package entity

import "fmt"

func PersonalChannel(userID string) string {
	return fmt.Sprintf("personal:%s", userID)
}
