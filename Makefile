GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_DIRTY = $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)

COMMON_LDFLAGS = -X main.GitCommit='$(GIT_COMMIT)$(GIT_DIRTY)'
COMMON_ARGS = -ldflags "$(COMMON_LDFLAGS)"

DEPS = $(wildcard *.go) $(wildcard cmd/xenstore/*.go)

OUTPUTS = \
	xenstore-linux-amd64 \
	xenstore-linux-amd64-static

.PHONY: all
all: $(OUTPUTS)

xenstore-linux-amd64: $(DEPS)
	env GOOS=linux GOARCH=amd64 go build $(COMMON_ARGS) -tags linux -o $@ ./cmd/xenstore

xenstore-linux-amd64-static: $(DEPS)
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s $(COMMON_LDFLAGS)" -tags linux -o $@ ./cmd/xenstore

.PHONY: clean
clean:
	rm -rf $(OUTPUTS)
