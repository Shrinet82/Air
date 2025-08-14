# AIR Quick Reference Guide

## Starting AIR
```
./air [filename]       Open or create a file
./air --help           Show this help text
```

## Modes
- **Normal Mode**: Navigation and commands (default)
- **Insert Mode**: Text editing (press `i` to enter)
- **Command Mode**: Execute commands (press `:` to enter)
- **Search Mode**: Search text (press `/` to enter)

## Global Shortcuts
- `Ctrl+A` - Toggle AI chat panel
- `Ctrl+S` - Save file
- `Ctrl+Z` - Undo
- `Ctrl+Y` - Redo
- `Ctrl+C` - Copy the most recent AI response

## Navigation (Normal Mode)
- `h` `j` `k` `l` - Move cursor (left, down, up, right)
- `w` / `b` - Move to next/previous word
- `gg` / `G` - Go to beginning/end of file

## Editing (Insert Mode)
- `Esc` - Return to Normal mode
- `Ctrl+V` - Paste clipboard content

## Commands
- `:w` - Save file
- `:q` - Quit (with check for unsaved changes)
- `:q!` - Force quit without saving
- `:wq` - Save and quit
- `:copy N` - Copy AI response #N
- `:[number]` - Go to line number

## AI Chat
1. Press `Ctrl+A` to toggle AI chat panel
2. Type your question and press Enter
3. Use `Ctrl+C` to copy the last response
4. Use `:copy N` to copy a specific response by number
5. Press `i` then `Ctrl+V` to paste in the editor

## Environment Setup
```bash
export GEMINI_API_KEY="your-api-key-here"
```
