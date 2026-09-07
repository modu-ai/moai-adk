---
id: SPEC-CODEX-SKILL-PATH-SLASH-001
title: "Codex skills.config path slash normalization — make `moai codex skills disable` operable on Windows without emitting a TOML escape"
version: "0.1.0"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, skills-config, windows, path-normalization, disable-verb, prune, doctor, t540"
related_specs: [SPEC-CODEX-SKILL-PATH-001, SPEC-CODEX-SKILLCONFIG-SHAPE-001]
---

# SPEC-CODEX-SKILL-PATH-SLASH-001 — Codex skills.config path slash normalization

## HISTORY

- 2026-09-08 (plan-phase, v0.1.0) Initial authoring. Card t540, worktree `.claude/worktrees/t540`, branch `WT-codex-path-escape`, base `9ce792637`. Tier M, Class C, cycle_type tdd. Direction (slash normalization, NOT parser escape-decoding) is an operator decision of 2026-09-08. Absorbs t502 sync-audit findings F4 (in scope) and F5 (partially — see §D.3).

## §A. User Story

**As a** MoAI-ADK user on Windows, **I want** `moai codex skills disable <skill>` to actually publish a `[[skills.config]]` entry instead of silently skipping every invocation, **so that** the verb is operable on my platform and does not report success while having done nothing.

**As a** MoAI-ADK maintainer, **I want** that publication to happen without the config file ever carrying a TOML escape sequence, **so that** the repository's verbatim-reading parser and Codex never disagree about which file an entry names — the disagreement that would let the `prune` verb delete a healthy registration.

## §B. Context and Background

### §B.1 The defect, as measured

The publisher refuses any path containing `"`, `\`, LF, or CR. **(I1, inferred — not observed on a Windows runtime in this tree)** every Windows path would contain `\`, so the verb would be inoperative on Windows, reporting rc=0 while doing nothing. The refusal itself (M1) is read from the source; the Windows consequence is the inference the table below marks as I1, and this opening sentence carries the marker rather than asserting it flat.

The table below separates what was **read in this tree at `9ce792637`** from what is **inferred** from
it. The split is visual and deliberate: no Windows runtime has been observed here, so every
Windows-behaviour statement in this document is an inference and is marked as one.

**MEASURED — read directly from the source at `9ce792637`:**

| # | Address | Fact as read |
|---|---|---|
| M1 | `internal/cli/codex_skills_disable.go:245` | `if strings.ContainsAny(skillPath, "\"\\\n\r") { return skip("the path contains a character this config format cannot carry verbatim (%q)", skillPath) }` |
| M2 | `internal/cli/codex_skills_disable.go:148,158,160` | `skillPath` is a `filepath.Join` product built by `resolveCodexSkillMirrorPath` |
| M3 | `internal/codexwiring/skills.go:174` | `skillPathKeyRe = regexp.MustCompile("^path\\s*=\\s*\"([^\"]*)\"\\s*(#.*)?$")` — the parser captures `[^"]*` verbatim and decodes NO TOML escapes |
| M4 | `internal/cli/codex_skills_disable.go:252` | the ONLY path comparison in the verb is `if e.Path == skillPath` |
| M5 | `internal/cli/codex_skills_prune.go:76-94,102` | `Eligible: true` requires BOTH `codexPathAbsolute`/`codexPathHomeRelative` classification AND `fs.ErrNotExist`; `codexPathRelative` and `codexPathOddlyFormed` are non-destructive skips |
| M6 | this host, `go run .moai/reports/t540/lab/ts.go` | `GOOS=darwin Separator='/'` → `ToSlash(backslash) changed=false`, `FromSlash(slash) changed=false`, `IsAbs(slash form)=false`, `IsAbs(backslash)=false` |

**INFERRED — not observed on a Windows runtime in this tree:**

| # | From | Inference |
|---|---|---|
| I1 | M2 + Go's `os.PathSeparator` contract | on Windows `skillPath` always carries `\`, so M1 refuses every invocation there |
| I2 | M5 + `internal/cli/doctor_codex.go:668-677` (`IsAbs` runs first) | a `C:\…` declaration classifies `codexPathAbsolute` on Windows and `codexPathOddlyFormed` on darwin/linux, so the destructive prune branch opens **on Windows only** |

Fact M3 is why the refusal's stated reasoning is **correct and must be preserved**. Emitting an escaped path (`path = "C:\\Users\\..."`) would produce an entry Codex reads correctly and every moai reader reads as a different, absent path. The `prune` verb deletes an entry it classifies as absolute-or-home-relative and cannot stat (M5), so on Windows that divergence is destructive rather than cosmetic (I2) — and Windows is precisely the platform this card exists to serve, so the qualification narrows the claim without reducing the risk. This SPEC removes the refusal's *reach over Windows paths*; it does not remove the property the refusal protects.

### §B.2 Reader census — the full set of consumers

`ParseSkillEntries` has EXACTLY 3 non-test call sites. Full evidence, with the census command and its verbatim output, is already written at `.moai/reports/t540/reader-census.md` — it is cited here, not restated.

| # | Site | What it does with `Path` | Consequence of a misread |
|---|---|---|---|
| 1 | `internal/cli/codex_skills_prune.go:120` (consumes at `:71-94`) | `classifyCodexSkillPath` → stat; deletes the entry **only when** the classification is absolute or home-relative **and** the stat reports `ErrNotExist` (M5) | destructive **on Windows only** (I2) |
| 2 | `internal/cli/doctor_codex.go:723` (consumes at `:824-857` — classification `:831-856`, stat `:857`) | `classifyCodexSkillPath` → stat; reports | false diagnostic |
| 3 | `internal/cli/codex_skills_disable.go:251` | `e.Path == skillPath` string equality (self-match) | duplicate entry published |

Sites 1 and 2 both route through `classifyCodexSkillPath` / `expandCodexHomeRelativePath` (`internal/cli/doctor_codex.go:667,686`). Site 3 is a pure string comparison and is addressed by REQ-CSPS-007 (§D.4).

### §B.3 The chosen direction — slash normalization

The publisher normalizes the path to the config's slash form before the `ContainsAny` check; the two
stat-ing readers convert back to the host's native form before stat. **`internal/codexwiring/skills.go`
is NOT modified.**

Two properties follow, and both are the reason this direction was chosen over reader-side escape decoding:

1. No TOML escape is ever emitted, so the Codex-reads-X / moai-reads-Y divergence never opens. The destructive `prune` misread path (I2) stays closed by construction rather than by a new decoder being correct.
2. The parser is untouched, so t533's ghost-entry counting basis is unchanged. A parser change would move the number t533 measured, and the two cards would no longer be talking about the same population.

#### §B.3.1 The conversion is separator-injected, not `filepath.ToSlash` directly

`filepath.ToSlash` and `filepath.FromSlash` rewrite `filepath.Separator`, and on this card's only
executable host that separator is already `/`. Both functions are therefore the **identity** here —
measured, not assumed (M6). Using them directly would leave every backslash-shaped test input
unchanged on darwin, so it would still trip the `:245` guard and AC-CSPS-002 could not pass on the
one host that can run it. The production behaviour on Windows would be correct; the **testability**
would not exist.

The conversion is therefore taken through an injectable separator, following this package's existing
seam convention (`osStatFn`, `codexUserHomeDir` — package-level vars / explicit parameters, not a new
abstraction):

```go
func toConfigPath(p string, sep rune) string   // sep == '/' → identity; else sep → '/'
func fromConfigPath(p string, sep rune) string // sep == '/' → identity; else '/' → sep
```

Production call sites pass `configPathSeparator`, a package-level `var configPathSeparator = filepath.Separator`.
A test overrides that var (with `t.Cleanup` restore, non-parallel) or calls the pure functions with
`'\\'` directly, and thereby measures Windows semantics on darwin. **AC-CSPS-002's
"platform-independent measurement" claim is true only on top of this seam** — without it the claim is
false, and it is stated here so it cannot be reintroduced.

**[HARD] The conversion MUST be separator-aware, never an unconditional `strings.ReplaceAll(p, "\\", "/")`.**
A unix filename may legally contain `\`, and such a path is today **refused** by the `:245` guard.
An unconditional replacement would convert that refusal into the publication of a *wrong path* —
strictly worse than the defect this card repairs. When `sep == '/'`, `toConfigPath` returns its input
untouched and the guard keeps refusing.

On a `/`-separator host the production separator is `/`, so both functions are identity and observable
behaviour must be byte-identical to base. That is stated as its own acceptance criterion (AC-CSPS-006)
rather than assumed.

### §B.4 Card ordering relative to t533 — t540 is the PRECEDING card

Two lead judgments of 2026-09-08, recorded here as background:

1. **No file conflict.** The t533 worktree has been disposed and `WT-ghost-skills-measure` is an ancestor of `origin/develop` (lead-measured, `git merge-base --is-ancestor` rc=0). No concurrent writer shares these files.
2. **The ordering constraint runs the other way.** t533's `AC-CGM-011/012/013` execute `prune`, and they wait on t540 landing. The destructiveness is platform-conditional, and stating it precisely does not weaken the constraint: **where the host running `prune` is Windows, the deletion branch is reachable and running it before the path-resolution defect is fixed would destroy healthy registrations; on a `/`-separator host the same entries classify `codexPathOddlyFormed` and are skipped non-destructively** (M5 + I2). Safety therefore rests on which host happens to run the verb — an accident, not a guarantee — so the ordering constraint stands: **t540 precedes t533's prune-execution ACs.**

### §B.5 Absorbed t502 debt

Verdicts are stated explicitly rather than left implicit; see §D.3.

- **F4** (t502 `sync-audit.md:38`, round2:63, `245.48` count=0) — the Windows-refusal guard at `:245` has ZERO test coverage. Independently reconfirmed in this tree: `/usr/bin/grep -rn 'cannot carry verbatim' --include='*.go' internal/` matches the production line only (`internal/cli/codex_skills_disable.go:246`). **IN SCOPE** — this SPEC's tests close it.
- **F5** (t502 `sync-audit.md:39`) — `skip` returns rc=0, so on Windows every call looks successful to a script. **PARTIALLY IN SCOPE.**

## §C. Requirements (GEARS)

- **REQ-CSPS-001** (Event-driven, gate) — When a publisher change is proposed for landing, the maintainer shall measure whether Codex resolves a forward-slash `path` value on Windows, and shall record that measurement as evidence under `.moai/reports/t540/`, before the change lands.
- **REQ-CSPS-002** (Event-driven) — When `upsertCodexSkillDisable` receives a skill path containing the host path separator, the publisher shall convert it to the config slash form via `toConfigPath(p, sep)` and publish it, rather than returning a skip verdict.
- **REQ-CSPS-003** (Ubiquitous) — The publisher shall continue to refuse any path that still contains `"`, LF, or CR after conversion, and shall never emit a TOML escape sequence into `config.toml`.
- **REQ-CSPS-004** (Event-driven) — When `prune` or `doctor` resolves a declared `path` to a filesystem location, the reader shall convert it back with `fromConfigPath(p, sep)` before the stat call.
- **REQ-CSPS-005** (Unwanted) — On a host whose separator is not `/`, the `prune` verb shall not classify a slash-shaped entry naming an existing file as absent, and shall not delete it. (The destructive branch is reachable only under `codexPathAbsolute`/`codexPathHomeRelative` + `ErrNotExist` (M5), which on a `/`-separator host a backslash declaration never reaches (I2).)
- **REQ-CSPS-006** (State-driven) — While the host path separator is `/`, `toConfigPath` and `fromConfigPath` shall be the identity function, and the observable behaviour of the publisher and both stat-ing readers shall be byte-identical to the behaviour at base `9ce792637`.
- **REQ-CSPS-007** (Ubiquitous) — `internal/codexwiring/skills.go` shall remain unmodified by this card.
- **REQ-CSPS-008** (Event-driven) — When `upsertCodexSkillDisable` compares a declared entry's path against the path it is about to publish, the comparison shall be performed on the `toConfigPath` form of both sides, so an existing backslash-shaped entry is matched and updated rather than duplicated.
- **REQ-CSPS-009** (Ubiquitous, safety) — Every experiment that writes a Codex config file shall run under an isolated `CODEX_HOME`; no experiment shall write the real `~/.codex/config.toml`.
- **REQ-CSPS-010** (Unwanted) — The conversion shall not be an unconditional backslash replacement: while the host separator is `/`, a path containing `\` shall reach the `:245` guard unchanged and be refused.

### §C.1 REQ → AC traceability

Full Given-When-Then bodies live in `acceptance.md`; this table is the coverage map only.

| REQ | Covering AC |
|---|---|
| REQ-CSPS-001 | AC-CSPS-001 |
| REQ-CSPS-002 | AC-CSPS-002 |
| REQ-CSPS-003 | AC-CSPS-003 |
| REQ-CSPS-004 | AC-CSPS-004, AC-CSPS-005 |
| REQ-CSPS-005 | AC-CSPS-004 |
| REQ-CSPS-006 | AC-CSPS-006 |
| REQ-CSPS-007 | AC-CSPS-008 |
| REQ-CSPS-008 | AC-CSPS-007 |
| REQ-CSPS-009 | AC-CSPS-001 |
| REQ-CSPS-010 | AC-CSPS-003 |

## §D. Decisions and Verdicts

### §D.1 The load-bearing unknown

It is **NOT verified** that Codex on Windows resolves a forward-slash `path` value. This SPEC is written with that question open, and AC-CSPS-001 is the gate that settles it. If the measurement comes back negative, the slash direction fails as a whole and the card falls back to option (b) — reader-side escape decoding — which re-opens the t533 counting-basis question that §B.3 property 2 currently keeps closed.

### §D.2 Ordering constraint inside `classifyCodexSkillPath`

`classifyCodexSkillPath` (`internal/cli/doctor_codex.go:667`) runs `filepath.IsAbs` first and returns `codexPathOddlyFormed` for any path containing `\`. Classification therefore happens on the DECLARED form and `fromConfigPath` is applied to the resolved stat target, not before classification. This ordering is a design constraint the implementation must respect, not a free choice; getting it backwards would reclassify legitimate Windows-absolute declarations.

`classifyCodexSkillPath` itself is **not** separator-injectable — it calls `filepath.IsAbs`, whose
behaviour is fixed by the build's `GOOS`. This card does not refactor it (§F). The consequence is
stated rather than hidden: the *conversion* half of the reader path is measurable on darwin through
the seam, while the *classification* half's Windows behaviour remains inference I2 and is settled
only on a Windows host (§G gap 4).

### §D.3 t502 F4 / F5 verdicts

- **F4 — IN SCOPE.** The guard at `:245` currently has no test. AC-CSPS-002 and AC-CSPS-003 together exercise both the refuse branch and the newly-opened publish branch, which is what closes F4. F4 is closed by this card.
- **F5 — PARTIALLY IN SCOPE.** The *Windows instance* of F5 dissolves once publication succeeds there: there is no longer a platform on which every call silently skips. The *general* question — should a deliberate skip exit non-zero, and how would a caller distinguish "skipped by policy" from "nothing to do"? — is an exit-code policy decision spanning every `skip` branch of the verb, and is **NOT** absorbed here. It remains open as a separate policy card. This SPEC states that explicitly so the residual is not lost by being silently folded into a card that did not decide it.

### §D.4 Reader site 3 — adopted, with one clause retracted

The lead's judgment of 2026-09-08 is **adopted**, and its reasoning is reproduced because it is the reason AC-CSPS-007 exists:

`codex_skills_disable.go:252` compares `e.Path == skillPath` — and per M4 that is the verb's **only** path comparison. Once the publisher normalizes, that comparison fails against an entry already stored in backslash form — one written by hand, or by another tool. The `len(matches) == 0` branch then falls through to `appendDisableEntry`, publishing a **second entry for the same skill**. On the next invocation the duplicate trips the `len(matches) > 1` skip at `:257` (`"%d entries declare this path; collapsing duplicates is \`moai clean --codex-skills\`'s job"`), and cleanup passes to t506's surface. The card would be manufacturing the debt t506 exists to clear.

Normalization must therefore apply at the **comparison** point as well as the publication point. This is the third and last of the three reader sites; omitting it would make the claim "the full reader census is covered" false.

**Retracted clause.** An earlier revision of this section claimed the `len(matches) > 1` branch and
the comparison "could disagree about what counts as a match". That premise is false and is withdrawn:
`:257` performs **no comparison** — it counts the slice the single comparison at `:252` filled. There
is no second comparison site to fall out of step with. The retraction is recorded rather than
silently deleted, because the AC arm it motivated is being kept for a different, valid reason
(below).

Two observations recorded alongside the adoption:

- **MEASURED.** The update branch (`:260-280`) rewrites only the `enabled` line via `setEnabledFalseInExtent`; it leaves the stored `path` line in its original form. That is correct and deliberate — rewriting a declaration this verb did not author is out of scope (§F).
- **INFERRED (I2, not observed).** The residual is that a legacy backslash entry keeps its shape. On Windows it would classify `codexPathAbsolute` and stat successfully; on a `/`-separator host it classifies `codexPathOddlyFormed` and is skipped non-destructively. Only the second half of that sentence has been read on a running host; the Windows half is inference.
- **Why AC-CSPS-007's second arm survives the retraction.** Its purpose is not to catch a disagreement between two comparison sites. It is to show that **after normalization is introduced at the one comparison site, the duplicate-detection branch still behaves correctly** — a config already holding two entries for the same skill (one backslash-shaped, one slash-shaped) must still reach the `len(matches) > 1` skip rather than being silently updated or tripled. Introducing normalization changes which entries land in `matches`, so that branch's input distribution changes and is worth an arm.

## §E. Constraints

- **Hard-to-reverse write.** Writing `~/.codex/config.toml` is hard to reverse. Every experiment runs under an isolated `CODEX_HOME` (REQ-CSPS-009). This is the established harness from SPEC-CODEX-SKILLCONFIG-SHAPE-001.
- **`internal/codexwiring/skills.go:9` declares itself READ-ONLY by construction.** Making the parser writable is not proposed and is not available as a fallback within this card.
- **Verification is scoped.** Touched packages only; never `go test ./...` locally. `internal/cli` needs `-timeout 1800s` (a lane measured 702.5s on 2026-09-07).
- **Absence claims use `/usr/bin/grep`**, not the shell's `grep` (a ugrep wrapper that skips silently).
- **No time predictions.** Priority labels and phase ordering only.
- **Every commit message on this branch contains `t540`.** Evidence path: `.moai/reports/t540/`.

## §F. Exclusions — what this SPEC does NOT build

This section states what is out of scope for this SPEC.

### Out of Scope — parser modification

- `internal/codexwiring/skills.go` is not modified. TOML escape decoding is not implemented, and the parser is not made writable.
- Changing the parser's regex, its verbatim-capture semantics, or its tri-state `enabled` handling.

### Out of Scope — exit-code policy

- The general F5 question ("should a deliberate skip exit non-zero?") is not decided here. Only the Windows instance dissolves, as a side effect of publication succeeding.
- No `skip` branch's return code is changed by this card.

### Out of Scope — duplicate collapse and existing-entry rewriting

- Collapsing duplicate `[[skills.config]]` entries stays with `moai clean --codex-skills` (t506).
- Rewriting the `path` line of an entry this verb did not author. The update branch continues to rewrite only `enabled`.
- The 49 ghost entries measured by t533 are not counted, re-counted, or removed here.

### Out of Scope — prune execution

- This SPEC does not run `prune` against any real config. t533's `AC-CGM-011/012/013` own prune execution and wait on this card (§B.4).

### Out of Scope — path-classification refactor

- `classifyCodexSkillPath` is not made separator-injectable. Its `filepath.IsAbs` call stays as-is, so its Windows behaviour remains inference I2 rather than a darwin-measurable fact (§G gap 4).
- `expandCodexHomeRelativePath` and the `codexUserHomeDir` seam are not changed.
- **No stat seam is introduced at `internal/cli/doctor_codex.go:857`.** That line calls `os.Stat` directly rather than `osStatFn`, so the doctor reader's stat argument cannot be observed by a recorder. Adding the seam is production-code change beyond this card's repair and is owned by a separate card; the consequence for this card's evidence is recorded as §G gap 5, not silently absorbed.

### Out of Scope — new verbs or surfaces

- No new CLI verb, flag, or config key. No change to `moai doctor`'s finding text beyond what REQ-CSPS-004 requires.

## §G. Gaps — what was NOT observed

1. **Windows execution has NOT been observed in this tree.** The chain "every Windows path carries `\` → every Windows invocation hits the refusal" is inference I1 — source reading plus Go's `os.PathSeparator` contract, not a runtime observation. AC-CSPS-002 converts the *publisher* half into a measured fact **only through the §B.3.1 separator seam**; without the seam that claim would be false on this card's only executable host (M6). The *Codex-side* half is AC-CSPS-001 and is genuinely open.
2. **Codex's Windows path resolution is unmeasured.** See §D.1.
3. **The destructive prune branch is narrower than a first reading suggests, and has not been reproduced.** `Eligible: true` requires BOTH an absolute/home-relative classification AND `fs.ErrNotExist` (M5, `codex_skills_prune.go:76-94` + `:102`); `codexPathRelative` and `codexPathOddlyFormed` are non-destructive skips. Because `classifyCodexSkillPath` runs `filepath.IsAbs` first, a `C:\…` declaration is oddly-formed on a `/`-separator host and absolute on Windows — so **the destructive path opens on Windows only** (I2). "Destructive on every platform" is false; "destructive on Windows" is what the source supports, and Windows is the platform this card serves, so the risk is undiminished and only the claim is narrowed. Neither branch has been demonstrated by a run. Corrected census: `.moai/reports/t540/reader-census.md` § 정정 2026-09-08.
4. **`classifyCodexSkillPath`'s Windows behaviour is not measurable on this host.** `filepath.IsAbs` is fixed by the build's `GOOS` and is not separator-injectable, and this card does not refactor it (§D.2). The conversion half of the reader path is measurable through the seam; the classification half is not.
5. **The doctor reader's stat argument is not observable in this card.** `internal/cli/doctor_codex.go:857` calls `os.Stat` **directly** — it does NOT route through the `osStatFn` seam that `codex_skills_prune.go:96` uses, so no recorder can be placed there. Measured: `/usr/bin/grep -n 'os\.Stat\|osStatFn' internal/cli/doctor_codex.go` → `459: os.Stat`, `857: os.Stat` (zero `osStatFn` occurrences). Compounding it, `codexStaleSkillFinding` has **zero existing tests**: `/usr/bin/grep -rn 'codexStaleSkillFinding' --include='*_test.go' internal/cli/` → no output, rc=1. AC-CSPS-004 arm A therefore establishes its RED on the **prune** reader only; the doctor half is a regression guard that claims no RED. Introducing a stat seam at `:857` is production-code change and is **out of scope** here (§F) — it belongs to a separate card.
6. **`filepath.ToSlash`/`FromSlash` identity on this host is measured, and it is the only such measurement.** M6 was taken on `GOOS=darwin` via `.moai/reports/t540/lab/ts.go`. No linux and no Windows run of that probe exists; the linux case is asserted from `Separator == '/'`, not observed.

## §H. Cross-references

- `.moai/reports/t540/reader-census.md` — the reader census evidence cited in §B.2, and its 2026-09-08 corrections (prune destructiveness scope; `ToSlash`/`FromSlash` identity)
- `.moai/reports/t540/lab/ts.go` — the separator probe producing M6
- `.moai/reports/t540/plan-audit.md` — plan-audit iteration 1/2 (FAIL); F-1..F-7 are answered in §B.3.1, §D.2, §D.4, §G, and `acceptance.md`
- SPEC-CODEX-SKILL-PATH-001 — the doctor path-shape semantics this SPEC extends
- SPEC-CODEX-SKILLCONFIG-SHAPE-001 — the isolated-`CODEX_HOME` measurement harness
- t533 (`WT-ghost-skills-measure`) — the ghost-entry counting basis §B.3 preserves; its prune-execution ACs wait on this card
- t506 — duplicate collapse (`moai clean --codex-skills`), the debt §D.4 avoids manufacturing
