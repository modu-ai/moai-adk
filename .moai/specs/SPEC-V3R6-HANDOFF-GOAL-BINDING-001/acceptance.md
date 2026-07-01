# Acceptance Criteria — SPEC-V3R6-HANDOFF-GOAL-BINDING-001

All AC are mechanically checkable via grep/build. Paths: LIVE = `.claude/...`; TMPL = `internal/template/templates/.claude/...`.

## §D. AC Matrix

| AC ID | REQ | Check | Pass condition |
|-------|-----|-------|----------------|
| AC-HGB-001 | REQ-HGB-001 | `/goal` line present in SSOT Block 1 | grep finds `/goal` in `session-handoff.md` Block 1 skeleton + Field-by-Field spec |
| AC-HGB-002 | REQ-HGB-001/005 | `/goal` line present in render §8 | grep finds `/goal` in `moai.md` §8 6-block skeleton |
| AC-HGB-003 | REQ-HGB-002 | Emit-trigger text present | grep finds run-phase AND verifiable trigger + "Default on ambiguity" omit rule in `session-handoff.md` |
| AC-HGB-004 | REQ-HGB-003 | Diet single-line item present | grep finds a `/goal` item in the Diet paste-ready-budget self-check; item-count label updated |
| AC-HGB-005 | REQ-HGB-004 | Kickoff-Approval invariant present | grep finds explicit "`/goal` does NOT authorize autonomous run-phase entry" text |
| AC-HGB-006 | REQ-HGB-005 | SSOT↔render Template-First + concern-qualifier parity | see §D.1 parity greps |
| AC-HGB-007 | REQ-HGB-006 | Localization 4-column parity — per-surface, per-locale | Each of en/ko/ja/zh present in the localization/header-translation region of BOTH surfaces (session-handoff.md SSOT AND moai.md §8) — 8 assertions (4 locales × 2 surfaces), all PASS; `/goal`/`✂`/`─` verbatim |
| AC-HGB-008 | REQ-HGB-007 | Omission anti-pattern present | grep finds `/goal` omission anti-pattern in `session-handoff.md` Anti-Patterns |
| AC-HGB-009 | REQ-HGB-008 | goal-directive cross-ref NEW delta = literal `Block 1` | grep finds the literal token **`Block 1`** co-located with `/goal` in `goal-directive.md` `ultrathink.` resume-pairing bullet. Baseline: 0 `Block 1` occurrences pre-edit — pinning `Block 1` prevents a vacuous pass on pre-existing `/goal` + `session-handoff.md` + re-set-after-`/clear` text |
| AC-HGB-010 | REQ-HGB-009 | Template-First mirror parity | per-file: LIVE `/goal` grep count == TMPL `/goal` grep count (all 4 files incl. context-window-management.md conditional) |
| AC-HGB-011 | REQ-HGB-010 | context-window-management non-contradiction | grep confirms resume section refs `session-handoff.md` SSOT; no contradicting inline `/goal` claim |
| AC-HGB-012 | REQ-HGB-009 | Re-embed clean | `make build` succeeds AND `go build ./...` passes |
| AC-HGB-013 | REQ-HGB-003/005 | Count-label == enumeration | In each self-check list, the bumped "N items" label integer equals the enumerated item count (SSOT paste-ready-budget 8→9 `- [ ]` items; render session-handoff-template-completeness 9→10 `;`-clauses) |
| AC-HGB-014 | REQ-HGB-009 | session-handoff.md template neutrality preserved | `session-handoff.md` TEMPLATE copy has 0 language-specific tokens (`go test`/`pytest`/`cargo`); baseline was 0 `go test` — the `/goal` Example edit must not introduce the first one |

## §D.1 Parity grep checks (canonical commands)

```bash
# AC-HGB-001 / AC-HGB-002 — /goal present in both surfaces
grep -n '/goal' .claude/rules/moai/workflow/session-handoff.md
grep -n '/goal' .claude/output-styles/moai/moai.md

# AC-HGB-006 (D2 mechanized) — assert EXACTLY 3 distinct concern-name qualifiers across
# BOTH surfaces. Extracts every `Pre-emit self-check (<qualifier>)` label, drops the numeric
# "(N items)" form, dedupes. A NEW 4th qualifier introduced by the /goal edit makes the count
# 4 → FAIL. (The old per-known-qualifier loop iterated only the 3 known strings and structurally
# could not detect a 4th.)
distinct=$(for s in .claude/rules/moai/workflow/session-handoff.md .claude/output-styles/moai/moai.md; do
             grep -ohE 'Pre-emit self-check \([a-z][^)]*\)' "$s"
           done | grep -vE '\([0-9]+ items?\)' | sed -E 's/.*\((.*)\)/\1/' | sort -u)
echo "$distinct"
count=$(printf '%s\n' "$distinct" | grep -c .)
[ "$count" -eq 3 ] && echo "QUALIFIER-COUNT-OK (3)" || echo "QUALIFIER-COUNT-FAIL ($count — a 4th qualifier was introduced)"
# Pass: count == 3 (localization render / paste-ready budget / session-handoff template completeness).

# AC-HGB-010 (D5 guard-add) — Template-First mirror parity: LIVE count == TMPL count, per file.
# context-window-management.md is included for the D.4 CONDITIONAL-edit case: if its edit fired
# (LIVE /goal count > 0) TMPL must match; if not edited both are 0 and parity holds trivially — so
# REQ-HGB-009's "every file this SPEC edits" is not violated should the D.4 contingent edit fire.
for f in \
  .claude/rules/moai/workflow/session-handoff.md \
  .claude/output-styles/moai/moai.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/context-window-management.md ; do
  live=$(grep -c '/goal' "$f")
  tmpl=$(grep -c '/goal' "internal/template/templates/$f")
  echo "$f  LIVE=$live  TMPL=$tmpl  $( [ "$live" = "$tmpl" ] && echo PARITY-OK || echo PARITY-FAIL )"
done

# AC-HGB-007 (D1 per-surface, per-locale) — assert each of en/ko/ja/zh appears in the
# localization/header-translation region of BOTH surfaces (8 assertions). The old single-surface
# OR-alternation would pass even if a locale column were silently dropped in moai.md.
# Locale representative tokens = the cut-line marker translations present in both surfaces.
loc_fail=0
for s in .claude/rules/moai/workflow/session-handoff.md .claude/output-styles/moai/moai.md; do
  echo "== $s =="
  grep -q 'Copy from here' "$s" && echo "  en OK" || { echo "  en FAIL"; loc_fail=1; }
  grep -q '여기부터 복사'   "$s" && echo "  ko OK" || { echo "  ko FAIL"; loc_fail=1; }
  grep -q 'ここからコピー'  "$s" && echo "  ja OK" || { echo "  ja FAIL"; loc_fail=1; }
  grep -q '从这里复制'      "$s" && echo "  zh OK" || { echo "  zh FAIL"; loc_fail=1; }
done
[ "$loc_fail" -eq 0 ] && echo "LOCALE-PARITY-OK (4 locales × 2 surfaces)" || echo "LOCALE-PARITY-FAIL"

# AC-HGB-013 (D4) — count-label integer == enumerated item count, per self-check list.
# SSOT Diet "paste-ready budget" list: label integer (— N items) == number of '- [ ]' checkbox items.
ssot=.claude/rules/moai/workflow/session-handoff.md
label_n=$(grep -oE 'Pre-emit self-check \(paste-ready budget\) — [0-9]+ items' "$ssot" | grep -oE '[0-9]+' | head -1)
enum_n=$(awk '/Pre-emit self-check \(paste-ready budget\)/{f=1;next} /^### /{f=0} f&&/^- \[ \]/{c++} END{print c+0}' "$ssot")
echo "SSOT budget: label=$label_n enum=$enum_n  $( [ "$label_n" = "$enum_n" ] && echo OK || echo DRIFT )"
# render "session-handoff template completeness" list: run-phase asserts the render '(N items)'
# label integer equals the count of ';'-separated clauses in the 'Covers:' enumeration bullet.
# Pass: both label==enum (SSOT 9==9 after 8→9 bump; render 10==10 after 9→10 bump).

# AC-HGB-014 (D6) — session-handoff.md TEMPLATE language-neutrality preserved.
# Baseline: 0 'go test' occurrences (verified pre-edit). The /goal Example edit must not introduce one.
tmpl=internal/template/templates/.claude/rules/moai/workflow/session-handoff.md
n=$(grep -cE 'go test|pytest|cargo (test|build)|npm test' "$tmpl")
[ "$n" -eq 0 ] && echo "TEMPLATE-NEUTRALITY-OK (0 language-specific tokens)" || echo "TEMPLATE-NEUTRALITY-FAIL ($n introduced)"

# AC-HGB-009 (D7) — goal-directive.md NEW delta: literal 'Block 1' co-located with /goal
grep -n 'Block 1' .claude/rules/moai/workflow/goal-directive.md   # baseline 0 pre-edit; expect ≥1 post-edit near /goal

# AC-HGB-012 — re-embed clean
make build && go build ./...
```

## §D.2 Given-When-Then Scenarios

### Scenario 1 — Run-phase SPEC with verifiable condition emits the `/goal` line

- **Given** the orchestrator is emitting a paste-ready resume message, and the next SPEC enters run-phase with a mechanically verifiable completion condition (`go test ./... exits 0 AND golangci-lint clean, or stop after 20 turns`),
- **When** Block 1 is rendered,
- **Then** Block 1 contains a single `/goal <completion-condition>` line (parallel to the `/effort ultracode` line), the condition text is on one line, and the message does NOT imply run-phase is pre-authorized.

### Scenario 2 — Plan/sync-phase or non-verifiable next SPEC omits the `/goal` line

- **Given** the orchestrator is emitting a resume message, and the next SPEC is a plan-phase or sync-phase SPEC (no machine-verifiable end-state), OR the phase/verifiability is ambiguous,
- **When** Block 1 is rendered,
- **Then** the `/goal` line is omitted entirely (the `ultrathink.` opener alone suffices), matching the `/effort ultracode` default-on-ambiguity-omit discipline.

### Scenario 3 — SSOT↔render + Template-First parity holds after the edit

- **Given** the run-phase has applied the `/goal` binding,
- **When** the §D.1 parity greps run,
- **Then** the `/goal` binding is present in both SSOT and render surface, each edited file's LIVE `/goal` count equals its TMPL count, exactly the 3 existing concern-name qualifiers remain (no new qualifier), and `make build` + `go build ./...` succeed.

### Scenario 4 — `/goal` line does not bypass Implementation Kickoff Approval

- **Given** a resumed session whose pasted resume Block 1 contains a `/goal` line,
- **When** the orchestrator reaches the run-phase entry point,
- **Then** the orchestrator still asks for explicit run-phase approval via `AskUserQuestion` (Implementation Kickoff Approval), because the binding text explicitly states the `/goal` line does not authorize autonomous run-phase entry.

## §D.3 Edge Cases

- **Ambiguous verifiability** → omit (default), never emit a speculative `/goal` line.
- **Multi-condition** → keep on one line joined with `AND`/`, or stop after N turns`; do NOT expand into a multi-line block (Diet REQ-HGB-003).
- **Locale = ja/zh/fr** → `/goal` literal preserved verbatim; only surrounding label follows localization; fr (not in table) falls back to English skeleton per the existing Localization fallback rule.

## §D.4 Definition of Done

- [ ] All 14 AC PASS.
- [ ] `/goal` binding present + identical in SSOT (`session-handoff.md`) and render (`moai.md §8`), both LIVE and TMPL.
- [ ] Diet single-line constraint honored; item-count labels updated on both self-check lists AND label integer == enumerated item count (AC-HGB-013).
- [ ] Kickoff-Approval invariant text present.
- [ ] Omission anti-pattern present.
- [ ] `goal-directive.md` cross-reference strengthened AND pins the literal `Block 1` NEW delta (AC-HGB-009).
- [ ] `context-window-management.md` confirmed non-contradicting (edited only if needed); included in the Template-First mirror-parity loop for the conditional-edit case (AC-HGB-010/D5).
- [ ] No new pre-emit self-check concern-name qualifier introduced — mechanically verified distinct-qualifier count == 3 (AC-HGB-006/D2).
- [ ] Localization per-surface per-locale parity — 8 assertions (4 locales × 2 surfaces) PASS (AC-HGB-007/D1).
- [ ] `session-handoff.md` template neutrality preserved — 0 language-specific tokens (AC-HGB-014/D6).
- [ ] `make build` + `go build ./...` clean.

## Quality Gate

- spec-lint: frontmatter 12-field schema valid; `OutOfScopeRule` satisfied (spec.md §C has `### Out of Scope —` H3 sub-headings with `-` bullets); GEARS syntax present.
- Doc-only change: no test-suite / coverage delta expected.
