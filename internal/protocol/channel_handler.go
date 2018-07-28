package protocol

import (
	"github.com/kyokan/drawbridge/internal/p2p"
	"github.com/kyokan/drawbridge/internal/wallet"
	"github.com/kyokan/drawbridge/internal/ethclient"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"math/big"
	"github.com/kyokan/drawbridge/pkg/wire"
	"errors"
	"github.com/lightningnetwork/lnd/lnwire"
	"sync"
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/kyokan/drawbridge/pkg/txout"
)

type chanId [32]byte

type ChannelHandler struct {
	peerBook        *p2p.PeerBook
	km              *wallet.KeyManager
	client          *ethclient.Client
	pendingChannels map[chanId]*pendingChannel
	mtx             sync.Mutex
	db              *db.DB
}

type pendingChannel struct {
	PendingChannelID chanId
	FundingAmount    *big.Int
	OurFundingKey    *crypto.PublicKey
	TheirFundingKey  *crypto.PublicKey
}

func (c *ChannelHandler) InitChannel(pub *crypto.PublicKey, amount *big.Int) error {
	peer := c.peerBook.FindPeer(pub)
	if peer == nil {
		return errors.New("peer not found")
	}

	cId, err := crypto.Rand32()
	if err != nil {
		return err
	}

	msg := &wire.OpenChannel{
		PendingChannelID: cId,
		FundingAmount:    amount,
		CsvDelay:         7,
		MaxAcceptedHTLCs: 2,
		FundingKey:       c.km.PublicKey(),
	}

	c.mtx.Lock()
	c.pendingChannels[msg.PendingChannelID] = &pendingChannel{
		PendingChannelID: msg.PendingChannelID,
		FundingAmount:    amount,
		OurFundingKey:    msg.FundingKey,
	}
	c.mtx.Unlock()

	return peer.Send(msg)
}

func (c *ChannelHandler) CanAccept(msg lnwire.Message) bool {
	switch msg.MsgType() {
	case wire.MsgOpenChannel:
		return true
	default:
		return false
	}
}

func (c *ChannelHandler) Accept(msg lnwire.Message) (lnwire.Message, error) {
	switch msg.MsgType() {
	case wire.MsgOpenChannel:
		return c.onOpenChannel(msg.(*wire.OpenChannel))
	case wire.MsgAcceptChannel:
		return c.onAcceptChannel(msg.(*wire.AcceptChannel))
	default:
		return nil, errors.New("unknown message type")
	}
}

func (c *ChannelHandler) onOpenChannel(msg *wire.OpenChannel) (lnwire.Message, error) {
	c.mtx.Lock()
	_, exists := c.pendingChannels[msg.PendingChannelID]
	if exists {
		defer c.mtx.Unlock()
		return nil, errors.New("duplicate pending channel id")
	}

	ourKey := c.km.PublicKey()
	c.pendingChannels[msg.PendingChannelID] = &pendingChannel{
		PendingChannelID: msg.PendingChannelID,
		FundingAmount:    msg.FundingAmount,
		TheirFundingKey:  msg.FundingKey,
		OurFundingKey:    ourKey,
	}
	c.mtx.Unlock()

	res := &wire.AcceptChannel{
		PendingChannelID: msg.PendingChannelID,
		CsvDelay:         msg.CsvDelay,
		MaxAcceptedHTLCs: msg.MaxAcceptedHTLCs,
		FundingKey:       ourKey,
	}

	return res, nil
}

func (c *ChannelHandler) onAcceptChannel(msg *wire.AcceptChannel) (lnwire.Message, error) {
	c.mtx.Lock()
	pending, exists := c.pendingChannels[msg.PendingChannelID]
	if !exists {
		defer c.mtx.Unlock()
		return nil, errors.New("no channel with that pending id found")
	}
	pending.TheirFundingKey = msg.FundingKey
	c.mtx.Unlock()

	out, err := c.db.Outputs.FindSpendableByOwnerAmount(c.km.PublicKey().ETHAddress(), pending.FundingAmount)
	if err != nil {
		return nil, err
	}

	paymentWitness := txout.NewPaymentWitness()
	multisigOutput := txout.NewMultisig(pending.OurFundingKey.ETHAddress(), pending.TheirFundingKey.ETHAddress())
	sigData, err := txout.SigData(&txout.SpendRequest{
		InputID: out.ID,
		Witness: paymentWitness,
		Values: []*big.Int {
			pending.FundingAmount,
		},
		Outputs: []txout.Output{
			multisigOutput,
		},
	})
	if err != nil {
		return nil, err
	}
	sig, err := c.km.SignData(sigData)
	if err != nil {
		return nil, err
	}

	return &wire.FundingCreated{
		PendingChannelID: pending.PendingChannelID,
		InputID: out.ID,
		Sig: sig.Bytes(),
	}, nil
}
