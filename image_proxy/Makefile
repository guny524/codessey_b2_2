COMPONENT = image_proxy
STATE = main
REPO_NAME ?= ghcr.io/guny524
VERSION = v00.00.01
ARCH = amd64
CI_COMMIT_SHORT_SHA ?= $(shell git rev-parse --short=8 HEAD)
BUILD_DATE = $(shell date +"%y%m%d")

TAG = ${STATE}-${VERSION}-${ARCH}-${CI_COMMIT_SHORT_SHA}-${BUILD_DATE}
IMG = ${REPO_NAME}/${COMPONENT}:${TAG}

lint:
	go fmt ./...
	golangci-lint run

test: lint
	go clean -testcache
	go test -race ./...

test-cover: lint
	go clean -testcache
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

build: test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${COMPONENT} .

run:
	go run .

clean:
	- rm -f ${COMPONENT} coverage.out

image-build: build
	docker build . -t ${IMG} --platform linux/${ARCH}

.PHONY: lint test test-cover build run clean image-build
