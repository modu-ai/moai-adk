# acceptance.md — SPEC-CLAUDE-INSTRUCTION-FILES-001

> stateless 산출물 (status 축 무상태). 검증 계층 — 모든 항목은 `AC-CIF-XXX` `Given/When/Then` 형식이며 이진 판정 가능해야 한다. Tier M 예산 준수 — AC 16건. **정적 확인(grep/ls)과 실행 검증(go test)을 모든 AC에서 구분하고, 미실행은 결코 PASS가 아니다.**

## §D.1 채택 규칙 (2-셀)

- 릴리스-블로킹 AC는 **RED-now 셀**(단일 호출 읽기 전용 명령 + 그 verbatim stdout + exit code + 트리 SHA)과 **green path 셀**(어느 마일스톤이 뒤집는지 + 통과 시 출력)을 쌍으로 갖는다(verification-completeness §2).
- 문서 수준 고정: 아래 표의 측정 트리는 **`cb30e05b3e`** (2026-10-27, 본 워크트리; 제품 트리는 감사 시점 HEAD `376a0a5cf0`에서 `cb30e05b3e`와 바이트 동일 — 감사 판정문 관측)이며, 이 고정은 자체 핀을 가지지 않는 모든 RED-now 셀을 묶는다. run 진입 트리에서는 재측정·재고정한다(plan.md §C/M0).
- RED는 "올바른 이유로 red"여야 한다 — 각 측정은 현행 트리가 요구사항과 반대 방향(AGENTS-정규)이거나 요구되는 표면이 부재함을 보이는 positive 관측이다. 방향 전환 대상 동사의 기존 거부 프리미티브(공재거부·Lstat)는 **구 방향**(CLAUDE→AGENTS 이동)에만 묶여 있으므로, red 사유는 프리미티브 부재가 아니라 **방향/이름 축**(정규 목적지 CLAUDE.local.md/CLAUDE.md를 향한 경로가 오늘 존재하지 않음)에 명시한다(D1).

### RED-now 기준선 (측정 트리 `cb30e05b3e`, 2026-10-27)

| # | 명령 (단일 호출) | verbatim stdout | exit |
|---|---|---|---|
| 1 | `go test ./internal/template/ -run '^TestCodexOnlyForceUpdateVariant$' -v` | `=== RUN   TestCodexOnlyForceUpdateVariant` ⏎ `--- PASS: TestCodexOnlyForceUpdateVariant (0.16s)` ⏎ `PASS` ⏎ `ok  	github.com/modu-ai/moai-adk/internal/template	0.460s` | 0 |
| 2 | `grep -n 'ReplaceAll(text, "CLAUDE.md", "AGENTS.md")' internal/template/harness_fs.go` | `197:	text = strings.ReplaceAll(text, "CLAUDE.md", "AGENTS.md")` | 0 |
| 3 | `grep -n 'Move CLAUDE.local.md to AGENTS.local.md' internal/cli/migrate_local_instructions.go` | `27:		Short: "Move CLAUDE.local.md to AGENTS.local.md, keeping a backup",` | 0 |
| 4 | `grep -n 'Path: "AGENTS.local.md", BlockType: BlockTypeLearnedLocal' internal/harness/curator/dispatch.go` | `45:	3: {Path: "AGENTS.local.md", BlockType: BlockTypeLearnedLocal, Tier: 3},` | 0 |
| 5 | `ls internal/template/templates/CLAUDE.md.tmpl` | `ls: internal/template/templates/CLAUDE.md.tmpl: No such file or directory` | 1 |
| 6 | `grep -c 'AGENTS.md' internal/cli/doctor_harness.go internal/hook/instructions_loaded.go internal/cli/init_resume.go` | `internal/hook/instructions_loaded.go:9` `internal/cli/init_resume.go:4` `internal/cli/doctor_harness.go:3` | 0 |
| 7 | `grep -n 'both exist' internal/cli/migrate_local_instructions.go` | `58:		return fmt.Errorf("refusing to migrate: %s and %s both exist; merge %s into %s by hand, then remove %s",` ⏎ `177:		return fmt.Sprintf("Advisory: %s and %s both exist; merge %s into %s by hand and remove it "` ⏎ `178:			"(`moai migrate local-instructions` refuses while both exist).",` | 0 |
| 8 | `grep -n 'os.Lstat(claudePath)' internal/cli/migrate_local_instructions.go` | `68:	info, err := os.Lstat(claudePath)` | 0 |
| 9 | `grep -c 'ModTime' internal/cli/migrate_local_instructions.go` | `0` | 1 |
| 10 | `grep -c '"AGENTS.md"' internal/cli/migrate_local_instructions.go` | `0` | 1 |

읽는 법: **행 1은 행동 기반 RED다** — `TestCodexOnlyForceUpdateVariant`는 오늘 AGENTS.md 배포를 단언하는 테스트(harness_fs_test.go:132-133)로 `--- PASS`가 관측되며, 이는 배포 동작이 AGENTS-정규임을 행동으로 입증하므로 CLAUDE-정규 기준으로 red다(산문 판정 규칙 없음, 리더 당부 1). 행 2·3·4·6·7·8(exit 0)은 현행 트리가 해당 요구사항과 **반대 방향**임 — 매치된 코드가 구 방향(AGENTS 목적지·CLAUDE 원본·AGENTS-우선)을 가리킨다. 행 5·9·10(exit 1)은 요구되는 표면 자체가 부재(정규 템플릿 없음·동시변경 검증 없음·공유 파일 이관 경로 없음). 특히 행 7·8의 거부 프리미티브는 구 방향의 목적지(`codexLocalInstructionName` = AGENTS.local.md)만 보호하므로, 정규 목적지 CLAUDE.local.md를 향한 충돌 거부·비정규 거부는 오늘 시험 대상으로 존재하지 않는다 — 이것이 AC-CIF-007/009의 red 사유다(D1).

## §D.2 AC 목록

### AC-CIF-001 — 정규 방향의 동작 입증 (maps REQ-CIF-001, GUIDE-01) [실행, 릴리스-블로킹]
- **Given** 제품 소스 트리와 배포·리더 경로 **When** 배포 경로(`TestCodexOnlyForceUpdateVariant` 계열)와 init 통합 테스트, 리더 해석 경로(`resolveInstructionFile` 계열) 실행 **Then** 관측 가능한 동작으로 역할 반전이 입증된다 — 배포 테스트가 CLAUDE.md 배치를 단언하고(AGENTS 파일 부재), init이 CLAUDE.md를 생성하며, 리더가 CLAUDE.md를 우선 해석한다; 모두 `--- PASS:`로 관측. 두 상수 `ClaudeMD`/`AgentsMD`는 모두 유지된 채 주석·역할만 반전된다(REQ-CIF-003 레거시 폴백·REQ-CIF-014 frozen 계약이 AGENTS.md 이름의 생존을 요구 — 상수·주석 텍스트 자체는 측정 대상이 아니다).
- RED-now: §D.1 표 1행(행동 기반) — 오늘 그 배포 테스트는 AGENTS.md 배치를 단언하며 PASS하므로 CLAUDE-정규 기준으로 red. Green path: M1 — 배포 주장이 CLAUDE.md로 재타겟되어 동일 테스트가 PASS + init 통합 테스트 PASS(AC-CIF-002와 별개 측정). (v0.1.x의 상수-삭제 grep은 계획 준수 구현에서 영구 exit 0인 불가능 기준이었고, v0.2.0 초안의 마커 토큰 grep은 산문 지정 규칙에 가까워 폐기 — 감사 D2 + 리더 당부 1: 동작으로 입증.)

### AC-CIF-002 — init 생성 방향 (maps REQ-CIF-002, GUIDE-01) [실행, 릴리스-블로킹]
- **Given** 임시 프로젝트 디렉터리(t.TempDir) **When** `moai init --llm claude`(및 codex/both) 실행 **Then** CLAUDE.md가 생성되고 AGENTS.md·AGENTS.local.md가 존재하지 않는다(테스트가 os.Stat으로 부재 관측).
- RED-now: §D.1 표 3행(생성 방향의 원인 동사가 구 방향). Green path: M1 — init 통합 테스트 PASS.

### AC-CIF-003 — 런타임 리더 폴백 방향 (maps REQ-CIF-003, GUIDE-08) [실행]
- **Given** (a) CLAUDE.md와 AGENTS.md가 모두 있는 fixture (b) CLAUDE.md만 있는 fixture (c) AGENTS.md만 있는 fixture **When** 리더(`resolveInstructionFile` 등)가 명령 파일을 해석 **Then** (a) CLAUDE.md 우선, (c) 폴백 경로 진입이 리터럴 라벨 토큰 `legacy AGENTS fallback` 포함 출력으로 관측되며, 모든 경우 어느 파일도 수정하지 않는다(사전/사후 해시 동일).
- RED-now: §D.1 표 6행(현재 AGENTS-우선 — 라벨 토큰·폴백 방향 모두 부재). Green path: M3.

### AC-CIF-004 — 참조 정규화·정규 템플릿 (maps REQ-CIF-004, GUIDE-01/10) [실행+정적, 릴리스-블로킹]
- **Given** 정규화 대상 템플릿 바이트와 template 트리 **When** `normalizeHarnessReferences` 적용 및 `ls internal/template/templates/CLAUDE.md.tmpl` 실행 **Then** 결과에 `"AGENTS.md"`/`"AGENTS.local.md"` 활성 참조가 0건이고 CLAUDE.md.tmpl이 존재한다(exit 0), 제품 생성 표면에 활성 AGENTS 참조가 없다.
- RED-now: §D.1 표 2행 + 표 5행. Green path: M1 — 정규화 단위 테스트 PASS + ls exit 0.

### AC-CIF-005 — 마이그레이션 방향 반전·프리미티브 재사용 (maps REQ-CIF-005, GUIDE-02/03) [실행, 릴리스-블로킹]
- **Given** AGENTS.local.md만 있는 임시 저장소와 AGENTS.md만 있는 임시 저장소 **각각**(두 이관 경로 모두 필수 — 테이블 주도, 어느 한쪽도 생략 불가) **When** 마이그레이션 verb 실행 **Then** 각각 CLAUDE.local.md·CLAUDE.md로 이동하고(쌍별 단언) 기존 프리미티브(공재 거부·Lstat 검사·타임스탬프 백업·O_EXCL·rollback)를 재사용한다(신규 메커니즘 부재 — 코드 검증 정적 확인).
- RED-now: §D.1 표 3행(현재 반대 방향). Green path: M2.

### AC-CIF-006 — 무손실 바이트 정확 백업 (maps REQ-CIF-006, GUIDE-02) [실행, 릴리스-블로킹]
- **Given** 알려진 바이트 내용과 권한(0600 등)의 AGENTS.local.md **When** 마이그레이션 실행 **Then** 백업 복사본이 원본과 바이트 동일(`cmp` exit 0)하고 권한 비트가 보존되며 최종 원자적 rename(원본→백업 경로) 성공 전까지 원본이 그 자리에 남는다.
- RED-now: §D.1 표 3행(방향이 반대라 이 동작 자체가 존재하지 않음 — AC-CIF-005와 동일 원인). Green path: M2.

### AC-CIF-007 — 공재·상이 거부 (maps REQ-CIF-007, GUIDE-04) [실행, 릴리스-블로킹]
- **Given** CLAUDE.local.md+AGENTS.local.md 쌍과 CLAUDE.md+AGENTS.md 쌍 **각각**이 모두 존재하고 내용이 다른 fixture(both pairs 필수 — 테이블 주도; 로컬 쌍만 거부하고 공유 쌍을 덮어쓰는 mutant는 이 AC에 실패해야 한다, N8) **When** 마이그레이션 실행 **Then** **쌍별로** exit non-zero, 양쪽 파일 모두 변경 없음(사전/사후 해시 동일 — 공유 쌍 포함), 오류 메시지가 두 경로와 수동 병합 지시를 보고.
- RED-now: §D.1 표 7행 — 공재 거부 프리미티브는 이미 존재하나 **구 방향의 목적지**(AGENTS.local.md로의 병합)만 보호한다; 정규 목적지 CLAUDE.local.md를 향한 충돌 거부는 오늘 시험 대상으로 존재하지 않는다(red 사유는 방향/이름 축, D1). Green path: M2 — 거부 테이블 테스트.

### AC-CIF-008 — 멱등 재실행 noop (maps REQ-CIF-008, GUIDE-07) [실행]
- **Given** 이미 완료된 이동(정규 있음·레거시 부재) **When** verb 재실행 **Then** 파일·백업 디렉터리에 아무 쓰기 없음(백업 디렉터리 수 불변 관측), noop 메시지 출력.
- Green path: M2 (RED: §D.1 표 3행과 동일 원인 — 정규 방향 완료 상태를 만드는 경로 자체가 부재).

### AC-CIF-009 — 비정상 입력 거부 (maps REQ-CIF-009, GUIDE-05) [실행, 릴리스-블로킹]
- **Given** symlink·FIFO·읽기불가(0000)·경로이탈 입력 fixture **When** 마이그레이션 실행 **Then** 각각 exit non-zero로 거부하고 링크 대상·원본을 따라가거나 덮어쓰지 않는다(Lstat 결과 관측).
- RED-now: §D.1 표 8행 — Lstat 비정규 검사는 **구 원본 경로**(`claudePath` = CLAUDE.local.md)에만 묶여 있다; 정규 방향의 원본(AGENTS.local.md/AGENTS.md)에 대한 거부 경로는 오늘 존재하지 않는다(red 사유는 방향/이름 축, D1). Green path: M2 — 거부 테이블.

### AC-CIF-010 — 원자적 최종 단계·무손실 불변 (maps REQ-CIF-010, REQ-CIF-011; GUIDE-06/11) [실행, 릴리스-블로킹]
- **Given** (a) 최종 단계 거부 주입 fixture — `os.Rename`(원본→신규 타임스탬프 백업 경로)이 EXDEV 등으로 거부 (b-1) 읽기와 쓰기 사이 원본 mtime/내용 변경 fixture (c) 목적지 쓰기(`writeExclusive`) 실패 주입 fixture (d) **목적지 생성 뒤 사용자가 목적지를 편집 + 최종 rename 거부(EXDEV) 동시 주입** fixture (e) **rollback rename 자체 실패 주입 fixture — 백업 디렉터리 쓰기 불가(EACCES 등)로 최종 rename과 rollback rename이 모두 거부** **When** 마이그레이션 실행 **Then** (a) 안전 중단 — 원본은 제자리, 부분 목적지는 **상태와 무관하게(미편집/편집 분류 없이) 백업 엔트리로 원자적 rename 보존**되고 정확한 부분 결과는 보존된 사본에서 보고되며, "Migrated" 성공 없음·반쪽 상태 없음 (b-1) 목적지 쓰기 전 중단, 양쪽 내용 보존 (c) 원본 불변·그대로, **부분 목적지는 (a)와 동일하게 상태와 무관하게 백업 엔트리로 원자적 rename 보존(해시 대조)**, 정확한 부분 결과는 보존된 사본에서 보고, 반쪽 상태 없음 — 쓰기 실패에서 보존·rename 실패에서 삭제하는 mutant는 이 행에 실패해야 한다(N9) (d) 주입된 목적지 편집 내용이 **백업 엔트리에 그대로 생존** — 무조건 보존이므로 어떤 목적지 상태에서나 성립(REQ-CIF-011, N7) (e) 원본과 부분 목적지가 **그 자리에 그대로**(양쪽 내용 모두 생존 — 무손실 유지), exit non-zero, 보고가 잔여 경로들과 보존 실패 사실을 명명(REQ-CIF-011 rollback 실패 출구, N10). **무손실 불변(모든 출구에서 단언)**: 어느 출구에서든 사용자의 최신 콘텐츠는 정규 경로, 백업, 또는 미변경 레거시 경로에 존재한다 — 최종 단계는 재검사+`os.Remove` 쌍이 아니라 단일 원자적 `os.Rename`이므로 재검사→remove 사이 소실 창(TOCTOU)이 원천적으로 없다(REQ-CIF-010, N2/N5/N6).
- RED-now: §D.1 표 9행 — 읽기→쓰기 가드(mtime/내용 비교)와 원자적 최종 `os.Rename` 어느 쪽 코드도 동사에 전혀 없다(grep 0건, exit 1). Green path: M2 — 실패·거부 주입 테스트(a/b-1/c/d/e) + 정상 이동에서 rename 후 백업 엔트리 존재 관측.

### AC-CIF-011 — 프라이버시·모호 분류 보존 (maps REQ-CIF-012, GUIDE-03/12) [실행+정적, 릴리스-블로킹]
- **Given** 비공개 문구를 포함한 AGENTS.local.md와 모호 공유/비공개 혼합 내용 fixture **When** 마이그레이션 및 자체 저장소 전환 실행 **Then** 비공개 문구가 CLAUDE.md에 0건(grep), 모든 로그·보고 출력에서 본문 미포함(출력 스캔), 모호 내용은 미해결로 보고되고 자동 게시되지 않는다.
- RED-now: §D.1 표 10행 — 공유 파일(AGENTS.md) 이관 경로가 동사에 전혀 없으므로(grep 0건, exit 1) "비공개 내용의 공유 파일 승격 금지"가 시험하는 코드 표면 자체가 부재. Green path: M2/M4.

### AC-CIF-012 — curator·예산 표면 재타겟 (maps REQ-CIF-013, GUIDE-08) [실행, 릴리스-블로킹]
- **Given** curator TierSurfaceMap과 token budget guard **When** 해당 단위 테스트 실행 **Then** Tier-3가 `CLAUDE.local.md`, Tier-4가 `CLAUDE.md` digest를 대상으로 하고 예산 가드 측정 표면이 정규 이름을 따른다.
- RED-now: §D.1 표 4행. Green path: M3.

### AC-CIF-013 — frozen 4-이름 계약 유지 (maps REQ-CIF-014, GUIDE-10) [실행, regression-guard]
- **Given** `internal/contract/frozen.go`와 `internal/hook/pre_tool.go` **When** `TestFrozenInstructionFiles` 계열 실행 **Then** 4-이름 목록이 그대로이고 두 목록이 동일(pin)하다 — 이관 전에도 green이어야 하며 이관 후에도 green이어야 한다(불변 유지 criterion; vacuous 아님 — 목록을 줄이는 mutant에 red).
- Green path: M3에서 동일 PASS 관측(현재도 green).

### AC-CIF-014 — 역사 보존 (maps REQ-CIF-015, GUIDE-10) [정적+실행]
- **Given** `.moai/backups/local-instructions/**` 기존 항목과 `.moai/reports/**` **When** 마이그레이션 및 자체 전환 완료 **Then** 기존 백업·보고 파일의 내용 불변(해시 비교, 사전/사후).
- Green path: M4 (불변 유지 — 현재도 green; 위반 시 red).

### AC-CIF-015 — 자체 저장소 전환 (maps REQ-CIF-016, GUIDE-02/03) [실행]
- **Given** 이 저장소의 AGENTS.md/AGENTS.local.md **When** 마이그레이션 verb 실행(커밋 경계) **Then** CLAUDE.md/CLAUDE.local.md가 존재하고 백업이 확보되었으며 init/doctor가 정규 파일에서 동작함을 실행으로 관측, `.claude/rules/**`·template mirror에 활성 AGENTS 지시가 남지 않았다(grep, historical/data-fixture 인용 제외).
- Green path: M4.

### AC-CIF-016 — 관찰자 합치 + update 경로 (maps REQ-CIF-001, REQ-CIF-003; GUIDE-08) [실행]
- **Given** CLAUDE-정규 프로젝트 fixture, 그리고 사용자 편집 AGENTS.md만 있는 레거시 fixture **When** `moai doctor`·init_resume 프로브·검증기 실행 및 레거시 fixture에서 `moai update` 실행 **Then** L3 instruction-file 검사가 CLAUDE.md를 1차 표적으로 보고, `deployedWiringIsGPT` 계열 프로브가 CLAUDE.md를 사용하며 validator의 존재 검사가 CLAUDE.md를 가리킨다; update는 사용자 편집 AGENTS.md를 보존하고 정규 CLAUDE.md 생성 전에 충돌/보존 순서를 지켜 사용자 지시가 비활성화되는 전환을 일으키지 않는다(D4).
- RED-now: §D.1 표 6행(doctor 3건·init_resume 4건 포함 — update의 관찰자·프로브 전반이 구 방향). Green path: M1/M3(update leg는 M1의 템플릿 투영 전환 뒤 M3 검증).

## §D.3 GUIDE-01..12 조정표 (스코프 맵 제안 ↔ 본 SPEC AC)

| GUIDE | 제안 단언 | 본 SPEC AC | 조정 |
|---|---|---|---|
| GUIDE-01 | init CLAUDE 배출 | AC-CIF-001/002/004 | 그대로 채택 |
| GUIDE-02 | 공유 원본 무손실 이동 | AC-CIF-005/006/015 | 그대로 채택 |
| GUIDE-03 | 로컬은 로컬 유지 | AC-CIF-005/011/015 | 그대로 채택 |
| GUIDE-04 | 상이 충돌 비파괴 거부 | AC-CIF-007 | 그대로 채택 |
| GUIDE-05 | 비정상 입력 거부 | AC-CIF-009 | 그대로 채택 |
| GUIDE-06 | 부분 실패 정확 보고 | AC-CIF-010 | 그대로 채택 |
| GUIDE-07 | 멱등 noop | AC-CIF-008 | 그대로 채택 |
| GUIDE-08 | 관찰자 합치(init/update/doctor/hooks/curator/budget) | AC-CIF-003/012/016 | 채택 — v0.2.0부터 update 관찰자 포함 완전 채택(AC-CIF-016 update leg, 감사 D4) |
| GUIDE-09 | MCP 감사 CLAUDE guidance 경로 | — (Out of Scope) | **미채택** — GUIDE-09는 t1984 단일 담당, 구체 AC는 t1984 plan에서 확정; 본 카드 run은 감사 launcher/RPC 3파일(codex_audit_launch.go·codex_audit_mcp.go·mcp_codex_audit.go)을 편집하지 않는다 |
| GUIDE-10 | 문서/소스/템플릿 패리티·역사 불변 | AC-CIF-004/013/014 | 그대로 채택 |
| GUIDE-11 | 동시 변경 중단 | AC-CIF-010 | 그대로 채택(쓰기→remove 창 포함, D3) |
| GUIDE-12 | 모호 콘텐츠 미해결 보존 | AC-CIF-011 | 그대로 채택 |

> GUIDE fixture는 미실행 제안이었다 — 이 표의 채택/미채택 판정이 조정이며, 선행 85-fixture audit-only 집계에는 개수를 더하지 않는다.

## §D.4 심각도·게이트

- 릴리스-블로킹: AC-CIF-001, 002, 004, 005, 006, 007, 009, 010, 011, 012 — 모두 2-셀 충족(RED-now §D.1 행 참조 + green path).
- regression-guard: AC-CIF-013, 014 (불변 유지). 그 외 실행 AC(003, 008, 015, 016)는 run 단계 PASS 관측 필수.
- Definition of Done: 위 AC 전항 관측 기록(progress.md §E.2), `moai spec lint` 무결, 패키지 스코프 테스트 커버리지 85%+, golangci-lint 통과, 자체 저장소 전환 완료, 역사 증거 불변 확인.
- 미관측·거부된 검증은 Gaps로 명시 — 공백 Gaps는 "관측 누락 없음" 주장일 때만 유효.
