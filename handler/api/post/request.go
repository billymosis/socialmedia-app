package request

type createPostRequest struct {
	Html string   `json:"postInHtml" validate:"required,min=2,max=500"`
	Tags []string `json:"tags" validate:"required,min=0"`
}

type createCommentRequest struct {
	PostId  string `json:"postId" validate:"required"`
	Comment string `json:"comment" validate:"required,min=2,max=500"`
}
