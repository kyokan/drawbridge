package api

import (
		"net/http"
		"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
	"github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/kyokan/drawbridge/internal/protocol"
	"math/big"
)

var csLog *zap.SugaredLogger

func init() {
	csLog = logger.Logger.Named("channel-service")
}

type SwapService struct {
	swapHandler *protocol.SwapHandler
}

func NewSwapService(swapHandler *protocol.SwapHandler) (*SwapService) {
	return &SwapService{
		swapHandler: swapHandler,
	}
}

type DoSwapArgs struct {
	PeerPubkey string
}

type DoSwapReply struct {
	Status string
}

func (f *SwapService) DoSwap(r *http.Request, args *DoSwapArgs, reply *DoSwapReply) error {
	csLog.Infow("performing swap")

	pub, err := crypto.PublicFromCompressedHex(args.PeerPubkey)

	if err != nil {
		return err
	}

	err = f.swapHandler.InitSwap(pub, big.NewInt(1000), big.NewInt(1000))
	if err != nil {
		return err
	}

	reply.Status = StatusOk
	return nil
}
