# SPEC-HOOK-DIAG-SINK-001 — 인수 기준 (v0.3.3)

> **개정 사유 (v0.3.0)**: 감사 D2·D3. 두 가지를 고쳤다.
>
> **(D3) 존재 확인의 적용 범위.** v0.1.x 의 판정 원칙 3·4 와 AC-HDS-001 은 이미 옳은 형식
> (`-list` 존재 확인 + `-run` 실행)을 갖고 있었으나, **그 형식이 적용된 AC 는 14개 중
> 1개뿐**이었다. 나머지 테스트 기반 AC 는 `-run` 종료코드만으로 판정했고, 그 명령은 가드가
> 존재하지 않는 미수정 트리에서 exit 0 으로 통과한다. 결함은 **형식의 부재가 아니라 적용
> 범위**였으므로, AC-HDS-001 의 형식을 그대로 나머지에 확산했다. 접혀 있던 AC 네 건
> (003 · 007 · 008 · 012)은 전용 가드를 받아 각자의 이름으로 앵커한다.
>
> **(D2) 역방향 검사의 부재.** 새 이름이 자유로운지 앞으로 묻는 부재 검사만 있었고,
> **바꾸려는 심볼에 대해 기존 테스트·주석이 무엇을 단언하고 있는지** 뒤로 묻는 검사가
> 없었다. 그 결과 `internal/cli/logging_test.go` 가 이 SPEC 이 의도적으로 뒤집는 단언을
> 들고 있다는 사실이 초판에 실려 있지 않았다(§ 기존 단언과의 충돌). 판정 원칙 6 으로
> 승격한다.
>
> 측정 기록: `.moai/reports/t1144/ac-baseline.md`. 그 보고서는 자기 정정을 포함한다 —
> "테스트 기반 AC 4개가 전부 공허" 라는 초판 진단이 철회되고, AC-HDS-001 은 이미 존재
> 확인을 갖추고 있었음이 전수로 확인됐다.

## 판정 원칙

1. **내용 앵커만 쓴다.** 절대 라인 번호와 커밋 SHA 는 판정에 쓰지 않는다. `spec.md` §1.1 의
   수치는 base `60017eb83` 에 귀속된 **측정 기록**이지 판정 기준이 아니다.
2. **호출이 아니라 효과를 본다.** "라우팅이 설치됐다"가 아니라 "레코드가 파일에 있다"를
   단언한다.
3. **[HARD] 테스트 기반 판정은 반드시 두 단이다 — 존재 확인이 먼저다.**

   ```
   # 1단 존재: 이름이 실제로 출력되어야 한다
   go test <pkg> -list '^<TestName>$' | grep -cE '^Test'    = 1
   # 2단 실행: 그 다음에야 통과 여부를 본다
   go test <pkg> -run  '^<TestName>$' -count=1
   ```

   **두 단이 모두 만족해야 충족이다.** 1단이 0 이면 2단 결과와 무관하게 미충족이다.
   1단 없이 2단만 보면 "가드가 아직 없는 상태"와 "가드가 있고 통과하는 상태"가 구별되지
   않는다 — `go test -run` 은 매칭되는 테스트가 없을 때 실패가 아니라 성공하므로,
   run-phase 가 테스트 이름을 바꾸거나 오타를 내도 조용히 통과한다.

   **공허 방지**: `| grep -c .` 는 후행 `ok <pkg>` 줄 때문에 무조건 ≥ 1 이므로 쓰지 않는다.
   반드시 `grep -cE '^Test'` 다. `grep -c` 는 0 건일 때 종료코드 1 을 내므로 이 형식은
   게이트로도 그대로 쓸 수 있다.

   **면제는 네 건뿐이고, 전부 테스트 이름을 선택자로 쓰지 않는다**: AC-HDS-011(grep —
   미수정 트리에서 이미 exit 1 로 변별), AC-HDS-013(크로스 빌드), AC-HDS-014(패키지 전체
   스위트 — 선택자가 없으므로 빈 스윕이 성립하지 않는다), AC-HDS-016(grep — 주석 갱신
   판정). 테스트 **이름**을 `-run` 에 거는 AC 는 예외 없이 두 단이다.

   기계적으로 셀 수 있다 — 16개 AC 중 **12개**가 두 단을 갖고 나머지 4개가 면제다:

   ```
   awk '/^## /{ac=""} /^### AC-HDS-/{ac=$2} /-list/{if(ac!=""){print ac; ac=""}}' acceptance.md
   ```

   > 이 awk 의 `/^## /{ac=""}` 는 **필요하다**. 그것이 없으면 AC 블록을 벗어난 산문
   > (§ Gaps 의 "`-list` 존재 확인은 이름을 지킬 뿐…")이 직전 AC 의 것으로 잘못 집계된다 —
   > 실제로 AC-HDS-016 이 그렇게 거짓 양성으로 잡혔다.

4. **`-list` 와 `-run` 은 같은 정규식을 쓰고**, 둘 다 정확한 이름으로 앵커한다(`^…$`).
   넓은 패턴으로 존재를 확인하고 좁은 패턴으로 판정하면 확인이 판정을 구속하지 못한다.
5. **양성 대조 없는 0 은 부재의 증거가 아니다.** 부재 팔만 재면 "아무것도 출력하지 않는
   명령"과 구별되지 않는다. 판정식을 채택하기 전에 **실재하는 테스트**로 한 번 발화시킨다.
6. **[HARD] 역방향 검사 — 바꾸려는 심볼의 기존 단언을 먼저 센다.**

   ```
   grep -rn '<심볼>' --include='*.go' internal/
   ```

   **새 이름의 부재 검사는 이 검사를 대신하지 않는다.** 부재 검사는 "내가 쓸 이름이
   자유로운가"를 앞으로 묻고, 역방향 검사는 "이미 무엇이 이 심볼에 대해 단언하고 있는가"를
   뒤로 묻는다. 두 질문은 서로를 함의하지 않으며, 후자를 빠뜨리면 **의도적으로 뒤집는
   단언**이 SPEC 에 실리지 않은 채 run-phase 에서 예상 못한 빨강으로 나타난다. 이 SPEC 의
   실제 적발 내역은 § 기존 단언과의 충돌에 있다.

7. **사전 존재 부채는 델타로 판정한다.** 무관한 기존 매치를 "0건"으로 요구하는 AC 는
   통과 불가이며 범위 밖 편집을 강요한다.
8. **판정 불가한 런타임 대조는 소스 수준 가드로 대신한다.** `internal/cli/logging_test.go`
   `TestLoggingNeverTargetsStdout` 의 주석이 기록하듯, `go test ./...` 아래에서 이 저장소의
   테스트 바이너리는 **두 스트림에 같은 `*os.File` 을 받는다** — 측정 결과 `dest == os.Stdout
   == os.Stderr` 가 같은 포인터였다. 따라서 런타임 동일성 검사는 프로덕션 선택이 옳은데도
   거짓 실패를 낸다. 그 자리는 (a) **결정값 단언**(`resolveLoggingDecision` 의 `dest`)과
   (b) **소스 토큰 부재 가드**가 대신한다. 이 저장소가 이미 쓰는 방식이며 새 발명이 아니다.

   **[HARD] 정적 축의 판독 범위는 훅 경로 목적지를 구성하는 파일 집합 전체이고, 요구
   토큰은 파일마다 다르다.** v0.1.x 에서 정적 축은 런타임 대조의 보조였으나 v0.2.0 에서
   **유일한 담지자**가 됐다. 그런데 기존 `TestLoggingNeverTargetsStdout` 의 판독 범위는
   `logging.go` 한 파일이므로, 이 SPEC 이 **새로 쓰는 싱크 파일은 그 범위 밖**이다 —
   새 파일이 `os.Stdout` 이나 `os.Stderr` 를 참조해도 정적 축이 침묵한다. 따라서:

   | 파일 | 금지 토큰 | 이유 |
   |---|---|---|
   | `internal/cli/logging.go` | `os.Stdout` 만 | **비-훅 경로가 `os.Stderr` 를 정당하게 쓴다**(`resolveLoggingDecision` 의 비-훅 분기). 여기에 `os.Stderr` 부재를 요구하면 옳은 코드가 실패한다 |
   | 싱크 writer 선언 파일 | `os.Stdout` **과** `os.Stderr` 둘 다 | 훅 경로 전용이므로 두 스트림 어느 쪽도 정당한 목적지가 아니다 |

   이 비대칭은 **의도된 것**이며, 같은 규칙을 두 파일에 적용하는 쪽이 오답이다. 집합은
   파일이 늘어날 때 조용히 빠지지 않도록 명시 열거하거나 디렉터리 단위로 읽고, 열거한
   파일이 없으면 **건너뛰지 말고 실패**해야 한다(AC-HDS-004).

## 가드 이름 (이 SPEC 소유)

| 가드 | 테스트 이름 | 패키지 | 판정 AC |
|---|---|---|---|
| 1 | `TestHookPathWarnRecordReachesSink` | `internal/cli` | AC-HDS-001 |
| 2 | `TestHookPathHookDestIsNeitherStdStream` | `internal/cli` | AC-HDS-004 · 009 |
| 3 | `TestHookPathSinkFailureIsFailOpen` | `internal/cli` | AC-HDS-005 · 006 |
| 4 | `TestHookSinkIsPrunedAtSessionEnd` | `internal/hook` | AC-HDS-010 |
| 5 | `TestHookSinkConcurrentAppendIsIntact` | `internal/cli` | AC-HDS-003 |
| 6 | `TestHookSinkLevelBoundary` | `internal/cli` | AC-HDS-007 · 008 |
| 7 | `TestHookSinkIsNotCreatedWithoutRecords` | `internal/cli` | AC-HDS-012 |

> 가드 5·6·7 은 v0.3.0 신설이다. v0.1.x 에서 AC-HDS-003 · 007 · 008 · 012 는 가드 1 또는
> 가드 3 에 **접혀** 있었고, 넷이 글자 그대로 같은 명령 한 줄을 공유했다 — 접힌 단언이
> **작성됐는지조차** 가르지 못하는 상태였다. 특히 AC-HDS-012(지연 개방)가 가드 3(개방 **실패**
> 경로)에 얹혀 있어, 지연 개방을 보지 않아도 통과했다. 지연 개방은 `plan.md` §B-1 이 그
> 설계를 고른 **유일한 근거**(REQ-HDS-007)이므로 전용 가드가 필요하다.
>
> 가드 2 의 이름은 v0.1.x 의 `TestHookPathLeavesStdoutAndStderrUntouched` 에서 바뀌었다 —
> 옛 이름은 판정 원칙 8 이 금지하는 런타임 바이트 대조를 약속하는 이름이었다.

**[HARD] 일곱 가드 전부 존재 확인을 갖는다.** 따라서 "이 AC 들은 미수정 트리에서 통과할 수
없다(공허 GREEN 불가)"는 단정은 이제 **일곱 가드가 판정하는 AC 전부에 대해 참**이다 —
v0.1.x 에서 AC-HDS-001 에만 참이었던 것을 확산으로 해소했다. 단정이 닿지 않는 범위는
판정 원칙 3 의 면제 4건(AC-HDS-011 · 013 · 014 · 016)과 AC-HDS-015 뿐이다 — 015 · 016 은
**사전 존재 자산에 대한 델타 판정**이라 성격이 다르다(미수정 트리에서 이미 통과하므로
"공허 GREEN 불가"가 적용될 대상이 아니며, 각자의 델타 축이 그 자리를 대신한다:
015 는 `grep` 3 유지, 016 은 `card t1144` 문구).

## 미수정 트리 기준선 (실측)

측정 트리 `.claude/worktrees/t1144`, HEAD **`9b973bc04`** — SPEC 커밋만 올라간 상태이며
`internal/` 에 구현 변경이 없다.

> **SHA 재핀 주석**: 측정 자체는 선행 커밋 `2f63958a9` 에서 수행했고, 그 커밋은 이후
> amend 되어 `9b973bc04` 가 됐다. 두 커밋의 `internal/` 트리는 **바이트 동일**하다 —
> `git diff --stat 2f63958a9 9b973bc04 -- internal/` 출력 없음, 전체 차이는 이 SPEC 의
> 문서 3개뿐. 아래 모든 명령은 `internal/` 만 대상으로 하므로 측정값은 그대로 유효하며,
> 살아 있는 커밋으로 다시 핀했다.

### (a) `-run` 단독 판정의 공허 통과 — D3 의 원인

```
$ go test ./internal/cli/... -run '^TestHookPathWarnRecordReachesSink$' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli/update/report  0.166s [no tests to run]
ok  github.com/modu-ai/moai-adk/internal/cli/wizard         0.177s [no tests to run]
ok  github.com/modu-ai/moai-adk/internal/cli/worktree       0.168s [no tests to run]
EXIT=0
```

`-run` 단독 명령은 네 가드 모두 exit 0 이다. **AC 단위의 공허 여부는 그 AC 가 존재 확인을
갖는지에 달려 있었다**: AC-HDS-001 은 존재 확인을 갖고 있어 막혔고, 공허했던 것은
AC-HDS-004 · 006 · 010 과 거기 접혀 있던 하위 케이스들이다(`ac-baseline.md` 자기 정정본).

### (b) 존재 확인 — 일곱 가드 전수 부재 실측

```
$ go test ./internal/cli/...  -list '^TestHookPathWarnRecordReachesSink$'       | grep -cE '^Test'   → 0   # 가드 1
$ go test ./internal/cli/...  -list '^TestHookPathHookDestIsNeitherStdStream$'  | grep -cE '^Test'   → 0   # 가드 2
$ go test ./internal/cli/...  -list '^TestHookPathSinkFailureIsFailOpen$'       | grep -cE '^Test'   → 0   # 가드 3
$ go test ./internal/hook/... -list '^TestHookSinkIsPrunedAtSessionEnd$'        | grep -cE '^Test'   → 0   # 가드 4
$ go test ./internal/cli/...  -list '^TestHookSinkConcurrentAppendIsIntact$'    | grep -cE '^Test'   → 0   # 가드 5
$ go test ./internal/cli/...  -list '^TestHookSinkLevelBoundary$'               | grep -cE '^Test'   → 0   # 가드 6
$ go test ./internal/cli/...  -list '^TestHookSinkIsNotCreatedWithoutRecords$'  | grep -cE '^Test'   → 0   # 가드 7
```

일곱 전부 1단에서 0 → **미충족으로 올바르게 판정**된다. (가드 2 는 v0.1.x 이름
`TestHookPathLeavesStdoutAndStderrUntouched` 로도 0 이었다.)

### (c) 양성 대조 — 이 형식이 눈멀지 않았는가 (두 패키지 각각)

```
$ go test ./internal/cli/...  -list '^TestLoggingHandlerSelection$' | grep -cE '^Test'
1
$ go test ./internal/hook/... -list '^TestPruneObservationLogsIncludesAgentModelAudit$' | grep -cE '^Test'
1
```

부재 0 / 존재 1 로 두 팔이 갈린다. 가드 4 가 사는 `internal/hook` 에서도 갈리므로 두
패키지 모두에서 판정식이 성립한다.

### (d) grep 기반 판정식 — 원래부터 변별한다

```
$ grep -rn '"\.moai/logs/hook-runtime\.log"' --include='*.go' internal/ | grep -v _test.go
EXIT=1   (무매치)
```

## 기존 단언과의 충돌 (판정 원칙 6 의 실제 적발 내역)

역방향 검사를 이 SPEC 이 바꾸는 심볼 전체에 돌린 결과다. HEAD `9b973bc04`.

### (1) 테스트 — M1 에서 반드시 FAIL 한다

```
$ grep -rn 'resolveLoggingDecision' --include='*.go' internal/ | grep _test
internal/cli/logging_test.go:82:  got := resolveLoggingDecision(tc.args)
internal/cli/logging_test.go:84:  t.Errorf("resolveLoggingDecision(%q).dest = ...")
```

`TestLoggingHandlerSelection` 의 훅 케이스 3건이 목적지를 `io.Discard` 로 단언한다:

| 케이스 이름 | 인자 | 현재 기대값 |
|---|---|---|
| `hook_discards` | `["hook", "pre-tool"]` | `io.Discard` |
| `hook_behind_a_flag_discards` | `["--verbose", "hook", "stop"]` | `io.Discard` |
| `hook_ignores_log_level_env` | `MOAI_LOG_LEVEL=debug`, `["hook", "session-start"]` | `io.Discard` |

REQ-HDS-001 이 훅 목적지를 바꾸므로 **이 3건은 M1 에서 반드시 빨개진다.** 회귀가 아니라
**의도된 계약 변경**이며, run-phase 는 기대값을 갱신한다(AC-HDS-015). 빨개진 것을 보고
구현을 되돌리는 것이 이 자리에서 가능한 가장 나쁜 대응이다.

`hook_ignores_log_level_env` 는 특히 주의한다 — 이 케이스의 취지(**`MOAI_LOG_LEVEL` 이 훅
경로의 stdout/stderr 를 열지 못한다**)는 REQ-HDS-002 로 **그대로 살아남는다**. 기대값만
`io.Discard` → 싱크로 바뀌고, "환경변수가 목적지를 바꾸지 못한다"는 단언은 유지된다.

### (2) 주석 — M1 에서 거짓이 된다 (테스트는 깨지지 않으므로 더 조용하다)

`resolveLoggingDecision` 역방향 검사가 프로덕션 주석 **3곳**을 찾아냈고, 델타 감사가 그 검사의
사정거리 밖에서 **1곳을 더** 찾아냈다(합 4곳). 전부 **현재의 폐기 동작을 사실로 서술**한다.

| 위치 | 서술 | M1 이후 |
|---|---|---|
| `internal/hook/instructions_loaded.go:48` (`appendRuleLoadAudit` 직전) | "The slog record above never reaches a reader … routes every `moai hook` invocation to io.Discard, unconditionally" | 거짓 — warn 이상은 싱크에 도달한다 |
| `internal/hook/config_change.go:192` (`runReload` 기록 근거) | "routes EVERY record emitted under `moai hook` to io.Discard, unconditionally … MOAI_LOG_LEVEL does not re-open the carve-out" | 앞 절은 거짓, `MOAI_LOG_LEVEL` 절은 **참으로 남는다**(REQ-HDS-002) |
| `internal/hook/factory_messages.go:144` (열화 반환 직전) | "a hook process discards slog records too …, so a degraded inspection is still not reported anywhere. **Closing that is card t1144.**" | 거짓 — 이 SPEC 이 바로 그 카드다 |
| `internal/hook/pre_tool.go:442` (`gateNotice` 선언 직전) — **델타 감사 추가 적발** | "the `moai hook` path installs a discarding handler, so a log record here would be silent by construction" | 거짓 — 그리고 이 주석은 폐기를 **설계 근거**로 삼는다(`gateNotice` 를 slog 대신 구조화 출력에 태우는 이유) |

**이 네 번째가 모집단 설계를 바꿨다.** `pre_tool.go:442` 는 `io.Discard` 도 `card t1144` 도
쓰지 않아 v0.3.1 의 두 토큰 명령 어디에도 잡히지 않았다 — 토큰으로 모집단을 잡으면 같은
계열이 조용히 빠진다는 실증이다. AC-HDS-016 이 그래서 **과다 포착 그물 + 전수 판정 기록**
으로 바뀌었다(22행 / 13파일).

네 주석 모두 **자기 옆 코드가 존재하는 이유**를 설명한다(감사 행을 따로 쓰는 이유,
`gateNotice` 가 구조화 출력을 타는 이유). 따라서 지우는 것이 아니라 **새 사실로 갱신**한다 —
그 이유는 남고(싱크는 레벨 게이트 뒤에 있으며 구조화 출력·감사 행과 용도가 다르다),
"어디에도 도달하지 않는다"는 부분이 바뀐다. AC-HDS-016 이 판정한다.

> 이 저장소는 거짓 주석을 결함으로 다룬다 — `SPEC-HOOK-TRACE-FLUSH-001` 이 `fan_in=24`
> 거짓 수치를 별도 요구사항으로 고친 선례가 있다.

### (3) 충돌 없음이 확인된 심볼

- `defaultLogLevel` — `internal/cli/logging.go` 안에서만 쓰이고, 값을 단언하는 외부 테스트가
  없다. 이 SPEC 은 값을 바꾸지 않는다(REQ-HDS-004 는 그 값을 **재사용**한다).
- `PruneStats` — `internal/hook/prune_logs.go` 에만 나타난다. 전체 구조체 비교
  (`PruneStats{…}` 리터럴 대조 / `reflect.DeepEqual`)는 **0건**이므로 필드 추가는
  기존 테스트를 깨지 않는다(REQ-HDS-009).

## REQ ↔ AC 커버리지

| REQ | AC |
|---|---|
| REQ-HDS-001 (파일 싱크 기록) | AC-HDS-001, AC-HDS-002 |
| REQ-HDS-002 (stdout/stderr 불변) | AC-HDS-004, AC-HDS-009, AC-HDS-015 |
| REQ-HDS-003 (해소 실패 시 fail-open) | AC-HDS-005, AC-HDS-006 |
| REQ-HDS-004 (기본 warn) | AC-HDS-007 |
| REQ-HDS-005 (`MOAI_LOG_LEVEL` 은 레벨만) | AC-HDS-008, AC-HDS-009 |
| REQ-HDS-006 (동시 append 무훼손) | AC-HDS-003 |
| REQ-HDS-007 (무기록 시 파일 미생성) | AC-HDS-012 |
| REQ-HDS-008 (경계 없는 대기 금지) | AC-HDS-014 |
| REQ-HDS-009 (기존 정리 편입) | AC-HDS-010 |
| REQ-HDS-010 (명명 상수) | AC-HDS-011 |
| REQ-HDS-011 (행동 가드) | AC-HDS-001 |
| REQ-HDS-012 (반증 가능성) | AC-HDS-002 |
| REQ-HDS-013 (계약 가드) | AC-HDS-004 |
| REQ-HDS-014 (fail-open 가드) | AC-HDS-006 |
| REQ-HDS-015 (이식성) | AC-HDS-013 |
| (범위 규율 — 기존 자산 갱신) | AC-HDS-015, AC-HDS-016 |

## 인수 기준

### AC-HDS-001 — warn 레코드가 싱크 파일에 실제로 도달한다

**Given** `t.TempDir()` 를 프로젝트 루트로 삼은 훅 인자 조합이 주어지고,
**When** 훅 경로의 로깅 결정을 설치한 뒤 `slog.Warn` 을 한 건 방출하면,
**Then** `<root>/.moai/logs/hook-runtime.log` 가 존재하고 그 안에 방출한 메시지가 한 줄로
들어 있다.

```
go test ./internal/cli/... -list '^TestHookPathWarnRecordReachesSink$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookPathWarnRecordReachesSink$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0.

### AC-HDS-002 — 가드 1 은 반증 가능하다

**Given** 가드 1 이 GREEN 인 상태에서,
**When** `resolveLoggingDecision` 의 훅 분기를 `io.Discard` 로 되돌리면,
**Then** 가드 1 의 **2단**이 FAIL 한다(1단은 1 을 유지한 채로 — 테스트는 그대로 있고 행동만
어긋나므로, 이것이 반증이 실제로 행동을 겨눴다는 증거다).

```
# 1) 훅 분기의 목적지를 io.Discard 로 되돌린다 (커밋하지 않는다)
go test ./internal/cli/... -list '^TestHookPathWarnRecordReachesSink$' | grep -cE '^Test'   # 1 유지
go test ./internal/cli/... -run  '^TestHookPathWarnRecordReachesSink$' -count=1              # FAIL 이어야 한다
# 2) 되돌린 편집을 폐기한다
git checkout -- internal/cli/logging.go
go test ./internal/cli/... -run  '^TestHookPathWarnRecordReachesSink$' -count=1              # 다시 PASS
```

판정: 1단계 2단이 non-zero, 2단계가 exit 0. run-phase 는 두 출력을 progress.md §E.2 에
인용한다.

### AC-HDS-003 — 동시 append 가 서로의 레코드를 훼손하지 않는다

**Given** 같은 싱크 파일을 가리키는 writer 가 N개 동시에 열려 있고,
**When** 각 writer 가 식별 가능한 레코드를 M건씩 방출하면,
**Then** 파일의 총 행 수가 N×M 이고, 각 행이 온전한 한 레코드다(행 중간에 다른 레코드가
끼어든 행이 0건).

```
go test ./internal/cli/... -list '^TestHookSinkConcurrentAppendIsIntact$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookSinkConcurrentAppendIsIntact$' -race -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 `-race` exit 0.
**행 순서는 판정하지 않는다**(`plan.md` §B-3).

### AC-HDS-004 — 훅 경로의 목적지가 두 표준 스트림 어느 쪽도 아니다

**Given** 훅을 가리키는 인자 조합(플래그 선행형 포함)이 주어지고,
**When** `resolveLoggingDecision` 을 해소하면,
**Then** 반환된 `dest` 가 `os.Stdout` 도 `os.Stderr` 도 아니다(포인터 동일성),
**And** 훅 경로 목적지를 구성하는 **파일 집합 전체**에 대해 판정 원칙 8 의 파일별 금지
토큰이 부재한다 — `logging.go` 는 `os.Stdout`, **싱크 writer 선언 파일은 `os.Stdout` 과
`os.Stderr` 둘 다**,
**And** 그 집합이 싱크 writer 선언 파일을 **실제로 포함**하며, 열거된 파일 중 읽을 수 없는
것이 있으면 건너뛰지 않고 **실패**한다.

```
go test ./internal/cli/... -list '^TestHookPathHookDestIsNeitherStdStream$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookPathHookDestIsNeitherStdStream$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0.

> **왜 세 번째 And 가 필요한가 (조용한 이탈 방지).** 집합을 손으로 열거하면 파일이 늘어날
> 때 새 파일이 집합 밖에 남아도 가드는 초록으로 통과한다 — 바로 이 SPEC 이 만든 상황이다
> (기존 가드가 `logging.go` 하나만 읽으므로 새 싱크 파일이 조용히 빠진다). 멤버십 단언이
> 그 재발을 막는 축이고, "읽을 수 없으면 실패"가 오타·이동으로 집합이 비는 경로를 막는다.
> 멤버십은 AC-HDS-011 이 이미 한 곳으로 고정한 싱크 경로 상수의 **선언 파일**로 건다 —
> 그래야 파일명이 바뀌어도 집합이 따라간다.

> **판정 방식의 근거는 판정 원칙 8.** 기존 `TestLoggingNeverTargetsStdout` 은 **델타 판정
> 대상이 아니다**(판정 원칙 7) — 이미 통과 중이며 계속 통과해야 한다. 그 가드는 좁은
> 사전 존재 단언으로 남고, **넓힌 집합은 가드 2 가 진다**(판독이 `logging.go`/`os.Stdout`
> 에서 겹치지만, 겹침은 비용이 아니라 이중 확인이다). AC-HDS-004 가 요구하는 것은 훅
> 목적지에 대한 **새 단언**이다.

### AC-HDS-005 — 루트를 해소할 수 없으면 폐기로 강등한다

**Given** `CLAUDE_PROJECT_DIR` 이 존재하지 않는 경로를 가리키고 cwd 도 쓰기 불가한 조건에서,
**When** 훅 경로의 로깅 결정을 해소하면,
**Then** 결정은 오류 없이 반환되고 이후 레코드 방출이 panic 도 non-zero 도 만들지 않는다.

```
go test ./internal/cli/... -list '^TestHookPathSinkFailureIsFailOpen$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookPathSinkFailureIsFailOpen$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0.

### AC-HDS-006 — 싱크 개방 실패가 훅을 실패시키지 않는다

**Given** 싱크 경로가 디렉터리로 선점되어 있는 등 `OpenFile` 이 실패하는 조건에서,
**When** 훅이 레코드를 방출하고 정상 경로를 끝까지 실행하면,
**Then** 훅은 exit 0 으로 끝나고 목적지는 폐기로 강등되어 있다.

```
go test ./internal/cli/... -list '^TestHookPathSinkFailureIsFailOpen$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookPathSinkFailureIsFailOpen$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0.

### AC-HDS-007 — 기본 구성에서 info / debug 는 기록되지 않는다

**Given** `MOAI_LOG_LEVEL` 이 설정되지 않은 훅 경로에서,
**When** debug 1건, info 1건, warn 1건을 방출하면,
**Then** 싱크 파일에는 warn 레코드만 있고 debug / info 레코드는 0건이다.

```
go test ./internal/cli/... -list '^TestHookSinkLevelBoundary$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookSinkLevelBoundary$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0.

### AC-HDS-008 — `MOAI_LOG_LEVEL` 이 낮아지면 info 도 기록된다

**Given** `MOAI_LOG_LEVEL=INFO` 인 훅 경로에서,
**When** info 레코드를 방출하면,
**Then** 싱크 파일에 그 레코드가 존재한다.

```
go test ./internal/cli/... -list '^TestHookSinkLevelBoundary$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookSinkLevelBoundary$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0. 환경변수는
`t.Setenv` 로 주입하고 **병렬 서브테스트에서 쓰지 않는다**(CLAUDE.local.md §6; 기존
`TestLoggingHandlerSelection` 도 같은 이유로 비병렬이다).

> AC-HDS-007 과 008 은 같은 가드(6)의 서로 다른 축을 판정한다 — 기본값에서의 침묵과
> 하향 시의 기록. 한 가드가 두 AC 를 지므로, run-phase 가 축을 분리하면 각 절의 이름을
> 갱신한다(§ Gaps).

### AC-HDS-009 — `MOAI_LOG_LEVEL` 은 목적지를 바꾸지 못한다

**Given** `MOAI_LOG_LEVEL` 을 DEBUG / INFO / WARN / ERROR 로 각각 설정한 훅 경로에서,
**When** 로깅 결정을 해소하면,
**Then** 어떤 값에서도 `dest` 가 `os.Stdout` 도 `os.Stderr` 도 아니다.

```
go test ./internal/cli/... -list '^TestHookPathHookDestIsNeitherStdStream$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookPathHookDestIsNeitherStdStream$' -count=1
```

판정: 1단 = 1 AND 2단 exit 0. 가드 2 가 값 테이블을 순회한다.

### AC-HDS-010 — 싱크 파일이 SessionEnd 정리 후보에 편입된다

**Given** 보존 임계값보다 오래된 `hook-runtime.log` 가 있는 로그 디렉터리에서,
**When** `PruneObservationLogs` 를 실행하면,
**Then** 해당 파일이 제거되고 그 사실이 `PruneStats` 의 필드로 보고되며, 임계값 이내의
파일은 보존된다.

```
go test ./internal/hook/... -list '^TestHookSinkIsPrunedAtSessionEnd$' | grep -cE '^Test'   # = 1
go test ./internal/hook/... -run  '^TestHookSinkIsPrunedAtSessionEnd$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0.

### AC-HDS-011 — 경로와 보존값이 명명 상수다

**Given** 구현이 완료된 트리에서,
**When** 싱크 경로 문자열의 비테스트 출현을 검색하면,
**Then** **상수 선언 1곳뿐**이다.

```
grep -rn '"\.moai/logs/hook-runtime\.log"' --include='*.go' internal/ | grep -v _test.go
```

기준선(HEAD `9b973bc04`): **exit 1, 무매치**. 판정: 출력 1행이고 그 행이 `const` 선언이다.
2행 이상이면 FAIL. (판정 원칙 3 의 두 단 면제 — 테스트 이름을 선택자로 쓰지 않는다.)

### AC-HDS-012 — 레코드가 없으면 파일을 만들지 않는다 (지연 개방)

**Given** 비어 있는 프로젝트 루트에서,
**When** 훅 경로의 로깅 결정을 설치하고 **아무 레코드도 방출하지 않은 채** 종료하면,
**Then** `<root>/.moai/logs/hook-runtime.log` 가 존재하지 않고, `.moai/logs/` 디렉터리도
이 호출 때문에 생성되지 않는다.

```
go test ./internal/cli/... -list '^TestHookSinkIsNotCreatedWithoutRecords$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestHookSinkIsNotCreatedWithoutRecords$' -count=1
```

기준선(HEAD `9b973bc04`): 1단 **0** → 미충족. 판정: 1단 = 1 AND 2단 exit 0.

> **전용 가드가 필요한 이유**: v0.1.x 는 이 판정을 가드 3(개방 **실패** 경로)에 얹었는데,
> 그 경로는 개방을 시도한 뒤의 이야기라 **지연 개방을 보지 않아도 통과**한다. 지연 개방은
> `plan.md` §B-1 이 그 설계를 고른 유일한 근거(REQ-HDS-007)이므로 자기 가드를 갖는다.

### AC-HDS-013 — 세 대상에 대해 빌드된다

**Given** 구현이 완료된 트리에서,
**When** 크로스 빌드를 수행하면,
**Then** 세 대상 모두 exit 0 이다.

```
GOOS=linux   GOARCH=amd64 go build ./... && \
GOOS=darwin  GOARCH=arm64 go build ./... && \
GOOS=windows GOARCH=amd64 go build ./...
```

판정: 마지막 exit 0. (판정 원칙 3 의 두 단 면제.)

### AC-HDS-014 — 훅 경로에 경계 없는 대기가 없다

**Given** 구현이 완료된 트리에서,
**When** 변경된 패키지의 테스트를 타임아웃과 함께 실행하면,
**Then** 타임아웃에 걸리지 않고 끝난다.

```
go test ./internal/cli/... ./internal/hook/... -count=1 -timeout 120s
```

판정: exit 0. 타임아웃 종료(exit 1 + `panic: test timed out`)는 FAIL 이며, 싱크 기록
경로에 도입된 대기를 먼저 의심한다. (판정 원칙 3 의 두 단 면제 — 특정 테스트 이름이 아니라
**패키지 전체**를 겨누므로 스위트가 비어 있을 수 없다.)

### AC-HDS-015 — 기존 훅 케이스 3건이 새 계약으로 갱신되고 통과한다

**Given** `internal/cli/logging_test.go` `TestLoggingHandlerSelection` 의 훅 케이스 3건
(`hook_discards`, `hook_behind_a_flag_discards`, `hook_ignores_log_level_env`)이 현재
`io.Discard` 를 기대하고 있고,
**When** M1 이 훅 목적지를 파일 싱크로 바꾸면,
**Then** 세 케이스의 기대값이 새 목적지를 반영하도록 갱신되어 있고 전체 테스트가 통과하며,
`hook_ignores_log_level_env` 는 **"환경변수가 목적지를 바꾸지 못한다"는 단언을 유지**한다.

```
go test ./internal/cli/... -list '^TestLoggingHandlerSelection$' | grep -cE '^Test'   # = 1
go test ./internal/cli/... -run  '^TestLoggingHandlerSelection$' -count=1
grep -c 'hook_discards\|hook_behind_a_flag_discards\|hook_ignores_log_level_env' internal/cli/logging_test.go
```

기준선(HEAD `9b973bc04`): 1단 **1**(이미 존재), 2단 **exit 0**(현재 계약으로 통과 중),
grep **3**. 판정: 1단 = 1 · 2단 exit 0 · grep 여전히 3.

> **사전 존재 테스트에 대한 델타 판정**(판정 원칙 7). 미수정 트리에서 이미 2단이 통과하므로
> 그것만으로는 변별력이 없고, **grep 3 유지**가 "케이스를 삭제해서 통과시키는" 경로를 막는
> 축이다. 기대값이 실제로 싱크를 가리키는지는 run-phase 감사 몫이다.

### AC-HDS-016 — 폐기 동작을 서술하는 주석이 전수 판독되고, 갱신 대상이 갱신된다

**Given** 기준선 SHA `bbc855f45` 에서 아래 **그물**이 `internal/hook` 비테스트 Go 파일의
`discard` 언급 **22행(13개 파일)** 을 반환하고,
**When** M1 이 훅 경로의 폐기 동작을 바꾸면,
**Then** 그 22행 **전수**에 대한 판정 기록이 존재하고(행당 정확히 한 건), 각 행이
`갱신 대상` / `무관` / `판정 유보` 중 하나로 분류되어 있으며, `갱신 대상` 으로 분류된 행은
실제로 갱신되어 있다.

```
git grep -niE 'discard' bbc855f45 -- internal/hook | grep '\.go:' | grep -v _test.go
```

기준선(HEAD `bbc855f45`, 실측): **22행 / 13파일**. 파일별 분포 —
`branch_guard.go` 4 · `session_start.go` 3 · `session_start_record.go` 2 · `registry.go` 2 ·
`pre_tool.go` 2 · `handoff_inject.go` 2 · `user_prompt_submit.go` 1 ·
`session_start_kanban.go` 1 · `session_start_drift_fill.go` 1 · `quality/gate.go` 1 ·
`instructions_loaded.go` 1 · `factory_messages.go` 1 · `config_change.go` 1.

**판정 기록의 위치와 형식** (이것이 고정되지 않으면 "행 수 일치"를 셀 대상이 없다):
`progress.md` §E.2 안의 표 하나, 행당 `<file>:<line> | 갱신 대상|무관|판정 유보 | <근거>`.
`<file>:<line>` 은 위 그물 명령이 출력한 값을 그대로 쓴다.

```
# 기계적 축 1 — 전수 완결성
git grep -niE 'discard' bbc855f45 -- internal/hook | grep '\.go:' | grep -v _test.go | wc -l   # = 22
grep -cE '^\| internal/hook/.*\.go:[0-9]+ \|' .moai/specs/SPEC-HOOK-DIAG-SINK-001/progress.md  # = 22

# 기계적 축 2 — 지정 4행의 분류
grep -E '^\| internal/hook/(config_change\.go:192|instructions_loaded\.go:48|factory_messages\.go:144|pre_tool\.go:442) \| 갱신 대상 \|' \
  .moai/specs/SPEC-HOOK-DIAG-SINK-001/progress.md | wc -l                                       # = 4
grep -cE '^\| internal/hook/session_start\.go:420 \| 판정 유보 \|' \
  .moai/specs/SPEC-HOOK-DIAG-SINK-001/progress.md                                               # = 1
```

**판정**: 위 네 수치가 각각 22 / 22 / 4 / 1 (기계적 축) **AND** `갱신 대상` 으로 분류된 행이
실제로 갱신됨 (run-phase 감사의 사람 판독).

| 그물 행 | 요구 분류 | 근거 |
|---|---|---|
| `internal/hook/config_change.go:192` | 갱신 대상 | `io.Discard` 로 "EVERY record" 라우팅을 사실로 서술. 단 `MOAI_LOG_LEVEL` 절은 **참으로 남는다**(REQ-HDS-002) |
| `internal/hook/instructions_loaded.go:48` | 갱신 대상 | "routes every `moai hook` invocation to io.Discard, unconditionally" |
| `internal/hook/factory_messages.go:144` | 갱신 대상 | "a hook process discards slog records too … still not reported anywhere. **Closing that is card t1144**" — 이 SPEC 이 그 카드이므로 미완료 서술로 남길 수 없다 |
| `internal/hook/pre_tool.go:442` | 갱신 대상 | "the `moai hook` path installs a discarding handler, so a log record here would be silent by construction" — 폐기를 사실로 놓고 그 위에 **설계 판단**(`gateNotice` 를 slog 대신 구조화 출력에 태움)을 얹었다 |
| `internal/hook/session_start.go:420` | **판정 유보** | 아래 유보 사유 참조 |

**[HARD] `session_start.go:420` 은 판정을 유보한다.** "in slog, whose stderr the hook wrappers
discard. The failure was recorded and invisible" — **과거형 서술**이고, 폐기 주체를
`resolveLoggingDecision` 이 아니라 **셸 래퍼의 stderr 처리**로 돌린다. 이 SPEC 은 stderr 를
열지 않으므로(REQ-HDS-002) M1 이후 이 문장이 거짓이 되는지 **단정할 수 없다.** 사람 판독
대상으로만 올리고, 갱신 여부는 run-phase 감사가 정한다 — 이 불확실성 자체가 기록이다.

> **왜 토큰 일치가 아니라 과다 포착 그물인가 (v0.3.2 개정 사유).** v0.3.1 은 모집단을
> `io.Discard` + `card t1144` 두 토큰으로 잡았는데, **두 겹으로 틀렸다.**
>
> **(a) 기준선이 재현되지 않았다.** "첫 명령 3행(세 파일)"이라고 적었으나 실측은 **2행**
> (`config_change.go:192` · `instructions_loaded.go:48`)이고, `factory_messages.go` 는 그
> 집합에 **없다** — 그 주석은 `io.Discard` 토큰을 쓰지 않고 산문으로만 서술한다.
> 재지 않고 적은 값이었다.
>
> **(b) 모집단이 닫혀 있지 않았다.** 토큰으로 모집단을 잡는데 대상 하나가 그 토큰을 안
> 쓴다면 같은 계열의 다른 주석도 사정거리 밖이다. 실제로 `pre_tool.go:442` 가 두 명령
> 어디에도 안 잡혔다 — AC-016 이 존재하는 이유에 정확히 해당하는 주석인데도.
> **D1 과 같은 계열의 반대 방향이다**(D1 = 과다 포착, 이것 = 과소 포착).
>
> 그래서 **정확한 그물을 포기하고 과다 포착 + 기록된 판독**으로 바꿨다. 그물은 무관한
> 행을 함께 잡고, **그중에는 주석이 아닌 행도 있다** — 예: `pre_tool.go:623` "the return is
> intentionally discarded"(자문용 반환값 폐기 주석, 이 SPEC 과 무관), `branch_guard.go` 4행,
> 그리고 `handoff_inject.go:92` 의 **사용자 표시 문자열**. 그것이 **설계 의도**다:
> 무엇이 갱신 대상인지는 여전히 사람이 읽지만, **누락이 조용히 일어나지 않는다.**

> **[HARD] 그물은 기준선 SHA 에 고정한다.** M1 이 주석을 고치면 그물 자체가 움직인다 —
> `갱신 대상` 행은 갱신 후 `discard` 를 더는 안 쓸 수 있어 현재 트리 그물에서 빠지고,
> 그러면 행 수 일치 판정이 공허해진다. `git grep <SHA>` 형태가 그 이동을 막는다
> (실측: 워킹트리 `grep -rn` 과 `git grep bbc855f45` 둘 다 22행 — 같은 모집단이다).

> **감사 측정과의 차이 (21 vs 22) — 원인은 주석 한정 필터다.** 감사가 보고한 21 은
> **주석으로 한정한 수**였다(`… | grep -E '//'`). 그 필터가 떨어뜨린 1행은
> `internal/hook/handoff_inject.go:92` 의 **사용자 표시 문자열**이다 —
> `"or run \`moai handoff clear\` to discard it.\n"`. 그물(22)은 이 행을 포함한다.
>
> **[HARD] 이 절의 초판 귀속(“원인은 `quality/gate.go` 1행”)은 틀렸다** — 감사의 명령도
> `internal/hook/` 을 재귀로 훑으므로 `quality/` 는 애초에 포함돼 있었다. 그 오답이
> 확증처럼 보인 이유는 **두 설명이 각각 다른 1행을 빼고 똑같이 21 을 내기 때문**이다:
> `quality/gate.go:903` 도 1행(주석), `handoff_inject.go:92` 도 1행(문자열). 수치가
> 재현된다는 것은 **그 설명이 옳다는 증거가 아니다** — 같은 값에 닿는 뺄셈이 둘 이상 있을
> 때, 실제로 어느 쪽이 일어났는지는 상대의 명령을 보지 않고는 정해지지 않는다. 초판은
> 그것을 보지 않고 "원인 규명됨"이라고 적었다.
>
> **`quality/gate.go` 포함 결정은 그대로 유지한다** — 이 귀속과 독립인 판단이다.
> `internal/hook/quality` 는 도달성 측정에서 OTHER-ONLY(훅 경로에 오지 않음, `spec.md` §1.1)
> 이므로 그 행의 판정은 `무관` 이 될 공산이 크지만, 제외 조항을 두는 쪽이 **조항을 잘못
> 쓸 위험**을 새로 만든다. 판정 기록 한 줄이 더 드는 비용이 더 싸다.

> 주석은 테스트를 깨뜨리지 않으므로 **AC-HDS-015 보다 조용히 틀린다**. 이 저장소가 거짓
> 주석을 결함으로 다루는 선례는 `SPEC-HOOK-TRACE-FLUSH-001`(`fan_in=24` 거짓 수치 정정).

## 품질 게이트

```
go vet ./internal/cli/... ./internal/hook/... ./internal/config/...
golangci-lint run internal/cli/... internal/hook/... internal/config/...
go test ./internal/cli/... ./internal/hook/... ./internal/config/... -count=1
```

판정: 세 명령 모두 exit 0. **전체 스위트(`go test ./...`)는 로컬에서 돌리지 않는다** —
전 패키지 판정은 CI 몫이다(CLAUDE.local.md §4).

> 세 번째 명령의 exit 0 은 AC-HDS-015 가 이행된 **뒤에만** 도달 가능하다 — 그 전에는
> `TestLoggingHandlerSelection` 의 훅 케이스 3건이 빨갛다. 이 의존을 명시해 두는 이유는,
> v0.1.x 가 그 사실을 적지 않아 품질 게이트가 기술된 대로는 도달 불가였기 때문이다.

## Gaps (이 인수 기준이 지키지 않는 것)

- **`-list` 존재 확인은 이름을 지킬 뿐 내용을 지키지 않는다.** 본문이 빈 테스트도 1단과
  2단을 모두 통과한다. 두 단 판정식은 "가드가 사라지거나 개명되는" 실패만 막으며, **그
  테스트가 무엇을 단언하는지의 적합성은 run-phase 감사 몫이다.** AC-HDS-002 의 반증 절차가
  이 Gap 을 부분적으로만 메운다 — 가드 1 하나에 대해서만 행동을 겨눴음을 보인다.
- **[HARD] 그물은 "주석 집합"이 아니다 — 문자열 리터럴도 포착한다.** 22행 중
  `internal/hook/handoff_inject.go:92` 는 주석이 아니라 **사용자 표시 문자열**이다
  (`"or run \`moai handoff clear\` to discard it.\n"`). 판정 기록에서 `무관` 으로 처리될
  것이므로 설계는 그대로 유효하지만, 모집단의 성격을 "주석 22곳"으로 읽으면 그물을 다시
  좁히려는 유혹이 생긴다 — 좁히는 순간 AC-016 초판의 과소 포착이 재현된다.
- **[HARD] 이 SPEC 에서 모집단 경계는 네 번 문제가 됐다.** run-phase 는 같은 자리를 먼저
  본다: **과다 포착**(D1) → **과소 포착**(AC-016 초판의 토큰 모집단) → **형식 불일치**
  (지정행의 짧은 이름 대 그물 출력의 전체 경로) → **귀속 오류**(21 vs 22 의 원인을 재지
  않고 추정). 네 번 모두 "모집단을 무엇으로 정의했는가"에서 갈렸고, 마지막 것은 **수치가
  재현된다는 사실을 설명이 옳다는 증거로 읽은** 오류다.
- **AC-HDS-016 의 그물은 의도적으로 과다 포착하므로, 무엇이 갱신 대상인지는 사람이 읽는다.**
  기계적 축은 **전수 판독의 완결성**(판정 기록 행 수 = 그물 22행)과 **지정 4행의 분류**뿐이고,
  `갱신 대상` 으로 분류된 주석이 *새 사실을 정확히* 서술하는지는 감사 판단이다. 이 AC 가
  막는 것은 "틀린 갱신"이 아니라 **"조용한 누락"** 이다.
- **판정 기록의 내용 정확성은 기계로 보지 않는다.** 22행 전부를 `무관` 으로 적어도 행 수
  일치는 통과한다 — 지정 4행의 분류 고정이 그 최악의 경우를 부분적으로만 막는다(4/22).
  나머지 18행의 분류 타당성은 run-phase 감사 몫이다.
- **빌드 실패 시 `-list` 의 행동은 미관측이다.** 컴파일이 깨진 트리에서 1단이 어떤 값을
  내는지 재지 않았다(`ac-baseline.md` Gaps 와 동일). 품질 게이트가 그 상태를 먼저 잡을
  것으로 보지만, 그것은 추론이지 측정이 아니다.
- **[HARD] 정적 축은 토큰의 부재만 보므로, 결정값을 받은 뒤 두 스트림으로 tee 하는 writer 는
  현행 판정 전부를 통과한다.** AC-HDS-004 의 세 축은 (a) `dest` 포인터가 두 스트림 어느
  쪽도 아님, (b) 파일 집합에 `os.Stdout` / `os.Stderr` 토큰 부재, (c) 집합 멤버십이다.
  싱크 writer 가 그 토큰을 쓰지 않고 — 예컨대 다른 경로로 얻은 파일 디스크립터나 주입된
  `io.Writer` 를 통해 — 스트림으로도 함께 내보내면 셋 다 초록이다.
  **이것은 판정이 REQ-HDS-002 의 효과("어떤 레코드도 내보내서는 안 된다")에서 한 층 아래
  (결정값 + 소스 토큰)로 내려온 데서 오는 잔여 위험이다.** 요구사항의 구속력은 그대로이고
  판정만 좁다 — 효과 층은 **run-phase 의 실 바이너리 관측 몫**이며, 그 관측 없이
  AC-HDS-004 PASS 를 "stdout/stderr 로 아무것도 나가지 않음"의 증거로 읽어서는 안 된다.
  (판정 원칙 8 이 런타임 대조를 이 저장소에서 판정 불가로 분류한 대가이기도 하다.)
- **AC-HDS-007 과 008 이 가드 6 하나를 공유한다.** 접힘은 v0.1.x 보다 크게 줄었으나 완전히
  사라지지는 않았다 — 두 축이 한 이름 뒤에 있으므로, 한 축만 작성해도 두 AC 가 함께
  통과한다. 축 분리는 run-phase 재량이며, 분리하면 각 절의 이름을 갱신한다.

## Definition of Done

- [ ] AC-HDS-001 … AC-HDS-016 전원 PASS — **테스트 기반 AC 는 1단 값(= 1)과 2단 출력을 모두** 인용
- [ ] 각 판정 명령의 출력이 progress.md §E.2 에 인용되고, 기준선에는 HEAD `9b973bc04` 가 동반됨
- [ ] AC-HDS-002 의 반증 절차가 실제로 수행되고 FAIL→PASS 두 출력이 인용됨
- [ ] AC-HDS-015 의 grep 이 3 을 유지함 (케이스 삭제로 통과시키지 않았음)
- [ ] AC-HDS-016 의 판정 기록이 progress.md §E.2 에 **22행 전수**로 존재하고(행당 한 건,
      `<file>:<line> | 갱신 대상|무관|판정 유보 | 근거` 형식), 지정 4행이 요구 분류를 가지며,
      `session_start.go:420` 이 `판정 유보` 로 기록됨
- [ ] 품질 게이트 3종 exit 0
- [ ] `spec.md` §4 의 범위 제외가 전부 지켜짐 — 특히 stdout/stderr 미개방, 템플릿 트리 미수정
- [ ] `internal/template/templates/` 하위 diff 0
- [ ] 모든 인용 수치에 base SHA 가 동반됨
