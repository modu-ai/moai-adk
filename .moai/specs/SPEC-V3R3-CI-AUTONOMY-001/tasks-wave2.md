# Wave 2 Atomic Tasks (Phase 1 Output)

> Companion to strategy-wave2.md. SPEC-V3R3-CI-AUTONOMY-001 Wave 2 — CI Watch Loop (T2).
> Generated: 2026-05-06. Methodology: TDD. Wave Base: origin/main 0b028bfaa.

## Atomic Task Table

| Task ID | Description | REQ-CIAUT-XXX | Dependencies | Planned Files | Acceptance | Status |
|---------|-------------|---------------|--------------|---------------|------------|--------|
| W2-T01 | Skill `moai-workflow-ci-watch` SKILL.md skeleton (YAML frontmatter + 3-tier Progressive Disclosure structure) | REQ-008, REQ-013 | None | `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL.md` | Frontmatter validates; sections present; line count <= 500 | pending |
| W2-T02 | Wave 1 mirror script contract preservation smoke test | REQ-013 | W2-T03 | `scripts/ci-watch/test/run_test.sh` (test_mirror_script_intact) | `scripts/ci-mirror/run.sh` callable from Wave 2 path; silent skip exit 0 verified | pending |
| W2-T03 | `scripts/ci-watch/run.sh` gh polling loop with mock injection | REQ-008, REQ-009, REQ-010, REQ-012 | W2-T04, W2-T10 | `scripts/ci-watch/run.sh`, `scripts/ci-watch/lib/_common.sh`, `scripts/ci-watch/test/run_test.sh` | 4 scenarios pass (all-pass, required-fail, aux-only-fail, mixed-pending) | pending |
| W2-T04 | Required-checks SSoT classifier (Go + shell) | REQ-009, REQ-013 | None (Wave 1 W1-T07 SSoT exists) | `internal/ciwatch/classifier.go`, `internal/ciwatch/classifier_test.go`, `scripts/ci-watch/lib/classify.sh` | TestIsRequired_TableDriven 6 cases pass; shell helper consumes same SSoT | pending |
| W2-T05 | 30s status report formatter (pure function) | REQ-010 | None | `internal/ciwatch/handoff.go` (FormatStatusUpdate), `internal/ciwatch/handoff_test.go` | TestFormatStatusUpdate 3 scenarios; no ANSI; max 200 char/line | pending |
| W2-T06 | SKILL.md Implementation/Advanced sections (300-600 lines content) | REQ-008, REQ-010, REQ-011, REQ-012 | W2-T01, W2-T03, W2-T08 | `internal/template/templates/.claude/skills/moai-workflow-ci-watch/SKILL.md` (extend), `modules/ci-watch-protocol.md`, `modules/trigger-handoff.md` | All required sections concrete; cross-references to rule/T3 valid | pending |
| W2-T07 | `ci-watch-protocol.md` rule file with HARD invocation contract | REQ-008, REQ-011, REQ-012 | W2-T06 | `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md` | Frontmatter paths set; HARD/WARN markers; T3 handoff format documented | pending |
| W2-T08 | Ready-to-merge AskUserQuestion handoff report (CLI emits structured stdout, orchestrator consumes) | REQ-011 | W2-T05, W2-T09 | `internal/cli/pr/watch.go`, `internal/cli/pr/watch_test.go` | TestReadyToMergeFlow asserts markdown report with `(권장)` first option; CLI does not call AskUserQuestion | pending |
| W2-T09 | T3 failure metadata handoff schema (FailedCheck struct + JSON serialization) | REQ-012 | None | `internal/ciwatch/handoff.go` (FailedCheck), `internal/ciwatch/handoff_test.go` | TestNewHandoff 3 cases; JSON shape stable for Wave 3 consumer | pending |
| W2-T10 | State file (`.moai/state/ci-watch-active.flag`) + heartbeat + abort flag + 30-min wall-clock | REQ-008, REQ-010 | None | `internal/ciwatch/state.go`, `internal/ciwatch/state_test.go`, `scripts/ci-watch/lib/timeout.sh`, `internal/cli/pr/watch.go` (--abort subflag) | TestStateFile_HeartbeatStale, TestStateFile_AbortFlag, test_30min_hard_stop pass | pending |

## File Ownership Assignment (Agent Teams)

### Implementer Scope (write access)

```
internal/ciwatch/                                    # All files (state.go, classifier.go, handoff.go + tests)
internal/cli/pr/                                     # watch.go + watch_test.go
internal/config/required_checks.go                   # EXTEND only — IsRequired helper if not present
scripts/ci-watch/                                    # All shell files (run.sh, lib/*.sh)
internal/template/templates/.claude/skills/moai-workflow-ci-watch/  # SKILL.md + modules/
internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md
```

### Tester Scope (write access for tests, read-only on production code)

```
internal/ciwatch/*_test.go                           # All test files
internal/cli/pr/watch_test.go
scripts/ci-watch/test/                               # All shell test harnesses
```

### Read-Only Scope (analyst/reviewer, no write)

```
.github/required-checks.yml                          # Wave 1 SSoT — read-only consumer
internal/config/required_checks.go                   # Wave 1 loader — read-only consumer (extension allowed only via implementer scope)
scripts/ci-mirror/                                   # Wave 1 module — contract-preserving read-only
```

### Implicit Read Access (all roles)

- `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/` (spec.md, plan.md, acceptance.md, strategy-wave2.md, this file)
- `.claude/rules/moai/**` (auto-loaded rules)
- Wave 1 deliverables (Makefile, scripts/ci-mirror/, internal/cli/github_init.go, internal/config/required_checks.go)

## AC Mapping

| Wave 2 Task | Drives AC | Validation |
|-------------|-----------|------------|
| W2-T03 + W2-T08 | AC-CIAUT-004 (CI watch auto-invocation post `/moai sync`) | Integration: skill activated within 30s of `/moai sync` PR-create completion; manual + scripted |
| W2-T04 + W2-T03 | AC-CIAUT-005 (Required vs auxiliary discrimination) | Unit: classifier table-driven (6 cases) + shell mock-gh scenario with mixed required/aux fails; emits "1 required failure, 2 advisory" |
| W2-T09 | Wave 3 prereq for AC-CIAUT-006/007/008 | JSON schema fixture consumed by Wave 3 expert-debug prompt |
| W2-T01 + W2-T06 + W2-T07 | REQ-CIAUT-008 (orchestrator invokes skill on PR create) | Skill auto-load on `/moai sync` Phase 4 completion via paths frontmatter in rule file |
| W2-T10 | R-CIAUT-2 mitigation | 30-min timeout test + abort flag test verify token-budget guard |

## TRUST 5 Targets (Wave 2 SPEC-Level DoD)

| Pillar | Target | Verification |
|--------|--------|--------------|
| **Tested** | Per-package coverage >= 85% for `internal/ciwatch/`; shell test harness covers run.sh 4 scenarios | `go test -cover ./internal/ciwatch/...`; `bash scripts/ci-watch/test/run_test.sh` |
| **Readable** | Zero golangci-lint warnings; shellcheck -s sh clean on all `scripts/ci-watch/` shell files | `make ci-local`; `shellcheck -s sh scripts/ci-watch/**/*.sh` |
| **Unified** | gofmt + goimports clean; shell scripts follow Wave 1 `_common.sh` style | `make fmt-check`; manual code review |
| **Secured** | No secrets in code or templates; `gh` CLI handles all auth (no token storage); state file contains no PR body content (only metadata) | `grep -rE 'GITHUB_TOKEN\|gh_pat' internal/ciwatch/ scripts/ci-watch/` returns empty |
| **Trackable** | All commits reference SPEC-V3R3-CI-AUTONOMY-001 W2; Conventional Commits format; `🗿 MoAI <email@mo.ai.kr>` trailer | `git log --grep='SPEC-V3R3-CI-AUTONOMY-001 W2'` shows 8 commits per pacing plan |

## Per-Wave DoD Checklist

- [ ] All 10 W2 tasks complete (table above)
- [ ] Template-First mirror verified: every new `.claude/` file has counterpart under `internal/template/templates/.claude/`
- [ ] `make build` regenerates `internal/template/embedded.go` cleanly
- [ ] `make ci-local` passes (Wave 1 framework consumes Wave 2 changes without regression)
- [ ] `go test ./...` and `go test -race ./internal/ciwatch/... ./internal/cli/pr/...` pass
- [ ] Wave 2-specific acceptance: AC-CIAUT-004, AC-CIAUT-005 manually validated against a live PR (smoke test)
- [ ] No hardcoded URLs/models/env keys in `internal/ciwatch/` or `scripts/ci-watch/`
- [ ] PR labeled with `type:feature`, `priority:P0`, `area:ci`, `area:templates`
- [ ] Conventional Commits + MoAI co-author trailer on all 8 commits

## Out-of-Scope (Wave 2 Exclusions)

- Auto-fix loop integration with `expert-debug` (Wave 3 = T3)
- Auxiliary workflow disable/cleanup (Wave 4 = T4)
- Worktree state guard (Wave 5 = T6)
- i18n validator (Wave 6 = T7)
- BODP (Wave 7 = T8)
- Multi-process concurrent watch (deferred follow-up SPEC; single-watch-per-repo model is Wave 2 contract)
- gh CLI < 2.50 deep compatibility (heuristic fallback only; full support out of scope)
- Cross-platform PowerShell variant of `scripts/ci-watch/run.sh` (Windows users use git-bash, same as Wave 1)

## Honest Scope Concerns

1. **gh CLI version drift over time**: `--json` field set may change in future gh releases. Mitigated by `moai doctor` version pin and fallback heuristics. Recommended to add `gh --version` capture to state file for future compatibility audits.
2. **State file race on concurrent invocation**: heartbeat-based staleness reclaim is sufficient for serial orchestrator use, but two simultaneous `moai pr watch` invocations could overlap during the 90s heartbeat window. Acceptable for Wave 2; future SPEC may add fcntl-based locking.
3. **30-min hard timeout**: empirically chosen. If post-merge data shows CI runs regularly exceeding this, propose config knob via follow-up SPEC; do not extend default in Wave 2.

No hard blockers. Wave 2 is ready for Phase 2 (manager-tdd) delegation upon strategy + tasks approval.

---

Version: 0.1.0
Status: pending implementation
Last Updated: 2026-05-06
