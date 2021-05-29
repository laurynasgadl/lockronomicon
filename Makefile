GO 		    := go
GOHOSTOS    := $(shell $(GO) env GOHOSTOS)
GOHOSTARCH  := $(shell $(GO) env GOHOSTARCH)

MODULE := $(shell env GO111MODULE=on $(GO) list -m)

BINARY := lockronomicon

BUILD_VERSION  := $(shell git describe --tags --always --match="v*" 2> /dev/null)
BUILD_DATE     := $(shell date '+%FT%T')
BUILD_REVISION := $(shell git rev-parse HEAD)

BUILD_PLATFORMS := darwin/amd64 freebsd/amd64 linux/amd64 windows/amd64 linux/386

LDFLAGS=-ldflags "-X $(MODULE)/build.Name=$(BINARY) -X $(MODULE)/build.Version=$(BUILD_VERSION) -X $(MODULE)/build.Date=$(BUILD_DATE) -X $(MODULE)/build.Revision=$(BUILD_REVISION)"


.PHONY: build-docker
build-docker: build-docker-bin
	docker build -t $(BINARY):$(BUILD_VERSION) -f Dockerfile .

.PHONY: build-docker-bin
build-docker-bin:
	GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -v -o $(BINARY) ./cmd/$(BINARY)


temp = $(subst /, ,$@)
OS   = $(word 1, $(temp))
ARCH = $(word 2, $(temp))

.PHONY: build-all $(BUILD_PLATFORMS)
build-all: $(BUILD_PLATFORMS)

$(BUILD_PLATFORMS):
	GO111MODULE=on GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(LDFLAGS) -v -o bin/$(BINARY)-$(BUILD_VERSION).$(OS)-$(ARCH) ./cmd/$(BINARY)

BINARIES := $(shell cd bin && find *)

.PHONY: tarballs $(BINARIES)
tarballs: dist-folder $(BINARIES)

dist-folder:
	rm -rf dist
	mkdir dist

$(BINARIES):
	tar -C bin -czf "dist/$@.tar.gz" "$@"

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist