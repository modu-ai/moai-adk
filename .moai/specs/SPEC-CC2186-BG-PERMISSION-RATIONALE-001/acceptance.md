# Acceptance Criteria — SPEC-CC2186-BG-PERMISSION-RATIONALE-001

> All ACs are grep-verifiable or test-verifiable per `.claude/rules/moai/core/verification-claim-integrity.md`. Each PASS in run/sync MUST cite the actually-observed command output (recorded in progress.md §E.2/§E.3). Line numbers may shift; verify by content token, not by line.

## §D. AC Matrix

| AC | REQ | Verification | PASS condition |
|----|-----|--------------|----------------|
| AC-BGR-001 | REQ-BGR-001 | grep | "cannot interact with the user" removed from agent-common-protocol § Background Agent Execution (live + mirror) |
| AC-BGR-002 | REQ-BGR-002 | grep | "auto-deny" descriptor corrected in CLAUDE.md §14 + zone-registry CONST-V3R2-020 (live + mirror) |
| AC-BGR-003 | REQ-BGR-003 | grep | `run_in_background: false` directive still present in all corrected surfaces (conclusion retained) |
| AC-BGR-004 | REQ-BGR-004 | grep | allowlist-non-inheritance sentence retained in agent-common-protocol (live + mirror) |
| AC-BGR-005 | REQ-BGR-005 | diff | per-locus DIRECT live-vs-mirror corrected-region diff exits 0 for all 3 loci (see §D.1 — no Go test backs these paths; see D2 note) |
| AC-BGR-006 | REQ-BGR-006 | test | `TestTemplateNeutralityAudit` + `TestTemplateNoInternalContentLeak` pass (no internal-content leak in template prose) |
| AC-BGR-007 | REQ-BGR-007 | build | `embedded.go` regenerated via `make build`; reflects corrected mirror; `go build ./...` succeeds |
| AC-BGR-008 | REQ-BGR-008 | evidence | run-phase doc re-confirmation evidence (2.1.186 surface verbatim from sub-agents doc) recorded in progress.md §E.2 before [HARD] rewrite |
| AC-BGR-009 | REQ (Exclusions) | grep | CONST-V3R2-044 clause UNCHANGED in both trees (scope discipline) |
| AC-BGR-010 | REQ-BGR-007 | test | full suite `go test ./...` green (no cascading failure) |

## §D.1 Grep-verifiable AC commands (canonical)

The run-phase agent runs these from repo root and records verbatim output in progress.md §E.2.

### AC-BGR-001 — primary drift clause removed (STRICT == 0)

```bash
grep -rc 'because they cannot interact with the user' \
  .claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
# Expected: both files report 0
```

### AC-BGR-002 — auto-deny descriptor corrected (STRICT == 0 for the old descriptor)

```bash
grep -rn 'auto-deny' CLAUDE.md internal/template/templates/CLAUDE.md \
  .claude/rules/moai/core/zone-registry.md \
  internal/template/templates/.claude/rules/moai/core/zone-registry.md
# Expected: 0 matches in CLAUDE.md §14 + CONST-V3R2-020 clause
# (CONST-V3R2-044 does NOT contain "auto-deny" — see AC-BGR-009)
```

### AC-BGR-003 — conclusion retained (STRICT >= 1 per surface family)

```bash
grep -rn 'run_in_background: false' CLAUDE.md internal/template/templates/CLAUDE.md
grep -rn 'run_in_background: false\|MUST NOT perform Write/Edit\|MUST NOT.*Write/Edit' \
  .claude/rules/moai/core/zone-registry.md \
  internal/template/templates/.claude/rules/moai/core/zone-registry.md
# Expected: the run_in_background:false directive (or the MUST NOT Write/Edit conclusion)
#           still present in each corrected surface — the restriction is NOT relaxed
```

### AC-BGR-004 — allowlist survivor retained (STRICT >= 1)

```bash
grep -rc 'does not fully inherit the parent session' \
  .claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
# Expected: both files report >= 1 (the allowlist-non-inheritance justification survives)
```

### AC-BGR-005 — live/mirror parity (DIRECT per-locus corrected-region diff)

> **D2 design note (plan-auditor SHOULD-FIX)**: there is NO existing Go test backing this parity guarantee. The real mirror-drift test `TestRuleTemplateMirrorDrift` (`internal/template/rule_template_mirror_test.go`) covers NONE of the 3 loci — CLAUDE.md and zone-registry.md are absent from its `workflowOptMirroredPaths` allowlist, and agent-common-protocol.md is explicitly in its leak-test-only list (the live and mirror agent-common-protocol.md are DIVERGENT overall because the live copy carries ~17 internal-content tokens stripped from the mirror). Therefore AC-BGR-005 MUST use a DIRECT per-locus `diff` of the corrected region only — NOT a whole-file diff (which would FAIL on agent-common-protocol.md) and NOT a Go test (which does not cover these paths). Each locus extracts the corrected sentence/clause from both trees and compares byte-for-byte; the AC FAILS if any of the 3 loci diverge live↔mirror after correction.

```bash
# Locus 1 — CLAUDE.md §14 "Background Agent Write Restriction" bullet (single corrected line)
diff <(grep -F 'Background Agent Write Restriction' CLAUDE.md) \
     <(grep -F 'Background Agent Write Restriction' internal/template/templates/CLAUDE.md) \
  && echo "LOCUS-1 PARITY OK (exit 0)"

# Locus 2 — agent-common-protocol § Background Agent Execution corrected paragraph.
#   Scope the diff to the corrected region ONLY (the §-heading → "Rules for agent spawning"
#   block). A whole-file diff is INVALID here: live and mirror diverge overall (leak-test-only,
#   ~17 internal tokens). The corrected paragraph itself must be byte-identical across trees.
diff <(sed -n '/## Background Agent Execution/,/Rules for agent spawning/p' .claude/rules/moai/core/agent-common-protocol.md) \
     <(sed -n '/## Background Agent Execution/,/Rules for agent spawning/p' internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md) \
  && echo "LOCUS-2 PARITY OK (exit 0)"

# Locus 3 — zone-registry CONST-V3R2-020 corrected clause.
#   The clause: line is at +5 from the id: line, so -A1 MISSES it (it stops at +1). Use -A6
#   to capture through the canary_gate line, OR anchor directly on the clause: text. Both forms
#   below are well-formed; the second is line-offset-independent (preferred).
diff <(grep -A6 -F 'id: CONST-V3R2-020' .claude/rules/moai/core/zone-registry.md) \
     <(grep -A6 -F 'id: CONST-V3R2-020' internal/template/templates/.claude/rules/moai/core/zone-registry.md) \
  && echo "LOCUS-3 PARITY OK (exit 0)"
# Offset-independent alternative for Locus 3 (anchors on the corrected clause text directly):
diff <(grep -F 'clause: "Background subagents' .claude/rules/moai/core/zone-registry.md | grep -F 'run_in_background') \
     <(grep -F 'clause: "Background subagents' internal/template/templates/.claude/rules/moai/core/zone-registry.md | grep -F 'run_in_background')

# Expected: each locus diff exits 0 (no output) → AC-BGR-005 PASS.
# Any non-empty diff for any locus → AC-BGR-005 FAIL.
```

> Note on AC-BGR-005 tooling: `sed`/`diff`/`grep` are used here as read-only verification (not as content-mutation tools). The Edit-vs-sed coding-standard governs *mutation*; read-only parity diffing via `diff`/`sed -n`/`grep` is acceptable for verification evidence.

### AC-BGR-009 — out-of-scope CONST-V3R2-044 unchanged (STRICT, exact clause preserved)

```bash
grep -c 'Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations' \
  .claude/rules/moai/core/zone-registry.md \
  internal/template/templates/.claude/rules/moai/core/zone-registry.md
# Expected: each file reports 1 (the pure-conclusion clause is preserved, NOT edited)
```

## §D.2 Test / build-verifiable AC

### AC-BGR-006 + AC-BGR-007 + AC-BGR-010

```bash
make build                                    # regenerate embedded.go (AC-BGR-007)
go build ./...                                # embedded.go compiles (AC-BGR-007)
# Neutrality (AC-BGR-006): these real tests DO cover the template prose. NOTE: TestEmbeddedMirror
# does NOT exist in the repo and is intentionally NOT invoked — AC-BGR-005 parity is verified by
# the direct per-locus diff in §D.1 instead (see D2 design note).
go test ./internal/template/... -run 'TestTemplateNeutralityAudit|TestTemplateNoInternalContentLeak'   # AC-BGR-006
go test ./...                                 # full suite green (AC-BGR-010)
```

## §D.3 Edge cases

- **EC-1 — "auto-deny" appears elsewhere legitimately**: confirm the AC-BGR-002 grep scope is limited to CLAUDE.md §14 + CONST-V3R2-020; if "auto-deny" survives in an unrelated context, scope the grep to the target loci (do not over-delete).
- **EC-2 — CONST-V3R2-044 accidentally matched by an auto-deny grep**: CONST-V3R2-044 uses "MUST NOT perform Write/Edit" (no "auto-deny" token), so it should not match AC-BGR-002. AC-BGR-009 independently confirms it is unchanged.
- **EC-3 — mirror byte-drift from invisible whitespace**: AC-BGR-005 per-locus DIRECT diff catches trailing-whitespace divergence in the corrected region. There is NO Go-test backstop for these 3 loci (`TestRuleTemplateMirrorDrift` covers none of them); the per-locus diff IS the parity authority — run it for all 3 loci, do not assume a CI test guards it.
- **EC-4 — make build dirties unrelated embedded content**: confirm `git diff --stat` after `make build` shows only `embedded.go` (expected) + the 6 source surfaces; no unrelated template churn.

## §D.4 Definition of Done

- [ ] All 10 ACs PASS with verbatim grep/test output recorded in progress.md §E.2 + §E.3
- [ ] Primary drift clause ("cannot interact with the user") == 0 across both agent-common-protocol copies
- [ ] "auto-deny" descriptor corrected in CLAUDE.md §14 + CONST-V3R2-020 (both trees)
- [ ] `run_in_background: false` conclusion retained in every corrected surface (restriction NOT relaxed)
- [ ] allowlist-non-inheritance survivor retained
- [ ] CONST-V3R2-044 unchanged (scope discipline verified)
- [ ] live == mirror for each locus (byte-identical prose)
- [ ] `make build` run; `embedded.go` reflects corrected mirror; `go build ./...` + `go test ./...` green
- [ ] Template neutrality test passes (no internal-content leak)
- [ ] Run-phase doc re-confirmation evidence (2.1.186 surface from official sub-agents doc) recorded before the [HARD] rewrite

## §D.5 Quality gate criteria

- LSP/build: zero errors after `make build` (`go build ./...` green).
- Test: `go test ./...` green (no cascading failure introduced).
- Neutrality: `TestTemplateNeutralityAudit` + `TestTemplateNoInternalContentLeak` green.
- Mirror: per-locus DIRECT live-vs-mirror diff (§D.1 AC-BGR-005) exits 0 for all 3 loci. NOTE: no Go test backs these 3 paths (`TestRuleTemplateMirrorDrift` covers none of them; `TestEmbeddedMirror` does not exist) — the per-locus diff is the parity authority for this SPEC.
