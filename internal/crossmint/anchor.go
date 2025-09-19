package crossmint

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/memo"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

// CrqCuK5y425w9CUoqGvXhi8xqh9SCRUz9P8K8ckWiG6F
type AnchorClient struct {
	client *rpc.Client
	Payer  solana.PrivateKey
}

func NewAnchorClient(rpcURL, KeypairPath string) (*AnchorClient, error) {
	c := rpc.New(rpcURL)
	k, err := solana.PrivateKeyFromSolanaKeygenFile(KeypairPath)
	if err != nil {
		return nil, fmt.Errorf("load keypair: %w", err)
	}
	// return address
	// fmt.Print(k.PublicKey().String())
	return &AnchorClient{client: c, Payer: k}, nil
}
func (a *AnchorClient) AnchorRoot(checksumHex string) (string, error) {
	recent, err := a.client.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("blockhash: %w", err)
	}
	bal, err := a.client.GetBalance(context.Background(), a.Payer.PublicKey(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("balance check failed: %w", err)
	}
	if bal.Value < 5000 {
		_, err := a.client.RequestAirdrop(context.Background(), a.Payer.PublicKey(), 1_000_000_000, rpc.CommitmentFinalized)
		if err != nil {
			return "", fmt.Errorf("aidrop request failed: %w", err)
		}
	}
	memoIx := memo.NewMemoInstruction([]byte(checksumHex), a.Payer.PublicKey())

	// 3. Create transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{memoIx.Build()},
		recent.Value.Blockhash,
		solana.TransactionPayer(a.Payer.PublicKey()),
	)
	if err != nil {
		return "", fmt.Errorf("tx: %w", err)
	}

	// 4. Sign with your payer key
	tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(a.Payer.PublicKey()) {
			return &a.Payer
		}
		return nil
	})

	// 5. Send to Devnet
	sig, err := a.client.SendTransactionWithOpts(
		context.Background(),
		tx,
		rpc.TransactionOpts{},
	)
	if err != nil {
		// unwrap JSONRPC “already processed” errors
		if jsonrpcErr, ok := err.(*jsonrpc.RPCError); ok {
			return "", fmt.Errorf("anchor failed: %v", jsonrpcErr.Message)
		}
		return "", err
	}

	return sig.String(), nil
}

