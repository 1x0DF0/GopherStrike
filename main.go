package main

import (
	"GopherStrike/pkg" // Import the pkg package to access exported functions
	"GopherStrike/pkg/tools"
	"GopherStrike/utils"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// Global variable to track if we're currently in a tool
var inTool bool = false

// ASCII art for each tool
var (
	subdomainScannerArt = `
    ███████╗██╗   ██╗██████╗ ██████╗  ██████╗ ███╗   ███╗ █████╗ ██╗███╗   ██╗    ███████╗ ██████╗ █████╗ ███╗   ██╗███╗   ██╗███████╗██████╗ 
    ██╔════╝██║   ██║██╔══██╗██╔══██╗██╔═══██╗████╗ ████║██╔══██╗██║████╗  ██║    ██╔════╝██╔════╝██╔══██╗████╗  ██║████╗  ██║██╔════╝██╔══██╗
    ███████╗██║   ██║██████╔╝██║  ██║██║   ██║██╔████╔██║███████║██║██╔██╗ ██║    ███████╗██║     ███████║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
    ╚════██║██║   ██║██╔══██╗██║  ██║██║   ██║██║╚██╔╝██║██╔══██║██║██║╚██╗██║    ╚════██║██║     ██╔══██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
    ███████║╚██████╔╝██████╔╝██████╔╝╚██████╔╝██║ ╚═╝ ██║██║  ██║██║██║ ╚████║    ███████║╚██████╗██║  ██║██║ ╚████║██║ ╚████║███████╗██║  ██║
    ╚══════╝ ╚═════╝ ╚═════╝ ╚═════╝  ╚═════╝ ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝    ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝
    `

	osintArt = `
     ██████╗ ███████╗██╗███╗   ██╗████████╗    ████████╗ ██████╗  ██████╗ ██╗     
    ██╔═══██╗██╔════╝██║████╗  ██║╚══██╔══╝    ╚══██╔══╝██╔═══██╗██╔═══██╗██║     
    ██║   ██║███████╗██║██╔██╗ ██║   ██║          ██║   ██║   ██║██║   ██║██║     
    ██║   ██║╚════██║██║██║╚██╗██║   ██║          ██║   ██║   ██║██║   ██║██║     
    ╚██████╔╝███████║██║██║ ╚████║   ██║          ██║   ╚██████╔╝╚██████╔╝███████╗
     ╚═════╝ ╚══════╝╚═╝╚═╝  ╚═══╝   ╚═╝          ╚═╝    ╚═════╝  ╚═════╝ ╚══════╝
    `

	webVulnArt = `
    ██╗    ██╗███████╗██████╗     ██╗   ██╗██╗   ██╗██╗     ███╗   ██╗    ███████╗ ██████╗ █████╗ ███╗   ██╗███╗   ██╗███████╗██████╗ 
    ██║    ██║██╔════╝██╔══██╗    ██║   ██║██║   ██║██║     ████╗  ██║    ██╔════╝██╔════╝██╔══██╗████╗  ██║████╗  ██║██╔════╝██╔══██╗
    ██║ █╗ ██║█████╗  ██████╔╝    ██║   ██║██║   ██║██║     ██╔██╗ ██║    ███████╗██║     ███████║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
    ██║███╗██║██╔══╝  ██╔══██╗    ╚██╗ ██╔╝██║   ██║██║     ██║╚██╗██║    ╚════██║██║     ██╔══██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
    ╚███╔███╔╝███████╗██████╔╝     ╚████╔╝ ╚██████╔╝███████╗██║ ╚████║    ███████║╚██████╗██║  ██║██║ ╚████║██║ ╚████║███████╗██║  ██║
     ╚══╝╚══╝ ╚══════╝╚═════╝       ╚═══╝   ╚═════╝ ╚══════╝╚═╝  ╚═══╝    ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝
    `

	s3ScannerArt = `
    ███████╗██████╗     ██████╗ ██╗   ██╗ ██████╗██╗  ██╗███████╗████████╗    ███████╗ ██████╗ █████╗ ███╗   ██╗███╗   ██╗███████╗██████╗ 
    ██╔════╝╚════██╗    ██╔══██╗██║   ██║██╔════╝██║ ██╔╝██╔════╝╚══██╔══╝    ██╔════╝██╔════╝██╔══██╗████╗  ██║████╗  ██║██╔════╝██╔══██╗
    ███████╗ █████╔╝    ██████╔╝██║   ██║██║     █████╔╝ █████╗     ██║       ███████╗██║     ███████║██╔██╗ ██║██╔██╗ ██║█████╗  ██████╔╝
    ╚════██║ ╚═══██╗    ██╔══██╗██║   ██║██║     ██╔═██╗ ██╔══╝     ██║       ╚════██║██║     ██╔══██║██║╚██╗██║██║╚██╗██║██╔══╝  ██╔══██╗
    ███████║██████╔╝    ██████╔╝╚██████╔╝╚██████╗██║  ██╗███████╗   ██║       ███████║╚██████╗██║  ██║██║ ╚████║██║ ╚████║███████╗██║  ██║
    ╚══════╝╚═════╝     ╚═════╝  ╚═════╝  ╚═════╝╚═╝  ╚═╝╚══════╝   ╚═╝       ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝
    `

	emailHarvesterArt = `
    ███████╗███╗   ███╗ █████╗ ██╗██╗         ██╗  ██╗ █████╗ ██████╗ ██╗   ██╗███████╗███████╗████████╗███████╗██████╗ 
    ██╔════╝████╗ ████║██╔══██╗██║██║         ██║  ██║██╔══██╗██╔══██╗██║   ██║██╔════╝██╔════╝╚══██╔══╝██╔════╝██╔══██╗
    █████╗  ██╔████╔██║███████║██║██║         ███████║███████║██████╔╝██║   ██║█████╗  ███████╗   ██║   █████╗  ██████╔╝
    ██╔══╝  ██║╚██╔╝██║██╔══██║██║██║         ██╔══██║██╔══██║██╔══██╗╚██╗ ██╔╝██╔══╝  ╚════██║   ██║   ██╔══╝  ██╔══██╗
    ███████╗██║ ╚═╝ ██║██║  ██║██║███████╗    ██║  ██║██║  ██║██║  ██║ ╚████╔╝ ███████╗███████║   ██║   ███████╗██║  ██║
    ╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚══════╝    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
    `

	dirBruteforceArt = `
    ██████╗ ██╗██████╗     ██████╗ ██████╗ ██╗   ██╗████████╗███████╗███████╗ ██████╗ ██████╗  ██████╗███████╗
    ██╔══██╗██║██╔══██╗    ██╔══██╗██╔══██╗██║   ██║╚══██╔══╝██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝██╔════╝
    ██║  ██║██║██████╔╝    ██████╔╝██████╔╝██║   ██║   ██║   █████╗  █████╗  ██║   ██║██████╔╝██║     █████╗  
    ██║  ██║██║██╔══██╗    ██╔══██╗██╔══██╗██║   ██║   ██║   ██╔══╝  ██╔══╝  ██║   ██║██╔══██╗██║     ██╔══╝  
    ██████╔╝██║██║  ██║    ██████╔╝██║  ██║╚██████╔╝   ██║   ███████╗██║     ╚██████╔╝██║  ██║╚██████╗███████╗
    ╚═════╝ ╚═╝╚═╝  ╚═╝    ╚═════╝ ╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝╚══════╝
    `

	reportGeneratorArt = `
    ██████╗ ███████╗██████╗  ██████╗ ██████╗ ████████╗     ██████╗ ███████╗███╗   ██╗███████╗██████╗  █████╗ ████████╗ ██████╗ ██████╗ 
    ██╔══██╗██╔════╝██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝    ██╔════╝ ██╔════╝████╗  ██║██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗
    ██████╔╝█████╗  ██████╔╝██║   ██║██████╔╝   ██║       ██║  ███╗█████╗  ██╔██╗ ██║█████╗  ██████╔╝███████║   ██║   ██║   ██║██████╔╝
    ██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║██╔══██╗   ██║       ██║   ██║██╔══╝  ██║╚██╗██║██╔══╝  ██╔══██╗██╔══██║   ██║   ██║   ██║██╔══██╗
    ██║  ██║███████╗██║     ╚██████╔╝██║  ██║   ██║       ╚██████╔╝███████╗██║ ╚████║███████╗██║  ██║██║  ██║   ██║   ╚██████╔╝██║  ██║
    ╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝        ╚═════╝ ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝
    `

	resolverArt = `
    ██████╗ ███████╗███████╗ ██████╗ ██╗    ██╗   ██╗███████╗██████╗ 
    ██╔══██╗██╔════╝██╔════╝██╔═══██╗██║    ██║   ██║██╔════╝██╔══██╗
    ██████╔╝█████╗  ███████╗██║   ██║██║    ██║   ██║█████╗  ██████╔╝
    ██╔══██╗██╔══╝  ╚════██║██║   ██║██║    ╚██╗ ██╔╝██╔══╝  ██╔══██╗
    ██║  ██║███████╗███████║╚██████╔╝███████╗╚████╔╝ ███████╗██║  ██║
    ╚═╝  ╚═╝╚══════╝╚══════╝ ╚═════╝ ╚══════╝ ╚═══╝  ╚══════╝╚═╝  ╚═╝
    `

	dependenciesArt = `
    ██████╗ ███████╗██████╗ ███████╗███╗   ██╗██████╗ ███████╗███╗   ██╗ ██████╗██╗███████╗███████╗
    ██╔══██╗██╔════╝██╔══██╗██╔════╝████╗  ██║██╔══██╗██╔════╝████╗  ██║██╔════╝██║██╔════╝██╔════╝
    ██║  ██║█████╗  ██████╔╝█████╗  ██╔██╗ ██║██║  ██║█████╗  ██╔██╗ ██║██║     ██║█████╗  ███████╗
    ██║  ██║██╔══╝  ██╔═══╝ ██╔══╝  ██║╚██╗██║██║  ██║██╔══╝  ██║╚██╗██║██║     ██║██╔══╝  ╚════██║
    ██████╔╝███████╗██║     ███████╗██║ ╚████║██████╔╝███████╗██║ ╚████║╚██████╗██║███████╗███████║
    ╚═════╝ ╚══════╝╚═╝     ╚══════╝╚═╝  ╚═══╝╚═════╝ ╚══════╝╚═╝  ╚═══╝ ╚═════╝╚═╝╚══════╝╚══════╝
    `

	mainBanner = `
    ██████╗  ██████╗ ██████╗ ██╗  ██╗███████╗██████╗ ███████╗████████╗██████╗ ██╗██╗  ██╗███████╗
    ██╔════╝ ██╔═══██╗██╔══██╗██║  ██║██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔══██╗██║██║ ██╔╝██╔════╝
    ██║  ███╗██║   ██║██████╔╝███████║█████╗  ██████╔╝███████╗   ██║   ██████╔╝██║█████╔╝ █████╗  
    ██║   ██║██║   ██║██╔═══╝ ██╔══██║██╔══╝  ██╔══██╗╚════██║   ██║   ██╔══██╗██║██╔═██╗ ██╔══╝  
    ╚██████╔╝╚██████╔╝██║     ██║  ██║███████╗██║  ██║███████║   ██║   ██║  ██║██║██║  ██╗███████╗
     ╚═════╝  ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝
    `
)

// displayBanner prints the GopherStrike ASCII art banner
func displayBanner() {
	fmt.Println(mainBanner)
}

// mainMenu displays and handles the main application menu
func mainMenu() {
	utils.ClearScreen()
	displayBanner() // this will have to get changed around
	fmt.Println("\nAvailable Tools:")
	fmt.Println("================")
	fmt.Println("1. Subdomain Scanner")
	fmt.Println("2. OSINT & Vulnerability Tool")
	fmt.Println("3. Web Application Security Scanner")
	fmt.Println("4. S3 Bucket Scanner")
	fmt.Println("5. Email Harvester")
	fmt.Println("6. Directory Bruteforcer")
	fmt.Println("7. Report Generator")
	fmt.Println("8. Host & Subdomain Resolver")
	fmt.Println("9. Check Dependencies")
	fmt.Println("10. Exit")

	// Get user input
	fmt.Printf("\n%s: ", "Enter your choice")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			// Handle EOF (Ctrl+D or non-interactive mode)
			fmt.Println("\nExiting GopherStrike. Goodbye!")
			os.Exit(0)
		}
		fmt.Printf("Error reading input: %v\n", err)
		fmt.Println("Exiting GopherStrike. Goodbye!")
		os.Exit(1)
	}
	
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("Invalid choice. Please try again.")
		utils.ClearScreen()
		mainMenu()
		return
	}
	
	choice, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid choice. Please enter a number between 1-10.")
		utils.ClearScreen()
		mainMenu()
		return
	}

	switch choice {
	case 1:
		utils.ClearScreen()
		fmt.Println(subdomainScannerArt)
		fmt.Println("\nRunning Subdomain Scanner...")
		// Run subdomain scanner
		if err := pkg.RunSubdomainScannerWithCheck(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 3:
		utils.ClearScreen()
		fmt.Println(osintArt)
		fmt.Println("\nRunning OSINT & Vulnerability Tool...")
		// Run OSINT tool
		if err := pkg.RunOSINTTool(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 4:
		utils.ClearScreen()
		fmt.Println(webVulnArt)
		fmt.Println("\nRunning Web Application Security Scanner...")
		// Call the web vulnerability scanner
		if err := pkg.RunWebVulnScanner(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 5:
		utils.ClearScreen()
		fmt.Println(s3ScannerArt)
		fmt.Println("\nRunning S3 Bucket Scanner...")
		// Call the S3 bucket scanner
		if err := tools.RunS3Scanner(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 6:
		utils.ClearScreen()
		fmt.Println(emailHarvesterArt)
		fmt.Println("\nRunning Email Harvester...")
		// Call the email harvester
		if err := tools.RunEmailHarvester(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 7:
		utils.ClearScreen()
		fmt.Println(dirBruteforceArt)
		fmt.Println("\nRunning Directory Bruteforcer...")
		// Call the directory bruteforcer
		if err := tools.RunDirBruteforcer(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 8:
		utils.ClearScreen()
		fmt.Println(reportGeneratorArt)
		fmt.Println("\nRunning Report Generator...")
		// Call the report generator
		if err := tools.RunReportingTools(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 9:
		utils.ClearScreen()
		fmt.Println(resolverArt)
		fmt.Println("\nRunning Host & Subdomain Resolver...")
		// Run host & subdomain resolver
		if err := pkg.RunHostResolver(); err != nil {
			fmt.Println("Error:", err)
		}
		utils.ClearScreen()
		mainMenu()
	case 10:
		utils.ClearScreen()
		fmt.Println(dependenciesArt)
		fmt.Println("\nChecking Dependencies...")
		// Run dependency check
		pkg.PrintDependencyStatus()
		utils.ClearScreen()
		mainMenu()
	case 11:
		utils.ClearScreen()
		fmt.Println(mainBanner)
		fmt.Println("\nExiting GopherStrike. Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid choice. Please try again.")
		utils.ClearScreen()
		mainMenu()
	}
}

// showHelp displays the help information
func showHelp() {
	fmt.Println(mainBanner)
	fmt.Println("\nGopherStrike - Advanced Security Reconnaissance Tool")
	fmt.Println("==================================================")
	fmt.Println("\nUsage:")
	fmt.Println("  ./GopherStrike              # Interactive mode")
	fmt.Println("  ./GopherStrike --help       # Show this help")
	fmt.Println("  ./GopherStrike -h           # Show this help")
	fmt.Println("\nAvailable Tools in Interactive Mode:")
	fmt.Println("=====================================")
	fmt.Println("1. Subdomain Scanner         - Discover subdomains of target domains")
	fmt.Println("2. OSINT & Vulnerability     - Open Source Intelligence gathering")
	fmt.Println("3. Web Application Scanner   - Web vulnerability assessment")
	fmt.Println("4. S3 Bucket Scanner         - AWS S3 bucket enumeration")
	fmt.Println("5. Email Harvester           - Email address collection")
	fmt.Println("6. Directory Bruteforcer     - Web directory discovery")
	fmt.Println("7. Report Generator          - Generate comprehensive reports")
	fmt.Println("8. Host & Subdomain Resolver - DNS resolution and validation")
	fmt.Println("9. Check Dependencies        - Verify required tools installation")
	fmt.Println("\nFor more information, visit: https://github.com/your-repo/GopherStrike")
}

// main is the entry point for the application
func main() {
	// Handle command line arguments
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "--help", "-h", "help":
			showHelp()
			return
		case "--version", "-v":
			fmt.Println(mainBanner)
			fmt.Println("\nGopherStrike v1.0.0")
			fmt.Println("Advanced Security Reconnaissance Tool")
			return
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[1])
			fmt.Println("Use --help for usage information")
			os.Exit(1)
		}
	}

	utils.ClearScreen() // clears the screen for the UI

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Handle Ctrl+C in a goroutine
	go func() {
		for range sigChan {
			if inTool {
				// If we're in a tool, just return to main menu
				fmt.Println("\n\nReturning to main menu...")
				inTool = false
				mainMenu()
			} else {
				// If we're in the main menu, exit the program
				utils.ClearScreen()
				fmt.Println(mainBanner)
				fmt.Println("\nExiting GopherStrike. Goodbye!")
				os.Exit(0)
			}
		}
	}()

	// Check for logs directory at startup with secure permissions
	if err := os.MkdirAll("logs", 0750); err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
	}

	// Create OSINT logs directory with secure permissions
	if err := os.MkdirAll("logs/osint", 0750); err != nil {
		fmt.Printf("Warning: Failed to create OSINT logs directory: %v\n", err)
	}

	// Create resolver logs directory with secure permissions
	if err := os.MkdirAll("logs/resolver", 0750); err != nil {
		fmt.Printf("Warning: Failed to create resolver logs directory: %v\n", err)
	}

	// Create webvuln logs directory with secure permissions
	if err := os.MkdirAll("logs/webvuln", 0750); err != nil {
		fmt.Printf("Warning: Failed to create webvuln logs directory: %v\n", err)
	}

	// Use the text-based menu directly
	mainMenu()
}
