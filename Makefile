LDFLAGS=-ldflags "-X 'main.version=`git describe --tags` (`date -u +%Y-%m-%d\ %H:%M:%S`)'"
PACKAGES=$(shell go list ./... | grep -v /vendor/)
default: build

describe:
	@go run $(LDFLAGS) main.go -version

build:
	@go build $(LDFLAGS) -v .

test:
	@go test $(LDFLAGS) $(PACKAGES)
