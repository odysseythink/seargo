package engine

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// EngineCache provides a per-engine key/value store backed by SQLite,
// ported from SearXNG's EngineCache.
type EngineCache struct {
	mu sync.RWMutex
	db *sql.DB
}

// NewEngineCache opens or creates an SQLite database at the given path.
func NewEngineCache(path string) (*EngineCache, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open engine cache: %w", err)
	}

	db.SetMaxOpenConns(10)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping engine cache: %w", err)
	}

	if err := createTable(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("create cache table: %w", err)
	}

	return &EngineCache{db: db}, nil
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS engine_cache (
			engine_name TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			expires_at INTEGER NOT NULL,
			PRIMARY KEY (engine_name, key)
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_expires ON engine_cache(expires_at)`)
	return err
}

// Set stores a value with a TTL in seconds.
func (c *EngineCache) Set(engineName, key, value string, ttl int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Now().Unix() + ttl
	if ttl <= 0 {
		expiresAt = 0
	}

	_, err := c.db.Exec(`
		INSERT OR REPLACE INTO engine_cache (engine_name, key, value, expires_at)
		VALUES (?, ?, ?, ?)
	`, engineName, key, value, expiresAt)
	return err
}

// Get retrieves a value. Returns (value, true) if found and not expired.
func (c *EngineCache) Get(engineName, key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var value string
	var expiresAt int64
	err := c.db.QueryRow(
		`SELECT value, expires_at FROM engine_cache WHERE engine_name = ? AND key = ?`,
		engineName, key,
	).Scan(&value, &expiresAt)

	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		return "", false
	}

	if expiresAt == 0 || time.Now().Unix() >= expiresAt {
		return "", false
	}

	return value, true
}

// Delete removes a key for an engine.
func (c *EngineCache) Delete(engineName, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(
		`DELETE FROM engine_cache WHERE engine_name = ? AND key = ?`,
		engineName, key,
	)
	return err
}

// Close closes the database connection.
func (c *EngineCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Close()
}

// PurgeExpired removes all expired entries.
func (c *EngineCache) PurgeExpired() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.db.Exec(`DELETE FROM engine_cache WHERE expires_at > 0 AND expires_at <= ?`, time.Now().Unix())
	return err
}
