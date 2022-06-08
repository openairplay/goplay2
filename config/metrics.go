package config

import (
	"encoding/json"
	"goplay2/globals"
	"io/ioutil"
	"log"
	"time"
)

type Metrics struct {
	CountDrop      uint32 `json:"count-drop"`
	CountSilence   uint32 `json:"count-silence"`
	DriftAverage   int64  `json:"drift-average"`
	PacketReceived int64  `json:"packet-received"`
}

func (m Metrics) Store(deviceName string) {
	data, err := json.Marshal(&m)
	if err != nil {
		log.Printf("Warning: impossible to marshal configuration in json")
	}
	err = ioutil.WriteFile(deviceName+"/metrics.json", data, 0660)
	if err != nil {
		log.Printf("Warning : impossible to store config file %s \n", deviceName+"/config.json")
	}
}

func (m *Metrics) Drop() {
	m.CountDrop += 1
	globals.MetricLog.Printf("Drop sequence because of drift")
}

func (m *Metrics) Silence() {
	globals.MetricLog.Printf("filling audio buffer with silence")
	m.CountSilence += 1
}

func (m *Metrics) Drift(drift time.Duration) {
	globals.MetricLog.Printf("drift : %v\n", drift)
	if drift < 0 {
		drift = -drift
	}
	m.PacketReceived += 1 // 1-based count ( to prevent divide by zero )
	m.DriftAverage = m.DriftAverage + (drift.Milliseconds()-m.DriftAverage)/m.PacketReceived
}
