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
| **warning, non-advisory only for SPECs created on/after a cutoff** | gated population at landing is the SPECs created on the landing date; REQ-VTA-014 requires it to be 0 | recorded count stays absent (= 0); any new violation in a new SPEC is an increase → red | **chosen** |

Consequences accepted: legacy active SPECs keep their findings as advisory (visible, not gated);
a backdated `created:` would evade the gate — that edit is visible in review and is recorded as
residual risk, not defended in code.

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
test fixture.

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
  `internal/spec/testdata/vacuous_assert_e2e/{red,green}/` (each a minimal project root with one
  SPEC carrying `era: V3R6`, `status: draft`, `created` on/after the cutoff, and a
  `baseline.json` without a `VacuousTestAssertion` entry; the two trees differ by exactly one
  line, blockquoted in `red`). Build the binary from the tree under test and run the CI
  invocation from each fixture root (AC-VTA-009).
- **M4 — Corpus landing measurement (Priority Medium).** Register the rule, set the cutoff to the
  registering commit's date, run the real-corpus baseline gate and the JSON census; any gated
  finding on a same-day SPEC is fixed (it is a real instance) before landing.
- **M5 — Mechanical finish (Priority Low).** Mutation probes, coverage, `go vet`,
  `golangci-lint`, `@MX:NOTE` on the rule type.

## §G Anti-Patterns

- Excluding blockquote lines to silence the rule on its own documentation (reintroduces deficit 4).
- Treating any line that contains the dollar-quote sequence as anchored (D15's first-branch miss).
- Using `go test -run` prefixes in this SPEC's own criteria.
- Rebaselining to absorb the corpus instead of the cutoff (DD-3).
- Declaring M3 done from a unit test: the CI command must be executed.

## §H Risks

| Risk | Mitigation |
|---|---|
| Fixture SPEC trips other rules (Tier artifacts, frontmatter), making the red arm red for the wrong reason | the green arm must exit 0; the arms differ by one line, so the red exit is attributable |
| Same-day SPECs carry gated findings at landing | M4 measures and fixes; REQ-VTA-014 |
| Markdown-table pipe escape misread | explicit conformant and detection table rows in AC-VTA-003 |
| CI lint step runtime (measured 527-571s in the workflow comment) grows | three file reads per SPEC, no process spawn; M4 records the lint wall time before/after |

## §I Cross-References

- `.moai/reports/t1243/decision-scope-reduction.md`, `.moai/reports/t1243/plan-audit-iter{1,2,3}.md`
- `internal/spec/lint_movingref.go` (sibling-artifact precedent), `internal/spec/lint_duplicate_acid.go` (header style)
- `internal/spec/lint_baseline.go`, `.github/workflows/spec-lint.yml`
