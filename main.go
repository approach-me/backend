package main

import (
	"log"
	"net"

	"github.com/approach.me/backend/managers"
	"github.com/approach.me/backend/protos"
	"github.com/approach.me/backend/services/lasn"
	"google.golang.org/grpc"
)

var neo4jManager *managers.Neo4jManager

func init() {
	neo4jManager = managers.NewNeo4jManager()
}

const host = "0.0.0.0:9090"

func main() {
	defer neo4jManager.Close()

	list, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	protos.RegisterLasnServiceServer(grpcServer, lasn.NewService(neo4jManager))

	log.Printf("Listening on %v...", host)
	grpcServer.Serve(list)
}
