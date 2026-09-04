package output

import (
	"encoding/csv"
	"io"
	"time"

	"github.com/niepres/logparse/pkg/api"
)

// CSVOutput formats log entries as CSV.
type CSVOutput struct {
	writer  *csv.Writer
	headers bool
}

// NewCSVOutput creates a new CSV outputter.
func NewCSVOutput() *CSVOutput {
	return &CSVOutput{}
}

// Write outputs a single entry as a CSV row.
func (o *CSVOutput) Write(w io.Writer, entry api.LogEntry) error {
	if o.writer == nil {
		o.writer = csv.NewWriter(w)
	}

	if !o.headers {
		if err := o.writer.Write([]string{"timestamp", "level", "message", "source"}); err != nil {
			return err
		}
		o.headers = true
	}

	timestamp := ""
	if !entry.Timestamp.IsZero() {
		timestamp = entry.Timestamp.Format(time.RFC3339)
	}

	record := []string{
		timestamp,
		entry.Level,
		entry.Message,
		entry.Source,
	}

	return o.writer.Write(record)
}

// Flush ensures all data is written.
func (o *CSVOutput) Flush(w io.Writer) error {
	if o.writer != nil {
		o.writer.Flush()
		return o.writer.Error()
	}
	return nil
}
