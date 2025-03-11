package pkg

import (
	"GopherStrike/pkg/tools"
	"fmt"
	"strings"
)

// cleanDomain removes http/https and any trailing slashes from the domain
func cleanDomain(input string) string {
	// Remove http:// or https://
	domain := strings.TrimPrefix(input, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Remove trailing slash
	domain = strings.TrimSuffix(domain, "/")

	// Remove any www. prefix
	domain = strings.TrimPrefix(domain, "www.")

	return domain
}

// RunSubdomainScannerWithCheck executes the subdomain scanner
func RunSubdomainScannerWithCheck() error {
	// Get target domain from user
	var domain string
	fmt.Print("Enter target domain (e.g., example.com): ")
	if _, err := fmt.Scanln(&domain); err != nil {
		return fmt.Errorf("error reading domain: %v", err)
	}

	// Clean the domain input
	domain = cleanDomain(domain)
	if domain == "" {
		return fmt.Errorf("invalid domain provided")
	}

	fmt.Printf("\nScanning subdomains for: %s\n", domain)

	// Set up default scan options
	options := tools.ScanOptions{
		WordlistPath: "pkg/tools/wordlists/small-wordlist/common_subdomains_5000.txt",
		Threads:      10,
		CheckHTTP:    true,
		CheckSSL:     true,
		Timeout:      5,
		ResolveIPs:   true,
	}

	// Run the scan
	result, err := tools.ScanSubdomains(domain, options)
	if err != nil {
		return fmt.Errorf("scan error: %v", err)
	}

	// Print the results
	tools.PrintResults(result)
	return nil
}
