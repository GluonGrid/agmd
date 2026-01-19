# agent-config for M4 MacBook Pro - Quick Setup

This guide is specifically for setting up agent-config on your M4 MacBook Pro (ARM64).

## TL;DR - Super Quick Setup

```bash
# 1. Run the automated setup script
cd agent-config
chmod +x setup-m4.sh
./setup-m4.sh

# 2. Restart your terminal or run:
source ~/.zshrc

# 3. Test it
agent-config --help

# 4. Use it in your projects
cd ~/projects/your-project
agent-config init --claude
agent-config add "Never use fixed package versions; use 'uv add pkg' or 'cargo add pkg' instead"
```

**That's it!** The script handles everything: Go installation, dependencies, building, and global installation.

## What the Setup Script Does

1. ✅ Checks/installs Homebrew
2. ✅ Installs Go (ARM64 version for M4)
3. ✅ Configures Go environment in `~/.zshrc`
4. ✅ Downloads project dependencies (`go mod download`)
5. ✅ Cleans up dependencies (`go mod tidy`)
6. ✅ Builds the binary for ARM64
7. ✅ Installs globally to `$GOPATH/bin`
8. ✅ Optionally installs dev tools (gopls, golangci-lint, delve)

## Manual Setup (If You Prefer)

### Step 1: Install Go

```bash
# Using Homebrew (recommended for M4 Mac)
brew install go

# Verify installation
go version
# Should show: go version go1.21.x darwin/arm64
```

### Step 2: Configure Go Environment

Add to `~/.zshrc`:

```bash
# Go environment
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
```

Then reload:
```bash
source ~/.zshrc
```

### Step 3: Build agent-config

```bash
cd agent-config

# Download dependencies (like 'uv sync' or 'cargo build')
go mod download

# Clean up dependencies
go mod tidy

# Build for M4 (ARM64)
go build -o agent-config .

# Test it
./agent-config --help
```

### Step 4: Install Globally

```bash
# Install to $GOPATH/bin
go install .

# Verify
which agent-config
agent-config --help
```

## Go Package Management for M4 Mac

### The Golden Rule: Never Edit go.mod Manually

Just like you use `uv add` and `cargo add`, in Go you use `go get`:

```bash
# Add a dependency (like 'uv add requests' or 'cargo add serde')
go get github.com/spf13/cobra@latest

# Remove a dependency (like 'uv remove' or 'cargo remove')
# Step 1: Remove the import from your code
# Step 2: Run:
go mod tidy

# Update all dependencies (like 'uv sync --upgrade' or 'cargo update')
go get -u ./...

# Clean up (like 'uv sync' or after 'cargo update')
go mod tidy
```

### Quick Command Reference

| Task | Python (uv) | Rust (cargo) | **Go (M4 Mac)** |
|------|-------------|--------------|-----------------|
| Add package | `uv add pkg` | `cargo add pkg` | **`go get pkg@latest`** |
| Remove package | `uv remove pkg` | `cargo remove pkg` | **Remove import + `go mod tidy`** |
| Update all | `uv sync` | `cargo update` | **`go get -u ./...`** |
| Build | `uv run` | `cargo build` | **`go build .`** |
| Run | `uv run main.py` | `cargo run` | **`go run .`** |

See [GO_PACKAGE_MANAGEMENT.md](../GO_PACKAGE_MANAGEMENT.md) for complete guide.

## M4-Specific Build Commands

```bash
# Build for M4 Mac (default)
go build -o agent-config .

# Build optimized (smaller binary)
go build -ldflags="-s -w" -o agent-config .

# Cross-compile for other platforms
GOOS=linux GOARCH=amd64 go build -o agent-config-linux .
GOOS=windows GOARCH=amd64 go build -o agent-config.exe .

# Build all platforms (uses Makefile)
make release
```

## Development Workflow on M4 Mac

```bash
# 1. Make changes to code
vim cmd/init.go

# 2. Format code (Go's built-in formatter)
go fmt ./...

# 3. Build and test
go build -o agent-config .
./agent-config init --claude

# 4. Run tests (if we add them)
go test ./...

# 5. Install updated version
go install .
```

## Common Issues on M4 Mac

### Issue 1: "go: command not found"

**Solution:**
```bash
# Check if Go is installed
which go

# If not, install it
brew install go

# Add to PATH in ~/.zshrc
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

# Reload shell
source ~/.zshrc
```

### Issue 2: Wrong Architecture

**Verify you're building for ARM64:**
```bash
go env GOARCH
# Should output: arm64

# If not, force ARM64:
GOARCH=arm64 go build -o agent-config .
```

### Issue 3: "Cannot find module"

**Solution:**
```bash
# Download dependencies
go mod download

# Clean up
go mod tidy

# Rebuild
go build -o agent-config .
```

## Using agent-config on M4 Mac

### Example 1: Python Project with uv

```bash
cd ~/projects/my-python-project

# Initialize with Claude
agent-config init --claude

# Add Python-specific rules
agent-config add "Use uv for Python dependency management"
agent-config add --section "Code Style" "Follow PEP 8, use Black formatter (line length 100)"
agent-config add --section "Commands" "uv run pytest - Run tests"
agent-config add --section "Commands" "uv run mypy . - Type check"

# View rules
agent-config list

# Commit to git
git add agent.md CLAUDE.md
git commit -m "Add AI agent configuration"
```

### Example 2: Rust Project with cargo

```bash
cd ~/projects/my-rust-project

# Initialize with all AI tools
agent-config init --all

# Add Rust-specific rules
agent-config add "Use cargo for Rust dependency management"
agent-config add --section "Code Style" "Follow Rust API guidelines"
agent-config add --section "Commands" "cargo test - Run tests"
agent-config add --section "Commands" "cargo clippy -- -D warnings - Lint"

# List rules
agent-config list
```

### Example 3: Multiple Projects Setup

```bash
# Project 1: Python/Django
cd ~/projects/datapred
agent-config init --claude --cursor
agent-config add "Django 5.0 project with PostgreSQL"
agent-config add "Use uv for dependencies"

# Project 2: Rust CLI
cd ~/projects/my-cli-tool
agent-config init --claude
agent-config add "Rust CLI using clap"
agent-config add "Use cargo for dependencies"

# Project 3: Go project
cd ~/projects/my-go-service
agent-config init --claude
agent-config add "Go microservice using Gin"
agent-config add "Use 'go get' for dependencies, never edit go.mod manually"
```

## Verifying Installation

After setup, verify everything works:

```bash
# Check Go installation
go version
# Expected: go version go1.21.x darwin/arm64

# Check Go environment
go env GOARCH
# Expected: arm64

# Check agent-config installation
which agent-config
# Expected: /Users/yourname/go/bin/agent-config

# Test agent-config
agent-config --help
# Should show help text

# Test in a project
cd /tmp
mkdir test-project && cd test-project
agent-config init --claude
ls -la
# Should show: agent.md and CLAUDE.md (symlink)
cat CLAUDE.md
# Should show content from agent.md
```

## Next Steps

1. **Read the full docs**: [README.md](README.md)
2. **Quick tutorial**: [QUICKSTART.md](QUICKSTART.md)
3. **Go package management**: [GO_PACKAGE_MANAGEMENT.md](../GO_PACKAGE_MANAGEMENT.md)
4. **Project plan**: [PROJECT_PLAN.md](../PROJECT_PLAN.md)

## Performance on M4 Mac

Expected performance metrics:

- **Build time**: <2 seconds
- **Binary size**: ~8MB (optimized: ~6MB)
- **Init command**: <100ms
- **Add command**: <50ms
- **List command**: <30ms

The M4's ARM64 architecture provides excellent performance for Go applications!

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## Support

- **Issues**: GitHub Issues
- **Documentation**: See README.md
- **Go Questions**: See GO_PACKAGE_MANAGEMENT.md

---

**Enjoy using agent-config on your M4 Mac!** 🚀
