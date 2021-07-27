package config

import (
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Configuration struct {
	Volume           float64 `json:"sound-volume"`
	DeviceUUID       string  `json:"device-uuid"`
	PulseSink        string  `json:"-"`
	DeviceName       string  `json:"-"`
	DisableAudioSync bool    `json:"-"`
	AudioMetrics     Metrics `json:"-"`
	exitsSignals     chan os.Signal
}

var Config = &Configuration{
	PulseSink:        "",
	Volume:           -999,
	DeviceUUID:       uuid.NewString(),
	DisableAudioSync: false,
}

func (c *Configuration) Load() {
	data, err := ioutil.ReadFile(c.DeviceName + "/config.json")
	if err != nil || json.Unmarshal(data, &c) != nil {
		log.Printf("%s is not valid - at new file will be created at program exit\n", c.DeviceName+"/config.json")
	}
	c.exitsSignals = make(chan os.Signal, 1)
	signal.Notify(c.exitsSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c.exitsSignals
		c.Store()
		c.AudioMetrics.Store(c.DeviceName)
		os.Exit(0)
	}()
}

func (c *Configuration) Store() {
	data, err := json.Marshal(&c)
	if err != nil {
		log.Printf("Warning: impossible to marshal configuration in json")
	}
	err = ioutil.WriteFile(c.DeviceName+"/config.json", data, 0660)
	if err != nil {
		log.Printf("Warning : impossible to store config file %s \n", c.DeviceName+"/config.json")
	}
}

func NetworkInfo(ifName string) (*net.Interface, string, []string) {
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

	return iFace, macAddress, ipStringAddr
}
