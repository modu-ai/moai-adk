---
title: CG Mode (Claude + GLM)
weight: 20
draft: false
---

## What is CG mode?

CG (Claude + GLM) mode is a hybrid mode where the leader uses the **Claude
API** and the workers use the **GLM API**. It is implemented with
tmux-session-level environment variable isolation and executes the Tokenomics
split — "Claude plans deep, GLM implements cheap" — inside a single session.
For implementation-heavy work, it saves roughly 60-70% of the cost.

## Architecture

```
moai cg runs
    │
    ├── 1. Inject GLM settings into the tmux session environment
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* variables)
    │
    ├── 2. Remove GLM environment variables from settings.local.json
    │      → the leader pane uses the Claude API
    │
    ├── 3. Set teammateMode: "tmux" in settings.local.json
    │      → workers inherit GLM env vars in new panes
    │
    └── 4. Launch Claude Code (replaces the current process)
```

```
┌─────────────────────────────────────────────────────────────┐
│  Leader (current tmux pane, Claude API)                      │
│  - Workflow orchestration                                    │
│  - Handles plan, quality, and sync phases                    │
│  - No GLM env vars → uses the Claude API                     │
└──────────────────────┬──────────────────────────────────────┘
                       │ Teammate spawn (new tmux pane)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Teammates (new tmux panes, GLM API)                         │
│  - Inherit tmux session env vars → use the GLM API           │
│  - Execute implementation work in the run phase              │
│  - Communicate with the leader via SendMessage               │
└─────────────────────────────────────────────────────────────┘
```

## How to use it

### Step 1: save your GLM API key (once)

```bash
moai glm sk-your-glm-api-key
```

The key is stored safely in `~/.moai/.env.glm`.

### Step 2: check your tmux environment

If you are already inside tmux, there is no need to create a new session.

```bash
# If you are not in tmux:
tmux new -s moai
```

> **Tip**: setting tmux as your VS Code terminal default lets you skip this step entirely.

### Step 3: launch CG mode

```bash
moai cg
```

`moai cg` automatically launches Claude Code in the current pane. There is no
need to run `claude` separately.

### Step 4: run your workflow

```bash
/moai "Implement user authentication feature"
```

From here it works as usual. The orchestrator (the leader, Claude) handles
planning, quality, and sync, while implementation-heavy work is delegated to
GLM teammates in new tmux panes.

> **Note**: the old `--team` flag (the Agent Teams static-orchestration layer)
> was retired in v3.0. Forcing it falls back to sub-agent mode. CG mode's
> leader/worker separation runs on Claude Code's built-in teammate runtime
> (tmux panes), and that runtime is preserved.

## Important notes

| Item | Description |
|------|------|
| **tmux environment** | No new session needed if you are already in tmux. Setting tmux as the VS Code terminal default is convenient |
| **Auto launch** | `moai cg` auto-launches Claude Code in the current pane. No separate `claude` command needed |
| **Session end** | The session_end hook automatically cleans up the tmux session env vars → the next session uses Claude |
| **Team communication** | Leader↔worker communication via the SendMessage tool |
| **Mode switching** | When switching from `moai glm`, `moai cg` auto-resets the GLM settings — no intermediate `moai cc` needed |

## tmux environment variable injection security model {#tmux-env-security}

Since v3.0.0, when `moai cg` injects the GLM token
(`ANTHROPIC_AUTH_TOKEN`) into the tmux session environment, it uses the
**source-file channel** (`tmux source-file <tmp>`) instead of the **argv
channel** (`tmux set-environment <KEY> <VALUE>`). The token is no longer
exposed in plaintext to `ps auxe`, `/proc/<pid>/cmdline`, auditd logs, sysmon
traces, or crash dumps (CWE-214).

### Injection flow

1. Create a temp file under `~/.moai/run/` with `mkstemp` (mode `0o600` enforced)
2. Write a single `set-environment -t <session> <KEY> <VALUE>` line
3. Have tmux read the file into the environment via `tmux source-file <tmp>`
4. Unlink with `os.Remove` immediately after injection

Only the temp file path appears in argv — the token itself is never exposed.

### Non-sensitive values stay on argv

Values that are not tokens — `CLAUDE_CONFIG_DIR`, `ANTHROPIC_BASE_URL`,
`ANTHROPIC_DEFAULT_*_MODEL`, etc. — keep the existing argv path (no security
threat).

### User responsibility

The `~/.moai/.env.glm` source file must keep `0o600` permissions in your
environment. The `moai glm` command sets this automatically:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### Self-check

Verify the token is not exposed in argv while CG mode is running:

```bash
# Inside a new tmux session after running moai cg
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# Expected: 0 matches (the token is not in argv)
```

For the detailed threat model, failure behavior (the
`ErrTmuxSensitiveInjectFailed` sentinel), and additional checks, see
[Security Notes — CWE-214](/en/advanced/security-notes/#cwe-214).

## Display mode (teammateMode)

`teammateMode` is a Claude Code built-in display setting, stored in
`settings.local.json`. It is a different concept from MoAI's team-mode (the old
`--team` flag, retired in v3.0) — the teammate runtime itself is provided by
Claude Code, and `teammateMode` controls only its display style.

| Value | Description | Leader/worker separation | CG mode |
|------|------|--------------|---------|
| `in-process` | Default, inline in the same terminal | No | Not used |
| `auto` | Auto-detect environment | Not supported | Not used |
| `tmux` | tmux split screen | Session env-var isolation | {{< icon check ok >}} Used |
| `iterm2` | iTerm2 split screen | Not supported | Not used |

`moai cg` and `moai glm` set `teammateMode` to `"tmux"` in
`settings.local.json`, and `moai cc` clears it to an empty value. The
`teammateMode` setting takes precedence over the old
`CLAUDE_CODE_TEAMMATE_DISPLAY` environment variable.

> **CG mode can only separate leader/worker APIs in the `tmux` display mode.**

## Mode comparison

| Command | Leader | Workers | tmux required | Cost savings | Use case |
|--------|------|------|----------|----------|------|
| `moai cc` | Claude | Claude | No | - | Complex work, highest quality |
| `moai glm` | GLM | GLM | Recommended | ~70% | Cost optimization |
| `moai cg` | Claude | GLM | **Required** | **~60%** | Quality + cost balance |

### When should you use CG mode?

**Good fit for CG mode:**
- Implementation-heavy SPEC execution (run phase)
- Code generation
- Test writing
- Documentation generation

**Good fit for Claude-only (cc):**
- Architecture design/planning (needs Opus reasoning)
- Security reviews (needs Claude's security training)
- Complex debugging (needs advanced reasoning)

## Troubleshooting

| Problem | Cause | Fix |
|------|------|------|
| Workers use the Claude API | tmux session env vars not set | Re-run `moai cg` inside tmux |
| Claude Code does not launch after `moai cg` | Run outside tmux | `tmux new -s moai`, then re-run |
| GLM env vars remain after the session ends | session_end hook failed | Clean up manually with `moai cc` |

## Next steps

- [Model Policy](/en/multi-llm/model-policy) — per-agent model assignments
- [FAQ](/en/getting-started/faq) — execution mode FAQ
- [CLI Reference](/en/getting-started/cli) — moai cc, moai glm, moai cg details
