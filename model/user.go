package model

import (
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	devSaltRounds  = 8
	prodSaltRounds = 10
)

type User struct {
	Id          int
	Name        string
	Password    string
	ImageUrl    string
	CreatedAt   time.Time
	FriendCount int
}

type UserAndCred struct {
	Id          int
	Name        string
	Password    string
	ImageUrl    string
	CreatedAt   time.Time
	FriendCount int
	Email       string
	Phone       string
}

func (user *User) HashPassword() error {
	saltRound, err := strconv.Atoi(os.Getenv("BCRYPT_SALT"))
	if err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), saltRound)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return nil
}

func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (user *UserAndCred) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
