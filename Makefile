.PHONY: run test migrate-up migrate-down

-include .env
export

run:
	go run ./cmd/api
test:
	go test ./...
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL_MIGRATION)" up
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL_MIGRATION)" down 1
