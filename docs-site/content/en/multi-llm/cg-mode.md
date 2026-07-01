---
title: CG Mode (Claude + GLM)
weight: 20
draft: false
---

## What is CG Mode?

CG (Claude + GLM) mode is a hybrid mode where the **leader uses Claude API** and **workers use GLM API**. It's implemented through tmux session-level environment variable isolation.

## Architecture

```
moai cg execution
    │
    ├── 1. Inject GLM settings into tmux session environment variables
    │      (ANTHROPIC_AUTH_TOKEN, BASE_URL, MODEL_* variables)
    │
    ├── 2. Remove GLM environment variables from settings.local.json
    │      → Leader pane uses Claude API
    │
    ├── 3. Set CLAUDE_CODE_TEAMMATE_DISPLAY=tmux
    │      → Workers inherit GLM environment variables in new panes
    │
    └── 4. Run Claude Code (replace current process)
```

```
┌─────────────────────────────────────────────────────────────┐
│  Leader (current tmux pane, Claude API)                     │
│  - Run /moai --team to orchestrate workflows                │
│  - Handle plan, quality, sync phases                        │
│  - No GLM environment variables → Use Claude API            │
└──────────────────────┬──────────────────────────────────────┘
                       │ Agent Teams (new tmux pane)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  Team Members (new tmux pane, GLM API)                      │
│  - Inherit tmux session environment variables → Use GLM API │
│  - Execute implementation work in run phase                 │
│  - Communicate with leader via SendMessage                  │
└─────────────────────────────────────────────────────────────┘
```

## How to Use

### Step 1: Save GLM API Key (First Time Only)

```bash
moai glm sk-your-glm-api-key
```

The key is securely stored in `~/.moai/.env.glm`.

### Step 2: Check tmux Environment

If you're already using tmux, no need to create a new session.

```bash
# If not using tmux:
tmux new -s moai
```

> **Tip**: Set VS Code terminal default to tmux to skip this step entirely.

### Step 3: Run CG Mode

```bash
moai cg
```

`moai cg` automatically runs Claude Code in the current pane. No need to run `claude` separately.

### Step 4: Execute Team Workflow

```bash
/moai --team "implement user authentication"
```

## Important Notes

| Item | Description |
|------|-------------|
| **tmux environment** | No new session needed if already using tmux. Set VS Code terminal default to tmux for convenience |
| **Auto-execution** | `moai cg` auto-runs Claude Code in current pane. No separate `claude` command needed |
| **Session end** | session_end hook automatically cleans up tmux session environment variables → Next session uses Claude |
| **Team communication** | SendMessage tool for leader↔worker communication |
| **Mode switching** | `moai cg` auto-initializes GLM settings when switching from `moai glm` — no intermediate `moai cc` needed |

## tmux Environment Variable Injection Security Model {#tmux-env-security}

Starting with v2.20.0-rc1, when `moai cg` injects GLM token (`ANTHROPIC_AUTH_TOKEN`) into tmux session environment variables, it uses **source-file channel** (`tmux source-file <tmp>`) instead of **argv channel** (`tmux set-environment <KEY> <VALUE>`). The token is no longer exposed in plaintext to `ps auxe`, `/proc/<pid>/cmdline`, auditd logs, sysmon tracing, or crash dumps (CWE-214).

### Injection Flow

1. Create a temporary file under `~/.moai/run/` with `mkstemp` (mode `0o600` enforced)
2. Write a single line: `set-environment -t <session> <KEY> <VALUE>`
3. `moai cg` reads the file via `tmux source-file <tmp>` to inject into environment
4. Delete the temporary file with `os.Remove` immediately after injection

Only the temporary file path is exposed in argv; the token itself is not.

### Non-sensitive Values Keep argv Channel

Non-sensitive values like `CLAUDE_CONFIG_DIR`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_*_MODEL` use the existing argv path (no security risk).

### User Responsibility

The `~/.moai/.env.glm` source file must maintain `0o600` permissions in your environment. The `moai glm` command sets this automatically:

```bash
stat -c '%a' ~/.moai/.env.glm    # Linux: 600
stat -f '%A' ~/.moai/.env.glm    # macOS: 600
```

### Self-Audit

During CG mode execution, verify the token is not exposed in argv:

```bash
# After running moai cg in new tmux session
ps auxe | grep -i 'tmux set-environment.*ANTHROPIC_AUTH_TOKEN'
# Expected: 0 matches (token not in argv)
```

For detailed threat model, failure behavior (`ErrTmuxSensitiveInjectFailed` sentinel), and additional audit procedures, see [Security Notes — CWE-214](/en/advanced/security-notes/#cwe-214).

## Display Modes

Agent Teams supports two display modes:

| Mode | Description | Communication | Leader/Worker Separation |
|------|-------------|----------------|--------------------------|
| `in-process` | Default, all terminals | ✅ SendMessage | ❌ Same environment |
| `tmux` | Split screen display | ✅ SendMessage | ✅ Session environment isolation |

> **CG mode can only separate leader/worker APIs in `tmux` display mode.**

## Mode Comparison

| Command | Leader | Worker | tmux Required | Cost Savings | Use Case |
|---------|--------|--------|---------------|--------------|----------|
| `moai cc` | Claude | Claude | No | - | Complex tasks, highest quality |
| `moai glm` | GLM | GLM | Recommended | ~70% | Cost optimization |
| `moai cg` | Claude | GLM | **Required** | **~60%** | Quality + cost balance |

### When Should I Use CG Mode?

**Good for CG Mode:**
- Implementation-heavy SPEC execution (run phase)
- Code generation work
- Test writing
- Documentation generation

**Better with Claude Only (cc):**
- Architecture design/planning (Opus reasoning needed)
- Security review (Claude's security training needed)
- Complex debugging (advanced reasoning needed)

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Workers using Claude API | tmux session environment variables not set | Re-run `moai cg` inside tmux |
| `moai cg` doesn't run Claude Code | Executed outside tmux | Run `tmux new -s moai` then retry |
| GLM environment variables persist after session end | session_end hook failed | Manually cleanup with `moai cc` |

## Next Steps

- [Model Policy](/en/multi-llm/model-policy) — Agent-specific model assignment
- [Dual Execution Mode](/en/getting-started/faq) — Sub-Agent vs Agent Teams
- [CLI Reference](/en/getting-started/cli) — Detailed `moai cc`, `moai glm`, `moai cg`
