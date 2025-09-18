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

var encoders = encoder.NewEncoder("rlp")

// var cfg = configs.Config{}
var hasher = crypto.NewHasher("sha256")
var JournalConfig configs.JournalConfig

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
	return post.SessionID
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

func SetJournalEncoder() {
	encoders = encoder.NewEncoder(JournalConfig.EncoderName)
}
func SetJournalHasher() {
	hasher = crypto.NewHasher(JournalConfig.HasherName)
}
func SetJournalConfigs(cfgs configs.JournalConfig) {
	JournalConfig = cfgs
}
func Name() string {
	return hasher.Name()
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
	return &PostEntry{}, nil
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

			} else {
				enc, err := entry.Encode()
				if err != nil {
					return err
				}
				item := hasher.Sum([]byte(entry.GetID()))
				j.store.BatchPut(item, enc, true)
				elem = append(elem, item)
				j.Post = j.Post[:0]
				if tree.Insert(elem) {
					if tree.Commit() {
						return nil
					}
				}

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
