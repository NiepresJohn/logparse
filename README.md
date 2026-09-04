# logparse

A fast, streaming CLI tool for parsing, filtering, and formatting log files.

[![CI](https://github.com/niepres/logparse/actions/workflows/ci.yml/badge.svg)](https://github.com/niepres/logparse/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/niepres/logparse)](https://goreportcard.com/report/github.com/niepres/logparse)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Auto-detects log formats** — JSON, Docker, Nginx, Syslog, plain text
- **Streaming** — constant memory usage, handles multi-GB files
- **Filter by level, time, regex, or fields**
- **Multiple outputs** — table, JSON (pipe-friendly), CSV
- **Unix philosophy** — pipes, stdin, composable with `jq`, `grep`, etc.
- **Shell completions** — Bash, Zsh, Fish, PowerShell

## Install

```bash
# Homebrew (macOS/Linux)
brew install niepres/tap/logparse

# Go
go install github.com/niepres/logparse/cmd/logparse@latest

# Docker
docker pull niepres/logparse

# Binary releases
# Download from https://github.com/niepres/logparse/releases
```

## Quick Start

```bash
# Parse a log file (auto-detects format)
logparse app.log

# Filter for errors only
logparse app.log --level ERROR

# Pipe from kubectl
kubectl logs pod-name | logparse --format docker --since 1h

# Output as JSON for jq
logparse app.log --output json | jq '.[] | select(.level=="ERROR")'

# Parse nginx logs with filter
logparse access.log --format nginx --grep "POST"

# Show summary statistics
logparse app.log --summary

# CSV for spreadsheet analysis
logparse app.log --output csv > report.csv
```

## Usage

```
logparse [files...] [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `auto` | Log format: `auto`, `json`, `docker`, `nginx`, `syslog`, `plain` |
| `-o, --output` | `table` | Output format: `table`, `json`, `csv` |
| `-l, --level` | | Minimum level: `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL` |
| `--since` | | Filter after time: `1h`, `30m`, or `2024-01-01` |
| `-g, --grep` | | Filter by regex pattern |
| `-F, --field` | | Filter by field=value (repeatable) |
| `--strict` | `false` | Fail on malformed lines |
| `--summary` | `false` | Show statistics instead of entries |

### Subcommands

| Command | Description |
|---------|-------------|
| `completion` | Generate shell completion script |
| `version` | Print version information |

## Examples

### Watch logs in real-time

```bash
tail -f /var/log/app.log | logparse --level ERROR
```

### Find slow requests in nginx logs

```bash
logparse access.log --format nginx --grep "api" --output json | jq '.[] | select(.fields.status | tonumber >= 400)'
```

### Compare error rates across services

```bash
logparse service-a.log service-b.log --level ERROR --summary
```

### Extract structured data

```bash
logparse app.log --output json > structured.json
```

### Docker logs

```bash
docker logs container 2>&1 | logparse --format docker --level WARN
```

## Log Formats

| Format | Detection | Example |
|--------|-----------|---------|
| JSON | Lines starting with `{` | `{"level":"INFO","message":"ok"}` |
| Docker | JSON with `log` + `time` fields | `{"log":"hello","time":"2024-01-15T10:30:00Z"}` |
| Nginx/Apache | IP + bracketed timestamp | `127.0.0.1 - - [15/Jan/2024:10:30:00 +0000] "GET / HTTP/1.1" 200` |
| Syslog | `<priority>` prefix | `<134>Jan 15 10:30:00 host app[1234]: message` |
| Plain | Fallback | Any text, infers level from keywords |

## As a Library

```go
import "github.com/niepres/logparse/internal/parser"

p, _ := parser.New(api.FormatJSON)
entries, errs := p.Parse(os.Stdin)

for entry := range entries {
    fmt.Println(entry.Level, entry.Message)
}
```

## Performance

- **Memory**: O(1) — streams line-by-line, never loads entire file
- **Speed**: ~100K lines/second on modern hardware
- **Binary size**: ~8MB static binary

## Build

```bash
make build    # Build binary
make test     # Run tests
make bench    # Run benchmarks
make lint     # Run linter
make release  # Cross-compile for all platforms
```

## Shell Completions

```bash
# Bash
source <(logparse completion bash)

# Zsh
source <(logparse completion zsh)

# Fish
logparse completion fish | source
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)
