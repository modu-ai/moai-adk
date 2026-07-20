# research.md — SPEC-INTERNAL-TEST-004

> Drift-cause analysis with `UPDATE_GOLDEN=1` + `git diff` byte-level evidence.
> Scope-boundary decision on the cache-hit emoji change (💾→♻️).
> Approach selection with rationale.
> This file corrects an imprecise characterization inherited from SPEC-INTERNAL-TEST-003 progress.md.

---

## §1. Baseline State (observed, not inferred)

| Signal | Value | Evidence |
|--------|-------|----------|
| HEAD (actual) | `bd97306af chore(SPEC-TEMPLATE-RULES-CLEANUP-001): 3-phase close` | `git log --oneline -1` |
| `origin/main...HEAD` | `0 0` (synced) | `git rev-list --count --left-right origin/main...HEAD` |
| `pkg/version/version.go` Version (HEAD-committed) | `v3.0.0-rc9` | `git show HEAD:pkg/version/version.go` + `grep` |
| version bump commit | `9edb72af5 chore(version): bump to v3.0.0-rc9` | `git log --oneline -- pkg/version/version.go` |
| Working-tree version.go modification | NONE (not in `git status`) | `git status --short pkg/version/` empty |
| Golden testdata committed state | expects `rc7` | `git show HEAD:internal/cli/testdata/doctor-light.golden` |

**Context-drift note (does not affect conclusions)**: the orchestrator briefing stated HEAD = `280c9dd71` and working-tree version.go = `rc8`. The actual HEAD is `bd97306af` and version.go is cleanly committed at `rc9` (no working-tree drift). This is a stale-context observation, not a research blocker — `origin/main...HEAD = 0 0` confirms the baseline is clean and synced.

---

## §2. Drift-Cause Analysis — the decisive `UPDATE_GOLDEN=1` + `git diff` evidence

### §2.1 FAIL detail (from `.moai/state/verify/538fe6ae/test-004-golden-fail-detail.log`)

6 golden tests FAIL in `internal/cli`:

| Test | File:line | Mismatch sections |
|------|-----------|-------------------|
| `TestDoctor_Current_Light` | `doctor_golden_test.go:91` | MoAI-ADK / System / Workspace (version line) |
| `TestDoctor_Current_Dark` | `doctor_golden_test.go:91` | (same pattern) |
| `TestDoctor_NoColor` | `doctor_golden_test.go:91` | (same pattern) |
| `TestStatus_Current_Light` | `status_golden_test.go:142` | Project / Configuration (version line) |
| `TestStatus_Current_Dark` | `status_golden_test.go:142` | (same pattern) |
| `TestStatus_NoColor` | `status_golden_test.go:142` | (same pattern) |

### §2.2 The `UPDATE_GOLDEN=1` regenerate run

Command (analysis-only; testdata reverted afterward per PRESERVE discipline):
```
UPDATE_GOLDEN=1 go test ./internal/cli/ -run 'TestDoctor_Current_Light|TestStatus_Current_Light|TestDoctor_Current_Dark|TestStatus_Current_Dark|TestDoctor_NoColor|TestStatus_NoColor' -count=1
```
Result: `ok github.com/modu-ai/moai-adk/internal/cli 0.563s` (all 6 PASS after regen).

### §2.3 Byte-level `git diff internal/cli/testdata/` — the decisive evidence

Diff stat: 6 files changed, 6 insertions(+), 6 deletions(-) — exactly 1 line per file.

Unique changed-line patterns (exhaustive — these are the ONLY byte changes across all 6 files):

```
-│  ✓  MoAI Version  moai-adk v3.0.0-rc7  ...    (doctor goldens)
+│  ✓  MoAI Version  moai-adk v3.0.0-rc9  ...

-│  ADK       moai-adk v3.0.0-rc7  ...            (status goldens)
+│  ADK       moai-adk v3.0.0-rc9  ...
```

### §2.4 Drift-cause conclusion

**The golden mismatch is caused SOLELY by the version string `rc7 → rc9`.**

- `pkg/version/version.go` was bumped from `rc7` to `rc9` (committed via `9edb72af5`; `rc8` was an intermediate release that is no longer the HEAD value).
- The 6 golden testdata files were last regenerated at `rc7` (by SPEC-INTERNAL-TEST-002 M1 `ffea91710`) and were NOT regenerated when version.go moved to `rc8` then `rc9`.
- The code output (rc9) is CORRECT — the goldens (rc7) are STALE.
- **No section-rendering logic changed.** The doctor sections (System / MoAI-ADK / Workspace) and status sections (Project / Configuration) are structurally identical between got/want — only the version string differs.
- **No cache-hit emoji appears in the golden output.** The golden test fixtures set `cacheStrategy.enabled: false`, so the cache-hit segment never renders. The 💾→♻️ change is invisible to these tests.

### §2.5 Correction of TEST-003's characterization

TEST-003 progress.md §E.2 AC-006 row characterized the debt as *"statusline golden drift ... vs uncommitted `internal/statusline/renderer.go` working-tree changes"* and named the remediation owner as *"statusline golden regeneration"*.

This characterization was **imprecise**. The actual root cause is a **version-bump golden drift** in `internal/cli/`, not a statusline drift. The `internal/statusline/renderer.go` working-tree change (💾→♻️) is in a DIFFERENT package and does NOT affect the `internal/cli/` golden tests at all (the goldens do not render the statusline cache-hit segment). TEST-004 corrects this: the remediation is golden testdata regeneration for the version bump, not statusline logic changes.

---

## §3. Cache-Hit Emoji Scope Decision (research Q2 + Q4)

### §3.1 Evidence

| Signal | Value | Source |
|--------|-------|--------|
| HEAD-committed `renderer.go:312` | `return fmt.Sprintf("💾 %d%%", pct)` | `git show HEAD:internal/statusline/renderer.go` |
| Working-tree `renderer.go:312` | `return fmt.Sprintf("♻️ %d%%", pct)` | `git diff internal/statusline/renderer.go` |
| Ancestor commit | `22220186c feat(SPEC-WEB-CONSOLE-011): M6 statusline cache_hit 노출 fan-out + segment-list SSOT` | `git log --oneline 22220186c -1` |
| `cache_hit_test.go` working-tree edit | 4 line changes 💾→♻️ (test expectation sync) | `git diff internal/statusline/cache_hit_test.go` |

### §3.2 Scope decision

**The cache-hit emoji change (💾→♻️) belongs to SPEC-WEB-CONSOLE-011, NOT TEST-004.**

Rationale:
1. The ancestor commit `22220186c` explicitly scopes the cache-hit fan-out to WEB-CONSOLE-011 M6.
2. The working-tree edits are the continuation of that M6 work that was not committed with the original commit.
3. The emoji change is in `internal/statusline/` — a package that TEST-004 does NOT touch.
4. `internal/statusline/` tests PASS in BOTH states: HEAD-committed (💾, both renderer + test use 💾) AND working-tree (♻️, both synced). The package is internally consistent regardless of commit state.
5. Whole-repo `go test ./...` exit 0 does NOT depend on the emoji commit decision — confirmed by running the full suite (see §4).

### §3.3 TEST-004 action on the emoji edits

**PRESERVE — do not commit, do not revert.** The working-tree edits (renderer.go + cache_hit_test.go) remain as-is for WEB-CONSOLE-011 to claim. TEST-004 touches ONLY `internal/cli/testdata/*.golden`.

---

## §4. Whole-Repo Exit 0 Verification (analysis-only, goldens reverted after)

Command (with goldens temporarily regenerated for analysis):
```
go test ./... > /tmp/moai-test004-research/wholerepo-current.log 2>&1; echo "exit=$?"
```
Result: `exit=0`, 93 `ok` packages, 0 FAIL.

**This confirms that regenerating the 6 goldens alone is SUFFICIENT to achieve whole-repo exit 0.** No other package has a failing test. The cache-hit emoji working-tree state (uncommitted ♻️) does not block green because `internal/statusline/` is internally consistent.

Evidence persisted at: `/tmp/moai-test004-research/wholerepo-current.log` (analysis-only; not a run-phase artifact).

---

## §5. Approach Selection

### §5.1 Decision: golden regenerate (code is correct, goldens are stale)

Selected over alternatives because the `UPDATE_GOLDEN=1` + `git diff` evidence (§2.3) proves the code output is correct and only the golden version string is stale. There is NO code regression to fix.

| Approach | Verdict | Reason |
|----------|---------|--------|
| **Golden regenerate** (`UPDATE_GOLDEN=1`) | **SELECTED** | git diff proves only the version string line changes; code is correct |
| Code fix | Rejected | No code regression exists — version.go rc9 is the intended release version |
| Both | Rejected | No code side to fix; golden regen alone is necessary and sufficient |

### §5.2 Risk assessment

- **Low risk**: golden regeneration is a mechanical, well-understood operation (`UPDATE_GOLDEN=1` is the project's documented golden-refresh mechanism).
- **Single concern**: 6 testdata files, 1 line each, identical transform (rc7→rc9 version string).
- **No logic change**: no Go source modified, no behavior change, no API change.
- **Reversible**: `git checkout internal/cli/testdata/` fully reverts if needed.

---

## §6. PRESERVE List (working-tree paths TEST-004 MUST NOT touch)

| Path | Owner | Why preserved |
|------|-------|---------------|
| `internal/statusline/renderer.go` (working-tree ♻️) | WEB-CONSOLE-011 | cache-hit emoji change, ancestor `22220186c` |
| `internal/statusline/cache_hit_test.go` (working-tree ♻️) | WEB-CONSOLE-011 | test sync for the emoji change |
| `.claude/agents/moai/manager-docs.md` (modified) | (other work) | unrelated |
| `.claude/agents/moai/manager-git.md` (modified) | (other work) | unrelated |
| `.moai/config/sections/llm.yaml` (modified) | (other work) | unrelated |
| `.moai/config/sections/workflow.yaml` (modified) | (other work) | unrelated |
| `.moai/specs/SPEC-OBSERVE-HYGIENE-001/progress.md` (modified) | (other work) | unrelated |
| `internal/template/catalog.yaml` (modified) | (other work) | unrelated |
| `internal/template/templates/.claude/agents/moai/manager-{docs,git}.md` (modified) | (other work) | unrelated |
| `internal/template/templates/.moai/config/sections/llm.yaml` (modified) | (other work) | unrelated |
| All `??` untracked paths (moai-easy.md, rules/.moai/, etc.) | (various) | unrelated |
| `pkg/version/version.go` | release process | rc9 is the committed release version — do NOT bump or revert |

**TEST-004 touches ONLY**: `.moai/specs/SPEC-INTERNAL-TEST-004/` (this SPEC's artifacts) during plan-phase, and `internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden` (6 files) during run-phase.

---

## §7. Cross-References

- SPEC-INTERNAL-TEST-003 `progress.md` §E.2 AC-006 row — debt transfer provenance
- SPEC-INTERNAL-TEST-003 `.moai/state/verify/538fe6ae/test-003-ac006-wholerepo.log` — original whole-repo FAIL evidence
- SPEC-INTERNAL-ARCH-001 plan-audit (memory `project_internal_arch_001_plan_entry`) — whole-repo-green M0 precondition consumer
- SPEC-WEB-CONSOLE-011 ancestor commit `22220186c` — cache-hit emoji scope owner
- SPEC-INTERNAL-TEST-002 M1 `ffea91710` — last golden regeneration (rc6→rc7); the rc7→rc9 gap is the drift window
