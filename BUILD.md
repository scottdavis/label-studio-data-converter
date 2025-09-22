# Build Instructions

## Quick Build

```bash
# Build for current platform
go build -o labelstudio-to-yolo

# Run
./labelstudio-to-yolo -help
```

## Using Makefile

```bash
# Download dependencies
make deps

# Build for current platform  
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run with coverage
make test-coverage
```

## Cross-Platform Builds

```bash
# Linux (64-bit)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o labelstudio-to-yolo_linux

# Windows (64-bit)  
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o labelstudio-to-yolo.exe

# macOS (64-bit)
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o labelstudio-to-yolo_macos

# macOS (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o labelstudio-to-yolo_macos_arm64
```

## Prerequisites

- Go 1.21+
- No other dependencies required

## Usage

```bash
# Basic usage
./labelstudio-to-yolo

# Custom options
./labelstudio-to-yolo -source /path/to/data -output /path/to/yolo -train-split 0.7
```
