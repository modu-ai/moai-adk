# Plan-Audit Verdict — SPEC-CODEX-COMMAND-SKILLS-001 (card t503)

Auditor: plan-auditor (independent) · Iteration 1/2 (Tier M ceiling) · 2026-09-07

## Claim

**Verdict: FAIL** — overall score **0.75** (Clarity 0.75 / Completeness 0.75 / Testability 0.75 / Traceability 0.75; harmonic = arithmetic here), below the Tier M plan-auditor PASS threshold **0.80** (`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier). Confidence in the verdict: **0.85**.

All seven must-pass criteria PASS (MP-1..MP-7, see Evidence). The FAIL rests on four blocking-class internal-consistency/verification-adoptability defects (D1–D4), each independently cheap to repair in one revision. The defects do not invalidate the SPEC's architecture — the five card-mandated decisions (D1–D5 of plan.md §) are all genuinely resolved and their resolutions are, on my code-side measurement, correct — but two cross-reference errors would misroute run-phase verification, one AC's mandated RED evidence is unobservable against the shipped mirror, and M3's R-011 conditional has a measurable answer the plan did not take. Reasoning context from the SPEC author: none was passed; the audit ran on the four artifacts only.

## Evidence

### Must-Pass Results

| # | Verdict | Evidence |
|---|---------|----------|
| MP-1 REQ numbering | PASS | R-001..R-011 sequential, no gaps, no duplicates (spec.md:75–149) |
| MP-2 GEARS (requirement layer only) | PASS | All 11 REQs match GEARS: Ubiquitous ×7 (R-001:77, R-002:83, R-005:105, R-006:113, R-008:129, R-009:134, R-010:141), Event-driven ×4 (R-003:89, R-004:97, R-007:121, R-011:147). ACs are Given-When-Then in acceptance.md — the correct verification-layer format (M3 § Scope); not penalized here |
| MP-3 frontmatter | PASS | spec.md:2–14: all 12 canonical fields, canonical names (no `created_at`/`labels`/`spec_id`), `version: "0.1.0"` quoted, `status: draft` ∈ 8-value enum, ISO dates, `phase: "v3.2.0 target"` carries a modern release prefix (not a prohibited stage name). plan.md/acceptance.md carry no `status:` field (artifact statelessness) |
| MP-4 language neutrality | N/A | Template-content publication SPEC; not multi-language tooling. Auto-pass |
| MP-5 D7 cross-SPEC | PASS | `SPEC-CODEX-DUAL-AGENTS-001` status `completed`, `SPEC-CODEX-SKILL-LOADER-001` status `completed` — both measured in this tree (`grep '^status:'`); no retired/superseded/archived reference needing reconciliation |
| MP-6 D8 syscall | PASS | `grep -rn 'syscall'` over the SPEC dir → 0 matches → D8-4 auto-pass |
| MP-7 clarification gate | PASS | `grep -rn 'NEEDS CLARIFICATION'` over the SPEC dir → 0 matches; research.md absent (Tier M 3-artifact set — permitted) |

### Re-measured baseline (independent; do-not-trust-report discipline)

- **16 command sources** confirmed: `clean, codemaps, e2e, feedback, fix, gate, goal, harness, loop, mx, plan, project, review, run, sync` (.md.tmpl) + `todo.md` — matches plan.md §A exactly.
- **34 canonical skill dirs** confirmed under `internal/template/templates/.claude/skills/`.
- **0 collisions** on both bare and `moai-`-prefixed derivations — re-ran the loop myself; scan completed with zero hits. The SPEC's 16/16-clear claim is TRUE.
- Command shape confirmed: locale-conditional `description` with English in the unconditional `{{else}}` branch (plan.md.tmpl:2), `argument-hint`, `allowed-tools: Skill`, one-line body `Use Skill("moai") with arguments: <verb> $ARGUMENTS` (plan.md.tmpl:7); `todo.md` plain-English description confirmed.
- **D1's build-time committed-artifact choice is sound.** go:embed drops symlinks from directory-pattern embeds (skill_mirror.go:11–13 states the constraint verbatim), so the mirror is deploy-time of necessity; real files embed cleanly via `//go:embed all:templates` (embed.go:28, dot-dirs included). The lighter alternative (deploy-time symlink like the 34-skill mirror) is format-infeasible: command sources carry Claude frontmatter (`argument-hint`/`allowed-tools`) and Go template syntax; a symlink would expose both, violating the codex contract the SPEC correctly pins (plan.md B4, R-003b). No overlooked lighter option.
- **Deploy-time distribution is mechanically supported** (R-010): `Deploy walks the embedded filesystem and writes every file to projectRoot` (deployer.go:134–137, walk at :160); SKILL.md has no `.tmpl` suffix → written raw; manifest tracking is automatic in the same walk (deployer.go:270–274, `m.Track(destRelPath, manifest.TemplateManaged, ...)`) — plan.md §B's manifest claim is confirmed.
- **`moai update` survival confirmed**: `ManagedCleanTargets` (internal/cli/update/deploy/deploy.go:56–86) roots are `.claude/settings.json`, `.claude/commands/moai`, `.claude/agents/moai`, `.claude/skills/moai*`, `.claude/rules/moai`, `.claude/output-styles/moai`, `.claude/hooks/moai` (+`.moai/config`) — **no `.agents/` root**. The marked Gap (managed-root membership) resolves favorably by measurement.
- **Catalog axis harmless by measurement**: SlimFS hides only *non-core catalog entries*; non-catalog paths pass through unchanged (slim_fs.go:4, :67–83), and `TestAllSkillsInCatalog` scopes to `.claude/skills` only (catalog_tier_audit_test.go:106). Published skills need no catalog entry to deploy, in full or slim mode. (Recorded as optional finding D8.)
- **Precedent wiring accurate**: `build: agents-emit-check templ-generate` — the drift check IS the first build prerequisite (Makefile:34), so plan.md D4's "same position" is feasible as described; the neutrality audit walks the whole templates root (`neutralityTemplatesRoot = "templates"`, template_neutrality_audit_test.go:63) so AC-012's command is real and auto-covers the new subtree; the CI neutrality workflow's path filter is `internal/template/templates/**` (template-neutrality-check.yaml:27,33) — plan.md's CI claim is accurate.
- **t494 survey claims faithful**: custom-prompts deprecation ("Use skills for reusable prompts", survey:48), C1 emitter adopted (:64), C2 `~/.codex/prompts` rejected (:65), REPO scan path `$REPO_ROOT/.agents/skills` (:147–149), `[[skills.config]]` disable-only / registration-by-path (:171–199) — all spot-verified in the t494 worktree file.

### Defects Found (machine-consumable fix route)

**Blocking (fix before verdict re-entry):**

- **D1** — spec.md:156 — Coverage map says `R-010→AC-013`, but AC-013 is "Cross-platform build stays green (B1)" (acceptance.md:145–151); R-010's distribution is covered by AC-009, which acceptance.md:102 itself labels "(R-011, R-010)". — Severity: major — Class: blocking — Required fix: correct the map line to `R-010→AC-009` (R-011→AC-009 stands).
- **D2** — plan.md:152–153 — M3 cites "mirror coexistence test (AC-009)"; the coexistence test is AC-008 (acceptance.md:91–100); AC-009 is deploy-safety + fresh-init. — Severity: major — Class: blocking — Required fix: change the AC reference in M3 to AC-008.
- **D3** — acceptance.md:99–100 (AC-008 `[RED-first required]`) + acceptance.md:171 (DoD requires RED for AC-005/008/010) vs plan.md:134–136 (§E names RED only for "collision-guard and drift-check tests" — two, not three) — AC-008's RED is **unobservable**: the shipped mirror already skips non-symlink occupants (`skill_mirror.go:198–207` — "Never remove or overwrite it — skip and report") and mirror names derive from the `.claude/skills` walk (deployer.go:152–156, :219–224) with 0/16 measured overlap, so every fixture the real mirror runs against yields skip → green-from-birth. Observing red requires modifying PRESERVE-listed code or stubbing the mirror (asserting nothing about shipped code). — Severity: major — Class: blocking — Required fix: reclassify AC-008 as `regression-guard` per acceptance.md's own preamble convention (:15–16), strike it from the DoD RED list, and align plan.md §E to name AC-005 + AC-010 only.
- **D4** — plan.md:152–156 — M3's R-011 conditional ("implemented only if the existing manifest/update path does not already provide it") is answerable today and the answer changes M3 scope: deployer.go:234 gates **all** provenance checks behind `if !d.forceUpdate`, so update-mode deploy overwrites any file at a template path unconditionally (:242 "template_managed files are safe to overwrite"); only init-mode has protection (deployer.go:229–248). R-011's update-mode half therefore **requires a deploy.go change** (path-scoped provenance for `.agents/skills/moai-*`), which under the plan's own rule means the blocker path is the expected outcome, not a contingency. — Severity: major — Class: blocking — Required fix: fold the measured fact into M3 (expected deploy.go change, its path-scoped shape, and the §24.4 namespace-contract interaction), or amend R-011 to scope the skip-and-report duty explicitly to init-mode with update-mode deferred.

**Optional (orchestrator discretion; surfaced, not verdict-driving):**

- **D5** — spec.md:105–109 (R-005 verbatim body) vs spec.md:136–137 (R-009 generated header "agentemit header precedent") — agentemit places its header as a body-level comment (writer.go:16–17, :105), which would break AC-006's body byte-compare; the constraints force frontmatter placement but no artifact states it. One sentence in plan.md D3/M1 ("header lives inside the YAML frontmatter block") removes the ambiguity.
- **D6** — acceptance.md:131–133 — AC-011's verify reads `0` only when the artifacts are already committed (order-dependent; the parenthetical acknowledges).
- **D7** — acceptance.md:55 — AC-004's `grep -rl '{{'` false-passes on an empty/wrong path (mitigated by AC-001's count check; note for the run-phase fixture test to carry the real weight).
- **D8** — plan.md §B — catalog membership unaddressed; measured harmless (see above). A one-line justification in plan.md §B preempts re-litigation.
- **D9** — spec.md:90–93 — R-003 references the `{{else}}` template branch (implementation detail inside a requirement); kept as disambiguation, defensible, noted for consistency discipline.

### AC-by-AC testability spot verdicts

AC-001/003/004/006/007/011/012/013: binary, named command, executable on this tree's tooling — PASS. AC-002: regression-guard classification sound (no emitter exists pre-run; red unobservable — correct disposition). AC-005: TDD-observable RED (test-first on a constructed fixture, guard implemented second) — sound. AC-010: RED observable on demand by mutating one committed artifact — sound. AC-008: RED-first **unsatisfiable as stated** (D3). AC-009: executable, but its occupancy half inherits D4's under-specified update-mode semantics.

## Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t503`, branch `WT-codex-command-skills`, HEAD `a205c181daff64779fea43323de67b1bda785da9` (short `a205c181d`), measured 2026-09-07, this run. All claims above are measurements taken against this tree in this run; SPEC artifacts are committed at this HEAD.
- Commands run (verbatim, this run): `git rev-parse HEAD && git branch --show-current && git rev-parse --show-toplevel`; `ls -1 …/templates/.claude/commands/moai/`; `ls -1d …/templates/.claude/skills/*/`; collision loop `[ -d … ]` over both derivations; `grep -n -A 25 'func CleanMoaiManagedPaths'` and `ManagedCleanTargets` in `internal/cli/update/deploy/deploy.go`; Read of `internal/template/skill_mirror.go`, `internal/template/embed.go`, `internal/template/deployer.go` (:100–318), `internal/template/slim_fs.go`, `internal/template/templates/.claude/commands/moai/{plan.md.tmpl,todo.md}`; `grep` over `Makefile`, `internal/template/agentemit/writer.go`, `internal/template/catalog.yaml`, `internal/template/catalog_tier_audit_test.go`, `internal/template/template_neutrality_audit_test.go`, `.github/workflows/template-neutrality-check.yaml`; `grep '^status:'` on both referenced SPECs; `grep -rn 'syscall'` and `'NEEDS CLARIFICATION'` over the SPEC dir; `grep` spot-checks of the t494 survey at `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t494/.moai/reports/t494/codex-doc-survey.md`.

## Gaps

What was explicitly NOT observed: (1) no `go test`/`make`/`go build` was executed — the audit is read-only and the emitter does not exist yet, so AC executability was judged from the named commands' feasibility against the measured code paths, not by running them; (2) the t494 survey was spot-checked by targeted grep, not read end-to-end; (3) the update-mode backup flow for user-modified non-`.agents` template files (update_archive.go) was not traced — R-011's finding is scoped to the deployer walk I read; (4) `catalog.yaml`'s 45 entries were counted and sampled, not exhaustively reconciled; (5) the run-phase fixture-test design for AC-005 (which package, which fixture layout) is deferred to M1 by the SPEC's own `<emitterpkg>` placeholder — only its observability was judged.

## Residual-risk

Despite the defects found: (1) if the orchestrator elects to treat D1/D2 as doc-only and override, run-phase verification routing will still be corrected by acceptance.md's own per-AC labels (they are the accurate surface) — the residual is wasted auditor/implementer time, not wrong artifacts; (2) D3 left unfixed risks fabricated E8 evidence for AC-008 mid-run — this is the highest-cost residual and the main reason the verdict is FAIL rather than PASS-with-notes; (3) D4 left unfixed produces a mid-M3 blocker round-trip; the blocker-report-first discipline contains it, but the plan currently prices it as unlikely when it is the measured outcome; (4) the published-body neutrality defect (`Use Skill("moai")`) is deliberately shipped verbatim per t497's boundary — codex-side usability of the 16 published skills is knowingly broken until t497 lands; this SPEC correctly scopes and flags it (R-005/AC-007), and my audit raises no defect there.

## Recommendation (numbered, for manager-spec — one revision cycle)

1. spec.md:156 — replace `R-010→AC-013` with `R-010→AC-009` (D1).
2. plan.md:152–153 — in M3, change "mirror coexistence test (AC-009)" to "(AC-008)" (D2).
3. acceptance.md:99–100 and :171 — reclassify AC-008 as `regression-guard` (green coexistence guard over the measured skip-and-report seam, citing skill_mirror.go:198–207), remove it from the RED-first DoD list; plan.md:134–136 — align §E's RED list to AC-005 + AC-010 (D3).
4. plan.md:152–156 — amend M3 with the measured deploy.go fact (forceUpdate overwrites unconditionally, deployer.go:234) and either pre-scope the path-scoped R-011 change or amend R-011 to init-mode scope (D4).
5. Optional, same pass: one sentence in plan.md D3 pinning the generated header inside the YAML frontmatter block (D5); one line in plan.md §B recording why catalog membership is not required (D8).

---

# Iteration 2 — Re-audit (repair round 1)

## Claim

**Verdict: PASS** — overall score **1.0** (Clarity 1.0 / Completeness 1.0 / Testability 1.0 / Traceability 1.0), at/above the Tier M PASS threshold 0.80. Confidence: **0.90**. Blocking findings: **0**. Score trajectory 0.75 → 1.0 (improving — no STOP signal). Scope: delta re-audit over the iter-1 defect list (D1–D9) per the Retry Loop Contract, plus a new-inconsistency sweep over the revised surfaces; artifacts are uncommitted working-tree revisions of the same three files.

## Evidence

### Regression check — prior-iteration defects (each verified against the revised files, not the author's report)

- **D1 RESOLVED** — spec.md:154-157 coverage map now `R-010→AC-009; R-011→AC-009`, with the parenthetical "(AC-013 is the cross-platform build gate, B1 — not an R-coverage target.)" The map is now correct in both directions against acceptance.md's per-AC labels.
- **D2 RESOLVED** — plan.md:170-171 M3 now cites "mirror coexistence test (AC-008)".
- **D3 RESOLVED, consistent on all three surfaces** — acceptance.md:96 retags AC-008 `[regression-guard]` and drops the RED-first clause from its Verify (:104 → plain `→ GREEN`); acceptance.md:106-113 carries the full classification rationale with the correct citations (skill_mirror.go:198-207 quote, 0/16 walk-derivation overlap, "both prohibited" blocking the mock-RED escape, "no E8 obligation"); DoD :190-192 reads "RED-first evidence present for exactly AC-005 and AC-010 (AC-008 is regression-guard — no E8 obligation; AC-002 likewise)"; plan.md §E :147-154 names "exactly two ACs — AC-005 … and AC-010". No surface retains an AC-008 RED-first obligation; the wording does not fabricate an observable RED (it honestly classifies green-from-birth).
- **D4 RESOLVED** — plan.md:172-185 M3 states the measured fact verbatim (`deployer.go:234` `if !d.forceUpdate` gates all provenance checks; :242 quote; init-only protection :229-248) and reclassifies the update-side path-scoped provenance change as "EXPECTED deploy.go change, not a contingency", scoped to `.agents/skills/moai-<command>` with the §24.4 `.claude/skills` namespace preserved. acceptance.md AC-009 :123-130 is updated to the same mode-split (init half = existing path; update half = the planned change; blocker only on shape divergence). The §A.5 PRESERVE line for skill_mirror.go (:41-42) remains coherent — the planned change is in deployer.go's walk, and M3 says so explicitly. Cross-layer revision sweep (verification-completeness §3): R-011's requirement text needed no change and now has mode-split operationalization downstream; no contradiction with R-010 (a safety modification inside the existing walk is not a new deploy-time distribution path).
- **D5 RESOLVED** — plan.md D3 :117-123 pins the generated header INSIDE the YAML frontmatter block, constrained by AC-004 and AC-006; exact form bounded at M1.
- **D6 RESOLVED** — acceptance.md AC-011 :149-152 rewritten to the order-independent snapshot-diff form (`cp -R` → `make <EMIT>` → `diff -r` → empty).
- **D7 RESOLVED** — acceptance.md AC-004 :55-61 two-stage non-vacuous assertion (count 16 first, then `grep -rl '{{' … | wc -l` → 0) with the honest scoping note ("this grep is the drift net, not the proof").
- **D8 RESOLVED** — plan.md §B :71-76 records catalog-independence with citations matching my iter-1 measurements exactly (slim_fs.go:4,67-83; catalog_tier_audit_test.go:106).
- **D9 RESOLVED** — spec.md R-003 :89-93 no longer names the `{{else}}` branch ("extraction mechanism: plan.md M1"); the mechanism lives in plan.md M1 :159-160 and AC-004's verify prose — the correct layers (ACs may name implementation specifics; requirements should not).

### New-inconsistency sweep (revisions introducing nothing broken)

- RED-first count is exactly 2 (AC-005 :63, AC-010 :132) in every surface that states it — acceptance tags, DoD, plan §E, and the §I repair log agree.
- Coverage map bidirectionally complete: R-001..R-011 all covered; AC-001..AC-012 all trace to existing REQs; AC-013 annotated as B1, not an R target.
- Reference numbers: no stale AC-009-for-coexistence or AC-013-as-distribution references remain anywhere in the three files (read in full).
- Structure unchanged where it should be: frontmatter 12/12 canonical fields with `status: draft`; R-001..R-011 sequential, all still GEARS-shaped (R-003 remains Event-driven after the parenthetical swap); Out of Scope H3 set intact; 13 ACs; plan.md §I is a plan.md body section (not a progress.md §-letter) and cannot collide with the era.go progress.md parser.
- AC-006's `tail -n +5` spot-check fragility remains explicitly subordinated to the binding emitter-package test (in-text) — not a defect.

## Baseline-attribution

Same tree, branch `WT-codex-command-skills`; revised artifacts read from the working tree on 2026-09-07, this run (artifacts intentionally uncommitted per the repair protocol). The revised text's code citations were verified against my iter-1 measurements of the same tree — `deployer.go:234/:242/:229-248/:270-274`, `skill_mirror.go:198-207`, `slim_fs.go:4/:67-83`, `catalog_tier_audit_test.go:106`, `deploy.go:56-86`, Makefile:34 — all previously measured by me in iter-1; no new code measurements were required because the repair cited only already-measured facts, and each quoted figure matched.

## Gaps

Same gaps as iter-1 (read-only audit; no test/make executed; emitter does not exist yet; AC executability judged against measured code paths). Additionally: the M3-planned deployer.go change's final shape is a run-phase deliverable — this audit validated its scoping and constraint set, not its implementation.

## Residual-risk

(1) The deployer.go update-side change touches a hot deploy path — its correctness is run-phase's verdict surface (fresh-init + update fixture tests of AC-009 cover it). (2) The published bodies ship knowingly Claude-only (t497 boundary) — unchanged, correctly scoped. (3) Header exact form deferred to M1 under a two-AC constraint set — self-correcting if violated.

## Iter-2 Recommendation

Proceed to run-phase entry per the normal gates (Implementation Kickoff Approval unchanged; the plan-audit skip contract applies only if the run-gate re-checks artifact-hash against this verdict). Run-phase carry-forward notes for the delegation prompt: RED-first obligations exist for exactly AC-005 and AC-010 (E8 verbatim evidence required); AC-008 and AC-002 carry no E8 obligation; M3's deployer.go change is planned scope with the §24.4 namespace note, blocker report only on shape divergence from plan.md M3.

