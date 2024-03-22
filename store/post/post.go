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
		       c.comment, c.created_at,
		       uc.id, uc.name, uc.image_url, uc.friend_count, uc.created_at,
		       u.id post_creator_id, u.name as post_creator_name, u.image_url, u.friend_count, u.created_at
		FROM posts p
		LEFT JOIN comments c ON p.id  = c.post_id  
		LEFT JOIN users u ON p.user_id  = u.id 
		LEFT JOIN users uc on c.user_id = uc.id
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
			q.Query(" AND p.tags @> ")
			q.Param("[\"" + element + "\"]")
		}

	}

	limitStr := queryParams.Get("limit")
	limit := 10

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return nil, nil
		}
	}
	q.Query("\nLIMIT ")
	q.Param(limit)

	offsetStr := queryParams.Get("offset")
	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return nil, nil
		}
	}
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

	rows, err := ps.db.Query(ctx, query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get posts")
	}
	commentMap := make(map[string][]model.CommentResponseValid)
	postMap := make(map[string]model.PostResponseData)
	order := make([]string, 0)
	for rows.Next() {
		var data model.PostResponseData
		var comment model.CommentResponse
		var tagsJSON []byte
		var postUserImage sql.NullString
		err := rows.Scan(&data.PostID, &data.PostContent.PostInHTML, &tagsJSON, &data.PostContent.CreatedAt, &comment.Comment, &comment.CreatedAt, &comment.Creator.UserID, &comment.Creator.Name, &comment.Creator.ImageURL, &comment.Creator.FriendCount, &comment.Creator.CreatedAt, &data.Creator.UserID, &data.Creator.Name, &postUserImage, &data.Creator.FriendCount, &data.Creator.CreatedAt)
		order = append(order, data.PostID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan posts")
		}
		if err := json.Unmarshal(tagsJSON, &data.PostContent.Tags); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal tags JSON")
		}
		_, ok := postMap[data.PostID]
		if !ok {
			postMap[data.PostID] = data
		}
		if comment.Comment.Valid {
			commentMap[data.PostID] = append(commentMap[data.PostID],
				model.CommentResponseValid{
					Comment:   comment.Comment.String,
					CreatedAt: comment.CreatedAt.Time,
					Creator: model.CreatorValid{
						UserID:      strconv.Itoa(int(comment.Creator.UserID.Int32)),
						Name:        comment.Creator.Name.String,
						ImageURL:    postUserImage.String,
						FriendCount: int(comment.Creator.FriendCount.Int32),
						CreatedAt:   comment.CreatedAt.Time,
					},
				},
			)
		}
	}
	var res model.PostResponse = model.PostResponse{
		Data: []model.PostResponseData{},
	}
	order = helper.MakeUnique(order)
	for _, ord := range order {
		el, ok := postMap[ord]
		mod, ok := commentMap[el.PostID]
		if ok {
			el.Comments = mod
		} else {
			el.Comments = make([]model.CommentResponseValid, 0)
		}
		res.Data = append(res.Data, el)
	}

	countQuery := strings.Split(strings.Split(query, q.Arr[0])[1], "LIMIT")[0]
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
