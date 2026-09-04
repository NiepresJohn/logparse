package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd.Use != "logparse [files...]" {
		t.Errorf("Use = %q, want 'logparse [files...]'", cmd.Use)
	}

	// Check required flags exist
	flags := []string{"format", "output", "level", "since", "grep", "field", "strict", "summary"}
	for _, f := range flags {
		if cmd.Flags().Lookup(f) == nil {
			t.Errorf("missing flag: %s", f)
		}
	}
}

func TestRun_EmptyStdin(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() with empty stdin should not error, got: %v", err)
	}
}

func TestRun_WithJsonFile(t *testing.T) {
	cmd := NewRootCommand()
	// Use -- to separate flags from positional args
	cmd.SetArgs([]string{"--output", "json", "--", "../testdata/sample.json.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("output should contain log entries, got: %s", output)
	}
}

func TestRun_WithLevelFilter(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--level", "ERROR", "--output", "json", "--", "../testdata/sample.json.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "ERROR") && !strings.Contains(line, "FATAL") {
			t.Errorf("filtered output should only contain ERROR/FATAL, got: %s", line)
		}
	}
}

func TestRun_WithInvalidLevel(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--level", "INVALID", "--", "../testdata/sample.json.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error with invalid level")
	}
}

func TestRun_WithSummary(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--summary", "--", "../testdata/sample.json.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stderr.String(), "Summary") {
		t.Errorf("summary output should be on stderr")
	}
}

func TestRun_WithCSVOutput(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--output", "csv", "--", "../testdata/sample.json.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "timestamp,level,message,source") {
		t.Errorf("CSV should have header row, got: %s", output)
	}
}

func TestRun_WithGrepFilter(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--grep", "ERROR|FATAL", "--output", "json", "--", "../testdata/sample.json.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, `"level":"INFO"`) {
		t.Error("grep filter should exclude INFO entries")
	}
}

func TestRun_WithNginxFormat(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--format", "nginx", "--output", "json", "--", "../testdata/sample.nginx.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("nginx output should contain method, got: %s", output)
	}
}

func TestRun_WithSyslogFormat(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--format", "syslog", "--output", "json", "--", "../testdata/sample.syslog"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "webserver") {
		t.Errorf("syslog output should contain host, got: %s", output)
	}
}

func TestRun_WithInvalidFile(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"nonexistent.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error with nonexistent file")
	}
}

func TestRun_WithInvalidGrep(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--grep", "[invalid", "--", "../testdata/sample.json.log"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Error("Execute() should error with invalid grep pattern")
	}
}
