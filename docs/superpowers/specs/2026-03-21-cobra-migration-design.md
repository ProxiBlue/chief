# Cobra CLI Migration Design

## Goal

Migrate the chief CLI from manual `os.Args` command routing to the Cobra framework for an extensible command structure. This is driven by upcoming commands (`login`, `logout`, `serve` for a server component) and the need for a scalable CLI architecture.

## Constraints

- Full backwards compatibility with all existing commands and flags
- Default behavior (no subcommand = TUI) must be preserved
- Existing business logic in `internal/cmd/` stays untouched

## Package Structure

```
cmd/chief/
├── main.go              # thin entry point: calls commands.Execute()
└── commands/
    ├── root.go          # root command (TUI default), persistent flags, wiggum
    ├── new.go           # chief new [name] [context]
    ├── edit.go          # chief edit [name]
    ├── status.go        # chief status [name]
    ├── list.go          # chief list
    ├── update.go        # chief update
    └── serve.go         # chief serve (placeholder for future)
```

Each command file defines a `cobra.Command` and registers it via `init()`. Commands are thin wrappers that parse flags/args and delegate to existing `internal/cmd` functions.

## Root Command

The root command launches the TUI when no subcommand is given:

```go
var rootCmd = &cobra.Command{
    Use:   "chief [prd-name|path]",
    Short: "AI-powered PRD execution engine",
    RunE:  runTUI,
}
```

### Persistent Flags (global, all commands)

| Flag | Type | Description |
|------|------|-------------|
| `--agent <provider>` | string | Agent provider (claude, codex, opencode, cursor) |
| `--agent-path <path>` | string | Custom path to agent CLI binary |
| `--verbose` | bool | Show raw agent output |

### Local Flags (root/TUI only)

| Flag | Type | Description |
|------|------|-------------|
| `--max-iterations` / `-n` | int | Iteration limit |
| `--no-retry` | bool | Disable auto-retry on crashes |
| `--merge` | bool | Auto-merge progress on conflicts |
| `--force` | bool | Auto-overwrite on conflicts |

### Version

Handled via Cobra's built-in `rootCmd.Version = version`, supporting both `--version` and `-v`.

### Environment Variable Fallbacks

`--agent` and `--agent-path` fall back to `CHIEF_AGENT` and `CHIEF_AGENT_PATH` respectively. This is already handled in `internal/agent/resolve.go` — no changes needed.

## Subcommands

### `new` — Create a new PRD

```
chief new [name] [context]
Args: cobra.MaximumNArgs(2)
```

Extracts positional args, calls `cmd.RunNew()`.

### `edit` — Edit an existing PRD

```
chief edit [name]
Args: cobra.MaximumNArgs(1)
```

### `status` — Show PRD progress

```
chief status [name]
Args: cobra.MaximumNArgs(1)
```

Defaults to `"main"` if no name given.

### `list` — List all PRDs

```
chief list
Args: cobra.NoArgs
```

### `update` — Self-update

```
chief update
Args: cobra.NoArgs
```

### `wiggum` — Easter egg (hidden)

Registered inline in `root.go` with `Hidden: true`.

## What Changes

| Component | Change |
|-----------|--------|
| `cmd/chief/main.go` | Gutted to just `commands.Execute()` |
| `cmd/chief/commands/` | New package (~7 files) |
| `go.mod` | Add `github.com/spf13/cobra` |
| `parseTUIFlags()` | Deleted (Cobra handles flag parsing) |

## What Stays the Same

- `internal/cmd/` — all command handler logic
- `internal/agent/` — provider resolution
- `internal/config/` — configuration loading
- `internal/loop/` — agent orchestration
- `internal/tui/` — Bubble Tea UI
- `internal/prd/` — PRD model and persistence
- `internal/git/` — git utilities
- `internal/update/` — self-update mechanism

## Free Wins from Cobra

- `chief completion bash|zsh|fish|powershell` — shell completions with no extra work
- Auto-generated help text for all commands and flags
- Consistent error messages for invalid args/flags
- Built-in `--help` / `-h` on every command

## Testing

The Cobra layer is thin — most coverage comes from existing `internal/cmd` tests. A lightweight test in `commands/` can verify subcommand registration and flag definitions.
