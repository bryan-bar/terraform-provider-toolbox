TF_ACC_EXTERNAL_TIMEOUT_TEST ?= 0

ifeq ($(TF_ACC_EXTERNAL_TIMEOUT_TEST),0)
	GO_TIMEOUT = 3m
else
	GO_TIMEOUT = 23m
endif

default: build

build:
	go build -v ./...

install: build
	go install -v ./...

# See https://golangci-lint.run/
lint:
	golangci-lint run

generate:
	go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=$(GO_TIMEOUT) -parallel=4 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout $(GO_TIMEOUT) ./...

.PHONY: build install lint generate fmt test testacc
