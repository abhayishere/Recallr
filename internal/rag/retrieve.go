package rag

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// SearchNotes retrieves top relevant note chunks for a query embedding
func SearchNotes(ctx context.Context, db *pgxpool.Pool, embClient interface {
	Embed([]string) ([][]float32, error)
}, query string) ([]map[string]interface{}, error) {
	vecs, err := embClient.Embed([]string{query})
	if err != nil {
		return nil, err
	}
	// Search top 5 similar chunks using pgvector
	rows, err := db.Query(ctx, `SELECT content, note_id, idx FROM note_chunks ORDER BY embedding <-> $1 LIMIT 5`, pgvector.NewVector(vecs[0]))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		var content string
		var noteID string
		var idx int
		if err := rows.Scan(&content, &noteID, &idx); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"content": content,
			"note_id": noteID,
			"idx":     idx,
		})
	}
	return results, nil
}
