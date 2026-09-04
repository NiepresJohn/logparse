package api

import "time"

// LogEntry represents a parsed log line with structured fields.
type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
	Raw       string         `json:"-"`
	Source    string         `json:"source,omitempty"`
}

// ParseStats holds summary statistics about a parse operation.
type ParseStats struct {
	TotalLines   int            `json:"total_lines"`
	ParsedLines  int            `json:"parsed_lines"`
	SkippedLines int            `json:"skipped_lines"`
	ByLevel      map[string]int `json:"by_level,omitempty"`
	TimeRange    *TimeRange     `json:"time_range,omitempty"`
}

// TimeRange represents the span of timestamps in parsed entries.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// FormatType identifies a log format.
type FormatType string

const (
	FormatAuto   FormatType = "auto"
	FormatJSON   FormatType = "json"
	FormatDocker FormatType = "docker"
	FormatNginx  FormatType = "nginx"
	FormatSyslog FormatType = "syslog"
	FormatPlain  FormatType = "plain"
)

// OutputType identifies an output format.
type OutputType string

const (
	OutputTable OutputType = "table"
	OutputJSON  OutputType = "json"
	OutputCSV   OutputType = "csv"
)

// LogLevel represents severity levels for filtering.
type LogLevel int

const (
	LevelUnknown LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of a LogLevel.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel converts a string to a LogLevel.
func ParseLogLevel(s string) LogLevel {
	switch s {
	case "DEBUG", "debug":
		return LevelDebug
	case "INFO", "info":
		return LevelInfo
	case "WARN", "warn", "WARNING", "warning":
		return LevelWarn
	case "ERROR", "error":
		return LevelError
	case "FATAL", "fatal":
		return LevelFatal
	default:
		return LevelUnknown
	}
}
