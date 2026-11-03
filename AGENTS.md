# Agent Instructions

agmd is a CLI task board: `agmd task ...` manages markdown task files under
`~/.agmd/task/<project>/` (or `$AGMD_HOME/task/<project>/`), and `agmd task web`
serves a local board over the same files. Several agents may run agmd against
the same project at the same time, so every write must stay safe under
concurrent processes.

## Concurrency Rules

- Every read-modify-write of a project's files (task `.md` files,
  `assignments.json`, `project.json`) must hold `lockProject(reg, project)`
  (`cmd/task_lock.go`). Never hold it while waiting for user input.
- Write files with `writeFileAtomic` (replace) or `writeFileExclusive`
  (create-once), never `os.WriteFile`.
- `tests/integration/task_concurrency_test.go` runs a built binary as parallel
  subprocesses; add a case there when you add a new mutating command.

## Layout

- `cmd/task.go` — task model, storage, and all `agmd task` subcommands
- `cmd/task_next.go` — `task next [--claim]`, the atomic work-claiming primitive
- `cmd/task_lock*.go` — per-project file lock and atomic file writes
- `cmd/task_project.go` — per-project metadata (`project.json`, emoji)
- `cmd/task_web.go`, `cmd/webui/index.html` — web board and Claude run console
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
