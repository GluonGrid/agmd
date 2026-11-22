# Agent Instructions

agmd is a CLI task board: `agmd task ...` manages markdown task files under
`~/.agmd/task/<project>/` (or `$AGMD_HOME/task/<project>/`), and `agmd task web`
serves a local board over the same files. Several agents may run agmd against
the same project at the same time, so every write must stay safe under
concurrent processes.

## Concurrency Rules

- Every read-modify-write of a project's files (task `.md` files,
  `assignments.json`, `project.json`, `comments/*.json`) must hold `lockProject(reg, project)`
  (`cmd/task_lock.go`). Never hold it while waiting for user input, and never
  call it while already holding it: flock is per open file, so a nested call
  in the same process blocks forever.
  A command that changes two projects takes both locks with
  `lockTwoProjects` (fixed name order, so it cannot deadlock).
- Write files with `writeFileAtomic` (replace) or `writeFileExclusive`
  (create-once), never `os.WriteFile`.
- `tests/integration/task_concurrency_test.go` runs a built binary as parallel
  subprocesses; add a case there when you add a new mutating command.

## Human Tasks

A task with `human: true` is a step only the human can do. Any code path that
hands work to an agent (`nextReadyTask`, `heldTask`, `task assign`, web runs)
must skip or refuse it, inside the same locked step that picks the task.

A human task may carry a `check` (shell command, exit 0 = done). Only
`task check` runs it, only for tasks flagged human (`runHumanCheck` refuses
others), without holding the project lock; the completion re-reads the task
under the lock. `createTask` and `task update`/`edit` refuse a check on a non-human
task. A replaced task is closed with `replaceTask` (`cmd/task_replace.go`),
which completes it, sets `replaced_by`/`replaces` and moves its dependents
in one locked step; never leave a replaced task open.

## Dependencies and Completion

- "Is this dependency done?" has one answer: `dependencySatisfied(dep, now)`
  (`cmd/task_reopen.go`). It is false for a missing task, an unfinished one,
  and a completion still inside its `--hold` window (`completion_counts_at`).
  Every reader that decides whether dependents are released (computed status,
  `next`/`--claim`, list filters, show, the web board) must go through it,
  via `computeTaskStatus`/`getPendingDependencies` or directly; never compare
  `Status == "completed"` for a dependency yourself.
- Completed is final once something was built on it. `setTaskStatus` still
  refuses to leave completed; `reopenTask` is the only way back and checks
  the dependents and writes under the same project lock.
- `saveTask` stamps `updated_at`; `createTask` sets `created_at`;
  `setTaskStatus` sets `completed_at`. Never invent timestamps for old tasks.

## Layout

- `cmd/task.go` — task model, storage, and all `agmd task` subcommands
- `cmd/task_next.go` — `task next [--claim]` (atomic claiming, stale-claim
  takeover) and `task heartbeat`
- `cmd/task_comment.go` — `task comment` (add, `--delete <n>`) and per-task
  comment storage
- `cmd/task_reopen.go` — `task reopen`, `status --hold`, and
  `dependencySatisfied`
- `cmd/task_update.go` — `task update` (text and fields in one locked step) and its
  compatibility alias `task edit`
- `cmd/task_dep.go` — `task dep add|remove` (cycle refusal) and the aliases
  `task blocked-by`/`task unblock`
- `cmd/task_replace.go` — `task replace` (close as replaced, move dependents)
- `cmd/task_check*.go` — `task check` (run checks of open human tasks)
- `cmd/task_lint.go` — `task lint` (unflagged `human-` tasks, replaced-but-open)
- `cmd/task_move.go` — `task move` (to another project; takes both project
  locks with `lockTwoProjects`, always in name order)
- `cmd/task_lock*.go` — per-project file lock and atomic file writes
- `cmd/task_project.go` — per-project metadata (`project.json`, emoji)
- `cmd/task_store.go` — task create/delete/claim-release and name validation,
  shared by the CLI and the web board
- `cmd/task_web.go` — web server, board state, patch/comment handlers
- `cmd/task_web_board.go` — create/delete/release endpoints, change feed,
  localhost-only guard
- `cmd/task_web_run.go` — Claude Code runs started from the board
- `cmd/webui/` — the board UI (plain HTML/CSS/JS, embedded; no build step):
  `ui.js` (App namespace, API client, icons, confirm/prompt dialogs),
  `board.js` (columns, cards, attention counts, keyboard, drag and drop,
  polling), `panel.js` (task panel), `runs.js` (run modal, console, runs list)
- `PRODUCT.md` — product record used for design work
- `cmd/task_completion.go` — shell completion for task names and flags
- `pkg/registry` — resolves the data directory (`AGMD_HOME` or `~/.agmd`)
- `tests/integration` — in-process CLI tests

## Code Quality

- Refactor in place; never create `-v2`/`-new` copies of files. Delete dead code
  instead of commenting it out.
- Keep files around 500–700 lines; extract cohesive pieces when they grow.
- Match the existing style and naming; prefer clarity over cleverness.
- Update `README.md` when user-facing behaviour changes.

## Integration Tests

### In-Process Testing
Tests call `cmd.ExecuteArgs(args)` directly — no subprocess, no binary build required.
Output is captured via `os.Pipe()` redirect of stdout/stderr.

### Key Helpers (tests/integration/helpers_test.go)
```go
run(t, "task", "list")        // run, return output (ignores error)
runE(t, "task", "list")       // run, return (output, error)
setup(t)                       // fresh AGMD_HOME, returns path
chdir(t)                       // temp CWD, auto-restored on cleanup
fileExists(path)               // uses os.Lstat — detects symlinks too
readFile(t, path)              // fatal if missing
```

### Flag Reset (cmd/root.go: ExecuteArgs)
Cobra does NOT reset flag values between Execute() calls.
`resetFlags(rootCmd)` is called before each ExecuteArgs to reset all
flags to their DefValue and clear f.Changed — prevents cross-test pollution
when global vars are shared between subcommands.

Because flags are package globals, `ExecuteArgs` must never be called from
several goroutines at once. Test concurrency either at the function level in
`cmd` (internal tests) or by running a built `agmd` binary as subprocesses.

### Commands Requiring --force
These commands prompt for confirmation; always pass --force in tests:
- `task delete <name> --force`
- `task clean --force`

### Commands Requiring EDITOR=true
`task new` without `--no-editor` or `--content` opens an editor; set
`t.Setenv("EDITOR", "true")` to avoid blocking.

## Operational Learnings

- `task.go`: priority was omitted from YAML when == 2 (default), causing P2 tasks
  to be read back as P0. Always write the priority field.
- `fileExists` must use os.Lstat, not os.Stat, so dangling symlinks are seen.
- Web server code (handlers and their goroutines) must not read CLI flag
  variables such as `taskStaleAfter`: they are package globals that the
  running command owns, and the in-process tests reset them while the server
  is still running. Use `envClaimTTL()` there, not `claimTTL()`. Run
  `go test -race ./...` before pushing; this race only showed up under `-race`.
