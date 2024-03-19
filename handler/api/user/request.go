package user

type loginUserRequest struct {
	Password        string `json:"password" validate:"required,min=5,max=15"`
	CredentialType  string `json:"credentialType" validate:"required,oneof=phone email"`
	CredentialValue string `json:"credentialValue" validate:"required"`
}

type createUserRequest struct {
	Name            string `json:"name" validate:"required,min=5,max=50"`
	Password        string `json:"password" validate:"required,min=5,max=15"`
	CredentialType  string `json:"credentialType" validate:"required,oneof=phone email"`
	CredentialValue string `json:"credentialValue" validate:"required"`
}

type linkEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type linkPhoneRequest struct {
	Phone string `json:"phone" validate:"required,startswith=+,min=7,max=13"`
}
