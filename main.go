package main

import (
	"GopherStrike/pkg" // Import the pkg package to access exported functions
	"GopherStrike/utils"
	"fmt"
	"os"
)

// displayBanner prints the GopherStrike ASCII art banner
func displayBanner() {
	banner := `
    ██████╗  ██████╗ ██████╗ ██╗  ██╗███████╗██████╗ ███████╗████████╗██████╗ ██╗██╗  ██╗███████╗
    ██╔════╝ ██╔═══██╗██╔══██╗██║  ██║██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔══██╗██║██║ ██╔╝██╔════╝
    ██║  ███╗██║   ██║██████╔╝███████║█████╗  ██████╔╝███████╗   ██║   ██████╔╝██║█████╔╝ █████╗  
    ██║   ██║██║   ██║██╔═══╝ ██╔══██║██╔══╝  ██╔══██╗╚════██║   ██║   ██╔══██╗██║██╔═██╗ ██╔══╝  
    ╚██████╔╝╚██████╔╝██║     ██║  ██║███████╗██║  ██║███████║   ██║   ██║  ██║██║██║  ██╗███████╗
     ╚═════╝  ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝
    `
	fmt.Println(banner)
}

// mainMenu displays and handles the main application menu
func mainMenu() {
	displayBanner() // this will have to get changed around
	fmt.Println("\nAvailable Tools:")
	fmt.Println("================")
	fmt.Println("1. Port Scanner")
	fmt.Println("2. Subdomain Scanner")
	fmt.Println("3. OSINT & Vulnerability Tool")
	fmt.Println("4. Check Dependencies")
	fmt.Println("5. Exit")

	var choice string
	fmt.Print("\nSelect a tool: ")
	if _, err := fmt.Scanln(&choice); err != nil {
		fmt.Println("Error reading input:", err)
		utils.ClearScreen()
		mainMenu()
		return
	}

	switch choice {
	case "1":
		// Use the properly exported function from the pkg package
		if err := pkg.RunNmapScannerWithPrivCheck(); err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln() // Ignoring error here is fine for user interaction
		utils.ClearScreen()
		mainMenu() // Return to main menu after tool completes
	case "2":
		// Run subdomain scanner
		if err := pkg.RunSubdomainScannerWithCheck(); err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		utils.ClearScreen()
		mainMenu()
	case "3":
		// Run OSINT tool
		if err := pkg.RunOSINTTool(); err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		utils.ClearScreen()
		mainMenu()
	case "4":
		// Run dependency check
		pkg.PrintDependencyStatus()
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		utils.ClearScreen()
		mainMenu()
	case "5":
		fmt.Println("Exiting GopherStrike. Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid choice, please try again")
		utils.ClearScreen()
		mainMenu()
	}
}

// main is the entry point for the application
func main() {
	utils.ClearScreen() // clears the screen for the UI

	// Check for logs directory at startup
	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
	}

	// Create OSINT logs directory
	if err := os.MkdirAll("logs/osint", 0755); err != nil {
		fmt.Printf("Warning: Failed to create OSINT logs directory: %v\n", err)
	}

	mainMenu()
}
