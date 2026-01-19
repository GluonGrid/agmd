# Universal Agent Guardrails

**Version:** 1.0.0
**Managed by:** agmd (Agent Markdown Manager)
**Purpose:** Language-agnostic, universal agent instruction base

This file contains core principles that apply to **all** software projects regardless of language, stack, or domain. Language/stack-specific extensions belong in profiles, and project-specific details belong in local configuration.

---

## How to Use

### In Your Project

Add this to the top of your project's `AGENTS.md` or `.agmd.md`:

```yaml
---
agmd:
  version: 1.0.0
  shared: https://agmd.dev/shared/AGENTS.md  # or ~/agmd/shared/AGENTS.md
  profiles: [typescript, node-cli]  # Your stack
---
```

Then add project-specific content below.

### With agmd CLI

```bash
# Initialize with profiles
agmd init --profile typescript --profile node-cli

# Show effective config (merged from all layers)
agmd show --merged

# Validate your config
agmd validate
```

---

## 1. Intake & Context Understanding

**Before making any changes:**

- **Read project configuration**
  - Read local agent config file (AGENTS.md, .agmd.md, CLAUDE.md)
  - Review project documentation (README, CONTRIBUTING, architecture docs)
  - Check for specialized docs (API.md, DEPLOYMENT.md, etc.)

- **Understand structure**
  - Identify entry points (main files, CLI commands, API routes)
  - Map directory organization
  - Note testing strategy and locations

- **Gather context**
  - Review recent commit history (understand evolution)
  - Check open issues/PRs (understand current focus)
  - Look for CI/CD configuration (understand workflows)

- **Ask when unclear**
  - Clarify ambiguous requirements before proceeding
  - Don't assume - ask for confirmation on architectural decisions
  - Surface trade-offs for user to decide

---

## 2. Code Quality Principles

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

---

## 3. Dependency Management

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
  - If project uses `npm`, continue using `npm`
  - If project uses `bun`, `cargo`, `go mod`, etc. - respect that
  - Swapping package managers requires explicit user approval

- **Respect lockfiles**
  - Don't edit lockfiles manually
  - Use package manager commands to update
  - Commit lockfile changes with dependency updates

- **Exact versions for patches**
  - If a dependency is patched/vendored, use exact version
  - No semver ranges (`^`, `~`) on patched dependencies
  - Patching dependencies requires explicit approval

---

## 4. Testing Philosophy

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

---

## 5. Version Control & Git

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

- **Multi-agent awareness**
  - **Don't create/apply/drop git stash** (unless requested)
  - **Don't switch branches** (unless requested)
  - **Don't modify git worktrees** (unless requested)
  - When you see unrecognized files, note them but continue with your work
  - Commit only YOUR changes when working in parallel

---

## 6. Documentation

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
  - Internal refactoring (unless it affects performance/behavior)
  - Test additions (unless they reveal a user-facing issue)
  - Dependency updates (unless they affect users)
  - Typo fixes in code comments

- **Format consistently**
  - Follow existing changelog format
  - Include issue/PR references when relevant
  - Thank contributors

---

## 7. Security & Secrets

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
  - If you must, explain why and get explicit approval

### Dependency Security

- **Never edit `node_modules` (or equivalent)**
  - Changes will be overwritten on next install
  - Use patch files if you must modify dependencies
  - Consider forking if major changes are needed

- **Don't edit lockfiles manually**
  - Use package manager commands
  - Lockfiles ensure reproducible builds

---

## 8. Build & Verification

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

---

## 9. Communication & Interaction

### Ask Questions

- **When requirements are ambiguous**
  - Ask for clarification
  - Surface trade-offs
  - Propose alternatives

- **When making architectural decisions**
  - Present options with pros/cons
  - Explain implications
  - Let user decide

- **When blocked**
  - Explain what you tried
  - Describe the blocker
  - Ask for guidance

### Provide Context

- **Explain changes**
  - Why did you make this change?
  - What problem does it solve?
  - What trade-offs did you make?

- **Surface risks**
  - Highlight potential breaking changes
  - Note performance implications
  - Mention security considerations

### Be Concise

- **Respect user's time**
  - Get to the point
  - Avoid unnecessary preamble
  - Use formatting to improve readability

- **Show, don't tell**
  - Provide code examples
  - Show diffs
  - Include command outputs

---

## 10. Language-Agnostic Workflows

### Issue Investigation

1. **Reproduce the issue**
   - Understand the expected behavior
   - Identify the actual behavior
   - Create minimal reproduction case

2. **Identify root cause**
   - Use debugging tools
   - Add logging/tracing
   - Narrow down the problem

3. **Fix and verify**
   - Implement fix
   - Add test proving fix works
   - Verify fix doesn't break other functionality

### Feature Implementation

1. **Understand requirements**
   - What problem are we solving?
   - Who is the user?
   - What does success look like?

2. **Design approach**
   - Identify affected files/modules
   - Consider edge cases
   - Plan testing strategy

3. **Implement incrementally**
   - Start with core functionality
   - Add edge case handling
   - Refine based on feedback

4. **Verify completeness**
   - Does it meet requirements?
   - Are there tests?
   - Is it documented?

### Refactoring

1. **Have a clear goal**
   - What are we improving?
   - What problem does current code have?
   - How will we measure success?

2. **Ensure safety net**
   - Tests must pass before starting
   - Tests should still pass after
   - Add tests if coverage is lacking

3. **Refactor incrementally**
   - Small, focused changes
   - Commit frequently
   - Verify tests pass between steps

4. **Clean up**
   - Remove dead code
   - Update documentation
   - Verify no regressions

---

## 11. Advanced: Long-Running Tasks

### Background Task Management

**When to use:**

- Commands that could hang
- Long-running builds
- Development servers
- File watchers

**Recommendations:**

- Use tmux/screen if available (check with `which tmux`)
- Document what you're running
- Clean up sessions when done
- Don't wrap in infinite polling loops

**Example pattern:**

```bash
# Start in background
tmux new-session -d -s myproject-build "npm run build"

# Check status
tmux has-session -t myproject-build 2>/dev/null && echo "Running"

# Attach to view
tmux attach -t myproject-build

# Kill when done
tmux kill-session -t myproject-build
```

---

## 12. Operational Learnings Pattern

**Purpose:** Capture hard-won knowledge and troubleshooting insights

**When to add:**

- After solving a tricky bug
- After discovering a non-obvious solution
- After learning something valuable about the project

**Example:**

```markdown
## Operational Learnings

### Database Connection Timeouts

**Problem:** Intermittent connection timeouts in production
**Root Cause:** Connection pool exhaustion during traffic spikes
**Solution:** Increased pool size + connection timeout
**Lesson:** Always load test with production-like traffic patterns
```

**Benefits:**

- Prevents repeated mistakes
- Speeds up onboarding
- Documents tribal knowledge

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

### Feedback

- Issues: <https://github.com/yourusername/agmd/issues>
- Discussions: <https://github.com/yourusername/agmd/discussions>
- Documentation: <https://agmd.dev>

---

## See Also

- **Language Profiles:** TypeScript, Swift, Go, Python, Rust
- **Stack Profiles:** Node CLI, Web API, macOS App, etc.
- **Templates:** CI/CD, Deployment, Security, etc.
- **agmd Documentation:** <https://agmd.dev/docs>

---

**Version:** 1.0.0
**Last Updated:** 2026-01-19
**License:** MIT
**Maintained by:** agmd community
