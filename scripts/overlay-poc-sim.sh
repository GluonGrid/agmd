#!/usr/bin/env bash
# PoC: Parallel Agents Workspace Isolation (Simulated)
# Tests the conflict detection logic without requiring FUSE

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log() { echo -e "${BLUE}[PoC]${NC} $1"; }
success() { echo -e "${GREEN}[✓]${NC} $1"; }
warn() { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[✗]${NC} $1"; }

POC_DIR="${1:-/tmp/overlay-poc-sim}"

cleanup() {
    rm -rf "$POC_DIR"
    success "Cleanup complete"
}

setup() {
    log "Setting up at $POC_DIR..."
    mkdir -p "$POC_DIR/base"
    mkdir -p "$POC_DIR/agents/agent-A/upper"
    mkdir -p "$POC_DIR/agents/agent-A/merged"
    mkdir -p "$POC_DIR/agents/agent-B/upper"
    mkdir -p "$POC_DIR/agents/agent-B/merged"
}

create_base() {
    log "Creating base repository..."

    cat > "$POC_DIR/base/main.ts" << 'EOF'
// Main entry point
import { Database } from './db';
import { Server } from './server';

const db = new Database();
const server = new Server(db);
server.start(3000);
EOF

    cat > "$POC_DIR/base/db.ts" << 'EOF'
// Database module
export class Database {
    connect() {
        console.log('Connecting...');
    }
    query(sql: string) {
        return [];
    }
}
EOF

    cat > "$POC_DIR/base/server.ts" << 'EOF'
// Server module
export class Server {
    constructor(private db: any) {}
    start(port: number) {
        console.log(`Starting on ${port}`);
    }
}
EOF

    cat > "$POC_DIR/base/config.json" << 'EOF'
{"port": 3000, "env": "dev"}
EOF

    # Copy base to each agent's merged view
    cp -r "$POC_DIR/base/"* "$POC_DIR/agents/agent-A/merged/"
    cp -r "$POC_DIR/base/"* "$POC_DIR/agents/agent-B/merged/"

    success "Base created: main.ts, db.ts, server.ts, config.json"
}

agent_edit() {
    local agent=$1
    local file=$2
    local content=$3

    echo "$content" > "$POC_DIR/agents/$agent/merged/$file"
    mkdir -p "$(dirname "$POC_DIR/agents/$agent/upper/$file")"
    echo "$content" > "$POC_DIR/agents/$agent/upper/$file"
}

simulate_work() {
    log "Simulating parallel agent work..."
    echo ""

    echo -e "${CYAN}Agent A${NC}: Adding authentication..."
    agent_edit "agent-A" "db.ts" '// Database - Modified by Agent A
export class Database {
    private authed = false;
    connect() { console.log("Connecting..."); }
    auth(user: string) { this.authed = true; }
    query(sql: string) {
        if (!this.authed) throw new Error("Not authenticated");
        return [];
    }
}'

    agent_edit "agent-A" "auth.ts" '// Auth module - Created by Agent A
export function validateToken(t: string): boolean {
    return t.length > 10;
}'

    agent_edit "agent-A" "main.ts" '// Main - Modified by Agent A
import { Database } from "./db";
import { Server } from "./server";
import { validateToken } from "./auth";

const db = new Database();
db.auth("admin");
const server = new Server(db);
server.start(3000);'

    echo -e "${CYAN}Agent B${NC}: Adding middleware support..."
    agent_edit "agent-B" "server.ts" '// Server - Modified by Agent B
export class Server {
    private middlewares: Function[] = [];
    constructor(private db: any) {}
    use(fn: Function) { this.middlewares.push(fn); }
    start(port: number) {
        console.log(`Starting on ${port} with ${this.middlewares.length} middlewares`);
    }
}'

    agent_edit "agent-B" "config.json" '{"port": 8080, "env": "prod", "middlewares": ["cors"]}'

    # Agent B also modifies main.ts -> CONFLICT!
    agent_edit "agent-B" "main.ts" '// Main - Modified by Agent B
import { Database } from "./db";
import { Server } from "./server";

const db = new Database();
const server = new Server(db);
server.use((req: any) => console.log(req));
server.start(8080);'

    success "Both agents completed their work"
}

detect_conflicts() {
    echo ""
    log "Analyzing changes..."
    echo ""

    # Use temp files instead of associative arrays for compatibility
    local tmp_files=$(mktemp)
    local conflicts=""
    local safe=""
    local new_files=""

    # Collect all modified files from both agents
    for agent in agent-A agent-B; do
        if [ -d "$POC_DIR/agents/$agent/upper" ]; then
            find "$POC_DIR/agents/$agent/upper" -type f 2>/dev/null | while read file; do
                rel="${file#$POC_DIR/agents/$agent/upper/}"
                echo "$rel|$agent" >> "$tmp_files"
            done
        fi
    done

    echo "┌────────────────────────────────────────────────────────────┐"
    echo "│                   CHANGE ANALYSIS                          │"
    echo "├────────────────────────────────────────────────────────────┤"

    # Process each unique file
    sort -u -t'|' -k1,1 "$tmp_files" | cut -d'|' -f1 | sort -u | while read file; do
        # Count how many agents modified this file
        count=$(grep "^$file|" "$tmp_files" | wc -l | tr -d ' ')
        agents=$(grep "^$file|" "$tmp_files" | cut -d'|' -f2 | tr '\n' ',' | sed 's/,$//')

        if [ -f "$POC_DIR/base/$file" ]; then
            if [ "$count" -gt 1 ]; then
                printf "│ ${RED}✗ CONFLICT${NC}  %-28s ${CYAN}[%s]${NC}\n" "$file" "$agents"
            else
                printf "│ ${GREEN}✓ SAFE${NC}      %-28s ${CYAN}[%s]${NC}\n" "$file" "$agents"
            fi
        else
            printf "│ ${BLUE}+ NEW${NC}       %-28s ${CYAN}[%s]${NC}\n" "$file" "$agents"
        fi
    done

    echo "└────────────────────────────────────────────────────────────┘"
    echo ""

    # Show conflict details for main.ts
    if [ -f "$POC_DIR/agents/agent-A/upper/main.ts" ] && [ -f "$POC_DIR/agents/agent-B/upper/main.ts" ]; then
        echo "╔════════════════════════════════════════════════════════════╗"
        echo -e "║ ${RED}CONFLICT DETAILS${NC}: main.ts"
        echo "╠════════════════════════════════════════════════════════════╣"
        echo "║ BASE version:"
        head -3 "$POC_DIR/base/main.ts" | sed 's/^/║   /'
        echo "║   ..."
        echo "╠════════════════════════════════════════════════════════════╣"
        echo -e "║ ${CYAN}agent-A${NC} version:"
        head -3 "$POC_DIR/agents/agent-A/upper/main.ts" | sed 's/^/║   /'
        echo "║   ..."
        echo "╠════════════════════════════════════════════════════════════╣"
        echo -e "║ ${CYAN}agent-B${NC} version:"
        head -3 "$POC_DIR/agents/agent-B/upper/main.ts" | sed 's/^/║   /'
        echo "║   ..."
        echo "╚════════════════════════════════════════════════════════════╝"
    fi

    rm -f "$tmp_files"

    echo ""
    echo "┌────────────────────────────────────────────────────────────┐"
    echo "│                      SUMMARY                               │"
    echo "├────────────────────────────────────────────────────────────┤"
    echo "│  This demonstrates:                                        │"
    echo "│  • Each agent works in isolation (merged/)                 │"
    echo "│  • Changes are tracked separately (upper/)                 │"
    echo "│  • Conflicts detected BEFORE merge attempt                 │"
    echo "│  • Non-conflicting changes merge automatically             │"
    echo "└────────────────────────────────────────────────────────────┘"
}

merge_safe_changes() {
    echo ""
    log "Merging non-conflicting changes..."

    mkdir -p "$POC_DIR/merged_result"
    cp -r "$POC_DIR/base/"* "$POC_DIR/merged_result/"

    # Apply agent-A's unique changes
    for file in db.ts auth.ts; do
        if [ -f "$POC_DIR/agents/agent-A/upper/$file" ]; then
            cp "$POC_DIR/agents/agent-A/upper/$file" "$POC_DIR/merged_result/$file"
            success "  Applied: $file (from agent-A)"
        fi
    done

    # Apply agent-B's unique changes
    for file in server.ts config.json; do
        if [ -f "$POC_DIR/agents/agent-B/upper/$file" ]; then
            cp "$POC_DIR/agents/agent-B/upper/$file" "$POC_DIR/merged_result/$file"
            success "  Applied: $file (from agent-B)"
        fi
    done

    warn "  Skipped: main.ts (CONFLICT - needs resolution)"

    echo ""
    log "Merged result:"
    ls -la "$POC_DIR/merged_result/"
}

show_diff() {
    echo ""
    log "Showing conflict diff for main.ts:"
    echo ""
    echo "--- Base ---"
    cat "$POC_DIR/base/main.ts"
    echo ""
    echo "--- Agent A (wants auth) ---"
    cat "$POC_DIR/agents/agent-A/upper/main.ts"
    echo ""
    echo "--- Agent B (wants middleware) ---"
    cat "$POC_DIR/agents/agent-B/upper/main.ts"
}

main() {
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║   Parallel Agents Workspace Isolation PoC                  ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""

    cleanup 2>/dev/null || true
    setup
    create_base
    simulate_work
    detect_conflicts
    merge_safe_changes

    echo ""
    success "PoC complete!"
    echo ""
    echo "Key insight: upperdirs ARE the diffs!"
    echo ""
    echo "Explore:"
    echo "  ls $POC_DIR/base/                    # Original files"
    echo "  ls $POC_DIR/agents/agent-A/upper/    # Only A's changes"
    echo "  ls $POC_DIR/agents/agent-B/upper/    # Only B's changes"
    echo "  ls $POC_DIR/merged_result/           # Safe merge result"
    echo ""
    echo "View conflict:"
    echo "  diff $POC_DIR/agents/agent-A/upper/main.ts $POC_DIR/agents/agent-B/upper/main.ts"
    echo ""
    echo "Cleanup: rm -rf $POC_DIR"
}

case "${1:-}" in
    -c|--cleanup) cleanup; exit 0 ;;
    -d|--diff) show_diff; exit 0 ;;
    -h|--help)
        echo "Usage: $0 [POC_DIR]"
        echo "  -c, --cleanup   Clean up"
        echo "  -d, --diff      Show conflict diff"
        echo "  -h, --help      Show help"
        exit 0
        ;;
    *) main ;;
esac
