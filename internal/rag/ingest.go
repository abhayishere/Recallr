package rag

import (
	"appleNotesRag/internal/chunk"
	"appleNotesRag/internal/emb"
	"context"
	"strings"

	// "io/ioutil" // removed, use os.ReadDir instead
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
)

func IngestNote(ctx context.Context, db *pgxpool.Pool, e emb.Client, title, raw, sourcePath string, noteID *uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	if noteID != nil {
		return *noteID, nil
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	if noteID == nil {
		_, err = tx.Exec(ctx, `INSERT INTO notes (id, title, raw_text, source_path) VALUES ($1,$2,$3,$4)`,
			id, strings.TrimSpace(title), raw, sourcePath)
	} else {
		_, err = tx.Exec(ctx, `UPDATE notes SET title=$2, raw_text=$3, source_path=$4, updated_at=now() WHERE id=$1`,
			id, strings.TrimSpace(title), raw, sourcePath)
	}
	if err != nil {
		return uuid.Nil, err
	}

	// wipe old chunks if updating
	if noteID != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM note_chunks WHERE note_id=$1`, id); err != nil {
			return uuid.Nil, err
		}
	}
	chunks := chunk.SplitByParagraph(raw, 1200, 200)
	if len(chunks) == 0 {
		chunks = []string{raw}
	}

	vecs, err := e.Embed(chunks)
	if err != nil {
		return uuid.Nil, err
	}

	batch := &pgx.Batch{}
	for i, c := range chunks {
		batch.Queue(`INSERT INTO note_chunks (note_id, idx, content, embedding) VALUES ($1,$2,$3,$4)`,
			id, i, c, pgvector.NewVector(vecs[i]))
	}
	br := tx.SendBatch(ctx, batch)
	for range chunks {
		if _, err := br.Exec(); err != nil {
			_ = br.Close()
			return uuid.Nil, err
		}
	}
	if err := br.Close(); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return id, nil

}
func IngestAllDocs(ctx context.Context, db *pgxpool.Pool, e emb.Client, docDir string) error {
	files, err := os.ReadDir(docDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".txt" {
			continue
		}
		path := filepath.Join(docDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = IngestNote(ctx, db, e, file.Name(), string(data), path, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
