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

test-sync:
	$(gotestsum) --debug --format testname -- -mod=readonly -tags=syncintegration -race -coverpkg=./... -covermode=atomic -coverprofile=coverage-sync.txt ./cmd/...

gen-test-infra:
	cd .infra/infra; terraform apply -auto-approve

destroy-test-infra:
	cd .infra/infra; terraform apply -destroy -auto-approve

gen-test-usage:
	cd .infra/infra; terraform output -json | go run ../usage