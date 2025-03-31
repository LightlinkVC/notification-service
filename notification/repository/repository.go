package repository

import "github.com/lightlink/notification-service/notification/domain/dto"

type Notifier interface {
	Receive() (<-chan dto.RawNotification, error)
}
