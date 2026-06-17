.PHONY: run build tidy test docker-build docker-up docker-down fmt vet

run:
	go run .

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/docflow-backend .

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

docker-build:
	docker build -t docflow-backend:latest .

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
