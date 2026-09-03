# t451 — `moai doctor` Codex Wiring 점검이 삼키던 침묵 두 건

카드: t451 · 브랜치: `WT-codex-wiring-doctor` · 기반: develop 팁 `e79c010b8`
워크트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t451`

---

## Claim (주장)

1. **미배선 프로젝트 침묵을 깼다.** `codex` 바이너리가 PATH에 잡히는데도 프로젝트에
   `.codex/hooks.json`과 `.codex/config.toml`이 둘 다 없으면, 지금까지는 `CheckOK` +
   "not wired (claude-only project) — skipped"로 조용히 넘어갔다. 이제 `CheckWarn`을
   내고, 결과(MCP 서버 미등록 + 생성된 훅이 못 뜬다)와 조치(`run moai init --agent codex`)를
   **Message에** 실어 보낸다. `Detail`은 `--verbose`에서만 렌더되므로 맨 `moai doctor`에
   보이려면 Message여야 한다.
2. **묵은 스킬 등록 침묵을 깼다.** 사용자 계층 `~/.codex/config.toml`의
   `[[skills.config]]` 항목 중 `path`가 실제로 없는 것들을 세어 보고한다. 전체 대비
   몇 개인지와 enabled/disabled 분포를 함께 낸다 — enabled인 채로 사라진 경로는 지금
   깨져 있는 등록이고, disabled는 남아 있는 찌꺼기라 심각도가 다르다.
3. **claude 전용 사용자는 여전히 조용하다.** `.codex/`도 없고 머신에 `codex`도 없으면
   이전과 **글자 그대로 같은** `CheckOK` 스킵을 낸다.
4. **아무것도 쓰지 않는다.** `.codex/` 파일을 만들거나 고치지 않고,
   `~/.codex/config.toml`도 건드리지 않는다. 읽기 전용 진단이며, 모든 입력 부재는
   실패가 아니라 침묵으로 떨어진다(fail-open).
5. **새 모듈 의존성이 없다.** TOML 파서는 `configtoml.go`와 같은 방식으로
   `regexp`/`strings`만 써서 손으로 짰다.

변경 파일 5개:

| 파일 | 성격 |
|---|---|
| `internal/codexwiring/skills.go` | 신규 — `[[skills.config]]` 읽기 전용 파서 |
| `internal/codexwiring/skills_test.go` | 신규 — 파서 단위 테스트 7건 |
| `internal/cli/doctor_codex.go` | 수정 — 두 하위 점검 + 홈 디렉터리 seam |
| `internal/cli/doctor_codex_test.go` | 수정 — 신규 테스트 5건 + seam 스텁 헬퍼 2개 |
| `internal/cli/doctor_golden_test.go` | 수정 — 골든 스냅샷 hermeticity 보강(아래 참조) |

---

## Evidence (증거 — 명령과 그 출력 그대로)

### RED 1단계 — 컴파일 실패 (seam 자체가 없음)

```
$ go test ./internal/codexwiring/...
# github.com/modu-ai/moai-adk/internal/codexwiring [.../codexwiring.test]
internal/codexwiring/skills_test.go:16:13: undefined: ParseSkillEntries
... (7건)
FAIL	github.com/modu-ai/moai-adk/internal/codexwiring [build failed]

$ go test -timeout 600s ./internal/cli/... -run 'Codex'
internal/cli/doctor_codex_test.go:70:2: undefined: codexWiringUserHomeDir
internal/cli/doctor_codex_test.go:71:21: undefined: codexWiringUserHomeDir
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

### RED 2단계 — seam만 넣고(파서는 `return nil` 스텁) 단언 수준 실패

컴파일 실패는 "테스트가 뭘 재는지"를 증명하지 못하므로, 빈 껍데기만 넣고 다시 재서
**현행 구현이 정확히 그 두 침묵을 낸다는 것**을 출력으로 받아 뒀다.

```
$ go test -timeout 600s ./internal/cli/ -run CheckCodexWiring
--- FAIL: TestCheckCodexWiring_UnwiredWithCodexInstalledWarns (0.00s)
    doctor_codex_test.go:113: unwired project with codex installed status = ok, want Warn: {Name:Codex Wiring Status:ok Message:not wired (claude-only project) — skipped Detail:}
    doctor_codex_test.go:116: action directive missing from Message (Detail is --verbose-only): {... Message:not wired (claude-only project) — skipped ...}
    doctor_codex_test.go:120: message does not name the absent path ".codex/hooks.json": {...}
    doctor_codex_test.go:120: message does not name the absent path ".codex/config.toml": {...}
--- FAIL: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)
    doctor_codex_test.go:148: stale home skills status = ok, want Warn: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
    doctor_codex_test.go:152: finding does not name the [[skills.config]] surface: {...}
    doctor_codex_test.go:155: finding does not point at ~/.codex/config.toml: {...}
    doctor_codex_test.go:160: finding does not quantify "3": {...}
    doctor_codex_test.go:160: finding does not quantify "4": {...}
    doctor_codex_test.go:160: finding does not quantify "1 enabled": {...}
    doctor_codex_test.go:160: finding does not quantify "2 disabled": {...}
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.842s

$ go test ./internal/codexwiring/... -run Skill
--- FAIL: TestParseSkillEntriesCanonicalOrder — returned 0 entries, want 2: []
--- FAIL: TestParseSkillEntriesReversedKeyOrder — returned 0 entries, want 1
--- FAIL: TestParseSkillEntriesEnabledAbsent — returned 0 entries, want 1
--- FAIL: TestParseSkillEntriesWhitespaceTolerance — returned 0 entries, want 1: []
--- FAIL: TestParseSkillEntriesSectionBoundary — returned 0 entries, want 1
FAIL	github.com/modu-ai/moai-adk/internal/codexwiring	0.574s
```

**공허성에 관한 정직한 구분.** 새로 넣은 doctor 테스트 5건 중 RED를 낸 것은 2건
(`UnwiredWithCodexInstalledWarns`, `StaleHomeSkillsReported`)이다. 나머지 3건
(`HealthyHomeSkillsNoFinding`, `AbsentHomeConfigSilent`, `ClaudeOnlyMachineStaysSilent`)은
성질상 변경 전후 모두 통과한다 — 새 행동을 요구하는 단언이 아니라 **깨지면 안 되는 불변**
(오탐 없음, 침묵 유지)을 못박는 회귀 가드다. 이 3건이 RED를 안 냈다는 사실 자체를
증거로 남긴다. 파서 테스트 7건 중 RED를 낸 것은 5건, `Empty`와 `SimilarHeaderRejected`는
스텁이 `nil`을 돌려주는 바람에 우연히 통과했다 — 진짜 파서를 넣은 뒤에야 의미가 생겼다.

### GREEN — 전 건 통과 (셀렉터가 0건을 잡은 게 아님을 이름으로 확인)

```
$ go test -timeout 600s -v ./internal/cli/ -run Codex
--- PASS: TestCheckCodexWiring_UnwiredWithCodexInstalledWarns (0.00s)
--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)
--- PASS: TestCheckCodexWiring_HealthyHomeSkillsNoFinding (0.00s)
--- PASS: TestCheckCodexWiring_AbsentHomeConfigSilent (0.00s)
--- PASS: TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent (0.00s)
--- PASS: TestCheckCodexWiring_InactiveProjectInformationalSkip (0.00s)
--- PASS: TestCheckCodexWiring_HealthyProjectOK (0.00s)
--- PASS: TestCheckCodexWiring_DivergenceAdvisesReTrust (0.00s)
--- PASS: TestCheckCodexWiring_ValidationFailureReported (0.00s)
--- PASS: TestCheckCodexWiring_MoaiNotOnPathReported (0.00s)
--- PASS: TestCheckCodexWiring_ConfigTableDriftReported (0.00s)
--- PASS: TestDoctor_CodexWiringRegistered (0.00s)
exit=0

$ go test -v ./internal/codexwiring/ -run Skill
--- PASS: TestParseSkillEntriesEmpty / CanonicalOrder / ReversedKeyOrder /
          EnabledAbsent / WhitespaceTolerance / SectionBoundary /
          SimilarHeaderRejected   (7/7)
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.462s
exit=0
```

### 런타임 관측 (실제 바이너리, 실제 미배선 프로젝트) — 필수 항목

이 머신에 `codex`는 실제로 깔려 있다. 코드 읽기가 아니라 실행으로 확인했다.

```
$ command -v codex
/Users/goos/.local/bin/codex
codex_exit=0

$ go build -o /tmp/t451-moai ./cmd/moai
build_exit=0

$ mkdir -p /tmp/t451-probe && git -C /tmp/t451-probe init -q .
$ cd /tmp/t451-probe && /tmp/t451-moai doctor 2>&1 | grep -iA2 codex
doctor_exit=0
```

경고가 실제로 렌더된 줄 (박스 테두리 제외, 그대로):

```
warn    Codex Wiring          codex is installed but this project is not wired (.codex/hooks.json and .codex/config.toml both absent) — the MoAI MCP server is not registered and the generated hooks cannot fire here; run moai init --agent codex; ~/.codex/config.toml declares 49 of 49 [[skills.config]] entries whose path no longer exists (0 enabled, 49 disabled) — remove the stale entries or restore the skill files
```

두 발견 모두 `--verbose` 없이 맨 `moai doctor`에서 보인다. 그리고 이 줄의
**49 / 49 / 0 enabled / 49 disabled는 건네받은 값이 아니라 이 실행이 직접 잰 값**이다.

반대편(claude 전용) 불변도 실행으로 확인했다:

```
$ cd /tmp/t451-probe && PATH=/usr/bin:/bin /tmp/t451-moai doctor 2>&1 | grep 'Codex Wiring'
doctor_exit=0
ok      Codex Wiring          not wired (claude-only project) — skipped
```

### 정적 검사 · 패키지 스위트

```
$ go vet ./internal/cli/... ./internal/codexwiring/...
vet_exit=0        (출력 없음)

$ gofmt -l internal/cli internal/codexwiring
→ 28개 파일이 나오지만 전부 선재 미포맷이며, 이번에 건드린 5개 파일은 목록에 없다.

$ go test ./internal/codexwiring/...
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.603s
codexwiring_exit=0

$ go test -timeout 600s ./internal/cli/...
cli_full_exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	521.676s
ok 패키지 17개, FAIL 0
```

### 부수 피해 1건 — 발견하고 귀속시키고 고쳤다

첫 전체 `internal/cli` 실행에서 `TestDoctorGolden_{Light,Dark,NoColor}` 3건이 깨졌다.
선재 실패인지 내 탓인지를 **가정하지 않고 재서** 갈랐다: 변경분을 stash하고 기반
`e79c010b8`에서 같은 셀렉터를 돌렸다.

```
$ git stash push --include-untracked ... && git rev-parse --short HEAD
e79c010b8
$ go test -timeout 600s ./internal/cli/ -run DoctorGolden
ok  	github.com/modu-ai/moai-adk/internal/cli	0.926s
baseline_golden_exit=0
```

기반이 초록이므로 내 탓이 맞다. 원인은 골든 하네스가 `HOME`은 고정하면서 PATH 조회는
고정하지 않는다는 데 있었다 — 미배선 프로젝트에 대한 판정이 이제 `codex` 설치 여부에
갈리므로, 스냅샷이 **개발자 머신의 PATH에 좌우되는** 상태가 된 것이다. CI에는 codex가
없고 이 머신에는 있으니, 골든을 재생성하는 쪽은 오답이다(이 머신 상태를 픽스처에 굽는
꼴). 하네스가 이미 `HOME`에 대해 쓰던 그 방식대로 `codex` 조회를 부재로 고정했고,
**골든 파일은 한 바이트도 안 바뀌었다**(`git status`에 `testdata` 변경 없음).

```
$ go test -timeout 600s ./internal/cli/ -run DoctorGolden
ok  	github.com/modu-ai/moai-adk/internal/cli	0.939s
golden_exit=0
```

---

## Baseline-attribution (귀속)

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t451`,
  브랜치 `WT-codex-wiring-doctor`, 측정 시점 `HEAD = e79c010b8` (develop 팁, 미커밋 작업트리).
- 위 모든 수치와 출력은 **이 트리에서 이번 실행으로** 얻은 것이다. 선행 카드나 리드 보고에서
  옮겨온 값은 없다. 특히 카드에 적혀 있던 "49 항목 / 49 결측 / enabled 0"은 근거로 삼지 않고
  런타임 관측으로 독립 재측정했으며, 결과가 일치했다.
- `internal/cli` 전체 스위트 521.676s는 이 머신의 로컬 측정이다. 전 패키지 판정은 CI 몫이며
  여기서 `go test ./...`는 돌리지 않았다(CLAUDE.local.md §4 부하 금지).
- 골든 선재/부수 판정은 같은 셀렉터를 stash 전후로 돌린 대조 측정에 귀속된다.

---

## Gaps (관측하지 않은 것)

- **`internal/cli`·`internal/codexwiring` 밖은 안 쟀다.** 다른 패키지에 대한 파급은
  이 실행이 말할 수 있는 범위가 아니다. CI 몫으로 남긴다.
- **CI에서 안 돌렸다.** 푸시하지 말라는 지시대로 브랜치를 밀지 않았으므로, darwin/windows
  매트릭스와 깨끗한 환경에서의 판정은 없다. 특히 골든 hermeticity 보강이 codex 없는
  러너에서 의도대로 동작하는지는 로컬 대칭 실험(`PATH=/usr/bin:/bin`)으로만 뒷받침된다.
- **커버리지 수치를 안 쟀다.** `-cover`를 돌리지 않았으므로 85% 기준 충족 여부를 주장하지
  않는다.
- **`golangci-lint`를 안 돌렸다.** `go vet`과 `gofmt`만 확인했다.
- **실제 codex 런타임과의 상호작용은 안 봤다.** 경고 문구가 사용자를 옳은 조치로 이끄는지는
  `moai doctor` 출력까지만 확인했고, `moai init --agent codex`를 실행해 실제로 배선이
  생기는지는 이번 범위 밖이다(카드가 쓰기를 금지했다).
- **`~/.codex/config.toml`의 49개 항목이 왜 전부 사라졌는지**는 조사하지 않았다. 이 카드는
  그 사실을 **보고**하게 만드는 일이지 원인을 밝히는 일이 아니다.

---

## Residual-risk (잔여 위험)

- **손으로 짠 TOML 파서의 한계.** 여러 줄 문자열, 리터럴 문자열(`'…'`), 인라인 테이블
  형태(`skills.config = [{...}]`)로 쓰인 항목은 인식하지 못한다. 그 경우 항목 수가 덜
  잡히고, 결과는 오탐이 아니라 **과소 보고**(침묵)로 떨어진다 — 진단이 fail-open이므로
  안전한 방향이지만, 침묵이 곧 건강함은 아니라는 점은 남는다.
- **경로 존재 판정은 `os.Stat` 한 번**이다. 권한 문제로 stat이 실패하는 경로는 "없음"으로
  집계된다. 홈 디렉터리 자기 소유 파일에서 흔한 상황은 아니지만 오탐 가능성은 0이 아니다.
- **`codex` PATH 조회로 "코덱스를 쓴다"를 추정한다.** 바이너리만 깔아 두고 이 프로젝트에서는
  쓸 생각이 없는 사용자에게는 원치 않는 경고가 된다. 완화책은 문구 자체가 조치 지시형이라
  무시 비용이 낮다는 점뿐이다.
- **골든 스냅샷의 결합.** 이제 골든 하네스가 `codexWiringLookPath`를 스텁한다. 이후 이
  seam의 이름이나 시그니처가 바뀌면 골든 테스트가 함께 깨진다 — 의도된 결합이지만
  결합이라는 사실은 남는다.
- **다른 진단 항목의 같은 종류 취약성은 손대지 않았다.** 골든 하네스는 PATH 조회 일반을
  고정하지 않으므로, 앞으로 머신 상태를 읽는 점검이 추가되면 같은 함정을 다시 밟는다.

---

## 범위 밖으로 남긴 것 (의도)

- `.codex/` 파일 생성·수리 없음, `~/.codex/config.toml` 수정 없음 — 카드의 [HARD] 제약.
- 묵은 스킬 항목 자동 정리 기능 없음. 보고만 한다.
- `moai init --agent codex` 경로 변경 없음.
- 푸시·develop 병합·PR 없음. 커밋만 이 브랜치에 남겼다.
