package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// --- Editor Initialization and Main Loop ---

func NewEditor() *Editor {
	e := &Editor{
		app:       tview.NewApplication(),
		mode:      ModeNormal,
		undoStack: make([][]string, 0),
		redoStack: make([][]string, 0),
	}

	// Initialize UI components
	e.mainView = tview.NewTextView().SetDynamicColors(true).SetWrap(false)
	e.statusBar = tview.NewTextView().SetDynamicColors(true)
	e.commandInput = tview.NewInputField().SetLabelColor(tcell.ColorWhite).SetFieldBackgroundColor(tcell.ColorBlack)
	e.chatView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetWrap(true)
	e.chatInput = tview.NewInputField().SetLabel("You: ").SetLabelColor(tcell.ColorYellow)

	// Assemble chat panel
	e.chatPanel = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.chatView, 0, 1, false).
		AddItem(e.chatInput, 3, 0, true)
	e.chatPanel.SetBorder(true).SetTitle("AI Chat")
	
	// Set up special input handling for chat view
	e.chatView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			e.copySelectedText()
			return nil
		}
		return event
	})
	
	// Enable mouse capture in the application
	e.app.EnableMouse(true)

	// Assemble main layout
	e.mainLayout = tview.NewFlex().SetDirection(tview.FlexColumn)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.mainLayout, 0, 1, true).
		AddItem(e.statusBar, 1, 0, false).
		AddItem(e.commandInput, 1, 0, false)

	// Set root and input captures
	e.app.SetRoot(layout, true).EnableMouse(true)
	e.app.SetInputCapture(e.globalInput)
	e.commandInput.SetDoneFunc(e.commandInputHandler)
	e.chatInput.SetDoneFunc(e.chatInputHandler)

	e.rebuildLayout()
	return e
}

func (e *Editor) Run() error {
	// Start a ticker to force redraws, which helps with async UI updates from the chat.
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			e.app.QueueUpdateDraw(func() {})
		}
	}()
	defer ticker.Stop()

	e.render() // Initial render
	return e.app.Run()
}

// --- UI Rendering and Layout ---

func (e *Editor) rebuildLayout() {
	e.mainLayout.Clear()
	if e.chatVisible {
		e.mainLayout.AddItem(e.mainView, 0, 1, true).AddItem(e.chatPanel, 40, 0, false)
	} else {
		e.mainLayout.AddItem(e.mainView, 0, 1, true)
	}
}

func (e *Editor) render() {
	if e.buffer == nil {
		return
	}
	e.recomputeLineStarts()
	e.scrollToCursor()

	var builder strings.Builder
	_, _, width, height := e.mainView.GetInnerRect()

	for y := 0; y < height; y++ {
		fileY := y + e.rowOffset
		if fileY >= len(e.buffer.Lines) {
			builder.WriteString("~")
		} else {
			line := e.buffer.Lines[fileY]
			if fileY == e.cy && e.mode == ModeInsert {
				// Special handling for the cursor line to draw the cursor manually
				e.calculateRx()
				
				// Safe substringing for horizontal scroll
				if e.colOffset > len(line) {
					line = ""
				} else {
					line = line[e.colOffset:]
				}
				
				// Truncate line to fit view width
				if len(line) > width {
					line = line[:width]
				}

				cursorX := e.rx - e.colOffset
				if cursorX < 0 {
					cursorX = 0
				}
				
				// Invert the character at the cursor position
				if cursorX < len(line) {
					line = fmt.Sprintf("%s[white:black]%c[-:-]%s", line[:cursorX], line[cursorX], line[cursorX+1:])
				} else if cursorX == len(line) {
					line = fmt.Sprintf("%s[white:black] [-:-]", line) // Show cursor at end of line
				}

			} else {
				// Regular line rendering
				line = highlightLineGo(line)
				if e.colOffset < len(line) {
					if e.colOffset+width > len(line) {
						line = line[e.colOffset:]
					} else {
						line = line[e.colOffset : e.colOffset+width]
					}
				} else {
					line = ""
				}
			}
			builder.WriteString(line)
		}
		builder.WriteString("\n")
	}

	e.mainView.SetText(builder.String())
	e.renderStatus()
}

func (e *Editor) renderStatus() {
	if e.buffer == nil {
		return
	}
	mode := fmt.Sprintf("[black:white] %s [-:-:-]", strings.ToUpper(string(e.mode)))
	file := e.buffer.BaseName()
	if e.buffer.Dirty {
		file += " [+]"
	}
	pos := fmt.Sprintf("%d:%d", e.cy+1, e.cx+1)

	status := fmt.Sprintf("%s %s - %s", mode, file, pos)
	if e.statusMsg != "" {
		status = e.statusMsg
	}

	debugInfo := ""
	if e.debugKeys && e.lastEvent != nil {
		keyStr := eventToKeyString(e.lastEvent)
		debugInfo = fmt.Sprintf(" | Last Key: %s", keyStr)
	}

	e.statusBar.SetText(status + debugInfo)
}

func (e *Editor) recomputeLineStarts() {
	e.lineStarts = e.lineStarts[:0]
	e.lineStarts = append(e.lineStarts, 0)
	totalLen := 0
	for _, line := range e.buffer.Lines {
		totalLen += len(line) + 1 // +1 for newline
		e.lineStarts = append(e.lineStarts, totalLen)
	}
}

func (e *Editor) scrollToCursor() {
	_, _, width, height := e.mainView.GetInnerRect()
	if height == 0 || width == 0 {
		return
	}

	// Vertical scrolling
	if e.cy < e.rowOffset {
		e.rowOffset = e.cy
	}
	if e.cy >= e.rowOffset+height {
		e.rowOffset = e.cy - height + 1
	}

	// Horizontal scrolling
	e.calculateRx()
	if e.rx < e.colOffset {
		e.colOffset = e.rx
	}
	if e.rx >= e.colOffset+width {
		e.colOffset = e.rx - width + 1
	}
}

func (e *Editor) calculateRx() {
	if e.cy < len(e.buffer.Lines) {
		e.rx = 0
		line := e.buffer.Lines[e.cy]
		for i := 0; i < e.cx; i++ {
			if i < len(line) && line[i] == '\t' {
				e.rx += 4 - (e.rx % 4) // Tab stop of 4
			} else {
				e.rx++
			}
		}
	}
}

// --- Input Handlers ---

func (e *Editor) globalInput(event *tcell.EventKey) *tcell.EventKey {
	Log(fmt.Sprintf("Global input received: Key=%s, Rune=%q", event.Name(), event.Rune()))
	e.lastEvent = event // For debugging
	e.statusMsg = ""    // Clear status message on new key press

	// If chat input has focus, let it handle the event.
	if e.app.GetFocus() == e.chatInput {
		return event
	}
	
	// If chat view has focus, handle special keys
	if e.app.GetFocus() == e.chatView {
		if event.Key() == tcell.KeyCtrlC {
			// Copy selected text from chat view
			e.copySelectedText()
			return nil
		}
		// Let other keys be handled normally
		return event
	}

	// Global commands that work in any mode
	switch event.Key() {
	case tcell.KeyCtrlA, tcell.KeyF2, tcell.KeyCtrlG:
		Log("Chat toggle key pressed")
		e.toggleChat()
		return nil
	case tcell.KeyCtrlS:
		e.Save()
		return nil
	case tcell.KeyCtrlZ:
		e.undo()
		return nil
	case tcell.KeyCtrlY:
		e.redo()
		return nil
	case tcell.KeyCtrlC:
		// Copy last AI response
		e.copyLastAIResponse()
		return nil
	}

	// Mode-specific handling
	switch e.mode {
	case ModeNormal:
		return e.normalModeInput(event)
	case ModeInsert:
		return e.insertModeInput(event)
	}

	e.render()
	return event
}

func (e *Editor) normalModeInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'i':
		e.mode = ModeInsert
	case ':':
		e.mode = ModeCommand
		e.commandInput.SetText(":")
		e.app.SetFocus(e.commandInput)
	case '/':
		e.mode = ModeSearch
		e.commandInput.SetText("/")
		e.app.SetFocus(e.commandInput)
	case 'h':
		e.moveCursor(-1)
	case 'l':
		e.moveCursor(1)
	case 'k':
		e.moveVertical(-1)
	case 'j':
		e.moveVertical(1)
	case 'w':
		e.moveWord(1)
	case 'b':
		e.moveWord(-1)
	case 'g':
		if e.lastKey == "g" {
			e.cy = 0
			e.cx = 0
			e.lastKey = "" // Reset
		} else {
			e.lastKey = "g"
		}
	case 'G':
		e.cy = len(e.buffer.Lines) - 1
		e.cx = 0
	case 'u':
		e.undo()
	case 'x':
		e.deleteChar()
	default:
		if e.lastKey == "g" && event.Rune() != 'g' {
			e.lastKey = ""
		}
	}
	e.render()
	return nil // Consume the event
}

func (e *Editor) insertModeInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		e.mode = ModeNormal
	case tcell.KeyEnter:
		e.insertNewline()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.backspace()
	case tcell.KeyRune:
		e.insertRune(event.Rune())
	case tcell.KeyLeft:
		e.moveCursor(-1)
	case tcell.KeyRight:
		e.moveCursor(1)
	case tcell.KeyUp:
		e.moveVertical(-1)
	case tcell.KeyDown:
		e.moveVertical(1)
	case tcell.KeyCtrlV:
		// Paste clipboard contents at cursor position
		if e.clipboard != "" {
			e.insertString(e.clipboard)
			e.statusMsg = "Text pasted from AI response"
		}
	}
	e.render()
	return nil // Consume the event
}

func (e *Editor) commandInputHandler(key tcell.Key) {
	if key == tcell.KeyEnter {
		cmdText := e.commandInput.GetText()
		e.commandInput.SetText("")
		e.app.SetFocus(e.mainView)

		if strings.HasPrefix(cmdText, ":") {
			e.exec(strings.TrimPrefix(cmdText, ":"))
			e.mode = ModeNormal
		} else if strings.HasPrefix(cmdText, "/") {
			e.search(strings.TrimPrefix(cmdText, "/"))
			e.mode = ModeNormal
		}
	} else if key == tcell.KeyEsc {
		e.commandInput.SetText("")
		e.mode = ModeNormal
		e.app.SetFocus(e.mainView)
	}
}

func (e *Editor) chatInputHandler(key tcell.Key) {
	if key == tcell.KeyEnter {
		userInput := e.chatInput.GetText()
		if strings.TrimSpace(userInput) != "" {
			e.sendChat(userInput)
		}
		e.chatInput.SetText("")
	} else if key == tcell.KeyEsc {
		e.app.SetFocus(e.mainView)
	}
}

// --- Commands and Actions ---

func (e *Editor) exec(cmd string) {
	parts := strings.Split(cmd, " ")
	switch parts[0] {
	case "q":
		if e.buffer.Dirty {
			e.statusMsg = "No write since last change (use q! to override)"
			return
		}
		e.app.Stop()
	case "q!":
		e.app.Stop()
	case "w":
		e.Save()
	case "wq":
		e.Save()
		if !e.buffer.Dirty {
			e.app.Stop()
		}
	case "chat":
		e.toggleChat()
	case "debugkeys":
		e.debugKeys = !e.debugKeys
		if e.debugKeys {
			e.statusMsg = "Key debugging enabled"
		} else {
			e.statusMsg = "Key debugging disabled"
		}
	case "copy":
		// Copy AI response by number: :copy 2
		if len(parts) < 2 {
			e.statusMsg = "Usage: copy <response-number>"
			return
		}
		num, err := strconv.Atoi(parts[1])
		if err != nil || num <= 0 {
			e.statusMsg = "Invalid response number"
			return
		}
		e.copyResponseByNumber(num)
	default:
		if lineNum, err := strconv.Atoi(parts[0]); err == nil {
			if lineNum > 0 && lineNum <= len(e.buffer.Lines) {
				e.cy = lineNum - 1
				e.cx = 0
			} else {
				e.statusMsg = "Invalid line number"
			}
		} else {
			e.statusMsg = fmt.Sprintf("Unknown command: %s", cmd)
		}
	}
}

func (e *Editor) search(query string) {
	e.searchQuery = query
	e.searchResults = [][2]int{}
	if query == "" {
		return
	}
	e.statusMsg = fmt.Sprintf("Search for '%s' (highlighted)", query)
}

func (e *Editor) toggleChat() {
	e.chatVisible = !e.chatVisible
	e.rebuildLayout()
	if e.chatVisible {
		e.statusMsg = "Chat panel opened. Click to select text, then Ctrl+C to copy selection."
		e.app.SetFocus(e.chatInput)
	} else {
		e.app.SetFocus(e.mainView)
	}
}

func (e *Editor) sendChat(userInput string) {
	msg := strings.TrimSpace(userInput)
	if msg == "" {
		return
	}

	e.chatHistory = append(e.chatHistory, ChatMessage{Role: "user", Content: msg})
	// Show a thinking message
	e.chatHistory = append(e.chatHistory, ChatMessage{Role: "model", Content: "..."})
	e.refreshChatView()

	go func(history []ChatMessage, prompt string) {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			e.app.QueueUpdateDraw(func() {
				// Replace thinking message with error
				e.chatHistory[len(e.chatHistory)-1] = ChatMessage{Role: "model", Content: "Error: GEMINI_API_KEY environment variable not set"}
				e.refreshChatView()
			})
			return
		}

		// Call the API
		ctx := context.Background()
		response, err := callGemini(ctx, apiKey, history[:len(history)-1], prompt)
		
		e.app.QueueUpdateDraw(func() {
			if err != nil {
				// Replace thinking message with error
				e.chatHistory[len(e.chatHistory)-1] = ChatMessage{Role: "model", Content: fmt.Sprintf("Error: %v", err)}
			} else {
				// Replace thinking message with response
				e.chatHistory[len(e.chatHistory)-1] = ChatMessage{Role: "model", Content: response}
			}
			e.refreshChatView()
		})
	}(append([]ChatMessage(nil), e.chatHistory...), msg)
}

func (e *Editor) refreshChatView() {
	if !e.chatVisible {
		return
	}
	var b strings.Builder
	
	// Number the AI responses for selection
	aiResponseCount := 0
	
	for _, m := range e.chatHistory {
		if m.Role == "user" {
			b.WriteString(fmt.Sprintf("[yellow]You:[-] %s\n", tview.Escape(m.Content)))
		} else {
			// Only number completed AI responses, not "..." placeholders
			if m.Content != "..." {
				aiResponseCount++
				b.WriteString(fmt.Sprintf("[green]AI #%d:[-] %s\n", aiResponseCount, tview.Escape(m.Content)))
			} else {
				b.WriteString(fmt.Sprintf("[green]AI:[-] %s\n", tview.Escape(m.Content)))
			}
		}
	}
	e.chatView.SetText(b.String())
	e.chatView.ScrollToEnd()
}

func (e *Editor) copyLastAIResponse() {
	// Find the last AI response
	for i := len(e.chatHistory) - 1; i >= 0; i-- {
		if e.chatHistory[i].Role == "model" && e.chatHistory[i].Content != "..." {
			e.clipboard = e.chatHistory[i].Content
			e.statusMsg = "Last AI response copied to clipboard. Press Ctrl+V in insert mode to paste."
			Log(fmt.Sprintf("Copied last AI response (index %d)", i))
			return
		}
	}
	Log("No AI response found to copy")
	e.statusMsg = "No AI response found to copy"
}

func (e *Editor) copySelectedText() {
	// Since direct text selection isn't available in tview,
	// we'll implement a simpler approach for now
	// Get the current chat view content
	text := e.chatView.GetText(false)
	if text != "" {
		// Store the text for later mouse selection support
		e.clipboard = text
		e.statusMsg = "Chat text copied. Press Ctrl+V in insert mode to paste. For specific responses, use :copy <number>"
	} else {
		e.statusMsg = "No text available to copy."
	}
}

func (e *Editor) copyResponseByNumber(num int) {
	Log(fmt.Sprintf("Attempting to copy AI response #%d", num))
	count := 0
	for _, m := range e.chatHistory {
		if m.Role == "model" && m.Content != "..." {
			count++
			if count == num {
				e.clipboard = m.Content
				e.statusMsg = fmt.Sprintf("AI response #%d copied. Press Ctrl+V in insert mode to paste.", num)
				Log(fmt.Sprintf("Successfully copied AI response #%d (length: %d chars)", num, len(m.Content)))
				return
			}
		}
	}
	Log(fmt.Sprintf("AI response #%d not found (total AI responses: %d)", num, count))
	e.statusMsg = fmt.Sprintf("AI response #%d not found", num)
}

func (e *Editor) insertString(s string) {
	// Insert each line of the string at the current cursor position
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return
	}
	
	e.pushUndo()
	
	// Insert first line at cursor position
	line := e.buffer.Lines[e.cy]
	if e.cx > len(line) {
		e.cx = len(line)
	}
	e.buffer.Lines[e.cy] = line[:e.cx] + lines[0] + line[e.cx:]
	e.cx += len(lines[0])
	
	// If we have more than one line, insert the rest as new lines
	if len(lines) > 1 {
		for _, l := range lines[1:] {
			// Create a new line
			e.cy++
			// Insert the rest of the lines as new lines
			e.buffer.Lines = append(e.buffer.Lines[:e.cy], append([]string{l}, e.buffer.Lines[e.cy:]...)...)
			e.cx = len(l)
		}
	}
	
	e.buffer.Dirty = true
}

// --- Editing Operations ---

func (e *Editor) insertRune(r rune) {
	e.pushUndo()
	line := e.buffer.Lines[e.cy]
	if e.cx > len(line) {
		e.cx = len(line)
	}
	e.buffer.Lines[e.cy] = line[:e.cx] + string(r) + line[e.cx:]
	e.cx++
	e.buffer.Dirty = true
}

func (e *Editor) insertNewline() {
	e.pushUndo()
	line := e.buffer.Lines[e.cy]
	if e.cx > len(line) {
		e.cx = len(line)
	}
	remaining := line[e.cx:]
	e.buffer.Lines[e.cy] = line[:e.cx]

	e.buffer.Lines = append(e.buffer.Lines[:e.cy+1], append([]string{remaining}, e.buffer.Lines[e.cy+1:]...)...)
	e.cy++
	e.cx = 0
	e.buffer.Dirty = true
}

func (e *Editor) backspace() {
	if e.cx == 0 && e.cy == 0 {
		return
	}
	e.pushUndo()

	if e.cx > 0 {
		line := e.buffer.Lines[e.cy]
		if e.cx > len(line) {
			e.cx = len(line)
		}
		e.buffer.Lines[e.cy] = line[:e.cx-1] + line[e.cx:]
		e.cx--
		e.buffer.Dirty = true
	} else {
		prevLine := e.buffer.Lines[e.cy-1]
		e.cx = len(prevLine)
		e.buffer.Lines[e.cy-1] = prevLine + e.buffer.Lines[e.cy]
		e.buffer.Lines = append(e.buffer.Lines[:e.cy], e.buffer.Lines[e.cy+1:]...)
		e.cy--
		e.buffer.Dirty = true
	}
}

func (e *Editor) deleteChar() {
	if e.cy >= len(e.buffer.Lines) {
		return
	}
	line := e.buffer.Lines[e.cy]
	if e.cx >= len(line) {
		return
	}
	e.pushUndo()
	e.buffer.Lines[e.cy] = line[:e.cx] + line[e.cx+1:]
	e.buffer.Dirty = true
}

// --- Cursor Movement ---

func (e *Editor) moveCursor(delta int) {
	if e.cy >= len(e.buffer.Lines) {
		return
	}
	newPos := e.cx + delta
	if newPos >= 0 && newPos <= len(e.buffer.Lines[e.cy]) {
		e.cx = newPos
	}
}

func (e *Editor) moveVertical(delta int) {
	newY := e.cy + delta
	if newY >= 0 && newY < len(e.buffer.Lines) {
		e.cy = newY
		if e.cx > len(e.buffer.Lines[e.cy]) {
			e.cx = len(e.buffer.Lines[e.cy])
		}
	}
}

func (e *Editor) moveWord(dir int) {
	// Implementation omitted for brevity
}

// --- Undo/Redo ---

func (e *Editor) pushUndo() {
	e.undoMutex.Lock()
	defer e.undoMutex.Unlock()

	// Create a snapshot
	snapshot := make([]string, len(e.buffer.Lines))
	copy(snapshot, e.buffer.Lines)

	e.undoStack = append(e.undoStack, snapshot)
	// If we have more than 100 undo states, trim the oldest one
	if len(e.undoStack) > 100 {
		e.undoStack = e.undoStack[1:]
	}
	// Any new action clears the redo stack
	e.redoStack = nil
}

func (e *Editor) undo() {
	e.undoMutex.Lock()
	defer e.undoMutex.Unlock()

	if len(e.undoStack) == 0 {
		return
	}

	// Pop from undo stack
	lastState := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]

	// Push current state to redo stack
	currentSnapshot := make([]string, len(e.buffer.Lines))
	copy(currentSnapshot, e.buffer.Lines)
	e.redoStack = append(e.redoStack, currentSnapshot)

	// Restore buffer
	e.buffer.Lines = lastState
	e.buffer.Dirty = true
	e.recomputeLineStarts()
	// TODO: Restore cursor position?
}

func (e *Editor) redo() {
	e.undoMutex.Lock()
	defer e.undoMutex.Unlock()

	if len(e.redoStack) == 0 {
		return
	}

	// Pop from redo stack
	nextState := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]

	// Push current state to undo stack
	currentSnapshot := make([]string, len(e.buffer.Lines))
	copy(currentSnapshot, e.buffer.Lines)
	e.undoStack = append(e.undoStack, currentSnapshot)

	// Restore buffer
	e.buffer.Lines = nextState
	e.buffer.Dirty = true
	e.recomputeLineStarts()
}

// --- File Operations ---

func (e *Editor) Save() {
	if err := e.buffer.Save(); err != nil {
		e.statusMsg = fmt.Sprintf("Error saving file: %v", err)
	} else {
		e.statusMsg = fmt.Sprintf("File '%s' saved", e.buffer.BaseName())
	}
}

func (e *Editor) openFile(path string) {
	b, err := NewBuffer(path)
	if err != nil {
		e.statusMsg = fmt.Sprintf("Error opening file: %v", err)
		return
	}
	e.buffer = b
	e.cy, e.cx = 0, 0
	e.rowOffset, e.colOffset = 0, 0
	e.undoStack = make([][]string, 0)
	e.redoStack = make([][]string, 0)
}

// --- Highlighting and Utility ---

func highlightLineGo(line string) string {
	keywords := map[string]bool{
		"func": true, "var": true, "const": true, "type": true, "struct": true,
		"interface": true, "package": true, "import": true, "return": true,
		"go": true, "defer": true, "for": true, "range": true, "if": true, "else": true,
		"switch": true, "case": true, "default": true, "map": true, "chan": true,
	}
	parts := strings.FieldsFunc(line, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	for _, part := range parts {
		if keywords[part] {
			line = strings.ReplaceAll(line, part, fmt.Sprintf("[blue]%s[-]", part))
		}
	}
	return line
}

func eventToKeyString(ev *tcell.EventKey) string {
	if ev.Key() == tcell.KeyRune {
		return string(ev.Rune())
	}
	return tcell.KeyNames[ev.Key()]
}
