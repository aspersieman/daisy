APP_NAME=daisy
SRC=main.go db.go models.go
BIN_DIR=bin
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
BIN_PATH=/usr/local/bin/$(APP_NAME)
ARCH=$(shell go env GOARCH)
SERVICE_PATH=/etc/systemd/system/$(APP_NAME).service
define SERVICE_DEFINITION
[Unit]
Description=Wellness Tracker API
After=network.target

[Service]
ExecStart=$(BIN_PATH)
WorkingDirectory=$(shell pwd)
Restart=always
User=$(shell whoami)
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
endef

.PHONY: all dev build install uninstall clean

all: build

## Development mode (requires air: https://github.com/air-verse/air)
dev:
	@command -v air >/dev/null 2>&1 || { echo >&2 "Air is not installed. Run 'go install github.com/air-verse/air@latest'"; exit 1; }
	air

## Build binary
build:
	@echo "Building..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=$(ARCH) go build -o $(BIN_DIR)/$(APP_NAME) $(SRC)
	@echo "Built binary at $(BIN_DIR)/$(APP_NAME)"

## Install to system path
export SERVICE_DEFINITION
install: build
	@echo "Installing $(APP_NAME)..."
	sudo cp $(BIN_DIR)/$(APP_NAME) $(BIN_PATH)
	@echo "Creating systemd service..."
	echo "$$SERVICE_DEFINITION" | sudo tee $(SERVICE_PATH) > /dev/null
	sudo systemctl daemon-reexec
	sudo systemctl daemon-reload
	sudo systemctl enable $(APP_NAME)
	sudo systemctl restart $(APP_NAME)
	@echo "$(APP_NAME) service installed and started."

## Uninstall from system path
uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	sudo systemctl stop $(APP_NAME)
	sudo systemctl disable $(APP_NAME)
	sudo rm -f $(SERVICE_PATH)
	sudo rm -f $(BIN_PATH)
	sudo systemctl daemon-reexec
	sudo systemctl daemon-reload
	@echo "$(APP_NAME) service removed."

## Clean build files
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@echo "Cleaned."
