# Final Config Format for agmd

**Focus:** How configs actually work in practice

---

## File Naming

### What Agents Actually Read

Agents look for these files (in order):
1. `AGENTS.md` (primary, most common)
2. `CLAUDE.md` (alternative naming)
3. Whatever symlink points to via `agmd symlink`

### What agmd Uses

**agmd only manages `AGENTS.md`** - no `.agmd.md` needed!

```
my-project/
├── AGENTS.md          # ← agmd creates/manages this
└── .agmd/
    └── cache.json     # ← agmd's internal cache
```

**If you want a different name:**
```bash
# Create symlink after agmd init
agmd init --profile typescript
ln -s AGENTS.md CLAUDE.md

# OR use agmd symlink command
agmd symlink CLAUDE.md  # Creates CLAUDE.md → AGENTS.md
```

---

## Final Config Structure

### Example: TypeScript Project with Node CLI Profile

**File: `AGENTS.md`**

```yaml
---
agmd:
  version: 1.0.0
  shared: ~/.agmd/shared/base.md
  profiles:
    - typescript
    - node-cli
  overrides:
    code-quality.file-size-limit: 800
---

# Project: my-api

## Project Structure

### Directory Layout
- `src/` - API implementation
  - `routes/` - Express routes
  - `models/` - Database models
  - `services/` - Business logic
- `tests/` - Vitest tests
- `docs/` - API documentation

## Build Commands

```bash
# Install dependencies
pnpm install

# Development
pnpm dev              # Start dev server (port 3000)

# Production build
pnpm build            # TypeScript compilation

# Testing
pnpm test             # Run all tests
pnpm test:watch       # Watch mode
pnpm test:coverage    # With coverage

# Quality checks
pnpm lint             # Run Biome
pnpm typecheck        # TypeScript only
pnpm check            # Full gate (lint + typecheck + test)
```

## Deployment

- **Platform:** Railway
- **CI:** GitHub Actions (`.github/workflows/deploy.yml`)
- **Staging:** Push to `develop` branch
- **Production:** Push to `main` branch
- **Environment:** Set `DATABASE_URL`, `JWT_SECRET` in Railway dashboard

## Special Rules

- Database migrations must be backward-compatible
- API versioning via `/v1/`, `/v2/` routes
- All endpoints require authentication except `/health`
- Rate limiting: 100 req/min per IP
```

---

## How Profiles Are Applied

### Resolution Flow

When agent reads `AGENTS.md`:

1. **agmd reads frontmatter:**
   ```yaml
   shared: ~/.agmd/shared/base.md
   profiles: [typescript, node-cli]
   ```

2. **agmd resolves inheritance:**
   ```
   Load: ~/.agmd/shared/base.md (Universal)
       → Merge: ~/.agmd/profiles/typescript.md
       → Merge: ~/.agmd/profiles/node-cli.md
       → Merge: AGENTS.md (project content)
       → Apply: overrides from frontmatter
   ```

3. **Agent sees effective config:**
   - All universal rules (9 core sections)
   - All TypeScript-specific rules
   - All Node CLI-specific rules
   - Project-specific details
   - Any overrides applied

---

## Profile Content Examples

### TypeScript Profile (`~/.agmd/profiles/typescript.md`)

```yaml
---
agmd:
  version: 1.0.0
  type: profile
  extends: universal
---

# TypeScript Profile

## Language Specifics

### Type System
- Use strict mode (`"strict": true` in tsconfig.json)
- Avoid `any` - use `unknown` and type guards instead
- Prefer `interface` for objects, `type` for unions/intersections
- Enable `noUncheckedIndexedAccess` for safer array access

### Module System
- ESM only (`"type": "module"` in package.json)
- Use `.js` extensions in imports (ESM requirement)
- Prefer named exports over default exports

### Common Patterns
- Use utility types: `Partial<T>`, `Pick<T, K>`, `Omit<T, K>`
- Async/await over raw Promises
- Error handling with explicit types, not `any`

## Common Tools

### Package Manager
Project declares which to use (respect it):
- `pnpm` (most common)
- `bun` (faster alternative)
- `npm` (fallback)

### Formatting & Linting
Prefer Biome (all-in-one):
- `biome check` - Lint + format
- `biome check --fix` - Auto-fix

Alternative:
- Prettier + ESLint/oxlint

### Testing
Vitest (preferred):
- Fast, Vite-powered
- Compatible API with Jest
- Built-in coverage (V8)

Alternative:
- Jest (older projects)

### Build
- `tsc` - Type checking + emit
- `esbuild` - Fast bundling
- `tsup` - Simple bundler
- `vite` - For apps

## File Conventions

```
src/
├── index.ts          # Entry point
├── types.ts          # Shared types
├── utils/            # Utilities
│   ├── index.ts
│   └── helpers.ts
└── features/         # Feature-based
    └── auth/
        ├── auth.ts
        └── auth.test.ts

tests/
├── fixtures/         # Test fixtures
├── setup.ts          # Test setup
└── e2e/              # E2E tests
    └── api.e2e.test.ts
```

## Build Pattern (Template)

Most TypeScript projects follow this pattern:

```bash
# Install
pnpm install

# Develop
pnpm dev

# Build
pnpm build          # Usually `tsc` or bundler

# Test
pnpm test           # Usually Vitest/Jest
pnpm test:coverage

# Lint
pnpm lint           # Usually Biome or ESLint
pnpm format         # Auto-fix formatting

# Type check only
pnpm typecheck      # tsc --noEmit

# Full gate
pnpm check          # lint + typecheck + test
```

**Note:** Actual commands are in project's `AGENTS.md`, not here.
This is just the typical pattern.
```

### Node CLI Profile (`~/.agmd/profiles/node-cli.md`)

```yaml
---
agmd:
  version: 1.0.0
  type: profile
  extends: typescript
---

# Node CLI Tools Profile

Extends: TypeScript

## CLI Specifics

### Entry Point
- `src/cli.ts` or `src/index.ts`
- Shebang: `#!/usr/bin/env node`
- Make executable: `chmod +x`

### Argument Parsing
Common libraries:
- `commander` - Full-featured, popular
- `yargs` - Alternative
- `minimist` - Minimal

### CLI Conventions
- Provide `--help` and `--version`
- Use `--verbose` / `-v` for debug output
- Exit codes: `0` (success), `1` (error), `2` (usage error)
- Errors to stderr, output to stdout

### Signal Handling
```typescript
// Handle SIGINT (Ctrl+C) gracefully
process.on('SIGINT', () => {
  console.log('\nGracefully shutting down...');
  // Cleanup here
  process.exit(0);
});
```

### Environment Variables
- Prefix with app name: `MYAPP_DEBUG`, `MYAPP_CONFIG`
- Use `dotenv` for `.env` file support
- Document all env vars in README

## Distribution

### Package.json
```json
{
  "bin": {
    "mycli": "./dist/cli.js"
  },
  "files": [
    "dist",
    "README.md",
    "LICENSE"
  ]
}
```

### Publishing
- `npm publish` (if public)
- `pnpm pack` (test package first)
- Consider standalone binaries (`pkg`, `bun compile`)

## Testing CLI

### Test Strategies
- Unit tests for logic
- Integration tests for commands
- Use `execa` to test CLI execution

```typescript
import { execa } from 'execa';

test('cli --version', async () => {
  const { stdout } = await execa('mycli', ['--version']);
  expect(stdout).toMatch(/\d+\.\d+\.\d+/);
});
```

## Build Pattern (Template)

```bash
# Development
pnpm dev              # Usually tsx watch

# Build
pnpm build            # tsc or bundler

# Test CLI locally
pnpm link             # Link globally
mycli --help          # Test command

# Publish
pnpm version patch    # Bump version
pnpm publish          # Publish to npm
```
```

---

## How `agmd show --merged` Works

### Command

```bash
$ cd my-project
$ agmd show --merged
```

### Output

```markdown
# Effective Configuration for my-api

## Inheritance Chain
✓ Universal: ~/.agmd/shared/base.md
✓ Profile: typescript
✓ Profile: node-cli
✓ Project: AGENTS.md
✓ Overrides: 1 applied

---

# 1. Intake & Context Understanding
[Content from universal base.md]

# 2. Code Quality Principles
[Content from universal base.md]
[Override applied: file-size-limit = 800]

# 3. Dependency Management
[Content from universal base.md]

# 4. Testing Philosophy
[Content from universal base.md]

# 5. Version Control & Git
[Content from universal base.md]

# 6. Documentation
[Content from universal base.md]

# 7. Security & Secrets
[Content from universal base.md]

# 8. Build & Verification
[Content from universal base.md]

# 9. Multi-Agent Safety
[Content from universal base.md]

---

# Language Specifics (from typescript profile)

## Type System
- Use strict mode
- Avoid `any`
[etc.]

---

# CLI Specifics (from node-cli profile)

## Entry Point
- src/cli.ts with shebang
[etc.]

---

# Project: my-api (from AGENTS.md)

## Project Structure
- `src/` - API implementation
[etc.]

## Build Commands
```bash
pnpm install
pnpm dev
[etc.]
```

## Deployment
- Platform: Railway
[etc.]
```

---

## agmd Commands in Practice

### Initialize New Project

```bash
$ cd my-new-api
$ agmd init --profile typescript --profile node-cli

✓ Created AGENTS.md
✓ Added universal shared: ~/.agmd/shared/base.md
✓ Added profile: typescript
✓ Added profile: node-cli

Next steps:
1. Edit AGENTS.md to add:
   - Project structure
   - Build commands
   - Deployment details
2. Run 'agmd show --merged' to see effective config
3. Run 'agmd validate' to check for issues
```

**Creates `AGENTS.md`:**

```yaml
---
agmd:
  version: 1.0.0
  shared: ~/.agmd/shared/base.md
  profiles:
    - typescript
    - node-cli
---

# Project: my-new-api

## Project Structure

[TODO: Add your project structure]

## Build Commands

[TODO: Add your build commands]
```

### Show Current Config

```bash
$ agmd show

# Shows AGENTS.md as-is (with frontmatter)
```

### Show Merged Config

```bash
$ agmd show --merged

# Shows fully resolved config from all layers
# (universal + profiles + project + overrides)
```

### Show in JSON (for tools)

```bash
$ agmd show --merged --json > effective-config.json
```

### Validate Config

```bash
$ agmd validate

✓ YAML frontmatter is valid
✓ Universal shared config found: ~/.agmd/shared/base.md
✓ Profile 'typescript' found
✓ Profile 'node-cli' found
✓ Override 'code-quality.file-size-limit' applied
! Warning: Consider adding 'deployment' section

Config is valid
```

---

## Symlink Support

### Creating Alternative Names

If you want agents to read `CLAUDE.md` instead:

```bash
# After agmd init
$ agmd symlink CLAUDE.md

Created symlink: CLAUDE.md → AGENTS.md
Agents can now read either file (same content)
```

### How It Works

```bash
$ ls -la
-rw-r--r--  AGENTS.md      # Real file (managed by agmd)
lrwxr-xr-x  CLAUDE.md → AGENTS.md  # Symlink
```

Both point to the same content. Update `AGENTS.md` and `CLAUDE.md` sees the change automatically.

---

## Override Examples

### Override Single Value

```yaml
---
agmd:
  overrides:
    code-quality.file-size-limit: 1000
---
```

### Override Multiple Values

```yaml
---
agmd:
  overrides:
    code-quality.file-size-limit: 800
    testing.coverage-threshold: 90
    testing.framework: jest  # Use Jest instead of Vitest
---
```

### How Overrides Work

**Without override:**
```markdown
## Code Quality (from universal)
- Keep files under 500 LOC
```

**With override:**
```yaml
overrides:
  code-quality.file-size-limit: 800
```

**Effective config:**
```markdown
## Code Quality
- Keep files under 800 LOC
```

---

## Directory Structure Reference

### User's Home Directory

```
~/.agmd/
├── shared/
│   └── base.md              # Universal shared (9 sections)
├── profiles/
│   ├── typescript.md
│   ├── swift.md
│   ├── go.md
│   ├── python.md
│   ├── rust.md
│   ├── node-cli.md          # Extends typescript
│   ├── web-api.md           # Extends typescript
│   └── custom/
│       └── my-custom.md     # User-created profiles
├── config.yml               # agmd settings
└── cache/                   # Cached resolved configs
    └── project-hash.json
```

### Project Directory

```
my-project/
├── AGENTS.md                # ← Main config (agmd manages this)
├── CLAUDE.md                # ← Optional symlink to AGENTS.md
├── .agmd/
│   └── cache.json           # Resolved config cache
├── src/
├── tests/
└── package.json
```

---

## Best Practices

### 1. Keep Project Config Minimal

**Good (minimal, project-specific):**
```yaml
---
agmd:
  shared: ~/.agmd/shared/base.md
  profiles: [typescript, node-cli]
---

# Project: my-api

## Project Structure
- `src/` - Source code
- `tests/` - Tests

## Build Commands
```bash
pnpm install
pnpm dev
pnpm build
pnpm test
```
```

**Bad (duplicating universal/profile content):**
```markdown
# Don't repeat universal rules here!

## Code Quality
- Refactor in place [← already in universal]
- Strict typing [← already in typescript profile]
[etc...]
```

### 2. Use Profiles for Stack-Specific Rules

If you have multiple projects with same stack, create a profile:

```bash
# Create custom profile
$ mkdir -p ~/.agmd/profiles/custom
$ cat > ~/.agmd/profiles/custom/my-stack.md
---
agmd:
  version: 1.0.0
  type: profile
  extends: typescript
---

# My Stack Profile

## Framework
- Next.js 14
- Tailwind CSS
- Prisma ORM

## Conventions
- Page components in `app/`
- API routes in `app/api/`
[etc...]
```

Then use it:
```yaml
profiles:
  - typescript
  - my-stack
```

### 3. Use Overrides Sparingly

Only override when project truly needs different value:

```yaml
# Good reason to override
overrides:
  code-quality.file-size-limit: 1200  # Large config file is OK

# Bad reason
overrides:
  testing.framework: jest  # Just use profile instead
```

---

## Summary

### File Naming
- ✅ `AGENTS.md` (primary, agmd manages this)
- ✅ `CLAUDE.md` (symlink via `agmd symlink`)
- ❌ `.agmd.md` (not needed, agents don't read hidden files)

### Config Structure
```yaml
---
agmd:
  shared: ~/.agmd/shared/base.md
  profiles: [typescript, node-cli]
  overrides:
    section.key: value
---

# Project content (structure, commands, special rules)
```

### Inheritance Flow
```
Universal (9 sections)
  → Profile 1 (language-specific)
  → Profile 2 (stack-specific)
  → Project (unique details)
  → Overrides (exceptional cases)
  → Effective Config (what agent sees)
```

### Commands
```bash
agmd init --profile typescript    # Create AGENTS.md
agmd show                         # Show raw
agmd show --merged                # Show resolved
agmd validate                     # Check validity
agmd symlink CLAUDE.md            # Create symlink
```

**Result:** Simple, clean, DRY agent configs that work with existing agent workflows.
