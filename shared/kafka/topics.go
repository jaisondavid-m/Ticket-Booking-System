package kafka

import (
	"log"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func CreateTopics(broker string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()
	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	controllerConn,err:= kafka.Dial("tcp",net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()
	topics := []kafka.TopicConfig{
		{
			Topic:             TopicBookingRequests,
			NumPartitions:     12,  // 12 partitions
			ReplicationFactor: 3,
		},
		{
			Topic:             TopicBookingResults,
			NumPartitions:     12,
			ReplicationFactor: 3,
		},
		{
			Topic:             TopicInventorySync,
			NumPartitions:     6,
			ReplicationFactor: 3,
		},
		{
			Topic:             TopicBookingDLQ,
			NumPartitions:     3,
			ReplicationFactor: 3,
		},
	}
	err = controllerConn.CreateTopics(topics...)
	if err != nil {
		log.Printf("[Kafka] topic creation (may already exist): %v", err)
	}
	log.Println("[Kafka] topics ensured")
	return nil
}