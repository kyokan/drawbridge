package channel

import (
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/lightningnetwork/lnd/lnwire"
)

type HTLCManager struct {
	db *db.DB
}

func NewHTLCManager(db *db.DB) *HTLCManager {
	return &HTLCManager{
		db: db,
	}
}

func (m *HTLCManager) OriginateHTLC(channelId [32]byte) error {
	return nil
}

func (m *HTLCManager) ReceiveHTLC(msg *lnwire.UpdateAddHTLC) error {
	return nil
}