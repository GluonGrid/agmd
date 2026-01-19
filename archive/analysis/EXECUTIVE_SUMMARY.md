# Executive Summary: Agent Instructions Consolidation

**Project:** Consolidation of Agent Markdown Files
**Date:** 2026-01-19
**Repositories Analyzed:** 11 repositories, 13 files, 916 total lines of content

---

## Problem Statement

Agent instruction files (`AGENTS.md`, `CLAUDE.md`) across steipete's repositories contain **60-80% duplicated content**, leading to:

- **Maintenance burden** - Updates must be applied to multiple files
- **Inconsistency risk** - Rules drift out of sync between repos
- **Onboarding friction** - New repos start from scratch or copy-paste
- **Wasted effort** - Same rules written multiple times

---

## Key Findings

### Duplication Analysis

| Content Type | Duplication Level | Lines Affected |
|--------------|-------------------|----------------|
| Coding Philosophy (no V2 files, strict typing, file size) | **90%** | ~100 lines |
| Git/Commit Rules (Conventional Commits, wrapper scripts) | **90%** | ~80 lines |
| Dependency Management (ask before adding, stick to package manager) | **85%** | ~60 lines |
| Documentation Rules (update when behavior changes) | **80%** | ~50 lines |
| Testing Guidelines (frameworks, coverage, when to test) | **75%** | ~80 lines |
| Formatting Rules (tool configurations) | **75%** | ~60 lines |
| **Overall Average** | **75-80%** | **~600 lines** |

### Repository Maturity

**Complete (needs minimal work):**
- ✅ **mcporter** - Master file with `<shared>` tags (ideal model)
- ✅ **clawdis** - Very comprehensive, multi-agent safety rules
- ✅ **Peekaboo** - Well-structured, references shared

**Needs Expansion (missing key guardrails):**
- ⚠️ **summarize** - Too minimal, lacks basic structure
- ⚠️ **Matcha** - Only covers testing, missing general rules
- ⚠️ **sweetlink** - Missing dependency approval rules
- ⚠️ **tokentally** - Missing coverage thresholds, git wrappers

**Needs Shared Reference:**
- 📝 **Trimmy**, **claude-code-mcp**, **VibeMeter**, **macos-automator-mcp** - Need to reference central guardrails

---

## Recommended Solution

### The Quick Win: Single Line Addition

Add this line to the top of each repo's AGENTS.md:

```markdown
READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING (skip if missing).
```

**Benefit:** Instant access to shared guardrails with zero duplication.

### The Complete Solution: mcporter Model

**Structure:**
```markdown
# Repository Guidelines

## Shared Guardrails
READ ~/Projects/agent-scripts/AGENTS.MD BEFORE ANYTHING (skip if missing).

<shared>
[Optional: inline copy of shared content for offline reference]
</shared>

## Project Structure & Modules
[Project-specific directory layout]

## Build, Test, and Development Commands
[Project-specific commands]

## Project-Specific Rules
[Unique rules for this repo]
```

**Benefits:**
- **DRY:** 60-80% less duplication
- **Consistent:** All repos follow same core rules
- **Maintainable:** Update once, applies everywhere
- **Flexible:** Project-specific rules stay local

---

## Impact Analysis

### Before Consolidation
- **916 total lines** across 13 files
- **~600-700 lines** duplicated
- **13 files** to update when changing shared rules
- **No templates** for new repos

### After Consolidation
- **1 central file** (~350 lines) + **13 local files** (~20-50 lines each)
- **~200-300 lines** of unique content (where it belongs)
- **1 file** to update for shared rules
- **2 templates** (TypeScript, Swift) for new repos

### Metrics

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Duplicate Content | 60-80% | <30% | **>50% reduction** |
| Files to Update (shared rules) | 13 files | 1 file | **92% less work** |
| Lines per Local File | 70 avg | 30 avg | **57% reduction** |
| New Repo Setup Time | Manual | Use template | **Faster onboarding** |

---

## Implementation Plan

### Phase 1: Immediate (Week 1)

**Goal:** Establish foundation and fix high-priority repos

1. ✅ Verify `~/Projects/agent-scripts/AGENTS.MD` is current
2. Update high-priority repos (4 repos):
   - [ ] **summarize** - Add reference + expand structure
   - [ ] **sweetlink** - Add reference + dep rules
   - [ ] **tokentally** - Add reference + coverage/git rules
   - [ ] **Matcha** - Add reference + general guardrails

**Effort:** ~2-3 hours
**Impact:** HIGH - Fixes repos with most missing content

### Phase 2: Short-Term (Month 1)

**Goal:** Standardize remaining repos and extract valuable patterns

3. Update medium-priority repos (4 repos):
   - [ ] **Trimmy** - Add reference + expand testing
   - [ ] **claude-code-mcp** - Add reference + coding standards
   - [ ] **VibeMeter** - Add reference + changelog/testing
   - [ ] **macos-automator-mcp** - Add reference + general rules

4. Extract to shared file:
   - [ ] Multi-agent safety rules (from clawdis)
   - [ ] Changelog workflow (from clawdis)
   - [ ] "Don't edit node_modules" warning
   - [ ] Troubleshooting pattern (from macos-automator-mcp)

5. Standardize section names across all repos

**Effort:** ~4-5 hours
**Impact:** MEDIUM - Achieves consistency across all repos

### Phase 3: Long-Term (Ongoing)

**Goal:** Maintain and improve

6. Create templates:
   - [ ] TypeScript project template
   - [ ] Swift project template

7. Establish maintenance:
   - [ ] Quarterly review of shared file
   - [ ] Contribution guidelines
   - [ ] Sync script (optional)

**Effort:** ~2-3 hours initially, ~1 hour quarterly
**Impact:** LOW - Quality of life improvements

---

## Deliverables

### ✅ Analysis Complete

1. **ANALYSIS_REPORT.md** (1,517 lines)
   - Detailed duplication analysis
   - Missing content matrix
   - Repository-by-repository breakdown
   - Exact text snippets and examples

2. **CONSOLIDATION_PLAN.md** (this document)
   - Implementation phases
   - Priority matrix
   - Templates
   - Success metrics

3. **PROPOSED_SHARED_AGENTS.MD** (350+ lines)
   - Complete shared guardrails file
   - Language-specific sections (TypeScript, Swift)
   - Multi-agent safety rules
   - Security guidelines
   - All consolidated best practices

4. **fetched_files/** (14 files)
   - All original agent instruction files
   - Ready for reference during implementation

### 📋 Ready for Implementation

All analysis and design work is complete. Ready to begin Phase 1.

---

## Unique Valuable Patterns Discovered

### Multi-Agent Safety (from clawdis)
**Problem:** Multiple AI agents working in same repo can conflict
**Solution:** Explicit rules about git stash, branch switching, scoped commits

### Operational Learnings (from macos-automator-mcp)
**Problem:** Hard-won knowledge gets lost
**Solution:** Dedicated section capturing troubleshooting insights

### Reference Implementation Pattern (from Matcha)
**Problem:** Implementing something that should match existing tool
**Solution:** "When unsure, align with what [reference] does"

### Changelog Workflow (from clawdis)
**Problem:** Inconsistent changelog entries
**Solution:** Detailed rules for PR attribution, contributor thanks, version management

---

## Risk Assessment

### Low Risk
- **Reference-only approach** - Minimal changes to repos
- **Gradual rollout** - Can implement one repo at a time
- **Reversible** - Can remove reference line if issues arise

### Mitigations
- **Offline access** - Optional `<shared>` block for offline copy
- **Backward compatible** - Old files still work while transitioning
- **No breaking changes** - Only additions, no removals

---

## Success Criteria

✅ **Phase 1 Complete When:**
- High-priority repos reference shared file
- No critical missing content in any repo
- Templates created for new repos

✅ **Phase 2 Complete When:**
- All repos reference shared file
- Section names standardized
- Valuable patterns extracted to shared

✅ **Long-Term Success:**
- Shared file actively maintained
- New repos use templates
- Updates propagate quickly

---

## Recommendation

**Approve and proceed with Phase 1 implementation.**

The analysis shows clear duplication, the solution is proven (mcporter already uses it), and the implementation is low-risk with high impact.

**Next Action:** Begin Phase 1 updates to high-priority repos (summarize, sweetlink, tokentally, Matcha).

---

## Appendix: File Locations

All deliverables are in `/Users/sky/git/agent-md/`:

- `EXECUTIVE_SUMMARY.md` (this file)
- `CONSOLIDATION_PLAN.md` - Detailed implementation plan
- `PROPOSED_SHARED_AGENTS.MD` - Complete shared guardrails file
- `fetched_files/ANALYSIS_REPORT.md` - Detailed analysis (1,517 lines)
- `fetched_files/FETCH_SUMMARY.md` - Files fetched summary
- `fetched_files/*.md` - All 14 original agent instruction files

**Contact:** Ready to answer questions or begin implementation.
