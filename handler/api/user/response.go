package user

type loginUserResponse struct {
	Message string `json:"message"`
	Data    struct {
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		Name        string `json:"name"`
		AccessToken string `json:"accessToken"`
	} `json:"data"`
}

type createUserResponse struct {
	Message string `json:"message"`
	Data    struct {
		Phone       string `json:"phone,omitempty" validate:"min=7,max13"`
		Email       string `json:"email,omitempty" validate:"email"`
		Name        string `json:"name" validate:"required,min=5,max=50"`
		AccessToken string `json:"accessToken" validate:"required,min=5,max=15"`
	} `json:"data"`
}
