# Recallr (RAG prototype in Go)

Simple Retrieval-Augmented Generation (RAG) prototype:
- Ingest text, chunk, embed, and store in PostgreSQL with pgvector.
- Retrieve top-k chunks by vector similarity.
- Chat endpoint calls an LLM with retrieved context.

## Prerequisites
- Go (1.21+ recommended)
- Docker (for PostgreSQL + pgvector)
- Ollama (or your preferred embedding/LLM provider)
- A configured `.env` (see example below)

## Quick Start

1) Start Postgres + pgvector (using docker-compose)

```sh
docker compose up -d
```

2) Create extension and tables

Option A: Use the Makefile (macOS/BSD `sed` supported)
```sh
make migrate MODEL_VECTOR_SIZE=768
```

Option B: Copy and run the migration manually
```sh
docker cp migrations/001_init.sql my-postgres:/001_init.sql
docker exec -it my-postgres psql -U myuser -d mydatabase -f /001_init.sql
```

Notes:
- If you change embedding model dimensions, update the `VECTOR(<dim>)` in `001_init.sql` (or pass `MODEL_VECTOR_SIZE` to the Makefile) and re-run the migration.
- Ensure `CREATE EXTENSION IF NOT EXISTS vector;` is present in SQL when initializing a fresh DB.

3) Configure environment

Create `.env` in project root (example):
```env
DATABASE_URL=postgres://rag:rag@localhost:5432/ragdb?sslmode=disable

EMBEDDINGS_PROVIDER=ollama
EMBEDDINGS_MODEL=bge-m3          # or nomic-embed-text, all-minilm-l6-v2
EMBEDDINGS_BASE_URL=http://localhost:11434

LLM_PROVIDER=ollama
LLM_MODEL=gemma3:1b
LLM_BASE_URL=http://localhost:11434
```
Make sure the models are pulled in Ollama (e.g., `ollama pull bge-m3`, `ollama pull gemma3:1b`).

4) Run the server
```sh
go run main.go
```
The API listens on `http://localhost:8080` (adjust if you wired a different port).

## API

- Health
```sh
curl -s http://localhost:8080/health
```

- Ingest a note
```sh
curl -s -X POST http://localhost:8080/ingest \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Doc 1",
    "text": "Your text content here...",
    "source_path": "doc/doc1.txt"
  }'
```

- Semantic search (top-k server default)
```sh
curl -s -X POST http://localhost:8080/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"What does the doc say about X?"}'
```

- Chat (RAG)
```sh
curl -s -X POST http://localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"query":"Ask your question"}'
```
If no relevant context is found, the system prompt can instruct the model to answer: "Sorry, I don't have any idea about this."

## Ingesting files from `doc/`
`internal/rag/ingest.go` provides `IngestAllDocs(ctx, db, emb, docDir)` to read `.txt` files and ingest them. You can wire a temporary CLI/endpoint to call it, or ingest via the `/ingest` endpoint repeatedly.

## Troubleshooting
- ERROR: type "vector" does not exist
  - Run `CREATE EXTENSION IF NOT EXISTS vector;` in your DB before creating tables.
- invalid input syntax for type vector
  - Ensure you pass embeddings as a float slice to pgx (use `pgvector-go`) and not as a string.
- FK constraint fails on `note_chunks.note_id`
  - Insert the parent note in `notes` before inserting chunks (transaction recommended).
- Poor retrieval quality
  - Use a stronger embedding model (e.g., `bge-m3`), ensure the same dimension is used in DB and client, and consider a similarity threshold when selecting top-k.
- Streaming returns only a partial token (e.g., "I")
  - Disable streaming in your LLM client for synchronous tests (set `stream: false`).

## Makefile helpers
```makefile
# Update vector size, copy migration, and run it inside the container
make migrate MODEL_VECTOR_SIZE=768

# Remove a locally pulled Ollama model
make clean-model MODEL=bge-m3
```

## Project Layout
- `cmd/server/` – server bootstrap
- `internal/http/` – HTTP handlers (health, ingest, search, chat)
- `internal/rag/` – ingest/retrieve utilities
- `internal/emb/` – embeddings client
- `internal/llm/` – LLM chat client
- `migrations/` – SQL schema

## License
MIT (or your choice)
