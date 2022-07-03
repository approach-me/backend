package main

import (
	"log"
	"net"

	"github.com/approach.me/backend/protos"
	"github.com/approach.me/backend/services/lasn"
	"google.golang.org/grpc"
)

const host = "0.0.0.0:9090"

func main() {
	list, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	protos.RegisterLasnServiceServer(grpcServer, lasn.NewService())

	log.Printf("Listening on %v...", host)
	grpcServer.Serve(list)
}
