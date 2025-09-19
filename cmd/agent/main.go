package main

import (
	"context"
	"crosstrace/internal/crossmint"
	"crosstrace/internal/journal"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mcpadapter "github.com/i2y/langchaingo-mcp-adapter"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/tools"
)

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
	journal.SetAllJournalConfigs(cfgs.Journal)
	crossmint.SetMintConfig(cfgs.Minting)
	client, _ := crossmint.NewAnchorClient(cfgs.Anchor.SolanaRPC, cfgs.Anchor.KeypairPath)
	cache := journal.NewJournalCache(&cfgs.Journal)

	customtools := []tools.Tool{
		&LogEventTool{Journal: cache},
		&SealBatchTool{Journal: cache, Anchor: client},
	}
	llm, _ := mistral.New(mistral.WithModel("open-mixtral-8x22b"),
		mistral.WithAPIKey(cfgs.Server.MistralAi),
	)
	instructions := ``
	alltools = append(alltools, customtools...)
	agent := agents.NewOneShotAgent(llm, alltools, agents.WithMaxIterations(3))
	executor := agents.NewExecutor(agent)
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
			result, err:= executor.Call(ctx, payload.Content)
		}
	})
}

func runTurn(ctx context.Context, llm *mistral.Model, exec *agents.Executor, userText string, toolindex map[string]tools.Tool) (string ,error){
	
}
