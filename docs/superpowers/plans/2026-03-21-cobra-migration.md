# Cobra CLI Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace manual `os.Args` command routing in `cmd/chief/main.go` with the Cobra framework for an extensible CLI structure.

**Architecture:** New `cmd/chief/commands/` package with one file per command. Each command is a thin Cobra wrapper delegating to existing `internal/cmd` handlers. Root command launches TUI by default. `main.go` becomes a one-liner calling `commands.Execute()`.

**Tech Stack:** Go 1.24, github.com/spf13/cobra

**Spec:** `docs/superpowers/specs/2026-03-21-cobra-migration-design.md`

---

## File Structure

```
cmd/chief/
├── main.go                    # MODIFY: gut to thin entry point
└── commands/
    ├── root.go                # CREATE: root command, persistent flags, TUI logic, wiggum, helpers
    ├── root_test.go           # CREATE: test flag/command registration
    ├── new.go                 # CREATE: chief new
    ├── edit.go                # CREATE: chief edit
    ├── status.go              # CREATE: chief status
    ├── list.go                # CREATE: chief list
    └── update.go              # CREATE: chief update
```

---

### Task 1: Add Cobra dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add cobra dependency**

Run: `cd /Users/codemonkey/projects/chief && go get github.com/spf13/cobra@latest`

- [ ] **Step 2: Tidy modules**

Run: `go mod tidy`

- [ ] **Step 3: Verify dependency added**

Run: `grep cobra go.mod`
Expected: `github.com/spf13/cobra` appears in require block

- [ ] **Step 4: Verify project compiles**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add spf13/cobra for CLI framework migration"
```

---

### Task 2: Create root command with persistent flags

**Files:**
- Create: `cmd/chief/commands/root.go`

- [ ] **Step 1: Create the commands package with root command**

Create `cmd/chief/commands/root.go`:

```go
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/minicodemonkey/chief/internal/agent"
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/minicodemonkey/chief/internal/config"
	"github.com/minicodemonkey/chief/internal/git"
	"github.com/minicodemonkey/chief/internal/loop"
	"github.com/minicodemonkey/chief/internal/prd"
	"github.com/minicodemonkey/chief/internal/tui"
	"github.com/spf13/cobra"
)

var (
	// Version is set from main.go
	Version = "dev"

	// Persistent flags (global)
	flagAgent     string
	flagAgentPath string
	flagVerbose   bool

	// Local flags (root/TUI only)
	flagMaxIterations int
	flagNoRetry       bool
	flagMerge         bool
	flagForce         bool
)

var rootCmd = &cobra.Command{
	Use:   "chief [prd-name|path]",
	Short: "Chief - Autonomous PRD Agent",
	Long:  "Chief orchestrates AI agents to implement PRDs (Product Requirements Documents) autonomously.",
	RunE:  runTUI,
	// Allow the root command to accept unknown flags that look like positional args
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVar(&flagAgent, "agent", "", "Agent CLI to use: claude (default), codex, opencode, or cursor")
	rootCmd.PersistentFlags().StringVar(&flagAgentPath, "agent-path", "", "Custom path to agent CLI binary")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Show raw agent output in log")

	// Local flags (root/TUI only)
	rootCmd.Flags().IntVarP(&flagMaxIterations, "max-iterations", "n", 0, "Set maximum iterations (default: dynamic)")
	rootCmd.Flags().BoolVar(&flagNoRetry, "no-retry", false, "Disable auto-retry on agent crashes")
	rootCmd.Flags().BoolVar(&flagMerge, "merge", false, "Auto-merge progress on conversion conflicts")
	rootCmd.Flags().BoolVar(&flagForce, "force", false, "Auto-overwrite on conversion conflicts")

	// Wiggum easter egg
	rootCmd.AddCommand(&cobra.Command{
		Use:    "wiggum",
		Hidden: true,
		Run: func(c *cobra.Command, args []string) {
			printWiggum()
		},
	})
}

// Execute is the main entry point called from main.go
func Execute(version string) {
	Version = version
	rootCmd.Version = version

	// Create the --version flag, then disable the -v shorthand to avoid
	// conflict with potential future --verbose shorthand
	rootCmd.InitDefaultVersionFlag()
	rootCmd.Flags().Lookup("version").Shorthand = ""

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveProvider loads config and resolves the agent provider.
func resolveProvider() (loop.Provider, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to load .chief/config.yaml: %w", err)
	}
	provider, err := agent.Resolve(flagAgent, flagAgentPath, cfg)
	if err != nil {
		return nil, err
	}
	if err := agent.CheckInstalled(provider); err != nil {
		return nil, err
	}
	return provider, nil
}

// findAvailablePRD looks for any available PRD in .chief/prds/
func findAvailablePRD() string {
	prdsDir := ".chief/prds"
	entries, err := os.ReadDir(prdsDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			prdPath := filepath.Join(prdsDir, entry.Name(), "prd.md")
			if _, err := os.Stat(prdPath); err == nil {
				return prdPath
			}
		}
	}
	return ""
}

// listAvailablePRDs returns all PRD names in .chief/prds/
func listAvailablePRDs() []string {
	prdsDir := ".chief/prds"
	entries, err := os.ReadDir(prdsDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			prdPath := filepath.Join(prdsDir, entry.Name(), "prd.md")
			if _, err := os.Stat(prdPath); err == nil {
				names = append(names, entry.Name())
			}
		}
	}
	return names
}

func runTUI(command *cobra.Command, args []string) error {
	// Validate --max-iterations (0 means dynamic/unset, negative is invalid)
	if flagMaxIterations < 0 {
		return fmt.Errorf("--max-iterations must be at least 1")
	}

	// Non-blocking version check on startup
	cmd.CheckVersionOnStartup(Version)

	provider, err := resolveProvider()
	if err != nil {
		return err
	}

	// Resolve PRD path from positional arg
	var prdPath string
	if len(args) > 0 {
		arg := args[0]
		if strings.HasSuffix(arg, ".md") || strings.HasSuffix(arg, ".json") || strings.HasSuffix(arg, "/") {
			prdPath = arg
		} else {
			prdPath = fmt.Sprintf(".chief/prds/%s/prd.md", arg)
		}
	}

	return runTUIWithOptions(prdPath, provider)
}

func runTUIWithOptions(prdPath string, provider loop.Provider) error {
	// If no PRD specified, try to find one
	if prdPath == "" {
		mainPath := ".chief/prds/main/prd.md"
		if _, err := os.Stat(mainPath); err == nil {
			prdPath = mainPath
		} else {
			prdPath = findAvailablePRD()
		}

		// If still no PRD found, run first-time setup
		if prdPath == "" {
			cwd, _ := os.Getwd()
			showGitignore := git.IsGitRepo(cwd) && !git.IsChiefIgnored(cwd)

			result, err := tui.RunFirstTimeSetup(cwd, showGitignore)
			if err != nil {
				return err
			}
			if result.Cancelled {
				return nil
			}

			cfg := config.Default()
			cfg.OnComplete.Push = result.PushOnComplete
			cfg.OnComplete.CreatePR = result.CreatePROnComplete
			if err := config.Save(cwd, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
			}

			newOpts := cmd.NewOptions{
				Name:     result.PRDName,
				Provider: provider,
			}
			if err := cmd.RunNew(newOpts); err != nil {
				return err
			}

			newPath := fmt.Sprintf(".chief/prds/%s/prd.md", result.PRDName)
			return runTUIWithOptions(newPath, provider)
		}
	}

	prdDir := filepath.Dir(prdPath)

	// Auto-migrate: if prd.json exists alongside prd.md, migrate status
	jsonPath := filepath.Join(prdDir, "prd.json")
	if _, err := os.Stat(jsonPath); err == nil {
		fmt.Println("Migrating status from prd.json to prd.md...")
		if err := prd.MigrateFromJSON(prdDir); err != nil {
			fmt.Printf("Warning: migration failed: %v\n", err)
		} else {
			fmt.Println("Migration complete (prd.json renamed to prd.json.bak).")
		}
	}

	app, err := tui.NewAppWithOptions(prdPath, flagMaxIterations, provider)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			fmt.Printf("PRD not found: %s\n", prdPath)
			fmt.Println()
			available := listAvailablePRDs()
			if len(available) > 0 {
				fmt.Println("Available PRDs:")
				for _, name := range available {
					fmt.Printf("  chief %s\n", name)
				}
				fmt.Println()
			}
			fmt.Println("Or create a new one:")
			fmt.Println("  chief new               # Create default PRD")
			fmt.Println("  chief new <name>        # Create named PRD")
			os.Exit(1)
		}
		return err
	}

	if flagVerbose {
		app.SetVerbose(true)
	}
	if flagNoRetry {
		app.DisableRetry()
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	// Handle post-exit actions
	if finalApp, ok := model.(tui.App); ok {
		switch finalApp.PostExitAction {
		case tui.PostExitInit:
			newOpts := cmd.NewOptions{
				Name:     finalApp.PostExitPRD,
				Provider: provider,
			}
			if err := cmd.RunNew(newOpts); err != nil {
				return err
			}
			newPath := fmt.Sprintf(".chief/prds/%s/prd.md", finalApp.PostExitPRD)
			return runTUIWithOptions(newPath, provider)

		case tui.PostExitEdit:
			editOpts := cmd.EditOptions{
				Name:     finalApp.PostExitPRD,
				Provider: provider,
			}
			if err := cmd.RunEdit(editOpts); err != nil {
				return err
			}
			editPath := fmt.Sprintf(".chief/prds/%s/prd.md", finalApp.PostExitPRD)
			return runTUIWithOptions(editPath, provider)
		}
	}

	return nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/chief/commands/...`
Expected: no errors (package compiles but isn't wired up yet)

- [ ] **Step 3: Commit**

```bash
git add cmd/chief/commands/root.go
git commit -m "feat: add Cobra root command with persistent flags and TUI logic"
```

---

### Task 3: Create subcommand files

**Files:**
- Create: `cmd/chief/commands/new.go`
- Create: `cmd/chief/commands/edit.go`
- Create: `cmd/chief/commands/status.go`
- Create: `cmd/chief/commands/list.go`
- Create: `cmd/chief/commands/update.go`

- [ ] **Step 1: Create `new.go`**

```go
package commands

import (
	"strings"

	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name] [context...]",
	Short: "Create a new PRD interactively",
	Args:  cobra.ArbitraryArgs,
	RunE: func(c *cobra.Command, args []string) error {
		opts := cmd.NewOptions{}
		if len(args) > 0 {
			opts.Name = args[0]
		}
		if len(args) > 1 {
			opts.Context = strings.Join(args[1:], " ")
		}

		provider, err := resolveProvider()
		if err != nil {
			return err
		}
		opts.Provider = provider

		return cmd.RunNew(opts)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
```

- [ ] **Step 2: Create `edit.go`**

```go
package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit an existing PRD interactively",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		opts := cmd.EditOptions{}
		if len(args) > 0 {
			opts.Name = args[0]
		}

		provider, err := resolveProvider()
		if err != nil {
			return err
		}
		opts.Provider = provider

		return cmd.RunEdit(opts)
	},
}

func init() {
	// Register --merge and --force on edit too (advertised in current help).
	// These use local variables since EditOptions doesn't consume them yet —
	// they exist purely for backwards compatibility with scripts that pass them.
	var editMerge, editForce bool
	editCmd.Flags().BoolVar(&editMerge, "merge", false, "Auto-merge progress on conversion conflicts")
	editCmd.Flags().BoolVar(&editForce, "force", false, "Auto-overwrite on conversion conflicts")
	rootCmd.AddCommand(editCmd)
}
```

- [ ] **Step 3: Create `status.go`**

```go
package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show progress for a PRD",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		opts := cmd.StatusOptions{}
		if len(args) > 0 {
			opts.Name = args[0]
		}
		return cmd.RunStatus(opts)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
```

- [ ] **Step 4: Create `list.go`**

```go
package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all PRDs with progress",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunList(cmd.ListOptions{})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
```

- [ ] **Step 5: Create `update.go`**

```go
package commands

import (
	"github.com/minicodemonkey/chief/internal/cmd"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Chief to the latest version",
	Args:  cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		return cmd.RunUpdate(cmd.UpdateOptions{
			Version: Version,
		})
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
```

- [ ] **Step 6: Add `printWiggum()` to `root.go`**

Move the `printWiggum()` function from `main.go` into `root.go` (the wiggum command is already registered in `root.go`'s `init()` from Task 2).

- [ ] **Step 7: Verify all files compile**

Run: `go build ./cmd/chief/commands/...`
Expected: no errors

- [ ] **Step 8: Commit**

```bash
git add cmd/chief/commands/root.go cmd/chief/commands/new.go cmd/chief/commands/edit.go cmd/chief/commands/status.go cmd/chief/commands/list.go cmd/chief/commands/update.go
git commit -m "feat: add Cobra subcommands for new, edit, status, list, update, wiggum"
```

---

### Task 4: Wire up main.go and remove old routing

**Files:**
- Modify: `cmd/chief/main.go`

- [ ] **Step 1: Replace main.go contents**

```go
package main

import "github.com/minicodemonkey/chief/cmd/chief/commands"

// Version is set at build time via ldflags
var Version = "dev"

func main() {
	commands.Execute(Version)
}
```

This removes: the switch/case routing, `parseTUIFlags()`, `parseAgentFlags()`, `runNew()`, `runEdit()`, `runStatus()`, `runList()`, `runUpdate()`, `runTUIWithOptions()`, `findAvailablePRD()`, `listAvailablePRDs()`, `resolveProvider()`, `printHelp()`, `printWiggum()`, `TUIOptions` struct — all now live in the `commands` package.

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/chief/...`
Expected: no errors

- [ ] **Step 3: Run existing tests**

Run: `go test ./...`
Expected: all tests pass

- [ ] **Step 4: Commit**

```bash
git add cmd/chief/main.go
git commit -m "refactor: replace manual CLI routing with Cobra framework

Guts main.go to a thin entry point that calls commands.Execute().
All command routing, flag parsing, and TUI logic moved to
cmd/chief/commands/ package. No behavioral changes."
```

---

### Task 5: Manual smoke tests

**Files:** None (verification only)

- [ ] **Step 1: Build binary**

Run: `go build -o chief ./cmd/chief/`

- [ ] **Step 2: Test version flag**

Run: `./chief --version`
Expected: `chief version dev`

- [ ] **Step 3: Test help**

Run: `./chief --help`
Expected: shows Cobra-generated help with all subcommands listed

- [ ] **Step 4: Test `chief help`**

Run: `./chief help`
Expected: same help output

- [ ] **Step 5: Test subcommand help**

Run: `./chief new --help`
Expected: shows help for new command with usage and args

- [ ] **Step 6: Test unknown flag**

Run: `./chief --bogus 2>&1`
Expected: error about unknown flag

- [ ] **Step 7: Test status subcommand**

Run: `./chief status`
Expected: shows status for default PRD (or appropriate error if no PRD exists)

- [ ] **Step 8: Test list subcommand**

Run: `./chief list`
Expected: lists PRDs (or empty output if none)

- [ ] **Step 9: Clean up binary**

Run: `rm ./chief`

---

### Task 6: Add command registration tests

**Files:**
- Create: `cmd/chief/commands/root_test.go`

- [ ] **Step 1: Write tests for command registration and flags**

```go
package commands

import (
	"testing"
)

func TestRootCommandHasExpectedSubcommands(t *testing.T) {
	expected := []string{"new", "edit", "status", "list", "update", "wiggum", "completion"}
	commands := rootCmd.Commands()

	found := make(map[string]bool)
	for _, cmd := range commands {
		found[cmd.Name()] = true
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestPersistentFlagsRegistered(t *testing.T) {
	flags := []string{"agent", "agent-path", "verbose"}
	for _, name := range flags {
		if rootCmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("expected persistent flag %q not found", name)
		}
	}
}

func TestLocalFlagsRegistered(t *testing.T) {
	flags := []string{"max-iterations", "no-retry", "merge", "force"}
	for _, name := range flags {
		if rootCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected local flag %q not found", name)
		}
	}
}

func TestMaxIterationsHasShorthand(t *testing.T) {
	f := rootCmd.Flags().Lookup("max-iterations")
	if f == nil {
		t.Fatal("max-iterations flag not found")
	}
	if f.Shorthand != "n" {
		t.Errorf("expected shorthand 'n', got %q", f.Shorthand)
	}
}

func TestEditCommandHasMergeAndForceFlags(t *testing.T) {
	if editCmd.Flags().Lookup("merge") == nil {
		t.Error("edit command missing --merge flag")
	}
	if editCmd.Flags().Lookup("force") == nil {
		t.Error("edit command missing --force flag")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./cmd/chief/commands/ -v`
Expected: all tests pass

- [ ] **Step 3: Commit**

```bash
git add cmd/chief/commands/root_test.go
git commit -m "test: add command registration tests for Cobra migration"
```
