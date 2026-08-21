# SPEC Review Report: SPEC-CLI-CLEAN-SYMLINK-001
Iteration: 1/2 (Tier M — `harness.plan_audit_tier_ceilings` M=2)
Verdict: FAIL
Overall Score: 0.92 (harmonic mean; Tier M PASS threshold 0.80 — score 충족하나 blocking-class 결함 2건이 판정을 지배, 하기 서술)

## Method Note

- **audit_multi 미실행(오케스트레이터 지시)**: 2차 백엔드(audit_multi)가 worktree-blind라는 t171 판정 결과 지시에 따라 호출하지 않았다. 본 판정은 전적으로 in-session 증거에 근거한다.
- **M1 Context Isolation**: 저자 추론 컨텍스트는 제공되지 않았다(제공됐다면 무시했을 것). 감사 입력은 Tier M 3산물 + progress.md + 도시에 + 코드 트리뿐이다.
- **판정 구조**: MP-1~MP-7 전부 PASS. FAIL 사유는 must-pass가 아니라 **M6 routing 분류상 blocking-class 결함 2건**(SPEC 내부 모순 1 + 계약 AC의 실현 분기 모호성 1)이며, 둘 다 1줄 수준 수정이다. iter2는 이 2건의 delta 재감사로 족하다.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: `REQ-CSL-001`…`REQ-CSL-012` — 12건 열거(grep 전수), 순차·무공백·무중복. 3자리 zero-padding 일관.
- **[PASS] MP-2 EARS/GEARS format compliance** (요구사항 계층에만 적용 — AC는 검증 계층으로 Given-When-Then가 정규 형식): 12 REQ 전수 대조. 001(While+shall not)·002(When)·003(When)·004(When)·005(Unwanted+Ubiquitous)·006(While)·007(Unwanted)·008(When)·009(When)·010(Where)·011(Where+shall not)·012(Where+shall not) — 전건 5대 GEARS 패턴 중 하나에 정합. 비정형 요구사항 0건.
- **[PASS] MP-3 YAML frontmatter validity**: 12 canonical 필드 전수 확인(id `SPEC-CLI-CLEAN-SYMLINK-001` 다중 세그먼트 도메인 — D7 규칙이 명시적으로 허용, title·version "0.1.0" 인용 semver, status draft, created/updated 2026-08-22 ISO, author, priority P1, phase "v3.1.3 target" — 금지된 lifecycle 단어명 아님, module, lifecycle spec-anchored, tags 쉼표 문자열). 기각 alias(created_at/updated_at/labels/spec_id) 0건. 선택 필수 아닌 tier: M·era: V3R6 전부 유효 enum. Tier M 산물 3종+progress.md — tier 계약 일치(research.md 부재는 Tier M에 합법).
- **[N/A→PASS] MP-4 Section 22 language neutrality**: 단일 언어(Go) 프로젝트 내부 CLI 경로에 한정된 SPEC — 16언어 템플릿 도구망 다루지 않음. 자동 통과.
- **[PASS] MP-5 D7 cross-SPEC reconciliation**: 참조 SPEC-ID 2건(자기 자신 + `SPEC-CODEX-SKILLS-CANONICAL-001`). 후자는 본 워크트리 `.moai/specs/`에 부재(D7-5 SHOULD 조건)하나, SPEC §F가 명시한 절대 경로(t81 워크트리)에서 **직독 확인**했다 — v0.7.0, `status: draft`(retired/superseded/archived 아님 → D7-4 BLOCKING 없음). 상세는 D6.
- **[PASS] MP-6 D8 cross-platform discipline**: `grep -c syscall` → spec.md 0 / plan.md 0 / acceptance.md 0. 자동 통과.
- **[PASS] MP-7 clarification gate**: `grep -rn '\[NEEDS CLARIFICATION' plan.md` → **0건**(exit 1). research.md 부재(Tier M — N/A 부분). progress.md의 2건 hit은 폐지된 마커를 언급한 서술("폐지되어 DECIDED 항목으로 대체됐다")이지 미해결 마커가 아니다.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 1-2건 미묘한 모호성 (합리적 엔지니어가 일관되게 해결 가능하나 해석 분기 실재) | D1(spec.md:81 "배치 2곳" — 자기 표·형제 산물과 모순), D2(AC-CSL-009 "템플릿이 보유하는 경로" — raw FS vs 렌더링 결과 경로) |
| Completeness | 1.0 | 전 섹션·전 필드 구비 | HISTORY(spec.md:20-33)·WHY(§A)·WHAT(§B)·REQUIREMENTS(§C, 12 REQ)·AC(acceptance.md, Tier M 2층 구조)·Out of Scope(§E — H3 소제목 5개 각각 구체적 불릿, spec.md:202-225)·frontmatter 12+2 필드 |
| Testability | 1.0 | 전 AC 이진 판정 가능, weasel word 0건 | 11 AC 전수 WalkDir-스킵 공허함 트랩 명시적 봉쇄(AC-CSL-004.4 "보조축으로만", acceptance.md:91-94)·양극 쌍 실재(AC-CSL-002 ↔ AC-CSL-007, AC-CSL-006.2 부재 단언)·"성공 추정"은 §D.3에서 run-phase 확인 항목로 정직 마크 |
| Traceability | 1.0 | 12 REQ 전부 ≥1 AC, 역방향 완전, 고아 0건 | acceptance.md §D.2 표(168-179행): REQ-CSL-001…012 전건 매핑, AC-CSL-001…011 전부 유효 REQ 지침. 마일스톤 닫힘 조건 합집합 = AC 11건 전량(plan.md §F M1-M4) |

## Ground-Truth Verification (판정 근거 — 전부 이 트리 HEAD 075672146에서 직접 관측)

1. **코드 앵커 무드리프트**: `git diff --stat 4b2f203fe..HEAD -- internal/ cmd/ pkg/` → **빈 출력**(이후 2커밋은 docs뿐). 도시에의 전 file:line 인용이 현 HEAD에서 그대로 유효. 직독 확인: `ManagedCleanTargets` deploy.go:50-82(7뿌리, 글로브 1개 `moai*` — 도시에 §2.1 표와 정확 일치)·`CleanMoaiManagedPaths` :101·글로브 팔 :115-137(Glob :116, 호출 :128)·루트 Stat :139 + Skip :140-146·config 제거 :168-182(사전검사 없음, :176 호출)·`backupThenRemove` :371-399(:372 os.Stat, :374-375 no-op, :380 파일 분기, :390 디렉터리 분기)·`backupUnmanagedTree` :435-459(**:441 `d.IsDir() || !d.Type().IsRegular()` 스킵** — WalkDir 루트 Lstat 확인)·`copyRegularFile` :465(os.ReadFile)·deployer.go `MkdirAll` :189-190(`template deploy mkdir %q` — Run D verbatim 에러와 일치)·`atomicWriteFile` :19-32(rename :28)·forceUpdate 존재검사 우회 :169-185.
2. **FX-1 MkdirAll fast-path 근거(제거+가시화 처분의 결정 근거) — 실증**: GOROOT `os/path.go` `MkdirAll` 직독 — fast-path가 `Stat`(링크 추적)으로 판정: 라이브 디렉터리 링크에서 `IsDir()==true` → **nil 반환** → 보존된 링크 하위 destDir 체인이 전부 사용자 외부 트리로 흘러들고 `atomicWriteFile`의 tmp+rename이 같은 장치에서 성공하며 기록이 외부 트리에 착지. dangling에서는 slow-path `Mkdir` EEXIST(Run D 오류 메커니즘과 동일체). spec.md §B.1의 "보존은 배포 기록이 사용자 외부 디렉터리 내부로 유입된다" — 코드로 정확. 처분(1) 채택 근거 성립.
3. **FX-3b 제거의 무데이터-손실 주장 — 성립**: dangling 링크의 대상은 부재(정의) → 링크 제거로 잃을 대상 데이터가 없다. 링크 자체(사용자 의도)의 상실은 plan.md §D-5에 **[HARD] 새 회귀면으로 명시**돼 있고 양극 픽스처(AC-CSL-002 must-remove ↔ AC-CSL-007 must-not-remove `badlink`)로 고정. 관리 네임스페이스 일관성: 글로브 `moai*` 아래 실 디렉터리는 현재도 백업 후 제거되므로 dangling 제거가 더 파괴적인 것이 아니다.
4. **FX-2 파일 링크 백업 — 성립**: :381 templateCarries(".claude/settings.json")=false(템플릿은 `.tmpl`만 보유 — 도시에 §1.1(b) ls 실측 + deployer.go:133/150 렌더링 경로 확인) → :384→:465 ReadFile 추적 백업 → RemoveAll 링크만 제거. Run B 실측(OUTSIDE-SETTINGS-v1 백업+복원)과 일치.
5. **Run D 실패 경로 — 코드와 정합**: :139 Stat→ENOENT→Skip(링크 잔존) → deploy :189 MkdirAll(".claude/agents/moai") fast-path Stat ENOENT → slow-path Mkdir EEXIST → Go의 재확인 Stat도 ENOENT → 에러 반환. verbatim 출력(`mkdir .claude/agents/moai: file exists`)과 문자열 단위 일치. "재실행 영구 루프"는 상태 비의존적 코드 경로 추적(도시에 gap 2에 정직 마크, run-phase 실측 전환 예정 — plan §C).
6. **t81(가) 인용 정확성 — 직독 확인**(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t81/.moai/specs/SPEC-CODEX-SKILLS-CANONICAL-001/spec.md`): v0.7.0·status draft. REQ-CSC-004(:200 복사 폴백)·REQ-CSC-005(:201 모드+경고 반환 — t173 §A의 "경고로 보고한다" 압축 서술은 의미 정확)·REQ-CSC-014(:208 실 항목 보존)·§D "폴백 플램폼 미러 고착" 고지(2회차 배포부터 고착 + 경고의 오귀속 — t173 §A 보충 항목 4의 서술과 문단 단위 일치). **판별자 부재도 t81 스스로 확인**("§C에는 … 판별자가 없으므로"). t81 §D의 "승계 SPEC(t173)이 닫는다" 문장과 리드 결정(아무도 소유 안 함, 리드 큐 후보)의 충돌은 progress.md에 명시 기록+해소(리드 몫 정리) — 은폐 없음. t81 감사 보고서 앵커도 실재·일치(plan-audit-iter4.md 273줄, D1 :95·실측 :117-123·D2 :137·D3 :160·D4 :173·판정 :258).
7. **양극/비공허(D4 이관) 점검 — 성립**: "백업 수==0"은 전 AC에서 단독 단언이 아니라 보조축(AC-CSL-004.4). AC-CSL-001은 5단언(링크 Lstat ENOENT + 메시지 + 무중단 + 재배포 + 재실행) 결합. AC-CSL-006.2/007.2가 must-not-flag 극으로 오탐 포착. 11 AC 모두 구현으로 반증 가능(무효 구현 시 001/002/003은 즉시 RED).
8. **수치 규율(D3 이관)**: 12 REQ·11 AC·5 형태가 spec(§C:124, §F:242)·plan(:20-23, §D-4 [HARD])·acceptance(:20-21, §A)·progress(§E.1) 4면 동일. 마일스톤 닫힘 조건의 AC 합집합 = 11건 전량, 잔여 없음. **단, 하위-카운트 1건 드리프트 발견 → D1.**
9. **리드 비준(DECIDED) 대조**: 리드 원문은 입력에 없어 verbatim 대조는 **UNVERIFIED**(M4 규율상 패스라고 단정하지 않는다). 대신 4산물 간 정합은 확인: progress.md "리드 결정 확정"(1 FX-1 제거+가시화, 2 FX-3b 제거, 3 판별자 라우팅 정정, 4 조건부 Kickoff) ↔ plan.md §D-5 DECIDED 2건 ↔ spec.md §B.1/§E/HISTORY — 내용 수준에서 상호 불모순.

## Defects Found

**D1. count-drift** — spec.md:L81-82 — "dangling은 **배치 2곳**이 같은 형태의 **두 진입 팔**이다"가 자기 문서의 §B 표(FX-3a/3b/3c = **3배치**, L88-90)·plan.md:20-23("3배치")·acceptance.md:30("3배치")·progress.md:98("3배치")와 모순. 5-형태 상위 카운트 자체는 4면 일치하나, 이 문장은 정확히 이 SPEC이 [HARD]로 못박은 D3/AP-2 드리프트 계급의 하위-카운트 오류다(plan.md §D-4, §G AP-2). — Severity: major — Class: **blocking** — Required fix: L81-82를 "dangling은 배치 3곳(비-글롭 뿌리·글로브 매치·config 뿌리)이 같은 형태의 진입 팔이다(이 중 실측은 2, config 팔은 코드 추적 — 도시에 §2.1 gap 4)"로 정정.

**D2. contract-AC 용어 미고정** — acceptance.md:L142-144(AC-CSL-009 Then-1) + plan.md:L142-143(M3) — "모든 비-글롭 청소 루트는 **템플릿이 보유하는 경로**다"의 "보유"가 (a) 임베디드 FS 원시 경로 (b) 렌더링 후 기록 경로(.tmpl 제거) 사이에서 미고정. **이 SPEC 자체의 증거가 (a) 독해를 반증한다** — 도시에 §1.1(b) 실측: templateCarries(".claude/settings.json")=false(템플릿은 `.claude/settings.json.tmpl`만 보유). (a)로 구현하면 루트 1에서 계약 테스트가 즉시 실패하고, 구현자가 테스트를 "고치는" 방향(파일 루트 제외)으로 약화하면 REQ-CSL-009가 조용히 파일 루트 적용을 잃는 실현 분기가 실재한다. — Severity: major — Class: **blocking** — Required fix: AC-CSL-009와 plan M3에 구성원 판정을 명시 — "보유 = 렌더링 후 기록 경로 집합(`.tmpl` 접미사 제거 포함)" — 하거나 파일 루트의 렌더링 경유를 명시적 예외 조항으로 적는다.

**D3. AC 커버리지 구멍(config 뿌리 진행줄)** — acceptance.md:L73-80(AC-CSL-003) — config 뿌리 dangling의 AC에 진행줄 단언이 없다. REQ-CSL-002("어디든 … 진행줄을 출력하며(shall)")와 REQ-CSL-005(전 링크 제거에 진행줄)가 이를 요구하는데, config 뿌리의 진행줄만 빼먹은 구현이 **11 AC 전부 통과**한다(REQ 위반 상태로). — Severity: minor — Class: optional(권장 수정) — Required fix: AC-CSL-003에 Then-4 "진행 출력에 해당 경로와 dangling임을 이름붙인 줄이 존재한다" 추가.

**D4. RED 라벨 귀속 과잉** — acceptance.md:L60 — AC-CSL-001 RED 기준이 "단언 1·3·5가 실패한다(**Run D 실측**)"로 적혀 있으나, 단언 5(재실행)는 실측이 아니라 코드 추적이다(도시에 §3.4·§5 gap 2가 스스로 명시 — 예산 4회 소진). RED Go 테스트가 관측하면 해소되나 문서 단계에서의 귀속 표기는 느슨하다. — Severity: minor — Class: optional — Required fix: "(Run D 실측; 5는 코드 추적 — M1 RED에서 직접 관측)"로 보강.

**D5. 배포 재생성 조항의 서술张力(모순 아님)** — spec.md:L142(REQ-CSL-003 말단 "배포는 … 실제 디렉터리를 재생성해야 한다(shall)") vs §D:197-198("배포 쪽 동작을 바꾸는 요구사항은 없다") — 링크 제거 후 MkdirAll의 현행 동작(도시에 §2.3, Run B) 재진술로 읽히므로 모순 아니다(의무 ≠ 동작 변경). — Severity: minor — Class: optional — Required fix: 없음(원하면 "(현행 동작 재진술)" 각주).

**D6. D7-5 SHOULD — 교차 워크트리 참조** — spec.md §F — 참조 `SPEC-CODEX-SKILLS-CANONICAL-001`이 본 워크트리 `.moai/specs/`에 없다(프로토콜상 SHOULD). 그러나 §F가 명시한 t81 워크트리 절대 경로에서 직독 확인(v0.7.0, draft → D7-4 BLOCKING 없음)했고 SPEC 스스로 "본 트리에 없음, 읽기 전용 참조"로 자기 기술했다. 병렬 카드 워크플로의 구조적 상황. — Severity: minor — Class: optional — Required fix: 없음. release/v3.1.3 통합 시점 §D.7-1 재실행에서 해소.

**D7. 라벨 조어** — spec.md 전반(~15회) — 도시에 dossier를 지칭하는 "도시에"라는 조어(문자적으로 "도장(都市)에")가 어색하나 첫 사용 시 정의(HISTORY :22-23)되고 전 문서 일관 단일 지칭이라 해석 분기는 없다. — Severity: minor — Class: optional — Required fix: 없음(원하면 "dossier(도시에)" 병기로 1회 정의 강화).

## Recommendation (FAIL — 수정 후 iter2 delta 재감사)

1. **[D1]** spec.md:81-82 하위-카운트 정정("배치 2곳/두 진입 팔" → 3배치 서술 + 실측/추적 구분). 1줄.
2. **[D2]** acceptance.md AC-CSL-009 Then-1 + plan.md M3에 "보유" 판정 집합 정의(렌더링 후 경로) 추가. 1-2줄.
3. **[D3](권장)** AC-CSL-003에 진행줄 단언(Then-4) 추가 — REQ-CSL-002/005의 어디든 조항을 AC가 실제로 걸게.
4. **[D4](권장)]** AC-CSL-001 RED 라벨 귀속 보강.
5. 수정 후 REQ/AC/형태 수치 12/11/5 불변을 유지할 것(위 수정은 전부 수치 불변). iter2는 D1+D2(권장사항 반영 시 D3·D4까지)의 delta 확인 + 본 보고서 회귀 체크로 충분하다(Tier M 상한 2 — 잔여 1회).

## Positive Findings (차기 감사 회귀 기준선)

- 처분표 전 행이 도시에 실측/추적과 정합 — FX-1(MkdirAll fast-path 직독 실증), FX-2(ReadFile 추적 백업 실측), FX-3a(Run D verbatim), FX-3b(무데이터-손실 + 회귀면 [HARD] 공개 + 양극 고정), FX-3c(trace-only 정직 마크 + AC-CSL-003 실측 전환), FX-4/5(대조 극).
- 스코프 펜스: t81(가) 미러 배포 배제(§E)·판별자 "리드 큐 후보" 라우팅(§A/§E + progress 충돌 기록까지 공개)·배포 측 게이팅 배제(§B.1 근거와 함께)·.moai/config 8번째 뿌리 매트릭스 포함(trace 상태 명시).
- REQ-CSL-008의 "최종 트리 상태" 한정 — 순서에 따라 진행줄 형태명(live vs dangling)이 갈리는 미세점을 트리 상태로 정확히 좁힘.
- 12 REQ GEARS 정합·11 AC 이진 판정·추적표 완전·마커 0·syscall 0·프론트매터 완전.
