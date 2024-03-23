package relationship

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/billymosis/socialmedia-app/helper"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type RelationshipStore struct {
	db       *pgxpool.Pool
	Validate *validator.Validate
}

func NewRelationshipStore(db *pgxpool.Pool, validate *validator.Validate) *RelationshipStore {
	return &RelationshipStore{
		db:       db,
		Validate: validate,
	}
}

func (ps *RelationshipStore) AddFriend(ctx context.Context, userAddId int, userId int) error {

	query := `
		SELECT EXISTS (
		    SELECT 1
		    FROM users
		    WHERE id = $1
		)
	`
	var exist bool
	err := ps.db.QueryRow(ctx, query, userId).Scan(&exist)
	if err != nil {
		return errors.Wrap(err, "failed check user exist")
	}
	if !exist {
		return errors.New("not exist")
	}
	query = `
		WITH inserted_relationship AS (
		  INSERT INTO relationships (user_first_id, user_second_id)
		  VALUES ($1, $2)
		  RETURNING user_first_id, user_second_id
		)
		UPDATE users
		SET friend_count = friend_count + 1
		WHERE id IN (SELECT user_first_id FROM inserted_relationship UNION SELECT user_second_id FROM inserted_relationship);
	`
	_, err = ps.db.Exec(ctx, query, userId, userAddId)
	if err != nil {
		return errors.Wrap(err, "failed to add relation")
	}
	return nil
}

func (ps *RelationshipStore) DeleteFriend(ctx context.Context, userAddId int, userId int) error {
	query := `
	WITH deleted_relationship AS (
		DELETE FROM relationships
		WHERE
		    (user_first_id = $1 AND user_second_id = $2)
		    OR
		    (user_first_id = $2 AND user_second_id = $1)
	)
		UPDATE users
		SET friend_count = friend_count - 1
		WHERE users.id = $1 OR users.id = $2;
	`
	_, err := ps.db.Exec(ctx, query, userId, userAddId)
	if err != nil {
		return errors.Wrap(err, "failed to add relation")
	}
	return nil
}

type Meta struct {
	Limit  int
	Offset int
	Total  int
}
type Friend struct {
	UserId      string
	Name        string
	ImageUrl    *string
	FriendCount int
	CreatedAt   time.Time
}

type GetFriendListRow struct {
	Friends []*Friend
	Meta    Meta
}

func (ps *RelationshipStore) GetFriendList(ctx context.Context, userId int, queryParams url.Values) (*GetFriendListRow, error) {
	q := helper.Query{}
	q.Query(
		`
		SELECT u.id as user_id, u.name as name, u.image_url as image_url, u.friend_count, u.created_at
		FROM users u
		LEFT JOIN relationships r ON u.id = r.user_first_id OR u.id = r.user_second_id WHERE`)

	if queryParams.Has("onlyFriend") && queryParams.Get("onlyFriend") == "" {
		return nil, errors.New("bad request: invalid sortBy parameter")
	}
	onlyFriendStr := queryParams.Get("onlyFriend")
	var onlyFriend bool = false
	if onlyFriendStr != "" {
		b, err := strconv.ParseBool(onlyFriendStr)
		onlyFriend = b
		if err != nil {
			return nil, errors.Wrap(err, "bad request")
		}
	}
	search := queryParams.Get("search")
	hasParams := false
	if onlyFriend || search != "" {
		hasParams = true
	}

	if onlyFriend {
		q.Query(" AND u.id <> ")
		q.Param(userId)
		q.Query(" AND r.user_first_id = ")
		q.Param(userId)
		q.Query(" OR r.user_second_id = ")
		q.Param(userId)
	}

	limit := 10
	limitStr := queryParams.Get("limit")
	if queryParams.Has("limit") && limitStr == "" {
		return nil, errors.New("bad request")
	}
	if queryParams.Has("limit") && limitStr != "" {
		limitx, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, errors.Wrap(err, "bad request")
		}
		limit = limitx
	}

	offset := 0
	offsetStr := queryParams.Get("offset")
	if queryParams.Has("offset") && offsetStr == "" {
		return nil, errors.New("bad request")
	}
	if queryParams.Has("offset") && offsetStr != "" {
		offsetx, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, errors.Wrap(err, "bad request")
		}
		offset = offsetx
	}

	if search != "" {
		q.Query(" AND u.name LIKE ")
		q.Param("%" + search + "%")
	}

	orderBy := "DESC"
	if queryParams.Has("orderBy") && queryParams.Get("orderBy") == "" {
		return nil, errors.New("bad request: invalid sortBy parameter")
	}
	if queryParams.Get("orderBy") != "" {
		orderBy = queryParams.Get("orderBy")
	}

	sortBy := "created_at"
	if queryParams.Has("sortBy") && queryParams.Get("sortBy") == "" {
		return nil, errors.New("bad request: invalid sortBy parameter")
	}
	switch sortByParam := queryParams.Get("sortBy"); sortByParam {
	case "createdAt":
		sortBy = "u.created_at"
	case "friendCount":
		sortBy = "u.friend_count"
	case "":
	default:
		return nil, errors.New("bad request: invalid sortBy parameter")
	}

	q.Query(" GROUP BY u.id ")
	q.Query(fmt.Sprintf("\nORDER BY %s", sortBy))
	q.Query(fmt.Sprintf(" %s", orderBy))

	q.Query("\nLIMIT ")
	q.Param(limit)

	q.Query(" OFFSET ")
	q.Param(offset)
	query, params := q.Get()

	if !hasParams {
		query = strings.Replace(query, "WHERE", "", 1)
	} else {
		if strings.Count(query, "AND") == 0 {
			query = strings.Replace(query, "WHERE", "", 1)
		}
		query = strings.Replace(query, "WHERE AND", "WHERE", 1)
	}
	logrus.Printf("QUERY: %+v\n", query)
	logrus.Printf("QUERY: %+v\n", params...)

	rows, err := ps.db.Query(ctx, query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query friend list")
	}

	defer rows.Close()

	var users []*Friend
	for rows.Next() {
		var user Friend
		if err = rows.Scan(&user.UserId, &user.Name, &user.ImageUrl, &user.FriendCount, &user.CreatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan friend data")
		}
		users = append(users, &user)
		logrus.Printf("QUERY: %+v\n", &user)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error while iterating over rows")
	}

	countQuery := strings.Split(query, "LIMIT")[0]
	countQuery = fmt.Sprintf(`
	SELECT COUNT (*) AS total_rows
	FROM (
		%s
	) AS subquery; 
	`, countQuery)
	params = params[:len(params)-2]
	var count int
	err = ps.db.QueryRow(ctx, countQuery, params...).Scan(&count)
	logrus.Println("WOX6")

	if err != nil {
		return nil, errors.Wrap(err, "failed to get total friend list")
	}
	friends := GetFriendListRow{
		Friends: users,
		Meta: Meta{
			Limit:  limit,
			Offset: offset,
			Total:  count,
		},
	}

	logrus.Println("AKHIR")

	return &friends, err
}
