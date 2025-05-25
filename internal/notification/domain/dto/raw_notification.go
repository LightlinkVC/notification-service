package dto

type RawNotification struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type ReadyNotification struct {
	Channel string `json:"channel"`
	Payload []byte `json:"payload"`
}
