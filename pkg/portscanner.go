package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// hasRequiredPrivileges checks if the current process has the required privileges
// for running the port scanner
func hasRequiredPrivileges() bool {
	if runtime.GOOS == "windows" {
		// On Windows, os.Geteuid() returns -1, so we can't use it
		// For now, we'll assume Windows users can handle privilege escalation
		// A more robust solution would use Windows APIs to check admin status
		return true
	}
	// On Unix-like systems, check if running as root (UID 0)
	return os.Geteuid() == 0
}

// RunPortScanner executes the Python port scanner
func RunPortScanner() error {
	fmt.Println("Initializing Port Scanner...")
	
	// Get the current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	
	// Look for the Python script
	scriptPath := filepath.Join(workDir, "NmapScript_copy.py")
	
	// Check if the script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("port scanner script not found at %s", scriptPath)
	}
	
	fmt.Printf("Found port scanner at: %s\n", scriptPath)
	fmt.Println("Launching Python port scanner...")
	fmt.Println("Note: This tool requires root privileges.")
	
	// Check if already running with appropriate privileges
	if !hasRequiredPrivileges() {
		if runtime.GOOS == "windows" {
			fmt.Println("Administrator privileges required. Please run GopherStrike as Administrator.")
			fmt.Println("Right-click Command Prompt and select 'Run as administrator'")
		} else {
			fmt.Println("Root privileges required. Please run GopherStrike with sudo.")
			fmt.Println("Example: sudo ./GopherStrike")
		}
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return nil
	}
	
	// Execute the Python script with proper environment
	cmd := exec.Command("python3", scriptPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run port scanner: %v", err)
	}
	
	fmt.Println("\nPort scanner completed.")
	fmt.Println("Press Enter to continue...")
	fmt.Scanln()
	
	return nil
}