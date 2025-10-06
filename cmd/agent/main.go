package main

import (
	"context"
	"crosstrace/cmd/agent/tools"
	"crosstrace/internal/configs"
	"crosstrace/internal/crossmint"
	"crosstrace/internal/journal"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/tools"
)

/*
	Agent should be able to call the tools and pass information to them directly parsing input from

Ai then trigerring tools based on them won't be needed
*/
var cfgs configs.Configs

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

func handleAgentOutput(ctx context.Context, output string, tools map[string]tools.Tool) (string, error) {
	var act AgentAction
	if err := json.Unmarshal([]byte(output), &act); err != nil {
		return "", fmt.Errorf("invalid agent output: %w", err)
	}
	if act.Action == "Final Answer" {
		return string(act.ActionInput), nil
	}
	tool, ok := tools[act.Action]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", act.Action)
	}
	return tool.Call(ctx, string(act.Action))
}
func main() {
	cfgs = Setup()
	log.Print("Starting CrossTrace")
	//set global var
	log.Print("Seting globals variable")
	journal.SetAllJournalConfigs(cfgs.Journal)
	crossmint.SetMintConfig(cfgs.Minting)
	client, _ := crossmint.NewAnchorClient(cfgs.Anchor.SolanaRPC, cfgs.Anchor.KeypairPath)
	cache := journal.NewJournalCache(&cfgs.Journal)
	var alltools []tools.Tool
	Logtool := crosstracetools.NewLogEventTool(cache)
	Sealtool := crosstracetools.NewSealBatchTool(cache, client)
	customtools := []tools.Tool{
		Logtool,
		Sealtool,
	}
	alltools = append(alltools, customtools...)
	llm, _ := mistral.New(mistral.WithModel("open-mixtral-8x22b"),
		mistral.WithAPIKey(cfgs.Server.MistralAi),
	)
	agent := agents.NewConversationalAgent(llm, alltools, agents.WithMaxIterations(3))
	// how to pass system message to llm model?
	_ = agent

}
func SetParams() {

}
