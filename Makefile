BIN := base58
VERSION := $$(make -s show-version)
VERSION_PATH := cmd/$(BIN)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION)"
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: all
all: build

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) ./cmd/$(BIN)

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./cmd/$(BIN)

.PHONY: show-version
show-version: $(GOBIN)/gobump
	@gobump show -r "$(VERSION_PATH)"

$(GOBIN)/gobump:
	@go install github.com/x-motemen/gobump/cmd/gobump@latest

.PHONY: cross
cross: $(GOBIN)/goxz CREDITS
	goxz -n $(BIN) -pv=v$(VERSION) -include _$(BIN) \
		-build-ldflags=$(BUILD_LDFLAGS) ./cmd/$(BIN)

$(GOBIN)/goxz:
	go install github.com/Songmu/goxz/cmd/goxz@latest

CREDITS: $(GOBIN)/gocredits go.sum
	go mod tidy
	gocredits -w .

$(GOBIN)/gocredits:
	go install github.com/Songmu/gocredits/cmd/gocredits@latest

.PHONY: test
test: build
	go test -v -race ./...

.PHONY: lint
lint: $(GOBIN)/staticcheck
	go vet ./...
	staticcheck -checks all ./...

$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: clean
clean:
	rm -rf $(BIN) goxz CREDITS
	go clean

.PHONY: bump
bump: $(GOBIN)/gobump
	test -z "$$(git status --porcelain || echo .)"
	test "$$(git branch --show-current)" = "main"
	@gobump up -w "$(VERSION_PATH)"
	git commit -am "bump up version to $(VERSION)"
	git tag "v$(VERSION)"
	git push --atomic origin main tag "v$(VERSION)"
