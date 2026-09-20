# SPEC Review Report: SPEC-WORKTREE-EXIT-RETURN-001

Iteration: 2 (Tier S 반복 상한은 1 — 이 회차는 오케스트레이터 지시로 상한 밖에서 수행됨. § 절차 기록 참조)
Verdict: **FAIL**
Overall Score: **0.80** (조화평균) — Tier S PASS 문턱 **0.75** 를 **넘는다**

> **이 FAIL 은 점수가 만든 것이 아니다.** 집계 점수는 문턱을 넘었고 iter-1 대비 모든 축이 올랐다.
> FAIL 을 만든 것은 **새로 생긴 critical 결함 하나**(D1′)이며, 아래에서 변이 실험으로 기계 입증한다.
> 근거와 그 판정 규칙은 § Verdict 근거 에 명시했다.

> **Reasoning context ignored per M1 Context Isolation.** 배차문이 서술한 저자 측 수정 보고 —
> 「비대칭 결정으로 D1 이 해소됐다」·「D5 의 공허성」·「D9 분할」 — 는 전부 **주장으로 취급**했고
> 파일과 실행에서 재측정했다. 배차문에서 받아들인 것은 감사 방향뿐이다.

## 측정 환경 (baseline attribution)

| 항목 | 값 | 어떻게 얻었나 |
|---|---|---|
| 트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t965` | `git rev-parse --show-toplevel` |
| 브랜치 | `WT-exitworktree-return` | `git branch --show-current` |
| 감사 시작 HEAD | `ce2e35355a0df4535505df22e05b1d0e5fd62e52` | `git rev-parse HEAD` |
| 감사 종료 HEAD | `ce2e35355` — **이동 없음** | 동일 명령 재실행 |
| 측정 일자 | 2026-09-20 | |
| 미변경 대조 | 감사 전후 `.claude/rules/…/worktree-integration.md` 와 그 미러의 sha256 = `3ac8498d62e37445b4345a62f8c8e7610b53fd933bbb09cd32b0bbd6d896c6e4` (양쪽 동일) | `shasum -a 256` |

**외래 커밋 검증 (배차 지시 — 믿지 않고 직접 측정).** `git log --format='%H %ci %s' develop..HEAD` 은
정확히 두 커밋을 낸다: `d20a9f022`(23:47:12) 와 `ce2e35355`(23:50:50). iter-1 판정서 파일의 mtime 은
`2026-09-19 23:58:52` 이므로 **외래 커밋은 iter-1 판정서가 쓰이기 전에 착지했고**, 레인의 자기 보고와
일치한다. **그 뒤로 착지한 커밋은 없다** — HEAD 가 여전히 `ce2e35355` 이고 브랜치 커밋 수는 2 로 동일하다.
작업 트리의 수정은 SPEC 산출물 4개뿐이다(`git status --porcelain`).

---

## Must-Pass Results (전량 재측정 — spec.md 가 개정됐으므로 델타 범위로 줄이지 않았다)

- **[PASS] MP-1 REQ 번호 일관성** — `grep -n '^- \*\*REQ-WXR-' spec.md` → 8행(`:168`,`:177`,`:182`,`:187`,`:192`,`:201`,`:205`,`:213`). `REQ-WXR-001`~`008`, 결번·중복 0, 3자리 제로패딩 일관. 총 8 = Tier S 상한 8(`spec-workflow.md:145-146`) 이내.
- **[PASS] MP-2 GEARS 형식 준수** — **요구사항 층(`spec.md` 의 `REQ-XXX`)에 대해 판정함.** 8개 전부 다섯 패턴 중 하나: 001 Ubiquitous(`:168`), 002 Event-driven(`:177`), 003 Event-driven(`:182`), 004 Ubiquitous(`:187`), 005 Where/capability-gate(`:192`), 006 Unwanted — `shall not` 정칙 부정형(`:201`), 007 While/state-driven(`:205`), 008 Ubiquitous(`:213`). 주어는 일관되게 `the doctrine`. `acceptance.md` 의 Given-When-Then 은 **검증 층**이므로 M3 § Scope 에 따라 여기서 감점하지 않고 Group 4 에서 채점했다. 폐기 예정 `IF/THEN` 0건.
- **[PASS] MP-3 YAML frontmatter 유효성** — 12 canonical 필드 항목별 대조(`spec.md:1-15`): `id`·`title`·`version`(`"0.2.0"` 인용 semver, 0.1.0 에서 상승)·`status`(`draft`)·`created`(2026-09-19)·`updated`(**2026-09-20 으로 갱신됨**)·`author`·`priority`(`P2`)·`phase`·`module`·`lifecycle`(`spec-anchored`)·`tags` 전부 존재·형 일치. 13번째 `tier: S` 는 스키마 등재 선택 필드. 거부 별칭(`created_at`/`updated_at`/`labels`/`spec_id`) 0건.
- **[N/A] MP-4 §22 언어 중립성** — 프로그래밍-언어 도구 이름 0건인 문서-계약 SPEC → 기준 비적용, auto-pass. (별축인 **템플릿 내부-내용 중립성**은 아래 D1′ 에서 다룬다 — MP-4 와 다른 축이다.)
- **[PASS] MP-5 D7 교차-SPEC 조정** — 검증 동사 실행: `grep -hEo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md plan.md acceptance.md | sort -u` → `SPEC-WORKTREE-EXIT-RETURN-001` 하나(자기 자신). 외부 SPEC 참조 0건 → `retired/superseded/archived` 조정 의무 미발생. BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼 규율** — `grep -c syscall spec.md plan.md acceptance.md` → 전부 `0`. D8-4 에 따라 auto-PASS.
- **[PASS] MP-7 clarification 게이트** — `grep -rn 'NEEDS CLARIFICATION' .` → exit 1, 적중 0건. `research.md` 는 Tier S 이므로 부재(정상).

**must-pass 7/7 통과 — iter-1 과 동일.** 이 FAIL 도 방화벽이 만든 것이 아니다.

---

## Category Scores (0.0–1.0) + iter-1 대비 단조성

| 축 | iter-1 | iter-2 | 델타 | 근거 |
|---|---|---|---|---|
| Clarity | 0.75 | **0.75** | ±0 | D4·D7·D8 이 실제로 해소됐다(아래 회귀 검사). 그럼에도 유지한 이유는 **새 모순** 때문이다 — REQ-WXR-008(`spec.md:213-221`)의 비대칭 의무와 §4 「코드 층 무변경」(`spec.md:264-270`)이 이 저장소에서 **동시에 만족될 수 없다**(D1′). 이행자 둘이 서로 다른 쪽을 어기게 되므로 결과가 예측 불가다. |
| Completeness | 0.75 | **0.75** | ±0 | 필수 절 전부 존재(HISTORY `:19`, WHY `:27`/`:47`, WHAT `:89`/`:166`, REQUIREMENTS `:166`, ACCEPTANCE `:224`+`acceptance.md`, Out of Scope `:240` H3 3개 각각 `-` 불릿). D3 해소. 유지 사유: **산출물을 지배하는 가드 하나가 제약 목록에서 빠져 있다** — `plan.md` §D(`:81-89`)·§G(`:117-134`)는 중립성 가드는 세 곳에서 인용하면서 같은 파일에 실제로 걸려 있는 **바이트 동일성 가드**를 한 번도 이름 대지 않는다(D1′). |
| Testability | 0.70 | **0.85** | **+0.15** | AC-003 이 두 기계 검사로 분해됐고(`acceptance.md:59-72`), AC-008 이 공허하지 않은 두 검사를 갖고(`:180-184`), AC-007 (2)(3) 이 기계 판별식이다. AC-007 (3) 은 **변이로 비공허성을 입증**했다(아래 § 변이 실험). 잔여: REQ-001 의 「표본을 문장에 싣는다」와 REQ-005 의 「도구 계약 인용으로 제시하지 않는다」가 어느 판별식에도 걸리지 않는다(D11·D12). |
| Traceability | 0.60 | **0.85** | **+0.25** | 8개 AC 전부 `**Covers**` 로 REQ 를 명시 인용(`acceptance.md:16,34,52,76,97,130,147,174`). `REQ-WXR` 적중 0 → 13. **8개 REQ 전부 커버**(001→AC-001·006, 002→001, 003→002, 004→003, 005→005, 006→005·006, 007→004·005, 008→007). D6 해소. 잔여: AC-008 만 REQ 가 아니라 `plan.md` §C 조항에 앵커돼 있어 엄밀히는 REQ 미앵커 AC 1건. |

조화평균 = 4 / (1/0.75 + 1/0.75 + 1/0.85 + 1/0.85) = 4 / 5.0196 = **0.797 ≈ 0.80**
(산술평균 0.80 으로도 같다. 결론은 평균 방식에 좌우되지 않는다.)

**단조성 판정: 0.69 → 0.80, 네 축 모두 하락 없음.** 점수 회귀는 없으므로 LEAN 워크플로의
STOP 에스컬레이션 조건(iter(N+1) < iter(N))은 **발화하지 않는다.**

---

## 회귀 검사 — iter-1 결함 9건의 현재 상태

| # | iter-1 결함 | 상태 | 증거 |
|---|---|---|---|
| D1 | template-neutrality-collision (critical) | **부분 해소 — 조건이 사라진 게 아니라 다른 가드로 이동** | 미러의 날짜·SPEC ID 는 실제로 제거 설계됐고 부재 검사가 붙었다(`acceptance.md:159-162`). 그러나 그 해법이 **바이트 동일성 가드**를 적색으로 만든다 → **D1′ 로 재기록** |
| D2 | cross-round-version-attribution (major) | **해소** | `acceptance.md:105-112` — (1) 관측 회차 `2.1.278`, (2) 재측정 회차는 AC-004 의 값, 「두 회차의 버전을 하나로 합쳐 쓴 문장은 0건이어야 한다」 명문화 |
| D3 | stale-gap-carriage (major) | **해소** | `spec.md:82-88` 이 「사후 기록됐고 … 사후 귀속으로 약화된 채 남는다」 + 「관측 4 는 그 귀속조차 공유하지 않는다」로 교체됨. 근거 파일 `observations.md:114-116,139-141` 과 문언 일치 확인 |
| D4 | independence-overclaim (major) | **해소** | `spec.md:168-176` 에서 `independent of` 소멸. §2 전체 grep 결과 전칭 술어 적중은 **금지 조항 자신 2곳뿐**(REQ-001 의 금지 목록, REQ-005 의 보장 문언 금지). 표본(깊이 1·2, 단일 출발-디렉터리 사례) 이 문장에 실렸고, 제보와의 공존 문장도 REQ-001 말미에 명시 |
| D5 | ordering-not-witnessable (major) | **해소 — 그리고 iter-1 의 처방이 틀렸다** | 아래 § iter-1 자신의 오류 참조 |
| D6 | uncovered-and-self-voiding-req (minor) | **해소** | REQ-007 가드가 「AC-004 가 수행되지 않은 동안」→「직접 판독으로 덮이지 않은 관측이 있는 동안」으로 바뀌어(`spec.md:205-212`) 재측정으로 자기소멸하지 않는다. AC-004(`:76`)·AC-005(`:97`) 가 Covers 로 덮고, AC-005 (4)(`:116-119`)가 「재측정 완료 뒤에도 통과해야 한다」를 명문화 |
| D7 | ac003-discriminant-not-binary (minor) | **해소** | `acceptance.md:59-72` — (i) 무자격 단독 사용 0건 / (ii) 정정 문장 1건 이상. 두 검사가 같은 줄을 반대 방향으로 분류하므로 한쪽만 만족시키는 이행이 존재하지 않는다는 설명까지 붙었다 |
| D8 | classification-step-unearned (minor) | **해소** | `spec.md:45` 제목이 「이 카드가 고칠 수 있는 것은 계약(문서) 층이다」로 축소. 읽기 (나) 가 이름을 달고 남았으며(`:47-60`) [HARD] 로 「기각하지 않는다」를 못 박고 §4(`:249-253`)에 「기각이 아니라 범위 밖」으로 배치. 「모호성」 격하도 `:104-106` 에서 「관측 4 에서 읽기 하나가 거짓」으로 교정 |
| D9 | untracked-evidence-citation (major) | **해소 — 분할이 건전하다** | 아래 § D9 분할 판정 참조 |

**미해소·화장(化粧)만 한 항목: 없다.** 9건 모두 문언이 실제로 바뀌었고, 각 수정이 자기 판별식을
동반한다. **D1 만 「해소」가 아니라 「이동」이며 그것이 이 회차의 FAIL 사유다.**

---

## 변이 실험 (배차 요청 — 「필요하면 변이를 심어라」)

세 변이를 심고 매번 복원했다. 복원 후 두 사본의 sha256 이 실험 전 값과 동일함을 확인했고
(`3ac8498d…`), `git status --porcelain` 은 SPEC 산출물 4개만 남겼다. 커밋은 하지 않았다.

**baseline (변이 전).** 두 티어 모두 초록:
```
go test ./internal/template/ -run TestTemplateNoInternalContentLeak -count=1          → ok, exit 0
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run … -count=1              → ok, exit 0
```

| # | 심은 것 | 두 사본 관계 | `TestRuleTemplateMirrorDrift` | 중립성 default | 중립성 strict |
|---|---|---|---|---|---|
| V1 | 날짜+내부경로+SPEC ID 를 **로컬에만** (= REQ-WXR-008 이 명령하는 비대칭) | 다름 | **FAIL** `RULE_TEMPLATE_MIRROR_DRIFT` | — | — |
| V2 | 같은 문자열을 **양쪽에** (= 파리티 유지) | 동일 | ok | **FAIL** `class=C1-spec-id-prefix match=SPEC-WORKTREE-EXIT-RETURN-001` | — |
| V3 | 날짜+내부경로+버전(SPEC ID 없음)을 **양쪽에** | 동일 | — | ok | **FAIL** `class=S1-internal-date match=2026-09-19` |

V1 의 원문 출력:
```
rule_template_mirror_test.go:174: RULE_TEMPLATE_MIRROR_DRIFT: source file
.claude/rules/moai/workflow/worktree-integration.md differs from its mirror at
…/internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md
(source 58440 bytes, mirror 58332 bytes); run 'cp …' and stage both files
```

**읽는 법.** V1 과 V2·V3 은 서로 배타다. 파리티를 지키면 누출 가드가 걸리고(V2·V3), 누출 가드를
피하려 비대칭을 만들면 파리티 가드가 걸린다(V1). **현재 SPEC 문언을 만족시키는 트리 상태가
존재하지 않는다.** 이것이 D1′ 이다.

부수 관측 — 무엇이 어느 가드에 걸리는지의 분해:
- **SPEC ID** → `C1-spec-id-prefix`, **default 티어**. strict 를 켜지 않아도 잡힌다.
- **날짜 리터럴** → `S1-internal-date`, **strict 티어 전용**.
- **`.moai/specs/` · `.moai/reports/` 내부 경로** → **어느 누출 클래스도 잡지 않는다**(V3 에서 경로가 있는데 default 초록, strict 은 날짜만 지목). `C5-memory-archive-path` 의 패턴은 `~/.claude/projects/-Users-` 와 `.moai/backups/agent-archive-` 뿐이다. 따라서 AC-WXR-007 (2)(b) 는 **CI 가 받쳐 주지 않는 유일한 부재 검사**이며, 그 자체로 결함은 아니나 잔여 위험으로 기록한다.

**배차가 물은 「AC-WXR-007 부재 검사가 위반 시 실제로 실패하는가」의 답: (2)(a)(c) 와 (3) 은
실패한다(V2·V3 로 입증). (2)(b) 는 AC 자신의 grep 으로만 실패하고 가드는 침묵한다.**

---

## Defects Found (구조화 결함 목록)

**D1′.** mirror-byte-parity-collision — `spec.md:L111-165`(§1.5) + `L213-221`(REQ-WXR-008) + `L264-270`(§4 코드 층) + `acceptance.md:L145-170`(AC-WXR-007) + `plan.md:L37-50`(§B.2b), `L81-89`(§D) — **iter-1 의 D1 은 제거된 것이 아니라 다른 가드로 옮겨 갔다. 이행하면 여전히 완료 시점에 CI 가 적색이 된다.**

측정: 이 SPEC 의 대상 파일 `.claude/rules/moai/workflow/worktree-integration.md` 는
`internal/template/rule_template_mirror_test.go:57` 의 `workflowOptMirroredPaths` **바이트 동일성
허용목록에 명시 등재**돼 있다. 등재 주석(`:53-58`)은 등재 사유를 "Both trees byte-identical
post-neutralization … 앞으로의 단일-트리 편집이 CI 에서 잡히도록 등록"이라고 적는다. 단언부는
`:174` 의 `bytes.Equal(srcContent, mirrorContent)` 이며 실패 시 센티널 `RULE_TEMPLATE_MIRROR_DRIFT`
를 낸다. 등재 전제는 **지금도 살아 있다** — 두 사본의 sha256 이 현재 동일하다(`3ac8498d…`).

즉 REQ-WXR-008 이 명령하는 비대칭(미러에는 날짜·내부 경로·SPEC ID 없음, 로컬에는 있음)은 이 파일에
한해 **기계적으로 금지돼 있다.** 변이 V1 이 이를 입증한다(위 표).

CI 도달성도 측정했다: `.github/workflows/ci.yml:210` 의 `test` 잡이 `go test ./...` 를 돌리고,
그 잡을 여는 `detect` 필터(`:76-95`)의 `go_code` 목록에 **`.claude/rules/moai/**` 와
`internal/template/templates/**` 가 둘 다 들어 있다.** 즉 이 카드의 문서-전용 변경만으로도
`go_code == 'true'` 가 되어 전체 스위트가 돌고, 그 안에서 이 테스트가 적색이 된다. 같은 필터의
주석이 바로 이 위험 계열을 이름 대어 설명한다 — "A PR that moves or edits any of the non-Go roots
below can turn the suite red with zero .go files changed … `internal/template parity tests`".

**SPEC 은 이 가드를 한 번도 이름 대지 않는다.** 측정: `grep -n 'rule_template_mirror|RULE_TEMPLATE_MIRROR|byte-identical|바이트' spec.md plan.md acceptance.md` → 적중 0.
「파리티」 적중 4건은 전부 이 SPEC 자신의 **규범 층 파리티**(판별식 결과 일치) 용법이며
바이트 동일성과 무관하다. `plan.md` §G(`:117-134`)는 중립성 가드와 그 CI 를 세 항목으로 인용하면서
같은 파일에 실제로 걸려 있는 파리티 가드를 빠뜨렸다.

**그리고 처방이 SPEC 자신의 [HARD] 와 충돌한다.** 이 저장소의 설계된 해법은
`sanitized_pair_parity_test.go:1-60` 이 서술하는 **「§25 sanitized pair」** 이다 — 바이트 동일성을
지킬 수 없는 쌍은 `workflowOptMirroredPaths` 에서 **빼고** `sanitizedPairPaths` 레지스트리에
**넣어** 구조적 드리프트 가드로 옮긴다(선례: `main-checkout-branch-guard.md`,
`rule_template_mirror_test.go:68-73` 주석). 두 편집 모두 `internal/template/*_test.go` 이므로
`plan.md:81`(「[HARD] 코드 변경 없음. `internal/`, `pkg/`, `cmd/` 를 건드리지 않는다」)·
`spec.md:264-270`(§4 Out of Scope — 코드 층)·`acceptance.md:199`(DoD 「코드 변경 0」)에 정면으로
걸린다. iter-1 이 D1 의 required fix 말미에 적어 둔 것과 **같은 모양의 충돌이 되풀이됐다.**

— Severity: **critical** — Class: **blocking** — Required fix (둘 중 하나를 골라 SPEC 에 명문화):

- **(경로 A — 코드 무변경 유지, 권장 검토 1순위)** 비대칭을 포기하고 **두 사본을 동일하게** 두되,
  **로컬 사본도** 날짜·내부 경로를 싣지 않고 **런타임 버전만으로 귀속**한다. §1.5 가 이미 찾아 둔
  전례(`kanban-dispatch.md` 미러의 "Measured on Claude Code 2.1.275/2.1.276")가 **바로 이 형태**이며,
  그 파일은 날짜 0건으로 strict CI 아래 초록이다(측정 확인). 증거 기록은
  `.moai/specs/SPEC-WORKTREE-EXIT-RETURN-001/evidence/` 에 반출하되 **독트린 본문이 그 경로를 인용하지
  않게** 하고, 인용은 SPEC·progress 층에 둔다. 이 경로를 택하면 AC-WXR-005 (1)(a)(b)·(3) 과 §1.5 표의
  「기록 경로: 로컬에 있음 / 측정 날짜: 로컬에 있음」 행을 고쳐야 한다.
- **(경로 B — 비대칭 유지)** 대상 파일을 `rule_template_mirror_test.go` 의 `workflowOptMirroredPaths`
  에서 제거하고 `sanitized_pair_parity_test.go` 의 `sanitizedPairPaths` 에 **등재**한다. **두 편집을
  모두** 해야 한다 — 제거만 하면 이 파일을 묶는 상시 가드가 **하나도 남지 않는다**(아래 D2′).
  이 경로를 택하면 `spec.md` §4 「코드 층」 배제와 `plan.md` §D [HARD] 와 DoD 「코드 변경 0」을
  **함께** 고쳐 두 테스트 파일 편집을 범위 안으로 들여야 한다.

어느 경로든 **AC-WXR-007 에 판별식 (5) 를 추가**해 `go test ./internal/template/ -run TestRuleTemplateMirrorDrift -count=1` 의 exit 0 을 완료 **이전에** 관측하게 한다. 현재 AC-007 (3) 은 중립성 가드만 돌리므로 **자기 인수조건을 전부 통과하면서도 CI 를 적색으로 만드는 트리**가 성립한다 — iter-1 D1 이 지적한 것과 같은 사각이다. (AC 를 늘리지 않고 판별식만 더하므로 Tier S 상한 8 을 넘지 않는다.)

**D2′.** parity-binding-lost-on-removal — `acceptance.md:L156-158`(AC-WXR-007 (1)) — **D1′ 경로 B 를
최소 변경으로 이행하면 두 사본을 묶는 상시 검사가 사라진다.** AC-WXR-007 (1) 의 「규범 층 파리티」는
**판별식 결과의 일치**를 요구할 뿐 문언 일치를 요구하지 않는다. 즉 같은 판별식을 만족하는 서로 다른
문장 두 개가 허용된다. 더 중요한 것은 시점이다 — 이 검사는 이 카드의 run/sync 회차에 **한 번** 도는
문서 검사이지, 이후 편집을 잡는 상시 CI 가드가 아니다. 현재는 바이트 동일성 가드가 그 역할을 하고
있는데(V1 으로 입증), 경로 B 에서 허용목록 제거만 하고 sanitized-pair 등재를 빠뜨리면 이 파일은
**어느 상시 가드에도 속하지 않는 상태**가 된다. 배차가 물은 「두 문장이 서로 갈라질 수 있는가」의 답이
여기다 — 지금은 갈라질 수 없고(가드가 금지), 분할을 이행하는 순간 **등재를 함께 하지 않으면**
갈라질 수 있게 된다. — Severity: **major** — Class: **blocking** — Required fix: D1′ 경로 B 를 택할
경우 `sanitizedPairPaths` 등재를 **같은 커밋에서 의무화**하는 문언을 `plan.md` M3 와 AC-WXR-007 에
넣는다. 경로 A 를 택하면 이 결함은 발생하지 않는다(바이트 동일성 가드가 그대로 남는다).

**D11.** req001-sample-clause-unchecked — `spec.md:L168-176`(REQ-WXR-001) vs `acceptance.md:L24-30`(AC-WXR-001 판별식) — **D4 수정이 넣은 핵심 의무에 판별식이 없다.** REQ-001 은 이제 「표본을 **문장 자체에** 싣는다 — 연쇄 깊이 1·2, 그리고 **단일** 비-primary 출발-디렉터리 사례」를 요구한다. 그런데 AC-001 의 세 검사는 (1) `primary checkout` 적중, (2) `not the previous worktree` 적중, (3) **전칭 술어 0건**뿐이다. 즉 **금지(부정 의무)는 기계로 막히는데 표본 명시(긍정 의무)는 아무것도 검증하지 않는다.** 표본 문장을 통째로 빠뜨린 이행이 AC-001 을 전부 통과한다. — Severity: **minor** — Class: **blocking** — Required fix: AC-WXR-001 에 검사 (4) 를 더한다 — 복귀 지점 문장 또는 그 직후 3줄 이내에 관측 깊이(`1`·`2`)와 「단일 사례」에 해당하는 한정 문구가 적중할 것. AC 수는 늘지 않는다.

**D12.** req005-quotation-clause-unchecked — `spec.md:L192-200`(REQ-WXR-005) vs `acceptance.md:L103-120`(AC-WXR-005 판별식) — REQ-005 는 세 가지를 요구한다: (a) 측정 관측으로 제시, (b) **도구 계약의 인용으로 제시하지 않을 것**, (c) 보장 문언 금지. AC-005 의 다섯 검사 중 (c) 는 (5) 가, (a) 는 (1)(2) 가 덮지만 **(b) 는 어느 검사에도 걸리지 않는다.** `plan.md:106` 의 안티패턴 목록이 같은 위험을 이름 대고 있음에도(「상위 도구 설명 인용문을 우리 독트린에 복사해 두고 그것이 우리 계약인 것처럼 읽히게 하기」) 판별식이 없다. — Severity: **minor** — Class: **blocking** — Required fix: AC-WXR-005 에 검사 (6) 을 더하거나 (5) 를 확장한다 — 상위 도구 설명 문장을 인용하는 줄이 있으면 같은 줄 또는 인접 줄에 「우리 계약이 아님/상위 런타임 설명임」 취지의 한정어가 함께 적중할 것.

### Optional findings (오케스트레이터 재량 — 이 목록만으로는 FAIL 을 만들지 않았다)

**O1.** ac003-line-scope-coarse — `acceptance.md:L62-68` — AC-003 의 (i)(ii) 는 **줄 단위** 판별식인데 대상 줄(`worktree-integration.md:220`)은 **553자짜리 한 문단**이다(`sed -n '220p' … | wc -m` → 553). 문단 어디에든 `primary checkout` 이 있으면 (i)(ii) 가 동시에 만족되므로, 한정어가 「originating checkout」에서 멀리 떨어진 이행도 통과한다. 이진성은 유지되므로 D7 은 해소된 것이 맞고, 이것은 정밀도 문제다. — Severity: minor — Class: optional.

**O2.** ac008-not-req-anchored — `acceptance.md:L174` — 8개 AC 중 AC-008 만 `Covers` 가 REQ 가 아니라 `plan.md` §C 조항을 가리킨다. REQ 상한 8 이 이미 찼으므로 REQ-009 신설은 예산 위반이 되고, 대응이 **명시 인용**으로 성립하므로 주제 추론에 기대지 않는다. 구조적으로 허용 가능한 선택으로 본다. — Severity: minor — Class: optional.

**O3.** probe-worktree-residue — iter-1 O1 의 재확인. 별도 측정하지 않았으므로 iter-1 기록을 그대로 운반한다(격상하지 않음). — Severity: minor — Class: optional.

---

## iter-1 자신의 오류 — D5 의 처방은 공허했고, 저자가 옳다

배차가 「저자가 옳으면 그렇다고 분명히 말하라」고 요구했다. **저자가 옳다.**

iter-1 D5 의 required fix 는 이렇게 적었다: "판별식은 `git merge-base --is-ancestor <M1 커밋> <M2 커밋>`
이 exit 0 을 낼 것." **이 한 검사만으로는 공허하다.** 측정:

```
git merge-base --is-ancestor HEAD HEAD ; echo $?     → 0
git merge-base --is-ancestor d20a9f022 ce2e35355     → 0
git merge-base --is-ancestor ce2e35355 d20a9f022     → 1   (양성 대조 — 방향성은 살아 있다)
```

**커밋은 자기 자신의 조상이다.** 따라서 M1 기록과 M2 개정이 **한 커밋에 함께 들어가도**
`<M1>` 과 `<M2>` 가 같은 SHA 로 해소되어 검사는 exit 0 을 낸다 — iter-1 이 금지하려던 바로 그
상황에서 통과한다. 이것은 iter-1 이 D5 본문에서 정확히 진단한 위험(`verification-claim-integrity.md`
§2.3, 「git 은 커밋 내부의 저작 순서를 원리상 기록하지 않는다」)을 **처방이 스스로 재현한** 모양이다.

저자는 이를 잡아 AC-WXR-008 에 두 번째 검사를 넣었다(`acceptance.md:180-184`):

> 2. 두 커밋이 **서로 다르다** — … (1) 은 같은 커밋에서도 exit 0 을 내므로 이 검사가 없으면
>    판별식이 공허해진다.

**두 검사가 함께 있어야 비공허하다는 판단이 맞고, 문언에 실제로 둘 다 들어 있다.** 더하여
`acceptance.md:192-194` 는 「한 커밋에 합쳐야만 하는 사정이 생기면 이 AC 를 통과한 것으로 처리하지
않는다」까지 명문화해 VCI §2.3 의 명시적 지시를 이행한다. **D5 는 해소됐고, iter-1 의 처방보다
나은 형태로 해소됐다.**

---

## D9 분할 판정 (배차 질문 4)

**분할은 건전하다.** 근거:

1. **반출 대상 경로가 실제로 인용 가능하다.** `git check-ignore -v .moai/specs/SPEC-WORKTREE-EXIT-RETURN-001/evidence/x.md` → **exit 1**(무시되지 않음). 같은 디렉터리의 기존 파일 4개는 `git ls-files --error-unmatch` 가 exit 0 을 낸다. iter-1 D9 가 지적한 「해소되지 않는 경로」 문제는 사라진다.
2. **미러의 대체 귀속 표면이 실재한다.** `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md:207,237` 이 "Measured on Claude Code 2.1.275" / "2.1.276" 을 싣고 있고, 그 파일의 날짜 리터럴은 **0건**(`grep -cE '\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b'` → 0, exit 1). strict CI 아래 초록인 배포 파일이 이미 같은 형태를 쓴다. §1.5 가 「새 형식을 발명하지 않고 그 형태를 따른다」고 쓴 것은 사실이다.
3. **귀속을 약화시키지 않는다.** 미러는 「측정된 관측이다」류 무귀속 문구로 뭉개지 않고 버전이라는 **낡음을 보이게 만드는 좌표**를 남긴다. 사용자 프로젝트 독자가 `claude --version` 과 대조할 수 있다는 논거도 성립한다.

**그러나 분할이 새 문제를 만든다 — 그것이 D1′ 과 D2′ 다.** 분할의 *논리*는 옳고 *기계적 실행
가능성*이 막혀 있다. 배차가 물은 「두 문장이 갈라질 수 있는가」의 답은 위 D2′ 에 적었다: 현재는
바이트 동일성 가드가 금지하므로 갈라질 수 없고, 분할을 이행하는 순간 **sanitized-pair 등재를 함께
하지 않으면** 갈라질 수 있게 된다.

## REQ-WXR-008 이 흐려졌는가 (배차 질문 2) — 아니다

iter-1 의 지시는 「미러 파리티를 재량으로 뭉개 해소하지 말 것」이었다. 재측정 결과 **뭉개지 않았다.**

- 허용되는 이탈이 **닫힌 목록**이다: 부재 3건(측정 날짜 리터럴 / 저장소 내부 증거 경로 / SPEC ID) + 대체 1건(런타임 버전 인라인). `spec.md:213-221` 이 문장으로, `:135-141` 이 표로, `acceptance.md:159-162` 가 기계 검사로 같은 목록을 세 번 고정한다.
- 묶이는 규범 층도 **열거**돼 있다: 복귀 지점 · 연쇄 케이스와 명시적 `not B` · 두 트리 응답 경고 · 정정된 `originating` 문언 · 비보장 헤지 · 잔여 위험 문구 · 열린 미재현 제보(`spec.md:215-218`).
- 「재량 판단」을 요구하는 문구가 이 조항에 없다. `acceptance.md:156-158` 은 어느 판별식이 두 사본에서 같은 결과를 내야 하는지를 **번호로** 지목한다.

**판정: 닫힌 형태이고 기계 판정 가능하다. 이행자 재량이 남지 않는다.** 이 축에서 FAIL 사유는 없다.

## M0 반출 유예 판정 (배차 질문 6) — 유예가 옳다

plan 단계의 산출물은 `spec.md`·`plan.md`·`acceptance.md` 이고, 반출은 파일을 만드는 행위이므로
run 단계 M0(`plan.md:61`)의 일이 맞다. 유예가 **평가 불가능한 AC 를 남기지 않는다**:
AC-WXR-005 (3) 은 run/sync 시점에 `git ls-files --error-unmatch` 로 판정되고, 그 판정이 성립할 수
있음을 위 D9 판정 1번에서 미리 측정해 두었다(대상 디렉터리가 무시되지 않음). DoD(`acceptance.md:202-205`)
도 같은 검사를 완료 조건으로 걸어 둔다. **plan 단계에서 증거를 만들지 않은 것은 결함이 아니라
단계 경계를 지킨 것이다.**

## 상한 압박 판정 (배차 질문 7)

REQ **8/8**, AC **8/8** — 둘 다 Tier S 상한에 **정확히 닿았다**(`spec-workflow.md:144-146`).

- **커버리지 손실은 없다.** 8개 REQ 전부 최소 1개 AC 가 덮는다(위 Traceability 표). 배차가 지목한
  REQ-WXR-007 은 AC 를 따로 받지 못했으나 AC-004(`:76`)와 AC-005(`:97`)가 `Covers` 로 명시 인용하고,
  AC-005 (4) 가 잔여 위험 문구의 생존을 직접 검사한다. **AC 를 늘리지 않은 판단이 커버리지를
  희생시키지 않았다.**
- **그러나 상한은 이제 실제로 압박한다.** D1′ 이 요구하는 파리티 가드 관측 자리가 새 AC 를 필요로
  하는데 예산이 0 이다. 그래서 위 required fix 를 **AC 신설이 아니라 AC-WXR-007 의 판별식 (5) 추가**로
  적었다. D11·D12 도 같은 이유로 판별식 추가 형태로 처방했다. 상한 안에서 해결 가능하다.
- `spec-workflow.md:148` 의 「상한 초과는 tier-up 또는 분할의 신호」에 비추어, **이 SPEC 이 이번
  수정에서 또 다른 의무를 얻는다면 Tier M 승격을 검토할 자리**임을 기록해 둔다(지시 아님).

---

## 무엇을 검사했고 무엇이 통과했나 (FAIL 을 만들 수 있었으나 만들지 않은 것)

1. **AC-WXR-007 (3) 은 공허하지 않다.** 심는 즉시 실패함을 변이 V2·V3 로 입증했다(`C1-spec-id-prefix`, `S1-internal-date` 적중). 배차가 의심한 「부재 검사가 위반 시 실제로 실패하는가」에 대해 **두 항목은 실패하고 한 항목은 AC 자신의 grep 에만 기댄다**는 것까지 분해했다.
2. **iter-1 결함 9건 중 8건이 실질 수정이다.** 문언을 문자열 단위로 대조했고, 「고친 척」에 해당하는 항목은 없었다. D3 은 근거 파일 `observations.md:114-116,139-141` 과 문언 일치까지 확인했다.
3. **전칭 술어가 다른 요구사항으로 옮겨 가지 않았다.** `awk '/^## 2. Requirements/,/^## 3. Acceptance/'` 로 §2 를 잘라내 `independent of|always|regardless of|항상|무관하게|guaranteed|never` 를 전수 검색한 결과, 적중은 **금지 조항 자신 2곳뿐**이었다. D4 가 한 요구사항에서 다른 요구사항으로 도망치지 않았다 — 이 축은 FAIL 사유가 될 수 있었으나 되지 않았다.
4. **독트린 표적이 여전히 실재한다.** `.claude/rules/moai/workflow/worktree-integration.md:216` 에 절 제목, `:220` 에 무수식 문장 "`ExitWorktree` returns to the originating checkout" 이 있고 미러도 같은 줄에 같은 문장을 담는다(두 사본 sha256 동일). 표적이 사라졌으면 SPEC 전체가 무너졌을 것이다.
5. **must-pass 7개 전수 재측정.** 델타 범위로 줄이지 않고 전부 다시 쟀다 — `spec.md` 가 개정됐으므로 iter-1 의 must-pass 통과를 상속하지 않는 것이 옳다. 7/7 유지.
6. **점수 회귀 없음.** 네 축 어느 것도 iter-1 아래로 내려가지 않았다. LEAN STOP 조건은 발화하지 않으며, 이 FAIL 은 「진전 없음」이 아니라 「새 결함 1건」이다.

---

## Verdict 근거 — 점수는 넘는데 왜 FAIL 인가

집계 0.80 은 Tier S 문턱 0.75 를 넘는다. 그런데도 FAIL 을 낸 근거를 숨기지 않고 적는다.

M5 방화벽의 7개 must-pass 는 전부 통과했고, **D1′ 는 그 7개 중 하나가 아니다.** 따라서 이 FAIL 은
방화벽 조항의 자동 발화가 아니다. 적용한 규칙은 MP-5·MP-6 이 세워 둔 **선례의 유추**다 — 그 두 조항은
「기계 검증된 단일 BLOCKING 발견은 집계 점수와 무관하게 `Verdict: FAIL` 을 강제한다」는 형태를
이 문서 안에 이미 확립해 두었다. D1′ 는 그와 같은 모양이다: **변이 실험으로 기계 입증됐고, SPEC 을
문언대로 이행하면 완료 시점에 저장소의 주 테스트 잡이 적색이 되며, 그 처방은 SPEC 자신의 [HARD]
배제 조항과 충돌한다.** 자기 제약을 어기지 않고는 완료할 수 없는 SPEC 은 루브릭 평균이 얼마든
승인 대상이 아니다.

**점수를 문턱 아래로 낮춰 판정과 맞추지 않았다.** 그렇게 하는 것은 판정을 정당화하려고 측정을
조정하는 일이고, 이 역할이 금지하는 합리화의 거울상이다. 점수는 정직하게 0.80 으로 보고하고,
판정은 그 자체의 증거 위에 세운다. 오케스트레이터는 이 둘을 따로 읽을 수 있다.

**남은 일은 좁다.** D1′ 의 두 경로 중 하나를 고르는 **결정 하나**와, 그에 딸린 문언 수정
(§1.5 표 1~2행 또는 §4·§D·DoD 3곳) + 판별식 3개 추가(D1′·D11·D12)다. **코드 변경은 경로 A 에서
0 이고 경로 B 에서 테스트 파일 2개다.** SPEC 의 구조·추적성·검증가능성은 이미 건전하다.

---

## Recommendation

되돌리기 비용 순.

1. **D1′ 의 경로 결정부터 (critical).** 경로 A(두 사본 동일 + 로컬도 버전 귀속만)와 경로 B(비대칭 유지 + sanitized-pair 등재)는 §1.5 표·§4·§D·DoD·AC-005·AC-007 의 형태를 **동시에** 정한다. 다른 수정보다 먼저 확정해야 한다. **경로 A 가 코드 무변경 제약을 지키고 §1.5 가 이미 인용한 전례와 정확히 같은 형태이므로 먼저 검토할 값어치가 있다** — 다만 그 대가로 로컬 사본이 증거 경로 인용을 잃으므로, 귀속 앵커를 SPEC·progress 층으로 옮기는 문언이 함께 필요하다.
2. **AC-WXR-007 에 판별식 (5) 추가 (critical 의 관측 자리).** `go test ./internal/template/ -run TestRuleTemplateMirrorDrift -count=1` exit 0 을 완료 **이전에** 관측한다. 경로 A·B 어느 쪽이든 필요하다 — A 에서는 파리티가 지켜졌음을, B 에서는 등재가 끝났음을 확인하는 자리가 된다.
3. **경로 B 를 택할 경우 D2′ (major).** `sanitizedPairPaths` 등재를 같은 커밋에서 의무화하는 문언을 `plan.md` M3 와 AC-WXR-007 에 넣는다. 허용목록 제거만 하는 이행은 상시 가드를 0 으로 만든다.
4. **D11·D12 (minor, blocking).** AC-WXR-001 에 표본 명시 검사를, AC-WXR-005 에 도구-계약-인용 한정 검사를 더한다. **둘 다 판별식 추가이므로 AC 수는 8 로 유지된다.**
5. **`plan.md` §G 에 파리티 가드 2건 추가.** `internal/template/rule_template_mirror_test.go`(바이트 동일성, 센티널 `RULE_TEMPLATE_MIRROR_DRIFT`)와 `internal/template/sanitized_pair_parity_test.go`(구조 드리프트, 센티널 `SANITIZED_PAIR_PARITY_DRIFT`). 이 SPEC 의 산출물을 지배하는 가드가 제약 목록에 없었던 것이 D1′ 의 발생 경로다.

O1·O2·O3 은 재량 항목이다. **이 셋을 근거로 FAIL 을 만들지 않았으며** 이행을 강제하지 않는다.

Tier S 반복 상한은 1 이므로(`harness.plan_audit_tier_ceilings`) 이 회차는 이미 상한 밖이다. iter-3
진입 여부는 오케스트레이터의 판단이며, 이 결함 목록은 **문언 수정과 결정 하나로 닫히는 좁은 델타**
이므로 재감사 범위를 D1′·D2′·D11·D12 로 한정하는 것으로 충분하다.

---

## Gaps (관측하지 않은 것 / 거부된 것)

`verification-claim-integrity.md` §3.1 — 거부는 Gap 이며, 대체로 한 일을 함께 적는다.

- **워크트리 가드가 두 명령을 거부했다.** (1) `for` 루프를 포함한 복합 비교 명령 → "too complex to verify"; (2) 런타임 변수 `$D` 를 `sed` 인자에 쓴 명령 → "a value computed at runtime where an option may stand". **대체로 추론하지 않았다** — 두 경우 모두 같은 질문을 평문 명령으로 쪼개 다시 실행해 **측정으로 답을 얻었다**(절 위치 확인, frontmatter 판독). 판정에 사용된 값은 전부 실제 실행 출력이다.
- **런타임 동작 자체는 재측정하지 않았다.** 이 감사는 문서 층 감사이며, `ExitWorktree` 복귀 지점을 직접 재지 않았다. 관측 1~4 의 진위는 이 판정의 대상이 아니다(AC-WXR-004 가 run 단계에서 닫을 일이다).
- **CI 를 실제로 돌리지 않았다.** D1′ 의 CI 적색 주장은 **로컬 `go test` 실행 결과 + `ci.yml` 의 트리거·필터 문언 판독**에 기댄다. GitHub Actions 실행으로 확인하지 않았다. 다만 필터 목록(`ci.yml:76-95`)에 두 경로가 문자로 들어 있고 `test` 잡이 `go test ./...`(`:210`)이므로 추론 구간은 짧다.
- **`TestSanitizedPairParity` 를 실행하지 않았다.** 등재 시 초록이 되는지는 재지 않았고, 그 테스트의 머리 주석 서술(`:1-40`)만 읽었다. 경로 B 를 택하면 run 단계에서 실측이 필요하다.
- **변이 실험의 복원은 sha256 으로 확인했으나**, 실험 도중 다른 세션이 같은 파일을 읽었을 가능성은 배제하지 않았다(감사 창 단일 쓰기 주체 가정에 기댄다).

## Residual-risk

- 변이 V1~V3 는 대상 파일 **말미에 한 줄을 덧붙이는** 형태였다. 실제 M2 개정은 §216-228 절 **안에서** 이뤄지므로 위치가 다르다. 가드 셋 모두 파일 전체 또는 바이트 전체를 보므로 판정은 위치에 무관하지만, 「절 안에서만 재면 다른 결과가 나온다」는 가능성을 0 으로 증명하지는 않았다.
- AC-WXR-007 (2)(b) 의 내부 경로 부재는 CI 가 받치지 않는다(측정으로 확인). 이행 후 누군가 미러에 `.moai/` 내부 경로를 넣어도 **어떤 가드도 발화하지 않는다.**
- 이 회차는 Tier S 상한(1) 밖의 iter-2 다. 상한 규약상 이 판정의 처리 방식은 오케스트레이터 몫이다.

---

*Audited by plan-auditor. M1 Context Isolation 적용. 모든 PASS·FAIL 판정은 위에 인용한 명령과 그 관측 출력에 귀속된다. 커밋 없음 — 작업 트리는 감사 전 상태로 복원됐다(sha256 대조).*
