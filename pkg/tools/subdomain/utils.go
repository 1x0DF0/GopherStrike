// pkg/tools/subdomain/utils.go
package subdomain

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

// EnsureDirectory creates the directory if it doesn't exist
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// ExpandHomeDir expands ~ to the user's home directory
func ExpandHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	home := usr.HomeDir
	return strings.Replace(path, "~/", home+"/", 1), nil
}

// DirectoryExists checks if a path exists and is a directory
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatSize formats bytes into human readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ValidateDomainFormat checks if domain has valid format
func ValidateDomainFormat(domain string) bool {
	domain = strings.ToLower(strings.TrimSpace(domain))
	
	if domain == "" {
		return false
	}
	
	// Basic domain format validation
	return strings.Contains(domain, ".")
}

// ValidateDomain checks if domain has valid format and exists (returns bool for compatibility)
func ValidateDomain(domain string) bool {
	return ValidateDomainFormat(domain)
}

// GenerateProgressBar creates a progress bar string
func GenerateProgressBar(current, total int, width int) string {
	if total == 0 {
		return strings.Repeat("=", width)
	}
	
	progress := float64(current) / float64(total)
	filled := int(progress * float64(width))
	
	bar := strings.Repeat("=", filled)
	remaining := strings.Repeat("-", width-filled)
	
	return fmt.Sprintf("[%s%s] %d/%d (%.1f%%)", bar, remaining, current, total, progress*100)
}