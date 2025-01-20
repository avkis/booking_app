.DEFAULT_GOAL := build

.PHONY:fmt vet build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build

run:
	go run ./cmd/web

test:
	go test -v ./cmd/web

test_cover:
	go test -cover ./cmd/web

test_cover_html:
	go test -coverprofile=coverage.out ./cmd/web && go tool cover -html=coverage.out

migrate:
	soda migrate

migrate_down:
	soda migrate down

mailhog_run:
	docker run -d -p 8025:8025 -p 1025:1025 mailhog


