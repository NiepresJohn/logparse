package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/niepres/logparse/pkg/api"
)

func TestNewOutputter(t *testing.T) {
	tests := []struct {
		name    string
		format  api.OutputType
		wantErr bool
	}{
		{"table", api.OutputTable, false},
		{"json", api.OutputJSON, false},
		{"csv", api.OutputCSV, false},
		{"invalid", api.OutputType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOutputter(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOutputter(%q) error = %v, wantErr %v", tt.format, err, tt.wantErr)
			}
		})
	}
}

func TestTableOutput(t *testing.T) {
	var buf bytes.Buffer
	out := NewTableOutput()

	entry := api.LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:     "ERROR",
		Message:   "Something went wrong",
	}

	if err := out.Write(&buf, entry); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := out.Flush(&buf); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("output should contain level, got: %s", output)
	}
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("output should contain message, got: %s", output)
	}
}

func TestTableOutput_LongMessage(t *testing.T) {
	var buf bytes.Buffer
	out := NewTableOutput()

	longMsg := strings.Repeat("x", 150)
	entry := api.LogEntry{
		Level:   "INFO",
		Message: longMsg,
	}

	if err := out.Write(&buf, entry); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := out.Flush(&buf); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, longMsg) {
		t.Error("long message should be truncated")
	}
	if !strings.Contains(output, "...") {
		t.Error("truncated message should contain ellipsis")
	}
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	out := NewJSONOutput()

	entry := api.LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:     "INFO",
		Message:   "Test message",
	}

	if err := out.Write(&buf, entry); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := out.Flush(&buf); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("output should contain JSON level, got: %s", output)
	}
	if !strings.Contains(output, `"message":"Test message"`) {
		t.Errorf("output should contain JSON message, got: %s", output)
	}
}

func TestCSVOutput(t *testing.T) {
	var buf bytes.Buffer
	out := NewCSVOutput()

	entry := api.LogEntry{
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Level:     "WARN",
		Message:   "Warning message",
		Source:    "test.log",
	}

	if err := out.Write(&buf, entry); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := out.Flush(&buf); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (header + data), got %d", len(lines))
	}
	if !strings.Contains(lines[0], "timestamp,level,message,source") {
		t.Errorf("header should contain field names, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "WARN") {
		t.Errorf("data should contain level, got: %s", lines[1])
	}
}

func TestCSVOutput_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	out := NewCSVOutput()

	entries := []api.LogEntry{
		{Level: "INFO", Message: "First"},
		{Level: "ERROR", Message: "Second"},
		{Level: "DEBUG", Message: "Third"},
	}

	for _, e := range entries {
		if err := out.Write(&buf, e); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}
	if err := out.Flush(&buf); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 { // header + 3 entries
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}
