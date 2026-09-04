package filter

import (
	"regexp"
	"strings"
	"time"

	"github.com/niepres/logparse/pkg/api"
)

// Chain combines multiple filters into a single filter.
type Chain struct {
	filters []api.Filter
}

// NewChain creates a filter chain that matches when ALL filters match.
func NewChain(filters ...api.Filter) *Chain {
	return &Chain{filters: filters}
}

// Match returns true if all filters in the chain match.
func (c *Chain) Match(entry api.LogEntry) bool {
	for _, f := range c.filters {
		if !f.Match(entry) {
			return false
		}
	}
	return true
}

// Add appends a filter to the chain.
func (c *Chain) Add(f api.Filter) {
	c.filters = append(c.filters, f)
}

// LevelFilter filters by minimum log level.
type LevelFilter struct {
	minLevel api.LogLevel
}

// NewLevelFilter creates a filter for entries at or above the given level.
func NewLevelFilter(minLevel api.LogLevel) *LevelFilter {
	return &LevelFilter{minLevel: minLevel}
}

// Match returns true if entry level >= minLevel.
func (f *LevelFilter) Match(entry api.LogEntry) bool {
	entryLevel := api.ParseLogLevel(entry.Level)
	return entryLevel >= f.minLevel
}

// TimeFilter filters entries within a time range.
type TimeFilter struct {
	since time.Time
	until time.Time
}

// NewTimeFilterSince creates a filter for entries after the given time.
func NewTimeFilterSince(since time.Time) *TimeFilter {
	return &TimeFilter{since: since}
}

// NewTimeFilterRange creates a filter for entries within [since, until].
func NewTimeFilterRange(since, until time.Time) *TimeFilter {
	return &TimeFilter{since: since, until: until}
}

// Match returns true if entry timestamp is within the filter's range.
func (f *TimeFilter) Match(entry api.LogEntry) bool {
	if entry.Timestamp.IsZero() {
		return true // Can't filter if no timestamp
	}
	if !f.since.IsZero() && entry.Timestamp.Before(f.since) {
		return false
	}
	if !f.until.IsZero() && entry.Timestamp.After(f.until) {
		return false
	}
	return true
}

// GrepFilter matches entries containing a pattern.
type GrepFilter struct {
	pattern *regexp.Regexp
	negate  bool
}

// NewGrepFilter creates a filter that matches entries containing the regex pattern.
func NewGrepFilter(pattern string) (*GrepFilter, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &GrepFilter{pattern: re}, nil
}

// NewGrepFilterInverse creates a filter that matches entries NOT containing the pattern.
func NewGrepFilterInverse(pattern string) (*GrepFilter, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &GrepFilter{pattern: re, negate: true}, nil
}

// Match returns true if the entry message matches (or doesn't, if negate).
func (f *GrepFilter) Match(entry api.LogEntry) bool {
	matches := f.pattern.MatchString(entry.Message)
	if f.negate {
		return !matches
	}
	return matches
}

// FieldFilter matches entries with a specific field value.
type FieldFilter struct {
	key   string
	value string
}

// NewFieldFilter creates a filter for entries with field[key] == value.
func NewFieldFilter(key, value string) *FieldFilter {
	return &FieldFilter{key: key, value: value}
}

// Match returns true if the entry has the field with the expected value.
func (f *FieldFilter) Match(entry api.LogEntry) bool {
	if entry.Fields == nil {
		return false
	}
	v, ok := entry.Fields[f.key]
	if !ok {
		return false
	}
	return strings.EqualFold(v.(string), f.value)
}

// NotEmptyFilter matches entries with non-empty messages.
type NotEmptyFilter struct{}

// Match returns true if the entry has a non-empty message.
func (f *NotEmptyFilter) Match(entry api.LogEntry) bool {
	return strings.TrimSpace(entry.Message) != ""
}
