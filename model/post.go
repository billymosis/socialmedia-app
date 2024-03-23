package model

import (
	"database/sql"
	"time"

)

type Post struct {
	Id        int
	Html      string
	UserId    int
	Tags      []string
	CreatedAt time.Time
}

type Comment struct {
	Id        int
	Comment   string
	PostId    int
	UserId    int
	CreatedAt time.Time
}

type PostData struct {
	PostInHTML string    `json:"postInHtml"`
	Tags       []string  `json:"tags"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CommentResponse struct {
	Comment   sql.NullString `json:"comment"`
	Creator   Creator        `json:"creator"`
	CreatedAt sql.NullTime   `json:"createdAt"`
}

type CommentResponseValid struct {
	Comment   string       `json:"comment"`
	Creator   CreatorValid `json:"creator"`
	CreatedAt time.Time    `json:"createdAt"`
}

type PostResponseData struct {
	PostID      string                 `json:"postId"`
	PostContent PostData               `json:"post"`
	Comments    []CommentResponseValid `json:"comments"`
	Creator     CreatorValid           `json:"creator"`
}

type Creator struct {
	UserID      sql.NullInt32  `json:"userId"`
	Name        sql.NullString `json:"name"`
	ImageURL    sql.NullString `json:"imageUrl"`
	FriendCount sql.NullInt32  `json:"friendCount"`
	CreatedAt   sql.NullTime   `json:"createdAt"`
}

type CreatorValid struct {
	UserId      int       `json:"userId"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"imageUrl"`
	FriendCount int       `json:"friendCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CommentAndUser struct {
	Comment Comment
	Creator CreatorValid
}

type PostResponse struct {
	Message string             `json:"message"`
	Data    []PostResponseData `json:"data"`
	Meta    Meta               `json:"meta"`
}
