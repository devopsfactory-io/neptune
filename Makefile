.PHONY: build test-all check-fmt
BINARY := neptune

build:
	go build -o $(BINARY) .

test-all:
	go test ./...

check-fmt:
	@test -z "$$(gofmt -s -l .)" || (echo "Run: gofmt -s -w ."; gofmt -s -l .; exit 1)

lint:
	golangci-lint run ./...
