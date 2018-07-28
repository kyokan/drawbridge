package api

import (
	"github.com/kyokan/drawbridge/internal/ethclient"
	"net/http"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/kyokan/drawbridge/internal/protocol"
)

var fsLog *zap.SugaredLogger

func init() {
	fsLog = logger.Logger.Named("funding-service")
}

type FundingService struct {
	client      *ethclient.Client
	chanHandler *protocol.ChannelHandler
}

func NewFundingService(client *ethclient.Client, chanHandler *protocol.ChannelHandler) (*FundingService) {
	return &FundingService{
		client:      client,
		chanHandler: chanHandler,
	}
}

type ApproveArgs struct {
	Amount string
}

type ApproveReply struct {
	TxHash string
	Status string
}

func (f *FundingService) Approve(r *http.Request, args *ApproveArgs, reply *ApproveReply) error {
	fsLog.Infow("received approve request",
		"amount", args.Amount,
	)

	amountBig, err := hexutil.DecodeBig(args.Amount)

	if err != nil {
		return err
	}

	tx, err := f.client.ApproveERC20(amountBig)

	if err != nil {
		return err
	}

	txHash := tx.Hash()
	reply.TxHash = hexutil.Encode(txHash[:])
	reply.Status = StatusOk
	fsLog.Infow("processed approve request",
		"amount", args.Amount,
		"txHash", reply.TxHash,
	)
	return nil
}

type DepositArgs struct {
	Amount string
}

type DepositReply struct {
	TxHash string
	Status string
}

func (f *FundingService) Deposit(r *http.Request, args *DepositArgs, reply *DepositReply) error {
	fsLog.Infow("received deposit request",
		"amount", args.Amount,
	)

	tokensBig, err := hexutil.DecodeBig(args.Amount)

	if err != nil {
		fsLog.Errorw("decoding failure",
			"error", err.Error(),
		)
		return err
	}

	tx, err := f.client.Deposit(tokensBig)

	if err != nil {
		return err
	}

	txHash := tx.Hash()
	reply.TxHash = hexutil.Encode(txHash[:])
	reply.Status = StatusOk
	fsLog.Infow("processed deposit request",
		"amount", args.Amount,
		"txHash", reply.TxHash,
	)
	return nil
}

type OpenChannelArgs struct {
	PeerPubkey string
	Amount     string
}

type OpenChannelReply struct {
	Status string
}

func (f *FundingService) OpenChannel(r *http.Request, args *OpenChannelArgs, reply *OpenChannelReply) error {
	fsLog.Infow("received open channel request",
		"peerId", args.PeerPubkey,
		"amount", args.Amount,
	)

	amountBig, err := hexutil.DecodeBig(args.Amount)

	if err != nil {
		fsLog.Errorw("decoding failure",
			"error", err.Error(),
		)
	}

	pub, err := crypto.PublicFromCompressedHex(args.PeerPubkey)

	if err != nil {
		return err
	}

	err = f.chanHandler.InitChannel(pub, amountBig)

	if err != nil {
		return err
	}

	reply.Status = StatusOk

	return nil
}
