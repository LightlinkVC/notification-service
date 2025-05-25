package main

import (
	"fmt"
	"log"
	"os"

	notificationDelivery "github.com/lightlink/notification-service/internal/notification/delivery/centrifugo"
	notificationRepo "github.com/lightlink/notification-service/internal/notification/repository/kafka"
	"github.com/lightlink/notification-service/internal/notification/usecase"
	userRepo "github.com/lightlink/notification-service/internal/user/repository/grpc"
	proto "github.com/lightlink/notification-service/protogen/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	notificationRepo, err := notificationRepo.NewNotificationKafkaRepository(
		"kafka:29092",
		"notification-group",
		"notifications",
		"http://schema_registry:9091",
	)
	if err != nil {
		panic(err)
	}

	client, _ := grpc.Dial(
		fmt.Sprintf("%s:%s", os.Getenv("USER_SERVICE_HOST"), os.Getenv("USER_SERVICE_PORT")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	userServiceClient := proto.NewUserServiceClient(client)
	userRepo := userRepo.NewUserGrpcRepository(&userServiceClient)

	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo, userRepo)

	centrifugoClient := notificationDelivery.NewCentrifugoClient(
		os.Getenv("CENTRIFUGO_HTTP_API_URL"),
		os.Getenv("CENTRIFUGO_HTTP_API_KEY"),
		notificationUsecase,
	)

	log.Println("notification service started")
	log.Fatal(centrifugoClient.ConsumeMessages())
}
