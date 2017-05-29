.DEFAULT_GOAL := build

test:
	go test -v -race $(shell go list ./... | grep -v /vendor/)

# The build targets allow to build the binary and docker image
.PHONY: build build.docker

BINARY        ?= hrp
SOURCES        = $(shell find . -name '*.go')
IMAGE         ?= zlangbert/$(BINARY)
VERSION       ?= $(shell git describe --tags --always --dirty)
BUILD_FLAGS   ?= -v
LDFLAGS       ?= -X github.com/zlangbert/hrp/config.version=$(VERSION) -w -s

build: build/$(BINARY)

build/$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build.docker: build/$(BINARY)
	docker build --rm --tag "$(IMAGE):$(VERSION)" .

build.push: build.docker
	docker push "$(IMAGE):$(VERSION)"

clean:
	@rm -rf build