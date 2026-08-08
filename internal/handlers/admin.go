package handlers

import (
	"net/http"
	"strings"

	appdb "github.com/armin/geoquiz-api/internal/db"
	"github.com/gin-gonic/gin"
)

// GetInviteCode returns the current invite code (admin only).
func (h *Handler) GetInviteCode(c *gin.Context) {
	code, err := appdb.GetSetting(c.Request.Context(), h.db, "invite_code")
	if err != nil {
		writeError(c, http.StatusInternalServerError, "could not load invite code")
		return
	}
	c.JSON(http.StatusOK, gin.H{"invite_code": code})
}

// UpdateInviteCode sets the invite code (admin only).
func (h *Handler) UpdateInviteCode(c *gin.Context) {
	var req struct {
		InviteCode string `json:"invite_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request")
		return
	}
	code := strings.TrimSpace(req.InviteCode)
	if code == "" {
		writeError(c, http.StatusBadRequest, "invite_code required")
		return
	}
	if err := appdb.SetSetting(c.Request.Context(), h.db, "invite_code", code); err != nil {
		writeError(c, http.StatusInternalServerError, "could not update invite code")
		return
	}
	c.JSON(http.StatusOK, gin.H{"invite_code": code})
}
