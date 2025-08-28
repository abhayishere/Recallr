package main

import (
	"context"
	"fmt"
	"path/filepath"

	"appleNotesRag/internal/db"
	"appleNotesRag/internal/emb"
	"appleNotesRag/internal/rag"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func cmdIngest() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest all .txt files from a directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = godotenv.Load()
			d := dir
			if d == "" {
				d = "doc"
			}
			// Resolve to absolute for clarity
			abs, _ := filepath.Abs(d)
			ctx := context.Background()
			pool, err := db.Connect(ctx)
			if err != nil {
				return err
			}
			defer pool.Close()
			e := emb.NewOllama(1024) // adjust if you change dims
			if err := rag.IngestAllDocs(ctx, pool, e, abs); err != nil {
				return err
			}
			fmt.Println("Ingestion complete from:", abs)
			return nil
		},
	}
	cmd.Flags().StringVarP(&dir, "dir", "d", "doc", "Directory with .txt files")
	return cmd
}
