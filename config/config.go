package config

import (
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type Configuration struct {
	Volume       float64 `json:"sound-volume"`
	DeviceUUID   string  `json:"device-uuid"`
	PulseSink    string  `json:"-"`
	DeviceName   string  `json:"-"`
	DataDirectory string `json:"data-directory"`
	exitsSignals chan os.Signal
	baseDir		 string
}

var Config = &Configuration{
	PulseSink:  "",
	Volume:     -999,
	DeviceUUID: uuid.NewString(),
}

func (c *Configuration) Load(baseDir string) error {
	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	c.baseDir = baseDir
	configFilePath := filepath.Join(c.baseDir, c.DeviceName, "config.json")
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil || json.Unmarshal(data, &c) != nil {
		log.Printf("%s is not valid - a new file will be created at program exit\n", configFilePath)
	}
	if c.DataDirectory == "" {
		c.DataDirectory = baseDir
	}

	c.exitsSignals = make(chan os.Signal, 1)
	signal.Notify(c.exitsSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c.exitsSignals
		err := c.Store()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()

	return nil
}

func (c *Configuration) Store() error {
	data, err := json.Marshal(&c)
	if err != nil {
		log.Printf("Warning: impossible to marshal configuration in json\n")
		return err
	}
	configFilePath := filepath.Join(c.baseDir, c.DeviceName, "config.json")
	dirPath := filepath.Join(c.baseDir, c.DeviceName)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0755)
		if err != nil{
			log.Printf("Warning : impossible to create config directory %s \n", filepath.Join(c.baseDir, c.DeviceName))
			return err
		}
	}
	err = ioutil.WriteFile(configFilePath, data, 0660)
	if err != nil {
		log.Printf("Warning : impossible to store config file %s \n", configFilePath)
	}
	return err
}
