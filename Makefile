.PHONY: build test-all check-fmt fmt e2e e2e.terramate e2e.localstacksfiles e2e.localdeclaredstacks lambda.build lambda.zip lambda.test
BINARY := neptune

build:
	go build -o $(BINARY) .

test-all:
	go test ./...

check-fmt:
	@test -z "$$(gofmt -s -l .)" || (echo "Run: gofmt -s -w ."; gofmt -s -l .; exit 1)

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./...

e2e: e2e.terramate e2e.localstacksfiles e2e.localdeclaredstacks

e2e.terramate:
	./e2e/scripts/run-terramate.sh

e2e.localstacksfiles:
	./e2e/scripts/run-local-stacks-files.sh

e2e.localdeclaredstacks:
	./e2e/scripts/run-local-declared-stacks.sh

# Lambda (separate Go module under lambda/). Binary name neptune-webhook; CloudFormation Handler must match.
lambda.build:
	cd lambda && GOOS=linux GOARCH=amd64 go build -o neptune-webhook .

lambda.zip: lambda.build
	cd lambda && zip -q neptune-webhook.zip neptune-webhook

lambda.test:
	cd lambda && go test ./...
