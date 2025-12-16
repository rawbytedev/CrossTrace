package journal

import (
	"bytes"
	"crosstrace/context"
	"crosstrace/internal/configs"
	"crosstrace/internal/crypto"
	"crosstrace/internal/encoder"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"testing"
	"time"
)

func NewJournalConfig() *configs.JournalConfig {
	return &configs.JournalConfig{
		CacheSize:   "10MB",
		DBPath:      "dbfolder",
		DBName:      "badgerdb",
		LogSize:     "10MB",
		EncoderName: "rlp",
		HasherName:  "sha256",
		MaxMsgSize:  30,
	}
}

func GeneRandomPreEntry(ctx *context.Context) []*PreEntry {
	var items []*PreEntry
	for range 7 {
		items = append(items, &PreEntry{
			ctx:        ctx,
			sender_id:  rand.Text()[1:],
			raw_msg:    rand.Text(),
			timestamp:  time.Now(),
			source:     "test",
			session_id: "testing_Id",
		})
	}
	items = append(items, &PreEntry{
		ctx:        ctx,
		sender_id:  "12",
		raw_msg:    "HelloThereThisisbeyongthemaxlenghtofmessage",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
		ctx:        ctx,
		sender_id:  "12",
		raw_msg:    "In", // add invalide utf-8 character
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})

	return items
}
func GeneConstantPreEntry(ctx *context.Context) []*PreEntry {
	var items []*PreEntry
	items = append(items, &PreEntry{
		ctx:        ctx,
		sender_id:  "12",
		raw_msg:    "HelloThereThisisbeyongthemaxlenghtofmessage",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
		ctx:        ctx,
		sender_id:  "129",
		raw_msg:    "Hello message",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
		ctx:        ctx,
		sender_id:  "152",
		raw_msg:    "World",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
		ctx:        ctx,
		sender_id:  "12",
		raw_msg:    "In", // add invalide utf-8 character
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	return items
}
func TestJournalInsert(t *testing.T) {
	cfg := NewJournalConfig()
	ctx := context.Context{Journal: *cfg}
	ctx.Encoder = encoder.NewEncoder(cfg.EncoderName)
	ctx.Hasher = crypto.NewHasher(cfg.HasherName)
	journal := NewJournalCache(&ctx)
	bad_entries := GeneRandomPreEntry(&ctx)
	var san_entries []JournalEntry
	// we expect to run into some bad entries
	for i, entry := range bad_entries {
		res, err := SanitizePreEntry(&ctx, entry)
		if err != nil {
			t.Log(err)
			t.Logf("Bad entry %d", i)
			continue
		}
		san_entries = append(san_entries, res)
	}
	for _, entry := range san_entries {
		_, err := journal.Append(entry)
		if err != nil {
			t.Fatal(err)
		}
	}
	err := journal.BuildTree()
	if err != nil {
		t.Fatal(err)
	}
	_, err = journal.Batch()
	if err != nil {
		t.Fatal(err)
	}
	err = journal.Commit()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range san_entries {
		v := NewPostEntryWithCtx(&ctx)
		data, err := journal.Get(Format(entry.GetID()))
		if err != nil {
			t.Fatal(err)
		}
		err = v.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		if v.Checksum != entry.GetID() {
			t.Fatal("checksum mismatch")
		}
	}
	err = journal.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// This test ensure that we can query database for data even after restarting
func TestJournalInsertGet(t *testing.T) {
	new := true

	if new {
		cfg := NewJournalConfig()
		ctx := context.Context{Journal: *cfg}
		ctx.Encoder = encoder.NewEncoder(cfg.EncoderName)
		ctx.Hasher = crypto.NewHasher(cfg.HasherName)
		journal := NewJournalCache(&ctx)
		bad_entries := GeneRandomPreEntry(&ctx)
		var san_entries []JournalEntry
		// we expect to run into some bad entries
		for i, entry := range bad_entries {
			res, err := SanitizePreEntry(&ctx, entry)
			if err != nil {
				t.Log(err)
				t.Logf("Bad entry %d", i)
				continue
			}
			san_entries = append(san_entries, res)
		}
		for _, entry := range san_entries {
			_, err := journal.Append(entry)
			if err != nil {
				t.Fatal(err)
			}
		}
		err := journal.BuildTree()
		if err != nil {
			t.Fatal(err)
		}
		err = journal.Commit()
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range san_entries {
			v := NewPostEntryWithCtx(&ctx)
			data, err := journal.Get(Format(entry.GetID()))
			if err != nil {
				t.Fatal(err)
			}
			err = v.Decode(data)
			if err != nil {
				t.Fatal(err)
			}
			if v.Checksum != entry.GetID() {
				t.Fatal("checksum mismatch")
			}
		}
		err = journal.Close()
		if err != nil {
			t.Fatal(err)
		}
	} else {
		cfg := NewJournalConfig()
		ctx := context.Context{Journal: *cfg}
		ctx.Encoder = encoder.NewEncoder(cfg.EncoderName)
		ctx.Hasher = crypto.NewHasher(cfg.HasherName)
		journal := NewJournalCache(&ctx)
		bad_entries := GeneRandomPreEntry(&ctx)
		var san_entries []JournalEntry
		// we expect to run into some bad entries
		for i, entry := range bad_entries {
			res, err := SanitizePreEntry(&ctx, entry)
			if err != nil {
				t.Log(err)
				t.Logf("Bad entry %d", i)
				break
			}
			san_entries = append(san_entries, res)
		}
		for _, entry := range san_entries {
			v := NewPostEntryWithCtx(&ctx)
			data, err := journal.Get(Format(entry.GetID()))
			if err != nil {
				t.Fatal(err)
			}
			err = v.Decode(data)
			if err != nil {
				t.Fatal(err)
			}
			if v.Checksum != entry.GetID() {
				t.Fatal("checksum mismatch")
			}

		}
		err := journal.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestBatchQuery(t *testing.T) {
	new := true
	if new {
		cfg := NewJournalConfig()
		ctx := context.Context{Journal: *cfg}
		ctx.Encoder = encoder.NewEncoder(cfg.EncoderName)
		ctx.Hasher = crypto.NewHasher(cfg.HasherName)
		journal := NewJournalCache(&ctx)
		bad_entries := GeneRandomPreEntry(&ctx)
		var san_entries []JournalEntry
		// we expect to run into some bad entries
		for i, entry := range bad_entries {
			res, err := SanitizePreEntry(&ctx, entry)
			if err != nil {
				t.Log(err)
				t.Logf("Bad entry %d", i)
				continue
			}
			san_entries = append(san_entries, res)
		}
		for _, entry := range san_entries {
			_, err := journal.Append(entry)
			if err != nil {
				t.Fatal(err)
			}
		}
		err := journal.BuildTree()
		if err != nil {
			t.Fatal(err)
		}
		com, err := journal.Batch()
		if err != nil {
			t.Fatal(err)
		}
		err = journal.Commit()
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range san_entries {
			v := NewPostEntryWithCtx(&ctx)
			data, err := journal.Get(Format(entry.GetID()))
			if err != nil {
				t.Fatal(err)
			}
			err = v.Decode(data)
			if err != nil {
				t.Fatal(err)
			}
			if v.Checksum != entry.GetID() {
				t.Fatal("checksum mismatch")
			}
		}
		/*
			Format Sequence Test
		*/
		for i := range com.Count {
			// retrieving checksum from batchID and at index i
			// the comparing with records
			data, err := journal.Get(FormatSeq(com.BatchID, int(i)))
			if err != nil {
				t.Log(err)
				t.Fatal("sequence while using Format for Seq")
			}
			if string(data) != san_entries[i].GetID() {
				t.Fatal("checksum mismatch")
			}
		}
		t.Log("Retrieving Batch")
		var v Commitment
		v.ctx = &ctx
		data, err := journal.Get(FormatBatch(com.BatchID))
		if err != nil {
			t.Fatal(err)
		}
		err = v.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		if v.Roothash != com.Root {
			t.Fatal("mismatch Roots")
		}
		err = journal.Close()
		if err != nil {
			t.Fatal(err)
		}
	} else {
		// only peform query doesn't write to db
		_ = CommitResult{
			BatchID: "09dd1d47d7f0e5dfac278513a723b6d424558669feb014aecf5afce040c18211",
			Root:    [32]byte{89, 82, 203, 230, 157, 145, 229, 24, 119, 35, 162, 39, 108, 37, 209, 71, 3, 171, 242, 49, 6, 1, 84, 104, 252, 65, 22, 173, 7, 180, 233, 189},
			Count:   0x3,
		}
		cfg := NewJournalConfig()
		ctx := context.Context{Journal: *cfg}
		ctx.Encoder = encoder.NewEncoder(cfg.EncoderName)
		ctx.Hasher = crypto.NewHasher(cfg.HasherName)
		journal := NewJournalCache(&ctx)
		bad_entries := GeneRandomPreEntry(&ctx)
		var san_entries []JournalEntry
		// we expect to run into some bad entries
		for i, entry := range bad_entries {
			res, err := SanitizePreEntry(&ctx, entry)
			if err != nil {
				t.Log(err)
				t.Logf("Bad entry %d", i)
				break
			}
			san_entries = append(san_entries, res)
		}
		for _, entry := range san_entries {
			v := NewPostEntryWithCtx(&ctx)
			data, err := journal.Get(Format(entry.GetID()))
			if err != nil {
				t.Fatal(err)
			}
			err = v.Decode(data)
			if err != nil {
				t.Fatal(err)
			}
			if v.Checksum != entry.GetID() {
				t.Fatal("checksum mismatch")
			}
		}
		err := journal.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

// seq == sequence it allow to track down event
/*
Assuming that a batch has 10 events we compute the batchid(unique identified) and stores it
with batch metadata
each event is store individually using chk:(id of event) as key and event content as value
each seq represent an event stored in order seq:%s:%d -> id of event
%s represent the batchid and %d is the position of event in the whole batch
*/
func TestFormatSeq(t *testing.T) {
	cfg := NewJournalConfig()
	ctx := context.Context{Journal: *cfg}
	ctx.Hasher = crypto.NewHasher(ctx.Journal.HasherName)
	s := hex.EncodeToString(ctx.Hasher.Sum(fmt.Appendf(nil, "seq:%s:%d", "12", 1)))
	b := FormatSeq("12", 1) // this only formats into seq:%s:%d
	dat1, _ := hex.DecodeString(s)
	val, _ := hex.DecodeString(b)
	dat2 := ctx.Hasher.Sum(val)
	if !bytes.Equal(dat1, dat2) {
		print("Bad")
	}
}
