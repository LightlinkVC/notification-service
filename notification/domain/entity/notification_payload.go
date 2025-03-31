package entity

type FriendRequestPayload struct {
	FromUserID int `json:"from_user_id"`
}

type IncomingMessagePayload struct {
	FromUserID int    `json:"from_user_id"`
	RoomID     int    `json:"room_id"`
	Content    string `json:"content"`
}

type IncomingCallPayload struct {
	FromUserID int `json:"from_user_id"`
	RoomID     int `json:"room_id"`
}
