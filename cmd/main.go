package main

import (
	"fmt"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"tages/config"
	ctr "tages/internal/controller/grpc"
	"tages/pb"

	"google.golang.org/grpc"
)

const (
	maxUploadConnections = 10
	maxListConnections   = 100
)

func main() {
	cfg := config.LoadConfig()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.GRPCServerPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterFileServiceServer(grpcServer, ctr.NewFileServer(
		maxUploadConnections, maxListConnections))

	reflection.Register(grpcServer)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("gRPC server is running on port :%v", cfg.GRPCServerPort)

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-stop
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}
