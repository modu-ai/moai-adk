---
id: SPEC-CODEX-REVIEW-OWNERSHIP-001
title: "codex 리뷰의 소유권을 재배치한다 — 카드 리뷰는 레인의 단계, 비카드 세션은 설정으로 제외, 모든 세션에 온디맨드 자기 리뷰"
version: "0.1.0"
status: in-progress
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
amendment_of: SPEC-CODEX-GATE-SCOPE-001
related_specs: [SPEC-CODEX-GATE-SCOPE-001, SPEC-CODEX-REVIEW-TARGET-001, SPEC-MOAI-MCP-SERVER-001, SPEC-CODEX-AUDIT-GATE-AXES-001]
tags: "codex, glm, review-gate, self-review, mcp, lane, kanban, scoping, t1422, t1404"
---

# SPEC-CODEX-REVIEW-OWNERSHIP-001 — codex 리뷰의 소유권을 재배치한다

카드: **t1422** (운영자 지시 2026-10-02, Class C — t1404 인접. 카드 본문의 수치는 리더 제공이며 §A.6에 출처를 표기한다)

## HISTORY

- 2026-10-02 · v0.1.0 · manager-spec · 최초 작성. 카드 t1422. 측정 원천: 본 트리 `.moai/worktrees/t1422`(브랜치 `WT-codex-review-lane-scope`)에서의 코드 좌표 직접 판독(§A). 결정 D1·D2·D3는 `.moai/reports/t1422/jev-decisions.md`(Jev 판정 기록 — 표시 전용 신호, 결정 권한은 레인·리더 소관)와 코디네이터 지시(D2 조건부 구속)를 근거로 하며, progress.md「Decision Log」에 원문을 옮겼다.
- 2026-10-02 · v0.1.0 · manager-spec · 결정 확정 반영(제자리 갱신, 버전 불변). 오픈 결정 Q1-Q8 이 운영자 위임(Jev `jev-1.13.0`)·리더 승인으로 해결됐다. 해결 표는 plan.md §G.
- 2026-10-02 · v0.1.0 · manager-spec · plan-audit 1회차(FAIL 0.76, `.moai/reports/t1422/plan-audit.md`) 반영(제자리 갱신, 버전 불변). CLI 거울(구 REQ-CRO-011)을 이 SPEC 에서 **삭제**했고(Jev 0.83), 요구를 14건으로 재번호했다. D4(skip 이 `WT-` 접두 세션을 삼키는 경로 차단)·D5(리더 제외는 SPEC 밖 운영자 행위; 설정 파일은 추적 파일의 로컬 수정본)·D7(REQ-CRO-008 을 검증 가능한 범위로 축소)·D8(비추적 경로 목록·`truncated`)·D13(리더 측 GAP 규칙)을 요구에 반영했다. Claude Stop 게이트의 asyncRewake 전환은 같은 카드의 **형제 SPEC**으로 분리됐다(§E).

- 2026-10-02 · v0.1.0 · manager-spec · plan-audit 2회차(FAIL 0.79, `.moai/reports/t1422/plan-audit-iter2.md`, 감사 트리 `a0d801409`; 마지막 허용 반복) 반영(제자리 갱신, 버전 불변). R1: advisory·범위 메타데이터를 두 스코프·비카드 조기 반환 전부에서 값으로 단정하고 빈 자료 경로를 정의했다(REQ-CRO-010, AC-010 확장 + AC-016 분리). R2: env 행렬에 런처가 실제로 쓰는 값을 넣고 정책 파일의 환경 참조 정적 가드를 더했다(AC-005). N1: 두 MCP 규칙 사본은 바이트 동일이어야 한다(문장 정정)와 두 가드 시험을 목록에 올렸다(REQ-CRO-014). N3: primary 로컬 설정은 이동하는 운영자 상태다 — 시각 표기 관측으로 낮추고 이 SPEC 이 그 값에 의존하지 않음을 명시(§A.3). N5: 리더 문장을 조건부로(REQ-CRO-013). N9: 진행 알림을 호출하지 않음을 범위 밖에 명시. N11: 형제 SPEC 이 이미 존재하므로 §E 서술을 고치고 핸들러 배선 경계를 교차 확인했다.

### Amendments

- 2026-10-02 — 후속 개정(successor amendment). 대상: `SPEC-CODEX-GATE-SCOPE-001` **REQ-CGS-003**("카드 세션이 아닌 세션의 대상은 변하지 않는다"). prior_completed_version: `0.1.0`. prior_completed_sha: `3f4000bbe`(본 트리 `git log -- .moai/specs/SPEC-CODEX-GATE-SCOPE-001/spec.md` 실측 — 그 SPEC의 3-phase close 커밋. 부모 progress.md §E.4 의 `sync_commit_sha` 는 측정 시점에 `pending-backfill-sync` 로 남아 있다). 근거: REQ-CGS-003 은 비카드 세션 전부를 "트리 전체 미커밋 리뷰"로 못 박아, 설정으로 그 리뷰를 끌 길이 없었다(§A.2). 범위: 적용 조건 한 줄만 좁힌다 — REQ-CGS-003 의 "카드 스코프가 아닌 세션"이 **"카드 스코프가 아니고 `WT-` 접두 브랜치도 아니며 `tree_scope` 가 `skip` 이 아닌 세션"** 이 된다(REQ-CRO-003). `skip` 이 아닐 때의 요청 형태·fail-open·REQ-CRT-006 회귀선은 그대로이고, REQ-CGS-001·002·004~010 은 건드리지 않는다. 부모 SPEC 본문은 수정하지 않는다(완료 SPEC 본문 불변 — 개정은 이 SPEC이 선언한다).

---

## §0 지배 원칙 [HARD]

> **리뷰는 변경을 귀속할 수 있는 세션이 자기 몫으로 한다. 턴 종료 게이트는 그 귀속이 성립하는 곳에서만 돈다.**

현행은 "누가 무엇을 리뷰하는가"를 턴마다 도는 강제 하나로 풀려 한다. 결과는 두 가지다. 첫째, 귀속할 수 없는 변경(공유 primary 체크아웃의 미커밋 전체)이 리뷰 대상이 되고, 그 판정이 턴을 막는다. 둘째, 리뷰를 **요청할 길**이 없다 — 게이트가 돌 때만 리뷰가 일어난다. 이 SPEC은 소유권을 셋으로 가른다. (1) 카드 리뷰는 레인의 몫이고 카드 diff로 한정한 카드 단계로 수행한다. (2) 리더 자신의 내부 산출물은 리더가 직접 리뷰한다 — **리더 세션의 턴 종료 게이트 제외는 이 SPEC의 코드가 아니라 SPEC 밖 운영자 행위로 전달된다**(§A.3, §E, progress.md 인계 항목). 이 SPEC이 만드는 것은 그 행위가 쓸 수 있는 설정 키다. (3) 어느 세션이든 자기 카드·자기 코드를 codex나 GLM에 직접 요청할 수 있다. 리뷰 품질·판정 문구·900s 예산·감사(audit) 영수증 의미는 별개 축이며 이 SPEC은 건드리지 않는다.

---

## §A 배경 (측정)

### A.0 측정 원천과 핀

좌표는 모두 본 트리 기준이며 줄 번호는 2026-10-02 판독값이다. 측정 HEAD 핀: `3ae43ed8e78ffa673ca238227df6ca7202c1ce70`(`git rev-parse HEAD`, 이 SPEC 디렉터리를 추가한 커밋 — 인용 경로는 c50da9c2f 와 동일), 브랜치 `WT-codex-review-lane-scope`. 이 레인이 라이브 레인·라이브 codex·라이브 GLM 을 실행해 관측한 항목은 없다 — 아래 서술은 코드·파일 판독이며, 관측하지 않은 전제는 해당 항목에 Gap으로 표기한다.

### A.1 현행 게이트의 구조 — 비카드 세션을 가르는 축이 없다

- `internal/cli/codex_review_gate.go:69-119` `HandleCodexReviewGate` 의 판정 순서: 비활성(:71) → `stop_hook_active`(:74) → 스코프 해상(:81, 로그 :82) → 스코프 셀프게이트(:83) → codex 조회(:87) → 리뷰(:103) → fail 차단(:112-117). 이 순서의 어디에도 트리 스코프 세션을 설정으로 제외하는 분기가 없다.
- `internal/cli/codex_review_scope.go:84-104` `resolveReviewScope`: `WT-` 접두 브랜치이고 `git merge-base develop HEAD` 가 계산될 때만 카드 스코프(:97-103), 그 외 전부 트리 스코프. `WT-` 브랜치인데 기저 계산이 실패하면 클래스는 트리이고 `Branch` 는 `WT-…`, `Basis` 는 "card branch detected but merge base unavailable"(:97-102). 기저 브랜치는 상수 `cardBaseBranch = "develop"`(:55). 요청 조립은 `:169-180` `reviewRequestParams` — 카드는 `baseBranch` 대상(재계산한 merge-base SHA)+카드 트리 cwd, 트리는 `uncommittedChanges`+해상 트리 cwd.
- Codex Stop 체인 멤버 6(`internal/cli/codex_stop_chain.go:613-669`)과 receipt 생산자(`internal/cli/codex_review_receipt.go:126-174`)는 같은 `reviewScopeResolver` 를 부른다(`:624`, `:134`). `moai verify codex-review` 는 사용자가 명시적으로 돌리는 명령이다(`--project-root` 플래그 보유, `moai verify codex-review --help` 실측).
- multi 리뷰 게이트(`internal/cli/multi_review_gate.go:75`)는 스코프 해상기를 쓰지 않고 `reviewGateChangeDetector(projectDir)` 를 직접 부른다 — codex 리뷰 게이트의 스코프 개념을 공유하지 않는다. 기본 OFF(`workflow.multi.review_gate.enabled: false`).

### A.2 REQ-CGS-003 이 설정 탈출구를 막고 있다

SPEC-CODEX-GATE-SCOPE-001 REQ-CGS-003 은 비카드 세션의 대상을 "해상 트리의 미커밋 전체, 사전 형태와 동일"로 고정했다(리더 세션을 명시적으로 포함). `workflow.codex.review_gate.enabled` 는 **세션 전부에 한 번에** 걸리는 스위치라서, "카드 세션은 리뷰하고 비카드 세션은 건너뛴다"를 표현할 수 없다.

### A.3 어느 설정 루트가 `enabled` 를 정하는가 — 그리고 리더 제외는 무엇이 전달하는가

- 셸 래퍼 `.claude/hooks/moai/handle-codex-review-gate.sh` 는 `CLAUDE_PROJECT_DIR`(없으면 `$PWD`)의 `.moai/config/sections/workflow.yaml` 에서 `workflow.codex.review_gate.enabled` 를 순수 셸로 읽고, off 면 exit 0 한다.
- Go 핸들러는 `runCodexReviewGate`(`codex_review_gate.go:192-211`)에서 `resolveProjectDirFromInput`(`:260-270` — ProjectDir → CWD → `CLAUDE_PROJECT_DIR`)으로 projectDir 를 정하고, `reviewGateConfigRoot`(`:220-226`, `auditreceipt.StoreRoot`)가 돌려준 루트의 `workflow.yaml` 로 `readCodexReviewGateEnabled`(`internal/cli/mcp_codex.go:2385`)를 호출한다(`:200`). 설정 루트는 `.moai` 를 추적하지 않는 저장소의 연결 워크트리일 때만 primary 로 바뀌고, 그 외에는 projectDir 그 자체다(SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-008). `HandleCodexReviewGate` 자체는 설정 루트가 아니라 `projectDir` 를 받는다(`:69`).
- Codex 체인 멤버 6은 `readCodexReviewGateEnabled(c.root)`(`codex_stop_chain.go:618`)로 **세션 트리를 그대로** 설정 루트로 쓴다 — `reviewGateConfigRoot` 를 거치지 않는다(기존 비대칭, 이 SPEC은 고치지 않는다).
- **설정 파일의 출처(관측, 본 트리, 이 레인이 실행).** `git ls-files --error-unmatch .moai/config/sections/workflow.yaml` → `.moai/config/sections/workflow.yaml` — **추적 파일**이다. `git show main:.moai/config/sections/workflow.yaml` 에서 `review_gate` 는 0 건(`grep -c` → 0, exit 1) — main 의 커밋본에는 게이트 키가 없다. primary 체크아웃의 HEAD 는 `cat /Users/goos/MoAI/moai-adk-go/.git/HEAD` → `ref: refs/heads/main` 이고, 그 작업 사본(`/Users/goos/MoAI/moai-adk-go/.moai/config/sections/workflow.yaml`)은 main 의 커밋본과 **대폭 다른 로컬 수정본**이다(최초 측정 `diff … | grep -c '^[<>]'` → 429 줄). 즉 primary 의 설정은 "비추적 로컬 파일"이 아니라 **추적 파일의 로컬 수정본**이다. 본 트리(develop 계열)의 추적본에도 `review_gate` 키가 없다(`grep -c review_gate .moai/config/sections/workflow.yaml` → 0).
- **이동하는 좌표 — 값에 의존하지 않는다.** 그 로컬 수정본의 `codex.review_gate.enabled` 값은 운영자 상태라 시간에 따라 바뀐다. 이 SPEC 의 이전 판은 `enabled: true`(`:132-134`)를 서술했으나 **2026-10-02 16:39 KST 재관측**에서는 `codex:\n    review_gate:\n      enabled: false` 였고 파일 수정 시각은 같은 날 15:32 였다(`sed -n 127,137p`, `ls -l`; 누가 바꿨는지는 관측하지 않았다). 이 관측은 그 시각의 값일 뿐 표준 사실이 아니며 이 SPEC 의 어떤 요구도 그 값에 의존하지 않는다: `enabled` 가 `false` 면 게이트가 꺼져 `tree_scope` 가 필요 없고, `true` 면 `tree_scope: skip` 이 리더를 가른다. `tree_scope` 의 가치는 저장소 전체에 `enabled: true` 가 걸린 배포에서 비카드 세션을 가르는 것이다.
- 코드 판독에 따른 귀결(라이브 레인 미관측): 카드 워크트리에 cwd 를 둔 레인 세션의 Go 측 `enabled` 는 그 워크트리의 추적 사본에서 읽혀 이 저장소에서는 off 로 읽힌다. primary 에 앉은 세션(리더 포함)의 `enabled` 는 위 로컬 수정본의 그때그때 값에 따른다.
- **리더 제외의 전달 경로(솔직한 서술).** 이 SPEC 이 넣는 것은 `tree_scope` 키와 그 판독·적용이다. 리더 세션이 게이트에서 빠지려면 누군가 primary 의 `workflow.yaml` 에 `tree_scope: skip`(또는 `enabled: false`)을 써야 한다. 그 쓰기는 **착지 뒤 리더가 하는 운영자 행위이며 이 SPEC 의 요구가 아니다.** 지속성: `moai update` 는 새 템플릿에 없는 기존 키를 유지한다(`internal/cli/update/backup/merge.go:20-22` 문서 주석 실측). 그러나 primary 는 `main` 에 있고 `main` 은 릴리스 PR 로만 전진하므로, 릴리스가 이 파일을 바꾸는 순간 로컬 수정 429 줄과 부딪힐 수 있다 — 이 충돌에서 키가 살아남는지는 **시도하지 않았고 미해결**이다. 인계 항목은 progress.md 에 둔다.

### A.4 리뷰 도구 보유 현황

`codex_audit`/`glm_audit` 를 `tools:` 에 가진 에이전트는 `plan-auditor`(`.claude/agents/moai/plan-auditor.md:7`)와 `sync-auditor`(`.claude/agents/moai/sync-auditor.md:9`) 둘뿐이다(`grep -l mcp__moai__codex_audit .claude/agents/moai/*.md` → `sync-auditor.md`, `plan-auditor.md` 만 출력, 본 트리). manager-develop·manager-docs·manager-lead 의 로컬(C1) `tools:` 줄은 각 `:9`/`:9`/`:10` 이고 배포 미러(C2)는 `:10`/`:9`/`:10` 이다. 메인 세션은 에이전트 `tools:` 제약을 받지 않으므로 직접 호출할 수 있다고 추정되나 이 레인은 메인 세션의 도구 목록을 관측하지 않았다 — **Gap**.

`codex_audit` 는 매 호출이 감사 영수증을 남긴다(`internal/cli/mcp_codex.go:1931`, `:1954` — `recordAuditReceipt` 호출 2곳, `grep -c recordAuditReceipt internal/cli/mcp_codex.go` → 2). 영수증 가드는 같은 트리에서 감사관 시작 이후에 만들어진 `codex_audit`/`audit_multi` 영수증을 PASS 근거로 인정한다(`internal/auditreceipt/store.go:596-601`). 또 `applyGateUnmet`(`mcp_codex.go:1929`, `:1952` 호출, 정의 `:1974`)은 `workflow.audit.gates.codex: required` 일 때 리뷰 부재를 `fail`+`gate_unmet` 로 바꾼다. `internal/mcp/catalog_test.go:30-37` 은 `codex_audit`·`audit_multi` 를 영수증 때문에 **쓰기 가능** 도구로 분류한다.

### A.5 `baseBranch` 대상 해상 결함 (b)

- `codex_audit` 의 `baseBranch` 대상은 호출자가 지정할 수 없고 서버가 정한다 — `coerceCodexReviewTarget`(`mcp_codex.go:1218`, `baseBranch` 분기 `:1230-1236`)→`resolveReviewBaseBranchName`(`internal/cli/mcp_review_material.go:131-144`): 원격 기본 헤드, 그다음 `main`. `glm_audit` 는 별개 체인으로 `resolveReviewMergeBase`(`mcp_review_material.go:92-103` 부근: `origin/HEAD`, `origin/main`, `main`)의 `git diff <base>...HEAD`(`reviewDiffArgs` `:73-87`)를 쓴다.
- 게이트의 카드 diff 기저는 이 둘과 다르다 — `git merge-base develop HEAD`(`codex_review_scope.go:132`, 로컬 `develop`, 매 평가 재계산).
- 본 트리 측정: `git symbolic-ref --short refs/remotes/origin/HEAD` → `origin/develop`. 그러므로 **이 저장소에서** codex 쪽은 이름 `develop` 으로 해상되어 "main 과 비교"하지는 않는다. 결함은 구조적이다: 호출자가 기저를 지정할 수 없고, 두 감사 도구의 체인이 서로 다르며, 어느 쪽도 게이트의 판별·재계산 규율을 쓰지 않는다. 로컬 `develop` 이 `origin/develop` 보다 앞서 있고(배치 push 전) 카드가 그 커밋을 흡수했다면 GLM 쪽 `origin/HEAD` 기저는 다른 카드의 커밋을 카드 diff 로 끌어들일 수 있다 — 이 추론은 구조 판독이며 라이브 재현은 하지 않았다 — **Gap**.
- 이 결함의 수리 범위: 자기 리뷰 도구의 카드 스코프로만 닫는다. 감사 도구 쪽은 카드 **t1426** 으로 분리됐다(§E).

### A.6 동기와 출처 표기

카드 본문(리더 제공)은 "리더 Stop 훅 1회당 codex 리뷰 9~12분·Bash 80~105회, primary 282파일 미커밋 diff 전체를 매 턴 재검토"를 기록한다. 본 레인은 이 수치를 독립 재측정하지 않았다 — **리더 제공**. 이 SPEC이 의존하는 것은 수치가 아니라 §A.1-A.3 의 구조(비카드 세션을 설정으로 가르는 축 부재)다. 인접 카드 t1404(게이트 스코핑 구멍)의 항목별 처분은 plan.md §D 에 코드 인용과 함께 둔다.

---

## §B 요구사항 (GEARS)

세 묶음이다. **A. 비카드 세션 정책**(REQ-CRO-001~006), **B. 온디맨드 자기 리뷰 표면**(REQ-CRO-007~010), **C. 보유·교리·정합**(REQ-CRO-011~014). 개수 규칙: `### REQ-` 제목 개수 = 14.

### A. 비카드 세션 정책 (결정 D1)

### REQ-CRO-001 — 비카드 세션 정책 키 (Ubiquitous)

The codex review gate configuration shall define `workflow.codex.review_gate.tree_scope` with the allowed values `review` and `skip`, read only from its place under the `workflow` root; a missing key, an empty value, a misplaced or commented-out key, or any other value shall read as `review`, and the comparison shall ignore surrounding whitespace and letter case.

키 이름과 값은 `workflow.codex.review_gate.tree_scope` = `review`|`skip` 으로 확정됐다(decision-index.md Q2, 운영자 위임 판정, 리더 최종). `review` 는 현행 동작이고 배포 기본이다. 알 수 없는 값을 `review` 로 읽는 방향은 "덜 검사하는 쪽으로 조용히 틀어지지 않는" 방향이다.

### REQ-CRO-002 — 브랜치 증거가 없는 트리 스코프 세션의 skip (Where + While)

**Where** the codex review gate is enabled, **While** the session's scope class is tree, the resolver found no `WT-` prefixed branch, and `tree_scope` is `skip`, the review gate shall allow the turn without running the self-gate detector, the reviewer lookup, or any review, and shall record the skip and its basis in the gate's structured log.

Claude Stop 훅과 Codex Stop 체인 멤버 6 두 자동 경로 모두에 적용된다(REQ-CRO-006). "`WT-` 접두 브랜치를 찾지 못함"은 해상기의 결과(브랜치 이름이 접두를 갖지 않거나 읽을 수 없음/detached)로 판정하며 §A.1 의 세 `Basis`("no session tree", "no card branch (unreadable or detached)", "no card branch: <branch>")가 이에 해당한다. 로그는 REQ-CGS-010 의 스코프 로그 행과 구별되는 한 행(클래스·키 값·skip 사유)이다.

### REQ-CRO-003 — 그 밖의 트리 스코프 세션의 기존 동작 유지 (While) — REQ-CGS-003 개정

**While** the session's scope class is tree and either `tree_scope` is not `skip` or the resolver found a `WT-` prefixed branch, the review gate shall scope the review to the whole uncommitted changes of the resolved tree, shape-identical to its form before SPEC-CODEX-GATE-SCOPE-001 — the same path SPEC-CODEX-REVIEW-TARGET-001 REQ-CRT-006 pins.

**개정 선언(명시).** 이 요구는 SPEC-CODEX-GATE-SCOPE-001 REQ-CGS-003 의 후속 개정이다. 바뀌는 것은 적용 조건 한 줄뿐이다 — "카드 스코프가 아닌 세션" → "카드 스코프가 아니고 `WT-` 접두 브랜치도 아니며 `tree_scope` 가 `skip` 이 아닌 세션". 바뀌지 않는 것: `skip` 이 아닐 때의 요청 형태(`uncommittedChanges`+해상 트리 cwd), fail-open(REQ-CGS-008), REQ-CRT-006 회귀선. 이유: REQ-CGS-003 은 비카드 세션의 리뷰를 항상 켜진 것으로 못 박아 설정으로 끌 수 없었다(§A.2). `tree_scope` 가 부재이거나 `review` 이면 REQ-CGS-003 의 이전 문장과 관측상 동일하다.

### REQ-CRO-004 — 카드 브랜치 세션은 정책의 영향을 받지 않는다 (While)

**While** the resolver found a `WT-` prefixed branch in the session tree — whether or not the merge base could be computed — the review gate shall not apply `tree_scope: skip`: it shall review as it does without the key (the card diff when the merge base is available, the whole uncommitted tree when it is not) and shall keep the scope log row whose basis names the unavailable merge base.

D4 수리. `WT-` 브랜치인데 `develop` 참조가 없거나 기저 계산이 실패한 세션은 클래스가 트리로 떨어지지만(§A.1) 카드 증거가 있으므로 skip 으로 삼켜져서는 안 된다 — 오늘은 전체 트리 리뷰로 떨어져 **리뷰 쪽으로 틀어진다**. `tree_scope` 는 "카드 증거가 없는 세션이 무엇을 하는가"만 정한다.

### REQ-CRO-005 — 정책 판정의 입력은 해상 결과와 키 값뿐이다 (Ubiquitous + shall not)

The tree-scope policy decision shall take exactly two inputs — the resolver's result for the session tree and the `tree_scope` value — and shall not be derived from session environment labels, a leader or lane role, or launcher mode.

REQ-CGS-004 의 연장이다. 리더를 env 로 검출하지 않는 근거: 진행 중인 카드 t1399 가 `MOAI_KANBAN*` env 계열을 삭제하고 있고, REQ-CGS-004 가 env 를 스코프 판별 입력에서 이미 배제했다. 역할 차이는 세션이 어느 트리에 앉았는가와 그 저장소의 설정으로만 표현된다.

### REQ-CRO-006 — 두 자동 경로와 설정 루트 (Ubiquitous)

The Claude Stop hook and the Codex Stop chain's codex review member shall apply the tree-scope policy through one shared function, reading `tree_scope` from exactly the config root each path already reads `enabled` from, so that the two paths cannot disagree on the policy for the same session state; the receipt producer `moai verify codex-review` is an explicit request rather than an automatic gate and shall run its review whatever `tree_scope` holds.

설정 루트는 경로마다 `enabled` 와 같다: Claude 훅은 `reviewGateConfigRoot(projectDir)`, Codex 체인은 `c.root`(§A.3). **배선 결정(D15):** `HandleCodexReviewGate` 의 시그니처는 바꾸지 않는다 — 스코프 클래스가 트리일 때만 핸들러 안에서 기존 `reviewGateConfigRoot(projectDir)` 로 루트를 다시 구해 주입 가능한 판독기로 키를 읽는다(카드 세션에는 추가 git 호출이 없다). 이 핸들러는 형제 SPEC(§E)이 시그니처·설정 루트 배선을 다시 만질 수 있으므로 이 SPEC 은 그 결정을 선취하지 않는다. 두 경로의 기존 비대칭은 유지하되 새 키가 `enabled` 와 다른 루트를 읽는 일은 없어야 한다. 명시 실행 `moai verify codex-review` 는 사용자가 요청한 리뷰라 정책으로 끄지 않는다(온디맨드 원칙).

### B. 온디맨드 자기 리뷰 표면 (결정 D2 — 코디네이터 조건부 구속, 리더 승인)

### REQ-CRO-007 — 자기 리뷰 도구 두 개 (Ubiquitous)

The moai MCP server shall expose two read-only tools, `codex_review` and `glm_review`, each requiring a `scope` input whose allowed values are `card` and `uncommitted`, accepting an optional `model` input and the `project_root` input under the contract the existing `project_root` tools share, and declaring in its output schema the result fields REQ-CRO-010 names.

`scope` 는 필수다 — 기본값이 있으면 레인이 `scope` 를 빠뜨렸을 때 카드의 커밋분을 놓치고 미커밋분만 보는 조용한 다른 리뷰가 된다. `focus` 입력은 두지 않는다(요구 없음). 두 도구는 감사(audit) 도구가 아니다 — `codex_audit`·`glm_audit`·`claude_audit`·`audit_multi` 의 입력·출력·영수증 의미는 바뀌지 않는다.

### REQ-CRO-008 — 카드 스코프는 게이트와 같은 해상기가 정한다 (When)

**When** a self-review tool is called with `scope` `card`, it shall resolve the scope through the same scope resolver the turn-end gate uses, recomputing the merge base on every call without pinning it and without introducing a second discriminator; **when** the resolved class is card, the `codex_review` tool shall send the review request the gate sends for that scope, and the `glm_review` tool shall build its material as the diff from that merge base to the card tree's working state with the runtime-managed prefixes excluded by pathspec; **when** the resolved class is not card, the tool shall return an `inconclusive` result naming the cause, shall not call the reviewer, and shall not review any other change in its place.

이것이 결함 (b)의 수리다 — 호출자가 게이트와 **같은** 판별·같은 재계산 규율로 측정된 카드 diff 를 요청할 수 있다. 판별 입력은 REQ-CGS-004 그대로 `WT-` 브랜치뿐이다. **이 SPEC 이 주장하지 않는 것:** codex 가 `baseBranch` 대상으로 미커밋·비추적 파일까지 읽는지, `branch` 필드에 SHA 를 받는지는 관측되지 않았다(§A.5 Gap, plan.md §G(a)). 그 관측은 plan.md M3 의 기록 단계(`.moai/reports/t1422/live-probe/`)로 수행하며 AC 의 근거가 아니다. 런타임 관리 접두 제외는 GLM 자료 구성에 대한 주장이다 — codex 가 읽는 파일에 대한 주장이 아니다.

### REQ-CRO-009 — 미커밋 스코프 (When)

**When** a self-review tool is called with `scope` `uncommitted`, the `codex_review` tool shall send a request whose shape is identical to the tree-scope request of the turn-end gate for the named tree, and the `glm_review` tool shall build its material as the diff of the named tree against its HEAD with the same runtime-managed prefix exclusions.

리더·일반 세션이 자기 내부 산출물을 리뷰하는 경로다. 해상기를 거치지 않고 트리 클래스 요청을 직접 만들어 `WT-` 트리에서도 미커밋 요청이 나가게 한다. primary 에서 부르면 공유 작업 트리 전체가 대상이며 경로 제한 입력은 없다 — 한계로 도구 설명과 교리에 적는다.

### REQ-CRO-010 — 자기 리뷰는 조언이며 결과가 자기 범위를 밝힌다 (Ubiquitous + shall not)

Every self-review result, whatever its verdict, shall carry a machine-readable `advisory` marker with the value true, together with the scope, the base commit, the backend, the tree reviewed, and a `truncated` marker; the `glm_review` tool shall name every untracked non-runtime path of the tree in `excluded_untracked` for both scopes and shall set `truncated` when it cut its material to the size cap; a self-review tool shall not mint or consume an audit receipt, shall not apply a `required` audit gate or the audit model pins, and shall fail open to `inconclusive` on every missing-reviewer, error, or empty-material path, as SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012 requires.

판정(verdict) 값은 `codex_audit` 와 같은 어휘(`pass`/`fail`/`inconclusive`)를 쓰되 **구속력이 없다** — 도구 설명·출력 스키마·교리 세 곳에 "advisory(non-binding)"로 적는다. 감사 영수증을 만들지도 소비하지도 않으므로 PASS 근거로 인용될 수 없다(`store.go:596-601` 가드의 입력에 들어가지 않는다). `required` 게이트의 `fail`+`gate_unmet` 변환(§A.4)을 받지 않는다 — codex 부재는 언제나 `inconclusive` 다. 모델은 호출자의 `model` 입력이 있으면 그것, 없으면 백엔드 기본이다(Q6 잠정 — 감사 핀 미적용; 되돌림 비용은 해상기 호출 한 번). `codex_review` 의 `truncated` 는 이 도구가 자료를 만들지 않으므로 항상 false 이고 `excluded_untracked` 는 비어 있다.

**메타데이터는 모든 경로에서 값으로 정해진다(plan-audit 2회차 R1).** 위 필드는 pass·fail·inconclusive 어느 verdict 에서도, `card`·`uncommitted` 어느 스코프에서도, 그리고 카드 스코프를 요청받았으나 카드 트리가 아니어서 리뷰어를 부르지 않고 돌려주는 조기 `inconclusive` 에서도 채워진다. 값의 정의: `advisory` 는 true; `scope` 는 요청 값; `backend` 는 `codex`|`glm`; `base` 는 카드 스코프에서 호출 시점의 `git merge-base develop HEAD` 출력(SHA)이고 그 밖의 모든 경로(미커밋 스코프, 비카드 조기 반환, 빈 자료)에서 빈 문자열; `tree` 는 `project_root` 인자의 철자가 아니라 심볼릭 링크를 해소한 정규 루트다. 상수 문자열이나 인자 그대로의 값은 위반이다.

**빈 자료(empty material)의 정의.** 요청한 스코프에 리뷰할 변경이 없는 경우다 — 카드 스코프는 `git diff <merge-base> -- . <런타임 접두 exclude>` 가 비어 있고 비추적 비런타임 파일이 없을 때, 미커밋 스코프는 `git diff HEAD -- . <exclude>` 가 비어 있고 비추적 비런타임 파일이 없을 때. 두 백엔드 모두 이 경우 리뷰어를 부르지 않고(`codex_review` 는 리뷰 RPC 0 회, `glm_review` 는 HTTP 0 회) verdict `inconclusive` 와 변경이 없다는 요약을 위 메타데이터와 함께 돌려준다.

### C. 보유·교리·정합

### REQ-CRO-011 — 도구 보유 (Ubiquitous + shall not)

The agent definitions of manager-develop, manager-docs, and manager-lead shall list `mcp__moai__codex_review` and `mcp__moai__glm_review` in their tool lists, and `mcp__moai__codex_audit` and `mcp__moai__glm_audit` shall remain listed only by plan-auditor and sync-auditor.

적용 사본: 로컬 `.claude/agents/moai/*.md`(C1), 배포 미러 `internal/template/templates/.claude/agents/moai/*.md`(C2), 기계 방출 `internal/template/templates/.codex/agents/moai/*.toml`(C3, `make agents-emit`). 메인 세션(리더·레인 오케스트레이터)은 MCP 를 직접 호출한다. 카드 t1424 가 manager-develop 의 `tools:` 줄에 codex·glm 위임 도구를 추가하는 작업을 가진다 — 같은 줄을 편집하므로 병합 순서 위험을 plan.md §G 에 둔다.

### REQ-CRO-012 — 레인의 카드 리뷰 단계 (Ubiquitous)

The kanban lane doctrine shall name a `card-review` stage in an ordered stage list, positioned after the run-exit verification and before the integration step, in which the lane requests a card-scope self-review, writes the result to `.moai/reports/<card-id>/card-review.md` with the backend, the base commit, the verdict, the findings, and a disposition for each finding, and cites that path in the card's progress record; the review shall remain advisory and shall never replace the leader's evidence read or an independent audit.

단계의 구체형은 plan.md §E 에 둔다. 리뷰 부재(codex 미설치·`inconclusive`·도구 부재)는 PASS 가 아니라 기록된 사유다. 이 단계는 Stop 훅이 아니라 레인의 카드 단계이므로 턴을 막지 않고 카드당 한 번(수리 후 재리뷰 최대 2회) 돈다.

### REQ-CRO-013 — 리더 측 규칙 (Ubiquitous)

The kanban lane doctrine shall state that, with `tree_scope: skip` configured for the leader's checkout, the leader session carries no turn-end codex review gate and reviews its own internal output directly with the same tools, and that a card whose progress record neither cites a readable `.moai/reports/<card-id>/card-review.md` nor records a reason for its absence is a gap in the leader's completion read and stays in its column; the leader's declared evidence list for a card shall include that path, so a stage that stopped being run surfaces at the leader's read.

연속 발화(continued firing) 답: 단계가 조용히 멈춰도 리더의 완료 판독이 증거 경로 목록에서 `card-review.md` 부재를 gap 으로 읽는다 — 별도 훅은 없다(§E). 교리 문장은 **조건부**다 — 배포된 교리가 모든 저장소에서 "리더 세션에 게이트가 없다"고 단정하면 `enabled: true` 이고 `tree_scope: skip` 이 없는 저장소에서는 거짓이다(plan-audit 2회차 N5). 조건절에 `tree_scope: skip` 토큰이 있고, 같은 문장에 L9 의 앵커 `turn-end codex review gate` 가 들어 있다. 실제 제외는 §A.3 의 운영자 행위가 전달한다.

### REQ-CRO-014 — 목록·문서·템플릿 정합 (Ubiquitous)

The tool catalogue, the server registration, the `project_root` inventory in both rule-file copies and in the four-locale docs-site pages, the console's per-tool i18n entries, and the MCP tool counts in doctrine shall describe the two new tools identically, each verified by the parity test that already guards it, and the distributed template shall carry `tree_scope` only as a commented example while the shipped-key inventory, the settings schema, the console fields, and the i18n entries for it remain unchanged.

도구 수 45→47, `project_root` 선언 도구 20→22, 쓰기 가능 도구 집합은 불변(두 도구 모두 읽기 전용). 템플릿 형태는 decision-index.md Q1 확정(주석 예시)이다. 정합 대상 전수와 마일스톤 배정은 plan.md §F. **카탈로그 문서 사본 두 쌍은 바이트 동일이어야 한다**(`moai-mcp-tools.md`·`moai-mcp-tools-catalogue.md` 의 로컬 사본과 배포 미러 — `TestMCPToolCatalogueDocsStayMirrorIdentical`)이고, 카탈로그의 모든 개수 수치(총 도구 수 문장·`Tool families (N of the M tools` 머리의 41→43 / 45→47·`(\d+)-tool`·`of the N tools` 계열)는 레지스트리와 일치해야 하며(`TestMCPToolCatalogueFiguresMatchRegistry`), `moai-mcp-tools.md` 의 "Four of the twenty REQUIRE it" 문장은 "Four of the twenty-two" 로 고친다. 이 두 시험이 이 문서 쌍의 유일한 기계 강제다.

---

## §C AC 형태에 대한 구속 [HARD]

- AC 는 **게이트·도구가 조립한 요청의 대상 필드**(target·cwd), **GLM 자료의 구성**, **설정 판독값**, **도구 목록**(`tools/list` 입출력 스키마·`tools:` 줄), **영수증 저장소의 호출 전후 목록**을 관측한다. verdict 값 단독은 어떤 AC 의 근거도 되지 못한다 — 스텁 codex 는 요청과 무관한 값을 돌려줄 수 있다. SPEC-CODEX-GATE-SCOPE-001 §C 와 같은 규율이다.
- 해상기·정책 판독기는 기존 주입 seam(`reviewScopeResolver`, `reviewGateChangeDetector`, `codexLookPath`, 리뷰 RPC 호출 지점, `glmKeyLoader`/HTTP 스텁)으로 검증한다 — 라이브 codex·라이브 GLM 의존 AC 는 두지 않는다. 라이브 프로브는 기록되는 관측이지 AC 가 아니며 skip 은 통과가 아니라 **미관측(Gap)**이다.
- 카드 픽스처는 실제 git worktree(`WT-` 브랜치, develop 분기 후 커밋 1건+추적 파일 미커밋 1건+비추적 파일 1건+**추적된 런타임 접두 경로 1건**) + primary 측 외부 WIP 로 구성한다. 셀렉터 0매칭 초록 방지: RED·GREEN 실행은 `-v` 로 `=== RUN` 을 함께 관측한다(verification-completeness §1.1).

## §D 실행 순서 구속

1. 회귀선을 먼저 초록으로 고정한다: 트리 스코프 `uncommittedChanges` 요청 형태, 카드 스코프 요청 형태, `WT-`+기저 불가 폴백, fail-open, 감사 도구 `tools:` 보유 집합, `codex_audit` 의 `required` 게이트 동작. 변경 전 트리에서 초록을 관측한 뒤에 정책·도구를 얹는다.
2. 도구 계약(이름·입력·출력 스키마·advisory 필드)과 설정 키 이름·값을 **먼저** 고정한다 — 둘은 외부 표면이라 되돌리기 가장 어렵다(plan.md §B 순서).

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — 감사(audit) 도구의 대상·영수증 의미 (카드 t1426)

- `codex_audit`·`glm_audit`·`claude_audit`·`audit_multi` 에 `cardDiff` 대상을 추가하는 일, 그들의 `baseBranch` 해상 체인(`resolveReviewBaseBranchName`·`resolveReviewMergeBase`)을 바꾸는 일은 하지 않는다. 결함 (b)는 새 자기 리뷰 도구의 카드 스코프 요청으로만 닫는다(REQ-CRO-008) — **감사 도구 쪽 `baseBranch` 해상은 이 SPEC 이후에도 그대로이며 카드 t1426 이 소유한다.** 카드 발행은 리더 소관이고 이 레인은 카드를 발행하지 않는다(decision-index Q8 확정).
- 감사 영수증 저장소·가드(`internal/auditreceipt`, `internal/hook/audit_receipt_guard.go`)는 건드리지 않는다.

### Out of Scope — 형제 SPEC: Claude Stop 게이트의 asyncRewake 전환

- 같은 카드(t1422)의 형제 SPEC `SPEC-CODEX-REVIEW-ASYNC-001`(draft, 이 SPEC 에 `depends_on`)이 Claude Stop 게이트를 asyncRewake 방식으로 바꾸는 일을 다룬다. 그 SPEC 은 이미 작성돼 있다(plan-audit 2회차 N11). **핸들러 배선 경계 교차 확인(형제 plan.md §B.1·§G 위험 2 판독):** 두 SPEC 모두 `HandleCodexReviewGate` 의 시그니처를 바꾸지 않는다. 이 SPEC 의 삽입은 스코프 해상 직후·셀프게이트 앞(스코프 클래스가 트리일 때 설정 루트 재해상+`tree_scope` 판독+skip)이고, 형제의 삽입은 codex 조회 뒤(트리 루트·git 디렉터리·락·상태 키·기록)이며 실행기(`runCodexReviewGate`)의 출력 매핑을 바꾼다 — 삽입 위치가 겹치지 않고 skip 이 락보다 앞선다(형제 REQ-CRA-010). 이 SPEC 이 먼저 착지하며 형제의 결정을 선취하지 않는다.

### Out of Scope — t1404 의 잔여 항목

- sync 게이트 `WCI_EXCLUDES` 의 `.moai/reports/**` 누락(`sync-phase-quality-gate.sh:257-266`), `moai gpt` 문서-CLI 드리프트(카드 t1406), 검토 트리의 develop 대비 진부함 감지, 기지 발견 장부(`ledger.jsonl`) 사전 분류, 비카드 스코프의 런타임 관리 경로 확장, 검토자 CLI 의 빌드 노후는 이 SPEC에 흡수하지 않는다(plan.md §D 의 항목별 처분). 이 SPEC 착지 뒤 리더가 t1404 를 잔여 항목으로 편집하는 일은 sync 단계 인계 항목이며 코드 변경이 아니다(착지 전에는 닫지 않는다, Q4 확정).

### Out of Scope — 카드 스코프 Stop 게이트의 제거·재설계와 리더 제외의 실행

- 카드 스코프 Stop 게이트(REQ-CGS-001·002·004~010)의 코드와 옵트인 성격은 그대로 둔다. 추적된 `workflow.yaml` 에는 `enabled: true` 를 커밋하지 않는다(Q5 확정).
- 리더 제외를 실제로 일으키는 것 — 착지 **뒤** 리더가 primary 의 `workflow.yaml`(추적 파일의 로컬 수정본)에 `tree_scope: skip` 을 쓰는 일 — 은 운영자 소유 파일에 대한 운영자 행위이며 이 SPEC 은 그 파일을 편집하지 않는다. 이 SPEC 은 그 행위를 요구로 만족시킨다고 주장하지 않는다(§A.3, progress.md 인계 항목).
- multi 리뷰 게이트(`HandleMultiReviewGate`)에 스코프 개념을 도입하는 일은 하지 않는다.

### Out of Scope — 역할 검출·env 파싱

- 리더나 레인을 env·프로세스 역할·런처 모드로 검출하는 어떤 메커니즘도 만들지 않는다(REQ-CRO-005). `MOAI_KANBAN*` 삭제는 카드 t1399 소관이다.

### Out of Scope — 리뷰 내용·예산·콘솔·CLI 거울

- codex·GLM 이 무엇을 좋다고 보는지, 프롬프트, 900s 예산, 감사 모델 핀 정책은 바꾸지 않는다. `tree_scope` 를 웹 콘솔 필드·`shipped_key_inventory.yaml` 항목으로 노출하는 일은 하지 않는다 — 템플릿에는 주석 예시로만 싣는다(Q1 확정).
- 자기 리뷰의 **CLI 거울은 이 SPEC 에서 삭제됐다**(Jev 0.83). 결과: 호스트 도구 타임아웃과 오래된 MCP 서버 프로세스(새 도구가 `tools/list` 에 없음)에 대해 **GLM 쪽 우회로가 없다.** codex 쪽 대체 경로는 기존 `moai verify codex-review --project-root <tree>` 뿐이다(plan.md §G 잔여 위험).
- 레인의 카드 리뷰 단계를 훅으로 기계 강제하는 일, 카드 리뷰 증거 파일의 존재를 검사하는 게이트를 만드는 일은 하지 않는다(교리로만 규정한다).
- Claude 백엔드 자기 리뷰(`claude_review`), 경로 제한(pathspec) 입력, 백그라운드 잡 모델은 만들지 않는다.
- **진행 알림(heartbeat)은 새 도구에 넣지 않는다(plan-audit 2회차 N9).** 감사 도구는 긴 호출 중 `notifyMCPProgress`(`internal/cli/mcp_progress.go`)로 Claude Code 의 idle watchdog(stdio 기본 30분)을 재설정한다. 새 도구는 이 알림을 호출하지 않는다 — 한 호출은 동기식이고 상한이 900s 로 그 watchdog 창보다 짧다. 호스트 도구 타임아웃(별개 층)이 이 알림으로 늘어나는지는 관측하지 않았다 — 잔여 위험(plan.md §G 위험 6). 알림을 넣으려면 두 백엔드 호출 지점에 한 줄씩이며, 넣는다면 REQ 와 AC 가 늘어난다(decision-index Q13).

---

## §F 제약

- **fail-open 불변**(SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012): reviewer 부재·오류·빈 자료·inconclusive 는 게이트와 자기 리뷰 도구 모두에서 ALLOW/`inconclusive` 다. `tree_scope: skip` 은 ALLOW 방향이며 차단 경로를 만들지 않는다.
- **REQ-CRT-006 회귀선**: `skip` 이 아닐 때의 트리 스코프 `uncommittedChanges` 직렬화는 본 SPEC 이후에도 shape-identical 이다.
- **단일 해상기**: 게이트 두 경로와 자기 리뷰 도구는 하나의 스코프 해상기·하나의 요청 조립기를 쓴다. 두 번째 판별기·두 번째 기저 계산은 금지다(REQ-CRO-008).
- **단순성**(AGENTS.md §5): 기존 해상기·요청 조립기·RPC 드라이버·영수증 형식을 재사용하고 새 의존성을 들이지 않는다. 최소 변경 규모 추정과 3배 점검은 plan.md §H.
- **Template-First**: `.claude/`·`.moai/config/` 의 배포 대상 변경은 `internal/template/templates/` 에 같은 변경으로 들어간다. 배포 미러는 SPEC ID·REQ 토큰·카드 번호·날짜·커밋 SHA 를 담지 않는 중립 본문이다. 에이전트 정의는 C1·C2 를 손으로 고치고 C3(`internal/template/templates/.codex/agents/moai/*.toml`)는 `make agents-emit` 으로만 만든다. 로컬 전용 파일(`.claude/rules/local/`, `AGENTS.local.md`)은 미러하지 않는다. 목록·마일스톤은 plan.md §F.
- **하드코딩 금지**: `review`/`skip` 값 이름·기본값은 `internal/config` 의 상수·`defaults.go` 에서 단일 원천으로 정의한다.
- **측정 규율**: 검증은 대상 패키지(`./internal/cli/...`, `./internal/mcp/...`, `./internal/config/...`, `./internal/template/...`, `./internal/web/...`) 스코프로 돌린다. 전체 스위트 로컬 금지.
- **문서 4개국어**: docs-site 를 건드리면 같은 PR 에서 ko·en·ja·zh 를 함께 고친다.

## §G 참조

- `internal/cli/codex_review_gate.go` · `codex_review_scope.go` · `codex_review_receipt.go` · `codex_stop_chain.go:613-669` — 게이트·스코프·영수증·멤버 6 (§A.1)
- `internal/cli/mcp_server.go:316,341,493` — `codex_audit`/`claude_audit`/`glm_audit` 등록. `internal/cli/mcp_codex.go:1218-1236,1891-1957,1974` — 대상 해상·핸들러·`applyGateUnmet`. `internal/cli/mcp_review_material.go:55-144,174` — GLM 자료·기저 체인·`truncateDiff`
- `internal/mcp/catalog.go` · `internal/mcp/catalog_test.go:19,30-37` — 도구 목록 단일 선언·크기 불변식·쓰기 가능 분류
- `internal/auditreceipt/store.go:596-601` · `internal/hook/audit_receipt_guard.go` — 영수증 가드
- `.claude/rules/moai/core/moai-mcp-tools.md` · `moai-mcp-tools-catalogue.md` — 도구 지도·`project_root` 규칙
- `.claude/rules/moai/workflow/kanban-dispatch.md` · `kanban-dispatch-detail.md` — 레인 단계·완료는 증거로 읽는다
- `.claude/rules/local/gitflow-lane-protocol.md` §1·§8 — `WT-` 불변식·merge-base 규율
- `.moai/reports/t1422/jev-decisions.md` — 결정 D1~D3 판정 기록(표시 전용 신호). `.moai/reports/t1422/plan-audit.md` — plan-audit 1회차(로컬 증거)
- 관련 SPEC(전부 completed): SPEC-CODEX-GATE-SCOPE-001(개정 대상) · SPEC-CODEX-REVIEW-TARGET-001 · SPEC-MOAI-MCP-SERVER-001 · SPEC-CODEX-AUDIT-GATE-AXES-001
