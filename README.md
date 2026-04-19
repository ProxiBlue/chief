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

Chief includes an optional adversarial evaluation system that validates generated code against story acceptance criteria. Standard evaluation and security evaluation are independent — use either, or both:

```bash
# Standard evaluation only (acceptance criteria)
chief --eval

# Security evaluation only (OWASP Top 10)
chief --security-eval

# Both together
chief --eval --security-eval

# With model overrides (use cheaper models for evaluators)
chief --eval --eval-model claude-haiku-4-5-20251001
chief --security-eval --security-eval-model claude-haiku-4-5-20251001
```

After each story is committed, evaluator agents score the output on a 1-10 scale, deliberate on findings, and produce a pass/fail verdict. Stories that fail are automatically retried.

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
2. If `--eval`: N evaluator agents (default 3) independently score each acceptance criterion
3. If `--security-eval`: security evaluator(s) (default 1) scan the diff for OWASP Top 10 vulnerabilities
4. All enabled evaluators run in parallel, then deliberate — challenging false positives and surfacing missed issues
5. A story passes only if all criteria from all enabled evaluators meet the threshold (default 7/10)
6. Failed stories are retried up to `maxRetries` times — the failure log shows exactly which criteria failed, their scores, and why

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

securityEvaluation:
  enabled: false        # or use --security-eval flag
  agents: 1             # number of parallel security evaluators
  passThreshold: 7      # minimum score per OWASP criterion (1-10)
  maxRetries: 3         # retry attempts on failure
  provider: ""          # LLM provider (defaults to eval provider, then main)
  model: ""             # model override for security evaluators
```

Standard evaluation and security evaluation are fully independent:

- **`--eval` only** — runs N acceptance-criteria evaluators, no security checks
- **`--security-eval` only** — runs security evaluators (OWASP Top 10), no acceptance-criteria checks
- **Both flags** — runs both in parallel, both must pass for the story to advance

The `model` option lets you use a cheaper or different model for evaluators while keeping the main generator on a more capable model. For example, generate with Opus and evaluate with Haiku to reduce costs. Each evaluation type can use its own model independently.

### Architecture

Evaluator agents are spawned as regular agent CLI subprocesses — the same CLI used for code generation (Claude Code, Codex, etc.). Each evaluator is a fresh, independent process that receives a one-shot prompt containing:

- The story's acceptance criteria (as JSON)
- The full git diff of the committed code (capped at 200KB)
- Instructions to score each criterion on a 1-10 scale and output structured JSON

Evaluators are allowed up to 3 tool calls (file reads, shell commands) to verify the implementation before scoring. When both `--eval` and `--security-eval` are enabled, all evaluators (standard + security) run in parallel, then a deliberation round shows each evaluator the others' scores and asks them to challenge false positives, agree with legitimate findings, and surface missed issues.

#### Security Evaluator

In addition to the standard acceptance-criteria evaluators, every evaluation cycle automatically includes a **security specialist evaluator**. This evaluator uses a dedicated prompt structured around the **OWASP Top 10 (2021)** framework, checking the diff for:

- **A01: Broken Access Control** — missing auth checks, privilege escalation, IDOR, CORS misconfiguration
- **A02: Cryptographic Failures** — weak encryption, hardcoded secrets, missing TLS, weak hashing
- **A03: Injection** — SQL injection, command injection, XSS, template injection, log injection
- **A04: Insecure Design** — missing rate limiting, business logic flaws, insecure defaults
- **A05: Security Misconfiguration** — verbose errors, unnecessary features, missing security headers
- **A06: Vulnerable Components** — known-vulnerable patterns, unsafe external library usage
- **A07: Auth Failures** — weak sessions, credential stuffing exposure, missing authentication
- **A08: Data Integrity Failures** — insecure deserialization, prototype pollution, unsigned updates
- **A09: Logging Failures** — sensitive data in logs, missing audit trails
- **A10: SSRF** — unvalidated URLs, missing allowlists for outbound requests

Plus cross-cutting concerns: race conditions, resource management, and file system safety.

Evaluation results reference OWASP category IDs (e.g., `A03:2021 No injection vulnerabilities`) for traceability. The security evaluator's scores are merged into the final results alongside the standard evaluators. If any security criterion scores below the pass threshold, the story fails — the same as any other criterion.

The security evaluator prompt can be customized via `.chief/prompts/security_evaluator_prompt.txt` (see [Custom Prompts](#custom-prompts)).

#### Failure Details

When an evaluation fails, the TUI log now shows exactly what went wrong:

```
❌ [STORY-1] Evaluation FAILED (attempt 1/3)
     Failed criteria:
     - No injection vulnerabilities [4/10]: api/handler.go:52 - user input passed to shell unsanitized
     - Input validation is sufficient [3/10]: auth/login.go:28 - no length limit on password field
```

This makes it easy to understand why a story was rejected without having to dig into the full evaluation result JSON.

**Important:** This is a "peer review by LLM" pattern, not a formal eval framework. Evaluators use the same model and tool access as the generator. They are effective at catching obvious acceptance criteria misses (missing features, broken mechanics, incomplete implementations) and common security anti-patterns, but they share the same limitations as the generator — they can hallucinate, miss subtle bugs, or be overly generous. For rigorous validation, complement with manual testing or a dedicated eval harness.

### Examples

See [chiefloopEVALexample](https://github.com/ProxiBlue/chiefloopEVALexample) for a side-by-side comparison of the same task (a Mage Lander game) built with and without evaluation enabled.

## Custom Prompts

Chief uses built-in prompts for agents, evaluators, and PRD generation. You can override any prompt by placing a custom file in `.chief/prompts/`:

| File | Purpose | Key Placeholders |
|------|---------|-----------------|
| `prompt.txt` | Agent instructions for story implementation | `{{PROGRESS_PATH}}`, `{{STORY_CONTEXT}}`, `{{STORY_ID}}`, `{{STORY_TITLE}}` |
| `init_prompt.txt` | PRD generator (used by `chief new`) | `{{PRD_DIR}}`, `{{CONTEXT}}` |
| `edit_prompt.txt` | PRD editor (used by `chief edit`) | `{{PRD_DIR}}` |
| `evaluator_prompt.txt` | Adversarial evaluator for acceptance criteria | `{{EVALUATOR_ID}}`, `{{STORY_CONTEXT}}`, `{{DIFF}}` |
| `security_evaluator_prompt.txt` | Security evaluator (OWASP Top 10) | `{{EVALUATOR_ID}}`, `{{STORY_CONTEXT}}`, `{{DIFF}}` |
| `deliberation_prompt.txt` | Deliberation round between evaluators | `{{EVALUATOR_ID}}`, `{{STORY_CONTEXT}}`, `{{OWN_FINDINGS}}`, `{{OTHER_FINDINGS}}` |
| `detect_setup_prompt.txt` | Project setup command detection | *(none)* |

### Getting Started

Example files (`.example`) are included in `.chief/prompts/` showing the current defaults. To override a prompt, copy the example to the active filename:

```bash
# Example: customize the security evaluator
cp .chief/prompts/security_evaluator_prompt.txt.example \
   .chief/prompts/security_evaluator_prompt.txt

# Edit your custom version
vim .chief/prompts/security_evaluator_prompt.txt
```

Override files use the same `{{PLACEHOLDER}}` syntax as the built-in prompts. Chief checks for overrides at startup — if a file exists in `.chief/prompts/`, it takes precedence over the compiled-in default. Remove the override file to revert to the built-in prompt.

### Use Cases

- **Tighten security evaluation** — Add domain-specific security rules (e.g., HIPAA, PCI-DSS) to the security evaluator prompt
- **Adjust evaluation strictness** — Modify scoring rubrics or add custom acceptance criteria categories
- **Customize agent behavior** — Change how the agent structures commits, writes progress reports, or handles quality checks
- **Project-specific PRD templates** — Tailor the PRD generator to your team's format and conventions

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
