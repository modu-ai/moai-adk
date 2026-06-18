# Acceptance Criteria — SPEC-V3R6-RULES-SSOT-DEDUP-001

Every de-dup AC follows the canonical triple: **(absent)** duplicated block gone from the
non-owner, **(pointer)** a cross-reference pointer remains, **(intact)** the SSOT owner still
holds the authoritative text. The file-deletion AC adds a 3rd clause (content survives + no
dangling referrer). All grep commands are run-phase evidence (recorded in progress.md §E.2).

### Distinctive-line delta convention (D5/D6 fix — required for every "(absent)" clause)

[HARD] A pointer-presence grep is INSUFFICIENT as an "(absent)" assertion: pointers are often
ALREADY present before the edit (verified live: agent-common-protocol.md has 2 `askuser-protocol.md`
mentions and moai-constitution.md has 1, BEFORE any edit), so `grep -c "<owner>.md" >= 1` proves
nothing. Every "(absent)" clause MUST instead assert a genuine **before(N)/after(<N or 0)** delta on
a DISTINCTIVE line of the duplicated block — a verbatim string that appears ONLY inside the
copied block, not in the surrounding prose. Each AC below records the BASELINE count (measured at
run-phase pre-flight) and the EXPECTED post-edit count. The AC-SSD-005 model (asserting
`(none) → draft == 0` as a true before(1)/after(0) drop) is the canonical pattern; the pointer-presence
grep is retained ONLY as a secondary "(pointer)" check, never as the "(absent)" proof.

### Mirror-class verification convention (D1/D2 fix — pick the AC per file's actual class)

[HARD] `TestRuleTemplateMirrorDrift` enrolls ONLY the `workflowOptMirroredPaths` byte-parity
allowlist. Live ground truth: of the 20 in-scope files, **only `hooks-system.md` is enrolled**.
Asserting "`TestRuleTemplateMirrorDrift` PASS" for a NON-enrolled file is VACUOUSLY green — a
missed mirror edit would still pass because the test never inspects that file. Each AC therefore
picks its mirror-verification clause from this table (re-confirm membership at run-phase pre-flight,
since the allowlist is the ground truth):

| Mirror class | Files (in scope) | Mirror AC clause to use |
|--------------|------------------|--------------------------|
| **byte-parity allowlist** (test-enrolled) | `hooks-system.md` ONLY | `go test ./internal/template/... -run TestRuleTemplateMirrorDrift` PASS (genuinely catches drift) |
| **byte-identical-by-discipline** (NOT enrolled — identical today, must stay identical) | moai-constitution, settings-management, context-window-management, spec-frontmatter-schema, team-protocol, team-pattern-cookbook, worktree-integration, skill-authoring, skill-writing-craft, skill-ab-testing, agent-patterns, orchestrator-templates, agent-authoring | explicit per-file `diff -q <deployed> <mirror>` → exit 0 (else the AC is vacuously green) |
| **§25-sanitized** (NOT byte-identical) | agent-common-protocol (17 tok), verification-batch-pattern (2 tok), agent-teams-pattern (1 tok, deleted) | `go test ./internal/template/... -run TestTemplateNoInternalContentLeak` PASS + a scoped `diff` showing ONLY the de-dup edit changed (not new divergence) |
| **§25-divergent registry** (13-line CONST-V3R6-001 delta, NOT enrolled, NOT in leak-comment list but diverges) | zone-registry | `TestTemplateNoInternalContentLeak` PASS + `diff <deployed> <mirror>` confirming ONLY the de-dup `clause:`/`paths:` edit changed (the pre-existing 13-line delta is unchanged) |
| **no mirror** (deployed-only) | lifecycle-sync-gate | `test ! -e internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md` |

---

## §D. AC Matrix

### AC-SSD-001 — AskUserQuestion 4-way de-dup (REQ-SSD-001)

**Given** `core/askuser-protocol.md` is the AskUserQuestion SSOT,
**When** the de-dup is applied,
**Then** all of:
- (absent — distinctive-line delta) `core/agent-common-protocol.md` §User Interaction Boundary no
  longer restates the full ToolSearch-preload / Socratic-interview / option-standards clauses.
  Run-phase pre-flight MUST first BASELINE a distinctive line that appears ONLY in the restated
  block (a verbatim clause that is NOT the pointer text and NOT in the locally-unique delta), then
  assert it drops to 0 post-edit. Example distinctive markers to baseline (run-phase confirms which
  are present): the verbatim `select:AskUserQuestion` preload string, the `(권장)` / `(Recommended)`
  first-option-label clause, or the `max 4 questions per round` Socratic constraint:
  ```bash
  # BASELINE (pre-edit, expected >=1): pick a distinctive restated-clause string actually present
  grep -c 'select:AskUserQuestion' .claude/rules/moai/core/agent-common-protocol.md   # baseline N
  # AFTER edit: the restated copy is gone (the SSOT owner keeps it)
  grep -c 'select:AskUserQuestion' .claude/rules/moai/core/agent-common-protocol.md   # expect 0 (or only the pointer line, not the full restatement)
  ```
  (If a chosen marker also legitimately appears in the retained local delta, pick a different
  distinctive marker — the run-phase delegation MUST record the exact baseline string + count.)
- (pointer — secondary check) a pointer line to `askuser-protocol.md` is present in BOTH files (this
  is a secondary check; pointers are ALREADY present pre-edit so this alone does NOT prove absence):
  ```bash
  grep -c "askuser-protocol.md" .claude/rules/moai/core/agent-common-protocol.md   # >= 1 (was 2 pre-edit)
  grep -c "askuser-protocol.md" .claude/rules/moai/core/moai-constitution.md       # >= 1 (was 1 pre-edit)
  ```
- (delta survives) the subagent-prohibition framing remains in agent-common-protocol; the
  orchestrator-obligation framing remains in moai-constitution:
  ```bash
  grep -q "Subagents MUST NOT" .claude/rules/moai/core/agent-common-protocol.md    # delta intact
  ```
- (intact) `core/askuser-protocol.md` is unchanged as the owner (no content removed).
- (mirror — per class) agent-common-protocol is §25-sanitized → `TestTemplateNoInternalContentLeak`
  PASS + scoped `diff` showing only the de-dup edit changed; moai-constitution is
  byte-identical-by-discipline (NOT allowlist-enrolled) → explicit
  `diff -q .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md`
  → exit 0.
- (forward-item) CLAUDE.md was NOT edited:
  ```bash
  git diff --name-only | grep -qx "CLAUDE.md" && echo VIOLATION || echo OK   # expect OK
  ```

### AC-SSD-002 — Hooks config/timeout de-dup + contradiction reconcile (REQ-SSD-002)

**Given** `core/hooks-system.md` is the hooks SSOT,
**When** the de-dup + reconciliation is applied,
**Then** all of:
- (absent — distinctive-line delta) `core/settings-management.md` §Hooks Configuration no longer
  carries the full hook JSON config block. Baseline a distinctive line of the copied hook JSON
  (e.g. the `handle-session-start.sh` wrapper path or the `$CLAUDE_PROJECT_DIR` quoting line that
  belongs to the config block, NOT the StatusLine delta) and assert it drops post-edit:
  ```bash
  # BASELINE (pre-edit): distinctive hook-JSON-config line present in settings-management
  grep -c 'handle-session-start.sh' .claude/rules/moai/core/settings-management.md   # baseline N
  # AFTER edit: the duplicated config block is gone (settings-management keeps only the StatusLine delta + pointer)
  grep -c 'handle-session-start.sh' .claude/rules/moai/core/settings-management.md   # expect < N (ideally 0)
  ```
- (pointer — secondary) a pointer to `hooks-system.md` is present:
  ```bash
  grep -c "hooks-system.md" .claude/rules/moai/core/settings-management.md   # >= 1
  ```
- (reconciled timeout — BOTH the JSON example AND the L323 timeout TABLE) hooks-system.md is
  internally consistent after the edit: the timeout TABLE (the row mapping
  `SessionStart, PreCompact, PreToolUse, PostToolUse` to a default) MUST be edited to separate
  PostToolUse from the 5s synchronous group and state it as the 10s+`async` exception, so the table
  no longer contradicts the JSON example (which already shows `timeout: 10` + `async: true`).
  Baseline the contradiction, assert it is resolved:
  ```bash
  # BASELINE the contradiction: the table groups PostToolUse under 5s while the JSON example uses 10s+async
  sed -n '320,326p' .claude/rules/moai/core/hooks-system.md   # pre-edit: PostToolUse in the 5s row
  # AFTER edit: PostToolUse appears as the 10s + async exception in BOTH the table AND the JSON example,
  # and the 5s default is documented as the synchronous-hook default with rationale.
  grep -n 'async' .claude/rules/moai/core/hooks-system.md | grep -qi 'PostToolUse\|10'   # async exception stated
  ```
- (reconciled alwaysLoad — specific introduction line, not a bare version sweep) settings-management.md
  states the `alwaysLoad` INTRODUCTION version exactly once; the v2.1.121 mention (if it refers to the
  separate `updatedToolOutput` extension, NOT the alwaysLoad introduction) is labeled distinctly.
  Do NOT use a bare `grep -o 'v2.1.1xx' | sort -u` (that legitimately yields 4 unrelated CC versions —
  v2.1.119/120/121 hook-change rows are valid). Instead grep the specific alwaysLoad introduction line:
  ```bash
  # the alwaysLoad introduction is attributed to ONE version (canonical: v2.1.119, the earlier)
  grep -n 'alwaysLoad' .claude/rules/moai/core/settings-management.md | grep -iE 'v2.1.1[0-9][0-9]'
  # expect: exactly one line attributing alwaysLoad INTRODUCTION to a single version; any v2.1.121 line
  # near alwaysLoad refers to the updatedToolOutput EXTENSION, labeled as such — not a 2nd introduction.
  ```
- (intact) hooks-system.md retains the full timeout table + JSON config (now internally consistent).
- (mirror) hooks-system.md mirror is **byte-identical** (it is the ONLY in-scope file in the
  `workflowOptMirroredPaths` byte-parity allowlist, so this test genuinely catches drift):
  ```bash
  go test ./internal/template/... -run TestRuleTemplateMirrorDrift   # PASS
  ```

### AC-SSD-003 — Context-window threshold 5-way de-dup (REQ-SSD-003)

**Given** `workflow/context-window-management.md` is the threshold SSOT,
**When** the de-dup is applied,
**Then**:
- (absent + pointer) each of the 3 non-owner copies carries a pointer and no longer restates the
  full per-model threshold table:
  ```bash
  for f in core/settings-management.md core/zone-registry.md development/skill-ab-testing.md; do
    grep -c "context-window-management.md" .claude/rules/moai/$f   # >= 1 each
  done
  ```
- (zone-registry safety) the CONST-V3R5-022 edit is a body/pointer edit, NOT a `clause:` blanking —
  `moai constitution validate` still passes (no new SentinelDrift):
  ```bash
  moai constitution validate --format json | grep '"drift_count"'   # 0 (or unchanged baseline)
  ```
- (untouched) `workflow/session-handoff.md` is NOT in the diff (already defers).
- (intact) context-window-management.md retains the authoritative table.
- (mirror — per class) settings-management + skill-ab-testing are byte-identical-by-discipline (NOT
  enrolled) → explicit `diff -q <deployed> <mirror>` exit 0 each; zone-registry is §25-divergent →
  `TestTemplateNoInternalContentLeak` PASS + `diff` confirming ONLY the CONST-V3R5-022 pointer edit
  changed (the pre-existing 13-line CONST-V3R6-001 delta unchanged):
  ```bash
  diff -q .claude/rules/moai/core/settings-management.md internal/template/templates/.claude/rules/moai/core/settings-management.md
  diff -q .claude/rules/moai/development/skill-ab-testing.md internal/template/templates/.claude/rules/moai/development/skill-ab-testing.md
  # zone-registry: leak-test + scoped diff (delta count stays at the pre-existing 13 + the de-dup line)
  go test ./internal/template/... -run TestTemplateNoInternalContentLeak
  ```

### AC-SSD-004 — 7-item verification batch: confirm-defer + re-sync sentinel (REQ-SSD-004, RE-TARGETED per D4)

**D4 premise correction**: `workflow/verification-batch-pattern.md` has **NO verbatim 7-command code
block to remove** — research confirmed it ALREADY correctly DEFERS to agent-common-protocol.md
§Parallel Execution (the only `coverprofile=cover.out` match in the file is the legitimately-retained
anti-pattern race-note prose at line 29: "Same-file writes (two `coverprofile=cover.out` runs race)").
T4 is therefore largely a **verified no-op**, re-targeted from "removal" to "confirm-defer + add a
re-sync sentinel note." The original `grep -c "coverprofile=cover.out" ... # 0` clause is REMOVED — it
would FALSE-FAIL on the retained race-note prose.

**Given** `core/agent-common-protocol.md` §Parallel Execution holds the verbatim 7-command block,
**When** the (re-targeted) de-dup is applied,
**Then**:
- (no removal needed — verified defer) verification-batch-pattern.md is confirmed to NOT carry a
  duplicate 7-command block. The single `coverprofile=cover.out` occurrence is the L29 race-note
  prose, which is RETAINED (it is an anti-pattern example, not a copy of the batch):
  ```bash
  grep -n 'coverprofile=cover.out' .claude/rules/moai/workflow/verification-batch-pattern.md
  # expect exactly 1 match at ~L29 ("Same-file writes ... runs race") — this is RETAINED, not a violation
  ```
- (re-sync sentinel added) verification-batch-pattern.md carries an explicit "re-sync if the
  7-item list changes" sentinel note pointing at agent-common-protocol §Parallel Execution (this is
  the only substantive edit for T4):
  ```bash
  grep -c "agent-common-protocol.md" .claude/rules/moai/workflow/verification-batch-pattern.md   # >= 1 (pointer/sentinel present)
  grep -iq 're-sync\|re-sync if the 7-item' .claude/rules/moai/workflow/verification-batch-pattern.md   # sentinel note present
  ```
- (intact — per-keyword loop, NOT a vacuous line-count) agent-common-protocol §Parallel Execution
  still has all 7 DISTINCT commands. The original `grep -c "<7 alternation>"` returned 17 (a vacuous
  line-count that never equals 7); replace with a per-keyword loop asserting EACH of the 7 distinct
  commands is present ≥ 1:
  ```bash
  f=.claude/rules/moai/core/agent-common-protocol.md
  for kw in 'go test' 'coverprofile' 'grep ' 'sentinel' 'cmd/moai' 'bench' 'lint'; do
    c=$(grep -c "$kw" "$f")
    [ "$c" -ge 1 ] && echo "OK   $kw ($c)" || echo "FAIL $kw (0)"
  done
  # expect 7 OK lines, zero FAIL
  ```
- (mirror) both files are §25-sanitized → `TestTemplateNoInternalContentLeak` PASS (verification-batch
  carries 2 stripped tokens; agent-common-protocol carries 17) + scoped `diff` showing only the
  sentinel-note edit changed in verification-batch-pattern.md.

### AC-SSD-005 — Status Transition Ownership Matrix de-dup (REQ-SSD-005)

**Given** `development/spec-frontmatter-schema.md` is the matrix SSOT,
**When** the de-dup is applied,
**Then**:
- (absent — canonical distinctive-line delta, the model for all other "(absent)" clauses)
  `workflow/lifecycle-sync-gate.md` no longer reproduces the full 7-row transition matrix table. The
  distinctive line `(none) → draft` appears in the copied matrix row; assert its genuine before(1)/
  after(0) drop (verified live: lifecycle-sync-gate.md L225 currently has this row):
  ```bash
  grep -c "(none) → draft" .claude/rules/moai/workflow/lifecycle-sync-gate.md   # BASELINE 1 → AFTER 0 (full matrix row gone)
  grep -c "spec-frontmatter-schema.md" .claude/rules/moai/workflow/lifecycle-sync-gate.md   # >= 1 (pointer — secondary check)
  ```
- (intact) spec-frontmatter-schema.md holds the canonical matrix (SSOT owner unchanged unless it
  lacked the close-subject one-liner; if edited, diff its mirror per below).
- (mirror — spec-frontmatter-schema only) spec-frontmatter-schema.md is byte-identical-by-discipline
  (NOT allowlist-enrolled) → IF it was edited, explicit
  `diff -q .claude/rules/moai/development/spec-frontmatter-schema.md internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md`
  exit 0; if unedited, no mirror action.
- (no mirror edit — lifecycle-sync-gate) `lifecycle-sync-gate.md` has NO template mirror — verify no
  mirror file was created or edited:
  ```bash
  test ! -e internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md && echo "OK no mirror"
  ```

### AC-SSD-006 — Skill 3-file SSOT boundary (REQ-SSD-006)

**Given** skill-authoring=schema, skill-writing-craft=prose, skill-ab-testing=A/B,
**When** the de-dup is applied,
**Then**:
- (absent — distinctive-line delta) `development/skill-writing-craft.md` no longer carries the
  duplicated frontmatter-schema field table. Baseline a distinctive table line (e.g. the
  `| Field | Required | Type | Notes |` header that belongs to the duplicated schema table at ~L214,
  NOT prose) and assert it drops post-edit:
  ```bash
  grep -c '| Field | Required | Type' .claude/rules/moai/development/skill-writing-craft.md   # BASELINE >=1 → AFTER 0
  grep -c "skill-authoring.md" .claude/rules/moai/development/skill-writing-craft.md           # >= 1 (pointer — secondary)
  ```
- (absent — distinctive-line delta) `development/skill-ab-testing.md` frontmatter/progressive-disclosure
  checklist reduced to a pointer. Baseline a distinctive checklist line (e.g. the
  `Frontmatter complete (description, paths, metadata)` checklist item at ~L161) and assert it drops:
  ```bash
  grep -c 'Frontmatter complete (description, paths' .claude/rules/moai/development/skill-ab-testing.md   # BASELINE >=1 → AFTER 0
  grep -c "skill-authoring.md" .claude/rules/moai/development/skill-ab-testing.md   # >= 1 (pointer — secondary)
  ```
- (intact) skill-authoring.md holds the canonical schema table.
- (mirror — per class) all three are byte-identical-by-discipline (NOT allowlist-enrolled) →
  explicit `diff -q <deployed> <mirror>` exit 0 for each EDITED file:
  ```bash
  for f in skill-authoring skill-writing-craft skill-ab-testing; do
    diff -q .claude/rules/moai/development/$f.md internal/template/templates/.claude/rules/moai/development/$f.md
  done
  ```

### AC-SSD-007 — Agent 3-file SSOT boundary (REQ-SSD-007)

**Given** agent-patterns owns the per-domain tool-whitelist table,
**When** the de-dup is applied,
**Then**:
- (intact + single owner) the per-domain tool-whitelist table exists in exactly one file
  (agent-patterns.md); agent-authoring.md and orchestrator-templates.md reference it rather than
  duplicate it:
  ```bash
  grep -lc "| backend | .Read, Write, Edit" .claude/rules/moai/development/agent-patterns.md   # owner
  grep -c "agent-patterns.md" .claude/rules/moai/development/agent-authoring.md                # >= 1 (pointer already exists, preserved)
  ```
- (residual-only) if no residual duplication was found, the milestone is a verified no-op documented
  as such (research showed the boundary is mostly clean) — this AC PASSES on "single owner confirmed".
- (mirror — per class) any EDITED file (agent-patterns / agent-authoring / orchestrator-templates) is
  byte-identical-by-discipline (NOT allowlist-enrolled) → explicit `diff -q <deployed> <mirror>` exit
  0 for each file actually edited; if T8 is a verified no-op, no mirror action.

### AC-SSD-008 — Team files 4 → 2 + 1 pointer + FILE DELETION (REQ-SSD-008)

**Given** team-protocol + team-pattern-cookbook are retained owners,
**When** the consolidation is applied,
**Then** ALL of (3-clause deletion AC):
- (file gone — both trees) `workflow/agent-teams-pattern.md` is deleted in deployed AND template trees:
  ```bash
  test ! -e .claude/rules/moai/workflow/agent-teams-pattern.md && echo "deployed gone"
  test ! -e internal/template/templates/.claude/rules/moai/workflow/agent-teams-pattern.md && echo "template gone"
  ```
- (content survives) the 5+1+1 composition is present in `team-pattern-cookbook.md` as a 6th pattern:
  ```bash
  grep -c "5+1+1\|5 implementer\|implementer.*5" .claude/rules/moai/workflow/team-pattern-cookbook.md   # >= 1
  ```
- (no dangling referrer — WIDENED to internal/ per D8, with worktree noise excluded) no file
  references `agent-teams-pattern` after deletion. The grep MUST scope to the real referrer surfaces
  (rules / skills / Go source / yaml) AND exclude `/worktrees/` (transient agent worktrees carry
  OTHER SPECs' artifacts mentioning the string — verified live: an unscoped grep returns ~3009 noise
  matches) AND exclude this SPEC's own artifacts. The original `.claude/ internal/template/templates/`
  scope MISSED 3 comment referrers in `internal/config/`:
  ```bash
  grep -rn "agent-teams-pattern" .claude/rules/ .claude/skills/ internal/ \
    --include="*.go" --include="*.md" --include="*.yaml" \
    | grep -v "/worktrees/" \
    | grep -v ".moai/specs/SPEC-V3R6-RULES-SSOT-DEDUP-001/" \
    | grep -v "agent-teams-pattern.md:"   # exclude the file being deleted (it self-references in its own body)
  # BASELINE (pre-edit): exactly 5 referrers (the 4 comment refs below + the workflow.yaml comment).
  # AFTER (all rewritten/removed): zero matches.
  ```
  Known referrers to rewrite/remove (verified live at plan-phase — comment-only, harmless to build
  but stale doc refs):
  - `internal/config/workflow_role_profiles_test.go:21` and `:45` (comments: "Teams composition
    documented in agent-teams-pattern.md" / "Modifying this list requires SPEC amendment per
    agent-teams-pattern.md") → repoint to `team-pattern-cookbook.md` 6th pattern.
  - `internal/config/defaults_test.go:527` (comment: "3-element subset per agent-teams-pattern 5+1+1")
    → repoint to `team-pattern-cookbook.md`.
  - `internal/template/rule_template_mirror_test.go:66` (§25-leak-list COMMENT naming the deleted path)
    → update the comment to remove the now-deleted entry (do NOT change test logic).
  - `internal/template/templates/.moai/config/sections/workflow.yaml:28` (comment referencing the rule
    path) → repoint to `team-pattern-cookbook.md`.
  These are ALL comment-only references (no Go logic depends on the file); rewriting the comments is
  safe and required to keep the dangling-grep clean.
- (4th copy → pointer, distinctive-line delta) `workflow/worktree-integration.md` §Team Protocol
  (the "merged from team-protocol.md" copy at L380) reduced to a cross-reference. Baseline a
  distinctive sub-heading of the copied Team Protocol block — verified live, the block carries
  `### Team Discovery`, `### Communication`, `### Task Management`, `### Error Recovery`,
  `### Shutdown Handling`, `### Idle States`, `### Context Isolation` sub-headings (these belong to
  the copied protocol body, NOT the surrounding worktree prose). Assert their count drops to 0:
  ```bash
  # BASELINE (pre-edit, expected >=1 each): the merged-protocol sub-headings
  grep -c '^### Team Discovery\|^### Shutdown Handling\|^### Idle States' .claude/rules/moai/workflow/worktree-integration.md   # BASELINE >=1 → AFTER 0
  grep -c "team-protocol.md" .claude/rules/moai/workflow/worktree-integration.md   # >= 1 (pointer — secondary)
  ```
- (roster reconciled, distinctive-line delta) `team-protocol.md` reconciles the
  `max_teammates default 10` vs 3-5 ceiling contradiction. The current L114 reject-ceiling line stays
  (it is a config safety limit) BUT must be re-framed alongside the 3-5 recommended-starting-size
  statement so the apparent contradiction is gone (design.md §6.3):
  ```bash
  # BASELINE: max_teammates default 10 present (the reject ceiling); AFTER: 3-5 recommended-size also stated
  grep -c "max_teammates" .claude/rules/moai/workflow/team-protocol.md   # reject-ceiling line retained
  grep -iq "3-5\|3 to 5\|recommended starting" .claude/rules/moai/workflow/team-protocol.md   # recommended-size framing added
  ```
- (mirror — per class) team-protocol / team-pattern-cookbook / worktree-integration are
  byte-identical-by-discipline (NOT allowlist-enrolled) → explicit `diff -q <deployed> <mirror>` exit
  0 each; agent-teams-pattern absence verified in BOTH trees (see file-gone clause above):
  ```bash
  for f in team-protocol team-pattern-cookbook worktree-integration; do
    diff -q .claude/rules/moai/workflow/$f.md internal/template/templates/.claude/rules/moai/workflow/$f.md
  done
  ```
- (dependency surfaced) §F CATALOG-SCRUB-001 dependency is noted in the run-phase report.

### AC-SSD-009 — zone-registry bloat reduction (REQ-SSD-009, SHOULD, CLI-gated)

**Given** the `moai constitution validate` drift check requires verbatim-source-substring clauses (B4),
**When** the reduction is applied,
**Then**:
- (no validate break) after the edit, `moai constitution validate` reports no NEW drift:
  ```bash
  moai constitution validate --format json | grep '"status"'   # "ok" (or unchanged baseline)
  ```
- (clause still verbatim) every reduced `clause:` is still a verbatim substring of its source file
  (no 3-5 word non-verbatim label introduced) — proven by validate passing.
- (paths narrowed) the `paths:` trigger is narrowed from `.claude/**,.moai/specs/**,.claude/rules/**`
  to the minimum that preserves intended load behavior:
  ```bash
  grep -n "^paths:" .claude/rules/moai/core/zone-registry.md   # narrowed value
  ```
- (blocker note) IF full label-only reduction was blocked by the CLI dependency, a blocker note
  recommending follow-up SPEC `SPEC-V3R6-CONST-VALIDATE-LABEL-001` is present in the run-phase report.
- (mirror — §25-divergent, NOT byte-parity) zone-registry.md is NOT byte-identical to its mirror
  today (the deployed tree carries a 13-line CONST-V3R6-001 block stripped from the neutral mirror;
  it is NOT in the byte-parity allowlist). Verify via `TestTemplateNoInternalContentLeak` PASS + a
  scoped `diff` confirming ONLY the de-dup `clause:`/`paths:` edit changed and the pre-existing
  13-line delta is unchanged (NOT a `diff -q` exit-0, which would FALSE-FAIL on the legitimate delta):
  ```bash
  go test ./internal/template/... -run TestTemplateNoInternalContentLeak   # PASS
  # confirm the diff still shows ONLY the pre-existing 13-line delta + the new de-dup edit (no new divergence)
  diff .claude/rules/moai/core/zone-registry.md internal/template/templates/.claude/rules/moai/core/zone-registry.md | grep -c '^[<>]'
  ```

### AC-SSD-010 — Template-First mirror discipline (REQ-SSD-010, cross-cutting)

**When** the full change set is committed,
**Then** the Go-test trio is green:
```bash
go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestTemplateNeutralityAudit'   # PASS
```
[HARD] The trio green is NECESSARY but NOT SUFFICIENT for mirror parity: `TestRuleTemplateMirrorDrift`
only inspects the `workflowOptMirroredPaths` allowlist (in scope: `hooks-system.md` ONLY). For the 13
byte-identical-by-discipline files (NOT allowlist-enrolled), the trio passing does NOT prove their
mirror was updated — a missed mirror edit is vacuously green. Therefore this AC ALSO requires the
explicit per-file `diff -q` assertions embedded in AC-SSD-001/002/003/005/006/007/008 to all pass:
```bash
# aggregate explicit diff sweep over every EDITED byte-identical-by-discipline file
# (each edited file's <deployed> and <mirror> must be byte-identical post-edit)
for f in <list of edited byte-identical-by-discipline files>; do
  diff -q .claude/rules/moai/$f internal/template/templates/.claude/rules/moai/$f || echo "MIRROR DRIFT: $f"
done   # expect zero "MIRROR DRIFT" lines
```
and no edit touched a non-existent `embedded.go`:
```bash
git diff --name-only | grep -qx "internal/template/embedded.go" && echo VIOLATION || echo OK   # OK
```

### AC-SSD-011 — Template neutrality preserved (REQ-SSD-011, cross-cutting)

`TestTemplateNoInternalContentLeak` PASS for every §25-sanitized mirror in the change set
(agent-common-protocol, verification-batch-pattern, and any other leak-list file edited).

### AC-SSD-012 — Pointer integrity (REQ-SSD-012, cross-cutting)

Every pointer introduced names both the SSOT owner file AND a section anchor; no pointer references
a deleted file. Spot-verified by reading each introduced `<owner>.md § <anchor>` reference and the
repo-wide dangling-reference grep in AC-SSD-008.

---

## §D.1 Severity

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-SSD-001..008 | MUST | Core de-dup deliverables |
| AC-SSD-009 | SHOULD | CLI-gated; partial acceptable with blocker note |
| AC-SSD-010..012 | MUST | Cross-cutting integrity (mirror/neutrality/pointer) |

## §D.2 Definition of Done

- All MUST ACs PASS with grep/test evidence in progress.md §E.2/§E.3.
- AC-SSD-009 PASS or PASS-WITH-BLOCKER-NOTE (follow-up SPEC recommended).
- Go-test trio green; `moai constitution validate` no new drift.
- No CLAUDE.md / Go-source edit (except the optional Go-test comment update for the deleted file path).
- §F dependency surfaced to orchestrator.

## §D.3 Closure Gates

- status `draft → in-progress` at M1 commit (manager-develop).
- status `in-progress → implemented` at sync (manager-docs).
- 4-phase close (Mx) after `moai spec audit` shows era V3R6, drift 0.
