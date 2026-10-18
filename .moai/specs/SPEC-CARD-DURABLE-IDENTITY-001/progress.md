# Progress: SPEC-CARD-DURABLE-IDENTITY-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-CARD-DURABLE-IDENTITY-001
card: t1842
tier: L
base_sha: f4bcd4f2cc65342559b505f1e5a29df80f992894
spec_version: 0.1.0
req_count: 22
ac_count: 24
decision_rows: 7
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md
plan_status: audit-ready
plan_complete_at: 2026-10-18
plan_artifact_hash: f48a8af397477554d847506010280f9db6a45ba9a5b084ec5922eec91af203a4
release_target: v3.2.0
audit_ready: true
```

What the token means here: the plan-phase artifact set is complete and ready to be audited. **No plan audit has run on this SPEC.** The verdict files go to `.moai/reports/t1842/plan-audit.md` (single iteration) or `.moai/reports/t1842/plan-audit-iter<N>.md`, carrying the machine lines `verdict`, `audited_sha`, `overall_score`, `must_pass_failed`, `blocking_count` and `plan_artifact_hash` (`.moai/docs/audit-artifact-convention.md`). The token was written at the lane's instruction before any audit; it is not an audit result, and a Kickoff still needs a PASS verdict file whose `plan_artifact_hash` equals the one above.

Plan-phase facts (manager-spec, card t1842): 22 requirements and 24 acceptance criteria against the Tier L ceilings of 25 and 25; seven decision rows, all `DEFAULT-APPLIED` and none product-level; the decision-index stamps (2026-10-18T00:53:21Z) were written after the last edit of every other hashed file. The plan-artifact hash binding is plan.md § E. Stage A (card t1841) is not a prerequisite and its SPEC is not touched; the seam is spec.md § E.

### Plan-phase verification record (this run, tree B0 plus the uncommitted SPEC directory)

- SPEC id check, executed: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS` for `SPEC-CARD-DURABLE-IDENTITY-001`; `ls .moai/specs | grep -c "SPEC-CARD-DURABLE-IDENTITY-001"` printed `0` before the directory was created.
- Counts, by `grep -c`: 22 lines `^- \*\*REQ-CDI-` in spec.md; 24 lines `^### AC-CDI-` in acceptance.md; 7 lines `^### Q` and 7 lines `^Operator verdict: DEFAULT-APPLIED` in decision-index.md.
- `moai spec lint SPEC-CARD-DURABLE-IDENTITY-001` and the same with `--strict` each printed `No findings — all SPEC documents are valid` (exit 0), run before progress.md and decision-index.md existed.
- Traceability, by a script of my own over the files (the agent definition's awk was not run for this one): 22 requirement definitions collected, 0 uncovered, 0 orphan, 24 acceptance headings. Ordering verb (the awk of the plan-auditor definition, copied to the scratchpad): `COLLECTED: 8 milestones in plan order (M0 M1 M2 M3 M4 M5 M6 M7), 0 exit bindings, 30 ordering candidates`, no `CONFLICT:` line; the candidates are ordering words in test descriptions, none binds a criterion to a milestone.
- Scope guards E23a and E23b printed nothing on B0; their red was observed on a constructed history (acceptance.md § Guard measurements).
- Every RED-now ledger entry E00 to E24b was run on B0 and its stdout and exit code recorded in acceptance.md. The SQLite behaviors the design rests on were observed on the SQLite command line only (research.md § 6). The live-database measurements are on a copy taken 2026-10-18T00:31:45Z (research.md § 4).
- Tool provenance (verification-claim-integrity § 2.2): the lint ran on the installed `moai` v3.2.0-rc.36-t1822c, built from commit `ee981a932`; `git merge-base --is-ancestor ee981a932 HEAD` exited 0, so that build is a strict ancestor of the tree HEAD and lags it. The lint and the `mcp__moai__spec_audit` results are attributed to a lagging build and are advisory for this tree.
- Plan-artifact hash `f48a8af397477554d847506010280f9db6a45ba9a5b084ec5922eec91af203a4`: computed by a Python reimplementation (`python3 -I`, hashlib) of the algorithm read in `internal/runtime/audit_cache.go:84-141` (names in the order acceptance, decision-index, design, plan, research, spec, tasks; present files only; `<name>:<length>`, a zero byte, the bytes, a zero byte). Sizes in bytes: acceptance.md 46417, decision-index.md 7454, design.md 27245, plan.md 13916, research.md 13580, spec.md 34849; tasks.md absent. The runtime `ComputeHash` was not executed.

### Gaps

- No `go build` and no `go test` was run (the host load average read 20.63 at 09:33). No behavior of the planned code was exercised, and the four existing retire tests of AC-CDI-002 were not observed to pass.
- The project's SQLite driver was not exercised; view reuse of the legacy name, `INSTEAD OF` triggers and the three immutability triggers were observed on SQLite 3.54.0 (command line).
- The stage-A and card-t1852 facts come from unmerged worktrees (t1841, t1852) and, for the t1852 audit result and the t1825 split, from memory records; they are re-read at run entry (plan.md § B).
- Stage A's iteration 1 to 3 defect bodies were read as headlines and opening sentences only (research.md § 3). Cards t1835 and t1843 have no SPEC that was read.
- The conflict rule (spec REQ-CDI-005, decision-index Q7) reads the card's "conflicting active owners" as non-terminal rows in two or more active runs; this is an interpretation recorded as a decision row, to be confirmed by whoever reads the card text.
- Two commands were refused by the worktree-session guard while authoring (a loop assembling a path in a variable; a `cd` to the shared checkout before `git`) and were rerun in plain form (acceptance.md § Guard measurements).

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
