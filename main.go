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
	var delay int64
	var configurationBaseDir string
	var flagsConfig config.Configuration

	// These two are being used to construct the config path
	// XXX if the config has been manually update, it could very well have a devicename that does not align with the directory it's in
	flag.StringVar(&configurationBaseDir, "c", "", "Configuration base directory (default to current directory)")
	flag.StringVar(&config.Config.DeviceName, "n", "goplay", "Specify device name")

	// These two should override whatever is in the currently store config
	flag.StringVar(&flagsConfig.DataDirectory, "d", "", "Data base directory (defaults to configuration directory)")
	flag.StringVar(&flagsConfig.PulseSink, "sink", config.Config.PulseSink, "Specify Pulse Audio Sink - Linux only")

	// These are not stored in permanent config
	flag.StringVar(&ifName, "i", "eth0", "Specify interface")
	flag.Int64Var(&delay, "delay", 0, "Specify hardware delay in ms")
	flag.Parse() // after declaring flags we need to call it

	// Load the possibly existing config
	err := config.Config.Load(configurationBaseDir)
	if err != nil {
		panic(err)
	}
	defer config.Config.Store()

	// Override config specifics with command-line flags
	if flagsConfig.DataDirectory != "" {
		config.Config.DataDirectory = flagsConfig.DataDirectory
	}
	if flagsConfig.PulseSink != "" {
		config.Config.PulseSink = flagsConfig.PulseSink
	}

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
	homekit.Device = homekit.NewAccessory(macAddress, config.Config.DeviceUUID, airplayDevice())
	log.Printf("Starting goplay for device %v", homekit.Device)
	homekit.Server, err = homekit.NewServer(macAddress, config.Config.DeviceName, ipStringAddr)

	server, err := zeroconf.Register(config.Config.DeviceName, "_airplay._tcp", "local.",
		7000, homekit.Device.ToRecords(), []net.Interface{*iFace})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	clock := ptp.NewVirtualClock(delay)
	ptpServer := ptp.NewServer(clock)

	// Divided by 100 -> average size of a RTP packet
	audioBuffer := audio.NewRing(globals.BufferSize / 100)

	player := audio.NewPlayer(clock, audioBuffer)

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
		handler, e := handlers.NewRstpHandler(config.Config.DeviceName, player)
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
