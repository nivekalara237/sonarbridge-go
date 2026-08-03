APP=sonarbrige

.PHONY: run cli test tidy build build-ci build-clid-prod

run:
	go run ./cmd/server

cli:
	go run ./cmd/cli

test:
	go test ./..

tidy:
	go mod tidy

build:
	go build -o bin/sonarbridge-server ./cmd/server

build-cli:
	go build -o bin/sonarbridge-cli-dev ./cmd/cli

build-cli-prod:
	go build \
	  -ldflags "\
	  -X 'sonarbridge-go/internal/cli.Version=1.0.2' \
	  -X 'sonarbridge-go/internal/cli.Commit=$(git rev-parse --short HEAD)' \
	  -X 'sonarbridge-go/internal/cli.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
	  -o bin/sonarbridge-cli-prod ./cmd/cli