APP=sonarbrige
VERSION ?= 1.0.2
COMMIT := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PKG_VERSION := sonarbridge-go/internal/build
VCS_DIST := bin.local/assets/binaries

BIN_DIR := ./bin

LDFLAGS := -s -w -X '$(PKG_VERSION).Version=$(VERSION)' \
           -X '$(PKG_VERSION).Commit=$(COMMIT)' \
           -X '$(PKG_VERSION).Date=$(DATE)'

.PHONY: run cli test tidy build-server build-server-prod build-ci build-clid-prod gor-gitlab-plugin gor-github-plugin

gor-gitlab-plugin:
	goreleaser release --clean --skip=validate,publish --config ./plugins/vcs/gitlab/.gor-config.yaml --verbose
	LOCAL_REGISTRY_URL=$(REGISTRY_URL) LOCAL_REGISTRY_UPLOAD_TOKEN=$(ARTIFACTS_TOKEN)
	./upload.sh --latest \
  	  --version=$$(jq -r .version $(VCS_DIST)/gitlab/metadata.json) \
  	  gitlab $(VCS_DIST)/gitlab/artifacts.json

gor-github-plugin:
	goreleaser release --clean --skip=validate,publish --config ./plugins/vcs/github/.gor-config.yaml --verbose
	LOCAL_REGISTRY_URL=$(REGISTRY_URL) LOCAL_REGISTRY_UPLOAD_TOKEN=$(ARTIFACTS_TOKEN)
	./upload.sh --latest \
  	  --version=$$(jq -r .version $(VCS_DIST)/github/metadata.json) \
  	  github $(VCS_DIST)/github/artifacts.json

run:
	go run ./cmd/server

cli:
	go run ./cmd/cli

test:
	go test ./..

tidy:
	go mod tidy

build-server:
	go build -o ${BIN_DIR}/sonarbridge-server-dev ./cmd/server

build-server-prod:
	go build \
	  -trimpath \
	  -ldflags "$(LDFLAGS)" \
	  -o ${BIN_DIR}/sonarbridge-server ./cmd/server

build-cli:
	go build -buildvcs=false -o ${BIN_DIR}/sonarbridge-cli-dev ./cmd/cli

build-cli-prod:
	go build -trimpath -ldflags "$(LDFLAGS)" -o ${BIN_DIR}/sonarbridge-cli ./cmd/cli