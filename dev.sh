#! /bin/bash

cleanup() {
    echo "Performing cleanup..."
    docker compose down
    echo "Cleanup complete."
}

trap cleanup EXIT

source ./env.sh
docker compose up -d
air
