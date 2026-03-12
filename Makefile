.PHONY: build test-all check-fmt fmt e2e e2e.terramate e2e.localstacksfiles e2e.localdeclaredstacks lambda.build lambda.zip lambda.test goreleaser.test
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

# Lambda (separate Go module under lambda/). Binary named bootstrap for provided.al2023 runtime.
lambda.build:
	cd lambda && GOOS=linux GOARCH=amd64 go build -o bootstrap .

lambda.zip: lambda.build
	cd lambda && zip -qj neptune-webhook.zip bootstrap

lambda.test:
	cd lambda && go test ./...

# Test GoReleaser locally (build + changelog; no publish). Requires goreleaser on PATH. Output: dist/ (artifacts, dist/CHANGELOG.md).
goreleaser.test:
	goreleaser release --skip=publish,validate --clean
