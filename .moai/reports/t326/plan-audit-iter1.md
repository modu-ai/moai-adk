# SPEC Review Report: SPEC-BINARY-LAG-VISIBILITY-001
Iteration: 1/2 (Tier M ceiling)
Verdict: PASS-WITH-DEBT
Overall Score: 0.80 (Tier M threshold 0.80)

Audit tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t326`, branch `WT-integration-lock-identity`, HEAD `25f7b0fe9` (verified `git rev-parse`).
Reasoning context ignored per M1 Context Isolation — the dispatch's history narrative was used only to choose where to press; every judgment below rests on a measurement made in this run, against this tree.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `grep -o 'REQ-BLV-00[0-9]' spec.md | sort -u` → exactly REQ-BLV-001..008, no gap, no duplicate, uniform zero-padding.
- [PASS] MP-2 GEARS format compliance (requirement layer only — `REQ-BLV-*` in spec.md §3): 001 ubiquitous; 002 event-driven (`When a session starts …`); 003 `Where` capability-gate; 004 compound ubiquitous + event-driven; 005 ubiquitous; 006 unwanted + event-driven; 007 unwanted (`shall not rest on`); 008 ubiquitous + unwanted. The Given-When-Then entries in `acceptance.md` are verification-layer `AC-BLV-*` and are graded under Group 4, not here.
- [PASS] MP-3 YAML frontmatter validity: all 12 canonical fields present in spec.md:1-15, no rejected snake_case alias. Independently confirmed with the domain tool: `moai spec lint --strict .moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/spec.md` → `✓ No findings — all SPEC documents are valid`.
- [N/A] MP-4 language neutrality: single-language SPEC (Go source + Makefile in this repository). No template-bound or 16-language content.
- [PASS] MP-5 D7 cross-SPEC reconciliation: only one external reference, `SPEC-AGENT-EMIT-LINEAGE-001` (t317). Its status is active, not retired/superseded/archived — no BLOCKING finding. It is absent from this tree's `.moai/specs/` (it lives in t317's worktree); the SPEC discloses that explicitly with the path and the snapshot date (spec.md §5 footnote), so the D7-5 "referenced SPEC not found" signal is disclosed rather than hidden. Recorded as D9 below, informational.
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall spec.md` → 0. Auto-PASS.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-BINARY-LAG-VISIBILITY-001/` returns exactly one line — progress.md:29, which *records that the count is zero*, not an open marker. No `research.md` exists (Tier M). No open marker.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Requirements are individually unambiguous, but spec.md:174 contradicts spec.md §1.7.1 / REQ-BLV-008 / plan §G on the single most consequential design point — where the advisory is written (D1). |
| Completeness | 0.75 | 0.75 | All sections present, frontmatter valid, Out of Scope carries 8 H3 sub-headings with specific bullets, §7.5 names the accepted cost. Gap: the scope-boundary constraints C-1/C-2/C-3 (spec.md §6) carry no requirement and no acceptance criterion (D2). |
| Testability | 0.85 | 0.75-1.0 | Anti-vacuity is designed in: each AC names a representative mutant, RED-now cells are pinned to a measured tree SHA, AC-BLV-008 forces judgment on `json.Marshal` output. Two weaknesses: D6, D7. |
| Traceability | 0.85 | 0.75-1.0 | All 8 REQ ids and all 8 AC ids present; every REQ has ≥1 AC; every AC heading carries its REQ id. Constraints sit outside the table (D2); progress.md calls the mapping "1:1" where REQ-BLV-002 maps to two ACs (cosmetic). |

Aggregate: (0.75 + 0.75 + 0.85 + 0.85) / 4 = 0.80.

## Press-point findings (the six the dispatch named)

### 1. Anti-vacuity — PASS, with one residual gap

**AC-BLV-008 does force serialization judgment.** The criterion as written states: "핸들러 출력을 `json.Marshal` 한 **뒤** 그 바이트에서 권고 문자열을 찾는다. 구조체 필드를 직접 읽지 않는다". Measured against the code, the named path is real and the named mutant is genuinely excluded:

- `HookSpecificOutput.AdditionalContext` carries `json:"additionalContext,omitempty"` (`internal/hook/types.go:333`); `HookOutput.HookSpecificOutput` carries `json:"hookSpecificOutput,omitempty"` (`:375`). The JSON path the AC names exists.
- `HookOutput.Data` carries `json:"-"` (`internal/hook/types.go:394`), directly under the comment `// Internal data (not serialized to JSON)`. The mutant "put a perfectly-formed advisory in `Data`" is therefore invisible in the marshaled bytes and is caught.
- Traced the actual advisory flow to confirm the SPEC's premise rather than inheriting it: `computeDeferredAdvisory` (`session_start.go:593`) returns a `map[string]any`; results are merged into `data` at `:264` (sync path) and via `resultCh` at `:574` (async path); `data` is marshaled at `:276` into `jsonData`; `jsonData` becomes `HookOutput{Data: jsonData}` at `:301`. The code's own comment at `:197-199` states the same conclusion. **The SPEC's premise holds as measured.**

**AC-BLV-004 is RED against the current tree, for the stated reason.** Measured here:
- `grep -rn 'BUILD_ID' Makefile pkg/version/` → exit 1, zero matches. No `BUILD_ID` exists — conditions (a) and (b) both fail pre-implementation.
- `git describe --tags --abbrev=0` → `v3.1.2`; `git describe --tags` → `v3.1.2-502-g25f7b0fe9`. The collapse is real: `Makefile:6` `VERSION ?= $(shell git describe --tags --abbrev=0 …)`.
- The SPEC pins its own citation to tree SHA `343399d2f` (494 commits). That is correct discipline, not staleness — `git merge-base --is-ancestor 343399d2f 25f7b0fe9` → true, so the pinned measurement is an ancestor of this tree and the difference (494 → 502) is expected commit accrual. Not a defect.

### 2. Representative mutants — both named claims verified plausible, not strawmen

**AC-BLV-007 mutant 1 (semver comparison) does get the measured inversion backwards.** `git merge-base --is-ancestor 22df80e90 343399d2f` → true. So the build reporting `v3.1.2` is the *descendant* of the build reporting `v3.1.3-rc.5`. A semver comparator classifies the newer binary as stale — the exact inversion the SPEC documents. The mechanism is confirmed at `Makefile:6` (`--abbrev=0` strips the distance suffix) and `pkg/version/version.go:8` (compile default `Version = "v3.1.3"`). Genuine mutant.

**AC-BLV-003 mutant (Fail on "no git") does flip `moai doctor` to exit 1 for distributed users.** `doctorExitStatus` (`internal/cli/doctor.go:140-146`) returns `exitCodeError{code: 1}` whenever `failCount != 0`. The mutant is plausible because §5's applicability predicate is *new* code in a *new* file (`binary_lag.go` per spec.md §4 row 2) — returning `Fail` for "no repository to compare against" is an easy default in fresh code. Genuine mutant.

Also confirmed the existing tolerance the SPEC promises to preserve: `checkBinaryFreshness` (`internal/cli/doctor.go:495-540`) has exactly the five `CheckOK` branches the SPEC enumerates — no commit metadata, `os.Getwd()` failure, `git rev-parse HEAD` failure, prefix match with HEAD, and non-ancestor — with `CheckWarn` on the single ancestor branch. Registration verified at `internal/cli/doctor.go:199` (`{"Binary Freshness", checkBinaryFreshness}`).

**Automatic invocation is genuinely zero.** `grep -rn "moai doctor\|runDoctor\|RunDoctor" .claude/hooks/ .claude/settings.json internal/hook/` returns no matches. The SPEC's central claim — the verdict exists but never reaches anyone — is confirmed independently.

### 3. Distributed-user blast radius — specified and judged, with one scoping defect in the DoD

REQ-BLV-003 states the not-applicable path as: report not-applicable, emit no advisory, leave the `moai doctor` exit status unchanged. AC-BLV-003 judges all three conjunctively — (a) status is not `Fail`, (b) no advisory, (c) exit status 0 — using a `t.TempDir()` non-git directory, plus the t317 D9 sub-directory anchor gate as a separate sub-case. plan.md §B item 2 and §G restate it. This is judged, not merely asserted.

Defect: the DoD item "비-git 임시 디렉터리에서 `moai doctor`가 exit 0을 유지한다" is scoped to the *whole* doctor run, which depends on every other check (MoAI Config, Claude Config, Constitution Registry, MCP checks) also not returning `Fail` in a bare directory — outside this SPEC's control. AC-BLV-003 (c) is correctly scoped ("그 판정만 포함한 `moai doctor` 실행"); the DoD item is not. See D8.

### 4. Scope boundary against t317 — the claim is NOT expressed at requirement level

The dispatch asked whether the "rewire, do not add a new check name" claim is expressed at requirement level and actually holds. Measured:

- spec.md §3 contains no requirement covering it. The eight requirements are about emission, applicability, build identity, seam singularity, fail-open, verdict basis, and emission surface — none constrains the doctor check-name registry.
- spec.md §3.1 traceability table has no row for it.
- acceptance.md has no AC for it.
- It exists in exactly two places: prose "제약 C-2" (spec.md §6) and a DoD checkbox in acceptance.md ("신규 doctor 검사 **이름**이 추가되지 않았다").

A DoD checkbox is a human checklist item, not a mechanically judged criterion. The boundary that keeps two live lanes from colliding is therefore unjudged. This is D2.

Separately, C-2's *justification* rests on an unverified premise. Measured in the t317 read-only snapshot (`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t317`, read 2026-08-27, no write performed): `plan.md:68` states "항목명 문자열은 run-phase 확정" — t317 has NOT yet fixed its check-name string. So "t317은 별도 이름의 항목을 추가하므로 이름 충돌이 발생하지 않는다" asserts a fact not yet determined. The practical risk is asymmetric (this SPEC adding no name is under its own control), but the stated reason is an unverified premise. This is D3.

### 5. The accepted cost in §7.5 — stated plainly, not softened; no DoD contradiction

§7.5 names the cost without hedging: "`VERSION`이 그대로이므로 `moai version`의 **제목 줄은 여전히 `v3.1.2`로 읽힌다.** 제목 줄만 보는 독자는 새 빌드를 옛 빌드로 여전히 오독할 수 있다 — 오늘 실제로 발생한 바로 그 오독이다." Scope-widening is forbidden in three places (§7 Out of Scope "moai version 제목 줄 축", §7.5 closing line, plan.md M3). The DoD **reinforces** rather than contradicts it: "`moai version`의 제목 줄이 `v3.1.2`로 남는 것은 기대된 결과이며 결함이 아니다". AC-BLV-004 (c) mechanically pins `VERSION ?=` byte-identity, and AC-BLV-004 mutant 1 is precisely the scope-widening implementation. No defect found on this axis.

The consumption paths cited as the reason for holding `VERSION` fixed are verified: `Makefile:14` `RELEASE_BINARY := moai-$(VERSION)-$(PLATFORM)`, `Makefile:35` writes `version.json` with `"version":"$(VERSION)"`.

### 6. Requirement-to-AC traceability — mapping verified, not just counted

Every REQ id and every AC id enumerated by grep; each of REQ-BLV-001..008 appears in the §3.1 table with at least one AC; each AC-BLV-001..008 heading carries its own REQ id(s). No requirement is left unjudged and no AC references a non-existent REQ. The mapping is 8→8 but not literally 1:1 (REQ-BLV-002 → AC-BLV-001 + AC-BLV-002); progress.md:26 describes it as a "1:1 추적표" — cosmetic imprecision, not a coverage gap. The substantive traceability gap is the constraints sitting outside the table (D2).

## Defects Found

D1. spec.md:174 — the §4 impact table, row 3, instructs `internal/hook/session_start.go` = "`computeDeferredAdvisory`에 권고 1건 추가" — Severity: major — Class: blocking — This prescribes exactly the anti-pattern the SPEC exists to close. Traced and verified: `computeDeferredAdvisory`'s returned map is merged into `data` (`session_start.go:264`, `:574`), marshaled at `:276`, and assigned to `HookOutput{Data: jsonData}` at `:301`, where `Data` carries `json:"-"` (`internal/hook/types.go:394`). Taken literally, row 3 produces AC-BLV-008's representative mutant and plan.md §G's first anti-pattern. Three other places say the opposite (§1.7.1, REQ-BLV-008, plan M2), so this is residual v0.1.0 text the v0.2.0 correction missed — but §4 is the file-level instruction surface a run-phase implementer reads to learn what to change. Required fix: rewrite row 3 to name the `AdditionalContext` append site (`session_start.go:343-346` / `:369` pattern) as the change target, and state that if the deferred-scan scheduling seam is reused, only the *scheduling* is borrowed and the result must not be written into the advisory map.

D2. spec.md §6 (constraint C-2) + acceptance.md (DoD) — the "no new doctor check name; rewire the existing `Binary Freshness` item" boundary carries no REQ and no AC — Severity: major — Class: blocking — Verified absent from spec.md §3, from the §3.1 traceability table, and from acceptance.md's eight criteria. It survives only as prose and a human DoD checkbox, so the cross-lane boundary against t317 is not mechanically judged. Required fix: add a requirement (e.g. "The lag verdict shall be surfaced through the existing `Binary Freshness` doctor item, and the implementation shall not register an additional doctor check name") with an AC that asserts the `moaiChecks` slice (`internal/cli/doctor.go:196-205`) gains no new entry — a count or a name-set comparison, not a grep for a literal.

D3. spec.md §6 (constraint C-2 rationale) — "t317은 별도 이름의 항목을 추가하므로 이름 충돌이 발생하지 않는다" rests on a fact t317 has not determined — Severity: minor — Class: optional — Measured in the t317 read-only snapshot: `plan.md:68` states "항목명 문자열은 run-phase 확정". The non-collision conclusion is still likely correct, but its stated premise is unverified at audit time. Required fix: restate the rationale as one-sided — this SPEC registers no new name, so a collision can only arise from t317's own choice — rather than asserting what t317 will do.

D4. spec.md §8 — cites `pkg/version/version.go:9` for the compile default `Version = "v3.1.3"` — Severity: minor — Class: optional — Measured: `Version = "v3.1.3"` is at `pkg/version/version.go:8`; `:9` is `Commit = "none"`. Off-by-one citation. Required fix: change `:9` to `:8`.

D5. spec.md §6 (constraint C-1) — "t317의 §Out of Scope가 `agentemit` 밖 임베드 검증을 명시적으로 제외해 이 자리를 비워 두었다" overstates the cited text — Severity: minor — Class: optional — Measured at t317 `spec.md:136-137`: the section declines to *investigate whether the same tautology exists in other embed tests* ("같은 동어반복이 `internal/template` 의 다른 임베드 테스트에도 있는지는 조사하지 않는다"). That is a narrower statement than "excludes embed verification outside agentemit". The practical non-overlap still holds because t317 scopes to `.codex/agents/moai/*.toml`, but the cited reason is not what the cited text says. Required fix: cite t317's positive scope (REQ-AEL-004's `.codex/agents/moai/*.toml` subject) as the non-overlap basis instead of its Out of Scope section.

D6. acceptance.md AC-BLV-006 — the representative mutant may not be RED under the criterion's own judgment method — Severity: minor — Class: blocking — The mutant is "시간 제한 없는 `exec.Command` 호출", but the 판정 injects a blocking stub *at the comparison seam* — i.e. it replaces the very code that would contain that `exec.Command`. If the time box lives caller-side (which is what the stated 판정 actually tests), the mutant is bounded anyway and is not a defect; if the time box is expected to live inside the seam implementation, the stub replaces it and the mutant is never exercised. The AC does not say which side owns the bound. Required fix: state the bound's owner explicitly (handler-side bounded join vs. seam-internal `context.WithTimeout`), and name a mutant the stated 판정 actually turns RED — e.g. a handler that joins the seam without a deadline.

D7. acceptance.md AC-BLV-008 — the judgment does not pin the advisory to the `hookSpecificOutput.additionalContext` key — Severity: minor — Class: optional — As written, the criterion searches the marshaled bytes for the advisory substring. The named mutant (`HookOutput.Data`) is still caught, because `Data` is `json:"-"`. But `SystemMessage` carries `json:"systemMessage,omitempty"` (`internal/hook/types.go:366`) and IS serialized, so an implementation writing the advisory there would pass a whole-document substring search — while violating lead decision (a), which plan.md M1 records as "`additionalContext` 단독. `systemMessage`도, 둘 다도 아니다". Required fix: unmarshal the bytes and assert the advisory appears under `hookSpecificOutput.additionalContext` specifically, and add the `systemMessage` placement as a second representative mutant.

D8. acceptance.md DoD, item "비-git 임시 디렉터리에서 `moai doctor`가 exit 0을 유지한다" — scoped to the whole doctor run rather than to this SPEC's check — Severity: minor — Class: optional — The whole-run exit status in a bare directory depends on every other registered check (MoAI Config, Claude Config, Constitution Registry, the two MCP checks) also not returning `Fail` there, which this SPEC neither controls nor measured. AC-BLV-003 (c) is correctly scoped ("그 판정만 포함한 `moai doctor` 실행"); the DoD item silently widens it. If another check Fails in a bare directory, the item is unachievable for a reason this SPEC did not cause. Required fix: either scope the DoD item to `moai doctor --check "Binary Freshness"` in a non-git directory (the `--check` flag is verified to exist at `internal/cli/doctor.go:58`), or record the whole-run baseline as a measured precondition in plan.md §C before making it a completion gate.

D9. spec.md §5 / §H — `SPEC-AGENT-EMIT-LINEAGE-001` is referenced but is not present in this tree's `.moai/specs/` — Severity: minor — Class: optional — D7-5 "referenced SPEC not found" signal. Mitigated: the SPEC discloses the exact foreign-worktree path and the snapshot date, and declares the tree read-only. Recorded for completeness; no fix required beyond keeping the disclosure.

## Gaps — what this audit did NOT observe

- **Real-binary non-git probe.** I could not run `moai doctor` in a directory outside this worktree. The worktree-isolation guard refused the compound form twice ("this command is too complex to verify that it stays inside the worktree"), and no non-compound form reaches a directory outside the tree. So the DoD item's real-world premise — that a bare non-git directory yields `moai doctor` exit 0 today — is **unverified by me**. It is the measurement D8 asks for.
- **The historical binary A/B (spec.md §1.3).** I verified the ancestry arithmetic (`22df80e90` → `343399d2f` → `25f7b0fe9`) but did NOT re-run the two binaries or re-observe the `held by a session that is gone (reclaimable)` versus `held` outputs. Those rows are inherited from the authoring session. They are §1 background and no requirement rests on them, but they are not my baseline.
- **The installed binary's own build commit.** I did not run `moai version` to confirm which commit the binary on PATH was built from, so the claim that today's installed binary is now fresh is unobserved here.
- **`git describe` behavior under a tagless clone.** `BUILD_ID`'s derivation from `git describe --tags` is specified but its fallback in a tag-less or shallow clone is not addressed by any AC. Not raised as a defect — it may belong to run-phase — but it was not observed either way.
- **t317's eventual check name.** Read-only snapshot only; the name is run-phase-determined there (D3).
- **AC-BLV-005's stub-injection feasibility.** I did not verify that a substitutable seam can be threaded through `checkBinaryFreshness`'s current shape without changing the doctor registration signature (`checkFunc` takes only `verbose bool`, `internal/cli/doctor.go:199`). plan.md §B item 1 acknowledges the injection point does not exist today; whether M2's extraction preserves the registration signature is unmeasured.

## Residual risk

- D1 is a one-cell edit, which makes it cheap to fix and easy to skip. If it is left, the run-phase implementer most likely to be misled is the one who reads §4 as the work list and treats §1.7.1 as background — the exact reading order the impact table invites.
- D2 leaves the two-lane boundary resting on a human checkbox. Both lanes are live and both touch `internal/cli/doctor.go`; a checkbox does not survive a run-phase that is confident it is doing the right thing.
- The SPEC's core premise chain (verdict exists → nothing invokes it → the surface that would carry it is `json:"-"`) is verified end to end in this run and is sound. The defects above are all in the instruction and judgment surfaces, not in the analysis.

## Recommendation

PASS-WITH-DEBT at 0.80, exactly at the Tier M threshold. All seven must-pass criteria clear, and the SPEC's central analysis survives independent re-measurement — including the two things most likely to have been inherited rather than verified (the `json:"-"` unreachability and the semver inversion). Clear D1, D2, and D6 before Implementation Kickoff Approval:

1. **D1** — rewrite spec.md:174 row 3 to name the `AdditionalContext` append site as the change target; state that only `computeDeferredAdvisory`'s *scheduling* is precedent.
2. **D2** — promote constraint C-2 to a requirement with a mechanical AC over the `moaiChecks` registration slice.
3. **D6** — state where AC-BLV-006's time box lives and name a mutant its own judgment method turns RED.

D3, D4, D5, D7, D8, D9 are optional-class: surface them to the orchestrator, and let it decide whether they are worth a revision round or a run-phase note. Do not manufacture a FAIL out of their count.
