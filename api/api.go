package api

import (
	"context"

	"github.com/jessfraz/bpfd/api/grpc"
)

type apiServer struct{}

// New returns grpc server instance.
func New() (grpc.APIServer, error) {
	return &apiServer{}, nil
}

func (s *apiServer) CreateRule(ctx context.Context, c *grpc.CreateRuleRequest) (*grpc.CreateRuleResponse, error) {
	return &grpc.CreateRuleResponse{}, nil
}

func (s *apiServer) RemoveRule(ctx context.Context, r *grpc.RemoveRuleRequest) (*grpc.RemoveRuleResponse, error) {
	return &grpc.RemoveRuleResponse{}, nil
}

func (s *apiServer) ListRules(ctx context.Context, r *grpc.ListRulesRequest) (*grpc.ListRulesResponse, error) {
	var rules []*grpc.Rule
	return &grpc.ListRulesResponse{Rules: rules}, nil
}
