package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"log"
	"os"
	"service/config"
	msg "service/gen/go"
)

/*
	auto.offset.reset : latest // Bắt đầu lấy message gửi đi sau khi consumer bắt đầu chạy
						earliest // Bắt đầu lấy message có offset được gửi đầu tiên
*/

func main() {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost",
		"group.id":           "myGroup",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	})

	if err != nil {
		panic(err)
	}
	client, err := schemaregistry.NewClient(schemaregistry.NewConfig(config.URLSchema))
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	deser, err := protobuf.NewDeserializer(client, serde.ValueSerde, protobuf.NewDeserializerConfig())
	if err != nil {
		log.Fatalf("Error creating deserialize: %v", err)
	}
	deser.ProtoRegistry.RegisterMessage((&msg.Message{}).ProtoReflect().Type())

	err = c.SubscribeTopics([]string{config.Topic}, nil)

	if err != nil {
		panic(err)
	}

	run := true

	for run {
		ev := c.Poll(100)
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case *kafka.Message:
			value, err := deser.Deserialize(*e.TopicPartition.Topic, e.Value)
			if err != nil {
				fmt.Printf("Failed to deserialize payload: %s\n", err)
				continue
			}

			fmt.Printf("%% Message on %s:\n%+v\n", e.TopicPartition, value)

			if e.Headers != nil {
				fmt.Printf("%% Headers: %v\n", e.Headers)
			}
			c.CommitMessage(e)

		case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
		default:
			fmt.Printf("Ignored %v\n", e)
		}
	}
	fmt.Printf("Closing consumer\n")
	c.Close()
}
