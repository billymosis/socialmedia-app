package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type jwtCustomClaims struct {
	UserId int `json:"user_id"`
	jwt.StandardClaims
}

func GenerateToken(id int) (string, error) {
	now := time.Now()
	var expiration time.Time
	environment := os.Getenv("ENVIRONMENT")
	if environment == "production" {
		expiration = now.Add(time.Minute * 2)
	} else {
		expiration = now.Add(time.Hour * 1)
	}
	claims := &jwtCustomClaims{
		UserId: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return t, nil
}

func GetUserId(ctx context.Context) (int, error) {
	props, _ := ctx.Value("userAuthCtx").(jwt.MapClaims)

	userId, err := strconv.Atoi(fmt.Sprintf("%v", props["user_id"]))
	if err != nil {
		return 0, err
	}

	return userId, nil
}
