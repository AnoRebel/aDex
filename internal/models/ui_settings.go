package models

import (
	"time"
)

// UISettings is the typed, file-persisted mirror of the frontend's
// `adex-settings` object (the SettingsData interface in
// AdexSettingsModal.vue). It is DELIBERATELY separate from AppSettings:
//
//   - AppSettings is the legacy eDEX-shaped 477-line config consumed by
//     the font / system / theme Go services with their own field
//     expectations. Cramming the lean frontend shape into it would
//     force changes across those consumers.
//   - UISettings instead matches the frontend SettingsData 1:1, field
//     for field, so the round-trip is a plain JSON marshal/unmarshal
//     with no lossy mapping layer.
//
// Persistence model: this struct is the SOURCE OF TRUTH (a real file
// under the user config dir, atomic write + .bak backup). The
// frontend's localStorage is only a fast reactive cache that is
// seeded from here on boot and flushed back here (debounced) on
// change. Backend always wins on a boot conflict.
//
// When the frontend adds a setting, add the matching field here too
// (the user chose typed reconciliation over an opaque blob precisely
// so the Go side stays a checkable contract).
//
// This is a plain value DTO — no embedded mutex. Concurrency is the
// owning store's responsibility (settings.UIStore), so the struct
// stays freely copyable for safe snapshot returns.
type UISettings struct {
	// Metadata — not part of the frontend shape; we stamp these on
	// save so the file is self-describing and migrations can branch
	// on Version later.
	Version   string    `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`

	Shell    UIShellSettings    `json:"shell"`
	Display  UIDisplaySettings  `json:"display"`
	Audio    UIAudioSettings    `json:"audio"`
	System   UISystemSettings   `json:"system"`
	Network  UINetworkSettings  `json:"network"`
	Advanced UIAdvancedSettings `json:"advanced"`
}

// UIShellSettings ← SettingsData.shell
type UIShellSettings struct {
	Path             string `json:"path"`
	Args             string `json:"args"`
	WorkingDirectory string `json:"workingDirectory"`
}

// UIDisplaySettings ← SettingsData.display
type UIDisplaySettings struct {
	Theme            string `json:"theme"`
	KeyboardLayout   string `json:"keyboardLayout"`
	TerminalFontSize int    `json:"terminalFontSize"`
	FontFamily       string `json:"fontFamily"`
	// Layout preset override. Empty = follow the active theme's layout.
	Layout string `json:"layout"`
}

// UIAudioSettings ← SettingsData.audio
type UIAudioSettings struct {
	Enabled          bool   `json:"enabled"`
	Volume           int    `json:"volume"`
	Soundpack        string `json:"soundpack"`
	MuteInBackground bool   `json:"muteInBackground"`
	Boot             bool   `json:"boot"`
	Shutdown         bool   `json:"shutdown"`
	// Per-category opt-outs (keyboard / destructive / interface clicks).
	CategoryKeyboard    bool `json:"categoryKeyboard"`
	CategoryDestructive bool `json:"categoryDestructive"`
	CategoryInterface   bool `json:"categoryInterface"`
}

// UISystemSettings ← SettingsData.system
type UISystemSettings struct {
	// "home" or "cwd" — where new shells / file manager open.
	InitialCwd string `json:"initialCwd"`
	// "12h" or "24h".
	ClockFormat     string `json:"clockFormat"`
	BootAnimation   bool   `json:"bootAnimation"`
	GridBackground  bool   `json:"gridBackground"`
	PerformanceMode bool   `json:"performanceMode"`
}

// UINetworkSettings ← SettingsData.network
type UINetworkSettings struct {
	PingTarget        string `json:"pingTarget"`
	Adapter           string `json:"adapter"`
	PingIntervalMs    int    `json:"pingIntervalMs"`
	TrafficIntervalMs int    `json:"trafficIntervalMs"`
	GeoipEnabled      bool   `json:"geoipEnabled"`
	GeoipEndpoint     string `json:"geoipEndpoint"`
}

// UIAdvancedSettings ← SettingsData.advanced
type UIAdvancedSettings struct {
	AllowWindowedMode         bool `json:"allowWindowedMode"`
	ExperimentalFeatures      bool `json:"experimentalFeatures"`
	WebglRenderer             bool `json:"webglRenderer"`
	DebugMode                 bool `json:"debugMode"`
	Scrollback                int  `json:"scrollback"`
	Nointro                   bool `json:"nointro"`
	ForceFullscreen           bool `json:"forceFullscreen"`
	HideDotfiles              bool `json:"hideDotfiles"`
	FsListView                bool `json:"fsListView"`
	ExperimentalGlobeFeatures bool `json:"experimentalGlobeFeatures"`
}

// DefaultUISettings mirrors getDefaultSettings() in AdexSettingsModal.vue
// exactly. Used when no settings file exists yet (first launch) and as
// the merge base so a partial/older file gets sane values for any new
// field rather than Go zero-values (false / 0 / "").
func DefaultUISettings() *UISettings {
	return &UISettings{
		Version:   "1.0",
		UpdatedAt: time.Now(),
		Shell: UIShellSettings{
			Path:             "",
			Args:             "--login",
			WorkingDirectory: "/",
		},
		Display: UIDisplaySettings{
			Theme:            "default",
			KeyboardLayout:   "en-US",
			TerminalFontSize: 14,
			FontFamily:       "'Fira Code', monospace",
			Layout:           "",
		},
		Audio: UIAudioSettings{
			Enabled:             true,
			Volume:              50,
			Soundpack:           "adex",
			MuteInBackground:    false,
			Boot:                true,
			Shutdown:            true,
			CategoryKeyboard:    true,
			CategoryDestructive: true,
			CategoryInterface:   true,
		},
		System: UISystemSettings{
			InitialCwd:      "home",
			ClockFormat:     "24h",
			BootAnimation:   true,
			GridBackground:  true,
			PerformanceMode: false,
		},
		Network: UINetworkSettings{
			PingTarget:        "8.8.8.8:53",
			Adapter:           "",
			PingIntervalMs:    2000,
			TrafficIntervalMs: 1000,
			GeoipEnabled:      true,
			GeoipEndpoint:     "",
		},
		Advanced: UIAdvancedSettings{
			AllowWindowedMode:         false,
			ExperimentalFeatures:      false,
			WebglRenderer:             true,
			DebugMode:                 false,
			Scrollback:                5000,
			Nointro:                   false,
			ForceFullscreen:           false,
			HideDotfiles:              true,
			FsListView:                false,
			ExperimentalGlobeFeatures: false,
		},
	}
}
