#!/bin/bash
# PoC: Parallel Agents with OverlayFS on macOS
# Tests isolated workspaces for multiple agents editing the same repo

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() { echo -e "${BLUE}[PoC]${NC} $1"; }
success() { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; }

# Configuration
POC_DIR="${1:-/tmp/overlay-poc}"
AGENTS=("agent-A" "agent-B")

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    if ! command -v fuse-overlayfs &> /dev/null; then
        error "fuse-overlayfs not found"
        echo ""
        echo "Install with:"
        echo "  brew install macfuse"
        echo "  brew install fuse-overlayfs"
        echo ""
        echo "Then approve macFUSE in System Settings → Privacy & Security"
        echo "Reboot after installation"
        exit 1
    fi

    success "fuse-overlayfs found"
}

# Clean up any existing mounts
cleanup() {
    log "Cleaning up..."

    for agent in "${AGENTS[@]}"; do
        mount_point="$POC_DIR/agents/$agent/merged"
        if mount | grep -q "$mount_point"; then
            umount "$mount_point" 2>/dev/null || fusermount -u "$mount_point" 2>/dev/null || true
        fi
    done

    rm -rf "$POC_DIR"
    success "Cleanup complete"
}

# Set up directory structure
setup_directories() {
    log "Setting up directories at $POC_DIR..."

    mkdir -p "$POC_DIR/base"

    for agent in "${AGENTS[@]}"; do
        mkdir -p "$POC_DIR/agents/$agent/upper"
        mkdir -p "$POC_DIR/agents/$agent/work"
        mkdir -p "$POC_DIR/agents/$agent/merged"
    done

    success "Directory structure created"
}

# Create base repository (simulated)
create_base_repo() {
    log "Creating base repository..."

    # Simulate a small codebase
    cat > "$POC_DIR/base/main.ts" << 'EOF'
// Main application entry point
import { Database } from './db';
import { Server } from './server';

const db = new Database();
const server = new Server(db);

server.start(3000);
console.log('Server running on port 3000');
EOF

    cat > "$POC_DIR/base/db.ts" << 'EOF'
// Database module
export class Database {
    private connection: any;

    connect() {
        console.log('Connecting to database...');
    }

    query(sql: string) {
        return [];
    }
}
EOF

    cat > "$POC_DIR/base/server.ts" << 'EOF'
// Server module
import { Database } from './db';

export class Server {
    constructor(private db: Database) {}

    start(port: number) {
        this.db.connect();
        console.log(`Starting server on port ${port}`);
    }
}
EOF

    cat > "$POC_DIR/base/config.json" << 'EOF'
{
    "port": 3000,
    "database": "postgres://localhost/app"
}
EOF

    success "Base repository created with 4 files"
    ls -la "$POC_DIR/base/"
}

# Mount overlays for each agent
mount_overlays() {
    log "Mounting overlays..."

    for agent in "${AGENTS[@]}"; do
        log "  Mounting overlay for $agent..."

        fuse-overlayfs \
            -o lowerdir="$POC_DIR/base" \
            -o upperdir="$POC_DIR/agents/$agent/upper" \
            -o workdir="$POC_DIR/agents/$agent/work" \
            "$POC_DIR/agents/$agent/merged"

        success "  $agent mounted at $POC_DIR/agents/$agent/merged"
    done

    success "All overlays mounted"
}

# Simulate agent work
simulate_agent_work() {
    log "Simulating agent work..."

    # Agent A: Modifies db.ts and adds a new file
    log "  Agent A: Adding authentication to database..."
    cat > "$POC_DIR/agents/agent-A/merged/db.ts" << 'EOF'
// Database module - Modified by Agent A
export class Database {
    private connection: any;
    private authenticated: boolean = false;

    connect() {
        console.log('Connecting to database...');
    }

    authenticate(user: string, pass: string) {
        // Agent A added authentication
        this.authenticated = true;
        console.log(`Authenticated as ${user}`);
    }

    query(sql: string) {
        if (!this.authenticated) throw new Error('Not authenticated');
        return [];
    }
}
EOF

    cat > "$POC_DIR/agents/agent-A/merged/auth.ts" << 'EOF'
// Auth module - Created by Agent A
export function validateToken(token: string): boolean {
    return token.length > 0;
}
EOF
    success "  Agent A: Modified db.ts, created auth.ts"

    # Agent B: Modifies server.ts and config.json
    log "  Agent B: Adding middleware to server..."
    cat > "$POC_DIR/agents/agent-B/merged/server.ts" << 'EOF'
// Server module - Modified by Agent B
import { Database } from './db';

export class Server {
    private middlewares: Function[] = [];

    constructor(private db: Database) {}

    use(middleware: Function) {
        // Agent B added middleware support
        this.middlewares.push(middleware);
    }

    start(port: number) {
        this.db.connect();
        console.log(`Starting server on port ${port}`);
        console.log(`Loaded ${this.middlewares.length} middlewares`);
    }
}
EOF

    cat > "$POC_DIR/agents/agent-B/merged/config.json" << 'EOF'
{
    "port": 8080,
    "database": "postgres://localhost/app",
    "middlewares": ["cors", "logging"]
}
EOF
    success "  Agent B: Modified server.ts, config.json"

    # Both agents modify main.ts (CONFLICT!)
    log "  Agent A: Updating main.ts imports..."
    cat > "$POC_DIR/agents/agent-A/merged/main.ts" << 'EOF'
// Main application entry point - Modified by Agent A
import { Database } from './db';
import { Server } from './server';
import { validateToken } from './auth';

const db = new Database();
db.authenticate('admin', 'secret');

const server = new Server(db);
server.start(3000);
console.log('Server running on port 3000');
EOF

    log "  Agent B: Updating main.ts server config..."
    cat > "$POC_DIR/agents/agent-B/merged/main.ts" << 'EOF'
// Main application entry point - Modified by Agent B
import { Database } from './db';
import { Server } from './server';

const db = new Database();
const server = new Server(db);

server.use((req: any) => console.log('Request:', req));
server.start(8080);
console.log('Server running on port 8080 with middleware');
EOF

    success "  Both agents modified main.ts (intentional conflict)"
}

# Detect conflicts by comparing upperdirs
detect_conflicts() {
    log "Detecting conflicts..."
    echo ""

    # Collect all modified files
    declare -A modified_by

    for agent in "${AGENTS[@]}"; do
        upper_dir="$POC_DIR/agents/$agent/upper"
        if [ -d "$upper_dir" ]; then
            for file in $(find "$upper_dir" -type f 2>/dev/null); do
                rel_path="${file#$upper_dir/}"
                if [ -z "${modified_by[$rel_path]}" ]; then
                    modified_by[$rel_path]="$agent"
                else
                    modified_by[$rel_path]="${modified_by[$rel_path]},$agent"
                fi
            done
        fi
    done

    echo "┌─────────────────────────────────────────────────────────┐"
    echo "│                    FILE MODIFICATIONS                   │"
    echo "├─────────────────────────────────────────────────────────┤"

    conflicts=()
    safe_files=()
    new_files=()

    for file in "${!modified_by[@]}"; do
        agents="${modified_by[$file]}"

        # Check if file exists in base
        if [ -f "$POC_DIR/base/$file" ]; then
            if [[ "$agents" == *","* ]]; then
                # Multiple agents modified same file
                conflicts+=("$file")
                printf "│ ${RED}CONFLICT${NC}  %-30s [%s]\n" "$file" "$agents"
            else
                safe_files+=("$file")
                printf "│ ${GREEN}SAFE${NC}      %-30s [%s]\n" "$file" "$agents"
            fi
        else
            new_files+=("$file")
            printf "│ ${BLUE}NEW${NC}       %-30s [%s]\n" "$file" "$agents"
        fi
    done

    echo "└─────────────────────────────────────────────────────────┘"
    echo ""

    # Show conflict details
    if [ ${#conflicts[@]} -gt 0 ]; then
        warn "Found ${#conflicts[@]} conflict(s)!"
        echo ""

        for file in "${conflicts[@]}"; do
            echo "┌─────────────────────────────────────────────────────────┐"
            echo "│ CONFLICT: $file"
            echo "├─────────────────────────────────────────────────────────┤"
            echo "│ Base version:"
            head -3 "$POC_DIR/base/$file" | sed 's/^/│   /'
            echo "│ ..."
            echo "├─────────────────────────────────────────────────────────┤"

            for agent in "${AGENTS[@]}"; do
                if [ -f "$POC_DIR/agents/$agent/upper/$file" ]; then
                    echo "│ $agent version:"
                    head -3 "$POC_DIR/agents/$agent/upper/$file" | sed 's/^/│   /'
                    echo "│ ..."
                fi
            done
            echo "└─────────────────────────────────────────────────────────┘"
            echo ""
        done
    else
        success "No conflicts detected!"
    fi

    # Summary
    echo ""
    echo "Summary:"
    echo "  Safe modifications: ${#safe_files[@]}"
    echo "  New files:          ${#new_files[@]}"
    echo "  Conflicts:          ${#conflicts[@]}"
}

# Show what each agent sees
show_agent_views() {
    log "Showing agent views..."
    echo ""

    for agent in "${AGENTS[@]}"; do
        echo "┌─────────────────────────────────────────────────────────┐"
        echo "│ $agent sees (in merged/):"
        echo "├─────────────────────────────────────────────────────────┤"
        ls -la "$POC_DIR/agents/$agent/merged/" | tail -n +2 | sed 's/^/│ /'
        echo "└─────────────────────────────────────────────────────────┘"
        echo ""
    done

    for agent in "${AGENTS[@]}"; do
        echo "┌─────────────────────────────────────────────────────────┐"
        echo "│ $agent modified (in upper/):"
        echo "├─────────────────────────────────────────────────────────┤"
        if [ -d "$POC_DIR/agents/$agent/upper" ] && [ "$(ls -A $POC_DIR/agents/$agent/upper 2>/dev/null)" ]; then
            ls -la "$POC_DIR/agents/$agent/upper/" | tail -n +2 | sed 's/^/│ /'
        else
            echo "│ (no modifications)"
        fi
        echo "└─────────────────────────────────────────────────────────┘"
        echo ""
    done
}

# Create merged result (for non-conflicting files)
create_merged_result() {
    log "Creating merged result..."

    mkdir -p "$POC_DIR/merged_result"

    # Start with base
    cp -r "$POC_DIR/base/"* "$POC_DIR/merged_result/"

    # Apply non-conflicting changes
    for agent in "${AGENTS[@]}"; do
        upper_dir="$POC_DIR/agents/$agent/upper"
        if [ -d "$upper_dir" ]; then
            for file in $(find "$upper_dir" -type f 2>/dev/null); do
                rel_path="${file#$upper_dir/}"

                # Skip if this file was modified by multiple agents
                conflict=false
                for other_agent in "${AGENTS[@]}"; do
                    if [ "$other_agent" != "$agent" ] && [ -f "$POC_DIR/agents/$other_agent/upper/$rel_path" ]; then
                        conflict=true
                        break
                    fi
                done

                if [ "$conflict" = false ]; then
                    mkdir -p "$(dirname "$POC_DIR/merged_result/$rel_path")"
                    cp "$file" "$POC_DIR/merged_result/$rel_path"
                    success "  Applied $rel_path from $agent"
                fi
            done
        fi
    done

    warn "  Skipped conflicting files (need manual resolution)"
    echo ""
    echo "Merged result in: $POC_DIR/merged_result/"
    ls -la "$POC_DIR/merged_result/"
}

# Main
main() {
    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║     OverlayFS Parallel Agents PoC for macOS               ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    check_prerequisites
    cleanup
    setup_directories
    create_base_repo
    echo ""

    mount_overlays
    echo ""

    simulate_agent_work
    echo ""

    show_agent_views

    detect_conflicts

    create_merged_result
    echo ""

    log "PoC complete!"
    echo ""
    echo "To explore:"
    echo "  Base:      $POC_DIR/base/"
    echo "  Agent A:   $POC_DIR/agents/agent-A/merged/"
    echo "  Agent B:   $POC_DIR/agents/agent-B/merged/"
    echo "  Merged:    $POC_DIR/merged_result/"
    echo ""
    echo "To clean up:"
    echo "  $0 --cleanup"
}

# Handle arguments
case "${1:-}" in
    --cleanup|-c)
        cleanup
        exit 0
        ;;
    --help|-h)
        echo "Usage: $0 [POC_DIR]"
        echo ""
        echo "Options:"
        echo "  --cleanup, -c    Clean up mounts and directories"
        echo "  --help, -h       Show this help"
        echo ""
        echo "Default POC_DIR: /tmp/overlay-poc"
        exit 0
        ;;
    *)
        main
        ;;
esac
