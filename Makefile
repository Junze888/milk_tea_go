.PHONY: run tidy build docker-up docker-down

run:
	go run ./cmd/server

tidy:
	go mod tidy

build:
	go build -o bin/milktea-server ./cmd/server

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
