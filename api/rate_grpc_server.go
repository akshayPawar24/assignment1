package api

import (
	ratepb "assignment1/grpc/proto"
	"assignment1/service"
	"context"
	"log"
)

type RateGRPCServer struct {
	ratepb.UnimplementedRateServiceServer
	Service *service.RateService
}

func NewRateGRPCServer(svc *service.RateService) *RateGRPCServer {
	return &RateGRPCServer{Service: svc}
}

func (s *RateGRPCServer) GetRate(ctx context.Context, req *ratepb.GetRateRequest) (*ratepb.GetRateResponse, error) {
	base := req.GetBase()
	target := req.GetTarget()
	resp := &ratepb.GetRateResponse{}

	if base == "" || target == "" {
		resp.Error = "Missing base or target parameter"
		return resp, nil
	}

	// Add this log
	log.Printf("GetRate called with base=%s, target=%s", base, target)

	data, err := s.Service.GetRate(base, target)

	// Add this log
	log.Printf("GetRate finished: err=%v, data=%+v", err, data)

	if err != nil {
		resp.Error = err.Error()
		return resp, nil
	}

	resp.Base = data.Base
	resp.Target = data.Target
	resp.Rate = data.Rate
	resp.UpdatedAt = data.UpdatedAt
	return resp, nil
}
