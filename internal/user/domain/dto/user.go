package dto

import proto "github.com/lightlink/notification-service/protogen/user"

type UserTransfer struct {
	Id       uint
	Username string
}

func GetUserResponseToTransfer(getResponse *proto.GetUserResponse) *UserTransfer {
	return &UserTransfer{
		Id:       uint(getResponse.Id),
		Username: getResponse.Username,
	}
}
