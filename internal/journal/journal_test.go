package journal

import (
	"crosstrace/internal/configs"
	"crypto/rand"
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

func GeneRandomPreEntry() []*PreEntry {
	var items []*PreEntry
	for range 7 {
		items = append(items, &PreEntry{
			sender_id:  rand.Text()[1:],
			raw_msg:    rand.Text(),
			timestamp:  time.Now(),
			source:     "test",
			session_id: "testing_Id",
		})
	}
	items = append(items, &PreEntry{
		sender_id:  "12",
		raw_msg:    "HelloThereThisisbeyongthemaxlenghtofmessage",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
		sender_id:  "12",
		raw_msg:    "In", // add invalide utf-8 character
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})

	return items
}
func GeneConstantPreEntry() []*PreEntry {
	var items []*PreEntry
	items = append(items, &PreEntry{
		sender_id:  "12",
		raw_msg:    "HelloThereThisisbeyongthemaxlenghtofmessage",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
		sender_id:  "129",
		raw_msg:    "Hello message",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
		sender_id:  "152",
		raw_msg:    "World",
		timestamp:  time.Now(),
		source:     "custom",
		session_id: "testing_Id",
	})
	items = append(items, &PreEntry{
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
	SetAllJournalConfigs(*cfg)
	journal := NewJournalCache(cfg)
	bad_entries := GeneRandomPreEntry()
	var san_entries []JournalEntry
	// we expect to run into some bad entries
	for i, entry := range bad_entries {
		res, err := SanitizePreEntry(entry)
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
		var v PostEntry
		data, err := journal.Get(entry.GetID())
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

}
// This test ensure that we can query database for data even after restarting
func TestJournalInsertGet(t *testing.T) {
	new := false
	if new {
		cfg := NewJournalConfig()
		SetAllJournalConfigs(*cfg)
		journal := NewJournalCache(cfg)
		bad_entries := GeneConstantPreEntry()
		var san_entries []JournalEntry
		// we expect to run into some bad entries
		for i, entry := range bad_entries {
			res, err := SanitizePreEntry(entry)
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
			var v PostEntry
			data, err := journal.Get(entry.GetID())
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
	} else {
		cfg := NewJournalConfig()
		SetAllJournalConfigs(*cfg)
		journal := NewJournalCache(cfg)
		bad_entries := GeneConstantPreEntry()
		var san_entries []JournalEntry
		// we expect to run into some bad entries
		for i, entry := range bad_entries {
			res, err := SanitizePreEntry(entry)
			if err != nil {
				t.Log(err)
				t.Logf("Bad entry %d", i)
				break
			}
			san_entries = append(san_entries, res)
		}
		for _, entry := range san_entries {
			var v PostEntry
			data, err := journal.Get(entry.GetID())
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
	}
}
