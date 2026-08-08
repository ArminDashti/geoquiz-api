package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/armin/geoquiz-api/internal/config"
	"github.com/armin/geoquiz-api/internal/data"
	"github.com/armin/geoquiz-api/internal/models"
	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP endpoints.
type Handler struct {
	store  *data.Store
	db     *sql.DB
	cfg    config.Config
}

// New creates a Handler.
func New(store *data.Store, db *sql.DB, cfg config.Config) *Handler {
	return &Handler{store: store, db: db, cfg: cfg}
}

// Health returns a simple liveness payload.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListCountries returns id/name/iso for all countries.
func (h *Handler) ListCountries(c *gin.Context) {
	c.JSON(http.StatusOK, h.store.List())
}

// CountriesGeoJSON returns the MapLibre-ready FeatureCollection.
func (h *Handler) CountriesGeoJSON(c *gin.Context) {
	c.Data(http.StatusOK, "application/geo+json", h.store.GeoJSON())
}

// CountryNeighbors returns land-adjacent countries for the given id.
func (h *Handler) CountryNeighbors(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country id"})
		return
	}

	neighbors, ok := h.store.Neighbors(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "country not found"})
		return
	}

	c.JSON(http.StatusOK, neighbors)
}

func (h *Handler) avatarURL(avatarPath *string) *string {
	if avatarPath == nil || *avatarPath == "" {
		return nil
	}
	url := "/uploads/" + filepath.ToSlash(*avatarPath)
	return &url
}

func (h *Handler) userPublic(u models.User) models.User {
	u.AvatarURL = h.avatarURL(u.AvatarPath)
	u.AvatarPath = nil
	return u
}

func scanUser(row interface {
	Scan(dest ...any) error
}) (models.User, error) {
	var u models.User
	var username, avatar sql.NullString
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName,
		&username, &avatar, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}
	if username.Valid {
		u.Username = &username.String
	}
	if avatar.Valid {
		u.AvatarPath = &avatar.String
	}
	return u, nil
}

const userSelect = `
	SELECT id, email, password_hash, first_name, last_name, username, avatar_path, is_admin, created_at, updated_at
	FROM users
`

func (h *Handler) getUserByID(c *gin.Context, id string) (models.User, error) {
	return scanUser(h.db.QueryRowContext(c.Request.Context(), userSelect+` WHERE id = $1`, id))
}

func (h *Handler) getUserByEmail(c *gin.Context, email string) (models.User, error) {
	return scanUser(h.db.QueryRowContext(c.Request.Context(), userSelect+` WHERE email = $1`, strings.ToLower(email)))
}

func (h *Handler) getUserByUsername(c *gin.Context, username string) (models.User, error) {
	return scanUser(h.db.QueryRowContext(c.Request.Context(), userSelect+` WHERE lower(username) = lower($1)`, strings.TrimSpace(username)))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isAdminEmail(cfgAdmin, email string) bool {
	if cfgAdmin == "" {
		return false
	}
	return normalizeEmail(cfgAdmin) == normalizeEmail(email)
}

func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func nullString(p *string) sql.NullString {
	if p == nil || *p == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

func fmtUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}

func absUpload(cfg config.Config, rel string) string {
	return filepath.Join(cfg.UploadDir, rel)
}

func safeExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return ext
	default:
		return ".jpg"
	}
}

func avatarRelPath(userID, ext string) string {
	return filepath.ToSlash(filepath.Join("avatars", fmt.Sprintf("%s%s", userID, ext)))
}
