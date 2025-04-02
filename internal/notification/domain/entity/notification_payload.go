package entity

type FriendRequestPayload struct {
	Type         string `json:"type"`
	FromUsername string `json:"from_username"`
	FromUserID   string `json:"from_user_id"`
	ToUserID     string `json:"to_user_id"`
}

type IncomingMessagePayload struct {
	Type       string `json:"type"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	RoomID     string `json:"room_id"`
	Content    string `json:"content"`
}

type IncomingCallPayload struct {
	Type       string `json:"type"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	RoomID     string `json:"room_id"`
}
