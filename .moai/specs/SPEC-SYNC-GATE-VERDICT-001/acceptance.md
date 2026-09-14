# SPEC-SYNC-GATE-VERDICT-001 — Acceptance Criteria

frontmatter-less sibling artifact. All H01 scenarios assume the M1 fixture recipe (temp git
repo under `/tmp`, sync-phase commit subject, detected Go language, non-zero code delta,
blocking mode default) and the measurement contract: every verdict-bearing command redirects
to a file and records its exit code; no pipes on verdict-bearing commands. Every current-tree
claim is an executed command in the claiming run (verification-claim-integrity §1.1).

## §A Scope of Verification

Three headline executions (arms A/B/C) with a positive control against the `2213871af` hook
copy; the doc-pair wording surface (H03 + SX-R05) verified by phrase greps whose patterns
are first proven against the correct pre-change tree per AC-SGV-005/006 (the `2213871af`
baseline doc for the H03 absence pattern; the current pre-M3 tree for the stale-clause
removal patterns — plan-audit F7); pairing and scope-fence checks. Two
release-blocking executions (AC-SGV-001, AC-SGV-006) carry the card's mandatory
execution/wording burden; the remainder are regression-guards and pairing checks.

## §B Measurement contract

For each executed arm, the evidence row records: the command, the redirected output file's
verbatim content, the exit code, and the tree (worktree + HEAD SHA). A second call's output
being "empty" is evidenced by a byte-count of the redirected file (0), not by absence of a
claim.

## §D AC Matrix

| AC | Requirement | Severity | Verification surface |
|----|-------------|----------|----------------------|
| AC-SGV-001 | REQ-SGV-001 | Must (release-blocking) | Positive control on `2213871af` hook + Arm A execution on current tree |
| AC-SGV-002 | REQ-SGV-002 | Must | Arm B (input change re-gate) |
| AC-SGV-003 | REQ-SGV-002 | Must | Arm C (pass suppresses re-delivery) + state-shape read |
| AC-SGV-004 | REQ-SGV-003 | Must | Conditional-repair fence audit (zero-diff or minimal-repair evidence) |
| AC-SGV-005 | REQ-SGV-004 + REQ-SGV-005 | Must | H03 wording greps (baseline-proven patterns) + gate-preservation diff |
| AC-SGV-006 | REQ-SGV-006 | Must (release-blocking) | SX-R05 unification greps, both copies |
| AC-SGV-007 | REQ-SGV-007 | Must | Template-first order + scoped-passage parity + neutrality scan |
| AC-SGV-008 | REQ-SGV-008 | Must | Evidence-path + process fences |
| AC-SGV-009 | REQ-SGV-006 (Phase 8 clause) | Must | Phase 8 additional-lens statement present + stale-clause removal greps, both copies |

### AC-SGV-001 — Same-HEAD failing check re-delivers the stored block (Arm A), with positive control

**Given** the M1 positive control: the `2213871af` hook copy, run twice consecutively against
the failing-check fixture in blocking mode
**When** both calls execute (each redirected to its own file, exit codes recorded)
**Then** call 1's output contains the block decision JSON and call 2's output is empty
(redirected file byte-count 0) — the audit's defect shape, reproduced — and ONLY after that
control is observed, the same two-call sequence against the current-tree hook yields: call 1
block JSON, call 2 non-empty output carrying the identical block decision byte-identically to
call 1, exit 0 on both calls, and the gate log showing no second check execution
(`decision=block-redelivered` or equivalent evidence that the checks did not re-run).

### AC-SGV-002 — Input change under unchanged HEAD re-executes the checks (Arm B)

**Given** the state left by Arm A (same HEAD, recorded failure, stored payload)
**When** the fixture's work-tree input changes (the deliberate fast-check failure is repaired
in the working tree, HEAD unchanged) and the hook is invoked again
**Then** the checks re-execute (fresh execution evidence in the gate log — a new
check-run record, not a re-delivery), the emitted decision reflects the repaired input (allow
→ empty stdout), and the record file now carries the new outcome for the same HEAD.

### AC-SGV-003 — Pass record suppresses re-delivery; state shape holds (Arm C)

**Given** a fixture invocation that passes (all fast checks green) for its HEAD
**When** the hook is invoked again under the same HEAD and unchanged work tree
**Then** stdout is empty (redirected file byte-count 0) and exit 0 — no stale block is
re-delivered; and across Arms A-C the record file always reads `<sha> <running|pass|fail>
<worktree-id>` with the failing run's payload file present (state-shape read recorded in the
evidence). **Boundary (F3):** this AC verifies post-hoc state shape only — it is NOT a
write-ordering proof. The payload-before-fail-record ordering is owned and verified by
SPEC-SYNC-GATE-FAILSTATE-001's torn-write shims (its AC-005 rows, mutants M23/M24), consumed
by REQ-SGV-002; no ordering claim is made here.

### AC-SGV-004 — Conditional-repair fence

**Given** the Arm A-C outcomes on the base tree
**When** the run phase closes the H01 surface
**Then** either (a) all arms passed and `git diff` on the two hook paths is EMPTY across the
whole SPEC, or (b) exactly the red-arm repair landed: a minimal diff applied to both hook
copies in one commit (template first), the pair byte-identical after the repair
(`diff -q` exit 0 re-proven), and the re-run arms green. Any hook change beyond the failing
arm's minimal scope is a FAIL.

### AC-SGV-005 — H03 wording truth and gate preservation

**Given** the M1-proven grep patterns (each ABSENCE pattern demonstrated with ≥1 hit against
the `2213871af` baseline doc copy BEFORE any absence is claimed on the current tree — the
계기 observer contract; the presence patterns' positive control is the current tree, where
they already hold)
**When** the patterns run against both current doc copies and the hook
**Then** the baseline false-promise phrasing ("vulnerability scan runs automatically" family)
has 0 hits in both copies; "not a vulnerability scan" is present in both copies (≥1 each) and
in the hook (≥1); the `deps_modified` mechanism description is present in both copies; and
the SPEC's diff deletes no gate check from either hook copy (the `deps_modified` block and
the per-language fast-check `case` survive verbatim).

### AC-SGV-006 — Severity-to-verdict unified across both copies

**Given** the M1-proven patterns for the superseded trio and the two stale relationship
clauses — where the pattern-proof for the stale-clause removal greps runs against the
CURRENT pre-M3 tree (1 hit each in both current copies; the clauses are absent from the
`2213871af` baseline doc, which predates t624's M3), while the `2213871af` baseline positive
control applies to the H01 hook-state arms and the H03 absence pattern only (plan-audit F7)
**When** the greps run against both doc copies after M3
**Then** "Only CRITICAL findings block", "HIGH findings are reported as warnings", and
"Continue with warning" each have 0 hits in BOTH copies; the two clauses the unified gate
falsifies — "CRITICAL-only stop gate" and "reports only as a warning" — each have 0 hits in
BOTH copies (plan-audit F1 option A: the clauses leave the text, they are not grep-watched
surviving); "Continue by approved exception" is present in both; each copy's Step 0.55.2
carries the single verdict mapping (Critical and High block; Medium and Low advisory;
user-approved exception record fields: finding ID, rationale, scope, approver, expiry,
review condition); and the template copy's edited passages reference no rule file absent
from the distributed template (inline contract, not a `security-decision-contract.md`
path).

### AC-SGV-007 — Template-first order, scoped-passage parity, neutrality

**Given** the M3 edit set
**When** the pairing and neutrality checks run
**Then** (a) the hook pair is byte-identical (`diff -q` exit 0); (b) the doc pair's SCOPED
passages (Step 0.55.1 severity application + Step 0.55.2 decision block + Phase 8
relationship paragraph) agree semantically, verified by this MECHANICAL PROXY (plan-audit
F4): each copy's scoped passage is extracted to a file, the one documented delta sentence
(the local copy's rule-path reference) is excluded, whitespace is normalized, and `diff` of
the two normalized files exits 0; (c) the template doc's edited passages contain no internal
SPEC IDs, no audit citations, no internal dates, and no commit SHAs (grep 0 hits over the
diff-added lines); (d) no whole-file copy occurred (the doc diff touches only the scoped
passages — the fenced divergence passages of spec.md §B show zero diff lines); (e) on the
hook copies, diff-ADDED lines carry no internal card IDs (plan-audit F5: pre-existing
citations on untouched lines are preserved verbatim and exempt).

### AC-SGV-008 — Evidence path and process fences

**Given** the whole run phase
**When** the close audit runs
**Then** all measurement evidence lives under `.moai/reports/t783/` as UNTRACKED files in
the primary checkout (nothing under that path is committed); no `make build` appears in the
run evidence; every verdict-bearing command's evidence row carries command + verbatim output
+ exit code + tree, in file-redirect form; and the SPEC's commits contain no `.moai/reports/`
paths.

### AC-SGV-009 — Phase 8 classified as an additional lens, stale clauses gone

**Given** both doc copies after M3 (plan-audit F1 option A)
**When** the Phase 8 relationship statement is read and grepped in each copy
**Then** each copy states that the sync-auditor rubric (Step 0.5.4, Critical/High → FAIL) is
canonical, that Phase 8 is an additional lens, and that its stop gate never clears an earlier
sync-auditor FAIL (≥1 hit of "sync-auditor FAIL" in both copies, in the canonical-direction
sentence) — AND the relationship paragraph no longer carries the two clauses the unified
gate falsifies: "CRITICAL-only stop gate" and "reports only as a warning" each have 0 hits
in both copies (same removal greps as AC-SGV-006; the alignment is the fix, the greps are
its evidence — not a guard over surviving text).

## §E Definition of Done

- AC-SGV-001..009 all PASS with executed evidence (or AC-SGV-004 branch (b) completed with
  re-run arms green).
- `.moai/specs/SPEC-SYNC-GATE-VERDICT-001/progress.md` §E.2/§E.3 populated with the
  attribution triple (command / verbatim output / baseline-attribution) per headline AC.
- No scope-fence violation recorded; if one is attempted and refused, it is reported as a
  blocker, never silently absorbed.
