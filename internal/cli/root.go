package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/niepres/logparse/internal/filter"
	"github.com/niepres/logparse/internal/output"
	"github.com/niepres/logparse/internal/parser"
	"github.com/niepres/logparse/pkg/api"
	"github.com/spf13/cobra"
)

// Version info - set by main at startup
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion sets the version information for display.
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

// NewRootCommand creates the root cobra command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logparse [files...]",
		Short: "Parse, filter, and format log files",
		Long: `logparse is a CLI tool for parsing, filtering, and formatting log files.
It auto-detects common log formats (JSON, Docker, Nginx, Syslog) and
supports streaming for large files.

Examples:
  logparse app.log
  cat app.log | logparse --level ERROR
  kubectl logs pod | logparse --format docker --since 1h
  logparse app.log --output json | jq '.[] | select(.level=="ERROR")'`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	cmd.Flags().StringP("format", "f", "auto", "log format: auto, json, docker, nginx, syslog, plain")
	cmd.Flags().StringP("output", "o", "table", "output format: table, json, csv")
	cmd.Flags().StringP("level", "l", "", "minimum log level: DEBUG, INFO, WARN, ERROR, FATAL")
	cmd.Flags().String("since", "", "show entries after this time (e.g., 1h, 30m, or 2024-01-01)")
	cmd.Flags().StringP("grep", "g", "", "filter entries matching regex pattern")
	cmd.Flags().StringArrayP("field", "F", nil, "filter by field=value (repeatable)")
	cmd.Flags().Bool("strict", false, "fail on malformed log lines")
	cmd.Flags().Bool("summary", false, "show summary statistics instead of entries")

	// Add subcommands
	cmd.AddCommand(newCompletionCommand())
	cmd.AddCommand(newVersionCommand())

	return cmd
}

// newVersionCommand creates a command for displaying version information.
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("logparse %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built: %s\n", date)
		},
	}
}

// newCompletionCommand creates a command for generating shell completions.
func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `To load completions:

Bash:
  $ source <(logparse completion bash)
  # To load completions for each session, execute once:
  $ logparse completion bash > /etc/bash_completion.d/logparse

Zsh:
  $ source <(logparse completion zsh)
  # To load completions for each session, execute once:
  $ logparse completion zsh > "${fpath[1]}/_logparse"

Fish:
  $ logparse completion fish | source
  # To load completions for each session, execute once:
  $ logparse completion fish > ~/.config/fish/completions/logparse.fish

PowerShell:
  $ logparse completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, execute once:
  $ logparse completion powershell > logparse.ps1
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}

func run(cmd *cobra.Command, files []string, stdout, stderr io.Writer) error {
	// Parse format
	formatStr, _ := cmd.Flags().GetString("format")
	format := api.FormatType(formatStr)

	// Parse output format
	outputFormatStr, _ := cmd.Flags().GetString("output")
	outputFormat := api.OutputType(outputFormatStr)
	outputter, err := output.NewOutputter(outputFormat)
	if err != nil {
		return err
	}

	// Build filter chain
	filterChain := filter.NewChain(&filter.NotEmptyFilter{})

	levelStr, _ := cmd.Flags().GetString("level")
	if levelStr != "" {
		level := api.ParseLogLevel(levelStr)
		if level == api.LevelUnknown {
			return fmt.Errorf("unknown log level: %s", levelStr)
		}
		filterChain.Add(filter.NewLevelFilter(level))
	}

	sinceStr, _ := cmd.Flags().GetString("since")
	if sinceStr != "" {
		since, err := parseSince(sinceStr)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
		filterChain.Add(filter.NewTimeFilterSince(since))
	}

	grepStr, _ := cmd.Flags().GetString("grep")
	if grepStr != "" {
		grepFilter, err := filter.NewGrepFilter(grepStr)
		if err != nil {
			return fmt.Errorf("invalid --grep pattern: %w", err)
		}
		filterChain.Add(grepFilter)
	}

	fieldFlags, _ := cmd.Flags().GetStringArray("field")
	for _, f := range fieldFlags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --field format: %s (expected key=value)", f)
		}
		filterChain.Add(filter.NewFieldFilter(parts[0], parts[1]))
	}

	// Parse strict flag
	strict, _ := cmd.Flags().GetBool("strict")
	summary, _ := cmd.Flags().GetBool("summary")

	// Create parser
	p, err := parser.New(format, api.WithStrict(strict))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		// Read from stdin
		stats, err := parseReader(p, os.Stdin, filterChain, outputter, stdout)
		if err != nil {
			return err
		}
		if summary {
			printSummary(stderr, stats)
		}
		return nil
	}

	// Read from files
	var totalStats *api.ParseStats
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %s: %w", path, err)
		}

		// Auto-detect format if needed
		currentParser := p
		var reader io.Reader = f
		if format == api.FormatAuto {
			detectedFormat, r := parser.DetectFormatAndReader(f)
			currentParser, _ = parser.New(detectedFormat, api.WithStrict(strict))
			reader = r
		}

		stats, err := parseReader(currentParser, reader, filterChain, outputter, stdout)
		f.Close()
		if err != nil {
			return err
		}
		if totalStats == nil {
			totalStats = stats
		} else {
			mergeStats(totalStats, stats)
		}
	}

	if summary {
		printSummary(stderr, totalStats)
	}

	return nil
}

func parseReader(p *parser.Parser, r io.Reader, f api.Filter, o api.Outputter, w io.Writer) (*api.ParseStats, error) {
	stats, err := p.ParseAndFilter(r, f, o, w)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func parseSince(s string) (time.Time, error) {
	// Try duration first
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d), nil
	}

	// Try date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse as duration or date: %s", s)
}

func printSummary(w io.Writer, stats *api.ParseStats) {
	if stats == nil {
		return
	}
	fmt.Fprintf(w, "\n--- Summary ---\n")
	fmt.Fprintf(w, "Total lines:   %d\n", stats.TotalLines)
	fmt.Fprintf(w, "Parsed lines:  %d\n", stats.ParsedLines)
	fmt.Fprintf(w, "Skipped lines: %d\n", stats.SkippedLines)
	if len(stats.ByLevel) > 0 {
		fmt.Fprintf(w, "By level:\n")
		for level, count := range stats.ByLevel {
			fmt.Fprintf(w, "  %s: %d\n", level, count)
		}
	}
}

func mergeStats(dst, src *api.ParseStats) {
	dst.TotalLines += src.TotalLines
	dst.ParsedLines += src.ParsedLines
	dst.SkippedLines += src.SkippedLines
	for level, count := range src.ByLevel {
		dst.ByLevel[level] += count
	}
}
