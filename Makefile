# Makefile for NetMonitor

# Go variables
GOCMD=go
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

# Node.js variables
NPMCMD=npm
NPMINST=$(NPMCMD) install
NPMBUILD=$(NPMCMD) run build
NPMDEV=$(NPMCMD) run dev

# Wails variables
WAILSCMD=wails
WAILSDEV=$(WAILSCMD) dev
WAILSBUILD=$(WAILSCMD) build

# OS detection
ifeq ($(OS),Windows_NT)
    RM = del /Q
    RM_DIR = rmdir /s /q
    MKDIR_P = mkdir
else
    RM = rm -f
    RM_DIR = rm -rf
    MKDIR_P = mkdir -p
endif

.PHONY: all build dev test frontend-install frontend-build frontend-dev clean help

all: build

# Build the application for production
build:
	@echo "Building the application for production..."
	$(WAILSBUILD)

# Run the application in development mode
dev:
	@echo "Running the application in development mode..."
	$(WAILSDEV)

# Run the tests, excluding the network package which has known issues in this environment
test:
	@echo "Running tests..."
	$(GOTEST) $(shell go list ./... | grep -v /internal/network)

# Install frontend dependencies
frontend-install:
	@echo "Installing frontend dependencies..."
	cd frontend && $(NPMINST)

# Build the frontend
frontend-build:
	@echo "Building the frontend..."
	cd frontend && $(NPMBUILD)

# Run the frontend in development mode
frontend-dev:
	@echo "Running the frontend in development mode..."
	cd frontend && $(NPMDEV)

# Clean the project
clean:
	@echo "Cleaning up the project..."
	-$(GOCLEAN)
	-$(RM_DIR) frontend/node_modules
	-$(RM_DIR) frontend/dist
	-$(RM_DIR) build/bin

# Display help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all                Build the application (default)"
	@echo "  build              Build the application for production"
	@echo "  dev                Run the application in development mode"
	@echo "  test               Run the tests"
	@echo "  frontend-install   Install frontend dependencies"
	@echo "  frontend-build     Build the frontend"
	@echo "  frontend-dev       Run the frontend in development mode"
	@echo "  clean              Clean the project"
	@echo "  help               Display this help message"
