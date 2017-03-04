GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
COMMON_ARGS = -ldflags "-X main.GitCommit='$(GIT_COMMIT)$(GIT_DIRTY)'"
DEPS = $(wildcard *.go) $(wildcard ../*.go)

default: xenstore

.PHONY: all
all: xenstore xenstore-static

xenstore: $(DEPS)
	env GOOS=linux GOARCH=amd64 go build $(COMMON_ARGS) -tags linux -o xenstore .

xenstore-static:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -X main.GitCommit='$(GIT_COMMIT)$(GIT_DIRTY)'" -tags linux -o xenstore-static .

.PHONY: clean
clean:
	rm -rf xenstore xenstore-static
