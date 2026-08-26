# SPEC Review Report: SPEC-ZONE-REGISTRY-RESYNC-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **FAIL**
Overall Score: **0.75** (Tier M PASS 임계 0.80)

감사 트리: `.claude/worktrees/t232` @ `880a43363`, 브랜치 `WT-zone-registry-drift`, 작업 트리 클린.
SPEC이 base 로 적은 `294b4b6ab` 와 HEAD 사이의 차이는 SPEC 저작 커밋 1개뿐이며, `git diff 294b4b6ab..HEAD -- .claude/rules internal/template/templates/.claude/rules` 가 빈 출력이다 — 규칙 트리가 움직이지 않았으므로 SPEC 이 인용한 모든 RED 값은 HEAD 에서도 그대로 유효하다.

M1 Context Isolation: 저자 추론 컨텍스트는 전달받지 않았고, 전달됐더라도 무시한다. 판정은 `spec.md` / `plan.md` / `acceptance.md` (Tier M 3종) 와 내가 직접 돌린 명령의 출력만을 근거로 한다. `.moai/reports/t232/` 의 레인 측정치는 **주장으로 취급**해 전부 재측정했다.

---

## 요약 — 왜 FAIL 인가

이 SPEC 은 스스로 "AC 는 전부 mutant 내성으로 작성한다"(`spec.md` §3)를 중심 품질 축으로 선언하고, `plan.md` §G 에 AC별 mutant 대응표를 둔다. 그 선언이 판정 기준이다.

측정 근거는 견고하다 — RED 값 7종을 전부 독립 재현했고 하나도 틀리지 않았다. 요구사항은 GEARS 를 지키고, must-pass 7개는 전량 통과하며, 리드가 지목한 Tier 압축(#6)과 순서 구속(#7)은 걱정할 근거가 없다.

문제는 선언한 그 축에서 난다. **AC 를 통과하면서 그 AC 가 지키려는 요구사항을 위반하는 구체적 구현 4개**를 찾았고, 그중 3개는 `plan.md` §G 표에 행이 없다. 특히 D-1 은 BLOCKING AC 9건 중 4건(001·002·003·004)을 **한 번에** 통과시키면서 수리를 전혀 하지 않는 경로다. §G 는 "이 노트가 비면 그 AC 는 약한 것이다"라고 스스로 적었는데, 비어 있는 칸이 가장 큰 구멍 위에 있다.

FAIL 의 사유는 작업이 얕아서가 아니다. 오히려 반대다 — 이 SPEC 은 자기가 무엇을 증명해야 하는지 정확히 알고 있고, 그래서 증명이 새는 지점이 그만큼 또렷하게 판정된다. 네 구멍의 수정은 전부 AC 본문 3~5줄 추가로 끝나며, 재감사는 그 델타에만 한정된다.

---

## Must-Pass Results

| | 판정 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | `grep -o 'REQ-ZRR-[0-9]\{3\}' spec.md \| sort -u` → `REQ-ZRR-001`..`015` 15개, 결번 0, 중복 0(`uniq -d` 빈 출력), zero-padding 일관. 단 **제시 순서가 어긋난다** — §4 "가드" 절에서 `REQ-ZRR-015` 가 `009` 와 `010` 사이에 놓여 있다. 집합은 완전하므로 PASS, 가독성 결함으로 D-7 에 기록 |
| MP-2 GEARS 형식 준수 | **PASS** | 요구사항 계층(`REQ-ZRR-*` in `spec.md`) 기준 판정. 15건 전부 GEARS 5패턴 중 하나에 붙는다 — Ubiquitous 9(`The zone registry shall carry…`), Event-driven 3(`When … shall`), Unwanted 6(`shall not`). AC 의 Given-When-Then 은 검증 계층 형식이므로 여기서 감점하지 않는다(Group 4 소관) |
| MP-3 YAML frontmatter 유효성 | **PASS** | canonical 12필드 전량 존재·타입 적합: `id`(패턴 적합) / `title`(quoted) / `version:"0.2.0"`(quoted semver) / `status: draft`(8-enum 내) / `created`·`updated`(ISO) / `author` / `priority: P1` / `phase:"v3.1.3 target"`(스테이지명 아님 — 금지값 회피) / `module` / `lifecycle: spec-anchored` / `tags`(CSV). 거부 alias(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. 추가 필드 `tier`·`era`·`related_specs` 는 스키마 위반 아님 |
| MP-4 §22 언어 중립성 | **N/A (auto-pass)** | 이 SPEC 은 다국어 툴체인을 다루지 않는다. `REQ-ZRR-013` 의 "중립성"은 16개 언어 균등 열거가 아니라 템플릿의 SPEC-ID/날짜/SHA 누출 금지 축이다 |
| MP-5 D7 cross-SPEC 조정 | **PASS** | `related_specs` 2건 실재 + 상태 확인: `SPEC-V3R6-ZONE-REGISTRY-PACKAGING-001` → `status: completed`, `SPEC-V3R5-CONSTITUTION-DUAL-001` → `status: completed`. 둘 다 {retired, superseded, archived} 밖이므로 조정 요구 없음. BLOCKING 0건 |
| MP-6 D8 cross-platform 규율 | **PASS (auto)** | `grep -c syscall spec.md plan.md acceptance.md` → `0 / 0 / 0`. cross-platform 관심사 없음 |
| MP-7 clarification gate | **PASS** | `grep -rn 'NEEDS CLARIFICATION'` 이 SPEC 디렉터리에서 rc=1(무매치). Tier M 이므로 `research.md` 없음 |

**must-pass 7건 전량 PASS** — FAIL 은 firewall 이 아니라 aggregate 점수(0.75 < 0.80)에서 나온다.

---

## Category Scores

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 요구사항은 단일 해석이고 §D 11개 구속 조건이 재논의 여지를 닫는다. 감점 근거 3건: `spec.md` §5 의 파일 수 오기(D-5, `acceptance.md` AC-ZRR-003 으로 전파), `REQ-ZRR-015` 의 복합문 뒷절이 시스템 거동이 아닌 프로세스 서술(D-7), AC-ZRR-007 의 변이 대상이 "임의의 한 엔트리"로만 지정(D-8) |
| Completeness | 0.75 | 0.75 | frontmatter 완전, Out of Scope 는 `### Out of Scope — <topic>` H3 3개 + 구체 불릿으로 `OutOfScopeRule` 규약을 정확히 만족, §7 미검증 항목을 스스로 열거한 점은 이 SPEC 의 강점이다. 감점: **HISTORY 섹션 부재**(D-6), §5 의 미해소 gap 을 열어 둔 채 진행(D-11 — 다만 내가 재보니 위험은 근거 없음, 아래 참조) |
| Testability | 0.75 | 0.75 | 14 AC 전부 Given-When-Then + 판정 명령 + RED 기준값 + 심각도 등급 + DoD 를 갖췄고 weasel word 가 없다. 그러나 **네 개의 구체적 구현이 AC 를 통과하면서 요구사항을 위반**한다(D-1..D-4). 이진 판정 가능성 자체는 유지되므로 0.50 밴드는 아니지만, 선언한 mutant 내성 축에서 새는 만큼 1.0 은 불가 |
| Traceability | 0.75 | 0.75 | §D.3 이 REQ 15 ↔ AC 14 전량 매핑을 싣고, 고아 AC 0건·미커버 REQ 0건을 확인했다. 감점 근거 1건: **`REQ-ZRR-003` 의 매핑이 간접적**이다 — AC-ZRR-004 는 anchor 가 해석되는지만 보고, "교리가 실제로 이사한 그 파일로 옮겼는지"는 판정하지 않는다. 규칙 밴드 0.75 의 정의("one AC references a REQ that exists but the mapping is indirect")에 정확히 해당. SPEC 이 §D.4 에서 이 사실을 숨기지 않고 적은 것은 정직하지만, 정직한 공백도 공백이다 |

**Aggregate = (0.75 + 0.75 + 0.75 + 0.75) / 4 = 0.75.** Tier M PASS 임계 0.80 미달.

---

## 리드가 지목한 7개 축 — 개별 판정

### 1. Mutant 내성 (최고 가치 축) — **미달**

14개 AC 각각에 대해 "이 AC 를 통과하면서 그 요구를 위반하는 구현"을 구성해 봤다. 통과하는 mutant 를 4개 찾았고, 그중 3개는 §G 에 행이 없다. D-1~D-4 로 아래 결함 목록에 기록한다.

§G 표 자체는 15행 중 11행이 유효한 방어를 제시한다 — 특히 AC-001 행(매처 약화 → 005/006/010 이 잡음), AC-014 행(규칙만 적고 다른 해석기를 쓰는 mutant → 004 가 잡음)은 정확하다. 문제는 남은 4행이 아니라 **표에 없는 행**이다.

### 2. RED baseline 정직성 — **통과** (7종 전량 독립 재현)

| AC | SPEC 이 적은 RED | 내가 실측한 값 | 일치 |
|---|---|---|---|
| AC-ZRR-001 | validate exit 1 / DRIFT 67 | `reconcile.py` → `validator_drift=67`, analyzer 68, 차이 1건 = `CONST-V3R5-009`(clause 값에 따옴표가 박혀 있어 YAML 파싱이 갈림) — SPEC 의 "analyzer 68 vs validator 67, 차이 1건은 파싱 아티팩트" 서술과 정확히 일치 | ✓ |
| AC-ZRR-002 | 로컬 리터럴 적중 25/101 | `cmp.py` → `validator_ok 33 / grepF_ok 25`, `grep_ok_but_not_validator 0` | ✓ |
| AC-ZRR-003 | 템플릿 리터럴 적중 25/101 | `tmplgrep.py` → `template_grepF_ok 25 of 101` | ✓ |
| AC-ZRR-004 | anchor 실패 17, 그중 clause 통과 9 | `analyze.py . ` → `entries=101 clause_fail=68 anchor_fail=17`, `clause_ok_anchor_fail=9` | ✓ |
| AC-ZRR-006 | 엔트리 101, raw `grep -c canary_gate` 는 104 | `grep -c '^- id:'` → `101`, `grep -c canary_gate` → `104` | ✓ |
| AC-ZRR-007 | `constitution validate` 배선 0건 | `grep -rn 'constitution validate' Makefile .github/workflows/` 무매치, `grep -c` → `0` | ✓ |
| AC-ZRR-008 | 유일한 constitution job 이 `continue-on-error: true` | `ci.yml:445` `constitution-check:` / `:449` `continue-on-error: true`, 스텝은 `constitution list` 뿐 | ✓ |
| AC-ZRR-011 | 두 미러 바이트 동일 | `diff -q` → 무출력(IDENTICAL) | ✓ |
| AC-ZRR-014 | 재사용 가능한 heading slug 헬퍼 부재 | `grep -rn 'func.*[Ss]lug' --include='*.go' internal/` → 5개 발견(SPEC 이 적은 3개 + `memoryProjectSlug`, `memorySlug`) — **전부 경로/i18n 키 slug 이고 heading 을 다루는 것은 0개**. 실질 주장은 성립, 열거만 불완전(D-10) | ✓(부분) |

**grep 형태 AC 의 0-히트 채택 조건**: 리드가 지목한 규율("사전 구현 트리에서 0을 반환해야 채택 가능")에 걸리는 AC 는 **AC-ZRR-012 하나**다. 세 grep 이 이미 `0 / 0 / 0` 이므로 이 AC 는 구현 전후로 아무것도 관측하지 않는다. SPEC 은 이 사실을 **먼저 밝히고**("비회귀 AC 다") 변이 기반 채택 근거를 제시했다 — 은폐가 아니다. 다만 그 변이가 AC 의 Then 절에 들어가 있지 않아, AC-ZRR-012 는 **아무것도 하지 않는 구현으로 충족된다**(D-3 계열, 심각도 minor로 별도 기록).

AC-ZRR-005/011 도 RED 가 "이미 참"이지만 이들은 보존 AC 이고 판정력이 사후 diff 에서 나오므로 같은 결함이 아니다.

재현 못 한 것 1건: **AC-ZRR-013(doctor Fail 0)** 의 RED 는 스크래치 프로젝트 `moai init` 이 필요해 이번 감사에서 재현하지 않았다. `validate` 층(67)이 재현됐고 doctor 가 그것을 그대로 싣는 구조라 신뢰하지만, **내가 관측한 값은 아니다**.

### 3. 가드-must-fail AC (AC-ZRR-007) — **부분 통과**

무엇을 변이시키는지(clause 한 글자 삽입 + anchor 1회 + 템플릿 미러 1회), 무엇이 관측인지(0 아닌 종료 코드 + **실패 메시지가 그 엔트리 ID 를 지목** + 그 출력을 `progress.md` §E.2 에 그대로 인용), 그리고 되돌린 뒤 재통과까지 명시돼 있다. "가드가 존재한다"는 서술로는 충족되지 않으며, `plan.md` §D10 과 §F M2 종료조건 6단계가 같은 것을 반복 고정한다. **여기까지는 잘 설계됐다.**

두 가지가 빠졌다. ① 변이 대상이 "임의의 한 엔트리"라 재현 가능한 지정이 아니다. ② 관측되는 것이 **가드 명령의 로컬 종료 코드**이지 CI job 의 결론이 아니다 — 이 틈이 D-3 를 만든다.

### 4. `continue-on-error: true` 구속 조건 — **주장 정확, AC 는 한 단계 약함**

`ci.yml` 실측: `constitution-check` job(`:445`)은 `continue-on-error: true`(`:449`)이고 스텝은 `constitution list` 두 개뿐이다. SPEC 의 주장은 정확하다.

권고 설계도 건전하다. `test` job(`:104`)이 `:183` 에서 `go test -coverprofile=... ./...` 를 돌리고 `continue-on-error` 가 없다 — 즉 진짜 차단 경로다. 게다가 `detect` job 의 `go_code` paths-filter 가 `.claude/rules/moai/**`, `CLAUDE.md`, `internal/template/templates/**`, `.moai/**` 를 이미 포함하므로, 인용 대상 17개 파일(16개가 `.claude/rules/moai/**`, 1개가 `CLAUDE.md` — 실측 열거함) 어느 것을 고쳐도 `test` job 이 돈다. **내가 의심했던 "markdown-only PR 이 Go 수트를 건너뛰어 가드가 안 돈다" 구멍은 오늘 트리에는 없다.**

그러나 AC-ZRR-008 의 Then 은 이 건전한 설계를 **요구하지 않는다**. 문자 그대로 읽으면 차단 job 안에 `|| true` 로 감싼 스텝도 만족한다(D-3). 그리고 `REQ-ZRR-003` 이 `file:` 재지정을 명시적으로 허용하므로, 수리가 필터 밖 경로(예: `.claude/skills/**`)로 엔트리를 옮기면 그 파일 편집에 대해 가드가 조용히 안 돈다 — 오늘의 결함은 아니고 **내일의 결함**이며, 어떤 AC 도 이를 막지 않는다(D-9).

### 5. 이중 트리 verbatim 구속 조건 (§5) — **사실관계 오류 발견, 그러나 gap 자체는 blocker 아님**

`file:` 17개를 로컬/템플릿에서 전수 diff 했다.

```
cited files: 17
divergent: 2
  differs .claude/rules/moai/development/coding-standards.md  entries: 2
  differs .claude/rules/moai/development/skill-authoring.md   entries: 1
```

`spec.md` §5 는 "17개 중 **3개**가 서로 다르다"라고 적었으나 **실측은 2개 파일 / 3개 엔트리**다. 3은 엔트리 수인데 파일 수 자리에 놓였다. 이어지는 "나머지 **14개** 파일(엔트리 98건)"도 산술이 어긋난다 — 17−2=**15**개 파일이며, 엔트리 98은 맞다. 정작 바로 위 표는 2행으로 올바르게 적혀 있어 **본문과 표가 서로를 반박**한다. 같은 오류가 `acceptance.md` AC-ZRR-003 으로 전파됐다(D-5).

**§7 이 열어 둔 gap("공통 구간이 실제로 존재하는지 세 엔트리 각각에 대해 아직 읽지 않았다")은 내가 대신 읽었고, 위험은 근거가 없다.**

- `coding-standards.md` 의 두 트리 차이는 **딱 한 줄**(로컬 `:124` 의 `git commit --no-verify` 불릿)이다. 영향 엔트리 2건은 `#language-policy` / `#thin-command-pattern` 절을 가리키며 그 한 줄과 무관하다 — 공통 구간이 파일의 거의 전부다.
- `skill-authoring.md` 의 차이는 3줄(`:352`, `:375-376`)이고 전부 SPEC-ID 중립화에서 온 것이다. 영향 엔트리 1건은 `#key-format-rules` 절을 가리키며 역시 무관하다.

즉 이 gap 은 **blocker 로 올라올 가능성이 사실상 없고**, 열어 둔 채 run-phase 에 넣는 판단은 정당하다. `plan.md` M1 이 "공통 구간이 없으면 blocker 보고(우회 금지)"라는 탈출로를 미리 못박아 둔 것도 적절하다. 이 축에서 감점하지 않는다 — 감점은 오직 숫자 오기(D-5) 몫이다.

### 6. 범위와 Tier — **통과, 압축 흔적 없음**

`spec-workflow.md` § SPEC Complexity Tier 실측: Tier M 의 ceiling 은 REQ **16** / AC **16** 이고 표가 둘을 **독립 열**로 싣는다(합산 아님). `plan.md` §A 의 해석이 맞다. 15/14 는 둘 다 여유 안에 있다.

리드가 경고한 "상한에 맞추려 요구를 병합·삭제한 흔적"을 찾으려 했고, **찾지 못했다.** 오히려 반대 신호가 있다:

- 세 축(clause 재동기화 / anchor 탐지·수리 / 가드)이 REQ 그룹으로 각각 온전히 남아 있다 — 데이터 수리 3, 보존 3, 가드 6, 배포 3.
- 병합했다면 가장 먼저 뭉쳤을 `REQ-ZRR-015`(slug 규칙 선언)가 `REQ-ZRR-009`(anchor 검증)에 흡수되지 않고 독립 요구로 서 있다. 이건 압축의 반대 방향 선택이다.
- `plan.md` §A 가 "여유가 REQ 1칸뿐"임을 **먼저 보고**하고 "들어가지 않았으면 Tier L 로 올렸을 것"이라고 적었다.

Tier M 자체도 타당하다 — 만지는 파일이 레지스트리 ×2 + 신규 테스트 + `ci.yml` + (선택) `Makefile` ≈ 5개로 M 의 5~15 구간이고, 15개를 넘지 않아 L 요건에 닿지 않는다.

### 7. 순서 구속 — **통과, 의존성으로 인코딩됨**

`plan.md` §F 가 `[HARD]` 태그로 "anchor 탐지 배선과 anchor 수리는 함께 착지한다. 나눌 경우 수리가 먼저"를 못박고, 이유까지 적는다 — 배선이 먼저 가면 17건이 깨진 채 새 가드가 즉시 붉어지고 그 붉음이 정상 신호와 구분되지 않아 "원래 붉은 것"으로 학습된다. 이어서 "M1 수리 → M2 가드 → M3 검증이며, 이는 되돌리기 어려움 순서가 아니라 **의존성**"이라고 성격을 명시한다. 권고가 아니라 의존성으로 읽히며, 리드의 판정과 정확히 일치한다.

덧붙여 §F 는 "가장 바뀔 가능성이 큰 결정(가드 표면)은 M2 로 미루지 않고 지금 확정한다"고 적고 M2 설계표로 대안 4종을 판정해 둔다 — 결정을 앞에, 착지를 뒤에 두는 형태로, 순서 규율을 지키면서 결정 리스크를 앞당긴 좋은 구성이다.

---

## Defects Found

**D-1** — `acceptance.md` AC-ZRR-001/002/003/004 (+ `plan.md` §G 표에 행 없음) — **자기참조 `file:` 재지정 mutant 가 BLOCKING AC 4건을 한 번에 통과시킨다.** 깨진 엔트리의 `file:` 을 `.claude/rules/moai/core/zone-registry.md`(레지스트리 자신)로 바꾸면, clause 값은 **정의상 그 파일 안에 리터럴로 존재한다**. 실측 확인: `grep -F -c -- '16-language neutrality' <registry>` → `1`, `grep -F -c -- 'Language-Aware Responses' <registry>` → `1`. 레지스트리에는 실제 heading 도 있으므로(`# Zone Registry`, `## HISTORY`, `## ID Allocation Policy`, `## Usage Guide` 등, `grep -c '^#'` → 50) anchor 도 해석된다. 두 미러가 바이트 동일하니 AC-ZRR-003 도 함께 통과한다. `AC-ZRR-006` 은 "달라진 필드는 `clause`/`anchor`/**`file`** 뿐"이라며 `file:` 변경을 **명시적으로 허용**하므로 막지 못한다. 결과: 수리를 한 줄도 하지 않고 validate exit 0 / 리터럴 101·101 / anchor 101 이 성립하며, 레지스트리는 아무것도 인용하지 않는 자기참조 껍데기가 된다. §G 는 "이 노트가 비면 그 AC 는 약한 것이다"라고 적었는데 이 mutant 에 해당하는 행이 없다 — Severity: **critical** — Class: **blocking** — Required fix: AC-ZRR-002/003 의 Then 에 두 줄 추가 — ① "어떤 엔트리의 `file:` 도 레지스트리 파일 자신이 아니다", ② "착지 전후로 `file:` 값이 바뀐 엔트리 목록(구 → 신)을 `progress.md` §E.2 에 인용하고, sync-phase 리뷰가 각 이동의 타당성을 판정한다". §G 에 대응 행을 신설할 것.

**D-2** — `acceptance.md` AC-ZRR-002/003 (`plan.md` §G AC-ZRR-006 행) — **빈 clause mutant 가 리터럴 체크를 그대로 통과한다.** AC 의 Then 은 "101건 전부가 적중한다"뿐이고, 빈 문자열은 어떤 파일에도 적중한다. 실측: `grep -F -q -- '' <file>` → rc `0`, Python `'' in text` → `True`. 즉 모든 clause 를 빈 문자열로 만들면 AC-ZRR-002/003 (둘 다 BLOCKING) 이 101/101 로 통과하고, `Validate` 도 `normalizedClause != ""` 가드 때문에 조용히 건너뛰어 AC-ZRR-001 까지 통과한다. §G 는 이 mutant 를 **정확히 식별해 놓고** 방어를 "이 요구를 M1 테스트 구현에 명시할 것"이라는 **plan 메모**로 남겼다 — AC 본문에는 들어가지 않았다. 식별된 mutant 가 AC 로 승격되지 않으면 방어가 아니다 — Severity: **major** — Class: **blocking** — Required fix: AC-ZRR-002/003 Then 에 "어떤 엔트리의 `clause:` 도 빈 문자열이 아니며, 빈 clause 는 적중이 아니라 **실패**로 센다" 한 줄 추가.

**D-3** — `acceptance.md` AC-ZRR-007 / AC-ZRR-008 (`plan.md` §G AC-ZRR-008 행의 방어 주장이 성립하지 않음) — **차단 경로 안의 `|| true` (또는 step-level `continue-on-error`) mutant 가 두 AC 를 함께 통과한다.** AC-ZRR-008 의 Then 은 "그 job 이 `continue-on-error: true` 가 아니거나, 가드가 기존 테스트 job 에서 실행된다"인데, `test` job 안에 `go test ./... || true` 로 넣은 구현은 후자를 **문자 그대로** 만족한다(job-level `continue-on-error` 는 없으므로 전자도 걸리지 않는다). §G 는 "007 이 종료 코드 관측으로 잡는다"고 적었으나, AC-ZRR-007 이 관측하는 것은 **로컬에서 돌린 가드 명령의 종료 코드**이지 CI job 의 결론이 아니다 — 로컬 `go test` 는 정상적으로 붉어지므로 007 도 통과한다. 결과적으로 `REQ-ZRR-007`("가드는 그 편집을 담은 PR 을 실패시킨다")을 검증하는 AC 가 **하나도 없다**. 같은 틈이 AC-ZRR-012 에도 있다(변이 근거가 Then 절 밖에 있어 "아무것도 안 하기"로 충족 — minor) — Severity: **major** — Class: **blocking** — Required fix: AC-ZRR-008 Then 을 "가드를 실행하는 **스텝**이 `|| true` / `continue-on-error` / 종료 코드 무시로 감싸이지 않았다"로 강화하고, AC-ZRR-007 에 "변이를 담은 커밋을 푸시해 해당 CI job 이 실제로 `fail` 로 결론나는 것을 `gh pr checks` 출력으로 관측·인용한다"를 추가.

**D-4** — `acceptance.md` AC-ZRR-005 / AC-ZRR-007 전반 — **가드가 평가한 엔트리 수를 단언하는 AC 가 없다.** AC-ZRR-006 은 *레지스트리에* 101 엔트리가 있음을 확인하지만, *가드가* 101건을 실제로 검사했음을 확인하는 AC 는 없다. 그 결과 가드 쪽 부분 순회·조기 반환·파일 제외 목록이 전부 살아남는다: AC-ZRR-007 의 변이 대상이 "**임의의** 한 엔트리"이므로, 제외 목록이 덮지 않은 엔트리에 변이가 떨어지면 가드는 정상적으로 붉어지고 AC 는 통과한다. §G 의 AC-ZRR-005 행("가드 쪽에 제외 목록을 넣는 mutant → 007/009 가 뚫고 실패해야 함")은 변이가 **제외된 엔트리에 떨어질 때만** 참이며, AC 가 그것을 요구하지 않는다 — Severity: **major** — Class: **blocking** — Required fix: AC-ZRR-007 에 "가드 출력이 평가한 엔트리 수를 보고하며 그 값이 `101` 이다(두 미러 각각)"를 추가하고, 변이 대상을 "임의의 한 엔트리"가 아니라 **명시된 ID 1건 + 무작위 1건**으로 고정.

**D-5** — `spec.md` §5 본문 / `acceptance.md` AC-ZRR-003 — **이중 트리 발산 파일 수 오기.** "17개 중 3개가 서로 다르다"로 적혔으나 실측은 **2개 파일 / 3개 엔트리**다(`coding-standards.md` 2엔트리, `skill-authoring.md` 1엔트리). 3은 엔트리 수인데 파일 수 자리에 놓였다. 파생 산술 "나머지 14개 파일"도 틀렸다 — 17−2=**15**(엔트리 98은 맞다). 바로 위 표는 2행으로 올바르므로 본문과 표가 서로를 반박하며, 같은 오기가 `acceptance.md` AC-ZRR-003 의 근거 문장으로 전파됐다. 이 SPEC 이 §D11 로 "숫자에는 측정 트리를 붙인다"를 스스로 규율한 것과 어긋난다 — Severity: **minor** — Class: **blocking**(명시된 구속 조건의 범위 서술이자 내부 모순, 수정 비용 2줄) — Required fix: §5 를 "17개 중 **2개 파일**이 다르고, 그 2개 파일이 **3개 엔트리**를 물고 있다 / 나머지 **15개** 파일(엔트리 98건)"으로 정정하고 `acceptance.md` AC-ZRR-003 도 함께 고칠 것.

**D-6** — `spec.md` (문서 전체) — **HISTORY 섹션 부재.** 표준 SPEC 섹션 중 HISTORY 가 없다(§1 문제 / §2 원인 / §3 대표 mutant / §4 요구사항 / §5 구속 조건 / §6 범위 밖 / §7 미검증). WHY(§1-2), WHAT(§4), Out of Scope(§6)는 이름만 다를 뿐 실질을 갖췄고 §6 의 H3 `### Out of Scope — <topic>` 3개는 lint 규약을 정확히 만족하므로 이 결함은 HISTORY 한 칸에 국한된다 — Severity: **minor** — Class: **optional** — Required fix: frontmatter 아래에 `## HISTORY` 를 추가하고 v0.1.0 → v0.2.0 변경 사유 한 줄을 남길 것.

**D-7** — `spec.md` §4 `REQ-ZRR-015` (및 제시 순서) — **복합 요구 + 시스템 거동 아닌 뒷절 + 번호 순서 어긋남.** `REQ-ZRR-015` 는 "가드는 slug 규칙을 선언한다"(Ubiquitous, 검증 가능) 와 "규칙이 바뀌면 그 변경은 구현 세부가 아니라 요구사항 변경으로 **취급된다**"(프로세스 서술, 시스템 거동 아님)를 한 요구에 담았다. AC-ZRR-014 는 앞절만 검증하며 뒷절은 어떤 AC 로도 판정되지 않는다. 또한 `REQ-ZRR-015` 가 §4 "가드" 절에서 `009` 와 `010` 사이에 배치돼 번호 순서가 어긋난다(집합은 완전하므로 MP-1 은 통과) — Severity: **minor** — Class: **optional** — Required fix: 뒷절을 `plan.md` §D8 구속 조건으로 이관(이미 거기 있다)하고 `REQ-ZRR-015` 를 선언 요구로만 남길 것. 번호는 `010` 뒤로 재배치.

**D-8** — `acceptance.md` AC-ZRR-007 — 변이 대상이 "**임의의** 한 엔트리"로만 지정돼 재현 가능한 판정이 아니다. 서로 다른 실행이 서로 다른 엔트리를 골라도 둘 다 "충족"으로 읽힌다 — Severity: **minor** — Class: **blocking**(D-4 의 수정에 흡수) — Required fix: D-4 와 함께 처리.

**D-9** — `spec.md` `REQ-ZRR-003` / `acceptance.md` AC-ZRR-008 — **재지정된 `file:` 이 CI paths-filter 밖으로 나갈 수 있다.** `REQ-ZRR-003` 은 교리 이사 시 `file:` 재지정을 명시적으로 허용한다. `detect` job 의 `go_code` 필터는 `.claude/rules/moai/**` · `CLAUDE.md` · `internal/template/templates/**` · `.moai/**` 를 덮지만 `.claude/skills/**` 는 덮지 않는다. 오늘의 17개 파일은 전부 필터 안이라 지금은 문제가 없으나(실측 열거 확인), 수리가 엔트리를 필터 밖 파일로 옮기면 **그 파일 편집에 대해 가드가 조용히 안 돌고** `test-skip-marker` 스텁이 초록을 보고한다 — `ci.yml:48-58` 주석이 기록한 #1557 실패 형태와 같다 — Severity: **major** — Class: **blocking**(수정 1줄) — Required fix: `plan.md` §D 에 구속 조건 추가 — "재지정된 `file:` 값은 `detect` job 의 `go_code` 필터가 덮는 경로 안이어야 한다. 밖으로 나가야 하면 같은 변경에서 필터를 확장한다."

**D-10** — `acceptance.md` AC-ZRR-014 RED 서술 — slug 헬퍼 열거가 불완전하다. 3개(`i18nSlug`, `slugify`, `projectSlug`)를 적었으나 실측은 5개(`internal/cli/memory.go:44 memoryProjectSlug`, `internal/cli/preference/cmd.go:159 memorySlug` 추가). 다섯 개 전부 경로/i18n 키 slug 이므로 **"heading 을 다루는 헬퍼가 없다"는 실질 주장은 성립**하고 AC 의 판정력에는 영향이 없다 — Severity: **minor** — Class: **optional** — Required fix: 열거를 5개로 보정하거나 "확인된 slug 헬퍼 전부가 경로·i18n 키 대상"으로 서술을 일반화.

**D-11** — `spec.md` §7 4번째 gap — **열어 둔 gap 이 실제로는 이미 닫혀 있다.** "3건 이중 트리 제약의 해소 가능성은 확인되지 않았다"고 적혔으나, 내가 두 파일을 diff 한 결과 발산은 `coding-standards.md` 1줄 / `skill-authoring.md` 3줄이고 셋 다 영향 엔트리의 인용 대상 절과 무관하다 — 공통 구간이 파일의 거의 전부다. **blocker 위험은 근거가 없으며, 이 gap 을 연 채 run-phase 로 가는 판단은 정당하다.** 결함이라기보다 정보 갱신이다 — Severity: **minor** — Class: **optional** — Required fix: §7 에 실측 결과("발산 총 4줄, 인용 대상 절과 무관 — 공통 구간 존재 확인")를 추가해 run-phase 가 같은 조사를 반복하지 않게 할 것.

---

## Regression Check

Iteration 1 — 이전 반복 없음.

---

## Recommendation

Tier M 이므로 재감사 기회는 **1회**(`plan_audit_tier_ceilings` S=1/M=2/L=3). iter2 는 아래 델타에만 한정되며, 전면 재감사가 아니다.

**blocking 6건 — 전부 AC/구속 조건 문안 수정이며 구조 변경이 아니다.**

1. **D-1** — `acceptance.md` AC-ZRR-002/003 Then 에 2줄: `file:` 이 레지스트리 자신이 아닐 것 + 변경된 `file:` 목록을 `progress.md` §E.2 에 인용하고 sync 리뷰가 판정할 것. `plan.md` §G 에 대응 행 신설.
2. **D-2** — AC-ZRR-002/003 Then 에 1줄: 빈 `clause:` 는 적중이 아니라 실패로 센다.
3. **D-3** — AC-ZRR-008 Then 을 스텝 단위로 강화(`|| true` / step-level `continue-on-error` 금지), AC-ZRR-007 에 CI job 결론 관측(`gh pr checks` 출력 인용) 추가.
4. **D-4 + D-8** — AC-ZRR-007 에 "가드가 평가한 엔트리 수 = 101(두 미러 각각)" 단언 추가, 변이 대상을 명시 ID 1건 + 무작위 1건으로 고정.
5. **D-9** — `plan.md` §D 에 구속 조건 1줄: 재지정된 `file:` 은 `detect` 의 `go_code` 필터 안에 머문다(밖이면 같은 변경에서 필터 확장).
6. **D-5** — `spec.md` §5 및 `acceptance.md` AC-ZRR-003 의 파일 수 정정(3→2 파일, 14→15 파일).

**optional 4건** (D-6 HISTORY, D-7 REQ-015 분리·재배치, D-10 열거 보정, D-11 §7 갱신) — 오케스트레이터 재량. 함께 처리하면 문서 정합성이 올라가지만 verdict 를 바꾸지는 않는다.

이 6건을 반영하면 Testability 와 Traceability 가 각각 1.0 / 0.75+ 로 올라가 aggregate 가 0.80 임계를 넘는다. **iter2 재감사는 위 6개 델타와, 수정이 기존 AC 를 약화시키지 않았는지에만 범위를 둔다.**

---

## 이 감사가 하지 않은 것 (Gaps)

- **AC-ZRR-013 의 RED 를 재현하지 않았다.** 스크래치 프로젝트 `moai init` + `moai doctor` 를 돌리지 않았으므로 `Pass 22 / Warn 2 / Fail 1` 은 레인 보고서를 근거로 인용했을 뿐 내가 관측한 값이 아니다.
- **나머지 14개 AC 를 실행해 보지는 않았다.** 판정한 것은 AC 의 *문안*이 mutant 를 막는지이며, 아직 존재하지 않는 구현의 거동이 아니다.
- **slug 규칙(§2.2 6단계)이 실제 markdown 렌더러의 anchor 해석과 바이트 단위로 같은지 확인하지 않았다.** SPEC 이 §7 에서 이미 미검증으로 선언한 항목이며, 나도 그것을 검증하지 않았다 — anchor 실패 17이라는 수치는 `analyze.py` 의 재구현 기준으로만 재현했다.
- **`.moai/reports/t232/analysis-devrepo.json` 은 읽지 않았다.** `analysis-repro.json` 과 크기가 같아(33955B) 동일 산출로 보이지만 확인하지 않았다.
