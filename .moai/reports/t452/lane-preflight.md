# t452 lane pre-flight measurements

Tree: `.claude/worktrees/t452` on `WT-codex-skill-wiring`, base `d592b0551` (= `origin/develop` at measurement time).
Date: 2026-09-03. Tool under measurement: `codex-cli 0.152.1` (`/Users/goos/.local/bin/codex`).

## Claim

The card premise "codex has no skill axis at all" is partly falsified. The skill EXPOSURE
axis landed and is completed; the unlanded axis is the agent-TOML skill-loader emission row.
Separately, the load-root premise the completed SPEC inherited is doubtful on this codex version.

## Evidence

| # | Command | Observed output |
|---|---|---|
| M1 | `grep -c '^\[\[skills.config\]\]' ~/.codex/config.toml` | `49` |
| M2 | `ls ~/.codex/skills/` | `.system`, `hatch-pet` only — none of the 49 registered paths exist |
| M3 | `grep -rn 'skills\.config' internal \| grep -v _test` | read-only parser + doctor only; **no writer** |
| M4 | `sed -n '1,14p' .moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/spec.md` | `status: completed`, version 0.7.0 |
| M5 | `go test ./internal/template/ -run 'Mirror' -count=1` | `ok  github.com/modu-ai/moai-adk/internal/template  1.517s` |
| M6 | `sed -n '117,123p' internal/template/agentemit/agents-codex.yaml` | `class: skill-loader` / `disposition: deferred-m1` |
| M7 | `strings /Users/goos/.local/bin/codex \| grep -nE 'agents/skills'` | exactly ONE line, inside an `external-agent-migration/src/detect/mod.rs` blob |
| M8 | `grep -n 'CODEX_HOME/skills' ~/.codex/skills/.system/skill-installer/SKILL.md` | lines 48, 58 — `$CODEX_HOME/skills` named as install + annotation root |
| M9 | `CODEX_HOME=/tmp/t452-probe/home codex doctor --json` | rc=1; enumerates no skill roots; only feature flags `skill_mcp_dependency_install`, `skill_search` |

## Baseline-attribution

Every row above was run in this run, in this tree, against `codex-cli 0.152.1`. M4-M6 read
files at `d592b0551`. No figure is carried over from another tree or point in time.

## Gaps

- Whether codex-cli 0.152.1 loads a repo-local `.agents/skills/<name>/SKILL.md` was NOT
  observed. M7-M9 raise doubt; none of them settles it. That measurement is REQ-CSL-001 of
  SPEC-CODEX-SKILL-LOADER-001 and belongs to the run phase.
- Whether the agent TOML accepts a `skills` key, and in what value shape, was NOT observed.
- No deployed user project was inspected; the mirror's real-world materialization is unmeasured.

## Residual-risk

- `strings` cannot observe a runtime-composed path. `.agents` appears as its own token in the
  same blob as M7, so a runtime `join(".agents", "skills")` would be invisible to M7. A branch-A
  outcome therefore remains fully possible despite M7-M9.
- The 49 registrations in M1/M2 were read from this machine's user layer. They are one machine's
  state and do not establish what any other user's config carries.

## Fixture left on disk for the run phase

`/tmp/t452-probe/proj/` — `.claude/skills/moai-probe-x/SKILL.md` plus
`.agents/skills/moai-probe-x -> ../../.claude/skills/moai-probe-x`. Traversal through the mirror
path was confirmed by reading the file through it. `/tmp` is volatile; the run phase rebuilds
rather than assumes it.
