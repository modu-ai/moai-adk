# acceptance.md — SPEC-CLAUDE-INSTRUCTION-FILES-001

> stateless 산출물 (status 축 무상태). 검증 계층 — 모든 항목은 `AC-CIF-XXX` `Given/When/Then` 형식이며 이진 판정 가능해야 한다. Tier M 예산 준수 — AC 16건. **정적 확인(grep/ls)과 실행 검증(go test)을 모든 AC에서 구분하고, 미실행은 결코 PASS가 아니다.**

## §D.1 채택 규칙 (2-셀)

- 릴리스-블로킹 AC는 **RED-now 셀**(단일 호출 읽기 전용 명령 + 그 verbatim stdout + exit code + 트리 SHA)과 **green path 셀**(어느 마일스톤이 뒤집는지 + 통과 시 출력)을 쌍으로 갖는다(verification-completeness §2).
- 문서 수준 고정: 아래 표의 측정 트리는 **`cb30e05b3e`** (2026-10-27, 본 워크트리)이며, 이 고정은 자체 핀을 가지지 않는 모든 RED-now 셀을 묶는다. t1984 흡수 후에는 재측정·재고정한다(plan.md §C).
- RED는 "올바른 이유로 red"여야 한다 — 각 측정은 현행 트리가 요구사항의 반대 방향(AGENTS-정규)임을 보이는 positive 관측이다.

### RED-now 기준선 (측정 트리 `cb30e05b3e`, 2026-10-27)

| 명령 (단일 호출) | verbatim stdout | exit |
|---|---|---|
| `grep -n 'AgentsMD = "AGENTS.md"' internal/defs/files.go` | `21:	AgentsMD = "AGENTS.md"` | 0 |
| `grep -n 'ReplaceAll(text, "CLAUDE.md", "AGENTS.md")' internal/template/harness_fs.go` | `197:	text = strings.ReplaceAll(text, "CLAUDE.md", "AGENTS.md")` | 0 |
| `grep -n 'Move CLAUDE.local.md to AGENTS.local.md' internal/cli/migrate_local_instructions.go` | `27:		Short: "Move CLAUDE.local.md to AGENTS.local.md, keeping a backup",` | 0 |
| `grep -n 'Path: "AGENTS.local.md", BlockType: BlockTypeLearnedLocal' internal/harness/curator/dispatch.go` | `45:	3: {Path: "AGENTS.local.md", BlockType: BlockTypeLearnedLocal, Tier: 3},` | 0 |
| `ls internal/template/templates/CLAUDE.md.tmpl` | `ls: internal/template/templates/CLAUDE.md.tmpl: No such file or directory` | 1 |
| `grep -c 'AGENTS.md' internal/cli/doctor_harness.go internal/hook/instructions_loaded.go internal/cli/init_resume.go` | `internal/hook/instructions_loaded.go:9` `internal/cli/init_resume.go:4` `internal/cli/doctor_harness.go:3` | 0 |

읽는 법: exit 0인 행은 현행 트리가 해당 요구사항과 **반대 방향**임(요구 만족 시 그 grep은 비일치·exit 1이 되어야 하므로 오늘의 0이 RED). `ls` 행은 exit 1 자체가 RED(정규 템플릿 부재). grep -c 행은 참조 카운트 > 0이 RED.

## §D.2 AC 목록

### AC-CIF-001 — 정규 상수 방향 (maps REQ-CIF-001, GUIDE-01) [실행+정적, 릴리스-블로킹]
- **Given** 제품 소스 트리 **When** `grep -n 'AgentsMD = "AGENTS.md"' internal/defs/files.go` 실행 **Then** 매치 0건(exit 1)이고 `ClaudeMD` 주석이 정규(canonical)를 가리키며 CLAUDE-정규를 반영한 단위 테스트가 `--- PASS:`로 관측된다.
- RED-now: §D.1 표 1행. Green path: M1 — grep exit 1 + 전환 테스트 PASS 라인.

### AC-CIF-002 — init 생성 방향 (maps REQ-CIF-002, GUIDE-01) [실행, 릴리스-블로킹]
- **Given** 임시 프로젝트 디렉터리(t.TempDir) **When** `moai init --llm claude`(및 codex/both) 실행 **Then** CLAUDE.md가 생성되고 AGENTS.md·AGENTS.local.md가 존재하지 않는다(테스트가 os.Stat으로 부재 관측).
- RED-now: §D.1 표 1행(생성 방향의 원인 상수). Green path: M1 — init 통합 테스트 PASS.

### AC-CIF-003 — 런타임 리더 폴백 방향 (maps REQ-CIF-003, GUIDE-08) [실행]
- **Given** (a) CLAUDE.md와 AGENTS.md가 모두 있는 fixture (b) CLAUDE.md만 있는 fixture (c) AGENTS.md만 있는 fixture **When** 리더(`resolveInstructionFile` 등)가 명령 파일을 해석 **Then** (a) CLAUDE.md 우선, (c) 레거시 라벨 폴백, 모든 경우 어느 파일도 수정하지 않는다(사전/사후 해시 동일).
- RED-now: §D.1 grep -c 행(현재 AGENTS-우선). Green path: M3.

### AC-CIF-004 — 참조 정규화·정규 템플릿 (maps REQ-CIF-004, GUIDE-01/10) [실행+정적, 릴리스-블로킹]
- **Given** 정규화 대상 템플릿 바이트와 template 트리 **When** `normalizeHarnessReferences` 적용 및 `ls internal/template/templates/CLAUDE.md.tmpl` 실행 **Then** 결과에 `"AGENTS.md"`/`"AGENTS.local.md"` 활성 참조가 0건이고 CLAUDE.md.tmpl이 존재한다(exit 0), 제품 생성 표면에 활성 AGENTS 참조가 없다.
- RED-now: §D.1 표 2행 + 표 5행. Green path: M1 — 정규화 단위 테스트 PASS + ls exit 0.

### AC-CIF-005 — 마이그레이션 방향 반전·프리미티브 재사용 (maps REQ-CIF-005, GUIDE-02/03) [실행, 릴리스-블로킹]
- **Given** AGENTS.local.md(또는 AGENTS.md)만 있는 임시 저장소 **When** 마이그레이션 verb 실행 **Then** CLAUDE.local.md(또는 CLAUDE.md)로 이동하고 기존 프리미티브(공재 거부·Lstat 검사·타임스탬프 백업·O_EXCL·rollback)를 재사용한다(신규 메커니즘 부재 — 코드 검증 정적 확인).
- RED-now: §D.1 표 3행(현재 반대 방향). Green path: M2.

### AC-CIF-006 — 무손실 바이트 정확 백업 (maps REQ-CIF-006, GUIDE-02) [실행, 릴리스-블로킹]
- **Given** 알려진 바이트 내용과 권한(0600 등)의 AGENTS.local.md **When** 마이그레이션 실행 **Then** 백업 복사본이 원본과 바이트 동일(`cmp` exit 0)하고 권한 비트가 보존되며 최종 remove 성공 전까지 원본이 그 자리에 남는다.
- Green path: M2 (RED: AC-CIF-005와 동일 원인 — 방향이 반대라 이 동작 자체가 존재하지 않음).

### AC-CIF-007 — 공재·상이 거부 (maps REQ-CIF-007, GUIDE-04) [실행, 릴리스-블로킹]
- **Given** CLAUDE.local.md와 AGENTS.local.md가 모두 존재하고 내용이 다른 fixture **When** 마이그레이션 실행 **Then** exit non-zero, 양쪽 파일 모두 변경 없음(사전/사후 해시 동일), 오류 메시지가 두 경로와 수동 병합 지시를 보고.
- Green path: M2 — 거부 테이블 테스트.

### AC-CIF-008 — 멱등 재실행 noop (maps REQ-CIF-008, GUIDE-07) [실행]
- **Given** 이미 완료된 이동(정규 있음·레거시 부재) **When** verb 재실행 **Then** 파일·백업 디렉터리에 아무 쓰기 없음(백업 디렉터리 수 불변 관측), noop 메시지 출력.
- Green path: M2.

### AC-CIF-009 — 비정상 입력 거부 (maps REQ-CIF-009, GUIDE-05) [실행, 릴리스-블로킹]
- **Given** symlink·FIFO·읽기불가(0000)·경로이탈 입력 fixture **When** 마이그레이션 실행 **Then** 각각 exit non-zero로 거부하고 링크 대상·원본을 따라가거나 덮어쓰지 않는다(Lstat 결과 관측).
- Green path: M2 — 거부 테이블.

### AC-CIF-010 — 부분 실패 원자성·동시 변경 중단 (maps REQ-CIF-010, REQ-CIF-011; GUIDE-06/11) [실행, 릴리스-블로킹]
- **Given** (a) 최종 remove 실패, 백업 실패, 쓰기 실패 주입 fixture (b) 읽기와 쓰기 사이 원본 mtime/내용 변경 fixture **When** 마이그레이션 실행 **Then** (a) 부분 목적지 쓰기가 rollback되고 정확한 부분 결과가 보고되며 "Migrated" 성공이 없음 (b) 목적지 쓰기 전 중단, 양쪽 내용 보존.
- Green path: M2 — 실패 주입 테스트.

### AC-CIF-011 — 프라이버시·모호 분류 보존 (maps REQ-CIF-012, GUIDE-03/12) [실행+정적, 릴리스-블로킹]
- **Given** 비공개 문구를 포함한 AGENTS.local.md와 모호 공유/비공개 혼합 내용 fixture **When** 마이그레이션 및 자체 저장소 전환 실행 **Then** 비공개 문구가 CLAUDE.md에 0건(grep), 모든 로그·보고 출력에서 본문 미포함(출력 스캔), 모호 내용은 미해결로 보고되고 자동 게시되지 않는다.
- Green path: M2/M4.

### AC-CIF-012 — curator·예산 표면 재타겟 (maps REQ-CIF-013, GUIDE-08) [실행, 릴리스-블로킹]
- **Given** curator TierSurfaceMap과 token budget guard **When** 해당 단위 테스트 실행 **Then** Tier-3가 `CLAUDE.local.md`, Tier-4가 `CLAUDE.md` digest를 대상으로 하고 예산 가드 측정 표면이 정규 이름을 따른다.
- RED-now: §D.1 표 4행. Green path: M3.

### AC-CIF-013 — frozen 4-이름 계약 유지 (maps REQ-CIF-014, GUIDE-10) [실행, regression-guard]
- **Given** `internal/contract/frozen.go`와 `internal/hook/pre_tool.go` **When** `TestFrozenInstructionFiles` 계열 실행 **Then** 4-이름 목록이 그대로이고 두 목록이 동일(pin)하다 — 이관 전에도 green여야 하며 이관 후에도 green이어야 한다(불변 유지 criterion; vacuous 아님 — 목록을 줄이는 mutant에 red).
- Green path: M3에서 동일 PASS 관측(현재도 green).

### AC-CIF-014 — 역사 보존 (maps REQ-CIF-015, GUIDE-10) [정적+실행]
- **Given** `.moai/backups/local-instructions/**` 기존 항목과 `.moai/reports/**` **When** 마이그레이션 및 자체 전환 완료 **Then** 기존 백업·보고 파일의 내용 불변(해시 비교, 사전/사후).
- Green path: M4.

### AC-CIF-015 — 자체 저장소 전환 (maps REQ-CIF-016, GUIDE-02/03) [실행]
- **Given** 이 저장소의 AGENTS.md/AGENTS.local.md **When** 마이그레이션 verb 실행(커밋 경계) **Then** CLAUDE.md/CLAUDE.local.md가 존재하고 백업이 확보되었으며 init/doctor가 정규 파일에서 동작함을 실행으로 관측, `.claude/rules/**`·template mirror에 활성 AGENTS 지시가 남지 않았다(grep, historical/data-fixture 인용 제외).
- Green path: M4.

### AC-CIF-016 — 관찰자 합치 (maps REQ-CIF-001, REQ-CIF-003; GUIDE-08) [실행]
- **Given** CLAUDE-정규 프로젝트 fixture **When** `moai doctor`·init_resume 프로브·검증기 실행 **Then** L3 instruction-file 검사가 CLAUDE.md를 1차 표적으로 보고, `deployedWiringIsGPT` 계열 프로브가 CLAUDE.md를 사용하며 validator의 존재 검사가 CLAUDE.md를 가리킨다.
- RED-now: §D.1 grep -c 행(doctor 3건·init_resume 4건 포함). Green path: M1/M3.

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
| GUIDE-08 | 관찰자 합치 | AC-CIF-003/012/016 | 그대로 채택 |
| GUIDE-09 | MCP 감사 CLAUDE guidance 경로 | — (Out of Scope) | **미채택 — t1984 소유**. 본 카드 run은 감사 launcher/RPC 3파일을 편집하지 않는다; 필요 시 후속 카드 |
| GUIDE-10 | 문서/소스/템플릿 패리티·역사 불변 | AC-CIF-004/013/014 | 그대로 채택 |
| GUIDE-11 | 동시 변경 중단 | AC-CIF-010 | 그대로 채택 |
| GUIDE-12 | 모호 콘텐츠 미해결 보존 | AC-CIF-011 | 그대로 채택 |

> GUIDE fixture는 미실행 제안이었다 — 이 표의 채택/미채택 판정이 조정이며, 선행 85-fixture audit-only 집계에는 개수를 더하지 않는다.

## §D.4 심각도·게이트

- 릴리스-블로킹: AC-CIF-001, 002, 004, 005, 006, 007, 009, 010, 011, 012 — 모두 2-셀 충족 요구.
- regression-guard: AC-CIF-013 (불변 유지). 그 외 실행 AC(003, 008, 014, 015, 016)는 run 단계 PASS 관측 필수.
- Definition of Done: 위 AC 전항 관측 기록(progress.md §E.2), `moai spec lint` 무결, 패키지 스코프 테스트 커버리지 85%+, golangci-lint 통과, 자체 저장소 전환 완료, 역사 증거 불변 확인.
- 미관측·거부된 검증은 Gaps로 명시 — 공백 Gaps는 "관측 누락 없음" 주장일 때만 유효.
