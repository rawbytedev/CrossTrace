package main

import (
	"crosstrace/internal/configs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"time"
)

var cfg = configs.Config{Port: 1023}

// PreEntry matches your journal.PreEntry shape
type PreEntry struct {
	SenderID  string `json:"sender_id"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

func logEventHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var pre PreEntry
	if err := json.NewDecoder(r.Body).Decode(&pre); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Log to stdout for now
	log.Printf("[logEvent] sender=%s session=%s msg=%q", pre.SenderID, pre.SessionID, pre.Message)

	// Respond with dummy root + timestamp for now
	resp := map[string]any{
		"status":    "ok",
		"received":  pre,
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"root":      fmt.Sprintf("%x", []byte("dummyroot")),
		"latency":   time.Since(start).String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func smain() {
	mux := http.NewServeMux()
	mux.HandleFunc("/logEvent", logEventHandler)

	addr := ":1023"
	log.Printf("CrossTrace test agent listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

/*
func main() {
	cfg, err := configs.LoadConfig(os.Getenv("CONFIG_PATH"))
if err != nil {
    log.Fatal("failed to load config:", err)
}

	mux := http.NewServeMux()
	mux.HandleFunc("/logEvent", logEvent)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Agent running on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
func  logEvent(w http.ResponseWriter, r *http.Request) {
    var pre journal.PreEntry
    if err := json.NewDecoder(r.Body).Decode(&pre); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    post, err := journal.SanitizePreEntry(pre)
    if err != nil {
        http.Error(w, "sanitize failed: "+err.Error(), http.StatusBadRequest)
        return
    }

    if err := a.journal.Append(post); err != nil {
        http.Error(w, "append failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Optional: trigger commit if batch full
    if a.journal.ShouldCommit() {
        root, err := a.journal.Commit()
        if err != nil {
            http.Error(w, "commit failed: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Anchor root to Solana
        sig, err := crossmint.AnchorChecksumViaMemo(a.cfg.Anchor.SolanaRPC, root)
        if err != nil {
            http.Error(w, "anchor failed: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Mint NFT
        nftURL, err := crossmint.MintReceiptNFT(a.cfg.Minting.CrossmintAPIKey, root)
        if err != nil {
            http.Error(w, "mint failed: "+err.Error(), http.StatusInternalServerError)
            return
        }

        json.NewEncoder(w).Encode(map[string]any{
            "root": fmt.Sprintf("%x", root),
            "anchor_tx": sig,
            "nft_url": nftURL,
        })
        return
    }

    w.WriteHeader(http.StatusAccepted)
}*/
