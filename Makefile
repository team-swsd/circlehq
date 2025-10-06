# --- Variables ---
WASM_OUTPUT := internal/server/static/main.wasm
WASM_SOURCE := ./wasm/main.go
BINARY_OUTPUT := bin/circlehq
CMD_PATH := cmd/circlehq/main.go
PORT := 35080

# Phony targets are targets that do not produce an output file.
.PHONY: all build run clean

# Default target executed when you just run `make`
all: build

# Build both WASM and the server binary
build: $(BINARY_OUTPUT)

# Rule to build the server binary. It depends on the WASM file.
$(BINARY_OUTPUT): $(WASM_OUTPUT) $(CMD_PATH)
	@echo "==> Building server binary..."
	@mkdir -p $(dir $(BINARY_OUTPUT))
	go build -o $(BINARY_OUTPUT) $(CMD_PATH)
	@echo "==> Build complete: $(BINARY_OUTPUT)"

# Run the server (ensures WASM is built first)
run: $(WASM_OUTPUT)
	@echo "==> Running server on port $(PORT)..."
	go run $(CMD_PATH) serve --port $(PORT)

# Rule to build the WASM file.
# This rule is triggered if wasm/main.go is newer than the existing main.wasm
$(WASM_OUTPUT): $(WASM_SOURCE)
	@echo "==> Building WebAssembly binary..."
	@mkdir -p $(dir $@)
	GOOS=js GOARCH=wasm go build -o $@ $<

# Clean up build artifacts
clean:
	@echo "==> Cleaning up..."
	@rm -f $(WASM_OUTPUT)
	@rm -rf bin