package journal

import (
	"crosstrace/context"
	mptree "crosstrace/internal/merkle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type RetrieveOptions int

const (
	Checksum RetrieveOptions = iota
	Sequence
	BatchID
)

// Entry point to Journalling

func NewJournalCache(ctx *context.Context) JournalStore {
	db, err := NewLocalStorage(ctx)
	if err != nil {
		fmt.Print(err)
		return nil
	}
	return &JournalCache{ctx: ctx, store: db}
}

// called by main
// testing
// this is received by ai
// Will remove this
func NewPreEntry(ctx *context.Context, raw_msg string, sender_id string, source string, session_id string) *PreEntry {
	return &PreEntry{ctx: ctx, raw_msg: raw_msg, sender_id: sender_id, source: source, session_id: session_id}
}
func NewPostEntryWithCtx(ctx *context.Context) *PostEntry {
	return &PostEntry{ctx: ctx}
}

func (pre *PreEntry) GetID() string {
	return pre.session_id
}
func (pre *PreEntry) GetTimestamp() time.Time {
	return pre.timestamp
}

// needs to improve encoders implementation for easier use
func (pre *PreEntry) Encode() ([]byte, error) {
	return pre.ctx.Encoder.Encode(pre)
}
func (pre *PreEntry) Decode(data []byte) error {
	return pre.ctx.Encoder.Decode(data, pre)
}
func (pre *PreEntry) Sum() []byte {
	data, err := pre.Encode()
	if err != nil {
		return nil
	}
	return pre.ctx.Hasher.Sum(data)
}

func (post *PostEntry) GetID() string {
	return post.Checksum
}
func (post *PostEntry) GetTimestamp() time.Time {
	return post.Timestamp
}

func (post *PostEntry) Encode() ([]byte, error) {
	return post.ctx.Encoder.Encode(post)
}
func (post *PostEntry) Decode(data []byte) error {
	return post.ctx.Encoder.Decode(data, post)
}
func (res *CommitResult) Encode() ([]byte, error) {
	return res.ctx.Encoder.Encode(res)
}
func (res *CommitResult) Decode(data []byte) error {
	return res.ctx.Encoder.Decode(data, res)
}
func (res *Commitment) Encode() ([]byte, error) {
	return res.ctx.Encoder.Encode(res)
}
func (res *Commitment) Decode(data []byte) error {
	return res.ctx.Encoder.Decode(data, res)
}

// set based on configuration
// that means if config are changed during run
// simply call set JournalConfigs and other setJournal
// to change package configuration

// Handle Sanitization : add global error vars
// change PreEntry/PostEntry to JournalEntry
func SanitizePreEntry(ctx *context.Context, pre *PreEntry) (JournalEntry, error) {
	// size check

	if len(pre.raw_msg) > ctx.Journal.MaxMsgSize {
		return &PostEntry{}, fmt.Errorf("message exceeds max size: %d > %d", len(pre.raw_msg), ctx.Journal.MaxMsgSize)
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
			if ctx.Journal.SafeMode {
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
	checksum := ctx.Hasher.Sum([]byte(checksumInput))
	checksumHex := hex.EncodeToString(checksum[:])
	_ = checksumHex
	// Return safe PostEntry
	// Create NewPostEntry fot return
	return &PostEntry{
		ctx:       ctx,
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

// prepares a batch
func (j *JournalCache) Batch() (*CommitResult, error) {
	if uint32(len(j.Post)) == 0 {
		if j.commitRes != nil {
			return j.commitRes, nil
		}
		return nil, nil
	}
	batch := CommitResult{
		ctx:          j.ctx,
		Root:         [32]byte(j.treeroot),
		Count:        uint32(len(j.Post)),
		WindowsStart: j.Post[0].GetTimestamp(),
		WindowsEnd:   j.Post[len(j.Post)-1].GetTimestamp(),
		version:      "v1",
	}
	enc, err := batch.Encode()
	if err != nil {
		return &CommitResult{}, err
	}
	batchcommitment := j.ctx.Hasher.Sum(enc)
	batchdata := &Commitment{
		ctx:          j.ctx,
		Roothash:     [32]byte(j.treeroot),
		Count:        uint32(len(j.Post)),
		WindowsStart: j.Post[0].GetTimestamp(),
		WindowsEnd:   j.Post[len(j.Post)-1].GetTimestamp(),
		commitment:   batchcommitment,
	}
	data, err := batchdata.Encode()
	if err != nil {
		return &CommitResult{}, err
	}
	j.batchid = batchcommitment
	batch.batchID = string(batchcommitment)
	j.commitRes = &batch
	return &batch, j.store.Put(fmt.Appendf(nil, "b:%x", batchcommitment), data)
}

func (j *JournalCache) Commit() error {
	_, err := j.Batch()
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	// j.Post get zerro when it goes low
	// we can't get size from it at that point
	// at this point seems like j contents get corrupted? need to investigate
	//this fix it temporaly
	size := len(j.Post)
	if size > 1 {
		return j.largeCommit()
	}
	return nil
}
func (j *JournalCache) hash(data []byte) []byte {
	return j.ctx.Hasher.Sum(data)
}

// used to peform largecommit when elems are >1
func (j *JournalCache) largeCommit() error {
	size := len(j.Post)
	batchid := j.batchid
	for i, entry := range j.Post {
		enc, err := entry.Encode()
		if err != nil {
			return err
		}
		err = j.store.BatchPut(fmt.Appendf(nil, "e:%x", entry.GetID()), enc)
		if err != nil {
			return err
		}
		err = j.store.BatchPut(fmt.Appendf(nil, "p:%x:%d", batchid, i), []byte(entry.GetID()))
		if err != nil {
			return err
		}
		err = j.store.BatchPut(fmt.Appendf(nil, "r:%s", entry.GetID()), batchid)
		if err != nil {
			return err
		}
		// once for last elem
		// for testing the problem
		// replace size-1 with len(j.Post)
		// might be because it's a pointer?
		// can be done after loop
		if i == size-1 {
			err = j.store.BatchPut(nil, nil)
			if err != nil {
				return err
			}
		}
		j.RoolBack()
	}
	return nil
}

// id == checksum == hash
// hash it according to type before calling it
func (j *JournalCache) Get(id string) ([]byte, error) {
	item, err := hex.DecodeString(id)
	if err != nil {
		return nil, err
	}
	return j.store.Get(item)
}

// clean everything / for now it can only clear
func (j *JournalCache) RoolBack() {
	j.Post = j.Post[:0]
	j.batchid = j.batchid[:0]
	j.treeroot = nil
}
func (j *JournalCache) Close() error {
	j.RoolBack()           // safety
	return j.store.Close() // direct close
}

// small Format implementation
func Format(s string, opts ...RetrieveOptions) string {
	return hex.EncodeToString(fmt.Appendf(nil, "e:%x", s))
}

func FormatSeq(s string, n int) string {
	return hex.EncodeToString(fmt.Appendf(nil, "p:%x:%d", []byte(s), n))
}
func FormatBatch(s string) string {
	return hex.EncodeToString(fmt.Appendf(nil, "b:%x", s))
}
