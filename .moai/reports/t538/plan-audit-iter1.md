# SPEC Review Report: SPEC-DOCS-LOCALE-PARITY-REPAIR-001 (card t538)

- Iteration: 1/2 (Tier M — `harness.plan_audit_tier_ceilings` M=2)
- Verdict: **FAIL**
- Overall Score: **0.71** (조화평균; 산술평균 0.75) — Tier M PASS 문턱 0.80 미달 + 차단 결함 6건
- 감사자: plan-auditor (독립 재측정 — 저자 추론 무시, M1 Context Isolation 적용)
- 입력 권위: `.moai/reports/t538/plan-phase.md` (측정 SSOT — SPEC 과 어긋나면 이 파일이 이김)
- 측정 도구: 전 수치 `/usr/bin/grep` (셸 grep = ugrep 래퍼 회피). 커밋 직전 요구 존중 — 읽기 전용 재측정만 수행.

## Must-Pass Results

- **[PASS] MP-1 REQ 번호 일관성** — REQ-001~REQ-013 연속, 갭·중복 0 (spec.md:73-123, `### REQ-00N` 헤딩 전수 확인).
- **[PASS] MP-2 GEARS 형식 (REQ 계층 판정)** — 13건 전부 shall / shall-not / When / Where 형태. REQ-011 Where [HARD] (spec.md:113-115), REQ-001/003/007/010 When, REQ-002/005/009/012 단언형. 판정 계층: `spec.md §D` REQ 항목 — AC 는 Group 4 로 별도 채점. 3건(REQ-004 :85-87, REQ-008 :101-103, REQ-013 :121-123)이 한 요구 안에 shall+shall-not 복합 — 비격식 아닌 형식 내 복합이라 MP-2 는 통과하되 선택 결함 F3 으로 분리.
- **[PASS] MP-3 YAML 프론트매터** — 12 정본 필드 전부 존재·타입 정상, 금지 별칭(created_at 등) 0 (spec.md:1-16). `related_specs:` 는 스키마 외 여분 키 — `internal/spec/lint.go` 에 바인딩 참조 0 이라 디코더가 무시(관측만, 결함 아님).
- **[N/A] MP-4 언어 중립성** — 프로그래밍-언어 도구 나열 SPEC 이 아님(4-로케일 문서 콘텐츠 수리). N/A 자동 통과.
- **[PASS] MP-5 D7 교차-SPEC 조정** — 참조 3건(SPEC-DESKTOP-NATIVE-E2E-001, SPEC-DOCS-V313-CATCHUP-001, SPEC-DOCS-CODEX-WIRING-CALLOUT-001) 모두 `.moai/specs/` 에 실재, `status: completed` ×3 → retired/superseded/archived 없음, reconciliation 의무 미발생. BLOCKING 없음.
- **[PASS] MP-6 D8 크로스플랫폼** — spec.md 내 `syscall` 0건 → 자동 통과.
- **[PASS] MP-7 clarification 게이트** — plan.md `[NEEDS CLARIFICATION` 0건. research.md 는 Tier M 아티팩트 셋에 없음(부재 — MP-4 선례대로 N/A, 사유 명시).

## Category Scores (0.0-1.0, 러브릭 앵커)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 밴드 (1-2개 요구의 경미한 모호) | 정밀한 행-앵커 REQ 다수(예: REQ-001 :75, REQ-007 :99). 단 REQ-012 zh 토큰(D4), REQ-010 배치 체인(D6), REQ-005 기술(F5) |
| Completeness | 1.0 | 1.0 밴드 | 12 필드(spec.md:1-16) · HISTORY :20-24 · §B/§C · §D 13 REQ(한도 16) · OOS H3 4개+불릿 :132-150 · acceptance.md 11 AC(한도 16) · progress.md §E.1 :6-25 |
| Testability | 0.50 | 0.50 밴드 (다수 AC 판정 불가) | AC-006 기댓값 오류(D1) · AC-005 호스트-OS 절 무감각(D2) · AC-010 서술 tension(F1) · §G.2 계수 오류(D5) |
| Traceability | 0.75 | 0.75 밴드 (1 REQ 실효 미커버) | REQ-005 (호스트 OS 문단) 실효 미커버(D2). AC↔REQ 매핑 자체는 11/11 존재 |

## Baseline 재측정 (독립 실행 — 전부 이번 실행, 이 트리)

트리 귀속: HEAD `cb2f8d375` (브랜치 `WT-docs-v313-locales`), `git diff --stat bce6d7e08 -- docs-site CHANGELOG.md` 빈 출력 — 문서 대상 영역은 base `bce6d7e08` 내용과 동일. 아래 전 수치 `/usr/bin/grep` 실측.

| # | 값 | SPEC/AC 주장 | 재측정 | 판정 |
|---|---|---|---|---|
| 1 | en deferral "not yet provided" | 1건 @en:170 (AC-001) | 1건 @:170 | 일치 |
| 2 | zh deferral "尚未提供" | 1건 @zh:170 (AC-002) | 1건 @:170 | 일치 |
| 3 | en·zh `desktop-native` (cs/-ci) | 0 (AC-003/004) | 0 / 0 | 일치 |
| 4 | ko·ja `desktop-native` | 플래그 행 외 무산문(REQ-012) | 각 1건 @:58 플래그 행 | 일치 |
| 5 | ko 데스크탑-네이티브 / ja デスクトップネイティブ | 존재(완비 근거) | 6 / 6 (ja :78-80·:84·:98·:176 직독) | 일치 |
| 6 | zh 원어 토큰 | REQ-012 `桌面原生` | **0건** — zh 실제 패밀리는 `原生桌面`(:170 1행) | **불일치 → D4** |
| 7 | doctor bold-내부-괄호 (AC-007) | ko1/en0/ja1/zh1 | 1/0/1/1 (ko:41 `**권고(advisory)**`, ja:39 `**勧告 (advisory)**`, zh:39 `**建议 (advisory)**` 직독) | 일치 |
| 8 | doctor permission/sandbox (AC-006) | ja/zh **0** → 2 | **ja 2 / zh 2** (명령표 :32-33), ko 4 / en 4 | **불일치 → D1** |
| 9 | skill-guide SVG060/070 ×4 (AC-008) | 0 | 0 ×4; svg-infographic ko:158·:163 | 일치 |
| 10 | e2e 행수/표행/헤딩 | 201/195/201/195 | ✓; 표행 46/42/46/42; **헤딩 18×4 — base 에서 이미 동일** | 일치하되 AC-005 보강절 무감각 → D2 |
| 11 | t535 Codex Wiring 절 불가침 근거 | ko:69·ja:67·zh:67 마커 없음 | ko:69 `조언(advisory)` 무마커, ja:67 `勧告 (advisory)` 무마커 직독 | 일치 |
| 12 | hugo.toml 버전 | v3.1.3 | :55 `version = "v3.1.3"` | 일치 |
| 13 | CHANGELOG 배치 | `[Unreleased] → ### Docs → ### Added` | :8 `[Unreleased]` → :10 `### Added` (t535 항목 :12). **`### Docs` 부재**(전체 문서에서 :1067 구 Release 블록에만 1건) | **불일치 → D6** |
| 14 | codex-dual-harness ×4 | 각 64행(plan-phase.md:12) | 각 **63행** | 경미 불일치 → F4 |
| 15 | G2 예시 블록 | ja/zh 4행 `hook` 종단 | ja:106-111 직독 — 4행 `moai doctor hook` 종단, permission/sandbox 누락 ✓; ko 예시 :119-121 | 일치 |
| 16 | D8/Mermaid/URL/이모지 | 제약 준수 대상 | Mermaid `flowchart TD` ×3 (en :24/:133/:177), 블랙리스트 URL 0, 대상 페이지 SPEC-/`.moai/` 토큰 0(skill-guide 기존 오염은 F5), ko 대상 구간 이모지 0 | 일치 |

3개 AC 패턴 스폿-실행 완료: AC-001(1건 확인), AC-006(2/2 반환 — 결함 실증), AC-007(1/0/1/1 — 정합). AC-007 정규식은 논리 검증도 수행: 수리 형태 `**권고** (advisory)` 는 매칭되지 않고 위반 형태만 매칭된다(올바른 도구 선택).

## Defects Found

**차단(blocking)** — 수리 후 판정 재검토 대상:

- **D1** — acceptance.md:15·:52-56 — AC-006 의 검증 패턴 `moai doctor permission\|moai doctor sandbox` 가 파일 전체를 스캔해 **명령표 행(:32-33)까지 센다** — baseline 실측 2(ja/zh), 수리 후 4. AC 가 선언한 `0 → 2` 는 양쪽 cell 모두 허위. SPEC 자신의 G2 서술(spec.md:61 「명령 표(:32-33)는 4-locale 모두 정상 — 갭은 예시 블록 한정」)와 모순. — Severity: major — Class: blocking — Required fix: baseline/기대값을 `2 → 4 (각)` 로 정정하고, 예시 블록 한정 판정을 추가(예: fenced 블록 범위 추출 후 계수, 또는 `sandbox` 매치 행이 `## 例`/`## 示例` 블록 시작 행보다 뒤인지 행번호 비교).
- **D2** — spec.md:89-91 / acceptance.md:46-50 — REQ-005(호스트 OS 규칙 문단)에 실효 AC 가 없다. AC-005 의 「호스트 OS 규칙 문단 존재는 헤딩 패리티로 보강」은 무감각: 헤딩 수는 **base 에서 이미 18/18/18/18 로 동일**(문단은 ko:84 의 평문 — 헤딩도 표행도 아님)이며, 평문 추가 후에도 변하지 않는다. 표행 카운트(`^|`)도 평문을 못 잡는다. — Severity: major — Class: blocking — Required fix: 문단 존재를 이진 판정하는 전용 검증 추가(원어 토큰 grep — ko `호스트 OS 규칙` 기준, en/zh 파생 토큰은 작성 시점 확정 — 예: en `host OS rule`, zh 파생어)를 AC-005 에 병합하거나 별도 AC 로.
- **D3** — plan.md:41·:62·:69 — AC 색인 드리프트. §E 「AC-001(간격위반), AC-003(예시행), AC-005(deferral), AC-007(SVG0)」 — 4개 라벨 전부 acceptance.md 실제 번호와 불일치(실제: AC-001=en deferral, AC-003=플래그 행, AC-005=표행 패리티, AC-007=간격 스캔, 예시행=AC-006, SVG=AC-008). M1 의 「AC-008 (ko·ja 무변경)」→ AC-009, M2 의 「AC-003 (RED-now 해소), AC-004 (간격위반 0)」→ AC-006·AC-007. run 페이즈 자가검증 지도가 오염된 상태. — Severity: major — Class: blocking — Required fix: acceptance.md §D 번호로 전면 재정렬.
- **D4** — spec.md:117-119 / acceptance.md:88 — REQ-012 가 지정한 zh 원어 토큰 `桌面原生` 이 zh 페이지 자체 용어와 충돌. 실측: `桌面原生` 0건, zh 실제 패밀리는 `原生桌面`(:170 `原生桌面应用`/`原生桌面自动化`). ja 토큰(デスクトップネイティブ 6건)·ko 토큰(데스크탑-네이티브 6건)은 검증됨. 파생이 기존 zh 용어를 따르면 REQ-012 를 충족 못 하는 wrong-reason-red. — Severity: minor — Class: blocking — Required fix: 토큰을 `原生桌面 계열` 로 정정하거나(acceptance.md:88 edge 1 은 이미 「계열」 표현), 파생 시점 토큰 확정 절차로 바꾼다.
- **D5** — acceptance.md:103 — 「RED-now AC **4건**(AC-001·002·003·006·007·008)」 — 괄호 안 6개. progress.md:22 도 6건. 계수-목록 모순. — Severity: minor — Class: blocking — Required fix: 「4건」→「6건」.
- **D6** — spec.md:109-111 / acceptance.md:84 / plan.md:77 — REQ-010·AC-011·M3 이 지정한 CHANGELOG 배치 경로 `[Unreleased] → ### Docs → ### Added` 는 실제 파일에 없는 `### Docs` 섹션을 가리킨다(실측: `[Unreleased]`:8 → `### Added`:10, t535 항목 :12 직접 배치; `### Docs` 는 구 Release 블록 :1067 에만 존재). SPEC 이 인용한 t535 선례 자체가 `### Added` 직하 배치다. — Severity: minor — Class: blocking — Required fix: 경로를 `[Unreleased] → ### Added` 최상단(t535 선례)으로 정정하거나, `### Docs` 를 새로 도입하려면 그 결정을 명시(Keep a Changelog 1.1.0 에 Docs 카테고리 없음).

**선택(optional)** — 노출만 하고 오케스트레이터 재량:

- **F1** — acceptance.md:78 — AC-010 의 `hugo --quiet` 는 Then 절이 요구하는 WARN/ERROR 계수 출력을 억제한다. 같은 AC 안의 대안 경로(verify 레시피)가 실행 가능해 tension 만 존재. — Severity: minor — Class: optional — Required fix: `--quiet` 제거 후 WARN/ERROR 행 계수, 또는 exit-code 단일 판정으로 문구 정리.
- **F2** — plan-phase.md:32(입력 권위) — 「SPEC-DESKTOP-NATIVE-E2E-001 이 제거하라고 명시한 deferral 문장이 en·zh 에 생존」은 과장. REQ-DNE-004(그 SPEC spec.md:84)가 제거를 명시한 verbatim 문장("There is no opt-in automation path for `desktop-native`.")은 workflow-skill/agent 트리 대상이며 en:170 은 다른 docs-site 인용 문장이다. §B.2 의 「docs 동기화 연기」 표현이 정확한 형태. — Severity: minor — Class: optional — Required fix: 측정 파일 문구 정정 또는 본 SPEC §B.2 에 정확한 귀속 유지(현행 유지도 무해).
- **F3** — spec.md:85-87·:101-103·:121-123 — REQ-004/008/013 이 한 요구에 shall+shall-not 복합; REQ-008 말미의 「판정은 전체-파일 스캔이 0건…」은 검증 방법(테스트 고도)을 요구 계층에 포함. — Severity: minor — Class: optional — Required fix: GEARS 단일-패턴 순수성을 위해 shall/ shall-not 분리(선택).
- **F4** — plan-phase.md:12 — codex-dual-harness ×4 「각 64행」 → 실측 각 63행. 존재 주장 자체는 참. — Severity: minor — Class: optional — Required fix: 측정 파일 수치 정정(후속 터치 시).
- **F5** — docs-site/content/ko/advanced/skill-guide.md:167-177 — 기존(pre-existing) 내부 경로 오염 10행(`.moai/reports/t272/...` 참조 — 공개 저장소 트리에 존재하지 않는 경로). G4 삽입 구역과 인접. 본 SPEC 이 만든 것이 아니며 계획된 콘텐츠도 아니다. — Severity: minor — Class: optional — Required fix: 본 카드 불요 — 향후 docs 위생 카드로 분리. G4 파생 시 이 스타일을 모방하지 않도록 run 지시에 주의 문구 권고.
- **F6** — spec.md:121-123 — REQ-013 의 전제 「알려진 저장-스크립트 문법 오류가 있는 `hns-oss-docs-run` 러너」의 근거 미인용 + 이 세션에서 미검증. 워크트리·primary 체크아웃(`.claude/skills/`)·`~/.claude/skills/`·`.moai/harness/`·`.agents/` 전부에서 디렉터리 부재(find 실측, 히트 0) — 그러나 세션 스킬 목록에는 `hns-oss-docs-run` 이 존재한다(하네스가 어딘가에서 해석하는 user-owned 스킬). 금지 자체는 방향이 안전하나 전제는 VCI §1.1 표면 3상 미검증 주장. — Severity: minor — Class: optional — Required fix: 전제의 출처 관측(카드/verdict 경로)을 spec.md 에 인용하거나, run 페이즈 진입 전 리드가 러너 실물 문법 검사로 전제를 확정.

## Regression Check (Iteration 2+ 전용)

N/A — iteration 1.

## Recommendation (FAIL — manager-spec 수리 경로)

1. (D1) acceptance.md §D.6 과 AC 매트릭스 :15 행: 기대값 `2 → 4 (각)` 로 정정 + 예시 블록 한정 판정 절차 추가. plan-phase.md 는 이미 표-블록 구분을 알고 있었으므로(그 파일 :42) AC 만 고치면 된다.
2. (D2) REQ-005 검증 추가: 원어-토큰 존재 grep 을 AC-005 에 병합(예: ko `호스트 OS` = 기존 ≥1, en/zh after ≥1)하거나 별도 AC-012 로. 현재로선 REQ-005 가 커버 없는 유일한 G1 축이다.
3. (D3) plan.md §E·M1·M2 의 AC 포인터를 acceptance.md §D 실번호로 재정렬(001/002 deferral, 003 플래그, 006 예시행, 007 간격, 008 SVG, 009 무변경).
4. (D4) REQ-012 zh 토큰 `桌面原生` → `原生桌面 계열` 정정(zh:170 기존 용어 근거).
5. (D5) acceptance.md:103 「4건」→「6건」.
6. (D6) REQ-010·AC-011·plan-M3 의 CHANGELOG 경로를 `[Unreleased] → ### Added` 최상단으로 정정(t535 선례 = :12 직하 배치).
7. (권장) F1 AC-010 문구, F4 측정 파일 63행 정정을 같은 수정 패스에서 처리. F3 분리는 재량.

수리 후 재감사는 본 결함 목록 delta 범위로만 수행한다(Retry Loop Contract — Tier M ceiling 2, 다음은 iteration 2/2).

## Evidence-Bearing Report (5-섹션)

**Claim**: SPEC-DOCS-LOCALE-PARITY-REPAIR-001 plan 아티팩트 4종은 Tier M 문턱(0.80) 미달 0.71 + 차단 결함 6건으로 FAIL이다. baseline 수치 16항목 중 12항목은 입력 권위와 정확히 일치했고, 4항목(AC-006 값, REQ-012 zh 토큰, CHANGELOG 배치 경로, plan AC 색인)이 문서 내부 또는 트리와 불일치한다.

**Evidence**: 본 보고서 §Baseline 재측정 표 전체. 핵심 원시 출력 — `/usr/bin/grep -c "desktop-native"` en/zh = 0·0, `-ci` = 0·0; `/usr/bin/grep -n "not yet provided" en` → `170:...`; `/usr/bin/grep -cE '\*\*[^*]*\([^)]*\)[^*]*\*\*'` ko/en/ja/zh = 1/0/1/1; `/usr/bin/grep -c "moai doctor permission\|moai doctor sandbox"` ja/zh/ko/en = 2/2/4/4 (ja 매치행 :32-33 직독); `wc -l` e2e 201/195/201/195; `/usr/bin/grep -c '^|'` = 46/42/46/42, `'^#'` = 18×4; `/usr/bin/grep -n 'version' docs-site/hugo.toml` → `55:  version = "v3.1.3"`; `/usr/bin/grep -n '### Docs' CHANGELOG.md` → :1067 단 1건; `git log --oneline -1 175d63f3f` → `docs(t274): v3.1.3 documentation catch-up (#1662)`; `git log --oneline -1 732609dcf` → `docs(SPEC-CODEX-EVENT-COVERAGE-001) ... (t496)`.

**Baseline-attribution**: (this run, this tree) — worktree `.claude/worktrees/t538` @ `cb2f8d375`, 브랜치 `WT-docs-v313-locales`; 문서 대상 영역은 base `bce6d7e08` 와 동일함을 `git diff --stat bce6d7e08 -- docs-site CHANGELOG.md` 빈 출력으로 소속 확인 후 측정. 전 재측정 명령은 `/usr/bin/grep`으로 실행(ugrep 래퍼 회피 — REQ-012 취지와 동일한 도구 규율을 감사자 스스로 적용).

**Gaps**: (1) `hns-oss-docs-run` 러너 본체를 도달 가능한 모든 뿌리에서 찾지 못해 문법-오류 전제를 검증하지 못했다(세션 스킬 목록에는 존재 — F6, find 히트 0 실측). (2) hugo 빌드를 실행하지 않았다 — AC-010 판정은 run 페이즈 산출물에 대한 것이고 plan 산출물 감사 범위 밖. (3) plan-phase.md §Gaps 가 명시한 3종(Fixed 전수 판정, 문단 단위 diff, SVG0 타-표면 전수)은 카드 불요로 계승 — 본 감사도 수행하지 않았다. (4) en/zh 페이지의 축 외 문단 전수 대조는 하지 않았다(G1 축 단위 감사). (5) `related_specs` 여분 키의 스키마-문서상 지위(선택 필드 목록에 없음)는 lint 구현 기준 무시 확인까지만 했다(문서-레벨 결정은 스키마 SSOT 소관).

**Residual-risk**: (1) D2 미수리 시 REQ-005 축은 run 에서 구현돼도 검증 없이 녹통과될 수 있다 — 무검증 통과가 가장 비싼 잔여 위험. (2) D4 미수리 시 zh 파생 번역이 기존 용어(原生桌面)를 따르는 올바른 결과가 REQ-012 에서 RED 로 판정될 수 있다. (3) AC-011 의 「파일 10개 한정」은 착지 후에만 기계 판정 가능 — plan 승인 시점엔 선언만 존재한다. (4) F5 의 기존 내부-경로 오염(ko skill-guide :167-177)은 본 카드 밖이지만 G4 파생자가 같은 스타일을 모방할 유인이 있다.
