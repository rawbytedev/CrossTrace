package journal

import (
	"crosstrace/internal/configs"
	"crosstrace/internal/crypto"
	"crosstrace/internal/encoder"
	mptree "crosstrace/internal/merkle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var encoders encoder.Encoder
var hasher crypto.Hasher
var JournalConfig configs.JournalConfig

// the way to enter the journal package
// call this first
// only NewJournalCache is needed to start using package
func NewJournalCache(cfg *configs.JournalConfig) JournalStore {
	db, err := NewLocalStorage(cfg.DBName)
	if err != nil {
		return nil
	}
	return &JournalCache{store: db}
}

// must be called from Main to set globally
func SetAllJournalConfigs(cfg configs.JournalConfig) {
	SetJournalConfigs(cfg)
	SetJournalEncoder()
	SetJournalHasher()
}

// called by main
// testing
// this is received by ai
// Will remove this
func NewPreEntry(maxsize uint64, raw_msg string, sender_id string, source string, session_id string) *PreEntry {
	return &PreEntry{raw_msg: raw_msg, sender_id: sender_id, source: source, session_id: session_id}
}

func (pre *PreEntry) GetID() string {
	return pre.session_id
}
func (pre *PreEntry) GetTimestamp() time.Time {
	return pre.timestamp
}
func (pre *PreEntry) Encode() ([]byte, error) {
	return encoders.Encode(pre)
}
func (pre *PreEntry) Decode(data []byte) error {
	return encoders.Decode(data, pre)
}

func (post *PostEntry) GetID() string {
	return post.Checksum
}
func (post *PostEntry) GetTimestamp() time.Time {
	return post.Timestamp
}

func (post *PostEntry) Encode() ([]byte, error) {
	return encoders.Encode(post)
}
func (post *PostEntry) Decode(data []byte) error {
	return encoders.Decode(data, post)
}
func (res *CommitResult) Encode() ([]byte, error) {
	return encoders.Encode(res)
}
func (res *CommitResult) Decode(data []byte) error {
	return encoders.Decode(data, res)
}

// set based on configuration
// that means if config are changed during run
// simply call set JournalConfigs and other setJournal
// to change package configuration
func SetJournalEncoder() {
	encoders = encoder.NewEncoder(JournalConfig.EncoderName)
}
func SetJournalHasher() {
	hasher = crypto.NewHasher(JournalConfig.HasherName)
}
func SetJournalConfigs(cfgs configs.JournalConfig) {
	JournalConfig = cfgs
}

// Handle Sanitization : add global error vars
// change PreEntry/PostEntry to JournalEntry
func SanitizePreEntry(pre *PreEntry) (JournalEntry, error) {
	// size check
	if len(pre.raw_msg) > JournalConfig.MaxMsgSize {
		return &PostEntry{}, fmt.Errorf("message exceeds max size: %d > %d", len(pre.raw_msg), JournalConfig.MaxMsgSize)
	}
	// UTF-8 validation
	if !utf8.ValidString(pre.raw_msg) {
		return &PostEntry{}, fmt.Errorf("invalid UTF-8 sequence detected")
	}

	// Character whitelist / suspicious content check
	allowed := func(r rune) bool {
		// Allow letters, numbers, common punctuation, whitespace
		return unicode.IsLetter(r) || unicode.IsNumber(r) ||
			unicode.IsPunct(r) || unicode.IsSpace(r)
	}

	total := 0
	suspicious := 0
	var cleanBuilder strings.Builder

	for _, r := range pre.raw_msg {
		total++
		if allowed(r) {
			cleanBuilder.WriteRune(r)
		} else {
			suspicious++
			if JournalConfig.SafeMode {
				// Replace with placeholder in safe mode
				cleanBuilder.WriteRune(' ')
			}
		}
	}
	// If too many suspicious chars, reject
	if float64(suspicious)/float64(total) > 0.15 {
		return &PostEntry{}, fmt.Errorf("message flagged as potentially malicious")
	}

	cleanMsg := strings.TrimSpace(cleanBuilder.String())

	// Metadata sanity checks
	if pre.sender_id == "" || pre.session_id == "" {
		return &PostEntry{}, fmt.Errorf("missing required metadata")
	}

	// Compute checksum (SHA-256 of clean message + sender + timestamp)
	// Change to Hasher interface
	checksumInput := fmt.Sprintf("%s|%s|%s", cleanMsg, pre.sender_id, pre.timestamp.UTC().Format(time.RFC3339Nano))
	checksum := hasher.Sum([]byte(checksumInput))
	checksumHex := hex.EncodeToString(checksum[:])
	_ = checksumHex
	// Return safe PostEntry
	// Create NewPostEntry fot return
	return &PostEntry{
		SenderID:  pre.sender_id,
		SessionID: pre.session_id,
		Source:    pre.source,
		Timestamp: pre.timestamp,
		CleanMsg:  cleanMsg,
		Checksum:  checksumHex,
	}, nil
}

// those are called by main
func (j *JournalCache) Entries() []JournalEntry {
	return j.Post
}
func (j *JournalCache) Append(entry JournalEntry) (string, error) {
	j.Post = append(j.Post, entry)
	return entry.GetID(), nil
}

// only call this when ready to commit
// do not insert after building tree
// if you insert rebuild tree or it won't match
func (j *JournalCache) BuildTree() error {
	tree := mptree.NewMerkleTree()
	var elem [][]byte
	for _, entry := range j.Post {
		// in this case Post Entry checksum
		elem = append(elem, []byte(entry.GetID()))
	}
	res := tree.Insert(elem)
	if !res {
		return fmt.Errorf("unable to insert into tree")
	}
	res = tree.Commit()
	if !res {
		return fmt.Errorf("unable to build tree")
	}
	j.treeroot = tree.Root()
	return nil
}

// this is related to commitresult needed to mint and anchor
// run after calling buildtree and before committing onto database
// needs len(j.post) j.treeroot timewindow
func (j *JournalCache) BatchInsert() (*CommitResult, error) {
	batch := CommitResult{
		BatchID: hex.EncodeToString(j.treeroot)[:4],
		Root:    [32]byte(j.treeroot),
		Count:   len(j.Post),
	}
	enc, err := batch.Encode()
	if err != nil {
		return &CommitResult{}, err
	}
	return &batch, j.store.Put(fmt.Appendf(nil, "batch:%s", j.treeroot), enc)
}

// only store post Entries
// entry are rehashed
// pattern
// chk:%s (checksum) -> PostEntry
// batch:%s (batchid) -> CommitResult
// seq:%s:%s (batchid) (n) -> checksum

func (j *JournalCache) Commit() error {
	if len(j.Post) > 2 {
		for i, entry := range j.Post {
			if i != len(j.Post)-1 {
				enc, err := entry.Encode()
				if err != nil {
					return err
				}
				err = j.store.BatchPut(hasher.Sum(fmt.Appendf(nil, "chk:%s", entry.GetID())), enc, false)
				if err != nil {
					return err
				}
				err = j.store.BatchPut(hasher.Sum(fmt.Appendf(nil, "seq:%s:%d", hex.EncodeToString(j.treeroot)[:4], i)), []byte(entry.GetID()), false)
				if err != nil {
					return err
				}
			} else {
				enc, err := entry.Encode()
				if err != nil {
					return err
				}
				err = j.store.BatchPut(hasher.Sum(fmt.Appendf(nil, "chk:%s", entry.GetID())), enc, false)
				if err != nil {
					return err
				}
				err = j.store.BatchPut(hasher.Sum(fmt.Appendf(nil, "seq:%s:%d", hex.EncodeToString(j.treeroot)[:4], i)), []byte(entry.GetID()), true)
				if err != nil {
					return err
				}
				j.Post = j.Post[:0]
				return nil
			}
		}
		return fmt.Errorf("out of loop")

	} else {
		for i, entry := range j.Post {
			enc, err := entry.Encode()
			if err != nil {
				return err
			}
			err = j.store.Put(hasher.Sum(fmt.Appendf(nil, "chk:%s", entry.GetID())), enc)
			if err != nil {
				return err
			}
			err = j.store.BatchPut(hasher.Sum(fmt.Appendf(nil, "seq:%s:%d", hex.EncodeToString(j.treeroot)[:4], i)), []byte(entry.GetID()), true)
			if err != nil {
				return err
			}
		}
		return nil // error
	}
}

// id == checksum == hash
func (j *JournalCache) Get(id string) ([]byte, error) {
	return j.store.Get(hasher.Sum(fmt.Appendf(nil, "chk:%s", id)))
}
