package repository

import "github.com/lightlink/notification-service/internal/notification/domain/dto"

type NotificationRepositoryI interface {
	Receive() (<-chan dto.RawNotification, error)
}
