---
id: SPEC-ARTIFACT-STATELESS-001
title: "비-spec.md SPEC 산출물 무상태 확정 — 규약 명문화 · 재발 방지 lint · status 라인 정리"
version: "0.3.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: ".claude/rules/moai/development, internal/spec, .moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, frontmatter, statelessness, corpus-cleanup, recurrence-prevention"
tier: M
related_specs: [SPEC-PHASE-FRONTMATTER-OWNER-001]
---

# SPEC-ARTIFACT-STATELESS-001 — 비-spec.md SPEC 산출물 무상태 확정

## HISTORY

| 날짜 | 버전 | 변경 |
|---|---|---|
| 2026-08-28 | 0.3.0 | plan-auditor 2차 **PASS-WITH-DEBT 0.92** (네 차원 단조 상승, 1차 결함 10건 중 9건 해소) 대응. **(N1 — 개정이 들여온 결함) 기준값 슬롯의 3중 가드.** `[0-9a-f]*`가 0회 반복을 허용해 **빈 슬롯 행에 매치하고 빈 문자열을 캡처**했고, `""..HEAD`가 `HEAD..HEAD`로 조용히 해석돼 `AC-08`·`-10`이 아무것도 검사하지 않고 PASS를 출력했다(1차의 `BASE=<…>` 자리표시자는 셸 리디렉션으로 요란하게 죽었으므로 **실패 모드가 악화**된 것이다). 세 AC 전부에 `\{7,\}` + `-n` 검사 + `git rev-parse --verify` 3중 가드를 넣고, `AC-07`의 둘째 조건(`총량 == baseline N`)을 **산문에서 명령으로 승격**했으며, `AC-08`·`-10`에 「정리 대상이 범위 안에 있는가」 비공허성 가드를 추가했다. baseline N 슬롯을 `progress.md` 표에 신설. **(R1 완화, 완전 해결 불가) 리터럴 앵커의 부정문 맹점** — 감사관이 네 리터럴을 전부 담되 의미를 뒤집은 소절로 전부 PASS시켰다. 같은 줄 부정어 근접 검사 + DoD 인간 판독 항목 + `REQ-AST-001-016`을 넣고, **잔여를 §5에 부채로 명시**했다(여러 줄 부정은 기계로 닫히지 않는다). **(N2) 소절 추출 종료 조건**을 `/^##/` → `/^#{1,3} /`로 좁혀 소절 안 `####` 하위 제목에서 끊기던 오탐 FAIL을 제거 — M1 작성자에게 암묵적 제약을 지우지 않는다. **(N3)** `plan.md` §B1의 종결 모집단 쌍(362→417)을 채택 모집단 기준으로 재기술. **(N4)** `REQ-AST-001-001`의 중첩 백틱을 코드펜스 「고정 리터럴 3종」 블록으로 대체. REQ 15 → 16, AC 11 (Tier M 상한 각 16). |
| 2026-08-28 | 0.2.0 | plan-auditor 1차 **PASS-WITH-DEBT 0.86** 대응, blocking 6건(D1~D6) + optional 1건(D9). **(D6) 모집단을 전 코퍼스 696(389 파일)으로 확정** — 무상태 선언이 Tier에 의존하지 않는 것과 같은 이유로 라이프사이클 status에도 의존하지 않으며, 종결로 좁히면 status 축에 같은 모양의 구멍이 생긴다(§1.6). 362는 389의 종결-한정 부분집합으로 병기해 추적성을 유지한다. **(D4) 소급 편집 구분을 논증으로 승격**(§1.5) — 백필은 기록된 값을 바꾸고 D1 정리는 기록이 된 적 없는 필드를 지운다. **(D1·D3) `AC-AST-001-01`·`-02`를 신설 앵커 `### Artifact Statelessness` 범위로 한정**하고 세 문장을 각각 별도 판정으로 분해 — 종전 판정은 미수정 파일에서 7개 중 6개가 이미 참이었고, 부정 검사는 선언한 PASS 문자열이 원리상 출력될 수 없었다. **(D2) `AC-AST-001-04`의 lint 호출을 심는 SPEC 하나로 한정** — 종전 전 코퍼스 호출은 M2 선행 시 정상 lint에도 FAIL했다. **(D5) 교차검증 주장 한정**(§1.6) — 두 스크립트가 `fm_of`를 공유하므로 검증된 것은 집계 로직뿐이다. **(D9) 템플릿 미러를 조건절에서 사실로**(REQ-AST-001-015 + `AC-AST-001-11`) — 실측 결과 실재하며 바이트 동일. REQ 13 → 15, AC 10 → 11. |
| 2026-08-28 | 0.1.0 | 최초 작성. 카드 t357의 두 전제(「규약이 4종 명시」·「Tier L 6종」)가 실측으로 뒤집힌 뒤, 운영자가 **안 C(무상태 선언)** 를 선택한 결과를 명세한다. 정리 정의는 **D1(status 라인만 제거)** 로 못박고, 재발 방지 lint를 이 SPEC의 본체로 둔다. |

## §1 배경

### §1.1 카드 전제 두 가지가 실측과 어긋났다

카드 t357은 "상태 전이 규약이 4종(spec/plan/acceptance/progress)만 명시하고 Tier L은 6종이라 design.md·research.md가 규약 밖에 남는다"고 적었다. 두 전제 모두 측정으로 뒤집혔다.

**측정 트리**: 이 워크트리, HEAD `c6aa61346` (= 측정 시점 `origin/develop`, 발산 `0 0`). 코퍼스 `.moai/specs/SPEC-*` **696개**. 전체 실측 보고: `.moai/reports/t357/plan-measurement.md`, 원자료 `t357_rows.tsv`·`t357_closed.tsv`·`t357_audit.txt`, 스크립트 `t357_measure.sh`.

| 카드 전제 | 실측 |
|---|---|
| 규약이 4종을 명시 | 규약은 **spec.md 1종만** 명시한다. `plan.md`·`acceptance.md`도 `design.md`·`research.md`와 똑같이 규약 밖이다 |
| Tier L은 6종 | 규약상 Tier L은 5종(spec/plan/acceptance/design/research). `progress.md`는 별도 진행 기록 |

따라서 이 카드가 정하는 것은 "4종 → 6종 확장"이 아니라 **"1종 유지냐, N종 확장이냐"** 다.

### §1.2 규약은 spec.md 1종만 구속한다

`.claude/rules/moai/development/spec-frontmatter-schema.md:13`:

> All SPEC documents (`spec.md`) MUST contain exactly these 12 fields in YAML frontmatter.

의무 대상이 **괄호로 `spec.md`까지 박혀** 있다. 나머지 산출물의 frontmatter는 스키마에 정의된 적이 없는 필드다. plan 워크플로도 같다 — `.claude/skills/moai/workflows/plan/spec-assembly.md:72,80,313,519`의 frontmatter 지시는 전부 spec.md 대상이고, 다른 산출물에 frontmatter를 쓰라는 지시는 0건이다.

### §1.3 검사도 정확히 1종만 본다 — 규약과 어긋나 있지 않다

| 지점 | 읽는 파일 | 근거 |
|---|---|---|
| `moai spec audit` | spec.md | `internal/spec/audit.go:347` `filepath.Join(specDir, "spec.md")` |
| lint 대상 발견 | spec.md | `internal/spec/lint.go:307,328` `discoverSPECs` — `SPEC-*/spec.md` 패턴 |
| 전이 소유권 lint | spec.md | `internal/spec/lint_ownership.go:11` — spec.md status 라인 변화만 추적 |
| `moai spec close` **쓰기** | spec.md + progress.md | `internal/spec/closer.go:312,335` — `os.WriteFile` 2곳뿐 |

`closer.go:627`이 acceptance.md를 읽긴 하나 **AC PASS 마커 판독용**이고 frontmatter는 건드리지 않는다.

`moai spec audit` 실측(트리 `c6aa61346`, 696 SPEC): `rc=0`, `[INFO] 481 / [WARN] 0 / [ERROR] 0`, `draft|design.md|research.md` 매치 **0건** (`.moai/reports/t357/t357_audit.txt`).

**결론**: 침묵은 검사 누락이 아니다. **규약이 그 파일에 아무 의무도 지우지 않기 때문**이다. `design.md`가 `draft`로 남은 것은 규약 위반이 아니라, 규약에 없는 필드를 에이전트가 임의로 붙였다가 아무도 옮기지 않은 것이다.

### §1.4 지배적 형태는 `draft`가 아니라 frontmatter 부재다

종결(closed = completed + implemented) SPEC **633**개 모집단, 트리 `c6aa61346`:

| 상태 | design | research | plan | acceptance |
|---|---:|---:|---:|---:|
| 파일 없음 | 527 | 470 | 51 | 74 |
| **frontmatter 없음** | **81** | **131** | **402** | **379** |
| status 필드 없음 | 2 | 3 | 25 | 25 |
| draft | 12 | 12 | 44 | 45 |
| completed | 9 | 9 | 61 | 57 |
| implemented | 0 | 0 | 28 | 26 |
| in-progress | 2 | 3 | 12 | 8 |
| 기타 | 0 | 8 | 10 | 19 |

카드가 지목한 `draft` 잔류는 design 12 / research 12 로 **소수 패턴**이다. 그리고 카드가 "규약 안"이라고 가정한 `plan.md`·`acceptance.md`도 402/379가 frontmatter 부재로, design/research와 같은 상태다.

### §1.5 왜 안 C인가 — 운영자 결정과 그 근거

운영자는 세 안 중 **안 C(무상태 선언)** 를 선택했다.

| | 안 A 전수 백필 | 안 B 시점 이후만 | **안 C 무상태 선언** |
|---|---|---|---|
| 해당 종결 SPEC | 544 | 0 | **170** |
| 실제 편집 파일 | 1,251 | 0 | **362** (D1 기준) |

> 이 표는 **운영자에게 제시됐던 그대로**이며 모집단이 종결 SPEC 633이다. 이후 §1.6이 모집단을 전 코퍼스 696으로 확대했으므로 안 C의 실제 편집 파일은 **389**다(362는 그 부분집합). 표를 사후에 고쳐 쓰지 않고 이 주석으로 관계를 남긴다 — 결정 시점의 근거가 무엇이었는지가 기록으로서 의미를 갖는다.

#### 소급 편집 — 안 A를 물린 근거가 안 C에는 적용되지 않는 이유

운영자가 안 A를 물린 근거 하나는 **소급 편집(이력 왜곡)** 이었다. 안 C도 닫힌 SPEC의 파일을 편집한다. 두 행위가 왜 다른지를 전제로 두지 않고 여기서 논증한다.

**두 행위는 축이 다르다.**

- **안 A의 백필은 기록된 값을 다른 값으로 바꾼다.** `design.md`의 `status: draft`를 `status: completed`로 옮기는 것은, 그 SPEC이 닫히던 시점에 그 파일이 무엇이었는지에 대한 **주장을 사후에 다시 쓰는** 일이다. 그 주장이 옳은지 그른지와 무관하게, 그것은 기록의 내용을 변경한다.
- **안 C의 D1 정리는 규약이 정의한 적 없는 필드를 지운다.** §1.3이 실측한 대로 그 필드는 규약이 요구한 적이 없고, 검사가 읽은 적도 없으며(`audit.go:347`·`lint.go:307,328`은 `spec.md`만 본다), `closer.go:312,335`가 쓴 적도 없다. §1.2가 인용한 규약 문언(`spec-frontmatter-schema.md:13`)은 의무 대상을 `spec.md`로 괄호까지 박아 놓았다.

따라서 지워지는 것은 **애초에 기록으로서의 지위를 가진 적이 없다** — §1.3의 표현대로 "규약 위반이 아니라, 규약에 없는 필드를 에이전트가 임의로 붙였다가 아무도 옮기지 않은 것"이다. 규약이 읽지 않고 검사가 보지 않으며 close 절차가 쓰지 않는 필드는, 그 SPEC이 닫히던 시점에 대해 아무것도 주장하지 않는다. 그것을 지우는 것은 이력의 삭제가 아니라 **즉흥의 잔여물 제거**다.

**뒤집어 말하면 이렇다**: 안 A는 기록이 말하는 바를 바꾸므로 소급 편집의 반대에 걸리고, 안 C는 기록이 아닌 것을 치우므로 걸리지 않는다. 이 구분이 성립하지 않는 유일한 경우는 그 필드가 실제로 무언가를 기록하고 있었을 때인데, §1.3의 4개 실측 좌표가 그렇지 않음을 보인다.

안 C가 닫는 구멍 하나를 기록해 둔다: **종결 SPEC 106건이 `design.md`/`research.md`를 보유하면서 `tier: L`이 아니다.** 규약을 "Tier L 한정"으로 좁혀 쓰는 안 A/B 형태였다면 이 106건은 규약 밖에 그대로 남는다. 안 C는 **Tier와 무관하게** 모든 비-spec.md 산출물을 규약 밖으로 확정하므로 이 106건도 빠짐없이 규칙 안에 들어온다. 이것이 선택된 안이 견고한 이유이며, 명세에 남겨야 할 사실이다.

### §1.6 정리 정의는 D1으로, 모집단은 전 코퍼스로 못박는다

두 정의를 모두 실측했다. **모집단이 두 가지라 수치도 두 벌**이라는 점을 먼저 분리해 적는다.

**모집단 = 전 코퍼스 696 SPEC** (이 SPEC이 채택하는 모집단, 트리 `3b1830b96`):

| 정의 | design | research | plan | acceptance | **합** |
|---|---:|---:|---:|---:|---:|
| **D1 — `status:` 라인만 제거** | 27 | 34 | 164 | 164 | **389** |
| D2 — frontmatter 블록 통째 제거 | — | — | — | — | (미측정; D2 미채택) |

```bash
bash .moai/reports/t357/t357_d1_by_artifact.sh .
# HEAD=3b1830b96 / design 27 / research 34 / plan 164 / acceptance 164 / total 389
```

**모집단 = 종결(closed = completed + implemented) SPEC 633** (참고값, 트리 `c6aa61346`):

| 정의 | design | research | plan | acceptance | **합** |
|---|---:|---:|---:|---:|---:|
| **D1** | 23 | 29 | 155 | 155 | **362** |
| D2 | 25 | 32 | 180 | 180 | 417 |

```bash
awk -F'\t' '$5=="yes"{d1[$2]++; t++} END{for(k in d1) print k, d1[k]; print "total", t}' \
  .moai/reports/t357/t357_fmrows.tsv
```

**362는 389의 종결-한정 부분집합이다.** 운영자에게 제시됐던 수치가 362이므로 폐기하지 않고 이 관계로 남긴다 — 차분 27건은 미종결 SPEC(draft·in-progress·archived·superseded·rejected)에 있다. 두 모집단을 같은 명령으로 한 번에 재측정한 결과:

```bash
bash .moai/reports/t357/t357_d1_all.sh .
# HEAD=3b1830b96 / D1 전체 696 모집단 = 389 / D1 종결(633) 모집단 = 362
```

#### 채택 — 정의는 D1, 모집단은 전 코퍼스 696

**정의 D1을 채택한다.** 근거: 무상태 선언이 지배하는 것은 **status 축**이다. `id`·`title`·`version`·`created`는 이 카드가 아무것도 판정하지 않은 **다른 축**이며, 그것까지 제거하는 것은 범위 확대다.

D2에만 걸리는 차분(종결 모집단 기준 55개 파일)의 성격도 실측된다 — 그 블록들은 **spec.md frontmatter의 복제본**이다. 필드 빈도는 `version` 371 · `status` 362 · `id` 356 · `created` 340 · `title` 289이고, D2 블록 417개 중 **224개가 12~14 필드**를 담고 있다. 즉 D2는 "부수 필드 정리"가 아니라 복제된 스키마 블록 전체를 지우는 별개의 결정이다.

**모집단은 전 코퍼스 696을 채택한다.** 근거는 §1.5의 논거와 같은 형태다: **무상태 선언은 무조건적이다.** 그것은 Tier에 의존하지 않고(§1.5가 106건으로 논증한 그 축), 라이프사이클 status에도 의존하지 않는다. 정리를 종결 여부로 좁히면 status 축에 **같은 모양의 구멍**을 새로 낸다 — 오늘 `in-progress`인 SPEC이 내일 닫히면서 규약 밖 필드를 그대로 달고 들어온다. 규칙의 조건이 하나면 정리의 조건도 하나여야 한다.

#### 교차검증의 범위 — 무엇이 검증됐고 무엇이 안 됐는가

D1의 362는 두 경로(`t357_rows.tsv` 재집계, `t357_fmrows.tsv` 집계)에서 같은 값이 나왔다. 다만 두 스크립트는 **동일한 `fm_of` awk 추출기를 공유한다**(`NR==1&&/^---/{f=1;next} f&&/^---/{exit} f`). 따라서 이 일치가 검증하는 것은 **집계 로직**이고, §5가 기록한 추출 단계의 공통 실패 모드(1행이 `---`가 아닌 파일을 전부 "frontmatter 없음"으로 계상)는 **두 경로에 동일하게 작용하므로 교차검증되지 않는다.** 계수 자체는 정확하나, 이 일치를 추출기 독립성의 근거로 읽지 않는다.

### §1.7 무상태 선언만으로는 재발을 막지 못한다 — lint가 이 SPEC의 본체다

무상태 선언은 문서 선언이지 기계 가드가 아니다. 선언만 하고 끝내면 에이전트가 다시 임의 frontmatter를 붙이고, 코퍼스는 같은 상태로 돌아온다. 따라서 **재발 방지 lint 규칙이 이 SPEC의 본체이며 부록이 아니다.**

lint 술어는 **D1 정의와 같은 축을 본다**: 비-spec.md SPEC 산출물(`plan.md`·`acceptance.md`·`design.md`·`research.md`)의 frontmatter에 있는 **`status:` 필드**를 거부한다. "frontmatter 일체 금지"가 아니다. 정의와 검사가 같은 축을 보게 만드는 것이 이 카드가 애초에 고치려던 어긋남이다.

**설계상 결합 — era/grandfather 예외를 두지 않는다.** D1 정리가 **이 SPEC 안에서 같이 착지**하므로, lint가 켜지는 시점에 코퍼스에는 위반이 0건이다. 따라서 `eraDemotableCodes`(`internal/spec/lint.go:248`) 편입이 필요 없다. 이것은 누락이 아니라 **의도된 설계 결정**이며, 그 성립 조건은 "정리와 lint가 함께 착지한다"이므로 AC로 못박는다.

### §1.8 lint 규칙의 구현 좌표 (설계 메모, 구현은 run-phase)

`Rule.Check(doc *SPECDoc, ...)`는 spec.md 문서를 받는다(`internal/spec/lint.go:433` `SPECDoc.Path`). 형제 산출물은 `filepath.Dir(doc.Path)`에서 유도되므로, 새 규칙은 기존 per-doc `Rule` 인터페이스 안에서 형제 파일을 스캔하는 형태로 성립한다 — 새로운 발견 경로(`discoverSPECs` 변경)는 필요 없다. `HaikuResidualRule`이 이미 SPEC 문서 밖 트리를 스캔하는 선례다(`lint.go:133` 인근 등록부).

## §2 목적

비-spec.md SPEC 산출물을 **무상태(stateless)** 로 확정한다. 세 가지를 함께 한다: 규약 명문화, 재발 방지 lint, D1 정리. 셋 중 하나라도 빠지면 목적이 성립하지 않는다 — 명문화만 하면 재발하고, lint만 하면 기존 위반(전 코퍼스 389건 @ `3b1830b96`)에 걸리며, 정리만 하면 다시 쌓인다.

## §3 요구사항 (GEARS)

### §3.1 규약 명문화 (M1)

M1의 신설 소절은 **고정 앵커 제목 `### Artifact Statelessness`** 를 갖는다. AC는 그 앵커로 소절 본문을 잘라낸 뒤 그 안에서만 판정하므로, 앵커가 없으면 어떤 판정도 통과할 수 없다. 아래 네 REQ가 그 소절이 담아야 할 것을 규정한다.

**고정 리터럴 3종** — AC가 `grep -F`로 찾는 문자열이다. 내부 백틱을 포함해 **아래 코드펜스에 적힌 그대로** 써야 한다:

```text
binds `spec.md` only
Tier-independent
Frontmatter itself is permitted
```

- REQ-AST-001-001: The frontmatter schema document SHALL carry a sub-section anchored by the literal heading `### Artifact Statelessness`, and that sub-section SHALL state that the canonical 12-field obligation binds spec.md only, using literal 1 of the three fixed literals above verbatim.
- REQ-AST-001-002: The `### Artifact Statelessness` sub-section SHALL state that the declaration is Tier-independent, using literal 2 above verbatim, so that a SPEC carrying `design.md` without `tier: L` is not left outside the rule.
- REQ-AST-001-003: The `### Artifact Statelessness` sub-section SHALL define statelessness on the status axis by naming all four artifacts (`plan.md`, `acceptance.md`, `design.md`, `research.md`) together with the prohibition of a `status:` field in their frontmatter.
- REQ-AST-001-014: The `### Artifact Statelessness` sub-section SHALL state that frontmatter as such remains permitted, using literal 3 above verbatim, and SHALL NOT carry any blanket prohibition of frontmatter.
- REQ-AST-001-016: Each of the three fixed literals SHALL appear as a plain declarative assertion, and SHALL NOT appear inside a negated sentence, a counter-example, or a "do not read this as …" construction.

### §3.2 재발 방지 lint (M2)

- REQ-AST-001-004: Where a SPEC directory contains a non-`spec.md` artifact among `plan.md`, `acceptance.md`, `design.md`, and `research.md`, the lint engine SHALL emit a finding when that artifact's YAML frontmatter carries a `status:` field.
- REQ-AST-001-005: The lint engine SHALL NOT emit that finding for an artifact whose frontmatter carries no `status:` field, nor for `spec.md`, nor for `progress.md`.
- REQ-AST-001-006: The new lint finding code SHALL be absent from `eraDemotableCodes`, because the corpus cleanup lands in the same SPEC and leaves no grandfathered violation to demote.
- REQ-AST-001-007: When the lint rule is exercised against a deliberately planted `status:` field in a non-`spec.md` artifact, the lint engine SHALL emit the rejection observably — a finding code in the report and a non-zero gating outcome under the rule's declared severity.

### §3.3 코퍼스 정리 (M3)

- REQ-AST-001-008: The corpus cleanup SHALL cover **every** SPEC directory under `.moai/specs/SPEC-*/` regardless of the SPEC's lifecycle status (population = all 696 at `3b1830b96`, 389 files), and SHALL remove only the `status:` line from a non-`spec.md` artifact's frontmatter, leaving every other frontmatter field in place (definition D1).
- REQ-AST-001-009: The corpus cleanup SHALL re-measure its target set at the run-phase HEAD rather than reuse the 389 counted at `3b1830b96`, because `origin/develop` has advanced by 26 commits to `48d8ef4be` since the original measurement.
- REQ-AST-001-010: The corpus cleanup and the lint rule SHALL land in the same SPEC, so that no era carve-out is required for the lint to hold against the live corpus.
- REQ-AST-001-011: The corpus cleanup SHALL NOT modify `spec.md` or `progress.md` in any SPEC directory.

### §3.3b 템플릿 미러 (M1에 부수)

미러는 **실재하며 현재 바이트 동일**하다 — 조건절이 아니라 사실이다:

```bash
diff -q .claude/rules/moai/development/spec-frontmatter-schema.md \
        internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
# (무출력) → MIRROR IDENTICAL @ 3b1830b96, 양쪽 23,317 bytes
```

- REQ-AST-001-015: The template mirror at `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md` SHALL receive the same `### Artifact Statelessness` sub-section in the same commit as the local rule file, so the two files remain byte-identical (Template-First, `CLAUDE.local.md` §2 [HARD]).

### §3.4 무엇을 하지 않는가

- REQ-AST-001-012: This SPEC SHALL NOT repair the twelve SPECs carrying `tier: 2` (7) or `tier: 3` (5) — values outside the S/M/L enum — and SHALL record them as an observation only.
- REQ-AST-001-013: This SPEC SHALL NOT backfill any closed SPEC's non-`spec.md` artifact status to match its `spec.md` status.

## §3.9 Acceptance Criteria 매핑

GWT 본문과 실행 가능한 판정 명령은 `acceptance.md`에 있다. 아래는 REQ ↔ AC 추적 매트릭스다.

- AC-AST-001-01: Given 규약에 `### Artifact Statelessness` 앵커가 없을 때, When M1 소절이 착지하면, Then 그 소절 안에서 spec.md 한정·4종 무상태·Tier 무관 세 문장이 각각 별도로 판정되고 부정문 안에 있지 않다 (maps REQ-AST-001-001, REQ-AST-001-002, REQ-AST-001-003, REQ-AST-001-016)
- AC-AST-001-02: Given 무상태가 frontmatter 금지로 읽힐 여지가 있을 때, When 그 소절을 잘라내 읽으면, Then frontmatter 허용 문장이 부정되지 않은 채 있고 일체 금지 문형이 0건이다 (maps REQ-AST-001-003, REQ-AST-001-014, REQ-AST-001-016)
- AC-AST-001-03: Given lint 엔진이 규칙 배열과 era 예외 목록을 가질 때, When M2가 착지하면, Then 새 코드가 등록되고 era 예외에는 없다 (maps REQ-AST-001-004, REQ-AST-001-006)
- AC-AST-001-04: Given 규칙이 작성됐으나 거부를 관측하지 못한 상태에서, When 비-spec.md 산출물에 status를 심고 lint를 돌리면, Then 거부를 관측하고 원복한다 (maps REQ-AST-001-007)
- AC-AST-001-05: Given status 없는 산출물과 spec.md·progress.md가 존재할 때, When lint를 돌리면, Then 그 셋에 대해 거부가 나오지 않는다 (maps REQ-AST-001-005, REQ-AST-001-011)
- AC-AST-001-06: Given run-phase HEAD에서 재측정한 D1 대상 N개가 있을 때, When M3가 착지하면, Then 재측정 잔여가 0이 된다 (maps REQ-AST-001-008, REQ-AST-001-009)
- AC-AST-001-07: Given D1이 status 라인만 지우는 정의일 때, When M3 diff를 읽으면, Then 제거된 비-status 라인이 0이다 (maps REQ-AST-001-008)
- AC-AST-001-08: Given 두 파일이 대상 밖일 때, When M3 diff를 읽으면, Then spec.md·progress.md 변경이 없다 (maps REQ-AST-001-011)
- AC-AST-001-09: Given era 예외 미설정이 동시 착지에 의존할 때, When 종결 시점 트리에서 판정하면, Then 규칙 등록과 D1 잔여 0이 함께 성립한다 (maps REQ-AST-001-006, REQ-AST-001-010)
- AC-AST-001-10: Given tier 값 정정과 백필이 스코프 밖일 때, When 전체 diff를 읽으면, Then tier 편집 0건이고 status 추가 0건이다 (maps REQ-AST-001-012, REQ-AST-001-013)
- AC-AST-001-11: Given 템플릿 미러가 실재하고 현재 바이트 동일할 때, When M1이 착지하면, Then 두 파일이 여전히 바이트 동일하고 양쪽에 앵커 소절이 있다 (maps REQ-AST-001-015)

## §4 스코프 제외

이 SPEC이 **만들지 않는** 것을 못박는다.

### Out of Scope — 스키마 위반 tier 값 정정

- `tier: 2` **7건**과 `tier: 3` **5건**(합 12건)은 `S`/`M`/`L` 열거값 밖의 값이다. 실측으로 확인했으나 **이 SPEC은 이것을 고치지 않는다** — 관측 기록으로만 남긴다.
- 근거: status 축과 tier 축은 별개다. 이 카드가 판정한 것은 status 축뿐이며, tier 값 정정은 별도 카드 소관이다.

### Out of Scope — D2(frontmatter 블록 통째 제거)

- `id`·`title`·`version`·`created` 등 status 밖 필드의 제거는 하지 않는다. D2 채택 시 대상은 417개이나, 그 차분 55개는 spec.md frontmatter 복제 블록으로서 **별개의 결정**을 요구한다(§1.6).
- 근거: 무상태 선언이 지배하는 축은 status 하나다. 다른 필드 제거는 이 카드가 아무 판정도 하지 않은 축에 대한 범위 확대다.

### Out of Scope — 종결 SPEC 산출물 상태 백필

- 닫힌 SPEC의 비-spec.md 산출물 상태를 사후에 `completed`로 옮기는 작업(안 A, 파일 1,251개)은 하지 않는다.
- 근거: 백필은 **기록된 값을 다른 값으로 바꾸는** 행위이고, 이 SPEC의 D1 정리는 **규약이 정의한 적 없는 필드를 제거하는** 행위다 — 축이 다르다. 전체 논증은 §1.5 「소급 편집」 절.

### Out of Scope — progress.md의 status 표기

- `progress.md`의 상태 표기는 건드리지 않는다. `closer.go`가 쓰는 두 번째 파일이지만 형태가 frontmatter가 아니라 본문 라인이라 축이 다르다.
- 근거: 이 SPEC의 술어는 "frontmatter의 `status:` 필드"이며, progress.md는 그 술어에 걸리지 않는다.

### Out of Scope — 카드가 인용한 원 사례의 재현

- `SPEC-AC-COUNT-DISCRIMINATOR-001`은 lane-7의 t338 브랜치 소관이며 develop 코퍼스에 **없다**. 이 SPEC은 그 사례를 재현하지 않고, 그것을 판정 근거로도 삼지 않는다.
- 근거: 그 SPEC이 develop에 없다는 사실 자체가 실측이며, 이 SPEC의 모든 계수는 develop 트리 `c6aa61346` 귀속이다(§5).

## §5 Gaps — 관측하지 않은 것

- **원 사례 미재현.** 카드가 지목한 `SPEC-AC-COUNT-DISCRIMINATOR-001`은 lane-7 t338 브랜치에 있고 develop 코퍼스에 없다. 원 사례를 직접 재현하지 못했다.
- **계수의 baseline 고정 — 두 트리가 섞여 있다.** 초기 실측(696 / 633 / 362 / 417 / 170 / 106 / 12)은 트리 **`c6aa61346`** 귀속이고, 전 코퍼스 D1 수치(389 = 27/34/164/164)와 미러 동일성 확인은 트리 **`3b1830b96`** 귀속이다. 두 SHA를 병기해 읽는다. 측정 이후 `origin/develop`은 **`48d8ef4be`(26 커밋)** 로 전진했으므로, **정리 마일스톤 착수 전에 run-phase HEAD에서 재측정한다**(REQ-AST-001-009). 이 문서의 숫자를 판정값으로 재사용하지 않는다.
- **lint 구현 비용 미측정.** §1.8은 구현 좌표를 적었을 뿐 규칙 신설의 실제 작업량은 재지 않았다.
- **상태 추출의 한계 — 그리고 그 실패 모드는 교차검증되지 않았다.** 판정은 frontmatter 첫 `status:` 라인 기준이며, 1행이 `---`가 아닌 파일은 전부 "frontmatter 없음"으로 계상된다. frontmatter가 2행부터 시작하는 변칙 파일이 있으면 오분류될 수 있다. 362를 도출한 두 경로가 **같은 `fm_of` 추출기를 공유**하므로 이 실패 모드는 두 경로에 동일하게 작용한다 — 일치가 검증한 것은 집계 로직뿐이다(§1.6).
- **배포 사용자 코퍼스 미측정.** 새 lint를 error 심각도로 두었을 때 배포 사용자 트리에 기존 위반이 있는지는 이 리포에서 잴 수 없다. 완화안은 `plan.md` §B3.
- **[부채] 리터럴 앵커는 주장의 극성(polarity)을 판정하지 못한다.** `AC-AST-001-01`·`-02`가 판정하는 것은 세 고정 리터럴의 **존재와 배치**(앵커 소절 안에 있는가)이지, 그 문장이 무엇을 **주장**하는가가 아니다. 감사관이 네 리터럴을 전부 담되 의미를 전부 뒤집은 소절(`This rule does NOT say that the obligation binds …`, `emphatically not Tier-independent`)을 지어 실행한 결과 **S1·S2·S3·AC-02가 모두 PASS**했다. 완화로 같은 줄 부정어 근접 검사를 넣었고(`\b(not|never|no longer|rejected|incorrect|false)\b|아니|않`) 위 적대적 사례와 흔한 비고의 경로(반례 문형, 윤문 중 부정형 전환)는 잡히지만, **여러 줄에 걸친 부정**(`… only.` 다음 줄에 `— 이것은 사실이 아니다`)은 잡지 못한다. 문자열 검사로는 닫히지 않는 잔여이며, `acceptance.md` DoD의 **인간 판독 항목**이 유일한 방어다. 정규식을 더 조이는 방향은 채택하지 않았다 — 오탐이 올바른 작업을 막는 비용이 잔여 위험보다 크다.

## §6 참조

- 실측 보고: `.moai/reports/t357/plan-measurement.md` (커밋 `7537ce693`)
- 원자료: `.moai/reports/t357/t357_rows.tsv`, `t357_closed.tsv`, `t357_audit.txt`, `t357_measure.sh`
- 감사 보고: `.moai/reports/t357/plan-audit.md` (D1~D9; 이 개정이 닫는 결함 목록)
- D1/D2 필드 축 원자료: `.moai/reports/t357/t357_fmrows.tsv`, `.moai/reports/t357/t357_fmfields.txt`
- 모집단별 재측정 스크립트: `.moai/reports/t357/t357_d1_all.sh` (전 코퍼스 vs 종결), `.moai/reports/t357/t357_d1_by_artifact.sh` (전 코퍼스 산출물별)
- 규약 SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- lint 엔진: `internal/spec/lint.go`, `internal/spec/lint_ownership.go`, `internal/spec/audit.go`, `internal/spec/closer.go`
