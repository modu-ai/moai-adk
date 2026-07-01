# moai agent lint — Subcommand Reference

`moai agent lint` scans every `.claude/agents/moai/*.md` file (template tree +
local tree) for 8 lint rules enforcing `agent-common-protocol.md` §User
Interaction Boundary and related [HARD] rules.

---

## Usage

```
moai agent lint [flags]

Flags:
  --path string     Path to agent directory (default: .claude/agents/moai/ and
                    internal/template/templates/.claude/agents/moai/)
  --format string   Output format: text or json (default: "text")
  --strict          Promote warnings to errors
  -h, --help        Help for lint
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No violations found |
| 1 | Violations found (errors, or warnings promoted by --strict) |
| 2 | Malformed YAML frontmatter (per-file isolated; other files continue) |
| 3 | IO error reading file |

---

## Lint Rules

| Rule | Default | --strict | Description |
|------|---------|----------|-------------|
| LR-01 | error | error | Literal `AskUserQuestion` in body text (excluding code blocks). Orchestrator agents declaring `AskUserQuestion` in `tools:` are exempt. |
| LR-02 | error | error | `Agent` token in `tools:` CSV list. Subagents cannot spawn sub-subagents. |
| LR-03 | error | error | Missing `effort:` field in frontmatter. Promoted from warning. |
| LR-04 | error | error | Dead hook entry: `matcher:` references a tool absent from `tools:` list. |
| LR-05 | error | error | Write-heavy agent missing `isolation: worktree`. Promoted from warning. |
| LR-06 | warning | error | `--deepthink flag:` boilerplate text in `description:` field (redundant activation instructions). |
| LR-07 | error | error | Duplicate Skeptical-Evaluator Mandate block (canonical copy lives in `agent-common-protocol.md` §Skeptical Evaluation Stance). |
| LR-08 | warning | warning | Skill-preload drift within same agent category (>=50% peer-omission threshold). |
| LR-09 | error | error | `isolation: worktree` on read-only agent (`permissionMode: plan`). |
| LR-10 | error | error | Static `team-*.md` agent file (v3r2 uses dynamic team generation only). |
| LR-12 | error | error | `effort:` value drifts from the canonical effort matrix. |
| LR-13 | error | error | Invalid `effort:` enum value (must be one of: low, medium, high, xhigh, max). |
| LR-14 | error | error | Fixed `budget_tokens:` value (Opus 4.7 Adaptive Thinking rejects HTTP 400). |

---

## JSON Output Schema

When `--format=json`, the output document has this shape (version "1.0" stable
through v3.0.x minor versions; breaking field changes bump to "2.0"):

```json
{
  "version": "1.0",
  "summary": {
    "total": 3,
    "errors": 2,
    "warnings": 1
  },
  "violations": [
    {
      "rule": "LR-01",
      "severity": "error",
      "file": ".claude/agents/moai/manager-strategy.md",
      "line": 59,
      "message": "Literal AskUserQuestion found in body text ..."
    }
  ]
}
```

Parse with `jq`:

```bash
# Show summary
moai agent lint --format=json | jq '.summary'

# Show only errors
moai agent lint --format=json | jq '.violations[] | select(.severity=="error")'

# Get JSON version field
moai agent lint --format=json | jq -r '.version'
```

---

## Pre-commit Hook Integration

Add to `.pre-commit-config.yaml` to run on every commit:

```yaml
repos:
  - repo: local
    hooks:
      - id: moai-agent-lint
        name: moai agent lint
        entry: moai agent lint --path .claude/agents/moai/
        language: system
        types: [markdown]
        pass_filenames: false
```

Install once per developer:

```bash
pip install pre-commit
pre-commit install
```

---

## CI Integration

The lint step runs in the **Lint** job in `.github/workflows/ci.yml`:

```yaml
- name: Run moai agent lint
  run: ./bin/moai agent lint
```

Non-zero exit blocks PR merge via branch-protection required-status checks.

---

## Orchestrator Agent Carve-out (LR-01)

Agents that declare `AskUserQuestion` in their `tools:` frontmatter field are
automatically exempt from LR-01. This covers orchestrator-class agents (e.g.,
`manager-brain`, the main MoAI session) that legitimately invoke
`AskUserQuestion`.

New orchestrator agents self-assert by adding `AskUserQuestion` to their
`tools:` list. No central allowlist maintenance required.

---

## Template-First Discipline

All agent definition files have a canonical copy under
`internal/template/templates/.claude/agents/moai/` and a local mirror at
`.claude/agents/moai/`. After editing templates, run `make build` to
recompile the binary (templates are embedded via `//go:embed all:templates`
in `embed.go`; there is no generated `embedded.go`) and synchronize local
mirrors.

`moai agent lint` scans both trees by default and emits `LINT_TREE_DRIFT`
warnings when violation counts differ between the template and local copies of
the same agent file.
