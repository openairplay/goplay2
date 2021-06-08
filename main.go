package main

import (
	"flag"
	"fmt"
	"github.com/brutella/hc/log"
	"github.com/grandcat/zeroconf"
	"goplay2/event"
	"goplay2/handlers"
	"goplay2/homekit"
	"goplay2/rtsp"
	"net"
	"os"
	"strings"
	"sync"
)

const deviceName = "aiwa"

func main() {

	var ifName string
	log.Debug.SetOutput(os.Stdout)
	flag.StringVar(&ifName, "i", "en0", "Specify interface")
	flag.Parse() // after declaring flags we need to call it

	iFace, err := net.InterfaceByName(ifName)
	if err != nil {
		panic(err)
	}
	macAddress := strings.ToUpper(iFace.HardwareAddr.String())
	homekit.Aiwa = homekit.NewAccessory(macAddress, aiwaDevice())
	fmt.Printf("Aiwa %v", homekit.Aiwa)
	homekit.Server, err = homekit.NewServer(macAddress, deviceName)

	server, err := zeroconf.Register(deviceName, "_airplay._tcp", "local.",
		7000, homekit.Aiwa.ToRecords(), []net.Interface{*iFace})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		event.RunEventServer()
		wg.Done()
	}()

	go func() {
		rtsp.RunRtspServer(handlers.NewRstpHandler())
		wg.Done()
	}()

	wg.Wait()
}
