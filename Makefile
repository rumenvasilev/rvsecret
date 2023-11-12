SHELL = /bin/bash

.PHONY: all build clean coverage lint test
.DEFAULT_GOAL := lint

# The name of the binary to build
#
ifndef pkg
pkg := $(shell pwd | awk -F/ '{print $$NF}')
endif

# Set the target OS
# Ex: windows, darwin, linux
#
ifndef target_os
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		target_os = linux
	endif
	ifeq ($(UNAME_S),Darwin)
		target_os = darwin
	endif
	UNAME_P := $(shell uname -p)

	ifeq ($(UNAME_P),x86_64)
		target_arch = amd64
	endif
endif

ifeq ($(target_os),windows)
	target_ext = .exe
endif

ifndef target_arch
	target_arch = amd64
endif

## Lint, build
all: pretty build

## Build
build: clean
	@GOOS=$(target_os) GOARCH=$(target_arch) go build -mod vendor -o ./bin/$(pkg)-$(target_os)

linux:
	@GOOS=linux GOARCH=$(target_arch) go build -mod vendor -o ./bin/$(pkg)-linux

windows:
	@GOOS=windows GOARCH=$(target_arch) go build -mod vendor -o ./bin/$(pkg)-windows.exe

## Clean binaries
clean:
	@rm -rf ./bin

## Lint
lint:
	@go fmt ./...
	@golangci-lint run

## Some tests depend on built binary, make sure you've built it beforehand.
test:
	@go test -v -mod vendor -race ./... 

coverage:
	@go test -mod vendor -coverprofile=coverage.out ./...
