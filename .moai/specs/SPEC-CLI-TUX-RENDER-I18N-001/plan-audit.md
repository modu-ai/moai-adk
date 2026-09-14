# SPEC Review Report: SPEC-CLI-TUX-RENDER-I18N-001

Iteration: 1/2 (tier M — `.moai/config/sections/harness.yaml` `plan_audit_tier_ceilings.M = 2`)
Verdict: **FAIL**
Overall Score: 0.925 (Clarity 0.95 · Completeness 0.75 · Testability 1.0 · Traceability 1.0)

> 판정 근거 요약: must-pass 7항목(MP-1~MP-7) 전부 통과였고 러브릭 점수도 Tier M 역치(0.80)를 넘지만, 차단(class: blocking) 결함 D1이 있다 — M2 의 유일하게 명명된 수리 경로(Inline(true))가 라이브러리 소스 구조상 요구사항 자체의 목표 형상과 양립하지 않는 것이 본 감사의 소스 판독으로 확인됐고, 대체 경로(fallback)가 계획 어디에도 명명돼 있지 않다. 차단 결함이 미해결인 상태에서의 PASS 는 러브릭 점수로 상쇄될 수 없다(M6). 러브릭 점수는 참고치이며, 본 FAIL 은 결함 목록 D1-D3 의 수리 후 iteration 2 delta 재감사로 재판정한다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- **Claim**: 본 SPEC 의 plan 산출물 4종(spec.md·plan.md·acceptance.md·progress.md)은 카드 t756 의 요구를 잔여 수리 + 가드 + 캡처 재검증으로 올바르게 재범위했으며, must-pass 7항목을 통과한다. 단, M2 수리 경로의 실행 가능성과 증거 기록의 사실 정확성에서 차단 결함 3건이 있다.
- **Evidence**: 이하 Must-Pass Results · Category Scores · Defects Found 의 전 항목. 핵심 관측 — (1) huh v2.0.3 `field_confirm.go` 모듈 캐시 판독: `:261` `if !c.inline && wroteHeader`, `:262-263` `sb.WriteString("\n")` ×2, `:254` inline 시 제목↔설명 개행 생략, `:293` buttonsRow 앞 개행 없이 직접 append, `:53` `buttonAlignment: lipgloss.Center`, `:136` `Inline`, `:361` `WithButtonAlignment`. (2) 본 감사 실행: `go test ./internal/cli/wizard/ -run 'TestLayout|TestOptionColumn|TestDrawInit' -count=1 -v` → `TestLayout_NoBlankBetweenFields` 1건만 스윕(하위테스트 init-first-page·profile-all-groups) PASS. (3) `MOAI_PTY_CAPTURE=1 go test ./internal/cli/wizard/ -run '^TestPtyCapture_' -count=1 -v` → NormalRun·SkipWithoutGate·FailWithoutTmux 3/3 PASS. (4) `go test ./internal/cli/ -run 'TestEmitAcceptEditsConfirmationAnchor' -count=1` PASS.
- **Baseline-attribution**: 이번 실행, 이 트리(HEAD `0ab6cf281`, branch `WT-tux-render`, base `a404132e7`). huh 라이브러리 판독은 `charm.land/huh/v2@v2.0.3` 모듈 캐시(`go.mod:9` 고정 버전) 대상.
- **Gaps**: (1) M1 이전 기준 프레임(수리 전 확인 필드 캡처)을 직접 측정하지 않았다 — AC-TRI-003 의 "수리 전 = 2" 값은 SPEC 기록을 인용했을 뿐이고 본 감사의 inline 해석은 소스 기반 판독이다. (2) `.moai/reports/init-tui-audit-20260909.html` §1(primary 체크아웃)은 읽지 않았다 — 트리 측 앵커를 전부 독립 검증했으므로 판정에 불요했다. (3) huh `ThemeBase` Title 스타일의 margin 유무(빈 행 2번째의 기여원)는 추적하지 않았다 — M1 캡처가 실측으로 대체한다.
- **Residual-risk**: (1) D1 수리(래퍼 경로)는 현재 계획보다 구현 물량을 늘린다 — Tier M 범위(5-15 파일) 안에는 유지된다. (2) huh 버전이 올라가면 D2 의 재고정 라인 번호도 다시 스테일해진다(moving-ref 성질). (3) M1 실측이 잔여 빈 행을 1로 판독하면 AC-TRI-003 의 전제값(2) 조정이 필요하다 — AC 구조상(수리 전 값 먼저 관측) 안전하게 흡수된다.

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성**: REQ-TRI-001~008 이 spec.md:59-66 에 순차 배치, 결번·중복 0. zero-padding 일관.
- **[PASS] MP-2 GEARS 형식 준수 (요구 계층)**: 판정 대상은 spec.md §C 의 REQ-XXX 8건(검증 계층 AC 는 M3 § Scope 에 따라 본 항목에서 제외 — acceptance.md §E 의 Given-When-Then 은 AC-1 에서 별도 판정). 8건 전부 5 GEARS 패턴 적중: Event(REQ-TRI-001 "When run phase 가 시작되면…shall", REQ-TRI-008 "When 기준 캡처가…shall"), State(REQ-TRI-004 "While consecutive fields render…the wizard shall"), Ubiquitous(REQ-TRI-002/003/005/006/007 "The …shall"). IF/THEN 0건, 비격식 서술("should"·"may") 0건.
- **[PASS] MP-3 YAML Frontmatter 유효성**: 12 정규 필드 전부 존재·타형 적중(spec.md:2-15) — id/title(quoted)/version("0.1.0")/status(draft)/created·updated(2026-09-14)/author/priority(P2)/phase("v3.2.0 target" — 금지 라이프사이클 토큰 아님)/module/lifecycle(spec-anchored)/tags(CSV 문자열). snake_case 별칭 0건. `id: SPEC-CLI-TUX-RENDER-I18N-001` 은 `internal/spec/lint.go:1146` `specIDPattern = ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` 적중(다중 세그먼트 허용 확인). 선택 필드 `tier: M` 존재.
- **[N/A] MP-4 Section 22 언어 중립성**: 단일 언어(Go) 프로젝트 대상 SPEC — 자동 통과.
- **[PASS] MP-5 D7 Cross-SPEC 조정**: 본문+frontmatter 이 참조하는 4개 SPEC 전부 존재·`status: completed` 확인 — SPEC-INIT-TUX-I18N-001, SPEC-CLI-WIZARD-RESTRUCTURE-001, SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-I18N-GOVERNANCE-001. retired/superseded/archived 없음 → BLOCKING 없음. 참고로 최대 겹처(SPEC-INIT-TUX-I18N-001 와의 흡수 관계)는 §A.1 대응표로 명시 조정돼 있다.
- **[PASS] MP-6 D8 Cross-Platform 규율**: SPEC 디렉터리 전체 `grep 'syscall'` 0건 → 자동 PASS.
- **[PASS] MP-7 clarification 게이트**: `grep '\[NEEDS CLARIFICATION'` plan.md 0건. research.md 부재(Tier M — 불요 아티팩트)로 N/A 경로, plan.md 가-clean 으로 충족. progress.md §E.1 의 `needs_clarification_markers: none` 과 일치.

## Category Scores (0.0-1.0, 러브릭 앵커)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.95 | 1.0 역 — 경미한 모호성 1건 | 전 REQ 단일 해석·정량 기준(빈 행 ≤1, 빈 줄 0, 표시 열 동일, 잔존 0). 감점: REQ-TRI-003(spec.md:61)의 "(인라인 모드 등)"이 명명하는 지점의 실제 레이아웃 의미가 요구의 목표 형상과 어긋남(D1 — 라이브러리 소스 판독으로 확인) |
| Completeness | 0.75 | 0.75 역 — 계획 본질 내용 결손 | frontmatter 12/12, 전 섹션 존재(HISTORY:22, §A:28, §B:51, §C:57, §D:68, §E:75, §F:79-91 — OoS H3 ×3 + 불릿). 감점: M2 중심 수리에 실행 가능한 명명 경로·fallback 부재(D1), 증거 기록 사실 오류 2건(D2·D3) |
| Testability | 1.0 | 1.0 역 | AC 9건 전부 이진 판정(프레임 행 수·열 동일·뮤턴트 FAIL·수리 diff 0). weasel word 0건. AC-TRI-003 의 전제 단정 구조는 inline 모드의 공허 통과(0행 몰아넣기)도 원리적으로 막는다 |
| Traceability | 1.0 | 1.0 역 | acceptance.md §F 매핑이 REQ 8↔AC 9 전체를 쌍방향 커버, 고아 0. 선행 SPEC 인용(REQ-ITI-001/014/015/016, AC-ITI-015/016/017/019)을 SPEC-INIT-TUX-I18N-001 실물에서 대조 확인(acceptance.md:82,141-143,150,188) |

## Defects Found

D1. **TRI-D1** — plan.md:52(M2) / spec.md:61(REQ-TRI-003) — M2 의 유일하게 명명된 수리 지점 `Inline(true)` 가 구조적으로 요구 형상과 양립하지 않는데 fallback 이 명명돼 있지 않음 — Severity: **major** — Class: **blocking** — Required fix: plan.md M2 에 대체 경로를 명명하라(예: `huh.Field` 인터페이스를 구현한 로컬 Confirm 래퍼 — 자체 View 합성으로 제목 줄·설명 줄·버튼 줄 분리를 유지하면서 빈 행 ≤1 달성). Inline 모드는 "관측용/예상 실패"로 강등하고 REQ-TRI-003 의 "(인라인 모드 등)"도 정정하라. **근거(본 감사 소스 판독)**: huh v2.0.3 `field_confirm.go` 에서 inline 모드는 (a) `:254` 조건으로 제목↔설명 개행을 생략하고, (b) `:261-263`의 `"\n\n"`을 생략한 뒤, (c) `:293`에서 buttonsRow 를 개행 없이 헤더 문자열에 직접 append 한다 — 즉 inline 시 제목·설명·버튼이 하나의 묶인 줄이 되고(width 래퍼가 임의 위치에서 접는다), "제목 줄과 버튼 줄 사이 빈 행 ≤1"(줄이 분리돼 있다는 전제 포함)과 양립하지 않는다. AC-TRI-003 의 전제 단정(제목 줄·버튼 줄 식별)이 이를 잡아 FAIL 로 전락하며, §D 가 upstream 패치를 금지하므로 M2 는 사전에 합의된 대안 없는 dead-end 에 진입한다.

D2. **TRI-D2** — spec.md:37 / plan.md §B(:18) / progress.md §C·§D(6행) — huh 고정 빈 줄 인용 `field_confirm.go:270-274` 가 실제 위치와 어긋남 — Severity: **minor** — Class: **blocking** — Required fix: `:261-263` 으로 재고정(조건 `:261`, writes `:262-263`, 블록 닫음 `:264`). **근거**: 모듈 캐시 `charm.land/huh/v2@v2.0.3/field_confirm.go` 직접 판독. 메커니즘 주장 자체(테마로 제거 불가한 하드코딩 `"\n\n"`, `:136` Inline 이 우리 쪽 지점)는 전부 참으로 확인됐다 — 포인터만 ~9-12줄 어긋나 있다. 스테일 참조 재고정을 사명으로 하는 SPEC(§E.1 `stale_ref_resolution: all card refs re-anchored`)의 재고정표 자체에 부정확 인용이 남는 것은 내적 일관성 결함이다.

D3. **TRI-D3** — progress.md §D 1행 — pre-flight 선택자가 존재하지 않는 검사 기호 2개를 명명 — Severity: **minor** — Class: **blocking** — Required fix: 선택자를 실제 검사명으로 정정(예: `TestLayout_NoBlankBetweenFields`). **근거**: `TestOptionColumn`·`TestDrawInit` 는 `internal/cli/wizard/` 전체에서 0건(grep). 본 감사가 동일 선택자를 `-v` 로 실행한 결과 스윕은 `TestLayout_NoBlankBetweenFields` 1건뿐(하위테스트 2개) — 기록된 `ok 0.607s` 는 실측이지만(본 감사 재현: `ok … 0.617s`) 명명된 3 패밀리 중 1/3 만 스윕한 green 이다(verification-completeness.md §1.1 부분 공스윕). 향후 재실행이 기록보다 넓은 커버리지로 오독될 지점.

Nits (class: optional — 판정 불변):
- N1 — acceptance.md 섹션 문자가 §A → §C 로 건너뛴다(§B 부재). 편집 잔재로 보인다.
- N2 — AC-TRI-007 이 검사 위치를 "run 단계에서 확정"으로 미룬다. 그러나 `TestEmitAcceptEditsConfirmationAnchor` 는 이미 `internal/cli/profile_setup_acceptEdits_test.go:21` 에 존재하고 본 감사 실행에서 PASS 였다 — 지금 확정 가능.
- N3 — plan.md §B 가 앵커 토큰 계약을 "REQ-CCI-006/AC-CCI-006"으로 소유 SPEC ID 없이 인용한다. (`profile_setup.go:30` 주석에서 소유 계약을 찾을 수는 있다.)

## 카드 t756 요구 검증 결과 (7 claims)

1. **Overlap scoping — 참으로 확인**. §B 전 앵커를 이 트리에서 라인 단위로 대조: `wizard.go:299` `optionColumnWidthCap = 80` ✓, `:317` `alignOptionLabels` ✓, `:544` `huh.NewConfirm` + `:555` `WithButtonAlignment(lipgloss.Left)` ✓, `:645` `FieldSeparator = "\n"` ✓, `downgrade_confirm.go:26`+`:31` ✓, `init.go:617` REQ-ITI-001 주석("NO profile entry") ✓, `profile_setup.go:330 부근` `profileWizardRunner` + `getProfileText` 로케일 해석 ✓. 비테스트 코드의 `huh.NewConfirm` 생성 지점은 정확히 2곳(grep). 카드 6판정 전부 AC 로 커버(누락 0): F11-(1)→REQ-TRI-002+AC-TRI-001/002, F11-(2)→REQ-TRI-003/004+AC-TRI-003/004, F11-(3)→REQ-TRI-005+AC-TRI-005, F10-(1)→REQ-TRI-006+AC-TRI-006(+007), F10-(2)→AC-TRI-001(6판정 대응)+AC-TRI-006, F10-(3)→REQ-TRI-007+AC-TRI-008. REQ-TRI-008 은 well-formed — AC-TRI-009 의 "수리 diff 0" 커밋 범위 확인이 올바른 코드의 수리를 기계적으로 막는다(빈 행 아님).
2. **잔여 결함 주장 — 실질 참, 인용 오류**. `"\n\n"` 하드코딩·테마 불가·Inline 지점(`:136`) 모두 소스로 참 확인. 측정 게이트 명명됨(M1 잔여 표가 M2/M3 범위 확정, progress.md D3 "M2 에서 캡처로 확정"). **fallback 미명명(D1)** — 더하여 inline 지점의 구조적 위험도 소스로 확인(D1).
3. **카드 [HARD] 준수 — 통과**. 렌더/i18n 수리 판정 AC(TRI-001/003/004/005/006/009) 전부 PTY 프레임 판정. TRI-002/008 은 뮤턴트 관측 가드로 카드 HARD("수리 판정은 캡처")의 대상이 아닌 별개 계층(합리적 판독). 캡처 명령이 참조하는 기존 기호 실재 확인: `TestPtyCapture_NormalRun/SkipWithoutGate/FailWithoutTmux`(ptycap_test.go:45/79/86) — 본 감사 실행 3/3 PASS(3.158s). `init-first-page` 케이스 실재(ptycap_test.go:20). `TestConfirmAlignmentSweep` 는 신규 예정 검사로 적히고 M2 산출 목록과 일치(기존/신규 구분 가능). 프레임 수출처가 SPEC 디렉터리 아래(evidence/)로 명시.
4. **Stale-ref purge — 통과**. 카드 09-09 참조(`huh_theme.go`·`601-602`·`wizard.go:549`·`288-292`·`260-265`·`field_confirm.go:64`)는 plan.md §B 매핑 열에만 존재(해당 파일 제외 grep 0건). REQ/AC 본문에 file:line 인용 0건 — `huh.NewConfirm`·앵커 토큰(`acceptEdits`·`settings.local.json`)은 관측 가능한 계약의 일부로 허용 범위.
5. **D1 결정 항목 — well-formed**. 권고(v2 흡수 유지)+근거+운영자 확인 지점(Kickoff 게이트) 갖춰짐. 전제 재검증: `go.mod:9` `charm.land/huh/v2 v2.0.3` 만 존재, `charmbracelet/huh"`(v1) import grep 0건 — huh v1 부재 확인.
6. **형식·등급·마일스톤·§E.1 — 통과**. GEARS 8/8(MP-2). `tier: M` = 카드 Tier M 일치, 하네스 standard. REQ 8·AC 9 각각 Tier M 한도(16) 이내. M1(캡처)→M2(렌더)→M3(i18n)→M4(종결) 순서 타당. progress.md §E.1 audit-ready 신호 13필드 완비, `plan_status: audit-ready` 기록.
7. **위생 — 통과**. `.moai/reports/` 쓰기 없음(감사 보고서는 primary 체크아웃 읽기전용 인용으로 명시). 한국어 산문+영어 식별자 준수. OoS 3건이 카드 요구를 배제하지 않음(기존 4 로케일 빈칸 메우기는 명시적으로 범위 안).

## Recommendation (manager-spec — iteration 2 수리 경로)

1. (D1) plan.md M2 에 fallback 을 명명한다: huh.Field 를 구현한 로컬 Confirm 래퍼(자체 View 합성 — 3표면 행 분리 유지 + 빈 행 ≤1)를 기본 경로로, Inline(true) 는 관측/검증용으로 강등. 근거는 huh v2.0.3 `field_confirm.go` `:254`·`:261-263`·`:293` 판독(본 보고서 D1). spec.md REQ-TRI-003 의 "(인라인 모드 등)" 표현도 함께 정정.
2. (D2) huh 고정 빈 줄 인용을 4곳(spec.md:37, plan.md §B, progress.md §C, progress.md §D 6행)에서 `:270-274` → `:261-263` 으로 재고정.
3. (D3) progress.md §D 1행 선택자를 `TestLayout_NoBlankBetweenFields`(실재 검사)로 정정. TestOptionColumn·TestDrawInit 삭제.

세 수정 모두 1-2줄 규모다. 수리 후 iteration 2 는 본 결함 델타(D1-D3)만 재판정한다.
