BIN := base58
VERSION := $$(make -s show-version)
VERSION_PATH := cli
CURRENT_REVISION := $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS := "-s -w -X github.com/itchyny/$(BIN)-go/cli.revision=$(CURRENT_REVISION)"
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: all
all: build

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) ./cmd/$(BIN)

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./...

.PHONY: show-version
show-version: $(GOBIN)/gobump
	@gobump show -r $(VERSION_PATH)

$(GOBIN)/gobump:
	@go install github.com/x-motemen/gobump/cmd/gobump@latest

.PHONY: cross
cross: $(GOBIN)/goxz CREDITS
	goxz -n $(BIN) -pv=v$(VERSION) -include _$(BIN) -build-ldflags=$(BUILD_LDFLAGS) ./cmd/$(BIN)

$(GOBIN)/goxz:
	go install github.com/Songmu/goxz/cmd/goxz@latest

CREDITS: $(GOBIN)/gocredits go.sum
	go mod tidy
	gocredits . > $@

$(GOBIN)/gocredits:
	go install github.com/Songmu/gocredits/cmd/gocredits@latest

.PHONY: test
test: build
	go test -v -race ./...

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint -set_exit_status ./...

$(GOBIN)/golint:
	go install golang.org/x/lint/golint@latest

.PHONY: clean
clean:
	rm -rf $(BIN) goxz CREDITS
	go clean

.PHONY: bump
bump: $(GOBIN)/gobump
ifneq ($(shell git status --porcelain),)
	$(error git workspace is dirty)
endif
ifneq ($(shell git rev-parse --abbrev-ref HEAD),master)
	$(error current branch is not master)
endif
	@gobump up -w "$(VERSION_PATH)"
	git commit -am "bump up version to $(VERSION)"
	git tag "v$(VERSION)"
	git push origin master
	git push origin "refs/tags/v$(VERSION)"

.PHONY: upload
upload: $(GOBIN)/ghr
	ghr "v$(VERSION)" goxz

$(GOBIN)/ghr:
	go install github.com/tcnksm/ghr@latest
