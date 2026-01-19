# agent-config Implementation Project Plan

## Development Environment

### Current Setup
- **Machine**: MacBook Pro with M4 chip (ARM64 architecture)
- **OS**: macOS (ARM64/Apple Silicon)
- **Target**: Cross-platform CLI tool in Go

### Prerequisites Installation

#### 1. Install Homebrew (if not installed)
```bash
# Check if Homebrew is installed
which brew

# If not installed, install it:
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 2. Install Go

```bash
# Install Go via Homebrew (recommended for M4 Mac)
brew install go

# Verify installation
go version
# Should show: go version go1.21.x darwin/arm64

# Check Go environment
go env
# Important values:
# GOARCH="arm64"
# GOOS="darwin"
# GOPATH="/Users/yourname/go"
```

**Expected Output:**
```
go version go1.21.5 darwin/arm64
```

#### 3. Configure Go Environment

Add to your `~/.zshrc` (or `~/.bashrc`):

```bash
# Go environment
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# Reload shell
source ~/.zshrc
```

## Go Package Management - Best Practices

### The Go Way (Similar to `cargo add` / `uv add`)

In Go, dependencies are managed through `go.mod` file (similar to `Cargo.toml` or `pyproject.toml`).

**IMPORTANT RULES (Same Philosophy as Rust/Python):**

✅ **DO THIS:**
```bash
# Add a new dependency (like 'cargo add' or 'uv add')
go get github.com/spf13/cobra@latest

# Add specific version
go get github.com/spf13/cobra@v1.8.0

# Add multiple packages
go get github.com/spf13/cobra@latest github.com/fatih/color@latest

# Clean up unused dependencies
go mod tidy
```

❌ **DON'T DO THIS:**
```bash
# DON'T manually edit go.mod with version numbers
# DON'T write dependencies by hand in go.mod
```

### Go Module Commands Reference

```bash
# Initialize a new Go module (like 'cargo init' or 'uv init')
go mod init github.com/username/projectname

# Download dependencies (like 'cargo build' downloads)
go mod download

# Add/update dependencies
go get package@version

# Remove unused dependencies and add missing ones
go mod tidy

# Verify dependencies
go mod verify

# Show dependency graph
go mod graph

# Update all dependencies to latest
go get -u ./...

# Update specific dependency
go get -u github.com/spf13/cobra@latest
```

### Comparison with Other Languages

| Task | Python (uv) | Rust (cargo) | Go |
|------|-------------|--------------|-----|
| Init project | `uv init` | `cargo init` | `go mod init` |
| Add package | `uv add requests` | `cargo add serde` | `go get package` |
| Remove package | `uv remove requests` | `cargo remove serde` | Remove import, `go mod tidy` |
| Update deps | `uv sync` | `cargo update` | `go get -u ./...` |
| Clean deps | `uv sync` | - | `go mod tidy` |
| Build | `uv run` | `cargo build` | `go build` |
| Run | `uv run main.py` | `cargo run` | `go run .` |

## Project Implementation Plan

### Phase 1: Setup (30 minutes)

#### Step 1.1: Install Go and Tools
```bash
# Install Go
brew install go

# Install useful Go tools
go install golang.org/x/tools/gopls@latest      # Language server
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linter
go install github.com/go-delve/delve/cmd/dlv@latest  # Debugger
```

#### Step 1.2: Create Project Structure
```bash
# Create project directory
mkdir -p ~/projects/agent-config
cd ~/projects/agent-config

# Initialize Go module
go mod init github.com/yourusername/agent-config

# Create directory structure
mkdir -p cmd internal/{config,parser,symlink}
```

#### Step 1.3: Add Dependencies
```bash
# Add Cobra (CLI framework)
go get github.com/spf13/cobra@latest

# Add color output library
go get github.com/fatih/color@latest

# Clean up (equivalent to cargo update or uv sync)
go mod tidy

# Verify go.mod was created correctly
cat go.mod
```

**Expected `go.mod`:**
```go
module github.com/yourusername/agent-config

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/fatih/color v1.16.0
)

require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/mattn/go-colorable v0.1.13 // indirect
    github.com/mattn/go-isatty v0.0.20 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    golang.org/x/sys v0.14.0 // indirect
)
```

### Phase 2: Core Implementation (4-6 hours)

#### Step 2.1: Implement Project Structure (1 hour)

**Priority Order:**
1. `main.go` - Entry point
2. `cmd/root.go` - Root command
3. `internal/config/template.go` - Templates and constants
4. `internal/parser/markdown.go` - Markdown parsing
5. `internal/symlink/manager.go` - Symlink operations

**Implementation Checklist:**
- [ ] Create `main.go`
- [ ] Create `cmd/root.go` with Cobra setup
- [ ] Define templates in `internal/config/template.go`
- [ ] Implement markdown parser in `internal/parser/markdown.go`
- [ ] Implement symlink manager in `internal/symlink/manager.go`

#### Step 2.2: Implement Commands (2-3 hours)

**Order of Implementation:**
1. `cmd/init.go` - Most important, creates foundation
2. `cmd/add.go` - Core functionality for adding rules
3. `cmd/list.go` - Reading functionality
4. `cmd/edit.go` - Simple editor integration
5. `cmd/symlink.go` - Symlink management

**Testing Strategy for Each Command:**
```bash
# After implementing each command, test it:
go build -o agent-config .

# Test init
./agent-config init --help
./agent-config init --claude

# Test add
./agent-config add "Test rule"
./agent-config add --section "Code Style" "Another test"

# Test list
./agent-config list
./agent-config list --section "Code Style"

# And so on...
```

#### Step 2.3: Build and Test (1 hour)

```bash
# Build for M4 Mac
go build -o agent-config .

# Test the binary
./agent-config --help
./agent-config version  # If we add version command

# Run Go tests (if we write them)
go test ./...

# Run with race detector
go test -race ./...
```

### Phase 3: Documentation and Polish (1-2 hours)

#### Step 3.1: Documentation
- [ ] Write comprehensive README.md
- [ ] Create QUICKSTART.md
- [ ] Write CONTRIBUTING.md
- [ ] Add code comments to exported functions
- [ ] Create usage examples

#### Step 3.2: Build Automation
- [ ] Create Makefile
- [ ] Add build targets
- [ ] Add cross-compilation support

```bash
# Test Makefile targets
make build
make test
make install
```

### Phase 4: Cross-Platform Testing (1 hour)

#### Step 4.1: Build for Multiple Platforms
```bash
# Create release builds
make release

# Or manually:
GOOS=darwin GOARCH=arm64 go build -o dist/agent-config-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o dist/agent-config-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o dist/agent-config-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o dist/agent-config-windows-amd64.exe .
```

#### Step 4.2: Test Symlinks
```bash
# Test on macOS (your M4 Mac)
./agent-config init --all
ls -la CLAUDE.md .cursorrules .windsurfrules

# Verify symlinks work
cat CLAUDE.md
cat .cursorrules
```

### Phase 5: Installation and Distribution (30 minutes)

#### Step 5.1: Local Installation
```bash
# Install to GOPATH/bin
go install .

# Verify installation
which agent-config
agent-config --help
```

#### Step 5.2: Create GitHub Repository
```bash
git init
git add .
git commit -m "Initial commit: agent-config CLI tool"
git remote add origin https://github.com/yourusername/agent-config.git
git push -u origin main
```

## Development Workflow

### Daily Development Cycle

```bash
# 1. Make changes to code
vim cmd/init.go

# 2. Format code (Go's built-in formatter)
go fmt ./...
# Or format all files
gofmt -s -w .

# 3. Check for issues
go vet ./...

# 4. Run tests
go test ./...

# 5. Build and test locally
go build -o agent-config .
./agent-config init --claude

# 6. Clean up dependencies
go mod tidy
```

### Adding New Dependencies

```bash
# Example: Adding a new testing library
go get github.com/stretchr/testify@latest

# The import will automatically be added to go.mod
# Then clean up
go mod tidy

# Verify
cat go.mod
```

### Removing Dependencies

```bash
# 1. Remove the import from your code
# 2. Run go mod tidy
go mod tidy

# This automatically removes unused dependencies
```

## Time Estimates

| Phase | Time | Tasks |
|-------|------|-------|
| **Phase 1: Setup** | 30 min | Install Go, create structure, add deps |
| **Phase 2: Core Implementation** | 4-6 hours | Implement all commands and core logic |
| **Phase 3: Documentation** | 1-2 hours | Write docs and examples |
| **Phase 4: Testing** | 1 hour | Cross-platform builds and testing |
| **Phase 5: Distribution** | 30 min | Git setup and installation |
| **Total** | **7-10 hours** | Full MVP implementation |

## Quick Start for M4 Mac

### Complete Setup Script

```bash
#!/bin/bash
# Quick setup script for M4 MacBook Pro

echo "=== agent-config Setup for M4 Mac ==="

# 1. Install Go via Homebrew
echo "Installing Go..."
brew install go

# 2. Verify installation
echo "Verifying Go installation..."
go version

# 3. Navigate to project
cd ~/projects
git clone https://github.com/yourusername/agent-config.git
cd agent-config

# 4. Download dependencies
echo "Downloading dependencies..."
go mod download

# 5. Build for M4 (ARM64)
echo "Building for M4 Mac..."
go build -o agent-config .

# 6. Test the binary
echo "Testing binary..."
./agent-config --help

# 7. Install globally
echo "Installing to GOPATH/bin..."
go install .

echo "✓ Setup complete!"
echo "Run 'agent-config --help' to get started"
```

## Troubleshooting

### Common Issues on M4 Mac

**Issue 1: "go: command not found"**
```bash
# Solution: Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.zshrc
source ~/.zshrc
```

**Issue 2: Architecture mismatch**
```bash
# Verify you're building for ARM64
go env GOARCH
# Should output: arm64

# If it shows amd64, force ARM64:
GOARCH=arm64 go build -o agent-config .
```

**Issue 3: Module not found**
```bash
# Solution: Download dependencies
go mod download
go mod tidy
```

**Issue 4: Import cycle**
```bash
# Solution: Reorganize packages
# Don't import main package from internal packages
# Keep dependencies flowing one way: main → cmd → internal
```

## Development Best Practices

### Go-Specific Rules

1. **Use `go fmt` always** - It's the standard, no configuration needed
2. **Run `go mod tidy` regularly** - Keeps dependencies clean
3. **Use `go vet` before committing** - Catches common mistakes
4. **Keep packages small and focused** - Unix philosophy
5. **Exported names start with capital letter** - `AddRule` not `addRule`

### Project-Specific Rules

Add these to `agent.md` in the project root:

```markdown
# Agent Configuration for agent-config Development

## General Rules

- Never manually edit go.mod with version numbers
- Always use `go get package@latest` to add dependencies
- Run `go mod tidy` after adding or removing dependencies
- Use `go fmt` before every commit

## Code Style

- Follow standard Go conventions (gofmt, go vet)
- Use descriptive variable names
- Keep functions small and focused
- Document all exported functions

## Commands

- `go build -o agent-config .` - Build binary
- `go test ./...` - Run all tests
- `go mod tidy` - Clean dependencies
- `make build` - Build with Makefile
- `make install` - Install to GOPATH/bin

## Architecture

- `cmd/` - Cobra commands (CLI interface)
- `internal/` - Internal packages (not importable)
- `internal/config/` - Templates and constants
- `internal/parser/` - Markdown parsing logic
- `internal/symlink/` - Symlink operations

## Important Notes

- This tool runs on ARM64 (M4 Mac) natively
- Test symlinks work correctly on macOS
- Keep binary size under 10MB
- Support cross-compilation for Linux/Windows

## Testing

- Test all commands manually after changes
- Verify symlinks work on macOS
- Check cross-compilation builds succeed
```

## Next Steps

1. **Install Go** (30 min)
   ```bash
   brew install go
   go version
   ```

2. **Clone project** (5 min)
   ```bash
   cd ~/projects
   # Extract the agent-config folder from outputs
   cd agent-config
   ```

3. **Setup dependencies** (5 min)
   ```bash
   go mod download
   go mod tidy
   ```

4. **Build and test** (10 min)
   ```bash
   go build -o agent-config .
   ./agent-config --help
   ./agent-config init --claude
   ```

5. **Install globally** (2 min)
   ```bash
   go install .
   agent-config --help
   ```

6. **Start using it!** (ongoing)
   ```bash
   cd ~/projects/datapred
   agent-config init --claude
   agent-config add "Use uv for Python dependency management"
   ```

## Success Criteria

- [ ] Go installed and working on M4 Mac
- [ ] Project builds without errors
- [ ] All commands work as expected
- [ ] Symlinks created successfully
- [ ] Binary installed to PATH
- [ ] Can use `agent-config` from any directory
- [ ] Cross-compilation works for Linux/Windows

---

**Estimated Total Time: 7-10 hours for full MVP implementation**

Start with Phase 1 (30 min) to get your environment set up, then work through the phases in order!
