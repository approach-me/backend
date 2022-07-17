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
	stream   protos.LasnService_SubscribeServer // stream is the server side of the RPC stream
	finished chan<- bool                        // finished is used to signal closure of a client subscription
}

func NewService(neo4jManager *managers.Neo4jManager) *Service {
	return &Service{neo4jManager: neo4jManager}
}

func (s *Service) CreateOrUpdateNode(ctx context.Context, request *protos.CreateOrUpdateNodeRequest) (*protos.CreateOrUpdateNodeResponse, error) {
	log.Printf("Received CreateOrUpdateNode request [%v]", request)
	err := s.neo4jManager.CreateOrUpdateNode(request.UserSummary)
	if err != nil {
		return nil, err
	}
	return &protos.CreateOrUpdateNodeResponse{}, nil
}

func (s *Service) Link(ctx context.Context, request *protos.LinkRequest) (*protos.LinkResponse, error) {
	log.Printf("Received Link request [%v]", request)

	err := s.neo4jManager.AddEdgeBetween(request.UserId, request.NearbyDeviceId)
	if err != nil {
		return nil, err
	}

	userSummaries, err := s.neo4jManager.RetrieveUsersConnectedTo(request.UserId)
	if err != nil {
		log.Printf("Failed to notify users about link. Reason: %v\n", err)
		return &protos.LinkResponse{}, nil
	}

	s.notifyUsersWith(userSummaries)
	return &protos.LinkResponse{}, nil
}

func (s *Service) Fetch(ctx context.Context, request *protos.FetchRequest) (*protos.FetchResponse, error) {
	userSummaries, err := s.neo4jManager.RetrieveUsersConnectedTo(request.UserId)
	if err != nil {
		log.Printf("Error while collecting results: [%v]", err)
		return nil, err
	}
	log.Printf("Fetch return %v records", len(userSummaries))
	return &protos.FetchResponse{UserSummaries: userSummaries}, nil
}

func (s *Service) Disconnect(ctx context.Context, request *protos.DisconnectRequest) (*protos.DisconnectResponse, error) {
	err := s.neo4jManager.UnlinkEdgesAndMergeDanglingComponentsFrom(request.UserId)
	if err != nil {
		return nil, err
	}

	v, ok := s.subscribers.Load(request.UserId)
	if !ok {
		// User is already not subscribed so just return.
		return &protos.DisconnectResponse{}, nil
	}

	select {
	case v.(subscription).finished <- true:
		log.Printf("Unsubscribed client: %v", request.UserId)
	default:
		// Default case is to avoid blocking in case client has already unsubscribed
	}
	s.subscribers.Delete(request.UserId)
	return &protos.DisconnectResponse{}, nil
}

func (s *Service) Subscribe(request *protos.SubscribeRequest, stream protos.LasnService_SubscribeServer) error {
	log.Printf("Received Subscribe request from UserID: [%v]", request.UserId)

	finished := make(chan bool)
	// Save the subscription according to the given UserID
	s.subscribers.Store(request.UserId, subscription{stream: stream, finished: finished})

	ctx := stream.Context()
	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-finished:
			log.Printf("UserID [%v] has disconnected", request.UserId)
			return nil
		case <-ctx.Done():
			log.Printf("UserID [%v] connection has closed", request.UserId)
			s.subscribers.Delete(request.UserId)
			return nil
		}
	}
}

func (s *Service) notifyUsersWith(userSummaries []*protos.UserSummary) {
	log.Println("Going to notify subscribed users")
	for _, userSummary := range userSummaries {
		v, ok := s.subscribers.Load(userSummary.UserId)
		if !ok {
			// That user does not have an active subscription (not active in the app).
			// TODO: send phone notification here.
			continue
		}

		log.Printf("Notifying User [%v]", userSummary.UserId)
		err := v.(subscription).stream.Send(&protos.SubscribeResponse{UserSummaries: userSummaries})
		if err != nil {
			log.Printf("Failed to send notification to user stream. Reason: %v", err)
			// TODO: send phone notification here instead.
		}
	}
}
