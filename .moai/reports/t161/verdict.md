# t161 — 토큰 원장 경로가 머신에 좌우되던 문제 (Class B)

- 카드: t161 (Class B). PR #1585 차단 요인 1/2
- 워크트리: `.claude/worktrees/t161`
- 브랜치: `WT-tokens-ledger-path` (base `76ef8a764`)
- push: 하지 않음 (카드 지시)

---

## 1. 원인 — 확정

`resolveTokensStateDir()` (`internal/cli/tokens.go:376`) 는 `findStateDir()` 의
**무한 walk-up** 에 의존한다. 이 walk 은 cwd 에서 파일시스템 루트까지 올라가며 처음 만나는
`.moai/state` 를 반환한다(`internal/cli/state.go`). 즉 **cwd 의 조상 어딘가에 `.moai/state`
가 있으면 원장이 그리로 간다.**

테스트는 `t.Chdir(tmp)` 만 해두고 원장이 `tmp/.moai/state` 에 생기기를 기대했다
(`tokens_test.go:190`). 조상이 가로채면 `tmp/.moai/state` 는 아예 만들어지지 않으므로,
읽기가 `ERROR_PATH_NOT_FOUND`("The system cannot find the **path** specified" — 파일이
아니라 **경로** 부재)로 실패한다. CI 로그의 에러 문구가 정확히 이 형태다.

**메커니즘 실측 확인** (이 워크트리, macOS):

```
cwd  = <root>/AppData/Local/Temp/TestX/001
got  = <root>/.moai/state                      ← 조상이 가로챔
leaf-local would be = <root>/AppData/Local/Temp/TestX/001/.moai/state
RESULT: ancestor hijack CONFIRMED
```

카드가 준 가설 3개 중 첫 번째("테스트가 cwd 를 tmp 로 안 옮김")는 **오답**이다 —
`tokens_test.go:171` 에 `t.Chdir(tmp)` 가 분명히 있다. 세 번째("findStateDir walk-up 이
Windows 에서 다른 지점 착지")가 맞고, 두 번째(8.3 단축경로)는 **원인이 아니다**: `RUNNER~1`
과 긴 이름은 NTFS 에서 같은 디렉터리로 해석되므로 그것만으로는 경로 부재가 나올 수 없다.

## 2. Linux 통과 / Windows 실패의 비대칭 — [HARD] 요구 해명

가로채기가 성립하려면 **두 조건**이 동시에 필요하다: (1) cwd 가 `.moai/state` 를 가진
디렉터리의 **하위**일 것, (2) 그런 `.moai/state` 가 실제로 존재할 것.

조건 (1) 이 OS 마다 갈린다. Go 의 `t.TempDir()` 은 `os.TempDir()` 아래에 만들어지는데:

| 러너 | temp 루트 | 홈 하위인가 | walk-up 이 홈을 지나는가 |
|---|---|---|---|
| windows-latest | `C:\Users\RUNNER~1\AppData\Local\Temp` | **예** | **지난다** |
| ubuntu-latest | `/tmp` (TMPDIR 미설정) | 아니오 | 안 지난다 |
| macos-latest | `/var/folders/...` | 아니오 | 안 지난다 |

Windows 의 temp 가 사용자 프로필 **안**이라는 것은 추정이 아니라 **CI 에러 메시지 자체가
증거**다 — 실패 경로가 `C:\Users\RUNNER~1\...\001\...` 이다. 그리고
`release-pr-multi-os.yml` 에는 `env:` 블록도 `TMPDIR` 설정도 없다(실측: grep 결과 없음).
따라서 리눅스 레그의 cwd 는 `/tmp/...` 이고, `/tmp` 와 `/` 어디에도 `.moai/state` 는 없으며
`/home/runner` 는 애초에 walk 경로에 오르지 않는다.

**요약**: 같은 코드가 리눅스에서 성립하는 이유는 "리눅스에는 `~/.moai/state` 가 없어서"가
아니라 **temp 가 홈 밖이라 walk-up 이 홈을 아예 지나지 않아서**다. Windows 만 홈을 지난다.
조건 (2) 를 만족시키는 `~/.moai/state` 는 제품이 정상 동작 중 만드는 디렉터리다
(실측: 이 개발 머신에 `/Users/goos/.moai/state` 존재; 생성 지점 후보는
`internal/statusline/{context_usage,model_cache}.go`, `internal/session/store.go` 등 다수).

## 3. 이건 회귀가 아니라 **새로 들어온 취약 테스트**다 — 카드 귀속 정정

카드는 `d73e9a669`(t118 네이밍 스윕)를 지목했다. 그 커밋이 `tokens_test.go` 를 마지막으로
건드린 것은 맞지만, 실제 변경은 **라벨 문자열과 gofmt 정렬뿐**이다(실측):

```
-		Role:     "leader",       →  +		Role:     "lead",
-	"role", "", "... lead, run, review, sync)"  →  "... lead, plan, run, sync)"
-	Model string             →  +	Model string          (정렬)
```

원인 코드와 테스트는 `dd060a191`("feat(cli): add moai tokens record", 2026-08-17) 에서
**신규 도입**됐고, 이 커밋이 이번 배치 소속이다. 즉 "PR #1582 는 39 pass" 와 모순되지
않는다 — 그때는 이 테스트가 존재하지 않았다. **기존 동작의 회귀가 아니라, Windows CI 를
한 번도 거치지 않은 채 들어온 환경 의존 테스트**다.

## 4. 조치

주입 seam 은 **이미 있었다** — `tokens_recordOpts.StateDir` (`tokens.go:255`, 소비 지점 :318).
테스트가 쓰지 않았을 뿐이다. 그래서 코드 동작은 건드리지 않고 세 가지를 했다.

**(a) 테스트가 stateDir 을 명시 주입하도록 변경** ([HARD] 요구사항)

`tokens_test.go` — 원장 위치가 **검증 대상이 아닌** 테스트 3곳에서 `StateDir` 을 주입해
walk-up / cwd 조회를 경로에서 완전히 제거했다:

- `TestRunTokensRecordAppendsLedger` — 실패하던 그 테스트
- `TestRunTokensRecordJSONFlagNoLedger` — "원장이 안 생겨야 한다" 를 tmp 기준으로 확인하던 곳
- `TestRunTokensRecordContextSnapshot` — 두 번째 절반(스냅샷 부재 시 `Context == nil`)이
  같은 이유로 Windows 에서 오탐 가능했다. 조상의 `context-usage.json` 을 주워 오면
  `Context != nil` 이 되어 실패한다. 보고된 실패 목록에는 없었지만 같은 클래스라 함께 고쳤다

읽기 쪽 리터럴 `"token-accounting.jsonl"` 도 상수 `tokensLedgerFilename` 으로 바꿨다(중복 제거).

**(b) 경로 결정 로직을 순수 함수로 분리** ([HARD] 요구사항)

`state.go` — `findStateDir()` 을 두 조각으로 나눴다. `findStateDirFrom(start string)` 이
walk 전체를 담고, `findStateDir()` 은 `os.Getwd()` 를 읽어 그 함수를 호출한다.
**동작 변화 0** — 프로세스에서 시작점을 읽는 부분만 밖으로 나왔다. 이제 walk 을 테스트가
소유한 트리 위에서 Windows 형태 입력으로 검사할 수 있다.

**(c) 회귀 테스트 추가** — `internal/cli/tokens_state_dir_test.go` (신규)

- `TestFindStateDirFromWalksUp` 3케이스: 조상 우선 / 시작점 우선 / 트리 안 무주장.
  레이아웃을 `AppData/Local/Temp/TestSomething/001` 로 잡아 Windows 형태를 그대로 재현한다
- `TestRunTokensRecordHonoursInjectedStateDir`: 조상에 미끼(decoy) `.moai/state` 를 두고
  cwd 를 그 하위로 옮긴 뒤 `StateDir` 을 주입 — 원장이 주입 위치에만 생기고 미끼에도
  cwd 에도 새지 않음을 확인. Windows 에서 벌어진 상황을 그대로 재현한 형태다

세 번째 케이스("트리 안 무주장")는 에러를 단정하지 않는다. walk 이 테스트 트리 밖으로
계속 올라가므로 "에러가 난다"는 곧 "러너에 `~/.moai/state` 가 없다"는 주장이 되고,
그건 이 테스트가 문서화하려는 환경 의존 그 자체이기 때문이다. 대신 "테스트 트리 안의
무엇도 반환되지 않는다"만 확인한다 — 이건 어느 머신에서나 결정적이다.

## 5. 검증 실측

| 항목 | 명령 | 결과 |
|---|---|---|
| 메커니즘 재현 | 임시 프로브(조상 hijack) | `RESULT: ancestor hijack CONFIRMED` |
| 신규 회귀 테스트 | `go test -count=1 -run 'TestFindStateDirFromWalksUp\|TestRunTokensRecordHonoursInjectedStateDir' ./internal/cli/` | 4/4 PASS |
| 기존 tokens 테스트 | `go test -count=1 -run TestRunTokensRecord ./internal/cli/` | 6/6 PASS |
| 영향 패키지 전체 | `go test -count=1 -timeout 20m ./internal/cli/` | **ok 223.986s** |
| 정적 분석 | `go vet ./internal/cli/` | 출력 없음 |
| 크로스 컴파일 | `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` | rc=0 |
| 포맷 | `gofmt -l <변경 3파일>` | 출력 없음 |

프로브 파일은 검증 후 삭제했다(커밋에 없음).

## 6. 변경 파일

| 파일 | 성격 |
|---|---|
| `internal/cli/state.go` | `findStateDirFrom` 추출 — 동작 변화 없음, 테스트 가능한 seam |
| `internal/cli/tokens_test.go` | 3개 테스트에 `StateDir` 주입 + 파일명 상수화 |
| `internal/cli/tokens_state_dir_test.go` | 신규 회귀 테스트 4케이스 |

프로덕션 동작은 바뀌지 않았다 — `resolveTokensStateDir()` 의 두 분기(walk-up → cwd 폴백)
그대로다.

---

## 미검증 (Gaps)

- **러너의 `~/.moai/state` 를 직접 관측하지 못했다.** 가로채기 *메커니즘* 은 확정했고
  Windows 만 홈을 지난다는 *비대칭* 도 확정했지만, windows-latest 러너에서 그 디렉터리를
  **무엇이** 만들었는지는 CI 안을 볼 수 없어 확인하지 못했다. 후보는 좁혔다(위 §2).
  **다만 수정은 그 답에 의존하지 않는다** — 주입으로 조건 (1) 자체를 없앴기 때문에
  어떤 프로세스가 만들었든 무관하다.
- **원래 실패를 이 머신에서 재현하지는 않았다.** 재현하려면 `$TMPDIR` 의 조상에
  `.moai/state` 를 심어야 하는데, 지금 이 머신에서 다른 레인들이 `go test` 를 돌리고 있어
  그들의 테스트까지 가로채게 된다. 검증이 부하·간섭을 만드는 전형이라 하지 않았다.
  대신 동일 구조를 테스트 소유 트리 안에서 재현했다(§4c).
- **전체 스위트 미실행** — 카드가 금지했다. `internal/cli` 만 돌렸다. 전 패키지·3-OS 판정은
  PR CI 몫이며, 이 카드의 진짜 판정도 windows-latest 레그다.
- **다른 walk-up 소비자 미점검** — `findStateDir()` 은 `clean.go`, `chain.go`, `state.go`
  에서도 쓰인다. 이들도 같은 환경 의존을 갖지만 이번 실패와 무관하고 범위 밖이라
  건드리지 않았다.

## 잔여 위험

- **프로덕션의 무한 walk-up 은 그대로다.** 사용자의 홈 아래에서 `moai tokens record` 를
  실행하면(예: `~/projects/foo` 에 `.moai/state` 가 없고 `~/.moai/state` 가 있으면) 원장이
  홈으로 간다. 의도된 동작인지 결함인지는 이 카드가 판단할 사안이 아니라 남겨둔다 —
  후속 카드 후보: walk-up 에 상한(리포지토리 루트 등)을 두거나 `--state-dir` 플래그를 노출.
- **같은 형태의 테스트가 더 있을 수 있다.** cwd 에 기대어 `.moai/state` 를 다루는 테스트는
  전부 Windows 에서 같은 함정을 밟는다. `internal/cli` 안에서 tokens 계열만 훑었고
  전수 조사는 하지 않았다.
