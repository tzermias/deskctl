OUTPUT=deskctl
#Version variables
VERSION?=dev
GOMODULE=github.com/tzermias/deskcli
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X $(GOMODULE)/cmd.Version=$(VERSION) -X $(GOMODULE)/cmd.Commit=$(COMMIT) -X $(GOMODULE)/cmd.BuildTime=$(BUILD_TIME)"


make-deps:
	go mod download
	go mod tidy

test: make-deps
	go test

fmt:
	gofmt -s -w .
vet:
	go vet ./...

build: fmt vet test
	go build $(LDFLAGS) -o bin/$(OUTPUT) -v ./

clean:
	go clean 
	rm -rf dist/ bin/
