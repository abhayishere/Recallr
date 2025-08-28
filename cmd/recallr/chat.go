package main

import (
	"appleNotesRag/internal/llm"
	"appleNotesRag/internal/rag"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type chatReq struct {
	Query string `json:"query"`
}

type chatResp struct {
	Response string   `json:"response"`
	Context  []string `json:"context"`
}

func cmdChat() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Interactive chat with the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("Enter ':exit' to quit.")
			for {
				fmt.Print("> ")
				line, _ := reader.ReadString('\n')
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				if line == ":exit" {
					return nil
				}
				if err := doAsk(line); err != nil {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}
		},
	}
	return cmd
}

func cmdAsk() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Single-turn ask",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			q := strings.TrimSpace(strings.Join(args, " "))
			if q == "" {
				return fmt.Errorf("empty question")
			}
			return doAsk(q)
		},
	}
	return cmd
}

func doAsk(q string) error {
	results, err := rag.SearchNotes(context.Background(), dbPool, embClient, q)
	if err != nil {
		return err
	}
	var contextTexts []string
	for _, res := range results {
		if text, ok := res["content"].(string); ok {
			contextTexts = append(contextTexts, text)
		}
	}
	// Build prompt for LLM
	prompt := "Context:\n" + strings.Join(contextTexts, "\n---\n") + "\n\nUser: " + q
	// Call LLM for response
	llmReq := []llm.Msg{{Role: "user", Content: prompt}}
	answer, err := llmClient.Complete("You are a helpful assistant. Only answer questions using the provided context. If the context does not contain the answer, reply: \"Sorry, I don't have any idea about this.\" Do not use any external knowledge or make assumptions.", llmReq)
	if err != nil {
		return err
	}
	fmt.Println(answer)
	return nil
}
