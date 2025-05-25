package centrifugo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lightlink/notification-service/internal/notification/usecase"
)

type PublishResponse struct {
	Error  *PublishErrorResponse   `json:"error,omitempty"`
	Result *PublishSuccessResponse `json:"result,omitempty"`
}

type PublishErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PublishSuccessResponse struct {
	Offset int    `json:"offset"`
	Epoch  string `json:"epoch"`
}

type NotificationDeliveryI interface {
	Publish(channel string, data interface{}) error
	ConsumeMessages() error
}

type CentrifugoClient struct {
	notificationUsecase usecase.NotificationUsecaseI
	httpClient          *http.Client
	apiURL              string
	apiKey              string
}

func NewCentrifugoClient(apiURL, apiKey string, notificationUsecase usecase.NotificationUsecaseI) *CentrifugoClient {
	return &CentrifugoClient{
		httpClient:          &http.Client{Timeout: 5 * time.Second},
		apiURL:              apiURL,
		apiKey:              apiKey,
		notificationUsecase: notificationUsecase,
	}
}

func (c *CentrifugoClient) ConsumeMessages() error {
	inputCh, err := c.notificationUsecase.ConsumeMessages()
	if err != nil {
		return err
	}

	for message := range inputCh {
		log.Printf("Publishing notification: %v\n", message)
		err := c.Publish(message.Channel, message.Payload)
		if err != nil {
			log.Printf("ERR: err publishing message in centrifugo: %v\n", err)
			continue
		}
	}

	return nil
}

func (c *CentrifugoClient) Publish(channel string, data interface{}) error {
	payload := map[string]interface{}{
		"channel": channel,
		"data":    data,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.apiURL+"/api/publish", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("centrifugo error: %s", resp.Status)
	}

	var publishResponse PublishResponse
	err = json.NewDecoder(resp.Body).Decode(&publishResponse)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if publishResponse.Error != nil {
		return fmt.Errorf("error: Code %d, Message: %s", publishResponse.Error.Code, publishResponse.Error.Message)
	}

	fmt.Println("Successfully published in channel")

	return nil
}
