# Acceptance Criteria — SPEC-STEERING-ALIGN-RULE-SCOPING-001

All criteria are GEARS-format and verified by a re-runnable command. Each command is read-only or a `git diff` inspection; reproduce against the post-run tree.

## D. AC Matrix

| AC ID | REQ | Severity | Verification |
|-------|-----|----------|--------------|
| AC-SARS-001 | REQ-SARS-001/006/009 | MUST | per-tree always-loaded drop: LIVE 15 → 11, TEMPLATE 13 → 10 |
| AC-SARS-002 | REQ-SARS-002/003/004/005 | MUST | each Class-A rule loads on matching file touch (`**/`-prefixed glob sanity) |
| AC-SARS-003 | REQ-SARS-008 | MUST | template + live `paths:` identical — MIRRORED targets only, both-files-exist guard (D4) |
| AC-SARS-004 | REQ-SARS-007 | MUST | no rule BODY content changed (frontmatter-only diff) |
| AC-SARS-005 | REQ-SARS-006 | MUST | NOTICE.md excluded from always-load, file retained on disk (both trees) |
| AC-SARS-006 | REQ-SARS-009 | SHOULD | always-on byte-sum reduced per tree (LIVE + TEMPLATE) |
| AC-SARS-007 | REQ-SARS-010 | MUST | no Class-B/Class-C rule was scoped (honesty guardrail) |
| AC-SARS-008 | REQ-SARS-008b | MUST | `lifecycle-sync-gate.md` LIVE-ONLY: scoped in live, ABSENT from template, NOT parity-checked (D1) |

---

## D.1 Detailed Given-When-Then scenarios

### AC-SARS-001 — Per-tree always-loaded count drop (LIVE 15 → 11, TEMPLATE 13 → 10)

> **When** the always-loaded rule set of EACH tree is enumerated after the change, the LIVE count **shall** be exactly 11 (from baseline 15) AND the TEMPLATE count **shall** be exactly 10 (from baseline 13). The two trees drop by different amounts because the LIVE-ONLY target `lifecycle-sync-gate.md` is not present in the template tree (D1).

```bash
# LIVE tree — re-runnable. Counts rule files lacking a `paths:` line (= always-loaded).
for f in $(find .claude/rules/moai -name '*.md'); do grep -q '^paths:' "$f" || echo "$f"; done | wc -l
# Expected (post-run): 11   |   Baseline (pre-run, 2026-06-22): 15

# TEMPLATE tree — re-runnable.
for f in $(find internal/template/templates/.claude/rules -name '*.md'); do grep -q '^paths:' "$f" || echo "$f"; done | wc -l
# Expected (post-run): 10   |   Baseline (pre-run, 2026-06-22): 13
```

The 4 files that MUST move out of the LIVE always-loaded set: `development/hook-independence.md`, `development/prompting-best-practices.md`, `workflow/lifecycle-sync-gate.md`, `NOTICE.md`. The 3 files that MUST move out of the TEMPLATE always-loaded set: `development/hook-independence.md`, `development/prompting-best-practices.md`, `NOTICE.md` (NOT lifecycle-sync-gate.md — it is absent from the template tree, so it was never in the template always-loaded set).

### AC-SARS-002 — Class-A rules still load on a matching file touch (behavior preserved)

> **When** Claude touches a file matching a newly-scoped Class-A rule's `paths` glob, the rule **shall** still be selected for loading (relevance preserved — only always-on residency removed).

```bash
# Verify each Class-A rule now carries the EXACT precedent glob (glob sanity = load-on-touch preserved).
grep -H '^paths:' .claude/rules/moai/development/hook-independence.md
# Expected: paths: "**/.claude/hooks/**"
grep -H '^paths:' .claude/rules/moai/development/prompting-best-practices.md
# Expected: paths: "**/.claude/agents/**,**/.claude/skills/**"
grep -H '^paths:' .claude/rules/moai/workflow/lifecycle-sync-gate.md
# Expected: paths: "**/internal/spec/**,**/.moai/specs/**"   (D5: **/-prefixed)
```

Given-When-Then (illustrative, runtime-level): GIVEN `hook-independence.md` is scoped to `**/.claude/hooks/**`; WHEN a session edits `.claude/hooks/moai/handle-session-start.sh`; THEN the hook-independence doctrine loads (matching the 46 existing scoped rules' documented `paths:` behavior).

### AC-SARS-003 — Template + live parity (MIRRORED targets only, both-files-exist guard)

> **Where** an edit target is MIRRORED (template mirror PRESENT — the 3 MIRRORED targets, NOT lifecycle-sync-gate.md), the `paths:` frontmatter **shall** be byte-identical between the template SSOT tree and the live deployed tree. The check applies ONLY to MIRRORED targets and MUST assert both files exist before diffing (D4 — a missing path must FAIL the AC, not false-pass on empty-string equality).

```bash
# D4: assert BOTH files exist before diffing the `paths:` line; scoped to the 3 MIRRORED targets only.
# lifecycle-sync-gate.md is EXCLUDED here (it is LIVE-ONLY — covered by AC-SARS-008).
fail=0
for r in development/hook-independence.md development/prompting-best-practices.md NOTICE.md; do
  tf="internal/template/templates/.claude/rules/moai/$r"
  lf=".claude/rules/moai/$r"
  if [ ! -f "$tf" ]; then echo "PARITY FAIL: $r — template file MISSING ($tf)"; fail=1; continue; fi
  if [ ! -f "$lf" ]; then echo "PARITY FAIL: $r — live file MISSING ($lf)"; fail=1; continue; fi
  t=$(grep '^paths:' "$tf"); l=$(grep '^paths:' "$lf")
  # guard against false-pass on empty-string equality ('' = ''): require non-empty paths line
  if [ -z "$t" ] || [ -z "$l" ]; then echo "PARITY FAIL: $r — empty paths: line (t='$t' l='$l')"; fail=1; continue; fi
  if [ "$t" = "$l" ]; then echo "PARITY OK: $r"; else echo "PARITY FAIL: $r ('$t' != '$l')"; fail=1; fi
done
[ "$fail" -eq 0 ] && echo "ALL MIRRORED PARITY OK" || echo "MIRRORED PARITY FAILED"
# Expected: 3× "PARITY OK" + "ALL MIRRORED PARITY OK"
```

### AC-SARS-004 — Frontmatter-only (no body content changed)

> The change **shall not** modify any rule BODY content; the diff for each changed rule **shall** consist solely of an inserted `paths:` frontmatter block.

```bash
# Each changed rule's diff vs its pre-change state must touch ONLY the frontmatter (top of file).
# Inspect: every added line is inside a `---` ... `---` block or the `paths:` line itself.
git diff --unified=0 -- \
  internal/template/templates/.claude/rules/moai/development/hook-independence.md \
  internal/template/templates/.claude/rules/moai/development/prompting-best-practices.md \
  internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md \
  internal/template/templates/.claude/rules/moai/NOTICE.md \
  .claude/rules/moai/development/hook-independence.md \
  .claude/rules/moai/development/prompting-best-practices.md \
  .claude/rules/moai/workflow/lifecycle-sync-gate.md \
  .claude/rules/moai/NOTICE.md
# Expected: only `+---`, `+paths: "..."`, `+---` additions; ZERO body-line deletions or modifications.
```

### AC-SARS-005 — NOTICE.md excluded from always-load, retained on disk

> The `NOTICE.md` legal-attribution rule **shall** no longer be always-loaded, **and** the file **shall** remain present on disk in both trees.

```bash
# (a) excluded from always-load: NOTICE.md now carries a paths: line
grep -q '^paths:' .claude/rules/moai/NOTICE.md && echo "EXCLUDED (live)" || echo "STILL ALWAYS-LOADED (live)"
grep -q '^paths:' internal/template/templates/.claude/rules/moai/NOTICE.md && echo "EXCLUDED (template)" || echo "STILL ALWAYS-LOADED (template)"
# Expected: "EXCLUDED (live)" + "EXCLUDED (template)"

# (b) retained on disk
ls -la .claude/rules/moai/NOTICE.md internal/template/templates/.claude/rules/moai/NOTICE.md
# Expected: both files present (legal attribution preserved)
```

### AC-SARS-006 — Always-on byte-sum reduced per tree (SHOULD; proxy for token reduction)

> **When** the byte-sum of the remaining always-loaded set is measured after the change, it **shall** be strictly less than the pre-change byte-sum for BOTH trees (the always-on context budget shrinks in each).

```bash
# LIVE tree byte-sum of currently-always-loaded rules.
total=0
for f in $(find .claude/rules/moai -name '*.md'); do grep -q '^paths:' "$f" || total=$((total + $(wc -c < "$f"))); done
echo "LIVE always-loaded byte-sum: $total"
# Baseline (pre-run, 2026-06-22): 211495 bytes (15 files)
# Expected (post-run): strictly < 211495 (4 live files removed: NOTICE ~9580 + hook-indep ~14313
#                       + prompting ~9180 + lifecycle-sync-gate ~18661 ≈ 51734 removed → ~159761)

# TEMPLATE tree byte-sum.
ttotal=0
for f in $(find internal/template/templates/.claude/rules -name '*.md'); do grep -q '^paths:' "$f" || ttotal=$((ttotal + $(wc -c < "$f"))); done
echo "TEMPLATE always-loaded byte-sum: $ttotal"
# Baseline (pre-run, 2026-06-22): 156308 bytes (13 files)
# Expected (post-run): strictly < 156308 (3 template files removed: NOTICE 4639 + hook-indep 14313
#                       + prompting 9180 ≈ 28132 removed → ~128176; lifecycle-sync-gate not in template)
```

> Note (verification-claim-integrity): the byte-sum is a PROXY for the token reduction the best-practices per-line test targets; it is not an exact tokenizer count. The SHOULD severity reflects this. The MUST signal is the per-tree count drop (AC-SARS-001) and parity (AC-SARS-003).

### AC-SARS-007 — Honesty guardrail: no Class-B/Class-C rule scoped

> The change **shall not** add a `paths:` field to any Class-B or Class-C rule.

```bash
# The 11 rules that MUST remain always-loaded (no paths: line):
for f in \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/askuser-protocol.md \
  .claude/rules/moai/core/glm-web-tooling.md \
  .claude/rules/moai/core/verification-claim-integrity.md \
  .claude/rules/moai/development/sprint-round-naming.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/dynamic-workflows.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/runtime-recovery-doctrine.md \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/rules/moai/workflow/verification-batch-pattern.md ; do
  grep -q '^paths:' "$f" && echo "VIOLATION (scoped): $f" || echo "ok (still always-loaded): $f"
done
# Expected: 11× "ok (still always-loaded)" — ZERO "VIOLATION"
```

### AC-SARS-008 — `lifecycle-sync-gate.md` is LIVE-ONLY (D1)

> The LIVE-ONLY target `lifecycle-sync-gate.md` **shall** be scoped in the live tree AND **shall** remain ABSENT from the template tree (no fabricated template edit), and **shall not** be subject to the MIRRORED parity check (AC-SARS-003).

```bash
# (a) live file carries the paths: line (scoped)
grep -q '^paths:' .claude/rules/moai/workflow/lifecycle-sync-gate.md \
  && echo "LIVE SCOPED ok" || echo "LIVE NOT SCOPED — FAIL"
grep -H '^paths:' .claude/rules/moai/workflow/lifecycle-sync-gate.md
# Expected: "LIVE SCOPED ok" + paths: "**/internal/spec/**,**/.moai/specs/**"

# (b) template file MUST remain absent (no fabricated template edit — AP-7)
[ -f internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md ] \
  && echo "TEMPLATE PRESENT — FAIL (stray file created)" || echo "TEMPLATE ABSENT ok (live-only preserved)"
# Expected: "TEMPLATE ABSENT ok (live-only preserved)"
```

This AC is the D1 guard: it confirms the run-phase agent edited ONLY the live file and did NOT create a stray template mirror that does not belong in the distribution.

---

## D.2 Edge cases

- **EC-1**: A target already having frontmatter — none of the 4 currently does (verified), so each edit is a clean top-of-file insertion. If a future rebase introduces frontmatter, the run-phase agent appends `paths:` as a new field rather than duplicating the `---` fence.
- **EC-2**: `make build` re-embed surfaces unrelated template drift — the run-phase agent scopes the commit to only the MIRRORED targets' frontmatter; any unrelated drift is reported, not absorbed.
- **EC-3**: NOTICE.md body divergence between trees — the parity check (AC-SARS-003) compares ONLY the `paths:` line, not the divergent body, so divergence does not fail parity.
- **EC-4 (D1)**: the LIVE-ONLY target `lifecycle-sync-gate.md` has no template mirror — the parity loop (AC-SARS-003) deliberately EXCLUDES it; AC-SARS-008 covers it separately. A parity check that included it would false-fail (template file missing) — the D4 both-files-exist guard would correctly FAIL it, which is why it is scoped out of AC-SARS-003 entirely rather than relying on the guard.

## D.3 Definition of Done

- [ ] AC-SARS-001 PASS (LIVE 15 → 11 AND TEMPLATE 13 → 10).
- [ ] AC-SARS-002 PASS (3 Class-A globs present, `**/`-prefixed, mirror precedent).
- [ ] AC-SARS-003 PASS (3× MIRRORED parity OK + both-files-exist guard).
- [ ] AC-SARS-004 PASS (frontmatter-only diff).
- [ ] AC-SARS-005 PASS (NOTICE excluded + retained, both trees).
- [ ] AC-SARS-006 PASS (byte-sum strictly reduced per tree).
- [ ] AC-SARS-007 PASS (0 Class-B/C violations).
- [ ] AC-SARS-008 PASS (lifecycle-sync-gate live-scoped + template-absent).
- [ ] `make build` succeeds; `go test ./internal/template/...` green (embed integrity).
- [ ] spec-lint clean on the SPEC artifacts.

## D.4 Quality gate

- Frontmatter-only change → run-phase LSP gate is trivially satisfied (no Go code touched).
- Template neutrality (CLAUDE.local.md §15): `paths:` glob additions are language-neutral; no forbidden content class introduced.
