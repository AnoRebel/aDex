package tests

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"

	"aDex/internal/services/terminal"
)

// TestTerminal_JSONMarshalable proves that the values returned to Wails
// from CreateTerminal / GetTerminalInfo are JSON-encodable. Wails serializes
// every bound method's return value through encoding/json and crashes the
// app at launch if it hits an unsupported type:
//
//   FAT | json: unsupported type: func() error
//
// Pre-fix, *Terminal contained `Command *exec.Cmd`, `PTY *os.File`,
// channels, and sync primitives — none of which JSON-encode. The fix in
// backend/services/terminal/service.go tags those fields `json:"-"`. This
// test guards the regression.
func TestTerminal_JSONMarshalable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY init differs on windows; tests for marshalability run on linux/darwin")
	}

	s := terminal.NewService()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	term, err := s.CreateTerminal(ctx, 80, 24)
	if err != nil {
		t.Skipf("CreateTerminal not available in this env: %v", err)
	}
	t.Cleanup(func() { _ = s.CloseTerminal(ctx, term.ID) })

	if _, err := json.Marshal(term); err != nil {
		t.Fatalf("Terminal is not JSON-marshalable: %v", err)
	}

	info, err := s.GetTerminalInfo(ctx, term.ID)
	if err != nil {
		t.Fatalf("GetTerminalInfo: %v", err)
	}
	if _, err := json.Marshal(info); err != nil {
		t.Fatalf("GetTerminalInfo result is not JSON-marshalable: %v", err)
	}

	stats := s.GetCWDStats()
	if _, err := json.Marshal(stats); err != nil {
		t.Fatalf("CWDStats is not JSON-marshalable: %v", err)
	}
}
