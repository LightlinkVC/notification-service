package usecase

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"log"

	"github.com/lightlink/notification-service/internal/notification/domain/dto"
	"github.com/lightlink/notification-service/internal/notification/domain/entity"
	notificationRepo "github.com/lightlink/notification-service/internal/notification/repository"
	userRepo "github.com/lightlink/notification-service/internal/user/repository"
)

const (
	FriendRequestNotificationType   = "friendRequest"
	IncomingMessageNotificationType = "incomingMessage"
	IncomingCallNotificationType    = "incomingCall"
)

type NotificationUsecaseI interface {
	ConsumeMessages() (chan dto.ReadyNotification, error)
}

type NotificationUsecase struct {
	notificationRepo notificationRepo.NotificationRepositoryI
	userRepo         userRepo.UserRepositoryI
}

var messageTypeMap = map[string]reflect.Type{
	"FriendRequestPayload":   reflect.TypeOf(entity.FriendRequestPayload{}),
	"IncomingMessagePayload": reflect.TypeOf(entity.IncomingMessagePayload{}),
	"IncomingCallPayload":    reflect.TypeOf(entity.IncomingCallPayload{}),
}

func NewNotificationUsecase(
	notificationRepo notificationRepo.NotificationRepositoryI,
	userRepo userRepo.UserRepositoryI,
) *NotificationUsecase {
	return &NotificationUsecase{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

func (uc *NotificationUsecase) getUsernameByUserID(userID string) (string, error) {
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return "", err
	}

	user, err := uc.userRepo.GetById(uint(userIDUint))
	if err != nil {
		return "", err
	}

	return user.Username, nil
}

func (uc *NotificationUsecase) preparePayload(payload map[string]interface{}, msgType string) (*dto.ReadyNotification, error) {
	structType, exists := messageTypeMap[msgType]
	if !exists {
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}

	structPtr := reflect.New(structType).Interface()

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, structPtr)
	if err != nil {
		return nil, err
	}

	var channel string
	var preparedPayload []byte
	switch v := structPtr.(type) {
	case *entity.FriendRequestPayload:
		channel = entity.PersonalChannel(v.ToUserID)
		v.Type = FriendRequestNotificationType

		senderUsername, err := uc.getUsernameByUserID(v.FromUserID)
		if err != nil {
			return nil, err
		}
		v.FromUsername = senderUsername
	case *entity.IncomingMessagePayload:
		channel = entity.PersonalChannel(v.ToUserID)
		v.Type = IncomingMessageNotificationType

		senderUsername, err := uc.getUsernameByUserID(v.FromUserID)
		if err != nil {
			return nil, err
		}
		v.FromUsername = senderUsername
	case *entity.IncomingCallPayload:
		channel = entity.PersonalChannel(v.ToUserID)
		v.Type = IncomingCallNotificationType

		senderUsername, err := uc.getUsernameByUserID(v.FromUserID)
		if err != nil {
			return nil, err
		}
		v.FromUsername = senderUsername
	default:
		return nil, fmt.Errorf("unknown payload type: %T", v)
	}

	preparedPayload, err = json.Marshal(structPtr)
	if err != nil {
		return nil, err
	}

	return &dto.ReadyNotification{
		Channel: channel,
		Payload: preparedPayload,
	}, nil
}

func (uc *NotificationUsecase) ConsumeMessages() (chan dto.ReadyNotification, error) {
	inputCh, err := uc.notificationRepo.Receive()
	if err != nil {
		return nil, err
	}

	outputCh := make(chan dto.ReadyNotification, 100)

	go func() {
		defer close(outputCh)

		for rawMessage := range inputCh {
			processedNotification, err := uc.preparePayload(rawMessage.Payload, rawMessage.Type)
			if err != nil {
				log.Printf("Ошибка обработки payload: %v\n", err)
				continue
			}
			if processedNotification == nil {
				log.Printf("Неизвестный тип сообщения: %s\n", rawMessage.Type)
				continue
			}

			outputCh <- *processedNotification
		}
	}()

	return outputCh, nil
}
