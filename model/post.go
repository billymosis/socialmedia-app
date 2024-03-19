package model

import "time"

type Post struct {
	Id     int
	Html   string
	UserId int
	tags   []string
	CreatedAt time.Time
}
