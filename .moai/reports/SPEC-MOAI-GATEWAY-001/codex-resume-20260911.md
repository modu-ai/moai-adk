# Gateway 작업 재개 기록

- 세션: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`
- 작업 트리: `.claude/worktrees/moai-proxy-unified`
- 브랜치·기준 커밋: `WT-unified-gateway`, `81c1d58f9`
- 입력: 사용자가 전달한 Claude 세션 기록과 “이어서 moai-proxy-unified WT에서 작업진행하고 계획대로 모두 완료 하자” 지시.

## 범위와 완료 조건

중단된 6차 변경분 감사를 재개하고, PASS 뒤 gateway 코어의 구현과 해당 인수 조건 검증을 진행한다. 기존 env 정리·GLM 슬롯·PICKER 출시 결합 결정을 유지한다. 현재 지시를 감사 PASS 후 구현 진행 승인으로 기록한다. push·PR·병합·워크트리 제거는 별도 지시 범위다.

완료는 `acceptance.md`의 유효 AC 24개에 실제 실행 근거 또는 허용된 Gap 판정을 남기고, 변경 패키지 검증 및 독립 구현 감사를 마친 상태다. M0 인증 측정과 M1 원본 캡처 게이트를 통과한 것으로 가정하지 않는다. 계획에 명시된 중단 조건이 관측되면 의존 구현을 멈추고 그 근거를 기록한다.

## Claim

6차 변경분 감사가 PASS 0.92로 완료되었다. G5-B1~B3와 G5-A1~A4 해소, 차단 결함 0건, 선택 개선 G6-A1 1건이다. 제품 구현 완료 판정은 아니다.

## Evidence

현재 작업 트리에서 실행:

```text
git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway
go build -o /tmp/moai-gateway-81c1d58f9 ./cmd/moai
[stdout/stderr 없음, exit 0 — 캐시 권한을 허용한 재실행]
/tmp/moai-gateway-81c1d58f9 spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
claude --version
2.1.268 (Claude Code)
```

감사 보고서 `plan-audit-iter6.md`를 저장 후 다시 읽은 결과:

```text
Iteration: 6 — 운영자가 허용한 변경분 한정 재감사
Verdict: PASS
Overall Score: 0.92 (변경분 점수; Tier L 기준 0.85)
```

Python 정규식으로 현재 문서의 유효 정의를 센 출력:

```text
live_REQ 24
live_AC 24
```

## Baseline-attribution

SPEC 0.7.0, 위 커밋의 기존 추적 코드와 현재 추적되지 않은 SPEC 파일을 대상으로 측정했다. 전체 문서 해시는 `plan-audit-iter6.md`에 있다. `git fetch origin main` 성공 후 `origin/main...HEAD`는 `0 2879`였다. 별도 fetch하지 않은 공유 원격 추적 참조 `origin/develop...HEAD`는 `33 0`이었다. 기존 세션의 `0 0`을 현재 상태로 사용하지 않았고 브랜치를 병합하지 않았다.

## Gaps

- M0 OAuth·로컬 인증 공존과 refresh, M1 실제 요청 원문·버전별 인식 조건은 별도 실행 대상이다.
- 네 형제 SPEC 디렉터리는 현재 존재하지 않는다. 제안 상태이며 출시 의존성이 충족된 것으로 보지 않는다.
- 설치본 대신 현재 트리 빌드를 사용했다. 최초 빌드는 exit 0과 모듈 캐시 기록 권한 경고를 함께 냈고, 권한을 허용한 재실행에서는 경고 없이 exit 0이었다.
- 아직 제품 테스트·실계정 provider 전환·Windows 실행·구현 독립 감사를 완료하지 않았다.

## Residual-risk

계획 감사 PASS는 실제 client 호환성을 보증하지 않는다. 현재 client는 이전 측정의 2.1.267과 다른 2.1.268이다. PICKER·GPT-AUTH 출시 조건과 tmux teammate 잔여 위험은 승인된 계약대로 남는다.

## 후속 구현용 코드 위치 조사

읽기 전용 탐색 에이전트가 같은 커밋의 코드와 `plan.md` M2~M9를 대조했다. 아래는 구현 시 연결할 위치이며 런타임 결함 판정이 아니다.

| 책임 | 기존 연결 지점 | 구현 시 확인할 점 |
|---|---|---|
| 실행 분기 | `internal/cli/launcher.go`, `launchClaudeDefault`, `execOrSpawnClaudeFunc` | `--continue`의 별도 `exec.Command(...).Run()`은 최종 env 조립보다 앞서므로 양쪽에 같은 계약 필요 |
| POSIX·Windows 수명 | `internal/cli/launch_exec_posix.go`, `launch_exec_windows.go` | POSIX exec의 PID 각인, Windows profile lease 이전과 `os.Exit` 보존 |
| 프로세스 식별 | `internal/homestate/profile_lease.go`, `CurrentProcessFingerprint`, `ProbeProcessIdentity` | PID 재사용을 고려한 기존 식별 seam 재사용 |
| 설정 파일 | `internal/cli/settings.go`, `mutateSettingsLocal` | 기존 lock과 원자적 0600 쓰기 재사용 |
| GLM 설정·credential | `internal/cli/glm.go`, `resolveGLMModels`; `internal/glmcred/glmcred.go`, `Load` | 네 슬롯 매핑은 재사용하되 direct endpoint·tmux 변경을 하는 `setGLMEnv` 전체 호출은 분리 |
| 세션 시작 | `internal/hook/session_start.go`, `runSettingsChain`, `ensureGLMCredentials`, `ensureTeammateMode` | chain 전체에서 signal 우선순위·in-process 설정 판정 |
| tmux·세션 종료 | `internal/hook/glm_tmux.go`, `ensureTmuxGLMEnv`; `session_end.go`의 두 cleanup 함수 | gateway signal 때 쓰기 0회 판정 가능한 seam 필요 |
| 초기 provider 표시 | `internal/cli/kanban.go`, `exportKanbanLaunchFacts`; `internal/kanban/record.go`; `internal/web/widgets.templ`, `backendBadge` | 초기 provider 표시이며 실제 과금 표시로 해석하지 않음 |

탐색에서는 테스트·실계정 요청을 실행하지 않았다. M0·M1 담당과 쓰기 경로가 겹치지 않는다.

## 이번 실행의 최종 상태

- 6차 계획 변경분 감사: PASS 0.92.
- M0: INCONCLUSIVE. 별도 헤더와 Bearer 공존은 관측했으나 upstream 429 14건으로 정상 응답·refresh 미확인. AUTH_TOKEN 대조군은 시험 토큰 대체와 401 14건. `m0-auth-gate.md` 참조.
- M1: BLOCKED. 실제 TUI 검증 요청 원문에서 `stream` 키 부재 확인. 현재 SPEC은 키 존재와 false를 요구하므로 명시적 중단 조건 성립. 첫 turn·이후 turn·제목 요청을 함께 캡처했다. `m1-capture-gate.md` 참조.
- 진행 기록 E1~E3 갱신. 제품 코드·인식기·고정 testdata는 만들지 않았다. 마지막 `git diff --stat` 출력은 비어 있었고 상태는 기존 네 미추적 디렉터리였다.
- 마지막 현재 트리 바이너리 lint: `✓ No findings — all SPEC documents are valid`, exit 0. 형식 통과가 M1 의미 충돌을 해소하지는 않는다.

검토할 수정안: `stream` 키 부재와 명시적 JSON `false`를 구분하여 허용할지 정하고, `null`·문자열·`true`는 계속 제외한다. `max_tokens` 정수 1·user 메시지 하나·tools 부재/빈 배열 조건 및 실제 turn·제목 요청 대조를 유지한다. `design.md` §4.1과 AC-MG-003(c)의 누락 변형 기대값을 함께 수정하고 변경분 재감사 후 인식기 작업을 재개한다. 아직 이 수정안을 SPEC이나 코드에 적용하지 않았다.
