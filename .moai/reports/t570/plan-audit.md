# SPEC Review Report: SPEC-CODEX-DOCTOR-PATH-GUARD-001 (card t570)

Iteration: **1/1** (Tier S ceiling = 1)
Verdict: **PASS** (Tier S threshold 0.75)
Overall Score: **0.87**

Auditor: plan-auditor. 측정 트리: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t570`,
branch `WT-codex-doctor-guard`, HEAD `a4855f0b2`. 감사 창 안에서 이 트리에 커밋·push·SPEC 편집 없음
(이 판정 파일 1개만 신규 작성). 지시문에 카드 맥락이 포함돼 있었으나 저자 추론 컨텍스트는 아니므로
M1 Context Isolation 위반 없음 — 지시문이 이미 확인했다고 밝힌 4건은 **전부 소스에서 독립 재유도**했고
아래 Evidence 에 명령과 출력을 붙였다.

MCP 부재 고지: 이 세션에서 `moai` MCP 서버가 연결에 실패해 `mcp__moai__spec_audit` /
`mcp__moai__audit_multi` 를 쓰지 못했다. 이것은 SPEC 의 결함이 아니라 **내 증거 커버리지의 간극**이며
Gaps §G-1 에 기록한다. 판정은 Read/Grep/Bash 실측만으로 내렸다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-CDPG-001…007` 연속, 결번·중복·패딩 불일치 없음.
  증거: `/usr/bin/grep -n "^### REQ-" spec.md` → `119,125,133,139,146,154,159` 7행, 001→007.
- **[PASS] MP-2 GEARS 형식** (판정 레이어: `spec.md` §C 의 REQ-XXX 요구 레이어만. `acceptance.md` 의
  Given-When-Then 은 검증 레이어의 정상 형식이므로 이 기준에서 제외했다) — 001/002/003 은
  `While … , when … , the test suite shall …` 복합형(state+event), 004 는 event-driven, 005 는
  event+unwanted, 006 은 state+unwanted, 007 첫 문장은 부정 ubiquitous. 비격식 언어 없음.
  단 007 둘째 문장 `Should a production change prove necessary, …` 은 다섯 패턴 어느 것도 아닌
  `Should` 조건절이다 — legacy If/then 호환 창(2026-11-22) 안이라 MINOR 로 처리(D6), FAIL 아님.
- **[PASS] MP-3 YAML frontmatter** — 12 정식 필드 전부 존재, 타입 정상 (spec.md L1-17):
  `id` / `title` / `version "0.1.0"` / `status draft` / `created 2026-09-08` / `updated 2026-09-08` /
  `author` / `priority P2` / `phase "v3.2.0 target"` / `module internal/cli` /
  `lifecycle spec-anchored` / `tags` CSV 문자열. 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`)
  없음. `tier: S` · `depends_on` · `related_specs` 는 선택 필드로, 형제
  `SPEC-CODEX-SKILL-PATH-READBACK-001/spec.md` L1-17 과 동일한 형태다.
- **[N/A] MP-4 언어 중립성** — 단일 프로그래밍 언어 범위(`module: internal/cli`, Go 단일 패키지).
  자동 통과.
- **[PASS] MP-5 D7 교차 SPEC** — 참조 3건 전부 실재하며 retired/superseded/archived 아님:
  `SPEC-CODEX-SKILL-PATH-READBACK-001 → status: completed`,
  `SPEC-CODEX-SKILL-PATH-SLASH-001 → status: implemented`,
  `SPEC-CODEX-STALE-SPLIT-FOURTH-001 → status: completed`. BLOCKING 소견 없음.
- **[PASS] MP-6 D8 크로스플랫폼** — `/usr/bin/grep -c syscall spec.md plan.md acceptance.md` → `0 0 0`.
  `syscall` 미언급이므로 D8 자동 통과.
- **[PASS] MP-7 clarification gate** — `/usr/bin/grep -rn 'NEEDS CLARIFICATION' plan.md spec.md acceptance.md`
  → rc=1, 매치 0. (`research.md` 는 Tier S 라 부재 — 해당 파일 축은 N/A.)

---

## Category Scores

| Dimension | Score | Rubric band | 근거 |
|---|---|---|---|
| Clarity | 0.90 | 0.75–1.00 사이 | 요구 7건 모두 단일 해석. 감점: AC-CDPG-001 이 픽스처 경로 문자열을 고정하지 않음(D7), §B.1 B-2 셀의 뮤턴트 좌표 부정확(D2), §B.2 의 출처 귀속 오류(D3) |
| Completeness | 0.95 | 1.00 근접 | HISTORY/User Story/Context/Requirements/Exclusions/Assumptions/Cross-Refs 전부 존재. `### Out of Scope —` H3 4개 각각 구체 불릿 보유(spec.md L166,173,178,184). frontmatter 12/12 |
| Testability | 0.90 | 0.75–1.00 사이 | AC 4건 전부 이진 판정 가능, weasel word 없음. 판별력은 아래 Evidence 에서 소스 대조로 확인. 감점: DoD 5번 항목(`statRecorder` grep)이 기대값을 못박지 않음, AC-001 픽스처 미고정 |
| Traceability | 0.75 | 0.75 | AC 4건 전부 실재 REQ 를 참조(001→001, 002→002, 003→003, 004→004). 그러나 REQ-005/006/007 에 대응 AC-XXX 가 **없다** — DoD 체크박스로만 덮인다(D4) |

가중 없는 평균 = (0.90+0.95+0.90+0.75)/4 = **0.875 → 0.87**. Tier S 임계 0.75 초과 → PASS.

---

## Evidence (명령 + 관측 출력)

### E-1 대상 라인과 조기 반환 — `internal/cli/doctor_codex.go`

`sed -n '790,975p' internal/cli/doctor_codex.go` 실측:

- `:837` `statPath = fromConfigPath(e.Path, configPathSeparator)` — SPEC §B.1 의 라인 힌트와 일치.
- `:847-853` `case codexPathRelative:` → `relativeCount++; continue`
- `:854-859` `default:` → `oddlyFormed++; continue`
- `:861` `_, serr := osStatFn(statPath)` — 위 두 분기보다 **뒤**
- `:905-912` `missing := …; unresolvedShape := relativeCount + oddlyFormed;`
  `if missing == 0 && unresolvedShape == 0 { return codexFinding{}, false }`
- `:954-958` `"; %d relative %s (not checked: the resolution base is not observed)"`
- `:959-963` `"; %d oddly-formed %s (not checked: backslash or ~other-user shape)"`

**판정**: (a) AC-CDPG-003 의 도달성 — `relativeCount==1` 이면 `unresolvedShape==1` 이므로 조기 반환을
타지 않고 `ok=true` 로 finding 이 나온다. 확인됨. (b) 「zero-stat 은 판별력이 없다」는 SPEC §B.3 의
주장 — 두 분기 모두 `:861` 앞에서 `continue` 하므로 참. 확인됨. (c) Detail 문구 두 갈래가 실제로
서로 다른 문자열이므로 AC-CDPG-003 의 판별식은 실재한다. 확인됨.

`sed -n '983,988p' internal/cli/doctor_codex.go` → `pluralCodexEntries(1) == "entry"` 이므로 렌더는
`"1 relative entry"`, AC 가 요구하는 부분문자열 `"relative entr"` 를 포함한다. 공허하지 않다.

### E-2 변환 함수와 분류기

```
/usr/bin/grep -rn "func fromConfigPath|func classifyCodexSkillPath" internal/cli/
→ internal/cli/codex_config_path.go:61 / internal/cli/doctor_codex.go:667
```

```go
func fromConfigPath(p string, sep rune) string {
	if sep == '/' { return p }
	return strings.ReplaceAll(p, "/", string(sep))
}

func classifyCodexSkillPath(p string) codexSkillPathShape {
	if filepath.IsAbs(p) { return codexPathAbsolute }
	if p == "~" || strings.HasPrefix(p, "~/") { return codexPathHomeRelative }
	if strings.HasPrefix(p, "~") || strings.ContainsRune(p, '\\') { return codexPathOddlyFormed }
	return codexPathRelative
}
```

**판정**: `C:/Users/u/SKILL.md` 는 darwin 에서 `IsAbs` false → `~` 없음 → 백슬래시 없음 →
`codexPathRelative`. 변환 선행(M-2 뮤턴트) 시에는 `C:\Users\u\SKILL.md` 가 되어 백슬래시 분기로
`codexPathOddlyFormed`. **두 순서가 서로 다른 Detail 문자열을 낸다 — AC-CDPG-003 은 판별한다.**

### E-3 darwin `IsAbs` 프로브 (독립 재측정)

```
go run /private/tmp/claude-501/t570probe/main.go
IsAbs("//?/C:/Users/u/skills/probe/SKILL.md")=true conv="\\\\?\\C:\\Users\\u\\skills\\probe\\SKILL.md" declHasBS=false
IsAbs("C:/Users/u/SKILL.md")=false conv="C:\\Users\\u\\SKILL.md" declHasBS=false
```

(`%q` 이중 이스케이프를 벗기면 변환형은 `\\?\C:\Users\u\skills\probe\SKILL.md`.) SPEC §B.2 의
프로브 출력과 값이 일치한다. AC-CDPG-002 의 도달성 전제와 기대값 모두 참.
프로브 디렉터리는 삭제했고 `git status --short` 는 `?? .moai/specs/SPEC-CODEX-DOCTOR-PATH-GUARD-001/`
한 줄로 프로브 전후 동일하다(내 판정 파일 추가 전 기준).

### E-4 파서 층까지의 도달성 (지시문에 없던 추가 검증)

```
internal/codexwiring/skills.go:174
skillPathKeyRe = regexp.MustCompile(`^path\s*=\s*"([^"]*)"\s*(#.*)?$`)
```

경로는 `[^"]*` 로 **원문 그대로** 캡처된다(TOML 이스케이프 해석 없음). 픽스처 작성기
`writeCodexHomeConfig`(`internal/cli/doctor_codex_test.go:122`)도 `path = "…"` 를 verbatim 으로 쓴다.
따라서 `//?/C:/…` 와 `C:/…` 둘 다 선언 문자열이 훼손되지 않고 `codexStaleSkillFinding` 에 도달한다.
이 층에서 공허해질 경로는 없다.

### E-5 이름 충돌 위험

```
/usr/bin/grep -rn "statRecorder" internal/cli/
internal/cli/codex_skills_prune_readback_test.go:38,39   (주석: 충돌 사고 기록)
internal/cli/doctor_codex_stale_skill_test.go:328,331,337,339
```

`type statRecorder struct` 선언은 `doctor_codex_stale_skill_test.go:331` **1건**. plan.md §B-3 의 좌표
(`:331`, 사고 기록 `codex_skills_prune_readback_test.go:37-41`)와 일치하고, REQ-CDPG-005 가 재선언을
명시적으로 금지한다. **SPEC 은 이 위험을 실제로 막고 있다.**

### E-6 t562 기록 대조 (§B.1 의 세 측정)

- B-1: `/usr/bin/grep -rln fromConfigPath internal/cli/*_test.go` →
  `codex_config_path_test.go`, `codex_skills_prune_readback_test.go` 2건. doctor 측 테스트 없음 — 참.
- B-2: t562 `progress.md:100-107` 은 **4건** 주입(bypass=prune`:83`, blanket-wrap=prune`:101`,
  reorder=prune`:76`, seam=doctor측 diff-count pin)으로 기록한다. 「3건 전부 prune 측」은 결론으로는
  참이나 좌표 표기가 부정확(D2).
- B-3: t562 `acceptance.md:93-109` AC-CSRB-006 은 "green-by-construction … no RED claimed" 와
  "the honest ceiling of doctor-side verification" 를 문자 그대로 담고 있다 — 참.
- 추가 확인: t562 `progress.md:287-295` 가 이 간극을 **자기 카드 안에서 명시적으로 남겨뒀다**
  ("Reverting the doctor-side conversion to `statPath = e.Path` is caught by nothing in this card").
  즉 이 카드의 전제는 t562 자신의 기록으로 교차 확증된다.

### E-7 규율 조항 배치

- 비병렬 + `t.Cleanup`: REQ-CDPG-006(spec L154-157), plan §D D-2(L70-71), AC-CDPG-001 Given(L22-23),
  DoD 1번(L117) — 구현자가 읽을 자리 네 곳 모두에 있다. (덤: `stubCodexHome` 이 `t.Setenv` 을 쓰므로
  런타임도 병렬을 거부한다 — SPEC 이 언급하진 않지만 규율을 강화한다.)
- `-timeout 1200s`: acceptance.md L8-16(모든 AC 공통 검증 명령), plan §D D-3(L72-74),
  §G 안티패턴(L121-122), DoD 2번(L118). **acceptance.md 에 확실히 실려 있다.**
- 테스트 전용 경계: REQ-CDPG-007 + plan D-1 + §G 마지막 항목. **생산 변경이 필요해지면 "블로커 보고
  후 정지"** 로 경로가 명시돼 있다(조용한 범위 확대 금지).

---

## Baseline-attribution

모든 판정은 이 실행에서, 이 트리(`WT-codex-doctor-guard` @ `a4855f0b2`)를 대상으로 잰 값에 귀속된다.
`go run` 프로브는 이 호스트(darwin) 에서 이번에 실행한 것이며, SPEC 이 기록한 프로브 값을 재사용하지
않았다. t562 인용은 이 트리의 `.moai/specs/SPEC-CODEX-SKILL-PATH-READBACK-001/` 사본을 직접 읽었다.

---

## Defects Found

- **D1** — `plan.md:57-59` — §C 표가 `writeCodexHomeConfig` / `stubCodexHome` 를 "same file"
  (= `doctor_codex_stale_skill_test.go`) 로 적었으나 실제 위치는
  `internal/cli/doctor_codex_test.go:122` / `:103` 이다. 같은 패키지라 호출은 되지만 좌표가 틀렸다.
  재현: `/usr/bin/grep -rn "func writeCodexHomeConfig" internal/cli/`.
  Severity: minor · Class: optional · 수정: 표의 Location 칸을 `doctor_codex_test.go:122 / :103` 로.
- **D2** — `spec.md:60` — B-2 셀 "all three executed mutants sit on the prune side,
  `codex_skills_prune.go:83`". t562 `progress.md:100-107` 은 4건 주입(prune `:83`/`:101`/`:76` +
  doctor 측 seam diff-pin)을 기록한다. 결론은 유지되나 좌표가 하나로 뭉개졌다.
  Severity: minor · Class: optional · 수정: 세 prune 좌표를 나열하고, 네 번째(doctor seam)는
  **행위 가드가 아니라 diff-count pin** 임을 한 줄로 밝힐 것.
- **D3** — `spec.md:73-75` — "It never named a path family for which the unconverted form actually
  fails. This card names one." 그러나 t562 `progress.md:292-295` 가 이미
  `//?/C:/… → \\?\C:\…` 를 sync-audit F6 로 지목했고 이 카드가 그 후속으로 발행됐다.
  Severity: minor · Class: optional · 수정: "t562 의 SPEC 본문은 명명하지 않았고, 그 카드의
  sync-audit F6 가 명명해 이 카드로 이월했다" 로 귀속을 정정.
- **D4** — `spec.md:146-162` + `acceptance.md:20-102` — **REQ-CDPG-005 / 006 / 007 에 대응하는
  AC-XXX 가 없다.** 셋 다 DoD 체크박스(`acceptance.md:117-123`)로 덮이긴 하나, AC↔REQ 매핑상은
  미커버다. 재현: `/usr/bin/grep -n "maps REQ-" acceptance.md` → 004 까지만 나온다.
  Severity: major · Class: **blocking** · 수정: AC-CDPG-005/006/007 을 추가하거나(각 DoD 명령을
  Given-When-Then 으로 승격), acceptance.md 에 `REQ-005→DoD5 / REQ-006→DoD1 / REQ-007→DoD4`
  명시 매핑 한 줄을 넣을 것. Tier S AC 상한 8 이므로 4건 추가해도 여유가 있다.
- **D5** — `acceptance.md:28-30` — AC-CDPG-001 의 RED-now 셀이 현재형으로 "This revert is executed,
  not hypothesised" 라고 쓰여 있다. plan 단계에서는 아직 아무것도 실행되지 않았으므로, 이 문장은
  **관측 주장으로 읽힐 수 있다**(이 카드가 t562 에게서 물려받은 실패 형태와 같은 계열).
  Severity: minor · Class: **blocking**(문서가 스스로 세운 「관측되지 않은 RED 는 채택하지 않는다」
  기준을 이 셀이 흐린다) · 수정: "M1 step 1 에서 실행하고 이 셀을 verbatim FAIL 출력 + 증거 경로로
  채운다 — 채워지기 전까지 이 셀은 의무이지 관측이 아니다" 로 시제·성격을 명시.
- **D6** — `spec.md:161-162` — REQ-CDPG-007 둘째 문장이 `Should …` 조건절이라 다섯 GEARS 패턴 어디에도
  들어맞지 않는다(legacy If/then 호환 창 안이라 MINOR).
  Severity: minor · Class: optional · 수정: `When a production change proves necessary, the
  implementer shall report it as a finding and stop.`
- **D7** — `acceptance.md:24-26` — AC-CDPG-001 이 "an absolute slash-form path" 로만 말하고 픽스처
  문자열을 고정하지 않는다. darwin 절대경로는 어느 것이든 `/` 를 포함하므로 **공허해지지는 않는다**
  (변환형 ≠ 선언형이 항상 성립). 다만 002/003 과 달리 자유도가 남는다.
  Severity: minor · Class: optional · 수정: 픽스처 경로를 002/003 처럼 고정하거나, "선언형이
  최소 하나의 `/` 를 포함할 것" 을 명시.

**공허성 재검(이 감사의 주 임무) — 결과: 재발 없음.**
AC-001/002/003 세 건 모두 (a) 지정 픽스처가 실제로 그 코드 경로에 도달하고(E-1·E-3·E-4),
(b) 올바른 구현과 명명된 결함을 구별한다(E-1·E-2). t562 AC-CSRB-006 의 darwin-identity 공허성은
**분리자 고정(`overrideSeparator`)** 으로 실제로 해소된다 — `fromConfigPath` 는 `sep=='/'` 일 때만
항등이고, `'\\'` 로 고정하면 항등이 아니다(E-2 소스). 그리고 SPEC 은 자기 손으로 판별력 없는
zero-stat 단언을 **비판별로 라벨링**하고 대체 판별식(Detail 문자열)을 세웠다 — 소스 대조 결과
그 라벨링이 정확하다.

---

## Gaps (관측하지 **않은** 것)

- **G-1** MCP 서버 미연결로 `mcp__moai__spec_audit` / `spec_drift` / `audit_multi` 를 실행하지 못했다.
  기계 린트 축(GEARS modality 검사, 교차모델 2차 의견)은 이번 판정에 없다. 내 GEARS·frontmatter
  판정은 수동 대조다.
- **G-2** 테스트를 **작성하거나 실행하지 않았다**. 「AC 가 통과 가능하다」는 소스 대조 추론이며,
  실제 컴파일·실행 결과가 아니다. 특히 `go test ./internal/cli/... -timeout 1200s` 는 돌리지 않았고
  「~615s 런타임」 주장도 재측정하지 않았다(레인 부하 규율 §4).
- **G-3** M-1/M-2 뮤턴트를 주입해보지 않았다. 「M-2 가 AC-003 을 FAIL 시킨다」는 분류기 소스로부터의
  연역이지 실행 관측이 아니다 — 그 실행은 run-phase 의 AC-CDPG-004 의무다.
- **G-4** Windows 에서 `\\?\` 해석 동작을 확인하지 않았다. SPEC 도 확인했다고 주장하지 않으며
  R-1 로 명시해 뒀다 — 주장과 증거가 일치한다.
- **G-5** `.moai/specs/` 전수 스윕으로 SPEC ID 유일성을 재검하지 않았다(progress.md 의 기록만 읽었다).

---

## Residual Risk

- **RR-1** AC-CDPG-003 의 판별식은 Detail 문자열이므로, `SPEC-CODEX-STALE-SPLIT-FOURTH-001` 계열의
  후속 변경이 relative/oddly-formed 렌더를 통합하면 가드가 조용히 약해진다. SPEC 은 A-2 로 이 위험을
  이미 적어뒀지만, 문서 주석일 뿐 기계적 경보는 아니다.
- **RR-2** 이 가드는 `osStatFn` 에 넘어가는 **인자**만 고정한다. 변환 이후 실제 파일시스템 결과는
  주장하지 않는다(SPEC R-1 과 동일). 즉 "Windows 가 변환형을 해석한다" 는 여전히 CI 매트릭스 문제다.
- **RR-3** 뮤턴트 창 동안 `internal/cli/doctor_codex.go` 는 공유 트리의 파일이다. acceptance.md
  L100-101 이 이를 짚었으나, 배타성 확인 방법은 명시하지 않는다(주입 직전 재확인 권고).
- **RR-4** D4 를 고치지 않고 run 으로 넘어가면 REQ-005/006/007 의 통과 근거가 DoD 체크박스 자기보고에
  머문다 — 자기보고는 관측이 아니다.

---

## Recommendation

**PASS (0.87 ≥ Tier S 0.75).** 이 SPEC 은 감사가 겨눈 실패(t562 의 공허한 AC 재발)를 **반복하지 않는다**:
세 AC 전부 도달성과 판별력이 소스에서 확인됐고, 판별력 없는 단언을 스스로 라벨링해 대체 판별식을 세웠다.

run 진입 전 처리 권고(둘 다 blocking 급, 각 1~3줄):

1. **D4** — REQ-CDPG-005/006/007 ↔ DoD 매핑을 acceptance.md 에 명시하거나 AC 3건을 추가한다.
2. **D5** — AC-CDPG-001 RED-now 셀을 「의무」 시제로 고쳐, 채워지기 전에 관측 주장으로 읽히지 않게 한다.

D1/D2/D3/D6/D7 은 optional — 운영자 재량. 다만 D2/D3 은 이 카드의 **자기 출처 기록**이므로 지금
고치는 편이 싸다(카드가 닫힌 뒤에는 귀속을 되돌리기 어렵다).

범위 비례성: 산출물 490행 대 코드 ~20행은 비율상 무겁다. 다만 acceptance.md 는 리드의 명시 요청으로
작성됐고(progress.md §E.1 에 기록), Tier S 상한(REQ 8 / AC 8)은 7 / 4 로 준수한다 — 팽창으로
판정하지 않는다.
