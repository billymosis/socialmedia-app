package relationship

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/billymosis/socialmedia-app/handler/render"
	"github.com/billymosis/socialmedia-app/model"
	"github.com/billymosis/socialmedia-app/service/auth"
	rs "github.com/billymosis/socialmedia-app/store/relationship"
	"github.com/sirupsen/logrus"
)

func Add(rs *rs.RelationshipStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addFriendRequest
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

		if err := rs.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}
		userIdRequest, err := strconv.Atoi(req.UserId)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		userId, err := auth.GetUserId(r.Context())
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = rs.AddFriend(r.Context(), userIdRequest, userId)
		if err != nil {
			
			render.BadRequest(w, err)
			return
		}
		w.WriteHeader(200)
	}
}

func Delete(rs *rs.RelationshipStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addFriendRequest
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

		if err := rs.Validate.Struct(req); err != nil {
			render.BadRequest(w, err)
			return
		}
		userIdRequest, err := strconv.Atoi(req.UserId)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		userId, err := auth.GetUserId(r.Context())
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = rs.DeleteFriend(r.Context(), userIdRequest, userId)
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		w.WriteHeader(200)
	}
}

func Get(rs *rs.RelationshipStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := auth.GetUserId(r.Context())
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		logrus.Println("WOI1")
		users, err := rs.GetFriendList(r.Context(), userId, r.URL.Query())
		logrus.Printf("%v\n", err)
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		logrus.Printf("WOI2 %v", users)
		var data []Friend = make([]Friend, 0)
		if users != nil {
			for _, user := range users.Friends {
				data = append(data, Friend{
					UserId:      user.UserId,
					Name:        user.Name,
					ImageUrl:    user.ImageUrl,
					FriendCount: user.FriendCount,
					CreatedAt:   user.CreatedAt,
				})
			}
		}
		logrus.Println("WOI3")
		res := GetFriendListRow{
			Message: "",
			Data:    data,
			Meta: model.Meta{
				Limit:  users.Meta.Limit,
				Offset: users.Meta.Offset,
				Total:  users.Meta.Total,
			},
		}
		render.JSON(w, res, 200)
	}
}
