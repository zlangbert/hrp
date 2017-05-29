.DEFAULT_GOAL := build

deps:
	dep ensure

test:
	go test -v -race $(shell go list ./... | grep -v /vendor/)

# The build targets allow to build the binary and docker image
.PHONY: build build.docker

BINARY        ?= hrp
SOURCES        = $(shell find . -name '*.go')
IMAGE         ?= quay.io/zlangbert/$(BINARY)
VERSION       ?= $(shell git describe --tags --always --dirty)
BUILD_FLAGS   ?= 
LDFLAGS       ?= -X github.com/zlangbert/hrp/config.version=$(VERSION) -w -s

build: build/$(BINARY)

build/$(BINARY): $(SOURCES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build.docker: build/$(BINARY)
	#docker build --rm --tag "$(IMAGE):$(VERSION)" .
	docker build --rm --tag "$(IMAGE):master" .

build.push: build.docker
	#docker push "$(IMAGE):$(VERSION)"
	docker push "$(IMAGE):master"

clean:
	@rm -rf build