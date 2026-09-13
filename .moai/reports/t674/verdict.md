# t674 Verdict — classifyError OOM "137" 부분문자열 오판 (t529 부수 발견)

Date: 2026-09-13 · Lane: lane-1 · Branch: `WT-oom-137-match` (develop `74d872aaf` 기점) · Tier S, Class B (run→sync)

## Claim

`classifyError`의 OOM 분기가 맨몸 부분문자열 `"137"`을 매치해, 카드 id를 딴 워크트리 경로(`.claude/worktrees/t137`)나 더 긴 숫자(`exit status 1375`)를 포함하는 무관한 실패를 OOMKilled로 오분류하던 결함을 경계 매치로 수리하고, 기존 분류 회귀를 보존한다.

## Changed files

- `internal/hook/post_tool_failure.go` — OOM 분기의 `strings.Contains(errorText, "137")` → 패키지 수준 컴파일 정규식 `oomExitCodeRe = \b137\b` (경계 매치). WorktreeGuardRefusal 선행 분기와 나머지 분기 순서 무변경.
- `internal/hook/post_tool_failure_oom_test.go` (신규) — 5케이스 경계 회귀
- `.moai/reports/t674/verdict.md` (본 문서)

## 수리 축 판정 (카드 후보 a vs b)

후보 (b) **구분자 경계 검사**(`\b137\b`)를 채택했다. 근거: 실제 OOM 이벤트 문자열("exit status 137", "exit code 137", "(137)")은 전부 공백/구두점으로 구분된 단독 토큰 형태라 경계 매치가 전부 흡수되고, 후보 (a)의 엄격한 형태(`exit (status|code) 137` 한정)는 "killed (137)" 같은 변형을 놓친다. 반면 "oom"/"out of memory" 부분문자열이 의미 신호의 주하중을 이미 운반하므로, 137의 역할은 보조 신호다 — 경계만으로 충분하다. `t137`은 `t`→`1` 사이 단어 경계가 없어(`\w` 연속) 매치되지 않고, `1375`는 `7`→`5`가 연속어로 배제된다.

## Evidence (본 run의 실측)

| # | Command | Observed |
|---|---------|----------|
| E1 | RED-first: `TestOOMExitCode137Boundary` 를 구형 트리에서 실행 | FAIL 2케이스 — `t137/run.sh: no such file...` → OOMKilled(오분류, 카드 재현 그대로) · `exit status 1375` → OOMKilled. 3케이스("exit status 137" 포함 t137 경로 동반 등) PASS |
| E2 | 수리 후 동일 테스트 | 5/5 PASS |
| E3 | 기존 분류 회귀 — 셀렉터는 RE2 교대(`TestOOMExitCode137Boundary`, `TestPostToolUseFailure`, `TestWorktreeGuard`, 백슬래시 없는 파이프 연결)로 go test 실행 | ok — 기존 "signal: killed (exit status 137)" → OOMKilled 픽스처 보존 확인. (교정: 이 셀의 명령 표기에 백슬래시가 섞인 것은 오기였다 — 실제 실행은 백슬래시 없는 교대로 3테스트 실제 선택·통과. 병합 트리 재측정은 교대 셀렉터에 `-v`를 붙여 PASS 수를 대조한다) |
| E4 | gofmt -l(빈 출력) + go vet ./internal/hook/(무소식) | 클린 |
| E5 | 패키지 전체 `go test ./internal/hook/ -count=1` | **ok 187.694s, exit 0** — 같은 실행에서 gofmt -l 무출력·go vet 무출력 확인 |

## 분기 순서 관련 관측 (카드 배경)

카드가 지적한 t529의 분기 순서(WorktreeGuardRefusal을 첫 분기로)는 유지 — refusal은 "명령이 실행조차 안 됐다"는 다른 종류의 주장이므로 순서 하중은 그대로 유효하다. 본 수리는 OOM 분기 내부의 매치 정밀도만 좁혔고 순서를 건드리지 않는다.

## Gaps

- E5 패키지 전체 실행은 완료 관측됐다(ok, exit 0) — 병합 창 재측정에서 동일 패키지를 다시 측정한다.
- `"oom"` 부분문자열 자체의 오탐 축(예: "room", "zoom" 포함 문구)은 본 카드 범위 밖 — 실측된 오분류 사례가 없어 손대지 않았다. (현실적 발화: "bedroom"류 단어가 실패문에 섞이는 경우 — 가능성 낮음, 기록만 남긴다.)
- 실제 OOM이 아닌 "port 137"(NetBIOS)류 문구는 경계 매치로는 여전히 OOMKilled다 — 이 도메인에서 관측된 바 없음, Residual 기재.

## Residual risk

- 경계 매치로도 의미 없는 단독 숫자 137이 실패문에 섞이면 오분류는 남는다 — 137 매치는 어디까지나 보조 신호이고, 근본적으로 숫자 한 개로 프로세스 사인을 추정하는 분기의 한계다.
- 본 수리 역시 릴리스 전까지 v3.1.x 사용자에게 전달되지 않는다(t680·t683과 동일 축).

🗿 MoAI
