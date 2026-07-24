# SPEC-SEC-DEEPSCAN-001 — Implementation Plan

> Milestones are ordered by decision-reversibility: the highest-change-likelihood decisions (surface/UX contract, adversarial-panel semantics, results-dir schema) lead; mechanical mirror/registration steps trail. Human review should focus on M1-M3.

## §A Context

### A.1 Work location & distribution model

- **Template source (author FIRST)**: `internal/template/templates/.claude/…` (Template-First, CLAUDE.local.md §2).
- **Local sync (author SECOND, lockstep)**: `.claude/…` siblings.
- **Build**: `make build` re-embeds the template tree via `//go:embed all:templates`.
- This SPEC is **markdown-first**: the run-phase deliverable is playbook prose + an argument-hint edit + an optional delegation-map edit. There is NO Go runtime change (REQ-SDS-062). This mirrors the SPEC-E2E-REVIVAL-001 distribution shape (workflow-skill + command edits, both trees).

### A.2 Extension points (verified 2026-07-24)

1. **Workflow skill** — `.claude/skills/moai/workflows/review.md` + template sibling (currently byte-identical, 20961 bytes; edits MUST stay lockstep). Today it carries flags `--staged --branch --security --file --design --critique --lean --repo` and a Mode-4 parallel 4-perspective fan-out with sync-auditor as verdict owner. The `--deep` mode is a NEW top-level section.
2. **Command wrapper** — `internal/template/templates/.claude/commands/moai/review.md.tmpl` → rendered local `.claude/commands/moai/review.md`. `argument-hint` currently `"[--staged] [--branch] [--security]"`; gains `[--deep]` and `[--patch]`.
3. **Delegation map** — `.moai/config/sections/delegation.yaml` `review:` entry (agents: `[sync-auditor]`; skills: `[moai-foundation-quality, moai-ref-owasp-checklist]`). OPTIONAL deep-mode note — the 4 security ref skills are injected per-phase via `Skill()` (REQ-SDS-011), NOT added to the static preload (adding 4 heavy skills to every review preload is a token regression).

### A.3 PRESERVE list (do NOT modify)

- The existing `--staged / --branch / --security / --file / --design / --critique / --lean / --repo` behavior in `review.md` (REQ-SDS: out-of-scope "existing single-pass lens unchanged").
- `catalog.yaml` agent/skill counts (`expectedAgentCount`, `expectedTotal`, `expectedSkillCount`) — no new agent, no new skill dir (REQ-SDS-062).
- All Go source under `internal/`, `pkg/`, `cmd/`.
- `.claude/workflows/*.js` (user-owned, not template-managed) — no shipped script.
- Runtime-managed files (`.moai/state/*`, `.moai/reports/*` existing, `.moai/cache/*`).

## §B Known Issues (auto-injected, filtered for a markdown-only SPEC)

- **B4 — Frontmatter canonical schema**: spec.md uses `created:`/`updated:`/`tags:` (done). Snake_case aliases prohibited.
- **B6 — spec-lint heading convention**: `## Out of Scope` (h2) alone → `MissingExclusions` ERROR. spec.md uses `### Out of Scope — <topic>` h3 sub-headings with `-` bullets (done — 6 sub-headings).
- **B8 / B10 — working-tree hygiene & scope discipline**: run-phase commits touch ONLY the review workflow skill (both trees), the review command wrapper (both trees), and optionally delegation.yaml. `git add` specific paths only; never `-A`.
- **B9 — direct commit + push (Route B for Tier L)**: Tier L routes through a PR per `spec-workflow.md` § SPEC Phase Discipline Route B. `manager-git` opens the PR; run-phase manager-develop commits on the feature branch.
- **B11 — AskUserQuestion boundary**: the playbook MUST NOT instruct any agent to prompt the user (REQ-SDS-063). This is a content-level constraint on the shipped prose itself.
- **Template neutrality (§25 / §15)**: the shipped playbook MUST NOT embed `SPEC-SEC-DEEPSCAN-001`, internal dates, or commit SHAs, and MUST NOT privilege Go. CI guard: `internal/template/internal_content_leak_test.go` + `template-neutrality-check.yaml`.

## §C Pre-flight (run before any edit)

```bash
# 1. Branch + baseline
git branch --show-current && git rev-parse HEAD
# 2. Confirm review.md template/local parity (must be byte-identical pre-edit)
diff .claude/skills/moai/workflows/review.md \
     internal/template/templates/.claude/skills/moai/workflows/review.md && echo "PARITY OK"
# 3. Confirm no /moai security residue would be reintroduced
grep -rn "moai security\|security subcommand" .claude/skills/moai/ | grep -v review || echo "no security-subcommand refs"
# 4. Spec-lint the new SPEC (dir arg NOT supported — lint from repo context)
go run ./cmd/moai spec lint 2>&1 | grep -i "SEC-DEEPSCAN" || echo "check strict subset"
# 5. Neutrality baseline
grep -rn "SPEC-SEC-DEEPSCAN\|2026-07-24" internal/template/templates/.claude/skills/moai/workflows/review.md || echo "neutral"
```

## §D Constraints (DO NOT VIOLATE)

- REQ-SDS-002 / REQ-SDS-062: no `/moai security` revival; no new agent; no Go runtime; no shipped `.js`.
- REQ-SDS-060 / REQ-SDS-061: Template-First + 16-language neutrality on all shipped content.
- REQ-SDS-063: no agent-prompts-user directive anywhere in the playbook.
- Both-tree lockstep: every workflow-skill edit lands in BOTH `.claude/` and `internal/template/templates/.claude/` in the same commit; the command wrapper edits the `.tmpl` (template) + rendered (local).
- Never `--no-verify`, never force-push, `git add` specific paths only.

## §E Self-Verification (run-phase completion evidence targets)

The run-phase completion report (manager-develop §E) MUST carry, per the 5-section evidence format:

- **E1 — AC matrix**: every AC in acceptance.md PASS/FAIL with the verbatim grep/diff command + output.
- **E-parity**: `diff` of the two `review.md` copies → identical (byte-parity).
- **E-neutrality**: neutrality grep on the edited template files → 0 internal-token matches.
- **E-lint**: `go run ./cmd/moai spec lint --strict` subset for this SPEC → 0 errors; repo-wide residual debt reported separately (repo-global exit code is NOT this SPEC's regression).
- **E-regression**: `go test ./internal/template/...` (commands audit, catalog counts, neutrality) → PASS; `go build ./...` → exit 0.

## §F Milestones (reversibility-ordered — highest-change-likelihood first)

### M1 — Surface & UX contract (`--deep` / `--patch` / `--commit`, job menu)
**Change-likelihood: HIGH (user-facing flag contract).**
- Add the `--deep` mode entry + `--patch` opt-in + `--commit <SHA>` scope to the `Supported Flags` section of `review.md` (both trees).
- Update the command `argument-hint` (`.tmpl` + rendered) to include `[--deep]` and `[--patch]`.
- Write the job-menu mapping (whole-repo / diff / commit / patch → existing scope flags) into the `--deep Mode` section header.
- Covers: REQ-SDS-001, -002, -003, -004.
- Review focus: is `--deep` the right composition with `--security`? Is `--commit` the right new scope token? (design.md §S records the surface decision + alternatives.)

### M2 — Six-phase pipeline + adversarial-panel semantics
**Change-likelihood: HIGH (verification rigor is the core value).**
- Author the `--deep Mode` playbook body: the six phases in order (architecture map → threat model → hunt → adversarial verify → report → patch), each with its agent role, read-only vs write scope, and `Skill()` injection for the hunt phase.
- Specify the 3-voter panel (REACHABILITY / IMPACT / DEFENSES), the 2-of-3 quorum admission rule, the non-unanimous "medium" confidence cap, voter independence, and rejected-candidate exclusion (appendix-only).
- Covers: REQ-SDS-010, -011, -012, -020, -021, -022, -023.
- Review focus: does the quorum + cap wording exactly match the absorbed contract? Is voter independence enforceable in prose?

### M3 — Results-directory schema
**Change-likelihood: HIGH (artifact contract downstream tools depend on).**
- Specify the timestamped results dir under `.moai/reports/security-deepscan-<timestamp>/`: `report.md` (F-IDs + impact/exploit/severity/confidence/recommendation), `findings.jsonl` (one finding per line), a revision stamp (scanned commit + effort tier + working-tree-included boolean), and a self-`.gitignore` (`*`).
- Classify explicitly as a REPORT (not a SPEC) per the classification rules.
- Covers: REQ-SDS-040, -041, -042, -043, -044.
- Review focus: is the JSONL line schema stable enough to be machine-consumed? Is the revision-stamp field set complete?

### M4 — Patch drafting + reviewer-vouch gate
**Change-likelihood: MEDIUM (behavior-shaped but bounded by never-apply invariant).**
- Specify the `--patch` flow: scratch-clone drafting via `Agent(isolation: "worktree")`, an independent reviewer agent (distinct from the drafter) vouching for the 3 claims, vouch-failure → short note, and the never-auto-apply / one-finding-one-patch-one-PR invariant.
- Covers: REQ-SDS-030, -031, -032, -033.
- Review focus: is "independent reviewer" clearly a different spawn than the drafter? Is the never-apply invariant unambiguous?

### M5 — Graceful degradation ladder
**Change-likelihood: MEDIUM (prerequisite-conditioned fallback).**
- Document the Dynamic Workflows prerequisite (v2.1.154+) + the availability signal(s).
- Author the ladder: PRIMARY = Workflow() (Mode 6); FALLBACK = Mode-4 bounded parallel (3-5 concurrent) fan-out preserving the 2-of-3 quorum; DEGRADED = single-pass `/moai review --security` + native `/security-review`.
- Covers: REQ-SDS-050, -051, -052.
- Review focus: does the degraded path really preserve the adversarial quorum, or does it silently drop rigor?

### M6 — Delegation-map + AskUserQuestion boundary + neutrality pass (mechanical)
**Change-likelihood: LOW (mechanical wiring + guard compliance).**
- OPTIONAL delegation.yaml note (per-phase `Skill()` injection documented; static preload unchanged).
- Verify the playbook contains no agent-prompts-user directive (REQ-SDS-063); route all decisions through the orchestrator.
- Run the neutrality grep + both-tree parity diff; align local ↔ template.
- Covers: REQ-SDS-060, -061, -062, -063.

## §G Anti-Patterns (avoid)

- **Shipping a `.js` workflow script** into the template tree (violates the "workflows not template-managed" rule + REQ-SDS-062). The Workflow is runtime-constructed from the playbook.
- **Adding the 4 security ref skills to the static review preload** (token regression) — inject per-phase via `Skill()`.
- **Writing results under `.moai/specs/`** (a scan is a REPORT, not a SPEC — REQ-SDS-040).
- **Degrading rigor, not just scale**, on the fallback path (REQ-SDS-052 forbids dropping the quorum).
- **Go-specific examples in the playbook** (16-language neutrality — REQ-SDS-061).

## §H Cross-References

- Surface decision + phase pipeline + panel mapping: `design.md`.
- Codebase research (review surface, dynamic-workflow primitives, ref skills, retirement constraint): `research.md`.
- Reused primitive: `.claude/rules/moai/workflow/dynamic-workflows.md`.
- Distribution precedent: SPEC-E2E-REVIVAL-001 (workflow-skill + command, both trees).
- Retirement constraint: SPEC-SUBCOMMAND-RETIRE-001 (`/moai security` removed).

## Resolved Decisions (settled — user-confirmed via AskUserQuestion)

- **Single-commit scope token** — RESOLVED: a new `--commit <SHA>` flag on `/moai review`, consistent with REQ-SDS-003 (which already commits to `--commit`). No REQ/AC/design softening; `--commit` is the firm requirement.
- **Results-dir retention** — RESOLVED: no auto-prune for the first cut; retention is left to the user, revisit in SPEC-2. Recorded in spec.md §C Out of Scope.
