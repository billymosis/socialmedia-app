package relationship

import (
	"time"

	"github.com/billymosis/socialmedia-app/model"
)

type Friend struct {
	UserId      string    `json:"userId"`
	Name        string    `json:"name"`
	ImageUrl    *string   `json:"imageUrl"`
	FriendCount int       `json:"friendCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GetFriendListRow struct {
	Message string    `json:"message"`
	Data    []Friend `json:"data"`
	Meta    model.Meta      `json:"meta"`
}
