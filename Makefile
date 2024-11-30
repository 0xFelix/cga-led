PROJECT_DIR := ${shell dirname ${abspath ${lastword ${MAKEFILE_LIST}}}}
LOCALBIN ?= ${PROJECT_DIR}/bin
${LOCALBIN}:
	mkdir -p ${LOCALBIN}

.PHONY: clean
clean:
	rm -rf bin

.PHONY: fmt
fmt: gofumpt ## Run gofumt against code.
	go mod tidy -compat=1.23
	$(GOFUMPT) -w -extra .

.PHONY: vendor
vendor:
	go mod vendor

GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint

.PHONY: lint
lint:
	test -s $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN)
	CGO_ENABLED=0 $(GOLANGCI_LINT) run --timeout 5m

GOFUMPT ?= $(LOCALBIN)/gofumpt

.PHONY: gofumpt
gofumpt: $(GOFUMPT) ## Download gofumpt locally if necessary.
$(GOFUMPT): $(LOCALBIN)
	test -s $(LOCALBIN)/gofumpt || GOBIN=$(LOCALBIN) go install mvdan.cc/gofumpt@latest

.PHONY: build
build: ## Build binary for host
	CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/cga-led main.go

.PHONY: build-arm
build-arm: ## Build binary for arm64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/cga-led-arm main.go
