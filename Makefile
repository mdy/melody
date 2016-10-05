VERSION=$(patsubst v%,%,$(shell git describe --tags))
LDFLAGS=-ldflags "-X 'main.version=$(VERSION) ($(shell date -u +%Y-%m-%d\ %H:%M:%S))'"
GOXFLAGS=-output "build/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}"
default: build

describe:
	@go run $(LDFLAGS) main.go -version

build:
	@go build $(LDFLAGS) -v .

test:
	@go test $(LDFLAGS) $(shell go list ./... | grep -v /vendor/)

release:
	@gox $(LDFLAGS) $(GOXFLAGS) .

install:
	git submodule update --init --recursive
	go get github.com/jteeuwen/go-bindata/...
