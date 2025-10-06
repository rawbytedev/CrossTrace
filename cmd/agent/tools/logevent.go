package crosstracetools

import (
	"context"
	"crosstrace/internal/journal"
	"fmt"
)

type LogEventTool struct {
	Journal journal.JournalStore
}

func (t *LogEventTool) Name() string        { return "log_event" }
func (t *LogEventTool) Description() string { return "Append a new event to the journal" }
// add input parsing for direct use
func (t *LogEventTool) Call(ctx context.Context, input string) (string, error) {
	pre := ParsePreEntry(input) // also sanitze
	post, err := t.Journal.Append(pre)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Event logged with checksum %s", post), nil
}

// input must be converted to journal.PreEntry before sanitization
func ParsePreEntry(input string) journal.JournalEntry {
	return &journal.PostEntry{}
}

func NewLogEventTool(cache journal.JournalStore) *LogEventTool {
	return &LogEventTool{Journal: cache}
}