# SPEC Review Report: SPEC-LEARN-CHANNEL-SCOPE-001

Iteration: 1 (Tier S — `plan_audit_tier_ceilings` S=1, 단일 전수 감사)
Verdict: **FAIL**
Overall Score: **0.63** (harmonic mean — Tier S PASS threshold 0.75 미달)
측정 트리: worktree `WT-learn-channel-gap` @ `d7ce6c6bd` (본 감사 전 수행 `git rev-parse` 관측값, 2026-09-01)

> 요약 한 줄: 카드의 전제(범위 갭은 실재하고 인간 루프가 실제 학습 채널이다)는 실측으로 재확인됐다. FAIL의 원인은 방향이 아니라 **SPEC 자신의 구조적 사실 주장이 틀렸기 때문**이다 — 인박스 writer는 `tool_failure` 단일 패밀리가 아니라, 배선된 제2 패밀리 `test_fail:<pkg>:`를 함께 보유한다. 정직화를 주장하는 SPEC이 스스로 과소-기술(과대가 아니라 반대 방향의 부정확)을 저질렀다.

---

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성**: `grep -on 'REQ-LCS-[0-9][0-9][0-9]'` → spec.md:83-89 연속 001-007, 갭·중복 0. AC도 151-157 연속 001-007 (§F의 111-114는 교차참조). `spec.md:L83-89`, `spec.md:L151-157`
- **[PASS] MP-2 GEARS 형식 (요구 레이어 기준)**: 7개 REQ 전부 패턴 적합 — REQ-001/002/003 ubiquitous ("The harness documentation shall…"), REQ-004/007 Where-절, REQ-005 unwanted ("shall not"), REQ-006 state-driven ("While … shall remain"). 판정 레이어: **요구 레이어(spec.md §C)** — §I의 AC는 Given-When-Then 검증 레이어로 본 기준의 대상 아님(M3 § Scope). `spec.md:L83-89`
- **[PASS] MP-3 YAML frontmatter 유효성**: 12 정본 필드 전부 존재 + 올바른 타입 (`spec.md:L2-16`). snake_case 별칭 0. `phase: "v3.2.0 target"` — 금지 전체값(plan/run/sync/mx) 아님. `era: V3R6` 명시(H-override 유효). `tier: S`, `related_specs` 선택 필드.
- **[N/A] MP-4 Section 22 언어 중립성**: 단일 도메인(문서 정밀화) SPEC — 다중 프로그래밍-언어 도구 내용 없음. N/A 자동 통과.
- **[PASS] MP-5 D7 교차-SPEC 조정**: 참조 2건 모두 존재 + `status: completed` (SPEC-LSEL-DRAIN-STALL-001, SPEC-LSEL-LOCAL-EVOLUTION-001 — frontmatter awk 판독, 본 감사 실행). retired/superseded/archived 없음 → 조정 요구 없음. BLOCKING finding 없음.
- **[PASS] MP-6 D8 크로스플랫폼**: `grep -c 'syscall'` → **0** → D8 자동 통과.
- **[PASS] MP-7 clarification 게이트**: `grep -rn 'NEEDS CLARIFICATION'` (spec.md/plan.md/progress.md) → **0매치**. research.md 없음(Tier S 정상 — 2-artifact 세트) → N/A.

**M5 방화벽: 전부 통과.** FAIL은 방화벽이 아니라 루브릭 점수(하모닉 평균 0.63 < 0.75)와 차단 결함 D1-D3에서 나온다.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 (1-2개 요구의 경미한 모호성) | AC-LCS-004 범위 서술이 REQ-LCS-004 트리거보다 넓다 (D3). `spec.md:L86` vs `spec.md:L154` |
| Completeness | 0.50 | 0.50 (핵심 근거의 실질 공백) | §B.1/§B.2/§H가 제2 writer 패밀리를 누락 — 정직성 SPEC의 증거 절 자체가 부정확 (D1). `spec.md:L47,L64,L142` |
| Testability | 0.50 | 0.50 (올바른 구현으로도 유지 불가한 AC 1건) | AC-LCS-001의 영구 불변식이 무관 상류 이벤트에 의존 (D2). `spec.md:L151` |
| Traceability | 1.00 | 1.00 | 7 REQ ↔ 7 AC 쌍방향 완전, 고아 0, M1-M3가 전 AC 커버 (plan.md §F: M1→AC-001, M2→AC-002/003/007, M3→AC-004/005/006) |

**하모닉 평균**: 4 / (1/0.75 + 1/0.50 + 1/0.50 + 1/1.00) = **0.63**

---

## 독립 재측정 (본 감사가 실제로 관측한 것 — 귀속 명시)

| 대상 | 명령 | 관측값 (본 감사, 2026-09-01, worktree `d7ce6c6bd` / primary 런타임 파일) |
|---|---|---|
| live 인박스 구성 | `jq -r '.event_key' <primary>/.moai/lessons-inbox.jsonl \| cut -d: -f1 \| sort \| uniq -c` | **5,932행 전부 `tool_failure`** — 오케스트레이터 5,916 → 저자 5,919 → 본 감사 5,930/5,932 (append-only 성장, 구성 결론 3중 일치) |
| 패밀리 상위 | `jq -r '.event_key' … \| sort \| uniq -c \| sort -rn` | `tool_failure:Bash:UnknownFailure` 5,089 외 전부 `tool_failure:*` — **`test_fail` 0행** |
| 인간 채널 | `find <memory-store> -maxdepth 1 -name 'feedback_*.md' \| wc -l` (+`-newermt 2026-08-25`) | **164 / 145** — SPEC §B.1 값과 일치 |
| constitution-detail L53 과대문구 | `sed -n '48,58p'` 로컬+미러 | "tool failures **and test failures** append structured stubs…" — 양쪽 바이트 동일 존재 확인 |
| writer 구조 | `grep -n` + Read `internal/hook/failure_observer.go` | **stub 생산자 2개 확정** — `:77` (`tool_failure:<tool>:<sig>`) + `:111` (`test_fail:<pkg>:`) |
| 제2 패밀리 배선 | Read `internal/hook/evidence_writer.go:560-592` | `logEvidence` PostToolUse → `rec.IsTestFail` → `recordTestFailEvent` (L589) — **살아있는 호출 경로**. 도입: `e70c77576` (SPEC-HARNESS-RATCHET-REWIRE-001 M1) |
| 표면 전수 | `grep -rln 'lessons-inbox'` (worktree, specs/reports/.git 제외) | 24파일 — 아래 표면 분석 참조 |
| anchor doc 대상 | `ls .moai/docs/learning-channel-scope.md` | 부재 (신규 생성 대상 정상) |
| `.moai/docs` 추적 | `git ls-files .moai/docs \| wc -l` | 21 — tracked + §2.3 삭제 뿌리 밖 (plan.md 주장 확인) |
| 템플릿 AGENTS.md | `grep -c 'LSEL'` | **0** — plan.md §A.4 주장 확인 (CLAUDE.local.md 미러 불필요) |

---

## 표면 전수 분석 (감사 항목 4 — plan이 빠뜨린 표면)

`grep -rln 'lessons-inbox'` 24파일을 성격별로 분류:

| 분류 | 파일 | plan 목록 여부 | 판정 |
|---|---|---|---|
| 범위 주장 표면 | `moai-constitution-detail.md` (로컬+미러), `hns-lsel-curator/SKILL.md`, `CLAUDE.local.md` §28 | ✅ 3건 모두 목록됨 | REQ-LCS-004 트리거("describes") 적중 — plan의 주 대상 정확 |
| 배제 선언 표면 | `moai-workflow-project/references/navigator.md:86,208` + `navigator-audit.md:215-218` (양쪽 모두 **템플릿 미러 존재**), `hns-lsel-applier/apply_test.sh`, `.claude/lsel/frozen-allowlist.json` | ❌ 미목록 | 인박스의 캡처 범위를 말하지 않고 **Navigator 자신의 읽기-집합 배제**를 말함 → REQ-LCS-004 트리거 비적중. 그러나 AC-LCS-004의 문자적 범위("all prose surfaces referencing")는 이들을 집어넣는다 → **D3의 실체** |
| 스테일 라이브 수치 | `hns-lsel-curator/SKILL.md:34` — "(624 stubs at M1 start…)" | 부분 (파일은 목록, 수치 미언급) | 실제 5,932행 — 본 SPEC 자신의 제약 7(라이브 수치는 anchor doc에만) 위반 상태. M2 편집 범위에 명시 안 됨 → **D4** |
| 코드/테스트 | `failure_observer.go`, `lessons_inbox_test.go`, `drain.sh` 외 curator 스크립트, navigator 테스트 4건, `lsel-drain-loop.js` | — | 산문 표면 아님. §G 무수정 범위와 일치 |
| 이력 | `CHANGELOG.md` | — | 이력 기록 — 대상 아님 |
| 본문(항상 로드) | `moai-constitution.md` | (무수정 결정) | 리터럴 `lessons-inbox` 토큰 0히트 — 본문은 "repo-local inbox drain contract" 포인터만 두고 상세를 companion에 위임. **무수정 결정은 방어 가능**함을 본 감사가 확인 |

---

## Defects Found (structured defect-list)

**D1.** LCS-WRITER-EXCLUSIVITY — `spec.md:L47` (§B.1 구조 근거 행), `spec.md:L83` (REQ-LCS-001), `spec.md:L64` (§B.2 판정), `spec.md:L120` (§G 예시), `spec.md:L142` (§H) — **"writer는 tool_failure 계열만 배출/유일 배출/records tool-call failure stubs only"는 구조적으로 거짓.** 실제: `internal/hook/failure_observer.go:87-112` `recordTestFailEvent`가 `test_fail:<pkg>:` 키로 같은 인박스에 stub을 append하고(:110-111), `internal/hook/evidence_writer.go:583-591`의 `logEvidence` PostToolUse 경로에서 `rec.IsTestFail`일 때 실제 호출되며, `e70c77576`(SPEC-HARNESS-RATCHET-REWIRE-001 M1)에서 배선됐다. stub 생산자는 정확히 2개(:77, :111). §B.2의 "미구현 서술" 가설은 거짓(구현돼 있음)이고, §G의 "test_failure 패밀리 추가" 예시는 이미 트리에 존재하는 기전을 미래 기전처럼 쓴다. live 파일에 `test_fail`이 0행인 것은 **경험적 사실**(3개 baseline 일치)이지 구조적 사실이 아니다 — 두 진술을 구분하지 않은 것이 결함의 핵심. — Severity: **critical** — Class: **blocking** — Required fix: 경계 주장을 **구성 기반 + 능력 정직형**으로 재프레임: (a) 측정된 구성 100% `tool_failure` (dated baseline + 재검증 명령), (b) writer 표면에는 배선된 제2 패밀리 `test_fail`이 존재하며 본 파일에는 0행, (c) 어느 패밀리도 도구-비가시 결함 계열(공허 초록·스테일 귀속 등)은 담지 않는다 — 그것들은 인간 루프로 흐른다. REQ-LCS-001/§B.1/§B.2/§G/§H 5개 표면 동시 수정.

**D2.** LCS-AC001-FRAGILE-INVARIANT — `spec.md:L151` (AC-LCS-001) — "the fresh tally again shows **zero non-`tool_failure:` rows**"는 올바른 구현이 영구히 유지할 수 없는 판정식이다. 제2 패밀리가 배선돼 있으므로(D1), 어느 세션의 실패한 테스트가 정상 경유로 `test_fail` 행을 하나 append하는 순간 재검증이 적색으로 뒤집힌다 — 본 작업과 무관한 상류 이벤트가 green을 결정하는 구조(verification-completeness §2의 green-path 실격 사유). — Severity: **major** — Class: **blocking** — Required fix: 판정식을 패밀리-집합 형태로 — "zero rows **outside the `{tool_failure, test_fail}` families**" — 또는 D1의 재프레임 문구(구성 주장 + 날짜 스코프)와 정렬.

**D3.** LCS-AC004-SCOPE-OVERREACH — `spec.md:L154` (AC-LCS-004) vs `spec.md:L86` (REQ-LCS-004) — AC 범위 "all prose surfaces **referencing** `.moai/lessons-inbox.jsonl`"이 REQ 트리거 "Where a prose surface **describes** the lessons-inbox"보다 넓다. 문자 집행 시 navigator 배제 선언(로컬+**템플릿 미러** 4파일), `apply_test.sh`, `frozen-allowlist.json`까지 편집 대상으로 끌어들여 Template-First 캐스케이드(미러 쌍 + make build)를 유발한다 — 문서 전용 제약과 충돌. 또한 AC가 참조하는 "the grep sweep list recorded in plan.md"가 plan.md에 열거돼 있지 않다(절차만 있음 — §E.5). — Severity: **major** — Class: **blocking** — Required fix: AC-LCS-004 범위를 "prose surfaces **describing the inbox's capture scope** (REQ-LCS-004의 표면 목록)"으로 정렬 + plan.md §E.5에 스윕 목록을 명시 열거(또는 spec.md §B.2 표 확장판 포인터)해 run-phase 집행자가 구체 대상 집합을 갖게 한다.

**D4.** LCS-STALE-COUNT-IN-PROSE — `.claude/skills/hns-lsel-curator/SKILL.md:L34` — "(624 stubs at M1 start…)" 라이브 카운트가 산문에 박혀 스테일화(실제 5,932). 본 SPEC 제약 7의 교리(라이브 수치는 anchor doc에만)가 자기 표면 목록 안의 기존 위반을 편집 범위에 넣지 않았다. — Severity: **minor** — Class: **optional** — Required fix: plan.md M2의 SKILL.md 항목에 "스테일 stub 카운트 제거/교체(anchor doc 포인터로)" 1절 추가.

**D5.** LCS-BASELINE-SHA-MISSING — `spec.md:L43-56` (§B.1 측정 표) — 명령·출력(요약)·날짜는 있으나 **tree SHA가 없다**(plan.md §A.1의 `d7ce6c6bd` 핀은 plan.md에만 존재). verification-completeness §2.1 4요소 중 2개 미충足 (SHA + exit code). — Severity: **minor** — Class: **optional** — Required fix: §B.1에 문서-수준 핀 1행("본 §B 측정은 worktree HEAD `d7ce6c6bd`에서 수행") 추가 — AC-LCS-001이 anchor doc에 요구하는 것과 동일한 기준을 SPEC 자신이 먼저 충족.

**D6.** LCS-MIRROR-EXCLUSION-LIST-GAP — `spec.md:L152` (AC-LCS-002), `plan.md:L97` (E3) — 미러 배제 목록이 "(dates, counts, SPEC IDs)"로 카드 id(t260)를 빠뜨린다. neutrality C-클래스는 내부 개발 상태 토큰 전반이다. — Severity: **minor** — Class: **optional** — Required fix: 양 목록에 "card ids" 추가.

**D7.** LCS-TWO-CELL-FORMALITY — `spec.md:L147-157` (§I) — AC가 green-form만 carry. RED-now는 실질로 존재(§B.2 인용된 과대문구 + §B.1 baseline)하나 셀 형태가 아니고, AC마다 뒤집는 마일스톤 태그가 없다(plan §F에서 파생 가능). — Severity: **minor** — Class: **optional** — Required fix: 각 AC에 flipping 마일스톤(M1/M2/M3) 태그 + RED 근거 행 인용 1절.

---

## Two-Cell Adoption / Live-File Baseline (감사 지시 1 — 판정)

- **라이브 파일 문제의 설계 대응은 올바르다**: AC-LCS-001이 "재검증 가능성"을 판정하고("the AC judges the claim's re-verifiability, not the count's immutability"), dated baseline(날짜+SHA)+재실행 명령을 anchor doc에 요구하며, 대상을 "primary checkout's live inbox"로 명명한다. 카운트 자체를 불변 조건으로 박지 않은 것은 정확한 선택 — 3개 시점 측정(5,916/5,919/5,932)이 append-only 성장을 보여주며 구성 결론은 불변이다.
- **그러나 D2**: "영구 zero non-tool_failure" 판정식만이 이 올바른 설계 안의 균열 — 배선된 제2 패밀리가 그것을 깬다.
- **RED-now의 실질 충족**: 문서 전용 SPEC의 RED = 측정된 부정확성 주장이며, §B.2가 그것을 파일+인용문으로 고정하고 있다. 4요소 형식 준수는 D5에서 별도(경미).
- **Green path**: plan.md §F가 M1→AC-001, M2→AC-002/003/007, M3→AC-004/005/006으로 매핑 — 실질 충족, 형식 태그는 D7.

## Decision-Axis Integrity (감사 지시 2 — 판정)

plan.md §A.3은 **중립적이다**: 각 축에 권고+귀속된 근거+반전 비용을 나열하고, 최종 결정을 "리드 경유 운영자"(kickoff)에 명시적으로 남기며, 각 축이 뒤집힐 경우의 영향까지 서술한다((a) NO → REQ-002/004/007 폐기 + (c) 단독 축소; (b) 기전화 → 후속 SPEC 분리; (c) NO → 폐기 권고). 측정을 정책 결정처럼 포장한 문구도 없다. REQ-LCS-005의 관측 가능성: AC-LCS-005(docs-only diff, `git diff --stat`) + AC-LCS-006(신규 기전 부재 — `.moai/config/sections/` 신규 0, 훅 배선 0, 스크립트 0)이 이진 grep/find로 검증한다. **단, D1 수정 후 (b)축의 문구도 같이 갱신해야 한다** — "자동 추출 파이프라인은 기전 추가다"라는 골격은 유효하나 그 근거 문장이 D1의 재프레임을 따라야 한다.

## Template-First Exposure (감사 지시 3 — 판정)

plan.md는 사이클 전체를 소유한다: 미러 쌍 수정(§A.4 EXTEND + spec §D 제약 3) + `make build`(M2/E4) + neutrality(E3/AC-002 + 제약 7 "라이브 수치의 미러 유입 금지"). 미러 존재는 본 감사가 직접 확인(`internal/template/templates/.claude/rules/moai/core/moai-constitution-detail.md` 존재, L53 문구 로컬과 바이트 동일). `.moai/docs`가 tracked(21파일) + 삭제 뿌리 밖인 것도 확인 — anchor doc 위치 안전. 템플릿 AGENTS.md LSEL 0회 확인 — CLAUDE.local.md 미러 불필요 주장 참. **갭은 D6(배제 목록에 카드 id 누락) 하나**. 참고: navigator 표면에 템플릿 미러가 존재하므로(D3에서 문자 집행 시 끌려들 파일) AC-004 범위 정렬이 Template-First 노출을 넉아내는 열쇠다.

## Milestone Ordering (감사 지시 5 — 판정)

M1(결정 착지+정준 문구) → M2(표면 전파) → M3(스윕+검증) 순서는 타당하다 — M1이 정준 문구를 고정한 뒤 M2가 전파해야 발산이 안 생기고, M3가 전수 검증한다. "의사결정 역전 가능성 순" 주장과 실제 의존성이 일치. 7 REQ 전부 커버(REQ-005/006은 제약+E2/E6+AC-005/006 경유), 고아 REQ 0, 고아 AC 0. M2의 SKILL.md 항목이 스테일 카운트를 빠뜨린 것만 D4로 별도.

---

## Regression Check (Iteration 2+ only)

해당 없음 — iteration 1 (Tier S 단일 감사). 참고: Tier S 천장이 1이므로 본 보고서의 defect-list가 곧 수정 경로다. 재감사가 필요하면 오케스트레이터가 천장 정책(사용자 선택지: PASS-with-debt / 범위 조정 / 명시 override)에 따라 결정한다.

## Recommendation

**FAIL — 수정 후 재판정 필요.** 카드의 방향은 옳고 전제는 실측으로 살아있다(도구-비가시 결함 계열은 인박스에 0행이고 인간 루프가 주 145건을 생산 — 본 감사 재확인). 수정은 범위 변경이 아니라 **정직성의 양방향 완성**이다:

1. **(D1, 최우선)** writer 단일성 주장 → 구성+능력 정직형 재프레임. 5개 표면(REQ-LCS-001, §B.1, §B.2, §G, §H) 동시 수정. "writer가 배출하는 것"과 "데이터가 보이는 것"을 구분하는 것은 plan.md §G가 이미 가진 정신이다 — 그 정신을 자기 §B에도 적용하는 일이다.
2. **(D2)** AC-LCS-001 판정식을 `{tool_failure, test_fail}` 패밀리-집합 형태로.
3. **(D3)** AC-LCS-004 범위를 REQ-LCS-004의 "describes"에 정렬 + plan.md에 스윕 목록 명시 열거.
4. **(D4-D7)** 경미 4건 — 각 1-2행 수정.

이 3건의 차단 결함이 해소되면 나머지 구조(의사결정 축, 마일스톤, 미러 소유, 라이브 baseline 설계)는 Tier S 기준으로 건전하다.
