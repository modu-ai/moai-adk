# plan.md — SPEC-USERASSET-DEPLOY-GUARD-001 (카드 t1591)

## §A Context

- 카드: t1591 (Class C — 설계 판단 포함·3-stage 전체). 리더 배차 2026-10-08, lane-20.
- 작업 트리: `.moai/worktrees/t1591`, 브랜치 `WT-user-asset-bundle` @ `db0c514d3` (카드 베이스 81786284e가 인테이크 시 16커밋 FF 흡수됨 — progress.md 인테이크 기록).
- 상위 계약: SPEC-USER-ASSET-INSTALL-001 (completed) — C2 경계·REQ-010·REQ-023·저널-매니페스트 원자성 승계.
- 본 SPEC의 성격: 13개 게이트 릴레이 결함의 수리 묶음. **모든 항목은 릴레이 가설로 시작하며**, 런 페이즈의 첫 행동은 본 트리에서의 RED-first 재현 관측이다. plan-phase에서 좌표를 재고정했고 그 결과(관측/릴레이/이미 수리/좌표 퇴역)가 spec.md §4에 반영되어 있다.
- 결함 가족 7종 → 마일스톤 8개(M0 포함). 파일 영향 약 12개 + 신규 테스트 파일 다수 → Tier L.

## §B Known Issues (재고정 발견 — 런 시작 전 숙지)

1. **원장 11a는 HEAD에서 이미 수리 관측**: `update.go:566-574`의 이전 코멘트("relocated by the repair round, gate r5 finding") + `update_template_sync.go:1185`의 확인창 뒤 호출. M0 재현이 GREEN(재현 불가)이면 REQ-SRF-006은 회귀 가드로 확정하고 원장 항목은 소관 종료 후보로 리더 보고.
2. **원장 5의 ListTemplates 분항도 이미 수리 관측**: `deployer.go:332-340` 코멘트(t1547 수리 라운드, 758→414 팬텀 측정). 이 분항을 다시 "수리"하지 않는다 — 회귀만 확인.
3. **좌표 퇴역 2건**: `internal/template/bundle.go` 부재(파일 퇴역), `deployer_mode.go:37` 메커니즘 소멸(M7 리타이어 → 생 면 `internal/cli/skills.go`).
4. **원장 10a 판별 불가**: `SchemaVersion=1`(manifest.go:25) 뿐이고 `LoadJournal`(journal.go:86-99)은 버전 게이트가 없다 — v1/v2 오판 주장의 좌표를 HEAD에서 찾을 수 없다. REQ-JRN-004(버전 게이트)로 재정식화.
5. **원장 13a/13b 미검증 릴레이**: 카드 본문 명시("미검증 전달", "재현 시 확정"). :584는 좌표 표류(실제 TOML 복사는 :392 경로). M0에서 재현 확정 전까지 구현 착수 금지.
6. **prune 측 부분 관측**: `remove.go`는 R-f-② 의존 유예 팔(RF2, DeferredDeps :115/:305)을 이미 갖는다. 원장 8b의 prune 절반이 `depends_on`을 읽는지는 재현으로 변별.
7. 이전 세대 /tmp 오버레이 스위트는 스테일 — 의존 금지(카드 지시).

## §C Pre-flight

- [x] 베이스 확인: HEAD `db0c514d3` = origin/main tip (인테이크 FF 흡수 관측 완료).
- [x] SPEC ID 정규식 검증: `SPEC-USERASSET-DEPLOY-GUARD-001` → Bash ERE `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → `PASS` (verbatim 인용, plan-phase 기록).
- [x] ID 유일성: `.moai/specs/` 카탈로그에 `SPEC-USERASSET-*` 없음 (grep 관측). 부모 `SPEC-USER-ASSET-INSTALL-001`과는 문자열·디렉터리 모두 상이.
- [x] frontmatter 12 필수 필드 스키마 적합 (spec.md).
- [ ] M0 RED 배터리 관측 (런 페이즈 최초 과업).
- 검증 규율: 카드 워크트리 안에서는 영향 패키지 스위트만 런-로컬로 판정하고 전체 스위트는 CI에 맡긴다(레인-로컬 검증 규율). 병렬 스위트 경합 회피 — 필요 시 slot 임대.

## §D Constraints

- 부모 계약 유지: C2 경계(confinedWrite/confinedMkdir), REQ-010 사용자 파일 불가침, REQ-023 백업, 저널-매니페스트 원자적 해제.
- 병합·베이스 축: 카드 워크트리 기저 브랜치는 **main**이다(AGENTS.local.md §4.1 — develop은 legacy, 카드 PR base `main`). 착지는 `moai factory complete` 경유(push·PR·병합 관측)이며 병합·push는 리더 일괄 소관이다. WT-* 브랜치와 커밋 메시지 카드 id(t1591) 추적성은 유지한다.
- 커밋 주제 규약: `fix(SPEC-USERASSET-DEPLOY-GUARD-001): M<N> ...` — 마일스톤 단위.
- 아티팩트 소관: progress.md §E.2/§E.3은 manager-develop, §E.4는 manager-docs 소관 — 런/싱크 에이전트만 채운다.
- 시간 예측 금지 — 우선순위와 순서만.

## §E Self-Verification

- 각 마일스톤 완료 판정: 해당 AC의 재현 테스트 GREEN + 영향 패키지 회귀 스위트 GREEN + 커버리지 기준(REQ-NFR-001).
- M0 산출물(RED 배터리 관측 기록)이 acceptance.md의 RED-now 칸을 채운다 — verbatim 출력은 progress.md §E.2로.
- 리팩터 드리프트 가드: 계획 파일 대비 실제 수정 30% 초과 시 재계획.

## §F Milestones

> 순서 원칙: 결정 가역성(데이터 모델·신규 인터페이스·사용자 가시 동작) 우선, 기계적 진단을 뒤로. 각 마일스톤의 첫 과업은 그 가족의 RED-first 재현이다(M0가 전체 배터리를 선행 관측).

### M0 — RED-first 재현 기준선 (Priority High)
- 목표: 13개 원장 항목 전체의 재현 테스트를 **본 트리**에서 작성하고, HEAD에서의 RED/GREEN 여부를 verbatim 관측한다.
- 산출: 재현 테스트 파일(가족별), 관측 기록 → progress.md §E.2, acceptance.md RED-now 칸 확정.
- 판정: 항목별 분류 확정 — RED-first는 `EXPECTED_RED` 관측, born-green 가드(11a·5-ListTemplates·5d-emission·2a-사용자 측)는 GREEN 상륙 기록, 게이트(AC-023/024)는 절차 판정.
- 위험: FIFO/잠금/중단 재현은 타이밍 민감 — 표 테스트+시임 주입으로 결정화.

### M1 — 저널 영속화·복구 무결성 (원장 3, 6a, 10, 10a-재정식) (Priority High)
- 좌표: `internal/userassets/install.go`(227-304, 239, 255, 297), `journal.go`(LoadJournal).
- 과업: (a) 회수 항목의 첫 스테이징 전 병합(REQ-JRN-001), (b) refresh 시 반입 항목 해시 갱신(REQ-JRN-002), (c) 파일별 즉시 플래그 영속화(REQ-JRN-003), (d) LoadJournal 버전 게이트(REQ-JRN-004).
- 설계 구속: design.md §2 — 저널 clear의 매니페스트 저장 원자성 유지, RF8/B3 계약 유지.
- 판정: AC-002~005 (TestJournalStageCarriesRecoveredEntries, TestJournalRefreshUpdatesCarriedHash, TestWriteCompletedPersistedPerFile, TestJournalUnknownSchemaRefused).
- 위험: 파일별 영속화가 쓰기 증폭 — stage 저널의 원자 쓰기 재사용으로 최소화.

### M2 — 잠금 가드 복구 (원장 1, 10-P1) (Priority High — P1)
- 좌표: `internal/userassets/lock_guard_windows.go`(전체), `lock_guard_unix.go`, `lock.go`(guardPath 54-59), `lock_owner_windows.go`.
- 과업: 마커에 PID 소유권 기록 + 사망 확인 후 회수(REQ-LOCK-001), windows/unix 의미 일치(REQ-LOCK-002).
- 설계 구속: design.md §3 — `lockOwnerGone` 자세 재사용, RF8 토큰 계약 유지.
- 판정: AC-001 (TestGuardMarkerReclaimAfterOwnerDeath) + `GOOS=windows go build` 게이트 (AC-024).
- 위험: CI windows 매트릭스 미검증(릴레이 선언) — 로컬 판정은 플랫폼 중립 표 테스트로.

### M3 — Codex 정규화 3-지점 변환 (원장 7a, 13a, 13b) (Priority High — P1)
- 좌표: `internal/userassets/install.go`(373, 392, 584 경유) + 기존 변환면 `internal/template/harness_fs.go`(`NormalizeCodexRoleForDeploy` :180·`normalizedOpen` :209)의 설치 경로 배선 — 신규 변환기 아님(design.md §4).
- 과업: 기존 변환기의 설치 경로 배선 — 기록과 해시가 동일 변환 바이트를 쓰게 한다(REQ-CNV-001), 미배포 참조의 파일 단위 보고(REQ-CNV-002).
- 선행 조건: **M0에서 13a/13b 재현 확정 후에만 착수** — 반박 시 범위 축소(7a만).
- 판정: AC-006, AC-007(13b 좌표 확정 팔 포함), AC-026 (TestCodexAgentTOMLReferencesConverted, TestCodexUserSkillInstallNoClaudeOnlyRefs, TestCodexRoleTOMLNotVerbatim, TestCodexUnconvertibleReferenceReported).
- 위험: 변환 규칙의 과잉 일반화 — 재현이 보인 참조 좌표만 변환, 나머지는 보고.

### M4 — 동결 가드 사용자 경로 보호 (원장 9a) (Priority High — P1)
- 좌표: `internal/hook/pre_tool.go:354-355`(config 적재면), `internal/hook/protected_zone_path.go`, 보호 집합 데이터 원천 — design.md §5.
- 과업: 보호 집합에 네 사용자 설치 루트 반영(REQ-GRD-001/002). import 방향(hook→userassets 미임포트 관측) 해결: 경로 상수 공유 계층 또는 config 목록+패리티 테스트.
- 판정: AC-009 (TestFrozenGuardDeniesUserPathDelete).
- 위험: 보호 과잉으로 사용자 합법 편집 차단 — 보호 대상은 "사용자 설치 루트 아래 moai 관리 파일"로 한정, 집합 정의를 design.md에 고정.

### M5 — 충돌 판정·confined 쓰기 안전화 (원장 8c, 12, 6b) (Priority Medium, 12=P1)
- 좌표: `internal/userassets/install.go`(211, 715-755).
- 과업: 충돌 사전 판정의 Lstat 선행(REQ-COL-001), 부모 핸들 고정 쓰기로의 전환(REQ-COL-002 — 경로 재검증 설계는 TOCTOU를 못 닫아 폐기, design.md §6), 설치 스크립트 실행 권한 보존(REQ-COL-003).
- 판정: AC-010, AC-011, AC-025 (TestCollisionPrecheckSkipsFifoWithoutBlock, TestConfinedWritePinnedToValidatedParent, TestConfinedWritePreservesExecBit).
- 위험: FIFO 재현은 mkfifo 플랫폼 차 — unix 표 테스트 + windows 빌드 게이트.

### M6 — 배포·번들 표면 정합화 (원장 4, 5-잔여, 8a, 8b, 8d, 11a-가드, 11b) (Priority Medium)
- 좌표: `internal/cli/migrate_project_assets.go`(89, 119-122), `internal/cli/doctor_agentemit_embed.go`(135-160), `internal/cli/skills.go`+`codex_skills_disable.go`, `internal/web/agentfm.go`(123-132, 278-285), `internal/userassets/install.go`(455-468)+`remove.go`, `internal/cli/update_template_sync.go`(회귀 가드 대상), `internal/cli/init.go`(905, 944-948).
- 과업: (a) 미등록 미러 사본 분류·보고·제거(REQ-SRF-001), (b) emission 미비보고 불변 가드 상륙(REQ-SRF-002 — CheckFail 유지, design.md §7), (c) skills disable의 사용자 계층 폴백(REQ-SRF-003), (d) 동명 에이전트 행의 단일 설정 통합(REQ-SRF-004 — 저장 계약 design.md §7, end-to-end AC-015), (e) depends_on 클로저 설치+보존 집합(REQ-SRF-005 — prune 절반은 M0 변별 후), (f) 취소 불변 회귀 가드(REQ-SRF-006), (g) init 재개 경로(REQ-SRF-007 — ensure + 미실행 후속 설정 단계, design.md §7).
- 판정: AC-012~018.
- 위험: (d)는 web 폼 계약 변경 — 폼 키 스코프화가 기존 소비자(post 핸들러)와 정합하는지 회귀. (g)는 UX 문구·재개 범위의 설계 판단 포함.

### M7 — doctor 진단 정합화 (원장 2, 7b, 9b) (Priority Medium)
- 좌표: `internal/cli/doctor_user_install.go`(130-187), `internal/cli/doctor_harness.go`(52-62).
- 과업: 프로젝트 측 적재 위장 수리 + 사용자 측 정직성 가드(REQ-DOC-001 — AC-019a RED / AC-019b 가드), 공허 일치 위장 제거(REQ-DOC-002), 필수 파일 존재 기반 루트 선택(REQ-DOC-003), 폴백 시 L1 프로젝트 유지(REQ-DOC-004).
- 판정: AC-019a/b, AC-020~022.
- 위험: 가장 기계적 — 진단 출력 형식 변경이 기존 doctor 테스트와 충돌할 수 있음.

## §G Anti-Patterns

- 분류 없는 GREEN 금지 — 모든 AC는 M0에서 RED-first / born-green 회귀 가드 / 빌드·측정 게이트로 분류 확정되며, RED-first만 관측 RED를 전제로 전환된다(가드의 RED 관측은 퇴행 결함 보고).
- 이미 수리된 좌표(11a, 5-ListTemplates)의 재수리 금지 — 회귀 가드만.
- 스테일 오버레이(`/tmp/t1547-review-overlay.json`) 참조 금지.
- 원장 좌표를 문자 그대로 믿고 표류 좌표(:584, bundle.go:145)를 구현 금지 — M0 재확정 좌표만.
- 부모 계약(C2/REQ-010/REQ-023) 우회 금지.
- `go test ./...` 전체 스위트의 레인 로컬 반복 금지 — 영향 패키지 + CI.

## §H Cross-References

- 부모: SPEC-USER-ASSET-INSTALL-001 (계약 승계).
- 상호참조: SPEC-PROGRESS-RECORD-IO-001 (t1560 이관분 원장).
- 형제: t1594 (t1547 r5–r12 잔여 — 본 SPEC 비소관), t1547 도메인 정합 축 (리더 보유 후속).
- 증거: `.moai/reports/t1591/` (런 페이즈 판정·card-review), progress.md §E.
