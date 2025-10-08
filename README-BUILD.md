# NetMonitor Build Environment Setup

This guide provides instructions for setting up the build environment for NetMonitor on Linux, Windows, and macOS.

## Prerequisites

Before you begin, you need to install the following tools:

- **Go**: The backend is written in Go.
- **Node.js**: The frontend is a web-based application that requires Node.js and npm.
- **Wails**: The framework used to build the cross-platform desktop application.
- **Make**: A build automation tool used to simplify the build and test process.

### Installation Guides

- **Go**: [Official Installation Guide](https://golang.org/doc/install)
- **Node.js**: [Official Installation Guide](https://nodejs.org/en/download/)
- **Wails**: [Official Installation Guide](https://wails.io/docs/gettingstarted/installation)
- **Make**:
  - **Linux**: Make is usually pre-installed. If not, you can install it using your distribution's package manager (e.g., `sudo apt-get install build-essential` on Debian/Ubuntu).
  - **Windows**: Make can be installed using Chocolatey (`choco install make`) or by installing [msys2](https://www.msys2.org/) and adding it to your PATH.
  - **macOS**: Make is included with the Xcode Command Line Tools. You can install it by running `xcode-select --install` in your terminal.

## Using the Makefile

Once you have all the prerequisites installed, you can use the `Makefile` to build and test the project. The `Makefile` provides several targets to simplify the development workflow.

### Makefile Targets

Here are the available targets:

- `make all`: Builds the application for production. This is the default target.
- `make build`: Builds the application for production, creating a distributable package.
- `make dev`: Runs the application in development mode with hot reload for frontend changes.
- `make test`: Runs all the tests for the Go backend.
- `make frontend-install`: Installs the frontend dependencies using npm.
- `make frontend-build`: Builds the frontend for production.
- `make frontend-dev`: Runs the frontend in development mode.
- `make clean`: Cleans up the project by removing build artifacts and frontend dependencies.
- `make help`: Displays a help message with all the available targets.

### Common Workflows

- **To build the application for the first time**:
  ```sh
  make frontend-install
  make build
  ```

- **To run the application in development mode**:
  ```sh
  make dev
  ```

- **To run the tests**:
  ```sh
  make test
  ```
