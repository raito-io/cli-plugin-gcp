GO := go

gotestsum := go run gotest.tools/gotestsum@latest

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
	go tool cover -html=coverage.txt -o coverage.html
