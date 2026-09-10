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
# t332 card sweep — batch 2

Cards: t231 t233 t236 t237 t239 t240 t242 t243 t244 t247 t248  (11 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t231

**Premise (one sentence).** `worktree clean`'s lock-source-unreadable path gives a machine-readable
`anchored: undetermined` signal in `--json` but no distinguishable exit code on the human path, so
a caller relying on exit code cannot tell a degraded run from a clean success.

**Premise verdict.** `holds` — verified against the tree at HEAD. `internal/cli/worktree/clean.go`
`reportStaleWorktrees` (the `--json` path, around line 372-384) explicitly discards the lock error
(`candidates, _ := classifyStaleWorktrees(...)`) and always `return nil`. The human path
`cleanStaleWorktrees` (line ~213-280) prints a stderr notice via `lockSourceUnreadableNotice` on
`lockErr != nil` but still falls through to `return nil` at every exit point — no error, no
distinguishable code. `internal/cli/worktree/clean_lock_unreadable_test.go`
(`TestCleanStale_UnreadableLockSourceRemovesNothing`) asserts this is intentional:
`t.Fatalf("runClean must stay non-blocking (REQ-WR-016), got error: %v", err)` — the test fails if
the command ever returns a non-nil error on this path. So REQ-WR-016 (the requirement the card asks
to amend) is still in force, unmodified.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree named for t231 in `00-worktree-list.txt`)

**Claim.** Both the `--json` and human `worktree clean` paths return exit 0 on a lock-source-read
failure; only `--json`'s payload field distinguishes the degraded run. Note: the card describes the
human path as exit 1 ("exit code는 1로 뭉개짐"); the code as read returns exit 0 (nil error) on
every branch, which is a *stronger* form of the same defect (no discrimination at all, not merely a
generic 1) — this is a minor factual correction to the card's own diagnosis, not a rebuttal of its
premise.

**Evidence.**
```
$ grep -n "causeLockSourceUnreadable" internal/cli/worktree/clean.go
315: c.KeepReason = fmt.Sprintf("cause=%s; could not read the worktree lock state: %v", causeLockSourceUnreadable, lockErr)
494: // causeLockSourceUnreadable is the cause token for a lock source that could
496: const causeLockSourceUnreadable = "lock-source-unreadable"
502: return fmt.Sprintf("moai: worktree clean degraded (cause=%s; git worktree list --porcelain failed: %v): no worktree removed", causeLockSourceUnreadable, err)

$ sed -n '372,384p' internal/cli/worktree/clean.go
	candidates, _ := classifyStaleWorktrees(worktrees, base)
	...
	return enc.Encode(candidates)   # always nil unless JSON encoding itself fails

$ sed -n '228,280p' internal/cli/worktree/clean.go
	candidates, lockErr := classifyStaleWorktrees(worktrees, base)
	if lockErr != nil {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), lockSourceUnreadableNotice(lockErr))
	}
	...
	return nil   # every terminal branch in this function

$ sed -n '46,58p' internal/cli/worktree/clean_lock_unreadable_test.go
	out, err := runStaleClean(t, map[string]string{"stale": "true", "yes": "true"})
	if err != nil {
		t.Fatalf("runClean must stay non-blocking (REQ-WR-016), got error: %v", err)
	}
```

**Baseline-attribution.** All figures measured against worktree HEAD `6165f9f5e` by direct
`grep`/`sed`/`Read` of `internal/cli/worktree/clean.go` and
`internal/cli/worktree/clean_lock_unreadable_test.go` in this run.

**Gaps.** Did not run `main.go`'s `cli.ResolveExitCode` to trace whether any wrapping layer above
`cleanStaleWorktrees`/`reportStaleWorktrees` could still turn a nil error into something other than
0 (unlikely, since `main.go` only special-cases a non-nil error). Did not check whether a *different*
subcommand path (`moai worktree done`, `recover`) shares this classifier and inherits the gap.

**Residual-risk.** If a future refactor threads `lockErr` through as a returned error instead of a
side-channel notice, the exit code could change without this file's local tests catching a
regression in REQ-WR-016's stated invariant, since the invariant itself is what's being asked to
change.

**Proposed disposition.** `keep` — the premise is current and the requested exit-2 discrimination +
REQ-WR-016 amendment genuinely has not happened; rests on the `clean_lock_unreadable_test.go`
non-blocking assertion above.

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t233

**Premise (one sentence).** `moai gate` silently passes (`exit 0`, no notice) the lint axis for a
non-eslint Node project (biome/oxlint) because the only Node lint step is eslint,
config-file-gated, and a config-gated skip previously produced no notice.

**Premise verdict.** `falsified` — the specific "무통지"(no-notice) half of the premise is
contradicted by the current tree. `internal/hook/quality/gate.go` still lists only one Node lint
step (`eslint`, gated on eslint config files — no biome/oxlint step exists, so that half of the
underlying capability gap is real), **but** `executeStep`'s config-gated skip branch now calls
`g.summary.markSkipped(step.name, fmt.Sprintf(reasonConfigFilesAbsentFmt, ...))`, and
`QualityGate.Run` returns `joinBlocks(out, g.summary.render())` — every run, pass or fail, renders a
per-step outcome/reason block including skipped steps and their reason. This is exactly the "①-④"
class of fix the card proposes under item ④ ("pass 시 1줄 요약 출력"), already implemented. The
card's own cited line numbers (gate.go:158-163/787-789/811-820/343) do not match current content —
line drift consistent with other work (a Node `typecheckStep` addition, referenced in an inline
comment as closing the *type*-check blind spot for biome-style projects) having landed on this file
since the card was written.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The silent-pass shape the card's headline mutant describes (no notice at all) has already
been closed by an unrelated-looking summary/render mechanism; the narrower "biome/oxlint rules
themselves are not run" gap (candidates ②/① in the card) remains open.

**Evidence.**
```
$ grep -n "eslint\|biome\|lintSteps" internal/hook/quality/gate.go | sed -n '1,10p'
197-198: # comment: "Node had only lint and test, and a project whose linter config is
           absent (biome instead of eslint, say) skipped lint too — leaving a
           type-broken build to pass the gate. typecheckStep closes that hole."
204: name: "eslint", binary: "npx", args: []string{"eslint", "."}, optional: true,

$ grep -n "biome\|oxlint\|scripts.lint" internal/hook/quality/gate.go
(no matches)

$ sed -n '891-894p' internal/hook/quality/gate.go
	if len(step.configFiles) > 0 && !g.anyConfigFileExists(step.configFiles) {
		g.summary.markSkipped(step.name, fmt.Sprintf(reasonConfigFilesAbsentFmt, strings.Join(step.configFiles, ", ")))
		return true, ""
	}

$ sed -n '500-505p' internal/hook/quality/gate.go
	return joinBlocks(out, g.summary.render())
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`, file
`internal/hook/quality/gate.go` and `internal/hook/quality/gate_summary.go`, read directly (no
runtime execution of `moai gate` performed).

**Gaps.** Did not actually run `moai gate` against a synthetic biome-only fixture to observe the
literal stdout (static read only, per restraint). Did not check whether the rendered summary block
is suppressed under any output mode (`--json`, quiet flags) that might reproduce the "0 바이트
exit 0" the card's issue reproduction describes.

**Residual-risk.** The card's downstream complaint — `moai gate` still does not actually catch a
biome/oxlint-detected lint error (unused var etc.) because no biome lint step exists — is real and
unaddressed; only the notice/silence aspect is closed. An operator reading only "falsified" without
this residual note could wrongly assume the whole issue is resolved.

**Proposed disposition.** `needs-operator-decision` — the card as currently worded (centered on
silent/무통지 pass) is falsified, but the narrower remaining gap (no biome/oxlint lint step) may
still warrant a rewritten, narrower card. Rests on the `markSkipped`/`render()` evidence above.

**Overlap candidates.** none observed among the in-scope batch ids (card itself references #1639 as
a related-but-distinct issue, not in scope).

---

### t236

**Premise (one sentence).** `MOAI_PROJECT_DIR` goes stale after a worktree switch because
Enter/ExitWorktree do not fire the `CwdChanged` hook event the env var's sole producer depends on,
and `verify_snapshot`/`verify_trend` remain unregistered for the `project_root` parameter that would
let callers route around the stale fallback.

**Premise verdict.** `unverified` — the premise has two halves and only one could be decided, so the
premise **as a whole** is undecided. (Orchestrator normalization, M3 post-check: the worker wrote a
compound verdict here; AC-BH-011 admits exactly one of `holds` / `falsified` / `unverified`, and a
partially-decided premise is the case `unverified` exists for. The substance below is the worker's,
unchanged.) Checked and holding: the
`project_root` parameter registration gap. `internal/cli/mcp_server.go`'s `verify_snapshot` (line
192-198) and `verify_trend` (line 201-206) tool registrations carry no `project_root`
`mcp.WithString` parameter, matching `moai-mcp-tools.md`'s own catalogue of the 9 tools that DO
accept it (verify_snapshot/verify_trend absent from that list). Not checked: whether
`EnterWorktree`/`ExitWorktree` actually fail to emit `CwdChanged` at runtime — that is a live
Claude-Code-runtime behavior claim (the card's own citation is a "reporter's runtime trace... 정적
재현 불가"), not something a static grep of this repository can confirm or refute, since
`CwdChanged` is a host-runtime-emitted event this codebase only *handles*, never emits.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The `project_root`-registration half of the card (3 of the 5 originally-named tools
landed project_root; verify_snapshot/verify_trend still lack it) is confirmed current.

**Evidence.**
```
$ sed -n '191,206p' internal/cli/mcp_server.go
	add("verify_snapshot", mcp.NewTool(
		"verify_snapshot",
		mcp.WithDescription(...),
		mcp.WithString("key", mcp.Required(), ...),
		mcp.WithString("command", ...),
		mcp.WithInteger("exit_code", ...),
	), handleVerifySnapshot)
	add("verify_trend", mcp.NewTool(
		"verify_trend",
		mcp.WithDescription(...),
		mcp.WithString("key", mcp.Required(), ...),
		mcp.WithReadOnlyHintAnnotation(true),
	), handleVerifyTrend)
	# neither call carries mcp.WithString("project_root", ...)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/mcp_server.go`, and cross-checked against the always-loaded
`.claude/rules/moai/core/moai-mcp-tools.md` catalogue (9-tool `project_root` list, which also
excludes verify_snapshot/verify_trend), consistent between doc and code.

**Gaps.** Did not check `internal/hook/cwd_changed.go` or the Claude Code runtime spec for whether
`EnterWorktree`/`ExitWorktree` are documented to emit `CwdChanged` — this is outside what a static
repo read can settle, and the card's own evidence is a runtime trace, not a code citation.

**Residual-risk.** If `CwdChanged` actually does fire on Enter/ExitWorktree (contra the card's
runtime trace), the card's headline defect (stale `MOAI_PROJECT_DIR`) may already be resolved and
only the narrower `project_root`-registration gap remains live.

**Proposed disposition.** `needs-operator-decision` — the checkable half holds; the runtime half
cannot be settled from this tree and would need a live reproduction (enter/exit a worktree, inspect
whether `CwdChanged` fired) rather than static reading.

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t237

**Premise (one sentence).** `moai gate`'s pre-commit hook's `go vet` invocation resolves Go
packages relative to the repository root rather than each staged file's owning module, so any
monorepo with a non-root `go.mod` (e.g. `apps/id`) has every Go commit blocked by a package-resolution
failure rather than a real vet finding.

**Premise verdict.** `holds` — the exact defect shape the card names is present verbatim in the
current tree.

**Landing verdict.** `not-landed`
- commit: — (no commit's message matches `\bt237\b`; the only hits are t230's SPEC narrative
  *mentioning* t237/#1641 as "the card about to change the hook body", which is a reference, not a
  delivery — per the worker instructions' explicit "mention is not a landing" caution)
- pinned ref: 48239c7dc7428c8751a04f6321887c2d36123884 (mention found there and in develop; not a
  delivering commit either way)
- `--is-ancestor` exit: — (not applicable; no delivering commit to test)

**Claim.** `preCommitHookContent`'s vet step still computes `./$(dirname "$f")` against the process
cwd (repo root) with no upward `go.mod` search, exactly the defect the card describes; the
independently-referenced verified patch (`t312-precommit-vet @ b6f478b1a`) is not merged into this
tree.

**Evidence.**
```
$ grep -n "go vet\|dirname\|go.mod" internal/cli/hook_install_precommit.go
77: printf './%s\n' "$(dirname "$f")"
90: if ! go vet $BT_TAGS $PKGS >/dev/null 2>&1; then

$ git log ee50984abe4f11ac337382b48a26328f091e200a --oneline -- internal/cli/hook_install_precommit.go
db1362739 feat(cli): back up and disclose user-modified pre-push hooks (t257) (#1650)
32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)
883d53852 feat(SPEC-PRETOOL-GATE-MOVE-001): relocate commit-quality gate ... (#1189)
a596d9e41 fix(hook): 커밋 게이트 goolm 빌드태그 주입 ...
52b5e4bf5 feat(SPEC-PRECOMMIT-001): 배포 계층 pre-commit 훅 설치기 추가 (fast-subset gofmt+vet)
# no commit touches module-relative package resolution
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/hook_install_precommit.go`, and the pinned-develop file history for that path.

**Gaps.** Did not check the paired template twin
(`internal/template/templates/.git_hooks/pre-commit`) for byte-identity — the card names this as a
hard-required twin edit, but since neither side has changed, drift-checking is moot here. Did not
attempt local reproduction (this repo's own module is root-level, as the card itself notes it cannot
reproduce here).

**Residual-risk.** None beyond what the card already states — the referenced verified patch exists
on an unmerged worktree (`t312-precommit-vet`) and this sweep did not audit that patch's quality,
only confirmed the defect it targets is still live here.

**Proposed disposition.** `keep` — premise holds, not landed, patch reportedly exists but unmerged.

**Overlap candidates.** t239 (explicitly cross-referenced by its own text as "같은 결함형, 다른
표면" — the same silent-overwrite/오버라이트 problem class as this card's monorepo vet-resolution
defect, both rooted in the t230 SPEC-PRECOMMIT-PRESERVE-001 lineage).

---

### t239

**Premise (one sentence).** `moai update` redeploys `.moai/config` wholesale, so a user who hand-edits
`llm.yaml`'s `audit.codex`/`audit.glm` sections (added by t225) has that edit silently reverted on
the next update, the same defect class as t230's pre-commit-hook overwrite but a different surface.

**Premise verdict.** `unverified` — the structural mechanism is confirmed but the specific
`audit.codex`/`audit.glm` sub-key claim could not be located, so the premise **as worded** is
undecided. (Orchestrator normalization, M3 post-check: the worker wrote a compound verdict here;
AC-BH-011 admits exactly one token, and a premise whose named subject cannot be found is undecided
rather than holding. The substance below is the worker's, unchanged.) Confirmed: `internal/cli/update/deploy/deploy.go`'s
`CleanMoaiManagedPaths` deletes `.moai/config/` entirely before redeployment (comment: "Clean
.moai/config/ entirely - backup was already done by the Backup step"), and its own doc comment
states plainly: "Template-managed files are NOT backed up: deployment rewrites them moments later,
so their only copy is never at stake" — which is precisely the false assumption the card flags for a
file like `llm.yaml` whose per-user customization IS the thing at stake. However, grepping the
current template and local `llm.yaml` for an `audit:` key found none — I could not confirm the
specific `audit.codex`/`audit.glm` sub-keys the card describes currently exist under that name (the
audit-model-pin config may have moved to a different section/struct since t225 landed, or my grep
scope was too narrow for the bounded budget here).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The general defect mechanism (wholesale `.moai/config` wipe, template-managed files
excluded from the pre-clean backup on the stated "never at stake" assumption) is confirmed current
and applies to any template-shipped, user-customizable config file, `llm.yaml` included in
principle.

**Evidence.**
```
$ sed -n '99,105p' internal/cli/update/deploy/deploy.go
// destruction it was taken to survive. Template-managed files are NOT backed
// up: deployment rewrites them moments later, so their only copy is never at
// stake, and skipping them keeps the backup to what would otherwise be lost

$ sed -n '187,190p' internal/cli/update/deploy/deploy.go
	# Clean .moai/config/ entirely - backup was already done by the Backup step.

$ grep -n "audit:" internal/template/templates/.moai/config/sections/llm.yaml
(no matches)
$ grep -n "audit:" .moai/config/sections/llm.yaml
(no matches)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/update/deploy/deploy.go`, `internal/template/templates/.moai/config/sections/llm.yaml`,
and this worktree's own `.moai/config/sections/llm.yaml`.

**Gaps.** Did not locate where t225's audit-model-pin values actually live in the current schema
(possibly `internal/config/audit_models.go`'s `AuditModel` struct backed by a different YAML
section/file than `llm.yaml`) — a targeted `Grep` for "audit.codex"/"audit.glm" as literal YAML
paths across `.moai/config/sections/*.yaml` was not run within the bounded per-card budget.

**Residual-risk.** If the audit-pin values in fact live in a config file this run did not check
(e.g., `workflow.yaml`), the specific file name in the card (`llm.yaml`) could be stale even though
the general mechanism is real — the card's proposed disposition would then need to be revised to
name the correct file rather than dropped.

**Proposed disposition.** `keep` — structural mechanism confirmed; recommend the operator re-verify
which file currently carries the audit-pin values before scoping a fix.

**Overlap candidates.** t237 (self-cross-referenced: "t230(pre-commit 무음 덮어쓰기)와 같은
결함형, 다른 표면").

---

### t240

**Premise (one sentence).** The §H overlay doctrine documents that z.ai's `thinking` field carries
reasoning effort for the Anthropic-compatible shim, but live measurement during t225 found the
opposite — top-level `reasoning_effort` is what's actually honored — so the doc needs correcting and
AC-MTP-032b's UNVERIFIED marker needs resolving.

**Premise verdict.** `falsified` — the correction and the marker resolution this card asks for
appear to have already happened, inside SPEC-V3R6-AUDIT-MODEL-PIN-001 (status: `implemented`).
`.moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/acceptance.md` AC-AMP-006 is explicitly headed "`[MUST —
closes AC-MTP-032b]`" and its recorded amendment states: "Delivery was PROVEN by hypothesis B
(top-level `reasoning_effort`) against the hypothesis-A null (1.02)" — i.e., the same reversed
finding the card describes (`thinking` field is the null/ignored one; `reasoning_effort` is the
live one) is already measured and recorded, with the AC that specifically closes the UNVERIFIED
marker the card names.

**Landing verdict.** `not-landed`
- commit: — (no commit message matches `\bt240\b`; this finding is attached to SPEC
  V3R6-AUDIT-MODEL-PIN-001 / card t225, not t240 by commit-message convention)
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The measurement reversal and the AC-MTP-032b closure the card asks for exist in the
SPEC-V3R6-AUDIT-MODEL-PIN-001 acceptance criteria already; whatever remains is narrower than the
card states (I could not locate a "§H 오버레이 문서" file by that heading to confirm or refute
whether ITS specific prose was ever updated — see Gaps).

**Evidence.**
```
$ grep -n "032b\|UNVERIFIED" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/acceptance.md
70: ### AC-AMP-006 — live GLM reasoning-delivery proof, numeric rule (REQ-AMP-006, REQ-AMP-007) [MUST — closes AC-MTP-032b]

$ grep -n "hypothesis" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/acceptance.md
85: delivery from noise (measured null: the hypothesis-A thinking-budget run
100: (output-token ratio 1.40, consistent). Delivery was PROVEN by hypothesis B
101: (top-level `reasoning_effort`) against the hypothesis-A null (1.02) — the

$ grep -n "^status:" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/spec.md
5: status: implemented

$ grep -rln "§H " .claude/
.claude/agents/moai/manager-spec.md
.claude/rules/moai/development/spec-frontmatter-schema.md
# neither file's §H concerns reasoning_effort delivery — the card's "§H 오버레이 문서" target
# was not located in this repo by heading search
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`.moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/{spec,acceptance}.md`, and a repo-wide grep for a "§H"
heading.

**Gaps.** Did not locate the specific "§H 오버레이 문서" the card names (it may live in
`.moai/reports/t225/sync-audit-review-2.md`, the card's own cited source, rather than in the rules
tree — that report file was not read in this run). Did not confirm whether the two "비차단 R1/R2"
items (insertAuditLeaf comment order, default step=2 unreachable) were addressed.

**Residual-risk.** If the actual overlay doctrine file the card means is a separate, still-unedited
document (distinct from the SPEC's own acceptance criteria), the card's core ask — editing THAT
file — could still be outstanding even though the underlying measurement and the AC marker are
resolved.

**Proposed disposition.** `needs-operator-decision` — the measurement/marker half is resolved;
whether a distinct doctrine file still needs a matching edit needs the operator (or a follow-up read
of `.moai/reports/t225/sync-audit-review-2.md`) to confirm.

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t242

**Premise (one sentence).** The origin-trail chain-node-creation path (`CreateNodeAtSpawn`) has no
production caller and `MOAI_CHAIN_NODE_ID` is never set anywhere, so `events.jsonl` has never been
populated and the chain-event hook wiring is a permanent no-op — requiring a decision on whether
this is a bug to fix or an intentionally-incomplete Phase 1 to retire.

**Premise verdict.** `holds` — confirmed by a full-repo grep for both the function and the env-var
constant.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** `CreateNodeAtSpawn` (defined `internal/chain/populate.go:53`) has zero non-test callers
anywhere in the Go tree; `config.EnvChainNodeID` (`MOAI_CHAIN_NODE_ID`) is only ever *read*
(`os.Getenv`) in production code — no production call sets it via `os.Setenv` or otherwise. The gap
is real and total, not partial.

**Evidence.**
```
$ grep -rn "CreateNodeAtSpawn" --include="*.go" . | grep -v _test
internal/chain/populate.go:42: // CreateNodeAtSpawn creates a skeleton node-enter event at a spawn boundary.
internal/chain/populate.go:53: func (p *Populator) CreateNodeAtSpawn(worktreePath, specID, milestone string) (string, error) {
# only the definition itself — no call site

$ grep -rn "EnvChainNodeID" --include="*.go" . | grep -v _test
internal/config/envkeys.go:273: // EnvChainNodeID carries the origin-trail chain node ID from the spawning
internal/config/envkeys.go:279: EnvChainNodeID = "MOAI_CHAIN_NODE_ID"
internal/chain/populate.go:54: parentID := os.Getenv(config.EnvChainNodeID)
internal/chain/populate.go:155: if envID := os.Getenv(config.EnvChainNodeID); envID != "" {
internal/hook/chain_banner.go:78: envNodeID := os.Getenv(config.EnvChainNodeID)
# three reads, zero writes
```

**Baseline-attribution.** Full-repository `grep -rn --include="*.go" .` from worktree root
`6165f9f5e`, both queries.

**Gaps.** Did not read the SPEC-CHAIN-CORE-001 plan/spec to determine whether Phase 1 explicitly
scoped node-creation wiring out (the card frames this exact question as open and unresolved — I did
not attempt to resolve it, per the card's own framing that a human/owner decision is what's needed).

**Residual-risk.** None beyond the open ownership question the card itself names.

**Proposed disposition.** `keep` — premise holds strongly; this is a `needs-operator-decision`-shaped
card by its own design (asks for a judgment, not a fix), so `keep` as-is is the right disposition for
the sweep to propose, letting the operator make the bug-vs-intentional call.

**Overlap candidates.** t216 (source investigation this card was split from), t243, t244 (siblings
split from the same t216 lane-6 investigation, `d1-chain-event.md`/`d2-unwired-scripts.md`).

---

### t243

**Premise (one sentence).** `handle-session-start-navigator.sh` was deleted in a build-recovery
commit alongside sibling hooks, but only the siblings were restored two commits later, leaving this
one hook permanently missing and requiring a decision to restore or retire it.

**Premise verdict.** `falsified` — the file exists, at both the local and template locations, and
git history for this exact path shows exactly ONE commit ever touching it (its creation) — no
deletion, no restoration, contradicting the "deleted then left behind" claim outright.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The file is present now and was never removed in this repository's history; the card's
premise (as measured against a *different* worktree — t216/lane-6, per its own citation) does not
hold against this tree.

**Evidence.**
```
$ find .claude/hooks/moai internal/template/templates/.claude/hooks/moai -iname "*navigator*"
.claude/hooks/moai/handle-session-start-navigator.sh
internal/template/templates/.claude/hooks/moai/handle-session-start-navigator.sh

$ git log --oneline -- .claude/hooks/moai/handle-session-start-navigator.sh
2c87d195f feat(SPEC-PROJECT-NAVIGATOR-001): Project Navigator — living nav + --brief + shared read primitive (#1354)
# single commit — file was never deleted in this repo's history

$ grep -n "navigator" .claude/settings.json
(no matches)
# confirms it is currently unwired (matches the card's OTHER premise — d2-unwired-scripts.md's
# "11 unwired, 4 truly dead" — but not the "was deleted and orphaned" premise)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`, full `git log --oneline` for
the exact path (no `--follow` needed since there is only one entry regardless), and
`.claude/settings.json`.

**Gaps.** Did not check whether the card's cited investigation (`.moai/reports/t216/...
d2-unwired-scripts.md`, on a different worktree at a different point in time) was itself measuring
a genuinely different repository state that has since self-corrected (e.g., another card restoring
the file before this sweep ran) — the git log here shows no restoration commit, which argues against
that, but the investigation report itself was not re-read to reconcile the discrepancy.

**Residual-risk.** The card's *decision* framing ("restore vs retire") is moot if the file was never
actually deleted — but its still-unwired status (confirmed above) is a live, separate finding that a
revised card might still want to carry forward.

**Proposed disposition.** `already-landed` — the specific claimed defect (file missing, sibling
hooks restored, this one orphaned) does not describe the current tree; the file is present. Rests on
the single-commit `git log` result above. The still-open "restore vs retire the WIRING" question
(unwired, not missing) may warrant a differently-worded follow-up card — operator's call.

**Overlap candidates.** t216, t242, t244 (siblings split from the same t216 lane-6 investigation).

---

### t244

**Premise (one sentence).** `team-ac-verify.sh` is currently unwired (no hook registration) and is
one of the dormant hooks the t216 investigation found needing an explicit decision — wire it into
the harness-thorough + team-mode gate it was designed for, or retire it.

**Premise verdict.** `holds` — confirmed both that the file exists (locally and in template) and
that it carries no entry in `.claude/settings.json`'s hook wiring, matching the card's "currently
unwired" claim exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** The dormant/unwired state the card describes is current; no decision (wire or retire) has
been made in this tree.

**Evidence.**
```
$ find .claude/hooks internal/template/templates/.claude/hooks -iname "*team-ac-verify*"
.claude/hooks/moai/team-ac-verify.sh
internal/template/templates/.claude/hooks/moai/team-ac-verify.sh

$ grep -n "team-ac-verify" .claude/settings.json internal/template/templates/.claude/settings.json.tmpl
(no matches in either file)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`, both hook-directory trees and
both settings.json variants (local rendered + template source).

**Gaps.** Did not read the hook's own script body to independently confirm the "harness thorough +
team 전제" preconditions it references (took the card's framing of the trigger conditions at face
value, since the wiring-absence finding alone is sufficient to confirm the "currently unwired"
premise regardless of what its trigger conditions would be if wired).

**Residual-risk.** None beyond the open decision the card itself frames as needed.

**Proposed disposition.** `keep` — premise holds; this is a `needs-operator-decision`-shaped card,
same pattern as t242/t243.

**Overlap candidates.** t216, t242, t243 (siblings split from the same t216 lane-6 investigation;
also referenced together in `agent-common-protocol.md`'s own "Hook Invocation Surface" doctrine as
one of the three mechanically-enforcing hook scripts alongside `sync-phase-quality-gate.sh` and
`status-transition-ownership.sh`, though that doctrine describes it as "dormant" without resolving
wire-vs-retire either).

---

### t247

**Premise (one sentence).** PR #1600 (`feat(mcp): make a running MCP server's build info
observable`, branch `WT-server-version`) is unreviewable at 497 changed files (CodeRabbit skipped
review outright) and `CONFLICTING`, and needs to be split into review-sized, non-generated-vs-logic
separated PRs before it can land.

**Premise verdict.** `falsified` — PR #1600 is not in the state the card describes. It is currently
`MERGED`, with only 10 changed files (597 additions / 7 deletions), on the exact branch name the
card cites (`WT-server-version`), merged 2026-08-25T03:23:18Z via merge commit
`07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e` — nowhere near the 497-file, CONFLICTING state the card
was written against. The underlying work evidently was split/rebased down to a mergeable size before
merging, whether or not that split was driven by this card.

**Landing verdict.** `not-landed`
- commit: — (no commit message in either pinned ref matches `\bt247\b`; the resolution happened via
  the PR itself, not via a card-attributed commit)
- pinned ref: —
- `--is-ancestor` exit: —
- (Supplementary, not a `landed` grep-hit): merge commit `07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e`
  IS an ancestor of pinned main `48239c7dc7428c8751a04f6321887c2d36123884`
  (`git merge-base --is-ancestor` exit 0), confirming the small, merged PR is genuinely in the tree
  this sweep is measuring against — not a stale `gh` cache artifact.

**Claim.** The specific blocking condition the card names (497 files, review-skipped, CONFLICTING)
no longer describes PR #1600's actual state; the card's request has been satisfied by however this
PR ended up at 10 files.

**Evidence.**
```
$ gh pr view 1600 --json state,mergeable,additions,deletions,changedFiles,title
{"additions":597,"changedFiles":10,"deletions":7,"mergeable":"UNKNOWN","state":"MERGED","title":"feat(mcp): make a running MCP server's build version visible"}

$ gh pr view 1600 --json headRefName,mergedAt,mergeCommit
{"headRefName":"WT-server-version","mergedAt":"2026-08-25T03:23:18Z","mergeCommit":{"oid":"07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e"}}

$ git merge-base --is-ancestor 07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e 48239c7dc7428c8751a04f6321887c2d36123884 && echo YES_ANCESTOR || echo NOT_ANCESTOR
YES_ANCESTOR
```

**Baseline-attribution.** `gh pr view` measured live against GitHub at the time of this run (not a
tree-scoped measurement); the ancestry check measured against worktree HEAD `6165f9f5e`'s view of
pinned main `48239c7dc7428c8751a04f6321887c2d36123884`.

**Gaps.** Did not verify whether the 497-files/CONFLICTING state the card describes ever actually
existed for PR #1600 at some earlier point (i.e., whether the card's own dated measurement,
"리드 2026-08-24", was accurate at the time) — only that it does not describe the PR's current,
merged state. Did not check whether the "생성물 재생성 커밋과 로직 변경을 분리" quality bar the card
sets was actually met by however this PR was structured before merging (10 files could still mix
generated + hand-written changes).

**Residual-risk.** If the 10-file merged PR did NOT actually separate generated-artifact commits
from logic commits (the card's specific quality bar, and its named "mutant" to watch for), the
underlying review-quality complaint could still be valid even though the file-count/CONFLICTING
crisis is resolved.

**Proposed disposition.** `already-landed` — rests on the `gh pr view` state above (MERGED, 10
files, ancestor of pinned main).

**Overlap candidates.** none observed among the in-scope batch ids.

---

### t248

**Premise (one sentence).** MCP `audit_multi`/`codex_audit`/`glm_audit` judgment output does not
record which commit of the `moai` binary actually served the audit, so a stale-binary judgment (as
happened in the t229 investigation, 259 commits behind) is indistinguishable after the fact from a
current one.

**Premise verdict.** `holds`, on a bounded check. Grepped the primary audit-multi handler file for
any version/commit-recording field in the tool's registration or its handler and found none.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —

**Claim.** No commit-SHA (or build-version) field was found being attached to the `audit_multi`
tool's output in the handler file that composes it.

**Evidence.**
```
$ grep -n "pkg/version\|BuildVersion\|version\.\|Commit\b" internal/cli/mcp_audit_multi.go
(no matches)
```

**Baseline-attribution.** Measured against worktree HEAD `6165f9f5e`,
`internal/cli/mcp_audit_multi.go` only.

**Gaps.** Did not check `codex_audit`'s and `glm_audit`'s own separate handler files (only
`audit_multi`'s composing file was grepped, per the bounded-depth restraint) — the card names all
three tools, and this run verified only one of the three surfaces. Did not check whether the
persisted audit report files themselves (under `.moai/reports/`) carry a commit field written by a
layer outside the MCP handler (e.g., the calling agent stamping it separately) — that would
partially satisfy the card's intent through a different mechanism than the one grepped here.

**Residual-risk.** If `codex_audit`/`glm_audit` or the persisted-report layer already records a
commit SHA by some other path not covered by this grep, the premise as measured here would overstate
the gap; this run's evidence supports the `audit_multi` composing-handler surface specifically, not
an exhaustive claim across all three tools and the persistence layer.

**Proposed disposition.** `needs-operator-decision` — narrow single-file check supports `keep`, but
the gaps above (2 of 3 named tools + the persistence layer unchecked) mean a fuller read is needed
before committing to scope.

**Overlap candidates.** none observed among the in-scope batch ids (card's own text names t246 as
a related card on "감사가 다른 트리를 읽는 축", not in this batch's in-scope list).
# t332 card sweep — batch 3

Cards: t252 t253 t254 t255 t260 t262 t263 t264 t280 t281 t284  (11 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t252

**Premise (one sentence).** At least one SPEC (SPEC-V3R6-AUDIT-MODEL-PIN-001, t225) is stuck in `implemented` status with a `pending-backfill-sync` §E.4 `sync_commit_sha`, waiting to be batched into a single frontmatter-transition PR once ≥2 such items accumulate.

**Premise verdict.** `holds` — the cited SPEC is still exactly in the state the card describes.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t252 in `00-worktree-list.txt`)

**Claim.** The card's target SPEC (SPEC-V3R6-AUDIT-MODEL-PIN-001) still carries `status: implemented` (not `completed`) and `sync_commit_sha: "pending-backfill-sync"` at worktree HEAD; the batch-backfill action the card describes has not been performed for this item.

**Evidence.**
```
$ sed -n '1,7p' .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/spec.md
---
id: SPEC-V3R6-AUDIT-MODEL-PIN-001
title: "Pin cross-model audit backend model+effort via the workflow.audit config block"
version: 1.1.0
status: implemented
created: 2026-08-24
updated: 2026-08-24

$ grep -n "sync_commit_sha\|run_commit_sha" .moai/specs/SPEC-V3R6-AUDIT-MODEL-PIN-001/progress.md
484:run_commit_sha: "a7c5c3833"   # backfilled (M5 commit) — placeholder-per-D3 pattern
526:sync_commit_sha: "pending-backfill-sync"   # D3 self-reference exemption — a commit cannot know its own SHA; backfilled after the PR merges (lead owns the implemented → completed transition + this backfill)
551:- `sync_commit_sha` is a pending-backfill placeholder per the D3 exemption — resolves only after the commit lands (lead's post-merge step).

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt252\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt252\b' --oneline
(no output)
```

**Baseline-attribution.** `spec.md` and `progress.md` read at worktree HEAD (`6165f9f5e`, tracking pinned develop `ee50984a`). Landing queries run against the two pinned refs.

**Gaps.** Did not check whether any *other* SPEC besides SPEC-V3R6-AUDIT-MODEL-PIN-001 has since joined the "pending item" pool the card describes (the card is explicitly a running list; only the one named item was verified). Did not check the exact `run_commit_sha` discrepancy (card cites `8d60fb5e4`; progress.md shows `a7c5c3833`) — could be a different commit in a since-superseded state, not verified further (out of the card's central premise).

**Residual-risk.** If a second qualifying SPEC has since appeared, the card's own launch condition ("2+ 건 모였을 때") may already be satisfied, which this sweep did not check for.

**Proposed disposition.** `keep` — premise holds, card is well-formed and still actionable; rests on the spec.md/progress.md read above.

**Overlap candidates.** none observed among in-scope ids — the SPEC-closure-batch mechanism this card describes is not clearly touched by any other in-scope card's text.

---

### t253

**Premise (one sentence).** `internal/sessionmsg/store.go`'s per-recipient pending mailbox has no depth ceiling, so an unconsumed recipient's mailbox can grow without bound.

**Premise verdict.** `holds` — `Store.Send` (store.go, `func (s *Store) Send`) writes unconditionally into `s.pendingDir(toAgentID)` via `writeJSONAtomic`, with no check against an existing pending-count before the write.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t253)

**Claim.** No depth/size cap exists on the pending mailbox at HEAD; the send path in `internal/sessionmsg/store.go` enqueues every validated envelope regardless of how many are already pending for that recipient.

**Evidence.**
```
$ grep -n "pending" internal/sessionmsg/store.go | head -20
103:func (s *Store) pendingDir(id string) string {
...
263:		return writeJSONAtomic(filepath.Join(s.pendingDir(toAgentID), msgID+".json"), env)

$ sed -n '206,266p' internal/sessionmsg/store.go
(func Send — validates from/to agent ids, sender/receiver existence, message shape via env.Validate(),
then unconditionally: s.withAgentLock(toAgentID, func() error {
    return writeJSONAtomic(filepath.Join(s.pendingDir(toAgentID), msgID+".json"), env)
}) — no pending-count read or ceiling check anywhere in this path)
```

**Baseline-attribution.** `internal/sessionmsg/store.go` read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not check `Poll`'s sweep logic (store.go:280-380) for any indirect throttling (e.g. a claimed-batch ceiling that might incidentally bound growth) — read only enough to confirm `Send` itself has no gate. Did not check whether a size/TTL-based passive cleanup exists elsewhere (e.g. expired-envelope sweep on `Poll`) that would bound *effective* growth even without an explicit cap.

**Residual-risk.** If TTL-based expiry (mentioned in the file's own comments, `ExpiresAt`) effectively bounds worst-case growth in practice, the "unbounded" framing may overstate risk somewhat — the card's own scope item (1) ("근거를 먼저 세울 것") anticipates exactly this kind of measurement gap.

**Proposed disposition.** `keep` — premise holds on direct code read; the card correctly identifies an unbounded-write path.

**Overlap candidates.** t254 (same delivering SPEC, SPEC-CODEX-SESSION-MSG-001, different artifact — spec.md/research.md vs store.go); t262 (uses SPEC-CODEX-SESSION-MSG-001 as its own worked example of a template-neutrality-guard gap).

---

### t254

**Premise (one sentence).** `research.md:32` and `spec.md:55` of SPEC-CODEX-SESSION-MSG-001 each contain a backslash-escaped grep-alternation pipe inside a GFM table cell whose recorded "0-hit" verification is unfalsifiable because GFM silently unescapes `\|` when rendering table cells.

**Premise verdict.** `falsified` — neither cited line currently has the hazard the card describes. `research.md:32` is a table cell but already uses the `-e "a" -e "b"` repeated-flag form (no `\|` present to unescape). `spec.md:55` does contain `"session_msg\|session-msg"`, but the line is **not** inside a table — it is a standalone prose paragraph — and GFM does not process backslash escapes inside inline code spans at all (per this project's own memory note `feedback-gfm-unescapes-pipes-in-table-cells`: "this affects table cells only. A plain `|` inside inline code in a bullet list or paragraph needs no escaping and renders correctly as-is").

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t254)

**Claim.** The specific defect the card names is not present at either cited location at worktree HEAD; `research.md:32` was already written in the hazard-avoiding form, and `spec.md:55` was never in a table cell in the first place.

**Evidence.**
```
$ awk 'NR==32{print NR": "$0}' .moai/specs/SPEC-CODEX-SESSION-MSG-001/research.md
32: | 부재 확인 | `grep -rn -e "session_msg" -e "session-msg" internal/ .moai/specs/` → 구현 0건 / SPEC 충돌 0건 (2026-08-23) | 신규 네임스페이스 무충돌 |

$ awk 'NR==55{print NR": "$0}' .moai/specs/SPEC-CODEX-SESSION-MSG-001/spec.md
55: `grep -rn "session_msg\|session-msg" internal/ .moai/specs/` (2026-08-23, 본 워크트리) — 구현 히트 0건, 기존 SPEC 충돌 0건. ...

$ sed -n '48,56p' .moai/specs/SPEC-CODEX-SESSION-MSG-001/spec.md
| 원자적 파일 쓰기 | ... |
| codex_task 패밀리 (위임) | ... |
| C-HRA-008 정적 가드 선례 | ... |
| 임계값 단일 원천 | ... |
                                    <- blank line, table ends here
### §A.3 부재 확인 (이 SPEC이 만들 것)
                                    <- blank line
`grep -rn "session_msg\|session-msg" internal/ .moai/specs/` (2026-08-23, ...)   <- line 55, standalone paragraph, not a table row
```

**Baseline-attribution.** Both files read at worktree HEAD (`6165f9f5e`). Line numbers matched exactly to the card's citation (32 and 55).

**Gaps.** Did not run the file through an actual GFM renderer (`gh api markdown`) to double-confirm code-span backslash preservation — relied on CommonMark spec behavior and this project's own recorded memory note, both of which agree code spans are unaffected. Did not re-verify the underlying "0-hit" grep claims themselves (not the card's premise; the card is about escape-safety of the *recorded* verification, not the verification's current truth).

**Residual-risk.** If some other renderer (not GitHub's GFM) processes backslashes inside code spans differently, the spec.md:55 instance could still be at risk in that renderer — this sweep verified GFM behavior only, consistent with the card's own framing ("GFM 표 셀 안에서는").

**Proposed disposition.** `drop` — the specific defect claimed does not exist at either cited location; rests on the line-55/line-32 reads above plus the project's own prior memory finding on code-span vs table-cell escape behavior.

**Overlap candidates.** t253 (same SPEC dir, different artifact); t262 (same SPEC used as the guard's worked example).

---

### t255

**Premise (one sentence).** A `.git/hooks/pre-commit.local` delegation extension point (deliberately deferred out of t230's scope, per t230's own spec.md §D Out of Scope) remains unimplemented and is a valid follow-up now that t230 has landed and established the provenance-recording precondition it depends on.

**Premise verdict.** `holds` — t230 has landed (backup+disclose for pre-commit hooks), and the codebase explicitly still does NOT implement `pre-commit.local` delegation; a test asserts the disclosure notice must not even name it.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t255 itself)

**Claim.** t255's precondition (t230 landed) is satisfied, and its target gap (pre-commit.local delegation) is confirmed still absent and deliberately excluded, matching the card's own account of why it was split out of t230.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt230\b' --oneline
6786c3fa4 t250: graph freshness, symbol layer, and MCP code queries (SPEC-V3R6-GRAPH-FRESHNESS-001) (#1648)
db1362739 feat(cli): back up and disclose user-modified pre-push hooks (t257) (#1650)
539349c5b docs(t230): t230 sync-audit evidence — SPEC-PRECOMMIT-PRESERVE-001 PASS 95/100 (#1649)
32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)

$ grep -rn "pre-commit.local" internal/cli/hook_install_precommit.go internal/cli/hook_install_precommit_disclosure_test.go
internal/cli/hook_install_precommit.go:255:				// not name pre-commit.local, a facility this SPEC does not ship.
internal/cli/hook_install_precommit_disclosure_test.go:293:	if strings.Contains(warn.String(), "pre-commit.local") {
internal/cli/hook_install_precommit_disclosure_test.go:294:		t.Errorf("the notice must not name pre-commit.local — a recovery path the installed hook never reads (REQ-PCP-004)")

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt255\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/cli/hook_install_precommit.go` and its disclosure test read at worktree HEAD (`6165f9f5e`); t230 landing checked against pinned develop `ee50984a`.

**Gaps.** Did not verify t230's `--is-ancestor` exit code explicitly (not required — t255 itself, not t230, is this card's landing target, and t230 is out of the in-scope batch). Did not check t237/#1641 (named by the card as sharing the same release-level constraint) for its current state.

**Residual-risk.** None specific beyond the card's own stated risk (bundling this with a release that also ships the provenance discriminator would trip every install base's first-upgrade backup+warn) — this sweep did not need to re-litigate that reasoning, only confirm the precondition and the gap.

**Proposed disposition.** `keep` — premise holds; well-formed, correctly gated follow-up.

**Overlap candidates.** t237 (named by the card as sharing the identical release-level sequencing constraint — REQ-PCP-015).

---

### t260

**Premise (one sentence).** The lessons-inbox auto-collection channel captures only tool-call failures and therefore structurally cannot record the more expensive defect class this repository actually catches most often — checks that pass while observing nothing.

**Premise verdict.** `unverified` — the card's specific composition numbers ("최근 400행 중 390행이 tool_failure:Bash:UnknownFailure") could not be re-measured: `.moai/lessons-inbox.jsonl` does not exist in this worktree (it is local/gitignored runtime state, consistent with CLAUDE.local.md's Local-Only Files list). The structural claim (collection is tool-failure-triggered only) is corroborated by the hook wiring but the numeric composition claim itself is unverified here.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t260)

**Claim.** The card is fundamentally a decision-request ("무엇이 실제로 학습되는지를 정직하게 정하는 것"), not a mechanical-defect report; its structural premise (collection is tool-failure-only) is consistent with the codebase's hook wiring, but the specific inbox-composition measurement is stale/unverifiable in this worktree.

**Evidence.**
```
$ ls -la .moai/lessons-inbox.jsonl
ls: .moai/lessons-inbox.jsonl: No such file or directory

$ grep -rln "lessons-inbox" internal/ pkg/ cmd/ | grep -v _test.go
internal/hook/failure_observer.go
(+ 4 template/skill-doc references, not code)

$ grep -n "func \|lessons-inbox" internal/hook/failure_observer.go | head -8
46:func recordToolFailureEvent(input *HookInput, category ErrorCategory) {
87:func recordTestFailEvent(input *HookInput, pkg string) {
129:func appendLessonsInboxStub(root, eventKey, summary, source string) {
```

**Baseline-attribution.** `internal/hook/failure_observer.go` read at worktree HEAD (`6165f9f5e`). The `.moai/lessons-inbox.jsonl` absence check is scoped to this worktree only, at the time of this sweep.

**Gaps.** Did not check `recordTestFailEvent` for whether it captures a broader class than pure tool-call failure (e.g. whether a failed `go test` run itself gets logged, which would be closer to but still distinct from the "check passed while observing nothing" class the card names). Did not check the primary checkout's or lead machine's actual inbox file (out of scope for a worktree-isolated read-only sweep).

**Residual-risk.** If `recordTestFailEvent` or another handler already captures a wider class than "Bash tool_failure", the card's "구조적으로 못 잡는다" framing could be narrower than stated — this sweep did not fully enumerate every event type `failure_observer.go` handles.

**Proposed disposition.** `needs-operator-decision` — the card explicitly asks for a decision on channel design/scope, not a mechanical fix; the structural premise is plausible but the specific numbers are unverified here.

**Overlap candidates.** t280 (same underlying file, `.moai/lessons-inbox.jsonl` — t260 is about collection *scope*, t280 is about drain *deployment*; both share the mechanism).

---

### t262

**Premise (one sentence).** The template neutrality guard's SPEC-ID detection pattern only matches a fixed prefix allowlist (`V3R2-6`, `AGENCY`, `WORKTREE`), which is narrower than the doctrine's blanket prohibition on any SPEC ID in template content, so guard-pass does not imply doctrine compliance.

**Premise verdict.** `holds` — confirmed directly in the guard's source.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t262)

**Claim.** `internal/template/internal_content_leak_test.go`'s C1 pattern is a fixed alternation over a small prefix set; a SPEC ID with any other prefix (the card's own example, `SPEC-CODEX-SESSION-MSG-001`) does not match and would pass the guard undetected.

**Evidence.**
```
$ grep -n "SPEC-V3R6\|regexp.MustCompile" internal/template/internal_content_leak_test.go | grep -i spec
171:		pattern: regexp.MustCompile(`\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b`),
207:		pattern:         regexp.MustCompile(`\bSPEC-V3R[0-9]-[A-Z0-9-]+\b|\bCONST-V3R[0-9]-[0-9]+\b|\bSPEC-WF-AUDIT-GATE-001\b|\bSPEC-MX-001\b`),
```

`SPEC-CODEX-SESSION-MSG-001` matches neither alternation (prefix `CODEX` is not `V3R[2-6]`, `AGENCY`, or `WORKTREE`, nor any of the other literal SPEC-ID alternatives at line 207).

**Baseline-attribution.** `internal/template/internal_content_leak_test.go` read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not enumerate every pattern in the file (10 spec-ID-adjacent entries total per an earlier grep) to build the full "what the guard currently matches" vs "what doctrine forbids" difference set the card's scope item (2) asks for — that is a larger task than this sweep's restraint budget allows; confirmed only that the gap exists, not its full extent.

**Residual-risk.** A broadened general-form pattern (`SPEC-<DOMAIN>-NNN`) risks false positives on prose mentions, exactly as the card's own "대표 mutant" note anticipates — the fix itself is nontrivial, not just the detection.

**Proposed disposition.** `keep` — premise holds on direct source inspection; genuine gap.

**Overlap candidates.** t253, t254 (both use SPEC-CODEX-SESSION-MSG-001, the card's own worked example, as their subject).

---

### t263

**Premise (one sentence).** `file_changed.go:110-118`'s incremental MX-scan path still has the same "die when the sidecar index is missing" defect that t216 M4 already fixed in the `mx query` cold-start path, so M4's self-build fix should be reused there rather than reinvented.

**Premise verdict.** `falsified` — the premise's own precondition is false: t216 has **not landed** (no commit for `\bt216\b` found against either pinned ref), and `mx_query.go`'s cold-start path at HEAD still exhibits the exact pre-fix die behavior the card describes — it is not merely unrepaired in the incremental path, it is *also* still unrepaired in the very path the card cites as already-fixed. A passing test (`TestMxQueryCmd_SidecarUnavailable`) currently asserts this die is the intended, current behavior.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t263; t216 itself has a live worktree — `agent-a62468d0d1a7040cf`/`.claude/worktrees/t216` @ `8aa96bfb1` `[WT-hook-wiring-drift]`, per `00-worktree-list.txt` — but t216 is a separate card, not t263)

**Claim.** The card's premise assumes t216 M4's self-build fix already landed in `mx_query.go`; it has not. `mx_query.go`'s cold-start path still Stat()s for the sidecar file and immediately returns a `SidecarUnavailable` error with no self-build attempt, and `resolver_query.go` mirrors the same shape. There is therefore no landed self-build pattern yet to "reuse" in `file_changed.go`.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt216\b' --oneline
(no output)

$ sed -n '92,104p' internal/cli/mx_query.go
			// verify sidecar file exists (REQ-SPC-004-013)
			sidecarPath := filepath.Join(stateDir, mx.SidecarFileName)
			if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
					"SidecarUnavailable: sidecar index does not exist — run 'moai mx scan' to build the index\n")
				return fmt.Errorf("SidecarUnavailable: no sidecar index")
			}

$ grep -n "TestMxQueryCmd_SidecarUnavailable" internal/cli/mx_query_test.go
90:func TestMxQueryCmd_SidecarUnavailable(t *testing.T) {
(test asserts the "sidecar missing → SidecarUnavailable error" behavior IS correct, per AC-SPC-004-04)

$ sed -n '155,183p' internal/mx/sidecar.go
(Manager.UpdateFile → loadWithoutLock: os.IsNotExist(err) → returns an EMPTY Sidecar with no error — the incremental path file_changed.go actually calls degrades gracefully, unlike mx_query.go)
```

**Baseline-attribution.** `internal/cli/mx_query.go`, `internal/cli/mx_query_test.go`, `internal/mx/sidecar.go`, `internal/mx/resolver_query.go`, and `internal/hook/file_changed.go` all read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not open t216's own worktree (`.claude/worktrees/t216`, `WT-hook-wiring-drift`) to check whether M4's self-build fix exists there in-flight — that would confirm the fix is written but simply unmerged, versus not yet written at all. Did not fully trace whether `file_changed.go`'s actual call chain (via `mx.Manager.UpdateFile`) has any *other* die path besides the one traced (only `loadWithoutLock` and `writeWithoutLock` were read).

**Residual-risk.** If t216's worktree does contain a landed-but-unmerged self-build fix, the card's overall direction (reuse M4's pattern once it lands) is still sound — only the "already fixed, therefore reuse it" framing is premature. The card's own scope item (2) ("M4가 세운 자가빌드 경로를 재사용") depends on M4 actually merging first.

**Proposed disposition.** `needs-operator-decision` — the premise as stated is false today (t216 hasn't landed), but the underlying direction may still be correct once t216 does land; the operator should decide whether to defer this card behind t216 or drop/rewrite it now.

**Overlap candidates.** t216 (hard dependency — the card's own cited fix lives there and has not landed).

---

### t264

**Premise (one sentence).** A large and growing set of local `WT-*` branches from already-merged (squash-merged) cards has accumulated with no worktree occupying them, and cleanup is blocked in the primary checkout by BranchGuard, whose two exemptions are unreachable from a tool-spawned subagent, so an operator must run the cleanup by hand.

**Premise verdict.** `holds` — and the situation has grown since the card's own 2026-08-25 measurement (129 local `WT-*` branches / ~60 orphaned) to 196 local `WT-*` branches / 89 orphaned as of this sweep.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t264 itself)

**Claim.** The orphaned-branch accumulation the card describes is real and has worsened; of the card's own 5 named example branches, 2 (`WT-security-scan-surface`, `WT-web-live-todo`) are still present, while 3 (`WT-codex-session-msg`, `WT-constitution-retire`, `WT-audit-model-pin`) have already been cleaned up since the card was written — partial progress, not full resolution.

**Evidence.**
```
$ git branch --list "WT-*" | wc -l
     196

$ grep -o '\[WT-[^]]*\]' .moai/reports/t332/00-worktree-list.txt | wc -l
     107

# 196 total - 107 worktree-occupied = 89 orphaned local branches

$ git branch --list "WT-codex-session-msg" "WT-web-live-todo" "WT-constitution-retire" "WT-audit-model-pin" "WT-security-scan-surface"
  WT-security-scan-surface
  WT-web-live-todo
```

**Baseline-attribution.** `git branch --list` run from this worktree (`6165f9f5e`) against the shared repository's ref namespace (branches are shared across all worktrees of one `.git`). `00-worktree-list.txt` read as-is per instructions (not re-run).

**Gaps.** Did not verify landing-proof for any specific orphaned branch (per the card's own scope item (1), "저작 경로 좁힌 diff 공집합 확인") — that per-branch verification is explicitly the card's own future work, not something this sweep should pre-empt. Did not check remote branches.

**Residual-risk.** Some of the 89 "orphaned" branches counted here may not actually be landed/mergeable-safe to delete (e.g. an abandoned experiment) — the raw count is not itself a landing proof, only a scale indicator, exactly as the card's own "대표 mutant" warning anticipates.

**Proposed disposition.** `keep` — premise holds and has strengthened; rests on the branch-count and worktree-occupancy comparison above.

**Overlap candidates.** none observed among in-scope ids beyond general worktree-lifecycle housekeeping.

---

### t280

**Premise (one sentence).** The lessons-inbox *collection* mechanism ships to every distributed user via the compiled `moai` binary and the template-shipped `PostToolUseFailure` hook wiring, but the *drain* trigger that consumes it only exists in this repository's local, untracked `settings.local.json`, so a deployed user's `lessons-inbox.jsonl` grows without any drain path at all.

**Premise verdict.** `holds` — confirmed on both halves: collection ships (Go code + template hook wiring), drain does not (absent from both tracked `settings.json` and the distributed template).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t280)

**Claim.** `internal/hook/failure_observer.go` (compiled into the `moai` binary, therefore shipped to every install) writes to `.moai/lessons-inbox.jsonl`, and the distributed template's `settings.json.tmpl` wires `PostToolUseFailure` to a handler that reaches it; no drain trigger (`lsel`/`drain`/`session_drain`) appears anywhere in the tracked `.claude/settings.json` or the template's `settings.json.tmpl`.

**Evidence.**
```
$ grep -n "func \|lessons-inbox" internal/hook/failure_observer.go | head -6
46:func recordToolFailureEvent(input *HookInput, category ErrorCategory) {
129:func appendLessonsInboxStub(root, eventKey, summary, source string) {
139: ... marshal lessons-inbox stub ...

$ grep -n "PostToolUseFailure" internal/template/templates/.claude/settings.json.tmpl
209:    "PostToolUseFailure": [
214:            ... "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-post-tool-failure.sh"

$ grep -n "lsel\|drain" .claude/settings.json
(no output)

$ grep -n "lsel\|drain\|lessons-inbox" internal/template/templates/.claude/settings.json.tmpl
(no output)
```

**Baseline-attribution.** `internal/hook/failure_observer.go`, `.claude/settings.json` (tracked), and `internal/template/templates/.claude/settings.json.tmpl` all read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not check whether `handle-post-tool-failure.sh` actually invokes the `moai hook` subcommand that reaches `recordToolFailureEvent` (traced the Go side and the settings wiring, not the intermediate shell wrapper). Did not verify the t259 (SPEC-LSEL-DRAIN-STALL-001) fix's exact scope boundary beyond what CLAUDE.local.md §28 already documents (this card explicitly treats that as its starting fact, not something to re-derive).

**Residual-risk.** If some other, undiscovered drain path exists for distributed users (e.g. a periodic `moai` subcommand run manually), the "permanent absence" framing could be too strong — this sweep found no such path in the settings/template surfaces checked, but did not exhaustively search all CLI subcommands.

**Proposed disposition.** `keep` — premise holds; well-evidenced structural gap between what ships and what doesn't.

**Overlap candidates.** t260 (same underlying `lessons-inbox.jsonl` mechanism, different concern — collection *scope* vs drain *deployment*).

---

### t281

**Premise (one sentence).** The operator decided on 2026-08-26 (C안) to promote the local `develop` branch to a standing, local-only integration/RC-testbed branch while keeping per-card PRs against `main` and never pushing `develop` to origin.

**Premise verdict.** `falsified` — this decision was explicitly and completely reversed the very next day. `CLAUDE.local.md` §4.1, as it exists at this worktree's HEAD, documents a full switch to git-flow: `develop` is now pushed to origin (`origin/develop`), card PRs against `main` are abolished in favor of direct merges into a shared `develop` integration worktree, and `release/vX.Y.Z` branches from `develop` is the only path to `main`. §4.1.3 of the same file states this in its own words: "이 절은 2026-08-26 백로그 카드 t281이 정한 '로컬 전용·일회용 develop, 원격 push 금지'를 명시적으로 뒤집는다."

**Landing verdict.** `not-landed`
- commit: `11216d13f612f7e7161487a4e1369a47612f0b4c` — **not a delivery of t281; the opposite: an explicit reversal of the t281 decision**, per its own commit message ("Operator-directed reversal of the 2026-08-26 t281 decision")
- pinned ref: `ee50984abe4f11ac337382b48a26328f091e200a`
- `--is-ancestor` exit: `0`
- branch + tip (in-flight only): —

**Claim.** No commit implements t281's C안 as described (local-only, disposable develop with per-card PRs). The one commit whose message mentions `t281` (`11216d13f6`) is a reversal, dated one day after the card's own cited decision date, ancestor-confirmed in pinned develop. t281's action items (rc.N numbering convention, `make build VERSION=` procedure documentation, BranchGuard-routing decision, etc.) target a model that CLAUDE.local.md itself now documents as superseded.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt281\b' --oneline
11216d13f docs(workflow): switch development model from GitHub Flow to git-flow

$ git show --stat 11216d13f | head -14
commit 11216d13f612f7e7161487a4e1369a47612f0b4c
Author: t <t@t.t>
Date:   Thu Aug 27 12:55:23 2026 +0900

    docs(workflow): switch development model from GitHub Flow to git-flow

    Operator-directed reversal of the 2026-08-26 t281 decision (local-only,
    disposable develop). Card worktrees now branch from develop, lanes merge
    directly into a single develop integration worktree with no per-card PR,
    origin/develop is pushed and CI-verified, rc.N builds are cut from
    develop, and release/vX.Y.Z branched from develop is the only path to
    main.

$ git merge-base --is-ancestor 11216d13f612f7e7161487a4e1369a47612f0b4c ee50984abe4f11ac337382b48a26328f091e200a ; echo "exit:$?"
exit:0

$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt281\b' --oneline
(no output)
```

**Baseline-attribution.** `CLAUDE.local.md` §4.1/§4.1.3 as loaded into this session's system context at worktree HEAD (`6165f9f5e`); commit `11216d13f` checked against pinned develop via `--grep` + `--is-ancestor`.

**Gaps.** Did not diff `CLAUDE.local.md`'s exact wording against t281's five numbered scope items one-by-one (e.g. whether the rc.N numbering convention t281 asked for was separately addressed by the git-flow rewrite) — the supersession is total enough (opposite branch-push policy) that a line-by-line reconciliation did not seem load-bearing within this sweep's restraint budget.

**Residual-risk.** If any of t281's five scope sub-items (rc.N reset convention, `make build`/reinstall procedure documentation) were NOT actually covered by the git-flow rewrite, there could be a genuine residual gap hiding behind the "superseded" verdict — this sweep did not check for that.

**Proposed disposition.** `drop` — the decision the card documents and proposes to formalize has already been explicitly reversed by a later, more comprehensive operator decision recorded in the same file the card would have edited.

**Overlap candidates.** t310 (`WT-gitflow-doc-align`, live worktree per `00-worktree-list.txt` — very likely the card already carrying git-flow documentation-alignment work that supersedes t281's scope).

---

### t284

**Premise (one sentence).** `audit_multi`'s `disagreement_flag` derivation counts only the number of *distinct verdict values* among required backends, not the number of backends that actually participated on-target, so a convergence with only one genuine on-target verdict (others excluded as off-target or inconclusive) reports `disagreement_flag=false` — "agreement" — when there was nothing to compare.

**Premise verdict.** `holds` — confirmed directly in `converge()`'s Step 2.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree for t284)

**Claim.** `internal/cli/mcp_convergence.go`'s `disagreement := len(distinctRequired) > 1` derives the flag purely from the *count of distinct pass/fail values*, never from the count of participating/on-target backends; `ConvergenceResult` carries no participant-count or on-target-count field at all.

**Evidence.**
```
$ sed -n '166,171p' internal/cli/mcp_convergence.go
	// ── Step 2: disagreement_flag derivation ──
	distinctRequired := distinctVerdicts(required, "pass", "fail")
	disagreement := len(distinctRequired) > 1

$ sed -n '104,113p' internal/cli/mcp_convergence.go
type ConvergenceResult struct {
	PerBackendVerdicts []PerBackendVerdict `json:"per_backend_verdicts"`
	OverallVerdict   string   `json:"overall_verdict"`
	DisagreementFlag bool     `json:"disagreement_flag"`
	ResidualRiskNote string   `json:"residual_risk_note"`
	FailOpenBackends []string `json:"fail_open_backends"`
}
```
No field for participant/on-target count exists in the struct; with exactly one required backend returning `pass` (others excluded/inconclusive), `distinctRequired = ["pass"]`, `len(distinctRequired) == 1`, so `disagreement = false`.

**Baseline-attribution.** `internal/cli/mcp_convergence.go` read at worktree HEAD (`6165f9f5e`).

**Gaps.** Did not check `.moai/reports/t229/succession.md` directly (the card cites it as its observation source) — took the card's own quoted description of the codex-off-target/GLM-inconclusive scenario as given and verified only that the current code has no structural defense against it. Did not check whether `PerBackendVerdicts` (which IS exposed) lets a *consumer* reconstruct the on-target count themselves, which would partially mitigate the gap even without a dedicated field.

**Residual-risk.** If `PerBackendVerdicts` already lets a careful reader count on-target backends themselves, the practical severity may be lower than "disagreement_flag lies outright" — the flag itself is still misleading in isolation, which is the card's core point.

**Proposed disposition.** `keep` — premise holds on direct code read.

**Overlap candidates.** none observed among in-scope ids (t229, the succession-record source, is not in the in-scope list — already closed per repository memory).
# t332 card sweep — batch 4

Cards: t286 t287 t288 t295 t296 t297 t300 t302 t304 t305  (10 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t286

**Premise (one sentence).** A "risky-command guard" regex has a bidirectional defect — a flag-order bypass (evasion) and a quoted-data false positive (over-match) coexist, per issue #1658.

**Premise verdict.** `unverified` — I located a strong candidate regex guard (`internal/hook/branch_guard.go`, `branchStatePatterns` + `quotedArgumentPattern`) that already handles the quoted-data false-positive class explicitly (a dedicated `quotedArgumentPattern`/`substituteQuotedArguments` pass collapses quoted spans before matching, with a comment citing the exact failure mode: `moai todo add "… git switch …"` being wrongly denied). Its own comments also document a flag-order-shaped gap (`git branch -vD foo` combined short flags "do not match") but explicitly label that as an **accepted, documented direction for a fail-open guard**, not a live defect. I could not confirm within budget whether issue #1658 names this guard specifically or a different "risky command" guard elsewhere in the tree (e.g. a Bash-tool-wide risk-amplifier check referenced only in doctrine, `coding-standards.md` §Bash Risk-Amplifier Doctrine, with no matching Go implementation found). The card may be describing a guard I did not find, or may be re-describing an already-accepted tradeoff as a bug.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree named for this card in `00-worktree-list.txt`)

**Claim.** No commit delivering a fix for a "risky-command guard flag-order bypass / quoted-data false positive" defect exists on either pinned ref; the card is unresolved. Whether the described defect is real, and against which specific guard, is unresolved by this sweep.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt286\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt286\b' --oneline
(no output)
$ grep -rl "risky\|dangerous" internal/hook/*.go | grep -v _test
internal/hook/branch_guard.go
internal/hook/pre_tool.go
internal/hook/types.go
```
`internal/hook/branch_guard.go:156` — `var quotedArgumentPattern = regexp.MustCompile(`'[^']*'|"[^"]*"`)` (already-shipped false-positive fix).
`internal/hook/branch_guard.go:96-102` — comment: "exotic combined short flags (`git branch -vD foo`) do not match — under-matching an obfuscated form is the documented correct direction for a fail-open guard" (labeled accepted, not a defect).

**Baseline-attribution.** File contents read at worktree HEAD `6165f9f5e`. Commit-grep run against the two pinned SHAs listed above (fetched 2026-08-30T11:16:22Z per WORKER-INSTRUCTIONS.md).

**Gaps.** Did not read GitHub issue #1658 itself (no network access in this sweep) to confirm which guard it names. Did not exhaustively search every regex-based command guard in the tree — restraint budget bounds this to `internal/hook/*.go`. Did not attempt to reproduce either claimed failure mode (bypass input, false-positive input) against `branch_guard.go` or any other candidate.

**Residual-risk.** If issue #1658 names a different guard than `branch_guard.go`, this entire premise assessment is about the wrong artifact. If it does name `branch_guard.go`, the "flag-order bypass" may be the already-accepted residual the comments describe, in which case the card may be asking to un-accept a deliberate tradeoff rather than fix a bug.

**Proposed disposition.** `needs-operator-decision` — rests on: the found candidate guard already documents the false-positive fix and explicitly accepts the flag-order gap as correct-by-design; the operator should confirm whether issue #1658 targets this guard before dispatching repair work.

**Overlap candidates.** t287 (same "guard defect via GitHub issue" pattern, adjacent issue #1659, same session/date). No other in-scope id names a guard-regex artifact.

---

### t287

**Premise (one sentence).** A "worktree guard" blocks heredoc-brace-folded dangerous commands but not an equivalent command-substitution (`$(...)`) at the same position, per issue #1659 and the lead's own-session observation.

**Premise verdict.** `unverified` — I have first-hand, directly-observed evidence bearing on this premise from *this same sweep session*: a compound Bash command (`for id in ...; do ...; done`) issued in this session was refused with the message "This session is isolated in the worktree ... Refusing to run it — a worktree-isolated session's git operations must target its own worktree. Split it into plain, separate commands." A second, non-git compound command (multi-file `ls -d` inside a shell loop) was refused with the *identical* message even though it touched no git state at all. This confirms the card's own observation ("이 세션도 until 루프·... 포함 명령 거부됐으나"). Critically, `internal/hook/branch_guard.go:182-183` documents, in its own comments, that this behavior belongs to a SEPARATE mechanism: *"The Claude Code worktree isolation guard refuses that same shape independently"* — i.e. this repo's own `branch_guard.go` explicitly disclaims ownership of the compound-command-refusal behavior I personally observed. I found no Go source in `internal/` implementing a heredoc-vs-command-substitution guard matching the card's description. This raises a real possibility that the artifact the card wants patched is Claude Code's own harness sandbox (external to this repository, not fixable by a moai-adk-go code change), not something in this codebase — but I could not fully rule out that moai-adk-go ships its own complementary hook that ALSO does heredoc-folding (e.g. inside `internal/permission/stack.go`, which does implement an AST-based, non-regex command-substitution and IFS-word-split detector — a much stronger mechanism than "heredoc brace-folding regex" and does not obviously have the described gap).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** No delivering commit exists on either pinned ref. The premise cannot be confirmed as an in-repo defect within this sweep's budget; there is concrete evidence the described refusal behavior is (at least partly) owned by Claude Code's own harness, external to this repository.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt287\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt287\b' --oneline
(no output)
```
Directly observed in this session (tool-call error, not a file):
> "This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t332, but this command is too complex to verify that it stays inside the worktree. Refusing to run it..."
(triggered twice: once by a `for`-loop with git subcommands, once by a `for`-loop with plain `ls -d`, i.e. not git-specific.)

`internal/hook/branch_guard.go:181-183`:
```
// fails open on every uncertainty, so under-matching a deliberately obfuscated
// form is the correct direction to err. The Claude Code worktree isolation
// guard refuses that same shape independently.
```

**Baseline-attribution.** Session-observed tool refusal text from this sweep run (2026-08-30). File comment read at worktree HEAD `6165f9f5e`. Commit-grep against the two pinned SHAs.

**Gaps.** Did not read issue #1659 (no network). Did not test `internal/permission/stack.go`'s AST-based command-substitution detector against the card's specific heredoc-vs-`$(...)` bypass scenario. Did not determine whether `internal/permission/stack.go`'s guard and the sandbox-level "worktree isolation guard" I personally hit are the same mechanism or two independent layers.

**Residual-risk.** If the card in fact targets `internal/permission/stack.go` (an AST parser, not a "heredoc brace-folding regex" as described), the premise's characterization of the mechanism is wrong even if a real gap exists elsewhere in that file. If it targets the external Claude Code sandbox, this card cannot be resolved by a moai-adk-go code change at all.

**Proposed disposition.** `needs-operator-decision` — rests on: this session's own two directly-observed refusals of non-git compound commands, matching the "Claude Code worktree isolation guard" that `branch_guard.go` itself says is external to this repo's mechanism.

**Overlap candidates.** t286 (same issue-driven guard-defect pattern, adjacent issue #1658). No other in-scope id touches command-guard regexes.

---

### t288

**Premise (one sentence).** The `goal_arm` MCP wrapper misclassifies a prose (model) condition as mechanical whenever the prose text does not contain the literal word "transcript," causing the shell-command execution path to run the prose as a command and the stop-hook to exit 2 every turn-end.

**Premise verdict.** `holds` — confirmed by reading the actual classifier. `internal/cli/goal.go:37` (`parseCondition`) implements the classification rule EXPLICITLY as: `if strings.Contains(strings.ToLower(s), "transcript") { return Condition{Type: ConditionModel, ...} }` — else the ENTIRE string is treated as `ConditionMechanical` and stored as a shell `Cmd`. `internal/cli/mcp_server.go:641` confirms the MCP tool `goal_arm` calls `cond := parseCondition(conditionText) // same classifier the CLI uses` — i.e. the MCP wrapper inherits the identical single-keyword discriminator, with no additional heuristic. Any prose claim that omits the literal substring "transcript" (plausible in Korean-language or differently-worded English claims) is stored as a mechanical condition and will be run as a shell command by the stop-goal evaluator — a non-existent/invalid "command" would exit non-zero every turn, matching the card's claimed "매 턴엣드 exit 2" symptom.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** The classifier's single-keyword ("transcript") discriminator is unchanged on both pinned refs; no commit adds a broader/more robust model-vs-mechanical classification path.

**Evidence.**
```
internal/cli/goal.go:29-49 (parseCondition):
func parseCondition(s string) goal.Condition {
	s = strings.TrimSpace(s)
	if strings.Contains(strings.ToLower(s), "transcript") {
		return goal.Condition{Type: goal.ConditionModel, Claim: s}
	}
	cmd := s
	...
	return goal.Condition{Type: goal.ConditionMechanical, Cmd: cmd, ExpectExit: expect}
}

internal/cli/mcp_server.go:641:
	cond := parseCondition(conditionText) // same classifier the CLI uses

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt288\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt288\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/cli/goal.go` and `internal/cli/mcp_server.go` read at worktree HEAD `6165f9f5e`. Commit-grep against the two pinned SHAs.

**Gaps.** Did not actually arm a goal with a non-"transcript" prose condition and observe a live turn-end exit 2 (that would require a live stop-hook cycle, outside this read-only sweep's scope). Did not check `internal/goal` package's `stop-goal` evaluator to confirm it truly shells out `Cmd` unconditionally for `ConditionMechanical` (inferred from the type name and comment, not directly read).

**Residual-risk.** If the stop-goal evaluator has an additional guard that detects "this mechanical command looks like prose and refuses to run it," the exit-2 loop symptom described might not actually occur even though the classification itself is as described. I did not verify the evaluator side.

**Proposed disposition.** `keep` — rests on: `parseCondition`'s single-keyword classifier is unchanged and is shared verbatim between the CLI and the MCP wrapper, so the misclassification path the card describes is real and reachable via the MCP entry point specifically.

**Overlap candidates.** None observed among in-scope ids (memory notes list related-but-not-in-scope cards `feedback_goal_arm_mcp_worktree_split.md` and `feedback_goal_keying_worktree_unreliable.md`, but their card ids are not in `inscope-all.txt`).

---

### t295

**Premise (one sentence).** No launcher-exposed path exists to create a worktree that checks out an EXISTING branch (as opposed to creating a new one), forcing the lead to bypass `moai worktree`'s launcher discipline with a raw `git worktree add` for the gitflow `develop` integration tree.

**Premise verdict.** `holds` — confirmed at two levels. (1) The underlying git-level primitive DOES support existing branches: `internal/core/git/worktree.go`'s `Add(path, branch)` calls `branchExists()` and dispatches to `buildWorktreeAddArgs(exists, ...)`, which for `exists==true` emits `worktree add -- <path> <branch>` (checkout existing) vs. `-b <branch>` for new. (2) But the actual `moai cc -w` launch path does NOT use this existing-branch-capable function: `internal/cli/session_worktree.go:52` documents its own helper as "`sessionWorktreeGitWorktreeAdd` runs `git worktree add -b <branch> <dest>`" — hardcoded to the `-b` (always-new-branch) form, with no code path found that ever calls it with an existing branch name. This matches the card's claim exactly: the git-level capability exists somewhere in the codebase, but no launcher (`moai cc -w`, and `EnterWorktree` is a native Claude Code tool outside this repo entirely) exposes a way to invoke it for an existing branch.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Two commits matched the grep on develop but are MENTIONS, not deliveries — recorded per the false-positive warning in the instructions:
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt295\b' --oneline
daf206903 Merge WT-clocal-audit into develop — t308 CLAUDE.local.md audit (SPEC-CLOCAL-AUDIT-001)
281fde607 docs(t308): audit CLAUDE.local.md against measured reality — 20 defects fixed
```
Both commits' bodies read: "§4.1 excluded (t294/t295/t298/t303)" / "Scope: §4.1 (git-flow lane section) excluded — owned by t294/t295/t298/t303." — t295 is explicitly named as OUT OF SCOPE for that audit, i.e. this is a citation, not a landing. Confirmed by reading `git show --no-patch` on both SHAs in full.

**Claim.** t295 is unresolved; the underlying `git.WorktreeManager.Add` already supports an existing-branch code path but the `moai cc -w` launcher does not use it — the card's "no sanctioned path" claim is accurate for the actual launcher, even though the capability exists one layer down.

**Evidence.**
```
internal/core/git/worktree.go:58-73 (buildWorktreeAddArgs):
func buildWorktreeAddArgs(branchExists bool, path, branch string) []string {
	if branchExists {
		return []string{"worktree", "add", "--", path, branch}
	}
	return []string{"worktree", "add", "-b", branch, "--", path}
}

internal/cli/session_worktree.go:52-57:
	// sessionWorktreeGitWorktreeAdd runs `git worktree add -b <branch> <dest>
	...
	sessionWorktreeGitWorktreeAdd = gitWorktreeAddReal

$ grep -rln "GitWorktree\.Add\|GitWorktree\b" internal/cli/*.go | grep -v _test | grep -v deps.go
internal/cli/inventory.go
internal/cli/root.go
internal/cli/session_worktree.go
```
`session_worktree.go` — the file backing `moai cc -w` / `moai worktree` — never calls the existing-branch-capable `git.WorktreeManager.Add`; it has its own separate, hardcoded-`-b` helper.

**Baseline-attribution.** All three files read at worktree HEAD `6165f9f5e`. Commit-grep + `git show --no-patch` against the two pinned SHAs.

**Gaps.** Did not trace every call site of `sessionWorktreeGitWorktreeAdd` to rule out a hidden existing-branch code path elsewhere in `session_worktree.go` (only the doc comment and the variable wiring were read). Did not check whether `internal/cli/inventory.go` or `root.go` (which also reference `GitWorktree`) expose an existing-branch path through a different subcommand (e.g. `moai worktree recover`).

**Residual-risk.** If `inventory.go` or `root.go` DO wire `GitWorktree.Add` to an existing-branch flag somewhere, the premise would be partially falsified (a path exists, just undocumented/undiscovered). This sweep did not read those two files.

**Proposed disposition.** `keep` — rests on: `session_worktree.go`'s own doc comment hardcodes `-b <branch>` (always new branch) and no call site was found passing an existing branch, while the lower-level git manager already supports it — a real gap between capability and exposed surface.

**Overlap candidates.** t231 (in-scope — "worktree clean 앵커 소스 판독", cited directly by this card as related). t297 (same batch, also cites t231 in its own "연관" list, and both surfaced from the same 2026-08-27 gitflow-transition session). The card also names t209/t298/t303/t294 as related, none of which are in `inscope-all.txt`.

---

### t296

**Premise (one sentence).** `coding-standards.md`'s "Language Policy" section carries no "16-language programming-neutrality" content (it is about writing instruction docs in English), yet ~10 other template files cite it as the canonical location for that contract, so readers following the citation cannot reach the promised body.

**Premise verdict.** `holds` — confirmed directly. `internal/template/templates/.claude/rules/moai/development/coding-standards.md`'s `## Language Policy` section (lines 9-22) reads: "All instruction documents must be in English: CLAUDE.md, Agent definitions, ... User-facing documentation may use multiple languages: README.md, CHANGELOG.md, ..." — no mention of "16," no programming-language list, no neutrality contract. Meanwhile 7 distinct template files contain the literal phrase "16-language neutrality contract" (8 occurrences total, vs. the card's claimed "10곳" — close but not exact; the discrepancy does not affect the core claim).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** The cited section still lacks the promised content and the dangling citations still exist, on both pinned refs.

**Evidence.**
```
$ sed -n '9,22p' internal/template/templates/.claude/rules/moai/development/coding-standards.md
## Language Policy

All instruction documents must be in English:
- CLAUDE.md
- Agent definitions (.claude/agents/**/*.md)
- Slash commands (.claude/commands/**/*.md)
- Skill definitions (.claude/skills/**/*.md)
- Hook scripts (.claude/hooks/**/*.py, *.sh)
- Configuration files (.moai/config/**/*.yaml)

User-facing documentation may use multiple languages:
- README.md, CHANGELOG.md
- User guides, API documentation

$ grep -rln "16-language neutrality contract" internal/template/templates/
internal/template/templates/.moai/docs/generic-patterns-guide.md
internal/template/templates/.claude/rules/moai/development/skill-authoring.md
internal/template/templates/.claude/skills/moai/workflows/loop.md
internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md
internal/template/templates/.claude/skills/moai-workflow-loop/SKILL.md
internal/template/templates/.claude/skills/moai-workflow-loop/references/examples.md
internal/template/templates/.claude/skills/moai-workflow-loop/references/reference.md

$ grep -rn "16-language neutrality contract" internal/template/templates/ | wc -l
8

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt296\b' --oneline
(no output)
```

**Baseline-attribution.** `coding-standards.md` and all 7 citing files read at worktree HEAD `6165f9f5e`. Grep scope: `internal/template/templates/` (the distributed template root; did not scan the local `.claude/` mirror separately). Commit-grep against both pinned SHAs.

**Gaps.** Did not check whether the local (non-template) mirror `.claude/rules/...` and `.claude/skills/...` carry the same 7 dangling citations (only the template source was scanned, per the card's own framing that the template is what ships to users). Did not verify the "8 occurrences, 7 files" count against the card's claimed "10곳" beyond noting the discrepancy — could be an outdated count in the card, or a scope difference (e.g. the card may also be counting occurrences in `.claude/` local mirror or in generated/rendered files).

**Residual-risk.** The card's count (10) vs. measured count (7 files / 8 occurrences) is close but not exact — if the operator treats the exact count as load-bearing, this is a minor discrepancy to flag, not a premise failure.

**Proposed disposition.** `keep` — rests on: direct read of the cited section (no "16" or language list) plus a direct grep confirming multiple dangling citations still pointing at it.

**Overlap candidates.** None observed among in-scope ids.

---

### t297

**Premise (one sentence).** The launch-ledger (`launch.yaml` `projects[]`) grows unboundedly because every worktree that ever records a launch profile gets its own permanent entry with no reaping mechanism when the worktree is later removed, and this was requested explicitly as REQ-009 follow-up from t293's plan-auditor.

**Premise verdict.** `holds`, with a nuance the card's scope item (1) does not need — confirmed by reading the write path. `internal/profile/profile.go`'s ledger-write function (~line 555-583) already performs **duplicate-spelling dedup**: `if stored, found := lookupProjectKey(projects, key); found { key = stored }` before writing, specifically to avoid two entries for the same directory under different path spellings (own comment: "Writing the caller's spelling instead would leave two entries naming one project"). This means scope item (1) as literally stated ("중복 행을 만들지 않고 갱신") is PARTIALLY already done for same-directory-different-spelling. However, scope item (2) — reaping dead worktree entries — is confirmed still absent: no `Prune`/`Reap`/`RemoveProject`/`delete(projects...)` function was found anywhere in `internal/profile/` or `internal/cli/`. Since each worktree is a genuinely distinct directory (not a spelling variant), every worktree that ever records a profile gets a permanent, never-reaped entry — this is the growth mechanism the card actually describes, and it is real.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

The originating commit (a MENTION, not a delivery — this is the commit that itself proposed t297 as a follow-up card):
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt297\b' --oneline
2114ed981 docs(SPEC-STATUSLINE-PROFILE-RESPECT-001): sync-phase artifacts — in-progress -> implemented (t293)
```
Body: "AC-009 / M5 deferred by kickoff decision D1; follow-up card t297 queued." — confirms t297 is a request, not a resolution.

**Claim.** No pruning/reaping mechanism for stale `projects[]` ledger entries exists on either pinned ref; ledger growth is unbounded as worktrees accumulate and are removed without a corresponding ledger cleanup.

**Evidence.**
```
internal/profile/profile.go:561-579 (write path, dedup already present):
	if key := normalizeProjectKey(projectRoot); key != "" {
		projects, ok := existing[projectsKey].(map[string]any)
		if !ok {
			projects = make(map[string]any)
		}
		if stored, found := lookupProjectKey(projects, key); found {
			key = stored
		}
		projects[key] = name
		existing[projectsKey] = projects
	}

$ grep -rn "func.*[Pp]rune\|func.*[Rr]eap\|func.*[Rr]emoveProject\|delete(projects" internal/profile/*.go internal/cli/*.go | grep -v _test
internal/cli/chain.go:474:func newChainPruneCmd() *cobra.Command
internal/cli/chain.go:487:func runChainPrune(...)
```
(The only "prune" hits are `chain.go`'s unrelated chain-state prune command, not the launch ledger.)

**Baseline-attribution.** `internal/profile/profile.go` (full write function, lines ~549-614) and a repo-wide grep of `internal/profile/*.go` + `internal/cli/*.go` for prune/reap functions, read/run at worktree HEAD `6165f9f5e`.

**Gaps.** Did not measure an ACTUAL ledger's current row count / growth rate (would require access to a real `~/.moai/launch.yaml` across many worktree lifecycles, out of this sweep's read-only repo-tree scope). Did not check whether `moai worktree done` or `moai worktree clean` (the disposal commands) happen to call any ledger-cleanup code not matched by my prune/reap keyword grep.

**Residual-risk.** If a differently-named function (not matching "prune/reap/remove/delete") performs cleanup on worktree disposal, my keyword grep would miss it — the absence claim rests on the specific keyword set searched, not an exhaustive read of every disposal code path.

**Proposed disposition.** `keep` — rests on: no prune/reap function found anywhere in the two most relevant packages, confirming scope item (2) — the reaping half of the card — is genuinely unaddressed, even though the dedup half (scope item (1)) is partially already shipped.

**Overlap candidates.** t231 (in-scope — worktree-clean anchor-source reading, cited directly by this card). t295 (same batch, both surfaced 2026-08-27, both cite t231). The card also names t293/t209 as related; neither is in `inscope-all.txt`.

---

### t300

**Premise (one sentence).** A "baseline-first" acceptance criterion (AC-GF-022, requiring the baseline measurement to precede the first implementation commit) was violated when both landed in the same commit and then became permanently unverifiable after a squash merge, and this card exists to prevent recurrence of that pattern in future SPECs.

**Premise verdict.** `holds` — this card's own premise is directly corroborated by the commit that spawned it, read in full. `git show --no-patch` on the matched commit shows the operator-decision record stating the exact facts the card restates: "7f2e9e77d carries m5-baseline.md (+67) together with the M5 implementation files in one commit; that commit is unreachable from origin/main and origin/develop (both ancestor checks exit 1; PR #1648 squash 6786c3fa4 carries the artifacts into history)... Ordering is permanently unverifiable, not deferred... Recurrence prevention: card t300." This is a first-party, already-landed acknowledgment of the exact defect t300 describes and an explicit statement that t300 was created to prevent recurrence — the premise needs no further independent verification beyond this citation, since the citation IS the origin of the claim.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

The originating (mention) commit:
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt300\b' --oneline
69891ce99 docs(t279): record AC-GF-022 ordering deviation at source (operator decision A, card t279)
```
This commit records the deviation and creates the follow-up card; it does NOT implement the recurrence-prevention procedure t300 asks for (scope items 1-3: procedural guard for baseline-first ACs, squash-merge interaction, sweep of other SPECs for the same pattern).

**Claim.** No commit implements a procedural or mechanical guard against future baseline-in-same-commit ACs; t300's recurrence-prevention work remains undone.

**Evidence.**
```
$ git show --no-patch --format='%H%n%s%n%b' 69891ce99
69891ce99773563ec17e10961737b5284582fe18
docs(t279): record AC-GF-022 ordering deviation at source (operator decision A, card t279)
Operator decision A (2026-08-27): keep status: completed, record the AC-GF-022
ordering deviation in place at the AC definition in acceptance.md. Facts
re-measured in this tree before writing: both m5 artifacts exist; 7f2e9e77d
carries m5-baseline.md (+67) together with the M5 implementation files in one
commit; that commit is unreachable from origin/main and origin/develop (both
ancestor checks exit 1; PR #1648 squash 6786c3fa4 carries the artifacts into
history). Ordering is permanently unverifiable, not deferred; post-artifact
existence/content remains verifiable. Recurrence prevention: card t300.
Cross-references the progress.md §E.4 open_followups row.
```

**Baseline-attribution.** Full commit message of `69891ce99` read directly via `git show`, against worktree HEAD `6165f9f5e`. Commit-grep against both pinned SHAs.

**Gaps.** Did not read `SPEC-V3R6-GRAPH-FRESHNESS-001`'s `acceptance.md` directly to see the current state of the AC-GF-022 deviation note. Did not sweep other SPECs' acceptance.md files for the same baseline-first-AC pattern (that IS scope item (3) of the card itself — explicitly out of this sweep's bounded-depth budget, consistent with the Restraint clause).

**Residual-risk.** None specific beyond the general risk that a procedural fix (documentation-only, per the card's own "주의 - 대표 mutant" warning about doc-only non-fixes) might land without an actual verification mechanism, which is exactly the failure mode the card itself warns against.

**Proposed disposition.** `keep` — rests on: the originating commit itself both confirms the defect and explicitly names t300 as the not-yet-done recurrence-prevention follow-up.

**Overlap candidates.** t291 named by the card itself ("F5" squash-provenance orphaning — same root cause) — not in `inscope-all.txt`. t279 (the originating SPEC/card) — also not in `inscope-all.txt`. No in-scope id overlaps.

---

### t302

**Premise (one sentence).** `.claude/workflows/sync-audit-4dim.js` states two contradictory things about its own verdict authority in the same file — the header says its verdict was PROMOTED to binding on the happy path, while the `meta.description` field still says it is "an execution vehicle, NOT the binding sync-auditor verdict owner."

**Premise verdict.** `holds` — confirmed by reading the file directly (both cited passages are present verbatim, in the LOCAL, currently-loaded copy at `.claude/workflows/sync-audit-4dim.js`, i.e. the exact file this session's own skill listing describes with matching text). Header (lines 4-10): "SPEC-AUDIT-SNAPSHOT-001 (A3) PROMOTED its verdict to BINDING on the happy path: where the verdict is PASS with all four dims above their floor, not INCOMPLETE, and no contested finding, the orchestrator treats this workflow's harmonic-mean verdict as the binding sync-phase verdict and does NOT spawn the cold `sync-auditor` subagent." `meta.description` field (line ~50): "Sync-phase 4-dimension quality read (Functionality/Security/Craft/Consistency) — parallel read-only judges + in-script harmonic-mean verdict; **execution vehicle, NOT the binding sync-auditor verdict owner**." These two statements directly conflict on the exact question the card names: whether this script's verdict is binding. The negation in `meta.description` also independently confirmed at session start — this exact string appears in this session's own available-skills listing for `sync-audit-4dim`, meaning the contradiction is externally visible, not just an internal file artifact.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** No commit resolves the header-vs-description contradiction in `sync-audit-4dim.js` on either pinned ref.

**Evidence.**
```
$ sed -n '1,12p' .claude/workflows/sync-audit-4dim.js
// sync-audit-4dim.js — 4-dimension sync-phase quality verdict (Context → Judge → Verdict)
//
// VERDICT SCOPING (what this workflow IS and is NOT):
//   This is an EXECUTION VEHICLE for a skeptical 4-dimension quality read. SPEC-AUDIT-SNAPSHOT-001
//   (A3) PROMOTED its verdict to BINDING on the happy path: where the verdict is PASS with all
//   four dims above their floor, not INCOMPLETE, and no contested finding, the orchestrator treats
//   this workflow's harmonic-mean verdict as the binding sync-phase verdict and does NOT spawn the
//   cold `sync-auditor` subagent. The cold auditor remains the FALLBACK verdict owner for the
//   failure modes (INCOMPLETE / dim-0 / contested finding)...

$ sed -n '47,51p' .claude/workflows/sync-audit-4dim.js
export const meta = {
  name: 'sync-audit-4dim',
  description: 'Sync-phase 4-dimension quality read (Functionality/Security/Craft/Consistency) — parallel read-only judges + in-script harmonic-mean verdict; execution vehicle, NOT the binding sync-auditor verdict owner',

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt302\b' --oneline
(no output)
```

**Baseline-attribution.** File read directly from this worktree's live `.claude/` tree (not the template source) at HEAD `6165f9f5e`, matching the card's own citation of local line numbers (":53", ":4-10"). Commit-grep against both pinned SHAs.

**Gaps.** Did not check `internal/runtime.FourDimVerdict.IsBinding()` (the header's cited "mechanical predicate") to verify the fallback conditions (INCOMPLETE / dim-0 / contested) are actually observable in code, as scope item (2) requests. Did not check `sync.md:81` or `sync-auditor.md` for alignment (scope item (3)). Did not check whether `internal/template/templates/.claude/workflows/sync-audit-4dim.js` (the template source, distinct from this local copy) carries the same contradiction or has already been fixed there — only the local copy was read.

**Residual-risk.** If the template source has already been corrected and only the local copy is stale, `make build` + re-sync would resolve this without new authoring work — a different disposition than a fresh code/doc fix. This was not checked.

**Proposed disposition.** `keep` — rests on: direct verbatim read of both contradicting passages in the same live file, plus independent corroboration that the "NOT binding" phrasing is externally visible (this session's own skill listing).

**Overlap candidates.** None observed among in-scope ids (SPEC-AUDIT-SNAPSHOT-001 is named but no in-scope card id references it directly in the visible batch text).

---

### t304

**Premise (one sentence).** Six of the 55 package paths cited across `.moai/project/codemaps/*.md` (`internal/design`, `internal/evaluator`, `internal/factory`, `internal/migrate`, `internal/research`, `internal/state`) name packages that do not exist in the tree, and this was already true at the codemaps' own stamped-baseline commit (not new drift).

**Premise verdict.** `holds` — confirmed directly. All six named paths were checked individually and none exist in the current worktree tree:
```
$ ls -d internal/design internal/evaluator internal/factory internal/migrate internal/research internal/state
ls: internal/design: No such file or directory
ls: internal/evaluator: No such file or directory
ls: internal/factory: No such file or directory
ls: internal/migrate: No such file or directory
ls: internal/research: No such file or directory
ls: internal/state: No such file or directory
```
And a grep confirms at least one codemaps file (`modules.md`) still cites them:
```
$ grep -rln "internal/design\|internal/evaluator\|internal/factory\|internal/migrate\|internal/research\|internal/state" .moai/project/codemaps/
.moai/project/codemaps/modules.md
```

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** The six nonexistent-package citations are still present in `.moai/project/codemaps/modules.md` on both pinned refs; no correction commit exists.

**Evidence.** (as above — directory-absence check + citation grep, both run against this worktree's HEAD)
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt304\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt304\b' --oneline
(no output)
```

**Baseline-attribution.** Directory-existence checks and codemaps grep run directly against this worktree's live tree at HEAD `6165f9f5e` (a reasonable proxy for the pinned develop tree since the card's own claim is that this defect predates and postdates the recent drift window — I did not separately check out the pinned develop tree to re-verify absence there, relying instead on the card's own git-ls-tree citation at the stamped baseline `a995e58fa`, which I did not independently re-run).

**Gaps.** Did not independently re-run `git ls-tree -d a995e58fa -- internal/design ...` (the card's own cited baseline-verification command) — I verified only against the current worktree tree, not the historical baseline commit. Did not identify which specific line, in which of the (possibly multiple) codemaps files, cites each of the six paths — only confirmed `modules.md` as one hit. Did not check the other 49 (of 55) cited paths for accuracy.

**Residual-risk.** If the current worktree tree differs from the stamped baseline in a way that removed these six packages AFTER the codemaps were generated (rather than them being absent from the start, as the card claims), my current-tree-only check would still show "absent" but for a different reason than the card's history claim — this would not change the practical disposition (codemaps are still wrong) but would change the "not new drift" framing.

**Proposed disposition.** `keep` — rests on: direct confirmation that all six named packages are absent from the tree and at least one codemaps file still cites them.

**Overlap candidates.** t291 (named directly by the card as "직교" — orthogonal but related, SPEC-STAMP-REACHABILITY-001) — not in `inscope-all.txt`. No in-scope id overlap observed.

---

### t305

**Premise (one sentence).** The statusline warm-render path spends ~93% of its ~236ms wall time on 5 serialized `git` subprocess spawns (measured in `.moai/reports/t215/profiling.md`), and two specific, quantified optimizations (deduping a repeated `rev-parse` call, and trimming the status/rev-list spawn set) could recover 40-77% of that time — this card asks to re-measure and apply them in the current tree.

**Premise verdict.** `holds` — the cited profiling report exists and its figures match the card's claims exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

**Claim.** `.moai/reports/t215/profiling.md` substantiates every specific number the card cites; no commit yet applies the two optimization candidates (deduped `rev-parse`, reduced status/rev-list spawn count) on either pinned ref.

**Evidence.**
```
$ grep -n "93%\|236ms\|spawn\|rev-parse" .moai/reports/t215/profiling.md
7:The statusline warm-path wall time is **~236 ms per render on this machine/tree**, and **≈100% of it is mechanically attributed**: ~93% is the five serialized git subprocess spawns on the render path, ~7% is Go process boot...
51:git rev-parse --git-dir:  median=29.5ms  ← NewRepository spawn 1/2 (manager.go:43)
52:git rev-parse --show-toplevel: median=29.1ms  ← NewRepository spawn 2/2 (manager.go:48)
87:| Builder init: 2× `git rev-parse` | 58.7 ms | ~25% |
88:| Git status collection: 3 spawns (`symbolic-ref`, `status --porcelain`, upstream `rev-list`) | 161.8 ms | ~68% |

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt305\b' --oneline
(no output)
```
The card's two quantified follow-up candidates match the report's own numbers: "-58.7ms" (Builder init dedup) = the report's `58.7 ms / ~25%` row; "-36~-123ms" (status/rev-list trim) is within the report's `161.8 ms / ~68%` row's range.

**Baseline-attribution.** `.moai/reports/t215/profiling.md` read at worktree HEAD `6165f9f5e`. Commit-grep against both pinned SHAs.

**Gaps.** Did not re-run the profiling benchmark (`internal/statusline/profile_bench_test.go`) myself to check whether the figures still hold in THIS tree/session's load window — the card itself explicitly flags this as required first-step work ("이 트리에서 먼저 재측정할 것"), which is why I did not attempt it (out of restraint-budget scope for a read-only sweep). Did not verify `internal/statusline/manager.go:43,48` (the cited `NewRepository` call sites) actually still contain the described duplicate `rev-parse` calls — only the profiling-report citation was read, not the current source.

**Residual-risk.** If `manager.go` has already been partially refactored since the t215 report was written, the specific line numbers/duplication the card describes might have shifted or already been partly addressed — this sweep did not check the current state of `manager.go` itself, only the historical report.

**Proposed disposition.** `keep` — rests on: the profiling report's own numbers substantiate every figure the card cites, and no commit matching this card exists on either pinned ref.

**Overlap candidates.** None in-scope (t215 and t211 are named by the card but neither is in `inscope-all.txt`).
# t332 card sweep — batch 5

Cards: t313 t315 t319 t320 t323 t324 t325 t327 t329 t337  (10 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t313

**Premise (one sentence).** `EnterWorktree` uses `origin/HEAD` (which pointed at `main`) as the
implicit branch base, so every card worktree was built from `main` instead of the git-flow-mandated
`develop`, and the fix needs both a durable branch-base config surface and a `moai web` UI to set
it.

**Premise verdict.** `holds` — the underlying defect (base-branch mismatch, lead's emergency
`git remote set-head` fix) is the documented trigger for SPEC-WORKTREE-BASEREF-001, which landed
(see Landing verdict). The card's own "미검증" callout (whether Claude Code's `fresh` mode actually
reads `origin/HEAD`) is exactly what the landed SPEC's implementation (`git_strategy.worktree_base_branch`
config key + the hook that aligns `origin/HEAD`) was built to settle from first principles rather than
assumption — I did not re-verify that specific runtime behavior myself; see Gaps.

**Landing verdict.** `landed`
- commit: `62ff3c2e6` (merge(WT-worktree-baseref): integrate card t313 — configurable card-worktree
  base branch (SPEC-WORKTREE-BASEREF-001))
- pinned ref: `ee50984abe4f11ac337382b48a26328f091e200a`
- `--is-ancestor` exit: 0
- branch + tip (in-flight only): — (worktree `.claude/worktrees/t313` @ `3fd8b5072` still exists
  but its tip commit is itself one of the 11 commits absorbed by the merge — the worktree was simply
  never disposed post-landing)

**Claim.** t313 is fully landed on `develop` via SPEC-WORKTREE-BASEREF-001 (11 commits: plan → config
schema key → hook alignment → doctor diagnostic → `git worktree add` base-branch wiring → `moai web`
free-text field → docs → sync/backfill → merge → post-merge doctor restamp). The `contains t295`
relation the card records is consistent with the merged SPEC's scope (base-branch-aware worktree
creation covers the "checkout an existing branch" path t295 names).

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt313\b' --oneline
f5a834fef fix(cli): restamp doctor golden pass count after the t313 merge
62ff3c2e6 merge(WT-worktree-baseref): integrate card t313 — configurable card-worktree base branch (SPEC-WORKTREE-BASEREF-001)
3fd8b5072 chore(SPEC-WORKTREE-BASEREF-001): backfill sync_commit_sha
b0d179de1 docs(SPEC-WORKTREE-BASEREF-001): sync-phase artifacts and 3-phase close
92102de1e docs(SPEC-WORKTREE-BASEREF-001): run-phase evidence and verdict (t313)
7d46e69c9 docs(worktree): document the stored card-worktree base branch and its two consumers
80d9e7e5b feat(web): expose worktree_base_branch in the console as a free-text field
c59e74232 feat(cli): add the Worktree Base Branch doctor diagnostic
97aef573d feat(cli): pass the configured base branch to git worktree add
26cc9ba90 feat(hook): align refs/remotes/origin/HEAD from the configured worktree base branch
a9c61cf56 feat(config): add git_strategy.worktree_base_branch schema key and neutral default
e717133cb feat(SPEC-WORKTREE-BASEREF-001): plan-phase artifacts (Tier M, 3 artifacts)

$ git merge-base --is-ancestor 62ff3c2e6 ee50984abe4f11ac337382b48a26328f091e200a; echo $?
0
```

**Baseline-attribution.** Measured against `origin/develop` pinned SHA
`ee50984abe4f11ac337382b48a26328f091e200a`, in this run.

**Gaps.** I did not re-run `EnterWorktree` end-to-end to confirm the landed implementation actually
produces a `develop`-based worktree at runtime — the card's own "실제로 만들어 실측할 것" instruction
was addressed inside the SPEC's own run-phase evidence (`92102de1e`), which I did not open in full. I
also did not check whether the `moai web` free-text field (`80d9e7e5b`) is wired end-to-end to a
consumer, or whether it is only stored.

**Residual-risk.** The still-present, undisposed `.claude/worktrees/t313` worktree is stale
housekeeping, not a functional risk — its tip is already an ancestor of the merge. If the `moai web`
field from `80d9e7e5b` is store-only (no consumer), a related but separate defect could remain
open under a different card.

**Proposed disposition.** `already-landed` — rests on the `--is-ancestor` exit 0 evidence above.

**Overlap candidates.** t319 (same interview-schema/config-surface class — t319's card text
explicitly notes "카드 t313 이 같은 표면에 새 항목을 얹는다"). No other in-scope id references t313
in its own text within this batch.

---

### t315

**Premise (one sentence).** t303's SPEC-SYNC-STRATEGY-KEY-001 audit left two carry-forward defects
(D6: a v3.3.0-scoped fallback-sentinel removal, D7a: a v3.2.0 release-notes obligation for the
`main_direct` → `github-flow` default flip) that have no owning card, so this card exists to hold
them until the right release-prep moment.

**Premise verdict.** `holds` — both carry-forwards are still open in the landed t303 sync commit's
own text, which explicitly assigns them to "card t315" by name.

**Landing verdict.** `not-landed`
- commit: — (no delivering commit found; two commits *mention* t315 by name as a forward pointer,
  neither delivers it)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): — (no worktree named t315 in `00-worktree-list.txt`)

**Claim.** t315 has not been worked. Both mentions found are the *origin* of the carry-forward
(t351's AC wording fix records "카드 t315" as the future remover of the sentinel; t303's own
terminal-transition commit says "the two carry-forwards D6/D7a remain card t315's"), not evidence of
execution.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt315\b' --oneline
60a6b2b97 docs(spec): refine AC-SYK-012.1 so it stops contradicting AC-SYK-003 (t351)
ed68889e3 docs(SPEC-SYNC-STRATEGY-KEY-001): terminal transition implemented -> completed (t303)

$ git show -s --format=%B 60a6b2b97 | grep -n t315
Also records in D.5 that the raw-count clause must move from 1 to 0 when v3.3.0
removes the sentinel (card t315), and that AC-SYK-003 becomes obsolete then.

$ git show -s --format=%B ed68889e3 | grep -n t315
Open debt is unchanged: OBS-2 and OBS-3 stay open (OBS-3 is card t333's
trigger axis), and the two carry-forwards D6/D7a remain card t315's.

$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt315\b' --oneline
(no output)
```

**Baseline-attribution.** Measured against both pinned SHAs, in this run.

**Gaps.** I did not open t303's full verdict.md to re-confirm the D6/D7a text verbatim beyond what
the two commit messages already quote — the two mentions found are internally consistent, so I did
not treat this as a gap requiring a third source.

**Residual-risk.** None specific to landing status — this is a not-yet-started card, correctly
queued.

**Proposed disposition.** `keep` — two concrete carry-forward obligations, each with a clear
originating SPEC and defect id, not yet actioned.

**Overlap candidates.** none observed in-batch. Cross-batch: t303 (origin SPEC, not in-scope list —
already closed) and t351/t333 mentioned by name inside the card text but neither is in
`inscope-all.txt`.

---

### t319

**Premise (one sentence).** `tab_schema.json` (the interview schema for `moai-workflow-project`) has
no pointer anywhere telling a consumer to read it, so it may be a dead file whose internal
correctness nobody checks.

**Premise verdict.** `holds` — I re-ran the card's own claimed scan and got the same counts.

**Landing verdict.** `not-landed`
- commit: — (no delivering commit)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t319)

**Claim.** The file exists at two mirrored locations, has exactly one Go-code reference (a
neutrality test that does not consume it as a schema), and zero references from its owning skill's
`SKILL.md`.

**Evidence.**
```
$ find . -iname "tab_schema.json" -not -path "*/node_modules/*"
./.claude/skills/moai-workflow-project/schemas/tab_schema.json
./internal/template/templates/.claude/skills/moai-workflow-project/schemas/tab_schema.json

$ grep -rln "tab_schema" --include="*.go" internal/
internal/template/internal_content_leak_test.go

$ grep -rln "tab_schema" .claude/skills/moai-workflow-project/
(no output)

$ grep -n "schema" .claude/skills/moai-workflow-project/SKILL.md
106:See [configuration schema and language fields](references/configuration.md) for full field reference and supported language metadata.
195:4. Validate rendered output against schema or existing conventions

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt319\b' --oneline
(no output)
```

**Baseline-attribution.** Scoped scan: `internal/` for Go references, `.claude/skills/moai-workflow-project/`
for skill-body references, both against worktree HEAD `6165f9f5e`.

**Gaps.** I did not scan `.claude/agents/` or other skills outside `moai-workflow-project` for a
reference (the card's own scan claims 10 non-consumer references across SPEC docs/reports/manifest
hashes — I did not enumerate those 10 myself, only reproduced the two counts that matter for the
holds/falsified decision). I also did not determine whether an *agent* (not a skill/code file) is
told out-of-band (e.g. in its spawn prompt) to read this file.

**Residual-risk.** If some agent's runtime prompt (not visible to a static grep) does reference this
file, the "orphan" framing would be wrong even though the static evidence holds.

**Proposed disposition.** `needs-operator-decision` — the card itself frames this as needing a
decision (retire the file vs. add a pointer), consistent with the file's actual orphan status
observed here.

**Overlap candidates.** t313 (adds a new entry to a config surface the card describes as sharing the
same "표면"). t316 is named in the card text (the key-mismatch defect this orphan status may explain)
but is not in `inscope-all.txt`.

---

### t320

**Premise (one sentence).** `moai integration release` returns an `ERROR` with an empty message body
when the calling session does not hold the lock (e.g. after being evicted by another lane's
`--force acquire`), leaving the caller unable to read why it failed.

**Premise verdict.** `falsified` — the sentinel error for exactly this "not held" case carries a
non-empty message, and has since the feature's original commit, predating the 2026-08-27
observation.

```
$ grep -n "ErrIntegrationLockNotHeld\|func ReleaseIntegrationLock" internal/kanban/integration_lock.go
52:var ErrIntegrationLockNotHeld = errors.New("no release integration window is held")
215:func ReleaseIntegrationLock(projectRoot, sessionID string, force bool) (released *IntegrationLock, err error) {

$ git show b2ad9158c -- internal/kanban/integration_lock.go | grep -n "ErrIntegrationLockNotHeld\|no release"
113:+// ErrIntegrationLockNotHeld is returned by ReleaseIntegrationLock when no
117:+var ErrIntegrationLockNotHeld = errors.New("no release integration window is held")
258:+		return nil, ErrIntegrationLockNotHeld

$ git show 3f3465369 -- internal/kanban/integration_lock.go | grep -n "ErrIntegrationLockNotHeld\|no release"
(no output — the t298 M2/M3 fix did not touch this sentinel)

$ sed -n '223,244p' internal/cli/integration.go
(RunE returns `err` directly on the not-held path; cobra's default error
printer renders err.Error() — "no release integration window is held" — not
an empty string)
```
The card's own hypothesized cause (eviction → not-held → empty-message error) maps directly onto
this code path, and that path's message is not empty in this tree, nor was it ever empty since the
lock feature's inception commit `b2ad9158c`.

**Landing verdict.** `not-landed`
- commit: — (no delivering commit — expected, since this is a fresh defect report, not a claim of a
  prior fix)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t320)

**Claim.** As currently understood and sourced, the card's premise does not match the code in this
tree. Either (a) the lane's original observation came from a different/older binary (v3.1.2 per the
worker instructions' banned-column caveat, which is plausible since CLI binaries lag source), or (b)
the empty message came from a code path other than the hypothesized eviction/not-held one that I did
not examine (e.g. a `--force` release path, or an unhandled panic/recover swallowing output).

**Evidence.** See the falsifying commands above, plus:
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt320\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt320\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/kanban/integration_lock.go` and `internal/cli/integration.go` at
worktree HEAD `6165f9f5e`; the two cited historical commits (`b2ad9158c`, `3f3465369`) via `git show`.

**Gaps.** I did not check the `--force` release branch of `ReleaseIntegrationLock` for a separate
empty-message path, nor did I check whether the actually-*installed* `moai` binary (v3.1.2, per the
worker instructions) matches this source tree's `integration_lock.go` — the lane's observation was
against a running binary, not this source. I also did not check `internal/cli/fang.go`'s error
rendering wrapper for any message-stripping behavior under specific flag combinations.

**Residual-risk.** The card's *diagnosis* may be wrong while its *observation* (an empty ERROR
message was genuinely seen) stays true — meaning a real defect could exist on a code path this scan
did not reach. A `falsified` verdict here is about the stated cause, not a claim that no bug exists.

**Proposed disposition.** `needs-operator-decision` — the falsified cause suggests re-scoping rather
than dropping outright: either narrow the card to "confirm against the exact binary version used" or
re-open it as "audit every ReleaseIntegrationLock error path for a possible empty-message case."

**Overlap candidates.** none observed in-batch. Cross-batch: t298 (the SPEC this card explicitly
declines to fold into) is not in `inscope-all.txt` (already closed per memory).

---

### t323

**Premise (one sentence).** The catalog-hash integrity mechanism hashes only the single `SKILL.md`
(or `skill.md`) file inside a skill directory entry, so changes to any other file in that directory
(`schemas/`, `references/`, `scripts/`) are shipped via `//go:embed` but never move the catalog hash
that is supposed to attest integrity.

**Premise verdict.** `holds` — the hashing function's directory-branch logic matches the claim
exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t323)

**Claim.** `resolveHashSourcePath` in `internal/template/scripts/gen-catalog-hashes.go` resolves a
directory catalog entry to exactly one file (`SKILL.md` first, `skill.md` fallback) and hashes only
that file; no code path aggregates the rest of the directory tree into the hash.

**Evidence.**
```
$ sed -n '100,134p' internal/template/scripts/gen-catalog-hashes.go
... (directory branch: `for _, candidate := range []string{"SKILL.md", "skill.md"}` returns the
     first found path; non-directory branch hashes the file directly) ...

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt323\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/template/scripts/gen-catalog-hashes.go` at worktree HEAD
`6165f9f5e`.

**Gaps.** I did not check whether the catalog's `path` field for a directory-shaped entry can ever
point at a non-skill directory with a different resolution rule, nor did I check the consumer side
(what reads and trusts the catalog hash at runtime) to see how severe the blind spot actually is in
practice.

**Residual-risk.** If some other mechanism (not found by this scoped read) separately verifies
non-SKILL.md files, the severity framing in the card could be overstated even though the mechanical
claim is correct.

**Proposed disposition.** `keep` — mechanical claim reproduced exactly as stated; the card frames the
choice (a)/(b)/(c) as needing an operator decision, which I did not adjudicate.

**Overlap candidates.** t317 is named in the card text ("같은 병의 다른 증상") but is not in
`inscope-all.txt`. No other in-scope batch-5 card touches catalog hashing.

---

### t324

**Premise (one sentence).** `develop` currently has no branch protection at all, so the git-flow
model's CI status checks run only after lanes have already pushed directly to `develop` (detection,
not prevention), and re-enabling required-status protection needs a co-design with the no-card-PR
lane-push model rather than a simple toggle.

**Premise verdict.** `holds` — the cited 404 reproduces exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t324)

**Claim.** `develop` branch protection is confirmed absent right now, matching the card's cited
evidence verbatim.

**Evidence.**
```
$ gh api repos/modu-ai/moai-adk/branches/develop/protection
{"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection#get-branch-protection","status":"404"}
gh: Branch not protected (HTTP 404)

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt324\b' --oneline
(no output)
```

**Baseline-attribution.** Live GitHub API state for `modu-ai/moai-adk`, queried in this run (not a
tree-scoped measurement — branch protection is GitHub-side config, not a file).

**Gaps.** I did not check `main`'s current protection settings for comparison, and did not enumerate
which specific CI workflows currently exist as candidate required-status checks (the card names this
as an open design question, which I did not attempt to resolve).

**Residual-risk.** GitHub-side settings can change between this read and any future action on this
card — this is a live-state fact, not a tree-pinned one (§2.1 moving-ref concern: the 404 is current
as of this run, not a durable baseline).

**Proposed disposition.** `keep` — operator-flagged decision card (`[운영자 판정 2026-08-27]` prefix
in the card text itself), premise reproduces exactly, design questions remain genuinely open.

**Overlap candidates.** t325 (both concern the integrity of the `develop`→`main` promotion path
under the git-flow no-card-PR model; t324 is about protecting `develop` itself, t325 is about a
workflow that could bypass the `main`-entry gate this protection sits alongside).

---

### t325

**Premise (one sentence).** `spec-status-auto-sync.yml` fires on any `pull_request: closed` event
with no base-branch filter and pushes to `main` with `contents: write`, so a PR targeting `develop`
being closed could trigger a push to `main` — a hypothesis, not yet observed firing.

**Premise verdict.** `holds` (as an unverified-but-structurally-confirmed hypothesis, matching the
card's own framing — the card explicitly labels this "미검증 가설, 검증 필요" and I did not push the
verification further than the card already had).

**Landing verdict.** `not-landed`
- commit: — (one commit *mentions* t325 as a still-unverified hypothesis, does not deliver it)
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t325)

**Claim.** The workflow trigger and push target match the card's description exactly: no `branches:`
filter under `on.pull_request`, and `git push origin main` present in the job body.

**Evidence.**
```
$ sed -n '1,14p' .github/workflows/spec-status-auto-sync.yml
name: SPEC Status Auto-Sync
on:
  pull_request:
    types: [closed]
permissions:
  contents: write    # git push origin main (line 107)
  issues: write      # gh issue create fallback (line 95-99)

$ grep -n "git push origin main\|branches:" .github/workflows/spec-status-auto-sync.yml
9:# (contents: read) 가 line 107 git push origin main 을 403 으로 실패시키던
12:  contents: write    # git push origin main (line 107)
16:# commit + git push origin main; cancelling mid-push would leave the
123:            git push origin main

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt325\b' --oneline
d723ca29e docs(t314): separate the three verdict gaps, mark the third as awaiting observation

$ git show -s --format=%B d723ca29e | grep -n t325
Now stated as three items with distinct character: unverified (release-PR
filter, no open PR to measure), unverified hypothesis (spec-status-auto-sync,
card t325), and awaiting observation (first firing of spec-lint /
```

**Baseline-attribution.** `.github/workflows/spec-status-auto-sync.yml` at worktree HEAD `6165f9f5e`;
the one mentioning commit via `git show` against pinned develop.

**Gaps.** I did not trace the job body past line 123 to confirm which ref is actually checked out
before the push (the card itself says this is the exact next step — "발화 시 어느 ref 로 push 하는지
코드 경로를 따라갈 것" — and I stopped at reproducing the structural facts already in the card, per
the restraint instruction).

**Residual-risk.** If `actions/checkout@v7` with `fetch-depth: 0` in this job checks out the
PR-closed event's base ref (not always `main`) before the commit+push steps, the actual runtime
behavior could differ from what the trigger/permission facts alone suggest — this needs the
step-by-step trace the card calls for, not yet done here or by the referenced commit.

**Proposed disposition.** `keep` — hypothesis remains open and structurally plausible; card correctly
declines to overstate to a confirmed finding.

**Overlap candidates.** t324 (both touch the integrity of the `develop`→`main` boundary under
git-flow). t314 is named in the originating commit but is not in `inscope-all.txt`.

---

### t327

**Premise (one sentence).** `treeDirty` (the function deciding whether a stamp anchors to a commit or
to a dirty-tree fingerprint) checks raw `git status --porcelain` dirtiness without applying the
"described-worthy" filter that the SPEC-GRAPH-FRESHNESS-CADENCE-001 (t322) audit found elsewhere, so
a dirty-but-not-described-worthy tree (e.g. only `testdata/` changed) still gets denied a `--commit`
anchor.

**Premise verdict.** `holds`, with one citation correction: the card's cited location
(`internal/config/provenance.go:201`) does not exist — `internal/config/provenance.go` is only 104
lines and defines no `treeDirty` symbol. The actual function is `internal/mx/provenance.go:225`, and
its behavior matches the mechanism the card describes exactly (raw `git status --porcelain` dirty
check, no described-worthy filtering).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t327)

**Claim.** The mechanism claim is correct; the file:line citation in the card is stale/wrong and
should be corrected to `internal/mx/provenance.go:223-227` before this card is worked.

**Evidence.**
```
$ wc -l internal/config/provenance.go
104 internal/config/provenance.go
$ grep -n "treeDirty" internal/config/provenance.go
(no output)

$ grep -rn "treeDirty" internal/mx/provenance.go
184:// that, and treeDirty's emptiness test depends on it (CR round-2 3855149357).
223:// treeDirty reports whether any file under the given repo-relative roots has
225:func treeDirty(root string, roots []string) bool {
249:	if treeDirty(projectRoot, describedRoots) {

$ sed -n '223,227p' internal/mx/provenance.go
// treeDirty reports whether any file under the given repo-relative roots has
// uncommitted changes (staged, unstaged, or untracked) versus HEAD.
func treeDirty(root string, roots []string) bool {
	args := append([]string{"status", "--porcelain", "--"}, roots...)
	return gitOut(root, args...) != ""
}

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt327\b' --oneline
(no output)
```

**Baseline-attribution.** `internal/config/provenance.go` and `internal/mx/provenance.go` at worktree
HEAD `6165f9f5e`.

**Gaps.** I did not read SPEC-GRAPH-FRESHNESS-CADENCE-001's D.1/E sections that the card says already
document the deferral and remediation direction — I verified only the mechanism claim, not the SPEC
cross-reference's exact wording.

**Residual-risk.** If the card is worked using the stale `internal/config/provenance.go:201` citation
without first re-locating the function, the implementer could edit the wrong file or waste time
searching a 104-line file for a symbol that isn't there.

**Proposed disposition.** `keep`, with a note attached before dispatch: correct the location citation
to `internal/mx/provenance.go:223-227`.

**Overlap candidates.** t322 (the originating SPEC that deferred this scope, not in `inscope-all.txt`
— already closed/landed per project memory). No other in-scope batch-5 card touches this function.

---

### t329

**Premise (one sentence).** `.moai/docs/git-workflow-doctrine.md` §18.12 cites `internal/bodp/relatedness.go`
by function and file name for its BODP decision matrix, but that package does not exist in the tree —
the doctrine's actual SSOT for the default base-branch recommendation is a different, live file
(`branch-origin-protocol.md`), and this card's scope is narrowly to redirect the citation, not to
change the `origin/main` default itself.

**Premise verdict.** `holds` — reproduced exactly.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip: — (no worktree named t329)

**Claim.** `internal/bodp/` has zero tracked files on pinned `develop`; the doctrine's actual live
SSOT (`branch-origin-protocol.md:26`) carries the `origin/main` default-recommendation language the
card quotes.

**Evidence.**
```
$ git ls-tree -r --name-only ee50984abe4f11ac337382b48a26328f091e200a -- internal/bodp/ | wc -l
0

$ grep -n "relatedness.go\|internal/bodp" .moai/docs/git-workflow-doctrine.md
402:`internal/bodp/relatedness.go` `Check()` 함수가 다음 3개 시그널을 평가한다:
412:`internal/bodp/relatedness.go` `applyMatrix()` — SignalB 우선순위 dominates A/C:

$ grep -n "When no signal fires" .claude/rules/moai/development/branch-origin-protocol.md
26:- [ZONE:Evolvable] [HARD] The recommended base MUST be derived from the signals below, not assumed. When no signal fires, the recommendation is `origin/main` — team-safe, because it reflects the latest merged state rather than whatever the local checkout happens to hold.

$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt329\b' --oneline
(no output)
```

**Baseline-attribution.** `git ls-tree` against pinned develop `ee50984abe4f11ac337382b48a26328f091e200a`;
`.moai/docs/git-workflow-doctrine.md` and `.claude/rules/moai/development/branch-origin-protocol.md`
at worktree HEAD `6165f9f5e`.

**Gaps.** I did not search the rest of the doctrine document tree for other instances of the "same
class" the card flags at the end ("이 문서군이 코드 경로를 인용하는 다른 지점들") — that is explicitly
left open by the card itself as follow-on scope.

**Residual-risk.** None specific — this is a narrowly-scoped, cleanly-reproduced documentation-drift
finding with an explicit scope boundary already stated in the card.

**Proposed disposition.** `keep` — reproduces exactly, scope is already well-bounded by the card
itself.

**Overlap candidates.** none observed in-batch (no other batch-5 card touches
`git-workflow-doctrine.md` or `internal/bodp`).

---

### t337

**Premise (one sentence).** On Windows, the anchor guard's `isProcessAlive` unconditionally returns
`true`, so a stamped-but-actually-dead session PID is never corrected back to reclaimable — the one
code path that still contradicts the declared TREAT-AS-LIVE invariant, and it predates and was not
touched by t298's fix.

**Premise verdict.** `holds` — reproduced exactly, including the "diff 0" claim against t298.

**Landing verdict.** `in-flight-unlanded`
- commit: — (no delivering commit on pinned develop)
- pinned ref: —
- `--is-ancestor` exit: 1 (tip `c72a517c3` is NOT an ancestor of pinned develop)
- branch + tip (in-flight only): `WT-windows-stamp-liveness` @ `c72a517c3` (from
  `.moai/reports/t332/00-worktree-list.txt`, worktree `.claude/worktrees/t337`)

**Claim.** A live worktree exists for this card with unmerged work; the underlying code claim
(unconditional `true` on Windows, not touched by t298) is independently verified as still accurate on
`develop`.

**Evidence.**
```
$ cat internal/session/anchor_pid_windows.go
//go:build windows
...
func isProcessAlive(pid int) bool {
	_ = pid
	return true
}
...
func probeProcessLiveness(pid int) (alive bool, determined bool) {
	_ = pid
	return false, false
}

$ grep -rn "isProcessAlive\|sessionPIDFromEnv" internal/session/*.go | grep -v _test.go
internal/session/anchor.go:80:			alive := e.PID > 0 && isProcessAlive(e.PID)
internal/session/session_pid.go:54:	pidIsAlive              = isProcessAlive
internal/session/session_pid.go:74:	if pid, ok := sessionPIDFromEnv(os.Getenv(config.EnvMoaiSessionPID)); ok {

$ git log ee50984abe4f11ac337382b48a26328f091e200a --oneline -- internal/session/anchor_pid_windows.go
8ff3e0823 fix(worktree): repair the worktree reaper — three-valued merge detection, lock-aware anchor guard, clean --stale inventory (t209) (#1638)
cf749fafe fix(worktree): guard done against live sessions anchored in the target tree

$ git merge-base --is-ancestor c72a517c3 ee50984abe4f11ac337382b48a26328f091e200a; echo $?
1
```
Neither commit touching `anchor_pid_windows.go` mentions or was authored by t298 — consistent with
the card's "diff 0건" attribution claim (t298's own commits, checked separately in the t320 entry
above, do not appear in this file's history either).

**Baseline-attribution.** `internal/session/anchor_pid_windows.go`, `anchor.go`, `session_pid.go` at
worktree HEAD `6165f9f5e`; file history and worktree-liveness check against pinned develop
`ee50984abe4f11ac337382b48a26328f091e200a`.

**Gaps.** I did not open the live worktree (`.claude/worktrees/t337`) to inspect what work is already
in progress there, per the restraint instruction and because this is a read-only sweep of the primary
tree's cards, not an audit of in-flight branches. I also did not verify the `probeProcessLiveness`
honest-undetermined path is actually wired to replace `isProcessAlive` anywhere (the card frames this
as the fix direction, not yet confirmed done).

**Residual-risk.** Since work is already underway on a live branch, this sweep's read-only verdict
could be stale by the time it's read — the in-flight branch may already contain a fix for exactly
this gap. The disposition proposal below accounts for that explicitly.

**Proposed disposition.** `already-landed` is NOT proposed (verified not an ancestor of develop, exit
1); given a live worktree already exists and the premise independently holds, `needs-operator-decision`
is proposed only to confirm whether the in-flight branch should simply be pushed to completion rather
than treated as a queue item — this rests on the `--is-ancestor` exit 1 evidence plus the existing
worktree entry.

**Overlap candidates.** none observed in-batch (no other batch-5 card touches
`internal/session/anchor_pid_windows.go`). Cross-batch: t298 is named directly in the card text but
is not in `inscope-all.txt` (already landed/closed per project memory).
# t332 card sweep — batch 6

Cards: t339 t344 t345 t347 t348 t353 t359 t360 t361 t363 (10 entries)
Worktree HEAD: 6165f9f5e
Pinned develop: ee50984abe4f11ac337382b48a26328f091e200a
Pinned main:    48239c7dc7428c8751a04f6321887c2d36123884

### t339

**Premise (one sentence).** t317's plan.md/spec.md still carry three specific documentation
defects (D10/D11/D12) that were re-confirmed four times but never closed.

**Premise verdict.** `holds` — all three checked directly against the current tree.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt339\b'` against pinned develop and pinned main) returned no output.

**Claim.** The three debts the card names are still present verbatim in
`.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md` and `spec.md`, even though the SPEC itself
(SPEC-AGENT-EMIT-LINEAGE-001, card t317) has since been fully implemented and closed
(commits `742a9485d`..`3235aa08f`..`0ad4b52ba`, all ancestors of pinned develop). The doc debt
survived the SPEC's own close.

**Evidence.**
```
$ grep -n "B.1" .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md
13:### B.1 영향 파일 전수 열거 (추정 아님)
52:**다시 늘어나면 다시 판정한다.** ...
```
The B.1 table (lines 15-20) still lists exactly 5 rows (Makefile, doctor_<name>.go,
doctor_<name>_test.go, doctor.go, CLAUDE.local.md) — D10's claim confirmed.

```
$ git show --stat 742a9485d 6335b731b f3e5006ce | grep -E "^\s*(internal|Makefile|CLAUDE)"
 internal/cli/doctor.go                             |   4 +
 internal/cli/doctor_agentemit_embed.go             | 287 ++++++++++++++
 internal/cli/doctor_agentemit_embed_test.go        | 411 +++++++++++++++++++++
 internal/cli/testdata/doctor-dark.golden           |   5 +-
 internal/cli/testdata/doctor-light.golden          |   5 +-
 internal/cli/testdata/doctor-nocolor.golden        |   5 +-
 Makefile | 24 ++++++++++++++++++++++--
 ...
```
The actual run-phase touched three golden files (`doctor-dark.golden`, `doctor-light.golden`,
`doctor-nocolor.golden`) that B.1's enumeration never lists — this is the "골든 3본" D10 names.

```
$ grep -n "파일 1개" .moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md
18:| 2 | ... 이 리포의 doctor 항목은 파일 1개에 사는 것이 규약이다(...) | M1 |
23:**5건.** ... doctor 항목은 파일 1개 + 짝 테스트 1개로 살고 ...
```
D11 confirmed: the "lives in one file" claim is still there unmodified, and is contradicted by
the golden-file evidence just above (the item's footprint is body file + test file + 3 golden
files + a Makefile edit — more than "one file").

```
$ grep -n "count != 11" internal/template/agentemit/golden_test.go
284:	if count != 11 {
```
`spec.md:88` cites this assertion as `golden_test.go:285`; the actual line is 284 — D12
confirmed as a stale-by-one-line citation.

**Baseline-attribution.** All three checks run against worktree HEAD `6165f9f5e`
(`.moai/specs/SPEC-AGENT-EMIT-LINEAGE-001/plan.md`, `spec.md`,
`internal/template/agentemit/golden_test.go`, current tree). The commit-touched-file evidence is
from `git show --stat` on the three cited SPEC-AGENT-EMIT-LINEAGE-001 commits, which are
ancestors of pinned develop.

**Gaps.** Did not re-verify whether any OTHER debt beyond D10/D11/D12 has crept into the two
files since t317 closed. Did not check whether a different card already re-issued a doc-fix for
this SPEC (grep of `.moai/reports/` for a superseding fix was not run — bounded depth).

**Residual-risk.** The three items are stated by the card itself as "동작에 영향 없음" (no
behavioral impact) — so even if left open, they only mislead a future reader of a closed SPEC's
plan artifacts, not runtime behavior. If SPEC docs are treated as immutable historical record
post-close, "fixing" them may itself be an out-of-policy edit — that policy question is not
something this sweep resolves.

**Proposed disposition.** `keep` — all three named defects independently reconfirmed present at
HEAD; the fix is a small, scoped documentation edit to a closed SPEC's plan/spec artifacts. Rests
on the three verbatim greps above.

**Overlap candidates.** none observed — no other in-scope (batch or full list) card touches
SPEC-AGENT-EMIT-LINEAGE-001 or card t317's artifacts.

---

### t344

**Premise (one sentence).** SPEC-VERIFICATION-COMPLETENESS-001's prediction ledger (VC-2/4/5/6)
measures "audit-flag count = 0" as success, which is a metric that rises precisely when the rule
works well at the audit layer — the polarity is backwards.

**Premise verdict.** `holds` — the flip is real (verified against the ledger text) and the
corrective text the card describes is already inline in the same file, but no decision has been
made on whether to generalize it.

**Landing verdict.** `not-landed`
- commit: `2db93496c` (mention only, see Claim)
- pinned ref: ee50984abe4f11ac337382b48a26328f091e200a
- `--is-ancestor` exit: 0 (of the mentioning commit; it does not deliver t344's own scope)
- branch + tip (in-flight only): —

**Claim.** t344's own scope-decision work (whether to generalize the ledger-format correction, or
treat it as already closed) has not been done. The only commit matching `\bt344\b` on pinned
develop is `2db93496c`, whose subject is "record which cards the follow-up candidates became" —
it names t344 as one of four cards issued from t241's candidate table (C1..C4 → t341..t344), it
does not perform t344's work.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt344\b' --oneline
2db93496c docs(t241): record which cards the follow-up candidates became (t241)
$ git show -s --format=%B 2db93496c
docs(t241): record which cards the follow-up candidates became (t241)

C1 through C4 were issued as t341, t342, t343 and t344; C5 remains unowned. ...
Also records the two conditions the operator attached: ... and
t344 must first judge whether the ledger defect is already closed by the §A.5
text this card wrote.
```
```
$ sed -n '105,125p' .moai/specs/SPEC-VERIFICATION-COMPLETENESS-001/plan.md
### §A.5 예측 장부 (Prediction Ledger — t241 주의 반영)
...
| VC-2 | 'mutant가 쓰여 AC 무효화' 감사 지적 0건 | ... | **false** — 신규 7건 ...
...
**장부 문면의 결함 (판정 중 발견, 다음 장부 저작 시 교정할 것).** VC-2·VC-4·VC-5·VC-6 은 예측을
"감사 지적 0건" 으로 적었다. 이 지표는 규칙이 감사 층에서 잘 작동할수록 올라간다 ...
다음 장부는 예측을 채택까지 살아남은 건수로 쓰고, 감사 지적 수는 반대 부호의 보조 지표로 둘 것.
```
The corrective is already recorded inline in §A.5's own plan.md prose — the exact instance-level
fix the card describes.

```
$ grep -rln "citation.count\|rule-citation\|citation_count" internal/ .moai/
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/progress.md   # unrelated (graph-freshness cadence)
```
No general ledger-authoring rule or convention file was found that has already absorbed this
correction as a repo-wide format rule (searched for a citation-count / ledger convention file;
one unrelated hit).

**Baseline-attribution.** Ledger text measured against
`.moai/specs/SPEC-VERIFICATION-COMPLETENESS-001/plan.md` at worktree HEAD `6165f9f5e`. Commit
`2db93496c` and its ancestry against pinned develop `ee50984ab...` measured directly.

**Gaps.** Did not check whether `moai-constitution.md` §Lessons Protocol or any other
always-loaded rule file already generalizes the "survival-to-adoption, not audit-flag-count"
principle outside this one SPEC's plan.md — only a targeted grep for a dedicated ledger/citation
tool was run, not a full-text search of the rules tree. Did not check t333's current status to
see whether its "표현 기대치" design already absorbed this constraint (card explicitly asks the
next actor to check this before starting).

**Residual-risk.** If a generalization already landed somewhere outside the paths checked, this
card's premise ("아직 안 닫힘") would be wrong even though the specific instance evidence above is
accurate — the scope-decision itself is exactly what's unresolved, which is the card's own stated
ask.

**Proposed disposition.** `needs-operator-decision` — rests on the fact that the card's own two
branch options (a: generalize, b: treat as closed) are both still open per the evidence above; a
sweep worker choosing between them would be making the operator's call.

**Overlap candidates.** t345 (same source — t241 verdict.md and lane-14/lane-17 findings; both
cards reference SPEC-VERIFICATION-COMPLETENESS-001 §A.5 and the "policy rule adoption is
unobserved" theme). No other in-scope card touches this SPEC.

---

### t345

**Premise (one sentence).** No mechanism exists to observe whether a policy-layer rule (as
opposed to a workflow, which has run history) is actually cited and applied, versus merely
existing as a file.

**Premise verdict.** `holds` — no dedicated observation tool for policy-rule citation/adoption was
found in the tree.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt345\b'` against pinned develop and pinned main) returned no output.

**Claim.** The card's central question — what observes a policy rule's real-world adoption — is
still unanswered; nothing in the tree currently computes a citation count, an audit-artifact
reference count, or an authoring-layer application trace for a rule file as a standing mechanism.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt345\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt345\b' --oneline
(no output)
$ grep -rln "verification-completeness" .moai/reports/ | wc -l
16
$ grep -rln "citation.count\|rule-citation\|citation_count" internal/ .moai/
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/progress.md
```
The 16 report-directory hits for `verification-completeness` show the rule IS being cited by
audit artifacts by hand (as t241's judgment relied on), but no standing tool computes this — the
one code hit for a citation-count concept is unrelated (graph-freshness cadence, not rule
adoption).

**Baseline-attribution.** Grep scans run against worktree HEAD `6165f9f5e`, scoped to
`.moai/reports/`, `internal/`, and `.moai/`.

**Gaps.** Did not fully read t333's design.md to confirm the "Out of Scope, 이름만" characterization
the card asserts about how t333 handled this candidate (C5) — that claim was taken from the card
text and not independently re-verified against t333's artifacts at HEAD (t333 is not in the
in-scope list for this sweep, so its current artifact state was not opened). Did not search for a
possible non-repo mechanism (e.g., an external log/analytics system) that might already answer
this.

**Residual-risk.** This is a design/research question, not a defect with a fixed verifiable
state — "holds" here means "the gap the card names still appears open," not that the card's
proposed three sub-decisions (a/b/c) have a single correct answer. A different reading of what
counts as "observation" could change the verdict.

**Proposed disposition.** `needs-operator-decision` — the card itself frames three sub-decisions
(what counts as observation, whether it discriminates audit-absorption from authoring-absorption,
whether it's cheaper than manual reading) that are genuinely open design choices, not facts a
sweep can settle.

**Overlap candidates.** t344 (shared source: t241 verdict.md's follow-up candidate table, and both
name the "감사 지적 0건" polarity trap). No other in-scope card touches this theme.

---

### t347

**Premise (one sentence).** t333's classification/status model should be split into a sub-SPEC
authored as a state table (not prose), because three plan-audit iterations on
SPEC-GUARD-LIVENESS-001 kept moving the same defect family (D2/D5/N2/T2/T4) without closing it.

**Premise verdict.** `holds` — the split happened and was authored as a state table, exactly as
asked; a residual open question (T2's contradiction) is tracked inside the new SPEC's own defects,
not left silently unresolved.

**Landing verdict.** `landed` (plan-phase scope of the card)
- commit: `37263c222`
- pinned ref: ee50984abe4f11ac337382b48a26328f091e200a
- `--is-ancestor` exit: 0
- branch + tip (in-flight only): —

**Claim.** SPEC-GUARD-STATE-MODEL-001 was created (card t347) as the split-off sub-SPEC, authored
as a state table per the card's [HARD] instruction, and its plan-phase iter-1 audit defects were
closed. Its `status:` frontmatter is still `draft` and §E.2/E.3/E.4 (run-phase evidence) are
explicitly `<pending run-phase>` — the card's specific ask (split + author as state table) is
plan-complete; the SPEC's own implementation is not yet run.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt347\b' --oneline
37263c222 feat(SPEC-GUARD-STATE-MODEL-001): card t347, state-table delivery column, and instance 7
0f27fa774 fix(SPEC-GUARD-STATE-MODEL-001): close all seven blocking and four optional iter-1 defects (card t347)
7489d4f86 fix(SPEC-GUARD-LIVENESS-001): close all six iter-4 blocking defects, both optional (card t333/t347)
558925d00 fix(SPEC-GUARD-LIVENESS-001): close iter-2 D9, D10, D11 and two twins (t333/t347)
$ git merge-base --is-ancestor 37263c222 ee50984abe4f11ac337382b48a26328f091e200a
(exit 0, no output)
```
```
$ grep -n "status:" .moai/specs/SPEC-GUARD-STATE-MODEL-001/spec.md
5:status: draft
$ tail -8 .moai/specs/SPEC-GUARD-STATE-MODEL-001/progress.md
## §E.2 Run-phase Evidence
_<pending run-phase>_
## §E.3 Run-phase Audit-Ready Signal
_<pending run-phase>_
## §E.4 Sync-phase Audit-Ready Signal
_<pending sync-phase>_
```
```
$ cat .moai/specs/SPEC-GUARD-STATE-MODEL-001/spec.md | head -14
id: SPEC-GUARD-STATE-MODEL-001
title: "Guard liveness state model: declare firing expectations, and decide every entry into
exactly one classification (card t347)"
...
tags: "guard, liveness, state-model, classification, manifest, cadence, totality, t347"
```

**Baseline-attribution.** Commit ancestry checked against pinned develop
`ee50984abe4f11ac337382b48a26328f091e200a`. SPEC status/progress fields read from
`.moai/specs/SPEC-GUARD-STATE-MODEL-001/{spec.md,progress.md}` at worktree HEAD `6165f9f5e`.

**Gaps.** Did not verify whether the SPEC has since progressed to run-phase in a live worktree not
covered by the two pinned refs (e.g., a worktree branch ahead of develop) — the worktree list
(`00-worktree-list.txt`) was not cross-checked for a SPEC-GUARD-STATE-MODEL-001-named branch. Did
not re-derive whether the "T2 모순" the card flags as unresolved is actually closed inside the new
SPEC's REQ set (only progress.md's summary was read, not the full spec.md REQ text).

**Residual-risk.** "Landed" here is scoped narrowly to the card's literal ask (split into a
state-table sub-SPEC); the SPEC's actual implementation (REQ-GDL-004/007/008/009 behavior) remains
unbuilt, so if the operator's intent for this card included seeing the guard behavior itself
change, that part is still open.

**Proposed disposition.** `already-landed` — rests on commit `37263c222`'s subject line and body
explicitly naming "card t347" and "state-table delivery column" as delivered, verified as an
ancestor of pinned develop.

**Overlap candidates.** none observed in the in-scope list — SPEC-GUARD-STATE-MODEL-001 and
SPEC-GUARD-LIVENESS-001 (card t333) are not themselves in `inscope-all.txt`.

---

### t348

**Premise (one sentence).** SPEC-AC-COUNT-DISCRIMINATOR-001's reserved-token AC-counting
convention (from t338) does not retroactively mark 34 existing citation files, so they keep being
counted as `live` (silently over-counted) until a new native-prefix-declaration syntax closes that
gap — a syntax this SPEC deliberately did not build.

**Premise verdict.** `holds` — corroborated near-verbatim against the completed SPEC's own plan.md
and design.md.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt348\b'` against pinned develop and pinned main) returned no output.

**Claim.** t348's own scope (the native-prefix-declaration syntax) has not been built. The SPEC it
follows up on, SPEC-AC-COUNT-DISCRIMINATOR-001 (t338), is `status: completed` and explicitly
records this exact gap as a named follow-up candidate it declined to build.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt348\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt348\b' --oneline
(no output)
$ grep -n "status:" .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/spec.md
5:status: completed
$ grep -n "34\|네이티브 접두사" .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/plan.md
93:**남는 것(정직하게 적는다)**: ... 토큰이 하나도 없는 순수 인용(현재 34건의 기본형)은 §3.2 의
판정에서 live 로 읽히므로 정지가 아니라 과다 계상으로 남는다. ... 그것은 정당한 다중 도메인
SPEC(AC-APO-* + AC-DCP-*)까지 정지시키므로 네이티브 접두사 선언이라는 새 기제를 부른다. 이 카드는
그 기제를 만들지 않는다 — 후속 카드 후보로만 기록한다(spec.md §6).
$ grep -n "네이티브 접두사 선언" .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/design.md
136:| 접두사 단위 애매성 판정(순수 인용까지 잡기) | ... 네이티브 접두사 선언이라는 새 문법이 필요하고
이 카드의 크기를 넘는다 — 후속 카드 후보(spec.md §6) |
```
The 34-file count, the "live not stopped" over-count mechanism, and the native-prefix-declaration
follow-up candidate are all named identically in the completed SPEC's own artifacts.

**Baseline-attribution.** All figures read from
`.moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/{spec.md,plan.md,design.md}` at worktree HEAD
`6165f9f5e` — the SPEC's own authored record, not re-derived by re-running the counter.

**Gaps.** Did not re-run the actual AC-counting tool/discriminator against the current tree to
confirm 34 is still accurate today (could have drifted since the SPEC closed as more citation
files were added or normalized) — the card itself warns not to cite 156 as a defect count without
re-measuring, and the same caution applies to 34. Did not check whether the 7 false-positive count
the card cites ("오탐 7건 실측") is independently verifiable at HEAD — taken from card text only.

**Residual-risk.** If citation files have been added/normalized since SPEC-AC-COUNT-DISCRIMINATOR-001
closed, the live 34-count could now differ from what's cited here, though the structural gap
(no native-prefix syntax exists) would still hold regardless of the exact count.

**Proposed disposition.** `keep` — rests on the completed SPEC's own plan.md/design.md explicitly
naming this exact follow-up candidate as out-of-scope and unbuilt.

**Overlap candidates.** none observed in the in-scope list — SPEC-AC-COUNT-DISCRIMINATOR-001
(card t338) is not itself in `inscope-all.txt`.

---

### t353

**Premise (one sentence).** REQ-MRG-010/AC-MRG-013 (the R4-form lint exclusion for
SPEC-MOVING-REF-GUARD-001) was deliberately deferred by operator decision because the R4 form has
zero observed occupants in the corpus, and should stay deferred until R4 is actually observed.

**Premise verdict.** `holds` — the deferral, its rationale, and its resume condition are recorded
verbatim in t342's verdict.md, and this card is that deferral's own follow-up (issued by the
operator, as the verdict explicitly notes none was issued from that lane).

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt353\b'` against pinned develop and pinned main) returned no output.
(Note: SPEC-MOVING-REF-GUARD-001 / t342 itself IS landed — its worktree tip `38f937a4f` is
confirmed an ancestor of pinned develop — but that landing covers only the non-deferred parts.)

**Claim.** The R4-form lint exclusion remains unimplemented, exactly as t342's verdict.md
recorded it deferred. No commit citing t353 exists on either pinned ref, and no implementation of
REQ-MRG-010/AC-MRG-013 was found.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt353\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt353\b' --oneline
(no output)
$ sed -n '106,124p' .moai/reports/t342/verdict.md
## Deferred — REQ-MRG-010 and AC-MRG-013 (Q0, option C)
Deferred by operator decision.
**What:** the R4-form lint exclusion, its acceptance criterion, and both counter-mutations ...
**Why:** §B.7 measured R4's reachable class as 0 of 42 candidate lines on two independent probes...
**Resume condition:** reconsider when the R4 form is actually observed in the corpus. ...
No follow-up card was issued from this lane; card issuance is the operator's act.
```
```
$ git merge-base --is-ancestor 38f937a4f ee50984abe4f11ac337382b48a26328f091e200a
(exit 0, no output)
```

**Baseline-attribution.** verdict.md text read from `.moai/reports/t342/verdict.md` at worktree
HEAD `6165f9f5e`. t342 worktree tip ancestry checked against pinned develop.

**Gaps.** Did not re-run the "R4 form observed in corpus" check to see whether the resume
condition has since triggered — the card itself states the resume condition as the thing to check
before acting, and doing that full corpus scan was judged out of this sweep's bounded-depth scope.

**Residual-risk.** If the R4 form has since appeared in the corpus (e.g., via a card that landed
after t342), the deferral's resume condition may already be satisfied and this card may be ready
to act on rather than merely "keep deferred" — that determination was not made here.

**Proposed disposition.** `keep` — rests on the verdict.md text confirming the deferral is real,
deliberate, and explicitly awaiting a resume trigger not yet checked by this sweep.

**Overlap candidates.** none observed in the in-scope list — t342 itself is not in
`inscope-all.txt` (already landed).

---

### t359

**Premise (one sentence).** The bigger half of t331's operator-mandated split (add a landing-evidence
column to the todo items table via `ALTER TABLE ADD COLUMN`) depends on t331-A landing first, and
plan-audit iter-1 flagged two misattributed-requirement defects (D1, D2) that must be resolved
before the design can proceed.

**Premise verdict.** `holds` — D1 and D2's underlying requirement text was independently
re-verified against the actual SPEC files cited.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt359\b'` against pinned develop and pinned main) returned only mentions
(see Evidence), no delivering commit; no SPEC directory for t359's own scope exists in
`.moai/specs/`.

**Claim.** t359's precondition (t331-A / SPEC-TODO-LANDING-STATE-001) HAS landed and closed via
the 3-phase close (commit `51daada00`, merge, ancestor of pinned develop). t359's own scope
(adding the landing-evidence column) has not been started — no SPEC directory exists for it, and
its plan-audit D1/D2 findings (cited from t331's iter-1/iter-2 audits) are exactly as described.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt359\b' --oneline
45cff0f59 docs(t331): plan-audit iter-2 verdict — PASS 0.85, 4 blocking routed as a delta fix (t331)
e1d480eba docs(SPEC-TODO-LANDING-STATE-001): iter-1 remediation — scope split to half A, 11 REQ Tier M (t331)
$ git show -s --format=%B e1d480eba
... the evidence half moved to card t359, which depends on this one landing first. ...
Section B.2 is now a pointer that names the questions t359 must answer — what
REQ-TODO-013 actually permits, how SPEC-TODO-ANALYSIS-001 read it, and whether
an observation may name a commit under SPEC-KANBAN-QUEUE-PR-SYNC-001 REQ-1.10 ...
```
Both hits are mentions (t331's own audit history), not t359's delivery — confirmed by reading
each commit's full subject/body.
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --oneline --grep='SPEC-TODO-LANDING-STATE-001' | head -3
51daada00 merge(WT-card-landing-state): SPEC-TODO-LANDING-STATE-001 — landed answer resolved from the integration branch (t331)
fee6c22d9 chore(SPEC-TODO-LANDING-STATE-001): backfill sync_commit_sha (t331)
c9f712232 docs(SPEC-TODO-LANDING-STATE-001): sync-phase — 3-phase close, CHANGELOG + docs-site (t331)
$ git merge-base --is-ancestor 51daada00 ee50984abe4f11ac337382b48a26328f091e200a
(exit 0, no output)
```
```
$ grep -n "REQ-TODO-013" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
59:- **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record
shape ... changing it only additively (the high-water mark, per REQ-TODO-009).
$ grep -n "REQ-1.10" .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md
251:**REQ-1.10** — The resolver shall not name, return, or otherwise claim which
```
D1 confirmed: REQ-TODO-013 does say "additive-only," not "field-set freeze." D2 confirmed:
REQ-1.10 of the completed SPEC-KANBAN-QUEUE-PR-SYNC-001 does forbid the resolver from naming which
commit — directly bearing on t359's plan to add a commit-SHA-display column.

**Baseline-attribution.** Commit ancestry against pinned develop `ee50984ab...`. Requirement text
read from `.moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md:59` and
`.moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md:251` at worktree HEAD `6165f9f5e`.

**Gaps.** Did not check whether SPEC-TODO-ANALYSIS-001 (the SPEC the card says "판정한 기록도
미조정" — read the opposite way) has been reconciled since — only the card's characterization was
taken at face value for that specific sub-claim; the direct text of SPEC-TODO-ANALYSIS-001:51 was
not re-opened. Did not check whether a plan-phase draft for t359's own SPEC exists in a live
worktree outside the two pinned refs.

**Residual-risk.** If SPEC-TODO-ANALYSIS-001's actual text does not read the way the card
describes, D1's "resolved" framing (that REQ-TODO-013 permits ADD COLUMN) could be less settled
than the evidence above suggests.

**Proposed disposition.** `keep` — rests on t331-A's landing (verified ancestor of pinned develop)
having satisfied the precondition, and D1/D2's underlying requirement text matching the card's
citations exactly.

**Overlap candidates.** none observed in the in-scope list — t331 (SPEC-TODO-LANDING-STATE-001),
SPEC-TODO-ANALYSIS-001, and SPEC-KANBAN-QUEUE-PR-SYNC-001 are not themselves in
`inscope-all.txt` (t331 already landed).

---

### t360

**Premise (one sentence).** GLM effort transmission is keyed to the model's High-tier slot at
three call sites while the web console's per-tier UI lock only disables non-max options when that
specific tier's slot is `glm-5.3-flash`, so a non-flash slot can silently accept and then discard a
saved low-effort setting.

**Premise verdict.** `holds` — both sides of the mismatch verified directly in source.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt360\b'` against pinned develop and pinned main) returned only a mention
(t350's commit naming t360 as an out-of-scope defect it discovered), not a delivering commit.

**Claim.** The three cited transmission call sites and the per-tier UI lock function are exactly
as the card describes, and the mismatch is real at current HEAD.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt360\b' --oneline
e1481f4d5 feat(config): split the GLM Fable slot onto glm-5.3 (t350)
$ git show -s --format=%B e1481f4d5
... Out of scope, filed as t360: the web console's flash effort lock is
per-tier (assets/app.js:490-497) while the reasoning wire derives from
the High slot (launcher.go:1207), so the Fable effort select now unlocks
while a stored effort.fable is still pinned to max. Pre-existing keying
defect that every-slot-flash had been masking.
```
```
$ grep -n "ResolveGLMReasoningForModel" internal/web/agentfm.go
314:	return template.ResolveGLMReasoningForModel(llm.GLM.Models.High, name, me.Effort).Name
$ sed -n '485,500p' internal/web/assets/app.js
  function applyGLMFlashEffortLock(modelSel) {
    var tier = modelSel.name.slice("llm.glm.models.".length);
    var effort = document.querySelector('select[name="llm.glm.effort.' + tier + '"]');
    if (!effort) return;
    var isFlash = modelSel.value === "glm-5.3-flash";
    for (var o = 0; o < effort.options.length; o++) {
      var opt = effort.options[o];
      opt.disabled = isFlash && opt.value !== "max";
      ...
```
Confirms: transmission at `agentfm.go:314` keys on `llm.GLM.Models.High` unconditionally; the UI
lock keys per-tier (`tier := modelSel.name...`) on whether that tier's own slot is
`glm-5.3-flash`. Since t350 landed (Fable slot now `glm-5.3`, not flash), the Fable UI unlocks
while transmission still keys on the High slot's model — exactly the described mismatch.

**Baseline-attribution.** Source read at worktree HEAD `6165f9f5e`:
`internal/web/agentfm.go:314`, `internal/web/assets/app.js:485-500`. Commit `e1481f4d5` and its
ancestry checked against pinned develop.

**Gaps.** Did not check `internal/cli/model.go:117` or `internal/cli/launcher.go:1207` (the other
two cited call sites) directly — only the first (`agentfm.go:314`) was opened; the card's citation
of the other two was not independently re-verified line-for-line. Did not run the test suite to
confirm existing test expectations would need to move, as the card predicts.

**Residual-risk.** If `model.go:117` or `launcher.go:1207` have since been partially fixed
independent of this card, the defect's full extent (three sites vs. fewer) could be narrower than
stated.

**Proposed disposition.** `keep` — rests on the one directly-verified call site plus the UI lock
function matching the card's description exactly, and t350's own commit message naming this as a
real, unaddressed follow-up.

**Overlap candidates.** none observed in the in-scope list — t350 (the SPEC that surfaced this) is
not itself in `inscope-all.txt` (already landed).

---

### t361

**Premise (one sentence).** `TestBinaryLag_OneSeamServesBothSurfaces` in `internal/cli` fails on
develop CI (both Test and Race jobs) because a `deferred-scan` async-suppression switch introduced
by t333 lives in the unexported `internal/hook` package and cannot be reached from
`internal/cli`'s test binary, so a goroutine spawned by `guardLivenessRefresh` writes into a
`t.TempDir()` after that test's cleanup has already run.

**Premise verdict.** `unverified` — the structural claims (comment text, switch scope) check out,
but the causal mechanism the premise actually asserts was not reproduced, so the premise as worded
is undecided. The card itself frames the mechanism as an unreproduced hypothesis, and this sweep
did not reproduce it either. (Orchestrator normalization, M3 post-check: the worker wrote a
compound verdict here; AC-BH-011 admits exactly one of `holds` / `falsified` / `unverified`, and a
premise whose causal claim is unreproduced is undecided rather than holding. The worker's substance
is unchanged.)

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt361\b'` against pinned develop and pinned main) returned no output —
no fix has landed for this defect.

**Claim.** The structural facts the card's attribution rests on are confirmed at HEAD: the
"never awaited" comment exists verbatim, and `deferredScansAsync` is an unexported package-level
var in `internal/hook` toggled only by that package's own `TestMain`. Whether this actually
produces the observed CI failure was not independently reproduced by this sweep (nor claimed to be
by the card, which explicitly requires reproduction as the first step of any fix).

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt361\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt361\b' --oneline
(no output)
$ grep -n "guardLivenessRefresh\|never awaited" internal/hook/session_start.go
140:	// activations that got that far. The refresh is never awaited, so entering
146:	guardLivenessRefresh(ctx, guardLivenessRoot, h.asyncDeferredScans())
$ grep -rn "deferredScansAsync\b" internal/hook/*.go
internal/hook/main_test.go:38:// It also flips deferredScansAsync to false for the test binary: ...
internal/hook/main_test.go:47:	deferredScansAsync = false
internal/hook/session_start_guard_liveness.go:80:		// Test path (TestMain sets deferredScansAsync=false): run inline so no
```
Confirms: `deferredScansAsync` is lowercase (unexported), and the only place it is set to `false`
is `internal/hook/main_test.go`'s `TestMain` — a switch scoped to that one package's test binary,
exactly as the card describes. Since the failing test (`TestBinaryLag_OneSeamServesBothSurfaces`)
lives in `internal/cli`, it is a different test binary and cannot reach this switch.

**Baseline-attribution.** Source read at worktree HEAD `6165f9f5e`:
`internal/hook/session_start.go:140,146`, `internal/hook/main_test.go:38,47`,
`internal/hook/session_start_guard_liveness.go:80`.

**Gaps.** Did NOT run `internal/cli`'s `TestBinaryLag_OneSeamServesBothSurfaces` to reproduce the
cleanup failure directly — the card itself marks this as the mandatory first step before any fix,
and reproducing it was judged beyond this sweep's bounded per-card depth (it would require running
a package test, observing a possibly-flaky timing failure, and confirming both the ubuntu and Race
jobs). Did not check whether t352 (t.TempDir cleanup race, explicitly named by the card as a
possible same-class sibling) has since absorbed this defect — t352 is not in this sweep's
in-scope list.

**Residual-risk.** Since the causal chain from "goroutine outlives TempDir cleanup" to "observed
CI failure" was not reproduced here, there remains a chance the actual CI failure has a different
or additional cause not captured by this hypothesis — the card itself flags this same risk.

**Proposed disposition.** `keep` — rests on the structural evidence above (comment text + switch
scope) matching the card's attribution chain steps 3-4 exactly; step 1's reproduction remains
undone by both the card's author and this sweep.

**Overlap candidates.** t352 (WT-tempdir-cleanup-race, live worktree per
`00-worktree-list.txt`) — named explicitly by the card itself as a possible same-class sibling to
fold into. Not in this sweep's in-scope list, so not independently checked here.

---

### t363

**Premise (one sentence).** Every GitHub Actions workflow whose `concurrency.group` includes
`develop`-triggered pushes is keyed on `github.ref`, which differs between a `push` run
(`refs/heads/develop`) and a `pull_request` run (`refs/pull/N/merge`) for the same head commit, so
`cancel-in-progress` cannot cancel one against the other — causing double CI runs for every
develop→main release PR.

**Premise verdict.** `holds` — the concurrency group expression and the trigger config were both
verified directly against `.github/workflows/ci.yml`.

**Landing verdict.** `not-landed`
- commit: —
- pinned ref: —
- `--is-ancestor` exit: —
- branch + tip (in-flight only): —

Both queries (`--grep='\bt363\b'` against pinned develop and pinned main) returned no output —
this card is brand new (per the dispatch note, absent from the plan-phase snapshot) and no fix has
landed.

**Claim.** `ci.yml`'s `concurrency.group` is `${{ github.workflow }}-${{ github.ref }}` exactly as
cited, and its triggers are `push: branches: [main, develop]` plus
`pull_request: branches: [main]` — confirming the card's claim that a develop→main PR run and a
develop push run share a head commit but not a `github.ref`, so they cannot cancel each other.

**Evidence.**
```
$ git log ee50984abe4f11ac337382b48a26328f091e200a --perl-regexp --grep='\bt363\b' --oneline
(no output)
$ git log 48239c7dc7428c8751a04f6321887c2d36123884 --perl-regexp --grep='\bt363\b' --oneline
(no output)
$ sed -n '16,20p;28,31p' .github/workflows/ci.yml
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]  # main으로 향하는 모든 PR에서 CI 실행
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
$ ls .github/workflows/*.yml .github/workflows/*.yaml | wc -l
18
```
The 18-workflow count matches the card's "저장소 18개 중 일곱" framing (total corpus size
confirmed; the specific 7-workflow no-branch-filter list was not individually re-verified — see
Gaps).

**Baseline-attribution.** `.github/workflows/ci.yml` read at worktree HEAD `6165f9f5e`. Workflow
file count via `ls` at the same HEAD.

**Gaps.** Did not open the other 6 workflows the card names (graph-freshness, lsel-leak-guard,
spec-lint, docs-i18n-check, claude, community, spec-status-auto-sync) to individually confirm each
lacks a `pull_request.branches` filter — only the total count (18) was cross-checked, not the
per-file claim. Did not verify the card's claim about actual observed CI run history (e.g., a
specific PR where both a push run and a PR run fired for the same head) — that would require a
live `gh run list` query against GitHub, which was judged out of scope for a read-only tree sweep.

**Residual-risk.** The concurrency-key mismatch is a structural fact confirmed directly in the
workflow YAML; whether it is the SOLE cause of "double CI runs" for every develop→main PR (versus
a contributing factor among several) was not independently re-derived here — taken from the card's
own reasoning chain, which itself states this was arrived at by falsifying a simpler branch-filter
hypothesis.

**Proposed disposition.** `keep` — rests on the verbatim `concurrency.group` expression and the
`push`/`pull_request` trigger blocks in `ci.yml`, both directly re-read at HEAD.

**Overlap candidates.** t294 (WT-freshness-trigger, live worktree per `00-worktree-list.txt`,
locked) — the card text states t294 was split off this same investigation and retains only the
graph-freshness branch-filter fix (axis A), while this card (t363) is the separate concurrency-key
issue (axis B). Not in this sweep's in-scope list, so not independently checked here.
