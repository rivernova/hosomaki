# config
BINARY     := hosomaki
CMD_DIR    := .
BUILD_DIR  := ./bin
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"

OLLAMA_URL := http://localhost:11434

# default
.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# build
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)
	@echo "✓ Built $(BUILD_DIR)/$(BINARY) ($(VERSION))"

.PHONY: install
install: build 
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "✓ Installed to /usr/local/bin/$(BINARY)"

.PHONY: uninstall
uninstall: 
	sudo rm -f /usr/local/bin/$(BINARY)
	@echo "✓ Uninstalled $(BINARY)"

# dev
.PHONY: run
run: check-ollama build 
	$(BUILD_DIR)/$(BINARY) $(ARGS)

.PHONY: dev
dev: check-ollama
	go run $(LDFLAGS) $(CMD_DIR) $(ARGS)

# ollama
.PHONY: check-ollama
check-ollama: 
	@curl -sf $(OLLAMA_URL) > /dev/null 2>&1 || \
		(echo "✗ Ollama is not running at $(OLLAMA_URL)"; \
		 echo "  Start it with: ollama serve"; \
		 exit 1)
	@echo "✓ Ollama is running"

.PHONY: ollama-start
ollama-start: 
	@curl -sf $(OLLAMA_URL) > /dev/null 2>&1 && \
		echo "✓ Ollama already running" || \
		(ollama serve &> /tmp/ollama.log & \
		 sleep 2 && \
		 curl -sf $(OLLAMA_URL) > /dev/null 2>&1 && \
		 echo "✓ Ollama started" || \
		 (echo "✗ Failed to start Ollama. Check /tmp/ollama.log"; exit 1))

.PHONY: ollama-stop
ollama-stop: 
	@pkill -f "ollama serve" && echo "✓ Ollama stopped" || echo "Ollama was not running"

.PHONY: ollama-status
ollama-status: 
	@curl -sf $(OLLAMA_URL) > /dev/null 2>&1 && \
		(echo "✓ Ollama running at $(OLLAMA_URL)"; ollama list) || \
		echo "✗ Ollama not running"

# test
.PHONY: test
test:
	go test ./... -v

.PHONY: test-cover
test-cover: 
	go test ./... -coverprofile=coverage.txt
	go tool cover -html=coverage.txt -o coverage.html
	@echo "✓ Coverage report: coverage.html"

.PHONY: lint
lint: 
	golangci-lint run ./...

.PHONY: fmt
fmt: 
	gofmt -w .
	goimports -w .

.PHONY: vet
vet: 
	go vet ./...

# deps
.PHONY: deps
deps:
	go mod download
	go mod tidy

# clean
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) coverage.txt coverage.html
	@echo "✓ Cleaned"
