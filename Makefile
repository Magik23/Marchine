APP := marchine

# Derive the development version from Git.
# Release builds can override this:
#   make build VERSION=1.22.0
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//')

LDFLAGS := -s -w -X main.version=$(VERSION)
GO := GOTOOLCHAIN=auto go

.PHONY: deps vendor-refresh build build-online test run demo install uninstall clean fmt vet

# Refresh/download dependencies only when intentionally updating the graph.
# Normal build/test/run use the vendored snapshot and work offline.
deps:
	$(GO) mod download

vendor-refresh:
	$(GO) mod tidy
	$(GO) mod vendor

build:
	mkdir -p dist
	CGO_ENABLED=0 $(GO) build -mod=vendor -trimpath -ldflags "$(LDFLAGS)" -o dist/$(APP) ./cmd/marchine

# Use the module graph directly when intentionally building without vendor/.
build-online:
	mkdir -p dist
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(APP) ./cmd/marchine

test:
	$(GO) test -mod=vendor ./...

run:
	$(GO) run -mod=vendor -ldflags "$(LDFLAGS)" ./cmd/marchine

demo:
	$(GO) run -mod=vendor -ldflags "$(LDFLAGS)" ./cmd/marchine --demo

install: build
	install -Dm755 dist/$(APP) $(HOME)/.local/bin/$(APP)

uninstall:
	rm -f $(HOME)/.local/bin/$(APP)

fmt:
	gofmt -w $$(find . -name '*.go' -type f -not -path './vendor/*')

vet:
	$(GO) vet -mod=vendor ./...

clean:
	rm -rf dist
