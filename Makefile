APP := marchine

# Development builds derive their version from Git when available. Source
# archives and exported trees have no .git directory, so they fall back to a
# visible, non-empty development identifier. Release/package builds should
# override VERSION explicitly (for example: make build VERSION=1.22.0).
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || true)
VERSION ?= $(if $(strip $(GIT_VERSION)),$(strip $(GIT_VERSION)),dev)

LDFLAGS := -s -w -X main.version=$(VERSION)
GO ?= go
GO_SOURCES := $$(find cmd internal -name '*.go' -type f)

.PHONY: deps vendor-refresh build build-online test run demo install uninstall clean fmt fmt-check vet check

# Refresh/download dependencies only when intentionally updating the graph.
# Normal build/test/run use the committed vendor snapshot and work offline.
deps:
	$(GO) mod download

vendor-refresh:
	$(GO) mod tidy
	$(GO) mod vendor

build:
	mkdir -p dist
	CGO_ENABLED=0 $(GO) build -mod=vendor -trimpath -buildvcs=false -buildmode=pie -ldflags "$(LDFLAGS)" -o dist/$(APP) ./cmd/marchine

# Use the module graph directly when intentionally building without vendor/.
build-online:
	mkdir -p dist
	CGO_ENABLED=0 $(GO) build -trimpath -buildvcs=false -buildmode=pie -ldflags "$(LDFLAGS)" -o dist/$(APP) ./cmd/marchine

test:
	$(GO) test -mod=vendor -count=1 ./...

run:
	$(GO) run -mod=vendor -ldflags "$(LDFLAGS)" ./cmd/marchine

demo:
	$(GO) run -mod=vendor -ldflags "$(LDFLAGS)" ./cmd/marchine --demo

install: build
	install -Dm755 dist/$(APP) $(HOME)/.local/bin/$(APP)

uninstall:
	rm -f $(HOME)/.local/bin/$(APP)

fmt:
	gofmt -w $(GO_SOURCES)

fmt-check:
	@unformatted="$$(gofmt -l $(GO_SOURCES))"; \
	if [ -n "$$unformatted" ]; then \
		printf '%s\n' "Go files need gofmt:" "$$unformatted"; \
		exit 1; \
	fi

vet:
	$(GO) vet -mod=vendor ./...

check: fmt-check test vet

clean:
	rm -rf dist
