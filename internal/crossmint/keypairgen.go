package crossmint

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
)

func Genekeypair() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	keyfile := append(priv.Seed(), pub...)
	data, err := json.Marshal(keyfile)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("keypair.json", data, 0600); err != nil {
		panic(err)
	}
	fmt.Println("Keypair saved to keypair.json")
}

