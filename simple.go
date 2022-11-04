package cellmodemd

import (
	"github.com/maltegrosse/go-modemmanager"
	"log"
)

type simpleConnector struct {
	logger        *log.Logger
	mmgr          modemmanager.ModemManager
	modem         modemmanager.Modem
	simplemodem   modemmanager.ModemSimple
	conproperties modemmanager.SimpleProperties
	Bearer        modemmanager.Bearer
}

type SimpleConnector interface {
	Connect() error
	GetBearer() modemmanager.Bearer
	WaitForDisconnect() (modemmanager.MMModemState, error)
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
	c := sc.modem.SubscribePropertiesChanged()
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
