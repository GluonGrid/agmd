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
agmd task new create-api --content "Create endpoints"
agmd task dep add create-api setup-db
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
agmd task new create-api --content "Create endpoints"
agmd task new setup-db --feature auth     # Scope task to a feature
agmd task new merge-pr --human --content "Merge PR #12"   # A step only you can do
agmd task new merge-pr --human --check "gh pr view 12 --json state -q .state | grep -qx MERGED"

# List tasks (auto-sorted: P0→P4, then ready → blocked → completed)
agmd task list                            # All tasks for current project
agmd task list --type bug --priority 0    # Critical bugs only
agmd task list --feature auth             # Filter by feature
agmd task list --status ready             # Filter by computed status
agmd task list --tree                     # Show dependency tree
agmd task list --all                      # Include completed tasks
agmd task list --human                    # Only tasks for the human (--no-human hides them)
agmd task list --sort age                 # Oldest-created first (default: --sort priority)

# Manage status and dependencies
agmd task status setup-db completed       # Update status
agmd task status merge-pr completed --hold 2m   # Done now, releases dependents in 2 minutes
agmd task reopen merge-pr                 # Undo a completion nothing was built on yet
agmd task update setup-db --content "New text"   # Replace the description
agmd task update setup-db --content-file notes.md  # ... from a file (- reads stdin)
cat notes.md | agmd task update setup-db  # ... from stdin
agmd task update setup-db --subject "Set up the database"  # Change the subject
agmd task update setup-db --priority 0 --type bug  # Change priority/type
agmd task update setup-db --feature auth  # Change the feature (--feature "" clears it)
agmd task update merge-pr --human         # Mark for the human (--human=false hands it back)
agmd task update merge-pr --check "<cmd>" # Set a human task's check (--check "" clears it)
agmd task update ui --set run.profile=frontend --unset run.model  # Set/clear run fields
agmd task replace old-task --by new-task  # Close as replaced, move its dependents
agmd task move setup-db --to-project infra  # Move to another project (keeps id, comments)
agmd task comment setup-db --delete 2     # Delete comment #2 (yours; others need --force)
agmd task check                           # Complete open human tasks whose check passes
agmd task lint                            # Find unflagged human- tasks and replaced-but-open ones (--all: completed too)
agmd task dep add create-api setup-db     # Add dependency (refuses cycles)
agmd task dep remove create-api setup-db  # Remove dependency

# View and delete
agmd task show setup-db                   # Show task content
agmd task show --all                      # Show all tasks with content
agmd task delete setup-db --force         # Delete task
agmd task clean                           # Delete all completed tasks
```

`task update` changes everything in one locked step and keeps the task's id, dependencies, comments, claim and `created_at`; `updated_at` is set to now. A completed task is refused with `invalid_transition` unless you pass `--force`. `--json` prints `{"ok":true,"action":"update",...}` like the other commands.

**Compatibility aliases.** Older names still work and keep their JSON `action`: `task edit` = `task update` (it also changes completed tasks without `--force` and never reads stdin), `task blocked-by` = `task dep add`, `task unblock` = `task dep remove`, and `task new --blocked-by a,b` (or `--depends-on a,b`) sets dependencies at creation. Use that shorthand when the new task must never be ready without its dependencies: between `task new` and `task dep add` an agent could claim it.

**Priority levels:** P0 (critical), P1 (high), P2 (medium/default, hidden), P3 (low), P4 (backlog)
**Task types:** bug, feature, task (default, hidden), chore

### Human tasks

Some steps only you can do: merge a pull request, log in somewhere, run a root command, make a decision. Create them with `--human` on the same board. Agents are never handed a human task: `task next` and `next --claim` skip it (if only human tasks are ready, the answer is `not_found`), and `task assign` refuses it with `conflict`, even with `--force`. The list shows a `[human]` tag, `task show` a `for: human` line, and JSON output carries `"human": true`.

To make agent work wait for you, block it on the human task. Completing the human task releases it as usual:

```bash
agmd task new merge-pr --human -p 0 --content "Merge PR #12"
agmd task new deploy --blocked-by merge-pr --content "Deploy after the merge"   # blocked from the start
agmd task status merge-pr completed       # deploy becomes ready for agents
```

**Checked automatically.** A step is often done outside the board (you merged on GitHub, or ran the command), and nobody ticks the task. Give a human task a check, a shell command whose exit 0 means the step is done: `task new --human --check "<cmd>"` or `task update <task> --check "<cmd>"` (`--check ""` clears it). A check is refused on a task that is not for the human. `agmd task check [--project p]` runs the checks of all open human tasks with `sh -c` in the current folder (`AGMD_PROJECT` and `AGMD_TASK` are set), each killed after `--timeout` (default `30s`). A task whose check exits 0 is completed with the comment `verified automatically: <first output line>` (author `--author`, else `$AGMD_AUTHOR`, else `check`), which releases its dependents as usual. Stdout gets one line per completed task, `<project>/<task>`, and nothing else; `--verbose` reports the checks that did not pass on stderr, and `--json` lists both. A check stored on a task that is not flagged human is never run. `task show` prints the check.

```bash
agmd task new merge-pr --human --check "gh pr view 12 --json state -q .state | grep -qx MERGED"
agmd task check --project myproj          # prints "myproj/merge-pr" once the PR is merged
```

**Replaced tasks.** When a task is renamed, split or superseded, close it instead of leaving a comment: `agmd task replace <old> --by <new>` completes the old task, records `replaced_by` on it and `replaces` on the new one, moves tasks that depended on the old task to the new one (so they wait for the replacement), releases a claim on the old task and comments on both. It fails with `invalid_transition` if the old task is already completed and with `conflict` if moving a dependent would create a cycle. `task list` and `task show` print `replaced by <new>`; JSON carries `replaced_by` and `replaces`. Never leave a replaced task open.

**Run fields.** A task can carry free string settings for whoever starts a worker on it, for example the profile, model or effort. They live under `run:` in the front matter and agmd does not interpret them. `--set run.<key>=<value>` sets one (repeatable, on `task new` and `task update`; an empty value clears it) and `--unset run.<key>` clears one (on `task update`); all changes of one command are written in one locked step. Keys are lowercase letters, digits, `-` and `_`; values are one line. `task show` prints `run: key=value, ...`, and `show --json`/`list --json` carry them as the object `"run"` (left out when empty). Move, replace and reopen keep them.

```bash
agmd task new ui-polish --set run.profile=frontend --set run.effort=high
agmd task update ui-polish --set run.model=opus --unset run.effort
```

**Moving a task.** `agmd task move <task> --to-project <project>` moves a task that was filed in the wrong project. It keeps its name, id, text, fields (human flag, check, priority…), comments and `created_at`/`completed_at`; `updated_at` records the move and the task gets the comment `Moved from project '<from>'.` (`--author`, else `$AGMD_AUTHOR`, else `user`). The move holds the locks of both projects (taken in name order, so opposite moves cannot deadlock). It fails with `conflict` when the task has dependencies or dependents (they would cross projects; remove them with `task dep remove` first), when the target already has a task of that name or id, or while another workflow holds a claim on it; pass `--workflow-id` to move your own claim along. `--json` prints `{"ok":true,"action":"move","project":"<to>","from_project":"<from>",...}`. `task replace --by` still only accepts a task of the same project: move the replacement first if needed.

**Lint.** `agmd task lint [--project p] [--all]` lists tasks named `human-…` that are not flagged human (they are hidden from `list --human`; fix with `task update <task> --human`) and open tasks with a comment stating they were `replaced by <task>` or `superseded by <task>` (fix with `task replace`). `<task>` must be a task of the project, a `<project>/<task>` reference or a hyphenated name, and a plan or negation (“should be replaced by…”, “not replaced by…”) does not count. Only open tasks are listed; `--all` also checks completed ones. One line per finding, `<project>/<task> <kind>: <detail>`; `--json` for programs. The exit status is 0 either way.

### Undoing a completion

A completed task can't change status again, but it isn't final until something was built on it. `agmd task reopen <task>` puts it back to pending as long as no task that depends on it is claimed, in progress or completed; dependents that were only ready become blocked again. Otherwise it fails with `conflict`, naming the dependent. Chains follow from this: if completing `a` released `b` and `b` is completed, `a` can no longer be reopened. The reopen is recorded as a comment (`--author`, else `$AGMD_AUTHOR`, else `user`). Reopen is meant for the human (for example a mis-tap in an app); agents should not reopen tasks they did not complete.

To leave room for an undo, complete with `--hold <duration>` (for example `10s` or `2m`, at most `1h`). The task is completed at once, but its dependents stay blocked, and no agent can claim them, until the hold has passed (rounded up to the second); a reopen during the hold always succeeds. There is no timer: every reader compares the time when it looks. `task show` notes `counts from HH:MM:SS` and JSON output carries `hold_until` while the hold runs.

```bash
agmd task status merge-pr completed --hold 2m
agmd task reopen merge-pr --author app    # within 2 minutes: nobody has started deploy
```

Tasks are auto-sorted by priority then status. The `--tree` view groups tasks by feature and shows dependency chains. Dependencies are validated when they are added: the dependency must exist, and `task dep add` refuses a link that would make a task depend on itself, directly or through other tasks (`conflict`).

## Projects

The project is derived from the repository you run agmd in. In a Git worktree, agmd resolves to the main repository, so every worktree of a repo shares one task list. Override it with `--project <name>` or `--repo-path <path>`.

```bash
agmd task emoji 🚀                        # Show an emoji next to the project name
```

## Notes

`agmd note` reads and adds to the user's notes: one Markdown file per note at `~/.agmd/note/<project>/<slug>.md`, the same files as the Notes tab of the Agents app. The front matter carries `title`, `created`, `updated` (UTC) and an optional `tasks` list. Every write holds the one notes lock (flock on `~/.agmd/note/.lock`, the lock the app takes), so the app, agmd and other agents never overwrite each other. The project is resolved as for tasks (`--project`, `--repo-path`).

```bash
agmd note list                                 # slug, title, last updated (newest first)
agmd note list --all                           # the projects that have notes
agmd note show login-screen-inspiration        # the text, with the title as a heading
agmd note new --title "API naming" "first line" # prints the slug: api-naming
agmd note append api-naming -- "- [ ] check the docs"   # a TEXT starting with - goes after --
echo "from stdin" | agmd note append api-naming -
agmd note path api-naming                      # the file: read it, write only through agmd note
```

The slug is made from the title when the note is made and never changes: lowercase `a-z0-9` joined by `-`, at most 60 characters, `-2`, `-3` when the name is taken. agmd never deletes notes and never touches the app's pictures folder or `note/.trash/`.

## Agents

agmd is built for several agents working on the same project at once. Each agent picks a workflow id and claims work with one command:

```bash
agmd task next --claim --workflow-id agent-1 --json   # Claim the next ready task
agmd task comment setup-db "Schema migrated" --author agent-1   # Log progress
agmd task status setup-db completed                   # Finish it
agmd task unassign setup-db                           # Release the claim (back to pending if unfinished)
```

`task show` numbers comments from `#1` (oldest). `agmd task comment <task> --delete <n> --author <you>` deletes comment `n`; a comment by another author is refused with `conflict` unless you pass `--force` (for example to remove a personal detail an agent posted). A comment on a completed task is still stored (exit 0), with a warning on stderr that it is easily missed: reopen the task with `agmd task reopen`, or file a new task.

`next --claim` atomically picks the highest-priority agent task (never a [human task](#human-tasks)) that is ready and unassigned, assigns it to the workflow and marks it `in_progress`; two agents never get the same task. If the workflow already holds an unfinished task, that task is returned instead (`"action":"resume"`), so a restarted agent picks up where it left off. Without `--claim`, `agmd task next` only shows the next task.

`agmd task --help` is the whole agent reference in one screen (command matrix, update and dependency patterns, claiming, human tasks, JSON, exit codes, examples); an agent can read it once per session instead of each command's help.

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
| `invalid_transition` | 5 | Status change not allowed (leaving completed other than by `reopen`, or reopening a task that isn't completed) |
| `conflict` | 6 | Task already assigned to another workflow (`assign --force` takes it over), a task for the human, or a reopen refused because a dependent was already claimed, started or completed |

```bash
agmd task list --json --status ready
agmd task assign setup-db --workflow-id agent-1 --worktree-path ../wt-1
agmd task unassign --all --workflow-id agent-1
```

### Task list JSON

`agmd task list --json` is flat and stable: one object per task, nothing nested beyond what is shown. `depends_on` lists the references as stored (task names); `dependencies` lists the same tasks resolved to their canonical `id` and `name` (a reference that matches no task keeps its name and gets `"missing": true`, no id). `pending_dependencies` appears on blocked tasks only. Fields such as `feature`, `human`, `check` and `assignment` are left out when empty.

```json
{"project":"demo","tasks":[
  {"id":"tsk_eaf579d3d99c6f8b","name":"setup-db","project":"demo","subject":"Setup Db","status":"pending","computed_status":"ready","priority":2,"type":"task","depends_on":[],"dependencies":[]},
  {"id":"tsk_33cbcfa186886107","name":"create-api","project":"demo","subject":"Create Api","status":"pending","computed_status":"blocked","priority":2,"type":"task","depends_on":["setup-db"],"dependencies":[{"id":"tsk_eaf579d3d99c6f8b","name":"setup-db"}],"pending_dependencies":["setup-db"]}],
 "counts":{"total":2,"ready":1,"blocked":1,"in_progress":0,"completed":0}}
```

`agmd task list --json --tree` returns the same object plus `edges`, one per dependency of each listed task, for drawing the graph: `from` depends on `to`. `satisfied` says whether `to` is done; `to` may be absent from `tasks` (completed and hidden without `--all`, filtered out, or `missing`). Timestamps are left out of both examples.

```json
{"project":"demo","tasks":[...same as above...],"counts":{...},
 "edges":[{"from":"create-api","from_id":"tsk_33cbcfa186886107","to":"setup-db","to_id":"tsk_eaf579d3d99c6f8b","satisfied":false}]}
```

Concurrent use is safe: every change to a project's files happens under a per-project file lock (`<project>/.lock`), and files are replaced atomically, so the web board and any number of CLI agents can work side by side.

## Web board

```bash
agmd task web                             # Serve on a free port and open the browser
agmd task web --no-open                   # Serve without opening a browser
agmd task web --port 8765                 # Serve on a fixed port
```

`agmd task web` serves a local kanban board over the same task files (no build step, works offline):

- **Columns** follow each task's state: Blocked (waiting on dependencies), Ready, In progress and Done. Drag a card to change its status. Marking a task done asks first, because the board can't undo it (`agmd task reopen` can, until something was built on it); moving a task a live agent holds back to Ready asks before releasing its claim.
- **Agents at a glance.** Every card shows who holds it: a live agent (green, with its last heartbeat), a stale claim (amber, with its age), a Claude run started from the board (purple), or an in-progress task nobody has claimed. In progress lists stuck work first, and blocked cards name the task they wait on and flag it when that task is stuck.
- **What needs you.** The header counts open, working, needing-attention and blocked tasks; click a count to filter by it. Projects with stuck work carry an amber badge in the sidebar. The task panel offers one-click Re-queue for in-progress tasks nobody holds, and Release and re-queue for stale claims.
- **Human tasks** carry a blue *human* badge, and the board won't start a Claude run on them.
- **Manage tasks.** Create tasks (`n`), edit title, status, priority, type, feature, dependencies and description in the side panel, comment, and delete. Each field saves on its own, so an agent's concurrent change to another field is kept.
- **Keyboard.** `j`/`k`/`h`/`l` or the arrow keys move between cards, Enter opens one, `x` selects it for a run, `/` filters, `n` creates, Esc closes. Filters and the selected project live in the URL.
- **Live.** Changes made by agents through the CLI show up within a couple of seconds; if the agmd server stops, the board says so instead of showing stale data as live.
- **Light and dark** themes follow the system, or pick one in the sidebar.

The board can also **run tasks with Claude Code**: select cards (shift-click, `x`, or the corner checkbox) or open one, click Run, and pick a permission mode (it starts on `acceptEdits` and remembers your last choice; blocked and done tasks are left out). agmd starts a `claude` session with a `/goal` prompt, moves the tasks to in progress and claims them as workflow `web-<session-id>` (kept alive while the run is going). The run's transcript streams into a console, and every run stays available from **Runs** in the sidebar, with a `claude --resume <id>` command to continue it.

The server binds to `127.0.0.1` and refuses requests from other sites (foreign `Origin` or non-IP `Host`), since it can start processes that modify your files.

## Storage

Each task is a markdown file with YAML frontmatter at `~/.agmd/task/<project>/<name>.md`. The frontmatter carries `created_at`, `updated_at` and `completed_at` (RFC 3339, UTC; tasks created before these existed simply lack them), and `completion_counts_at` for a completion with `--hold`. Claims live in `<project>/assignments.json` and comments in `<project>/comments/<task-id>.json`. Set `AGMD_HOME` to use another directory. Because tasks are plain files, you can version `~/.agmd` with git to carry it between machines.

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
