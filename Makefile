init:
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.61.0

tidy:
	go mod tidy

format: tidy
	goimports -local github.com/2754github/oracle -w .

lint: format
	go vet ./...
	$(shell go env GOPATH)/bin/golangci-lint run

test: lint
	go clean -testcache
	go test -v ./...

run:
	go run main.go

.PHONY: init tidy format lint test run
