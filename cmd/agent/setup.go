package main

import (
	"context"
	"crosstrace/internal/configs"
	"crosstrace/internal/journal"
	"fmt"
	"os"

	//"os"

	mcpadapter "github.com/i2y/langchaingo-mcp-adapter"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/mistral"
)

var journalcfg journal.JournalConfig
var cfgs configs.Configs

func main() {
	ctx := context.Background()
	coralUrl := "http://localhost:5555/devmode/exampleApplication/privkey/session1/sse?agentId=setup"
	url := os.Getenv("CORAL_SSE_URL")
	id := os.Getenv("CORAL_AGENT_ID")
	link := fmt.Sprintf("%s?%s", url, id)
	fmt.Print(link)
	_ = coralUrl
	mpc, err := mcpclient.NewSSEMCPClient(coralUrl, nil)
	if err != nil {
		fmt.Print(err)
	}
	mpc.Start(ctx)
	defer mpc.Close()
	//defer mpc.Stop()
	adapter, err := mcpadapter.New(mpc)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print("Tools:")
	mcptools, err := adapter.Tools()
	fmt.Print(mcptools)
	if err != nil {
		fmt.Print(err)
	}
	llm, err := mistral.New(
		mistral.WithAPIKey("APIKEY"),
	)
	if err != nil {
		fmt.Print(err)
	}

	agent := agents.NewOneShotAgent(llm, mcptools, agents.WithMaxIterations(3))
	exevutor := agents.NewExecutor(agent)
	qest := "can you tell me the wvalue of life"
	result, err := chains.Run(
		ctx, exevutor, qest,
	)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(result)
}
func LaunchCoral() {
	agen := agents.NewExecutor(&agents.OneShotZeroAgent{})
	agen.GetInputKeys()
	agen.Agent.GetTools()

}
