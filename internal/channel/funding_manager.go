package channel

import (
	"github.com/kyokan/drawbridge/internal/p2p"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/kyokan/drawbridge/internal/wallet"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/kyokan/drawbridge/pkg"
	"math"
	"github.com/btcsuite/btcutil"
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/kyokan/drawbridge/internal/ethclient"
	"github.com/kyokan/drawbridge/pkg/crypto"
)

type FundingManager struct {
	peerBook *p2p.PeerBook
	db       *db.DB
	km       *wallet.KeyManager
	client   *ethclient.Client
	config   *pkg.Config
}

func NewFundingManager(peerBook *p2p.PeerBook, db *db.DB, km *wallet.KeyManager, client *ethclient.Client, config *pkg.Config) *FundingManager {
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
	return nil, nil
}

func (m *FundingManager) onAcceptChannel(msg *lnwire.AcceptChannel) (*lnwire.FundingCreated, error) {
	return nil, nil
}

func (m *FundingManager) onFundingCreated(msg *lnwire.FundingCreated) (*lnwire.FundingSigned, error) {
	return nil, nil
}

func (m *FundingManager) onFundingSigned(msg *lnwire.FundingSigned) (*lnwire.FundingLocked, error) {
	return nil, nil
}