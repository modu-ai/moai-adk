---
id: SPEC-CLAUDE-INSTRUCTION-FILES-001
title: "활성 instruction-file 표면의 CLAUDE.md/CLAUDE.local.md 정규 쌍 전환"
version: "0.2.5"
status: draft
created: 2026-10-27
updated: 2026-10-27
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/defs, internal/contract, internal/cli, internal/hook, internal/harness, internal/template, internal/core/project"
lifecycle: spec-anchored
tags: "instruction-files, claude, agents, migration, template, frozen-contract, privacy"
tier: M
related_specs: [SPEC-INSTRUCTION-FILES-UNIFY-001, SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001, SPEC-AGENTS-MD-CANON-001]
---

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-27 | 최초 plan-phase 초안 (card t1985, lane-16). 운영자 지시(2026-10-27) — 활성 정규 쌍을 CLAUDE.md(공유) + CLAUDE.local.md(로컬/비공개)로 전환 — 를 GEARS 요구사항으로 인코딩. 현행 지도는 HEAD `cb30e05b3e`에서 소스 직독으로 재검증했고, 스코프 맵 `/tmp/moai-factory-analysis-20261027/adk-native-codex-withdrawal-claude-guidance-scope.json`의 `instruction_migration.inventory`(763 match lines / 208 files, 분류 계수이며 삭제 개수가 아님)과 `acceptance_fixture_definitions` GUIDE-01..12(제안이며 미실행)를 본 SPEC의 AC에 조정해 흡수함(GUIDE-09는 t1984 소유로 미채택). Tier M 예산(REQ≤16, AC≤16)에 맞추어 초기 17+17안을 16+16으로 통합 작성. 선행 `SPEC-INSTRUCTION-FILES-UNIFY-001`(completed, AGENTS.md 정규화)의 방향을 반전하는 후속이며, 해당 SPEC의 `superseded` 전환은 run/sync 단계에서 소유자별로 수행한다. |
| 0.1.1 | 2026-10-27 | **Tier M 한도 정합화 (리더 수리 지시와 동일 내용, 처방과 다른 번호 체계로 적용).** 초기 초안 17 REQ/17 AC가 Tier M 상한(각 16)을 초과해, 통합 후 16/16으로 재작성 — 요구사항 통합: 정규화+템플릿(구 004+005→신 004, 마이그레이션 계열 전반 재부번), AC 통합: 정규화·템플릿(→AC-CIF-004), 부분 실패+동시 변경(→AC-CIF-010), doctor·init_resume 프로브 합치(→AC-CIF-016). 리더 처방은 비연속 ID 잔존안이었으나 감사 통과 전 초안이라 연속 재부번을 택했고 RED-now 기준선과 GUIDE 조정표는 현재 ID로 전부 정합. `maps REQ-…` 커버리지 형식 전환 및 "CLAUDEE" 오타 수정 포함. `moai spec lint` 무결("✓ No findings") 관측. |
| 0.2.0 | 2026-10-27 | **감사 iter1 FAIL(0.85, MP-8 + codex 필수 게이트) 수리 + 순서 전환(리더 §15.124 확정).** Part A(순서): t1984 흡수 게이트 삭제 — t1985가 plan→run→sync를 merge-ready까지 먼저 수행(직렬 슬롯 해제), t1984의 임대·plan은 t1985 merge-ready 뒤, t1985 PR 병합·main 흡수가 t1984 RUN에 선행. run 진입 게이트 = iter2+ 독립 감사 PASS(최신 plan HEAD에 결속) + 리더 킥오프 판정뿐. 겹침(harness_fs.go·install.go·AGENTS.md.tmpl·docs)은 순서-겹침이지 의존 아님; 감사 launcher/RPC 3파일은 instruction-file 참조 0건(레인·리더 양측 측정)이므로 t1984 소유·본 카드 비편집 원칙만 유지. Part B(D1~D4): D1 — AC-CIF-007/009/010/011 RED-now 셀 신설(§D.1 행 7~10, 방향/이름 축으로 red 사유 명시), AC-CIF-006 포인터 명시; D2 — AC-CIF-001의 불가능한 Then(상수 grep)을 **관측 가능한 동작**(배포 테스트 `TestCodexOnlyForceUpdateVariant` 계열의 CLAUDE.md 재타겟 + init·리더 경로 실행 관측)으로 교체(§D.1 행 1을 행동 기반 RED로 재기록 — 오늘 그 테스트는 AGENTS.md 배치를 단언하며 PASS, plan M1이 재타겟 소유; 마커 토큰안은 산문 지정 규칙에 가까워 리더 당부로 폐기); D3 — REQ-CIF-010 확장: 최종 remove 직전 원본 mtime/내용 재검증(변경 시 중단·보존), AC-CIF-010 fixture (b)에 write 후/remove 전 변경 주입 추가(atomic-rename 처방은 프리미티브 재사용 제약 위반이라 미채택); D4 — AC-CIF-016 Given/When에 `moai update` 레거시 fixture 실행 포함(보존·충돌 순서 단언), 예산 16/16 유지. 선택 D5(REQ-CIF-012/015 정규 GEARS 전환)·D6(REQ-CIF-003 레거시 폴백 라벨 토큰 `legacy AGENTS fallback` 명명·AC-CIF-003 Then에 고정) 적용, D7은 Part A로 대체(흡수 문구 삭제). |
| 0.2.2 | 2026-10-27 | **감사 iter2 FAIL(0.94, D1~D7 전건 RESOLVED 확인·MP 전 PASS) 델타 수리 — N1·N2(+선택 N3).** N1(critical): progress.md §E.1 `run_gate:`가 삭제된 게이트("t1984 착지 + 리더 해제 전까지 run 금지")를 그대로 인용해 새 순서와 순환 대기를 이뤘음 — run 진입 게이트(iter2+ 독립 감사 PASS + 리더 킥오프 판정)와 t1984 선행 순서 서술로 1행 교체. N2(major): REQ-CIF-010의 "최종 remove 직전 재검증"은 TOCTOU 창을 줄일 뿐 닫지 못함(재검사→remove 사이 편집 소실) — 최종 단계를 재검사+`os.Remove` 대신 **단일 원자적 `os.Rename`(원본→신규 타임스탬프 백업 경로)**로 교체, 그 순간의 바이트가 구성상 백업에 남고 모든 출구에서 최신 콘텐츠가 정규 경로 또는 백업에 존재. 신규 메커니즘 아님(백업 디렉터리·writeExclusive·rollback은 기존 프리미티브, `os.Rename`은 stdlib; 기각된 전면 atomic-rename 재설계와 별개). AC-CIF-010은 기존 창 fixture가 사라지므로 채택 메커니즘에서 재유도: rename 거부(EXDEV)·읽기→쓰기 창·쓰기 실패 주입 + 무손실 불변 단언. plan.md M2 동일 전환, N3(선택): M1에 `embed_manifest_gen.go:647` embed 지시자 이름 갱신(재생성) 명시. |
| 0.2.3 | 2026-10-27 | **감사 iter3 5앵커 문구 수리 — N4·N5·N6 (운영자 지시로 (b) 인계-부채 안 대신 (a) 경량 문구 패스 채택).** N4(완료 판정 정합): rename 채택이 REQ-CIF-010에서 멈춰 REQ-CIF-011("final remove succeeded")·REQ-CIF-006/AC-CIF-006(remove 문구)이 메커니즘과 어긋나 이행자의 완료 판정을 충족불능으로 만들던 것을 세 앵커 모두 최종 원자적 rename 완료 판정으로 정렬. N5(출구 열거): REQ-CIF-010 무손실 불변에 "또는 미변경 레거시 경로" 추가 — 읽기→쓰기 가드 중단·백업 실패 출구에서 최신 콘텐츠는 레거시 경로에 그대로 있으므로 3 출구 계층 전부를 열거. N6(편집된 목적지 보존): 목적지 생성→사용자 편집→최종 rename 거부→rollback-remove가 편집을 지우는 창을 봉쇄 — REQ-CIF-011의 rollback은 편집된 목적지를 삭제가 아니라 백업 엔트리로 rename 보존(최종 단계와 동일 프리미티브 패턴), AC-CIF-010에 (d) 동시 주입 fixture 신설. 변경은 5앵커(REQ-CIF-006/010/011, AC-CIF-006/010)+버전/HISTORY로 국한. |
| 0.2.4 | 2026-10-27 | **감사 iter4 FAIL(0.94, N4/N5/N6 전건 RESOLVED 확인) 수리 — N7·N8 (레인 판정 (a): hunk 패스 1회 + reread 델타).** N7(major): v0.2.3의 편집/미편집 목적지 분류 rollback 자체가 check-then-act(분류 통과→사용자 저장→rollback-remove가 편집 삭제 — N2가 최종 단계에서 제거한 경쟁을 rollback 지점에 재도입) — REQ-CIF-011을 **무조건 형태로 단순화**: rollback은 분류 없이 부분 목적지를 백업으로 원자적 rename 보존(최종 단계와 동일 프리미티브), 부분 결과 보고는 보존된 사본에서 작성. AC-CIF-010 (a)/(d)를 무조건 형태에서 재유도 — (a)는 상태 무관 백업 엔트리 보존+사본 기반 보고, (d)는 EXDEV+편집 동시 주입 유지하되 무조건 보존으로 어떤 상태에서나 성립. N8(major): AC-CIF-005/007의 Given이 `.local.md` 쌍만 시험하고 "또는"이 공유 쌍 생략을 허용(로컬 충돌만 거부하고 공유 충돌을 덮어쓰는 mutant가 양 AC 통과 — codex가 실제 mutant 구축·실행) — 두 AC의 Given을 **both pairs 필수**(공유 CLAUDE.md/AGENTS.md + 로컬 CLAUDE.local.md/AGENTS.local.md, 테이블 주도)로 교체, Then에 쌍별 단언. 변경은 3앵커(REQ-CIF-011, AC-CIF-010, AC-CIF-005/007)+버전/HISTORY로 국한. |
| 0.2.5 | 2026-10-27 | **감사 iter5 FAIL(0.94 보합, N7/N8 RESOLVED 확인) 최종 통합 패스 — N9·N10·N11 (감사자 예고: 6라운드의 더 얇은 결함은 run 부채로 전환되므로 plan 산출물 라운드는 이번이 마지막).** N9(major): AC-CIF-010 fixture (c)(쓰기 실패)의 Then에 (a)와 동일한 무조건 보존 단언 부착 — 부분 목적지는 상태 무관 백업 엔트리로 원자적 rename 보존(해시 대조)·보고는 보존된 사본에서; "rename 실패 시 보존, 쓰기 실패 시 삭제" mutant가 이 행에 실패 (mutant-probe HARD). N10(major): REQ-CIF-011에 rollback rename 자체 실패 출구 신설 — "When the rollback rename itself fails, the migration shall leave the original and the partial destination untouched in place, exit non-zero, and report the exact residual paths and the failed preservation." (미정의 출구의 불가능 클래스 봉쇄; 무손실은 유지 — 양쪽 내용 생존, 동사의 완료 계약만 정의) + AC-CIF-010에 (e) fixture(백업 디렉터리 쓰기 불가 EACCES 주입 → 원본·부분 목적지 그 자리, exit non-zero, 잔여 경로·보존 실패 보고) 신설. N11(SHOULD-FIX): plan.md M2 마감 단언을 3-위치 불변("정규 경로, 백업, 또는 미변경 레거시 경로")으로 정합. 변경은 3앵커(AC-CIF-010, REQ-CIF-011, plan.md M2 gloss)+버전/HISTORY로 국한. |

## §A Context

본 저장소의 현행 제품은 **AGENTS.md-정규화** 방향이다(`SPEC-INSTRUCTION-FILES-UNIFY-001`가 2026-09-27에 완료). 운영자 지시(2026-10-27)는 이 표면의 방향을 반전하여 **활성 정규 쌍을 CLAUDE.md(공유) + CLAUDE.local.md(로컬/비공개)로** 만들고, AGENTS.md / AGENTS.local.md의 모든 생성·탐색·참조·설치·갱신·마이그레이션·문서·템플릿·테스트 경로를 전환하거나 폐기하라는 것이다. 단 **사용자 콘텐츠는 무손실**로 보존되어야 한다 — 충돌 감지 뒤에만 이동하고, 눈먼 전역 치환·덮어쓰기·비공개 콘텐츠의 공유 파일 승격은 금지된다.

### 현행 지도 (HEAD `cb30e05b3e`에서 소스 직독, 2026-10-27)

| 위치 | 관측된 사실 |
|---|---|
| `internal/defs/files.go` | `ClaudeMD = "CLAUDE.md"`는 retired 주석(레거시 탐지·frozen 계약용으로만 생존), `AgentsMD = "AGENTS.md"`(line 21)이 정규 상수 |
| `internal/contract/frozen.go:10` | `FrozenInstructionFiles = ["CLAUDE.md","CLAUDE.local.md","AGENTS.md","AGENTS.local.md"]` — 4개 이름 모두 frozen. **이 덮개는 이관 중/후에도 유지한다** |
| `internal/hook/pre_tool.go:1678` | `frozenInstructionFiles` 동일 4이름 — harness-learner write-deny sentinel |
| `internal/core/project/initializer.go:303-306` | init은 CLAUDE.md를 만들지 않고 AGENTS.md를 템플릿 배포자가 투영 |
| `internal/core/project/validator.go:140-145,211-212` | AGENTS.md 존재 검사, 레거시 CLAUDE.md는 불활성 |
| `internal/cli/init_resume.go:222-230` | 레이아웃 프로브가 `.claude/` + AGENTS.md 사용 (`deployedWiringIsGPT`) |
| `internal/cli/update/plan/plan.go:26-28,46` | `highRiskFiles`가 AGENTS.md/CLAUDE.md 둘 다 나열, 병합 전략도 둘 다 — 4-이름 리스크 분류는 유지 |
| `internal/template/harness_fs.go:197` | `normalizeHarnessReferences`가 `"CLAUDE.md"` → `"AGENTS.md"`로 재작성 (line 197에서 `grep` 관측; `:181` `NormalizeCodexRoleForDeploy`와 `normalizedOpen` 경유 호출). 방향 반전 대상 |
| `internal/template/templates/AGENTS.md.tmpl` | 정규 공유 템플릿 (CLAUDE.md.tmpl은 부재 — `ls` exit 1 관측). CLAUDE.md 투영으로 재명명/재투영한다 |
| `internal/cli/migrate_local_instructions.go` (205 lines) | `moai migrate local-instructions`가 CLAUDE.local.md → AGENTS.local.md 이동 (**요청과 반대 방향**). 재사용할 검증된 프리미티브: 공재거부 무기록(:57-59, REQ-IFU-010b 선례), `os.Lstat`+`!IsRegular` 거부(:73-75), 타임스탬프 백업 디렉터리 `.moai/backups/local-instructions/`, `writeExclusive`(O_EXCL), 최종 remove 실패 시 신규 파일 rollback-remove(:90-97), 검증 완료 후 재실행 noop |
| 런타임 리더 (AGENTS-우선 + 레거시 CLAUDE 폴백) | `internal/cli/doctor_harness.go:84-89`, `internal/hook/instructions_loaded.go:79-80,145`(`resolveInstructionFile`, @import 폐쇄), `internal/harness/curator/dispatch.go:41-43`(Tier3→AGENTS.local.md append-only, REQ-IFU-014; Tier4→AGENTS.md digest), `internal/harness/layer5.go:68,115,194-208`(AGENTS @import 문자열), `internal/config/token_budget_guard.go`(측정 표면 고정 슬롯), `internal/cli/harness/install.go`(CLAUDE.md 마커 블록 명명 — 주석 직독 확인) |
| MCP 감사 입력 | `internal/cli/codex_audit_launch.go:182-222`는 `role.instructions`를 `developer_instructions`로 주입하며 CLAUDE/AGENTS 파일 로드 경로가 없음. 신규 guidance 입력은 card t1984 조정 대상(§D Out of Scope) |
| 이 저장소 자체 디스크 상태 | AGENTS.md (22,511 B, regular, 사용자 편집) + AGENTS.local.md (53,522 B, regular, 사용자 소유) 존재, CLAUDE 파일 부재 (`wc -c`/`ls` 관측) |

규모(스코프 맵 inventory, 분류 계수): 94 code + 60 template + 251 instructions/docs + 242 tests + 116 historical/data-fixture = 208 files / 763 match lines. 이 중 Go 파일 81개, template 트리 18개 파일이 `AGENTS.md`를 참조한다(본 워크트리 `grep -rl` 관측).

### ASSUMPTIONS

1. **t1984와의 순서(의존 아님)**: 리더 확정(§15.124) 순서 — t1985가 plan→run→sync를 merge-ready까지 먼저 수행해 직렬 슬롯을 해제하고, t1984(네이티브 Codex 철수)의 임대·plan은 t1985 merge-ready 뒤에 시작하며, t1985의 PR 병합·main 흡수가 t1984 RUN에 선행한다. `harness_fs.go`·`install.go`·`AGENTS.md.tmpl`·docs 겹침은 **순서-겹침**이지 의존이 아니다. 감사 launcher/RPC 3파일(`codex_audit_launch.go` / `codex_audit_mcp.go` / `mcp_codex_audit.go`)은 instruction-file 참조 0건(레인·리더 양측 grep 측정)으로 본 SPEC과 의미상 무관하다 — t1984 소유·본 카드 비편집 원칙만 유지된다.
2. **레거시 읽기 폴백 유지**: 방향 반전은 현재 제품이 레거시 CLAUDE.md에 대해 하던 것과 대칭으로, 레거시 AGENTS.md를 읽기 폴백으로 유지한다(미이관 저장소의 사용자 콘텐츠 보존). 폴백은 보존-보고이며 자동 해소하지 않는다.
3. **frozen 계약 불변**: 4-이름 frozen 덮개는 이관 중에도 유지되며, 본 SPEC은 이를 줄이지 않는다.
4. **분류 불확실성의 기본 처지**: 공유/비공개 모호 콘텐츠는 미해결로 보존되고 로컬에 남는다 — 자동 게시 금지.
5. **스코프 맵의 GUIDE-01..12는 제안**: 미실행 fixture다. 본 SPEC은 이들을 AC에 조정해 흡수하되, 선행 85-fixture audit-only 집계에는 개수를 더하지 않는다(supersedes 조항).

## §B Goal

활성 instruction-file 표면이 CLAUDE.md(공유) + CLAUDE.local.md(로컬/비공개) 단일 정규 쌍으로 전환되고, 기존 AGENTS 사용자 콘텐츠가 충돌 감지 뒤 무손실(바이트 정확 백업)로 이동하며, 프라이버시 분리(비공개 내용은 공유 파일·로그에 못 들어감)와 frozen 계약 덮개가 이관 전·중·후 내내 유지되는 것.

## §C Requirements (GEARS)

도메인 토큰: **CIF** (CLAUDE Instruction Files). Tier M 예산 준수 — REQ 16건.

### C.1 정규 쌍과 생성 경로

- **REQ-CIF-001** — The system shall treat `CLAUDE.md` as the sole shared active instruction file and `CLAUDE.local.md` as the sole local/private active instruction file, in every creation, navigation, reference, install, update, migration, documentation, template, and test path of the product.
- **REQ-CIF-002** — When `moai init` scaffolds a project for any `--llm` value, the system shall deploy `CLAUDE.md` from the canonical shared template and shall not create `AGENTS.md` or `AGENTS.local.md`.
- **REQ-CIF-003** — Where a runtime reader resolves the project's instruction file(s), the reader shall prefer the canonical CLAUDE pair and shall read a legacy AGENTS file only as an explicit legacy fallback labelled with the literal token `legacy AGENTS fallback` that modifies nothing.

### C.2 참조 정규화와 템플릿

- **REQ-CIF-004** — Where the product names, rewrites, or generates instruction-file references (`internal/template/harness_fs.go` normalization and the canonical template itself), the reference and the deployed template shall resolve to the canonical CLAUDE names (`CLAUDE.md.tmpl` 투영), and newly generated or deployed content shall contain no reference that creates or directs a user to an AGENTS file as an active surface.

### C.3 마이그레이션 동사 — 무손실 방향 전환

- **REQ-CIF-005** — When the local-instructions migration verb runs on a repository whose canonical destination file is absent and whose legacy AGENTS counterpart exists, the system shall move the legacy file to the canonical name (AGENTS.local.md → CLAUDE.local.md, 그리고 shared AGENTS.md → CLAUDE.md 경로) by reusing the existing coexistence-refusal, `os.Lstat` regular-file check, timestamped backup, exclusive write (`O_EXCL`), and rollback-remove primitives; the system shall not introduce a parallel migration mechanism, a new state ledger, or a new product subsystem for this.
- **REQ-CIF-006** — The migration shall record a byte-exact backup of each moved file — content and permission bits — under a fresh timestamped backup directory before the destination write, and shall keep the original in place until the final atomic-rename step (the original's move into the backup) succeeds.
- **REQ-CIF-007** — When the canonical file and its legacy counterpart both exist and their contents differ, the migration shall refuse without modifying either file and shall report both paths with the hand-merge instruction; the system shall not overwrite either file.
- **REQ-CIF-008** — Where a migration has already verifiably completed (canonical present, legacy absent), re-running the migration verb shall be a noop that writes no file and creates no second backup.
- **REQ-CIF-009** — The migration shall refuse symlinked, non-regular, unreadable, and path-escaping inputs without following, resolving, or overwriting the target.
- **REQ-CIF-010** — When the source file's content or modification time changes between the migration's read and its write, the migration shall abort before the destination write and shall leave both files exactly as they were; and the migration's final step shall move the original into a fresh timestamped backup path by a single atomic `os.Rename` — not a re-check-then-remove pair — so that whatever bytes exist at that instant land in the backup by construction and, at every exit (success, refusal, or failure), the user's latest content exists at the canonical path, in the backup, or at the untouched legacy path — never nowhere.
- **REQ-CIF-011** — When the backup, the exclusive write, or the final atomic-rename step fails, the migration shall roll back losslessly by renaming the partial destination into the backup — unconditionally, with no prior untouched-vs-edited classification (the same atomic primitive as the final step; any classification before the preserve would re-introduce the check-then-act window the final step removed) — and shall report the exact partial outcome from the preserved backup entry; the verb shall not report a migrated result unless the final step (the atomic rename) succeeded. When the rollback rename itself fails, the migration shall leave the original and the partial destination untouched in place, exit non-zero, and report the exact residual paths and the failed preservation.

### C.4 프라이버시, 분류, 학습 표면

- **REQ-CIF-012** — While the migration or any instruction-file transition is handling instruction-file content, the system shall keep instruction-file bodies out of logs and reports, shall keep private and local content out of the shared `CLAUDE.md`, and shall preserve ambiguous shared/private classification as unresolved, local, and reported — never auto-published.
- **REQ-CIF-013** — The harness curator's Tier-3 learning surface shall target `CLAUDE.local.md` (append-only) and its Tier-4 surface the `CLAUDE.md` digest, and the always-loaded token-budget guard's measured surface shall follow the canonical CLAUDE names.

### C.5 계약과 역사 보존

- **REQ-CIF-014** — The frozen instruction-file contract shall keep covering all four names — `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `AGENTS.local.md` — in `internal/contract/frozen.go` and `internal/hook/pre_tool.go`, and the pinning test shall keep the two lists equal, throughout and after the migration.
- **REQ-CIF-015** — While the migration or the repository's own transition is running, the system shall preserve historical AGENTS backups (`.moai/backups/local-instructions/**`) and past audit evidence (`.moai/reports/**`) unchanged.

### C.6 자체 저장소 전환과 룰 표면

- **REQ-CIF-016** — This repository's own instruction surface (`AGENTS.md` 22,511 B 사용자 편집본, `AGENTS.local.md` 53,522 B 사용자 소유본) shall transition through the same migration-verb semantics with conflict detection — no blind substitution — and the distributed rule files and their template mirror shall stop directing agents to AGENTS files as active instruction surfaces.

## §D Out of Scope

### Out of Scope — MCP 감사 launcher의 CLAUDE guidance 입력 경로 (GUIDE-09)

GUIDE-09(신규 bounded explicit guidance 입력)는 **t1984 단일 담당(solo ownership)이며, 구체 AC는 t1984 plan에서 확정한다**. 본 SPEC은 이 파일 3개(`internal/cli/codex_audit_launch.go`·`codex_audit_mcp.go`·`mcp_codex_audit.go`)를 편집하지 않는다(동시 쓰기 금지), 그리고 명세되지 않은 CLI config 옵션을 발명하지 않는다.

- 신규 MCP 감사 guidance 주입 경로 설계·구현
- 감사 launcher/RPC 파일의 동시 편집 (t1984와 절대 병행 금지)

### Out of Scope — 네이티브 Codex 철수 자체

- ADK-네이티브 Codex 진입점·설치·템플릿·문서의 철수 (t1984가 소유; 본 SPEC은 instruction-file 방향만 담당)

### Out of Scope — 신규 메커니즘

- 병렬 마이그레이션 메커니즘, 신규 상태 원장(ledger), 방향 전환에 필요한 것 이외의 신규 하위시스템
- AGENTS 파일의 자동 내용 병합/자동 분류 (모호 콘텐츠는 미해결 보존 — 수동 병합만)

### Out of Scope — 역사·외부 자산

- `.moai/backups/local-instructions/**` 및 `.moai/reports/**` 재작성·삭제
- 사용자 Codex 애플리케이션, 무관 저장소/제품, 실행 중 프로세스/세션 종료
- 스코프 맵 GUIDE fixture의 선행 85-fixture audit-only 집계에 개수를 더하는 행위

## §E Cross-references

- 선행: `SPEC-INSTRUCTION-FILES-UNIFY-001` (AGENTS.md 정규화 — 본 SPEC이 방향 반전; 완료 시 superseded 전환은 run/sync 소유), `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001` (REQ-IFU-007~012/020~022의 원 소유), `SPEC-AGENTS-MD-CANON-001`
- 순서(의존 아님): card t1984 (lane-15) — t1985 merge-ready가 t1984 임대·plan에 선행, t1985 PR 병합·main 흡수가 t1984 RUN에 선행 (리더 §15.124 확정; plan.md §C Pre-flight)
- 스코프 맵: `/tmp/moai-factory-analysis-20261027/adk-native-codex-withdrawal-claude-guidance-scope.json` (operator-sanctioned, 2026-10-27; GUIDE-01..12는 제안 → acceptance.md에서 조정)
- 계약: `internal/contract/frozen.go` / `internal/hook/pre_tool.go` frozen instruction-files invariant
- 산출물: `plan.md`, `acceptance.md`, `progress.md` (같은 디렉터리)
