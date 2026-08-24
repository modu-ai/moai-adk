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
go files=2438 wouldSKIP=22 rate=0.9%
js files=81  wouldSKIP=78 rate=96.3%
py files=14  wouldSKIP=12 rate=85.7%
```

(Figures corrected in iteration 2 — see below. The iteration-1 figures were go 1.2% / py 92.9%,
inflated by an incomplete token set.)

The auditor's reading was correct that the shipped `sec-hardcoded-credential` is `kind:` +
`regex:` in all four covered languages, and that v0.1.0's §C.2 would have classified it
underivable — making A1 a no-op everywhere. §C.2 now carries a full extraction table with two
added rows: regex top-level alternation (union of per-branch mandatory prefixes, sound by the
same disjunction logic as `any:`) and `kind:` + `regex:` (derive from the regex conjunct, sound
because `kind` narrows and never widens). Under that table the rule is derivable with tokens
`{sk-, AKIA, ghp_, xox, AIza}`, and the measured skip rate is 96.3% / 85.7% for js-ts / python
against 0.9% for Go. A1 is kept because the saving is decisively non-trivial for three of the
four covered languages, and cutting it on the strength of the 0.9% Go figure would be deciding a
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

### Audit iteration 2 remediation (PASS 0.925 — two blocking findings closed)

Verdict read from `.moai/reports/t217/plan-audit-2.md`. Both findings verified independently
before editing.

**E1 — the measurement script did not implement the rule it claimed to. Confirmed and fixed.**
`sec-command-injection-shell` (python, error) is a **four**-branch `any:` — `subprocess.call`,
`subprocess.run`, `subprocess.Popen`, `os.system` — and the script listed two. §C.2's `any:` row
requires the full union, so the omission inflated the reported saving. Root cause of my own
error: the branches were read through `grep -B4 -A8 "severity: error"`, whose 8-line window cut
the tail of the block — a bounded-output read presented as a complete one.

Auditing the other two token sets while fixing it found a **second** incomplete set the audit did
not flag: **go** carried only `,` and `=` (the first rule's tokens) on the argument that they
bound the rate. The argument is sound — a union can only lower a skip rate — but it made the
figure an upper bound, not a measurement. Both are now enumerated rule-by-rule.

- **js / ts — audited, already complete.** The `any:` has exactly two branches
  (`child_process.exec`, `cp.exec`) and both were present; the credential prefixes were present.
  Rate unchanged at 96.3%.
- **python — 2 tokens added** (`subprocess.run`, `subprocess.Popen`). 92.9% → **85.7%**, matching
  the auditor's corrected figure exactly.
- **go — 7 tokens added** (`for`, `defer`, `const`, `SignedString`, `exec.Command`,
  `template.HTML`, `md5.New`, plus the credential prefixes). 1.2% → **0.9%**.

Both corrections move in the deflating direction. The conclusion is unchanged: 0.9% for Go
against 85.7-96.3% elsewhere still carries the exit-(a) decision.

**E2 — soundness claim made honest.** §C.2's underivable row now names two further regex forms:
an inline flag (`(?i)`, `(?s)`, `(?m)`) and an optional or quantified leading literal. Both are
absent from the shipped ruleset, so the row is a soundness closure rather than machinery — the
row states that explicitly so no one builds an extractor for a case that does not exist.

**Optional items, all three taken.** AC-SSS-016's pair check now runs
`driftcheck.sh pairaxis` (the deployed ↔ template axis REQ-SSS-015 actually speaks to) as the
primary check, with the templates-internal `guarded` mode retained as a secondary signal;
AC-SSS-015 cites `~/go/bin/moai` instead of the untracked `./bin/moai` (verified: `exit=0`); and
§C.3's saving claim is softened to describe this sample rather than project a rate.

**Open gap left open, marked.** The `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/`
row in AC-SSS-016 is now labelled **UNVERIFIED** rather than "pass" — the run exceeded 120 s with
empty output, so its value is a Gap, not a Claim (`verification-claim-integrity.md` §2). M4 must
measure it with an explicit longer timeout before citing it.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
