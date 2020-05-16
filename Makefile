GO111MODULE=on
GONFIGD_BINARY_NAME=gonfigd
GONFIGD_VERSION=1.0.0
BUILD_FLAGS=-ldflags "-X main.version=v${GONFIGD_VERSION}"

fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	go test -v ./gonfig... ./fswatcher... ./kv/... ./pubsub/... -coverprofile cover.out

tidy:
	go mod tidy

build: fmt vet test tidy
	go build ${BUILD_FLAGS} -o bin/${GONFIGD_BINARY_NAME} main.go
