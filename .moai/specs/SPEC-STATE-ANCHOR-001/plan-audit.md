# Plan-Audit — SPEC-STATE-ANCHOR-001

Iteration: 1/2 (Tier M ceiling = 2, `harness.plan_audit_tier_ceilings`)
Verdict: **FAIL** (차단 결함 1건 — 수리 후 delta 재감사)
Overall Score: **0.85** (Tier M PASS threshold 0.80 — 점수 단독으로는 통과선이나, 차단 결함이 판정을 지배한다)
측정 트리: `.claude/worktrees/t510` @ `668b10721` (브랜치 `WT-state-write-locus`; base `0b1e27877` 대비 커밋 2건 모두 docs 전용 — `git diff --stat 0b1e27877..HEAD` = `.moai` 6파일 752 insertions, 코드 변경 0. 코드 좌표 인용은 base와 동일 트리다.)

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — REQ-SA-001..011, 11건 연속·중복 0·패딩 일관 (spec.md:85-101). AC-SA-001..012, 12건 연속 (acceptance.md:12-24).
- **[PASS] MP-2 GEARS 형식 (요구사항 층)** — 11건 전부 5 패턴 문법 적합: REQ-SA-001/002/009/010 Ubiquitous, REQ-SA-003/006 Event-driven("When … shall"), REQ-SA-004/008 Unwanted("shall not"), REQ-SA-007 GEARS 정준 부정형("shall not create … under directories merely visited"), REQ-SA-011 When-트리거형. AC(Given-When-Then)는 검증 층으로 M3 § Scope에 따라 이 기준의 대상이 아니다. 형태 비고 3건은 Defects D7(경미)로 분리 — 문법 위반 아니다.
- **[PASS] MP-3 YAML frontmatter** — 12 정식 필드 전부 존재·타입 적합 (spec.md:1-18): `id`/`title`(quoted)/`version`("0.1.0")/`status: draft`/`created`·`updated`(ISO)/`author`/`priority: P1`/`phase: "v3.2.0 target"`(릴리스 라벨, 금지값 아님)/`module`/`lifecycle`/`tags`(CSV). 기각 대상 별칭(`created_at` 등) 0. `tier: M`·`era: V3R6` 선택 필드. 기계적 확인: `moai spec lint .moai/specs/SPEC-STATE-ANCHOR-001/spec.md` → **0 error(s), 11 warning(s)**, exit 0 — `FrontmatterInvalid`·`FrontmatterPhaseInvalid`·`MissingExclusions` 미발생.
- **[N/A] MP-4 언어 중립성** — 단일 언어(Go) 내부 수리 SPEC, 다중 언어 도구 내용 없음. N/A 자동 통과.
- **[PASS] MP-5 D7 cross-SPEC** — 본문 SPEC-ID 참조는 `SPEC-SESSION-TELEMETRY-001` (frontmatter `related_specs` + spec.md §6 + plan.md §H)뿐. 실재 확인: `.moai/specs/SPEC-SESSION-TELEMETRY-001/spec.md` 존재, `status: completed` — retired/superseded/archived 아님, 화해 조항 불요. BLOCKING 없음.
- **[PASS] MP-6 D8 syscall** — `grep -c syscall spec.md` = 0. D8 자동 통과.
- **[PASS] MP-7 clarification gate** — `grep -c "NEEDS CLARIFICATION"` plan.md 0 / spec.md 0 / acceptance.md 0. research.md은 Tier M 산출물이 아니라 미존재가 정상(N/A).

## Category Scores

| Dimension | Score | 근거 |
|-----------|-------|------|
| Clarity | 0.75 | REQ-SA-004(spec.md:93)·plan.md D2(:48)·M1(:85)·AC-SA-006/012(acceptance.md:104,166)가 표시 경로의 소스를 `current_dir`로 서술했으나 트리는 반대다 — `extractProjectDirectory`(internal/statusline/builder.go:406-438)는 **1순위 `workspace.project_dir`**(416-418행), `types.go:184` 주석도 `project_dir`을 "(used for display)"로 문서화, builder.go:235 주석 "prefer project_dir per documentation". 불변성 의도는 명확하고 AC-SA-006 골든 diff가 실질 보호막이지만, 글자대로 따르면 표시 경로를 "current_dir 소스로 고쳐야 한다"는 오지시가 된다. 그 외 요구사항은 단일 해석 가능. |
| Completeness | 1.00 | 필수 섹션 전부 존재: HISTORY(spec.md:22-28), 문제/원인(§1-2), mutant 대비(§3), 요구사항(§4), 운영자 미결 결정의 명시적 표면화(§5), Out of Scope H3 4개 + 구체 bullet(§6), Gaps(§7). frontmatter 완비. lint ERROR 0이 구조를 기계적으로 corroborate. |
| Testability | 0.90 | 전 AC 이진 판정 가능. weasel word 0. 빈-스윕 방어가 AC 본문에 구속돼 있다 — AC-SA-010 "N=0이면 테스트가 스스로 실패한다"(acceptance.md:145)는 셀렉터 0매치=초록 반패턴의 직접 차단이고 plan §G·M4에도 이중 명시. 뮤턴트 규율(AC-SA-011)은 verification-completeness §2 + 프로젝트 교훈(「못 잡은 뮤턴트도 보고」)을 정확히 이행. B4의 2단 RED(유도 지점 고정 → 착지 단언)는 미확정 Gap을 실행 가능한 계획으로 바꿔놓았다. 감점: AC-SA-006 골든 코퍼스가 `project_dir ≠ current_dir` 입력을 요구하지 않음(D2 참조). |
| Traceability | 0.75 | §D.2 대응표 12행을 판정서 전문과 대조 — 전 행이 실재하는 판정서 근거로 추적된다(B-1/B-2/B-3, A-5/A-6, Gaps, Residual-risk 1-3, 수리 방향 1/3, 카드 [HARD]). 역방향도 성립: A-1..A-7·B-1..B-3·Gaps·Residual 3건·수리 방향 3건 전부 SPEC 어딘가에 착지한다(D12←A-7, §5←A-4+Claim 4, M5←수리 방향 3, D7←카드 [HARD]). 감점: 판정서 **부칙(B-표 v2, verdict.md:195-216)** 이 열거를 증보한 B2b(github counts — "R1 단일 시접의 적용 대상"으로 명시)와 B7(session-memo — "run-phase에서 범위 재판정")을 SPEC이 전사하지 않았다(D3 참조). |

종합: (0.75 + 1.00 + 0.90 + 0.75) / 4 = **0.85**

## Defects Found

**D1 — 표시 경로 서술이 트리와 반대다 (차단)** — spec.md:93 (REQ-SA-004) · plan.md:48 (D2) · plan.md:85 (M1 3번 bullet) · acceptance.md:104,166 (AC-SA-006/012) — Severity: major — Class: **blocking**
REQ-SA-004는 표시 이름 유도가 "세션의 current_dir에서 계속 유출된다"고 서술하고 D2는 이를 재논의 금지 구속 조건으로 고정했다. 트리 관측: `extractProjectDirectory`(internal/statusline/builder.go:415-438)의 우선순위는 **project_dir > current_dir > CWD > Getwd**이고 416-418행이 `ProjectDir`을 1순위로 반환한다. `WorkspaceInfo.ProjectDir` 주석(types.go:184)도 "(used for display)"다. 이 오류는 판정서 수리 방향 1("표시용 basename은 종전대로 current_dir에서")에서 왔고 SPEC이 전사했다 — 불변성이라는 **결론은 옳지만 논거가 거짓**이다. 글자대로 이행하면 수리 담당자가 표시 경로를 current_dir 소스로 *변경*하게 되는데, 이는 REQ-SA-004 자신의 의도(무변경)가 금지하는 행위고, cd한 세션에서 실제 표시 변화를 일으킨다(project_dir basename → current_dir basename).
**필수 수리**: 네 표면을 "표시 유도는 **기존 `extractProjectDirectory` 동작(project_dir > current_dir > CWD > Getwd) 그대로 — 수리가 이 함수를 수정하지 않는다**"로 고친다. "current_dir에서 유도"라는 메커니즘 서술을 삭제한다. 판정서 산문과의 불일치이므로 SPEC에 정정 사실을 한 줄 기록한다(트리 관측이 판정서 산문에 우선 — 본 SPEC의 전사 원칙상 정정은 예외가 아니라 의무다).

**D2 — AC-SA-006 골든 코퍼스에 발산 입력이 없다 (차단, D1과 동일 근원)** — acceptance.md:100-108 — Severity: major — Class: **blocking**
골든 diff는 "동일 stdin 입력, 수리 전후 출력 동일"인데, 표시 소스가 실제로 갈라지는 지점은 `project_dir ≠ current_dir` 입력(AC-SA-001이 쓰는 A/B 2디렉터 형태)뿐이다. 코퍼스가 동일-디렉터 입력만 담으면, D1의 글자대로 표시를 고친 이행도 골든 테스트를 통과한다.
**필수 수리**: AC-SA-006의 Given에 "project_dir(A) ≠ current_dir(B)인 입력을 포함한다"를 추가한다 — AC-SA-001과 같은 fixture를 재사용하면 1줄이다.

**D3 — 부칙 증보 멤버 B2b·B7 미전사** — spec.md:38-47(멤버 표+경계선언) · spec.md:85(REQ-SA-001 멤버 목록) · acceptance.md:24(AC-SA-012) · acceptance.md:174-191(§D.2) · plan.md §H — Severity: major — Class: **optional (권고)**
판정서 부칙(verdict.md:195-216)은 `state/github/counts.json`(B2b)을 "R1 단일 시접의 적용 대상은 B1·B2(**+B2b**)·B3·B4"로 명시했고 B7(session-memo)에 "run-phase에서 범위 재판정"을 지시했다. SPEC은 B-1 표만 전사하고 부칙을 dropped했다. **행위적 구멍은 아니다 — 코드로 검증했다**: B2b는 독립 앵커가 없다. `builder.go:255`의 `boardRoot := resolveBoardRoot(input)` 하나가 `resolveGitHubCounts`·`maybeRefreshGitHubCounts`(builder.go:265/267)·landed 3형제에 같이 흐르고, `githubCachePath`(github.go:60-62)는 그 boardRoot 아래에만 쓴다. B2 수리가 B2b를 자동으로 운반한다. B7은 `internal/hook/memo/writer.go`로 AC-SA-008의 diff-0 가드가 기계적으로 지킨다.
**권고 수리**: ① REQ-SA-001 괄호 목록과 §1.1 경계 선언에 B2b를 명기("B2 board root — landed·github counts 양 소비자 포함"). ② AC-SA-002의 GREEN에 github counts 경로 단언 1줄(비용 거의 0). ③ plan.md D4 또는 M2에 "B7은 훅 사슬 정리와 함께 run-phase 범위 재판정 대상(판정서 부칙)" 1줄 — 지금은 spec+plan만 읽는 run-phase 담당자가 그 지시의 존재를 알 방법이 없다.

**D4 — 등급 라벨과 두-셀 규율의 긴장** — acceptance.md:5,14-17 — Severity: minor — Class: optional
acceptance.md 자신의 정의로 blocking = "RED-now 셀 + green path 셀을 갖춘 릴리스 게이트"인데 AC-SA-002/003/004/005는 "blocking" 등급에 예정형 RED("(M2/M3 RED 관측 예정)")를 올려놨다. 헤더가 이를 정직하게 공개하고 D8(멤버별 committed RED → RED 출력 §E.2 보존 → GREEN)·E8·마일스톤 종료 조건이 관측을 기계적으로 강제하므로 규율의 실질은 지켜진다 — 채택을 관측 후로 미루는 자세 자체는 verification-completeness §2 적합이다. 라벨만 삐어져 있다. 권고: 등급을 "blocking(RED 관측 전 — 채택은 run-phase)"으로 표기하거나 채택 상태 칼럼을 추가.

**D5 — AC-SA-011 뮤턴트 맥락 형태가 부정확하다** — acceptance.md:152-156 — Severity: minor — Class: optional
Given의 "live-like 저장소 컨텍스트"는 오염을 만드는 형태가 아니다. 가드 판별식을 읽었다: `liveTodoQueueRootReason`(internal/cli/todo_queue_root_test.go:226-234)은 **루트가 OS temp 트리 안이면 침묵**하고, 그 외(라이브 리포, 실제 HOME 폴백)엔 발화한다. 판정서 A-5의 기전 사슬대로 canary HOME 오염은 `CLAUDE_PROJECT_DIR` = **비git temp 디렉터** + canary HOME에서 난다 — "live-like 저장소"(git repo) 맥락이면 git 해석이 fixture 리포로 가서 HOME 오염이 안 난다. "못 잡은 뮤턴트 보고" 조항이 mis-shaped 시도를 흡수하니 판별 가능성은 보존되지만, Given이 정확한 형태를 이름으로 쓰는 것이 뮤턴트-증명의 취지에 맞다. 권고: Given을 "CLAUDE_PROJECT_DIR가 비git temp 디렉터를 가리키는 컨텍스트(verdict A-5 사슬 형태)"로 교정.

**D6 — M2 "기존 테스트 무파괴"와 의도된 변경의 충돌 여지** — plan.md:97 vs acceptance.md:76 — Severity: minor — Class: optional
B3 수리는 cd-기반 goal 읽기에 의존하던 기존 테스트를 *갱신*할 수 있는데(의도된 변경), M2 종료 조건의 "기존 테스트 무파괴"가 갱신을 허용하는지 모호하다. 또 그 확인 범위는 이번 감사에서 좁혀질 수 있다: goal 읽기의 소비자는 `builder.go:286`(생산 1곳) + `handoff_goal_suppress_test.go` + `profile_bench_test.go:256/357`이 전부다. 권고: M2에 "의도된 변경으로 갱신되는 테스트는 §E.2에 목록화, 무파괴는 '미갱신 테스트 무실패'로 정의" 1줄 + 위 3 좌표를 체크리스트로 명기.

**D7 — 형식 비고 묶음** — Severity: minor — Class: optional
① REQ-SA-005(spec.md:94)는 (Unwanted) 라벨에 "shall preserve" 긍정형 — 문법은 Ubiquitous 적합이라 MP-2 위반은 아니나 라벨이 어긋난다. ② REQ-SA-011(spec.md:101) 라벨 "Event-detected"는 5 패턴 정식명이 아니고(구조는 Event-driven), 둘째 shall절("a mutant … shall be reported")이 무조건절로 한 REQ에 합성돼 있다 — 분할하거나 유지해도 문법 위반은 아니다. ③ AC-SA-011이 뮤턴트 헬퍼 "관측 후 제거"를 허용 — 관측 기록이 GREEN 내용이므로 허용 가능하나, 프로젝트 교훈(「못 잡은 뮤턴트도 남긴다」)의 강한 형태는 `...GuardBypassMutant` 이름으로 트리에 남기는 쪽이다. ④ 승계 인용의 줄번호 미세 드리프트: 판정서 B2행 "호출 builder.go:274"의 실제 `resolveBoardRoot` 호출은 builder.go:255(274는 `resolveLandedCounts`) — 실질 사슬은 정확. ⑤ 그 외 검증된 정밀 일치: context_usage.go:176/278, builder.go:178/286, backlog.go:24, landed.go:78/118, cache.go:58, manager.go:74, types.go:184, todo_test.go:42, todo_root.go:96, memory.go:73, registry.go:174, memo/writer.go:12.

## 관측 사항 (결함 아님)

- **lint CoverageIncomplete 11건은 Tier M false positive다** — `moai spec lint`가 spec.md만 스캔해 REQ→AC 참조를 못 보는데, Tier M에서 AC는 acceptance.md에 살고 §D 매트릭스가 11 REQ 전부를 참조한다(AC-SA-006|REQ-SA-004 등). 매 SPEC마다 재발할 도구 맹점 — `/moai:feedback` 후보(본 감사는 read-only라 제출하지 않음, 오케스트레이터 라우팅 몫).
- **"가드 표면 2 헬퍼" 주장 정확** — `runTodo`(todo_test.go:42-46) + `runTodoWithClosedStdin`(todo_relate_test.go:319, 내부 321행 자체 게이트) = 정확히 2. `newTodoCmd()` 직접 Execute 우회 시 무가드 전제도 성립(runTodo가 게이트 후 `newTodoCmd()`를 호출하는 구조). REQ-SA-011의 뮤턴트 전제 유효.
- **B4 전제의 간접 확인** — 내 판독: `ConfigManager.Load(projectRoot)`(manager.go:63-74)가 `configDir = <projectRoot>/.moai` + `MOAI_CONFIG_DIR` override로 `LoadWithCache`에 넘긴다. cwd 연결은 projectRoot를 넘기는 호출자 쪽에 있다 — M3 1단 RED가 고정할 지점이 실재하며, GH #1694 부칙의 150건 `config-cache.json` 계수가 생산 cwd-앵커링의 외부 필드 증거다. AC-SA-004의 **결과** 명제(캐시=프로젝트 앵커, cwd 무오염)는 유도 지점이 어디로 밝혀져도 안정적이라 2단 설계는 견고하다.
- **판정서 커밋 검증** — `e7a078970` "test(cli): fail-loud guard — todo tests must isolate the queue root (card t422)" 2026-09-02 03:13, todo_queue_root_test.go +96 / todo_test.go +28 / todo_relate_test.go +7 — 판정서 주장과 일치. (커밋 메시지 본문 인용은 verbatim 확인 대상에서 제외 — stat과 가드 코드가 실질을 corroborate.)
- **plan-phase 산출물만 존재** — base 대비 diff가 `.moai` 문서 6파일뿐이므로 "SPEC이 스스로 구현하지 않는다" 확인. 카드 추적성: 브랜치 커밋 2건 모두 메시지에 `card t510` 명기(§3 운반체 요건 충족).

## Regression Check

N/A — iteration 1.

## Residual Risks (수용)

1. **REQ-SA-002 우선순위 vs 워크트리 세션의 `project_dir` 의미론** — Claude Code가 워크트리 세션에서 `project_dir`에 워크트리 루트를 담는지 원래 체크아웃을 담는지 미관측이다. project_dir가 1순위면 워크트리 세션은 git common-dir 해석(「모든 체크아웃에 하나의 루트」)에 도달하기 전에 앵커가 워크트리로 굳을 수 있다. 폴백 3단이 있어 기능 파열은 없고, 체인 순서는 D1으로 판정서 전사가 맞다 — 다만 M1/C0에서 실제 페이로드를 한 번 관측하면 이 리스크가 소멸한다. 수용 사유: 외부 런타임 동작이라 plan-phase에서 검증 불가.
2. **statusline stdin에 `project_dir` 실재 여부**(판정서 Gaps와 동일) — 체인이 폴백을 갖춰 B1 재현·수리에 무관.
3. **무프로젝트 skip의 사용 빈도 미측정**(SPEC §7 자기 공개) — REQ-SA-003 설계 판단은 판정서 수리 방향 1의 직접 전사로 수용.

## Recommendation (iteration 2 delta 재감사 범위)

1. **[차단 D1]** REQ-SA-004(spec.md:93)·plan.md D2(:48)·plan.md M1(:85)·acceptance.md AC-SA-006(:104)/AC-SA-012(:166)의 표시-소스 서술을 "기존 `extractProjectDirectory` 동작 불변 — 수리가 이 함수를 수정하지 않는다"로 교정 + 판정서 산문 정정 사실 1줄 기록.
2. **[차단 D2]** AC-SA-006 Given에 `project_dir ≠ current_dir` 입력 포함 조항 추가.
3. **[권고 D3]** B2b를 REQ-SA-001 괄호 목록·§1.1 경계에 명기, AC-SA-002 GREEN에 github counts 경로 1줄, plan.md에 B7 run-phase 재판정 지시 1줄.
4. [선택 D4-D7] 등급 표기 정밀화, AC-SA-011 Given 교정, M2 무파괴 정의 + goal 소비자 3 좌표 명기, REQ-SA-005/011 라벨 정리.

1·2가 반영되면 iteration 2에서 이 두 결함의 해소만 재판정한다(나머지는 optional로 오케스트레이터 재량).

---

## 측정 로그 (본 감사가 실행한 명령)

| 명령 | 관측 |
|---|---|
| `git branch --show-current` / `git rev-parse --short HEAD` | `WT-state-write-locus` / `668b10721` |
| `git log --oneline -3` | 668b10721(SPEC 산출물) ← 17257d9af(판정서) ← 0b1e27877(=base) |
| `git diff --stat 0b1e27877..HEAD` | `.moai` 6파일 752 insertions — 코드 0 |
| `git show --stat e7a078970` | t422 가드 커밋 실재·내용 일치 |
| `moai spec lint .moai/specs/SPEC-STATE-ANCHOR-001/spec.md` | 0 error, 11 warning(CoverageIncomplete — Tier M false positive, 위 관측 참조). 비고: `moai spec lint SPEC-STATE-ANCHOR-001`(ID 인자)은 ParseFailure — t500에서 이미 알려진 CLI 인체공학 결함 |
| `grep -c "NEEDS CLARIFICATION"` plan/spec/acceptance | 0 / 0 / 0 |
| `grep -c syscall spec.md` | 0 |
| `grep "^status:" .moai/specs/SPEC-SESSION-TELEMETRY-001/spec.md` | `status: completed` |
| `grep -rn "audit_model" .moai/config/` | 0매치 — cross-model 팬아웃 미설정, Claude-only 감사 경로 |
| 코드 판독(Read) | builder.go(160-310, 400-454), context_usage.go(160-292), backlog.go(1-47), github.go(전체), landed.go(70-124), todo_test.go(1-90), todo_queue_root_test.go(210-289), todo_root.go(55-134), cache.go(1-80), manager.go(40-99), memory.go(60-104), types.go(165-204), registry.go(165-189), hook/memo/writer.go(1-25) |

감사자: plan-auditor (독립 판정 — 작성자 추론 맥락 없이 산출물과 트리만으로 판정. M1 Context Isolation 준수)

---
---

# Plan-Audit — SPEC-STATE-ANCHOR-001 · Iteration 2 (delta 재판정)

Iteration: 2/2 (Tier M ceiling = 2 — 최종 iteration)
Verdict: **PASS**
Overall Score: **0.98** (Tier M threshold 0.80 — 통과. iter-1 0.85 대비 상승 — LEAN score-regression STOP 신호 없음)
측정 트리: `.claude/worktrees/t510` @ `4e89df2cb` (v0.1.1 커밋 "docs(t510): verdict correction — display-path prose inverted vs tree (plan-audit iter-1 D1+D2, card t510)"). base `0b1e27877` 대비 diff는 `.moai` 문서 6파일뿐 — **코드 변경 0** 확인.
범위 규율: iter-1 차단 D1+D2의 해소 + advisory D3-D6의 착지 확인만. 그 밖 재판정 없음(델타 판정).

## 차단 결함 재판정

| 결함 | 판정 | 근거 |
|---|---|---|
| **D1** 표시 경로 오지시 | **해소** | 6표면 전부 트리-정확 서술로 교정: REQ-SA-004(spec.md:98 — "existing `extractProjectDirectory` behavior (`builder.go:415-438`: `project_dir` > `current_dir` > `CWD` > `os.Getwd()`) untouched, and the repair shall not modify that function"), plan D2(plan.md:48), plan M1(plan.md:85), AC-SA-006 Then(acceptance.md:105), AC-SA-012(acceptance.md:167 — "current_dir는 기존 표시 유도(extractProjectDirectory — 불변)에만 남는다" — 수리 후 참이 된 서술), + 자발 발견 2건(spec.md:79 §3 mutant 2, plan.md:129 §G). 정정 기록 1줄(spec.md:96) 착지. **판정서 정정도 실측 확인**: verdict.md:222-229에 원문 보존 + 병기 정정 부칙("트리는 반대다…원문은 삭제하지 않고 이 정정과") — 「결론은 옳고 논거를 정정」 형태로 프로젝트 교훈을 그대로 이행. |
| **D2** 골든 코퍼스 발산 입력 | **해소** | AC-SA-006 Given에 "골든 코퍼스에 `project_dir`(A) ≠ `current_dir`(B)인 입력이 포함된다(AC-SA-001과 같은 2디렉터 fixture 재사용)"(acceptance.md:103) + 판정 명령에 "발산 입력 골든 비교 1건 추가"(:107) + GREEN에 "발산 입력 포함 코퍼스 PASS"(:109). 처방 그대로. |

**차단 잔존: 0건.**

## Advisory 착지 확인

| 항목 | 착지 | 근거 |
|---|---|---|
| D3 B2b/B7 전사 | **예** | REQ-SA-001 멤버 목록에 "B2 board root — covering both its landed and github-counts consumers"(spec.md:88) · §1.1 부칙 전사 문단(spec.md:50 — B2b 독립 앵커 없음 + B7 run-phase 재판정) · B2 행에 부칙 비고 + iter-1의 줄번호 드리프트(builder.go:274→255)까지 정정(spec.md:44) · AC-SA-002에 github 캐시 경로 단언 And-절(acceptance.md:59 — "github 캐시를 visited 디렉터나 별도 위치에 쓰는 수리는 실패다") + GREEN 셀(14) · plan M2에 B2b 운반 확인 + B7 재판정 지시 흡수(plan.md:95-96). |
| D4 등급 라벨 | **예** | 「blocking (RED 관측 전 — 채택은 run-phase)」 라벨 + 용어 단락 추가(acceptance.md:5, AC-SA-002~005 행 :14-17) — §D8·E8 기계 강제와 정합. |
| D5 뮤턴트 맥락 | **예** | AC-SA-011 Given이 "`CLAUDE_PROJECT_DIR`가 비git temp 디렉터를 가리키는 컨텍스트(판정서 §A-5 기전 사슬)"로 교정 + git-repo 맥락이 오염을 안 만드는 경계 설명 동봉(acceptance.md:155). |
| D6 M2 무파괴 정의 | **예** | 「무파괴 = 미갱신 테스트 무실패, 갱신분은 §E.2 목록화」(plan.md:99) + goal 소비자 3좌표 체크리스트(builder.go:286 / handoff_goal_suppress_test.go / profile_bench_test.go:256-357)(plan.md:97). |
| D7 형식 비고 | **미반영 (수용)** | REQ-SA-005 (Unwanted) 라벨·REQ-SA-011 (Event-detected) 라벨+합성절·뮤턴트 제거 허용 — 전부 iter-1에서 minor/optional로 분류한 재량 항목. MP-2 위반 아님(문법은 5패턴 적합). 잔존하지만 판정에 영향 없음. |

## 새 결함 스윕 (수정 영역)

수정된 6표면 + 신규 문단에서 새로 생긴 허위 주장 0건. delta 범위의 유일한 미검증 주장이었던 「판정서에 정정이 부쳐졌다」(spec.md:96)를 verdict.md:222-229 직접 판독으로 검증했다. 부수 확인: HISTORY 0.1.1 행 정확(spec.md:27), frontmatter version 0.1.1, lint 재실행 **0 error / 11 warning**(동일한 Tier M CoverageIncomplete false positive — 도구 맹점, 변화 없음).

## Category Scores (iter-2)

| Dimension | Score | 변동 근거 |
|---|---|---|
| Clarity | 1.00 | 허위 메커니즘 서술 소멸 — 전 REQ 단일 해석. |
| Completeness | 1.00 | 불변. |
| Testability | 0.95 | 발산 입력 코퍼스로 AC-SA-006 강화. 잔여: 뮤턴트 제거-허용(보존 선호는 optional 잔존). |
| Traceability | 0.95 | 부칙 B2b/B7이 경계·REQ·AC·M2에 전사. 잔여 nit: §D.2 표의 AC-SA-002 행은 부칙 인류를 갱신하지 않았으나 AC 본문이 인용해 추적성은 완결(분산 존재). |

종합: (1.00 + 1.00 + 0.95 + 0.95) / 4 = **0.975 → 표기 0.98**

## 잔여 리스크 (iter-1에서 수용한 것 변동 없음)

워크트리 세션의 `project_dir` 의미론 미관측(폴백 3단 존재 — M1/C0에서 한 번 관측 권고), statusline stdin의 `project_dir` 실재 여부(외부 런타임), 무프로젝트 skip 빈도 미측정(SPEC §7 자기 공개).

## 판정

**PASS — run-phase 진입 가능.** 반복 소관 종료: Tier M ceiling 2에 도달했고 최종 판정은 PASS다. Implementation Kickoff Approval(plan→run HUMAN GATE)은 이 PASS로 자동 우회되지 않는다 — 별도 게이트로 유지된다. 남은 D7 형식 비고 3건은 오케스트레이터 재량 optional이며 수리 의무가 아니다.

감사자: plan-auditor (iter-2 delta 재판정 — iter-1 기록은 상단에 보존됨)
