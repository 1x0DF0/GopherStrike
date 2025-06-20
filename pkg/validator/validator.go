// Package validator provides comprehensive input validation for security
package validator

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Validator interface for all input validation
type Validator interface {
	Validate(input string) error
	Sanitize(input string) string
}

// IPValidator validates IP addresses
type IPValidator struct{}

func (v *IPValidator) Validate(input string) error {
	if input == "" {
		return fmt.Errorf("IP address cannot be empty")
	}
	
	if net.ParseIP(input) == nil {
		return fmt.Errorf("invalid IP address: %s", input)
	}
	
	return nil
}

func (v *IPValidator) Sanitize(input string) string {
	return strings.TrimSpace(input)
}

// DomainValidator validates domain names
type DomainValidator struct{}

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func (v *DomainValidator) Validate(input string) error {
	if input == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	
	// Remove protocol if present
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "https://")
	
	// Remove path if present
	if idx := strings.Index(input, "/"); idx != -1 {
		input = input[:idx]
	}
	
	// Extract hostname from host:port format
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		input = parts[0] // Take just the hostname part
	}
	
	// Check if it's an IP address (which is valid for some tools)
	if net.ParseIP(input) != nil {
		return nil
	}
	
	// For host:port format, validate the hostname part
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ":")
		hostname := parts[0]
		
		// Check if hostname is IP
		if net.ParseIP(hostname) != nil {
			return nil
		}
		
		// Validate hostname part
		if !domainRegex.MatchString(hostname) {
			return fmt.Errorf("invalid domain format in host:port: %s", hostname)
		}
		
		// Validate port part
		if len(parts) == 2 {
			portValidator := &PortValidator{}
			if err := portValidator.Validate(parts[1]); err != nil {
				return fmt.Errorf("invalid port in host:port: %w", err)
			}
		}
		return nil
	}
	
	// Validate domain format
	if !domainRegex.MatchString(input) {
		return fmt.Errorf("invalid domain format: %s", input)
	}
	
	return nil
}

func (v *DomainValidator) Sanitize(input string) string {
	// Remove protocol and trailing slashes
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimSuffix(input, "/")
	
	// Remove path if present
	if idx := strings.Index(input, "/"); idx != -1 {
		input = input[:idx]
	}
	
	// For domain validation, we typically want just the hostname
	// but for some tools, host:port format might be acceptable
	// So we'll keep the full host:port format in sanitize
	
	return input
}

// PortValidator validates port numbers and ranges
type PortValidator struct{}

func (v *PortValidator) Validate(input string) error {
	if input == "" {
		return fmt.Errorf("port cannot be empty")
	}
	
	// Check if it's a single port
	if port, err := strconv.Atoi(input); err == nil {
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
		return nil
	}
	
	// Check if it's a port range
	if strings.Contains(input, "-") {
		parts := strings.Split(input, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid port range format")
		}
		
		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		
		if err1 != nil || err2 != nil {
			return fmt.Errorf("invalid port numbers in range")
		}
		
		if start < 1 || start > 65535 || end < 1 || end > 65535 {
			return fmt.Errorf("ports must be between 1 and 65535")
		}
		
		if start > end {
			return fmt.Errorf("start port must be less than end port")
		}
		
		return nil
	}
	
	return fmt.Errorf("invalid port format")
}

func (v *PortValidator) Sanitize(input string) string {
	return strings.TrimSpace(input)
}

// URLValidator validates URLs
type URLValidator struct {
	AllowedSchemes []string
}

func NewURLValidator() *URLValidator {
	return &URLValidator{
		AllowedSchemes: []string{"http", "https"},
	}
}

func (v *URLValidator) Validate(input string) error {
	if input == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	parsedURL, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	
	// Check scheme
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http:// or https://)")
	}
	
	schemeAllowed := false
	for _, allowed := range v.AllowedSchemes {
		if parsedURL.Scheme == allowed {
			schemeAllowed = true
			break
		}
	}
	
	if !schemeAllowed {
		return fmt.Errorf("URL scheme must be one of: %s", strings.Join(v.AllowedSchemes, ", "))
	}
	
	// Check host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include host")
	}
	
	return nil
}

func (v *URLValidator) Sanitize(input string) string {
	return strings.TrimSpace(input)
}

// FilePathValidator validates file paths
type FilePathValidator struct {
	MustExist    bool
	AllowedExts  []string
	MaxSizeBytes int64
}

func (v *FilePathValidator) Validate(input string) error {
	if input == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	
	// Clean the path
	cleanPath := filepath.Clean(input)
	
	// Convert to absolute path to detect traversal attempts
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	
	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected")
	}
	
	// Additional security checks
	if strings.Contains(absPath, "..") {
		return fmt.Errorf("absolute path contains traversal elements")
	}
	
	// Check for dangerous paths (Unix/Linux)
	dangerousPaths := []string{
		"/etc/passwd", "/etc/shadow", "/etc/hosts", "/proc/", "/sys/",
		"/root/", "/var/log/", "/tmp/", "/dev/",
	}
	
	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(absPath, dangerous) {
			return fmt.Errorf("access to system path denied: %s", dangerous)
		}
	}
	
	// Check for Windows dangerous paths
	if strings.Contains(strings.ToLower(absPath), "windows\\system32") ||
	   strings.Contains(strings.ToLower(absPath), "windows/system32") {
		return fmt.Errorf("access to Windows system directory denied")
	}
	
	// Check if file exists if required
	if v.MustExist {
		info, err := os.Stat(cleanPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file does not exist: %s", cleanPath)
			}
			return fmt.Errorf("error accessing file: %w", err)
		}
		
		// Check file size if specified
		if v.MaxSizeBytes > 0 && info.Size() > v.MaxSizeBytes {
			return fmt.Errorf("file size exceeds maximum allowed (%d bytes)", v.MaxSizeBytes)
		}
	}
	
	// Check file extension if specified
	if len(v.AllowedExts) > 0 {
		ext := filepath.Ext(cleanPath)
		allowed := false
		for _, allowedExt := range v.AllowedExts {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension must be one of: %s", strings.Join(v.AllowedExts, ", "))
		}
	}
	
	return nil
}

func (v *FilePathValidator) Sanitize(input string) string {
	return filepath.Clean(strings.TrimSpace(input))
}

// CommandValidator validates command arguments to prevent injection
type CommandValidator struct {
	AllowedChars *regexp.Regexp
}

func NewCommandValidator() *CommandValidator {
	// Only allow alphanumeric, dash, underscore, dot, and slash
	return &CommandValidator{
		AllowedChars: regexp.MustCompile(`^[a-zA-Z0-9\-_./]+$`),
	}
}

func (v *CommandValidator) Validate(input string) error {
	if input == "" {
		return nil // Empty is allowed for command args
	}
	
	// Check for common injection patterns
	dangerousPatterns := []string{
		";", "&&", "||", "|", "`", "$", "(", ")", "{", "}", "[", "]",
		">", "<", "&", "\n", "\r", "\x00",
	}
	
	for _, pattern := range dangerousPatterns {
		if strings.Contains(input, pattern) {
			return fmt.Errorf("potentially dangerous character detected: %s", pattern)
		}
	}
	
	// Check against allowed characters
	if !v.AllowedChars.MatchString(input) {
		return fmt.Errorf("input contains disallowed characters")
	}
	
	return nil
}

func (v *CommandValidator) Sanitize(input string) string {
	// Remove any potentially dangerous characters
	safe := strings.TrimSpace(input)
	
	// Replace dangerous characters with safe alternatives
	replacements := map[string]string{
		";":  "",
		"&&": "",
		"||": "",
		"|":  "",
		"`":  "",
		"$":  "",
		"(":  "",
		")":  "",
		"{":  "",
		"}":  "",
		"[":  "",
		"]":  "",
		">":  "",
		"<":  "",
		"&":  "",
		"\n": "",
		"\r": "",
	}
	
	for old, new := range replacements {
		safe = strings.ReplaceAll(safe, old, new)
	}
	
	return safe
}

// EmailValidator validates email addresses
type EmailValidator struct{}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (v *EmailValidator) Validate(input string) error {
	if input == "" {
		return fmt.Errorf("email cannot be empty")
	}
	
	if !emailRegex.MatchString(input) {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

func (v *EmailValidator) Sanitize(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

// IntegerValidator validates integer inputs within a range
type IntegerValidator struct {
	Min int
	Max int
}

func (v *IntegerValidator) Validate(input string) error {
	if input == "" {
		return fmt.Errorf("integer value cannot be empty")
	}
	
	num, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("invalid integer: %s", input)
	}
	
	if num < v.Min || num > v.Max {
		return fmt.Errorf("value must be between %d and %d", v.Min, v.Max)
	}
	
	return nil
}

func (v *IntegerValidator) Sanitize(input string) string {
	return strings.TrimSpace(input)
}

// Helper functions for common validations

// ValidateIP validates and sanitizes an IP address
func ValidateIP(input string) (string, error) {
	validator := &IPValidator{}
	if err := validator.Validate(input); err != nil {
		return "", err
	}
	return validator.Sanitize(input), nil
}

// ValidateDomain validates and sanitizes a domain name
func ValidateDomain(input string) (string, error) {
	validator := &DomainValidator{}
	if err := validator.Validate(input); err != nil {
		return "", err
	}
	return validator.Sanitize(input), nil
}

// ValidatePort validates and sanitizes a port or port range
func ValidatePort(input string) (string, error) {
	validator := &PortValidator{}
	if err := validator.Validate(input); err != nil {
		return "", err
	}
	return validator.Sanitize(input), nil
}

// ValidateURL validates and sanitizes a URL
func ValidateURL(input string) (string, error) {
	validator := NewURLValidator()
	if err := validator.Validate(input); err != nil {
		return "", err
	}
	return validator.Sanitize(input), nil
}

// ValidateFilePath validates and sanitizes a file path
func ValidateFilePath(input string, mustExist bool) (string, error) {
	validator := &FilePathValidator{MustExist: mustExist}
	if err := validator.Validate(input); err != nil {
		return "", err
	}
	return validator.Sanitize(input), nil
}

// ValidateCommand validates and sanitizes command arguments
func ValidateCommand(input string) (string, error) {
	validator := NewCommandValidator()
	if err := validator.Validate(input); err != nil {
		return "", err
	}
	return validator.Sanitize(input), nil
}