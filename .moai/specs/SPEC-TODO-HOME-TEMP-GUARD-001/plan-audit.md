# SPEC Review Report: SPEC-TODO-HOME-TEMP-GUARD-001

Iteration: 1/1 (Tier S 상한 — `harness.plan_audit_tier_ceilings` S=1)
Verdict: **FAIL**
Overall Score: **0.75** (Tier S 임계값 0.75 — 임계값에 정확히 걸림, 여유 0)

측정 기준 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536` · `WT-home-fallback` @ `412c8cb14`.
아래 모든 인용·측정은 이 트리에서 이 감사 실행 중 직접 관측한 것이다.

M1 Context Isolation: 저자의 추론 맥락은 전달받지 않았고, 카드 지시문이 제공한 "확립된 사실"은
**그대로 채택하지 않고 이 트리에서 재측정**했다. 재측정 결과는 §측정 기록에 있다.

**FAIL 근거는 점수가 아니라 blocking 결함 목록이다.** must-pass 7건은 전부 통과했고 총점은
Tier S 임계값을 정확히 충족한다. 그러나 아래 D1-D4는 릴리스 게이트 산출물의 사실 오류·내부
모순이며, M6 분류상 blocking이다 — 수정 후 판정을 다시 받는 것이 옳다. 결함의 성격은 전부
**국소적이고 값싼 수정**이며, SPEC의 설계 골자(판별식·범위 경계·fail-open 방향)는 견고하다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-THG-001`..`008` 연속, 결번 0, 중복 0, zero-padding 일관.
  측정: `grep -o 'REQ-THG-[0-9]*' spec.md | sort -u` → 8행 001..008.
- **[PASS] MP-2 GEARS 형식 준수 (요구사항 계층에 대해 판정)** — `spec.md:102-109`의 REQ 8건 전수가
  GEARS 5패턴 중 하나에 해당한다. 판정 계층 명시: **`spec.md` §6의 `REQ-XXX` 항목에 대해서만**
  판정했다. `acceptance.md`의 Given-When-Then 항목은 검증 계층(AC)이므로 이 기준으로 감점하지
  않았다(Group 4에서 별도 채점).
  - Event-driven: 001(`When ... and ..., the resolver shall not ...`), 006(`When the guard refuses ..., the command shall surface ...`)
  - Ubiquitous: 002, 003
  - Unwanted: 005, 007, 008 (`shall not ...`)
  - 004: 본문은 `When normalization fails, ..., the discriminant shall report ...` — 유효한 Event-driven.
    다만 **라벨이 `(Event-detected)`** 로 적혀 있고 이는 GEARS 5패턴 이름이 아니다 → D5(minor).
    문장 자체는 준수하므로 MP-2는 PASS.
- **[PASS] MP-3 YAML frontmatter 유효성** — 정본 12필드 전수 존재, 타입 적합, 거부 별칭
  (`created_at`/`updated_at`/`labels`/`spec_id`) 사용 0.
  `spec.md:2-13` — id·title·version("0.1.0" 인용부호 유지)·status(draft)·created/updated(ISO
  `2026-09-08`)·author·priority(P2)·phase·module·lifecycle(spec-anchored)·tags(쉼표 구분 문자열).
  선택 필드 `tier: S`·`era`·`related_specs`는 추가 필드로 허용 범위.
- **[N/A] MP-4 언어 중립성** — 단일 언어(Go) 프로젝트 내부 SPEC이며 템플릿 바인딩 콘텐츠가 아니다.
  16개 프로그래밍 언어 열거 의무가 발생하지 않는다 → 자동 통과.
- **[PASS] MP-5 D7 cross-SPEC 정합** — 본문 참조 SPEC 2건 전수 실재 + 상태 확인.
  `SPEC-STATE-ANCHOR-001` → `status: completed`, `SPEC-WEB-TODO-QUEUE-001` → `status: completed`.
  `retired`/`superseded`/`archived` 0건 → BLOCKING 소견 없음.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c 'syscall' spec.md` → **0**. 언급 없음 → 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/`
  → rc=1, 매치 0. 미해결 마커 없음. (`research.md`는 Tier S로 부재 — `plan.md`가 존재하므로 N/A가
  아니라 실측 PASS다.)

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 (하단) | REQ 8건이 코드 좌표에 고정돼 있고 판별식 규칙이 `plan.md:45` D2로 못박혀 있다. 감점: `REQ-THG-001`(`spec.md:102`)이 순수 부정형이라 가드 발화 시 `ResolveTodoQueueRoot`의 **반환값이 미정의**(D1). 구현자가 `base`/`resolveStateDir(base)`/project-local 중 무엇을 고르느냐에 따라 콘솔 동작이 갈린다. |
| Completeness | 0.75 | 0.75 | 전 섹션 존재(HISTORY `spec.md:21`, WHY §1, WHAT §3/§6, REQUIREMENTS §6, AC `acceptance.md`, Out of Scope §7에 `### Out of Scope — <topic>` H3 5개 + 각 `-` 불릿). §8 Gaps가 이례적으로 정직하다. 감점: §8이 D1(대체 루트)과 D6(루트 집합 `/var/tmp` 누락 근거)을 기록하지 않았다. |
| Testability | 0.75 | 0.75 | AC 8건 전수 Given-When-Then + 판정 명령 + RED/GREEN 분리, `blocking` / `blocking (RED 관측 전)` / `invariant-guard` 3등급 어휘가 채택 판정 자격을 구분한다(`acceptance.md:5`) — 부재 가드를 RED-now로 채택하지 않는 규율이 명시적이다. weasel word 0건. 감점: `acceptance.md:37`의 RED 서술이 **현재 트리에서 재현되지 않는다**(D2), AC-THG-005가 자기 Then절과 GREEN 주석이 어긋난다(D3). |
| Traceability | 0.75 | 0.75 | REQ→AC 전수 피복(001→AC-001/007, 002→AC-002/006, 003→AC-002, 004→AC-004, 005→AC-003, 006→AC-005, 007→AC-005/008, 008→AC-008). 고아 AC 0, 미피복 REQ 0. 감점: `plan.md:88` M2 종료 조건이 **AC 번호를 오기**했다(D4). |

**Aggregate: (0.75 + 0.75 + 0.75 + 0.75) / 4 = 0.75**

---

## Defects Found (structured defect-list)

**D1. 가드 발화 시 순수 리졸버의 반환값이 어느 요구사항에도 정의돼 있지 않다**
— `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/spec.md:102` (REQ-THG-001)
— `REQ-THG-001`은 "홈 큐 루트로 **해석하지 않고 만들지도 않는다**"는 순수 부정형이다. 그런데
`ResolveTodoQueueRoot`의 시그니처는 `func ResolveTodoQueueRoot(base string) string`
(`internal/kanban/todo_root.go:64` — 직접 판독)로 **오류도 옵셔널도 아닌 `string` 하나를 반드시
돌려준다**. 즉 가드가 발화하면 무언가는 반환돼야 하는데, 그 무언가를 고정하는 요구사항이 없다.
`spec.md §8`의 미검증 항목은 U1(CLI 종료 형태)만 기록하고 **순수 경로의 대체 루트는 기록조차
하지 않았다** — 미결인 줄 모르는 미결이다. `AC-THG-005`(`acceptance.md:84`)는 "대신 해석된
루트"를 이름 부르라고 요구하여 그런 루트가 **이미 정의된 것처럼 전제**한다.
이것은 U1(종료 코드)과 **별개의 축**이다: 웹 콘솔은 종료 코드가 없으므로 U1을 어느 쪽으로
확정하든 순수 경로는 반드시 루트를 반환한다. 구현자가 `base`(기존 `fallbackTodoQueueRoot`의
`:148` `return base` 형태), `resolveStateDir(base, false)`(`:121` 형태), project-local 중
무엇을 고르느냐에 따라 콘솔이 렌더하는 큐가 달라진다 — 요구사항이 지배하지 않는 사용자 가시 동작이다.
— Severity: **major** — Class: **blocking**
— Required fix: REQ를 하나 추가하여(예: `REQ-THG-009`, Event-driven) 가드 발화 시
`ResolveTodoQueueRoot`가 해석하는 대체 루트를 고정한다. 그리고 `AC-THG-005`의 "대신 해석된 루트"가
그 REQ를 참조하도록 한다. 대안으로 U1과 묶어 `plan.md §D`의 미결 결정으로 **명시 등재**하고
`spec.md §8`에 Gap으로 기록하되, 그 경우 `AC-THG-005`의 Then절에서 "대신 해석된 루트를 이름
부른다"를 제거해야 한다(D3과 동시 해결).
— Tier 영향: REQ 8건 = Tier S 상한(`spec-workflow.md:140` 표: S=8/8)이므로 REQ 추가 시
**상한 초과 → Tier 재판정**이 발생한다. 등재 경로(REQ 추가 없음)를 택하면 Tier S가 유지된다.

**D2. `AC-THG-001`의 RED 서술이 현재 트리에서 재현되지 않는다 (사실 오류)**
— `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/acceptance.md:37`
— 인용: "adopting 경로는 `MkdirAll`까지 간다(`:175`)". **이 주장은 AC의 Given 조건 아래에서
거짓이다.** `adoptLocalTodoQueue`는 `:172-174`에서
`if _, err := os.Stat(local); err != nil { return }`로 **조기 반환**한다(직접 판독). 즉 base에
project-local `backlog.json`이 **존재할 때만** `:175`의 `MkdirAll`에 도달한다. 그런데
`AC-THG-001`의 Given(`acceptance.md:30`)은 "git 저장소가 아닌 임시 디렉터(`t.TempDir()`)"만
규정하고 **local backlog 파일을 세우지 않는다**.
귀결이 둘이다. ① `acceptance.md:37`의 RED 근거 문장이 사실이 아니다. ② `AC-THG-001`의 And절
(`:33` "canaryHOME/.moai/todo 아래에 어떤 디렉터도 생성되지 않는다")은 **가드 없는 현재
트리에서도 이미 참**이다 — 즉 그 절반은 공허하게 통과한다. SPEC은 이 위험 자체를 §5 mutant 4로
정확히 지목하고 `AC-THG-007` 뮤턴트로 방어하고 있어 설계는 옳으나, **뮤턴트도 같은 픽스처를 쓰면
오염을 재현하지 못한다**(판별식을 `false`로 무력화해도 local backlog가 없으면 `MkdirAll`에 도달
하지 않으므로 canary HOME 아래에 디렉터가 생기지 않는다). 즉 D2는 `AC-THG-007`의 판별 증거까지
공허하게 만든다.
— Severity: **major** — Class: **blocking**
— Required fix: `AC-THG-001`의 Given에 "base에 project-local `backlog.json`이 존재하는" 픽스처를
**명시 추가**한다(그래야 `:175` `MkdirAll` 경로가 실제로 도달 가능해진다). Given을 두 갈래로
분리하는 편이 더 정확하다 — (a) local queue 없음 → 반환 루트만 판정, (b) local queue 있음 →
반환 루트 + canary HOME 디렉터 0 판정. `AC-THG-007`의 뮤턴트 절차(`acceptance.md:108-113`)도
(b) 픽스처를 쓰도록 명시해야 오염 재현이 성립한다. `acceptance.md:37`의 문장은 이 조건을
명시하도록 정정한다.

**D3. `AC-THG-005`가 자기 Then절과 GREEN 주석이 어긋나고, U1 중립성 주장이 과장됐다**
— `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/acceptance.md:84` vs `:91`
— 카드가 명시적으로 검증을 요구한 항목이다. **판정: 완전한 outcome-neutral이 아니다 — 한쪽 답을
밀수한다.**
Then절(`:84`)은 세 가지를 요구한다: ⑴ 매치된 임시 루트를 이름 부름, ⑵ **대신 해석된 루트를
이름 부름**, ⑶ canary HOME 디렉터 0. 그런데 GREEN 주석(`:91`)은 "판정 대상은 「홈 디렉터 0 +
안내 관측」이다"라며 **두 가지로 축소**한다. 같은 AC 안에서 판정 대상이 3항 → 2항으로 바뀐다.
U1(a)(비영 종료 + 안내)를 택하면 명령이 큐를 쓰지 않고 죽으므로 "대신 **해석된** 루트"의
지시대상이 존재하지 않는다 — ⑵는 U1(b)(안내 후 임시-로컬 큐로 계속)를 전제한다. 즉 Then절은
U1(b) 편향이고, 중립성은 GREEN 주석이 Then절을 사후 축소함으로써만 성립한다.
같은 전제가 요구사항에도 박혀 있다: `REQ-THG-006`(`spec.md:107`) "naming the matched temp root
**and the root the queue resolved to instead**".
따라서 `spec.md:136`의 "AC-THG-005는 두 형태 중 어느 쪽에서도 판정 가능하도록 쓰였다"는 주장은
**AC가 실제로 쓰인 형태에 대해 참이 아니다** — 축소된 판정 대상에 대해서만 참이다.
— Severity: **major** — Class: **blocking**
— Required fix: 둘 중 하나. ⒜ Then절 ⑵를 "가드가 발화한 이유와, 이 실행에서 큐가 어떻게
처리되는지를 안내가 명시한다"처럼 U1 양쪽에서 지시대상을 갖는 표현으로 고쳐 쓰고 `REQ-THG-006`도
같이 고친다. ⒝ U1을 M2 이전에 운영자 확인으로 **먼저 확정**하고(그편이 `plan.md:58`의 권고와도
일치한다) AC를 그 형태로 확정한다. 어느 쪽이든 Then절과 GREEN 주석의 판정 대상을 **같게** 만든다.

**D4. `plan.md` M2 종료 조건이 AC 번호를 오기했다**
— `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/plan.md:88`
— 인용: "**AC-THG-007**(순수성·키 유도 불변) GREEN". 그러나 `acceptance.md:21-22`에서
`AC-THG-007`은 **뮤턴트 증명**(REQ-THG-001)이고, 순수성·키 유도 불변은 **`AC-THG-008`**
(REQ-THG-007, 008)이다. 괄호 안 설명은 008을 가리키는데 번호는 007이다.
동시에 `plan.md:92`(M3)는 `AC-THG-007`을 뮤턴트로 **올바르게** 인용한다. 즉 M2 종료 조건만
어긋나 있고, 그대로 실행하면 M2가 M3 소관인 뮤턴트 관측을 게이트로 요구하거나(순서 구속 §F 위반)
반대로 불변 방어가 M2에서 판정되지 않고 새어 나간다.
— Severity: **major** — Class: **blocking**
— Required fix: `plan.md:88`을 "AC-THG-**008**(순수성·키 유도 불변) GREEN"으로 정정한다.

**D5. `REQ-THG-004`의 패턴 라벨 `(Event-detected)`가 GEARS 5패턴에 없다**
— `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/spec.md:105`
— GEARS 정본 패턴은 Ubiquitous / Event-driven / State-driven / Where / Unwanted 다섯이다.
`Event-detected`는 그중 어느 것도 아니다. 문장 본문(`When normalization fails, ..., the
discriminant shall report *not temporary*`)은 Event-driven으로 완전히 준수하므로 MP-2 판정에는
영향이 없고, 결함은 라벨 하나에 국한된다.
— Severity: **minor** — Class: **blocking** (문서가 스스로 선언한 분류 체계와의 내부 불일치이며
수정 비용이 한 단어다)
— Required fix: `(Event-detected)` → `(Event-driven)`.

**D6. 임시 루트 집합에서 `/var/tmp`가 빠졌고, 누락 근거가 기록돼 있지 않다**
— `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/spec.md:103` (REQ-THG-002)
— 루트 집합은 `os.TempDir()`, `/tmp`, `/var/folders` 셋으로 고정돼 있다. 이 기계에서 실측:
`/var/tmp`와 `/private/var/tmp`가 **둘 다 실재하는 디렉터**다(`drwxrwxrwt root wheel`). `/var/tmp`는
POSIX·리눅스 양쪽에서 관습적 임시 루트이며 `os.TempDir()`가 가리키지 않는다(이 기계
`TMPDIR=/var/folders/kt/.../T/`). 즉 `/var/tmp` 아래의 비git 실행은 가드를 통과해 홈 큐를 만든다.
fail-open 방향(`REQ-THG-004`)이므로 **회귀는 아니고** 오늘 동작이 유지될 뿐이며, 루트 추가는
한 행짜리 작업이다. 문제는 누락 자체보다 **`spec.md §8`이 이 결정을 Gap으로 기록하지 않은 것**이다
— 루트 집합의 완결성이 검토된 흔적이 없다.
— Severity: **minor** — Class: **optional**
— Required fix: 루트 집합에 `/var/tmp`를 추가하거나, 추가하지 않는 근거를 `spec.md §8`에 Gap으로
등재한다(예: "`/var/tmp`는 지속성 임시 디렉터라 소멸 전제가 성립하지 않아 제외").

**D7. `proj-325ca0b6`의 임시-디렉터 기원이 어느 SPEC에서도 측정으로 확립되지 않았다**
— `.moai/specs/SPEC-TODO-HOME-TEMP-GUARD-001/spec.md:42`
— `spec.md:42`는 생산 기원 2건(`proj-325ca0b6`, `t203-probe-d7a16ea2`)을 "비git 임시 디렉터에서
`moai todo`를 호출해 생긴 것들"이라고 **단정**한다. 선행 SPEC(`SPEC-STATE-ANCHOR-001` §5,
`:110`)도 같은 단정을 하며, 양쪽 모두 그 기원 경로를 복원한 측정을 인용하지 않는다.
나는 이 감사에서 `TodoQueueProjectKey`의 유도(`basename + "-" + sha256(abs)[:4]`,
`todo_root.go:197-203`)를 역산해 후보 15개 루트 × 2개 basename을 전수 대조했다. 결과:
**`t203-probe-d7a16ea2` ← `/tmp/t203-probe` 로 정확히 복원됐다**(§측정 기록 참조). 이로써
그 1건의 임시 기원은 **추정에서 측정으로 승격**된다. 그러나 `proj-325ca0b6`은 후보 15개 중
어느 것과도 일치하지 않아 **복원되지 않았다** — 임시 기원인지 아닌지 나는 판정하지 못한다.
결론(생산 축이 살아 있고 임시 기원이 실재한다)은 `t203-probe` 1건만으로도 성립하므로 SPEC의
논지는 무너지지 않는다. 결함은 **단정의 폭이 근거보다 넓다**는 것이다.
— Severity: **minor** — Class: **optional**
— Required fix: `spec.md:42`를 측정된 범위로 좁힌다 — 예: "`t203-probe-d7a16ea2`는 키 역산으로
`/tmp/t203-probe`가 확인됐다(미해석 `/tmp` 철자). `proj-325ca0b6`의 기원 경로는 복원되지 않았고
임시 기원 여부는 미검증이다." 그리고 후자를 `spec.md §8`에 Gap으로 옮긴다.

**D8. `internal/cli/launcher.go`의 행 범위 인용이 어긋난다**
— `spec.md:84`, `plan.md:45`(D2), `plan.md:114`(§H) — 세 곳 모두 `457-486`
— 실측: `resolveSymlinks`의 doc 주석은 **456**행에서 시작하고(`// resolveSymlinks returns the
symlink-resolved form of path.`), 함수 선언은 **475**행, 본문 종료는 **484**행이다. 486행은
함수 밖 공백이다. 즉 실제 범위는 `456-484`이고 인용은 시작에서 1행, 끝에서 2행 어긋난다.
인용된 **내용**(존재하면 `EvalSymlinks`, 없으면 어휘적 `Clean` + GOOS 발산 논거 + `@MX:REASON`)은
그 범위 안에 전부 실재하므로 논거 자체는 건재하다.
— Severity: **minor** — Class: **optional**
— Required fix: 세 인용을 `internal/cli/launcher.go:456-484`로 정정한다.

---

## 카드가 명시적으로 물은 항목에 대한 판정

### ① 범위 경계 — 비임시 비git base의 홈 폴백을 회수하는 요구사항/AC/마일스톤이 있는가

**없다. 위반 0건.** 전수 판독 결과 경계가 네 겹으로 방어돼 있다:

| 층 | 좌표 | 내용 |
|---|---|---|
| 요구사항 | `spec.md:106` REQ-THG-005 | "shall not withdraw the home fallback for a non-git project outside the temp-root set; 'any non-git base' is explicitly NOT the trigger" |
| 검증 | `acceptance.md:55-66` AC-THG-003 | 비임시 비git base가 `canaryHOME/.moai/todo/<key>`로 **종전대로** 해석됨을 단언. 등급이 `blocking` (RED 관측 전 아님) — 기각된 넓은 판독을 구현하면 FAIL |
| 계획 구속 | `plan.md:44` D1 | "이 경계를 넓히는 구현은 요구사항 위반이다" |
| 범위 밖 | `spec.md:121-123` | `### Out of Scope — 「모든 비git base」로의 확대` |

추가로 `plan.md:103` §G가 "비git이면 어차피 임시나 마찬가지"라는 단순화를 안티패턴으로 명시하고,
`spec.md:68`이 1차 전달의 확대를 **기각된 판독**으로 기록한다. 운영자 결정이 정확히 전사됐다.

### ② 판별식이 세 도달 형태에 실제로 발화하는가 — 그리고 거짓 양성 자세는 정직한가

**발화한다. 세 형태 전수 확인.** 판별식 규격(`spec.md:103-104`: 양쪽 정규화 + 구성요소 포함,
루트 집합 `os.TempDir()`/`/tmp`/`/var/folders`)을 실측 경로에 대입한 결과:

| 도달 형태 | 실측 경로 형태 | 정규화 후 | 매치 루트 | 발화 |
|---|---|---|---|---|
| ① 기존 테스트 `todo_root_test.go:98` | `t.TempDir()` → `$TMPDIR/...` | `/private/var/folders/kt/.../T/...` | `os.TempDir()`→`/private/var/folders/.../T` | ✅ |
| ② 생산 바이너리 `t203-probe-d7a16ea2` | **`/tmp/t203-probe`** (키 역산으로 복원) | `/private/tmp/t203-probe` | `/tmp`→`/private/tmp` | ✅ |
| ③ 테스트 기원 `001-*` 341개 | `$TMPDIR/TestX/001` | `/private/var/folders/...` | `os.TempDir()` | ✅ |

②의 복원은 판별식 설계를 **사후적으로 정당화한다**: 생산 오염이 `/private/tmp`가 아니라
**미해석 `/tmp` 철자**로 도달했다. 즉 루트 집합에 `/tmp`가 반드시 있어야 하고
(`os.TempDir()`만으로는 이 기계에서 `/var/folders/...`라 안 잡힌다), 양쪽 정규화가 없으면
`/tmp/t203-probe` vs `/private/tmp` 비교가 **조용히 어긋난다**. `spec.md §4`의 논거가 실측으로
확증됐다 — 저자가 이 역산을 하지 않고도 옳은 설계에 도달했다.

**거짓 양성 자세는 정직하다 — 손 흔들기가 아니다.** `spec.md:90`은 잔여 거짓 양성 모집단을
"임시 루트 아래에 사는 비git 진짜 프로젝트"로 좁히고, git 저장소는 임시 루트 아래에 있어도
`primaryCheckoutRoot`가 먼저 답해 폴백 가지에 **도달하지 않는다**고 주장하며 `:64-69`/`:76-83`을
인용한다. **직접 판독으로 검증했다**: `ResolveTodoQueueRoot`(`:64-69`)는 `primaryCheckoutRoot`가
`ok`면 즉시 반환하고 `fallbackTodoQueueRoot`에 도달하지 않는다. `ResolveTodoQueueRootAdopting`
(`:76-83`)도 동일. **인용 좌표와 주장이 모두 정확하다.** 나아가 `spec.md:138`이 그 모집단의
실제 빈도를 측정하지 않았음을 Gap으로 스스로 기록한다 — 수용을 근거 없이 통과시키지 않았다.

**과도한 관대함은 없다.** 다만 `/var/tmp` 누락(D6)이 유일한 커버리지 구멍이며, fail-open
방향이라 회귀가 아니다.

### ③ U1 처리의 적정성

**부적정 — D3.** `plan.md:56-58`에 미결로 등재한 것 자체는 옳고, "선택 전에 blocker 보고로
운영자 확인을 받는 편이 안전하다"는 권고도 적절하다. 그러나 "`AC-THG-005`가 두 형태 모두에서
판정 가능하다"는 주장(`spec.md:136`, `plan.md:58`, `acceptance.md:91`)은 AC가 실제로 쓰인 형태에
대해 참이 아니다. 상세와 수정 지시는 D3 참조.

### ④ Linux 미측정 처리의 적정성

**적정 — 저자의 두 주장 모두 사실이다.**
- `spec.md:135`가 Linux `/tmp`·`/private/tmp` 실측 부재를 **명시적 Gap**으로 기록하고,
  "설계 근거이지 측정이 아니다"라고 스스로 격하한다. **확인됨.**
- `acceptance.md:53`이 AC-THG-002의 리눅스 셀 각주로 "**로컬 macOS PASS를 리눅스 판정으로
  재사용하지 않는다**"고 금지하고, CI linux 매트릭스 결과를 `progress.md §E.2`에 인용하도록
  구속한다. **확인됨.**

이 처리는 baseline-integrity attribution 규율(다른 트리/시점의 측정을 이번 측정으로 쓰지 않기)을
AC 수준에서 강제한 것이고, 이 SPEC에서 가장 잘 쓰인 부분이다.

### ⑤ Tier S 유지 여부

**유지된다.** REQ 8 / AC 8은 Tier S 상한(`spec-workflow.md:146-150` 표: S = 8/8)에 **정확히
걸쳐 있고 초과하지 않는다**. 접촉 패키지 2개(`internal/kanban` 주 + `internal/cli` 안내 1행),
마일스톤 3개, 신규 아키텍처 없음. `plan.md:16`이 "정규화 규칙을 공유 위치로 추출해야 한다는
판단이 서면 Tier 재판정을 blocker로 보고한다"는 승격 트리거까지 미리 못박아 뒀다.

**단, 여유가 0이다.** D1을 REQ 추가로 해결하면 상한을 넘어 Tier 재판정이 발생한다. D1의 수정
경로로 "미결 등재"를 권하는 이유가 이것이다.

### ⑥ 정리(t542) / `stateanchor`(t537) 경계 침범 여부

**침범 없음.** 정리·삭제·이관을 정의한 요구사항은 0건이며 `REQ-THG-008`(`spec.md:109`)이 명시적
금지, `plan.md:50` D7이 구속, `acceptance.md:126`이 `~/.moai/todo` 계수 불변을 단언한다.
`spec.md:119`의 "`internal/kanban/todo_root.go`는 `internal/stateanchor`를 임포트하지 않는다"는
주장은 import 블록(`todo_root.go:29-37`) 직접 판독으로 **참**이다 — `crypto/sha256`, `fmt`,
`os`, `path/filepath`, `gitcore`, `paths` 뿐이다.

---

## 측정 기록 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

**Baseline-attribution**: 아래 전부 이 감사 실행 중, 이 트리
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536`, `WT-home-fallback` @ `412c8cb14`)에서
직접 실행한 명령의 관측 출력이다. 카드 지시문이 제공한 수치를 그대로 옮긴 것은 없다.

| # | Claim | Evidence (명령 → 관측) |
|---|---|---|
| 1 | 좌표 정확 | `cat -n internal/kanban/todo_root.go` → `:64-69` ResolveTodoQueueRoot, `:76-83` Adopting, `:96` primaryCheckoutRoot, `:112` homeTodoQueueRoot, `:175` MkdirAll, `:197` TodoQueueProjectKey, `:8-17` 순수/adopting 분할 doc, `:62-63` "exactly one queue" 주석 — **spec.md 인용 전수 일치** |
| 2 | 소비자 2곳 | `/usr/bin/grep -rn 'ResolveTodoQueueRoot' internal/ --include='*.go' \| grep -v _test.go` → 생산 히트 2건: `internal/cli/todo.go:72`(Adopting), `internal/web/todo_queue_read.go:33`(순수). 나머지는 전부 주석 |
| 3 | 기존 테스트 실재 | `grep -rn 'func TestResolveTodoQueueRoot_FallbackNoGit\|_PureFallbackWritesNothing' internal/kanban/` → `todo_root_test.go:98`, `:125` |
| 4 | **생산 기원 1건 역산 성공** | `TodoQueueProjectKey` 역산(python3, 후보 15루트×2 basename) → `MATCHES: [('t203-probe-d7a16ea2', '/tmp/t203-probe')]`. `proj-325ca0b6`은 **미복원** |
| 5 | 심링크 사실 | `readlink /var` → `private/var`; `readlink /tmp` → `private/tmp`; `TMPDIR=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/` — spec.md §4와 일치 |
| 6 | `/var/tmp` 실재 | `ls -d /var/tmp /private/var/tmp` → 둘 다 `drwxrwxrwt root wheel` (D6 근거) |
| 7 | 고아 계수 | `find ~/.moai/todo -maxdepth 1 -type d \| wc -l` → **344** — `plan.md:38` 기준값과 일치 |
| 8 | 참조 SPEC 상태 | `grep '^status:' .moai/specs/SPEC-{STATE-ANCHOR,WEB-TODO-QUEUE}-001/spec.md` → 둘 다 `completed` |
| 9 | 선행 §5 서술 | `SPEC-STATE-ANCHOR-001/spec.md:110` → "표본 2건(`proj-325ca0b6`, `t203-probe-d7a16ea2`)" — `spec.md §1.2`의 "생산 기원 한 쌍만 센 것" 서술이 **정확** |
| 10 | launcher 실제 행 | `grep -n 'resolveSymlinks returns\|^func resolveSymlinks'` → 456(doc 시작), 475(func) — 인용 `457-486`과 불일치 (D8) |
| 11 | adopt 조기 반환 | `sed -n '169,177p' todo_root.go` → `os.Stat(local)` 실패 시 `:174` return, `:175` MkdirAll 미도달 (D2 근거) |
| 12 | 마커 부재 | `grep -rn '\[NEEDS CLARIFICATION' <specdir>` → rc=1; `grep -c syscall spec.md` → 0 |

**Gaps — 이 감사에서 관측하지 **못한** 것 (명시)**
1. `proj-325ca0b6`의 기원 경로. 후보 15루트로는 복원되지 않았고, sha256 역상은 전수 탐색이
   불가능하다. **임시 기원인지 아닌지 나는 모른다** — 아니라고 판정한 것이 아니다.
2. Linux에서의 `/tmp`·`/var/folders` 정규화 거동. 이 기계는 darwin이며 linux 셀은 실행하지
   않았다. SPEC 자신이 Gap으로 기록한 항목과 동일하다.
3. 판별식의 실제 구현 거동. **코드가 존재하지 않는다**(사전 구현 감사) — 위 발화 판정은 SPEC이
   규정한 규격을 실측 경로에 대입한 **설계 수준 추론**이지 실행 관측이 아니다.
4. `go test` 실행. 사전 구현 감사이므로 테스트를 돌리지 않았다. `plan.md §C`의 baseline
   (`go test ./internal/kanban/... -count=1` ok)은 **재측정하지 않았다** — run-phase 진입 시
   `plan.md:30`의 지시대로 재측정 대상이다.
5. `~/.moai/todo` 아래 343개 디렉터의 내용 전수(`backlog.json` 211개, 카드 텍스트 분포). 계수
   1건(344)만 재측정했고 카드 텍스트 분포는 카드 지시문 값을 **채택하지 않고 미검증으로 남겼다**.

**Residual-risk — 위 관측에도 불구하고 여전히 틀릴 수 있는 것**
- 도달 형태 ②의 역산은 `TodoQueueProjectKey`가 **당시에도 지금과 같은 유도식**이었음을 전제한다.
  유도식이 그 사이 바뀌었다면 `/tmp/t203-probe` 복원은 우연의 일치일 수 있다(4바이트 = 32비트
  충돌 공간이므로 후보 30개 중 우연 일치 확률은 낮으나 0은 아니다).
- 판별식 발화 판정(②)은 구현이 SPEC 규격을 정확히 따를 때만 유효하다. run-phase가 규격에서
  이탈하면 이 판정은 무효가 되며, 그것을 잡는 것이 `AC-THG-007` 뮤턴트의 역할이다 — 단 D2가
  해결되지 않으면 그 뮤턴트도 공허해진다.
- D1의 미정의 반환값이 구현에서 어떻게 메워지든, 이 감사는 그 선택의 콘솔 측 파급을 측정하지
  않았다.

---

## Recommendation

**FAIL. 아래 순서로 수정한 뒤 재감사를 받는다.** 재감사는 이 결함 델타에만 범위를 한정한다
(Tier S 상한이 1회이므로, 재감사는 상한 밖의 확인 통과로 처리하거나 운영자 판단으로 연장한다).

1. **D4 정정 (1행, 즉시)** — `plan.md:88`의 `AC-THG-007` → `AC-THG-008`.
2. **D5 정정 (1단어, 즉시)** — `spec.md:105`의 `(Event-detected)` → `(Event-driven)`.
3. **D8 정정 (3곳, 즉시)** — `spec.md:84`, `plan.md:45`, `plan.md:114`의 `launcher.go:457-486`
   → `456-484`.
4. **D2 수정 (필수, 채택 판정의 유효성이 걸려 있다)** — `AC-THG-001`의 Given을 두 갈래
   ((a) local queue 없음 / (b) local queue 있음)로 분리하고, `acceptance.md:37`의 RED 서술에
   `MkdirAll` 도달 조건을 명시한다. `AC-THG-007`의 뮤턴트 절차가 (b) 픽스처를 쓰도록 못박는다.
   **이 수정 없이는 `AC-THG-007`의 뮤턴트가 판별 증거로 성립하지 않는다.**
5. **D3 수정 (필수)** — U1을 M2 이전에 운영자 확인으로 확정하거나(권장 — `plan.md:58`의 자체
   권고와도 일치), `AC-THG-005`의 Then절 ⑵와 `REQ-THG-006`을 U1 양쪽에서 지시대상을 갖는
   표현으로 고쳐 쓴다. Then절과 GREEN 주석의 판정 대상을 같게 만드는 것이 수용 기준이다.
6. **D1 수정 (필수)** — 가드 발화 시 `ResolveTodoQueueRoot`의 대체 루트를 고정한다.
   **Tier S 유지를 위해 REQ 추가보다 `plan.md §D`의 미결 결정 등재 + `spec.md §8` Gap 등재를
   권한다**(REQ 8건이 이미 Tier S 상한이다). 등재를 택하면 D3의 ⒜ 경로와 함께 처리해야 한다 —
   대체 루트가 미결인 채로 `AC-THG-005`가 그것을 이름 부르라고 요구할 수는 없다.
7. **D6·D7 (optional, 운영자 재량)** — 루트 집합의 `/var/tmp` 추가 여부와 `spec.md:42`의 단정
   범위 축소. 판정을 막지 않으므로 오케스트레이터가 라우팅 여부를 결정한다. D7은 위 측정 기록 #4의
   역산 결과를 그대로 인용하면 정정이 끝난다.

**감사자의 총평.** 이 SPEC은 결함 6-8건에도 불구하고 **설계의 골자가 옳다**. 범위 경계가 네 겹으로
방어돼 운영자 결정이 정확히 전사됐고(카드가 가장 걱정한 축), fail-open 방향의 비대칭 논거가
코드 좌표에 고정돼 있으며, 거짓 양성 자세가 손 흔들기가 아니라 `:64-69` 직접 판독에 근거하고,
부재 가드를 RED-now로 채택하지 않는 규율(`acceptance.md:5`, `plan.md:52` D9)과 로컬 macOS PASS를
리눅스 판정으로 재사용하지 않는 금지(`acceptance.md:53`)가 명시돼 있다. 나아가 판별식 설계는
저자가 수행하지 않은 키 역산에 의해 **사후적으로 확증됐다** — 생산 오염이 실제로 미해석 `/tmp`
철자로 도달했고, 이는 SPEC이 §4에서 예측한 바로 그 형태다.

FAIL은 이 설계에 대한 판정이 아니라 **산출물 4곳의 사실 오류·내부 모순과 1곳의 사양 공백**에
대한 판정이다. D1-D5는 전부 국소 수정이며, 그중 D2·D3은 방치하면 run-phase가 재현되지 않는 RED를
기록하거나 축소된 판정 대상으로 AC를 통과시키게 된다 — 즉 **공허한 초록**을 만든다. 이 SPEC이
스스로 §5에서 지목한 바로 그 실패 형태다.

---

_감사자: plan-auditor · 이 트리에서 직접 측정 · `412c8cb14`_

---
---

# SPEC Review Report — 2차 (수리 후 재감사): SPEC-TODO-HOME-TEMP-GUARD-001

> **1차 기록은 위에 그대로 있다.** 이 절은 덮어쓰기가 아니라 **추가**이며, 1차에서 무엇을
> 찾았는지의 기록은 보존된다. 1차 판정(FAIL / 0.75)은 그 시점 산출물(`version: 0.1.0`)에
> 대한 것이고, 아래는 `version: 0.1.1`에 대한 별개 판정이다.

Iteration: 2 (Tier S 상한은 1 — 이 2차는 **결함 델타 확인 감사**이며 신규 전수 감사가 아니다.
판정 권한은 여전히 이 에이전트에 있다.)
Verdict: **FAIL**
Overall Score: **0.875** (Tier S 임계값 0.75 — 임계값은 넘겼으나 blocking 결함이 남았다)

측정 기준 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536` · `WT-home-fallback`.
아래 모든 인용·측정은 이 감사 실행 중 이 트리에서 직접 관측했다.

M1 Context Isolation: 카드 지시문이 제공한 **수리 주장은 하나도 그대로 채택하지 않았다** —
각 항목을 파일과 코드에서 재측정했다. 결과는 §9에 있다.

**점수는 올랐고(0.75 → 0.875) 결함 6건이 실제로 닫혔다. 그럼에도 FAIL인 이유는 하나다:**
D1 수리가 공백을 메우면서 **틀린 값으로 메웠다**. 대체 루트로 고정된 `resolveStateDir(base, false)`는
이 코드베이스의 「루트」 계층보다 **한 단계 아래**이며, 그대로 구현하면 SPEC 자신이 `""`를 금지한
근거(`plan.md` D12: "거부가 아니라 **조용한 오작동**")와 **같은 실패**를 낸다. 새 결함 D9다.
나머지는 전부 수리됐고, **D5는 내 오류였다 — 철회한다.**

---

## 1. 1차 결함의 처분 (전수)

| # | 처분 | 근거 (이 감사에서 직접 관측) |
|---|---|---|
| D1 | **부분 수리 → 새 결함 D9로 승계** | 공백 자체는 닫혔다: `spec.md:114` REQ-THG-001이 대체 루트를 명시하고, `plan.md:55` D12가 `""` 금지를 근거와 함께 등재했으며, `spec.md:149`가 U1과의 축 분리를 기록했다. **그러나 고정된 값이 틀렸다** — §2 |
| D2 | **수리 (확인됨)** | `acceptance.md:34-43`이 Given을 (a)/(b)로 분리했고, `:47`의 RED 서술이 `MkdirAll` 도달을 **조건부**로 정정했으며, `:49`가 종전 문장의 거짓을 정정 기록으로 남겼다. `AC-THG-007`(`:122-126`)이 (b) 픽스처를 뮤턴트 재현 담당으로 명시했다. **판별력 검증은 §3** |
| D3 | **수리 (확인됨)** | `spec.md:119` REQ-THG-006과 `acceptance.md:96` Then절 ⑵가 「계속하면 대체 루트 이름 / 멈추면 큐 미사용 명시」의 **선언**으로 다시 쓰였다 — U1 양쪽에 지시대상이 있다. `acceptance.md:103` GREEN 주석이 **"판정 대상은 ⑴⑵⑶ 세 항목 전부이며, GREEN이라고 해서 축소되지 않는다"**로 바뀌어 3항→2항 축소가 제거됐다. 카드가 특정해 물은 항목이며, **축소는 사라졌다** |
| D4 | **수리 (확인됨)** | `plan.md:92` = "AC-THG-**008**(순수성·키 유도 불변) GREEN". `:94`에 정정 사유가 기록됐고, M3(`:98`)의 007 인용은 뮤턴트를 가리켜 여전히 옳다 |
| D5 | **철회 — 내 판정이 뒤집혔다** | §4 |
| D6 | **수리 (등재로 해결)** | `spec.md:152`가 제외 근거 + 재검토 조건을 Gap으로 등재했고 `plan.md:56` D13이 구속한다. 1차 요구는 「추가하거나 근거를 등재하라」였고 후자가 이행됐다. 근거 강도는 §5 D11 |
| D7 | **수리 (확인됨)** | `spec.md:43`이 `t203-probe-d7a16ea2`만 측정으로 주장하고 `proj-325ca0b6`을 **미검증**으로 명시했으며, `:153`이 Gap으로 등재했다. 종전 단정이 `SPEC-STATE-ANCHOR-001`에서 근거 없이 옮겨온 것이라는 사실까지 기록했다. **키 값은 내가 독립 재유도해 일치를 확인했다**(§9 #3) |
| D8 | **수리 (확인됨) — 내 1차 값도 1행 틀렸다** | 실측: 456 = doc 시작, 475 = `func`, 484 = `return filepath.Clean(path)`, **485 = 닫는 중괄호**, 486 = 함수 밖. 정본 범위는 `456-485`이고 세 인용(`spec.md:85`, `plan.md:45`, `plan.md:120`)이 전부 그 값으로 정정됐다. 1차가 제시한 `456-484`는 **끝에서 1행 짧았다** — 수리 쪽이 옳다 |

---

## 2. 신규 blocking 결함

### D9. REQ-THG-001이 고정한 대체 루트가 「루트」 계층이 아니다 — 조용한 오작동을 사양으로 못박았다

— `spec.md:114` (REQ-THG-001) · `plan.md:55` (D12) · `acceptance.md:36`, `:42` (AC-THG-001 Then)
— Severity: **major** — Class: **blocking**

**이 코드베이스에서 「루트」는 그 아래에 `.moai/state/<name>/`이 걸리는 경로다.** 직접 판독:

- `BacklogPathForRoot(root)`(`internal/kanban/state_dir.go:129-132`) = `resolveStateDir(root,false)` +
  `backlog.json`, 즉 **`root/.moai/state/{todo|kanban}/backlog.json`**. 함수 doc(`:126`)이 스스로
  "under a project root"라고 말한다.
- 온디스크 실측: `~/.moai/todo/001-a49ab4ee/.moai/state/kanban/backlog.json` — 홈 루트
  (`home/.moai/todo/<key>`) 아래에 `.moai/state/...`가 한 번 더 걸린다. 루트 계층이 실물로 확인된다.
- 소비자 전수: 콘솔(`internal/web/todo_queue_read.go:33-35`)은 반환값을 `vm.Root`에 싣고
  `BacklogPathForRoot(root)`로 읽는다. CLI(`internal/cli/todo.go:56-57`, `:71-79`)는
  `BacklogPathForRootAdopting(root)`로 읽는다. **두 소비자 모두 반환값에 `.moai/state/...`를 덧붙인다.**

그런데 REQ-THG-001이 고정한 값은 `resolveStateDir(base, adopt=false)`(`todo_root.go:121`) —
`StateDirForRoot(base)`(`state_dir.go:46-48`) = **`base/.moai/state/todo`**, 즉 **상태 디렉터
그 자체**다. 이것을 루트로 돌려주면 소비자가 계산하는 큐 경로는

```
base/.moai/state/todo/.moai/state/todo/backlog.json
```

이 된다 — **아무도 쓰지 않고 아무도 읽지 않는 경로**다. 그리고 결정적으로, 운영자의 실제 로컬 큐는
`base/.moai/state/todo/backlog.json`에 있다(`seedLocalQueue`, `todo_root_test.go:65-67`이 정확히
그 경로에 쓴다 — **AC-THG-001 (b)가 세우는 바로 그 픽스처다**). 즉 가드가 발화한 실행에서
**명령과 콘솔은 존재하는 로컬 큐를 못 보고 빈 큐를 렌더한다.**

**이것은 SPEC이 스스로 금지한 실패와 같은 부류다.** `plan.md:55` D12는 `""` 반환을 금지하며 그
근거로 "콘솔은 반환값을 `vm.Root`에 그대로 싣고 `BacklogPathForRoot(root)`로 읽으므로 … 거부가
아니라 **조용한 오작동**"이라고 적는다. **그 논거의 전제가 그대로 이 선택도 반박한다** — 같은
합성(`BacklogPathForRoot(반환값)`)이 계층이 어긋난 값에서도 엉뚱한 경로를 만든다.

**같은 부류의 사고가 이 리포에 이미 기록돼 있다.** `internal/kanban/todo_root_contract_test.go:9-13`
(`TestAdoptionLandsWhereConsumersRead`) doc: "adoption used to write `<root>/backlog.json` while
callers resolve the store through `BacklogPathForRoot` — `<root>/.moai/state/todo/backlog.json` — so
an adopted queue was moved somewhere nothing reads and **the operator's cards silently disappeared**".
방향만 반대인 같은 계층 오류이고, 그때는 테스트가 잡았다.

**`base` 기각은 뒤집힌 판단이다.** 수리는 `base`를 "덜 정밀"하다며 기각했으나, `base`야말로 계층이
맞는 값이고 **이 파일이 이미 그 이름으로 쓰고 있다**: `fallbackTodoQueueRoot`의 read-through가
`:148`에서 `return base`를 하고, 파일 머리말(`todo_root.go:22-24`)이 그것을 "resolves to the
**PROJECT-LOCAL root**"라고 부른다. 반면 `:121`의 반환은 같은 함수(`homeTodoQueueRoot`)의 다른
가지(`:124`, 프로젝트-루트 형태)와도 계층이 어긋난다 — **`:121`은 이 파일에서 유일하게 계층이
다른 반환이며, 기존의 잠복 결함으로 보인다**(기존 테스트 `todo_root_test.go:203-217`은 반환값만
단언하고 `BacklogPathForRoot`와 합성해보지 않아 이 잠복을 덮고 있다). 수리는 그 잠복을 **본 카드의
주 경로 사양으로 승격시켰다** — `HomeDirFn` 실패라는 희귀 경로에만 있던 것이, 이제 임시-기원
거부가 발화하는 **모든** 실행에서 돈다.

— **Required fix (택일)**:
  ⒜ 대체 루트를 **`base`**로 고정한다(`:148`의 read-through가 이미 돌려주는 값, 머리말이
     "PROJECT-LOCAL root"라 부르는 것). 한 낱말 수정이다.
  ⒝ 값이 아니라 **성질**로 쓴다 — "the substitute root shall be a root through which
     `BacklogPathForRoot` resolves the project's existing local queue" — 구현 좌표를 REQ에 박지
     않아도 되고(RQ-4), 계층 오류가 요구사항 수준에서 표현 불가능해진다. **권장.**
  어느 쪽이든 `plan.md` D12의 `base` 기각 문장과 `acceptance.md:36`·`:42`의 Then절을 함께 고친다.
  `todo_root.go:121`의 잠복 계층 오류는 **본 카드 범위 밖**이므로 고치지 말고 별도 카드로 등재하되,
  REQ가 그 좌표를 **본뜨지 않도록** 하는 것이 이 수정의 요점이다.

### D10. AC-THG-001이 「반환 루트로 큐가 읽히는가」를 단언하지 않는다 — D9가 GREEN으로 통과한다

— `acceptance.md:36`, `:42-43` — Severity: **major** — Class: **blocking** (D9와 한 뿌리이며,
D9를 ⒜로만 고치면 이 구멍은 남는다)

(a)의 Then은 "**project-local state directory를 반환한다**"는 **문자열 동일성**만 단언하고, (b)의
Then은 "반환 루트는 (a)와 같고 … 로컬 `backlog.json`은 제자리에 남아 있다"까지만 간다. **둘 중
어느 것도 「그 반환 루트로 로컬 큐가 실제로 읽히는가」를 묻지 않는다.** 그래서 D9의 계층 오류는
두 갈래 모두에서 **GREEN으로 통과한다** — 카드가 (b) 픽스처로 힘들여 만든 실제 큐가 렌더에서
사라지는 채로.

이 리포에는 그 성질을 정확히 고정하는 선례가 이미 있다 — `todo_root_contract_test.go:22-26`이
`os.Stat(BacklogPathForRoot(root))`로 "consumers read where the queue is"를 단언한다. AC에는
대응하는 단언이 없다.

— **Required fix**: AC-THG-001 (b)에 단언 1개 추가 — "**And** 반환 루트에 대해
`BacklogPathForRoot(root)`가 (b)에서 세운 로컬 `backlog.json`을 가리키고 그 파일이 stat 가능하다".
이 단언이 있으면 D9는 RED로 잡히고, 없으면 잡히지 않는다.

---

## 3. 카드가 「가장 중요하다」고 지목한 항목 — D2 수리의 판별력 검증

**판정: 진짜로 판별한다. 두 갈래가 서로 다른 기제로 실패한다.** 코드 판독으로 전 경로를 따라갔다.

뮤턴트(판별식 상수 `false`) 아래에서 동작은 오늘의 동작으로 되돌아간다:

| 갈래 | 기제 | 실패하는 단언 |
|---|---|---|
| (a) 로컬 큐 없음 | `primaryCheckoutRoot` 실패 → `homeTodoQueueRoot` ok → `adoptLocalTodoQueue`가 `:172-174`에서 조기 반환 → `fallbackTodoQueueRoot`가 홈 루트 반환 | **반환 루트**가 `canaryHOME/.moai/todo/<key>`로 되돌아간다 → (a) Then FAIL. 디렉터는 생기지 않으므로 이 갈래는 오염을 재현하지 **못한다** — `acceptance.md:125`가 그 사실을 정확히 적었다 |
| (b) 로컬 `backlog.json` 있음 | `os.Stat(target)` 실패 → `os.Stat(local)` **성공**(픽스처가 `BacklogPathForRoot(base)`에 씀) → **`:175` `MkdirAll` 실행** → `:179` `Rename`으로 로컬이 홈으로 이동 | `canaryHOME/.moai/todo/<key>/.moai/state/todo`가 **실제로 생성된다** → (b) And절("디렉터 0") FAIL. 덤으로 "로컬 `backlog.json`이 제자리에 남아 있다"도 FAIL — 이동됐으므로 |

즉 (b)에서 **오염이 실제로 재현된다**는 주장은 참이다. 1차 D2가 지적한 공허("가드 없는 트리에서도
And절이 이미 참")는 (b) 픽스처 도입으로 닫혔다 — `MkdirAll` 경로가 도달 가능해졌기 때문이다.
`acceptance.md:150-151`(DoD)이 "(b)가 `MkdirAll` 경로를 실제로 밟는다는 것이 RED 실측 출력으로
보인다"까지 요구해, 그 사실 자체가 관측 대상으로 걸려 있다. **수리 확인.**

다만 §2의 D10이 이 판별력의 **범위**를 제한한다: 뮤턴트는 「가드가 발화했는가」를 판별하지만,
「발화 후 돌려준 루트가 쓸 수 있는 루트인가」는 어느 단언도 묻지 않는다.

---

## 4. D5 적출 — **내 1차 판정이 뒤집혔다. 철회한다.**

**철회한다. `(Event-detected)`는 GEARS 정본 5패턴의 다섯 번째가 맞다.** 이 트리에서 직접 판독:

- `.claude/skills/moai-workflow-spec/SKILL.md:59` — GEARS 5패턴 표의 다섯째 행:
  `Event-detected (replaces IF/THEN)` | `"**When** <undesired-condition-detected>, the <subject>
  shall <behavior>"`. 표 어디에도 `Unwanted`는 없다.
- 같은 파일 `:82` — **legacy EARS** 표의 넷째 행에 `Unwanted`가 있다.
- `.claude/agents/moai/plan-auditor.md:72`(이 트리의 판본) — "Event-detected: … **the fifth GEARS
  pattern**", `:73` — "Unwanted … **NOT a GEARS pattern** — legacy EARS negative usage only".

**내가 1차에서 쓴 열거({Ubiquitous, Event-driven, State-driven, Where, Unwanted})는 두 표를 섞은
것이었다** — 다섯째 자리에 GEARS의 `Event-detected` 대신 legacy EARS의 `Unwanted`를 끼워 넣었다.
그 결과 정본 라벨을 결함으로 적고, 진짜 legacy 라벨 3건(`REQ-THG-005/007/008`의 `(Unwanted)`)은
지적하지 않는 **정확히 반대의 판정**을 냈다.

수리 쪽 처리도 옳다: `(Event-detected)`를 유지하고, `(Unwanted)` 3건은 호환 창(2026-11-22) 안의
legacy로 두었으며, `spec.md:112`에 **어느 표에서 온 라벨인지**를 명시했다 — 같은 오독이 재발하지
않게 하는 처리다. 지적을 받아들이는 대신 **근거를 대고 반박한 것이 옳았고, 그 반박이 이겼다.**

(잔여 관찰 1건은 §5 D12로 등재)

---

## 5. optional 결함 (판정을 막지 않음 — 오케스트레이터 재량)

**D11. `/var/tmp` 제외 근거 한 문장이 측정되지 않은 사전확률을 단정형으로 말하고, 재검토 조건이 관측 불가능할 수 있다**
— `spec.md:152` — Severity: minor — Class: **optional**
근거의 뼈대(지속성 임시 루트라 고아화 전제가 약하다 + 측정된 오염 기원은 `/tmp`다 + fail-open이라
회귀가 아니다)는 **건전하고**, 1차가 요구한 「등재」는 이행됐다. 두 군데가 약하다.
⑴ "장기 거주 디렉터가 실제로 살 확률은 `/tmp`보다 높다"는 **측정되지 않은 사전확률**인데 서술이
단정형이다 — 같은 §8이 다른 항목에서는 "설계 근거이지 측정이 아니다"라고 스스로 격하하므로, 그
자세를 이 문장에도 적용하면 된다.
⑵ 재검토 조건 "`/var/tmp` 기원의 고아가 **실측되면**"은 관측 가능성이 의심스럽다 — 고아 디렉터명은
sha256 4바이트 다이제스트라 기원 경로가 복원되지 않을 수 있고, **`proj-325ca0b6`이 바로 복원 실패
사례다**(`spec.md:153`). 트리거가 영원히 발화하지 않을 수 있다.
— Required fix: ⑴을 판단형으로 다시 쓰고, ⑵를 관측 가능한 트리거로 바꾼다(예: "`/var/tmp` 아래
비git 실행이 보고되면", 또는 후보 루트에 `/var/tmp`를 넣은 키 역산을 정기 수행).

**D12. `(Unwanted)` 3건은 legacy EARS 라벨이며 호환 창이 2026-11-22에 닫힌다**
— `spec.md:118`, `:120`, `:121` — Severity: minor — Class: **optional**
정본 GEARS 5패턴이 아니라 legacy EARS 표(`SKILL.md:82`)의 이름이다. **호환 창 안이므로 감점
사유가 아니고**, `spec.md:112`가 그 사실을 명시했으므로 내부 불일치도 아니다. 창이 닫히기 전
어느 시점에 정본형으로 옮길 여지가 있다는 사실만 남긴다 — **지금 고칠 필요는 없다.**

**D13. REQ-THG-001이 구현 좌표(`resolveStateDir(base, adopt=false)`)를 요구사항 본문에 담는다**
— `spec.md:114` — Severity: minor — Class: **optional** (D9의 ⒝ 경로를 택하면 함께 해소)
Group 3 RQ-4는 요구사항이 함수명 같은 구현 세부를 담지 않기를 요구한다. 이 리포 SPEC들이 좌표
고정을 관례로 쓰고 있어 관용 범위이나, **이 건에서는 그 관례가 대가를 치렀다** — 값을 성질이 아니라
호출로 못박은 탓에 계층 오류가 요구사항 안으로 그대로 복사됐다(D9). 성질로 쓰였다면 표현 자체가
불가능했을 오류다.

---

## 6. 카드가 물은 나머지 항목

**① 범위 경계(비임시 비git base의 홈 폴백)** — **여전히 깨끗하다. 위반 0건.**
`spec.md:118` REQ-THG-005(문구 무변경), `acceptance.md:67-78` AC-THG-003(등급 `blocking` 유지),
`plan.md:44` D1, `spec.md:133-135` Out of Scope, `plan.md:109` §G 안티패턴 — 네 겹 방어가 1차와
동일하게 서 있다. 수리 과정에서 **넓어진 요구사항은 하나도 없다**(REQ 8건 전문 대조).

**② Tier S** — **유지된다.** REQ 8 / AC 8(실측: `grep -c '^- \*\*REQ-THG-' spec.md` → 8,
`grep -c '^### AC-THG-' acceptance.md` → 8). D1을 REQ-THG-009 신설이 아니라 REQ-THG-001 확장으로
접은 것이 상한을 지켰다. **접기가 정직한가**: 정직하다고 판정한다 — 확장된 절은 「거부 시 리졸버가
무엇을 돌려주는가」이고, 이는 REQ-THG-001이 이미 지배하는 **한 해석 결과의 나머지 절반**이다
(거부만 말하고 결과를 말하지 않는 요구사항이 오히려 미완이었다). 두 번째 요구사항을 숨긴 것이
아니다. 다만 접힌 절에 딸린 정당화 문장("returns a bare `string` …")은 요구가 아니라 근거이므로
§4나 §8로 옮기는 편이 깔끔하다(optional).

**③ U1이 열린 채로 정직하게 다뤄지는가** — **그렇다. 이제 두 형태 모두에서 판정 가능하다.**
`REQ-THG-006`(`spec.md:119`)과 `AC-THG-005` Then절 ⑵(`acceptance.md:96`)가 선언형으로 다시 쓰여
(a)·(b) 어느 쪽에서도 지시대상을 갖는다. `acceptance.md:103`이 판정 대상 축소를 명시적으로
금지했고, `spec.md:151`·`plan.md:62`가 종전 과장을 **정정 기록으로 남겼다**(조용히 지우지 않았다).
`plan.md:60`은 "선택 전에 blocker 보고로 운영자 확인" 구속을 유지한다. **U1이 열려 있다는 것 자체는
결함이 아니고, 열린 채로 결정 가능하게 쓰였다** — 추가 AC 편집 없이 어느 형태로 확정되든 AC 8건은
그대로 판정된다. 단 하나의 예외가 D9다: (b)로 확정되면 명령이 「대체 루트로 계속」하는데, 그 루트가
D9 때문에 쓸 수 없는 경로다. **중립성은 성립하고, 중립성이 가리키는 값이 틀렸다.**

**④ 정리(t542) / `stateanchor`(t537) 경계** — **침범 없음.** `spec.md:127`·`:131` Out of Scope,
`REQ-THG-008`(`:121`), `plan.md:50` D7, `acceptance.md:141` 계수 불변 단언이 그대로다. D9의
Required fix에서 `todo_root.go:121`을 **본 카드에서 고치지 말라**고 명시한 것도 같은 규율이다.

**⑤ `plan.md:38`의 344** — **재측정 일치.** `find ~/.moai/todo -maxdepth 1 -type d | wc -l` → **344**.
행이 "자기 자신 포함 — 343 + 1"이라고 명시하므로 서술과 값이 어긋나지 않는다.

---

## 7. Must-Pass Results (2차 — 전수 재측정)

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-THG-001..008` 연속, 결번 0, 중복 0.
- **[PASS] MP-2 GEARS 형식 준수 (판정 계층: `spec.md` §6의 `REQ-XXX` 요구사항 계층)** — 8건 전수가
  GEARS 5패턴 또는 호환 창 안의 legacy EARS 등가다. Ubiquitous 002·003 / Event-driven 001·006 /
  **Event-detected 004(정본 — §4)** / legacy Unwanted 005·007·008. `acceptance.md`의 Given-When-Then은
  검증 계층이므로 이 기준으로 감점하지 않았다(Group 4에서 채점).
- **[PASS] MP-3 YAML frontmatter 유효성** — `spec.md:2-16` 정본 12필드 전수 존재·타입 적합
  (`version: "0.1.1"` 인용부호 유지, `created`/`updated` ISO). 거부 별칭 사용 0.
- **[N/A] MP-4 언어 중립성** — 단일 언어(Go) 내부 SPEC, 템플릿 바인딩 아님 → 자동 통과.
- **[PASS] MP-5 D7 cross-SPEC** — 참조 2건 실재 + `status: completed` ×2. retired/superseded/archived 0.
- **[PASS] MP-6 D8 크로스플랫폼** — `grep -c syscall spec.md` → **0** → 자동 PASS.
- **[PASS] MP-7 clarification gate** — `grep -n 'NEEDS CLARIFICATION' plan.md` → rc=1, 매치 0.
  `research.md` 부재(Tier S)이나 `plan.md`가 존재하므로 N/A가 아니라 실측 PASS.

**must-pass 7건 전수 통과. FAIL은 D9/D10(blocking)에 대한 판정이다.**

---

## 8. Category Scores (2차)

| Dimension | Score | 1차 대비 | Rubric Band | Evidence |
|-----------|-------|---|-------------|----------|
| Clarity | 0.75 | = | 0.75 | 1차 감점 사유(대체 루트 미정의)는 해소됐고 REQ-THG-006의 U1 편향도 사라졌다. 그러나 `spec.md:114`가 대체 루트를 **성질이 아니라 호출 좌표**로 못박아, 문자 그대로 읽는 구현자와 의도대로 읽는 구현자가 **다른 값**에 도달한다(D9·D13). |
| Completeness | 1.00 | ↑ | 1.0 | 전 섹션 + frontmatter 12필드 + `### Out of Scope — <topic>` H3 **5개** 각각 `-` 불릿(실측). 1차 감점이던 §8의 두 공백(대체 루트 축 `:149`, `/var/tmp` `:152`)이 등재됐고 미복원 1건(`:153`)까지 추가됐다. |
| Testability | 0.75 | = | 0.75 | AC 8건 전수 Given-When-Then + 판정 명령 + RED/GREEN 분리, weasel word **0건**(실측). 1차 감점 2건(거짓 RED, GREEN 축소) 해소. 새 감점: AC-THG-001이 반환 루트의 **문자열 동일성만** 단언해 D9가 GREEN 통과한다(D10). |
| Traceability | 1.00 | ↑ | 1.0 | REQ→AC 전수 피복(001→001·007, 002→002·006, 003→002, 004→004, 005→003, 006→005, 007→005·008, 008→008). 고아 AC 0, 미피복 REQ 0, 없는 REQ 참조 0. 1차 감점(`plan.md` M2 AC 오기)이 `:92`에서 정정됐다. |

**Aggregate: (0.75 + 1.00 + 0.75 + 1.00) / 4 = 0.875** (1차 0.75 → **상승**, 회귀 없음 → STOP 신호 없음)

---

## 9. 2차 측정 기록 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

**Baseline-attribution**: 아래 전부 이 감사 실행 중 이 트리(`.claude/worktrees/t536`)에서 직접
실행한 명령의 출력이거나 직접 판독한 파일 내용이다. **카드 지시문의 수리 주장 중 그대로 채택한
것은 없다.**

| # | Claim | Evidence |
|---|---|---|
| 1 | 루트 계층 사실 | `internal/kanban/state_dir.go:129-132` `BacklogPathForRoot(root)` = `resolveStateDir(root,false)` + `backlog.json`; `:46-48` `StateDirForRoot` = `root/.moai/state/todo`; doc `:126` "under a project root" |
| 2 | 온디스크 계층 실물 | `find ~/.moai/todo -name backlog.json \| head -3` → `~/.moai/todo/001-a49ab4ee/.moai/state/kanban/backlog.json` |
| 3 | 키 역산 **독립 재유도** | `python3` sha256 → `/tmp/t203-probe` → **d7a16ea2**, `/private/tmp/t203-probe` → **a8822e43**. `spec.md:95`의 두 값과 **정확히 일치** |
| 4 | adopt 조기 반환 / MkdirAll | `todo_root.go:169-170` target Stat, `:172-174` local Stat 조기 반환, `:175` `MkdirAll`, `:179` `Rename` |
| 5 | 픽스처 경로 | `todo_root_test.go:65-67` `seedLocalQueue` → `BacklogPathForRoot(base)` = `base/.moai/state/todo/backlog.json` |
| 6 | 잠복 계층 오류 | `todo_root.go:121` `resolveStateDir(base,false)` vs `:124` `home/.moai/todo/<key>` — 같은 함수 두 가지의 계층 불일치. 기존 테스트 `todo_root_test.go:203-217`은 반환값만 단언(합성 미검증) |
| 7 | 선례 사고 기록 | `todo_root_contract_test.go:9-13` doc — 같은 계층 오류의 반대 방향 사고와 결과("cards silently disappeared") |
| 8 | 소비자 합성 | `internal/web/todo_queue_read.go:33-35`(`vm.Root` + `BacklogPathForRoot`), `internal/cli/todo.go:56-57`(`BacklogPathForRootAdopting`) |
| 9 | launcher 실제 행 | 456 doc / 475 `func` / 484 `return filepath.Clean(path)` / **485 `}`** / 486 함수 밖 → 정본 `456-485` |
| 10 | GEARS 정본 표 | `SKILL.md:59` `Event-detected (replaces IF/THEN)`(GEARS 표) vs `:82` `Unwanted`(legacy EARS 표); `plan-auditor.md:72-73` |
| 11 | 구조 계수 | REQ 8 / AC 8 / Out of Scope H3 5 / weasel word 0 / `syscall` 0 / `NEEDS CLARIFICATION` 0 |
| 12 | 고아 계수 | `find ~/.moai/todo -maxdepth 1 -type d \| wc -l` → **344** (`plan.md:38` 기준값 일치) |
| 13 | 참조 SPEC 상태 | `grep -m1 '^status:'` → 둘 다 `completed` |

**Gaps — 이 감사에서 관측하지 못한 것 (명시)**

1. **D9의 경로 중복을 실행으로 관측하지 않았다.** 사전 구현 감사라 코드를 돌리지 않았다(그리고
   돌려서는 안 된다). D9는 두 함수(`BacklogPathForRoot`, `resolveStateDir`) **전문 판독에서
   기계적으로 유도한 합성**이며, 온디스크 계층(#2)과 기존 픽스처 경로(#5)가 각 항을 뒷받침한다.
   **실행 관측은 아니다.**
2. `todo_root.go:121`이 **의도된 설계인지 잠복 결함인지** 판정하지 못했다. 그 가지를
   `BacklogPathForRoot`와 합성해 단언하는 테스트도, 계층을 설명하는 주석도 찾지 못했다 — 부재는
   부재의 증거가 아니므로 "잠복 결함으로 **보인다**"까지만 말한다. D9의 판정은 이 물음과
   **독립적이다**: 그 가지가 무엇이든, 본 SPEC의 주 경로가 그 형태를 취하면 소비자 합성이 어긋난다.
3. Linux 셀 거동. 1차와 동일 — 이 기계는 darwin이고 SPEC 자신이 Gap으로 기록한 항목이다(`spec.md:147`).
4. `proj-325ca0b6`의 기원. 1차에서 복원 실패했고 이번에 재시도하지 않았다. SPEC이 미검증으로
   등재했으므로(`:153`) 결함이 아니다.
5. `go test` 실행. 사전 구현 감사이므로 돌리지 않았다. `plan.md §C` baseline은 run-phase 진입 시
   재측정 대상으로 남아 있다.

**Residual-risk**

- D9의 수정 방향 ⒜(`base` 고정)는 read-through 선례(`:148`)와 일치하지만, `base`가 큐 루트일 때
  `adoptLocalTodoQueue`가 호출되지 않는다는 전제(`plan.md:86`)가 지켜져야 부작용이 없다. 이
  상호작용은 설계 수준에서만 따라갔고 실행으로 확인하지 않았다.
- D2 수리의 판별력 판정(§3)은 구현이 SPEC 규격을 따를 때만 유효하다. run-phase가 이탈하면 무효이며,
  그것을 잡는 것이 AC-THG-007의 역할이다.
- 나는 1차에서 D5를 **정본을 결함으로 적는** 형태로 틀렸다. 같은 종류의 오류가 이번 판정 어딘가에
  또 있을 수 있다 — 그래서 이번 GEARS 관련 판정은 전부 파일 인용(#10)으로만 했다.

---

## 10. Recommendation

**FAIL. 남은 것은 두 건이며 둘 다 국소 수정이다.**

1. **D9 (필수)** — REQ-THG-001의 대체 루트를 고친다. **권장은 ⒝(성질로 기술)**: "the substitute
   root shall be a root through which `BacklogPathForRoot` resolves the project's existing
   project-local queue". 값으로 쓰겠다면 **`base`**다(`todo_root.go:148`이 이미 돌려주고 머리말이
   "PROJECT-LOCAL root"라 부르는 값). `plan.md` D12의 `base` 기각 문장과 `acceptance.md:36`·`:42`의
   Then절을 함께 고친다. `todo_root.go:121`은 **건드리지 않는다**(별도 카드).
2. **D10 (필수)** — AC-THG-001 (b)에 단언 1개 추가: 반환 루트에 대해 `BacklogPathForRoot(root)`가
   (b)의 로컬 `backlog.json`을 가리키고 stat 가능하다. 이 단언이 없으면 D9를 고쳐도 **같은 계층
   오류가 다음에 다시 GREEN으로 통과한다.**
3. **D11·D12·D13 (optional)** — 라우팅 여부는 오케스트레이터 재량. 판정을 막지 않는다. D13은 D9를
   ⒝로 고치면 함께 해소된다.

**감사자의 총평.** 수리는 **정직하다.** 결함 6건이 실제로 닫혔고, 닫는 방식이 좋았다 — 종전의 거짓
문장을 **지우지 않고 정정 기록으로 남겼고**(`acceptance.md:49`, `:105`, `spec.md:151`, `plan.md:62`,
`:94`), 근거를 대고 **내 오판 하나를 반박해 이겼다**(D5). 반박이 옳았다는 사실 자체가 이 수리의
품질을 말한다.

FAIL은 수리의 성실성이 아니라 **한 값의 계층**에 대한 판정이다. D1은 「미결인 줄 모르는 미결」
이었고 수리는 그것을 결정으로 바꿨다 — 옳은 방향이다. 다만 결정의 값을 고를 때, 이 파일에서
**유일하게 계층이 어긋난 반환**을 본떴다. 그리고 SPEC 자신이 `""`를 금지하며 적어 둔 논거가 그
선택도 함께 반박한다. 두 문장 고치면 닫힌다.

_감사자: plan-auditor · 2차 · 이 트리에서 직접 측정_

---

# SPEC Review Report — 3차 (0.1.2 재감사): SPEC-TODO-HOME-TEMP-GUARD-001

> **1·2차 기록은 위에 그대로 있다.** 이 절은 추가이며 덮어쓰기가 아니다. 1차(FAIL 0.75, `0.1.0`)와
> 2차(FAIL 0.875, `0.1.1`)는 각각 그 시점 산출물에 대한 판정이고, 아래는 `version: 0.1.2`에 대한
> 별개 판정이다.

Iteration: 3 / 3 (**하드 상한 도달**)
Verdict: **FAIL**
Overall Score: **0.8125** (Tier S 임계값 0.75 — 임계값은 넘겼으나 blocking 결함이 남았다)
**STOP 신호 발화** (아래 §7)

측정 기준 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t536` · `WT-home-fallback` @ `412c8cb14`.
아래 모든 인용·측정은 이 감사 실행 중 이 트리에서 직접 관측했다.

M1 Context Isolation: 카드 지시문이 제시한 수리 주장은 **하나도 그대로 채택하지 않았다.** 각 항목을
파일과 코드에서 재측정했고, 한 건은 **실행으로** 반증했다(§3).

**2차가 지적한 것은 셋 다 제대로 닫혔다.** D9의 계층 오류는 요구사항 수준에서 표현 불가능해졌고,
D10의 읽힘 단언은 D9를 실제로 RED로 잡으며, U1 가지치기에 잔여물이 없다. 그럼에도 FAIL인 이유는
**2차가 재지 않은 축에서 결함 셋이 나왔기 때문**이다 — 점수 하락은 산출물의 악화가 아니라 측정
면적의 확대다(§7에서 이 구분을 명시한다).

---

## 1. 2차 결함의 처분 (전수)

| # | 처분 | 근거 (이 감사에서 직접 관측) |
|---|---|---|
| D9 | **수리 (확인됨)** — 판정은 §2 | `spec.md:115` REQ-THG-001이 대체 루트를 `base`로 정정하고 성질로 기술. `plan.md:58` D12가 종전 값과 「덜 정밀」 기각 문장을 인용한 채 판단 역전을 기록. `plan.md:90` M2가 `:121` 형태를 본뜨지 말라고 구속. `acceptance.md:36`·`:43` Then절이 함께 바뀜 |
| D10 | **수리 (확인됨, 단 부수 주장 1건 거짓 — D14)** | `acceptance.md:37`이 (a)에 경로 동일성, `:44`가 (b)에 `os.Stat(BacklogPathForRoot(반환 루트))` + 파일 동일성을 추가. 선례 인용 `todo_root_contract_test.go:22-26` 실측 확인, doc `:9-13` 실측 확인 |
| D11 | **수리 (확인됨)** | `spec.md:152`가 사전확률을 "**판단한다** — 이는 설계 판단이지 측정이 아니다"로 격하하고 측정된 사실 1건을 분리. 재검토 트리거가 「`/var/tmp` 아래 비git 실행이 보고되거나 재현되면」으로 교체 — 관측 가능하다(§5) |
| D12 (optional) | **재론 없음** | `(Unwanted)` 3건은 호환 창(2026-11-22) 안이고 `spec.md:113`이 라벨 출처를 명시. 2차에서 D5를 철회한 사안이므로 **새 근거 없이 새 형태로 재기하지 않는다** |
| D13 (optional) | **수리 (확인됨)** | `resolveStateDir` 좌표가 어느 REQ·AC 본문에도 남아 있지 않다(전수 grep: HISTORY `spec.md:25`, §8 관측 `:155`, `plan.md:58`·`:90` 정정 기록뿐). RQ-4 위반 해소 |

**2차 인용 정정 확인**: 카드가 지적한 대로 내 2차 인용 `internal/cli/todo.go:56-57`·`:71-79`는 어긋났다.
실측 — `:57` = `return kanban.BacklogPathForRootAdopting(root)`, `:72` = `return
kanban.ResolveTodoQueueRootAdopting(resolveProjectDir())`. **수리 쪽 좌표가 옳다.**
`internal/cli/launcher.go:456-485`도 재실측 — `456` doc 시작, `475` `func`, `484` `return
filepath.Clean(path)`, `485` 닫는 중괄호. 정본 범위 맞다.

---

## 2. D9 성질 정식화에 대한 판정 — **성립한다. 느슨하지 않다.**

카드가 특정해 물은 항목이다. 판정: **그 성질은 틀린 루트를 받아들이지 않는다.**

근거는 `BacklogPathForRoot`가 루트에 대해 **단사**라는 것이다. 직접 판독:
`BacklogPathForRoot(R)`(`state_dir.go:129-132`) = `resolveStateDir(R,false)` + `backlog.json` =
`R/.moai/state/{todo|kanban}/backlog.json`. `R`이 결과 경로의 접두 성분으로 그대로 나타나므로
서로 다른 두 루트가 같은 큐 경로를 낼 수 없다. 따라서 「`BacklogPathForRoot(<반환 루트>)`가
프로젝트의 로컬 백로그 파일을 이름 부른다」를 만족하는 `R`은 **`base` 하나뿐**이고,
`resolveStateDir(base,false)`는 이중 경로(`base/.moai/state/todo/.moai/state/todo/backlog.json`)를
내므로 **성질 자체가 배제한다.** 0.1.1의 계층 오류는 요구사항 수준에서 표현 불가능해졌다 —
수리가 주장한 그대로다.

`todo|kanban` 두 이름 중 어느 쪽이 뽑히느냐(`state_dir.go:36`·`:42` 상수, 레거시 재배치 규칙)는
**루트 모호성을 만들지 않는다** — 어느 이름이 뽑히든 루트는 여전히 `base`다.

**다만 성질은 자족적이지 않다** — optional 결함 D17로 등재한다(§6). 성질의 지시대상이
「프로젝트의 **existing** project-local backlog file」인데, 가드가 발화하는 흔한 경우인
갈래 (a)(비git 임시 디렉터, 로컬 큐 없음)에는 **그 파일이 존재하지 않는다.** 문자 그대로 읽으면
그 갈래에서 성질은 지시대상을 잃어 아무것도 구속하지 못한다. 지금 구속을 지탱하는 것은 성질이
아니라 같은 문장의 **값 절**("shall instead resolve to the launch base itself")이고,
`acceptance.md:37`이 (a)를 경로 동일성으로 따로 닫아 실질적 구멍은 없다. 그래서 blocking이
아니라 optional이다 — 값 절을 빼고 성질만 남기면 그때 blocking이 된다.

---

## 3. D10 뮤턴트 주장에 대한 판정 — **결론은 참, 그러나 기재된 기제 하나가 거짓이다 (D14)**

카드가 특정해 물은 항목이고, **이 SPEC에서 한 번 공허해진 전력이 있는 자리**다. 그래서 판독으로
멈추지 않고 **실행으로 쟀다.**

**핵심 관측 — 지금 이 트리가 곧 뮤턴트다.** AC-THG-007의 뮤턴트는 「판별식을 상수 `false`로
무력화」한 상태이고, 가드가 아직 없는 현재 트리의 동작이 정확히 그 상태다. 그러므로 뮤턴트 아래
갈래 (b)의 거동은 **오늘 실행해서 관측할 수 있다.**

```
$ go test ./internal/kanban/ -run 'TestAdoptionLandsWhereConsumersRead|TestResolveTodoQueueRootAdopting_AdoptsLocalQueue' -count=1 -v
=== RUN   TestAdoptionLandsWhereConsumersRead
--- PASS: TestAdoptionLandsWhereConsumersRead (0.16s)
=== RUN   TestResolveTodoQueueRootAdopting_AdoptsLocalQueue
--- PASS: TestResolveTodoQueueRootAdopting_AdoptsLocalQueue (0.17s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	1.083s
```
(this run, this tree, HEAD `412c8cb14`)

두 테스트의 픽스처는 **비git `t.TempDir()` base + `seedLocalQueue`** — `acceptance.md:41`이 갈래 (b)의
픽스처 모형으로 이름 부른 바로 그것이다. 그 픽스처에서 관측된 것:

- `TestAdoptionLandsWhereConsumersRead`(`todo_root_contract_test.go:22-26`)는
  `ResolveTodoQueueRootAdopting(proj)`의 반환 루트에 대해 `os.Stat(BacklogPathForRoot(root))`가
  **성공**해야만 통과한다. 통과했다 → 뮤턴트 아래 (b)에서 그 경로는 **stat 가능하다**.
- `TestResolveTodoQueueRootAdopting_AdoptsLocalQueue`(`todo_root_test.go:229-247`)는 반환 루트가
  `home/.moai/todo/<key>`이고 그 루트로 로드한 큐가 픽스처의 3항목임을 단언한다. 통과했다 →
  거기 있는 파일은 **픽스처 자신의 큐**다(내용 동일).

기제는 코드가 그대로 말한다: `adoptLocalTodoQueue`의 target이 `BacklogPathForRoot(fallbackRoot)`
(`todo_root.go:168`)이고 `:179`의 `Rename`이 로컬 파일을 **정확히 그 경로로** 옮긴다. 그 뒤
`fallbackTodoQueueRoot`의 `:142`가 그 파일을 보고 홈 루트를 반환한다.

### D14 (blocking, major) — `acceptance.md:134`가 뮤턴트 FAIL 기제를 거꾸로 적었다

— `acceptance.md:134` — Severity: **major** — Class: **blocking**

기재: "(b)에서는 그 경로가 **stat 불가**(픽스처는 `base` 아래에 썼고, 게다가 `Rename`으로 이동됐다)라
FAIL한다."

**두 군데가 거짓이다.** ⑴ 그 경로는 stat **가능**하다(위 실행). ⑵ 인과가 뒤집혔다 — `Rename`은
stat 실패의 이유가 아니라 **stat 성공의 이유**다. 옮겨간 목적지가 바로
`BacklogPathForRoot(반환 루트)`이기 때문이다(`todo_root.go:168`·`:179`).

**귀결**: 0.1.2가 D10 수리로 추가한 읽힘 단언은 뮤턴트 아래 **갈래 (b)에서 공허하게 통과한다** —
FAIL 사유를 하나 더하기는커녕, 반환 루트가 홈으로 되돌아간 상태에서도 만족된다. 그래서
`acceptance.md:134`의 결론 문장("새 단언은 각 갈래에 FAIL 사유를 하나씩 더할 뿐")도 (b)에 대해
거짓이다.

**결론 자체는 무너지지 않는다.** (b)는 나머지 단언 셋으로 여전히 FAIL한다 — 반환 루트 ≠ `base`,
canary HOME 아래 디렉터 ≠ 0, 로컬 `backlog.json`이 제자리에 없음(이동됨). 그리고 (a)에서는 새
단언이 **진짜로** FAIL을 더한다: 반환 루트가 홈 루트이므로 `BacklogPathForRoot(홈 루트)`는 로컬
canonical 경로와 불일치한다. 또한 새 단언은 **D9를 잡는 일은 제대로 한다**(§2). 즉 손상된 것은
「뮤턴트 판별 증거」의 **기재된 근거**이지 판별력 자체가 아니다.

**그럼에도 blocking인 이유 둘.** ⑴ 1차 D2가 **같은 부류**(현재 트리에서 재현되지 않는 RED/FAIL
서술)를 blocking으로 매겼다 — 같은 SPEC 안에서 등급을 달리할 근거가 없다. ⑵ run-phase에 실질
위험이 있다: 이 문장을 읽고 뮤턴트 관측을 설계한 구현자는 stat 실패를 기대하다가 성공을 보고
「테스트가 틀렸다」로 오진하거나, 반대로 그 단언 하나만 보고 (b)의 뮤턴트 포착을 확인했다고
기록할 수 있다.

— **Required fix**: `acceptance.md:134`를 실제 기제로 다시 쓴다. (b)의 FAIL 사유는 **반환 루트 ·
디렉터 계수 · 로컬 파일 이동** 셋이고, 읽힘 단언은 (b)에서는 뮤턴트에 **공허**하며 (a)에서
FAIL을 더하고 D9 계열 계층 오류를 양 갈래에서 잡는다 — 이렇게 적으면 참이다. 위 실행 출력을
근거로 인용할 수 있다.

---

## 4. 신규 blocking 결함 — 2차가 재지 않은 축

### D15. 가드가 깨뜨릴 기존 테스트가 열거되지 않았고, 살아남는 테스트의 공허화도 보이지 않는다

— `plan.md:37` (§C 사전 점검 행) · `plan.md:95` (M2 무파괴 정의) · `acceptance.md:161` (DoD)
— Severity: **major** — Class: **blocking**

계획은 가드 착지로 **의도적으로 갱신되는 테스트를 정확히 1개** 이름 부른다 —
`TestResolveTodoQueueRoot_FallbackNoGit`. 그리고 무파괴를 「**미갱신** 테스트 무실패」로 정의한다
(`plan.md:95`). 이 트리를 재보면 그 전제가 성립하지 않는다.

리졸버를 호출하는 테스트를 전수로 뽑고(`ResolveTodoQueueRoot` 전수 grep, `--include='*_test.go'`)
base가 `t.TempDir()`(= 임시 기원 ⇒ 가드 발화)인 것을 판독한 결과:

**확실히 깨진다 (계획에 없음, 3건 · 2개 패키지)**

| 테스트 | 좌표 | 깨지는 단언 | 소관 AC |
|---|---|---|---|
| `TestResolveTodoQueueRoot_PopulatedFallbackWins` | `internal/kanban/todo_root_test.go:196` | `got != fallbackRoot` — 가드가 `base`를 반환하므로 홈 루트 단언 실패 | 읽기 우선순위(D-2) |
| `TestResolveTodoQueueRootAdopting_AdoptsLocalQueue` | `internal/kanban/todo_root_test.go:229`·`:233` | 반환 루트 = 홈 루트 실패 + 「로컬 파일이 사라졌다」 실패(가드가 adopt를 건너뛰므로 제자리에 남음) | **`SPEC-WEB-TODO-QUEUE-001` AC-WTQ-008** |
| `TestResolveTodoQueueRoot_FallbackNoGit` (CLI 미러) | `internal/cli/todo_queue_root_test.go:117-129` | 홈 루트 단언 실패. `dir := t.TempDir()` + `CLAUDE_PROJECT_DIR=dir` | — |

**가드 배치에 따라 깨진다 (1건)** — `TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing`
(`internal/kanban/todo_root_test.go:210-213`). D16 참조.

**통과하지만 공허해진다 (4건)** — 가드가 `base`를 돌려주는 덕에 단언은 만족되지만, 그 테스트들이
쓰인 이유인 **홈 폴백·adopt 가지를 더는 밟지 않는다**:
`TestResolveTodoQueueRoot_PureFallbackWritesNothing`(`:125`, AC-WTQ-006),
`TestResolveTodoQueueRoot_ReadThroughToProjectLocal`(`:156`, AC-WTQ-007),
`TestAdoptionLandsWhereConsumersRead`(`todo_root_contract_test.go:14` — 「카드가 조용히 사라진」
사고의 회귀 가드),
`TestAdoptingAndPureResolversAgreeWhenAdoptionFails`(`:40`).
`internal/web/todo_section_test.go:194` `TestTodoSectionReadsThroughToProjectLocalQueue`(AC-WTQ-007
콘솔 측)도 같은 모양이다 — **세 번째 패키지**다.

**왜 blocking인가 셋.**
⑴ `plan.md:95`의 무파괴 판정이 **M2에서 반드시 위반된다** — 계획이 스스로 세운 종료 조건이 그
시점에 거짓이 된다.
⑵ 깨지는 것 중 둘이 **다른 SPEC(`SPEC-WEB-TODO-QUEUE-001`)의 AC 산출 테스트**다. 결정 없이 새
동작에 맞춰 고쳐 쓰면 AC-WTQ-008(adopt-not-shadow)을 **조용히 철회**하게 된다. `plan.md:95`의
보존 지시("원 의도를 비임시 비git base 픽스처로 옮겨 보존한다")는 **이름 불린 1건에만 걸려 있다.**
⑶ 공허화는 **무파괴 기준으로 원리상 보이지 않는다** — 공허하게 통과하는 테스트는 실패하지
않는다. 이 리포가 반복해서 대가를 치른 「공허한 초록」 부류이고, 그중 하나는 회귀 가드다.

— **Required fix**: `plan.md` §C에 영향 테스트 전수(위 표 + 공허화 5건)를 등재하고, §F M2에
**테스트별 처분**을 적는다 — 「비임시 비git 픽스처로 이관해 원 의도 보존」인지 「의도적 갱신」인지.
`SPEC-WEB-TODO-QUEUE-001`의 AC 산출 테스트 3건(AC-WTQ-006/007/008)은 **보존 쪽으로 명시**한다.
공허화 5건에는 「비임시 픽스처 사본을 남겨 원 가지를 계속 밟게 한다」를 M2 종료 조건에 넣는다 —
그러지 않으면 무파괴 판정이 그것들을 볼 수 없다. `acceptance.md:161`의 DoD 문장도 1건 → 전수로
넓힌다.

### D16. 「임시 기원 ∧ 홈 해석 불가」 교차에서 어느 가지가 이기는지 정해져 있지 않다

— `spec.md:115` (REQ-THG-001) · `spec.md:158` (§8 처분) · `plan.md:90-91` (M2 배선)
— Severity: **minor** — Class: **blocking**

`spec.md:158`은 본 SPEC과 `:121` 관측을 갈라놓는 근거로 「트리거가 다르다 — 본 SPEC은 **임시 기원**,
그 가지는 **홈 해석 불가**」를 든다. **두 트리거는 배타적이지 않다.** 비git 임시 base에서
`HomeDirFn()`이 오류를 내면 둘 다 성립하고, 그 교차에서 무엇을 반환해야 하는지 어느 요구사항도
말하지 않는다.

REQ-THG-001은 발화 조건을 「홈 폴백 가지에 도달(git이 답하지 못함) ∧ 임시 기원」으로 쓴다 —
홈 해석 실패 가지도 그 조건 안이다. 그래서 술어를 `fallbackTodoQueueRoot`의 `ok` 검사 **앞**에
넣으면 `base`를, **뒤**에 넣으면 `base/.moai/state/todo`를 돌려주게 되고, 후자는 §8이 관측으로
등재한 바로 그 계층 어긋난 값이다 — **본 카드가 본뜨지 않기로 한 형태를 배치 순서만으로 다시
집어 든다.** `plan.md:90-91`은 삽입 지점을 「`homeTodoQueueRoot`/`fallbackTodoQueueRoot` 가지」로
두 함수를 함께 지목할 뿐 순서를 정하지 않는다.

관측 가능한 귀결이 이미 있다: `TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing`
(`internal/kanban/todo_root_test.go:204-217`)은 `dir := t.TempDir()`(임시) + `HomeDirFn` 오류로
정확히 이 교차를 밟으며 `got == dir/.moai/state/todo`를 단언한다. 배치에 따라 깨지거나 안 깨진다 —
**구현자가 고르게 되어 있고, 계획은 그 선택을 blocker로도 결정으로도 다루지 않는다.**

— **Required fix**: REQ-THG-001(또는 `plan.md` D12)에 한 절 추가 — 교차에서 어느 반환이 이기는지.
성질에 비춘 정합한 답은 `base`다(그 성질을 만족하는 유일한 루트, §2). 그렇게 정하면
`HomeUnresolvableWritesNothing`이 **의도적 갱신 대상**이 되므로 D15의 목록에 함께 올린다.

---

## 5. optional 결함의 처분

- **D11 재검토 트리거의 관측 가능성 — 해결로 본다.** 새 트리거(`spec.md:152`, 「`/var/tmp` 아래
  비git 실행이 보고되거나 재현되면」)는 4바이트 다이제스트 역상에 의존하지 않는다. **수동적**
  트리거이긴 하다 — 아무것도 이 검사를 예약하지 않는다 — 그러나 fail-open 방향의 제외 항목에
  대해 그 정도는 균형이 맞는다. 재기하지 않는다.
- **`(Unwanted)` 3건** — 호환 창 안. **재기하지 않는다**(2차 D5 철회 사안, 새 근거 없음).

---

## 6. 신규 optional 결함

**D17. REQ-THG-001의 성질 절이 갈래 (a)에서 지시대상을 잃는다**
— `spec.md:115` — Severity: minor — Class: **optional**
성질이 「the project's **existing** project-local backlog file」을 지시하는데, 가드가 발화하는 흔한
경우인 로컬 큐 없는 임시 디렉터에는 그 파일이 없다. 문자 그대로는 그 갈래에서 성질이 아무것도
구속하지 못하고, 구속을 지탱하는 것은 같은 문장의 값 절이다. 실질 구멍은 없다 —
`acceptance.md:37`이 (a)를 「로컬 큐가 **놓이는** canonical 경로와의 동일성」으로 닫는다.
— Required fix(선택): 성질을 존재 무관형으로 다시 쓴다 — "…names the canonical location of this
project's project-local backlog file (whether or not it exists yet)". 값 절을 유지하는 한 지금
고치지 않아도 판정을 막지 않는다.

---

## 7. STOP 신호 — 발화하되, 읽는 법을 함께 남긴다

점수가 0.875(2차) → **0.8125**(3차)로 **하락**했다. LEAN 조항의 문자대로 `STOP`을 발화한다.

**다만 이 하락을 산출물의 악화로 읽으면 오독이다.** 2차가 잰 모든 축에서 0.1.2는 개선됐거나
유지됐다(§1 전수 표). 하락은 2차가 **재지 않은 축**(가드가 기존 테스트에 미치는 영향, 가드와
no-home 가지의 교차)을 3차가 재면서 Completeness·Clarity가 내려간 결과다. 구조적 결함이 남아
반복 수리가 무의미한 상황 — STOP 조항이 겨냥한 그 상황 — 이 **아니다.**

동시에 **iteration 3 = 하드 상한**이므로 이 카드는 운영자 판단으로 넘어간다. 남은 결함 셋은
전부 **국소 편집**이고 설계를 바꾸지 않는다: D14는 문장 하나 재작성, D15는 목록 등재 + 테스트별
처분 명시, D16은 절 하나 추가. 새 측정도, 새 결정도 필요하지 않다.

---

## 8. Must-Pass Results (3차 — 전수 재측정)

- **[PASS] MP-1 REQ 번호 일관성** — `REQ-THG-001..008` 8개 distinct, 결번·중복·패딩 불일치 없음
  (`REQ-THG-[0-9]*` 전수 추출 후 정렬). `AC-THG-001..008`도 동일.
- **[PASS] MP-2 GEARS 형식 준수 (요구사항 계층에 대해 판정)** — `spec.md:115-122` 8건 전부 GEARS
  5패턴 또는 호환 창 안의 legacy EARS 등가형. `spec.md:113`이 라벨 출처(`SKILL.md:55-59` 정본 표 /
  `:82` legacy 표)를 명시한다. **`acceptance.md`의 Given-When-Then은 검증 계층이므로 이 기준으로
  판정하지 않았다**(Group 4에서 판정 — 아래 Testability).
- **[PASS] MP-3 YAML frontmatter 유효성** — 정본 12필드 전수 확인(`spec.md:2-13`): `id` `title`
  `version:"0.1.2"`(따옴표) `status:draft` `created:2026-09-08` `updated:2026-09-08` `author`
  `priority:P2` `phase` `module` `lifecycle:spec-anchored` `tags`(CSV 문자열). 거부 별칭
  (`created_at`/`updated_at`/`labels`/`spec_id`) 0건.
- **[N/A] MP-4 언어 중립성** — Go 단일 언어 SPEC. 자동 통과.
- **[PASS] MP-5 D7 교차-SPEC 정합** — 산출물이 참조하는 SPEC은 둘: `SPEC-STATE-ANCHOR-001`
  (`status=completed`), `SPEC-WEB-TODO-QUEUE-001`(`status=completed`). 둘 다 실재하고
  {retired, superseded, archived}에 없다 → BLOCKING 없음. 미발견 참조 0건.
- **[PASS] MP-6 D8 크로스 플랫폼** — `syscall` 문자열 0건(`spec.md`/`plan.md`/`acceptance.md`
  각 0). 자동 통과.
- **[PASS] MP-7 clarification gate** — `[NEEDS CLARIFICATION` 전수 grep(4개 산출물) 매치 0(rc=1).
  U1은 `plan.md:63`에서 **결정 완료**로 기록됐고 미결 항목은 `progress.md:10`이 "없다"로
  단언한다 — 재측정 결과와 일치.

**must-pass 7항 전부 통과 또는 N/A.** FAIL은 blocking 결함 3건(D14·D15·D16)에서 온다.

---

## 9. Category Scores (3차)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 (한두 건의 경미한 모호성) | REQ-THG-001(`spec.md:115`)의 성질 절이 갈래 (a)에서 지시대상을 잃는다(D17). 「임시 기원 ∧ 홈 해석 불가」 교차의 반환이 미정(D16, `spec.md:158`·`plan.md:90-91`). 나머지 REQ 7건은 단일 해석 |
| Completeness | 0.75 | 0.75 (비핵심 항목 1건 희소) | 12필드 frontmatter 완비, 필수 섹션 전부, `### Out of Scope` H3 5개 각각 구체 불릿 보유(`spec.md:126,130,134,138,142`). **영향 범위 열거가 불완전**하다 — 가드가 깨뜨릴 기존 테스트 3건 + 공허화 5건이 `plan.md` §C/§F에 없다(D15) |
| Testability | 0.75 | 0.75 (1건이 정밀한 이진 판정이 아님) | AC 8건 전부 이진 판정 가능하고 판정 명령·픽스처·RED/GREEN이 붙어 있다. weasel word 0건. **AC-THG-007(`acceptance.md:134`)의 뮤턴트 FAIL 기제 1건이 실행으로 반증됐다**(D14, §3) — 결론은 참이나 기재된 근거가 거짓 |
| Traceability | 1.00 | 1.0 | 매트릭스(`acceptance.md:15-22`) 전수 검증: REQ-001→AC-001/007, 002→AC-002/006, 003→AC-002, 004→AC-004, 005→AC-003, 006→AC-005, 007→AC-005/008, 008→AC-008. **미커버 REQ 0, 고아 AC 0** |

**Overall: (0.75 + 0.75 + 0.75 + 1.00) / 4 = 0.8125.** Tier S 임계값 0.75 초과 — 그러나 임계값은
필요조건이지 충분조건이 아니며, blocking 결함이 남은 이상 판정은 FAIL이다(2차가 0.875에서 FAIL을
낸 것과 같은 규칙).

---

## 10. Defects Found (3차 · 구조화 목록)

D14. 뮤턴트 기제 오기 — `acceptance.md:134` — 갈래 (b)에서 `BacklogPathForRoot(반환 루트)`가
「stat 불가라 FAIL」이라 적었으나 실행 관측상 **stat 가능**하며, `Rename`(`todo_root.go:168`·`:179`)은
실패의 원인이 아니라 성공의 원인이다. 새 읽힘 단언은 (b)에서 뮤턴트에 **공허**하다 —
Severity: major — Class: **blocking** — Required fix: (b)의 실제 FAIL 사유 셋(반환 루트 · 디렉터
계수 · 로컬 파일 이동)으로 재작성하고, 읽힘 단언의 판별 범위를 (a) + D9 계열로 정확히 적는다.
§3의 실행 출력을 근거로 인용 가능.

D15. 영향 테스트 미열거 + 공허화 미포착 — `plan.md:37`, `plan.md:95`, `acceptance.md:161` —
계획이 갱신 대상 1건만 이름 부르나 **확실히 깨지는 것이 3건**(`internal/kanban/todo_root_test.go:196`,
`:229`·`:233`, `internal/cli/todo_queue_root_test.go:117-129`)이고 그중 둘은
`SPEC-WEB-TODO-QUEUE-001`의 AC 산출 테스트다. **통과하지만 공허해지는 것이 5건**(3개 패키지)이며
무파괴 기준으로는 원리상 보이지 않는다 — Severity: major — Class: **blocking** — Required fix:
§C에 영향 전수 등재, §F M2에 테스트별 처분(보존 이관 / 의도적 갱신) 명시, AC-WTQ-006/007/008 산출
테스트는 보존으로 지정, 공허화분에 비임시 픽스처 사본 유지를 M2 종료 조건에 추가, DoD를 전수로 확대.

D16. 가드 ∧ no-home 교차의 반환 미정 — `spec.md:115`, `spec.md:158`, `plan.md:90-91` — 두 트리거가
배타적이라는 §8의 분리 근거가 성립하지 않으며, 술어 삽입 순서만으로 §8이 「본뜨지 않는다」고 한
계층 어긋난 값이 다시 반환될 수 있다. 관측 지점은
`internal/kanban/todo_root_test.go:204-217` — Severity: minor — Class: **blocking** —
Required fix: 교차에서 `base`가 이긴다는 절을 REQ-THG-001 또는 `plan.md` D12에 추가하고, 해당
테스트를 D15 목록에 의도적 갱신으로 올린다.

D17. 성질 절의 지시대상이 갈래 (a)에 없다 — `spec.md:115` — Severity: minor —
Class: **optional** — Required fix(선택): "…names the canonical location of this project's
project-local backlog file (whether or not it exists yet)"로 존재 무관형 재작성.

---

## 11. Regression Check (2차 결함 전수)

- D9 — **RESOLVED**: `spec.md:115` 값+성질 정정, `plan.md:58` 판단 역전 기록, `plan.md:90` M2 구속.
  성질의 판별력을 `BacklogPathForRoot`의 루트-단사성으로 독립 확인(§2).
- D10 — **RESOLVED (부수 주장 1건은 D14로 승계)**: `acceptance.md:37`·`:44` 단언 추가로 D9 계열
  계층 오류가 양 갈래에서 RED로 잡힌다. 뮤턴트에 대한 기재는 §3에서 반증.
- D11 — **RESOLVED**: `spec.md:152` 판단형 격하 + 관측 가능한 트리거 교체.
- D12 — 재론 없음(호환 창 안). D13 — **RESOLVED**(좌표가 REQ·AC 본문에서 전부 제거됨).

**정체(stagnation) 신호 없음** — 세 회차에 걸쳐 변하지 않은 결함은 하나도 없다. 매 회차 지적이
실제로 닫혔고, 신규 결함은 새로 측정된 면적에서 나왔다.

---

## 12. 3차 측정 기록 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

**Claim 1** — 뮤턴트 아래 갈래 (b)에서 `os.Stat(BacklogPathForRoot(반환 루트))`는 성공한다.
**Evidence** — §3의 `go test` 실행 출력 전문(2 PASS). 두 테스트가 그 stat 성공과 파일 동일성을
각각 단언한다(`todo_root_contract_test.go:23-26`, `todo_root_test.go:236-241`).
**Baseline-attribution** — this run, this tree, HEAD `412c8cb14`, 가드 미구현 = 뮤턴트 등가 상태.
**Gaps** — 뮤턴트를 실제로 주입해 돌리지는 않았다(구현이 없으므로 주입할 판별식이 없다). 등가성은
「가드 부재 = 판별식 상수 false」라는 판독 논증이다. **Residual-risk** — 구현된 판별식이 상수
false와 다른 방식으로 무력화될 경우(예: 배치 지점이 달라 adopt 호출 자체가 남지 않는 경우) (b)의
FAIL 사유 구성이 달라질 수 있다.

**Claim 2** — 가드 착지로 최소 3건의 기존 테스트가 확실히 깨지고 5건이 공허해진다.
**Evidence** — 리졸버 호출 테스트 전수 grep + 각 테스트의 base 생성(`t.TempDir()`)과 단언 대상을
직접 판독(§4의 표에 좌표 기재). **Baseline-attribution** — this run, this tree, HEAD `412c8cb14`.
**Gaps** — **실행으로 확인하지 않았다** — 가드가 아직 없으므로 원리상 불가능하다. 판독 유도다.
**Residual-risk** — 구현이 술어를 `ResolveTodoQueueRoot`이 아닌 더 좁은 지점에 넣으면 파급 집합이
줄 수 있다. 다만 그 경우에도 `plan.md`가 파급을 **열거하지 않았다**는 결함은 남는다.

**Claim 3** — must-pass 7항 전부 통과 또는 N/A.
**Evidence** — §8의 항목별 명령과 관측값. **Baseline-attribution** — this run, this tree.
**Gaps** — `plan.md:38`의 기준값 344(실 HOME 아래 디렉터 계수)는 **재측정하지 않았다** — 워크트리
격리 가드가 실 HOME을 겨눈 명령을 거절했다. 카드가 불변으로 지정한 값이고 본 판정의 어느 근거도
그 수에 의존하지 않는다. **Residual-risk** — 그 값이 그 사이 변했다면 §C 사전 점검이 run-phase
진입 시 멈추게 된다(계획이 그렇게 설계돼 있으므로 fail-safe 방향이다).

**Gaps (감사 전체)** — ⑴ 리눅스 셀 미측정(SPEC 스스로 `spec.md:148`에 Gap으로 등재; CI 소관).
⑵ `proj-325ca0b6` 기원 미복원(SPEC이 Gap으로 등재, 결론에 무영향). ⑶ 실 HOME 계수 미재측정(위).
⑷ 판별식 자체가 없으므로 정규화 규칙의 실동작은 어느 것도 실행 검증하지 못했다 — 전부 run-phase
소관이며 AC가 그것을 겨눈다.

---

## 13. Recommendation

**운영자의 implementation-kickoff 게이트에는 아직 올릴 수 없다.** 막는 것은 셋이고, 셋 다
국소 편집이다. 설계·범위·결정은 건드릴 필요가 없다.

1. **`acceptance.md:134` 재작성** (D14) — 뮤턴트 아래 갈래 (b)의 FAIL 사유를 실제 셋으로 적고,
   0.1.2가 추가한 읽힘 단언이 (b)에서는 뮤턴트에 공허하며 (a)와 D9 계열에서 판별한다고 명시한다.
   §3의 실행 출력을 근거로 인용하면 재측정이 필요 없다.
2. **`plan.md` §C·§F M2에 영향 테스트 전수 등재 + 테스트별 처분** (D15) — 확실히 깨지는 3건,
   배치 의존 1건, 공허해지는 5건. `SPEC-WEB-TODO-QUEUE-001`의 AC 산출 테스트 3건은 **보존**으로
   지정하고, 공허화분에 비임시 픽스처 사본 유지를 M2 종료 조건에 넣는다.
3. **교차 반환 1절 추가** (D16) — 「임시 기원 ∧ 홈 해석 불가」에서 `base`가 이긴다.
   `HomeUnresolvableWritesNothing`을 2번 목록에 올린다.

D17은 재량이다 — 값 절이 구속을 지탱하는 한 판정을 막지 않는다.

**이 셋이 닫히면 PASS로 본다.** 그때 run-phase에서 계속 지켜볼 것 셋: ⑴ **뮤턴트가 실제로
무엇을 잡았는지의 출력 전문** — 이 SPEC은 같은 자리에서 이미 한 번 공허해졌고, 위 D14가 두 번째
사례다. 「못 잡은 뮤턴트도 보고한다」(D9)를 문자 그대로 이행하는지. ⑵ **갱신된 테스트의 원 의도가
비임시 픽스처로 실제로 옮겨졌는지** — 단언을 지우는 갱신이 다른 SPEC의 AC를 조용히 철회하는
경로다. ⑶ **리눅스 셀 판정을 macOS 통과로 대체하지 않는지**(`acceptance.md:69`가 스스로 금지한다).

**iteration 3 = 하드 상한 도달 + STOP 신호 발화**(§7). 이 카드는 운영자 판단으로 넘어간다.
판단 재료로 남기는 사실 둘: 남은 셋은 전부 국소 편집이고 새 측정·새 결정을 요구하지 않는다는 것,
그리고 점수 하락이 산출물의 악화가 아니라 측정 면적의 확대라는 것.

---

# SPEC Review Report — 4차 (0.1.3 부분 재감사, 운영자 예외 개설): SPEC-TODO-HOME-TEMP-GUARD-001

> **1·2·3차 기록은 위에 그대로 있다.** 이 절은 추가이며 덮어쓰기가 아니다.

Iteration: 4 (하드 상한 3 **초과** — 운영자가 D14·D15·D16 세 축을 닫기 위해 예외로 개설했다)
**범위 제한(운영자 지시)**: D14 · D15 · D16 **세 축 + `AC-THG-003` 픽스처 1건**만 판정한다. 1-3차에서
닫은 축(D9 성질, U1 가지치기, `spec.md §8` 관측의 정직성, `/var/tmp` 제외, 리눅스 조항, 범위 경계)은
재론하지 않는다. **D17은 여전히 열린 채**이며 새로 등급 매기지 않는다.
**Tier 판정은 하지 않는다** — 감사 실행 중 리드가 결정했다(fold 기각 → `REQ-THG-009` 분리, Tier M).
아래 판정은 **현재 파일 상태(`0.1.3`, 접힌 형태, `tier: S`)**에 대한 것이고, unfold가 바꿀 항목은
그 의존을 명시한다.

Verdict: **FAIL** (세 축 중 **둘 닫힘 · 하나 미닫힘**)
측정 기준 트리: `.claude/worktrees/t536` · `WT-home-fallback` @ `412c8cb14`. 아래 인용·측정은 전부
이 감사 실행 중 이 트리에서 직접 관측했다. 셸 `grep`은 ugrep 래퍼라 **모든 부재·계수 주장에
`/usr/bin/grep`을 썼다.**

M1 Context Isolation: 카드가 제시한 수리 주장은 하나도 그대로 채택하지 않았다. 두 건은 코드 판독으로
확인했고, **한 건(D15)은 열거 자체를 다시 수행해 반증했다**(§3).

---

## 1. D14 — **닫혔다.** 재작성된 절은 참이고, 분리는 실질이다

**⑴ 기재가 참인가 — 참이다.** `acceptance.md:154`가 갈래 (b)의 FAIL 사유 셋을 든다: 반환 루트가
`base`가 아니다 · canary HOME 아래 디렉터 계수가 0이 아니다 · 로컬 `backlog.json`이 제자리에 없다.
셋 다 코드에서 성립한다 — `adoptLocalTodoQueue`가 `todo_root.go:172`의 조기 반환을 넘어 `:175`
`MkdirAll` → `:179` `os.Rename(local, target)`을 실행하고, `:168`이 `target := BacklogPathForRoot(fallbackRoot)`다
(네 좌표 전부 이 감사에서 `/usr/bin/grep`으로 실측: `168 / 172 / 175 / 179`). 그리고 셋은 각각
`acceptance.md:43`·`:45`의 단언에 대응한다 — **기재된 사유가 실재하는 단언에 걸린다.**

**⑵ 읽힘 단언의 판별 범위 기재가 참인가 — 참이다.** `acceptance.md:136`이 (a)에서는 FAIL 사유를
더하고 (b)에서는 **뮤턴트에 공허하며** 양 갈래에서 D9 계열 계층 오류를 잡는다고 적는다. 셋 다
성립한다: (a)에서 뮤턴트 아래 반환 루트가 홈 루트이므로 경로 동일성이 깨지고, (b)에서는
`Rename` 목적지가 곧 `BacklogPathForRoot(반환 루트)`라 stat이 성공하며(3차 §3 실행 출력이
`acceptance.md:144-148`에 그대로 인용돼 있다), 반환 루트가 `base/.moai/state/todo`인 D9형 오류에서는
합성 경로가 `.../.moai/state/todo/.moai/state/todo/backlog.json`이 되어 (b)의 stat이 실패하고 (a)의
경로 동일성도 깨진다.

**⑶ 분리가 실질인가 — 실질이다.** `acceptance.md:152`가 「결론은 살아남는다 / 기재된 논거는
거짓이었다」를 **두 항으로 갈라 적고**, 「결론이 맞으니 됐다」로 넘기지 않겠다는 문장을 붙인 뒤
**구체적 오진 시나리오 둘**(stat 성공을 보고 「테스트가 틀렸다」로 오진 / 그 단언 하나로 (b) 포착을
확인했다고 기록)을 든다. 종전 문장도 지우지 않고 인용해 뒀다(`:140`). 장식이 아니라 run-phase
독자의 행동을 바꾸는 기재다.

**세 번의 공허한 초록을 `spec.md §8` 관측으로만 담은 처분 — 옳다.** `spec.md:172-175`가 세 인스턴스와
공통 모양을 적고 「각 인스턴스만 고치고 반복 자체를 요구사항으로 만들지 않는다 · 후속 카드 후보,
발행은 리드 소관」으로 끝낸다. 두 가지 이유로 맞는 봉쇄다. ⑴ 반복의 대상은 이 SPEC의 가드가 아니라
**검증 설계 일반**이라, 요구사항으로 접으면 REQ가 자기 범위 밖을 구속한다. ⑵ 카드 발행은 운영자의
행위이며 SPEC이 스스로 큐에 넣지 않는 것이 큐 규율과 일치한다. 잔여: 트리거가 수동적이라 아무것도
이 후속을 예약하지 않는다 — D11 처분과 같은 수준이므로 재기하지 않는다.

---

## 2. D16 — **닫혔다.** 값 고정이 의존을 실제로 제거한다

`spec.md:116` REQ-THG-001이 교차에서 **`base`가 이긴다**를 못박고, 이어서 술어를 홈 해석 결과보다
**먼저** 평가하라고 적는다. `spec.md:159` §8이 종전의 「트리거가 다르다」를 거짓으로 정정하고 교차의
실재를 인정한다. `plan.md:117` D12에 같은 구속이 미러돼 있고 관측 지점(`todo_root_test.go:204-217`)이
명시된다.

**의존이 제거됐는가 — 됐다. 그리고 제거하는 것은 순서 절이 아니라 값 절이다.** 요구사항이 교차의
반환값을 하나로 고정하는 순간, 어떤 배치를 택하든 그 값을 내지 못하는 구현은 요구사항 위반이 된다 —
구현자의 선택지가 「이기는 값」에서 「그 값을 내는 방법」으로 줄어든다. 순서 절은 그 값을 내는 **한
가지 방법**을 이름 부르는 파생 문장이다. 3차가 요구한 것("교차에서 어느 반환이 이기는지")은 값 절
하나로 충족되고, 순서 절은 잉여이되 해롭지 않다(요구사항 본문에 함수명이 없어 RQ-4 위반은 아니고,
함수 좌표는 `plan.md` D12에만 있다).

실측 확인: 그 테스트는 `dir := t.TempDir()` + `HomeDirFn` 오류로 정확히 교차를 밟고
`want := filepath.Join(dir, ".moai","state","todo")`를 단언한다(`todo_root_test.go:204-217`, 직접 판독).
`plan.md:63-66` B행이 이를 **의도적 갱신**으로 등재하고 `acceptance.md:183` DoD가 근거 인용을
요구한다 — 3차가 요구한 목록 등재도 이행됐다.

**잔여 (optional, 아래 D22)**: 이 새 절에는 **산출 AC가 없다.** AC-THG-001의 Given은 canary HOME이
스텁된(=해석 가능한) 상태만 다루므로 교차를 밟지 않는다. 지금 이 절을 지키는 것은 AC가 아니라
DoD 한 줄이다.

---

## 3. D15 — **닫히지 않았다.** 열거가 여전히 불완전하고, 새 판정식이 5건 중 4건에 공허하다

카드가 특정해 물은 축이고, 3차가 「전수」를 요구한 자리다. 그래서 판독으로 멈추지 않고 **열거를
처음부터 다시 수행했다.**

### D18 (blocking, major) — §C.1의 「전수 8건」이 전수가 아니다. 확실히 깨지는 것이 **최소 2건 더** 있고, 그중 하나는 **완료된 형제 SPEC의 AC 산출 테스트**다

— `plan.md:53-60`(A표) · `plan.md:80`(측정 방법) · `internal/cli/todo_queue_root_test.go:153` ·
`internal/cli/todo_axisa_guard_test.go:150` — Severity: **major** — Class: **blocking**

**⑴ `TestTodoQueue_FallbackAdoptsExistingLocalQueue`** (`internal/cli/todo_queue_root_test.go:153`) —
`dir := t.TempDir()`(비git) + `CLAUDE_PROJECT_DIR=dir` + `userHomeDirFn` 스텁. 본문이
`root := resolveTodoQueueRoot()` 뒤 `want := filepath.Join(home, ".moai","todo", kanban.TodoQueueProjectKey(dir))`
동일성을 단언하고, 이관된 3항목·상태·`last_seq=7`까지 단언한다. 가드가 착지하면 반환값이 `base`가
되어 **확실히 깨진다.** 주석이 스스로 `[HARD] verification 3 in code form`, 「ADOPTED — never
shadowed」라고 적는다 — **A표의 `AdoptsLocalQueue`와 같은 부류의 adopt-not-shadow 단언**이고, 지금
§C.1에 없다.

**⑵ `TestGuardBypassMutant_ObserveHomePollution`** (`internal/cli/todo_axisa_guard_test.go:150`) —
이것이 더 무겁다. 이 테스트는 canary HOME + 비git `t.TempDir()`에서 `newTodoCmd()`를 직접 실행해
**홈 오염이 발생하는 것을 단언**하며, 오염이 **없으면** `t.Fatalf`로 죽는다("mutant produced NO home
pollution … report the guard boundary this reveals"). 본 SPEC의 가드가 하는 일이 정확히 그 오염을
불가능하게 만드는 것이므로 **확실히 깨진다.**
그 테스트는 `SPEC-STATE-ANCHOR-001`의 **AC-SA-011 / REQ-SA-011 산출 테스트**다(실측:
`.moai/specs/SPEC-STATE-ANCHOR-001/acceptance.md:23`·`:152`, `spec.md:106`, 그 SPEC의
`status: completed`). 그리고 그 SPEC의 §5가 **본 카드가 이행하는 결정** 그 자체다.

**왜 blocking인가.** D15 수리의 존재 이유가 「다른 SPEC의 AC를 조용히 철회하는 경로를 막는 것」인데,
그 경로가 **막히지 않은 채 하나 더 열려 있다** — 그것도 이 카드가 이행하는 바로 그 SPEC을 향해.
`AC-WTQ-008`에 대해서는 리드가 전제 정합 판정을 문서로 남겼고(`spec.md:169` 이하, 아래 §5에서
검증한다), **AC-SA-011에 대해서는 그런 판정이 없다.** 이 상태로 run-phase에 들어가면 구현자는
빨간 테스트 하나를 만나 「새 동작에 맞춰」 고치게 되고, 그것이 곧 조용한 철회다.

**놓친 기제는 측정 가능하다 (D20과 같은 뿌리).** `plan.md:80`이 열거 방법으로
`/usr/bin/grep -rn 'ResolveTodoQueueRoot' internal/ --include='*_test.go'`를 든다. 이 패턴은
**대문자 `R`로 시작하므로 CLI의 소문자 래퍼 `resolveTodoQueueRoot()`(`internal/cli/todo.go:71` —
그 본문이 `kanban.ResolveTodoQueueRootAdopting`을 부른다)를 부르는 테스트를 볼 수 없다.** 위 두 건이
정확히 그 경로다. 같은 이유로 `internal/web`의 간접 경로(`todoBodyFor` → `readTodoQueue`)도 그
grep에 잡히지 않는다 — 실측으로 확인했다: `internal/web`에서 그 패턴이 잡는 파일은
`todo_queue_read_test.go`(심볼 목록 테스트) 하나뿐이고, §C.1 C행이 든
`todo_section_test.go:194`는 **그 grep의 산출물이 아니다.**

— **Required fix**: 열거를 두 축으로 다시 수행해 §C.1에 등재한다. ⑴ 대소문자 무관 심볼 검색
(`resolveTodoQueueRoot` 포함) ⑵ 간접 소비자 경유(`todoBodyFor` / `readTodoQueue` / `newTodoStore` /
`newTodoCmd`). 위 2건은 A행(확실히 깨짐)에 넣고, **AC-SA-011에 대해 `AC-WTQ-008`과 같은 급의
교차-SPEC 판정을 `spec.md §8`에 남긴다** — 철회인지 전제 정합인지. (판정 재료 하나를 남긴다:
`REQ-SA-011` 본문은 「오염을 만들지 못한 뮤턴트는 그 사실과 가드 경계를 **보고하라**」고 이미
적는다 — 즉 그 요구사항 자신이 비오염 뮤턴트를 *발견*으로 다루므로, 비임시 비git base로 뮤턴트를
재기초하는 「보존 이관」이 정합해 보인다. 그러나 이것은 리드·운영자의 판정 사항이지 감사자의
결정이 아니다.) 덧붙여 `TestAxisACanaryHomeSweep_TodoFamily`(`todo_axisa_guard_test.go:69`)는
깨지지는 않으나 **0-오염이 가드에 의해 구조적으로 보장되면서 `runTodo` 게이트의 증거이기를 그친다** —
C행(공허화) 후보로 함께 등재한다.

### D19 (blocking, major) — C행(공허화 5건)의 새 판정식이 **5건 중 4건에서 성립하지 않는다**

— `plan.md:142` — Severity: **major** — Class: **blocking**

카드가 특정해 물은 항목이다: 「공허하게 구성된 종료 조건의 대체물이니, 새로운 방식으로 공허해서는
안 된다.」 **새로운 방식으로 공허하다.**

기재된 최소 형태: *「사본의 단언 대상이 홈 루트(또는 adopt 목적지)인 것 — 그 단언은 가드가 발화하면
성립할 수 없으므로, PASS 자체가 그 가지를 밟았다는 증거가 된다.」* 5건 각각에 대해 그 형태가
가능한지 본문을 직접 판독했다.

| C행 테스트 | 그 테스트의 단언 대상 (실측) | 최소 형태가 성립하는가 |
|---|---|---|
| `PureFallbackWritesNothing` (`todo_root_test.go:125`) | 로컬 파일 mtime 불변 + **`os.Stat(fallbackRoot)`가 `IsNotExist`** (`:147-149`) | **아니다.** 홈 루트가 등장하지만 **부재 단언**이다 — 가드가 발화해도 홈 루트는 생기지 않으므로 그대로 참이다. 기재된 근거("가드가 발화하면 성립할 수 없다")가 이 건에 대해 거짓 |
| `ReadThroughToProjectLocal` (`:156`) | `got != dir → Fatal`, 즉 **`got == dir`** (`:164-166`) | **아니다.** read-through 가지는 정의상 project-local 루트를 돌려주므로 홈 루트를 단언할 수 없고, 단언 대상 `dir`이 **가드의 반환값과 같다** — PASS가 두 상태를 구별하지 못한다 |
| `AdoptionLandsWhereConsumersRead` (`todo_root_contract_test.go:14`) | `os.Stat(BacklogPathForRoot(root))` 성공 (`:23`) | **그렇다.** 사본이 `root == 홈 루트`를 더할 수 있고 그 단언은 가드 아래 성립 불가다 — **5건 중 유일** |
| `AdoptingAndPureResolversAgreeWhenAdoptionFails` (`:40`) | `adopting == pure` + stat 성공 (`:61-67`) | **아니다.** 이 픽스처는 adopt를 **고의로 실패**시키므로 두 리졸버 모두 read-through로 `proj`를 돌려준다 — adopt 목적지도, 홈 루트도 단언 대상이 될 수 없고, `proj`는 다시 가드의 반환값과 같다 |
| `TodoSectionReadsThroughToProjectLocalQueue` (`internal/web/todo_section_test.go:194`) | 렌더 3행 + mtime 불변 + **`fallbackRoot` 부재** (`:207-220`) | **아니다.** 1행과 2행의 결함을 합쳐 갖는다 — 부재 단언 + 콘솔은 애초에 루트를 반환하지 않는다 |

즉 **4/5에서 「PASS 자체가 증거」가 성립하지 않는다.** 그 4건의 사본은 이음매 스텁이 실패해 가드가
발화해도 그대로 초록이다 — 3차가 폐기시킨 「미갱신 테스트 무실패」와 **판별력이 같다.** 이 SPEC이
이미 세 번 대가를 치른 부류(D2·D10·D14)의 네 번째가 예약된 자리이고, 그중
하나(`AdoptionLandsWhereConsumersRead`)는 「카드가 조용히 사라진」 사고의 회귀 가드다.

— **Required fix**: 판정식의 관측 대상을 **반환값에서 판별식 자체로** 옮긴다. 각 비임시 사본이 자기
base에 대해 **판별식이 「임시 아님」을 보고한다는 것을 직접 단언**하도록 한다(M1이 만드는
`TempOriginReason(base)`의 `isTemp == false`) — 이것은 긍정 단언이고, 이음매 스텁이 듣지 않으면
그 자리에서 RED가 된다. 원 단언은 그대로 둔다. 그러면 판정식이 5건 전부에 대해 성립한다:
「사본이 PASS **하고** 그 사본 안의 판별식 단언이 비임시를 보고했다」. 3행(`AdoptionLandsWhereConsumersRead`)만
추가로 홈 루트 동일성을 더할 수 있으므로 그 건은 두 겹이 된다.

### D20 (optional, minor) — 「전수」의 귀속이 인용된 방법과 맞지 않는다

— `plan.md:80` — Severity: minor — Class: **optional**
인용된 grep은 대소문자 구분이라 CLI 래퍼 경로를 원리상 볼 수 없고(D18), 동시에 그 grep의 산출물이
아닌 항목(`todo_section_test.go:194`)이 표에 실려 있다. 즉 **결과가 방법보다 넓고 동시에 좁다** —
어느 쪽이든 「전수 8건」은 인용된 방법으로 재유도되지 않는다. 실질 피해는 D18에서 이미 blocking으로
계상했으므로 여기서는 귀속 문제만 optional로 남긴다.
— Required fix(선택): 두 축(심볼 대소문자 무관 · 간접 소비자 경유)을 방법으로 적고, 각 행이 어느
축에서 나왔는지 표시한다. 제외한 후보(예: `internal/web`의 렌더 계열 `todo_section_test.go:81`·`:133`·`:155` —
비git 임시 루트로 리졸버를 지나지만 단언 대상이 렌더 결과라 의미가 변하지 않는다)는 **제외 기준과
함께** 적는다.

---

## 4. `AC-THG-003` 픽스처 (범위 밖 1건, 카드가 판정을 요청) — **정확한 자기 신고이고, 다시 만족 가능하다**

**⑴ 종전 Given이 실제로 성립 불가였는가 — 그렇다.** canary HOME은 `stubHome`가 만드는 `t.TempDir()`
(`internal/kanban/todo_root_test.go:55-61` 실측)이므로 그 아래 디렉터는 임시 루트 아래다. 「비임시
base」를 그렇게 구성하면 가드가 발화해 Given이 자기 모순이 된다. **3차 감사는 이것을 잡지 못했다** —
AC-THG-003을 blocking으로 등급만 매기고 픽스처의 성립 가능성을 재지 않았다. 내 누락이다.

**⑵ 「D15의 파급 반경 안, 명명된 세 결함 밖」이라는 성격 규정이 옳은가 — 옳다.** 이 불가능성은
「가드 + [HARD] `t.TempDir()` 격리 규율」의 곱에서 나오고, 그 곱을 처음 계산한 것이 D15의 이음매
논의다. D14(문장 재작성)·D16(교차 결정)과는 접점이 없다.

**⑶ 다시 만족 가능한가 — 그렇다.** `acceptance.md:74`의 새 픽스처는 base를 `t.TempDir()`로 두어
격리 규율을 지키고, 그 실행 동안 **임시 루트 집합을 base를 포함하지 않는 값으로 스텁**해 「비임시」를
선언한다. 이 형태에서 Given의 「임시 루트 밖」은 *주입된 집합 기준*으로 읽히며 자기 모순이 없다.
선례(`HomeDirFn` + `stubHome`)와 같은 급이고, `plan.md:126`(M1 이음매)과 `plan.md:130`(M1 종료
조건에 스텁 출력 요구)이 그 수단을 M2 이전에 세운다 — **순서도 맞다.**
잔여 위험(결함 아님, 기록만): 이 AC는 판별식을 *주입된 집합*에 대해 검증하며, 생산 집합
(`os.TempDir()`, `/tmp`, `/var/folders`)에 대한 「물리적으로 비임시인 디렉터」는 여전히 검증하지
않는다. 이음매를 쓰는 어떤 설계에서도 불가피하고, AC-THG-002/006이 생산 집합 쪽을 겨눈다.

**의존 (unfold 이후 확인 필요)**: 이 픽스처는 이음매 요구사항에 **의존**한다. 이음매가
`REQ-THG-009`로 분리되면 `acceptance.md:17` 매트릭스 행이 `REQ-THG-005` **및 `REQ-THG-009`**를
가리켜야 한다. 지금은 아래 D21이다.

---

## 5. 카드가 제시한 교차-SPEC 판정의 검증 (`AC-WTQ-008`) — **동의한다**

리드의 판정을 그대로 채택하지 않고 `SPEC-WEB-TODO-QUEUE-001` 산출물에서 직접 재측정했다.
- `acceptance.md:129-135`의 AC-WTQ-008이 전제를 `:105-111`의 AC-WTQ-006에서 물려받고, 그 Given이
  "a working directory that **git cannot resolve to a primary checkout**"이다 — 「임시」는 없다. **일치.**
- `spec.md:125-130` REQ-WTQ-004는 resolve와 adopt의 **분리**를 요구하며 임시성이 축에 없다. **일치.**
- `todo_root_test.go:224`의 `t.TempDir()`는 비git 디렉터를 얻는 관용 수단이다. **일치.**

따라서 t536이 하는 일은 홈 폴백에 도달하는 비git base의 **모집단을 좁히는 것**이고, AC-WTQ-008의
보장은 비임시 비git base에 대해 그대로 참이다 — 철회가 아니라 전제 정합이라는 판정에 **동의한다.**
`SPEC-WEB-TODO-QUEUE-001` 산출물이 수정되지 않았음도 확인했다.
다만 **같은 판정이 AC-SA-011에는 없다**(D18) — 판정의 논리가 옳다는 것과 그 논리가 필요한 곳에
전부 적용됐다는 것은 다른 주장이다.

---

## 6. 이번 회차 결함 목록 (구조화)

D18. §C.1 열거 불완전 — `plan.md:53-60`·`plan.md:80` — 확실히 깨지는 테스트가 최소 2건 누락
(`internal/cli/todo_queue_root_test.go:153` adopt-not-shadow, `internal/cli/todo_axisa_guard_test.go:150`
= 완료된 `SPEC-STATE-ANCHOR-001`의 **AC-SA-011 산출 테스트**이며 오염 **발생**을 단언한다).
누락 기제는 인용된 grep이 대소문자 구분이라 CLI 소문자 래퍼(`internal/cli/todo.go:71`)를 못 본 것 —
Severity: major — Class: **blocking** — Required fix: 두 축으로 재열거해 A행에 등재하고,
AC-SA-011에 대해 `AC-WTQ-008`과 같은 급의 교차-SPEC 판정을 `spec.md §8`에 남긴다.
`todo_axisa_guard_test.go:69`는 C행 후보로 추가.

D19. C행 판정식이 5건 중 4건에 공허 — `plan.md:142` — 「사본의 단언 대상이 홈 루트」라는 최소 형태가
`PureFallbackWritesNothing`·`TodoSectionReadsThroughToProjectLocalQueue`(홈 루트가 **부재** 단언),
`ReadThroughToProjectLocal`·`AdoptingAndPureResolversAgreeWhenAdoptionFails`(단언 대상이 가드의
반환값과 동일)에서 성립하지 않는다 — Severity: major — Class: **blocking** — Required fix: 판정식을
반환값이 아니라 **판별식의 직접 단언**(`isTemp == false`)에 걸고 원 단언을 병기한다.

D20. 「전수」의 방법 귀속 불일치 — `plan.md:80` — 결과가 인용된 방법으로 재유도되지 않는다 —
Severity: minor — Class: **optional** — Required fix: 두 축을 방법으로 적고 제외 기준을 남긴다.

D21. 이음매 절의 산출 AC가 매트릭스에 없다 — `acceptance.md:16-17`·`spec.md:117` — `spec.md:171`이
이음매의 산출 AC를 AC-THG-003이라 적지만, 매트릭스의 AC-THG-003 행은 `REQ-THG-005`만 가리킨다.
현재 판본에서 REQ-THG-002를 겨눈 AC(002·006)는 어느 쪽도 이음매를 밟지 않는다 — Severity: minor —
Class: **optional** — **unfold 의존**: `REQ-THG-009` 분리 시 AC-THG-003 행을
`REQ-THG-005, REQ-THG-009`로 고치면 닫힌다. (리드에게: 이 항목은 unfold 델타에서 함께 확인할 것.)

D22. 교차(임시 ∧ 홈 해석 불가) 절에 산출 AC가 없다 — `spec.md:116`·`acceptance.md:34` — REQ-THG-001의
새 절을 지키는 것이 AC가 아니라 DoD 한 줄(`acceptance.md:183`)이다 — Severity: minor —
Class: **optional** — Required fix(선택): AC-THG-001에 갈래 (c)를 **And 한 줄**로 더한다
(「canary HOME 대신 `HomeDirFn` 오류를 주입해도 반환 루트는 `base`다」). AC 개수는 늘지 않는다.

D17(3차, optional)은 **여전히 열려 있다** — 이번 범위 밖이라 재판정하지 않았고 등급도 그대로다.

---

## 7. 회귀 점검 (3차 결함 중 이번 범위)

- **D14 — RESOLVED.** §1. 기재가 참이고 분리가 실질이다.
- **D16 — RESOLVED.** §2. 값 고정이 배치 의존을 제거한다.
- **D15 — UNRESOLVED.** 수리는 실질적 진전이다(1건 → 8건 열거, 판정식 3분할, `AC-WTQ-008` 교차 판정).
  그러나 3차가 요구한 두 가지 **모두**가 아직 성립하지 않는다: 열거가 전수가 아니고(D18), 공허화를
  볼 수 있게 하려던 판정식이 4/5에서 공허하다(D19). **정체(stagnation)는 아니다** — 두 결함 다
  0.1.3이 새로 연 표면에서 나왔고, 3차가 지적한 문장은 실제로 고쳐졌다.

## 8. 측정 기록 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

**Claim A** — 확실히 깨지는 테스트가 §C.1의 4건 외에 최소 2건 더 있다.
**Evidence** — `/usr/bin/grep`으로 `internal/cli`의 `resolveTodoQueueRoot()` · `liveTodoQueueRootReason`
호출자 집합을 뽑고 두 테스트 본문을 직접 판독: `todo_queue_root_test.go:153-201`(홈 루트 동일성
단언 `want := filepath.Join(home, ".moai","todo", ...)`), `todo_axisa_guard_test.go:150-179`(오염 부재
시 `t.Fatalf`). 래퍼 좌표 `internal/cli/todo.go:71-73`이 `kanban.ResolveTodoQueueRootAdopting`을 부른다.
`SPEC-STATE-ANCHOR-001` 소속 확인: 그 SPEC의 `acceptance.md:23`·`:152`, `spec.md:106`, `status: completed`.
**Baseline-attribution** — this run, this tree, HEAD `412c8cb14`.
**Gaps** — **실행으로 확인하지 않았다** — 가드가 없으므로 원리상 불가능하다. 판독 유도다.
**Residual-risk** — 구현이 술어를 더 좁은 지점에 넣으면 파급이 줄 수 있다. 그 경우에도 열거가
불완전하다는 결함 자체는 남는다.

**Claim B** — C행 판정식이 5건 중 4건에서 성립하지 않는다.
**Evidence** — 5개 테스트 본문 전수 판독(§3의 표에 좌표와 단언문 기재). 부재 단언 2건은
`os.Stat(fallbackRoot)`의 `IsNotExist` 형태를 직접 확인했다.
**Baseline-attribution** — this run, this tree, HEAD `412c8cb14`.
**Gaps** — 사본이 아직 존재하지 않으므로 사본을 실행해 반증하지는 못했다. 판정은 「사본이 그 최소
형태를 취할 수 있는가」에 대한 것이고, 그것은 원본의 단언 구조와 가드의 반환값에서 결정된다.
**Residual-risk** — run-phase가 최소 형태를 넘어 더 강한 사본을 쓰면 실제 판별력은 회복될 수 있다.
그러나 **기재된 판정식이 그것을 요구하지 않는다** — 그것이 결함이다.

**Claim C** — `AC-THG-003`의 새 픽스처는 만족 가능하다.
**Evidence** — `stubHome`(`todo_root_test.go:55-61`)가 `t.TempDir()` 홈을 만든다는 실측 + `acceptance.md:74`의
픽스처 서술 + `plan.md:126`·`:130`(M1이 이음매를 M2 이전에 세운다).
**Baseline-attribution** — this run, this tree.
**Gaps** — 이음매가 아직 없으므로 픽스처를 세워보지 못했다.
**Residual-risk** — 주입 집합 기준의 검증이라 생산 집합에 대한 물리적 비임시 케이스는 미검증(§4).

**Gaps (이번 회차 전체)** — ⑴ **범위 제한으로 재측정하지 않은 축**: MP-1~MP-7, 카테고리 점수,
1-3차에서 닫힌 축 전부. 이번 회차는 **종합 점수를 산출하지 않는다** — 부분 감사에서 산출한 수치는
전수 감사 수치와 비교 가능하지 않아 STOP 조항의 비교 대상이 될 수 없다. ⑵ `plan.md:38`의 344는
이번에도 재측정하지 않았다(워크트리 격리; `plan.md:45`가 이미 Gap으로 등재하고 run-phase 재측정을
설계했다 — fail-safe 방향). ⑶ 리눅스 셀 미측정(SPEC이 스스로 Gap 등재, CI 소관). ⑷ 판별식이
없으므로 어떤 실동작도 실행 검증하지 못했다 — run-phase 소관이며 AC가 겨눈다.

---

## 9. Recommendation

**운영자의 implementation-kickoff 게이트에는 아직 올릴 수 없다.** 막는 것은 **D15 축 하나**이고,
그 안에 둘이다. D14·D16은 닫혔다.

1. **§C.1 재열거 + AC-SA-011 교차 판정** (D18) — 대소문자 무관 심볼 축과 간접 소비자 축으로 다시
   훑어 `todo_queue_root_test.go:153`과 `todo_axisa_guard_test.go:150`을 A행에 넣는다. 후자는
   `SPEC-STATE-ANCHOR-001` AC-SA-011의 산출 테스트이므로 `AC-WTQ-008`과 같은 급의 판정을
   `spec.md §8`에 남긴다 — **이것이 이번 회차에서 가장 무거운 항목**이다. 판정 없이 run-phase에
   들어가면 조용한 철회가 그 자리에서 일어난다.
2. **C행 판정식 교체** (D19) — 관측 대상을 반환값에서 판별식(`isTemp == false`)의 직접 단언으로
   옮기고 원 단언을 병기한다. 한 문장 교체이며 새 측정을 요구하지 않는다.

둘 다 국소 편집이고 설계·범위를 바꾸지 않는다. D20·D21·D22는 재량이며 D21은 **unfold 델타에서
함께 처리**하면 된다.

**닫히면 PASS로 본다.** 그때 run-phase에서 계속 지켜볼 것 넷: ⑴ 뮤턴트가 실제로 무엇을 잡았는지의
출력 전문(같은 자리에서 세 번 공허했다), ⑵ 보존 이관된 테스트의 원 단언이 **지워지지 않았는지**,
⑶ C행 사본이 판별식 단언을 실제로 들고 있는지, ⑷ 리눅스 셀 판정을 macOS 통과로 대체하지 않는지.

**Tier**: 이번 회차는 판정하지 않는다(감사 중 리드가 결정 — fold 기각, `REQ-THG-009` 분리, Tier M).
Tier M의 plan-audit 상한 2는 **이후 회차에만** 적용되며 운영자가 이전 tier 아래 개설한 이번 회차에는
소급하지 않는다. 위 D21은 unfold 델타와 함께 확인하면 닫힌다.
