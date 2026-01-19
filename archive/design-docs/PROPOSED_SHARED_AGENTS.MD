# AGENTS.MD

**Shared Agent Guardrails**

This file contains consolidated rules distilled from agent instruction files across multiple repositories. Local repositories should reference this file at the start of their own AGENTS.md, then add project-specific details.

**Usage in local repos:**
```markdown
READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING (skip if missing).
```

---

## General Principles

### Intake & Scoping

- **Read context at session start:**
  - Open local agent instructions (AGENTS.md, CLAUDE.md)
  - Review any `docs:list` summaries
  - Check referenced tmux panes, CI logs, or failing command transcripts
  - Re-run doc helpers when docs may have changed

- **Understand before acting:**
  - Study existing code patterns before introducing new ones
  - Review recent commit history to understand project evolution
  - Ask questions when requirements are unclear

---

## Code Quality & Style

### Core Philosophy

- **Refactor in place**
  - Never create duplicate files with suffixes like "V2", "New", "Fixed", or "Backup"
  - Update the canonical file and remove obsolete paths entirely
  - If a file needs major restructuring, do it incrementally in the same file

- **Strict typing**
  - Avoid `any` in TypeScript
  - Avoid `Any` in Swift
  - Prefer utility types and concrete type definitions
  - Use strict compiler settings

- **File size management**
  - Keep files at manageable size (~500-700 LOC guideline, not a hard limit)
  - When a file grows unwieldy, extract helpers or new modules
  - Don't let files bloat—refactor proactively

- **Match established style**
  - Study the repo's existing code before adding new patterns
  - Follow existing conventions for naming, organization, and structure
  - When in doubt, ask rather than introducing inconsistency

---

## Dependencies & Tooling

### Dependency Management

- **Ask before adding dependencies**
  - Never add dependencies without explicit user approval
  - Provide rationale: why is this needed? What problem does it solve?
  - When discussing dependencies, always provide a GitHub URL

- **Respect package manager**
  - Stick to the declared package manager (pnpm, bun, npm, swift, go, etc.)
  - Never swap package managers without explicit approval
  - Don't mix package managers in the same project

- **Exact versions for patched dependencies**
  - Any dependency with patches (`pnpm.patchedDependencies`, overrides, vendored changes) must use exact version
  - No `^` or `~` semver ranges on patched deps
  - Patching dependencies requires explicit approval

### Command Wrappers

- **Use provided wrappers**
  - Prefer repo-specific wrappers: `./runner`, `scripts/committer`, `pnpm mcp:*`
  - Don't bypass wrappers with raw commands unless necessary
  - Wrappers often contain important project-specific logic

- **Don't edit tooling files without approval**
  - Never edit `node_modules` (changes will be overwritten on next install)
  - Don't modify global installs (Homebrew, npm global, git installs)
  - Ask before changing build tooling or project-wide configuration

---

## Testing

### Testing Philosophy

- **Treat bug fixes as test opportunities**
  - Every bug fix should add or extend automated tests
  - Tests prove the behavior and prevent regression
  - If you can't write a test, document why

- **Run full gate before handoff**
  - Always run complete test suite before completing work
  - Don't hand off failing tests
  - Fix or explain any test failures

### Framework Conventions

**TypeScript Projects:**
- Test framework: Vitest (preferred) or Jest
- Naming: `*.test.ts` for unit tests, `*.e2e.test.ts` for e2e
- Location: `tests/` or co-located with source
- Coverage: Aim for 70-80% (check repo-specific thresholds)

**Swift Projects:**
- Test framework: Swift Testing (preferred for new code) or XCTest
- Naming: `*Tests.swift`
- Location: `Tests/` directory
- Annotations: Use `@Test`, `@Suite`, `#expect()` for Swift Testing

### Coverage Guidelines

- Prioritize testing critical paths and edge cases
- Don't chase 100% coverage for the sake of metrics
- Test public APIs and user-facing behavior
- Mock external dependencies (network, filesystem, etc.)

---

## Git, Commits & Releases

### Git Workflow

- **Use wrapper scripts when available**
  - Prefer `scripts/committer` or `./git` over raw git commands
  - Wrappers often handle staging scoping and other project-specific logic
  - If no wrapper exists, use git directly

- **Only commit when asked**
  - Never auto-commit without explicit user request
  - Ask for confirmation before committing
  - Explain what you're committing and why

- **Don't delete unfamiliar files**
  - Before deleting or renaming files, double-check with user or repo instructions
  - Unknown files might be important for other workflows

### Commit Message Format

- **Use Conventional Commits**
  - Format: `type(scope): summary`
  - Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `build`, `ci`, `style`, `perf`
  - Scope: optional, use when helpful (e.g., `cli`, `api`, `ui`)
  - Summary: concise, imperative mood ("add feature" not "added feature")

- **One logical change per commit**
  - Keep commits focused and atomic
  - Don't mix unrelated changes
  - Group related changes together

- **Add rationale when helpful**
  - Include "why" when behavior changes are non-obvious
  - Link to issues or PRs when relevant
  - Keep messages concise but informative

### Release Process

- **Follow documented checklists**
  - Don't invent new release steps
  - Check for `RELEASE.md`, `RELEASING.md`, or similar docs
  - Follow version bump conventions (semver, calver, etc.)

- **Run full verification**
  - Build, test, lint everything before release
  - Check that docs are up-to-date
  - Verify changelog is accurate

---

## Documentation

### Update Rules

- **Update when behavior changes**
  - Keep docs in sync with code
  - Update README, API docs, config examples
  - Include front-matter metadata if used by docs tooling

- **Only create new docs when requested**
  - Edit existing docs by default
  - Don't create redundant documentation
  - Ask before creating new doc files

- **Edit canonical files in place**
  - Update the authoritative doc, don't create duplicates
  - If a doc exists, enhance it rather than creating a new one

### Changelog Guidelines

- **Add entries for user-facing changes**
  - New features, bug fixes, breaking changes
  - Not for internal refactoring, test additions, or dependency bumps (unless they affect users)

- **Follow repo's changelog format**
  - Some use "Unreleased" section at top
  - Others keep latest release at top
  - Check existing changelog for conventions

- **Include attribution**
  - Reference PR numbers: `(#123)`
  - Thank contributors: `Thanks @username!`
  - Link to issues when relevant

---

## Security & Configuration

### Secrets Management

- **Never commit secrets**
  - Use environment variables
  - Use keychain/credential managers
  - Use `.env` files with `.gitignore`
  - Check for common secret patterns before committing

- **Configuration files**
  - Provide example configs (`.example.json`, `.env.example`)
  - Document required environment variables
  - Keep sensitive defaults out of version control

### Safe Defaults

- **Ask about project-wide config changes**
  - Changes to `.gitignore`, `tsconfig.json`, `Package.swift`, etc. need approval
  - These affect all contributors and CI systems
  - Explain rationale before making changes

- **Respect existing security boundaries**
  - Don't bypass CORS, CSP, or other security measures
  - Don't disable security features without explicit approval
  - Document security implications of changes

---

## Multi-Agent Safety

When multiple AI agents may work in the same repository:

- **Do not create/apply/drop git stash entries**
  - Unless explicitly requested
  - Assume other agents may be working
  - Keep unrelated WIP untouched

- **Do not switch branches**
  - Unless explicitly requested
  - Stick to current branch
  - Don't checkout other branches without permission

- **Do not create/remove git worktree checkouts**
  - Unless explicitly requested
  - Don't modify `.worktrees/*` directories

- **When you see unrecognized files**
  - Keep going, focus on your changes
  - Commit only what you've been working on
  - List unrecognized files at end of session if relevant

- **Scope commits to your changes**
  - When user says "commit": commit only your changes
  - When user says "commit all": commit everything in grouped chunks
  - Never accidentally commit another agent's work

- **Focus reports on your edits**
  - Avoid excessive disclaimers unless truly blocked
  - If multiple agents touch same file, continue if safe
  - End with brief "other files present" note only if relevant

---

## Build & Verification

### Pre-Handoff Gate

Before completing work, run the full "green gate" for that repo:

1. **Lint** - Format and style checks pass
2. **Type-check** - No type errors (TypeScript, Swift, etc.)
3. **Test** - All tests pass with expected coverage
4. **Build** - Production build succeeds
5. **Docs** - Doc generation succeeds (if applicable)

### Failure Handling

- **Surface failures clearly**
  - Include exact command output
  - Don't hide or minimize errors
  - Explain what failed and why

- **Don't stop running watchers**
  - Keep existing dev servers and watchers running
  - Don't disrupt other development workflows

- **Use tmux for long-running tasks** (optional)
  - Commands that could hang should run in tmux
  - Don't wrap in infinite polling loops
  - Sleep briefly (≤30s), capture output
  - Document sessions created, clean up when done

---

## Language-Specific Addenda

### TypeScript Projects

**Stack:**
- ESM + strict TypeScript configuration
- Package manager: Respect `pnpm` / `bun` / `npm` as declared
- Formatting: Typically Biome or Prettier (2-space indent common)
- Linting: Biome, oxlint, or ESLint
- Testing: Vitest preferred
- Build: Typically `tsc` to `dist/`

**Guidelines:**
- Maintain strict typing—avoid `any`
- Prefer utility helpers already provided by the repo
- Treat `lint`, `typecheck`, and `test` as mandatory gates
- Keep `node_modules` pristine (don't edit)

**Common Commands:**
```bash
pnpm install       # Install dependencies
pnpm dev          # Development mode
pnpm build        # Production build
pnpm test         # Run tests
pnpm lint         # Lint and format
pnpm typecheck    # Type checking only
```

### Swift Projects

**Stack:**
- Swift 6+ with strict concurrency checking
- Formatting: SwiftFormat (typically 4-space indent, 120 col width)
- Linting: SwiftLint
- Testing: Swift Testing preferred for new code, XCTest for legacy
- Build: SwiftPM or Xcode

**Guidelines:**
- Validate changes with `swift build` and filtered test suites
- Keep concurrency annotations accurate (`Sendable`, `@MainActor`, actors)
- Maintain structured concurrency (no raw threads)
- Use explicit `self` when required (SwiftFormat often enforces)

**Common Commands:**
```bash
swift build              # Build project
swift test               # Run tests
swiftformat .            # Format code
swiftlint lint           # Lint code
swift test --filter ...  # Run specific tests
```

---

## Optional Features

### Tools List

Projects may define custom tools in a `<tools>` block or separate `TOOLS.MD` file. These are project-specific helpers that extend the agent's capabilities.

Example tools:
- `docs:list` - List and summarize documentation
- `mcp:*` - MCP server operations
- Custom scripts in `scripts/` or `Scripts/`

Refer to local repo instructions for available tools.

### Project-Specific Sections

Local AGENTS.md files should include:

- **Project Structure & Modules** - Directory layout, module organization
- **Build, Test, and Development Commands** - Specific commands for this repo
- **Project-Specific Rules** - Unique conventions or requirements
- **Configuration & Secrets** - Where configs live, how to set up
- **Deployment & Release** - How to deploy or publish (if applicable)

---

## Maintenance

### Keeping This File Updated

- **Contribute improvements:** When you notice rules that recur across repos, add them here
- **Sync to local repos:** Update `<shared>` blocks in local files when this file changes
- **Review periodically:** Quarterly review to ensure rules are still relevant
- **Document rationale:** Explain why rules exist, not just what they are

### Sync Options

**Option A: Reference Only (Simplest)**
- Local repos just reference this file
- No `<shared>` block duplication
- Requires network/filesystem access to shared location

**Option B: Copy with Tags (Current mcporter Model)**
- Copy shared content into local `<shared>` blocks
- Provides offline reference
- Requires manual sync to stay current

**Option C: Script-Based Sync (Advanced)**
- Create sync script to update all `<shared>` blocks
- Maintains consistency automatically
- More complex to set up

---

## Version

**Last Updated:** 2026-01-19
**Version:** 1.0.0
**Source:** Consolidated from 13 files across 11 repositories

---

## References

- mcporter/AGENTS.md - Primary source for shared guardrails structure
- clawdis/AGENTS.md - Multi-agent safety rules, changelog workflow
- macos-automator-mcp/CLAUDE.md - Operational learnings pattern
- Matcha/CLAUDE.md - Testing framework migration guidance
- VibeMeter/CLAUDE.md - Release and signing process details

For project-specific details, always consult the local AGENTS.md or CLAUDE.md in each repository.
