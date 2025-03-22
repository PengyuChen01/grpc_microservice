package main

import (
	"context"
	// "encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	pb "microservice_go/remote-build"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// Server struct
type server struct {
	pb.UnimplementedBuildServiceServer
	producer    *kafka.Producer
	resultStore map[string]*pb.ResultResponse // Store compiled files
	mu          sync.Mutex                    // Mutex for concurrent access
}

// 创建 Kafka 生产者
func NewKafkaProducer() *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	return p
}

// 发送任务到 Kafka
func (s *server) SendRequest(_ context.Context, clientTask *pb.ClientRequest) (*pb.ServerResponse, error) {
	log.Printf("Received Command: %v", clientTask.Command)

	topic := "task-queue"
	message := fmt.Sprintf("%s|%s|%s", clientTask.Command, clientTask.File, clientTask.FileContent)
	err := s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)
	if err != nil {
		log.Printf("Kafka Produce error: %v", err)
		return nil, err
	}

	log.Printf("Task pushed to Kafka: %s", message)
	return &pb.ServerResponse{ServerResponse: "Task sent to Kafka"}, nil
}

// 监听 Kafka `result-queue` 并存储结果
func (s *server) listenForResults() {
	topic := "result-queue"

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "server-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	err = consumer.Subscribe(topic, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	log.Println("Server listening for task results...")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Kafka error: %v", err)
			continue
		}

		// Extract file name from Kafka Key
		fileName := string(msg.Key)
		fileData := string(msg.Value)

		// Debugging log
		log.Printf("Received from Kafka - File: %s, Data: %s", fileName, fileData)

		// Split fileContent and status from Value
		data := strings.Split(fileData, "|")
		if len(data) != 2 {
			log.Println("Invalid result format")
			continue
		}

		fileContent, status := data[0], data[1]

		// Only store successful compilations
		if status == "Success" {
			x := strings.TrimSuffix(fileName, ".c") + ".o"
			log.Printf("Stored result: %s", x)

			s.mu.Lock()
			s.resultStore[fileName] = &pb.ResultResponse{
				FileName:    fileName,
				FileContent: fileContent, // Send base64 content
			}
			s.mu.Unlock()
		}
	}
}

// 让客户端获取结果
func (s *server) GetResult(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, exists := s.resultStore[req.FileName]
	if !exists {
		return nil, fmt.Errorf("No result found for file: %s", req.FileName)
	}

	return result, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	producer := NewKafkaProducer()
	defer producer.Close()

	s := grpc.NewServer()
	serverInstance := &server{
		producer:    producer,
		resultStore: make(map[string]*pb.ResultResponse),
	}
	pb.RegisterBuildServiceServer(s, serverInstance)
	log.Printf("Server listening at %v", lis.Addr())

	go serverInstance.listenForResults()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
