# Research — SPEC-SESSION-HANDOFF-ALIGN-001

> Evidence base for the SPEC scope. Two sections:
> - §A: 18-file workflow/ coverage audit table (REQ-SHA-008) — **iter-2 re-measured 2026-06-17, verbatim audit output pasted**
> - §B: Diet/V0 + `/cd` cache-preserving neutral-vs-dev content classification (input to design.md §B)
>
> **iter-2 correction note (2026-06-17)**: iter-1's §A.1 table was hand-transcribed and undercounted the LOCAL directory by 1 file (it claimed 17 LOCAL files; the actual LOCAL count is 18). The iter-1 table also carried incorrect line counts for `session-handoff.md` (claimed local=310/diff=112; actual local=314/diff=117) and `orchestration-mode-selection.md` (claimed 230; actual 234). Per the Verification-Claim Integrity doctrine, every attributed numerical claim must match an actually-observed baseline. iter-2 re-runs the audit command verbatim and pastes the real output below; all downstream findings and the FL-1 deferral are re-examined against the corrected baseline.

## §A. 18-file workflow/ coverage audit table

**Method**: for each `internal/template/templates/.claude/rules/moai/workflow/*.md`, diff against the local canonical at `.claude/rules/moai/workflow/<same-name>.md`. Classify drift severity. Additionally, scan the LOCAL directory for files that have no template mirror (the `template-missing` severity class).

**Command** (preserved for reproducibility — this is the exact command iter-2 executed):
```bash
echo "=== LOCAL workflow/ files ==="
ls .claude/rules/moai/workflow/*.md | wc -l
echo "=== TEMPLATE workflow/ files ==="
ls internal/template/templates/.claude/rules/moai/workflow/*.md | wc -l
echo "=== per-file diff (local vs template) ==="
for f in internal/template/templates/.claude/rules/moai/workflow/*.md; do
  b="${f##*/}"; tl=".claude/rules/moai/workflow/$b"
  if [ -f "$tl" ]; then
    d=$(diff "$f" "$tl" | grep -c '^[<>]')
    echo "$b | both | diff_lines=$d"
  else
    echo "$b | TEMPLATE-ONLY | (no local)"
  fi
done
echo "=== local-only (in local, not in template) ==="
for f in .claude/rules/moai/workflow/*.md; do
  b="${f##*/}"
  [ -f "internal/template/templates/.claude/rules/moai/workflow/$b" ] || echo "$b | LOCAL-ONLY | (no template mirror)"
done
echo "=== mirror test enrollment ==="
grep -n 'workflow/' internal/template/rule_template_mirror_test.go
```

> **Metric note**: the `diff_lines` column produced by this command counts `^[<>]` lines (the actual changed `<` and `>` lines, excluding the `---` separators). An alternative metric (`diff ... | wc -l`) includes the `---` separator lines and is therefore ~6 lines higher for the session-handoff.md case (111 vs 117). Both are reported below for completeness; the severity classification (`> 20 OR local-only sections` → major-drift) is unaffected by the choice of metric.

**Measurement date**: 2026-06-17 (iter-2 re-measurement). **Measurement baseline**: `main` HEAD pre-SPEC, `session-handoff.md` LOCAL=314 lines / TEMPLATE=209 lines (verified by `wc -l`).

### §A.0 Verbatim audit output (iter-2, observed 2026-06-17)

```
=== LOCAL workflow/ files ===
      18
=== TEMPLATE workflow/ files ===
      17
=== per-file diff (local vs template) ===
agent-teams-pattern.md | both | diff_lines=0
archived-agent-rejection.md | both | diff_lines=0
ci-autofix-protocol.md | both | diff_lines=0
ci-watch-protocol.md | both | diff_lines=0
context-window-management.md | both | diff_lines=0
dynamic-workflows.md | both | diff_lines=0
goal-directive.md | both | diff_lines=0
moai-memory.md | both | diff_lines=0
mx-tag-protocol.md | both | diff_lines=0
orchestration-mode-selection.md | both | diff_lines=0
session-handoff.md | both | diff_lines=111
spec-workflow.md | both | diff_lines=0
team-pattern-cookbook.md | both | diff_lines=0
team-protocol.md | both | diff_lines=0
verification-batch-pattern.md | both | diff_lines=0
worktree-integration.md | both | diff_lines=0
worktree-state-guard.md | both | diff_lines=0
=== local-only (in local, not in template) ===
lifecycle-sync-gate.md | LOCAL-ONLY | (no template mirror)
=== mirror test enrollment ===
46:	".claude/rules/moai/workflow/spec-workflow.md",
59:	//   - .claude/rules/moai/workflow/ci-watch-protocol.md (1 token)
61:	//   - .claude/rules/moai/workflow/agent-teams-pattern.md (1 token)
62:	//   - .claude/rules/moai/workflow/verification-batch-pattern.md (2 tokens)
```

### §A.1 Audit table (iter-2, reconciled with verbatim output)

The 17 TEMPLATE files (all mirrored in LOCAL) plus 1 LOCAL-ONLY file (`lifecycle-sync-gate.md`):

| # | File | local lines | template lines | diff lines (change-only, `^[<>]`) | diff lines (raw, `wc -l`) | severity | note |
|---|------|-------------|----------------|-----------------------------------|---------------------------|----------|------|
| 1 | `agent-teams-pattern.md` | — | — | 0 | 0 | in-sync | — |
| 2 | `archived-agent-rejection.md` | — | — | 0 | 0 | in-sync | — |
| 3 | `ci-autofix-protocol.md` | — | — | 0 | 0 | in-sync | — |
| 4 | `ci-watch-protocol.md` | — | — | 0 | 0 | in-sync | — |
| 5 | `context-window-management.md` | 81 | 81 | 0 | 0 | in-sync | — |
| 6 | `dynamic-workflows.md` | 115 | 115 | 0 | 0 | in-sync | — |
| 7 | `goal-directive.md` | 63 | 63 | 0 | 0 | in-sync | — |
| 8 | `moai-memory.md` | 158 | 158 | 0 | 0 | in-sync | — |
| 9 | `mx-tag-protocol.md` | 179 | 179 | 0 | 0 | in-sync | — |
| 10 | `orchestration-mode-selection.md` | 234 | 234 | 0 | 0 | in-sync | — |
| **11** | **`session-handoff.md`** | **314** | **209** | **111** | **117** | **MAJOR-DRIFT (local-only-sections)** | **THIS SPEC's target — 105-line net content delta, generic doctrine trapped local-only** |
| 12 | `spec-workflow.md` | 448 | 448 | 0 | 0 | in-sync | already enrolled in mirror test (L46) |
| 13 | `team-pattern-cookbook.md` | 312 | 312 | 0 | 0 | in-sync | — |
| 14 | `team-protocol.md` | 126 | 126 | 0 | 0 | in-sync | — |
| 15 | `verification-batch-pattern.md` | 66 | 66 | 0 | 0 | in-sync | — |
| 16 | `worktree-integration.md` | 425 | 425 | 0 | 0 | in-sync | — |
| 17 | `worktree-state-guard.md` | 140 | 140 | 0 | 0 | in-sync | — |
| **18** | **`lifecycle-sync-gate.md`** | **394** | **—** | **—** | **—** | **TEMPLATE-MISSING** | **LOCAL-ONLY (17564 bytes, 2026-06-16); carries internal SPEC-ID tokens — see §A.5** |

**Totals**: 18 LOCAL files, 17 TEMPLATE files, 16 in-sync (diff=0), 1 major-drift (`session-handoff.md`), 1 template-missing (`lifecycle-sync-gate.md`).

### §A.2 Severity classification rubric

- **in-sync**: `diff_lines == 0`. Local and template are byte-identical. No action.
- **minor-drift**: `diff_lines` in `[1, 20]` AND no local-only sections of substance. Typically whitespace, typo, or single-line phrasing. Triage optional.
- **major-drift**: `diff_lines > 20` OR local-only sections of substance (entire doctrine blocks absent from template). Requires a SPEC to close.
- **local-only-sections**: local has entire `##` sections absent from template (the session-handoff.md pattern — Diet + V0 + `/cd` cache-preserving).
- **template-missing**: file exists locally but has no template mirror at all (the `lifecycle-sync-gate.md` case — see §A.5).

### §A.3 Findings

- **F-A1**: Of 18 workflow/ files, **16 are in-sync** (diff=0), **1 is major-drift** (`session-handoff.md`, change-only diff=111 / raw diff=117), and **1 is template-missing** (`lifecycle-sync-gate.md`). This SPEC closes the major-drift outlier (session-handoff.md) and explicitly scopes the template-missing finding as a follow-up (§A.5).
- **F-A2**: Only **1 of 17** mirrored files (`spec-workflow.md`) is enrolled in the mirror test. The other 16 mirrored files are unguarded drift candidates — but today they happen to be in sync. The coverage gap is systemic: the mirror set was built incrementally per-SPEC and never audited for coverage.
- **F-A3**: **Bulk enrollment decision (FL-1, deferred — still holds post-iter-2)**: enrolling all 16 in-sync mirrored files in the mirror test would provide defense-in-depth against future drift recurrence. Cost: every future single-tree edit on these 16 files requires lockstep template edits. Benefit: CI catches drift on PR rather than at the next ad-hoc audit. **iter-2 re-check**: the corrected audit (16 in-sync mirrored files, not 15) strengthens the case for a deliberate, evidence-based deferral rather than an ad-hoc one — the count is now accurate and the FL-1 carve-out (EXCL-005) correctly names "16 in-sync siblings" (updated from the iter-1 "15"). This decision is OUT OF SCOPE for this SPEC — the audit table informs it but does not authorize it. A follow-up SPEC may propose bulk enrollment if the maintenance cost is judged acceptable.
- **F-A4**: The 6-file §25 carve-out in `rule_template_mirror_test.go` L50-63 (manager-develop-prompt-template, ci-watch-protocol, agent-common-protocol, agent-teams-pattern, verification-batch-pattern, plan-auditor) is a DIFFERENT mechanism — those files retain intentional internal-content tokens and are covered by the leak test instead of byte-parity. session-handoff.md does NOT join this carve-out; it achieves true byte-parity via neutralization.

### §A.4 Recommendation

Close the session-handoff.md outlier via this SPEC (M1-M6). Defer the bulk-enrollment decision to a follow-up audit SPEC once the maintainer has experience with the lockstep-edit cost on session-handoff.md + spec-workflow.md (the 2 enrolled files).

### §A.5 lifecycle-sync-gate.md template-missing finding (iter-2 NEW, from D1 remediation)

**Observation**: `lifecycle-sync-gate.md` exists LOCAL-ONLY (`.claude/rules/moai/workflow/lifecycle-sync-gate.md`, 394 lines, 17564 bytes, 2026-06-16). There is no template mirror at `internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md`. This is a `template-missing` case per the §A.2 rubric.

**Content profile (generic doctrine vs internal-only)**: the file's body is **predominantly generic V3R6 era doctrine** — the era classification heuristic table (H-1..H-6), the grandfather clause policy, the `era:` frontmatter field semantics, and the Status Transition Ownership Matrix cross-reference. These are all concepts that apply to ANY user project that adopts the MoAI SPEC lifecycle.

**However, the file carries internal-content tokens** that would fail the `template-neutrality-check.yaml` / `internal_content_leak_test.go` CI guard if ported naively:
- L235: `SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001` (named as the close-subject-mandate origin)
- L235: `SPEC-CCSYNC-CLAUDEMD-001` (used as a close-commit subject example)
- L235: `SPEC-CCSYNC` (used as a prohibited combined-scope example)
- L394: `SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M5` (origin footer)
- Canonical Go implementation paths: `internal/spec/era.go`, `internal/spec/audit.go`, `internal/spec/drift.go`, `internal/spec/CLAUDE.md`, `internal/spec/audit_test.go` (these reference the moai-adk Go source tree, which IS shipped via the `moai init` template, so they are arguably neutral — but the SPEC-ID citations are not).

**Scope decision for THIS SPEC**: **defer porting to a follow-up SPEC**. Rationale: (a) porting lifecycle-sync-gate.md is a distinct doctrinal-port task with its own neutralization work (the SPEC-ID citations in L235 + the L394 origin footer must be stripped or generalized, mirroring the Diet/V0 neutralization this SPEC performs on session-handoff.md); (b) this SPEC's scope is the session-handoff.md major-drift outlier — widening it to include lifecycle-sync-gate.md would violate the surgical-scope principle (CLAUDE.md §16 anti-overengineering); (c) the file's LOCAL-ONLY status is NOT silently shipped — this §A.5 documents it explicitly, and EXCL-006 (added in spec.md §B.3) carves it out of this SPEC's scope so a follow-up SPEC can pick it up.

**Irony note**: this SPEC's own spec.md §D constraint #4 and §H cross-reference L145 depend on lifecycle-sync-gate.md for the V3R6 era doctrine (the commit_sha conventions and the §E marker names). The file is load-bearing for the SPEC's own lifecycle discipline, yet its template-missing status was invisible to iter-1's audit. iter-2 surfaces this dependency explicitly so the follow-up SPEC that ports it is aware that lifecycle-sync-gate.md is referenced by active V3R6 SPECs.

---

## §B. Diet / V0 / `/cd` cache-preserving neutral-vs-dev content classification

This classification is the operational input to design.md §B (token-stripping enumeration). Every paragraph of the three local-only content blocks is classified as **neutral-generic** (ships to template) or **dev-incident-provenance** (stays local-only OR moves to lesson memory).

> **iter-2 addition (from D7 remediation)**: a THIRD local-only content block was identified during the iter-1 audit's second pass — the `### /cd cache-preserving alternative (CC 2.1.169+)` section at LOCAL L249. iter-1's §B covered only Diet and V0; iter-2 adds §B.3 for this block and folds it into the REQ-SHA-006 mirror-back scope. The LOCAL↔TEMPLATE delta therefore consists of THREE local-only content blocks (Diet + V0 + `/cd`) plus the 3 stale SPEC-ID lines (REQ-SHA-005) — not two blocks as iter-1 implied.

### §B.1 Diet Constraints (L127-183 local)

| Content block | Local lines | Classification | Disposition |
|---------------|-------------|----------------|-------------|
| Opening paragraph: "paste-ready resume message는 next session minimum executable context이다 — audit trail, history record, ceremonial commitment record가 아니다" | L129 | **MIXED** — the philosophy is generic; the trailing parenthetical "(LIFECYCLE-SYNC-GATE-001 line C 1~14차 + HARNESS-NAMESPACE Phase 1B line B 1~5차에서 동일 비대화 패턴 관측)" is dev-incident | **Strip the parenthetical**; reword as "차수 누적 retry 진행 시 본문에 history/lesson/directive escalation prose를 append-only로 누적하는 것은 empirical 입증된 anti-pattern이다." (empirical without the SPEC-ID evidence chain) |
| `### Block 2 applied lessons 제약` (≤4 refs, 1-line identifier) | L131-135 | **neutral-generic** | Ship verbatim |
| `### Block 4 precondition 제약` (≤200 chars, no history prose) | L137-142 | **neutral-generic** | Ship verbatim |
| `### Block 5 실행 제약` (single primary action, no ceremonial reminder) | L144-148 | **neutral-generic** | Ship verbatim |
| `### Block 6 후속 제약` (≤2 lines) | L150-153 | **neutral-generic** | Ship verbatim |
| `### Doctrine reference 패턴` (N차 sustained history → lesson memory) | L155-158 | **neutral-generic** | Ship verbatim |
| `### Anti-pattern catalogue` (AP-D-001..005) | L160-166 | **neutral-generic** (the bodies describe generic shapes; only AP-D-005's example "B8/B15 discipline 엄수" / "plan.md §F.3 line 130-143" is illustrative-of-the-shape, not internal-token-bearing in a way that leaks a SPEC) | Ship verbatim (optionally generalize AP-D-005's example placeholders, per Finding #19 optional note) |
| `### Pre-emit self-check (8 items)` | L168-177 | **neutral-generic** | Ship verbatim |
| `### 적용 범위` bullet 1 ("모든 신규 paste-ready resume message") | L181 | **neutral-generic** | Ship verbatim |
| `### 적용 범위` bullet 2 ("차수 누적 retry paste-ready") | L182 | **neutral-generic** | Ship verbatim |
| `### 적용 범위` bullet 3 ("Cross-line 일관 적용 (LIFECYCLE-SYNC-GATE / HARNESS-NAMESPACE / SESSION-AUTO-RESUME 등 모든 SPEC line)") | L183 | **MIXED** — the scope statement is generic; the parenthetical names 3 internal SPEC lines | **Strip the parenthetical**; reword as "Cross-line 일관 적용 (모든 SPEC line)" |

**Net**: the entire Diet section ships to template, with 2 parenthetical strips (L129, L183). All budgets, AP-D catalogue, and Pre-emit self-check are verbatim.

### §B.2 V0 Abort Gate Doctrine (L185-226 local)

| Content block | Local lines | Classification | Disposition |
|---------------|-------------|----------------|-------------|
| Opening: "paste-ready Block 4 V0 precondition은 lsof + cwd 교차 검증을 사용한다. `ps aux` raw count는 environmental baseline noise..." | L187 | **MIXED** — the doctrine is generic; the trailing "(cross-line empirical 입증)" is borderline (the phrase "empirical" without the SPEC-ID chain is acceptable) | **Strip "cross-line empirical 입증" parenthetical**; keep "empirical 입증된 anti-pattern이다" generic form |
| `### V0 검증 명령 (canonical)` — V0-a `ps aux` informational baseline | L192-193 | **neutral-generic** | Ship verbatim |
| `### V0 검증 명령 (canonical)` — V0-b `lsof -a -c claude +D "$PWD"` STRICT 0 | L195-199 | **MIXED** — the command + STRICT criterion is generic; the inline comment "Hugo docs 서버 PID 1개 → 8 entry 오탐, cross-line 입증" references an internal incident | **Strip the "cross-line 입증" and Hugo docs server reference** from the inline comment; keep the generic "파일명에 'claude' 포함된 콘텐츠까지 매칭하는 false-positive 결함이 있다" + the COMMAND-column-filter prescription |
| `### V0 검증 명령 (canonical)` — V0-c `lsof -a -c claude -d cwd` STRICT ≤2 | L201-202 | **neutral-generic** | Ship verbatim |
| `### Abort 의무` (V0-b ≥1 OR V0-c ≥3 → abort, no spawn, no override) | L205-211 | **neutral-generic** | Ship verbatim |
| `### Cross-pollination 이력` (Line C 9차/10차, Line A 13차, Line B 14차) | L213-219 | **100% dev-incident log** — pure internal iteration history with SPEC-IDs | **DROP entirely from template**; collapse in local to a 1-line lesson reference (REQ-SHA-014). The iteration history is retained in lesson memory, not in the rule body. |
| `### Anti-pattern` — AP-V-001 (`ps aux` raw count as sole V0) | L223 | **neutral-generic** | Ship verbatim |
| `### Anti-pattern` — AP-V-002 (N회 미이행 추적 guilt-trip) | L224 | **neutral-generic** | Ship verbatim |
| `### Anti-pattern` — AP-V-003 (AskUserQuestion override option) | L225 | **neutral-generic** | Ship verbatim |
| `### Anti-pattern` — AP-V-004 (`lsof +D | grep -iE 'claude'` filename false-positive) | L226 | **MIXED** — the generic lesson (use COMMAND-column filter, not filename-grep) is generic; the trailing "(claude-md-guide.md·claude-design-handoff.md 등)... LIFECYCLE-SYNC-GATE-001 M4 1·2차에서 동일 false abort 유발" is internal-file provenance | **Strip the internal-file list and SPEC-ID provenance**; keep the generic lesson: "AP-V-004: V0-b 측정에 `lsof +D \"$PWD\" | grep -iE 'claude'` 사용 → 파일명에 'claude' 포함된 콘텐츠까지 매칭하는 false-positive. COMMAND 컬럼 프로세스 필터 `lsof -a -c claude +D \"$PWD\"` 필수." |

**Net**: the V0 section ships to template with 3 strips (L187 parenthetical, L195-199 Hugo/cross-line inline comment, L226 internal-file provenance) and 1 full drop (L213-219 Cross-pollination 이력 block). The 3 canonical commands, abort obligation, and AP-V-001..004 generic bodies are verbatim.

### §B.3 `/cd` cache-preserving alternative (LOCAL L249) — iter-2 NEW (from D7 remediation)

The LOCAL file has a `### /cd cache-preserving alternative (CC 2.1.169+)` subsection at L249, sitting inside the `## Worktree-Anchored Resume Pattern` section (L228) and before `## Cross-references` (L302). The TEMPLATE file does NOT have this subsection — the TEMPLATE's Worktree-Anchored section jumps from the Block 0 format directly to the Cross-references section.

| Content block | Local lines | Classification | Disposition |
|---------------|-------------|----------------|-------------|
| `### /cd cache-preserving alternative (CC 2.1.169+)` — documents the `claude --resume` / `/cd` cache-preserving entry pattern introduced in CC 2.1.169 | L249 (block) | **neutral-generic** — the content references only Claude Code platform behavior (CC version `2.1.169+`) and generic worktree-resume mechanics; no internal SPEC-IDs, no internal filenames, no dev-incident provenance | **Ship to template verbatim** (port under REQ-SHA-006 mirror-back scope, alongside Diet + V0). The block is ~14 lines of platform-feature documentation that applies to any user project using Claude Code worktree-resume. |

**Net**: the `/cd` block ships to template verbatim — no token stripping required. iter-1 missed this block entirely (§B covered only Diet + V0); iter-2 folds it into REQ-SHA-006's mirror-back scope so the LOCAL↔TEMPLATE delta closes completely, not partially.

**Consolidated LOCAL↔TEMPLATE delta (iter-2 reconciled)**: the session-handoff.md major-drift consists of (a) 3 stale SPEC-ID lines at L68/L69/L122 (REQ-SHA-005, local→template realignment), (b) Diet Constraints section trapped local-only (REQ-SHA-001/002), (c) V0 Abort Gate section trapped local-only (REQ-SHA-003/004), (d) `/cd` cache-preserving subsection trapped local-only (REQ-SHA-006 mirror-back, iter-2 NEW), plus (e) the M4 structural relocation of Diet + V0 from mid-file to after-Worktree-Anchored (REQ-SHA-016). The net content delta (105 lines = raw 117 − separator/heading overhead) collapses to zero on canonical + neutralized-appendix content post-M4.

### §B.4 Aggregate token-strip enumeration (input to design.md §B)

The exact tokens to strip during porting (Diet + V0 only — the `/cd` block needs no stripping):

1. `LIFECYCLE-SYNC-GATE-001` (Diet L129 parenthetical, V0 L213-219 Cross-pollination, V0 L226 AP-V-004 trailing)
2. `HARNESS-NAMESPACE` (Diet L129 parenthetical, V0 L213-219 Cross-pollination)
3. `SESSION-AUTO-RESUME` / `SESSION-AUTO-RESUME-001` (Diet L183 parenthetical, V0 L213-219 Cross-pollination)
4. `line C 1~14차`, `line B 1~5차`, `Line C 9차`, `Line C 10차`, `Line A 13차`, `Line B 14차` (Diet L129 parenthetical, V0 L213-219 entire block)
5. `Hugo docs 서버 PID` / `Hugo docs 서버` (V0 L196-198 inline comment, V0 L226 AP-V-004 trailing)
6. `claude-md-guide.md`, `claude-design-handoff.md` (V0 L226 AP-V-004 trailing)
7. `M4 1·2차` / `LIFECYCLE-SYNC-GATE-001 M4` (V0 L226 AP-V-004 trailing)
8. `cross-line empirical 입증` / `cross-line 입증` (Diet L129, V0 L187, V0 L196-198 — reword to "empirical 입증" without the cross-line chain)
9. `SPEC-V3R6-MULTI-SESSION-COORD-001`, `REQ-COORD-009` (the 3 stale local lines L68, L69, L122 — REQ-SHA-005)

**Post-port verification grep** (AC-SHA-002 + AC-SHA-004):
```bash
grep -nE 'SPEC-V3R[0-9]-[A-Z]|LIFECYCLE-SYNC-GATE|HARNESS-NAMESPACE|SESSION-AUTO-RESUME|Hugo docs|claude-md-guide|claude-design-handoff|cross-line (empirical )?입증|line [ABC] [0-9]+차|M4 1·2차' \
  internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
# Expected: zero matches on the ported Diet+V0+`/cd` content (historic references in HISTORY/changelog sections outside the ported blocks are acceptable if any exist; none expected here)
```

### §B.5 What stays local-only (after M2/M3 mirror-back)

After REQ-SHA-006 mirrors the neutralized versions back to local, the local file has NO local-only content left — both trees are byte-identical. The dev-incident provenance (the Cross-pollination 이력 iteration history) is retained in **lesson memory** (referenced by the 1-line lesson pointer REQ-SHA-014 produces), NOT in either rule file. This is the correct layer for iteration history per the Diet doctrine itself (AP-D-002: history → lesson memory, not rule body).
