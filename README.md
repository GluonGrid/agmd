<div align="center">

# agmd

**A task board for you and your coding agents, driven from the CLI.**

Plain markdown task files, dependency-aware ordering, JSON output for agents, and a local web board — no server, no database, no account.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/GluonGrid/agmd.svg)](https://github.com/GluonGrid/agmd/releases)

</div>

## Quick Start

```bash
curl -fsSL https://gluongrid.dev/agmd/install.sh | bash

cd your-repo
agmd task new setup-db --content "Set up database schema"
agmd task new create-api --content "Create endpoints" --blocked-by setup-db
agmd task list
agmd task web
```

There is nothing to initialize: the first write creates `~/.agmd/task/<project>/`.

## Tasks

```bash
# Create tasks with priority and type
agmd task new setup-db --content "Set up database schema"
agmd task new fix-auth -t bug -p 0 --content "Critical auth bug"
agmd task new add-dark-mode -t feature -p 1 --content "Add dark mode"
agmd task new create-api --content "Create endpoints" --blocked-by "setup-db"
agmd task new setup-db --feature auth     # Scope task to a feature

# List tasks (auto-sorted: P0→P4, then ready → blocked → completed)
agmd task list                            # All tasks for current project
agmd task list --type bug --priority 0    # Critical bugs only
agmd task list --feature auth             # Filter by feature
agmd task list --status ready             # Filter by computed status
agmd task list --tree                     # Show dependency tree
agmd task list --all                      # Include completed tasks

# Manage status and dependencies
agmd task status setup-db completed       # Update status
agmd task edit setup-db --priority 0 --type bug  # Change priority/type
agmd task blocked-by create-api setup-db  # Add dependency
agmd task unblock create-api setup-db     # Remove dependency

# View and delete
agmd task show setup-db                   # Show task content
agmd task show --all                      # Show all tasks with content
agmd task delete setup-db --force         # Delete task
agmd task clean                           # Delete all completed tasks
```

**Priority levels:** P0 (critical), P1 (high), P2 (medium/default, hidden), P3 (low), P4 (backlog)
**Task types:** bug, feature, task (default, hidden), chore

Tasks are auto-sorted by priority then status. The `--tree` view groups tasks by feature and shows dependency chains. Dependencies via `--blocked-by` are validated on creation.

## Projects

The project is derived from the repository you run agmd in. In a Git worktree, agmd resolves to the main repository, so every worktree of a repo shares one task list. Override it with `--project <name>` or `--repo-path <path>`.

```bash
agmd task emoji 🚀                        # Show an emoji next to the project name
```

## Agents

agmd is built for several agents working on the same project at once. Each agent picks a workflow id and claims work with one command:

```bash
agmd task next --claim --workflow-id agent-1 --json   # Claim the next ready task
agmd task comment setup-db "Schema migrated" --author agent-1   # Log progress
agmd task status setup-db completed                   # Finish it
agmd task unassign setup-db                           # Release the claim (back to pending if unfinished)
```

`next --claim` atomically picks the highest-priority task that is ready and unassigned, assigns it to the workflow and marks it `in_progress`; two agents never get the same task. If the workflow already holds an unfinished task, that task is returned instead (`"action":"resume"`), so a restarted agent picks up where it left off. Without `--claim`, `agmd task next` only shows the next task.

### Stale claims

If an agent dies mid-task, its claim expires so the work isn't stuck. A claim stays alive while its workflow shows signs of life: claiming or resuming with `next --claim`, `agmd task heartbeat --workflow-id agent-1`, or a comment whose `--author` is the workflow id. After `--stale-after` without any (default `$AGMD_CLAIM_TTL`, else `2h`; `0` disables expiry), `next --claim` may hand the task to another workflow (`"action":"takeover"`, `"taken_over_from":"agent-1"`) and records the takeover as a comment. Ready, unclaimed tasks are still preferred over stale ones of the same priority, and a task whose dependencies aren't done is never handed out.

```bash
agmd task heartbeat --workflow-id agent-1             # Refresh all of agent-1's claims
agmd task heartbeat setup-db --workflow-id agent-1    # Fails with "conflict" if the claim was lost
AGMD_CLAIM_TTL=30m agmd task next --claim --workflow-id agent-2
```

### Comments

Comments are a progress log on a task, shown by `agmd task show`, included in `--json` output, and can be added from the web board. The author defaults to `$AGMD_AUTHOR`, or `user`.

### JSON and errors

Every `agmd task` command takes `--json`. Errors are printed to stderr as `{"error":{"code":"...","message":"..."}}` with a non-zero exit status, so agents can branch on the code:

| Code | Exit | Meaning |
| --- | --- | --- |
| `invalid_input` | 2 | Bad arguments |
| `not_found` | 3 | No such task, or nothing ready for `next` |
| `project_mismatch` | 4 | Task reference names another project |
| `invalid_transition` | 5 | Status change not allowed |
| `conflict` | 6 | Task already assigned to another workflow (`assign --force` takes it over) |

```bash
agmd task list --json --status ready
agmd task assign setup-db --workflow-id agent-1 --worktree-path ../wt-1
agmd task unassign --all --workflow-id agent-1
```

Concurrent use is safe: every change to a project's files happens under a per-project file lock (`<project>/.lock`), and files are replaced atomically, so the web board and any number of CLI agents can work side by side.

## Web board

```bash
agmd task web                             # Serve on a free port and open the browser
agmd task web --no-open                   # Serve without opening a browser
agmd task web --port 8765                 # Serve on a fixed port
```

`agmd task web` launches a self-contained local UI (no build step, works offline) that aggregates tasks across every project. Filter by project/status/priority/type/feature, edit status and priority inline, or open a task to edit its subject/type/feature/content. Edits write straight back to the task files, so they show up immediately in `agmd task list`.

The web UI can also **run selected tasks with Claude Code**: tick the tasks you want, click "▶ Run selected", pick a permission mode, and agmd spawns a `claude` session with a `/goal` prompt to implement them. The chosen tasks are flipped to in-progress and claimed by workflow `web-<session-id>` (kept alive while the run is going, so CLI agents leave them alone), the run's transcript streams live in a console panel, and you get a `claude --resume <id>` hint to continue the session interactively once it finishes. You can launch several runs at once, minimize or close the console, and reopen any run later from the **Sessions** menu in the top bar. The server stays bound to `127.0.0.1` and asks for confirmation before launching, since this spawns processes that can modify your files.

## Storage

Each task is a markdown file with YAML frontmatter at `~/.agmd/task/<project>/<name>.md`. Claims live in `<project>/assignments.json` and comments in `<project>/comments/<task-id>.json`. Set `AGMD_HOME` to use another directory. Because tasks are plain files, you can version `~/.agmd` with git to carry it between machines.

## Installation

### Quick Install

```bash
curl -fsSL https://gluongrid.dev/agmd/install.sh | bash
```

### From Source

```bash
go install github.com/GluonGrid/agmd@latest
```

### Manual Download

Download binaries from [Releases](https://github.com/GluonGrid/agmd/releases).

### Shell Completion

```bash
# Bash (add to ~/.bashrc)
eval "$(agmd completion bash)"

# Zsh (add to ~/.zshrc)
eval "$(agmd completion zsh)"

# Fish (add to ~/.config/fish/config.fish)
agmd completion fish | source
```

## License

MIT
