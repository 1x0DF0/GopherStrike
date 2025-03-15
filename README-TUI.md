# GopherStrike Terminal UI

This document explains how to use the Terminal User Interface (TUI) features of GopherStrike.

## Overview

GopherStrike includes a rich terminal UI library that provides interactive components like:

- Panels with borders and titles
- Buttons
- Labels
- Select lists
- Progress bars
- Modal dialogs

These components can be used to create a more interactive and user-friendly interface for the GopherStrike toolkit.

## Current Status

The TUI library is currently in development. The basic components are implemented in the `pkg/termui` directory, but they are not yet integrated with the main application.

## Integration Plan

The plan is to integrate the TUI with the main application in the following steps:

1. Complete the implementation of the TUI components
2. Implement the menu system using the TUI components
3. Integrate the TUI menu with the main application
4. Add TUI interfaces for each tool

## Using the TUI Library

To use the TUI library in a tool, you would typically:

1. Import the TUI components:

```go
import "github.com/yourusername/GopherStrike/pkg/termui"
```

2. Create an application:

```go
app, err := termui.NewApplication()
if err != nil {
    // Handle error
}
```

3. Create and configure UI components:

```go
// Create a panel
panel := termui.NewPanel(0, 0, 80, 24)
panel.SetTitle("My Tool")

// Create a button
button := termui.NewButton(10, 10, 20, 3, "Click Me")
button.SetOnClick(func() {
    // Handle button click
})

// Add the button to the panel
panel.AddChild(button)
```

4. Add components to the application and run it:

```go
app.AddWidget(panel)
app.SetFocus(button)
app.Run()
```

## Fixing Linter Errors

If you encounter linter errors when working with the TUI library, here are some common issues and how to fix them:

1. Duplicate declarations of `Alignment` type and constants:
   - Keep these declarations only in `pkg/termui/component.go` and import them in other files

2. Missing fields in the `Application` struct:
   - Make sure the struct has the following fields: `mu`, `eventHandlers`, `styles`, and `focusManager`

3. Undefined color constants:
   - Check the tcell documentation for the available color constants
   - Common replacements: `tcell.ColorCyan` → `tcell.ColorTeal`

## Contributing to the TUI

Contributions to improve the TUI are welcome! Here are some areas that need work:

1. Implementing missing components
2. Improving the styling system
3. Adding documentation and examples
4. Integrating with the main application

Please follow the project's coding style and submit a pull request with your changes. 