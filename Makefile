GOPATH=$(CURDIR)

all: build

build:
	GOPATH=$(GOPATH) go build

test:
	GOPATH=$(GOPATH) go test -v ./...

clean:
	GOPATH=$(GOPATH) go clean
