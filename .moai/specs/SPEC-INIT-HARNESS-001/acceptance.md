# acceptance.md — SPEC-INIT-HARNESS-001

## §A 판정 규약

- 모든 AC는 **2셀 규율**(verification-completeness.md §2)을 따른다: RED-now 셀(구현 전 나무에서 빨간 이유가 명시된 관측) + green-path 셀(어느 마일스퀀스가 뒤집는가).
- **트리 고정**: 본 문서의 RED-now 관측 기준 나무는 워크트리 HEAD **`a404132e7`**(트리 `aebfa5bea`)다. 개별 AC에 핀이 없으면 이 핀이 적용된다.
- **verbatim 출력 슬롯**: RED 관측 중 실제 init 실행을 수반하는 AC(AC-IH-002/003/004/009/011)는 t583의 홈 지문 슬롯 규약(REQ-IQW-014/016 — 선언 후 실행, 전후 실제 홈 8항목 지문, stderr 분리, 불일치 시 작업 멈춤·리드 보고)을 준용해 **run 단계 해당 마일스퀀스의 첫 슬롯에서 verbatim 출력을 채운다**. 이 셀들의 verbatim 칸이 비어 있는 것은 "측정 예정"의 표시지 않고 "plan 단계에서는 실행하지 않기로한 홈 안전 규약의 결과"다 — 첫 슬롯 기록이 채우며, 빈 채로 sync에 도달하면 그 시점에 FAIL로 판정한다.
- 파일집 단정 명령은 공통 형식을 쓴다: `$MOAI_INIT_ROOT` = init이 실행된 임시 프로젝트 루트(t.TempDir() 아래), 부정 단정은 `test -e` 실패 = 통과.
- 상태 토큰: `PASS` / `FAIL` / `PASS-WITH-DEBT`(사유 명시 필수).

## §B AC Matrix (AC-IH-001..016)

### AC-IH-001 — `llm.harness` 영속화 (REQ-IH-001, REQ-IH-002)

- **RED-now**: 신규 테스트 `TestInitPersistsHarnessKey`(제안) — init 실행 후 `llm.yaml`에 `llm.harness` 단정. 오늘 나무에서 키가 쓰이지 않으므로 RED.
  - command: `go test ./internal/cli/ -run TestInitPersistsHarnessKey -count=1`
  - verbatim: _(run 단계 M1 첫 슬롯에서 채움 — §A 슬롯 규약)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
  - 빨간 이유: 오늘 init은 `llm.harness` 키를 쓰는 코드가 없다(존재하지 않는 키) — wrong-reason red 아님, 이 카드의 첫 변경(M1)이 정확히 이 키를 쓴다.
- **green-path**: M1 — 3값(claude/codex/both) 각각 기록 + 미설정 claude 기록까지 포함해 PASS.

### AC-IH-002 — codex 단독 부정 단정 (REQ-IH-005)

- **RED-now**: `--llm codex` init 후 `.claude/` 부재 단정. 오늘 나무는 `.claude/`를 심는다 — RED.
  - command: `go test ./internal/cli/ -run TestInitCodexOnlyDeploysNoClaudeSurface -count=1` (제안)
  - verbatim: _(run 단계 M2 첫 슬롯 — 홈 지문 규약)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
  - 빨간 이유: `deployer.go:160-284` walk에 필터가 없어 `--llm codex`에도 `.claude/**`·`CLAUDE.md`·`.mcp.json`이 기록된다(research.md §1 전제 1·6) — 이 카드 M2의 harnessFS가 제거하는 바로 그 동작.
  - 단정 목록: `$MOAI_INIT_ROOT/.claude` 부재, `CLAUDE.md` 부재, `.mcp.json` 부재, `.claudeignore` 부재, `.moai/status_line.sh` 부재.
- **green-path**: M2 — 5개 부정 단정 전부 PASS.

### AC-IH-003 — codex 단독 긍정 단정 (REQ-IH-005)

- **RED-now**: codex 단독 필수 표면 단정 — 오늘은 `.claude/`까지 심긴 하지만 이 AC가 요구하는 **정확한 파일집**(특히 `.moai/status_line.sh` 부재 조건 포함 집합)은 성립하지 않는다. RED의 형태: AC-IH-002가 먼저 실패하므로 본 AC는 독립 실행 시 status_line.sh 존재로 실패.
  - command: `go test ./internal/cli/ -run TestInitCodexOnlyRequiredSurfaces -count=1` (제안)
  - verbatim: _(run 단계 M2 슬롯 — 홈 지문 규약)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
- **green-path**: M2 — 단정: `AGENTS.md` 존재, `.codex/hooks.json`·`.codex/config.toml`·신뢰 사이드카 존재, `.codex/agents/moai/*.toml` 11개, `.agents/skills/moai-{plan,run,sync,fix,gate,goal,loop,mx,clean,codemaps,e2e,feedback,harness,project,review,todo}/SKILL.md` 16종, 카탈로그 재매핑 34종 존재, `.moai/config/sections/llm.yaml` 존재, `.gitignore` 존재.

### AC-IH-004 — 재매핑 무결성 (REQ-IH-006, REQ-IH-007)

- **RED-now**: 재매핑된 스킬이 실디렉터이고 링크가 0개라는 단정 — 오늘은 `.agents/skills`의 카탈로그 항목이 `../../.claude/skills/...` 심볼릭 링크다.
  - command: `go test ./internal/template/ -run TestCodexOnlySkillRemapIntegrity -count=1` (제안)
  - verbatim: _(run 단계 M2 슬롯)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
  - 빨간 이유: `skill_mirror.go:174-176` — 오늘 배포는 카탈로그를 링크로 만든다; codex 단독 재매핑 실디렉터는 M2 신규 동작.
- **green-path**: M2 — `$MOAI_INIT_ROOT/.agents/skills` 아래 symlink 0, `<name>/SKILL.md` 판독 가능, 충돌 skip 시나리오(사전 점유 더미 배치 → skip + 경고 기록) PASS.

### AC-IH-005 — claude 배포 보존 — 파일집·로직 (REQ-IH-003)

- **RED-now**: (보존형 — 신규 RED 없음) 기존 claude init 실행 테스트 전체가 `a404132e7`에서 green이라는 관측이 시작 관측이다.
  - command: `go test ./internal/cli/ -run 'TestRunInit' -count=1`
  - verbatim: `ok  	github.com/modu-ai/moai-adk/internal/cli` 형태의 기존 green — M2 착지 전후 동일 출력 유지가 본 AC의 판정이다.
  - exit code: 0 → 0 (불변 유지)
  - tree: `a404132e7`
- **green-path**: M2/M3 — claude 배포의 **파일집과 배포 로직**에 변경이 없음을 "기존 테스트 본문 무변경 + 전부 green"으로 증명. REQ-IH-002(`llm.harness` 키)와 REQ-IH-008(AGENTS.md 내용 보강)이 명시하는 내용 갱신은 이 단정의 범위 밖이며 각자의 AC(AC-IH-001, AC-IH-007)가 잰다. 테스트 본문 수정이 발생하면 blocker 보고.

### AC-IH-006 — both 보존 (REQ-IH-004)

- **RED-now**: (보존형) `internal/cli/init_agent_flag_test.go:83-91`(AC-CW-005 — both 배선+사이드카)과 MCP 강제 규칙 테스트가 `a404132e7`에서 green — 시작 관측.
  - command: `go test ./internal/cli/ -run 'TestRunInit_AgentBoth|TestRunInit_CallsCodexWiring' -count=1`
  - verbatim: 기존 green 출력 — 동일성 유지가 판정.
  - exit code: 0 → 0
  - tree: `a404132e7`
- **green-path**: M2 — both 배포에 새 필터 경로가 탑재되지 않음을 확인(파일집·로직 보존; `llm.harness` 기록은 AC-IH-001, AGENTS.md 내용은 AC-IH-007이 각자 판정).

### AC-IH-007 — AGENTS.md 공개 완전성 (REQ-IH-008)

- **RED-now**: 동결 목록 8개 각각이 배포본 AGENTS.md에서 매칭된다는 단정 — 오늘 템플릿 AGENTS.md는 Agent 소환·output style·슬래시 명령·Workflow 스크립트 4건을 전무하고, 스킬 로더는 주소·도달 축만 산문 공개(`:27-33`) 상태다(research.md §4 갭 분석).
  - command: `go test ./internal/template/ -run TestAgentsDisclosureCompleteness -count=1` (제안 — 동결 목록 상수 기반)
  - verbatim: `--- FAIL: ... missing disclosures: [agent-spawning output-style slash-commands workflow-scripts skill-loader:non-equivalence]` 형태
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
- **스킬 로더 판정 (D5)**: 템플릿 AGENTS.md :27-33은 `skill-loader`가 "모든 하네스가 갖는 능력"이며 미러 경로(`.agents/skills/<name>/SKILL.md`)에서 동등하게 도달한다고 이미 산문으로 공개한다 — 이 부분 공개는 유효한 포인터로 **인정**한다. REQ-IH-008이 스킬 로더에 요구하는 나머지 조각은 "codex 런타임 스킬 로더가 아직 미등가(deferred) 상태"임을 명시하는 절 1개다 — 표 행 전면 재저작이 아니라 기존 산문 뒤 포인터 보강으로 충족함을 명시적으로 판정한다.
- **green-path**: M3 — 템플릿 AGENTS.md 보강 후 PASS + `internal/config/token_budget_guard.go` `CodexContractByteCeiling`(24,576B) 이내 유지 + `make build` green.

### AC-IH-008 — 문서 (REQ-IH-009)

- **RED-now**: README init 절에 3-way 하니스 선택("Codex only") 서술 부재.
  - command: `grep -c "Codex only" README.md`
  - verbatim:
    ```
    0
    ```
  - exit code: 1 (grep -c는 0매치에서 exit 1) → 문서 추가 후 출력 1 이상·exit 0
  - tree: `a404132e7` — plan 단계 2026-09-14 이 트리 실행 결과가 위 verbatim이다(단일 호출형, D4 정정).
- **green-path**: M3 — README("Codex only" ≥1, 4-locale 동기) + docs-site init 가이드 절 착지(oss-docs 4-locale 동기 규칙 준용 — docs-site 표면의 판정식은 run 단계에서 같은 단일 호출 형태로 확정한다).

### AC-IH-009 — update 부활 방지 (REQ-IH-010)

- **RED-now**: codex 단독 고정 프로젝트(시뮬레이션: `.claude/` 부재 + `llm.harness: codex`)에 `moai update` 재배포 후 `.claude/` 부재 단정 — 오늘 update는 전면 재배포하므로 `.claude/`가 부활한다.
  - command: `go test ./internal/cli/ -run TestUpdateCodexOnlyNoClaudeResurrection -count=1` (제안)
  - verbatim: _(run 단계 M4 슬롯 — 홈 지문 규약)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
  - 빨간 이유: update 재배포 경로(`runCleanReinstall` → deployer.Deploy)에 하니스 인지가 없다(research.md §2.3).
- **green-path**: M4 — 재배포 후 `.claude/` 0 + codex 표면 갱신 + CleanMoaiManagedPaths 부재 경로 무해 확인.

### AC-IH-010 — doctor 하니스 조건부 (REQ-IH-011)

- **RED-now**: codex 단독 프로젝트 doctor 결과에 claude 표면 finding(settings.json·hooks·statusline 부재 경고)이 오늘은 발화 — 하니스 조건이 없다.
  - command: `go test ./internal/cli/ -run TestDoctorCodexOnly -count=1` (제안 — golden 시나리오 추가)
  - verbatim: _(run 단계 M4 — 기존 finding 목록을 golden 스냅샷으로 먼저 채집)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
- **green-path**: M4 — claude 표면 finding이 INFO 강등(`claude surface: not deployed (harness=codex)`)되고 codex 검사 finding은 유지. 기존 golden 시나리오 무변경 통과.

### AC-IH-011 — 비대화형 parity (REQ-IH-012)

- **RED-now**: `--non-interactive --llm codex` 파일집 == 대화형 codex 선택 파일집 단정 — M2 착지 전까지 대화형 codex 자체가 미구현이라 RED.
  - command: `go test ./internal/cli/ -run TestInitCodexNonInteractiveParity -count=1` (제안)
  - verbatim: _(run 단계 M5 슬롯 — 홈 지문 규약)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
- **green-path**: M5 — 두 경로 파일집 diff 0. 비대화형 MCP 비대칭(REQ-IQW-006)은 codex에서 `.mcp.json`을 어차피 안 쓰므로 관찰 불가 — parity 단정에서 자연 제외됨을 테스트 주석으로 명시.

### AC-IH-012 — 위저드 문항 갱신 (REQ-IH-013)

- **RED-now**: `agent_wiring` 옵션 설명이 배포 결과를 서술하지 않는다 — 오늘 서술은 "연결(wiring)"뿐(questions.go:380-385).
  - command: `go test ./internal/cli/wizard/ -run TestAgentWiringOptions -count=1` (제안 — 3옵션·값 동결·배포 서술 키워드 단정)
  - verbatim: `--- FAIL: ... option codex description does not describe deployment` 형태
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
- **green-path**: M3 — Label/Desc 갱신 + 번역 3블록(`translations.go:107/:193/:279`) 동기 + 값(claude/codex/both)·옵션 수(3)·질문 수(4) 동결 단정 + reconfigure 문항 무변경 기존 테스트 green 유지.

### AC-IH-013 — AC-CW-004 보존 (REQ-IH-014·§4 제약)

- **RED-now**: (보존형) `internal/cli/init_agent_flag_test.go` 전체가 `a404132e7`에서 green — 시작 관측.
  - command: `go test ./internal/cli/ -run TestInitAgentFlag -count=1 && go test ./internal/cli/ -run 'TestValidateInitFlags_AgentClosedSet|TestRunInit_AgentAbsentLeavesNoCodexFiles|TestRunInit_AgentClaudeLeavesNoCodexFiles' -count=1`
  - verbatim: 기존 green 출력 — 본문 무변경 + green 유지가 판정.
  - exit code: 0 → 0
  - tree: `a404132e7`
- **green-path**: M5 — 플래그 부재 ≡ claude ≡ `.codex/` 배선 0 불변 + 닫힌 집합 fail-loud 불변.

### AC-IH-014 — `tool enable codex` 무변경 (Out of Scope 경계)

- **RED-now**: (보존형) `go test ./internal/cli/ -run TestToolEnableCodex -count=1`이 `a404132e7`에서 green — 시작 관측.
  - command: 위 그대로
  - verbatim: 기존 green — 유지가 판정.
  - exit code: 0 → 0
  - tree: `a404132e7`
- **green-path**: M5 — 이 SPEC이 tool.go를 만지지 않음으로써 증명(diff에 tool.go 부재).

### AC-IH-015 — `llm.harness` 병합 생존 (REQ-IH-002·design D5)

- **RED-now**: init이 쓴 `llm.harness`가 update 3-way 병합(RestoreMoaiConfigRetained) 후 생존하는가 — M1 전까지는 키 자체가 없어 RED(AC-IH-001의 상위 결합).
  - command: `go test ./internal/cli/ -run TestUpdatePreservesHarnessKey -count=1` (제안)
  - verbatim: _(run 단계 M4 슬롯)_
  - exit code: 1 (RED) → 0 (GREEN)
  - tree: `a404132e7`
- **green-path**: M4 — init(하니스 codex) → update → `llm.harness: codex` 생존 단정 PASS.

### AC-IH-016 — 품질 게이트 (마감)

- **RED-now**: (보존형) baseline — §C pre-flight의 스위트·vet·lint·windows 크로스빌드 결과를 run 진입 시 기록.
- **green-path**: M5 — `go test ./internal/cli/... ./internal/template/... ./internal/core/project/...` + `go vet` + `golangci-lint run` + `GOOS=windows GOARCH=amd64 go build ./...` 전부 green, NEW 결함 0.

## §C 추적 표 (REQ ↔ AC ↔ Milestone)

| REQ | AC | Milestone |
|---|---|---|
| REQ-IH-001, REQ-IH-002 | AC-IH-001, AC-IH-015 | M1, M4 |
| REQ-IH-003 | AC-IH-005 | M2/M3 |
| REQ-IH-004 | AC-IH-006 | M2 |
| REQ-IH-005 | AC-IH-002, AC-IH-003 | M2 |
| REQ-IH-006, REQ-IH-007 | AC-IH-004 | M2 |
| REQ-IH-008 | AC-IH-007 | M3 |
| REQ-IH-009 | AC-IH-008 | M3 |
| REQ-IH-010 | AC-IH-009 | M4 |
| REQ-IH-011 | AC-IH-010 | M4 |
| REQ-IH-012 | AC-IH-011 | M5 |
| REQ-IH-013 | AC-IH-012 | M3 |
| REQ-IH-014 | AC-IH-013 | M5 |
| (경계) | AC-IH-014 | M5 |
| (마감) | AC-IH-016 | M5 |

## §D Definition of Done

1. AC-IH-001..016 전부 PASS(보존형 4개는 무변경 green 유지로 PASS).
2. CHANGELOG에 `--llm codex` 의미 변경 Breaking 라벨 기재(sync 단계).
3. `progress.md` §E.2/§E.3에 슬롯 지문 포함 run 증거 착지.
4. sync-auditor PASS.
