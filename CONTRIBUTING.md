# Contributing to logparse

Thanks for your interest in contributing! This document will help you get started.

## Development Setup

1. Install Go 1.24 or later
2. Clone the repository
3. Run `go mod download` to install dependencies
4. Run `make test` to verify everything works

## Project Structure

```
logparse/
├── cmd/logparse/        # CLI entry point
├── internal/            # Private application code
│   ├── parser/          # Log parsing engine
│   ├── filter/          # Entry filtering
│   ├── output/          # Output formatters
│   └── cli/             # CLI command setup
├── pkg/api/             # Public API types and interfaces
└── testdata/            # Test fixtures
```

## Making Changes

1. Create a feature branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Run `make test` and `make lint`
5. Submit a pull request

## Code Style

- Follow standard Go conventions
- Use `gofmt` to format code
- Run `golangci-lint run` before submitting
- Write tests for all new features

## Testing

```bash
# Run all tests
make test

# Run benchmarks
make bench

# Run with race detector
go test -race ./...
```

## Release Process

Releases are automated via GoReleaser. To create a release:

1. Tag the version: `git tag -a v0.1.0 -m "Release v0.1.0"`
2. Push the tag: `git push origin v0.1.0`
3. GitHub Actions will build and publish the release

## Reporting Issues

- Use GitHub Issues to report bugs or request features
- Include steps to reproduce for bugs
- Specify your Go version and operating system
