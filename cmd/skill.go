package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"agmd/pkg/markdown"
	"agmd/pkg/registry"
	"agmd/pkg/skills"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage Agent Skills",
	Long: `Install, link, and manage Agent Skills.

Skills are folders containing a SKILL.md file with instructions for AI agents.
They follow the open Agent Skills format (https://agentskills.io).

Quick start:
  agmd skill install owner/repo          # install from GitHub
  agmd skill link my-skill               # symlink into agent dirs
  agmd skill list                        # list installed skills

In directives.md, use :::skills to generate the <available_skills> XML block:
  :::skills                              # auto-detect .claude/skills or .agents/skills
  :::skills .claude/skills               # explicit path`,
}

var skillInstallLocal bool
var skillInstallName string

var skillInstallCmd = &cobra.Command{
	Use:   "install <owner/repo> [subdir]",
	Short: "Install a skill from GitHub",
	Long: `Install an Agent Skill from a GitHub repository.

The repository must contain a SKILL.md file at its root (or in the given subdir).
Uses git clone --depth 1 for fast, shallow cloning.

Examples:
  agmd skill install anthropics/skills/pdf          # install from subdir
  agmd skill install owner/repo                     # install from repo root
  agmd skill install owner/repo --name my-pdf       # install with custom name
  agmd skill install owner/repo --local             # install to project .agmd/`,
	Args: cobra.ExactArgs(1),
	RunE: runSkillInstall,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	RunE:  runSkillList,
}

var skillShowCmd = &cobra.Command{
	Use:               "show <name>",
	Short:             "Show SKILL.md content",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSkillName,
	RunE:              runSkillShow,
}

var skillLinkTarget string
var skillLinkCmd = &cobra.Command{
	Use:   "link <name>",
	Short: "Symlink a skill into agent directories",
	Long: `Create a symlink to an installed skill in the agent skills directory.

Target is auto-detected based on project structure:
  - .claude/ exists → .claude/skills/<name>
  - .agents/ exists → .agents/skills/<name>
  - neither → .agents/skills/<name> (created)

Override with --target claude|agents|both.

Examples:
  agmd skill link pdf                    # auto-detect target
  agmd skill link pdf --target claude    # link to .claude/skills/
  agmd skill link pdf --target both      # link to both`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSkillName,
	RunE:              runSkillLink,
}

var skillUnlinkCmd = &cobra.Command{
	Use:               "unlink <name>",
	Short:             "Remove skill symlinks from agent directories",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSkillName,
	RunE:              runSkillUnlink,
}

var skillDeleteCmd = &cobra.Command{
	Use:               "delete <name>",
	Short:             "Delete an installed skill from the registry",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSkillName,
	RunE:              runSkillDelete,
}

var skillUpdateCmd = &cobra.Command{
	Use:               "update <name>",
	Short:             "Re-fetch and update an installed skill",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSkillName,
	RunE:              runSkillUpdate,
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillInstallCmd)
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillShowCmd)
	skillCmd.AddCommand(skillLinkCmd)
	skillCmd.AddCommand(skillUnlinkCmd)
	skillCmd.AddCommand(skillDeleteCmd)
	skillCmd.AddCommand(skillUpdateCmd)

	skillInstallCmd.Flags().BoolVar(&skillInstallLocal, "local", false, "Install to project-local .agmd/ instead of ~/.agmd/")
	skillInstallCmd.Flags().StringVar(&skillInstallName, "name", "", "Override skill name (default: repo or subdir name)")
	skillLinkCmd.Flags().StringVar(&skillLinkTarget, "target", "", "Link target: claude, agents, or both (default: auto-detect)")
}

// --- helpers ---

func skillRegistryBase(reg *registry.Registry, local bool) (string, error) {
	if local {
		if reg.LocalPath == "" {
			return "", fmt.Errorf("no local registry found — run 'agmd init --local' first")
		}
		return reg.LocalPath, nil
	}
	return reg.BasePath, nil
}

// parseGitHubRef parses "owner/repo" or "owner/repo/subdir/path" into
// clone URL, subdir (may be empty), and default skill name.
func parseGitHubRef(ref string) (cloneURL, subdir, defaultName string) {
	parts := strings.SplitN(ref, "/", 3)
	owner := parts[0]
	repo := ""
	if len(parts) >= 2 {
		repo = parts[1]
	}
	if len(parts) == 3 {
		subdir = parts[2]
		// skill name = last component of subdir
		subParts := strings.Split(subdir, "/")
		defaultName = subParts[len(subParts)-1]
	} else {
		defaultName = repo
	}
	cloneURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	return
}

// --- command implementations ---

func runSkillInstall(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	ref := args[0]
	cloneURL, subdir, defaultName := parseGitHubRef(ref)

	skillName := skillInstallName
	if skillName == "" {
		skillName = defaultName
	}
	if skillName == "" {
		return fmt.Errorf("could not determine skill name — use --name to set it")
	}

	reg, err := registry.New()
	if err != nil {
		return err
	}
	if !reg.Exists() {
		return fmt.Errorf("registry not found — run 'agmd setup' first")
	}

	base, err := skillRegistryBase(reg, skillInstallLocal)
	if err != nil {
		return err
	}

	destDir := filepath.Join(base, "skill", skillName)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("skill '%s' already installed — run 'agmd skill update %s' to update", skillName, skillName)
	}

	// Check git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH — git is required to install skills")
	}

	fmt.Printf("%s Installing skill '%s' from %s\n", blue("->"), skillName, ref)

	// Clone into a temp dir then move the relevant subdir
	tmpDir, err := os.MkdirTemp("", "agmd-skill-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("%s Cloning %s\n", blue("->"), dim(cloneURL))
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "--quiet", cloneURL, tmpDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed")
	}

	// Determine source dir (repo root or subdir)
	srcDir := tmpDir
	if subdir != "" {
		srcDir = filepath.Join(tmpDir, subdir)
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			return fmt.Errorf("subdir '%s' not found in repository", subdir)
		}
	}

	// Validate SKILL.md exists before installing
	if _, err := skills.ParseSkillMD(srcDir, ""); err != nil {
		return fmt.Errorf("invalid skill: %w", err)
	}

	// Copy skill dir (can't rename across tmp filesystem)
	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}
	if err := copySkillDir(srcDir, destDir); err != nil {
		return fmt.Errorf("failed to copy skill: %w", err)
	}

	// Parse and display result
	skill, _ := skills.ParseSkillMD(destDir, "")
	fmt.Printf("%s Installed skill '%s'\n", green("ok"), skillName)
	if skill != nil {
		fmt.Printf("  %s\n", dim(skill.Description))
	}
	fmt.Printf("\nNext:\n")
	fmt.Printf("  agmd skill link %s     # symlink into agent dirs\n", skillName)

	return nil
}

func runSkillList(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return err
	}
	if !reg.Exists() {
		return fmt.Errorf("registry not found — run 'agmd setup' first")
	}

	var allSkills []*skills.Skill
	seen := make(map[string]bool)

	// Local first
	if reg.LocalPath != "" {
		localSkills, _ := skills.ListSkills(reg.LocalPath, "local")
		for _, s := range localSkills {
			seen[s.Name] = true
			allSkills = append(allSkills, s)
		}
	}
	// Global
	globalSkills, _ := skills.ListSkills(reg.BasePath, "global")
	for _, s := range globalSkills {
		if !seen[s.Name] {
			allSkills = append(allSkills, s)
		}
	}

	if len(allSkills) == 0 {
		fmt.Println("No skills installed.")
		fmt.Println("Install one with: agmd skill install owner/repo")
		return nil
	}

	fmt.Printf("Skills (%d installed)\n\n", len(allSkills))
	for _, s := range allSkills {
		src := dim("global")
		if s.Source == "local" {
			src = yellow("local")
		}
		fmt.Printf("  %s  %s  %s\n", green(s.Name), src, dim(s.Description))
	}
	return nil
}

func runSkillShow(cmd *cobra.Command, args []string) error {
	reg, err := registry.New()
	if err != nil {
		return err
	}

	skill, err := skills.FindSkill(args[0], reg.BasePath, reg.LocalPath)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(skill.SkillMDPath)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func runSkillLink(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return err
	}

	skillName := args[0]
	skill, err := skills.FindSkill(skillName, reg.BasePath, reg.LocalPath)
	if err != nil {
		return err
	}

	// Determine targets
	target := skillLinkTarget
	if target == "" {
		target = skills.DetectLinkTarget()
		if target == "" {
			target = "agents" // default if nothing detected
		}
	}

	var targets []string
	if target == "both" {
		targets = []string{"claude", "agents"}
	} else {
		targets = []string{target}
	}

	for _, t := range targets {
		linkPath := skills.SkillLinkPath(skillName, t)

		// Create parent dir
		if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Create absolute symlink target path
		absSkillPath, err := filepath.Abs(skill.Path)
		if err != nil {
			absSkillPath = skill.Path
		}

		// Check if already correctly linked
		if existing, err := os.Readlink(linkPath); err == nil && existing == absSkillPath {
			fmt.Printf("%s Already linked: %s\n", green("ok"), linkPath)
			continue
		}

		// Remove existing symlink or stale entry
		if _, err := os.Lstat(linkPath); err == nil {
			os.Remove(linkPath)
		}

		fmt.Printf("%s Linking %s → %s\n", blue("->"), linkPath, absSkillPath)
		if err := os.Symlink(absSkillPath, linkPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		fmt.Printf("%s Linked (%s)\n", green("ok"), t)
	}

	// Update :::skills block in directives.md
	updateDirectivesSkillsBlock(skillName, true, blue)

	return nil
}

func runSkillUnlink(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	skillName := args[0]
	removed := 0

	for _, t := range []string{"claude", "agents"} {
		linkPath := skills.SkillLinkPath(skillName, t)
		if _, err := os.Lstat(linkPath); err == nil {
			fmt.Printf("  %s Removing %s\n", dim("·"), linkPath)
			os.Remove(linkPath)
			removed++
		}
	}

	if removed == 0 {
		fmt.Printf("No symlinks found for skill '%s'\n", skillName)
	} else {
		fmt.Printf("%s Unlinked '%s'\n", green("ok"), skillName)
		// Update :::skills block in directives.md
		blue := color.New(color.FgBlue).SprintFunc()
		updateDirectivesSkillsBlock(skillName, false, blue)
	}
	return nil
}

func runSkillDelete(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return err
	}

	skillName := args[0]
	skill, err := skills.FindSkill(skillName, reg.BasePath, reg.LocalPath)
	if err != nil {
		return err
	}

	fmt.Printf("%s Delete skill '%s' at %s? (y/N): ", red("!"), skillName, skill.Path)
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(strings.TrimSpace(response)) != "y" {
		return fmt.Errorf("cancelled")
	}

	if err := os.RemoveAll(skill.Path); err != nil {
		return fmt.Errorf("failed to delete skill: %w", err)
	}

	fmt.Printf("%s Deleted skill '%s'\n", green("ok"), skillName)
	return nil
}

func runSkillUpdate(cmd *cobra.Command, args []string) error {
	blue := color.New(color.FgBlue).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return err
	}

	skillName := args[0]
	skill, err := skills.FindSkill(skillName, reg.BasePath, reg.LocalPath)
	if err != nil {
		return err
	}

	// Check if it's a git repo (has .git dir or was cloned)
	gitDir := filepath.Join(skill.Path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Can do git pull
		fmt.Printf("%s Updating skill '%s' via git pull\n", blue("->"), skillName)
		c := exec.Command("git", "pull", "--quiet")
		c.Dir = skill.Path
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	// No git — need to re-install. But we don't know the original URL.
	return fmt.Errorf("skill '%s' was not installed via git clone — delete and re-install manually:\n  agmd skill delete %s\n  agmd skill install owner/repo", skillName, skillName)
}

// copySkillDir recursively copies src to dst (used for skill installation)
func copySkillDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

// updateDirectivesSkillsBlock adds or removes a skill name from the :::skills block
// in directives.md. add=true adds, add=false removes. Prints info on failure.
func updateDirectivesSkillsBlock(name string, add bool, blue func(...interface{}) string) {
	const directivesFile = "directives.md"
	content, err := os.ReadFile(directivesFile)
	if err != nil {
		return // No directives.md in current dir — skip silently
	}
	var updated []byte
	if add {
		updated, err = markdown.AddToSkillsBlock(content, name)
	} else {
		updated, err = markdown.RemoveFromSkillsBlock(content, name)
	}
	if err != nil {
		fmt.Printf("%s Could not update %s: %v\n", blue("ℹ"), directivesFile, err)
		return
	}
	if err := os.WriteFile(directivesFile, updated, 0644); err != nil {
		fmt.Printf("%s Could not write %s: %v\n", blue("ℹ"), directivesFile, err)
		return
	}
	if add {
		fmt.Printf("%s Added '%s' to :::skills block in %s\n", blue("ℹ"), name, directivesFile)
	} else {
		fmt.Printf("%s Removed '%s' from :::skills block in %s\n", blue("ℹ"), name, directivesFile)
	}
}
