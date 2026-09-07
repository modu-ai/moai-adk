# t506 — baseline measurement (axis 1: ghost `[[skills.config]]` prune verb)

Card: t506 · Class C · Tier S~M
Tree: `.claude/worktrees/t506`, branch `WT-codex-ghost-skills`, HEAD `ace1c5440` (= `origin/develop` at dispatch)
Date: 2026-09-07

## Axis split (decided at kickoff)

The card carries two independent axes. Only **axis 1** is in scope here.

| Axis | Scope | Disposition |
|---|---|---|
| 1 — ghost `[[skills.config]]` prune verb | this card | IN SCOPE |
| 2 — `codex_measured_version` 0.147.0 → 0.153.4 full re-measurement | t496 (hook axis) + t507 (plugin manifest) must land first | DEFERRED — split reported to lead, follow-up card is the lead's call |

Reason for the split: raising the manifest stamp before the two axes now being
measured elsewhere have landed would claim coverage this card does not hold —
the exact fault `progress.md:118` of SPEC-CODEX-SKILL-LOADER-001 recorded when
it declined to raise the stamp the first time.

## Claim

This machine's `~/.codex/config.toml` declares 49 `[[skills.config]]` entries;
all 49 are absolute-shaped, all 49 declare `enabled = false`, and all 49 point
at a path that does not exist.

## Evidence

Lead's cited measurement, reproduced verbatim:

```
$ grep -c '^\[\[skills.config\]\]' ~/.codex/config.toml
49
```

Shape / flag / existence split, measured directly from the file:

```
$ grep -A2 '^\[\[skills.config\]\]' ~/.codex/config.toml | grep '^path = ' \
    | sed 's/^path = "//;s/"$//' > /tmp/t506-paths.txt
paths:       49
absolute: 49
enabled=false: 49
enabled=true: 0
exists=0 missing=49
```

The 49 declared paths are exported verbatim to `observed-skill-paths.txt`
beside this file.

Independent confirmation from the product's own detector, run from a binary
built from THIS tree (`go build -o /tmp/t506-moai ./cmd/moai`, rc 0 — not the
installed `~/go/bin/moai`, which would carry an unknown lag):

```
$ /tmp/t506-moai doctor --check "Codex Wiring"
warn  Codex Wiring  codex installed, project not wired — run moai init --agent codex;
                    ~/.codex/config.toml: 49 stale skill entries
```

## Baseline-attribution

Every figure above was produced in this run, against this tree
(`ace1c5440`), reading this machine's live `~/.codex/config.toml` (629 lines;
the entry block begins at line 324). The detector figure comes from a binary
compiled from this tree in this run, so the tool-provenance coordinate is the
tree itself.

## Consequence for the design

Mapped onto the classifier the doctor already carries
(`classifyCodexSkillPath`, `internal/cli/doctor_codex.go`), this machine's
population is:

| Class | Count | Prune-eligible |
|---|---|---|
| absolute → `fs.ErrNotExist`, `enabled = false` | 49 | yes |
| absolute → exists | 0 | no |
| home-relative | 0 | yes only after expansion + `ErrNotExist` |
| relative | 0 | **never** |
| oddly-formed | 0 | **never** |
| indeterminate (stat error other than `ErrNotExist`) | 0 | **never** |
| entry declaring no `path` key | 0 | **never** |

The four never-classes are all empty on this machine, which means this
machine's population **cannot exercise the safety boundary**. The boundary is
therefore a test obligation, not something the live config will demonstrate —
see the mutant requirement in the SPEC.

## Gaps

- The `enabled = true` missing case is unobserved here (0 instances). It is in
  scope by construction (the doctor counts it) but no live instance exists.
- The doctor renders only its `summary`; the richer `detail` string it builds
  (the per-class split) is not surfaced by the current render path. Observed,
  not investigated — out of scope for this card.
- Whether Codex itself ever prunes these entries is not measured; the doctor's
  own comment asserts it does not, and that assertion is inherited, not
  re-verified here.

## Residual risk

The 49 entries on this machine are a live observation target for t504 (path
value shape) and t502 (`[[skills.config]]` producer). This card **must not
prune them** — the verb is the deliverable, its execution against this machine
is not.
