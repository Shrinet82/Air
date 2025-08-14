# AIR Editor - AI-Integrated Text Editor

![Build Status](https://github.com/Shrinet82/Air/workflows/AIR%20Editor%20CI/badge.svg)

AIR is a terminal-based text editor with integrated AI capabilities, allowing you to seamlessly interact with a Gemini AI assistant while editing your files.

## Table of Contents
- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Editor Modes](#editor-modes)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [AI Chat Features](#ai-chat-features)
- [Command Reference](#command-reference)
- [Customization](#customization)
- [Troubleshooting](#troubleshooting)

## Installation

### Prerequisites
- Go 1.18 or higher
- Gemini API key (get one from [Google AI Studio](https://ai.google.dev/))

### Building from Source
```bash
git clone https://github.com/yourusername/air.git
cd air
go build -o air
```

### Setting up the API Key
Set your Gemini API key as an environment variable:

```bash
export GEMINI_API_KEY="your-api-key-here"
```

For permanent setup, add this line to your `~/.bashrc` or `~/.zshrc` file.

## Basic Usage

### Starting AIR
```bash
./air [filename]
```

If the file doesn't exist, a new one will be created. If no filename is provided, a new blank buffer will be opened.

## Editor Modes

AIR is a modal editor inspired by Vim, with the following modes:

### Normal Mode
The default mode for navigation and commands. Press `Esc` to return to Normal mode from other modes.

### Insert Mode
For typing and editing text. Press `i` in Normal mode to enter Insert mode.

### Command Mode
For executing editor commands. Press `:` in Normal mode to enter Command mode.

### Search Mode
For searching within the file. Press `/` in Normal mode to enter Search mode.

## Keyboard Shortcuts

### Global Shortcuts (Work in Any Mode)
- `Ctrl+A` - Toggle AI chat panel
- `Ctrl+S` - Save file
- `Ctrl+Z` - Undo
- `Ctrl+Y` - Redo
- `Ctrl+C` - Copy the most recent AI response

### Normal Mode
- `h` - Move cursor left
- `j` - Move cursor down
- `k` - Move cursor up
- `l` - Move cursor right
- `w` - Move to next word
- `b` - Move to previous word
- `gg` - Go to beginning of file
- `G` - Go to end of file
- `x` - Delete character at cursor
- `i` - Enter Insert mode
- `:` - Enter Command mode
- `/` - Enter Search mode

### Insert Mode
- `Esc` - Return to Normal mode
- `Ctrl+V` - Paste clipboard content (including AI responses)
- Arrow keys - Navigate the cursor

### Command Mode
Enter commands by typing `:` followed by the command and pressing Enter.

## AI Chat Features

AIR integrates with the Gemini API to provide AI assistance during your editing session.

### Using the AI Chat
1. Press `Ctrl+A` to toggle the AI chat panel
2. Type your question or request in the input field
3. Press `Enter` to send your message
4. The AI will respond in the chat panel

### Working with AI Responses

#### Copying AI Responses
- `Ctrl+C` - Copy the most recent AI response to clipboard
- `:copy N` - Copy a specific AI response by number (e.g., `:copy 2`)

#### Pasting AI Responses
1. Enter Insert mode by pressing `i`
2. Press `Ctrl+V` to paste the copied AI response at the cursor position

#### Response Numbering
Each AI response is numbered for easy reference:
```
You: What is Go?
AI #1: Go is a statically typed, compiled programming language...

You: How do I create a function in Go?
AI #2: To create a function in Go, use the `func` keyword...
```

## Command Reference

### File Operations
- `:w` - Save file
- `:q` - Quit (fails if there are unsaved changes)
- `:q!` - Force quit without saving
- `:wq` - Save and quit

### Editor Commands
- `:chat` - Toggle AI chat panel
- `:debugkeys` - Toggle key debugging mode
- `:[number]` - Go to specified line number

### AI Commands
- `:copy [number]` - Copy the specified AI response by number

## Customization

AIR currently supports basic customization through the source code. More configuration options will be added in future versions.

## Troubleshooting

### AI Chat Not Working
- Ensure `GEMINI_API_KEY` environment variable is set correctly
- Check your internet connection
- Verify the API key is valid and has necessary permissions

### Editor Display Issues
- Ensure your terminal supports the required features (colors, key combinations)
- Try resizing your terminal window for better layout

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
# Air
