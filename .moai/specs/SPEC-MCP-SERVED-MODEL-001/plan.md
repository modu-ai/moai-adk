# SPEC-MCP-SERVED-MODEL-001 — 구현 계획

## §A 맥락

- 카드: t1284 (Tier S, Class C). 워크트리 `.claude/worktrees/t1284`, 브랜치 `WT-mcp-served-model`, base `e9577de4f`.
- 전제: `.moai/reports/t1284/premise.md` (LIVE).
- 형제: SPEC-SERVED-MODEL-AUDIT-001 (t1282) — "기록하고 경고하되 거부하지 않는다" 자세를 그대로 따른다. 형제의 관측기(`internal/hook/served_model.go`)와 doctor 점검(`internal/cli/doctor_served_model.go`), 채택 거부 저장소(`internal/auditreceipt/store.go`)는 재사용하지 않는다(§B.1, §B.4).
- 개발 방식: TDD (quality.yaml 기본).

---

## §B 결정 기록 (되돌리기 어려운 순)

### §B.1 t1282 부채 F1 — 범위 밖 (결정)

**결정: 이 카드에 넣지 않는다. F1 은 별도 후속 카드로 남긴다.**

근거:

1. 이 카드의 비교는 **같은 백엔드 응답 안에서** 요청 id 와 서빙 id 를 맞대는 것이다. 형제의 `expectedServedModel`(`internal/hook/served_model.go:124`)과 `servedMatches`(`:145`)를 쓰지 않는다. 두 함수는 `internal/hook` 패키지의 비공개 함수이고, 기대값을 에이전트 선언·프로필 해석에서 끌어오며, `servedMatches` 는 Claude 계열 별칭(`opus` → `claude-opus-*`) 규칙을 담고 있다. GLM·codex 비교에는 그 어느 것도 필요 없다.
2. GLM 모델 id 는 `[1m]` 접미사를 달지 않는다. `internal/config/defaults.go:224-226` 이 밝히듯 z.ai 는 접미사 붙은 id 를 모르는 모델로 거부하므로, 요청 id 에 접미사가 있으면 호출이 완료되지 않고 서빙 비교 자체가 일어나지 않는다(REQ-MSM-005 는 완료되지 않은 호출에 경고를 붙이지 않는다).
3. codex 경로는 서빙 모델을 비교하지 않는다(`unknown` 고정, REQ-MSM-006).
4. 따라서 F1 은 이 카드가 내는 어떤 판정도 뒤집을 수 없다. 반대로 이 카드에서 접미사 정규화를 넣으면 REQ-MSM-005 의 "비교 전에 id 를 고치지 않는다"와 충돌하고, 형제 비교기의 동작 변경이 이 카드의 인수 범위 밖에서 일어난다.

run 단계 확인 의무: 새 비교 함수가 `internal/hook` 을 import 하지 않는다는 것을 `go list -deps` 가 아닌 소스 판독으로 확인해 §E.2 에 적는다(import 가 생기면 이 결정의 전제가 깨진 것이므로 blocker 로 보고).

### §B.2 glm_audit — 범위 밖 (결정)

`glm_audit` 도 `glmMessagesResponse` 를 공유한다(`mcp_glm.go:343`). 이 카드는 그 구조체에 `model` 필드를 **더하기만** 하므로 glm_audit 의 파싱 결과는 바뀌지 않는다. 그러나 glm_audit 의 결과에 서빙 모델을 **싣는 것**은 범위 밖이다. glm_audit 은 `ReviewOutput` 을 돌려주는데 이 타입은 codex_audit·glm_audit·audit_multi 의 수렴 계약이며, 여기에 필드를 더하면 세 백엔드와 수렴 로직, 그 소비자(plan-auditor·sync-auditor)의 판독 규약을 함께 건드리게 된다. 단순성 원칙상 별도 카드 몫이다. AC-MSM-004 가 공유 구조체 변경의 무회귀를 잰다.

### §B.3 codex 서빙 모델 — `unknown` 명시 (결정)

codex_task 는 app-server JSON-RPC 세션을 쓰고, 소비하는 응답 어디에서도 모델 식별자를 읽지 않으며, 테스트 픽스처의 `thread/start` 응답도 스레드 id 만 싣는다(spec.md §A.2). 모양이 확립되지 않은 필드(예: `result.model`)를 추측해 읽는 코드를 만들면 합성 픽스처로만 초록이 나는 검증 불가 코드가 된다. 그래서 서빙 모델은 `unknown` 을 명시 기록하고, 요청 모델만 실제 값으로 남긴다. 요청 모델은 세션이 `thread/start` 에 싣는 값과 같은 해석기(`resolveCodexModelEffort`, `mcp_codex.go:223`)로 구한다 — 새 해석 규칙을 만들지 않는다.

`unknown` 이 모든 호출에서 구조적으로 나오므로 호출마다 경고를 붙이지 않는다(REQ-MSM-007). 모든 호출에 붙는 경고는 정보가 없고 경고를 무시하는 습관만 만든다. `unknown` 값 자체가 "관측되지 않음"의 신호이며 `ok` 로 읽힐 수 없다.

후속 가능성(이 카드 아님): 실제 codex app-server 세션 한 건을 캡처해 `thread/start` 응답이나 턴 알림이 모델을 담는지 확립하면, 그 모양으로 서빙 모델 판독을 추가하는 카드를 낼 수 있다.

### §B.4 doctor 점검 — 넣지 않음 (결정)

형제의 doctor 사후 스캔은 서브에이전트 **트랜스크립트** 판독기다. MCP 위임의 흔적은 트랜스크립트가 아니라 동기 결과(휘발)와 `.moai/state/glm-jobs/`·`.moai/state/codex-jobs/` 의 작업 기록이다. 재사용할 부분이 없고, 새 점검을 만들면 binary_lag 허용 목록 등록 등 부수 비용이 따른다. 결과와 기록에 경고가 실리므로 관측 목적은 이 카드만으로 달성된다.

### §B.5 필드 이름 제안 (run 단계에서 확정)

- 동기 결과와 기록 모두 기존 `model` 키는 **요청 모델**의 의미를 유지한다.
- 새 필드: `served_model` (GLM 은 관측값 또는 빈 값, codex 는 `"unknown"`).
- 경고: GLM 동기 결과는 기존 `note` 에 덧붙인다(공장 모드 안내 등 기존 문장 보존). GLM 작업 기록은 기존에 note 필드가 없으므로 `omitempty` 경고 필드 하나를 더한다.
- codex 결과·기록: 요청 모델용 `model`(비어 있으면 생략 가능)과 `served_model` 을 더한다.

필드 이름은 JSON 계약의 일부이므로 run 단계가 바꾸려면 blocker 로 보고한다.

---

## §C 사전 점검 (run 단계 첫 행동)

```bash
git rev-parse --short HEAD
git branch --show-current
git grep -n 'glmMessagesResponse' -- 'internal/**/*.go'
go test ./internal/cli/ -run 'GLM|Glm|Codex' -count=1
```

마지막 명령의 출력은 기존 초록 기준선으로 §E.2 에 남긴다(좁은 셀렉터의 형제 가드 누락 방지 — AC-MSM-005 의 재측정과 같은 셀렉터).

---

## §D 마일스톤 (우선순위 순, 시간 추정 없음)

### M1 — GLM 서빙 모델 관측 (Priority High)

- RED: `TestGLMTask_ServedModel`(동기 표), `TestGLMJob_ServedModel`(백그라운드 기록) 작성. 가짜 HTTP doer 가 돌려주는 응답 봉투에 `model` 을 싣거나 빼서 표를 만든다.
- GREEN: `glmMessagesResponse` 에 `model` 필드, `callGLMTask` 가 텍스트와 서빙 모델을 함께 돌려주도록, 동기 결과·작업 기록에 서빙 모델과 경고.
- 비교: `strings.EqualFold(strings.TrimSpace(req), strings.TrimSpace(served))` 수준 — 별칭·접미사 처리 없음.

### M2 — codex 요청 모델 기록 + 서빙 `unknown` (Priority High)

- RED: `TestCodexTask_ServedModelUnknown` — 기존 가짜 세션 픽스처로 동기·백그라운드 양쪽 확인.
- GREEN: `CodexTaskResult`·`CodexJobRecord`·`codexJobSpec` 에 필드 추가, 요청 모델은 `resolveCodexModelEffort(turnParams).Model`.

### M3 — 무회귀와 린트 (Priority Medium)

- glm_audit 공유 구조체 무회귀 행(`TestGLMAudit_ServedModelFieldIgnored` 또는 기존 `parseGLMReview` 표에 `model` 포함 봉투 행 추가).
- `acceptance.md` §C 재측정 묶음 전체 실행.

---

## §E run 단계가 건드릴 파일

| 파일 | 변경 |
|------|------|
| `internal/cli/mcp_glm.go` | `glmMessagesResponse` 에 `model` 필드 추가(공유 구조체 — audit 은 읽지 않음) |
| `internal/cli/glm_task.go` | `callGLMTask` 가 서빙 모델 반환, `GLMTaskResult` 에 서빙 모델, 불일치·부재 경고, 백그라운드 완료 시 기록 갱신 |
| `internal/cli/glm_jobs.go` | `GLMJobRecord` 에 서빙 모델·경고 필드 |
| `internal/cli/codex_task.go` | `CodexTaskResult` 에 요청 모델·서빙 `unknown`, 기록 생성 시 전달 |
| `internal/cli/codex_jobs.go` | `CodexJobRecord`·`codexJobSpec` 에 요청 모델·서빙 모델 필드 |
| `internal/cli/glm_task_test.go` 또는 새 `internal/cli/glm_task_served_model_test.go` | M1 테스트 |
| `internal/cli/codex_task_test.go` 또는 새 `internal/cli/codex_task_served_model_test.go` | M2 테스트 |
| `internal/cli/mcp_glm_parse_test.go` (또는 `mcp_glm_test.go`) | M3 무회귀 행 |

건드리지 않음: `internal/hook/**`, `internal/auditreceipt/**`, `internal/cli/doctor*.go`, `internal/template/templates/**`, git 을 실행하는 테스트.

---

## §F 위험

| 위험 | 대응 |
|------|------|
| z.ai 가 `model` 을 싣지 않으면 모든 완료 호출에 부재 경고가 붙는다 | 의도된 동작이다 — 부재는 `ok` 가 아니다. 실측 후 경고 문구 조정은 후속 카드 |
| z.ai 가 요청 id 를 다른 표기(예: 대문자)로 돌려준다 | 대소문자 무시 비교로 흡수. 그 밖의 표기 차이는 경고만 붙고 호출은 성공 — 거짓 경고는 비용이 낮다 |
| 공유 구조체 변경이 glm_audit 을 깨뜨린다 | 필드 추가만 하며 AC-MSM-004 가 잰다 |
| 좁은 셀렉터가 형제 가드를 놓친다 | §C 사전 점검과 AC-MSM-005 가 `GLM|Glm|Codex` 전체를 재측정 |

---

## §G 금지 사항

- `internal/hook` 비교기 재사용 금지(§B.1 전제).
- codex 응답에서 모양이 확립되지 않은 필드 판독 금지(§B.3).
- 로컬 전체 스위트(`go test ./...`) 금지 — CLAUDE.local.md §4.
