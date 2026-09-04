package filter

import (
	"testing"
	"time"

	"github.com/niepres/logparse/pkg/api"
)

func TestLevelFilter(t *testing.T) {
	f := NewLevelFilter(api.LevelWarn)

	tests := []struct {
		level string
		want  bool
	}{
		{"DEBUG", false},
		{"INFO", false},
		{"WARN", true},
		{"ERROR", true},
		{"FATAL", true},
	}

	for _, tt := range tests {
		entry := api.LogEntry{Level: tt.level}
		if got := f.Match(entry); got != tt.want {
			t.Errorf("LevelFilter(%q) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

func TestTimeFilterSince(t *testing.T) {
	since := time.Now().Add(-1 * time.Hour)
	f := NewTimeFilterSince(since)

	tests := []struct {
		name      string
		timestamp time.Time
		want      bool
	}{
		{"recent entry", time.Now(), true},
		{"old entry", time.Now().Add(-2 * time.Hour), false},
		{"no timestamp", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := api.LogEntry{Timestamp: tt.timestamp}
			if got := f.Match(entry); got != tt.want {
				t.Errorf("TimeFilter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrepFilter(t *testing.T) {
	f, err := NewGrepFilter(`error|fail`)
	if err != nil {
		t.Fatalf("NewGrepFilter error: %v", err)
	}

	tests := []struct {
		message string
		want    bool
	}{
		{"something went wrong", false},
		{"an error occurred", true},
		{"operation failed", true},
		{"all good", false},
	}

	for _, tt := range tests {
		entry := api.LogEntry{Message: tt.message}
		if got := f.Match(entry); got != tt.want {
			t.Errorf("GrepFilter(%q) = %v, want %v", tt.message, got, tt.want)
		}
	}
}

func TestGrepFilterInverse(t *testing.T) {
	f, err := NewGrepFilterInverse(`debug`)
	if err != nil {
		t.Fatalf("NewGrepFilterInverse error: %v", err)
	}

	if f.Match(api.LogEntry{Message: "debug info"}) {
		t.Error("inverse filter should not match 'debug info'")
	}
	if !f.Match(api.LogEntry{Message: "important message"}) {
		t.Error("inverse filter should match 'important message'")
	}
}

func TestFieldFilter(t *testing.T) {
	f := NewFieldFilter("method", "POST")

	tests := []struct {
		name   string
		fields map[string]any
		want   bool
	}{
		{"matching field", map[string]any{"method": "POST"}, true},
		{"different value", map[string]any{"method": "GET"}, false},
		{"missing field", map[string]any{"path": "/api"}, false},
		{"no fields", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := api.LogEntry{Fields: tt.fields}
			if got := f.Match(entry); got != tt.want {
				t.Errorf("FieldFilter.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChain(t *testing.T) {
	chain := NewChain(
		NewLevelFilter(api.LevelWarn),
		&NotEmptyFilter{},
	)

	tests := []struct {
		name  string
		entry api.LogEntry
		want  bool
	}{
		{"warn with message", api.LogEntry{Level: "WARN", Message: "alert"}, true},
		{"error with message", api.LogEntry{Level: "ERROR", Message: "fail"}, true},
		{"info with message", api.LogEntry{Level: "INFO", Message: "ok"}, false},
		{"warn empty message", api.LogEntry{Level: "WARN", Message: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chain.Match(tt.entry); got != tt.want {
				t.Errorf("Chain.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChain_Add(t *testing.T) {
	chain := NewChain()
	chain.Add(NewLevelFilter(api.LevelError))

	entry := api.LogEntry{Level: "WARN"}
	if chain.Match(entry) {
		t.Error("chain should not match WARN level")
	}

	chain.Add(&NotEmptyFilter{})
	entry = api.LogEntry{Level: "ERROR", Message: "fail"}
	if !chain.Match(entry) {
		t.Error("chain should match ERROR with message")
	}
}

func TestNotEmptyFilter(t *testing.T) {
	f := &NotEmptyFilter{}

	tests := []struct {
		message string
		want    bool
	}{
		{"has content", true},
		{"", false},
		{"   ", false},
		{"\t\n", false},
	}

	for _, tt := range tests {
		entry := api.LogEntry{Message: tt.message}
		if got := f.Match(entry); got != tt.want {
			t.Errorf("NotEmptyFilter(%q) = %v, want %v", tt.message, got, tt.want)
		}
	}
}

