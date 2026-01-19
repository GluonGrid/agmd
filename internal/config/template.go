package config

// AgentMdFilename is the main configuration file
const AgentMdFilename = "AGENTS.md"

// Legacy filenames that might exist in projects
var LegacyFilenames = []string{
	"CLAUDE.md",
	"AGENT.md",
	"agent.md",
	".cursorrules",
	".windsurfrules",
	"GEMINI.md",
	".continuerules",
}

// Tool configuration filenames (symlink targets)
const (
	ClaudeFilename    = "CLAUDE.md"
	CursorFilename    = ".cursorrules"
	WindsurfFilename  = ".windsurfrules"
	CopilotFilename   = ".github/copilot-instructions.md"
	AiderFilename     = ".aider.conf.yml"
)

// ToolConfig represents a tool's symlink configuration
type ToolConfig struct {
	Name     string
	Filename string
	NeedsDir bool // true if we need to create a directory (e.g., .github/)
}

// AvailableTools returns all supported tools
func AvailableTools() []ToolConfig {
	return []ToolConfig{
		{Name: "claude", Filename: ClaudeFilename, NeedsDir: false},
		{Name: "cursor", Filename: CursorFilename, NeedsDir: false},
		{Name: "windsurf", Filename: WindsurfFilename, NeedsDir: false},
		{Name: "copilot", Filename: CopilotFilename, NeedsDir: true},
		{Name: "aider", Filename: AiderFilename, NeedsDir: false},
	}
}

// GetToolByName returns a tool configuration by name
func GetToolByName(name string) *ToolConfig {
	for _, tool := range AvailableTools() {
		if tool.Name == name {
			return &tool
		}
	}
	return nil
}

// DefaultTemplate is the initial agent.md template
const DefaultTemplate = `# Agent Configuration

This file contains rules and guidelines for AI coding assistants working on this project.

## General Rules

- Follow the existing code style and conventions in this project
- Write clear, maintainable code with appropriate comments
- Run tests before committing changes
- Keep commits atomic and well-described

## Project Overview

[Add a brief description of what this project does]

## Code Style

[Add project-specific code style guidelines]

## Commands

[Add common commands used in this project]
- Build: [command]
- Test: [command]
- Run: [command]

## Architecture

[Add notes about the project architecture]

## Important Notes

[Add any project-specific quirks, gotchas, or important information]

---
*Generated with agmd*
`

// MinimalTemplate is a minimal starting template
const MinimalTemplate = `# Agent Configuration

## General Rules

- Follow existing code conventions
- Write clear, maintainable code
- Test before committing

## Commands

- Build: [your build command]
- Test: [your test command]

---
*Generated with agmd*
`
