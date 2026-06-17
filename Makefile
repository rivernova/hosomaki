# config
BINARY    := hosomaki
CMD_DIR   := .
BUILD_DIR := ./bin
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-X main.version=$(VERSION)"
OLLAMA_URL := http://localhost:11434

# default
.DEFAULT_GOAL := help

.PHONY: help
help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# build

.PHONY: build
build: ## build binary to ./bin/hosomaki
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)
	@echo "✓ Built $(BUILD_DIR)/$(BINARY) ($(VERSION))"

.PHONY: install
install: build ## install binary to /usr/local/bin
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "✓ Installed to /usr/local/bin/$(BINARY)"

.PHONY: uninstall
uninstall: ## remove binary from /usr/local/bin
	sudo rm -f /usr/local/bin/$(BINARY)
	@echo "✓ Uninstalled $(BINARY)"

.PHONY: setup
setup: install ## install hosomaki and print next steps for Ollama
	@echo ""
	@echo "Next: make sure Ollama is running and pull a model:"
	@echo "  ollama serve"
	@echo "  ollama pull llama3.1:8b"
	@echo ""
	@echo "Then try:"
	@echo "  hosomaki status"

# dev

.PHONY: run
run: check-ollama build ## build and run with ARGS= (e.g. make run ARGS='status')
	$(BUILD_DIR)/$(BINARY) $(ARGS)

.PHONY: dev
dev: check-ollama ## run without building (go run) with ARGS=
	go run $(LDFLAGS) $(CMD_DIR) $(ARGS)

# ollama

.PHONY: check-ollama
check-ollama: ## check that Ollama is reachable
	@curl -sf $(OLLAMA_URL) > /dev/null 2>&1 || \
		(echo "✗ Ollama is not running at $(OLLAMA_URL)"; \
		 echo "  Start it with: ollama serve"; \
		 exit 1)
	@echo "✓ Ollama is running"

.PHONY: ollama-start
ollama-start: ## start Ollama in the background if not already running
	@curl -sf $(OLLAMA_URL) > /dev/null 2>&1 && \
		echo "✓ Ollama already running" || \
		(ollama serve &> /tmp/ollama.log & \
		 sleep 2 && \
		 curl -sf $(OLLAMA_URL) > /dev/null 2>&1 && \
		 echo "✓ Ollama started" || \
		 (echo "✗ Failed to start Ollama — check /tmp/ollama.log"; exit 1))

.PHONY: ollama-stop
ollama-stop: ## stop the background Ollama process
	@pkill -f "ollama serve" && echo "✓ Ollama stopped" || echo "Ollama was not running"

.PHONY: ollama-status
ollama-status: ## show Ollama status and loaded models
	@curl -sf $(OLLAMA_URL) > /dev/null 2>&1 && \
		(echo "✓ Ollama running at $(OLLAMA_URL)"; ollama list) || \
		echo "✗ Ollama not running"

# test and quality

.PHONY: test
test: ## run tests
	go test ./...

.PHONY: test-verbose
test-verbose: ## run tests with verbose output
	go test ./... -v

.PHONY: test-cover
test-cover: ## run tests and generate HTML coverage report
	go test ./... -coverprofile=coverage.txt
	go tool cover -html=coverage.txt -o coverage.html
	@echo "✓ Coverage report: coverage.html"

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## format code with gofmt and goimports
	gofmt -w .
	goimports -w .

.PHONY: vet
vet: ## run go vet
	go vet ./...

# deps

.PHONY: deps
deps: ## download and tidy dependencies
	go mod download
	go mod tidy

# clean

.PHONY: clean
clean: ## remove build artifacts and coverage files
	rm -rf $(BUILD_DIR) coverage.txt coverage.html
	@echo "✓ Cleaned"

# packaging

PKG_ARCHES   := amd64 arm64
DIST_DIR     := ./dist
NFPM_VERSION := $(VERSION:v%=%)

.PHONY: package
package: $(addprefix package-,$(PKG_ARCHES)) ## build .deb and .rpm packages for all target arches

.PHONY: package-%
package-%: ## build .deb and .rpm for a single arch, e.g. make package-amd64
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=$* go build $(LDFLAGS) -o $(BUILD_DIR)/hosomaki-linux-$* $(CMD_DIR)
	@ln -sf hosomaki-linux-$* $(BUILD_DIR)/hosomaki
	GOARCH=$* VERSION=$(NFPM_VERSION) nfpm package \
		--config .nfpm.yaml \
		--packager deb \
		--target $(DIST_DIR)/hosomaki_$(NFPM_VERSION)_$*.deb
	GOARCH=$* VERSION=$(NFPM_VERSION) nfpm package \
		--config .nfpm.yaml \
		--packager rpm \
		--target $(DIST_DIR)/hosomaki_$(NFPM_VERSION)_$*.rpm
	@rm -f $(BUILD_DIR)/hosomaki
	@echo "✓ Packaged $* ($(NFPM_VERSION))"

.PHONY: package-clean
package-clean: ## remove packaging artifacts
	rm -rf $(DIST_DIR)
	@echo "✓ Cleaned packages"