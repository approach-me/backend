package lasn

import (
	"context"
	"log"
	"sync"

	"github.com/approach.me/backend/managers"
	"github.com/approach.me/backend/protos"
)

type Service struct {
	neo4jManager *managers.Neo4jManager
	subscribers  sync.Map
	protos.UnimplementedLasnServiceServer
}

type subscription struct {
	stream protos.LasnService_SubscribeServer // stream is the server side of the RPC stream
}

func NewService(neo4jManager *managers.Neo4jManager) *Service {
	return &Service{neo4jManager: neo4jManager}
}

func (s *Service) Link(ctx context.Context, request *protos.LinkRequest) (*protos.LinkResponse, error) {
	log.Printf("Received request [%v]", request)

	// TODO: May need to query actual userid from firebase
	err := s.neo4jManager.AddEdgeBetween(request.UserId, request.NearbyUserId)
	if err != nil {
		return nil, err
	}

	userIDs, err := s.neo4jManager.RetrieveUsersConnectedTo(request.UserId)
	if err != nil {
		log.Printf("Failed to notify users about link. Reason: %v\n", err)
		return &protos.LinkResponse{}, nil
	}

	s.notifyUsers(append(userIDs, request.UserId))
	return &protos.LinkResponse{}, nil
}

func (s *Service) Fetch(ctx context.Context, request *protos.FetchRequest) (*protos.FetchResponse, error) {
	userIDs, err := s.neo4jManager.RetrieveUsersConnectedTo(request.UserId)
	if err != nil {
		log.Printf("Error while collecting results: [%v]", err)
		return nil, err
	}
	log.Printf("Fetch return %v records", len(userIDs))
	userSummaries := make([]*protos.UserSummary, len(userIDs))
	for i, userID := range userIDs {
		userSummaries[i] = &protos.UserSummary{UserId: userID}
	}
	return &protos.FetchResponse{UserSummaries: userSummaries}, nil
}

func (s *Service) Subscribe(request *protos.SubscribeRequest, stream protos.LasnService_SubscribeServer) error {
	log.Printf("Received subscribe request from UserID: [%v]", request.UserId)

	// Save the subscriber stream according to the given UserID
	s.subscribers.Store(request.UserId, subscription{stream: stream})
	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-ctx.Done():
			log.Printf("UserID [%v] has disconnected", request.UserId)
			s.subscribers.Delete(request.UserId)
			return nil
		}
	}
}

func (s *Service) notifyUsers(userIDs []int64) {
	log.Println("Going to notify subscribed users")
	for _, userID := range userIDs {
		v, ok := s.subscribers.Load(userID)
		if !ok {
			// That user does not have an active subscription (not active in the app).
			// TODO: send phone notification here.
			log.Printf("UserID [%v] is not subscribed. Ignore...", userID)
			continue
		}

		log.Printf("Notifying UserID: %v", userID)
		err := v.(subscription).stream.Send(&protos.SubscribeResponse{})
		if err != nil {
			log.Printf("Failed to send notification to user stream. Reason: %v", err)
			// TODO: send phone notification here instead.
		}
	}
}
