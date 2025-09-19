package main

import (
	"context"
	"crosstrace/internal/crossmint"
	"crosstrace/internal/journal"
	"encoding/hex"
	"fmt"
)

type LogEventTool struct {
	Journal journal.JournalStore
}

func (t *LogEventTool) Name() string        { return "log_event" }
func (t *LogEventTool) Description() string { return "Append a new event to the journal" }
func (t *LogEventTool) Call(ctx context.Context, input string) (string, error) {
	pre := ParsePreEntry(input) // also sanitze
	post, err := t.Journal.Append(pre)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Event logged with checksum %s", post), nil
}

type SealBatchTool struct {
	Journal journal.JournalStore
	Anchor  *crossmint.AnchorClient
}

func (t *SealBatchTool) Name() string { return "seal_batch" }
func (t *SealBatchTool) Description() string {
	return "commit journal, anchor root on solana, and mint NFT receipt"
}
func (t *SealBatchTool) Call(ctx context.Context, input string) (string, error) {
	err := t.Journal.BuildTree()
	if err != nil {
		return "", err
	}
	batch, err := t.Journal.BatchInsert()
	if err != nil {
		return "", err
	}
	err = t.Journal.Commit()
	if err != nil {
		return "", err
	}
	txhash, err := t.Anchor.AnchorRoot(hex.EncodeToString(batch.Root[:]))
	if err != nil {
		return "", err
	}
	result, err := crossmint.MintReceiptNFT(ctx, *batch, txhash)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Batch %s sealde. Root=%x, Tx=%s, Claim=%s",
		batch.BatchID, batch.Root[:], txhash, result), nil
}
