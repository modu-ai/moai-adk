# Implementation Plan — SPEC-CC2178-TEAM-API-ALIGN-001

## §A. Context

Docs/doctrine alignment SPEC tracking the Claude Code 2.1.178 Agent Teams API change. Three axes (T1-1 team-API doctrine, T1-3 `/config` command docs, D1 `attribution.sessionUrl`). All edits are documentation/template content; no team-subsystem behavior change. Tier M because the work spans a 16-file template-mirrored doctrine corpus + a 4-locale docs-site surface but is mechanically bounded (text replacement against confirmed upstream facts, no code-logic change).

Authoritative blast radius, measured live at authoring time (not assumed):

| Surface | Files | Notes |
|---------|-------|-------|
| Doctrine `.claude/**` (TeamCreate/TeamDelete) | 15 | rules ×7, skills ×8 |
| `CLAUDE.md` §15 + §4 | 1 | §15 Team APIs enumeration + cleanup line; §4 consistency-only |
| Template mirrors | 16 | byte-identical to the 16 above (symmetric) |
| docs-site team-API | 9 | en/ko/ja/zh × {what-is-moai-adk, introduction} = 8 + ko/advanced/hooks-guide = 1 |
| README locales | 4 | README.{md,ko,ja,zh}.md Mermaid node + bullet |
| `settings.json.tmpl` (D1) | 1 | `attribution` block, add `sessionUrl` |
| `settings-management.md` (D1) | 1 | add attribution doctrine line |
| Go references (assess) | 4 | all comment/text at authoring time; none live-behavior |

## §B. Known Issues / Findings from authoring-time investigation

1. **Go references are NOT live-behavior** (REQ-GO-001 pre-assessment): all four named Go files reference `TeamCreate`/`TeamDelete` only in comments, `t.Log` strings, `fmt.Printf` log text, or a mock type name (`mockTeamCreator`). None invoke a removed Claude Code tool. The mock models the *gate-before-spawn ordering* (a MoAI-internal Plan Audit Gate contract), which survives the upstream change conceptually. Run-phase MUST re-confirm this and record the finding; the default is no Go-code change.

2. **`skills_audit_test.go` does NOT depend on the removed tool names**: its `requiredPatterns` for `team/run.md` are `Phase 0.5: Plan Audit Gate`, `plan-auditor`, `--skip-audit`, `INCONCLUSIVE`, `.moai/reports/plan-audit/` — NOT `TeamCreate`/`TeamDelete`. The line-50 `TeamCreate` reference is a *code comment*. Therefore doctrine edits to `team/run.md` will NOT break this test as long as the gate markers are preserved. This must be re-verified after the `team/run.md` edit.

3. **Locale divergence in hooks-guide**: `docs-site/content/ko/advanced/hooks-guide.md` (L132-139) carries an obsolete 3-step "TeamDelete cleanup" instruction (3 matches); the en/ja/zh hooks-guide siblings have 0 matches. Run-phase MUST reconcile per REQ-LOC-002 — either remove the obsolete ko block (cleanup is now automatic) or confirm the sibling locales legitimately never had it. The likely correct action is to remove/rewrite the ko block to the automatic-cleanup model and verify 0 across all four locales.

4. **`attribution.sessionUrl` schema unconfirmed offline**: the template block has `commit` + `pr` string keys. The exact value type of `sessionUrl` (boolean to omit, vs string) is NOT confirmable from the offline corpus. Run-phase MUST confirm against the official CC settings reference before writing the key. See §G open question.

5. **`/config` grep noise**: 86 files match bare `/config`; only 1 (`dynamic-workflows.md`) uses the backtick-quoted command form. The precise grep gate (`` `/config` `` or command-form) is mandatory (REQ-CFG-003).

## §C. Pre-flight checks (run-phase entry)

1. `git fetch origin main` + divergence count (multi-session race mitigation per agent-common-protocol §Pre-Spawn Sync Check). This repo has known parallel-session activity (RC2-README SPEC in-flight per git status).
2. Re-run the live blast-radius greps (§A) to confirm counts have not drifted since authoring.
3. **Capture the `.moai/config/`-path baseline count for AC-CFG-002** (D2). AC-CFG-002 asserts a post-edit invariant: the number of `.moai/config/` filesystem-path references in the doctrine corpus MUST be unchanged by the M4 `/config` edits. That invariant is only checkable against a captured pre-edit baseline. Record this number in `progress.md §E.2` at run-phase entry, before any M4 edit:

   ```bash
   # Baseline: total count of '.moai/config/' path references in the doctrine corpus.
   grep -rc '\.moai/config/' .claude/ CLAUDE.md \
     | grep -v '\.claude/worktrees/' | grep -v '/backups/' \
     | awk -F: '{s+=$2} END{print s}'
   # Record the printed integer as CFG_CONFIG_PATH_BASELINE in progress.md §E.2.
   # AC-CFG-002 PASS condition: the same command post-M4 yields the identical integer.
   ```

4. Confirm `attribution.sessionUrl` schema against the maintainer's own CC instance before M5 (resolves §G OQ-1 — see §G for the resolution status: type NOT confirmable from an authoritative machine-readable source yet; D1/attribution step is conditional on confirming the type in-instance).
5. Confirm `teammateMode` `auto→in-process` and idle-row-hide facts against `.moai/research/cc-update-2.1.164-to-2.1.183.md`.

## §D. Constraints

- **Template-mirror parity** (REQ-MIR-001): every `.claude/**` / `CLAUDE.md` edit mirrored byte-identically; `make build` after (REQ-MIR-002).
- **Template neutrality** (REQ-NEU-001): no internal-content leak in template files; `TestTemplateNeutralityAudit` must pass.
- **4-locale sync** (REQ-LOC-001): docs-site edits span en/ko/ja/zh equally.
- **No team-subsystem behavior change**: doctrine/template/docs only.
- **Quality gate** (REQ-QG-001): `go test ./...`, `golangci-lint`, `TestTemplateNeutralityAudit` green post-edit.

## §E. Self-Verification (plan-phase)

- [x] Blast radius measured live (16==16 symmetric; 9 docs-site; 4 README; 4 Go assessed).
- [x] Go references pre-assessed as comment/text (not live-behavior).
- [x] `skills_audit_test.go` requiredPatterns confirmed independent of removed-tool names.
- [x] hooks-guide locale divergence identified (ko 3, en/ja/zh 0).
- [x] `/config` precise-grep gate validated (86 noisy → 1 genuine).
- [x] SPEC ID self-check PASS; frontmatter 12-field schema valid; Out of Scope present.

## §F. Milestones (priority-ordered, no time estimates)

### M1 — Axis T1-1 doctrine corpus (15 `.claude/**` + CLAUDE.md) [High]
Rewrite TeamCreate/TeamDelete live-action instructions to the implicit-team model across the 15 doctrine files + CLAUDE.md §15. Apply REQ-TAA-001..005, 007. Reflect teammateMode/idle-row facts only where `teammateMode` is already discussed (REQ-TAA-006: `worktree-integration.md`, `moai-workflow-worktree/SKILL.md`). Verify CLAUDE.md §4 consistency (REQ-TAA-008, no scope expansion).

### M2 — Axis T1-1 template mirrors (16 files) + build [High]
Apply the byte-identical M1 edits to the 16 `internal/template/templates/**` mirrors. Run `make build` to regenerate `embedded.go` (REQ-MIR-001, REQ-MIR-002). Re-run `skills_audit_test.go` to confirm gate markers preserved.

### M3 — Axis T1-1 docs-site (9 files, 4-locale) + README (4 files) [High]
Reconcile the docs-site team-API references: rewrite Mermaid `TeamCreate → SendMessage` nodes and the `TeamCreate, SendMessage, TaskList` bullets to the implicit-team surface across en/ko/ja/zh `what-is-moai-adk.md` + `introduction.md` (OQ-2 default: rewrite to `Agent(name=…) → SendMessage` unless maintainer overrides at Implementation Kickoff Approval). Reconcile the ko `advanced/hooks-guide.md` obsolete 3-step cleanup block per REQ-LOC-002 (OQ-3 default: rewrite to preserve the tmux-pane warning + drop the `TeamDelete` step, unless maintainer overrides to deletion). Rewrite the 4 repo-root README locale Mermaid nodes + bullets (AC-LOC-003 — README is NOT template-mirrored). Verify per-locale parity (docs-site via AC-LOC-001/002; README via AC-LOC-003).

### M4 — Axis T1-3 `/config` command docs [Med]
Precise-grep the genuine `/config` command references (REQ-CFG-003). Add the `/config key=value`, `/config --help`, and Enter/Space/Esc toggle-behavior documentation (REQ-CFG-001, 002). Mirror any `.claude/**` edits to the template (REQ-MIR-001).

### M5 — Axis D1 `attribution.sessionUrl` [Med]
**Conditional on OQ-1 in-instance type confirmation.** Confirm the `sessionUrl` value type in the maintainer's own CC instance (§G OQ-1 resolution status: NOT confirmable from an authoritative machine-readable source; boolean most plausible per release-note but unverified). ONLY after in-instance confirmation, add the key with the confirmed literal type to the `settings.json.tmpl` `attribution` block (REQ-ATT-001). If the type cannot be confirmed in-instance during run-phase, M5 returns a blocker report rather than pinning an unverified literal (AP-3). The `settings-management.md` doctrine line (REQ-ATT-002) MAY be added per the release-note omit-session-link semantics WITH a "verify type before pinning a default" caveat — and mirrored. `make build`.

### M6 — Quality gate + Go assessment record [High]
Run the read-only verification batch: `go test ./...`, `golangci-lint run`, `go test ./internal/template/... -run TestTemplateNeutralityAudit`. Record the REQ-GO-001 assessment finding (default: no Go change). Confirm zero residual live-action `TeamCreate`/`TeamDelete` instructions across the full corpus.

## §G. Anti-Patterns / Open Questions

### Anti-patterns to avoid
- **AP-1**: Editing transient `.claude/worktrees/agent-*` copies (out of scope §J). Edit only the live tree + template mirror.
- **AP-2**: Blanket find/replace of bare `/config` — would corrupt 85 `.moai/config/` path references. Use the precise command-form grep only.
- **AP-3**: Pinning a literal `sessionUrl` value into `settings.json.tmpl` without confirming the type in the maintainer's own CC instance (schemastore is stale per OQ-1; the authoritative machine-readable source does not yet carry the key). Asserting an unverified type violates the verification-claim-integrity invariant — return a blocker instead.
- **AP-4**: Removing the `TeamCreate` *comment* in `audit_gate.go`/`team_run_audit_gate_test.go` as if it were a behavior change — the gate-before-spawn ordering is a MoAI-internal contract that survives; do not break the mock's intent.
- **AP-5**: Editing one locale and missing the 3 siblings (4-locale parity drift — the canonical recurring docs failure mode).

### Open questions for the maintainer (OQ)

- **OQ-1** [run-phase blocker — RESOLUTION STATUS RECORDED, type NOT yet confirmed]: What is the exact value type and schema of `attribution.sessionUrl`?

  **Verification finding (recorded, no schema asserted)**: Two authoritative sources were checked. (1) The schemastore JSON schema (`https://www.schemastore.org/claude-code-settings.json`) lists only `commit` / `pr` under `attribution` — the schema is stale (predates CC 2.1.183), so `sessionUrl` is absent there; its absence is NOT evidence the key does not exist, only that the machine-readable schema has not caught up. (2) The official `/en/settings` page does not expose the sub-key detail. The CC 2.1.183 release-note semantics ("Added `attribution.sessionUrl` setting to omit the claude.ai session link from commits and PRs in web and Remote Control sessions") most plausibly indicate a **boolean**, but the type is NOT confirmable from an authoritative machine-readable source yet. Per the verification-claim-integrity invariant, the SPEC does NOT assert a schema it cannot verify.

  **Run-phase contract (conditional)**: The D1 / `attribution` step (M5) is EXPLICITLY conditional on confirming the value type in the maintainer's own CC instance before writing any literal value into `settings.json.tmpl`. Acceptable in-instance confirmation methods: the local `claude` settings schema, `/config` autocomplete on the `attribution.sessionUrl` key, or an authoritative updated schemastore/docs entry. Until the type is confirmed in-instance, M5 MUST NOT pin a literal `sessionUrl` value into the template (writing an unconfirmed literal would violate REQ-NEU-001-adjacent correctness + the verification-claim-integrity invariant — AP-3). The `settings-management.md` doctrine line (REQ-ATT-002) MAY document the key per the release-note semantics (omit-session-link behavior) WITH an explicit "verify type before pinning a default" caveat — documenting the documented behavior is permitted; pinning an unverified template literal is not.

- **OQ-2** [scope decision — recommended default encoded, maintainer may override at Implementation Kickoff Approval]: For the README/docs-site Mermaid nodes that read `TeamCreate → SendMessage`, should the node be (a) rewritten to `Agent(name=…) → SendMessage` (implicit-team spawn), or (b) removed entirely as a stale implementation detail in a user-facing diagram?

  **Recommended default (run-phase applies UNLESS overridden)**: (a) — rewrite to `Agent(name=…) → SendMessage`; keep the coordination concept, name the current mechanism. Run-phase (M3) SHALL apply default (a) unless the maintainer overrides to (b) at the Implementation Kickoff Approval gate. AC-LOC-003 Part B is written to accommodate either outcome (Part A — zero stale refs — gates the AC under both; Part B asserts the implicit-team form only under default (a)).

- **OQ-3** [scope decision — recommended default encoded, maintainer may override at Implementation Kickoff Approval]: Should the ko `hooks-guide` 3-step TeamDelete block be deleted (cleanup is automatic) or rewritten to describe the automatic-cleanup behavior + the remaining tmux-pane caveat?

  **Recommended default (run-phase applies UNLESS overridden)**: rewrite (not delete) — preserve the genuine tmux-pane-lifecycle warning (the `shutdown_response` does not kill the tmux pane process) while dropping the obsolete `TeamDelete` call step (cleanup is automatic on session exit). Run-phase (M3) SHALL apply this rewrite default unless the maintainer overrides to full deletion at the Implementation Kickoff Approval gate. AC-LOC-002 (zero obsolete `TeamDelete`/`TeamCreate` matches across all 4 hooks-guide locales) gates the AC under either outcome.

> **Maintainer override channel**: OQ-2 and OQ-3 are recorded as maintainer-pending. The orchestrator surfaces them at the Implementation Kickoff Approval (plan→run human gate) via `AskUserQuestion`. If the maintainer makes no override selection, run-phase proceeds with the recommended defaults above. This SPEC does NOT assert maintainer consent for either default — the defaults are encoded as the no-override fallback, with the human gate preserved as the override channel.

## §H. Cross-References

- `.moai/research/cc-update-2.1.164-to-2.1.183.md` — research basis.
- `.claude/rules/moai/workflow/team-protocol.md`, `team-pattern-cookbook.md`, `orchestration-mode-selection.md` — primary team doctrine.
- CLAUDE.local.md §2/§24 (template-mirror), §15/§25 (neutrality), §17 (4-locale).
- Sibling SPECs: `SPEC-CC2178-DOCS-ALIGN-001`, `SPEC-CC2178-MODEL-POLICY-REPAIR-001`.
