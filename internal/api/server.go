package api

import (
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net/http"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
)

const StatusOk = "OK"

var sLog *zap.SugaredLogger

func init() {
	sLog = logger.Logger.Named("api-server")
}

func Start(container *ServiceContainer, addr string, port string) {
	sLog.Infow("starting services", "listen-ip", addr, "listen-port", port)
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	container.RegisterServices(s)
	http.Handle("/rpc", s)
	err := http.ListenAndServe(addr + ":" + port, s)

	if err != nil {
		sLog.Fatalw("failed to start HTTP listener", "err", err.Error())
	}
}