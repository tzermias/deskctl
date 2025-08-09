
deps:
	go mod download
	go mod tidy

test: deps
	go test -v ./...

fmt:
	gofmt -s -w .
vet:
	go vet ./...

build: fmt vet test
	PKG=$(PKG) goreleaser build --clean --snapshot

clean:
	go clean 
	rm -rf dist/ bin/
