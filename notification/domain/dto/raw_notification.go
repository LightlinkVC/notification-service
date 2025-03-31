package dto

type RawNotification struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}
