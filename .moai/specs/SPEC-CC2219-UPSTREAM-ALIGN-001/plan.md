# plan.md — SPEC-CC2219-UPSTREAM-ALIGN-001

## §A Context

Umbrella alignment of MoAI-ADK doctrine/agents/templates/Go code with Claude Code 2.1.208..2.1.219. Evidence: `.moai/research/cc-update-2.1.207-to-2.1.219.md` (§4 detail, §9 file inventory). GD-3 excluded (PR #1146 hotfix). Fresh 2026-07-25 probe resolved the GD-1 three-way conflict: nesting is default-ON on 2.1.219 (single depth-1 trial; ceiling not probed).

Milestones are ordered by **decision reversibility** — the safety-doctrine decision (M1) and the Go data-model change (M2) lead; mechanical doc sweeps trail.

## §B Known Issues / Risks

1. **Upstream reversal risk**: 2.1.217→2.1.219 flipped the nesting default twice in 2 days. Mitigation: doctrine text cites the probe date + binary version, so a future flip is a dated correction, not a contradiction.
2. **GD-4 @MX:ANCHOR cascade**: `ModelAliasTable` fan_in ≥ 3 (`launcher.go` `expandModelString`, `profile_setup.go` `normalizeModel`, `settings/schema.go` `modelOptions`). Line numbers in the report drift — re-anchor by content token before editing.
3. **Parent-mode precedence** (GD-2): the maintainer's own `defaultMode: bypassPermissions/acceptEdits` (CLAUDE.local.md §22.1) means even `permissionMode: plan` frontmatter is not a fix — only tool restriction is. Do not "migrate" `mode:` to `permissionMode:`; migrate to tool restriction.
4. **Mirror-parity classes vary**: some surfaces are byte-identical mirrors, others sanitized pairs — re-measure per file at run time (do not assume from memory).
5. **Line numbers in report §9 are drift-prone** — treat as hints; grep the quoted stale text as the anchor.

## §C Pre-flight

1. Re-run per-cluster greps from report §4 against current HEAD (report measured at `c2fd0bf8c`; other sessions are active on this checkout).
2. Confirm PR #1146 landed the GD-3 matcher on main: `git show origin/main:.claude/settings.json | grep fork` (merged as commit 714270085). The current checkout may predate the merge and still lack `fork` locally — rebase onto origin/main before run-phase, then re-verify the local file.
3. `grep -rn '@MX:ANCHOR' internal/template/model_policy.go` — confirm anchor + enumerate fan-in callers freshly.
4. M1 decision gate: surface REQ-GD1-005 option (a) retire `Agent` from sync-auditor vs (b) retain pilot with tool-restricted-children rationale, via orchestrator AskUserQuestion, BEFORE editing sync-auditor.md.

## §D Constraints

- Template-First: every `.claude/` edit pairs with `internal/template/templates/` mirror + `make build` (REQ-X-001); neutrality guards must stay green (REQ-X-002).
- No edits to `.moai/specs/`/`.moai/reports/` historical artifacts (REQ-GD2-003 exemption).
- Concurrency safeguard prose (one write-capable agent) preserved verbatim (REQ-X-003).
- Probe caveats must ride every ceiling/propagation assertion (REQ-GD1-004).

## §E Self-Verification

Per milestone: (E1) AC matrix rows for the milestone's cluster PASS with verbatim grep/build output; (E2) `go build ./...` + `go test ./...` green after any Go edit; (E4) template-mirror parity check; (E5) neutrality guard test green; stale-token absence greps (see acceptance.md) return 0 on live+template surfaces.

## §F Milestones (priority-ordered, no time estimates)

### M1 — Child A: nesting + permission-mode doctrine (GD-1 + GD-2) — HIGHEST-CHANGE-LIKELIHOOD DECISIONS
- Resolve the REQ-GD1-005 sync-auditor pilot decision (gate in §C.4) first.
- Rewrite the 7 GD-1 surfaces + 6 GD-2 surfaces (+ mirrors): default-ON depth 3, `=1` disables, double-guarantee removed, `mode: "plan"` → tool-restriction grounding, probe caveats + provenance citations.
- Add the SPEC-SUBAGENT-NESTING-DOCTRINE-001 supersession cross-reference note (REQ-GD1-006).

### M2 — Child C: Opus 5 migration, Go-code portion (GD-4) — **@MX:ANCHOR flagged**
- `model_policy.go`: `opus` alias → `claude-opus-5`; deprecated row for `claude-opus-4-8`; `opus[1m]` re-check. **This edits an @MX:ANCHORed symbol (`ModelAliasTable`, fan_in ≥ 3)** — verify all 3 consumers, run full test suite, `make build`.
- `internal/web/root_templ.go` appbar + `appbar_context_test.go` (templ regen via pinned version).

### M3 — Child C: Opus 5 doctrine/naming sweep (doc portion of GD-4)
- `model-policy.md`, `moai-constitution.md` heading, CLAUDE.md §12, `context-window-management.md` table (+ Opus 5 row, 256K-row disambiguation), `quality.yaml.tmpl`, `harness.yaml` (local-only), ~10-file naming sweep + mirrors.

### M4 — Children D + E: mechanical doc sync (GD-5/6/7/8/9)
- `native-invocation-model.md` rows + Axis A annotation; CLAUDE.md L211; `dynamic-workflows.md` (manual `/deep-research`, size default/enum/`workflowSizeGuideline`); `settings-management.md` key row; `hooks-system.md` `DirectoryAdded`; `agent-authoring.md` fork/`/subtask`; `skill-authoring.md` + foundation-cc reference `context: fork` background default. All + mirrors.
- **Double-edit hazard**: PR #1146 already touched `hooks-system.md` fork-source documentation on main. M4 MUST re-measure the post-rebase state of `hooks-system.md` (live + template) before editing — apply only the `DirectoryAdded` delta still missing; do not re-apply fork-matcher docs.

### M5 — Child F: docs-site 4-locale + README propagation
- Grep-driven sweep for the four stale claim families; 4-locale same-PR parity; record 0-match evidence where no counterpart exists (REQ-F-002).

### M6 — Closure verification
- Full stale-token absence batch (acceptance.md §D.4), `make build`, `go test ./...`, neutrality + mirror guards, spec lint.

## §G Anti-Patterns

- Blind sed across locales/mirrors (per-file judgment; CJK-aware greps).
- Rewriting doctrine beyond the probe's evidence (claiming depth-3 observed).
- Editing sync-auditor tools list before the M1 decision gate.
- Treating report line numbers as current (content-token anchoring only).

## §H Cross-References

- `.moai/research/cc-update-2.1.207-to-2.1.219.md` §4/§6/§7/§9
- SPEC-SUBAGENT-NESTING-DOCTRINE-001 (superseded encoding), PR #1133, PR #1146
- CLAUDE.local.md §2 (Template-First), §25 (neutrality), §17 (docs-site i18n)
