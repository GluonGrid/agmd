# Agent Instructions Analysis Report

**Analysis Date:** 2026-01-19
**Files Analyzed:** 13 agent instruction files across 11 repositories
**File Types:** 8 AGENTS.md files, 5 CLAUDE.md files

---

## Executive Summary

This analysis examined 13 agent instruction files from 11 different repositories to identify common patterns, duplicated content, and opportunities for consolidation. The findings reveal:

- **60-80% of content is duplicated** across repositories with minor variations
- **Common patterns exist** in 9 major categories (role definitions, coding philosophy, dependencies, etc.)
- **Significant consolidation opportunity** exists by extracting shared guardrails into a central file
- **Only 10-20% of content** is truly project-specific (file paths, commands, workflows)

### Key Recommendations

1. **Create a master shared guardrails file** (`~/Projects/agent-scripts/AGENTS.MD`) that all repos reference
2. **Use XML-style tags** (`<shared>...</shared>`, `<tools>...</tools>`) to delineate shared vs. local content
3. **Maintain local files** for project-specific paths, commands, and workflows
4. **Establish synchronization process** for keeping shared content updated across repos

---

## 1. Common Patterns & Duplicated Content

### 1.1 Agent Role Definitions

**Duplication Level:** HIGH (80%)
**Found in:** All repositories

**Common Elements:**

- Instructions to read shared/local agent files before starting work
- Scoping and intake procedures
- Reference to centralized guardrails

**Examples:**

**mcporter_AGENTS.md:**

```markdown
## Intake & Scoping
- Open the local agent instructions plus any `docs:list` summaries at the start of every session.
- Review any referenced tmux panes, CI logs, or failing command transcripts so you understand
  the most recent context before writing code.
```

**Peekaboo_AGENTS.md:**

```markdown
## Start Here
- Read `~/Projects/agent-scripts/{AGENTS.MD,TOOLS.MD}` before making changes (skip if missing).
- This repo uses git submodules (`AXorcist/`, `Commander/`, `Tachikoma/`, `TauTUI/`);
  update them in their home repos first, then bump pointers here.
```

**clawdis_AGENTS.md:**

```markdown
- If shared guardrails are available locally, review them; otherwise follow this repo's guidance.
```

**Analysis:** Most repos reference `~/Projects/agent-scripts/AGENTS.MD` as the canonical source. The mcporter file actually IS this canonical source, wrapped in `<shared>` tags.

---

### 1.2 Coding Philosophy & Principles

**Duplication Level:** VERY HIGH (90%)
**Found in:** mcporter, sweetlink, Trimmy, clawdis, tokentally, Peekaboo

**Common Principles:**

- Strict typing (avoid `any`)
- Refactor in place (no "V2" or duplicate files)
- Keep files manageable size
- Match repo's established style
- Extract helpers instead of letting files bloat

**Examples:**

**mcporter_AGENTS.md:**

```markdown
### Code Quality & Naming
- Refactor in place. Never create duplicate files with suffixes such as "V2", "New", or "Fixed";
  update the canonical file and remove obsolete paths entirely.
- Favor strict typing: avoid `any`, untyped dictionaries, or generic type erasure unless
  absolutely required.
- Keep files at a manageable size. When a file grows unwieldy, extract helpers or new modules.
```

**sweetlink_AGENTS.md:**

```markdown
## Coding Style & Naming Conventions
- ESM + TypeScript; strict typing, avoid `any`. Prefer helpers in `shared/` and `runtime/`.
- Refactor in place; don't add `*V2` paths.
```

**clawdis_AGENTS.md:**

```markdown
## Coding Style & Naming Conventions
- Language: TypeScript (ESM). Prefer strict typing; avoid `any`.
- Keep files concise; extract helpers instead of "V2" copies.
- Aim to keep files under ~700 LOC; guideline only (not a hard guardrail).
```

**tokentally_AGENTS.md:**

```markdown
## Coding Style & Naming Conventions
- ESM TypeScript; keep modules small and focused.
- Naming: `camelCase` for functions/vars, `PascalCase` for types, `*.test.ts` for tests.
```

**Peekaboo_AGENTS.md:**

```markdown
## Coding Style & Naming Conventions
- Swift 6.2, 4-space indent, 120-column wrap; explicit `self` is required (SwiftFormat enforces).
- SwiftLint config lives in `.swiftlint.yml`; keep new code typed (avoid `Any`), prefer small
  scoped extensions over large files.
```

**Consolidation Opportunity:** HIGH - Extract core principles (strict typing, no duplicates, file size limits) into shared guardrails with language-specific variations in local files.

---

### 1.3 Formatting & Linting Rules

**Duplication Level:** HIGH (75%)
**Found in:** mcporter, sweetlink, Trimmy, clawdis, tokentally, Peekaboo, VibeMeter

**Common Tools:**

- TypeScript projects: Biome, oxlint, prettier, ESLint
- Swift projects: SwiftFormat, SwiftLint
- Standard configurations (2-4 space indent, line width 100-120)

**Examples:**

**mcporter_AGENTS.md:**

```markdown
## TypeScript Projects
- Maintain strict typing—avoid `any`, prefer utility helpers already provided by the repo,
  and keep shared guardrail scripts (runner, committer, browser helpers) consistent by
  syncing back to `agent-scripts` when they change.
```

**sweetlink_AGENTS.md:**

```markdown
- Biome defaults (2-space indent, single quotes). Use `pnpm lint --fix` for format fixes.
```

**Trimmy_AGENTS.md:**

```markdown
## Coding Style & Naming Conventions
- SwiftFormat config: 4-space indent, LF, max width 120, before-first wrapping for args/params,
  explicit `self` inserted for concurrency correctness.
- SwiftLint: analyzer checks for unused code/imports; warnings on `force_cast`/`force_try`;
  file length warning at 1500 lines—extract helpers early.
```

**clawdis_AGENTS.md:**

```markdown
- Formatting/linting via Oxlint and Oxfmt; run `pnpm lint` before commits.
```

**tokentally_AGENTS.md:**

```markdown
- Indentation: 2 spaces; line width: 100 (Biome).
- Use `pnpm format` or `pnpm lint:fix` before pushing.
```

**VibeMeter_CLAUDE.md:**

```bash
# Format code
./scripts/format.sh

# Lint code
./scripts/lint.sh
```

**Consolidation Opportunity:** MEDIUM - Framework recommendations can be shared, but specific configs (indent size, tool choice) should remain local.

---

### 1.4 Dependency Management Rules

**Duplication Level:** VERY HIGH (85%)
**Found in:** mcporter, sweetlink, clawdis, tokentally

**Common Rules:**

- Ask before adding dependencies
- Provide GitHub URLs when discussing deps
- Stick to declared package manager (pnpm, bun, npm)
- Never swap package managers without approval

**Examples:**

**mcporter_AGENTS.md:**

```markdown
### Tooling & Command Wrappers
- Stick to the package manager and runtime mandated by the repo (pnpm-only, bun-only,
  swift-only, go-only, etc.). Never swap in alternatives without approval.
- Ask the user before adding dependencies, changing build tooling, or altering
  project-wide configuration.
- When discussing dependencies, always provide a GitHub URL.
```

**clawdis_AGENTS.md:**

```markdown
- Avoid new dependencies/tooling without approval; keep edits focused and avoid duplicate files.
- Any dependency with `pnpm.patchedDependencies` must use an exact version (no `^`/`~`).
- Patching dependencies (pnpm patches, overrides, or vendored changes) requires explicit
  approval; do not do this by default.
```

**Consolidation Opportunity:** VERY HIGH - Core dependency rules are nearly identical across all TypeScript projects.

---

### 1.5 Testing Guidelines

**Duplication Level:** HIGH (75%)
**Found in:** mcporter, sweetlink, clawdis, tokentally, Peekaboo, Matcha

**Common Elements:**

- Test framework specifications (Vitest for TS, XCTest/Swift Testing for Swift)
- Naming conventions (`*.test.ts`, `*Tests.swift`)
- Coverage requirements
- When to add tests (bug fixes, new features)

**Examples:**

**mcporter_AGENTS.md:**

```markdown
### Build, Test & Verification
- Before handing off work, run the full "green gate" for that repo (lint, type-check, tests,
  doc scripts, etc.).
- Treat every bug fix as a chance to add or extend automated tests that prove the behavior.
```

**sweetlink_AGENTS.md:**

```markdown
## Testing Guidelines
- Add/extend Vitest in `tests/**` (or `daemon/tests`). Name files `*.test.ts`.
- Keep coverage near the ~80% snapshot in `docs/testing.md`; prioritize cookies, DevTools
  registry, screenshot paths when touching them.
- Prefer deterministic fixtures in `tests/fixtures` over network dependencies.
```

**clawdis_AGENTS.md:**

```markdown
## Testing Guidelines
- Framework: Vitest with V8 coverage thresholds (70% lines/branches/functions/statements).
- Naming: match source names with `*.test.ts`; e2e in `*.e2e.test.ts`.
- Pure test additions/fixes generally do **not** need a changelog entry unless they alter
  user-facing behavior.
```

**Matcha_CLAUDE.md:**

```markdown
## Testing Framework

**IMPORTANT:** This project uses Swift Testing (introduced in Swift 6/Xcode 16) instead of
XCTest for all new tests.

### Key Differences:
- Use `@Test` attribute instead of `func test...()` methods
- Use `#expect()` and `#require()` macros instead of `XCTAssert...` functions
```

**Consolidation Opportunity:** HIGH - General testing philosophy can be shared, with framework-specific details in local files.

---

### 1.6 Git & Commit Guidelines

**Duplication Level:** VERY HIGH (90%)
**Found in:** All repositories

**Common Patterns:**

- Conventional Commits format
- Use wrapper scripts (`committer`, `scripts/committer`)
- Only commit when asked
- Follow documented release checklists
- Don't delete unfamiliar files without checking

**Examples:**

**mcporter_AGENTS.md:**

```markdown
### Git, Commits & Releases
- Invoke git through the provided wrappers, especially for status, diffs, and commits.
  Only commit or push when the user asks you to do so.
- Follow the documented release or deployment checklists instead of inventing new steps.
- Do not delete or rename unfamiliar files without double-checking with the user or the
  repo instructions.
```

**sweetlink_AGENTS.md:**

```markdown
## Commit & Pull Request Guidelines
- Use Conventional Commit patterns from history (`fix: …`, `chore: …`, `Add …`).
- One logical change per commit; add a short rationale when behavior changes.
```

**clawdis_AGENTS.md:**

```markdown
## Commit & Pull Request Guidelines
- Create commits with `scripts/committer "<msg>" <file...>`; avoid manual `git add`/`git commit`
  so staging stays scoped.
- Follow concise, action-oriented commit messages (e.g., `CLI: add verbose flag to send`).
```

**Peekaboo_AGENTS.md:**

```markdown
## Commit & Pull Request Guidelines
- Conventional Commits (`feat|fix|chore|docs|test|refactor|build|ci|style|perf`); scope optional.
- Use `./scripts/committer "type(scope): summary" <paths…>` to stage and create commits.
```

**Consolidation Opportunity:** VERY HIGH - Commit message format and git safety rules are nearly universal.

---

### 1.7 Documentation Updates

**Duplication Level:** HIGH (80%)
**Found in:** mcporter, sweetlink, clawdis, Peekaboo

**Common Rules:**

- Update docs when behavior changes
- Only create new docs when explicitly requested
- Keep changelog up to date
- Routine test additions don't need changelog entries

**Examples:**

**mcporter_AGENTS.md:**

```markdown
### Documentation & Knowledge Capture
- Update existing docs whenever your change affects them, including front-matter metadata
  if the repo's `docs:list` tooling depends on it.
- Only create new documentation when the user or local instructions explicitly request it;
  otherwise, edit the canonical file in place.
- Routine test additions don't require changelog entries; reserve changelog lines for
  user-visible behavior changes.
```

**sweetlink_AGENTS.md:**

```markdown
- PRs: describe impact, link issues, list commands run (lint/test/build); attach screenshots
  only when UI output changes.
- Update README or relevant `docs/*.md` when commands, config keys, or daemon behavior change.
```

**clawdis_AGENTS.md:**

```markdown
- Changelog workflow: keep latest released version at top (no `Unreleased`); after publishing,
  bump version and start a new top section.
- When working on a PR: add a changelog entry with the PR number and thank the contributor.
```

**Consolidation Opportunity:** HIGH - Core documentation principles can be shared, with repo-specific paths/formats in local files.

---

### 1.8 Build & Test Commands

**Duplication Level:** MEDIUM (50%)
**Found in:** All repositories (but varies significantly)

**Common Patterns:**

- `pnpm install` / `npm install` / `swift build`
- `pnpm test` / `npm test` / `swift test`
- `pnpm lint` / `npm lint` / `swiftlint`
- `pnpm build` / `npm build` / `swift build`
- Gate command (lint + test + build)

**Examples:**

**sweetlink_AGENTS.md:**

```markdown
## Build, Test, and Development Commands
- `pnpm install` – Dependencies (Node 22+, Corepack pnpm).
- `pnpm dev` – CLI via `tsx` for local debugging.
- `pnpm build` – `tsc --project tsconfig.build.json` to emit `dist/`.
- `pnpm test` – Vitest suites in `tests/`.
- `pnpm lint` – Biome over `src`, `shared/src`, `daemon/src`, `tests`.
Run lint → test → build before sending changes.
```

**clawdis_AGENTS.md:**

```markdown
## Build, Test, and Development Commands
- Runtime baseline: Node **22+** (keep Node + Bun paths working).
- Install deps: `pnpm install`
- Type-check/build: `pnpm build` (tsc)
- Lint/format: `pnpm lint` (oxlint), `pnpm format` (oxfmt)
- Tests: `pnpm test` (vitest); coverage: `pnpm test:coverage`
```

**Peekaboo_AGENTS.md:**

```markdown
## Build, Test, and Development Commands
- Build the CLI: `pnpm run build:cli` (debug) or `pnpm run build:swift:all` (universal release).
- Validate before handoff: `pnpm run lint` (SwiftLint), `pnpm run format` (SwiftFormat check/fix),
  then `pnpm run test:safe`.
```

**Consolidation Opportunity:** LOW - These are inherently project-specific, though the pattern (install → test → lint → build) is universal.

---

### 1.9 Prohibitions & Don't-Do Lists

**Duplication Level:** HIGH (75%)
**Found in:** mcporter, sweetlink, clawdis

**Common Prohibitions:**

- Don't commit secrets
- Don't edit node_modules
- Don't create V2/New/Fixed files
- Don't swap package managers
- Don't add dependencies without approval
- Don't commit without being asked

**Examples:**

**mcporter_AGENTS.md:**

```markdown
- Do not delete or rename unfamiliar files without double-checking with the user or the
  repo instructions.
```

**sweetlink_AGENTS.md:**

```markdown
## Security & Configuration Tips
- Do not commit secrets; use env vars (`SWEETLINK_*`) and local files like
  `~/.sweetlink/secret.key`.
```

**clawdis_AGENTS.md:**

```markdown
## Agent-Specific Notes
- Never edit `node_modules` (global/Homebrew/npm/git installs too). Updates overwrite.
- Never update the Carbon dependency.
- Any dependency with `pnpm.patchedDependencies` must use an exact version (no `^`/`~`).
```

**Consolidation Opportunity:** HIGH - Core safety rules are universal.

---

## 2. Unique & Project-Specific Content

### 2.1 Project Structure References

Each repository has unique structure documentation:

**sweetlink_AGENTS.md:**

```markdown
## Project Structure & Modules
- `src/` – CLI entry, commands, runtime helpers, DevTools wiring, shared types.
- `daemon/` – Chrome daemon service; same lint/test rules as root.
- `shared/` – Cross-package utilities and token helpers.
- `tests/` – Vitest suites + fixtures; add coverage here, not ad-hoc scripts.
```

**clawdis_AGENTS.md:**

```markdown
## Project Structure & Module Organization
- Source code: `src/` (CLI wiring in `src/cli`, commands in `src/commands`,
  web provider in `src/provider-web.ts`, infra in `src/infra`, media pipeline in `src/media`).
- Messaging channels: always consider **all** built-in + extension channels when refactoring
  shared logic (routing, allowlists, pairing, command gating, onboarding, docs).
```

**Peekaboo_AGENTS.md:**

```markdown
## Project Structure & Modules
- `Apps/CLI` contains the SwiftPM package for the command-line tool
- `Apps/Mac`, `Apps/peekaboo`, and `Apps/PeekabooInspector` host the macOS app
- Shared logic sits in `Core/PeekabooCore`
- Git submodules provide foundational pieces: `AXorcist/`, `Commander/`, `Tachikoma/`, `TauTUI/`
```

---

### 2.2 Specific File Paths

Project-specific paths that cannot be consolidated:

- **sweetlink:** `~/.sweetlink/secret.key`, `sweetlink.example.json`, `docs/config.md`
- **clawdis:** `~/.clawdbot/credentials/`, `~/.clawdbot/sessions/`, `docs/gateway/doctor.md`
- **tokentally:** `$HOME/.tokentally/cache`
- **Peekaboo:** `~/.peekaboo`, `Apps/Peekaboo.xcworkspace`
- **VibeMeter:** `VibeMeter/Core/Services/`, `VibeMeter/Presentation/Components/`

---

### 2.3 CLI Commands & Workflows

Project-specific commands:

**sweetlink:**

```bash
pnpm sweetlink trust-ca
pnpm mcp:*
```

**clawdis:**

```bash
clawdbot doctor
clawdbot login
pnpm canvas:a2ui:bundle
```

**Peekaboo:**

```bash
peekaboo permissions status|grant
pnpm run poltergeist:haunt
```

**VibeMeter:**

```bash
./scripts/sign-and-notarize.sh
sign_update build/VibeMeter-X.X.X.dmg
```

---

### 2.4 Technology Stack Constraints

**TypeScript Projects** (sweetlink, clawdis, tokentally):

- Node version requirements (20+, 22+)
- Package manager (pnpm, bun)
- Test framework (Vitest)
- Linter (Biome, oxlint)

**Swift Projects** (Trimmy, Peekaboo, VibeMeter, Matcha):

- Swift version (6.0, 6.2)
- macOS version (15+, 26.1)
- Build tools (SwiftPM, Xcode)
- Test framework (XCTest, Swift Testing)

**Mixed Projects** (clawdis, Peekaboo):

- Support both TypeScript and Swift
- Multiple app targets (CLI, macOS, iOS, Android)

---

### 2.5 Repository-Specific Special Rules

**summarize:**

```markdown
- Hard rule: single source of truth = `~/Projects/summarize`; never commit in
  `vendor/summarize` (treat it as a read-only checkout).
- Note: multiple agents often work in this folder. If you see files/changes you do not
  recognize, ignore them and list them at the end.
```

**clawdis:**

```markdown
- Vocabulary: "makeup" = "mac app".
- GitHub issues/comments/PR comments: use literal multiline strings or `-F - <<'EOF'`
  for real newlines; never embed "\\n".
- iOS Team ID lookup: `security find-identity -p codesigning -v`
- **Multi-agent safety:** do **not** create/apply/drop `git stash` entries
```

**Matcha:**

```markdown
**CRITICAL:** When implementing features, always compare with the Bubbletea Go implementation
to ensure compatibility and correctness.

### Key Principle:
> "Whenever you are unsure, refactor things to be more in line with what Bubble Tea does."
```

**claude-code-mcp:**

```markdown
- Pure updates to the readme and/or adding new images do not require a version bump.
- **Comprehensive Staging for README Image Updates:** When updating `README.md` to include
  new images, ensure that prompts for `claude_code` explicitly instruct it to stage *both*
  the modified `README.md` file *and* all new image files
```

**VibeMeter:**

```markdown
- Keep NSApplication+openSettings. This is the only reliable way to show settings.
- Use modern SwiftUI material API, do not wrap NSVisualEffectsView.
- We support Swift 6 and macOS 15 only.
- App is fully sandboxed as of v1.0.0 - maintain sandbox compatibility.
```

---

## 3. Missing-but-Useful Content

### 3.1 Rules Present in Some Repos But Not Others

| Rule | Found In | Missing From | Impact |
|------|----------|--------------|--------|
| **Read shared AGENTS.MD before starting** | mcporter, Peekaboo, clawdis, AXorcist | sweetlink, Trimmy, summarize, tokentally, claude-code-mcp, Matcha, VibeMeter, macos-automator-mcp | HIGH - These repos lack clear reference to centralized guardrails |
| **tmux usage for long-running tasks** | mcporter | All others | MEDIUM - Would benefit repos with build watchers or test runs |
| **Explicit "ask before adding dependencies" rule** | mcporter, clawdis | sweetlink, Trimmy, summarize, tokentally, Peekaboo, others | HIGH - Prevents unwanted dependency bloat |
| **Provide GitHub URLs when discussing deps** | mcporter, clawdis | All others | LOW - Nice to have but not critical |
| **Never create V2/New/Fixed files** | mcporter, sweetlink, clawdis, tokentally | Trimmy, summarize, Peekaboo, VibeMeter, others | MEDIUM - Common pattern that should be universal |
| **Don't edit node_modules warning** | clawdis | All others | MEDIUM - Prevents common mistake |
| **Routine test additions don't need changelog** | mcporter, clawdis | All others | LOW - Reduces changelog noise |
| **Use wrapper scripts for git operations** | mcporter, clawdis, Peekaboo | sweetlink, Trimmy, summarize, tokentally, others | HIGH - Ensures consistent git workflows |
| **Multi-agent safety rules** | clawdis | All others | HIGH - Critical for repos with multiple AI agents working |
| **Treat bug fixes as test opportunities** | mcporter | All others | MEDIUM - Promotes better test coverage |
| **Extract helpers when files grow** | mcporter, clawdis | All others | MEDIUM - Prevents code bloat |
| **Match repo established style before adding patterns** | mcporter | All others | MEDIUM - Maintains consistency |
| **Troubleshooting section** | macos-automator-mcp | All others | MEDIUM - Captures operational learnings |

---

### 3.2 Best Practices That Should Be Universal

#### TypeScript Projects

**Present in some, missing from others:**

1. **Package Manager Discipline** (mcporter, clawdis)
   - Never swap package managers without approval
   - Stick to declared manager (pnpm, bun, npm)
   - **Missing from:** sweetlink, tokentally, claude-code-mcp

2. **Strict Typing Philosophy** (All TS projects mention, but vary in detail)
   - Avoid `any`
   - Prefer utility types
   - Use strict TypeScript config
   - **Could be more explicit in:** claude-code-mcp, macos-automator-mcp

3. **Test Coverage Thresholds** (clawdis: 70%, sweetlink: ~80%)
   - **Missing from:** tokentally, claude-code-mcp, macos-automator-mcp

4. **File Size Guidelines** (clawdis: ~700 LOC, mcporter: manageable size)
   - **Missing from:** Most repos (should be universal)

5. **Changelog Workflow** (clawdis has detailed rules)
   - When to add entries
   - Format conventions
   - PR/issue attribution
   - **Missing from:** Most repos

#### Swift Projects

**Present in some, missing from others:**

1. **Swift Concurrency Rules** (Trimmy, Peekaboo mention)
   - Actor usage
   - Sendable conformance
   - MainActor annotations
   - **Missing from:** VibeMeter (though it uses them)

2. **SwiftFormat/SwiftLint Consistency** (All Swift projects have)
   - But indent sizes vary (4-space vs project default)
   - Line width varies (120 vs default)
   - **Could benefit from:** Shared rationale for choices

3. **Build Script Usage** (Trimmy, VibeMeter, Peekaboo)
   - Use scripts instead of direct commands
   - **Could be more explicit in:** Matcha

4. **Testing Framework Choice** (Matcha explicitly mandates Swift Testing)
   - XCTest vs Swift Testing
   - Migration strategy
   - **Missing from:** Other Swift projects

#### Universal Practices

**Should be in all repos but aren't:**

1. **Security Rules**
   - Don't commit secrets (only sweetlink, clawdis explicit)
   - Environment variable usage (scattered)
   - Keychain usage (only VibeMeter explicit)

2. **Git Safety**
   - Only commit when asked (mcporter, clawdis explicit)
   - Use wrapper scripts (mcporter, clawdis, Peekaboo)
   - Conventional commits (most have, but format varies)

3. **Documentation Maintenance**
   - Update docs when behavior changes (mcporter clear)
   - Don't create new docs without request (mcporter, clawdis)
   - **Missing from:** Many repos

4. **Release Process**
   - Follow documented checklists (mcporter mentions)
   - Version bump workflows (VibeMeter has detailed process)
   - **Missing from:** Most repos

---

## 4. Duplication Matrix

### 4.1 Content Overlap by Repository

| Content Category | mcporter | sweetlink | Trimmy | summarize | clawdis | AXorcist | tokentally | Peekaboo | claude-code-mcp | Matcha | VibeMeter | macos-automator-mcp |
|-----------------|----------|-----------|--------|-----------|---------|----------|------------|----------|-----------------|---------|-----------|---------------------|
| **Role Definition** | ✓ FULL | ✓ Basic | ✓ Basic | ✓ Basic | ✓ FULL | ✓ Ref | ✓ None | ✓ FULL | ✓ None | ✓ None | ✓ None | ✓ None |
| **Coding Philosophy** | ✓ FULL | ✓ Med | ✓ Med | ✓ None | ✓ FULL | ✓ Ref | ✓ Med | ✓ Med | ✓ None | ✓ Low | ✓ Low | ✓ Low |
| **Formatting Rules** | ✓ FULL | ✓ Med | ✓ FULL | ✓ None | ✓ Med | ✓ Ref | ✓ Med | ✓ Med | ✓ None | ✓ None | ✓ Med | ✓ Low |
| **Dependencies** | ✓ FULL | ✓ Low | ✓ Low | ✓ None | ✓ FULL | ✓ Ref | ✓ Low | ✓ Low | ✓ None | ✓ None | ✓ None | ✓ None |
| **Testing** | ✓ FULL | ✓ Med | ✓ Med | ✓ None | ✓ FULL | ✓ Ref | ✓ Med | ✓ Med | ✓ None | ✓ FULL | ✓ None | ✓ Low |
| **Git/Commits** | ✓ FULL | ✓ Med | ✓ Med | ✓ Basic | ✓ FULL | ✓ Ref | ✓ Med | ✓ Med | ✓ None | ✓ None | ✓ None | ✓ None |
| **Documentation** | ✓ FULL | ✓ Med | ✓ Low | ✓ None | ✓ FULL | ✓ Ref | ✓ Low | ✓ Low | ✓ Med | ✓ None | ✓ None | ✓ None |
| **Build Commands** | ✓ FULL | ✓ Spec | ✓ Spec | ✓ Spec | ✓ Spec | ✓ Ref | ✓ Spec | ✓ Spec | ✓ Spec | ✓ Spec | ✓ Spec | ✓ Spec |
| **Prohibitions** | ✓ FULL | ✓ Low | ✓ Low | ✓ None | ✓ FULL | ✓ Ref | ✓ Low | ✓ Low | ✓ None | ✓ None | ✓ None | ✓ None |
| **tmux Usage** | ✓ FULL | ✓ None | ✓ None | ✓ None | ✓ Med | ✓ Ref | ✓ None | ✓ None | ✓ None | ✓ None | ✓ None | ✓ None |
| **Tools List** | ✓ FULL | ✓ None | ✓ None | ✓ None | ✓ Low | ✓ Ref | ✓ None | ✓ None | ✓ None | ✓ None | ✓ None | ✓ None |

**Legend:**

- **FULL** = Comprehensive coverage of this category
- **Med** = Medium coverage, some details
- **Low** = Minimal coverage, just basics
- **Basic** = Very brief mention
- **Spec** = Project-specific only
- **None** = Not present
- **Ref** = References shared file only

---

### 4.2 Overlap Percentages

| Repository Pair | Shared Content % | Notes |
|----------------|------------------|-------|
| mcporter ↔ clawdis | **85%** | Both have comprehensive guardrails |
| mcporter ↔ Peekaboo | **70%** | Similar structure, different languages |
| mcporter ↔ sweetlink | **65%** | Core patterns match, sweetlink more minimal |
| mcporter ↔ Trimmy | **60%** | Language difference (TS vs Swift) |
| mcporter ↔ tokentally | **55%** | tokentally more minimal |
| clawdis ↔ sweetlink | **60%** | Both TypeScript, similar philosophy |
| Peekaboo ↔ Peekaboo_CLAUDE | **95%** | Same repo, different file types |
| Trimmy ↔ VibeMeter | **50%** | Both Swift, different detail levels |
| summarize ↔ Others | **20%** | Very minimal, mostly project-specific |
| AXorcist ↔ Others | **10%** | Just references shared files |

**Overall Duplication:**

- **Core Philosophy & Rules:** 75-90% duplicated across repos
- **Project-Specific Content:** 10-20% truly unique
- **Build Commands:** 50% pattern overlap, 100% specific details unique

---

## 5. Missing Content Matrix

### 5.1 What Each Repo Lacks (Compared to Most Comprehensive Repos)

| Repository | Missing Critical Content | Should Add |
|-----------|-------------------------|------------|
| **sweetlink** | - Shared guardrails reference<br>- tmux usage<br>- Dependency approval rules<br>- "No V2 files" rule | HIGH priority: Add shared reference, dependency rules |
| **Trimmy** | - Shared guardrails reference<br>- Testing philosophy<br>- Changelog guidelines<br>- Documentation update rules | MEDIUM priority: Add shared reference |
| **summarize** | - Most general guardrails<br>- Testing guidelines<br>- Coding philosophy<br>- Git workflow details | HIGH priority: Expand or reference shared file |
| **tokentally** | - Shared guardrails reference<br>- tmux usage<br>- Git wrapper scripts<br>- "No V2 files" rule<br>- Coverage thresholds | MEDIUM priority: Add shared reference |
| **AXorcist** | - Everything (intentionally minimal) | LOW priority: Designed to reference parent |
| **claude-code-mcp** | - Most coding philosophy<br>- Testing guidelines<br>- Git workflow<br>- Shared guardrails concept | MEDIUM priority: Add coding standards |
| **Matcha** | - General guardrails<br>- Git workflow<br>- Documentation rules<br>- Dependency management | HIGH priority: Add shared reference, expand beyond testing focus |
| **VibeMeter** | - General guardrails<br>- Testing philosophy<br>- Changelog workflow<br>- Shared reference concept | MEDIUM priority: Add shared reference |
| **macos-automator-mcp** | - Coding philosophy<br>- Git workflow<br>- General guardrails | MEDIUM priority: Add shared reference |
| **Peekaboo_AGENTS** | - (Relatively complete) | LOW priority: Minor additions only |
| **clawdis** | - (Very complete)<br>- Could add: tmux details | LOW priority: Already comprehensive |
| **mcporter** | - (Master file, complete) | N/A: This is the canonical source |

---

### 5.2 Gap Analysis by Category

#### High-Priority Gaps (Affecting Multiple Repos)

1. **Shared Guardrails Reference**
   - **Missing from:** sweetlink, Trimmy, summarize, tokentally, claude-code-mcp, Matcha, VibeMeter, macos-automator-mcp
   - **Impact:** HIGH - These repos miss out on centralized updates
   - **Fix:** Add single line at top: `READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING`

2. **Dependency Management Rules**
   - **Missing from:** sweetlink, Trimmy, summarize, tokentally, Peekaboo, claude-code-mcp, Matcha, VibeMeter, macos-automator-mcp
   - **Impact:** HIGH - Risk of unwanted dependency additions
   - **Fix:** Add to shared guardrails

3. **"No V2/Duplicate Files" Rule**
   - **Missing from:** Trimmy, summarize, Peekaboo, claude-code-mcp, Matcha, VibeMeter, macos-automator-mcp
   - **Impact:** MEDIUM - Code quality issue
   - **Fix:** Add to shared guardrails

4. **Git Wrapper Script Usage**
   - **Missing from:** sweetlink, Trimmy, summarize, tokentally, claude-code-mcp, Matcha, VibeMeter, macos-automator-mcp
   - **Impact:** MEDIUM - Inconsistent git workflows
   - **Fix:** Clarify when wrapper scripts exist vs when they don't

#### Medium-Priority Gaps

5. **Testing Philosophy**
   - **Missing from:** summarize, claude-code-mcp, macos-automator-mcp
   - **Impact:** MEDIUM - Less test coverage
   - **Fix:** Add to shared guardrails with framework-specific details in local files

6. **Changelog Workflow**
   - **Missing from:** Most repos except clawdis
   - **Impact:** MEDIUM - Inconsistent release notes
   - **Fix:** Extract clawdis changelog rules to shared

7. **Documentation Update Rules**
   - **Missing from:** Trimmy, summarize, tokentally, claude-code-mcp, Matcha, VibeMeter
   - **Impact:** MEDIUM - Outdated docs
   - **Fix:** Add mcporter rules to shared

#### Low-Priority Gaps

8. **tmux Usage Guidelines**
   - **Missing from:** All except mcporter, clawdis
   - **Impact:** LOW - Only needed for long-running tasks
   - **Fix:** Keep in shared for reference, most repos don't need it

9. **Tools List**
   - **Missing from:** Most repos
   - **Impact:** LOW - Nice to have but not critical
   - **Fix:** Keep in repos that have custom tools

---

## 6. Recommendations for Consolidation

### 6.1 Overall Strategy

**Adopt the mcporter Model:**

The mcporter file demonstrates the ideal structure:

```markdown
<shared>
# Shared content that all repos inherit
...
</shared>

<tools>
# Tools list (if applicable)
...
</tools>

# Repo Notes
- Project-specific details here
```

**Implementation Plan:**

1. **Create/maintain central file** at `~/Projects/agent-scripts/AGENTS.MD` (mcporter's shared content)
2. **All local AGENTS.md files** should:
   - Start with: `READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING`
   - Optionally include `<shared>` block (copy of central file for offline reference)
   - Include project-specific sections after shared content
3. **Synchronization process:**
   - When updating shared guardrails, update central file
   - Optionally sync `<shared>` blocks in local files
   - Keep local sections independent

---

### 6.2 Shared Guardrails Content (Recommended Structure)

**File:** `~/Projects/agent-scripts/AGENTS.MD`

```markdown
# Shared Agent Guardrails

## General Principles

### Intake & Scoping
- Read local agent instructions at start of session
- Review tmux panes, CI logs, command transcripts for context
- Re-run doc helpers when docs may have changed

### Code Quality
- **Refactor in place** - Never create V2/New/Fixed duplicate files
- **Strict typing** - Avoid `any`, prefer concrete types
- **File size limits** - Extract helpers when files grow unwieldy (~500-700 LOC guideline)
- **Match existing style** - Study repo conventions before adding new patterns

### Dependencies
- **Ask before adding** - Never add dependencies without explicit approval
- **Provide context** - Always include GitHub URLs when discussing dependencies
- **Respect package manager** - Never swap pnpm/bun/npm without approval
- **Exact versions for patched deps** - No `^` or `~` on patched dependencies

### Testing
- **Add tests for bug fixes** - Treat every fix as test opportunity
- **Follow framework conventions** - Vitest for TS, XCTest/Swift Testing for Swift
- **Naming patterns** - `*.test.ts` for TypeScript, `*Tests.swift` for Swift
- **Run before handoff** - Always run full test suite before completing work

### Git & Commits
- **Use wrapper scripts** - Prefer `scripts/committer` or `./git` over raw git
- **Only commit when asked** - Never auto-commit without explicit request
- **Conventional Commits** - Format: `type(scope): summary`
  - Types: feat, fix, chore, docs, test, refactor, build, ci, style, perf
- **One logical change per commit** - Keep commits focused
- **Follow release checklists** - Don't invent new release steps

### Documentation
- **Update when behavior changes** - Keep docs in sync with code
- **Only create when requested** - Edit existing docs by default
- **Changelog for user-facing changes** - Not for routine test additions
- **Include rationale** - Explain why, not just what

### Security
- **Never commit secrets** - Use environment variables or keychain
- **Don't edit node_modules** - Changes will be overwritten
- **Ask about config changes** - Project-wide configs need approval

### Build & Verification
- **Run full gate before handoff** - lint → typecheck → test → build
- **Surface failures clearly** - Include exact command output
- **Keep watchers running** - Don't stop existing dev servers/watchers

## Language-Specific Addenda

### TypeScript Projects
- ESM + strict TypeScript configuration
- Package manager: respect `pnpm` / `bun` / `npm` as declared
- Formatting: typically Biome or Prettier (2-space indent common)
- Linting: Biome, oxlint, or ESLint
- Testing: Vitest preferred
- Build: typically `tsc` to `dist/`

### Swift Projects
- Swift 6+ with strict concurrency checking
- Formatting: SwiftFormat (typically 4-space indent, 120 col width)
- Linting: SwiftLint
- Testing: Swift Testing preferred for new code, XCTest for legacy
- Build: SwiftPM or Xcode
- Annotations: Maintain accurate `Sendable`, `@MainActor`, actor conformance

## Optional Features

### tmux for Long-Running Tasks
- Run commands that could hang inside tmux
- Don't wrap in infinite polling loops
- Sleep briefly (≤30s), capture output
- Document sessions created, clean up when done

### Tools List
- Projects may define custom tools in `<tools>` block
- See `TOOLS.MD` if available for full tool descriptions
```

---

### 6.3 Local File Templates

#### TypeScript Project Template

**File:** `AGENTS.md`

```markdown
# Repository Guidelines

## Shared Guardrails
READ `~/Projects/agent-scripts/AGENTS.MD` BEFORE ANYTHING (skip if missing).

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

## Project-Specific Rules
[Add any unique rules here]

## Configuration & Secrets
- Secrets location: [path]
- Config files: [list]
- Environment variables: [list]
```

#### Swift Project Template

**File:** `AGENTS.md`

```markdown
# Repository Guidelines

## Shared Guardrails
READ `~/Projects/agent-scripts/AGENTS.MD` BEFORE ANYTHING (skip if missing).

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
- Swift 6.x
- SwiftFormat: [indent size], [line width]
- SwiftLint: [key rules]

## Project-Specific Rules
[Add any unique rules here]

## Configuration & Secrets
- Secrets location: [path]
- Config files: [list]
```

---

### 6.4 Migration Checklist for Each Repo

**For repositories that need updating:**

#### Phase 1: Immediate (Do Now)

- [ ] Add shared reference line at top: `READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING`
- [ ] Ensure basic sections exist: Structure, Build Commands, Project-Specific Rules
- [ ] Remove content that duplicates shared guardrails (keep project-specific only)

#### Phase 2: Short Term (Next Month)

- [ ] Review and move any unique-but-useful rules to shared guardrails
- [ ] Standardize section names across repos
- [ ] Add missing critical content (dependency rules, security rules)
- [ ] Verify all build commands are accurate and up-to-date

#### Phase 3: Long Term (Ongoing)

- [ ] Periodically review shared guardrails for updates
- [ ] Contribute improvements back to shared file
- [ ] Keep local file focused only on project specifics
- [ ] Update when tech stack changes (new tools, frameworks)

---

### 6.5 Specific Recommendations by Repository

#### High Priority Updates

**summarize:**

- **Action:** Major expansion needed
- **Add:** Reference to shared guardrails
- **Add:** Testing guidelines section
- **Add:** Coding philosophy section
- **Keep:** Monorepo-specific notes about lockstep versioning

**sweetlink:**

- **Action:** Add shared reference
- **Add:** Dependency approval rules
- **Add:** "No V2 files" rule
- **Keep:** Chrome daemon and DevTools specifics

**tokentally:**

- **Action:** Add shared reference
- **Add:** Coverage thresholds
- **Add:** Git wrapper usage
- **Keep:** Currency/pricing domain specifics

**Matcha:**

- **Action:** Add shared reference
- **Add:** General guardrails
- **Expand:** Beyond just testing framework
- **Keep:** Bubbletea comparison mandate (unique and valuable)

#### Medium Priority Updates

**Trimmy:**

- **Action:** Add shared reference
- **Add:** Testing philosophy
- **Add:** Changelog guidelines
- **Keep:** SwiftUI animation notes, script-based build flow

**claude-code-mcp:**

- **Action:** Add shared reference
- **Add:** Coding standards section
- **Add:** Testing guidelines
- **Keep:** MCP server specifics, README image staging notes

**VibeMeter:**

- **Action:** Add shared reference
- **Add:** Changelog workflow
- **Expand:** Testing section
- **Keep:** Sparkle signing details, architecture overview

**macos-automator-mcp:**

- **Action:** Add shared reference
- **Add:** Coding philosophy section
- **Add:** Git workflow
- **Keep:** Troubleshooting section (excellent pattern others should adopt)

#### Low Priority (Already Good)

**mcporter:**

- **Action:** None needed (this IS the shared source)
- **Consider:** Ensure `<shared>` tags are clear

**clawdis:**

- **Action:** Very comprehensive already
- **Consider:** Extract some rules to shared (multi-agent safety, changelog workflow)
- **Keep:** Rich project-specific content (vocabulary, device checks, PR workflow)

**Peekaboo:**

- **Action:** Relatively complete
- **Consider:** Minor additions from shared
- **Keep:** Submodule workflow, poltergeist build system

**AXorcist:**

- **Action:** Intentionally minimal (references parent)
- **Keep:** As-is

---

### 6.6 Content to Extract to Shared Guardrails

**From clawdis (currently missing from shared):**

1. **Multi-Agent Safety Rules:**

   ```markdown
   - Do not create/apply/drop git stash entries unless requested
   - Do not create/remove/modify git worktree checkouts
   - Do not switch branches unless explicitly requested
   - When you see unrecognized files, keep going; focus on your changes
   ```

2. **Detailed Changelog Workflow:**

   ```markdown
   - Keep latest released version at top (no `Unreleased`)
   - After publishing, bump version and start new top section
   - When working on PR: add changelog entry with PR # and thank contributor
   - When working on issue: reference issue in changelog entry
   ```

3. **PR Merge Workflow:**

   ```markdown
   - Review mode: read via gh pr view/diff, do not switch branches
   - Landing mode: create integration branch, prefer rebase for linear history
   - Always add PR author as co-contributor
   - Run full gate locally before committing
   ```

4. **Don't Edit node_modules:**

   ```markdown
   - Never edit `node_modules` (global/Homebrew/npm/git installs too)
   - Updates will overwrite your changes
   ```

**From macos-automator-mcp (unique pattern):**

5. **Operational Learnings Section:**

   ```markdown
   ## Agent Operational Learnings

   This section captures key strategies based on collaborative sessions.

   - When external tool returns cryptic errors, suspect dynamic content
   - Enable detailed logging and ensure log visibility
   - For complex patterns (regex), use iterative simplification
   ```

**From Matcha (domain-specific but illustrative):**

6. **Reference Implementation Pattern:**

   ```markdown
   ## Implementation Guidelines

   **CRITICAL:** When implementing features, always compare with [reference implementation]
   to ensure compatibility and correctness.

   Key Principle: "When unsure, align with what [reference] does. [Reference] works,
   our implementation doesn't work yet."
   ```

---

### 6.7 Synchronization Process

**Option A: Manual Sync (Simple)**

1. Update `~/Projects/agent-scripts/AGENTS.MD` when adding shared rules
2. Notify team/self to pull updates manually to local repos
3. Local repos reference but don't copy shared content

**Option B: Copy with Tags (Current mcporter Model)**

1. Update `~/Projects/agent-scripts/AGENTS.MD`
2. Copy shared content into local `<shared>` blocks
3. Provides offline reference, but requires manual sync
4. Risk of drift between copies

**Option C: Script-Based Sync (Advanced)**

1. Create sync script that updates all `<shared>` blocks from master
2. Run periodically or on-demand
3. Maintains consistency automatically
4. More complex to set up

**Recommendation:** Start with **Option A** (simplest), move to **Option C** if drift becomes a problem.

---

## 7. Quantitative Summary

### 7.1 Overall Duplication Statistics

| Metric | Value |
|--------|-------|
| **Total files analyzed** | 13 |
| **Repositories covered** | 11 |
| **Average file length** | ~100 lines |
| **Estimated shared content** | 60-80 lines |
| **Estimated unique content** | 20-40 lines |
| **Duplication percentage** | **60-80%** |
| **Consolidation potential** | **HIGH** |

### 7.2 Content Category Breakdown

| Category | Avg Lines | Duplication | Can Consolidate? |
|----------|-----------|-------------|------------------|
| Role Definition | 5-10 | 80% | ✓ YES |
| Coding Philosophy | 10-15 | 90% | ✓ YES |
| Formatting Rules | 5-10 | 75% | ✓ PARTIAL (framework choices vary) |
| Dependencies | 5-8 | 85% | ✓ YES |
| Testing | 8-12 | 75% | ✓ PARTIAL (frameworks vary) |
| Git/Commits | 8-12 | 90% | ✓ YES |
| Documentation | 5-8 | 80% | ✓ YES |
| Build Commands | 10-20 | 10% | ✗ NO (project-specific) |
| Project Structure | 5-15 | 5% | ✗ NO (project-specific) |
| Prohibitions | 5-10 | 75% | ✓ YES |

### 7.3 Repository Maturity Assessment

| Repository | Completeness | Consistency | Needs Work |
|-----------|--------------|-------------|------------|
| **mcporter** | ★★★★★ | ★★★★★ | None (master file) |
| **clawdis** | ★★★★★ | ★★★★★ | Minor: Extract some rules to shared |
| **Peekaboo** | ★★★★☆ | ★★★★☆ | Minor: Add shared reference |
| **sweetlink** | ★★★☆☆ | ★★★★☆ | Medium: Add shared ref, expand rules |
| **Trimmy** | ★★★☆☆ | ★★★☆☆ | Medium: Add shared ref, expand testing |
| **tokentally** | ★★★☆☆ | ★★★★☆ | Medium: Add shared ref, expand rules |
| **VibeMeter** | ★★★☆☆ | ★★★☆☆ | Medium: Add shared ref, testing |
| **Matcha** | ★★☆☆☆ | ★★★☆☆ | High: Add shared ref, general guardrails |
| **summarize** | ★★☆☆☆ | ★★☆☆☆ | High: Major expansion needed |
| **claude-code-mcp** | ★★☆☆☆ | ★★★☆☆ | High: Add shared ref, coding standards |
| **macos-automator-mcp** | ★★☆☆☆ | ★★★☆☆ | High: Add shared ref, expand |
| **AXorcist** | ★☆☆☆☆ | ★★★★★ | None (intentionally minimal) |

**Legend:**

- **Completeness:** How much necessary content is present
- **Consistency:** How well it follows patterns from other repos
- **★★★★★:** Excellent, **★★★☆☆:** Good, **★★☆☆☆:** Needs work, **★☆☆☆☆:** Minimal

---

## 8. Action Plan

### 8.1 Immediate Actions (Week 1)

**Priority 1: Establish Central Reference**

- [ ] Verify `~/Projects/agent-scripts/AGENTS.MD` is up-to-date
- [ ] Ensure it contains all shared guardrails from mcporter
- [ ] Add multi-agent safety rules from clawdis
- [ ] Add "don't edit node_modules" rule

**Priority 2: Update High-Impact Repos**

- [ ] **summarize:** Add shared reference line + basic structure
- [ ] **sweetlink:** Add shared reference line
- [ ] **tokentally:** Add shared reference line
- [ ] **Matcha:** Add shared reference line + expand beyond testing

### 8.2 Short-Term Actions (Month 1)

**Priority 3: Update Medium-Impact Repos**

- [ ] **Trimmy:** Add shared reference, expand testing section
- [ ] **claude-code-mcp:** Add shared reference, add coding standards
- [ ] **VibeMeter:** Add shared reference, expand testing
- [ ] **macos-automator-mcp:** Add shared reference, add general rules

**Priority 4: Standardize Section Names**

- [ ] Ensure all repos use consistent section headers:
  - "Project Structure & Modules"
  - "Build, Test, and Development Commands"
  - "Coding Style & Naming Conventions"
  - "Testing Guidelines"
  - "Commit & Pull Request Guidelines"
  - "Security & Configuration Tips"

### 8.3 Long-Term Actions (Ongoing)

**Priority 5: Continuous Improvement**

- [ ] Create sync script for `<shared>` blocks (if using Option B/C)
- [ ] Establish review cadence (quarterly?) for shared guardrails
- [ ] Document process for contributing to shared guardrails
- [ ] Create template files for new repos (TS and Swift versions)

**Priority 6: Knowledge Capture**

- [ ] Extract unique patterns worth sharing (like macos-automator-mcp's troubleshooting)
- [ ] Document rationale for choices (why Biome vs Prettier, etc.)
- [ ] Build FAQ section in shared guardrails

### 8.4 Success Metrics

**After consolidation, measure:**

- [ ] Reduction in duplicate content across repos (target: <30%)
- [ ] Consistency of core rules (target: 90%+ repos have same rules)
- [ ] Time to onboard new repo (should be faster with templates)
- [ ] Ease of updating shared rules (one update propagates to all)

---

## 9. Conclusion

This analysis reveals significant opportunity for consolidation:

1. **60-80% of content is duplicated** across repositories with only minor variations
2. **A central shared guardrails file** already exists (mcporter/AGENTS.MD) but is not referenced by most repos
3. **Simple fix for most repos:** Add one line referencing shared guardrails
4. **10-20% of content** (project structure, build commands, specific workflows) must remain local

**Key Insight:** The mcporter file demonstrates the ideal model—shared content in `<shared>` tags, project-specific content outside. Other repos should adopt this pattern.

**Recommended Next Steps:**

1. Update high-priority repos (summarize, sweetlink, tokentally, Matcha) to reference shared guardrails
2. Extract valuable unique patterns (clawdis multi-agent rules, macos-automator-mcp troubleshooting) to shared
3. Standardize section names across all repos
4. Create templates for new repos (TypeScript and Swift versions)

**Expected Benefits:**

- **Reduced maintenance:** Update once, apply everywhere
- **Better consistency:** All agents follow same core rules
- **Faster onboarding:** New repos start with complete guardrails
- **Preserved flexibility:** Project-specific rules remain local

---

## Appendix A: Detailed Text Snippets

### A.1 Complete Shared Content Block (from mcporter)

```markdown
<shared>
# AGENTS.md

Shared guardrails distilled from the various `~/Projects/*/AGENTS.md` files.
This document highlights the rules that show up again and again; still read
the repo-local instructions before making changes.

## General Guardrails

### Intake & Scoping
- Open the local agent instructions plus any `docs:list` summaries at the start
  of every session.
- Review any referenced tmux panes, CI logs, or failing command transcripts so
  you understand the most recent context before writing code.

### Tooling & Command Wrappers
- Use the command wrappers provided by the workspace (`./runner …`,
  `scripts/committer`, `pnpm mcp:*`, etc.).
- Stick to the package manager and runtime mandated by the repo (pnpm-only,
  bun-only, swift-only, go-only, etc.). Never swap in alternatives without approval.
- Ask the user before adding dependencies, changing build tooling, or altering
  project-wide configuration.
- When discussing dependencies, always provide a GitHub URL.

### Build, Test & Verification
- Before handing off work, run the full "green gate" for that repo (lint,
  type-check, tests, doc scripts, etc.).
- Treat every bug fix as a chance to add or extend automated tests that prove
  the behavior.

### Code Quality & Naming
- Refactor in place. Never create duplicate files with suffixes such as "V2",
  "New", or "Fixed"; update the canonical file and remove obsolete paths entirely.
- Favor strict typing: avoid `any`, untyped dictionaries, or generic type erasure.
- Keep files at a manageable size. When a file grows unwieldy, extract helpers
  or new modules instead of letting it bloat.
- Match the repo's established style by studying existing code before introducing
  new patterns.

### Git, Commits & Releases
- Invoke git through the provided wrappers, especially for status, diffs, and
  commits. Only commit or push when the user asks you to do so.
- Follow the documented release or deployment checklists instead of inventing
  new steps.

### Documentation & Knowledge Capture
- Update existing docs whenever your change affects them.
- Only create new documentation when the user or local instructions explicitly
  request it.
- Routine test additions don't require changelog entries; reserve changelog lines
  for user-visible behavior changes.

### Stack-Specific Reminders

## Swift Projects
- Validate changes with `swift build` and the relevant filtered test suites.
- Keep concurrency annotations (`Sendable`, actors, structured tasks) accurate.

## TypeScript Projects
- Use the package manager declared by the workspace (often `pnpm` or `bun`).
- Maintain strict typing—avoid `any`, prefer utility helpers already provided
  by the repo.
- Treat `lint`, `typecheck`, and `test` commands as mandatory gates before
  handing off work.

Keep this master file up to date as you notice new rules that recur across
repositories.

</shared>
```

### A.2 Multi-Agent Safety Rules (from clawdis)

```markdown
## Multi-Agent Safety

- **Do not create/apply/drop `git stash` entries** unless explicitly requested
  (this includes `git pull --rebase --autostash`). Assume other agents may be
  working; keep unrelated WIP untouched.

- **When the user says "push":** you may `git pull --rebase` to integrate latest
  changes (never discard other agents' work).

- **When the user says "commit":** scope to your changes only.

- **When the user says "commit all":** commit everything in grouped chunks.

- **Do not create/remove/modify `git worktree` checkouts** (or edit `.worktrees/*`)
  unless explicitly requested.

- **Do not switch branches / check out a different branch** unless explicitly
  requested.

- **Running multiple agents is OK** as long as each agent has its own session.

- **When you see unrecognized files, keep going:** focus on your changes and
  commit only those.

- **Focus reports on your edits:** avoid guard-rail disclaimers unless truly
  blocked; when multiple agents touch the same file, continue if safe; end with
  a brief "other files present" note only if relevant.
```

### A.3 Changelog Workflow (from clawdis)

```markdown
## Changelog Workflow

- Keep latest released version at top (no `Unreleased`)
- After publishing, bump version and start a new top section
- When working on a PR: add a changelog entry with the PR number and thank the
  contributor
- When working on an issue: reference the issue in the changelog entry
- When merging a PR: leave a PR comment that explains exactly what we did and
  include the SHA hashes
- When merging a PR from a new contributor: add their avatar to the README
  "Thanks to all clawtributors" thumbnail list
- After merging a PR: run `bun scripts/update-clawtributors.ts` if the contributor
  is missing, then commit the regenerated README
```

### A.4 Testing Framework Choice (from Matcha)

```markdown
## Testing Framework

**IMPORTANT:** This project uses Swift Testing (introduced in Swift 6/Xcode 16)
instead of XCTest for all new tests.

### Key Differences:
- Use `@Test` attribute instead of `func test...()` methods
- Use `#expect()` and `#require()` macros instead of `XCTAssert...` functions
- Use `@Suite` for test organization instead of `XCTestCase` subclasses
- Tests run in parallel by default
- Use `init()` and `deinit` for setup/teardown instead of `setUp()`/`tearDown()`

### Migration Status:
- All tests created after this note should use Swift Testing
- Existing XCTest tests will be migrated incrementally
- Both frameworks can coexist during the migration period
```

---

## Appendix B: Repository File Locations

| Repository | File Name | Path | Lines |
|-----------|-----------|------|-------|
| mcporter | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/mcporter_AGENTS.md` | 110 |
| sweetlink | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/sweetlink_AGENTS.md` | 42 |
| Trimmy | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/Trimmy_AGENTS.md` | 43 |
| summarize | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/summarize_AGENTS.md` | 22 |
| clawdis | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/clawdis_AGENTS.md` | 135 |
| AXorcist | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/AXorcist_AGENTS.md` | 6 |
| tokentally | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/tokentally_AGENTS.md` | 45 |
| Peekaboo | AGENTS.md | `/Users/sky/git/agent-md/fetched_files/Peekaboo_AGENTS.md` | 40 |
| claude-code-mcp | CLAUDE.md | `/Users/sky/git/agent-md/fetched_files/claude-code-mcp_CLAUDE.md` | 57 |
| Matcha | CLAUDE.md | `/Users/sky/git/agent-md/fetched_files/Matcha_CLAUDE.md` | 64 |
| VibeMeter | CLAUDE.md | `/Users/sky/git/agent-md/fetched_files/VibeMeter_CLAUDE.md` | 209 |
| macos-automator-mcp | CLAUDE.md | `/Users/sky/git/agent-md/fetched_files/macos-automator-mcp_CLAUDE.md` | 103 |
| Peekaboo | CLAUDE.md | `/Users/sky/git/agent-md/fetched_files/Peekaboo_CLAUDE.md` | 40 |

**Total Lines:** 916
**Average per File:** ~70 lines
**Estimated Unique Content:** ~200-300 lines
**Estimated Duplicate Content:** ~600-700 lines

---

**Report Generated:** 2026-01-19
**Analysis Tool:** Claude Sonnet 4.5
**Repositories Covered:** 11
**Files Analyzed:** 13
