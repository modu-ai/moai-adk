# SPEC Review Report: SPEC-CODEX-SKILLS-CANONICAL-001

Iteration: 3/3 (Tier M 2회 상한에 대한 리드 명시적 예외)
Verdict: **FAIL**
Overall Score: **0.7875** (산술평균) / **0.775** (조화평균) — Tier M 임계 **0.80** 미달. 점수와 별개로 아래 blocking 3건이 독립적으로 FAIL 을 강제한다.

Reasoning context ignored per M1 Context Isolation. 리드 메시지의 전제(무엇이 닫혔다는 주장)와 선행 감사 3건의 결론은 **검증 대상**으로만 취급했고, 판정 근거로 인용하지 않았다.

## §0. 감사 대상 고정 확인

착수 시점에 직접 실행했다.

```
$ git log --oneline -1
2145e2b2a docs(spec): add SPEC-CODEX-SKILLS-CANONICAL-001 plan-phase artifacts (card t81)
$ git status --short
(출력 없음)
$ git branch --show-current
WT-skills-canonical
$ git rev-parse HEAD
2145e2b2ab88220fb27757acc9f6a89a95708491
```

**감사 종료 시점에도 동일함을 재확인했다** (probe 디렉터리 생성·삭제 후 `git status --short` → 무출력). 이 감사는 단일 스냅샷 `2145e2b2a` 에 대해 수행됐고, 도중 아티팩트가 움직이지 않았다.

## §1. Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — `grep -o 'REQ-CSC-[0-9]\{3\}' spec.md | sort -u` → `REQ-CSC-001 … REQ-CSC-016` 연속 16개. 결번·중복 0, zero-padding 3자리 일관.
- **[PASS] MP-2 GEARS 형식 준수 (요구사항 계층 판정)** — `spec.md:210-225` 의 REQ-CSC-001~016 을 항목별로 읽었다. Event-driven(`~할 때(When)`: 003·005·008·011·012·013·014), State-driven(`~하는 동안(While)`: 009), Where(`~환경에서(Where)`: 004), Ubiquitous shall(001·006·015), Unwanted shall not(002·007·010·016). 비정형 문장·REQ 자리의 Given-When-Then 0건. **판정 계층(AC-CSC-*)의 Given-When-Then 은 검증 계층 정본 형식이므로 이 기준으로 감점하지 않았고, Group 4 에서 별도 채점했다.**
- **[PASS] MP-3 YAML frontmatter** — `spec.md:1-15`. 12개 필드(`id`/`title`/`version`/`status`/`created`/`updated`/`author`/`priority`/`phase`/`module`/`lifecycle`/`tags`) 전부 존재하고 타입 일치. `version: "0.4.0"` quoted semver, `created`·`updated` ISO date, `priority: P2` enum, `tags` 콤마 구분 문자열. 거부 alias(`created_at`/`updated_at`/`labels`/`spec_id`) 0건. `tier: M` 은 추가 필드로 허용.
- **[N/A] MP-4 언어 중립성** — 본 SPEC 의 대상은 moai-adk 자체의 Go 코드(`internal/template`, `internal/cli/update/deploy`)이며 사용자 프로젝트 언어 매트릭스를 다루지 않는다. 단일 언어 스코프이므로 N/A(자동 통과). §E 의 "템플릿 중립성" 제약은 별개 축이며 AC-CSC-015(4)가 기존 게이트로 판정한다.
- **[PASS] MP-5 D7 교차 SPEC 정합** — `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` 를 3개 아티팩트에 실행한 결과 자기 ID 하나뿐이다. 외부 SPEC 참조 0건이므로 retired/superseded 참조도 0건. BLOCKING 없음.
- **[PASS] MP-6 D8 크로스 플랫폼 규율** — `grep -c 'syscall'` → spec 0 / plan 0 / acceptance 0. `syscall` 미등장이므로 자동 PASS. (Windows 축은 `os.Symlink` 권한 문제로 다루며, REQ-CSC-004 폴백 + `GOOS=windows go vet` 게이트로 처리 — 빌드 태그 규율 대상이 아니다.)
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/` → rc=1, 매치 0건. (`research.md` 는 Tier M 이라 부재. `plan.md` 는 존재하며 마커 없음.)

**must-pass 는 7개 전부 통과다.** 아래 FAIL 은 must-pass 위반이 아니라 판정 계층(Group 4)의 blocking 결함과 점수 미달에서 나온다.

## §2. Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | 서술은 정밀하고 모호한 대명사 없음. 감점 사유는 **load-bearing 인용 4건이 엉뚱한 절을 가리킨다**(D1) + `§D.4` 문서 간 참조가 spec.md 안에서 해석되지 않음(D4) + §A.11 이 코드 주석에 없는 문장을 주석 인용으로 제시(D5). |
| Completeness | 0.90 | 0.75–1.0 | 필수 절 전부 존재: HISTORY(`spec.md:19`), 맥락/근거 §A(`:38`), 설계 §B(`:184`), 요구사항 §C(`:208`), 범위 밖 §D(`:231`, `### Out of Scope —` H3 4개 각각 구체 불릿 보유), 판정 acceptance.md 전체, frontmatter 12/12. 감점은 §A 가 "실측" 절인데 `io.Writer` 부재 실측(REQ-CSC-005 의 유일 근거)만 §A 에 없고 plan/acceptance 에만 있다는 비대칭. |
| Testability | 0.65 | 0.50–0.75 | 16개 AC 중 **MUST 2개가 작성된 그대로는 통과 불가**(D2: AC-CSC-008, D3: AC-CSC-012). 나머지 14개는 명령·경로·개수로 기계 판정 가능하고 weasel word 0건. AC-CSC-009 는 판정력이 얇음을 스스로 명시(감점 아님 — 알고 두는 것과 모르고 두는 것은 다르다). |
| Traceability | 0.85 | 0.75–1.0 | §D.3 매트릭스 실측 대조: REQ 16개 → 각각 ≥1 AC, AC 16개 → 각각 유효 REQ. 고아 AC 0, 미커버 REQ 0, 은퇴 흔적 0. 감점은 plan §F M3 닫힘 조건이 `AC-CSC-012 (양팔)` 로 **팔 개수를 잘못 세어** 3번 팔이 어느 마일스톤 소유인지 문서상 미확정(D3-b). |

산술평균 `(0.75+0.90+0.65+0.85)/4 = 0.7875`. 조화평균 `0.775`. 어느 방식이든 0.80 미만.

## §3. 선행 지적 대조 (닫힘/미해결)

보고서 약칭: **a#1** = `plan-audit.md`(iter-2, 0.78), **a#2** = `plan-audit-iter2.md`(iter-2, 0.7625), **a#0** = `plan-audit-iter1.md`(iter-1, 0.775).

### 3.1 닫힌 것 — 전부 아티팩트에서 직접 확인

| ID | 제기 | 결함 | 상태 | 근거 |
|---|---|---|---|---|
| a#1 N1 / a#2 N4 | 양쪽 | dangling 링크로 청소가 미러를 하나도 못 지움 | **CLOSED** | REQ-CSC-008(`spec.md:222`)이 세 절을 함께 건다: 글롭 등록 + **`os.Lstat` 기준 판정 + dangling 제거**(shall) + **순서 배치**(shall). §B.D6(`:197-203`)이 (a) 본체 / (b) 이중 방어로 하중을 명시하고 `[HARD] (b) 를 중복이라며 지우지 않는다`로 못박음. |
| a#1 N2 / a#2 N5 | 양쪽 | fixture 가 실 디렉터리라 위 결함 미검출 | **CLOSED (단, D2 신규 결함 동반)** | AC-CSC-008(`acceptance.md:92-107`) fixture 4형태 — `moai-live`(살아있는 링크) / `moai-gone`(**dangling**) / `moai-copied`(실 디렉터리) / `hns-user-owned`. 단언 2번이 dangling 팔. `[HARD] 2번이 이 AC 의 존재 이유다` 명시. |
| a#1 N3 / a#2 N6 | 양쪽 | 존재하지 않는 "Deploy 출력 버퍼" 전제 | **CLOSED (단, D1 인용 오류 동반)** | REQ-CSC-005 가 반환값 seam 으로 확정(`spec.md:218`), AC-CSC-006 `[HARD]` 가 버퍼 캡처 형태 금지(`acceptance.md:79`), plan M2 가 seam 을 산출물로 소유(`plan.md:68`). |
| a#1 N4 / a#2 N3 | 양쪽 | §H 판별 기준이 현역 스킬 `moai` 를 삭제 후보로 분류 | **CLOSED** | plan §H(`:145-146`)가 `[HARD] 판별자는 위 하나뿐이다` 로 템플릿 디렉터리 이름 집합 단일 기준을 세우고 `EmbeddedMoaiSkillNames()` 사용을 금지. AP-15 도 동일 금지. |
| a#1 N10 | a#1 | plan §F M3 이 REQ-CSC-010 과 모순 | **CLOSED** | M3(`plan.md:74-83`) 제목·본문이 "기록·백업 대상에서 **제외**"로 뒤집혀 있음. |
| a#1 N11 | a#1 | plan §F M5 가 REQ-CSC-016 미추종 | **CLOSED** | M5(`plan.md:102-110`) Priority High, `.agents/skills/moai*` 명시, `.agents/` 전체 금지 `[HARD]`. |
| a#1 N12 | a#1 | 백업 제외 로직 무소유 | **CLOSED (부분 — D3-b 참조)** | M3 2번 항목이 `backupThenRemove` 처리를 명시적으로 소유. |
| a#2 N1 | a#2 | REQ-CSC-001 무조건 shall 이 REQ-011·014 와 모순 | **CLOSED** | `spec.md:210` 에 예외 절 추가("단 REQ-CSC-014 … REQ-CSC-011 … 은 예외이며, 그 두 경우에는 경고가 접근 경로의 자리를 대신한다"). AC-CSC-002 도 단서 문단으로 동기화(`acceptance.md:60`). |
| a#2 N2 | a#2 | 무백업 데이터 손실 경로 | **CLOSED (단, D3 신규 결함 동반)** | REQ-CSC-010(`spec.md:220`)이 백업 금지를 "이번 실행이 곧바로 다시 만들 미러"로 한정하고, 걸리지 않는 항목은 기존 백업 규칙을 따르라고 shall. AC-CSC-012(3)가 판정. |
| a#2 N7 | a#2 | `.gitignore` 가 `.agents/` 전체를 무시 | **CLOSED** | REQ-CSC-016(`spec.md:226`) `.agents/skills/moai*` 한정 + `.agents/` 전체 금지 shall not. AC-CSC-015(2)가 금지형을 판정. §B.D7 이 사유 보유. |
| a#2 N8 | a#2 | AC-CSC-010 이 REQ-CSC-007 을 증명하지 못함 | **CLOSED (범위 축소 + 갭 공시)** | REQ-CSC-007 이 "**미러 기능의 활성 여부가** … 변화시켜서는 안 된다"로 좁혀졌고, 잃은 축을 §D.4 1회 대조가 덮으며 §D.2 `[HARD]` 가 "그 대조는 AC 가 아니므로 회귀 가드가 아니다"를 명시. 갭을 숨기지 않고 적은 형태 — 수용. |
| a#2 N9 | a#2 | §G 가 §F 앞에 옴 | **CLOSED** | 현재 순서 §F(`:250`) → §G(`:256`). |
| a#1 N5 | a#1 | Codex 노출 제외 사유가 t91 방법론과 모순 | **CLOSED** | §D.4(`acceptance.md:222`)가 사유를 "codex 바이너리 부재 + 버전 의존"으로 정정하고 이전 사유가 틀렸음을 명기. |
| a#1 N6 | a#1 | §A.2 embed 서술에 조건 미명시 | **CLOSED** | §A.2(`spec.md:52`)에 "**조건 명시**" 문단 추가 — 디렉터리 패턴 임베드 한정, 직접 지목 시 빌드 오류. |
| a#1 N8 | a#1 | Codex 리스팅 예산 미확인 | **CLOSED (범위 밖 명시)** | §A.3 말미가 예산 축을 적고 "본 SPEC 의 요구사항에는 영향이 없고, 예산 축의 실측은 범위 밖"으로 경계를 그음. |
| a#1 N9 | a#1 | §F 교차 참조 1건이 worktree 에서 미해석 | **CLOSED — 나도 재측정함** | §F(`:254`)가 "primary 체크아웃 기준 경로"임을 명기. 실측: worktree 에 부재(`ls` rc=1), primary 에 존재(`-rw-r--r-- 7425 Aug 17 01:11`). 주석이 정확하다. |

**미해결로 넘어온 blocking 은 0건이다.** 정체(stagnation) 신호 없음 — 같은 결함이 3회 반복된 항목은 하나도 없다.

### 3.2 은퇴 항목 점검 (리드 지시 5번)

HISTORY 는 "예산: REQ 16 / AC 16 불변 — 은퇴시킨 항목 없음"이라고 적는다. **직접 세어 확인했다**: REQ 16개(001–016 연속), `^### AC-CSC-` 헤딩 16개(001–016 연속). §D.3 추적 매핑에 고아 AC 0 / 미커버 REQ 0. iter-3(v0.3.0)에서 이미 12→16 / 15→16 으로 상한에 도달해 있었으므로 v0.4.0 에서 은퇴가 **필요하지 않았고**, 실제로 없었다. **커버리지가 조용히 떨어진 흔적은 없다.** 이 주장은 참이다.

## §4. Defects Found (구조화 목록)

### D1 — 출력 seam 의 근거 인용 4건이 전부 엉뚱한 절을 가리킨다 (blocking)

`spec.md:218` (REQ-CSC-005) · `acceptance.md:79` (AC-CSC-006 [HARD]) · `plan.md:29` (R15) · `plan.md:68` (M2 [HARD]) — Severity: **major** — Class: **blocking**

네 곳 모두 반환값 출력 seam 의 근거로 `§B.D6` 을 든다. 그런데 `§B.D6`(`spec.md:195-203`)의 제목과 내용은 **"dangling 링크 제거: `Lstat` 판정이 본체, 순서 배치는 이중 방어"** 이며, 출력·관측 가능성에 대한 문장이 **한 줄도 없다**. 관측 가능성을 다루는 절은 `§B.D3`("폴백은 복사, 그리고 관측 가능해야 한다", `spec.md:192`)이고, `io.Writer` 부재 실측은 `§A` 에 아예 없이 `plan.md:68` 과 `acceptance.md:79` 에만 있다.

같은 `§B.D6` 인용이 dangling/순서/폭발 반경 맥락에서는 **정확하게** 쓰인다(`spec.md:155`, `acceptance.md:88`·`237`·`247`, `plan.md:30`·`94`·`96`). 즉 한 절 번호가 두 주제에 걸려 있고 한쪽이 틀린 형태다 — v0.4.0 이 §B 에 D6 을 새로 끼워 넣으면서 이전 판본의 절 번호를 그대로 옮겨 적은 잔재로 보인다.

run-phase 실무자가 "근거는 §B.D6"을 따라가면 Lstat 이야기만 읽고 왜 반환값이어야 하는지는 못 찾는다. 근거 사슬이 끊겨 있다.

Required fix: 네 곳의 `§B.D6` 을 출력 seam 의 실제 근거로 교체한다 — `§B.D3` + `io.Writer` 부재 실측. 그리고 그 실측을 `§A`(검증된 기준선)에 절로 승격해 근거가 spec 안에서 닫히게 한다. §A 승격이 예산에 걸리지 않는 이유는 §A 가 REQ/AC 개수 예산의 대상이 아니기 때문이다.

### D2 — AC-CSC-008 의 마지막 단언은 올바른 구현에서 **반드시 실패한다** (blocking)

`acceptance.md:107` — Severity: **critical** — Class: **blocking**

AC-CSC-008 은 fixture 를 심고 `CleanMoaiManagedPaths` 를 1회 실행한 뒤, 네 단언에 더해 이렇게 요구한다:

> 추가로, **정본 `.claude/skills/moai-live/SKILL.md` 가 링크를 통해 삭제되지 않았다**

실측: `CleanMoaiManagedPaths`(`internal/cli/update/deploy/deploy.go:101`)는 `targets := ManagedCleanTargets(projectRoot)` 를 받아 **7개(신규 등록 후 8개) 전부를 순회한다**. 그 목록의 4번째 항목이 `.claude/skills/moai*` 글롭이다(`ManagedCleanTargets` 본문 직접 확인). fixture 가 심은 정본 이름은 `moai-live` 이므로 이 글롭에 **매치되고**, `backupThenRemove` 가 `os.RemoveAll` 한다. 링크를 따라가서가 아니라 **정본 자신의 청소 대상에 걸려서** 지워진다.

따라서 청소가 완전히 올바르게 구현되어도 실행 후 `.claude/skills/moai-live/SKILL.md` 는 존재하지 않는다. 단언은 거짓이 되고, MUST AC 한 개가 **원리적으로 통과 불가**다. 순서 배치 (b) 를 적용하든 안 하든 결과는 같다 — (b) 는 미러가 먼저 처리되게 할 뿐 정본 글롭의 실행 자체를 막지 않는다.

이 문장의 출처는 a#2 의 Required fix(`plan-audit-iter2.md:199`)이며, 저자가 그대로 옮겼다. **a#2 도, a#1 도 이 단언이 실행 불가임을 지적하지 않았다.** 내가 새로 찾은 것이며, 근거는 위 코드 실측이다.

관측 가능성 문제이기도 하다: "링크를 통해 삭제됐는지"는 파일 존재 여부만으로는 **원리적으로 구분되지 않는다**. 두 경로 모두 결과가 "부재"다.

Required fix: 이 축을 관측 가능한 형태로 다시 쓴다. 두 방향 중 하나 — (i) 정본을 청소 글롭 밖 이름으로 심고(예: `.claude/skills/hns-live-src`, 링크 대상도 그리로) `moai*` 미러 링크만 청소가 지우는지 본다, 또는 (ii) 정본을 `moai-live` 로 유지하되 `.agents/skills/moai*` 단일 대상만 처리하는 좁은 호출 경로로 판정하고(전체 `CleanMoaiManagedPaths` 대신), 전체 실행 판정은 별도 단언으로 분리한다. 어느 쪽이든 "링크 추종 여부"가 파일 존재로 환원되지 않게 해야 한다.

### D3 — AC-CSC-012 는 2번 단언과 3번 단언이 **같은 fixture 에서 동시에 참일 수 없다** (blocking)

`acceptance.md:150-163` — Severity: **critical** — Class: **blocking**

AC 는 하나의 Given(1회 배포 + 1회 청소)에 대해 "두 단언이 모두 참이어야 한다"고 쓴 뒤 **세 개**를 나열한다. 그중:

- 2번: "`.moai-backups/**/pre-clean/.agents/` 아래 **파일 수가 0**이다."
- 3번: "`.moai-backups/**/pre-clean/.agents/skills/moai-retired/` 아래에 **보존되어 있어야** 한다."

3번이 요구하는 파일은 2번이 세는 서브트리 **안에** 있다. 2번의 한정어("템플릿이 같은 이름의 스킬을 가진 미러 항목에 대해")는 의도를 말하지만, 실제로 적힌 **측정 지표**는 `.agents/` 서브트리 전체 파일 수다. 같은 fixture 에서 3번을 만족시키면 그 수는 0 이 아니다. 테스터는 어느 쪽을 문자 그대로 재현할지 판단해야 하고, 그 순간 이 AC 는 기계적으로 참·거짓이 갈리지 않는다 — acceptance.md 서두가 스스로 요구한 성질을 잃는다.

이 결함이 위험한 이유는 완화 방향이다. 2번을 문자 그대로 구현하면 `.agents/` 를 백업에서 통째로 제외하게 되는데, 그것은 AC 자신이 바로 아래에서 금지한 형태다("2번만 있고 3번이 없으면 구현자는 `.agents/` 를 백업에서 통째로 제외해 AC 를 통과시키고, 그 순간 3번이 막으려는 손실이 발생한다"). **AC 가 자기 경고를 자기 문구로 유도한다.**

Required fix: 2번의 측정 범위를 판별자와 같은 폭으로 좁힌다 — "`.agents/skills/<템플릿에 대응 스킬이 있는 이름>/` 아래 파일 수가 0" 로 이름을 한정하고, 서브트리 전체 카운트를 쓰지 않는다. 동시에 헤딩 `(양팔)` 과 본문 "두 단언"을 3개로 정정한다(D3-c).

### D3-b — plan §F M3 이 AC-CSC-012(3) 이 금지하는 형태를 해법으로 처방하고, 닫힘 조건에서 팔 하나를 누락한다 (blocking)

`plan.md:81`, `plan.md:83` — Severity: **major** — Class: **blocking**

M3 본문은 이렇게 쓴다:

> 미러 뿌리를 **백업 대상에서 제외하는 명시적 분기**를 두거나 그 뿌리를 managed 로 간주하게 만든다 — 어느 쪽이든 AC-CSC-012 의 **2번 팔이** 판정한다.

앞의 선택지는 `.agents/` 뿌리를 통째로 백업에서 빼는 절대 형태이며, 이는 REQ-CSC-010 의 v0.4.0 한정(다시 만들지 않을 항목은 기존 백업 규칙을 따라야 한다)과 AC-CSC-012(3), 그리고 같은 문서의 **AP-16**("백업 금지를 `.agents/` 전체에 절대 형태로 적용한다. 사용자의 `moai` 접두 실 항목이 경고도 백업도 없이 사라진다")과 정면 충돌한다. plan 이 자기 안티패턴 목록이 금지한 것을 마일스톤 본문에서 처방한다.

닫힘 조건도 `AC-CSC-012 (양팔)` 로 적혀 3번 팔이 어느 마일스톤 소유인지 문서상 비어 있다. 같은 plan 의 위험 표는 3번 팔을 알고 있다(`plan.md:28` R14 → "AC-CSC-012(3번 팔)") — plan 내부에서도 어긋난다. v0.4.0 이 REQ-CSC-010 을 한정하고 AC 에 3번 팔을 더하면서 M3 만 iter-3 문구로 남은 잔재다.

Required fix: M3 본문의 "백업 대상에서 제외하는 명시적 분기"를 삭제하고, "**이번 실행이 다시 만들 미러인지**를 판별하는 분기"로 바꾼다(REQ-CSC-010 의 판별자 문구 그대로). 닫힘 조건을 `AC-CSC-012 (3팔)` 로 정정한다.

### D4 — `§D.4` 참조가 spec.md 안에서 해석되지 않는다 (optional)

`spec.md:216` (REQ-CSC-007) — Severity: minor — Class: **optional**

REQ-CSC-007 이 "§D.4 의 1회 대조"를 근거로 든다. 그런데 spec.md 자신의 `§D` 는 "범위 밖 (Exclusions)"이고 번호 하위절이 없다(`### Out of Scope — …` 4개). 실제 대상은 **acceptance.md 의 §D.4**다. 같은 참조가 acceptance.md 안(`:230`)에서는 정확히 해석된다 — 문제는 문서를 건너뛰는 쪽 하나뿐이다.

Required fix: `spec.md:216` 의 `§D.4` 를 `acceptance.md §D.4` 로 명시한다.

### D5 — §A.11 이 코드 주석에 없는 문장을 주석 인용으로 제시한다 (optional)

`spec.md:180` — Severity: minor — Class: **optional**

§A.11 은 이렇게 쓴다:

> 판별자는 이미 코드에 있다. `backupThenRemove` 는 템플릿이 같은 경로를 가진 파일을 백업하지 않는데, **그 사유가 주석에 적혀 있다** — *배포가 곧바로 다시 쓰므로 유일본이 위태로운 적이 없다*.

실측(`internal/cli/update/deploy/deploy.go:356-370`, doc comment 전문 확인): 주석은 **동작만** 적는다 — "A FILE target (settings.json) is backed up unless the template carries the exact path". 인용된 *사유* 문장은 주석에 없다. `templateCarries` 함수 주석(`:425`)도 "reports whether tmplFS holds a regular file at rel" 한 줄뿐이다.

판별 **기준**이 코드에 있다는 부분은 참이다(`templateCarries` 분기). 틀린 것은 그 기준의 근거가 주석에 문장으로 있다는 주장이다. §A 는 "검증된 기준선(실측)" 절이므로 이 자리의 부정확한 인용은 다른 §A 항목의 신뢰도에 얹힌다.

부가 관측 하나: `templateCarries` 판별자는 `!info.IsDir()` 분기, 즉 **파일 대상**에만 걸린다. 미러 항목은 디렉터리(복사 모드) 또는 링크이므로 실제로는 `templateManagedPaths` 경로를 타고, 그쪽은 템플릿에 `.agents/` 가 없어 항상 공집합이다(§A.7 이 이미 실측한 바). 즉 "그대로 쓴다"는 재사용이 아니라 **새 분기 작성**이다. plan M3 이 그 작업을 인정하므로 커버리지 갭은 아니지만, §A.11 의 "이미 코드에 있다" 프레이밍은 필요한 작업량을 낮춰 읽게 한다.

Required fix: 인용 형태를 버리고 "판별 **기준**(`templateCarries` 의 template-carries 비교)이 코드에 있으며, 디렉터리 경로에는 적용되지 않으므로 미러용 분기는 새로 작성한다"로 사실에 맞춘다.

### D6 — 단언 개수 표기가 세 곳에서 실제 개수와 어긋난다 (optional)

`acceptance.md:150`(헤딩 `(양팔)`) · `:153`("두 단언") · `:184`("세 단언"인데 4개 나열) — Severity: minor — Class: **optional**

- AC-CSC-012: 헤딩 "(양팔)" + 본문 "두 단언" ↔ 실제 3개 (본문 뒤쪽은 "세 단언의 관계가 이 AC 의 요점"이라고 옳게 씀 — 같은 절 안에서 2와 3이 공존한다).
- AC-CSC-015: "세 단언이 모두 참이어야 한다" ↔ 실제 4개(1·2·3·4).

셋 다 v0.4.0 이 팔을 늘리면서 개수 문구를 갱신하지 않은 편집 잔재다. 개별로는 사소하지만 **§D.6 Definition of Done 이 "MUST AC 15개 전부"를 개수로 판정**하는 문서에서 개수 표기가 흔들리는 것은 판정 기준 자체의 신뢰도 문제다. (MUST 15 / SHOULD 1 = 16 은 직접 세어 확인했고 그쪽은 정확하다.)

Required fix: `(양팔)`→`(3팔)`, "두 단언"→"세 단언", AC-CSC-015 "세 단언"→"네 단언".

### D7 — `io.Writer` 부재 실측이 §A 에 없다 (optional)

`spec.md` §A 전체 — Severity: minor — Class: **optional**

HISTORY 는 "출력 seam 을 REQ-CSC-005 에 반환값 형태로 확정했고(iter-2 N6 — `internal/template` 에 `io.Writer` 가 **없다**는 실측)"이라고 적지만, §A.1~§A.11 어디에도 그 실측 절이 없다. 실측 자체는 참이다 — 내가 재측정했다(§5.2 항목 8). 문제는 §A 가 "본 SPEC 이 근거로 삼는 실측을 모아 두는 절"이라는 자기 규약과 어긋난다는 것이며, D1 의 인용 오류가 발생한 구조적 원인이기도 하다(가리킬 §A 절이 없어서 §B 절을 가리켰다).

Required fix: D1 과 함께 처리한다 — §A 에 절을 신설하고 REQ-CSC-005·plan M2·AC-CSC-006 이 그것을 가리키게 한다.

## §5. 재측정 기록 (모든 수치를 직접 측정 — 선행 보고서·preflight 노트에서 이월한 값 0건)

### 5.1 SPEC 이 §A 에 적은 값과의 대조

| # | SPEC 주장 | 절 | 내 측정 | 판정 |
|---|---|---|---|---|
| 1 | 템플릿 배포 대상 스킬 34 | §A.1 | `find internal/template/templates/.claude/skills -mindepth 1 -maxdepth 1 -type d \| wc -l` → **34** | 일치 |
| 2 | 로컬 스킬 44, 그중 `hns-*` 10 | §A.1 | **44** / **10** (44 = 34 + 10) | 일치 |
| 3 | 템플릿 스킬 중 `SKILL.md` 보유 34 | §A.1 | `find … -mindepth 2 -maxdepth 2 -name SKILL.md \| wc -l` → **34** | 일치 |
| 4 | 템플릿 트리 심볼릭 링크 0 | §C 사전점검 | `find internal/template/templates -type l \| wc -l` → **0** | 일치 |
| 5 | `.agents/` 는 로컬·템플릿 양쪽에 부재 | §A.1 | `ls` 둘 다 rc=1 (No such file or directory) | 일치 |
| 6 | 카탈로그 스킬 34 = core 21 + optional-pack 13, `harness_generated.skills` 는 빈 슬롯 | §A.3 | YAML 파싱: 총 **34**, `core 21` / `optional-pack:devops 5` `:frontend 4` `:backend 3` `:design 1` = **13**. `catalog.yaml:262` `skills: []` | 일치 (부분합까지) |
| 7 | `moaiSkillPrefix = "moai-"`, `grep -cv '^moai'`→0, `grep -cv '^moai-'`→1, 그 이름은 `moai` | §A.9 | `skills_manifest.go:15` `const moaiSkillPrefix = "moai-"`. 실측 **0** / **1** / `moai`. 카탈로그에 이름 `moai` 존재 확인 | 일치 |
| 8 | `internal/template` 에 `io.Writer` 없음, `Deploy` 서명은 `(ctx, projectRoot, m manifest.Manager, tmplCtx *TemplateContext) error` | plan M2 / AC-006 | `grep -rn 'io.Writer' internal/template/` → Go 소스 매치 **0건**(템플릿 마크다운 1건뿐). `deployer.go:100` 서명 일치 | 일치 |
| 9 | `ManagedCleanTargets` 7개, 전부 `.claude/` 하위, `.agents/` 없음, `.claude/skills/moai*` 가 4번째 | §A.4 / §A.10 | 함수 본문 직접 확인: 항목 **7**개, 전부 `defs.ClaudeDir` 기반, 4번째가 `SkillsSubdir + "moai*"` (`IsGlob: true`) | 일치 |
| 10 | clean 이 deploy 보다 앞, `:297` / `:323` | §A.10 | `update_template_sync.go:297` `deploy.CleanMoaiManagedPaths`, `:323` `deployer.Deploy` | 일치 |
| 11 | `backupThenRemove` 가 `os.Stat` → `IsNotExist` → `return 0, nil` | §A.10 | `deploy.go:371-378` 그대로 | 일치 |
| 12 | dangling 링크에서 `Stat` 실패·`Lstat` 성공·`Glob` 매치 | §A.10 | **직접 재현**(§5.2) | 일치 |
| 13 | `Track`→`HashFile`→`os.Open`+`io.Copy`, 디렉터리 링크에서 EISDIR | §A.6 | `manifest.go:136` / `hasher.go:17,29`. 직접 재현: `open dirlink err: <nil>` / `read dirlink err: … is a directory` | 일치 |
| 14 | 템플릿 `.gitignore` 에 `.agents` 항목 없음 | §A.8 | `grep -n 'agents' internal/template/templates/.gitignore` → rc=1 | 일치 |
| 15 | 재설계 문서가 worktree 에 없고 primary 에 있음 | §F | worktree rc=1 / primary `7425 bytes` | 일치 |
| 16 | 주석에 "배포가 곧바로 다시 쓰므로…" 사유가 적혀 있음 | §A.11 | **불일치 — D5** | **틀림** |

**16개 중 15개가 정확하다.** §A 의 실측 품질은 높다. 어긋난 하나는 코드 주석 인용이며 D5 로 올렸다.

### 5.2 직접 실행한 재현 (worktree 내부 임시 프로그램, 실행 후 삭제)

```
Stat err: stat …/mirror/moai-x: no such file or directory  IsNotExist: true
Lstat err: <nil>
glob matches: 1 […/mirror/moai-x]
open dirlink err: <nil>
read dirlink err: read …/lnk: is a directory
```

`filepath.Glob` 은 dangling 링크를 매치하고 `os.Stat` 은 `IsNotExist` 를 낸다 — §A.10 의 결함 메커니즘을 독립적으로 확인했다. 디렉터리 링크에서 `os.Open` 성공 + 읽기 EISDIR 도 확인했다 — §A.6 확인.

probe 디렉터리는 `rm -r` 로 제거했고 `git status --short` 무출력을 재확인했다.

## §6. SPEC·선행 감사·preflight 노트의 주장과 어긋나는 것 (별도 명시)

리드 지시대로 이 항목만 따로 모은다.

1. **§A.11 의 주석 인용은 코드와 어긋난다** — `backupThenRemove` doc comment 에 "배포가 곧바로 다시 쓰므로 유일본이 위태로운 적이 없다"는 문장이 없다. 판별 *기준*은 있고 *사유 문장*은 없다. (D5)

2. **§A.11 의 "판별자는 이미 코드에 있다 … 그대로 쓴다"는 미러 경로에 대해 성립하지 않는다** — `templateCarries` 는 `!info.IsDir()` 분기 전용이다. 미러는 링크 또는 디렉터리라 `templateManagedPaths` 를 타고, 템플릿에 `.agents/` 가 없어 그쪽은 항상 공집합이다(SPEC 자신이 §A.7 에 적은 사실). 두 절이 서로를 반박한다. (D5)

3. **a#2 가 처방한 fix 문장 하나가 실행 불가였고, 저자가 그것을 그대로 채택했다** — `plan-audit-iter2.md:199` 의 "정본 `.claude/skills/moai-live/` 가 링크를 통해 지워지지 않았음". 정본 이름이 `moai-live` 인 이상 `.claude/skills/moai*` 청소 대상에 걸려 같은 실행에서 삭제된다. **감사가 처방한 것이 곧 옳다고 볼 수 없다는 사례이며, a#1·a#2 어느 쪽도 이를 지적하지 않았다.** (D2)

4. **HISTORY 가 "optional 도 전부 반영했다"고 적지만, 반영 목록에 든 §A.3 리스팅 예산 축은 "범위 밖" 선언이지 반영이 아니다** — 결정 자체는 정당하고 근거도 적혀 있으므로 결함으로 올리지 않았다. 다만 "반영"과 "범위 밖으로 확정"은 다른 처분이며, 문구가 둘을 합쳐 읽게 한다. 기록만 남긴다.

5. **plan §G 의 AP-16 과 plan §F M3 본문이 서로를 부정한다** — 같은 문서 안의 모순이다. (D3-b)

6. **preflight 노트(`m1-preflight-measurements.md`)의 "36개"는 이 감사에서 채택하지 않았다** — 나는 34 를 직접 측정했고, SPEC §A.1 의 정정(`ls | wc -l` 별칭에 의한 +2)과 일치한다. preflight 노트의 수치는 어떤 판정에도 인용하지 않았다.

7. **`catalog.yaml` 의 `tier:` 문자열 전역 카운트는 45(core 31 / optional-pack 13 / harness-generated 1)이며, 이는 에이전트 항목을 포함한 값이다.** SPEC §A.3 의 "core 21"은 **스킬 항목만** 센 값이고 그쪽이 옳다. 단순 `grep -c 'tier: core'` 로 재검하면 31 이 나와 SPEC 이 틀린 것처럼 보이므로, 다음 감사자를 위해 여기 적어 둔다 — 스킬 항목만 파싱해야 21 이 나온다.

## §7. 이번 판본이 새로 들여온 결함 (리드 지시 — 3차 개정 고유 산물)

| ID | 형태 | 도입 경위(추정) |
|---|---|---|
| **D2** | 판정 계층: MUST AC 한 개가 올바른 구현에서 반드시 FAIL | a#2 의 처방 문장을 검증 없이 채택. fixture 이름 `moai-live` 가 정본 청소 글롭과 충돌한다는 점을 양쪽 모두 놓침 |
| **D3** | 판정 계층: 같은 AC 의 두 단언이 동시에 참일 수 없음 | REQ-CSC-010 한정에 맞춰 3번 팔을 추가하면서 2번 팔의 측정 범위를 좁히지 않음 |
| **D3-b** | 계획 계층: plan 이 자기 AP-16 이 금지한 형태를 처방 + 팔 개수 누락 | spec/acceptance 는 갱신되고 plan M3 만 iter-3 문구로 잔존 |
| **D1** | 근거 사슬: 출력 seam 인용 4건이 §B.D6 을 가리킴 | §B 에 D6(dangling)이 신설되면서 이전 판본의 절 번호가 그대로 이월 |
| **D6** | 개수 표기 3건 드리프트 | 팔을 늘리면서 개수 문구 미갱신 |

**공통 형태가 하나 있다**: v0.4.0 의 수정은 전부 "기존 번호 안의 절 추가"로 이뤄졌고(HISTORY 가 스스로 밝힌 전략), 그 결과 **절이 늘어난 항목의 메타 정보 — 팔 개수, 절 번호 인용, 소유 마일스톤 — 가 함께 갱신되지 않았다.** 예산 상한(16/16)을 지키느라 번호를 못 늘린 것 자체는 정당하지만, 한 번호 안에 절을 겹쳐 쌓는 방식이 이 부류의 잔재를 만든다. 다음 개정에서는 절을 추가할 때마다 **그 항목을 가리키는 모든 지점**(위험 표, 마일스톤 닫힘 조건, 개수 문구, 교차 참조)을 함께 훑는 것을 권한다.

## §8. Regression Check (iteration 2 → 3)

선행 blocking 총 11건(a#1 7건 + a#2 6건, 중복 2쌍 제외) — **전부 RESOLVED**. 증거는 §3.1 표의 각 행. 3회 연속 반복된 결함 0건 → **정체(stagnation) 없음**.

다만 두 건은 "닫혔으나 그 자리에 새 결함이 생긴" 형태다:
- a#1 N2 / a#2 N5 (fixture) → **RESOLVED**, 그러나 확장된 fixture 가 D2 를 낳았다.
- a#2 N2 (무백업 손실) → **RESOLVED**, 그러나 추가된 3번 팔이 D3 를 낳았다.

이것은 진전이다. 결함이 **요구사항 계층에서 판정 계층의 문구 수준으로 내려왔고**, 남은 것은 재작성이 아니라 문장 정정이다.

## §9. 점수 회귀 확인 (LEAN — STOP 신호 판정)

| iter | 점수 | 산출 |
|---|---|---|
| 1 | 0.775 | a#0 |
| 2 | 0.78 / 0.7625 | a#1 / a#2 |
| 3 | **0.7875** | 본 보고서 |

iter(3) 0.7875 > iter(2) 최고값 0.78 → **상승**. 회귀가 아니므로 **STOP 신호를 발하지 않는다.** 다만 상한(3회)에 도달했으므로 다음 처분은 오케스트레이터가 사용자에게 세 선택지로 물어야 한다(PASS-with-debt / 범위 축소 / 상한 연장). 내 권고는 §10 에 적는다.

## §10. Recommendation

**FAIL 이지만, 남은 작업은 문서 재설계가 아니라 좁은 정정 5건이다.** 요구사항 계층은 튼튼하다 — GEARS 16개 전부 적법, 실측 16건 중 15건 정확, 추적 매핑 무결, 선행 blocking 11건 전부 닫힘. 무너진 곳은 판정 계층의 **문구**이며 전부 국소적이다.

착수 순서(위험 큰 것부터):

1. **AC-CSC-008 마지막 단언을 관측 가능한 형태로 재작성한다** (D2, critical). 정본을 청소 글롭 밖 이름으로 옮기거나, 전체 `CleanMoaiManagedPaths` 실행과 좁은 대상 판정을 분리한다. 이 SPEC 이 존재하는 이유인 dangling 팔(단언 2)은 **건드리지 않는다** — 그 팔은 옳다.
2. **AC-CSC-012 2번 팔의 측정 범위를 판별자와 같은 폭으로 좁힌다** (D3, critical). "`.agents/` 서브트리 전체 파일 수 0" → "템플릿에 대응 스킬이 있는 이름의 미러 디렉터리 아래 파일 수 0".
3. **plan §F M3 본문에서 "백업 대상에서 제외하는 명시적 분기"를 제거하고 REQ-CSC-010 의 판별자 문구로 교체, 닫힘 조건을 `AC-CSC-012 (3팔)` 로 정정한다** (D3-b, major).
4. **출력 seam 근거를 §B.D3 + 신설 §A 절로 옮기고, `§B.D6` 인용 4곳을 교체한다** (D1 + D7, major). §A 신설은 예산 대상이 아니다.
5. **문구 정정 3건**: `spec.md:216` 의 `§D.4` → `acceptance.md §D.4` (D4); §A.11 의 주석 인용을 사실에 맞게 (D5); 개수 표기 3건 (D6).

**예산 여유는 필요 없다.** 다섯 건 모두 기존 REQ/AC 번호 안의 문장 교체이며 새 번호를 요구하지 않는다. §G 가 스스로 못박은 원칙 — "예산 논거는 blocking 수정에는 적용되지 않는다" — 도 이번에는 발동할 일이 없다.

**재감사 범위**: 위 5건의 델타에 한정한다. §3.1 의 닫힘 11건과 §5.1 의 실측 16건은 이 스냅샷에서 확인됐으므로 재검이 불필요하다 — 단, 정정이 §A·§B 절 번호를 건드리면 그 절을 가리키는 인용 전수를 다시 훑어야 한다(이번 결함의 발생 형태가 정확히 그것이다).

---

감사자: plan-auditor (iteration 3, 단독)
대상 스냅샷: `2145e2b2a` (감사 시작·종료 시점 모두 clean 확인)
