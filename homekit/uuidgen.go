package homekit

import (
	"github.com/google/uuid"
)

func GetUUID(deviceName string) uuid.UUID {
	newUUID := uuid.New()
	return newUUID
}

