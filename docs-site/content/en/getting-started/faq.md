---
title: Frequently Asked Questions
weight: 100
draft: false
---

Frequently asked questions and answers about using MoAI-ADK.


---

## Q: What is the difference between `moai` and `/moai`?

They are two completely different things. This is the most common confusion, so let's clear it up first.

| | `moai` (terminal CLI) | `/moai` (slash subcommand) |
|---|---|---|
| **Where it runs** | Terminal shell | Claude Code chat input |
| **What it is** | Go binary | Claude Code skill invocation |
| **Purpose** | Project setup, template deployment | AI agent development workflows |
| **Example** | `moai init my-project` | `/moai plan "auth feature"` |

- Running `moai plan` in the terminal does nothing — `/moai plan` is only valid inside Claude Code.
- Typing `/moai init` in Claude Code does nothing — `moai init` is a terminal command.

---

## Q: What does the version display in the statusline mean?

The MoAI statusline shows version information together with an update notification:

```text
🗿 v3.1.2 -> 🗿 v3.1.3
```

- **`🗿 v3.1.2`**: The currently installed version
- **`-> 🗿 v3.1.3`**: A newer version available for update, joined by the ASCII arrow `->`

When you are on the latest version, only the version number is shown:

```text
🗿 v3.1.3
```

**How to update**: Run `moai update` and the update notification disappears.

{{< callout type="info" >}}
**Note**: This is different from Claude Code's built-in version display (`🔅 v2.1.172`). The MoAI display tracks the MoAI-ADK version, while Claude Code displays its own version separately.
{{< /callout >}}

---

## Q: How do I customize the segments shown in the statusline?

The statusline is toggled one segment at a time. Turn each segment on or off to keep only the information you want. There are no display presets — the configuration is just a theme and a set of segments.

Configure it in the `moai init` or `moai update -c` wizard, or edit `.moai/config/sections/statusline.yaml` directly:

```yaml
statusline:
  segments:
    model: true
    context: true
    output_style: false
    directory: false
    git_status: true
    claude_version: false
    moai_version: false
    git_branch: true
```

With no `segments:` block, every segment is enabled by default.

{{< callout type="info" >}}
For details, see [SPEC-STATUSLINE-001](https://github.com/modu-ai/moai-adk/blob/main/.moai/specs/SPEC-STATUSLINE-001/spec.md).
{{< /callout >}}

---

## Q: How do I choose a model policy?

Since v3.2 there is no per-agent model policy to choose. **Subagents inherit the main session's model and effort** — pass neither `model` nor `effort` when spawning a subagent, and MoAI agent definitions declare neither. What remains is one session-level choice: the **Session model policy** in `moai profile setup`, which sets the default reasoning effort of the Claude session launched with the profile, applied when no effort level is chosen.

### Session Model Policy Comparison

| Value | Meaning |
|------|------|
| **high** | Session effort fallback `high` |
| **medium** (default) | Session effort fallback `medium` — the knee of the cost/score curve |
| **low** | Session effort fallback `low` — economical within the same model |

{{< callout type="warning" >}}
**Why does the effort choice matter?** Lowering the effort mostly lowers *reasoning depth*, not model class. On a long-horizon agentic task, Opus at `low` effort scores higher and costs less per task than Sonnet at any effort — the bill is set by how many steps a model spends finishing, not by the per-token rate. The session-level effort is where that economy lives now; the per-agent assignment tables of earlier versions are retired.
{{< /callout >}}

### What Changed from the Per-Agent Era

Through v3.1, MoAI-ADK assigned `{model, effort}` to each of the 13 catalog agents through a profile matrix whose column the tier selected. That apparatus was retired in SPEC-AGENT-MODEL-INHERIT-001 — measurement showed fewer than 1% of spawns ever carried a model argument, so the assignment moved to the session itself. The old `--model-policy`, `--profile`, `--high`, `--medium-alias`, and `--low` flags survive as deprecated stubs that print a warning and do nothing.

### How to Configure

```bash
# Configure the session model policy (Session model policy question)
moai profile setup

# Reconfigure an existing project
moai update -c                # Re-run the setup wizard
```

{{< callout type="info" >}}
The default effort fallback is `medium`. Change it in `moai profile setup`, or adjust the session effort as you go with `/effort` or `ultrathink` — every subagent spawned afterwards inherits it.
{{< /callout >}}

---

## Q: I see an "Allow external CLAUDE.md file imports?" warning

When opening a project, Claude Code may show a security prompt about external file imports:

```
External imports:
  /Users/<user>/.moai/config/sections/quality.yaml
  /Users/<user>/.moai/config/sections/user.yaml
  /Users/<user>/.moai/config/sections/language.yaml
```

{{< callout type="info" >}}
**Recommended action:** Choose **"No, disable external imports"**.
{{< /callout >}}

**Why:**
- These files already exist in your project's `.moai/config/sections/`
- Project-level settings take precedence over global settings
- The essential settings are already included in the CLAUDE.md text
- Disabling external imports is safer and does not affect functionality

**What the files are:**
- `quality.yaml`: TRUST 5 framework and development methodology settings
- `language.yaml`: Language settings (conversation, comments, commits)
- `user.yaml`: User name (optional, used for Co-Authored-By)

---

## Q: What is the difference between the TDD and DDD methodologies?

MoAI-ADK v2.5.0+ lets you choose between two methodologies (TDD or DDD only). The hybrid mode was removed for clarity and consistency.

TDD writes the test first and then makes it pass, which suits new development; DDD pins existing behavior down with characterization tests and then works on it in small steps, which suits code that has almost no tests. The step-by-step procedure for each cycle is covered in [SPEC-Based Development](/en/core-concepts/spec-based-dev) and [DDD](/en/core-concepts/ddd).

### Methodology Selection Table

| Project State | Test Coverage | Recommended Methodology | Reason |
|--------------|---------------|-------------|------|
| New project | N/A | TDD | Test-first development |
| Existing project | 50%+ | TDD | A test base exists |
| Existing project | 10-49% | TDD | Tests can be extended |
| Existing project | < 10% | DDD | Incremental characterization tests needed |

### How to Configure

```bash
# Auto-detected during project initialization
moai init my-project          # Can be specified with the --mode <ddd|tdd> flag

# Manual configuration
# Edit .moai/config/sections/quality.yaml
development_mode: tdd         # or ddd
```


---

## Q: Why does my code have no @MX tags?

This is **completely normal**. The @MX tag system is designed to mark only the most dangerous and important code the AI should look at first.

| Question | Answer |
|------|------|
| Is it a problem if there are no tags? | **No.** Most code does not need tags. |
| When are tags added? | Only for **high fan_in** (callers >= 3), **complex logic** (complexity >= 15), and **risky patterns** (goroutines without context). |
| Is it similar across projects? | **Yes.** In every project, most code carries no tags. |

### Tag Priorities

| Priority | Condition | Tag Type |
|---------|------|----------|
| **P1 (critical)** | fan_in >= 3 | `@MX:ANCHOR` |
| **P2 (risky)** | goroutines, complexity >= 15 | `@MX:WARN` |
| **P3 (context)** | magic constants, missing godoc | `@MX:NOTE` |
| **P4 (missing)** | no test file | `@MX:TODO` |

To scan your codebase for @MX tags:

```bash
/moai mx --all        # Full scan
/moai mx --dry        # Preview
/moai mx --priority P1  # Critical items only
```

---

## More Questions?

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) — Questions, ideas, feedback
- [Issues](https://github.com/modu-ai/moai-adk/issues) — Bug reports, feature requests
- [Discord Community](https://discord.gg/Z7E7Mdc5aN) — Real-time chat, tips
