package terminal

import (
	"context"
	"strings"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
)

// EventPublisher handles terminal-specific event publishing
type EventPublisher struct {
	eventBus *events.IEventBus
	logger   *logger.Logger
}

// NewEventPublisher creates a new terminal event publisher
func NewEventPublisher(eventBus events.IEventBus) *EventPublisher {
	return &EventPublisher{
		eventBus: &eventBus,
		logger:   logger.GetDefaultLogger(),
	}
}

// PublishSessionCreated publishes a session created event
func (ep *EventPublisher) PublishSessionCreated(ctx context.Context, session *models.TerminalSession) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: session.ID,
		Shell:     session.Shell,
		CWD:       session.CWD,
		Size:      session.Size,
		CreatedAt: session.CreatedAt,
		Active:    session.Active,
		User:      session.User,
	}

	err := (*ep.eventBus).Publish(ctx, models.TerminalCreated, data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal created event", err, map[string]interface{}{
			"session_id": session.ID,
		})
	}
}

// PublishSessionClosed publishes a session closed event
func (ep *EventPublisher) PublishSessionClosed(ctx context.Context, sessionID string, reason string) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		ClosedAt:  time.Now(),
		Reason:    reason,
	}

	err := (*ep.eventBus).Publish(ctx, models.TerminalClosed, data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal closed event", err, map[string]interface{}{
			"session_id": sessionID,
			"reason":     reason,
		})
	}
}

// PublishSessionResized publishes a session resized event
func (ep *EventPublisher) PublishSessionResized(ctx context.Context, sessionID string, size *models.TerminalSize) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Size:      size,
		ResizedAt: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, models.TerminalResized, data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal resized event", err, map[string]interface{}{
			"session_id": sessionID,
			"rows":       size.Rows,
			"cols":       size.Cols,
		})
	}
}

// PublishOutput publishes terminal output data
func (ep *EventPublisher) PublishOutput(ctx context.Context, sessionID string, data []byte) {
	if ep.eventBus == nil {
		return
	}

	// Split large output into chunks to avoid event size limits
	maxChunkSize := 4096 // 4KB chunks
	for i := 0; i < len(data); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[i:end]
		chunkData := models.TerminalEventData{
			SessionID:   sessionID,
			Data:        chunk,
			DataLength:  len(chunk),
			ChunkIndex:  i / maxChunkSize,
			IsLastChunk: end >= len(data),
			Timestamp:   time.Now(),
		}

		err := (*ep.eventBus).Publish(ctx, models.TerminalOutput, chunkData, "terminal-service")
		if err != nil {
			ep.logger.Error("Failed to publish terminal output event", err, map[string]interface{}{
				"session_id":  sessionID,
				"chunk_index": chunkData.ChunkIndex,
				"data_length": chunkData.DataLength,
			})
		}
	}
}

// PublishInput publishes terminal input data
func (ep *EventPublisher) PublishInput(ctx context.Context, sessionID string, data []byte) {
	if ep.eventBus == nil {
		return
	}

	dataStr := string(data)

	// Don't publish sensitive input like passwords
	if isSensitiveInput(dataStr) {
		ep.logger.Debug("Skipping publication of sensitive input", map[string]interface{}{
			"session_id":  sessionID,
			"data_length": len(data),
		})
		return
	}

	eventData := models.TerminalEventData{
		SessionID:  sessionID,
		Data:       data,
		DataLength: len(data),
		Timestamp:  time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, models.TerminalInput, eventData, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal input event", err, map[string]interface{}{
			"session_id":  sessionID,
			"data_length": len(data),
		})
	}
}

// PublishCommand publishes a terminal command event
func (ep *EventPublisher) PublishCommand(ctx context.Context, command *models.TerminalCommand) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: command.SessionID,
		Command:   command.Command,
		Arguments: command.Arguments,
		CWD:       command.CWD,
		StartTime: command.StartTime,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.command.started", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal command event", err, map[string]interface{}{
			"session_id": command.SessionID,
			"command":    command.Command,
		})
	}
}

// PublishCommandCompleted publishes a command completion event
func (ep *EventPublisher) PublishCommandCompleted(ctx context.Context, command *models.TerminalCommand) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: command.SessionID,
		Command:   command.Command,
		Arguments: command.Arguments,
		CWD:       command.CWD,
		StartTime: command.StartTime,
		EndTime:   command.EndTime,
		ExitCode:  command.ExitCode,
		Duration:  command.Duration,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.command.completed", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal command completed event", err, map[string]interface{}{
			"session_id": command.SessionID,
			"command":    command.Command,
			"exit_code":  command.ExitCode,
		})
	}
}

// PublishBell publishes a terminal bell event
func (ep *EventPublisher) PublishBell(ctx context.Context, sessionID string) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "bell",
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.bell", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal bell event", err, map[string]interface{}{
			"session_id": sessionID,
		})
	}
}

// PublishFocusChanged publishes a terminal focus changed event
func (ep *EventPublisher) PublishFocusChanged(ctx context.Context, sessionID string, focused bool) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "focus_changed",
		Focused:   &focused,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.focus.changed", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal focus changed event", err, map[string]interface{}{
			"session_id": sessionID,
			"focused":    focused,
		})
	}
}

// PublishTitleChanged publishes a terminal title changed event
func (ep *EventPublisher) PublishTitleChanged(ctx context.Context, sessionID, title string) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "title_changed",
		Title:     title,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.title.changed", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal title changed event", err, map[string]interface{}{
			"session_id": sessionID,
			"title":      title,
		})
	}
}

// PublishDirectoryChanged publishes a working directory changed event
func (ep *EventPublisher) PublishDirectoryChanged(ctx context.Context, sessionID, cwd string) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "directory_changed",
		CWD:       cwd,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.directory.changed", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal directory changed event", err, map[string]interface{}{
			"session_id": sessionID,
			"cwd":        cwd,
		})
	}
}

// PublishProcessStarted publishes a process started event
func (ep *EventPublisher) PublishProcessStarted(ctx context.Context, sessionID string, process *models.TerminalProcess) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "process_started",
		PID:       process.PID,
		Process:   process,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.process.started", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal process started event", err, map[string]interface{}{
			"session_id": sessionID,
			"pid":        process.PID,
			"name":       process.Name,
		})
	}
}

// PublishProcessEnded publishes a process ended event
func (ep *EventPublisher) PublishProcessEnded(ctx context.Context, sessionID string, process *models.TerminalProcess) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "process_ended",
		PID:       process.PID,
		Process:   process,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.process.ended", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal process ended event", err, map[string]interface{}{
			"session_id": sessionID,
			"pid":        process.PID,
			"name":       process.Name,
		})
	}
}

// PublishThemeChanged publishes a terminal theme changed event
func (ep *EventPublisher) PublishThemeChanged(ctx context.Context, sessionID string, theme string) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "theme_changed",
		Theme:     theme,
		Timestamp: time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, models.TerminalThemeChanged, data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal theme changed event", err, map[string]interface{}{
			"session_id": sessionID,
			"theme":      theme,
		})
	}
}

// PublishError publishes a terminal error event
func (ep *EventPublisher) PublishError(ctx context.Context, sessionID string, err error, context string) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID: sessionID,
		Type:      "error",
		Error:     err.Error(),
		Context:   context,
		Timestamp: time.Now(),
	}

	publishErr := (*ep.eventBus).Publish(ctx, models.ErrorOccurred, data, "terminal-service")
	if publishErr != nil {
		ep.logger.Error("Failed to publish terminal error event", publishErr, map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

// PublishNotification publishes a terminal notification event
func (ep *EventPublisher) PublishNotification(ctx context.Context, notification *models.TerminalNotification) {
	if ep.eventBus == nil {
		return
	}

	data := models.TerminalEventData{
		SessionID:    notification.SessionID,
		Type:         "notification",
		Notification: notification,
		Timestamp:    time.Now(),
	}

	err := (*ep.eventBus).Publish(ctx, "terminal.notification", data, "terminal-service")
	if err != nil {
		ep.logger.Error("Failed to publish terminal notification event", err, map[string]interface{}{
			"session_id": notification.SessionID,
			"type":       notification.Type,
			"title":      notification.Title,
		})
	}
}

// Helper functions

// isSensitiveInput checks if input contains sensitive information
func isSensitiveInput(input string) bool {
	sensitivePatterns := []string{
		"password",
		"passwd",
		"secret",
		"token",
		"key",
		"auth",
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(inputLower, pattern) {
			return true
		}
	}

	// Check for input that looks like it might be a password (no echo)
	if strings.Contains(inputLower, "askpass") ||
		strings.Contains(inputLower, "getpass") {
		return true
	}

	return false
}

// TerminalEventData represents data for terminal events
type TerminalEventData struct {
	SessionID    string                       `json:"session_id"`
	Type         string                       `json:"type,omitempty"`
	Data         []byte                       `json:"data,omitempty"`
	DataLength   int                          `json:"data_length,omitempty"`
	ChunkIndex   int                          `json:"chunk_index,omitempty"`
	IsLastChunk  bool                         `json:"is_last_chunk,omitempty"`
	Shell        string                       `json:"shell,omitempty"`
	CWD          string                       `json:"cwd,omitempty"`
	Size         *models.TerminalSize         `json:"size,omitempty"`
	Command      string                       `json:"command,omitempty"`
	Arguments    []string                     `json:"arguments,omitempty"`
	StartTime    time.Time                    `json:"start_time,omitempty"`
	EndTime      *time.Time                   `json:"end_time,omitempty"`
	ExitCode     *int                         `json:"exit_code,omitempty"`
	Duration     time.Duration                `json:"duration,omitempty"`
	Process      *models.TerminalProcess      `json:"process,omitempty"`
	PID          int                          `json:"pid,omitempty"`
	Focused      *bool                        `json:"focused,omitempty"`
	Title        string                       `json:"title,omitempty"`
	Theme        string                       `json:"theme,omitempty"`
	Notification *models.TerminalNotification `json:"notification,omitempty"`
	Error        string                       `json:"error,omitempty"`
	Context      string                       `json:"context,omitempty"`
	Reason       string                       `json:"reason,omitempty"`
	Active       bool                         `json:"active,omitempty"`
	User         string                       `json:"user,omitempty"`
	CreatedAt    time.Time                    `json:"created_at,omitempty"`
	ClosedAt     time.Time                    `json:"closed_at,omitempty"`
	ResizedAt    time.Time                    `json:"resized_at,omitempty"`
	Timestamp    time.Time                    `json:"timestamp"`
}
