package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/tools"
)



func handleAgentOutput(ctx context.Context, output string, tools map[string]tools.Tool) (string, error) {
	var act AgentAction
	if err := json.Unmarshal([]byte(output), &act); err != nil {
		return "", fmt.Errorf("invalide agent output: %w", err)
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
