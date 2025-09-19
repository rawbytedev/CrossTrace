package crossmint

import (
	"bytes"
	"context"
	"crosstrace/internal/configs"
	"crosstrace/internal/journal"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var MintConfig configs.MintingConfig


func SetMintConfig(cfg configs.MintingConfig) {
	MintConfig = cfg
}

type MintResponse struct {
	ID       string `json:"id"`
	ActionID string `json:"actionId"`
	OnChain  struct {
		Status string `json:"status"`
		Chain  string `json:"chain"`
	} `json:"onChain"`
}
type Minter struct{

}

// MintReceiptNFT sends a mint request to Crossmint and returns the mint ID.
func MintReceiptNFT(ctx context.Context, rec journal.CommitResult, tx string) (string, error) {

	endpoint := "https://staging.crossmint.com/api/2022-06-09"
	if MintConfig.CrossmintBaseURL != "" {
		endpoint = MintConfig.CrossmintBaseURL
	}
	payload := map[string]interface{}{
		"recipient": MintConfig.Recipient,
		"metadata": map[string]interface{}{
			"name":        "CrossTrace Batch Receipt",
			"image":       "https://placehold.co/600x400.png", // required field
			"description": fmt.Sprintf("Merkle root receipt for batch %s", rec.BatchID),
			"attributes": []map[string]string{
				{"trait_type": "merkle_root", "value": fmt.Sprintf("%x", rec.Root[:])},
				{"trait_type": "batch_id", "value": rec.BatchID},
				{"trait_type": "count", "value": fmt.Sprintf("%d", rec.Count)},
				{"trait_type": "solana_tx", "value": tx},
			},
			"sendNotification": true,
		},
	}

	body, _ := json.Marshal(payload)
	// change to MintConfig.BaseUrl
	url := fmt.Sprintf("%s/collections/%s/nfts", endpoint, MintConfig.CrossmintCollectionID)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", MintConfig.CrossmintAPIKey)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("crossmint request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err!= nil{

		}
		return "", fmt.Errorf("crossmint mint failed: %d: %s", resp.StatusCode, body)
	}
	var out MintResponse

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.OnChain.Status, nil
}
