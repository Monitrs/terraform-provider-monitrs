.PHONY: build test testacc docs fmt

BINARY   := terraform-provider-monitrs
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

test:
	go test -v ./...

testacc:
	TF_ACC=1 go test -v -timeout 120m ./...

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

fmt:
	gofmt -w .
	terraform fmt -recursive examples/
