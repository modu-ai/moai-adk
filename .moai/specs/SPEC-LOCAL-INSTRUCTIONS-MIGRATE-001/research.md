# SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001 — research

> Authored at v0.2.0 by card t1259 with the Tier raise to L. Every figure here was measured **in
> the t1259 worktree on 2026-09-26**, with the command that produced it. Nothing is carried from
> the carve, and the carve's two figures that this run contradicts are recorded as contradictions
> rather than replaced silently. The parent SPEC's `research.md` §B source survey is
> cross-referenced, not duplicated.

## §A Q4 — the launcher's diagnostic surface

**Question (inherited from the parent SPEC).** Does the Codex launcher's fallback branch have a
diagnostic surface on which to emit the deprecation advisory without polluting
`developer_instructions`?

**Answer: yes — at the caller, and it is already in use for operator diagnostics.**

Producer, and its callers:

```
$ grep -rn 'codexLocalDeveloperInstructionArgs' internal/cli/*.go | grep -v _test
internal/cli/codex_launcher.go:121:// codexLocalDeveloperInstructionArgs reads both local inputs fresh on each
internal/cli/codex_launcher.go:123:func codexLocalDeveloperInstructionArgs(projectRoot string) ([]string, error) {
internal/cli/codex_launcher.go:825:	localArgs, err := codexLocalDeveloperInstructionArgs(projectRoot)
```

One definition, **one** call site. Read at `codex_launcher.go:123-143`, the producer's signature is
`(projectRoot string) ([]string, error)`: no writer, no logger, no `cobra.Command`. Its loop is

```go
for _, name := range []string{codexClaudeLocalName, codexLocalInstructionName} {
```

and everything it accumulates is `json.Marshal`-ed into `developer_instructions`. Emitting from
inside it therefore lands in the model payload — the outcome Q4 exists to avoid.

The call site at `:825` is inside `runCodexLaunch(cmd *cobra.Command, …)`, which already writes
operator diagnostics to `cmd.ErrOrStderr()` — the install hint at `:804` and the worktree error at
`:842`. That is the surface, it pre-dates this SPEC, and it is already the channel for messages
addressed to the operator rather than the model.

**Q4 is closed.** plan.md M2 no longer opens with it. Design consequence — which of the two
plumbing shapes carries the "which file did I read" signal to the caller — is `design.md` §B.

## §B `CLAUDE.local.md` size — the carve's figure does not reproduce

**Claim under test.** The carve recorded the canonical copy at 44,381 characters and derived a
reduction floor of "at least 4,382" from it.

**Measured, this tree, 2026-09-26:**

```
$ git show origin/develop:CLAUDE.local.md | wc -m
   44740
$ git show develop:CLAUDE.local.md | wc -m
   44740
```

**44,740**, and the remote-tracking and local `develop` copies agree. The carve's 44,381 is
**359 characters short** and does not reproduce.

The two readings are not in conflict about the method — both used `wc -m` against the
`develop`-committed copy, which is the copy `CLAUDE.local.md` §0.1 declares canonical. The file
moved between the measurements. It is a live maintainer document and sibling cards keep landing in
it, so it will move again before M3 runs.

**What this changes.** A reduction floor is a constant derived from a moving quantity, so it is
stale by construction. v0.1.1 had already repaired this floor once, for an off-by-one at the
`< 40000` boundary — a correct arithmetic repair to a figure that should never have been fixed,
which is the more instructive part: the repair made a drifting constant look authoritative.
`AC-IFU-007` now asserts only `after <= 39,999`, which does not drift, and requires the
before-value to be measured at the milestone.

At 44,740 the expected reduction is ~4,741 characters (~11%). That is an expectation for planning,
not a bound.

## §C The docs-site six pages — the cited path matches nothing, and one stem is ambiguous

**Claim under test.** `AC-IFU-023` cited six page stems in `docs-site/content/{ko,en,ja,zh}/`.

**Measured:**

```
$ ls docs-site/content/ko/*claude-md-guide*
zsh: no matches found
```

No page sits at the cited depth. Resolving each stem instead:

```
$ for p in claude-md-guide codex-dual-harness harness-learning memory quickstart update; do
    find docs-site/content -name "$p.md"; done
```

| Page | Path under each locale | Files |
|---|---|---|
| claude-md-guide | `advanced/claude-md-guide.md` | 4 |
| codex-dual-harness | `advanced/codex-dual-harness.md` | 4 |
| harness-learning | `advanced/harness-learning.md` | 4 |
| **memory** | **two distinct pages — see below** | **8** |
| quickstart | `getting-started/quickstart.md` | 4 |
| update | `cli-reference/update.md` | 4 |

Two findings, both material:

**(1) The glob was vacuous.** Every page is one or two directories deeper than the criterion said.
A literal reading matches the empty set, and a criterion asserting "all four locales report a
match" over an empty set is not a check.

**(2) `memory` resolves to two files per locale, and the criterion named neither.**
`cli-reference/memory.md` and `claude-code/context-memory/memory.md` both exist in ko, en, ja, zh.
Disambiguated by content:

```
$ grep -c 'CLAUDE.local.md\|CLAUDE.md' docs-site/content/ko/cli-reference/memory.md
0
$ grep -c 'CLAUDE.local.md\|CLAUDE.md' docs-site/content/ko/claude-code/context-memory/memory.md
28
```

`cli-reference/memory.md` is the `moai memory` CLI reference and says nothing about instruction
files. `claude-code/context-memory/memory.md` is the Claude Code memory-file concept page and
carries the discussion `REQ-IFU-020` is about. **The target is the latter.** `REQ-IFU-020` now
names all six by path.

The failure mode this avoids is quiet: editing `cli-reference/memory.md` would satisfy a
stem-based grep for `AGENTS.local.md` while leaving the page that actually documents the structure
untouched.

## §D Tier — the file-count axis

```
6 pages × 4 locales                                  = 24
internal/cli (launcher, contract, new verb + tests,
              update, doctor, their tests)           ≈  8
CLAUDE.local.md → AGENTS.local.md + .moai/docs/      ≈  3
                                                      ----
                                                      ≈ 35
```

The Tier L threshold is `> 15` files (`spec-workflow.md` § SPEC Complexity Tier). The carve's
provisional `M` rested on the requirement count, which was within the M ceiling and is not the
binding axis. The LOC axis agrees: `migrate_agency.go` — the precedent the verb is modelled on —
is 25,790 bytes, with a 30,580-byte test file beside it.

**Tier L.** Artifact set 5 files, plan-auditor PASS threshold 0.85.

## §E What was NOT measured

- The parent SPEC's M2 landing state. plan.md §B holds it as a dependency; whether it has landed
  is the lead's read at dispatch, not a plan-phase measurement.
- The t1175 rules-diet dependency, same reason.
- Whether `TestCodexLocalInstructions_DualFileMatrix` currently passes. `AC-IFU-029` cites it as
  existing and asserting the provenance preamble, which the carve recorded; this run confirmed
  neither, and the criterion's own `--- PASS:` assertion is what settles it at run time.
- The `.moai/docs/` split point for M3 — which sections of `CLAUDE.local.md` are procedure and
  which are rules. That is M3's first act, and doing it here would be re-deciding the document's
  content, which spec.md §D puts out of scope.

## §F Cross-references

- `SPEC-INSTRUCTION-FILES-UNIFY-001/research.md` §B — the shared source survey, attributed to base
  develop `553e224f3`. Read rather than duplicated; every figure re-measured before use.
- `.moai/reports/t1243/plan-audit-iter1.md` — the audit whose D2 arithmetic forced the carve.
- `CLAUDE.local.md` §0 — the canonical-copy discriminant §B rests on.
