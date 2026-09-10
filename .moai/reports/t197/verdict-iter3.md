# t197 — plan-audit iteration 3 판정

| 항목 | 값 |
|---|---|
| 대상 | `SPEC-CODEX-LAUNCHER-001` (spec / plan / acceptance) |
| 내용 핀 | `6bfb076bc` (부모 `746177017`), 트리 깨끗함 — 감사 중 이동 없음 |
| 실행 | `mcp__moai__audit_multi`, `project_root=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t197`, `target=HEAD` |
| 일자 | 2026-08-24 |

## VERDICT — **FAIL**

**점수 0.63 / 1.00** (codex 백엔드 채점). Tier M 통과선 **0.80** 미달.

| 차원 | 점수 |
|---|---:|
| 명확성 | 0.75 |
| 완전성 | 0.50 |
| 시험 가능성 | 0.50 |
| 추적성 | 0.75 |

## 백엔드별 판정

| 백엔드 | 게이트 | verdict 필드 | 본문 판정 | 채택 |
|---|---|---|---|---|
| claude | required | pass | pass | pass |
| codex | required | **pass** | **FAIL 0.63 · 차단 5건** | **FAIL** (본문 채택) |
| glm | advisory | inconclusive | — | fail-open |

### 판정 필드 ↔ 본문 불일치 (수렴 엔진 결함, 2회 연속)

수렴 엔진이 낸 `overall_verdict` 는 `pass` 이고 `disagreement_flag` 도 `false` 다. 그러나 codex 백엔드의 **본문** 은 `FAIL — 0.63 / 1.00`, 5건 모두 `Status: Blocking` 이다. 엔진은 구조화된 `verdict` 필드만 읽고 본문을 읽지 않는다.

이 불일치는 **iteration 2 에서도 동일하게 발생했다** (당시에도 필드 `pass` / 본문 `FAIL — 병합 차단` 5건). 2회 연속 재현이므로 우연이 아니며, `audit_multi` 의 codex 어댑터가 verdict 를 추출하는 경로의 결함으로 본다. **이 판정서는 본문을 채택한다** — 필드를 근거로 통과시키면 관측하지 않은 통과 주장이 된다.

→ 별도 카드 후보: `audit_multi` codex 어댑터의 verdict 추출이 본문 판정과 어긋난다 (재현 2/2).

## 잔여 차단 지적 5건

| ID | 심각도 | 내용 | 위치 |
|---|---|---|---|
| D1 | High | `auth_mode=chatgpt` 의 조건이 "`tokens` 객체 존재" 뿐이라 `{"auth_mode":"chatgpt","tokens":{}}` 가 통과한다. stale-file 구멍 미폐쇄 | plan §C.2 표, AC-CL-008 |
| D2 | High | "비밀 필드 없는 구조체" 라 선언해 놓고 바로 아래 구조체가 `APIKey string` 으로 **키 전문을 역직렬화** 한다. AC 의 리플렉션 검사는 토큰 계열만 금지해 이걸 놓친다 | plan §C.2, AC-CL-008 |
| D3 | High | `codexLoginStatusRunner` 가 **이미 결합된** `combined []byte` 를 반환하므로 "stdout 비고 stderr 만 있음" 상태를 스텁으로 표현할 수 없다. 즉 이 SPEC 이 고치려는 baseline 결함(프로덕션의 stderr 폐기)이 회귀 시험을 우회한다 | plan §C.2, AC-CL-008 |
| D4 | Medium | AC-CL-008 이 "반환된 오류의 `Error()`" 를 검사하는데, 계획된 시그니처 `classifyCodexAuthFile(raw []byte) (string, bool)` 에는 오류도 경로 인자도 없다 — 작성 불가능한 AC | plan §C.2, AC-CL-008 |
| D5 | Medium | `measurement.md` 가 "모든 항목에 복사 실행 가능한 명령 + rc" 라 선언하지만 `${PIPESTATUS[…]}` 는 bash 전용이고 실행 셸에서는 unset, `time` 줄에 rc 없음, MCP 항목은 재실행 표현 없음 | measurement.md |

### 자체 확인

5건 전부 직접 확인했고 **5건 다 유효하다**. D5 는 보고보다 더 나쁘다 — 문제는 셸 호환성이 아니라, `${PIPESTATUS[…]}` 형태의 명령이 **내가 실제로 실행한 명령이 아니라 문서에 옮겨 적으며 재구성한 형태** 라는 것이다. rc 값 자체는 다른 형태의 명령에서 얻었다. "명령·rc·출력 전문" 이라는 선언과 실제 기록 방식이 어긋났다.

## 해결된 항목 (iteration 2 → 3)

| 항목 | 상태 |
|---|---|
| 전체 행 문법이 `Logged in state unavailable: API key missing` 반례를 거부하는가 | **해결** — AC-CL-009 11행 표가 그 반례를 직접 고정 |
| AC-CL-002 (argv·cwd·인용 보존) | **해결** — 해당 범위의 불완전 구현을 거부 |
| AC-CL-007 (sentinel 전파) | **해결** |
| measurement.md 미관측 절 명시 | **해결** |

## 근거 명령

```
$ git rev-parse --short HEAD
6bfb076bc

$ git status --short | wc -l
0

$ /tmp/moai-t197 spec lint .moai/specs/SPEC-CODEX-LAUNCHER-001/spec.md
✓ No findings — all SPEC documents are valid

REQ=14  AC=16   (Tier M 상한 16/16 이내)
```

감사 백엔드가 별도로 확인한 사항: 변경 범위는 SPEC/plan/acceptance/measurement 4개 마크다운뿐이며 scope creep 없음, 실제 자격 증명 값 노출 없음.

## 잔여 위험

구현 코드가 아직 없으므로 크로스 플랫폼 실행·subprocess 스트림 결합·앱 기동은 이 iteration 에서 관측하지 않았다. 이들은 run-phase 게이트 항목이다.

## 다음 단계

1. D1~D5 반영 (본 판정 이후 별도 커밋)
2. 운영자 신규 요구 — 배선 부재 시 초기화 + `CLAUDE.md ← @AGENTS.md` 규약 확보 — 를 REQ 로 접기
3. 최종 확인 감사 1회 → `verdict-final.md`
