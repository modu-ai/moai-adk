---
id: SPEC-V3R4-CI-FASTTRACK-001
title: "CI/CD Fast Track for Single-Developer Workflow (Path-Filter + Review Bot Consolidation)"
version: "0.1.0"
status: completed
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, cd, github-actions, paths-filter, review-bot, single-developer, productivity"
---

# Design — SPEC-V3R4-CI-FASTTRACK-001

## 1. Architectural Decisions

### AD-001 — Paths-Filter Strategy: `dorny/paths-filter@v3` Detect Job

**Decision**: 신규 `detect` job 에서 `dorny/paths-filter@v3` 를 사용하여 PR diff 를
`docs_only` / `go_code` 두 카테고리로 분류하고, 이후 `test` job 은
`needs: detect` + conditional `if:` 가드로 매트릭스 실행을 결정한다.

**Alternatives considered**:

- (a) **GitHub native `paths:` / `paths-ignore:` trigger keys**: 가장 간단해 보이지만 두
  가지 본질적 한계가 있다.
  1. branch protection 의 required status checks 는 workflow 가 트리거되지 않은 경우
     "expected" 상태로 영구 pending 처리될 수 있음 (GitHub 문서에 명시된 불일관성). 사용자
     보고와 일치하는 known issue.
  2. 단일 workflow 의 일부 job 만 conditional 으로 만들 수 없음 (workflow 전체가
     trigger 되거나 아예 안 됨). ci.yml 의 Build / Lint 등 다른 job 은 항상 동작해야
     함으로 부적합.

- (b) **`dorny/paths-filter@v3` detect job + conditional matrix** (chosen): job-level
  conditional 이 가능하며, branch protection 의 required check name 을 skip-marker 로
  pass 시그널 제공 가능. dorny/paths-filter 는 well-maintained (10K+ stars, monthly
  releases), GitHub Actions ecosystem 의 사실상 표준. 단 한 step 의 추가 overhead
  (~5초) 만 발생.

- (c) **On-merge-only test (PR 단계 skip)**: 회귀 탐지 latency 가 너무 큼. main 으로 머지된
  후에 처음 fail 이 발견되면 revert / 재작업 비용이 큼. 1인 개발이라도 PR 단계 ubuntu
  test 는 유지해야 신뢰성 확보.

**Rationale**: 옵션 (b) 만이 (1) branch protection 호환성, (2) job-level conditional
실행, (3) skip-marker 패턴 적용을 동시에 만족. dorny/paths-filter 는 verified action 이며
self-hosted 의존성 없음.

**Consequence**:

- `ci.yml` 에 ~30 LOC 의 detect job 추가.
- 모든 PR 에 detect job 의 ~5초 overhead 추가 (실제 wait 단축 효과: 5초 추가 vs 3-5분 절감
  = net 압도적 positive).
- paths-filter 의 정확성에 대한 의존성 발생 (R1 risk; nightly safety-net 으로 완화).

### AD-002 — Skip-Marker Pattern over Paths Trigger Keys

**Decision**: Docs-only PR 에서도 branch protection 의 required check (`Test
(ubuntu-latest)`, `CodeQL`) 를 만족시키기 위해, 동일 이름을 가진 별도 "skip-marker"
job 을 실행하여 즉시 success 를 발행한다. 이 패턴은 ci.yml (T1) 과 codeql.yml (T2) 양쪽에
일관 적용한다 — codeql.yml 에 bare `paths-ignore` 를 사용하는 것은 explicit reject.

**Empirical evidence (community)**: GitHub Actions 의 "two jobs with identical `name`,
mutually exclusive `if:` guards" 패턴은 GitHub 의 공식 문서에는 명시되지 않았으나
empirically 작동함이 다수의 community 사례로 확인됨. 본 SPEC 은 단순 추론이 아닌 plan.md
Wave 0 (T0) 의 sandbox PoC 로 동작을 binary 검증한 후 Wave 1 진입.

- Canonical community reference: https://github.com/orgs/community/discussions/13690
  ("Required checks not respecting paths filter" — skip-marker pattern 이 권고되는
  canonical thread).
- 추가 known precondition (community 사례에서 도출):
  1. Workflow 가 `on.paths` / `on.paths-ignore` 으로 전체 gating 되면 안 됨 (그러면
     workflow 자체가 트리거되지 않아 required check 가 영구 pending).
  2. 두 mutually exclusive job 중 정확히 한 개만 한 PR 의 한 run 에서 report 되어야 함.
  3. Required check matching 은 check-run **name** 으로 수행됨. workflow name 매칭은
     legacy fallback 으로만 동작 — Wave 0 (T0) 에서 어느 매칭 모드가 실제로 satisfy
     하는지 binary 확인.

**Alternatives considered**:

- (a) **GitHub built-in "skipped = success" 동작에만 의존**: GitHub 의 required check
  semantics 는 *skipped* 상태에 대해 일관성 없음. paths-filter 로 workflow 자체가 트리거되지
  않은 경우는 명시적으로 "expected, never reported" 상태가 되어 PR 이 영구 pending 됨.
  workflow 내부의 job 이 `if:` 가드로 skip 된 경우는 success 로 간주되는 경우가 다수지만,
  matrix job + branch protection 조합에서는 검증되지 않은 영역. risk 가 큼.

- (b) **별도 skip-marker job + 동일 이름 매칭** (chosen): branch protection 의 required
  check 는 `Test (ubuntu-latest)` 라는 정확한 이름과 매칭. skip-marker job 이 동일 matrix
  name 으로 즉시 success 신호 발행 → 보호 규칙 자동 satisfaction. GitHub Actions
  ecosystem 의 well-known workaround (community discussion #13690 evidence).

- (c) **required check 자체를 모두 제거**: 의도와 어긋남. 보호 규칙은 회귀 안전망이며
  사용자 (B) 결정에서도 4개 항목은 의도적으로 유지.

- (d) **codeql.yml 에 bare `paths-ignore` 적용** (rejected): 위 known precondition #1
  위반. codeql.yml 의 workflow 자체가 트리거되지 않아 `CodeQL` required check 가 영구
  pending 으로 PR block. 본 SPEC 의 plan-audit iteration 1 에서 P0 defect D1 으로 명시
  지적됨.

**Rationale**: (b) 만이 결정성 (deterministic check satisfaction) + branch protection
보존 + 의도 명확성을 모두 충족. (d) 의 explicit rejection 으로 ci.yml + codeql.yml 양쪽이
**동일 skip-marker pattern** 으로 일관성 유지.

**Wave 0 PoC verification gate**:

Wave 1 (T1, T2) implementation 진입 전, plan.md T0 가 sandbox PR 로 다음을 binary 검증:

1. CodeQL 의 실제 satisfying check-run name (`gh api ... --jq '.check_runs[].name' |
   grep -i codeql`).
2. Skip-marker job 이 docs-only PR 의 `gh pr checks` 에서 SUCCESS 로 보고되는지.
3. branch protection 의 required check 가 skip-marker pass 로 satisfy 되는지.

Wave 0 PASS 결과 (observed satisfying name + SUCCESS verification) 를 본 AD-002 의 별도
section "Wave 0 PoC Result" 에 run-phase 가 기록.

**Consequence**:

- `Test (ubuntu-latest)` 이름이 docs-only PR 과 go-code PR 에서 서로 다른 job 으로 매칭됨
  — 의도된 설계. PR 상세 페이지에서 사용자가 "어느 job 이 실행됐는가" 를 보려면 workflow
  run 을 확인해야 함 (minor UX overhead).
- matrix name template (`Test (${{ matrix.os }})`) 의 *정확한* 일치 필수. 오타 1자만
  있어도 보호 규칙 미충족 → PR block. T1 verification 에서 binary 확인.
- codeql.yml 의 `analyze-skip-marker` job 의 `name:` 은 Wave 0 결과에 따라 `Analyze (Go)`
  (matrix language: [go] 와 함께 → `Analyze (Go) (go)` emit) 로 설정. legacy workflow-name
  matching 의존시 단일 job `name: CodeQL`.

### AD-003 — Review Bot Consolidation: Keep Claude-Code-Review, Remove 5

**Decision**: 5개 review workflow (codex-review / gemini-review / glm-review / llm-panel /
claude-code-review.optional) 를 제거하고, `claude-code-review.yml` 단 1개를 유지한다.
`claude.yml` (issue/comment trigger) 와 `review-quality-gate.yml` (Claude Code Review
severity parser) 는 codex-independent 이므로 PRESERVE.

**Decision matrix (per workflow)**:

| Workflow | Trigger | Codex 의존? | 실제 가치 | 결정 |
|----------|---------|-------------|-----------|------|
| `claude-code-review.yml` | PR opened/synchronize/reopened | NO (Anthropic API) | HIGH (사용자 신뢰, 실제 회귀 탐지) | PRESERVE |
| `codex-review.yml` | PR | YES (codex CLI 미설치) | RED 영구 | DELETE |
| `gemini-review.yml` | PR | YES (gemini CLI 미설치) | RED 영구 | DELETE |
| `glm-review.yml` | PR | partial (glm 설정 미완) | redundancy with claude-code-review | DELETE |
| `llm-panel.yml` | PR (depends on 3 above) | YES (aggregator) | orphan after DELETE | DELETE |
| `claude-code-review.optional.yml` | PR | NO | duplicate of main | DELETE |
| `claude.yml` | issue/comment `@claude` | NO | independent value (사용자 요청 응답) | PRESERVE |
| `review-quality-gate.yml` | check_run completed | NO (gh API + Claude Code Review check_run) | severity gating | PRESERVE |

**Why claude-code-review.yml survives**:

- Anthropic OAuth credential 이 repo secrets 에 이미 active.
- 실제 회귀 탐지 사례 다수 (사용자 보고에 따른 신뢰).
- LLM panel 5개 중 단 1개를 골라야 한다면 가장 robust signal source.

**Why the other 4 die**:

- `codex` / `gemini` CLI 가 runner machine 에 없음 → 워크플로우 step 자체가 `command not
  found` 로 실패 → 모든 PR 에서 RED. 신호 가치 0.
- `glm-review` 는 review setting 미완 (사용자 보고).
- `llm-panel` 은 위 3개의 결과를 aggregating 하므로 orphan.

**Why claude.yml + review-quality-gate.yml live**:

- `claude.yml`: 사용자가 issue 또는 PR comment 에서 `@claude` 를 호출할 때만 트리거. PR
  자동 트리거 아님. codex 비의존. 별도 가치 (사용자 직접 호출 기능). T4 audit grep 으로
  검증.
- `review-quality-gate.yml`: `claude-code-review.yml` 의 check_run 종료 시점에 트리거되어
  severity JSON 을 파싱하고 PR comment 작성. codex 비의존 (gh API 만 사용). claude-code-review
  의 부속 dependency 이므로 함께 PRESERVE.

**Alternatives considered**:

- (a) **모든 review bot 제거, manual review 만**: 회귀 탐지 안전망 손실. 1인 개발일수록
  LLM review 의 second-pair-of-eyes 가치가 큼. user 가 명시적으로 claude-code-review.yml
  보존 요청.
- (b) **GLM-only (cost optimization)**: GLM 의 review 신뢰도가 검증되지 않음. claude API
  비용은 PR 당 ~$0.10 수준으로 acceptable.
- (c) **Workflow PRESERVE 하되 codex CLI install step 추가**: codex 의 API key /
  credential 분배 복잡도가 큼. 1인 개발자가 codex 와 claude 양쪽 quota 를 관리할 동기
  부족.

**Rationale**: 사용자 보고 + grep audit (T4) + ROI 분석 모두 chosen decision matrix 를
지지.

**Consequence**: review 영역은 단일 source (claude-code-review.yml) 로 단순화. private-guard
job 은 자동 소멸 (codex-review.yml + llm-panel.yml 둘 다 DELETE 됨). 미래에 사용자가
review bot 다양화를 원하면 별도 SPEC 으로 root cause (CLI / credential / ROI) 처리.

### AD-004 — `claude.yml` / `review-quality-gate.yml` / `private-guard` Decision Matrix

**Decision Template** (run-phase T4 가 audit 수행 후 이 매트릭스 채움):

| Item | Audit method | Expected result | Decision |
|------|-------------|-----------------|----------|
| `claude.yml` workflow | `grep -nE 'codex\|gemini\|glm' .github/workflows/claude.yml` | 0 matches | **PRESERVE** |
| `review-quality-gate.yml` workflow | `grep -nE 'codex\|gemini\|glm' .github/workflows/review-quality-gate.yml` | 0 matches | **PRESERVE** |
| `private-guard` job name | `grep -rln 'private-guard\|private_guard' .github/workflows/` | matches in `codex-review.yml` + `llm-panel.yml` only | **자동 소멸** (T3 에서 두 파일 DELETE) |

PRE-PLAN 정찰 결과 (orchestrator 가 plan-PR 작성 시점에 grep 으로 사전 확인) 와 일치하므로
T4 audit 는 검증 단계 (no edits). 만약 grep 결과가 예상과 다르면 (예: review-quality-gate
가 codex 를 invoke 하면) 결정을 GUARD (graceful `if: command -v codex` wrapper + 
`continue-on-error`) 로 전환. run-phase 가 결정.

**Why audit-first vs delete-first**:

- T4 audit 를 T3 *이전* 에 수행하여 PRESERVE 대상이 우발 삭제되는 것을 방지 (R5
  mitigation).
- audit 결과를 design.md AD-004 에 기록함으로써 미래 reviewer 가 결정의 근거를 추적 가능.

### AD-005 — Nightly Schedule over Per-PR Multi-OS Test

**Decision**: macOS / Windows 매트릭스를 per-PR required check 에서 제거하고, daily 03:00
UTC + release tag push 시점의 nightly workflow 로 이전한다.

**Trade-off analysis**:

| 차원 | Per-PR multi-OS | Nightly + tag-push |
|------|-----------------|---------------------|
| 회귀 탐지 latency | 즉시 (PR 단위) | 24h (nightly) 또는 release 시점 |
| PR-level wait time | 5-6분+ | 1-2분 (~70% 절감) |
| Runner cost | 매 PR × 3 OS | 매일 1회 × 3 OS |
| Cross-platform 안전망 | 강함 | 강함 (지연되지만 보장됨) |
| 1인 개발 ROI | 부정적 (wait > 가치) | 긍정적 (안전망 유지 + wait 절감) |

**1인 개발 trade-off**:

- 1인 개발 cadence 는 hot-fix 가 즉시 production 으로 가지 않음. 보통 plan → run → sync
  3단계 PR 사이클 후 사용자가 명시적으로 release tag 발행. nightly + tag-push 가 release
  전에 항상 한 번은 트리거됨 → release 차단 가능.
- PR 단계의 즉시 회귀 탐지는 ubuntu-latest 만으로도 충분 (Go stdlib + runtime 일관성;
  OS-specific bug 는 빈도 낮음).

**Karpathy / forrestchang anti-pattern catalog 검토**:

- "Premature optimization" anti-pattern 의 명시적 *negation*. 본 SPEC 의 trigger 는
  empirical (사용자 직접 5-6분 wait 보고) + measured (20개 workflow + branch protection
  6 check 의 실측). 이론적 가설이 아닌 측정값 기반 right-sizing.
- "Over-engineering" anti-pattern 의 명시적 해소. review bot 4개 RED 가 실제로 신호 잡음.
  removal 자체가 anti-pattern 제거.

**Alternatives considered**:

- (a) **Per-PR multi-OS 유지, runner 만 self-hosted 로**: cost prohibitive (월 인프라
  + 유지보수 부담).
- (b) **PR 에서 macOS 만 제거, Windows 만 유지**: 임의적 비대칭. Windows 가 macOS 보다
  flake 가 잦은 platform 임에도 유지하는 근거 없음.
- (c) **Per-PR multi-OS 유지, `go test -short` 만 실행**: race detector 가 main hot-loop;
  -short 만으로는 wait 감소가 marginal.

**Rationale**: (B) 결정 (외부 적용) + nightly safety-net 결합이 단일 일관성 있는 trade-off.

**Consequence**:

- 회귀 탐지 latency 가 24h (nightly) 까지 지연될 수 있음. release tag 발행 직전에는 항상
  최신 main HEAD 의 매트릭스 결과 가용 (tag push trigger).
- 만약 nightly 가 RED 라면 issue 자동 생성 → 다음 PR 작성 전에 사용자 인지.

## 2. Failure Modes & Recovery

### FM-1 — paths-filter Misclassifies Go Change as Docs-Only

**Scenario**: T1 의 paths-filter 패턴이 일부 Go 변경 (예: `.github/workflows/*.yml` 외의
CI-related 파일, 혹은 새로운 build script 도입) 을 `docs_only` 로 잘못 분류. skip-marker
job 이 false success 반환. 회귀가 ubuntu-latest 테스트 없이 main 으로 머지됨.

**Recovery**:

1. nightly full-matrix (T6) 가 24h 이내 회귀 탐지 → issue 자동 생성.
2. 회귀 commit 식별 후 hotfix PR.
3. paths-filter 패턴을 *해당 파일 카테고리 추가* 로 업데이트 (run-phase 후속 작업으로 별도
   PR).
4. 또는 workflow_dispatch 로 nightly 즉시 실행해 main 상태 사전 검증 가능.

**근본 mitigation**: paths-filter 패턴이 `go_code` 카테고리를 inclusive 하게 정의 (T1
deliverable 의 `**/*.go` + Makefile + workflow yaml + go.mod/sum 모두 포함). 보수적
filter 설계.

### FM-2 — Skip-Marker Job Name Mismatch

**Scenario**: T1 구현 시 skip-marker job 의 `name:` 이 실제 test job 의 `name:` 과
타이포 1자 차이 (e.g., `Test (Ubuntu-latest)` vs `Test (ubuntu-latest)`). branch
protection 의 required check 가 매칭되지 않아 PR 영구 pending.

**Recovery**:

1. PR 본문에서 `gh pr checks <PR> --json statusCheckRollup` 로 mismatch 검출.
2. ci.yml T1 hunk 재편집해 정확한 matrix name 적용.
3. branch protection api 결과의 contexts 와 정확히 일치하는지 binary 검증.

**근본 mitigation**: T1 verification 의 checklist 에 binary 일치 검증 포함. matrix name
template 사용 (`Test (${{ matrix.os }})`) 으로 두 job 이 같은 source-of-truth 참조.

### FM-3 — lefthook Pre-Push Gate Slowness Annoyance

**Scenario**: 개발자가 매 push 마다 `make preflight` 실행 (~1-2분) 으로 인해 lefthook 을
우회하기 시작 (`LEFTHOOK=0 git push` 남용). 결과적으로 pre-flight 안전망 무효.

**Recovery**:

1. `LEFTHOOK=0` 우회 사용 빈도 monitoring (git log + bash history 분석은 user-side
   responsibility).
2. preflight 가 너무 느리면 (a) `test -short -race` 만 유지하고 (b) lint-fast 만 유지로
   분기 분할 가능 (별도 SPEC).
3. lefthook 의 graceful degradation: 미설치 시 `git push` 정상 동작 (R3 mitigation 과
   동일).

**근본 mitigation**: T5 deliverable 의 `test-race-short` 만 runtime 으로 (race + short 만).
실제 시간은 ubuntu-latest CI 의 race 전체 대비 60-70% 단축 예상.

### FM-4 — Nightly Full-Matrix Flake Noise

**Scenario**: Windows ETXTBSY flake (CLAUDE.local.md §18.11 known issue) 또는 macOS-specific
flake 가 nightly 마다 발생. dedup 로직이 24h window 만 보호하므로 같은 flake 가 매일 issue
생성.

**Recovery**:

1. T6 의 dedup 가 24h 안에서는 댓글로 통합. 25h+ 부터는 새 issue.
2. dedup window 를 7d 등으로 확장하는 별도 PR (run-phase 후속).
3. 알려진 flake 는 `t.Skip("known flake — TestSupervisor_NonZeroExit ETXTBSY race")` 로
   sourcecode 차원에서 skip 처리 (CLAUDE.local.md §6 Go Test Execution Rules 참조).

**근본 mitigation**: nightly title 에 commit SHA short 포함 → 같은 commit 의 동일 OS
실패는 자연 dedup. 다른 commit 이지만 같은 flake 패턴이면 24h window 내 댓글 통합.

### FM-5 — `review-quality-gate.yml` 또는 `claude.yml` 우발 삭제

**Scenario**: T3 의 batch delete 명령에서 audit 결과 (T4) 가 잘못 적용되어 PRESERVE 대상이
함께 삭제됨.

**Recovery**:

1. git revert run-PR T3 hunk.
2. design.md AD-004 가 PRESERVE 결정을 명시했음에도 삭제가 발생한 경우 → run-phase agent
   의 reviewer self-check 실패. acceptance gate AC-CIFT-004 binary 검증으로 차단됨.
3. PR diff 가 변경 파일 목록을 자동 노출 → 사용자 검토 가능.

**근본 mitigation**: T4 audit-first 순서 + T3 의 명시적 5-file naming + acceptance gate
이중 검증.

## 3. Cross-Cutting Concerns

### CC-1 — Token Budget Impact

본 SPEC 의 run-PR 산출물 중 LLM context 에 영향:

- ci.yml + codeql.yml + lefthook.yml + Makefile + release-pr-multi-os.yml: GitHub
  Actions / Makefile 형식 파일들로 progressive disclosure system 의 paths trigger 와
  무관. PR review 시에만 reviewer 의 context 에 로드됨 — 일회성.
- CLAUDE.local.md §18.7 갱신: ~15 LOC 변경. CLAUDE.local.md 는 매 세션 자동 로드되므로
  session-level token cost ~30-50 tokens 증가.
- lessons.md #19: ~65 LOC. lessons.md 는 stale 검사 후 선택적 로드 (24h mtime rule per
  moai-memory.md). 신규 session 에는 token cost 영향 미미.

**Net 토큰 영향**: 세션 당 ~50-100 tokens 증가 (CLAUDE.local.md). PR-level wait 절감
3-4분 / PR 대비 trivial.

paths-filter detection 자체는 LLM context 가 아닌 GitHub Actions runtime 비용:
- detect job overhead: ~5초 / PR (모든 PR).
- Net effect: docs-only PR 의 경우 5초 추가 vs 3-5분 절감 = 압도적 positive.

### CC-2 — Lessons-Protocol Compliance

T8 의 lessons.md #19 entry 는 MoAI Lessons Protocol (`.claude/rules/moai/core/moai-constitution.md`
§ Lessons Protocol) 의 5-section structure (Category / Incorrect / Correct / Why / How
to apply) 를 엄격 준수. REQ-CIFT-008 본문이 verbatim template.

### CC-3 — No Production Code Changes

본 SPEC 의 모든 변경은 CI / build / doctrine layer:

- Go code (internal/, pkg/, cmd/): 0 LOC 수정
- 테스트 파일 (`*_test.go`): 0 LOC 수정
- production binary 동작: 영향 없음

`go test ./...` 게이트 (AC-CIFT-009) 는 본 SPEC 의 sanity check (의도되지 않은 code touch
검출용) 으로만 사용. 모든 Go test 가 PASS 상태를 유지해야 함.

### CC-4 — Cross-SPEC Coupling: Minimal

- **SPEC-V3R4-WORKFLOW-SPLIT-001** (Bundle F): orthogonal. Bundle F 는
  `.claude/skills/moai/workflows/*.md` (entry router + sub-skill split) 작업. 본 SPEC 은
  `.github/workflows/*.yml` (CI YAML) 작업. 두 SPEC 이 동일 파일을 수정하지 않음.
  conflict probability 0 으로 평가.
- **SPEC-WORKTREE-DOCS-001** (sibling): documentation-only SPEC, 본 SPEC 과 file
  scope overlap 없음.
- **미래 release SPEC (v3.0.0-rc1)**: 본 SPEC 의 release-pr-multi-os.yml 이 tag-push
  트리거를 포함하므로, release tag 발행 시점에 release-PR run 이 sequence 안에 자연 통합됨.
  본 SPEC 이 release readiness 의 precondition.

### CC-5 — Meta-Level Benefit (Future SPEC Latency)

본 SPEC 의 paths-filter / review consolidation 효과는 *모든 미래 SPEC* 의 PR cycle 에
적용됨. 추정:

- 미래 SPEC 평균 PR 수: 3 (plan-PR + run-PR + sync-PR)
- PR 당 wait 단축: 3-4분 (docs-heavy PR 에서 더 크게)
- 미래 SPEC 100건 가정 시 누적 절감: 100 × 3 × 3-4분 = 900-1200분 (15-20시간)

이 meta-level benefit 이 본 SPEC 의 P1 priority 정당화.

## 4. References

- **Triggering**: 사용자 보고 2026-05-17 (1인 개발 CI 비용 비대칭, review bot 영구 RED).
- **(B) 결정 baseline**: branch protection PATCH 적용 시점 (main HEAD `680232c82`).
- **Current main HEAD**: `41b6f37dc` (plan-PR base).
- **Anti-pattern catalog**: Karpathy "premature optimization" + "over-engineering" —
  본 SPEC 은 두 anti-pattern 의 명시적 negation (empirical trigger + ROI-driven scope).
- **Workflow files (current)**: ci.yml (226 LOC), codeql.yml (52 LOC), claude.yml (64
  LOC), review-quality-gate.yml (134 LOC).
- **External actions**: `dorny/paths-filter@v3`, `actions/checkout@v5`,
  `actions/setup-go@v6`, `actions/github-script@v7`, `evilmartians/lefthook` (CLI).
- **MoAI rule references**:
  - `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol (T8 format spec)
  - `.claude/rules/moai/workflow/moai-memory.md` (lessons.md token budget)
  - `.claude/rules/moai/development/spec-frontmatter-schema.md` (SPEC frontmatter SSOT)
- **CLAUDE.local.md**: §18.7 (target of T7), §18.9 (Release Drafter + GoReleaser 보존),
  §18.11 (v2.14.0 Case Study lessons — Windows flake 인지).
