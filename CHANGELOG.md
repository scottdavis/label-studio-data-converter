# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-09-22

### Added
- Initial release of Label Studio to YOLO converter
- Convert Label Studio exports to YOLO training data format
- Support for train/validation dataset splits (configurable ratio)
- Automatic YOLO directory structure creation
- YAML configuration file generation for YOLO
- Label validation and statistics reporting
- Cross-platform support (Linux, Windows, macOS)
- Support for multiple architectures (amd64, arm64)
- Comprehensive test suite with 68%+ coverage
- Command-line interface with configurable options
- Reproducible splits with seeded random number generation
- Support for multiple image formats (jpg, jpeg, png, bmp, tiff, webp)
- Automated CI/CD with GitHub Actions
- Pre-built binaries for all major platforms

### Features
- **Single class detection**: Currently optimized for single-class object detection (books)
- **Fast performance**: Go implementation provides 3-5x speed improvement over Python
- **Memory efficient**: Low memory footprint for large datasets
- **Portable**: Single binary with no external dependencies
- **Validated output**: Automatic checking of label format and coordinate normalization

### Technical Details
- Go 1.21+ required for building
- Uses `gopkg.in/yaml.v3` for YAML generation
- Cross-compilation support for multiple platforms
- Automated testing and benchmark suite
- SHA256 checksums for release verification

### Usage
```bash
# Basic usage
./labelstudio-to-yolo -source . -output ./yolo_dataset

# Custom split ratio
./labelstudio-to-yolo -train-split 0.7

# With custom seed for reproducible splits
./labelstudio-to-yolo -seed 123
```

### Supported Platforms
- Linux (x64, ARM64)
- Windows (x64)
- macOS (Intel, Apple Silicon)

[Unreleased]: https://github.com/yourusername/labelstudio-to-yolo/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yourusername/labelstudio-to-yolo/releases/tag/v1.0.0
