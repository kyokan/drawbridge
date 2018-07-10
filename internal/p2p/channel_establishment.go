package p2p

import (
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/kyokan/drawbridge/pkg"
	"errors"
	"math"
	"github.com/kyokan/drawbridge/internal/db"
)

type ChannelEstablishment struct {
	Config *pkg.Config
	DB     *db.DB
}

func (c *ChannelEstablishment) HandleOpenChannel(peer *Peer, req *lnwire.OpenChannel) (*lnwire.AcceptChannel, error) {
	if !c.Config.ChainHashes.ValidChainHash(req.ChainHash) {
		return nil, errors.New("unsupported chain hash")
	}

	return &lnwire.AcceptChannel{
		PendingChannelID:     req.PendingChannelID,
		DustLimit:            req.DustLimit,
		MaxValueInFlight:     math.MaxUint64,
		ChannelReserve:       req.ChannelReserve,
		HtlcMinimum:          req.HtlcMinimum,
		MinAcceptDepth:       8,
		CsvDelay:             8,
		MaxAcceptedHTLCs:     math.MaxUint16,
		FundingKey:           c.Config.SigningPubkey,
		RevocationPoint:      c.Config.SigningPubkey,
		PaymentPoint:         c.Config.SigningPubkey,
		DelayedPaymentPoint:  c.Config.SigningPubkey,
		HtlcPoint:            c.Config.SigningPubkey,
		FirstCommitmentPoint: c.Config.SigningPubkey,
	}, nil
}

func (c *ChannelEstablishment) HandleAcceptChannel(peer *Peer, req *lnwire.AcceptChannel) (*lnwire.FundingCreated, error) {
	return nil, nil
}
