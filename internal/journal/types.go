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

// Complex due to the needs
// building the tree separatly to allow fast debugging
type JournalStore interface {
	Append(entry JournalEntry) (journalID string, err error)
	Commit() error
	Entries() []JournalEntry
	BuildTree() error
	Get(id string) ([]byte, error)
	BatchInsert() (*CommitResult, error)
	Close() error // shutdows
}
type CommitResult struct {
	BatchID      string
	Root         [32]byte
	Count        uint32
	WindowsStart time.Time // first j.Port // assuming that it is ordered
	WindowsEnd   time.Time // last j.Post
}

// Default format when received / Unsafe
type PreEntry struct {
	sender_id  string
	raw_msg    string
	timestamp  time.Time
	source     string
	session_id string
}

// PostEntry is the sanitized event
type PostEntry struct {
	SenderID  string    `json:"sender_id"`
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
	CleanMsg  string    `json:"clean_msg"`
	Source    string    `json:"source"`
	Checksum  string    `json:"checksum"`
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

// treeroot is left untouched unless buildtree is called
// this avoid having to recompute tree if something fails along the way

type JournalCache struct {
	store    database.StorageDB
	Post     []JournalEntry
	treeroot []byte
	batchid  string
}
