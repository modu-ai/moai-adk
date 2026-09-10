# t453 sync-audit 판정 — AlwaysLoadedTokenBudget 76,400 → 77,200 인상

- 감사자: sync-auditor (독립 사후 감사, Class B · SPEC 없음)
- 대상: 커밋 `9044aa214` (단일 커밋, `internal/config/token_budget_guard.go` + `.moai/reports/t453/verdict.md`)
- 트리: `WT-token-budget` @ `9044aa214`, base `400f37eb9` (로컬 develop 팁) — 감사일 2026-09-03, 본 레인 워크트리 로컬 측정

## Verdict: PASS-WITH-DEBT (4차원 92.5/100)

| Dimension | Score | Verdict | Evidence (요약; 전문은 아래 5절) |
|---|---:|---|---|
| Functionality (40%) | 88 | PASS | 적색→녹색 재현·확인(76,939/77,200/261, 17항목); 파일 귀속 산술 전수 정확(+810, 76,129 조정 일치). 감점: 카드별 분배 오기(F1), 충실성 주장 과대(F2) |
| Security (25%) | 100 | PASS | diff = 상수 1 + 주석 + 마크다운. 비밀 없음, 동작 변경 없음, 테스트 미약화(diff 0행) |
| Craft (20%) | 90 | PASS | 2파일 최소 diff, 테스트 파일 무변경(빈 diff 실측), 증거문서 5절 완비·Gaps 정직. 감점: F1 내부 합 불일치(225+192+95=512≠510), F3 |
| Consistency (15%) | 95 | PASS | 주석 언어·체재 파일 기존 관례 일치, @MX 태그 영어, Conventional commit + card id + 증거 경로 규약 준수 |

가중 92.5 · 조화평균 93.0. must-pass(Functionality·Security) 모두 통과. PASS-WITH-DEBT인 이유: (a) 인상 체인 자체가 명명된 부채(@MX:DEBT + 측정 기반 @MX:UPGRADE 표적 — 수용 가능, 퇴출구 못박힘), (b) F1을 병합 전 수리 권고(주석·마크다운만, ≤4줄)로 지정. F1은 문서 수치 오기로 동작 면이 0이라 FAIL까지는 과잉이나, 이 카드의 산출물이 곧 귀속 기록인 만큼 방치하지 않는다.

## Findings (심각도순)

- **F1 [Medium] [병합 전 수리 — 주석/문서만]** `internal/config/token_budget_guard.go:61-62` · `.moai/reports/t453/verdict.md:57` — 카드별 성장 분배가 실측과 다르다. 주석·판정서: t224 +510(kanban-dispatch +225) / t386 +143. blob 실측: t386(`8925d89c7`) kanban-dispatch 32,800→33,203 B = **+100 tok**(+403 B); t224 라인(02cf8ec39^→826a63ebf) kanban-dispatch +1,065 B = **+267 tok** → t224 합계 **+554**(267+192+95). 주석의 괄호 안 분해(225+192+95=512)조차 스스로의 총계 +510과 안 맞는다. 검산: 554+100+20+136=810 — **파일 귀속·총계(+810)·76,129 조정은 정확**하며 오류는 카드 간 분배(≈44 tok)에 국한된다. 수리: 두 파일의 카드별 수치를 실측치(t224 +554 / t386 +100, kanban-dispatch t224 +267·t386 +100)로 고친다.
- **F2 [Medium-Low] [optional]** `.moai/reports/t453/verdict.md:67-68` — "17개 열거 항목 전수가 매 턴 주입 집합과 정확히 일치"는 양방향으로 읽면 과대다. `CLAUDE.local.md`(git 추적, 49,597 B ≈ 12,399 tok)는 매 턴 주입됨을 이번 감사 세션 컨텍스트에서 직접 관측했으나 열거 밖이다(사용자 레벨 auto-memory도 주입되나 트리 밖 — 부외는 타당). 선결 범위 설계이며 이번 산술·인상 판단을 바꾸지 않는다(계수에 넣으면 적색이 오히려 커진다). 수리: "배포 always-loaded 표면(룰 트리 + 3 고정 슬롯)과 일치"로 한정 서술하거나, 다이어트 카드에 범위 질문을 추가한다.
- **F3 [Low] [optional]** `internal/config/token_budget_guard.go:73` — @MX:DEBT 체인 표기 "(76,000 -> 76,400 -> 77,200)"이 76,210 단계를 생략해 판정서 Residual-risk의 "4회째(74,317 → 76,000 → 76,210 → 76,400 → 77,200)" 열거와 어긋난다. 본문 주석엔 전 단계가 있어 태그 행의 생략이지만, 다음 열람자가 두 표기를 대조하면 어긋난다.
- **F4 [Low] [optional]** — 여유 261 tok(최근 조항 20~367 대비 단일 조항분)은 판단의 여지가 있는 선택이다(측정+1의 더 작은 대안 존재). 단 체인 관례(t421의 135)와 일치하고 최근 성장 속도(+810/약 2일)에서 다음 조항 착지에 재트립된다 — 예산 풀어주기가 아니다. 조치 불요, 기록만.

## Claim (주장)

1. 적색은 실재했고 본 커밋이 녹색으로 만든다 — 이 감사가 독립 재현·확인했다.
2. 성장의 파일 귀속(+810 tok, 5개 표면 파일, 76,939−810=76,129의 t421 기록과 조정)은 **정확하다**. 카드별 분배는 오기가 있다(F1).
3. 세 갈래 대안 검토(측정 대상 변경·문서 축소·상한 인상) 중 인상이 가장 작은 정직한 수리였다는 판정은 **뒷받침된다**(아래 근거). 단 "주입 집합 정확히 일치" 주장은 F2의 한정이 필요하다.
4. 가드 무결성: 테스트 미약화 없음(diff 0행), 비밀 없음, 동작 변경은 상수뿐.

## Evidence (증거 — 실행한 명령 + 출력)

**적색 재현(도출 재현)**: `git show 9044aa214^:internal/config/token_budget_guard.go | grep 76400` → `61:const AlwaysLoadedTokenBudget = 76400`. `git show 9044aa214 --stat` → 변경 2파일(가드+판정서)로 표면 파일 불변 → 표면 측정치는 두 커밋에서 동일. 테스트 판정 조건 `token_budget_guard_test.go:72` `if total > AlwaysLoadedTokenBudget` → 76,939 > 76,400 (overflow 539) = 적색. (구 상수를 물리적으로 재실행하지는 않았다 — Gaps 참조.)

**녹색 확인 [본 감사, HEAD `9044aa214`]**: `go test -count=1 -run 'TestAlwaysLoadedTokenBudget$' ./internal/config/ -v` →
```
token_budget_guard_test.go:69: always-loaded surface = 76939 tokens (budget 77200, headroom 261, 17 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.01s)
```
`go test -count=1 ./internal/config/` → `ok github.com/modu-ai/moai-adk/internal/config 3.223s` (패키지 전체, CodexContractByteCeiling 포함 — AGENTS.md 14,774 B ≤ 24,576).

**파일 귀속 산술 [git ls-tree -r -l, b9efb3626 ↔ 400f37eb9]**: 변경 .md 11개 중 6개(coding-standards, moai-memory, spec-workflow, worktree-integration, kanban-dispatch-detail, moai-constitution-detail)는 frontmatter `paths:` 보유로 표면 밖(전수 확인 — awk 재판정, 가드 `hasPathsRestriction`와 동일 논리). 표면 5개: kanban-dispatch 32,800→34,268 (+1,468 B, +367 tok), agent-common-protocol 25,969→26,739 (+770, +192), AGENTS.md 14,229→14,774 (+545, +136), moai-constitution 16,126→16,506 (+380, +95), moai-mcp-tools 6,320→6,403 (+83, +20). 합 +3,246 B = **+810 tok**(파일별 floor 산술 실측 일치). 76,939 − 810 = **76,129** = t421 주석 기록(`token_budget_guard.go:48` "실측 76,129 토큰(여유 81)")과 정확히 일치.

**열거 충실성**: 표면 열거 17 = no-paths 룰 14 + 고정 슬롯 3. 이 감사 세션 자체가 주입 집합의 직접 관측이다 — 항상 주입된 룰이 열거된 14개와 정확히 일치(kaanban-dispatch 등 3개 무-frontmatter/무-paths 직접 확인; go.md 등 조건부 룰은 paths: 보유 확인). t368형 역방향(주입 안 되는 외래물 계수) 없음 확인.

**t224 5표면 설계 [git show 02cf8ec39 --stat]**: kanban-dispatch.md / kanban-dispatch-detail.md / agent-common-protocol.md / moai-constitution.md / manager-lead.md (+템플릿 미러·codex emit) — 커밋 메시지 스스로 "five doctrine surfaces" 명명. agent-common 재진술(+192)은 5개 의도 배치 중 하나로, de-dup는 착지 설계 재소송. 확인.

**테스트 미약화**: `git diff 9044aa214^ 9044aa214 -- internal/config/token_budget_guard_test.go` → 빈 출력(0행).

**카드별 분배 실측 (F1 근거)**: `git ls-tree -r -l 8925d89c7{,^}` kanban-dispatch → 32,800→33,203 (+403 B = +100 tok, t386). `02cf8ec39^`(t224 분기 기저 32,800) → `826a63ebf`(t224 창 수리 후) 34,268 — t224 라인 +1,065 B = +267 tok. AGENTS.md `936adb4b0{,^}` 14,229→14,774 (+136, t196 ✓). moai-mcp-tools `21734f9e9{,^}` 6,320→6,403 (+20, t236 ✓).

**창 대기 브랜치 부신호**: `git diff --name-only 400f37eb9...<branch> -- <표면 경로>` → WT-pipeline-fallthrough(t447)·WT-gears-canon-rubric(t367)·WT-doctor-freshness-reds(t444) 전부 0행. 8개 중 3개 표본 확인.

**최소 변경 판정**: 축소 대안은 실행 불가 — 유일 de-dup 후보(+192)는 초과분 539를 못 덮고, t224+t196 롤백(646)은 두 카드의 의도 설계(위 5표면 확인)를 되돌린다. 대형 파일 다이어트는 stub+lazy-loading 설계가 필요한 별도 카드로 명시적으로 연기돼 왔다(t421 주석 2회). 인상(실측+261)이 주어진 증거에서 가장 작은 정직한 수리다.

## Baseline-attribution (baseline 귀속)

- 전 측정 본 감사자가 본 레인 워크트리(`9044aa214`)에서 직접 실행. 적색 도출의 두 좌표: 구 상수 `9044aa214^`(=`400f37eb9`) git 객체에서 직접 판독 + 표면 76,939는 HEAD에서 `go test -v` 실측(커밋이 표면 파일을 건드리지 않음을 --stat으로 확인).
- 델타 산술의 두 좌표: `b9efb3626` ↔ `400f37eb9` blob 크기(`git ls-tree -r -l`). 카드별 분배는 `936adb4b0`/`31566c117`·`21734f9e9`/`8925d89c7`/`02cf8ec39`·`826a63ebf` 경계 blob.
- 판정 문서(t453 verdict.md)의 수치는 신뢰하지 않고 전수 재실측했다 — 파일 귀속·총계는 일치, 카드별 분배는 불일치(F1).

## Gaps (미검증)

- 구 상수 76,400에서의 적색을 물리적으로 실행하지 않았다(stash 금지·트리 무결성). 도출 재현(상수 확인 + 표면 불변 + 판정 조건 소스)으로 대체 — 논리적으로 닫히지만 관측은 아니다.
- CI 판정 미확인 — develop 병합·push는 리드 소관, 병합 후 push CI가 최종 판정.
- 창 대기 8브랜치 중 5개(t300·t302·t348·t353·t339)의 카드↔브랜치 대조는 이번 감사도 수행하지 않았다(3개 표본만). 구속 검사는 창 안 rev-list 재측정 소관.
- `internal/template/templates/**` 미러 트리의 대응 규칙 파일은 이번 감사도 재지 않았다(가드가 재포 트리만 재는 것이 설계).
- CLAUDE.local.md가 아닌 사용자 레벨 auto-memory·시스템 리마인더의 토큰 기여는 트리 밖이라 측정 자체가 불가능하다(부외 타당).

## Residual-risk (잔여 위험)

- 인상 4회째가 DEBT 규율을 마모시키는 경로 위험이 남는다 — 완화(체인 전체 기록·"다음 트립에 자동 정당성은 없다" 상수 주석 명문화·@MX:UPGRADE 측정 표적 4개)를 확인했으나, 다음 트립에서 그 문구가 실제로 다이어트 카드 착수로 이어지는지는 미래 사건이다.
- F1을 병합 전에 고치지 않으면 다음 인상 감사가 잘못된 카드별 기준(+510/+143)에 맞춰 조정하게 된다.
- 여유 261 tok는 단일 조항분 — 다음 always-loaded 조항 착지 시 재트립되며, 그것은 가드의 설계된 기능이다(결함이 아님).
- char/4 추정 오차(±약 15%)는 근본 한계로 그대로 남는다(가드는 상대 트립와이어 — 근사 무의존은 의도적).
