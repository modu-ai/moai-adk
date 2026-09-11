# SPEC Review Report: SPEC-RESOURCE-SLOT-LEASE-001

Iteration: 2/2 (Tier M 상한 2 — `harness.plan_audit_tier_ceilings`, 마지막 회차)
Verdict: **PASS**
Overall Score: **0.90** (Tier M 통과 기준 0.80, 1회차 0.78 대비 상승 — STOP 신호 없음)

감사 대상 트리: 워크트리 t607, 브랜치 `WT-heavy-test-slot`, HEAD `e79d6761f`(SPEC v0.2.0). 수리 커밋은 `e50cfea93..e79d6761f` 한 개(`e79d6761f fix(SPEC-RESOURCE-SLOT-LEASE-001): plan-audit iter1 repair, v0.2.0 (card t607)`)이고, 코드 트리(`internal cmd pkg`)는 `c4ce42eca`, `e50cfea93`, `e79d6761f` 세 트리에서 동일하다(Evidence E1).

작성자 추론 맥락은 받지 않았다(M1 맥락 격리). 입력은 spec.md, plan.md, acceptance.md, progress.md, 1회차 보고서 `.moai/reports/t607/plan-audit-iter1.md`, 운영자 결정 `.moai/reports/t607/verdict.md`, 그리고 SPEC이 기대는 코드다. 이번 회차는 1회차 결함 D1–D12의 변화분과, 수리가 새 결함을 만들지 않았는지에 한정했다.

적용한 정책 규칙: `verification-claim-integrity.md` §1·§2·§2.2, `verification-completeness.md` §1.1·§2·§2.1, `.claude/rules/local/gitflow-lane-protocol.md` §8, `spec-workflow.md` § SPEC Complexity Tier(REQ/AC 상한).

---

## 결론 요약

blocking 결함 D1–D6은 모두 본문에서 해결됐고, optional D7–D12도 모두 반영됐다. 번호 이동(옛 AC-RSL-002 → 001b, 옛 AC-RSL-016 → 003c, 옛 REQ-RSL-008 → REQ-RSL-005) 뒤에도 REQ 16개 모두가 AC를 가지며 고아 AC는 없다. 필수 통과 기준 일곱 개는 모두 통과한다.

새로 찾은 결함은 네 건이고 모두 minor·optional이다. 가장 무게 있는 것은 N1이다. 가드는 훅 루트를 git common dir로 정규화하지만 CLI 쪽은 여전히 `CLAUDE_PROJECT_DIR`를 정규화 없이 1순위로 쓴다. 그래서 REQ-RSL-008의 "CLI가 쓰는 바로 그 공유 루트"라는 문장은 CLI 환경의 `CLAUDE_PROJECT_DIR`가 링크된 워크트리를 가리키는 구성에서 성립하지 않는다. 이 세션에서 잰 CLI(Bash) 환경의 `CLAUDE_PROJECT_DIR`는 비어 있으므로(E6) 기본 경로에서는 일어나지 않고, plan이 이미 "같은 해석 함수 하나"를 권장하므로 run 단계에서 한 줄 결정으로 닫을 수 있다.

---

## Must-Pass Results

- [PASS] **MP-1 REQ 번호 일관성** — spec.md:63-111에 REQ-RSL-001부터 016까지 빈칸·중복 없이 이어진다. REQ-RSL-008은 가드 루트 정규화(spec.md:86)로 교체됐고, 옛 pid 0 조항은 REQ-RSL-005(spec.md:76)로 옮겨졌다. 번호 재사용 사실이 spec.md:78과 HISTORY(spec.md:157)에 적혀 있다. 요구사항 층에서 판정했다.
- [PASS] **MP-2 GEARS 형식** — 요구사항 층(spec.md §C)에서 판정했다. 바뀐 문장 다섯 개를 다시 봤다. REQ-RSL-004(spec.md:72)는 State-driven 두 절, REQ-RSL-005(:76)는 Event-driven 두 절, REQ-RSL-008(:86)은 Ubiquitous + When 복합형, REQ-RSL-012(:98)는 Where(정적 설정 상태), REQ-RSL-014(:104)는 While+When 복합형이다. 모두 GEARS 패턴과 맞는다. 트리에서 빌드한 lint가 `ModalityUnjudged`를 포함해 발견 0건을 냈다(E2). 인수 기준 층의 Given-When-Then은 검증 층 형식이므로 여기서 보지 않았다. 표기상 흠(REQ-RSL-002의 소문자 "where"가 조건절로 쓰임, REQ-RSL-014 라벨 "Event-driven")은 N4로 따로 적었다.
- [PASS] **MP-3 프론트매터** — spec.md:2-14에 12개 정식 필드가 모두 있다: `id`, `title`, `version: "0.2.0"`(따옴표), `status: draft`, `created`/`updated: 2026-09-12`, `author`, `priority: P1`, `phase: "v3.2.0 target"`, `module`, `lifecycle: spec-anchored`, `tags`(쉼표 문자열), 그리고 `tier: M`. 거부 별칭은 없다. plan.md와 acceptance.md 프론트매터에도 `status:`가 없다(무상태 규칙 준수).
- [PASS] **MP-4 언어 중립성** — REQ-RSL-013(spec.md:101), REQ-RSL-015(:108)가 내장 언어 목록과 언어 명명 자리표시자를 금지한다. 1회차 D6이 지적한 기계 검증 공백은 AC-RSL-014 (c)(d)(e)와 16개 언어 도구 토큰 명시 목록(acceptance.md:306-312)으로 닫혔다. 16개 언어가 모두 같은 무게로 열거돼 있다.
- [PASS] **MP-5 D7 교차 SPEC** — spec·plan·acceptance가 참조하는 SPEC은 `SPEC-INTEGRATION-LOCK-ATOMIC-001`, `SPEC-INTEGRATION-LOCK-LIVENESS-001` 둘이고 모두 `status: completed`다(E5). retired·superseded·archived가 없어 BLOCKING이 없다.
- [PASS] **MP-6 D8 교차 플랫폼** — 세 파일의 `syscall` 등장 수는 각각 0이다(E5). 자동 통과다.
- [PASS] **MP-7 명확화 게이트** — `grep -rn 'NEEDS CLARIFICATION'`이 SPEC 디렉터리 전체에서 종료 코드 1로 끝났다(E5). research.md는 Tier M이라 없다.

## Operator Decision Conformance (verdict.md 결정표 재대조)

| 결정 항목 | 0.2.0 반영 위치 | 판정 |
|---|---|---|
| 범용 임대 명령 `moai slot acquire\|status\|release --resource` | spec.md:34, REQ-RSL-001(:63), plan.md:123 | 준수 |
| 선택형 PreToolUse 가드, 기본 꺼짐 | REQ-RSL-011/012(:95, :98), Exclusions(:185-186) | 준수 |
| 템플릿·바이너리 배포, 템플릿 문서 최소 | REQ-RSL-015(:108), plan.md M5(:136-141) | 준수 |
| 기록 필드(자원·세션 id/이름·pid·명령·시작 시각·선언 상한) | REQ-RSL-002(:66), AC-RSL-003a/b | 준수. 이름·명령은 생략 시 빈 값으로 기록되고 `(not given)`으로 표시된다. 필드 자체는 항상 기록되므로 결정의 취지를 유지한다 |
| 두 세션 대조군 필수 | AC-RSL-001a(acceptance.md:114-126) | 준수. 구성 결함(1회차 D2)은 해결 |
| kanban 락의 코드 방식만 재사용, `moai integration`과 분리 | REQ-RSL-001, spec.md §D(:116), AC-RSL-013 | 준수 |
| 두 문서 편집은 t637 병합 게이트 뒤 | REQ-RSL-016(:111), AC-RSL-015, plan.md M6 | 준수. 게이트는 여전히 닫혀 있다(E4) |

## Regression Check (1회차 결함 D1–D12)

| 결함 | 판정 | 근거 (파일:줄, 코드 대조) |
|---|---|---|
| D1 GUARD-ROOT-DIVERGENCE (blocking) | **RESOLVED** | spec.md:47이 훅 해석을 `CLAUDE_PROJECT_DIR` → `os.Getwd()` 두 단계로 정정했다. 코드와 일치한다: `internal/hook/path_resolve.go`:78 `resolveProjectRootFromEnvAt`는 환경변수 다음 `os.Getwd()`만 쓰고, `preToolHandler.projectRoot()`(`pre_tool.go`:400)가 이 해석을 쓴다. CLI는 `internal/cli/integration.go`:46 `integrationLockRoot`에서 `CLAUDE_PROJECT_DIR` → `git rev-parse --git-common-dir` → cwd 순서다. 정규화 요구는 REQ-RSL-008(spec.md:86), 설계는 plan.md:40, 검증은 AC-RSL-016(acceptance.md:345-361)이다. AC-RSL-016은 wt-deny, primary-deny, no-git 세 행과 "정규화를 지운 뮤턴트는 wt-deny에서 허용" 대조 행을 갖는다. 남은 비대칭은 N1 |
| D2 CONTROL-TEST-DEADLOCK (blocking) | **RESOLVED** | acceptance.md:116-121이 "B의 완료 또는 풀림 타임아웃(대기 예산의 1/3 이하) 중 먼저 오는 쪽"으로 다시 썼고, busy와 STALLED 미도달을 하네스 결함으로 규정하며, 소유자 pid를 부모 테스트 프로세스로 고정한다. plan.md:104-106도 같다. 출처 대조: `integration_lock_cross_test.go`:63 `integrationStallReleaseTimeout = 500 * time.Millisecond`, 설명 주석 :46-62, `ownerPID := os.Getpid()` :186. 인용 줄 번호 `:180-183`은 실제 주석 위치(:182-186)와 두어 줄 어긋난다. 1회차 보고서 자신의 인용도 같았다(N3에 포함) |
| D3 BUSY-OUTCOME-MISSING (blocking) | **RESOLVED** | REQ-RSL-004(spec.md:72)가 busy 결과를 추가했다. busy는 held와 구별되고, busy 판정 참·held 판정 거짓이며, 기록을 바꾸지 않는다. AC-RSL-002(acceptance.md:138-148)는 acquire_busy·release_busy·held_is_not_busy를 양방향으로 확인하고 뮤턴트 두 방향을 적었다. 코드 선례 `integration_lock_mutation.go`:125-128이 예산 소진 시 `ErrIntegrationLockBusy`를 `%w`로 반환한다(E3) |
| D4 LITERAL-BASELINE-PIN (blocking) | **RESOLVED** | AC-RSL-013(c)(acceptance.md:288-295), AC-RSL-015 닫힘 가지(:333-341), §D.5(:395), plan.md:148, 반패턴(plan.md:155)이 모두 `CARD_BASE=$(git merge-base develop HEAD)` + 대조군 ≥ 1 + "0이면 측정 불가" + 병합 전 전용 + 병합 후 트리 동일성 대체 형태다. lane 규칙 §8(`.claude/rules/local/gitflow-lane-protocol.md`:94, :100)과 문구까지 일치한다. `BASELINE_SHA` 리터럴은 세 파일 어디에도 남지 않았다 |
| D5 REQ-014-012-CONFLICT (blocking) | **RESOLVED** | REQ-RSL-012(spec.md:98)가 "설정을 읽을 수 없음"을 조용한 꺼짐으로 명시했다. REQ-RSL-014(:104)는 "`enabled`가 참으로 읽힌 뒤"로 좁혀졌고 :106에 경계를 적었다. AC-RSL-010 (iii) nil 설정(acceptance.md:233)과 AC-RSL-012 (e) 잘못된 자원 항목(:270-271)이 두 쪽을 각각 검증한다. (e)의 도달성 조건(자원 구획을 관대하게 해석)은 plan.md:59에 설계로 고정됐다 |
| D6 NEUTRALITY-UNVERIFIED (blocking) | **RESOLVED** | AC-RSL-014에 16개 언어 도구 토큰 명시 목록(acceptance.md:306-312), 설정·규칙 두 파일의 부재 판정 (c)(d), 스크래치 사본 양성 대조 (e), 자리표시자 구조 판정 (b)가 추가됐다. EL-7(부재)과 EL-8(양성 대조, 7건)을 다시 재 보니 기록과 같았다(E2). 명령이 `<TOOL_TOKENS>` 이름을 치환해야 실행되는 형태라는 점은 N3 |
| D7 OQ1-MEASURABLE-NOW (optional) | RESOLVED | plan.md:93이 OQ-1을 1회차 E3 측정으로 닫았다. spec.md §G(:143)와 Exclusions "세션 내부 병렬 실행"(:167-168)이 추가됐다. 런타임 버전은 E3에 없다고 정직하게 적었다 |
| D8 FACTUAL-SLIP-PRIVATE (optional) | RESOLVED | plan.md:42가 `FactoryProcessAlive`를 공개 함수로 고쳤다. 배치 근거는 비공개인 `acquireBoardLockImpl`·`boardLockWaitBudget`으로 좁혀졌다 |
| D9 DATED-DEVELOP-SHA (optional) | RESOLVED | spec.md:57과 EL-5(acceptance.md:52-55)가 develop SHA를 측정 시점 값(`eb50af5a8` → `e82ef5565` → `ac6c42c2d`)으로 표시하고 핀하지 않는다. 이번 실행의 develop은 `ac6c42c2d`, 게이트는 종료 코드 1이다(E4) |
| D10 TEST-ENV-SCRUB (optional) | RESOLVED | spec.md:122, plan.md:87, acceptance.md:17이 `CLAUDE_CODE_SESSION_ID`·`MOAI_SESSION_PID`를 고정·비움 목록에 넣었다 |
| D11 OPTIONAL-FIELDS-UNSPECIFIED (optional) | RESOLVED | REQ-RSL-002(spec.md:66)가 빈 값 기록과 `(not given)` 표시를 정했다. AC-RSL-003b(acceptance.md:153)가 키 존재와 빈 문자열, 사람용 출력을 확인한다 |
| D12 ALLOW-BRANCH-ATTRIBUTION (optional) | RESOLVED | AC-RSL-011 표(acceptance.md:249-259)가 허용 행마다 감사 사유(`allow-self`/`allow-stale`/`allow-expired`/`allow-unheld`/감사 줄 없음)를 단언한다. 뮤턴트 설명(:263)이 "보유자 있음" 검사 삭제가 `allow-expired`로 새는 경로를 사유 단언이 잡는다고 적는다 |

미해결 결함은 없다. 세 회차에 걸쳐 그대로 남은 정체 결함도 없다.

## 수리 이후 구조 점검

- **추적성(번호 이동 후).** §D 행렬(acceptance.md:82-99)과 §D.3(:365-382)을 줄 단위로 대조했다. REQ-RSL-001 → 003c·013, 002 → 003a/b, 003 → 001a/b, 004 → 002·004, 005 → 005, 006 → 006, 007 → 007, 008 → 016, 009 → 008, 010 → 009, 011·013 → 011, 012 → 010, 014 → 012, 015 → 014, 016 → 015. 16개 REQ 모두 AC가 있고, 16개 AC 모두 존재하는 REQ를 가리킨다. 세 파일에 걸친 AC 참조 전수 추출(E7)에서 옛 의미의 AC-RSL-002나 AC-RSL-016을 가리키는 잔존 참조는 없다. plan 마일스톤의 AC 배정(plan.md:119, :127, :134, :141)도 새 번호와 맞는다.
- **REQ/AC 예산.** REQ 16개, 최상위 AC 16개로 Tier M 상한(각각 16)에 딱 맞는다. 하위 문자 AC의 관례 판단은 이렇다. 005a-c와 006a-d는 한 요구사항의 사례 분할이라 관례 안이다. 001b는 001a의 뮤턴트 짝으로, 다른 AC가 "뮤턴트 짝" 항목으로 안에 두는 것과 성격이 같아 관례 안이다(심각도 등급만 따로 매겼다, acceptance.md:106). 003c(CLI 표면 + 교차 플랫폼 빌드)는 REQ-RSL-001과 §D 플랫폼 제약을 검증하는 별개 관심사를 "기록 필드" AC 아래로 옮긴 것이라 경계선에 있다. 추적성은 온전하고(§D.3이 REQ-RSL-001 → 003c로 적는다) lint도 발견 0건이므로 blocking이 아니다. 다만 상한 회피용 재라벨링의 선례가 될 수 있어 N2로 적는다.
- **내부 일관성.** REQ-RSL-012와 014의 경계(설정 로드 실패 = 조용한 꺼짐, 켜진 뒤 불확실성 = fail-open)는 spec, plan B4, AC-010/012가 서로 모순 없이 맞물린다. REQ-RSL-013의 허용 경로 넷은 AC-011 사유 넷과 일대일이다. REQ-RSL-004의 busy는 acquire·release 모두에 걸리고 AC-002 (a)(b)가 둘 다 본다.
- **원장 EL-7/EL-8 재측정.** 기록과 같다(E2). EL-7은 `e50cfea93`에서 쟀다고 되어 있는데, `e50cfea93`과 `e79d6761f` 사이 코드 트리가 같으므로(E1) 이번 측정도 같은 대상을 잰 것이다.
- **AC-RSL-014 (f)의 `\b`.** POSIX ERE에서 `\b`가 단어 경계로 동작하지 않아 전부 무출력이 나는 알려진 위험을 확인했다. 이 트리의 `/usr/bin/grep`(BSD grep 2.6.0-FreeBSD)은 `-E`에서 `\b`를 경계로 처리해 `t607`을 적중시켰다(E8). 판정식은 공허하지 않다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75 밴드 상단 | 1회차의 세 모호점(D1·D5·D11)은 해소됐다(spec.md:47, :98, :104, :66). 남은 것은 N1(CLI 루트 비정규화로 REQ-RSL-008의 "같은 공유 루트" 주장이 한 구성에서 깨짐)과 N4(no-git 경우 감사 로그 위치 미정, 조건절 "where" 표기)다. 합리적인 엔지니어라면 대부분 같은 해석에 이른다 |
| Completeness | 1.0 | 1.0 | HISTORY §H(spec.md:152-157), 배경 §A, 요구사항 §C, 제약 §D, 위험 §G, plan, acceptance, Exclusions H3 9개(spec.md:161-186, 각 `-` 항목)가 모두 있다. 프론트매터 12필드도 완전하다 |
| Testability | 0.85 | 0.75 밴드 상단 | D2(교착)와 D4(흡수 후 거짓 실패)가 해소돼 모든 AC가 이진 판정 가능하다. 남은 흠은 둘이다. AC-RSL-014 (c)(d)와 EL-7/8이 `<TOOL_TOKENS>` 치환을 거쳐야 실행된다(N3). AC-RSL-016 no-git 행의 감사 줄 위치가 정해져 있지 않다(N4) |
| Traceability | 0.90 | 0.75 밴드 상단 | REQ 16개 모두 AC가 있고 고아 AC가 없다(§D.3). REQ-RSL-014 "unavailable configuration" 절과 REQ-RSL-015 언어 중립 절이 이제 AC-010 (iii)과 AC-014 (c)(d)(e)로 추적된다. AC-016 no-git 행이 요구하는 감사 줄은 REQ-RSL-008 문장에 없고(REQ는 advisory만 명시) plan B2에만 있다(N4) |

집계: (0.85 + 1.0 + 0.85 + 0.90) / 4 = **0.90**. Tier M 기준 0.80 이상이다. 필수 통과 기준은 모두 통과했고 blocking 결함은 남아 있지 않다.

## Defects Found (structured defect-list)

D1–D12(1회차)는 모두 RESOLVED다(위 Regression Check). 이번 회차의 새 발견은 아래와 같다.

N1. CLI-ROOT-NOT-NORMALIZED — spec.md:86 (REQ-RSL-008), plan.md:39 — REQ-RSL-008은 가드가 "the same shared root the slot lease CLI writes to"를 읽는다고 쓰고, 그 수단으로 훅 루트를 git common dir의 부모로 정규화한다. 그런데 plan B2(:39)의 CLI 해석은 통합 창과 같은 순서라 `CLAUDE_PROJECT_DIR`가 있으면 그 값을 정규화 없이 쓴다(`internal/cli/integration.go`:46-49). CLI 프로세스 환경의 `CLAUDE_PROJECT_DIR`가 링크된 워크트리를 가리키면 CLI는 워크트리에 쓰고 가드는 primary를 읽는다. 1회차 D1의 거울상이다. 이 세션의 Bash 환경에서 잰 값은 빈 문자열이라(E6) 기본 경로에서는 CLI가 git common dir를 거쳐 primary에 쓴다. 따라서 이 어긋남은 사용자가 셸에서 그 변수를 워크트리로 내보낸 구성에 한정된다. plan.md:40은 "CLI와 가드가 같은 해석 함수 하나"를 권장만 한다. — Severity: minor — Class: optional — Required fix: plan B2(:39)에 "CLI도 `CLAUDE_PROJECT_DIR`로 얻은 루트를 git common dir의 부모로 정규화한다(가드와 같은 함수)"를 한 줄 넣고, 권장을 결정으로 올린다. 또는 REQ-RSL-008 문장을 "the primary root derived from the git common directory"로 바꿔 CLI 동치 주장을 빼고, CLI 쪽 동작을 §G 위험으로 적는다. 어느 쪽이든 `internal/cli/integration.go`는 건드리지 않는다(AC-RSL-013(c) 프로브 범위).

N2. AC-003C-BUDGET-RELABEL — acceptance.md:154, :367 — 옛 AC-RSL-016(CLI 표면 + 교차 플랫폼 빌드)이 "기록 필드" AC 아래 003c로 옮겨졌다. 검증 대상은 REQ-RSL-001과 §D 플랫폼 제약이고 마일스톤도 M3로 달라, 주제상 AC-003에 속하지 않는다. 추적성은 온전하고 상한 규칙은 AC id를 세므로 위반은 아니다. 하지만 이 방식이 선례가 되면 "상한을 넘으면 tier를 올리거나 쪼갠다"는 예산 규칙이 재라벨링으로 우회된다. — Severity: minor — Class: optional — Required fix: 조치 불요. 다음 개정 때 AC를 하나 더 늘릴 일이 생기면 003c를 독립 AC로 되돌리고 tier 판단을 다시 한다는 메모를 HISTORY에 남기는 정도로 충분하다.

N3. TOKEN-PLACEHOLDER-NOT-VERBATIM — acceptance.md:65, :71 (EL-7/EL-8), :317-318 (AC-RSL-014 (c)(d)), 부수적으로 :116과 plan.md:106의 `integration_lock_cross_test.go:180-183` 인용 — EL-7/EL-8과 AC-014 (c)(d)의 명령은 `<TOOL_TOKENS>` 이름을 목록 문자열로 바꿔 넣어야 실행된다. `verification-completeness.md` §2.1은 원장 명령이 "원문 그대로 읽어 수정 없이 실행되는" 형태이기를 요구한다. 치환 규칙이 한 줄로 명확하고(acceptance.md:78) 이 감사에서 치환해 재현했으므로(E2) 판정은 가능하다. 다만 셸 인용이 틀리면(목록에 공백이 있는 토큰이 여럿 있다) 무출력이 "부재"로 잘못 읽힐 수 있다. 부수적으로 부모 pid 고정 출처 인용 `:180-183`은 실제 주석 :182-185와 대입 :186에서 두어 줄 어긋난다(E3). — Severity: minor — Class: optional — Required fix: 토큰 목록을 SPEC 디렉터리의 한 줄 파일(예: `tool-tokens.txt`, 한 줄에 토큰 하나)로 두고 판정을 `/usr/bin/grep -nwiE -f <tokens-file> <target>` 단일 호출로 바꾼다. 인용 줄 번호는 run 단계에서 다시 잰다.

N4. NOGIT-AUDIT-LOCATION / WHERE-AS-CONDITION — acceptance.md:355 (AC-RSL-016 no-git 행), spec.md:86, :104, plan.md:40-41; 표기상 spec.md:66, :104 — (a) AC-016 no-git 행은 "감사 `fail-open`"을 요구하지만 REQ-RSL-008은 advisory만 명시한다. 감사 로그는 plan.md:41에 따르면 `<primary>/.moai/logs/`인데 이 경우는 primary를 구할 수 없는 경우라, 테스트가 어느 경로의 로그를 단언할지 정해져 있지 않다. REQ-RSL-014의 불확실성 열거에도 "정규화 실패"가 없다(plan B4 :57에는 있다). (b) REQ-RSL-002의 "where the caller omits…"는 GEARS의 Where(능력·정적 설정 게이트)가 아니라 입력 조건이다. REQ-RSL-014는 While+When 복합형인데 라벨이 "Event-driven"이다. — Severity: minor — Class: optional — Required fix: (a) REQ-RSL-014 열거에 "a hook root that cannot be normalized to the shared root"를 추가하고, AC-016 no-git 행에 감사 로그 경로를 "훅 루트(정규화 전) 아래 `.moai/logs/slot-lease-audit.jsonl`"로 명시한다. (b) REQ-RSL-002의 절을 "if the caller omits"로 바꾸기보다 "and, for an omitted session name or command, shall record …"처럼 응답 절로 흡수하고, REQ-RSL-014 라벨을 "State+Event"로 고친다. lint는 현재 두 표기를 받아들이므로 급하지 않다.

blocking 결함: 없음.

## Recommendation

PASS. 필수 통과 기준별 근거:

- MP-1: spec.md:63-111, REQ-RSL-001..016 연속, REQ-RSL-008 재사용은 :78과 HISTORY :157에 기록.
- MP-2: 요구사항 층 16문장 GEARS, lint `ModalityUnjudged` 0(E2).
- MP-3: spec.md:2-14 12필드 + `tier: M`, 형식 적합.
- MP-4: 16개 언어 도구 토큰 전수 열거와 부재 판정·양성 대조(acceptance.md:306-319), 기준선 EL-7/EL-8 재현(E2).
- MP-5/6/7: 참조 SPEC 둘 모두 `completed`, `syscall` 0, `NEEDS CLARIFICATION` 0(E5).

optional N1–N4는 오케스트레이터 재량이다. 이 중 N1은 run 단계 M4 착수 전에 plan B2 한 줄로 닫는 편이 가장 싸다. 공유 루트 해석 함수를 하나로 정하면 N1과 1회차 잔여 위험(두 경로가 따로 풀면 다시 어긋남)이 함께 닫힌다. Tier M 상한 2회에 도달했으므로 이 결함들을 이유로 plan 감사를 더 돌리지 않는다.

Implementation Kickoff Approval은 이 PASS와 무관하게 필수다. 킥오프 때 운영자에게 보여야 할 SPEC 자체의 결정은 셋이다. 선언 상한 만료 시 강제 없는 인수(REQ-RSL-006, OQ-4), 한 세션 안의 병렬 무거운 실행은 직렬화하지 않음(Exclusions), 통합 창 가드의 같은 루트 노출은 후속 카드로 넘김(Exclusions).

---

## 5-Section Evidence

### Claim

- C1. 1회차 blocking D1–D6과 optional D7–D12는 현재 파일 본문에서 모두 해결됐다.
- C2. 필수 통과 기준 MP-1..MP-7은 HEAD `e79d6761f`에서 모두 통과한다.
- C3. 원장 EL-7·EL-8은 다시 재도 기록과 같다.
- C4. 번호 이동 뒤 REQ↔AC 추적성은 양방향으로 온전하고, 옛 번호 의미의 잔존 참조는 없다.
- C5. 새 결함 N1–N4는 모두 minor·optional이다.
- C6. 집계 0.90으로 Tier M 기준 0.80을 넘는다.

### Evidence

- E1 (트리 동일성): `git diff --stat c4ce42eca e50cfea93 -- internal cmd pkg` → 출력 없음, `diff1 exit=0`. `git diff --stat e50cfea93 e79d6761f -- internal cmd pkg` → 출력 없음, `diff2 exit=0`. `git log --oneline e50cfea93..e79d6761f` → `e79d6761f fix(SPEC-RESOURCE-SLOT-LEASE-001): plan-audit iter1 repair, v0.2.0 (card t607)`.
- E2 (lint, EL-7/EL-8 재측정):
  - `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=e79d6761f" -o <scratchpad>/moai-e79 ./cmd/moai` → `build exit=0`. `<scratchpad>/moai-e79 version` → ` v3.1.3   e79d6761f   built unknown`.
  - `<scratchpad>/moai-e79 spec lint --strict .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001` → `✓ No findings — all SPEC documents are valid`, `lint rc=0`.
  - EL-7: `/usr/bin/grep -nwiE "$T" internal/template/templates/.moai/config/sections/workflow.yaml`(`T`는 acceptance.md:309의 목록 문자열 그대로) → 출력 없음, `EL-7 exit=1`.
  - EL-8: `/usr/bin/grep -cwiE "$T" internal/template/templates/.claude/rules/moai/languages/python.md` → `7`, `EL-8 exit=0`.
- E3 (코드 대조): `grep -n` → `internal/hook/path_resolve.go:78:func resolveProjectRootFromEnvAt(...)`, `internal/cli/integration.go:46:func integrationLockRoot() string {`, `integration_lock_cross_test.go:63: integrationStallReleaseTimeout = 500 * time.Millisecond`, `:186: ownerPID := os.Getpid()`. `pre_tool.go`:400 `projectRoot()`는 `h.resolve(&h.projectDir)`이고 생성자(:364-373)가 `resolveProjectRootFromEnv`로 늦게 푼다. `integration_lock_mutation.go`:125-128이 예산 소진 시 `ErrIntegrationLockBusy`를 `%w`로 반환한다. lane 규칙 `.claude/rules/local/gitflow-lane-protocol.md`:94 `## 8. 검증은 레인-로컬`, :100 `CARD_BASE=$(git merge-base develop HEAD)`.
- E4 (M6 게이트): `git merge-base --is-ancestor WT-acquire-branch-record develop` → `gate exit=1`. `git rev-parse --short develop` → `ac6c42c2d`, `git rev-parse --short WT-acquire-branch-record` → `f680dab46`.
- E5 (D7/D8/MP-7): spec·plan·acceptance의 SPEC 참조 추출 → `SPEC-INTEGRATION-LOCK-ATOMIC-001`, `SPEC-INTEGRATION-LOCK-LIVENESS-001`(자기 id 제외). `grep '^status:'` → 둘 다 `status: completed`. `grep -c syscall` → spec.md:0, plan.md:0, acceptance.md:0. `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-RESOURCE-SLOT-LEASE-001/` → 출력 없음, `mp7 exit=1`.
- E6 (N1 관측): 감사자 Bash 환경 `echo "CPD=[${CLAUDE_PROJECT_DIR}]"` → `CPD=[]`. `git rev-parse --git-common-dir`(워크트리 안) → `/Users/goos/MoAI/moai-adk-go/.git`.
- E7 (AC 참조 전수): `/usr/bin/grep -no 'AC-RSL-[0-9]\{3\}[a-d]\{0,1\}' spec.md plan.md acceptance.md | … | uniq -c`의 결과에 AC-RSL-001, 001a, 001b, 002, 003, 003a, 003b, 003c, 004–016이 나타나고, 목록 밖의 id는 없다. `grep -n 'REQ-RSL-008\|REQ-RSL-005'` → REQ-RSL-008은 spec.md:47, :78, :86, :157, plan.md:40, :154, acceptance.md:99, :374에서만 새 의미(루트 정규화)로 쓰인다.
- E8 (`\b` 동작): 스크래치 파일(`see t607 here` / `SPEC-ABC-001` / `2026-09-12` / `nothing`)에 `/usr/bin/grep -nE '\bt[0-9]{3}\b'` → `1:see t607 here`, `b-only exit=0`. `/usr/bin/grep -V` → `grep (BSD grep, GNU compatible) 2.6.0-FreeBSD`.

### Baseline-attribution

- 모든 측정은 이번 실행에서 워크트리 t607, HEAD `e79d6761f`에 대해 했다. 문서 핀 `c4ce42eca`와 EL-7/8 트리 `e50cfea93`는 코드 트리가 HEAD와 같다(E1). 그래서 인용된 줄 번호와 원장 값을 HEAD에서 다시 잰 것이 같은 대상을 잰 것이다.
- lint는 HEAD `e79d6761f`에서 커밋 ldflags를 넣어 빌드한 스크래치 바이너리를 경로로 호출했다. 판정 빌드 좌표는 바이너리가 스스로 보고한 `e79d6761f`이고 트리 HEAD와 같다(§2.2). 설치본(`ed71054d3`)은 쓰지 않았다.
- develop과 게이트 값(`ac6c42c2d`, 종료 코드 1)은 움직이는 ref의 감사 시점 값이다(SUBJECT형, R4). 판정 명령은 run 종료 시점에 다시 실행해야 한다.

### Gaps

- spec lint 음성 대조(`phase: plan` 주입)는 이번에도 재현하지 못했다. 스크래치 사본은 만들었지만, 워크트리 격리 가드가 스크래치 경로 파일 편집을 거부했다(`Path traversal detected`). 발견 0건의 판별력은 progress.md:34의 작성자 대조군 기록에 기댄다.
- cross-model 감사(codex/GLM)는 돌리지 않았다. `.moai/config/sections/`에 `audit_model` 키가 없어 이 저장소가 교차 모델 의견을 요청하지 않는다고 판단했다. 이 판단 자체는 설정 부재를 읽은 것이다.
- N1의 어긋남이 실제로 일어나는 세션 형태(CLI 환경의 `CLAUDE_PROJECT_DIR`가 워크트리를 가리키는 경우)는 재지 못했다. 이 세션 한 번의 관측(빈 값)만 있다.
- 1회차 E3의 OQ-1 관측(서브에이전트 = 부모 세션 id)은 이번에 다시 재지 않았다. 이 회차의 결함 판정에 쓰지 않았다.
- `-run` 필터 AC들과 뮤턴트 표의 실제 판별력은 run 단계 산출물이 없어 설계 수준에서만 판정했다.

### Residual-risk

- 공유 루트를 가드와 CLI가 서로 다른 코드 경로로 풀면 N1 같은 어긋남이 다시 생긴다. plan은 단일 함수를 권장만 한다.
- AC-RSL-016은 `CLAUDE_PROJECT_DIR`를 워크트리로 두어 조건을 만든다. 실제 런타임이 훅 환경에 어떤 값을 주는지와 무관하게 결과가 같도록 요구하므로 SPEC 안에서는 닫혀 있다. 다만 통합 창 가드의 같은 노출은 이 SPEC 밖에 남는다(Exclusions).
- AC-RSL-012 (e)는 설정 로더가 자원 구획을 관대하게 해석해야 도달 가능하다(plan.md:59). 로더가 워크플로 섹션을 엄격하게 디코드하면 run 단계에서 설계 비용이 예상보다 클 수 있다.
- Windows 변경 락 잔재 정리의 행동 증거는 CI 전까지 없다(acceptance.md:386).
