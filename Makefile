# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOBUILDRACE=$(GOCMD) build -race
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BIN_NAME=redisgraph-benchmark-go
DISTDIR = ./dist
DIST_OS_ARCHS = "linux/amd64 linux/arm64 linux/arm windows/amd64 darwin/amd64 darwin/arm64"
LDFLAGS = "-X 'main.GitSHA1=$(GIT_SHA)' -X 'main.GitDirty=$(GIT_DIRTY)'"

# Build-time GIT variables
ifeq ($(GIT_SHA),)
GIT_SHA:=$(shell git rev-parse HEAD)
endif

ifeq ($(GIT_DIRTY),)
GIT_DIRTY:=$(shell git diff --no-ext-diff 2> /dev/null | wc -l)
endif

.PHONY: all test coverage build checkfmt fmt
all: test coverage build checkfmt fmt

build:
	$(GOBUILD) \
        -ldflags=$(LDFLAGS) .

build-race:
	$(GOBUILDRACE) \
        -ldflags=$(LDFLAGS) .

checkfmt:
	@echo 'Checking gofmt';\
 	bash -c "diff -u <(echo -n) <(gofmt -d .)";\
	EXIT_CODE=$$?;\
	if [ "$$EXIT_CODE"  -ne 0 ]; then \
		echo '$@: Go files must be formatted with gofmt'; \
	fi && \
	exit $$EXIT_CODE

lint:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run

fmt:
	$(GOFMT) ./...

get:
	$(GOGET) -t -v ./...

test: get
	$(GOFMT) ./...
	$(GOTEST) -race -covermode=atomic ./...

coverage: get test
	$(GOTEST) -race -coverprofile=coverage.txt -covermode=atomic .

flow-test: build-race
	./$(BIN_NAME) -n 100000 -query "CREATE(n)" -query-ratio 0.33 -query "MATCH (n) RETURN n LIMIT 1" -query-ratio 0.67

release:
	$(GOGET) github.com/mitchellh/gox
	$(GOGET) github.com/tcnksm/ghr
	GO111MODULE=on gox  -osarch ${DIST_OS_ARCHS} -output "${DISTDIR}/${BIN_NAME}_{{.OS}}_{{.Arch}}" \
	    -ldflags=$(LDFLAGS) .

publish: release
	@for f in $(shell ls ${DISTDIR}); \
	do \
	echo "copying ${DISTDIR}/$${f}"; \
	aws s3 cp ${DISTDIR}/$${f} s3://benchmarks.redislabs/tools/${BIN_NAME}/unstable/$${f} --acl public-read; \
	done
