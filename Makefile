PROJECT_DIR := ${shell dirname ${abspath ${lastword ${MAKEFILE_LIST}}}}
LOCALBIN ?= ${PROJECT_DIR}/bin
${LOCALBIN}:
	mkdir -p ${LOCALBIN}

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: vet golangci-lint ## Lint source code.
	${GOLANGCILINT} run -v --timeout 4m0s ./...

.PHONY: build
build: ## Build manager binary
	CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/cga-led main.go

.PHONY: build-arm
build-arm: ## Build manager binary
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/cga-led-arm main.go

.PHONY: golangci-lint
GOLANGCILINT := ${LOCALBIN}/golangci-lint
golangci-lint: ${GOLANGCILINT} ## Download golangci-lint
${GOLANGCILINT}: ${LOCALBIN}
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${LOCALBIN}