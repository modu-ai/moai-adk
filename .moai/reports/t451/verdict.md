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

---

# 수리 라운드 (repair round) — 감사 FAIL 대응

독립 sync-auditor가 `04f89d39d`에 대해 **FAIL**(Functionality 66 / Security 88 / Craft 68 /
Consistency 60)을 냈다. 보고서는 `.moai/reports/t451/sync-audit.md`(이 라운드에서 함께 커밋).
[HARD] 제약 3종은 전부 지켜졌다고 판정됐고, 떨어진 이유는 **건강한 설정에 거짓 지적을 내고
패널을 깨뜨린다**는 것이다. 아래는 여섯 결함의 수리와 그 증거다.

## D1 — `CODEX_HOME` 무시 (F1)

`codexStaleSkillFinding`이 자체 home seam(`codexWiringUserHomeDir`) + `filepath.Join(home,
".codex", "config.toml")` 하드코딩으로 경로를 만들었다. 같은 패키지의 `resolveCodexHomeDir()`
(`mcp_codex.go`)를 호출하도록 바꾸고, 파일명은 `path.Base(codexwiring.ConfigRelPath)`로
재사용했다. 중복 seam은 **삭제**했다 — 재사용이 seam을 옮겼으므로 죽은 seam을 남기지 않는다.
테스트의 `stubCodexHome`은 이제 `codexUserHomeDir` seam과 `CODEX_HOME` **양쪽**을 고정한다
(seam만 고정하면 개발자가 export한 `CODEX_HOME`이 판정을 좌우한다).

증거 — 실바이너리, 빈 `CODEX_HOME`:

```
$ cd /tmp/t451-repair-probe && CODEX_HOME=/tmp/t451-empty-codexhome \
    /tmp/t451-repair-moai doctor 2>&1 | grep -ac "stale skill"
0
$ ... | grep -a "warn    Codex Wiring"
│    warn    Codex Wiring          codex installed, project not wired — run moai init --agent codex   │
```

회귀 테스트 2종: `TestCheckCodexWiring_CodexHomeHonoured`(빈 `CODEX_HOME` ⇒ 지적 없음),
`TestCheckCodexWiring_CodexHomeConfigRead`(`CODEX_HOME` 안의 묵은 설정 ⇒ 지적, 표시 문자열은
`$CODEX_HOME/config.toml`, Detail은 실제 해석된 절대경로를 인용).

## D2 — `os.Stat` 오류 종류 뭉개기 (F2)

`serr != nil`을 3분기로 나눴다.

| stat 결과 | 판정 | 근거 |
|---|---|---|
| `err == nil` (파일) | missing 아님 | 존재한다 |
| `err == nil` (**디렉터리**) | missing 아님 | 존재한다. Codex가 디렉터리를 어떻게 다루는지는 **미관측**이라 추측으로 삭제를 권고하지 않는다 |
| `errors.Is(err, fs.ErrNotExist)` | **missing** | 관측된 부재 |
| 그 외(권한 거부·심링크 루프·I/O) | **미판정(indeterminate)** | missing 카운트에 절대 넣지 않는다. `--verbose` Detail에만 `N could not be checked and are NOT counted as missing`으로 공개 |

메시지가 "remove the stale entries"라는 파괴적 조치를 권고하므로, 완료하지 못한 stat을 부재로
집계하면 멀쩡한 등록을 지우라고 조언하게 된다. 그래서 미판정은 **분자에 들어가지 않는다**.
테스트: `TestCheckCodexWiring_IndeterminateStatNotMissing`(심링크 루프),
`TestCheckCodexWiring_DirectoryPathNotMissing`.

## D3 — multi-line 문자열 유령 항목 (F4) + 형제 위험

`ParseSkillEntries`에 `"""` / `'''` 리터럴 상태를 넣었다. 한 줄에서 구분자가 **홀수 번**
나오면 리터럴을 연 것으로 보고, 닫을 때까지 그 구간을 통째로 건너뛴다. 짝수(인라인
`x = """foo"""`, 구분자를 두 번 인용한 주석)는 여는 것으로 보지 않는다. 편향은 보수적으로
잡았다 — 애매하면 리터럴 안으로 취급해 **과소 보고(침묵)** 쪽으로 떨어뜨린다. 유령 항목은
유효한 TOML에 대한 거짓 지적이 되지만, 침묵은 그렇지 않기 때문이다. regexp/strings만 쓰며
새 의존성은 없다.

형제 위험 4종을 감사 지적대로 확인하고 각각 테스트를 붙였다.

| 위험 | 이미 옳았나 | 조치 |
|---|---|---|
| 뒤따르는 **다른 테이블**의 `path` 키 | 예 (`anyTableRe` 경계) | 회귀 테스트 추가 `…LaterTablePathDoesNotLeak` |
| 주석(헤더·양 키 뒤) | 예 | 회귀 테스트 추가 `…Comments` |
| CRLF 줄바꿈 | 예 (`TrimSpace`가 `\r` 제거) | 회귀 테스트 추가 `…CRLF` |
| 따옴표 친 `enabled = "true"` | **아니오** — false로 강등 | 파서 수정 + `…EnabledQuoted` |

추가로 리터럴 상태가 **일방통행 함정이 아님**을 고정했다(`…MultilineStringCloses`): 닫힌
리터럴 뒤의 실제 항목은 여전히 파싱된다.

## D4 — 패널 폭 붕괴 (F3, 사용자 체감 최상위)

**측정한 실제 천장**: 커밋된 골든 픽스처의 메시지 컬럼 최장은 **113 runes**이다(감사 보고서의
108이 아니라). 픽스처 박스 폭 자체는 152 runes.

```
$ python3 …  # doctor-nocolor.golden, 메시지 컬럼(offset 34) 최장
114 |  zone-registry.md not found at ".claude/rules/moai/core/zone-registry.md" — run `moai constitution list` to verify
   (선행 공백 1 제외 ⇒ 113)
$ 박스 테두리 행: 152 runes
```

구조를 바꿨다. 각 지적은 이제 `codexFinding{summary, detail}` 2단이며, `Message`는 summary만,
`Detail`은 전문을 싣는다. `joinCodexSummaries`가 113 runes를 **경계로 강제**하고, 넘치면
**꼬리부터** 버린 뒤 `(+N more, see --verbose)`로 개수를 공개한다 — 지시문을 실은 지적은 항상
선두이므로 절단에서 살아남는다. 카드가 요구한 두 조건이 함께 성립한다:

```
│    warn    Codex Wiring          codex installed, project not wired — run moai init --agent codex; ~/.codex/config.toml: 49 stale skill entries        │
```

`moai init --agent codex`가 **plain `moai doctor`에 그대로 보인다**. 실측 폭:

| 실행 | 이 라운드 | 감사 라운드 |
|---|---|---|
| `moai doctor` 패널 최장 (runes) | **154** | — |
| 같은 값 (bytes, 감사와 같은 단위) | **462** | **1272** |
| `Codex Wiring` 메시지 (runes) | **110** (≤ 113) | 390 |

`--verbose`의 Detail도 같은 폭 밴드로 **줄바꿈**한다(`wrapCodexDetail`) — Detail은
`--verbose`에서 자기 행이 되므로 여기서 묶지 않으면 폭 경계가 반쪽만 성립한다. 실측 Detail
행: 124 / 109 / 18 / 109 / 90 runes, 전부 152 밴드 안.

> **관측했으나 고치지 않은 것**: `--verbose` 전체 패널은 여전히 2607 runes까지 벌어진다.
> 원인은 이 카드 밖의 **기존** 점검 `Home Disk Usage`의 Detail(단일 행 2604 runes)이다.
> 범위 규율에 따라 손대지 않았고, 사실만 기록한다.

## D5 — 공허한 단언 (C3)

원인은 감사 진단대로였다: `check.Message + " " + check.Detail` 위에서 `"3"`, `"4"` 같은
**맨 부분문자열**을 찾았기 때문에 자리 뒤바꿈이 통과했고, "지시문은 Message에 있어야 한다"는
요구는 이 분기에서 아예 검사되지 않았다. 다시 썼다:

- **순서 고정 구절**로 단언 (`declares 5 [[skills.config]] entries`,
  `3 with a path that no longer exists`, `(1 enabled, 2 disabled, 0 unspecified)`).
- **뒤바뀐 형태의 부재**도 함께 단언 — 자릿수만 들어 있으면 통과하는 일이 없게.
- Message 요구는 **`check.Message` 단독**으로 단언 (연결 문자열 아님).
- 빈 경로 가드는 **카운트에 미치는 효과**로 단언 (`TestCheckCodexWiring_EmptyPathEntryNotCountedMissing`:
  path 키 없는 항목 2개가 분자에 들어가지 않고 분모에는 들어간다).

세 뮤턴트를 **스크래치 사본에 `-overlay=`로 주입**해 직접 재현했다(작업 트리 무변경):

| 뮤턴트 | 변경 | 감사 라운드 | 이번 |
|---|---|---|---|
| M6 | `Sprintf` 분자/분모 뒤집기 | 생존 (ok) | **사망** `M6_EXIT=1` |
| M4 | `— remove the stale entries or restore the skill files` 삭제 | 생존 (ok) | **사망** `M4_EXIT=1` |
| M7 | `if e.Path == ""` → `if false` | 생존 (ok) | **사망** `M7_EXIT=1` |

## D6 — 미검증 전제 (F8, F9)

**(a) `enabled` 삼상태.** `SkillEntry.Enabled`를 `bool`에서 `SkillEnabled`
(`Unspecified` / `True` / `False`)로 바꿨다. 키 부재를 false로 접는 것은 **이 저장소가 관측한
적 없는 Codex 기본값을 단언**하는 일이므로 하지 않는다. 분류는 **선언된 대로** 보고하며
(`0 enabled, 49 disabled, 0 unspecified`), 메시지는 Codex의 해석에 대해 아무것도 주장하지
않는다. 따옴표 친 `"true"`는 액면대로 true로 읽는다 — false로 읽으면 **살아 있는 등록을 묵은
기록으로 강등**하는, 둘 중 더 해로운 오독이 된다. 분류는 폭 규율에 따라 Detail에만 싣는다.

**(b) "the MoAI MCP server is not registered" 삭제.** 그 단정은 **프로젝트 레이어** 두 파일의
부재만 보고 내린 전역 주장이었다. 문구를 관측 범위로 좁혔다: `this project registers no MoAI
MCP server`. 사용자 레이어를 읽지 않은 채 전역 상태를 주장하지 않는다. 회귀 테스트가 옛 문구의
**부재**를 단언한다.

## 골든 커버리지 부채

핀은 유지했다(감사도 옳은 수정으로 판정). warn 분기 커버리지는 **골든 픽스처가 아니라 직접
단언**으로 넣었다. 이유:

> warn 분기의 Detail은 **해석된 설정 경로**를 인용한다. 테스트에서는 `t.TempDir()`,
> 실사용에서는 사용자의 실제 홈이다. 바이트 비교 스냅샷은 (a) 머신 의존이 되거나 — 핀이
> 막으려던 바로 그 hermeticity 실패 — (b) 그 경로를 정규화해 **단언하려는 증거 자체를
> 지워야** 한다. 깨진 불변식은 **폭**이므로 폭을 직접, 실제 렌더 출력 위에서, plain과
> verbose 양쪽에서 단언한다.

`TestCheckCodexWiring_RenderedPanelStaysInBand`가 49개 항목(1272-컬럼을 만든 실제 인구)으로
`renderDoctorGroups`를 통과시키고 박스 최장 행을 152 runes 밴드에 대해 검사한다. 이 테스트는
D4 수리 **전에 실제로 RED였다**(verbose 318 runes) — Detail 줄바꿈을 넣고서야 GREEN이 됐다.
`TestCheckCodexWiring_MessageWidthStaysInBand`가 Message 쪽 경계와 지시문 보존을 함께 고정한다.

## Verification (이 라운드, 이 트리 `WT-codex-wiring-doctor`)

```
$ go vet ./internal/cli/... ./internal/codexwiring/...
VET_EXIT=0
$ go test -count=1 ./internal/codexwiring/...
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.654s      CODEXWIRING_EXIT=0
$ go test -count=1 -race ./internal/codexwiring/...
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	1.697s      CODEXWIRING_RACE_EXIT=0
$ go test -count=1 -race -timeout 900s ./internal/cli/ -run 'Codex|codex|Doctor|doctor'
ok  	github.com/modu-ai/moai-adk/internal/cli	197.440s            CLI_CODEX_RACE_EXIT=0
   (DATA RACE 경고 0건 — 감사가 남긴 "새 전역 seam은 race로 관측 안 됨" 간극을 닫는다)
$ grep -nE "WriteFile|os\.Create|MkdirAll|os\.Remove|os\.Rename|OpenFile|Chmod|Truncate|io\.Copy" \
    internal/cli/doctor_codex.go internal/codexwiring/skills.go
WRITE_PRIMITIVE_GREP_EXIT=1        (무매치)
$ shasum -a 256 ~/.codex/config.toml     # doctor 여러 번 실행 전후 동일
8adec56f13e1cb6baafdb89a406f907f98b002d9136ca1cbf6f40a23194b4ae7
$ git status --short -- internal/cli/testdata/
   (무출력 — 픽스처 바이트 불변)
```

### `internal/cli` 전체 스위트 — 승계된 flake 1건

전체 `./internal/cli/...` 1회차에서 `TestResolveWorktreeExistingBranch_RejectsBadUsage` /
`…_MaterializeErrorPropagates` 2건이 실패했다. **이 카드가 만든 결함이 아니다.** 귀속을 실측했다:

```
# 내가 바꾼 4개 파일을 HEAD 내용으로 되돌린 overlay(작업 트리 무변경)
$ go test -count=1 -timeout 600s -overlay=o_BASE.json ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	438.601s      BASE_CLI_EXIT=0
# 같은 baseline을 -race로
$ go test -count=1 -race -overlay=o_BASE.json ./internal/cli/ -run 'TestResolveWorktreeExistingBranch|TestLauncherWorktree'
WARNING: DATA RACE  (7건)
--- FAIL: TestResolveWorktreeExistingBranch_NoFlagIsNoop
--- FAIL: TestResolveWorktreeExistingBranch_RejectsBadUsage
--- FAIL: TestResolveWorktreeExistingBranch_WiresAndStrips        BASE_WT_RACE_EXIT=1
```

`worktree_branch_flag_test.go`의 세 테스트는 `t.Parallel()`을 선언한 채 패키지 전역
(`findProjectRootFn`, `launcherWorktreeMaterialize`)을 동기화 없이 덮어쓴다 — **수리 이전
baseline에서 이미 data race**다. 이 라운드의 변경은 인터리빙을 흔들어 그 잠복 결함을 한 번
드러냈을 뿐이다. 격리 실행(`-run 'TestResolveWorktreeExistingBranch'`)은 통과하고,
재실행에서는 재현되지 않았다. **범위 규율에 따라 고치지 않았고, 이름을 밝혀 남긴다.**

## 이 라운드에서 결함이 아니라고 판단한 것 (운영자 판정 요청)

- **F5(TOML 이스케이프 미해제) / F6(`~`·상대경로 미해석) / F7(리터럴 문자열 `path = '…'`)**:
  카드가 지시한 6개 결함에 포함되지 않았고, 감사도 optional로 분류했다. 셋 다 방향이
  **거짓 지적 쪽**이라 무해하지 않다는 점은 인정한다(특히 `~` 경로는 항상 stat 실패 →
  거짓 missing). 다만 감사가 실제 `~/.codex/config.toml` 표본에서 기계가 쓴 항목은 전부
  절대경로임을 확인했고, 범위 규율상 지시받지 않은 파서 표면을 넓히지 않았다. 별도 카드 감.
- **F10(억제 수단 부재)**: 설계 선택이며 카드의 수리 목록에 없다. 경고 피로 위험은 실재하지만
  opt-out 도입은 새 설정 표면이라 이 라운드의 범위가 아니다.
- **S3(다른 테스트 파일의 home seam 미고정)**: 이번 수리가 `codexUserHomeDir`로 seam을
  옮기면서 `doctor_disk_test.go` / `coverage_improvement_test.go` 쪽 미고정은 **그대로**다.
  그 경로들은 단언이 내용에 의존하지 않아 실패하지 않으나, hermeticity는 여전히 부분적이다.
  카드가 지시한 6개에 없어 손대지 않았다.
- **`Home Disk Usage`의 `--verbose` 2604-rune Detail 행**: 이 카드 밖 기존 점검이다. D4 수리로
  Codex Wiring은 밴드 안에 들어왔지만 verbose 패널 전체 폭은 이 항목 때문에 여전히 넓다.

### `internal/cli` 전체 스위트 — 3회 실측 (승계 flake 확정)

| 회차 | 명령 | 결과 |
|---|---|---|
| 1 | `go test -count=1 -timeout 600s ./internal/cli/...` | `CLI_EXIT=1` — worktree 2건 실패 (402s) |
| 2 | 같은 명령 | `CLI_RUN2_EXIT=1` — **10분 테스트 타임아웃**(머신 부하). worktree 실패 재현 안 됨 |
| 3 | `go test -count=1 -timeout 1200s ./internal/cli/` | `CLI_RUN3_EXIT=0` — `ok … 947.874s`, **실패 0건** |

3회 중 1회만 실패했고, 그 실패는 위에서 baseline overlay + `-race`로 **수리 이전에도 존재하는
data race**임을 실측했다. 결론: 승계된 flake이며 이 카드의 회귀가 아니다.
