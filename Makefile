.PHONY: build test-all check-fmt e2e lambda.build lambda.zip lambda.test
BINARY := neptune

build:
	go build -o $(BINARY) .

test-all:
	go test ./...

check-fmt:
	@test -z "$$(gofmt -s -l .)" || (echo "Run: gofmt -s -w ."; gofmt -s -l .; exit 1)

lint:
	golangci-lint run ./...

e2e:
	./e2e/run.sh

# Lambda (separate Go module under lambda/). Binary name bootstrap is for provided.al2023 runtime.
lambda.build:
	cd lambda && GOOS=linux GOARCH=amd64 go build -o bootstrap .

lambda.zip: lambda.build
	cd lambda && zip -q neptune-webhook.zip bootstrap

lambda.test:
	cd lambda && go test ./...
