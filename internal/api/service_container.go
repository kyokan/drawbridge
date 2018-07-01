package api

import "github.com/gorilla/rpc"

type ServiceContainer struct {
	FundingService *FundingService
}

func (s *ServiceContainer) RegisterServices(server *rpc.Server) {
	server.RegisterService(s.FundingService, "")
}