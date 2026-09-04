package parser

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/niepres/logparse/pkg/api"
)

// jsonLog represents a generic JSON log entry.
type jsonLog struct {
	Timestamp string `json:"timestamp"`
	Time      string `json:"time"`
	Ts        string `json:"ts"`
	Level     string `json:"level"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Msg       string `json:"msg"`
	Log       string `json:"log"`
}

// jsonParser handles JSON-formatted log lines.
type jsonParser struct {
	strict bool
}

func newJSONParser(strict bool) *jsonParser {
	return &jsonParser{strict: strict}
}

func (p *jsonParser) ParseLine(line string) (api.LogEntry, error) {
	// Handle potential BOM or whitespace
	line = strings.TrimSpace(line)
	if line == "" {
		return api.LogEntry{}, &parseError{msg: "empty line"}
	}

	var raw jsonLog
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return api.LogEntry{}, err
	}

	level := normalizeLevel(raw.Level + raw.Severity)
	message := raw.Message + raw.Msg + raw.Log

	entry := api.LogEntry{
		Level:   level,
		Message: message,
		Raw:     line,
	}

	// Extract timestamp from various common fields
	for _, ts := range []string{raw.Timestamp, raw.Time, raw.Ts} {
		if ts != "" {
			if t, err := parseTimestamp(ts); err == nil {
				entry.Timestamp = t
				break
			}
		}
	}

	return entry, nil
}

// dockerParser handles Docker JSON log files.
type dockerParser struct {
	jsonParser
}

func newDockerParser(strict bool) *dockerParser {
	return &dockerParser{jsonParser: *newJSONParser(strict)}
}

func (p *dockerParser) ParseLine(line string) (api.LogEntry, error) {
	var raw struct {
		Log   string `json:"log"`
		Time  string `json:"time"`
		Level string `json:"level"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return api.LogEntry{}, err
	}

	entry := api.LogEntry{
		Level:   normalizeLevel(raw.Level),
		Message: strings.TrimSpace(raw.Log),
		Raw:     line,
	}

	if raw.Time != "" {
		if t, err := time.Parse(time.RFC3339Nano, raw.Time); err == nil {
			entry.Timestamp = t
		}
	}

	return entry, nil
}

// nginxRegex matches the nginx combined log format.
var nginxRegex = regexp.MustCompile(`^(\S+) \S+ \S+ \[(.+?)\] "(\S+) (\S+) \S+" (\d+) (\d+)`)

type nginxParser struct{}

func newNginxParser() *nginxParser {
	return &nginxParser{}
}

func (p *nginxParser) ParseLine(line string) (api.LogEntry, error) {
	matches := nginxRegex.FindStringSubmatch(line)
	if matches == nil {
		return api.LogEntry{}, &parseError{msg: "line does not match nginx format"}
	}

	timestamp, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
	if err != nil {
		timestamp = time.Time{}
	}

	return api.LogEntry{
		Timestamp: timestamp,
		Level:     inferHTTPLvl(matches[5]),
		Message:   matches[3] + " " + matches[4] + " → " + matches[5],
		Raw:       line,
		Fields: map[string]any{
			"host":       matches[1],
			"method":     matches[3],
			"path":       matches[4],
			"status":     matches[5],
			"bytes_sent": matches[6],
		},
	}, nil
}

// syslogRegex matches standard syslog format.
var syslogRegex = regexp.MustCompile(`^<(\d+)>(\w+ \d+ \d+:\d+:\d+) (\S+) (\S+): (.+)$`)

type syslogParser struct{}

func newSyslogParser() *syslogParser {
	return &syslogParser{}
}

func (p *syslogParser) ParseLine(line string) (api.LogEntry, error) {
	matches := syslogRegex.FindStringSubmatch(line)
	if matches == nil {
		return api.LogEntry{}, &parseError{msg: "line does not match syslog format"}
	}

	timestamp, _ := time.Parse("Jan 02 15:04:05", matches[2])

	return api.LogEntry{
		Timestamp: timestamp,
		Level:     "INFO",
		Message:   matches[5],
		Raw:       line,
		Fields: map[string]any{
			"priority": matches[1],
			"host":     matches[3],
			"process":  matches[4],
		},
	}, nil
}

// plainParser handles plain text log lines (fallback).
type plainParser struct{}

func newPlainParser() *plainParser {
	return &plainParser{}
}

func (p *plainParser) ParseLine(line string) (api.LogEntry, error) {
	return api.LogEntry{
		Level:   inferLevelFromText(line),
		Message: line,
		Raw:     line,
	}, nil
}

// Helper functions

func normalizeLevel(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG", "DBG":
		return "DEBUG"
	case "INFO", "INFORMATION", "INF":
		return "INFO"
	case "WARN", "WARNING":
		return "WARN"
	case "ERROR", "ERR":
		return "ERROR"
	case "FATAL", "CRITICAL", "CRIT":
		return "FATAL"
	default:
		return "INFO"
	}
}

func inferLevelFromText(line string) string {
	upper := strings.ToUpper(line)
	if strings.Contains(upper, "FATAL") || strings.Contains(upper, "CRITICAL") {
		return "FATAL"
	}
	if strings.Contains(upper, "ERROR") {
		return "ERROR"
	}
	if strings.Contains(upper, "WARN") {
		return "WARN"
	}
	if strings.Contains(upper, "DEBUG") {
		return "DEBUG"
	}
	if strings.Contains(upper, "INFO") {
		return "INFO"
	}
	return "INFO"
}

func inferHTTPLvl(status string) string {
	if len(status) == 0 {
		return "INFO"
	}
	switch status[0] {
	case '4':
		return "WARN"
	case '5':
		return "ERROR"
	default:
		return "INFO"
	}
}

func parseTimestamp(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999Z07:00",
		"2006-01-02T15:04:05.999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &parseError{msg: "unable to parse timestamp: " + s}
}

type parseError struct {
	msg string
}

func (e *parseError) Error() string {
	return e.msg
}
