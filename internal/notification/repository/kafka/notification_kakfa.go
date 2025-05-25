package kafka

import (
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lightlink/notification-service/internal/notification/domain/dto"
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
		log.Println("Схема не найдена, создаём новую...")

		schemaStr := `{
			"type": "record",
			"name": "RawNotification",
			"fields": [
				{"name": "type", "type": "string"},
				{"name": "payload", "type": [
					{"type": "record", "name": "FriendRequestPayload", "fields": [
						{"name": "from_user_id", "type": "string"},
						{"name": "to_user_id", "type": "string"}
					]},
					{"type": "record", "name": "IncomingMessagePayload", "fields": [
						{"name": "from_user_id", "type": "string"},
						{"name": "to_user_id", "type": "string"},
						{"name": "room_id", "type": "string"},
						{"name": "content", "type": "string"}
					]},
					{"type": "record", "name": "IncomingCallPayload", "fields": [
						{"name": "from_user_id", "type": "string"},
						{"name": "to_user_id", "type": "string"},
						{"name": "room_id", "type": "string"}
					]}
				]}
			]
		}`

		schema, err = schemaRegistryClient.CreateSchema(subject, schemaStr, srclient.Avro)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания схемы в Registry: %v", err)
		}
		log.Println("Схема успешно создана!")
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
			timeoutDuration := time.Second * 5

			msg, err := repo.consumer.ReadMessage(timeoutDuration)
			if err != nil {
				kafkaErr := err.(kafka.Error)
				if kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("Ошибка при чтении сообщения: %v", err)
				continue
			}

			native, _, err := repo.codec.NativeFromBinary(msg.Value)
			if err != nil {
				log.Printf("Ошибка декодирования Avro: %v", err)
				continue
			}

			avroData := native.(map[string]interface{})
			rawPayload := avroData["payload"].(map[string]interface{})

			var realPayload map[string]interface{}
			var payloadType string

			for key, val := range rawPayload {
				payloadType = key
				realPayload = val.(map[string]interface{})
				break
			}

			notification := dto.RawNotification{
				Type:    payloadType,
				Payload: realPayload,
			}

			fmt.Printf("KAFKA: Get value from queue: %v\n", notification)
			ch <- notification
		}
	}()

	return ch, nil
}
