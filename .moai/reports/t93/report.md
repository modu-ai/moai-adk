# t93 — 백그라운드 서브에이전트 Task 도구 제거 여부 실측

> 카드 원문(backlog t93): "[3.1.1][중대] 백그라운드 서브에이전트 Task 도구 제거 여부 실측 — MoAI 에이전트 11/11이 tools에 TaskCreate/Update/List/Get 선언. 공식 문서의 백그라운드 유지 도구 목록에 Task 계열 부재 → v2.1.198 기본 백그라운드부터 선언한 Task 도구가 제거된 상태일 수 있음(manager-kanban TaskList 조율 조용히 미동작 가능). 백그라운드 서브에이전트에서 TaskCreate 실제 호출로 확인. … 고치기 전에 재는 게 순서. Tier S."

## 판정

**제거 확인(서브에이전트 경로)** — 카드의 의심이 실측으로 확정됨.

| 경로 | Task 계열 도구 | 비고 |
|---|---|---|
| 무명 백그라운드 서브에이전트 (E2) | **부재** — TaskCreate/TaskUpdate/TaskList/TaskGet 모두 스키마에 없음 | ToolSearch 자체도 부재 → 사전 로드 불가. `Task`(구명)도 없음(근사치: `Agent`, `TaskStop`는 존재) |
| named 스폰 / teammate 경로 (E1) | **미측정** — 결과 반환이 불가해 보고 수집 못 함 | 공식 문서는 teammate 경로에 Task 도구가 추가된다고 명시(t92 §C.1 계보 참조) |
| 메인 세션 | 존재 | TaskList로 기존 태스크 #1~#5 관측(대조군) |

## 증거 (E2 원시 결과 — raw/e2-result.txt)

- (1) ToolSearch 호출: **불가** — 서브에이전트 스키마에 도구 자체가 없음(도구 부재가 곧 발견).
- (2) TaskCreate 스킵(전제 실패) — 마커 태스크 미생성(세션 TaskList에 잔류 없음, 대조 관측).
- (3) 자체 컨텍스트 도구 목록: TaskCreate ✗ / TaskUpdate ✗ / TaskList ✗ / TaskGet ✗ / Task ✗ / ListAgents ✗ / ToolSearch ✗ / **SendMessage ✓**(피어 `main`·`t92-probe-named` 주소 가능 확인) / `TaskStop` ✓(근사 도구).

## 파급

- MoAI 에이전트 11/11의 `tools:` Task 계열 선언이 백그라운드 경로에서 무효 — `manager-kanban`의 TaskList 조율은 백그라운드 스폰 시 조용히 미동작(카드 예측 확정).
- SendMessage은 잔존 → 파일 기반·메시지 기반 조율은 대안 가능.

## 권장 후속 (카드가 제시한 (a)/(b) — 본 카드는 측정만, 수정은 별도 카드)

- (a) Task 조율이 필요한 에이전트 스폰에 `background: false` 명시 — 단, 헌법 §Background Agent Execution은 `background:` 미설정 규율을 유지하므로 정책 정합 검토 필요.
- (b) 파일 기반 진행 기록(예: progress.md)으로 조율 경로 전환 — kanban 보드의 "완료는 읽는다" 원칙과 이미 정합.
- 어느 쪽이든 에이전트 `tools:` 정의 정리(Task 계열 제거 또는 경로 주석)가 뒤따름 — t92에서 `orchestration-mode-selection.md` §C.1 마지막 불릿에 측정 사실을 문서화해 둠.

## 재현

```bash
# E2: 무명 백그라운드 서브에이전트 스폰(프롬프트 전문: raw/e2-result.txt 부제 참조)
#   Agent(subagent_type: general-purpose, prompt: 'ToolSearch select:TaskCreate,TaskList →
#          TaskCreate 마커 → TaskList 확인 → 도구 가용성 자가 보고 → E2-DONE')
# 관측: 결과 없음? 아님 — 정상 반환했으나 내용이 "도구 부재"(위 표). 메인 세션 TaskList에 마커 없음(정합).
```

## Gaps / 잔여 위험

- teammate 경로 미측정(E1 무반환) — 공식 문서 명시만 근거.
- `background: false`가 도구 집합을 회복시키는지 직접 검증 안 함(옵션 (a) 채택 시 선행 실측 필요).
- 측정 시점 CC 2.1.233 — 버전업 시 재측정 권장.
