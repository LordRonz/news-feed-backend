CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS authors (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS feeds (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  author_id UUID REFERENCES authors(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reactions (
  feed_id UUID PRIMARY KEY REFERENCES feeds(id) ON DELETE CASCADE,
  likes INT DEFAULT 0,
  haha INT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_feeds_created_at ON feeds (created_at DESC);

CREATE TABLE IF NOT EXISTS feed_events (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  feed_id UUID NOT NULL,
  author_id UUID NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);