package journal

import (
	"crosstrace/internal/configs"
	"crosstrace/internal/crypto"
	"crosstrace/internal/encoder"
	"crosstrace/internal/journal/database"
	"time"
)

var encoders = encoder.NewEncoder("yaml")
var cfg = configs.Config{}
var hasher = crypto.NewHasher("sha256")

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

type JournalConfig struct {
	MaxMsgSize  int
	SafeMode    bool
	nameencoder string
	nameHasher  string
}
type BatchJournal struct {
	entry []PostEntry
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
func NewJournalCache(name string, cfg configs.Config) JournalStore {
	db, err := NewLocalStorage(name, cfg)
	if err != nil {
		return nil
	}
	return &JournalCache{Post: make([]JournalEntry, 10), store: db, cfg: cfg}
}
func NewLocalStorage(name string, cfg configs.Config) (database.StorageDB, error) {
	switch name {
	case "badgerdb":
		return database.NewBadgerdb(cfg)
	case "pebbledb":
		return database.NewPebbledb(cfg)
	default:
		return database.NewBadgerdb(cfg)
	}
}

type JournalCache struct {
	cfg   configs.Config
	store database.StorageDB
	Post  []JournalEntry
}

// those are called by main
func (j *JournalCache) Entries() []JournalEntry {
	return j.Post
}
func (j *JournalCache) Append(entry JournalEntry) (string, error) {
	j.Post = append(j.Post, entry)
	return entry.GetID(), nil
}
func (j *JournalCache) Commit() error {
	for _, entry := range j.Post {
		if len(j.Post) < 2 {
			enc, err := entry.Encode()
			if err != nil {
				return err
			}
			j.store.Put([]byte(entry.GetID()), enc)
		}
		enc, err := entry.Encode()
		if err != nil {
			return err
		}
		err = j.store.BatchPut([]byte(entry.GetID()), enc, false)
		if err != nil {
			return err
		}
	}
	return j.store.BatchPut(nil, nil, true)
}
