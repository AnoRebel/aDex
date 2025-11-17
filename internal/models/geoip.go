package models

import (
	"fmt"
	"time"
)

// GeoIPInfo represents geographical information for an IP address
type GeoIPInfo struct {
	IP           string  `json:"ip"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"countryCode"`
	Region       string  `json:"region"`
	RegionCode   string  `json:"regionCode"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ISP          string  `json:"isp"`
	Organization string  `json:"organization"`
	AS           string  `json:"as"`
	Timezone     string  `json:"timezone"`
	IsVPN        bool    `json:"isVpn"`
	IsProxy      bool    `json:"isProxy"`
	IsMobile     bool    `json:"isMobile"`
	Timestamp    time.Time `json:"timestamp"`
}

// NetworkConnectionWithGeo extends NetworkConnection with GeoIP data
type NetworkConnectionWithGeo struct {
	// Network connection fields
	LocalAddr     string    `json:"localAddr"`
	RemoteAddr    string    `json:"remoteAddr"`
	State         string    `json:"state"`
	PID           int       `json:"pid"`
	ProcessName   string    `json:"processName"`
	Protocol      string    `json:"protocol"`
	BytesSent     uint64    `json:"bytesSent"`
	BytesRecv     uint64    `json:"bytesRecv"`
	Established   time.Time `json:"established"`
	LastActivity  time.Time `json:"lastActivity"`

	// GeoIP fields
	RemoteGeoIP  *GeoIPInfo `json:"remoteGeoIP,omitempty"`
	CountryFlag  string     `json:"countryFlag,omitempty"`
}

// GeoIPSettings contains user preferences for GeoIP functionality
type GeoIPSettings struct {
	Enabled         bool          `json:"enabled"`
	Provider        string        `json:"provider"`
	ShowFlags       bool          `json:"showFlags"`
	ShowCity        bool          `json:"showCity"`
	ShowISP         bool          `json:"showISP"`
	CacheTimeout    time.Duration `json:"cacheTimeout"`
	AutoLookup      bool          `json:"autoLookup"`
	PrivacyMode     bool          `json:"privacyMode"`
}

// GeoIPCacheStats represents cache statistics
type GeoIPCacheStats struct {
	Size       int       `json:"size"`
	MaxSize    int       `json:"maxSize"`
	LastUpdate time.Time `json:"lastUpdate"`
	Enabled    bool      `json:"enabled"`
}

// GeoIPProvider represents a GeoIP service provider
type GeoIPProvider struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	FreeTier    bool   `json:"freeTier"`
	AuthType    string `json:"authType"` // "none", "apikey"
	Description string `json:"description"`
}

// GetCountryCode returns the ISO 3166-1 alpha-2 country code
func (g *GeoIPInfo) GetCountryCode() string {
	if len(g.CountryCode) == 3 {
		// Convert ISO 3166-1 alpha-3 to alpha-2 (basic conversion for common countries)
		countryMap := map[string]string{
			"USA": "US",
			"CAN": "CA",
			"GBR": "GB",
			"AUS": "AU",
			"DEU": "DE",
			"FRA": "FR",
			"ITA": "IT",
			"JPN": "JP",
			"CHN": "CN",
			"IND": "IN",
			"BRA": "BR",
			"MEX": "MX",
			"ESP": "ES",
			"RUS": "RU",
			"KOR": "KR",
			"NLD": "NL",
			"BEL": "BE",
			"SWE": "SE",
			"NOR": "NO",
			"DNK": "DK",
			"FIN": "FI",
			"CHE": "CH",
			"AUT": "AT",
			"NZL": "NZ",
			"IRL": "IE",
			"PRT": "PT",
			"GRC": "GR",
			"TUR": "TR",
			"POL": "PL",
			"CZE": "CZ",
			"HUN": "HU",
			"ROU": "RO",
			"BGR": "BG",
			"HRV": "HR",
			"SVK": "SK",
			"SVN": "SI",
			"EST": "EE",
			"LVA": "LV",
			"LTU": "LT",
		}
		if code, exists := countryMap[g.CountryCode]; exists {
			return code
		}
	}
	return g.CountryCode
}

// GetCountryFlag returns the Unicode flag emoji for the country
func (g *GeoIPInfo) GetCountryFlag() string {
	code := g.GetCountryCode()
	if len(code) != 2 {
		return ""
	}

	// Convert country code to flag emoji
	flagOffset := 0x1F1E6 // Regional Indicator Symbol Letter A
	asciiOffset := 0x41 // ASCII 'A'

	r1 := rune(flagOffset + int(code[0]) - asciiOffset)
	r2 := rune(flagOffset + int(code[1]) - asciiOffset)

	return string([]rune{r1, r2})
}

// GetLocationString returns a formatted location string
func (g *GeoIPInfo) GetLocationString() string {
	if g.City == "" {
		return g.Country
	}
	if g.Region == "" {
		return fmt.Sprintf("%s, %s", g.City, g.Country)
	}
	return fmt.Sprintf("%s, %s, %s", g.City, g.Region, g.Country)
}

// IsEmpty returns true if the GeoIP info contains no meaningful data
func (g *GeoIPInfo) IsEmpty() bool {
	return g.IP == "" && g.Country == "" && g.City == ""
}

// GetAvailableProviders returns available GeoIP providers
func GetAvailableGeoIPProviders() []GeoIPProvider {
	return []GeoIPProvider{
		{
			Name:        "ipapi.co",
			DisplayName: "IPAPI.co",
			FreeTier:    true,
			AuthType:    "none",
			Description: "Free IP geolocation API with no authentication required",
		},
		{
			Name:        "ip-api.com",
			DisplayName: "IP-API.com",
			FreeTier:    true,
			AuthType:    "none",
			Description: "Free IP geolocation API with comprehensive data",
		},
		{
			Name:        "maxmind",
			DisplayName: "MaxMind GeoIP2",
			FreeTier:    false,
			AuthType:    "apikey",
			Description: "Commercial GeoIP database with high accuracy",
		},
	}
}