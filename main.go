package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// findScript tries to locate the script in multiple possible locations
func findScript(scriptName string) (string, error) {
	// List of possible paths to search
	searchPaths := []string{
		// Mac specific path
		"/Users/leog/GolandProjects/GopherStrike/scripts",

		// Relative to current directory
		filepath.Join(".", "scripts"),
		filepath.Join(".", scriptName),

		// Relative to executable
		getExecutableDirPath("scripts"),
	}

	// Search each path
	for _, basePath := range searchPaths {
		path := filepath.Join(basePath, scriptName)
		if fileExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("Script %s not found in any of the expected locations", scriptName)
}

// getExecutableDirPath returns a path relative to the executable directory
func getExecutableDirPath(subPath string) string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(execPath), subPath)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// getPythonCommand determines the appropriate Python command for the current platform
func getPythonCommand() string {
	// Try python3 first
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}

	// Fallback to python (which might be python3 on some systems)
	if _, err := exec.LookPath("python"); err == nil {
		return "python"
	}

	// Windows-specific python executable names
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("py"); err == nil {
			return "py"
		}
	}

	// Default to python and let the system report the error if it's missing
	return "python"
}

// setupVirtualEnv creates a virtual environment for Python if not exists
func setupVirtualEnv() (string, error) {
	// Define virtual environment path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	venvPath := filepath.Join(homeDir, ".gopherstrike_venv")
	venvBinPath := filepath.Join(venvPath, "bin")

	// Check if virtual environment already exists
	if fileExists(filepath.Join(venvPath, "bin", "python")) ||
		fileExists(filepath.Join(venvPath, "Scripts", "python.exe")) {
		fmt.Println("Using existing virtual environment.")
	} else {
		// Create virtual environment
		fmt.Println("Creating Python virtual environment...")
		pythonCmd := getPythonCommand()
		cmd := exec.Command(pythonCmd, "-m", "venv", venvPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to create virtual environment: %v", err)
		}
		fmt.Println("Virtual environment created successfully.")
	}

	// Return the path to the Python executable in the virtual environment
	if runtime.GOOS == "windows" {
		return filepath.Join(venvPath, "Scripts", "python.exe"), nil
	}
	return filepath.Join(venvBinPath, "python"), nil
}

// checkPythonModule checks if a Python module is installed in the virtual environment
func checkPythonModule(pythonCmd, module string) bool {
	// Map Python import names to their actual module names
	moduleImportMap := map[string]string{
		"nmap":  "nmap",
		"scapy": "scapy",
	}

	// Get the import name for the module
	importName := moduleImportMap[module]

	cmd := exec.Command(pythonCmd, "-c", fmt.Sprintf("import %s", importName))
	return cmd.Run() == nil
}

// getPipPackageName maps a module import name to its pip package name
func getPipPackageName(module string) string {
	// Map module names to their pip package names
	packageMap := map[string]string{
		"nmap":  "python-nmap",
		"scapy": "scapy",
	}

	if pkgName, exists := packageMap[module]; exists {
		return pkgName
	}
	return module
}

// installPythonModule installs a Python module in the virtual environment
func installPythonModule(pythonCmd, module string) error {
	// Get the correct pip package name
	pipPackage := getPipPackageName(module)
	fmt.Printf("Installing Python module: %s (package: %s)\n", module, pipPackage)

	// Use the Python in the virtual environment to run pip
	cmd := exec.Command(pythonCmd, "-m", "pip", "install", pipPackage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// checkAndInstallDependencies ensures all required Python modules are installed
func checkAndInstallDependencies(pythonCmd string) error {
	// List of required modules for NmapScript.py
	requiredModules := []string{"nmap", "scapy"}

	missingModules := []string{}

	// Check for missing modules
	for _, module := range requiredModules {
		if !checkPythonModule(pythonCmd, module) {
			missingModules = append(missingModules, module)
		}
	}

	// If no missing modules, return nil
	if len(missingModules) == 0 {
		return nil
	}

	// Ask user if they want to install missing modules
	fmt.Printf("The following Python modules are required but not installed: %s\n", strings.Join(missingModules, ", "))
	fmt.Print("Do you want to install them now? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		return fmt.Errorf("cannot continue without required Python modules")
	}

	// Install missing modules
	for _, module := range missingModules {
		if err := installPythonModule(pythonCmd, module); err != nil {
			return fmt.Errorf("failed to install %s: %v", module, err)
		}
		fmt.Printf("Successfully installed %s\n", module)
	}

	return nil
}

// createModifiedScriptForSudo creates a temporary version of the script with the no-sudo flag
func createModifiedScriptForSudo(originalScript string) (string, error) {
	// Create a directory for temporary files if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "gopherstrike")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Read the original script
	content, err := os.ReadFile(originalScript)
	if err != nil {
		return "", fmt.Errorf("failed to read script: %v", err)
	}

	// Modify the script to disable sudo checks
	modifiedContent := []byte(
		"# Modified by GopherStrike to run without sudo\n" +
			"import os\ndef check_root():\n    # Disable sudo check\n    pass\n\n" +
			string(content))

	// Write the modified script to a temporary file
	tempScript := filepath.Join(tempDir, "nmap_no_sudo.py")
	if err := os.WriteFile(tempScript, modifiedContent, 0755); err != nil {
		return "", fmt.Errorf("failed to write temp script: %v", err)
	}

	return tempScript, nil
}

func main() {
	fmt.Println("Welcome to GopherStrike")

	// Find the Python script
	scriptPath, err := findScript("NmapScript.py")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	// Set up virtual environment
	venvPythonCmd, err := setupVirtualEnv()
	if err != nil {
		fmt.Printf("Error setting up virtual environment: %s\n", err)
		return
	}

	// Check and install dependencies in the virtual environment
	if err := checkAndInstallDependencies(venvPythonCmd); err != nil {
		fmt.Printf("Dependency error: %s\n", err)
		return
	}

	// Determine whether to use sudo directly or run through terminal
	var cmd *exec.Cmd
	var cmdString string

	// Create a modified version of the script without sudo check
	tempScript, err := createModifiedScriptForSudo(scriptPath)
	if err != nil {
		fmt.Printf("Error preparing script: %s\n", err)
		return
	}

	fmt.Println("This program requires elevated privileges for scanning.")
	fmt.Println("Please enter your password when prompted in the terminal window.")

	// Prepare sudo command for terminal
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		// On macOS or Linux, use a terminal to run sudo
		cmdString = fmt.Sprintf("sudo %s %s", venvPythonCmd, tempScript)

		var terminalCmd string
		var terminalArgs []string

		if runtime.GOOS == "darwin" {
			// Use Terminal.app on macOS
			terminalCmd = "osascript"
			terminalArgs = []string{
				"-e",
				fmt.Sprintf(`tell application "Terminal" to do script "%s"`, cmdString),
			}
		} else {
			// Use xterm or gnome-terminal on Linux
			if _, err := exec.LookPath("gnome-terminal"); err == nil {
				terminalCmd = "gnome-terminal"
				terminalArgs = []string{"--", "bash", "-c", cmdString + "; echo 'Press Enter to close'; read"}
			} else if _, err := exec.LookPath("xterm"); err == nil {
				terminalCmd = "xterm"
				terminalArgs = []string{"-e", "bash", "-c", cmdString + "; echo 'Press Enter to close'; read"}
			} else {
				fmt.Println("No supported terminal found. Please run the script manually with sudo.")
				fmt.Printf("Command to run: sudo %s %s\n", venvPythonCmd, scriptPath)
				return
			}
		}

		cmd = exec.Command(terminalCmd, terminalArgs...)
	} else if runtime.GOOS == "windows" {
		// On Windows, use cmd.exe to run as admin
		cmdString = fmt.Sprintf(`%s %s`, venvPythonCmd, tempScript)
		cmd = exec.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", cmdString)
	} else {
		fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
		return
	}

	// Execute the command through terminal
	fmt.Printf("Starting NmapScript.py in a terminal window...\n")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error executing terminal command: %s\n", err)
		return
	}

	fmt.Println("Terminal window has been opened to run the script with sudo privileges.")
	fmt.Println("You can now return to this window after completing the scan.")
}
