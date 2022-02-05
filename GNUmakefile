.PHONY: fun-tests unit-tests test

export

GORELEASER_VERSION := 0.179.0
LINTER_VERSION := 1.43.0
UPLOAD_TARGET=https://nexus.api.bskyb.com/nexus/content/repositories/nova-packages
PATH := $(shell pwd)/bin:$(PATH)
SHELL := bash

.ONESHELL:
.SHELLFLAGS := -ec

bash-required-version:
	$(if $(filter oneshell,${.FEATURES}) \
	  , \
	  ,$(error oneshell not supported - update your make))

check-bash: bash-required-version
	@test $$BASH_VERSINFO = 5 || test $$BASH_VERSINFO = 4

bin: check-bash
	@mkdir -p bin/

bin/golangci-lint: bin
	@test -f $@ ||
	(
		curl -fsL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v${LINTER_VERSION}
		bin/golangci-lint version
	)

bin/goreleaser: bin
	@test -f $@ ||
	(
		curl -fsL https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_`uname`_x86_64.tar.gz -o goreleaser.tgz
		tar xvf goreleaser.tgz -C bin goreleaser
		rm goreleaser.tgz
		bin/goreleaser --version
	)

tools: bin/goreleaser bin/golangci-lint

pre-check: tools
	[[ `(gofmt -l .)`x == x ]] || (echo "go fmt failed" && gofmt -l . && exit 1)
	go vet ./...
	golangci-lint run ./...

release: build
	bin/vergo check release --tag-prefix vergo || exit 0
	bin/vergo bump minor --tag-prefix vergo
	BUILT_BY="`goreleaser --version | head -n1`, `go version`" \
	GORELEASER_CURRENT_TAG=`bin/vergo get latest-release --tag-prefix vergo -p` \
	GORELEASER_PREVIOUS_TAG=`bin/vergo get previous-release --tag-prefix vergo -p` \
	goreleaser release --skip-validate --rm-dist
	bin/vergo push --tag-prefix vergo

unit-tests: pre-check
	go clean -testcache
	go test ./...

fun-tests: build
	./fun-tests/test.sh

test: build unit-tests fun-tests
test-compile:
	go test --exec=true ./...

build: pre-check
	BUILT_BY="`goreleaser --version | head -n1`, `go version`" \
	GORELEASER_CURRENT_TAG=0+`git rev-parse --short HEAD` \
	goreleaser build --snapshot --rm-dist
	@dist/vergo_`uname | tr A-Z a-z`_amd64/vergo version
	@cp dist/vergo_`uname | tr A-Z a-z`_amd64/vergo bin/vergo

extended-linter: pre-check
	golangci-lint run --enable-all --disable wrapcheck,nlreturn,exhaustivestruct,wsl,gofumpt,gci ./...