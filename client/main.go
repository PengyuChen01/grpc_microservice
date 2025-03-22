package main

import (
	"context"
	"encoding/base64"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	pb "microservice_go/remote-build"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultCommand = "gcc"
	defaultFile    = "main.c"
)

var (
	addr    = flag.String("addr", "localhost:50051", "the address to connect to")
	command = flag.String("command", defaultCommand, "Command to execute")
	file    = flag.String("file", defaultFile, "Files to send (separated by space)")
)

func main() {
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBuildServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10) // Increased timeout
	defer cancel()

	eachFile := strings.Split(*file, " ")
	expectedResults := len(eachFile)
	receivedResults := 0
	fileNames := make([]string, 0)

	// Step 1: Send all files first
	for _, filex := range eachFile {
		data, err := os.ReadFile(filex)
		if err != nil {
			log.Printf("Skipping file %s (could not read): %v", filex, err)
			continue
		}

		_, err = c.SendRequest(ctx, &pb.ClientRequest{
			Command:     *command,
			File:        filex,
			FileContent: string(data),
		})


		pb.ClientRequest {
			targets : map [
				string "hello": pb.Command {
					command: "gcc",
					inputs : [
						"main.o",
						"a.o",
						"b.o"
						"zhao.o"

					]
					"hello.o": {
						"command": "gcc",
						"input": ["hello.c"], os.ReadFile("hello.c")
								 
					  },
					  "a.o": {
						"command": "gcc",
						"input": ["a.c"], os.ReadFile("a.c")
					  },
					  "b.o": {
						"command": "gcc",
						"input": ["b.c"], os.ReadFile("b.c")
					  },
					  "zhao.o": {
						"command": "gcc",
						"input": ["zhao.c"], os.ReadFile("zhao.c")
					  }
				}
			]
			}
		}

		if err != nil {
			log.Printf("Skipping file %s (SendRequest failed): %v", filex, err)
			continue
		}

		log.Printf("Sent file: %s", filex)
		fileNames = append(fileNames, strings.TrimSuffix(filex, ".c")+".o")
	}

	log.Println("All files sent. Waiting for compilation results...")

	// Step 2: Wait and fetch results for all files
	for receivedResults < expectedResults {
		for _, fileName := range fileNames {
			// Skip if the file already exists
			if _, err := os.Stat(fileName); err == nil {
				log.Printf("File %s already exists, skipping retrieval.", fileName)
				receivedResults++
				continue
			}

			// Request file from server
			res, err := c.GetResult(ctx, &pb.ResultRequest{FileName: fileName})
			if err != nil {

				continue
			}

			// Decode base64 and save file
			binaryData, err := base64.StdEncoding.DecodeString(res.FileContent)
			if err != nil {
				log.Printf("Skipping file %s (Base64 decode failed): %v", fileName, err)
				continue
			}

			err = os.WriteFile(res.FileName, binaryData, 0755) // Set executable permissions
			if err != nil {
				log.Printf("Skipping file %s (File write failed): %v", res.FileName, err)
				continue
			}

			log.Printf("Executable file received and saved: %s", res.FileName)
			receivedResults++
		}
	}

	log.Println("All results received and saved.")
	log.Println("Exiting program.")
}
