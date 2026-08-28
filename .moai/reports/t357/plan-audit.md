# SPEC 감사 보고 — SPEC-ARTIFACT-STATELESS-001

반복: 1/2 (Tier M 상한)
판정: **PASS-WITH-DEBT**
종합 점수: **0.86** (Tier M 임계 0.80)

- 감사 트리: `.claude/worktrees/t357` · HEAD `3b1830b96` · 브랜치 `WT-tierl-status-transitions`
- 산출물 입력: `spec.md` · `plan.md` · `acceptance.md` · `progress.md` (Tier M 계약)
- **M1 문맥 격리 적용**: 작성자의 추론 맥락은 무시했다. 판정 근거는 산출물 파일과 `.moai/reports/t357/`·`.moai/cache/`의 원자료를 이 트리에서 직접 재측정한 결과뿐이다.
- 재감사가 열린 이유: must-pass 7항 전부 통과하고 점수가 임계를 넘었으나, **blocking 등급 결함 4건**이 인수 계층에 남아 있다. 아래 D1~D4는 Implementation Kickoff Approval 전에 닫는다.

---

## Must-Pass 결과

| | 항목 | 판정 | 근거 |
|---|---|---|---|
| MP-1 | REQ 번호 일관성 | **PASS** | `spec.md:L119-149` — `REQ-AST-001-001` ~ `-013` 연속 13개. 결번·중복 0, 3자리 zero-padding 일관 |
| MP-2 | GEARS 형식 (요구사항 계층) | **PASS** | 13개 전수 확인. Where형 `REQ-AST-001-004`("Where a SPEC directory contains…, the lint engine SHALL emit"), When형 `-007`, Unwanted형 `-005`/`-011`/`-012`/`-013`(`SHALL NOT`), 나머지 Ubiquitous형. **판정 계층은 `spec.md`의 `REQ-XXX` 항목이며**, `acceptance.md`의 Given-When-Then은 검증 계층이므로 이 항목에서 감점하지 않았다 |
| MP-3 | YAML frontmatter 유효성 | **PASS** | 12필드 전수 존재(`spec.md:L2-13`), `version: "0.1.0"` 인용, `created`/`updated` ISO, `priority: P2`, `lifecycle: spec-anchored`, `tags` CSV 문자열. 거부 alias(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. 기계 확인: `moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md` → `✓ No findings`, rc=0 |
| MP-4 | 언어 중립성 | **N/A (auto-pass)** | 이 SPEC은 이 리포의 Go lint 엔진과 자체 SPEC 코퍼스만 대상으로 한다. 16개 프로그래밍 언어 도구를 열거하는 성격이 아니며, 언어별 도구명 하드코딩 0건 |
| MP-5 | D7 교차-SPEC 정합 | **PASS** | 참조 2건. `SPEC-PHASE-FRONTMATTER-OWNER-001` → 존재, `status: completed` (retired/superseded/archived 아님, 화해 불필요). `SPEC-AC-COUNT-DISCRIMINATOR-001` → `.moai/specs/`에 부재 = D7-5 SHOULD 등급이나, `spec.md:L196-198`과 `§5`가 부재 사실과 근거를 명시 기록하고 판정 근거에서 배제했으므로 해소됨. **BLOCKING 0건** |
| MP-6 | D8 크로스플랫폼 규율 | **PASS (auto)** | `grep -c syscall spec.md` = **0**. 관심사 부재 |
| MP-7 | clarification gate | **PASS** | `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-ARTIFACT-STATELESS-001/` → 매치 0 (rc=1). 미해결 마커 없음 |

**must-pass 7항 전부 통과.** M5 방화벽에 걸리는 항목 없음.

---

## 항목별 점수

| 차원 | 점수 | 루브릭 대역 | 근거 |
|---|---|---|---|
| Clarity | 0.80 | 0.75~1.0 | 결정을 번복 가능성 순으로 배열(`plan.md:L7`)하고 각 결정에 번복 지점을 명시한 구성은 모범적이다. 감점 사유 둘: (1) D1 모집단이 **종결 SPEC 633** 기준(`spec.md:L106`)인데 `REQ-AST-001-008`은 모집단을 명시하지 않고 `acceptance.md`의 `count_d1`은 **전 코퍼스 696** 기준이다 — 두 엔지니어가 다르게 구현할 수 있다. (2) `plan.md:L108`의 "순서는 M2 → M3든 M3 → M2든 무방"이 `AC-AST-001-04` 판정 1항과 충돌한다(D2 참조) |
| Completeness | 0.95 | 1.0 대역 근접 | HISTORY·배경(WHY)·목적(WHAT)·요구사항·AC 매핑·Out of Scope·Gaps·참조 전부 존재. `### Out of Scope — <주제>` H3 소제목 **5개**(`spec.md:L176,181,186,191,196`), 각각 구체 `-` 항목과 근거 보유. frontmatter 완전. Tier M 산출물 3종 + progress.md 충족이며 design/research 부재를 `progress.md:L7`이 계약 근거와 함께 명시 |
| Testability | 0.70 | 0.50~0.75 | AC 10개 중 3개(`-03`·`-05`·`-06`)는 견고하고 `-04`는 비공허성 가드까지 갖췄다. 그러나 `AC-AST-001-02`는 **선언한 PASS 문자열이 원리상 출력될 수 없고**(D1), `AC-AST-001-01`은 7개 conjunct 중 6개가 현재 파일에서 이미 참이라 **키워드 1개 검사로 붕괴**한다(D3). `-09`는 미정의 셸 함수를 참조하고 `-07`/`-08`/`-10`은 `BASE=<…>` 자리표시자를 남긴다 |
| Traceability | 1.00 | 1.0 | REQ 13개 전부 최소 1개 AC 보유(`spec.md:L153-162`). AC 10개 전부 실존 REQ를 지시. 고아 AC 0, 미커버 REQ 0. 교차 확인: `-001`→AC-01, `-002`→AC-01, `-003`→AC-01/02, `-004`→AC-03, `-005`→AC-05, `-006`→AC-03/09, `-007`→AC-04, `-008`→AC-06/07, `-009`→AC-06, `-010`→AC-09, `-011`→AC-05/08, `-012`→AC-10, `-013`→AC-10 |

산술 평균 = (0.80 + 0.95 + 0.70 + 1.00) / 4 = **0.8625 → 0.86**

---

## 지시된 7개 점검 항목의 판정

### 1. 정리 정의 고정 (D1 vs D2) — **충족, 수치 전부 독립 재측정으로 확인**

`spec.md:L100-116`이 D1을 명시 채택하고 근거("무상태 선언이 지배하는 것은 status 축")를 적었다. 수치를 **인용하지 않고 이 트리에서 직접 다시 셌다**:

```
$ awk -F'\t' '$5=="yes"{d[$2]++;t++} END{...}' .moai/cache/t357_fmrows.tsv
research 29 · acceptance 155 · design 23 · plan 155 · TOTAL 362

$ awk -F'\t' '$4=="yes"{...}' .moai/cache/t357_fmrows.tsv
research 32 · acceptance 180 · design 25 · plan 180 · TOTAL 417
```

`spec.md:L106-109`의 표(D1: 23/29/155/155 = 362, D2: 25/32/180/180 = 417)와 **셀 단위까지 일치**한다. 부수 수치도 전부 맞다: D2 블록 417개 중 12~14필드 **224개**, 필드 빈도 `version` 371 · `status` 362 · `id` 356 · `created` 340 · `title` 289 — 전부 재현됨.

다만 `spec.md:L116`의 "두 계수는 **서로 다른 스크립트**에서 나왔다"는 교차검증 주장은 과장이다 → D5.

### 2. lint 술어 정렬 — **충족, era 예외 미설정의 성립 조건도 명시적으로 못박음**

`REQ-AST-001-004`가 거부 대상을 "비-spec.md 산출물 frontmatter의 `status:` 필드"로, `REQ-AST-001-003`이 "frontmatter 자체를 금지하지 않음"으로 각각 못박았다. `plan.md:L21-27`의 3×3 표가 통과/거부를 파일 유형별로 전개한다. **정의(D1)와 검사가 같은 축을 본다** — 카드가 고치려던 어긋남을 새로 만들지 않는다.

era carve-out 불필요 주장이 "정리와 lint의 동시 착지"에 의존한다는 점을 SPEC이 **스스로 인지하고 AC로 못박았다**: `spec.md:L96-98`, `REQ-AST-001-010`, `AC-AST-001-09`("둘 중 하나만 성립하면 FAIL"), `plan.md:L31-35`("M3를 미룬다와 예외를 두지 않는다는 동시에 성립할 수 없다"). 지적할 것이 없다.

구현 좌표도 확인했다: `eraDemotableCodes`는 `internal/spec/lint.go:247`에 실재(SPEC 표기 248은 map 항목 시작행 — 허용 오차), 규칙 배열은 `lint.go:127`의 `l.rules = []Rule{…}`, 트리 스캔형 선례 `HaikuResidualRule`도 실재, `Rule.Check(doc *SPECDoc, all []*SPECDoc)` 시그니처가 형제 파일 유도 방식을 실제로 허용한다.

### 3. 관측 인수 (AC-AST-001-04) — **비공허성 가드는 정확히 설계됨. 다만 실행 순서 결함 있음**

판정 1항 "`before` 매치 수 = 0"이 정확히 요구된 가드다. 아무것도 보지 않는 lint는 before=0·after=0이 되어 판정 2항("after ≥ 1")에서 걸린다. 심기 → 관측 → 원복 3단계가 실행 가능한 명령으로 적혔고, DoD가 "한쪽만 인용하면 미충족"까지 못박았다. **이 부분은 모범적이다.**

`moai spec lint`가 실재하고(`internal/cli/spec_lint.go:28`), 인자 없이 전 코퍼스를 검사하며, `printTable`이 `CODE` 열을 출력하고(`spec_lint.go:121-131`), error 발견 시 exit 1을 낸다(`spec_lint.go:93-95`) — 명령이 실제로 작동할 수 있는 형태다.

결함은 순서다 → **D2**.

### 4. 스코프 경계 — **(a) 충족, (b) 충족하며 논거가 특히 강함**

(a) `tier: 2` 7건 / `tier: 3` 5건: `REQ-AST-001-012`가 "관측 기록으로만" 못박고, `spec.md:L176-179`이 Out of Scope 절로, `AC-AST-001-10`이 "tier 편집 0건"으로 기계 판정한다. 재측정 확인: `awk` 집계 결과 `2`=7, `3`=5 — 정확.

(b) 106건: `spec.md:L92-94`가 이 구멍을 **선택된 안의 강점으로** 서술한다 — "규약을 Tier L 한정으로 좁혀 쓰는 안 A/B 형태였다면 이 106건은 규약 밖에 그대로 남는다. 안 C는 Tier와 무관하게 … 이 106건도 빠짐없이 규칙 안에 들어온다." 이것이 지시가 요구한 정확한 서술이다. `REQ-AST-001-002`가 Tier 무관성을 규약 문언 의무로 승격시켰다. 재측정 확인: 종결 SPEC 중 design 또는 research 보유 & `tier != L` = **106** — 정확.

### 5. baseline 귀속 — **충족, 그리고 재측정이 요구사항으로 승격됨**

모든 계수가 트리 `c6aa61346`에 귀속된다(`spec.md:L28`, `§5`, `plan.md:L11`, `progress.md:L6`). `REQ-AST-001-009`가 "362를 재사용하지 말고 run-phase HEAD에서 재측정하라"를 요구사항으로 못박았고, `acceptance.md` 머리말이 "이 문서의 판정 명령은 어떤 계수도 리터럴로 고정하지 않는다"를 선언했으며, DoD가 "`c6aa61346` 리터럴 재사용 금지"를 항목으로 둔다. `origin/develop`이 `48d8ef4be`(26 커밋)로 전진한 사실도 3개 산출물에 일관되게 기록됐다.

### 6. 미재현 Gap — **충족**

`SPEC-AC-COUNT-DISCRIMINATOR-001`의 부재를 `spec.md:L196-198`(Out of Scope)과 `§5`(Gaps), `progress.md:L8`이 각각 기록하고, "그것을 판정 근거로도 삼지 않는다"고 명시했다. 이 트리에서 확인: `.moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001` → `No such file or directory`. SPEC의 어떤 결론도 이 사례에 기대지 않는다 — 모든 논거가 develop 코퍼스 696건 실측에서 나온다.

### 7. 자기 정합 — **충족**

```
$ head -2 plan.md        → "# SPEC-ARTIFACT-STATELESS-001 — 구현 계획"
$ head -2 acceptance.md  → "# SPEC-ARTIFACT-STATELESS-001 — 인수 기준"
$ head -2 progress.md    → "# SPEC-ARTIFACT-STATELESS-001 — 진행 기록"
```

세 파일 모두 frontmatter 블록 자체가 없다. 자기 규칙을 위반하지 않는다. `spec.md`만 12필드 frontmatter를 보유 — 이것이 이 SPEC이 선언하는 상태 그대로다.

---

## 자체 판단 항목

### Tier M 분류는 옳다

M2(Go 규칙 1개 + 테이블 테스트)와 M1(문서 소절 1개)은 명백히 Tier M 규모다. 쟁점은 M3의 362~389 파일이나, **파일 수는 복잡도가 아니라 반복이다** — 편집 단위가 파일당 1행이고, 정확성이 `AC-AST-001-07`의 diff 판독(제거된 비-status 라인 = 0)으로 기계 판정된다. 설계 결정면은 실제로 좁다: lint 술어 1개, 정리 정의 1개, 심각도 1개.

Tier L로 올리면 `design.md`/`research.md`가 필수가 되는데, 이 SPEC의 주제가 바로 그 두 파일을 무상태로 선언하는 것이라는 아이러니는 **분류 근거가 아니다** — 그것이 Tier 판정을 좌우해서는 안 된다. 그럼에도 Tier M이 옳다고 본다. 다만 Tier M은 plan-audit 반복 상한을 2로 묶으므로, 아래 blocking 4건은 다음 1회 안에 닫혀야 한다.

### 종결 SPEC 362건 정리의 정당화는 **불충분하다** → D4

운영자가 안 A를 물린 이유 하나가 소급 편집이다. 그런데 안 C도 종결 SPEC 362건을 소급 편집한다. 이 둘이 왜 다른 행위인지 — **"규약이 존재해서는 안 된다고 말하는 필드를 지우는 것"과 "기록된 상태 값을 바꾸는 것"은 다르다** — 를 SPEC은 **논증하지 않고 전제한다**.

재료는 이미 갖고 있다. `spec.md:L74`("규약 위반이 아니라, 규약에 없는 필드를 에이전트가 임의로 붙였다가 아무도 옮기지 않은 것")와 Out of Scope 백필 절이 함께 읽히면 그 구분이 도출되지만, 어느 문장도 그것을 **말하지 않는다**. 운영자가 명시적으로 제기한 반대를, 그 SPEC이 실제로 수행하는 행위에 대해 답하지 않은 채 남겨두었다.

### 술어가 검증 불가한 요구사항 / 아무것도 증명하지 못하는 인수 명령

- `REQ-AST-001-001`~`-003`은 문서 문언 의무라 본질적으로 문자열 검사로만 판정되며, 현재 그 검사가 붕괴해 있다(D1·D3).
- `AC-AST-001-02`의 첫 명령 `grep -qE 'status:' "$F"`는 현재 파일에서 **이미 12회 매치**한다 — M1이 착지하기 전에도 참이다. 완전히 공허하다.
- 나머지 AC(`-03`·`-05`·`-06`·`-07`·`-08`·`-10`)는 잘못 수행되면 실제로 실패한다. `-06`의 `count_d1`은 이 트리에서 직접 돌려 **389**를 얻었다(작동하는 술어다).

---

## 발견된 결함

**D1. `AC-AST-001-02`의 판정이 원리상 성립할 수 없다** — `acceptance.md:L42-47` — 심각도: **major** — 등급: **blocking**

두 명령 모두 결함이다.

1. `grep -qE 'status:' "$F"` — 현재 미수정 파일에서 이미 12회 매치한다. M1 착지 여부와 무관하게 항상 `status-axis: present`를 출력한다. **공허하다.**
2. `grep -inE '(no|never|금지).{0,40}frontmatter' "$F" || echo "no blanket-frontmatter prohibition: PASS"` — 이 grep은 현재 파일에서 **이미 5행 이상 매치한다**(예: `L75` "NO new `amended` enum value … `amendment_of:` optional frontmatter", `L93` "This matrix does not cover non-transition frontmatter corrections"). 따라서 `||` 분기가 결코 실행되지 않고, AC가 선언한 판정 문자열 `no blanket-frontmatter prohibition: PASS`는 **어떤 경우에도 출력될 수 없다.**

실행 확인 (이 트리, 미수정 파일):
```
$ grep -inE '(no|never|금지).{0,40}frontmatter' .claude/rules/moai/development/spec-frontmatter-schema.md | head -5
8: ... 75: ... 93: ... 127: ... 131: ...   ← rc=0, 5행 매치
```

**필요한 수정**: (a) 첫 명령을 신설 소절 범위로 한정한다 — 예: 소절 앵커(`## Artifact Statelessness` 등)를 정하고 `awk '/^## <앵커>/,/^## /'`로 잘라낸 뒤 그 안에서 `status:`를 찾는다. (b) 둘째 명령의 부정 검사를 신설 소절 범위로 한정하고, 매치 패턴을 "frontmatter 자체 금지" 문형(예: `frontmatter.{0,20}(MUST NOT|금지)`)으로 좁힌다. 두 명령 모두 M1 착지 **전에 FAIL, 후에 PASS**가 되는지 실행해 확인한 뒤 확정한다.

---

**D2. `AC-AST-001-04`의 before=0 요건과 `plan.md`의 "M2·M3 순서 무방" 주장이 충돌한다** — `acceptance.md:L100` vs `plan.md:L108` — 심각도: **major** — 등급: **blocking**

`AC-AST-001-04`는 전 코퍼스 대상 `moai spec lint`를 인자 없이 돌리고 판정 1항으로 `before` 매치 수 = **0**을 요구한다. 그런데 M3(정리)가 아직 착지하지 않은 상태에서 M2(lint)만 착지하면, 코퍼스에 위반이 남아 있으므로 before 매치가 0이 아니다.

이 트리에서 실측한 잔여량:
```
$ bash .moai/cache/t357_audit_d1live.sh
D1 live = 389
```

즉 M2 → M3 순서로 진행하면 `AC-AST-001-04`는 **정상 동작하는 lint에 대해서도 FAIL한다.** 그럼에도 `plan.md:L108`은 "순서는 M2 → M3든 M3 → M2든 무방하나"라고 적었다. 두 문서가 서로 다른 것을 말한다.

**필요한 수정**: 둘 중 하나. (a) `plan.md:L108`을 "M3 → M2 순서로 고정한다 — `AC-AST-001-04`의 before=0 요건이 정리 선행을 요구한다"로 정정한다. (b) `AC-AST-001-04`의 lint 호출을 심는 SPEC 하나로 한정한다(`moai spec lint .moai/specs/SPEC-ARTIFACT-STATELESS-001/spec.md`) — 이때 before=0이 순서와 무관해지며, 비공허성 가드는 그대로 유지된다. (b)가 더 견고하다.

---

**D3. `AC-AST-001-01`의 판정이 키워드 1개 검사로 붕괴한다** — `acceptance.md:L22-30` — 심각도: **major** — 등급: **blocking**

Then 절은 세 가지 **명시 문장**(spec.md 한정 · 나머지 4종 무상태 · Tier 무관)을 요구한다. 그런데 7개 conjunct 중 6개가 **현재 미수정 파일에서 이미 참이다**:

```
spec.md.*(only|한정|1종)  1   ← L122가 이미 매치
plan.md                    6
acceptance.md              4
design.md                  1
research.md                1
stateless|무상태           0   ← 유일하게 거짓
tier                       2   ← L83, L152가 이미 매치
```

따라서 AC 전체가 "파일 어딘가에 `stateless` 또는 `무상태`라는 단어가 있는가" 한 문항으로 축소된다. 소절을 쓰지 않고 기존 문장에 그 단어 하나만 끼워 넣어도 PASS한다 — Then 절이 요구한 세 문장 중 어느 것도 검증되지 않는다.

(전체로는 현재 FAIL이므로 완전 공허는 아니다. 그러나 판정력이 요구한 것의 1/3에도 못 미친다.)

**필요한 수정**: 신설 소절에 안정적인 앵커 제목을 정하고, 판정을 그 소절 본문으로 한정한 뒤 **세 문장을 각각 별도 grep으로 판정**한다. 각 grep이 M1 착지 전 FAIL·후 PASS임을 실행해 확인한다.

---

**D4. 종결 SPEC 소급 편집의 정당화가 전제로만 남아 있다** — `spec.md:L186-189` — 심각도: **major** — 등급: **blocking**

운영자가 안 A를 물린 근거 하나가 소급 편집(이력 왜곡)이다. 안 C도 종결 SPEC 362건을 소급 편집하며, 그 둘이 왜 다른 행위인지에 대한 논증이 SPEC 어디에도 없다. Out of Scope 절은 "운영자가 안 C를 선택했다"를 근거로 들지만, 이것은 **결정의 인용이지 근거가 아니다** — 그리고 그 결정이 물린 것이 바로 이 SPEC이 수행하는 것과 같은 종류의 편집이다.

**필요한 수정**: `§1.5` 또는 `§1.6`에 한 문단을 추가한다. 요지는 두 행위의 축이 다르다는 것이다 — 백필(안 A)은 **기록된 상태 값을 사후에 다른 값으로 바꾸는** 행위로 종결 시점의 기록을 왜곡하지만, D1 정리는 **규약이 정의한 적 없는 필드를 제거하는** 행위이므로 지워지는 것이 애초에 기록으로서의 지위를 가진 적이 없다(그 근거는 이미 `spec.md:L74`에 있다). 이미 가진 재료를 문장으로 세우기만 하면 된다.

---

**D5. "서로 다른 스크립트에 의한 독립 교차검증" 주장이 과장이다** — `spec.md:L116` — 심각도: **minor** — 등급: **blocking**

362가 두 경로에서 나온 것은 사실이고 나도 재현했다(`t357_fmrows.tsv` 집계 362, `t357_rows.tsv` 재집계 362). 그러나 `t357_measure.sh`와 `t357_fmscope.sh`는 **동일한 `fm_of` awk 추출기를 공유한다**:

```awk
NR==1&&/^---/{f=1;next} f&&/^---/{exit} f
```

`§5`가 스스로 기록한 한계("1행이 `---`가 아닌 파일은 전부 frontmatter 없음으로 계상")는 두 스크립트에 **동일하게** 작용한다. 즉 공통 실패 모드에 대해 두 계수는 서로를 검증하지 못한다. 진짜 독립 검증이 아니라 **집계 로직의 상호 검증**이다.

**필요한 수정**: `spec.md:L116`을 "동일한 frontmatter 추출기를 공유하므로 추출 단계의 공통 실패 모드는 교차검증되지 않는다 — 검증된 것은 집계 로직이다"로 한정한다. 계수 자체는 정확하므로 수치는 그대로 둔다.

---

**D6. D1 대상 모집단이 문서 간에 다르다 (종결 633 vs 전 코퍼스 696)** — `spec.md:L106` vs `acceptance.md:L152-166` — 심각도: **minor** — 등급: **blocking**

`§1.6`은 D1 = 362를 "모집단 = 종결 SPEC"으로 명시한다. 그런데 `REQ-AST-001-008`은 모집단을 말하지 않고, `AC-AST-001-06`의 `count_d1`은 `.moai/specs/SPEC-*/` 전부를 순회한다 — 종결 여부를 걸러내지 않는다. 이 트리 실측:

```
종결 전용 (c6aa61346 원자료 재집계) = 362
전 코퍼스 (HEAD 3b1830b96, count_d1 그대로 실행) = 389
```

차이 27건이 미종결 SPEC에 있다. 실무 결과는 오히려 넓은 쪽이 낫지만(`AC-06`이 0을 요구하므로 어차피 전부 지워야 한다), **문서가 말하는 대상 수와 판정이 요구하는 대상 수가 다르다**. 게다가 `plan.md:L96`의 "재측정값이 362에서 ±20% 초과 시 원인 규명" 가드가 **서로 다른 모집단의 수를 비교하게 된다**(389는 362 대비 +7.5%라 우연히 걸리지 않으나, 그 통과는 근거가 아니라 우연이다).

**필요한 수정**: `REQ-AST-001-008`에 모집단을 명시한다 — 전 코퍼스로 하는 것이 `AC-06`과 정합적이다. 그 경우 `§1.6`에 "362는 종결 SPEC 기준 참고값이며 실제 대상은 전 코퍼스 재측정값"을 한 줄 덧붙이고, `plan.md:L96`의 ±20% 가드도 같은 모집단으로 재기술한다.

---

**D7. `AC-AST-001-09`가 미정의 셸 함수를 참조한다** — `acceptance.md:L216-220` — 심각도: **minor** — 등급: **optional**

`[ "$(count_d1)" -eq 0 ]`가 `AC-AST-001-06`에서 정의된 함수를 쓰는데, 각 코드 블록은 별개 셸에서 실행되므로 그대로 붙여넣으면 `command not found`로 실패한다. 주석("`AC-AST-001-06`의 `count_d1`을 재사용")이 의도를 밝히므로 실행자가 복구할 수 있다.

**필요한 수정**: `count_d1` 정의를 `AC-09` 블록에도 인라인하거나, 정의를 문서 상단 공용 블록으로 올린다.

---

**D8. `AC-07`/`-08`/`-10`이 자리표시자를 남긴 채로 있다** — `acceptance.md:L177,197,231` — 심각도: **minor** — 등급: **optional**

`BASE=<M3 착수 직전 커밋 SHA>` / `BASE=<SPEC 착수 직전 커밋 SHA>`는 실행 시 채워야 하는 값이다. 의도가 명확하고 run-phase에 자연히 확정되므로 결함으로서는 약하다.

**필요한 수정**: `progress.md §E.2`에 "M3 착수 직전 SHA"와 "SPEC 착수 직전 SHA"를 기록할 자리를 미리 만들어 두고, AC 본문이 그 자리를 가리키게 한다.

---

**D9. 템플릿 미러가 실재하는데 조건절로 남아 있고 AC가 없다** — `plan.md:L82`, `acceptance.md:L239` — 심각도: **minor** — 등급: **optional**

`plan.md` M1은 "미러돼 있는지 확인하고, **있으면** 같은 커밋에서 함께 수정"이라고 조건절로 적었고 품질 게이트도 "템플릿 미러가 **존재하면**"이다. 그런데 미러는 지금 확인 가능하며 실재한다:

```
$ ls -la internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md
-rw-r--r--  23317  spec-frontmatter-schema.md        ← 로컬본과 동일 크기
```

지금 알 수 있는 사실을 미지로 남겨두었고, 미러가 드리프트했을 때 실패하는 AC가 없다. Template-First는 `CLAUDE.local.md §2`에서 [HARD]다.

**필요한 수정**: 조건절을 사실로 바꾸고("미러가 존재한다 — 같은 커밋에서 함께 수정한다"), 미러 정합을 판정하는 명령을 품질 게이트에 넣는다(예: 두 파일에 같은 앵커 소절이 존재하는지 grep, 또는 `diff`).

---

**D10. `SPEC-AC-COUNT-DISCRIMINATOR-001` 참조가 D7-5 SHOULD를 유발한다** — `spec.md:L196` — 심각도: **minor** — 등급: **optional (해소됨)**

참조된 SPEC이 `.moai/specs/`에 없다. 통상 D7-5 SHOULD 소견이나, 이 SPEC은 부재를 실측 사실로 기록하고 판정 근거에서 배제했으므로 실질적으로 해소됐다. 기록 목적으로만 남긴다. 조치 불필요.

---

## 권고

점수 0.86은 Tier M 임계 0.80을 넘고 must-pass 7항은 전부 통과했다. 그러나 **D1~D6 여섯 건은 blocking**이며, 그중 셋(D1·D2·D3)은 성격이 같다 — **잘못 수행돼도 통과하거나, 옳게 수행돼도 실패하는 인수 명령**이다. 이 SPEC의 주제 자체가 "존재하지만 아무것도 보지 않는 검사"를 없애는 것이므로(`plan.md:L128`이 카드 t355의 실패 형태로 직접 지목한다), 그 결함이 자기 인수 계층에 남아 있는 것은 특히 무겁다.

Implementation Kickoff Approval 전에 다음 순서로 닫는다.

1. **D2** — `AC-AST-001-04`의 lint 호출을 심는 SPEC 하나로 한정하고, `plan.md:L108`의 순서 무방 주장과 정합시킨다. (가장 저렴하고 파급이 크다)
2. **D1 · D3** — `AC-01`·`AC-02`를 신설 소절 범위로 한정하고, **각 판정을 미수정 파일에 대해 실제로 실행해 FAIL이 나오는 것을 확인한 뒤** 확정한다. 확인하지 않은 판정 명령을 다시 넣지 않는다.
3. **D4** — `§1.5`/`§1.6`에 소급 편집 구분 한 문단. 재료는 `spec.md:L74`에 이미 있다.
4. **D6** — `REQ-AST-001-008`에 모집단 명시(전 코퍼스 권고), `§1.6`과 `plan.md:L96` ±20% 가드를 같은 모집단으로 재기술.
5. **D5** — `spec.md:L116`의 교차검증 주장 한정.
6. D7~D9는 optional. 오케스트레이터 재량이며, D9는 Template-First가 [HARD]인 만큼 함께 처리하기를 권한다.

재감사는 위 열거된 결함 델타에만 범위를 두며(Tier M 상한 2회 중 1회 남음), 전면 재감사는 하지 않는다.

**강점 기록** (수정 시 훼손하지 말 것): 실측이 카드 전제를 뒤집은 경위를 `§1.1`이 정면으로 기록한 점, 106건 구멍을 선택안의 강점으로 논증한 `§1.5`, era 예외 미설정의 성립 조건을 스스로 찾아 `AC-09`로 못박은 `§1.7`/`plan.md §B2`, `REQ-AST-001-009`가 baseline 재측정을 요구사항으로 승격시킨 것, `AC-AST-001-04`의 before=0 비공허성 가드, 그리고 인용한 모든 계수(696/633/362/417/224/170/106/12/389)가 이 감사의 독립 재측정과 **한 건도 어긋나지 않은 것**.

---

**감사 baseline 귀속**: 모든 판정은 이 워크트리(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t357`) HEAD `3b1830b96`에서 이번 실행에 직접 실행한 명령의 출력에 귀속된다. 인용된 수치 중 다른 트리·다른 시점에서 옮겨온 것은 없다.

**미관측 (Gaps)**: (1) `moai spec lint`를 전 코퍼스 대상으로 실제 실행하지는 않았다 — 단일 SPEC 대상 실행(rc=0, 소견 0)만 관측했다. (2) 신설될 lint 규칙은 아직 존재하지 않으므로 그 동작은 코드 좌표의 성립 가능성만 확인했고 실행하지 않았다. (3) 배포 사용자 코퍼스의 기존 위반 여부는 이 리포에서 측정 불가하다 — SPEC이 `plan.md §D`에 Gap으로 이미 기록한 것과 같은 항목이다. (4) 템플릿 미러의 내용 동일성은 파일 크기 일치만 확인했고 `diff`로 대조하지 않았다.

**잔여 위험**: D2를 (a) 방식(순서 고정)으로 닫으면 M3 → M2 순서가 강제되는데, 그러면 정리가 lint 없이 선행하므로 정리 중 새로 유입되는 위반을 막을 장치가 그 구간에 없다. (b) 방식(단일 SPEC 한정)이 이 위험을 만들지 않는다. 또한 D6을 전 코퍼스로 확정하면 편집 대상이 389 이상으로 늘어나며, 그중 미종결 SPEC의 산출물은 아직 작업 중일 수 있어 다른 레인의 미커밋 변경과 충돌할 여지가 `plan.md §D`의 기록보다 커진다.
