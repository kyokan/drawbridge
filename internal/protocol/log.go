package protocol

import (
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
)

var log *zap.SugaredLogger

func init() {
	log = logger.Logger.Named("protocol")
}