package journal

import (
	"crosstrace/internal/configs"
	"crosstrace/internal/crypto"
	"crosstrace/internal/encoder"
	"crosstrace/internal/journal/database"
	mptree "crosstrace/internal/merkle"
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

type CommitResult struct {
	BatchID string
	Root    [32]byte
	Count   int
	Entries []PostEntry
}
type JournalConfig struct {
	MaxMsgSize  int
	SafeMode    bool
	Nameencoder string
	NameHasher  string
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
	db, err := NewLocalStorage(name)
	if err != nil {
		return nil
	}
	return &JournalCache{Post: make([]JournalEntry, 10), store: db}
}
func NewLocalStorage(name string) (database.StorageDB, error) {
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
	tree := mptree.NewMerkleTree()
	var elem [][]byte
	if len(j.Post) > 2 {
		for i, entry := range j.Post {
			if i != len(j.Post)-1 {
				enc, err := entry.Encode()
				if err != nil {
					return err
				}
				item := hasher.Sum([]byte(entry.GetID()))
				err = j.store.BatchPut(item, enc, false)
				elem = append(elem, item)
				if err != nil {
					return err
				}

			}
		}
		if tree.Insert(elem) {
			if tree.Commit() {
				return j.store.BatchPut(nil, nil, true)
			}

		}
		return nil // error

	} else {
		for _, entry := range j.Post {
			enc, err := entry.Encode()
			item := hasher.Sum([]byte(entry.GetID()))
			if err != nil {
				return err
			}
			elem = append(elem, item)
			j.store.Put(item, enc)
		}
		if tree.Insert(elem) {
			if tree.Commit() {
				return nil
			}
		}
		return nil // error
	}
}
