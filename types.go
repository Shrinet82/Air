package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Mode defines the editor's current operational mode.
type Mode string

const (
	ModeNormal  Mode = "normal"
	ModeInsert  Mode = "insert"
	ModeCommand Mode = "command"
	ModeSearch  Mode = "search"
)

// Editor holds the entire state of the application.
type Editor struct {
	app          *tview.Application
	mainView     *tview.TextView
	statusBar    *tview.TextView
	commandInput *tview.InputField
	mainLayout   *tview.Flex
	chatPanel    *tview.Flex
	chatView     *tview.TextView
	chatInput    *tview.InputField

	buffer *Buffer
	mode   Mode
	cx, cy int // Cursor position in the buffer
	rx     int // Rendered cursor x position (for tabs)

	rowOffset int // Top row of the file being displayed
	colOffset int // Leftmost column of the file being displayed

	undoStack [][]string
	redoStack [][]string
	undoMutex sync.Mutex

	statusMsg   string
	chatVisible bool
	chatHistory []ChatMessage
	clipboard   string // For storing copied text

	searchQuery   string
	searchResults [][2]int // [line, char_pos]

	lastKey   string
	lastEvent *tcell.EventKey // For debugging
	debugKeys bool

	lineStarts []int
}

// ChatMessage represents a single message in the chat history.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Buffer encapsulates the editable text, file metadata, and dirty state.
type Buffer struct {
	Lines    []string
	FilePath string
	ReadOnly bool
	Dirty    bool
}

// NewBuffer creates a new buffer, loading from a file if it exists.
func NewBuffer(filePath string) (*Buffer, error) {
	b := &Buffer{
		FilePath: filePath,
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		b.Lines = []string{""} // Start with one empty line for new files
	} else {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		b.Lines = strings.Split(string(content), "\n")
		// Reading a file might add a trailing empty line if the file ends with a newline.
		// Let's remove it for consistency.
		if len(b.Lines) > 0 && b.Lines[len(b.Lines)-1] == "" {
			b.Lines = b.Lines[:len(b.Lines)-1]
		}
		if len(b.Lines) == 0 { // Ensure there's always at least one line
			b.Lines = []string{""}
		}
	}
	return b, nil
}

// BaseName returns a display-friendly name for the buffer.
func (b *Buffer) BaseName() string {
	if b.FilePath == "" {
		return "[No Name]"
	}
	return filepath.Base(b.FilePath)
}

// Save writes the buffer's content to its file path.
func (b *Buffer) Save() error {
	if b.FilePath == "" {
		return fmt.Errorf("no file path specified")
	}
	content := strings.Join(b.Lines, "\n")
	err := os.WriteFile(b.FilePath, []byte(content), 0644)
	if err != nil {
		return err
	}
	b.Dirty = false
	return nil
}
