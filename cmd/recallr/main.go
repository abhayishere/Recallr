package main

import (
	"appleNotesRag/internal/emb"
	"appleNotesRag/internal/llm"
	"context"

	"fmt"
	"os"

	"appleNotesRag/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	flagAPI   string
	embClient emb.Client     // Replace 'Embedder' with the actual type returned by emb.NewOllama
	llmClient llm.ChatClient // Replace 'LLM' with the actual type returned by llm.NewOllama
	dbPool    *pgxpool.Pool
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	embClient = emb.NewOllama(1024)
	llmClient = llm.NewOllama()
	var err error
	dbPool, err = db.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer dbPool.Close()

	root := &cobra.Command{
		Use:   "recallr",
		Short: "Recallr CLI (RAG tools)",
	}

	root.PersistentFlags().StringVar(&flagAPI, "api", "http://localhost:11434", "Base URL for HTTP API")

	root.AddCommand(cmdChat())
	root.AddCommand(cmdAsk())
	root.AddCommand(cmdIngest())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
