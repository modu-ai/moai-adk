# SPEC Review Report: SPEC-CODEMAPS-REFRESH-002

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.74** (Tier M PASS threshold 0.80)

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t475` · 브랜치 `WT-codemaps-stale` · HEAD `52f863f3666c9ec754253a06b96ed1fe844f1590` (= `origin/develop`)

M1 Context Isolation: 저작자의 추론 맥락은 무시했다. 판정은 `spec.md` / `plan.md` / `acceptance.md` (Tier M 3-아티팩트) 와 이 트리에서 직접 실행한 명령의 출력에만 근거한다.

---

## Claim

이 SPEC은 **must-pass 7항목을 전부 통과**하며 인용한 수치는 **전수 재현된다.** 그럼에도 **선정 집합 2개가 재현 불가능**하고, **마일스톤 명령 순서가 복구 불가능한 증거 손실을 유발**하며, **MUST AC 3개가 카드의 실제 문제가 남아 있는 채로 통과 가능**하다. 따라서 Tier M 임계 0.80에 미달한다.

---

## Evidence

### Must-Pass Results

| # | 항목 | 판정 | 증거 |
|---|------|------|------|
| MP-1 | REQ 번호 일관성 | **PASS** | `grep -c "^- \*\*REQ-CM2-" spec.md` → `13`. REQ-CM2-001~013 연속, 결번·중복·제로패딩 불일치 0 |
| MP-2 | GEARS 형식 준수 (요건 층) | **PASS** | 13개 REQ 전부가 5개 GEARS 패턴 중 하나에 적중. 판정 대상은 `spec.md`의 `REQ-XXX` 요건 층이며, `acceptance.md`의 Given-When-Then `AC-XXX`는 검증 층 정상 형식이므로 여기서 감점하지 않았다 |
| MP-3 | YAML frontmatter | **PASS** | `moai spec lint .moai/specs/SPEC-CODEMAPS-REFRESH-002/spec.md` → `0 error(s), 13 warning(s)`. 12 정본 필드 전수 존재. `phase: "v3.2.0 target"` 는 릴리스 타깃이며 금지된 lifecycle 토큰이 아님 |
| MP-4 | §22 언어 중립성 | **N/A** | 이 저장소 내부 문서 SPEC. 16 프로그래밍 언어 툴링을 다루지 않음 (auto-pass) |
| MP-5 | D7 교차-SPEC 정합 | **PASS** | 참조 SPEC 8건 전부 `.moai/specs/` 에 실재하며 전부 `status: completed`. `retired`/`superseded`/`archived` 0건 → BLOCKING 없음 |
| MP-6 | D8 크로스플랫폼 | **PASS** | `grep -c "syscall"` → spec/plan/acceptance/progress 전부 `0` (자동 PASS) |
| MP-7 | clarification 게이트 | **PASS** | `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEMAPS-REFRESH-002/` → exit 1 (무매치). `research.md` 부재는 Tier M 규정상 정상 |

### 수치 전수 재측정 — SPEC이 인용한 모든 값

| SPEC 주장 | 재측정 | 관측 | 일치 |
|---|---|---|---|
| `described-source-diff value=64 threshold=40 verdict=stale`, exit 1 | `.moai/reports/t475/graph-check-baseline.txt` (수출본) | 동일 | ✓ |
| 64 의 정체 | `awk '{print $2}' <diff> \| grep '\.go$' \| grep -v '_test\.go$' \| wc -l` → `64`; `internal/mx/described_worthy.go:13-26` 의 술어(`.go` && !`_test.go` && !`testdata/`)와 일치 | 64 | ✓ |
| 앵커 `25a3212a9`, `described_roots [internal,cmd,pkg]`, `dirty:false`, `tree_root=.../t476` | `cat .moai/project/codemaps/provenance.json` | 동일 | ✓ |
| `78 A / 1 D / 121 M` | `git diff --name-status 25a3212a9 HEAD -- internal cmd pkg \| awk '{print $1}' \| sort \| uniq -c` | `78 A / 1 D / 121 M` (총 200행) | ✓ |
| 수출 증거의 무결성 | `diff -q <live-diff> .moai/reports/t475/described-roots-diff-since-anchor.txt` | **EXPORTED-DIFF-IDENTICAL** | ✓ |
| `go list ./internal/... ./cmd/... ./pkg/... \| wc -l` = 136 | 동일 명령 | `136` | ✓ |
| 미히트 48, 그중 `internal/harness/*` 12 | `wc -l < packages-absent-from-codemaps.txt`; `grep -c '^internal/harness/'` | `48` / `12` | ✓ |
| 6개 단위 히트 0 | 6문서 연결 텍스트에 `/usr/bin/grep -c -F` | 6개 전부 `0` | ✓ |
| 부모 히트 `internal/template`=19, `internal/web`=5, `internal`=236, `internal/harness`=4 | 동일 규약 | `19 / 5 / 236 / 4` | ✓ |
| 집계 규약(적중 **행** 수; 발생 수는 281) | `grep -c -F "internal"` → `236`; `grep -o -F "internal" \| wc -l` → `281` | 동일 | ✓ |
| 상위 변경 구간 `cli 59 / web 13 / statusline 12 / kanban 11 / template 10 / codexwiring 6 / commandemit 5` | `awk '{print $2}' <diff> \| xargs -n1 dirname \| sort \| uniq -c \| sort -rn` | 전부 동일 | ✓ |
| 팬텀 6개(`internal/{design,evaluator,factory,migrate,research,state}`)가 트리 부재 + 히트 0 | `[ -d internal/<p> ]` + `grep -c -F` | 6개 전부 `dir=no cmhits=0` | ✓ |
| §A.4 `check.go:341-346` 주석 / `:456` / `:557` TreeRoot 비교 | `sed -n '330,350p'` + `grep -n "TreeRoot != "` | 주석 341-346 위치 정확, 비교는 `:456`·`:557` 두 곳뿐이며 `checkCodemaps`(315-444) 경로에 없음 | ✓ |
| CADENCE-001 "임계 40 유지" | `.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/spec.md:368` — *"D.2 — (b) Is 40 still the right threshold? **Yes. Retain it, on a corrected justification.**"* | 일치 | ✓ |
| REFRESH-001 REQ-CMR-002~005 계승 관계 | `SPEC-CODEMAPS-REFRESH-001/spec.md:83,85,87,89` — REQ-CMR-005 도 `origin/develop` merge-base 명시 | 일치 | ✓ |
| `moai graph stamp codemaps --commit` 플래그 실재 | `./bin/moai graph stamp codemaps --help` → `--commit  Explicit commit anchor (any rev-parse expression...)` | 실재 | ✓ |
| `/moai codemaps --force` 플래그 실재 | `grep -n -- "--force" .claude/skills/moai/workflows/codemaps.md:36` | 실재 | ✓ |

**인용된 수치 중 재현되지 않은 것은 하나도 없다.** 카드 텍스트의 144 → 64 정정도 옳다.

### RED-now 대상 실재 확인

| RED 대상 | 현재 상태 |
|---|---|
| `.moai/reports/t475/codemaps-accuracy-verification.md` | 부재 (`No such file or directory`) |
| `.moai/reports/t475/verdict.md` | 부재 |
| `.moai/reports/t475/pre-regen/` | 부재 |

---

## Baseline-attribution

이 감사의 모든 측정은 **이 실행에서, 이 트리에 대해** 수행됐다: `git rev-parse --show-toplevel` = `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t475`, `git rev-parse HEAD` = `52f863f3666c9ec754253a06b96ed1fe844f1590`, `git branch --show-current` = `WT-codemaps-stale`, `git merge-base HEAD origin/develop` = `52f863f3666c9ec754253a06b96ed1fe844f1590` (HEAD 와 동일). `git status --porcelain` 은 `?? .moai/reports/t475/` 와 `?? .moai/specs/SPEC-CODEMAPS-REFRESH-002/` 두 항목만 보고한다 — 감사 시작 시점 트리는 이 카드의 산출물 외에 깨끗하다.

판정 도구: `./bin/moai` (이 워크트리에서 빌드된 바이너리; `bin/` 은 gitignored 라 `git status` 에 나타나지 않는다).

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | 판별식(§A.3(a1))은 정확하고 모범적이다. 그러나 선정 집합 2개(6단위 / 6구간)가 재현 불가하며(D1, D5), `acceptance.md:161` 의 "관측 3항목" 이 REQ-CM2-013·AC-CM2-012 의 "2항목" 과 직접 충돌한다(D4) |
| Completeness | 0.75 | 0.75 | HISTORY/WHY/WHAT/HOW/REQUIREMENTS/AC/Out of Scope 전부 존재, `### Out of Scope — <주제>` H3 4개 + `-` 불릿 존재, frontmatter 12/12. 다만 `docs-truth.md` 의 생성 주체가 미기술이며(D3), M2 사전 사본 절차가 명령 순서상 뒤에 놓여 있다(D2) |
| Testability | 0.65 | 0.50~0.75 | AC-CM2-006/008 이 추출 명령을 명명하지 않고(D6), AC-CM2-005 의 두 번째 분기가 반증 불가능한 자유 서술이며(D7), AC-CM2-002 가 부모 산문 인용을 요구하지 않아 미관측 전제를 허용한다(D8). 12개 MUST AC 중 9개에 RED-now 셀이 없다(D9) |
| Traceability | 0.80 | 0.75~1.0 | §D.2 가 12 AC ↔ 12 REQ 를 명시 매핑. REQ-CM2-011 만 단일 AC 대응이 없으나 그 이유(구조 요건)가 명시돼 있고 실질적으로 AC-CM2-006~008 이 담당한다. `moai spec lint` 의 13건 `CoverageIncomplete` 는 린터가 `spec.md` 단독 스캔 시 `acceptance.md` 의 AC 를 못 보는 알려진 형태이며(REFRESH-001 도 동일), SPEC 결함이 아니다 |

산술 평균 = (0.75 + 0.75 + 0.65 + 0.80) / 4 = **0.7375 → 0.74**. Tier M 임계 0.80 미달.

---

## Defects Found

**D1. 6단위 선정 집합이 재현 불가능하며, 카드가 고치려는 바로 그 종류의 단위를 빠뜨린다** — `spec.md:68` (§A.3(a) 표제 "문서 전체에서 히트 0인 실재 단위 **6개**"), `spec.md:75-82` (표) — Severity: **critical** — Class: **blocking**

측정: `internal/template/agentemit` 는 (a) `go list` 가 열거하는 실재 패키지이고(`packages-absent-from-codemaps.txt:46`), (b) SPEC 자신의 집계 규약으로 히트 **0** 이며(`grep -c -F "internal/template/agentemit"` → `0`), (c) 앵커 이후 변경됐고(`M internal/template/agentemit/agents-codex.yaml`), (d) 운영자가 명명한 omission 사례 `internal/template/commandemit` 와 **구조적 쌍둥이**다 — 둘 다 `internal/template` 하위의 기계 방출(emit) 패키지이며 부모의 19개 적중 행 어디에도 그 발행 책임이 서술돼 있지 않다. 그런데 6에 없다.

또한 exact-path 기준으로 히트 0인 패키지는 **48개**다. 6을 48에서 어떻게 골랐는지 규칙이 서술돼 있지 않고, 어떤 후보 규칙으로도 재현되지 않는다 — "앵커 이후 추가된 단위"라면 `internal/chain` 이 탈락하고(`git log -1 -- internal/chain` → `1444583bf 2026-09-03 style(t457): gofmt entire tree`, 신규 추가가 아님), "앵커 이후 변경된 단위"라면 `agentemit` 이 포함돼야 한다.

**결과가 곧 공허한 통과다.** REQ-CM2-002/AC-CM2-002 는 M1 판별을 정확히 이 6개에만 건다. `agentemit` 은 M1 을 통과하지 않고 AC-CM2-007 의 **기록 전용** 분류(`분류는 기록 대상이지 수정 대상이 아니다`, `spec.md:196`)로 흘러간다. 즉 SPEC 자신의 M3 가 "누락"으로 분류할 단위가 문서에 편입되지 않은 채 AC 12개가 전부 PASS 로 닫힌다 — 카드의 목표("genuinely-absent units 를 편입한다")가 남은 채로.

Required fix: §A.3(a) 의 6단위 집합에 대해 **명명된 명령으로 재현 가능한 선정 규칙**을 서술하고, 그 규칙을 실제로 적용해 결과 집합을 확정한다. `internal/template/agentemit` 는 그 규칙에 따라 포함하거나, 포함하지 않는다면 **`commandemit` 과 동일한 책임 질문**으로 제외 근거를 §A.3(a1) 형식으로 기술한다. (접힘 정책 자체는 건드리지 않는다 — 운영자 결정 2·§B.2 준수.)

---

**D2. M2 의 명령 순서가 AC-CM2-005 가 "복구 불가"로 규정한 증거를 파괴한다** — `plan.md:99-108` — Severity: **major** — Class: **blocking**

M2 본문은 `# /moai codemaps --force (스킬 실행)` 를 첫 코드블록(`plan.md:100`)에 두고, 재생성 **전** 사본을 뜨는 `mkdir -p .moai/reports/t475/pre-regen && cp ...` 를 그보다 **아래**(`plan.md:107`)에 둔다. 산문은 "미리 떠 두어야 한다"고 말하지만, 실행자가 마일스톤 명령을 위에서 아래로 따르면 사본은 재생성 **후**에 떠진다. AC-CM2-005(`acceptance.md:90`)는 "재생성 전 사본이 없으면 이 AC 는 판정 불가이며 FAIL 로 처리한다(사후 복구 불가)"라고 명시한다. 현재 `.moai/reports/t475/pre-regen/` 는 부재이므로 이 함정은 살아 있다.

Required fix: `cp` 블록을 M2 의 **첫 단계**로 올리고, `/moai codemaps --force` 는 그 뒤에 둔다. 사본 존재 확인(`ls .moai/reports/t475/pre-regen/ | wc -l` = 6)을 재생성 실행의 선행 조건으로 명시한다.

---

**D3. `docs-truth.md` 는 `/moai codemaps --force` 의 산출물이 아닌데 AC-CM2-003 이 "재생성 완전성"으로 통과 판정한다** — `spec.md:188` (REQ-CM2-003), `acceptance.md:71-76` (AC-CM2-003), `plan.md:100-102` — Severity: **major** — Class: **blocking**

측정: `.claude/skills/moai/workflows/codemaps.md` Phase 3 의 산출 파일 목록은 **5개**다 — `overview.md`, `modules.md`, `dependencies.md`, `entry-points.md`, `data-flow.md`. 같은 파일 전체에 `docs-truth` 는 **0회** 등장한다(`grep -n "docs-truth" .claude/skills/moai/workflows/codemaps.md` → 무출력). 실제 `docs-truth.md` 는 손으로 저작된 Docs-v3 코호트 산출물이다(파일 머리말: *"Canonical Facts Checklist for the Docs-v3 Cohort — Navigation aid, NOT a new SSOT"*; 저작 SPEC `SPEC-V3R6-DOCS-CODEMAPS-V3-001`, `SPEC-V3R6-DOCS-V3-README-001/plan.md:185` 이 인용).

따라서 AC-CM2-003 의 "6개 문서가 **전부 재생성 대상으로 보고**된다. 일부만 갱신되면 FAIL" 은 명시된 실행면으로는 만족될 수 없다. 반면 같은 AC 의 두 번째 증거인 `ls .moai/project/codemaps/`(7개 항목)는 `docs-truth.md` 가 손대지 않은 채 남아 있어도 **그대로 통과한다** — 공허한 통과다. plan.md 어디에도 `docs-truth.md` 를 손으로 갱신하라는 지시가 없다. 11,162 byte 의 산문이 조용히 낡은 채 남는다.

Required fix: `docs-truth.md` 가 스킬 산출물이 아니라 **손으로 유지되는 문서**임을 §B 또는 M2 에 명시하고, 그 갱신을 별도 단계로 지시한다(최소한 §1 에이전트 카탈로그 표 전수 대조 — REFRESH-001 REQ-CMR-004 가 이미 그 검증 경계를 정해 두었다). AC-CM2-003 문구를 "생성기 산출 5문서 + 손 유지 1문서"로 이분해 각각의 증거를 요구한다.

---

**D4. `acceptance.md` §D.3 의 "관측 3항목" 이 REQ-CM2-013·AC-CM2-012 의 "2항목" 과 충돌한다** — `acceptance.md:161` vs `spec.md:208` / `acceptance.md:132` / `plan.md:136` — Severity: **major** — Class: **blocking**

`acceptance.md:161` Definition of Done: "`verdict.md`(관측 **3항목**)". REQ-CM2-013(`spec.md:208`), AC-CM2-012(`acceptance.md:132`), plan.md M5(`plan.md:136`) 는 전부 **2항목**이다. 그리고 AC-CM2-012 는 "verdict.md 에 wrong-tree 항목이 실려 있으면 **FAIL**"이라고 못박는다(`acceptance.md:133`) — DoD 를 따라 세 번째 항목을 채우려는 실행자가 가장 먼저 손을 뻗을 항목이 정확히 그것이다. 두 규정이 서로 반대를 지시한다.

Required fix: `acceptance.md:161` 을 "관측 2항목"으로 정정한다.

---

**D5. §A.3(b) "상위 구간" 6개 선정에 컷오프가 없고, 동일 수치의 구간 하나가 누락돼 있다** — `spec.md:112` / `spec.md:192` (REQ-CM2-005) / `acceptance.md:88` (AC-CM2-005) — Severity: **major** — Class: **blocking**

측정(`awk '{print $2}' <diff> | xargs -n1 dirname | sort | uniq -c | sort -rn`):

```
59 internal/cli        13 internal/web         12 internal/statusline
11 internal/kanban     10 internal/template     8 internal/template/templates/.claude/rules/moai/workflow
 6 internal/settings    6 internal/codexwiring   5 internal/template/templates/.codex/agents/moai
 5 internal/template/commandemit                 5 internal/spec
```

SPEC 이 나열한 7개 수치는 전부 정확하다. 그러나 `internal/settings` 는 **6** 으로 포함된 `internal/codexwiring` 과 **동률**인데 목록에 없다. REQ-CM2-005 와 AC-CM2-005 는 재기술 의무를 명시된 6구간에만 걸고 "구간이 누락되면 FAIL" 이라고 하므로, 누락 자체가 판정 범위 밖으로 나가 조용히 통과한다.

Required fix: 선정 컷오프를 명시한다(예: "described roots 하위, `internal/template/templates/**` 의 비-Go 템플릿 콘텐츠를 제외한 변경 파일 수 ≥6"). 그 규칙을 적용하면 `internal/settings` 가 포함되며 목록은 7구간이 된다. AC-CM2-005 의 구간 수를 그에 맞춘다.

---

**D6. AC-CM2-006/008 이 추출 명령을 명명하지 않는다 — 저장소 안에 정본 추출식이 이미 있는데도** — `spec.md:194,198` / `acceptance.md:92-96,104-108` / `plan.md:114,116` — Severity: **major** — Class: **blocking**

AC-CM2-006 은 "유니크 경로가 하나라도 누락되면 FAIL" 로 이분 판정한다고 선언하지만, 무엇을 경로로 셀지 정하는 명령·정규식이 어디에도 없다. 느슨한 추출로 짧은 표를 만들어도 그 사실을 판별할 근거가 없으므로 실질적으로 판정 불가다. AC-CM2-008 은 더 나쁘다 — "식별자" 의 정의도, 추출 명령도 없다.

그런데 정본이 트리에 있다: `internal/graph/check_citations.go:23`

```go
var citedPathPattern = regexp.MustCompile(`\b(?:internal|pkg|cmd)/[A-Za-z0-9_/.-]*`)
```

여기에 후행 구두점 절삭(`:35` `citedPathTrailingPunct = ".,;:)]}\"'"`)과 **blockquote(부정 인용) 면제**가 붙어 있다. 손으로 만든 추출이 이 면제를 빠뜨리면, 일부러 부존재를 인용한 행(제거 기록·rename 이력)이 전부 "absent" new-finding 으로 잘못 분류된다. 더구나 `moai graph check` 의 `citations` 계층이 같은 측정을 이미 기계적으로 수행한다(`positive-cited-path-absence`, threshold 0) — spec.md §A.1 이 그 행을 인용하면서도 REQ-CM2-006 과 연결하지 않는다.

Required fix: REQ-CM2-006/AC-CM2-006 이 `check_citations.go:23` 의 정규식(후행 구두점 절삭 + blockquote 면제 포함)을 추출 규약으로 명명하고, `moai graph check --json` 의 `citations` 계층 출력을 교차 대조 근거로 인용한다. AC-CM2-008 은 "식별자" 를 명명된 추출 명령으로 정의한다.

---

**D7. AC-CM2-005 의 두 번째 분기가 반증 불가능한 자유 서술이다** — `acceptance.md:89` — Severity: **minor** — Class: **blocking**

"각 행이 '서술이 바뀐 지점' **또는** '바뀌지 않았고 그것이 옳은 이유'를 담는다." 두 번째 분기는 명령으로 반증되지 않는다. 뮤턴트가 존재한다: 재생성이 그 구간에 대해 아무것도 하지 않았어도 6행 전부에 "변경 없음 — 기존 서술이 여전히 정확함" 을 적으면 통과한다. 대조 명령도 명명돼 있지 않다("전후 서술을 대조하면").

Required fix: 구간별로 `diff -u .moai/reports/t475/pre-regen/<doc>.md .moai/project/codemaps/<doc>.md` 를 명명하고, "변경 없음" 행은 **빈 diff 출력**을 증거로 첨부하도록 요구한다.

---

**D8. AC-CM2-002 가 부모 산문의 인용을 요구하지 않아 미관측 전제를 허용한다** — `acceptance.md:67-69` — Severity: **minor** — Class: **blocking**

AC-CM2-002 의 삼분 절단(히트 수만 → FAIL / 변경 파일 수만 → FAIL / `commandemit` 을 fold 로 → FAIL)은 세 함정을 정확히 막는다. 그러나 "부모의 기존 서술이 그 책임을 담는지/담지 않는지를 **진술한다**" 는 **진술**만 요구하고, 실행자가 실제로 어느 부모 산문을 읽었는지는 요구하지 않는다. 통과하는 뮤턴트: "`internal/chain`: 체인 노드 구성 책임. 부모 `internal` 의 서술이 이를 담지 않음 → omission" — 패키지 이름에서 유도한 책임 서술 + 부모 산문을 한 줄도 읽지 않은 부정. 이는 저장소 자신의 `verification-claim-integrity.md` §1.1 surface 4(권고 전제 주장)가 금지하는 형태다.

Required fix: 각 판정 행에 **검사한 부모 산문의 위치**(`<문서>.md:L<n>` 또는 인용문)를 요구한다. `fold` 판정은 책임을 담는 그 줄을 **인용**해야 하고, `omission` 판정은 검사한 부모 적중 행의 범위(예: `internal/template` 19행 전수)를 명시해야 한다.

---

**D9. MUST 선언된 12개 AC 중 9개에 RED-now 셀이 없다** — `acceptance.md:24-53` (§B RED-now 원장), `acceptance.md:137` (§D.1 "전 항목 MUST") — Severity: **minor** — Class: **blocking**

§B 는 RED-1/RED-2/RED-3 세 개만 담는다. §D.1 은 12개 AC 전부를 MUST(= release-blocking)로 선언한다. `verification-completeness.md` §2 는 "한 셀만 있는 기준은 미채택" 이고 §2.1 은 release-blocking 기준에 4요소(명령 / verbatim stdout / exit code / 트리 SHA)를 요구한다. AC-001/003/006/007/008/009/011/012 및 AC-005 의 실제 술어에 대해 RED 셀이 없다.

부수적으로 RED-3 은 두 가지가 어긋난다. (a) stdout 자리에 verbatim 출력(200행) 대신 **집계**("78 A / 1 D / 121 M")를 적었다 — 실제 verbatim 은 `.moai/reports/t475/described-roots-diff-since-anchor.txt` 에 수출돼 있고 내가 바이트 동일함을 확인했으나, 원장이 그 경로를 인용하지 않는다. (b) RED-3 은 AC-CM2-005 의 술어(증거 파일에 6행이 존재하는가)가 아니라 드리프트 동기를 측정한다 — 원장 자신이 그 사실을 인정한다(`acceptance.md:49`).

수리 비용은 낮다. 대부분의 RED 는 한 줄이며 지금 실제로 붉다: `ls .moai/reports/t475/codemaps-accuracy-verification.md` → `No such file or directory`, exit 1, tree `52f863f36`(내가 실행해 확인).

Required fix: 각 MUST AC 에 4요소 RED-now 셀을 붙인다. RED-3 은 수출 파일 경로를 인용하고 집계는 "파생값" 으로 라벨한다.

---

### Optional findings (blocking 아님 — 리드 재량)

- **O1.** `plan.md:61,123` 이 스탬프 리비전을 `/tmp/t475-stamp-rev` 에 저장하고 §E 자체검증이 그 파일을 읽는다. `plan.md:52` 와 `acceptance.md:22` 는 "`/tmp` 저장은 판정 근거로 인정되지 않는다" 고 선언한다 — 내부 긴장. `provenance.json` 의 `commit_sha` 를 직접 읽는 편이 더 강하다(AC-CM2-009 는 이미 그 일치를 요구한다).
- **O2.** REQ-CM2-001 과 REQ-CM2-003 은 `(Ubiquitous)` 로 라벨돼 있으나 본문은 각각 "run 시작 시", "`/moai codemaps --force` 실행 시" 로 event-driven 이다. 패턴 자체는 GEARS 에 적중하므로 MP-2 는 통과하며, 라벨만 부정확하다.
- **O3.** `acceptance.md` 에 §C 가 없다(§B → §D 로 건너뜀). 표기상 문제일 뿐이다.
- **O4.** 재발 원인의 귀속이 한 칸 비어 있다. §A.2 는 CADENCE-001 의 누적 지표 메커니즘을 인용해 "왜 다시 붉어지는가"에 답하며 — **이 답은 옳고, 절차 반복이 이 카드의 정답이다** — 다만 현재 앵커 `25a3212a9` 를 찍은 주체가 REFRESH-001(2026-09-02)이 아니라 **다른 워크트리 t476 이 2026-09-03T18:18:34Z 에** 찍은 것이라는 사실(`provenance.json` 직독)이 어디에도 없다. 5일 만에 64 에 도달한 것은 CADENCE-001 의 "corrected-40 이 약 1.6일에 교차" 와 정합하며, REQ-CM2-013 이 요구하는 "앵커 이후 누적 속도" 관측의 좋은 재료다 — verdict.md 에 담을 것을 권한다.
- **O5.** 형제 아티팩트(`plan.md` / `acceptance.md`)에 대한 `moai spec lint` 는 `FrontmatterInvalid`(status/priority/lifecycle/tags 누락) + `MissingExclusions` + `DuplicateSPECID` 를 낸다. 이는 **린터 쪽 형태**이며 SPEC 결함이 아니다 — `spec-frontmatter-schema.md` § Artifact Statelessness 가 형제 아티팩트의 `status:` 부재를 **요구**하고, `completed` 상태로 통과한 REFRESH-001 도 동일한 6건을 낸다. `spec.md` 단독 린트는 `0 error(s)`.

---

## Gaps — 명시적으로 관측하지 않은 것

1. **`/moai codemaps --force` 를 실행하지 않았다.** 재생성이 실제로 어떤 문서를 건드리는지, `internal/template/commandemit` 을 자동 편입하는지는 미관측이다. D3 은 스킬 문서의 **선언된 산출 목록**과 `docs-truth.md` 파일 머리말에 근거하며, 실행 관측이 아니다.
2. **`moai graph stamp codemaps --commit` 을 실행하지 않았다.** `--commit` 플래그의 실재와 help 문구만 확인했다. 재스탬프 후 `graph check` 가 실제로 `value=0 verdict=fresh` 를 내는지는 미관측이다 — 코드 독해(`resolveContentAnchor` Rule A, `check.go:252-294`)상 그렇게 되어야 하지만 실행 증거는 없다.
3. **48개 미히트 각각의 접힘/누락 성질을 판정하지 않았다.** `agentemit` 하나만 `commandemit` 과의 구조적 동형성을 근거로 검사했다. 나머지 46개 중 D1 과 같은 사례가 더 있는지는 미측정이다.
4. **`origin/develop` 을 fetch 하지 않았다.** `merge-base` 와 `origin/develop` 판독은 이 워크트리의 원격 추적 ref 현재값에 근거한다. 이 감사 시점 이후 `origin/develop` 이 움직였을 수 있다.
5. **다른 감사 백엔드(codex / GLM)를 호출하지 않았다.** 이 판정은 Claude 단독 앵커다.
6. **6문서의 산문 정확성 자체를 읽지 않았다.** 이 감사는 plan-phase 감사이며 codemaps 내용의 사실성은 M3 의 소관이다.

---

## Residual-risk

1. **선정 집합 결함이 D1·D5 두 곳에서 같은 형태로 나타났다** — 이 카드의 판단이 측정보다 앞서가는 경향(리드가 지적한 이력)이 아직 살아 있다는 신호다. 세 번째 미발견 사례가 남아 있을 수 있다(Gap 3).
2. **게이트 녹색은 재생성 품질과 무관하다.** `resolveContentAnchor` Rule A 에 따라 재스탬프만으로 `value=0` 이 나온다 — 재생성을 건너뛰어도 AC-CM2-010 은 통과한다. SPEC 은 이를 알고 REQ-CM2-011 + AC-CM2-006~008 로 방어하며 RED-1 의 green path 도 정직하게 "M4 재스탬프가 이 값을 0으로 만든다" 라고 적는다. 그러나 그 방어층이 D6/D7/D8 로 약해져 있으므로, **세 결함이 함께 남으면 카드 전체가 "스탬프만 갱신" 으로 퇴화할 수 있다** — 카드가 명시적으로 금지한 실패 형태다.
3. **기준선이 시간에 민감하다.** 다른 레인이 `origin/develop` 을 밀면 value 가 다시 오른다. plan.md §B.3 이 이를 인지하고 §C 재측정으로 흡수하지만, run 종료와 병합 사이의 창에서 재적색이 될 수 있다.
4. **`merge-base` 가 현재 HEAD 와 동일하다**(둘 다 `52f863f36`). 지금 시점에는 bare HEAD 스탬프와 merge-base 스탬프가 구분되지 않으므로 AC-CM2-009 가 물지 않는다. run 이 자체 커밋을 쌓는 순간 구분되므로 실질 위험은 낮으나, 실행자가 "지금 같으니 아무거나" 로 읽지 않도록 주의가 필요하다.
5. **`docs-truth.md` 는 REFRESH-001 시점에도 같은 6문서 주장 아래 있었고**(`SPEC-CODEMAPS-REFRESH-001/spec.md:81`, `acceptance.md:30` — 문구가 거의 동일), 그 SPEC 은 `completed` 다. 즉 D3 은 이 SPEC 이 새로 만든 결함이 아니라 **계승된 결함**이다. 지금 고치지 않으면 -003 에서 다시 나타난다.

---

## Recommendation

**FAIL — D1~D9 를 수리한 뒤 iteration 2 로 재감사한다.** must-pass 7항목은 전부 통과했고 인용 수치는 전수 재현되므로, 수리는 전면 재저작이 아니라 **열거된 델타**로 충분하다. 우선순위 순:

1. **D1** (critical) — 6단위 선정 규칙을 명명된 명령으로 재현 가능하게 만들고, `internal/template/agentemit` 을 포함하거나 `commandemit` 과 동일한 책임 질문으로 제외 근거를 쓴다. 접힘 정책은 건드리지 않는다.
2. **D2** — `plan.md` M2 의 `cp` 블록을 첫 단계로 올린다(한 줄 이동으로 복구 불가 손실 차단).
3. **D3** — `docs-truth.md` 의 손-유지 성격을 명시하고 갱신 단계와 증거를 분리한다.
4. **D4** — `acceptance.md:161` "3항목" → "2항목".
5. **D5** — §A.3(b) 컷오프를 명시하고 `internal/settings` 를 반영한다.
6. **D6** — `check_citations.go:23` 정규식(+ 구두점 절삭 + blockquote 면제)을 AC-CM2-006 의 추출 규약으로 명명하고, AC-CM2-008 의 "식별자" 를 추출 명령으로 정의한다.
7. **D7** — AC-CM2-005 에 `diff -u` 를 명명하고 "변경 없음" 행에 빈 diff 증거를 요구한다.
8. **D8** — AC-CM2-002 판정 행에 검사한 부모 산문의 위치·인용을 요구한다.
9. **D9** — 나머지 9개 MUST AC 에 4요소 RED-now 셀을 붙이고, RED-3 에 수출 파일 경로를 인용한다.

**운영자 결정은 어디에서도 재개하지 않았다.** 임계 40 / `gate.yaml` / `DefaultThresholds()` / 접힘 정책 / Go 코드는 전부 소관 밖으로 유지했고, 판별식은 §A.3(a1) 그대로 인용했으며 재정의하지 않았다. §A.4 의 `tree_root` 는 SPEC 의 판정대로 **결함이 아니다** — 나도 `check.go:341-346` 주석과 `:456`/`:557` 두 곳뿐인 비교를 직접 확인했고, 어떤 결함 항목도 여기서 파생시키지 않았다.

정책 규칙 적용 기록(`verification-completeness.md` § Policy-rule application evidence): 이 감사는 `.claude/rules/moai/development/verification-completeness.md` §1.1(빈 집합 위의 통과), §2(두 셀 채택 규율), §2.1(RED-now 4요소), 그리고 `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3·4 와 §2 를 적용해 판정했다.

---

## Addendum A — 리드 교차 측정에 대한 판정 (렌즈 3·4)

리드가 독립 측정 3건을 전달했다. 각각을 SPEC 본문에 대고 재측정해 판정한다. **판정 결과 신규 blocking 0건, 확언적 종결 1건, O1 정련 1건. 종합 판정(FAIL 0.74)은 변하지 않는다.**

### A.1 — 해소된 SHA 하드코딩 조건절 (신규 결함 없음)

리드의 조건절: *"If the SPEC hard-codes a resolved SHA instead of the command, that is worth a finding."* **그 전건은 거짓이다.** 재스탬프 리비전을 명명하는 네 자리 전부가 실행 시점 해석 명령 형식이다:

| 위치 | 형식 |
|---|---|
| `spec.md:200` (REQ-CM2-009) | `--commit <merge-base of HEAD and origin/develop>` |
| `plan.md:40` (§C pre-flight) | `git merge-base HEAD origin/develop` |
| `plan.md:123` (M4) | `REV="$(git merge-base HEAD origin/develop)"` |
| `acceptance.md:113` (AC-CM2-009) | `--commit <merge-base>` |

측정: `grep -nE '\b[0-9a-f]{40}\b'` 는 SPEC 3파일 전체에서 **단 한 곳**을 반환한다 — `spec.md:49` `commit_sha: 25a3212a93b4c811cbb22e3c0b34d43571fa65b4`. 이는 **교체 대상인 현재 앵커의 실측 판독**이지 스탬프 타깃이 아니다. 결함 없음.

### A.2 — O1 정련: /tmp 재판독은 의도를 검증하지 효과를 검증하지 않는다

리드의 *"resolved once vs re-evaluated at restamp time"* 지적은 새 결함을 만들지 않지만 **O1 을 더 날카롭게 만든다.** `plan.md:61` 의 §E 자체검증은 `git merge-base --is-ancestor "$(cat /tmp/t475-stamp-rev)" origin/develop` 로, M4 가 **찍으려 한** 리비전의 조상 성립을 검증한다 — **실제로 찍힌** 리비전이 아니다. 둘이 갈라지는 경로(스탬프 실패, 부분 쓰기, 다른 행위자의 개입)에서 이 검증은 조용히 통과한다.

`acceptance.md:114` 의 AC-CM2-009 는 `provenance.json` 의 `commit_sha` 가 `<REV>` 와 일치할 것을 요구하므로 인정 기준 층은 이 간극을 덮는다. plan.md §E 만 덮지 않는다.

Required fix(O1 갱신, 여전히 optional): `plan.md:61` 을 `/tmp` 대신 `provenance.json` 의 `commit_sha` 직독으로 바꾼다 — `--is-ancestor "$(jq -r .commit_sha .moai/project/codemaps/provenance.json)" origin/develop`. 그러면 §E 가 의도가 아니라 효과를 검증하고, `plan.md:52` / `acceptance.md:22` 의 "`/tmp` 는 판정 근거로 인정되지 않는다" 선언과의 긴장도 함께 사라진다.

### A.3 — value=0 기대는 명시돼 있고, 물며, 리드가 짚은 것보다 견고하다

리드는 *"whether the SPEC states that expectation, and whether any AC would still bite if it turned out false, is yours to judge"* 라고 남겼다. 판정: **명시돼 있고, 문다.** `acceptance.md:120` (AC-CM2-010) — "codemaps 행이 `verdict=fresh` 이고 value < 40(**기대 0**)". `verdict` 와 `value` 양쪽에 이분 판정이 걸려 있으므로 0 이 아니면 잡힌다.

리드의 추론(카드 커밋이 described roots 밖이므로 diff 0)은 코드 독해로 확인된다: `resolveContentAnchor` Rule A(`check.go:252-294`)는 워킹트리 body 가 스탬프 시점 body 와 다르면 앵커를 S 로 잡고, `internal cmd pkg` 에 대한 S↔워킹트리 diff 는 0 이다.

한 가지 덧붙인다 — **이 성질은 `origin/develop` 이 움직여도 유지된다**, 양방향 모두: (a) 레인이 `origin/develop` 을 흡수하면 merge-base 가 새 tip 으로 전진하고 diff 는 여전히 0, (b) 흡수하지 않으면 merge-base 가 `52f863f36` 에 머물고 diff 도 여전히 0. 따라서 Residual-risk #4(merge-base == HEAD)는 그대로 두되, Residual-risk #3 은 **run 창 내부가 아니라 병합 이후**에 대한 진술로 좁혀 읽어야 한다 — 카드가 develop 에 들어간 뒤 다른 레인의 described-root 커밋이 쌓이면 누적 지표가 다시 붉어지는데, 그것은 CADENCE-001 이 이미 판정한 메커니즘이지 이 SPEC 의 결함이 아니다.

### A.4 — 렌즈 4 확언적 종결: gate.yaml 을 만져야만 만족되는 AC 는 없다

리드가 준 좌표를 재측정했다 — `.moai/config/sections/gate.yaml:72-78`:

```yaml
  graph_freshness:
    enabled: true
    blocking: false
    codemaps_changed_files: 40
```

AC 12개 전수에 대해 "이 AC 를 만족시키는 유일한 경로가 gate.yaml 수정인가" 를 물었다. **하나도 없다.**

- **AC-CM2-010** 은 재스탬프로 만족된다(A.3). 임계를 올릴 필요가 없다.
- **AC-CM2-012** 는 관측 보고를 요구하면서 **설정 변경이 동반되면 FAIL** 이라고 명시한다(`acceptance.md:132`) — 방향이 반대다.
- **AC-CM2-011** 은 변경 집합을 허용 3경로로 제한하므로, gate.yaml 을 만지는 순간 **기계적으로 FAIL** 한다(`acceptance.md:126` 이 `gate.yaml` 을 이름으로 열거한다).

즉 운영자 결정 2(임계값·게이트 설정 불가침)는 산문 선언에 그치지 않고 **AC-CM2-011 이라는 기계적 경계로 집행된다.**

추가 관측 — **REQ-CM2-011 의 `Where` 절이 공허하지 않다.** `spec.md:204` 는 "**Where** the graph freshness gate runs advisory" 라는 역량 게이트로 조건화돼 있다. 그 전제는 `blocking: false` 이며, 방금 `gate.yaml:74` 에서 **참으로 측정**됐다. 전제가 거짓이면 REQ-CM2-011 전체가 공허해졌을 것이므로, 이 확인은 형식적이지 않다.

이 절은 본문 Recommendation 의 "(b) 스코프 누출 없음" 을 **측정으로 대체한다** — 종전 진술은 AC 전수 검토에 근거했으나 `gate.yaml` 좌표를 인용하지 않았다. 리드의 좌표가 그 간극을 메웠다.
