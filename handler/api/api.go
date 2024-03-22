package api

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	x "github.com/billymosis/socialmedia-app/handler/api/post"
	"github.com/billymosis/socialmedia-app/handler/api/relationship"
	"github.com/billymosis/socialmedia-app/handler/api/user"
	AppMiddleware "github.com/billymosis/socialmedia-app/middleware"
	"github.com/billymosis/socialmedia-app/service/image"
	pss "github.com/billymosis/socialmedia-app/store/post"
	rs "github.com/billymosis/socialmedia-app/store/relationship"
	us "github.com/billymosis/socialmedia-app/store/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	Users         *us.UserStore
	Relationships *rs.RelationshipStore
	Posts         *pss.PostStore
	S3Client      *s3.Client
}

func New(users *us.UserStore, relationships *rs.RelationshipStore, posts *pss.PostStore, s3client *s3.Client) Server {
	return Server{
		Users:         users,
		Relationships: relationships,
		Posts:         posts,
		S3Client:      s3client,
	}
}
func prometheusHandler() http.Handler {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	handler := promhttp.Handler()
	return handler
}

func (s Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Use(AppMiddleware.WrapWithPrometheus)

		r.Route("/user", func(r chi.Router) {
			r.Post("/login", user.HandleAuthentication(s.Users))
			r.Post("/register", user.HandleRegistration(s.Users))
			r.With(AppMiddleware.ValidateJWT).Patch("/", user.HandleUpdateUser(s.Users))
			r.Route("/link", func(r chi.Router) {
				r.Use(AppMiddleware.ValidateJWT)
				r.Post("/", user.HandleLinkEmail(s.Users))
				r.Post("/phone", user.HandleLinkPhone(s.Users))
			})
		})
		r.Route("/friend", func(r chi.Router) {
			r.Use(AppMiddleware.ValidateJWT)
			r.Get("/", relationship.Get(s.Relationships))
			r.Post("/", relationship.Add(s.Relationships))
			r.Delete("/", relationship.Delete(s.Relationships))
		})

		r.Route("/post", func(r chi.Router) {
			r.Use(AppMiddleware.ValidateJWT)
			r.Get("/", x.GetPost(s.Posts))
			r.Post("/", x.Create(s.Posts))
			r.Post("/comment", x.CreateComment(s.Posts))
		})

	})

	r.Route("/v1/image", func(r chi.Router) {
		r.Use(AppMiddleware.ValidateJWT)
		r.Post("/", image.Upload(s.S3Client))
	})
	return r
}
