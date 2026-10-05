# SPEC Review Report: SPEC-WORKTREE-EXIT-RETURN-001

Iteration: 1/1 (Tier S ceiling = 1, per `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL**
Overall Score: **0.69** (조화평균) — Tier S PASS 문턱 **0.75** 미달
(문턱 출처: `.claude/rules/moai/workflow/spec-workflow.md:138` — Tier S 행 "plan-auditor PASS threshold" = 0.75)

> **Reasoning context ignored per M1 Context Isolation.** 배차문이 서술한 SPEC 의 결론·분류·「line 220 이 이러이러하다」는 주장은 저자 측 추론으로 취급해 채택하지 않았고, 전부 파일에서 직접 재측정했다. 배차문에서 받아들인 것은 **감사 방향**(어디를 의심할지) 뿐이다.

## 측정 환경 (baseline attribution)

| 항목 | 값 |
|---|---|
| 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t965` (`git rev-parse --show-toplevel`) |
| 브랜치 | `WT-exitworktree-return` |
| 감사 시작 HEAD | `d20a9f022` |
| 감사 종료 HEAD | `ce2e35355` — **감사 창 안에서 이동함, 아래 § 절차 결함 참조** |
| 측정 일자 | 2026-09-19 |

감사 대상 경로는 두 HEAD 사이에서 변하지 않았다. 측정:

```
git diff --stat d20a9f022..ce2e35355 -- \
  .claude/rules/moai/workflow/worktree-integration.md \
  internal/template/templates/ internal/template/internal_content_leak_test.go .moai/specs/
→ (출력 없음)
```

따라서 아래 측정값은 두 커밋 모두에 귀속된다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-WXR-001` ~ `REQ-WXR-008`. 측정: `grep -o 'REQ-WXR-[0-9]*' spec.md | sort -u` → 8개, 결번·중복 없음, 3자리 제로패딩 일관. REQ 총수 8 = Tier S 상한 8 (`spec-workflow.md:146`) 이내.
- **[PASS] MP-2 GEARS 형식 준수** — **요구사항 층(`spec.md` 의 `REQ-XXX`)에 대해 판정함.** 8개 전부 다섯 GEARS 패턴 중 하나에 대응한다: 001 Ubiquitous(`spec.md:94`), 002 Event-driven(`:98`), 003 Event-driven(`:103`), 004 Ubiquitous(`:108`), 005 Where/capability-gate(`:113`), 006 Unwanted — `shall not` 정칙 부정형(`:118`), 007 While/state-driven(`:122`), 008 Ubiquitous(`:127`). 주어는 일반화된 `the doctrine` 으로 일관. `acceptance.md` 의 Given-When-Then 항목은 **검증 층**이므로 M3 § Scope 에 따라 여기서 감점하지 않았고 Group 4 에서 채점했다. 폐기 예정 `IF/THEN` 구문 0건.
- **[PASS] MP-3 YAML frontmatter 유효성** — `.claude/rules/moai/development/spec-frontmatter-schema.md` 의 canonical 12 필드를 항목별로 대조: `id`·`title`·`version`(`"0.1.0"` 인용 semver)·`status`(`draft`, 8-값 enum 내)·`created`·`updated`(ISO)·`author`·`priority`(`P2`)·`phase`(`"v3.1.4 target"` — 금지 라이프사이클 토큰 `plan/run/sync/mx` 아님)·`module`·`lifecycle`(`spec-anchored`)·`tags` 전부 존재·형 일치. 13번째 `tier: S` 는 스키마 `:167` 에 등재된 선택 필드이므로 위반 아님. 거부 대상 snake_case 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건.
- **[N/A] MP-4 §22 언어 중립성** — 이 SPEC 은 프로그래밍-언어 도구 이름을 전혀 담지 않는 문서-계약 SPEC 이다. 기준 비적용 → auto-pass. (별개 축인 **템플릿 내부-내용 중립성** 위반은 실재하며 D1 에 기록했다 — MP-4 와 다른 축임을 명시한다.)
- **[PASS] MP-5 D7 교차-SPEC 조정** — 검증 동사 실행함. `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` 적중은 자기 자신(`SPEC-WORKTREE-EXIT-RETURN-001`) 뿐이며 외부 SPEC 참조 0건 → `retired/superseded/archived` 조정 의무가 발생하지 않는다. BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c 'syscall' spec.md` → `0`. D8-4 에 따라 auto-PASS.
- **[PASS] MP-7 clarification 게이트** — `grep -rn '\[NEEDS CLARIFICATION' spec.md plan.md acceptance.md` → exit 1, 적중 0건. `research.md` 는 Tier S 이므로 부재(정상).

**must-pass 7개 전부 통과.** 따라서 이 FAIL 은 방화벽이 아니라 **루브릭 점수와 열거된 결함**이 만든 것이다.

---

## Category Scores (0.0–1.0, 루브릭 기준)

| 축 | 점수 | 루브릭 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 대부분 단일 해석. 다만 `spec.md:94-96` REQ-001 의 "independent of chain depth / launch directory" 가 표본 범위 주장인지 전칭 주장인지 구분되지 않고(D4), `acceptance.md:30-31` 의 「무자격·단독」이 판단을 요구한다(D7). |
| Completeness | 0.75 | 0.75 | 필수 절 전부 존재 — HISTORY(`:19`), WHY(`:27`/`:45`), WHAT(`:63`/`:89`), REQUIREMENTS(`:89`), ACCEPTANCE(`:131` + `acceptance.md`), Out of Scope(`:138`, H3 3개 각각 `-` 불릿 보유). frontmatter 완전. 감점 사유: `spec.md:86` 이 「그대로 옮겼다」고 선언한 Gaps 를 실제로는 낡은 판으로 옮겼다(D3). |
| Testability | 0.70 | 0.50–0.75 사이, 0.75 쪽 | AC-004 는 모범적(비공허·값 미리단정 없음). 다만 AC-002(`:22-23` "구분을 선언하는 문장이 있다")·AC-003(`:30-31` "무자격 단독")·AC-006(`:61-62` "「재현되지 않았다」류 문구") 세 항목이 판정자 재량을 남긴다. 각각 인용 문자열이라는 구체 앵커를 갖고 있어 0.50 밴드보다는 낫다고 보아 상향 조정했다. |
| Traceability | 0.60 | 0.50–0.75 사이 | `REQ-WXR-007` 을 덮는 AC 가 **없다**(D6). 더하여 `acceptance.md` 전체에서 `REQ-WXR` 문자열 적중이 **0건** — 7개 AC 어느 것도 자기 REQ 를 명시 인용하지 않으며 대응은 주제 추론으로만 성립한다. 「REQ 1개 미커버」(0.75 밴드)보다 나쁘고 「다수 미커버」(0.50 밴드)까지는 아니다. |

조화평균 = 4 / (1/0.75 + 1/0.75 + 1/0.70 + 1/0.60) = 4 / 5.7619 = **0.694**.
(산술평균으로 계산해도 0.700 으로 문턱 미달이며, 결론은 평균 방식에 좌우되지 않는다.)

---

## Defects Found (구조화 결함 목록)

**D1.** template-neutrality-collision — `.moai/specs/SPEC-WORKTREE-EXIT-RETURN-001/acceptance.md:L52-54, L68-69` + `spec.md:L127-129` — **인수조건이 CI 가드가 잡도록 설계된 위반을 지시한다.** AC-WXR-005 는 개정 절이 「(b) 측정 날짜」를 담을 것을 요구하고, AC-WXR-007 의 판별식은 "AC-WXR-001 · AC-WXR-005 의 판별 문자열이 **두 사본 모두에서** 적중한다" 이며, REQ-WXR-008 이 템플릿 미러 동시 반영을 의무화한다. 즉 `2026-09-19` 가 `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` 에 들어가는 것이 **완료 조건**이다. 측정: strict-tier 클래스 `S1-internal-date` 의 패턴 `\b202[5-9]-[0-1][0-9]-[0-3][0-9]\b` (`internal/template/internal_content_leak_test.go:L364`)가 그 문자열에 적중한다. 면제 경로 둘 다 닿지 않는다 — DC-1 은 unfenced frontmatter `updated:` 한정, DC-4 는 `dc4AttributionFiles = [".claude/rules/moai/NOTICE.md"]` 한정(`:L956-958`). 현재 미러의 날짜 리터럴은 0건(`grep -nE '20[0-9]{2}-[0-9]{2}-[0-9]{2}'` → 무출력)이고, 두 티어 baseline 은 green 이다(`go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1` → `ok`, exit 0; `MOAI_TEMPLATE_LEAK_STRICT=1` 동일 → `ok`, exit 0). CI 는 `.github/workflows/template-neutrality-check.yaml` 에서 `MOAI_TEMPLATE_LEAK_STRICT: '1'` 로 이 테스트를 돌리며 트리거 경로에 `internal/template/templates/**` 가 포함된다. 즉 SPEC 을 문언대로 이행하면 **새 적중이 생기고 CI 가 적색이 된다.** — Severity: **critical** — Class: **blocking** — Required fix: AC-WXR-007 의 판별식에서 AC-WXR-005 의 날짜 문자열 요구를 제거하고, 귀속(날짜·기록 경로)은 **로컬 사본 전용**으로 범위를 좁힌다. 미러는 날짜 없는 중립 서술을 담는다. 굳이 미러에도 날짜를 넣겠다면 `dateAllowlist` 에 (파일, 날짜) 항목을 **선언된 예외로** 추가하는 변경을 SPEC 범위에 명시적으로 넣어야 하며, 현재 §4 Out of Scope 의 「코드 층 무변경」(`spec.md:L158-161`)과 충돌하므로 그 배제 항목도 함께 고쳐야 한다.

**D2.** cross-round-version-attribution — `acceptance.md:L52-53` — **한 귀속 3항 중 두 개와 나머지 하나가 서로 다른 측정 회차 소속이다.** (a) `.moai/reports/t965/observations.md` 와 (b) 측정 날짜는 2026-09-19 관측 회차의 것이고, (c) 는 명문으로 "**AC-WXR-004 가 기록한** 런타임 버전" — 즉 아직 수행되지 않은 재측정 회차의 것이다. 세 항목이 한 묶음으로 제시되면 독자는 하나의 귀속으로 읽는다. 재측정이 더 새 빌드에서 돌면 독트린은 「2026-09-19 에 관측, 런타임 <더 새 버전>」이라는 **거짓 귀속**을 싣게 된다. `verification-claim-integrity.md` §2 가 금지하는 「다른 시점의 값을 이 측정의 baseline 으로 쓰는」 모양이다. 덧붙여 `observations.md:L131` 은 관측 1~3 의 버전 `2.1.278` 을 **이미 기록하고 있는데** AC-005 는 그것을 쓰지 않는다. — Severity: **major** — Class: **blocking** — Required fix: 귀속을 회차별로 분리한다. 관측 1~3 서술에는 `observations.md` 자신의 `2.1.278`(사후 귀속이라는 단서 포함)을, 재측정 서술에는 AC-WXR-004 의 버전을 붙인다. 두 회차를 한 3항 묶음으로 합치지 않는다.

**D3.** stale-gap-carriage — `spec.md:L86` vs `.moai/reports/t965/observations.md:L114-116, L128-141` — **「그대로 운반한다」고 선언한 절이 근거 파일의 낡은 판을 운반한다.** `spec.md:L78-79` 는 §1.4 가 조사 파일의 Gaps 를 "**그대로** 옮긴 것"이라고 선언한다. 그러나 `spec.md:L86` 은 "**런타임 버전이 기록되지 않았다** — 관측은 버전에 귀속되지 않는다" 인 반면, 근거 파일은 "관측 **도중**에 재지 않았다. 리드 지시로 **사후 기록했고** … **사후 귀속으로 약화된 채 남는다** — 특히 관측 4 는 그 귀속조차 공유하지 않는다"(`:L114-116`)이고 `2.1.278` 을 표로 싣고 있다(`:L128-133`). 「미기록」과 「사후 기록되어 약화됨」은 다른 상태다. 이 오기는 무해하지 않다 — REQ-WXR-005(`spec.md:L113-116`)가 런타임 버전 귀속을 **요구**하므로, §1.4 를 읽은 이행자는 존재하지 않는다고 적힌 값을 요구받는 모순에 놓이고, 그 모순의 회피책이 D2 의 교차-회차 귀속이다. — Severity: **major** — Class: **blocking** — Required fix: `spec.md:L86` 을 근거 파일의 실제 문언으로 교체한다 — 「사후 기록되어 귀속이 약화됨, 관측 4 는 그 귀속조차 공유하지 않음」.

**D4.** independence-overclaim — `spec.md:L94-96` (REQ-WXR-001) — **증거가 표본 2점·1점인 두 축에 대해 「무관함」을 서술하라고 지시한다.** 문언은 "shall state that the observed return point **is independent of** the `EnterWorktree` chain depth **and of** the session's launch directory" 이다. 여기서 `observed` 는 `return point` 를 수식할 뿐 `independent` 를 한정하지 않는다. 실제 표본: 연쇄 깊이는 1(관측 1)·2(관측 2) **두 점**뿐이고, 세션 출발 디렉터리는 **비-primary 한 값 하나**(관측 4)뿐이며 그 관측은 재실행 불가(`observations.md:L110-111`)이자 버전 귀속 미공유(`:L139-141`)다. 근거 파일 자신은 "…고정되는 것으로 **보인다**"(`:L80-81`)로 유보했고 `spec.md:L41-43` 도 같은 유보를 쓰는데, 정작 규범 문장인 REQ-001 이 이를 성질 주장으로 굳힌다. 결정적으로 **REQ-WXR-005 의 Where 가드가 이 문장을 덮지 못한다** — 가드 조건은 "rests on the **upstream tool description**" 인데 독립성 주장은 상위 설명이 아니라 관측에 기대므로 가드가 발화하지 않는다. SPEC 에서 가장 강한 일반화가 하필 귀속 요구의 사각에 앉아 있다. 또한 이 주장은 REQ-WXR-006 이 열어 두라고 명령한 agent-5 반대 제보(`observations.md:L106-109`)에 대한 반증으로 기능하면서 그 사실을 말하지 않는다 — 배차가 물은 「다른 요구사항이 원 제보가 해명된 것처럼 가정하는가」의 답이 여기다. — Severity: **major** — Class: **blocking** — Required fix: 표본을 문장에 실어 한정한다. 예: 「관측된 깊이 1·2 에서, 그리고 관측된 단일 출발-디렉터리 사례에서 복귀 지점은 동일했다」. `independent of` 라는 전칭 술어를 쓰지 않는다.

**D5.** ordering-not-witnessable — `plan.md:L46-52` — **M1→M2 순서가 산문으로만 단언되고 커밋 그래프로 검증되지 않는다.** `plan.md:L51-52` 는 순서의 이유를 정확히 댄다("문서를 먼저 쓰고 나중에 재는 순서는, 측정이 문서를 확인하는 도구로 전락한다"). 그런데 어떤 AC 도 M1 기록물의 커밋이 M2 독트린 커밋의 **조상**일 것을 요구하지 않는다. AC-WXR-007(`acceptance.md:L64-69`)은 독트린과 미러가 **같은** 커밋에 있을 것만 요구하고 M1 기록의 커밋에 대해 침묵한다. AC-WXR-005(c) 가 만드는 데이터 의존(독트린이 AC-004 의 버전을 인용)은 약한 함의일 뿐이며, 둘이 한 커밋에 같이 들어가면 git 은 커밋 내부의 저작 순서를 원리상 증언하지 못한다 — `verification-claim-integrity.md` §2.3 의 정확한 모양이다. `plan.md:L82-83` 이 VCI §2 를 교차참조하면서 §2.3 을 놓쳤다. — Severity: **major** — Class: **blocking** — Required fix: AC 를 하나 추가한다. M1 재측정 기록은 **자기 커밋**에 먼저 착지하고, 판별식은 `git merge-base --is-ancestor <M1 커밋> <M2 커밋>` 이 exit 0 을 낼 것. 한 커밋에 합쳐야만 하는 사정이 있다면 순서 조항을 커밋 그래프가 검증 가능한 문언으로 다시 쓴다.

**D6.** uncovered-and-self-voiding-req — `spec.md:L122-126` (REQ-WXR-007) + `acceptance.md` 전역 — **덮는 AC 가 없고, 완료 시점에 스스로 무효가 된다.** 측정: `acceptance.md` 에서 잔여-위험 의무를 검증하는 항목 적중 0건(`grep -n "잔여\|residual\|응답 문자열" acceptance.md` → AC-004 의 제목·본문만 적중, 의무 조항 아님). 구조적으로도 REQ-007 은 "**While** the re-measurement of AC-WXR-004 **has not been performed**" 를 가드로 갖는데, Definition of Done(`acceptance.md:L73`)이 AC-WXR-004 PASS 를 요구하므로 완료 시점에 가드는 거짓이 되고 REQ-007 은 아무 의무도 남기지 않는다. 더 문제는 실질이다 — AC-004 가 더하는 것은 **직접 판독 1회**이고, 관측 1~4 자체는 여전히 응답 문자열에 기대며 관측 4 는 재실행조차 불가하다. 즉 REQ-007 을 무효화하는 사건이 REQ-007 이 지목한 위험을 실제로는 해소하지 않는다. — Severity: **minor** — Class: **blocking** — Required fix: REQ-007 에 대응 AC 를 부여하거나, 재측정 이후에도 **재측정이 덮지 못하는 관측들에 한정해** 잔여-위험 문구가 남도록 문언을 고친다.

**D7.** ac003-discriminant-not-binary — `acceptance.md:L30-31` — **판별식이 이진이 아니고 REQ-WXR-004 와 긴장한다.** 문언은 "개정 후 그 절에 **무자격** 「originating checkout」 **단독** 문장이 남아 있지 않다(적중 0)" 이다. 「무자격」·「단독」은 판정자 재량이며, 문자열 grep 0건으로 읽으면 REQ-WXR-004(`spec.md:L108-111`)와 충돌한다 — REQ-004 는 기존 문장 "`ExitWorktree` returns to the originating checkout" 을 **reconcile** 하라고 요구하고, 조정은 통상 그 문장을 인용한다. 이행자 둘이 서로 다르게 구현할 여지가 있다. — Severity: **minor** — Class: **blocking** — Required fix: 두 개의 기계 검사로 분해한다 — (i) 인용 맥락 밖 무수식 사용 0건을 판정할 구체 문자열 조건, (ii) 정정 문장이 담아야 할 필수 문자열 1건 이상.

**D8.** classification-step-unearned — `spec.md:L45` (§1.2 제목) — **제목이 본문 근거보다 한 걸음 더 나간다.** 제목은 "런타임 결함이 **아니라** 계약(문서) 결함이다" 로 단정하는데, 제시된 근거는 "복귀는 일관되고 예측 가능하다"(`:L47`)뿐이다. **일관성은 상위 계약과의 합치를 입증하지 않는다** — 문서가 말하는 바를 일관되게 어기는 런타임도 일관적이다. 실제로 같은 절 `:L52-54` 가 상위 설명 (2) 가 관측 4 에서 "**거짓으로 읽힌다**" 고 적고 있다. 즉 SPEC 자신의 증거는 「상위 런타임이 자기 설명을 위반한다」는 읽기를 지지하며, 그 경우 고칠 표면이 런타임이라는 결론도 같은 관측에서 나온다. 그런데 `:L73` 은 이를 「모호성」으로 낮춰 부른다. 다행히 **범위 논거는 건전하고 그 자체로 충분하다** — `:L65-67` 의 [HARD] 와 §4(`:L142-147`)가 「우리 산출물이 아니라 요구사항을 걸 수 없다」를 명확히 한다. 분류 단정은 그 위에 얹힌 불필요하고 미입증인 한 걸음이다. — Severity: **minor** — Class: **blocking** — Required fix: 제목을 우리 표면으로 한정한다(예: 「이 카드가 고칠 수 있는 것은 계약(문서) 층이다」). 런타임 결함 읽기를 **명시적으로 이름 붙여 두고**, 기각이 아니라 **범위 밖**이라는 이유로 치워 둔다. `:L73` 의 「모호성」은 「관측 4 에서 거짓으로 읽힘」으로 바로잡는다.

**D9.** untracked-evidence-citation — `acceptance.md:L52` (+ `L75`) — **독트린이 인용하도록 지시된 증거 경로가 git 에 없다.** 측정: `git ls-files --error-unmatch .moai/reports/t965/observations.md` → `did not match any file(s) known to git`; `git check-ignore -v` → `.gitignore:227:.moai/reports/*`. 즉 이 저장소의 어떤 clone·CI 러너·다른 머신에도 그 파일은 존재하지 않는다. AC-WXR-005 는 개정 절이 바로 그 경로를 담을 것을 요구하고, REQ-WXR-008/AC-WXR-007 을 통해 **모든 사용자 프로젝트로 배포되는 템플릿 미러**에도 같은 경로가 실린다. 결과적으로 독트린의 유일한 귀속 앵커가 어느 독자에게도 해소되지 않는다 — `agent-common-protocol.md` § Parallel Execution 의 증거 반출 의무(「인용 대상은 추적되는 경로」)에 어긋나고, 인용 경로가 해소되지 않는 주장은 VCI §2 의 미귀속 주장이다. (참고: 이 보고서의 출력 경로 `.moai/reports/t965/plan-audit-verdict.md` 도 같은 규칙에 걸리나, 배차가 명시 지정한 경로이므로 그대로 따르고 사실만 기록해 둔다.) 부기 — 「내부 report 경로」 클래스는 leak 가드에 **기계적으로 등재되어 있지 않다**(측정: 패턴은 `~/\.claude/projects/-Users-|\.moai/backups/agent-archive-` 뿐). 따라서 이 항목은 CI 가 잡아 주지 않는 **독트린 위반**이며, 기계가 잡는 D1 과는 성격이 다르다. — Severity: **major** — Class: **blocking** — Required fix: 관측 기록과 M1 재측정 기록을 추적되는 경로로 **먼저 반출**한 뒤 그 경로를 인용한다. 템플릿 미러에는 저장소-내부 report 경로를 싣지 않는다(D1 의 수정과 같은 방향).

### Optional findings (오케스트레이터 재량 — 이 목록만으로 FAIL 을 만들지 않았다)

**O1.** probe-worktree-residue — `observations.md:L123-124` — 조사가 「정리 대상」으로 지목한 실험 잔재가 남아 있다. 측정: `git worktree list` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t965-probe-b  fe18be5e2 [worktree-t965-probe-b]`. `spec.md` §1.4 는 이 항목을 운반하지 않고 어떤 AC·DoD 도 닫지 않는다. 브랜치명도 `WT-` 접두 규약(`kanban-dispatch.md` § Isolation)을 따르지 않는다. — Severity: minor — Class: optional.

**O2.** plan-commit-subject-convention — 커밋 `d20a9f022` — 스키마(`spec-frontmatter-schema.md:L98`)가 plan 단계 커밋 제목으로 `feat(SPEC-{ID}): plan-phase artifacts ({tier}, {N} artifacts)` 를 규정하나 실제 제목은 `docs(spec): add SPEC-WORKTREE-EXIT-RETURN-001, …` 이다. 문서 산출물 성격상 `docs:` 선택에 합리가 있어 재량 항목으로 둔다. — Severity: minor — Class: optional.

---

## 무엇을 검사했고 무엇이 통과했나 (FAIL 을 만들 수 있었으나 만들지 않은 것)

FAIL 판정이 단순 흠집내기가 아님을 보이기 위해, **결함을 찾으러 갔다가 실제로 건전함을 확인한** 항목을 명시한다.

1. **AC-WXR-004 는 공허하지 않다 — 배차가 의심한 지점에서 SPEC 이 옳다.** `acceptance.md:L36-46` 은 (i) 「Exit 직후 작업 디렉터리를 **직접 판독**」과 (ii) 「같은 회차에 **런타임 버전을 기록**」을 **둘 다** 요구하고, (iii) 결정적으로 `L45-46` 에서 "**이 AC 는 판독값이 A 일 것을 요구하지 않는다** — 판독과 버전 기록이 존재할 것을 요구한다" 라고 못 박는다. 특정 값을 요구했다면 측정이 고무도장이 되었을 텐데 그러지 않았고, `L44-45` 는 어긋남이 나오면 「어긋남 자체를 기록」하며 두 회차를 합쳐 「고정」이라 쓰지 않을 것까지 규정한다. 이 항목은 **감점 없이 통과**했다. (잔여 비대칭만 D4·D6 에 기록: 재측정은 연쇄-깊이 축만 닫고 출발-디렉터리 축은 닫지 못한다.)
2. **상위 도구 표면으로의 요구사항 밀수 없음.** REQ-001~008 을 전수 확인한 결과 모든 규범 주어가 `the doctrine` 이다. REQ-WXR-005(`spec.md:L113-116`)는 상위 설명을 **조건절에서만** 언급하고 주어는 독트린이며, REQ-WXR-003 은 상위 응답 문자열을 **우리 문서가 서술할 것**을 요구할 뿐 상위의 변경을 요구하지 않는다. `spec.md:L65-67` 의 [HARD] 와 §4 배제가 이를 뒷받침한다. **밀수 0건 — 이 축은 FAIL 사유가 될 수 있었으나 되지 않았다.**
3. **독트린 표적이 실재한다 — 배차문을 믿지 않고 직접 읽었다.** `.claude/rules/moai/workflow/worktree-integration.md:220` 이 무수식 문장 "`ExitWorktree` returns to the originating checkout" 을 담고 있음을 확인했고, 절 제목 `### EnterWorktree / ExitWorktree Tools` 는 `:216` 에 실재한다. 템플릿 미러도 **같은 220행**에 같은 문장을 담는다(`grep -n "originating checkout"` 양쪽 적중). 표적이 허구였다면 SPEC 전체가 무너졌을 것이다.
4. **REQ-WXR-006 의 미재현 처리는 정직하다.** `spec.md:L118-120` 은 재현 주장과 원인 주장을 **둘 다** 금지하고, `L81-83`·§4(`L151-152`)·`plan.md:L57`·`plan.md:L73-74` 가 같은 금지를 네 곳에서 반복 고정한다. `acceptance.md:L56-62` 가 이를 검증한다. **원 제보를 해명된 것처럼 다루는 요구사항은 REQ-006 계열에는 없다** — 다만 REQ-001 의 독립성 주장이 간접적으로 그 제보를 반증하면서 침묵한다는 별개 문제가 있어 D4 에 기록했다.
5. **Tier S 예산 준수.** REQ 8개(상한 8), AC 7개(상한 8) — `spec-workflow.md:146` 기준 초과 없음. Tier S 가 2파일 규정임에도 `acceptance.md` 를 둔 편차는 `spec.md:L133` 과 `progress.md:L5` 에 카드 지시로 근거가 명시돼 있고, `acceptance.md` 가 frontmatter 를 갖지 않아 statelessness 규칙에도 저촉되지 않는다. 위반으로 보지 않았다.
6. **Out of Scope 절이 형식·실질 모두 유효.** `### Out of Scope — <주제>` H3 3개(`L142`, `L149`, `L158`) 각각이 구체적 `-` 불릿을 갖는다. 「없음」식 공허 항목 0건.

---

## 절차 결함 (SPEC 결함 아님 — 리드 보고 대상)

**감사 창 안에서 HEAD 가 이동했다.** 감사 시작 시 `d20a9f022`, 종료 시 `ce2e35355`(`docs: mark the three memory pointers as naming the dormant store (t965)`, `CLAUDE.local.md` 3행 변경). `agent-common-protocol.md` § Background Agent Execution 의 [HARD] — "감사 중인 워크트리는 쓰기 주체가 정확히 하나이며, 예기치 않은 HEAD 이동이나 외래 커밋 관측은 절차 결함이므로 조용히 계속하지 말고 즉시 리드에게 보고하고 진행 기록에 남긴다" — 에 따라 보고한다.

**판정에 미친 영향: 없음.** 위 § 측정 환경의 diff 측정이 감사 대상 경로 전부가 두 커밋 사이에서 불변임을 보인다. 그러나 「영향이 없었다」는 사후 측정의 결과이지 규율이 지켜졌다는 뜻이 아니다.

---

## Recommendation

**FAIL — 재작업 후 재감사.** must-pass 7개는 모두 통과했으므로 이 SPEC 은 구조적으로 건전하며, 결함은 전부 **문언 층에서 고칠 수 있다**. 코드 변경은 필요 없다. 아래 순서는 되돌리기 비용 순이다.

1. **D1 먼저 (critical).** AC-WXR-007 의 판별식에서 AC-WXR-005 의 날짜 요구를 떼어 내고, 귀속(날짜·기록 경로)을 로컬 사본 전용으로 한정한다. 미러는 날짜·내부 경로 없는 중립 서술만 담는다. 이 결정이 D9 의 절반과 AC-005 의 형태를 동시에 정하므로 가장 먼저 확정해야 한다.
2. **D3 → D2 순서로 (major).** 먼저 `spec.md:L86` 을 근거 파일의 실제 문언으로 바로잡는다(버전은 사후 기록되어 존재하되 약화됨, 관측 4 는 미공유). 그러면 AC-WXR-005(c) 를 회차별 귀속으로 분리할 재료가 생긴다 — 관측 회차에는 `2.1.278`, 재측정 회차에는 AC-004 의 값. 순서를 뒤집으면 D2 를 고칠 근거가 SPEC 안에 없다.
3. **D4 (major).** REQ-WXR-001 에서 `independent of` 전칭 술어를 제거하고 표본(깊이 1·2, 출발-디렉터리 단일 사례)을 문장에 싣는다. 같은 수정에서, 이 주장이 REQ-WXR-006 이 열어 둔 agent-5 제보와 어떻게 공존하는지 한 문장으로 명시한다.
4. **D5 (major).** M1 기록의 커밋이 M2 커밋의 조상임을 `git merge-base --is-ancestor` 로 판정하는 AC 를 추가한다. `plan.md:L82-83` 의 교차참조에 VCI **§2.3** 을 더한다.
5. **D9 잔여 (major).** 관측 기록·재측정 기록을 추적되는 경로로 반출한 뒤 그 경로를 인용하도록 AC-WXR-005 를 고친다.
6. **D6·D7·D8 (minor, blocking).** REQ-007 에 AC 를 주거나 재측정 이후에도 살아남을 문언으로 고친다. AC-003 판별식을 두 기계 검사로 분해한다. §1.2 제목을 우리 표면으로 한정하고 `L73` 의 「모호성」을 「관측 4 에서 거짓으로 읽힘」으로 바로잡는다.
7. **횡단 개선 (Traceability 0.60 → 0.75+).** 7개 AC 각각에 대응 `REQ-WXR-XXX` 를 명시 인용한다. 현재 `acceptance.md` 의 `REQ-WXR` 적중은 0건이며, 이 한 번의 편집이 D6 의 발견 비용도 없앤다.

O1·O2 는 재량 항목이다. **이 둘을 근거로 FAIL 을 만들지 않았으며**, 이행을 강제하지 않는다.

재감사는 위 열거된 결함 델타로 범위를 한정해 수행하면 된다. 단 Tier S 의 반복 상한은 1회이므로(`harness.plan_audit_tier_ceilings`), 재감사 진입 여부는 오케스트레이터가 상한 규약에 따라 판단한다.

---

*Audited by plan-auditor. M1 Context Isolation 적용. 모든 PASS 판정은 위에 인용한 명령과 그 관측 출력에 귀속된다.*
