package models

import (
	"fmt"
	"time"
)

// TerminalSession represents a terminal session state
type TerminalSession struct {
	ID         string                 `json:"id"`
	PID        int                    `json:"pid,omitempty"`
	Shell      string                 `json:"shell"`
	CWD        string                 `json:"cwd"`
	Env        map[string]string      `json:"env"`
	Size       *TerminalSize          `json:"size"`
	Active     bool                   `json:"active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	LastSeen   time.Time              `json:"last_seen"`
	User       string                 `json:"user"`
	Title      string                 `json:"title,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TerminalSize represents terminal dimensions
type TerminalSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// TerminalOptions represents options for creating a terminal session
type TerminalOptions struct {
	Shell      string            `json:"shell,omitempty"`
	CWD        string            `json:"cwd,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Rows       uint16            `json:"rows,omitempty"`
	Cols       uint16            `json:"cols,omitempty"`
	Profile    string            `json:"profile,omitempty"`
	User       string            `json:"user,omitempty"`
	WorkingDir string            `json:"working_dir,omitempty"`
}

// TerminalData represents data sent to or from a terminal
type TerminalData struct {
	SessionID string    `json:"session_id"`
	Data      []byte    `json:"data"`
	Type      string    `json:"type"` // "input", "output", "error", "resize"
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // "user", "process", "system"
}

// TerminalCommand represents a command executed in a terminal
type TerminalCommand struct {
	SessionID   string        `json:"session_id"`
	Command     string        `json:"command"`
	CWD         string        `json:"cwd"`
	Arguments   []string      `json:"arguments,omitempty"`
	Environment []string      `json:"environment,omitempty"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	ExitCode    *int          `json:"exit_code,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	User        string        `json:"user,omitempty"`
	Title       string        `json:"title,omitempty"`
}

// TerminalHistory represents command history for a session
type TerminalHistory struct {
	SessionID string            `json:"session_id"`
	Commands  []TerminalCommand `json:"commands"`
	MaxSize   int               `json:"max_size"`
	Current   int               `json:"current"`
}

// TerminalTheme represents terminal color scheme and appearance
type TerminalTheme struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Colors      TerminalColors    `json:"colors"`
	Font        TerminalFont      `json:"font"`
	Opacity     float64           `json:"opacity"`
	Blur        bool              `json:"blur"`
	CursorStyle string            `json:"cursor_style"`
	Animations  bool              `json:"animations"`
}

// TerminalColors represents terminal color palette
type TerminalColors struct {
	Background   string `json:"background"`
	Foreground   string `json:"foreground"`
	Cursor       string `json:"cursor"`
	Selection    string `json:"selection"`
	Black        string `json:"black"`
	Red          string `json:"red"`
	Green        string `json:"green"`
	Yellow       string `json:"yellow"`
	Blue         string `json:"blue"`
	Magenta      string `json:"magenta"`
	Cyan         string `json:"cyan"`
	White        string `json:"white"`
	BrightBlack  string `json:"bright_black"`
	BrightRed    string `json:"bright_red"`
	BrightGreen  string `json:"bright_green"`
	BrightYellow string `json:"bright_yellow"`
	BrightBlue   string `json:"bright_blue"`
	BrightMagenta string `json:"bright_magenta"`
	BrightCyan   string `json:"bright_cyan"`
	BrightWhite  string `json:"bright_white"`
}

// TerminalFont represents terminal font configuration
type TerminalFont struct {
	Family   string `json:"family"`
	Size     int    `json:"size"`
	Weight   string `json:"weight"`
	LineHeight float64 `json:"line_height"`
	Ligatures bool   `json:"ligatures"`
}

// TerminalTab represents a terminal tab in the UI
type TerminalTab struct {
	ID          string          `json:"id"`
	SessionID   string          `json:"session_id"`
	Title       string          `json:"title"`
	Active      bool            `json:"active"`
	Icon        string          `json:"icon,omitempty"`
	Color       string          `json:"color,omitempty"`
	Badge       string          `json:"badge,omitempty"`
	Position    int             `json:"position"`
	Pinned      bool            `json:"pinned"`
	Modified    bool            `json:"modified"`
	LastActivity time.Time      `json:"last_activity"`
}

// TerminalNotification represents terminal-related notifications
type TerminalNotification struct {
	SessionID string    `json:"session_id"`
	Type      string    `json:"type"` // "bell", "activity", "error", "completion"
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
}

// TerminalProcess represents a process running in a terminal
type TerminalProcess struct {
	PID         int                    `json:"pid"`
	PPID        int                    `json:"ppid"`
	Name        string                 `json:"name"`
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	Env         []string               `json:"env"`
	CWD         string                 `json:"cwd"`
	User        string                 `json:"user"`
	Group       string                 `json:"group"`
	Status      string                 `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	CPUTime     time.Duration          `json:"cpu_time"`
	Memory      uint64                 `json:"memory"`
	TTY         string                 `json:"tty"`
	SessionID   string                 `json:"session_id"`
	Parent      *TerminalProcess       `json:"parent,omitempty"`
	Children    []*TerminalProcess     `json:"children,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TerminalBuffer represents the terminal output buffer
type TerminalBuffer struct {
	Lines      []string `json:"lines"`
	CursorRow  int      `json:"cursor_row"`
	CursorCol  int      `json:"cursor_col"`
	Scrollback int      `json:"scrollback"`
	MaxLines   int      `json:"max_lines"`
	Modified   bool     `json:"modified"`
}

// TerminalClipboard represents clipboard operations
type TerminalClipboard struct {
	Content   string    `json:"content"`
	Type      string    `json:"type"` // "text", "html", "image"
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // "selection", "copy", "paste"
}

// TerminalSearch represents search functionality in terminal
type TerminalSearch struct {
	Query     string    `json:"query"`
	Results   []int     `json:"results"`
	Current   int       `json:"current"`
	CaseSensitive bool  `json:"case_sensitive"`
	Regex     bool      `json:"regex"`
	Direction string    `json:"direction"` // "forward", "backward"
	Timestamp time.Time `json:"timestamp"`
}

// TerminalBookmark represents a bookmark in terminal output
type TerminalBookmark struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Line      int       `json:"line"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	Note      string    `json:"note,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// TerminalProfile represents saved terminal configuration profiles
type TerminalProfile struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Shell       string                 `json:"shell"`
	Env         map[string]string      `json:"env,omitempty"`
	WorkingDir  string                 `json:"working_dir,omitempty"`
	Options     *TerminalOptions       `json:"options,omitempty"`
	Theme       string                 `json:"theme,omitempty"`
	Shortcuts   map[string]string      `json:"shortcuts,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsDefault   bool                   `json:"is_default"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewTerminalSession creates a new terminal session
func NewTerminalSession(options *TerminalOptions) *TerminalSession {
	now := time.Now()

	// Set default shell if not provided
	shell := options.Shell
	if shell == "" {
		shell = getDefaultShell()
	}

	// Set default CWD if not provided
	cwd := options.CWD
	if cwd == "" {
		cwd = "/tmp"
	}

	// Set default size if not provided
	rows := options.Rows
	if rows == 0 {
		rows = 24
	}

	cols := options.Cols
	if cols == 0 {
		cols = 80
	}

	return &TerminalSession{
		ID:        generateTerminalSessionID(),
		Shell:     shell,
		CWD:       cwd,
		Env:       options.Env,
		Size: &TerminalSize{
			Rows: rows,
			Cols: cols,
		},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		LastSeen:  now,
		User:      options.User,
		Metadata:  make(map[string]interface{}),
	}
}

// NewTerminalTab creates a new terminal tab
func NewTerminalTab(sessionID, title string, position int) *TerminalTab {
	return &TerminalTab{
		ID:           generateTabID(),
		SessionID:    sessionID,
		Title:        title,
		Active:       false,
		Position:     position,
		Pinned:       false,
		Modified:     false,
		LastActivity: time.Now(),
	}
}

// NewTerminalCommand creates a new terminal command record
func NewTerminalCommand(sessionID, command, cwd string, args []string) *TerminalCommand {
	return &TerminalCommand{
		SessionID: sessionID,
		Command:   command,
		CWD:       cwd,
		Arguments: args,
		StartTime: time.Now(),
	}
}

// Helper functions

func generateTerminalSessionID() string {
	return fmt.Sprintf("term-%d", time.Now().UnixNano())
}

func generateTabID() string {
	return fmt.Sprintf("tab-%d", time.Now().UnixNano())
}

func getDefaultShell() string {
	// This would be implemented to detect the default shell on the current platform
	// For now, return a common default
	return "/bin/bash"
}