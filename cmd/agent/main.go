package main

import (
	"crosstrace/internal/configs"
	"fmt"
	"log"
	"net/http"
)

var cfg = configs.Config{Port: 1023}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/logEvent", logEvent)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Agent running on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
func logEvent(w http.ResponseWriter, r *http.Request) {
	data := make([]byte, 30)
	r.Body.Read(data)
	print(string(data))
}
