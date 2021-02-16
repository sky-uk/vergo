.PHONY: fun-tests unit-tests test

export

SHELL := /bin/bash
LINTER_VERSION := 1.36.0
UPLOAD_TARGET=https://nexus.api.bskyb.com/nexus/content/repositories/nova-packages
PATH := $(shell pwd)/bin:$(PATH)

bin:
	mkdir -p bin/

bin/golangci-lint: bin
	curl -fsL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v${LINTER_VERSION}

command-available:
	#this empty target is for 'if' syntax to work

pre-check: $(if $(shell which golangci-lint), command-available, bin/golangci-lint)
	[[ `(gofmt -l .)`x == x ]] || (echo "go fmt failed" && gofmt -l . && exit 1)
	go vet ./...
	golangci-lint run ./...

release: bin build-test
	bin/vergo bump minor --repository-location ".." --tag-prefix vergo
	GORELEASER_CURRENT_TAG=`bin/vergo get latest-release --repository-location ".." --tag-prefix vergo -p` \
	GORELEASER_PREVIOUS_TAG=`bin/vergo get previous-release --repository-location ".." --tag-prefix vergo -p` \
	goreleaser release --skip-validate --rm-dist
	bin/vergo push --repository-location ".." --tag-prefix vergo

unit-tests: pre-check
	go clean -testcache
	go test ./...

fun-tests: build-test
	./fun-tests/test.sh

test: unit-tests fun-tests
test-compile:
	go test --exec=true ./...

build-test: bin pre-check
	GORELEASER_CURRENT_VERSION=`git rev-parse --short HEAD` goreleaser build --snapshot --rm-dist
	@dist/vergo_`uname | tr A-Z a-z`_amd64/vergo version
	@cp dist/vergo_`uname | tr A-Z a-z`_amd64/vergo bin/vergo

extended-linter: pre-check
	golangci-lint run --enable-all --disable wrapcheck,nlreturn,exhaustivestruct,wsl,gofumpt,gci ./...