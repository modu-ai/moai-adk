---
title: moai ast-grep / ast-edit Structural Search and Rewrite
weight: 77
draft: false
---

`moai ast-grep` scans code by syntax tree, and `moai ast-edit` rewrites the code it matches. Unlike text-based `grep`, both match on syntactic structure, so differences in whitespace, line breaks, and variable names do not throw the match off.

Both commands drive the [ast-grep](https://ast-grep.github.io/) CLI (`sg`), and what they do when `sg` is missing differs on purpose. `ast-grep` is a detector, often wired into CI as a gate, so it writes install guidance to stderr and **exits non-zero** — a scan that never ran must not be read as a clean one. `ast-edit` is a rewriter, for which "nothing to apply" is a genuine no-op, so it prints a notice and exits 0. Install `sg` from the [ast-grep quick start](https://ast-grep.github.io/guide/quick-start.html); `moai doctor` also reports whether it is present.

> **Reading and writing are separate commands.** `ast-grep` never modifies a file; `ast-edit` does. Because they are distinct commands, granting `Bash(moai ast-grep:*)` does not also grant write access.

## moai ast-grep — scan (read-only)

| Flag | Description |
|--------|------|
| `--format` | Output format: `text` (default), `json`, `sarif` |
| `--lang` | Scan only the given language (e.g. `go`, `python`, `typescript`) |
| `--severity` | Minimum severity to display (`error`, `warning`, `info`) |
| `--rules-dir` | Rules directory (default `.moai/config/astgrep-rules`) |
| `--dry` | Print only the rules that would be applied, skipping the scan |

```bash
# Scan the whole project
moai ast-grep ./

# Go only, SARIF output (for GitHub code scanning upload)
moai ast-grep --format=sarif --lang=go ./internal/

# Show error severity only
moai ast-grep --severity=error ./
```

## moai ast-edit — rewrite (modifies files)

Run without `--dry` and the command **modifies files in place**. Preview with `--dry` first.

| Flag | Description |
|--------|------|
| `--dry` | Print what would change without touching any file |
| `--pattern` | ast-grep pattern to match (used with `--rewrite`) |
| `--rewrite` | Replacement pattern (used with `--pattern`) |
| `--rule` | Apply only the rule with this ID (rule mode) |
| `--lang` | Language of the target code |
| `--rules-dir` | Rules directory (default `.moai/config/astgrep-rules`) |
| `--format` | Output format: `text` (default), `json` |

### Pattern mode

Give `--pattern` and `--rewrite` together to rewrite every match. Supplying only one of the two is rejected.

```bash
# Preview first
moai ast-edit --dry --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/

# Apply once the preview looks right
moai ast-edit --pattern 'foo($A)' --rewrite 'bar($A)' --lang go ./internal/
```

### Rule mode

Run without `--pattern` and the command reads the rules directory and applies only the rules that declare a `fix:` field. Rules without one are detection-only; they are skipped and their count is reported.

```bash
# Apply every fix: rule (preview)
moai ast-edit --dry ./internal/

# Apply a single rule
moai ast-edit --rule my-rule-id ./internal/
```

The shipped ruleset (`go/hardcoding`, `security/credentials`, `security/crypto`, `security/injection`) is **detection-only** throughout. No `fix:` was added, deliberately: an automatic rewrite there would change meaning or break compilation. Declare `fix:` in your own project rules when you want automatic rewrites.

## Rules location

Both commands read `.moai/config/astgrep-rules/` by default. `sgconfig.yml` declares which rule directories are active.

## Related

- [moai loop](/en/cli-reference/loop) — diagnostic-driven iterative fix loop
- [CLI overview](/en/getting-started/cli)
