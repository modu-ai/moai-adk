---
id: SPEC-CODEX-REVIEW-OWNERSHIP-001
title: "codex 리뷰의 소유권을 재배치한다 — 카드 리뷰는 레인의 단계, 비카드 세션은 설정으로 제외, 모든 세션에 온디맨드 자기 리뷰"
version: "0.1.0"
status: draft
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

- 2026-10-02 · v0.1.0 · manager-spec · 최초 작성. 카드 t1422. 측정 원천: 본 트리 `.moai/worktrees/t1422`(브랜치 `WT-codex-review-lane-scope`, HEAD `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`)에서의 코드 좌표 직접 판독(§A). 결정 D1·D2·D3는 `.moai/reports/t1422/jev-decisions.md`(Jev 판정 기록 — 표시 전용 신호, 결정 권한은 레인·리더 소관)와 코디네이터 지시(D2 조건부 구속)를 근거로 하며, progress.md「Decision Log」에 원문을 옮겼다. 인접 SPEC 충돌 사전 검사: SPEC-CODEX-GATE-SCOPE-001(completed)의 REQ-CGS-003 한 조항만 좁힌다 — 아래 Amendments.

- 2026-10-02 · v0.1.0 · manager-spec · 결정 확정 반영(제자리 갱신, 버전 불변). 오픈 결정 Q1-Q8 이 운영자 위임(Jev `jev-1.13.0`)·리더 승인으로 해결되어 REQ-CRO-001(키는 템플릿에 주석 예시로만 배포)·§E 범위 밖(감사 도구 후속 카드 필요, 로컬 primary 설정 비편집, t1404 착지 후 인계)을 맞췄다. 처분 원문은 progress.md「Decision Log」, 해결 표는 plan.md §G.

### Amendments

- 2026-10-02 — 후속 개정(successor amendment). 대상: `SPEC-CODEX-GATE-SCOPE-001` **REQ-CGS-003**("카드 세션이 아닌 세션의 대상은 변하지 않는다"). prior_completed_version: `0.1.0`. prior_completed_sha: `3f4000bbe`(본 트리 `git log -- .moai/specs/SPEC-CODEX-GATE-SCOPE-001/spec.md` 실측 — 그 SPEC의 3-phase close 커밋. 부모 progress.md §E.4 의 `sync_commit_sha` 는 측정 시점에 `pending-backfill-sync` 로 남아 있다). 근거: REQ-CGS-003 은 비카드 세션 전부를 "트리 전체 미커밋 리뷰"로 못 박아, 설정으로 그 리뷰를 끌 길이 없었다(§A.2). 범위: 적용 조건 한 줄만 좁힌다 — REQ-CGS-003 의 "카드 스코프가 아닌 세션"이 **"카드 스코프가 아니고 `tree_scope` 가 `skip` 이 아닌 세션"** 이 된다(REQ-CRO-003). `skip` 이 아닐 때의 요청 형태·fail-open·REQ-CRT-006 회귀선은 그대로이고, REQ-CGS-001·002·004~010 은 건드리지 않는다. 부모 SPEC 본문은 수정하지 않는다(완료 SPEC 본문 불변 — 개정은 이 SPEC이 선언한다).

---

## §0 지배 원칙 [HARD]

> **리뷰는 변경을 귀속할 수 있는 세션이 자기 몫으로 한다. 턴 종료 게이트는 그 귀속이 성립하는 곳에서만 돈다.**

현행은 "누가 무엇을 리뷰하는가"를 턴마다 도는 강제 하나로 풀려 한다. 결과는 두 가지다. 첫째, 귀속할 수 없는 변경(공유 primary 체크아웃의 미커밋 전체)이 리뷰 대상이 되고, 그 판정이 턴을 막는다. 둘째, 리뷰를 **요청할 길**이 없다 — 게이트가 돌 때만 리뷰가 일어난다. 이 SPEC은 소유권을 셋으로 가른다. (1) 카드 리뷰는 레인의 몫이고 카드 diff로 한정한 카드 단계로 수행한다. (2) 리더 자신의 내부 산출물은 리더가 직접 리뷰하며 리더 세션에는 턴 종료 리뷰 게이트를 두지 않는다. (3) 어느 세션이든 자기 카드·자기 코드를 codex나 GLM에 직접 요청할 수 있다. 리뷰 품질·판정 문구·900s 예산·감사(audit) 영수증 의미는 별개 축이며 이 SPEC은 건드리지 않는다.

---

## §A 배경 (측정)

### A.0 측정 원천과 핀

좌표는 모두 본 트리 기준이며 줄 번호는 2026-10-02 판독값이다. 측정 HEAD 핀: `c50da9c2f8aa1227073bd77caa07ca1c75b8d81b`(`git rev-parse HEAD`), 브랜치 `WT-codex-review-lane-scope`. 이 레인이 라이브 레인·라이브 codex를 실행해 관측한 항목은 없다 — 아래 서술은 코드·파일 판독이며, 관측하지 않은 전제는 해당 항목에 Gap으로 표기한다.

### A.1 현행 게이트의 구조 — 비카드 세션을 가르는 축이 없다

- `internal/cli/codex_review_gate.go:69-119` `HandleCodexReviewGate` 의 판정 순서: 비활성(:71) → `stop_hook_active`(:74) → 스코프 해상(:81, 로그 :82) → 스코프 셀프게이트(:83) → codex 조회(:87) → 리뷰(:103) → fail 차단(:112-117). 이 순서의 어디에도 트리 스코프 세션을 설정으로 제외하는 분기가 없다.
- `internal/cli/codex_review_scope.go:84-104` `resolveReviewScope`: `WT-` 접두 브랜치이고 `git merge-base develop HEAD` 가 계산될 때만 카드 스코프(:97-103), 그 외 전부 트리 스코프. 요청 조립은 `:169-180` `reviewRequestParams` — 카드는 `baseBranch` 대상(재계산한 merge-base SHA)+카드 트리 cwd, 트리는 `uncommittedChanges`+해상 트리 cwd.
- Codex Stop 체인 멤버 6(`internal/cli/codex_stop_chain.go:613-669`)과 receipt 생산자(`internal/cli/codex_review_receipt.go:126-174`)는 같은 `reviewScopeResolver` 를 부른다(`:624`, `:134`). `moai verify codex-review` 는 사용자가 명시적으로 돌리는 명령이다.
- multi 리뷰 게이트(`internal/cli/multi_review_gate.go:75`)는 스코프 해상기를 쓰지 않고 `reviewGateChangeDetector(projectDir)` 를 직접 부른다 — codex 리뷰 게이트의 스코프 개념을 공유하지 않는다. 기본 OFF(`workflow.multi.review_gate.enabled: false`).

### A.2 REQ-CGS-003 이 설정 탈출구를 막고 있다

SPEC-CODEX-GATE-SCOPE-001 REQ-CGS-003 은 비카드 세션의 대상을 "해상 트리의 미커밋 전체, 사전 형태와 동일"로 고정했다(리더 세션을 명시적으로 포함). `workflow.codex.review_gate.enabled` 는 **세션 전부에 한 번에** 걸리는 스위치라서, "카드 세션은 리뷰하고 비카드 세션은 건너뛴다"를 표현할 수 없다.

### A.3 어느 설정 루트가 `enabled` 를 정하는가

- 셸 래퍼 `.claude/hooks/moai/handle-codex-review-gate.sh` 는 `CLAUDE_PROJECT_DIR`(없으면 `$PWD`)의 `.moai/config/sections/workflow.yaml` 에서 `workflow.codex.review_gate.enabled` 를 순수 셸로 읽고, off 면 exit 0 한다.
- Go 핸들러는 `runCodexReviewGate`(`codex_review_gate.go:192-211`)에서 `resolveProjectDirFromInput`(`:260-270` — ProjectDir → CWD → `CLAUDE_PROJECT_DIR`)으로 projectDir 를 정하고, `reviewGateConfigRoot`(`:220-226`, `auditreceipt.StoreRoot`)가 돌려준 루트의 `workflow.yaml` 로 `readCodexReviewGateEnabled`(`internal/cli/mcp_codex.go:2385`)를 호출한다. 설정 루트는 `.moai` 를 추적하지 않는 저장소의 연결 워크트리일 때만 primary 로 바뀌고, 그 외에는 projectDir 그 자체다(SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-008).
- Codex 체인 멤버 6은 `readCodexReviewGateEnabled(c.root)`(`codex_stop_chain.go:618`)로 **세션 트리를 그대로** 설정 루트로 쓴다 — `reviewGateConfigRoot` 를 거치지 않는다(기존 비대칭, 이 SPEC은 고치지 않는다).
- 본 트리의 추적된 `.moai/config/sections/workflow.yaml` 에는 `review_gate` 키가 없다(`grep -c review_gate .moai/config/sections/workflow.yaml` → 0, 본 트리). primary 체크아웃 사본에는 `codex.review_gate.enabled: true` 가 있다(`/Users/goos/MoAI/moai-adk-go/.moai/config/sections/workflow.yaml:132-134` 직접 판독). 그 파일이 추적 파일의 로컬 미커밋 수정인지는 워크트리 가드 제약으로 git 으로 확인하지 못했다(리더가 로컬 미커밋으로 전달) — **Gap**.
- 코드 판독에 따른 귀결(라이브 레인 미관측): 카드 워크트리에 cwd 를 둔 레인 세션의 Go 측 `enabled` 는 그 워크트리의 추적 사본에서 읽혀 이 저장소에서는 off 로 읽힌다. 즉 현재 이 저장소에서 게이트가 도는 곳은 `enabled: true` 가 있는 primary 에 앉은 세션(리더 포함)이다.

### A.4 리뷰 도구 보유 현황

`codex_audit`/`glm_audit` 를 `tools:` 에 가진 에이전트는 `plan-auditor`(`.claude/agents/moai/plan-auditor.md:7`)와 `sync-auditor`(`.claude/agents/moai/sync-auditor.md:9`) 둘뿐이다(`grep -l mcp__moai__codex_audit .claude/agents/moai/manager-develop.md …manager-docs.md …manager-lead.md …plan-auditor.md …sync-auditor.md` → `sync-auditor.md`, `plan-auditor.md` 만 출력, 본 트리). manager-develop·manager-docs·manager-lead 의 `tools:` 줄에는 없다(각 `:9`/`:9`/`:10`). 메인 세션은 에이전트 `tools:` 제약을 받지 않으므로 직접 호출할 수 있다고 추정되나 이 레인은 메인 세션의 도구 목록을 관측하지 않았다 — **Gap**.

`codex_audit` 는 매 호출이 감사 영수증을 남긴다(`internal/cli/mcp_codex.go:1931`, `:1954` — `recordAuditReceipt` 호출 2곳, `grep -c recordAuditReceipt internal/cli/mcp_codex.go` → 2). 영수증 가드는 같은 트리에서 감사관 시작 이후에 만들어진 `codex_audit`/`audit_multi` 영수증을 PASS 근거로 인정한다(`internal/auditreceipt/store.go:596-601`). 또 `applyGateUnmet`(`mcp_codex.go:1929`, `:1952` 호출, 정의 `:1974`)은 `workflow.audit.gates.codex: required` 일 때 리뷰 부재를 `fail`+`gate_unmet` 로 바꾼다.

### A.5 `baseBranch` 대상 해상 결함 (b)

- `codex_audit` 의 `baseBranch` 대상은 호출자가 지정할 수 없고 서버가 정한다 — `coerceCodexReviewTarget`(`mcp_codex.go:1218`, `baseBranch` 분기 `:1230-1236`)→`resolveReviewBaseBranchName`(`internal/cli/mcp_review_material.go:131-144`): 원격 기본 헤드, 그다음 `main`. `glm_audit` 는 별개 체인으로 `resolveReviewMergeBase`(`mcp_review_material.go:92-103` 부근: `origin/HEAD`, `origin/main`, `main`)의 `git diff <base>...HEAD`(`reviewDiffArgs` `:73-87`)를 쓴다.
- 게이트의 카드 diff 기저는 이 둘과 다르다 — `git merge-base develop HEAD`(`codex_review_scope.go:132`, 로컬 `develop`, 매 평가 재계산).
- 본 트리 측정: `git symbolic-ref --short refs/remotes/origin/HEAD` → `origin/develop`. 그러므로 **이 저장소에서** codex 쪽은 이름 `develop` 으로 해상되어 "main 과 비교"하지는 않는다. 결함은 구조적이다: 호출자가 기저를 지정할 수 없고, 두 감사 도구의 체인이 서로 다르며, 어느 쪽도 게이트의 판별·재계산 규율(로컬 develop, `WT-` 판별)을 쓰지 않는다. 로컬 `develop` 이 `origin/develop` 보다 앞서 있고(배치 push 전) 카드가 그 커밋을 흡수했다면 GLM 쪽 `origin/HEAD` 기저는 다른 카드의 커밋을 카드 diff 로 끌어들인다(gitflow-lane-protocol §8 이 막는 오염과 같은 형태) — 이 추론은 구조 판독이며 라이브 재현은 하지 않았다 — **Gap**.

### A.6 동기와 출처 표기

카드 본문(리더 제공)은 "리더 Stop 훅 1회당 codex 리뷰 9~12분·Bash 80~105회, primary 282파일 미커밋 diff 전체를 매 턴 재검토"를 기록한다. 본 레인은 이 수치를 독립 재측정하지 않았다 — **리더 제공**. 이 SPEC이 의존하는 것은 수치가 아니라 §A.1-A.3 의 구조(비카드 세션을 설정으로 가르는 축 부재)다. 인접 카드 t1404(게이트 스코핑 구멍)의 항목별 처분은 plan.md §D 에 코드 인용과 함께 둔다.

---

## §B 요구사항 (GEARS)

세 묶음이다. **A. 비카드 세션 정책**(REQ-CRO-001~006), **B. 온디맨드 자기 리뷰 표면**(REQ-CRO-007~011), **C. 보유·교리·정합**(REQ-CRO-012~014).

### A. 비카드 세션 정책 (결정 D1)

### REQ-CRO-001 — 비카드 세션 정책 키 (Ubiquitous)

The codex review gate configuration shall define `workflow.codex.review_gate.tree_scope` with the allowed values `review` and `skip`; an absent key, an empty value, or any other value shall read as `review`, and the comparison shall ignore surrounding whitespace and letter case. The distributed template shall carry the key only as a commented example; the config struct shall parse it, and no shipped-key inventory entry, settings-schema field, console field, or i18n entry shall be added for it.

키 이름과 값은 `workflow.codex.review_gate.tree_scope` = `review`|`skip` 으로 확정됐다(decision-index.md Q2, 운영자 위임 판정). 배포 형태는 템플릿 `workflow.yaml` 의 주석 예시뿐이다(Q1 확정) — 살아 있는 shipped 키가 아니므로 인벤토리·설정 스키마·콘솔·i18n 은 변하지 않는다. `review` 는 현행 동작(트리 스코프 세션의 전체 미커밋 리뷰)이고 배포 기본이다. 알 수 없는 값을 `review` 로 읽는 방향은 "덜 검사하는 쪽으로 조용히 틀어지지 않는" 방향이다.

### REQ-CRO-002 — 트리 스코프 세션의 skip (Where + While)

**Where** the codex review gate is enabled, **While** the session's scope class is tree and `tree_scope` is `skip`, the review gate shall allow the turn without running the self-gate detector, the reviewer lookup, or any review, and shall record the skip and its basis in the gate's structured log.

Claude Stop 훅과 Codex Stop 체인 멤버 6 두 자동 경로 모두에 적용된다(REQ-CRO-006). 로그는 REQ-CGS-010 의 스코프 로그 행과 구별되는 한 행(클래스·키 값·skip 사유)이다.

### REQ-CRO-003 — 트리 스코프 세션의 기존 동작 유지 (While) — REQ-CGS-003 개정

**While** the session's scope class is tree and `tree_scope` is not `skip`, the review gate shall scope the review to the whole uncommitted changes of the resolved tree, shape-identical to its form before SPEC-CODEX-GATE-SCOPE-001 — the same path SPEC-CODEX-REVIEW-TARGET-001 REQ-CRT-006 pins.

**개정 선언(명시).** 이 요구는 SPEC-CODEX-GATE-SCOPE-001 REQ-CGS-003 의 후속 개정이다. 바뀌는 것은 적용 조건 한 줄뿐이다 — "카드 스코프가 아닌 세션" → "카드 스코프가 아니고 `tree_scope` 가 `skip` 이 아닌 세션". 바뀌지 않는 것: `skip` 이 아닐 때의 요청 형태(`uncommittedChanges`+해상 트리 cwd), fail-open(REQ-CGS-008), REQ-CRT-006 회귀선. 이유: REQ-CGS-003 은 비카드 세션의 리뷰를 항상 켜진 것으로 못 박아 설정으로 끌 수 없었다(§A.2). `tree_scope` 가 부재이거나 `review` 이면 REQ-CGS-003 의 이전 문장과 관측상 동일하다.

### REQ-CRO-004 — 카드 스코프는 정책의 영향을 받지 않는다 (While)

**While** the session's scope class is card, the review gate shall review the card diff as REQ-CGS-002 defines it, whatever value `tree_scope` holds.

`tree_scope` 는 "카드가 아닌 세션이 무엇을 하는가"만 정한다. 카드 세션의 게이트 동작은 `enabled` 와 REQ-CGS-002 가 정한 그대로다.

### REQ-CRO-005 — 정책 판정의 입력은 둘뿐이다 (Ubiquitous + shall not)

The tree-scope policy decision shall take exactly two inputs — the resolved scope class and the `tree_scope` value — and shall not be derived from session environment labels, a leader or lane role, or launcher mode.

REQ-CGS-004 의 연장이다. 리더를 env 로 검출하지 않는 근거: 진행 중인 카드 t1399 가 `MOAI_KANBAN*` env 계열을 삭제하고 있고, REQ-CGS-004 가 env 를 스코프 판별 입력에서 이미 배제했다. 리더·레인이라는 역할은 게이트에 입력이 아니며, 역할 차이는 세션이 어느 트리에 앉았는가(스코프 클래스)와 그 저장소의 설정으로만 표현된다.

### REQ-CRO-006 — 두 자동 경로와 설정 루트 (Ubiquitous)

The Claude Stop hook and the Codex Stop chain's codex review member shall apply the tree-scope policy through one shared function, reading `tree_scope` from exactly the config root each path already reads `enabled` from, so that the two paths cannot disagree on the policy for the same session state; the receipt producer `moai verify codex-review` is an explicit request rather than an automatic gate and shall run its review whatever `tree_scope` holds.

설정 루트는 경로마다 `enabled` 와 같다: Claude 훅은 `reviewGateConfigRoot(projectDir)`, Codex 체인은 `c.root`(§A.3). 두 경로의 기존 비대칭은 유지하되 새 키가 `enabled` 와 다른 루트를 읽는 일은 없어야 한다. 명시 실행 `moai verify codex-review` 는 사용자가 요청한 리뷰라 정책으로 끄지 않는다(온디맨드 원칙).

### B. 온디맨드 자기 리뷰 표면 (결정 D2 — 코디네이터 조건부 구속)

### REQ-CRO-007 — 자기 리뷰 도구 두 개 (Ubiquitous)

The moai MCP server shall expose two read-only tools, `codex_review` and `glm_review`, each requiring a `scope` input whose allowed values are `card` and `uncommitted`, and each accepting the `project_root` input under the contract the existing `project_root` tools share.

`scope` 는 필수다 — 기본값이 있으면 레인이 `scope` 를 빠뜨렸을 때 카드의 커밋분을 놓치고 미커밋분만 보는 조용한 다른 리뷰가 된다(SPEC-CODEX-REVIEW-TARGET-001 이 닫은 "다른 변경을 말없이 리뷰하는" 형태의 한 칸 옆). 두 도구는 감사(audit) 도구가 아니다 — `codex_audit`·`glm_audit`·`claude_audit`·`audit_multi` 의 입력·출력·영수증 의미는 바뀌지 않는다.

### REQ-CRO-008 — 카드 스코프 해상은 게이트와 같은 해상기다 (When)

**When** a self-review tool is called with `scope` `card`, the tool shall resolve the scope through the same scope resolver the turn-end gate uses and review the card diff — the changes from the recomputed `git merge-base develop HEAD` to the card branch HEAD plus the uncommitted changes of the card worktree, runtime-managed prefixes excluded — without introducing a second scope discriminator and without pinning the merge base; **when** the named tree does not resolve to card scope, the tool shall return an `inconclusive` result naming the cause, shall not call the reviewer, and shall not review any other change in its place.

이것이 결함 (b)의 수리다 — 호출자가 게이트와 **같은** 판별·같은 재계산 규율로 측정된 카드 diff 를 요청할 수 있다. 판별 입력은 REQ-CGS-004 그대로 `WT-` 브랜치뿐이다. codex 요청은 `reviewRequestParams` 가 만드는 형태(`baseBranch` 대상=재계산 SHA, cwd=카드 트리)를 그대로 쓴다. GLM 자료는 같은 합집합을 `git diff <MergeBase>` 로 잰다(커밋분+추적 파일의 미커밋분). GLM 자료에서 빠지는 비추적 비런타임 경로는 REQ-CRO-010 이 결과에 명시한다.

### REQ-CRO-009 — 미커밋 스코프 (When)

**When** a self-review tool is called with `scope` `uncommitted`, the tool shall review the whole uncommitted changes of the named tree, with a request shape identical to the tree-scope request of the turn-end gate.

리더·일반 세션이 자기 내부 산출물을 리뷰하는 경로다.

### REQ-CRO-010 — 자기 리뷰는 조언이다 (Ubiquitous + shall not)

Every self-review result shall carry a machine-readable `advisory` marker with the value true, together with the scope, the base commit, the backend, and the tree reviewed; a self-review tool shall not mint or consume an audit receipt, shall not apply a `required` audit gate, and shall fail open to `inconclusive` on every missing-reviewer, error, or empty-material path, as SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012 requires; where the GLM material excludes untracked non-runtime paths, the result shall name them.

판정(verdict) 값은 `codex_audit` 와 같은 어휘(`pass`/`fail`/`inconclusive`)를 쓰되 **구속력이 없다** — 도구 설명·출력 스키마·교리 세 곳에 모두 "advisory(non-binding)"로 적는다. 감사 영수증을 만들지도 소비하지도 않으므로 PASS 근거로 인용될 수 없다(`store.go:596-601` 가드의 입력에 들어가지 않는다). `required` 게이트의 `fail`+`gate_unmet` 변환(§A.4)을 받지 않는다 — codex 부재는 언제나 `inconclusive` 다.

### REQ-CRO-011 — CLI 거울 (Ubiquitous)

The `moai` command line shall expose the same two reviews through one command backed by the same implementation as the MCP tools, shall print the same result fields as JSON, and shall exit with status 0 for any review verdict.

MCP 서버가 오래된 빌드로 떠 있거나 에이전트가 MCP 도구를 보유하지 않은 자리, 운영자가 터미널에서 직접 요청하는 자리를 위한 경로다. 종료 코드는 조언성(advisory)을 따라 판정과 무관하게 0 이고, 사용법 오류·사용 불가 루트만 비0 이다.

### C. 보유·교리·정합

### REQ-CRO-012 — 도구 보유 (Ubiquitous + shall not)

The agent definitions of manager-develop, manager-docs, and manager-lead shall list `mcp__moai__codex_review` and `mcp__moai__glm_review` in their tool lists, and `mcp__moai__codex_audit` and `mcp__moai__glm_audit` shall remain listed only by plan-auditor and sync-auditor.

적용 사본: 로컬 `.claude/agents/moai/*.md`, 배포 미러 `internal/template/templates/.claude/agents/moai/*.md`, 기계 방출 `.codex/agents/moai/*.toml`(`make agents-emit`). 메인 세션(리더·레인 오케스트레이터)은 MCP 를 직접 호출한다. 카드 t1424 가 manager-develop 의 `tools:` 줄에 codex·glm 위임 도구를 추가하는 작업을 가진다 — 같은 줄을 편집하므로 병합 순서 위험을 plan.md §G 에 둔다.

### REQ-CRO-013 — 레인의 카드 리뷰 단계와 리더 (Ubiquitous)

The kanban lane doctrine shall name a `card-review` stage in the card's stage list, positioned after run-phase convergence and lane-local verification and before the integration step: the lane shall request a card-scope self-review, write the result to `.moai/reports/<card-id>/card-review.md` with the backend, the base commit, the verdict, the findings, and a disposition for each finding, and cite that path in the card's progress record, while the review remains advisory and never replaces the leader's evidence read or an independent audit; the same doctrine shall state that the leader session carries no turn-end codex review gate and reviews its own internal output directly with the same tools.

단계의 구체형은 plan.md §E 에 둔다. 리뷰 부재(codex 미설치·`inconclusive`)는 PASS 가 아니라 기록된 사유다. 이 단계는 Stop 훅이 아니라 레인의 카드 단계이므로 턴을 막지 않고, 카드당 한 번(재리뷰는 수리 후 한정 횟수) 돈다. 카드 스코프 Stop 게이트는 코드·옵트인 기능으로 남지만 이 저장소의 레인 리뷰 수단은 이 단계다(plan.md §E).

### REQ-CRO-014 — 목록·문서 정합 (Ubiquitous)

The tool catalogue, the server registration, the `project_root` inventory in the rule files and in the four-locale docs-site pages, the console's per-tool i18n entries, and the MCP tool counts in doctrine shall describe the two new tools identically, each verified by the parity test that already guards it.

도구 수 45→47, `project_root` 선언 도구 20→22, 쓰기 가능 도구 집합은 불변(두 도구 모두 읽기 전용). 정합 대상 전수는 plan.md §F.

---

## §C AC 형태에 대한 구속 [HARD]

- AC 는 **게이트·도구가 조립한 요청의 대상 필드**(target·cwd), **설정 값**(`tree_scope` 판독 결과), **도구 목록**(`tools/list` 스키마·`tools:` 줄), **영수증 저장소의 변화**(호출 전후 목록 동일)를 관측한다. verdict 값 단독은 어떤 AC 의 근거도 되지 못한다 — 스텁 codex 는 요청과 무관한 값을 돌려줄 수 있다. SPEC-CODEX-GATE-SCOPE-001 §C 와 같은 규율이다.
- 해상기·정책 판독기는 기존 주입 seam(`reviewScopeResolver`, `reviewGateChangeDetector`, `codexLookPath`, 리뷰 RPC 변수)으로 검증한다 — 라이브 codex·라이브 GLM 의존 AC 는 두지 않는다. skip 은 통과가 아니라 **미관측**이다.
- 카드 픽스처는 실제 git worktree(`WT-` 브랜치, develop 분기 후 커밋 1건+추적 파일 미커밋 1건+비추적 파일 1건) + primary 측 외부 WIP 로 구성한다. 셀렉터 0매칭 초록 방지: RED·GREEN 실행은 `-v` 로 `=== RUN` 을 함께 관측한다(verification-completeness §1.1).

## §D 실행 순서 구속

1. 회귀선을 먼저 초록으로 고정한다: 트리 스코프 `uncommittedChanges` 요청 형태(AC-CGS-002 계열), 카드 스코프 요청 형태, fail-open, 감사 도구 `tools:` 보유 집합, `codex_audit` 의 `required` 게이트 동작. 변경 전 트리에서 초록을 관측한 뒤에 정책·도구를 얹는다.
2. 도구 계약(이름·입력·출력 스키마·advisory 필드)과 설정 키 이름·값을 **먼저** 고정한다 — 둘은 외부 표면이라 되돌리기 가장 어렵다(plan.md §B 순서).

---

## §E 범위 밖 (Out of Scope)

### Out of Scope — 감사(audit) 도구의 대상·영수증 의미

- `codex_audit`·`glm_audit`·`claude_audit`·`audit_multi` 에 `cardDiff` 대상을 추가하는 일, 그들의 `baseBranch` 해상 체인(`resolveReviewBaseBranchName`·`resolveReviewMergeBase`)을 바꾸는 일은 하지 않는다. 결함 (b)는 새 자기 리뷰 도구의 카드 스코프 요청으로만 닫는다(REQ-CRO-008) — **감사 도구 쪽 `baseBranch` 해상은 이 SPEC 이후에도 그대로이며 후속 카드가 필요하다.** 카드 발행은 리더 소관이고 이 레인은 카드를 발행하지 않는다(decision-index Q8 확정).
- 감사 영수증 저장소·가드(`internal/auditreceipt`, `internal/hook/audit_receipt_guard.go`)는 건드리지 않는다.

### Out of Scope — t1404 의 잔여 항목

- sync 게이트 `WCI_EXCLUDES` 의 `.moai/reports/**` 누락(`sync-phase-quality-gate.sh:257-262`), `moai gpt` 문서-CLI 드리프트(카드 t1406), 검토 트리의 develop 대비 진부함 감지, 기지 발견 장부(`ledger.jsonl`) 사전 분류, 비카드 스코프의 런타임 관리 경로 확장, 검토자 CLI 의 빌드 노후는 이 SPEC에 흡수하지 않는다(plan.md §D 의 항목별 처분). 이 SPEC 착지 뒤 리더가 t1404 를 잔여 항목으로 편집하는 일은 sync 단계 인계 항목이며 코드 변경이 아니다(착지 전에는 닫지 않는다, Q4 확정).

### Out of Scope — 카드 스코프 Stop 게이트의 제거·재설계

- 카드 스코프 Stop 게이트(REQ-CGS-001·002·004~010)의 코드와 옵트인 성격은 그대로 둔다. 추적된 `workflow.yaml` 에는 `enabled: true` 를 커밋하지 않는다(Q5 확정). 리더가 primary 의 비추적 로컬 설정에 `tree_scope: skip` 을 적는 일은 운영자 소유 파일에 대한 운영자 행위이며, 이 SPEC 은 그 파일을 편집하지 않는다.
- multi 리뷰 게이트(`HandleMultiReviewGate`)에 스코프 개념을 도입하는 일은 하지 않는다.

### Out of Scope — 역할 검출·env 파싱

- 리더나 레인을 env·프로세스 역할·런처 모드로 검출하는 어떤 메커니즘도 만들지 않는다(REQ-CRO-005). `MOAI_KANBAN*` 삭제는 카드 t1399 소관이다.

### Out of Scope — 리뷰 내용·예산·콘솔

- codex·GLM 이 무엇을 좋다고 보는지, 프롬프트, 900s 예산, 모델 핀 정책은 바꾸지 않는다. `tree_scope` 를 웹 콘솔 필드·`shipped_key_inventory.yaml` 항목으로 노출하는 일은 하지 않는다 — 템플릿에는 주석 예시로만 싣는다(decision-index Q1 확정).
- 레인의 카드 리뷰 단계를 훅으로 기계 강제하는 일, 카드 리뷰 증거 파일의 존재를 검사하는 게이트를 만드는 일은 하지 않는다(교리로만 규정한다).
- Claude 백엔드 자기 리뷰(`claude_review`)는 만들지 않는다.

---

## §F 제약

- **fail-open 불변**(SPEC-MOAI-MCP-SERVER-001 REQ-MCP-012): reviewer 부재·오류·빈 자료·inconclusive 는 게이트와 자기 리뷰 도구 모두에서 ALLOW/`inconclusive` 다. `tree_scope: skip` 은 ALLOW 방향이며 차단 경로를 만들지 않는다.
- **REQ-CRT-006 회귀선**: `skip` 이 아닐 때의 트리 스코프 `uncommittedChanges` 직렬화는 본 SPEC 이후에도 shape-identical 이다.
- **단일 해상기**: 게이트 두 경로와 자기 리뷰 도구는 하나의 스코프 해상기·하나의 요청 조립기를 쓴다. 두 번째 판별기·두 번째 diff 측정식은 금지다(REQ-CRO-008).
- **단순성**(AGENTS.md §5): 기존 해상기·요청 조립기·RPC 드라이버·영수증 형식을 재사용하고 새 의존성을 들이지 않는다. 최소 변경 규모 추정과 3배 점검은 plan.md §H.
- **Template-First**: `.claude/`·`.moai/config/` 의 배포 대상 변경은 `internal/template/templates/` 에 같은 변경으로 들어간다. 배포 미러는 SPEC ID·REQ 토큰·카드 번호·날짜·커밋 SHA 를 담지 않는 중립 본문이다(템플릿 중립성). 에이전트 정의는 C1·C2 를 손으로 고치고 C3(`.codex/agents/moai/*.toml`)는 `make agents-emit` 으로만 만든다. 로컬 전용 파일(`.claude/rules/local/`, `AGENTS.local.md`)은 미러하지 않는다. 목록은 plan.md §F.
- **하드코딩 금지**: `review`/`skip` 값 이름·기본값은 `internal/config` 의 상수·`defaults.go` 에서 단일 원천으로 정의한다.
- **측정 규율**: 검증은 대상 패키지(`./internal/cli/...`, `./internal/mcp/...`, `./internal/config/...`, `./internal/template/...`) 스코프로 돌린다. 전체 스위트 로컬 금지.
- **문서 4개국어**: docs-site 를 건드리면 같은 PR 에서 ko·en·ja·zh 를 함께 고친다.

## §G 참조

- `internal/cli/codex_review_gate.go` · `codex_review_scope.go` · `codex_review_receipt.go` · `codex_stop_chain.go:613-669` — 게이트·스코프·영수증·멤버 6 (§A.1)
- `internal/cli/mcp_server.go:316,341,493` — `codex_audit`/`claude_audit`/`glm_audit` 등록. `internal/cli/mcp_codex.go:1218-1236,1891-1957,1974` — 대상 해상·핸들러·`applyGateUnmet`. `internal/cli/mcp_review_material.go:55-144` — GLM 자료·기저 체인
- `internal/mcp/catalog.go` · `internal/mcp/catalog_test.go:19` — 도구 목록 단일 선언·크기 불변식
- `internal/auditreceipt/store.go:596-601` · `internal/hook/audit_receipt_guard.go` — 영수증 가드
- `.claude/rules/moai/core/moai-mcp-tools.md` · `moai-mcp-tools-catalogue.md` — 도구 지도·`project_root` 규칙
- `.claude/rules/moai/workflow/kanban-dispatch.md` · `kanban-dispatch-detail.md` — 레인 단계·완료는 증거로 읽는다
- `.claude/rules/local/gitflow-lane-protocol.md` §1·§8 — `WT-` 불변식·merge-base 규율
- `.moai/reports/t1422/jev-decisions.md` — 결정 D1~D3 판정 기록(표시 전용 신호)
- 관련 SPEC(전부 completed): SPEC-CODEX-GATE-SCOPE-001(개정 대상) · SPEC-CODEX-REVIEW-TARGET-001 · SPEC-MOAI-MCP-SERVER-001 · SPEC-CODEX-AUDIT-GATE-AXES-001
