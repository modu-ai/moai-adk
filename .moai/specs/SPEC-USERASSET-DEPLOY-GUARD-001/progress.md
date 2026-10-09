# progress.md — SPEC-USERASSET-DEPLOY-GUARD-001

status: completed
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

### M1 후속 (게이트 14 버킷 B) + M0.1 테스트 수리 (게이트 13/14, 2026-10-09)

**버킷 B — M1 축 프로덕션 수리 (RED 관측: stash 롤백 verbatim — `the refusal does not
carry the schema diagnosis: pending-install journal corrupt — preserved at
...corrupt-20261008T193558`)**

| 항목 | 수리 | 검증 |
|---|---|---|
| B3 journal.go:101 | 스키마 불일치 저널은 corrupt 사이드카로 라우팅되지 않는다 — `JournalSchemaError` 타입 신설, Install이 **원본 경로 보존 + 재거절**. corrupt(파싱 실패)만 사이드카로 | TestJournalUnsupportedSchemaPreservedInPlace GREEN (수정 전 RED 관측: 사이드카 라우팅 + 진단 은닉) |
| B4 install.go:302 | refresh된 반입 항목이 해시와 **provenance(MoaiVersion·InstalledAt)를 함께 동기화** — REQ-006 (디스크 바이트를 만든 빌드가 기록된다) | AC-003 캡처 단정에 MoaiVersion 검증 추가 — GREEN |

**버킷 A — M0.1 테스트 수리 6건 (전부 적용, 영향 셀렉터 재실행)**

| 결함 | 수리 | 재관측 |
|---|---|---|
| A1 [P1] Mkfifo가 windows 컴파일 파괴 | `m0_fifo_{unix,windows}_test.go` 빌드태그 분할 — 런타임 skip이 아닌 컴파일타임 분리 | **`GOOS=windows GOARCH=amd64 go test -c ./internal/userassets` OK (P1 합격)**, hook 동일 OK, 전체 windows 빌드 OK |
| A2 watcher 조건 광역 일치 | 저널을 파싱해 recoveredKey 항목의 해시를 **특정 검사** | AC-003 GREEN 유지 |
| A3 watcher가 자기 shuttle을 제품 쓰기로 오판 | 센티널 판정을 **실 목적지명(file.txt) 한정** — .ua-write-* 셔틀은 probe 장치 | TOCTOU RED 재관측 (올바른 근거) |
| A4 decideBash가 Handle의 zone 셸 분기 우회 | **Handle 실제 진입** + overlay 보호 구성 + harness-learner 신원 + control 팔 | control(프로젝트측 covered 삭제)=**DENY 관측**(가드 생존 증명), 결함 팔(사용자측)=ALLOW → RED 성립. 실측 필드: `HookSpecificOutput.PermissionDecision` |
| A5 FIFO 설치 고루틴이 cleanup 생존 | RED 팔에서 write-end 개방·폐쇄로 EOF 주입 → 고루틴 종료 대기 후 종료 | TempDir cleanup 오류 소멸, teardown 보장 |
| A6 lock 경로 항법 동일성(tautology) | **획득이 실제 만든 파일**을 디스크 diff로 관측 — 실측 마커명 `user-assets.acquire-guard.guard` (게이트 지적명 `user-assets.lock.guard`와 다름 — 본 트리 실측값 기록) | PASS + 보유 중 2차 획득 거절(직렬화 실측) |
| A7 TOCTOU 무성공 경로 | 3-way 판정: 탈출=RED / 고정·안전거절=**GREEN 성공 경로** / 미관측=재시도 후 도구 소관 | HEAD에서 RED 유지 (M5 수리 시 GREEN 전환 가능) |

**버킷 C — 구조 처분 (레인 재정, §E.2 기록)**: 의도 RED 테스트는 빌드태그 격리 없이 기본
run에 남는다. 각 마일스톤이 자기 가족을 GREEN으로 전환하며, 미착지 브랜치에서의 중간
적색은 예상되고 설명된 상태다 — 영향 패밀리 판정이 레인의 실행 면이고, 저장소 전체
판정은 전 가족 GREEN 시점의 통합 CI가 소관이다.

**게이트**: lint 0 issues (userassets+hook), gofmt 청결, userassets 전체 스위트 = 의도
RED 8건(M2-M7 소관)만 잔존, M1 GREEN 4종+버킷 B 1종 유지.

### 게이트 15 — 쓰기 증폭 게이트 (design §2 제약, 2026-10-09)

- **install.go 플래그 영속화 게이트**: per-file WriteCompleted 영속화는 **실제 쓰기가
  발생한 대상에만** 발화한다(applyTarget의 `wrote` 신호 — confinedWrite 팔 한정).
  up-to-date 재실행은 대상당 저널 재직렬화를 지불하지 않는다(게이트 실측: 미변경 6파일이
  저널 쓰기를 2→7로 증폭). stateManifestStale(기록만 수리, 무기록)은 플래그 없음·해시
  동기화는 유지 — 디스크 바이트가 이미 shipped라 반입 항목과 일치해야 하기 때문.
- 검증: TestJournalFlagPersistOnlyForWrittenFiles GREEN — up-to-date 재실행에서
  "persist completion flag" 실패 0건 + installed=0/refreshed=0 확인. AC-004
  (TestWriteCompletedPersistedPerFile) GREEN 유지 — 실제 쓰기 파일의 플래그는 계속
  즉시 영속된다. 전체 스위트 의도 RED 8건 유지, lint 0, gofmt 청결.

### 게이트 17 — 버전 헤더 선검사·armed 순서 정정 (2026-10-09)

- **journal.go 버전 헤더 선검사**: schema_version 헤더를 본문 디코드 **전에** 검증한다.
  미래 스키마가 본문 필드 **타입**을 바꾸면 본문 디코드가 먼저 실패해 corrupt 사이드카로
  라우팅되어 원본 경로가 사라지는 결함(게이트 재현)을 차단 — 헤더 peek은 미지 필드를
  타입 검사 없이 건너뛰므로, 타입 충돌 본문도 원본 경로 보존+재거절로 간다. 헤더 자체가
  파싱 불가인 본문은 복원 가능한 버전이 없으므로 corrupt가 옳다. 검증:
  TestJournalFutureSchemaTypeConflictPreservedInPlace GREEN (entries를 문자열로 타입
  충돌시킨 schema_version=99 저널 → JournalSchemaError + 원본 경로 유지 + 사이드카
  부재). 수리 과정에서 Install의 스키마 거절 래핑에 %w 누락이 발견되어 함께 수리
  (errors.As 체인 복원).
- **TOCTOU probe armed 순서 정정**: 판정 순서를 escape → **!armed (probeLost — 재시도)**
  → pinned/refused (probeHeld) 로 재배치. 지연 watcher가 arm에 실패한 시도의 부모 내
  기록을 성공으로 grading하던 결함 제거 — 관측되지 않은 경쟁 창은 성공이 아니라
  재시도 대상이다. HEAD에서 RED 유지(탈출 관측), M5 수리 시 arm→pinned 경로로 GREEN.
- 게이트: lint 0, gofmt 청결, 전체 스위트 의도 RED 8건 유지, 스키마 가족 6종 GREEN.

### M2 — 잠금 가드 복구: proven-death reclaim (2026-10-09, HEAD c41e67817 기반)

프로덕션 변경: `internal/userassets/lock.go`(guardMarkerDead·reclaimGuardMarker —
lockOwnerGone 자세 재사용 + ClassifyGuardMarker·GuardMarkerPath 신설) +
`lock_guard_windows.go`(마커에 PID 기록 + 사망 입증 회수 루프) +
`internal/cli/doctor_lock_markers.go` 신설(User Lock 행 — 가시 회복 경로) + doctor 등록.
`lock_guard_unix.go` 무변경 — unix는 flock 기반으로 커널이 사망을 증명하며, windows 마커가
unix 자세로 수렴하는 것이 REQ-LOCK-002의 방향이다(design §3).

**AC 전환 (AC-001 + AC-024)**

| 팔 | M0 관측 | M2 관측 |
|---|---|---|
| AC-001 windows dead-owner (pid 기록) | SKIP (darwin 미실행) | 정책 표 GREEN — pid 기록+사망 확인 마커가 회수 가능 판정+실제 회수 (플랫폼 중립, darwin 판정 가능). windows 런타임 팔은 windows 실행 환경 소관 (darwin Gap 선언) |
| AC-001 windows ownerless (무 pid 기록) | SKIP | 계약 전환: **자동 회수 금지** — 정책 표 GREEN (ownerless≠사망, 재시도도 거부) + windows 런타임 팔은 거부·마커 생존 단정 (skip, Gap) |
| AC-001 unix orphan guard | born-green | GREEN 유지 |
| AC-001 case 2 (무 pid 마커 — 연령 회수 금지) | born-green (거부 팔) | 거부 팔 GREEN 유지 + **보고·명시 제거 팔 착지**: doctor User Lock 행 (ownerless → never-auto-reclaimed 명시 + rm 절차 + report-only 단정), dead-owner → 자동 회수 안내, alive → 보유 경고 |
| AC-024 windows 빌드 | exit 0 | **exit 0 유지 + `GOOS=windows go test -c ./internal/userassets` OK + windows vet OK** |

**정책 표 (마커 상태 × 소유자 생존 — design §3, 플랫폼 중립)**: pid+사망=회수 가능·실제
회수 / ownerless=회수 불가·파일 생존 / 생존 pid=회수 불가 / 기형 기록(pid=invalid·0·-1·
4294967296)=실패 폐쇄 / 판독 불가(디렉터리)=실패 폐쇄. reclaimGuardMarker는 자체 재입증
(fail closed) — 미입증 마커를 옮기는 호출 자체를 불가능하게 한다. 실측 dead pid는
helper-process 재확인 방식(헬퍼 종료 후 pid). .lock 파일 자체의 stale 인수 경로
(TestLockStaleDeadOwnerTakenOver)도 착지 — 인수가 신규 소유자로 재기록하는 것까지 단정.

**AC-023 커버리지 (unix, 전체 스위트 결합)**

| 파일 | 문 커버리지 | 90% 게이트 |
|---|---|---|
| lock.go | 65/73 = **89.0%** (81.8%→89.0%) | 미달 1.0%p — 잔여 8문은 환경적 오류 臂 (mkdir 실패·deadline·release 오류·rename 경합 패배 — Gaps 선언) |
| lock_guard_unix.go | 15/16 = **93.8%** | PASS |
| journal.go | 30/32 = **93.8%** | PASS |
| install.go | 308/373 = **82.6%** | FAIL — M3/M5 착지가 상승 수단 |
| 패키지 합계 | **82.7%** | FAIL (85%) — 위 상승분이 해소 |

**Gaps**: lock_guard_windows.go의 런타임 커버리지는 windows 실행 환경 소관 (darwin에서
IgnoredGoFiles — AC-023의 windows 판정 절차에 속함). lock.go의 환경적 오류 臂 8문.

**게이트**: lint 0 (userassets+cli), gofmt 청결, windows build+test compile+vet OK,
전체 스위트 의도 RED 8건(M2 제외 — 이 마일스톤 RED는 darwin-skip 팔) 유지, doctor
가족 3종 GREEN, 정책 표 7종 GREEN.

### 게이트 18/19 — M2 정밀화 9건 (P1 회수 신원 경합 포함, 2026-10-09)

| # | 수리 | 검증 |
|---|---|---|
| 18-1 [P1] | **회수 신원 경합**: reclaimGuardMarker가 rename 전후 바이트 동일성을 검증한다 — 예비 판독(사망 입증)과 이후 rename 사이에 경합 승자의 새 마커가 끼어들면 displaced 바이트가 불일치하고 **원위 복원** 후 실패 반환. windows 마커 기록은 나노초 타임스탬프로 바이트 유일화. (미세 노출: 복원 창의 과도 displacement는 남는다 — 삭제가 아니라 자기치유 displacement이며, 제2 기록자 재개창은 닫힌다) | TestGuardMarkerReclaimPolicy/reclaim_identity_survives_concurrent_acquirers — 200라운드 경합 후 정지 상태 안전 속성 (dead 바이트 잔존 없음·외부 바이트 없음·소실 없음) GREEN |
| 18-2 | **소유-vs-미기록 구분**: reconcile 사례 3에서 플래그 없어도 매니페스트 추적 대상은 REQ-023 divergence(백업+보존+보고)로 분류 — "이번 런 미기록"을 "무소유"로 읽던 결함 수리 | TestJournalTrackedMismatchIsDivergenceNotCollision GREEN — divergence 분류+백업 존재+사용자 바이트 불변 |
| 18-3/19-2 | **unix flock 잔여물 오분류 제거**: pid-less 잔여 분류를 플랫폼 잔여 함수로 분리 — unix는 flock 상태가 진실(자유=clean leftover→absent, 점유=alive), windows는 존재=점유 증거(무 pid=ownerless) | TestUnixFlockLeftoverClassification + TestDoctorUserLockMarkersUnixFlockLeftoverIsOK GREEN |
| 18-4/19-3 | **IsNotExist-only 부재 + Lstat 타입 판정**: marker 경로의 비정규 객체(FIFO 포함)는 절대 읽지 않고 Irregular 경고로 노출 — doctor 행이 FIFO에서 hang하지 않는다 (이 테스트가 반환하는 것 자체가 증명) | TestGuardMarkerReclaimPolicy/irregular + TestDoctorUserLockMarkersIrregularIsSurfaced GREEN; lockOwnerGone도 동일 Lstat 폐쇄 |
| 19-4 | **회수 조건 안내 정합**: doctor의 .lock dead-owner 안내가 stale 창(DefaultStaleAfter) 조건을 명시 — guard marker(무 연령 게이트)와 .lock(연령+사망)의 실제 조건 구분 | TestDoctorUserLockMarkersDeadOwnerIsInformational — "stale window" 문구 단정 |
| 18-5 | **frozen-guard 재현 배선(M4 준비)**: 결함 팔이 temp home의 절대 경로로 사용자 twin을 겨냥 + 사용자 매니페스트가 대상을 추적(관리 파일 증거 — REQ-GRD-002 한정의 자격 증명). control 팔(프로젝트측 DENY) 유지 | RED 유지 — 올바른 근거 (HEAD에서 사용자 루트 미보호), M4가 user-root 해석을 착지하면 flip |

**커버리지 변동**: lock.go 89.0%→**88.2%** (신설 플랫폼 잔여·Irregular·identity 분기로
분모 증가 — 잔여 미커버는 환경적 오류 臂), lock_guard_unix.go 93.8%→**92.3%** (flock
probe 신설 분기), 패키지 82.7%→**83.2%**. 게이트: lint 0, gofmt 청결, windows build +
test compile OK, 의도 RED 8건 유지.

### M3 — Codex 정규화 배선 + 게이트 20 (7a+13a+13b 흡수, 2026-10-09)

프로덕션 변경: `internal/userassets/install.go`(installTarget에 codexFace·data 단일
판독 필드 — targetBytes가 Codex 면에 NormalizeCodexRoleForDeploy 적용, applyTarget이
동일 바이트 기록, REQ-CNV-002 스캔, afterTargetPersist 시음) + `Result.Unconverted`
신설 + `remove.go` readShippedForKey의 Codex 면 백업 바이트 동일 변환 +
`internal/cli/user_asset_phase.go` 요약 렌더 + `lock.go` ClassifyLockFile 분리(게이트 20).

**AC 전환 (AC-006 + AC-007 13a/13b + AC-026) — RED→GREEN 4종**

| AC | 테스트 | M0/M3 | M3 관측 |
|---|---|---|---|
| AC-006 (7a) | TestCodexAgentTOMLReferencesConverted | RED | **GREEN** — 설치 TOML이 변환 면(.moai/policies·.moai/workflows·AGENTS.md)으로 착지, 무변환 참조 0 |
| AC-007 13a | TestCodexUserSkillInstallNoClaudeOnlyRefs | RED (재현 확정) | **GREEN** — Codex 루트 스킬 사본 무변환 참조 0 |
| AC-007 13b | TestCodexRoleTOMLNotVerbatim | RED (AC-006 흡수 확정) | **GREEN** — 설치 TOML ≠ 소스 byte-identical |
| AC-026 | TestCodexUnconvertibleReferenceReported | RED | **GREEN** — 변환 후에도 `.claude/` 참조 잔존 파일이 Result.Unconverted에 파일 단위 나열 (신규 보고 면 — M0 테스트 코멘트가 예고한 flip) |

출처 규율(REQ-CNV-001): installTarget.data = **단일 판독의 변환 바이트** — 기록 바이트,
manifest sha, journal 해시가 전부 같은 바이트에서 산출된다(원문 해시+변환 기록 혼합이
구조적으로 불가능). Claude 면은 verbatim 유지(참조가 옳은 면) — 기존 verbatim 단정
테스트(TestInstallFirstRunLandsL0)가 무참조 fixture에서 계속 GREEN으로 이를 지킨다.
백업 바이트(readShippedForKey)도 Codex 면 동일 변환으로 정렬.

**M0 시음 이관 (단일 판독 설계의 부수 효과)**: applyTarget이 소스를 재판독하지 않게
되어 M0의 read-counting park 지점이 소멸 — 중단 재현 시음을 Installer.afterTargetPersist
(정렬 순서 N번째 기록 target의 플래그 영속 직후, 프로덕션 nil)으로 이관하고 두 저널
테스트를 재조준. 단정 계약 동일, 전문은 M0/M1 절 보존.

**게이트 20**

| # | 수리 | 검증 |
|---|---|---|
| 20-1 [P1] | flock 테스트 본문의 빌드태그 분리 — 이미 f80590426에 착지됨(재검증: m2_lock_guard_unix_test.go만 syscall.Flock 보유, `GOOS=windows go test -c` OK) | 현재 HEAD 실측 OK |
| 20-2 | **.lock 파일의 flock 오분류 제거**: ClassifyGuardMarker(가드 파일 — unix flock 진실)와 ClassifyLockFile(.lock — O_EXCL 기반, 무 flock — pid-less 잔여는 전 플랫폼 ownerless=수동 복구 상태)로 분리, doctor가 파일 종류별로 선택 | TestLockFilePidlessLeftoverIsOwnerlessNotAbsent GREEN — 분류=ownerless + 획득=ErrLocked 일치 (게이트 재현 상태) |

**라운드 19 잔여 A-D — 이미 6627769b5에 착지 (재검증)**: (A) Lstat 선입검사 lock.go:119/
166/228 실측, (B) flock 상태 판정 classifyPidlessMarker + TestDoctorUserLockMarkers
UnixFlockLeftoverIsOK, (C) IsNotExist-only + Irregular 노출, (D) stale window 안내.
리더 실측은 6627769b5 이전 트리 기준이었음으로 재보고.

**AC-023 커버리지 (unix, 전체 스위트 결합)**

| 파일 | 문 커버리지 | 게이트 |
|---|---|---|
| install.go | 322/391 = **82.4%** | 미달 — M3 전환 분기는 커버되나 충돌 패밀리(FIFO/TOCTOU/exec-bit) applyTarget 팔이 여전히 미실행: **M5 착지가 계획된 riser** |
| lock.go | 84/96 = **87.5%** | ClassifyLockFile 분리로 분모 증가 |
| journal.go 93.8% / lock_guard_unix.go 92.3% | | PASS |
| 패키지 합계 | **82.9%** | FAIL (85%) — M4-M7 착지분이 해소 |

**게이트**: lint 0, gofmt 청결, windows build+test compile OK, userassets 잔여 RED 4건
(M5/M6 소관), cli 배터리 의도 RED 8건 변화 없음, 저널 가족 6종 + Codex 가족 4종 GREEN.

### 게이트 21/22 — 해시 판별자 정정 + 컴파일 클래스 4번째 인스턴스 (2026-10-09)

| # | 수리 | 검증 |
|---|---|---|
| 21(a) | **컴파일 클래스 4번째**: cli m2_doctor_lock_markers_test.go의 syscall.Mkfifo를 `//go:build unix` 파일(m2_doctor_lock_markers_unix_test.go)로 이동 — 런타임 가드는 컴파일을 못 막는 정정. **상설 규칙 선포: unix 전용 syscall(Mkfifo·Flock 등)은 첫 작성부터 빌드태그 파일+windows 스텁으로 — M4-M7 테스트 전부 적용** | `GOOS=windows go test -c ./internal/cli` OK (userassets도 OK) |
| 21(b)/22-1 [P1] | **중단 갱신 오판 정정 — 해시 판별자**: case 3의 분류기를 `tracked` 플래그만 쓰던 형태에서 **해시 비교 판별자**로 정정 — (i) 디스크 바이트 == 기존 manifest 해시 = 미완료 pending refresh → 무분류로 일반 refresh 경로 위임 (applyTarget이 stateManifestMatch로 v2 완성; manifest를 여기서 건드리지 않음), (ii) 바이트 변경 = 사용자 수정 → divergence, 단 **이번 런이 쓰지 않은 파일의 기존 기록은 verbatim 보존** (저널 해시는 플래그 완결 런 자기 설치에만 권한). v2 해시를 v1 바이트에 못박아 갱신이 영구 막히던 결함(게이트 재현: 2차 재시도까지 v1 고착) 차단 | TestJournalIncompletePendingRefreshCompletes GREEN — 재시도가 v2로 refresh + divergence/collision 0 + manifest=v2 + 3차 런 멱등(고착 없음); 게이트 18 divergence 테스트도 GREEN 유지 |
| 22-2 | 저널 테스트 중단점 — **이미 d90df83f0에 착지** (단일 판독 설계의 read-park 소멸 → afterTargetPersist 시음 이관, 두 테스트 재조준). 게이트 실측은 커밋 중간 트리 기준 | 현재 HEAD에서 두 테스트 GREEN (전체 스위트 관측) |

**게이트**: lint 0, gofmt 청결, windows build + userassets/cli test compile OK, 의도 RED
세트 불변 (userassets 4건 = M5/M6, cli 8건 = M6/M7).

### M4 — 동결 가드 사용자 루트 보호 (원장 9a, REQ-GRD-001/002, 2026-10-09)

프로덕션 변경: `internal/config/protected_zone.go`(**ZoneUserRoot** 진입 종류 신설 —
`user-root:<slug>/…` 파싱, Sub 접미 의미, Match의 namespaced 형태 판정; 절대 경로·..
·대문자 슬러그는 프리픽스 유무와 무관하게 기각 — repository-relative 계약 존중) +
`internal/hook/user_root_zone.go` 신설(슬러그→디렉터리 표, userRootForms namespaced
형태 산출, userRootFormTracked 매니페스트 추적 한정) + `protected_zone_path.go`
(resolveZoneTarget의 user-root 팔) + shell/guard 두 deny 경로의 **추적 한정 필터**
(REQ-GRD-002 과잉 보호 방지 — 무추적 파일은 매니페스트 디렉터리 안에서도 편집 가능).
import 방향: hook ← config만 (userassets는 테스트 전용 패리티 임포트).

**AC-009 전환 — RED→GREEN (이중 형태 probe, 실제 Handle 진입)**

| 팔 | M0/M4-전 관측 | M4 관측 |
|---|---|---|
| control (프로젝트측 covered 삭제) | DENY (게이트 18 배선) | **DENY 유지** — 가드·구성 생존 증명 |
| (i) 절대 항목 기각 | config 계약 | **user-root:/etc/passwd·C:/·..·대문자 슬러그 전부 기각** (파서 테스트 GREEN) |
| (ii) 사용자 파일 매치 → deny | **ALLOW (RED — 원장 9a)** | **DENY로 전환** — 절대 경로·~ alias 양형 모두, 매니페스트 추적 파일 한정 |
| REQ-GRD-002 과잉 보호 방지 | — | 무추적 사용자 파일 = **ALLOW** (매니페스트 디렉터리 안에서도) — 보호는 관리 자산에 한정 |

**REQ-GRD-001 정직성 갱신**: "집합 불포함" 주장이 relayed였던 것이 AC-009의 실제 적재
RED-now(관측: ALLOW, HEAD 6627769b5 — 게이트 18-5 배선분)과 M4 전환 GREEN(299bc6bd7
이후)으로 관측 완료 — 원장 항목 9a의 소관 판정 근거가 되었다.

**라운드 22 #1 정정**: 해시 판별자(미완료 pending refresh → refresh 경로 위임)는
`299bc6bd7`에 이미 착지 — TestJournalIncompletePendingRefreshCompletes(재시도 v2
완료·3차 런 멱등)가 정확히 그 재현이며 현재 HEAD에서 GREEN. 리더 실측은 커밋 이전
트리 기준이었음으로 재보고.

**패리티 (REQ-GRD-002)**: hook의 슬러그→디렉터리 표 4종이 userassets.ResolveRoots와
전부 일치 (테스트 전용 임포트 패리티 테스트 — production 의존 방향 불변). 드리프트 시
적색.

**테스트**: config 파서 3종(kinds·rejections·match) + hook 4종(parity·forms·tracked
scoping·AC-009 이중 형태) GREEN. **게이트**: lint 0, gofmt 청결, windows build +
4패키지 test compile OK, hook/config/userassets 전체 스위트 GREEN (userassets 의도
RED 4건 = M5/M6 소관).

### 게이트 23-26 — 잠금 판독·doctor 인용·저널 주입·user-root 경계 (2026-10-09)

| # | 수리 | 검증 |
|---|---|---|
| 23-1 [P1] | **readLockRecord 플랫폼 분할**: unix는 O_NONBLOCK 오픈 → **핸들 fstat** → 핸들 read — Lstat→read 스왑 창(path가 FIFO로 교체되면 read가 hang)을 fd 레벨 봉쇄. windows는 Lstat+ReadFile (FIFO 경로 위험 없음) | FIFO 마커 → Irregular 판정, hang 없음 (정책 표 irregular 팔 GREEN) |
| 23-2 | doctor 제거 절차 **POSIX 단일 인용** (`rm '<path>'`, 내부 `'\''` 이스케이프) — 이중 인용의 `$(...)` 실행 재현 차단 | TestDoctorUserLockMarkersReportsOwnerless — `rm '` 포함 + `rm "` 부재 단정 |
| 24-2 | 저널 테스트 실패 주입 3건을 **플랫폼 독립 디렉터리-스wap**으로 교체 (경로를 디렉터리가 점유 → 모든 OS rename 실패; windows 무효 dir-Chmod 제거) | 저널 가족 전부 GREEN (windows 포함) |
| 25-2 [P1] | **템플릿 source에 user-root 4항목 착지** (moai_managed/runtime_paths, 뉴트럴 코멘트) + runtime allowlist 승인 + **dogfood 매니페스트 byte-identical 동기화** | TestProtectedZone 3subtest GREEN |
| 26-1 [P1] | user-root 매치를 **resolved 대상**으로 판정 (zoneResolve + home resolve — macOS /var→/private/var 우회 봉쇄); /./·/../ 표기 수렴 단정 | TestUserRootFormsEmitNamespacedForm GREEN |
| 26-2 [P1] | **디렉터리 포함 검사**: 정확 매치 + 접두 컨테인먼트 (`rm -rf` 부모 변형 커버) | TestUserRootFormTrackedScoping (dir/root containment) GREEN |
| 26-3 [P2] | user-root 형태를 **독립 분기**로 분리 — 베이스라인·프로젝트 규칙이 user-root 형태를 못 보고, 프로젝트 형태도 ZoneUserRoot와 매치 불가 (무추적 basename 차단 누수 제거) | hook 전체 스위트 GREEN |

### 게이트 27-30 — user-root 경계 6건 + 매니페스트 FIFO (2026-10-09)

| # | 수리 | 검증 |
|---|---|---|
| 27-1 | 배포 yaml의 4개 user-root 항목 — 25-2로 착지 (현재 HEAD 재검증) | TestProtectedZone GREEN |
| 27-2 [P1] | **루트 자체 보호**: rest=="" naked-root 형태 발행 + 디렉터리 포함 검사로 tracked 하위 커버 (`rm -rf ~/.claude/skills` 차단) | TestUserRootFormTrackedScoping (dir/root containment) GREEN |
| 27-3 [P1] | **케이스 별칭**: 형태 키와 매니페스트 키 **양쪽 fold** 후 비교 — case-insensitive FS의 skill.md 별칭 봉쇄 | 동일 테스트 GREEN |
| 27-4 [P1] | **설치 루트 자체 zoneResolve**: dotfile-manager 심링크(~/.claude → 외부) 재배치 대응 | AC-009 probe GREEN (심링크 홈) |
| 27-5 [P1] | **디렉터리 후보 슬래시 없음**: ZoneDir/Prefix 매치가 rest와 rest+"/" 양쪽 판정 (shell 판정과 동일) | config Match 테스트 |
| 27-6 | 패리티 dot-path 테스트 — 원형 결합으로 수정 (Join이 .·..을 정규화하는 버그) | GREEN |
| 29-NEW-1 [P1] | 템플릿 코멘트의 SPEC 식별자 제거 (뉴트럴 재작성 — C1-spec-id-prefix 차단) | TestProtectedZone/Neutrality GREEN |
| 29-NEW-2 | **백슬래시 보존**: userRootZoneForms가 zoneSlash 사전 변환 없이 **원형 후보를 zoneResolve** (POSIX에서 \은 파일명 문자 — `alias\dir` 심링크가 잘못된 경로를 검사하던 누수) | 코드 정정 |
| 30 | **사용자 매니페스트 FIFO**: userRootTracksAny 판독을 비차단 오픈+핸들 fstat으로 (zone_user_read_{unix,windows}.go 신설 — 게이트 23 패턴) | FIFO 매니페스트 → untracked 판정, hang 없음 |

### M5 — 충돌 안전 (원장 8c/12/6b, REQ-COL-001/002/003, 2026-10-09)

프로덕션 변경: `install.go` — (a) RF5 충돌 사전 판정 + reconcile 판독의 **Lstat 선행
타입 판정** (비정규 대상은 읽지 않고 collision 분류), (b) **confinedWrite 핸들 고정** —
검증된 부모를 `os.OpenRoot`로 핀하고 temp-create·write·chmod·rename을 전부 핸들 경유로
(경로 재해석 제거 — C2 선언 경합 봉쇄; 기존 C2 검사 전부 존속, 핀은 rename의 경로
재해석을 대체할 뿐 약화가 아님), (c) `installMode(rel)` 신설 (.sh → 0755, 나머지 0644).
**M1 상호작용 검증**: 플래그 영속은 실쓰기 대상 한정(게이트 15 게이트) 불변 — 저널 가족
GREEN 유지.

**AC 전환 — RED→GREEN 3종 (잔여 RED 4건 → 1건)**

| AC | 테스트 | M0 관측 | M5 관측 |
|---|---|---|---|
| AC-010 (8c) | TestCollisionPrecheckSkipsFifoWithoutBlock | RED — FIFO에서 10s+ 블록 | **GREEN (0.01s)** — Lstat가 FIFO를 읽지 않고 collision-skip |
| AC-011 (12) | TestConfinedWritePinnedToValidatedParent | RED — 외부 센티널 탈출 | **GREEN (probeHeld)** — os.Root 핀이 검증 inode에 고정, 스왑 후 rename도 핸들 경유 |
| AC-025 (6b) | TestConfinedWritePreservesExecBit | RED — .sh 0644 하락 | **GREEN** — 설치 .sh 0755 (installMode) |

3-way probe 판정 유지: 탈출=RED / 고정·안전거절=GREEN / 미관측=재시도. 핀은 재검증을
**대체**하는 것이 아니라 rename 경로를 핸들로 바꾼 것 — C2 검사 전부 존속.

**AC-023 커버리지 (unix, 전체 스위트 결합)**

| 파일 | 문 커버리지 | 게이트 |
|---|---|---|
| install.go | 341/414 = **82.4%** | 미달 — M5 팔(선입검사·핀·모드)은 실행되나 신설 가드문으로 분모 증가; 잔여 미커버는 M6 번들·prune 상호 팔 |
| lock.go | 83/101 = **82.2%** | ClassifyGuardMarker 재배선 분모 증가 |
| lock_guard_unix.go | 36/40 = **90.0%** | **PASS** (readLockRecord 포함) |
| journal.go | 30/32 = **93.8%** | PASS |
| 패키지 합계 | **82.5%** | FAIL (85%) — M6/M7 착지분이 해소 |

**게이트**: lint 0, gofmt 청결, windows build + 5패키지(userassets/hook/config/cli/template)
test compile OK, hook/config/template 전체 GREEN (hook 305s), cli DoctorUserLock GREEN,
userassets 잔여 RED 1건 (AC-016 번들 클로저 — M6 소관).

### M6 — 배포·번들 표면 (원장 4/5-잔여/8a/8b/8d/11a/11b + 부채 R4, 2026-10-09)

프로덕션 변경: `migrate_project_assets.go`(무등록 미러 사본 파일 단위 분류·보고) +
`install.go` collectEntries(**DependsOn 클로저 전개** — 순환 안전 gathered 가드) +
`codex_skills_disable.go`(**프로젝트 미러 부재 시 사용자 설치 면 폴백** — userassets
설치 면 2종) + `web/agentfm.go`(**동명 행 통합** — 다중 디렉터리 스캔의 동명 행을 첫
히트 하나로; 이름 기반 폼 키의 이중 제출 덮어쓰기 제거) + `init.go`/**init_resume.go**
신설(**init 재개** — ensure 실패 후 재실행이 "already initialized" 대신 ensure 부족분
+ 미실행 후속 단계를 완료).

**AC 전환 — RED→GREEN 5종 + 확인 2종 (userassets 잔여 RED 4건 → 0건)**

| AC | 테스트 | M0 관측 | M6 관측 |
|---|---|---|---|
| AC-012 (4) | TestMigrationClassifiesUnregisteredMirrorCopy | RED — 무보고 방치 | **GREEN** — 무등록 미러 사본이 파일 단위로 분류·보고 (kept — provenance/approval 명시) |
| AC-013 (5-잔여) | TestDoctorAgentEmissionUncomparedNotOK | GREEN (회귀 가드) | **GREEN 유지** — CheckFail 불변 확인 (변경 없음) |
| AC-014 (8d) | TestSkillsDisableResolvesUserInstalledSkill | RED — MirrorAbsent 무시 | **GREEN** — 프로젝트 미러 부재 시 사용자 설치 면(~/.claude/skills 등)으로 resolve |
| AC-015 (8a) | TestAgentFormSameNameConsolidatedEndToEnd | RED — 2행 중복 렌더 | **GREEN** — 동명 행 1행 통합 + 2단계 저장-재독록에서 저장소가 항상 마지막 제출과 일치 |
| AC-016 (8b) | TestBundleDependsOnClosureInstall | RED — 클로저 부재 | **GREEN** — depends_on 순환 안전 전개로 의존 번들 동시 설치 |
| AC-017 (11a) | TestUpdateCancelKeepsProjectAssetsIntact | GREEN (회귀 가드) | **GREEN 유지** — 취소 불변 확인만 (재수리 없음) |
| AC-018 (11b+R4) | TestInitResumeAfterUserAssetEnsureFailure | RED — already initialized 거절 | **GREEN + DEBT R4 폐쇄** — 재개가 ensure 부족분을 완료하고 후속 단계 실행; 단정을 존재성에서 **내용으로 강화** (llm.yaml harness 라인 + .mcp.json moai 엔트리) |

**DEBT R4 폐쇄 (AC-018 내용 단정)**: ApplyHarness의 실제 산출물은 llm.yaml의 harness
라인 재작성이고 MCP 등록의 산출물은 .mcp.json의 moai 엔트리 — 존재성 단정을 이 두
내용 단정으로 교체. (`.moai/harness/`는 런타임 아티팩트로 init 산출이 아니어서 단정
대상에서 제외 — 기존 단정의 오류 정정.)

**init 재개 설계**: `init_resume.go` 신설 — resumeInitializedProject가 (1) 파손된
사용자 매니페스트를 quarantine (timestamped .corrupt-* — doctor의 rebuild-from-fresh-
init 회복 경로), (2) ensureUserAssetsLocked로 부족분 완성, (3) ApplyHarness → MCP →
Codex 배선을 생산 순서대로 실행. runInit의 already-initialized 분기가 이 재개를
호출하고 성공 시 nil 반환.

**AC-023 커버리지 (unix, 전체 스위트 결합)**

| 파일 | 문 커버리지 | 게이트 |
|---|---|---|
| install.go | 349/423 = **82.5%** | 미달 — 클로저 전개 팔 실행, 신설 분기로 분모 증가; 잔여 미커버는 M6/M7 표면의 나머지 오류·경합 팔 |
| journal.go **93.8%** / lock_guard_unix.go **90.0%** | | PASS |
| 패키지 합계 | **82.5%** | FAIL (85%) — M7 + 나머지 오류 팔이 해소 |

**게이트**: lint 0, gofmt 청결, windows build + 6패키지 test compile OK, web 에이전트
패밀리 GREEN, userassets **RED 0건**.

### M7 — doctor 진단 무결성 (원장 2/7b/9b, REQ-DOC-001/002/003/004, 2026-10-09)

프로덕션 변경: `doctor_user_install.go` checkProjectVsLock(**선-판독 + 파손 실패 등급** —
mgr.Load 도달 전 ReadFile+JSON probe로 파손을 판정해 quarantine rename을 유발하지 않음;
파손 = CheckFail, 무설치 = CheckOK 유지, 판독 불가 = Warn) + **vacuous 정직 보고**(비교
가능 면 0 = Warn "the lock comparison is vacuous") + `doctor_harness.go`(**필수 파일
존재 기반 workflow root 선택** — 4개 필수 파일+import 미보유 프로젝트 dir는 선택 패배,
**workflowsDir만 교체** — skillsDir는 프로젝트 스코프 유지) + `init_resume.go`(
**quarantine 판독을 비차단 오픈+핸들 fstat으로** — 게이트 36, FIFO 매니페스트 hang 제거).

**AC 전환 — RED→GREEN 5종 (구조적 적색 시대 종료)**

| AC | 테스트 | M0 관측 | M7 관측 |
|---|---|---|---|
| AC-019a (2) | TestDoctorProjectManifestLoadFailureNotDisguised | RED — CheckOK 위장 | **GREEN** — 파손 = CheckFail (mgr.Load 미도달, quarantine 유발 없음) |
| AC-019a 보존 | TestDoctorProjectManifestPreservedOnLoadFailure | RED — 원본 이탈+.corrupt 파괴 | **GREEN** — 원본 바이트·경로 유지 + 기존 .corrupt 미파괴 (선-판독으로 rename 자체가 미발생) |
| AC-019b (2a) | TestDoctorUserInstallHonestFailureAndReadOnly | GREEN (회귀 가드) | **GREEN 유지** |
| AC-020 (2b) | TestDoctorLockCheckVacuousNotReportedMatch | RED — 공허 OK 위장 | **GREEN** — 비교 가능 면 0 = Warn "vacuous" 보고 |
| AC-021 (9b) | TestDoctorWorkflowRootSelectedByRequiredFile | RED — 빈 dir가 선택 | **GREEN** — 필수 파일+import 보유 루트가 선택 |
| AC-022 (7b) | TestDoctorFallbackKeepsProjectScopeL1 | RED — 통째 교체 | **GREEN** — workflowsDir만 교체, L1:FAIL(프로젝트 결함 가시)+L4:PASS 동시 달성 |
| 게이트 36 | init_resume quarantine 판독 | — | FIFO 매니페스트 → 비차단 판독으로 hang 제거 (플랫폼 분할 readManifestRecord) |

**라운드 35 잔여 확인**: (i) 재개 게이트 — 35-3으로 착지(299bc6bd7 이후 f691c077e):
initResumeCheckpoint(pending 저널 or 파손 매니페스트) + TestInitResumeGateKeepsHealthy
Redirect(건전한 재실행 = 기존 redirect 유지 단정). (ii) autonomy tier — 35-6으로 착지:
resumeInitializedProject가 applyAutonomyTierBundleFn 포함.

**AC-023 커버리지 (unix, 전체 스위트 결합 — 최종 보고)**

| 파일 | 문 커버리지 | 게이트 |
|---|---|---|
| userassets 패키지 | journal **93.8%** / lock_guard_unix **90.0%** / install.go **82.6%** / lock.go **82.2%** / **합계 82.5%** | critical 2파일 PASS; 나머지 미달분은 M6/M7 표면의 오류·경합 팔 + M4 이후 신설 가드문 분모 |
| cli 패키지 (전체) | **41.9%** — checkProjectVsLock 58.3% (선-판독·vacuous·파손 분기 포함), checkPluginMigrationAdvisory 100% | cli 전체는 본 SPEC 이전부터의 대형 미커버 표면 다수 — 본 SPEC의 doctor 가족 분기는 전부 실행 |

**게이트**: lint 0, gofmt 청결, windows build + 6패키지 test compile OK. cli 전체 스위트의
잔여 FAIL 17건은 **문서화된 구조적 상태**(카드 워크트리 환경 계열 + binary-lag 계열) —
M7 이전 커밋(152c3ef7c)에서 동일 실패 재현으로 기존 환경 실패 확인 (M6/M7 도입분 아님).
영향 패밀리(doctor+resume)는 전부 GREEN — 구조적 적색 시대 종료.

### M7.1 — 게이트 42 마감 + 커버리지 보강 (2026-10-09)

**게이트 42 [P3] — pinned temp O_EXCL 변수 결함 수리.** install.go confinedWrite의
O_EXCL 재시도 루프가 스테일 외부 `err`(항상 nil)을 판정해 첫 EEXIST 충돌이
재시도 없이 설치 실패로 귀결될 수 있었다. 수리: 성공은 `createErr == nil`,
비재시도 오류는 `!errors.Is(createErr, os.ErrExist)`, EEXIST는 새 이름 재시도.
계약 검증은 경합 형태: `TestConfinedWriteExclusiveCreationSurvivesContention`
— 같은 부모에 24개 동시 confinedWrite, 전부 성공·전부 착지, `-count=2` GREEN.

**커버리지 보강 (게이트 42 후속, 리더 M7.1 지시).**
- `m7_coverage_topup_test.go` (플랫폼 무관): 잠금 가족 결정적 팔 5종(nil Release·
  회수됨 Release 스킵·String unknown 폴백·ClassifyLockFile absent/ownerless/
  irregular/owner-dead/owner-alive), 획득 오류 2종(잠금 홈 파일 점유 MkdirAll
  실패·잠금 경로 디렉터 점유 비-EEXIST), confinedWrite 거절 8종(무효 rel·부모
  부재·부모 심볼릭 링크 탈출·리프 심링크·디렉터 점유 목적지 rename 실패·파일
  점유 세그먼트·세그먼트 심링크 탈출/합법), Install 진입 5종(nil catalog·미지
  번들·root 해석 실패·schema-99 저널 현장 보존 거부·저널 선택 복원), now 폴백,
  오염 소스 3종(codex TOML 부재·소스 바이트 부재·무효 엔트리명).
- `m7_coverage_topup_unix_test.go` (unix 전용 — M0 상위 규칙대로 첫 작성부터
  빌드 태그): read-only 부모의 비-EEXIST create 오류 서면화(루트 스킵 가드),
  guard flock 경합 타임아웃, read-only .moai의 stage-journal fail-closed 거부.

**AC-023 커버리지 — M7.1 보강 후 재측정 (M7 표의 install/lock 수치 대체)**

| 파일 | 문 커버리지 | 게이트 |
|---|---|---|
| userassets 패키지 | **85.8%** (82.1% → 85.8%) | 상승 |
| install.go | **84.6%** (272 문 중 42 미커버) | 90% 미달 — 잔여 분류 아래 |
| lock.go | **90.7%** (75 문 중 7 미커버) | 90% 도달 |
| 함수 이동 | classifyLockRecord 47.1→88.2% · confinedWrite 67.9→77.4% · confinedMkdir 78.3→91.3% · acquireUserLockStale 81.5→88.9% · Install 87.8→89.4% | |

**잔여 미커버의 정직한 분류 (결정적 주입으로 도달 불가 — 전부):**
(i) I/O 실패 주입 팔 — WriteJournal 영속 4곳 + confinedWrite의 Write/Chmod/
Close 오류 (열린 fd 오류 주입은 권한 의존 또는 런타임 재현 불가);
(ii) 경합 창 팔 — confinedWrite 부모 재검증 재평가·OpenRoot 실패·
reclaimGuardMarker displaced-mismatch 복귀·journal sidecar rename 실패
(단일 스레드에서 그 창에 들어갈 방법이 없음);
(iii) 예측 불가 이름 — O_EXCL 8회 소진 팔 (임시명이 pid+UnixNano 파생,
외부에서 충돌명 선점 불가; 경합 테스트가 계약을 대신 검증);
(iv) 방어 팔 — withinRoot의 join 후 재검사(ValidateRelPath가 `..`를 차단해
confinedWrite 자체 조인으로는 도달 불가)·confinedMkdir seg-continue·
callerIsStored 멤버십 루프(유효 팩 2개 fixture 필요, 현 fixture는 1팩);
(v) withinRoot의 cross-volume filepath.Rel 오류(windows 드라이브 문자 —
unix 실행에서 도달 불가).
(vi) lock.go 잔여 7문도 같은 (i)(ii)류 — 69(비-EEXIST open 오류의 플랫폼
분기)·84(reclaim rename 경합)·107(회수 Release — 보강 테스트가 닫음,
재측정 잔여는 107의 ReadFile 오류 갈래)·215(방어 폴백)·281/287(record 판독
경합).

**게이트**: `go test ./internal/userassets/ -count=1` ok 4.5s (env 스크럽) ·
`go test ./internal/cli/ -run 'TestDoctorGolden|TestBinaryLag'` ok 3.5s ·
golden 스냅샷 3종 재생성(UPDATE_GOLDEN=1, User Lock 행 포함) ·
`golangci-lint run ./internal/userassets/...` 0 issues ·
`GOOS=windows go build ./...` + `go vet` (userassets/cli/hook, 테스트 컴파일
포함) OK · internal/hook 전체 스위트는 이 커밋 미변경 패키지 회귀 확인용으로
백그라운드 진행(완료 시 구조 보고에 합산).

### 게이트 43/44 — 매니페스트 보호 정밀화 + 이중 형식 짝 규칙 + 플랫폼 분리 (2026-10-09)

**43-1 [P1] 매니페스트 보호의 Write/Edit 확장.** 동결 가드의 매니페스트 보호가
Bash 경로만 판정해 `Write` 도구로 user-assets.json을 `{"files":{}}`로
덮어쓰면 허용됐다(리더 Handle 재현). 수리: checkProtectedZone의
ZoneStateOK 분기에 userManifestProtectForms 판정을 선행 배치 — 셸 arm과
동일한 무-추적성 거부(선언된 ZoneUserRoot 항목의 카테고리, 미선언 시
컴파일형 user_manifest). 재현: TestWriteToolCannotReplaceUserManifest
(~/$HOME 스펠링 거부 + 무추적 형제 허용 통제 + 추적 자산 거부 통제).

**43-2 [P2] 과잉 보호 축소.** 최종 접두 조건이 ~/.moai/** 전부를
거부했다(무추적 user-note.txt 삭제 거부 실측). 축소: 매니페스트 자체 +
포함 디렉터(rest == ".moai")만 보호, 나머지 하위 파일은 정상 판정으로.
뒤따른 슬래시 스펠링("~/.moai/")은 trim 후 같은 판정. 재현:
TestUserManifestProtectFormsArePrecise (보호 6스펠링 × 과잉 5스펠링 표).

**44-1 [P1] 이중 형식 짝 규칙.** user_root_zone_forms가 미해결 후보를
해결된 루트와 비교해 루트 자체가 심볼릭 링크일 때(~/.claude → real-claude)
링크 경로 형식을 상실 — 링크 경로로 기록된 매니페스트 키(demo/SKILL.md)가
미등록 real-demo/SKILL.md로 판독될 수 있었다. 수리: 미해결 후보 ↔ 미해결
루트, 해결 후보 ↔ 해결 루트의 동일 기저 짝 비교. 점 세그먼트 rest는 매니페스트
키가 가질 수 없어(hasDotSegment 필터) 별칭 오염 없음. 재현:
TestDualFormsPairedByBasis (이중 심볼릭 링크 + 링크 경로 키).

**44-2 [P1] exec-bit 단언의 플랫폼 분리.** TestConfinedWritePreservesExecBit의
0755 단언은 windows 실행 시 항상 실패(Chmod가 exec 분리를 못 표현 — 정규
파일 문서화 Perm() 0666). 플랫폼 기대값 분리: windows는 문서화 0666 단언,
unix는 0755 단언 (런타임 동작 클래스 — M0 상위 규칙의 형제).

**44-3 [P1] 잠금 마커 형상의 플랫폼 분리.** TestLockGuardPathShape의
.guard 접미 단언은 windows에서 실패 — windows 가드는 acquire 반환 전
.guard를 제거해 .guard.serialize만 남음(리더 고립 windows 경로 실측).
마커 판별에 windows .guard.serialize 사례 추가, 가족 접두·별개 파일 단언은
양 플랫폼 공통 유지.

**44-4 [P2] resume의 ApplyDeployMode + git 훅.** resumeInitializedProject가
deployment_mode를 기록하지 않아 빈 값으로 남고 다음 update가 마이그레이션
경로로 오라우팅됐다(정상 init은 "local" 기록). 수리: ApplyHarness 직후
template.ApplyDeployMode(opts.DeployMode) 삽입 + 미도달 git 훅 설치
(pre-push REQ-CIAUT-002·pre-commit REQ-PC-001, 정상 init과 같은 비치명적
처리) — 대기 중이던 git-hooks 항목과 병합.

**재제기 검증 (현 HEAD에서 재관측)**: project-scoped resume checkpoint
[8th] — initResumeCheckpoint의 프로젝트 범위 arm(gate 37-3) + 비차단 판독
(gate 38-3/30) 착지, TestInitResumeGateKeepsHealthyRedirect(타 프로젝트
저널 무시)로 관측 · readLockRecord [10th] — 플랫폼 분리 파일
(lock_guard_{unix,windows}.go + init_resume_read_{unix,windows}.go, gate
36) 착지, ClassifyGuardMarker/guardMarkerDead 가족 테스트로 간접 관측.

**게이트**: hook zone 계열
(TestFrozenGuard|TestUserRoot|TestWriteToolCannot|TestUserManifestProtect|
TestDualForms|TestProtectedZone|TestShellMatrix|TestShellZone) ok 2.2s ·
`go test ./internal/userassets/ -count=1` ok 4.3s ·
`go test ./internal/cli/ -run 'TestInit'` ok 70.8s ·
golangci-lint (hook/cli/userassets) exit 0 ·
GOOS=windows build + vet (테스트 컴파일 포함) OK ·
internal/hook 전체 스위트 백그라운드 완료 시 구조 보고에 합산.

### 게이트 46/47/48/49 — 보호 기준 경로·실경로 의미론·마이그레이션 정합 (2026-10-09)

**46-1/47-1 [P1] 보호 기준 경로의 케이스 폴드 + 46-2 [P1] 심링크된 ~/.moai 양
스코프 + 48-1 [P1] 심링크+점 결합.** userManifestProtectForms가 rest 비교를
케이스 폴드(~/.MOAI/user-assets.json 동일 아이노드 — resolver가 호출자
스펠링을 반환하므로 폴드가 유일한 정규화)하고, 후보를 미해결(링크 경로)·
해결(실경로) 양 스코프로 판정(~/.moai가 홈 밖으로 링크되면 해결 경로가 홈
접두를 잃음), 점 세그먼트는 텍스트 탈락(~/.moai/./user-assets.json 결합 우회
차단), 그리고 실경로 직접 대등(후보가 매니페스트의 실제 파일/디렉터로
resolve되면 보호 — 심링크+`..` 조합이 홈 상대 rest로 명명 불가인 케이스의
파일시스템 답).

**47-2 [P1] + 48-3 [P2] 이중 형식 별칭 — 점 텍스트 탈락, `..` 실경로 의미론
보존.** userRootZoneForms의 미해결 별칭: `.` 세그먼트는 실경로 효과가 없어
텍스트 탈락(skills/./moai-alpha/SKILL.md가 링크 경로 키로 판정됨), `..`
세그먼트는 절대 문자열 축소 금지 — 심링크 뒤 `..`은 실제로 untracked 파일을
가리키므로(moai-alpha/../moi-alpha/SKILL.md → nested/...) 축소하면 건전한
편집이 거부됨. `..` 포함 스펠링은 해결 형식(파일시스템 워크)이 판정. 재현:
TestDotSegmentAliasInterpretsTrackedKey·TestUserRootPhysicalDotDot (원시
연결 스펠링 — Join 정리 함정 회피)·TestReviewManifestResolvedAliases.

**43-1 후속 — Write/Edit 매니페스트 보호**는 46-48 기준 경로 수리를 그대로
상속(같은 userManifestProtectForms).

**46-3 [P1, 9-11차 재제기] 프로젝트 스코프 체크포인트 — 배포 완성 마커.**
"완성된 MCP/harness config" 프로젝트가 user-global 스테일 저널에 탈취되어
resume이 타 프로젝트의 pending intent를 소비하는 재현 확정. 핵심 발견:
.mcp.json과 llm.yaml의 harness/moai 내용은 템플릿이 이미 품고 있어 존재
판정은 아무것도 구분 못함 — **deployment_mode가 유일한 post-path 전용
마커**(템플릿 미보유, ApplyDeployMode만 기록). 체크포인트 완성 arm을
deployment_mode 존재로 판정: 있으면 완료→redirect(탈취 차단), 없으면 M6
창(배포됨, 설정 미완)→resume. 잔여: MCP-거절 프로젝트는 .mcp.json을 안
쓰지만 deployment_mode로 판정되므로 무관. 재현:
TestInitResumeCheckpointReadsBothDeploymentFaces·TestJournalClearedAfterResume.

**47-4 [P2] resume의 applyJevFromWizard.** wizard의 Jev 답이 resume에서
유실됨(설정이 false로 남음). 수리: wizardRan+wizardResult를
resumeInitializedProject로 전달, ApplyHarness 전 applyJevFromWizard 실행
(정상 init 순서와 동일). 재현: TestInitResumeAppliesJevSelection.

**47-3 [P2] doctor L6 양 스코프.** L4→user-workflow 폴백이 프로젝트
skillsDir만 L6에 넘겨 user-home hns-* 참조가 dangling으로 판정됨. 수리:
checkLayer6AgentActivation에 userSkillsDir 제2 스코프 — 양쪽 모두 부재일
때만 dangling(L1은 프로젝트 스코프 유지). 재현: TestLayer6ResolvesUserScope.

**49 NEW-1 [P2] 워크플로 홈 앵커.** 변환기의 무조건 재작성이
~/.claude/skills/moai/workflows/ → ~/.moai/workflows/로 보냄 — 양쪽 어디에도
없는 대상(실제 변환 복사본은 ~/.agents/skills/moai/workflows/). 수리:
홈 앵커(~ 접두)는 .agents 설치 루트로, 프로젝트 상대형은 .moai/workflows로
구분 재작성.

**49 NEW-2 [P2] 마이그레이션 확인의 converted-vs-converted.**
userCounterpartConfirmed가 RAW 소스와 설치 바이트(CONVERTED)를 비교해
정상 설치된 .toml이 항상 false — 스테일 프로젝트 사본이 영구 고정. 수리:
codex-agents/agents-skills 페이스는 NormalizeCodexRoleForDeploy 적용 후
비교. 재현: TestCodexInstallMigrationAndWorkflowRefs.

**49 [12차 재제기] readLockRecord — 잠금 레코드 판독의 핸들 바인딩.**
classifyLockRecord의 Lstat→ReadFile 쌍이 두 단계 사이 경로를 재해석(동시
스왑 시 판독자가 못 본 레코드를 분류). 수리: readLockRecord 경유로 전환
(비차단 open + 핸들 fstat + 동일 핸들 판독). 재현:
TestLockSwapSubprocess(자식 프로세스가 원자 rename으로 스왑하는 동안 부모
800+ 판정 — 온전한 레코드만, 최소 1회 온전 레코드 관측 보장, ×5 GREEN).

**게이트**: `go test ./internal/userassets/ -count=1` ok 8.1s (스왑 테스트
포함) · hook zone 계열 ok 1.2s · cli 가족(TestInitResume|TestInit|Migrate|
TestCodexInstall|TestDoctor|TestBinaryLag) 백그라운드 · template 전체
백그라운드(변환기 변경) · hook 전체 스위트 백그라운드 · golangci-lint 4
패키지 백그라운드 · GOOS=windows build+vet 백그라운드 — 완료 결과는 구조
보고에 합산. 재현 이름은 게이트가 지정한 4종(TestReviewManifestResolved
Aliases|UserRootPhysicalDotDot|CodexInstallMigrationAndWorkflowRefs|
LockSwapSubprocess)에 바인딩.

### 게이트 50 — 역방향 이중 형식·실경로 조상·참여동의 가족 완성 (2026-10-09)

**50-1 [P1] 이중 형식 역방향 매핑.** 매니페스트가 LINK 경로를 추적할 때
REAL 경로로의 접근이 미등록으로 판돡되어 허용됐다(링크 거부·실경로 허용
실측). 수리: userRootTracksAny가 폴드 키 미적중 시 후보 실경로와 각 추적 키
실경로를 resolve해 비교 — 후보↔추적 양방향 링크 매핑. 재현:
TestDualFormsPairedByBasis 역방향 팔.

**50-2 [P1] 매니페스트 실경로 조상 전체 포함.** ~/.moai → outside 링크에서
rm -rf outside가 허용됐다(파일+직접 부모만 비교). 수리: resolved 매니페스트
경로의 전체 조상 사슬 포함 검사 — 매니페스트는 보호 모델의 루트이므로 임의
조상 수준의 삭제가 모두 보호. 무관 디렉터 통제 팔 포함. 재현:
TestReviewManifestResolvedAliases 조상 팔.

**50-3 [P1] resume의 applyParticipationFromWizard.** 참여동의 wizard 답이
resume에서 유실 — 거절 선택이 기록되지 않아 consent가 enabled로 판돡.
수리: wizardRan+wizardResult 전달 경로에 applyParticipationFromWizard 추가
(Jev ✓·deploy mode ✓·참여동의 ← 이번 — init의 wizard 적용 3단계 가족
완성). 재현: TestInitResumeAppliesJevSelection decline 단언(Asked:true,
Enabled:false).

**48-3↔50-1 조정 기록.** 48-3의 "건전한 편집 거부" 전제는 50-1 역방향
매핑 하에서 소멸: 링크+`..` 스펠링이 착지하는 실파일은 추적 키의 실경로
그 자체(EvalSymlinks 실측 — 링크가 `..` 적용 전 해석됨)라 수정은 관리
자산에 대한 것이고 거부가 옳다. TestUserRootPhysicalDotDot은 실경로
의미론 계약으로 재보정 — 착지 대상이 추적 자산 실파일이면 거부(거부 사유가
해결된 실경로 형식을 명명하는지 검증 — 텍스트 축소 판정 금지), 미추적
파일이면 허용.

**게이트**: hook 팔 4종 ok(재조정 후) · cli TestInitResumeAppliesJevSelection
ok 42.7s · golangci-lint 4패키지·GOOS=windows build+vet·hook 전체 스위트
백그라운드 — 완료 결과는 구조 보고에 합산.

### 게이트 51 — 실경로 포함·체크포인트 수명·하니스 보존·재개방·클로저 prune (2026-10-09)

**51-1 [P1] 실경로 포함 검사.** rm -rf real-demo(추적 파일의 링크가 담긴
실경로 디렉터)가 허용됐다 — 역방향 매핑이 등일만 비교. 수리:
userRootTracksAny의 실경로 비교에 디렉터 포함(tracked 실경로가 후보
실경로의 하위면 추적) 추가. 재현: TestDualFormsPairedByBasis 디렉터 팔.

**51-2 [P2] 체크포인트 수명.** 자산 설치 성공 시 원 증거(저널/파손
매니페스트)가 소비되어, 이후 단계(autonomy 등)의 일시 실패가 다음 실행을
"already initialized"로 막음. 수리: resume-pending 마커(.moai/resume-
pending) — resume 진입 시 기록, 모든 post-step 완료 시에만 제거,
initResumeCheckpoint가 마커를 체크포인트로 판정. 재현:
TestInitResumeCheckpointRetiresOnCompletion(완료 시 소멸 단언 포함).

**51-3 [P2] 하니스 선택 보존.** --llm gpt init이 자산 설치에서 실패 뒤
옵션 없는 재실행이 harness:claude 기본값을 기록하고 Codex 와이어링을
생략 — 배포된 GPT 레이아웃과 설정이 영구 불일치. 수리: 배포 레이아웃이
중단 기록(codex-only 배포자는 CLAUDE.md를 숨기고 AGENTS.md를 투영 —
AGENTS.md 존재+CLAUDE.md 부재 = gpt 기록), resume이 wiring 복원. 재현:
TestInitResumeRestoresGPTSelection.

**51-4 [P2] 두 번째 open의 재개방 창.** classifyPidlessMarker의 flock
프로브 open이 플레인 os.Open — 판독 뒤 FIFO 스왑이 재open에서 블로킹
(실측). 수리: readLockRecord와 동일 규율(O_NONBLOCK + 이 핸들 바운드 타입
검사, 비정규 fail-closed held). 재현: TestUnixPidlessMarkerFIFOOpenNon
Blocking(구현은 hang, 수리는 즉시 반환 — 채널 타임아웃 가드).

**51-5 [P2] RemoveBundle의 클로저 prune.** 라벨 전용 열거가 부모 전용
번들(deployment → extras, 자체 자산 없음)의 의존 자산을 남김 — 의존
자산은 의존 번들 라벨로 기록되기 때문. 수리: 제거 자격 = 전/후 의존
클로저 차등(collectEntries(manifest.Bundles) − collectEntries(remaining)),
소유 엔트리 이름 기준 매니페스트 열거, 공유/보존 게이트 유지. 재현:
TestRemoveBundlePrunesDependencyClosure(파일·기록·L0 보존 3중 단언).

**게이트**: 게이트 51 신규 테스트 5종 전부 GREEN · lint 0 issues(errcheck
2건 수리 후) · GOOS=windows build+vet OK · userassets/template 전체
백그라운드 · hook 전체(단독 부하)·cli 가족 백그라운드 — 완료 결과는 구조
보고에 합산.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-09
- **원장 22건 전부 전환**: M0의 22 의도 RED → M1-M7을 거쳐 **0건** (M1:4 · M2:1+가드2 · M3:4 · M4:1 dual-shape · M5:3 · M6:5+확인2 · M7:5).
- **커밋 계통 20건** (push 없음 — 레인 통합 소관): 58540639d(M0) → 5e6974b39(M1) → ad2ee5a62 → f80590426(M0.1) → c41e67817 → 411ddc6c6 → 8ffa65021(M2) → 6627769b5 → d90df83f0(M3) → 299bc6bd7 → 9671d9922(M4) → 8c1befdb3 → 152c3ef7c(M5) → 00c1b4b3f(M6) → f691c077e → efa062b5d(M7) → 07143ecf7 → 63ad7dc33 → bf9154368 → 470f9ca6e(M7.1-38/42) → f9a58fa4e(M7.1-43/51).
- **턴종료 게이트 51라운드 흡수**: 원장 좌표 외 게이트 발견 ~80건 전부 수리·재검증 — 대표 축:windows reclaim LockFileEx 직렬화(이중 잠금 차단)·user-root 경계 행렬 완성(조상 포함·케이스 fold·심링크 양면·POSIX 원형)·manifest 보호(Write/Edit 확장·freeze·narrowing)·resume post-path 완성(3 wizard 단계+체크포인트 수명)·O_EXCL 배타 생성·golden 스냅샷.
- **AC-023 정직 선언**: journal 93.8%·lock_guard_unix 90.0%·lock.go 92.9% PASS / install.go **84.6%(90% 미달 — formal FAIL 유지)**: 잔여 42문은 §E.2 M7.1행의 5유형(I/O 주입·경합 창·예측 불가 임시명·방어 팔·cross-volume Rel) 주입 불가 팔 — sync-auditor 재판정 대상으로 Gaps 표 이관.
- **레인 최종 검증 배치(독립 관측, HEAD f9a58fa4e)**: go build + GOOS=windows build exit 0 / userassets 9.7s ok · hook 전체 383.5s ok · config 54.3s ok · web 125.8s ok (suites-exit=0) / lint 4패키지 0 issues(구현자 관측+레인 스팟 일치).
- **작성 경위**: manager-develop이 M7.1 통합 보고 후 5시간 사용량 한도 진입(리셋 14:01) — 잔여 §E.3 기입만 남았으므로 레인이 구현자의 §E.2 측정 기록 + 상기 독립 배치를 인용해 대필 (작성자: orchestrator on behalf of manager-develop, 카드 t1591 레인-20).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-10-09
sync_commit_sha: pending-backfill-sync   # 커밋은 자신의 해시를 인용할 수 없음 — 다음 커밋에서 backfill
b12_self_test_a: >-
  pre-emission dedup grep: grep -c 'SPEC-USERASSET-DEPLOY-GUARD-001' CHANGELOG.md → 0 (exit 1) —
  중복 없음 확인 후 등록 진행 (관측 시점 HEAD 0feeb453f)
b12_self_test_b: >-
  AC 카운터(acceptance.md): live=27 excluded=0 ambiguous=0 — CHANGELOG 메타데이터의
  "27 acceptance criteria AC-001..026" 참조와 일치
b12_self_test_c: >-
  파일 경로 검증: CHANGELOG 엔트리가 인용하는 .moai/specs/SPEC-USERASSET-DEPLOY-GUARD-001/spec.md
  존재(ls) · 실행 커밋 3종(58540639d, f9a58fa4e, 0feeb453f) git cat-file -t=commit 관측
changelog_entry_position: "[Unreleased] 첫 배치 ### Fixed 선두 (SPEC-RECEIPT-REUSE-001 엔트리 위)"
frontmatter_status_transitions:
  spec_md: "draft → completed (3-phase close — 단일 sync 커밋 탑재, updated: 2026-10-09 유지)"
  plan_acceptance_design_research: "frontmatter 블록 부재 — status/updated 갱신 대상 없음, 본문 무변경"
canary_compliance_check: "N/A — 본 SPEC은 자기 sync 테스트를 정의하는 선제 정책을 갖지 않음"
docs_claim_check: >-
  docs-site/content 4-locale + docs/: protected-zone 설정 섹션 자체가 미문서화(grep 0적중),
  doctor/init 페이지의 기존 주장과 본 카드 변경 충돌 없음 — 위조된 주장 0건, 4-locale 후속 불필요
```

## §F Phase 4 Mode Selection

- 입력 파라미터: tier L · 범위 약 15-20 파일(구현+신규 재현 테스트) · 도메인 4(internal/userassets, internal/cli, internal/template, internal/web) · 언어 혼합 Go 100% · 병렬 이득 LOW(coding-heavy) · agent-team 전제 N/A(미요청).
- 모드 평가: direct 미해당(다중 파일·의미 변경) / fanout 미선택(coding-heavy — Anthropic coding-task 병렬성 유의) / sweep 미선택(기계 균일 변환 아님·파일 30 미만·상호 의존 마일스톤) / **serial 선택** / agent-team 미요청(실험적 명시 전용).
- **Decision: serial** — 마일스톤 M0→M7 순서에 1회 1개 manager-develop 스폰(트리 내 단일 작성자 — 팩토리 레인 one-writer 규율과 합치).
- 근거: Anthropic coding-task caveat(코딩 과업의 순차 기본) + 결정 가역성 순서(plan §F)가 마일스톤 간 의존을 만드는 구조. 팩토리 레인 크론+태스크 규율이 연속성 담당 — ac_converge goal 미무장(중복 감시자 방지).
- Boundary Case 해당 없음(기준 여유 있음).
