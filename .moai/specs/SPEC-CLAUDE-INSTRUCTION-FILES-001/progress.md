# progress.md — SPEC-CLAUDE-INSTRUCTION-FILES-001

> 카드 t1985, lane-16. §E 섹션 골격 — §E.1만 plan-phase에서 작성, §E.2~§E.4는 placeholder.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-audit
plan_complete_at: 2026-10-27
artifacts: spec.md (16 REQ, GEARS), plan.md (M0~M5), acceptance.md (AC-CIF-001~016, GUIDE-01..12 조정표, RED-now 기준선 @cb30e05b3e)
run_gate: iter2+ 독립 감사 PASS(최신 plan HEAD에 결속) + 리더 킥오프 판정까지 run 금지; t1984는 본 카드 merge-ready 뒤 임대·plan, 본 카드의 PR 병합·main 흡수가 t1984 RUN에 선행 (plan.md §C, 리더 §15.124 확정 순서)

## §E.2 Run-phase Evidence

### M0 — run 진입 게이트 (측정 트리 `746b8af14e`, 커밋 전, 2026-10-27)

- Pre-flight: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · golangci-lint exit 1 = 사전 존재 기준선 1건(`internal/discovery/leader_readers_darwin.go:93` SA1019, 본 카드 비관련) · t1984 소유 3파일 존재·비편집 확인.
- RED-now 재측정: acceptance.md §D.1 10행 + 행동 행 전부 verbatim 부합(행 1 `--- PASS: TestCodexOnlyForceUpdateVariant` = AGENTS 배치 단언 → CLAUDE-정규 기준 red; 행 5/9/10 exit 1 표면 부재). 리더 파일 목록 승인(2026-10-27T17:1x) 후 M1 착수.

### M1 — 정규 이름·생성 경로 전환 (RED→GREEN, 2026-10-27)

- **E8 RED (구현 전환 전, 트리 `746b8af14e`+테스트 재타겟)**: `TestCodexOnlyForceUpdateVariant` FAIL(`CLAUDE.md missing from force-update codex deploy` + `AGENTS.md deployed … must not be created`) · `TestCodexOnlyDeployerWalkIntegrity` FAIL(`retired AGENTS instruction file "AGENTS.md" visible` + `"CLAUDE.md" missing from listing`) · `TestHarnessProfilesResolveSharedReferences` FAIL(3 프로파일 전부 CLAUDE.md 부재/AGENTS.md 존재) · `TestValidate/existing_CLAUDE.md_triggers_warning` FAIL(`expected warning containing "CLAUDE.md already exists", got […]`) · `TestValidateMoAI_MissingClaudeMD` FAIL(`expected warning about missing CLAUDE.md, got [.claude/ … AGENTS.md not found…]`) · `TestDeployedWiringIsGPTReadsTheClaudeSurface/codex-only_layout` FAIL(`deployedWiringIsGPT = false, want true`) · `TestInitCmd_NonInteractiveExecution` FAIL(`expected CLAUDE.md to be created` + `AGENTS.md must not be created by init`). 대칭적 2건(`TestInitResumeRefusesGPTLayoutWithoutRecord`, `TestInitResumeDoesNotInferBothFromInstructionFile`)은 현행 green — 회귀 가드.
- **전환**: `defs/files.go` 상수 역할 반전(둘 다 유지) · `templates/AGENTS.md.tmpl`→`CLAUDE.md.tmpl` git mv · `embed_manifest_gen.go` :647 재생성(make embed-manifest, diff 1행) · `harness_fs.go` codexHiddenPaths에서 CLAUDE.md 제외 + normalizeHarnessReferences AGENTS→CLAUDE 양쌍 반전 · `validator.go` CLAUDE.md 존재 검사/AGENTS.md 불활성 · `init_resume.go` deployedWiringIsGPT 프로브 CLAUDE.md 축 · `initializer.go` 주석.
- **따라옴(이름 참조)**: template 패키지 테스트 11파일 + agentemit 테스트 2파일 + `obligations.yaml` claude_path 2행 + config/constitution 테스트 이름 참조. `token_budget_guard.go:141` 미러 경로만 CLAUDE.md.tmpl로(루트 축 :131은 M3/M4와 함께 — 조기 전환 시 이 저장소 측정 소멸). zone-registry `file: AGENTS.md` 14행 전환은 시도 후 **보류 복원** — registry 전환은 저장소 루트 CLAUDE.md(=M4 verb 전환)를 요구하므로, M1 중간 상태에서는 `registry_sync_test.go`의 배포 미러 매핑(rel AGENTS.md → source CLAUDE.md.tmpl)으로 브리지. M4에서 registry 플립+브리지 제거로 종결.
- **GREEN**: `go test ./internal/template/ ./internal/template/agentemit/` ok(369.9s/0.6s) · `./internal/core/project/` ok · `./internal/config/ ./internal/constitution/ ./internal/cli/harness/` ok · cli 대상(init/resume/headroom/import-resolution) ok. `go build ./...` exit 0 · windows build exit 0 · lint = 기준선 1건 동일(NEW 0).
- **Gaps**: internal/cli 전체 스위트 실행은 M5 총괄로 연기(리소스 규율) — 대상 실행으로 방향 민감 표면 커버.

### M2 — 마이그레이션 동사 방향 전환 + 무손실 원칙 + 자체 저장소 전환 (RED→GREEN, 커밋 `8884b02f4a`)

- **E8 RED (구현 전, 구 방향 구현 against)**: 9개 테스트 그룹 전부 FAIL — `TestMigrateLocalInstructions_{RefusesCoexistence(공유쌍 exit 0), MovesWithBackup(canonical 부재), RerunIsNoop, RefusesNonRegular(exit 0), RefusesUnreadable(exit 0), SourceChangedDuringRead(exit 0), FinalRenameRefusal(exit 0), DestinationWriteFailure(exit 0), EEXISTForeignDestination(exit 0), RollbackRenameFailure(exit 0)}` — verbatim 출력 §E.2 원본 기록(커밋 시점 로그).
- **구현**: 쌍 테이블(공유+로컬 양쪽 필수) · 결정 우선(공재=전면 거부·무기록) · Lstat 비정규 거부(경로이탈=링크 추종 거부로 대표, 인자 없는 고정 이름) · 백업 사본(바이트·권한) → 읽기→쓰기 가드(mtime+내용) → `writeExclusiveKeepPartial`(O_EXCL) → 최종 단일 원자적 `os.Rename`(원본→신규 타임스탬프 백업 경로) · 실패 출구 전부 무조건 보존(상태 무관) + 보존된 사본 기반 보고 · rollback rename 자체 실패=양쪽 제자리+잔여 경로 보고 · **N12 반영**: EEXIST 목적지=타 프로세스 소유(O_EXCL 자기 생성만 소유권 증명) — 보존·이동 없이 양쪽 부재 변화 없이 거부, 코드 주석+테스트 행으로 명시.
- **advisory 방향 반전**(update/doctor 공용 1문본) + doctor L3 CLAUDE-우선.
- **자체 저장소 전환 (M4 코어, M2 커밋에 동반)**: 이유 — zone-registry `file:` 플립(M1 재명의 배포 정합에 필수)은 저장소 루트 CLAUDE.md 없이는 local-arm 검증이 불가해 M1↔M4 순환. verb 실물 실행(`go run ./cmd/moai migrate local-instructions`): 양쌍 이동, `cmp` 바이트 동일 관측(LOCAL/SHARED 모두), 원본 보존 `.moai/backups/local-instructions/20261027T083748Z-*/`(gitignored, 역사 항목 불변). zone-registry+template mirror 플립(cmp IDENTICAL), token_budget_guard 루트 슬롯·미러 CLAUDE 축, `.gitignore` `/AGENTS.local.md`·`/CLAUDE.local.md` 제거(추적 쌍 유지), git rename 100% 유사도로 기록.
- **GREEN**: `go test ./internal/cli/ -run '^TestMigrateLocalInstructions|^TestLocalInstructions'` ok · `./internal/config/` ok · `./internal/constitution/` ok · template 대상(AgentsLocal|SweepGuard|Section20) ok.

### M3 — 런타임 리더/작성기 재타겟 (RED→GREEN, 2026-10-27)

- **E8 RED**: curator 행동 RED — `TestSurfaceForTier_Tier3` FAIL(`Tier 3 Path = "AGENTS.local.md", want CLAUDE.local.md`)·`Tier4` FAIL(`Path = "AGENTS.md", want CLAUDE.md`)·`TestPrepareTierDispatch_Tier3` FAIL(동일 축). hook reader — 신규 API(`resolveInstructionFileDetail`)는 구현 전 상태가 컴파일 실패(`undefined: resolveInstructionFileDetail/resolveLocalInstructionFileDetail`, verbatim 캡처)이며 이것이 신규 API의 사전 관측; 방향 민감 기존 테스트(`TestInstructionFileSetDerivation` 앵커 CLAUDE 재타겟)는 동일 컴파일 단에서 관측.
- **구현 (N13 반영)**: `instructions_loaded.go` — `resolveInstructionPairDetail(root, canonical, legacy)` 일반화 + 공유/로커 양 세부 resolver, CLAUDE-우선, 레거시 진입 시 리터럴 라벨 `legacy AGENTS fallback` 반환 + 핸들러 warn 로그로 관측 가능화 + 핸들러가 로컬 쌍도 예산 검사; **`TestResolveInstructionPairPreference` 테이블 = 양쌍(공유·로컬) × 3형(coexist/canonical-only/legacy-only) — 선택·라벨 토큰·사전/사후 해시 불변 단언, 로컬 쌍 우선 mutant 통과 불가(N13 소실 관측)**. `curator/dispatch.go` TierSurfaceMap Tier3→CLAUDE.local.md·Tier4→CLAUDE.md. `layer5.go` @import/마커 문자열. `observer.go` eagerWeightSources 순서·주석. `harness/install.go` 마커 주입 CLAUDE-우선+레거시 AGENTS 폴백·오류 문구. `update/plan/plan.go` 4-이름 리스크 분류 유지(AGENTS.md·CLAUDE.md 둘 다 high+SectionMerge — 직독 확인, 편집 불요).
- **frozen 계약 (AC-CIF-013)**: `go test ./internal/contract/` ok — 4-이름 목록·pin 불변 (현행 green 유지 관측).
- **GREEN**: `./internal/hook/` ok(383s 전체) · `./internal/harness/curator/` ok · `./internal/harness/ ./internal/contract/` ok · `./internal/cli/harness/` ok.

### M4 — 룰/스킬/미러 스윕 (자체 전환 본체는 M2 커밋에 동반, 2026-10-27)

- **스윕**: `.claude/rules/**` 13파일 + local 2파일, `.claude/skills/**` 11파일, `.claude/output-styles/moai/moai.md`, template mirror 13+6파일 — 전부 활성 참조(§13·§4.0 인용, paths: 스코프, 마커 계약, 메모리 위계)로 판별되어 CLAUDE 축으로 재지정. **의도적 보존**: hns-lsel-* 2건(REQ-CIF-014 frozen 4-이름 계약 서술 — AGENTS.md 보호 이름 생존이 서술의 전제), protected-zone/frozen 테이블(frozen 계약), t1984 소유 3파일, `.moai/policies/**`(AC 범위 밖·일부 금지).
- **미러 큐레이션 존중**: agent-common-protocol-reference.md 미러는 내부 토큰 제거 fork — cp 재유입(누출 검출) 후 복원, 해당 미러는 원래부터 AGENTS 참조 0건이었음이 확인됨.
- **ledger 재동결 (TMD-001 캐스케이드)**: binding_ledger.json 6행(ALB-0049/0080/0121/0135/0138/0197) after_text=CLAUDE 본문, before_text=git 앵커(a2a184ad3) 원문 보존 · output_style_ledger.json 2행(moai-0015/0245) 재동결(before=after=신규 본문) + moai.md anchor_body_sha256 재계산(5380fc2f…) · catalog.yaml 해시 재생성(gen-catalog-hashes --all, moai·moai-workflow-testing 갱신).
- **AC-CIF-011 (프라이버시)**: 전환은 순수 이동(병합 없음) — CLAUDE.md ≡ 구 AGENTS.md 바이트 동일(cmp 관측)이므로 로컬 비공개 내용의 승격은 구성상 0건. 동사 출력·오류는 경로만 포함(본문 없음). 모호 콘텐츠 자동 분류·게시 경로 부재(수동 병합 안내만).
- **AC-CIF-014 (역사 불변)**: `.moai/reports/**` git status 변경 0건 · `.moai/backups/local-instructions/` 기존 항목 없음(본 워크트리), 신규 4디렉터리만(verb MkdirTemp).
- **GREEN**: `go test ./internal/template/` ok(356.7s 전체 — binding ledger·output-style ledger·catalog hash·내부누출·korean-leftover 전부).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
