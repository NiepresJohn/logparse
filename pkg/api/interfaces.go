package api

import "io"

// Parser defines the interface for log parsing implementations.
type Parser interface {
	// Parse reads from r and returns a channel of LogEntry.
	// The channel is closed when all input is consumed or an error occurs.
	// Errors are sent on the error channel before the entry channel is closed.
	Parse(r io.Reader) (<-chan LogEntry, <-chan error)
}

// ParserFactory creates a Parser for a given format.
type ParserFactory interface {
	NewParser(format FormatType) (Parser, error)
}

// Filter defines the interface for filtering log entries.
type Filter interface {
	// Match returns true if the entry passes the filter.
	Match(entry LogEntry) bool
}

// FilterFunc is a function type that implements Filter.
type FilterFunc func(LogEntry) bool

// Match implements Filter.
func (f FilterFunc) Match(entry LogEntry) bool {
	return f(entry)
}

// Outputter defines the interface for formatting log entries.
type Outputter interface {
	// Write outputs a single entry to w.
	Write(w io.Writer, entry LogEntry) error
	// Flush writes any remaining data (headers, footers) to w.
	Flush(w io.Writer) error
}

// ParseOption configures a Parser.
type ParseOption func(*ParseOptions)

// ParseOptions holds configuration for the Parser.
type ParseOptions struct {
	Strict    bool
	MaxLines  int
	BatchSize int
}

// WithStrict sets whether parsing should fail on malformed lines.
func WithStrict(strict bool) ParseOption {
	return func(o *ParseOptions) {
		o.Strict = strict
	}
}

// WithMaxLines sets the maximum number of lines to parse (0 = unlimited).
func WithMaxLines(n int) ParseOption {
	return func(o *ParseOptions) {
		o.MaxLines = n
	}
}

// WithBatchSize sets the channel buffer size for parsed entries.
func WithBatchSize(n int) ParseOption {
	return func(o *ParseOptions) {
		o.BatchSize = n
	}
}
