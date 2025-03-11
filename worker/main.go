package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
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
	log.Printf("Received Command: %v", serverTask.Command)
	log.Printf("Received File: %v", serverTask.File)
	execCommand := exec.Command(serverTask.Command, serverTask.File, "-o", "main.o")
	log.Printf("Command: %v", execCommand)
	err := execCommand.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	// execCommand.Output()
	return &pb.WorkerResponse{CompleteTask: "successfully convert input file " + serverTask.File + " to main.o"}, nil
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
