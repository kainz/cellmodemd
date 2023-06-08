package cellmodemd

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/renameio"
	"github.com/maltegrosse/go-modemmanager"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"
)

type simpleConnector struct {
	logger        *log.Logger
	mmgr          modemmanager.ModemManager
	modem         modemmanager.Modem
	simplemodem   modemmanager.ModemSimple
	conproperties modemmanager.SimpleProperties
	Bearer        modemmanager.Bearer
}

var ErrNoModem = errors.New("No modem found")

type SimpleConnector interface {
	Connect() error
	GetBearer() modemmanager.Bearer
	WaitForDisconnect() (modemmanager.MMModemState, error)
	TriggerInterface() error
}

func GetConnector(mmgr modemmanager.ModemManager, index int, apn string, logger *log.Logger) (SimpleConnector, error) {
	var c simpleConnector

	c.mmgr = mmgr
	c.logger = logger
	c.conproperties.Apn = apn

	return &c, c.init(index)
}

func (sc *simpleConnector) init(index int) error {
	modems, err := sc.mmgr.GetModems()
	if err != nil {
		return err
	}
	if len(modems) == 0 {
		return ErrNoModem
	}

	sc.modem = modems[index]

	md, err := sc.modem.MarshalJSON()
	if err != nil {
		return err
	}
	log.Print("selected modem # ", index, " data: ", string(md))

	sc.simplemodem, err = sc.modem.GetSimpleModem()
	if err != nil {
		return err
	}

	return nil
}

func (sc *simpleConnector) WaitForDisconnect() (modemmanager.MMModemState, error) {
	// A slight misnomer, we wait and return on any non connecting state. TODO: make this some sort of unidirectional DFA to manage state changes as seen
	c := sc.modem.SubscribeStateChanged()
	var err error
	for v := range c {
		allowed_states := []modemmanager.MMModemState{
			modemmanager.MmModemStateConnecting,
			modemmanager.MmModemStateConnected}
		log.Println(v)
		oldState, newState, reason, err := sc.modem.ParseStateChanged(v)
		if err == nil {
			log.Println("got state change from ", oldState, " to ", newState, " because ", reason)
			exit := true
			for t := range allowed_states {
				if newState == allowed_states[t] {
					exit = false
					break
				}
			}
			if exit {
				return newState, nil
			}
			return newState, err
		} else {
			log.Println("eating error in parsed state because ", err)
		}
	}
	return modemmanager.MmModemStateUnknown, err
}

func (sc *simpleConnector) Connect() error {
	bearer, err := sc.simplemodem.Connect(sc.conproperties)
	if err != nil {
		return err
	}

	sc.Bearer = bearer

	status, err := sc.simplemodem.GetStatus()
	if err != nil {
		sc.logger.Println("Connected but got error fetching status:", err.Error())
		return err
	}

	sc.logger.Println("Successful connection with status: ", status)
	return nil
}

func (sc *simpleConnector) GetBearer() modemmanager.Bearer {
	return sc.Bearer
}

func (sc *simpleConnector) TriggerInterface() error {
	// writeout a systemd network file for the simple connection bearor, then trigger a networkctl reconfigure
	// TODO: use dbus to trigger systemd directly?
	// TODO: write effective MTU to dnsmasq and do whatevers needed there as well
	var methodValues = template.FuncMap{
		"MmBearerIpMethodUnknown": func() modemmanager.MMBearerIpMethod { return modemmanager.MmBearerIpMethodUnknown },
		"MmBearerIpMethodStatic":  func() modemmanager.MMBearerIpMethod { return modemmanager.MmBearerIpMethodStatic },
	}
	var tmpl = template.Must(template.New("systemd-network-template").Funcs(methodValues).Parse(systemd_network_template))

	var netdata bytes.Buffer
	netwriter := io.Writer(&netdata)
	netreader := io.Reader(&netdata)

	err := tmpl.Execute(netwriter, sc.Bearer)
	if err != nil {
		return err
	}

	intf, err := sc.Bearer.GetInterface()
	if err != nil {
		return err
	}

	targetpath := filepath.Join(systemd_network_prefix, intf+".network")

	log.Println("will generate network file at ", targetpath, " with content\n", netdata.String())
	targetMode := os.FileMode(0644)

	err = os.MkdirAll(systemd_network_prefix, os.FileMode(0755))
	if err != nil {
		return err
	}

	dir := renameio.TempDir(systemd_network_prefix)

	o, err := renameio.TempFile(dir, targetpath)
	if err != nil {
		return err
	}
	defer o.Cleanup()
	err = o.Chmod(targetMode)
	if err != nil {
		return err
	}
	io.Copy(o, netreader)
	err = o.CloseAtomicallyReplace()
	if err != nil {
		return err
	}

	log.Println("networkctl reloading")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	networkcmd := exec.CommandContext(ctx, "networkctl", "reload")

	output, err := networkcmd.CombinedOutput()
	log.Println("network ctl output was (", err, "):", string(output))
	if err != nil {
		return err
	}

	return nil
}

const (
	// TODO: determine what to do for multistack
	systemd_network_prefix string = "/run/systemd/network"
	// Disable LLDP, doesnt work on the cellmodems
	systemd_network_template string = `
[Match]
Name={{.GetInterface}}
{{$ip4 := .GetIp4Config -}}
{{- $ip6 := .GetIp6Config -}}
[Link]
{{ if gt $ip6.Mtu 0 }}
MTUBytes={{ $ip6.Mtu }}
{{ else }}
MTUBytes={{ $ip4.Mtu }}
{{ end }}

[Network]
LLDP=false
EmitLLDP=false
{{ if eq $ip4.Method MmBearerIpMethodUnknown }}
{{ else if ne $ip4.Method MmBearerIpMethodStatic }}
DHCP=yes
{{ else }}
{{ if $ip4.Dns1 }}DNS={{$ip4.Dns1}}{{end}}
{{ if $ip4.Dns2 }}DNS={{$ip4.Dns2}}{{end}}
{{ if $ip4.Dns3 }}DNS={{$ip4.Dns3}}{{end}}
[Address]
Address={{ $ip4.Address }}/{{ $ip4.Prefix }}
[Route]
Gateway={{ $ip4.Gateway }}
Metric=401
{{ end }}
{{ if eq $ip6.Method MmBearerIpMethodUnknown }}
{{ else if ne $ip6.Method MmBearerIpMethodStatic }}
{{ else }}
[Network]
LLDP=false
EmitLLDP=false
{{ if $ip6.Dns1 }}DNS={{$ip6.Dns1}}{{end}}
{{ if $ip6.Dns2 }}DNS={{$ip6.Dns2}}{{end}}
{{ if $ip6.Dns3 }}DNS={{$ip6.Dns3}}{{end}}
[Address]
Address={{ $ip6.Address }}/{{ $ip6.Prefix }}
[Route]
Gateway={{ $ip6.Gateway }}
Metric=401
{{ end -}}
	`
)
