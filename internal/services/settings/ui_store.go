package settings

import (
	"aDex-UI/internal/appdir"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
)

// uiSettingsFileName is the on-disk source of truth for the frontend's
// `adex-settings`. It lives alongside the legacy settings.json but is a
// SEPARATE file with a SEPARATE shape on purpose (see models.UISettings).
const uiSettingsFileName = "adex-ui-settings.json"

// UIStore persists models.UISettings to a single JSON file with an
// atomic write (.tmp + rename) and a one-deep .bak so a corrupt write
// or bad payload can't brick the user's config. It is the persistence
// authority; the frontend localStorage is only a reactive cache that
// is seeded from / flushed to this store.
type UIStore struct {
	logger *logger.Logger
	file   string

	mu       sync.RWMutex
	settings *models.UISettings
}

// NewUIStore opens (or creates) the UI settings file under the user
// config dir. On first run it writes DefaultUISettings(). A corrupt
// file is recovered from .bak, and failing that from defaults — load
// never returns an error that would block app start.
func NewUIStore(log *logger.Logger) (*UIStore, error) {
	appConfigDir := appdir.Config()
	if err := os.MkdirAll(appConfigDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	s := &UIStore{
		logger: log,
		file:   filepath.Join(appConfigDir, uiSettingsFileName),
	}
	s.settings = s.load()
	return s, nil
}

// Get returns a value copy of the current settings (no shared mutable
// state escapes the store).
func (s *UIStore) Get() *models.UISettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp := *s.settings
	return &cp
}

// Save replaces the settings from a raw JSON payload (the frontend
// sends the `adex-settings` object verbatim). Unknown / missing fields
// are tolerated: we unmarshal onto a defaults base so a partial or
// older-shaped payload still yields a complete, valid struct. The
// frontend remains the UX source of truth for *values*; this store is
// the persistence + crash-recovery authority.
func (s *UIStore) Save(jsonPayload string) error {
	// Start from defaults so any field the frontend omitted (e.g. a
	// setting added in a newer build than the one that wrote the
	// localStorage cache) keeps a sane value instead of a Go zero.
	next := models.DefaultUISettings()
	if err := json.Unmarshal([]byte(jsonPayload), next); err != nil {
		return fmt.Errorf("invalid UI settings payload: %w", err)
	}

	next.Version = "1.0"
	next.UpdatedAt = time.Now()

	s.mu.Lock()
	s.settings = next
	s.mu.Unlock()

	if err := s.persist(next); err != nil {
		return err
	}
	s.logger.Debug("UI settings saved", map[string]interface{}{"file": s.file})
	return nil
}

// File exposes the on-disk path (used by tests / diagnostics).
func (s *UIStore) File() string { return s.file }

// load reads the settings file, falling back through .bak then
// defaults. It always returns a usable struct.
func (s *UIStore) load() *models.UISettings {
	if v, ok := s.tryRead(s.file); ok {
		s.logger.Info("Loaded UI settings", map[string]interface{}{"file": s.file})
		// Repair values naming a theme or layout the app does not ship, and
		// write the repair back so the same fault is not re-reported on every
		// start. Without this an unresolvable id persists indefinitely.
		if validateDisplay(v) {
			s.logger.Warn("Repaired invalid display settings", map[string]interface{}{
				"theme": v.Display.Theme, "layout": v.Display.Layout,
			})
			if err := s.persist(v); err != nil {
				s.logger.Warn("Failed to persist repaired UI settings", map[string]interface{}{"error": err.Error()})
			}
		}
		return v
	}

	// Primary unreadable/corrupt — try the backup.
	if v, ok := s.tryRead(s.file + ".bak"); ok {
		s.logger.Warn("UI settings recovered from backup", map[string]interface{}{"file": s.file + ".bak"})
		// Rewrite primary from the recovered copy.
		if err := s.persist(v); err != nil {
			s.logger.Warn("Failed to rewrite UI settings from backup", map[string]interface{}{"error": err.Error()})
		}
		return v
	}

	// First run (or unrecoverable) — write defaults.
	s.logger.Info("No UI settings file; writing defaults", map[string]interface{}{"file": s.file})
	def := models.DefaultUISettings()
	if err := s.persist(def); err != nil {
		s.logger.Warn("Failed to write default UI settings", map[string]interface{}{"error": err.Error()})
	}
	return def
}

// tryRead reads + unmarshals one path onto a defaults base. ok=false
// means the file is missing or unparseable (caller falls back).
func (s *UIStore) tryRead(path string) (*models.UISettings, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	out := models.DefaultUISettings()
	if err := json.Unmarshal(data, out); err != nil {
		s.logger.Warn("Failed to parse UI settings file", map[string]interface{}{
			"file":  path,
			"error": err.Error(),
		})
		return nil, false
	}
	return out, true
}

// persist writes settings atomically: rotate current → .bak, write
// .tmp, fsync-free rename over the primary. A failed rename cleans up
// the temp file so we never leave a half-written primary.
func (s *UIStore) persist(v *models.UISettings) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal UI settings: %w", err)
	}

	// Best-effort backup of the last good file before overwriting.
	if cur, err := os.ReadFile(s.file); err == nil {
		if err := os.WriteFile(s.file+".bak", cur, 0o644); err != nil {
			s.logger.Warn("Failed to write UI settings backup", map[string]interface{}{"error": err.Error()})
		}
	}

	tmp := s.file + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp UI settings file: %w", err)
	}
	if err := os.Rename(tmp, s.file); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("failed to rename temp UI settings file: %w", err)
	}
	return nil
}

// GPUAccelerationDisabled reports whether the user turned off webview GPU
// acceleration in Settings → Advanced.
//
// Read directly from the file rather than through the store, because the
// window is constructed before services start — there is no coordinator to
// ask yet.
func GPUAccelerationDisabled() bool {
	data, err := os.ReadFile(filepath.Join(appdir.Config(), uiSettingsFileName))
	if err != nil {
		return false
	}
	var probe struct {
		Advanced struct {
			DisableGPU bool `json:"disableGpu"`
		} `json:"advanced"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return false
	}
	return probe.Advanced.DisableGPU
}
