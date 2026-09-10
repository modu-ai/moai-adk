# t540 — 읽는 쪽 전수 (plan-phase 사전 측정)

- 카드: t540
- 트리: `.claude/worktrees/t540`, 브랜치 `WT-codex-path-escape`
- base: `9ce792637` (origin/develop tip, 2026-09-08 진입 시점 실측)

## Claim

`[[skills.config]]` 의 `path` 를 읽는 비테스트 지점은 **정확히 3곳**이고, 그중 경로를 stat 하는
소비자는 **2곳**이다.

## Evidence

`/usr/bin/grep -rn "ParseSkillEntries" --include='*.go' internal/` (정의 1 + 호출 3):

```
internal/cli/codex_skills_prune.go:120      entries := codexwiring.ParseSkillEntries(content)
internal/cli/doctor_codex.go:723            entries = codexwiring.ParseSkillEntries(raw)
internal/cli/codex_skills_disable.go:251    for _, e := range codexwiring.ParseSkillEntries(content) {
internal/codexwiring/skills.go:215          func ParseSkillEntries(...)   ← 정의
```

경로 소비 형태:

| # | 지점 | Path 로 무엇을 하는가 | 오독의 결과 |
|---|---|---|---|
| 1 | `codex_skills_prune.go:71-80` | `classifyCodexSkillPath` → stat | **windows 에서만 파괴적** (아래 정정) |
| 2 | `doctor_codex.go:825-835` | `classifyCodexSkillPath` → stat | 진단 오보 |
| 3 | `codex_skills_disable.go:251` | 문자열 동등 비교(자기 매칭) | 중복 엔트리 발행 |

### [정정 2026-09-08] 1번의 파괴성은 무조건이 아니라 플랫폼 조건부다

이 문서의 최초판은 1번을 「못 찾으면 삭제」라고 **한정 없이** 적었다. 그건 과장이다.
plan-audit(D3)이 지적했고, `internal/cli/codex_skills_prune.go:75-94` 를 다시 읽어 확인했다.

`Eligible: true` 는 두 조건이 **모두** 설 때만 나온다:

1. `classifyCodexSkillPath` 가 `codexPathAbsolute` 또는 `codexPathHomeRelative` 이고,
2. `osStatFn` 이 `fs.ErrNotExist` 를 낸다.

`codexPathRelative` 와 `codexPathOddlyFormed` 는 **비파괴 skip** 이다
(`"relative path — no observed resolution base"` / `"oddly-formed path — not resolvable here"`).

그런데 `classifyCodexSkillPath` (`doctor_codex.go:668-677`) 는 `filepath.IsAbs` 를 먼저 돌린다:

- **darwin/linux**: `IsAbs(C:\...)` = false → `\` 포함 → `codexPathOddlyFormed` → **비파괴 skip**
- **windows**: `IsAbs(C:\...)` = true → `codexPathAbsolute` → stat 실패 → **삭제**

즉 파괴 경로는 실재하지만 **windows 에서만** 열린다. 하필 이 카드가 다루는 플랫폼이라 위험의
크기는 줄지 않는다 — 줄어드는 것은 주장의 범위다. 「모든 플랫폼에서 파괴적」은 거짓이고
「windows 에서 파괴적」이 참이다.

### [정정 2026-09-08] `filepath.ToSlash` / `FromSlash` 는 unix 에서 항등이다

이 호스트에서 직접 쟀다 (`go run .moai/reports/t540/lab/ts.go`):

```
GOOS=darwin Separator='/'
ToSlash(backslash)  = "C:\\Users\\u\\.agents\\skills\\foo\\SKILL.md" changed=false
FromSlash(slash)    = "C:/Users/u/.agents/skills/foo/SKILL.md" changed=false
IsAbs(slash form)   = false
IsAbs(backslash)    = false
```

`ToSlash` 는 `Separator` 를 `/` 로 바꾸는 함수이고 darwin 에서 `Separator == '/'` 라 아무것도
바꾸지 않는다. **결과**: 백슬래시 리터럴을 darwin 에서 발행자에 먹여도 `ToSlash` 뒤에 `\` 가
그대로 남아 `:245` 가드에 걸린다 — 이 함수를 그대로 쓰면 「플랫폼 무관 단위 실측」이 불가능하다.

수리는 분리자를 **주입 가능한 인자**로 빼는 것이다. 이 저장소에 이미 있는 seam 관례
(`osStatFn`, `codexUserHomeDir`) 와 같은 형태다. 그러면 테스트가 `'\\'` 를 직접 넘겨 darwin
에서도 windows 의미를 잰다.

## 파서가 이스케이프를 디코드하지 않는다

`internal/codexwiring/skills.go:174`

```go
skillPathKeyRe = regexp.MustCompile(`^path\s*=\s*"([^"]*)"\s*(#.*)?$`)
```

`[^"]*` 를 그대로 캡처한다 — `\\` 를 `\` 로 되돌리지 않는다. 게다가 `\"` 를 담은 경로는
정규식이 첫 `"` 에서 끊어 아예 다른 값을 잡는다.

## 발행자의 거부 조건

`internal/cli/codex_skills_disable.go:245`

```go
if strings.ContainsAny(skillPath, "\"\\\n\r") {
    return skip("the path contains a character this config format cannot carry verbatim (%q)", skillPath)
}
```

`skillPath` 는 `resolveCodexSkillMirrorPath` 가 `filepath.Join` 으로 만든다 (`:148`, `:158`, `:160`).
windows 에서 `filepath.Join` 은 `\` 로 잇는다 → 모든 windows 경로가 이 조건에 걸린다.

## Gaps (관측하지 않은 것)

1. **windows 실행 미관측.** 위 연결고리는 소스 판독 + Go stdlib 의 `os.PathSeparator` 규약에
   기댄 구조적 추론이다. 이 머신에서 windows 바이너리를 돌린 적 없다. 백슬래시 경로를
   `upsertCodexSkillDisable` 에 직접 먹이는 단위 테스트가 플랫폼 무관하게 이를 실측으로
   바꾼다 — run-phase 첫 산출물 후보.
2. **거부 분기에 테스트가 0건이다.** `/usr/bin/grep -rn 'cannot carry verbatim' --include='*.go' internal/`
   → 생산 코드 1줄만 매치. 이 [HARD] 안전 성질을 지키는 테스트가 없다.
3. prune 이 실제로 오독 경로를 지우는 것을 재현하지 않았다(구조적 판단).

## Residual-risk

t533(유령 49건)이 같은 파일군을 만진다. 파서를 고치면 t533 의 계수 기준이 바뀔 수 있다 —
순서를 리드에 확인해야 한다(카드 [HARD] 명시).
