package journal

import (
	"crosstrace/context"
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
	Batch() (*CommitResult, error) // used for manual batch creation
	Close() error                        // shutdows
}
type CommitResult struct {
	ctx          *context.Context
	BatchID      string
	Root         [32]byte
	Count        uint32
	WindowsStart time.Time // first j.Port // assuming that it is ordered
	WindowsEnd   time.Time // last j.Post
	Version      string
}
type Commitment struct {
	ctx          *context.Context // anonymous
	Roothash     [32]byte
	Count        uint32
	WindowsStart time.Time
	WindowsEnd   time.Time
	Commitment   []byte
}

// Default format when received / Unsafe
type PreEntry struct {
	ctx        *context.Context
	sender_id  string
	raw_msg    string
	timestamp  time.Time
	source     string
	session_id string
}

// PostEntry is the sanitized event
type PostEntry struct {
	ctx       *context.Context
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

func NewLocalStorage(ctx *context.Context) (database.StorageDB, error) {
	switch ctx.Journal.DBName {
	case "badgerdb":
		return database.NewBadgerdb(ctx.Journal)
	case "pebbledb":
		return database.NewPebbledb(ctx.Journal)
	default:
		return database.NewBadgerdb(ctx.Journal)
	}
}

// treeroot is left untouched unless buildtree is called
// this avoid having to recompute tree if something fails along the way

type JournalCache struct {
	ctx       *context.Context
	store     database.StorageDB
	Post      []JournalEntry
	treeroot  []byte
	batchid   []byte
	commitRes *CommitResult
}
