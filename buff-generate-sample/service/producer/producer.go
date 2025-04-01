package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"log"
	"service/config"
	msg "service/gen/go"
)

func main() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": config.Host})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(config.URLSchema))

	if err != nil {
		log.Fatalf("Error schemaregistry.NewClient : %v", err)
	}

	ser, err := protobuf.NewSerializer(client, serde.ValueSerde, protobuf.NewSerializerConfig())

	if err != nil {
		log.Fatalf("Error protobuf.NewSerializer : %v", err)
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
		value := msg.Message{
			Content: word,
		}

		payload, err := ser.Serialize(config.Topic, &value)
		if err != nil {
			log.Fatalf("Error serializing %v : %v", config.Topic, err)
		}

		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &config.Topic, Partition: kafka.PartitionAny},
			Value:          payload,
		}, nil)
	}

	p.Flush(15 * 1000)
}
