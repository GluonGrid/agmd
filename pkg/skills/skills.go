package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill represents an installed Agent Skill
type Skill struct {
	Name        string            // from SKILL.md frontmatter
	Description string            // from SKILL.md frontmatter
	License     string            // optional
	Compat      string            // optional compatibility field
	AllowedTools string           // optional space-delimited tools
	Metadata    map[string]string // optional arbitrary metadata
	Path        string            // absolute path to skill directory
	SkillMDPath string            // absolute path to SKILL.md
	Source      string            // "global" or "local"
}

// skillFrontmatter is used for YAML parsing only
type skillFrontmatter struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	License      string            `yaml:"license"`
	Compat       string            `yaml:"compatibility"`
	AllowedTools string            `yaml:"allowed-tools"`
	Metadata     map[string]string `yaml:"metadata"`
}

// ParseSkillMD parses a SKILL.md file and returns a Skill
func ParseSkillMD(skillDir, source string) (*Skill, error) {
	// Support both SKILL.md and skill.md
	var skillMDPath string
	for _, name := range []string{"SKILL.md", "skill.md"} {
		p := filepath.Join(skillDir, name)
		if _, err := os.Stat(p); err == nil {
			skillMDPath = p
			break
		}
	}
	if skillMDPath == "" {
		return nil, fmt.Errorf("no SKILL.md found in %s", skillDir)
	}

	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		return nil, err
	}

	fm, err := extractSkillFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("invalid SKILL.md frontmatter in %s: %w", skillDir, err)
	}
	if fm.Name == "" {
		return nil, fmt.Errorf("SKILL.md in %s missing required 'name' field", skillDir)
	}
	if fm.Description == "" {
		return nil, fmt.Errorf("SKILL.md in %s missing required 'description' field", skillDir)
	}

	return &Skill{
		Name:         fm.Name,
		Description:  fm.Description,
		License:      fm.License,
		Compat:       fm.Compat,
		AllowedTools: fm.AllowedTools,
		Metadata:     fm.Metadata,
		Path:         skillDir,
		SkillMDPath:  skillMDPath,
		Source:       source,
	}, nil
}

// extractSkillFrontmatter extracts YAML frontmatter from SKILL.md content
func extractSkillFrontmatter(data []byte) (*skillFrontmatter, error) {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("missing frontmatter (file must start with ---)")
	}

	lines := strings.Split(content, "\n")
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return nil, fmt.Errorf("frontmatter not closed (missing closing ---)")
	}

	fmYAML := strings.Join(lines[1:end], "\n")
	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(fmYAML), &fm); err != nil {
		return nil, err
	}
	return &fm, nil
}

// ListSkills scans a registry base path for installed skills
func ListSkills(basePath, source string) ([]*Skill, error) {
	skillDir := filepath.Join(basePath, "skill")
	entries, err := os.ReadDir(skillDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var skills []*Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(skillDir, entry.Name())
		skill, err := ParseSkillMD(dir, source)
		if err != nil {
			continue // skip invalid skills
		}
		skills = append(skills, skill)
	}
	return skills, nil
}

// FindSkill finds a skill by name, checking local first then global
func FindSkill(name, globalBase, localBase string) (*Skill, error) {
	if localBase != "" {
		dir := filepath.Join(localBase, "skill", name)
		if s, err := ParseSkillMD(dir, "local"); err == nil {
			return s, nil
		}
	}
	dir := filepath.Join(globalBase, "skill", name)
	if s, err := ParseSkillMD(dir, "global"); err == nil {
		return s, nil
	}
	return nil, fmt.Errorf("skill '%s' not found", name)
}

// DetectLinkTarget auto-detects where to link skills based on project structure.
// Returns "claude", "agents", or "" (unknown).
func DetectLinkTarget() string {
	if _, err := os.Stat(".claude"); err == nil {
		return "claude"
	}
	if _, err := os.Stat(".agents"); err == nil {
		return "agents"
	}
	return ""
}

// SkillLinkPath returns the symlink path for a skill given the target
func SkillLinkPath(skillName, target string) string {
	switch target {
	case "claude":
		return filepath.Join(".claude", "skills", skillName)
	case "agents":
		return filepath.Join(".agents", "skills", skillName)
	default:
		return filepath.Join(".agents", "skills", skillName)
	}
}

// LinkedSkills scans skill link dirs and returns skills with their SKILL.md paths
type LinkedSkill struct {
	Name        string
	Description string
	SkillMDPath string // absolute path
}

// FindLinkedSkills scans .claude/skills/ and .agents/skills/ for linked skills
func FindLinkedSkills(searchPath string) ([]*LinkedSkill, error) {
	// Resolve to absolute for the <location> field
	abs, err := filepath.Abs(searchPath)
	if err != nil {
		abs = searchPath
	}

	entries, err := os.ReadDir(abs)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var linked []*LinkedSkill
	for _, entry := range entries {
		if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		skillDir := filepath.Join(abs, entry.Name())

		// Resolve symlink if needed
		resolved, err := filepath.EvalSymlinks(skillDir)
		if err != nil {
			resolved = skillDir
		}

		skill, err := ParseSkillMD(resolved, "")
		if err != nil {
			continue
		}

		// Use the absolute SKILL.md path in the linked location (not the resolved target)
		linked = append(linked, &LinkedSkill{
			Name:        skill.Name,
			Description: skill.Description,
			SkillMDPath: filepath.Join(skillDir, "SKILL.md"),
		})
	}
	return linked, nil
}
