# SPEC-CLOCAL-AUDIT-001 — Implementation Plan

## A. Context

Card t308 audits `CLAUDE.local.md` @ develop `d29b8942e` (663 lines, branch `WT-clocal-audit` in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t308`). The frozen claims inventory (76 items, INV-001..INV-141; Group X1 added at plan-audit iter-1 repair) lives at `.moai/reports/t308/claims-inventory.md`. This is a Tier M documentation-audit: no production code changes; the deliverable is a corrected `CLAUDE.local.md` plus an evidence-bearing report pack.

Development mode context: quality gates are scoped to documentation (no Go tests required); TRUST 5 applies as Readable/Unified/Trackable on the corrected prose and Trackable evidence trail.

## B. Known Issues

1. The git-flow transition commit rewrote §4.1 hours before this card (lines 273–337 at HEAD). Residual §4.1 defects are already registered under t294/t295/t298/t303 — do not re-find (REQ-CLOCAL-004).
2. Internal contradiction candidate confirmed at freeze time: the Local-Only block (line 118) frames `.claude/hooks/moai/handle-*.sh` as "Generated hook wrappers (not templates)" while §2.3 (drift subsection) discusses the same wrappers as `.sh`/`.sh.tmpl` PAIRS living in `internal/template/templates/.claude/hooks/moai/` requiring synchronized edits. Exactly one framing can be true NOW (INV-033).
3. Header claims "Last Updated: 2026-05-25" while heavy edits landed through August — near-certain DEFECT-CONFIRMED pending the mechanical date check (INV-001).
4. Two consecutive `---` blocks observed between §11 close and §12 heading (confirmed in sed output around lines 488–497) create an empty section — cosmetic-but-real (INV-002).

## C. Pre-flight

- [x] Worktree pinned: HEAD `d29b8942e5b237ef7180749dc04273d83647c745`, branch `WT-clocal-audit`.
- [x] `wc -l CLAUDE.local.md` → 663; `git diff --stat HEAD -- CLAUDE.local.md` clean at plan time.
- [x] SPEC ID self-check executed: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS` → PASS.
- [x] Duplicate-ID sweep: no existing `CLOCAL` SPEC in `.moai/specs/` (name scan + content grep both empty).
- [x] Frontmatter validated against `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields.
- [ ] Run-phase re-verifies HEAD immediately before first edit and before commit (REQ-CLOCAL-001).

## D. Constraints

- Write-capable paths: `.moai/specs/SPEC-CLOCAL-AUDIT-001/**`, `.moai/reports/t308/**`, `CLAUDE.local.md` (run-phase only).
- No full-suite test runs; no background load; simple individual Bash commands only (shared hardware with active factory lanes).
- PRIMARY checkout strictly read-free except the authorized §28 probes.
- Every cited command output is recorded in `.moai/reports/t308/checks-transcript.md` with exit codes — a claim without recorded output is a Gap, not a Pass.

## E. Self-Verification

E1 Inventory closure — every INV item carries exactly one REQ-CLOCAL-007 verdict class.
E2 Evidence attribution — every DEFECT-CONFIRMED names command + verbatim output + exit code; five-section verdict draft complete.
E3 Diff discipline — final `git diff` of CLAUDE.local.md contains ONLY minimal corrections traceable to INV IDs.
E4 Scope guard — no file outside the three write-capable paths modified; `git status --short` proves it.
E5 No-history rewrite — historical/incident records retain narrative with dated correction markers where anchors drifted.

## F. Milestones (priority-ordered by decision-reversibility — highest-change-likelihood decisions first; no time estimates)

### M1 (High) — Contradiction and trap resolution [Priority High]
Decision-laden items that shape later milestones:
- INV-033 handle-*.sh contradiction: inspect `internal/template/templates/.claude/hooks/moai/` for `handle-*.tmpl` presence AND pair state of locals; fix whichever framing is false now (plus the companion `.sh`/`.sh.tmpl` drift-check recipe INSIDE §2.3 stays excluded unless the defect text sits outside it — the contradiction text itself lives at line 118, OUTSIDE §4.1, so it is IN scope).
- INV-056 local-linear-integration.md tracked-vs-untracked (`git ls-files --error-unmatch` + filesystem check); state which surfaces see it.
- INV-057 ambiguous-path finding for `memory/audit_sweep_patterns.md` Pattern A and `memory/lessons.md` #4/#5 (no directory anchor; determine whether in-repo or global Claude projects memory referent).
- INV-049 astgrep relocation narrative: verify `.moai/astgrep-rules/` six yml files + gate.yaml wiring + CLI `--rules-dir` default still old path (source grep of `internal/cli/gate.go`). (INV-012, its §1 embed-pointer sibling, is scheduled with Group G in M7.)

### M2 (High) — Cross-document consistency [Category F]
- INV-110 §18 row summary vs current `.moai/docs/git-workflow-doctrine.md` content (transition edits landed).
- INV-111 §23 row summary vs current `.moai/docs/git-local-workflow-doctrine.md` content.
- INV-112 §2.3 reinstall recipe (line ~180-183): verify worktree HEAD `.moai/config/sections/git-strategy.yaml`; if HEAD carries template github-flow values instead of committed gitflow keys, the recipe's restore step is broken → DEFECT-CONFIRMED.
- INV-055 References-row existence sweep (11 `.moai/docs/*.md` pointers + SKILL.md) — mechanically cheap, folded here.
- INV-059 §19 row pointer files exist; INV-113 pointer-validation for out-of-§4.1 sections referencing §4.1 subsections.
- iter-1 additions: INV-137 (§12 skill-authoring/agent-authoring pointer pair, L499); INV-140 (References §27 Agent-Skill row surfaces, L645).

### M3 (High) — Symbol + line citations [Category B]
Verify symbol name AND quoted line range hold at `d29b8942e`: INV-070 (`deploy.go:29` CleanMoaiManagedPaths), INV-071 (`deploy.go:122` → `.moai/config` root specifically; whole 8-root list matches lines 152–154), INV-072 (`plan.go:73` `if IsMoaiManaged(...) { continue }`), INV-073 (`update.go:513` archiveLegacySkills call site — citation check only, defect NOT re-adjudicated), INV-074 (`gate.go:160-161` empty-rules_dir fallback), INV-075 (`Makefile:9` LDFLAGS; `Makefile:38` install recipe), INV-076 (`pkg/version` Commit="none"/Date="unknown").

### M4 (High) — Defaults, env-var surface [Categories C, D]
- INV-080 AstGrepGate literal `{Enabled: true, BlockOnError: false, WarnOnlyMode: true}` in `internal/config/defaults.go`; loader chain `loader_gate.go` ← `Loader.Load`.
- INV-081 hook timeout "default 5초" + `{"timeout": 60}` recipe vs actual template settings hook timeouts (cite real values).
- INV-082 coverage targets 85% / 90%+ enforcement location hunt.
- INV-090 `applyEnvOverrides` switch cases == claimed five-var set; `envkeys.go` constants.
- INV-091 zero-reader assertion for `MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG` — full-tree grep `internal/ pkg/ cmd/` WITH named scan scope recorded.
- INV-092 configuration-priority order + section file inventory (config.yaml main + sections list).
- iter-1 addition: INV-138 (§13 loadGLMKey() symbol existence in internal//cmd scope + ~/.moai/.env.glm path convention, L521).

### M5 (Medium) — Template-tree structure + tooling surface [Categories E, G]
- INV-100 embed.go dual directive + no generated embedded.go; INV-101 actual template settings filename (`settings.json` vs `settings.json.tmpl` — internal inconsistency candidate between §8 line 402 and Local-Only line 117); INV-102 TemplateContext fields + constructor; INV-103 `deployer.Deploy(ctx, projectRoot, mgr, ctx)` signature; INV-104 status_line.sh.tmpl PATH export; INV-105 renderer.go claudeCodePassthroughTokens ($HOME + posixPath); INV-106 project_markers detection; INV-107 16-language list + Dart-canonical-flutter + catalog.yaml parity.
- iter-1 addition: INV-136 (§3 Code Standards cluster L230–238 — coding-standards.md auto-load pointer existence, snake_case naming, fmt.Errorf %w wrapping vs rules content).
- INV-120 no top-level `moai build` (CLI help/source); INV-121 slash-workflow wiring (.claude/skills/moai/workflows/*); INV-122 `SPLIT_HARNESS_NAMESPACE_LEAK` sentinel grep; INV-123 make targets build/install/help with `##` help.

### M6 (Medium) — LSEL §28 via PRIMARY [Category H]
Reads at the authorized primary surfaces ONLY (NAME THE SURFACE per finding): INV-130 two SessionStart lsel entries each `"timeout": 30` in primary `.claude/settings.local.json` (jq); INV-131 wrapper scripts `session_drain.sh`/`backlog_check.sh` existence + tracked-status + `--inbox/--state-dir` flag surface; INV-132 underlying `drain.sh` exists and overwrite-behavior description matches; INV-133 `.moai/state/lsel/clusters-history/` exists in primary; INV-134 default threshold 25 + `LSEL_BACKLOG_THRESHOLD` key located (script/SKILL.md); INV-135 `.claude/skills/hns-lsel-curator/SKILL.md` § Durable operations + § Verification headings exist (tracked, worktree-readable).

### M7 (Low) — Header + formatting + Group T1 + historical anchors
- INV-001 Last Updated staleness; INV-002 double `---` empty section; INV-003 §-numbering drift from consolidation stubs.
- Group T1 items named explicitly: INV-021 CI-guard `.github/workflows/template-neutrality-check.yaml` existence (subject L110); INV-022 template-isolation-doctrine §25.1/§25.3 anchors + sibling `internal_content_leak_test.go` existence (subject L112).
- Historical-anchor spot checks: INV-044 #1557 `ed04e40e6` commit exists; INV-083 v3.0.1 tag exists.
- iter-1 additions: INV-139 (§11 inward cross-ref reachability to §6 Go Test Execution Rules — links INV-075/076 which own the same-line Makefile/pkg-version citations); INV-141 (§20 Vercel pricing row recorded as EXTERNAL-UNVERIFIABLE, zero further effort).
- All remaining inventory sweeps not absorbed above (INV-010..011, 020, 030–049 pass/fail attestations for Local-Only block entries, INV-121 complements).

### M8 (Low) — Fixes, report consolidation
Apply fix routing (REQ-CLOCAL-008) per milestone as findings land rather than batching; M8 closes: populate `verdict-draft.md` five sections, ensure transcript completeness (E2), paste progress notes into `SPEC-CLOCAL-AUDIT-001/progress.md`, final `git status --short` scope proof (E4), HEAD re-read before commit.

## G. Anti-Patterns (forbidden in run-phase)

- Text-pattern inference promoted to DEFECT-CONFIRMED without an executed command (the failure mode that caused the loader_gate incident).
- Editing cross-doc staleness directly in `.moai/docs/*` instead of recording new-card recommendations.
- Re-running historical incidents to "prove" a narrative.
- Sweeping §4.1 for already-registered defects.
- Full-suite or background-load verification on shared hardware.
- Dropping unresolved hypotheses silently.

## H. Cross-References

- Claims inventory: `.moai/reports/t308/claims-inventory.md`
- Verdict draft: `.moai/reports/t308/verdict-draft.md`
- Check transcript: `.moai/reports/t308/checks-transcript.md`
- Sibling cards carrying §4.1 defects: t294, t295, t298, t303
