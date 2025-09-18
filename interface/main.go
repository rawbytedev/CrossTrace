package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
)

// Step is the incoming payload from Coral Server
type Step struct {
    ID     string      `json:"id"`
    Input  interface{} `json:"input"`
    Params interface{} `json:"params,omitempty"`
}

// StepResult is what we send back to Coral
type StepResult struct {
    ID     string      `json:"id"`
    Output interface{} `json:"output"`
    Status string      `json:"status"` // "success" or "error"
    Error  string      `json:"error,omitempty"`
}

// HandlerFunc is your agent's logic
type HandlerFunc func(Step) StepResult

// Run starts the agent HTTP server
func Run(handler HandlerFunc) error {
    port := os.Getenv("PORT")
    if port == "" {
        port = "5555" // fallback for local testing
    }

    http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
        var step Step
        if err := json.NewDecoder(r.Body).Decode(&step); err != nil {
            http.Error(w, "invalid JSON", http.StatusBadRequest)
            return
        }
        log.Printf("[CoralAgent] Received step: %+v", step)

        result := handler(step)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    })

    addr := fmt.Sprintf(":%s", port)
    log.Printf("[CoralAgent] Listening on %s", addr)
    return http.ListenAndServe(addr, nil)
}



func main() {
    Run(func(step Step) StepResult {
        // Echo back the input with a greeting
        output := fmt.Sprintf("Hello from Go! You said: %v", step.Input)
        return StepResult{
            ID:     step.ID,
            Output: output,
            Status: "success",
        }
    })
}