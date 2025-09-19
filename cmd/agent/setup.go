package main

import (
	"context"
	"crosstrace/internal/configs"

	"fmt"
	"log"
	"time"

	//"os"

	mcpadapter "github.com/i2y/langchaingo-mcp-adapter"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/tools"
)

var cfgs configs.Configs

func setuomain() {
	ctx := context.Background()
	coralUrl := "http://localhost:5555/devmode/exampleApplication/privkey/session1/sse?agentId=crosstrace"
	/*url := os.Getenv("CORAL_SSE_URL")
	id := os.Getenv("CORAL_AGENT_ID")
	link := fmt.Sprintf("%s?%s", url, id)
	fmt.Print(link)
	_ = coralUrl*/
	mpc, err := mcpclient.NewSSEMCPClient(coralUrl)
	if err != nil {
		fmt.Print(err)
	}
	err = mpc.Start(ctx)
	if err != nil {
		fmt.Print(err)
	}
	res, err := mpc.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		fmt.Print(err)
		fmt.Print(res.Result)
	}
	defer mpc.Close()
	//defer mpc.Stop()
	adapter, err := mcpadapter.New(mpc)
	if err != nil {
		log.Fatal("Adapter init failed:", err)
	}
	var avtools []tools.Tool
	for i := 0; i < 10; i++ {
		avtools, err = adapter.Tools()
		if err == nil {
			break
		}
		log.Println("Waiting for transport to initialize...")
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to fetch tools: %v", err)
	}
	llm, err := mistral.New(
		mistral.WithAPIKey("qQ4QUBN7nkh6c9zZEPYLnaEYbH1gt1Sc"),
		mistral.WithModel("mistral-large"),
	)
	if err != nil {
		fmt.Print(err)
	}
	//agents.WithPromptFormatInstructions()
	agent := agents.NewOneShotAgent(llm, avtools)
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

func setup() error {
	cfg, err := configs.LoadConfig("config.yaml")
	if err != nil {
		return err
	}
	cfgs = *cfg // copy over to global cfgs
	return nil
}
func LaunchCoral() {
	agen := agents.NewExecutor(&agents.OneShotZeroAgent{})
	agen.GetInputKeys()
	agen.Agent.GetTools()

}
