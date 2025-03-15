// pkg/subdomain/utils.go
package subdomain

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// Cache for domain validation to prevent redundant lookups
	domainCache     = make(map[string]bool)
	domainCacheLock sync.RWMutex

	// Precompiled regex patterns for better performance
	domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
)

// CleanDomain removes http/https and any trailing slashes from the domain
func CleanDomain(input string) string {
	// Convert to lowercase for consistency
	domain := strings.ToLower(strings.TrimSpace(input))

	// Remove http:// or https:// prefixes
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Remove trailing slash
	domain = strings.TrimSuffix(domain, "/")

	// Remove any www. prefix
	domain = strings.TrimPrefix(domain, "www.")

	// Remove any paths, query parameters, etc.
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port number if specified
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// ValidateDomainFormat checks if a domain name format is valid
func ValidateDomainFormat(domain string) bool {
	// Basic validation
	if domain == "" {
		return false
	}

	// Check domain name format using precompiled regex
	return domainRegex.MatchString(domain)
}

// ValidateDomain checks if a domain name is valid and exists
func ValidateDomain(domain string) bool {
	// Format validation first (fast)
	if !ValidateDomainFormat(domain) {
		return false
	}

	// Check cache first
	domainCacheLock.RLock()
	if result, found := domainCache[domain]; found {
		domainCacheLock.RUnlock()
		return result
	}
	domainCacheLock.RUnlock()

	// Try to resolve the domain with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resolver net.Resolver
	_, err := resolver.LookupHost(ctx, domain)

	// Cache the result
	result := err == nil
	domainCacheLock.Lock()
	domainCache[domain] = result
	domainCacheLock.Unlock()

	return result
}

// FormatSize formats a file size in bytes to a human-readable string
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

// GenerateProgressBar creates a simple ASCII progress bar
func GenerateProgressBar(progress, total int, width int) string {
	if total <= 0 {
		return "[----------] Unknown progress"
	}

	percent := float64(progress) / float64(total)
	completeWidth := int(percent * float64(width))

	bar := "["
	for i := 0; i < width; i++ {
		if i < completeWidth {
			bar += "="
		} else if i == completeWidth {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	return fmt.Sprintf("%s %.1f%% (%d/%d)", bar, percent*100, progress, total)
}

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirectoryExists checks if a directory exists
func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// EnsureDirectory ensures a directory exists, creating it if necessary
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ExpandHomeDir expands the tilde in a file path to the user's home directory
func ExpandHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return strings.Replace(path, "~/", home+"/", 1), nil
}

// getDefaultThreadCount returns the optimal default thread count based on system resources
func getDefaultThreadCount() int {
	// Use number of CPU cores with a minimum of 4 and maximum of 20
	numCPU := runtime.NumCPU()
	if numCPU < 4 {
		return 4
	}
	if numCPU > 20 {
		return 20
	}
	return numCPU
}

// GetDomainInput gets and validates the target domain from user input
func GetDomainInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter target domain (e.g., example.com): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading domain: %v", err)
		}

		// Clean the domain input
		domain := CleanDomain(input)
		if domain == "" {
			fmt.Println("Error: Invalid domain provided. Please enter a valid domain name.")
			continue
		}

		// Basic domain format validation
		if !ValidateDomainFormat(domain) {
			fmt.Println("Error: Invalid domain format. Please enter a valid domain name.")
			continue
		}

		// Check if domain exists (not mandatory, can be skipped)
		fmt.Printf("Checking if domain %s exists... ", domain)

		// Set a timeout for the validation
		domainCheckChan := make(chan bool)
		go func() {
			domainCheckChan <- ValidateDomain(domain)
		}()

		select {
		case exists := <-domainCheckChan:
			if exists {
				fmt.Printf("✓ Domain exists.\n")
			} else {
				fmt.Printf("⚠ Unable to verify domain.\n")
				fmt.Print("Continue anyway? (y/n): ")
				continueAnyway, _ := reader.ReadString('\n')
				if strings.ToLower(strings.TrimSpace(continueAnyway)) != "y" {
					continue
				}
			}
		case <-time.After(3 * time.Second):
			fmt.Printf("⚠ Verification timed out.\n")
			fmt.Print("Continue anyway? (y/n): ")
			continueAnyway, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(continueAnyway)) != "y" {
				continue
			}
		}

		return domain, nil
	}
}

// GetWordlistPath gets the wordlist path from user input
func GetWordlistPath() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nWordlist:")
	fmt.Println("=========")
	fmt.Println("You must provide your own wordlist file.")
	fmt.Println("Examples:")
	fmt.Println("- Kali Linux SecLists: /usr/share/seclists/Discovery/DNS/")
	fmt.Println("- OWASP Amass: /usr/share/amass/wordlists/")
	fmt.Println("- Custom wordlists: ~/wordlists/subdomains.txt")

	for {
		fmt.Print("\nEnter full path to wordlist: ")
		wordlistPath, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading path: %v", err)
		}

		wordlistPath = strings.TrimSpace(wordlistPath)
		if wordlistPath == "" {
			fmt.Println("Error: Wordlist path cannot be empty.")
			continue
		}

		// Expand home directory if using ~
		expandedPath, err := ExpandHomeDir(wordlistPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		wordlistPath = expandedPath

		// Check if the file exists
		if !FileExists(wordlistPath) {
			fmt.Printf("Error: Wordlist not found at: %s\n", wordlistPath)
			continue
		}

		// Check if it's a directory
		if DirectoryExists(wordlistPath) {
			fmt.Println("Error: The provided path is a directory, not a file.")
			continue
		}

		// Get file size
		fileSize, err := GetFileSize(wordlistPath)
		if err != nil {
			fmt.Printf("Error: Cannot get file size: %v\n", err)
			continue
		}
		fmt.Printf("Wordlist size: %s\n", FormatSize(fileSize))

		return wordlistPath, nil
	}
}

// CustomizeOptions allows user to customize scanning options
func CustomizeOptions(options ScanOptions) (ScanOptions, error) {
	reader := bufio.NewReader(os.Stdin)

	// Set optimal defaults
	if options.Threads == 0 {
		options.Threads = getDefaultThreadCount()
	}

	fmt.Print("\nCustomize scan options? (y/n): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return options, fmt.Errorf("error reading input: %v", err)
	}

	if strings.ToLower(strings.TrimSpace(input)) != "y" {
		return options, nil
	}

	// Thread count
	for {
		fmt.Printf("Thread count (1-100, default: %d): ", options.Threads)
		input, err = reader.ReadString('\n')
		if err != nil {
			return options, fmt.Errorf("error reading input: %v", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			break // Keep default
		}

		threads, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Error: Invalid number. Please enter a number between 1 and 100.")
			continue
		}

		if threads < 1 || threads > 100 {
			fmt.Println("Error: Thread count must be between 1 and 100.")
			continue
		}

		options.Threads = threads
		break
	}

	// HTTP check
	fmt.Printf("Check HTTP connectivity? (y/n, default: %t): ", options.CheckHTTP)
	input, err = reader.ReadString('\n')
	if err != nil {
		return options, fmt.Errorf("error reading input: %v", err)
	}

	input = strings.TrimSpace(input)
	if input != "" {
		options.CheckHTTP = strings.ToLower(input) == "y"
	}

	// SSL check
	fmt.Printf("Check SSL/TLS? (y/n, default: %t): ", options.CheckSSL)
	input, err = reader.ReadString('\n')
	if err != nil {
		return options, fmt.Errorf("error reading input: %v", err)
	}

	input = strings.TrimSpace(input)
	if input != "" {
		options.CheckSSL = strings.ToLower(input) == "y"
	}

	// Timeout
	for {
		fmt.Printf("Connection timeout in seconds (1-60, default: %d): ", options.Timeout)
		input, err = reader.ReadString('\n')
		if err != nil {
			return options, fmt.Errorf("error reading input: %v", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			break // Keep default
		}

		timeout, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Error: Invalid number. Please enter a number between 1 and 60.")
			continue
		}

		if timeout < 1 || timeout > 60 {
			fmt.Println("Error: Timeout must be between 1 and 60 seconds.")
			continue
		}

		options.Timeout = timeout
		break
	}

	// Resolve IPs
	fmt.Printf("Resolve IP addresses? (y/n, default: %t): ", options.ResolveIPs)
	input, err = reader.ReadString('\n')
	if err != nil {
		return options, fmt.Errorf("error reading input: %v", err)
	}

	input = strings.TrimSpace(input)
	if input != "" {
		options.ResolveIPs = strings.ToLower(input) == "y"
	}

	return options, nil
}

// ScanOptions defines options for subdomain scanning
type ScanOptions struct {
	WordlistPath string // Path to wordlist file
	Threads      int    // Number of concurrent goroutines
	CheckHTTP    bool   // Whether to check HTTP status
	CheckSSL     bool   // Whether to check SSL certificates
	Timeout      int    // Timeout in seconds for each check
	ResolveIPs   bool   // Whether to resolve IPs
}
