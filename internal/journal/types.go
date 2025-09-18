package journal

import (
	"crosstrace/internal/journal/database"
	"time"
)
type JournalEntry interface {
	GetID() string
	GetTimestamp() time.Time
	Encode() ([]byte, error)
	Decode([]byte) error
}
type JournalStore interface {
	Append(entry JournalEntry) (journalID string, err error)
	Commit() error
	Entries() []JournalEntry
}
type CommitResult struct {
	BatchID string
	Root    [32]byte
	Count   int
	Entries []PostEntry
}
// Default format when received / Unsafe
type PreEntry struct {
	sender_id  string
	raw_msg    string
	timestamp  time.Time
	source     string
	session_id string
	maxsize    int64
}

// PostEntry is the sanitized event
type PostEntry struct {
	SenderID  string            `json:"sender_id"`
	SessionID string            `json:"session_id"`
	Timestamp time.Time         `json:"timestamp"`
	CleanMsg  string            `json:"clean_msg"`
	Source    string            `json:"source"`
	Meta      map[string]string `json:"meta"`
	Checksum  string            `json:"checksum"`
}

type Event struct {
	ts         string
	actor      string
	action     string
	payload    string
	tags       string
	confidence int
	comment    string
}

// called by main
func NewPreEntry(maxsize uint64, raw_msg string, sender_id string, source string, session_id string) *PreEntry {
	return &PreEntry{raw_msg: raw_msg, sender_id: sender_id, source: source, session_id: session_id}
}
func NewJournalCache(name string) JournalStore {
	db, err := NewLocalStorage(name)
	if err != nil {
		return nil
	}
	return &JournalCache{Post: make([]JournalEntry, 10), store: db}
}
func NewLocalStorage(name string) (database.StorageDB, error) {
	switch name {
	case "badgerdb":
		return database.NewBadgerdb(JournalConfig)
	case "pebbledb":
		return database.NewPebbledb(JournalConfig)
	default:
		return database.NewBadgerdb(JournalConfig)
	}
}

type JournalCache struct {
	store database.StorageDB
	Post  []JournalEntry
}

