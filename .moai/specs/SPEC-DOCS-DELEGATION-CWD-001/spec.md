---
id: SPEC-DOCS-DELEGATION-CWD-001
title: "레인 플로우 스폰 cwd 고정 결함 수리 — 격리 스폰 정착 검증과 싱크 위임 화해 절차"
version: "0.1.1"
status: in-progress
created: 2026-10-01
updated: 2026-10-01
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/template
lifecycle: spec-anchored
tier: M
related_specs: [SPEC-CODEX-GATE-SCOPE-001, SPEC-STALE-RUN-LABEL-001, SPEC-WORKTREE-SWEEP-001, SPEC-BRANCHGUARD-EXEMPT-REACH-001]
tags: "lane, factory, kanban, subagent-cwd, agent-worktree, isolation, sync-commit, manager-docs, reconciliation, t1387"
---

# SPEC-DOCS-DELEGATION-CWD-001 — 레인 플로우 스폰 cwd 고정 결함 수리

카드: **t1387** (리드 발행 2026-10-01 — 카드 본문에 worker-66(t1383) §E.4 소유 예외 기록과 t1373 레인의 양상태 실측이 기재돼 있다)

## §A 배경과 실측 원천

### §A.1 결함 한 줄

서브에이전트의 작업 트리 부착(cwd attachment)은 **런타임이 정하고 레인이 통제할 수 없다** — 같은 세션에서 같은 스폰 방식으로 스폰한 두 전문가가 서로 다르게 착지했다(t1373 실측). 격리되면 자기 트리에 갇혀 카드 트리 쓰기가 가드에 거부되고, 비격리되면 레인 트리에서 동작한다. 싱크 단계의 manager-docs 위임은 **단일 싱크 커밋**(implemented → completed 전이 + progress.md §E.4 `sync_commit_sha:` + 3-phase close)에 닫히는 계약이라, 그 산출물이 에이전트 브랜치에 착지하면 레인의 화해 순서가 3-phase close 계약의 생사를 가른다 — 그 순서를 정한 문서가 오늘 존재하지 않는다.

### §A.2 실측 사례 1 — t1373 레인 (격리 착지)

- `manager-develop` 스폰 → 런타임이 `.claude/worktrees/agent-a703d55100343647b` 로 격리(에이전트가 스스로 브랜치를 `WT-stale-run-gate` 로 개명). 트리 현재 존재 확인: `git worktree list` — `13806938c [WT-stale-run-gate]`.
- 격리 가드가 교차 트리 조작을 전부 거부했다("a worktree-isolated agent's git operations must target its own worktree"). 카드 트리 쓰기 프로브도 "Edit the worktree copy" 로 거부.
- 증거 파일 좌초: 격리 에이전트의 progress 기록(`SPEC-STALE-RUN-LABEL-001/progress.md:137`) — `evidence_file: ".moai/reports/t1373/run-ac-green.md (this worktree; card-tree write refused by the isolation guard)"`. `.moai/reports/` 는 gitignore 라 병합에 동행하지 않는다.
- 레인이 손으로 화해: 미추적 사본 diff 검증 → 대체된 사본 제거 → `git merge --ff-only` → 증거 수확. `WT-stale-run-gate` 는 develop 에 병합됨을 확인(`git merge-base --is-ancestor` — 본 SPEC 작성 시점 재측정).
- 같은 세션, 같은 스폰 방식: `manager-docs` 스폰 → **비격리** → 레인 트리에서 직접 동작. **격리 여부는 레인이 예측할 수 없다.**

### §A.3 실측 사례 2 — worker-66 / t1383 (구조적 2회 실패 → 소유 예외)

`SPEC-CODEX-GATE-SCOPE-001/progress.md:84` §E.4 소유 예외 기록 원문: "sync 커밋은 레인이 직접 수행 — manager-docs 위임이 두 번 구조적으로 실패(스폰 격리가 자체 트리 생성 / 비격리 스폰이 레인 트리 고정 — 서브에이전트 cwd는 스폰 세션 트리에 고정, manager-docs 차단 보고 `feedback_sync_dispatch_tree_anchor.md` 참조)". 레인이 manager-docs의 사전 점검과 초안을 그대로 실행했고 소유 예외를 완료 보고에 기록했다. 격리 에이전트 트리 `agent-af69b18e474caa5e2`(브랜치 `WT-gate-scope-impl`, tip `3f4000bbe` — sync 산출물 커밋 포함)도 존재 확인. 단, 차단 보고 메모리 파일 `feedback_sync_dispatch_tree_anchor.md` 는 양쪽 메모리 저장소(활성 프로필·휴면 저장소) 모두에서 검색 불가 — **본 SPEC 은 그 파일 내용을 인용하지 않는다**(§G 미검증).

### §A.4 트리거 통제 가능성 — 부정 판정 (근거 포함)

1. **에이전트 정의에 isolation 선언이 없다.** `manager-develop`·`manager-docs`·`manager-spec` 세 정의 모두 프론트매터에 `isolation:` 필드가 없다(본 트리 `.claude/agents/moai/*.md` 머리 30행 직접 판독). `isolation: worktree` 는 문서화된 유일한 격리 선택 채널(`worktree-integration.md` § `isolation: worktree` in Agent Frontmatter)인데 셋 다 미선언이므로, t1373 의 격리는 프론트매터가 아니라 런타임 층의 결정이다.
2. **MoAI 는 `background:` 를 설정하지 않는다.** `agent-common-protocol.md` § Background Agent Execution — "MoAI takes the default and does not set the `background:` frontmatter field". 런타임이 골라야 한다.
3. **자동 격리의 발동 조건을 서술한 문서가 없다.** `kanban-dispatch-mechanics.md` 전체에서 isolation/agent-worktree 언급 0건(grep 실측), `worktree-integration.md` 는 opt-in 플래그만 서술. 런타임 자동 격리의 트리거(크기·쓰기 가능성·캐시 상태·무작위성 여부)는 **레인이 읽을 수 있는 문서가 없고, 레인이 넘길 수 있는 매개변수도 없다**.
4. 결론: **결정론적 격리는 레인 측에서 달성 불가** — 수리 부류는 스폰 매개변수가 아니라 "양쪽 착지를 모두 다루는 절차 + 문서 + 가드"가 된다(카드 연구 질문 1-2의 답).

### §A.5 가드 경계 (착지 거부 형상)

`worktree-integration-ops.md` § Refused Commands — 교차 트리 `-C`/`--git-dir` 거부문 7종 관측, 분류기 앵커 `isolated in the worktree`. 반대 방향: **비격리 배경 서브에이전트는 자기 앵커가 없다** — 호출마다 세션의 현재 앵커로 cwd 를 재해석(t741 실측: 경로 없는 호출 12회, 부모 이동 2회 동안 cwd 가 따라 움직이며 거부 0). 그래서 비격리 전문가가 살아 있는 동안 레인 세션이 트리를 옮기면, 전문가의 후반 작업이 **말없이** 다른 트리에 착지한다(무음 재앵커링 위험).

### §A.6 계약 표면 — 단일 싱크 커밋

`spec-frontmatter-schema.md` § Status Transition Ownership Matrix — manager-docs 는 **단일 싱크 커밋**에서 `in-progress → implemented → completed` 전이 + progress.md §E.4 (`sync_commit_sha:`) + 3-phase close 를 수행하고, 커밋 주부는 정확히 하나의 풀 SPEC-ID 를 명명한다. §E.4 는 era.go 가 문자열 매칭하는 파서 적재 표면이다. 화해 순서(에이전트 싱크 커밋의 병합이 레인 측 후속 커밋보다 앞선다)가 어기면 close 계약과 드리프트 판정이 깨진다.

## §B 요구사항 (GEARS)

### REQ-DSC-001 — 부착은 런타임 결정이다 (Ubiquitous)

The lane flow shall treat a phase-specialist spawn's working-tree attachment as a runtime decision the lane neither controls nor predicts — the lane shall not assume isolation or non-isolation for any specialist spawn.

근거: §A.4 부정 판정. t1373 레인의 양상태 실측(§A.2)이 증명한 실패가 정확히 "어느 쪽이 올지 가정한 스폰"이다.

### REQ-DSC-002 — 스폰 뒤 착지 트리를 검증한다 (When)

**When** a lane spawns a write-capable phase specialist (manager-spec / manager-develop / manager-docs / 지정 감사자) whose artifacts or evidence must land in the lane tree, the lane shall verify where the work actually landed — the lane tree or an agent worktree (`git worktree list` 의 `agent-*` 항목, 에이전트 보고의 `git rev-parse --show-toplevel`) — before advancing the card stage.

근거: t1373 격리 발각이 전부 사후 손검증이었다. 검증 없는 진행은 §A.2 의 재현이다.

### REQ-DSC-003 — 격리 착지는 레인이 화해한다 (When, detected)

**When** the specialist landed in an agent worktree, the lane shall reconcile the work into the lane tree itself using guard-safe forms — plain git inside the lane tree (`git merge --ff-only <agent-branch>`), plain-file evidence harvest — and shall not instruct the isolated agent to write or run git across trees.

근거: 가드는 기능이다(카드 제약). 거부 형상 7종(§A.5)을 우회 지시로 되풀이하는 게 t1383 의 실패 형상이다.

### REQ-DSC-004 — 에이전트 브랜치는 병합 전에 WT- 를 얻는다 (While)

**While** reconciling an isolated spawn, the lane shall ensure the agent branch carries the `WT-<slug>` prefix before merging — renaming it when the agent did not — so the branch stays reachable by the `WT-` sweep and disposal policy after remote landing.

근거: t1373·t1383 의 에이전트가 이미 스스로 개명한 선례(`WT-stale-run-gate`, `WT-gate-scope-impl`). 미개명 에이전트 브랜치(`worktree-agent-*`)는 sweep 후보에서 보이지 않는다(`worktree-integration.md` § The rename is also a disposal-path switch).

### REQ-DSC-005 — 싱크 위임의 화해는 단일 싱크 커밋을 그대로 옮긴다 (When, specialist is manager-docs)

**When** the isolated specialist is manager-docs, the lane shall merge the agent's single sync commit as-is and shall not add its own edits to `spec.md` / `plan.md` / `acceptance.md` bodies — the sync commit carries the `implemented → completed` transition, progress.md §E.4 with `sync_commit_sha:`, the 3-phase close, and exactly one full SPEC-ID in its subject.

근거: §A.6 계약 표면. 레인이 싱크 커밋을 "고쳐" 병합하면 소유 매트릭스와 드리프트 close-주부 규약을 동시에 깬다.

### REQ-DSC-006 — run→sync 인접성은 화해가 지킨다 (While, reconciling docs)

**While** reconciling a docs-phase isolated spawn, the lane shall merge the agent's commits before any lane-side follow-up commit (SHA 백필 등) and shall not interleave a lane commit between the specialist's run and sync commits.

근거: `sync_commit_sha` 백필 인접성(§A.6)과 커밋 그래프 순서 증거 규율(`verification-claim-integrity.md` §2.3 — 순서의 유일한 목격자는 커밋 그래프다).

### REQ-DSC-007 — 좌초 증거는 처분 전에 수확한다 (When, evidence stranded)

**When** the specialist's evidence lives under `<agent-tree>/.moai/reports/`, the lane shall copy it into the lane tree's `.moai/reports/` before the agent tree is disposed — gitignored evidence does not travel with the merge — and the disposal follows the hoist-before-dispose rule of `worktree-integration.md` § Hoist.

근거: t1373 좌초 실측(§A.2). 증거가 에이전트 트리의 유일본인 채 트리가 사라지면 검증 주장의 근거가 소멸한다.

### REQ-DSC-008 — 레인은 위임 우회로 산출물을 쓰지 않는다 (Ubiquitous + shall not)

The lane shall not execute a phase specialist's SPEC-artifact edits as a flow — the reconciliation protocol grants no editing authority over `spec.md` / `plan.md` / `acceptance.md` bodies. **Where** a docs delegation has failed structurally twice with refusal evidence, the lane shall restrict its substitute execution to the specialist's own permitted surface — sync outputs (CHANGELOG, progress.md §E.4, spec.md frontmatter `status`/`updated` transition), the single sync commit included — and shall record the ownership exception in progress.md §E.4, name it in the completion report, and report it to the leader.

근거: standing spawn authority 가 닫는 결함이 정확히 "전문가를 안 띄우고 레인이 직접 편집"이었다(kanban-dispatch-detail.md § Factory in-lane 3-stage — tk8hce 관측). t1383 예외는 흐름이 아니라 **트리거·기록 의무가 붙는 최후 수단**으로 성문화한다(카드 제약: 소유 매트릭스 유지).

### REQ-DSC-009 — 비격리 전문가 살아 있는 동안 레인은 트리를 옮기지 않는다 (While)

**While** a non-isolated specialist is live, the lane shall not move its own session's tree — the specialist re-resolves its cwd against the session's current anchor at each call, so a mid-run move silently re-anchors the specialist's later writes.

근거: §A.5 t741 무음 재앵커링 실측. 거부 기록도 남지 않는 방향의 실패다.

### REQ-DSC-010 — 절차 문서가 존재한다 (Ubiquitous, docs)

The lane-flow documentation shall carry the detection-verification duty and the reconciliation procedure — `kanban-dispatch-mechanics.md` § Isolation 아래 화해 절차 소절과 `agent-common-protocol.md` § User Interaction Boundary 의 lane-session 단락에 넣는 항상 적재 의무 문장(필수 단계: 착지 검증, ff-only 병합, 증거 수확, 교차 트리 지시 금지, 인접성) — and the always-loaded clause shall carry the duty with a pointer to the procedure body.

근거: 오늘 그 문서가 0건(§A.4-3 실측)이기에 t1383 이 재발했다.

### REQ-DSC-011 — 문서 절차에 기계 가드가 있다 (Where, repository carries the check)

**Where** the repository's Go test surface runs, a mechanical check shall fail when the reconciliation procedure section or any of its mandatory steps disappears from the rule documentation — the check verifies section presence and the mandatory-step markers in `kanban-dispatch-mechanics.md` and the duty sentence in `agent-common-protocol.md`.

근거: 문서-only 수리는 조용히 삭제될 수 있다(룰 다이어트·스플릿 사례 다수). 콘텐츠 스캔 가드 선례: `internal/template/` 의 룰 콘텐츠 테스트들(`workflow_rule_paths_pinned_test.go` 등).

## §C 인수 기준

Tier M — `acceptance.md` §D AC 매트릭스. AC-DSC-001..008 + 003a, 전 항 기계 검증 가능(단일 명령 + 종료 코드 + RED-now 핀).

## §D 제약

- 소유 매트릭스 불변(카드 [HARD]): 이 SPEC 이 추가하는 권한은 레인의 **병합·수확·기록**뿐이다. `spec.md` / `plan.md` / `acceptance.md` 본문 편집 권한은 없다 — 보호면은 소유 매트릭스의 forbidden crossings 표면이고, sync 산출물(CHANGELOG·progress.md §E.4·status 전이)만이 REQ-DSC-008 예외의 유일 허용면이다.
- 가드 우회 금지(카드 [HARD]): 절차가 가르치는 형상은 레인 트리 안의 plain git 과 plain-file 복사뿐이다.
- 룰 저작 예산: `agent-common-protocol.md` 는 항상 적재 표면 — 추가 문장은 1,000 바이트 이내로 유지하고(`rule-authoring.md` § Threshold calibration) 초과 시 변경 설명에 비용 산언을 붙인다. 절차 본문은 paths 스코프 파일(`kanban-dispatch-mechanics.md`)에 둔다.
- 템플릿 중립성(카드 [HARD]): 룰 편집은 C1 로컬 + C2 템플릿 미러 쌍으로 반영하고 `TestRuleTemplateMirrorDrift` 를 통과시킨다. 카드 내력(t1387, SPEC-ID 인용은 룰 본문에서 금지)은 미러에 넣지 않는다.
- 본 SPEC 은 런타임(Claude Code 바이너리)을 바꾸지 않는다 — `worktree-integration-ops.md` 가 "per-agent anchor registration" 을 바이너리 변경으로 기각한 것과 같은 이유다.

## §E 인접 SPEC 과의 비중복

- **SPEC-CODEX-GATE-SCOPE-001**(completed, card t1383): review 게이트의 **대상 스코핑** — §E Out of Scope 직접 확인. 레인 플로우·화해 절차·문서는 다루지 않는다. 이 SPEC 은 그 카드의 §E.4 소유 예외 기록을 **계기**로만 인용하고 그 스코프에 들어가지 않는다.
- **SPEC-STALE-RUN-LABEL-001**(card t1373): stale-run 라벨 결함 자체는 그 카드 소관. 이 SPEC 은 같은 세션의 **스폰 격리 양상태 실측**만 원천으로 인용한다.
- **SPEC-WORKTREE-SWEEP-001**: sweep 의 안전 술어와 WT- 처분은 기존 정책을 그대로 인용 — 새 처분 정책을 만들지 않는다.

## §F 범위 밖 (Out of Scope)

### Out of Scope — 런타임 자동 격리 트리거의 역공학

- Claude Code 바이너리가 격리를 결정하는 내부 조건(크기·캐시·무작위성)을 규명하거나 문서화하지 않는다. 레인이 통제할 수 없다는 **부정 판정**(§A.4)만으로 수리 부류가 정해지고, 트리거 지식은 수리에 입력이 아니다.

### Out of Scope — 스폰 매개변수·설정 추가

- "비격리 고정"이나 "목표 트리 고정" 같은 스폰 파라미터, 설정 키, 에이전트 프론트매터 필드 추가는 만들지 않는다. 문서화된 선택 채널(`isolation: worktree` 옵트인) 외의 채널은 런타임이 소유한다.

### Out of Scope — manager-docs 위임의 흐름화(레인 직접 실행)

- 레인이 싱크 단계 산출물을 직접 쓰는 흐름은 만들지 않는다. standing spawn authority 와 tk8hce 결함이 닫는 방향의 회귀다(REQ-DSC-008). t1383 예외는 트리거·기록 의무가 붙는 최후 수단으로만 성문화된다.

### Out of Scope — sweep·hoist·가드 코드의 변경

- `moai worktree sweep` / `hoist` 와 worktree 가드의 코드는 건드리지 않는다. 기존 동사가 화해 절차의 재료다(REQ-DSC-007). 가드 거부 형상 카탈로그도 유지한다.

### Out of Scope — 격리 에이전트 트리의 자동 처분

- 미원격 착지 에이전트 트리의 자동 폐기는 만들지 않는다 — 미푸시 브랜치의 트리는 작업의 유일본이다(`worktree-integration.md` § An unpushed worktree branch).

## §G 미검증 (Gaps)

- 차단 보고 메모리 파일 `feedback_sync_dispatch_tree_anchor.md` — t1383 §E.4 기록이 인용하는 파일이 활성 프로필 저장소(`~/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/`)와 휴면 저장소(`~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`) 모두에서 검색 불가(2026-10-01 find/grep 실측). 본 SPEC 은 그 내용을 인용하지 않고, t1383 위험만 progress.md 원문으로 인용한다. run/sync 단계에서 발견되면 §A.3 을 보강한다.
- 런타임 자동 격리의 발동 조건 — 관측 불가(§A.4). "무작위성 여부"를 포함한 내부 기준은 부정도 입증도 안 된 미측정 상태로 남긴다(부재 ≠ 측정됨).
- t1383 의 비격리 실패 두 번째 형상("비격리 스폰이 레인 트리 고정"이 왜 manager-docs 위임 실패로 이어졌는지의 세부)은 §E.4 한 줄 기록으로만 존재하고 독립 재현은 안 된다 — REQ-DSC-002(착지 검증)가 두 형상을 모두 조기 발각하므로 수리는 두 형상에 정합이다.

## §H 이력 (HISTORY)

- 2026-10-01 · v0.1.0 · manager-spec · 최초 작성. 카드 t1387. 측정 원천: 본 트리(`.moai/worktrees/t1387`, 브랜치 `WT-docs-delegation-cwd`, HEAD `e927266be` 기점)에서 `.claude/agents/moai/{manager-develop,manager-docs,manager-spec}.md` 프론트매터 직접 판독, `SPEC-STALE-RUN-LABEL-001/progress.md:137` 좌초 증거 판독, `SPEC-CODEX-GATE-SCOPE-001/progress.md:84` §E.4 예외 기록 판독, `git worktree list` 로 두 격리 트리 존재 확인, `git merge-base --is-ancestor` 로 양 에이전트 브랜치의 develop 병합 확인, 룰 문서 격리 트리거 기재 0건 grep 실측. 차단 보고 메모리 파일 미발견 — §G 기재.
- 2026-10-01 · v0.1.1 · manager-spec · plan-audit r1(FAIL 0.875 — MP-8 가림) 델타 수리. **AC-DSC-004 처분: (a) 유도 변이 RED — 단, 기제 교체와 함께.** 원 인용 `TestRuleTemplateMirrorDrift` 는 프로브 1(C1 mechanics.md description 변이)에서 GREEN 유지가 관측돼(`ok 0.360s`) 이 쌍에 도달하지 않는다는 기계적 증명 확보 — mechanics 는 어느 바이트 패리티 목록에도 미등록이고 `agent-common-protocol.md` 는 §25 정화 쌍으로 바이트 패리티에서 설계상 제외(`rule_template_mirror_test.go` 코멘트). 교체 기제: `TestTemplateNoInternalContentLeak`(템플릿 루트 전역 WalkDir) + `TestSanitizedPairParity`(agent-common-protocol 등록). 프로브 2(C2 mechanics 에 SPEC-ID 주입)로 RED 관측 — `FAIL github.com/modu-ai/moai-adk/internal/template`, 위반 라인 `class=C1-spec-id-prefix | match=SPEC-DOCS-DELEGATION-CWD-001`, 복원 후 sha256 전후 일치(`4e04d962…94fa`). AC-005/006/007/008 RED-now 셀 추가(본 인 재측정치 — 감사자 지시대로 보고 수치 이월 없음). D3(REQ-008 shall형 재서술 + 예외 범위 명명)·D4(§D 보호면 명명)·D6(AC 범위 표기) 접수. D2(decision-index Q1/Q2)는 SPEC 편집 불가 — kick-off 전 운영자 직답 유지.
