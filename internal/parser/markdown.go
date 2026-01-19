package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Section represents a markdown section with its content
type Section struct {
	Title   string
	Content []string
	Level   int // Heading level (1 for #, 2 for ##, etc.)
}

// AgentConfig represents the parsed agent.md file
type AgentConfig struct {
	Sections []Section
	RawContent string
}

// Parse reads and parses an agent.md file
func Parse(filename string) (*AgentConfig, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	config := &AgentConfig{
		RawContent: string(content),
		Sections:   []Section{},
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var currentSection *Section

	for scanner.Scan() {
		line := scanner.Text()

		// Check if it's a heading
		if strings.HasPrefix(line, "#") {
			// Save previous section if exists
			if currentSection != nil {
				config.Sections = append(config.Sections, *currentSection)
			}

			// Parse heading level
			level := 0
			for i := 0; i < len(line) && line[i] == '#'; i++ {
				level++
			}

			title := strings.TrimSpace(line[level:])
			currentSection = &Section{
				Title:   title,
				Level:   level,
				Content: []string{},
			}
		} else if currentSection != nil {
			// Add line to current section
			currentSection.Content = append(currentSection.Content, line)
		}
	}

	// Add the last section
	if currentSection != nil {
		config.Sections = append(config.Sections, *currentSection)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return config, nil
}

// GetSection returns a section by title (case-insensitive)
func (c *AgentConfig) GetSection(title string) *Section {
	lowerTitle := strings.ToLower(title)
	for i := range c.Sections {
		if strings.ToLower(c.Sections[i].Title) == lowerTitle {
			return &c.Sections[i]
		}
	}
	return nil
}

// AddRule adds a rule to a specific section
func (c *AgentConfig) AddRule(sectionTitle, rule string) error {
	section := c.GetSection(sectionTitle)
	if section == nil {
		return fmt.Errorf("section '%s' not found", sectionTitle)
	}

	// Add the rule as a bullet point
	section.Content = append(section.Content, fmt.Sprintf("- %s", rule))

	return nil
}

// AddSection adds a new section with a rule
func (c *AgentConfig) AddSection(title, rule string) {
	section := Section{
		Title:   title,
		Level:   2, // Default to ## level
		Content: []string{"", fmt.Sprintf("- %s", rule)},
	}
	c.Sections = append(c.Sections, section)
}

// ToString converts the config back to markdown string
func (c *AgentConfig) ToString() string {
	var builder strings.Builder

	for i, section := range c.Sections {
		// Write heading
		heading := strings.Repeat("#", section.Level) + " " + section.Title
		builder.WriteString(heading)
		builder.WriteString("\n")

		// Write content
		for _, line := range section.Content {
			builder.WriteString(line)
			builder.WriteString("\n")
		}

		// Add spacing between sections (except for last one)
		if i < len(c.Sections)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// Save writes the config to a file
func (c *AgentConfig) Save(filename string) error {
	content := c.ToString()
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// ListRules returns all rules from a specific section
func (c *AgentConfig) ListRules(sectionTitle string) ([]string, error) {
	section := c.GetSection(sectionTitle)
	if section == nil {
		return nil, fmt.Errorf("section '%s' not found", sectionTitle)
	}

	var rules []string
	for _, line := range section.Content {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			// Remove the bullet point marker
			rule := strings.TrimSpace(trimmed[1:])
			if rule != "" {
				rules = append(rules, rule)
			}
		}
	}

	return rules, nil
}

// ListAllRules returns all rules from all sections
func (c *AgentConfig) ListAllRules() map[string][]string {
	result := make(map[string][]string)

	for _, section := range c.Sections {
		rules, _ := c.ListRules(section.Title)
		if len(rules) > 0 {
			result[section.Title] = rules
		}
	}

	return result
}
