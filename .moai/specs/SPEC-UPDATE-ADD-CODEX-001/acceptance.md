# Acceptance — SPEC-UPDATE-ADD-CODEX-001

판정 기준선은 측정 트리 `.claude/worktrees/t589` @ `5caddeb2d` (branch `WT-add-codex-verb`)다. 모든 GREEN 판정 명령은 구현이 착지한 트리에서 실행하며, 각 실행의 HEAD SHA 를 함께 기록한다(verification-claim-integrity §2 귀속).

## §D AC 매트릭스

| AC | 요구 (REQ) | 판정 명령 (요지) | 분류 |
|---|---|---|---|
| AC-UAC-001 | REQ-UAC-001/008 | `./bin/moai update --help` 에 `--add-codex` 존재 | release-blocking (RED §D.5 EV-1/2) |
| AC-UAC-002 | REQ-UAC-001 | claude-only 스크랫치 + `update --add-codex` → 3 산출물 생성 | release-blocking (RED EV-1) |
| AC-UAC-003 | REQ-UAC-002 | 플래그 없는 `update` → `.codex/` 무생성 | regression-guard |
| AC-UAC-004 | REQ-UAC-003 | `--add-codex` 전후 `.mcp.json` sha256 동일 | release-blocking (RED EV-1) |
| AC-UAC-005 | REQ-UAC-004 | 재실행 멱등 — sidecar sha 불변 + 무안내 | release-blocking (RED EV-1) |
| AC-UAC-006 | REQ-UAC-001 | 사용자 config.toml 항목 보존 | release-blocking (RED EV-1) |
| AC-UAC-007 | REQ-UAC-006 | `--dry-run` 조합 — 프리뷰만, 무기록 | release-blocking (RED EV-1) |
| AC-UAC-008 | REQ-UAC-007 | `--check --add-codex` → fail-loud 거절 | release-blocking (RED EV-1) |
| AC-UAC-009 | REQ-UAC-009 | 5 섹션 제목 + 본문 앵커 존재 | release-blocking (RED EV-6) |
| AC-UAC-010 | REQ-UAC-010 | `wc -c` ≤ 24,576 + 가드 테스트 green | regression-guard (기준 EV-5) |
| AC-UAC-011 | REQ-UAC-011 | CLAUDE.md 18 제목 + `@AGENTS.md` import 행 불변 + AGENTS.md 품질 게이트 포인터 | release-blocking (RED EV-8) |
| AC-UAC-012 | REQ-UAC-012 | AGENTS.md 마커 0 유지 + curator 불변 | regression-guard (기준 EV-3) |
| AC-UAC-013 | REQ-UAC-013 | init --force --agent both 안내 출력 | release-blocking (RED EV-4) |
| AC-UAC-014 | REQ-UAC-014 | 영향 패키지 테스트 green + 고정 테스트 삭제 금지 | regression-guard |

## §D.1 AC 상세

### AC-UAC-001 — update 도움말이 --add-codex 를 안다

- **Given** 구현이 착지된 트리에서 빌드한 `bin/moai`
- **When** `./bin/moai update --help` 실행
- **Then** 출력에 `--add-codex` 와 그 용도(기존 프로젝트에 codex 덧붙임)가 존재한다
- 판정: `./bin/moai update --help | grep -c -- --add-codex` ≥ 1

### AC-UAC-002 — claude-only 프로젝트에 3 산출물을 만든다

- **Given** `moai init` (claude, wiring 부재 — `.codex/` 없음)으로 만든 스크래치 프로젝트
- **When** `moai update --add-codex` 실행
- **Then** (1) `.codex/hooks.json` 존재(레지스트리 계약 훅 포함), (2) `.codex/config.toml`에 `[mcp_servers.moai]`와 `[tui].status_line` 존재, (3) `.moai/state/codex-wiring.json` 사이드카 존재, (4) 최초 생성 안내(FirstTrustGuidance) 출력
- 판정: 스크랫치에서 3 파일 존재 + `grep -c "mcp_servers.moai" .codex/config.toml` ≥ 1

### AC-UAC-003 — 플래그가 없으면 아무것도 만들지 않는다 (회귀 방어)

- **Given** 같은 claude-only 스크랫치 프로젝트
- **When** 플래그 없는 `moai update` 실행
- **Then** `.codex/` 디렉터리가 새로 생성되지 않는다 (SPEC-CODEX-WIRING-001 AC-CW-008 계승)
- 판정: 실행 전후 `test -e .codex && echo exists || echo absent` — absent 불변

### AC-UAC-004 — .mcp.json 을 건드리지 않는다

- **Given** 사용자가 편집한 `.mcp.json` 이 있는 프로젝트 (sha256 사전 기록)
- **When** `moai update --add-codex` 실행
- **Then** `.mcp.json` 바이트 동일 (결정 D2)
- 판정: `shasum -a 256 .mcp.json` 전후 일치

### AC-UAC-005 — 재실행은 멱등이다

- **Given** AC-UAC-002 통과 프로젝트 (sidecar sha256 기록)
- **When** `moai update --add-codex` 재실행
- **Then** sidecar sha 불변, hooks.json 재기록 없음, 재신뢰 안내 출력 없음 (`wireProject` 불변-재생성 무음 계약)
- 판정: `shasum -a 256 .moai/state/codex-wiring.json` 전후 일치 + 안내 문구 부재

### AC-UAC-006 — 사용자가 추가한 config.toml 항목을 보존한다

- **Given** `.codex/config.toml`에 사용자 자신의 테이블/키를 추가한 프로젝트
- **When** `moai update --add-codex` 실행
- **Then** 사용자 항목 존재 유지 + `[mcp_servers.moai]` 존재 (create-if-absent 병합)
- 판정: 사용자 항목 grep ≥ 1 전후 유지

### AC-UAC-007 — dry-run 은 쓰지 않는다

- **Given** claude-only 스크랫치 프로젝트
- **When** `moai update --add-codex --dry-run` 실행
- **Then** 예정 wiring 동작이 출력되고 `.codex/`·sidecar 미생성
- 판정: `test -e .codex` → absent

### AC-UAC-008 — --check 조합은 거절된다

- **Given** 구현이 착지된 바이너리
- **When** `moai update --check --add-codex` 실행
- **Then** 상호배타 검증기가 fail-loud 거절 (exit ≠ 0)
- 판정: exit 코드 ≠ 0 + 오류 메시지에 두 플래그 언급

### AC-UAC-009 — 5 섹션이 존재하고 내용이 있다

- **Given** M2 착지 후 템플릿
- **When** `internal/template/templates/AGENTS.md` 를 grep
- **Then** plan.md §D6 의 5 제목(`## 8. Codex Web Console` ~ `## 12. Status Line Tokens`) 각각 ≥ 1회 + 각 섹션별 본문 앵커 2차 grep ≥ 1회 (제목 껍데기 배제 — 앵커 목록: (a) `console` 계열, (b) `hook` + 이벤트명 ≥ 2, (c) `.moai/config/sections`, (d) 표 헤더 또는 동사명 ≥ 3, (e) `statusline`/`MOAI_STATUSLINE` 계열)
- 판정: `rg -c '^## (8|9|10|11|12)\. '` = 5

### AC-UAC-010 — 바이트 상한이 지켜진다

- **Given** M2 착지 후 템플릿
- **When** `wc -c internal/template/templates/AGENTS.md`
- **Then** ≤ 24,576 그리고 `go test ./internal/config -run TestCodexContractByteCeiling` ok (루트 AGENTS.md + 템플릿 미러 쌍방)
- 판정: 두 명령 모두 green — 상한 위반은 가드 테스트가 fail-closed 로 잡는다

### AC-UAC-011 — universal 은 이동하고 껍데기는 남는다

- **Given** M2 착지 후 템플릿 쌍
- **When** 템플릿 CLAUDE.md 의 `^## [0-9]` 헤딩을 센다
- **Then** 개수 18 불변(§0–§17, pure wrapper 봉쇄) AND `@AGENTS.md` import 행 유지(헤딩 카운트만으로는 잡지 못하는 상실 — plan-audit F2) AND 템플릿 AGENTS.md 에 품질 게이트 포인터 존재(`rg -c "Quality Gates|TRUST"` ≥ 1 — 착지 전 0, §D.5 EV-8)
- 판정: `rg -c '^## [0-9]' internal/template/templates/CLAUDE.md` = 18 AND `grep -c '@AGENTS.md' internal/template/templates/CLAUDE.md` ≥ 1

### AC-UAC-012 — 학습 마커는 CLAUDE.md 에만 산다

- **Given** M2 착지 후 템플릿
- **When** `rg -n "MOAI:LEARNED-WORKFLOW" internal/template/templates/AGENTS.md`
- **Then** 0 matches 유지(착지 전부터 0 — 회귀 방어) AND `internal/harness/curator` 의 `TierSurfaceMap` 이 여전히 Tier 4 → CLAUDE.md 앵커(간접: `rg -n 'Path: "CLAUDE.md"' internal/harness/curator/` ≥ 1)
- 판정: 첫 grep 0 + 둘째 grep ≥ 1

### AC-UAC-013 — codex-add 목적의 --force 재초기화는 안내를 출력한다

- **Given** 초기화된 프로젝트
- **When** `moai init --force --agent both --non-interactive` 실행 (스크래치 복제 위 SAFE 위치에서)
- **Then** stdout/stderr 어딘가에 `moai update --add-codex` 안내 문구 존재, 재초기화 자체는 요청대로 진행 (redirect-not-block)
- 판정: 출력에 `update --add-codex` 문자열 ≥ 1

### AC-UAC-014 — 영향 패키지가 green 이고 고정 테스트는 살아 있다

- **Given** M3 착지 후 트리
- **When** `go test ./internal/cli/... ./internal/codexwiring/... ./internal/config/...`
- **Then** ok + `rg -c '^func Test' internal/cli/init_agent_flag_test.go` 가 M3 직전 값 이상 (삭제로 "갱신" 금지 — §G mutant)
- 판정: 테스트 ok + 함수 수 비감소

## §D.2 심각도

- **release-blocking** (RED-now 가 §D.5 장부로 관측됨): AC-UAC-001, 002, 004, 005, 006, 007, 008, 009, 011, 013 — 이 중 004/005/006/007/008 의 RED 는 "동사 자체가 없음"(EV-1)으로 관측된다. 새 능력의 AC 는 그 주체가 없으면 관측 불가하며, 주체 부재가 곧 RED 다.
- **regression-guard** (오늘 이미 green — 유지 계약): AC-UAC-003, 010, 012, 014. 이들은 RED-now 를 요구하지 않는 분류다 — 존재하는 행동의 보존이 계약이며, 깨졌을 때 잡는 것이 역할이다.

## §D.3 추적성

| REQ | AC |
|---|---|
| REQ-UAC-001 | AC-UAC-002, AC-UAC-006 |
| REQ-UAC-002 | AC-UAC-003 |
| REQ-UAC-003 | AC-UAC-004 |
| REQ-UAC-004 | AC-UAC-005 |
| REQ-UAC-005 | AC-UAC-002 (거부 경로는 codexwiring 패키지 테스트가 1차 — 간접, §D.4) |
| REQ-UAC-006 | AC-UAC-007 |
| REQ-UAC-007 | AC-UAC-008 |
| REQ-UAC-008 | AC-UAC-001 |
| REQ-UAC-009 | AC-UAC-009 |
| REQ-UAC-010 | AC-UAC-010 |
| REQ-UAC-011 | AC-UAC-011 |
| REQ-UAC-012 | AC-UAC-012 |
| REQ-UAC-013 | AC-UAC-013 |
| REQ-UAC-014 | AC-UAC-014 |

## §D.4 간접 검증 항목

1. **REQ-UAC-005 (fail-loud 거부) — 기계 판정 2층** (plan-audit F1 승격):
   - **기존 판정자 green 유지**: `TestWireValidationRefusalWritesNothing`(`internal/codexwiring/wire_test.go:164`) — 거부 시 무기록을 `wireProject` 본체 레벨에서 검증한다. `--add-codex` 가 이 본체를 호출하는 한 거부 무기록은 공유된다.
   - **신규 판정자 존재 필수**: `TestUpdateAddCodex_ValidationRefusalFailsLoud`(`internal/cli`) — `--add-codex` 경로가 `ErrValidationRefused` 를 삼키지 않고 exit ≠ 0 으로 전파함을 CLI 래퍼 레벨에서 핀다. 이 테스트가 없으면 "경고로 삼키는 래퍼 mutant"(감사 F1)가 전 AC 를 통과한다 — 형제 래퍼 `refreshCodexWiringBestEffortAt`(`update_codex_wiring.go:16-20`)이 그 모델이다. 테스트 부재는 AC-UAC-002/REQ-UAC-005 의 FAIL 이다.
2. **Template-First 규율** — `make build` 후 `make embed-check` (수동): 바이너리 임베드가 커밋본과 일치.
3. **템플릿 중립성** — `template-neutrality-check.yaml` CI 가드 green: 5 섹션 본문에 SPEC ID·내부 날짜·커밋 SHA 없음.
4. **agents-emit 무영향** — 에이전트 정의 미변경이므로 `internal/template/agentemit` 드리프트 0 (기존 검사 green 유지).

## §D.5 RED-now 증거 장부

측정: 2026-09-09, 트리 `5caddeb2d` (branch `WT-add-codex-verb`). `bin/moai` 는 이 트리에서 `go build -o bin/moai ./cmd/moai` 로 빌드.

```
EV-1
  command: ./bin/moai update --add-codex
  stdout:
     ERROR
  Unknown flag: --add-codex.
  Try --help for usage.
  exit code: 1
  tree: 5caddeb2d
  cited by: AC-UAC-001, 002, 004, 005, 006, 007, 008 (동사 부재 = 가족 전체의 RED)

EV-2
  command: rg -n "add-codex" internal/cli/update.go
  stdout: (empty — 0 matches)
  exit code: 1 (rg no-match)
  tree: 5caddeb2d
  cited by: AC-UAC-001

EV-3
  command: rg -n "MOAI:LEARNED-WORKFLOW" internal/template/templates/AGENTS.md
  stdout: (empty — 0 matches)
  exit code: 1 (rg no-match)
  tree: 5caddeb2d
  cited by: AC-UAC-012 (green-now 기준선 — 유지 계약)

EV-4
  command: rg -n "add-codex" internal/cli/init.go
  stdout: (empty — 0 matches)
  exit code: 1 (rg no-match)
  tree: 5caddeb2d
  cited by: AC-UAC-013

EV-5
  command: wc -c internal/template/templates/AGENTS.md
  stdout: 15415 internal/template/templates/AGENTS.md
  exit code: 0
  tree: 5caddeb2d
  cited by: AC-UAC-010 (green-now 기준선; 상한 24576 대비 여유 9161 B)

EV-6
  command: rg -n '^## ' internal/template/templates/AGENTS.md
  stdout:
    35:## 1. Evidence and verification claims
    59:## 2. Git, branches, and the shared checkout
    103:## 3. Worktrees
    132:## 4. How verification is run
    158:## 5. Core behaviors
    204:## 6. Output, language, and format
    233:## 7. Tools and command output
  exit code: 0
  tree: 5caddeb2d
  cited by: AC-UAC-009 (## 8~12 부재 = RED) 및 §D6 배치 근거

EV-7
  command: rg -c '^## [0-9]' internal/template/templates/CLAUDE.md
  stdout: 18
  exit code: 0
  tree: 5caddeb2d
  cited by: AC-UAC-011 (보존 기준선)

EV-8
  command: rg -n "Quality Gates|TRUST" internal/template/templates/AGENTS.md
  stdout: (empty — 0 matches)
  exit code: 1 (rg no-match)
  tree: 5caddeb2d
  cited by: AC-UAC-011 (포인터 부재 = RED)

EV-9
  command: ls internal/template/templates/.agents/skills/
  stdout: moai-clean … moai-todo — 16개 moai-* 디렉터리
  exit code: 0
  tree: 5caddeb2d
  cited by: spec.md §1.1 M6 (--add-codex 실현 가능성 근거; AC 아님)

EV-10
  command: echo "SPEC-UPDATE-ADD-CODEX-001" | grep -E '^(SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3})$'
  stdout: SPEC-UPDATE-ADD-CODEX-001
  exit code: 0
  tree: 5caddeb2d
  cited by: SPEC ID 사전 검증 (spec-workflow 매니저 프로토콜; 중복 검사 rg 0 matches 병행)
```

비고: EV-2/3/4/8 은 rg no-match(empty stdout) 판독이다. 파일 존재는 각각 선행 측정(update.go/init.go/AGENTS.md 판독)으로 확인돼 있어 "빈 출력 = 0 matches" 판독이 성립한다 — 대상 파일 부재에 의한 빈 출력이 아니다.
