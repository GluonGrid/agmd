#!/bin/bash

# Repos with AGENTS.md
AGENTS_REPOS=(
  "agent-scripts"
  "mcporter"
  "sweetlink"
  "Trimmy"
  "summarize"
  "clawdis"
  "AXorcist"
  "tokentally"
  "Peekaboo"
)

# Repos with CLAUDE.md
CLAUDE_REPOS=(
  "claude-code-mcp"
  "Matcha"
  "VibeMeter"
  "macos-automator-mcp"
  "Peekaboo"
)

OUTPUT_DIR="fetched_files"

echo "Fetching AGENTS.md files..."
for repo in "${AGENTS_REPOS[@]}"; do
  echo "Fetching from steipete/$repo..."
  gh api repos/steipete/$repo/contents/AGENTS.md --jq '.content' 2>/dev/null | base64 -d > "$OUTPUT_DIR/${repo}_AGENTS.md" 2>/dev/null
  if [ $? -eq 0 ] && [ -s "$OUTPUT_DIR/${repo}_AGENTS.md" ]; then
    echo "  ✓ Downloaded ${repo}/AGENTS.md"
  else
    echo "  ✗ Failed or not found: ${repo}/AGENTS.md"
    rm -f "$OUTPUT_DIR/${repo}_AGENTS.md"
  fi
done

echo ""
echo "Fetching CLAUDE.md files..."
for repo in "${CLAUDE_REPOS[@]}"; do
  echo "Fetching from steipete/$repo..."
  gh api repos/steipete/$repo/contents/CLAUDE.md --jq '.content' 2>/dev/null | base64 -d > "$OUTPUT_DIR/${repo}_CLAUDE.md" 2>/dev/null
  if [ $? -eq 0 ] && [ -s "$OUTPUT_DIR/${repo}_CLAUDE.md" ]; then
    echo "  ✓ Downloaded ${repo}/CLAUDE.md"
  else
    echo "  ✗ Failed or not found: ${repo}/CLAUDE.md"
    rm -f "$OUTPUT_DIR/${repo}_CLAUDE.md"
  fi
done

echo ""
echo "Summary:"
echo "========"
ls -lh "$OUTPUT_DIR" | tail -n +2
