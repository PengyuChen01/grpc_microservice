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

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"log"
	pb "microservice_go/remote-build"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultCommand = "gcc"
	defaultFile = "main.c"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	command = flag.String("command", defaultCommand, "Name to greet")
	file = flag.String("file", defaultFile, "File to send")
)

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBuildServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var eachFile = strings.Split(*file, " ")
	for  i := 0; i < len(eachFile); i++ {
		var filex = eachFile[i]
		data, err := os.ReadFile(filex)
		if err != nil {
			log.Fatalf("could not read file: %v", err)
		}
		r, err := c.SendRequest(ctx, &pb.ClientRequest{Command: *command, File: eachFile[i] ,FileContent: string(data[:])})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("%s",r.ServerResponse) 
	}
	

}
