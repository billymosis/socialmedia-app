package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/billymosis/socialmedia-app/handler/render"
	"github.com/billymosis/socialmedia-app/model"
	"github.com/billymosis/socialmedia-app/service/auth"
	us "github.com/billymosis/socialmedia-app/store/user"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func HandleAuthentication(us *us.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginUserRequest

		body, err := io.ReadAll(r.Body)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &req); err != nil {
			render.BadRequest(w, err)
			return
		}

		if err := us.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}
		if req.CredentialType == "phone" {
			err = us.Validate.Var(req.CredentialValue, "required,min=7,max=13,startswith=+")
			if err != nil {
				render.BadRequest(w, errors.New("Invalid phone format"))
				return
			}
		}
		if req.CredentialType == "email" {
			fmt.Printf(req.CredentialValue)
			err = us.Validate.Var(req.CredentialValue, "email,required")
			if err != nil {
				render.BadRequest(w, errors.New("Invalid email format"))
				return
			}

		}
		var user *model.User
		user, err = us.GetByCredential(r.Context(), req.CredentialValue)

		if err != nil {
			render.NotFound(w, errors.New("User not found"))
			logrus.Info("api: cannot find user")
			return
		}

		validUser := user.CheckPassword(req.Password)
		if !validUser {
			render.BadRequest(w, errors.New("Invalid username or password"))
			return

		}

		token, err := auth.GenerateToken(user.Id)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		var res loginUserResponse
		res.Message = "User logged successfully"
		res.Data.Name = user.Name
		res.Data.AccessToken = token
		if req.CredentialType == "email" {
			res.Data.Email = req.CredentialValue
			res.Data.Phone = ""
		}
		if req.CredentialType == "phone" {
			res.Data.Phone = req.CredentialValue
			res.Data.Email = ""
		}

		render.JSON(w, res, http.StatusOK)
	}
}

func HandleRegistration(us *us.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req createUserRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &req); err != nil {
			render.BadRequest(w, err)
			return
		}

		if err := us.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}

		user := model.User{
			Name:     req.Name,
			Password: req.Password,
		}
		credential := model.Credential{
			CredentialType:  req.CredentialType,
			CredentialValue: req.CredentialValue,
		}

		err = user.HashPassword()
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		userId, err := us.CreateUser(r.Context(), &user, &credential)
		if err != nil {
			render.ErrorCode(w, err, 409)
			return
		}

		token, err := auth.GenerateToken(userId)
		if err != nil {
			render.InternalError(w, err)
			return
		}

		var res createUserResponse
		res.Message = "User registered successfully"
		res.Data.Name = user.Name
		res.Data.AccessToken = token
		if req.CredentialType == "phone" {
			res.Data.Phone = req.CredentialValue
		}
		if req.CredentialType == "email" {
			res.Data.Email = req.CredentialValue
		}

		render.JSON(w, res, http.StatusCreated)
	}
}

func HandleLinkEmail(us *us.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req linkEmailRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &req); err != nil {
			render.BadRequest(w, err)
			return
		}

		if err := us.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}

		userId, err := auth.GetUserId(r.Context())
		err = us.UpdateUserEmail(r.Context(), req.Email, userId)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					render.ErrorCode(w, err, 409)
					return
				}
			}
			render.BadRequest(w, err)
			return
		}
		render.JSON(w, map[string]interface{}{}, 200)
	}
}

func HandleLinkPhone(us *us.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req linkPhoneRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &req); err != nil {
			render.BadRequest(w, err)
			return
		}

		if err := us.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}

		userId, err := auth.GetUserId(r.Context())
		err = us.UpdateUserPhone(r.Context(), req.Phone, userId)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					render.ErrorCode(w, err, 409)
					return
				}
			}
			render.BadRequest(w, err)
			return
		}
		render.JSON(w, map[string]interface{}{}, 200)
	}
}
