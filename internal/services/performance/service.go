package performance

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"aDex/internal/logger"
)

// Service provides performance monitoring and optimization features
type Service struct {
	logger    *logger.Logger
	profiles  map[string]*Profile
	mutex     sync.RWMutex
	profiling bool
	startTime time.Time
}

// Profile represents a performance profile for a specific component
type Profile struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Enabled     bool               `json:"enabled"`
	StartTime   time.Time          `json:"startTime"`
	EndTime     time.Time          `json:"endTime,omitempty"`
	Duration    time.Duration      `json:"duration"`
	Samples     []Sample           `json:"samples"`
	Metrics     map[string]float64 `json:"metrics"`
	Hotspots    []Hotspot          `json:"hotspots"`
	Summary     *ProfileSummary    `json:"summary"`
	mutex       sync.Mutex         `json:"-"`
}

// Sample represents a single performance measurement
type Sample struct {
	Timestamp time.Time         `json:"timestamp"`
	Metric    string            `json:"metric"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	CallStack []Frame           `json:"callStack,omitempty"`
}

// Frame represents a stack frame in call stack
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

// Hotspot represents a performance bottleneck
type Hotspot struct {
	Location       string  `json:"location"`
	Function       string  `json:"function"`
	Impact         float64 `json:"impact"`
	TotalCalls     int     `json:"totalCalls"`
	AverageTime    float64 `json:"averageTime"`
	MaxTime        float64 `json:"maxTime"`
	Recommendation string  `json:"recommendation"`
}

// ProfileSummary provides aggregated performance statistics
type ProfileSummary struct {
	TotalSamples    int64              `json:"totalSamples"`
	TotalDuration   time.Duration      `json:"totalDuration"`
	AverageLatency  time.Duration      `json:"averageLatency"`
	P95Latency      time.Duration      `json:"p95Latency"`
	P99Latency      time.Duration      `json:"p99Latency"`
	MinLatency      time.Duration      `json:"minLatency"`
	MaxLatency      time.Duration      `json:"maxLatency"`
	MemoryUsage     int64              `json:"memoryUsage"`
	GoroutineCount  int                `json:"goroutineCount"`
	CPUUsage        float64            `json:"cpuUsage"`
	TopMetrics      map[string]float64 `json:"topMetrics"`
	Issues          []string           `json:"issues"`
	Recommendations []string           `json:"recommendations"`
}

// PerformanceIssue represents a detected performance problem
type PerformanceIssue struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Threshold   float64 `json:"threshold"`
	Resolution  string  `json:"resolution"`
}

// NewService creates a new performance service
func NewService(logger *logger.Logger) *Service {
	service := &Service{
		logger:    logger,
		profiles:  make(map[string]*Profile),
		profiling: false,
	}

	// Initialize default profiles
	service.initDefaultProfiles()

	return service
}

// StartProfiling begins performance profiling
func (s *Service) StartProfiling() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.profiling = true
	s.startTime = time.Now()

	s.logger.Info("Performance profiling started")

	// Start background profiling goroutine
	go s.backgroundProfiler()
}

// StopProfiling stops performance profiling
func (s *Service) StopProfiling() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.profiling {
		return
	}

	s.profiling = false

	s.logger.Info("Performance profiling stopped", map[string]interface{}{"duration": time.Since(s.startTime).String()})
}

// IsProfiling returns true if profiling is currently active
func (s *Service) IsProfiling() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.profiling
}

// CreateProfile creates a new performance profile
func (s *Service) CreateProfile(name, description string) *Profile {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	profile := &Profile{
		Name:        name,
		Description: description,
		Enabled:     true,
		StartTime:   time.Now(),
		Samples:     make([]Sample, 0),
		Metrics:     make(map[string]float64),
		Hotspots:    make([]Hotspot, 0),
		Summary: &ProfileSummary{
			TopMetrics:      make(map[string]float64),
			Issues:          make([]string, 0),
			Recommendations: make([]string, 0),
		},
	}

	s.profiles[name] = profile
	s.logger.Info("Created performance profile", map[string]interface{}{"name": name})

	return profile
}

// GetProfile returns a profile by name
func (s *Service) GetProfile(name string) *Profile {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if profile, exists := s.profiles[name]; exists {
		return profile
	}

	return nil
}

// ListProfiles returns all available profiles
func (s *Service) ListProfiles() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	names := make([]string, 0, len(s.profiles))
	for name := range s.profiles {
		names = append(names, name)
	}

	return names
}

// EnableProfile enables a profile for collection
func (s *Service) EnableProfile(name string) error {
	profile := s.GetProfile(name)
	if profile == nil {
		return fmt.Errorf("profile not found: %s", name)
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	profile.Enabled = true
	profile.StartTime = time.Now()

	s.logger.Info("Enabled performance profile", map[string]interface{}{"name": name})
	return nil
}

// DisableProfile disables a profile
func (s *Service) DisableProfile(name string) error {
	profile := s.GetProfile(name)
	if profile == nil {
		return fmt.Errorf("profile not found: %s", name)
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	profile.Enabled = false
	profile.EndTime = time.Now()
	profile.Duration = profile.EndTime.Sub(profile.StartTime)

	s.logger.Info("Disabled performance profile", map[string]interface{}{"name": name})
	return nil
}

// RecordSample records a performance sample
func (s *Service) RecordSample(profileName, metric string, value float64, labels map[string]string) error {
	profile := s.GetProfile(profileName)
	if profile == nil {
		return fmt.Errorf("profile not found: %s", profileName)
	}

	if !profile.Enabled {
		return nil
	}

	sample := Sample{
		Timestamp: time.Now(),
		Metric:    metric,
		Value:     value,
		Labels:    labels,
	}

	// Capture call stack if needed
	if metric == "function_call_time" {
		sample.CallStack = s.captureCallStack()
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	profile.Samples = append(profile.Samples, sample)
	profile.Metrics[metric] += value

	// Update hotspots
	s.updateHotspots(profile, sample)

	return nil
}

// RecordFunctionCall records a function call with timing
func (s *Service) RecordFunctionCall(profileName string, function string, duration time.Duration) error {
	labels := map[string]string{
		"function": function,
	}

	return s.RecordSample(profileName, "function_call_time", float64(duration.Nanoseconds())/1e6, labels)
}

// RecordMemoryUsage records memory usage
func (s *Service) RecordMemoryUsage(profileName string) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Record various memory metrics
	metrics := []struct {
		name   string
		value  float64
		labels map[string]string
	}{
		{"heap_alloc", float64(m.HeapAlloc), nil},
		{"heap_sys", float64(m.HeapSys), nil},
		{"heap_idle", float64(m.HeapIdle), nil},
		{"heap_inuse", float64(m.HeapInuse), nil},
		{"stack_inuse", float64(m.StackInuse), nil},
		{"gc_pause_total", float64(m.PauseTotalNs) / 1e6, nil},
	}

	for _, metric := range metrics {
		if err := s.RecordSample(profileName, metric.name, metric.value, metric.labels); err != nil {
			s.logger.Warn("Failed to record memory sample", map[string]interface{}{"metric": metric.name, "error": err.Error()})
		}
	}

	return nil
}

// RecordSystemMetrics records system-level metrics
func (s *Service) RecordSystemMetrics(profileName string) error {
	// Get goroutine count
	goroutines := runtime.NumGoroutine()

	// Record goroutine count
	if err := s.RecordSample(profileName, "goroutine_count", float64(goroutines), nil); err != nil {
		s.logger.Warn("Failed to record goroutine count", map[string]interface{}{"error": err.Error()})
	}

	// This would be extended to collect system metrics
	// using gopsutil or similar libraries
	return nil
}

// GetProfileSummary returns summary statistics for a profile
func (s *Service) GetProfileSummary(name string) (*ProfileSummary, error) {
	profile := s.GetProfile(name)
	if profile == nil {
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	summary := profile.Summary

	// Calculate summary statistics
	if len(profile.Samples) > 0 {
		summary.TotalSamples = int64(len(profile.Samples))

		// Calculate duration
		if !profile.EndTime.IsZero() {
			summary.TotalDuration = profile.EndTime.Sub(profile.StartTime)
		}

		// Calculate latency statistics
		latencies := make([]float64, 0)
		for _, sample := range profile.Samples {
			if sample.Metric == "function_call_time" || sample.Metric == "api_response_time" {
				latencies = append(latencies, sample.Value)
			}
		}

		if len(latencies) > 0 {
			sort.Float64s(latencies)
			summary.MinLatency = time.Duration(latencies[0]) * time.Nanosecond
			summary.MaxLatency = time.Duration(latencies[len(latencies)-1]) * time.Nanosecond

			// Calculate percentiles
			index := int(float64(len(latencies)) * 0.95)
			if index < len(latencies) {
				summary.P95Latency = time.Duration(latencies[index]) * time.Nanosecond
			}

			index = int(float64(len(latencies)) * 0.99)
			if index < len(latencies) {
				summary.P99Latency = time.Duration(latencies[index]) * time.Nanosecond
			}

			// Calculate average
			var sum float64
			for _, latency := range latencies {
				sum += latency
			}
			summary.AverageLatency = time.Duration(sum/float64(len(latencies))) * time.Nanosecond
		}

		// Update top metrics
		for metric, value := range profile.Metrics {
			summary.TopMetrics[metric] = value
		}
	}

	return summary, nil
}

// DetectPerformanceIssues analyzes profiles for performance problems
func (s *Service) DetectPerformanceIssues() []PerformanceIssue {
	var issues []PerformanceIssue

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, profile := range s.profiles {
		if !profile.Enabled {
			continue
		}

		profile.mutex.Lock()
		profileIssues := s.analyzeProfile(profile)
		profile.mutex.Unlock()

		issues = append(issues, profileIssues...)
	}

	return issues
}

// OptimizeBasedOnProfile provides optimization recommendations
func (s *Service) OptimizeBasedOnProfile(profileName string) ([]string, error) {
	profile := s.GetProfile(profileName)
	if profile == nil {
		return nil, fmt.Errorf("profile not found: %s", profileName)
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	var recommendations []string

	// Analyze hotspots
	for _, hotspot := range profile.Hotspots {
		if hotspot.AverageTime > 10.0 { // 10ms threshold
			recommendations = append(recommendations,
				fmt.Sprintf("Consider optimizing %s in %s (avg: %.2fms)",
					hotspot.Function, hotspot.Location, hotspot.AverageTime))
		}

		if hotspot.TotalCalls > 1000 && hotspot.AverageTime > 5.0 {
			recommendations = append(recommendations,
				fmt.Sprintf("Add caching to %s in %s (%d calls)",
					hotspot.Function, hotspot.Location, hotspot.TotalCalls))
		}
	}

	// Analyze memory usage
	if profile.Metrics["heap_inuse"] > 100*1024*1024 { // 100MB threshold
		recommendations = append(recommendations,
			"Consider reducing memory allocation in profile")
	}

	// Check goroutine count
	if profile.Metrics["goroutine_count"] > 100 {
		recommendations = append(recommendations,
			"Consider reducing goroutine usage to prevent memory leaks")
	}

	// Add profile-specific recommendations
	profile.Summary.Recommendations = recommendations

	return recommendations, nil
}

// GetPerformanceReport generates a comprehensive performance report
func (s *Service) GetPerformanceReport() map[string]interface{} {
	report := make(map[string]interface{})

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	report["timestamp"] = time.Now()
	report["profiling_active"] = s.profiling
	report["profiles"] = make(map[string]interface{})

	profiles := make(map[string]interface{})
	for name, profile := range s.profiles {
		profile.mutex.Lock()
		summary, _ := s.GetProfileSummary(name)
		profile.mutex.Unlock()

		profiles[name] = map[string]interface{}{
			"enabled":  profile.Enabled,
			"duration": profile.Duration,
			"samples":  len(profile.Samples),
			"summary":  summary,
			"hotspots": len(profile.Hotspots),
		}
	}
	report["profiles"] = profiles

	// Add system-wide metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	report["system"] = map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"heap_alloc": memStats.HeapAlloc,
		"heap_inuse": memStats.HeapInuse,
		"heap_sys":   memStats.HeapSys,
	}

	return report
}

// ClearProfile removes all samples from a profile
func (s *Service) ClearProfile(name string) error {
	profile := s.GetProfile(name)
	if profile == nil {
		return fmt.Errorf("profile not found: %s", name)
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	profile.Samples = make([]Sample, 0)
	profile.Metrics = make(map[string]float64)
	profile.Hotspots = make([]Hotspot, 0)
	profile.StartTime = time.Now()
	profile.EndTime = time.Time{}

	s.logger.Info("Cleared performance profile", map[string]interface{}{"name": name})
	return nil
}

// DeleteProfile removes a profile
func (s *Service) DeleteProfile(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.profiles[name]; exists {
		delete(s.profiles, name)
		s.logger.Info("Deleted performance profile", map[string]interface{}{"name": name})
		return nil
	}

	return fmt.Errorf("profile not found: %s", name)
}

// ExportProfile exports a profile to JSON
func (s *Service) ExportProfile(name string) ([]byte, error) {
	profile := s.GetProfile(name)
	if profile == nil {
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	profile.mutex.Lock()
	defer profile.mutex.Unlock()

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ImportProfile imports a profile from JSON
func (s *Service) ImportProfile(data []byte) error {
	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.profiles[profile.Name] = &profile
	s.logger.Info("Imported performance profile", map[string]interface{}{"name": profile.Name})

	return nil
}

// Private methods

// initializeDefaultProfiles sets up default performance profiles
func (s *Service) initDefaultProfiles() {
	defaultProfiles := []struct {
		name, description string
	}{
		{"terminal", "Terminal emulator performance"},
		{"system_monitor", "System monitoring performance"},
		{"file_browser", "File browser performance"},
		{"network_monitor", "Network monitoring performance"},
		{"ui", "UI responsiveness performance"},
		{"memory", "Memory allocation performance"},
	}

	for _, profile := range defaultProfiles {
		s.CreateProfile(profile.name, profile.description)
	}
}

// backgroundProfiler runs continuous profiling when profiling is enabled
func (s *Service) backgroundProfiler() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !s.IsProfiling() {
				return
			}

			// Record system metrics for all active profiles
			for _, profile := range s.profiles {
				if profile.Enabled {
					s.RecordSystemMetrics(profile.Name)
					s.RecordMemoryUsage(profile.Name)
				}
			}

		case <-time.After(10 * time.Second):
			// Cleanup any inactive profiles
			s.cleanupInactiveProfiles()
		}
	}
}

// cleanupInactiveProfiles disables profiles that haven't been updated recently
func (s *Service) cleanupInactiveProfiles() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for name, profile := range s.profiles {
		if !profile.Enabled {
			continue
		}

		profile.mutex.Lock()
		lastActivity := profile.EndTime
		if lastActivity.IsZero() {
			lastActivity = profile.StartTime
		}
		profile.mutex.Unlock()

		// Disable if inactive for more than 5 minutes
		if time.Since(lastActivity) > 5*time.Minute {
			s.DisableProfile(name)
		}
	}
}

// updateHotspots updates performance hotspots based on samples
func (s *Service) updateHotspots(profile *Profile, sample Sample) {
	if sample.Metric != "function_call_time" && sample.Metric != "api_response_time" {
		return
	}

	location := "unknown"
	if file, ok := sample.Labels["file"]; ok {
		location = file
	}

	function := "unknown"
	if fn, ok := sample.Labels["function"]; ok {
		function = fn
	}

	// Find existing hotspot
	var hotspot *Hotspot
	for i, h := range profile.Hotspots {
		if h.Location == location && h.Function == function {
			hotspot = &profile.Hotspots[i]
			break
		}
	}

	if hotspot == nil {
		// Create new hotspot
		hotspot = &Hotspot{
			Location:    location,
			Function:    function,
			Impact:      0,
			TotalCalls:  0,
			AverageTime: sample.Value,
			MaxTime:     sample.Value,
		}
		profile.Hotspots = append(profile.Hotspots, *hotspot)
	}

	// Update hotspot statistics
	hotspot.TotalCalls++
	hotspot.AverageTime = (hotspot.AverageTime*float64(hotspot.TotalCalls-1) + sample.Value) / float64(hotspot.TotalCalls)
	hotspot.MaxTime = max(hotspot.MaxTime, sample.Value)

	// Update impact based on frequency and duration
	hotspot.Impact = float64(hotspot.TotalCalls) * hotspot.AverageTime

	// Sort hotspots by impact
	sort.Slice(profile.Hotspots, func(i, j int) bool {
		return profile.Hotspots[i].Impact > profile.Hotspots[j].Impact
	})
}

// analyzeProfile analyzes a profile for performance issues
func (s *Service) analyzeProfile(profile *Profile) []PerformanceIssue {
	var issues []PerformanceIssue

	// Check for high memory usage
	if heapInuse, ok := profile.Metrics["heap_inuse"]; ok && heapInuse > 200*1024*1024 {
		issues = append(issues, PerformanceIssue{
			Type:        "memory",
			Severity:    "warning",
			Description: "High memory usage detected",
			Location:    profile.Name,
			Metric:      "heap_inuse",
			Value:       heapInuse,
			Threshold:   200 * 1024 * 1024,
			Resolution:  "Consider optimizing memory allocation patterns",
		})
	}

	// Check for too many goroutines
	if goroutineCount, ok := profile.Metrics["goroutine_count"]; ok && goroutineCount > 500 {
		issues = append(issues, PerformanceIssue{
			Type:        "concurrency",
			Severity:    "warning",
			Description: "High goroutine count detected",
			Location:    profile.Name,
			Metric:      "goroutine_count",
			Value:       goroutineCount,
			Threshold:   500,
			Resolution:  "Check for goroutine leaks",
		})
	}

	return issues
}

// captureCallStack captures the current goroutine's call stack
func (s *Service) captureCallStack() []Frame {
	pc := make([]uintptr, 32)
	n := runtime.Callers(2, pc)
	pc = pc[:n]

	frames := make([]Frame, 0, n)
	for _, pcVal := range pc {
		fn := runtime.FuncForPC(pcVal)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pcVal)
		frames = append(frames, Frame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
			Package:  extractPkgName(fn.Name()),
		})
	}

	return frames
}

// extractPkgName extracts the package name from a fully qualified function name
func extractPkgName(funcName string) string {
	// funcName is like "github.com/user/repo/pkg/subpkg.FuncName"
	lastSlash := -1
	for i := len(funcName) - 1; i >= 0; i-- {
		if funcName[i] == '/' {
			lastSlash = i
			break
		}
	}
	if lastSlash == -1 {
		return ""
	}
	afterSlash := funcName[lastSlash+1:]
	dotIndex := -1
	for i, c := range afterSlash {
		if c == '.' {
			dotIndex = i
			break
		}
	}
	if dotIndex == -1 {
		return funcName[:lastSlash+1+len(afterSlash)]
	}
	return funcName[:lastSlash+1+dotIndex]
}

// max returns the maximum of two values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
