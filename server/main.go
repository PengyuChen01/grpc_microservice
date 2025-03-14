package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"

	pb "microservice_go/remote-build"
	"net"

	"github.com/confluentinc/confluent-kafka-go/kafka"

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
	numPartitions := 3 // num of partition in task-queue
	partition := int32(rand.Intn(numPartitions))
	err := s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: partition},
		Value:          []byte(message),
	}, nil)
	if err != nil {
		log.Printf("Kafka Produce error: %v", err)
		return nil, err
	}

	log.Printf("Task pushed to Kafka: %s", message)
	return &pb.ServerResponse{ServerResponse: "Task sent to Kafka"}, nil
}

// 监听 Kafka `result-queue` 并打印结果
func listenForResults() {
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

		data := strings.Split(string(msg.Value), "|")
		if len(data) != 2 {
			log.Println("Invalid result format")
			continue
		}

		file, status := data[0], data[1]
		log.Printf("Received task result: %s -> %s", file, status)
	}
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

	// 启动监听 Kafka `result-queue`
	go listenForResults()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
