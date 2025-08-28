package emb

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type Client interface {
	Embed(text []string) ([][]float32, error)
	Dim() int
}

type Ollama struct {
	BaseURL    string
	Model      string
	Dimensions int
}

func NewOllama(dim int) Client {
	return &Ollama{
		BaseURL:    os.Getenv("OLLAMA_BASE_URL_EMB"),
		Model:      os.Getenv("OLLAMA_MODEL_EMB"),
		Dimensions: dim,
	}
}

type ollamaRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Stream bool     `json:"stream"`
	Input  []string `json:"input"`
}

type ollamaResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Embedding  []float32   `json:"embedding"`
}

func (o *Ollama) Embed(text []string) ([][]float32, error) {
	body, _ := json.Marshal(ollamaRequest{
		Model:  o.Model,
		Stream: false,
		Input:  text,
	})

	resp, err := http.Post(o.BaseURL+"/api/embed", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out.Embeddings, nil
}

func (o *Ollama) Dim() int {
	return o.Dimensions
}
