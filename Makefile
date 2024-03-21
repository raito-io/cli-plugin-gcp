GO := go

gotestsum := go run gotest.tools/gotestsum@latest
gotestconv := go run github.com/vladopajic/go-test-coverage/v2@main

.PHONY: generate wire lint test test-sync unit-test generate

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
	$(gotestsum) --debug --format pkgname  -- -mod=readonly -tags=integration -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./...

test-sync:
	$(gotestsum) --debug --format testname -- -mod=readonly -tags=syncintegration -race -coverpkg=./... -covermode=atomic -coverprofile=sync-coverage.txt ./cmd/...

check-coverage: test test-sync
	$(gotestconv) --config=./.testcoverage.yml
