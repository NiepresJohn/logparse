package output

import (
	"io"
	"strings"

	"github.com/niepres/logparse/pkg/api"
	"github.com/olekukonko/tablewriter"
)

// TableOutput formats log entries as a human-readable table.
type TableOutput struct {
	table  *tablewriter.Table
	count  int
	writer io.Writer
}

// NewTableOutput creates a new table outputter.
func NewTableOutput() *TableOutput {
	return &TableOutput{}
}

// Write outputs a single entry. On first call, initializes the table.
func (o *TableOutput) Write(w io.Writer, entry api.LogEntry) error {
	if o.table == nil {
		o.table = tablewriter.NewWriter(w)
		o.table.SetHeader([]string{"TIME", "LEVEL", "MESSAGE"})
		o.table.SetAutoWrapText(false)
		o.table.SetAutoFormatHeaders(true)
		o.table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		o.table.SetAlignment(tablewriter.ALIGN_LEFT)
		o.table.SetCenterSeparator("")
		o.table.SetColumnSeparator("")
		o.table.SetRowSeparator("")
		o.table.SetHeaderLine(false)
		o.table.SetBorder(false)
		o.table.SetTablePadding("\t")
		o.table.SetNoWhiteSpace(true)
		o.writer = w
	}

	timestamp := ""
	if !entry.Timestamp.IsZero() {
		timestamp = entry.Timestamp.Format("15:04:05")
	}

	message := entry.Message
	if len(message) > 100 {
		message = message[:97] + "..."
	}

	o.table.Append([]string{timestamp, colorizeLevel(entry.Level), message})
	o.count++

	// Flush periodically for streaming output
	if o.count%100 == 0 {
		o.table.Render()
		o.table = tablewriter.NewWriter(w)
		o.table.SetHeader([]string{"TIME", "LEVEL", "MESSAGE"})
		o.table.SetAutoWrapText(false)
		o.table.SetAutoFormatHeaders(true)
		o.table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		o.table.SetAlignment(tablewriter.ALIGN_LEFT)
		o.table.SetCenterSeparator("")
		o.table.SetColumnSeparator("")
		o.table.SetRowSeparator("")
		o.table.SetHeaderLine(false)
		o.table.SetBorder(false)
		o.table.SetTablePadding("\t")
		o.table.SetNoWhiteSpace(true)
	}

	return nil
}

// Flush renders any remaining rows.
func (o *TableOutput) Flush(w io.Writer) error {
	if o.table != nil && o.count%100 != 0 {
		o.table.Render()
	}
	return nil
}

// colorizeLevel adds ANSI color codes to log levels.
func colorizeLevel(level string) string {
	if !isTerminal() {
		return level
	}
	switch strings.ToUpper(level) {
	case "DEBUG":
		return "\033[37m" + level + "\033[0m" // gray
	case "INFO":
		return "\033[36m" + level + "\033[0m" // cyan
	case "WARN":
		return "\033[33m" + level + "\033[0m" // yellow
	case "ERROR":
		return "\033[31m" + level + "\033[0m" // red
	case "FATAL":
		return "\033[35m" + level + "\033[0m" // magenta
	default:
		return level
	}
}

// isTerminal checks if stdout is a terminal (simplified).
func isTerminal() bool {
	// In a real implementation, check if os.Stdout is a tty
	// For now, return false to keep output pipe-friendly
	return false
}

