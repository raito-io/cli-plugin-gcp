GO := go

gotestsum := go run gotest.tools/gotestsum@latest
gocheckcov := go run github.com/vladopajic/go-test-coverage/v2@latest

.PHONY: generate wire build unit-test lint test test-sync check-coverage

generate:
	go get github.com/raito-io/enumer
	go generate ./...

wire:
	go run github.com/google/wire/cmd/wire ./...

build: generate wire
	go build ./...

unit-test:
	$(gotestsum) --debug --format pkgname -- -mod=readonly -coverpkg=./... -covermode=atomic -coverprofile=unit-test-coverage.txt ./...

lint:
	golangci-lint run ./...
	go fmt ./...

test:
	$(gotestsum) --debug --format pkgname -- -mod=readonly -tags=integration -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./...

test-sync:
	$(gotestsum) --debug --format testname -- -mod=readonly -tags=syncintegration -race -coverpkg=./... -covermode=atomic -coverprofile=coverage-sync.txt ./cmd/...

check-coverage:
	$(gocheckcov) --config=./.testcoverage.yml
