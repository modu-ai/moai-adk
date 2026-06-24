# Acceptance Criteria — SPEC-CC2178-TEAM-API-ALIGN-001

All grep commands run from the repo root `/Users/goos/MoAI/moai-adk-go`. Worktree (`.claude/worktrees/`) and backup (`.moai/backups/`) paths are excluded from every corpus grep because they are out of scope (spec.md §J).

## §D. AC Matrix

| AC ID | REQ | Severity | Axis |
|-------|-----|----------|------|
| AC-TAA-001 | REQ-TAA-001/002 | MUST | T1-1 |
| AC-TAA-002 | REQ-TAA-003 | MUST | T1-1 |
| AC-TAA-003 | REQ-TAA-004 | MUST | T1-1 |
| AC-TAA-004 | REQ-TAA-005 | MUST | T1-1 |
| AC-TAA-005 | REQ-TAA-006 | SHOULD | T1-1 |
| AC-TAA-006 | REQ-TAA-007 | MUST | T1-1 |
| AC-TAA-007 | REQ-TAA-008 | SHOULD | T1-1 |
| AC-CFG-001 | REQ-CFG-001/002 | SHOULD | T1-3 |
| AC-CFG-002 | REQ-CFG-003 | MUST | T1-3 |
| AC-ATT-001 | REQ-ATT-001 | MUST | D1 |
| AC-ATT-002 | REQ-ATT-002 | SHOULD | D1 |
| AC-MIR-001 | REQ-MIR-001/002 | MUST | cross |
| AC-NEU-001 | REQ-NEU-001 | MUST | cross |
| AC-LOC-001 | REQ-LOC-001 | MUST | cross |
| AC-LOC-002 | REQ-LOC-002 | MUST | cross |
| AC-LOC-003 | REQ-LOC-003 | MUST | T1-1 (README) |
| AC-QG-001 | REQ-QG-001 | MUST | cross |
| AC-GO-001 | REQ-GO-001 | MUST | cross |

---

## Given-When-Then Scenarios

### AC-TAA-001 — Zero residual live-action TeamCreate/TeamDelete in doctrine corpus (MUST)

- **Given** the 16-file doctrine corpus (15 `.claude/**` + CLAUDE.md) has been aligned.
- **When** the corpus is grepped for live-action removed-tool instructions.
- **Then** no instruction tells an actor to call `TeamCreate`/`TeamDelete` as a live action; any surviving mention is a migration note ("removed in v2.1.178", "formerly `TeamCreate`") only.

```bash
# Live-action residue check — expect each surviving line to be a migration/historical note, not an instruction
grep -rn 'TeamCreate\|TeamDelete' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Manual review: every remaining line must be a "removed/formerly" migration note, NOT "use TeamCreate" / "call TeamDelete".
```

### AC-TAA-002 — Implicit-team spawn mechanism documented (MUST)

- **Given** the doctrine describes spawning a teammate.
- **When** the spawn mechanism is read.
- **Then** the Agent tool's `name` parameter is presented as the canonical spawn mechanism, and the doctrine states teams form implicitly on first spawn (no setup step).

```bash
# At least one doctrine surface documents implicit-team spawn via name parameter
grep -rln 'implicit team\|name parameter\|spawn teammates directly\|no setup step' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: ≥1 match.
```

### AC-TAA-003 — team_name accepted-but-ignored + hook-payload deprecation documented (MUST)

- **Given** the doctrine mentions the `team_name` parameter or hook payload.
- **When** the `team_name` semantics are read.
- **Then** the doctrine states the Agent-tool `team_name` is accepted but ignored, and the `team_name` field in `TaskCreated`/`TaskCompleted`/`TeammateIdle` hook payloads is deprecated (session-derived name).

```bash
grep -rn 'accepted but ignored\|team_name.*deprecated\|deprecated.*team_name\|session-derived' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: ≥1 match covering both the Agent-tool param and the hook-payload field.
```

### AC-TAA-004 — Automatic-cleanup, no manual TeamDelete teardown (MUST)

- **Given** doctrine previously instructed "Call TeamDelete only after all teammates have shut down".
- **When** the corpus is grepped for that manual-teardown instruction.
- **Then** the instruction is absent, replaced by the automatic-cleanup-on-session-exit statement.

```bash
# Manual teardown instruction must be gone
grep -rn 'Call TeamDelete only after\|TeamDelete only after all teammates' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: 0 matches.
# Automatic-cleanup statement present:
grep -rln 'automatic.*session exit\|cleanup.*automatic\|automatically.*session' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: ≥1 match.
```

### AC-TAA-005 — teammateMode auto→in-process reflected only where already discussed (SHOULD)

- **Given** the two files that already discuss `teammateMode` (`worktree-integration.md`, `moai-workflow-worktree/SKILL.md`).
- **When** they are read.
- **Then** they reflect the `auto → in-process` default (v2.1.179) and idle-row-hide (v2.1.181); no NEW file gains `teammateMode` discussion.

```bash
# Files discussing teammateMode must be the same set (no new ones)
grep -rln 'teammateMode' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: the same set known at authoring (worktree-integration.md + moai-workflow-worktree/SKILL.md), plus any pre-existing.
# in-process default reflected:
grep -rn 'in-process' .claude/rules/moai/workflow/worktree-integration.md .claude/skills/moai-workflow-worktree/SKILL.md
# Expect: ≥1 match.
```

### AC-TAA-006 — CLAUDE.md §15 Team APIs rewritten + mirror parity (MUST)

- **Given** CLAUDE.md §15 listed `TeamCreate, SendMessage, TaskCreate/Update/List/Get, TeamDelete` and a "Call TeamDelete..." line.
- **When** §15 and its template mirror are read.
- **Then** both carry the implicit-team API surface (no `TeamCreate`, no `TeamDelete` as live APIs), and the two files are byte-identical for the edited region.

```bash
grep -n 'TeamCreate\|TeamDelete' CLAUDE.md internal/template/templates/CLAUDE.md
# Expect: 0 live-API listings (migration notes only, if any).
diff <(grep -n 'Team\|team_name\|implicit' CLAUDE.md) \
     <(grep -n 'Team\|team_name\|implicit' internal/template/templates/CLAUDE.md)
# Expect: identical (parity).
```

### AC-TAA-007 — CLAUDE.md §4 nested-subagent note consistency, no scope expansion (SHOULD)

- **Given** the §4 "Watch (Claude Code 2.1.172)" note.
- **When** it is read after alignment.
- **Then** it remains factually consistent with the implicit-team model and was not expanded beyond team-API facts.

```bash
grep -n 'Watch (Claude Code 2.1.172)' CLAUDE.md internal/template/templates/CLAUDE.md
# Manual review: note unchanged or only minimally consistency-edited; no new nested-delegation adoption text.
```

### AC-CFG-001 — /config command behavior documented (SHOULD)

- **Given** the genuine `/config` command reference(s).
- **When** the documentation is read.
- **Then** it describes ALL THREE distinct facts, each as actual added doctrine content (not an incidental match): (1) the `/config key=value` direct-set form, (2) the `/config --help` shorthand-key listing, (3) the toggle-key behavior change where Enter AND Space both change the selected setting and Esc saves-and-closes.

This AC is tightened to be non-vacuous: instead of a single loose alternation (`Esc.*save | Enter.*Space`, which can pass on incidental prose), it requires each of the three facts to be present as a dedicated, separately-counted assertion anchored to the literal token(s) the run-phase edit introduces.

```bash
# Fact 1 — direct-set form: literal '/config key=value' or '/config <key>=<value>'.
grep -rEn '/config[[:space:]]+[A-Za-z._]+=' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: ≥1 match.

# Fact 2 — help listing: literal '/config --help'.
grep -rn '/config --help' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: ≥1 match.

# Fact 3 — toggle-behavior sentence: the added sentence MUST name Enter, Space, AND Esc together
# (a single sentence/line mentioning all three keys — incidental matches that mention only one key do not count).
grep -rEn 'Enter.*Space.*Esc|Esc.*(Enter|Space).*(Space|Enter)' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/'
# Expect: ≥1 match — the toggle-behavior sentence naming all three keys (Enter+Space change, Esc saves-and-closes).
# Anchoring to all three keys in one line prevents a vacuous pass on prose that merely mentions "Esc" or "Enter" elsewhere.
```

### AC-CFG-002 — /config edits scoped to genuine command refs only (MUST)

- **Given** bare `/config` is noisy (matches `.moai/config/` paths).
- **When** the diff of this SPEC's run-phase is inspected.
- **Then** no `.moai/config/` filesystem-path reference was altered; only genuine `/config` command references were edited.

```bash
# Count of .moai/config/ path references must be unchanged by this SPEC's edits.
grep -rc '\.moai/config/' .claude/ CLAUDE.md \
  | grep -v '\.claude/worktrees/' | grep -v '/backups/' | awk -F: '{s+=$2} END{print s}'
# Compare against the pre-edit baseline captured in §C pre-flight. Expect: identical count.
```

### AC-ATT-001 — attribution.sessionUrl present in template (MUST)

- **Given** the `settings.json.tmpl` `attribution` block had `commit` + `pr`.
- **When** the block is read after the edit.
- **Then** it carries a `sessionUrl` sub-key with the value type confirmed against the official CC settings reference.

```bash
grep -n 'sessionUrl' internal/template/templates/.claude/settings.json.tmpl
# Expect: ≥1 match inside the attribution block.
# Validate JSON template still renders (build/test):
go test ./internal/template/... 2>&1 | tail -5
```

### AC-ATT-002 — attribution documented in settings-management.md (SHOULD)

- **Given** `settings-management.md` had zero attribution coverage.
- **When** it is read after the edit.
- **Then** it documents the `attribution` block including `sessionUrl` in ≥1 line, with template-mirror parity.

```bash
grep -in 'attribution\|sessionUrl' .claude/rules/moai/core/settings-management.md \
  internal/template/templates/.claude/rules/moai/core/settings-management.md
# Expect: ≥1 match in each, parity.
```

### AC-MIR-001 — Template-mirror parity + embedded.go regenerated (MUST)

- **Given** every `.claude/**` / `CLAUDE.md` edit must have an identical template mirror.
- **When** parity is checked and `make build` is run.
- **Then** local and mirror are byte-identical for edited files, and `embedded.go` is regenerated.

```bash
# For each edited doctrine file, local vs template mirror must be identical:
for f in $(git diff --name-only origin/main -- '.claude/**' 'CLAUDE.md' | grep -v '\.claude/worktrees/'); do
  mirror="internal/template/templates/$f"
  [ -f "$mirror" ] && diff "$f" "$mirror" >/dev/null && echo "PARITY OK: $f" || echo "PARITY FAIL: $f"
done
# Expect: every line PARITY OK.
make build  # regenerates internal/template/embedded.go
```

### AC-NEU-001 — Template neutrality preserved (MUST)

- **Given** template-side edits must be language-neutral and leak-free.
- **When** the neutrality audit runs.
- **Then** `TestTemplateNeutralityAudit` passes (no internal SPEC IDs, REQ tokens, dates, SHAs introduced).

```bash
go test ./internal/template/... -run TestTemplateNeutralityAudit -count=1
# Expect: PASS.
```

### AC-LOC-001 — docs-site 4-locale parity (MUST)

- **Given** docs-site edits span en/ko/ja/zh.
- **When** each locale's team-API surface is grepped.
- **Then** all four locales carry the equivalent implicit-team wording (no locale left stale).

```bash
for loc in en ko ja zh; do
  printf "%-4s what-is: " "$loc"; grep -c 'TeamCreate\|TeamDelete' "docs-site/content/$loc/core-concepts/what-is-moai-adk.md"
  printf "%-4s intro:   " "$loc"; grep -c 'TeamCreate\|TeamDelete' "docs-site/content/$loc/getting-started/introduction.md"
done
# Expect: every count 0 (stale live-tool refs removed/rewritten across all 4 locales equally).
```

### AC-LOC-002 — hooks-guide locale divergence reconciled (MUST)

- **Given** ko `advanced/hooks-guide.md` had 3 obsolete TeamDelete-cleanup matches; en/ja/zh had 0.
- **When** all four hooks-guide locales are grepped after reconciliation.
- **Then** all four are consistent: 0 obsolete `TeamDelete` manual-cleanup-step matches across every locale.

```bash
for loc in en ko ja zh; do
  printf "%-4s hooks-guide: " "$loc"
  grep -c 'TeamDelete\|TeamCreate' "docs-site/content/$loc/advanced/hooks-guide.md" 2>/dev/null || echo "(absent)"
done
# Expect: every count 0 (ko reconciled to automatic-cleanup model; siblings already 0).
```

### AC-LOC-003 — README 4-locale stale team-API references removed (MUST)

- **Given** the 4 repo-root README locale files (`README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`) each carried a stale `TeamCreate → SendMessage` Mermaid node and a `TeamCreate, SendMessage, TaskList` bullet (live-measured at authoring: README.md L398/L421, README.ko.md L444/L467, README.ja.md L447/L470, README.zh.md L447/L470). README files are repo-root project-owned — they are NOT template-mirrored (so AC-MIR-001 does not cover them) and NOT under docs-site (so AC-LOC-001/002 do not cover them). Without this AC, the README edits planned in plan.md §F (M3) / design.md §C.4 are an orphan edit class with no acceptance coverage.
- **When** all four README locale files are grepped after the M3 edit.
- **Then** every README locale file contains zero stale `TeamCreate`/`TeamDelete` live-API references; and if the Mermaid nodes are rewritten per OQ-2 (recommended default `Agent(name=…) → SendMessage`, maintainer may override at the Implementation Kickoff Approval gate), each rewritten node shows the implicit-team form rather than the removed-tool form.

```bash
# Part A — zero stale live-API references across all 4 README locales:
for f in README.md README.ko.md README.ja.md README.zh.md; do
  printf "%-12s " "$f:"; grep -c 'TeamCreate\|TeamDelete' "$f"
done
# Expect: every count 0 (any surviving mention, if intentionally kept, must be a "removed in v2.1.178" migration note — verify by manual review of any nonzero line).

# Part B — if OQ-2 default applied (Mermaid node rewritten to implicit-team form), the new form is present:
#   (conditional: only assert Part B when the maintainer did NOT override OQ-2 to "remove node entirely")
for f in README.md README.ko.md README.ja.md README.zh.md; do
  printf "%-12s " "$f:"; grep -c 'Agent(name' "$f"
done
# Expect (under OQ-2 default): ≥1 match per locale. Under maintainer "remove node" override: 0 expected and Part A alone gates this AC.
```

### AC-QG-001 — Quality gate green (MUST)

- **Given** the run-phase edits are complete.
- **When** the verification batch runs.
- **Then** `go test ./...`, `golangci-lint run`, and `TestTemplateNeutralityAudit` all pass.

```bash
go test ./... 2>&1 | tail -5         # Expect: no FAIL
golangci-lint run --timeout=2m       # Expect: 0 issues
go test ./internal/template/... -run TestTemplateNeutralityAudit -count=1  # Expect: PASS
# Specifically re-run the skills audit (must still pass — gate markers preserved):
go test ./internal/template/... -run TestSkills 2>&1 | tail -3
```

### AC-GO-001 — Go reference assessment recorded, no spurious code change (MUST)

- **Given** the 4 named Go files reference `TeamCreate`/`TeamDelete` in comments/text/mock-names only.
- **When** the run-phase assessment is recorded.
- **Then** the finding states each reference is non-live-behavior (the default — no code change), OR documents the specific live reference that required a change with justification.

```bash
grep -n 'TeamCreate\|TeamDelete' \
  internal/cli/team_spawn_test.go \
  internal/cli/team_run_audit_gate_test.go \
  internal/runtime/audit_gate.go \
  internal/template/skills_audit_test.go
# Manual review per line: classify as {comment | t.Log | fmt.Printf | mock-name | requiredPattern}.
# Authoring-time finding: all are comment/log/mock-name; skills_audit_test requiredPatterns do NOT include the tool names.
# Run-phase MUST re-confirm and record in progress.md §E.2.
```

---

## §D.1 Severity / Closure Gates

- **MUST-FIX (block close)**: AC-TAA-001, -002, -003, -004, -006; AC-CFG-002; AC-ATT-001; AC-MIR-001; AC-NEU-001; AC-LOC-001; AC-LOC-002; AC-LOC-003; AC-QG-001; AC-GO-001.
- **SHOULD (PASS-WITH-DEBT eligible if justified)**: AC-TAA-005, -007; AC-CFG-001; AC-ATT-002.

## §D.2 Definition of Done

- [ ] All MUST-FIX ACs PASS with reproduced grep/command evidence (verification-claim-integrity: each PASS attributed to an actually-run command + observed output).
- [ ] Template-mirror parity verified for all edited files (16-file symmetric corpus).
- [ ] `make build` run; `embedded.go` regenerated and committed.
- [ ] 4-locale docs-site parity verified per-locale (no glob-total masking).
- [ ] README 4-locale stale team-API references removed (AC-LOC-003 — README is NOT covered by AC-MIR-001 or AC-LOC-001/002).
- [ ] Go-reference assessment recorded in progress.md §E.2 (REQ-GO-001).
- [ ] Open questions OQ-1 (sessionUrl schema), OQ-2 (Mermaid node fate), OQ-3 (hooks-guide ko block) resolved at run-phase entry per plan.md §G.
- [ ] Quality gate green.
