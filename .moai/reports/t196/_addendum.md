
---

# Addendum — iter-1 round 2 (two lead-supplied leads)

Same tree, same HEAD `297a21ea7`. Two items were handed to me as the lead's own measurements with
the instruction not to copy the judgment. I re-measured both. **One of the lead's two hypotheses is
false; the other is correctly observed but misdiagnosed. Chasing them surfaced one new blocking
defect the SPEC does not mention at all.**

Revised score: **0.75** (Completeness 0.85 → 0.75 on D10). Verdict unchanged: **FAIL**.

---

## D10 — the shipped rule that teaches skill authors instructs the exact opposite of B.D5, and sits outside every scope statement in the SPEC — `internal/template/templates/.claude/rules/moai/development/skill-authoring.md`:L219, L226, L301; `.../workflow/worktree-integration.md`:L386 — Severity: **critical** — Class: **blocking**

**Verified by running.** Found while chasing lead item (1); not in the lead's report and not in the
SPEC.

Measured spread of the token across the whole template tree:

```
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/ | wc -l                     → 50
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/.claude/skills | wc -l       → 46
$ grep -rn 'CLAUDE_SKILL_DIR' internal/template/templates/ | grep -vc '/.claude/skills/' → 4
$ grep -rl 'CLAUDE_SKILL_DIR' internal/template/templates/ | wc -l                     → 11
```

The SPEC counts 46 lines / 9 files. The tree carries **50 lines / 11 files**. The 4-line, 2-file
remainder lives in the shipped **rules** tree, and REQ-CSN-006 / 007 / 008 are all scoped to
`internal/template/templates/.claude/skills/**`, as is REQ-CSN-010's regression guard. §D names four
out-of-scope topics; none of them is the rules tree. So these 4 lines are neither in scope nor
declared out of scope — they are simply unseen.

That would be a counting quibble if the content were incidental. It is not. Verbatim, from the
shipped rule that teaches skill authors how to write skills:

```
skill-authoring.md:226
  Use `${CLAUDE_SKILL_DIR}` for referencing files within the skill directory
  instead of relative paths. This is more reliable across different invocation contexts.

skill-authoring.md:301
  - Use `${CLAUDE_SKILL_DIR}` for self-referencing paths within skill content

skill-authoring.md:219
  | `${CLAUDE_SKILL_DIR}` | Absolute path to the skill's own directory | v2.1.69 |

worktree-integration.md:386
  | Read-only references | Skills, configs via `${CLAUDE_SKILL_DIR}` | YES | ... |
```

Line 226 is the adopted design (B.D5: replace the variable with a project-root-relative path)
stated in reverse, as normative guidance, shipped to every user project, with a justification
("more reliable across different invocation contexts") that this SPEC's §A.4 measured to be false
under codex.

The consequence is concrete and reachable. §E.2 declares the SPEC complete when
`templates/.claude/skills/**` carries zero tokens. That state is satisfiable while the shipped rule
still instructs the next skill author to reintroduce the token — and REQ-CSN-010's guard, scoped to
the skills tree, would then fire on the author who followed the project's own documented rule. The
guard would be right and the author would be right; the rule is what is wrong, and nothing in this
SPEC touches it.

Required fix: pick one and state it. Either (a) extend REQ-CSN-006/007/008 scope from
`templates/.claude/skills/**` to the token's real footprint and correct `skill-authoring.md`
L219/L226/L301 and `worktree-integration.md` L386 — noting that L219 is a factual capability table
and may be *retained* while L226/L301's normative preference is inverted; or (b) declare the rules
tree explicitly out of scope in §D with a named successor card, and add to §E.2 that the completion
state leaves shipped guidance contradicting B.D5. Silently leaving it unseen is the one option that
is not available, because the SPEC's own §A.1 unit-mismatch discipline — where the wide unit was
chosen precisely because the narrow one missed a real breakage site — applies here with the same
force one level up.

---

## Lead item (1) — the passthrough registration

### (1a) The premise "REQ-CSN-008 orphans the registration" is **not established** — no defect

**Verified by running.** The renderer only executes on `.tmpl` files:
`internal/template/deployer.go:189` — `isTemplate := strings.HasSuffix(path, ".tmpl")`; the two
`Render()` call sites (deployer.go:199, 359) are both behind that suffix test, and 359 skips
non-`.tmpl` outright.

Every file in the template tree carrying the token is markdown, none is a template:

```
$ grep -rl 'CLAUDE_SKILL_DIR' internal/template/templates/ | sed 's/.*\.//' | sort | uniq -c
      11 md
```

So skill bodies never reach the renderer, and the passthrough registration never consumed them.
Removing the token from `.md` files changes nothing about that line's consumer set — it was already
zero with respect to the template tree before this SPEC proposed anything. The registration is a
defensive entry in a validation-suppression list, not a live coupling.

The lead's observation that "no REQ/AC owns this line" is factually right and is the **correct**
outcome: touching `renderer.go` would be a drive-by outside REQ-CSN-006/007/008's stated scope.

### (1b) The braced form **does** ride the same registration — the lead's hypothesis is false

**Verified by running.** Mechanism, `internal/template/renderer.go:110-113`:

```go
for _, tok := range claudeCodePassthroughTokens {
    masked = strings.ReplaceAll(masked, tok, "")                 // $CLAUDE_SKILL_DIR
    masked = strings.ReplaceAll(masked, "${"+tok[1:]+"}", "")    // ${CLAUDE_SKILL_DIR}
}
```

One registration, both forms — the second line reconstructs the braced spelling from the same
entry. Confirmed by a throwaway probe (`internal/template/zz_audit_probe_test.go`, run then
deleted; working tree verified clean before and after):

```
case=bare                         err=<nil>  out="{"p": "$CLAUDE_SKILL_DIR/workflows/plan.md", ...}"
case=braced                       err=<nil>  out="{"p": "${CLAUDE_SKILL_DIR}/workflows/plan.md", ...}"
case=control-unregistered-braced  err=template: unexpanded dynamic token detected: found "${TOTALLY_UNREGISTERED}"
```

The control matters: it proves the braced arm of the detector is live, so the braced pass is a real
registration hit rather than an unreached code path. Had I only run the braced case, its green
would have been uninterpretable.

**So §A.4's citation is not a category error — but it is weaker than it reads.** The passthrough
list is a *validation-suppression* list, not a substitution path: the renderer expands `{{...}}`
only and never expands `$VAR` in any spelling, and skill `.md` files are not rendered at all. The
registration therefore proves the token is *deliberately left for runtime resolution* — it does not
prove the token *is unset under codex*. That second claim rests entirely on §A.4's other cited
ground, the absence of any exporter in `codex_launcher.go`. §A.4 states both grounds, so its
conclusion stands; the presentation just gives the weaker of the two the leading position.

### (1c) New minor hazard the SPEC should name — Severity: **minor** — Class: **optional**

Because the registration looks orphaned after REQ-CSN-008 lands, a run-phase tidy-up is a live
temptation. Deleting it turns `TestPassthroughTokensCompleteness` red —
`internal/template/renderer_test.go:429-446` asserts `"$CLAUDE_SKILL_DIR": true` MUST be present in
the list. The failure is loud and CI-caught, so the risk is low, but a one-line anti-pattern entry
in `plan.md` §G ("do not remove the `renderer.go` passthrough registration — it is not orphaned,
and a completeness test requires it") costs nothing and closes it.

---

## Lead item (2) — the dogfood copy

**Verified by running.** The lead's measurement reproduces exactly:

```
$ grep -rn 'CLAUDE_SKILL_DIR' .claude/skills | wc -l   → 46      (template: 46)
$ grep -rl 'CLAUDE_SKILL_DIR' .claude/skills | wc -l   → 9       (template: 9)
$ grep -rn 'CLAUDE_SKILL_DIR' .claude/rules  | wc -l   → 4       (template: 4)
```

Same nine names, same line counts, in both trees — the local rules copy mirrors the D10 remainder
too.

**My judgment: primarily (a), explicitly NOT (b), and not (c).**

**(a) is most of it.** Template-First is already a project-wide HARD obligation, and this SPEC
restates it as REQ-CSN-012. A SPEC does not need to re-derive a standing rule.

**(b) is wrong, and would be actively harmful.** `moai update` wipes and redeploys
`.claude/skills/moai*` from the *embedded binary*, not from the tree — so the local copy is a
derived artifact whose contents track whichever binary is installed. A guard over both trees would
go red whenever the installed binary lags the working tree, and this repository has a documented
history of exactly that lag. That red would read "the token came back" while meaning "your binary
is stale" — a guard whose failure does not mean what its name says is worse than no guard, because
the next person learns to discount it. Guarding a derived artifact against its own generator's
staleness generates false signal by construction.

**Not (c), because there is a real residue underneath the misdiagnosis.** This repository is itself
a codex target, so a missed sync leaves the very defect the SPEC exists to remove sitting in the
tree that this project's own codex sessions read. REQ-CSN-012 and `plan.md` §D both say the edit
origin is the template and that `make build` follows — neither says the resulting local state is
ever *looked at*. The gap is evidence, not enforcement.

**Recommended fix — one step, no new guard:** add to M3's closing steps that after `make build` the
local census is recorded beside the template census (`grep -rn 'CLAUDE_SKILL_DIR' .claude/skills | wc -l`
alongside the template count), as a paired pre/post value in the same style AC-CSN-008 already uses.
That makes the sync evidenced rather than assumed, costs one command, and produces no false red when
the binary lags — a mismatch there is simply reported, with both numbers visible, for a human to
attribute.

---

## Revised scores and defect list

| Dimension | iter-1 | revised | reason |
|---|---|---|---|
| Clarity | 0.85 | 0.85 | unchanged |
| Completeness | 0.85 | **0.75** | D10 — the token's real footprint (50 lines / 11 files) exceeds the SPEC's scope (46 / 9), and the remainder is neither in scope nor declared out of scope |
| Testability | 0.70 | 0.70 | unchanged |
| Traceability | 0.70 | 0.70 | unchanged |

Aggregate = mean(0.85, 0.75, 0.70, 0.70) = **0.75** < 0.80. Verdict remains **FAIL**.

Blocking defects for iter-2, in fix order: **D1**, **D10**, **D3**, **D4**, **D2**, **D5**, **D6**.
Optional: **D7**, **D8**, **D9**, **(1c)**, and the M3 census step from item (2).

## Probe hygiene

One throwaway artifact was created and removed in this round:
`internal/template/zz_audit_probe_test.go` (braced-form probe). The working tree was verified clean
before and after — only the two untracked directories this card owns
(`.moai/reports/t196/`, `.moai/specs/SPEC-CODEX-SKILL-NEUTRAL-001/`) are present, and no tracked
file was modified at any point during this audit.
