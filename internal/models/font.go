package models

import (
	"time"
)

// FontConfiguration represents a complete font configuration for the terminal
type FontConfiguration struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Family      string            `json:"family"`
	Size        int               `json:"size"`
	Weight      string            `json:"weight"`
	LineHeight  float64           `json:"lineHeight"`
	LetterSpacing float64         `json:"letterSpacing"`
	Ligatures   bool              `json:"ligatures"`
	Antialias   bool              `json:"antialias"`
	Hinting     FontHinting       `json:"hinting"`
	Features    map[string]bool   `json:"features"`
	IsBuiltIn   bool              `json:"isBuiltIn"`
	IsDefault   bool              `json:"isDefault"`
	Author      string            `json:"author,omitempty"`
	Version     string            `json:"version,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FontHinting represents font hinting options
type FontHinting string

const (
	FontHintingNone   FontHinting = "none"
	FontHintingSlight FontHinting = "slight"
	FontHintingMedium FontHinting = "medium"
	FontHintingFull   FontHinting = "full"
)

// FontWeight represents font weight options
type FontWeight string

const (
	FontWeightThin       FontWeight = "100"
	FontWeightExtraLight FontWeight = "200"
	FontWeightLight      FontWeight = "300"
	FontWeightNormal     FontWeight = "400"
	FontWeightMedium     FontWeight = "500"
	FontWeightSemiBold   FontWeight = "600"
	FontWeightBold       FontWeight = "700"
	FontWeightExtraBold  FontWeight = "800"
	FontWeightBlack      FontWeight = "900"
)

// SystemFont represents a system font that can be used in the terminal
type SystemFont struct {
	Family      string       `json:"family"`
	DisplayName string       `json:"displayName"`
	Style       string       `json:"style"`
	Weight      string       `json:"weight"`
	IsMonospace bool         `json:"isMonospace"`
	IsInstalled bool         `json:"isInstalled"`
	FilePath    string       `json:"filePath,omitempty"`
	Format      FontFormat   `json:"format"`
	Size        int64        `json:"size,omitempty"`
	Checksum    string       `json:"checksum,omitempty"`
	Metadata    FontMetadata `json:"metadata"`
}

// FontFormat represents the format of a font file
type FontFormat string

const (
	FontFormatTTF      FontFormat = "ttf"
	FontFormatOTF      FontFormat = "otf"
	FontFormatWOFF     FontFormat = "woff"
	FontFormatWOFF2    FontFormat = "woff2"
	FontFormatEOT      FontFormat = "eot"
	FontFormatSVG      FontFormat = "svg"
	FontFormatUnknown  FontFormat = "unknown"
)

// FontMetadata contains additional metadata about a font
type FontMetadata struct {
	PSName      string            `json:"psName,omitempty"`
	Version     string            `json:"version,omitempty"`
	Copyright   string            `json:"copyright,omitempty"`
	Trademark   string            `json:"trademark,omitempty"`
	Manufacturer string           `json:"manufacturer,omitempty"`
	Designer    string            `json:"designer,omitempty"`
	Description string            `json:"description,omitempty"`
	VendorURL   string            `json:"vendorUrl,omitempty"`
	DesignerURL string            `json:"designerUrl,omitempty"`
	License     string            `json:"license,omitempty"`
	LicenseURL  string            `json:"licenseUrl,omitempty"`
	Panose      []int             `json:"panose,omitempty"`
	VendorID    string            `json:"vendorId,omitempty"`
	CSSFamily   string            `json:"cssFamily,omitempty"`
	CSSWeight   string            `json:"cssWeight,omitempty"`
	CSSStyle    string            `json:"cssStyle,omitempty"`
	UnicodeRange []string         `json:"unicodeRange,omitempty"`
	SupportedScripts []string     `json:"supportedScripts,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// FontPreview represents a font preview configuration
type FontPreview struct {
	Text          string            `json:"text"`
	SampleText    string            `json:"sampleText"`
	FontSize      int               `json:"fontSize"`
	LineHeight    float64           `json:"lineHeight"`
	LetterSpacing float64           `json:"letterSpacing"`
	Colors        FontPreviewColors `json:"colors"`
	ShowLineNumbers bool            `json:"showLineNumbers"`
	ShowCharacterInfo bool          `json:"showCharacterInfo"`
	RenderingHints []string         `json:"renderingHints"`
}

// FontPreviewColors represents colors for font preview
type FontPreviewColors struct {
	Background   string `json:"background"`
	Foreground   string `json:"foreground"`
	Selection    string `json:"selection"`
	LineNumbers  string `json:"lineNumbers"`
	Cursor       string `json:"cursor"`
	Accent       string `json:"accent"`
}

// FontSettings represents font-related settings
type FontSettings struct {
	DefaultFontID    string    `json:"defaultFontId"`
	FallbackFont     string    `json:"fallbackFont"`
	FontSize         int       `json:"fontSize"`
	LineHeight       float64   `json:"lineHeight"`
	LetterSpacing    float64   `json:"letterSpacing"`
	EnableLigatures  bool      `json:"enableLigatures"`
	EnableAntialias  bool      `json:"enableAntialias"`
	Hinting          FontHinting `json:"hinting"`
	FontFeatureSettings map[string]bool `json:"fontFeatureSettings"`
	AllowCustomFonts bool      `json:"allowCustomFonts"`
	FontDirectories  []string  `json:"fontDirectories"`
	AutoDetectFonts  bool      `json:"autoDetectFonts"`
	LastScanTime     time.Time `json:"lastScanTime"`
}

// FontValidationResult represents the result of font validation
type FontValidationResult struct {
	IsValid     bool     `json:"isValid"`
	Errors      []string `json:"errors"`
	Warnings    []string `json:"warnings"`
	Suggestions []string `json:"suggestions"`
	Features    []string `json:"features"`
	Limitations []string `json:"limitations"`
}

// FontMetrics represents font metrics information
type FontMetrics struct {
	Family         string  `json:"family"`
	Size           int     `json:"size"`
	Weight         string  `json:"weight"`
	Ascent         float64 `json:"ascent"`
	Descent        float64 `json:"descent"`
	LineGap        float64 `json:"lineGap"`
	CapHeight      float64 `json:"capHeight"`
	XHeight        float64 `json:"xHeight"`
	AvgCharWidth   float64 `json:"avgCharWidth"`
	MaxCharWidth   float64 `json:"maxCharWidth"`
	IsMonospace    bool    `json:"isMonospace"`
	CharacterCount int     `json:"characterCount"`
	SupportedChars []rune  `json:"supportedChars,omitempty"`
}

// FontImportRequest represents a request to import a font
type FontImportRequest struct {
	Source      string            `json:"source"` // file path, URL, or base64 data
	Name        string            `json:"name"`
	Family      string            `json:"family"`
	Weight      string            `json:"weight"`
	Style       string            `json:"style"`
	License     string            `json:"license,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Overwrite   bool              `json:"overwrite"`
	ValidateOnly bool             `json:"validateOnly"`
}

// FontImportResult represents the result of a font import operation
type FontImportResult struct {
	Success      bool               `json:"success"`
	Font         *SystemFont        `json:"font,omitempty"`
	Configuration *FontConfiguration `json:"configuration,omitempty"`
	Errors       []string           `json:"errors"`
	Warnings     []string           `json:"warnings"`
	ImportPath   string             `json:"importPath,omitempty"`
	Checksum     string             `json:"checksum,omitempty"`
}

// NewFontConfiguration creates a new font configuration
func NewFontConfiguration(id, name, family string) *FontConfiguration {
	now := time.Now()
	return &FontConfiguration{
		ID:           id,
		Name:         name,
		Family:       family,
		Size:         14,
		Weight:       string(FontWeightNormal),
		LineHeight:   1.4,
		LetterSpacing: 0,
		Ligatures:    true,
		Antialias:    true,
		Hinting:      FontHintingSlight,
		Features:     make(map[string]bool),
		IsBuiltIn:    false,
		IsDefault:    false,
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]string),
	}
}

// NewDefaultFontSettings creates default font settings
func NewDefaultFontSettings() *FontSettings {
	return &FontSettings{
		DefaultFontID:         "default-mono",
		FallbackFont:          "monospace",
		FontSize:              14,
		LineHeight:            1.4,
		LetterSpacing:         0,
		EnableLigatures:       true,
		EnableAntialias:       true,
		Hinting:               FontHintingSlight,
		FontFeatureSettings:   make(map[string]bool),
		AllowCustomFonts:      false,
		FontDirectories:       []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			"~/.fonts",
			"~/.local/share/fonts",
		},
		AutoDetectFonts: true,
		LastScanTime:    time.Now(),
	}
}

// IsValid checks if the font configuration is valid
func (fc *FontConfiguration) IsValid() bool {
	if fc.Family == "" {
		return false
	}
	if fc.Size < 6 || fc.Size > 72 {
		return false
	}
	if fc.LineHeight < 0.8 || fc.LineHeight > 3.0 {
		return false
	}
	if fc.LetterSpacing < -2 || fc.LetterSpacing > 10 {
		return false
	}
	return true
}

// Clone creates a deep copy of the font configuration
func (fc *FontConfiguration) Clone() *FontConfiguration {
	clone := *fc

	// Deep copy maps
	if fc.Features != nil {
		clone.Features = make(map[string]bool)
		for k, v := range fc.Features {
			clone.Features[k] = v
		}
	}

	if fc.Metadata != nil {
		clone.Metadata = make(map[string]string)
		for k, v := range fc.Metadata {
			clone.Metadata[k] = v
		}
	}

	return &clone
}

// GetCSSString returns the CSS font string
func (fc *FontConfiguration) GetCSSString() string {
	css := ""

	// Font style and weight
	css += fc.Weight + " "

	// Font size and line height
	css += string(rune(fc.Size)) + "px/" +
		string(rune(fc.LineHeight)) + " "

	// Font family
	css += "'" + fc.Family + "', monospace"

	return css
}

// ToSystemFont converts the configuration to a system font
func (fc *FontConfiguration) ToSystemFont() *SystemFont {
	return &SystemFont{
		Family:      fc.Family,
		DisplayName: fc.Name,
		Weight:      fc.Weight,
		IsMonospace: true,
		IsInstalled: true,
		Format:      FontFormatUnknown,
		Metadata: FontMetadata{
			CSSFamily: fc.Family,
			CSSWeight: fc.Weight,
		},
	}
}