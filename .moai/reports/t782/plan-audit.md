# SPEC-GIT-PROC-SAFE-001 Plan-Audit Report (iteration 1/1)

- Auditor: plan-auditor (independent, single-Claude verdict — cross-model second opinion not invoked: 모든 판정 근거가 본 트리 직접 측정이고 `audit_model` 설정 키가 `.moai/config/sections/`에서 검색되지 않음)
- Baseline attribution: worktree `WT-gitproc-audit`, HEAD `c9ceff175`, 2026-09-14, 본 세션 직접 판독. 감사 창 동안 HEAD 불변 + 외래 커밋 없음(`git status --porcelain` = `?? .moai/specs/SPEC-GIT-PROC-SAFE-001/` 단일 항목) — 단일 서술자 전제 유지.
- Reasoning context from the SPEC author / orchestrator reproduction claims: 입력으로 취급해 검증 대상으로만 사용. M1 Context Isolation 준수.

## Verdict: FAIL

Aggregate Score: 8.1 / 10 (Tier M 자칭이나 frontmatter `tier:` 부재로 감사 계약상 Tier L 기준 0.85 적용 — 어느 기준이든 must-pass 방화벽 위반이 판정을 결정)

**FAIL 사유 (M5 must-pass 방화벽)**: MP-3 YAML frontmatter 타입 위반 (발견 F1). 그 외 F2(blocking)가 수리 설계의 교리 정합성 결함. 두 건 모두 소규모 표적 수리로 제거 가능 — 결함 구조 자체(문서 절차 교정 접근법)는 건전하다.

## Must-Pass Results

| # | 판정 | 근거 |
|---|------|------|
| MP-1 REQ 번호 일관성 | **PASS** | `grep '^### REQ-'` → REQ-GP-001..004, REQ-SX-001, REQ-AC11-001, REQ-TF-001, REQ-TN-001. 8건, 도메인 접두 체계 내 연속, 중복 0건 (spec.md:37-77) |
| MP-2 GEARS 형식 | **PASS** (통보된 이탈 2건) | 6/8 엔트리가 5패턴 정합: REQ-GP-001은 State-driven(`**While** ... **인 동안** ... 규정해야 한다`, spec.md:39), REQ-GP-002/003/004/SX-001/TN-001은 Ubiquitous(+금지형 `~않아야 한다`). 이탈: REQ-AC11-001 2번째 문장이 서술형 산문(spec.md:69), REQ-TF-001이 shall-표지 없는 선언형 `적용한다`(spec.md:73). **판정 논리**: 본 감사는 어휘가 아니라 구조적 내용 검사(각 엔트리가 검증 가능한 규범 의무를 지시문 형태로 보유)를 적용 — 두 엔트리 모두 이진 검증 가능한 의무를 가지므로 must-pass 통과, 러브릭 점수는 0.75 밴드로 반영(F5로 수리 권고) |
| MP-3 frontmatter 유효성 | **FAIL** | F1 참조. `internal/spec/lint.go:516` `Tags string` 선언 + `:726` `yaml.Unmarshal` 오류 경로(`:728` "YAML parsing error" 반환) 대비 spec.md:13이 YAML 흐름 시퀀스 사용 |
| MP-4 언어 중립성 | **N/A** | git 절차 교리 단일 도메인 SPEC — 다중 프로그래밍 언어 도구 나열 없음. 템플릿 중립성은 REQ-TN-001로 별도 규정됨 |
| MP-5 D7 cross-SPEC | **PASS** | `related_specs` 2건 모두 존재: SPEC-GITFLOW-DOCTRINE-ALIGN-001(status: in-progress), SPEC-GIT-DELIVERY-PROCEDURE-001(status: completed). retired/superseded/archived 없음 → 재화해 요구 없음 |
| MP-6 D8 cross-platform | **PASS** | `grep -c syscall` → 5개 아티팩트 전부 0 → auto-PASS |
| MP-7 clarification gate | **PASS** (스캔 유해성 1건 기록) | `[NEEDS CLARIFICATION: <topic>]` 마커(콜론+토픽 형식) 0건. 정규식 레벨 적중 2건은 모두 부정문: plan.md:82("[NEEDS CLARIFICATION] 없음"), progress.md:21("[NEEDS CLARIFICATION] 0건") — 미해결 토픽 부재. F6에 스캔 오탐 방지 표현 수리 권고 |

## Dimension Scores (D1-D6)

### D1 Evidence Integrity — 9.5/10

본 트리에서 SPEC이 인용한 모든 file:line을 재판독 — **불일치 0건**:

- AC-01 로컬 `.claude/agents/moai/manager-git.md`: 98행 `git checkout main && git pull origin main`(Phase A), 112행 `git switch -c feat/SPEC-XXX`(Phase C), 121-124행(`checkout main`/`fetch`/`reset --hard origin/main`/`pull origin main`, Phase D), 129행 실패 복구, 131행 "canonical step ordering" 교차참조, 139행 `main_late_branch` + `reset --hard origin/main` — 전부 그대로 존재. § Late-Branch Invocation Pattern 헤딩 90행, 섹션 종료 131행 — SPEC 범위 표기(~90-131) 정확.
- 템플릿 동일 파일: 96/110/119-122/127/137행 — 전부 일치.
- `spec-workflow.md`: 50행 Route B 사전조건(`git rev-parse --abbrev-ref HEAD == main` ... `git switch -c plan/SPEC-XXX` 원문 일치), 53-62행 Step 4 closure(56-59행 명령 블록, 62행 post-condition + 교차참조) — 전부 일치. 템플릿 사본은 `diff` 결과 **0행**(바이트 동일 미러) → 로컬 행 번호가 템플릿에 그대로 이전.
- AC-11 "이미 수리됨" 판정 **사실로 확인**: 로컬 158행 / 템플릿 156행에 acceptance.md가 인용한 처방 문구("run `git fetch` first and wait until it completes, then run `git rev-list --count --left-right`")가 원문 존재. 검증 전용 프레이밍 타당.
- SX-R04: 로컬 delivery.md 355행(Step 3.4 헤딩), 367행 "Mode conditions (same as...)" + 368-369행 2불릿 재인용, 385-391행 Auto-Merge Execution 5단계(390행 "4. Checkout target branch, fetch latest" 포함), 393-397행 Failures, 399행 Post-Merge Automatic Cleanup — 전부 일치. 템플릿 330/342/360/365/368/374행 일치. `grep -ci auto-merge` = 로컬 18 / 템플릿 18 — SPEC 주장(사본당 18적중) 정확.
- manager-git.md 사본 분기(~5행): `diff` 출력 6행, 오직 frontmatter description 블록(5-7행) — 관측대로.
- research.md의 측정치 전부 재현 성공. 동 배치 grep 오아티팩트를 스스로 폐기·재측정한 Gaps 기록(research.md:63-65)은 근거 무결성의 모범 사례.

감점 0.5: research.md:60의 "~94 diff 행" 수치는 재측정하지 않았음(길이 델타 32행 + 마커 오프셋 일관성으로 분기 존재는 간접 확인).

### D2 Requirements Quality — 8.0/10

- 8 REQ / 7 AC — Tier M 상한(16/16) 내. 6/8 GEARS 정합(위 MP-2). 요구는 행위/결과 중심이고 구현 세부를 끌어들이지 않음(HEAD 재작성이 아닌 문서 텍스트 대상).
- AC-11 검증 전용 설계 우수: "편집 0건 + 양 사본 폐쇄 텍스트 존재 관측 + AC-11 관련 편집 커밋 부재"의 3중 관측으로 shaped 되어 run-phase가 편집을 발명할 여지를 차단(acceptance.md:62-66). 문서 도메인 RED/GREEN 2셀 전환(RED=위반 텍스트+file:line, GREEN=대체 텍스트 검사)도 문서 SPEC에 적법한 형태.
- 감점: REQ-TN-001 무 AC 커버리지(F4), REQ-AC11-001/REQ-TF-001 등기 이탈(F5).

### D3 Feasibility & Constraint Compliance — 6.0/10

**핵심 설계는 기계적으로 타당하다**: "하나의 브랜치는 하나의 워크트리에만 체크아웃" → "main 커밋 적립 모델은 워크트리 격리와 양립 불가" → 커밋은 워크트리 자기 브랜치에 적립(spec.md:33). Phase D 소멸 논리(main이 커밋을 받지 않음)도 성립. REQ-GP-002의 허용 형태 한정(`git branch -m`, `git fetch`, 자기 브랜치 커밋, push)은 AGENTS.md §2 금지 목록과 정확히 보완 관계.

**그러나 대체 모델의 승격 단계가 이 저장소의 관측된 구성과 충돌한다 — F2 (BLOCKING)**:

- 측정: `.moai/config/sections/git-strategy.yaml` 9행 `workflow: git-flow` + 29행 `branch_creation.auto_enabled: false`(git_strategy.manual 블록). Late-Branch 절차의 트리거 술어는 정확히 `auto_enabled == false`(manager-git.md 로컬 92행/템플릿 90행) → **본 저장소에서 그 절차가 발동되는 유일 모드가 git-flow 모드**.
- git-flow 모드의 통합 경로는 PR이 아니다: `gitflow-lane-protocol.md` §2(통합 워크트리 `.claude/worktrees/develop`에서 `git merge --no-ff`, delivery.md Step 3.2가 절차 정본, 그 push 단계는 이 리포에서 EXCLUDED) + §4(레인 push 금지, 리드 일괄) + `repo-local-pr-policy.md`(카드 PR 없음) + `delivery-policy.md` §2("The only no-PR path is an explicitly configured git-flow WT-* integration branch").
- 반면 수리 후 텍스트는 PR 승격을 단일 모델로 명시하도록 요구한다: plan.md M1 3번째 불릿("PR 시점에 해당 브랜치 push + `gh pr create` → 해석된 merge_method로 병합"), AC-GP-01a GREEN Then(3)("커밋 적립-브랜치 → **PR 승격 모델이 명시되고**"). 이 텍스트가 로컬 사본에 들어가면 `auto_enabled == false` 술어 아래에서 develop-병합 교리와 정면 모순 — SPEC이 제거하려는 결함과 동일 클래스의 교차 문서 모순을 새로 만든다.
- research.md:70이 "`git-strategy.yaml`의 현재 운영 값... 판독하지 않았다"고 명시 — 이 Gap이 위 충돌을 가렸다.

**필요한 수리 (소규모)**: 대체 텍스트의 승격 단계를 모드 조건부로 한정 — github-flow/Route B → 브랜치 push + `gh pr create` + merge_method 병합; git-flow → delivery.md Step 3.2 통합 워크트리 병합(WT-* → develop). spec.md §1 대체 모델 문장, plan M1, AC-GP-01a GREEN 셀 3곳에 동일 규정 반영 + research.md Gap 갱신. 템플릿 사본(하류 github-flow 사용자)은 PR 모델이 그대로 유효하므로 템플릿 텍스트는 불변 가능하나, 로컬 사본 분기 여부를 run-phase에서 명시적 결정으로 기록할 것.

추가 정합성 확인(통과): 브랜치 이중 체크아웃 없음(워크트리 자기 브랜치만 사용), Phase D 제거 후 post-condition은 SPEC 위험 절(spec.md:112)이 새 형태를 명시, spec-workflow Route A/B 표와 manager-git 재작성의 정합은 M2가 교차참조 갱신으로 처리.

### D4 Scope Discipline — 9.0/10

- 6개 파일(2사본 × manager-git.md / spec-workflow.md / delivery.md) 일관 유지 — plan M1-M3, acceptance AC 매트릭스, REQ 스코프 전부 동일 집합. Out of Scope 4개 H3 소제목 + 불릿(spec.md:91-107)으로 사본 분기·런타임 코드·AC-11 재수리·타 교리 파일을 명시 배제.
- 선존재 분기(manager-git frontmatter ~5행, delivery 분기)는 관측으로만 기록되고 작업 항목이 아님 — 적절.
- 감점: git-flow 교리와의 상호작용(F2)이 Out of Scope 절의 "이미 워크트리 모델을 서술하는 파일" 목록에도, 수리 범위 논의에도 등장하지 않음 — 범위 경계 판단의 누락.

### D5 Template Neutrality & Mirror Discipline — 8.0/10

- REQ-TN-001(내부 SPEC ID·dev-only 로컬 경로(`.claude/rules/local/*`) 미인용·16언어 중립) 존재. plan §G anti-pattern으로 재확인.
- 전체 파일 `cp` 미러링 금지 + 범위 밖 분기 비접촉(REQ-TF-001, AC-MIRROR-01, plan §G 1번) — 선존재 분기 보호 설계 적절. 로컬 사본의 이미 git-flow를 반영한 description(diff로 확인)이 통째 cp로 소멸할 위험도 차단됨.
- 감점: REQ-TN-001에 대응하는 검증 AC 부재(F4와 동일 근거) — run-phase §E 자체검증 목록(plan §E 4항목)에도 중립성 grep이 없다.

### D6 Milestone/AC Traceability — 8.0/10

| AC | 마일스톤 | REQ | 판정 |
|----|----------|-----|------|
| AC-GP-01a | M1 | REQ-GP-001/002 | 커버 |
| AC-GP-01b | M2 | REQ-GP-003 | 커버 |
| AC-GP-01c | M1 | REQ-GP-001(139행) | 커버 |
| AC-SX-01 | M3 | REQ-SX-001 | 커버 |
| AC-AC11-01 | M4 | REQ-AC11-001 | 커버 |
| AC-XREF-01 | M1/M2 | REQ-GP-004 | 커버 |
| AC-MIRROR-01 | M4 | REQ-TF-001 | 커버 |
| — | — | REQ-TN-001 | **미커버 (F4)** |

- 검증 동사 전부 실행 가능(grep/`git diff`/직접 판독 — 문서 SPEC에 적합). M1→M2 의존(교차참조 정합) 명시. 마일스톤 우선순위(절차 모델 교체 선행) 근거 있음.
- 감점: REQ-TN-001 무 AC. AC-GP-01b RED 셀이 로컬 행만 핀 — 본 감사에서 템플릿 바이트 동일성 확인했으므로 전이 유효하나, RED 셀 자체로는 양 사본 자기완결이 아님(F7).

## Defects Found

- **F1** — spec.md:13 — frontmatter `tags: [git, doctrine, manager-git, sync, worktree]`가 YAML 흐름 시퀀스. `internal/spec/lint.go:516`(`Tags string`)·`:726`(Unmarshal 오류 시 "YAML parsing error" 반환)과 타입 불일치 — 린트 파서가 frontmatter 전체를 디코딩 실패시키는 결함. 카탈로그 관례(예: `tags: "sync-phase, parallelization, ..."`)와도 불일치 — Severity: critical — Class: **blocking (MP-3)** — Required fix: `tags: "git, doctrine, manager-git, sync, worktree"` (따옴표 감싼 콤마 문자열 1행 교체).
- **F2** — spec.md:33·39 + plan.md M1 + acceptance.md AC-GP-01a GREEN — PR 승격 단일 모델이 `auto_enabled == false` 술어를 공유하는 git-flow 모드의 develop 통합 교리(gitflow-lane-protocol.md §2/§4, delivery-policy.md §2, 관측: git-strategy.yaml:9/29)와 모순; research.md:70이 해당 설정 미판독을 Gap으로 남김 — Severity: critical — Class: **blocking** — Required fix: 대체 텍스트 승격 단계를 모드 조건부로 한정(github-flow/Route B → PR 승격; git-flow → delivery.md Step 3.2 통합 워크트리 병합), spec.md §1/plan M1/AC-GP-01a GREEN 3곳 반영, research.md Gap 갱신.
- **F3** — spec.md frontmatter — `tier:` 필드 부재. plan.md §A·progress.md §E.1은 Tier M을 기록했으나 스키마상 `tier:`는 spec.md frontmatter에 실려야 하고, 부재 시 Tier L로 취급(감사 임계·이터레이션 상한 변경) — Severity: major — Class: optional — Required fix: frontmatter에 `tier: M` 1행 추가.
- **F4** — spec.md:75-77 / acceptance.md — REQ-TN-001에 대응 AC 없음(7/8 커버). 중립성은 plan §G anti-pattern에만 존재해 run-phase 판정 가능 근거가 없다 — Severity: major — Class: blocking (verdict에는 불영향, F1·F2 수리와 동일 패스에서 처리 권고) — Required fix: acceptance.md에 AC-TN-01 추가(예: 수리 후 템플릿 2사본에서 `SPEC-[A-Z]`·`.claude/rules/local` grep 적중 0건).
- **F5** — spec.md:69, :73 — REQ-AC11-001 2번째 문장이 서술형 산문, REQ-TF-001이 shall-표지 없는 선언형 — GEARS 등기 이탈 2건 — Severity: minor — Class: optional — Required fix: REQ-AC11-001을 금지형+의무형으로 재구성("어떤 파일도 편집해서는 안 되며, ... 문서화해야 한다"), REQ-TF-001을 "~적용해야 한다"로 통일.
- **F6** — plan.md:82, progress.md:21 — 부정문 안의 리터럴 토큰 `[NEEDS CLARIFICATION]`이 마커 스캔 오탐 유발 — Severity: minor — Class: optional — Required fix: "확인 필요 항목 없음" 등으로 표현 교체(또는 감사 문서에 부정문 예외를 명시 — 본 감사는 부정문으로 판정했음을 여기 기록).
- **F7** — acceptance.md:33-38 — AC-GP-01b RED 셀이 로컬 spec-workflow.md 행만 핀. 템플릿 사본은 바이트 동일 미러(본 감사 `diff` 0행 확인)라 전이되나, 셀 자체가 양 사본을 자기완결적으로 고정하지 않음 — Severity: minor — Class: optional — Required fix: RED 셀에 "템플릿 사본은 바이트 동일 미러(행 번호 동일)" 1문장 추가.

## Regression Check

Iteration 1 — 이전 이터레이션 결함 없음.

## Recommendation (수리 경로 — FAIL 판정에 따름)

1. (F1) spec.md:13 tags를 콤마 문자열로 교체 — 1행.
2. (F2) 대체 절차 텍스트의 승격 단계 모드 조건부화 — spec.md §1/§2(REQ-GP-001 두 번째 불릿 인접), plan.md M1 3번 불릿, AC-GP-01a GREEN Then(3). git-strategy.yaml 판독 결과를 research.md에 귀속 기록.
3. (F3) `tier: M` 추가.
4. (F4) AC-TN-01 신설 + plan §E 자체검증에 중립성 grep 1항 추가.
5. (F5/F6/F7) 등기·표현·핀 보강.
6. 위 수리 후 **재감사는 본 결함 델타(F1-F7)에 한정**된 스코프로 실행 가능 — 전면 재감사 불요.

## Gaps (미검증 항목 — 명시적)

- `moai spec lint` 바이너리를 실행하지 않음(설치 바이너리 버전 미검증) — MP-3 판정은 `internal/spec/lint.go` 소스(516/726/1146행) 직접 판독으로 대체. yaml.v3의 seq→string TypeError 동작은 소스 근거이며 런타임 재현은 아님.
- delivery.md 로컬↔템플릿 전체 diff의 "~94 diff 행" 수치는 재측정하지 않음(길이 델타 32행, 마커 오프셋 일관성으로 분기 존재만 간접 확인).
- 6개 대상 파일 외 late-branch/reset --hard 모델의 잔존 표면 전수 스윕 미실행 — SPEC 자체가 run-phase M4 2차 스윕으로 위임(plan.md:62, spec.md:111). 본 감사도 이를 Gap으로 존중.
- 템플릿 delivery.md의 auto-merge 18적중의 행별 분포 재계수는 하지 않음(`grep -ci` 총계만 검증).
- 크로스모델 2차 감사(audit_multi) 미호출 — 근거 전부 직접 텍스트/설정 측정이고 설정에 `audit_model` 키 부재 확인.

## Residual-risk

F2 수리 후에도 manager-git.md 로컬 사본의 Late-Branch 섹션이 본 저장소의 실제 카드 흐름(레인이 merge를 직접 수행, manager-git은 release 경로)과 어느 정도 추상적으로만 정렬될 수 있다 — 이는 후속 카드 후보(로컬/템플릿 description 분기 정리, research.md:59)와 함께 다룰 문제다. 또한 본 감사의 행 번호는 `c9ceff175` 기준 — 흡수(merge) 이후 재판독 시 재측정 필요.

---

# Iteration 2 — Scoped Delta Re-audit (F1-F7)

- Scope: iteration-1 결함 델타(F1-F7) + 신규 lint 경고 1건 판정. 전면 재감사 아님(Retry Loop Contract; Tier M 상한 2회 중 2회).
- Baseline attribution: 동일 워크트리 `WT-gitproc-audit` @ `c9ceff175`, 2026-09-14. 감사 창 단일 서술자 유지 — `git status --porcelain` = `?? .moai/reports/t782/` + `?? .moai/specs/SPEC-GIT-PROC-SAFE-001/` 두 비추적 디렉터리뿐, 추적 파일 무변경.
- **회귀 확인**: 수리 대상 6개 파일(manager-git.md ×2, spec-workflow.md ×2, delivery.md ×2)에 대해 `git diff --stat` → **빈 출력(무변경)** — 수리가 SPEC 디렉터리에만 국한됨을 관측. spec.md/plan.md/acceptance.md/research.md/progress.md의 비수리 영역(전체 RED 인용 블록, Out of Scope, §B-§H) 대조 판독 — 이터레이션-1 판독치와 동일. 수리 발신 보고의 "12 edits" 집계와 관측 편집 지점 수(프런트매터 2필드 + §1 + REQ 2건 재작성 등)에 사소한 계산 차이가 있으나, 실질 내용은 F1-F7에 정확히 귀속 — 결함 아님.

## Verdict: **PASS**

Aggregate Score: **9.4 / 10** (iteration-1: 8.1 → 상승 — STOP 신호 해당 없음). Tier M 임계 0.80 충족(`tier: M` 이제 frontmatter에 기록됨).

## Per-Finding Disposition

| 발견 건 | 처분 | 본 감사 직접 증거 |
|---------|------|------------------|
| **F1** tags 타입 (MP-3) | **RESOLVED** | spec.md:14 = `tags: "git, doctrine, manager-git, sync, worktree"` (인용 콤마 문자열). 본인이 `moai spec lint SPEC-GIT-PROC-SAFE-001` 직접 실행: `0 error(s), 2 warning(s)`, exit=0 — ParseFailure 소멸 직접 관측(1차 Gap이던 바이너리 검증, RED→GREEN 관측으로 보완) |
| **F2** git-flow 모순 | **RESOLVED** | 3개 규범 표면 전부 확인: spec.md:34(승격 단계 모드 조건부 + 설정 측정치 인용: git-strategy.yaml 9행 `workflow: git-flow`, 29행 `auto_enabled: false`), plan.md:45 M1 3번 불릿(PR 모드 → PR 승격 / git-flow → 통합 창 규율로 develop 통합 워크트리 병합, 템플릿 중립 표현 유지 명시), acceptance.md:31 AC-GP-01a GREEN Then(3) 동일 조건부. research.md Gaps 절이 1차 감사(F2)가 발각한 조사 공백을 정직하게 기록하고 재측정치(2행 `mode: manual`, 9행, 29행)로 갱신 — 2행/9행/29행 본인 재판독으로 일치 확인. 신규 텍스트에 SPEC ID·dev-only 경로 없음(REQ-TN-001 정합) |
| **F3** tier 부재 | **RESOLVED** | spec.md:13 `tier: M` |
| **F4** REQ-TN-001 무 AC | **RESOLVED** | acceptance.md:16(AC 매트릭스 행) + :76-80 AC-TN-01 — 3개 템플릿 사본 대상 grep 조건(SPEC-ID 패턴 0건, `.claude/rules/local/` 0건, 언어 중립 훼손 신규 도입 없음), 이진 검증 가능. REQ→AC 커버리지 8/8 |
| **F5** GEARS 등기 이탈 | **RESOLVED** | spec.md:70 REQ-AC11-001 → `**When** ... 편집도 수행해서는 안 되며 ... 관측해 ... 남겨야 한다`(Event-driven); spec.md:74 REQ-TF-001 → `**When** ... 적용해야 하며 ... 접촉해서는 안 된다`(Event-driven+금지). MP-2 이제 8/8 엔트리 5패턴 정합 |
| **F6** 마커 토큰 오탐 | **RESOLVED** | plan.md:82 "미해결 명확화 항목이 없다...", progress.md:21 "미해결 명확화 표식 없음" — 리터럴 토큰 제거. 본인 `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-GIT-PROC-SAFE-001/` → **적중 0건**(exit 1) 직접 관측 |
| **F7** AC-GP-01b 템플릿 핀 | **RESOLVED** | acceptance.md:40 — 템플릿 50/55-60/62행 핀 추가("동일 내용·동일 행"). 주장은 참: 이터레이션-1에서 본인 `diff`로 바이트 동일(0행) 확인済 |

## New Finding N1 — MovingRefUnpinned 경고 2건 (판정: NICE-TO-HAVE)

`moai spec lint`가 acceptance.md:39와 research.md:29에 `MovingRefUnpinned` WARNING을 보고(둘 다 `git rev-parse main == origin/main` 문구) — 이터레이션-1 F1-F7 목록 밖 신규 항목으로 본 델타 패스에서 판정했다.

- **판정: 위음성(None-blocking) 잔존 관측, NICE-TO-HAVE.** 두 적중 모두 **인용문 내부**다 — 수리 대상인 spec-workflow.md의 폐기 예정 post-condition 텍스트(RED 근거 인용)를 그대로 옮긴 것. VCI §2.1 앵커-대-주제 판별식에서 둘 다 해당하지 않는다: ANCHOR 아님(그 ref에서 측정한 것이 없음), SUBJECT 아님(origin/main의 현재 위치에 대한 주장이 아님 — 삭제할 텍스트의 인용). RED 셀의 측정 baseline은 이미 커밋 SHA로 고정돼 있다(acceptance.md:3, research.md:3 `develop c9ceff175`).
- **차단성 없음의 기계적 근거**: `internal/spec/lint_movingref_test.go:163-165`(AC-MRG-005) — MovingRefUnpinned는 `Advisory`로 발령되며 `moai spec lint`와 `--strict` 모두 exit에 영향 없음(라이브 코퍼스에 110건 존재가 그 전례). 본인 관측 exit=0과 일치.
- **권고(선택)**: 형태 탐지기를 정식으로 침묵시키려면 두 행에 R3 면제 표기(`<!-- moving-ref-ok: RED 셀이 인용하는 폐기 대상 텍스트; 측정 앵커 아님 -->`)를 추가. run-phase 편집과 무관한 문서 미화 — 수리 필수 아님.

## Remaining Findings (수리 요구 없음 — 잔존 관측)

- R1 (NICE-TO-HAVE): AC-TN-01이 마일스톤 검증 열(plan M1-M4 각각의 "검증:" 행)에 명시적 소유자가 없다 — AC 자체에 실행 시점(M1-M3 완료 후)은 기록돼 있고 M4의 "§E 전 항목"이 최종 포괄하므로 실행 가능. 다음 문서 정리 시 M3 또는 M4 검증 열에 "AC-TN-01" 토큰 추가 권고.
- N1 (NICE-TO-HAVE): 위 MovingRefUnpinned R3 면제 표기.

## Iteration-2 Category Scores

| Dimension | iter1 | iter2 | 근거 |
|-----------|-------|-------|------|
| D1 Evidence integrity | 9.5 | 9.5 | 전체 인용 재확인 + AC-GP-01b:40 신규 핀 참(바이트 동일은 1차에서 직접 측정). research.md 신규 측정치 2/9/29행 재판독 일치 |
| D2 Requirements quality | 8.0 | 9.5 | GEARS 8/8 정합, REQ→AC 8/8(AC-TN-01), AC-AC11-01 검증 전용 설계 유지 |
| D3 Feasibility & constraints | 6.0 | 9.5 | F2 해소 — 승격 단계가 관측된 git-flow 구성과 조건부 정합, Phase D 소멸 논리 유지, 3규범 표면 일관 |
| D4 Scope discipline | 9.0 | 9.5 | 수리가 SPEC 디렉터리에만 국함(6 대상 파일 `git diff` 빈 출력 직접 관측), F2의 범위 경로 기록도 research.md에 보강 |
| D5 Neutrality & mirror | 8.0 | 9.5 | AC-TN-01 구체 grep 조건, M1 불릿에 템플릿 중립 표현 유지 지시 내장 |
| D6 Milestone/AC traceability | 8.0 | 9.0 | REQ↔AC 전수 매핑. 감점 0.5: AC-TN-01 명시적 마일스톤 검증열 소유자 부재(R1) |

## Iteration-2 Must-Pass

MP-1 PASS(8 REQ, 번호 일관) · MP-2 PASS(8/8 GEARS) · **MP-3 PASS**(lint exit 0 본인 관측) · MP-4 N/A · MP-5 PASS · MP-6 PASS · MP-7 PASS(grep 적중 0건 본인 관측).

## Iteration-2 Gaps

- plan.md "×2" 발신 보고와 관측 편집 지점 수의 계산 차이는 원인 규명하지 않음(내용 귀속 검증으로 대체 — 결함 영향 없음).
- delivery.md "~94 diff 행" 수치는 이번 패스도 재측정하지 않음(범위 밖 관측치; 길이 델타·마커 오프셋으로 분기 존재는 확인됨).
- 2차 late-branch 잔존 표면 전수 스윕은 SPEC 계약상 run-phase M4 몫으로 유지.
- MovingRefUnpinned의 Advisory 판정은 `internal/spec/lint_movingref_test.go:163-165` 소스 판독 기준 — `--strict` 실제 실행 재현은 하지 않음(테스트 코드가 두 동사 모두를 단정).

## Run-phase Advisory

plan-audit PASS — 다만 일반 게이트와 별개로 Implementation Kickoff Approval(인간 게이트)은 그대로 요구된다. run-phase 착수 시: (1) AC-TN-01 grep을 M1-M3 편집 직후 실행해 즉시 피드백 확보, (2) M4의 2차 late-branch grep 스윕은 blocker 보고 경로 유지, (3) `make agents-emit` 필요성 보고는 plan §D 5번 그대로.
