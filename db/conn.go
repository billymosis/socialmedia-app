package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func Connection(driver, host, database, username, password, port string) (*pgxpool.Pool, error) {
	dsn, err := parseDSN(driver, host, database, username, password, port)
	logrus.Printf("SQL CONN: %+v\n", dsn)
	if err != nil {
		return nil, err
	}

	db, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if err := pingDatabase(db); err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), db)

	return pool, err
}

func pingDatabase(db *pgxpool.Config) error {
	pool, err := pgxpool.NewWithConfig(context.Background(), db)
	if err != nil {
		log.Fatal(err)
		pool.Close()
		return errPingDatabase
	}
	r := 3
	for i := 0; i < r; i++ {
		err := pool.Ping(context.Background())
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	pool.Close()
	return errPingDatabase
}

func parseDSN(driver, host, database, username, password, port string) (string, error) {

	switch driver {
	case "postgres":
		return postgreParseDSN(host, database, username, password, port), nil
	default:
		return "", errUnSupportedDriver
	}
}

func postgreParseDSN(host, database, username, password, port string ) string {
	dbUrl := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?%s",
		username,
		password,
		host,
		port,
		database,
		os.Getenv("DB_PARAMS"),
	)
	return dbUrl
}
