# SPEC 감사 보고 — SPEC-CODEX-SKILL-PATH-SLASH-001 (카드 t540)

- Iteration: 1/2 (Tier M ceiling)
- **Verdict: FAIL**
- Overall Score: 0.81 (Tier M 임계 0.80) — **FAIL 은 점수가 아니라 blocking 결함 F-1/F-2 에서 나온다**
- 감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t540`, 브랜치 `WT-codex-path-escape`, HEAD `9ce792637`
- 저자 추론 맥락은 M1 Context Isolation 에 따라 무시했다 (Reasoning context ignored per M1 Context Isolation). 판독 대상은 `spec.md` / `plan.md` / `acceptance.md` / `progress.md` 와 인용된 Go 원본뿐이다.
- `audit_model` 설정이 `.moai/config/sections/` 에 없다 → Claude 단독 감사. 교차 모델 백엔드 호출 없음.

---

## 1. Claim (무엇을 주장하는가)

1. **F-1 [critical]** REQ-CSPS-002 / plan M2 가 지정한 정규화 수단 `filepath.ToSlash` 는 **darwin·linux 에서 백슬래시를 변환하지 않는다**(항등 함수). 따라서 AC-CSPS-002 가 스스로 주장하는 "platform-independent — darwin, linux, windows 에서 같은 뜻으로 돈다" 는 **거짓**이고, AC-CSPS-002·AC-CSPS-007 은 이 카드가 실행 가능한 유일한 호스트에서 **통과할 수 없다**. `spec.md §G` gap 1 의 "AC-CSPS-002 가 구조적 추론을 플랫폼 무관 실측으로 바꾼다" 도 같이 무너지고, t502 **F4 종결 주장**도 근거를 잃는다.
2. **F-2 [major]** M3(reader-side `FromSlash`) 는 **어느 호스트에서도 RED 를 세울 수 없다**. AC-CSPS-004·AC-CSPS-005 는 카드 호스트에서 base 와 바이트 동일한 코드 경로를 밟는다. cycle_type=tdd 와 `acceptance.md §D.2`("추가된 모든 테스트는 변경 전 RED 를 시연한다")를 M3 가 충족할 방법이 명시돼 있지 않다.
3. **F-3 [major]** `prune` 오독의 **파괴성 전제가 이 카드가 만들어내는 경로 형태에 대해 과장돼 있다**. 삭제는 분류가 absolute/home-relative 이고 stat 이 `ErrNotExist` 일 때만 일어난다. slash·backslash 오독이 떨어지는 두 분류(`codexPathRelative`, `codexPathOddlyFormed`)는 **비파괴 skip** 이다. `§G` gap 3 은 `:102` 만 인용하고 그 앞의 분기 스위치(`:76-94`)를 빠뜨린다 — VCI §1.1 surface 4(권고 전제 주장).
4. **F-4 [minor]** `len(matches) > 1` 분기는 `:257`(문서는 `:256`)이며 **비교를 수행하지 않는다**. 비교 지점은 `:252` 하나뿐이다. `§D.4` 마지막 관측과 plan M1 두 번째 항목이 "두 분기가 매치의 정의를 놓고 어긋날 수 있다"고 쓴 것은 존재하지 않는 두 번째 비교 지점을 전제한다.
5. **F-5 [minor]** `§B.1` "Measured facts" 표 2행의 뒷절("so on Windows it always carries `\`")과 `§D.4` 의 "on Windows it still classifies `codexPathAbsolute` … and stats fine" 는 **미관측 windows 런타임 추론**인데 측정 사실로 적혀 있다. `§G` gap 1 이 앞의 것만 정정하고, 뒤의 것은 gap 목록에 없다.
6. **F-6 [minor]** AC-CSPS-006 의 "기존 스위트 전부 통과" 절반은 셀렉터 0매치·미컴파일에서 공허하게 초록이 된다. AC-CSPS-008 과 달리 **대조(실행된 테스트 수)가 없다**.
7. **F-7 [minor]** REQ-CSPS-001 의 GEARS 형태가 다섯 패턴 밖이다("Before …, the maintainer shall have measured").
8. **PASS 로 판정한 것**: AC-CSPS-001 게이트는 장식이 아니라 실효 게이트다. AC-CSPS-003·AC-CSPS-007 의 판별력, AC-CSPS-008 의 t543 규율, 읽는 쪽 전수 3건, `§D.2` 순서 제약, F4/F5 범위 서술은 원본과 일치한다.

---

## 2. Evidence (명령과 그 출력 그대로)

### E-1 — `filepath.ToSlash` 는 darwin 에서 백슬래시를 바꾸지 않는다 (F-1)

```
$ go doc path/filepath.FromSlash
func FromSlash(path string) string
    FromSlash returns the result of replacing each slash ('/') character in
    path with a separator character. Multiple slashes are replaced by multiple
    separators.
```

```
$ go run /tmp/t540_pchk.go
GOOS: darwin
in="C:\\Users\\u\\SKILL.md" ToSlash="C:\\Users\\u\\SKILL.md" FromSlash="C:\\Users\\u\\SKILL.md" IsAbs=false
in="C:/Users/u/SKILL.md" ToSlash="C:/Users/u/SKILL.md" FromSlash="C:/Users/u/SKILL.md" IsAbs=false
in="/var/folders/x/SKILL.md" ToSlash="/var/folders/x/SKILL.md" FromSlash="/var/folders/x/SKILL.md" IsAbs=true
```

(`ToSlash` 는 `Separator` 를 `/` 로 바꾸는 함수이고 darwin 에서 `Separator == '/'` 이므로 항등이다. 백슬래시는 그대로 남는다.)

이에 대응하는 SPEC 문면:

- `spec.md:87` — "the publisher shall normalize it with `filepath.ToSlash` and publish it"
- `plan.md:74` — "Apply `filepath.ToSlash` to `skillPath` before the `ContainsAny` guard"
- `acceptance.md:52-57` — "When it is called with a Windows-shaped path literal such as `C:\Users\u\.codex\skills\probe\SKILL.md`, **Then** … the emitted content contains the line `path = "C:/Users/u/.codex/skills/probe/SKILL.md"` … Platform-independent: the input is a string literal, so the test runs and means the same thing on darwin, linux, and windows."

darwin 에서 `ToSlash` 를 적용해도 `skillPath` 는 백슬래시를 유지 → `:245` 의 `ContainsAny` 에 그대로 걸림 → `codexSkillDisableSkipped`. AC-CSPS-002 의 Then 은 darwin 에서 성립 불가다.

### E-2 — 게이트가 남긴 stat 경로: 두 리더는 변화가 없다 (F-2)

E-1 의 세 번째 행: darwin 에서 `FromSlash` 도 항등이다. `acceptance.md:88-90` 은 이를 스스로 인정한다 — "On a `/`-separator host both arms exercise the identity path". 즉 AC-CSPS-004·005 는 base 트리에서도 동일하게 초록이다. 그리고 E-1 두 번째 행(`IsAbs("C:/Users/u/SKILL.md") = false`)은 darwin 에서 windows 형태 slash 경로를 픽스처로 쓰는 우회로도 막는다 — 그 경로는 `codexPathRelative` 로 분류돼 stat 에 도달하지 못한다.

### E-3 — prune 의 삭제 분기와 그 앞의 가드 (F-3)

```
$ /usr/bin/grep -n "" internal/cli/codex_skills_prune.go | sed -n '75,109p'
 75:	var statPath string
 76:	switch classifyCodexSkillPath(e.Path) {
 77:	case codexPathAbsolute:
 78:		statPath = e.Path
 79:	case codexPathHomeRelative:
 ...
 87:	case codexPathRelative:
 91:		return skip("relative path — no observed resolution base")
 92:	default:
 93:		return skip("oddly-formed path — not resolvable here")
 94:	}
 96:	_, err := osStatFn(statPath)
 98:	case err == nil:
101:		return skip("the path resolves")
102:	case errors.Is(err, fs.ErrNotExist):
103:		return codexSkillPruneVerdict{Entry: e, Eligible: true}
```

`Eligible: true` 는 `codexPathAbsolute` / `codexPathHomeRelative` + `ErrNotExist` 조합에서만 나온다.

### E-4 — 비교 지점은 하나뿐이다 (F-4)

```
$ /usr/bin/grep -n "" internal/cli/codex_skills_disable.go | sed -n '245,259p'
245:	if strings.ContainsAny(skillPath, "\"\\\n\r") {
246:		return skip("the path contains a character this config format cannot carry verbatim (%q)", skillPath)
247:	}
249:	lines, term := codexwiring.SplitConfigLines(content)
250:	var matches []codexwiring.SkillEntry
251:	for _, e := range codexwiring.ParseSkillEntries(content) {
252:		if e.Path == skillPath {
253:			matches = append(matches, e)
254:		}
255:	}
257:	if len(matches) > 1 {
258:		return skip("%d entries declare this path; collapsing duplicates is `moai clean --codex-skills`'s job, not this verb's", len(matches))
```

### E-5 — 읽는 쪽 전수 재유도 (감사자가 직접 실행)

```
$ /usr/bin/grep -rn "ParseSkillEntries" --include='*.go' internal/
internal/cli/codex_skills_prune_enabled_test.go:26,120,123      (테스트)
internal/cli/codex_skills_prune.go:120:	entries := codexwiring.ParseSkillEntries(content)
internal/cli/codex_skills_prune_test.go:202,253,293             (테스트)
internal/cli/doctor_codex.go:723:	entries = codexwiring.ParseSkillEntries(raw)
internal/cli/codex_skills_disable_test.go:37,50                 (테스트)
internal/cli/codex_skills_disable.go:251:	for _, e := range codexwiring.ParseSkillEntries(content) {
internal/codexwiring/skills.go:215:func ParseSkillEntries(content []byte) []SkillEntry   ← 정의
(이하 codexwiring/*_test.go)

$ /usr/bin/grep -rn "ParseSkillEntries" --include='*.go' . | /usr/bin/grep -v '^./internal/'
(무출력 — internal/ 밖에는 호출자가 없다)
```

**비테스트 호출 3건 — 전수는 옳다.** SPEC `§B.2` 의 3행 표와 `reader-census.md` 는 정확하다. 세 지점 모두 AC 로 덮인다(1·2 → AC-CSPS-004/005, 3 → AC-CSPS-007). 이 항목은 결함 없음.

### E-6 — `classifyCodexSkillPath` 순서 제약 (SPEC 서술이 옳다)

```
$ sed -n '663,678p' internal/cli/doctor_codex.go
663:// classifyCodexSkillPath maps a declared path onto its shape. Ordering is
664:// load-bearing: IsAbs runs before the backslash check ...
667:func classifyCodexSkillPath(p string) codexSkillPathShape {
668:	if filepath.IsAbs(p) { return codexPathAbsolute }
671:	if p == "~" || strings.HasPrefix(p, "~/") { return codexPathHomeRelative }
674:	if strings.HasPrefix(p, "~") || strings.ContainsRune(p, '\\') { return codexPathOddlyFormed }
677:	return codexPathRelative
```

`spec.md §D.2`(:118-120) 와 `plan.md` M3(:88-91) 의 서술 — "IsAbs 먼저, `\` 는 oddly-formed, 분류는 DECLARED form 에, `FromSlash` 는 분류 **이후** stat 대상에" — 은 원본과 일치한다. **뒤집힘 없음.**

### E-7 — Must-Pass 기계 검증

```
$ /usr/bin/grep -o 'REQ-CSPS-[0-9]*' spec.md | sort -u   → 001..009, 결번·중복 0
$ /usr/bin/grep -o 'AC-CSPS-[0-9]*' acceptance.md | sort -u → 001..008, 결번·중복 0
$ /usr/bin/grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/ ; rc=1 (무매치)
$ /usr/bin/grep -c 'syscall' spec.md → 0
$ SPEC-CODEX-SKILL-PATH-001:        status: completed
$ SPEC-CODEX-SKILLCONFIG-SHAPE-001: status: completed
$ /usr/bin/grep -rn 'cannot carry verbatim' --include='*.go' .
  ./internal/cli/codex_skills_disable.go:246   ← 생산 코드 1줄만. F4(테스트 0건) 재확인.
```

---

## 3. Baseline-attribution (무엇에 대고 쟀는가)

- 트리: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t540`, `git branch --show-current` → `WT-codex-path-escape`, `git rev-parse --short HEAD` → `9ce792637`. 이 감사 회차, 이 트리에서 읽었다.
- 모든 부재 주장은 `/usr/bin/grep` 으로 냈다(셸 `grep` 은 ugrep 래퍼라 조용히 건너뛴다).
- `ToSlash`/`FromSlash`/`IsAbs` 실측은 이 호스트(darwin, `GOOS: darwin`)에서 `go run /tmp//t540_pchk.go` 로 이번 회차에 냈다. 다른 트리·다른 시점의 수치를 옮겨 쓴 것이 없다.
- Tier M PASS 임계 0.80 은 `.claude/rules/moai/workflow/spec-workflow.md:141` 표에서 읽었다.

---

## 4. Gaps (관측하지 않은 것)

1. **테스트를 한 건도 돌리지 않았다.** `go test ./internal/cli/...` 미실행(감사 지시의 부하 제약). F-2 의 "RED 를 못 세운다" 는 **소스·문면 판독에 근거한 판단**이며, 실제 테스트 실행으로 반증되지 않았다.
2. **Windows 호스트 미관측.** `filepath.ToSlash` 가 windows 에서 `\`→`/` 로 바꾼다는 것, Windows `os.Stat` 이 forward slash 를 받는다는 것 — 둘 다 Go 계약에 근거한 **가설**이며 이 감사에서 측정하지 않았다. F-1 의 판정은 windows 동작에 의존하지 않는다(darwin 실측만으로 성립).
3. **Codex 의 windows slash 해석 미측정.** SPEC 이 스스로 게이트로 세운 항목이며 이 감사도 그것을 풀지 않았다.
4. **파괴적 prune 오독 미재현.** F-3 은 원본 분기 판독이지 실행 재현이 아니다.
5. **교차 모델 2차 의견 없음** (`audit_model` 미설정 → Claude 단독).
6. **`design.md` / `research.md` 없음** — Tier M 입력 계약상 정상이며 결함이 아니다.

---

## 5. Residual-risk (관측했는데도 틀릴 수 있는 것)

- F-1 을 "저자가 `filepath.ToSlash` 를 느슨하게 이름 붙였을 뿐, 실제로는 명시적 치환을 의도했다" 로 읽을 여지가 있다. 그러나 REQ·plan·AC 세 문서가 모두 `filepath.ToSlash` 를 **명시**하므로, 그 독법을 채택하려면 문면이 바뀌어야 한다 — 감사 판정은 문면 기준이다.
- F-3 은 "그 지점이 가진 힘"(파괴적 삭제가 가능한 유일한 지점)을 서술한 것으로 방어할 수 있다. 다만 REQ-CSPS-005 와 AC-CSPS-004 가 **삭제되지 않음**을 요구 사항으로 세운 이상, 그 요구가 base 에서 이미 만족된다는 사실은 남는다.
- Tier M 회차 상한이 2 이므로, 아래 수정 후 재감사는 1회만 남는다.

---

## Must-Pass Results

| | 결과 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | E-7: REQ-CSPS-001..009, 결번·중복 0, zero-padding 일관 |
| MP-2 GEARS 형식 | **PASS** (요구층 기준) | `spec.md:86-94` 9개 REQ 중 8개가 다섯 패턴에 정합. REQ-CSPS-001 만 이탈(F-7, minor). `acceptance.md` 의 Given-When-Then 은 검증층이므로 여기서 감점하지 않았다 |
| MP-3 YAML frontmatter | **PASS** | `spec.md:1-17` 12개 정본 필드 전부 존재, snake_case alias 0, `status: draft`·`priority: P2`·날짜 ISO 형식 정합 |
| MP-4 언어 중립성 | **N/A** | 단일 언어(Go) 내부 CLI 카드. 16개 프로그래밍 언어 도구를 다루지 않는다 |
| MP-5 D7 교차 SPEC | **PASS** | E-7: 참조 2건 모두 실재, status `completed` — retired/superseded/archived 없음 |
| MP-6 D8 크로스플랫폼 | **PASS(auto)** | E-7: `syscall` 0회 |
| MP-7 clarification gate | **PASS** | E-7: `[NEEDS CLARIFICATION` 무매치 (rc=1) |

**Must-Pass 실패는 없다.** FAIL 은 아래 blocking 결함에서 나온다.

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | F-4·F-5 의 사소한 사실 오차 외에는 해석이 갈리지 않는다 |
| Completeness | 1.00 | 1.0 | HISTORY/§A/§B/§C/§F(Out of Scope H3+불릿)/§G 전부 존재, frontmatter 완전 |
| Testability | 0.50 | 0.50 | 8개 중 4개(002·004·005·007)가 이 카드의 호스트에서 판별력을 잃는다 |
| Traceability | 1.00 | 1.0 | `spec.md:100-110` 매핑표: REQ 9개 전부 AC 보유, 고아 AC 0 |

Aggregate = 0.81 (Tier M 임계 0.80 위) — **그럼에도 FAIL.** M6 에 따라 blocking 결함(정확성·내부 일관성)은 점수로 상쇄되지 않는다.

## Defects Found

- **D1** — F-1 — `spec.md:87`, `plan.md:74`, `acceptance.md:52-57` — REQ-CSPS-002/M2 가 지정한 `filepath.ToSlash` 는 darwin·linux 에서 항등이라 `\` 를 변환하지 못한다(E-1 실측). 그 결과 AC-CSPS-002 의 Then 과 "platform-independent" 주석이 서로 모순이고, AC-CSPS-002·AC-CSPS-007 은 이 카드의 호스트에서 통과 불가다. `§G` gap 1 의 "플랫폼 무관 실측으로 바꾼다" 와 t502 F4 종결 주장도 함께 무효가 된다 — Severity: **critical** — Class: **blocking** — 필요한 수정: 정규화 수단을 플랫폼 무관 형태(예: `\` → `/` 명시적 치환)로 REQ-CSPS-002·plan M2·AC-CSPS-002 세 곳에서 동시에 다시 쓰거나, `filepath.ToSlash` 를 유지하되 AC-CSPS-002/007 의 "platform-independent" 주장을 철회하고 windows 전용 실행 조건을 명시하고 F4 종결 주장을 그에 맞춰 낮춘다. 두 길 중 하나를 고르되, 지금처럼 양쪽을 동시에 주장하는 상태로 두지 말 것.
- **D2** — F-2 — `acceptance.md:77-103`, `plan.md:83-95` — AC-CSPS-004·005 는 카드 호스트에서 base 와 동일한 코드 경로를 밟으므로 M3(`FromSlash`)에 대해 RED 를 세우지 못한다. windows 형태 slash 경로를 픽스처로 쓰는 우회도 `IsAbs=false → codexPathRelative` 로 막힌다(E-1·E-6) — Severity: **major** — Class: **blocking** — 필요한 수정: (a) stat 대상 산출을 순수 함수 seam 으로 분리해 `FromSlash` 적용을 직접 단정하는 AC 를 세우거나, (b) M3 를 "방어적 정규화, 관측 가능한 동작 변화 없음" 으로 명시하고 AC-CSPS-004/005 의 판별 주장을 회귀 고정(regression pin)으로 격하한 뒤 `§D.2` 의 RED 의무에서 M3 를 명시적으로 면제한다. 어느 쪽이든 문서에 적힌 채로 남아야 한다.
- **D3** — F-3 — `spec.md:53`(§B.2 1행), `spec.md:90`(REQ-CSPS-005), `spec.md:181`(§G gap 3) — 삭제는 분류가 absolute/home-relative 이고 `ErrNotExist` 일 때만 일어난다(E-3). slash·backslash 오독이 떨어지는 `codexPathRelative`·`codexPathOddlyFormed` 는 비파괴 skip 이며, gap 3 은 `:102` 만 인용하고 `:76-94` 가드를 빠뜨렸다. REQ-CSPS-005 는 base 에서 이미 만족된다 — Severity: **major** — Class: **blocking** — 필요한 수정: 파괴성 주장을 **이스케이프 형태(옵션 b)의 windows 상 위험**으로 범위 제한하고, `§G` gap 3 에 분기 가드(`:76-94`)를 명시하며, REQ-CSPS-005 를 "base 에서 이미 성립하는 회귀 고정"으로 라벨링한다.
- **D4** — F-4 — `spec.md:131,138`, `plan.md:65-66` — `len(matches) > 1` 은 `:257`(문서 `:256`)이며 비교를 수행하지 않는다. 비교 지점은 `:252` 하나뿐이라 "두 분기가 매치 정의를 놓고 어긋난다"는 위험은 현 코드에서 성립하지 않는다 — Severity: **minor** — Class: **blocking**(구현자가 없는 지점을 찾다가 arm 을 버릴 수 있다) — 필요한 수정: 줄번호를 `:257` 로 고치고, "비교는 `:252` 한 곳이며 거기서 정규화하면 두 분기가 자동으로 일치한다"로 다시 쓴다. AC-CSPS-007 의 duplicate arm 자체는 판별력이 있으므로 **유지**한다.
- **D5** — F-5 — `spec.md:42`(§B.1 2행), `spec.md:137`(§D.4) — "Measured facts" 표에 미관측 windows 런타임 추론이 실려 있고, §D.4 의 "on Windows it still classifies `codexPathAbsolute` … and stats fine" 는 gap 목록에 없다 — Severity: **minor** — Class: **blocking**(VCI §1 무관측 주장) — 필요한 수정: 2행의 뒷절을 추론으로 라벨링하고, §D.4 의 windows 문장을 `§G` gap 으로 올린다.
- **D6** — F-6 — `acceptance.md:106-118` — AC-CSPS-006 의 "기존 스위트 전부 통과" 절반에 대조가 없어 셀렉터 0매치·미컴파일에서 공허한 초록이 된다 — Severity: **minor** — Class: **optional** — 권장 수정: 실행된 테스트 수를 함께 인용한다(예: `--- PASS` 행 수, 뒤에 공백이 오는 형태로 셀 것 — `$` 앵커는 go 의 `(0.06s)` 접미 탓에 상시 0 이 된다).
- **D7** — F-7 — `spec.md:86` — REQ-CSPS-001 이 GEARS 다섯 패턴 밖("Before …, the maintainer shall have measured") — Severity: **minor** — Class: **optional** — 권장 수정: "When a publisher change is proposed for landing, the maintainer shall measure …" 로 다시 쓴다.

## Recommendation

FAIL. 아래 순서로 고친 뒤 재감사(회차 2/2, 남은 회차 1회).

1. **D1 을 먼저 결정한다.** 정규화 수단을 플랫폼 무관 치환으로 바꾸느냐, `filepath.ToSlash` 를 유지하고 AC 의 플랫폼 무관 주장·F4 종결 주장을 철회하느냐. 이 선택이 D2 의 해법과 AC 매트릭스 전체를 결정하므로 다른 어떤 수정보다 앞선다.
2. **D2** — M3 를 "판별 가능한 AC 를 가진 변경"으로 만들지, "동작 무변화 방어적 정규화"로 격하할지 문서에 적는다.
3. **D3** — 파괴성 주장의 범위를 이스케이프 형태로 좁히고 gap 3 에 분기 가드를 넣는다.
4. **D4·D5** — 줄번호와 무관측 표지를 고친다.
5. **D6·D7** 은 오케스트레이터 재량(optional).

**바뀌지 않아야 하는 것** (이번 감사에서 근거를 확인했으니 재작성 중 잃지 말 것): AC-CSPS-001 게이트와 그 fallback 절(양팔 실패 = 하네스 실패이지 반증이 아니라는 구분 포함), AC-CSPS-008 의 재유도 좌변·zero-control="not measurable" 규율, AC-CSPS-003 의 refuse-branch 보존, AC-CSPS-007 의 두 팔, `§D.2` 순서 제약, 읽는 쪽 전수 3건, `§F` 의 out-of-scope 5개 블록, F5 를 흡수하지 않고 별도 정책 카드로 남긴 판단.

---

# SPEC 감사 보고 — 회차 2/2 (SPEC-CODEX-SKILL-PATH-SLASH-001, 카드 t540)

- Iteration: 2/2 (Tier M ceiling — 이 회차 이후 자동 재감사 없음)
- **Verdict: FAIL**
- Overall Score: 0.88 (Tier M 임계 0.80) — 회차 1의 0.81 대비 **상승**, 점수 역행 없음 → STOP 신호 미발화
- 감사 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t540`, 브랜치 `WT-codex-path-escape`, HEAD `9ce792637`
- 저자 추론 맥락은 M1 Context Isolation 에 따라 무시했다 (Reasoning context ignored per M1 Context Isolation).
- 이 회차는 **위임된 5개 수리 + 저자 자진 수정 2건 + 비례성 판정**으로 범위를 좁혔다. 회차 1이 명시적으로 통과시킨 항목(읽는 쪽 전수 3건, `classifyCodexSkillPath` 순서 서술, AC-CSPS-001 게이트·fallback, AC-CSPS-008 의 t543 규율, AC-CSPS-007 두 팔의 판별력, F5 분리 기록)은 **수리가 그것을 깨뜨렸는지만** 확인했고, 깨진 것은 없다.

---

## 1. Claim (무엇을 주장하는가)

1. **R2-1 [major / blocking]** — D2 수리는 **읽는 쪽 두 곳 중 한 곳에서만** 성립한다. `doctor_codex.go:857` 의 stat 은 `osStatFn` 심(seam)이 아니라 **직접 `os.Stat`** 이다. AC-CSPS-004 arm A 는 "`osStatFn` 을 recorder 로 교체 … **BOTH readers** 에서 recorder 가 `\tmp\x\SKILL.md` 를 관측한다"고 쓰고, `plan.md:151-155` 는 `TestCodexStaleSkillFinding_SeparatorConversion` 을 이름 붙이고 "BEFORE the change the recorder observes `/tmp/x/SKILL.md`" 라고 단정한다. **doctor 팔에 대해 그 문장은 거짓이다** — 그 경로에 recorder 가 놓일 자리가 없다. 게다가 `codexStaleSkillFinding()` 은 인자 없는 함수로 실제 `CODEX_HOME` 을 해석해 실제 파일을 읽으며, 기존 테스트가 **0건**이다. 따라서 M3 의 TDD 의무는 prune 쪽에서만 이행되고, doctor 쪽은 RED 를 세울 수 없다.
2. **R2-2 [minor / blocking]** — 파괴성 주장의 windows 한정이 `spec.md:123`(§B.4) 에서 **누락된 채 살아남았다**. "running `prune` before the path-resolution defect is fixed would perform destructive deletion through the defective logic" 는 문서 자신의 I2(파괴 분기는 windows 에서만 열림)와 어긋나며, t533 의 prune 실행이 어느 호스트에서 도는지 이 문서 어디에도 없다. VCI §1.1 surface 4(권고 전제 주장).
3. **R2-3 [minor / blocking]** — `spec.md:35` 는 표지 없는 I1 단정이며, 바로 아래 `:37-39` 의 "every Windows-behaviour statement in this document is an inference and is **marked as one**" 과 정면으로 충돌한다. D5 가 세운 규율이 자기 문단 세 줄 위에서 깨져 있다.
4. **R2-4 [minor / optional]** — AC-CSPS-006 의 실행-테스트 대조가 "base count" 를 참조하지만, `plan.md §G` 에도 `acceptance.md` 에도 **base count 를 산출하는 명령이 없다**. 판정어 "materially below" 도 사람 판단을 요구한다.
5. **R2-5 [minor / optional]** — doctor 소비 지점을 `:825-835` 로 인용하지만(§B.2 2행, `plan.md:133`), 이 카드가 실제로 고쳐야 할 stat 은 `:857` 이다. 하필 누락된 그 줄이 R2-1 의 발원지다.
6. **R2-6 [minor / optional]** — `§C` 가 REQ-CSPS-010 을 REQ-CSPS-009 **앞에** 선언한다. 결번·중복은 없으므로 MP-1 실패가 아니다(표기 문제).
7. **통과로 판정한 것** — D1 완전 수리, D3(R2-2 잔여 제외) 수리, D4 완전 수리, D5(R2-3 잔여 제외) 수리, 자진 수정 2건 모두 **건전(유지)**, 표면 확대는 **비례적**(REQ-CSPS-010 은 load-bearing, AC-CSPS-004 3팔 분할은 복잡도를 벌고 있다).

---

## 2. Evidence (명령과 그 출력 그대로)

### E2-1 — doctor 리더의 stat 은 심을 지나지 않는다 (R2-1)

```
$ /usr/bin/grep -rn 'osStatFn' --include='*.go' internal/cli/ | /usr/bin/grep -v '_test.go'
internal/cli/codex_skills_prune.go:96:	_, err := osStatFn(statPath)
internal/cli/update_preserve_inventory.go:44:// osStatFn is the injectable stat seam used by mergeBackPreserveInventory to
internal/cli/update_preserve_inventory.go:59:var osStatFn = os.Stat
internal/cli/update_preserve_inventory.go:439:		if _, statErr := osStatFn(srcPath); statErr != nil {
internal/cli/codex_skills_disable.go:149:	if st, err := osStatFn(projMirror); err != nil || !st.IsDir() {
internal/cli/codex_skills_disable.go:168:		if st, err := osStatFn(c.file); err == nil && st.Mode().IsRegular() {

$ /usr/bin/grep -n 'os\.Stat\|os\.Lstat' internal/cli/doctor_codex.go
441:		info, lerr := os.Lstat(entryPath)
459:		_, serr := os.Stat(entryPath)
857:		_, serr := os.Stat(statPath)

$ /usr/bin/grep -rn 'codexStaleSkillFinding' --include='*_test.go' internal/cli/
(무출력 — 기존 테스트 0건)

$ /usr/bin/grep -n 'func codexUserSkillConfig' -A14 internal/cli/doctor_codex.go
713:func codexUserSkillConfig() (cfgPath, display string, entries []codexwiring.SkillEntry, ok bool) {
714-	codexHome, source := resolveCodexHomeDir()
...
719-	raw, err := os.ReadFile(cfgPath)
723-	entries = codexwiring.ParseSkillEntries(raw)
```

`osStatFn` 을 소비하는 비테스트 지점은 4곳이고 **`doctor_codex.go` 는 그중에 없다.** `codexStaleSkillFinding()`(`:815`) 은 인자를 받지 않고 `codexUserSkillConfig()` → `resolveCodexHomeDir()` → `os.ReadFile` 로 실제 config 를 읽은 뒤 `:857` 에서 실제 `os.Stat` 을 부른다.

이에 대응하는 SPEC 문면:

- `acceptance.md:94-99` — "**Given** … `osStatFn` replaced by a recorder … **Then** the recorder observes `\tmp\x\SKILL.md` … from **BOTH** readers."
- `plan.md:147` — "Fixture: `configPathSeparator = '\\'`; `osStatFn` replaced by a recorder capturing its argument"
- `plan.md:152,154` — "`go test … -run 'TestCodexStaleSkillFinding_SeparatorConversion'` … BEFORE the change the recorder observes `/tmp/x/SKILL.md` → assertion fails (RED)"

**prune 팔은 성립한다.** `codex_skills_prune.go:76-78` 이 `classifyCodexSkillPath("/tmp/x/SKILL.md")` → `codexPathAbsolute` → `statPath = e.Path` 로 두고 `:96` 이 `osStatFn(statPath)` 를 부르므로, `/tmp/x/SKILL.md` 픽스처는 실제로 stat 호출에 도달하고 recorder 가 인자를 잡는다. 변경 전 `/tmp/x/SKILL.md`, 변경 후 `\tmp\x\SKILL.md` → **진짜 RED.** M6 의 `IsAbs(slash form)=false` 때문에 `C:/…` 픽스처를 쓸 수 없다는 저자의 서술도 맞다(E2-2).

### E2-2 — M6 재측정 (이번 회차, 이 트리, 이 호스트)

```
$ go run .moai/reports/t540/lab/ts.go
GOOS=darwin Separator='/'
ToSlash(backslash)  = "C:\\Users\\u\\.agents\\skills\\foo\\SKILL.md" changed=false
FromSlash(slash)    = "C:/Users/u/.agents/skills/foo/SKILL.md" changed=false
IsAbs(slash form)   = false
IsAbs(backslash)    = false
```

`spec.md:50` 의 M6 행과 **바이트 대응**한다. 기억에서 옮겨 적은 수치가 아니다.

### E2-3 — MEASURED 표 6행 전수 대조 (D5)

```
$ git status --porcelain | /usr/bin/grep -E '\.go$' ; echo "rc=$?"
rc=1        ← Go 파일 수정 0건. 워킹트리 Go 원본 == 9ce792637
```

| 행 | 주소 | 대조 결과 |
|---|---|---|
| M1 | `codex_skills_disable.go:245` | `if strings.ContainsAny(skillPath, "\"\\\n\r") {` — 일치 |
| M2 | `codex_skills_disable.go:148,158,160` | `filepath.Join(projectRoot, …)` / `filepath.Join(homeDir, …)` / `filepath.Join(homeMirror, skill, "SKILL.md")` — 일치 |
| M3 | `codexwiring/skills.go:174` | `skillPathKeyRe = regexp.MustCompile("^path\\s*=\\s*\"([^\"]*)\"\\s*(#.*)?$")` — 일치 |
| M4 | `codex_skills_disable.go:252` | `if e.Path == skillPath {` — 일치. `:245-259` 전체에서 유일한 경로 비교 |
| M5 | `codex_skills_prune.go:76-94,102` | switch 4분기 + `case errors.Is(err, fs.ErrNotExist): return …{Eligible: true}` — 일치 |
| M6 | 이 호스트의 probe | E2-2 — 일치 |

I2 가 인용한 `doctor_codex.go:668-677` 도 확인: `:668 filepath.IsAbs` → `:674 strings.ContainsRune(p, '\\')` → `:677 return codexPathRelative`. **6행 전부 이 트리에서 판독 가능하다.**

### E2-4 — I1 / I2 하류 사용처 전수 (R2-2, R2-3)

```
$ /usr/bin/grep -ni 'destructive\|delete\|deletion' spec.md plan.md acceptance.md progress.md
spec.md:29   … the disagreement that would let the `prune` verb delete a healthy registration.   ← 무한정 (사용자 스토리)
spec.md:49   … `codexPathRelative` and `codexPathOddlyFormed` are non-destructive skips          ← MEASURED
spec.md:57   … so the destructive prune branch opens **on Windows only**                          ← I2 표지
spec.md:59   … so on Windows that divergence is destructive rather than cosmetic (I2)             ← 한정 + 표지
spec.md:67   … destructive **on Windows only** (I2)                                               ← 한정 + 표지
spec.md:81   The destructive `prune` misread path (I2) stays closed by construction               ← 표지
spec.md:123  … running `prune` before the path-resolution defect is fixed would perform
             destructive deletion through the defective logic.                                    ← 무한정 (R2-2)
spec.md:138  On a host whose separator is not `/`, the `prune` verb shall not … (M5) … (I2)        ← 한정 + 표지
spec.md:201  INFERRED (I2, not observed) …                                                        ← 표지
spec.md:250  … the destructive path opens on Windows only (I2). "Destructive on every platform"
             is false …  (`codex_skills_prune.go:76-94` + `:102`)                                 ← 한정 + 두 조건 게이트
```

`§B.2` 1행 · `§B.1` 산문 · `§G` gap 3 · REQ-CSPS-005 — **위임된 네 지점은 모두 두 조건 게이트(`:76-94` + `:102`)와 windows 한정을 달고 있다.** 남은 무한정 서술은 `:29`(동기 산문, 거짓은 아님)와 `:123`(**전제를 실어 나르는 판단문 — R2-2**) 둘뿐이다.

`spec.md:35` 원문:

```
35:The publisher refuses any path containing `"`, `\`, LF, or CR. Every Windows path contains `\`.
   The verb is therefore inoperative on Windows, and reports rc=0 while doing nothing.
...
38:Windows-behaviour statement in this document is an inference and is marked as one.
```

두 번째·세 번째 문장이 정확히 I1 인데 표지가 없고, 세 줄 뒤 문장이 "전부 표시돼 있다"고 주장한다(R2-3).

### E2-5 — 정규화 수단: 맨 stdlib 처방이 남아 있는가 (D1)

```
$ /usr/bin/grep -n 'ToSlash\|FromSlash' spec.md plan.md acceptance.md progress.md
spec.md:50    (M6 측정 행)
spec.md:84    §B.3.1 제목 — "not `filepath.ToSlash` directly"
spec.md:86    "…are therefore the **identity** here — measured, not assumed (M6)"
spec.md:252   §G gap 5 — 측정 범위 서술
spec.md:256   교차참조
plan.md:16,23,71   근거 서술 (identity 이므로 심이 필요하다)
plan.md:223   §H 리스크 표 — "Seam skipped, `filepath.ToSlash` used directly" (안티패턴)
acceptance.md:58,102   "the un-seamed `filepath.FromSlash` formulation could not produce" (안티패턴)
```

**REQ 도, AC 의 Then 도, plan 의 구현 지시도 맨 `filepath.ToSlash`/`FromSlash` 를 처방하지 않는다.** 전부 서술·안티패턴 문맥이다. REQ-CSPS-002/004/006 은 `toConfigPath(p, sep)` / `fromConfigPath(p, sep)` 를 명시한다(`spec.md:135,137,139`).

`[HARD]` 반-`ReplaceAll` 절이 산문에서 요구로 승격됐는지:

```
spec.md:142  - **REQ-CSPS-010** (Unwanted) — The conversion shall not be an unconditional backslash
             replacement: while the host separator is `/`, a path containing `\` shall reach the
             `:245` guard unchanged and be refused.
spec.md:160  | REQ-CSPS-010 | AC-CSPS-003 |
acceptance.md:79-85  Second arm (REQ-CSPS-010) … "A mutation replacing `toConfigPath` with the
             unconditional form must fail this arm."
plan.md:190  - **The unconditional-replacement mutant is mandatory** … must FAIL AC-CSPS-003 arm 2.
acceptance.md:237  DoD — the unconditional-`ReplaceAll` mutant (must fail AC-CSPS-003 arm 2)
```

**요구 + 테스트 팔 + 뮤턴트 + DoD 항목 — 네 곳 전부 존재한다.** 산문에 머물러 있지 않다.

### E2-6 — `:257` / `:252` 대조와 철회 문단 (D4)

```
$ /usr/bin/grep -n "" internal/cli/codex_skills_disable.go | sed -n '249,259p'
249:	lines, term := codexwiring.SplitConfigLines(content)
250:	var matches []codexwiring.SkillEntry
251:	for _, e := range codexwiring.ParseSkillEntries(content) {
252:		if e.Path == skillPath {
253:			matches = append(matches, e)
254:		}
255:	}
257:	if len(matches) > 1 {

$ /usr/bin/grep -n ':256\b' .moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/*.md ; echo "rc=$?"
rc=1        ← 낡은 :256 인용 0건
```

철회 문단은 `spec.md:191-196` 에 실재한다("**Retracted clause.** … `:257` performs **no comparison** — it counts the slice the single comparison at `:252` filled. … The retraction is recorded rather than silently deleted"). AC-CSPS-007 둘째 팔의 **새 근거**(`spec.md:202`, `acceptance.md:191-196`)는 "정규화가 `matches` 에 들어가는 항목 집합을 바꾸므로 중복 분기의 입력 분포가 바뀐다" 이며, 원본 `:250-258` 에서 재유도된다: 백슬래시 항목 1 + 슬래시 항목 1 인 config 에서 변경 **전** `len(matches)==1`(update), 변경 **후** `len(matches)==2`(skip). **동작이 실제로 바뀌므로 팔은 판별력을 가진다.** 철회된 전제("두 비교 지점이 어긋난다")의 재포장이 아니다.

### E2-7 — Must-Pass 기계 검증 (재실행)

```
$ /usr/bin/grep -o 'REQ-CSPS-[0-9]*' spec.md | sort -u   → 001..010, 결번 0 · 중복 0
   (단, §C 선언 순서는 001..008, 010, 009 — R2-6)
$ /usr/bin/grep -o 'AC-CSPS-[0-9]*' acceptance.md | sort -u → 001..008, 결번 0 · 중복 0
$ /usr/bin/grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/ ; rc=1 (무매치)
$ /usr/bin/grep -c 'syscall' spec.md → 0
$ SPEC-CODEX-SKILL-PATH-001: completed / SPEC-CODEX-SKILLCONFIG-SHAPE-001: completed
```

---

## 3. Baseline-attribution (무엇에 대고 쟀는가)

- 트리: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t540`; `git branch --show-current` → `WT-codex-path-escape`; `git rev-parse --short HEAD` → `9ce792637`. 이 회차, 이 트리에서 읽었다.
- Go 원본 판독은 워킹트리 파일에 대해 냈고, `git status --porcelain | /usr/bin/grep -E '\.go$'` 가 rc=1(무출력)이므로 **워킹트리 Go 원본 == `9ce792637`** 이다(E2-3). MEASURED 표의 "read at `9ce792637`" 주장은 이 회차에 재확인됐다.
- 모든 부재 주장은 `/usr/bin/grep` 으로 냈다(셸 `grep` 은 ugrep 래퍼라 조용히 건너뛴다).
- `ToSlash`/`FromSlash`/`IsAbs` 수치는 `go run .moai/reports/t540/lab/ts.go` 를 **이번 회차에 재실행**해 얻었다(E2-2). 회차 1의 수치를 옮겨 쓰지 않았다.
- Tier M PASS 임계 0.80 은 `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier 표에서 읽었다.
- 이 파일 상단의 회차 1 결함 D1~D7 이 이 회차의 델타 범위다.

---

## 4. Gaps (관측하지 않은 것)

1. **테스트를 한 건도 돌리지 않았다.** 감사 지시의 부하 제약. R2-1 은 **원본 판독**(심의 부재, `codexStaleSkillFinding` 의 시그니처와 실제 `os.Stat`)에 근거한 판단이며, 실제 테스트 작성·실행으로 반증되지 않았다.
2. **"AC-CSPS-002 가 이제 darwin 에서 통과한다"를 실행으로 확인하지 않았다.** `toConfigPath(p,'\\')` 가 백슬래시를 `/` 로 바꾸고 `:245` 의 문자 집합(`"`, `\`, LF, CR)에 걸리지 않는다는 것은 `:245` 를 직접 읽고 명세된 순수 함수의 계약과 합성해 낸 **원본 유도**다. 프로브를 새로 작성하지는 않았다.
3. **Windows 호스트 미관측.** I1·I2 는 이 감사에서도 측정되지 않았다. 이 회차의 어떤 판정도 windows 동작에 의존하지 않는다.
4. **t533 의 49개 유령 항목의 경로 형태를 재지 않았다.** 그래서 R2-2 는 "§B.4 의 순서 제약이 틀렸다"가 아니라 "§B.4 가 든 **전제**가 문서 자신의 I2 와 어긋난다"까지만 주장한다. 순서 제약 자체는 다른 근거로 옳을 수 있다.
5. **파괴적 prune 오독 미재현.** 회차 1과 동일.
6. **교차 모델 2차 의견 없음** — `.moai/config/sections/` 에 `audit_model` 키가 없어 Claude 단독 감사다. `mcp__moai__audit_multi` 는 호출하지 않았다.
7. **`design.md` / `research.md` 없음** — Tier M 입력 계약(spec/plan/acceptance)상 정상이며 결함이 아니다.

---

## 5. Residual-risk (관측했는데도 틀릴 수 있는 것)

- **R2-1 을 "arm A 의 doctor 절반은 recorder 가 아니라 다른 관측 수단을 쓰라는 뜻"으로 읽을 여지**가 있다. 그러나 `acceptance.md:94-99` 와 `plan.md:147,154` 가 recorder 를 명시하고 "BOTH readers" 로 묶었으므로, 그 독법을 채택하려면 문면이 바뀌어야 한다 — 그것이 곧 요구되는 수정이다.
- **수리 (a) 를 고르면 새 위험이 하나 붙는다.** `doctor_codex.go:857` 을 `osStatFn` 로 돌리면 `update_preserve_inventory.go` 의 기존 `osStatFn` 오버라이드 테스트와 doctor 가 같은 패키지 변수를 공유하게 된다. `prune` 이 이미 같은 결합을 갖고 있으므로 신종 위험은 아니지만, 비-병렬 + `t.Cleanup` 규율이 doctor 테스트에도 확장된다. AC-CSPS-006(base 와 바이트 동일)은 `var osStatFn = os.Stat` 이 기본값이므로 깨지지 않는다.
- **R2-2 는 "t533 의 항목이 절대 unix 경로라 darwin 에서도 파괴적"이라는 별도 근거로 방어될 수 있다.** 그 근거는 이 문서에도 이 감사에도 없다(Gap 4). 방어하려면 측정해서 적어야 한다.
- **AC-CSPS-004 arm A 의 `/tmp/x/SKILL.md` 픽스처는 windows 생산 경로와 형태가 다르다**(생산에서는 `C:/…` 선언이 windows `IsAbs` 로 absolute 가 된다). darwin 에서 배선을 재는 대리 픽스처로서는 타당하나, 분류 절반의 windows 동작은 여전히 I2 다 — 문서가 §G gap 4 로 이미 인정한다.

---

## Must-Pass Results (회차 2)

| | 결과 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | E2-7: REQ-CSPS-001..010, 결번 0·중복 0·zero-padding 일관. §C 선언 순서만 010↔009 역전(R2-6, 결번·중복 아님) |
| MP-2 GEARS 형식 (요구층) | **PASS** | `spec.md:134-143` 10개 REQ 전부 다섯 패턴 정합. 회차 1의 이탈(REQ-CSPS-001)은 event-driven 으로 재작성돼 해소. `acceptance.md` 의 Given-When-Then 은 **검증층**이므로 이 항목에서 감점하지 않았다 |
| MP-3 YAML frontmatter | **PASS** | `spec.md:1-17` 정본 12필드 전부 존재, snake_case alias 0 |
| MP-4 언어 중립성 | **N/A** | 단일 언어(Go) 내부 CLI 카드 — 16개 프로그래밍 언어 도구를 다루지 않는다 |
| MP-5 D7 교차 SPEC | **PASS** | E2-7: 참조 2건 실재, 둘 다 `completed` |
| MP-6 D8 크로스플랫폼 | **PASS(auto)** | E2-7: `syscall` 0회 |
| MP-7 clarification gate | **PASS** | E2-7: `[NEEDS CLARIFICATION` 무매치 (rc=1) |

**Must-Pass 실패는 없다.** FAIL 은 아래 blocking 결함 R2-1 에서 나온다.

## Category Scores (회차 2)

| 차원 | 회차1 | 회차2 | 밴드 | 근거 |
|---|---|---|---|---|
| Clarity | 0.75 | 0.75 | 0.75 | MEASURED/INFERRED 분리와 철회 기록으로 크게 좋아졌으나 R2-2·R2-3·R2-5 가 남았다 |
| Completeness | 1.00 | 1.00 | 1.0 | HISTORY/§A~§H, `### Out of Scope — <topic>` 6블록 전부 불릿 보유, frontmatter 완전 |
| Testability | 0.50 | 0.75 | 0.75 | AC-002·003(2팔)·004 arm A(prune)·arm B·007(2팔)·008 이 이 호스트에서 판별력을 얻었다. 남은 하나 — 004 arm A 의 doctor 절반 — 이 실행 불가(R2-1) |
| Traceability | 1.00 | 1.00 | 1.0 | `spec.md:149-160` 매핑표: REQ 10개 전부 AC 보유, `acceptance.md` 8개 AC 전부 유효 REQ 참조, 고아 0 |

Aggregate = **0.88** (Tier M 임계 0.80 초과, 회차 1 대비 +0.07 — 점수 역행 없음).
**그럼에도 FAIL.** M6 에 따라 blocking 결함(정확성·내부 일관성)은 점수로 상쇄되지 않는다. 회차 1이 0.81 에서 같은 논리로 FAIL 했으므로 판정 일관성상 R2-1 도 같은 취급을 받는다.

## Defects Found (회차 2)

- **R2-1** — `acceptance.md:94-99`, `plan.md:147,151-156` — `doctor_codex.go:857` 은 `osStatFn` 이 아니라 직접 `os.Stat` 이다(E2-1). arm A 의 "recorder 가 **BOTH readers** 에서 관측한다" 와 plan 의 "BEFORE the change the recorder observes `/tmp/x/SKILL.md`" 는 doctor 팔에 대해 **성립 불가**다. `codexStaleSkillFinding()` 은 인자 없이 실제 `CODEX_HOME` config 를 읽고 기존 테스트가 0건이라 대체 관측 경로도 확립돼 있지 않다 — Severity: **major** — Class: **blocking** — 필요한 수정(**둘 중 하나를 택해 문면에 남길 것**):
  **(a)** `doctor_codex.go:857` 의 stat 을 `osStatFn` 심으로 돌린다. 이 경우 그것은 이 카드가 아직 허가하지 않은 생산 코드 변경이므로 **REQ 로 명시**하고 `plan.md` M3 에 단계로 넣고 `§F` 의 "path-classification refactor 는 out of scope" 와의 경계를 진술한다. `var osStatFn = os.Stat` 가 기본값이므로 AC-CSPS-006 의 바이트 동일 주장은 유지된다. 부수 위험은 §5 잔여 위험에 적힌 대로 `update_preserve_inventory` 테스트와의 패키지 변수 공유다.
  **(b)** arm A 의 RED 주장을 `judgeCodexSkillEntry` **한 곳으로 한정**하고, doctor 절반은 회귀 가드로 격하해 `§G` 에 gap 으로 올린다("reader 2 의 변환은 이 호스트에서 인자 수준으로 관측할 수 없다 — stat 이 심을 지나지 않는다"). `plan.md` M3 의 RED 블록에서 `TestCodexStaleSkillFinding_SeparatorConversion` 의 RED 주장을 삭제하고, DoD 의 "RED source named per AC" 항목에 그 사실을 반영한다.
  지금 문면(양쪽을 동시에 주장)은 어느 쪽으로도 두지 말 것 — 구현자가 없는 심을 찾다가 **말없이 추가**하거나(허가되지 않은 범위) **RED 없이 GREEN 을 보고**하게 된다.
- **R2-2** — `spec.md:123` — §B.4 의 순서 근거가 파괴성의 windows 한정을 떨어뜨린다(E2-4). 문서 자신의 I2 와 어긋나고, t533 의 prune 실행 호스트는 어디에도 없다 — Severity: **minor** — Class: **blocking**(VCI §1.1 surface 4, 권고 전제) — 필요한 수정: 한 문장. "on a Windows host that deletion is destructive (I2); on a `/`-separator host the ordering holds for <실제 근거>" 로 한정하거나, t533 항목의 경로 형태를 재서 근거를 바꾼다.
- **R2-3** — `spec.md:35` — 표지 없는 I1 단정이 `:37-39` 의 "every Windows-behaviour statement … is marked as one" 을 세 줄 만에 반증한다(E2-4) — Severity: **minor** — Class: **blocking**(D5 가 세운 규율의 자기 위반) — 필요한 수정: `:35` 의 2·3번째 문장에 `(I1)` 표지를 달거나 그 문장들을 표 아래로 옮긴다.
- **R2-4** — `acceptance.md:150-162`, `plan.md:168-171,196-200` — 실행-테스트 대조가 "base count" 를 참조하지만 그것을 산출하는 명령이 어느 문서에도 없고, "materially below" 는 판정을 사람 판단에 맡긴다 — Severity: **minor** — Class: **optional** — 권장 수정: `§G` 에 base 측정 단계를 넣고(M1 착지 **전** 같은 명령으로 1회) 판정식을 구체화한다(예: `after >= before && after > 0`, 아니면 그냥 "not measurable").
- **R2-5** — `spec.md:68`(§B.2 2행), `plan.md:133` — doctor 소비 지점을 `:825-835` 로 인용하지만 이 카드가 고쳐야 할 stat 은 `:857` 이다(E2-1) — Severity: **minor** — Class: **optional** — 권장 수정: 범위를 `:824-857`(분류 `:831-856` + stat `:857`)로 고친다. R2-1 을 (b) 로 닫더라도 이 줄번호는 고쳐 두는 편이 낫다.
- **R2-6** — `spec.md:142-143` — §C 가 REQ-CSPS-010 을 009 앞에 선언한다. 결번·중복 없음이므로 MP-1 실패가 아니다 — Severity: **minor** — Class: **optional** — 권장 수정: 두 줄 자리바꿈.

## Regression Check (회차 1 결함의 해소 여부)

| 회차1 | 상태 | 근거 |
|---|---|---|
| **D1** (`filepath.ToSlash` 가 darwin 에서 항등 → AC-002/007 통과 불가) | **RESOLVED** | `toConfigPath(p, sep)`/`fromConfigPath(p, sep)` + `var configPathSeparator = filepath.Separator` 심이 `spec.md §B.3.1`·REQ-CSPS-002/004/006·`plan.md` M1 에 들어갔다. 심은 순수 문자열 함수라 GOOS 에 의존하지 않으므로, `sep='\\'` 로 오버라이드한 테스트는 darwin 에서 `C:\…` → `C:/…` 를 만들고 `:245`(E2-3 M1 의 문자 집합)를 통과한다 — AC-CSPS-002 의 Then 이 이 호스트에서 성립한다. "platform-independent" 주장은 네 곳 전부에서 **심을 전제로** 조건화됐다(E2-5). 반-`ReplaceAll` 절은 REQ-CSPS-010 + AC-CSPS-003 arm 2 + 필수 뮤턴트 + DoD 로 승격됐다. 맨 stdlib 처방은 REQ·AC·plan 지시 어디에도 남아 있지 않다 |
| **D2** (M3 가 RED 를 못 세운다) | **PARTIALLY RESOLVED → R2-1** | prune 팔은 해소(`/tmp/x/SKILL.md` 픽스처가 `IsAbs`→`codexPathAbsolute`→`osStatFn` 에 실제로 도달, E2-1). 회귀 가드 3개(004-C, 005, 006)는 `acceptance.md:120-121,135` 와 DoD `:230-234` 에서 **명시적으로 RED 아님**으로 라벨링돼 TDD 근거로 계상될 수 없다 — 이 부분은 요구대로다. doctor 팔은 미해소 |
| **D3** (파괴성 전제 과장) | **RESOLVED (R2-2 잔여)** | §B.2 1행·§B.1 산문·§G gap 3·REQ-CSPS-005 네 지점 전부 두 조건 게이트(`:76-94` + `:102`)와 windows 한정 보유(E2-4). `spec.md:123` 만 남았다 |
| **D4** (`:256` 오기 + 존재하지 않는 두 번째 비교) | **RESOLVED** | `:256` 인용 0건, `:257`/`:252` 원본 대조 일치, 철회 문단 `§D.4:191-196` 실재, arm 2 의 새 근거(`matches` 입력 분포 변화)는 원본에서 재유도되며 철회 전제의 재포장이 아니다(E2-6) |
| **D5** (MEASURED 표에 미관측 추론) | **RESOLVED (R2-3 잔여)** | MEASURED 6행 전수 대조 통과, INFERRED I1/I2 분리, 하류 사용처 전부 표지 보유. `spec.md:35` 만 표지 누락 |
| **D6** (AC-006 대조 부재, optional) | **RESOLVED — 자진 수정 건전, 유지 권고** | `acceptance.md:150-165`·`plan.md:196-200` 이 `--- PASS: ` 계수를 도입하고 `[HARD]` 로 **뒤 공백 형태**를 못 박았다("`$` 앵커는 go 의 ` (0.06s)` 접미 탓에 상시 0"). 0 또는 base 미만은 "not measurable" 로 보고하도록 규정돼 있고 `§I` 안티패턴에도 반영됐다. **되돌리지 말 것.** 남은 정련은 R2-4(optional) 하나 |
| **D7** (REQ-CSPS-001 GEARS 이탈, optional) | **RESOLVED — 자진 수정 건전, 유지 권고** | `spec.md:134` "When a publisher change is proposed for landing, the maintainer shall measure … and shall record …, before the change lands." — canonical event-driven `When [trigger], the <subject> shall [response]`. `<subject>` 가 `maintainer` 인 것은 GEARS 의 일반화된 subject 허용 범위 안이다. **되돌리지 말 것** |

## 비례성 판정 (표면 확대 10 REQ / 8 AC, Tier M 상한 16/16)

- **REQ-CSPS-010 은 load-bearing 이다 — 유지.** "있으면 좋은" 요구가 아니라, 이 카드에서 가장 그럴듯한 오구현 하나(무조건 `strings.ReplaceAll(p, "\\", "/")`)를 금지한다. 그 오구현의 결과는 **지금의 결함보다 나쁘다** — `\` 를 담은 정당한 unix 파일명이 오늘은 `:245` 에서 **거부**되는데, 무조건 치환은 그것을 **틀린 경로의 발행**으로 바꾼다. 요구는 판별 가능한 AC 팔(AC-CSPS-003 arm 2)과 전용 뮤턴트를 갖고 있고 `§H` 리스크 표에 행을 하나 차지한다. 요구 없이 산문으로만 두면 구현자가 지키지 않아도 아무것도 실패하지 않는다 — 승격이 옳다.
- **AC-CSPS-004 의 3팔 분할은 복잡도를 벌고 있다 — 유지.** 세 팔의 실패 모드가 서로 다르고 각자 대응 뮤턴트가 다르다: **A** 는 배선(리더가 변환을 호출하는가 — identity-stub 뮤턴트가 여기서 죽는다), **B** 는 순수 함수의 계약(`fromConfigPath("C:/…", '\\')`), **C** 는 종단 판정 의미(`Eligible` false/true) — 그리고 C 의 **음성 팔**은 "모든 항목이 resolve 된다"로 만드는 뮤턴트를 잡는 유일한 자리다(`plan.md:186`). 세 팔을 AC 세 개로 쪼개면 REQ→AC 매핑만 부풀고 판별력은 그대로다. 한 AC 안의 3팔이 더 싸다. (R2-1 은 arm A 를 **좁히지**, 없애지 않는다.)
- **총량**: REQ 10/16, AC 8/16 — Tier M 상한 대비 여유가 있고, 각 REQ 가 최소 1개 AC 로 덮이며 고아 AC 가 0 이다(E2-7, `spec.md:149-160`). **과설계 소견 없음.** 회차 1 대비 늘어난 REQ 1개(010)·AC 0개는 감사가 요구한 수리에서 직접 파생된 것이다.

## Recommendation

**FAIL.** 반드시 바뀌어야 하는 것은 **정확히 하나**다.

1. **R2-1** — AC-CSPS-004 arm A 와 `plan.md` M3 의 RED 블록에서 **doctor 리더에 대한 recorder RED 주장을 해소**한다. (a) `doctor_codex.go:857` 을 `osStatFn` 심으로 돌리고 그것을 REQ·plan·§F 경계에 명시하거나, (b) arm A 의 RED 를 `judgeCodexSkillEntry` 로 한정하고 doctor 절반을 회귀 가드로 격하해 `§G` gap 으로 올린다. **(b) 는 순수 문서 편집이며 새 범위를 요구하지 않는다.**
2. **R2-2 · R2-3** — 각각 한 문장 수정(한정어 추가 / `(I1)` 표지 추가).
3. **R2-4 · R2-5 · R2-6** — 오케스트레이터 재량(optional).

**바뀌지 않아야 하는 것**(이번 회차에 근거를 확인했다): 분리자 심(`toConfigPath`/`fromConfigPath` + `configPathSeparator`)과 그 위에 조건화된 "platform-independent" 주장, REQ-CSPS-010 과 그 뮤턴트, AC-CSPS-004 의 3팔 구조와 arm C 의 음성 팔, AC-CSPS-006 의 `--- PASS: ` 대조와 뒤-공백 `[HARD]` 절, REQ-CSPS-001 의 event-driven 재작성, `§D.4` 의 철회 문단과 AC-CSPS-007 둘째 팔의 새 근거, MEASURED/INFERRED 2표 분리, DoD 의 "RED source named per AC" 와 회귀 가드 3개의 명시적 비-RED 라벨링, 그리고 회차 1이 통과시킨 항목 전부.

**회차 상한 도달.** Tier M ceiling 2/2 를 소진했으므로 자동 재감사는 없다. 오케스트레이터는 운영자에게 세 선택지를 제시한다 — (1) **PASS-with-debt**: R2-1 을 (b) 로 닫는 문서 편집만 적용하고 잔여 R2-2/R2-3 을 부채로 기록한 뒤 run 으로 진행, (2) **범위 축소**, (3) **명시적 상한 연장(iter3)**. 이 감사자의 소견으로는 (1) 이 합당하다: 남은 blocking 3건은 전부 **문면 수정**이고 새 측정이나 새 설계를 요구하지 않으며, 점수는 상승(0.81 → 0.88)했고 must-pass 실패는 없다. 다만 그 편집이 **실제로 들어간 것을 읽고 확인**한 뒤에 kickoff 게이트를 열 것 — 편집 주장은 편집이 아니다.
