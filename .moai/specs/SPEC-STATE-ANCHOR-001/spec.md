---
id: SPEC-STATE-ANCHOR-001
title: "상태 앵커 단일 시접 — cwd 오염 수리(GH #1694) + 홈 오염 정지 검증"
version: "0.1.0"
status: draft
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/statusline, internal/config, internal/cli, internal/kanban"
lifecycle: spec-anchored
tags: "state-anchor, statusline, cwd-pollution, gh-1694, home-pollution, t510"
tier: M
era: V3R6
related_specs: [SPEC-SESSION-TELEMETRY-001]
issue_number: 1694
---

# SPEC: 상태 앵커 단일 시접 — cwd 오염 수리(GH #1694) + 홈 오염 정지 검증

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-07 | manager-spec | 최초 작성 — 카드 t510 plan-phase. 판정서 `.moai/reports/t510/verdict.md`의 수리 방향을 SPEC으로 전사. 축 B 수리(단일 상태-앵커 시접) + 축 A 검증 전용 AC + 홈 폴백 가드를 운영자 미결 결정으로 표면화 |

카드: **t510** (Class B · Tier M · 외부 재현 GH #1694). 측정 기준 트리: `.claude/worktrees/t510`, 브랜치 `WT-state-write-locus` @ `0b1e27877` (= origin/develop, 세션 시작 시 fetch로 재확인). 모든 코드 좌표와 RED 근거는 이 트리의 것이다.

## 1. 문제 — 측정된 형태

moai의 런타임 상태 파일이 **의도하지 않은 자리에 쓰인다**. 두 축이 있고, 둘은 다른 리졸버의 다른 결함이다(판정서 Claim 1 — 관측이 만나는 지점 없음, 재론하지 않는다).

### 1.1 축 B — cwd 오염 (사용자 영향, GH #1694)

외부 제보(moai-adk 3.1.2, macOS, 프로젝트 루트에 정상 `.moai/` 보유): "그때 만지고 있던 디렉터에 쓴다. 프로젝트 안에 stray `.moai` 226개(내용 있는 것 221)". 판정서 B-2의 HEAD 격리 재현이 이 기전을 직접 확인했다 — statusline이 세션 텔레메트리를 stdin `workspace.current_dir`에 기록한다. 세션이 cd한 디렉터마다 렌더당 1개의 레코드가 착지하므로, cd를 많이 하는 에이전트 세션에서 제보자가 본 226개가 그대로 설명된다.

생산 cwd/env-앵커 상태-쓰기 패밀리 (판정서 B-1 표 전사 — 이 표가 본 SPEC의 멤버 경계다):

| # | 쓰기/읽기 표면 | 좌표 | 현재 앵커 소스 | 착지 위치 |
|---|---|---|---|---|
| B1 | 세션 텔레메트리 쓰기 | `internal/statusline/context_usage.go:176` `writeContextUsage` / `:278` `resolveProjectDir` (호출 `builder.go:178`) | stdin `workspace.current_dir` → `input.CWD` → `os.Getwd()` — **`workspace.project_dir`은 후보에 없음** (`types.go:184`에 필드 존재), git 해석 없음 | `<앵커>/.moai/state/context-usage/<sid>.json` 매 렌더 |
| B2 | landed 카운트 | `internal/statusline/landed.go:78` `landedCachePath` / `:118` `maybeRefreshLandedCounts` (호출 `builder.go:274` → `backlog.go:24` `resolveBoardRoot`) | `worktree.original_cwd` 있으면 그것 — **없으면 B1과 동일 current_dir 사슬** | `<앵커>/.moai/state/landed/counts.json` |
| B3 | goal 상태 **읽기** | `internal/statusline/builder.go:286` | B1과 동일 리졸버 | 읽기: `<앵커>/.moai/state/goal/...` — cd한 세션이 프로젝트 루트가 아닌 곳에서 읽음 |
| B4 | config 캐시 | `internal/config/cache.go:58` `cacheFilePath(<configDir>/state/config-cache.json)` ← `manager.go:74` `LoadWithCache` | CLI 사슬의 configDir(코드 전반에서 env → Getwd 패턴; **정확한 유도 지점은 미확정 — run-phase RED 테스트가 고정할 것**, 판정서 Gaps) | `<configDir>/state/config-cache.json` |

경계 밖 멤버: **B5**(훅 계열, `internal/hook/path_resolve.go:66` 외 — 훅 맥락에선 env 우선이라 저위험)와 **B6**(세션 레지스트리, `internal/session/registry.go:174` — cwd는 엔트리 필드일 뿐 쓰기 앵커가 아님)은 판정서가 read-only-noted로 분류했고 본 SPEC이 건드리지 않는다(§6).

### 1.2 축 A — 홈 오염 (이미 수리돼 있다 — 검증만 남는다)

`~/.moai/todo/001-<hash>` 형태의 디렉터 341개가 홈에 쌓였다. 범인은 `internal/cli` todo 테스트 3형제이고, 커밋 `e7a078970`(2026-09-02, card t422 fail-loud guard)이 3건을 수리 + 가드를 세웠다. 판정서 A-6이 "고쳐짐"의 요건(활발한 트리거 + 부재)을 충족시켰다: Sep 3~7 동안 full `internal/cli` 스위트가 여러 차례 실행됐는데(t498·t500·t503 종결 검증 포함) 그 기간 신규 `001-*` 디렉터 0건. 가드 코드도 HEAD에 존재한다(`internal/cli/todo_test.go:42` `runTodo` 게이트).

**본 SPEC은 축 A에 코드 수리를 하지 않는다.** 하는 일은 검증 AC — (a) 가드 활성 실측, (b) canary-HOME 스윕 0 오염, (c) 가드가 실제로 무엇을 잡는지의 뮤턴트 증명 — 뿐이다. 0-오염 관측만으로는 부재 가드가 조용히 통과할 수 있으므로(§3), 뮤턴트가 유일한 판별 증거다.

### 1.3 잔여 리스크 — 수리로 사라지지 않는 두 조각

1. **t422 가드의 표면은 2 헬퍼뿐** — `runTodo`/`runTodoWithClosedStdin`. `newTodoCmd()` 직접 `Execute` 우회 시 무가드(t422 판정이 이미 지적). 새 테스트 헬퍼 추가 시 재발 가능하다(REQ-SA-011).
2. **생산 홈 폴백이 설계로 살아있다** — `internal/kanban/todo_root.go`는 git 해석 실패 시 `~/.moai/todo/<key>`로 fail-open한다(함수 주석이 문서화된 의도). 비git 임시 디렉터에서 생산 바이너리로 `moai todo`를 호출하면 지금도 홈에 큐가 생긴다(판정서 A-4 표본: `proj-*`, `t203-probe-*` 2건). 이 설계를 임시 디렉터에서 가드할지는 **운영자 미결 결정**이다(§5).

## 2. 원인 — 시접이 하나 없어서 4개 멤버가 각자 앵커를 고른다

`resolveProjectDir`(`context_usage.go:278`)의 사슬은 `current_dir` → `input.CWD` → `os.Getwd()`다. `workspace.project_dir` 필드가 `types.go:184`에 이미 존재하면서도 후보에 없고, git 루트 해석도 없다. B2(`backlog.go:24`)는 워크트리 세션만 바로잡는다 — `worktree.original_cwd`가 있으면 그것을 쓰지만(이 리포의 실사용 워크트리 환경이라 대부분 올바르게 동작), 없으면 B1과 같은 current_dir 사슬로 떨어진다. B3는 읽는 쪽이라 오염 대신 **가시성 결함**이다 — cd한 세션이 goal 상태를 못 본다. B4는 별도의 CLI 사슬이지만 같은 질문(앵커가 어디냐)에 cwd로 답한다.

같은 리포 안에 이미 정답 형태가 두 개 있다(판정서 대조군):

- `internal/statusline/memory.go:73` `readLLMYAMLContextWindows` — 현재 위치에서 조상 방향 워크업, read-only, 실패 시 폴백.
- `internal/kanban/todo_root.go:96` `primaryCheckoutRoot` — `gitcore.ResolveGitDirs`(`internal/core/git/checkout.go:56`)의 common dir 부모로 **모든 워크트리/체크아웃에서 하나의 루트**를 답하는 기존 시접.

수리는 새 메커니즘의 발명이 아니라, 이 둘이 이미 보여준 형태를 **하나의 상태-앵커 리졸버**로 통일하는 것이다.

## 3. 대표 mutant — 이 SPEC의 AC가 어떻게 만족될 수 있는가

"렌더가 깨지지 않는다"만 요구하면 다음 구현들이 전부 통과하면서 결함을 남긴다.

1. **부분 수리** — B1만 고치고 B2/B3/B4를 남긴다. 4멤버 각각의 행동 AC(AC-SA-001..004)가 멤버별로 재므로 탐지된다.
2. **표시에까지 앵커를 퍼뜨리는 수리** — `project_dir`을 상태 앵커와 표시 이름에 동시에 적용하면 statusline 표시가 바뀐다. 읽기·쓰기·표시의 관심 분리(판정서 Residual-risk 3)가 요구사항이며 AC-SA-006이 잡는다.
3. **throttle 제거** — 앵커 수리를 하며 write-if-changed skip을 잃으면 렌더당 디스크 쓰기가 돌아온다. AC-SA-007.
4. **skip 경로가 소란을 피우는 것** — 무프로젝트에서 쓰기 생략이 오류 로그나 렌더 실패로 이어지면 "no project, no state"이 아니다. AC-SA-005가 "정상 완료"까지 재며.
5. **부재 가드의 공허 통과** — canary 스윕 0-오염은 셀렉터가 0개 테스트를 돌려도 초록이다. swept count 명시(AC-SA-010) + 뮤턴트 오염 관측(AC-SA-011)으로 닫는다.

## 4. 요구사항 (GEARS)

### 수리 — 축 B (R1: 단일 상태-앵커 시접)

- **REQ-SA-001** (Ubiquitous) — The statusline and CLI state surfaces shall resolve the **state anchor** — the project root under which `.moai/state/` is read and written — through a single shared resolver; no member (B1 telemetry write, B2 board root, B3 goal read, B4 config cache) shall derive its own anchor from the session's current directory.
- **REQ-SA-002** (Ubiquitous) — The shared state-anchor resolver shall resolve the anchor in this fixed precedence: stdin `workspace.project_dir`, then `worktree.original_cwd`, then a git resolution of the anchor directory reusing `gitcore.ResolveGitDirs` (the git common directory's parent — one root for every checkout and worktree of the repository, the `primaryCheckoutRoot` shape).
- **REQ-SA-003** (Event-driven) — **When** the anchor cannot be resolved (no `project_dir`, no `original_cwd`, and the directory is not inside a git repository), the writer shall skip the state write silently — no project, no state — and the render shall complete normally.
- **REQ-SA-006** (Event-driven) — **When** the statusline reads the armed-goal state (B3), it shall read it from the anchored root, not from the session's current directory — a session that has cd'd elsewhere still sees the project's goal state.
- **REQ-SA-007** (Unwanted) — The statusline shall not create `.moai` state directories under directories merely visited by the session (the GH #1694 mechanism): one telemetry record per render shall land only under the anchored root.

### 보존 — 수리가 파괴해서는 안 되는 것

- **REQ-SA-004** (Unwanted) — The repair shall not change what the statusline displays: the display-name derivation (basename for rendering) shall remain sourced from the session's current directory, kept separate from the state anchor.
- **REQ-SA-005** (Unwanted) — The repair shall preserve the write-if-changed throttle and the best-effort silent-failure semantics of every repaired writer (REQ-THRESHOLD-009/REQ-THRESHOLD-012 semantics preserved).
- **REQ-SA-008** (Unwanted) — The repair shall not modify B5 (hook-family state: `internal/hook/path_resolve.go`, `file_changed.go`) or B6 (session registry: `internal/session/registry.go`); the verdict classified both read-only-noted, and a change is permitted only after a plan amendment demonstrates R1's seam cannot hold without it.

### 검증 — 축 A (R2: 코드 수리 없음, AC만)

- **REQ-SA-009** (Ubiquitous) — The Axis-A fail-loud guard (card t422, commit `e7a078970`: `runTodo` refusing a live repository root) shall remain active, and the `internal/cli` todo test family shall run only through guarded helpers.
- **REQ-SA-010** (Ubiquitous) — A canary-HOME sweep of the `internal/cli` todo test family shall create zero directories under the canary `HOME/.moai/todo`, with the swept test count reported explicitly — an empty swept set asserts nothing.
- **REQ-SA-011** (Event-detected) — **When** the Axis-A guard is bypassed (a test helper executing `newTodoCmd()` directly), the canary sweep shall observe the resulting pollution as mutant evidence that the guard — not coincidence — is what keeps the sweep at zero; a mutant that produces no pollution shall be reported alongside the one that does, naming the guard boundary it reveals.

## 5. 미결 설계 결정 — 운영자 소관 (본 SPEC에서 구현하지 않는다)

생산 경로의 홈 폴백(`internal/kanban/todo_root.go` fail-open → `~/.moai/todo/<key>/`)을 **비git/임시-디렉터 기원에 대해 거부할지**는 결정하지 않는다. 근거 상황: 판정서 A-4의 표본 2건(`proj-325ca0b6`, `t203-probe-d7a16ea2`)은 비git 임시 디렉터에서 생산 바이너리로 `moai todo`를 호출해 만들어졌고, 그 경로는 지금도 살아있다.

- **거부하는 안** — 임시-디렉터 기원(TempDir/`/tmp` 판별)에서는 홈 큐를 만들지 않고 안내만: 홈 오염이 구조적으로 끊긴다. 대가: fail-open 설계가 문서화한 "no git metadata still gets exactly one queue" 가용성이 임시 디렉터에서 사라진다.
- **유지하는 안** — 의도 설계 존중, 아무것도 바꾸지 않음: 홈 오염은 계속 가능하지만(단발 호출 시) 테스트 경로의 오염은 t422 가드가 이미 막았다.

이 결정은 본 SPEC의 어느 REQ에도 영향을 주지 않는다(축 B 수리와 축 A 검증은 이 폴백과 독립적이다). `plan.md` §D11이 구속 조건으로 고정한다 — 본 SPEC의 run-phase는 이 결정을 기다리지 않는다. 본 결정의 구현은 별도 후속 카드로 발행하는 것이 적절하다.

## 6. 범위 밖 (Non-goals)

### Out of Scope — 생산 홈 폴백의 임시-디렉터 가드

- `internal/kanban/todo_root.go`의 fail-open 홈 폴백에 대한 임시-디렉터 거부(§5)는 본 SPEC이 구현하지 않는다 — 운영자 결정 대기. 본 SPEC의 `internal/kanban` 접촉은 이 파일의 **코드 변경 없는** 시접 참조(REQ-SA-002의 재사용)뿐이다.

### Out of Scope — B5(훅 계열)와 B6(세션 레지스트리)

- 판정서 B-1이 B5를 저위험(훅 맥락에선 env 우선, `cwd_fallback:true` 로그 존재), B6을 쓰기-앵커 아님으로 분류했다. 본 SPEC은 이 두 멤버의 코드를 변경하지 않는다(REQ-SA-008, AC-SA-008).

### Out of Scope — 기존 오염 디렉터의 정리

- `~/.moai/todo/` 아래 오염 디렉터 343개(`001-*` 341 + `proj-*`/`t203-probe-*` 각 1)는 **증거로 보존**한다 — 본 SPEC의 어떤 AC도 삭제하지 않고, 삭제는 운영자 결정 사항이다(카드 [HARD] 제약).
- 제보자(binsworld) 환경의 226개 stray `.moai`는 제보자 자신의 정리 대상이다 — 본 SPEC은 sync 단계의 GH #1694 회신으로 안내만 한다(plan.md §F M5).

### Out of Scope — 상태 파일의 스키마·경로 체계 변경

- `.moai/state/` 아래 하위 구조(`context-usage/`, `landed/`, `goal/`, `config-cache.json`)와 파일 형식은 불변이다. 고치는 것은 앵커(어느 루트 아래에 두는가)뿐이다. 세션 텔레메트리 스키마(`SPEC-SESSION-TELEMETRY-001`)는 건드리지 않는다.

## 7. 미검증 항목 (Gaps)

- **B4의 configDir 정확한 유도 지점은 미확정이다** — 판정서 Gaps가 명시한 대로, 본 SPEC은 그 유도를 run-phase M3의 RED 테스트가 함수 단위로 고정하게 한다. AC-SA-004의 RED 셀은 그 관측 후 확정된다. lane-1 t507 관측(`fixtures/.moai/state/config-cache.json`)과 B4의 형태 일치는 간접 근거다.
- **B5는 코드 판독만 하고 실행 재현은 하지 않았다**(판정서) — 저위험 분류가 env-우선 설계에 기반한다. 본 SPEC이 B5를 건드리지 않으므로 이 gap은 수리 범위 밖에 남는다.
- **Claude Code가 statusline 프로세스에 `CLAUDE_PROJECT_DIR`을 주는지 미검증**(판정서) — 상태-앵커 체인이 stdin 우선이므로 B1 재현에는 무관하지만, B4 앵커 설계의 env 우선 여부와 상호작용할 수 있다. M3에서 RED 테스트가 실제 사슬을 고정한다.
- **canary 스윕의 선택자와 swept count는 run-phase가 확정한다** — 본 SPEC은 "선택자가 0개 테스트에 매치하면 판정 아님"을 AC-SA-010에 못박는다. 실제 선택자 표현식과 N은 M4 관측에서 확정된다.
- **무프로젝트 skip의 실사용 빈도는 미측정이다** — git repo가 아니면서 상태를 기대하는 사용 사례(예: bare repo, 권한 없는 마운트)가 얼마나 있는지는 관측하지 않았다. REQ-SA-003의 설계 판단("no project, no state")은 판정서 수리 방향 1의 직접 전사다.
