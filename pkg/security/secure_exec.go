// Package security provides secure execution and operations for GopherStrike
package security

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// SecureCommandOptions contains options for secure command execution
type SecureCommandOptions struct {
	// WorkingDirectory sets the working directory for the command
	WorkingDirectory string
	
	// Environment variables to set (in addition to inherited ones)
	Environment map[string]string
	
	// Timeout for command execution
	Timeout time.Duration
	
	// AllowedCommands is a whitelist of allowed command names
	AllowedCommands []string
	
	// RequireAbsolutePath requires commands to be specified with absolute paths
	RequireAbsolutePath bool
	
	// DisableShell prevents shell interpretation
	DisableShell bool
	
	// User and Group IDs to run as (Unix only)
	UID *int
	GID *int
}

// SecureCommand represents a securely configured command
type SecureCommand struct {
	cmd     *exec.Cmd
	options SecureCommandOptions
	args    []string
}

// NewSecureCommand creates a new secure command with validation
func NewSecureCommand(name string, args []string, options SecureCommandOptions) (*SecureCommand, error) {
	// Default options
	if options.Timeout == 0 {
		options.Timeout = 30 * time.Second
	}
	
	if options.DisableShell {
		// Ensure no shell metacharacters in command or args
		if err := validateNoShellMetachars(name); err != nil {
			return nil, fmt.Errorf("command name contains shell metacharacters: %w", err)
		}
		
		for i, arg := range args {
			if err := validateNoShellMetachars(arg); err != nil {
				return nil, fmt.Errorf("argument %d contains shell metacharacters: %w", i, err)
			}
		}
	}
	
	// Validate command is in whitelist if specified
	if len(options.AllowedCommands) > 0 {
		allowed := false
		baseName := filepath.Base(name)
		for _, allowedCmd := range options.AllowedCommands {
			if baseName == allowedCmd || name == allowedCmd {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, fmt.Errorf("command '%s' is not in the allowed commands list", name)
		}
	}
	
	// Validate absolute path if required
	if options.RequireAbsolutePath && !filepath.IsAbs(name) {
		return nil, fmt.Errorf("command must be specified with absolute path: %s", name)
	}
	
	// Create command with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	cmd := exec.CommandContext(ctx, name, args...)
	
	// Set working directory
	if options.WorkingDirectory != "" {
		// Validate working directory exists and is safe
		if err := validateWorkingDirectory(options.WorkingDirectory); err != nil {
			cancel()
			return nil, fmt.Errorf("invalid working directory: %w", err)
		}
		cmd.Dir = options.WorkingDirectory
	}
	
	// Set environment
	cmd.Env = os.Environ()
	for key, value := range options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	// Set user/group on Unix systems
	if runtime.GOOS != "windows" && (options.UID != nil || options.GID != nil) {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		if options.UID != nil {
			cmd.SysProcAttr.Credential = &syscall.Credential{
				Uid: uint32(*options.UID),
			}
		}
		if options.GID != nil {
			if cmd.SysProcAttr.Credential == nil {
				cmd.SysProcAttr.Credential = &syscall.Credential{}
			}
			cmd.SysProcAttr.Credential.Gid = uint32(*options.GID)
		}
	}
	
	// Store cancel function for cleanup
	secureCmd := &SecureCommand{
		cmd:     cmd,
		options: options,
		args:    append([]string{name}, args...),
	}
	
	// Set up automatic cleanup
	go func() {
		<-ctx.Done()
		cancel()
	}()
	
	return secureCmd, nil
}

// Run executes the command and waits for completion
func (sc *SecureCommand) Run() error {
	return sc.cmd.Run()
}

// Start starts the command but doesn't wait for completion
func (sc *SecureCommand) Start() error {
	return sc.cmd.Start()
}

// Wait waits for the command to complete
func (sc *SecureCommand) Wait() error {
	return sc.cmd.Wait()
}

// Output runs the command and returns its standard output
func (sc *SecureCommand) Output() ([]byte, error) {
	return sc.cmd.Output()
}

// CombinedOutput runs the command and returns its combined output
func (sc *SecureCommand) CombinedOutput() ([]byte, error) {
	return sc.cmd.CombinedOutput()
}

// Kill terminates the command
func (sc *SecureCommand) Kill() error {
	if sc.cmd.Process != nil {
		return sc.cmd.Process.Kill()
	}
	return nil
}

// validateNoShellMetachars ensures no shell metacharacters are present
func validateNoShellMetachars(input string) error {
	// Common shell metacharacters that could be dangerous
	dangerous := []string{
		";", "&", "|", "&&", "||", "$", "`", "$(", "${", ">", "<", ">>", "<<",
		"*", "?", "[", "]", "{", "}", "~", "!", "#", "\n", "\r", "\t",
	}
	
	for _, char := range dangerous {
		if strings.Contains(input, char) {
			return fmt.Errorf("contains dangerous character: %s", char)
		}
	}
	
	return nil
}

// validateWorkingDirectory ensures the working directory is safe
func validateWorkingDirectory(dir string) error {
	// Clean the path
	cleanDir := filepath.Clean(dir)
	
	// Check if directory exists
	info, err := os.Stat(cleanDir)
	if err != nil {
		return fmt.Errorf("directory does not exist: %w", err)
	}
	
	// Ensure it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", cleanDir)
	}
	
	// Check permissions (Unix only)
	if runtime.GOOS != "windows" {
		// Ensure directory is readable and executable
		if info.Mode().Perm()&0500 != 0500 {
			return fmt.Errorf("directory lacks required permissions: %s", cleanDir)
		}
	}
	
	return nil
}

// PrivilegeEscalationManager handles secure privilege escalation
type PrivilegeEscalationManager struct {
	method EscalationMethod
}

// EscalationMethod represents different privilege escalation methods
type EscalationMethod int

const (
	EscalationSudo EscalationMethod = iota
	EscalationPkexec
	EscalationOSAScript
	EscalationPowerShell
)

// NewPrivilegeEscalationManager creates a new privilege escalation manager
func NewPrivilegeEscalationManager() *PrivilegeEscalationManager {
	method := detectBestEscalationMethod()
	return &PrivilegeEscalationManager{method: method}
}

// detectBestEscalationMethod detects the best privilege escalation method for the current system
func detectBestEscalationMethod() EscalationMethod {
	switch runtime.GOOS {
	case "darwin":
		// Check if osascript is available
		if _, err := exec.LookPath("osascript"); err == nil {
			return EscalationOSAScript
		}
		return EscalationSudo
	case "linux":
		// Prefer pkexec if available
		if _, err := exec.LookPath("pkexec"); err == nil {
			return EscalationPkexec
		}
		return EscalationSudo
	case "windows":
		return EscalationPowerShell
	default:
		return EscalationSudo
	}
}

// ExecuteWithElevatedPrivileges executes a command with elevated privileges securely
func (pem *PrivilegeEscalationManager) ExecuteWithElevatedPrivileges(command string, args []string, options SecureCommandOptions) error {
	// Validate all inputs first
	if err := validateNoShellMetachars(command); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}
	
	for i, arg := range args {
		if err := validateNoShellMetachars(arg); err != nil {
			return fmt.Errorf("invalid argument %d: %w", i, err)
		}
	}
	
	switch pem.method {
	case EscalationOSAScript:
		return pem.executeWithOSAScript(command, args, options)
	case EscalationPkexec:
		return pem.executeWithPkexec(command, args, options)
	case EscalationSudo:
		return pem.executeWithSudo(command, args, options)
	case EscalationPowerShell:
		return pem.executeWithPowerShell(command, args, options)
	default:
		return fmt.Errorf("unsupported escalation method")
	}
}

// executeWithOSAScript executes with macOS osascript
func (pem *PrivilegeEscalationManager) executeWithOSAScript(command string, args []string, options SecureCommandOptions) error {
	// Build command array securely
	fullCmd := append([]string{command}, args...)
	
	// Create the AppleScript command securely
	// We use separate arguments instead of string concatenation
	osascriptArgs := []string{
		"-e",
		fmt.Sprintf("do shell script %q with administrator privileges", strings.Join(fullCmd, " ")),
	}
	
	secureCmd, err := NewSecureCommand("osascript", osascriptArgs, SecureCommandOptions{
		Timeout:         options.Timeout,
		AllowedCommands: []string{"osascript"},
		DisableShell:    true,
	})
	
	if err != nil {
		return err
	}
	
	return secureCmd.Run()
}

// executeWithPkexec executes with Linux pkexec
func (pem *PrivilegeEscalationManager) executeWithPkexec(command string, args []string, options SecureCommandOptions) error {
	pkexecArgs := append([]string{command}, args...)
	
	secureCmd, err := NewSecureCommand("pkexec", pkexecArgs, SecureCommandOptions{
		Timeout:         options.Timeout,
		AllowedCommands: []string{"pkexec"},
		DisableShell:    true,
	})
	
	if err != nil {
		return err
	}
	
	return secureCmd.Run()
}

// executeWithSudo executes with sudo
func (pem *PrivilegeEscalationManager) executeWithSudo(command string, args []string, options SecureCommandOptions) error {
	sudoArgs := append([]string{command}, args...)
	
	secureCmd, err := NewSecureCommand("sudo", sudoArgs, SecureCommandOptions{
		Timeout:         options.Timeout,
		AllowedCommands: []string{"sudo"},
		DisableShell:    true,
	})
	
	if err != nil {
		return err
	}
	
	return secureCmd.Run()
}

// executeWithPowerShell executes with Windows PowerShell
func (pem *PrivilegeEscalationManager) executeWithPowerShell(command string, args []string, options SecureCommandOptions) error {
	// For Windows, we need to handle PowerShell elevation differently
	powershellArgs := []string{
		"Start-Process",
		command,
		"-ArgumentList",
		fmt.Sprintf("'%s'", strings.Join(args, "', '")),
		"-Verb",
		"RunAs",
		"-Wait",
	}
	
	secureCmd, err := NewSecureCommand("powershell", powershellArgs, SecureCommandOptions{
		Timeout:         options.Timeout,
		AllowedCommands: []string{"powershell"},
		DisableShell:    true,
	})
	
	if err != nil {
		return err
	}
	
	return secureCmd.Run()
}

// IsElevated checks if the current process is running with elevated privileges
func IsElevated() bool {
	switch runtime.GOOS {
	case "windows":
		// On Windows, check if running as administrator
		return isWindowsElevated()
	default:
		// On Unix-like systems, check if running as root
		return os.Geteuid() == 0
	}
}

// isWindowsElevated checks if running with elevated privileges on Windows
func isWindowsElevated() bool {
	// This is a simplified check - in a real implementation, you'd use Windows APIs
	// For now, we'll assume non-elevated since we can't easily check without CGO
	return false
}