package geoip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"aDex/internal/events"
)

// GeoIPService provides geographical IP information
type GeoIPService struct {
	eventBus      events.IEventBus
	config        *GeoIPConfig
	cache         map[string]*GeoIPData
	mutex         sync.RWMutex
	isEnabled     bool
	lastUpdate    time.Time
	client        *http.Client
}

// GeoIPConfig contains configuration for the GeoIP service
type GeoIPConfig struct {
	Enabled           bool          `json:"enabled"`
	CacheTimeout      time.Duration `json:"cacheTimeout"`
	RequestTimeout    time.Duration `json:"requestTimeout"`
	MaxCacheSize      int           `json:"maxCacheSize"`
	Provider          string        `json:"provider"` // "ipapi", "ipapi.co", "maxmind"
	APIKey           string        `json:"apiKey,omitempty"`
	UpdateInterval    time.Duration `json:"updateInterval"`
}

// GeoIPData contains geographical information for an IP address
type GeoIPData struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionCode  string  `json:"regionCode"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
	Organization string `json:"organization"`
	AS          string  `json:"as"`
	Timezone    string  `json:"timezone"`
	IsVPN       bool    `json:"isVpn"`
	IsProxy     bool    `json:"isProxy"`
	IsMobile    bool    `json:"isMobile"`
	Timestamp   time.Time `json:"timestamp"`
}

// GeoIPProvider interface for different GeoIP providers
type GeoIPProvider interface {
	GetGeoIPData(ctx context.Context, ip string) (*GeoIPData, error)
}

// IPApiCoProvider implements GeoIP provider using ipapi.co
type IPApiCoProvider struct {
	client *http.Client
}

// IPApiProvider implements GeoIP provider using ip-api.com
type IPApiProvider struct {
	client *http.Client
}

// NewGeoIPService creates a new GeoIP service instance
func NewGeoIPService(eventBus events.IEventBus) *GeoIPService {
	config := &GeoIPConfig{
		Enabled:        false, // Disabled by default for privacy
		CacheTimeout:   24 * time.Hour,
		RequestTimeout: 5 * time.Second,
		MaxCacheSize:   1000,
		Provider:       "ipapi.co", // Free tier
		UpdateInterval: 1 * time.Hour,
	}

	service := &GeoIPService{
		eventBus:   eventBus,
		config:     config,
		cache:      make(map[string]*GeoIPData),
		isEnabled:  config.Enabled,
		client: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}

	return service
}

// SetConfig updates the GeoIP service configuration
func (s *GeoIPService) SetConfig(config *GeoIPConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config = config
	s.isEnabled = config.Enabled
	s.client.Timeout = config.RequestTimeout

	// Clear cache if provider changed
	if config.Provider != s.config.Provider {
		s.cache = make(map[string]*GeoIPData)
	}

	return nil
}

// GetConfig returns the current GeoIP configuration
func (s *GeoIPService) GetConfig() *GeoIPConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modification
	configCopy := *s.config
	return &configCopy
}

// IsEnabled returns whether GeoIP service is enabled
func (s *GeoIPService) IsEnabled() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isEnabled
}

// Enable enables the GeoIP service
func (s *GeoIPService) Enable() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.isEnabled = true
	s.config.Enabled = true

	// Emit event
	if s.eventBus != nil {
		s.eventBus.Publish(context.Background(), "geoip:enabled", map[string]interface{}{
			"timestamp": time.Now(),
		}, "geoip")
	}

	return nil
}

// Disable disables the GeoIP service and clears cache
func (s *GeoIPService) Disable() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.isEnabled = false
	s.config.Enabled = false
	s.cache = make(map[string]*GeoIPData) // Clear cache for privacy

	// Emit event
	if s.eventBus != nil {
		s.eventBus.Publish(context.Background(), "geoip:disabled", map[string]interface{}{
			"timestamp": time.Now(),
		}, "geoip")
	}

	return nil
}

// GetGeoIPData retrieves geographical information for an IP address
func (s *GeoIPService) GetGeoIPData(ctx context.Context, ip string) (*GeoIPData, error) {
	// Check if service is enabled
	if !s.IsEnabled() {
		return nil, fmt.Errorf("GeoIP service is disabled")
	}

	// Validate IP address
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Skip private IP addresses
	if s.isPrivateIP(ip) {
		return nil, fmt.Errorf("GeoIP lookup not available for private IP addresses: %s", ip)
	}

	s.mutex.RLock()
	// Check cache first
	if cached, exists := s.cache[ip]; exists {
		if time.Since(cached.Timestamp) < s.config.CacheTimeout {
			s.mutex.RUnlock()
			return cached, nil
		}
		// Cache expired, remove it
		delete(s.cache, ip)
	}
	s.mutex.RUnlock()

	// Fetch fresh data
	data, err := s.fetchGeoIPData(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GeoIP data for %s: %w", ip, err)
	}

	// Cache the result
	s.mutex.Lock()
	s.cache[ip] = data
	s.lastUpdate = time.Now()
	s.mutex.Unlock()

	// Emit event
	if s.eventBus != nil {
		s.eventBus.Publish(context.Background(), "geoip:data-updated", map[string]interface{}{
			"ip":        ip,
			"data":      data,
			"timestamp": time.Now(),
		}, "geoip")
	}

	return data, nil
}

// GetCachedGeoIPData retrieves cached geographical information for an IP address
func (s *GeoIPService) GetCachedGeoIPData(ip string) (*GeoIPData, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("GeoIP service is disabled")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cached, exists := s.cache[ip]
	if !exists {
		return nil, fmt.Errorf("no cached data for IP: %s", ip)
	}

	if time.Since(cached.Timestamp) >= s.config.CacheTimeout {
		return nil, fmt.Errorf("cached data expired for IP: %s", ip)
	}

	return cached, nil
}

// ClearCache clears the GeoIP cache
func (s *GeoIPService) ClearCache() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cache = make(map[string]*GeoIPData)

	// Emit event
	if s.eventBus != nil {
		s.eventBus.Publish(context.Background(), "geoip:cache-cleared", map[string]interface{}{
			"timestamp": time.Now(),
		}, "geoip")
	}

	return nil
}

// GetCacheStats returns cache statistics
func (s *GeoIPService) GetCacheStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"size":       len(s.cache),
		"maxSize":    s.config.MaxCacheSize,
		"lastUpdate": s.lastUpdate,
		"enabled":    s.isEnabled,
	}
}

// CleanupExpiredCache removes expired entries from cache
func (s *GeoIPService) CleanupExpiredCache() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for ip, data := range s.cache {
		if time.Since(data.Timestamp) >= s.config.CacheTimeout {
			delete(s.cache, ip)
		}
	}
}

// fetchGeoIPData fetches GeoIP data from the configured provider
func (s *GeoIPService) fetchGeoIPData(ctx context.Context, ip string) (*GeoIPData, error) {
	var provider GeoIPProvider

	switch s.config.Provider {
	case "ipapi.co":
		provider = &IPApiCoProvider{client: s.client}
	case "ip-api.com":
		provider = &IPApiProvider{client: s.client}
	default:
		return nil, fmt.Errorf("unsupported GeoIP provider: %s", s.config.Provider)
	}

	return provider.GetGeoIPData(ctx, ip)
}

// isPrivateIP checks if an IP address is private
func (s *GeoIPService) isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// Provider implementations

// GetGeoIPData implements IPApiCoProvider
func (p *IPApiCoProvider) GetGeoIPData(ctx context.Context, ip string) (*GeoIPData, error) {
	url := fmt.Sprintf("http://ipapi.co/%s/json/", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse struct {
		IP          string  `json:"ip"`
		Country     string  `json:"country_name"`
		CountryCode string  `json:"country_code_iso3"`
		Region      string  `json:"region"`
		RegionCode  string  `json:"region_code"`
		City        string  `json:"city"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		ISP         string  `json:"org"`
		Timezone    string  `json:"timezone"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &GeoIPData{
		IP:          apiResponse.IP,
		Country:     apiResponse.Country,
		CountryCode: apiResponse.CountryCode,
		Region:      apiResponse.Region,
		RegionCode:  apiResponse.RegionCode,
		City:        apiResponse.City,
		Latitude:    apiResponse.Latitude,
		Longitude:   apiResponse.Longitude,
		ISP:         apiResponse.ISP,
		Timezone:    apiResponse.Timezone,
		Timestamp:   time.Now(),
	}, nil
}

// GetGeoIPData implements IPApiProvider
func (p *IPApiProvider) GetGeoIPData(ctx context.Context, ip string) (*GeoIPData, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse struct {
		Query       string  `json:"query"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"regionName"`
		RegionCode  string  `json:"region"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		AS          string  `json:"as"`
		Timezone    string  `json:"timezone"`
		Mobile      bool    `json:"mobile"`
		Proxy       bool    `json:"proxy"`
		Hosting     bool    `json:"hosting"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &GeoIPData{
		IP:           apiResponse.Query,
		Country:      apiResponse.Country,
		CountryCode:  apiResponse.CountryCode,
		Region:       apiResponse.Region,
		RegionCode:   apiResponse.RegionCode,
		City:         apiResponse.City,
		Latitude:     apiResponse.Lat,
		Longitude:    apiResponse.Lon,
		ISP:          apiResponse.ISP,
		Organization: apiResponse.Org,
		AS:           apiResponse.AS,
		Timezone:     apiResponse.Timezone,
		IsMobile:     apiResponse.Mobile,
		IsProxy:      apiResponse.Proxy || apiResponse.Hosting,
		Timestamp:    time.Now(),
	}, nil
}