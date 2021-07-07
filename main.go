package main

import (
	"flag"
	"github.com/grandcat/zeroconf"
	"goplay2/config"
	"goplay2/event"
	"goplay2/handlers"
	"goplay2/homekit"
	"goplay2/ptp"
	"goplay2/rtsp"
	"log"
	"net"
	"strings"
	"sync"
)

func main() {
	var ifName string
	var deviceName string
	var delay int64

	flag.StringVar(&ifName, "i", "eth0", "Specify interface")
	flag.Int64Var(&delay, "delay", 50, "Specify hardware delay in ms (useful on slow computer)")
	flag.StringVar(&deviceName, "n", "goplay", "Specify device name")
	flag.StringVar(&config.Config.AlsaPortName, "alsa", "pcm.default", "Specify Alsa Device - Linux only")
	flag.Parse() // after declaring flags we need to call it

	iFace, err := net.InterfaceByName(ifName)
	if err != nil {
		panic(err)
	}
	macAddress := strings.ToUpper(iFace.HardwareAddr.String())
	homekit.Device = homekit.NewAccessory(macAddress, airplayDevice())
	log.Printf("Device %v", homekit.Device)
	homekit.Server, err = homekit.NewServer(macAddress, deviceName)

	server, err := zeroconf.Register(deviceName, "_airplay._tcp", "local.",
		7000, homekit.Device.ToRecords(), []net.Interface{*iFace})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	clock := ptp.NewVirtualClock(delay)
	ptp := ptp.NewServer(clock)

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go func() {
		event.RunEventServer()
		wg.Done()
	}()

	go func() {
		ptp.Serve()
	}()

	go func() {
		rtsp.RunRtspServer(handlers.NewRstpHandler(clock))
		wg.Done()
	}()

	wg.Wait()
}
