package handlers

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/armin/geoquiz-api/internal/auth"
	"github.com/gin-gonic/gin"
)

type updateAccountRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Username  *string `json:"username"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UpdateAccount patches the current user's profile fields.
func (h *Handler) UpdateAccount(c *gin.Context) {
	var req updateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request")
		return
	}

	userID := auth.UserIDFromContext(c)
	user, err := h.getUserByID(c, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load user")
		return
	}

	first := user.FirstName
	last := user.LastName
	username := user.Username
	if req.FirstName != nil {
		first = strings.TrimSpace(*req.FirstName)
	}
	if req.LastName != nil {
		last = strings.TrimSpace(*req.LastName)
	}
	if req.Username != nil {
		u := strings.TrimSpace(*req.Username)
		if u == "" {
			username = nil
		} else {
			username = &u
		}
	}

	_, err = h.db.ExecContext(c.Request.Context(), `
		UPDATE users
		SET first_name = $1, last_name = $2, username = $3, updated_at = NOW()
		WHERE id = $4
	`, first, last, nullString(username), userID)
	if err != nil {
		if fmtUniqueViolation(err) {
			writeError(c, http.StatusConflict, "username already taken")
			return
		}
		writeError(c, http.StatusInternalServerError, "could not update account")
		return
	}

	user, err = h.getUserByID(c, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load user")
		return
	}
	c.JSON(http.StatusOK, h.userPublic(user))
}

// UploadAvatar stores a profile picture for the current user.
func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	file, err := c.FormFile("avatar")
	if err != nil {
		writeError(c, http.StatusBadRequest, "avatar file required")
		return
	}
	if file.Size > 2<<20 {
		writeError(c, http.StatusBadRequest, "avatar too large (max 2MB)")
		return
	}

	ext := safeExt(file.Filename)
	rel := avatarRelPath(userID, ext)
	abs := absUpload(h.cfg, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		writeError(c, http.StatusInternalServerError, "could not store avatar")
		return
	}

	src, err := file.Open()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not read upload")
		return
	}
	defer src.Close()

	dst, err := os.Create(abs)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not store avatar")
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		writeError(c, http.StatusInternalServerError, "could not store avatar")
		return
	}

	_, err = h.db.ExecContext(c.Request.Context(), `
		UPDATE users SET avatar_path = $1, updated_at = NOW() WHERE id = $2
	`, rel, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not update avatar")
		return
	}

	user, err := h.getUserByID(c, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load user")
		return
	}
	c.JSON(http.StatusOK, h.userPublic(user))
}

// ChangePassword updates the current user's password.
func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request")
		return
	}

	userID := auth.UserIDFromContext(c)
	user, err := h.getUserByID(c, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load user")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		writeError(c, http.StatusForbidden, "current password is incorrect")
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not hash password")
		return
	}
	_, err = h.db.ExecContext(c.Request.Context(), `
		UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2
	`, hash, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not update password")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeleteAccount removes the current user and related data.
func (h *Handler) DeleteAccount(c *gin.Context) {
	userID := auth.UserIDFromContext(c)
	user, err := h.getUserByID(c, userID)
	if err != nil && err != sql.ErrNoRows {
		writeError(c, http.StatusInternalServerError, "could not load user")
		return
	}
	if user.AvatarPath != nil {
		_ = os.Remove(absUpload(h.cfg, *user.AvatarPath))
	}
	_, err = h.db.ExecContext(c.Request.Context(), `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not delete account")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetProfile returns a public profile by username.
func (h *Handler) GetProfile(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	if username == "" {
		writeError(c, http.StatusBadRequest, "username required")
		return
	}

	user, err := scanUser(h.db.QueryRowContext(c.Request.Context(), userSelect+` WHERE username = $1`, username))
	if err == sql.ErrNoRows {
		writeError(c, http.StatusNotFound, "profile not found")
		return
	}
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load profile")
		return
	}
	if user.Username == nil {
		writeError(c, http.StatusNotFound, "profile not found")
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT id, user_id, quiz_type, correct, total, created_at
		FROM scores
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, user.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load scores")
		return
	}
	defer rows.Close()

	scores := make([]gin.H, 0)
	for rows.Next() {
		var id, uid, quizType string
		var correct, total int
		var createdAt interface{}
		if err := rows.Scan(&id, &uid, &quizType, &correct, &total, &createdAt); err != nil {
			writeError(c, http.StatusInternalServerError, "could not load scores")
			return
		}
		scores = append(scores, gin.H{
			"id":         id,
			"user_id":    uid,
			"quiz_type":  quizType,
			"correct":    correct,
			"total":      total,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"username":   *user.Username,
		"avatar_url": h.avatarURL(user.AvatarPath),
		"scores":     scores,
	})
}

// ListScores returns recent high scores for the score-board (public).
func (h *Handler) ListScores(c *gin.Context) {
	quizType := strings.TrimSpace(c.Query("quiz_type"))
	if quizType != "" && quizType != "flag" && quizType != "map" {
		writeError(c, http.StatusBadRequest, "quiz_type must be flag or map")
		return
	}

	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			writeError(c, http.StatusBadRequest, "invalid limit")
			return
		}
		if n > 100 {
			n = 100
		}
		limit = n
	}

	query := `
		SELECT s.id, u.username, s.quiz_type, s.correct, s.total, s.created_at
		FROM scores s
		INNER JOIN users u ON u.id = s.user_id
		WHERE u.username IS NOT NULL
	`
	args := make([]interface{}, 0, 2)
	if quizType != "" {
		query += ` AND s.quiz_type = $1`
		args = append(args, quizType)
		query += ` ORDER BY s.correct DESC, s.created_at DESC LIMIT $2`
		args = append(args, limit)
	} else {
		query += ` ORDER BY s.correct DESC, s.created_at DESC LIMIT $1`
		args = append(args, limit)
	}

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load scores")
		return
	}
	defer rows.Close()

	scores := make([]gin.H, 0)
	for rows.Next() {
		var id, username, qt string
		var correct, total int
		var createdAt interface{}
		if err := rows.Scan(&id, &username, &qt, &correct, &total, &createdAt); err != nil {
			writeError(c, http.StatusInternalServerError, "could not load scores")
			return
		}
		scores = append(scores, gin.H{
			"id":         id,
			"username":   username,
			"quiz_type":  qt,
			"correct":    correct,
			"total":      total,
			"created_at": createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		writeError(c, http.StatusInternalServerError, "could not load scores")
		return
	}

	c.JSON(http.StatusOK, scores)
}

// CreateScore stores a quiz score for the current user.
func (h *Handler) CreateScore(c *gin.Context) {
	var req struct {
		QuizType string `json:"quiz_type" binding:"required"`
		Correct  int    `json:"correct"`
		Total    int    `json:"total"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request")
		return
	}
	if req.QuizType != "flag" && req.QuizType != "map" {
		writeError(c, http.StatusBadRequest, "quiz_type must be flag or map")
		return
	}
	if req.Correct < 0 || req.Total <= 0 || req.Correct > req.Total {
		writeError(c, http.StatusBadRequest, "invalid score values")
		return
	}

	userID := auth.UserIDFromContext(c)
	var id string
	err := h.db.QueryRowContext(c.Request.Context(), `
		INSERT INTO scores (user_id, quiz_type, correct, total)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, req.QuizType, req.Correct, req.Total).Scan(&id)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not save score")
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":        id,
		"quiz_type": req.QuizType,
		"correct":   req.Correct,
		"total":     req.Total,
	})
}
