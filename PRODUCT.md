# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

The primary user is the author: a developer who runs several coding agents (Claude Code, Codex and similar) on their own repositories at the same time, often one agent per git worktree. agmd is tuned for that person first; it is published (GluonGrid/agmd, install script, releases) so other developers in the same situation can use it later, but no outside audience has been designed for yet.

There are two kinds of user of the same data:

- **Agents** use the CLI (`agmd task ...`, usually with `--json`) to claim work, report progress and finish tasks. They never see the board.
- **The human** uses the web board (`agmd task web`) and the CLI to plan and supervise.

## Product Purpose

agmd is a task board shared by a developer and their coding agents. It answers three questions at any moment: what should be worked on next, who (which agent) is working on what, and what is stuck.

The board serves two jobs equally:

- **Planning:** writing tasks, setting priority, type, feature and dependencies so agents pick up the right work in the right order.
- **Supervising:** seeing which agent holds which task and when it last checked in, spotting stale or orphaned work, and stepping in (release a claim, re-queue, comment, run a task with Claude Code).

Success means agents working in parallel never collide or pick up blocked work, and the human can tell at a glance what is happening without asking an agent.

## Positioning

- **Plain files you own.** Each task is a markdown file with YAML frontmatter in `~/.agmd/task/<project>/`; claims, comments and project metadata are small JSON files beside them. No database, no server process to keep alive, no account. The data diffs and syncs with git.
- **Built for parallel agents.** Atomic claiming (`task next --claim`), per-project file locking, heartbeats, stale-claim expiry and takeover, and workflow ids mean many agents can share one project safely.
- **Dependency-aware planning.** Priority (P0–P4), type, feature and `depends_on` decide what is ready; blocked tasks are never handed out.
- **CLI-first, not MCP.** Agents use a small shell CLI with JSON output and coded exit statuses. The web board is for the human.

## Operating Context

- Runs locally on the developer's machine. The board is served by the same binary on `127.0.0.1` and refuses cross-site requests, because it can start Claude Code runs that modify files.
- The project is derived from the git repository the command runs in; worktrees resolve to the main repository so every worktree shares one task list.
- Agents are instructed through a global skill (`~/.claude/skills/agmd`) to use agmd as their default task manager.
- The board updates live (polling a change feed) while agents work through the CLI; the human typically keeps it open beside their terminals.
- The board can start a Claude Code session for selected tasks and stream its transcript.

## Capabilities and Constraints

- Statuses are `pending`, `in_progress` and `completed`; the board shows computed columns Blocked, Ready, In progress and Done. Completed is final and cannot be reopened.
- Claims belong to a workflow id; a claim is active, stale (no heartbeat within the TTL, default 2h), held by a board-started run (`web-<session>`), or absent. An in-progress task with no claim is a known state the board must make visible.
- Tasks carry a name (kebab-case slug), subject, description (Markdown), priority, type (bug, feature, task, chore), feature, dependencies and comments. Tasks do not record created or updated timestamps.
- The CLI cannot edit a task's subject or description after creation; the board can.
- The board is plain HTML, CSS and JavaScript embedded in the Go binary (`cmd/webui/`), with no build step and no external requests. It must keep working offline.
- Several projects can exist; the board has an all-projects view.
- Terminology: task, project, feature, claim, workflow id, heartbeat, stale claim, run.

## Evidence on Hand

- README.md documents the CLI, agent workflow, board and storage.
- No users other than the author, no testimonials, adoption numbers, benchmarks or press exist. Do not invent any.

## Product Principles

1. **The files are the truth.** Every view, CLI or board, reads and writes the same plain files; nothing important lives only in the UI or in memory.
2. **Never let agents collide.** Any new way to change tasks must be safe under concurrent agents (locked, atomic) before it ships.
3. **Show who is doing what.** Claim state (active, stale, run, orphaned) is first-class information wherever tasks appear.
4. **Agents get terse, humans get clarity.** CLI output is short and machine-readable; the board explains state in plain language.
5. **Local and private by default.** No accounts, no network services, no telemetry.
