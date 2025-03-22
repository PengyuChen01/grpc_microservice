package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// 创建 Kafka 生产者
func NewKafkaProducer() *kafka.Producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	return p
}

// 发送结果到 Kafka
func sendResultToKafka(producer *kafka.Producer, file string, fileContent string, status string) {
	topic := "result-queue"

	message := fileContent + "|" + status

	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:           []byte(file),
		Value:         []byte(message),
	}, nil)
	if err != nil {
		log.Printf("Kafka Produce error: %v", err)
	} else {
		log.Printf("Successfully sent file: %s", file)
	}
}

func main() {
	taskTopic := "task-queue"
	producer := NewKafkaProducer()
	defer producer.Close()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":             "localhost:9092",
		"group.id":                      "worker-group",
		"auto.offset.reset":             "earliest",
		"enable.auto.commit":            "true",
		"partition.assignment.strategy": "range",
	})
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Close()

	err = consumer.Subscribe(taskTopic, nil)
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
		fileName := strings.TrimSuffix(file, ".c")

		// Compile the file
		execCommand := exec.Command(command, file, "-o", fileName+".o")
		log.Printf("Executing: %v", execCommand)
		err = execCommand.Run()
		if err != nil {
			log.Printf("Command failed: %v", err)
			sendResultToKafka(producer, file, "", "Failed")
		} else {
			log.Printf("Successfully compiled %s to %s.o", file, fileName)

			// Read and encode the compiled .o file
			binaryData, err := ioutil.ReadFile(fileName + ".o")
			if err != nil {
				log.Printf("Failed to read compiled file: %v", err)
				sendResultToKafka(producer, file, "", "Failed")
				continue
			}

			base64Encoded := base64.StdEncoding.EncodeToString(binaryData)
			sendResultToKafka(producer, file, base64Encoded, "Success")
		}
	}
}
