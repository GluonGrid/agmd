# Universal Agent Guardrails

**Version:** 1.0.0
**Managed by:** agmd
**License:** MIT

This file contains language-agnostic principles for AI agents working on software projects.

---

<!-- agmd:protected-structure-start -->
<!--
  ⚠️ AGENTS: READ THIS BEFORE MODIFYING

  This file is managed by agmd (Agent Markdown Manager).

  ✅ YOU MAY:
  - Add examples and clarifications within sections
  - Add subsections (### or ####) under existing ## sections
  - Fix typos and improve clarity
  - Add new sections AFTER "## 9. Multi-Agent Safety"

  ❌ YOU MUST NOT:
  - Change section titles (## 1. through ## 9. must stay exactly as-is)
  - Reorder sections 1-9
  - Remove marker comments like this one
  - Add content before "## 1. Intake"

  WHY: agmd merges configs by matching section titles. Changes break inheritance.

  🎯 FOR PROJECT-SPECIFIC RULES:
  Add them to your project's AGENTS.md instead, not here.

  📖 FULL DOCS: See AGENT_MODIFICATION_RULES.md
-->
<!-- agmd:protected-structure-end -->

---

## 1. Intake & Context Understanding

<!-- agmd:extensible-section -->
**Before making any changes:**

- **Read project configuration**
  - Read local agent config file (AGENTS.md, CLAUDE.md)
  - Review project documentation (README, CONTRIBUTING, etc.)
  - Check for specialized docs (ARCHITECTURE.md, API.md, DEPLOYMENT.md)

- **Understand structure**
  - Identify entry points (main files, CLI commands, API routes)
  - Map directory organization
  - Note testing strategy and locations

- **Gather context**
  - Review recent commit history
  - Check open issues/PRs
  - Look for CI/CD configuration

- **Ask when unclear**
  - Clarify ambiguous requirements before proceeding
  - Don't assume - ask for confirmation on architectural decisions
  - Surface trade-offs for user to decide

<!-- agmd:extensible-section-end -->

---

## 2. Code Quality Principles

<!-- agmd:extensible-section -->

### Refactoring Philosophy

- **Refactor in place - never create duplicate files**
  - ❌ NEVER: `utils-v2.js`, `helpers-new.ts`, `api-fixed.go`, `backup-service.swift`
  - ✅ INSTEAD: Update the canonical file and remove obsolete code entirely
  - If major restructuring is needed, do it incrementally in the same file
  - Delete old code, don't comment it out (git preserves history)

- **Keep files at manageable size**
  - Guideline: ~500-700 lines of code (not a hard limit)
  - When a file grows too large, extract cohesive modules
  - Prefer many small, focused files over few large ones
  - Extract helpers, utilities, or new modules proactively

- **Match existing code style**
  - Study the project's patterns before adding new ones
  - Follow established naming conventions
  - Maintain consistency in structure and organization
  - When in doubt, ask rather than introducing new patterns

### Code Clarity

- **Write self-documenting code**
  - Use descriptive names for variables, functions, and types
  - Prefer clarity over cleverness
  - Complex logic deserves comments explaining "why", not "what"

- **Avoid premature abstraction**
  - Don't create abstractions until patterns emerge
  - Three instances of duplication before abstracting
  - Keep it simple - add complexity only when needed

- **Type safety (when applicable)**
  - Use strong typing when the language supports it
  - Avoid dynamic types (`any`, `interface{}`, `object`) unless necessary
  - Leverage type systems to catch errors early

<!-- agmd:extensible-section-end -->

---

## 3. Dependency Management

<!-- agmd:extensible-section -->

### Adding Dependencies

- **Always ask before adding dependencies**
  - Provide rationale: What problem does this solve?
  - Explain why existing code can't solve it
  - Provide GitHub URL (or equivalent) for review
  - Wait for explicit approval

- **Evaluate carefully**
  - Consider maintenance burden (is it actively maintained?)
  - Check license compatibility
  - Assess bundle size impact (for frontend)
  - Look for security issues

- **Prefer established solutions**
  - Well-maintained > cutting edge
  - Community-vetted > new and shiny
  - Standard library > third-party when possible

### Package Manager Discipline

- **Never swap package managers**
  - If project uses `pnpm`, continue using `pnpm`
  - If project uses `npm`, `bun`, `cargo`, `go mod`, etc. - respect that
  - Swapping requires explicit user approval

- **Respect lockfiles**
  - Don't edit lockfiles manually
  - Use package manager commands to update
  - Commit lockfile changes with dependency updates

- **Exact versions for patches**
  - If a dependency is patched/vendored, use exact version
  - No semver ranges (`^`, `~`) on patched dependencies
  - Patching dependencies requires explicit approval

<!-- agmd:extensible-section-end -->

---

## 4. Testing Philosophy

<!-- agmd:extensible-section -->

### When to Add Tests

- **Always add/extend tests for bug fixes**
  - Tests prove the fix works
  - Tests prevent regression
  - If you can't write a test, document why

- **Add tests for new features**
  - Test happy path
  - Test edge cases
  - Test error handling

- **Don't obsess over coverage numbers**
  - Aim for meaningful tests, not 100% coverage
  - Focus on critical paths and business logic
  - Test public APIs, not implementation details

### Testing Best Practices

- **Write deterministic tests**
  - Avoid flaky tests (time-based, network-dependent)
  - Use fixtures and mocks for external dependencies
  - Seed random generators for reproducibility

- **Keep tests fast**
  - Unit tests should run in milliseconds
  - Integration tests should run in seconds
  - If tests are slow, fix them - slow tests don't get run

- **Run full test suite before handoff**
  - ALWAYS run tests before marking work complete
  - Fix or explain any failures
  - Don't hand off broken tests

<!-- agmd:extensible-section-end -->

---

## 5. Version Control & Git

<!-- agmd:extensible-section -->

### Commit Discipline

- **Only commit when explicitly asked**
  - Never auto-commit without user request
  - Ask for confirmation before committing
  - Explain what you're committing and why

- **Use conventional commit format**
  ```
  type(scope): summary

  Optional longer description

  Co-authored-by: [Agent Name] <agent@email.com>
  ```

  **Types:**
  - `feat`: New feature
  - `fix`: Bug fix
  - `chore`: Maintenance (deps, config, etc.)
  - `docs`: Documentation changes
  - `test`: Test additions/changes
  - `refactor`: Code restructuring (no behavior change)
  - `perf`: Performance improvements
  - `style`: Formatting, whitespace (no logic change)
  - `ci`: CI/CD changes
  - `build`: Build system changes

- **One logical change per commit**
  - Keep commits focused and atomic
  - Don't mix unrelated changes
  - Group related changes together

### Git Safety

- **Check status before destructive operations**
  - Use wrapper scripts if provided (`scripts/committer`, `./git`, etc.)
  - Never delete unfamiliar files without checking
  - Don't force-push without explicit permission

- **Follow documented workflows**
  - Check for CONTRIBUTING.md or similar
  - Follow release checklists if they exist
  - Don't invent new git workflows

<!-- agmd:extensible-section-end -->

---

## 6. Documentation

<!-- agmd:extensible-section -->

### When to Update Documentation

- **Update docs when behavior changes**
  - Keep README accurate
  - Update API documentation
  - Update configuration examples
  - Update diagrams if they exist

- **Only create new docs when requested**
  - Don't create redundant documentation
  - Edit existing docs rather than creating new ones
  - Ask before creating new doc files

### Changelog Guidelines

- **Add entries for user-facing changes**
  - New features
  - Bug fixes
  - Breaking changes
  - Deprecations

- **Don't add entries for**
  - Internal refactoring (unless affects performance/behavior)
  - Test additions (unless they reveal a user-facing issue)
  - Dependency updates (unless they affect users)
  - Typo fixes in code comments

- **Format consistently**
  - Follow existing changelog format
  - Include issue/PR references when relevant
  - Thank contributors

<!-- agmd:extensible-section-end -->

---

## 7. Security & Secrets

<!-- agmd:extensible-section -->

### Never Commit Secrets

- **Use environment variables**
  - API keys, tokens, passwords → env vars
  - Provide `.env.example` with dummy values
  - Document required environment variables

- **Use secure storage**
  - Keychain/credential managers for local development
  - Secrets management services for production
  - Never hardcode credentials

- **Check before committing**
  - Review diffs for accidental secret exposure
  - Use `.gitignore` for sensitive files
  - Consider git pre-commit hooks

### Configuration Safety

- **Ask before changing project-wide config**
  - Changes to `.gitignore`, `tsconfig.json`, build configs, etc. need approval
  - These affect all contributors and CI systems
  - Explain rationale before making changes

- **Don't disable security features**
  - Don't bypass CORS, CSP, or other security measures
  - Don't disable SSL verification
  - Don't weaken authentication/authorization

### Dependency Security

- **Never edit node_modules (or vendor directories)**
  - Changes will be overwritten on next install
  - Applies to: `node_modules/`, `vendor/`, `Pods/`, etc.
  - Use patch files if you must modify dependencies

- **Don't edit lockfiles manually**
  - Use package manager commands
  - Lockfiles ensure reproducible builds

<!-- agmd:extensible-section-end -->

---

## 8. Build & Verification

<!-- agmd:extensible-section -->

### Pre-Handoff Gate

Before completing work, run the full verification pipeline:

1. **Format** - Code formatting passes
2. **Lint** - Linting passes (no errors, minimal warnings)
3. **Type-check** - No type errors (if applicable)
4. **Test** - All tests pass
5. **Build** - Production build succeeds
6. **Docs** - Documentation generation succeeds (if applicable)

**The order matters:** Format → Lint → Type → Test → Build

### Surface Failures Clearly

- **Include exact command output**
  - Copy full error messages
  - Include stack traces
  - Show the command that failed

- **Don't hide or minimize errors**
  - Be explicit about what failed
  - Explain what you tried
  - Ask for guidance if you're stuck

- **Don't stop development watchers**
  - Keep dev servers running
  - Don't kill background processes
  - Let user manage their development environment

<!-- agmd:extensible-section-end -->

---

## 9. Multi-Agent Safety

<!-- agmd:extensible-section -->

When multiple AI agents may work in the same repository:

- **Do not create/apply/drop git stash**
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

<!-- agmd:extensible-section-end -->

<!-- agmd:end-core-sections -->
<!--
  Core sections end here.
  Agents MAY add new sections below this marker.
  Language/stack-specific content belongs in profiles, not here.
-->

---

## Communication & Interaction

**Ask Questions**

- When requirements are ambiguous
- When making architectural decisions
- When blocked or uncertain

**Provide Context**

- Explain changes (why, what problem it solves, trade-offs)
- Surface risks (breaking changes, performance, security)

**Be Concise**

- Respect user's time
- Get to the point
- Use formatting for readability

---

## Maintenance

### Keeping This File Updated

- This file is maintained as part of the `agmd` project
- Contributions welcome via pull request
- Changes should be language-agnostic and universally applicable
- Language-specific content belongs in profiles

### Version Compatibility

- **Breaking changes** increment major version (2.0.0)
- **New sections** increment minor version (1.1.0)
- **Clarifications/fixes** increment patch version (1.0.1)

---

**Version:** 1.0.0
**Last Updated:** 2026-01-19
**License:** MIT
**Maintained by:** agmd community
**Issues:** https://github.com/yourusername/agmd/issues
