package main

import (
	"crosstrace/internal/journal"
	"testing"
)

func TestSettings(t *testing.T) {
	journalcfg = journal.JournalConfig{NameHasher: "md5"}
	journal.SetJournalHasher(&journalcfg)
	t.Log(journal.Name())
}
