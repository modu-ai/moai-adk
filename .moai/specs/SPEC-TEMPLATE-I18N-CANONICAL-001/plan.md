# SPEC-TEMPLATE-I18N-CANONICAL-001 — Implementation Plan

Tier L. Markdown-only for M1–M3 (always-loaded rule files); M4 adds one test class + optional doctrine amendment. Route A (Hybrid Trunk main-direct: commit + push to `main`, no PR unless `--pr` per CLAUDE.local.md §23 Tier-based routing — Tier L default is PR, so expect PR creation at sync-phase).

## §A Context

- **Repo**: `/Users/goos/MoAI/moai-adk-go`, branch `main`, plan baseline SHA `39c74d77787621b6645aebe81e470277ba3c97cb`.
- **SPEC artifacts**: `.moai/specs/SPEC-TEMPLATE-I18N-CANONICAL-001/{spec.md, research.md, plan.md, acceptance.md, progress.md}` (Tier L 5-artifact set).
- **Targets (Template-First order, M1–M3)**:
  1. `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` (EDIT — translate § Recommendation Placement Principles + 3 worked examples to English-canonical)
  2. `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` (EDIT — English-first skeleton, drop "(canonical)", resolve tiering contradiction, add memory-heading loc-table row, fix goal-first example)
  3. `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` (EDIT — English-canonical resume format + cross-ref session-handoff.md)
  4. `internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md` (EDIT — add ja/zh rows for the new memory-heading element per REQ-I18N-006)
  5. `internal/template/template_neutrality_audit_test.go` (EDIT M4 — new `C9-natural-language-canonical-form` class)
  6. `.moai/docs/template-internal-isolation-doctrine.md` and/or `CLAUDE.local.md` (EDIT M4 — §25 and/or §15 line 611 natural-language neutrality sub-section)
  7. Local mirrors `.claude/rules/moai/...` for items 1–4 (SYNC — byte-identical)
- **PRESERVE (never touch)**: the 2 clean baseline files (`core/agent-common-protocol.md`, `workflow/goal-directive.md` × 2 trees), all other rule files, all Go production source (except the M4 test file), all skills, all other SPEC dirs, `.moai/state/*`.
- **Plan-time facts** (baseline-attributed, measured at `39c74d777` with ripgrep): askuser-protocol.md 1,176 Hangul syllables; session-handoff.md 186; context-window-management.md 18; agent-common-protocol.md 0; goal-directive.md 0. Localization Table row count in session-handoff.md = 9 (the floor for REQ-I18N-001). **Template ↔ local askuser-protocol.md parity**: at the `39c74d777` baseline the two trees were byte-identical (477 = 477 lines). At current HEAD they are INTENTIONALLY DIVERGED (24 diff lines) — the template tree is §25-neutral (0 `SPEC-V3R6-ASKUSER-DECISION-MEMORY-001` provenance refs) while the local tree retains 6 dev-trace provenance pointers (`AC-ADM-005..017`, Epic 7 TMC-001, §24 namespace). This is §25-intentional drift, not a sync defect. The sweep's parity target (AC-I18N-019) is narrowed accordingly: sweep-induced changes sync; pre-existing §25-intentional dev-trace drift is preserved.

## §B Known Issues (filtered to relevant categories)

- **B4 Frontmatter schema**: SPEC artifacts use `created:`/`updated:`/`tags:` canonical names (done at plan). Rule files have no frontmatter schema constraint — they are markdown bodies.
- **B6 spec-lint heading**: spec.md carries `### Out of Scope — <topic>` H3 subsections (done); do not restructure during run.
- **B8/B10 Working-tree hygiene & scope**: commit with specific paths only (`git add <path>`); parallel sessions are historically active on this checkout — run the pre-spawn sync check (`git fetch` + `rev-list --left-right`) before commits; never `git add -A`.
- **B9 Commit/push**: manager-develop commits + pushes per milestone. Conventional Commits: `feat(SPEC-TEMPLATE-I18N-CANONICAL-001): M<N> <subject>` + `🗿 MoAI` trailer. First run-phase commit flips frontmatter `draft → in-progress`.
- **Domain-specific pitfalls**:
  - *Hangul-count tooling*: use `rg -oN '[가-힣]'` (ripgrep) for all Hangul-range verification — for portability across locale-handling differences and to avoid the windowed-grep undercount risk (`sed -n <window> | grep` pipelines; `feedback_windowed_grep_undercount_authoring`). NOTE: an earlier plan-draft claimed "BSD grep returns 0 on macOS" — retracted at v0.2.0 after verification showed `grep -oE` (ugrep 7.5.0 on this Darwin 25.5.0 env) and `rg` both return the correct 1176 count; the 0-count was a broken `ggrep|wc -l` fallback masking a missing binary, not a grep defect. The AC commands (`rg`) are unchanged.
  - *Localization mechanism is sacrosanct*: the 4-locale table, the render-time substitution, and the English-fallback rule MUST survive the sweep. M2 edits to session-handoff.md touch the canonical-skeleton block + loc-table header annotation + the 2 leak sites — NEVER the loc-table's Korean column data, NEVER the render-time substitution instruction's semantics, NEVER the English-fallback rule.
  - *`session-handoff-examples.md` parity*: the ja/zh extension columns live there. When REQ-I18N-006 adds a memory-heading row, the row MUST land in BOTH the inline table (en/ko) AND the examples file (ja/zh) — split-site editing.
  - *`make build`*: M1–M3 are markdown-only (no embed regeneration needed); M4 touches a Go test file so `go test ./internal/template/` is the gate (not `make build`).

## §C Pre-flight (before any edit)

```bash
cd /Users/goos/MoAI/moai-adk-go
git branch --show-current && git rev-parse HEAD
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# Ripgrep-verified Hangul baselines (portability across locale handling; avoids windowed-grep undercount)
for f in internal/template/templates/.claude/rules/moai/core/askuser-protocol.md \
         internal/template/templates/.claude/rules/moai/workflow/session-handoff.md \
         internal/template/templates/.claude/rules/moai/workflow/context-window-management.md \
         internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
         internal/template/templates/.claude/rules/moai/workflow/goal-directive.md; do
  printf "%s\t%s\n" "$(rg -oN '[가-힣]' "$f" | wc -l | tr -d ' ')" "$f"
done
# Localization Table row baseline (floor for REQ-I18N-001)
grep -cE '^\|.*(English|Korean|Japanese|Chinese|Copy from here|여기부터|Block|Cut-line)' \
  internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
# askuser template↔local parity (D2: intentionally diverged at current HEAD — §25-neutral template vs local dev-trace)
diff -rq internal/template/templates/.claude/rules/moai/core/askuser-protocol.md .claude/rules/moai/core/askuser-protocol.md \
  && echo "askuser: identical" \
  || echo "askuser: diverged — confirm §25-intentional (template neutral, local retains SPEC-V3R6-ASKUSER provenance). If NOT §25-intentional, emit blocker report to choose dev-trace-preserve vs neutral-sync before M1."
diff -rq internal/template/templates/.claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/session-handoff.md
```

## §D Constraints (DO NOT VIOLATE)

- **Localization mechanism sacrosanct** (REQ-I18N-001/013): the 4-locale table data, the render-time substitution instruction, and the English-fallback rule survive the sweep unchanged in semantics. The sweep changes canonical *framing* (default skeleton + "(canonical)" label + tiering prose), never locale *coverage*.
- **Write whitelist**: items 1–7 in §A Targets + `.moai/specs/SPEC-TEMPLATE-I18N-CANONICAL-001/*`. Nothing else.
- **The 2 clean baseline files are byte-frozen** (REQ-I18N-011) — verified by `git diff 39c74d777..HEAD -- <8 paths>` (4 files × 2 trees) empty.
- **Hangul verification uses ripgrep** (`rg -oN '[가-힣]'`) for portability across locale-handling differences; windowed-grep pipelines (`sed -n <window> | grep`) are prohibited (plan-phase known-issue §B).
- No `--no-verify`, no `--amend` on pushed commits, no force-push.
- **Blockers** (e.g., doctrine-amendment contention in M4, loc-table edit ambiguity) → structured blocker report to orchestrator; never AskUserQuestion, never silent scope change.

## §E Self-Verification Deliverables (run-phase completion report)

Reported per `verification-claim-integrity.md` 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk):

- **E1** — AC binary PASS/FAIL matrix for all 23 checks (acceptance.md §D) with executed command (ripgrep-based for Hangul counts) + verbatim output per mechanical/hybrid AC.
- **E2** — `go test ./internal/template/ -run TestTemplateNeutrality` exit 0 (M4: the new C9 class passes; M1–M3: no regression to C1–C8); `go build ./...` exit 0 (sanity — no production Go touched).
- **E3** — Coverage: n/a (markdown-only for M1–M3); M4's new test class is its own coverage.
- **E4** — Byte-parity evidence: `diff -rq` template vs local for the 3–4 modified rule files → identical.
- **E5** — Byte-freeze evidence: `git diff 39c74d777..HEAD -- <8 baseline-file paths>` → empty (the 2 clean files × 2 trees untouched).
- **E6** — Localization mechanism preservation evidence: Localization Table row count post-sweep ≥ 9 (pre-sweep floor) + the 4 locale columns present + the render-time substitution instruction intact (grep for the substitution clause).
- **E7** — Commit SHAs per milestone + `git push origin main` result.

## §F Milestones (priority-ordered, no time estimates)

### M1 — askuser-protocol.md English-canonical (largest surface)

- Translate § "Recommendation Placement Principles" (~lines 127-175, ~50 lines, 5 HARD principles) from Korean prose to English-canonical prose. Preserve verbatim in English: `[ZONE:Evolvable]` tags, Fisher-information formula `I=p(1−p)` and `p ≈ 0.5` cold-start heuristic, 5-principle numbering (1. emission timing / 2. question ordering / 3. statistical-majority default / 4. precondition statement / 5. adaptive strength), and cross-references (`verification-claim-integrity.md §1.1 surface 3`, `design.md §A.4` / `§B.2`).
- Translate the 3 Korean worked examples to English-canonical payloads: anti-pattern block (~line 32), Epic-8 AskUserQuestion worked example (~line 246), correct-pattern example (~line 436). Preserve the `(권장)` / `(Recommended)` token discussion (locale-aware by design).
- Exit: `rg -oN '[가-힣]' askuser-protocol.md` count drops from 1,176 toward ~0 (residual only in the `(권장)` token discussion, which is locale-aware); AC-I18N-002/003/004/005 green.
- Commit: `feat(SPEC-TEMPLATE-I18N-CANONICAL-001): M1 askuser-protocol.md English-canonical policy section + examples` (flips `draft → in-progress`).

### M2 — session-handoff.md skeleton + loc-table + 2 leaks

- Flip the canonical 6-block skeleton (§ Canonical Format) to English-first headers (`Preconditions:` / `Run:` / `After merge:` / `Follow-up:`) OR locale-neutral placeholders. Korean renderings move to the Localization Table Korean column only.
- Localization Table: drop `"(canonical)"` from the Korean column header (→ `Korean`); the column labels become `Element | English | Korean` (equal-tier).
- Tiering prose: change `"primary locales"` (line ~67) to `"inline locales"`; resolve the line ~67/82 contradiction (the declared structural default must match the English-fallback behavior).
- REQ-I18N-006 (memory heading): add a Localization Table row — en `## Next Session Entry Point` / ko `## 다음 세션 시작점` inline; ja `## 次セッション開始点` / zh `## 下一会话起点` in `session-handoff-examples.md`. Extend the render-time substitution instruction to name the memory heading as substitutable. Update the line ~229 cross-reference to match-by-content.
- REQ-I18N-007 (goal-first example): rewrite line ~163 English-canonical with a locale note.
- Exit: AC-I18N-006/007/008/009/010/011/012 green; Localization Table row count ≥ 10 (was 9, +1 memory heading); `Korean (canonical)` substring returns 0 hits.
- Commit: `feat(SPEC-TEMPLATE-I18N-CANONICAL-001): M2 session-handoff.md English-first skeleton + loc-table fixes`.

### M3 — context-window-management.md resume format

- Rewrite the canonical resume-message format example (§ Orchestrator Responsibilities, lines ~66-71) English-canonical.
- Add a cross-reference to `session-handoff.md § Localization Table` for locale renderings; do NOT redefine a parallel format.
- Exit: `rg -oN '[가-힣]' context-window-management.md` count drops from 18 toward 0; AC-I18N-013/014 green.
- Commit: `feat(SPEC-TEMPLATE-I18N-CANONICAL-001): M3 context-window-management.md English-canonical resume format`.

### M4 — CI lint class + doctrine amendment (regression prevention)

- Add a new neutrality class `C9-natural-language-canonical-form` to `internal/template/template_neutrality_audit_test.go`'s `neutralityClasses` slice. The detector flags natural-language canonical-form prose bias OUTSIDE explicit localization tables (e.g., a `"(canonical)"` column-label detector; a Hangul-concentration heuristic for canonical-skeleton blocks; a symmetric-locale-column assertion). Mirror the existing class-detector structure (regexp + allow-list pattern).
- Propose the doctrine amendment: a new sub-section under `.moai/docs/template-internal-isolation-doctrine.md` §25 and/or `CLAUDE.local.md` §15 (line 611, "템플릿 언어 중립성") documenting natural-language canonical-form neutrality as a governed dimension. The amendment scope is limited to a new sub-section — it MUST NOT alter existing C1–C8 class definitions.
- **Conditional halt**: if the doctrine amendment is contentious (e.g., reviewer argues natural-language neutrality is out-of-scope for §25's internal-trace focus), emit a blocker report and split M4 into a follow-up SPEC. Do not force the amendment.
- Exit: `go test ./internal/template/ -run TestTemplateNeutrality` exit 0; AC-I18N-015/016 green.
- Commit: `feat(SPEC-TEMPLATE-I18N-CANONICAL-001): M4 CI lint C9 + doctrine §25/§15 natural-language neutrality amendment`.

### M5 — Parity, full AC matrix, close-out

- Sync local tree (items 1–4 + session-handoff-examples.md) → `diff -rq` byte-identical for each.
- Run the full AC matrix (23 checks); populate `progress.md` §E.2/§E.3.
- Localization mechanism preservation sweep: confirm ko-locale and en-locale render paths still produce correct localized output (AC-I18N-001c/d); confirm the 4 locale columns survive; confirm row count ≥ 10.
- Commit: `feat(SPEC-TEMPLATE-I18N-CANONICAL-001): M5 template↔local parity + AC matrix close-out`; push.

## §G Risks & mitigations

| # | Risk | Mitigation |
|---|------|------------|
| R1 | M2 loc-table edit accidentally drops a Korean cell or a ja/zh column | REQ-I18N-013 invariant + AC-I18N-001a/b mechanical row-count + 4-column check; split-site editing (inline + examples file) verified separately |
| R2 | M1 translation of the policy section loses the Fisher-information math semantics | REQ-I18N-002 preserves the math verbatim in English; AC-I18N-004 greps for `I=p(1−p)` and the 5-principle numbering |
| R3 | Windowed-grep pipeline (`sed -n <window> \| grep`) used for Hangul counts → undercount → false PASS | plan.md §B known-issue; all AC commands specify `rg -oN '[가-힣]'`; pre-flight uses ripgrep (the v0.2.0 BSD-grep retraction does NOT change this — windowed-grep undercount is the real, preserved hazard) |
| R4 | M4 doctrine amendment is contentious → scope creep | M4 conditional-halt clause: blocker-report + split into follow-up SPEC rather than force |
| R5 | M4 CI lint C9 false-positives on legitimate Korean content (loc-table cells, examples meant to demonstrate Korean) | the detector scopes to OUTSIDE explicit localization tables; allow-list for loc-table regions; tune threshold at run-phase |
| R6 | Parallel-session/worktree race on shared checkout — concrete active sources at plan baseline: `SPEC-AGENT-ARCH-V2-001` (worktree branch `97c0f8aaf` on `worktree-agent-ada0603163d6379ce`, pushed to origin) + multiple live `worktree-agent-*` checkouts on this repo | pre-spawn sync check (`git fetch` + `rev-list --left-right`); specific-path commits; re-fetch before each push |
| R7 | Translating askuser worked examples breaks the `(권장)` token discussion (which is locale-aware) | REQ-I18N-003 explicitly preserves the `(권장)`/`(Recommended)` token; AC verifies the token discussion survives |
| R8 | Pre-existing askuser template↔local §25-intentional drift (D2) — the sweep could accidentally "sync" away the local dev-trace provenance, or push dev-traces into the template (§25 violation) | AC-I18N-019 narrowed to sweep-induced-only; pre-flight askuser-parity check emits a blocker report if drift is non-§25-intentional; M1 commits use specific paths |

## §H Cross-references

- `spec.md` / `acceptance.md` / `research.md` / `progress.md` — this SPEC's sibling artifacts.
- `research.md` §C — doctrinal blind-spot analysis grounding M4.
- `research.md` §E — methodological hazards (Hangul-count tooling portability + v0.2.0 BSD-grep retraction, severity reconciliation).
- `SPEC-TEMPLATE-RULES-CLEANUP-001` / `SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001` — Tier L template-rules / neutrality-audit precedents.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter + status transition ownership.
- `CLAUDE.local.md` §2 (Template-First) / §15 (16-language neutrality) / §23 (Tier-based PR routing) / §25 (Template Internal-Content Isolation).
- `.moai/docs/template-internal-isolation-doctrine.md` §25 + `CLAUDE.local.md` §15 (line 611) — doctrinal SSOTs (M4 amendment targets).
- `.claude/rules/moai/core/verification-claim-integrity.md` — evidence format for E1–E7.
