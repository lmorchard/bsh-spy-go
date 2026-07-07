package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS tracks (
	spotify_id TEXT PRIMARY KEY,
	artist     TEXT,
	title      TEXT,
	added_at   TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS mystery_songs (
	id      INTEGER PRIMARY KEY AUTOINCREMENT,
	artist  TEXT,
	title   TEXT,
	query   TEXT,
	seen_at TEXT NOT NULL DEFAULT (datetime('now'))
);`

// Track is a de-duplicated playlist entry.
type Track struct {
	SpotifyID string
	Artist    string
	Title     string
}

// Store wraps the SQLite cache database.
type Store struct {
	db *sql.DB
}

// Open opens (and if needed creates) the SQLite database at path and ensures schema.
func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Has reports whether spotifyID is already in the cache.
func (s *Store) Has(spotifyID string) (bool, error) {
	var one int
	err := s.db.QueryRow(`SELECT 1 FROM tracks WHERE spotify_id = ?`, spotifyID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Add records a track as seen. Idempotent on spotify_id.
func (s *Store) Add(t Track) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO tracks (spotify_id, artist, title) VALUES (?, ?, ?)`,
		t.SpotifyID, t.Artist, t.Title,
	)
	return err
}

// RecordMystery logs a song that produced no Spotify match.
func (s *Store) RecordMystery(artist, title, query string) error {
	_, err := s.db.Exec(
		`INSERT INTO mystery_songs (artist, title, query) VALUES (?, ?, ?)`,
		artist, title, query,
	)
	return err
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }
