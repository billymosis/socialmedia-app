package user

import (
	"context"

	"github.com/billymosis/socialmedia-app/model"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type UserStore struct {
	db       *pgxpool.Pool
	Validate *validator.Validate
}

func NewUserStore(db *pgxpool.Pool, validate *validator.Validate) *UserStore {
	return &UserStore{
		db:       db,
		Validate: validate,
	}
}

func (us *UserStore) GetById(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	query := "SELECT * FROM users WHERE id = $1 LIMIT 1"
	err := us.db.QueryRow(ctx, query, id).Scan(
		&user.Id,
		&user.Name,
		&user.Password,
		&user.ImageUrl,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user by ID")
	}
	return &user, nil
}

func (us *UserStore) GetByCredential(ctx context.Context, credentialValue string) (*model.UserAndCred, error) {
	var user model.UserAndCred
	query := `
		SELECT u.id AS user_id, u.password as password, u.name as name 
		FROM users u
		JOIN user_credentials uc ON u.id = uc.user_id
		WHERE uc.credential_value = $1
	`
	err := us.db.QueryRow(ctx, query, credentialValue).Scan(
		&user.Id,
		&user.Password,
		&user.Name,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user by username")
	}
	query = `
	SELECT credential_type, credential_value
	FROM user_credentials
	WHERE user_id = $1
	`
	rows, err := us.db.Query(ctx, query, user.Id)
	for rows.Next() {
		var cred model.Credential
		rows.Scan(&cred.CredentialType, &cred.CredentialValue)
		if cred.CredentialType == "email" {
			user.Email = cred.CredentialValue
		}
		if cred.CredentialType == "phone" {
			user.Phone = cred.CredentialValue
		}
	}

	return &user, nil
}

func (us *UserStore) CreateUser(ctx context.Context, user *model.User, credential *model.Credential) (int, error) {
	query := "INSERT INTO users (name, password) VALUES($1,$2) RETURNING id"
	err := us.db.QueryRow(ctx, query,
		&user.Name,
		&user.Password,
	).Scan(&user.Id)
	if err != nil {
		return user.Id, errors.Wrap(err, "failed to create user")
	}
	query = "INSERT INTO user_credentials (credential_type, credential_value, user_id) VALUES($1,$2,$3)"
	_, err = us.db.Exec(ctx, query, credential.CredentialType, credential.CredentialValue, user.Id)
	if err != nil {
		return user.Id, errors.Wrap(err, "failed to create credentials")
	}
	return user.Id, nil
}

func (us *UserStore) UpdateUserEmail(ctx context.Context, email string, userId int) error {
	query := `
		SELECT EXISTS (
		    SELECT 1
		    FROM user_credentials
		    WHERE user_id = $1
		    AND credential_type = $2
		)
	`
	var exist bool
	err := us.db.QueryRow(ctx, query, userId, "email").Scan(&exist)
	if err != nil {
		return errors.Wrap(err, "failed to create credentials")
	}
	if exist {
		return errors.New("email already exist")

	}
	query = "INSERT INTO user_credentials (credential_type, credential_value, user_id) VALUES($1,$2, $3)"
	_, err = us.db.Exec(ctx, query, "email", email, userId)
	if err != nil {
		return errors.Wrap(err, "failed to create credentials")
	}
	return nil
}

func (us *UserStore) UpdateUserPhone(ctx context.Context, phone string, userId int) error {
	query := `
		SELECT EXISTS (
		    SELECT 1
		    FROM user_credentials
		    WHERE user_id = $1
		    AND credential_type = $2
		)
	`
	var exist bool
	err := us.db.QueryRow(ctx, query, userId, "phone").Scan(&exist)
	if err != nil {
		return errors.Wrap(err, "failed to create credentials")
	}
	if exist {
		return errors.New("phone already exist")

	}
	query = "INSERT INTO user_credentials (credential_type, credential_value, user_id) VALUES($1,$2, $3)"
	_, err = us.db.Exec(ctx, query, "phone", phone, userId)
	if err != nil {
		return errors.Wrap(err, "failed to create credentials")
	}
	return nil
}

func (us *UserStore) UpdateUser(ctx context.Context, image string, name string, userId int) error {
	query := `
	UPDATE users SET name = $1, image_url = $2 WHERE id = $3 
	`
	_, err := us.db.Exec(ctx, query, name, image, userId)
	if err != nil {
		return errors.Wrap(err, "failed to update users")
	}
	return nil
}
