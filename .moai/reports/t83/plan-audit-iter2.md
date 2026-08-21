# SPEC Review Report: SPEC-CODEX-HOOK-ADAPTER-001 (Iteration 2 — delta-scoped)

Iteration: 2/2 (Tier M ceiling — final round within the retry cap)
Verdict: **PASS**
Overall Score: **0.92** (harmonic mean; Tier M PASS threshold 0.80; iteration 1 was FAIL 0.63 — score improved, no STOP signal)

Audited tree: worktree `.claude/worktrees/t83`, branch `WT-hook-adapter`, fixed commit `3556ca1de`
(verified at start and re-verified at end of audit — unchanged; worktree clean at start). Branch
commits `main..HEAD` (4) touch only `.moai/reports/t83/**` and `.moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/**`
— no code files, consistent with REQ-7's premise. Artifacts audited (Tier M trio): spec.md (295L,
v0.2.0), plan.md (145L), acceptance.md (169L). Reasoning context ignored per M1 Context Isolation;
only the SPEC artifacts and the measured-evidence files they cite were used. Every claimed fix below
was verified against evidence re-derived by this auditor, not taken from the author's claim.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -c '^### REQ-' spec.md` → 7; REQ-1..REQ-7
  sequential, no gaps, no duplicates (spec.md:L139, L162, L175, L185, L196, L203, L209).
- **[PASS] MP-2 EARS/GEARS format compliance** (judged against the requirement layer, spec.md §C):
  REQ-1/2/3/4/5/6 use `WHERE [condition], the <subject> SHALL [response]` (incl. REQ-3's
  `WHERE the underlying hook exited 2, the adapter SHALL NOT …` — valid negative form under a WHERE
  gate; REQ-4 combines SHALL + SHALL NOT under one WHERE, matching the measured asymmetry);
  REQ-7 is the Ubiquitous form "The adapter SHALL … and the change set SHALL …" (L209-212).
  All 7 match GEARS patterns. The Given-When-Then entries in acceptance.md §D.2 are ACs
  (verification layer) and are correctly not penalized here.
- **[PASS] MP-3 YAML frontmatter validity** — judged by running the domain tool:
  `moai spec lint .moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/spec.md` (worktree) →
  `✓ No findings — all SPEC documents are valid`, exit 0. All 12 canonical fields present with
  correct types (spec.md:L2-16); `created`/`updated`/`tags`/`id` canonical names; version quoted
  semver `0.2.0`; `status: draft` in enum; `phase: "v3.2"` is a release label (not a prohibited
  stage token); optional `era: V3R6`, `tier: M` valid.
- **[PASS/N/A] MP-4 language neutrality** — N/A: single-language scoped SPEC (Go codebase,
  codex-cli target); no 16-language enumeration obligation. Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md
  | sort -u` → only the SPEC's own ID; `related_specs: []`. t83/t88/t91 are card ids, not SPEC-IDs.
  No reconciliation duty, no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → 0 → auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md` → no matches
  (research.md does not exist at Tier M; plan.md checked clean).

Must-pass firewall: clean. The PASS verdict additionally rests on all five iteration-1 blocking
defects being verified resolved (below).

---

## Defect Dispositions (iteration-1 D1–D9 + AC budget) — all verified against re-derived evidence

**D1 (REQ-1 count/counterpart) — RESOLVED.**
REQ-1 now carries the full 11-row table (spec.md:L145-157): Codex event → dispatcher argument →
adapted flag. Cross-checked against the actual dispatcher in the fixed tree
(`internal/cli/hook.go:46-71`): exactly **26** registered subcommands, and all **eleven** table
arguments are real — `pre-tool`(:47), `post-tool`(:48), `session-start`(:46), `session-end`(:49),
`stop`(:50), `user-prompt-submit`(:55), `compact`(:51, EventPreCompact), `post-compact`(:62),
`permission-request`(:56), `subagent-start`(:54), `subagent-stop`(:59). Six adapted / five
recognized-then-refused; the stated counts ("Eleven Codex events, eleven dispatcher counterparts,
six adapted", L159) match the enumerated table exactly. §B is rescoped as measurement-coverage
scoping with an explicit falsified-claim retraction: "An earlier draft asserted the absence; that
was false" (L120-121). AC-REQ-1a asserts the table row count against a constant (11) and folds in
the mechanical registration cross-check against `hook.go` (acceptance.md:L36-48).

**D2 (golden-fixture availability) — RESOLVED.**
`git ls-files` on the branch shows 8 payload files under
`.moai/specs/SPEC-CODEX-HOOK-ADAPTER-001/testdata/hook-payloads/` (six event payloads + two
`subagent-*` captures) plus a provenance `testdata/README.md`. All 8 are **byte-identical** to the
primary `.moai/reports/t91/hook-payloads/` originals (diffed file-by-file by this auditor). REQ-6
(L204-207) and AC-REQ-6 (acceptance.md:L137-144) point at the vendored tracked path; plan.md §C
pre-flight 2 requires ≥6 tracked payloads and explicitly forbids depending on the untracked t91
directory; plan §G carries the corresponding anti-pattern. The fixtures now resolve in a worktree
and in CI.

**D3 (Out-of-Scope convention) — RESOLVED, with independently reproduced negative control.**
Four `### Out of Scope — <topic>` H3 headings each with `-` bullets (spec.md:L103, L111, L117,
L129). Verified mechanically in this audit: `moai spec lint` on the revised spec →
`✓ No findings` (exit 0); negative control — a /tmp copy with the headings reverted to the old
bold-lead-in format → `ERROR MissingExclusions … 'Out of Scope' section has no items — minimum one
item required`. The author's negative-control claim is reproduced exactly.

**D4 (AC-REQ-5 vacuity) — RESOLVED.**
REQ-5 is re-aimed at "the validator this SPEC ships" (L197-201; M5 delivers it, plan.md:L109-112
explains why a validator rather than prose). AC-REQ-5a's negative specimen is the measured killer:
the exact shape from `probe/run-versionkey-kills-file.jsonl`, whose error item this auditor
re-read in the fixed tree — `"unknown field 'version', expected 'description' or 'hooks' at line 2
column 11"` — so the top-level whitelist {description, hooks} is measured whitelist evidence, not
invention. AC-REQ-5b adds per-level rejection with named key+level. The per-level whitelists
{matcher, hooks} / {type, command, timeout} are grounded in the round-2 working-config shape
(`precondition-measurement.md` L84: `{ "hooks": { "PreToolUse": [ { "hooks": [ { "type": "command",
"command": "…", "timeout": 10 } ] } ] } }`; L87 records the `timeout`-vs-`timeoutSec` distinction;
round-3 L91: matcher → fires). Both ACs are now failable against a subject that exists in scope.

**D5 (Finding D accuracy) — RESOLVED.**
Finding D restated as "the failure **is** reported, but only as an item in the `--json` event
stream … There is no interactive warning and the process still exits 0" with the error quoted
(§A, L66-80); §E adds the two-failure-modes distinction ("Only the second is silent, and only it
justifies a 'nothing reports this' framing", L251-253); §F's "nothing reporting that" is now
confined to the project-level non-firing case (verified: `run-projectlevel-nofire.jsonl` carries
no error items). Retraction banners verified in the measurement report:
`precondition-measurement.md` L7 ("정정 (3차 실측 후) … 철회") and L11 ("**정정 2 (plan-audit D5
수용)**: `version` 키 실패를 '조용히/경고 없이'로 적은 서술도 틀렸다") — a correction banner that
explicitly accepts audit finding D5.

**D6 (REQ-3 sink) — RESOLVED.**
Sink named: `.moai/logs/codex-adapter.jsonl` (REQ-3, L179); the second WHERE clause forbids
stderr writes on the exit-2 path with the REQ-4 interaction stated (L181-183); new MUST-PASS
AC-REQ-3c (acceptance.md:L99-105) makes the no-stderr rule binary-testable; plan D4 records the
exfiltration rationale for length-not-content.

**D7 (REQ-4 evidence tiers) — RESOLVED.**
REQ-4's normative classes are the two measured ones only (PreToolUse blocking reason, Stop
continuation prompt); `UserPromptSubmit` carries an explicit declared-not-measured annotation;
"Events excluded from adaptation by §B receive no stderr class here" (L186-194) removes the dead
PermissionRequest/SubagentStop branches. AC-REQ-4c verifies annotation presence and §B-event
absence. The related AC-REQ-2b overstatement is also corrected — "the binary carries that error
string for `Stop` and `SubagentStop`, the two events where it was located" (acceptance.md:L69).

**D8 (REQ-7 scope) — RESOLVED.**
REQ-7 pins the adapter package outside `internal/hook/` and states the invariance mechanically:
"zero modifications to files that existed under `internal/hook/` before this SPEC" (L209-212).
AC-REQ-7 is mechanical: `git diff --name-only <base>..HEAD -- internal/hook/` intersected with the
base commit's file list, count must be zero (acceptance.md:L146-152) — no judgment call. The three
divergent phrasings are collapsed; §A now frames the card premise as an assumption, not a claim.

**D9 (AC-REQ-3b enforcement) — RESOLVED.**
AC-REQ-3b binds the discard-branch enumeration to "a branch-count constant declared beside the
implementation" (acceptance.md:L91-97), making an untested added branch a compile-or-assert
failure; plan M3 wires the constant into the milestone that ships the discard paths.

**AC budget — RESOLVED.** `grep -c '^| AC-' acceptance.md` → **16** (= Tier M cap 16, no longer
17). The registration cross-check was folded into AC-REQ-1a rather than dropped. REQ count 7 ≤ 16.
REQ-7 is covered by AC-REQ-7 (MUST-PASS).

---

## Regression Check

Iteration-1 defects D1–D9: **all RESOLVED** (dispositions above, each with re-derived evidence).
No iteration-1 defect is unresolved, so no automatic FAIL applies. Two new minor observations
surfaced by the revision, neither blocking:

- **R1 (wording tension, zero decision surface)** — spec.md §B says of the four unmeasured events
  "No payload capture and no behavioral observation exists for any of them" (L122-124), while the
  SubagentStop paragraph and `testdata/README.md` record M0's observation that "`SubagentStart` /
  `SubagentStop` never fire in this build" — an observation of absence that touches SubagentStart,
  which sits in the unmeasured list. The operative disposition is identical either way and is
  already enumerated per-row in REQ-1's table (both readings → recognized-and-refused), so no
  engineer can diverge. — Severity: minor — Class: optional.
- **R2 (disclosed conditional in REQ-2)** — "for events where `decision: "block"` is honored"
  (L164-167) leaves the honored-event set implicit (measured instance: Stop only, per Finding B).
  §E discloses exactly this ("REQ-2's event coverage is a range to confirm at run-phase, not an
  established fact", L248-250) and plan §B/§C make widening it a run-phase gate, so the openness
  is deliberate and delegated rather than accidental. — Severity: minor — Class: optional.

No blocking regressions. The vendored `subagent-*` fixtures beyond REQ-6's six events are
documented provenance (they are the evidence for the SubagentStop retirement) and are not a defect.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | Minor ambiguity in one or two requirements that a reasonable engineer would resolve consistently | The three iteration-1 blockers are gone: REQ-1's count reconciled by an enumerated 11-row table verified against hook.go; REQ-3's sink named; REQ-7 single mechanical scope. Remaining: R1 (§B "no behavioral observation" vs M0's never-fire observation for SubagentStart) and R2 (REQ-2's honored-event set implicit) — both disclosed/bounded, both resolving to a single implementation via the REQ-1 table and the run-phase measurement gate. |
| Completeness | 1.00 | All required sections present; frontmatter complete; Out-of-Scope in project convention | HISTORY (§H two entries incl. revision record), WHY (§A five evidence-attributed findings), WHAT/scope (§B), REQUIREMENTS (§C, 7 REQ), AC (acceptance.md §D with severity split + residual risk §D.3), four `### Out of Scope` H3 sections with bullets; frontmatter lint-clean (mechanical, this audit); evidence §D attributes every §A claim to a tracked probe artifact; goldens vendored with per-file provenance README. |
| Testability | 1.00 | Every AC binary-testable; no weasel words | All four iteration-1 testability defects fixed and verified: AC-REQ-6 executable from tracked byte-identical fixtures; AC-REQ-5a/5b failable against the shipped validator with the measured `version` specimen; AC-REQ-7 mechanical zero-count diff check; AC-REQ-3b branch-count constant; plus new AC-REQ-3c. Count-asserted table (constant `11`), byte-identical pass-through, marker-grep live checks with exit-code-is-never-evidence stated up front. Budget 16/16 AC, 7/16 REQ. |
| Traceability | 1.00 | Full bidirectional coverage | Every REQ has ≥1 AC (1:1a/1b; 2:2a-2d; 3:3a/3b/3c; 4:4a/4b/4c; 5:5a/5b; 6:6; 7:7 — §D matrix); every AC maps to an existing REQ; §D.1 documents the MUST-PASS/SHOULD split; no orphans either direction. |

Harmonic mean = 4 / (1/0.75 + 1/1.00 + 1/1.00 + 1/1.00) = **0.92** ≥ 0.80 → **PASS**.
Iteration-over-iteration: 0.63 → 0.92 (improvement; no STOP escalation).

---

## Recommendation

PASS. All seven must-pass criteria hold with mechanical evidence (lint executed, counts grepped,
dispatcher registrations read, fixtures diffed, probe error re-read), and all nine iteration-1
defects plus the AC-budget breach are verified resolved — five of them against facts this auditor
re-derived rather than accepted (hook.go's 26 subcommands and the 11 argument targets; the 8
byte-identical vendored goldens; the lint negative control; the probe's `version` error item and
its `{description, hooks}` whitelist; the round-2 report's D5-acceptance correction banner and the
working-config shape grounding REQ-5's per-level whitelists). The two residual observations (R1,
R2) are optional-class with no implementation divergence and are left to the author's discretion;
they do not gate run-phase entry. Standard gate reminder: this PASS feeds the Plan Audit Gate's
skip-eligibility only — Implementation Kickoff Approval remains mandatory and score-independent.
