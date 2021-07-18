package homekit

import (
	"github.com/brutella/hc/db"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/util"
)

// reference : https://github.com/ejurgensen/pair_ap
const airportExpressHardcodedPin = "3939"

type server struct {
	Context  hap.Context
	Database db.Database
	Device   hap.SecuredDevice
	Name     string
	Ips      []string
}

func NewServer(macAddress string, deviceName string, ips []string) (*server, error) {

	storage, err := util.NewFileStorage(deviceName)
	if err != nil {
		return nil, err
	}
	database := db.NewDatabaseWithStorage(storage)

	device, err := hap.NewSecuredDevice(macAddress, airportExpressHardcodedPin, database)
	if err != nil {
		return nil, err
	}

	return &server{Context: hap.NewContextForSecuredDevice(device),
		Database: database,
		Device:   device,
		Name:     deviceName,
		Ips:      ips,
	}, nil
}

var Server *server
