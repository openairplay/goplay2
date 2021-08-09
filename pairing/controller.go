package pairing

import (
	"github.com/brutella/hc/db"
	"github.com/brutella/hc/hap/pair"
	"github.com/brutella/hc/util"
	"goplay2/config"
	"path/filepath"
)

type Controller struct {
	ctrl *pair.PairingController
	db   db.Database
}

func (c Controller) Handle(cont util.Container) (util.Container, error) {
	method := pair.PairMethodType(cont.GetByte(pair.TagPairingMethod))
	if method == 0x5 { // List pairing
		entities, err := c.db.Entities()
		if err != nil {
			return nil, err
		}

		out := util.NewTLV8Container()
		out.SetByte(pair.TagSequence, 0x2)

		for index, entity := range entities {
			out.SetString(pair.TagUsername, entity.Name)
			out.SetBytes(pair.TagPublicKey, entity.PublicKey)
			out.SetByte(pair.TagPermission, 0x01)
			if index != 0 {
				out.SetByte(255, 0)
			}
		}
		return out, nil
	}
	return c.ctrl.Handle(cont)
}

func NewController(deviceName string) (*Controller, error) {
	storage, err := util.NewFileStorage(filepath.Join(config.Config.DataDirectory, "db", deviceName))
	if err != nil {
		return nil, err
	}
	db := db.NewDatabaseWithStorage(storage)
	return &Controller{ctrl: pair.NewPairingController(db), db: db}, nil
}
