package model

import "time"

type Comments struct {
	Id        int
	comment   string
	UserId    int
	PostId    int
	CreatedAt time.Time
}
