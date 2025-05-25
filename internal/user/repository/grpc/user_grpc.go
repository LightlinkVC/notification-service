package grpc

import (
	"context"

	"github.com/lightlink/notification-service/internal/user/domain/dto"
	proto "github.com/lightlink/notification-service/protogen/user"
)

type UserGrpcRepository struct {
	client proto.UserServiceClient
}

func NewUserGrpcRepository(client *proto.UserServiceClient) *UserGrpcRepository {
	return &UserGrpcRepository{
		client: *client,
	}
}

func (repo *UserGrpcRepository) GetById(id uint) (*dto.UserTransfer, error) {
	getUserByIdRequest := &proto.GetUserByIdRequest{
		Id: uint32(id),
	}

	userResponseProto, err := repo.client.GetUserById(context.Background(), getUserByIdRequest)
	if err != nil {
		return nil, err
	}

	userModel := dto.GetUserResponseToTransfer(userResponseProto)

	return userModel, nil
}
