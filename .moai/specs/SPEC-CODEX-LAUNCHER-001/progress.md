# SPEC-CODEX-LAUNCHER-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-25

plan-phase 는 8차 반복으로 닫혔다. 궤적과 각 라운드의 판정서는 `.moai/reports/t197/` (`verdict-iter3` ~ `verdict-iter8`, `verdict-init-1` ~ `verdict-init-4`), 규칙 적용 기록은 같은 디렉터리 `rules-applied.md`.

마지막 확정 감사(`verdict-iter8.md`)는 FAIL 0.82 였고, 그 시점의 잔여 지적 중 **모순 3건** 은 착수 전에 정리했으며(운영자 판정) 나머지는 run-phase 부채로 인계됐다. 부채 목록은 아래 §E.2 에 기록한다.

mutant 표: 라운드 시작 시점 16/16 MUTANT-WRITABLE → 현재 13/16 MUTANT-FREE.

## §F Phase 4 Mode Selection

입력 파라미터:

| 항목 | 값 |
|---|---|
| tier | M |
| scope (파일 수) | 신규 2 + 수정 1 = 3 (`codex_launcher.go` · `codex_readiness.go` · `mcp_codex.go`) + 시험 파일 |
| domain count | 1 (Go — `internal/cli`) |
| file language mix | 100% Go |
| concurrency benefit | LOW — 코딩 중심 작업 |

모드 평가:

| 모드 | 선택 | 사유 |
|---|---|---|
| `direct` | 아니오 | 사소하지 않다 — 신규 커맨드 표면 + 분류기 재설계 |
| `serial` | **선택** | 코딩 중심 + 단일 도메인. Anthropic 의 coding-task 병렬성 유보에 따라 순차가 기본 |
| `fanout` | 아니오 | 다도메인이 아니고 연구 중심도 아니다 |
| `sweep` | 아니오 | 기계적 대량 변환이 아니다 — 새 코드 작성이다 |

Decision: serial

정당화: 단일 Go 패키지에 새 코드를 쓰는 작업이라 병렬화 가능한 독립 단위가 사실상 없다. M1 은 독립적 가치가 있고(웹 콘솔·MCP 도구의 auth 오표시가 그 자체로 해소된다) M2→M3 는 순차 의존, M4 는 마지막이므로 마일스톤 단위 순차 위임이 그대로 작업 구조와 일치한다. Implementation Kickoff Approval 은 통과했다 (운영자 판정, 리드 전달).

## §E.2 Run-phase Evidence

(마일스톤별로 아래에 append 한다.)

### 인계된 run-phase 부채

확정 감사가 남긴 지적 중 모순이 아닌 것들. 구현 중 닫히는 것과 못 닫는 것을 갈라 기록한다 — 못 닫으면 sync 로 넘기지 않고 그 시점에 보고한다 (리드 지침).

| id | AC | 내용 |
|---|---|---|
| D2 | AC-CL-008 · AC-CL-010 | REQ-CL-008 이 요구한 "거부된 `auth.json` → 명령 프로브 하강" 을 어느 AC 도 단언하지 않는다. 통합 3칸이 (유효 파일 / 파일 부재) 뿐이라 `readCodexAuthFile` 이 `ok=false` 를 돌려줄 때 러너를 부르지 않는 구현이 전 AC 를 통과한다 |
| D4 | AC-CL-012 | 무쓰기 관측 범위가 여전히 열거(격리 홈 트리 전체). `os.TempDir()` 을 거치지 않는 하드코딩 절대 경로 쓰기가 관측되지 않는다 |
| D5 | AC-CL-009 | 산문의 케이스 수(11)가 표의 데이터 행(13)과 어긋난다 — 단언은 표에 연동돼 있어 약화는 없다 |
| D6 | REQ-CL-004 vs AC-CL-011 | REQ 는 5행을 열거하고 AC 는 6라벨 폐집합을 요구한다 |
| D7 | plan §C.2 vs AC-CL-009 | 두 문서의 참조 문법이 다르다 (acceptance 판이 구속력을 가진다) |
| D8 | plan §C.3 | `CODEX_HOME` 부재 조치(`codex login`)를 어느 AC 도 판정하지 않는다 |

### 실측해야 할 것

- `internal/cli` 패키지 실행 시간 — 단독 실측 336초 기준선. 이 SPEC 이 시험을 얹으므로 M 단위로 재실측하고, `-timeout 1200s` 로도 모자라면 그 자리에서 상향해 기록한다.
- tty 왕복 — CI 에서 관측 불가. `os.Stdin`/`os.Stdout`/`os.Stderr` 값 항등만 단언하고 왕복은 Gap 으로 남는다. 실제로 깨지는 것이 관측되면 빌드 태그 문제가 운영자에게 되돌아간다 (spec.md 「판정 제외」).

### 중단된 위임 1건 (ledger 닫기)

M1 의 첫 `manager-develop` 위임이 사전 점검 단계에서 **중단** 됐다 — 반환값 없음, 사유는 계정 사용량 한도(작업 내용과 무관). 관측: 중단 직후 `git status --short` 에 그 위임이 만든 변경 0건, `internal/cli/` 에 신규 파일 0건, HEAD 무변동(`92987e653`). 즉 **아무것도 쓰지 않았고 재작업 대상도 없다.**

이 항목을 남기는 이유: 중단은 blocker report 와 다르다. blocker 는 돌아온 것이고 중단은 돌아오지 않은 것이라, 기록이 없으면 다음 읽는 사람이 "M1 이 한 번 돌았는데 결과가 없다" 로 읽는다.
