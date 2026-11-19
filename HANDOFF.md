# Handoff: agmd

Written 2026-11-19 for someone new to the project. It covers what agmd is, where the work stands, why it is built this way, and what to do next. For day-to-day coding rules, read [AGENTS.md](AGENTS.md); for user-facing documentation, read [README.md](README.md); for product intent, read [PRODUCT.md](PRODUCT.md).

## In one paragraph

agmd is a task board for a developer who runs several AI coding agents (Claude Code, Codex…) on the same repositories at once. Tasks are plain markdown files under `~/.agmd/task/<project>/`. Agents use the `agmd task …` CLI to claim work atomically, post progress comments and finish tasks; a human plans and supervises, originally through a local kanban web board (`agmd task web`). It is a single Go binary with no database and no account. The owner uses it as a personal tool first; it is published at [GluonGrid/agmd](https://github.com/GluonGrid/agmd) for other people later.

## Goal

Let many agents work the same project in parallel without colliding, and let the human see at a glance what is being worked on, by whom, and what is stuck. The principles behind every decision are in [PRODUCT.md](PRODUCT.md#product-principles): the files are the truth, agents must never collide, show who is doing what, terse output for agents and clear explanations for humans, local and private by default.

## Current state

### The code (this repo)

All work is on the branch `feat/tasks-only`, in open pull request [GluonGrid/agmd#1](https://github.com/GluonGrid/agmd/pull/1) against `main`. It is not merged. The working tree is clean apart from the untracked `.impeccable/` folder (see [Loose ends](#loose-ends)).

What the branch contains, oldest first:

| Commit | What it did |
| --- | --- |
| `57708ef`–`ea1a3f0` (5 commits, also on `feat/interactive-cli`) | Stable task ids, `--json` output with coded errors, assign/unassign, worktree-aware project names, the first web UI |
| `2651d99` | Removed the old product (see [History](#history)): ~3,170 files, 1.07M lines |
| `1cf50c9` | Per-project file lock, atomic writes, `task next --claim` |
| `cc1f94b` | `task comment` |
| `459117b` | Stale-claim expiry and `task heartbeat` |
| `af8c26f` | Blocked tasks are never handed out; `unassign` re-queues; one-line JSON errors |
| `1c40d81` | `.gitignore` cleanup |
| `b00620f` | Kanban web board with live claim status |
| `6395d1e` | Fix: `task show` reported finished dependencies as pending |
| `171a812` | PRODUCT.md |
| `c24bafa`, `c466add` | Board fixes from a design critique: stuck-work counts, confirmations, keyboard access, contrast, focus |
| `dd77da6` | This handoff |
| `ff68158` | Fix a data race: web server goroutines read a CLI flag global (found by `go test -race`) |
| (PR against `feat/tasks-only`) | Human tasks: a `human: true` field (`task new --human`, `task edit --human[=false]`, `task edit --feature`), never handed out by `next`/`assign`/board runs, `list --human`/`--no-human`, a badge on the board. Replaces the old `human-` name prefix and wrapper script. Also fixes a race in the run response |

Tests: `go test ./...` passes, including with `-race`. There is no CI; run the tests locally (see [How to build, run and test](#how-to-build-run-and-test)).

### How it is deployed today (outside this repo)

Since 11 November 2026 the live board does **not** run on the Mac. It moved to an always-on homelab container, **CT 1000**, and is reached through a separate project, the `agents` repo (`~/git/agents`, remote `gitlab.com:lucas.zweili/agents`). The design and its reasons are in that repo's `docs/designs/app-next.md`. In short:

```
Mac: Claude sessions ─ agmd (Python shim) ─┐
                                           ├─ HTTP on the tailnet ─── ask-service ─ agmd CLI ─ ~/.agmd/task
phone and desktop "Agents" app ────────────┘                          (CT 1000)    (CT 1000)
CT 1000: orchestrator and agent-run workers ─ agmd CLI (local) ─────────────────────────────┘
```

- **CT 1000** runs this repo's binary as the only writer of `~/.agmd/task`. One writing machine is what makes agmd's file lock correct (see [Decisions](#decisions-and-why)).
- **ask-service** (`agents/trial/bin/ask-service`) exposes `POST /app/api/agmd` and runs `agmd task <args> --project <p>` on CT 1000, returning stdout, stderr and the exit code. It never starts `agmd task web`, because that server has no login.
- **On the Mac**, `~/.local/bin/agmd` is a Python shim (`agents/tools/agmd-remote`, installed by `agents/tools/install-agmd-remote`). It sends `agmd task …` to ask-service and runs everything else (help, version) with the real binary, `~/.local/bin/agmd-local`. On 19 Nov `agmd-local` was byte-identical to a build of `feat/tasks-only`. The shim refuses `task web`, `delete` and `clean`, and adds exit code `7 unavailable` when CT 1000 can't be reached. Config: `~/.config/agmd-remote/config.json` (holds a token; never commit or share it).
- **The board humans use is now the Agents app** (Tasks tab, phone and desktop), built in the `agents` repo. The kanban board in this repo (`cmd/webui/`) still works locally, but the owner's setup no longer uses it.
- **Agents learn the workflow** from a global Claude Code skill at `~/.claude/skills/agmd/SKILL.md` (also linked into `~/.agents/skills/agmd`). It is not in this repo. It tells agents to pick a workflow id, `next --claim`, comment, heartbeat, finish or `unassign`, and not to edit task files by hand. It was updated for the CT 1000 setup.

Not verified here: how the binary on CT 1000 gets updated when this repo changes. Ask the owner, or check the `agents` repo's deploy scripts (`trial/bin/agents-deploy`).

## How it works

- **A task** is `~/.agmd/task/<project>/<name>.md`: YAML frontmatter (`id`, `subject`, `status`, `priority`, `type`, `feature`, `depends_on`, optional `human`) plus a Markdown description. `AGMD_HOME` moves the whole data folder.
- **Status** is stored as `pending`, `in_progress` or `completed`. The board shows computed columns: Blocked (pending with unfinished dependencies), Ready, In progress, Done. Completed is final.
- **The project** is the main repository folder's name, from `git rev-parse --git-common-dir`, so all worktrees of a repo share one list. Override with `--project` or `--repo-path`.
- **Claims** live in `<project>/assignments.json`: task id → workflow id, worktree path, last heartbeat. `next --claim` picks the highest-priority ready, unclaimed task and claims it in one locked step. A claim with no sign of life for `AGMD_CLAIM_TTL` (default 2h) can be taken over; claiming, `heartbeat` and a comment by the claiming workflow all count as signs of life.
- **Comments** live in `<project>/comments/<task-id>.json`.
- **Concurrency:** every read-modify-write holds `<project>/.lock` (`flock`), and every file is written with temp-file-and-rename (`cmd/task_lock.go`). Rules for contributors are in [AGENTS.md](AGENTS.md#concurrency-rules).
- **Errors:** with `--json`, failures print one JSON line on stderr, with exit codes 2 `invalid_input`, 3 `not_found`, 4 `project_mismatch`, 5 `invalid_transition`, 6 `conflict`. The Mac shim adds 7 `unavailable`.
- **The web board** (`agmd task web`) is embedded HTML/CSS/JS in `cmd/webui/` with no build step. It polls `GET /api/changes` and reloads when task files change. It can also start Claude Code runs for selected tasks (`cmd/task_web_run.go`).

A map of the source files is in [AGENTS.md](AGENTS.md#layout).

## Decisions and why

| Decision | Why |
| --- | --- |
| Strip agmd down to tasks only | It started as an AGENTS.md generator; the owner only kept using the task board, and the rest was 1M lines of upkeep. |
| CLI for agents, not MCP | The owner compared it with MCP-based boards (e.g. kanzo) and preferred a small shell CLI with JSON output: any agent can use it, and workers can run without MCP. |
| Plain files, no database | Tasks are diffable, greppable and syncable with git, and there is no server process to keep alive. |
| Per-project `flock` plus atomic writes | Several agents run `agmd` at the same moment. Subprocess tests in `tests/integration/task_concurrency_test.go` lose up to 8 of 12 updates without the lock. |
| `task next --claim` as the main primitive | One atomic step to pick and claim work, so two agents never get the same task. It resumes a task the workflow already holds. |
| Claims expire (2h) with heartbeats | A crashed agent must not hold a task forever; a takeover is recorded as a comment. |
| Completed is final | Keeps history honest and dependents stable. The board asks before completing. |
| One writing machine (CT 1000) | `flock` only excludes processes on the same machine, and file sync (Syncthing) creates conflict copies agmd would read as duplicate tasks. Decided in the `agents` repo. |
| The web server stays on localhost | It has no login and can start Claude with file access; it refuses foreign `Origin` headers and non-IP `Host` names to block cross-site and DNS-rebinding attacks. |
| Board columns stay status-based | The owner chose header counts and sorting to surface stuck work instead of a separate "Needs you" lane. |

## How to build, run and test

You need Go 1.25 or newer (`go.mod` says 1.25.3) and git. Use a scratch data folder so you never touch real tasks:

```bash
export AGMD_HOME="$(mktemp -d)"
```

Build:

```bash
go build -o agmd .
```

Try the agent workflow from this repo's folder (the project is `agent-md`, stored in your scratch `AGMD_HOME`):

```bash
./agmd task new setup-db --content "Create the schema"
./agmd task new create-api --blocked-by setup-db --content "Endpoints"
./agmd task next --claim --workflow-id me --json
./agmd task comment setup-db "Schema drafted" --author me
./agmd task status setup-db completed
./agmd task list --all
```

Run the board locally (it binds to 127.0.0.1):

```bash
./agmd task web --no-open --port 8765
```

To try "Run with Claude Code" without starting a real Claude session, point `AGMD_CLAUDE_BIN` at a stub script that just sleeps.

Run the tests (the concurrency tests build the binary themselves):

```bash
go test ./...
```

```bash
go test -race -count=1 ./...
```

Release: `scripts/release.sh v0.2.0` cross-compiles darwin and linux binaries into `dist/`. `install.sh` downloads the latest GitHub release. Nothing has been released since the rewrite, and the version string is still `0.1.0` (`cmd/root.go`).

## Branches and pull requests

| Branch | Where | State | What to do |
| --- | --- | --- | --- |
| `feat/tasks-only` | local and origin | All current work; open PR [#1](https://github.com/GluonGrid/agmd/pull/1) | Review and merge into `main` |
| `feat/interactive-cli` | local and origin | Its 5 commits are included in PR #1 | Delete after PR #1 merges |
| `main` | local and origin | The old instruction-registry product (Feb 2026) | Gets everything when PR #1 merges |
| `gh-pages` | origin only | Last updated 2026-02-27; site for the old product | Update or retire |
| `claude/laughing-kilby-979d7e` | local only | Same commit as `main`; left over from an old agent session | Safe to delete |

## Next steps

In rough priority order:

1. **Merge PR #1**, then delete `feat/interactive-cli` and `claude/laughing-kilby-979d7e`. The PR description predates PRODUCT.md and the board fixes; add a line about them.
2. **Change the server's default run mode.** If an API caller leaves `permissionMode` out of `POST /api/run`, the server still uses `bypassPermissions` (`cmd/task_web_run.go:319`). The UI now sends `acceptEdits` by default, but the server default should match.
3. **Decide how CT 1000 picks up new versions** of this repo, write it down here, and cut a release (`v0.2.0`) so `install.sh` serves the task-only CLI.
4. **Add CI** (GitHub Actions running `go test -race ./...`); there is none today.
5. **Bring the agent skill into the repo.** `~/.claude/skills/agmd/SKILL.md` documents the agent workflow but only exists on the owner's machine. Commit a copy (for example under `skills/agmd/`) and keep the CT 1000 notes out of the generic version.
6. **CLI gaps:** `task edit` can change priority, type, feature and the human flag, but not the title or description. Tasks have no created or updated timestamps.
7. **Board gaps** from the last design pass, if the local board stays in use: the list of Claude runs can't be opened on phones; the run chip on a card can't be reached with Tab; no screen-reader test has been done. A fresh `/impeccable critique` would re-score it (the last score was 23/40, before the fixes).
8. **Windows** doesn't build because of the `Setpgid` call in `cmd/task_web_run.go`. Windows isn't a release target.
9. **A takeover doesn't stop the old agent.** Only `task heartbeat <task>` tells it the claim was lost, and its `task status completed` still succeeds.

## Loose ends

- `.impeccable/` (untracked) holds the design-critique history used by the Impeccable design tool. It is neither committed nor ignored; decide which.
- `~/.agmd/` on the Mac still contains the old product's folders (`rule`, `guide`, `profile`, `doc`, …). They are unused; the owner hasn't decided whether to delete them. The task data moved to CT 1000.
- The previous Go binaries removed from `~/.local/bin` are in the Trash (`agmd-old-*`).

## History

Until November 2026, agmd was a different product: a registry of reusable AI instructions (rules, workflows, profiles) that generated `AGENTS.md` from a `directives.md` file. The task board started as one subcommand of it. In November the owner decided the task board was the only part worth keeping, compared it with [kanzo](https://github.com/emanuele-em/kanzo) (an MCP-based kanban for agents), and chose a CLI-first, file-based design built for parallel agents. `main` still holds the old product until PR #1 merges.
