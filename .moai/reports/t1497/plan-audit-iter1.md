auditor-model: glm-5.3-flash[1m]

# SPEC Review Report: SPEC-HARNESS-DETACHED-PRUNE-001 (card t1497)
Iteration: 1/2 (Tier M ceiling)
Verdict: FAIL
Overall Score: 0.88
Plan Artifact Hash: d6df92aa64fd00c5cde347b0a849c2e3a07e2faf6d127860f97cc6e3fc7e0709
Auditor Version: plan-auditor/v1 (GLM lane)

verdict: FAIL
audited_sha: 080578aeca7285426dbe0a31b65633ef60dbf78f
overall_score: 0.88
must_pass_failed: 0
blocking_count: 1
plan_artifact_hash: d6df92aa64fd00c5cde347b0a849c2e3a07e2faf6d127860f97cc6e3fc7e0709
scope: full
fix_scope: .moai/specs/SPEC-HARNESS-DETACHED-PRUNE-001/acceptance.md#AC-DP-001-green-cell
defect_class: ac-wording (per blocking finding D1)
reread_hunks: .moai/specs/SPEC-HARNESS-DETACHED-PRUNE-001/acceptance.md#AC-DP-001-green-cell

Header facts: auditor model glm-5.3-flash[1m] (GLM lane, stated per delegation); tree SHA `080578aeca7285426dbe0a31b65633ef60dbf78f` (branch `WT-harness-prune-detached`, HEAD re-read this run); artifacts commit `080578aec` (the audit-read commit; `git show --stat` verified: 4 files, 285 insertions, `Authored-By-Agent: manager-spec`); date 2026-10-05. `plan_artifact_hash` computed with the `internal/runtime/audit_cache.go` `ComputeHash` algorithm replicated over the Tier M subject set (acceptance.md, plan.md, spec.md; 3 files hashed) — formula: per file in `planArtifactNames` order, `fmt.Fprintf(h, "%s:%d\x00", name, len(data))` + bytes + `\0`, hex-encoded sha256.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-DP-001..008 sequential, no gaps, no duplicate definitions (`grep -o "REQ-DP-00[0-9]" spec.md | sort | uniq -c` — all 8 present; multiple hits per id are cross-references, one definition each in spec.md §B).
- [PASS] MP-2 EARS/GEARS format compliance (judged against the requirement layer in spec.md §B): all 8 REQs carry labeled GEARS forms — Ubiquitous (REQ-DP-001/002/003/005/007/008), Event-driven "When the spawn fails" (REQ-DP-004, spec.md:46), Where-gate "Where the platform is Windows" (REQ-DP-006, spec.md:50). REQ-DP-001/007/008 embed "shall not/shall never" negative clauses inside Ubiquitous shells — counted as legacy-EARS negative equivalents under the M3 Score-1.0 allowance, never as a fifth pattern.
- [PASS] MP-3 YAML frontmatter validity: all 12 canonical fields present with correct types (spec.md:2-13; `phase: "v3.2.0 target"` is a release label, not a stage name; optional `tier: M` + `related_specs` permitted). No snake_case aliases. Re-run claim verified: `go run ./cmd/moai spec lint SPEC-HARNESS-DETACHED-PRUNE-001 --strict` → `✓ No findings — all SPEC documents are valid`, exit 0 (this run, this tree, source-built binary).
- [N/A] MP-4 Section 22 language neutrality: single-language Go scope (`internal/harness`, `internal/cli`) — N/A auto-pass.
- [PASS] MP-5 D7 cross-SPEC reconciliation: one referenced SPEC, SPEC-V3R3-HARNESS-LEARNING-001; exists at `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`, `status: completed` (grep measured) — not retired/superseded/archived, so no D7-4 trigger; the REQ-HL-011 amendment relationship is declared explicitly regardless (spec.md §A para 4 + §C C6). No SHOULD findings.
- [PASS] MP-6 D8 cross-platform discipline: `syscall` literal appears in spec.md only in §C C4, whose own paragraph carries the `//go:build !windows` / `//go:build windows` constraint (awk verb: no output on spec.md). plan.md §E tripped the section-scoped verb; adjudicated non-blocking by reading — the §E sentence names file 5 and cites C4 by name, §B rows 4-5 of the same document carry the literal build tags, and the risk D8 guards is structurally closed by C4-mandatory split + GOOS=windows build gating every milestone exit + this run's measured green baseline (`GOOS=windows GOARCH=amd64 go build ./...` exit 0). Recorded as advisory D5.
- [PASS] MP-7 clarification gate: `grep -rn "\[NEEDS CLARIFICATION" .moai/specs/SPEC-HARNESS-DETACHED-PRUNE-001/` → no matches. research.md absent — expected at Tier M (N/A for that file).
- [PASS] MP-8 RED-now cell re-execution: all 8 ledger cells (LEDGER-DP-A..H) re-executed this run against the current tree, all commands conforming single invocations, all REDs reproduce: A `2`/exit 0; B `ok ... 0.791s` no test names; C no output/exit 1; D `ok ... 0.520s` no names; E both files absent/exit 1 (verbatim match); F no output/exit 1; G `Unknown flag: --log.` exit 1 (verbatim match); H `ok ... 1.363s` no names. Timing tokens vary as expected; substantive outputs match. `-list` corroboration cells were judged on the count of tests actually executed (zero names listed = genuinely-new tests), not on the `ok` token. Carrier is the §2.1-sanctioned fenced evidence ledger. (Note: MP-8 governs the RED side; blocking defect D1 below sits on the GREEN side of AC-DP-001 — a two-cell adoption defect outside MP-8's scope, caught by AC-6/§2.)
- [PASS] MP-9 cross-artifact ordering consistency: CN-4 verb output — `COLLECTED: 4 milestones in plan order (M1 M2 M3 M4), 0 exit bindings, 3 ordering candidates`; zero CONFLICT lines. The 3 candidates read by hand: acceptance.md:95 ("stamp written before the work" — algorithm-internal ordering, not milestone ordering), :98 ("no input turns it red before the work" — RED-now semantics, not ordering), :102 ("recorded ... before run-phase exit" — aligns with M4-last). `0 exit bindings` is a verb-form mismatch (plan uses `**Exit (count-first):**` headings, the verb binds lowercase `Exit:`); milestone order was collected, and the AC GREEN cells' milestone attributions (M1/M2+M3/M3/M2+M4/M3/M1+M2) are consistent with the declared M1→M2→M3→M4 order read by hand.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Minor ambiguities: F1's spawn-rate quantifier overclaims (spec.md:89, D2); REQ-DP-006 vs plan file-5 flag divergence (spec.md:50 vs plan.md:17, D3); stamp-read-error path unspecified in REQ-DP-002 (resolved consistently by any engineer via the child-side re-check) |
| Completeness | 1.0 | 1.0 | All sections present: §A background, §B 8 REQs, §C 6 constraints, §D 6 decisions, §E Out of Scope H3 with 6 specific bullets, §F 5 risks, §G history; frontmatter complete (12/12) |
| Testability | 0.75 | 0.75 | 6/7 ACs binary-testable with sound executable cells; AC-DP-001's GREEN command is non-functional as written (D1 — reproduced); no weasel words anywhere in acceptance.md; RED side immaculate (8/8 reproduced) |
| Traceability | 1.0 | 1.0 | Verb output: `COLLECTED: 8 REQ definitions (acceptance input: read)`, zero UNCOVERED, zero ORPHAN; mapping table hand-verified (REQ-DP-002/007→AC-DP-006, REQ-DP-003/005→AC-DP-003/004, all 8 covered) |

Aggregate 0.88 = mean of the four dimensions; above the Tier M threshold (0.80) — the FAIL verdict comes from the blocking finding, not the score (M5/M6 separation).

## Defects Found (structured defect-list)

D1. BLOCKING — .moai/specs/SPEC-HARNESS-DETACHED-PRUNE-001/acceptance.md:92 — AC-DP-001's GREEN cell carries `go test -run '^(TestRecordExtendedEventDoesNotPrune\|TestRecordEventDoesNotPrune)$' ./internal/harness`; in a markdown table cell the pipe is escaped as `\|`, and in Go regexp `\|` is a LITERAL pipe, not alternation — executed verbatim (the §2.1 verbatim standard) the command matches zero tests and exits 0 even after the tests exist: a vacuous green gate on a release-blocking AC. Reproduced this run by positive control: `go test -run '^(TestPruneStaleEntriesRemovesOldEvents\|TestRecordEvent)$'` → `ok ... [no tests to run]` (zero matched) vs the bare-`|` form → runs the existing test. Independently found by the codex backend (adversarial, its overlay experiment: two failing tests injected, verbatim command ran zero of them, exit 0). Severity: critical — Class: blocking — Required fix: move the GREEN commands out of the table into a fenced evidence-ledger entry (the §2.1-recommended carrier — a table cell cannot carry a bare `|`) or drop the inline alternation from the cell (name the tests; keep the command form the plan's M1 exit uses). Same class check on the other GREEN cells: clean — only this cell carries an inline regex (plan.md:44's body-text `-run` pattern correctly uses bare `|`).

D2. ADVISORY — spec.md:89 (§F F1) — "at most once per hour per project" overclaims the spawn rate. The stamp is written by the CHILD after it wins the lock (retention.go:195); a stamp-write failure (retention.go:195-196 skips the prune and leaves the stamp absent) or a slow first child leaves the gate's view stale, so every subsequent event in the interval spawns another child — each spawn collapses at the lock re-check (the WORK is once per interval; the SPAWN rate is not). SPEC §E exclusion 2 already declares the Windows burst but not the POSIX stamp-write-failure burst. Severity: minor — Class: optional — Required fix: correct F1's quantifier ("once per interval when the child's stamp write succeeds") and name the stamp-failure/slow-child repeat-spawn residual in §E or F1.

D3. ADVISORY — spec.md:50 (REQ-DP-006) vs plan.md:17 (file 5) — the requirement normatively names `CREATE_NEW_PROCESS_GROUP|DETACHED_PROCESS` while the plan substitutes `CREATE_NEW_PROCESS_GROUP|CREATE_NO_WINDOW` with "DETACHED_PROCESS semantics via the creation-flags combination" and a run-start hedge. A cross-layer wording divergence (verification-completeness.md §3 sweep form): an implementer following the plan can ship a combo missing the bit the REQ names, and nothing in build+vet checks the bit. Severity: minor — Class: optional — Required fix: align the REQ to the semantic contract ("detached, own process group, no console flash — exact Windows console-suppression flags finalized at run start per C4") or pin the plan to the REQ's combo. (DETACHED_PROCESS and CREATE_NO_WINDOW are alternative console-suppression modes — the implementer must pick per the pinned Go constants, not blindly both.)

D4. ADVISORY — acceptance.md:95 (AC-DP-004) / spec.md:48 (REQ-DP-005) — parent-exit-without-wait / child-reparenting has no executable AC: AC-DP-004 covers double-spawn collapse only, and REQ-DP-007's no-real-child policy (house rule, correct) forecloses the runtime experiment that would observe detachment. The property is asserted by construction (no `Wait` call) — an honest, declared trade, but not stated as such in the AC. Severity: minor — Class: optional — Required fix: one clause in AC-DP-004 or §D declaring the construction-only verification basis (or, optionally, a bounded cleanup-guaranteed process experiment consistent with AGENTS.md §4).

D5. OBSERVATION — plan.md:62 (§E) — `syscall` mentioned in a section without an inline `//go:build` mention; the constraint lives in the same document's §B rows 4-5 and in spec.md §C C4, which the sentence cites. Cosmetic. Severity: minor — Class: optional — Required fix: name the build tags in the §E sentence.

(Iteration 1 — no prior-iteration defects. Regression check section: N/A.)

## Open Questions Adjudicated (per delegation)

- **Q1 — Windows CreationFlags hedge**: ACCEPTABLE in mechanism, with the D3 wording fix. Constant availability genuinely depends on the pinned Go version's exported syscall surface — the one axis unverifiable from this lane (no Windows host; runtime half declared unobserved per F3/AC-DP-005). Pinning a constant combo now would create a constraint nothing on this tree can falsify. But the hedge must not leave the REQ naming a bit the plan may drop — apply D3's alignment.
- **Q2 — AC-DP-005 split**: SOUND. Build+vet is executable and gateable on this lane (baseline measured green this run); the runtime half has no observable RED here, and per verification-completeness §2.1's undecidable disposition it correctly loses release-blocking eligibility and is declared unobserved rather than fabricated. Exactly the honest handling.
- **Q3 — progress.md §E.1 `pending-audit` flip**: NOT the auditor's — the auditor owns only the verdict file; §E.1 is owned by manager-spec per the progress.md Section Map. The authoring lane flips `plan_status: audit-ready` after consuming the verdict and pinning hashes — and on this FAIL verdict the lane must NOT flip; it routes D1's fix first, re-audits (reread_hunks above), and flips only on the confirming PASS.

## Cross-Model Convergence Record (backends named per delegation)

- Plan (`moai verify audit-plan`): `config_status: ok`, `cross_model_required: true`, enforced_required `[codex]`.
- `mcp__moai__codex_audit` (adversarial, this tree): **verdict fail**, 4 findings (P2 ×4) — its central finding (the `\|` vacuous selector) independently CONFIRMED by this auditor's positive control; its other three findings adopted as D2/D3/D4 with my own severity/classification. Required gate MET (answered; `convergence_check.ok: true`, `unmet: []` after digest write + re-verify).
- `mcp__moai__audit_multi` (no anchor — I am a GLM session, not a Claude main session): refused to synthesize, zero backends run — recorded, no synthesis falsely claimed.
- `mcp__moai__glm_audit` (probe, uncommittedChanges): inconclusive — empty diff (artifacts are committed; my probe-target error, recorded as such).
- `mcp__moai__codex_audit` (native, baseBranch): rejected by codex (`missing field 'branch'`) — recorded as a launcher-surface gap; the adversarial surface carried the review instead.
- No `audit_receipt` id returned on the codex result → cited as `receipts=none`.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: SPEC-HARNESS-DETACHED-PRUNE-001 at 080578aec is high-quality (0.88) but not audit-ready: one release-blocking AC's executable green gate is vacuous as written.
- **Evidence**: all commands and outputs cited inline above (8/8 ledger reproductions; lint strict 0 findings; trace verb COLLECTED 8/0/0; CN-4 COLLECTED 4 milestones/0 conflicts; positive-control pair `[no tests to run]` vs test-executed; cross-compile exit 0).
- **Baseline-attribution**: every figure measured this run, this tree (`080578aec`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1497`), source-built `go run ./cmd/moai` (tool provenance: build made from the measured tree).
- **Gaps**: parent verdict t1467 §4 J2 + §5 M2 not readable here (`.moai/reports` disk-only) — the design baseline was taken from the inlined delegation text, as instructed; GOOS=windows `go vet` and golangci-lint not re-run (GREEN-path obligations for run-phase M4, not RED cells); Windows runtime behavior unobserved (declared by the SPEC itself, F3).
- **Residual-risk**: my positive control proves the regex semantics, not the post-implementation run; codex's overlay experiment (failing tests injected, verbatim command ran zero) is the post-implementation shape and I adopt it on the strength of the identical mechanism both measurements exhibit. The plan-artifact hash will change with the D1 fix — the reread is scoped to the named hunk.

## Recommendation (numbered fix route for manager-spec)

1. acceptance.md §C AC-DP-001 GREEN cell: relocate the two `go test` commands into a fenced ledger entry (e.g. `LEDGER-DP-GREEN-A`) or drop the inline `\|` regex from the cell — the table cannot carry a bare `|`. Keep the M1 exit's count-first form (plan.md:32) intact.
2. spec.md §F F1: replace "at most once per hour per project" with the interval-conditional phrasing and add the stamp-write-failure repeat-spawn residual (D2).
3. spec.md REQ-DP-006 / plan.md file 5: align the flag wording per D3.
4. acceptance.md AC-DP-004 or §D: state the construction-only basis of the detachment claim (D4).
5. On the confirming re-audit (reread hunk: acceptance.md#AC-DP-001-green-cell): flip progress.md §E.1 to `plan_status: audit-ready` and pin the new hashes (Q3).
