package homekit

import (
	"fmt"
	"goplay2/globals"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type UUID struct {
	uuid.UUID
}

type Accessory struct {
	Srcvers   string
	Deviceid  string
	Features  globals.Features
	flags     string
	model     string
	protovers string
	acl       string
	rsf       string
	Pi        UUID
	Gid       UUID
	Psi       UUID
	gcgl      string
	igl       string
	pk        string
}

func NewAccessory(deviceId string, features globals.Features) *Accessory {
	var currentUuid = GetUUID(deviceId)
	return &Accessory{
		Srcvers:   "366.0",
		Deviceid:  deviceId,
		Features:  features,
		flags:     "0x4",
		model:     "GoPlay2",
		protovers: "1.1",
		acl:       "0",
		rsf:       "0x0",
		Pi:        UUID{currentUuid},
		Gid:       UUID{currentUuid},
		Psi:       UUID{currentUuid},
		gcgl:      "0",
		igl:       "0",
		pk:        "b07727d6f6cd6e08b58ede525ec3cdeaa252ad9f683feb212ef8a205246554e7",
	}
}

func (t *Accessory) String() string {
	return fmt.Sprintf("Pi: %s, guid: %s, Psi: %s", t.Pi, t.Gid, t.Psi)
}

func (uid UUID) ToRecord() string {
	return uid.String()
}

func (t *Accessory) ToRecords() []string {

	fields := reflect.TypeOf(*t)
	values := reflect.ValueOf(*t)

	numField := values.NumField()
	results := make([]string, numField)

	for i := 0; i < numField; i++ {
		results[i] = strings.ToLower(fields.Field(i).Name) + "="
		value := values.Field(i)
		switch fields.Field(i).Type.Name() {
		case "string":
			results[i] += value.String()
		case "UUID":
			results[i] += value.Interface().(UUID).ToRecord()
		case "Features":
			results[i] += value.Interface().(globals.Features).ToRecord()
		default:
			panic(fields.Field(i).Type.Name())
		}
	}
	return results

}

var Device *Accessory
