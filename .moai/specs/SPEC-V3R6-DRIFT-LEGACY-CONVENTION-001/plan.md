# Implementation Plan — SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001

> Tier M plan. Section A-E delegation template (per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — REQUIRED for Tier M). Milestone-split run-phase. No time estimates — priority labels + phase ordering only.

---

## A. Context

### A.1 Location + branch

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Module: `internal/spec` (Go code — NOT template content; §25 template-neutrality does NOT gate the Go edits)
- Branch: per Hybrid Trunk 1-person OSS policy (`.moai/docs/git-local-workflow-doctrine.md`), Tier M may push directly to `main` with CI 4-check + pre-push guard. Run-phase manager-develop creates `feat/SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001` commits OR main-direct per `git-workflow-doctrine.md` Tier M.

### A.2 SPEC artifact paths

- `.moai/specs/SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001/spec.md` (GEARS REQ ①②③④ + ⑤doctrine, §A 5-mechanism table, §B exclusions, §3 AC matrix)
- `.moai/specs/SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001/plan.md` (this file)
- `.moai/specs/SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001/acceptance.md` (Given-When-Then per AC)

### A.3 Predecessor + relationship

- **Predecessor**: SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001 (status: completed, v0.2.0). Added `closeInfixMatch` (`4-phase close` / `Mx-phase audit-ready` → completed), `isSyncPhaseDocs` (`docs(SPEC-XXX): … sync-phase` → implemented), narrow backfill-skip in `shouldSkipCommitTitle`. Reduced drift 67→50. Named this SPEC explicitly in its §B Out of Scope as the legacy-convention follow-up.
- **This SPEC**: COMPREHENSIVE detector alignment — fixes the 4 residual false-positive mechanisms (①②③④) + aligns drift.go with audit.go's era/grandfather/terminal policy. Genuine-incomplete backfill (⑤) is OUT OF SCOPE (separate operational follow-up).

### A.4 Existing infrastructure (PRESERVE + EXTEND)

| Component | File | Disposition |
|-----------|------|-------------|
| `DetectDrift` / `getGitImpliedStatus` walker | `drift.go` | EXTEND (add era exemption, terminal authority, combined-scope mapping) |
| `commitMatchesSPECID` / `ExtractSPECIDs` | `drift.go` / `transitions.go` | EXTEND for ① (combined-scope), PRESERVE word-boundary (LSGF-001) |
| `shouldSkipCommitTitle` (chore + backfill skip) | `drift.go` | PRESERVE unchanged (AC-LSCSK-003, AC-DCA-003/008) |
| `closeInfixMatch` / `isSyncPhaseDocs` | `transitions.go` | PRESERVE (predecessor's positive close/sync signals) |
| `transitionRules` `{sync→completed}` / `{docs(sync)→completed}` | `transitions.go` | FIX (② — correct to implemented per 4-phase model) |
| `ClassifyEra` / `LoadEraSignalsFromDir` / `Era.EraFinal` | `era.go` | CONSUME READ-ONLY (④ — drift.go reuses, era.go NOT modified) |
| `Audit` / `checkV3R6Drift` terminal early-return | `audit.go` | REFERENCE pattern (③/④ — audit.go NOT modified) |
| Status Transition Ownership Matrix | `lifecycle-sync-gate.md` + `spec-frontmatter-schema.md` | AMEND (⑤doctrine — full-ID mandate) |

### A.5 PRESERVE list (DO NOT MODIFY)

- `internal/spec/era.go` — consumed READ-ONLY (no behavior change)
- `internal/spec/audit.go` — REFERENCE only (no behavior change; AC-DLC-007 asserts byte-identical audit output)
- `internal/spec/drift_chore_skip_test.go` — existing regression tests MUST stay green
- `internal/spec/drift_specid_grep_test.go` — LSGF-001 word-boundary tests MUST stay green
- `shouldSkipCommitTitle` metadata-sweep + narrow backfill-skip behavior — unchanged
- All other SPEC directories (parallel-session plan-phase artifacts) — do NOT touch
- runtime-managed files (`.moai/state/`, `.moai/cache/`, `.moai/logs/`) — do NOT touch

---

## B. Known Issues (filtered for this Tier M Go-code SPEC)

Per template §B, B1-B8 filtered by domain relevance. This is `internal/spec` Go code touching a doctrine-mirror file. Relevant categories:

- **B2 — Cross-SPEC policy conflict scan**: predecessor DRIFT-CONVENTION-ALIGN-001 introduced `closeInfixMatch`/`isSyncPhaseDocs`/backfill-skip. This SPEC EXTENDS them; do NOT revert. Run `grep -n 'closeInfixMatch\|isSyncPhaseDocs\|specIDScopedChorePattern' internal/spec/transitions.go internal/spec/drift.go` before editing to map the existing seam.
- **B4 — Frontmatter canonical schema**: this SPEC's own frontmatter uses `created:`/`updated:`/`tags:` + `era: V3R6` + `tier: M` (no snake_case aliases). Verified.
- **B5 — CI 3-tier awareness**: spec-lint + golangci-lint + Test (per OS) each fail independently. The Go edits run on all OSes; the doctrine edits affect spec-lint mirror tests if the two doctrine files are template-mirrored (see §H).
- **B6 — spec-lint heading convention**: this spec.md uses `### Out of Scope — <X>` H3 sub-headings under §B (verified — avoids `MissingExclusions`).
- **B8 / B10 — working-tree hygiene + untouched-paths PRESERVE**: parallel sessions may be active (see memory `project_ccsync_2spec_4phase_close` — MEMORY-CONFIG-CLEANUP run in parallel). `git add` specific paths only; never `git add -A`. Do NOT touch other SPEC dirs.
- **B9 — git commit + push self-perform** (Hybrid Trunk): manager-develop commits per-milestone with Conventional Commits (`feat(SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001): M{N} …`), `🗿 MoAI` trailer, never `--no-verify`.
- **B11 — AskUserQuestion forbidden in subagent**: detector/lint rules are observation-only; never call AskUserQuestion. Blocker → structured report to orchestrator.

Not relevant (omit): B1 cross-platform build tags (no syscall), B3 C-HRA-008 subagent-boundary grep (drift.go is not a harness/hook domain — though the observation-only discipline of `internal/spec/CLAUDE.md` still applies), B7 observer.go capture path.

---

## C. Pre-flight Check List (run before any code change)

```bash
# 1. Baseline drift state (SSOT for the count assertion)
moai spec drift --json | jq '[.Records[] | select(.Drifted==true)] | length'   # expect ~51
moai spec drift --count                                                          # expect ~53 (live)

# 2. Baseline audit state (must stay identical post-fix, AC-DLC-007)
moai spec audit --json > /tmp/audit-baseline.json
jq '{grandfathered, modern_era_clean, n: (.drift_findings|length)}' /tmp/audit-baseline.json

# 3. Existing test baseline (must stay green)
go test ./internal/spec/... 2>&1 | tail -5
go build ./...

# 4. Map the existing classifier seam (predecessor's additions)
grep -n 'closeInfixMatch\|isSyncPhaseDocs\|specIDScopedChorePattern\|transitionRules' internal/spec/transitions.go internal/spec/drift.go

# 5. Confirm era.go consumption surface (READ-ONLY)
grep -n 'func ClassifyEra\|func LoadEraSignalsFromDir\|func (e Era) EraFinal' internal/spec/era.go

# 6. Verify whether the two doctrine files are template-mirrored (decides §H scope)
ls internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md 2>/dev/null && echo "MIRRORED: lifecycle-sync-gate" || echo "NOT mirrored: lifecycle-sync-gate"
ls internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md 2>/dev/null && echo "MIRRORED: spec-frontmatter-schema" || echo "NOT mirrored: spec-frontmatter-schema"

# 7. D1 CONFIRMATION — prove the per-SPEC window does NOT contain the combined-scope close,
#    but the scope-prefix window DOES (justifies the M3 secondary-grep design).
echo "--- per-SPEC window (close commit MUST be ABSENT) ---"
git log main --oneline --grep="SPEC-CCSYNC-CLAUDEMD-001" -10   # expect: sync da2fbcedf + feat a97206dc7 + 2 plan; NO 4-phase close
echo "--- scope-prefix window (close commit MUST be PRESENT) ---"
git log main --oneline --grep="SPEC-CCSYNC" -12 | grep -i "4-phase close"  # expect: cf7d78a9c chore(SPEC-CCSYNC): … 4-phase close (CLAUDEMD + TOOLCAT …)
```

---

## D. Constraints (DO NOT VIOLATE)

- **PRESERVE** the §A.5 enumeration: `era.go`, `audit.go`, `drift_chore_skip_test.go`, `drift_specid_grep_test.go`, the metadata-sweep + backfill skip behavior.
- **DO NOT modify `era.go` or `audit.go`** — consume READ-ONLY. AC-DLC-007 asserts byte-identical audit output; any audit.go/era.go change fails the AC.
- **DO NOT weaken the LSGF-001 word-boundary filter** (REQ-DLC-002 / AC-DLC-006 / AC-DLC-012). The combined-scope secondary prefix-grep fallback (①, M3) must use hyphen-delimited prefix boundary (`SPEC-{PREFIX}-`), NOT substring containment, AND must apply all 3 gates (FALLBACK-ONLY + `closeInfixMatch` + distinguishing-segment cross-check).
- **DO NOT infer `completed` from sync/feat/docs** after fixing ② — `completed` only via close-infix (REQ-DLC-004 / DRIFT-CONVENTION-ALIGN AP-2).
- **DO NOT silently clear genuine-⑤ drift** — mechanism ⑤ (no close commit) MUST remain reported (§B, §G AP-1).
- **Forbidden commands**: `--no-verify`, `--amend` on pushed commits, force-push to main, `git add -A`.
- **Required**: Conventional Commits, `🗿 MoAI` trailer, observation-only detector rules (no `os.Exec` write, no file mutation in the classifier path beyond reading git log).
- **Tier M delegation**: use the full Section A-E manager-develop prompt template.

---

## E. Self-Verification Deliverables (manager-develop completion report)

manager-develop MUST self-verify and report:

- **E1. AC Binary PASS/FAIL Matrix** — all 12 AC (AC-DLC-001..011 + AC-DLC-012 non-sibling-prefix collision guard) with verification command + actual output.
- **E2. Build** — `go build ./...` → exit 0.
- **E3. Coverage** — `go test -cover ./internal/spec/...` ≥ 85% (no regression on modified files).
- **E4. Observation-only discipline** — `grep -n 'os.Exec\|WriteFile\|os.Remove' internal/spec/drift.go internal/spec/transitions.go` shows only the existing read-only `git log` exec (no new write primitive).
- **E5. Lint** — `golangci-lint run --timeout=2m` → NEW issues explicitly reported; pre-existing baseline marked.
- **E6. Audit parity** — `moai spec audit --json` diffed against `/tmp/audit-baseline.json` → identical (AC-DLC-007).
- **E7. Drift delta** — `moai spec drift --count` before/after, with the named-exemplar transitions (AC-DLC-001..004) listed.
- **E8. Blocker report** (if any) — structured, no AskUserQuestion.

---

## F. Milestones (priority-ordered, no time estimates)

Milestone ordering is dependency-driven. M1 (era + terminal) and M2 (stale-rule) are independent; M3 (combined-scope) depends on the classifier seam being stable; M4 (doctrine) is independent of Go code; M5 (verification) is the gate.

### M1 — Era exemption + terminal-state authority in `drift.go` (Priority High) — REQ-DLC-005, REQ-DLC-006

The single largest false-positive reducer (mechanisms ③ + ④, ~17 records). Aligns `DetectDrift` with `audit.go` policy.

- In `DetectDrift`, after `ParseStatus(specDir)`, before the git compare:
  - Load era signals: `signals, _ := LoadEraSignalsFromDir(specDir)` then `era, _ := ClassifyEra(signals)`. If `era.EraFinal()` → **append a `DriftRecord` with `Drifted: false`** (do NOT `continue` before appending — preserve the record), then `continue`. (④ — grandfather exemption, mirrors `auditSpec`.) **D3 normative**: the record IS appended with `Drifted: false`; it is NOT dropped. This keeps the `drift.go:73-84` one-record-per-SPEC append pattern intact so `Records[]` consumers (M5 residual classifier) see the SPEC.
  - If `frontmatterStatus` is a terminal state (`superseded` / `archived` / `rejected`) → **append a `DriftRecord` with `Drifted: false`** (D3 normative — preserve the record), then `continue`. (③ — terminal authority, mirrors `checkV3R6Drift` early-return.)
  - **D3 contract reminder**: both branches construct the `DriftRecord` (with `GitImpliedStatus` set to whatever the inference yielded, or a sentinel like `"exempt"`/the frontmatter value), set `Drifted: false`, append, and `continue`. Driftcount is NOT incremented for these. NEVER skip the append — `Records[]` must contain every SPEC.
- **DDD ANALYZE note**: `DetectDrift` currently has fan_in = 1 (CLI `moai spec drift`). `getGitImpliedStatus` has fan_in = 2 (DetectDrift + StatusGitConsistencyRule). Place era/terminal logic in `DetectDrift` ONLY — do NOT alter `getGitImpliedStatus` (it is shared with the lint rule, which has its own grandfather handling via `Era.IsModern` upstream). Add `@MX:NOTE`/`@MX:REASON` on the new era-exemption branch.
- **Era subtlety (§G AP-3)**: rely on `ClassifyEra`'s H-1..H-4 progress.md signals as-computed; do NOT add a standalone `created`-date comparison (would misclassify `SPEC-V3R2-SPC-002` created 2026-04-23 > threshold).
- Tests: fixture-backed `drift_era_terminal_test.go` (NEW) — H-1 V2.x exemption, terminal-state early-return, V3R6-still-detected control.
- Commit: `feat(SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001): M1 era exemption + terminal-state authority in DetectDrift`

### M2 — Correct stale `sync`/`docs(sync)`→completed rule (Priority High) — REQ-DLC-003, REQ-DLC-004

Mechanism ② (~3 records). The 4-phase model: sync-phase = `implemented`, `completed` only via close-infix.

- In `transitions.go` `transitionRules`, change `{"docs(sync)", {"sync-merge", "completed"}}` and `{"sync", {"sync-merge", "completed"}}` → target status `"implemented"` (sync-merge category retained). The existing `isSyncPhaseDocs` already returns `implemented`; this aligns the legacy bare-`sync`/`docs(sync)` prefixes with it.
- **Guard**: verify `closeInfixMatch` still fires BEFORE these rules in `ClassifyPRTitle` (it does — close-infix is checked first, line ~122). So a `sync(...): … 4-phase close` still resolves `completed` (close-infix wins); a plain `sync(...): … implemented` now resolves `implemented`.
- **DCA preservation**: `isSyncPhaseDocs` (the predecessor's `docs(SPEC-XXX): … sync-phase` matcher) is unchanged — it already returns `implemented`. This M2 only corrects the two LEGACY bare prefixes that still said `completed`.
- Tests: extend `transitions_test.go` — `sync(SPEC-X-001): lifecycle complete — implemented` → `implemented`; `sync(SPEC-X-001): … 4-phase close` → `completed` (close-infix still wins); `docs(sync): legacy` → `implemented`.
- Commit: `fix(SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001): M2 correct legacy sync→completed rule to 4-phase implemented`

### M3 — Combined-scope sibling mapping via SECONDARY PREFIX-GREP FALLBACK (Priority Medium) — REQ-DLC-001, REQ-DLC-002

Mechanism ① (~2 records). The most delicate change — must NOT weaken LSGF-001. **REDESIGNED per plan-audit D1.**

> **D1 root cause (the original M3 design was architecturally dead)**: `getGitImpliedStatus` fetches its walker window via `git log main --grep=<full-specID>`. For `SPEC-CCSYNC-CLAUDEMD-001`, that window is `{da2fbcedf sync→implemented, a97206dc7 feat→implemented, 2 plan}` — the combined-scope close commit `cf7d78a9c chore(SPEC-CCSYNC): … 4-phase close (CLAUDEMD + TOOLCAT …)` is **NEVER fetched** because its subject names only `SPEC-CCSYNC`. An in-window gate (the original M3) can never see it. Empirically confirmed this session: `git log --grep=SPEC-CCSYNC-CLAUDEMD-001` does NOT contain `cf7d78a9c`; `git log --grep=SPEC-CCSYNC` DOES. Therefore the fix MUST run a SECONDARY grep with a broader key.

- **Fix design (minimum mechanism) — secondary scope-prefix grep fallback in `DetectDrift` (or a dedicated helper called from it)**:
  1. Run the per-SPEC primary walk (`getGitImpliedStatus(<full-specID>)`) UNCHANGED. If it returns `completed` or a terminal/aligned signal, done (no fallback).
  2. ONLY if the primary walk yields no `completed`/terminal signal (i.e. the SPEC would otherwise DRIFT against a frontmatter `completed`), derive the **scope-prefix** by stripping the trailing distinguishing-segment(s)+number: `SPEC-CCSYNC-CLAUDEMD-001` → `SPEC-CCSYNC`. (Strip the last `-<SEGMENT>-<NNN>` pair; if multiple distinguishing segments, strip only the final `-<SEGMENT>-<NNN>` — the scope-prefix is everything before it.)
  3. Run a secondary `git log main --grep=<scope-prefix> --oneline --no-merges -<N>`. Scan newest-first for a candidate combined-scope close commit whose subject: (a) has prefix `chore(SPEC-<PREFIX>)` / `docs(SPEC-<PREFIX>)` where `<PREFIX>` carries NO trailing `-NNN`, AND (b) satisfies `closeInfixMatch(subject) == true`, AND (c) the subject references THIS SPEC's distinguishing segment (e.g. `CLAUDEMD`) OR explicitly closes all siblings under the prefix.
  4. If a qualifying candidate is found → resolve this SPEC to `completed`. Otherwise → no fallback, the SPEC keeps its primary-walk status (and DRIFTs if genuinely incomplete — preserves mechanism ⑤).
- **New helper signatures (illustrative — minimum surface)**: `deriveScopePrefix(specID string) string` (returns `SPEC-CCSYNC` from `SPEC-CCSYNC-CLAUDEMD-001`), `combinedScopeCloseMatches(subject, specID string) bool` (gates (a)+(b)+(c)). The distinguishing-segment cross-check (c) parses the segment(s) stripped in step 2 and verifies the subject contains it (case-insensitive token match).
- **LSGF-001 preservation (REQ-DLC-002 — 3 tight gates)**: (i) FALLBACK-ONLY — the secondary grep runs ONLY after the exact-token primary walk fails to yield `completed`/terminal; the existing `commitMatchesSPECID` full-token path is UNCHANGED (additive only). (ii) `closeInfixMatch == true` REQUIRED on the candidate (gate b) — a bare combined-scope feat/chore is never a fallback hit. (iii) distinguishing-segment cross-check (gate c) — `SPEC-CCSYNC-CLAUDEMD-001` resolves only via a close naming `CLAUDEMD`, so it does NOT absorb an unrelated `SPEC-CCSYNC-OTHER-002` the close does not name. Prefix boundary is hyphen-delimited (`SPEC-<PREFIX>-`), so `SPEC-CCSYNC` scope does NOT apply to `SPEC-CCSYNCEXTRA-001` (different hyphen boundary — `deriveScopePrefix` and the `HasPrefix(specID, prefix+"-")` check both enforce this).
- **DDD ANALYZE note**: place the fallback in `DetectDrift` (fan_in=1, CLI-only), NOT in `getGitImpliedStatus` (fan_in=2, shared with the lint rule). The secondary grep is a NEW read-only `git log` exec — acceptable per observation-only discipline (no write primitive). Add `@MX:NOTE`/`@MX:REASON` on the fallback branch documenting the D1 rationale.
- Tests: NEW `drift_combined_scope_test.go`:
  - Deterministic fixture (AC-DLC-001 BINDING): `chore(SPEC-ABC): … 4-phase close (FOO + BAR)` + `feat(SPEC-ABC-FOO-001): M1` + `feat(SPEC-ABC-BAR-001): M1` → BOTH `SPEC-ABC-FOO-001` and `SPEC-ABC-BAR-001` resolve `completed` via the prefix-grep fallback.
  - Collision guard (AC-DLC-012): `SPEC-ABC-OTHER-002` (under same prefix, NOT named in close) → NOT `completed`; `SPEC-ABCD-001` (substring-only, different hyphen boundary) → scope-prefix grep does not apply.
  - Negative: combined-scope WITHOUT close-infix → no fallback hit.
  - Real-repo integration: `SPEC-CCSYNC-CLAUDEMD-001` + `SPEC-CCSYNC-TOOLCAT-001` both absent from DRIFT.
- Commit: `feat(SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001): M3 combined-scope secondary prefix-grep fallback (LSGF-001-safe, 3-gate)`

### M4 — Close-subject doctrine amendment + template mirror (Priority Medium) — REQ-DLC-011

Mechanism ① recurrence prevention (doctrine). Independent of Go code.

- Amend the Status Transition Ownership Matrix in BOTH:
  - `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (§ Status Transition Ownership Matrix)
  - `.claude/rules/moai/development/spec-frontmatter-schema.md` (§ Status Transition Ownership Matrix)
- Amendment text: mandate individual full-IDs in close-commit subjects. Add a HARD constraint: close commits MUST carry one full SPEC-ID (`chore(SPEC-{FULL-ID}): … 4-phase close`); combined/abbreviated scope (`chore(SPEC-CCSYNC): … (CLAUDEMD + TOOLCAT)`) is prohibited because the drift detector's exact-token extraction cannot map an abbreviated prefix to siblings. Cross-reference this SPEC.
- **Mirror obligation (§H)**: if M4 pre-flight (step 6) confirms either doctrine file is template-mirrored under `internal/template/templates/.claude/rules/...`, update the mirror to byte-consistency and run the mirror-drift test (`go test ./internal/template/... -run TestRuleTemplateMirror` or equivalent). If NOT mirrored, no mirror edit needed (note this in the commit body).
- **Template-neutrality (§25)**: the doctrine amendment is generic mechanism prose — the close-subject convention is a permanent rule, not internal dev state. Acceptable per C2/C5. The SPEC-ID cross-reference (`SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001`) is in the SOURCE rule file (acceptable), and MUST be generalized or omitted in the template mirror per §25 forbidden-class C2 (internal SPEC IDs). Run-phase: in the source file cite the SPEC ID; in the mirror, use generic phrasing (e.g. "per the drift-detector close-subject convention").
- Tests: AC-DLC-011 grep assertion (both files contain the amendment) + mirror byte-parity (if mirrored).
- Commit: `docs(SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001): M4 close-subject full-ID doctrine amendment + mirror`

### M5 — Verification gate + drift residual classification (Priority High) — REQ-DLC-009, REQ-DLC-010, all AC

- Run full `go test ./internal/spec/...` (AC-DLC-010), `moai spec audit` parity (AC-DLC-007), `moai spec drift --count` delta (AC-DLC-009).
- Classify the post-fix residual: confirm remaining drift records are genuine-⑤ (no close commit). Produce the deferred-⑤ list for the follow-up SPEC (informational — do NOT remediate).
- The no-false-positive-on-grandfathered-sibling subtest (AC-DLC-008) is part of M1/M3 test files; M5 confirms the aggregate.
- Commit (if separate): `test(SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001): M5 verification gate + genuine-⑤ residual classification`

---

## G. Anti-Patterns (explicit anti-goals)

- **AP-1 — clearing genuine-⑤ drift**: any change that makes the drift walker report a terminal/completed status for a modern V3R6 SPEC lacking the corresponding close commit MASKS a real lifecycle gap. The era/terminal/combined-scope fixes target ①②③④ ONLY; ⑤ must remain reported.
- **AP-2 — inferring `completed` without close-infix** (inherited from DRIFT-CONVENTION-ALIGN): after M2 corrects the stale sync rule, NO sync/feat/docs commit may resolve `completed` — close-infix is the sole positive `completed` signal.
- **AP-3 — era classification via `created` date alone**: `SPEC-V3R2-SPC-002` has `created: 2026-04-23` > `modernEraThreshold` (2026-04-01). Relying on the date would misclassify it as V3R6 and re-introduce the drift. The era exemption MUST use `ClassifyEra`'s full H-1..H-4 progress.md signal chain (H-1 progress.md-absent → V2.x; H-2 no §E.* → V3R2-R4) which fires BEFORE the H-5 date tie-breaker.
- **AP-4 — combined-scope substring mapping**: mapping `SPEC-CCSYNC` to siblings via substring containment (not hyphen-delimited prefix) would weaken LSGF-001 and map to unrelated SPECs (`SPEC-CCSYNCEXTRA-001`). Use `deriveScopePrefix` + `HasPrefix(specID, prefix+"-")` hyphen-delimited boundary.
- **AP-5 — over-broad combined-scope mapping / missing distinguishing-segment check**: resolving a SPEC to `completed` from a combined-scope close WITHOUT verifying the close names this SPEC's distinguishing segment (gate c) would over-attribute — `SPEC-CCSYNC-OTHER-002` would be falsely cleared by a `(CLAUDEMD + TOOLCAT)` close that does not close it. The 3 gates (FALLBACK-ONLY + `closeInfixMatch` + distinguishing-segment cross-check) are all mandatory.
- **AP-6 — modifying era.go / audit.go**: these are the REFERENCE. AC-DLC-007 asserts byte-identical audit output. Any edit there is out of scope and breaks the AC.
- **AP-7 — in-window-only combined-scope gate (the dead D1 design)**: gating combined-scope mapping on the in-window commits of the per-SPEC `git log --grep=<full-specID>` walk is architecturally dead — the combined-scope close commit is NEVER in that window (it names only the scope-prefix). The fix MUST use a SECONDARY `git log --grep=<scope-prefix>` fetch (M3). An in-window-only gate silently does nothing for the motivating CCSYNC case and would make AC-DLC-001 fail at run-phase.

---

## H. Cross-References + Mirror Obligation

### Mirror obligation (template-mirror)

The Go code in `internal/spec/` is NOT template content (no mirror). However M4's doctrine edits touch `.claude/rules/`, which MAY be template-mirrored under `internal/template/templates/.claude/rules/`. Run-phase pre-flight (§C step 6) determines whether `lifecycle-sync-gate.md` and `spec-frontmatter-schema.md` are mirrored:

- If mirrored → update the mirror to byte-consistency, generalize the SPEC-ID cross-reference per §25 forbidden-class C2 (internal SPEC IDs prohibited in template), run the mirror-drift test.
- If NOT mirrored → no mirror edit; note in M4 commit body.

### Cross-references

- `.moai/specs/SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001/` — predecessor (close-infix + backfill-skip; named this SPEC as the follow-up)
- `internal/spec/drift.go`, `internal/spec/transitions.go` — run-phase Go targets
- `internal/spec/era.go`, `internal/spec/audit.go` — READ-ONLY reference (consume `ClassifyEra` / terminal pattern)
- `internal/spec/drift_chore_skip_test.go`, `internal/spec/drift_specid_grep_test.go` — regression guards (MUST stay green)
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — era heuristic + grandfather + Ownership Matrix (M4 amend target)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Ownership Matrix (M4 amend target)
- `internal/spec/CLAUDE.md` — module conventions (observation-only, table-driven, sibling PRESERVE)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E delegation (Tier M REQUIRED)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — Phase 0.95 mode (Tier M Go-code → Mode 5 sub-agent default per Finding A4 coding-task caveat)
