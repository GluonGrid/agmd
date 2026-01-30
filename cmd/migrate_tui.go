package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginBottom(1)

	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170"))

	previewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(2)

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("46")).
			MarginTop(1)
)

// migrateModel is the bubbletea model for migrate interactive mode
type migrateModel struct {
	sections     []MigrateSection
	currentIndex int
	state        migrateState
	textInput    textinput.Model
	viewport     viewport.Model
	width        int
	height       int
	done         bool
	quit         bool

	// Stats
	rules      int
	workflows  int
	guidelines int
	skipped    int
}

type migrateState int

const (
	stateChooseType migrateState = iota
	stateEnterName
	stateEditing
)

// newMigrateModel creates a new migrate TUI model
func newMigrateModel(sections []MigrateSection) migrateModel {
	ti := textinput.New()
	ti.Placeholder = "item-name"
	ti.CharLimit = 64
	ti.Width = 40

	vp := viewport.New(80, 10)

	return migrateModel{
		sections:     sections,
		currentIndex: 0,
		state:        stateChooseType,
		textInput:    ti,
		viewport:     vp,
	}
}

func (m migrateModel) Init() tea.Cmd {
	return nil
}

func (m migrateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = min(10, msg.Height/3)
		return m, nil

	case editDoneMsg:
		if msg.err != nil {
			// Edit failed, go back to choose type
			m.state = stateChooseType
			return m, nil
		}
		// Read edited content
		if msg.tmpFile != "" {
			edited, err := os.ReadFile(msg.tmpFile)
			os.Remove(msg.tmpFile)
			if err == nil {
				// Strip the ## header if present
				content := string(edited)
				if idx := strings.Index(content, "\n"); idx > 0 && strings.HasPrefix(content, "## ") {
					content = strings.TrimSpace(content[idx:])
				}
				m.sections[m.currentIndex].Content = content
			}
		}
		m.state = stateChooseType
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateChooseType:
			return m.handleChooseType(msg)
		case stateEnterName:
			return m.handleEnterName(msg)
		case stateEditing:
			return m.handleEditing(msg)
		}
	}

	return m, nil
}

func (m migrateModel) handleChooseType(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	section := &m.sections[m.currentIndex]

	switch msg.String() {
	case "q", "ctrl+c":
		m.quit = true
		return m, tea.Quit

	case "r":
		section.ItemType = "rule"
		m.state = stateEnterName
		m.textInput.SetValue(slugify(section.Header))
		m.textInput.Focus()
		return m, textinput.Blink

	case "w":
		section.ItemType = "workflow"
		m.state = stateEnterName
		m.textInput.SetValue(slugify(section.Header))
		m.textInput.Focus()
		return m, textinput.Blink

	case "g":
		section.ItemType = "guideline"
		m.state = stateEnterName
		m.textInput.SetValue(slugify(section.Header))
		m.textInput.Focus()
		return m, textinput.Blink

	case "s":
		section.ItemType = ""
		m.skipped++
		return m.nextSection()

	case "e":
		// Edit - open in editor
		m.state = stateEditing
		return m, m.editSection()

	case "enter":
		// Default to rule
		section.ItemType = "rule"
		m.state = stateEnterName
		m.textInput.SetValue(slugify(section.Header))
		m.textInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

func (m migrateModel) handleEnterName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	section := &m.sections[m.currentIndex]

	switch msg.String() {
	case "enter":
		name := m.textInput.Value()
		if name == "" {
			name = slugify(section.Header)
		}
		section.ItemName = slugify(name)

		// Update stats
		switch section.ItemType {
		case "rule":
			m.rules++
		case "workflow":
			m.workflows++
		case "guideline":
			m.guidelines++
		}

		m.textInput.Blur()
		return m.nextSection()

	case "esc":
		m.state = stateChooseType
		m.textInput.Blur()
		return m, nil

	case "ctrl+c":
		m.quit = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m migrateModel) handleEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m migrateModel) editSection() tea.Cmd {
	section := m.sections[m.currentIndex]

	// Create temp file with section content
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("agmd-section-%s.md", slugify(section.Header)))

	fullContent := fmt.Sprintf("## %s\n\n%s", section.Header, section.Content)
	if err := os.WriteFile(tmpFile, []byte(fullContent), 0644); err != nil {
		return func() tea.Msg {
			return editDoneMsg{err: err}
		}
	}

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, tmpFile)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			os.Remove(tmpFile)
			return editDoneMsg{err: err, tmpFile: ""}
		}
		return editDoneMsg{tmpFile: tmpFile}
	})
}

type editDoneMsg struct {
	tmpFile string
	err     error
}

func (m migrateModel) nextSection() (tea.Model, tea.Cmd) {
	m.currentIndex++
	if m.currentIndex >= len(m.sections) {
		m.done = true
		return m, tea.Quit
	}
	m.state = stateChooseType
	return m, nil
}

func (m migrateModel) View() string {
	if m.done || m.quit {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render(fmt.Sprintf("Migrating Sections (%d/%d)", m.currentIndex+1, len(m.sections))))
	b.WriteString("\n")

	// Current section
	section := m.sections[m.currentIndex]
	b.WriteString(sectionHeaderStyle.Render(fmt.Sprintf("## %s", section.Header)))
	b.WriteString("\n\n")

	// Preview
	b.WriteString(previewStyle.Render("Preview:"))
	b.WriteString("\n")

	previewLines := strings.Split(section.Content, "\n")
	maxPreview := 5
	if len(previewLines) < maxPreview {
		maxPreview = len(previewLines)
	}
	for i := 0; i < maxPreview; i++ {
		line := previewLines[i]
		if len(line) > 70 {
			line = line[:67] + "..."
		}
		b.WriteString(previewStyle.Render("│ " + line))
		b.WriteString("\n")
	}
	if len(previewLines) > maxPreview {
		b.WriteString(previewStyle.Render(fmt.Sprintf("│ ... (%d more lines)", len(previewLines)-maxPreview)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// State-specific content
	switch m.state {
	case stateChooseType:
		b.WriteString("Choose type:\n")
		b.WriteString(helpStyle.Render("  [r] rule  [w] workflow  [g] guideline  [s] skip  [e] edit  [q] quit"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press Enter for rule (default)"))

	case stateEnterName:
		b.WriteString(fmt.Sprintf("Creating %s:\n", selectedStyle.Render(section.ItemType)))
		b.WriteString("Name: ")
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press Enter to confirm, Esc to go back"))

	case stateEditing:
		b.WriteString(helpStyle.Render("Opening editor..."))
	}

	// Stats bar
	b.WriteString("\n")
	b.WriteString(statusStyle.Render(fmt.Sprintf("Progress: %d rules, %d workflows, %d guidelines, %d skipped",
		m.rules, m.workflows, m.guidelines, m.skipped)))

	return b.String()
}

// runMigrateTUI runs the bubbletea-based interactive migrate
func runMigrateTUI(sections []MigrateSection) ([]MigrateSection, error) {
	m := newMigrateModel(sections)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	fm := finalModel.(migrateModel)
	if fm.quit {
		return fm.sections[:fm.currentIndex], nil
	}

	return fm.sections, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
