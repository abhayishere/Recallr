CREATE EXTENSION IF NOT EXISTS vector;
DROP TABLE IF EXISTS note_chunks CASCADE;
DROP TABLE IF EXISTS notes CASCADE;
DROP TABLE IF EXISTS chats CASCADE;
DROP TABLE IF EXISTS messages CASCADE;
-- Original note file
CREATE TABLE notes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id  TEXT,                    -- Apple Notes id if you have it
  title        TEXT NOT NULL,
  raw_text     TEXT NOT NULL,
  source_path  TEXT,                    -- where it came from on disk
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Chunked passages
-- Adjust 768 if your embedding model uses a different dimension
CREATE TABLE note_chunks (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  note_id      UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  idx          INT  NOT NULL,
  content      TEXT NOT NULL,
  embedding    VECTOR(1024) NOT NULL
);

CREATE INDEX ON note_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX ON note_chunks (note_id, idx);

CREATE TABLE chats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('user','assistant','system')),
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
