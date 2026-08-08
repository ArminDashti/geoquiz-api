package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/armin/geoquiz-api/internal/auth"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open connects to PostgreSQL.
func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// Migrate runs SQL files in migrationsDir in lexical order.
func Migrate(db *sql.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, name := range files {
		body, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

// SeedInviteCode upserts the invite code from configuration.
func SeedInviteCode(ctx context.Context, db *sql.DB, code string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES ('invite_code', $1)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, code)
	return err
}

// SeedBootstrapUser creates the default admin user when missing (by username).
// If the username already exists, password/admin flags are synced so local seed
// credential corrections take effect without a manual DB reset.
func SeedBootstrapUser(ctx context.Context, db *sql.DB, username, email, password, firstName, lastName string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	var exists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE lower(username) = lower($1))
	`, username).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		_, err = db.ExecContext(ctx, `
			UPDATE users
			SET password_hash = $1, is_admin = TRUE, updated_at = NOW()
			WHERE lower(username) = lower($2)
		`, hash, username)
		return err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, username, is_admin)
		VALUES ($1, $2, $3, $4, $5, TRUE)
		ON CONFLICT (email) DO NOTHING
	`, strings.ToLower(strings.TrimSpace(email)), hash, firstName, lastName, username)
	return err
}

// GetSetting returns a settings value.
func GetSetting(ctx context.Context, db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&value)
	return value, err
}

// SetSetting upserts a settings value.
func SetSetting(ctx context.Context, db *sql.DB, key, value string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	return err
}

