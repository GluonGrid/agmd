# Agent Modification Rules for agmd-Managed Files

**Purpose:** Guidelines for AI agents modifying `AGENTS.md` files managed by agmd

---

## File Structure Markers

### In Universal Shared Config (`~/.agmd/shared/base.md`)

Add these markers to protect structure:

```markdown
# Universal Agent Guardrails

**Version:** 1.0.0
**Managed by:** agmd

<!-- agmd:protected-structure-start -->
<!--
  PROTECTED SECTION - DO NOT MODIFY
  - Section titles (## 1. Name, ## 2. Name, etc.)
  - Section order (1-9 are core sections)
  - Marker comments like this one
  Changes here will break agmd's merge functionality
-->
<!-- agmd:protected-structure-end -->

---

## ⚠️ Agent Modification Guidelines

**If you are an AI agent modifying this file:**

### ✅ You MAY:
- Add examples and clarifications within sections
- Add new subsections (### or ####) under existing sections
- Fix typos and improve clarity
- Add new sections **after** section 9 (Multi-Agent Safety)
- Add operational learnings at the end

### ❌ You MUST NOT:
- Change section titles (## 1. through ## 9.)
- Remove or reorder core sections (1-9)
- Remove marker comments (<!-- agmd:... -->)
- Add content before "## 1. Intake & Context"

### 🔒 Reason:
agmd merges configs by section title. Changing titles breaks inheritance.

---

## 1. Intake & Context Understanding
...
```

### In Profile Files (`~/.agmd/profiles/*.md`)

```yaml
---
agmd:
  version: 1.0.0
  type: profile
  extends: universal
---

<!-- agmd:profile-metadata
  DO NOT modify the frontmatter above
  DO NOT change the "extends: universal" field
-->

# TypeScript Profile
...
```

### In Project Files (`AGENTS.md`)

```yaml
---
agmd:
  version: 1.0.0
  shared: ~/.agmd/shared/base.md
  profiles: [typescript, node-cli]
---

<!-- agmd:project-config
  Agent Instructions:
  - You MAY modify content below this marker
  - You MUST NOT modify the YAML frontmatter above
  - You MAY add new sections
  - You MAY suggest profile additions (add to 'profiles:' array)
  - You MUST preserve project-specific sections
-->

# Project: my-api

## Project Structure
...
```

---

## Modification Patterns

### ✅ SAFE: Adding Examples

**Before:**
```markdown
## 2. Code Quality Principles

### Refactor in place
- Never create duplicate files with "V2" suffixes
```

**After (SAFE):**
```markdown
## 2. Code Quality Principles

### Refactor in place
- Never create duplicate files with "V2" suffixes

**Examples of what NOT to do:**
- ❌ `utils-v2.ts`
- ❌ `helpers-new.ts`
- ❌ `api-backup.ts`

**What to do instead:**
- ✅ Edit `utils.ts` directly
- ✅ Use git history if you need old version
```

### ✅ SAFE: Adding Subsections

**Before:**
```markdown
## 3. Dependency Management

### Adding Dependencies
- Ask before adding
```

**After (SAFE):**
```markdown
## 3. Dependency Management

### Adding Dependencies
- Ask before adding

### Evaluating Dependencies
- Check maintenance status
- Review license compatibility
- Assess bundle size impact

### Avoiding Dependency Bloat
- Prefer native APIs when available
- Consider utility functions over libraries
```

### ❌ UNSAFE: Changing Section Titles

**Before:**
```markdown
## 2. Code Quality Principles
```

**After (BREAKS AGMD):**
```markdown
## 2. Code Quality & Best Practices  ← BREAKS MERGING!
```

**Why it breaks:**
agmd merges by looking for "## 2. Code Quality Principles". If you change the title, agmd can't find the section and will duplicate it.

### ❌ UNSAFE: Reordering Sections

**Before:**
```markdown
## 1. Intake & Context
## 2. Code Quality
## 3. Dependencies
```

**After (BREAKS AGMD):**
```markdown
## 1. Code Quality          ← WRONG ORDER!
## 2. Intake & Context      ← BREAKS MERGING!
## 3. Dependencies
```

**Why it breaks:**
agmd expects sections in a specific order for proper merging.

### ✅ SAFE: Adding Project-Specific Sections

**In project's AGENTS.md:**

```markdown
<!-- After the frontmatter -->

# Project: my-api

## Project Structure
[Existing content]

## Build Commands
[Existing content]

<!-- SAFE: Add new sections here -->

## Operational Learnings

### Authentication Bug (2026-01-15)
**Problem:** JWT tokens expiring too quickly
**Cause:** Clock skew between servers
**Fix:** Added 30s grace period
**Lesson:** Always account for clock drift

## Special Workflows

### Database Migrations
1. Create migration: `pnpm db:migrate:create`
2. Test locally: `pnpm db:migrate:up`
3. Review SQL in PR
4. Deploy to staging first
```

---

## Project File (AGENTS.md) Rules

### Frontmatter Protection

**The YAML frontmatter is SACRED:**

```yaml
---
agmd:
  version: 1.0.0              # DO NOT CHANGE
  shared: ~/.agmd/shared/base.md    # DO NOT CHANGE
  profiles: [typescript]      # You MAY add profiles here
  overrides:                  # You MAY add overrides here
    section.key: value
---
```

### What Agents CAN Change

1. **Add profiles:**
   ```yaml
   profiles:
     - typescript
     - node-cli     # ← Agent can add this
   ```

2. **Add overrides:**
   ```yaml
   overrides:
     code-quality.file-size-limit: 800  # ← Agent can add this
   ```

3. **Modify ALL content below frontmatter:**
   - Project structure
   - Build commands
   - Special rules
   - Deployment notes
   - Operational learnings

### What Agents CANNOT Change

1. **agmd version field** - Critical for compatibility
2. **shared path** - Changing breaks inheritance
3. **Frontmatter structure** - Must remain valid YAML

---

## Override Mechanism (Safe Way to Change Rules)

Instead of modifying universal/profile files, use overrides:

### Example: Project Needs Larger Files

**❌ WRONG (modifying universal file):**
```markdown
# In ~/.agmd/shared/base.md
## 2. Code Quality
- Keep files under 800 LOC  ← Don't edit universal file!
```

**✅ RIGHT (using override):**
```yaml
# In project's AGENTS.md frontmatter
---
agmd:
  overrides:
    code-quality.file-size-limit: 800
---
```

### How Overrides Work

**Universal says:**
```markdown
## 2. Code Quality
- Keep files manageable size (~500-700 LOC guideline)
```

**Project overrides:**
```yaml
overrides:
  code-quality.file-size-limit: 800
```

**Agent sees (after merge):**
```markdown
## 2. Code Quality
- Keep files under 800 LOC
```

---

## Validation

### After Modification, Run:

```bash
agmd validate
```

**Good output:**
```
✓ YAML frontmatter is valid
✓ Shared config found
✓ All profiles found
✓ Section structure preserved
✓ No duplicate sections

Config is valid
```

**Bad output (agent broke something):**
```
✗ YAML frontmatter is invalid (syntax error on line 3)
✗ Section "## 2. Code Quality Principles" not found (was it renamed?)
✗ Duplicate section "## 3. Dependencies" found

Config is INVALID - please fix errors
```

---

## Summary for AI Agents

### When modifying `AGENTS.md` (project file):

1. **✅ DO:**
   - Add content below frontmatter
   - Add new sections (structure, commands, learnings)
   - Add profiles to frontmatter array
   - Add overrides to frontmatter
   - Improve examples and clarity

2. **❌ DON'T:**
   - Modify `agmd.version` or `agmd.shared`
   - Break YAML syntax in frontmatter
   - Remove existing sections
   - Add content before frontmatter

3. **✅ VALIDATE:**
   - Run `agmd validate` after changes
   - Check that file still merges correctly

### When modifying universal/profile files:

1. **✅ DO:**
   - Add examples within sections
   - Add subsections (### or ####)
   - Clarify existing content
   - Fix typos

2. **❌ DON'T:**
   - Change section titles (## 1. through ## 9.)
   - Reorder sections
   - Remove core sections
   - Remove marker comments

3. **🎯 BETTER APPROACH:**
   - Use project-specific overrides instead
   - Suggest adding a profile for common patterns
   - Add operational learnings to project file

---

## Marker Comment Reference

### Universal Config Markers

```markdown
<!-- agmd:protected-structure-start -->
Core sections that must not be renamed/reordered
<!-- agmd:protected-structure-end -->

<!-- agmd:extensible-section -->
This section can have subsections added
<!-- agmd:extensible-section-end -->

<!-- agmd:end-core-sections -->
New sections may be added after this marker
```

### Profile Markers

```yaml
---
agmd:
  type: profile
  extends: universal
---

<!-- agmd:profile-metadata -->
DO NOT modify frontmatter above
```

### Project Markers

```yaml
---
agmd:
  version: 1.0.0
  shared: ~/.agmd/shared/base.md
  profiles: [typescript]
---

<!-- agmd:project-config -->
You MAY modify everything below this line
```

---

## Example: Agent Adding Feature-Specific Guidance

**Scenario:** Agent is implementing OAuth and wants to add security guidance

**❌ WRONG:**
Modify universal file's security section:
```markdown
# In ~/.agmd/shared/base.md
## 7. Security & Secrets
- Never commit secrets
- OAuth tokens must use PKCE flow  ← TOO SPECIFIC!
```

**✅ RIGHT:**
Add to project file:
```markdown
# In project's AGENTS.md

## Security Guidelines

### OAuth Implementation
- Use PKCE flow for public clients
- Refresh tokens must be stored in httpOnly cookies
- Access tokens expire in 15 minutes
- Never log tokens or sensitive data

### References
- OAuth spec: RFC 6749
- PKCE spec: RFC 7636
```

---

## Testing Agent Modifications

### Before committing changes:

1. **Validate:**
   ```bash
   agmd validate
   ```

2. **Check merge:**
   ```bash
   agmd show --merged
   ```
   - Ensure no duplicate sections
   - Check that content appears correctly

3. **Verify YAML:**
   ```bash
   agmd show --json > /dev/null
   ```
   - If this fails, YAML is broken

---

**Last Updated:** 2026-01-19
**Version:** 1.0.0
**For:** agmd v1.0.0+
