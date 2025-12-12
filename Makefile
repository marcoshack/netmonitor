.PHONY: build test coverage clean

# Detect OS
ifeq ($(OS),Windows_NT)
	# Windows commands
MKDIR = if not exist build\test mkdir build\test
RM = if exist
RM_FLAGS = del /q
RM_REC = if exist build\test rmdir /s /q build\test
PATH_SEP = \\
COVERAGE_OUT = build\test\coverage.out
COVERAGE_HTML = build\test\coverage.html
else
	# POSIX commands (Linux/Mac)
MKDIR = mkdir -p build/test
RM = rm -f
RM_FLAGS = 
RM_REC = rm -rf build/test
PATH_SEP = /
COVERAGE_OUT = build/test/coverage.out
COVERAGE_HTML = build/test/coverage.html
endif

all: coverage build

# Build the application
build:
	wails build

# Run tests
test:
	go test ./...

# Generate coverage report
coverage:
	$(MKDIR)
	go test ./... -coverprofile=$(COVERAGE_OUT)
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)

# Clean build artifacts
clean:
	go clean
	$(RM_REC)
