package crosstracetools

import (
	"context"
	"crosstrace/internal/crossmint"
	"crosstrace/internal/journal"
	"encoding/hex"
	"fmt"
)

type SealBatchTool struct {
	Journal journal.JournalStore
	Anchor  *crossmint.AnchorClient
}

func (t *SealBatchTool) Name() string { return "seal_batch" }
func (t *SealBatchTool) Description() string {
	return "commit journal, anchor root on solana, and mint NFT receipt"
}

// Ai doesn't need to pass any input all input are sent through LogEvent
func (t *SealBatchTool) Call(ctx context.Context, input string) (string, error) {
	err := t.Journal.BuildTree()
	if err != nil {
		return "", err
	}
	batch, err := t.Journal.Batch()
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
func NewSealBatchTool(cache journal.JournalStore, client *crossmint.AnchorClient) *SealBatchTool {
	return &SealBatchTool{Journal: cache, Anchor: client}
}
