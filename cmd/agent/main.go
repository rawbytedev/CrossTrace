package main

import (
	"context"
	"crosstrace/internal/configs"
	"crosstrace/internal/crossmint"
	"crosstrace/internal/journal"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mcpadapter "github.com/i2y/langchaingo-mcp-adapter"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

var cfgs configs.Configs

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

func ParsePreEntry(input string) journal.JournalEntry {
	return &journal.PostEntry{}
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

type AgentAction struct {
	Action      string          `json:"action"`
	ActionInput json.RawMessage `json:"action_input"`
}

func Setup() configs.Configs {
	cfg, err := configs.LoadConfig("config.yaml")
	if err != nil {
		return configs.Configs{}
	}
	return *cfg
}
func NewLogEventTool(cache journal.JournalStore) *LogEventTool {
	return &LogEventTool{Journal: cache}
}
func NewSealBatchTool(cache journal.JournalStore, client *crossmint.AnchorClient) *SealBatchTool {
	return &SealBatchTool{Journal: cache, Anchor: client}
}
func main() {
	ctx := context.Background()
	coralUrl := "http://localhost:5555/devmode/exampleApplication/privkey/session1/sse?agentId=crosstrace"
	/*url := os.Getenv("CORAL_SSE_URL")
	id := os.Getenv("CORAL_AGENT_ID")
	link := fmt.Sprintf("%s?%s", url, id)
	fmt.Print(link)
	_ = coralUrl*/
	log.Printf("Starting MCP client and connecting to SSE server : %s", coralUrl)
	mpc, err := mcpclient.NewSSEMCPClient(coralUrl)
	if err != nil {
		log.Fatalf("error while setting mcpclient %v", err)
	}
	err = mpc.Start(ctx)
	if err != nil {
		log.Fatalf("error while starting mcpclient %v", err)
	}
	_, err = mpc.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		log.Fatalf("error while initializing mcp client %v", err)
	}
	defer mpc.Close()
	adapter, err := mcpadapter.New(mpc)
	if err != nil {
		log.Fatal("Adapter init failed:", err)
	}
	// retry in case server connection takes time
	// in case of issue with server
	var alltools []tools.Tool
	for i := 0; i < 10; i++ {
		alltools, err = adapter.Tools()
		if err == nil {
			break
		}
		log.Println("Waiting for transport to initialize...")
		time.Sleep(1 * time.Second)
	}
	// initilizing internal startup
	cfgs = Setup()
	journal.SetAllJournalConfigs(cfgs.Journal)
	crossmint.SetMintConfig(cfgs.Minting)
	client, _ := crossmint.NewAnchorClient(cfgs.Anchor.SolanaRPC, cfgs.Anchor.KeypairPath)
	cache := journal.NewJournalCache(&cfgs.Journal)
	tool1 := NewLogEventTool(cache)
	tool2 := NewSealBatchTool(cache, client)
	customtools := []tools.Tool{
		tool1,
		tool2,
	}
	llm, _ := mistral.New(mistral.WithModel("open-mixtral-8x22b"),
		mistral.WithAPIKey(cfgs.Server.MistralAi),
	)
	//instructions := ``
	alltools = append(alltools, customtools...)

	agent := agents.NewConversationalAgent(llm, alltools, agents.WithMaxIterations(3))
	agent.Chain = chains.NewLLMChain(llm, prompts.ChatPromptTemplate{
		Messages: []prompts.MessageFormatter{
			prompts.SystemMessagePromptTemplate{prompts.PromptTemplate{
				Template: "You are CrossTrace, an AI assistant that helps users log supply chain events and mint NFT receipts on Solana via Crossmint."},
			},
		},
	}, chains.WithMaxTokens(1000000))

	mpc.OnNotification(func(notification mcp.JSONRPCNotification) {
		log.Printf("Got notification: %s", notification.Method)
		if notification.Method == "message" {
			var payload struct {
				SessionId string `json:"sessionid"`
				Sender    string `json:"sender"`
				Content   string `json:"content"`
			}
			data, err := notification.Params.MarshalJSON()
			if err != nil {
				log.Fatalf("error while marshalling: %s", err)
			}
			if err = json.Unmarshal(data, &payload); err != nil {
				log.Fatalf("error while unmarshalling: %s", err)
			}
			fmt.Printf("User %s said: %s\n", payload.Sender, payload.Content)
			//result, err:= executor.Call(ctx, payload.Content)
		}
	})
}
