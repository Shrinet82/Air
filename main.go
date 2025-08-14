package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func showHelp() {
	// Try to read the quick reference file
	quickRefPath := filepath.Join(filepath.Dir(os.Args[0]), "QUICKREF.md")
	content, err := os.ReadFile(quickRefPath)
	if err != nil {
		// Fallback to a simple help message
		fmt.Println("AIR Editor - AI-Integrated Text Editor")
		fmt.Println("Usage: ./air [filename]")
		fmt.Println("Press Ctrl+A to toggle AI chat, :q to quit")
		fmt.Println("For full documentation, see README.md")
		return
	}
	
	fmt.Println(string(content))
}

func main() {
	// Check for help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		showHelp()
		return
	}

	editor := NewEditor()
	var initialFile string
	if len(os.Args) > 1 {
		initialFile = os.Args[1]
	}

	buffer, err := NewBuffer(initialFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	editor.buffer = buffer

	if err := editor.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

