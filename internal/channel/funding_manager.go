package channel

import (
	"github.com/kyokan/drawbridge/internal/p2p"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/kyokan/drawbridge/internal/wallet"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/kyokan/drawbridge/pkg"
	"math"
	"github.com/roasbeef/btcutil"
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/kyokan/drawbridge/internal/eth"
	btcwire "github.com/roasbeef/btcd/wire"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/kyokan/drawbridge/pkg/wire"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type FundingManager struct {
	peerBook *p2p.PeerBook
	db       *db.DB
	km       *wallet.KeyManager
	client   *eth.Client
	config   *pkg.Config
}

func NewFundingManager(peerBook *p2p.PeerBook, db *db.DB, km *wallet.KeyManager, client *eth.Client, config *pkg.Config) *FundingManager {
	return &FundingManager{
		peerBook: peerBook,
		db:       db,
		km:       km,
		client:   client,
		config:   config,
	}
}

func (m *FundingManager) InitChannel(pub *crypto.PublicKey, amount *big.Int) error {
	peer := m.peerBook.FindPeer(pub)

	if peer == nil {
		return errors.New("peer not found")
	}

	pendingCidSlice, err := wallet.Rand32()

	if err != nil {
		return err
	}

	var pendingCid [32]byte
	copy(pendingCid[:], pendingCidSlice)

	commitmentSeed, firstCommitmentPoint, err := wallet.FirstCommitmentPoint()

	if err != nil {
		return err
	}

	signingKey := m.config.SigningPubkey.BTCEC()

	msg := &lnwire.OpenChannel{
		ChainHash:            m.config.ChainHashes.TestToken,
		PendingChannelID:     pendingCid,
		FundingAmount:        btcutil.Amount(amount.Int64()),
		PushAmount:           0,
		DustLimit:            0,
		MaxValueInFlight:     math.MaxUint64,
		ChannelReserve:       0,
		HtlcMinimum:          0,
		FeePerKiloWeight:     0,
		CsvDelay:             7,
		MaxAcceptedHTLCs:     math.MaxUint16,
		FundingKey:           signingKey,
		RevocationPoint:      signingKey,
		PaymentPoint:         signingKey,
		DelayedPaymentPoint:  signingKey,
		HtlcPoint:            signingKey,
		FirstCommitmentPoint: firstCommitmentPoint,
		ChannelFlags:         0,
	}

	err = m.db.Channels.CreateLocalChannel(msg, commitmentSeed)

	if err != nil {
		return err
	}

	return peer.Send(msg)
}

func (m *FundingManager) CanAccept(msg lnwire.Message) bool {
	switch msg.MsgType() {
	case lnwire.MsgOpenChannel, lnwire.MsgAcceptChannel, lnwire.MsgFundingCreated, lnwire.MsgFundingSigned:
		return true
	default:
		return false
	}

	return false
}

func (m *FundingManager) Accept(msg lnwire.Message) (lnwire.Message, error) {
	switch msg.MsgType() {
	case lnwire.MsgOpenChannel:
		return m.onOpenChannel(msg.(*lnwire.OpenChannel))
	case lnwire.MsgAcceptChannel:
		return m.onAcceptChannel(msg.(*lnwire.AcceptChannel))
	case lnwire.MsgFundingCreated:
		return m.onFundingCreated(msg.(*lnwire.FundingCreated))
	case lnwire.MsgFundingSigned:
		return m.onFundingSigned(msg.(*lnwire.FundingSigned))
	default:
		return nil, errors.New("unsupported message type")
	}
}

func (m *FundingManager) onOpenChannel(open *lnwire.OpenChannel) (*lnwire.AcceptChannel, error) {
	commitmentSeed, firstCommitmentPoint, err := wallet.FirstCommitmentPoint()

	if err != nil {
		return nil, err
	}

	signingKey := m.config.SigningPubkey.BTCEC()

	accept := &lnwire.AcceptChannel{
		PendingChannelID:     open.PendingChannelID,
		DustLimit:            open.DustLimit,
		MaxValueInFlight:     open.MaxValueInFlight,
		ChannelReserve:       open.ChannelReserve,
		HtlcMinimum:          open.HtlcMinimum,
		MinAcceptDepth:       7,
		CsvDelay:             open.CsvDelay,
		MaxAcceptedHTLCs:     open.MaxAcceptedHTLCs,
		FundingKey:           signingKey,
		RevocationPoint:      signingKey,
		PaymentPoint:         signingKey,
		DelayedPaymentPoint:  signingKey,
		HtlcPoint:            signingKey,
		FirstCommitmentPoint: firstCommitmentPoint,
	}

	err = m.db.Channels.CreateRemoteChannel(open, accept, commitmentSeed)

	if err != nil {
		return nil, err
	}

	return accept, nil
}

func (m *FundingManager) onAcceptChannel(msg *lnwire.AcceptChannel) (*lnwire.FundingCreated, error) {
	err := m.db.Channels.AcceptLocalChannel(msg)

	if err != nil {
		return nil, err
	}

	ch, err := m.db.Channels.GetPendingChannel(msg.PendingChannelID)

	if err != nil {
		return nil, err
	}

	ourAddr := ch.OurFundingAddress.ETHAddress()
	theirAddr := ch.TheirFundingAddress.ETHAddress()
	input, err := m.db.UTXOs.FindSpendableByOwnerAmount(ch.FundingAmount, ourAddr)

	if err != nil {
		return nil, err
	}

	multisig := wire.NewMultisig(input.ID, ourAddr, theirAddr)
	multisigId := multisig.ID()
	multisigHash := multisig.Hash()

	fmt.Println(hexutil.Encode(multisigId[:]), hexutil.Encode(multisigHash[:]))

	err = m.db.Channels.FinalizeChannelId(msg.PendingChannelID, multisigId, input.ID[:])

	if err != nil {
		return nil, err
	}

	sig, err := m.km.SignData(multisigHash[:])

	if err != nil {
		return nil, err
	}

	err = m.db.Channels.FinalizeChannelSignatures(multisigId, sig, nil)

	return &lnwire.FundingCreated{
		PendingChannelID: msg.PendingChannelID,
		FundingPoint: btcwire.OutPoint{
			Hash:  multisigId,
			Index: 0,
		},
		CommitSig: sig.Wire(),
	}, nil
}

func (m *FundingManager) onFundingCreated(msg *lnwire.FundingCreated) (*lnwire.FundingSigned, error) {
	err := m.db.Channels.FinalizeChannelId(msg.PendingChannelID, msg.FundingPoint.Hash, nil)

	if err != nil {
		return nil, err
	}

	sig, err := m.km.SignData(msg.FundingPoint.Hash[:])

	if err != nil {
		return nil, err
	}

	err = m.db.Channels.FinalizeChannelSignatures(msg.FundingPoint.Hash, sig.Bytes(), crypto.SignatureFromWire(msg.CommitSig))

	return &lnwire.FundingSigned{
		ChanID:    [32]byte(msg.FundingPoint.Hash),
		CommitSig: sig.Wire(),
	}, nil
}

func (m *FundingManager) onFundingSigned(msg *lnwire.FundingSigned) (*lnwire.FundingLocked, error) {
	err := m.db.Channels.FinalizeChannelSignatures(msg.ChanID, nil, crypto.SignatureFromWire(msg.CommitSig))

	if err != nil {
		return nil, err
	}

	ch, err := m.db.Channels.GetFinalizedChannel(msg.ChanID)

	if err != nil {
		return nil, err
	}

	_, err = m.client.DepositMultisig(&eth.DepositMultisigOpts{
		InputID: ch.InputID,
		OurAddress: ch.OurFundingAddress.ETHAddress(),
		TheirAddress: ch.TheirFundingAddress.ETHAddress(),
		OurSignature: ch.OurSignature,
		TheirSignature: ch.TheirSignature,
	})

	if err != nil {
		return nil, err
	}

	next, err := crypto.RandomPublicKey()

	if err != nil {
		return nil, err
	}

	return &lnwire.FundingLocked{
		ChanID: msg.ChanID,
		NextPerCommitmentPoint: next.BTCEC(),
	}, nil
}