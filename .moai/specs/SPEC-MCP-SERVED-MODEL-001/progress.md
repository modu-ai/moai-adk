# Progress — SPEC-MCP-SERVED-MODEL-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-28
- card: t1284 (class C)
- tier: S — 산출물 spec.md, plan.md, acceptance.md, progress.md (카드 지시로 acceptance.md 분리)
- REQ: 8건 (REQ-MSM-001..008) · AC: 5건 (AC-MSM-001..005)
- 근거: `.moai/reports/t1284/premise.md` + spec.md §A.2 plan 단계 코드 판독
- 측정 트리: `e9577de4f` (WT-mcp-served-model)
- 결정: F1 범위 밖(plan.md §B.1) · glm_audit 범위 밖(§B.2) · codex 서빙 `unknown` 명시(§B.3) · doctor 점검 없음(§B.4)
- 미확인 항목(spec.md §A.3): z.ai 응답의 `model` 실제 채움 여부, codex app-server 응답의 모델 식별자 존재 여부
- open_clarifications: 0
- AC 스냅숏: 새 acceptance.md 추가는 방아쇠가 아님(`.moai/docs/ac-count-baseline-refresh.md` §2) — 스냅숏 재생성 없음

## §E.2 Run-phase Evidence

측정 트리: 구현 커밋 `b4fea44a2` (base `e9577de4f`, `git merge-base develop HEAD` = `e9577de4f`). 증거 파일은 모두 `.moai/reports/t1284/run/` 아래에 있다(gitignore 대상 — 워크트리에만 존재). 방식: TDD (cycle_type=tdd).

### RED (구현 전)

- 명령: `go test ./internal/cli/ -run '^Test[GC][a-zA-Z]*_ServedModel' -count=1 -v` → `01-red.txt`, exit 1.
- 관측: `--- FAIL: TestGLMTask_ServedModel` · `--- FAIL: TestGLMJob_ServedModel` · `--- FAIL: TestCodexTask_ServedModelUnknown` · `--- PASS: TestGLMAudit_ServedModelFieldIgnored`. 실패 사유는 모두 `served_model` 키 부재와 경고 부재다(예: `result carries no served_model key`). AC-MSM-004 는 무회귀 가드라 base 에서도 통과하는 것이 정상이다.

### AC 판정 (`^…$` 고정 셀렉터, `-v`, HEAD `b4fea44a2`)

| AC | 테스트 | 명령 출력 | 판정 |
|----|--------|-----------|------|
| AC-MSM-001 | `TestGLMTask_ServedModel` (행 a–h 8개) | `10-ac-msm-001.txt`: `--- PASS: TestGLMTask_ServedModel` · `ok`, exit 0 | PASS |
| AC-MSM-002 | `TestGLMJob_ServedModel` (행 a,b,c,e + 옛 형식 기록 d) | `11-ac-msm-002.txt`: `--- PASS: TestGLMJob_ServedModel` · `ok`, exit 0 | PASS |
| AC-MSM-003 | `TestCodexTask_ServedModelUnknown` (행 a–d) | `12-ac-msm-003.txt`: `--- PASS: TestCodexTask_ServedModelUnknown` · `ok`, exit 0 | PASS |
| AC-MSM-004 | `TestGLMAudit_ServedModelFieldIgnored` (문자열·숫자 `model`, 무 `model` 기준) | `13-ac-msm-004.txt`: `--- PASS: TestGLMAudit_ServedModelFieldIgnored` · `ok`, exit 0 | PASS |
| AC-MSM-005 | 재측정 묶음 (아래) | 아래 표 | **문언상 미충족** — 명령 1 이 HEAD 에서 exit 1. 유일한 실패는 base 에서도 재현되는 기존 불안정 테스트(아래). 판정은 리드 몫 |

네 파일 어디에도 `no tests to run` 은 없다.

### AC-MSM-005 재측정 묶음

| # | 명령 | 증거 | 결과 |
|---|------|------|------|
| 1 | `go test ./internal/cli/ -run 'GLM\|Glm\|Codex' -count=1` (HEAD `b4fea44a2`) | `03-glm-codex-final.txt` | exit 1 · `FAIL github.com/modu-ai/moai-adk/internal/cli 216.664s` · `--- FAIL` 1건 = `TestCodexTaskBackgroundHandshakeHonorsTaskBound` 뿐 · `no tests to run` 0건. 같은 HEAD 에서 그 한 테스트만 `-skip '^TestCodexTaskBackgroundHandshakeHonorsTaskBound$'` 한 실행: `03f-glm-codex-final-skip-flake.txt` → `ok … 198.085s`, exit 0 |
| 2 | `go test ./internal/cli/ -run TestRunDiagnosticChecks -count=1 -v` (리드 지정 형제 가드) | `02b-run-diagnostic-checks.txt` | `--- PASS` 6건 · `ok github.com/modu-ai/moai-adk/internal/cli 12.536s` |
| 3 | `go test ./internal/gitenv/... -count=1` | — | 해당 없음 — git 을 실행하는 테스트를 쓰지도 고치지도 않았다 |
| 4 | `go test ./internal/spec/... -count=1` | `04-internal-spec.txt` | `ok github.com/modu-ai/moai-adk/internal/spec 163.559s` |
| 5 | `go vet ./internal/cli/` | `05-go-vet.txt` | exit 0, 출력 없음 |
| 6 | `golangci-lint run ./internal/cli/...` (`~/go/bin/golangci-lint`, CI 핀 `.github/workflows/ci.yml:464` 과 같은 v2.1.6 — `06a-golangci-version.txt`) | `06-golangci-v2.1.6.txt` | `0 issues.` |

### 형제 비교기 미사용 (plan.md §B.1, acceptance.md §D)

- 판정: `git diff e9577de4f HEAD --output=.moai/reports/t1284/run/07-d3-diff-internal-cli.patch -- internal/cli` 후 `grep -nE '^\+.*(hook\.ObserveServedModel|expectedServedModel|servedMatches)'` → 빈 출력, `exit=1`. (워크트리 가드가 `git diff … | grep` 파이프를 거부해 diff 를 파일로 쓰고 grep 했다. 같은 판정식이다.)
- 대조: 같은 diff 파일에 바뀐 파일 7개(`+++ b/` 7행) — 판정이 빈 입력 위에서 돈 것이 아니다. `grep -cE 'hook\.ObserveServedModel|expectedServedModel|servedMatches' internal/cli/doctor_served_model.go` → `2`.
- 범위: `git diff --name-only e9577de4f HEAD` → 11개 파일(대조군 ≥1). `internal/hook/**`, `internal/auditreceipt/**`, `internal/cli/doctor*.go`, `internal/template/templates/**` 는 하나도 없다.

### 변이 대조 (AC-MSM-001 h · AC-MSM-004 b 의 이빨)

- HEAD 를 `git archive` 로 스크래치에 풀고 `glmMessagesResponse.Model` 을 `string` 으로 되돌린 변이체에서 네 테스트를 실행 → `14-mutant-typed-string.txt`, exit 1. 죽인 행은 정확히 `TestGLMTask_ServedModel/h_numeric_model_…` 과 `TestGLMAudit_ServedModelFieldIgnored/b_numeric_model` 두 개다. 라이브 트리는 건드리지 않았다.

### 기존 테스트 변경 1건

- `TestGLMTaskFactoryModeIgnoresModelOverride` 의 비공장 하위 테스트: 가짜 응답이 요청 모델(`caller-picked-model`)을 `model` 로 싣도록 픽스처 한 줄을 추가했다. 단언(`note == ""`)은 그대로다. 이유: REQ-MSM-004 에 따라 `model` 없는 응답은 부재 경고를 `note` 에 붙이므로, 옛 픽스처로는 "공장 안내가 없다"는 원래 의도를 잴 수 없었다. 수정 전 실패 출력: `03a-glm-codex-first.txt` (`note = "served model not reported: …", want none outside factory mode`), 수정 후: `03d-factory-test-fixed.txt` 두 하위 테스트 PASS.

### 형제 가드 셀렉터 안의 기존 불안정 테스트

- `TestCodexTaskBackgroundHandshakeHonorsTaskBound` (100 ms 경계 + 실제 자식 프로세스) 가 첫 묶음 실행에서 실패했다(`03a-glm-codex-first.txt`, `real child never reached the handshake`).
- 격리 3회: 이 카드 트리 `03b-handshake-isolated.txt` → PASS 1 / FAIL 2. base `e9577de4f` 를 스크래치에 풀어 같은 3회: `03c-handshake-on-base.txt` → PASS 1 / FAIL 2. 같은 모양이므로 이 카드와 무관한 부하 의존 불안정이다(측정 당시 load average 18–42). 중간 묶음 실행(`03e-glm-codex-intermediate.txt`)에서는 통과했다(`ok … 189.425s`).
- 무효 처리한 실행: plan §C 기준선 실행은 구현 편집과 시간이 겹쳐(일부 테스트가 라이브 트리로 바이너리를 빌드함) 오염됐으므로 기준선으로 인용하지 않는다.

### 필드 이름 확정 (plan.md §B.5 제안 그대로)

- GLM 동기 결과: `model`(요청, 의미 불변) + `served_model` + 경고는 기존 `note` 에 덧붙임.
- GLM 작업 기록: `model` + `served_model` + `served_model_warning`(omitempty). 완료 상태에서만 채운다.
- codex 결과·기록: `model`(omitempty, 요청 모델 — `resolveCodexModelEffort(turnParams).Model`) + `served_model` = `"unknown"`. 경고 필드 없음.

### @MX

- 추가 1건: `internal/cli/mcp_glm.go` `glmMessagesResponse` 위 `@MX:NOTE` (+ `@MX:SPEC`) — `Model` 을 `json.RawMessage` 로 두는 이유(glm_audit 과 공유). 새 공개 함수 없음, fan_in ≥3 신규 함수 없음 → ANCHOR·WARN 없음.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-28
run_commit_sha: b4fea44a2
run_status: pass-with-debt   # AC-MSM-005 명령 1 이 기존 불안정 테스트로 exit 1 (base 재현)
ac_pass_count: 4   # AC-MSM-001..004
ac_fail_count: 0   # AC-MSM-005 는 문언상 미충족이나 카드 귀속 실패 아님 — 리드 판정 대기
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run   # 레인 규율상 push 없음; HEAD·브랜치는 커밋 직전 재확인(30b1cff99 / WT-mcp-served-model)
l44_post_push_fetch: not-applicable   # push 하지 않음(리드 일괄)
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6 "0 issues.", go vet 무출력
cross_platform_build:
  darwin: measured-by-tests   # 이 머신에서 컴파일·테스트
  linux: not-measured
  windows: not-measured
total_run_phase_files: 8   # Go 7 + spec.md 상태 전이; progress.md 는 이 증거 커밋
m1_to_mN_commit_strategy: "구현 커밋 1 + 증거 커밋 1"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
