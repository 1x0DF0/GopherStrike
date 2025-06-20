# GopherStrike Security Documentation

## Overview

GopherStrike has been designed with security as a primary concern. This document outlines the security measures implemented throughout the application, best practices for secure usage, and guidelines for reporting security vulnerabilities.

## Security Architecture

### 1. Input Validation and Sanitization

**Location**: `pkg/validator/`

All user inputs are validated and sanitized before processing:

- **IP Address Validation**: Ensures only valid IPv4/IPv6 addresses are processed
- **Domain Validation**: Validates domain format and prevents malicious domains
- **URL Validation**: Enforces secure URL schemes and prevents SSRF attacks
- **File Path Validation**: Prevents path traversal attacks with comprehensive checks
- **Command Validation**: Prevents command injection by sanitizing shell metacharacters
- **Port Validation**: Ensures port numbers are within valid ranges

**Example Usage**:
```go
import "GopherStrike/pkg/validator"

// Validate IP address
ip, err := validator.ValidateIP(userInput)
if err != nil {
    return fmt.Errorf("invalid IP: %w", err)
}

// Validate domain
domain, err := validator.ValidateDomain(userInput)
if err != nil {
    return fmt.Errorf("invalid domain: %w", err)
}
```

### 2. Secure Command Execution

**Location**: `pkg/security/secure_exec.go`

All system commands are executed through a secure wrapper that:

- Prevents shell injection attacks
- Validates command whitelist
- Implements timeout controls
- Uses proper privilege escalation
- Sanitizes all arguments

**Example Usage**:
```go
import "GopherStrike/pkg/security"

// Create secure command
options := security.SecureCommandOptions{
    Timeout:         30 * time.Second,
    AllowedCommands: []string{"nmap", "python3"},
    DisableShell:    true,
}

cmd, err := security.NewSecureCommand("nmap", []string{"-sS", "target.com"}, options)
if err != nil {
    return err
}

// Execute safely
err = cmd.Run()
```

### 3. Encrypted Key Storage

**Location**: `pkg/security/keystore.go`

API keys and sensitive data are stored using AES-256-GCM encryption:

- **Encryption**: AES-256-GCM with PBKDF2 key derivation
- **Key Derivation**: PBKDF2 with SHA-256, 100,000 iterations
- **File Permissions**: 0600 (owner read/write only)
- **Salt**: Random 32-byte salt for each encryption
- **Nonce**: Random nonce for each encryption operation

**Example Usage**:
```go
import "GopherStrike/pkg/security"

// Create encrypted keystore
keystore, err := security.NewSecureKeyStore("/path/to/keys.enc", "password")
if err != nil {
    return err
}

// Store API key securely
err = keystore.Set("shodan_api_key", "your-api-key-here")
if err != nil {
    return err
}

// Retrieve API key
apiKey, err := keystore.Get("shodan_api_key")
if err != nil {
    return err
}
```

### 4. Network Security

**TLS Configuration**:
- Minimum TLS 1.2 enforced
- Certificate validation enabled by default
- Explicit user confirmation required to disable TLS verification

**HTTP Client Security**:
- Connection timeouts implemented
- Rate limiting support
- SSRF protection through URL validation

### 5. File System Security

**File Permissions**:
- Log files: 0600 (owner read/write only)
- Log directories: 0750 (owner full, group read/execute)
- Configuration files: 0600 (owner read/write only)
- Key stores: 0600 (owner read/write only)

**Path Security**:
- Path traversal prevention
- System directory access restriction
- Symlink attack prevention

### 6. Logging Security

**Log Sanitization**: `pkg/logging/logger.go`

All log messages are automatically sanitized to remove:
- API keys and tokens
- Passwords and secrets
- Credit card numbers
- SSH private keys
- Database connection strings
- Authorization headers

**Example**:
```go
// Before sanitization
log.Info("API key: sk-1234567890abcdef")

// After sanitization
log.Info("API key: ***REDACTED***")
```

### 7. Error Handling

**Location**: `pkg/errors/`

Comprehensive error handling system:
- Structured error types with severity levels
- Stack trace capture for debugging
- Context preservation
- Secure error messages (no sensitive data exposure)

## Security Best Practices

### For Users

1. **Keep Software Updated**
   - Regularly update GopherStrike to the latest version
   - Update dependencies and system packages

2. **Use Strong Passwords**
   - Use strong, unique passwords for encrypted keystores
   - Consider using a password manager

3. **Secure API Keys**
   - Store API keys in the encrypted keystore, not configuration files
   - Rotate API keys regularly
   - Use environment variables for CI/CD environments

4. **Network Security**
   - Use VPN when scanning external targets
   - Be aware of local network policies and restrictions

5. **File Permissions**
   - Ensure proper file permissions on logs and configuration
   - Regularly review access to sensitive files

### For Developers

1. **Input Validation**
   - Always validate user input using the validator package
   - Never trust external data

2. **Command Execution**
   - Use the secure command execution framework
   - Never use shell string interpolation

3. **Logging**
   - Use the logging framework for all log messages
   - Be mindful of what information is logged

4. **Error Handling**
   - Use the error handling framework
   - Don't expose internal details in error messages

## Threat Model

### Assets
- Scan results and reports
- API keys and credentials
- System access and privileges
- Network access

### Threats
- Command injection attacks
- Path traversal attacks
- Credential theft
- Privilege escalation
- Information disclosure
- Network attacks (SSRF, etc.)

### Mitigations
- Input validation and sanitization
- Secure command execution
- Encrypted storage
- Proper file permissions
- Network security controls
- Comprehensive logging and monitoring

## Security Testing

### Automated Testing
Run the security-focused test suite:
```bash
go test -v ./pkg/validator/...
go test -run="TestSecurity" ./...
```

### Manual Security Testing
1. **Input Validation Testing**
   - Test with malicious inputs
   - Verify command injection prevention
   - Test path traversal attempts

2. **File Permission Testing**
   ```bash
   ls -la logs/
   ls -la ~/.gopherstrike/
   ```

3. **Network Security Testing**
   - Verify TLS configuration
   - Test with invalid certificates
   - Check for SSRF vulnerabilities

## Vulnerability Reporting

If you discover a security vulnerability in GopherStrike:

### DO
- Report the issue privately via email or security contact
- Provide detailed reproduction steps
- Include potential impact assessment
- Allow reasonable time for fixes before public disclosure

### DON'T
- Publicly disclose the vulnerability before it's fixed
- Access systems you don't own
- Use vulnerabilities for malicious purposes

### Contact
For security issues, please contact: [security@gopherstrike.com]

## Security Checklist

Before deploying GopherStrike in production:

- [ ] All user inputs validated
- [ ] Secure command execution implemented
- [ ] API keys stored in encrypted keystore
- [ ] Proper file permissions set
- [ ] TLS validation enabled
- [ ] Logging sanitization active
- [ ] Security tests passing
- [ ] Dependencies updated
- [ ] Configuration reviewed
- [ ] Access controls implemented

## Compliance and Standards

GopherStrike follows:
- OWASP Top 10 security guidelines
- NIST Cybersecurity Framework recommendations
- Industry standard cryptographic practices
- Secure coding practices

## Security Monitoring

### Log Analysis
Monitor logs for:
- Failed authentication attempts
- Suspicious input patterns
- Command execution failures
- File access violations

### File Integrity
Monitor changes to:
- Configuration files
- Key stores
- Log files
- Executable binaries

### Network Monitoring
Monitor for:
- Unusual network connections
- DNS requests to suspicious domains
- Large data transfers
- Connection failures

## Incident Response

In case of a security incident:

1. **Immediate Response**
   - Isolate affected systems
   - Preserve evidence
   - Assess impact

2. **Investigation**
   - Analyze logs
   - Identify attack vectors
   - Determine scope

3. **Remediation**
   - Apply security patches
   - Update configurations
   - Rotate compromised credentials

4. **Recovery**
   - Restore systems
   - Verify security
   - Update monitoring

5. **Lessons Learned**
   - Document incident
   - Update procedures
   - Improve security measures

## Regular Security Maintenance

### Weekly
- Review security logs
- Check for software updates
- Monitor security advisories

### Monthly
- Update dependencies
- Review file permissions
- Test backup procedures
- Review access controls

### Quarterly
- Security audit
- Penetration testing
- Update threat model
- Review security policies

## Conclusion

Security is an ongoing process, not a destination. GopherStrike implements multiple layers of security controls, but users and developers must also follow best practices to maintain a secure environment. Regular updates, monitoring, and security testing are essential for maintaining the security posture of the application.

For questions about security features or to report issues, please contact the security team.