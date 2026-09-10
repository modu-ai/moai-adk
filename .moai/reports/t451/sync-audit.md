# t451 sync-audit — `moai doctor` Codex Wiring 확장 독립 감사

- 대상 커밋: `04f89d39dba5015f76ba0e837a266f05d93e28ff`
- 브랜치 / 트리: `WT-codex-wiring-doctor` @ `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t451`
- 감사 일자: 2026-09-03
- 감사자: sync-auditor (적대적 자세, 소스 읽기 전용)

## Overall Verdict: **FAIL**

must-pass 차원인 **Functionality**가 임계값 미달이라 다른 점수와 무관하게 전체 FAIL이다. 카드가 명시한 HARD 제약(쓰기 금지 / fail-open / claude-only 무침묵)은 **전부 지켜졌다**. 떨어진 이유는 제약 위반이 아니라, 새로 추가한 판정이 **healthy한 설정에 대해 거짓 지적(false finding)을 내는 경로가 재현 가능하게 존재**하기 때문이다. 카드 지시문 자체가 "healthy config에 경고를 내는 쪽이 더 나쁜 실패"라고 못박았고, 그 실패를 이 트리·이 런에서 실제로 재현했다.

---

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 66/100 | **FAIL** (must-pass, 임계 80) | `declares 49 of 49` — `CODEX_HOME`를 빈 디렉터리로 지정해도 동일 출력 (F1). 실데이터 참양성은 확인됨(49건 실제 부재). |
| Security (25%) | 88/100 | PASS | `grep -nE "WriteFile\|os.Create\|MkdirAll\|os.Remove\|os.Rename\|OpenFile\|Chmod" internal/cli/doctor_codex.go internal/codexwiring/skills.go` → `exit=1` (무매치) |
| Craft (20%) | 68/100 | FAIL(비-must-pass) | 패널 폭 `159` → `1248` 컬럼 (C1); 뮤턴트 M6/M4/M7 전부 생존 (C3) |
| Consistency (15%) | 60/100 | FAIL(비-must-pass) | 같은 패키지의 `resolveCodexHomeDir()`(`mcp_codex.go:1757`) 미재사용, 세 번째 home seam 신설 (S1/S2) |

---

## Findings

### 확인된 결함 (blocking)

**F1 [High][blocking] `internal/cli/doctor_codex.go:175-179` — `CODEX_HOME`를 무시한다.**
`codexStaleSkillFinding()`이 `filepath.Join(home, ".codex", "config.toml")`을 하드코딩한다. 이 바이너리는 다른 곳(`internal/cli/mcp_codex.go:1757 resolveCodexHomeDir`, `codex_readiness.go`, `codex_launcher.go`)에서 `CODEX_HOME` 환경변수를 1급 규약(REQ-CL-005)으로 이미 존중한다. 즉 doctor의 이 하위 판정만 규약을 어긴다.
- 트리거 입력: `CODEX_HOME=<빈 디렉터리>` 상태로 `moai doctor --check "Codex Wiring"`.
- 실측: 여전히 `~/.codex/config.toml`을 읽어 `declares 49 of 49` 를 출력.
- 결과 분류: **거짓 지적** (Codex가 읽지도 않는 파일에 대해 경고) 이면서 동시에 **누락**(진짜 사용 중인 `$CODEX_HOME/config.toml`은 검사 안 함). 양방향 오답이다.
- 필요한 수리: `resolveCodexHomeDir()`를 호출하고, 표시 문자열(`codexHomeConfigDisplay`)도 해석된 경로로 바꾼다.

**F2 [Medium][blocking] `internal/cli/doctor_codex.go:193` — `os.Stat` 오류 종류를 뭉갠다.**
`if _, serr := os.Stat(e.Path); serr != nil` 는 ENOENT가 아닌 오류까지 "경로가 더 이상 존재하지 않는다"로 보고한다. 실측한 두 종류가 모두 여기 걸린다.
- 권한 거부: `err=... permission denied  isNotExist=false`
- 심볼릭 링크 루프: `err=... too many levels of symbolic links  isNotExist=false`
- 트리거 입력: 조회 권한이 없는 부모 디렉터리 아래에 놓인 SKILL.md(macOS에서는 Full Disk Access 미부여 터미널에서 `~/Documents` 계열이 같은 오류를 낸다), 또는 순환 심볼릭 링크.
- 결과 분류: **거짓 지적**. 게다가 메시지가 "remove the stale entries" 라고 **파괴적 행동을 권고**하므로, 멀쩡한 등록을 지우게 만들 수 있다.
- 필요한 수리: `os.IsNotExist(serr)` 인 경우에만 missing으로 세고, 나머지 오류는 침묵(또는 별도 "unreadable" 카운트)으로 강등한다.

**F3 [Medium][blocking] `internal/cli/doctor_codex.go:213-215` + `uikit` 렌더 — 메시지 길이 무제한이라 표가 깨진다.**
두 지적이 `strings.Join(problems, "; ")`로 한 줄에 합쳐져 390자 메시지가 되고, 패널 폭이 터미널 폭과 무관하게 넓어진다.
- 실측: 같은 명령의 패널 최장 줄이 짧은 체크(`Glamour Cache`) 159자 → `Codex Wiring` 1248자. 터미널은 `tput cols` = 80.
- 골든 픽스처의 기존 행 최장은 108자다. 즉 이 커밋이 doctor 출력에 넣은 문자열이 기존 최장의 3.6배다.
- 대상 인구가 정확히 "codex 설치 + 미배선 프로젝트"이므로, 이 기능을 보게 될 사용자 전원이 깨진 표를 본다.
- 필요한 수리: 지시문을 유지하되 문장을 짧게 자르거나, 지적을 개행/다중 행으로 렌더하도록 한다.

**F4 [Medium][blocking] 파서가 multi-line 문자열 안의 `[[skills.config]]`를 실제 항목으로 오인한다.**
`ParseSkillEntries`는 줄 단위이며 `"""` / `'''` 상태를 갖지 않는다.
- 트리거 입력 A: `description = """` 다음 줄에 `[[skills.config]]` + `path = "/nonexistent/phantom"`.
  실측 결과: `[{Path:/nonexistent/phantom Enabled:true}]` — 존재하지 않는 유령 항목 생성.
- 트리거 입력 Q: `'''` 리터럴 블록도 동일 (`{Path:/nonexistent/p2 Enabled:false}`).
- 결과 분류: **거짓 지적**. 유효한 TOML이 healthy한데도 경고가 뜬다.
- 참고: `args = [ "[[skills.config]]", ... ]` 형태의 배열 원소는 줄이 `"`로 시작해 매치되지 않아 안전(케이스 R = `[]`).

### 확인된 결함 (optional)

**F5 [Low][optional] TOML 이스케이프를 전혀 풀지 않는다.** `path = "C:\\Users\\x\\S.md"` → 실측 `Path:C:\\Users\\x\\S.md` (역슬래시 2개가 그대로). Windows 경로나 `\t`, `\u` 를 포함한 경로는 `os.Stat`이 반드시 실패 → 거짓 "missing". macOS/Linux 실사용에서는 희박하다.

**F6 [Low][optional] `~` 및 상대 경로 미해석.** 실측 `Path:~/.codex/skills/a/S.md`, `Path:skills/a/S.md` 가 원문 그대로 나온다. `~`는 항상 stat 실패(거짓 지적), 상대 경로는 **doctor를 실행한 CWD 기준**으로 판정되어 같은 설정이 실행 위치에 따라 다른 답을 낸다(비결정성). 기계가 쓴 항목은 절대경로라 실사용 위험은 낮지만(실제 `~/.codex/config.toml` 표본 확인), 손으로 편집한 항목에서는 실재한다.

**F7 [Low][optional] TOML 리터럴 문자열(작은따옴표) 미지원.** `path = '/a/S.md'` → `Path:` (빈 문자열). 방향은 안전(누락)하지만 분모(`len(entries)`)에는 계속 잡혀 "N of M"의 M이 부풀려진다. 동일하게 `path` 키가 없는 항목(케이스 I)도 분모에만 잡힌다 — 코드 주석은 이를 의도라 밝히지만, 메시지 문구("declares N of M ... entries whose path no longer exists")는 M을 "경로를 선언한 항목 수"로 읽히게 해 오독을 유발한다.

**F8 [Low][optional] `enabled` 오해석 두 가지.** `enabled = "true"`(문자열)는 실측 `Enabled:false`로 읽혀 **live breakage를 stale bookkeeping으로 강등**한다(심각도 하향 오보). 또한 `enabled` 키 부재를 false로 확정한 근거는 코드 주석의 "Codex가 미등록 스킬에 내리는 판정과 같다"는 **미검증 전제**다 — Codex의 실제 기본값을 확인한 증거는 커밋 어디에도 없다. 기본이 true라면 심각도가 통째로 뒤집힌다.

**F9 [Low][optional] 미배선 메시지의 단정이 과하다.** `"the MoAI MCP server is not registered"`는 **프로젝트 레이어** `.codex/config.toml`만 보고 내린 결론이다. 같은 함수가 직후에 읽는 사용자 레이어 `~/.codex/config.toml`에 `[mcp_servers.moai]`가 있으면 이 단정은 거짓이 된다. 바이트를 이미 손에 쥔 시점이라 확인 비용이 0에 가까운데도 확인하지 않는다(VCI §1 미관측 전제).

**F10 [Low][optional] 억제 수단이 없다.** codex를 설치했지만 이 프로젝트는 의도적으로 claude-only인 사용자는 `moai doctor`를 돌릴 때마다 영구히 경고를 본다. 게다가 stale-skill 지적은 **머신 전역**이라 codex와 무관한 모든 프로젝트에서 같은 문구가 반복된다. 카드 제약(#3, codex 없는 머신)은 지켰으나 opt-out 경로가 전무하다.

### 테스트 공허성 (Craft)

**C3 [Medium][blocking] 새 doctor 테스트 3종의 단언이 뮤테이션을 잡지 못한다.** 실제 뮤턴트를 `go test -overlay=`로 주입해 확인했다(트리 미변경).

| 뮤턴트 | 변경 | 결과 |
|---|---|---|
| M6 | `Sprintf(..., missing, len(entries), ...)` → `..., len(entries), missing, ...` (분자/분모 뒤집기, "4 of 3") | **생존 (ok)** |
| M4 | 메시지 끝의 지시문 `— remove the stale entries or restore the skill files` 삭제 | **생존 (ok)** |
| M7 | `if e.Path == ""` → `if false` (문서화된 빈 경로 가드 제거) | **생존 (ok)** |
| M1 | `case anyTableRe.MatchString(line):` → `case false:` | 사망 (FAIL) — 파서 테스트는 유효 |
| M2 | `SkillEntry{}` → `SkillEntry{Enabled: true}` | 사망 (FAIL) — 파서 테스트는 유효 |

원인: `TestCheckCodexWiring_StaleHomeSkillsReported`가 `combined := check.Message + " " + check.Detail` 위에서 `"3"`, `"4"` 같은 **부분 문자열**만 찾는다. 두 숫자가 자리만 바뀌어도 둘 다 여전히 존재하므로 통과한다. 또한 `combined`를 쓰는 탓에 **"지시문은 Message에 실려야 한다"는 카드 요구가 이 하위 판정에서는 전혀 검증되지 않는다** (미배선 분기 테스트는 `check.Message`만 보므로 이 요구를 제대로 지킨다 — 두 테스트의 강도가 비대칭이다). 저자가 밝힌 "doctor 테스트 5개 중 2개만 RED" 라는 자기 신고와 일치하는, 실재하는 공허성이다.

### 일관성 (Consistency)

**S1 [Medium][blocking] 기존 home 해석기 미재사용.** `internal/cli` 패키지 안에 이미 `codexUserHomeDir` seam + `resolveCodexHomeDir()`(CODEX_HOME 규약 포함)이 있는데, 세 번째 seam `codexWiringUserHomeDir`을 새로 만들었다(패키지 내 home seam이 `homeDirFn`, `codexUserHomeDir`, `codexWiringUserHomeDir` 3개가 됐다). F1의 직접 원인이다.

**S2 [Low][optional] 문자열 하드코딩.** `".codex"` 는 같은 패키지에 `codexHomeDirName` 상수로 이미 존재한다(`mcp_codex.go:1731`). `"config.toml"` 도 `codexwiring.ConfigRelPath` 와 개념이 겹친다. CLAUDE.local.md §14(하드코딩 방지)에 어긋난다.

**S3 [Low][optional] 테스트 위생 미완.** 이 커밋은 자기가 만진 `doctor_codex_test.go` 전 테스트에 `stubCodexHome`을 달았지만, `internal/cli/doctor_disk_test.go:222`·`doctor_golden_test.go:425`의 `runGroupedChecks(false, "")`와 `coverage_improvement_test.go`의 다수 `runDoctor(...)` 호출은 여전히 home seam을 고정하지 않는다. 그 경로들은 개발자의 실제 `~/.codex/config.toml`을 읽고 실제 경로를 stat한다(이 머신에서는 49개 항목). 단언이 내용에 의존하지 않아 실패는 안 나지만, 커밋이 표방한 hermeticity는 자기가 만진 파일에서만 성립한다.

### 판정: 골든 테스트 hermeticity 수정 (카드 항목 4)

**옳은 수정이다.** 근거:
- `git diff --name-only 04f89d39d^ 04f89d39d | grep testdata` → 무출력. `git show --stat | grep -c testdata` → `0`. 픽스처는 실제로 손대지 않았다.
- 핀을 제거한 뮤턴트(M8: `if name == "codex"` → `if false`)를 overlay로 주입해 돌리면 이 머신에서 `TestDoctorGolden_Light/Dark/NoColor` 3개가 전부 실패한다. 이 머신에는 `codex`가 `/Users/goos/.local/bin/codex`에 있다. 즉 핀이 없으면 스냅샷 판정이 실행 머신의 PATH에 좌우된다는 주장은 참이며, 핀은 필수다.
- 다만 **부작용은 있다**: 새 warn 분기의 렌더가 골든에 전혀 covered 되지 않는다. F3(폭 붕괴)이 골든에 잡히지 않은 이유가 바로 이것이다. 핀은 유지하되 warn 분기용 골든을 하나 더 두는 것이 옳다. (blocking 아님, optional 부채.)

### 판정: 카드 HARD 제약 3종

| 제약 | 판정 | 증거 |
|---|---|---|
| `.codex/` 생성·수리 금지, `~/.codex/config.toml` 쓰기 금지 | **PASS** | 두 파일에 쓰기 프리미티브 0건(위 grep, exit=1). 호출 대상은 `os.ReadFile`/`os.Stat`/`exec.LookPath`/`regexp` 뿐. 추가 라인 grep에서도 `WriteFile\|Create\|MkdirAll\|Remove\|Rename\|OpenFile\|Chmod` 무매치(주석의 `RefreshWiring` 언급 1건 제외). |
| advisory·fail-open, 게이트 금지, 미확인 입력은 침묵 | **부분 PASS** | 모든 분기가 `DiagnosticCheck` 반환이고 error를 만들지 않음. home 미해석/파일 부재/엔트리 0은 `""` 반환으로 침묵. 단 F2(권한 오류·심링크 루프)는 침묵이 아니라 **거짓 지적**으로 degrade한다 — fail-open 정신 위반. 패닉·행 경로는 발견 못 함(정규식 3개 모두 선형, 재귀 없음). |
| claude-only 사용자 무침묵 | **PASS** | `!wired && !codexInstalled` 조기 반환이 첫 분기. `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent`가 stale home config가 있어도 OK임을 검증하고, 이 단언은 실효성이 있다(반대 뮤테이션 시 기존 `MoaiNotOnPathReported`가 잡음). |

### 스코프 규율 (카드 항목 6)

카드 범위 밖 drive-by는 **없다**. 변경 6파일 전부가 (a) 두 신규 판정, (b) 골든 hermeticity 핀, (c) 자기 증거 파일(`.moai/reports/t451/verdict.md`) 에 귀속된다. `doctor_codex.go`의 대규모 들여쓰기 변경은 `else` 블록 도입에 따른 필연이며 기존 분기 의미를 바꾸지 않는다(배선된 프로젝트 경로는 이전과 동일한 검사 순서·동일 문구).

---

## 5-Section Evidence

### Claim

1. 커밋 `04f89d39d`는 `.codex/` 및 `~/.codex/config.toml`에 아무것도 쓰지 않는다.
2. 새 판정은 실데이터에서 참양성을 낸다(이 머신: 49건 실제 부재).
3. 새 판정은 `CODEX_HOME`을 무시해 거짓 지적을 낸다.
4. `os.Stat` 오류 종류 뭉개기로 권한 오류·심링크 루프가 거짓 "missing"이 된다.
5. 파서는 multi-line 문자열 안의 헤더를 실제 항목으로 오인한다.
6. 새 doctor 테스트는 숫자 자리 뒤바꿈·지시문 삭제·빈 경로 가드 제거를 잡지 못한다.
7. 골든 핀은 필수이며 픽스처는 미변경이다.
8. 합쳐진 메시지가 doctor 패널 폭을 1248 컬럼으로 밀어낸다.

### Evidence (명령 + 축자 출력)

```
$ git rev-parse HEAD
04f89d39dba5015f76ba0e837a266f05d93e28ff
$ git branch --show-current
WT-codex-wiring-doctor
```

```
$ go vet ./internal/cli/... ./internal/codexwiring/...
VET_EXIT=0
$ go test ./internal/codexwiring/... -count=1 -race
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	1.917s
$ go test ./internal/cli/... -count=1 -timeout 600s -run 'Codex|Doctor|doctor'
ok  	github.com/modu-ai/moai-adk/internal/cli	259.133s
[exited with code 0]
```

C1 — 쓰기 프리미티브 부재:
```
$ grep -nE "WriteFile|os\.Create|MkdirAll|os\.Remove|os\.Rename|OpenFile|Chmod" internal/cli/doctor_codex.go internal/codexwiring/skills.go
exit=1
```

C2 — 실데이터 참양성 + 메시지 실물:
```
$ go run ./cmd/moai doctor --check "Codex Wiring"
warn    Codex Wiring  codex is installed but this project is not wired (.codex/hooks.json and .codex/config.toml both absent) — the MoAI MCP server is not registered and the generated hooks cannot fire here; run moai init --agent codex; ~/.codex/config.toml declares 49 of 49 [[skills.config]] entries whose path no longer exists (0 enabled, 49 disabled) — remove the stale entries or restore the skill files
$ grep -c "skills.config" /Users/goos/.codex/config.toml
49
$ ls /Users/goos/.codex/skills
.system  hatch-pet
$ ls -l /Users/goos/.codex/skills/moai-design-tools
ls: /Users/goos/.codex/skills/moai-design-tools: No such file or directory
```

F1 — `CODEX_HOME` 무시:
```
$ CODEX_HOME=<빈 디렉터리> NO_COLOR=1 go run ./cmd/moai doctor --check "Codex Wiring" | grep -o "declares [0-9]* of [0-9]*"
declares 49 of 49
$ grep -n "codexUserHomeDir =" internal/cli/*.go
internal/cli/mcp_codex.go:1753:var codexUserHomeDir = os.UserHomeDir
```

F2 — `os.Stat` 오류 종류(별도 스크래치 프로그램, 트리 미변경):
```
.../statfix/locked/SKILL.md   err=stat ...: permission denied              isNotExist=false
.../statfix/loopA             err=stat ...: too many levels of symbolic links  isNotExist=false
.../statfix/broken            err=stat ...: no such file or directory       isNotExist=true
.../statfix/adir              err=<nil>                                     isNotExist=false
```

F4~F8 — 파서 적대적 입력 (스크래치에 `skills.go` 사본 + `anyTableRe`/`splitLines` 재현, `go run`):
```
A multiline-basic-string phantom   -> [{Path:/nonexistent/phantom Enabled:true}]
Q multiline-literal phantom        -> [{Path:/nonexistent/p2 Enabled:false}]
R array element phantom            -> []
B escaped quote in path            -> [{Path: Enabled:true}]
C windows escaped backslash        -> [{Path:C:\\Users\\x\\S.md Enabled:true}]
D literal single-quoted            -> [{Path: Enabled:true}]
E tilde path                       -> [{Path:~/.codex/skills/a/S.md Enabled:true}]
F relative path                    -> [{Path:skills/a/S.md Enabled:true}]
G enabled as string                -> [{Path:/x Enabled:false}]
H duplicate path keys              -> [{Path:/second Enabled:false}]
I no path key                      -> [{Path: Enabled:true}]
J CRLF                             -> [{Path:/a/S.md Enabled:true}]
K inline-table array form          -> []
L spaced header                    -> []
M comment after header             -> [{Path:/a Enabled:true}]
N quotes inside comment            -> [{Path:/a Enabled:true}]
O comment inside entry             -> [{Path:/a Enabled:true}]
P later table has path             -> [{Path:/a Enabled:false}]
```

C3 — 뮤테이션(overlay 주입, 트리 미변경):
```
$ go test -overlay=overlay_M6.json ./internal/cli/ -run 'TestCheckCodexWiring' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.996s        # 분자/분모 뒤집기 생존
$ go test -overlay=overlay_M4.json ./internal/cli/ -run 'TestCheckCodexWiring|TestDoctor_CodexWiring' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.931s        # 지시문 삭제 생존
$ go test -overlay=overlay_M7.json ./internal/cli/ -run 'TestCheckCodexWiring|TestDoctor_CodexWiring' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	0.937s        # 빈 경로 가드 제거 생존
$ go test -overlay=overlay_M1.json ./internal/codexwiring/ -run TestParseSkillEntries -count=1
--- FAIL: TestParseSkillEntriesSectionBoundary (0.00s)
$ go test -overlay=overlay_M2.json ./internal/codexwiring/ -run TestParseSkillEntries -count=1
--- FAIL: TestParseSkillEntriesEnabledAbsent (0.00s)
--- FAIL: TestParseSkillEntriesSectionBoundary (0.00s)
```

골든 — 핀 필요성 + 픽스처 미변경:
```
$ git diff --name-only 04f89d39d^ 04f89d39d | grep testdata
TESTDATA_LIST_EXIT=1        (무출력)
$ command -v codex
/Users/goos/.local/bin/codex
$ go test -overlay=overlay_M8.json ./internal/cli/ -run 'Golden' -count=1     # 핀 제거 뮤턴트
--- FAIL: TestDoctorGolden_Light (0.00s)
    doctor_golden_test.go:183: doctor output mismatch for doctor-light
--- FAIL: TestDoctorGolden_Dark (0.00s)
--- FAIL: TestDoctorGolden_NoColor (0.00s)
```

F3 — 렌더 폭:
```
$ NO_COLOR=1 go run ./cmd/moai doctor --check "Codex Wiring" | awk '{print length($0)}' | sort -rn | head -2
1248
1248
$ NO_COLOR=1 go run ./cmd/moai doctor --check "Glamour Cache" | awk '{print length($0)}' | sort -rn | head -2
159
159
$ tput cols
80
$ NO_COLOR=1 go run ./cmd/moai doctor --check "Codex Wiring" | grep -o 'codex is installed.*skill files' | awk '{print length($0)}'
390
$ sed -n 38,48p internal/cli/testdata/doctor-nocolor.golden | ... | sort -rn | head -3
456   108   89        # 기존 골든 행 최장 108자
```

### Baseline-attribution

모든 수치는 이 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t451`, HEAD `04f89d39dba5015f76ba0e837a266f05d93e28ff`, 브랜치 `WT-codex-wiring-doctor`, 2026-09-03 이 감사 세션에서 직접 실행해 관측한 출력이다. 커밋 메시지가 인용한 저자 측 수치는 **재사용하지 않았고**, 테스트·vet·doctor 실행을 이 트리에서 새로 돌렸다. 뮤테이션은 전부 `go test -overlay=`로 주입해 **작업 트리 파일은 한 바이트도 바뀌지 않았다**(`git status --short`의 추적 파일 변경 0). 스크래치 산출물은 세션 스크래치패드에만 있다.

### Gaps (관측하지 않은 것)

- `go test ./...` 전체 스위트는 **의도적으로 돌리지 않았다**(로컬 전체 스위트 금지). 판정 범위는 `./internal/cli/...`(race 미적용, 259s)와 `./internal/codexwiring/...`(race 적용)뿐이다. race 검출은 `internal/cli`에 대해 관측하지 않았다 — 새 패키지 전역 seam 2개(`codexWiringLookPath`, `codexWiringUserHomeDir`)가 병렬 테스트와 경합하는지는 미검증이다.
- Windows / Linux에서의 동작을 관측하지 않았다. F5(역슬래시 이스케이프)의 실제 피해 규모는 darwin에서만 추론했다.
- Codex CLI가 `enabled` 키 부재를 어떻게 해석하는지, 프로젝트 레이어 `.codex/config.toml`을 실제로 어떤 우선순위로 읽는지 **확인하지 않았다**. F8·F9는 코드가 세운 전제가 미검증이라는 지적이지, Codex의 실제 동작에 대한 주장이 아니다.
- macOS TCC(전체 디스크 접근) 하에서의 stat 오류 코드는 관측하지 않았다. F2는 일반 권한 거부와 심링크 루프로만 재현했다.
- 쓰기 부재는 **정적 증거**(호출 그래프상 write 프리미티브 무매치)로만 확인했다. `strace`/`dtruss` 수준의 런타임 쓰기 관측은 하지 않았다.
- `--fix`/`--export` 경로에서 이 체크가 어떤 부작용도 갖지 않음은 호출자 1곳(`doctor.go:235`) 확인으로만 판단했고, export 산출물 내용은 관측하지 않았다.
- 저자의 `verdict.md` 내용은 검증 대상으로 읽지 않았다(독립 판정 유지).

### Residual-risk

- 관측한 것이 전부 참이어도, 파서가 손으로 편집된 `~/.codex/config.toml`의 **아직 시도해보지 않은 TOML 표면**(인라인 테이블 안의 `skills.config`, 배열 값의 여러 줄 지속, `\u` 이스케이프, BOM 선두)에서 또 다른 거짓 지적을 낼 여지는 남는다. 손으로 만든 18개 케이스는 표면을 전부 덮지 않는다.
- F2를 `os.IsNotExist`로 좁혀 고치더라도, 심볼릭 링크가 가리키는 대상이 없을 때(케이스: broken symlink)는 여전히 ENOENT로 "missing" 판정된다. 그 판정이 옳은지는 Codex가 링크를 어떻게 다루는지에 달려 있고 확인하지 않았다.
- F10(억제 수단 부재)는 결함이라기보다 설계 선택이지만, codex를 전역 설치한 사용자 전원이 무관한 프로젝트에서 머신 전역 경고를 반복해서 보게 되므로, 시간이 지나면 doctor warn 전체가 무시되는 2차 손상(경고 피로)을 낳을 수 있다. 이건 이 감사에서 측정할 수 없는 위험이다.
- `internal/cli` 스위트를 `-race` 없이 돌렸으므로, 새 전역 seam이 유발하는 잠재적 데이터 레이스는 이 판정으로 배제되지 않는다.

---

## Recommendations (우선순위 순)

1. **F1**: `resolveCodexHomeDir()`를 호출하도록 바꾸고 표시 문구도 해석된 경로로. (blocking)
2. **F2**: `os.IsNotExist(serr)`로 좁히고, 그 외 오류는 침묵 또는 별도 카운트. (blocking)
3. **F3**: 메시지를 짧게 자르거나 지적을 다중 행으로 렌더. 80컬럼 터미널에서 표가 유지되는지 확인. (blocking)
4. **C3**: stale 테스트를 `check.Message` 단독 단언 + 전체 문구 완전 일치(또는 `"3 of 4"` / `"1 enabled, 2 disabled"` 같은 순서 고정 부분 문자열)로 강화. 빈 경로 항목 픽스처 1건 추가. (blocking)
5. **F4**: 파서에 `"""` / `'''` 블록 상태를 추가하거나, 최소한 multi-line 진입을 감지해 그 구간을 통째로 건너뛴다. (blocking)
6. **S1/S2**: 신설 seam 제거, `codexHomeDirName` 상수 재사용. (optional, 그러나 F1 수리와 함께 하면 비용 0)
7. **F9**: 이미 읽은 `~/.codex/config.toml` 바이트에 `InspectMCPTable`을 한 번 더 돌려 "not registered" 단정을 사실에 맞춘다. (optional)
8. **골든**: 핀은 유지하고 warn 분기 골든 픽스처를 추가해 F3 같은 렌더 회귀가 스냅샷에 잡히게 한다. (optional)
9. **S3**: `runGroupedChecks`/`runDoctor`를 호출하는 기존 테스트에도 home seam 고정을 확산. (optional)
