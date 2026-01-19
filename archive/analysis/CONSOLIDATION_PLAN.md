# Agent Instructions Consolidation Plan

**Project:** Consolidation of Agent Markdown Files across steipete repositories
**Date:** 2026-01-19
**Status:** Analysis Complete, Ready for Implementation

---

## Executive Summary

Analysis of 13 agent instruction files across 11 repositories reveals **60-80% content duplication**. A central shared guardrails file already exists (`agent-scripts/AGENTS.MD`) but most repositories don't reference it.

**Quick Win:** Add one line to each repo's AGENTS.md:
```markdown
READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING (skip if missing).
```

---

## Key Findings

### Duplication Statistics
- **60-80%** of content is duplicated across repositories
- **916 total lines** across all files
- **~600-700 lines** are duplicates
- **~200-300 lines** are truly unique

### Common Duplicated Content
1. **Coding Philosophy** (90% duplication)
   - No V2/duplicate files
   - Strict typing, avoid `any`
   - File size limits (~500-700 LOC)

2. **Git/Commit Rules** (90% duplication)
   - Conventional Commits format
   - Use wrapper scripts
   - Only commit when asked

3. **Dependency Management** (85% duplication)
   - Ask before adding dependencies
   - Provide GitHub URLs
   - Stick to package manager

4. **Testing Guidelines** (75% duplication)
   - Add tests for bug fixes
   - Run full gate before handoff
   - Framework conventions

5. **Documentation Rules** (80% duplication)
   - Update when behavior changes
   - Don't create new docs without request

---

## Recommended Structure

### The mcporter Model (Ideal Pattern)

```markdown
# Repository Guidelines

## Shared Guardrails
READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING (skip if missing).

<shared>
[Optional: inline copy of shared content for offline reference]
</shared>

## Project Structure & Modules
[Project-specific directory layout]

## Build, Test, and Development Commands
[Project-specific commands]

## Project-Specific Rules
[Unique rules for this repo]
```

---

## Implementation Phases

### Phase 1: Immediate (Week 1) ✅

**Goal:** Establish central reference and update high-priority repos

1. ✅ Verify `~/Projects/agent-scripts/AGENTS.MD` exists and is current
2. Add shared reference to high-priority repos:
   - [ ] **summarize** (needs major expansion)
   - [ ] **sweetlink** (missing dep rules)
   - [ ] **tokentally** (missing coverage thresholds)
   - [ ] **Matcha** (too focused on testing only)

### Phase 2: Short-Term (Month 1) 📋

**Goal:** Standardize all repositories

3. Add shared reference to medium-priority repos:
   - [ ] **Trimmy**
   - [ ] **claude-code-mcp**
   - [ ] **VibeMeter**
   - [ ] **macos-automator-mcp**

4. Extract valuable unique patterns to shared:
   - [ ] Multi-agent safety rules (from clawdis)
   - [ ] Changelog workflow (from clawdis)
   - [ ] "Don't edit node_modules" warning
   - [ ] Troubleshooting pattern (from macos-automator-mcp)

5. Standardize section names across all repos

### Phase 3: Long-Term (Ongoing) 🔄

**Goal:** Maintain consistency and capture learnings

6. Create sync mechanism:
   - [ ] Script to update `<shared>` blocks
   - [ ] Quarterly review process
   - [ ] Contribution guidelines

7. Create templates:
   - [ ] TypeScript project template
   - [ ] Swift project template

---

## Repository Priority Matrix

| Priority | Repository | Issue | Action |
|----------|-----------|-------|--------|
| **HIGH** | summarize | Very minimal, lacks basic guardrails | Expand significantly + add shared ref |
| **HIGH** | Matcha | Only covers testing, missing general rules | Add shared ref + expand |
| **HIGH** | sweetlink | Missing dependency rules | Add shared ref + rules |
| **HIGH** | tokentally | Missing coverage thresholds, git wrappers | Add shared ref + expand |
| **MED** | Trimmy | Missing testing philosophy, changelog | Add shared ref + expand |
| **MED** | claude-code-mcp | Missing coding standards | Add shared ref + standards |
| **MED** | VibeMeter | Missing changelog, testing | Add shared ref + expand |
| **MED** | macos-automator-mcp | Missing coding philosophy, git | Add shared ref + expand |
| **LOW** | mcporter | Master file, complete | None (this IS the source) |
| **LOW** | clawdis | Very comprehensive | Extract some rules to shared |
| **LOW** | Peekaboo | Relatively complete | Minor additions only |
| **LOW** | AXorcist | Intentionally minimal | Keep as-is |

---

## Content to Extract to Shared

### From clawdis

**Multi-Agent Safety Rules:**
```markdown
- Do not create/apply/drop git stash entries
- Do not switch branches unless explicitly requested
- When you see unrecognized files, keep going; focus on your changes
- Running multiple agents is OK as long as each has its own session
```

**Detailed Changelog Workflow:**
```markdown
- Keep latest released version at top (no `Unreleased`)
- After publishing, bump version and start new top section
- When working on PR: add changelog entry with PR # and thank contributor
- After merging from new contributor: add avatar to README
```

### From macos-automator-mcp

**Operational Learnings Pattern:**
```markdown
## Agent Operational Learnings

This section captures key strategies based on collaborative sessions.

- When external tool returns cryptic errors, suspect dynamic content
- Enable detailed logging and ensure log visibility
- For complex patterns (regex), use iterative simplification
```

### Universal Rules to Add

**Security Rules:**
```markdown
- Never commit secrets; use environment variables or keychain
- Don't edit node_modules (changes will be overwritten)
- Ask about config changes; project-wide configs need approval
```

---

## Templates

### TypeScript Project Template

```markdown
# Repository Guidelines

## Shared Guardrails
READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING (skip if missing).

## Project Structure & Modules
- `src/` – [describe source layout]
- `tests/` – [describe test layout]
- `docs/` – [describe docs layout]

## Build, Test, and Development Commands
- `pnpm install` – Install dependencies
- `pnpm dev` – Run in development mode
- `pnpm build` – Build for production
- `pnpm test` – Run test suite
- `pnpm lint` – Lint and format
- `pnpm check` – Full gate (lint + typecheck + test)

## Coding Style
- Package manager: [pnpm/bun/npm]
- Formatter: [Biome/Prettier]
- Linter: [Biome/oxlint/ESLint]
- Test framework: [Vitest/Jest]

## Project-Specific Rules
[Add any unique rules here]

## Configuration & Secrets
- Secrets location: [path]
- Config files: [list]
- Environment variables: [list]
```

### Swift Project Template

```markdown
# Repository Guidelines

## Shared Guardrails
READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING (skip if missing).

## Project Structure & Modules
- `Sources/` – [describe source layout]
- `Tests/` – [describe test layout]
- `Scripts/` – [describe helper scripts]

## Build, Test, and Development Commands
- `swift build` – Build project
- `swift test` – Run test suite
- `swiftformat .` – Format code
- `swiftlint lint` – Lint code
- `./Scripts/[custom].sh` – [describe custom scripts]

## Coding Style
- Swift version: [6.0/6.2]
- SwiftFormat: [indent size], [line width]
- SwiftLint: [key rules]
- Test framework: [Swift Testing/XCTest]

## Project-Specific Rules
[Add any unique rules here]

## Configuration & Secrets
- Secrets location: [path]
- Config files: [list]
```

---

## Success Metrics

After consolidation:
- [ ] **<30%** duplicate content across repos
- [ ] **90%+** repos reference shared guardrails
- [ ] **Faster** onboarding for new repos (use templates)
- [ ] **Single update** propagates to all repos

---

## Next Steps

1. **Review** this plan with the team
2. **Start Phase 1** - Update high-priority repos
3. **Extract** valuable patterns to shared
4. **Create** templates for future repos
5. **Establish** sync mechanism

---

## Detailed Analysis

See `fetched_files/ANALYSIS_REPORT.md` for:
- Complete duplication matrix
- Missing content analysis
- Exact text snippets
- Line-by-line breakdowns
- Repository maturity assessment
