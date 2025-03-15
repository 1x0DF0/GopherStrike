package pkg

import (
	"GopherStrike/pkg/termui"
	"fmt"
)

// The ASCII banner to display in the TUI
var banner = `
    ██████╗  ██████╗ ██████╗ ██╗  ██╗███████╗██████╗ ███████╗████████╗██████╗ ██╗██╗  ██╗███████╗
    ██╔════╝ ██╔═══██╗██╔══██╗██║  ██║██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔══██╗██║██║ ██╔╝██╔════╝
    ██║  ███╗██║   ██║██████╔╝███████║█████╗  ██████╔╝███████╗   ██║   ██████╔╝██║█████╔╝ █████╗  
    ██║   ██║██║   ██║██╔═══╝ ██╔══██║██╔══╝  ██╔══██╗╚════██║   ██║   ██╔══██╗██║██╔═██╗ ██╔══╝  
    ╚██████╔╝╚██████╔╝██║     ██║  ██║███████╗██║  ██║███████║   ██║   ██║  ██║██║██║  ██╗███████╗
     ╚═════╝  ╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝╚══════╝
    `

// Tool function type for defining tool callbacks
type ToolFunction func() error

// ToolInfo holds information about each tool
type ToolInfo struct {
	Name string
	Fn   ToolFunction
}

// Tools mapping
var tools = []ToolInfo{
	{"Port Scanner", RunNmapScannerWithPrivCheck},
	{"Subdomain Scanner", RunSubdomainScannerWithCheck},
	{"OSINT & Vulnerability Tool", RunOSINTTool},
	{"Web Application Security Scanner", RunWebVulnScanner},
	// Add more tools as they become available
}

// RunTerminalMenu runs the terminal-based menu using the TUI components
func RunTerminalMenu() error {
	// Create a new TUI application
	app, err := termui.NewApplication()
	if err != nil {
		return fmt.Errorf("failed to create TUI application: %w", err)
	}

	// Get screen dimensions
	width, height := app.GetScreen().Size()

	// Create main panel
	mainPanel := termui.NewPanel(0, 0, width, height)
	mainPanel.SetTitle("GopherStrike")
	mainPanel.SetCollapsible(false)

	// Add the ASCII banner at the top
	bannerLines := []string{}
	var line string
	for _, r := range banner {
		if r == '\n' {
			if line != "" {
				bannerLines = append(bannerLines, line)
				line = ""
			}
		} else {
			line += string(r)
		}
	}
	if line != "" {
		bannerLines = append(bannerLines, line)
	}

	// Add each line of the banner
	for i, line := range bannerLines {
		label := termui.NewLabel(2, 2+i, line)
		label.SetStyle(app.Styles().Get("title"))
		mainPanel.AddChild(label)
	}

	// Calculate starting y position for menu items (after the banner)
	menuStartY := len(bannerLines) + 4

	// Add menu options as buttons
	for i, tool := range tools {
		btn := termui.NewButton(10, menuStartY+i*3, 40, 3, tool.Name)
		btn.SetOnClick(func(index int) func() {
			return func() {
				app.Quit()
				// The actual tool will be called after the app quits
			}
		}(i))
		mainPanel.AddChild(btn)
	}

	// Add exit button
	exitBtn := termui.NewButton(10, menuStartY+len(tools)*3, 40, 3, "Exit")
	exitBtn.SetOnClick(func() {
		app.Quit()
	})
	mainPanel.AddChild(exitBtn)
	// Add the panel to the application

	// Run the application
	app.Run()

	return nil
}
