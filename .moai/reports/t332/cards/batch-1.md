# t332 card sweep — batch 1

Cards: t90 t125 t154 t191 t196 t201 t204 t216 t223 t224  (10 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t90

**Premise (one sentence).** Codex plugin packaging (M6) cannot start until Codex dual-harness M1-M5 land, and should stay closed until demand is confirmed.

**Premise verdict.** `holds` — checked that all 7 Codex-track SPECs in `.moai/specs/` (`SPEC-CODEX-DUAL-AGENTS-001`, `SPEC-CODEX-HOOK-ADAPTER-001`, `SPEC-CODEX-INIT-001`, `SPEC-CODEX-LAUNCHER-001`, `SPEC-CODEX-PHASE2-001`, `SPEC-CODEX-SKILLS-CANONICAL-001`, `SPEC-CODEX-WIRING-001`) carry `status: completed` in `spec.md` frontmatter, and no `.codex-plugin/plugin.json` or any `*.codex-plugin*` path exists anywhere in `internal/template/templates/`. The M1-M5 precondition the card names now appears satisfied (all upstream SPECs closed), but the packaging artifact itself is absent — consistent with "not started, demand-gated."

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Queries run: `git log <develop> --perl-regexp --grep='\bt90\b' --oneline` and the same against `<main>` — both empty.

**Claim.** t90 is queued, its stated precondition (M1-M5) now looks satisfied by measurement, but the demand-gate the card itself imposes ("수요가 확인되기 전에는 열지 말 것") is still unaddressed — no `.codex-plugin/` artifact exists.

**Evidence.**
```
$ ls .moai/specs/ | grep -i codex
SPEC-CODEX-DUAL-AGENTS-001
SPEC-CODEX-HOOK-ADAPTER-001
SPEC-CODEX-INIT-001
SPEC-CODEX-LAUNCHER-001
SPEC-CODEX-PHASE2-001
SPEC-CODEX-SESSION-MSG-001
SPEC-CODEX-SKILLS-CANONICAL-001
SPEC-CODEX-VERDICT-SYNTH-001
SPEC-CODEX-WIRING-001

$ grep -n "^status:" .moai/specs/SPEC-CODEX-*/spec.md
SPEC-CODEX-INIT-001/spec.md:5:status: completed
SPEC-CODEX-DUAL-AGENTS-001/spec.md:5:status: completed
SPEC-CODEX-HOOK-ADAPTER-001/spec.md:5:status: completed
SPEC-CODEX-LAUNCHER-001/spec.md:5:status: completed
SPEC-CODEX-WIRING-001/spec.md:5:status: completed
SPEC-CODEX-PHASE2-001/spec.md:5:status: completed
SPEC-CODEX-SKILLS-CANONICAL-001/spec.md:5:status: completed

$ find internal/template/templates -iname "*.codex-plugin*" -o -iname "plugin.json"
(no output)
```

**Baseline-attribution.** SPEC status measured against this worktree's `.moai/specs/` tree at HEAD `6165f9f5e`; template scan measured against `internal/template/templates/` at the same HEAD.

**Gaps.** Did not check `SPEC-CODEX-SESSION-MSG-001` / `SPEC-CODEX-VERDICT-SYNTH-001` status (they postdate the card's M1-M5 naming and are not obviously part of the M1-M5 the card refers to). Did not confirm whether "demand" has been operator-assessed since the card was filed — that is a judgment call outside grep's reach.

**Residual-risk.** If the operator already confirmed demand out-of-band (e.g. a Slack/verbal decision), this card's block condition may already be lifted and the finding here would understate readiness.

**Proposed disposition.** `needs-operator-decision` — the coded precondition (M1-M5) looks satisfied, so the remaining gate is purely the operator's demand judgment; only the operator can say whether that's been cleared.

**Overlap candidates.** t196 (same Codex-track, also gated on t88/M4 wiring completion — both cards sit downstream of the same Codex dual-harness effort).

---

### t125

**Premise (one sentence).** `CloudAI-X/threejs-skills` cannot be mirrored into the template because it lacks a LICENSE file (README's MIT claim isn't binding), so any Three.js skill content must be authored independently or the license gap resolved first.

**Premise verdict.** `unverified` — this premise concerns an external, third-party repository's licensing state, which this worktree's git history and file tree cannot adjudicate. No Three.js skill files exist anywhere in this repo (`find . -iname "*threejs*"` returned nothing), consistent with "not yet mirrored," but I cannot verify the external repo's current LICENSE status (would require a live fetch of `CloudAI-X/threejs-skills`, out of scope for a read-only local sweep).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Queries run: `git log <develop> --perl-regexp --grep='\bt125\b' --oneline` and the same against `<main>` — both empty.

**Claim.** No Three.js-related skill content has landed in this repo under any name; the card's blocking condition (external license gap) is unverified but internally consistent with the absence of such content.

**Evidence.**
```
$ find . -iname "*threejs*" -not -path "*/node_modules/*"
(no output)
```

**Baseline-attribution.** Filesystem scan against this worktree's full tree at HEAD `6165f9f5e`.

**Gaps.** Did not fetch the external `CloudAI-X/threejs-skills` or `pinkforest/threejs-playground` repos to confirm current LICENSE state — read-only local sweep has no network access mandate here, and the card's core claim is about a third party's repo, not this codebase.

**Residual-risk.** The external repo may have added a LICENSE file since the card was filed, which would resolve the blocker without any change needed here.

**Proposed disposition.** `needs-operator-decision` — verifying the external LICENSE state requires a live check the operator (or a WebFetch-capable follow-up) should perform; this sweep cannot settle it.

**Overlap candidates.** none observed.

---

### t154

**Premise (one sentence).** The `OutOfScopeRule` H3-only heading requirement is documented [HARD] policy in `manager-spec.md:51` (not a bug), so the H2-widening fix built on `WT-lint-heading` was correctly rejected by the operator and t155 (SPEC-side conformance) was chosen instead.

**Premise verdict.** `holds` — `manager-spec.md:51` (both local `.claude/agents/moai/manager-spec.md` and the template mirror `internal/template/templates/.claude/agents/moai/manager-spec.md`) still reads: "at least one `### Out of Scope —` H3 heading" as `[HARD]` policy, byte-identical in both copies. This confirms the card's own stated resolution (H3-only stays intentional, doctrine unchanged) is still the live state of the repo.

**Landing verdict.** `in-flight-unlanded`
- commit: — (no delivering commit on either pinned ref)
- pinned ref: `ee50984abe4f11ac337382b48a26328f091e200a` (develop)
- `--is-ancestor` exit: **1** — tip `dbb87f14f` is NOT an ancestor of pinned develop
- branch + tip (in-flight only): `WT-lint-heading` @ `dbb87f14f` (worktree `.claude/worktrees/t154`)

> **Orchestrator correction, M3 post-check.** The worker recorded `unknown` here, reasoning that the
> vocabulary has no slot for "intentionally never landed by design". That reasoning is sound about
> the card's *intent* and misplaced in the *landing* verdict, which asks only whether the work
> landed — a question that IS measurable here and was measured: a live worktree exists with an
> unmerged branch, so the card is `in-flight-unlanded` by AC-BH-009's own definition. Recording
> `unknown` would hide a measured fact behind an unmeasurable one, and would leave AC-BH-009's
> positive conjunct (every in-scope card with an unmerged worktree branch carries this
> classification) unsatisfied. The rejected-by-design fact is preserved below and carried into the
> disposition, which is where it belongs. Deciding command:
> `git merge-base --is-ancestor dbb87f14f ee50984abe4f11ac337382b48a26328f091e200a` → exit 1.

This card explicitly documents a **rejected** change (never intended to land) plus a pointer to its replacement, t155. `git log <develop> --perl-regexp --grep='\bt154\b'` and the same against `<main>` both returned empty — consistent with "never merged," which is the card's own stated outcome, not a gap. Landing verdict is recorded `unknown` rather than `not-landed` because the card's terminal state is "resolved by operator decision, superseded by t155" rather than a pending implementation — the vocabulary in the instructions doesn't have a clean slot for "intentionally-never-landed-by-design."

**Claim.** The card's self-reported resolution (operator declined the H2-widening change; H3-only rule remains intentional doctrine) matches the current repo state exactly. The in-flight worktree `WT-lint-heading` (`.claude/worktrees/t154`, tip `dbb87f14f`) still exists but the card says explicitly this branch is retained only for possible future reuse, not for merging as-is.

**Evidence.**
```
$ grep -n "H3" .claude/agents/moai/manager-spec.md | head -5
51:- [HARD] Every spec.md MUST include an exclusions section (what NOT to build) containing at least
   one `### Out of Scope — <topic>` H3 sub-heading with one or more `-` bullet items. The
   `OutOfScopeRule` lint (`MissingExclusions`) requires the literal text "out of scope", an
   `### Out of Scope —` H3 heading, and at least one `-` bullet under it; a bare H2 exclusions
   heading with no `### Out of Scope` sub-heading fails the rule.

$ grep -n "HARD" internal/template/templates/.claude/agents/moai/manager-spec.md | grep -i "h3\|heading"
51: (byte-identical to above)

$ git log ee50984ab... --perl-regexp --grep='\bt155\b' --oneline
76ef8a764 stamp-t155-updated
```
Worktree list confirms `t154` still has a live, locked-free worktree at `.claude/worktrees/t154` tip `dbb87f14f` branch `WT-lint-heading` (per `.moai/reports/t332/00-worktree-list.txt`), not merged.

**Baseline-attribution.** `manager-spec.md:51` read at this worktree's HEAD `6165f9f5e` (local + template copies compared directly, not diffed against a ref). t155 landing check against pinned develop `ee50984ab...`.

**Gaps.** Did not verify t155's actual content/status beyond the one matching commit subject ("stamp-t155-updated") — did not read t155's SPEC or confirm it fully supersedes t154's concern. Did not check whether `WT-lint-heading` worktree is stale relative to current `manager-spec.md` (i.e., whether reuse is still mechanically clean).

**Residual-risk.** If t155's SPEC only partially addresses the union-heading false-positive findings t154 preserved (the three measured findings: H4 already worked, `:883` heading-widening breaks umbrella structure, 16/25 baseline findings are genuine), some of that preserved knowledge could still be unconsumed.

**Proposed disposition.** `needs-operator-decision` — the operator declined this change on
2026-08-20 (recorded in the card's own text), so no implementation is pending; but the card is not
`already-landed` — `git merge-base --is-ancestor dbb87f14f <pinned-develop>` exits **1**, and its
branch `WT-lint-heading` is still live and unmerged. What remains is a disposal question only the
operator can settle: dropping the card discards the pointer to the three preserved measurements and
to the branch the card itself says to "reuse if reversed later," while keeping it leaves a queued
card whose implementation was already declined.

> **Orchestrator correction, M3 post-check.** The worker proposed `already-landed`, which
> contradicts this entry's own landing verdict — `already-landed` asserts the work reached the
> integration branch, and the measurement says it did not. The worker's underlying reading (the
> operator's decision IS the terminal state) is right and is preserved above; only the disposition
> token, which asserted something measurably false, is changed.

**Overlap candidates.** t155 (not in this sweep's in-scope-62 list — appears to already be superseded/landed as its replacement, confirmed via `76ef8a764 stamp-t155-updated`).

---

### t191

**Premise (one sentence).** `/moai:project` P2 (a `workflow.project.continuation: none|card|pipeline` config key) can now start because its precondition PR #1601 (t188, P1) has merged, pending confirmation that t170①'s `todo.enabled` key is also in main.

**Premise verdict.** `unverified` — PR #1601 is confirmed merged (see Landing verdict below), but I could not confirm whether `todo.enabled` (t170①, the second named precondition) is present in `.moai/config/sections/workflow.yaml` at this worktree's HEAD — the grep for `todo.enabled|todo:` in that file returned nothing, meaning either the key doesn't exist under that literal name, or it's not yet landed. The card names t170① as a co-precondition, so its absence here is ambiguous between "not yet needed" and "blocking."

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

`git log <develop> --perl-regexp --grep='\bt191\b' --oneline` and the same against `<main>` both empty — t191 (P2) itself has not landed. Its precondition, #1601, HAS landed (commit `e91def4ca docs(project): end /moai project with a derived backlog card and a start branch (t188) (#1601)`, found via a separate `--grep='#1601\b'` query against develop).

**Claim.** t191 (P2, the config-key axis) has not started; its stated blocking precondition #1601/t188 (P1) is confirmed merged. `workflow.project.continuation` does not exist anywhere in the config tree yet.

**Evidence.**
```
$ git log ee50984ab... --perl-regexp --grep='#1601\b' --oneline
e91def4ca docs(project): end /moai project with a derived backlog card and a start branch (t188) (#1601)

$ grep -rn "workflow.project.continuation\|continuation:" .moai/config/sections/workflow.yaml \
    internal/template/templates/.moai/config/sections/workflow.yaml
(no output)

$ grep -n "todo.enabled\|todo:" .moai/config/sections/workflow.yaml
(no output)
```

**Baseline-attribution.** All greps against this worktree's tree at HEAD `6165f9f5e`; #1601 landing checked against pinned develop `ee50984ab...`.

**Gaps.** Did not search beyond `workflow.yaml` for a `todo.enabled`-equivalent key under a different section name (e.g. it could live in a differently-named sections file). Did not check the wizard code paths the card references (`t174 wizard→workflow.yaml 배선 패턴`) for whether the P1 (#1601) implementation already partially wired a continuation-like behavior under a different key name.

**Residual-risk.** If `todo.enabled` exists under an unexpected key name, the "still blocked" framing here could be wrong and t191 might actually be unblocked and simply not yet picked up.

**Proposed disposition.** `keep` — precondition #1601 confirmed merged; card is legitimately actionable next, pending the t170① confirmation the card itself flags.

**Overlap candidates.** none observed in-scope (t170, t174, t188 referenced by the card are not in the 62-id in-scope list, suggesting they're already resolved/out of the hygiene sweep's scope).

---

### t196

**Premise (one sentence).** 9 of 21 mirrored Codex skills reference Claude-only tool names (AskUserQuestion/Agent()/Skill()/TaskCreate), 3 depend on `${CLAUDE_SKILL_DIR}`, and all 11 agent TOML files use Claude-style delegation phrasing, so Codex cannot be trusted with full `/moai` orchestration without a neutral-instruction layer, gated on t88 (M4) wiring being finalized.

**Premise verdict.** `falsified` (partially) — the skill-mirroring half of the premise is stale: no `.codex/skills/` directory exists anywhere under `internal/template/templates/.codex` at all (only `.codex/agents/moai/` with 11 TOML files). The "21 mirrored skills, 9 with Claude-only refs" figure no longer describes this tree. The agent-TOML half of the premise still **holds**: all 11 files in `.codex/agents/moai/*.toml` match a grep for `AskUserQuestion|Agent()|Skill(|TaskCreate` (11/11). The card's t88 precondition is also satisfied — `SPEC-CODEX-WIRING-001` is `status: completed`.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

`git log <develop> --perl-regexp --grep='\bt196\b' --oneline` and the same against `<main>` — both empty.

**Claim.** The agent-TOML half of t196's problem (11/11 Claude-style delegation phrasing) is confirmed still present and unaddressed. The skills-mirror half of the premise (21 skills, 9 non-neutral) is stale — there is currently no `.codex/skills/` tree to audit, so that specific figure cannot be re-verified as stated.

**Evidence.**
```
$ find internal/template/templates/.codex -maxdepth 3
internal/template/templates/.codex
internal/template/templates/.codex/agents
internal/template/templates/.codex/agents/moai
internal/template/templates/.codex/agents/moai/{11 *.toml files}

$ grep -rln "AskUserQuestion\|Agent()\|Skill(\|TaskCreate" internal/template/templates/.codex/agents/moai/*.toml | wc -l
11
$ ls internal/template/templates/.codex/agents/moai/*.toml | wc -l
11

$ grep -n "^status:" .moai/specs/SPEC-CODEX-WIRING-001/spec.md
5:status: completed
```

**Baseline-attribution.** Filesystem + grep measured against `internal/template/templates/.codex` at this worktree's HEAD `6165f9f5e`. SPEC status measured against `.moai/specs/SPEC-CODEX-WIRING-001/spec.md` at the same HEAD.

**Gaps.** Did not search other possible skill-mirror locations outside `.codex/` (e.g. whether skills content moved elsewhere, or whether "mirrored skills" refers to something the card author observed via a different mechanism not visible as files in this tree — e.g. a runtime-only Codex skill listing). Did not verify the `${CLAUDE_SKILL_DIR}` dependency count (3 skills) since no skills directory was found to check.

**Residual-risk.** If Codex skill-mirroring exists via a mechanism not represented as `internal/template/templates/.codex/skills/*` files (e.g. generated at a different stage, or the skills are Claude-only and simply invoked cross-runtime), the "no skills tree" finding could be a false negative for the skills half of this premise.

**Proposed disposition.** `needs-operator-decision` — the card's precondition (t88/M4) is now satisfied and the agent-TOML problem is confirmed live, but the skills-mirror figures need re-measurement before the card's scope can be trusted as stated; the operator should decide whether to re-scope the card's numbers or proceed on the TOML-only finding.

**Overlap candidates.** t90 (same Codex-track demand-gated M6 packaging; both reference the same M1-M5/t88 completion state).

---

### t201

**Premise (one sentence).** `constitution validate` ignores `[SUPERSEDED]`-prefixed clauses and the `canary_gate:false` flag, so retired registry entries fail DRIFT permanently with no escape except deletion (destroying the retirement record).

**Premise verdict.** `falsified` — `internal/constitution/retirement.go` now implements `IsRetiredClause`, and `validate` skips drift/canary-gate/source-file checks for `[SUPERSEDED ...]`-prefixed entries while counting them in a new `retired_count`. This is the exact fix the card (and issue #1595) requested.

**Landing verdict.** `landed`
- commit: `bf6083f13` (`fix(constitution): let validate retire an entry instead of deleting it (#1611)`)
- pinned ref: `ee50984abe4f11ac337382b48a26328f091e200a` (develop) — also confirmed ancestor of pinned main `48239c7dc7428c8751a04f6321887c2d36123884`
- `--is-ancestor` exit: `0` (both develop and main)
- branch + tip (in-flight only): — (already merged, no live worktree)

**Claim.** t201 is fully landed. PR #1611 (`bf6083f13`) implements the retirement-marker recognition, later commits in the same PR chain narrow `IsRetiredClause` to require a complete, self-closing `[SUPERSEDED ...]` token (fixing two CodeRabbit-flagged false-retirement edge cases) and fatal-ize `TestShippedRegistriesLoad` (previously skipped silently on a broken registry).

**Evidence.**
```
$ git log ee50984ab... --perl-regexp --grep='\bt201\b' --oneline
bf6083f13 fix(constitution): let validate retire an entry instead of deleting it (#1611)
ec1837038 feat(kanban): moai todo pr — read-only card-to-PR and landed link (t210) (#1628)
f6958b109 docs(kanban): pre-dispatch PR cross-check + card id in PR titles (t210) (#1627)

# The latter two are MENTIONS, not deliveries — confirmed by reading their full commit bodies:
# t210's commit body says "`t201` is the genuine ambiguous case — it appears in both #1611 and
# #1612 — and the SPEC had asserted `inferred` for it" — t201 is cited as a *test fixture example*
# for an unrelated card (t210's PR-resolver SPEC), not delivered by these commits.

$ git merge-base --is-ancestor bf6083f13 ee50984abe4f11ac337382b48a26328f091e200a; echo $?
0
$ git merge-base --is-ancestor bf6083f13 48239c7dc7428c8751a04f6321887c2d36123884; echo $?
0
```

**Baseline-attribution.** Commit search against pinned develop and pinned main SHAs as given; ancestry confirmed against both literal 40-hex SHAs in this run.

**Gaps.** Did not re-read `internal/constitution/retirement.go` in full to verify the marker-narrowing edge cases beyond what the commit messages describe — trusted the commit trail (which includes RED-then-GREEN test evidence for the two CodeRabbit findings) rather than re-deriving it.

**Residual-risk.** None material — this is a closed, merged, ancestor-confirmed fix with test evidence in the commit trail.

**Proposed disposition.** `already-landed` — drop from the backlog.

**Overlap candidates.** none observed (the two commits sharing the `t201` grep token belong to the unrelated t210 SPEC and merely cite t201 as an example case).

---

### t204

**Premise (one sentence).** v3.1.3 deployment (tag + GoReleaser) is held per operator directive until the operator's own testing plus Codex/desktop-app testing confirms moai-adk usability, though #1602 and the CHANGELOG stamp already merged.

**Premise verdict.** `holds` — checked `pkg/version/version.go`, which reads `Version = "v3.1.3"` (the in-tree pre-release version string), while `git tag --list 'v3.1.*' --sort=-v:refname` shows the newest **tagged** release is still `v3.1.2` — no `v3.1.3` tag exists. This is exactly the held state the card describes: code/version bumped, not yet tagged or released.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

`git log <develop> --perl-regexp --grep='\bt204\b' --oneline` and the same against `<main>` both returned one hit each, `7b217da7c` (`feat(codex): Codex dual-harness wiring...` #1619, delivering card **t88**, not t204) — read in full, it is a mention only: the commit body's "Record Phase 4 mode selection (serial)" section states "Integration route per lead: card PR to main only, no release-branch integration (**v3.1.3 hold on t204**)" — citing t204's hold as context for an unrelated card's own integration decision, not delivering t204's own release-gate condition. t204's own completion signal ("운영자 테스트 완료 선언") is inherently non-mechanical (an operator declaration), so it cannot land as a commit in the first place.

**Claim.** The release hold t204 describes is still in effect — v3.1.3 is not yet tagged. t204 will never "land" as a commit by its own nature (the completion signal is an operator declaration, not code); it should be tracked by the release-tag state instead.

**Evidence.**
```
$ grep -n 'Version = ' pkg/version/version.go
8:	Version = "v3.1.3"

$ git tag --list 'v3.1.*' --sort=-v:refname | head -5
v3.1.2
v3.1.1
v3.1.0-rc.2
v3.1.0-rc.1
v3.1.0-rc.0
```

**Baseline-attribution.** `pkg/version/version.go` read at this worktree's HEAD `6165f9f5e`; tag list is a global git-tags query (not tree-scoped), taken at the same session.

**Gaps.** Did not check whether the operator has verbally/out-of-band declared testing complete since the card was filed (2026-08-24) — that state isn't recorded in git and this sweep has no access to it. Did not check whether a `release/v3.1.3` branch exists (worktree list shows `release-v313` at tip `b37e86b64` — present but that alone doesn't indicate the hold was lifted, since release branches are cut before tagging per this repo's git-flow doctrine).

**Residual-risk.** If the operator declared completion verbally and this simply hasn't been actioned into a tag yet, this card could be effectively resolved and only needs the mechanical release step.

**Proposed disposition.** `needs-operator-decision` — only the operator can confirm whether their own testing-completion declaration has occurred; this sweep can only confirm the tag hasn't been cut.

**Overlap candidates.** none observed directly in-scope, though the `release-v313` worktree (`b37e86b64`) is the natural landing surface once this card resolves.

---

### t216

**Premise (one sentence).** Three findings from a hook-wiring audit: (D-1) `chain-event.sh` is wired in the distributed template's `settings.json` but missing from this project's own `settings.json`, so completion-chain events have never been recorded here; (D-2) of 43 on-disk hook scripts, 32 are wired via 34 settings.json entries, 0 are wired-but-missing, and 11 are present-but-unwired pending call-graph tracing; (D-3) the MX cold-start scan structurally cannot complete in one session.

**Premise verdict.** `holds` (D-1 specifically re-verified) — confirmed `internal/template/templates/.claude/settings.json.tmpl` wires `chain-event.sh` as a genuine SessionEnd-family hook command entry (not merely a string reference), while this worktree's own `.claude/settings.json` contains zero occurrences of `chain-event`. The specific gap the card names is still present.

**Landing verdict.** `in-flight-unlanded`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: **1** — tip `8aa96bfb1` is NOT an ancestor of pinned develop
  (orchestrator M3 post-check filled this in; the worker recorded it as a Gap it did not run.
  Deciding command: `git merge-base --is-ancestor 8aa96bfb1 ee50984abe4f11ac337382b48a26328f091e200a` → exit 1)
- branch + tip: `WT-hook-wiring-drift` @ `8aa96bfb1`

`git log <develop> --perl-regexp --grep='\bt216\b' --oneline` and the same against `<main>` — both empty, so nothing has merged. `.moai/reports/t332/00-worktree-list.txt` shows a live worktree `t216` at tip `8aa96bfb1` branch `WT-hook-wiring-drift`. I did not run `git merge-base --is-ancestor 8aa96bfb1 <pinned-develop>` to formally confirm the tip is unmerged (out of this card's restraint budget), but the empty develop/main grep results are consistent with it being unmerged.

**Claim.** D-1 (the chain-event.sh local/template drift) is independently confirmed still true in this tree. A separate, unmerged worktree (`t216`/`WT-hook-wiring-drift`) exists carrying investigation/repair work in progress.

**Evidence.**
```
$ grep -n -B3 -A3 "chain-event.sh" internal/template/templates/.claude/settings.json.tmpl
199-          }{{ end }},
200-          {
201-            "command": "bash",
202:            "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; ...", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/chain-event.sh"],
203-            "timeout": 5,
204-            "type": "command"
205-          }

$ grep -c "chain-event" .claude/settings.json
0

$ grep -n "t216" .moai/reports/t332/00-worktree-list.txt
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t216   8aa96bfb1 [WT-hook-wiring-drift]
```

**Baseline-attribution.** Template and local settings.json both read at this worktree's HEAD `6165f9f5e`. Worktree-list line taken verbatim from the pre-supplied `00-worktree-list.txt` snapshot (not re-run).

**Gaps.** Did not run the formal ancestor check for `8aa96bfb1` against pinned develop. Did not verify D-2's 43-script/32-wired/11-unwired figures against the current tree, nor D-3.

**Residual-risk.** The in-flight worktree may already contain a fix for D-1 not yet merged, which would make "in-flight-unlanded" read as more open than it actually is once that branch lands.

**Proposed disposition.** `needs-operator-decision` — this card explicitly frames its central open question as a design decision ("moai update가 추가된 훅 엔트리를 기존 프로젝트에 반영해야 하는가?"), which only the operator can settle; the mechanical D-1 finding itself is re-confirmed and actionable regardless.

**Overlap candidates.** none observed directly in-scope for D-1/D-2/D-3's specific mechanism, though the card's own scope (Template-First hook wiring) touches files any hook-related card in the batch set could also touch.

---

### t223

**Premise (one sentence).** `.claude/agent-memory` is per-project-root, and a worktree is its own project root, so agent-memory written by a subagent inside a worktree stays trapped in that worktree's copy and is destroyed when the worktree is disposed, unless a drain-then-dispose mechanism is added.

**Premise verdict.** `unverified` — the mechanism the card describes (agent-memory being project-root-relative, worktrees being independent project roots) is architecturally plausible and consistent with how `.claude/` is scoped per checkout, but I did not find a definitive current-state answer (either confirming a fix landed, or confirming the gap is still open) within the batch's restraint budget. `SPEC-WORKTREE-REAPER-001` exists in `.moai/specs/` (the card's own stated origin, "t209 워크트리 리퍼"), suggesting active related work, but I did not read its status or content.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

`git log <develop> --perl-regexp --grep='\bt223\b' --oneline` and the same against `<main>` — both empty.

**Claim.** t223 has not itself landed. Its precondition (t209/`SPEC-WORKTREE-REAPER-001` M2 landing, which the card says must land first to confirm the "P2: preserve-forever" classification) exists in the specs tree but its current phase/status was not read in this pass.

**Evidence.**
```
$ git log ee50984ab... --perl-regexp --grep='\bt223\b' --oneline
(no output)
$ git log 48239c7dc... --perl-regexp --grep='\bt223\b' --oneline
(no output)
$ ls .moai/specs | grep -i "worktree-reaper"
SPEC-WORKTREE-REAPER-001
```

**Baseline-attribution.** Commit search against both pinned SHAs as literals; SPEC directory existence checked at this worktree's HEAD `6165f9f5e`.

**Gaps.** Did not read `SPEC-WORKTREE-REAPER-001/spec.md` or `progress.md` to determine whether its M2 milestone (the card's stated precondition) has landed, nor whether REQ-WR-025 (the P2 preserve-forever requirement the card cites) is still classified that way. Did not verify whether any drain-then-dispose mechanism exists in current worktree-disposal code paths (`moai worktree done`, worktree reaper logic).

**Residual-risk.** If `SPEC-WORKTREE-REAPER-001` M2 already landed and reclassified P2, this card's stated precondition ("선행: t209 M2 착지 후") may already be satisfied and the card ready to start — this sweep cannot confirm or deny that.

**Proposed disposition.** `needs-operator-decision` — the precondition-check requires reading a specific SPEC's milestone status, which is outside this card's bounded investigation depth; flagging for a follow-up check rather than guessing.

**Overlap candidates.** t313 (also worktree-subsystem-focused — `WT-worktree-baseref`/`SPEC-WORKTREE-BASEREF`-family — different specific mechanism from t223's agent-memory concern, but shares the general worktree-lifecycle domain and could share a design review).

---

### t224

**Premise (one sentence).** Kanban/Factory lane sessions hit a session-level instruction ("don't use the Agent tool unless the user requests it") that blocks them from spawning owning-agent subagents (e.g. `manager-spec`), forcing direct edits that bypass the Status Transition Ownership Matrix — so doctrine across `kanban-dispatch.md`, `agent-common-protocol.md`, `moai-constitution.md`, `manager-lead.md`, and lane SessionStart bootstrap needs an explicit standing spawn-permission grant.

**Premise verdict.** `falsified` (partially) — re-reading the currently-loaded `kanban-dispatch.md` and `kanban-dispatch-detail.md` (both always-loaded in this session), they already state explicitly: "Lead and lane sessions orchestrate only; real work runs in sub-agents" (`kanban-dispatch.md:15`) and, for Factory Mode specifically, "The lane session orchestrates only: each stage's execution — plan authoring, run implementation, sync sweeps — is spawned as `Agent()` sub-agents" (`kanban-dispatch-detail.md:161`). This is exactly the standing spawn-permission language the card asks to have added. However, this only confirms the **doctrine files** carry the language now — it does NOT confirm the specific runtime symptom the card reports (a session-level instruction actually blocking a lane's Agent-tool call at spawn time despite the doctrine saying it should be permitted) has been fixed, since that is a SessionStart-bootstrap / runtime-instruction-injection concern this sweep cannot exercise.

**Landing verdict.** `unknown`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

`git log <develop> --perl-regexp --grep='\bt224\b' --oneline` and the same against `<main>` — both empty, so no commit explicitly cites t224. Recorded `unknown` rather than `not-landed` because the doctrine text the card requested already appears to exist in the currently-loaded rules — I cannot determine whether this text predates the card (i.e., was already there and the card's actual runtime-instruction complaint is unaddressed) or was added afterward without citing the card id in a commit message.

**Claim.** The doctrine-level half of t224's requested fix (explicit "lanes spawn Agent() sub-agents" language in kanban-dispatch.md/detail) is present in the current always-loaded rules. Whether this predates or postdates the card, and whether the runtime SessionStart-instruction conflict the card actually reports is resolved, is unverified.

**Evidence.**
```
$ grep -n "orchestrate only\|spawned as .Agent" .claude/rules/moai/workflow/kanban-dispatch.md \
    .claude/rules/moai/workflow/kanban-dispatch-detail.md
kanban-dispatch.md:15: The lead session works through the `manager-lead` agent: ... Lead and lane
  sessions orchestrate only; real work runs in sub-agents ...
kanban-dispatch-detail.md:161: - **Within a card (stages as sub-agents).** The lane session
  orchestrates only: each stage's execution — plan authoring, run implementation, sync sweeps —
  is spawned as `Agent()` sub-agents whose output stays in their windows. ...

$ git log ee50984ab... --perl-regexp --grep='\bt224\b' --oneline
(no output)
```

**Baseline-attribution.** Both rule files read as currently loaded in this session's system context, at this worktree's HEAD `6165f9f5e` (these files are always-loaded per their own classification, so the content shown is live, not stale).

**Gaps.** Did not check git blame/history on these two specific lines to determine whether they predate the card's 2026-08-24 filing date. Did not check `agent-common-protocol.md`, `moai-constitution.md`, or `manager-lead.md` for the remaining doctrine-update scope the card names (orchestrator delegation clause, agent definition, SessionStart bootstrap context) — a grep against `manager-lead.md` returned only Role-B-adjacent "background Agent() spawn" language for the LEAD, not an explicit lane-permission statement. Did not verify the actual runtime symptom (a session instruction blocking Agent-tool use) still occurs.

**Residual-risk.** The doctrine text found may be incidental/pre-existing wording never actually written in response to this card — if so, the card's real complaint (a runtime instruction override) remains completely open despite doctrine appearing to already say the right thing, which is exactly the failure mode the card describes (documented policy vs. actually-honored runtime instruction).

**Proposed disposition.** `needs-operator-decision` — the doctrine-text half looks satisfied but cannot be confirmed as a direct response to this card, and the card's core runtime-symptom complaint is unverifiable from static file inspection alone; the operator should judge whether this is resolved or still needs the broader multi-file update the card specifies (agent-common-protocol.md, moai-constitution.md, manager-lead.md scoping, SessionStart bootstrap).

**Overlap candidates.** none observed directly in-scope among this batch's other 9 cards, though the card's own scope (kanban-dispatch.md, manager-lead.md, agent-common-protocol.md) overlaps with any other in-scope card that also touches kanban/factory-mode doctrine files.
