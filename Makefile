PROJECT_DIR := ${shell dirname ${abspath ${lastword ${MAKEFILE_LIST}}}}
LOCALBIN ?= ${PROJECT_DIR}/bin
${LOCALBIN}:
	mkdir -p ${LOCALBIN}

.PHONY: clean
clean:
	rm -rf bin

.PHONY: fmt
fmt: ## Run go fmt against code.
	go mod tidy
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
.PHONY: lint
lint: vet
	test -s $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN)
	CGO_ENABLED=0 $(GOLANGCI_LINT) run --timeout 5m

.PHONY: build
build: ## Build binary for host
	CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/cga-led main.go

.PHONY: build-arm
build-arm: ## Build binary for arm64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/cga-led-arm main.go
