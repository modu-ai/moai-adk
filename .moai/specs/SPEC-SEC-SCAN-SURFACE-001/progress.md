# Progress — SPEC-SEC-SCAN-SURFACE-001

Card: t217 · Branch: `security-scan-surface` · Worktree: `.claude/worktrees/t217`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (spec.md + plan.md + acceptance.md + progress.md). 16 REQ / 16 AC — at the Tier M
  ceiling, verified by count.
- Class: C — a design change spanning `internal/hook`, `internal/cli`, and the distributed
  template; carries a decision (`spec.md` §C) that a reviewer may disagree with.
- Evidence base: `.moai/reports/t217/investigation.md`, measured on tree `c4e90cd58`.
  Authoring-time measurements are tabulated in `spec.md` §A.1.
- Two card premises were found inaccurate and corrected in `spec.md` §A.2 rather than inherited.
- Status: `draft`. Awaiting re-audit (iteration 2).

### Audit iteration 1 remediation (FAIL 0.65 → resubmitted)

Verdict read from `.moai/reports/t217/plan-audit.md`. All six blocking findings were
independently re-verified before editing; none was auditor error.

| Finding | Disposition |
|---|---|
| D1 `tags` sequence ⇒ frontmatter ParseFailure | Fixed — `tags:` is now a quoted comma-separated string per the schema SSOT (`spec-frontmatter-schema.md` §Field Reference, `tags` = string). Lint clean, output below |
| D2 pre-filter may be a no-op | Resolved by measurement; exit **(a)** taken — see below |
| D3 AC-SSS-004 vacuous | Fixed — the criterion now counts scanner-side (PASS 0, today 1) *and* caller-side (PASS 1, today 0) resolutions; the two invert, so it cannot pass on the untouched tree |
| D4 drift command fails today | Fixed — `[ -f "$b" ]` guard restored (measured: guarded 0 lines, unguarded 31), and the pair check re-scoped to files this SPEC's diff changes. Pre-existing `handle-pre-tool.sh` comment drift recorded as out of scope |
| D5 measurement seam does not exist | Fixed — M0 step 1 now owns creating it. Also corrects the auditor's proposed remedy: a handler-level stub is *not* possible today either, because `preToolHandler.scanner` is a concrete `*security.SecurityScanner` (`pre_tool.go:325`) |
| D6 20 REQ > Tier M ceiling 16 | Fixed by consolidation to exactly 16, no content dropped (verified: `grep -c "^- \*\*REQ-SSS-"` → 16) |
| D7 short-circuit order dependency | Adopted — reachability clause added to REQ-SSS-012, evidence recorded in `spec.md` §A.3, asserted by AC-SSS-012 |
| D8 empty derived language set | Adopted — folded into REQ-SSS-006, asserted by AC-SSS-007 |

**D2 exit taken: (a) — specify the extractor, keep A1.**

Measurement (`python3 .moai/reports/t217/skiprate.py .`, worktree `c4e90cd58`):

```
go files=2438 wouldSKIP=30 rate=1.2%
js files=81  wouldSKIP=78 rate=96.3%
py files=14  wouldSKIP=13 rate=92.9%
```

The auditor's reading was correct that the shipped `sec-hardcoded-credential` is `kind:` +
`regex:` in all four covered languages, and that v0.1.0's §C.2 would have classified it
underivable — making A1 a no-op everywhere. §C.2 now carries a full extraction table with two
added rows: regex top-level alternation (union of per-branch mandatory prefixes, sound by the
same disjunction logic as `any:`) and `kind:` + `regex:` (derive from the regex conjunct, sound
because `kind` narrows and never widens). Under that table the rule is derivable with tokens
`{sk-, AKIA, ghp_, xox, AIza}`, and the measured skip rate is 96.3% / 92.9% for js-ts / python
against 1.2% for Go. A1 is kept because the saving is decisively non-trivial for three of the
four covered languages, and cutting it on the strength of the 1.2% Go figure would be deciding a
16-language tool's behaviour from the fact that this repository happens to be written in Go
(`CLAUDE.local.md` §15). The Go degeneracy and the measurement's gaps are now stated in §C.3
rather than asserted away.

**D1 closure evidence** — `~/go/bin/moai spec lint .moai/specs/SPEC-SEC-SCAN-SURFACE-001/spec.md`:

```
✓ No findings — all SPEC documents are valid
rc=0
```

(Before the fix, the same command returned
`ERROR ParseFailure ... line 13: cannot unmarshal !!seq into string`, `rc=1`.)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
