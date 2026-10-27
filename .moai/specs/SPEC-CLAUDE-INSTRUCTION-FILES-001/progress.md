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

### M5 — 검증 총괄 (측정 트리 = M4 커밋 `1b016070ee` 작업 트리, 2026-10-27)

#### §D.1 green 방향 재검증 (전 행)

| 행 | 관측 | exit |
|---|---|---|
| 1 | `--- PASS: TestCodexOnlyForceUpdateVariant (0.18s)` — CLAUDE.md 배치·AGENTS.md 부재 단언으로 재타겟된 테스트 green | 0 |
| 2 | 구형 `ReplaceAll(text, "CLAUDE.md", "AGENTS.md")` grep exit 1 · 신형 `:208 text = strings.ReplaceAll(text, "AGENTS.md", "CLAUDE.md")` (+ `:207 AGENTS.local.md→CLAUDE.local.md`) 관측 | 0 |
| 3 | 구형 Short 문구 exit 1 · `:62 Short: "Move AGENTS.md/AGENTS.local.md to CLAUDE.md/CLAUDE.local.md, keeping backups"` | 0 |
| 4 | `:45 3: {Path: "CLAUDE.local.md", BlockType: BlockTypeLearnedLocal, Tier: 3}` · Tier4 `CLAUDE.md` | 0 |
| 5 | `ls internal/template/templates/CLAUDE.md.tmpl` → 22,499B 존재 | 0 |
| 6 | `instructions_loaded.go:5`(레거시 폴백/라벨 관련)·`init_resume.go:2`·`doctor_harness.go:3` — 감소+맥락 전환 | 0 |
| 7 | `:105 refusing to migrate: … both exist; merge … by hand` + `:355-356` advisory | 0 |
| 8 | `:113 info, err := os.Lstat(legacyPath)` — 정규 방향 원본 검사 | 0 |
| 9 | `grep -c ModTime` = 2 (읽기→쓰기 가드) | 0 |
| 10 | `grep -c '"AGENTS.md"'` = 1 (쌍 테이블 레거시 엔트리) | 0 |

#### AC 매트릭스 (§E.2)

| AC | 상태 | 근거 (명령 → 관측) |
|---|---|---|
| AC-CIF-001 | PASS | `go test ./internal/template/ -run '^TestCodexOnlyForceUpdateVariant$' -v` → `--- PASS`; `TestHarnessProfilesResolveSharedReferences` 3프로파일 PASS(스위트 내); 상수 ClaudeMD/AgentsMD 둘 다 유지(주석·역할만 반전) |
| AC-CIF-002 | PASS | `TestInitCmd_NonInteractiveExecution`(CLAUDE.md 생성·AGENTS.md/AGENTS.local.md 부재 os.Stat 관측) + `TestInit_NoInstructionStubOnFallback` — internal/cli 스위트 내 green |
| AC-CIF-003 | PASS | `TestResolveInstructionPairPreference` — **양쌍(공유·로컬)×3형** 선택·`legacy AGENTS fallback` 라벨 토큰·사전/사후 해시 불변(N13 소실) + `TestInstructionFileSetDerivation`(레거시 전용 앵커 폴백) — internal/hook ok |
| AC-CIF-004 | PASS | `normalizeHarnessReferences` 반전 관측(위 표 행 2) + 정규화 단위 테스트(스위트 내) + `ls CLAUDE.md.tmpl` exit 0 |
| AC-CIF-005 | PASS | `TestMigrateLocalInstructions_MovesWithBackup` 양쌍(테이블 주도) — 프리미티브 재사용(공재거부·Lstat·타임스탬프 백업·O_EXCL·보존), 신규 메커니즘 부재 |
| AC-CIF-006 | PASS | 동일 테스트 — 백업 사본 바이트 동일(digest)+권한 0640 보존+원본은 최종 rename 전 제자리 |
| AC-CIF-007 | PASS | `RefusesCoexistence` 양쌍 — 쌍별 exit≠0·양쪽 해시 불변·백업 0건·수동 병합 지시 |
| AC-CIF-008 | PASS | `RerunIsNoop` — 재실행 noop 메시지·파일·백업 수 불변 |
| AC-CIF-009 | PASS | `RefusesNonRegular`(양쌍 symlink, 대상 생존) + `RefusesUnreadable`(0000) — 경로이탈=인자 없는 고정 이름+링크 비추종으로 구조적 차단(코드 주석 명시) |
| AC-CIF-010 | PASS | `SourceChangedDuringRead`(사전 쓰기 중단)·`FinalRenameRefusal`(미편집/편집 — 상태 무관 보존+사본 기반 보고)·`DestinationWriteFailure`(N9 mutant 차단)·`EEXISTForeignDestination`(N12 race 행)·`RollbackRenameFailure`((e) 잔여 경로+보존 실패 보고) — 전 출구 3-위치 무손실 불변 단언 |
| AC-CIF-011 | PASS | 순수 이동(cmp 바이트 동일 M2 관측) — 비공개 내용 승격 구성상 0건·동사 출력 경로만 포함·자동 분류 경로 부재 |
| AC-CIF-012 | PASS | `TestSurfaceForTier_Tier3/Tier4`(CLAUDE.local.md/CLAUDE.md) + token_budget 고정 슬롯·미러 CLAUDE 축 테스트 green |
| AC-CIF-013 | PASS | `go test ./internal/contract/` ok — 4-이름 목록·pin 불변(이관 전후 green) |
| AC-CIF-014 | PASS | `.moai/reports/**` git 변경 0건·`.moai/backups/local-instructions/` 신규 4디렉터리만 |
| AC-CIF-015 | PASS | 루트 CLAUDE.md/CLAUDE.local.md 존재(verb, git rename 100% 유사도)·doctor/init 정규 파일 동작(advisory 테스트가 실물 update+doctor 구동)·`.claude/rules/**`+mirror 활성 AGENTS 지시 0건(잔여: hns-lsel 2건=frozen 계약 서술, frozen/protected 테이블=t frozen 계약) |
| AC-CIF-016 | PASS | doctor L3 CLAUDE-우선·init_resume 프로브(`TestDeployedWiringIsGPTReadsTheClaudeSurface`)·update leg: 실물 `moai update`가 사용자 AGENTS.local.md 보존+CLAUDE.local.md 비생성(`TestLocalInstructions_UpdateDoctorPreserveFile`)+공유쌍: AGENTS.md를 덮쓰는 템플릿 부재로 보존 구성 보장+공재 시 advisory→verb 거부 순서(`AdvisoryParity/both`) |

#### E2/E3/E4/E5 관측

- E2: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (이 트리, 이번 실행).
- E3 (패키지): core/project 88.9 · config 85.5 · constitution 88.4 · contract 96.3 · harness/curator 91.4 · template/agentemit 90.4 · harness 86.4 · hook 87.9 · **template 82.6 / cli-harness 80.8 — 수정 파일 축으로는 harness_fs.go 핵심 함수 100%(normalizeHarnessReferences/isHidden)·install.go RunInstall 92.9%/NewInstallCmd 87.5%, 패키지 총계 부족분은 비수정 파일의 기존 부채** · internal/cli 전체(-cover): 1차 실행 `panic: test timed out after 10m0s`(계측 지연, 테스트 결함 아님 — verbatim 관측) · 2차(-timeout 30m, 백그라운드): 선언 창 내 결과 미수신 — **관측되지 않음(아래 Gaps)**, 리더 판정(2026-10-27)으로 touched-package 축으로 기록하고 카드를 막지 않음.
- E4: `grep -rn 'AskUserQuestion' internal/hook/ internal/harness/ | grep -v _test | grep -v '// '` → 2건 — `pre_tool.go:856`(툴 이름 비교 — 호출 아님, frozen 파일 미수정)·`proposalgen/scaffolder.go:164`(인용 산문, 미수정 파일). 신규 호출 0건.
- E5: `golangci-lint run --timeout=4m` exit 1 — 이슈 1건 = 기준선(`leader_readers_darwin.go:93` SA1019) 동일, **NEW 0건**.
- 기타: `moai spec lint` exit 0 — 0 error(전 SPEC 경고 8,592건은 기존).

#### Gaps

- **internal/cli 전체-suite 커버리지: 관측되지 않음.** 1차 `go test ./internal/cli/ -cover` = `panic: test timed out after 10m0s`(go test 기본 타임아웃, 계측 지연 — 패닉 로그 verbatim 관측). 2차 `-timeout=30m` 백그라운드 재실행의 결과가 선언 창 내 도착하지 않아(통지 채널 불안정) 관측 없음 — 커버리지 판정은 touched-package 축(hook 87.9·harness 86.4·curator 91.4·cli/harness 수정 파일 87.5~92.9 등)으로 대체하고 본 항목을 Gap으로 명시(verification-claim-integrity §3.1 — 관측 안 된 검증은 결코 PASS가 아님). internal/cli의 방향 민감 표면은 대상 실행(verb/advisory·init/resume·headroom·import-resolution)으로 전부 green 관측됨.
- AC-CIF-016 공유쌍 update leg는 구성 증명(AGENTS.md 템플릿 부재+deployer 보호 스킵+advisory→verb 순서)으로 관측 — 단일 통합 테스트로 직접 주입하지는 않음.
- token_budget_guard.go의 코멘트 서술 다듬기(AGENTS 시대 근거 문구)는 동작 무관 — sync 단계 문서 정리로 이월 가능.

#### N12/N13 부채 소실 판정 (run-exit 재감사용)

- **N12 소실**: REQ-CIF-011 구현체(`migrate_local_instructions.go` EEXIST 분기 + 주석 "ownership proof = this process's own O_EXCL creation")와 `TestMigrateLocalInstructions_EEXISTForeignDestination`(양쌍 — 외부 파일 백업 이동 0건·양쪽 제자리 단언)로 명시화.
- **N13 소실**: `TestResolveInstructionPairPreference` fixture 테이블이 로컬 쌍을 포함(공유·로컬 × coexist/canonical-only/legacy-only) — 로컬 쌍 우선 mutant 불통과.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete
run_complete_at: 2026-10-27T13:48:34Z
run_commit_sha: bcdc715517
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "git rev-list --count origin/main...HEAD = 0 10 (관측 2026-10-27, 커밋 전 HEAD 1b016070ee)"
l44_post_push_fetch: not-applicable-no-push-lane-owns-delivery
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 98   # git diff --name-only 746b8af14e..HEAD (M5 docs 커밋 전 기준)
m1_to_mN_commit_strategy: "per-milestone commits M1..M5 (feat/docs, SPEC id in subject, card t1985 in body, Authored-By-Agent: manager-develop); no push — lane owns delivery"
coverage_note: "touched-package axis recorded in §E.2 M5; internal/cli full-suite coverage unobserved (Gaps)"
```

_run_status의 `complete`는 run 단계 관측 종결을 뜻한다(사양 상태 전이 아님 — spec.md는 M1 커밋에서 이미 `in-progress`; `in-progress → implemented → completed`는 manager-docs 소유)._

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-10-27T14:00:03Z
sync_commit_sha: pending-backfill-sync   # a commit cannot cite its own hash; the lane backfills the real SHA in the next commit
changelog_entries_added: 1
changelog_entry_position: "[Unreleased] > Changed — SPEC-CLAUDE-INSTRUCTION-FILES-001 (appended after the SPEC-CI-CODEX-REVIEW-001 entry)"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed on the single sync commit (3-phase close)"
  plan_md: "stateless — no status field (file carries no YAML frontmatter block; nothing to refresh)"
  acceptance_md: "stateless — no status field (file carries no YAML frontmatter block; nothing to refresh)"
b12_self_test_a: "grep -c 'SPEC-CLAUDE-INSTRUCTION-FILES-001' CHANGELOG.md = 0 observed before emission (exit 1, no duplicate)"
b12_self_test_b: "AC live count = 16 (AC_FILE=acceptance.md, tier M; live=16 excluded=0 ambiguous=0); CHANGELOG entry references the same 16 criteria AC-CIF-001..016"
b12_self_test_c: "every path named in the entry verified present (internal/defs/files.go, internal/cli/migrate_local_instructions.go, internal/template/harness_fs.go, internal/cli/harness/install.go, internal/cli/update.go, internal/cli/doctor.go)"
readme_sync: "README.md / README.ko.md / README.ja.md / README.zh.md harness-table rows and common-assets paragraph corrected (AGENTS.md -> CLAUDE.md); docs-site untouched (out of card scope)"
lint_observation: "moai spec lint SPEC-CLAUDE-INSTRUCTION-FILES-001 = 'No findings' observed pre-commit at HEAD 5637461a26"
```

_sync_status의 `complete`는 sync 단계 산출물 종결을 뜻한다(3-phase close: `in-progress → implemented → completed` 전이가 본 싱크 커밋에 탑재)._
