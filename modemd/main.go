// build for primary target with
// GOARCH=arm GOARM=7 GOOS=linux go build -o rpi-cellmodemd gopkg.in/kainz/cellmodemd.v0/modemd 
package main


import (
	"github.com/maltegrosse/go-modemmanager"
	"log"
	"fmt"
	"flag"
	cellmodemd "gopkg.in/kainz/cellmodemd.v0"
)

func main() {
		mmgr, err := modemmanager.NewModemManager()
		if err != nil {
			log.Fatal(err.Error())
		}
		version, err := mmgr.GetVersion()
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println("ModemManager version", version)

		apnPtr := flag.String("a", "", "APN to connect")
		v6Mode := flag.Bool("6", false, "Enable ipv6 call")

		flag.Parse()

		if *apnPtr == "" {
			log.Fatal("must specify apn with -a")
		}

		connector, err := cellmodemd.GetConnector(mmgr, 0, *apnPtr, *v6Mode, log.Default())
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println("Connecting device")

		err = connector.Connect()
		if err != nil {
			log.Fatal(err.Error())
		}

		bearer := connector.GetBearer()
		//bjson, err := bearer.MarshalJSON()
		//if err != nil {
		//	log.Fatal(err.Error())
		//}

		c_if, err := bearer.GetInterface()
		if err != nil {
			log.Fatal(err.Error())
		}

		c_ip4, err := bearer.GetIp4Config()
		if err != nil {
			log.Fatal(err.Error())
		}

		c_ip6, err := bearer.GetIp6Config()
		if err != nil {
			log.Fatal(err.Error())
		}

		c_bprops, err := bearer.GetProperties()
		if err != nil {
			log.Fatal(err.Error())
		}

		//log.Println(string(bjson))
		log.Println("bearer properties ", c_bprops)

		log.Println("should configure interface ",
					c_if,
					" with ip4 ",
					c_ip4,
					" with ip6 ",
					c_ip6)

		log.Println("templating interface file and reloading")
		err = connector.TriggerInterface()

		if err != nil {
			log.Fatal("error triggering interface",err.Error())
		}

		log.Println("waiting for state change")

		state, err := connector.WaitForDisconnect()
		if err != nil {
			log.Fatal("error waiting for state change ",err.Error())
		}

		log.Println("exiting due to non-connect(ed/ing) state: ", state)

}
