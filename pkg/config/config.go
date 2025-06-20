// Package config provides configuration management for GopherStrike
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config represents the application configuration
type Config struct {
	// General settings
	General GeneralConfig `json:"general"`
	
	// Security settings
	Security SecurityConfig `json:"security"`
	
	// Network settings
	Network NetworkConfig `json:"network"`
	
	// Scanning settings
	Scanning ScanningConfig `json:"scanning"`
	
	// Output settings
	Output OutputConfig `json:"output"`
	
	// Tool-specific settings
	Tools ToolsConfig `json:"tools"`
}

// GeneralConfig contains general application settings
type GeneralConfig struct {
	LogLevel        string `json:"log_level"`        // debug, info, warning, error
	MaxConcurrency  int    `json:"max_concurrency"`  // Maximum concurrent operations
	TempDirectory   string `json:"temp_directory"`   // Directory for temporary files
	DataDirectory   string `json:"data_directory"`   // Directory for data files
	UpdateCheck     bool   `json:"update_check"`     // Check for updates on startup
	TelemetryEnabled bool  `json:"telemetry_enabled"` // Send anonymous usage statistics
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	RequireAuth      bool   `json:"require_auth"`       // Require authentication
	APIKeyFile       string `json:"api_key_file"`       // Path to API key file
	MaxLoginAttempts int    `json:"max_login_attempts"` // Maximum login attempts
	SessionTimeout   int    `json:"session_timeout"`    // Session timeout in minutes
	SecureMode       bool   `json:"secure_mode"`        // Enable additional security checks
}

// NetworkConfig contains network-related settings
type NetworkConfig struct {
	Timeout         int      `json:"timeout"`           // Default timeout in seconds
	MaxRetries      int      `json:"max_retries"`       // Maximum retry attempts
	RetryDelay      int      `json:"retry_delay"`       // Delay between retries in seconds
	ProxyURL        string   `json:"proxy_url"`         // HTTP/HTTPS proxy URL
	UserAgent       string   `json:"user_agent"`        // Default user agent
	DNSServers      []string `json:"dns_servers"`       // Custom DNS servers
	RateLimit       int      `json:"rate_limit"`        // Requests per second
}

// ScanningConfig contains scanning-related settings
type ScanningConfig struct {
	DefaultThreads   int      `json:"default_threads"`    // Default number of threads
	DefaultTimeout   int      `json:"default_timeout"`    // Default scan timeout
	PortRanges       []string `json:"port_ranges"`        // Default port ranges
	SkipHostCheck    bool     `json:"skip_host_check"`    // Skip host availability check
	SaveAllResults   bool     `json:"save_all_results"`   // Save all results, not just positive
	AutoSaveInterval int      `json:"auto_save_interval"` // Auto-save interval in seconds
}

// OutputConfig contains output-related settings
type OutputConfig struct {
	DefaultFormat    string   `json:"default_format"`     // json, csv, txt, html
	OutputDirectory  string   `json:"output_directory"`   // Default output directory
	LogDirectory     string   `json:"log_directory"`      // Log files directory
	Verbose          bool     `json:"verbose"`            // Verbose output
	ColorOutput      bool     `json:"color_output"`       // Use colored output
	TimestampFormat  string   `json:"timestamp_format"`   // Timestamp format
	CompressResults  bool     `json:"compress_results"`   // Compress result files
	ExportFormats    []string `json:"export_formats"`     // Enabled export formats
}

// ToolsConfig contains tool-specific settings
type ToolsConfig struct {
	PortScanner     PortScannerConfig     `json:"port_scanner"`
	SubdomainScanner SubdomainScannerConfig `json:"subdomain_scanner"`
	WebVulnScanner  WebVulnScannerConfig  `json:"web_vuln_scanner"`
	OSINTScanner    OSINTScannerConfig    `json:"osint_scanner"`
}

// PortScannerConfig contains port scanner settings
type PortScannerConfig struct {
	ScanTechnique   string   `json:"scan_technique"`    // SYN, TCP, UDP, etc.
	ServiceDetection bool    `json:"service_detection"` // Enable service detection
	OSDetection     bool     `json:"os_detection"`      // Enable OS detection
	CommonPorts     []int    `json:"common_ports"`      // List of common ports
	ExcludePorts    []int    `json:"exclude_ports"`     // Ports to exclude
	NmapPath        string   `json:"nmap_path"`         // Path to nmap binary
}

// SubdomainScannerConfig contains subdomain scanner settings
type SubdomainScannerConfig struct {
	DefaultWordlist  string   `json:"default_wordlist"`   // Default wordlist path
	ResolveIPs       bool     `json:"resolve_ips"`        // Resolve subdomain IPs
	CheckHTTP        bool     `json:"check_http"`         // Check HTTP/HTTPS
	RecursiveScan    bool     `json:"recursive_scan"`     // Enable recursive scanning
	WildcardCheck    bool     `json:"wildcard_check"`     // Check for wildcard domains
	DNSProviders     []string `json:"dns_providers"`      // DNS providers to use
}

// WebVulnScannerConfig contains web vulnerability scanner settings
type WebVulnScannerConfig struct {
	PayloadLevel     int      `json:"payload_level"`      // 1-5, higher = more payloads
	TestAllParams    bool     `json:"test_all_params"`    // Test all parameters
	FollowRedirects  bool     `json:"follow_redirects"`   // Follow HTTP redirects
	MaxRedirects     int      `json:"max_redirects"`      // Maximum redirects
	CustomPayloads   string   `json:"custom_payloads"`    // Path to custom payloads
	ExcludePatterns  []string `json:"exclude_patterns"`   // URL patterns to exclude
}

// OSINTScannerConfig contains OSINT scanner settings
type OSINTScannerConfig struct {
	EnabledSources   []string `json:"enabled_sources"`    // Enabled OSINT sources
	APIKeys          map[string]string `json:"api_keys"`  // API keys for services
	MaxDepth         int      `json:"max_depth"`          // Maximum search depth
	SaveRawData      bool     `json:"save_raw_data"`      // Save raw API responses
	CacheResults     bool     `json:"cache_results"`      // Cache OSINT results
	CacheDuration    int      `json:"cache_duration"`     // Cache duration in hours
}

var (
	instance *Config
	once     sync.Once
	mu       sync.RWMutex
)

// Get returns the global configuration instance
func Get() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.LoadDefaults()
	})
	return instance
}

// LoadDefaults loads default configuration values
func (c *Config) LoadDefaults() {
	c.General = GeneralConfig{
		LogLevel:        "info",
		MaxConcurrency:  10,
		TempDirectory:   filepath.Join(os.TempDir(), "gopherstrike"),
		DataDirectory:   filepath.Join(getHomeDir(), ".gopherstrike", "data"),
		UpdateCheck:     true,
		TelemetryEnabled: false,
	}
	
	c.Security = SecurityConfig{
		RequireAuth:      false,
		APIKeyFile:       filepath.Join(getHomeDir(), ".gopherstrike", "api_keys.json"),
		MaxLoginAttempts: 3,
		SessionTimeout:   30,
		SecureMode:       true,
	}
	
	c.Network = NetworkConfig{
		Timeout:    30,
		MaxRetries: 3,
		RetryDelay: 5,
		UserAgent:  "GopherStrike/1.0",
		DNSServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
		RateLimit:  10,
	}
	
	c.Scanning = ScanningConfig{
		DefaultThreads:   10,
		DefaultTimeout:   30,
		PortRanges:       []string{"1-1000", "8080", "8443"},
		SkipHostCheck:    false,
		SaveAllResults:   false,
		AutoSaveInterval: 300,
	}
	
	c.Output = OutputConfig{
		DefaultFormat:    "json",
		OutputDirectory:  filepath.Join(getHomeDir(), ".gopherstrike", "results"),
		LogDirectory:     filepath.Join(getHomeDir(), ".gopherstrike", "logs"),
		Verbose:          false,
		ColorOutput:      true,
		TimestampFormat:  time.RFC3339,
		CompressResults:  false,
		ExportFormats:    []string{"json", "csv", "txt"},
	}
	
	c.Tools = ToolsConfig{
		PortScanner: PortScannerConfig{
			ScanTechnique:    "SYN",
			ServiceDetection: true,
			OSDetection:      false,
			CommonPorts:      []int{21, 22, 23, 25, 80, 443, 445, 3306, 3389, 8080},
			ExcludePorts:     []int{},
		},
		SubdomainScanner: SubdomainScannerConfig{
			DefaultWordlist: "",
			ResolveIPs:      true,
			CheckHTTP:       true,
			RecursiveScan:   false,
			WildcardCheck:   true,
			DNSProviders:    []string{"8.8.8.8:53", "1.1.1.1:53"},
		},
		WebVulnScanner: WebVulnScannerConfig{
			PayloadLevel:    3,
			TestAllParams:   false,
			FollowRedirects: true,
			MaxRedirects:    5,
			CustomPayloads:  "",
			ExcludePatterns: []string{},
		},
		OSINTScanner: OSINTScannerConfig{
			EnabledSources: []string{"shodan", "censys", "virustotal"},
			APIKeys:        make(map[string]string),
			MaxDepth:       3,
			SaveRawData:    false,
			CacheResults:   true,
			CacheDuration:  24,
		},
	}
}

// LoadFromFile loads configuration from a JSON file
func (c *Config) LoadFromFile(filename string) error {
	mu.Lock()
	defer mu.Unlock()
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return nil
}

// SaveToFile saves configuration to a JSON file
func (c *Config) SaveToFile(filename string) error {
	mu.RLock()
	defer mu.RUnlock()
	
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write with secure permissions
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetString returns a string value from the config
func (c *Config) GetString(path string) string {
	mu.RLock()
	defer mu.RUnlock()
	
	// Simple path resolver - can be enhanced
	switch path {
	case "general.log_level":
		return c.General.LogLevel
	case "network.user_agent":
		return c.Network.UserAgent
	case "output.default_format":
		return c.Output.DefaultFormat
	default:
		return ""
	}
}

// GetInt returns an integer value from the config
func (c *Config) GetInt(path string) int {
	mu.RLock()
	defer mu.RUnlock()
	
	switch path {
	case "general.max_concurrency":
		return c.General.MaxConcurrency
	case "network.timeout":
		return c.Network.Timeout
	case "scanning.default_threads":
		return c.Scanning.DefaultThreads
	default:
		return 0
	}
}

// GetBool returns a boolean value from the config
func (c *Config) GetBool(path string) bool {
	mu.RLock()
	defer mu.RUnlock()
	
	switch path {
	case "general.update_check":
		return c.General.UpdateCheck
	case "security.secure_mode":
		return c.Security.SecureMode
	case "output.verbose":
		return c.Output.Verbose
	default:
		return false
	}
}

// Set updates a configuration value
func (c *Config) Set(path string, value interface{}) error {
	mu.Lock()
	defer mu.Unlock()
	
	// Simple setter - can be enhanced with reflection
	switch path {
	case "general.log_level":
		if v, ok := value.(string); ok {
			c.General.LogLevel = v
		}
	case "network.timeout":
		if v, ok := value.(int); ok {
			c.Network.Timeout = v
		}
	case "output.verbose":
		if v, ok := value.(bool); ok {
			c.Output.Verbose = v
		}
	default:
		return fmt.Errorf("unknown configuration path: %s", path)
	}
	
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	mu.RLock()
	defer mu.RUnlock()
	
	// Validate log level
	validLogLevels := []string{"debug", "info", "warning", "error"}
	valid := false
	for _, level := range validLogLevels {
		if c.General.LogLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log level: %s", c.General.LogLevel)
	}
	
	// Validate network settings
	if c.Network.Timeout < 1 || c.Network.Timeout > 300 {
		return fmt.Errorf("network timeout must be between 1 and 300 seconds")
	}
	
	if c.Network.MaxRetries < 0 || c.Network.MaxRetries > 10 {
		return fmt.Errorf("max retries must be between 0 and 10")
	}
	
	// Validate scanning settings
	if c.Scanning.DefaultThreads < 1 || c.Scanning.DefaultThreads > 100 {
		return fmt.Errorf("default threads must be between 1 and 100")
	}
	
	return nil
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	return home
}

// ConfigManager provides high-level configuration management
type ConfigManager struct {
	configFile string
	config     *Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configFile string) *ConfigManager {
	return &ConfigManager{
		configFile: configFile,
		config:     Get(),
	}
}

// Initialize loads or creates the configuration
func (cm *ConfigManager) Initialize() error {
	// Try to load existing config
	if err := cm.config.LoadFromFile(cm.configFile); err != nil {
		// If file doesn't exist, create default config
		if os.IsNotExist(err) {
			cm.config.LoadDefaults()
			if err := cm.config.SaveToFile(cm.configFile); err != nil {
				return fmt.Errorf("failed to save default config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}
	
	// Validate configuration
	if err := cm.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	return nil
}

// Get returns the configuration
func (cm *ConfigManager) Get() *Config {
	return cm.config
}

// Save saves the current configuration
func (cm *ConfigManager) Save() error {
	return cm.config.SaveToFile(cm.configFile)
}