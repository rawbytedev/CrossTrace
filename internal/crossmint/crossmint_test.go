package crossmint

import (
	"crosstrace/internal/configs"
	"crosstrace/internal/journal"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

func TestAnchoring(t *testing.T) {
	new := false
	if new {
		Genekeypair()
	}

	anchor, err := NewAnchorClient(rpc.DevNet_RPC, "keypair.json")
	if err != nil {
		t.Fatal(err)
	}
	rec := GeneConstant()
	root := rec.Root
	result, err := anchor.AnchorRoot(hex.EncodeToString(root[:]))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("signature: %s", result)
}
func TestMinting(t *testing.T) {
	root := make([]byte, 32)
	rand.Read(root)
	rec := journal.CommitResult{
		BatchID:      "Testing123",
		Root:         [32]byte(root),
		Count:        0,
		WindowsStart: time.Now(),
		WindowsEnd:   time.Now(),
	}
	SetMintConfig(configs.MintingConfig{
		CrossmintAPIKey:       "sk_staging_6CJFQGekazgd2bECdmNUF66m7JPD8Ev8JSZerTmSvKX6hAaPUL8jfeRBaaUqVLD1MprP9zgG64AedkkW3xzxe4LiZmWofxwX7KuuxXezvFU4bxBwiGLhkAUnptBZMS8EzFdRx4SrZ6545o1SbHyoS23xz6wNrqvCohx2Q6NwTcjTZx8uwYSm1Zozj3pyNVWzi96qKKFLjZuUQkSvC2DNGzj1", // use you're own API
		CrossmintCollectionID: "cc222c91-a5b9-4bd5-8135-9ba5efc7512b",                                                                                                                                                                                                // use you're own collectionID
		CrossmintBaseURL:      "",
		Recipient:             "email:radiationbolt@gmail.com:solana", // replace 'radiationbolt@gmail.com'
	})
	url, err := MintReceiptNFT(t.Context(), rec, "0x1235")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("claim url: %s", url)
}
func TestMint(t *testing.T) {
}
func GeneConstant() journal.CommitResult {
	cfg := configs.JournalConfig{
		CacheSize:   "10MB",
		DBPath:      "dbfolder",
		DBName:      "badgerdb",
		LogSize:     "10MB",
		EncoderName: "rlp",
		HasherName:  "sha256",
		MaxMsgSize:  30,
	}
	// set all cfg first
	journal.SetAllJournalConfigs(cfg)
	cache := journal.NewJournalCache(&cfg)
	preentry := []*journal.PreEntry{
		journal.NewPreEntry("test1", "user12", "coral", "testing"),
		journal.NewPreEntry("test3", "user4", "coral", "testing"),
		journal.NewPreEntry("test34", "user42", "coral", "testing"),
		journal.NewPreEntry("test5", "user43", "coral", "testing"),
	}
	// sanitize
	var entry []journal.JournalEntry
	for _, en := range preentry {
		post, err := journal.SanitizePreEntry(en)
		if err != nil {
			return journal.CommitResult{}
		}
		entry = append(entry, post)
	}
	for _, en := range entry {
		cache.Append(en)
	}
	cache.BuildTree()
	res, err := cache.BatchInsert()
	if err != nil {
		return journal.CommitResult{}
	}
	err = cache.Commit()
	if err != nil {
		return journal.CommitResult{}
	}
	return *res

}
