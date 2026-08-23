# SPEC Review Report: SPEC-TODO-ENABLE-FLAG-001

Iteration: 2/2 (Tier M ceiling)
Verdict: **PASS**
Overall Score: **0.87** (harmonic mean) — Tier M PASS 임계 **0.80**
Score trend: iter1 0.78 → iter2 0.87 (**회귀 없음** — LEAN STOP 에스컬레이션 조건 미충족)

Reasoning context ignored per M1 Context Isolation. 지시문의 측정치(REQ 6 / AC 11)는 입력이 아니라 재검증 대상으로 다뤘다.

Tree: `git rev-parse --short HEAD` → `0375e6842` — 핀과 일치.
독립 재측정: `grep -n '^### REQ-' spec.md` → 6건(REQ-1…REQ-6), `grep -c '^### AC-T-' acceptance.md` → **11**. 둘 다 Tier M 상한 16 이내. 지시문 측정치와 일치.

범위: iter1 보고서(`.moai/reports/t170/plan-audit-todo.md`)가 열거한 결함 델타 + 회귀 점검 + "개정이 깨뜨린 것" 탐색. iter1 보고서는 수정하지 않았다.

---

## 1. Claim (주장)

1. iter1의 블로킹 5건(D1·D2·D4·D5·D6)이 **모두 실제로 닫혔다** — 문구가 아니라 기제 수준에서.
2. 선택 4건(D3·D7·D8·D9) 중 3건이 닫혔고, D9는 `spec.md` 에서만 닫히고 `plan.md` 에 잔존한다.
3. 개정이 **새 결함 2건**을 들여왔다 — 둘 다 iter1이 벌한 것과 같은 계열(관측이 요구와 어긋남 / 관측이 부작용을 낳음)이며, 각각 한 줄 수정으로 닫힌다.
4. 형제 SPEC과의 공유 파일 표는 **소유권 축에서 모순이 없다** — 리드가 지정한 MUST-FIX 기준(마법사 질문·번역 항목·스키마 줄의 소유) 3축 전부 일치.
5. must-pass 7/7 통과, 총점 0.87 > 0.80 ⇒ PASS. 단 새 결함 2건은 run-phase 진입 **전에** 고쳐야 한다.

---

## 2. Evidence (증거)

### 2.1 Must-Pass Results (전량 재측정)

- **[PASS] MP-1 REQ 번호 일관성** — `spec.md:88,100,116,122,131,139` → REQ-1…REQ-6. 결번·중복·자릿수 불일치 없음.
- **[PASS] MP-2 GEARS 준수 (요구 계층 판정)** — 판정 대상은 `spec.md` 의 `REQ-XXX` 요구 계층이며 `acceptance.md` 의 Given-When-Then 11건은 **검증 계층**이라 여기서 채점하지 않았다(M3 § Scope). 각 REQ가 조건절 + `shall`/`shall not` 양태를 갖는다: REQ-1 `:90` "키가 설정에 없는 경우 … 해석돼야 한다" + `:94` "만들어서는 안 된다(shall not)"; REQ-2 `:102` "…`false`인 경우 … 출력해서는 안 된다(shall not)" + 신설 `:110` "…정상 동작해야 한다(shall) / 거부하거나 무시해서는 안 된다(shall not)"; REQ-3 `:118`; REQ-4 `:124` "`moai init`이 대화형으로 실행되는 경우"(event-driven); REQ-5 `:133`; REQ-6 `:141`. 신설 [HARD] 조항도 양태를 유지했다 — 개정이 GEARS를 훼손하지 않았다.
- **[PASS] MP-3 frontmatter** — `spec.md:2-17` 12개 정식 필드 전량 존재, 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `version: "0.2.0"` 은 D8 수정으로 **인용부호가 붙었다**(iter1의 유일한 스키마 이탈 해소).
- **[PASS] MP-4 언어 중립성** — `grep -rn 'SPEC-TODO-ENABLE-FLAG\|REQ-' internal/template/templates/.moai/config/sections/workflow.yaml internal/template/templates/.claude/skills/moai/SKILL.md` → 0건, AC-T-010(`acceptance.md:177`)이 그 기대를 고정. 언어별 도구명 0건. **주의**: 아래 N1이 이 축에 인접한 새 위험이나, MP-4의 정의(언어별 도구명 하드코딩)에는 해당하지 않으므로 must-pass를 실패시키지 않는다.
- **[PASS] MP-5 D7 cross-SPEC** — 참조 2건. `SPEC-FEEDBACK-AUTO-SUBMIT-001` → `status: draft`, `SPEC-KANBAN-TODO-CLI-001` → `status: in-progress`. 둘 다 retired/superseded/archived 아님 ⇒ BLOCKING 없음.
- **[PASS] MP-6 D8 cross-platform** — `grep -c 'syscall' spec.md` → `0`. auto-PASS(D8-4).
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-ENABLE-FLAG-001/ | wc -l` → `0`.

### 2.2 지시된 5건의 폐쇄 검증 (수용이 아니라 검증)

**D1 — 닫힘 (기제 확인).**
`grep -n '^func Test' internal/web/schema_label_test.go` → `TestSchemaEmptyLabelParity:16`, `TestI18nKeySetParity:74`, `TestI18nSegmentKeysRemovedFromWebDictionary:133`. `acceptance.md:156` 이 가리키는 `TestI18nKeySetParity` 는 **실재한다**. 이름만 바꾼 게 아니라 그 테스트가 실제로 신규 필드를 관측하는지 본문을 읽었다: `internal/web/schema_label_test.go:87` `for _, f := range settings.AllFields()` 를 돌며 각 필드의 `.title`/`.desc` 가 4로케일에 있는지 검사한다 — 신규 필드가 등록되면 자동으로 사정권에 든다. `internal/web` 쪽에서 무언가를 실제로 관측한다는 요구를 충족한다.
납품 테스트 축: `grep -c 'func TestWorkflowTodoEnabledFieldRegistered' internal/settings/schema_sections_test.go` → `0` — 아직 존재하지 않는 것이 **정상**이다. `acceptance.md:163` 이 "(M5 납품물)"로 명시하고 `plan.md` M5(§F, `schema_sections_test.go` 항목)가 단언 내용(존재 + `TypeBool` + `Persist.Kind == PersistSeam`)까지 규정한다. iter1의 공허함(존재하지 않는 이름을 `-run` 으로 가리켜 "no tests to run" 초록)과 구조가 다르다 — 이번엔 SPEC 자신이 그 이름을 만들어 내는 마일스톤을 갖고, `acceptance.md:203` 이 판정 전 `grep -c 'func <TestName>'` 확인을 명령한다.

**D5 — 실질 폐쇄, 부분 잔여.**
기록된 기준값이 AC 본문에 실제로 있다(`acceptance.md:97-99`) — "기록되지 않은 기준값에 대한 단언"이라는 되풀이 함정을 피했다. 세 명령을 현재 트리에서 그대로 실행:

```
grep -Fxc -e '  feedback, review, clean, codemaps, gate, e2e, harness, goal, todo) to' .claude/skills/moai/SKILL.md   → 1
grep -Fc  -e '- **todo** (aliases: backlog): Backlog queue — the slash surface covers two acts' …                    → 1
grep -Fc  -e '- Backlog language (add to the backlog, note this for later, …) routes to **todo**' …                  → 1
```

`sed -n '6p;81p;105p'` 로 읽은 실제 줄과 대조해 세 문자열이 그 줄들의 것임을 확인했다. working-tree `git diff` 의존이 사라졌으므로 커밋 이후에도 판정이 유지된다. `-e` 필수 주석(`:101`)도 사실이다 — 2·3번 패턴이 `-` 로 시작한다.
**잔여(N3)**: 1번만 `-Fxc`(줄 전체 일치)이고 2·3번은 `-Fc`(부분 문자열)다. 실제 81·105행은 기록된 문자열보다 길다(위 `sed` 출력) — 즉 그 줄의 **꼬리 부분이 편집돼도 관측되지 않는다**. "목록 3줄 불변"이라는 기대치의 2/3이 접두사 단언이다.

**D6 — 닫힘.**
`plan.md:11` 은 이제 "**Route B (PR) — 전 티어.**" 로 시작하고 `.claude/rules/local/repo-local-pr-policy.md` 를 근거로 든다. 그 규칙 파일을 직접 읽어 대조: *"[HARD] In THIS repository, the `spec-workflow.md` Route A … is DISABLED. `main` is protected with `enforce_admins: true` + required PR … ALL tiers (S / M / L) use **Route B (PR)**"* — `plan.md` 서술과 정확히 일치한다. iter1의 "Tier M이므로 Route A가 기본" 문장은 트리에서 사라졌다. `acceptance.md:217` 도 Closure Gate에 PR 경로 항목을 추가했다.

**D2 — 닫힘.**
`spec.md:110` 에 [HARD] 조항 신설: 억제 대상은 **추론 라우팅**이며, 사용자가 `/moai todo` 를 이름으로 직접 호출하면 "플래그 값과 무관하게 **정상 동작해야 한다**(shall). 거부하거나 무시해서는 안 된다(shall not)." 방향이 정해졌고(iter1이 권한 쪽 — REQ-3과 같은 근거), `:112` 가 왜 그 방향인지("이 결정이 없으면 구현자가 임의로 정하게 되고 CLI와 슬래시가 갈라진다")까지 남겼다.
관측: `acceptance.md:85` 이 AC-T-004에 한정 문장 관측을, `:109-111` 이 AC-T-005에 **행동** 관측(플래그 false에서 `todo add` → `todo` 왕복)을 추가했다. iter1이 지적한 "문장의 존재만 grep 한다"는 상태를 벗어나 실제 동작을 보는 축이 생겼다. **단 두 관측이 각각 새 결함을 하나씩 안고 있다(N1·N2 아래).**

**D4 — 닫힘, 그리고 지시된 가장 어려운 부분도 통과.**
`spec.md:182-190`: "충돌 없이 얹혀야 한다"가 사라지고 "**텍스트 충돌은 예외가 아니라 예상되는 결과다**" + 5조 해소 규칙으로 바뀌었다. 1조가 "두 번째로 착지하는 쪽이 충돌 해소 소유자다"로 소유자를 명명하고, 2조(양쪽 보존)·3조(재배치 금지)·4조(AC-T-011 재실행이 유일한 근거)·5조(실패 시 되돌리고 블로커 보고, 테스트를 고쳐 통과시키지 않음)가 따른다. `acceptance.md:195` 이 재실행 의무를 AC 본문에 못 박았다.
`depends_on` 근거 재작성 — **iter1의 논증에 직접 응답했다**. `spec.md:192`: *"이것은 '의존이 없다'는 관찰이 아니라 **선택**이다. `depends_on` 을 선언했다면 Phase 1 Depends_on Pre-flight 가 … run-phase 진입을 막아 두 SPEC을 **직렬화**했을 것이고, 위의 공유 파일 위험 9종이 통째로 사라졌을 것이다. 그 대신 **동시성을 택했다**"*, `:194`: *"기능 축에서 의존이 없다는 것 … 은 사실이지만, 그것만으로는 `depends_on` 생략을 정당화하지 못한다 — 생략이 사는 것은 동시성이고 치르는 것은 병합 충돌이다."* 기능 축 주장의 재진술이 아니라, iter1이 제기한 직렬화 대안을 명시적으로 평가하고 기각한 뒤 그 대가와 뒤집는 방법("양쪽 SPEC에 `depends_on` 을 추가하고 이 절을 함께 고친다")까지 적었다. 요구한 것을 그대로 했다.

### 2.3 선택 4건

- **D3 — 닫힘**: `spec.md:96` 이 잘못된 값 → 활성, 그리고 `loadWorkflowSection`(`internal/config/loader.go:226-237`)의 **섹션 단위** 폴백 폭발 반경을 기록. "판독의 *형태*를 계승하지 *기제*를 계승하지 않는다"는 문장으로 iter1이 지적한 미서술 괴리도 함께 해소. AC-T-001에 4번째 케이스 추가(`acceptance.md:42-44`).
- **D7 — 닫힘**: `spec.md:35` "**런타임/세션 표면 9개**" 로 한정 + `:70-72` 에 "배포 문서의 todo 서술" Out of Scope H3 신설.
- **D8 — 닫힘**: `version: "0.2.0"`.
- **D9 — `spec.md` 만 닫힘**: `spec.md:49` → `:95`, `:126` → `:87` 로 정정 확인. 그러나 `plan.md:19` 는 여전히 `translations_completeness_test.go:89`, `plan.md:20` 은 여전히 `questions_test.go:101`. **같은 결함이 같은 개정에서 한 파일에만 적용됐다**(N4, 선택).

### 2.4 개정이 깨뜨렸거나 들여온 것

**(a) 신규 AC의 중복·접힘 점검 — 문제 없음.**
AC 11건 전부를 읽었다. 중복 없음. 두 방향을 접은 AC는 3건이나 **셋 다 실패 방향을 분간할 수 있다**: AC-T-004는 5개 명령이 각각 다른 것을 관측하고(조건 문장 2사본 / 한정 문장 / 목록 3줄), AC-T-005는 서로 다른 명명 테스트 2개(`TestTodoCommandRegisteredRegardlessOfFlag` = 등록, `TestTodoVerbsUnaffectedByFlag` = 왕복), AC-T-009는 서로 다른 패키지의 테스트 2개(settings / web)로 갈린다. 어느 것이 실패해도 어느 방향이 깨졌는지 출력이 말해 준다.
매트릭스 대조(`acceptance.md:12-24`): REQ-1→001, REQ-2→002·003·004·005, REQ-3→005, REQ-4→006·007·008, REQ-5→009, REQ-6→010, 전체→011. 고아 AC 0건, 미커버 REQ 0건 — Traceability 유지.

**(b) 형제 SPEC 공유 파일 표 교차 검증 — 모순 없음(MUST-FIX 아님).**
이 SPEC `spec.md:168-178` vs 형제 `SPEC-FEEDBACK-AUTO-SUBMIT-001/spec.md:244-252`. 9행 전부 대조했다. 리드가 지정한 3축:

- 마법사 질문: 이쪽 `todo_enabled` 1개 / 형제 `feedback_auto_submit` 1개 — 양쪽 표가 서로의 열에 동일하게 기재. 일치.
- 번역 항목: 양쪽 모두 "ko/ja/zh 3블록 × 1쌍". 일치.
- 스키마 줄: 양쪽 모두 "필드 1줄". 일치.

`acceptance.md:191` 의 AC-T-011이 형제 테스트 이름 `TestFeedbackAutoSubmitQuestion` 을 부르는데, 형제 `acceptance.md:274` 가 같은 이름을 자기 납품물로 쓴다 — 이름도 일치한다.
**모순 아님이나 기록해 둘 비대칭 2건**: ① `shipped_key_inventory.yaml` 행에서 이 SPEC은 "(싣는 경우) 항목 1개"(조건부, M6 기본안은 싣지 않음)인데 형제 표의 대응 열은 "항목 1개"(무조건)로 적었다 — 소유권이 아니라 조건성의 불일치. ② 형제 §E.1은 **아직 iter1 문구**다("나중 것이 텍스트 충돌 **없이 얹혀야 한다**", `depends_on` 근거 = "기능 의존이 없다"). 즉 충돌 규율에서 두 SPEC이 갈렸다 — 이쪽이 강한 판본이고 정정 소관은 형제 쪽 감사관이다. 리드가 지정한 MUST-FIX 기준(질문·번역·스키마 줄 소유)에는 해당하지 않으므로 **이 SPEC의 MUST-FIX로 올리지 않는다**.

**(c) 새 결함 2건** — 아래 §5 N1·N2.

---

## 3. Baseline-attribution (baseline 귀속)

모든 수치는 이 트리(`0375e6842`), 이번 실행에서 관측했다. iter1 보고서의 수치는 재인용하지 않고 전부 다시 쟀다.

| 주장 | 명령 | 관측 |
|---|---|---|
| REQ 6건 | `grep -n '^### REQ-' spec.md` | 88,100,116,122,131,139 |
| AC 11건 | `grep -c '^### AC-T-' acceptance.md` | 11 |
| 웹 테스트 실재 | `grep -n '^func Test' internal/web/schema_label_test.go` | 16 / 74 / 133 |
| 납품 테스트 미존재(정상) | `grep -c 'func TestWorkflowTodoEnabledFieldRegistered' internal/settings/schema_sections_test.go` | 0 |
| D5 기준값 3건 | `grep -Fxc` / `grep -Fc` (AC 본문 그대로) | 1 / 1 / 1 |
| 목록 줄 실제 내용 | `sed -n '6p;81p;105p' .claude/skills/moai/SKILL.md` | 3줄 출력, 기록값이 각 줄의 접두 |
| PR 정책 | `sed -n '1,40p' .claude/rules/local/repo-local-pr-policy.md` | ALL tiers Route B, `enforce_admins: true` |
| SKILL.md 한국어 존재 | `grep -cP '[\x{AC00}-\x{D7A3}]' .claude/skills/moai/SKILL.md` (+ 템플릿 미러) | 12 / 12 (전부 `:302-320` 사용자 대면 오류 메시지) |
| 스킬 영어 규칙 | `grep -n "always English" CLAUDE.md` | `CLAUDE.md:111` "Commands/Agents/Skills instructions always English" |
| 큐 루트 해석 | `sed -n '35,105p' internal/cli/todo.go` | `resolveTodoQueueRoot` → git common dir, 실패 시 `~/.moai/todo/<key>` |
| 홈 격리 seam 관례 | `grep -rn 'userHomeDirFn =' internal/cli/*_test.go` | `todo_queue_root_test.go:122`, `glm_tools_test.go:35` |
| 형제 상태 | `grep -m1 '^status:' …/SPEC-FEEDBACK-AUTO-SUBMIT-001/spec.md` | `draft` |

---

## 4. Category Scores (0.0-1.0, rubric-anchored)

| Dimension | iter1 | iter2 | Rubric Band | Evidence |
|-----------|-------|-------|-------------|----------|
| Clarity | 0.75 | **0.85** | 0.75–1.0 | 유일했던 해석 여지(D2, 명시적 호출 미정의)가 `spec.md:110` [HARD]로 해소. 남은 감점은 요구 계층에 섞인 구현 세부 — REQ-4 `:128`(`saveBoolAnswer`/`WritePhase1Configs`/`yamlpatch.PatchFile` 경로 명시), REQ-5 `:133`(축자 Go 호출). WHAT이 아니라 HOW지만 두 곳 다 "따라야 할 살아 있는 경로"를 지목하려는 의도라 iter1과 같은 수준으로만 감점. |
| Completeness | 0.80 | **0.90** | 0.75–1.0 | 전 섹션 존재(HISTORY `§G:219`, WHY `§A:21`, WHAT `§B:51`, REQUIREMENTS `§C:86`, AC는 `acceptance.md`, Out of Scope H3 **5개** `:64,70,74,80,83` 각각 `-` 불릿 보유). iter1의 두 공백(D3 잘못된 값 / D4 해소 절차)이 모두 채워짐. 남은 감점: `plan.md` 인용 2건 미정정(N4). |
| Testability | 0.65 | **0.75** | 0.75 | iter1의 두 축이 모두 닫혔다 — 공허 `-run`(D1) 제거, 커밋 후 무력화되는 `git diff`(D5) 제거, 게다가 `acceptance.md:203-204` 가 두 실패 양식을 §D.3에 항구적 점검 항목으로 승격. 감점 사유는 신규 2건: AC-T-004의 관측이 한국어 리터럴에 묶여 준수 구현에서 실패한다(N1), AC-T-005 왕복이 격리 seam을 명명하지 않아 실제 홈을 오염시킬 수 있다(N2). 밴드 정의 "하나의 AC가 정밀하게 이진적이지는 않다"에 해당. |
| Traceability | 1.00 | **1.00** | 1.0 | 매트릭스 `acceptance.md:12-24` 재대조. 모든 REQ에 ≥1 AC, 모든 AC가 실재 REQ 지시. 고아 0, 미커버 0. 신규 조항(REQ-2 D2)도 AC-T-004·005로 양방향 연결(`spec.md:114`). |

Aggregate = harmonic mean(0.85, 0.90, 0.75, 1.00) = 4 / (1.17647 + 1.11111 + 1.33333 + 1.00000) = 4 / 4.62091 = **0.8656 → 0.87**

0.87 > Tier M 임계 0.80. 점수 회귀 없음(0.78 → 0.87)이므로 LEAN STOP 신호를 내지 않는다.

---

## 5. Defects Found (structured defect-list)

**N1. AC-T-004의 한정-문장 관측이 한국어 리터럴에 묶여 있다 — 준수 구현이면 실패하고, 통과시키면 규칙을 위반한다** — `acceptance.md:91` — Severity: **major** — Class: **blocking** — Required fix:
`grep -c '명시적' .claude/skills/moai/SKILL.md   # >= 1`. `.claude/skills/moai/SKILL.md` 는 **영어 지시문 파일**이고(`CLAUDE.md:111` "Commands/Agents/Skills instructions **always English**"), `internal/template/templates/.claude/skills/moai/SKILL.md` 로 미러돼 16언어 사용자에게 배포된다. 게다가 `moai update` 는 `.claude/skills/moai*` 글롭을 통째로 지우고 템플릿에서 재배포하므로(CLAUDE.local.md §2.3) 로컬 사본은 템플릿과 같아야 한다 — 즉 이 grep을 통과시키려면 한국어 지시 산문이 **배포 템플릿에 들어가야 한다**.
현재 SKILL.md의 한국어 12줄(`:302-320`)은 전부 사용자에게 그대로 출력되는 오류 메시지 텍스트로, 지시 산문과 계열이 다르다 — 선례로 쓸 수 없다.
결과는 양자택일이며 둘 다 나쁘다: 구현자가 영어로 한정을 쓰면(요구는 충족) MUST-PASS AC가 **거짓 실패**하고, 한국어로 쓰면 AC는 통과하되 영어 전용 규칙과 템플릿 중립성 방향을 위반한다. AC-T-010의 중립성 grep은 SPEC-ID/REQ 토큰만 보므로 이것을 잡지 못한다. iter1이 벌한 "관측이 요구와 어긋난 AC"와 같은 계열이다.
**수정**: 관측을 특정 언어 토큰이 아니라 행동/구조에 건다 — 영어 토큰으로 고정하거나(예: 라우팅 조건 문장 블록 안에 명시적 호출 보전 문장이 존재함을 영어로 관측), 이 관측을 AC-T-005의 행동 관측에 위임하고 AC-T-004에서 삭제한다. 어느 쪽이든 **한국어 리터럴은 제거**한다. 91행이 로컬 사본만 보는 점(두 사본 모두에 걸어야 함)도 함께 반영.

**N2. AC-T-005의 add→list 왕복이 격리 seam을 명명하지 않아 실제 사용자 홈을 오염시킬 수 있다** — `acceptance.md:106-111`, `plan.md` M3(`internal/cli/todo_test.go` 신규 테스트 2건) — Severity: **major** — Class: **blocking** — Required fix:
AC는 "프로젝트 루트는 `t.TempDir()`"만 규정한다. 그러나 `resolveTodoQueueRoot`(`internal/cli/todo.go:66-72`)는 프로젝트 루트를 쓰지 않는다 — `gitcore.ResolveGitDirs` 로 **primary 체크아웃**을 찾고, 실패하면 `fallbackTodoQueueRoot` → `userHomeDirFn()` → `~/.moai/todo/<key>` 로 간다. `t.TempDir()` 은 git 저장소가 아니므로 폴백을 타고, `userHomeDirFn` 을 덮어쓰지 않으면 테스트가 **개발자의 실제 홈에 큐를 만들고 카드를 남긴다**. 더 나쁜 경우 `adoptLocalTodoQueue` 가 기존 프로젝트 로컬 큐를 옮긴다.
`t.TempDir()` 만으로 충분하다는 `spec.md:202`(§E.2) 규정이 이 경로에서는 성립하지 않는다. 리포에 이미 관례가 있다 — `internal/cli/todo_queue_root_test.go:122` 와 `glm_tools_test.go:35-36` 이 `userHomeDirFn` 을 덮고 `t.Cleanup` 으로 복구한다.
**수정**: AC-T-005의 Given에 "`userHomeDirFn` 을 `t.TempDir()` 로 대체하고 `t.Cleanup` 으로 복구한다"를 한 줄 추가하고, `plan.md` M3의 신규 테스트 항목에 같은 seam을 명시한다. `spec.md` §E.2의 "신규 테스트는 `t.TempDir()`" 도 이 경로에 한해 홈 seam이 추가로 필요함을 한 줄로 보완하면 더 낫다.

**N3. D5의 내용 단언 3건 중 2건이 줄 전체가 아닌 접두사 일치다** — `acceptance.md:98-99` — Severity: **minor** — Class: **optional** — Required fix: 1번만 `-Fxc`(줄 전체)이고 2·3번은 `-Fc`(부분 문자열)이며, 실제 81·105행은 기록된 문자열보다 길다. 따라서 그 두 줄의 **꼬리가 편집돼도 통과**한다 — "목록 3줄 불변"이라는 기대치의 2/3만 실제로 고정된다. iter1 결함의 재발은 아니고(커밋 후에도 판정이 유지되는 성질은 확보됨) 그물이 성긴 것이다. 세 줄 전문을 기록해 셋 다 `-Fxc` 로 통일하거나, 접두사 단언임을 AC 본문에 명시해 기대치를 정직하게 좁힌다.

**N4. D9 정정이 `spec.md` 에만 적용되고 `plan.md` 에 남았다** — `plan.md:19`, `plan.md:20` — Severity: **minor** — Class: **optional** — Required fix: `plan.md:19` 는 여전히 `translations_completeness_test.go:89`(실제 `:95`), `:20` 은 여전히 `questions_test.go:101`(실제 `:87`)를 인용한다. `spec.md:49`·`:126` 은 정정됐다. 두 줄 수정.

**N5. AC-T-001의 소제목이 케이스 수와 어긋난다** — `acceptance.md:30` — Severity: **minor** — Class: **optional** — Required fix: 소제목은 "키 해석 3케이스"인데 매트릭스(`:14`)와 본문(`:42-44`, `:47`)은 4케이스다. D3 추가 시 소제목을 갱신하지 않았다. 한 단어 수정.

**N6. 형제 SPEC 표의 인벤토리 행 조건성이 비대칭이다 (교차 관측, 이 SPEC의 MUST-FIX 아님)** — `spec.md:178` vs `SPEC-FEEDBACK-AUTO-SUBMIT-001/spec.md:252` — Severity: **minor** — Class: **optional** — Required fix: 이 SPEC은 `shipped_key_inventory.yaml` 항목을 "(싣는 경우)" 조건부로, 형제 표의 대응 열은 무조건으로 적었다. M6 기본안이 "템플릿에 싣지 않음"이므로 실제로는 0건이 될 공산이 크다. 질문·번역·스키마 줄의 **소유권 축에는 모순이 없으므로** 리드가 지정한 MUST-FIX 기준에 해당하지 않는다. 어느 한쪽 표의 조건성을 맞추면 된다.

---

## 6. Regression Check (iter1 결함 대조)

| iter1 | Class | 상태 | 근거 |
|---|---|---|---|
| D1 공허 AC-T-009 | blocking | **RESOLVED** | `TestI18nKeySetParity` 실재(`:74`) + `AllFields()` 순회 본문 확인; 납품 테스트 이름·단언 내용이 `plan.md` M5에 규정 |
| D2 명시적 호출 미정의 | blocking | **RESOLVED** | `spec.md:110` [HARD] 방향 확정 + AC-T-004/005 관측 (단 N1·N2 동반) |
| D4 충돌 규율/`depends_on` | blocking | **RESOLVED** | 5조 해소 규칙 `:184-190`, 트레이드오프 재작성 `:192-194` — iter1 논증에 직접 응답 |
| D5 커밋 후 공허 AC-T-004 | blocking | **RESOLVED**(부분 잔여 N3) | 기준값 기록 + 현재 트리에서 1/1/1 실측 |
| D6 잘못된 PR 경로 | blocking | **RESOLVED** | `plan.md:11` Route B 전 티어, 규칙 파일 원문 대조 |
| D3 잘못된 값 | optional | **RESOLVED** | `spec.md:96` + AC-T-001 4번째 케이스 |
| D7 표면 수 무한정 | optional | **RESOLVED** | `spec.md:35` 한정 + `:70-72` Out of Scope 신설 |
| D8 `version` 미인용 | optional | **RESOLVED** | `version: "0.2.0"` |
| D9 드리프트 인용 2건 | optional | **PARTIAL** | `spec.md` 정정, `plan.md:19-20` 잔존(N4) |

정체(stagnation) 없음 — 두 이터레이션에 걸쳐 불변인 결함이 존재하지 않는다.

**iter1이 "유지 판정"으로 인용했던 것들의 회귀 점검**: 범위 경계 서술(`spec.md:33`, `:42`, `:64-68`, `:206`, `acceptance.md:5`, `:202`, `:214`) 전부 잔존. 대조 케이스(AC-T-002 `:56-58`, AC-T-003 `:69-74`) 잔존. AC-T-008의 사장 코드 방어(`:135-144`) 잔존. 충족 불가능한 문구의 AC 0건 — 11개 AC 전량 재독으로 확인. 개정이 되돌린 것은 없다.

---

## 7. Gaps (미검증)

- **납품 테스트 3종의 실제 통과 여부**는 검증하지 않았다 — `TestWorkflowTodoEnabledFieldRegistered`, `TestTodoCommandRegisteredRegardlessOfFlag`, `TestTodoVerbsUnaffectedByFlag` 는 아직 존재하지 않는 run-phase 산출물이다. 검증한 것은 "이름이 실재 파일에 없다"와 "SPEC이 그 이름을 만들 마일스톤과 단언 내용을 갖는다"까지다.
- **AC 명령을 실행해 통과를 관측하지 않았다** — plan-phase 감사이며 구현이 없다. 실행한 것은 D5 기준값 3건과 구조 측정뿐이다.
- **N1의 "영어로 쓰면 AC가 실패한다"는 추론이다** — 구현이 없으므로 실패를 관측한 것이 아니라, `grep -c '명시적'` 이 한국어 리터럴을 요구한다는 사실과 SKILL.md가 영어 지시문이라는 규칙에서 도출했다.
- **N2의 오염을 재현하지 않았다** — `resolveTodoQueueRoot` 의 폴백 경로를 코드로 읽어 확인했을 뿐, 테스트를 작성해 실제 홈 쓰기를 관측하지는 않았다(관측하려면 실제 홈을 오염시켜야 한다).
- **형제 SPEC은 공유 파일 표와 §E.1만 읽었다** — 리드가 지정한 교차 검증 범위 밖(마스킹·취약점 분류 등)은 보지 않았다. 형제 SPEC의 자체 감사는 다른 감사관 소관.
- **`spec_audit` / `spec_drift` 도메인 도구를 돌리지 않았다** — 이번 판정은 어떤 lifecycle drift 주장도 하지 않으므로 필요하지 않았다.
- **cross-model 2차 의견 없음** — 프로젝트 `audit_model` 이 `multi` 임을 확인하지 않았고 백엔드를 호출하지 않았다. 이 판정은 단일 Claude 감사다.

---

## 8. Residual-risk (잔여 위험) — PASS가 run-phase에 넘기는 부채

1. **N1·N2는 PASS와 함께 그냥 넘어가지 않는다.** 총점(0.87)과 must-pass(7/7)가 PASS를 지시하므로 점수를 깎아 FAIL로 뒤집지 않았다(M6 — 결함 나열로 FAIL을 제조하지 않는다). 그러나 둘은 **blocking 계열**이며 각각 한 줄 수정이다. run-phase 진입(Implementation Kickoff Approval) **전에** 고쳐야 하며, 고치지 않으면 N1은 MUST-PASS AC의 거짓 실패나 배포 템플릿 언어 규칙 위반 중 하나를, N2는 개발자 홈 오염을 실제로 일으킨다.
2. **병합 충돌은 여전히 예상되는 결과다.** 해소 규칙이 생겼을 뿐 충돌 자체는 그대로다. 두 번째 착지자가 5조를 지키는지는 AC-T-011 재실행으로만 관측되며, 그 재실행을 **누가 강제하는지**는 규약 수준이다 — 재실행 없이 "해소 완료"를 보고하는 경로가 기계적으로 막혀 있지 않다.
3. **형제 SPEC의 §E.1이 아직 iter1 문구다.** 두 SPEC이 같은 충돌 상황에 서로 다른 규율을 적는 상태이며, 두 번째 착지자가 형제 쪽 문서를 읽으면 "충돌은 일어나지 않아야 한다"는 낡은 기대를 받는다. 형제 감사에서 정렬되지 않으면 run-phase에서 리드가 이쪽 판본을 정본으로 지정해야 한다.
4. **N3의 접두사 단언**은 목록 줄의 꼬리 편집을 놓친다 — 범위 밖 위반을 잡는 그물이 의도보다 성기다.
5. **M6의 템플릿 결정이 아직 미정이다.** AC-T-010의 기대값이 결정에 종속되며(`acceptance.md:180`), 결정을 커밋 메시지에 남기는 것이 유일한 고정 장치다. 결정 자체는 run-phase로 이월된다.
6. **억제 범위가 부분적이라는 사실**은 해소되지 않았고 해소할 수도 없다(§A.1 P4). SPEC은 이를 정직하게 경계로 선언했고 완료 보고 문구까지 규정했다(`acceptance.md:214`) — 이 부채는 설계된 것이지 결함이 아니다.

---

VERDICT: PASS 0.87

**Tier M PASS 임계 0.80** 대비 +0.07. must-pass 7/7 통과, 점수 회귀 없음(0.78 → 0.87), 정체 결함 없음. iter1의 블로킹 5건은 문구가 아니라 기제 수준에서 닫혔음을 각각 별도 명령으로 확인했다.

**run-phase 진입 전 필수 수정 2건**(verdict를 뒤집지는 않으나 넘길 수 없는 부채, 각 한 줄):

1. **N1** — `acceptance.md:91`: `grep -c '명시적'` 을 언어 중립적 관측으로 교체(또는 AC-T-005에 위임하고 삭제). 한국어 리터럴을 영어 전용·템플릿 미러 스킬 본문의 통과 조건으로 두지 않는다.
2. **N2** — `acceptance.md:106-111` + `plan.md` M3: `userHomeDirFn` 격리 seam을 명명한다(`t.TempDir()` 만으로는 `resolveTodoQueueRoot` 폴백을 막지 못한다). 선례 `internal/cli/todo_queue_root_test.go:122`.

**선택**(오케스트레이터 재량): N3(접두사 단언 → 줄 전체), N4(`plan.md:19-20` 인용 정정), N5(AC-T-001 소제목 "3케이스" → "4케이스"), N6(형제 표 인벤토리 행 조건성 정렬).
