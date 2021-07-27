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
	"sync"
)

func main() {
	var ifName string
	var delay int64
	var metricsFileName string
	var err error

	flag.StringVar(&config.Config.DeviceName, "n", "goplay", "Specify device name")
	flag.StringVar(&ifName, "i", "eth0", "Specify interface")
	flag.Int64Var(&delay, "delay", 0, "Specify hardware delay in ms")
	flag.StringVar(&config.Config.PulseSink, "sink", config.Config.PulseSink, "Specify Pulse Audio Sink - Linux only")
	flag.StringVar(&metricsFileName, "metrics", "/dev/null", "File name to logs audio sync - temp param")
	flag.BoolVar(&config.Config.DisableAudioSync, "nosync", config.Config.DisableAudioSync, "Disable multi-room/audio-sync. On slow CPU multi-room can cause audio jitter")
	flag.Parse() // after declaring flags we need to call it

	config.Config.Load()

	globals.ErrLog = log.New(os.Stderr, "Error:", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)

	iFace, macAddress, ipStringAddr := config.NetworkInfo(ifName)
	homekit.Device = homekit.NewAccessory(macAddress, config.Config.DeviceUUID, airplayDevice())
	log.Printf("Starting goplay for device %v", homekit.Device)
	homekit.Server, err = homekit.NewServer(macAddress, config.Config.DeviceName, ipStringAddr)

	server, err := zeroconf.Register(config.Config.DeviceName, "_airplay._tcp", "local.",
		7000, homekit.Device.ToRecords(), []net.Interface{*iFace})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	if metricFile, err := os.Create(metricsFileName); err == nil {
		globals.MetricLog = log.New(metricFile, "", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)
	} else {
		log.Panicf("Impossible to open metrics file %s", metricsFileName)
	}

	clock := ptp.NewVirtualClock(delay)
	ptpServer := ptp.NewServer(clock)

	player := audio.NewPlayer(clock, &config.Config.AudioMetrics)

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
