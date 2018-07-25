package channel

import (
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/lightningnetwork/lnd/lnwire"
	"math/big"
	"github.com/kyokan/drawbridge/internal/wallet"
)

type HTLCManager struct {
	db *db.DB
}

func NewHTLCManager(db *db.DB) *HTLCManager {
	return &HTLCManager{
		db: db,
	}
}

func (m *HTLCManager) OriginateHTLC(channelId [32]byte, amount *big.Int) (*lnwire.UpdateAddHTLC, error) {
	id, err := m.db.HTLCs.LastHTLCID(channelId)

	if err != nil {
		return nil, err
	}

	nextId := id + 1
	paymentHash, err := wallet.Rand32Array()

	if err != nil {
		return nil, err
	}

	return &lnwire.UpdateAddHTLC{
		ChanID:      channelId,
		ID:          uint64(nextId),
		Amount:      lnwire.MilliSatoshi(amount.Uint64()),
		PaymentHash: paymentHash,
		Expiry:      1000,
	}, nil
}

func (m *HTLCManager) ReceiveHTLC(msg *lnwire.UpdateAddHTLC) (lnwire.Message, error) {
	return nil, nil
}
