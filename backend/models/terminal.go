package models

import "time"

// TerminalSession represents an active terminal session
type TerminalSession struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Shell       string                 `json:"shell"`
	WorkingDir  string                 `json:"working_dir"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	CreatedAt   time.Time              `json:"created_at"`
	LastActive  time.Time              `json:"last_active"`
	IsActive    bool                   `json:"is_active"`
	Environment map[string]string      `json:"environment"`
	Properties  map[string]interface{} `json:"properties"`
}

// TerminalCommand represents a command execution in terminal
type TerminalCommand struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Command   string    `json:"command"`
	Arguments []string  `json:"arguments"`
	ExitCode  int       `json:"exit_code"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int64     `json:"duration"`
	Output    string    `json:"output"`
	Error     string    `json:"error"`
}

// TerminalConfig represents terminal configuration
type TerminalConfig struct {
	DefaultShell    string            `json:"default_shell"`
	DefaultFont     string            `json:"default_font"`
	DefaultFontSize int               `json:"default_font_size"`
	Colors          map[string]string `json:"colors"`
	Scrollback      int               `json:"scrollback"`
	BellEnabled     bool              `json:"bell_enabled"`
	CursorBlink     bool              `json:"cursor_blink"`
	Opacity         int               `json:"opacity"`
	Hotkeys         map[string]string `json:"hotkeys"`
}

// TerminalProfile represents a terminal profile
type TerminalProfile struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Shell       string            `json:"shell"`
	Environment map[string]string `json:"environment"`
	WorkingDir  string            `json:"working_dir"`
	Args        []string          `json:"args"`
	IsDefault   bool              `json:"is_default"`
}

// TerminalEvent represents terminal-related events
type TerminalEvent struct {
	Type      string      `json:"type"`
	SessionID string      `json:"session_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// BufferLine represents a line in terminal buffer
type BufferLine struct {
	Text      string    `json:"text"`
	Attributes []Attribute `json:"attributes"`
	Timestamp time.Time `json:"timestamp"`
}

// Attribute represents text formatting attributes
type Attribute struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// TerminalHistory represents command history
type TerminalHistory struct {
	SessionID string      `json:"session_id"`
	Commands  []string    `json:"commands"`
	Index     int         `json:"index"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  interface{} `json:"metadata"`
}
