package graph

import (
	"database/sql"
	"fmt"
)

// Schema returns SQL schema for graph database
const Schema = `
CREATE TABLE IF NOT EXISTS nodes (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	type TEXT NOT NULL,
	content TEXT,
	embedding BLOB,
	metadata TEXT,
	file_path TEXT,
	line_number INTEGER,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_nodes_name ON nodes(name);
CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
CREATE INDEX IF NOT EXISTS idx_nodes_file ON nodes(file_path);

CREATE TABLE IF NOT EXISTS edges (
	id TEXT PRIMARY KEY,
	source_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	target_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	relation_type TEXT NOT NULL,
	weight REAL,
	metadata TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source_id);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target_id);
CREATE INDEX IF NOT EXISTS idx_edges_type ON edges(relation_type);

CREATE TABLE IF NOT EXISTS node_embeddings (
	node_id TEXT PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
	embedding BLOB NOT NULL,
	dimension INTEGER,
	model TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

// InitSchema initializes the graph database schema
func InitSchema(db *sql.DB) error {
	if _, err := db.Exec(Schema); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	return nil
}
