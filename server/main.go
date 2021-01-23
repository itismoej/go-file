package main

import (
	"github.com/mjafari98/go-file/models"
	"github.com/mjafari98/go-file/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

const GRPCPort = ":50061"

var DB = models.ConnectAndMigrate()

func main() {
	fileServer := FilesServer{}

	// start gRPC server
	listener, err := net.Listen("tcp", GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterFilesServer(grpcServer, &fileServer)
	reflection.Register(grpcServer)

	log.Printf("server gRPC is starting in localhost%s ...\n", GRPCPort)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start GRPC server: ", err)
	}
	// end of gRPC server
}
