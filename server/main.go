/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	pb "microservice_go/remote-build"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)



var (
	port = flag.Int("port", 50051, "The server port")
	addr = flag.String("addr", "localhost:50052", "the address to connect to")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedBuildServiceServer
}


// SayHello implements helloworld.GreeterServer
func (s *server) SendRequest(_ context.Context, clientTask *pb.ClientRequest) (*pb.ServerResponse, error) {
	log.Printf("Received Command: %v", clientTask.Command)
	
	log.Printf("Received File: %v", clientTask.File)
	log.Printf("Received File Content: %v", clientTask.FileContent)

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewWorkServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.AssignTask(ctx, &pb.TaskRequest{Command: clientTask.Command, File: clientTask.File, FileContent: clientTask.FileContent})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf(" %s", r.GetCompleteTask())
	return &pb.ServerResponse{ServerResponse:  r.GetCompleteTask()}, nil
}


func main() {
	// server set up
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBuildServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}


}
