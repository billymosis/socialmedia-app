package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/billymosis/socialmedia-app/handler/render"
	"github.com/billymosis/socialmedia-app/model"
	"github.com/billymosis/socialmedia-app/service/auth"
	ps "github.com/billymosis/socialmedia-app/store/post"
)

func Create(ps *ps.PostStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createPostRequest

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

		if err := ps.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}

		post := model.Post{
			Html: req.Html,
			Tags: req.Tags,
		}
		for _, el := range post.Tags {
			if el == "" {
				render.BadRequest(w, errors.New("Bad Tags"))
				return
			}
		}
		userId, err := auth.GetUserId(r.Context())
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = ps.Create(r.Context(), &post, userId)
		if err != nil {
			render.InternalError(w, err)
			return
		}
		w.WriteHeader(200)
	}

}

func CreateComment(ps *ps.PostStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createCommentRequest

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

		if err := ps.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}
		userId, err := auth.GetUserId(r.Context())
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		postid, err := strconv.Atoi(req.PostId)
		if err != nil {
			render.NotFound(w, err)
			return
		}

		comment := model.Comment{
			PostId:  postid,
			Comment: req.Comment,
		}

		err = ps.CreateComment(r.Context(), &comment, userId)
		if err != nil {
			render.InternalError(w, err)
			return
		}
		w.WriteHeader(200)
	}

}

func GetPost(ps *ps.PostStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := ps.GetPostList(r.Context(), r.URL.Query())
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		render.JSON(w, response, http.StatusOK)
	}

}
