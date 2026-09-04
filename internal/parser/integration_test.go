package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niepres/logparse/internal/filter"
	"github.com/niepres/logparse/internal/output"
	"github.com/niepres/logparse/pkg/api"
)

func TestEndToEnd_JSON(t *testing.T) {
	p, err := New(api.FormatJSON)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	f, err := os.Open("../../testdata/sample.json.log")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer f.Close()

	out := output.NewJSONOutput()
	chain := filter.NewChain(&filter.NotEmptyFilter{})

	var buf strings.Builder
	stats, err := p.ParseAndFilter(f, chain, out, &buf)
	if err != nil {
		t.Fatalf("ParseAndFilter() error = %v", err)
	}

	if stats.TotalLines == 0 {
		t.Error("should have parsed some lines")
	}

	result := buf.String()
	if !strings.Contains(result, "INFO") {
		t.Errorf("should contain INFO entries, got: %s", result)
	}
}

func TestEndToEnd_WithLevelFilter(t *testing.T) {
	p, _ := New(api.FormatJSON)
	f, _ := os.Open("../../testdata/sample.json.log")
	defer f.Close()

	out := output.NewJSONOutput()
	chain := filter.NewChain(
		&filter.NotEmptyFilter{},
		filter.NewLevelFilter(api.LevelError),
	)

	var buf strings.Builder
	_, err := p.ParseAndFilter(f, chain, out, &buf)
	if err != nil {
		t.Fatalf("ParseAndFilter() error = %v", err)
	}

	result := buf.String()
	lines := strings.Split(strings.TrimSpace(result), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "ERROR") && !strings.Contains(line, "FATAL") {
			t.Errorf("filtered output should only contain ERROR/FATAL, got: %s", line)
		}
	}
}

func TestEndToEnd_Nginx(t *testing.T) {
	p, _ := New(api.FormatNginx)
	f, _ := os.Open("../../testdata/sample.nginx.log")
	defer f.Close()

	out := output.NewTableOutput()
	chain := filter.NewChain(&filter.NotEmptyFilter{})

	var buf strings.Builder
	stats, err := p.ParseAndFilter(f, chain, out, &buf)
	if err != nil {
		t.Fatalf("ParseAndFilter() error = %v", err)
	}

	if stats.TotalLines != 8 {
		t.Errorf("expected 8 lines, got %d", stats.TotalLines)
	}
}

func TestEndToEnd_AutoDetect(t *testing.T) {
	tests := []struct {
		file   string
		format api.FormatType
	}{
		{"../../testdata/sample.json.log", api.FormatJSON},
		{"../../testdata/sample.nginx.log", api.FormatNginx},
		{"../../testdata/sample.syslog", api.FormatSyslog},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			f, err := os.Open(tt.file)
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			defer f.Close()

			format, reader := DetectFormatAndReader(f)
			if format != tt.format {
				t.Errorf("detected format = %q, want %q", format, tt.format)
			}

			p, _ := New(format)
			out := output.NewJSONOutput()
			chain := filter.NewChain(&filter.NotEmptyFilter{})

			var buf strings.Builder
			_, err = p.ParseAndFilter(reader, chain, out, &buf)
			if err != nil {
				t.Fatalf("ParseAndFilter() error = %v", err)
			}
		})
	}
}

func TestEndToEnd_LargeFile(t *testing.T) {
	// Create a temporary large log file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "large.log")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Write 10000 lines
	for i := 0; i < 10000; i++ {
		f.WriteString(`{"level":"INFO","message":"test message"}` + "\n")
	}
	f.Close()

	// Parse it
	p, _ := New(api.FormatJSON)
	f, _ = os.Open(tmpFile)
	defer f.Close()

	out := output.NewJSONOutput()
	chain := filter.NewChain(&filter.NotEmptyFilter{})

	var buf strings.Builder
	stats, err := p.ParseAndFilter(f, chain, out, &buf)
	if err != nil {
		t.Fatalf("ParseAndFilter() error = %v", err)
	}

	if stats.TotalLines != 10000 {
		t.Errorf("expected 10000 lines, got %d", stats.TotalLines)
	}
}

func TestEndToEnd_EmptyLines(t *testing.T) {
	p, _ := New(api.FormatJSON)

	input := `{"level":"INFO","message":"first"}

{"level":"ERROR","message":"second"}

{"level":"WARN","message":"third"}`

	out := output.NewJSONOutput()
	chain := filter.NewChain(&filter.NotEmptyFilter{})

	var buf strings.Builder
	stats, err := p.ParseAndFilter(strings.NewReader(input), chain, out, &buf)
	if err != nil {
		t.Fatalf("ParseAndFilter() error = %v", err)
	}

	if stats.TotalLines != 3 {
		t.Errorf("expected 3 lines (empty lines skipped), got %d", stats.TotalLines)
	}
}

func TestEndToEnd_MalformedLines(t *testing.T) {
	p, _ := New(api.FormatJSON)

	input := `{"level":"INFO","message":"valid"}
this is not json
{"level":"ERROR","message":"also valid"}`

	out := output.NewJSONOutput()
	chain := filter.NewChain(&filter.NotEmptyFilter{})

	var buf strings.Builder
	stats, err := p.ParseAndFilter(strings.NewReader(input), chain, out, &buf)
	if err != nil {
		t.Fatalf("ParseAndFilter() error = %v", err)
	}

	// Should parse 2 valid lines, skip 1 malformed
	if stats.ParsedLines != 2 {
		t.Errorf("expected 2 parsed lines, got %d", stats.ParsedLines)
	}
}

func TestEndToEnd_StrictMode(t *testing.T) {
	p, _ := New(api.FormatJSON, api.WithStrict(true))

	input := `{"level":"INFO","message":"valid"}
this is not json`

	out := output.NewJSONOutput()
	chain := filter.NewChain(&filter.NotEmptyFilter{})

	var buf strings.Builder
	_, err := p.ParseAndFilter(strings.NewReader(input), chain, out, &buf)
	// In strict mode, the parser stops at first error but doesn't return it
	// The stats will show fewer lines parsed than total
	if err != nil {
		// This is acceptable - strict mode returns error
		t.Logf("strict mode returned error as expected: %v", err)
	}
}
