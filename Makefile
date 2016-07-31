all: clean build

build: deps
	go build -o build/base58 ./cmd/base58

install: deps
	go install ./...

deps:
	go get -d -v ./...

test: testdeps build
	go test -v ./...

testdeps:
	go get -d -v -t ./...
	go get -u github.com/golang/lint/golint

LINT_RET = .golint.txt
lint: testdeps
	go vet
	rm -f $(LINT_RET)
	golint ./... | tee $(LINT_RET)
	test ! -s $(LINT_RET)

GOFMT_RET = .gofmt.txt
gofmt: testdeps
	rm -f $(GOFMT_RET)
	gofmt -s -d *.go | tee $(GOFMT_RET)
	test ! -s $(GOFMT_RET)

clean:
	rm -rf build
	go clean

.PHONY: build install deps test testdeps lint gofmt clean
