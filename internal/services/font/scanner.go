package font

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"aDex/internal/models"
)

// FontScanner handles scanning and importing of system fonts
type FontScanner struct {
	mu          sync.RWMutex
	scanCache   map[string]*models.SystemFont
	knownFonts  map[string]string // family -> file path mapping
}

// NewFontScanner creates a new font scanner
func NewFontScanner() *FontScanner {
	return &FontScanner{
		scanCache:  make(map[string]*models.SystemFont),
		knownFonts: make(map[string]string),
	}
}

// ScanSystemFonts scans system directories for available fonts
func (fs *FontScanner) ScanSystemFonts(directories []string) (map[string]*models.SystemFont, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fonts := make(map[string]*models.SystemFont)
	seenPaths := make(map[string]bool)

	for _, dir := range directories {
		// Expand user directory
		if strings.HasPrefix(dir, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				dir = filepath.Join(home, dir[2:])
			}
		}

		// Scan directory
		dirFonts, err := fs.scanDirectory(dir, seenPaths)
		if err != nil {
			// Log warning but continue with other directories
			continue
		}

		// Merge results
		for id, font := range dirFonts {
			fonts[id] = font
		}
	}

	// Cache results
	fs.scanCache = fonts

	return fonts, nil
}

// ImportFont imports a font from a specific source
func (fs *FontScanner) ImportFont(source string, req *models.FontImportRequest) (*models.SystemFont, error) {
	// Check if source is a file path
	if _, err := os.Stat(source); err == nil {
		return fs.importFontFromFile(source, req)
	}

	// For now, only file-based imports are supported
	// URL and base64 imports could be added later
	return nil, fmt.Errorf("unsupported import source: %s", source)
}

// GetFontMetrics retrieves metrics for a font
func (fs *FontScanner) GetFontMetrics(font *models.SystemFont, size int) (*models.FontMetrics, error) {
	// This is a simplified implementation
	// In a real implementation, you would use font parsing libraries
	// to extract actual metrics from font files

	metrics := &models.FontMetrics{
		Family:       font.Family,
		Size:         size,
		Weight:       font.Weight,
		Ascent:       float64(size) * 0.8,
		Descent:      float64(size) * 0.2,
		LineGap:      float64(size) * 0.2,
		CapHeight:    float64(size) * 0.7,
		XHeight:      float64(size) * 0.5,
		AvgCharWidth: float64(size) * 0.6,
		MaxCharWidth: float64(size) * 0.9,
		IsMonospace:  font.IsMonospace,
	}

	// For monospace fonts, calculate exact character width
	if font.IsMonospace {
		metrics.AvgCharWidth = float64(size) * 0.6
		metrics.MaxCharWidth = metrics.AvgCharWidth
	}

	return metrics, nil
}

// scanDirectory scans a directory for font files
func (fs *FontScanner) scanDirectory(dir string, seenPaths map[string]bool) (map[string]*models.SystemFont, error) {
	fonts := make(map[string]*models.SystemFont)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fonts, nil
	}

	// Walk through directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files that can't be accessed
		}

		// Skip directories and already seen files
		if info.IsDir() || seenPaths[path] {
			return nil
		}

		// Check if file is a font file
		if fs.isFontFile(path) {
			font, err := fs.parseFontFile(path)
			if err != nil {
				// Log warning but continue
				return nil
			}

			if font != nil {
				fonts[font.Family] = font
				seenPaths[path] = true
			}
		}

		return nil
	})

	return fonts, err
}

// isFontFile checks if a file is a font file based on extension
func (fs *FontScanner) isFontFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ttf", ".otf", ".woff", ".woff2", ".eot":
		return true
	default:
		return false
	}
}

// parseFontFile parses a font file and extracts metadata
func (fs *FontScanner) parseFontFile(path string) (*models.SystemFont, error) {
	// This is a simplified implementation
	// In a real implementation, you would use font parsing libraries
	// like github.com/go-text/typesetting/font or similar

	filename := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))
	name := strings.TrimSuffix(filename, ext)

	// Determine font format
	format := models.FontFormatUnknown
	switch ext {
	case ".ttf":
		format = models.FontFormatTTF
	case ".otf":
		format = models.FontFormatOTF
	case ".woff":
		format = models.FontFormatWOFF
	case ".woff2":
		format = models.FontFormatWOFF2
	case ".eot":
		format = models.FontFormatEOT
	}

	// Create system font
	font := &models.SystemFont{
		Family:      fs.guessFontFamily(name),
		DisplayName: name,
		Weight:      fs.guessFontWeight(name),
		IsMonospace: fs.guessIsMonospace(name),
		IsInstalled: true,
		FilePath:    path,
		Format:      format,
		Metadata: models.FontMetadata{
			CSSFamily: fs.guessFontFamily(name),
			CSSWeight: fs.guessFontWeight(name),
		},
	}

	// Get file size
	if info, err := os.Stat(path); err == nil {
		font.Size = info.Size()
	}

	return font, nil
}

// guessFontFamily attempts to guess the font family from the filename
func (fs *FontScanner) guessFontFamily(filename string) string {
	// Remove common suffixes
	name := strings.ToLower(filename)

	suffixes := []string{
		"-regular", "-bold", "-italic", "-oblique", "-light", "-medium",
		"-thin", "-black", "-extrabold", "-semibold", "-demibold",
		"regular", "bold", "italic", "oblique", "light", "medium",
		"thin", "black", "extrabold", "semibold", "demibold",
	}

	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
		name = strings.TrimSuffix(name, strings.Title(suffix))
	}

	// Remove common font variations
	variations := []string{"mono", "code", "sans", "serif"}
	for _, variation := range variations {
		if strings.HasSuffix(name, variation) {
			name = strings.TrimSuffix(name, variation)
			name += " " + strings.Title(variation)
			break
		}
	}

	// Convert to title case
	parts := strings.Fields(name)
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}

	return strings.Join(parts, " ")
}

// guessFontWeight attempts to guess the font weight from the filename
func (fs *FontScanner) guessFontWeight(filename string) string {
	name := strings.ToLower(filename)

	// Check for weight indicators
	if strings.Contains(name, "thin") {
		return string(models.FontWeightThin)
	}
	if strings.Contains(name, "extralight") || strings.Contains(name, "ultralight") {
		return string(models.FontWeightExtraLight)
	}
	if strings.Contains(name, "light") {
		return string(models.FontWeightLight)
	}
	if strings.Contains(name, "medium") {
		return string(models.FontWeightMedium)
	}
	if strings.Contains(name, "semibold") || strings.Contains(name, "demibold") {
		return string(models.FontWeightSemiBold)
	}
	if strings.Contains(name, "bold") {
		return string(models.FontWeightBold)
	}
	if strings.Contains(name, "extrabold") || strings.Contains(name, "ultrabold") {
		return string(models.FontWeightExtraBold)
	}
	if strings.Contains(name, "black") || strings.Contains(name, "heavy") {
		return string(models.FontWeightBlack)
	}

	// Default to normal
	return string(models.FontWeightNormal)
}

// guessIsMonospace attempts to guess if the font is monospace
func (fs *FontScanner) guessIsMonospace(filename string) bool {
	name := strings.ToLower(filename)

	// Check for monospace indicators
	monospaceIndicators := []string{
		"mono", "code", "console", "terminal", "courier",
		"monaco", "inconsolata", "source", "fira", "jetbrains",
		"cascadia", "ibm plex mono", "ubuntu mono", "anonymous pro",
	}

	for _, indicator := range monospaceIndicators {
		if strings.Contains(name, indicator) {
			return true
		}
	}

	return false
}

// importFontFromFile imports a font from a file
func (fs *FontScanner) importFontFromFile(path string, req *models.FontImportRequest) (*models.SystemFont, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("font file not found: %s", path)
	}

	// Parse font file
	font, err := fs.parseFontFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font file: %v", err)
	}

	// Override with request data if provided
	if req.Name != "" {
		font.DisplayName = req.Name
	}
	if req.Family != "" {
		font.Family = req.Family
		font.Metadata.CSSFamily = req.Family
	}
	if req.Weight != "" {
		font.Weight = req.Weight
		font.Metadata.CSSWeight = req.Weight
	}
	if req.Style != "" {
		font.Style = req.Style
		font.Metadata.CSSStyle = req.Style
	}

	// Add metadata from request
	if req.License != "" {
		font.Metadata.License = req.License
	}
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			font.Metadata.Custom[k] = v
		}
	}

	return font, nil
}

// GetCommonMonospaceFonts returns a list of common monospace fonts
func (fs *FontScanner) GetCommonMonospaceFonts() []*models.SystemFont {
	commonFonts := []*models.SystemFont{
		{
			Family:      "JetBrains Mono",
			DisplayName: "JetBrains Mono",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatTTF,
			Metadata: models.FontMetadata{
				CSSFamily: "JetBrains Mono",
				CSSWeight: string(models.FontWeightNormal),
				Description: "JetBrains Mono typeface for developers",
			},
		},
		{
			Family:      "Fira Code",
			DisplayName: "Fira Code",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatTTF,
			Metadata: models.FontMetadata{
				CSSFamily: "Fira Code",
				CSSWeight: string(models.FontWeightNormal),
				Description: "Fira Code monospace font with programming ligatures",
			},
		},
		{
			Family:      "Source Code Pro",
			DisplayName: "Source Code Pro",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatOTF,
			Metadata: models.FontMetadata{
				CSSFamily: "Source Code Pro",
				CSSWeight: string(models.FontWeightNormal),
				Description: "Source Code Pro by Adobe",
			},
		},
		{
			Family:      "Cascadia Code",
			DisplayName: "Cascadia Code",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatTTF,
			Metadata: models.FontMetadata{
				CSSFamily: "Cascadia Code",
				CSSWeight: string(models.FontWeightNormal),
				Description: "Cascadia Code by Microsoft",
			},
		},
		{
			Family:      "IBM Plex Mono",
			DisplayName: "IBM Plex Mono",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatOTF,
			Metadata: models.FontMetadata{
				CSSFamily: "IBM Plex Mono",
				CSSWeight: string(models.FontWeightNormal),
				Description: "IBM Plex Mono",
			},
		},
		{
			Family:      "Ubuntu Mono",
			DisplayName: "Ubuntu Mono",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatTTF,
			Metadata: models.FontMetadata{
				CSSFamily: "Ubuntu Mono",
				CSSWeight: string(models.FontWeightNormal),
				Description: "Ubuntu Mono font",
			},
		},
		{
			Family:      "Consolas",
			DisplayName: "Consolas",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatTTF,
			Metadata: models.FontMetadata{
				CSSFamily: "Consolas",
				CSSWeight: string(models.FontWeightNormal),
				Description: "Consolas font by Microsoft",
			},
		},
		{
			Family:      "Monaco",
			DisplayName: "Monaco",
			Weight:      string(models.FontWeightNormal),
			IsMonospace: true,
			IsInstalled: true,
			Format:      models.FontFormatTTF,
			Metadata: models.FontMetadata{
				CSSFamily: "Monaco",
				CSSWeight: string(models.FontWeightNormal),
				Description: "Monaco font by Apple",
			},
		},
	}

	// Sort by display name
	sort.Slice(commonFonts, func(i, j int) bool {
		return commonFonts[i].DisplayName < commonFonts[j].DisplayName
	})

	return commonFonts
}

// GetCachedFonts returns the cached font scan results
func (fs *FontScanner) GetCachedFonts() map[string]*models.SystemFont {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	result := make(map[string]*models.SystemFont)
	for id, font := range fs.scanCache {
		result[id] = font
	}
	return result
}

// ClearCache clears the font scan cache
func (fs *FontScanner) ClearCache() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.scanCache = make(map[string]*models.SystemFont)
	fs.knownFonts = make(map[string]string)
}