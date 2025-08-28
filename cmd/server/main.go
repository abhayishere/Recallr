package server

import (
	"appleNotesRag/internal/db"
	"appleNotesRag/internal/emb"
	httpapi "appleNotesRag/internal/http"
	"appleNotesRag/internal/llm"
	"context"
	"net/http"

	"github.com/joho/godotenv"
)

func Serve() {
	_ = godotenv.Load()
	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	embClient := emb.NewOllama(1024)
	llmClient := llm.NewOllama()

	s := httpapi.NewServer(embClient, llmClient, pool)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.Health)
	mux.HandleFunc("/ingest", s.Ingest)
	mux.HandleFunc("/search", s.Search)
	mux.HandleFunc("/chat", s.Chat)

	http.ListenAndServe(":8080", mux)
}
