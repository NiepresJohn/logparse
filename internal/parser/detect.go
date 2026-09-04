package parser

import (
	"bufio"
	"bytes"
	"io"

	"github.com/niepres/logparse/pkg/api"
)

const sampleLines = 5

// DetectFormat analyzes the first lines of input to determine the log format.
func DetectFormat(r io.Reader) api.FormatType {
	// We need to peek at the first lines without consuming the reader.
	// Read a sample, then return a new reader that prepends the sample.
	peeked := make([]byte, 64*1024) // 64KB sample
	n, _ := r.Read(peeked)
	peeked = peeked[:n]

	format := detectFromSample(bytes.NewReader(peeked))
	return format
}

// DetectFormatAndReader detects format and returns a reader that includes the consumed bytes.
func DetectFormatAndReader(r io.Reader) (api.FormatType, io.Reader) {
	peeked := make([]byte, 64*1024)
	n, _ := r.Read(peeked)
	peeked = peeked[:n]

	format := detectFromSample(bytes.NewReader(peeked))
	// Return a new reader that prepends the sample back
	return format, io.MultiReader(bytes.NewReader(peeked), r)
}

func detectFromSample(r io.Reader) api.FormatType {
	scanner := bufio.Scanner(r)
	lines := make([]string, 0, sampleLines)

	for scanner.Scan() && len(lines) < sampleLines {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return api.FormatPlain
	}

	// Check each format
	jsonCount := 0
	for _, line := range lines {
		if looksLikeJSON(line) {
			jsonCount++
		}
	}
	if jsonCount == len(lines) {
		// Check if it's Docker format
		if looksLikeDocker(lines[0]) {
			return api.FormatDocker
		}
		return api.FormatJSON
	}

	if looksLikeNginx(lines[0]) {
		return api.FormatNginx
	}

	if looksLikeSyslog(lines[0]) {
		return api.FormatSyslog
	}

	return api.FormatPlain
}

func looksLikeJSON(line string) bool {
	trimmed := trimLeftSpace(line)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	// Quick check: does it have a closing brace and quotes
	return trimmed[len(trimmed)-1] == '}' && bytes.Contains([]byte(trimmed), []byte(`"`))
}

func looksLikeDocker(line string) bool {
	// Docker logs have "log", "time" fields
	return bytes.Contains([]byte(line), []byte(`"log"`)) &&
		bytes.Contains([]byte(line), []byte(`"time"`))
}

func looksLikeNginx(line string) bool {
	// Starts with IP address pattern
	if len(line) < 20 {
		return false
	}
	// Quick heuristic: has bracketed timestamp
	return bytes.Contains([]byte(line), []byte(`[`)) &&
		bytes.Contains([]byte(line), []byte(`"`)) &&
		bytes.Contains([]byte(line), []byte(` HTTP/`))
}

func looksLikeSyslog(line string) bool {
	// Starts with <priority>
	if len(line) < 10 || line[0] != '<' {
		return false
	}
	// Look for closing > followed by month abbreviation
	for i := 1; i < len(line) && i < 10; i++ {
		if line[i] == '>' {
			// Check if next chars look like a month
			rest := line[i+1:]
			return looksLikeMonth(rest)
		}
	}
	return false
}

func looksLikeMonth(s string) bool {
	if len(s) < 3 {
		return false
	}
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	prefix := s[:3]
	for _, m := range months {
		if prefix == m {
			return true
		}
	}
	return false
}

func trimLeftSpace(s string) string {
	for i, c := range s {
		if c != ' ' && c != '\t' {
			return s[i:]
		}
	}
	return ""
}
