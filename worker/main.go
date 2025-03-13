package main

import (
	"log"
	"os/exec"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	topic := "task-queue"

	// 创建 Kafka 消费者
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "worker-group",
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

	log.Println("Worker started, waiting for tasks...")

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Kafka error: %v", err)
			continue
		}

		log.Printf("Received task: %s", string(msg.Value))
		data := strings.Split(string(msg.Value), "|")
		if len(data) != 3 {
			log.Println("Invalid task format")
			continue
		}

		command, file:= data[0], data[1]
		wholeFile := strings.Split(file, ".")
		fileName := wholeFile[0]

		execCommand := exec.Command(command, file, "-o", fileName+".o")
		log.Printf("Executing: %v", execCommand)
		err = execCommand.Run()
		if err != nil {
			log.Printf("Command failed: %v", err)
		} else {
			log.Printf("Successfully converted %s to %s.o", file, fileName)
		}
	}
}
