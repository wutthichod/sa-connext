PROTO_DIR := proto
PROTO_SRC := $(wildcard $(PROTO_DIR)/*.proto)
GO_OUT := .

.PHONY: generate-proto
generate-proto:
	protoc \
        --proto_path=$(PROTO_DIR) \
        --go_out=$(GO_OUT) \
        --go-grpc_out=$(GO_OUT) \
        $(PROTO_SRC)


# --- Configuration ---
BIN_DIR := bin

# Service Paths
SERVICE_A_PATH := services/api-gateway
SERVICE_B_PATH := services/user-service
SERVICE_C_PATH := services/chat-service
SERVICE_D_PATH := services/event-service
SERVICE_E_PATH := services/notification-service

# Service Names (Binaries will be named after these)
SERVICE_A_NAME := api-gateway
SERVICE_B_NAME := user-services
SERVICE_C_NAME := chat-service
SERVICE_D_NAME := event-service
SERVICE_E_NAME := notification-service

# Binary Paths
BIN_A := $(BIN_DIR)/$(SERVICE_A_NAME)
BIN_B := $(BIN_DIR)/$(SERVICE_B_NAME)
BIN_C := $(BIN_DIR)/$(SERVICE_C_NAME)
BIN_D := $(BIN_DIR)/$(SERVICE_D_NAME)
BIN_E := $(BIN_DIR)/$(SERVICE_E_NAME)

# List of all binaries
BINARIES := $(BIN_A) $(BIN_B) $(BIN_C) $(BIN_D) $(BIN_E)

# --- Main Targets ---

.PHONY: all build run clean stop stop-unix

# Target to build and run all
all: build run
 
# 1. Target to build all binaries
# This target depends on the list of all binaries.
# Make will find the specific rules below to build each one.
build: $(BINARIES)
	@echo "--- All services built successfully in $(BIN_DIR)/ ---"

# 2. Target to run all programs concurrently
run: build
	@echo "\n--- Running all compiled services concurrently ---"
	@echo "Starting $(SERVICE_A_NAME)..."
	./$(BIN_A) &
	@echo "Starting $(SERVICE_B_NAME) (Cobra)..."
	./$(BIN_B) serve &
	@echo "Starting $(SERVICE_C_NAME)..."
	./$(BIN_C) &
	@echo "Starting $(SERVICE_D_NAME)..."
	./$(BIN_D) &
	@echo "Starting $(SERVICE_E_NAME)..."
	./$(BIN_E) &
	@echo "\nAll services started. Press Ctrl+C to stop."
	@wait

# 4. Target to clean up compiled binaries and directory
clean:
	@echo "--- Cleaning up binaries ---"
	rm -rf $(BIN_DIR)
	@echo "Cleanup complete."

# 5. Target to stop all running services (Windows Version)
# NOTE: The EXEC_EXT variable was removed as it's not used here.
stop:
	@echo "--- Stopping all services on Windows (using taskkill) ---"
	@cmd /c "taskkill /F /IM $(SERVICE_A_NAME) /T 2>nul" || true
	@cmd /c "taskkill /F /IM $(SERVICE_B_NAME) /T 2>nul" || true
	@cmd /c "taskkill /F /IM $(SERVICE_C_NAME) /T 2>nul" || true
	@cmd /c "taskkill /F /IM $(SERVICE_D_NAME) /T 2>nul" || true
	@cmd /c "taskkill /F /IM $(SERVICE_E_NAME) /T 2>nul" || true
	@echo "All services stopped."

# 6. Target to stop all running services (macOS/Linux Version)
# Use 'make stop-unix' to run this target.
stop-unix:
	@echo "--- Stopping all services on macOS/Linux (using pkill -f) ---"
    # pkill -f matches the full command line, which includes the service name.
    # The '-' prefix ignores the error if the process is not found.
	-@pkill -f $(SERVICE_A_NAME) || true
	-@pkill -f $(SERVICE_B_NAME) || true
	-@pkill -f $(SERVICE_C_NAME) || true
	-@pkill -f $(SERVICE_D_NAME) || true
	-@pkill -f $(SERVICE_E_NAME) || true
	@echo "All services stopped."


# --- Utility & Build Rules ---

# 3. Utility to ensure the binary directory exists
# This is an "order-only prerequisite" for the binaries
$(BIN_DIR):
	@echo "--- Creating binary directory: $(BIN_DIR) ---"
	mkdir -p $(BIN_DIR)

# Pattern rules for building each binary
# These rules are triggered by the 'build' target's dependencies
# The '| $(BIN_DIR)' ensures the directory is created before trying to build.
$(BIN_A): | $(BIN_DIR)
	@echo "Building $(SERVICE_A_NAME) from $(SERVICE_A_PATH)..."
	go build -o $@ ./$(SERVICE_A_PATH)

$(BIN_B): | $(BIN_DIR)
	@echo "Building $(SERVICE_B_NAME) from $(SERVICE_B_PATH)..."
	go build -o $@ ./$(SERVICE_B_PATH)

$(BIN_C): | $(BIN_DIR)
	@echo "Building $(SERVICE_C_NAME) from $(SERVICE_C_PATH)..."
	go build -o $@ ./$(SERVICE_C_PATH)

$(BIN_D): | $(BIN_DIR)
	@echo "Building $(SERVICE_D_NAME) from $(SERVICE_D_PATH)..."
	go build -o $@ ./$(SERVICE_D_PATH)

$(BIN_E): | $(BIN_DIR)
	@echo "Building $(SERVICE_E_NAME) from $(SERVICE_E_PATH)..."
	go build -o $@ ./$(SERVICE_E_PATH)