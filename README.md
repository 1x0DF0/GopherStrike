# GopherStrike

A comprehensive red team security framework written in Go for penetration testing, vulnerability assessment, and OSINT operations.

## Features

- **Port Scanner** - Network port scanning with nmap integration
- **Subdomain Scanner** - Advanced subdomain enumeration
- **OSINT & Vulnerability Tool** - Intelligence gathering and vulnerability identification
- **Web Application Security Scanner** - Tests for XSS, SQL injection, and other web vulnerabilities
- **S3 Bucket Scanner** - Identifies misconfigured AWS S3 buckets
- **Email Harvester** - Collects email addresses associated with target domains
- **Directory Bruteforcer** - Discovers hidden directories on web servers
- **Report Generator** - Creates comprehensive security assessment reports
- **Host & Subdomain Resolver** - DNS resolution and verification
- **Dependencies Checker** - Verifies required tools installation

## Requirements

- Go 1.16 or higher
- Git
- Optional: nmap (for port scanning functionality)
- Optional: SecLists (for enhanced wordlists)

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/GopherStrike.git
cd GopherStrike

# Install dependencies
go mod download

# Build the application
go build -o GopherStrike main.go
```

## Usage

Run the application:

```bash
./GopherStrike
```

You'll be presented with a menu:

```
===============================
      GopherStrike Menu
===============================
1. Port Scanner
2. Subdomain Scanner
3. OSINT & Vulnerability Tool
4. Web Application Security Scanner
5. S3 Bucket Scanner
6. Email Harvester
7. Directory Bruteforcer
8. Report Generator
9. Host & Subdomain Resolver
10. Dependencies Checker
11. Exit
===============================
Enter your choice:
```

Select a tool by entering its number and follow the prompts.

## Configuration

GopherStrike uses a JSON-based configuration system. Create a `config.json` file to customize settings:

```json
{
  "general": {
    "logLevel": "info",
    "logFile": "logs/gopherstrike.log"
  },
  "network": {
    "timeout": 30,
    "maxConcurrency": 50,
    "rateLimit": 100
  },
  "security": {
    "verifySSL": true,
    "useProxy": false
  }
}
```

## Output

All scan results and logs are saved in the `logs/` directory:
- JSON files for structured data
- Text files for summaries
- Optional CSV/HTML export formats

## Project Structure

```
GopherStrike/
├── main.go                 # Entry point
├── cmd/                    # Command implementations
├── pkg/                    # Core packages
│   ├── config/            # Configuration management
│   ├── tools/             # Security tools
│   ├── security/          # Security utilities
│   └── ...                # Additional packages
├── utils/                  # Utility functions
└── logs/                   # Output directory
```

## Security Considerations

- Always obtain proper authorization before scanning targets
- Use responsibly and ethically
- Be aware of rate limiting to avoid disrupting services
- Review logs for sensitive information before sharing

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This tool is for authorized security testing only. Users are responsible for complying with all applicable laws and regulations. The authors assume no liability for misuse or damage caused by this software.

## Acknowledgments

- Built with Go and the amazing Go community
- Inspired by various security tools and frameworks
- Thanks to all contributors and testers