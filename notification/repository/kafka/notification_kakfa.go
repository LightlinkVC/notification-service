package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lightlink/notification-service/notification/domain/dto"
	"github.com/linkedin/goavro"
	"github.com/riferrei/srclient"
)

type NotificationKafkaRepository struct {
	consumer *kafka.Consumer
	codec    *goavro.Codec
	topic    string
}

func NewNotificationKafkaRepository(brokers, groupID, topic, schemaRegistryURL string) (*NotificationKafkaRepository, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	schemaRegistryClient := srclient.CreateSchemaRegistryClient(schemaRegistryURL)
	subject := topic + "-value"
	schema, err := schemaRegistryClient.GetLatestSchema(subject)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения схемы: %v", err)
	}

	codec, err := goavro.NewCodec(schema.Schema())
	if err != nil {
		return nil, fmt.Errorf("ошибка создания Avro-кодека: %v", err)
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, err
	}

	return &NotificationKafkaRepository{
		consumer: consumer,
		codec:    codec,
		topic:    topic,
	}, nil
}

func (repo *NotificationKafkaRepository) Receive() (<-chan dto.RawNotification, error) {
	ch := make(chan dto.RawNotification)

	go func() {
		defer close(ch)

		for {
			/*TODO deadline*/
			msg, err := repo.consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("Ошибка при чтении сообщения: %v", err)
				continue
			}

			native, _, err := repo.codec.NativeFromBinary(msg.Value)
			if err != nil {
				log.Printf("Ошибка декодирования Avro: %v", err)
				continue
			}

			jsonData, err := json.Marshal(native)
			if err != nil {
				log.Printf("Ошибка преобразования Avro в JSON: %v", err)
				continue
			}

			var notification dto.RawNotification
			err = json.Unmarshal(jsonData, &notification)
			if err != nil {
				log.Printf("Ошибка парсинга JSON: %v", err)
				continue
			}

			ch <- notification
		}
	}()

	return ch, nil
}
