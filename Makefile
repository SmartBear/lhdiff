.PHONY: gen lint test install

VERSION := "0.0.1"

gen:
	go generate ./...

lint: gen
	golangci-lint run

test: lint
	go test -v --race ./...

install: test
	go install -a -ldflags "-X=main.version=$(VERSION)" ./...
