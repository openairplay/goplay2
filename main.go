package main

import (
	"flag"
	"github.com/grandcat/zeroconf"
	"goplay2/audio"
	"goplay2/config"
	"goplay2/event"
	"goplay2/globals"
	"goplay2/handlers"
	"goplay2/homekit"
	"goplay2/ptp"
	"goplay2/rtsp"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

func main() {
	var ifName string
	var deviceName string
	var syncLogFile string
	var delay int64

	flag.StringVar(&ifName, "i", "eth0", "Specify interface")
	flag.Int64Var(&delay, "delay", 50, "Specify hardware delay in ms (useful on slow computer)")
	flag.StringVar(&deviceName, "n", "goplay", "Specify device name")
	flag.StringVar(&syncLogFile, "sync", "/dev/null", "Specify audio sync log")
	flag.StringVar(&config.Config.AlsaPortName, "alsa", "pcm.default", "Specify Alsa Device - Linux only")
	flag.Parse() // after declaring flags we need to call it

	globals.ErrLog = log.New(os.Stderr, "Error:", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)

	iFace, err := net.InterfaceByName(ifName)
	if err != nil {
		panic(err)
	}
	macAddress := strings.ToUpper(iFace.HardwareAddr.String())
	ipAddresses, err := iFace.Addrs()
	if err != nil {
		panic(err)
	}
	var ipStringAddr []string
	for _, addr := range ipAddresses {
		switch v := addr.(type) {
		case *net.IPNet:
			ipStringAddr = append(ipStringAddr, v.IP.String())
		case *net.IPAddr:
			ipStringAddr = append(ipStringAddr, v.IP.String())
		}
	}
	homekit.Device = homekit.NewAccessory(macAddress, deviceName, airplayDevice())
	log.Printf("Starting goplay for device %v", homekit.Device)
	homekit.Server, err = homekit.NewServer(macAddress, deviceName, ipStringAddr)

	server, err := zeroconf.Register(deviceName, "_airplay._tcp", "local.",
		7000, homekit.Device.ToRecords(), []net.Interface{*iFace})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	clock := ptp.NewVirtualClock(delay)
	ptpServer := ptp.NewServer(clock)

	// Divided by 100 -> average size of a RTP packet
	audioBuffer := audio.NewRing(globals.BufferSize / 100)
	syncFile, err := os.OpenFile(syncLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		globals.ErrLog.Printf("Can not open %s, logging audio sync to stderr\n", syncLogFile)
		syncFile = os.Stderr
	}
	player := audio.NewPlayer(syncFile, clock, audioBuffer)

	wg := new(sync.WaitGroup)
	wg.Add(4)

	go func() {
		player.Run()
		wg.Done()
	}()

	go func() {
		event.RunEventServer()
		wg.Done()
	}()

	go func() {
		ptpServer.Serve()
		wg.Done()
	}()

	go func() {
		handler, e := handlers.NewRstpHandler(deviceName, player)
		if e != nil {
			panic(e)
		}
		e = rtsp.RunRtspServer(handler)
		if e != nil {
			panic(e)
		}
		wg.Done()
	}()

	wg.Wait()
}
