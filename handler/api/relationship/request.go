package relationship

type addFriendRequest struct {
	UserId string `json:"userId" validate:"required"`
}
