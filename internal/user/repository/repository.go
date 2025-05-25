package repository

import "github.com/lightlink/notification-service/internal/user/domain/dto"

type UserRepositoryI interface {
	GetById(id uint) (*dto.UserTransfer, error)
}
