// Package pkg provides network scanning functionality and tools for the GopherStrike framework
package pkg

import (
	"GopherStrike/pkg/security"
	"GopherStrike/pkg/validator"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CheckRoot determines if the program is running with elevated privileges
// Returns true if running as root/admin, false otherwise
func CheckRoot() bool {
	// On Unix-like systems, root has UID 0
	if runtime.GOOS != "windows" && os.Geteuid() != 0 {
		return false
	}
	return true
}

// RunNmapScannerWithPrivCheck executes the nmap scanner with appropriate privileges
// It handles privilege escalation differently based on the operating system
func RunNmapScannerWithPrivCheck() error {
	if !CheckRoot() {
		if runtime.GOOS == "darwin" {
			// Get the path to the Python in the virtual environment
			venvPythonPath := "./.venv/bin/python3"
			_, err := os.Stat(venvPythonPath)
			if err == nil {
				// Virtual env Python exists, use it
				scriptPath, err := filepath.Abs("pkg/tools/NmapScript.py")
				if err != nil {
					return fmt.Errorf("error getting script path: %w", err)
				}

				// Get IP target first since osascript won't pass stdin
				fmt.Print("Enter target IP: ")
				var targetIPInput string
				fmt.Scanln(&targetIPInput)
				
				// Validate IP address to prevent injection
				targetIP, err := validator.ValidateIP(targetIPInput)
				if err != nil {
					return fmt.Errorf("invalid IP address: %w", err)
				}

				// Get port range
				fmt.Println("\nSelect port range:")
				fmt.Println("1. Common ports (1-1024)")
				fmt.Println("2. Extended range (1-5000)")
				fmt.Println("3. Full range (1-65535)")
				fmt.Println("4. Custom range")
				fmt.Print("\nEnter choice (1-4): ")
				var portChoice string
				fmt.Scanln(&portChoice)

				var portArgs string
				if portChoice == "4" {
					fmt.Print("Enter start port: ")
					var startPortInput string
					fmt.Scanln(&startPortInput)

					fmt.Print("Enter end port: ")
					var endPortInput string
					fmt.Scanln(&endPortInput)
					
					// Validate ports
					startPort, err := validator.ValidatePort(startPortInput)
					if err != nil {
						return fmt.Errorf("invalid start port: %w", err)
					}
					endPort, err := validator.ValidatePort(endPortInput)
					if err != nil {
						return fmt.Errorf("invalid end port: %w", err)
					}

					portArgs = fmt.Sprintf("--port-range %s %s", startPort, endPort)
				} else {
					portArgs = fmt.Sprintf("--port-choice %s", portChoice)
				}

				fmt.Println("Launching with admin privileges using virtual environment...")
				absVenvPath, _ := filepath.Abs(venvPythonPath)
				// Create logs directory in advance to avoid permission issues
				if err := os.MkdirAll("logs", 0755); err != nil {
					fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
				}

				// Use secure command execution
				pem := security.NewPrivilegeEscalationManager()
				cmdArgs := []string{scriptPath, "--target", targetIP}
				
				// Add port arguments safely
				if strings.Contains(portArgs, "--port-range") {
					parts := strings.Fields(portArgs)
					for _, part := range parts {
						if part != "" {
							cmdArgs = append(cmdArgs, part)
						}
					}
				} else if strings.Contains(portArgs, "--port-choice") {
					parts := strings.Fields(portArgs)
					for _, part := range parts {
						if part != "" {
							cmdArgs = append(cmdArgs, part)
						}
					}
				}
				
				options := security.SecureCommandOptions{
					Timeout:         5 * time.Minute,
					AllowedCommands: []string{"python3", "python"},
					DisableShell:    true,
				}
				
				if err := pem.ExecuteWithElevatedPrivileges(absVenvPath, cmdArgs, options); err != nil {
					return fmt.Errorf("error running with admin privileges: %w", err)
				}

				// Show the scan results after the scan completes
				summaryFile := fmt.Sprintf("logs/lastscan_%s.txt", targetIP)
				if _, err := os.Stat(summaryFile); err == nil {
					fmt.Println("\nScan Results:")
					fmt.Println("=============")

					data, err := os.ReadFile(summaryFile)
					if err != nil {
						fmt.Printf("Error reading scan results: %v\n", err)
					} else {
						fmt.Println(string(data))
					}
				}

				return nil
			}

			// Fallback to system Python if venv not found
			scriptPath, err := filepath.Abs("pkg/tools/NmapScript.py")
			if err != nil {
				return fmt.Errorf("error getting absolute path: %w", err)
			}

			fmt.Println("Launching with admin privileges...")
			pem := security.NewPrivilegeEscalationManager()
			options := security.SecureCommandOptions{
				Timeout:         5 * time.Minute,
				AllowedCommands: []string{"python3", "python"},
				DisableShell:    true,
			}
			
			err = pem.ExecuteWithElevatedPrivileges("python3", []string{scriptPath}, options)
			if err != nil {
				// Provide specific advice for common errors
				if strings.Contains(err.Error(), "No module named") {
					fmt.Println("\nMissing Python dependencies. Please install them with:")
					fmt.Println("sudo pip3 install python-nmap scapy")
					return fmt.Errorf("missing dependencies: %w", err)
				}
				return fmt.Errorf("error running with admin privileges: %w", err)
			}
			return nil
		} else if runtime.GOOS == "linux" {
			// Use secure privilege escalation for Linux
			scriptPath, err := filepath.Abs("pkg/tools/NmapScript.py")
			if err != nil {
				return fmt.Errorf("error getting absolute path: %w", err)
			}

			venvPythonPath := "./.venv/bin/python3"
			pythonPath := "python3"
			if _, err := os.Stat(venvPythonPath); err == nil {
				absVenvPath, _ := filepath.Abs(venvPythonPath)
				pythonPath = absVenvPath
			}

			fmt.Println("Launching with admin privileges...")
			pem := security.NewPrivilegeEscalationManager()
			options := security.SecureCommandOptions{
				Timeout:         5 * time.Minute,
				AllowedCommands: []string{"python3", "python"},
				DisableShell:    true,
			}
			
			if err := pem.ExecuteWithElevatedPrivileges(pythonPath, []string{scriptPath}, options); err != nil {
				return fmt.Errorf("error running with elevated privileges: %w", err)
			}
			return nil

			// Fallback to terminal sudo method
			fmt.Println("Please run from terminal: sudo go run main.go")
			return fmt.Errorf("cannot elevate privileges in this environment")
		} else if runtime.GOOS == "windows" {
			// Windows-specific elevation
			scriptPath, err := filepath.Abs("pkg/tools/NmapScript.py")
			if err != nil {
				return fmt.Errorf("error getting absolute path: %w", err)
			}

			// Use secure privilege escalation for Windows
			pem := security.NewPrivilegeEscalationManager()
			options := security.SecureCommandOptions{
				Timeout:         5 * time.Minute,
				AllowedCommands: []string{"python", "python3"},
				DisableShell:    true,
			}
			
			if err := pem.ExecuteWithElevatedPrivileges("python", []string{scriptPath}, options); err != nil {
				return fmt.Errorf("error running with admin privileges: %w", err)
			}
			return nil
		} else {
			// Other OS implementations...
			fmt.Println("Please run from terminal: sudo go run main.go")
			return fmt.Errorf("cannot elevate privileges in this environment")
		}
	}

	// We have root already, check for venv first
	venvPythonPath := "./.venv/bin/python3"
	pythonToUse := "python3"

	if _, err := os.Stat(venvPythonPath); err == nil {
		absVenvPath, _ := filepath.Abs(venvPythonPath)
		pythonToUse = absVenvPath
		fmt.Println("Using Python from virtual environment")
	}

	// Create logs directory in advance to avoid permission issues
	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
	}

	// Get working directory to ensure correct path resolution
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %w", err)
	}

	// Use secure command execution for all platforms
	scriptPath := "pkg/tools/NmapScript.py"
	
	// Use secure privilege escalation
	pem := security.NewPrivilegeEscalationManager()
	options := security.SecureCommandOptions{
		Timeout:            5 * time.Minute,
		WorkingDirectory:   workDir,
		AllowedCommands:    []string{"python3", "python"},
		DisableShell:       true,
	}
	
	if err := pem.ExecuteWithElevatedPrivileges(pythonToUse, []string{scriptPath}, options); err != nil {
		return fmt.Errorf("error running port scanner: %w", err)
	}

	return nil
}
