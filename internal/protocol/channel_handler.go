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
	"github.com/ethereum/go-ethereum/common"
	"context"
	"time"
	)

type ChannelHandler struct {
	peerBook           *p2p.PeerBook
	km                 *wallet.KeyManager
	client             *ethclient.Client
	db                 *db.DB
	pendingChannels    map[common.Hash]*pendingChannel
	finalizingChannels map[common.Hash]*pendingChannel
	mtx                sync.Mutex
}

type pendingChannel struct {
	InputID          common.Hash
	ChannelID        common.Hash
	PendingChannelID common.Hash
	FundingAmount    *big.Int
	OurFundingKey    *crypto.PublicKey
	TheirFundingKey  *crypto.PublicKey
	OurSignature     crypto.Signature
	SentLocked       bool
	ReceivedLocked   bool
	FundingOutput common.Hash
}

func NewChannelHandler(peerBook *p2p.PeerBook, km *wallet.KeyManager, client *ethclient.Client, db *db.DB) *ChannelHandler {
	return &ChannelHandler{
		peerBook:           peerBook,
		km:                 km,
		client:             client,
		db:                 db,
		pendingChannels:    make(map[common.Hash]*pendingChannel),
		finalizingChannels: make(map[common.Hash]*pendingChannel),
	}
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
	case wire.MsgOpenChannel, wire.MsgAcceptChannel, wire.MsgFundingCreated, wire.MsgFundingSigned, wire.MsgFundingLocked:
		return true
	default:
		return false
	}
}

func (c *ChannelHandler) Accept(envelope *p2p.Envelope) (lnwire.Message, error) {
	msg := envelope.Msg
	switch msg.MsgType() {
	case wire.MsgOpenChannel:
		return c.onOpenChannel(msg.(*wire.OpenChannel))
	case wire.MsgAcceptChannel:
		return c.onAcceptChannel(msg.(*wire.AcceptChannel))
	case wire.MsgFundingCreated:
		return c.onFundingCreated(msg.(*wire.FundingCreated))
	case wire.MsgFundingSigned:
		return c.onFundingSigned(msg.(*wire.FundingSigned))
	case wire.MsgFundingLocked:
		return c.onFundingLocked(msg.(*wire.FundingLocked))
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

	paymentScript := txout.NewPayment(c.km.PublicKey().ETHAddress())
	input, err := c.db.Outputs.FindSpendableByOwnerAmount(paymentScript, pending.FundingAmount)
	if err != nil {
		return nil, err
	}

	sigHash, err := genFundingSigHash(
		input.ID,
		pending.FundingAmount,
		pending.OurFundingKey.ETHAddress(),
		pending.TheirFundingKey.ETHAddress(),
	)
	sig, err := c.km.SignData(sigHash)
	if err != nil {
		return nil, err
	}

	c.mtx.Lock()
	pending.InputID = input.ID
	outputId, err := genMultisigId(pending)
	if err != nil {
		c.mtx.Unlock()
		return nil, err
	}
	chanId := finalizeChannelID(outputId)
	c.finalizingChannels[chanId] = pending
	pending.ChannelID = chanId
	pending.OurSignature = sig
	c.mtx.Unlock()

	return &wire.FundingCreated{
		PendingChannelID: pending.PendingChannelID,
		InputID:          input.ID,
		Sig:              sig,
	}, nil
}

func (c *ChannelHandler) onFundingCreated(msg *wire.FundingCreated) (lnwire.Message, error) {
	c.mtx.Lock()
	pending, exists := c.pendingChannels[msg.PendingChannelID]
	if !exists {
		defer c.mtx.Unlock()
		return nil, errors.New("no channel with that pending id found")
	}
	c.mtx.Unlock()

	sigHash, err := genFundingSigHash(
		msg.InputID,
		pending.FundingAmount,
		pending.OurFundingKey.ETHAddress(),
		pending.TheirFundingKey.ETHAddress(),
	)
	if err != nil {
		return nil, err
	}
	ok := msg.Sig.Verify(sigHash, pending.TheirFundingKey)
	if !ok {
		return nil, errors.New("signature verification failed")
	}

	// TODO: set up commitment transaction so allow for non-cooperative exit
	// for now just sign the sigHash myself and generate the chan ID

	sig, err := c.km.SignData(sigHash)
	if err != nil {
		return nil, err
	}

	c.mtx.Lock()
	pending.InputID = msg.InputID
	outputId, err := genMultisigId(pending)
	if err != nil {
		c.mtx.Unlock()
		return nil, err
	}
	chanId := finalizeChannelID(outputId)
	pending.FundingOutput = outputId
	c.finalizingChannels[chanId] = pending
	pending.ChannelID = chanId
	c.mtx.Unlock()

	return &wire.FundingSigned{
		ChannelID: chanId,
		Sig:       sig,
	}, nil
}

func (c *ChannelHandler) onFundingSigned(msg *wire.FundingSigned) (lnwire.Message, error) {
	c.mtx.Lock()
	finalizing, exists := c.finalizingChannels[msg.ChannelID]
	if !exists {
		defer c.mtx.Unlock()
		return nil, errors.New("no channel with that id found")
	}
	c.mtx.Unlock()

	sigHash, err := genFundingSigHash(
		finalizing.InputID,
		finalizing.FundingAmount,
		finalizing.OurFundingKey.ETHAddress(),
		finalizing.TheirFundingKey.ETHAddress(),
	)
	if err != nil {
		return nil, err
	}
	ok := msg.Sig.Verify(sigHash, finalizing.TheirFundingKey)
	if !ok {
		return nil, errors.New("signature verification failed")
	}

	spendReq := genSpendRequest(
		finalizing.InputID,
		finalizing.FundingAmount,
		finalizing.OurFundingKey.ETHAddress(),
		finalizing.TheirFundingKey.ETHAddress(),
	)
	_, err = c.client.DepositMultisig(spendReq, finalizing.OurSignature)

	outputIds, err := txout.GenOutputIDs(spendReq)
	if err != nil {
		return nil, err
	}
	outputId := outputIds[0]
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute*5)
	defer cancel()
	_, err = ethclient.AwaitOutput(ctx, c.db, outputId)
	if err != nil {
		return nil, err
	}

	c.mtx.Lock()
	finalizing.SentLocked = true
	c.mtx.Unlock()

	err = c.db.Channels.Save(&db.ETHChannel{
		ID: finalizing.ChannelID,
		FundingOutput: outputId,
		Counterparty: finalizing.TheirFundingKey.ETHAddress(),
	})
	if err != nil {
		return nil, err
	}

	return &wire.FundingLocked{
		ChannelID: finalizing.ChannelID,
	}, err
}

func (c *ChannelHandler) onFundingLocked(msg *wire.FundingLocked) (lnwire.Message, error) {
	c.mtx.Lock()
	finalizing, exists := c.finalizingChannels[msg.ChannelID]
	if !exists {
		defer c.mtx.Unlock()
		return nil, errors.New("no channel with that id found")
	}
	c.mtx.Unlock()

	var res lnwire.Message

	if !finalizing.SentLocked {
		res = &wire.FundingLocked{
			ChannelID: finalizing.ChannelID,
		}

		c.db.Channels.Save(&db.ETHChannel{
			ID: finalizing.ChannelID,
			FundingOutput: finalizing.FundingOutput,
			Counterparty: finalizing.TheirFundingKey.ETHAddress(),
		})
	}

	c.mtx.Lock()
	delete(c.pendingChannels, finalizing.PendingChannelID)
	delete(c.finalizingChannels, finalizing.ChannelID)
	c.mtx.Unlock()

	return res, nil
}

func genMultisigId(finalizing *pendingChannel) (common.Hash, error) {
	var res common.Hash
	spendReq := genSpendRequest(
		finalizing.InputID,
		finalizing.FundingAmount,
		finalizing.OurFundingKey.ETHAddress(),
		finalizing.TheirFundingKey.ETHAddress(),
	)
	outputIds, err := txout.GenOutputIDs(spendReq)
	if err != nil {
		return res, err
	}
	return outputIds[0], nil
}

func genSpendRequest(inputId common.Hash, amount *big.Int, us common.Address, them common.Address) *txout.SpendRequest {
	paymentWitness := txout.NewPaymentWitness()
	multisigOutput := txout.NewMultisig(us, them)
	return &txout.SpendRequest{
		InputID: inputId,
		Witness: paymentWitness,
		Values: []*big.Int{
			amount,
		},
		Outputs: []txout.Output{
			multisigOutput,
		},
	}
}

func genFundingSigHash(inputId common.Hash, amount *big.Int, us common.Address, them common.Address) ([]byte, error) {
	return txout.SigData(genSpendRequest(inputId, amount, us, them))
}

func finalizeChannelID(inputId common.Hash) common.Hash {
	// flip bits as per bolt 2 chanid generation. funding output id is always zero
	// in this model, so XOR by zero
	// TODO: currently broken :(
	var chanId common.Hash
	copy(chanId[:], inputId[:])
	chanId[31] ^= byte(0)
	chanId[30] ^= byte(0)
	return chanId
}
