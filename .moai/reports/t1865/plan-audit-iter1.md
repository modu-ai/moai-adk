auditor-model: claude-sonnet-5-5
verdict: FAIL
audited_sha: a8ad48c08d91bc759fedbf3a6055674aa1e97ae0
overall_score: 0.68
must_pass_failed: 1
blocking_count: 5
plan_artifact_hash: e77f413483348cf50ac23475bbf57c519888458e18385bc6c83f55ac91f95c50
required_backend: codex fail
convergence_overall: fail
scope: full
fix_scope: acceptance.md#AC-OWR-02 acceptance.md#AC-OWR-08 acceptance.md#AC-OWR-10 acceptance.md#RED-cell-rule design.md#D-3 spec.md#REQ-OWR-010 spec.md#K plan.md#M1b acceptance.md#AC-OWR-16
defect_class: D1=ac-wording D2=ac-wording D3=format D4=ac-wording D5=design

# SPEC Review Report: SPEC-UPDATE-OWNERSHIP-REPAIR-001

Iteration: 1/3
Verdict: FAIL
Overall Score: 0.68
Plan Artifact Hash: e77f413483348cf50ac23475bbf57c519888458e18385bc6c83f55ac91f95c50 (supplied by the lane from runtime ComputeHash; not recomputed)
Auditor Version: plan-auditor/v1 (single Claude leg plus the required codex leg via audit_multi)
Tier: L (threshold 0.85). Reasoning context ignored per M1 Context Isolation. Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865, HEAD a8ad48c08d91bc759fedbf3a6055674aa1e97ae0.

Overall: FAIL for three independent reasons. (1) The required codex backend returned `fail` (unmet required gate, not merely a reviewed defect). (2) MP-8 fails. (3) The score (0.68) is below the Tier L threshold 0.85. The structure is strong (15 REQ, 19 AC, traceability verb clean, line references mostly accurate, no profile-root or removal-switch violation in the requirement text); the defects are concentrated in AC verifiability and one design/requirement contradiction.

## Audit plan and backend result

- `moai verify audit-plan --project-root <tree>`: `config_status: ok`, model codex, codex gate `required` (source config.gates), claude `off`, glm `off`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- `mcp__moai__audit_multi` called once (project_root = tree, target baseBranch, focus: provenance-rule enforceability against mutants, profile-root non-access, AC verifiability, Q2 fallback, no `gates` argument). Result: `overall_verdict: fail`, `per_backend_verdicts[codex] = {gate: required, verdict: fail}`, `residual_risk_note: "required-backend FAIL: codex"`, `plan_source: config`, `audit_receipt: rcpt-8a4bab9510ae7274bc21b2b6`, `build_commit: ee981a932` (installed binary is a strict ancestor of the tree HEAD a8ad48c08, so backend output describes the build named, per the server's own lag notice).
- The required backend is `codex`; it did not pass. Overall verdict therefore cannot be PASS or PASS-WITH-DEBT. This is an unmet required gate for the codex leg, and separately the Claude-side reviewed defects below stand on their own.

### Codex findings, quoted verbatim, each checked by this auditor

1. "[P1] 바이트 일치 소유권 변이를 검증할 픽스처의 템플릿 경로가 존재하지 않는다 — .moai/specs/SPEC-UPDATE-OWNERSHIP-REPAIR-001/acceptance.md:34 / 신뢰도: 높음. HEAD `a8ad48c08`에서 지정한 `moai-user-authored/SKILL.md` 두 경로 모두 템플릿에 없다. 따라서 “같은 경로의 내장 템플릿과 바이트 동일”이라는 전제가 성립하지 않아, 해당 일치만으로 managed로 승격하는 변이의 실행을 보장하지 못한다. / 권고: 실제 배포 경로의 템플릿 바이트를 복사하고 매니페스트 기록만 제거하라. 템플릿 조회 성공을 먼저 단언하고, 입력 변경 대신 잘못된 승격 구현을 주입해 시험 실패를 확인하라."
   - CONFIRMED. Observed: `ls internal/template/templates/.claude/skills | grep -c user-authored` → `0`; the `.agents/skills` template tree lists `moai-clean`, `moai-codemaps`, … and no `moai-user-authored`. See D1.
2. "[P1] 프로필 무접근 기준은 읽기 오류를 무시하는 구현을 탐지하지 못한다 — acceptance.md:102 / 신뢰도: 높음. 임시 디렉터리 탐침에서 0o000 디렉터리 읽기는 `PermissionError`를 냈지만 이름·바이트·수정 시각은 그대로였다. 오류를 버리는 읽기를 스캔 대상 밖의 갱신 헬퍼에 넣으면, 동작 결과와 지정한 소스 검사 모두 그대로 통과할 수 있다. / 권고: 프로필 경로의 파일시스템 접근 시도 자체를 계수하거나 실패시키는 검증을 추가하고, 갱신 전체 호출 경로에서 오류를 무시하는 읽기 변이도 검증하라."
   - CONFIRMED by reading acceptance.md:102-107: the pass condition is "오류가 없다 (읽기 시도가 있었다면 권한 오류가 났을 것이다)", which assumes the error propagates; a swallowed read passes. The static scan covers only `migrate_ownership*.go` and `migrate_project_assets.go`. The 0o000 probe is not re-run here (inferred from the AC text; a root-run CI also bypasses 0o000, inferred). See D2.
3. "[P1] 이전 매니페스트 백업 실패 뒤에도 제거를 허용한다 — design.md:73 / 신뢰도: 높음. 설계는 `manifest.prior.json` 쓰기 실패를 경고로 처리하고 계속 진행한다. 이후 파일과 매니페스트 항목을 제거하면 파일 사본만 남아, REQ-OWR-010과 AC-OWR-15가 요구하는 실행 전 매니페스트 복원이 불가능하다. / 권고: 이전 매니페스트 보존 실패 시 제거를 차단하고, 해당 백업 경로의 쓰기 실패를 강제하는 인수 시험을 추가하라."
   - CONFIRMED against design.md:73 ("쓰기 실패는 경고이며 실행을 막지 않는다") versus spec.md:63 (REQ-OWR-010: "When the migration removes a file, the run backup shall hold ... a copy of the prior manifest"). See D5.
4. "[P2] Q2 거절 분기가 유지되는 인수 기준의 구현 의존성까지 제거한다 — plan.md:113 / ... M1b 전체를 접으면 그 단계에 배치된 증거 해석기도 사라지지만, AC-OWR-05는 증거 주입 시험으로 남는다. M5와 AC-OWR-13의 이전 실행 디렉터리 정리도 이전 세대 기록에 의존하는데, 거절 분기는 그 기록을 생산·읽지 않는다. / 권고: 증거 해석기를 영속 기록 생성과 분리하고, Q2 거절 시 유지·축소할 M5 및 AC-OWR-13 범위까지 명시하라."
   - CONFIRMED. plan.md:111 puts `provenanceEvidence` in M1b, plan.md:113 folds all of M1b on "(나)", acceptance.md:72 says AC-OWR-05 survives at function level; spec spec.md:61 / design.md:102 make REQ-OWR-008 depend on prior records. See D6.
5. "[P2] commands의 frontmatter 조건이 템플릿 바이트 동일 조건과 양립하지 않는다 — acceptance.md:124 / ... 현재 commands 템플릿 17개 모두 frontmatter에 `name`이 없다. 따라서 템플릿과 바이트가 같으면서 모든 파일에 비어 있지 않은 `name`과 `description`이 있어야 한다는 기준은 충족할 수 없다. / 권고: commands와 agents의 메타데이터 조건을 분리하고 각각 실제 형식에 맞는 탐색 조건을 단언하라."
   - CONFIRMED. Observed over `internal/template/templates/.claude/commands/moai/*` (17 files: 15 `.md.tmpl`, `gtd.md`, `todo.md`): every file prints `name=0 desc=1`. The agents tree has 0 files lacking `name:`. See D4. This auditor rates it BLOCKING, not P2, because the AC cannot go green for any correct implementation.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-OWR-001 … REQ-OWR-015 sequential, no gap, no duplicate (spec.md:54-68). Traceability verb: `COLLECTED: 15 REQ definitions (acceptance input: read)`.
- [PASS] MP-2 EARS/GEARS compliance (judged on the requirement layer, `REQ-OWR-*` in spec.md §C; ACs were not graded here): every REQ opens with a GEARS pattern token (Ubiquitous 001/006/009/012/015, Unwanted 002/005/011/014, Event-driven 003/007/008/010/013 with `**When**`, State-driven 004 with `**While**`). Minor wording notes only (D13).
- [PASS] MP-3 frontmatter: all 12 canonical fields present with correct types (spec.md:2-14; version quoted `"0.1.0"`, dates ISO, priority P1, lifecycle spec-anchored, tags a comma string, `phase: "v3.0.0 target"`-style release label, not a lifecycle token). Extra fields `tier: L`, `related_specs` are optional. plan.md/acceptance.md/design.md carry no `status:` (artifact statelessness honored).
- [N/A] MP-4 language neutrality: single-language (Go) repository-internal SPEC; no multi-language tool enumeration.
- [PASS] MP-5 D7: referenced SPECs and their `status:` read from disk: SPEC-UPDATE-MIGRATION-001 completed; SPEC-UPDATE-REPAIR-RC30-001 completed; SPEC-UPDATE-REMOVAL-FAMILY-001 `in-progress` (not retired/superseded/archived; reconciled in spec.md:147 and why it is not in `depends_on`). No `REVIEW:` candidate, no BLOCKING.
- [PASS] MP-6 D8: `grep -c syscall` = 0 in all five artifacts. Auto-pass.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → no match (rc=1). Note: Q2 is an open decision-index row, not a marker.
- [FAIL] MP-8 RED-now cell re-execution: applicable, because acceptance.md labels criteria "릴리스 게이트" (for example AC-OWR-01 at acceptance.md:30 "TDD RED 마일스톤 (M1a) 의 릴리스 게이트", and AC-OWR-02/-03/-04/-05/-06/-07/-08/-11/-12/-13/-14/-15/-16/-17/-18 likewise). No criterion carries a RED-now cell: acceptance.md records a command and an expected output, but no verbatim stdout, no exit code and no tree SHA (the only SHA, `cdd72a814`, appears inside AC-OWR-08's diff command, not as a pin). acceptance.md:14 defers the cells to the run phase ("plan.md 의 RED 작업이 구현 전에 기록한다 … 그 기록이 있기 전의 기준은 회귀 가드이며 릴리스 게이트가 아니다"), which contradicts the per-AC "릴리스 게이트" lines. Re-execution: not performed, because the cited tests do not yet exist in the tree (new files in plan.md §F), so a `-run` selector would sweep zero tests and print `ok` / `[no tests to run]`, which is neither RED nor a pass (acceptance.md:15 and AC-OWR-08:106 say the same); running a package-level `go test ./internal/cli/` is also outside this audit's execution discipline. See D3.
- [PASS] MP-9 cross-artifact ordering: CN-4 verb → `COLLECTED: 11 milestones in plan order (M0 M1 M2 M3 M4 M5 M6 M7 M8 M9 M10), 0 exit bindings, 4 ordering candidates`; plan.md carries no `Exit:` lines so no `CONFLICT:` can be produced. The four candidates (acceptance.md:125, 126, 180, 237) are test names containing "After"/"Before" (`...AfterInterruptedRun`, `TestMigrationBackupBeforeRemoval`) and order nothing. Manual read: M1b (manifest.prior.json) precedes its consumers M5 and M8; M1a precedes M7; no clause demands the reverse. No conflict.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | One interpretation split on where `removal-eligible` is decided: design.md:28 says the verdict marks candidates "앞까지" of the t1668 gate (switch, backup, counterpart check applied after), while acceptance.md:25 (AC-OWR-01), :68 (AC-OWR-05) and :188 (AC-OWR-16) put the switch and the counterpart check inside the verdict (`keep(removal-gate-off)`). AC-OWR-09:114 allows two outcomes with "또는". |
| Completeness | 0.75 | 0.75 | All sections present (HISTORY spec.md:20; Out of Scope as H3 `### Out of Scope — …` with bullets spec.md:172-196). Incomplete: no RED-now cells (D3); the Q2 "no" drop list omits REQ-OWR-008 and AC-OWR-13/-18 sub-cases (D6). |
| Testability | 0.50 | 0.50 | AC-OWR-02 fixture cannot stage its own premise (D1); AC-OWR-08 passes a swallowed profile-root read (D2); AC-OWR-10 unsatisfiable for commands (D4); AC-OWR-16 verifies a function while REQ-OWR-011 constrains the command (D7). Most other ACs have an anchored `-run '^…$'` selector, an expected `--- PASS:` line, and a stated mutant. |
| Traceability | 0.75 | 0.75 | Verb output: 15 definitions, no `UNCOVERED:`, no `ORPHAN:`. Chain item → REQ → AC → milestone is complete (spec.md:88-104, plan.md §G cites AC per milestone). Indirect: AC-OWR-06(b) asserts manifest-level `Version`/`DeployedAt` behavior that REQ-OWR-015's text does not state. |

Aggregate (mean) 0.69, reported 0.68 after rounding down for the MP-8 failure and the unmet required backend. Below the Tier L PASS threshold 0.85.

## Defects Found

BLOCKING

D1. AC-OWR-02 fixture path has no template counterpart — acceptance.md:34 — Severity: BLOCKING — Class: blocking — The Given states the files `.claude/skills/moai-user-authored/SKILL.md` and `.agents/skills/moai-user-authored/SKILL.md` are byte-identical to "내장 템플릿의 같은 경로 파일", but neither path exists in the embedded template (observed: 0 matches for `user-authored` under `internal/template/templates/.claude/skills`; `.agents/skills` lists `moai-clean`, `moai-codemaps`, …). Mutant constructed: an implementation that promotes any byte-equal-to-template file to `managed` and removes it never fires on this fixture, because no same-path template file exists, so the mutant passes AC-OWR-02 while violating REQ-OWR-002. This is the card's required preservation fixture 1 ("template-identical user file = keep"). Required fix: build the fixture from a path the embedded tree really carries (for example a file under an existing `.claude/skills/moai-*/` and its `.agents/skills/` mirror), delete only the manifest entry, assert as the first step that the template lookup for that path succeeds and the bytes are equal, and add the mutant (byte-equal + confirmed counterpart + switch on → removed) as a negative control the test must see go red. State the same requirement for AC-OWR-14's 313-file fixture (acceptance.md:164 says only "템플릿과 바이트가 같다").

D2. AC-OWR-08 cannot detect a swallowed or out-of-scope profile-root read — acceptance.md:100-107 — Severity: BLOCKING — Class: blocking — The behavioral check passes when "오류가 없다 (읽기 시도가 있었다면 권한 오류가 났을 것이다)"; an implementation that reads the 0o000 directory and ignores the error passes. The static scan covers only `migrate_ownership*.go` and `migrate_project_assets.go`, so a helper in any other file (`update_template_sync.go`, `update_dryrun_preview.go`, a new `userassets` helper) is unscanned. Hard constraint (a) is "no profile-root access of any kind". The 0o000 trick also does nothing when the suite runs as root (inferred). Required fix: (i) widen the source scan to the whole call path of the update flow (all non-test files of the packages the migration and preview reach), matching `claude-skills@`, `claude-agents@` and any use of `CLAUDE_CONFIG_DIR`/profile base dir as a read target; (ii) add an access-counting seam (an injectable filesystem or a wrapper that fails the test when any path under `profileDir` is opened/stat'd, errors included); (iii) drop the root-sensitive 0o000 reliance or guard it. Also pin AC-OWR-08's file-invariance diff to a recorded SHA and say what happens once main moves (the current `cdd72a814` pin is correct for now but silently becomes a false red if a legitimate t1702 change lands in this tree).

D3. Release-blocking ACs carry no RED-now cell, and the SPEC contradicts itself on whether they are release gates — acceptance.md:14, :30 and the per-AC "게이트" lines — Severity: BLOCKING (forces MP-8 FAIL) — Class: blocking — verification-completeness.md §2/§2.1 requires command + verbatim stdout + exit code + tree SHA per release-blocking criterion; none is present, and acceptance.md:14 defers them to run time while the AC bodies still say "릴리스 게이트". Required fix: choose one. Either (a) reclassify every AC whose RED cannot be observed at plan time as `regression-guard` in the "게이트" lines and state that the run-phase RED task records the cell before the first GREEN commit (keeping :14), or (b) record the cells now by landing the failing tests as a separate preceding commit (also satisfying verification-claim-integrity.md §2.3 ordering). Note the stated RED reasons for AC-OWR-01/-05 ("빌드 실패 … undefined: classifyOwnership") describe the state after the test file exists; before it exists the selector sweeps zero tests. Say so, as acceptance.md:106 already does for AC-OWR-08.

D4. AC-OWR-10 is unsatisfiable for commands — acceptance.md:124 — Severity: BLOCKING — Class: blocking — The Then requires "두 디렉터리는 템플릿이 운반하는 집합과 이름·바이트가 같고, 각 파일의 frontmatter 에 비어 있지 않은 name/description 이 있다". Observed: all 17 template command files (`internal/template/templates/.claude/commands/moai/*`) have a `description:` but no `name:` in their first 12 lines (`name=0 desc=1` for each). A byte-identical command file can therefore never also carry a non-empty `name`. (Agents are fine: 0 agent templates lack `name:`.) A correct implementation cannot turn this AC green, which makes it a wrong-reason red / impossible criterion. Required fix: split the condition: commands assert name-and-bytes equality plus non-empty `description` (and the presence the loader actually requires, which the SPEC should cite); agents assert non-empty `name` and `description`. Also state that `.md.tmpl` files are compared after rendering for the fixture's language, since "bytes same as template" is ambiguous for templated files.

D5. Design lets removal proceed without the prior-manifest copy REQ-OWR-010 requires — design.md:73 versus spec.md:63 — Severity: BLOCKING — Class: blocking — design.md:73: "쓰기 실패는 경고이며 실행을 막지 않는다". REQ-OWR-010: when the migration removes a file, the run backup "shall hold ... a copy of the prior manifest, such that returning both ... reproduces the pre-run bytes of ... the manifest". An implementation following the design passes AC-OWR-15 (which exercises only the success path) and still removes a file and its manifest entry with no manifest copy, violating card constraint (7), byte-identical backup/restore. Required fix: either make the failure of the `manifest.prior.json` write block removal for the run (keep-and-record reason `backup-failed`, already in design.md:30's reason set), or weaken REQ-OWR-010 to say what holds on failure; add an AC sub-case that forces the write to fail and asserts no file is removed (acceptance.md AC-OWR-15 or AC-OWR-18(a)).

SHOULD-FIX

D6. The Q2 "no" fallback is under-specified — spec.md:70, decision-index.md:19 (option 나), plan.md:111-113, design.md:102 — Severity: SHOULD-FIX — Class: blocking (fix before Kickoff, since Q2 is the operator's gate) — Stated as dropped on "(나)": REQ-OWR-015, the manifest-copy half of REQ-OWR-010, and the operational halves of AC-OWR-05/-06/-15. Not stated: (i) REQ-OWR-008 ("whether this run or an earlier one emptied it") and AC-OWR-13 cases A and C, plus AC-OWR-18(c), also need the prior-generation records that "(나)" never produces (design.md:102 collects records from S1 and S2; with S2 gone only S1 remains); (ii) the evidence interpreter `provenanceEvidence`/`provenanceSource` is placed in M1b (plan.md:111) yet AC-OWR-05 survives "(나)" at function level (acceptance.md:72), so folding M1b (plan.md:113) removes code the surviving AC needs. Required fix: move the interpreter and its interface to M1a (or a new M1a'), and list REQ-OWR-008/AC-OWR-13(A,C)/AC-OWR-18(c) explicitly as kept-reduced or dropped on "(나)". The decision row itself is clear and correctly placed (FOUNDER, product-level, blocks autonomous Kickoff at spec.md:200); what it lacks is the full consequence list the operator should see.

D7. AC-OWR-16 verifies a function, not the command REQ-OWR-011 constrains — acceptance.md:187 — Severity: SHOULD-FIX — Class: blocking (REQ coverage) — REQ-OWR-011 says "The dry-run shall not create, modify, or delete any file or directory". The AC drives `previewReconciliation` only ("`moai update --dry-run` 에 해당하는 미리보기"). Another dry-run code path (for example the `.moai-fscase-probe` write of `probeCaseSensitiveFS`, update_cleanup.go:387, which design.md:86 admits writes into the project root) is not exercised. Required fix: drive the full dry-run entry (the function the CLI dispatches for `--dry-run`) over the mixed fixture and compare the tree hash, not only the preview function.

D8. `removal-eligible` is defined inconsistently — design.md:28 versus acceptance.md:25, :68, :188 — Severity: SHOULD-FIX — Class: blocking (internal consistency) — Design: the verdict names candidates "앞까지" of the t1668 gate and the gate (switch, counterpart, boundary) acts afterwards; ACs: the switch and the counterpart result are inputs to the verdict (`keep(removal-gate-off)`). REQ-OWR-011 equates the preview's announced set with the verdict's `removal-eligible` set, so the two readings give different sets when the switch is off. Required fix: state once which gates are inside the verdict, and align D-1 and the ACs to it.

D9. M6's "measure first" is sound as a fallback but its fixture may miss the diagnosed cause — plan.md:150, acceptance.md:112-117 — Severity: SHOULD-FIX — Class: optional — The field symptom came from a project whose template record (rc.35) differed from the running binary (rc.36 dirty); the AC-OWR-09 fixture is described as four file states with no version delta or stale `TemplateTreeHash`, so "not reproduced" could be a false negative of the fixture (inferred; the cycle was not reproduced by code reading either, research.md G-2). Falling back to regression guards plus a report to the leader is acceptable; the risk is stopping after one same-version negative. Required fix: include a version-delta/stale-record variant in the M6 measurement and record it in the RED cell.

D10. TMPDIR directory is not created by any step — acceptance.md:13, plan.md:47 — Severity: SHOULD-FIX — Class: optional — Every judging command is prefixed `TMPDIR=/private/tmp/own-fixture-d1`; the directory existed at audit time (`drwxr-xr-x ... Oct 18 12:20`, created after the plan artifacts), but no pre-flight step makes it, and Go's `t.TempDir` fails when `TMPDIR` is missing. Add `mkdir -p` to plan.md §C.

D11. AC-OWR-06(b) asserts behavior with no REQ text — acceptance.md:79 — Severity: SHOULD-FIX — Class: optional — manifest-level `Version` refresh and `DeployedAt` change rule appear in design.md D-3 and AC-OWR-06/-17 but REQ-OWR-015 states only per-entry version and prior-manifest retention. Fold the rule into REQ-OWR-015 or REQ-OWR-012.

MINOR

D12. Off-by-one reference — spec.md:162 cites `internal/core/project/initializer.go:607-608` for `Version`/`DeployedAt`; observed `mf.Version = version.GetVersion()` at 606 and `mf.DeployedAt = …` at 607. — Class: optional. (Ten other citations spot-checked correct: migrate_project_assets.go :38-39, :296-302, :255-272, :84, :323, :332-336, :400, :436, manifest/types.go:93-113, update_cleanup.go:387, update_dryrun_preview.go:84/:120, update_template_sync.go :285-286, :369, :672-673.)

D13. Requirement wording carries implementation detail — spec.md:61 (`empty-dir-no-provenance`), :68 (field-level "record the version of the moai binary"), :58 (file names); AC-OWR-01 "분류 상수의 집합은 정확히 이 네 값이며 시험이 그 수를 단언한다" (acceptance.md:25) is satisfiable by a fifth constant outside the counted slice. — Class: optional.

D14. AC-OWR-09 Then offers an either/or ("보존되거나 `.moai-new` 곁파일과 함께 병합되며") — acceptance.md:114 — Class: optional; pick one observable outcome per fixture state.

D15. Merge-conflict exposure — plan.md §F touches `update_template_sync.go` in four milestones (M1b, M3, M6, M8) and `migrate_project_assets.go` in four (M1a, M2, M3, M5); plan.md §D5 already narrows the changes and §A/§H record that landing follows t1668. Acceptable; track that a later main merge may conflict in these two files. — Class: optional.

## Probe notes (what was tried, for the iteration-2 delta)

- Mutant probe for "byte-equality alone is not provenance": AC-OWR-05's four mutants (drop version, relocated key, absent prior record, hash mismatch) do force the three-evidence rule at function level; AC-OWR-14 forces it at scale only if its 313 files really come from the embedded tree (see D1). The remaining way for a mutant to survive is the fixture-premise gap in D1, not the evidence logic.
- Profile-root non-access: requirement text (REQ-OWR-005, spec.md:58) and the F.1 constraint are correct and consistent with Q5; defect is verification strength only (D2). `internal/config/userroots.go` is absent in this tree and `protected_zone.go` is present; `git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go` → exit 0 (observed), cdd72a814 is an ancestor of HEAD (observed).
- Safety gate: no REQ or milestone enables `removalEnabled` in production (REQ-OWR-014, AC-OWR-19, plan.md §H); no milestone needs the real home, mo.ai.kr, a `moai update` re-run or a binary swap.
- Tier L: 15 REQ and 19 AC are within the 25/25 ceilings; classification and five artifacts are consistent.

## Evidence-bearing sections

### Claim
The SPEC does not meet the plan-audit bar: one must-pass criterion (MP-8) fails, the required codex backend failed, five blocking defects stand, and the score is 0.68 against 0.85.

### Evidence
Commands run in the tree, observed outputs: `git rev-parse HEAD` → `a8ad48c08d91bc759fedbf3a6055674aa1e97ae0`; `moai verify audit-plan --project-root …` → `enforced_required: ["codex"]`; `mcp__moai__audit_multi` → `overall_verdict: fail`, `audit_receipt: rcpt-8a4bab9510ae7274bc21b2b6`; traceability verb → `COLLECTED: 15 REQ definitions (acceptance input: read)` with no UNCOVERED/ORPHAN; CN-4 verb → `COLLECTED: 11 milestones … 0 exit bindings, 4 ordering candidates`; MP-7 grep rc=1; D8 `syscall` count 0 in five files; D7 statuses listed above; template greps for `user-authored` (0) and command `name:` (0 of 17); file-existence checks for the 13 baseline/helper test functions (all found under internal/cli) and for `classifyOwnership`/`provenanceEvidence`/`resolveCaseAlias` (none exist yet, as planned).

### Baseline-attribution
Tree HEAD a8ad48c08d91bc759fedbf3a6055674aa1e97ae0, this run. Judging build for the moai MCP tools: installed binary commit ee981a932, a strict ancestor of HEAD (server lag notice); the backend text describes that build. Plan artifact hash supplied by the lane, not recomputed.

### Gaps
- The cited judging tests do not exist, so no AC command was executed; MP-8 is decided on the absent cells, not on a reproduced or failed RED.
- The 0o000 permission probe quoted by codex (finding 2) was not re-run; confirmed from the AC text only.
- The convergence-result digest check (`moai verify audit-plan --result-file …`) was not run: this audit was restricted to writing only the verdict file. The unmet gate is already visible in the audit_multi result itself.
- Whether `profile.ResolveLaunchProfileForProject` (migrate_project_assets.go:184) reads anything under a profile base directory was not measured; AC-OWR-08 would not catch it either way.
- research.md and the research claim "11 baseline tests GREEN" were not re-executed; only the existence of their test functions was checked.
- No tests, build, lint or coverage were run (plan-phase audit).

### Residual-risk
- Fixing D1-D5 may not suffice for the codex leg; its remaining P2 items are folded into D4 and D6, and it may raise new items against the revised ACs.
- Even with a strong AC-OWR-02/-14, the attested-managed path (Q2 "(가)") makes `manifest.prior.json` under the user-writable `.moai-backups` directory a trust input; this audit rates that as optional because the same user can already edit the files, but the operator should see it with Q2.
- Landing after t1668 may conflict in the two files named in D15.

## Operational Notes (unverified)

- Measure whether the template command files are rendered before byte comparison — command: `grep -c '{{' internal/template/templates/.claude/commands/moai/plan.md.tmpl` — status: assumption (output above shows template directives in `plan.md.tmpl`).
- Measure which package test selectors still sweep zero tests after the RED commit — command: `go test ./internal/cli/ -run '^TestOwnershipVerdictClasses$' -count=1 -v` — status: inferred (tests absent at HEAD a8ad48c08; not executed here).

## Regression Check

Iteration 1: no prior-iteration defects.

## Iteration history

Iteration 1 of 3 (Tier L ceiling per `harness.plan_audit_tier_ceilings`). Score 0.68. Next iteration scope: delta on D1-D5 (required), D6-D8 (required before Kickoff), plus the full CN-4 ordering verb re-run.

## Recommendation

1. D1: rebuild AC-OWR-02 (and state it for AC-OWR-14) on a real embedded-tree path with a template-lookup success assertion and a red-observed mutant.
2. D2: widen AC-OWR-08 to the whole update call path and add an access-recording seam; stop relying on 0o000 read errors.
3. D3: reconcile the release-gate labels with acceptance.md:14 (regression-guard until the run-phase RED cell exists, or land the failing tests first and record four-element cells).
4. D4: split AC-OWR-10 into a command condition (description only, byte-equal) and an agent condition (name and description).
5. D5: make a failed `manifest.prior.json` write block removal (or amend REQ-OWR-010) and test the forced failure.
6. D6-D8: complete the Q2 "no" consequence list and move the evidence interpreter to M1a; drive AC-OWR-16 through the full dry-run entry; define `removal-eligible` once.
7. Optional: D9-D15.

AUDIT-VERDICT: FAIL spec=SPEC-UPDATE-OWNERSHIP-REPAIR-001 receipts=rcpt-8a4bab9510ae7274bc21b2b6
