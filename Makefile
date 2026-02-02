.PHONY: tidy test run mongo-up mongo-down

tidy:
	go mod tidy

test:
	go test ./...

run:
	go run ./cmd/api

mongo-up:
	docker compose up -d

mongo-down:
	docker compose down -v

