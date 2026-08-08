package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotEnv loads KEY=VALUE pairs from a .env file into the process environment
// when the key is not already set. Missing file is ignored.
func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

// Config holds runtime settings for the API.
type Config struct {
	Addr          string
	DatabaseURL   string
	JWTSecret     string
	InviteCode    string
	AdminEmail    string
	UploadDir     string
	MigrationsDir string
	GeoJSONPath   string
	CORSOrigins   []string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		Addr:          envOr("ADDR", ":8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     envOr("JWT_SECRET", "dev-jwt-secret-change-me"),
		InviteCode:    envOr("INVITE_CODE", "armin"),
		AdminEmail:    strings.ToLower(strings.TrimSpace(envOr("ADMIN_EMAIL", "armin@geoquiz.local"))),
		UploadDir:     envOr("UPLOAD_DIR", "uploads"),
		MigrationsDir: envOr("MIGRATIONS_DIR", "migrations"),
		GeoJSONPath:   envOr("COUNTRIES_GEOJSON", "data/countries.geojson"),
		CORSOrigins:   splitCSV(envOr("CORS_ORIGINS", defaultCORSOrigins)),
	}
}

const defaultCORSOrigins = "http://localhost:5051,http://127.0.0.1:5051,http://localhost:5173,http://127.0.0.1:5173,http://localhost:5174,http://127.0.0.1:5174,http://localhost:5175,http://127.0.0.1:5175,http://localhost:8080,http://127.0.0.1:8080,http://localhost,http://127.0.0.1,https://geo-quiz.xaigrok.ir"

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
