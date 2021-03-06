VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/gotesplit.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

.PHONY: devel-deps
devel-deps:
	sh -c '\
      tmpdir=$$(mktemp -d); \
      cd $$tmpdir; \
      go get ${u} \
        golang.org/x/lint/golint            \
        github.com/Songmu/godzil/cmd/godzil \
        github.com/tcnksm/ghr; \
      rm -rf $$tmpdir'

.PHONY: test
test:
	go test

.PHONY: lint
lint: devel-deps
	golint -set_exit_status

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/gotesplit

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./cmd/gotesplit

.PHONY: release
release: devel-deps
	godzil release -branch=main

CREDITS: deps devel-deps
	godzil credits -w

DIST_DIR = dist/v$(VERSION)
.PHONY: crossbuild
crossbuild: CREDITS
	rm -rf $(DIST_DIR)
	env CGO_ENABLED=0 godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
      -os=linux,darwin,windows -d=$(DIST_DIR) ./cmd/*
	cd $(DIST_DIR) && shasum -a 256 $$(find * -type f -maxdepth 0) > SHA256SUMS

.PHONY: upload
upload:
	ghr -body="$$(godzil changelog --latest -F markdown)" v$(VERSION) dist/v$(VERSION)
