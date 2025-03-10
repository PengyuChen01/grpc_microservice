package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "microservice_go/remote-build"
)

var (
	port = flag.Int("port", 50052, "The server port")

)

type server struct {
	pb.UnimplementedWorkServiceServer
}

func (s *server) AssignTask(Message context.Context, serverTask *pb.TaskRequest) (*pb.WorkerResponse, error) {
	log.Printf("Received: %v", serverTask)
	return &pb.WorkerResponse{CompleteTask: " Complete" + serverTask.GetTask()}, nil
}


func main() {
	flag.Parse()
	// Set up a connection to the server.
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWorkServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	

}
