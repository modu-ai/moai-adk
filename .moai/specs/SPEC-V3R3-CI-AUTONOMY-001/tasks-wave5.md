---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 5
version: 0.1.0
status: draft
created_at: 2026-05-08
updated_at: 2026-05-08
author: manager-strategy
---

# Wave 5 Atomic Tasks (Phase 1 Output)

> Companion to strategy-wave5.md. SPEC-V3R3-CI-AUTONOMY-001 Wave 5 — T6 Worktree State Guard.
> Generated: 2026-05-08. Methodology: TDD for Go primitives (W5-T01..T06) + verify-via-grep for markdown (W5-T07, T08). Wave Base: `origin/main 311f27a2a`.

---

## Atomic Task Table

| Task ID | Description | Files (provisional) | REQ | AC | Dependencies | File Ownership Scope | Status |
|---------|-------------|---------------------|-----|----|--------------|---------------------|--------|
| W5-T01 | State snapshot capture 함수. `git rev-parse HEAD`, `git rev-parse --abbrev-ref HEAD`, `git status --porcelain` (sorted), `git ls-files --others --exclude-standard .moai/specs/` 결과를 `Snapshot` struct 로 캡처. JSON 직렬화 + `.moai/state/worktree-snapshot-<UUID>.json` 저장 (schema_version "1.0.0"). 30s timeout via `exec.CommandContext`. detached HEAD / empty repo graceful degradation (`HeadSHA = ""` + warning). | `internal/worktree/state_guard.go` (new), `internal/worktree/snapshot_io.go` (new), `internal/worktree/doc.go` (new) | REQ-CIAUT-031 | AC-CIAUT-014 | none (foundational) | implementer | pending |
| W5-T02 | State diff 함수 — pre vs post Snapshot 비교. 4-dimension binary detection: HeadChanged, BranchChanged, UntrackedAdded (set diff), UntrackedRemoved (set diff), PorcelainDelta (line-by-line). `Divergence.IsDivergent()` = 4 boolean OR + 2 slice non-empty. Stable ordering (sort.Strings) for deterministic output. | `internal/worktree/state_guard.go` (extend) | REQ-CIAUT-032 | AC-CIAUT-014 | W5-T01 | implementer | pending |
| W5-T03 | Orchestrator wrapper / CLI subcommands. `moai worktree snapshot [--out <path>] [--agent-name <name>]` (capture + JSON 저장; exit 0 + stdout snapshot ID), `moai worktree verify --snapshot <path> [--agent-response <json>]` (pre-state 로드 + post-state 캡처 + Divergence + SuspectFlag 계산; exit 0 clean / 1 divergence / 2 suspect / 3 both; stdout JSON report). cobra subcommand 등록을 `internal/cli/worktree/root.go` 에 추가. | `internal/cli/worktree/guard.go` (new), `internal/cli/worktree/root.go` (extend) | REQ-CIAUT-032 | AC-CIAUT-014 | W5-T01, W5-T02 | implementer | pending |
| W5-T04 | Divergence 발생 시 markdown structured report writer. 위치 `.moai/reports/worktree-guard/<YYYY-MM-DD>.md` (per-day rolling, append). Schema: ISO-8601 timestamp + Snapshot ID + Agent name + 4 dimension breakdown + JSON sidecar path. JSON sidecar 도 같이 작성 (`<DATE>-<id>.json`). orchestrator 가 AskUserQuestion 책임 (Go 는 report + exit code 만; agent-common-protocol §User Interaction Boundary HARD 준수). | `internal/worktree/divergence_log.go` (new), `internal/cli/worktree/guard.go` (extend `verify` subcommand) | REQ-CIAUT-033 | AC-CIAUT-014 | W5-T03 | implementer | pending |
| W5-T05 | Empty `worktreePath` suspect detection. `--agent-response <json>` 입력의 `worktreePath` 필드가 `{}` / missing / empty string 일 때 `SuspectFlag` 생성 + `.moai/state/worktree-suspect-<id>.flag` 파일 작성. `verify` subcommand 의 exit code 2 (suspect) 또는 3 (both divergence + suspect). Suspect flag 파일 stale 처리 (1시간 이상 → warning only, hard block 안 함). | `internal/cli/worktree/guard.go` (extend) | REQ-CIAUT-034 | AC-CIAUT-015 | W5-T03 | implementer | pending |
| W5-T06 | State restore 옵션. `moai worktree restore --snapshot <path> [--dry-run]` subcommand. 명령: `git restore --source=<HeadSHA> --staged --worktree :/`. Untracked file paths 는 stdout 에 list 출력 + "manual recreation required (untracked files cannot be restored from git)" 명시 (C-3 honest scope concern). `--dry-run` 시 실행 없이 명령 + untracked list 만 출력. | `internal/cli/worktree/guard.go` (extend) | REQ-CIAUT-036 | AC-CIAUT-014 (restore path) | W5-T03 | implementer | pending |
| W5-T07 | NEW agent: `claude-code-guide.md`. Anthropic upstream `Agent(isolation:)` 회귀 조사 + bug report 작성 책임. YAML frontmatter (description + EN/KO/JA/ZH keywords + tools: Read/Grep/Glob/WebFetch/WebSearch + model: sonnet + permissionMode: plan + memory: project + skills: moai-foundation-core). 본문: identity + workflow (regression detection → local analysis → user-facing report). Placeholder report 작성 `.moai/reports/upstream/agent-isolation-regression.md` (frontmatter only, body `_TBD_`). **NOTE: claude-code-guide 는 dev project 와 user-template 양쪽에 부재 → NEW 작성 (extension 아님; strategy-wave5 §14 C-1 참조).** | `internal/template/templates/.claude/agents/moai/claude-code-guide.md` (new), `.moai/reports/upstream/agent-isolation-regression.md` (new placeholder) | REQ-CIAUT-035 | AC-CIAUT-015 | none (independent) | implementer | pending |
| W5-T08 | NEW rule doc: `worktree-state-guard.md`. paths frontmatter (예: `**/*.go,**/.claude/agents/**`). 내용: when to snapshot (orchestrator 매 `Agent(isolation: "worktree")` 호출 직전), divergence threshold (binary, OQ4 결정), escalation path (suspect → AskUser → restore/accept/abort), invocation pattern (Bash CLI 호출 시퀀스), cross-reference (worktree-integration.md, agent-common-protocol.md). | `internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md` (new) | REQ-CIAUT-031~036 (governance) | AC-CIAUT-014 + AC-CIAUT-015 transitive | none (independent) | implementer | pending |

---

## Dependency Graph

```
W5-T01 (snapshot capture)
   │
   ├─→ W5-T02 (diff function)
   │       │
   │       └─→ W5-T03 (CLI subcommands snapshot + verify)
   │              │
   │              ├─→ W5-T04 (divergence logging)
   │              │
   │              ├─→ W5-T05 (suspect flag)
   │              │
   │              └─→ W5-T06 (restore subcommand)
   │
   └── (parallel) W5-T07 (claude-code-guide NEW agent)
                       │
                       └── (parallel) W5-T08 (worktree-state-guard rule)
```

**Sequential**: T01 → T02 → T03 → {T04, T05, T06}.
**Parallel**: T07 + T08 (markdown only, no Go dependency) can run in parallel with T01-T06.

**Suggested commit cadence** (Conventional Commits + 🗿 MoAI co-author trailer):

1. `test(worktree): W5-T01 RED — Snapshot capture cases`
2. `feat(worktree): W5-T01 implement Snapshot + git rev-parse/porcelain/ls-files`
3. `test(worktree): W5-T02 RED — Diff cases (4 dimensions)`
4. `feat(worktree): W5-T02 implement Diff + IsDivergent`
5. `feat(cli): W5-T03 add moai worktree snapshot|verify subcommands`
6. `feat(worktree): W5-T04 add divergence markdown logger`
7. `feat(cli): W5-T05 detect empty worktreePath + suspect flag file`
8. `feat(cli): W5-T06 add moai worktree restore subcommand`
9. `chore(agents): W5-T07 add claude-code-guide agent + upstream report placeholder`
10. `chore(rules): W5-T08 add worktree-state-guard.md rule`

---

## File Ownership Assignment (Solo 모드 — sub-agent + --branch 패턴, lessons #13)

### Implementer Scope (write access)

```
internal/worktree/state_guard.go                                          # W5-T01, W5-T02
internal/worktree/snapshot_io.go                                          # W5-T01 (JSON IO 분리)
internal/worktree/divergence_log.go                                       # W5-T04 (markdown writer)
internal/worktree/doc.go                                                  # W5-T01 (package doc)
internal/cli/worktree/guard.go                                            # W5-T03, W5-T05, W5-T06
internal/cli/worktree/root.go                                             # extend — register guard subcommand (read existing first)
internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md  # W5-T08 NEW
internal/template/templates/.claude/agents/moai/claude-code-guide.md      # W5-T07 NEW
.moai/reports/upstream/agent-isolation-regression.md                      # W5-T07 placeholder
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave5.md                   # this strategy
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave5.md                      # this tasks file
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/progress.md                         # extend (Phase 1 + 1.5 entries)
```

After `make build` (derived files — not directly hand-edited):

```
.claude/rules/moai/workflow/worktree-state-guard.md                       # auto-derived from template
.claude/agents/moai/claude-code-guide.md                                  # auto-derived from template
internal/template/embedded.go                                             # auto-regenerated by make build
```

### Tester Scope (test 파일만 write, production 코드 read-only)

```
internal/worktree/state_guard_test.go                                     # 4 mandatory cases (no-diff / untracked-added / untracked-removed / branch-changed)
internal/worktree/snapshot_io_test.go                                     # JSON roundtrip
internal/cli/worktree/guard_test.go                                       # CLI integration (cobra-driven)
```

### Read-Only Scope (cross-Wave consumer / SSoT source)

```
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md                             # frozen for Wave 5
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/plan.md                             # frozen for Wave 5
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/acceptance.md                       # frozen for Wave 5
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave[1-4].md               # prior waves audit trail
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave[1-4].md                  # prior waves audit trail
internal/cli/worktree/new.go, status.go, list.go, switch.go, sync.go      # existing cobra patterns (read-only reference)
internal/cli/worktree/root_test.go, subcommands_test.go                   # existing test patterns (read-only reference)
.claude/rules/moai/workflow/worktree-integration.md                       # existing worktree rule (Wave 5 worktree-state-guard.md cross-references)
.claude/rules/moai/core/agent-common-protocol.md                          # AskUserQuestion HARD rule (orchestrator-only)
.claude/agents/moai/expert-debug.md                                       # agent frontmatter pattern reference (W5-T07 작성 시)
.claude/agents/moai/researcher.md                                         # agent frontmatter pattern reference
CLAUDE.md, CLAUDE.local.md                                                # project rules
```

### Implicit Read Access (모든 task)

- `.claude/rules/moai/**` (auto-loaded rules — agent-common-protocol.md, askuser-protocol.md, worktree-integration.md)
- 모든 Wave 1-4 산출물 (read-only consumer)

---

## AC Mapping

| Wave 5 Task | Drives AC | Validation |
|-------------|-----------|------------|
| W5-T01 + W5-T02 | AC-CIAUT-014 (worktree state divergence detection) | unit test 4 cases (no-diff/added/removed/branch-changed) PASS; `t.TempDir()` isolation 검증 |
| W5-T03 + W5-T04 | AC-CIAUT-014 (divergence detection + reporting) | CLI integration test: `moai worktree verify` exit code 0/1; markdown report 파일 생성 + JSON sidecar |
| W5-T05 | AC-CIAUT-015 (empty worktreePath suspect handling) | CLI integration test: empty `worktreePath` fixture → exit 2 + suspect flag 파일 생성 |
| W5-T06 | AC-CIAUT-014 (restore path for divergence recovery) | CLI integration test: `--dry-run` → 명령 + untracked list stdout; manual replay (실제 restore 동작) |
| W5-T07 | AC-CIAUT-015 (claude-code-guide upstream investigation) | YAML frontmatter valid (description + tools + model + permissionMode + memory + skills 필드 존재); placeholder report 파일 frontmatter valid |
| W5-T08 | AC-CIAUT-014 + AC-CIAUT-015 transitive (governance) | rule doc grep checks: "## When to Snapshot" 헤딩 존재, "## Divergence Threshold" 헤딩 존재, "## Escalation Path" 헤딩 존재, worktree-integration.md cross-reference 존재 |

---

## TRUST 5 Targets (Wave 5 SPEC-Level DoD)

| Pillar | Target | Verification |
|--------|--------|--------------|
| **Tested** | `internal/worktree/` package coverage ≥ 85%; `internal/cli/worktree/guard.go` ≥ 80%; 4 mandatory unit cases (no-diff/untracked-added/untracked-removed/branch-changed) + JSON roundtrip + 5 CLI integration cases (snapshot/verify-clean/verify-diff/verify-suspect/restore-dry-run) 모두 PASS; `t.TempDir()` 격리 100%; `go test -race` clean | `go test -cover ./internal/worktree/...`; `go test -cover ./internal/cli/worktree/...`; `go test -race ./internal/...` |
| **Readable** | godoc 모든 exported type/func + package doc.go; `golangci-lint run ./internal/...` 0 issue; markdown rule + agent file 의 H2/H3 구조 명확; godoc 검증 (`go doc internal/worktree`) | `golangci-lint run`; `markdownlint internal/template/templates/.claude/rules/moai/workflow/worktree-state-guard.md`; `markdownlint internal/template/templates/.claude/agents/moai/claude-code-guide.md` |
| **Unified** | gofmt + goimports 일관성; cobra subcommand naming `<verb>` (snapshot/verify/restore); rule doc 헤딩 구조 worktree-integration.md 와 동일 톤; agent file frontmatter 가 expert-debug/researcher 패턴 따름 | `gofmt -l ./internal/...` empty; `goimports -l` empty; rule doc cross-reference 검증 |
| **Secured** | snapshot 파일에 secrets 미노출 (porcelain 이 path 만 포함, file content 미포함); `git restore` 호출은 orchestrator 측 AskUser 후에만 (Go CLI 는 `--dry-run` 가 아닌 호출도 명시적 confirm 메시지 출력); `.moai/state/` permission 0644; restore 가 워킹 트리 외 파일 변경 안 함 (`:/` scope) | snapshot JSON 검사: 모든 path 만 포함 + content 0; `moai worktree restore` 출력에 "this will discard local changes" 명시; restore 가 git pathspec `:/` 사용 검증 |
| **Trackable** | 모든 commit 이 SPEC-V3R3-CI-AUTONOMY-001 W5 reference 포함; Conventional Commits + 🗿 MoAI co-author trailer; divergence log 가 audit trail 보존 (`.moai/reports/worktree-guard/<DATE>.md`); suspect flag 파일이 audit trail 보존 (`.moai/state/worktree-suspect-<id>.flag`); claude-code-guide 호출 시 placeholder report 갱신 (W5-T07 본문 _TBD_ 는 향후 sessions 에 작성됨) | `git log --grep='SPEC-V3R3-CI-AUTONOMY-001 W5'`; divergence log + suspect flag 파일 schema 검증 |

---

## Per-Wave DoD Checklist

- [ ] 모든 8개 W5 task 완료 (W5-T01 ~ W5-T08, 위 표 참조)
- [ ] **Template-First mirror 검증** — `worktree-state-guard.md` rule + `claude-code-guide.md` agent 양쪽이 `internal/template/templates/.claude/...` 에 source 위치 (CLAUDE.local.md §2)
- [ ] `make build` 후 embedded.go 갱신 + `.claude/rules/moai/workflow/worktree-state-guard.md` + `.claude/agents/moai/claude-code-guide.md` derived 사본 자동 생성 확인
- [ ] `internal/worktree/` 신규 패키지 작성 완료 (state_guard.go + snapshot_io.go + divergence_log.go + doc.go)
- [ ] `internal/cli/worktree/guard.go` 신규 작성 + `internal/cli/worktree/root.go` 에 subcommand 등록
- [ ] 4 mandatory unit tests PASS (no-diff / untracked-added / untracked-removed / branch-changed)
- [ ] JSON roundtrip test PASS (snapshot Marshal → Unmarshal 동일성)
- [ ] CLI integration tests PASS (snapshot / verify clean / verify diff / verify suspect / restore dry-run, 총 5 cases)
- [ ] `go test -race ./...` 통과 (concurrency safety; orchestrator 의 sequential invocation 가정 하에서도 race detector clean)
- [ ] `go test -cover ./internal/worktree/...` ≥ 85%
- [ ] `go test -cover ./internal/cli/worktree/...` ≥ 80%
- [ ] `golangci-lint run ./internal/...` 0 issue
- [ ] `make ci-local` 통과 (Wave 1 framework 회귀 없음)
- [ ] `claude-code-guide.md` agent YAML frontmatter valid (`name`, `description`, `tools` (CSV string), `model`, `permissionMode`, `memory`, `skills` 필드 모두 존재 + `tools` CSV string 형식, `skills` YAML array)
- [ ] `.moai/reports/upstream/agent-isolation-regression.md` placeholder 파일 작성 (frontmatter only, 본문 `_TBD_`)
- [ ] `worktree-state-guard.md` rule 이 worktree-integration.md cross-reference + invocation pattern (Bash CLI 호출 시퀀스) 명시
- [ ] `worktree-state-guard.md` 가 paths frontmatter 보유 (예: `paths: "**/*.go,**/.claude/agents/**"` — Phase 2 가 정확한 패턴 결정)
- [ ] AC-CIAUT-014 manual replay 통과 (4회 연속 fake Agent invocation 시뮬레이션 → 1회차 alert)
- [ ] AC-CIAUT-015 manual replay 통과 (empty worktreePath fixture → suspect flag + warning)
- [ ] No release/tag automation 도입 (Wave 5 산출물 어떤 것도 `git tag` / `gh release create` / `goreleaser` 호출 안 함)
- [ ] No hardcoded paths/URLs/models (CLAUDE.local.md §14): `.moai/state/`, `.moai/reports/worktree-guard/`, untracked scope `.moai/specs/`, default exclusions list 모두 const 추출
- [ ] PR labeled with `type:feature`, `priority:P1`, `area:cli` + `area:workflow`
- [ ] Conventional Commits + 🗿 MoAI co-author trailer 모든 commit
- [ ] CHANGELOG.md 에 Wave 5 머지 entry

---

## Out-of-Scope (Wave 5 Exclusions)

- Wave 6 (T7 i18n validator) — 독립적, 별도 Wave
- Wave 7 (T8 BODP) — 독립적, 별도 Wave
- Anthropic upstream `Agent(isolation:)` 회귀 fix — claude-code-guide 가 보고만, 본 wave 는 guard layer 만 (spec.md §2 Out of Scope; strategy-wave5 §14 C-1)
- Orchestrator (Skill body) 측 invocation wiring — 본 wave 는 primitive + rule 만 (strategy-wave5 §14 C-2). `/moai run` / `/moai sync` 등 workflow body 의 실제 wiring 은 별도 SPEC.
- Untracked 파일 content snapshot — paths-only restoration; content 복원은 사용자 수동 (strategy-wave5 §14 C-3)
- Configurable divergence threshold — binary detection 만 (OQ4)
- AskUserQuestion 직접 호출 from Go — orchestrator-only HARD; Go 는 exit code + JSON 만 (strategy-wave5 §14 C-5)
- Multi-platform `git` binary 호환성 별도 검증 — Wave 1 패턴 계승 (Windows git-bash 가정)
- 16-language neutrality — Wave 5 는 Go-only orchestrator primitive (사용자 프로젝트 언어와 무관; CLAUDE.local.md §15 적용 대상 아님)
- claude-code-guide 의 자동 Anthropic console scrape — 사용자 수동 보고만; W5-T07 placeholder 는 정적 markdown
- Concurrency safety for parallel `moai worktree snapshot` invocations — 단일 invocation 가정 (orchestrator 가 sequential)
- Rollback to pre-snapshot state for tracked file deletions — `git restore` 가 처리하지 않는 케이스 는 wave 5 외
- spec.md REQ-CIAUT-035 wording 정정 ("delegate Anthropic upstream investigation" → claude-code-guide 가 NEW agent 임을 명시) — strategy-wave5 §14 C-1 follow-up commit (Wave 5 implementation 자체와는 무관)
- `internal/template/templates/.claude/agents/moai/` 디렉터리 자체 신규 생성 (현재 비어있음) — W5-T07 의 부수 작업 (디렉터리 생성 + .gitkeep 또는 첫 파일 추가)

---

## Honest Scope Concerns

1. **claude-code-guide 가 NEW 인지 EXTENSION 인지 plan.md wording 모호**: `ls .claude/agents/moai/` + `ls internal/template/templates/.claude/agents/moai/` 결과 양쪽 모두 부재. plan.md §7 W5-T07 wording ("`claude-code-guide.md` (extend with upstream investigation prompt)") 은 부정확. **NEW agent 작성 필요**. Phase 2 commit 메시지는 `feat(agents): W5-T07 add claude-code-guide` 사용; spec.md REQ-CIAUT-035 wording 정정은 Wave 5 외 follow-up. (strategy-wave5 §14 C-1)

2. **`internal/orchestrator/` 패키지 부재 — plan.md 가정 부정확**: dev project 에 `internal/orchestrator/` 디렉터리가 존재하지 않음. strategy 가 `internal/worktree/` (신규 lib) + `internal/cli/worktree/guard.go` (CLI) 패턴으로 정정 (OQ1). Phase 2 manager-tdd 위임 시 이 정정사항 명시 필요. (strategy-wave5 §7 OQ1)

3. **Orchestrator 측 wiring 본 Wave scope 외**: Go primitive + CLI 만 제공. `Skill("moai")` body 또는 `/moai run`/`/moai sync` workflow body 에서 실제 invoke 는 별도 SPEC. `worktree-state-guard.md` rule 이 invocation pattern 정의하지만 actual wiring 코드 변경은 본 wave 외. 본 wave 머지 후 즉시 활성화 안 됨 — 단계적 rollout. (strategy-wave5 §14 C-2)

4. **Untracked 파일 content snapshot 부재 → W5-T06 restore 한계**: snapshot 은 path 만 저장. `git restore` 가 untracked 파일 복원 불가 (git 에 staged/tracked 안 됨). W5-T06 stdout 에 "manual recreation required" 안내 필수 + `.moai/reports/worktree-guard/<DATE>.md` 에 untracked path list 보존하여 사용자가 외부 source (이전 commit, 다른 worktree, backup) 에서 복원하도록 가이드. (strategy-wave5 §14 C-3)

5. **AskUserQuestion 호출은 본 wave 산출물 외**: Go CLI 는 exit code + JSON report 만 출력. AskUser 는 orchestrator (Skill body) 책임. 본 wave 의 rule doc (W5-T08) 이 orchestrator 측 책임 명시하지만 Skill body 변경은 wave 외 (C-2 와 동일 follow-up). (strategy-wave5 §14 C-5)

6. **Wave 4 manager-tdd/expert-devops sub-agent 1M context inheritance error 재발 가능성**: Phase 2 위임 시 동일 error 발생 가능. Mitigation (W5-R7): 본 wave Go 코드 분량 (~600 LOC 추정, 4 신규 파일 + 1 extend) 은 main-session 직접 구현 가능 범위. Phase 2 가 sub-agent 위임 실패 시 즉시 main-session 직접 구현으로 fallback. 본 결정사항을 progress.md Phase 2 entry 에 명시.

7. **`.moai/reports/evaluator-active/` 등 정상 untracked 가 false-positive 위험**: OQ3 의 untrackedScope 가 `.moai/specs/` 로 한정 → reports/ 자동 제외. 단, future-proof 를 위해 `defaultExclusions` const 에 `.moai/reports/`, `.moai/cache/`, `.moai/logs/`, `.moai/state/` 명시. Phase 2 가 `.gitignore` 와 정합성 확인 필수.

8. **Performance — `git ls-files --others --exclude-standard .moai/specs/` 가 거대 specs/ 에서 느릴 가능성**: 현재 dev project specs/ 는 ~50 directories 로 빠름. 사용자 프로젝트가 1000+ SPEC 보유 시 저하 가능. Mitigation: scope 한정 (OQ3) + Phase 2 perf 측정. 5s 초과 시 `--max-depth` flag 또는 caching 도입 검토 (Wave 5 외 follow-up).

9. **Test design for "branch changed" case**: host project 의 git state 와 격리 필수. `t.TempDir()` + `exec.Command("git", "init", tmpDir)` + 모든 git 명령은 `-C tmpDir` flag 또는 `Dir = tmpDir` 설정. Host repo 의 HEAD 변경 절대 금지 (CI 환경에서도 동일). Phase 2 가 test fixture 작성 시 verbose error: "test must not modify host repo state".

10. **`internal/cli/worktree/root.go` 에 guard subcommand 등록 위치**: 기존 root.go 의 cobra `rootCmd.AddCommand(...)` 패턴 inspect 후 동일 패턴 적용. `internal/cli/worktree/root_test.go` (existing) 가 등록 회귀 즉시 catch. Phase 2 가 root.go read-before-edit 강제.

No hard blockers identified. Wave 5 ready for Phase 2 (manager-tdd 위임 또는 main-session 직접 구현 fallback) upon strategy + tasks approval.

---

Version: 0.1.0
Status: pending Phase 2 (manager-tdd 위임 대기, sub-agent context error 시 main-session fallback)
Last Updated: 2026-05-08
