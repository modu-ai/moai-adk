# SPEC-FACTORY-BOOT-MINIMAL-001 — 수용 기준

> 문서 수준 트리 고정: 모든 RED-now 관측은 `59959629908cda723c37e26c7c15718194d6b38a`에서 측정했다(verification-completeness.md §2.1 — 자체 고정이 없는 기준은 이 SHA에 묶인다). 원문 출력 파일은 `probe/`에 있으며 이 문서의 증거 장부(§B)가 그 파일을 인용한다.

## §A 기준 목록

| AC | 분류 | 요구사항 | 한 줄 요약 | 증거 장부 |
|---|---|---|---|---|
| AC-FBM-001 | release-blocking | REQ-FBM-001, REQ-FBM-002, REQ-FBM-004, REQ-FBM-005 | 탐침: 72칸 전부 고정분 ≤ 2,847, 합본 ≤ 9,146, 최소 원소 존재, 자동 모드 설명 부재 | E1 |
| AC-FBM-002 | release-blocking | REQ-FBM-008, REQ-FBM-009 | 저장소 내 예산 테스트가 실제 렌더 고정분을 칸별로 측정·단언 | E2 |
| AC-FBM-003 | release-blocking | REQ-FBM-003 | 자동 모드 안내 필드가 메시지 테이블에서 제거 | E3 |
| AC-FBM-004 | release-blocking | REQ-FBM-010 | AGENTS.md ≤ 24,276바이트 | E4 |
| AC-FBM-005 | regression-guard | REQ-FBM-010 | AGENTS.md의 배타성 문구 「live only under」 보존 | E5 |
| AC-FBM-006 | regression-guard | REQ-FBM-009, REQ-FBM-010 | Codex 상한 테스트 2개 PASS, skip 없음 | E6 |
| AC-FBM-007 | regression-guard | REQ-FBM-007 | 역할 게이트 규칙 2개(배포·템플릿) 무변경 | E7 |
| AC-FBM-008 | regression-guard | REQ-FBM-011 | 착지 바이너리로 6레인+리더 라이브 합본 ≤ 10,000, 넘침 0 | E8 |
| AC-FBM-009 | regression-guard | REQ-FBM-006, REQ-FBM-002 | 레인 규칙 토큰·스폰 권한·core 의무 요소 테스트 PASS | E9 |

regression-guard 분류 사유: AC-005·007·009는 보존 단언이라 기준 트리에서 이미 녹색(RED-now 없음). AC-006은 카드가 인용한 CI RED(main@3f81377dc9)가 현재 트리에서 재현되지 않는다(§2.1 undecidable disposition — release 게이트 자격 상실, pass로 기록하지 않음). AC-008은 라이브 세션이라 plan 단계에서 재실행 가능한 RED가 아니다(역사적 관측만 있음).

## §B 시나리오 (Given-When-Then)

### AC-FBM-001
- **Given** 기준 트리 이후의 작업 트리와 `probe/overlay.json`
- **When** `go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run '^TestT1923BootBudgetProbe$' -v ./internal/hook/` 를 실행하면
- **Then** exit 0, 출력에 `PROBE swept=72`, `--- FAIL` 0건, `--- PASS: TestT1923BootBudgetProbe ` (이름 뒤 공백) 가 있다. (E1의 RED 관측은 앵커 없는 `-run 'TestT1923BootBudgetProbe$'` 로 실행했다 — 이 패키지에 같은 접두 이름의 테스트가 없어 선택 집합은 같다.)
- RED-now: E1 (exit 1, 실패 칸 24개(72칸 중) — 사유: 영어 lane 고정분 3,751 > 2,847·합본 10,006 > 9,146, 일본어 lane 2,852 > 2,847, 4개 언어 리더·레인 공지에 `--auto-leader`/`--auto-lane` 설명 잔존). 올바른 이유의 RED: 실패 사유가 모두 이 SPEC이 바꿀 렌더 텍스트다.
- Green 경로: M3(메시지 테이블 축소)가 뒤집는다. 통과 출력은 칸별 `PROBE ... fixed=<≤2847> composite=<≤9146> core=4048` 72줄 + swept=72.
- 변이 탐침: 공지를 비워 크기만 통과시키는 변이 → startup 칸의 최소 원소 단언(run id·리더 라벨·레인 라벨·Matrix 경로·docs URL)이 잡는다. 자동 모드 설명을 다른 필드 이름으로 옮겨 남기는 변이 → `--auto-leader`/`--auto-lane` 부재 단언이 잡는다. role core를 깎아 합본만 통과시키는 변이 → AC-FBM-007이 잡는다.

### AC-FBM-002
- **Given** M1~M3 적용 트리
- **When** `go test -count=1 -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' -v ./internal/hook/` 를 실행하면
- **Then** exit 0, 두 테스트가 `--- PASS`(SKIP 아님)이고, 출력이 en·ko·ja·zh 4개 언어 × startup·clear·compact 칸마다 팩토리 고정분 수치를 보고하며, 순회 칸 수 줄이 있다.
- RED-now: E2 — 기준 트리 출력에는 고정분 수치 줄이 없고(합성 생산자 4,797), 언어는 en·ko 두 개, 소스는 startup 하나뿐이다. `lane_en_glm` 9,946이 9,146 위인데도 PASS — 테스트가 실제 위험을 단언하지 않는다는 것이 RED 사유.
- Green 경로: M1이 테스트를 바꿔 기준 렌더에서 RED(E8 증거), M3가 녹색으로 뒤집는다.
- 변이 탐침: 상수만 바꾸고 생산자를 여전히 합성 문자열로 두는 변이 → 칸별 고정분 줄 부재로 실패. 순회가 0칸이 되는 변이 → 순회 칸 수 단언이 잡는다.

### AC-FBM-003
- **Given** M3 적용 트리
- **When** `grep -c autoModeGuide internal/hook/session_start_factory_i18n.go` 를 실행하면
- **Then** 출력 `0`, exit 1.
- RED-now: E3 — 출력 `11`, exit 0. 사유: 자동 모드 2플래그 설명 필드(구조체 2 + 4개 언어 × 2 + 주석)가 아직 테이블에 있다.
- Green 경로: M3.
- 변이 탐침: 필드 이름만 바꾸고 텍스트를 남기는 변이 → AC-FBM-001의 `--auto-*` 부재 단언과 짝으로 잡힌다.

### AC-FBM-004
- **Given** M5 적용 트리
- **When** `wc -c AGENTS.md` 를 실행하면
- **Then** 첫 필드가 24,276 이하.
- RED-now: E4 — `   24501 AGENTS.md`, exit 0. 사유: 여유 75바이트(< 300).
- Green 경로: M5 — 운영자 패치 문단 + `only` 복원, 측정 후보 24,221바이트.
- 변이 탐침: 아무 의무 문단이나 지워 크기만 맞추는 변이 → AC-FBM-005(배타성 문구)와 plan.md §B.5 의미 대조 체크리스트(sync 감사 대상)가 잡는다.

### AC-FBM-005
- **When** `grep -c "live only under" AGENTS.md` → **Then** 출력 `1`, exit 0 (기준 트리에서도 `1` — 보존 단언).

### AC-FBM-006
- **When** `go test -count=1 -run '^(TestCodexContractByteCeiling|TestCodexNestedTemplateDiscoveryBudget)$' -v ./internal/config/` → **Then** exit 0, 두 테스트 `--- PASS`(SKIP 아님), `contract document AGENTS.md = <≤24276> bytes (ceiling 24576, ...)`. `CodexContractByteCeiling` 값 24576 무변경.

### AC-FBM-007
- **When** `git log --oneline 59959629908cda723c37e26c7c15718194d6b38a..HEAD -- .claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/cross-session-messaging.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md` → **Then** 카드 t1923 커밋이 0건(main 흡수 전 카드 브랜치에서 판정; 흡수 뒤에는 다른 카드 커밋이 나타날 수 있으므로 커밋 메시지의 카드 id로 귀속을 확인).

### AC-FBM-008
- **Given** 착지 바이너리가 설치되고 6레인 + 리더가 startup으로 기동된 뒤
- **When** `python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/live_scan.py <claude projects dir> <설치 시각 ISO>` 를 실행하면
- **Then** 행 ≥ 7(lane ≥ 6, leader ≥ 1), 모든 행의 합본 ≤ 10000, 마지막 줄 `overflow 0`.
- 역사 관측(재실행 RED 아님): E8 — 10-21 기준 `count 15` / `overflow 10`(모두 05:42Z 이전, 역할 블록 18,367).

### AC-FBM-009
- **When** `go test -count=1 -run '^(TestSD_AC019_NextCardRuleInjection|TestLaneRuleCarriesStandingRecheckCron|TestFactoryWorkerNoticeCarriesSpawnAuthority|TestInjectedCoreCarriesObligationElements)$' -v ./internal/hook/` → **Then** exit 0, 네 테스트 모두 `--- PASS`. (plan 단계 미실행 — §D Gaps.)

## §C 증거 장부

### E1 — AC-FBM-001 RED-now
```text
command: go test -count=1 -overlay .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/overlay.json -run 'TestT1923BootBudgetProbe$' -v ./internal/hook/
exit: 1
tree: 59959629908cda723c37e26c7c15718194d6b38a
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-probe.txt
excerpt (verbatim lines from that file):
    zz_t1923_boot_probe_test.go:84: factory fixed part 3751 > ceiling 2847
    zz_t1923_boot_probe_test.go:87: composite 10006 > budget 9146
    zz_t1923_boot_probe_test.go:84: factory fixed part 2852 > ceiling 2847
    zz_t1923_boot_probe_test.go:123: PROBE swept=72
FAIL	github.com/modu-ai/moai-adk/internal/hook	4.193s
"--- FAIL" lines: 25 (grep -c -- "--- FAIL" on the raw file) = 24 failing subtests + 1 parent ("^--- FAIL" count 1)
```

### E2 — AC-FBM-002 RED-now
```text
command: go test -count=1 -run 'TestRoleInjectionAssemblyBudget$|TestCompositeBudgetRealAssembly$' -v ./internal/hook/
exit: 0
tree: 59959629908cda723c37e26c7c15718194d6b38a
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-intree.txt
excerpt:
    role_composite_e2e_test.go:90: RESULT lane_ko_glm: composite utf16=9030 (cap 10000, margin=970)
    role_composite_e2e_test.go:90: RESULT lane_en_glm: composite utf16=9946 (cap 10000, margin=54)
    role_composite_e2e_test.go:90: RESULT leader_ko: composite utf16=7882 (cap 10000, margin=2118)
    role_composite_e2e_test.go:90: RESULT leader_en: composite utf16=8806 (cap 10000, margin=1194)
    role_injection_budget_test.go:163: deployed tree: role factory-lane composite=9100 (budget 9146) core=4048 (ceiling 4092, target 3800)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.508s
red reason: no per-cell factory-fixed line, only en/ko, only startup; 9946 > 9146 passes.
```

### E3 — AC-FBM-003 RED-now
```text
command: grep -c autoModeGuide internal/hook/session_start_factory_i18n.go
exit: 0
stdout: 11
tree: 59959629908cda723c37e26c7c15718194d6b38a
```

### E4 — AC-FBM-004 RED-now
```text
command: wc -c AGENTS.md
exit: 0
stdout:    24501 AGENTS.md
tree: 59959629908cda723c37e26c7c15718194d6b38a
```

### E5 — AC-FBM-005 기준 관측
```text
command: grep -c "live only under" AGENTS.md
exit: 0
stdout: 1
tree: 59959629908cda723c37e26c7c15718194d6b38a
```

### E6 — AC-FBM-006 기준 관측
```text
command: go test -count=1 -run 'TestCodexContractByteCeiling$|TestCodexNestedTemplateDiscoveryBudget$' -v ./internal/config/
exit: 0
tree: 59959629908cda723c37e26c7c15718194d6b38a
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/red-now-5995962990-ceiling.txt
excerpt:
    token_budget_guard_test.go:109: contract document AGENTS.md = 24501 bytes (ceiling 24576, headroom 75)
    token_budget_guard_test.go:109: contract document internal/template/templates/AGENTS.md.tmpl = 23784 bytes (ceiling 24576, headroom 792)
--- PASS: TestCodexContractByteCeiling (0.34s)
--- PASS: TestCodexNestedTemplateDiscoveryBudget (1.38s)
```

### E7 — AC-FBM-007
기준 트리 자체가 범위 시작점이라 plan 단계 관측 대상 없음(빈 범위). run 단계에서 측정.

### E8 — AC-FBM-008 역사 관측
```text
command: python3 -I .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/live_scan.py /Users/goos/.moai/claude-profiles/moai-adk/projects 2026-10-21T00:00:00Z
exit: 0
stdout (raw file): .moai/specs/SPEC-FACTORY-BOOT-MINIMAL-001/probe/research/live-scan-20261021.txt
last lines:
2026-10-22T18:26:15.251Z leader 5762 1467 4295 False -Users-goos-MoAI-moai-adk-go/3936623d-62a7-4ce8-b44e-74ba0ebe94b8.json
count 15
overflow 10
```

### E9 — AC-FBM-009
plan 단계 미실행(§D Gaps). run 단계 E1 매트릭스에서 관측.

## §D Gaps / Residual-risk

- Gaps: AC-FBM-009 테스트 4개는 plan 단계에서 실행하지 않았다. 리더 수동 디스패치·Codex(gpt) owned-card 규칙 변형은 탐침에 없다(gpt 레인 규칙은 owned-card 경로로 lane 자동 규칙보다 짧다 — 측정 안 함).
- Residual-risk: 비-팩토리 허용분 1,950은 GLM 레인의 단일 라이브 세션 분해 기반이다. 다른 생산자 조합(예: degraded 알림이 여러 개 겹치는 경우)은 1,950을 넘을 수 있다. 봉투 밖 생산자는 여전히 넘침을 만들 수 있으며 폴백이 백스톱이다.

## §E 완료 정의

- release-blocking 4개(AC-FBM-001~004) 녹색, regression-guard 5개 관측 기록.
- run 단계 E8(RED 출력)에 M1 직후의 예산 테스트 실패 원문 포함.
- sync 단계에서 AC-FBM-008 라이브 확인 기록.
