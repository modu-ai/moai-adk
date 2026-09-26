# plan.md — SPEC-SPEC-LINT-VACUOUS-ASSERT-001

Card t1269 · Tier M · base develop `e464fd5d0` · worktree `.claude/worktrees/t1269`, branch
`WT-vacuous-lint-rule` · run-phase evidence target: **`.moai/reports/t1269/verdict.md`**.

## §A Context

t1243 removed its prose enumeration commands under scope reduction and named this card as the
forward-looking owner of the pattern judgment (decision record
`.moai/reports/t1243/decision-scope-reduction.md`). This SPEC puts that judgment in a Go lint rule
that reuses the existing `Rule` interface, the registration slice, and the existing SPEC Lint CI
step. No Epic grouping.

## §B Design Decisions (ordered by likelihood of change — review these first)

### §B.1 DD-3 — severity and rollout (highest change-likelihood)

Plan-time **estimate** (not a measurement of the rule — the rule does not exist yet). Method: a
throwaway Python approximation of the §B.1 definitions over `spec.md`/`plan.md`/`acceptance.md`
of every SPEC directory in this worktree at `e464fd5d0`, split by frontmatter status (terminal =
completed/superseded/archived/rejected). The script lived in the session scratchpad and is not
persisted; run-phase re-measures with the real rule. Verbatim output:

```
run: terminal=2544 rows/420 dirs  active=403 rows/65 dirs
pass: terminal=357 rows/57 dirs  active=54 rows/9 dirs
```

The "active" figures are an upper bound on the gated population, because grandfather-era
demotion (which also marks warnings advisory) was not modelled.

Options weighed against the lead's constraint that a finding must turn the existing CI step red:

| Option | Effect on landing | Effect afterwards | Verdict |
|---|---|---|---|
| error | red on landing for every active finding; errors are never absorbable by the baseline | — | rejected |
| advisory | green on landing | the baseline gate never counts advisory, so no document can ever turn CI red — REQ-VTA-013 unsatisfiable | rejected |
| warning + audited rebaseline (REQ-SLGS-008) | record roughly the estimated active count (up to ~457) | every active SPEC that later reaches `completed` lowers the count while the recorded ceiling stays put (the file is never rewritten implicitly), so each closure hands out permanent headroom for new violations | rejected |
| **warning, non-advisory only for SPECs created on/after a cutoff** | gated population at landing is 0 by construction, because the cutoff is chosen strictly later than every existing `created` | recorded count stays absent (= 0); any new violation in a new SPEC is an increase → red | **chosen** |

**Cutoff selection rule (plan-audit iter-1 D3).** An earlier draft set the cutoff to the
registering commit's date. The audit measured that at 2026-09-26 two draft SPECs owned by other
cards (created that day, non-terminal, quoting the defect inside blockquotes as documentation)
would become gated and turn the landing commit red — and the only fix would have been editing
another card's SPEC. The rule is now: **cutoff = the day after the newest `created` value over
every `.moai/specs/*/spec.md` in the registering tree.** Plan-time measurement in this worktree
(`e464fd5d0`), with the throwaway approximator's date column:

```
$ grep -h -m1 "^created:" .moai/specs/*/spec.md | sed 's/created: *//; s/"//g' | sort | uniq -c | tail -4
   5 2026-09-23
   4 2026-09-24
   2 2026-09-25
  14 2026-09-26
$ grep -h -m1 -E "^created:" .moai/specs/*/spec.md | sed 's/created: *//; s/"//g' | awk '$0>="2026-09-27"' | wc -l
       0
```

So a cutoff of 2026-09-27 gates zero existing SPECs. The same measurement found 947 `spec.md`
files of which 946 carry a well-formed `created`; the one exception
(`SPEC-V3R5-INIT-WIZARD-EXPANSION-001`, legacy `created_at:` alias, status `implemented`,
grandfathered era) is why a missing `created` maps to advisory (REQ-VTA-009). M4 re-measures at
the registering tree; if newer SPECs have landed by then, the cutoff moves with them.

Consequences accepted:

- Legacy active SPECs keep their findings as advisory (visible, not gated).
- A backdated `created:` would evade the gate — visible in review; residual risk.
- **Pre-cutoff SPECs that are still active can gain new vacuous criteria after landing and are
  never gated** (plan-audit iter-1 D9) — e.g. an in-flight SPEC whose run phase or follow-up card
  edits its acceptance criteria. Those findings appear as advisory only.
- SPECs created later on the cutoff-minus-one day (after the measurement) are not gated.

### §B.2 DD-1 — detection axes

- (a) run-pattern anchoring, per §B.1 "Anchored". Covers the first-branch-unanchored case the
  repaired prose detector missed (t1243 iter-3 probe table).
- (b) outcome-assertion delimiter, per §B.1 "Delimited". Judged on every line, not only
  test-invocation lines, because criteria often phrase it as prose ("the output contains …").
- (c) out of scope — spec.md §E.

### §B.3 DD-2 — where the rule looks

Explicit artifact allowlist rather than `MovingRefUnpinnedRule`'s every-`.md` scan: `spec-compact.md`
duplicates `spec.md` (double counting), `progress.md` records commands that already ran and is
owned by other agents, `research.md`/`design.md` describe existing state. Lines are judged raw —
no markdown parsing — so fenced code, inline code, tables and blockquotes are all seen (deficit 4
of the decision record is resolved because the detector no longer lives in the document).

### §B.4 DD-4 — suppression

`lint.skip: [VacuousTestAssertion]` in frontmatter is the only escape. A SPEC that must quote a
defective example (for instance an audit of this class) uses it or moves the example into a Go
test fixture. Measured cost (plan-audit iter-1 D3): 2 of the 4 non-terminal SPECs created on
2026-09-26 quote the defect as documentation, and for such a SPEC whole-code `lint.skip` also
silences its real criteria. Accepted as a trade because the cutoff keeps every existing SPEC
advisory; if new SPECs show the need repeatedly, a reasoned per-line marker becomes a follow-up
card rather than a change here.

## §C Pre-flight (run-phase entry)

1. `git rev-parse --show-toplevel` → the t1269 worktree; `git branch --show-current` →
   `WT-vacuous-lint-rule`.
2. Re-read `internal/spec/lint.go` registration slice and `applyEraDemotion`; line numbers drift.
3. Confirm `normalizeEra` accepts `V3R6` for the fixture frontmatter `era:` override (read at plan
   time in `internal/spec/era.go`).

## §D Constraints

- No new gate, workflow, job, step, or dependency (REQ-VTA-012). `go.mod`/`go.sum` untouched.
- Every `-run` pattern written in tests, fixtures meant to be conformant, or any SPEC artifact is
  anchored at both ends; this SPEC must lint clean under its own rule (AC-VTA-011).
- Verification is scoped to `./internal/spec/...` locally; the full suite is CI's.

## §E Self-Verification (run phase fills §E.2 of progress.md)

Every AC result goes to `.moai/reports/t1269/verdict.md` as command + verbatim output + tree SHA.

## §F Milestones (priority-ordered; highest change-likelihood first)

- **M1 — Detection semantics (Priority High).** Write the two-arm unit tests first (RED):
  run-pattern axis, outcome-assertion axis, markdown contexts incl. nested blockquotes, D15
  regression inputs, shell-expansion skip. Then implement `lint_vacuous_assertion.go` to GREEN.
- **M2 — Gating (Priority High).** Cutoff constant + pinning test; advisory/non-advisory split;
  `lint.skip` suppression; confirm era/terminal demotion still applies.
- **M3 — End-to-end CI path (Priority High).** Commit fixture trees
  `internal/spec/testdata/vacuous_assert_e2e/{red,green}/`, each a minimal project root holding
  one SPEC `SPEC-FIXTURE-VTA-001` (testdata-only id; never under the repository's
  `.moai/specs/`) with `era: V3R6`, `status: draft`, `created` on or after the cutoff, and
  `.moai/spec-lint-baseline.json` without a `VacuousTestAssertion` entry (plan-time check:
  `git check-ignore` exits 1 on both planned fixture paths — not ignored). The two trees differ
  by exactly one line, blockquoted in `red`. Run `go run ../../../../../cmd/moai spec lint
  --baseline .moai/spec-lint-baseline.json` from each fixture root (AC-VTA-009) — the CI argument
  vector verbatim, the relative package path being the only difference.
- **M4 — Corpus landing measurement (Priority Medium).** Measure the newest `created` over the
  registering tree, set the cutoff to the following day, register the rule, then run the
  real-corpus baseline gate, the newest-created check, and the JSON census (AC-VTA-010). **No
  other card's SPEC is edited** under any outcome; if the gate is not green, the cutoff rule was
  applied wrongly, and that is what gets fixed.
- **M5 — Mechanical finish (Priority Low).** Mutation probes, coverage, `go vet`,
  `golangci-lint`, `@MX:NOTE` on the rule type.

## §G Anti-Patterns

- Excluding blockquote lines to silence the rule on its own documentation (reintroduces deficit 4).
- Treating any line that contains the dollar-quote sequence as anchored (D15's first-branch miss).
- Using `go test -run` prefixes in this SPEC's own criteria.
- Rebaselining to absorb the corpus instead of the cutoff (DD-3).
- Declaring M3 done from a unit test: the CI command must be executed.
- Checking the grouped form as "starts with `^(`, ends with `)$`" without paren matching (D4).
- Accepting the regex word boundary as a delimiter (D2).
- Editing another card's SPEC to make the landing gate green (D3).

## §H Risks

| Risk | Mitigation |
|---|---|
| Fixture SPEC trips other rules (Tier artifacts, frontmatter), making the red arm red for the wrong reason | the green arm must exit 0; the arms differ by one line, so the red exit is attributable |
| A SPEC created between the M4 measurement and the registering commit is gated at landing | M4 measures on the registering tree itself; AC-VTA-010 re-checks newest `created` < cutoff |
| Base-relative checks fail because `develop` absorption brings other cards' `.github/` or `go.mod` edits | AC-VTA-010/012 diff from `git merge-base HEAD develop`, recomputed at verification time |
| Markdown-table pipe escape misread | explicit conformant and detection table rows in AC-VTA-003 |
| CI lint step runtime (measured 527-571s in the workflow comment) grows | three file reads per SPEC, no process spawn; M4 records the lint wall time before/after |

## §I Cross-References

- `.moai/reports/t1243/decision-scope-reduction.md`, `.moai/reports/t1243/plan-audit-iter{1,2,3}.md`
- `internal/spec/lint_movingref.go` (sibling-artifact precedent), `internal/spec/lint_duplicate_acid.go` (header style)
- `internal/spec/lint_baseline.go`, `.github/workflows/spec-lint.yml`
