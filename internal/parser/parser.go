package parser

import (
	"bufio"
	"fmt"
	"io"

	"github.com/niepres/logparse/pkg/api"
)

const (
	maxScanTokenSize = 1024 * 1024 // 1MB
	defaultBatchSize = 100
)

// lineParser is the interface for format-specific line parsers.
type lineParser interface {
	ParseLine(line string) (api.LogEntry, error)
}

// Parser is the core log parsing engine.
type Parser struct {
	format    api.FormatType
	strict    bool
	batchSize int
}

// New creates a new Parser with the given options.
func New(format api.FormatType, opts ...api.ParseOption) (*Parser, error) {
	options := &api.ParseOptions{}
	for _, opt := range opts {
		opt(options)
	}

	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	return &Parser{
		format:    format,
		strict:    options.Strict,
		batchSize: batchSize,
	}, nil
}

// Parse reads from r and returns channels for entries and errors.
// Implements the api.Parser interface.
func (p *Parser) Parse(r io.Reader) (<-chan api.LogEntry, <-chan error) {
	entries := make(chan api.LogEntry, p.batchSize)
	errors := make(chan error, 10)

	go func() {
		defer close(entries)
		defer close(errors)

		lp := p.getLineParser()
		scanner := bufio.NewScanner(r)
		buf := make([]byte, maxScanTokenSize)
		scanner.Buffer(buf, maxScanTokenSize)

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if line == "" {
				continue
			}

			entry, err := lp.ParseLine(line)
			if err != nil {
				if p.strict {
					errors <- &lineError{line: lineNum, err: err}
					return
				}
				continue
			}

			entries <- entry
		}

		if err := scanner.Err(); err != nil {
			errors <- err
		}
	}()

	return entries, errors
}

// ParseAndFilter reads from r, applies the filter, and sends matching entries to w via the outputter.
func (p *Parser) ParseAndFilter(r io.Reader, filter api.Filter, output api.Outputter, w io.Writer) (*api.ParseStats, error) {
	entries, errs := p.Parse(r)
	stats := &api.ParseStats{
		ByLevel: make(map[string]int),
	}

	for entry := range entries {
		stats.TotalLines++

		if filter != nil && !filter.Match(entry) {
			continue
		}

		stats.ParsedLines++
		if level := entry.Level; level != "" {
			stats.ByLevel[level]++
		}

		if err := output.Write(w, entry); err != nil {
			return stats, err
		}
	}

	// Drain error channel
	for err := range errs {
		stats.SkippedLines++
		_ = err // In non-strict mode, we just count skipped lines
	}

	if err := output.Flush(w); err != nil {
		return stats, err
	}

	return stats, nil
}

func (p *Parser) getLineParser() lineParser {
	switch p.format {
	case api.FormatJSON:
		return newJSONParser(p.strict)
	case api.FormatDocker:
		return newDockerParser(p.strict)
	case api.FormatNginx:
		return newNginxParser()
	case api.FormatSyslog:
		return newSyslogParser()
	case api.FormatPlain:
		return newPlainParser()
	default:
		return newPlainParser()
	}
}

// lineError represents an error at a specific line number.
type lineError struct {
	line int
	err  error
}

func (e *lineError) Error() string {
	return fmt.Sprintf("line %d: %v", e.line, e.err)
}

// Ensure Parser implements api.Parser
var _ api.Parser = (*Parser)(nil)
