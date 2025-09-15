package journal

import (
	"crosstrace/internal/configs"
	"crosstrace/internal/crypto"
	"crosstrace/internal/encoder"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

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

// Handle globalization of configs
func NewJournalConfig(cfgs configs.Config) *JournalConfig {
	cfg = cfgs
	return &JournalConfig{}
}
func SetJournalEncoder(cfg *JournalConfig) {
	encoders = encoder.NewEncoder(cfg.Nameencoder)
}
func SetJournalHasher(cfg *JournalConfig) {
	hasher = crypto.NewHasher(cfg.NameHasher)
}
func SetConfigs(cfgs *configs.Config) {
	cfg = *cfgs
}
func Name() string {
	return hasher.Name()
}

// Handle Sanitization : add global error vars
// change PreEntry/PostEntry to JournalEntry
func SanitizePreEntry(cfg *JournalConfig, pre *PreEntry) (JournalEntry, error) {
	// size check
	if len(pre.raw_msg) > cfg.MaxMsgSize {
		return &PostEntry{}, fmt.Errorf("message exceeds max size: %d > %d", len(pre.raw_msg), cfg.MaxMsgSize)
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
			if cfg.SafeMode {
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
