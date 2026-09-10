# t448 probe-results — 지연 goroutine 부수효과 5축 재현 관측

카드: t448 (Class B). 트리: `WT-deferred-side-effects` @ develop `5a8449859` (ff 기반,
측정 시점 커밋). 도구: 본 워크트리에서 빌드한 `bin/moai` (LDFLAGS commit `5a8449859`,
버전 문자열 `t448-probe 5a8449859` — 도구 귀속). 플랫폼 darwin.
측정 일시 2026-09-03 08:50-09:04 KST.

측정 배경: 배차문의 `file_changed 0/5` 는 미귀속 수치로 확인됨 — 리드 본인이 생성한 값
(어느 레인의 실측도 아님, 리드 통보 2026-09-03). t216 기록(d3-mx-cold-start.md Gap 5)은
"asserted from code, not measured" 가 정직한 상태. 본 문서가 그 축의 첫 실측.

## Claim (축별)

1. **file_changed (MX 증분 스캔)**: die-at-exit 유실이 **실측 재현됨** — 직접 CLI 발화
   (`bin/moai hook file-changed`, **matcher 우회**) 10회 전부 산출물
   (`.moai/state/mx-index.json`) 미착지 (0/10). 인라인 대조군(대기하는 테스트)은 통과 →
   goroutine 로직 정상, **프로세스 수명 경합**이 메커니즘으로 확정.
   ⚠️ **정정 (2026-09-03, lane-5 t454 + 리드 교차측정 후)**: 초판의 "배포판 등록 경로는
   matcher와 확장자 게이트가 **서로소**라 goroutine 미생성" 주장은 matcher를 **리터럴로
   읽은** 것으로, **무조건적 주장으로서 철회**. matcher를 정규식으로 읽으면 `.`이 임의
   문자라 `(.env|…)` 가 `/env`("config/envkeys.go")·`/environment`("environment_test.go")
   에 매칭 — 교집합은 비어 있지 않음(본인 재확인 grep + lane-5·리드 실측 일치). CC
   런타임이 matcher를 정규식으로 적용하는지·무엇에 매칭하는지는 **미측정** — 필드 노출은
   "미검증"이 정직한 상태. 템플릿 내 다른 matcher들(`Write|Edit|Bash`,
   `startup|resume|…`)이 파이프 교대를 쓰는 관례는 정규식 해석 쪽을 지지하는 정황 증거.
2. **config_change (debounce+검증+reload)**: 유실 개념이 **성립하지 않음** — 산출물이
   애초에 비내구. reload 는 프로세스 메모리만 바꾸고 slog 는 훅 경로에서 `io.Discard`
   (internal/cli/logging.go:59-62). probe 2런(유효·무효 YAML) 모두 rc=0, 신규 파일 0개.
3. **notification**: 유실 재현 **안 됨** — 비동기 본체의 산출은 slog 한 줄뿐이고 그 싱크가
   `io.Discard`. 게이트를 전부 개방하고 올바른 cwd 에서 실행해도 stderr 0바이트(디버그
   레벨 포함). 유일한 내구 산출물(레지스트리 트레이스)은 goroutine 밖 **동기** 경로에서
   착지함을 관측(146B). 템플릿 미등록(RETIRE-OBS-ONLY) + 3중 게이트 → 필드 노출 0.
4. **task_created**: notification 과 동일 관측.
5. **edges refresh**: 본 카드에서 재측정하지 않음 — t435 실측(현장 0/60, probe2 착지/
   probe4 유실)이 그 축의 근거. lane-1 과의 측정 정합(2/153 vs 0/60)은 세 축 정합으로
   합의 완료(2026-09-03 메시지, 조건 2개 명시).

## Evidence

### E1 — file_changed 직접 CLI 발화 10회 (신규 프로젝트마다)

명령 형태 (i = 1..10, 리터럴 경로):

```
mkdir -p /tmp/t448-fc-$i
printf 'package main\n\n// @MX:NOTE: probe file for t448 P1\nfunc main() {}\n' > /tmp/t448-fc-$i/main.go
CLAUDE_PROJECT_DIR=/tmp/t448-fc-$i bin/moai hook file-changed \
  <<< '{"session_id":"t448-p1","file_path":"/tmp/t448-fc-$i/main.go","change_type":"modified","cwd":"/tmp/t448-fc-$i"}'
test -f /tmp/t448-fc-$i/.moai/state/mx-index.json
```

관측: `1 LOST` … `10 LOST` — **10/10 미착지**.

### E2 — file_changed 캡처 런 (run 6)

stdout `{}` (정상 빈 HookOutput), rc=0, stderr 0바이트, 사이드카 부재.
→ 오류 경로 아님. 정상 응답 후 프로세스가 종료하며 유실.

### E3 — file_changed 인라인 대조군 (프로세스가 goroutine 을 기다리는 경로)

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/ -run 'FileChanged' -count=1 -v
```

관측(발췌):
```
--- PASS: TestFileChanged_AsyncReturn_Under100ms (0.00s)
--- PASS: TestFileChanged_SideEffectsCompleted (0.00s)
INFO mx tag delta (async) summary="MX tag delta on …/tagged.go: ANCHOR=1"
PASS  ok github.com/modu-ai/moai-adk/internal/hook 0.609s
```
→ 대기하면 사이드카 기록 완료. E1 과의 유일한 차이는 프로세스 수명.

### E4 — matcher·확장자 게이트 (초판 "서로소" 주장 → 정정)

- 음성 probe: `.env` 경로 발화 → 사이드카 부재 ("NEG no-sidecar"). — 이 probe는 `.env`
  자체가 확장자 게이트를 통과 못 한다는 것만 재확인 (논점과 무관).
- 초판 코드판독: template matcher `.env|.envrc|.gitignore`(351-362행) vs
  `supportedExtensions` 21종(file_changed.go:21-43) — **리터럴** 해석으로 교집합 ∅.
- **정정**: 정규식 해석에서는 교집합 비어 있지 않음. 본인 재확인:
  `echo "internal/config/envkeys.go" | grep -E "(.env|.envrc|.gitignore)"` → 매칭
  (`.env` 의 `.`이 임의 문자 → `/env` 히트). lane-5(t454)의 Go regexp 실측과 리드의
  재측정도 동일: unescaped matcher → `['internal/config/envkeys.go',
  'internal/hook/environment_test.go']`.
- **미측정으로 남는 것**: CC 런타임이 matcher를 실제로 정규식으로 취급하는지, 매칭
  대상이 전체 경로/상대 경로/basename 중 무엇인지. 본 카드의 0/10 은 **직접 CLI 발화**
  (matcher 우회)라 이 Gap 을 닫지 못한다.
- 로컬 dev `.claude/settings.json` FileChanged 등록은 **matcher 없음**(327행) →
  dev 저장소에서는 모든 파일 변경에 발화 가능(배포판과 다름) — 이 부분은 초판과 동일.

### E5 — config_change 2런

유효 YAML:
```
{} rc=0  files_created_after_start=0
```
무효 YAML (`:\n  - [broken`):
```
{} rc=0  (신규 파일 0개)
```
→ 어떤 경우에도 내구 산출물 없음. reload 는 in-memory(config_change.go:114-127),
검증 경고는 slog→io.Discard.

### E6 — notification / task_created (게이트 전개방, P3c)

설정: `/tmp/t448-nt3/.moai/config/sections/system.yaml` (`hook:` 루트 — `systemFileWrapper`
는 루트 래퍼 없이 `hook:` 을 직접 받음, types.go:1526-1528. opt_in.enabled=true +
observability_events: [notification, taskCreated]) + observability.yaml enabled=true.
cwd 를 프로젝트로 맞춰 실행 (deps.Config 는 Getwd 기반, deps.go:129):

```
rc=0  stderr_bytes=0
logs: trace-t448-p3c.jsonl (146 bytes)
```

- 비동기 본체의 방출(slog.Info)은 stderr 에 **0바이트** — `io.Discard` (logging.go:59-62)
  를 관측으로 확인. `MOAI_LOG_LEVEL=debug` 에서도 0바이트(레벨 게이트 아닌 폐기).
- 레지스트리 트레이스는 **착지**(146B) — 트레이스는 Dispatch 동기 경로에서 기록
  (registry.go:124-211)되므로 goroutine 사망과 무관.
- 오경험 기록: `system:` 루트로 쓴 1차 시도(/tmp/t448-nt2)는 게이트 전체 거부로
  "미생성"을 잰 무효 측정이었음 — 스키마 확인 후 재측정. 빈 결과는 해석 전 입력
  검증이 필요하다는 교훈.

### E7 — notification / task_created 등록 상태

template settings.json.tmpl 에 `"Notification"`/`"TaskCreated"` 이벤트 키 없음
(RETIRE-OBS-ONLY 헤더 주석과 일치). CLI 레지스트리에는 등록(deps.go:258, 272).

## Baseline-attribution

- 모든 관측: 본 워크트리 `WT-deferred-side-effects` @ `5a8449859` (develop 팁 ff),
  빌드 `bin/moai` commit 스탬프 `5a8449859` — 측정 도구 = 측정 트리.
- 테스트 대조군(E3): 같은 트리, `-count=1` (캐시 무시).
- 직접 CLI probe(E1/E2/E4/E5/E6): `/tmp/t448-*` 신규 프로젝트, 프로세스 수명은
  실제 훅 호출과 동일한 단발 CLI.

## Gaps

- **file_changed 착지율의 분모 확대 미실시** — 10회 표본. "0/10" 은 착지율 상한 추정이지
  정밀한 확률이 아니다(메커니즘상 join bound 가 없어 ~0 에 수렴하는 구조이지만, 본
  측정만으로는 상한만 말할 수 있다).
- **Claude Code 가 실제로 FileChanged/TaskCreated/Notification 이벤트를 발화하는지**는
  본 카드에서 검증 불가(CC 런타임 관측 필요). 본 측정은 "발화되었다면"의 조건부.
- **CC 런타임의 matcher 해석(정규식 여부·매칭 대상 문자열)은 미측정** — 본 카드 0/10 은
  `moai hook file-changed` 직접 발화로 matcher 를 우회했으므로 이 Gap 을 닫지 못한다.
  t454(lane-5) 축.
- edges 축 재측정 없음(t435 근거 인용). CC 훅 프로세스 수명 분포 자체는 t435 와 동일하게
  미측정(현장 간접 근거).
- windows/linux 에서의 수명 특성 미측정(darwin 단일).

## Residual-risk

- **처방 함의**: 5축 중 die-at-exit 유실이 "처방의 대상"이 되는 것은 edges(session_start,
  처방은 M4 패턴 이전 추천 — join bound 내 재배치는 t435 지연실측 6.01s와 충돌해 정정).
  file_changed 는 유실이 실측됐으나 **배포판 노출 여부는 matcher 런타임 해석 미측정으로
  미결**(초판 "dead" 주장은 리터럴 해석 한정으로 격하, E4 정정 참조) — 게이트 정합 판정은
  t454 의 런타임 측정을 기다리는 것이 순서.
  config_change·notification·task_created 는 "완료해도 남지 않는" 별도 성격으로,
  비동기 전제 자체의 정당성 문제(REQ-HAE-002/003/004 의 ≤100ms AC 와 충돌)라 리드·
  운영자 판정 영역.
- **주석 정정 소관**: file_changed.go 의 "it runs to completion or the asyncDeadline
  expires" 세 번째 경 누락은 t454(lane-5)가 이미 수정 — 본 카드 미중복. task_created.go
  8-9행의 "JSONL append" 서술은 존재하나 조건부(slog 핸들러가 JSONL append 를 담당한다는
  서술)이며 훅 경로에 JSONL slog 핸들러 설치 코드가 없어(logging.go:73 이 유일한
  SetDefault 사이트, TextHandler) 기술하지 않는 메커니즘 — 정정 후보는 이 파일 한정.
- dev 로컬 settings 의 matcher-less FileChanged 등록은 dev 저장소 한정 상태이며,
  배포 사용자에게는 해당 없음(다만 dev dogfood 환경에서는 유실이 실제 일어나는 유일한
  필드 지점일 수 있음).
- t216(session_start.go 미푸시 12커밋, 시그니처 변경)과의 의미 충돌 자리는 본 카드가
  코드를 고칠 경우 그대로 적용 — 코드 변경 전 리드 시퀀싱 판정 필요.
