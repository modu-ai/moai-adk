# SPEC Review Report: SPEC-BINARY-LAG-VISIBILITY-001
Iteration: 2/2 (Tier M ceiling)
Verdict: PASS-WITH-DEBT
Overall Score: 0.85 (Tier M threshold 0.80)

Audit tree: `.claude/worktrees/t326`, branch `WT-integration-lock-identity`, HEAD `5fc676bbe` (verified `git rev-parse --short HEAD`), clean tree.
Reasoning context ignored per M1 Context Isolation — the dispatch's narrative of what changed was used only to choose where to press. Every judgment below rests on a measurement made in this run, against this tree. The iter-1 report was read as a defect list to re-verify, not as established fact; where I agree with it, I say from which of my own measurements.

Score trend: iter-1 0.80 → iter-2 0.85. Monotone improvement; no STOP escalation.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `grep -o 'REQ-BLV-[0-9]*' spec.md | sort -u` → exactly REQ-BLV-001..009. No gap, no duplicate, uniform zero-padding. The two new ids (009 in spec.md:150, AC-BLV-009 at acceptance.md:152) extend the run without breaking it.
- [PASS] MP-2 GEARS compliance (requirement layer only, `REQ-BLV-*` in spec.md §3, lines 134-150): 001 ubiquitous; 002 event-driven; 003 `Where` capability-gate (the capability "a repository HEAD exists" is absent); 004 compound ubiquitous + event-driven + unwanted; 005 ubiquitous; 006 unwanted + event-driven; 007 unwanted + ubiquitous; 008 ubiquitous + unwanted; **009 (new) ubiquitous + unwanted** — "The lag verdict shall be surfaced through …" / "shall not register an additional doctor check name". Both canonical. The Given-When-Then entries in acceptance.md are verification-layer `AC-BLV-*`, graded under Group 4, not here.
- [PASS] MP-3 YAML frontmatter validity: all 12 canonical fields present and non-empty at spec.md:2-13, plus `tier: M`. No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id` all absent). **Judged by direct field inspection, not delegated to the tool** — `moai spec lint --strict` returns `✓ No findings`, but I measured what that proves: `internal/spec/lint.go:748-773` checks the 12 fields for non-empty presence only and validates no enum. So the lint result is evidence of presence, not of enum conformance. See N8 for the one value that does not match the schema's enum casing; it is minor and mechanically inert.
- [N/A] MP-4 language neutrality: single-language SPEC (Go source + Makefile in this repository). No template-bound or 16-programming-language content.
- [PASS] MP-5 D7 cross-SPEC reconciliation: one external reference, `SPEC-AGENT-EMIT-LINEAGE-001` (t317). Read in the read-only snapshot: its status is active (v0.5.0, iter-3 PASS 0.90 per its own HISTORY line 28), not retired/superseded/archived. No BLOCKING finding. Absent from this tree's `.moai/specs/`, but disclosed with path and snapshot date at spec.md:195.
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall spec.md` → 0. Auto-PASS.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/` returns exactly one line — progress.md:29, which records that the residual count is zero. Not an open marker. No `research.md` exists (Tier M).

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75-1.0 | The D1 contradiction is gone and the document now carries no instruction pointing at a non-serialized destination (two independent sweeps, below). Eight of nine requirements are unambiguous. One residual: AC-BLV-006's [HARD] ownership clause (acceptance.md:102) rejects "seam 내부의 `context.WithTimeout`" and then cites a precedent whose mechanism is exactly `context.WithTimeout` — two reasonable engineers resolve this differently, and the two resolutions differ in whether the AC passes (N4). |
| Completeness | 0.90 | 0.75-1.0 | All sections present; frontmatter 12/12; 8 `### Out of Scope` H3s each with specific bullets (spec.md:223-248); §7.5 names the accepted cost without hedging; §8 records open observations as non-requirements. iter-1's D2 gap is closed — C-2 now carries REQ-BLV-009 + AC-BLV-009 + a §3.1 row. Residual: AC-BLV-009 says its base tree is "base SHA로 고정" but names no SHA (N6), where its sibling AC-BLV-004 names one. |
| Testability | 0.85 | 0.75-1.0 | Anti-vacuity is designed in and now verified mechanically: AC-BLV-008's keyed unmarshal genuinely excludes the `SystemMessage` mutant (measured below); AC-BLV-009 is falsifiable and its mutant does turn it RED; AC-BLV-004's RED-now cell reproduces in this tree. Offsetting: AC-BLV-006's judgment method has an unaddressed test-path feasibility problem (N5) and an under-specified stub contract (N4), and AC-BLV-009 is narrower than the requirement it judges (N7). Net unchanged from iter-1: D7 closed, D6 half-closed, two new weaknesses. |
| Traceability | 0.85 | 0.75-1.0 | 9 REQ ids, 9 AC ids, §3.1 table has 9 rows, every REQ has ≥1 AC, every AC heading carries its REQ id, no orphan AC. The "1:1" correction is accurate, not a reframing — I enumerated the mapping and REQ-BLV-002 is the only one-to-two (AC-BLV-001's heading carries both REQ-BLV-001 and REQ-BLV-002; AC-BLV-002 carries REQ-BLV-002). Offsetting: three citations do not resolve to what they claim (N1, N2, N3), and N1 has a mechanical consequence on the newest AC. |

Aggregate: (0.80 + 0.90 + 0.85 + 0.85) / 4 = **0.85**.

## Press-point findings

### 1. REQ-BLV-009 / AC-BLV-009 — audited from scratch

**REQ-BLV-009 (spec.md:150)** is GEARS-conformant (ubiquitous + unwanted) and states a real, checkable constraint.

**AC-BLV-009 is falsifiable — the "GREEN from the start" framing is honest.** The test: does its named mutant genuinely turn it RED? Measured — `moaiChecks` is a slice literal of `{name, fn}` pairs (`internal/cli/doctor.go:195-212`). The named mutant appends `{"Binary Lag", checkBinaryLag}`. That adds an element to the after-set, the before/after set comparison differs, RED. So the criterion can fail, and the thing that makes it fail is precisely the thing the requirement forbids. Not a criterion that can never fail.

Two properties strengthen it beyond the instruction: set identity also catches a **rename** or **removal** of `Binary Freshness` (which would break the DoD's `--check "Binary Freshness"` gate), and `plan.md:102` (M4) makes the mutant plant the *only* validity evidence for this AC and mandates it — the correct handling for an invariant guard that starts GREEN.

Two defects found against it, both new: the cited literal range is wrong in a way that could hide the append site (N1), and the criterion judges one registry where the requirement binds all three (N7).

### 2. The AC-BLV-009 deviation — sound, and the citation holds

The instruction was to assert the slice "gains no new entry"; the author instead compares the before/after name **set**, pinned to a base SHA, citing t317.

**The citation is accurate.** Read in the read-only snapshot at `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t317` (snapshot taken 2026-08-27T14:53:48Z; no write performed). `plan.md:68` states both halves the SPEC relies on, verbatim: "등록 지점은 `internal/cli/doctor.go` 의 `moaiChecks` 슬라이스(선례 `:201`), 항목명 문자열은 run-phase 확정." So t317 does register into the same slice, and its name string is not yet determined.

**The deviation is sound, and it is not actually a deviation from iter-1.** iter-1's own required fix offered both options — "a count or a name-set comparison, not a grep for a literal" (iter-1 report line 88). The author chose one of the two named options and gave a measured reason for rejecting the other: an absolute-count assertion breaks the moment t317 lands, for a reason unrelated to this SPEC. That reason is verified above.

**Does it weaken the criterion — could a violating implementation slip past?** For the violation the requirement names (registering an additional check name in `moaiChecks`), no: any added name changes the set. The set form is strictly stronger than "no new entry" because it also catches rename and removal. The one hole is scope, not form — see N7.

One residual risk, not a defect: `moaiChecks` entry `:201` uses the constant `mcpServerVersionCheckName` rather than a string literal, so a naive literal-extraction misses one name. Harmless for a before/after delta (missing on both sides), but the extraction must be name-set-of-the-slice, not literals-in-the-range.

### 3. The D1 sweep — document-scoped, and it did finish

Two independent sweeps across all four artifacts:

1. `grep -rn -e computeDeferredAdvisory -e 'HookOutput.Data' -e '권고 맵' -e 'Data 맵' -e 'json:"-"'` — 20 hits.
2. `grep -rn -e jsonData -e resultCh -e systemMessage -e SystemMessage -e Data` — 9 hits after excluding the first sweep's prohibitions.

**Every surviving mention is a prohibition, a mutant description, a HISTORY entry, or a scheduling-only precedent. None is an instruction to write there.** Checked each instruction-shaped surface individually:

- spec.md:176 (§4 row 3, the cell iter-1 flagged): now reads "**`AdditionalContext` append 지점**(`:343-346`·`:369` 패턴)에 지연 권고를 덧붙인다" — the correct target. Verified those lines exist and are the append-if-non-empty pattern (`session_start.go:343-346`, `:369`).
- spec.md:181 (new binding clause): prohibition.
- plan.md:33 (§B item 3, the ninth instance the author found): now carries the prohibition inline.
- plan.md:70, :72, :126, :127, :145: prohibitions and the scheduling-only split.
- acceptance.md:102: cites `computeDeferredAdvisory` on the **timeout** axis, which is the scheduling axis — permitted by the split. (Its accuracy is a separate matter — N4.)
- acceptance.md:144, :146: mutant descriptions.

No third instance survives. From my own measurement, the sweep is document-scoped.

### 4. D2-D8 verified against iter-1's statement of the defect

- **D2 (blocking) — closed.** REQ-BLV-009 at spec.md:150, AC-BLV-009 at acceptance.md:152-166, §3.1 row at spec.md:164. The boundary is now mechanically judged, not a human checkbox. Two new defects against the repair (N1, N7).
- **D3 (optional) — closed.** spec.md:215 restates the rationale one-sided: this SPEC registers no name, so a collision can only arise from t317's own choice. Verified against t317 `plan.md:68` myself.
- **D4 (optional) — closed.** Measured: `pkg/version/version.go:8` is `Version = "v3.1.3"`; `:9` is `Commit = "none"`. spec.md:260 now cites `:8`.
- **D5 (optional) — closed in substance.** spec.md:209-211 replaces the Out-of-Scope basis with t317's positive scope. I read the cited t317 section: the heading is at `spec.md:137` and the bullet at `:138`, reading "같은 동어반복이 `internal/template` 의 다른 임베드 테스트에도 있는지는 조사하지 않는다" — the narrower statement, exactly as the correction says. The line range cited is off by one (N3).
- **D6 (blocking) — partially closed.** The mutant is genuinely repaired: it is now "기한 없이 seam을 조인하는 핸들러", which lives caller-side and is therefore **not** replaced by the injected stub, so the stated judgment does exercise it — the structural flaw iter-1 identified is gone. But the ownership clause the repair added cites a precedent that contradicts it (N4), and the judgment's feasibility in the test binary is unaddressed (N5).
- **D7 (optional) — closed, and I confirmed it forces the keyed unmarshal.** acceptance.md:138 now requires unmarshalling and extracting `hookSpecificOutput.additionalContext` specifically, and explicitly forbids both struct-field reads and whole-document substring search. Measured that this turns the `SystemMessage` mutant RED: `SystemMessage` carries `json:"systemMessage,omitempty"` (`internal/hook/types.go:366`) and IS serialized, so a whole-document search would pass it — but it is not reachable under the `hookSpecificOutput` key. The keyed path exists: `HookOutput.HookSpecificOutput` is `json:"hookSpecificOutput,omitempty"` (`:375`) and `HookSpecificOutput.AdditionalContext` is `json:"additionalContext,omitempty"` (`:333`). Mutant 1 (`HookOutput.Data`, `json:"-"` at `:394`) remains excluded. The repair does what it claims.
- **D8 (optional) — closed and mechanically verified.** The DoD (acceptance.md:179) and plan.md:113 now scope to `moai doctor --check "Binary Freshness"`. Measured that this form works: `internal/cli/doctor.go:232` filters by exact, case-sensitive equality against the registered name, so the flag selects exactly the `:199` entry. I ran it in this tree — rc=0, one warn, no other check executed.
- **D9 (informational) — retained.** Disclosure intact at spec.md:195.

### 5. Re-checking what iter-1 passed

- **Counts and traceability after the 8→9 change.** Enumerated independently; see the Traceability row. The "every REQ has at least one AC" correction at progress.md:25 is accurate — REQ-BLV-002 is genuinely the only one-to-two mapping, and AC-BLV-001's heading carries both REQ ids, which is what makes the claim true rather than convenient.
- **The AC-BLV-004 tree-SHA pin.** I reach iter-1's conclusion from my own measurement, and from a second one it did not make. `git merge-base --is-ancestor 343399d2f 5fc676bbe` → true, so the pinned measurement is an ancestor of this tree and the 494 → 514 drift (`git describe --tags` → `v3.1.2-514-g5fc676bbe`) is expected commit accrual. **Second measurement:** `moai doctor --check "Binary Freshness"` reports `binary is behind source tree (binary: 343399d2f, HEAD: 5fc676bbe)`. So `343399d2f` is the commit the *installed binary* was built from — a live anchor, not a stale artifact. Correct discipline; keep the pin.
- **AC-BLV-004 is still RED against this tree, for the stated reason.** `grep -rn 'BUILD_ID' Makefile pkg/version/` → rc=1, zero matches; `git describe --tags --abbrev=0` → `v3.1.2` against `git describe --tags` → `v3.1.2-514-g5fc676bbe`. Conditions (a) and (b) both fail pre-implementation, and they fail because no monotone identity exists — not because of an unrelated file.
- **AC-BLV-007's inversion premise.** `git merge-base --is-ancestor 22df80e90 343399d2f` → true. The build reporting `v3.1.2` is the descendant of the one reporting `v3.1.3-rc.5`. Mutant 1 (semver comparison) gets the measured case backwards. Genuine.
- **The five-branch tolerance §5 promises to preserve.** Read `checkBinaryFreshness` (`internal/cli/doctor.go:495-541`): exactly five `CheckOK` branches (`:499` no commit metadata, `:506` getwd failure, `:513` not a git tree, `:520` prefix match, `:535` non-ancestor) and one `CheckWarn` (`:528`). No `Fail` branch anywhere — so AC-BLV-003's mutant applies to the *new* seam code, as the SPEC frames it. Registration confirmed at `:199`.
- **The §7.5 accepted-cost statement.** Re-read; unchanged and unhedged, and the DoD (acceptance.md:181) reinforces rather than contradicts it. `Makefile:6`, `:14`, `:35` and `internal/update/local.go:65` all verified as cited — the consumption paths that justify holding `VERSION` fixed are real.
- **Frontmatter.** Verified field-by-field rather than by tool, after measuring what the tool actually checks. See MP-3 and N8.

## Defects Found

N1. acceptance.md:166 (and plan.md:86) — the `moaiChecks` literal is cited as `internal/cli/doctor.go:196-205`; measured, the literal spans **195-212** — Severity: major — Class: blocking — The declaration is at `:195`, entries run `:196-211`, closing brace at `:212`. The cited window stops mid-entry (the Constitution Registry closure body ends at `:206`) and omits four further entries: Harness 5-Layer `:207`, Migration `:208`, Plugin Deployment `:209`, Home Disk Usage `:211`. This is not merely cosmetic: **the natural place to append a new check is the end of the slice, which is outside the cited window.** AC-BLV-009 labels this range "판정 근거 위치" (the location of the judgment basis), so an implementer who extracts names from lines 196-205 would not see the mutant the criterion exists to catch. The AC's own sentence is correct ("`moaiChecks` 슬라이스가 등록하는 검사 이름 집합"), which is what keeps this from being fatal. Note the range was inherited from iter-1's own required-fix text, so it is not an authoring regression. Required fix: change `196-205` to `195-212` in both places, and state that the extraction scope is the whole slice literal, not a line window.

N2. acceptance.md:102 — AC-BLV-006's [HARD] ownership clause cites a precedent that implements the form the clause rejects, and the real precedent is elsewhere — Severity: major — Class: blocking — The clause reads "seam 내부의 `context.WithTimeout`이 아니라, 핸들러가 seam을 **기한 있는 조인**으로 호출한다. 선례는 `computeDeferredAdvisory`의 `driftTimeout` 처리(`internal/hook/session_start.go:593` 인접)이며 같은 계약을 따른다." Measured: that handling is `session_start.go:622-624` — `driftCtx, cancel := context.WithTimeout(context.Background(), driftTimeout)` wrapping the injected `driftFn`, a synchronous ctx-bounded **call**, not a join. The actual caller-side bounded join in this file is `session_start.go:243-257`: `joinTimer := time.NewTimer(deferredScanJoinBound)` with a `select` over `advisoryCh`. The difference is mechanical, not stylistic: a ctx-wrap does not bound the handler when the injected stub ignores cancellation, whereas a timer+select join does — and AC-BLV-006's judgment injects "제한을 넘겨 블록하는 스텁" without specifying whether the stub honors cancellation. Under a ctx-ignoring stub, an implementation that follows the cited precedent fails the criterion; under a ctx-honoring stub, both forms pass. So the clause, the precedent, and the judgment do not jointly determine one compliant implementation. Required fix: cite `session_start.go:243-257` as the bounded-join precedent, and state in the 판정 whether the blocking stub honors context cancellation.

N3. acceptance.md:110 — AC-BLV-006's judgment method is not feasible on the default test path, and no artifact says so — Severity: minor — Class: optional — Measured: `internal/hook/main_test.go:47` sets `deferredScansAsync = false` for the entire `internal/hook` test binary, so `deferredScansAsyncEnabled()` (`session_start.go:1436`) is false and the handler takes the **inline** branch (`session_start.go:258-274`), which has no bounded join at all. A test written per AC-BLV-006 would therefore hang under a blocking stub regardless of the implementation's correctness. The mechanism is reachable — `session_start_parallel_test.go:315-321` flips the flag back to true with restore — so this is a run-phase note, not an impossibility. Required fix: name the flag flip (with restore) in AC-BLV-006's 판정, citing the existing precedent test.

N4. spec.md:181 and plan.md:33 — the marshal is cited at `session_start.go:276`; measured, `json.Marshal(data)` is at `:277` (`:276` is a blank line) — Severity: minor — Class: optional — The surrounding chain is otherwise correct: the advisory merge is `maps.Copy(data, advisory)` at `:266` (the call cited as `:264` is two lines above the merge), the async path is `:574`, and `out := &HookOutput{Data: jsonData}` is at `:301` exactly as cited. Required fix: `:276` → `:277`; optionally cite `:266` as the merge site rather than `:264`.

N5. spec.md:211 — the t317 Out-of-Scope section is cited as `spec.md:136-137`; measured in the read-only snapshot, the heading is at `:137` and the bullet at `:138` — Severity: minor — Class: optional — The quoted text is accurate; only the range is off by one. `:136` is a blank line. Required fix: `136-137` → `137-138`.

N6. acceptance.md:154 — AC-BLV-009 says its before-tree is pinned ("base SHA로 고정") but names no SHA — Severity: minor — Class: optional — Its sibling AC-BLV-004 names `343399d2f` explicitly (acceptance.md:71), so the inconsistency is visible within one file. The practical risk is low because a before/after **delta** is robust to the base choice as long as the base is this SPEC's own pre-change tree, but an unfilled pin invites a run-phase implementer to pick a moving reference. Required fix: name the base SHA, or state explicitly that the base is "this SPEC's first commit's parent" and that no fixed SHA is needed.

N7. spec.md:150 vs acceptance.md:155 — REQ-BLV-009 binds every doctor check registry; AC-BLV-009 judges only `moaiChecks` — Severity: minor — Class: optional — The requirement says "shall not register an additional doctor check name", unqualified. Measured, `internal/cli/doctor.go` has three sibling registries that all feed the same verdict list: `systemChecks` (`:187`), `moaiChecks` (`:195`), `workspaceChecks` (`:214`), merged at `:245-249` into `allChecks` (`:93-95`) and thence into `doctorExitStatus`. An implementation registering `{"Binary Lag", …}` into `workspaceChecks` violates REQ-BLV-009 while passing AC-BLV-009. `moaiChecks` is the likely site (it holds the existing `Binary Freshness`), so this is an under-scoping, not an open door. Required fix: extend the extraction to all three registries, or narrow REQ-BLV-009's wording to the `moaiChecks` registry.

N8. spec.md:9 (mirrored in acceptance.md:9, plan.md:9, progress.md:9) — `priority: HIGH` is not the schema's enum casing, and spec.md:4's `version: 0.3.0` is unquoted where the three sibling artifacts quote it — Severity: minor — Class: optional — The frontmatter schema's `priority` enum is `P0|P1|P2|P3` or `High|Medium|Low|Critical`; `HIGH` is a case variant. Measured that nothing enforces it: `internal/spec/lint.go:748-773` validates presence only, and no non-test Go consumer string-matches a priority value (`grep -rn 'fm.Priority\|\.Priority ==' internal/` → no consumer). So `moai spec lint` passing is not evidence of enum conformance, and the value is mechanically inert. Recorded so the SPEC's own reliance on the lint result (progress.md:26) is not read as broader than it is. Required fix: `HIGH` → `High`, and quote `version` in spec.md for parity with its three siblings.

## Regression Check (iter-1 defects)

| iter-1 defect | Class | Status | Basis |
|---|---|---|---|
| D1 §4 row 3 anti-pattern instruction | blocking | **RESOLVED** | spec.md:176 rewritten to the `AdditionalContext` append site; binding clause spec.md:181 added; two document-wide sweeps found no surviving instruction to a non-serialized destination |
| D2 C-2 unjudged | blocking | **RESOLVED** | REQ-BLV-009 spec.md:150 + AC-BLV-009 acceptance.md:152 + §3.1 row spec.md:164. New defects against the repair: N1, N7 |
| D3 C-2 rationale premise | optional | **RESOLVED** | spec.md:215 one-sided; t317 `plan.md:68` re-read and confirms both cited facts |
| D4 `version.go:9` off-by-one | optional | **RESOLVED** | measured `pkg/version/version.go:8` = `Version = "v3.1.3"` |
| D5 C-1 rationale overstated | optional | **RESOLVED** (line range off — N5) | t317 `spec.md:137-138` re-read; quoted text accurate |
| D6 AC-BLV-006 mutant not RED-able | blocking | **PARTIALLY RESOLVED** | mutant now caller-side and genuinely exercised by the stated judgment; but N2 (contradictory precedent) and N3 (test-path feasibility) remain |
| D7 AC-BLV-008 key not pinned | optional | **RESOLVED** | acceptance.md:138 keyed unmarshal + acceptance.md:146 `systemMessage` mutant; `types.go:333/366/375/394` all re-measured, mutant confirmed RED-able |
| D8 DoD scope too wide | optional | **RESOLVED** | acceptance.md:179 + plan.md:113 narrowed; `--check` exact-match verified at `doctor.go:232`; ran the command, rc=0 |
| D9 t317 SPEC not in this tree | optional | **RETAINED** (informational) | disclosure intact at spec.md:195 |

All nine iter-1 defects were examined. None was left unexamined.

## Gaps — what this audit did NOT observe

- **Real non-git probe.** I could not run `moai doctor` in a directory outside this worktree; the isolation guard refuses the compound form and no non-compound form reaches outside the tree. The DoD's real-world premise — that a bare non-git directory yields `--check "Binary Freshness"` exit 0 today — is **unverified by me**, exactly as in iter-1. plan.md M4 still owns this measurement.
- **AC-BLV-005 stub feasibility — inspected, not demonstrated.** I read the registration type (`checkFunc{name string; fn func(bool) DiagnosticCheck}`, `doctor.go:182-186`) and observed that entries `:200-209` already use closures capturing `cwd`, so a closure seam is available without changing the registration signature or the name set — which means AC-BLV-005 and AC-BLV-009 are not in tension. I did **not** write or compile such a seam, so this is an inspection, not a demonstration.
- **The §1.3 historical binary A/B table.** I verified the ancestry arithmetic and observed the installed binary's commit (`343399d2f`, via the doctor check), but did **not** re-run the two binaries or re-observe the `reclaimable` versus `held` outputs. Those rows remain inherited from the authoring session. No requirement rests on them.
- **`git describe --tags` fallback in a tagless or shallow clone.** Still unaddressed by any AC and unobserved by me either way — `BUILD_ID`'s derivation would degrade there. Raised as a run-phase consideration, not a defect.
- **t317's live-tree stability.** I read its snapshot at 2026-08-27T14:53:48Z UTC; its SPEC directory mtime was roughly three minutes earlier. It is another lane's live worktree and may have changed after my read. No write was performed there.
- **Go test execution.** I did not run `go test ./internal/cli/... ./internal/hook/...` — plan §C's baseline is a run-phase precondition, not an audit obligation, and the local-full-suite prohibition applies.
- **`hookSpecificOutput` non-nil on the lag path.** AC-BLV-008's keyed extraction assumes `hookSpecificOutput` is present in the marshaled output. `session_start.go:302` gates its construction on `input.SessionID != "" && input.ProjectDir != ""`; whether the lag advisory path guarantees that object exists is unmeasured by me and unaddressed by any AC.

## Residual risk

- N1 and N2 are both single-cell edits, which makes them cheap and easy to skip. N2 is the more consequential: it is the one place where the document does not determine a single compliant implementation, and the requirement it governs (fail-open at session start) is the one whose failure mode is a hung session start.
- The SPEC's core premise chain — the verdict exists, nothing invokes it, and the field that would carry it is `json:"-"` — is verified end to end again in this run, and this time with a live observation: the doctor check right now prints `binary is behind source tree (binary: 343399d2f, HEAD: 5fc676bbe)` at rc=0 and no one asked for it. That is the SPEC's thesis, reproduced. The defects above are all in the instruction and judgment surfaces, not in the analysis.
- Because this is the Tier M iteration ceiling, N1 and N2 will not be re-audited by this stream. If they are cleared, they are cleared without an independent verdict on the repair.

## Recommendation

**PASS-WITH-DEBT at 0.85**, clearing the Tier M threshold of 0.80 with margin, all seven must-pass criteria cleared, and a monotone improvement from iter-1's 0.80. The eight repairs land: five are fully closed, one is closed with a line-number slip, one is closed and independently confirmed to force the behavior it claims, and one is half-closed. The two brand-new artifacts (REQ-BLV-009 / AC-BLV-009) survive the full anti-vacuity standard applied from scratch.

Clear these two before Implementation Kickoff Approval — both are single-cell edits:

1. **N1** — correct the `moaiChecks` range to `195-212` in acceptance.md:166 and plan.md:86, and state that the extraction scope is the whole slice literal. As cited, the append site an implementer would most naturally use falls outside the window.
2. **N2** — repoint AC-BLV-006's precedent to `session_start.go:243-257` (the timer+select bounded join) and state whether the blocking stub honors context cancellation. As written, the clause, its precedent, and its judgment do not jointly pick one compliant implementation.

N3 through N8 are optional-class. Surface them and let the orchestrator decide whether they warrant an edit or a run-phase note. Do not manufacture a FAIL from their count — the score clears, the must-pass firewall clears, and the SPEC's analysis is sound.

This is iteration 2 of the Tier M ceiling of 2. Per the retry-loop contract the orchestrator now chooses among: accept PASS-WITH-DEBT and proceed to Implementation Kickoff Approval (recommended, after the two edits above), reduce scope, or explicitly override the cap for an iter-3.
