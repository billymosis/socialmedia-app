package post

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/billymosis/socialmedia-app/helper"
	"github.com/billymosis/socialmedia-app/model"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PostStore struct {
	db       *pgxpool.Pool
	Validate *validator.Validate
}

func NewPostStore(db *pgxpool.Pool, validate *validator.Validate) *PostStore {
	return &PostStore{
		db:       db,
		Validate: validate,
	}
}

func (ps *PostStore) Create(ctx context.Context, post *model.Post, userId int) error {
	tagsJSON, err := json.Marshal(post.Tags)
	if err != nil {
		return errors.Wrap(err, "failed to marshal tags to JSON")
	}
	query := `
	INSERT INTO posts
	(html, user_id, tags)
	VALUES($1,$2,$3)
	`

	_, err = ps.db.Exec(ctx, query, post.Html, userId, tagsJSON)
	if err != nil {
		return errors.Wrap(err, "failed to create posts")
	}
	return nil
}

func (ps *PostStore) CreateComment(ctx context.Context, comment *model.Comment, userId int) error {
	query := `
		INSERT INTO comments
		(comment, post_id, user_id)
		VALUES($1,$2,$3)
	`

	_, err := ps.db.Exec(ctx, query, comment.Comment, comment.PostId, userId)
	if err != nil {
		return errors.Wrap(err, "failed to create comments")
	}
	return nil
}

func (ps *PostStore) GetPostList(ctx context.Context, queryParams url.Values) (*model.PostResponse, error) {
	q := helper.Query{}
	q.Query(`
		SELECT p.id, p.html, p.tags, p.created_at,
		       u.id post_creator_id, u.name as post_creator_name, u.image_url, u.friend_count, u.created_at
		FROM posts p
		LEFT JOIN users u ON p.user_id  = u.id 
		WHERE`)
	var err error

	search := queryParams.Get("search")
	tags := queryParams["searchTag"]
	hasParams := false
	if len(tags) > 0 || search != "" {
		hasParams = true
	}

	if search != "" {
		q.Query(" AND p.html LIKE ")
		q.Param("%" + search + "%")
	}

	if len(tags) > 0 {
		for _, element := range tags {
			q.Query(" AND p.tags ? ")
			q.Param(element)
		}
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
		if limitx < 0 {
			return nil, errors.New("bad request")
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
		if offsetx < 0 {
			return nil, errors.New("bad request")
		}
		offset = offsetx
	}

	q.Query(fmt.Sprintf("\nORDER BY %s", "p.created_at"))
	q.Query(fmt.Sprintf(" %s", "DESC"))

	q.Query(" LIMIT ")
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
	logrus.Printf("QUE: %+v\n", query)
	logrus.Printf("QUE: %+v\n", params...)

	rows, err := ps.db.Query(ctx, query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get posts")
	}
	order := make([]*model.PostResponseData, 0)
	orderID := make([]string, 0)
	for rows.Next() {
		var data model.PostResponseData
		var tagsJSON []byte
		var postUserImage sql.NullString
		err := rows.Scan(&data.PostID, &data.PostContent.PostInHTML, &tagsJSON, &data.PostContent.CreatedAt, &data.Creator.UserId, &data.Creator.Name, &postUserImage, &data.Creator.FriendCount, &data.Creator.CreatedAt)
		order = append(order, &data)
		orderID = append(orderID, data.PostID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan posts")
		}
		if err := json.Unmarshal(tagsJSON, &data.PostContent.Tags); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal tags JSON")
		}
	}
	var res model.PostResponse = model.PostResponse{
		Data: []model.PostResponseData{},
	}

	query2 := fmt.Sprintf(`
		SELECT c.id, c.comment, c.post_id, c.user_id,  c.created_at,
		u.id, u.name, u.image_url, u.friend_count, u.created_at
		FROM comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.post_id IN (%s)
	`, strings.Join(orderID, ","))
	logrus.Printf("Q2: %+v\n", query2)
	rows, err = ps.db.Query(ctx, query2)
	var comments = make(map[int][]*model.CommentAndUser, 0)
	for rows.Next() {
		var mod model.CommentAndUser
		err = rows.Scan(&mod.Comment.Id, &mod.Comment.Comment, &mod.Comment.PostId, &mod.Comment.UserId, &mod.Comment.CreatedAt, &mod.Creator.UserId, &mod.Creator.Name, &mod.Creator.ImageURL, &mod.Creator.FriendCount, &mod.Creator.CreatedAt)

		if err != nil {
			return nil, errors.Wrap(err, "failed to scan comments")
		}

		comments[mod.Comment.PostId] = append(comments[mod.Comment.PostId], &mod)
	}
	logrus.Printf("OR: %+v\n", order)
	logrus.Printf("OR3: %+v\n", comments)
	for _, ord := range order {
		id, err := strconv.Atoi(ord.PostID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert")
		}
		ord.Comments = make([]model.CommentResponseValid, 0)
		current, ok := comments[id]
		if ok {
			for _, el := range current {
				logrus.Printf("OR2: %+v\n", el)
				ord.Comments = append(ord.Comments, model.CommentResponseValid{
					Comment: el.Comment.Comment,
					Creator: model.CreatorValid{
						UserId:      el.Creator.UserId,
						Name:        el.Creator.Name,
						ImageURL:    el.Creator.ImageURL,
						FriendCount: el.Creator.FriendCount,
						CreatedAt:   el.Creator.CreatedAt,
					},
				})

			}

		}
		res.Data = append(res.Data, *ord)
	}

	countQuery := strings.Split(strings.Split(query, q.Arr[0])[1], "ORDER")[0]
	params = params[:len(params)-2]
	if len(params) > 0 {
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM posts p WHERE %s", countQuery)
	} else {
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM posts p %s", countQuery)

	}

	var count int
	err = ps.db.QueryRow(ctx, countQuery, params...).Scan(&count)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get total posts list")
	}
	res.Meta = model.Meta{
		Limit:  limit,
		Offset: offset,
		Total:  count,
	}

	return &res, nil
}
