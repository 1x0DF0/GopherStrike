package pkg

import (
	"GopherStrike/pkg/subdomain"
)

// RunSubdomainScannerWithCheck executes the subdomain scanner
func RunSubdomainScannerWithCheck() error {
	return subdomain.RunScanner()
}
