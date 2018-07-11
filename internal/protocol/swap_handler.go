package protocol

import (
	"github.com/kyokan/drawbridge/internal/lndclient"
	"github.com/kyokan/drawbridge/internal/ethclient"
	"sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/lightningnetwork/lnd/lnwire"
	"math/big"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/kyokan/drawbridge/pkg/wire"
	"errors"
	"github.com/kyokan/drawbridge/pkg/txout"
	"github.com/kyokan/drawbridge/internal/wallet"
	"crypto/sha256"
	"github.com/kyokan/drawbridge/internal/p2p"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type SwapHandler struct {
	peerBook     *p2p.PeerBook
	lnd          *lndclient.Client
	eth          *ethclient.Client
	db           *db.DB
	km           *wallet.KeyManager
	mtx          sync.Mutex
	pendingSwaps map[common.Hash]*pendingSwap
}

type pendingSwap struct {
	SwapID       [32]byte
	PaymentHash  [32]byte
	ETHChannelID common.Hash
	BTCChannelID uint64
	ETHAmount    *big.Int
	BTCAmount    *big.Int
	ETHCommitSig crypto.Signature
	Invoice      *lnrpc.Invoice
	Preimage     [32]byte
}

func NewSwapHandler(pb *p2p.PeerBook, lnd *lndclient.Client, eth *ethclient.Client, d *db.DB, km *wallet.KeyManager) *SwapHandler {
	return &SwapHandler{
		peerBook: pb,
		lnd:lnd,
		eth:eth,
		db: d,
		km:km,
		pendingSwaps: make(map[common.Hash]*pendingSwap),
	}
}

func (s *SwapHandler) InitSwap(pub *crypto.PublicKey, ethAmount *big.Int, btcAmount *big.Int) error {
	peer := s.peerBook.FindPeer(pub)
	if peer == nil {
		return errors.New("peer not found")
	}

	swapId, err := crypto.Rand32()
	if err != nil {
		return err
	}
	preimage, err := crypto.Rand32()
	if err != nil {
		return err
	}
	ethChan, err := s.db.Channels.FindByAmount(ethAmount)
	if err != nil {
		return err
	}
	if ethChan == nil {
		return errors.New("no suitable channel found")
	}
	paymentHash := sha256.Sum256(preimage[:])

	spendReq := &txout.SpendRequest{
		InputID: ethChan.ID,
		Witness: txout.NewMultisigWitness(),
		Values: []*big.Int{
			ethAmount,
		},
		Outputs: []txout.Output{
			&txout.OfferedHTLC{
				Delay:             big.NewInt(5),
				RedemptionAddress: ethChan.Counterparty,
				TimeoutAddress:    s.km.PublicKey().ETHAddress(),
				PaymentHash:       paymentHash,
			},
		},
	}
	sigHash, err := txout.SigData(spendReq)
	if err != nil {
		return err
	}
	sig, err := s.km.SignData(sigHash)
	if err != nil {
		return err
	}

	s.mtx.Lock()
	s.pendingSwaps[swapId] = &pendingSwap{
		SwapID: swapId,
		PaymentHash: paymentHash,
		ETHChannelID: ethChan.ID,
		ETHAmount: ethAmount,
		BTCAmount: btcAmount,
		ETHCommitSig: sig,
		Preimage: preimage,
	}
	s.mtx.Unlock()

	msg := &wire.InitiateSwap{
		SwapID: swapId,
		PaymentHash: paymentHash,
		ETHChannelID: ethChan.ID,
		ETHAmount: ethAmount,
		ETHCommitmentSignature: sig,
		SendingAddress: s.km.PublicKey(),
		RequestedAmount: btcAmount,
	}
	return peer.Send(msg)
}

func (s *SwapHandler) CanAccept(msg lnwire.Message) bool {
	switch msg.MsgType() {
	case wire.MsgInitiateSwap, wire.MsgSwapAccepted, wire.MsgInvoiceGenerated, wire.MsgInvoiceExecuted:
		return true
	default:
		return false
	}
}

func (s *SwapHandler) Accept(envelope *p2p.Envelope) (lnwire.Message, error) {
	msg := envelope.Msg
	switch msg.MsgType() {
	case wire.MsgInitiateSwap:
		return s.onInitiateSwap(msg.(*wire.InitiateSwap), envelope.Peer)
	case wire.MsgSwapAccepted:
		return s.onSwapAccepted(msg.(*wire.SwapAccepted), envelope.Peer)
	case wire.MsgInvoiceGenerated:
		return s.onInvoiceGenerated(msg.(*wire.InvoiceGenerated))
	case wire.MsgInvoiceExecuted:
		return s.onInvoiceExecuted(msg.(*wire.InvoiceExecuted))
	default:
		return nil, errors.New("unknown message type")
	}
}

func (s *SwapHandler) onInitiateSwap(msg *wire.InitiateSwap, peer *p2p.Peer) (*wire.SwapAccepted, error) {
	s.mtx.Lock()
	_, exists := s.pendingSwaps[msg.SwapID]
	if exists {
		defer s.mtx.Unlock()
		return nil, errors.New("duplicate swap id")
	}
	s.mtx.Unlock()

	spendReq := &txout.SpendRequest{
		InputID: msg.ETHChannelID,
		Witness: txout.NewMultisigWitness(),
		Values: []*big.Int{
			msg.ETHAmount,
		},
		Outputs: []txout.Output{
			&txout.OfferedHTLC{
				Delay:             big.NewInt(5),
				RedemptionAddress: s.km.PublicKey().ETHAddress(),
				TimeoutAddress:    msg.SendingAddress.ETHAddress(),
				PaymentHash:       msg.PaymentHash,
			},
		},
	}
	sigHash, err := txout.SigData(spendReq)
	if err != nil {
		return nil, err
	}
	ok := msg.ETHCommitmentSignature.Verify(sigHash, msg.SendingAddress)
	if !ok {
		return nil, errors.New("signature verification failed")
	}

	btcChan, err := s.lnd.ChannelByCounterparty(peer.LNDIdentity)
	if err != nil {
		return nil, err
	}
	if btcChan == nil {
		return nil, errors.New("no suitable lnd channel found")
	}

	s.mtx.Lock()
	s.pendingSwaps[msg.SwapID] = &pendingSwap{
		SwapID: msg.SwapID,
		PaymentHash: msg.PaymentHash,
		ETHChannelID: msg.ETHChannelID,
		ETHAmount: msg.ETHAmount,
		BTCAmount: msg.RequestedAmount,
		ETHCommitSig: msg.ETHCommitmentSignature,
		BTCChannelID: btcChan.ChanId,
	}
	s.mtx.Unlock()

	return &wire.SwapAccepted{
		SwapID: msg.SwapID,
		BTCChannelID: btcChan.ChanId,
	}, nil
}

func (s *SwapHandler) onSwapAccepted(msg *wire.SwapAccepted, peer *p2p.Peer) (*wire.InvoiceGenerated, error) {
	s.mtx.Lock()
	swap, exists := s.pendingSwaps[msg.SwapID]
	if !exists {
		defer s.mtx.Unlock()
		return nil, errors.New("no swap with that ID found")
	}
	s.mtx.Unlock()

	// TODO: check the chan exists on our side

	res, err := s.lnd.AddInvoice(swap.BTCAmount.Int64(), swap.Preimage[:])
	if err != nil {
		return nil, err
	}

	return &wire.InvoiceGenerated{
		SwapID: swap.SwapID,
		PaymentRequest: res.PaymentRequest,
	}, nil
}

func (s *SwapHandler) onInvoiceGenerated(msg *wire.InvoiceGenerated) (*wire.InvoiceExecuted, error) {
	s.mtx.Lock()
	swap, exists := s.pendingSwaps[msg.SwapID]
	if !exists {
		defer s.mtx.Unlock()
		return nil, errors.New("no swap with that ID found")
	}
	s.mtx.Unlock()

	res, err := s.lnd.PayInvoice(msg.PaymentRequest)
	if err != nil {
		return nil, err
	}

	log.Infow("successfully received payment preimage", "preimage", hexutil.Encode(res.PaymentPreimage))

	s.mtx.Lock()
	delete(s.pendingSwaps, msg.SwapID)
	s.mtx.Unlock()

	// TODO: logging and persistence

	return &wire.InvoiceExecuted{
		SwapID: swap.SwapID,
	}, nil
}

func (s *SwapHandler) onInvoiceExecuted(msg *wire.InvoiceExecuted) (lnwire.Message, error) {
	s.mtx.Lock()
	_, exists := s.pendingSwaps[msg.SwapID]
	if !exists {
		defer s.mtx.Unlock()
		return nil, errors.New("no swap with that ID found")
	}
	delete(s.pendingSwaps, msg.SwapID)
	s.mtx.Unlock()
	return nil, nil
}