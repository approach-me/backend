package lasn

import (
	"context"
	"log"

	"github.com/approach.me/backend/protos"
)

type Service struct {
	protos.UnimplementedLasnServiceServer
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Link(ctx context.Context, request *protos.LinkRequest) (*protos.LinkResponse, error) {
	log.Printf("Received request from user [%v]", request.UserId)
	return &protos.LinkResponse{}, nil
}

func (s *Service) Fetch(ctx context.Context, request *protos.FetchRequest) (*protos.FetchResponse, error) {
	return &protos.FetchResponse{}, nil
}

func (s *Service) Subscribe(request *protos.SubscribeRequest, stream protos.LasnService_SubscribeServer) error {
	return stream.Send(&protos.SubscribeResponse{})
}
