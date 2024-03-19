#! /bin/bash
migrate -path ./db/migrations -database "postgres://myuser:mypassword@localhost:5432/mydatabase?sslmode=disable" $1
