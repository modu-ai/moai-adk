---
id: SPEC-UPDATE-DOC-DRIFT-001
title: "always-loaded instruction drift: maintainer documentation that asserts a mechanism the code contradicts is not a stale comment — it is an input that misdirects every agent session"
version: "0.3.0"
status: draft
created: 2026-07-31
updated: 2026-08-14
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "docs"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "documentation, drift, claude-local, instruction-correctness, always-loaded, config, envkeys, agency"
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-DATA-SURVIVAL-001, SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-KEY-HONESTY-001, SPEC-UPDATE-CI-GUARD-001, SPEC-INTERNAL-ARCH-001]
depends_on: []
---

# SPEC-UPDATE-DOC-DRIFT-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Epic SPEC 6 of 6 — the closing SPEC of the four-lens audit of `moai update` / `.moai/config`. Findings F1-F5 each re-verified while authoring; F1 found false in three independent ways rather than the two supplied; F3 found false at the template path as well as the local path. Three drifts recorded (§A.6). |
| 0.2.0 | 2026-07-31 | Plan-audit revision (iteration 1 verdict **FAIL, 0.65**; Testability 0.50). D1-D11 resolved, D12 folded into D1, D16 resolved; D13-D15 and D17 deferred. §A.5's option-B cost framing found inverted and rewritten against measurement; the `--dry-run` decision settled as **option B**. `acceptance.md` gained §A clause 7 (anti-vacuity); ten criteria rewritten under it. |
| 0.3.0 | 2026-08-14 | **Staleness rewrite.** Every one of REQ-UDD-001..013 re-measured against worktree HEAD `7f61332ef` (branch `docs/spec-doc-drift-rewrite`); the v0.2.0 baseline `d5336214e` is **not an ancestor of this tree** (`git merge-base --is-ancestor` exits `1`), so every v0.2.0 `file:line` and count was re-observed rather than carried over. Since v0.2.0, `CLAUDE.local.md` was consolidated: former §18-§27 were externalized into `.moai/docs/*.md` and the file now runs §1-§17 plus a `## References` table (511 lines). **Retired with evidence (4)**: REQ-UDD-005 (E4 landed and the reconciliation was performed), REQ-UDD-011 and REQ-UDD-012 (option B was implemented by the sibling E1 as REQ-RIL2-024/025/026, with the early-return constraint honoured), REQ-UDD-013 (its escalation target is moot). **Re-anchored (3)**: REQ-UDD-004 to `.moai/docs/local-dev-settings-intent.md`, REQ-UDD-007 to `.moai/docs/version-management.md`, and REQ-UDD-002/003's site from `CLAUDE.local.md:141` to `:146`. **Kept live, narrowed (2)**: REQ-UDD-002 and REQ-UDD-003 — §2.2 gained a dated correction on 2026-08-01 that retracts all four false claims, so the falsehood half of each is discharged; what survives is a *new* misstatement (§A.1) and two unmet positive obligations. **Kept live (5)**: REQ-UDD-001, REQ-UDD-006 (with `.agency/` folded in as its concrete target), REQ-UDD-008, REQ-UDD-009, REQ-UDD-010. **REQ-UDD-004's polarity inverted**: two production readers for `auto_cleanup` have appeared since v0.2.0, so the drift is no longer "the doc claims an enforcement that does not exist" but its mirror image. **`REQ-ARCH-006` duplicate resolved** in favour of this SPEC (§A.7). |

## §A Problem / Motivation

The five preceding SPECs of this Epic are about code and config that lie to the **user**. This one is
about documentation that lies to the **maintainer** — and therefore to every agent session, because
`CLAUDE.local.md` and `internal/config/CLAUDE.md` are loaded as project instructions on every turn.

That loading property is the whole reason this is a SPEC rather than a cleanup chore. A false
statement inside an always-loaded instruction file is not a stale comment sitting unread next to the
code it describes. It is an **input**: it enters the model's context before any work begins, it is
read as authoritative project doctrine, and it actively steers work in the wrong direction. A
maintainer who reads "nothing here can block a commit" does not think to check whether something
blocks. An agent that reads the same sentence has no independent reason to check either — the
instruction file is precisely the surface it is told to trust.

**Baseline.** Worktree HEAD `7f61332ef`, branch `docs/spec-doc-drift-rewrite`, based on `origin/main`.
Every figure below was observed on this tree while authoring v0.3.0. The v0.2.0 baseline `d5336214e`
is not an ancestor of it, so nothing was carried over.

### A.0 What the always-loaded surface is, after the consolidation

The v0.1.0/v0.2.0 drafts treated `CLAUDE.local.md` §1-§27 as one always-loaded body. That is no
longer the shape of the file. Since v0.2.0 the maintainer guide was consolidated: §18-§27 were
externalized into `.moai/docs/*.md`, and the file now runs §1-§17 plus a `## References` table that
maps each retired §-number to the file that received its content. `wc -l CLAUDE.local.md` → `511`.

Two consequences bind this SPEC's requirements:

- **The always-loaded surface is now `CLAUDE.local.md` (§1-§17 + References) and
  `internal/config/CLAUDE.md`.** The externalized `.moai/docs/*.md` files are load-on-demand
  pointers, read when a task reaches for them. Drift there is still a defect — the always-loaded
  file cites them as authoritative — but it is a *reference* defect, not an *input* defect, and
  requirements re-anchored into `.moai/docs/` carry that reduced severity explicitly (REQ-UDD-004,
  REQ-UDD-007).
- **§-number citations elsewhere in the repo still resolve.** 44 Go comments plus 8 rule/doc sites
  cite `CLAUDE.local.md §18`-`§27`. They are *not* dangling: the `## References` table preserves the
  §-number → file mapping (`§22 Dev Settings Intent → .moai/docs/local-dev-settings-intent.md`), and
  the receiving files retain their original §-numbered headings (`local-dev-settings-intent.md:58`
  is still headed `### §22.8 …`). This was checked before it was assumed, precisely so a
  repo-wide "stale citation" sweep is not proposed on a false premise. See §C.

### A.1 `CLAUDE.local.md` §2.2 was corrected in place, and the correction introduced one new error (F1, narrowed)

At `CLAUDE.local.md:146` (§2.2 — the section moved from `:141` during the consolidation) the four
false claims the v0.1.0 draft found are **gone**. A dated in-place correction was added on
2026-08-01:

```
$ grep -c '2026-08-01 정정' CLAUDE.local.md
1
```

It states that the loader exists (`internal/config/loader_gate.go`, called from
`internal/config/loader.go` `Loader.Load`) and that the real compiled default is
`AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}`. Both re-verified:

```
$ grep -rn 'loadGateSection' internal/config/ | grep -v '^.*://'
internal/config/loader_gate.go:20:func (l *Loader) loadGateSection(dir string, cfg *Config) {
internal/config/loader.go:89:	l.loadGateSection(sectionsDir, cfg)
internal/config/slice.go:35:	"gate":          (*Loader).loadGateSection,

$ grep -n -A5 'AstGrepGate: AstGrepGateConfig' internal/config/defaults.go
438:		AstGrepGate: AstGrepGateConfig{
439-			Enabled:      true,
440-			BlockOnError: false,
441-			WarnOnlyMode: true,
442-		},
```

(The `defaults.go` site moved from `:316-322` to `:438-442`; the values are unchanged.)

**The two false-claim requirements are therefore discharged** — REQ-UDD-002's "shall not assert the
absence of `loadGateSection` / of a shipped `gate.yaml` / a compiled default of `false`" clause, and
REQ-UDD-003's "shall not state that impact requires an explicit `moai ast-grep` invocation with `sg`
installed" clause. Neither claim survives (`grep -oE 'sg 설치|moai ast-grep' <the section>` → no
match).

**What survives is a new misstatement of the same class.** The correction concludes:

> 실제 기본값은 … `AstGrepGate{Enabled: true, BlockOnError: false, WarnOnlyMode: true}` — 즉
> **차단 없는 권고 모드로 켜져 있고**, 차단(blocking)만이 `gate.yaml` opt-in이다.

"Blocking is opt-in only" is false. `RunAstGrepGateV2` runs two steps, and step 1's block is
unconditional on `WarnOnlyMode`:

```
$ sed -n '41,62p' internal/hook/quality/astgrep_gate.go
41:func RunAstGrepGateV2(ctx context.Context, projectDir string, cfg *AstGrepGateConfig) (bool, string) {
42:	if cfg == nil || !cfg.Enabled {
43:		return true, ""
44:	}
46:	// ── 1. Suppression policy check (sg-independent, pure-Go) ─────────────────
48:	sourceFiles := walkSourceFiles(projectDir)
53:	if len(allViolations) > 0 {
60:		return false, strings.TrimSpace(sb.String())
61:	}
63:	// ── 2. ast-grep scan (depends on sg CLI) ─────────────────────────────────
```

`WarnOnlyMode` is passed only into the step-2 scanner config (`:66-71`); the step-1 `return false` at
`:60` is reached whenever `cfg.Enabled` is true and any suppression-policy violation is found — which
is the shipped default. So a maintainer whose commit is refused by the gate reads §2.2, concludes
blocking requires an opt-in they never made, and looks elsewhere. This is the same failure the
original §2.2 caused, one correction later.

**The scope qualifier is also still absent.** The gate is not evaluated on every tool call:

```
$ grep -n -B2 -A2 'IsGitCommit(command)' internal/hook/pre_tool.go
447:		if quality.IsGitCommit(command) && !config.IsAutonomyTierCommitGateOff(config.AutonomyTier()) {
448:			gate := quality.NewQualityGate(h.loadGateConfig())
```

(The site moved from `:430-431` to `:447-448`, and gained a second conjunct — an autonomy-tier
opt-out — that did not exist at v0.2.0.) §2.2 states no invocation scope at all
(`grep -cE 'git commit|IsGitCommit'` over the section → `0`), so a reader has no way to bound the
blast radius of a default-on gate.

REQ-UDD-002 and REQ-UDD-003 are therefore **kept live and narrowed** to these two residuals.

### A.2 `CLAUDE.local.md` §22.8 moved to `.moai/docs/`, and its claim inverted under it (F2, re-anchored)

§22.8 no longer exists in `CLAUDE.local.md`: `grep -n '§22' CLAUDE.local.md` returns one line, `:506`,
which is the References entry pointing at `.moai/docs/local-dev-settings-intent.md`. The content is
there, still under its original heading at `:58`.

The relocated text was itself updated to state per-toggle reader status — which is what REQ-UDD-004
asked for — and cites `SPEC-CONFIG-KEY-HONESTY-001` M5 as the source, discharging REQ-UDD-005's
reconciliation obligation. `local-dev-settings-intent.md:62-63`:

> - `auto_create`: 프로덕션 리더가 **있다** — `internal/cli/worktree_advisory.go::readWorktreeAutoCreate`가
>   읽는다. 단, 이 리더는 두 `fmt.Fprintln` 문구 중 하나를 고르는 용도일 뿐 …
> - `auto_merge` / `auto_cleanup`: 프로덕션 리더가 **없다** (declared but not read). 어떤 코드 경로도
>   이 값을 소비하지 않는다.

**The `auto_cleanup` half is now false, and false in the direction that matters.** Two production
readers exist:

```
$ grep -rn 'Worktree\.AutoCleanup\|Worktree\.AutoMerge' --include='*.go' internal cmd pkg | grep -v '_test.go'
internal/cli/session_worktree.go:584:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
internal/cli/session_worktree_prmerge.go:122:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
```

Neither is advisory wording. `session_worktree.go:584` sits inside `cleanupSessionWorktree`, and the
read gates disposal outright — `false` returns early and the worktree persists, `true` disposes it.
`auto_cleanup` is now a behaviour-gating key. `auto_merge` remains unread
(`grep -rn 'Worktree\.AutoMerge' … | grep -v '_test.go'` → no output, exit `1`), so the sentence is
half true, which is the worst available state: a reader who checks one key and finds the doc right
extends that confidence to the other.

**The polarity of the defect has inverted.** At v0.2.0 the doc claimed an enforcement that did not
exist; today it claims an absence that does not hold. The requirement's force is unchanged — the
documented reader status must be the measured one — but its target text and its expected correction
are the opposite of what v0.2.0 specified, which is why the requirement is re-stated rather than
carried.

**Severity, stated honestly.** This text is no longer always-loaded (§A.0). It is read when a task
reaches for the settings-intent doctrine — which is exactly when the reader is deciding whether a
`false` default is intentional. Lower exposure, undiminished consequence at the moment of use.

An adjacent observation, recorded and deliberately not made a requirement: both reader sites carry a
comment citing `CLAUDE.local.md §22.8`. Those citations resolve through the References table (§A.0),
so they are stale-looking but navigable, and rewriting 44+ such Go comments is a separate scope (§C).

### A.3 The nonexistent `config.yaml` survives at three sites, and `.agency/` joins it (F3, kept live)

`.moai/config/config.yaml` still does not exist, at either path, and is still named as the main
configuration file:

```
$ ls .moai/config/config.yaml internal/template/templates/.moai/config/config.yaml
ls: .moai/config/config.yaml: No such file or directory
ls: internal/template/templates/.moai/config/config.yaml: No such file or directory

$ grep -n 'config/config.yaml' CLAUDE.local.md
328:**Project config** (`.moai/config/config.yaml`):

$ grep -nE 'config\.yaml.{0,2} \(main\)|Main .?config\.yaml' internal/config/CLAUDE.md
5:…the layered configuration tree under `.moai/config/` — `config.yaml` (main) plus `sections/*.yaml`…
11:- **Section-file layout (CLAUDE.local.md §9)**: … Main `config.yaml` aggregates references. …
```

Three sites, down from four: the §5 release-checklist site left `CLAUDE.local.md` with the
consolidation and is now REQ-UDD-007's re-anchored target (§A.4). The remaining `CLAUDE.local.md:328`
sits under §9 and calls the nonexistent file "Main configuration file" — with the actual
`sections/*.yaml` layout described immediately below it, so the file contradicts itself within four
lines.

**A second instance of the same requirement, not previously named.** `CLAUDE.local.md` instructs, in
a `[HARD]` Template-First rule, that new files under `.agency/` be added to the template tree first:

```
$ grep -n '\.agency' CLAUDE.local.md
88:internal/template/templates/.agency/
94:When adding new files to `.claude/`, `.moai/`, or `.agency/`:
106:- New agency files (`.agency/`)
108:**Verification**: Before committing, check that every new file under `.claude/`, `.moai/`, or `.agency/` has a corresponding file in `internal/template/templates/`.

$ ls -d .agency internal/template/templates/.agency
ls: .agency: No such file or directory
ls: internal/template/templates/.agency: No such file or directory
```

`:88` names a template-source directory that does not exist, and `:94`/`:106`/`:108` instruct an
agent to mirror files into it — an unperformable instruction inside a `[HARD]` rule, which is the
strongest form of the defect this SPEC exists to remove.

The direction of the correction is measured, not assumed. `.agency/` is alive in the code, but
**inbound only**: it is a legacy layout that `moai migrate agency` reads *out of* a user's project
(`internal/cli/migrate_agency.go:200`, `internal/cli/v2_detection.go:282`,
`internal/cli/update_residue_cleanup.go:84`, `internal/defs/dirs.go:134`). Nothing ships into it.
So the Template-First rule's `.agency/` arm is not merely pointing at a missing directory — it names
a direction that never existed.

### A.4 The §5 release checklist moved to `.moai/docs/version-management.md`, and its unperformable line moved with it (F3b, re-anchored)

`CLAUDE.local.md` §5 is now a three-line pointer; the *Files Requiring Version Sync* list lives at
`.moai/docs/version-management.md:66-78`. The unperformable entry survived the move verbatim:

```
$ sed -n '76,78p' .moai/docs/version-management.md
**Configuration Files:**
- .moai/config/sections/system.yaml (moai.version)
- internal/template/templates/.moai/config/config.yaml (moai.version)
```

`:77` is correct — `.moai/config/sections/system.yaml` exists and carries `version: v3.1.0-rc.2`.
`:78` names the file §A.3 just showed does not exist, so the checklist line is unperformable: a
releaser either skips it silently or creates a config file nothing loads.

**The replacement was measured rather than guessed** (plan.md AP-6). There is no
`internal/template/templates/.moai/config/sections/system.yaml` either — the template ships
`system.yaml.tmpl`, whose version field is a render-time substitution:

```
$ grep -n -A2 '^moai:' internal/template/templates/.moai/config/sections/system.yaml.tmpl
4:moai:
5-  # MoAI-ADK version
6-  version: "{{.Version}}"
```

So the template side carries no hand-bumped version at all: it is injected when `moai init` /
`moai update` renders the template. The correct edit is **deletion of `:78`**, with a one-line note
recording that the template side is render-time-injected and needs no manual bump — not substitution
of another path, which would re-create the defect in a new location.

### A.5 `--dry-run` reached the plan renderer — the sibling SPEC implemented option B (F5, retired)

REQ-UDD-011 and REQ-UDD-012 are **satisfied**, and were satisfied by
`SPEC-UPDATE-REINSTALL-LOOP-002` (status `completed`) rather than by this SPEC. The v2 fingerprint is
now computed *inside* the dry-run branch:

```
$ sed -n '344,355p' internal/cli/update.go
344:		// SPEC-UPDATE-REINSTALL-LOOP-002 REQ-RIL2-024/025 (M4): the v2
345:		// fingerprint is computed HERE — inside the dry-run branch, above the
346:		// deny-rule migration below — so the clean-reinstall and
347:		// residue-cleanup plans become reachable from `moai update --dry-run`.
353:		// The early return itself does NOT move (REQ-RIL2-026): it stays
354:		// ABOVE stripRetiredV2DenyEntries, which rewrites settings.json.
355:		return emitDryRunReinstallPlan(cmd.Context(), cwd, getBoolFlag(cmd, "force"), out, th)
```

`emitDryRunReinstallPlan` (`internal/cli/update.go:546-600`) renders the clean-reinstall plan for a
v2 fingerprint and the residue-cleanup plan otherwise, and its doc comment records that it writes
nothing. The help text keeps both halves of its promise
(`internal/cli/update.go:81` — "Show planned archive and install operations without modifying the
filesystem"), and the promise is now met. A reachability test exists
(`internal/cli/update_dry_run_reach_test.go:186` `TestUpdateDryRun_EmitsCleanReinstallPlan`).

This is option B, implemented under exactly the constraint this SPEC and its sibling both imposed:
the early return did not move past `stripRetiredV2DenyEntries`. The v0.2.0 concern that the two
SPECs might push the return in opposite directions did not materialise. REQ-UDD-013, whose only
content was routing an escalation to that sibling, is moot.

### A.6 `internal/config/CLAUDE.md` still contradicts `CLAUDE.local.md` §9 (F4, kept live)

This is the one finding that has not moved at all. `internal/config/CLAUDE.md:12` still states:

> **Configuration priority order**: (1) Environment variables (`MOAI_USER_NAME`,
> `MOAI_CONVERSATION_LANG`, ...) override file values. … Tests MUST verify this priority via
> `t.Setenv` + fixture file combinations.

and `:13` still gives `EnvUserName = "MOAI_USER_NAME"` as the `envkeys.go` worked example. Neither
name is implemented:

```
$ grep -c 'MOAI_USER_NAME\|MOAI_CONVERSATION_LANG' internal/config/envkeys.go
0                                              # exit 1

$ grep -n -A13 'func applyEnvOverrides' internal/config/manager.go
398:func applyEnvOverrides(cfg *Config) {
399-	if mode := os.Getenv(EnvDevelopmentMode); mode != "" { … }
402-	if level := os.Getenv(EnvLogLevel); level != "" { … }
405-	if format := os.Getenv(EnvLogFormat); format != "" { … }
408-	if noColor := os.Getenv(EnvNoColor); noColor == "true" || noColor == "1" { … }
411-}
```

(`applyEnvOverrides` moved from `:393-406` to `:398-411`; the four overrides are unchanged.)
`CLAUDE.local.md` §9 states the correct fact, so two always-loaded files contradict each other and
whichever a reader weights wins silently. The contradiction is resolved against the code, not by
recency.

The second-order defect is unchanged and is the sharpest thing in this SPEC: `:12` does not merely
misstate, it *instructs* — "Tests MUST verify this priority via `t.Setenv`". For an unimplemented
variable that is an instruction to write a test that cannot pass, or to implement the override so it
can, which is an unrequested behaviour change originating in a documentation error.

### A.7 Duplicate ownership with `SPEC-INTERNAL-ARCH-001` REQ-ARCH-006 — resolved in favour of this SPEC

`SPEC-INTERNAL-ARCH-001` (status `draft`, Tier L, P1) carries `REQ-ARCH-006` at `spec.md:82`, which
claims the same fix: document the implemented override set, and stop citing `MOAI_USER_NAME` /
`MOAI_CONVERSATION_LANG` or an `EnvUserName` constant. Its `AC-ARCH-007a` is a single grep over all
three tokens.

**Resolution: this SPEC owns the fix (REQ-UDD-008, REQ-UDD-009, REQ-UDD-010).
`SPEC-INTERNAL-ARCH-001` REQ-ARCH-006 is superseded and should be reduced to a cross-reference.**
Three reasons, in order of weight:

1. **Landing latency.** `SPEC-INTERNAL-ARCH-001` is Tier L with `depends_on: [SPEC-INTERNAL-TEST-001]`
   and a `REQ-ARCH-007` cross-cutting behaviour-preservation invariant binding every milestone. Its
   M6 is a one-commit doc fix, but it can only land inside that SPEC's lifecycle, behind a large
   behaviour-preserving refactor. This SPEC has no `depends_on` and is doc-only. A two-line
   correction to an always-loaded file should not wait on a refactor Epic.
2. **Coverage.** `AC-ARCH-007a` is an absence-grep over three tokens plus a prose clause. It does not
   cover REQ-UDD-010 — the `t.Setenv` instruction scoping — at all, and it has no positive
   replacement assertion, so it is satisfiable by deleting the bullet. Assigning ownership to the
   coarser requirement would silently drop the second-order defect (§A.6), which is the part of this
   finding that can cause an unrequested code change.
3. **Subject fit.** An instruction file directing an agent toward an unpassable test is an
   always-loaded-instruction hazard, which is this SPEC's declared subject; it is not an architecture
   concern.

**Left undone deliberately, and named so it is not lost:** the reciprocal edit to
`SPEC-INTERNAL-ARCH-001` — marking REQ-ARCH-006 superseded and pointing AC-ARCH-007 here — is *not*
made by this rewrite. That SPEC is under concurrent review in this session, and editing another
agent's artifact would race it. The owner of `SPEC-INTERNAL-ARCH-001` must apply the cross-reference;
until they do, the duplicate is resolved on one side only. AC-UDD-024 is the guard that detects the
still-unresolved half rather than assuming it away.

### A.8 Drift recorded while rewriting

Four discrepancies between the v0.2.0 artifacts and the tree measured at `7f61332ef`.

**Drift 1 — the v0.2.0 baseline is not an ancestor of this tree.**
`git merge-base --is-ancestor d5336214e HEAD` exits `1`. Every `file:line`, count, and verbatim
output in v0.2.0 was therefore unattributable here and was re-observed rather than carried. Line
numbers moved in five cited files (`CLAUDE.local.md` §2.2 `:141`→`:146`; `defaults.go`
`:316-322`→`:438-442`; `pre_tool.go` `:430-431`→`:447-448`; `manager.go` `:393-406`→`:398-411`;
`update.go` help text `:69`→`:81`).

**Drift 2 — the `pre_tool.go` gate condition gained a conjunct.** `:447` now reads
`if quality.IsGitCommit(command) && !config.IsAutonomyTierCommitGateOff(config.AutonomyTier())`. The
autonomy-tier opt-out did not exist at v0.2.0. §A.1's scope statement is written against the current
two-conjunct form; a correction to §2.2 that names only `git commit` would be incomplete on this tree.

**Drift 3 — `auto_cleanup` gained two production readers, inverting F2 (§A.2).** The v0.2.0 baseline
recorded `grep -rn 'Worktree\.AutoCleanup\|Worktree\.AutoMerge' … | wc -l` → `0`; measured here it is
`2`. This reverses the direction of REQ-UDD-004's correction and is the single most consequential
finding of the rewrite: had REQ-UDD-004 been executed as written in v0.2.0, it would have written a
false statement into the doc.

**Drift 4 — several v0.2.0 acceptance criteria are now unsatisfiable as written, independent of the
requirement they serve.** AC-UDD-014 expected `grep -c '로더 부재' CLAUDE.local.md` → `0`; measured
it is `1`, because the 2026-08-01 correction **quotes the retracted claim** in order to retract it
("종전 이 절은 … 라고 적었으나 두 주장 모두 현재 main에서 거짓이다"). An absence-grep cannot
distinguish an assertion from its quoted retraction. This is a general hazard for
documentation-correctness ACs and is now excluded by construction in `acceptance.md` §A clause 8.

## §B Requirements (GEARS)

Retired requirements keep their IDs and are recorded in place with the evidence that discharged them;
they are never renumbered or silently deleted.

### B.1 Always-loaded instruction correctness

**REQ-UDD-001** — **Where** an always-loaded instruction file (`CLAUDE.local.md`,
`internal/config/CLAUDE.md`) states a mechanism, the stated mechanism shall match the code at the
cited site. **When** a reader follows the stated mechanism to the code, the code shall exhibit the
stated behaviour. **Where** an always-loaded file delegates a topic to a `.moai/docs/*.md` file via
the `## References` table, that delegated file shall be held to the same correctness standard at the
point of use.

**REQ-UDD-002** *(narrowed at v0.3.0)* — `CLAUDE.local.md` §2.2 shall name the ast-grep gate's actual
invocation scope: evaluated on `git commit` Bash invocations, and additionally suppressed when the
autonomy tier disables the commit gate (`internal/hook/pre_tool.go:447`,
`quality.IsGitCommit` && `!config.IsAutonomyTierCommitGateOff`). **Where** §2.2 states the gate's
default, it shall not do so without the scope qualifier, so the correction does not imply the gate
fires on every tool call.

> *Discharged at v0.3.0*: the clause forbidding assertion of an absent `loadGateSection`, an absent
> shipped `gate.yaml`, or a compiled default of `false`. All three claims were retracted by the
> 2026-08-01 in-place correction (§A.1). AC-UDD-015 is retained as the regression guard.

**REQ-UDD-003** *(narrowed and sharpened at v0.3.0)* — §2.2 shall state that the suppression-policy
check (`internal/hook/quality/astgrep_gate.go` step 1) is sg-independent pure Go and returns a
blocking result **irrespective of `WarnOnlyMode`**, and shall not state or imply that blocking is
reachable only via a `gate.yaml` opt-in. **Where** §2.2 describes the sg-dependent scan (step 2), it
shall state that it degrades gracefully when `sg` is absent.

> *Discharged at v0.3.0*: the clause forbidding the claim that impact requires an explicit
> `moai ast-grep` invocation with `sg` installed. That claim was retracted (§A.1).
> *Added at v0.3.0*: the `WarnOnlyMode`-irrespective half, which the 2026-08-01 correction newly
> contradicts.

### B.2 Documented policy versus code reachability

**REQ-UDD-004** *(re-anchored and polarity-inverted at v0.3.0)* — **Where**
`.moai/docs/local-dev-settings-intent.md` §22.8 states the production-reader status of the
`workflow.worktree.*` toggles, that status shall be the measured one: `auto_cleanup` **is** read, at
`internal/cli/session_worktree.go:584` and `internal/cli/session_worktree_prmerge.go:122`, and the
read gates worktree disposal rather than selecting advisory wording; `auto_merge` has no production
reader; `auto_create` is read at `internal/cli/worktree_advisory.go:60` and selects advisory wording
only. The section shall not assert that no code path consumes `auto_cleanup`.

**REQ-UDD-005** — **RETIRED at v0.3.0.** The reconciliation with `SPEC-CONFIG-KEY-HONESTY-001` §A.6
that this requirement made an obligation has been performed: that SPEC reached `status: completed`
(v0.3.0, 2026-08-12), and `.moai/docs/local-dev-settings-intent.md:60,65` cites its M5 as the source
of the recorded toggle status and of the template/`defaults.go` alignment. The obligation is
discharged; the residual factual error in the reconciled text is REQ-UDD-004's, not this
requirement's. AC-UDD-006 is retired with it.

### B.3 References to nonexistent paths

**REQ-UDD-006** *(targets enumerated at v0.3.0)* — An always-loaded instruction file shall not name a
path that does not exist, and shall not instruct a reader to write to one. Three concrete targets:

1. `CLAUDE.local.md:328` (§9) shall stop describing `.moai/config/config.yaml` as the project's main
   configuration file, and shall describe the actual `sections/*.yaml`-only layout.
2. `internal/config/CLAUDE.md:5` and `:11` shall stop asserting an aggregating main `config.yaml`.
3. `CLAUDE.local.md:88`, `:94`, `:106`, `:108` shall stop naming `internal/template/templates/.agency/`
   as a template-source directory and shall stop instructing that new `.agency/` files be mirrored
   into it. **Where** `.agency/` is retained in the document at all, it shall be described as what
   the code implements — a legacy layout read *out of* a user project by `moai migrate agency` and
   the v2 fingerprint detector — not as a template-managed output directory.

**REQ-UDD-007** *(re-anchored at v0.3.0)* — **Where** `.moai/docs/version-management.md` §*Files
Requiring Version Sync* lists files requiring a version bump at release, the list shall contain only
files that exist, so every checklist line is performable. The
`internal/template/templates/.moai/config/config.yaml (moai.version)` entry at `:78` shall be
**removed**, and shall not be replaced by another template path: the template side carries no
hand-bumped version, because `internal/template/templates/.moai/config/sections/system.yaml.tmpl:6`
renders `version: "{{.Version}}"` at deploy time. **Where** the removal would leave a reader assuming
the template version is unmanaged, a one-line note shall record the render-time injection.

### B.4 Contradiction between two always-loaded files

**REQ-UDD-008** — Two always-loaded instruction files shall not assert contradictory facts about the
same mechanism. `internal/config/CLAUDE.md:12`'s configuration-priority statement shall be corrected
to match the implemented override set — `MOAI_DEVELOPMENT_MODE`, `MOAI_LOG_LEVEL`, `MOAI_LOG_FORMAT`,
`MOAI_NO_COLOR` (per `applyEnvOverrides`, `internal/config/manager.go:398-411`), plus
`MOAI_CONFIG_DIR` for the config-directory location — and shall stop citing `MOAI_USER_NAME` /
`MOAI_CONVERSATION_LANG` as implemented overrides.

**REQ-UDD-009** — **Where** `internal/config/CLAUDE.md:13` illustrates the `envkeys.go` constant
convention, the example shall name a constant that `envkeys.go` actually declares. The convention
itself — constants live in `envkeys.go`, no inline `os.Getenv("MOAI_*")` — is correct and shall be
preserved; only the worked example changes.

**REQ-UDD-010** — An instruction file shall not require a test whose subject is unimplemented.
`internal/config/CLAUDE.md:12`'s "Tests MUST verify this priority via `t.Setenv` + fixture file
combinations" shall be scoped to the implemented override set, so following the instruction literally
produces neither an unpassable test nor an unrequested behaviour change.

**REQ-UDD-014** *(new at v0.3.0)* — **Where** this SPEC takes ownership of a fix another SPEC also
claims, the superseded requirement shall be reduced to a cross-reference by that SPEC's owner, so the
duplicate is not resolved on one side only. `SPEC-INTERNAL-ARCH-001` REQ-ARCH-006 / AC-ARCH-007 is
the instance (§A.7). This SPEC shall not edit that SPEC; it shall detect and report the unresolved
state.

### B.5 Flag contract agreement

**REQ-UDD-011** — **RETIRED at v0.3.0, satisfied.** The `--dry-run` help text and behaviour agree.
`internal/cli/update.go:81` promises "planned archive and install operations"; the dry-run branch at
`:332-355` computes the v2 fingerprint in place and returns `emitDryRunReinstallPlan` (`:546-600`),
which renders the clean-reinstall or residue-cleanup plan without mutating. Implemented by
`SPEC-UPDATE-REINSTALL-LOOP-002` REQ-RIL2-024/025 (§A.5). AC-UDD-001 is retained as a regression
guard; AC-UDD-002's bespoke tests are retired in favour of the sibling's
`TestUpdateDryRun_EmitsCleanReinstallPlan`.

**REQ-UDD-012** — **RETIRED at v0.3.0, satisfied.** The A-vs-B choice was recorded as an explicit
decision at v0.2.0 (option B) and executed under the stated constraint: the early return did not
move past `stripRetiredV2DenyEntries`, as `internal/cli/update.go:353-354` records verbatim
(REQ-RIL2-026). The demonstration obligation this requirement attached to option B is discharged by
the sibling's own no-mutation guarantee, documented at `internal/cli/update.go:551-555`.

**REQ-UDD-013** — **RETIRED at v0.3.0, moot.** Its content was to route a REQ-UDD-011 escalation to
`SPEC-UPDATE-REINSTALL-LOOP-002` if the plan could not be rendered without mutation. That SPEC
reached `status: completed` having rendered the plan without mutation, so the escalation has no
subject.

### B.6 Non-functional

**NFR-UDD-001** — Every test added by this SPEC shall confine its filesystem writes to `t.TempDir()`.

**NFR-UDD-002** *(extended at v0.3.0)* — No file under `internal/template/templates/**` shall be
modified by this SPEC. `CLAUDE.local.md`, `internal/config/CLAUDE.md`, and the two re-anchored
`.moai/docs/` targets are repo-local maintainer documentation and are never mirrored — verified:
`ls internal/template/templates/.moai/docs/` lists exactly `agent-lint.md` and
`generic-patterns-guide.md`, and neither `version-management.md` nor `local-dev-settings-intent.md`
is present.

**NFR-UDD-003** — Go sources added by this SPEC shall use `snake_case.go` filenames, wrap errors with
`fmt.Errorf("...: %w", err)`, and carry English comments and godoc.

**NFR-UDD-004** — Each documentation correction shall cite the code site it was verified against
(`file:line` or a content-anchored symbol name), so a future reader can re-verify without
re-deriving the measurement.

## §C Exclusions

### Out of Scope — the worktree-key triage

- Deciding whether `workflow.worktree.auto_cleanup` / `auto_merge` / `auto_create` should be
  implemented, deprecated, marked reserved, or removed; adding or removing readers for them; and any
  change to `internal/config/defaults.go` or `internal/config/types.go` on their account. Owned by
  `SPEC-CONFIG-KEY-HONESTY-001` (now `completed`). REQ-UDD-004 corrects only the recorded reader
  status.
- In particular: the fact that `auto_cleanup` acquired behaviour-gating readers while the shipped
  default remains `false` may be a config-honesty question. This SPEC records the reader status and
  raises no position on it.

### Out of Scope — clean-reinstall and `--dry-run` behaviour

- Whether the clean-reinstall loop should trigger, its v2-fingerprint detection, its sequencing, and
  the `--dry-run` reachability implementation. Owned by `SPEC-UPDATE-REINSTALL-LOOP-002`, which
  landed it (§A.5). This SPEC retains only regression guards.

### Out of Scope — repo-wide `§`-citation rewriting

- Rewriting the 44 Go comments and 8 rule/doc sites that cite the retired `CLAUDE.local.md`
  `§18`-`§27` numbers. Measured before being proposed: those citations still resolve, because the
  `## References` table preserves the §-number → file mapping and the receiving `.moai/docs/*.md`
  files retain their original §-numbered headings. They are stale-looking, not dangling, so no defect
  is claimed and no sweep is specified here.

### Out of Scope — editing `SPEC-INTERNAL-ARCH-001`

- Marking `REQ-ARCH-006` superseded, or amending `AC-ARCH-007`, inside `SPEC-INTERNAL-ARCH-001`.
  That SPEC is under concurrent review; its owner applies the reciprocal cross-reference.
  REQ-UDD-014 and AC-UDD-024 detect and report the unresolved half rather than editing across the
  boundary.

### Out of Scope — CI guard pattern design

- Extending, generalising, or restructuring the leak-detection or neutrality patterns, coverage
  gating, and the `paths-filter` gating set. Owned by `SPEC-UPDATE-CI-GUARD-001`. This SPEC adds no
  CI workflow change and no guard-pattern change.

### Out of Scope — template mirroring of maintainer documentation

- Mirroring `CLAUDE.local.md`, `internal/config/CLAUDE.md`, `.moai/docs/version-management.md`, or
  `.moai/docs/local-dev-settings-intent.md` into `internal/template/templates/**`. All four are
  repo-local; a mirror would violate the template internal-content isolation doctrine (NFR-UDD-002).

### Out of Scope — implementing the two unimplemented env vars

- Adding `MOAI_USER_NAME` / `MOAI_CONVERSATION_LANG` overrides to `envkeys.go` or
  `applyEnvOverrides`. REQ-UDD-008/009/010 correct the documentation to match the code; the reverse
  direction, if ever wanted, is a separate SPEC.

### Out of Scope — a general documentation audit

- Auditing `CLAUDE.local.md` sections other than §2 (Template-First / §2.2), §9, and the References
  table, `.moai/docs/` files other than the two re-anchored targets, or `internal/config/CLAUDE.md`
  statements other than the `config.yaml`, configuration-priority, and `envkeys.go`-convention
  bullets. This SPEC fixes the measured drifts; a general sweep is a separate scope.
