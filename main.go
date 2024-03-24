package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/billymosis/socialmedia-app/db"
	"github.com/billymosis/socialmedia-app/handler/api"
	pss "github.com/billymosis/socialmedia-app/store/post"
	rs "github.com/billymosis/socialmedia-app/store/relationship"
	us "github.com/billymosis/socialmedia-app/store/user"
	"github.com/go-playground/validator/v10"
	// "github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {

	// if err := godotenv.Load(); err != nil {
	// 	panic(err)
	// }

	host := os.Getenv("DB_HOST")
	database := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	region := os.Getenv("S3_REGION")

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("S3_ID"), os.Getenv("S3_SECRET_KEY"), "",
			)))

	s3Client := s3.NewFromConfig(cfg)

	db, err := db.Connection("postgres", host, database, user, password, port)
	if err != nil {
		log.Fatal(err)
	}
	validate := validator.New()

	userStore := us.NewUserStore(db, validate)
	relationStore := rs.NewRelationshipStore(db, validate)
	postStore := pss.NewPostStore(db, validate)

	r := api.New(userStore, relationStore, postStore, s3Client)
	h := r.Handler()

	logrus.Info("application starting billy fixed env")

	log.Println("application starting")

	go func() {
		s := http.Server{
			Addr:           ":8080",
			Handler:        h,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20, //1mb
		}

		err := s.ListenAndServe()
		if err != nil {
			log.Println("application failed to start")
			panic(err)
		}
	}()
	log.Println("application started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Info("application shutting down")

	log.Println("database closing")
	db.Close()
	log.Println("database closed")
}
