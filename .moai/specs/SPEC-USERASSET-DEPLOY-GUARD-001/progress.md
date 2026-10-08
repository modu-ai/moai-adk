# progress.md — SPEC-USERASSET-DEPLOY-GUARD-001

status: draft
card: t1591 (lane-20)
tree: .moai/worktrees/t1591 @ WT-user-asset-bundle

## Card Provenance

- 리더 배차 2026-10-08 (Class C — plan→run→sync 전체). 본 SPEC은 카드의 13개 원장 항목을 그대로 범위로 한다.
- 원장 출처: t1561 게이트 P2 4건 + 배포 표면 (리더 발행 2026-10-08) + t1547 릴레이 r10 (원장 8) + t1574 3차 릴레이 (원장 9·10) + t1498 게이트 원장 계통 + lane-15 (원장 11)·lane-5 (원장 12)·lane-7 대장 (원장 13) 릴레이.
- 계통: t1509 "사용자 자산·배포 축" 보존 계약 후속 (릴레이 누계 7건). t1560 게이트 이관분은 SPEC-PROGRESS-RECORD-IO-001 원장에 상호참조 존속 — 본 카드 비소관.
- 비소관 축: t1547 도메인 정합 (update/reconcile.go:169/:590/:204/:184, update_template_sync.go:541 — 리더 보유 후속 카드), 형제 카드 t1594 (t1547 r5-r12 잔여).

## Gate-Relay Provenance (라운드 1-13)

- 누적: t1547기 턴종료 게이트에서 라운드 1-13. **오버레이 재현만 수행, 전체 CI/windows 미검증은 게이트 스스로 선언.**
- 수렴: r8에서 8좌표 수렴 선언, r9-13이 잔여 추가.
- r4 검증 라인 (역사적 verbatim, 고정 HEAD 40e6c310):
  `go test -p 2 -overlay /tmp/t1547-review-overlay.json ./internal/userassets ./internal/cli -run '^TestReview(FIFOInstall|RemovePendingBundle|InitRetryAfterInstallFailure|CodexInstalledSchemaReference)$' -count=1 -timeout 120s -v` → `--- FAIL` 4케이스.
- 스테일 경고: `/tmp/t1547-review-overlay.json` 등 이전 세대 오버레이는 의존 금지 — 본 카드의 재현은 본 트리에서 새로 작성 (M0).

## Intake Record

- 베이스 흡수: 81786284e → db0c514d3, 16커밋, FF 병합 — 인테이크 시 관측 완료. t1547 수리 라운드 일부 포함 → 릴레이 일부가 이미 수리 상태 (재고정 표 11a·5b 행).
- 브랜치 개명: WT-t1561-p2-4 → WT-user-asset-bundle (관측). 트리 디렉터리는 카드 id 유지(t1591).
- 임대 기록 (양립 기록): 인테이크 시 `factory next` 거부 **2회** (직렬 슬롯 — 진행 중 형제 카드 점유) → 이후 임대 **성공** (2026-10-09 ~00:04, factory next가 카드 행 t1591 stage=- worktree=t1591 반환). 장부(record) 정산은 리더 소관.
- 경계 준수: 본 에이전트(manager-spec)는 `.moai/specs/SPEC-USERASSET-DEPLOY-GUARD-001/` 외 수정 없음. 커밋 없음(레인이 스테이지 경계에서 커밋), 큐 변경 없음, 사용자 질의 없음.

## Re-Anchor Table (HEAD db0c514d3 — 관측 요약)

전문(인용 라인 포함)은 research.md §1. 라벨: observed / relayed / repaired / retired / indeterminate.

| # | 원장 좌표 | 판정 |
|---|---|---|
| 1, 10-P1 | lock_guard_windows.go:23 | observed — 마커 무소유권·무회수 (lock.go .lock과 비대칭) |
| 2 | doctor_user_install.go:134/:165 | observed — 적재 오류 OK 위장·공허 일치 |
| 3 | install.go:239 | observed — 첫 스테이징 전 회수 항목 병합 부재 |
| 4 | migrate_project_assets.go:89/:119 | observed — 미등록 미러 사본 영구 보존 |
| 5 | deployer.go:192 / :347 / bundle.go:145 | observed / **repaired**(ListTemplates 제외, 758→414 코멘트) / **retired**(template/bundle.go 부재) / emission compared 축은 doctor_agentemit_embed.go로 재고정 |
| 6a | install.go:297 | observed — 반입 항목 해시 미갱신 |
| 6b | install.go:743 | observed — 0o644 하드코딩 (.sh 실행권 상실) |
| 7a | install.go:392 | observed(무변환 복사 경로) — 내용 주장 relayed |
| 7b | doctor_harness.go:58 | observed — skillsDir 통째 교체 |
| 8a | agentfm.go:128/:282 | observed — 중복 행·PostFormValue 첫 행 |
| 8b | install.go:459 | observed — 클로저 전개 부재 (prune R-f-② 유예 팔은 존재 — 변별 재현 필요) |
| 8c | install.go:211 | observed — ReadFile 직접 판정 (FIFO 블록) |
| 8d | deployer_mode.go:37 | **retired** — M7 퇴역, 생 면 internal/cli/skills.go로 재표현 |
| 9a | paths.go:38 | observed(루트 정의) — 집합 불포함 relayed (config 적재면 pre_tool.go:355 관측) |
| 9b | doctor_harness.go:57 | observed — 빈 디렉터리 존재만으로 선택 |
| 10 | install.go:255 | observed — 플래그 일괄 영속화 |
| 10a | install.go:260 | **indeterminate** — SchemaVersion=1 뿐, v1/v2 좌표 부재 → REQ-JRN-004 재정식화 |
| 11a | update.go:565 | **repaired** — 확인창 뒤 이전 코멘트+호출 관측 → 회귀 가드 |
| 11b | init.go:947/:905 | observed — 재개 경로 부재 |
| 12 | install.go:750 | observed — rename 직전 재검증 없음 |
| 13a/13b | install.go:373/:584 | observed(경로) / **relayed-unverified** (런 재현 확정 대상 — 카드 명시) |

관측 방법 선언: 본 표의 모든 "observed"는 plan-phase(2026-10-09)의 Read/grep 관측이다. 어떤 테스트도 실행하지 않았다 — 결함의 런타임 재현은 전부 런 M0 소관이다.

## Open Questions for Plan-Audit

1. **원장 11a 이미 수리** — RED-first가 재현 불가(GREEN)로 끝나면: REQ-SRF-006 회귀 가드 확정 + 원장 항목 소관 종료 표기로 처리할지 (리더 처분 동반).
2. **원장 5 부분 수리** — ListTemplates 제외 분항 재수리 금지 확인. 생존 분항(SRF-002/003)만 본 SPEC이 소관.
3. **원장 8d 좌표 퇴역** — skills.go 재표현이 원장 의도와 일치하는지 (deployer_mode.go:37의 원래 메커니즘 소멸 확인됨).
4. **원장 10a 재정식화** — v1/v2 오판 주장 → REQ-JRN-004(버전 게이트) 전환이 타당한지, 아니면 원장 좌표 소관 종료인지.
5. **원장 13a/13b 미검증 릴레이** — M0 재현이 반박 시 REQ-CNV 적용 범위 축소 절차 (7a만 잔존) 사전 승인 요청.
6. **template/bundle.go:145 부재** — 원장 좌표 소관 종료 또는 cli/bundle.go 재지정 판정.
7. **원장 8b prune 절반** — remove.go R-f-② 유예 팔과의 관계 변별을 M0에 포함할지.
8. **windows 커버리지** — 원장 1의 검증 구성(표 테스트 + GOOS 빌드 게이트, REQ-LOCK-002)이 plan-audit 기준에 충분한지.
9. **web 폼 계약 변경(8a)** — 폼 키 스코프화의 호환 기간 필요성.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-09
- 아티팩트 셋: Tier L 5종 (spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md — 동시 발행 완료.
- 사전 검증: SPEC ID 정규식 `PASS` (Bash verbatim), frontmatter 12 필수 필드 적합, ID 유일성 확인(`SPEC-USERASSET-*` 0건).
- 독립 감사 종결 — plan-audit iter3+Addendum 5 **PASS-WITH-DEBT 0.94**(기준 0.85 상회·blocking 0·audited_sha 7d2e2a36e·receipts rcpt-cf62fd3b4adf15332e74b11e, rcpt-9941027c067ed9cab2bcf821)·부채 2건(R4→run M6, R5→sync).

## §E.2 Run-phase Evidence

### M0 — RED-first 재현 기준선 (2026-10-09, HEAD e594047b2, 트리 WT-user-asset-bundle)

관측 방법: 신규 테스트 파일 8개(B10 — 프로덕션 코드 무변경)를 본 트리에 작성하고,
acceptance.md §C의 앵커 -run 셀렉터로 1회 실행. 판정 규율: `EXPECTED_RED`(의도 단정
실패)만 RED로 인정. verbatim 전문은 아래 E8에, 실행 로그 전체는 레인 로컬 스크래치에
있으나 본 문서가 유일한 인용면이다(스크래치 미인용 규율).

**E1. 항목별 매트릭스 (27 셀렉터 sweep — 22 의도 RED + 5 born-green PASS + windows축 SKIP)**

| 항목 | 테스트 | 분류 | 관측 vs EXPECTED |
|---|---|---|---|
| 7a→AC-006 | TestCodexAgentTOMLReferencesConverted | RED-first | RED 일치 — verbatim 복사 3참조 전부 잔존 |
| 13a→AC-007 | TestCodexUserSkillInstallNoClaudeOnlyRefs | RED-first(조건부) | **REPRODUCED** — Codex 루트 스킬 무변환 참조 잔존 |
| 13b→AC-007 | TestCodexRoleTOMLNotVerbatim | RED-first(조건부) | **REPRODUCED** — TOML이 소스와 byte-identical → AC-006 동일 결함으로 확정, 13b는 AC-006에 흡수 |
| 13→AC-026 | TestCodexUnconvertibleReferenceReported | RED-first | RED 일치 — 결과 보고에 파일 단위 나열 부재 |
| 8b→AC-016 | TestBundleDependsOnClosureInstall | RED-first | RED 일치 — depends_on 클로저 미전개 |
| 8c→AC-010 | TestCollisionPrecheckSkipsFifoWithoutBlock | RED-first | RED 일치 — FIFO에서 10초+ 블록 (watchdog 판정) |
| 12→AC-011 | TestConfinedWritePinnedToValidatedParent | RED-first | RED 일치 — 검증 후 부모 swap에서 외부 센티널 탈출 관측 |
| 6b→AC-025 | TestConfinedWritePreservesExecBit | RED-first | RED 일치 — 설치 .sh 모드 644 (하드코딩) |
| 10→AC-004 | TestWriteCompletedPersistedPerFile | RED-first | RED 일치 — 중단 시점 저널 플래그 false |
| 3→AC-002 | TestJournalStageCarriesRecoveredEntries | RED-first | RED 일치 — 첫 스테이징에 회수 항목 부재 |
| 6a→AC-003 | TestJournalRefreshUpdatesCarriedHash | RED-first | RED 일치 — 반입 항목 구 해시 유지 (새로고침 발생 확인 후) |
| 10a→AC-005 | TestJournalUnknownSchemaRefused | RED-first | RED 일치 — schema_version=99 무음 디코드 |
| 1/10-P1→AC-001 | TestGuardMarkerReclaimAfterOwnerDeath | 분할 | unix 2팔 PASS(born-green) + windows 팔 SKIP(darwin 미실행 — E Gaps) |
| (형태 문서) | TestLockGuardPathShape | born-green | PASS — guard 경로 파생 플랫폼 공통 고정 |
| 4→AC-012 | TestMigrationClassifiesUnregisteredMirrorCopy | RED-first | RED 일치 — 미등록 사본 무보고(untouched 집계만) |
| 2→AC-019a | TestDoctorProjectManifestLoadFailureNotDisguised | RED-first | RED 일치 — CheckOK 위장 |
| 2→AC-019a | TestDoctorProjectManifestPreservedOnLoadFailure | RED-first | RED 일치 — 원본 이탈 + .corrupt 파괴(manifest.go:86 rename) |
| 2a→AC-019b | TestDoctorUserInstallHonestFailureAndReadOnly | born-green | **GREEN 상륙** — 정직 보고 + 읽기전용 유지 |
| 2b→AC-020 | TestDoctorLockCheckVacuousNotReportedMatch | RED-first | RED 일치 — 0면 공허 일치 OK 위장 |
| 9b→AC-021 | TestDoctorWorkflowRootSelectedByRequiredFile | RED-first | RED 일치 — 빈 프로젝트 dir가 존재만으로 승리, L4 FAIL |
| 7b→AC-022 | TestDoctorFallbackKeepsProjectScopeL1 | RED-first | RED 일치 — L1이 사용자 스코프로 교체, 프로젝트 결함 은닉 |
| 5d→AC-013 | TestDoctorAgentEmissionUncomparedNotOK | born-green | **GREEN 상륙** — compared 1/2 → CheckFail 유지 |
| 8d→AC-014 | TestSkillsDisableResolvesUserInstalledSkill | RED-first | RED 일치 — 미러 부재 시 사용자 면 도달 불가 (MirrorAbsent) |
| 11a→AC-017 | TestUpdateCancelKeepsProjectAssetsIntact | born-green | **GREEN 상륙** — 취소 시 프로젝트 자산 byte-intact (r5 수리 유지) |
| 11b→AC-018 | TestInitResumeAfterUserAssetEnsureFailure | RED-first | RED 일치 — 재실행 "project already initialized" 거절 |
| 8a→AC-015 | TestAgentFormSameNameConsolidatedEndToEnd | RED-first | RED 일치 — 동명 에이전트 2행 렌더 |
| 9a→AC-009 | TestFrozenGuardDeniesUserPathDelete | RED-first | RED 일치 — 사용자 루트 삭제 4형태 전부 allow |

**E2. 빌드 게이트**

```
$ go build ./...                           → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ GOOS=windows GOARCH=amd64 go build ./internal/userassets ./internal/cli → exit 0  (AC-024)
```

**E5. Lint (baseline vs new)**

```
$ golangci-lint run --timeout=2m ./internal/userassets/... ./internal/cli/... ./internal/web/... ./internal/hook/...
→ exit 0, "0 issues." — 신규(m0_* 8파일) 발분 0건, baseline 발분 0건 (this run, this tree e594047b2)
```

**E8. RED 실패 출력 (verbatim — 의도 단정 실패의 관측 라인)**

- AC-006: `installed TOML still carries the unconverted reference ".claude/rules/moai/"` + `".claude/skills/moai/workflows/"` + `"CLAUDE.md"` + `no converted face found in the installed TOML — NormalizeCodexRoleForDeploy is not wired into the install path (install.go:392)`
- AC-007/13a: `REPRODUCED 13a: the installed Codex-root skill still carries the unconverted reference ".claude/rules/moai/"` + `"CLAUDE.md"`
- AC-007/13b: `REPRODUCED 13b: the installed role TOML is byte-identical to the reference-laden source — same verbatim-copy defect as AC-006 (install.go:392); 13b absorbs into AC-006`
- AC-026: `the install result report names no file for the unconvertible reference — report was: (빈 보고)`
- AC-016: `the depends_on dependency bundle (extras2) was not installed with the selected bundle (extras) — collectEntries has no closure expansion`
- AC-010: `the collision precheck blocked for over 10s on a FIFO target — os.ReadFile at install.go:211 reads the FIFO with no writer and never returns`
- AC-011: `the write escaped the validated parent after the post-validation swap — external sentinel received .../002/sub/file.txt (rename at install.go:750 followed the swapped parent)`
- AC-025: `installed .sh mode = 644, want 755 — confinedWrite hardcodes 0o644 (install.go:743) and drops the exec bit`
- AC-004: `journal on disk mid-run shows WriteCompleted=false for the already-written first file claude-agents/manager-x.md — flags are batched after the loop instead of persisted per file`
- AC-002: `the journal file staged at :239 lacks the recovered entry claude-skills/moai-beta/extra.md (present only after the post-loop carry-in) — an interruption before :301 drops it`
- AC-003: `captured: the carried entry hash 8f05c321...d2510 != the refreshed bytes hash 0f86c04a...d1823 — the :297 carry-in wrote the stale hash over a file the same run had already refreshed`
- AC-005: `LoadJournal silently decoded an unknown schema_version=99 journal (returned &{SchemaVersion:99 ...}) — no version gate exists`
- AC-012: `the migration report never names the unregistered mirror copy ".agents/skills/moai-unregistered/SKILL.md" — it is counted as untouched with no classification or visibility`
- AC-019a(위장): `a corrupt project manifest load produced CheckOK "no project manifest (nothing to compare)" — the load failure is disguised as nothing-to-compare (doctor_user_install.go:134-137)`
- AC-019a(보존): `the corrupt original was moved away from .../.moai/manifest.json during a read-only doctor check`
- AC-020: `a tree with zero lock-comparable files reported "project tree matches the lock file" — the vacuous comparison is disguised as a match`
- AC-021: `L4 failed although the USER workflow tree is healthy — the empty project workflows dir won the existence-only selection gate (:57). Message: "L1:PASS L2:FAIL L3:FAIL L4:FAIL L5:FAIL L6:PASS"`
- AC-022: `L1 reported "L1:PASS ... L4:PASS ..." with a broken project-side harness skill present — the wholesale skillsDir swap handed L1 the USER scope (doctor_harness.go:58), so the project defect is invisible`
- AC-014: `a user-installed skill with no project mirror resolved to outcome 1 (the project skill mirror .../.agents/skills does not exist ...) — the user hierarchy face is unreachable, the verb answers nothing-to-disable instead`
- AC-018: `the re-run after an ensure failure is refused with "initialization failed: project already initialized" — no resume path exists (init.go:905)`
- AC-015: `the same-name agent rendered 2 rows (want 1 consolidated row with scope provenance) — agentDirsFor concatenates both scopes without dedup`
- AC-009: `delete of a user-installed managed asset was allowed (decision "", want "deny") — command: rm -f ~/.claude/agents/moai/plan-auditor.md; the user install roots are not in the protected set` (4형태 전부 동일 관측)

**AC-023 커버리지 측정 (M0 baseline — 판정 아님)**

`go test -coverprofile` 절차(§C AC-023)로 측정: userassets 패키지(전체 스위트 결합)
81.3%, critical 경로 파일 수준 — journal.go LoadJournal 88.9%/WriteJournal 66.7%,
lock.go acquireUserLockStale 81.5%. web 0.4% / hook 1.5% / cli 6.4%는 M0 셀렉터 전용
실행의 baseline 수치다. 85%/90% 판정은 각 수리 마일스톤에서 §B 결합 스위트로 재측정한다.

**13a/13b 재현 판정 (M3 범위 게이트 — 카드 명시 사항)**

- **13a REPRODUCED**: Codex 루트(.agents/skills)에 설치된 스킬 사본이 `.claude/rules/moai/`·`CLAUDE.md` 참조를 그대로 운반한다. M3 스킬 축 착수 조건 충족.
- **13b REPRODUCED (AC-006 동일 결함 확정)**: 역할 TOML(:392 경로)이 소스와 byte-identical. acceptance.md §C AC-007의 흡수 조항대로 13b 판정은 AC-006에 흡수되며, 원장 :584 좌표는 표류 확정(실 경로 :392 — research.md §1 재고정과 일치).
- **미변환 참조 잔존 목록**: `.claude/rules/moai/`(→.moai/policies 매핑 존재, 설치 경로 무배선), `.claude/skills/...`(→.agents/skills 매핑 존재, 무배선), `CLAUDE.md`(→AGENTS.md 매핑 존재, 무배선), `.claude/agents/moai/`(매핑 자체 부재 — REQ-CNV-002 보고 대상). 결론: 변환기 매핑은 존재하나 설치 경로 배선이 전무 — M3의 배선 범위는 7a+13a+13b 전체로 확정.

**Gaps (미관측)**

- AC-001 windows 마커 팔: darwin에서 실행 불가(B1) — SKIP으로 기록. 런타임 RED 관측은 windows 실행 환경(AC-023의 windows 커버리지 수집과 동일 환경)에서 수행 필요. 소스 재독 확인(lock_guard_windows.go:23 O_EXCL·무소유권)과 GOOS 빌드 게이트 exit 0은 관측됨.
- AC-023의 85%/90% 판정: M0는 baseline 측정만 수행 — 판정은 수리 마일스톤의 §B 결합 스위트 재측정 소관.
- AC-015의 저장-재독록 2단계 팔: 렌더 팔 RED로 조기 종료 — arm 2는 M6 GREEN 전환 시 실행된다.
- web/hook/cli 전체 스위트: 레인-로컬 규율(전체 스위트 금지)에 따라 셀렉터만 실행 — 저장소 전체 판정은 CI 소관(보고 시점 PENDING).
- windows 유닛 매트릭스: 본 실행 미관측 — CI windows 잡 소관.

**Residual-risk (잔여 위험)**

- AC-011/AC-003의 재현은 watcher 폴링 probe다 — 결정화 설계(대용량 payload·판별자·재시도)에도 커널 타이밍에 의존하며, M5/M1 수리 후 GREEN 전환 검증 시 동일 probe의 재사용 가능성(같은 눈금자로 재측정)을 권장한다.
- AC-018의 ensure 실패 주입은 "이미 초기화된 프로젝트 + 파손 사용자 매니페스트" 상태를 만드는 간접 형태다 — run 1의 실패 원문("user-asset install failed")은 관측되어 올바른 상태를 증명하지만, 실행 중단(crash) 형태의 재개와는 다른 진입 경로일 수 있다.
- 신규 테스트 8개 파일은 프로덕션 무변경(B10)이며, 의도 RED 22건이 패키지 스위트를 적색으로 만든다 — 본 배터리는 관측 기록이며 수리 마일스톤(M1-M7)에서 전환된다.

### M1 — 저널 영속화·복구 무결성 (2026-10-09, HEAD feb6ec390 기반, 트리 WT-user-asset-bundle)

프로덕션 변경: `internal/userassets/install.go`(스테이징 병합·파일별 플래그 영속화·반입 해시
동기화·applyTarget 서명) + `journal.go`(버전 게이트) — design §2 불변 보존(journal clear는
manifest save 성공 뒤에만, B3 플래그는 현행 run의 stage 저널, RF8 토큰 계약, C2 confinedWrite
무변경).

**AC 전환 매트릭스 (RED→GREEN)**

| AC | 테스트 | M0 관측 | M1 관측 (verbatim tail) |
|---|---|---|---|
| AC-002 | TestJournalStageCarriesRecoveredEntries | RED | **GREEN** — 첫 스테이징이 회수 항목을 포함 (parked 관측) |
| AC-003 | TestJournalRefreshUpdatesCarriedHash | RED | **GREEN** — `captured — the carried entry carries the refreshed hash 0f86c04a...d1823` (강제 save 실패 후 디스크 저널에서) |
| AC-004 | TestWriteCompletedPersistedPerFile | RED | **GREEN** — 중단 시점 디스크 저널에 첫 파일 플래그 true |
| AC-005 | TestJournalUnknownSchemaRefused | RED | **GREEN** — `LoadJournal`가 schema_version=99를 진단과 함께 거절 |

회귀: userassets 전체 스위트에서 M1 4종 GREEN + 기존 테스트 전부 PASS + 미처리 RED 8건
(M2-M7 소관)만 잔존. cli 배터리 셀렉터 11종 재실행 — FAIL 8건은 M0 RED 그대로(변화 없음),
PASS 3건 유지. lint 0 issues, gofmt 청결.

**AC-023 커버리지 (신규 파일별 집계 파이프라인, unix 실행, 전체 스위트 결합)**

| 파일 | 문 커버리지 | 90% 게이트 |
|---|---|---|
| journal.go | 26/27 = **96.3%** | **PASS** — 미커버 1블록은 MarshalIndent 오류臂(구조적 도달 불가, Gaps 선언) |
| lock_guard_unix.go | 15/16 = **93.8%** | **PASS** |
| install.go | 299/364 = **82.1%** | FAIL — 미커버는 Codex 변환·divergence 백업·오류 臂 (M3/M5 착지 시 상승) |
| lock.go | 36/44 = **81.8%** | FAIL — stale 재클레임·오류 臂 (M2 착지 시 상승) |
| 패키지 합계 | **81.8%** | FAIL (85% 기준) — 위 두 파일의 상승분이 해소 |

판정: M1 시점 게이트는 journal.go·lock_guard_unix.go 통과, install.go·lock.go·패키지 합계는
미달 — 후속 마일스톤의 수리가 곧 커버리지 상승 수단이며, 최종 마일스톤에서 재판정한다.

**이상 항목**

- AC-003 관측 캐리어를 codex-agents TOML(단일 루트)로 재구성했다 — 최초의 moai-beta 스킬
  트리 캐리어는 미청구 쌍둥이 루트(agents-skills)의 stage 항목이 첫 스테이징부터 새 해시를
  싣워 watcher가 해시 동기화 전에 발화하는 결함이 있었다(실측). 단일 루트 캐리어로 제거.
  테스트 파일 코멘트에 기록.
- AC-003 테스트의 트리거는 M1 영속화 시점에 맞게 재조준되었다(M0는 post-loop 재영속
  :301을 감시 — 파일별 영속화가 그 자리를 대체). 단정 계약(중단 뒤 디스크 저널의 반입
  해시)은 동일하며, M0 RED 전문은 위 M0 절에 보존되어 있다.

## §E.3 Run-phase Audit-Ready Signal

_(pending run-phase — manager-develop 소관.)_

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase — manager-docs 소관. sync_commit_sha: )_

## §F Phase 4 Mode Selection

- 입력 파라미터: tier L · 범위 약 15-20 파일(구현+신규 재현 테스트) · 도메인 4(internal/userassets, internal/cli, internal/template, internal/web) · 언어 혼합 Go 100% · 병렬 이득 LOW(coding-heavy) · agent-team 전제 N/A(미요청).
- 모드 평가: direct 미해당(다중 파일·의미 변경) / fanout 미선택(coding-heavy — Anthropic coding-task 병렬성 유의) / sweep 미선택(기계 균일 변환 아님·파일 30 미만·상호 의존 마일스톤) / **serial 선택** / agent-team 미요청(실험적 명시 전용).
- **Decision: serial** — 마일스톤 M0→M7 순서에 1회 1개 manager-develop 스폰(트리 내 단일 작성자 — 팩토리 레인 one-writer 규율과 합치).
- 근거: Anthropic coding-task caveat(코딩 과업의 순차 기본) + 결정 가역성 순서(plan §F)가 마일스톤 간 의존을 만드는 구조. 팩토리 레인 크론+태스크 규율이 연속성 담당 — ac_converge goal 미무장(중복 감시자 방지).
- Boundary Case 해당 없음(기준 여유 있음).
