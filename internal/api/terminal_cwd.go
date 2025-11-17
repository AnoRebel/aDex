package api

import (
	"context"
	"net/http"

	"aDex-UI/internal/services/terminal"

	"github.com/labstack/echo/v4"
)

// TerminalCWDHandler handles terminal CWD tracking API endpoints
type TerminalCWDHandler struct {
	terminalService *terminal.Service
}

// NewTerminalCWDHandler creates a new terminal CWD handler
func NewTerminalCWDHandler(terminalService *terminal.Service) *TerminalCWDHandler {
	return &TerminalCWDHandler{
		terminalService: terminalService,
	}
}

// GetCurrentCWD handles GET /api/terminal/:sessionId/cwd
func (h *TerminalCWDHandler) GetCurrentCWD(c echo.Context) error {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Session ID is required",
		})
	}

	// Get current working directory
	cwd, err := h.terminalService.GetCurrentCWD(sessionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "Session not found",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"session_id": sessionID,
		"cwd":        cwd,
	})
}

// GetCWDStats handles GET /api/terminal/cwd/stats
func (h *TerminalCWDHandler) GetCWDStats(c echo.Context) error {
	stats := h.terminalService.GetCWDTrackerStats()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"stats": stats,
	})
}

// RegisterTerminalCWDRoutes registers CWD tracking routes
func (h *TerminalCWDHandler) RegisterRoutes(e *echo.Echo) {
	// CWD tracking routes
	api := e.Group("/api")
	{
		// General CWD stats
		api.GET("/terminal/cwd/stats", h.GetCWDStats)

		// Session-specific CWD endpoints
		api.GET("/terminal/session/:sessionId/cwd", h.GetCurrentCWD)
	}
}