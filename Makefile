GO := go

gotestsum := go run gotest.tools/gotestsum@latest
gotestconv := go run github.com/vladopajic/go-test-coverage/v2@latest

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
	$(gotestsum) --debug --format pkgname -- -mod=readonly -tags=integration,syncintegration -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt ./...

check-coverage: test
	$(gotestconv) --config=./.testcoverage.yml
