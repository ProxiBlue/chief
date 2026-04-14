# Chief

<p align="center">
  <img src="assets/hero.png" alt="Chief" width="500">
</p>

Build big projects with Claude. Chief breaks your work into tasks and runs Claude Code in a loop until they're done.

**[Documentation](https://minicodemonkey.github.io/chief/)** · **[Quick Start](https://minicodemonkey.github.io/chief/guide/quick-start)**

![Chief TUI](https://minicodemonkey.github.io/chief/images/tui-screenshot.png)

## Install

```bash
brew install minicodemonkey/chief/chief
```

Or via install script:

```bash
curl -fsSL https://raw.githubusercontent.com/MiniCodeMonkey/chief/refs/heads/main/install.sh | sh
```

### Install this fork (with adversarial evaluation)

```bash
curl -fsSL -o /usr/local/bin/chief https://raw.githubusercontent.com/ProxiBlue/chief/main/bin/chief && chmod +x /usr/local/bin/chief
```

## Usage

```bash
# Create a new project
chief new

# Launch the TUI and press 's' to start
chief
```

Chief runs Claude in a [Ralph Wiggum loop](https://ghuntley.com/ralph/): each iteration starts with a fresh context window, but progress is persisted between runs. This lets Claude work through large projects without hitting context limits.

## How It Works

1. **Describe your project** as a series of tasks
2. **Chief runs Claude** in a loop, one task at a time
3. **One commit per task** — clean git history, easy to review

See the [documentation](https://minicodemonkey.github.io/chief/concepts/how-it-works) for details.

## Adversarial Evaluation

Chief includes an optional adversarial evaluation system that validates generated code against story acceptance criteria. Enable it with the `--eval` flag:

```bash
chief --eval

# Use a specific (cheaper) model for evaluators
chief --eval --eval-model claude-haiku-4-5-20251001
```

After each story is committed, multiple independent evaluator agents score the output on a 1-10 scale, deliberate on findings, and produce a pass/fail verdict. Stories that fail are automatically retried.

### Why Use Eval Over the Standard Loop?

In the standard loop, the generator agent is the only judge of its own work. It commits code and moves to the next story. If it *thinks* it implemented a feature but didn't (hallucination), or misunderstood an acceptance criterion, nothing catches it. The loop trusts the agent completely.

The eval loop adds independent review after every commit. The concrete benefits:

- **Catches missing features** — The generator often claims "done" when it only partially implemented something. Evaluators read the diff and flag what's actually missing.
- **Catches misunderstood requirements** — The generator may interpret a criterion differently than intended. Multiple evaluators comparing notes in deliberation reduces this.
- **Forces completion over speed** — Without eval, the loop optimises for moving through stories fast. With eval, it can't advance until criteria are actually met.
- **Creates an audit trail** — Every evaluation is persisted to `.evaluation/` with full scores, reasoning, and deliberation transcripts. You can review exactly why a story passed or failed.

**The tradeoff:** More LLM calls (3 evaluators + deliberation per story, plus retries), so it's slower and more expensive. But the output quality is measurably higher — in [side-by-side testing](https://github.com/ProxiBlue/chiefloopEVALexample), the eval version implemented all core game mechanics (terrain, landing pads, scoring, levels), while the standard version was missing core features entirely.

### How It Works

1. The generator agent commits code for a user story
2. N evaluator agents (default 3) independently score each acceptance criterion
3. Evaluators deliberate — challenging false positives and surfacing missed issues
4. A story passes only if all criteria meet the threshold (default 7/10)
5. Failed stories are retried up to `maxRetries` times

### Configuration

In `.chief/config.yaml`:

```yaml
evaluation:
  enabled: false        # or use --eval flag
  agents: 3             # number of parallel evaluators
  passThreshold: 7      # minimum score per criterion (1-10)
  maxRetries: 3         # retry attempts on failure
  provider: ""          # LLM provider (defaults to main provider)
  model: ""             # model override for evaluators (e.g. "claude-sonnet-4-5-20250514")
```

The `model` option lets you use a cheaper or different model for evaluators while keeping the main generator on a more capable model. For example, generate with Opus and evaluate with Sonnet to reduce costs. When left blank, evaluators use the provider's default model.

### Architecture

Evaluator agents are spawned as regular agent CLI subprocesses — the same CLI used for code generation (Claude Code, Codex, etc.). Each evaluator is a fresh, independent process that receives a one-shot prompt containing:

- The story's acceptance criteria (as JSON)
- The full git diff of the committed code (capped at 200KB)
- Instructions to score each criterion on a 1-10 scale and output structured JSON

Evaluators are allowed up to 3 tool calls (file reads, shell commands) to verify the implementation before scoring. N evaluators (default 3) run in parallel, then a deliberation round shows each evaluator the others' scores and asks them to challenge false positives, agree with legitimate findings, and surface missed issues.

**Important:** This is a "peer review by LLM" pattern, not a formal eval framework. Evaluators use the same model and tool access as the generator. They are effective at catching obvious acceptance criteria misses (missing features, broken mechanics, incomplete implementations), but they share the same limitations as the generator — they can hallucinate, miss subtle bugs, or be overly generous. For rigorous validation, complement with manual testing or a dedicated eval harness.

### Examples

See [chiefloopEVALexample](https://github.com/ProxiBlue/chiefloopEVALexample) for a side-by-side comparison of the same task (a Mage Lander game) built with and without evaluation enabled.

## Requirements

- **[Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code)**, **[Codex CLI](https://developers.openai.com/codex/cli/reference)**, or **[OpenCode CLI](https://opencode.ai)** installed and authenticated

Use Claude by default, or configure Codex or OpenCode in `.chief/config.yaml`:

```yaml
agent:
  provider: opencode
  cliPath: /usr/local/bin/opencode   # optional
```

Or run with `chief --agent opencode` or set `CHIEF_AGENT=opencode`.

## License

MIT

## Acknowledgments

- [@Simon-BEE](https://github.com/Simon-BEE) — Multi-agent architecture and Codex CLI integration
- [@tpaulshippy](https://github.com/tpaulshippy) — OpenCode CLI support and NDJSON parser
- [snarktank/ralph](https://github.com/snarktank/ralph) — The original Ralph implementation that inspired this project
- [Geoffrey Huntley](https://ghuntley.com/ralph/) — For coining the "Ralph Wiggum loop" pattern
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
