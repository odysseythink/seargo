package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type sqliteBackend struct {
	db          *sql.DB
	maxValueLen int
	closeCh     chan struct{}
	closeOnce   sync.Once
}

func newSQLiteBackend(opts Options) (*sqliteBackend, error) {
	path := opts.SQLitePath
	if path == "" {
		path = "data/seargo.db"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite backend: open: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite backend: ping: %w", err)
	}
	// Enable WAL
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite backend: wal: %w", err)
	}
	// Create table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
			key        TEXT PRIMARY KEY,
			value      BLOB  NOT NULL,
			expires_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_kv_expires ON kv(expires_at);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite backend: schema: %w", err)
	}

	backend := &sqliteBackend{
		db:          db,
		maxValueLen: opts.MaxValueLen,
		closeCh:     make(chan struct{}),
	}

	// Run initial maintenance
	backend.maintenance()

	// Start maintenance goroutine
	interval := opts.Maintenance
	if interval <= 0 {
		interval = time.Hour
	}
	go backend.maintenanceLoop(interval)

	return backend, nil
}

func (s *sqliteBackend) Get(ctx context.Context, key string) ([]byte, bool, error) {
	var value []byte
	var expiresAt int64
	err := s.db.QueryRowContext(ctx,
		"SELECT value, expires_at FROM kv WHERE key = ?", key,
	).Scan(&value, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if expiresAt > 0 && time.Now().Unix() >= expiresAt {
		return nil, false, nil
	}
	return value, true, nil
}

func (s *sqliteBackend) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s.maxValueLen > 0 && len(value) > s.maxValueLen {
		return fmt.Errorf("value size %d exceeds max %d", len(value), s.maxValueLen)
	}
	expiresAt := int64(0)
	if ttl > 0 {
		expiresAt = time.Now().Unix() + int64(ttl.Seconds())
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO kv (key, value, expires_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value,
		                               expires_at=excluded.expires_at
	`, key, value, expiresAt)
	return err
}

func (s *sqliteBackend) Delete(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM kv WHERE key = ?", key)
	return err
}

func (s *sqliteBackend) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	_, ok, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	return true, s.Set(ctx, key, value, ttl)
}

func (s *sqliteBackend) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var value []byte
	var expiresAt int64
	err = tx.QueryRowContext(ctx,
		"SELECT value, expires_at FROM kv WHERE key = ?", key,
	).Scan(&value, &expiresAt)

	now := time.Now().Unix()
	var newVal int64
	var newExpires int64

	if ttl > 0 {
		newExpires = now + int64(ttl.Seconds())
	}

	if err == sql.ErrNoRows {
		newVal = 1
	} else if err != nil {
		return 0, err
	} else {
		old, parseErr := strconv.ParseInt(string(value), 10, 64)
		if parseErr != nil {
			old = 0
		}
		newVal = old + 1
		if expiresAt > 0 && expiresAt <= now {
			// expired — reset
			newVal = 1
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO kv (key, value, expires_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value,
		                               expires_at=excluded.expires_at
	`, key, []byte(strconv.FormatInt(newVal, 10)), newExpires)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newVal, nil
}

func (s *sqliteBackend) Expire(ctx context.Context, key string, ttl time.Duration) error {
	expiresAt := int64(0)
	if ttl > 0 {
		expiresAt = time.Now().Unix() + int64(ttl.Seconds())
	}
	_, err := s.db.ExecContext(ctx,
		"UPDATE kv SET expires_at = ? WHERE key = ?", expiresAt, key,
	)
	return err
}

func (s *sqliteBackend) Close() error {
	var err error
	s.closeOnce.Do(func() {
		close(s.closeCh)
		err = s.db.Close()
	})
	return err
}

func (s *sqliteBackend) WithNamespace(namespace string) KV {
	panic("not implemented: use through New()")
}

func (s *sqliteBackend) BackendName() string {
	return "sqlite"
}

func (s *sqliteBackend) maintenance() {
	now := time.Now().Unix()
	s.db.Exec("DELETE FROM kv WHERE expires_at > 0 AND expires_at <= ?", now)
	s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
}

func (s *sqliteBackend) maintenanceLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.closeCh:
			return
		case <-ticker.C:
			s.maintenance()
		}
	}
}
