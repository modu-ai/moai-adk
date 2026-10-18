auditor-model: claude-sonnet-5-5
verdict: FAIL
audited_sha: a2a7b8bb24b620c1390bc8e8efa0aacee411bdf1
overall_score: 0.84
must_pass_failed: 0
blocking_count: 2
plan_artifact_hash: a1688548678f6e2160dde24081e576950a88a15933b99a5686dbc355eff912c5
required_backend: codex fail
convergence_overall: fail
scope: full
fix_scope: acceptance.md#AC-OWR-21 design.md#D-3 plan.md#M1b plan.md#M8 plan.md#M9 acceptance.md#M-self-check
defect_class: N12=design N13=other

# SPEC Review Report: SPEC-UPDATE-OWNERSHIP-REPAIR-001

Iteration: 3/3 (the plan-audit ceiling for Tier L; ceiling 3, hold-and-split)
Outcome: FAIL
Overall Score: 0.84 (iteration 1: 0.68, iteration 2: 0.76, iteration 3: 0.84 — improving, so no STOP signal)
Plan Artifact Hash: a1688548678f6e2160dde24081e576950a88a15933b99a5686dbc355eff912c5 (supplied by the lane from the runtime ComputeHash at this HEAD; not recomputed)
Auditor Version: plan-auditor/v1 (Claude leg in-session, required codex leg through audit_multi)
Tier: L (pass threshold 0.85). Reasoning context ignored per M1 Context Isolation. Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865, branch WT-update-ownership-repair, HEAD a2a7b8bb24b620c1390bc8e8efa0aacee411bdf1, `git status --short` empty.

## Summary

The iteration-2 repair is substantive. Of the eleven prior defects, ten are closed on my own commands and one (N4) is only partly closed: the acceptance criterion was rewritten but plan.md milestone M8 still carries the old impossible whole-tree restore condition (and M9 carries a second stale line contradicting REQ-OWR-016). One new test-design gap is blocking: AC-OWR-21 claims to catch a non-atomic (direct-write) prior-manifest writer and cannot. No must-pass criterion fails, no requirement or milestone needs profile-root access, a protected-file change, `removalEnabled` activation, a separate security permission, or card t1702 scope.

The overall verdict is FAIL for three independent reasons: (1) the required backend `codex` returned `fail` (an unmet required gate — it is a gate result, not only a set of reviewed defects); (2) two blocking findings (N12, N13) are open; (3) the aggregate 0.84 is below the Tier L threshold 0.85.

## Audit plan and backend result

- `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`: `config_status: ok`, model codex, codex gate `required` (source config.gates), claude `off`, glm `off`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- `mcp__moai__audit_multi` was called once (project_root = the tree, target baseBranch, focus naming N1-N5 closure, AC-OWR-08 enforceability, provenance mutants, REQ-OWR-016/-017, AC-OWR-20/-21, Q2 (가) conditions; no `gates` argument). Result: `overall_verdict: fail`, `per_backend_verdicts[codex] = {gate: required, verdict: fail}`, `residual_risk_note: "required-backend FAIL: codex"`, `plan_source: config`, `audit_receipt: rcpt-48fbda4048c92f444d7ba40d`, `build_commit: ee981a932`. The installed binary (ee981a932) is a strict ancestor of HEAD a2a7b8bb2 (server lag notice); the backend text describes that build, but every finding concerns SPEC text and is checked below against the tree.
- The required backend is `codex` and it did not pass, so the overall verdict cannot be PASS or PASS-WITH-DEBT. This is an unmet gate, named here: codex.

### Codex findings, verbatim, each checked by this auditor

Quoted from the backend result (Korean, unedited; the backend's own "Verdict: fail" header line is omitted from the quotation only to keep the machine lines of this file unique):

> [P1] 프로필 읽기 결과를 버리는 변이가 AC-OWR-08의 검사망을 통과할 수 있다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/acceptance.md:125
> 확신: 높음. `strings.Join([]string{"claude","profiles"}, "-")`로 경로를 만들고 CLI에서 직접 `os.ReadDir`하면 검사 대상 심볼·토큰·환경 변수 호출은 늘지 않는다. HEAD a2a7b8bb2의 overlay 탐침에서 `successful-profile-directory-reads=3 normalized-output-equality=true`를 관측했다.
> 권고: 이 변이를 필수 대조군에 추가하고 실제 접근을 기록하는 검증 장치를 마련하라. 수용 부채라는 기재만으로 REQ-OWR-005의 접근 금지 조건이 충족되지는 않는다.

Auditor check: CONFIRMED as fact, and the SPEC states it itself (acceptance.md AC-OWR-08 limits (1) and (3); design.md D-8 L1/L3 "못 잡는 것"; decision-index.md accepted debt). A path assembled at run time from fragments that name no listed symbol or token passes L1, L3 and L4. I did not re-run the overlay (codex's measurement). Classified N16, Class optional: a disclosed, leader-owned residual with no plan-level mechanical fix short of the access-recording seam the SPEC deliberately rejected for merge-conflict cost. It stays an obligation of REQ-OWR-005 that review (not AC-OWR-08) carries; the leader must accept it by name before Kickoff.

> [P1] 서로 충돌하는 이전 세대에서 오래된 양성 기록을 선택하는 승격을 막지 않는다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/design.md:60
> 확신: 높음. 최신 세대가 P를 `UserCreated`로 기록하고 이전 세대가 같은 바이트의 P를 `TemplateManaged`로 기록한 경우의 우선순위가 없다. 최초로 조건에 맞는 양성 기록을 찾는 구현은 AC-OWR-05의 단일 세대 출처값 시험들을 만족하면서 P를 승격할 수 있다.
> 권고: 최신 경로 기록의 권위와 세대 충돌 시 보존 규칙을 명시하고, `UserCreated`·`UserModified`·배포 버전 누락 기록 뒤에서 오래된 양성 기록을 찾는 변이를 추가하라.

Auditor check: CONFIRMED as a specification gap (N14). REQ-OWR-002 (spec.md:62) says "a prior manifest record of that path whose recorded provenance is the template-managed value"; design.md D-2 (L52-53, L60) reads "최근 것부터 최대 8개"; AC-OWR-06 (c) reads "최신 세대부터 최대 8개". Nothing says which generation governs when generations disagree. Class optional: the requirement's literal text (an existential "a prior record") is satisfied by an any-generation reader, and the realistic path to the conflicting state is thin (the deployer records `UserCreated` only for a path with no entry; a start-of-run copy that holds `UserCreated` for P normally means the current manifest holds it too, so P is registered, not unregistered). The harm is bounded (the bytes equal a template file, the user-root counterpart is confirmed, a byte backup exists).

> [P1] `DeployedHash` 대체 일치를 허용하면 수정 파일이 항목 유실 후 managed로 승격된다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/design.md:55
> 확신: 높음. 기존 픽스처를 실행해 `provenance=template_managed disk==DeployedHash:true disk==TemplateHash:false`를 확인했다. 등록 상태에서는 수정 파일인 동일 바이트가 이전 세대 증거로 읽힐 때는 명시된 OR 조건을 만족한다.
> 권고: `TemplateHash != DeployedHash == 현재 해시`인 음성 사례를 추가하고, `DeployedHash`만의 일치로 수정 파일을 승격하지 못하게 하라.

Auditor check: CONFIRMED that design.md:55 says "이전 기록의 `TemplateHash` (또는 `DeployedHash`) 와 같아야 한다" while the registered-entry rule (design.md D-1, `managed` = bytes equal `TemplateHash`) has no such alternative, so the attested path is looser than the registered path. The fixture shape codex used is a hand-built test manifest; I read the production writers instead: `Track` computes `DeployedHash` from the file just written (`internal/manifest/manifest.go:135-155`, deployer `internal/template/deployer.go:294`), so for common-asset roots `DeployedHash == TemplateHash` in deployer-written entries, and `retrackManifestFiles` (`internal/cli/update_manifest_retrack.go:54`) re-tracks only entries the update wrote. The divergence is therefore a design-text looseness with a small production footprint (N15, Class optional).

> [P1] 디렉터리 충돌 주입은 Backup 단계의 실패 인계를 여전히 독립적으로 검증하지 못한다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/acceptance.md:271
> 확신: 높음. 설계 D-3은 제거 전에 사본의 유효성도 재검사한다. 따라서 실패 결과를 무시하는 변이도 사례 (a)의 디렉터리를 발견해 보존하고, (b)는 파일 백업 실패로 보존하며, (c)는 제거한다. 세 사례가 모두 기대 결과를 내면서 실패 인계는 빠질 수 있다.
> 권고: 유효한 사본이 존재하지만 쓰기 함수가 오류를 반환하는 사례를 추가해, 디스크 재검사와 별개로 실패 결과가 제거를 차단하는지 검증하라.

Auditor check: NOT CONFIRMED as a defect. REQ-OWR-016 (spec.md:76) is behavioral: "When the run backup cannot hold the byte copy ... the migration shall remove no file in that run". An implementation that drops the error hand-off but re-validates the copy on disk (design.md D-3 item 4 (ii)) still removes nothing when the copy cannot be held, which is the requirement. The mutant that matters — a Backup stage and migration with neither the hand-off nor the on-disk check — goes red in case (a) (removal list non-empty). The recommended extra case (writer returns an error while a valid copy exists) tests an implementation detail the requirement does not state. No finding raised.

> [P2] 실패 사례에 정상 사본 생성을 요구하여 AC-OWR-20이 모순된다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/acceptance.md:272
> 확신: 높음. 여기서는 규칙 7을 만족하도록 요구하지만, 규칙 7(ii)는 새 정상 `manifest.prior.json`과 원본 바이트 일치를 요구한다. 사례 (a)는 바로 그 파일 생성을 실패시키는 구성이다.
> 권고: 실패 사례의 도달 증거를 writer 호출·마이그레이션 진입으로 분리하고, 정상 사본 생성 조건은 성공 사례에만 적용하라.

Auditor check: CONFIRMED as a wording contradiction (N18). acceptance.md:272 says each case "규칙 7 의 조건을 만족한다 (skipped == false; 흐름이 Backup 단계와 마이그레이션에 도달했음)"; the parenthetical narrows the intent, but the literal "규칙 7" (acceptance.md:25) includes (ii), a new valid `manifest.prior.json` byte-equal to the pre-run manifest, which case (a) is built to prevent. Class optional (the parenthetical resolves it for a careful reader; one sentence fixes it).

> [P2] 디렉터리 mtime을 포함하는 공통 해시는 강제 반복 갱신의 멱등 기준을 실패시킨다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/plan.md:105
> 확신: 높음. 기존 흐름을 강제로 세 번 실행한 overlay 탐침에서 모두 건너뛰지 않았고, 실행 2·3 사이 비교 제외 영역 밖 디렉터리 mtime이 58곳 달랐다. 같은 바이트를 임시 파일과 rename으로 배포해도 발생하는 차이다.
> 권고: mtime 검사는 dry-run 무변경 검증에 유지하고, AC-OWR-17의 내용 멱등 해시에서는 제외하라.

Auditor check: CONFIRMED by code reading (N17). plan.md M0 item 2 (L105) prescribes one shared tree hash "경로 + 바이트 + 모드 + 디렉터리 mtime"; acceptance.md AC-OWR-17 and AC-OWR-18 say only "트리 해시 (제외: …)" without saying mtime is excluded; the deployer rewrites every template-managed file unconditionally through a temp file and rename (`internal/template/deployer.go`, the `atomicWriteFile` call before the `Track` at L294), which changes directory mtimes even when the bytes are identical. The count 58 is codex's. The existing helper `hashDryRunTree` (`internal/cli/update_dry_run_reach_test.go:115`) hashes paths and bytes only, so an mtime-free comparison is available; the AC just does not say which hash it means. Class optional.

> [P2] 원자성 시험의 수동 중단 상태는 최종 이름에 직접 쓰는 변이를 검증하지 않는다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/acceptance.md:286
> 확신: 높음. 직접 쓰기도 부모·목적지 충돌에서는 오류를 반환한다. 중단 사례는 시험이 임시 파일만 직접 만들기 때문에 실제 writer를 거치지 않으며, 정상 완료 뒤 임시 파일이 없다는 조건도 직접 쓰기가 만족한다.
> 권고: 실제 writer의 쓰기와 rename 사이에 실패를 주입하거나 파일 연산을 기록하여, 쓰는 동안 최종 이름이 노출되지 않고 rename을 거쳤음을 검증하라.

Auditor check: CONFIRMED and upgraded to BLOCKING (N12). Detail in the defect list: the three AC-OWR-21 (a) injections are a temp-write failure by directory collision, a rename failure by a directory at the final name, and a hand-built temp-only state. A writer that calls `os.WriteFile(finalName, …)` returns an error under the first two (EISDIR / unwritable destination) and is never exercised by the third (the test builds the state itself); AC-06 (b)'s "임시 파일 흔적이 성공 뒤 남지 않는다" is also satisfied by a direct write. Yet acceptance.md:294 lists "이름 바꾸기 없이 최종 이름에 직접 쓰는 구현 (중단 모사에서 부분 파일이 유효 세대로 읽혀 붉음)" as a mutant the AC catches. It does not.

> [P2] M8에는 N4의 불가능한 전체 트리 복원 조건이 그대로 남아 있다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/plan.md:178
> 확신: 높음. acceptance.md:208–212는 정상 배포·설정 변경 때문에 전체 트리 동일 조건을 철회했지만, M8은 여전히 제거 파일과 매니페스트만 복원한 뒤 실행 전 전체 트리 해시와 같도록 요구한다.
> 권고: M8을 복원 영역 R의 바이트 동일성과 통제 실행 T0 대비 차이 검사로 맞춰라.

Auditor check: CONFIRMED and BLOCKING (N13). plan.md:178 reads "…이전 매니페스트 사본이 `<runDir>/manifest.prior.json`, 사본을 원위치에 되돌린 트리의 해시 == 실행 전 트리 해시 (런 백업 디렉터리 제외)", the exact condition AC-OWR-15 (acceptance.md:208) calls unsatisfiable. I also found a second stale line of the same kind in M9 (plan.md:185), see N13.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: judged on spec.md. `grep -n '^- \*\*REQ-OWR-'` lists REQ-OWR-001 to REQ-OWR-017 on spec.md:61-77, sequential, no gap, no duplicate, uniform zero-padding. Traceability verb output: `COLLECTED: 17 REQ definitions (acceptance input: read)`, no `UNCOVERED:` line and no `ORPHAN:` line. 21 `### AC-OWR-` headings at acceptance.md:31-282, AC-OWR-01 to AC-OWR-21.
- [PASS] MP-2 GEARS/EARS format compliance: judged on the requirement layer only (the 17 `REQ-OWR-*` entries, spec.md:61-77); the Given-When-Then criteria in acceptance.md were not graded here. Ubiquitous: 001, 006, 009, 012, 015, 017. Event-driven with `**When**`: 003, 007, 008, 010, 013, 016. State-driven `**While**`: 004. Legacy EARS negative form (counted as legacy equivalent): 002, 005, 011, 014. No informal "should". Wording notes only (REQ-OWR-005 and -015 are long compound sentences; N21).
- [PASS] MP-3 YAML frontmatter validity: spec.md:2-15 carries all 12 canonical fields with correct types (`version: "0.3.0"` quoted; `created`/`updated` ISO; `priority: P1`; `phase: "v3.2.0 target"` is a release label; `lifecycle: spec-anchored`; `tags` a comma string); `tier: L` and `related_specs` are optional. `grep -ln '^status:'` over the seven artifacts lists spec.md only (statelessness of plan, acceptance, design, research honored).
- [N/A] MP-4 Section 22 language neutrality: single-language, repository-internal Go SPEC.
- [PASS] MP-5 D7 cross-SPEC reconciliation: statuses read from disk this run: SPEC-UPDATE-MIGRATION-001 `completed`, SPEC-UPDATE-REPAIR-RC30-001 `completed`, SPEC-UPDATE-REMOVAL-FAMILY-001 `in-progress` (not retired, superseded or archived; its absence from `depends_on` is reconciled at spec.md:161). No `REVIEW:` candidate, no BLOCKING.
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall` = 0 in all seven artifacts. Auto-pass.
- [PASS] MP-7 clarification gate: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` returned no match (rc=1).
- [N/A] MP-8 RED-now cell re-execution. Reason: no acceptance criterion is classified release-blocking. acceptance.md:19-20 states "21개 기준 모두 plan 시점에는 regression-guard" and all 21 criteria end with `등급: regression-guard → 승격 지점 <Mx>` (counted: 21 lines). `grep -n 'release-blocking\|릴리스 게이트'` over the seven artifacts hits only the rule text at acceptance.md:17, :19, :20. The tests the criteria cite do not exist in the tree (`classifyOwnership`, `provenanceSource`, `priorManifestWriter`, `runUpdateFlowForced`, `removalBackupWriter`, `priorManifestGenerations` each grep to 0 files), so a `-run` selector would print `ok [no tests to run]`, which MP-8's verdict rule says reproduces nothing; the SPEC says so itself (acceptance.md:20, :22) and fills no RED cell in advance. verification-completeness.md § 2.1 undecidable disposition applies: not recorded as a pass. The A-tests, B-cell, C-implementation commit order is the form the commit graph can witness (verification-claim-integrity.md § 2.3).
- [PASS] MP-9 cross-artifact ordering consistency: CN-4 verb output `COLLECTED: 11 milestones in plan order (M0 M1 M2 M3 M4 M5 M6 M7 M8 M9 M10), 0 exit bindings, 5 ordering candidates`. The five candidates (acceptance.md:154, :155, :215, :338, :428) are test names, a RED-reason sentence and the baseline-test list; none orders work across milestones, and with no `Exit:` binding no `CONFLICT:` line can be produced. Manual read: commit order A (tests), B (RED cell), C (implementation) is the same in acceptance.md:21 and plan.md:100; M0 (baseline freeze) precedes M2 (scan), M1a precedes M1b and M5; AC-OWR-08 command 3 (positive control) is ordered before reading command 2 inside the AC. No unsatisfiable pair found. The plan.md M8/M9 stale lines (N13) are a consistency defect, not an ordering conflict.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | `removal-eligible` defined once (spec.md:49) and cited by design.md:28; the single-exception boundary is stated (spec.md:51, :65). Remaining: generation precedence unspecified (design.md:52-53, :60; N14); the extra `(또는 DeployedHash)` alternative (design.md:55; N15); REQ-OWR-005 says both "the frozen existing contacts" and "the single existing contact" (spec.md:65; N21); AC-OWR-20 "규칙 7" parenthetical (acceptance.md:272; N18); AC-OWR-17/18 "트리 해시" undefined against the M0 helper (N17); REQ-OWR-015 bundles three obligations in one sentence (spec.md:75). |
| Completeness | 0.90 | 0.75-1.0 band | All sections present: HISTORY (spec.md:20-26), background (A), terms (B), GEARS requirements (C), judgment table (D), trace table (E), constraints (F), AC overview (G), cross references (H), repository facts (I), five `### Out of Scope — …` H3 headings with `-` bullets (spec.md:190-221), decision rows (K); five Tier L artifacts plus decision-index and progress. Frontmatter complete. Deduction: plan.md M8/M9 carry stale obligations (N13); the no-run-directory edge is unspecified (N19). |
| Testability | 0.75 | 0.75 | Strong: anchored `-run '^…$'` selectors, `--- PASS` lines, zero-match rule (acceptance.md:22), reach controls (rule 8), mutant self-check for every criterion (§ M), forced-run reachability (rule 7). Gaps: AC-OWR-21 (a) cannot tell a direct write from temp+rename although it claims to (N12); AC-OWR-08 cannot see a fragmented runtime path (disclosed limit, N16); AC-OWR-17/18 hash ambiguity (N17); AC-OWR-20 rule-7 wording (N18). |
| Traceability | 0.95 | 0.75-1.0 band | Verb: 17 definitions, no UNCOVERED, no ORPHAN. Chain card item to REQ to AC to milestone is complete (spec.md:97-115, plan.md § F/G). Q2 partition arithmetic still checks (plan.md § G.1, decision-index.md Q2). Deduction: plan.md M8/M9 trace to ACs whose text they now contradict (N13). |

Aggregate (mean of 0.75, 0.90, 0.75, 0.95) = 0.8375, reported as 0.84. Below the Tier L PASS threshold 0.85.

## Prior-iteration closure (N1..N11, checked with my own commands, not the disposition table)

| Id | Result | Evidence |
|----|--------|----------|
| N1 profile non-access scan scope | CLOSED for the vectors named in iteration 2; residual disclosed (N16) | `go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./internal/cli` counts 142 module-local packages at this HEAD, and `internal/config`, `internal/profile`, `internal/manifest`, `internal/template`, `internal/core/project`, `internal/userassets` are all in the closure. The exported surface of `internal/profile` (18 functions, `GetProfileDir` at profile.go:193) is covered by the "every exported identifier" rule; AC-OWR-08 now lists mutants M-a (discarded `os.ReadDir` through `profile.GetProfileDir`), M-b (token splice), M-c (second `os.Stat` in the excepted function), M-d, M-e, M-f; the L4 history command lists `internal/profile`, so any card commit under that package fails command 2. Command 2 prints empty, rc 0; positive control (command 3) prints 4 lines, rc 0. A fragmented, run-time-built path escapes (limit (1)); the verb list also omits `filepath.Glob`, `io/ioutil.*`, `os.Readlink`, which L4 covers only inside `internal/profile`. |
| N2 prior-record provenance | CLOSED for a single generation | REQ-OWR-002 (spec.md:62) and § B (spec.md:54) name `manifest.TemplateManaged`; constants verified at `internal/manifest/types.go:17,21,25,29,35`; the deployer records an untracked existing file as `UserCreated` and skips (`deployer.go:265-267`, read this run); AC-OWR-05 mutants (e)(f)(g) and the production-source cases (h1)(h2) use the real constants. Multi-generation precedence and the `DeployedHash` alternative remain (N14, N15). |
| N3 Backup-to-migration hand-off | CLOSED | `priorManifestWriter` and `runUpdateFlowForced` do not exist in the tree (grep: 0 files) and are specified as new: design.md D-3 item 5 and plan.md M1b (file list, § F) for the seam; plan.md M0 item 3a and § F for the helper. They can drive the production path: `runTemplateSyncWithProgress(cmd, userAssetsInstalled) (skipped bool, err error)` at update_template_sync.go:1225 returns the skip flag, and the migration is invoked inside the Backup step at update_template_sync.go:815-817; the existing helper registers a `force` flag and never sets it (migrate_removal_backup_test.go:30-57), so a forced variant is a one-line change. A directory at `manifest.prior.json` makes a file-over-directory rename fail with EISDIR irrespective of uid. |
| N4 whole-tree restore equality | PARTLY CLOSED | AC-OWR-15 (acceptance.md:206-218) is rewritten correctly (restore region R plus the T0 comparison, D subset of E). plan.md:178 (M8) still requires "사본을 원위치에 되돌린 트리의 해시 == 실행 전 트리 해시" — carried as N13 (blocking). |
| N5 version-match skip | CLOSED | Gate re-read at update_template_sync.go:1239-1248 (`packageVersion == projectVersion && !forceUpdate && versionMatchSkipsSync`, returns `true, nil`); a second gate at :249 for the reporter path. OQ-6 (plan.md:235) and REQ-OWR-012/-013 (spec.md:72-73) state clearly that a plain same-version re-run still skips and recovery needs `--force` or a version/template change. Rule 7 (acceptance.md:25) adds reach assertions and a swept-run count; no AC silently depends on the skip not happening — each multi-run AC uses `--force` plus a separate skipped-run case. Note: `--force` also feeds `archiveLegacySkills(projectRoot, out, forceBackup)` (update_template_sync.go:215, :505); fixtures without legacy skills are unaffected, but forced runs are not byte-for-byte the plain production path. |
| N6 AC-OWR-08 (b) volatile fields | CLOSED | Normalization named: temp root to `<ROOT>`, backup directory name to `<RUN>`, timestamps and `deployed_at` to `<TS>`, with the rule asserted first (acceptance.md:127; design.md D-8 L3). |
| N7 Q2 (가) field consequence | CLOSED | decision-index.md:18 states "nothing for the 313 existing `.agents/skills` files"; plan.md:231 repeats it. |
| N8 REQ-OWR-005 carve-out | CLOSED | spec.md:65 names the single tolerated `os.Stat` inside `launchCandidateIsUsable` (profile.go:439, read this run: `info, err := os.Stat(filepath.Join(baseDir, name))`) and states no other function, call, file or package is excepted; mutant M-c (second `os.Stat`, count 2) is red. Wording nit only (N21). |
| N9 stale counts | CLOSED | plan.md:20-21 state 21 criteria and 17 requirements; the 16 test files in § F count to 16 (M0 1, M1a 2, M1b 3, M2 2, M3 to M10 one each). |
| N10 AC-OWR-17 partition | CLOSED | plan.md § G.1 "남음" row (plan.md:204) contains no AC-OWR-17, -13, -15 or -18; decision-index.md:23 lists AC-OWR-17 under Reduced. |
| N11 research version | CLOSED | research.md frontmatter `version: "0.3.0"` (head of file read this run). |

## Constraint review: could any REQ or milestone need a forbidden thing?

- Profile-root access: none required. REQ-OWR-004 records the profile-side state as `not-checked` and keeps profile-mapped copies unconditionally (Q5 conservative reading, decision-index.md:57); M2 step 3 stops on a blocker report if reading a profile root would become necessary (plan.md:134).
- Protected files: `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go` rc 0; `internal/config/userroots.go` does not exist in this tree; no REQ or milestone touches either file.
- `removalEnabled`: default false (`migrate_project_assets.go:38-39`); `grep -rn 'removalEnabled *=' internal --include='*.go'` excluding tests lists only the declaration at :39; REQ-OWR-014 and AC-OWR-19 bind it.
- Separate security permission, real home, mo.ai.kr, `moai update` against a real project, binary replacement: none required; every criterion runs on fixtures with a temporary HOME and `TMPDIR` under /private/tmp.
- Card t1702 scope: none; spec.md:164 and :197 exclude it.

## Mutant construction (task 3): would the criteria go red?

| Mutant implementation | Result | Evidence |
|-----------------------|--------|----------|
| M-1 byte-equality-alone promotion: `classifyOwnership` returns `managed` for an unregistered file whose bytes equal an embedded template file | RED | AC-OWR-01 variant x2 (embedded same-path bytes) must be `unregistered` (acceptance.md:33-35); AC-OWR-02 (A) with the premise step proves the template lookup succeeds, and its (B) negative control injects the promoting source and requires a non-empty problem list; AC-OWR-14 313 copies carry real template bytes (premise asserts each lookup). |
| M-2 profile-root read hidden behind a helper | RED when the helper or call names a listed symbol or token (`profile.GetProfileDir`, `profile.HasClaudeConfig`, `CLAUDE_CONFIG_DIR`, `claude-profiles`, constant-folded splices, non-constant `os.Getenv`, a second `os.Stat` in the excepted function); NOT RED when the path is assembled from run-time fragments naming no listed symbol (codex measured `successful-profile-directory-reads=3`) | acceptance.md AC-OWR-08 limits (1)(3); N16. |
| M-3 removal after a failed backup (Backup stage swallows the copy error and no on-disk recheck) | RED | AC-OWR-20 (a): a directory at `manifest.prior.json` makes the writer fail; the removal list must be empty and three managed copies survive; (b) file-copy failure keeps all three; (c) twin removes all three. |
| M-4 promotion of a user file through `UserCreated`/`UserModified` provenance (single prior generation) | RED | AC-OWR-05 (e)(f)(g) with real constants, and (h1) through the production source after the real deployer wrote `UserCreated`. |
| M-4b the same promotion when an older generation says `template_managed` and a newer one says `user_created` | NOT RED | no AC varies generations against each other (N14). |
| M-5 non-atomic prior-manifest write (direct `WriteFile` to the final name) | NOT RED | AC-OWR-21 (a): both error injections also fail a direct write; the interrupt state is built by the test, bypassing the writer; AC-06 (b) temp-trace check passes (N12, blocking). |
| M-6 `deployed_version` stamped on the deployer's skipped `UserCreated` entry (`deployer.go:265-267`) | NOT RED | AC-OWR-06 (a) names old-JSON and re-track entries, not this one (N20, optional; harmless to ownership because `UserCreated` is not evidence). |

## Defects Found (structured defect-list)

D-ids continue the iteration-2 numbering. Classes are blocking or optional; severity BLOCKING, SHOULD-FIX or MINOR.

N12. acceptance.md:282-297 (AC-OWR-21 (a), mutant line L294) and design.md D-3 item 3 (L77) — AC-OWR-21 cannot verify atomic creation of `manifest.prior.json` — Severity: BLOCKING — Class: blocking — Atomic temp-plus-rename creation is a stated requirement (REQ-OWR-015, spec.md:75: "creating that copy atomically (a temporary file renamed into place)") and leader condition (3) of the Q2 (가) decision. The three injections in (a) are a temp-write failure by directory collision, a rename failure by a directory at the final name, and an interruption state the test builds by hand. A writer that opens the final name directly fails under the first two (the destination is unwritable or a directory) and is never reached by the third; the clean-run check "임시 파일 흔적이 성공 뒤 남지 않는다" (AC-OWR-06 (b)) a direct write also satisfies. The mutant the AC lists as caught ("이름 바꾸기 없이 최종 이름에 직접 쓰는 구현 … 붉음", acceptance.md:294) therefore passes — under verification-completeness.md § 2 (mutant probe) the criterion is too shallow to adopt. Consequence is bounded (a truncated file is not valid JSON, and the reader treats it as no evidence), but the leader-conditioned property is unverifiable as written. Required fix: give the writer an observable seam in the style the SPEC already uses (for example package-level `priorManifestTempWrite`/`priorManifestRename` function variables, recorded and restored by `t.Cleanup`), and assert through the real writer that (i) the bytes are written to a name different from `manifest.prior.json`, (ii) `manifest.prior.json` is absent until the rename is called, (iii) a failure injected between the temp write and the rename leaves no `manifest.prior.json` and no valid generation, and (iv) the rename is called temp-to-final. Reflect the seam in design.md D-3 item 3, plan.md M1b (file list and GREEN step) and acceptance.md § M row 21.

N13. plan.md:178 (M8 step 1) and plan.md:185 (M9 step 1 (a)) — stale obligations contradict the rewritten criteria and REQ-OWR-016 — Severity: BLOCKING — Class: blocking — (i) M8 still requires "사본을 원위치에 되돌린 트리의 해시 == 실행 전 트리 해시 (런 백업 디렉터리 제외)", the condition acceptance.md:208-212 (AC-OWR-15) calls unsatisfiable (the full flow rewrites `llm.yaml`, deploys template files and changes directory mtimes); this is the unresolved remainder of iteration-2 N4, and an unresolved prior-iteration defect is an automatic FAIL. (ii) M9 step 1 (a) reads "둘째 파일의 백업 디렉터리를 쓸 수 없게 만들어 부분 진행을 만든다", but REQ-OWR-016 (spec.md:76) and AC-OWR-18 (a) (acceptance.md:247, :249) say that run removes zero files, so no partial progress can exist. A run-phase implementer following the plan would write an unsatisfiable test and a fixture that cannot occur. Required fix: replace M8 step 1 with AC-OWR-15's restore-region equality (each removed file and the manifest file byte-equal after restore) plus the T0 comparison (D subset of E); replace M9 (a) with "an unwritable backup path for one candidate; the interrupted run removes nothing". Check the same two milestones against plan.md § G.1 so the (나) subtraction map stays consistent.

N14. design.md D-2 (L52-53, L60) and spec.md:62 (REQ-OWR-002) — no precedence among prior generations — Severity: SHOULD-FIX — Class: optional — When generations disagree about a path, the text does not say whether the newest record governs or any `template_managed` record suffices. A first-match reader promotes a path the newest generation records as `user_created`/`user_modified`/without `deployed_version`, and passes every AC-OWR-05 case (each case uses one generation). Realistic frequency is low (see the codex check above) and the removal stays reversible (byte backup, confirmed counterpart). Required fix (cheap): state in design.md D-2 and REQ-OWR-002 that the newest generation containing the path decides and that a newer non-`template_managed` or version-less record preserves the file, and add one two-generation negative case to AC-OWR-05. Carry as debt if the lane prefers not to widen the delta.

N15. design.md:55 (D-2, "또는 `DeployedHash`") — attested `managed` is looser than registered `managed` — Severity: SHOULD-FIX — Class: optional — Registered entries are `managed` only when the bytes equal `TemplateHash` (design.md D-1, L24); the attested path additionally accepts `DeployedHash`, with no stated reason and no AC (AC-OWR-05 (d) mismatch case is defined against `H`, the recorded template hash). In deployer-written entries the two hashes coincide, so the footprint is small, but a record with `DeployedHash != TemplateHash` would let a modified file be removed after an entry loss while an identical registered file is preserved. Required fix: drop the alternative (require equality with `TemplateHash`) and add a `TemplateHash != DeployedHash == current` negative case.

N16. acceptance.md:125-134 (AC-OWR-08 limits), design.md D-8, decision-index.md "Accepted debt" — fragmented run-time path with a discarded read is undetectable — Severity: SHOULD-FIX — Class: optional — Confirmed (codex measurement not re-run). This is the SPEC's disclosed limit (1) and (3); the access-recording seam was rejected for merge-conflict cost (design.md D-8, acceptance.md G-L). It does not satisfy REQ-OWR-005 by being recorded; it is a residual that review, not AC-OWR-08, carries. Required action (no text change needed to proceed): the leader accepts or rejects the residual by name in decision-index.md before Kickoff (the present wording makes acceptance the default). Optional hardening: add the fragmented-path mutant to the meta-test as a pinned known-escape control (asserting the scan does NOT flag it, so the limit is executable documentation), and add `filepath.Glob`, `io/ioutil.*`, `os.Readlink` to the `internal/profile` verb list.

N17. plan.md:105 (M0 item 2) with acceptance.md AC-OWR-17 (L238) and AC-OWR-18 (L249) — "트리 해시" undefined; the prescribed shared helper includes directory mtime — Severity: SHOULD-FIX — Class: optional — Forced runs 2 and 3 rewrite every template-managed file through a temp file and rename, which changes directory mtimes while the bytes stay identical, so an mtime-bearing hash cannot be equal across runs. AC-OWR-16 defines its hash with mtime on purpose (acceptance.md:224); AC-OWR-17/18 do not say which hash they mean. Required fix: say in AC-OWR-17 and AC-OWR-18 that the compared hash is content-only (paths, bytes, mode; the existing `hashDryRunTree` shape), and give the M0 helper a mode switch (mtime on for AC-OWR-16 only).

N18. acceptance.md:272 (AC-OWR-20 When) versus acceptance.md:25 (rule 7 (ii)) — failure cases are told to satisfy a rule that requires a valid new `manifest.prior.json` — Severity: SHOULD-FIX — Class: optional — Case (a) is built to prevent that file. The parenthetical "(skipped == false; 흐름이 Backup 단계와 마이그레이션에 도달했음)" shows the intent. Required fix: state that for cases (a) and (b) reach is evidenced by the writer being invoked and the migration being entered, and apply rule 7 (ii) to case (c) only.

N19. spec.md:75-76 (REQ-OWR-015/-016), design.md D-3 items 3-4, update_template_sync.go:728-740 — no run directory when there is no `.moai/config` to back up — Severity: MINOR — Class: optional — When `configBackupPath` is empty ("No config to backup"), the Backup stage has no run directory, and the SPEC does not say where `manifest.prior.json` goes or whether the missing copy counts as a backup failure under REQ-OWR-016 (zero removals). Required fix: one sentence (create the run directory through the existing on-demand helper, or define the empty case as "no managed removal candidates exist").

N20. acceptance.md AC-OWR-06 (a) and AC-OWR-05 (h1) — the deployer's skipped `UserCreated` entry is not asserted to carry no `deployed_version` — Severity: MINOR — Class: optional — REQ-OWR-015 forbids stamping any entry the run did not deploy; `deployer.go:265-267` tracks that file without writing it. Required fix: assert the field is empty on that entry in (h1).

N21. spec.md:65 (REQ-OWR-005) and acceptance.md:271 (AC-OWR-20 Given (b)) — wording — Severity: MINOR — Class: optional — (i) REQ-OWR-005 speaks of "the frozen existing contacts" and of "the single existing contact"; the AC-OWR-08 baseline freezes every existing file-system call in `internal/profile` (about 28 sites at profile.go, preferences.go, sync.go, read this run) and names one of them as the exception; say so in one clause. (ii) AC-OWR-20 (b) mentions a `removalBackupWriter` seam that design.md D-3 item 5 does not define; say that the `priorManifestWriter` stand-in, which receives `runDir`, also creates the directory at `removed-assets/<rel>`.

Spot-check of file:line coordinates (all read this run at HEAD a2a7b8bb2, all correct): update_template_sync.go:1225 (signature), :1239-1248 (skip gate), :286 (`mgr.Load`), :369 (`applyDeferredSkillRenames`), :609 (`reconcileManagedRoots`), :672-673 (`recordTemplateTree` and `mgr.Save`), :815-817 (migration call inside the Backup step), :944 (`PruneBackups(projectRoot, 5, …)`); migrate_project_assets.go:38-39 (`removalEnabled`), :84 (`backupRemovedAsset`), :185 (`profile.ResolveLaunchProfileForProject`), :195-197 (`d.IsDir()` return), :255-272 (profile-reach keep), :296-302 (gate), :323 (`removedPaths`), :332-336 (`pruneEmptiedSkillDir`), :377 (`userCounterpartConfirmed`); profile.go:439 (excepted `os.Stat`), :193 (`GetProfileDir`), :461 (`ResolveLaunchProfileForProject`); deployer.go:265-267 and :294; update_template_sync_guard.go:55; update_dryrun_preview.go:84 and :120; update.go:361 (dry-run branch); update/reconcile.go:563 (`LoadManifestReadOnly`); update_dry_run_reach_test.go:84 (`runDryRunUpdateInProject`) and :115 (`hashDryRunTree`); update_cleanup.go:387 (`probeCaseSensitiveFS`); update/backup/retention_prune.go:82 (`PruneBackups`); manifest/types.go:93 (`Manifest`); core/project/initializer.go:607-608 (`Version`, `DeployedAt`). 28 coordinates, no mismatch.

## Regression Check

Defects from iteration 2 (N1-N11): ten closed, one (N4) partly closed and carried as N13. No iteration-2 defect is unchanged. Score 0.68 then 0.76 then 0.84: improving, no STOP signal.

## Evidence-bearing sections

### Claim

N1, N2, N3, N5-N11 are closed; N4 is closed in acceptance.md but not in plan.md (N13). Two blocking defects (N12 atomicity unverifiable, N13 stale plan obligations) and the required-backend failure keep the verdict at FAIL; score 0.84 against 0.85. Seven further findings (N14-N21) are optional and may be carried as recorded debt.

### Evidence

Commands run in the tree with observed outputs: `git rev-parse HEAD` = a2a7b8bb24b620c1390bc8e8efa0aacee411bdf1; `moai verify audit-plan --project-root …` = `config_status: ok`, `enforced_required: ["codex"]`; `mcp__moai__audit_multi` = `overall_verdict: fail`, `per_backend_verdicts[codex] = required/fail`, receipt rcpt-48fbda4048c92f444d7ba40d; traceability verb = `COLLECTED: 17 REQ definitions (acceptance input: read)`, no UNCOVERED/ORPHAN; CN-4 verb = `COLLECTED: 11 milestones … 0 exit bindings, 5 ordering candidates`, no CONFLICT; D7 statuses (completed, completed, in-progress); `grep -c syscall` = 0 in 7 files; `grep -rn 'NEEDS CLARIFICATION'` rc 1; `go list -deps … ./internal/cli` = 142 module-local packages with the six core packages present; `git merge-base --is-ancestor cdd72a814 HEAD` rc 0; `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go` rc 0; AC-OWR-08 command 2 and AC-OWR-21 command 2 print nothing (rc 0), command 3 prints 4 lines; `grep -rn 'removalEnabled *=' internal --include='*.go'` (non-test) lists only `migrate_project_assets.go:39`; `grep -n 'Provenance = ' internal/manifest/types.go` = five constants at :17,21,25,29,35; `grep -n '^func [A-Z]' internal/profile/*.go` = 18 exported functions; the new-symbol grep (`runUpdateFlowForced|priorManifestWriter|provenanceSource|classifyOwnership|removalBackupWriter|priorManifestGenerations`) = 0 files; 21 `등급: regression-guard` lines; reads of update_template_sync.go (:215-290, :365-372, :605-612, :668-676, :710-760, :810-820, :940-948, :1225-1260), migrate_project_assets.go (:36-110, :178-345, :376-455), profile.go (:50-76, :425-445), manifest.go (:60-262), deployer.go (:255-300), update_manifest_retrack.go, migrate_removal_backup_test.go (:25-140); 28 line coordinates compared.

### Baseline-attribution

Tree HEAD a2a7b8bb24b620c1390bc8e8efa0aacee411bdf1, this run. Judging build for the moai MCP and CLI tools: installed binary commit ee981a932, a strict ancestor of HEAD (server lag notice); plan_artifact_hash supplied by the lane, not recomputed. Every conclusion about SPEC text rests on file reads and grep/go-list commands run against this tree, not on the lagging binary. SPEC line counts at this HEAD: spec.md 225, acceptance.md 587, plan.md 243, design.md 154, decision-index.md 69, progress.md 78, research.md 195.

### Gaps

- No acceptance-criterion command was executed: the new tests do not exist (a selector would print `ok [no tests to run]`). The baseline-11 command (acceptance.md L-4, about 45 s) was not re-run; it is carried as the SPEC's measurement. I confirmed by grep that no existing test asserts a `backupRemovedAsset` failure or the "removal refused" text, so REQ-OWR-016's collect, verify, remove restructuring cannot turn an existing t1668 test red on its failure path; the success-path tests (`TestMigrationBackupBeforeRemoval`, `TestEmptiedSkillDirRemoved`) were read, not run, and the restructuring moves `pruneEmptiedSkillDir` and the `removedPaths` tally after the walk, which only a run-phase RED/GREEN pair can confirm.
- The codex overlay measurements (`successful-profile-directory-reads=3`, 58 differing directory mtimes, `disk==DeployedHash:true`) are codex's; I checked each by code reading, not by re-running the overlay.
- The convergence digest check (`moai verify audit-plan --result-file …`) was not run: it requires writing `.moai/state/audit-plan-result.json`, which this lane forbade (the verdict file is the only permitted write). The unmet gate is visible in the `audit_multi` result itself (`overall_verdict: fail`, codex `fail`).
- `go list -deps` honors the current platform's build tags; files built only for other operating systems are outside the closure the scan would derive on this host.
- Whether a retention census treats `manifest.prior.json` as a content file (acceptance.md G-K) was not measured; the SPEC records it as an unmeasured possibility within the existing retention budget.
- Whether Claude Code discovers commands by path or by `name` (acceptance.md G-G) and whether the stale-record variant of AC-OWR-09 can be driven through existing seams (G-A) remain unmeasured, as the SPEC states.

### Residual-risk

- Even with N12 and N13 fixed, the profile non-access guarantee stays an inference tool: a fragmented run-time path read with a discarded result is undetectable without an access-recording seam (N16). The leader's by-name acceptance is a Kickoff-time item.
- The MP-8 N/A rests on classification. If a run-phase milestone lands without its B-commit RED cell, nothing before the sync audit notices; the Definition of Done (acceptance.md:337) makes promotion of all 21 a completion condition.
- Landing after t1668 may conflict in `update_template_sync.go` and `migrate_project_assets.go`; the collect, verify, remove restructuring widens that surface (accepted debt, decision-index.md, acceptance.md G-L).
- Plain same-version re-runs still skip (OQ-6); recovery paths are reachable only with `--force` or a template/version change. This is stated clearly and is not a defect.
- Codex may raise further items against the revised hunks on a delta re-audit.

## Iteration history

Iteration 3 of ceiling 3 (Tier L ceiling per `harness.plan_audit_tier_ceilings`). Iteration 1: FAIL 0.68 (audited a8ad48c08, MP-8 failed, 5 blocking). Iteration 2: FAIL 0.76 (audited 3c334dfdd, 0 must-pass failed, 5 blocking, codex fail). Iteration 3: FAIL 0.84 (audited a2a7b8bb2, 0 must-pass failed, 2 blocking, codex fail). Both blocking findings carry a scoped file#anchor fix (see `fix_scope`) and neither changes the REQ/AC id sets, so a delta round stays eligible only if the repair diff stays inside those anchors plus progress.md and .moai/reports/**; the optional findings N14-N21 touch other anchors (design.md#D-2, acceptance.md AC-OWR-05/-17/-18/-20, spec.md REQ-OWR-002/-005) and are therefore best carried as recorded debt unless the leader widens the scope. The CN-4 ordering verb ran in full this iteration. A hunk-limited re-read does not suffice because the required backend must be re-run on the changed hunks; no `reread_hunks` line is written.

## Recommendation

1. N13 (cheapest, unresolved prior blocker): rewrite plan.md M8 step 1 to AC-OWR-15's restore-region equality plus the T0 comparison, and M9 step 1 (a) to the zero-removal interrupted run.
2. N12: add an observable seam to the prior-manifest writer (temp write and rename), assert write-to-temp, final-name absence until rename, and fault-between-write-and-rename through the real writer; update design.md D-3 item 3, plan.md M1b and the § M row 21.
3. If the lane reopens the SPEC at all, fold N14 and N15 (generation precedence; drop the `DeployedHash` alternative), N17 and N18 (one sentence each) into the same commit; otherwise carry N14-N21 as recorded debt. Each is a one-sentence or one-case change.
4. N16: the leader records an accept-or-reject of the fragmented-path residual by name in decision-index.md before Kickoff.
5. Re-run the independent audit with the codex gate; do not treat the unmet codex gate as a reviewed defect — it needs a passing required-backend result.

## Operational Notes (unverified)

- Measure whether two `--force` runs differ in directory mtimes — command: a forced `runTemplateSyncWithProgress` twice in a scratch fixture, then compare a path, byte and mtime hash — status: inferred (the deployer rewrites template-managed files through a temp file and rename unconditionally; codex reported 58 differing directories).
- Measure that a direct-write prior-manifest writer passes the AC-OWR-21 (a) injections — command: write a 10-line direct-write variant behind the `priorManifestWriter` seam and run the three injections — status: inferred (a directory at the final name and an unwritable destination also fail a direct write; the interruption state is hand-built).
- Measure the fragmented-path escape — command: add `strings.Join([]string{"claude","profiles"}, "-")` plus `os.ReadDir` in a scratch overlay and run the scan — status: measured by codex (not re-run here).
- Measure `go list -deps` per target OS — command: `GOOS=windows go list -deps ./internal/cli` and compare the module-local package count with the 142 measured on this host — status: assumption.

AUDIT-VERDICT: FAIL spec=SPEC-UPDATE-OWNERSHIP-REPAIR-001 receipts=rcpt-48fbda4048c92f444d7ba40d
