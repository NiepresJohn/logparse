package output

import (
	"encoding/json"
	"io"

	"github.com/niepres/logparse/pkg/api"
)

// JSONOutput formats log entries as JSON, one per line (JSONL).
type JSONOutput struct {
	encoder *json.Encoder
}

// NewJSONOutput creates a new JSON outputter.
func NewJSONOutput() *JSONOutput {
	return &JSONOutput{}
}

// Write outputs a single entry as a JSON line.
func (o *JSONOutput) Write(w io.Writer, entry api.LogEntry) error {
	if o.encoder == nil {
		o.encoder = json.NewEncoder(w)
	}
	return o.encoder.Encode(entry)
}

// Flush is a no-op for JSON streaming.
func (o *JSONOutput) Flush(w io.Writer) error {
	return nil
}
