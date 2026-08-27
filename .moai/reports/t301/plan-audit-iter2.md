# SPEC 감사 보고서 (iteration 2): SPEC-FULL-SUITE-DOCTRINE-001 v0.2.0

Iteration: **2/2** (Tier M 상한 도달 — `harness.yaml:77` `plan_audit_tier_ceilings.M: 2` 실측)
Verdict: **FAIL**
Overall Score: **0.75** (iter-1 0.57 대비 **+0.18**) · Tier M PASS 기준선 **0.80** (`spec-workflow.md:141` 실측) — 미달
Skip-eligible: **아니오** (review 스트림 최종 판정이 FAIL이므로 run-gate 건너뛰기 자격 없음)

측정 트리: 워크트리 `.claude/worktrees/t301`, HEAD `d29b8942e`, branch `WT-full-suite-doctrine`. iter-1과 동일 트리.
저자 추론 맥락은 M1 Context Isolation에 따라 무시했다 (Reasoning context ignored per M1 Context Isolation). 저자가 진술한 조치 내용도 **주장으로만** 취급하고, 아래 모든 판정은 이 세션에서 직접 실행한 명령 출력에 귀속한다.

---

## 요약 — 무엇이 닫혔고 무엇이 남았나

iter-1의 차단 결함 9건 중 **7건이 완전히 닫혔고 2건이 부분적으로 닫혔다.** optional 2건은 근거를 갖춘 판단으로 처리됐고, 그중 D12 기각은 **정당하다**(독립 확인함).

그러나 iter-2에서 **새 차단 결함 1건**이 드러났다. 이것 하나로 FAIL이 확정되며, 점수와 무관하다: 같은 drift를 문자 그대로 담은 **배포 대상 사본이 하나 더 있는데 SPEC이 그 존재를 한 번도 언급하지 않는다.** 15개 AC 전부가 그 사본에 눈이 멀어 있어, 수리가 그 표면에서 **무효(inert)** 가 된다 — 이 SPEC이 §B2로 스스로 금지한 바로 그 실패 형태다.

---

## Must-Pass 결과 (7/7 통과)

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o '^- \*\*REQ-FSD-[0-9]*' spec.md | sort | uniq -c` → REQ-FSD-001~011 각 1회, 결번·중복 없음, 3자리 패딩 일관. 신설 REQ-FSD-007 삽입으로 인한 재번호가 순차성을 깨지 않았다.
- **[PASS] MP-2 GEARS 형식 준수 (요구사항 층)** — 11개 전부 GEARS. 001/004/005 Ubiquitous · 002/003/007/008/010 Unwanted(`shall not`) · 006/011 Event-driven · 009 Capability gate(`Where`). **iter-1의 경계 지적이 해소됐다**: 구 REQ-FSD-010(`its duration shall be recorded`, 수동태 산출물 주어)은 REQ-FSD-011에서 `the run-phase actor shall record …` 로 바뀌어 행위 주체가 문면에 들어왔다(`spec.md:107`).
- **[PASS] MP-3 YAML 프론트매터 유효성** — 실측 필드 13개: 정본 12개 + 선택 필드 `tier`(`spec.md:13`, 값 `M` — enum S\|M\|L 유효). snake_case 별칭 0건.
- **[N/A] MP-4 §22 언어 중립성** — 여전히 다중 언어 툴링 서술이 아니다. D11 부분 수용으로 지시문 층/예시 층을 명시 분리했고(`spec.md §D`), 16개 중 일부만 열거하는 새 서술을 만들지 않는다.
- **[PASS] MP-5 D7 교차 SPEC 정합** — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → 자기 참조 1건. retired/superseded/archived 0건.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c 'syscall'` → 4개 산출물 모두 `0`. D8-4 자동 통과.
- **[PASS] MP-7 clarification 게이트** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-FULL-SUITE-DOCTRINE-001/` → 매치 없음(rc=1).

부수 확인 — **Tier M 예산 준수**: REQ 11 / AC 15, 상한은 각 16(`spec-workflow.md:148-150`). 둘 다 통과하되 **AC는 상한까지 1칸** 남았다. 아래 권고가 AC를 추가하므로 이 여유를 의식해야 한다.

**FAIL은 M5 방화벽이 아니라 신규 차단 결함 N1과 집계 점수(0.75 < 0.80)에서 나온다.**

---

## 1. 전 AC 사전 baseline 재실행 결과

`acceptance.md` 의 **모든** AC 명령을 이 트리에서 직접 실행했다. 저자 기재값과 대조:

| AC | 유형 | 기재 baseline | 실측 | 판정 |
|---|---|---|---|---|
| AC-FSD-001 | 부재 | 3 | `3` | ✓ |
| AC-FSD-002 | 부재 | 3 | `3` | ✓ |
| AC-FSD-003 | 부재 | 각 1 | 로컬 `1` / 템플릿 `1` | ✓ |
| AC-FSD-004 | 부재 | 각 1 | 로컬 `1` / 템플릿 `1` | ✓ |
| AC-FSD-005 | 부재(행) | 각 1 | 로컬 `1` / 템플릿 `1` | ✓ |
| AC-FSD-006 | 부재(구간) | 각 1, 구간 4줄 | 로컬 `1` / 템플릿 `1`, 구간 **4줄** | ✓ |
| AC-FSD-007 | 존재(구간) | 각 0, 구간 **9줄** | 로컬 `0` / 템플릿 `0`, 구간 **10줄** | 값 ✓ / **줄수 불일치** (N2) |
| AC-FSD-008 | 동등 | `1a2 > isolation: worktree`, 무출력, 무출력 | 동일 | ✓ |
| AC-FSD-009 | 중립성 | diff 공백, 훑은 줄 `0` | `exit=0`, `grep -c '^+'` → `0`, 파일 0바이트 | ✓ |
| AC-FSD-012 | 존재 | 각 0 | 로컬 `0` / 템플릿 `0` | ✓ |
| AC-FSD-013 | 존재(구간) | 각 0 | 로컬 `0` / 템플릿 `0` | ✓ |
| AC-FSD-014 | 부재 | 각 3 | 로컬 `3` / 템플릿 `3` | ✓ |
| AC-FSD-015 | 부재(행) | 각 1 | 로컬 `1` / 템플릿 `1` | ✓ |

**형식 불량 패턴 0건. 사후 기대치와 같은 값에서 출발하는 항목 0건**(아래 두 예외는 유형상 정상):

- **AC-FSD-008**(동등형)은 사전과 사후 기대치가 같다. 불변 보존형이라 당연하며, "아무 작업도 안 함"을 잡는 것은 이 항목의 일이 아니다 — 부재·존재형이 담당한다. 결함 아님.
- **AC-FSD-010**(`make build` exit 0)도 무변경 상태에서 통과한다. 같은 이유로 결함 아님.
- **AC-FSD-009**는 사전에 성립하지 **않는다**(훑은 줄 0 < 하한 1). 저자가 `:229`에서 "이것이 의도한 형태"라고 명시했고, 실제로 그것이 D6를 막는 장치다. **정확히 옳은 설계다.**

또한 저자가 자진 신고한 iter-1 측정 오류(`| head` 를 통해 `rc` 를 읽어 `head` 의 종료 코드를 본 건)와 별개로, **파이프 없는 형태로 재확인**했다: `grep -rl 'origin/develop' internal/template/templates/` → 출력 없음, `rc=1`. `spec.md:77` 의 주장은 참이다.

---

## 2. 구간 추출 AC와 존재형 AC의 실제 동작

**AC-FSD-006 / AC-FSD-013 (`awk '/^# 1\./,/^# 2\./'`)** — 구간이 의도대로 닫힌다(실측 4줄). 그리고 이 두 항목은 **서로를 지킨다**: 구간 종료 패턴(`# 2.`)이 편집 중 깨져 범위가 EOF까지 달아나면, 범위 밖 예시 산문 `:65`·`:76`이 딸려 들어와 AC-FSD-006이 `2` 를 반환해 **RED**가 된다. 즉 AC-013의 baseline-0(2번 항목의 기존 `internal/<pkg>` 가 구간 밖이라 계수되지 않는다는 전제)이 조용히 무너질 수 없다. 저자가 `:172`에서 이 의존성을 인지하고 적어뒀는데, **실제로 짝이 그것을 기계적으로 지키고 있다.** 설계상의 강점으로 기록한다.

**AC-FSD-007 (`awk '/^### STEP 5/,/^### Checkpoint/'`)** — 구간은 열리고 닫히지만 **10줄**이다(로컬·템플릿 동일 실측). `acceptance.md:188` 이 기재한 "9줄" 과 어긋난다. awk 범위 연산자는 종료 패턴 줄을 **포함**하므로 `### Checkpoint and resume (both)` 헤딩까지 들어온다 — 실질 STEP 5 본문은 9줄, 추출은 10줄이다. 값 판정(각 `0`)에는 영향이 없다.

문제는 **이 줄 수가 Then에 들어 있지 않다**는 점이다. AC-006/013과 달리 AC-007에는 구간 폭주를 잡아줄 짝이 없다: 종료 헤딩이 편집으로 바뀌면 범위가 EOF까지 달아나고, 그러면 파일 어디에 있는 문구든 계수된다 — v0.1.0의 파일 전체 판정으로 조용히 되돌아간다. 위치 고정이 이 AC의 존재 이유인데, 그 고정을 보증하는 유일한 기록이 Then 밖에 있고 게다가 값이 틀렸다.

**AC-FSD-012 (`the tests the change can affect`)** — 사전 0히트 실측 확인. 상위 계약(`AGENTS.md:117`)이 이미 쓰는 문구라 배포 중립성과 충돌하지 않는다는 주장도 참이다.

---

## 3. D3 판정 — mutant 재구성 (질문 3)

**저자 주장**: 정본 문자열을 요구사항 본문에 고정했으므로 동의어 재작성이 "구조적으로" AC를 무효화한다.
**판정**: **좁은 형태에서만 참이다. 일반적으로는 거짓이다.**

**M-C (저자의 주장이 성립하는 형태 — 이제 잡힌다).** iter-1의 원래 mutant, 즉 위반 문장을 `Step 5 runs the complete suite regardless of project size.` 로 바꾸고 **다른 어디에도 정본 문구를 넣지 않는** 형태. AC-FSD-012가 `0` 을 반환해 **RED**. 저자의 조치가 실제로 이 구멍을 막았다.

**M-A (신규 mutant — 여전히 전 항목 통과).** AC-FSD-012는 **파일 단위 존재 판정(`grep -c … ≥1`)** 이지 지점 단위가 아니다. 따라서:

| 지점 | mutant가 쓰는 문면 |
|---|---|
| S1 | `Step 5 runs the complete suite regardless of project size.` |
| S2 | `run tests — the tests the change can affect` ← 정본 문구를 여기 한 번만 심는다 |
| S3 | 삭제 |
| S4 | `Issue the independent read-only verifications (tests, coverage, lint, boundary greps) …` |

결과 — AC-001 `0` ✓ / AC-002 `0` ✓ / AC-003 `0` ✓ / AC-012 `1` ✓ / AC-014 `0` ✓ (mutant 문면에 `LARGE_SCALE` 없음) / AC-007 (문장 심으면) ✓. **MUST 전부 초록인데 S1은 여전히 전량 실행을 처방한다.**

구멍의 크기는 iter-1보다 확실히 작아졌다(이제 mutant가 리터럴 4종을 피하면서 **동시에** 정본 문구를 심어야 한다). 그러나 닫히지는 않았다.

**값싼 마감**: 부재형에 형태족 패턴을 하나 더한다 — 예컨대 `(full|complete|entire|whole)[ -](test )?suite` 를 대소문자 무시로 세어 사후 `0`. 이 트리 사전 baseline은 비영이므로 RED-now가 확보된다. 형용사를 갈아끼우는 재작성 공간 대부분이 닫힌다.

**M-B (AC-FSD-007을 통과하는 mutant).** `grep -c -e A -e B` 는 **둘 중 하나**만 있어도 매치 줄을 센다. 직접 확인:

```
$ printf 'x integration branch y\nz nothing\n' | grep -c -e 'integration branch' -e 'PENDING at report time'
1
```

따라서 STEP 5에 `… delegated to the CI run on the project's integration branch.` 만 쓰고 `PENDING at report time` 을 **빠뜨려도** AC-FSD-007은 `1` 로 초록이다. 그런데 미결 상태를 말하지 않는 위임 문장은 보고 시점에 통과로 읽힌다 — **B5(조용한 삭제)가 정확히 그 형태로 되살아난다.** `acceptance.md:186` 의 "두 문구가 같은 문장에 있다" 는 여전히 눈 확인이며, `:189` 의 "위치·주체·미결 상태 셋 다 이 판정에 들어 있다" 는 주장은 기계 판정 층에서 성립하지 않는다.

**값싼 마감**: 두 토큰을 **각각** 세고 둘 다 `≥1` 을 요구한다(명령 2줄이면 된다).

---

## 4. `LARGE_SCALE` 제거의 안전성 (질문 4) — 여기서 N1이 나왔다

토큰 자체의 의존성은 **없다.** 리포 전역 검색 결과 `LARGE_SCALE` 를 담은 파일은 SPEC 산출물·감사 기록을 빼면 세 개뿐이고, 어느 Go 코드·훅·CI 워크플로도 이 토큰을 읽지 않는다. `.moai/specs/SPEC-AGENT-PARALLEL-OPT-001/progress.md:195` 의 언급은 과거 통합 기록이지 의존이 아니다. `spec.md §A.2` 의 "유일한 귀결은 타깃 실행으로의 전환" 도 세 줄을 읽어 확인했다. **제거 결정 자체는 안전하고, D4는 옳게 닫혔다.**

그런데 그 검색이 **세 번째 파일**을 드러냈다.

---

## 신규 차단 결함

**N1. 같은 drift를 문자 그대로 담은 세 번째 배포 사본을 SPEC이 인지하지 못한다 — `internal/template/templates/.codex/agents/moai/manager-develop.toml` — Severity: critical — Class: blocking**

같은 에이전트 정의의 **Codex용 사본**이 배포 템플릿 트리 안에 있다. 네 위반 지점이 전부, 문자 그대로 들어 있다:

```
:80   … LARGE_SCALE = test files > 500 … Step 5 always runs the full suite regardless of scale.     ← S1
:114  3. **Verify behavior**: run tests — targeted when `ddd` LARGE_SCALE, otherwise the full suite  ← S2
:120  - Run the COMPLETE test suite (always full, regardless of LARGE_SCALE; …)                      ← S3
:123  - Issue the independent read-only verifications (full suite, coverage, lint, boundary greps) …  ← S4
```

측정 근거:

- AC-FSD-001/002의 패턴 집합을 이 파일에 걸면 **`3`** — 즉 이 파일은 지금 `manager-develop.md` 와 **같은 상태**다.
- `grep -rln 'always runs the full suite'` 와 `grep -rln 'full suite, coverage'` 둘 다 이 `.toml` 을 포함해 세 파일을 반환한다(`.md` 로컬·템플릿 + `.toml`).
- 배포 여부: `internal/template/embed.go:28` `//go:embed all:templates` — `.codex/` 는 `templates/` 하위이므로 바이너리에 임베드돼 **배포된다.**
- 형제 11개(`builder-harness.toml` … `sync-auditor.toml`)가 함께 있는 정식 미러 트리다.
- SPEC의 인지 여부: `grep -rn 'codex' .moai/specs/SPEC-FULL-SUITE-DOCTRINE-001/` → **매치 없음(rc=1)**. 네 산출물 어디에도 언급이 없다 — 수리 대상으로도, 범위 밖 선언으로도 없다.

귀결이 무겁다. **MUST AC 15개 전부가 이 파일에 눈이 멀어 있다**: AC-001/002/003/012/014는 `.md` 두 경로를 하드코딩하고, AC-FSD-008의 델타 판정은 로컬↔템플릿 **쌍** 구조를 전제하는데 `.codex/` 에는 로컬 사본이 아예 없다(`ls .codex/agents/moai/manager-develop.toml` → No such file or directory). AC-FSD-009조차 이 파일을 훑기는 하지만 **추가된 줄의 지역 내용만** 보므로, 손대지 않은 파일은 무사통과한다.

따라서 "`.md` 두 사본만 고치고 `.toml` 을 그대로 둔다" 는 수리가 **15개 AC를 전부 초록으로 통과하면서 drift를 배포한다.** 이것은 `plan.md §B2` 가 "무효 수리" 로 이름 붙인 실패 형태이며, `§A.3` 이 배치 정의를 범위에 넣은 논리 — "한쪽을 두면 수리가 inert 해진다" — 와 정확히 같은 논리가 여기에도 적용되는데 적용되지 않았다.

`spec.md §A.1` 의 제목 "측정된 위반 지점 — 4곳" 도 이 발견 앞에서 부정확해진다. 사본 단위로는 3개, 지시 지점 단위로는 12개(4 × 3)다.

**범위 한정 사실 하나**: 배치 정의 두 파일에는 codex 미러가 **없다**(`ls internal/template/templates/.codex/rules` → No such file or directory). 따라서 N1은 에이전트 정의 한 축에 국한된다. `.toml` 자체도 배치 정의를 인라인으로 품고 있지 않다(`grep -c 'go test'` → `0`).

**필요한 수정** — 둘 중 하나를 **명시적으로** 택한다:

- **(가) 범위에 넣는다.** `spec.md §A.1` 표에 세 번째 사본 열을 추가하고, REQ-FSD-002의 "four measured sites" 를 사본 3개 기준으로 재서술하며, AC-FSD-001/002/012/014에 `.toml` 경로를 인자로 추가한다(새 AC 없이 기존 항목의 파일 목록 확장만으로 가능하므로 AC 예산 15/16을 건드리지 않는다). 사전 baseline은 이 트리에서 `3`·`0`·`3` 로 이미 실측돼 있다.
- **(나) 범위 밖으로 선언한다.** `§D` 에 근거를 적은 Out of Scope 소절을 신설한다. 다만 이 경로를 택하면 "배포되는 drift를 알면서 남긴다" 를 문서가 감당해야 하고, `§A.3` 이 배치 정의에 적용한 inert 논리와의 비대칭을 설명해야 한다.

**무언(無言)만이 허용되지 않는다.** 지금 상태는 판단이 내려지지 않은 것이지 범위가 좁혀진 것이 아니다.

---

**N2. AC-FSD-007의 구간 폭주를 막는 장치가 Then 밖에 있고, 그 기록값이 틀렸다 — `acceptance.md:188` — Severity: minor — Class: blocking**

실측 10줄인데 `9줄` 로 적혀 있다(로컬·템플릿 동일). 값 판정에는 영향이 없으나, 이 줄 수가 **위치 고정을 보증하는 유일한 기록**인데 Then에 들어 있지 않고 게다가 부정확하다. AC-006/013이 짝을 통해 폭주를 기계적으로 잡는 것과 대비된다(§2 참조).

**필요한 수정**: 추출 줄 수를 Then에 넣는다(정확히 10, 또는 상·하한). 한 줄이면 된다.

---

## iter-1 결함 처리 현황

| iter-1 결함 | 상태 | 근거 (이 세션 실측) |
|---|---|---|
| D1 브랜치명 유출 | **닫힘** | REQ-FSD-006이 `integration branch` 로 재작성(`spec.md:96`), REQ-FSD-007 신설(`:97`), `origin/develop` 은 `§A.4` 배경에만 존재. AC-FSD-009 패턴에 `origin/develop`·`origin/main` 추가(`acceptance.md:223`). `grep -rl 'origin/develop' internal/template/templates/` → rc=1 (파이프 없이 재확인) |
| D2 AC-006 지연·오기대 | **닫힘** | 구간 판정으로 재작성(`acceptance.md:99-100`), 구간 4줄·전량 호출 1 실측. 위임 문장 삭제됨. `:65`·`:76` 배제 근거가 AC와 `spec.md §A.3`·§D 양쪽에 기록 |
| D3 부재형 전용 | **부분** | AC-012·013 신설, 사전 0히트 실측 확인. 그러나 파일 단위 존재 판정이라 지점 단위 mutant M-A가 생존 — §3 |
| D4 dangling 무방비 | **닫힘** | 판별자 제거로 확정(`spec.md §A.2`), REQ-FSD-003 Unwanted 재작성, AC-FSD-014 신설(baseline 3 실측). plan.md의 존치 권고 삭제 확인. 토큰 의존성 리포 전역 검색으로 없음 확인 — §4 |
| D5 시간 추정치 무AC | **닫힘** | AC-FSD-015 신설, 행 한정 판정, baseline 각 1 실측 |
| D6 unstaged 전용 diff | **닫힘** | 기준 SHA 고정 + 훑은 줄 하한. `git diff <SHA> -- path` 는 커밋·스테이지·워킹트리 변경을 모두 본다. 하한이 공허 통과를 막는 것도 baseline(훑은 줄 0)이 사전에 성립하지 않음으로 확인 |
| D7 `tier:` 부재 | **닫힘** | `spec.md:13` `tier: M`. plan.md:13·progress.md:5와 일치, 모순 없음 (질문 6) |
| D8 report-not-verdict | **닫힘** | AC-005가 행 추출 후 카운트로 재작성(`acceptance.md:84-85`), AC-009 첫 블록 삭제 확인 |
| D9 AC-007 얕음 | **부분** | 구간 고정은 실효(§2). 그러나 `-e A -e B ≥1` 이 **둘 다**를 요구하지 않아 mutant M-B 생존 — §3 |
| D11 언어 중립성 (optional) | **처리됨** | 지시문 층/예시 층 분리를 `spec.md §D` 에 명문화. 근거 있는 판단 |
| D12 REQ 린트 (optional) | **기각 — 정당** | `internal/spec/lint.go:447` 이 `\d{3}-\d{3}` 두 그룹 꼬리를 요구함을 독립 확인. `REQ-FSD-001` 은 표기를 어떻게 바꿔도 매치 불가. 표기 변경은 GEARS 라벨 가독성만 잃고 안전망은 못 얻는다. 공허 린트 사실을 `progress.md:45-52` 에 기록한 처리가 옳다 |

**질문 5 (재번호 잔재)** — 네 산출물 전체에서 `REQ-FSD-*` 참조를 수집해 대조했다. 구 번호를 가리키는 참조는 **0건**이다. `acceptance.md §D.2` 는 REQ-FSD-007 → AC-FSD-009(중립성), REQ-FSD-008 → AC-FSD-007(존재), REQ-FSD-011 → AC-FSD-011(관측)로 신 번호와 정합한다. `plan.md` M6도 REQ-FSD-011을 가리킨다. `progress.md:31` 의 "REQ-FSD-007 신설" 도 신 번호다.

**질문 6 (tier 충돌)** — 없음. 세 곳 모두 M으로 일치하며, Tier M 산출물 집합(3종 + progress) 및 REQ/AC 예산(11/16, 15/16)과도 모순이 없다.

---

## 항목별 점수 (루브릭 앵커)

| 차원 | iter-1 | iter-2 | 밴드 | 근거 |
|---|---|---|---|---|
| Clarity | 0.50 | **0.75** | 0.75 | 요구사항 간 모순(D1) 소멸, AC-006 범위 모호성 소멸, C6가 계측점을 문자로 고정. 남은 모호성 1건: REQ-FSD-009의 "a mirror under `internal/template/templates/`" 가 `.codex/` 를 포함하는지 불명 — 그 모호성이 N1의 원인이다 |
| Completeness | 0.75 | **0.75** | 0.75 | 구조·프론트매터는 오히려 개선(`tier` 추가, Out of Scope H3 6개). 그러나 `§A.1` 의 위반 지점 열거가 배포 사본 하나를 통째로 빠뜨린다(N1) — 구조적 완비와 별개인 실질 누락 |
| Testability | 0.50 | **0.75** | 0.75 | AC-005 자기완결화, AC-006/013 구간 한정 + 상호 보증, AC-014/015 신설, AC-009 SHA 고정 + 훑은 줄 하한. 남은 비이진 항목 1개: AC-FSD-007의 "두 문구가 같은 문장에" 가 여전히 눈 확인이고 카운트가 둘 다를 요구하지 않는다 |
| Traceability | 0.60 | **0.75** | 0.75 | `§D.2` 가 절 단위로 분리돼 공허 매핑 대부분 해소(REQ-004·005 각각 2 AC). 구 번호 잔재 0건. 남은 공허 매핑 1건: REQ-FSD-008(조용한 생략 금지) → AC-FSD-007인데, 그 AC는 `PENDING` 생략을 구별하지 못한다 |

조화평균: `4 / (1/0.75 × 4)` = **0.75**. **Δ = +0.18** (0.57 → 0.75). 점수 회귀 없음 — LEAN의 STOP-on-regression 조항은 발동하지 않는다.

---

## 남은 결함 목록

| ID | 요지 | Severity | Class |
|---|---|---|---|
| **N1** | `.codex/agents/moai/manager-develop.toml` — 네 위반 지점을 문자 그대로 담은 세 번째 **배포** 사본을 SPEC이 수리 대상으로도 범위 밖으로도 선언하지 않음. 15개 AC 전부가 이 파일에 눈이 멂 → 수리가 그 표면에서 무효 | critical | blocking |
| **D3'** | AC-FSD-012가 파일 단위 존재 판정이라 지점 단위 동의어 mutant(M-A)가 MUST 전 항목을 통과 | major | blocking |
| **D9'** | AC-FSD-007의 `-e A -e B ≥1` 이 두 토큰 **모두**를 요구하지 않아, `PENDING at report time` 누락 mutant(M-B)가 통과 → B5 재발 경로 | major | blocking |
| **N2** | AC-FSD-007 구간 줄 수가 Then 밖에 있고 기록값이 틀림(기재 9 / 실측 10). 구간 폭주 시 v0.1.0의 파일 전체 판정으로 조용히 회귀 | minor | blocking |

optional 잔여 없음 — D11·D12는 근거 있는 판단으로 종결됐다.

---

## 권고 (manager-spec, 우선순위 순)

1. **N1을 명시적으로 처분한다.** 범위에 넣는다면 `spec.md §A.1` 표에 세 번째 사본을 추가하고 AC-FSD-001·002·012·014의 파일 인자에 `.toml` 경로를 더한다 — **새 AC 없이** 가능하므로 AC 예산 15/16을 넘지 않는다. 사전 baseline은 이 보고서에 실측돼 있다(`3` / `0` / `3`). 범위 밖으로 둔다면 `§D` 에 근거를 적은 소절을 신설하고, `§A.3` 의 inert 논리와의 비대칭을 설명한다.
2. **AC-FSD-007을 토큰별 판정 2줄로 나눈다** — `integration branch` 와 `PENDING at report time` 을 각각 세어 둘 다 `≥1`. (D9')
3. **부재형에 형태족 패턴을 하나 추가한다** — `(full|complete|entire|whole)[ -](test )?suite` 대소문자 무시, 사후 `0`. 기존 AC-FSD-001·002의 `-e` 목록에 얹으면 새 AC가 필요 없다. (D3')
4. **AC-FSD-007의 추출 줄 수를 Then에 넣는다** — 정확값 10(로컬·템플릿 동일 실측). `acceptance.md:188` 의 `9줄` 도 정정. (N2)

네 건 모두 **문면 편집이며 새 조사가 필요 없다.** 필요한 측정은 전부 이 보고서에 실려 있다.

---

## 반복 상한 · 다음 행동

Tier M 상한 2회에 도달했다(`harness.yaml:77`). 점수는 회귀하지 않고 +0.18 개선됐으므로 STOP-on-regression 조항은 발동하지 않으며, 남은 결함 4건은 전부 좌표와 기대치가 확정된 국소 편집이다. 오케스트레이터는 상한 도달 절차에 따라 운영자에게 세 갈래를 제시해야 한다:

1. **iter-3으로 상한 연장** — 남은 델타가 작고 전부 열거돼 있어 회수 가능성이 높다. **이 감사의 권고 경로다.**
2. **PASS-with-debt** — 권장하지 않는다. N1은 배포되는 drift를 알면서 남기는 것이고, 그 상태에서 run-phase가 15개 AC를 전부 통과시키면 완료 보고가 "수리 완료" 로 읽힌다 — VCI §1.1 surface 2의 관측되지 않은 완료 주장이 SPEC 설계 단계에서 예약되는 셈이다.
3. **scope-reduction** — N1을 별도 카드로 분리하고 이 SPEC은 `.md` 두 사본에 한정한다. 택한다면 `§D` 의 범위 밖 선언이 **필수**다(권고 1의 (나)).

---

## Gaps (관측하지 않은 것)

- 대체 문면은 여전히 존재하지 않는다(run-phase 산출). §3의 mutant 판정은 **AC 집합의 판별력**에 대한 것이지 아직 쓰이지 않은 텍스트에 대한 것이 아니다.
- 전체 테스트 스위트를 포함해 어떤 테스트도 실행하지 않았다.
- `.codex/` 트리의 나머지 10개 `.toml` 이 각자의 `.md` 원본과 어떤 동기화 규율로 묶이는지는 조사하지 않았다. N1의 판정에 필요하지 않았고, 그 규율 자체는 이 카드 소관이 아니다. (`plan-auditor.toml` 이 `go test ./...` 를 담고 있다는 사실만 관측했고, 이 SPEC의 대상 파일이 아니므로 판정하지 않았다.)
- `.claude/skills/` 층의 같은 형태 지시(`spec.md §D` 가 범위 밖으로 선언)는 세지 않았다.
- 어떤 SPEC 산출물·에이전트 정의·규칙·템플릿도 수정하지 않았다.

**감사 산출물**: `/Users/goos/MoAI/moai-adk-go/.moai/reports/t301/plan-audit-iter2.md` (iter-1 보고서 `plan-audit.md` 는 그대로 둠)
**측정 트리**: `.claude/worktrees/t301` @ `d29b8942e`
