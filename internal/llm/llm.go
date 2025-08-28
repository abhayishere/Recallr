package llm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type Msg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatClient interface {
	Complete(system string, msg []Msg) (string, error)
}

type Ollama struct {
	BaseURL string
	Model   string
}

func NewOllama() ChatClient {
	return &Ollama{
		BaseURL: os.Getenv("OLLAMA_BASE_URL"),
		Model:   os.Getenv("OLLAMA_MODEL"),
	}
}

type ollamaChatRequest struct {
	Model    string `json:"model"`
	Messages []Msg  `json:"messages"`
	Stream   bool   `json:"stream"` // Make sure this is set to false

}

type ollamaChatResponse struct {
	Message Msg `json:"message"`
}

func (o *Ollama) Complete(system string, msg []Msg) (string, error) {
	body, _ := json.Marshal(ollamaChatRequest{
		Model:    o.Model,
		Messages: msg,
		Stream:   false,
	})

	resp, err := http.Post(o.BaseURL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.Message.Content, nil
}
