package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"aDex-UI/internal/models"
	"aDex-UI/internal/services/theme"
)

// ThemeTestSuite provides a comprehensive test suite for the theme system
type ThemeTestSuite struct {
	suite.Suite
	themeService   *theme.Service
	tempDir        string
	legacyTheme    *models.LegacyTheme
	sampleTheme    *models.Theme
	converter      *theme.LegacyThemeConverter
	variableGen    *theme.VariableGenerator
}

// SetupSuite runs once before all tests
func (suite *ThemeTestSuite) SetupSuite() {
	// Create temporary directory for test themes
	tempDir, err := os.MkdirTemp("", "dex-theme-test-*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// Initialize theme service
	config := &theme.ThemeServiceConfig{
		ThemeDirectory: tempDir,
		AutoSave:       true,
		AutoReload:     true,
		CacheEnabled:   true,
	}

	service := theme.NewThemeService(config)
	err = service.Initialize()
	require.NoError(suite.T(), err)
	suite.themeService = service

	// Create test legacy theme
	suite.legacyTheme = &models.LegacyTheme{
		Name:        "Test Legacy Theme",
		DisplayName: "Test Legacy Theme",
		Description: "A test theme for legacy conversion",
		IsDark:      true,
		Colors: models.LegacyColors{
			Primary:    "#00ff41",
			Secondary:  "#00cc33",
			Accent:     "#ffaa00",
			Background: "#0a0a0a",
			Surface:    "#1a1a1a",
			Text:       "#ffffff",
			TextSecondary: "#cccccc",
			Border:     "#333333",
			Success:    "#00ff41",
			Warning:    "#ffaa00",
			Error:      "#ff3333",
			Info:       "#00aaff",
		},
		Typography: models.LegacyTypography{
			FontFamily: "Inter, sans-serif",
			FontSize: models.LegacyFontSize{
				XS:   "0.75rem",
				SM:   "0.875rem",
				Base: "1rem",
				LG:   "1.125rem",
				XL:   "1.25rem",
			},
			FontWeight: models.LegacyFontWeight{
				Light:  300,
				Normal: 400,
				Medium: 500,
				Bold:   700,
			},
			LineHeight: models.LegacyLineHeight{
				Tight:   1.25,
				Normal:  1.5,
				Relaxed: 1.75,
			},
		},
		Effects: models.LegacyEffects{
			Glow:       true,
			Animation:  true,
			Shadows:    true,
			Transitions: true,
			Gradients:  false,
			Backdrop:   false,
		},
		IsCustom: true,
	}

	// Create sample modern theme
	suite.sampleTheme = models.NewTheme("test-theme", "Test Theme")
	suite.sampleTheme.Description = "A test theme for unit testing"
	suite.sampleTheme.Author = "Test Suite"
	suite.sampleTheme.Version = "1.0.0"

	// Initialize converter and variable generator
	suite.converter = theme.NewLegacyThemeConverter()
	suite.variableGen = theme.NewVariableGenerator()
}

// TearDownSuite runs once after all tests
func (suite *ThemeTestSuite) TearDownSuite() {
	// Clean up temporary directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// SetupTest runs before each test
func (suite *ThemeTestSuite) SetupTest() {
	// Reset any test-specific state here
}

// TestLegacyThemeConversion tests the conversion from legacy to modern theme format
func (suite *ThemeTestSuite) TestLegacyThemeConversion() {
	// Test successful conversion
	convertedTheme, err := suite.converter.ConvertFromLegacy(suite.legacyTheme)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), convertedTheme)
	assert.Equal(suite.T(), "test-legacy-theme", convertedTheme.ID)
	assert.Equal(suite.T(), "Test Legacy Theme", convertedTheme.Name)
	assert.Equal(suite.T(), "Test Legacy Theme", convertedTheme.DisplayName)
	assert.Equal(suite.T(), suite.legacyTheme.Description, convertedTheme.Description)
	assert.True(suite.T(), convertedTheme.IsDark)
	assert.True(suite.T(), convertedTheme.IsCustom)

	// Test color conversion
	assert.Equal(suite.T(), "#00ff41", convertedTheme.Colors.Accent.Primary)
	assert.Equal(suite.T(), "#0a0a0a", convertedTheme.Colors.Background.Primary)
	assert.Equal(suite.T(), "#ffffff", convertedTheme.Colors.Foreground.Primary)

	// Test font conversion
	assert.Equal(suite.T(), "Inter, sans-serif", convertedTheme.Fonts.Families.Primary)
	assert.Equal(suite.T(), "1rem", convertedTheme.Fonts.Sizes.Base)
	assert.Equal(suite.T(), "400", convertedTheme.Fonts.Weights.Normal)

	// Test effects conversion
	assert.True(suite.T(), convertedTheme.Effects.Animations.Duration.Normal != "") // Animation enabled
	assert.Len(suite.T(), convertedTheme.Effects.Shadows, 0) // No gradients by default
}

// TestLegacyThemeConversionWithNilInput tests conversion with nil input
func (suite *ThemeTestSuite) TestLegacyThemeConversionWithNilInput() {
	convertedTheme, err := suite.converter.ConvertFromLegacy(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), convertedTheme)
	assert.Contains(suite.T(), err.Error(), "legacy theme cannot be nil")
}

// TestLegacyThemeConversionWithEmptyData tests conversion with empty theme data
func (suite *ThemeTestSuite) TestLegacyThemeConversionWithEmptyData() {
	emptyTheme := &models.LegacyTheme{
		Name: "",
		Colors: models.LegacyColors{},
	}

	convertedTheme, err := suite.converter.ConvertFromLegacy(emptyTheme)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), convertedTheme)
	assert.Equal(suite.T(), "unknown-theme", convertedTheme.ID)
	assert.Equal(suite.T(), "Unknown Theme", convertedTheme.Name)
}

// TestLegacyThemeBatchConversion tests batch conversion of multiple themes
func (suite *ThemeTestSuite) TestLegacyThemeBatchConversion() {
	themes := []*models.LegacyTheme{
		suite.legacyTheme,
		{
			Name: "Second Theme",
			Colors: models.LegacyColors{
				Primary: "#ff0000",
				Background: "#ffffff",
			},
		},
	}

	results := suite.converter.ConvertBatch(themes)

	assert.Len(suite.T(), results, 2)
	assert.NotNil(suite.T(), results[0].Theme)
	assert.NoError(suite.T(), results[0].Error)
	assert.NotNil(suite.T(), results[1].Theme)
	assert.NoError(suite.T(), results[1].Error)
}

// TestThemeValidation tests theme validation functionality
func (suite *ThemeTestSuite) TestThemeValidation() {
	// Test valid theme
	result := suite.themeService.ValidateTheme(suite.sampleTheme)
	assert.True(suite.T(), result.Valid)
	assert.Empty(suite.T(), result.Errors)
	assert.Empty(suite.T(), result.Warnings)

	// Test invalid theme (missing required fields)
	invalidTheme := &models.Theme{
		ID:   "", // Missing ID
		Name: "", // Missing name
	}

	result = suite.themeService.ValidateTheme(invalidTheme)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
}

// TestThemeSerialization tests theme save/load functionality
func (suite *ThemeTestSuite) TestThemeSerialization() {
	// Create a test theme
	testTheme := models.NewTheme("serialization-test", "Serialization Test")
	testTheme.Description = "Testing theme serialization"
	testTheme.Author = "Test Suite"
	testTheme.Version = "1.0.0"

	// Save theme
	err := suite.themeService.SaveTheme(testTheme)
	require.NoError(suite.T(), err)

	// Load theme back
	loadedTheme, err := suite.themeService.LoadTheme("serialization-test")
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), testTheme.ID, loadedTheme.ID)
	assert.Equal(suite.T(), testTheme.Name, loadedTheme.Name)
	assert.Equal(suite.T(), testTheme.Description, loadedTheme.Description)
	assert.Equal(suite.T(), testTheme.Author, loadedTheme.Author)
	assert.Equal(suite.T(), testTheme.Version, loadedTheme.Version)
}

// TestCSSVariableGeneration tests CSS variable generation from themes
func (suite *ThemeTestSuite) TestCSSVariableGeneration() {
	// Generate CSS variables
	css, err := suite.variableGen.GenerateCSSVariables(suite.sampleTheme)

	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), css)
	assert.NotEmpty(suite.T(), css.Variables)
	assert.Equal(suite.T(), suite.sampleTheme.ID, css.ThemeID)
	assert.Greater(suite.T(), css.VariableCount, 0)
	assert.NotEmpty(suite.T(), css.Hash)

	// Check that variables contain expected values
	assert.Contains(suite.T(), css.Variables, "--dex-theme-colors-background-primary")
	assert.Contains(suite.T(), css.Variables, "--dex-theme-fonts-families-primary")
}

// TestCSSVariableCaching tests CSS variable caching functionality
func (suite *ThemeTestSuite) TestCSSVariableCaching() {
	// Generate CSS variables first time
	css1, err := suite.variableGen.GenerateCSSVariables(suite.sampleTheme)
	require.NoError(suite.T(), err)

	// Generate again (should use cache)
	css2, err := suite.variableGen.GenerateCSSVariables(suite.sampleTheme)
	require.NoError(suite.T(), err)

	// Should be identical (from cache)
	assert.Equal(suite.T(), css1.Variables, css2.Variables)
	assert.Equal(suite.T(), css1.Hash, css2.Hash)
	assert.Equal(suite.T(), css1.GeneratedAt, css2.GeneratedAt)
}

// TestThemeServiceOperations tests theme service operations
func (suite *ThemeTestSuite) TestThemeServiceOperations() {
	// Test getting available themes
	themes := suite.themeService.GetAvailableThemes()
	assert.NotEmpty(suite.T(), themes)

	// Test setting current theme
	err := suite.themeService.SetCurrentTheme(suite.sampleTheme.ID)
	require.NoError(suite.T(), err)

	currentTheme := suite.themeService.GetCurrentTheme()
	assert.NotNil(suite.T(), currentTheme)
	assert.Equal(suite.T(), suite.sampleTheme.ID, currentTheme.ID)

	// Test getting built-in themes
	builtinThemes := suite.themeService.GetBuiltInThemes()
	assert.NotEmpty(suite.T(), builtinThemes)
}

// TestThemeEventSubscription tests theme event subscription system
func (suite *ThemeTestSuite) TestThemeEventSubscription() {
	eventReceived := false
	var eventTheme *models.Theme

	// Subscribe to theme change events
	unsubscribe := suite.themeService.SubscribeToThemeChanges(func(theme *models.Theme) {
		eventReceived = true
		eventTheme = theme
	})

	// Change current theme to trigger event
	err := suite.themeService.SetCurrentTheme(suite.sampleTheme.ID)
	require.NoError(suite.T(), err)

	// Wait a bit for event processing
	time.Sleep(100 * time.Millisecond)

	// Verify event was received
	assert.True(suite.T(), eventReceived)
	assert.NotNil(suite.T(), eventTheme)
	assert.Equal(suite.T(), suite.sampleTheme.ID, eventTheme.ID)

	// Unsubscribe
	unsubscribe()
}

// TestThemeImportExport tests theme import/export functionality
func (suite *ThemeTestSuite) TestThemeImportExport() {
	// Create test theme for export
	exportTheme := models.NewTheme("export-test", "Export Test")
	exportTheme.Description = "Testing theme export functionality"
	exportTheme.Author = "Test Suite"

	// Export theme
	exportData, err := suite.themeService.ExportTheme(exportTheme)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), exportData)
	assert.Equal(suite.T(), exportTheme.ID, exportData.Themes[0].ID)

	// Import theme back
	importResult, err := suite.themeService.ImportTheme(exportData)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), importResult.Success)
	assert.Len(suite.T(), importResult.ImportedThemes, 1)
}

// TestThemeConflictResolution tests theme conflict handling
func (suite *ThemeTestSuite) TestThemeConflictResolution() {
	// Create two themes with the same ID
	theme1 := models.NewTheme("conflict-test", "Theme 1")
	theme1.Description = "First theme"

	theme2 := models.NewTheme("conflict-test", "Theme 2")
	theme2.Description = "Second theme"

	// Save first theme
	err := suite.themeService.SaveTheme(theme1)
	require.NoError(suite.T(), err)

	// Try to import second theme (should create conflict)
	exportData := &theme.ThemeExportData{
		Themes: []*models.Theme{theme2},
		Version: "1.0.0",
		ExportDate: time.Now().Format(time.RFC3339),
	}

	importResult, err := suite.themeService.ImportTheme(exportData)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), importResult.Success) // Should fail due to conflict
	assert.Len(suite.T(), importResult.Conflicts, 1)
	assert.Equal(suite.T(), themeConflictTypeDuplicate, importResult.Conflicts[0].Type)
}

// TestThemeErrorHandling tests various error conditions
func (suite *ThemeTestSuite) TestThemeErrorHandling() {
	// Test loading non-existent theme
	theme, err := suite.themeService.LoadTheme("non-existent-theme")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), theme)

	// Test setting non-existent current theme
	err = suite.themeService.SetCurrentTheme("non-existent-theme")
	assert.Error(suite.T(), err)

	// Test validating nil theme
	result := suite.themeService.ValidateTheme(nil)
	assert.False(suite.T(), result.Valid)
	assert.NotEmpty(suite.T(), result.Errors)
}

// TestPerformanceWithLargeDataset tests performance with many themes
func (suite *ThemeTestSuite) TestPerformanceWithLargeDataset() {
	const numThemes = 100

	// Create many themes
	themes := make([]*models.LegacyTheme, numThemes)
	for i := 0; i < numThemes; i++ {
		themes[i] = &models.LegacyTheme{
			Name:        fmt.Sprintf("Performance Test Theme %d", i),
			DisplayName: fmt.Sprintf("Perf Theme %d", i),
			Description: fmt.Sprintf("A theme for performance testing #%d", i),
			Colors: models.LegacyColors{
				Primary:    fmt.Sprintf("#%02x%02x%02x", i%256, (i*2)%256, (i*3)%256),
				Background: "#0a0a0a",
				Text:       "#ffffff",
			},
		}
	}

	// Test batch conversion performance
	start := time.Now()
	results := suite.converter.ConvertBatch(themes)
	duration := time.Since(start)

	// Verify all conversions succeeded
	assert.Len(suite.T(), results, numThemes)
	for _, result := range results {
		assert.NoError(suite.T(), result.Error)
		assert.NotNil(suite.T(), result.Theme)
	}

	// Performance should be reasonable (less than 1 second for 100 themes)
	assert.Less(suite.T(), duration, time.Second)
}

// Run the test suite
func TestThemeSystem(t *testing.T) {
	suite.Run(t, new(ThemeTestSuite))
}

// BenchmarkThemeConversion benchmarks the legacy theme conversion performance
func BenchmarkThemeConversion(b *testing.B) {
	converter := theme.NewLegacyThemeConverter()
	legacyTheme := &models.LegacyTheme{
		Name: "Benchmark Theme",
		Colors: models.LegacyColors{
			Primary:    "#00ff41",
			Background: "#0a0a0a",
			Text:       "#ffffff",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.ConvertFromLegacy(legacyTheme)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCSSVariableGeneration benchmarks CSS variable generation performance
func BenchmarkCSSVariableGeneration(b *testing.B) {
	variableGen := theme.NewVariableGenerator()
	theme := models.NewTheme("benchmark", "Benchmark Theme")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := variableGen.GenerateCSSVariables(theme)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to create a comprehensive test theme
func createTestTheme(id, name string) *models.Theme {
	theme := models.NewTheme(id, name)
	theme.Description = "A comprehensive test theme"
	theme.Author = "Test Suite"
	theme.Version = "1.0.0"

	// Set all color properties
	theme.Colors.Background.Primary = "#0a0a0a"
	theme.Colors.Background.Secondary = "#1a1a1a"
	theme.Colors.Background.Tertiary = "#2a2a2a"
	theme.Colors.Foreground.Primary = "#ffffff"
	theme.Colors.Foreground.Secondary = "#cccccc"
	theme.Colors.Foreground.Tertiary = "#999999"
	theme.Colors.Accent.Primary = "#00ff41"
	theme.Colors.Accent.Secondary = "#00cc33"
	theme.Colors.Status.Success = "#00ff41"
	theme.Colors.Status.Warning = "#ffaa00"
	theme.Colors.Status.Error = "#ff3333"
	theme.Colors.Status.Info = "#00aaff"

	// Set terminal colors
	theme.Colors.Terminal = []string{
		"#000000", "#ff0000", "#00ff00", "#ffff00",
		"#0000ff", "#ff00ff", "#00ffff", "#ffffff",
		"#808080", "#ff8080", "#80ff80", "#ffff80",
		"#8080ff", "#ff80ff", "#80ffff", "#c0c0c0",
	}

	// Set UI colors
	theme.Colors.UI.ButtonBackground = "#1a1a1a"
	theme.Colors.UI.ButtonForeground = "#ffffff"
	theme.Colors.UI.ButtonHover = "#2a2a2a"
	theme.Colors.UI.ButtonActive = "#00ff41"
	theme.Colors.UI.InputBackground = "#0a0a0a"
	theme.Colors.UI.InputForeground = "#ffffff"
	theme.Colors.UI.InputBorder = "#333333"
	theme.Colors.UI.InputFocus = "#00ff41"
	theme.Colors.UI.Border = "#333333"
	theme.Colors.UI.Shadow = "rgba(0, 0, 0, 0.5)"

	// Set font properties
	theme.Fonts.Families.Primary = "\"Inter\", system-ui, sans-serif"
	theme.Fonts.Families.Secondary = "\"Roboto\", Arial, sans-serif"
	theme.Fonts.Families.Monospace = "\"JetBrains Mono\", \"Fira Code\", monospace"
	theme.Fonts.Sizes.ExtraSmall = "0.75rem"
	theme.Fonts.Sizes.Small = "0.875rem"
	theme.Fonts.Sizes.Base = "1rem"
	theme.Fonts.Sizes.Large = "1.125rem"
	theme.Fonts.Sizes.ExtraLarge = "1.25rem"
	theme.Fonts.Sizes.DoubleExtraLarge = "1.5rem"
	theme.Fonts.Weights.Light = "300"
	theme.Fonts.Weights.Normal = "400"
	theme.Fonts.Weights.Medium = "500"
	theme.Fonts.Weights.SemiBold = "600"
	theme.Fonts.Weights.Bold = "700"
	theme.Fonts.LineHeights.Tight = "1.25"
	theme.Fonts.LineHeights.Normal = "1.5"
	theme.Fonts.LineHeights.Relaxed = "1.75"
	theme.Fonts.LetterSpacing.Tight = "-0.025em"
	theme.Fonts.LetterSpacing.Normal = "0"
	theme.Fonts.LetterSpacing.Wide = "0.025em"

	// Set effects
	theme.Effects.BorderRadius.Small = "0.25rem"
	theme.Effects.BorderRadius.Medium = "0.5rem"
	theme.Effects.BorderRadius.Large = "1rem"
	theme.Effects.BorderRadius.ExtraLarge = "1.5rem"
	theme.Effects.BorderRadius.Full = "9999px"
	theme.Effects.Shadows = []string{
		"0 1px 3px rgba(0, 0, 0, 0.12), 0 1px 2px rgba(0, 0, 0, 0.24)",
		"0 4px 6px rgba(0, 0, 0, 0.16), 0 2px 4px rgba(0, 0, 0, 0.12)",
	}
	theme.Effects.Gradients = []string{
		"linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		"linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
	}
	theme.Effects.Animations.Duration.Fast = "150ms"
	theme.Effects.Animations.Duration.Normal = "300ms"
	theme.Effects.Animations.Duration.Slow = "500ms"
	theme.Effects.Animations.Easing.Linear = "linear"
	theme.Effects.Animations.Easing.EaseIn = "cubic-bezier(0.4, 0, 1, 1)"
	theme.Effects.Animations.Easing.EaseOut = "cubic-bezier(0, 0, 0.2, 1)"
	theme.Effects.Animations.Easing.EaseInOut = "cubic-bezier(0.4, 0, 0.2, 1)"
	theme.Effects.Blur.Small = "4px"
	theme.Effects.Blur.Medium = "8px"
	theme.Effects.Blur.Large = "16px"
	theme.Effects.Blur.ExtraLarge = "24px"

	// Set settings
	theme.Settings.DisabledOpacity = 0.5
	theme.Settings.HoverOpacity = 0.8
	theme.Settings.ActiveOpacity = 1
	theme.Settings.TransitionDuration.Fast = "150ms"
	theme.Settings.TransitionDuration.Normal = "300ms"
	theme.Settings.TransitionDuration.Slow = "500ms"
	theme.Settings.ZIndexScale = map[string]interface{}{
		"dropdown":      1000,
		"sticky":        1020,
		"fixed":         1030,
		"modalBackdrop": 1040,
		"modal":         1050,
		"popover":       1060,
		"tooltip":       1070,
		"toast":         1080,
	}

	return theme
}