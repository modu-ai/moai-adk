# plan.md — SPEC-SETTINGS-ORIGIN-001

Tier S investigation card. Harness: minimal. No production code changes — run-phase output is analysis + one verdict artifact.

> Artifact-set note: Tier S's minimal set is spec.md + plan.md; the t487 dispatch explicitly enumerates acceptance.md, so this plan emits the additive third file. The tier remains S (plan-auditor threshold 0.75; investigation card, no code).

## §A Context

The dirty `.claude/settings.json` working copy from the t452 window is preserved (`b669972dc738d1bf925281dcc90f152e`, re-verified 2026-09-05). t480 eliminated six origin candidates and disproved the loss premise, leaving "writer unidentified" as its Gap 1. t485 C4 bounds the method: state forensics of the persisted artifact is valid; expecting repo/config state to NAME the writer process is dead. Live axes: code-path inventory (which writers could exist), distribution (systematic vs one-off), external inflow.

Known code-path entry points to inventory (starting map, NOT exhaustive — REQ-001 requires the sweep to prove exhaustiveness, not assume this list):

- `internal/cli/update/deploy/` — `CleanMoaiManagedPaths` + template redeploy (`deploy.go`) — writes `.claude/settings.json` on `moai update` (deletes-then-redeploys; CLAUDE.md §2.3).
- `internal/template/templates/.claude/settings.json.tmpl` + renderer — the render source (`TemplateContext`, `renderer.go`).
- SessionStart / launcher flows (`moai cc` / `moai glm` / `moai cg`) — modify settings files at runtime (CLAUDE.md §2 settings.local.json notes claim local-only, but REQ-002 must verify whether any flow touches the tracked file).
- `.claude/hooks/moai/*.sh` wrappers and `.sh.tmpl` pairs — verify none write settings.json.
- `Makefile` targets (`make build`, `make install`) — embed/regenerate, verify no direct settings write.

## §B Known Issues

- The `.tmpl` template file is Go-template syntax — `jq` cannot parse it directly (measured: parse error at line 130). Key-order comparison vs the template must extract keys textually with `grep -nE '^[[:space:]]*"[a-zA-Z$]+"' <tmpl>` (auditor-verified: yields all 13 top-level keys with the exact line numbers cited in spec §1; the naive `grep -n '^[a-zA-Z]'` is vacuous on this file — keys are indented and quoted, verified 0 matches) or render first.
- t480's E4 table measured release templates **only on the ask-list axis**; full-file key-order vs older releases is an open axis (H2) — but the ask-list comparisons themselves must NOT be re-run (REQ-008).
- The t452 worktree was restored after the merge; its current settings.json is the develop version — the dirty state exists only in the preserved copy. Any "check the original tree" shortcut is therefore void.
- 70+ worktrees expected in the sweep; output must be bounded per tree (quiet git forms, redirect-to-file contract for the full table).

## §C Pre-flight (before M1)

1. Confirm the preserved copies still match their md5 (one command, both files) — the entire SPEC keys off that byte-preserved artifact.
2. Confirm `git worktree list` enumerates (count it; record N).
3. Confirm this worktree is at base `25a3212a9` / branch `WT-settings-origin` (branch and HEAD read-back).

## §D Constraints

Verbatim from spec.md §4 (binding): byte-preserve dirty copies (path+md5 only); sweep is read-only; no background load; no full-suite test runs (no test runs expected at all); verification-claim integrity on every claim; dispose of no worktree; push nothing; do not re-run t480's six eliminations.

## §E Self-Verification

Run-phase close self-check (each item = command + observed output recorded in the verdict):

- [ ] Q1 inventory: a cross-check search (e.g., repo-wide grep for write-target strings like `settings.json` in write contexts — `os.WriteFile`, `Write` targets, `jq`/redirect writes in shell), bounded only by the repository root (including `scripts/`, `.claude/workflows/`, `.github/workflows/`), surfaces no code path absent from the inventory table.
- [ ] REQ-008 baseline reuse: t480's six eliminations appear as cited baseline only (no re-execution in the verdict Evidence); §1 fingerprint observations appear as hypotheses (H1–H4), not conclusions.
- [ ] SWEEP: tree count enumerated == trees checked; every dirty tree carries path+md5+3 markers; the primary checkout is included.
- [ ] Q2 executed iff Q1 found no producer (conditional honored — state which branch was taken and why).
- [ ] Q3: exactly one recommendation; it names its evidence premise and its t485-C4 compliance (process-level if it proposes live observation).
- [ ] verdict.md exists at the expected path with all 5 sections; empty-output and absence-of-signal claims are labeled Gaps, not zeros.

## §F Milestones (priority-ordered; ordered by decision-reversibility — the measurements most likely to change the conclusion come first)

- **M1 — SWEEP: worktree distribution (REQ-003)** — Priority High. `git worktree list` → per-tree `git --no-optional-locks -C <wt> status --porcelain -- .claude/settings.json` (quiet, read-only; `--no-optional-locks` is mandatory — plain `git status` takes index write locks in trees other lanes may be using) → for dirty trees: md5 + ask-list length + matcher form + key order. The primary checkout is entry 1 of the listing. Output: a distribution table in the verdict. *Rationale for first position: this single measurement discriminates H4 (systematic vs one-off) and re-weights every later hypothesis; it is also pure read-only so it carries zero risk.*
- **M2 — Q1: code-path inventory (REQ-001)** — Priority High. Repo-wide search for every writer of `.claude/settings.json` across `internal/`, `pkg/`, `cmd/`, `.claude/hooks/` (both `.sh` and `.sh.tmpl`), template sources, Makefile. Each entry: file:line + what it writes.
- **M3 — Q1: shape reconciliation (REQ-002)** — Priority High. For each M2 entry, could it produce the §1 fingerprint? (Key-order reproduction test: for render paths, compare rendered order vs dirty order; for runtime-edit paths, does the edit preserve or rewrite key order?)
- **M4 — Q2: external inflow (REQ-004)** — Priority Medium, CONDITIONAL on M3 finding no producer. Other checkouts on the machine (`~`, other project roots — read-only), old binaries on disk, TypeScript moai-adk predecessor templates (npm, if fetchable — cite the fetched source or record "not fetchable" as a Gap), Claude Code runtime write signature (does CC append keys on upgrade? — verify from CC docs/changelog, not assumption), manual hand-edit signature (formatting, key-addition order).
- **M5 — Q3 + verdict close (REQ-005, REQ-006)** — Priority Medium. One grounded recommendation (process-level if it proposes observation); write `.moai/reports/t487/verdict.md` (5 sections); run §E self-check; report card completion with branch/HEAD/evidence path. No sync phase expected (evidence-only close, t480 precedent — noted in spec.md §5).

## §G Anti-Patterns (named, from this repo's lessons)

- **Config-state digging for the writer's name** — t485 C4 dead axis. State forensics of the artifact itself is in scope; expecting state to name a process is not.
- **Re-running t480's eliminations** — accepted baseline; re-running burns the window and adds nothing.
- **Empty output read as zero** / **absence-of-signal read as evidence** — every sweep row must quote the command output or be labeled a Gap.
- **Modifying a found dirty copy** "to compare" — byte-preservation is absolute (user env values may be mixed in).
- **Sweep as background load** — serial or modestly-parallel read-only git only; many lanes share this machine.
- **Hand-enumerating a discriminator regenerates its own defect** — the M2 inventory must be grep/cross-check-driven, not memory-driven.

## §H Cross-References

- spec.md §1 (fingerprint + hypothesis tree H1-H4) — M3/M4 consume it; M1 discriminates H4.
- acceptance.md AC-001..AC-007 — the close gate for §E.
- t480 verdict Evidence E1-E4 / Gaps — the baseline this SPEC extends (Gap 1 = this card; Gap 3 = the M1 sweep).
- `verification-claim-integrity.md` §1.1 surface 3 — defect/identification claims need the domain tool, not text-pattern inference.
