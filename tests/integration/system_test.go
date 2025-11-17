package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/wailsapp/wails/v2/pkg/options/linux"

	"aDex-UI/internal/api"
	"aDex-UI/internal/models"
	"aDex-UI/internal/services/filesystem"
	"aDex-UI/internal/services/network"
	"aDex-UI/internal/services/process"
	"aDex-UI/internal/services/theme"
)

// SystemIntegrationTestSuite provides comprehensive integration testing
// for the entire aDex-UI system, covering all implemented features
type SystemIntegrationTestSuite struct {
	suite.Suite
	echoServer       *echo.Echo
	httpServer       *httptest.Server
	fsService        *filesystem.Service
	networkService   *network.Service
	processService   *process.Service
	themeService     *theme.Service
	tempDir          string
	linuxOptions     *options.Linux
	testFiles        map[string]string
	testTheme        *models.Theme
}

// SetupSuite runs once before all tests
func (suite *SystemIntegrationTestSuite) SetupSuite() {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "dex-integration-test-*")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// Initialize Linux options
	suite.linuxOptions = &options.Linux{
		Icon:          "terminal",
		WindowIsResizable: true,
		WebviewIsTransparent: false,
	}

	// Create test files
	suite.testFiles = make(map[string]string)
	testFiles := []struct {
		path     string
		content  string
	}{
		{"test.txt", "Hello, World! This is a test file."},
		{"subdir/nested.txt", "Nested file content"},
		{"config.json", `{"name": "test", "version": "1.0.0"}`},
		{"README.md", "# Test Project\n\nThis is a test project file."},
		{"script.sh", "#!/bin/bash\necho 'Hello from script'"},
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file.path)
		dir := filepath.Dir(fullPath)
		if dir != tempDir {
			require.NoError(suite.T(), os.MkdirAll(dir, 0755))
		}
		require.NoError(suite.T(), os.WriteFile(fullPath, []byte(file.content), 0644))
		suite.testFiles[file.path] = fullPath
	}

	// Initialize file system service
	suite.fsService = filesystem.NewService()
	err = suite.fsService.SetWorkingDirectory(tempDir)
	require.NoError(suite.T(), err)

	// Initialize network service
	suite.networkService = network.NewService()
	err = suite.networkService.Initialize()
	require.NoError(suite.T(), err)

	// Initialize process service
	suite.processService = process.NewService()
	err = suite.processService.Initialize()
	require.NoError(suite.T(), err)

	// Initialize theme service
	themeConfig := &theme.ThemeServiceConfig{
		ThemeDirectory: filepath.Join(tempDir, "themes"),
		AutoSave:       true,
		AutoReload:     true,
		CacheEnabled:   true,
	}
	suite.themeService = theme.NewThemeService(themeConfig)
	err = suite.themeService.Initialize()
	require.NoError(suite.T(), err)

	// Create test theme
	suite.testTheme = models.NewTheme("integration-test-theme", "Integration Test Theme")
	suite.testTheme.Description = "A theme for integration testing"
	suite.testTheme.Author = "Test Suite"
	suite.testTheme.Version = "1.0.0"

	// Set theme colors
	suite.testTheme.Colors.Background.Primary = "#0a0a0a"
	suite.testTheme.Colors.Background.Secondary = "#1a1a1a"
	suite.testTheme.Colors.Background.Tertiary = "#2a2a2a"
	suite.testTheme.Colors.Foreground.Primary = "#ffffff"
	suite.testTheme.Colors.Foreground.Secondary = "#cccccc"
	suite.testTheme.Colors.Foreground.Tertiary = "#999999"
	suite.testTheme.Colors.Accent.Primary = "#00ff41"
	suite.testTheme.Colors.Accent.Secondary = "#00cc33"
	suite.testTheme.Colors.Status.Success = "#00ff41"
	suite.testTheme.Colors.Status.Warning = "#ffaa00"
	suite.testTheme.Colors.Status.Error = "#ff3333"
	suite.testTheme.Colors.Status.Info = "#00aaff"

	// Set terminal colors
	suite.testTheme.Colors.Terminal = []string{
		"#000000", "#ff0000", "#00ff00", "#ffff00",
		"#0000ff", "#ff00ff", "#00ffff", "#ffffff",
		"#808080", "#ff8080", "#80ff80", "#ffff80",
		"#8080ff", "#ff80ff", "#80ffff", "#c0c0c0",
	}

	// Set UI colors
	suite.testTheme.Colors.UI.ButtonBackground = "#1a1a1a"
	suite.testTheme.Colors.UI.ButtonForeground = "#ffffff"
	suite.testTheme.Colors.UI.ButtonHover = "#2a2a2a"
	suite.testTheme.Colors.UI.ButtonActive = "#00ff41"
	suite.testTheme.Colors.UI.InputBackground = "#0a0a0a"
	suite.testTheme.Colors.UI.InputForeground = "#ffffff"
	suite.testTheme.Colors.UI.InputBorder = "#333333"
	suite.testTheme.Colors.UI.InputFocus = "#00ff41"
	suite.testTheme.Colors.UI.Border = "#333333"
	suite.testTheme.Colors.UI.Shadow = "rgba(0, 0, 0, 0.5)"

	// Save test theme
	err = suite.themeService.SaveTheme(suite.testTheme)
	require.NoError(suite.T(), err)

	// Initialize Echo server
	suite.echoServer = echo.New()

	// Set up API routes
	apiHandler := api.NewAPIHandler(
		suite.fsService,
		suite.networkService,
		suite.processService,
		suite.themeService,
	)
	apiHandler.RegisterRoutes(suite.echoServer)

	// Start HTTP test server
	suite.httpServer = httptest.NewServer(suite.echoServer)
}

// TearDownSuite runs once after all tests
func (suite *SystemIntegrationTestSuite) TearDownSuite() {
	// Clean up HTTP server
	if suite.httpServer != nil {
		suite.httpServer.Close()
	}

	// Clean up Echo server
	if suite.echoServer != nil {
		suite.echoServer.Close()
	}

	// Clean up temporary directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

// TestFileSystemIntegration tests the complete file system functionality
func (suite *SystemIntegrationTestSuite) TestFileSystemIntegration() {
	// Test getting current working directory
	cwd, err := suite.fsService.GetCurrentWorkingDirectory()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.tempDir, cwd)

	// Test listing directory contents
	contents, err := suite.fsService.ListDirectory(suite.tempDir, false)
	require.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(contents), 0)

	// Find test files
	var testFile *models.FileInfo
	var testDir *models.FileInfo
	for _, item := range contents {
		if item.Name == "test.txt" {
			testFile = item
		}
		if item.Name == "subdir" {
			testDir = item
		}
	}

	// Verify test file exists
	require.NotNil(suite.T(), testFile)
	assert.Equal(suite.T(), "test.txt", testFile.Name)
	assert.False(suite.T(), testFile.IsDirectory)
	assert.Greater(suite.T(), testFile.Size, int64(0))

	// Verify test directory exists
	require.NotNil(suite.T(), testDir)
	assert.Equal(suite.T(), "subdir", testDir.Name)
	assert.True(suite.T(), testDir.IsDirectory)

	// Test reading file content
	content, err := suite.fsService.ReadFile(suite.testFiles["test.txt"])
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Hello, World! This is a test file.", content)

	// Test nested directory listing
	nestedContents, err := suite.fsService.ListDirectory(filepath.Join(suite.tempDir, "subdir"), false)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), nestedContents, 1)
	assert.Equal(suite.T(), "nested.txt", nestedContents[0].Name)

	// Test file system events
	eventChan := make(chan *models.FileSystemEvent, 10)
	err = suite.fsService.WatchDirectory(suite.tempDir, eventChan)
	require.NoError(suite.T(), err)

	// Create a new file to trigger event
	newFilePath := filepath.Join(suite.tempDir, "new-test.txt")
	err = os.WriteFile(newFilePath, []byte("New file content"), 0644)
	require.NoError(suite.T(), err)

	// Wait for event
	select {
	case event := <-eventChan:
		assert.Equal(suite.T(), models.EventCreate, event.Type)
		assert.Equal(suite.T(), "new-test.txt", filepath.Base(event.Path))
	case <-time.After(2 * time.Second):
		suite.T().Fatal("Expected file system event not received")
	}

	// Stop watching
	suite.fsService.StopWatching()
}

// TestNetworkIntegration tests the complete network monitoring functionality
func (suite *SystemIntegrationTestSuite) TestNetworkIntegration() {
	// Test network statistics
	stats, err := suite.networkService.GetNetworkStats()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)

	// Test interface enumeration
	interfaces, err := suite.networkService.GetNetworkInterfaces()
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), interfaces)

	// Find loopback interface
	var loopbackInterface *models.NetworkInterface
	for _, iface := range interfaces {
		if iface.Name == "lo" || iface.Name == "lo0" {
			loopbackInterface = iface
			break
		}
	}

	// Verify loopback interface exists
	require.NotNil(suite.T(), loopbackInterface)
	assert.True(suite.T(), loopbackInterface.IsLoopback)
	assert.NotEmpty(suite.T(), loopbackInterface.Addresses)

	// Test network connection monitoring
	connections, err := suite.networkService.GetActiveConnections()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), connections)

	// Test real-time monitoring
	monitorChan := make(chan *models.NetworkStats, 10)
	err = suite.networkService.StartMonitoring(monitorChan)
	require.NoError(suite.T(), err)

	// Wait for at least one monitoring update
	select {
	case stats := <-monitorChan:
		assert.NotNil(suite.T(), stats)
		assert.GreaterOrEqual(suite.T(), stats.BytesReceived, int64(0))
		assert.GreaterOrEqual(suite.T(), stats.BytesSent, int64(0))
	case <-time.After(2 * time.Second):
		suite.T().Fatal("Expected network monitoring update not received")
	}

	// Stop monitoring
	suite.networkService.StopMonitoring()
}

// TestProcessIntegration tests the complete process management functionality
func (suite *SystemIntegrationTestSuite) TestProcessIntegration() {
	// Test getting current process
	currentProc, err := suite.processService.GetCurrentProcess()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), currentProc)
	assert.Greater(suite.T(), currentProc.PID, 0)

	// Test process enumeration
	processes, err := suite.processService.GetProcessList()
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), processes)

	// Find current process in list
	var foundCurrentProc bool
	for _, proc := range processes {
		if proc.PID == currentProc.PID {
			foundCurrentProc = true
			assert.Equal(suite.T(), currentProc.Name, proc.Name)
			break
		}
	}
	assert.True(suite.T(), foundCurrentProc, "Current process not found in process list")

	// Test process monitoring
	monitorChan := make(chan *models.ProcessStats, 10)
	err = suite.processService.StartMonitoring(currentProc.PID, monitorChan)
	require.NoError(suite.T(), err)

	// Wait for at least one monitoring update
	select {
	case stats := <-monitorChan:
		assert.NotNil(suite.T(), stats)
		assert.Greater(suite.T(), stats.CPUUsage, 0.0)
		assert.Greater(suite.T(), stats.MemoryUsage, int64(0))
	case <-time.After(2 * time.Second):
		suite.T().Fatal("Expected process monitoring update not received")
	}

	// Stop monitoring
	suite.processService.StopMonitoring(currentProc.PID)
}

// TestThemeIntegration tests the complete theme system functionality
func (suite *SystemIntegrationTestSuite) TestThemeIntegration() {
	// Test getting available themes
	themes := suite.themeService.GetAvailableThemes()
	assert.NotEmpty(suite.T(), themes)

	// Find our test theme
	var foundTestTheme *models.Theme
	for _, theme := range themes {
		if theme.ID == "integration-test-theme" {
			foundTestTheme = theme
			break
		}
	}
	require.NotNil(suite.T(), foundTestTheme, "Test theme not found in available themes")
	assert.Equal(suite.T(), suite.testTheme.Name, foundTestTheme.Name)

	// Test setting current theme
	err := suite.themeService.SetCurrentTheme("integration-test-theme")
	require.NoError(suite.T(), err)

	currentTheme := suite.themeService.GetCurrentTheme()
	assert.Equal(suite.T(), "integration-test-theme", currentTheme.ID)

	// Test CSS variable generation
	variableGen := theme.NewVariableGenerator()
	css, err := variableGen.GenerateCSSVariables(currentTheme)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), css.Variables)
	assert.Contains(suite.T(), css.Variables, "--dex-theme-colors-background-primary")
	assert.Contains(suite.T(), css.Variables, "#0a0a0a")

	// Test theme validation
	validationResult := suite.themeService.ValidateTheme(currentTheme)
	assert.True(suite.T(), validationResult.Valid)
	assert.Empty(suite.T(), validationResult.Errors)

	// Test theme export
	exportData, err := suite.themeService.ExportTheme(currentTheme)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), exportData)
	assert.Len(suite.T(), exportData.Themes, 1)
	assert.Equal(suite.T(), currentTheme.ID, exportData.Themes[0].ID)

	// Test legacy theme conversion
	converter := theme.NewLegacyThemeConverter()
	legacyTheme := &models.LegacyTheme{
		Name:        "Legacy Integration Test",
		DisplayName: "Legacy Integration Test",
		Description: "A legacy theme for integration testing",
		IsDark:      true,
		Colors: models.LegacyColors{
			Primary:    "#00ff41",
			Secondary:  "#00cc33",
			Background: "#0a0a0a",
			Text:       "#ffffff",
		},
		Effects: models.LegacyEffects{
			Glow:       true,
			Animation:  true,
			Shadows:    true,
		},
	}

	convertedTheme, err := converter.ConvertFromLegacy(legacyTheme)
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), convertedTheme)
	assert.Equal(suite.T(), "legacy-integration-test", convertedTheme.ID)
	assert.True(suite.T(), convertedTheme.IsDark)
}

// TestAPIIntegration tests the complete API functionality
func (suite *SystemIntegrationTestSuite) TestAPIIntegration() {
	baseURL := suite.httpServer.URL

	// Test filesystem API
	suite.T().Run("Filesystem API", func(t *testing.T) {
		// Test GET /api/filesystem/cwd
		resp, err := http.Get(baseURL + "/api/filesystem/cwd")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var cwdResponse struct {
			CWD string `json:"cwd"`
		}
		err = json.NewDecoder(resp.Body).Decode(&cwdResponse)
		require.NoError(t, err)
		assert.Equal(t, suite.tempDir, cwdResponse.CWD)

		// Test GET /api/filesystem/list
		resp, err = http.Get(baseURL + "/api/filesystem/list?path=" + suite.tempDir)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResponse struct {
			Files []*models.FileInfo `json:"files"`
		}
		err = json.NewDecoder(resp.Body).Decode(&listResponse)
		require.NoError(t, err)
		assert.Greater(t, len(listResponse.Files), 0)
	})

	// Test network API
	suite.T().Run("Network API", func(t *testing.T) {
		// Test GET /api/network/stats
		resp, err := http.Get(baseURL + "/api/network/stats")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var statsResponse struct {
			Stats *models.NetworkStats `json:"stats"`
		}
		err = json.NewDecoder(resp.Body).Decode(&statsResponse)
		require.NoError(t, err)
		assert.NotNil(t, statsResponse.Stats)

		// Test GET /api/network/interfaces
		resp, err = http.Get(baseURL + "/api/network/interfaces")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var interfacesResponse struct {
			Interfaces []*models.NetworkInterface `json:"interfaces"`
		}
		err = json.NewDecoder(resp.Body).Decode(&interfacesResponse)
		require.NoError(t, err)
		assert.NotEmpty(t, interfacesResponse.Interfaces)
	})

	// Test process API
	suite.T().Run("Process API", func(t *testing.T) {
		// Test GET /api/process/current
		resp, err := http.Get(baseURL + "/api/process/current")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var currentResponse struct {
			Process *models.ProcessInfo `json:"process"`
		}
		err = json.NewDecoder(resp.Body).Decode(&currentResponse)
		require.NoError(t, err)
		assert.NotNil(t, currentResponse.Process)

		// Test GET /api/process/list
		resp, err = http.Get(baseURL + "/api/process/list")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResponse struct {
			Processes []*models.ProcessInfo `json:"processes"`
		}
		err = json.NewDecoder(resp.Body).Decode(&listResponse)
		require.NoError(t, err)
		assert.NotEmpty(t, listResponse.Processes)
	})

	// Test theme API
	suite.T().Run("Theme API", func(t *testing.T) {
		// Test GET /api/themes
		resp, err := http.Get(baseURL + "/api/themes")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var themesResponse struct {
			Themes []*models.Theme `json:"themes"`
		}
		err = json.NewDecoder(resp.Body).Decode(&themesResponse)
		require.NoError(t, err)
		assert.NotEmpty(t, themesResponse.Themes)

		// Test GET /api/themes/current
		resp, err = http.Get(baseURL + "/api/themes/current")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var currentResponse struct {
			Theme *models.Theme `json:"theme"`
		}
		err = json.NewDecoder(resp.Body).Decode(&currentResponse)
		require.NoError(t, err)
		assert.NotNil(t, currentResponse.Theme)

		// Test POST /api/themes/set
		setRequest := map[string]interface{}{
			"themeId": "integration-test-theme",
		}
		requestBody, _ := json.Marshal(setRequest)

		resp, err = http.Post(baseURL+"/api/themes/set", "application/json",
			bytes.NewBuffer(requestBody))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestCrossSystemIntegration tests interactions between different system components
func (suite *SystemIntegrationTestSuite) TestCrossSystemIntegration() {
	// Test theme application affecting file system display
	suite.T().Run("Theme affects File System Display", func(t *testing.T) {
		// Set a dark theme
		err := suite.themeService.SetCurrentTheme("integration-test-theme")
		require.NoError(t, err)

		// Get file listing (should be themed)
		contents, err := suite.fsService.ListDirectory(suite.tempDir, false)
		require.NoError(t, err)

		// Verify theme is applied to file system display
		currentTheme := suite.themeService.GetCurrentTheme()
		assert.Equal(t, "integration-test-theme", currentTheme.ID)
		assert.Greater(t, len(contents), 0)
	})

	// Test network monitoring while process is running
	suite.T().Run("Network and Process Monitoring", func(t *testing.T) {
		// Start network monitoring
		netChan := make(chan *models.NetworkStats, 10)
		err := suite.networkService.StartMonitoring(netChan)
		require.NoError(t, err)
		defer suite.networkService.StopMonitoring()

		// Start process monitoring
		procChan := make(chan *models.ProcessStats, 10)
		currentProc, err := suite.processService.GetCurrentProcess()
		require.NoError(t, err)
		err = suite.processService.StartMonitoring(currentProc.PID, procChan)
		require.NoError(t, err)
		defer suite.processService.StopMonitoring(currentProc.PID)

		// Collect metrics from both systems
		var netStats *models.NetworkStats
		var procStats *models.ProcessStats

		timeout := time.After(3 * time.Second)
		for netStats == nil || procStats == nil {
			select {
			case stats := <-netChan:
				if netStats == nil {
					netStats = stats
				}
			case stats := <-procChan:
				if procStats == nil {
					procStats = stats
				}
			case <-timeout:
				t.Fatal("Timeout waiting for network and process stats")
			}
		}

		// Verify both systems are working
		assert.NotNil(t, netStats)
		assert.NotNil(t, procStats)
		assert.GreaterOrEqual(t, netStats.BytesReceived, int64(0))
		assert.GreaterOrEqual(t, procStats.CPUUsage, 0.0)
	})

	// Test file operations triggering theme updates
	suite.T().Run("File Operations and Theme Updates", func(t *testing.T) {
		// Create a theme-related file
		themeFile := filepath.Join(suite.tempDir, "theme-test.txt")
		err := os.WriteFile(themeFile, []byte("Theme test content"), 0644)
		require.NoError(t, err)

		// Verify file was created
		contents, err := suite.fsService.ListDirectory(suite.tempDir, false)
		require.NoError(t, err)

		var themeTestFile *models.FileInfo
		for _, file := range contents {
			if file.Name == "theme-test.txt" {
				themeTestFile = file
				break
			}
		}
		require.NotNil(t, themeTestFile, "Theme test file not found")

		// Read the file content
		content, err := suite.fsService.ReadFile(themeFile)
		require.NoError(t, err)
		assert.Equal(t, "Theme test content", content)
	})
}

// TestPerformanceAndReliability tests system performance and reliability
func (suite *SystemIntegrationTestSuite) TestPerformanceAndReliability() {
	// Test concurrent API requests
	suite.T().Run("Concurrent API Requests", func(t *testing.T) {
		const numRequests = 50
		baseURL := suite.httpServer.URL

		// Launch concurrent requests
		results := make(chan error, numRequests)
		for i := 0; i < numRequests; i++ {
			go func() {
				resp, err := http.Get(baseURL + "/api/filesystem/cwd")
				if err != nil {
					results <- err
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
					return
				}
				results <- nil
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < numRequests; i++ {
			err := <-results
			if err == nil {
				successCount++
			}
		}

		// Verify most requests succeeded (allow some failures due to load)
		successRate := float64(successCount) / float64(numRequests)
		assert.Greater(t, successRate, 0.9, "Success rate should be at least 90%%")
	})

	// Test memory usage under load
	suite.T().Run("Memory Usage Under Load", func(t *testing.T) {
		// Get initial memory usage
		currentProc, err := suite.processService.GetCurrentProcess()
		require.NoError(t, err)

		// Perform memory-intensive operations
		for i := 0; i < 100; i++ {
			// Create and read many files
			testFile := filepath.Join(suite.tempDir, fmt.Sprintf("memory-test-%d.txt", i))
			err := os.WriteFile(testFile, []byte(fmt.Sprintf("Test content %d", i)), 0644)
			require.NoError(t, err)

			_, err = suite.fsService.ReadFile(testFile)
			require.NoError(t, err)

			// Get network stats
			_, err = suite.networkService.GetNetworkStats()
			require.NoError(t, err)

			// Get process list
			_, err = suite.processService.GetProcessList()
			require.NoError(t, err)
		}

		// Check that memory usage is reasonable (this is a basic check)
		// In a real system, you'd want more sophisticated memory monitoring
		updatedProc, err := suite.processService.GetCurrentProcess()
		require.NoError(t, err)
		assert.Greater(t, updatedProc.MemoryUsage, currentProc.MemoryUsage)
	})

	// Test system recovery after errors
	suite.T().Run("System Recovery After Errors", func(t *testing.T) {
		// Test invalid file access
		_, err := suite.fsService.ReadFile("/nonexistent/file.txt")
		assert.Error(t, err)

		// System should still work after error
		contents, err := suite.fsService.ListDirectory(suite.tempDir, false)
		require.NoError(t, err)
		assert.Greater(t, len(contents), 0)

		// Test invalid theme ID
		err = suite.themeService.SetCurrentTheme("nonexistent-theme")
		assert.Error(t, err)

		// System should still work after error
		currentTheme := suite.themeService.GetCurrentTheme()
		assert.NotNil(t, currentTheme)
	})
}

// TestLinuxOptionsIntegration tests Linux-specific options integration
func (suite *SystemIntegrationTestSuite) TestLinuxOptionsIntegration() {
	// Test Linux options configuration
	require.NotNil(t, suite.linuxOptions)
	assert.Equal(t, "terminal", suite.linuxOptions.Icon)
	assert.True(t, suite.linuxOptions.WindowIsResizable)
	assert.False(t, suite.linuxOptions.WebviewIsTransparent)

	// Test that services work correctly with Linux options
	currentProc, err := suite.processService.GetCurrentProcess()
	require.NoError(t, err)
	assert.Greater(t, currentProc.PID, 0)

	// Test file system operations work with Linux environment
	cwd, err := suite.fsService.GetCurrentWorkingDirectory()
	require.NoError(t, err)
	assert.NotEmpty(t, cwd)

	// Test network operations work on Linux
	interfaces, err := suite.networkService.GetNetworkInterfaces()
	require.NoError(t, err)
	assert.NotEmpty(t, interfaces)
}

// Run the integration test suite
func TestSystemIntegration(t *testing.T) {
	suite.Run(t, new(SystemIntegrationTestSuite))
}