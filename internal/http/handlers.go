package http

import (
	"appleNotesRag/internal/emb"
	"appleNotesRag/internal/llm"
	"appleNotesRag/internal/rag"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DB  *pgxpool.Pool
	Emb emb.Client
	LLM llm.ChatClient
}

func NewServer(emb emb.Client, llm llm.ChatClient, db *pgxpool.Pool) *Server {
	return &Server{
		DB:  db,
		Emb: emb,
		LLM: llm,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *Server) Ingest(w http.ResponseWriter, r *http.Request) {
	// var req struct {
	// 	Title      string `json:"title"`
	// 	Text       string `json:"text"`
	// 	SourcePath string `json:"source_path"`
	// 	NoteID     string `json:"note_id"`
	// }
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	// 	return
	// }
	// var id *uuid.UUID
	// if strings.TrimSpace(req.NoteID) != "" {
	// 	tmp, err := uuid.Parse(req.NoteID)
	// 	if err != nil {
	// 		writeJSON(w, 400, map[string]string{"error": "invalid note_id"})
	// 		return
	// 	}
	// 	id = &tmp
	// }

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// noteID, err := rag.IngestNote(ctx, s.DB, s.Emb, req.Title, req.Text, req.SourcePath, id)
	// if err != nil {
	// 	writeJSON(w, 500, map[string]string{"error": err.Error()})
	// 	return
	// }

	// writeJSON(w, http.StatusOK, map[string]string{"status": "ingested", "note_id": noteID.String()})
	err := rag.IngestAllDocs(ctx, s.DB, s.Emb, "./doc")
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ingested"})
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	results, err := rag.SearchNotes(ctx, s.DB, s.Emb, req.Query)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "results": results})
}

// ChatRequest is the request payload for chat
type ChatRequest struct {
	Query string `json:"query"`
}

// ChatResponse is the response payload for chat
type ChatResponse struct {
	Response string   `json:"response"`
	Context  []string `json:"context"`
}

// Chat handler: accepts a query, retrieves relevant chunks, and responds using LLM
func (s *Server) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Retrieve top relevant chunks using embedding similarity
	results, err := rag.SearchNotes(ctx, s.DB, s.Emb, req.Query)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	var contextTexts []string
	for _, res := range results {
		if text, ok := res["content"].(string); ok {
			contextTexts = append(contextTexts, text)
		}
	}
	// Build prompt for LLM
	prompt := "Context:\n" + strings.Join(contextTexts, "\n---\n") + "\n\nUser: " + req.Query
	// Call LLM for response
	llmReq := []llm.Msg{{Role: "user", Content: prompt}}
	answer, err := s.LLM.Complete("You are a helpful assistant. Only answer questions using the provided context. If the context does not contain the answer, reply: \"Sorry, I don't have any idea about this.\" Do not use any external knowledge or make assumptions.", llmReq)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ChatResponse{Response: answer, Context: contextTexts})
}
