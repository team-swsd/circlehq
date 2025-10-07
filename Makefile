# ==============================================================================
# Variables
# ==============================================================================

# Go commands
GO := go
GOBUILD := $(GO) build
GORUN := $(GO) run

# Source and Target files for WebAssembly
WASM_INDEX_SRC := wasm/index/main.go
WASM_INDEX_TARGET := internal/server/static/index.wasm

WASM_DASHBOARD_SRC := wasm/dashboard/main.go
WASM_DASHBOARD_TARGET := internal/server/static/dashboard.wasm

# Main application sources and binary target
MAIN_APP_SRC_DIR := cmd/circlehq
MAIN_APP_GO_FILES := $(wildcard $(MAIN_APP_SRC_DIR)/*.go)
MAIN_APP_TARGET := bin/circlehq

# Configuration file for the server
CONFIG_FILE := ./config.toml

# ==============================================================================
# Phony Targets (Targets that are not files)
# ==============================================================================

.PHONY: all run build wasm clean

# Default target when 'make' is run without arguments
all: build

# ==============================================================================
# Main Targets
# ==============================================================================

# Target: run
# Builds WASM files only if their sources have changed, then runs the server.
run: $(WASM_INDEX_TARGET) $(WASM_DASHBOARD_TARGET)
	@echo "==> Running server..."
	$(GORUN) $(MAIN_APP_SRC_DIR)/main.go serve -c $(CONFIG_FILE)

# Target: build
# Builds all WASM files and the main application binary.
build: wasm $(MAIN_APP_TARGET)
	@echo "==> Build complete."

# Helper target to build all WASM files
wasm: $(WASM_INDEX_TARGET) $(WASM_DASHBOARD_TARGET)

# ==============================================================================
# Build Rules
# ==============================================================================

# Rule to build the index WASM file.
# This command runs only if wasm/index/main.go is newer than the target WASM file.
$(WASM_INDEX_TARGET): $(WASM_INDEX_SRC)
	@echo "==> Building index WASM..."
	GOOS=js GOARCH=wasm $(GOBUILD) -o $@ $<

# Rule to build the dashboard WASM file.
# This command runs only if wasm/dashboard/main.go is newer than the target WASM file.
$(WASM_DASHBOARD_TARGET): $(WASM_DASHBOARD_SRC)
	@echo "==> Building dashboard WASM..."
	GOOS=js GOARCH=wasm $(GOBUILD) -o $@ $<

# Rule to build the main application binary.
$(MAIN_APP_TARGET): $(MAIN_APP_GO_FILES)
	@echo "==> Building main application binary..."
	@mkdir -p $(@D)
	$(GOBUILD) -o $@ $(MAIN_APP_SRC_DIR)/main.go

# ==============================================================================
# Utility Targets
# ==============================================================================

# Target: clean
# Removes all generated files.
clean:
	@echo "==> Cleaning up built artifacts..."
	@rm -f $(WASM_INDEX_TARGET) $(WASM_DASHBOARD_TARGET) $(MAIN_APP_TARGET)
	@rmdir bin 2>/dev/null || true
