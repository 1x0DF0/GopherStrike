# GopherStrike API Documentation

## Table of Contents
1. [Overview](#overview)
2. [Security Framework](#security-framework)
3. [Validation Framework](#validation-framework)
4. [Error Handling](#error-handling)
5. [Configuration Management](#configuration-management)
6. [Resource Management](#resource-management)
7. [Logging System](#logging-system)
8. [Command Execution](#command-execution)
9. [Key Storage](#key-storage)
10. [Examples](#examples)

## Overview

GopherStrike provides a comprehensive set of APIs for secure penetration testing and security assessment. All APIs are designed with security-first principles and include extensive validation, error handling, and logging.

## Security Framework

### Package: `pkg/security`

#### Secure Command Execution

```go
type SecureCommandOptions struct {
    WorkingDirectory    string
    Environment        map[string]string
    Timeout           time.Duration
    AllowedCommands   []string
    RequireAbsolutePath bool
    DisableShell      bool
    UID              *int
    GID              *int
}
```

**Functions:**

```go
// NewSecureCommand creates a new secure command with validation
func NewSecureCommand(name string, args []string, options SecureCommandOptions) (*SecureCommand, error)

// Methods on SecureCommand
func (sc *SecureCommand) Run() error
func (sc *SecureCommand) Start() error
func (sc *SecureCommand) Wait() error
func (sc *SecureCommand) Output() ([]byte, error)
func (sc *SecureCommand) CombinedOutput() ([]byte, error)
func (sc *SecureCommand) Kill() error
```

**Example:**
```go
options := security.SecureCommandOptions{
    Timeout:         30 * time.Second,
    AllowedCommands: []string{"nmap"},
    DisableShell:    true,
}

cmd, err := security.NewSecureCommand("nmap", []string{"-sS", "target.com"}, options)
if err != nil {
    return err
}

output, err := cmd.Output()
```

#### Privilege Escalation

```go
type PrivilegeEscalationManager struct {
    method EscalationMethod
}

// Methods
func NewPrivilegeEscalationManager() *PrivilegeEscalationManager
func (pem *PrivilegeEscalationManager) ExecuteWithElevatedPrivileges(command string, args []string, options SecureCommandOptions) error
func IsElevated() bool
```

**Example:**
```go
pem := security.NewPrivilegeEscalationManager()
options := security.SecureCommandOptions{
    Timeout:         5 * time.Minute,
    AllowedCommands: []string{"nmap"},
    DisableShell:    true,
}

err := pem.ExecuteWithElevatedPrivileges("nmap", []string{"-sS", "target.com"}, options)
```

## Validation Framework

### Package: `pkg/validator`

#### Core Validators

```go
type Validator interface {
    Validate(input string) error
    Sanitize(input string) string
}
```

**Available Validators:**
- `IPValidator` - Validates IPv4/IPv6 addresses
- `DomainValidator` - Validates domain names
- `PortValidator` - Validates ports and port ranges
- `URLValidator` - Validates URLs with scheme restrictions
- `FilePathValidator` - Validates file paths with security checks
- `CommandValidator` - Validates command arguments
- `EmailValidator` - Validates email addresses
- `IntegerValidator` - Validates integers within ranges

**Helper Functions:**
```go
func ValidateIP(input string) (string, error)
func ValidateDomain(input string) (string, error)
func ValidatePort(input string) (string, error)
func ValidateURL(input string) (string, error)
func ValidateFilePath(input string, mustExist bool) (string, error)
func ValidateCommand(input string) (string, error)
```

**Examples:**
```go
// Validate IP address
ip, err := validator.ValidateIP("192.168.1.1")
if err != nil {
    return fmt.Errorf("invalid IP: %w", err)
}

// Validate domain with custom validator
domainValidator := &validator.DomainValidator{}
err := domainValidator.Validate("example.com")
if err != nil {
    return err
}
sanitized := domainValidator.Sanitize("https://example.com/path")
// Result: "example.com"

// Validate file path
fileValidator := &validator.FilePathValidator{
    MustExist:    true,
    AllowedExts:  []string{".txt", ".json"},
    MaxSizeBytes: 10 * 1024 * 1024, // 10MB
}
path, err := fileValidator.Sanitize("/path/to/file.txt")
if err := fileValidator.Validate(path); err != nil {
    return err
}
```

## Error Handling

### Package: `pkg/errors`

#### Error Types and Severity

```go
type ErrorType int
const (
    ValidationError ErrorType = iota
    NetworkError
    FileError
    SecurityError
    ConfigError
    SystemError
    UserError
)

type ErrorSeverity int
const (
    SeverityLow ErrorSeverity = iota
    SeverityMedium
    SeverityHigh
    SeverityCritical
)
```

#### Error Creation Functions

```go
func New(errType ErrorType, severity ErrorSeverity, message string) *AppError
func Wrap(err error, errType ErrorType, message string) *AppError

// Helper functions
func ValidationFailed(field string, reason string) *AppError
func NetworkFailed(operation string, err error) *AppError
func FileFailed(operation string, path string, err error) *AppError
func SecurityFailed(issue string) *AppError
func ConfigFailed(setting string, reason string) *AppError
func SystemFailed(operation string, err error) *AppError
func UserInputError(input string, reason string) *AppError
```

**Examples:**
```go
// Create validation error
err := errors.ValidationFailed("ip_address", "invalid format")

// Wrap existing error
err = errors.Wrap(originalErr, errors.NetworkError, "failed to connect to target")

// Add context
err = err.WithContext("target", "192.168.1.1").WithSeverity(errors.SeverityHigh)

// Use error handler
handler := errors.NewErrorHandler()
handler.SetLogFunction(customLogFunc)
handler.Handle(err)
```

## Configuration Management

### Package: `pkg/config`

#### Configuration Structure

```go
type Config struct {
    General   GeneralConfig   `json:"general"`
    Security  SecurityConfig  `json:"security"`
    Network   NetworkConfig   `json:"network"`
    Scanning  ScanningConfig  `json:"scanning"`
    Output    OutputConfig    `json:"output"`
    Tools     ToolsConfig     `json:"tools"`
}
```

#### Configuration Manager

```go
type ConfigManager struct {
    configFile string
    config     *Config
}

// Functions
func NewConfigManager(configFile string) *ConfigManager
func (cm *ConfigManager) Initialize() error
func (cm *ConfigManager) Get() *Config
func (cm *ConfigManager) Save() error
```

**Examples:**
```go
// Initialize configuration
cm := config.NewConfigManager("/path/to/config.json")
err := cm.Initialize()
if err != nil {
    return err
}

// Get configuration
cfg := cm.Get()
timeout := cfg.Network.Timeout

// Update configuration
cfg.Set("network.timeout", 60)
err = cm.Save()
```

## Resource Management

### Package: `pkg/resources`

#### Resource Manager

```go
type Manager struct {
    mu        sync.RWMutex
    resources map[string]Resource
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    closed    bool
}

// Functions
func NewManager(ctx context.Context) *Manager
func (m *Manager) Register(resource Resource) error
func (m *Manager) Unregister(id string) error
func (m *Manager) Get(id string) (Resource, bool)
func (m *Manager) Close() error
```

#### Helper Functions

```go
func WithResource(manager *Manager, resource Resource, fn func(Resource) error) error
func WithFile(manager *Manager, path string, flag int, perm os.FileMode, fn func(*os.File) error) error
func WithHTTPClient(manager *Manager, timeout time.Duration, fn func(*http.Client) error) error
func SafeClose(closer io.Closer, name string)
```

**Examples:**
```go
// Create resource manager
ctx := context.Background()
manager := resources.NewManager(ctx)
defer manager.Close()

// Use managed file
err := resources.WithFile(manager, "/path/to/file", os.O_RDONLY, 0644, func(file *os.File) error {
    // Use file safely
    return nil
})

// Use managed HTTP client
err = resources.WithHTTPClient(manager, 30*time.Second, func(client *http.Client) error {
    resp, err := client.Get("https://example.com")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
})
```

## Logging System

### Package: `pkg/logging`

#### Logger Configuration

```go
type Logger struct {
    mu             sync.Mutex
    level          LogLevel
    writers        map[LogLevel][]io.Writer
    formatter      Formatter
    enableConsole  bool
    consoleLevel   LogLevel
    showTimestamp  bool
    showSource     bool
    sourceRelative bool
}
```

#### Log Levels

```go
const (
    DEBUG LogLevel = iota
    INFO
    WARNING
    ERROR
    CRITICAL
)
```

#### Logger Functions

```go
func New(level LogLevel) *Logger
func (l *Logger) AddFileHandler(filePath string, level LogLevel) error
func (l *Logger) SetLevel(level LogLevel)
func (l *Logger) SetFormatter(formatter Formatter)

// Logging methods
func (l *Logger) Debug(format string, args ...interface{})
func (l *Logger) Info(format string, args ...interface{})
func (l *Logger) Warning(format string, args ...interface{})
func (l *Logger) Error(format string, args ...interface{})
func (l *Logger) Critical(format string, args ...interface{})
```

**Examples:**
```go
// Create logger
logger := logging.New(logging.INFO)

// Add file handler
err := logger.AddFileHandler("/path/to/app.log", logging.INFO)

// Log messages (automatically sanitized)
logger.Info("Starting scan for target: %s", target)
logger.Error("Scan failed: %v", err)

// Get module-specific logger
scannerLogger := logging.GetModuleLogger("scanner")
scannerLogger.Info("Scanner initialized")
```

## Command Execution

### Package: `pkg/security`

#### Secure Command Pattern

```go
// Instead of dangerous:
cmd := exec.Command("nmap", userInput) // VULNERABLE!

// Use secure pattern:
validated, err := validator.ValidateIP(userInput)
if err != nil {
    return err
}

secureCmd, err := security.NewSecureCommand("nmap", []string{"-sS", validated}, security.SecureCommandOptions{
    Timeout:         30 * time.Second,
    AllowedCommands: []string{"nmap"},
    DisableShell:    true,
})
if err != nil {
    return err
}

err = secureCmd.Run()
```

## Key Storage

### Package: `pkg/security`

#### Secure Keystore

```go
type SecureKeyStore struct {
    filePath   string
    masterKey  []byte
    data       map[string]string
    mutex      sync.RWMutex
    gcm        cipher.AEAD
}

// Functions
func NewSecureKeyStore(filePath string, password string) (*SecureKeyStore, error)
func (ks *SecureKeyStore) Set(key, value string) error
func (ks *SecureKeyStore) Get(key string) (string, error)
func (ks *SecureKeyStore) Delete(key string) error
func (ks *SecureKeyStore) List() []string
func (ks *SecureKeyStore) Exists(key string) bool
func (ks *SecureKeyStore) ChangePassword(oldPassword, newPassword string) error
func (ks *SecureKeyStore) Backup(backupPath string) error
func (ks *SecureKeyStore) Wipe() error
```

**Examples:**
```go
// Create encrypted keystore
keystore, err := security.NewSecureKeyStore("/path/to/keys.enc", "strong-password")
if err != nil {
    return err
}

// Store API key
err = keystore.Set("shodan_api_key", "your-api-key")
if err != nil {
    return err
}

// Retrieve API key
apiKey, err := keystore.Get("shodan_api_key")
if err != nil {
    return err
}

// List available keys
keys := keystore.List()
for _, key := range keys {
    fmt.Printf("Available key: %s\n", key)
}
```

## Examples

### Complete Secure Scanning Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "GopherStrike/pkg/validator"
    "GopherStrike/pkg/security"
    "GopherStrike/pkg/resources"
    "GopherStrike/pkg/logging"
    "GopherStrike/pkg/errors"
)

func SecureScan(target string) error {
    // Initialize logging
    logger := logging.GetModuleLogger("scanner")
    
    // Create resource manager
    ctx := context.Background()
    resourceManager := resources.NewManager(ctx)
    defer resourceManager.Close()
    
    // Validate target
    validatedTarget, err := validator.ValidateIP(target)
    if err != nil {
        // Try as domain
        validatedTarget, err = validator.ValidateDomain(target)
        if err != nil {
            return errors.ValidationFailed("target", "must be valid IP or domain")
        }
    }
    
    logger.Info("Starting scan for validated target: %s", validatedTarget)
    
    // Create secure command
    options := security.SecureCommandOptions{
        Timeout:         5 * time.Minute,
        AllowedCommands: []string{"nmap"},
        DisableShell:    true,
    }
    
    cmd, err := security.NewSecureCommand("nmap", []string{
        "-sS",
        "-T4",
        validatedTarget,
    }, options)
    
    if err != nil {
        logger.Error("Failed to create secure command: %v", err)
        return errors.SystemFailed("command_creation", err)
    }
    
    // Execute scan
    output, err := cmd.Output()
    if err != nil {
        logger.Error("Scan execution failed: %v", err)
        return errors.SystemFailed("scan_execution", err)
    }
    
    logger.Info("Scan completed successfully for target: %s", validatedTarget)
    fmt.Printf("Scan results:\n%s\n", string(output))
    
    return nil
}
```

### Configuration Example

```go
package main

import (
    "GopherStrike/pkg/config"
    "GopherStrike/pkg/logging"
)

func InitializeApplication() error {
    // Initialize configuration
    configManager := config.NewConfigManager("~/.gopherstrike/config.json")
    if err := configManager.Initialize(); err != nil {
        return fmt.Errorf("failed to initialize config: %w", err)
    }
    
    cfg := configManager.Get()
    
    // Set up logging based on configuration
    logger := logging.New(logging.INFO)
    if cfg.Output.Verbose {
        logger.SetLevel(logging.DEBUG)
    }
    
    // Add file logging
    logPath := cfg.Output.LogDirectory + "/app.log"
    if err := logger.AddFileHandler(logPath, logging.INFO); err != nil {
        return fmt.Errorf("failed to setup logging: %w", err)
    }
    
    logger.Info("Application initialized with config from: %s", configManager.configFile)
    
    return nil
}
```

### Error Handling Example

```go
package main

import (
    "GopherStrike/pkg/errors"
    "GopherStrike/pkg/logging"
)

func HandleScanError(err error) {
    logger := logging.GetModuleLogger("error_handler")
    
    // Check error type and severity
    if errors.IsType(err, errors.SecurityError) {
        logger.Critical("Security error detected: %v", err)
        // Implement additional security measures
        return
    }
    
    severity := errors.GetSeverity(err)
    switch severity {
    case errors.SeverityCritical:
        logger.Critical("Critical error: %v", err)
        // Implement emergency procedures
    case errors.SeverityHigh:
        logger.Error("High severity error: %v", err)
        // Implement error recovery
    case errors.SeverityMedium:
        logger.Warning("Medium severity error: %v", err)
        // Log and continue
    case errors.SeverityLow:
        logger.Info("Low severity error: %v", err)
        // Log for debugging
    }
    
    // Get error context for additional information
    if context := errors.GetContext(err); context != nil {
        for key, value := range context {
            logger.Debug("Error context - %s: %v", key, value)
        }
    }
}
```

## Best Practices

1. **Always Validate Input**: Use the validator package for all user inputs
2. **Use Secure Commands**: Never use raw exec.Command with user input
3. **Handle Errors Properly**: Use the error handling framework with appropriate severity
4. **Manage Resources**: Use the resource manager to prevent leaks
5. **Log Securely**: Use the logging framework which automatically sanitizes sensitive data
6. **Store Keys Securely**: Use the encrypted keystore for all API keys and secrets
7. **Follow Principle of Least Privilege**: Use minimal required permissions

## Migration Guide

### From Raw Commands to Secure Commands

**Old (Vulnerable)**:
```go
cmd := exec.Command("nmap", userInput)
cmd.Run()
```

**New (Secure)**:
```go
validated, err := validator.ValidateIP(userInput)
if err != nil {
    return err
}

secureCmd, err := security.NewSecureCommand("nmap", []string{"-sS", validated}, security.SecureCommandOptions{
    AllowedCommands: []string{"nmap"},
    DisableShell:    true,
})
if err != nil {
    return err
}

err = secureCmd.Run()
```

### From Basic Logging to Secure Logging

**Old**:
```go
fmt.Printf("API Key: %s", apiKey) // Exposes sensitive data!
```

**New**:
```go
logger := logging.GetModuleLogger("app")
logger.Info("API Key configured") // Automatically sanitized
```

This documentation provides comprehensive coverage of all security-focused APIs in GopherStrike. All APIs are designed to be secure by default and include extensive validation and error handling.