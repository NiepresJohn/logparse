package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/niepres/logparse/pkg/api"
)

func TestJSONParser_ParseLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantErr  bool
		wantLvl  string
		wantMsg  string
	}{
		{
			name:    "valid json log",
			line:    `{"timestamp":"2024-01-15T10:30:00Z","level":"INFO","message":"test message"}`,
			wantLvl: "INFO",
			wantMsg: "test message",
		},
		{
			name:    "error level",
			line:    `{"level":"ERROR","message":"something failed"}`,
			wantLvl: "ERROR",
			wantMsg: "something failed",
		},
		{
			name:    "debug level",
			line:    `{"level":"DEBUG","message":"debug info"}`,
			wantLvl: "DEBUG",
			wantMsg: "debug info",
		},
		{
			name:    "alternative msg field",
			line:    `{"level":"WARN","msg":"warning text"}`,
			wantLvl: "WARN",
			wantMsg: "warning text",
		},
		{
			name:    "invalid json",
			line:    `not json at all`,
			wantErr: true,
		},
	}

	p := newJSONParser(false)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := p.ParseLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if entry.Level != tt.wantLvl {
					t.Errorf("Level = %q, want %q", entry.Level, tt.wantLvl)
				}
				if entry.Message != tt.wantMsg {
					t.Errorf("Message = %q, want %q", entry.Message, tt.wantMsg)
				}
			}
		})
	}
}

func TestNginxParser_ParseLine(t *testing.T) {
	line := `192.168.1.100 - - [15/Jan/2024:10:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234`
	p := newNginxParser()

	entry, err := p.ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}

	if entry.Level != "INFO" {
		t.Errorf("Level = %q, want INFO", entry.Level)
	}
	if entry.Fields["method"] != "GET" {
		t.Errorf("method = %q, want GET", entry.Fields["method"])
	}
	if entry.Fields["status"] != "200" {
		t.Errorf("status = %q, want 200", entry.Fields["status"])
	}
}

func TestNginxParser_ParseLine_500(t *testing.T) {
	line := `192.168.1.100 - - [15/Jan/2024:10:30:00 +0000] "POST /api/orders HTTP/1.1" 500 0`
	p := newNginxParser()

	entry, err := p.ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}

	if entry.Level != "ERROR" {
		t.Errorf("Level = %q, want ERROR for 500 status", entry.Level)
	}
}

func TestSyslogParser_ParseLine(t *testing.T) {
	line := `<134>Jan 15 10:30:00 webserver nginx[1234]: Server started`
	p := newSyslogParser()

	entry, err := p.ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine() error = %v", err)
	}

	if entry.Message != "Server started" {
		t.Errorf("Message = %q, want 'Server started'", entry.Message)
	}
	if entry.Fields["host"] != "webserver" {
		t.Errorf("host = %q, want webserver", entry.Fields["host"])
	}
}

func TestParser_Parse(t *testing.T) {
	input := `{"level":"INFO","message":"first"}
{"level":"ERROR","message":"second"}
{"level":"WARN","message":"third"}`

	p := &Parser{
		format:    api.FormatJSON,
		batchSize: 10,
	}

	entries, errs := p.Parse(strings.NewReader(input))

	var results []api.LogEntry
	for entry := range entries {
		results = append(results, entry)
	}

	for err := range errs {
		t.Errorf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("got %d entries, want 3", len(results))
	}

	if results[1].Level != "ERROR" {
		t.Errorf("second entry level = %q, want ERROR", results[1].Level)
	}
}

func TestDetectFormat_JSON(t *testing.T) {
	input := `{"level":"INFO","message":"test"}
{"level":"ERROR","message":"fail"}`

	format := detectFromSample(strings.NewReader(input))
	if format != api.FormatJSON {
		t.Errorf("detected format = %q, want json", format)
	}
}

func TestDetectFormat_Nginx(t *testing.T) {
	input := `192.168.1.100 - - [15/Jan/2024:10:30:00 +0000] "GET / HTTP/1.1" 200 1234`

	format := detectFromSample(strings.NewReader(input))
	if format != api.FormatNginx {
		t.Errorf("detected format = %q, want nginx", format)
	}
}

func TestDetectFormat_Syslog(t *testing.T) {
	input := `<134>Jan 15 10:30:00 webserver nginx[1234]: Server started`

	format := detectFromSample(strings.NewReader(input))
	if format != api.FormatSyslog {
		t.Errorf("detected format = %q, want syslog", format)
	}
}

func TestDetectFormatAndReader(t *testing.T) {
	original := `{"level":"INFO","message":"test"}`
	r := strings.NewReader(original)

	format, newR := DetectFormatAndReader(r)

	if format != api.FormatJSON {
		t.Errorf("detected format = %q, want json", format)
	}

	// Verify the reader still has the original content
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(newR); err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	if buf.String() != original {
		t.Errorf("reader content = %q, want %q", buf.String(), original)
	}
}

func TestPlainParser_ParseLine(t *testing.T) {
	p := newPlainParser()

	tests := []struct {
		line    string
		wantLvl string
	}{
		{"ERROR: something broke", "ERROR"},
		{"WARN: be careful", "WARN"},
		{"INFO: all good", "INFO"},
		{"DEBUG: details", "DEBUG"},
		{"just a plain message", "INFO"},
	}

	for _, tt := range tests {
		entry, err := p.ParseLine(tt.line)
		if err != nil {
			t.Errorf("ParseLine(%q) error = %v", tt.line, err)
			continue
		}
		if entry.Level != tt.wantLvl {
			t.Errorf("ParseLine(%q) level = %q, want %q", tt.line, entry.Level, tt.wantLvl)
		}
	}
}

func BenchmarkJSONParser(b *testing.B) {
	line := `{"timestamp":"2024-01-15T10:30:00Z","level":"INFO","message":"Application started on port 8080"}`
	p := newJSONParser(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.ParseLine(line)
	}
}

func BenchmarkParser_Parse(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		buf.WriteString(`{"level":"INFO","message":"test message"}` + "\n")
	}
	input := buf.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := &Parser{format: api.FormatJSON, batchSize: 100}
		entries, _ := p.Parse(strings.NewReader(input))
		for range entries {
			// consume
		}
	}
}

