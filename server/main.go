package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	pb "microservice_go/remote-build"
	"net"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedBuildServiceServer
	producer *kafka.Producer
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

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	producer := NewKafkaProducer()
	defer producer.Close()

	s := grpc.NewServer()
	pb.RegisterBuildServiceServer(s, &server{producer: producer})
	log.Printf("Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
