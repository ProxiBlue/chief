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

Note: `--merge` and `--force` are also registered as local flags on the `edit` subcommand (the current help text advertises them for edit). The `edit` command handler will accept these flags even though `internal/cmd/edit.go` does not currently use them — this preserves backwards compatibility with scripts that may pass them.

### Version

Handled via Cobra's built-in `rootCmd.Version = version`, supporting `--version`. The `-v` shorthand for version is explicitly disabled to avoid conflict with a potential future `-v` shorthand for `--verbose`. `--verbose` has no short form.

### Environment Variable Fallbacks

`--agent` and `--agent-path` fall back to `CHIEF_AGENT` and `CHIEF_AGENT_PATH` respectively. This is already handled in `internal/agent/resolve.go` — no changes needed.

## Subcommands

### `new` — Create a new PRD

```
chief new [name] [context...]
Args: cobra.ArbitraryArgs
```

Uses `cobra.ArbitraryArgs` because the current CLI allows unquoted multi-word context: `chief new auth JWT authentication for REST API` joins all args after the first with spaces. The handler extracts `args[0]` as name and `strings.Join(args[1:], " ")` as context.

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

## Implementation Notes

### `chief help` subcommand

Cobra handles `chief help` and `chief help <command>` automatically — no custom registration needed.

### `CheckVersionOnStartup`

Called only in `runTUI` (not `PersistentPreRun`), preserving the current behavior of only checking for updates when launching the TUI.

### Provider resolution

A shared `resolveProvider(cmd)` helper in `root.go` reads the persistent flags and calls `agent.Resolve()`. Used by `runTUI`, `new`, and `edit` commands.

### PRD path resolution

The `runTUI` function in `root.go` carries the full PRD resolution and first-time-setup flow currently in `runTUIWithOptions()` — this function is migrated largely intact.

### Flag validation

`--max-iterations` is validated `>= 1` in a `PreRunE` on the root command, matching current behavior.

## Free Wins from Cobra

- `chief completion bash|zsh|fish|powershell` — shell completions with no extra work
- Auto-generated help text for all commands and flags
- Consistent error messages for invalid args/flags
- Built-in `--help` / `-h` on every command

## Testing

The Cobra layer is thin — most coverage comes from existing `internal/cmd` tests. A lightweight test in `commands/` can verify subcommand registration and flag definitions.
