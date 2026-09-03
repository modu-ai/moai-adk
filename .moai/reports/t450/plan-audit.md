# SPEC Plan Audit Report: SPEC-PLAN-AUDITOR-RESIDUE-001

- 카드: t450 · 반복: 1/3 · 감사 트리: `7835148d3` (브랜치 `WT-plan-auditor-residue`, SPEC 산출물 untracked 상태 감사)
- 감사자: plan-auditor (독립 감사 — 작성자 추론 컨텍스트 배제, M1 Context Isolation 준수)
- 판정 기준 트리: 이 문서의 모든 측정값은 `git rev-parse --short HEAD` = `7835148d3` 트리에서 이번 실행으로 직접 재측정한 값이다.

## Verdict

**FAIL** (반복 1/3 — 차단 결함 D1 수리 후 결함 델타 범위 재감사)

**Overall Score: 0.90** — 네 차원 모두 루브릭 1.0이지만, release-blocking AC 미채택 결함(D1)에 대한 채택 규율 공제로 0.90로 보고한다.

- Tier 해석: spec.md frontmatter에 `tier:` 필드가 없다(plan.md는 Tier M 판정, progress.md §E.1도 `tier: M`). 재시도 천장·PASS 문턱의 기계 해석 규칙상 frontmatter 부재는 Tier L로 resolution된다(천장 3회, 문턱 0.85). 본 감사는 더 엄격한 Tier L 문턱을 적용했다. 0.90 ≥ 0.85 — 점수 기준으로는 문턱을 넘지만, 차단 결함 1건이 남아 있어 판정은 FAIL이다. skip-eligibility는 PASS 판정을 요하므로 어느 쪽으로도 skip 대상이 아니다.

## Must-Pass Results (자동통과 아닌 항목만)

- **[PASS] MP-1 REQ 번호 일관성** — REQ-001(spec.md:48)부터 REQ-008(spec.md:62)까지 연속, 공백·중복 0, zero-padding 일관. 재측정: `grep -Eo 'REQ-[0-9]{3}' spec.md | sort -u` → 8개 연속 확인.
- **[PASS] MP-2 GEARS 형식 준수 (판정 계층: 요구 계층 — spec.md의 REQ-XXX 항목)** — 8건 전부 패턴 적합: REQ-001·003·004·005·006 Ubiquitous("~해야 한다"), REQ-002 Unwanted("~해서는 안 된다" — t367이 착지시킨 현행 루브릭 문면상 legacy-equivalent 허용 범위, Score 1.0 backward-compat 창 내 유효), REQ-007·008 Event-driven("When … 면/때, … 해야 한다" — t367이 정의한 제5패턴 Event-detected 형태). AC 계층(Given-When-Then)은 M3 § Scope에 따라 이 기준에서 제외 — acceptance.md의 Given-When-Then은 검증 계층의 올바른 형식이다.
- **[PASS] MP-3 YAML frontmatter 유효성** — 12개 정식 필드 전부 존재·형식 적합(spec.md:2-13): `id`/`title`(quoted)/`version` "1.0.0"(quoted semver)/`status` draft/`created`·`updated` 2026-09-03(ISO)/`author`/`priority` P2/`phase` "v3.2.0 target"(릴리스 목표 라벨, 금지 어휘 아님)/`module`/`lifecycle` spec-anchored/`tags`(CSV string). snake_case 별칭(`created_at` 등) 0건.
- **[N/A] MP-4 언어 중립성** — 단일 에이전트 정의 파일(plan-auditor.md) 범위 SPEC으로 다국어 도구 다룸 없음. 자동통과.
- **[PASS] MP-5 D7 cross-SPEC 조정** — spec.md 본문의 SPEC-ID 추출 결과: `SPEC-DOMAIN-WO-001`, `SPEC-EXAMPLE-DOMAIN-001`(.moai/specs/에 없음 — D7-5 SHOULD). 단 둘 다 plan-auditor.md:338의 기존 드리프트 hunk를 **인용한 예시 식별자**일 뿐 의존 참조가 아니므로 BLOCKING 아님(기록 결함 D6). plan.md §H의 `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001`은 존재하며 `status: implemented`(직접 판독) — retired/superseded/archived 아님, 조정 불요. BLOCKING finding 없음.
- **[PASS] MP-6 D8 cross-platform 규율** — spec.md 내 `syscall` 0매치(재측정: grep -c → 0, exit 1). D8-4 자동통과.
- **[PASS] MP-7 clarification 게이트** — plan.md에 `[NEEDS CLARIFICATION: <topic>]` 규약 형식(`.claude/skills/moai-workflow-spec/SKILL.md:178` — 콜론+토픽 필수) 마커 0건. plan.md:20의 `[NEEDS CLARIFICATION 없음]`은 **부정문**이지 미해결 마커가 아니므로 게이트 위반이 아니다(형태와 역할의 구분 — 형태 매칭 게이트의 거짓 양성 유발 소지는 D5로 기록). research.md는 Tier M 산출물 셋에 없어 N/A.

## Category Scores (0.0-1.0, 루브릭 앵커 적용)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 1.0 | 1.0 | REQ-001이 반출 위치를 파일 패밀리 수준에서 못박음(spec.md:48, 규약 § Where :37-40 패밀리와 일치). REQ-004는 쌍둥이 일치 판정을 조항 구간으로 한정하고 기존 2 hunk를 명시적으로 제외(spec.md:54 — 실측으로 정확히 2 hunk 존재 확인). REQ-007은 대상 문면을 :72(Event-detected 불릿)·:73(Unwanted legacy-only 불릿)로 특정 — 실측 일치. 대명사 모호성·다의 해석 0건 |
| Completeness | 1.0 | 1.0 | HISTORY(:18)·WHY(:24)·WHAT(:38)·REQUIREMENTS(:46)·ACCEPTANCE CRITERIA(:64)·OUT OF SCOPE(:75, `### Out of Scope — <topic>` H3 3개 + 구체 불릿) 전부 존재. frontmatter 12필드 완비. 결함 3건(금지 경로·거짓 교차참조·곁말 미반영)이 모두 이번 실행 직접 측정치로 뒷받침됨 |
| Testability | 1.0 | 1.0 | AC-001~008 전부 이진 판정 가능(grep 횟수·diff 0·go test 출력). weasel word 0건. 5개 RED-now 셀의 명령·stdout·종료코드를 감사자가 전부 재실행해 바이트 수준 일치 확인(아래 Q4). 채택 규율 결함(D1·D2)은 이 차원이 아닌 Defects로 운반 |
| Traceability | 1.0 | 1.0 | REQ→AC 전피복(REQ-001→AC-001 … REQ-007→AC-007, REQ-008→AC-006), AC→REQ 역참조 전부 유효(AC-006→REQ-006+REQ-008, AC-008→REQ-006). 고아 AC 0건, 미피복 REQ 0건(acceptance.md §D 매트릭스 대조) |

## Defects Found

- **D1** — acceptance.md:56-58 (AC-004) — release-blocking으로 분류된 AC-004(조항 쌍둥이 일치)에 RED-now 셀과 green-path 셀이 **모두 없다**. `verification-completeness.md` §2 [HARD](두 셀 채택 규율) 위반이며, 조항 추출+diff 도구의 적색이 관측된 적이 없어 §1.1상 도구 미완결 상태다. 구체적 위험: 조항 구간 추출 프록시가 항상 0을 반환하는 고장 형태(쌍둥이 파일이 이미 2 hunk로 갈라져 있고 AC-004는 그것을 제외하므로, 추출 경계가 틀리면 편집 후에도 무조건 초록)에서 AC-004 PASS가 기록될 수 있다. — Severity: **major** — Class: **blocking** — Required fix: AC-004에 (a) RED-now 셀 — 전체 파일 diff가 현재 2 hunk·exit 1임을 관측해 diff 다리의 적색 감지를 증명하고, 조항 구간 추출분은 양쪽 0매치 baseline임을 명시(공집합 가시화), (b) green-path 셀 — M2가 유의 상태로 전환하며 통과 출력 형태("조항 구간 diff 0, exit 0")를 명시. 또는 §2.1 undecidable 성격을 명문화해 회귀 가드로 재분류하는 문서화된 근거를 남긴다.
- **D2** — acceptance.md:62-70 (AC-005) — 녹색 판별 프록시가 뒤집히지 않는다. RED-now의 grep(`reports/plan-audit/`)은 편집 후에도 run-gate 스트림(`internal/runtime/audit_report.go` 소관, 본 SPEC 범위 밖) 서술 때문에 0이 될 수 없다. (a)·(b) 조건은 사람이 문장을 읽으면 판정 가능하나 "통과 시 출력이 무엇이 되는가"가 셀에 없어 §2의 green-path 정의를 충족하지 않는다. — Severity: minor — Class: optional — Required fix: 녹색에서 실제로 뒤집히는 프록시로 교체·병기(예: spec-workflow.md 내 `plan-audit/{SPEC-ID}-review` 매치 0, 또는 :407 문장의 plan-auditor 연결 구절 제거 확인).
- **D3** — spec.md frontmatter — `tier:` 필드 부재. progress.md §E.1은 `tier: M`을 선언하지만 프론트매터가 없어 기계 해석(감사 천장·skip 문턱)은 Tier L로 간다. 엄격 방향이라 실해는 없으나 선언된 판정과 해석 값의 이중 운반체 불일치다. — Severity: minor — Class: optional — Required fix: spec.md frontmatter에 `tier: M` 한 줄 추가.
- **D4** — plan.md:20 — "범위 emit" 표현 부정확. emit 도구(`AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission`, Makefile:39)에는 범위 모드·`--entry` 플래그가 없고 실제 기법은 **전체 재생성 후 sync-auditor.toml을 develop 값으로 복원**이다(2549f775f 본문 및 stat가 그 end-state를 증명 — catalog.yaml + plan-auditor.toml만 변경, sync-auditor.toml 미포함). M3의 실행 단계 서술은 정확하므로 실행 가능성에 영향 없음. — Severity: minor — Class: optional — Required fix: §A 문구를 "전체 emit 후 sync-auditor.toml 복원"으로 정정.
- **D5** — plan.md:20 — `[NEEDS CLARIFICATION 없음]` 대괄호 형태가 MP-7 형태 게이트(`grep '\[NEEDS CLARIFICATION'`)의 거짓 양성을 유발한다. 규약 형식은 콜론+토픽(`SKILL.md:178`)이므로 게이트 위반은 아니지만, 모양만 보는 후속 게이트가 이 줄을 미해결 마커로 오판할 수 있다. — Severity: minor — Class: optional — Required fix: "NEEDS CLARIFICATION 마커: 없음" 등 대괄호를 벗긴 표현으로 변경.
- **D6** — spec.md:54·83 — `SPEC-DOMAIN-WO-001`·`SPEC-EXAMPLE-DOMAIN-001`이 .moai/specs/에 없음(D7-5 SHOULD). 둘 다 plan-auditor.md:338 기존 문면의 인용 조각이라 의존 참조가 아니므로 실질 결함 아님 — 기록 목적. — Severity: minor — Class: optional — Required fix: 불요(인용임을 나타내는 표현이 이미 유지되면 그대로 둔다).

## Audit Questions (Q1-Q6)

**Q1. 지연 조항 2건의 식별 정밀도·출처 정확성 — 예.**
- 곁말 규약: 커밋 SHA 재유도 결과 `f47d7f5a9`("docs(t387): codify side-talk rule for audit artifacts…")로 SPEC 표기가 정확하다(감사 지시서의 "f47d5a9*"는 불정확 — 정본은 SPEC 쪽). f47d7f5a9 본문에 plan.md §A가 인용한 문장이 **축자**로 존재한다: "Agent-definition reflection (plan-auditor / sync-auditor) deferred with t386's export-mandate clause until the lane-9 rubric/ownership cards settle (verdict.md Gaps)." 규약 문서 § Side-talk은 :87에 존재(로컬·템플릿 미러 동일). `.moai/reports/t387/verdict.md` Gaps 절(:39-42)이 plan-auditor·sync-auditor 양쪽 반영 보류를 명시하며, 이 파일은 **tracked**(`git ls-files` 확인)라 인용 대상으로 유효하다. REQ-003의 3요소(별도 미검증 절/측정 지시 형태/measured·inferred·assumption 라벨)는 f47d7f5a9 본문의 3규칙 서술과 정확히 대응한다.
- 반출 조항: `4244c4a06` 본문에 plan.md §A 인용문이 축자로 존재한다: "The plan-auditor counterpart stays deferred until t367 (rubric revision) closes." 동 커밋 stat는 sync-auditor.md 2벌(로컬+미러)만 +2행 — sync-auditor 전용 착지라는 SPEC 서술과 일치. 모델 문면은 sync-auditor.md:92 [HARD] 단락(양쪽 사본 확인)으로 실재하며, plan.md §A의 "같은 자리·같은 문체" 지시는 이 실물을 가리킨다.
- 전제 확인: `git merge-base --is-ancestor 18fc2c9ef 7835148d3` → 조상 맞음("18fc2c9ef IS ancestor"), 머지 커밋 `a3f4bc617`("Merge branch 'WT-gears-canon-rubric' into develop (card t367)") 실재.

**Q2. t367 :72 무겹침 주장 — 참.**
`git show 18fc2c9ef`의 plan-auditor.md diff는 hunk 1개(`@@ -69,7 +69,8 @@`)로 루브릭 불릿 목록만 건드린다. 현행 :72 = Event-detected 제5패턴 불릿, :73 = Unwanted legacy-only 불릿 — SPEC 서술과 정확히 일치. 본 SPEC의 편집 영역은 § Output Format(:393-395, "Write the audit report to `.moai/reports/plan-audit/…`" 실측)과 그 뒤에 붙는 조항 — 텍스트적 겹침 없음. AC-007 보존 판정 재측정: `grep -c "the fifth GEARS pattern"` 양쪽 사본 각 1.

**Q3. 범위 재생성 기법의 실재성·경계 — 실재하며 올바르게 경계 설정됨(표현 1건 부정확 = D4).**
- `gen-catalog-hashes`는 `--entry NAME` 플래그를 공식 지원한다(`internal/template/scripts/gen-catalog-hashes.go:10` — "--entry NAME    Update a single named entry").
- emit은 전체 재생성(`AGENTEMIT_UPDATE=1` 골든 테스트)이며, 선례 `2549f775f`의 실제 기법이 이것이다: 본문 "Repaired scoped: gen-catalog-hashes --entry plan-auditor + make agents-emit, then restored sync-auditor.toml (t443/4244c4a06-owned drift deliberately NOT repaired)" — stat도 catalog.yaml(2±) + plan-auditor.toml(3±2)만 변경되고 sync-auditor.toml은 미포함 = end-state가 정확히 M3가 서술한 바다.
- AC-006 경계의 현재 기준선: `go test ./internal/template/agentemit/...` 재측정 → `--- FAIL: TestGoldenCommittedArtifactsMatchEmission` / `golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing` / exit 1 — 적색이 정확히 sync-auditor.toml 1건. AC-006의 합격 조건("plan-auditor 녹색 + 적색 sync-auditor.toml 1건 한정")은 이 기준선과 정합적이다. AC-006이 인용하는 TestCatalogHashParity·TestManifestHashFormat도 실재(`internal/template/catalog_tier_audit_test.go`), catalog.yaml의 plan-auditor 항목도 실재(:142-146).

**Q4. RED-now 셀 완결성 — 6건이 아니라 5건이며, release-blocking 1건(AC-004)이 셀 없음 = D1.**
- 셀 있는 곳: AC-001·002·003·005·006 — 5개 셀 전부 명령+축자 stdout+종료코드를 갖추고, 트리 SHA는 문서 수준 핀(acceptance.md:3 "기준 트리: 7835148d3")이 무핀 기준을 결속(§2.1 허용 형태). 5개 명령은 전부 단일 호출 형태(파이프·체인 없음). 감사자 재실행 결과: AC-001 `Export mandate` 양쪽 0매치 exit 1, AC-002 `:395` 히트 축자 일치 exit 0, AC-003 `Side-talk` 0매치 exit 1, AC-005 `:407` 히트 축자 일치("Two report streams coexist deliberately in `.moai/reports/plan-audit/`…"), AC-006 골든 FAIL 출력 축자 일치 exit 1 — **전부 바이트 수준 재현**.
- AC-004는 release-blocking인데 셀 0개(→ D1, blocking). AC-005는 RED는 있으나 녹색 출력 형태 미정의(→ D2, optional).
- 나머지 2건의 비-release-blocking 분류는 올바르다: AC-007은 GREEN-now 보존 판정(편집 전 적색이 원리상 불가능 — §2.1 undecidable 성격에 맞는 회귀 가드 분류), AC-008은 신규 편집분 0매치 검사로 도착 시 녹색이 설계상 확정 — 회귀 가드 분류가 규율상 정확한 처치다. 형제 가드 기준선: `go test -run 'TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit' ./internal/template/` → ok (이번 실행, 트리 7835148d3).

**Q5. 중립성 0매치 기준의 구체성 — 구체적이고 grep 가능하다.**
AC-008은 토큰 4류(SPEC-ID 일반형·카드 id·커밋 SHA·내부 날짜)를 §25.1 C1-C8 카탈로그에 연결하고, CI 가드 실물(`.github/workflows/template-neutrality-check.yaml` — TestTemplateNeutralityAudit + TestTemplateNoInternalContentLeak + `MOAI_TEMPLATE_LEAK_STRICT=1` 엄격 티어)과 형제 Go 가드를 검증 경로로 명명한다. 플레이스홀더 안전성 주장도 검증됐다: `<card-id>`는 이미 템플릿 sync-auditor.md:92에 착지돼 가드를 통과한 선례다(4244c4a06산). §25.1 카탈로그 대비 "카드 id"가 SPEC이 추가하는 더 엄격한 요소라 지원 가드의 클래스 목록과 1:1이 아닌 점만 유의 — 초집합 요구라 결함 아님.

**Q6. 3축 범위 준수 — 이탈 없음.**
측정된 편집 대상 8파일(로컬+템플릿 plan-auditor.md, C3 plan-auditor.toml, catalog.yaml, spec-workflow.md 양쪽, audit-artifact-convention.md 양쪽)은 정확히 카드 3축(a 쌍둥이+C3, b 교차참조 양 미러, c 중립성)에 귀속된다. M1-M4가 3축만 다루고, §G Anti-Patterns가 침범 5종(t443 몰래 수리, 기존 드리프트 drive-by, .toml 손편집, :72 재작성, sync-auditor 곁말 동봉)을 금지하며, Out of Scope 3개 H3가 경계를 문서화했다. verdict.md 작성은 lane 소관으로 SPEC 작성 범위에서 명시적으로 분리돼 있다(plan.md §M4).

## Gaps (이 감사가 관측하지 않은 것)

- run-phase 이후의 행위 확인(반출 파일이 `.moai/reports/t450/`에 실제 생성되는지) — plan-phase에서 관측 불가. acceptance.md §D.5가 run/sync 몫으로 명시.
- 실제 편집 수행 후의 neutrality 가드 통과 — 편집 전 트리 기준선만 측정(녹색 확인).
- `make agents-emit` 실행 자체는 수행하지 않았다(감사는 read-only; AC-006의 RED만 재측정). 범위 emit 경로의 실행 가능성은 2549f775f의 착지 end-state로 간접 증명.
- sync-auditor 쪽 곁말 반영의 현재 상태 — Out of Scope라 확인하지 않았다(확인 안 함이 원칙).
- AC-004 수리 시 제안한 "조항 구간 추출"의 구체 구현 문법 — 본 감사는 셀 구조만 검증했다.

## Residual-risk

- develop이 움직이면(`t443` 등) AC-006의 "적색 1건" 기준선이 변할 수 있다 — acceptance.md가 재측정 규정을 이미 갖고 있어(§D 머리글) 절차는 갖춰져 있다.
- 2549f775f 당시 골든 적색이 "4 canonical exclusions"였다는 선례 본문 서술과 현재 1건의 차이는 시대 차이(t441 카탈로그 커버리지 이전/이후)로 판단되나, 본 감사는 현재 트리 1건만 검증했다.
- 조항 문안 자체(M1 산출)의 중립성 통과 여부는 문안 확정 전까지 미결이다 — `<card-id>`/`<SPEC-ID>` 선례가 안전성 근거로 제시돼 있으나(검증됨), 최종 문면에 대한 가드 실행은 run-phase 몫이다.

## Recommendation

FAIL 사유는 단일 blocking 결함 D1이다. 수리 지시:

1. (D1, 필수) acceptance.md AC-004에 RED-now 셀과 green-path 셀을 추가한다. RED-now: `diff .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md` → 2 hunk·exit 1(본 감사에서 재현 확인 — diff 다리의 적색 감지 증명) + 조항 구간 추출 양쪽 0매치 baseline 명시. green-path: "M2가 조항을 양쪽에 적용하면 조항 구간 diff = 0·exit 0".
2. (D2) AC-005에 녹색에서 뒤집히는 기계 프록시를 병기한다(예: spec-workflow.md 내 `plan-audit/{SPEC-ID}-review` 매치 0).
3. (D3) spec.md frontmatter에 `tier: M` 추가.
4. (D4·D5) plan.md:20의 "범위 emit" → "전체 emit 후 sync-auditor.toml 복원", `[NEEDS CLARIFICATION 없음]` → 대괄호 없는 표현.
5. D6은 기록 목적으로 무수리 권고.

이 감사의 판정 파일은 `.moai/reports/plan-audit/`가 아니라 카드 증거 경로 `.moai/reports/t450/plan-audit.md`에 기록됐다 — 이는 본 SPEC이 정론화하는 audit-artifact-convention.md § Where(.moai/reports/plan-audit/ FORBIDDEN, "writing there is disposal, not export")의 그 자체 적용이다.

---

# Iteration 2 — Delta Re-audit (수리 커밋 8faea987c)

- 반복: 2/2 (Tier M 천장 — `tier: M`이 frontmatter에 추가돼 기계 해석도 M으로 resolution)
- 감사 트리: `8faea987c` (브랜치 `WT-plan-auditor-residue`, 워킹트리 클린 — `git status --porcelain` 0행 관측)
- 범위: 반복 1 결함 D1-D5의 델타 + 회귀 검사. 전면 재감사 아님(재시도 계약의 델타 스코프).

## Verdict

**PASS** — **Overall Score: 1.0** (반복 1의 0.90에서 상승 — 단조성 요건 0.90 이상 충족; Tier M 문턱 0.80 통과)

차단 결함 없음. 반복 1의 5건 결함 전부 해소 확인. 신규 결함 없음(수리 커밋은 SPEC 산출물 4개 + 반복 1 판정문만 추가 — 에이전트·템플릿·문서 표면 무변경).

## 이전 반복 결함 회귀 검사

- D1 (blocking) — **RESOLVED**: 아래 델타 판정 참조.
- D2 — **RESOLVED**: AC-005 재계측화 확인.
- D3 — **RESOLVED**: spec.md:13에 `tier: M` 추가 확인. 프론트매터 13필드(정식 12 + 선택 `tier`), snake_case 별칭 여전히 0건 — MP-3 회귀 없음.
- D4 — **RESOLVED**: plan.md §A가 실제 메커니즘으로 재서술됨.
- D5 — **RESOLVED**: `grep 'NEEDS CLARIFICATION' plan.md` → 0매치(exit 1, 이번 실행). "[NEEDS CLARIFICATION 없음]" 대괄호 형태 소멸, "미해결 질의는 없다 — …" 평서문으로 대체. MP-7은 반복 1보다 더 엄격한 상태(형태 매칭도 0매치).
- D6 (기록 목적) — 변화 없음, 무수리 권고 유지.

## 델타 판정 (D1-D5)

**D1 — AC-004 채택 규율: 충족.** acceptance.md:58의 판정 대상 문단이 도구를 명시한다 — 조항 범위 diff가 판정자, 전체 파일 diff는 적색 관측 수단일 뿐, 제외된 2 hunk의 존속은 실패가 아님. 이로써 도착 시 녹색인 판정 도구의 공허함이 침묵이 아니라 문서화됐다. 두 셀 모두 재실행으로 검증:

- RED-now 셀(acceptance.md:64-70): `diff .claude/agents/moai/plan-auditor.md internal/template/templates/.claude/agents/moai/plan-auditor.md` → 이번 실행 재측정 좌표 `338c338` / `441c441` / `443,444d442` — 셀 기록 좌표와 정확히 일치, `diff -q` 재실행 exit=1 직접 관측. 트리 핀 7835148d3 명시.
- 보조 기준선 셀(acceptance.md:72-79): `grep -c "Side-talk"` 양쪽 사본 → `.claude/agents/moai/plan-auditor.md:0` / `internal/template/templates/.claude/agents/moai/plan-auditor.md:0`, exit 1 — 축자 일치 재현.
- Green 경로 셀(acceptance.md:81): M2가 전환 주체, 합격 출력 = 조항 범위 diff 0, 2 hunk 존속 합법 명시.

§2.1 네 요소(명령/축자 stdout/종료코드/트리 SHA)가 두 셀에 갖춰졌고, 도구의 적색이 알려진 실패 입력(기존 2 hunk 드리프트)에서 관측됐다(verification-completeness §1.1/§1.2(b)). 채택 완료로 판정한다.

- 잔여 관찰(optional, 수리 불요): 전체 파일 diff RED-now 셀의 출력란은 hunk 본문 전체가 아니라 좌표 헤더+주석 형태다. 좌표 헤더 자체는 diff의 축자 출력이고 적색을 유일하게 식별하며, 판정 대상 문단이 이 셀의 역할을 적색 관측으로 명시적으로 강등해 뒀으므로 규율 실질은 충족이다.

**D2 — AC-005 재계측화: 녹색 형태가 실제로 뒤집힌다.** 판정 대상 문단(acceptance.md:83-85)이 저장소 전역 매치를 명시적으로 판정 대상에서 제외(런게이트 스트림의 합법적 잔존 인정). (a) 조건은 쌍둥이 사본 grep `reports/plan-audit/` → 양쪽 0 — 뒤집힘 검증: 각 사본의 유일 출현은 :395 한 곳이고(이번 실행 재측정, 양쪽 모두 동일 라인), REQ-002가 그것을 제거하므로 1→0 전환이 기계적으로 가능하다. RED-now 셀이 :395로 재지향됐고 축자 일치 재현(exit 0). (b)는 문서 판독 조건으로 종속 배치. 녹색 출력 형태가 Then 절에 명시됐다 — 반복 1의 "프록시가 뒤집히지 않음" 결함 해소.

**D3 — `tier: M` 추가.** spec.md:13 확인. 이로써 선언된 Tier와 기계 resolution이 일치(문턱 0.80, 천장 2).

**D4 — §A 메커니즘 서술 정정.** plan.md §A 신규 문단: 전체 재생성 명령 `AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission` — Makefile:39와 축자 일치(반복 1에서 직접 검증한 바와 동일). 스포트 체크: `internal/template/agentemit/golden_test.go:78-82` — "With AGENTEMIT_UPDATE=1 it (re)writes the committed .toml artifacts" + `emitRealSet(t)` + `update := os.Getenv("AGENTEMIT_UPDATE") == "1"` — "커밋된 .toml 전부 재작성" 서술이 코드와 일치. 카탈로그 `gen-catalog-hashes --entry plan-auditor` 및 2549f775f 기록 인용 유지. "범위 emit" 오서술 소멸.

**D5 — 대괄호 부정문 제거.** 위 회귀 검사 참조 — 0매치.

## 회귀 검사 (반복 1 기준 유지 확인)

- REQ 계층: REQ-001..008이 :49-63(frontmatter tier 추가로 +1행 밀림)에 연속, 문면 반복 1과 동일 — MP-1·MP-2 회귀 없음.
- AC 매트릭스(acceptance.md:7-16)와 AC-006 RED 셀·AC-007 GREEN 셀: 반복 1 판독과 동일 — 변경 없음.
- AC-006 기준선(적색 sync-auditor.toml 1건): 반복 1에서 동일 트리 내용으로 재측정 완료, 수리 커밋은 코드 무변경이므로 유효.
- MP-5: 수리 커밋이 새 SPEC-ID 참조를 추가하지 않음(5파일 diff 대상 확인) — D6 기록 유지, BLOCKING 없음.
- MP-6: `syscall` 추가 없음.

## Iteration 2 Gaps

- run-phase 행위 확인(반출 파일의 카드 디렉터리 생성) — 여전히 plan-phase 관측 불가(§D.5 소관).
- 조항 최종 문안(M1 산출)에 대한 중립성 가드 실행 — run-phase 몫.
- `make agents-emit` 실행은 수행하지 않음(감사 read-only; 선례 end-state로 간접 증명, 반복 1과 동일).

## Iteration 2 Residual-risk

- AC-004의 조항 구간 추출 프록시 구체 구현은 run-phase M2에서 확정 — 추출 경계 오류는 AC-004를 무력화할 수 있으나, 이제 적색 관측 장치와 판정 대상 문단이 있어 무음 고장은 아니다.
- develop 이동 시 AC-006의 "적색 1건" 기준선 재측정 필요 — acceptance.md 재측정 규정이 절차를 갖고 있다.

## 최종 권고

수리 없이 run-phase 진행을 승인한다. 반복 1 차단 결함 D1이 규율상 요구되는 두 셀과 함께 해소됐고, 나머지 결함은 전부 기록·해소됐으며, 신규 표면 변경이 없다.
