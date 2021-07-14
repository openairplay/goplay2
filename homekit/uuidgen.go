package homekit

import (
	"fmt"
	"io/ioutil"

	"github.com/google/uuid"
)

func GetUUID(deviceName string) uuid.UUID {
	uuidStore := fmt.Sprintf("%s/uuid.cfg", deviceName)
	content, err := ioutil.ReadFile(uuidStore)
	if err != nil || len(content) == 0 {
		newUUID := uuid.New()
		ioutil.WriteFile(uuidStore, []byte(newUUID.String()), 0644)
		return newUUID
	}
	return uuid.MustParse(string(content))
}

