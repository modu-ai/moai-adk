# SPEC-MCP-SERVED-MODEL-001 — 인수 기준

---

## §A 판정 규약

1. 모든 명령은 워크트리 루트(`.claude/worktrees/t1284`)에서 실행한다.
2. `go test` 기반 기준은 종료 코드 0 **그리고** `-v` 출력에 `--- PASS: <테스트 이름> ` 줄이 있을 때만 PASS 다. `testing: warning: no tests to run` 또는 `[no tests to run]` 이 나오면 FAIL 이다.
3. 단일 테스트 기준의 `-run` 패턴은 `^…$` 로 고정한다. 형제 사례는 그 테스트 안의 표 기반 하위 테스트로 묶고, 아래 시나리오의 각 행이 표에 있어야 한다(행 누락은 FAIL).
4. 부정형 단언("경고가 없다", "바뀌지 않는다")을 담는 표는 같은 테스트 안에 양성 대조 행을 둔다.
5. 린트는 CI 판 golangci-lint v2.1.6 으로 잰다.
6. 기준 트리: 문서 수준 핀 `e9577de4f` 가 자체 핀이 없는 모든 기준에 적용된다.

---

## §B RED-now (문서 수준, base `e9577de4f`)

| 항목 | 값 |
|------|----|
| 명령 | `git grep -n -E 'served_model|ServedModel' -- internal/cli/glm_task.go internal/cli/glm_jobs.go internal/cli/mcp_glm.go internal/cli/codex_task.go internal/cli/codex_jobs.go` |
| stdout | (빈 출력) |
| exit | 1 |
| 양성 대조 | `git grep -c -E 'served_model|ServedModel' -- internal/hook/served_model.go` → `internal/hook/served_model.go:20` |
| 빨간 이유 | 다섯 파일 어디에도 서빙 모델 필드가 없다(요구사항이 겨누는 바로 그 부재). 이 작업이 필드를 더하면 초록으로 뒤집힌다 — 무관한 파일에 기대는 빨강이 아니다. |

AC-MSM-001..003 의 테스트는 base 에 존재하지 않으므로 `-run '^…$'` 는 규약 2 에 따라 FAIL 이다. 초록 경로는 각 기준의 "초록 경로" 줄에 적는다.

---

## §C 기준

### AC-MSM-001 — glm_task 동기 결과 표 (maps REQ-MSM-001, REQ-MSM-002, REQ-MSM-004, REQ-MSM-005)

- **Given** 가짜 z.ai 응답을 돌려주는 HTTP doer 와 표의 각 행:
  - (a) 요청 `glm-5.3-flash`, 응답 `model` = `glm-5.3-flash`
  - (b) 요청 `glm-5.3-flash`, 응답 `model` = `glm-5.3`
  - (c) 요청 `glm-5.3-flash`, 응답에 `model` 필드 없음
  - (d) 요청 `glm-5.3-flash`, 응답 `model` = `GLM-5.3-FLASH` (대소문자만 다름)
  - (e) 요청 `glm-5.3-flash`, 응답 `model` = `glm-5.3-flash[1m]` (접미사만 다름)
  - (f) z.ai 가 HTTP 500 을 돌려줌
  - (g) 공장 모드 모델 무시 안내가 이미 붙은 호출 + 응답 `model` = `glm-5.3`
- **When** 각 행으로 `glm_task` 를 동기 호출하면
- **Then** (a) 상태 completed, 서빙 모델 `glm-5.3-flash`, 서빙 경고 없음(양성 대조: 경고가 없어야 할 행) · (b) completed, 출력 불변, 서빙 모델 `glm-5.3`, 경고에 두 id 가 모두 있음 · (c) completed, 서빙 모델 빈 값, 부재를 말하는 경고 · (d) 경고 없음 · (e) 경고 있음(비교 전에 id 를 고치지 않음) · (f) 상태 failed, 서빙 경고 없음 · (g) 기존 안내 문장이 그대로 남고 서빙 경고가 덧붙음. 모든 행에서 기존 `model` 값은 요청 모델이다.
- **초록 경로**: M1 이 뒤집는다.

```bash
go test ./internal/cli/ -run '^TestGLMTask_ServedModel$' -count=1 -v
```

### AC-MSM-002 — 백그라운드 GLM 작업 기록 (maps REQ-MSM-003, REQ-MSM-004)

- **Given** 표의 각 행: (a) 응답 `model` = 요청과 같음 · (b) 응답 `model` 이 다름 · (c) 응답에 `model` 없음 · (d) 서빙 필드가 없는 옛 형식 기록 파일
- **When** 백그라운드 `glm_task` 가 완료되고 `glm_job_status` 로 기록을 읽으면(행 d 는 기록 파일을 직접 읽으면)
- **Then** (a) 기록에 요청·서빙 모델이 있고 경고 없음 · (b) 서빙 모델과 두 id 를 담은 경고 · (c) 서빙 모델 빈 값과 부재 경고 · (d) 오류 없이 읽히고 서빙 모델은 빈 값. (a)-(c) 모두 상태 completed, 출력 불변.
- **초록 경로**: M1 이 뒤집는다.

```bash
go test ./internal/cli/ -run '^TestGLMJob_ServedModel$' -count=1 -v
```

### AC-MSM-003 — codex_task 요청 모델과 서빙 `unknown` (maps REQ-MSM-006, REQ-MSM-007)

- **Given** 기존 가짜 codex 세션 픽스처와 표의 각 행: (a) 호출자가 `model` = `o4-mini` 를 지정한 동기 호출 · (b) 모델을 지정하지 않고 해석기도 빈 값을 내는 동기 호출 · (c) (a) 와 같은 백그라운드 호출
- **When** 각 행을 실행하면
- **Then** (a) 결과의 요청 모델 `o4-mini`, 서빙 모델 `unknown` · (b) 요청 모델 빈 값, 서빙 모델 `unknown` · (c) 작업 기록에 요청 모델 `o4-mini`, 서빙 모델 `unknown`. 어느 행에도 서빙 경고가 없고, 어느 행도 서빙 모델을 요청 모델과 같은 값으로 적지 않는다. 행 (a) 의 세션 송신 기록에서 `thread/start` 파라미터 `model` 이 `o4-mini` 임을 같은 테스트가 확인한다(요청 모델이 실제 송신값과 같은 해석기에서 나왔다는 양성 대조).
- **초록 경로**: M2 가 뒤집는다.

```bash
go test ./internal/cli/ -run '^TestCodexTask_ServedModelUnknown$' -count=1 -v
```

### AC-MSM-004 — glm_audit 무회귀 (maps REQ-MSM-005, REQ-MSM-008)

- **Given** 최상위 `model` 필드를 담은 z.ai 응답 봉투 안에 유효한 리뷰 JSON 이 든 경우와, `model` 이 없는 같은 봉투
- **When** GLM 리뷰 파서가 두 봉투를 파싱하면
- **Then** 두 결과의 verdict·summary·findings 가 같고, 리뷰 출력 타입에는 새 필드가 없다.
- **초록 경로**: M3. 기존 GLM 감사 테스트는 계속 통과한다.

```bash
go test ./internal/cli/ -run '^TestGLMAudit_ServedModelFieldIgnored$' -count=1 -v
```

### AC-MSM-005 — 재측정 묶음 (maps REQ-MSM-001, REQ-MSM-002, REQ-MSM-003, REQ-MSM-004, REQ-MSM-005, REQ-MSM-006, REQ-MSM-007, REQ-MSM-008)

- **Given** M1-M3 이 끝난 트리
- **When** 아래 명령을 차례로 실행하면
- **Then** 모든 명령이 exit 0 이고, 첫 명령의 출력에 `[no tests to run]` 이 없으며, 린트는 `0 issues.` 를 출력한다.

```bash
go test ./internal/cli/ -run 'GLM|Glm|Codex' -count=1
go test ./internal/spec/... -count=1
golangci-lint run ./internal/cli/...
```

조건부 형제 가드(해당 시에만 묶음에 추가하고, 추가 여부를 §E.2 에 적는다):

- `go test ./internal/cli/ -run TestRunDiagnosticChecks -count=1` — doctor 코드를 건드린 경우. 이 SPEC 은 doctor 를 건드리지 않는다(plan.md §B.4).
- `go test ./internal/gitenv/... -count=1` — git 을 실행하는 테스트를 건드리거나 새로 만든 경우. 계획상 해당 없음(가짜 HTTP doer 와 가짜 codex 세션만 쓴다).

린트 버전 확인: `golangci-lint version` 출력이 v2.1.6 이어야 한다. 다른 판이면 결과를 근거로 쓰지 않는다.

---

## §D 완료 정의

- AC-MSM-001..005 가 §A 규약대로 PASS.
- `internal/hook/**`, `internal/auditreceipt/**`, `internal/cli/doctor*.go`, `internal/template/templates/**` 가 변경되지 않았다: 카드 기여 범위 `git diff --name-only "$(git merge-base develop HEAD)"..HEAD` 에 이 경로가 없다(대조군: 같은 범위의 전체 파일 수가 1 이상).
- 새 비교 코드가 `internal/hook` 을 import 하지 않음을 소스 판독으로 확인해 progress.md §E.2 에 적는다(plan.md §B.1).
