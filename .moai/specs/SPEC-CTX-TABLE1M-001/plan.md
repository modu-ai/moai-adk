---
id: SPEC-CTX-TABLE1M-001
title: "Plan — context-window 1M-default correction"
version: "0.1.1"
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
---

# Plan — SPEC-CTX-TABLE1M-001

> Stateless artifact (no `status:` field, per spec-frontmatter-schema.md §
> Artifact Statelessness). Baseline tree for this plan: worktree t1415, HEAD
> `c50da9c2f` (branch `WT-ctx-table-1m-fix`), working tree clean — all RED
> observations in acceptance.md were measured on this tree.

## §A Edit sequence (Template-First)

Per-file order is **mirror first, then live**, so the `go:embed` source is
never stale relative to the working rule; both copies of a pair receive the
same bytes and `cmp` proves it before the pair's milestone closes. The run
phase edits BOTH copies of each file — a one-side-only fix is the recurring
defect this card family repairs and fails AC-CTM-003 by construction.

```
M1  context-window-management pair
    1. Edit mirror   internal/template/templates/.claude/rules/moai/workflow/context-window-management.md   line 16  → E-SPEC C1
    2. Edit live     .claude/rules/moai/workflow/context-window-management.md                              line 16  → E-SPEC C1 (identical bytes)
    3. cmp live vs mirror          → exit 0, no output
    4. AC-CTM-001 green greps      → all green (acceptance.md §AC-CTM-001)

M2  model-policy pair
    5. Edit mirror   internal/template/templates/.claude/rules/moai/development/model-policy.md   lines 90, 22, 28, 66 → E-SPEC P1..P4
    6. Edit live     .claude/rules/moai/development/model-policy.md                               lines 90, 22, 28, 66 → E-SPEC P1..P4 (identical bytes)
    7. cmp live vs mirror          → exit 0, no output
    8. AC-CTM-002 green greps      → all green (acceptance.md §AC-CTM-002)

M3  embed refresh + close verification
    9.  make build                 → exit 0 (embed refresh; agents-emit-check precedes, read-only)
    10. Final verification batch   → acceptance.md §Final Batch (greps + cmp x2 + wc -c + neutrality + git status)
    11. Commits (convention below)
```

Edits are full-line (or unique-fragment) Edit-tool replacements — never sed,
never regex tooling. Each `old_string` below is unique in its file (verified:
each matches exactly once; the D3 fragment ends "budgets 200K." with a period
that line 90's continuation "budgets 200K and ..." does not carry).

## §B Edit specifications (authoritative old → new text)

Every new line below was byte-measured at plan time (awk length, drafts under
/tmp/t1415-drafts/). Apply to the mirror AND the live copy; the two copies
must end byte-identical.

### C1 — context-window-management.md line 16 (both copies)

OLD (382 chars, verbatim):

```
| 200K sessions — Sonnet 4.6 / Opus 4.6 without `[1m]`; Opus 4.8+ running with a 200K window (e.g. on Bedrock / Google Cloud / Foundry); any native-1M model under `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`; `sonnet` behind an LLM gateway (non-Anthropic `ANTHROPIC_BASE_URL`) unless `sonnet[1m]` is selected; Sonnet 4.5 / Opus 4.5 and earlier | 200,000 tokens | **90%** | ~180,000 tokens |
```

NEW (481 bytes, byte-measured, verbatim; REQ-CTM-001):

```
| 200K sessions — any native-1M model under `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`; Sonnet 4.6 / Opus 4.6 without `[1m]`; models below the 1M-default line behind an LLM gateway (non-Anthropic `ANTHROPIC_BASE_URL`; CC 2.1.285 / 2.1.287 default 1M on gateways for Sonnet 5+ / Opus 4.7+ / Fable and on Bedrock / Vertex / Foundry for Opus 4.7+ and Fable); a 200K-capped gateway — run `/autocompact 200k`; Sonnet 4.5 / Opus 4.5 and earlier | 200,000 tokens | **90%** | ~180,000 tokens |
```

Design notes (repair iteration 1): flag path listed first (the load-bearing
200K path); the surviving 200K segments follow the dispatch guidance; the
Bedrock/Foundry 200K clause is dropped entirely (2.1.287 made 1M the default
there for Opus 4.7+ / Fable — the table no longer asserts the removed
expectation). Iteration-1 additions: the parenthetical carries BOTH upstream
defaults with the per-surface model lists exactly as the changelog states
them — gateways for Sonnet 5+ / Opus 4.7+ / Fable (2.1.285), Bedrock / Vertex
/ Foundry for Opus 4.7+ and Fable (2.1.287); a merged single-list phrasing
would overstate the provider case for Sonnet, which no bullet supports
(audit D3, faithful form applied). A 200K-capped-gateway segment carries the
`/autocompact 200k` remedy from the 2.1.285 bullet (audit D2). Line 19
(precedence sentence) is NOT edited (REQ-CTM-002).

### P1 — model-policy.md line 90 (both copies)

OLD (555 chars, verbatim; the dispatch-named bullet):

```
- The `[1m]` doctrine still binds two surviving paths: (1) the **opus`[1m]`** path (Opus variants still expose `[1m]` selection), and (2) the **gateway / older-model** path — behind an LLM gateway (`ANTHROPIC_BASE_URL` non-Anthropic) or with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K and `sonnet[1m]` selects the 1M window for Sonnet 5.5. The flag's blast radius widened at CC 2.1.223 — it now holds every native-1M Claude model (Opus 5.5 included) to 200K, so setting it changes the Opus-side budget too, not only the sonnet-side one.
```

NEW (814 chars, verbatim; REQ-CTM-003):

```
- The `[1m]` doctrine still binds two surviving paths: (1) the **opus`[1m]`** path (Opus variants still expose `[1m]` selection), and (2) the **flag / older-model** path — with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K, and behind an LLM gateway (`ANTHROPIC_BASE_URL` non-Anthropic) the 200K budget now applies only to models below the default-1M line (CC 2.1.285 defaults gateways to 1M for Sonnet 5+ / Opus 4.7+ / Fable — a gateway `sonnet` no longer budgets 200K by default; Sonnet 4.x and older still do); on those paths `sonnet[1m]` selects the 1M window where the model exposes the suffix. The flag's blast radius widened at CC 2.1.223 — it now holds every native-1M Claude model (Opus 5.5 included) to 200K, so setting it changes the Opus-side budget too, not only the sonnet-side one.
```

Design notes: path (2) renamed "gateway / older-model" → "flag / older-model"
(no external cross-references — swept, `.claude/` + `internal/` +
`.moai/docs/`, phrase occurs only on this line in both trees); the flag
sub-path and the gateway sub-path are split; the CC 2.1.223 blast-radius
sentence is preserved verbatim; "selects the 1M window for Sonnet 5.5" is
generalized to "where the model exposes the suffix" (line 89 of the same file
records that Anthropic-API Sonnet 5 has no `[1m]` suffix — the old
Sonnet-5.5-specific claim was unverifiable in the flag context).

### P2 — model-policy.md line 22, fragment replacement (both copies)

OLD fragment (verbatim, unique in file):

```
Behind an LLM gateway or with `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K.
```

NEW fragment (verbatim; REQ-CTM-003):

```
With `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, `sonnet` budgets 200K; behind an LLM gateway, the 200K budget now applies only to models below the default-1M line (CC 2.1.285 defaults gateways to 1M for Sonnet 5+ / Opus 4.7+ / Fable), so a gateway `sonnet` no longer budgets 200K by default.
```

Design notes: line grows 950→1146 chars; the rest of line 22 (Sonnet 5.5
entry, the CC 2.1.198 re-scope pointer, the CC 2.1.223 flag note, the CC
2.1.225 doc-lag note) is untouched.

### P3 — model-policy.md line 28, fragment replacement (both copies)

OLD fragment (verbatim, unique in file):

```
Opus 4.8 and later run with a 200K window on some providers, such as Amazon Bedrock, Google Cloud, and Microsoft Foundry.
```

NEW fragment (verbatim; REQ-CTM-004):

```
Since CC 2.1.287, the 1M window also defaults on Bedrock / Google Cloud (Vertex) / Microsoft Foundry and the Claude apps gateway for Opus 4.7+ and Fable — a 200K window on those surfaces is now the opt-in, not the default expectation.
```

Design notes: line grows 515→630 chars (iteration 1 names the fourth 2.1.287
surface, the Claude apps gateway — audit D4; "providers" → "surfaces" since
the gateway is not a provider); the first sentence (Anthropic-API 1M
default) and the fast-mode / Explore sentences on the same line are untouched.
This is the "provider-window paragraph at the top of this file" that line 66
cross-references — corrected before its referrer reads coherently.

### P4 — model-policy.md line 66, fragment replacement (both copies)

OLD fragment (verbatim, unique in file):

```
This resolution does not carry to 200K-budget paths — `sonnet` behind an LLM gateway or under `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, and Opus on a 200K provider such as Amazon Bedrock (see the `sonnet` alias entry and the provider-window paragraph at the top of this file): there
```

NEW fragment (verbatim; REQ-CTM-004):

```
This resolution does not carry to 200K-budget paths — under `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`, and for models below the default-1M line on an LLM gateway or a cloud provider (CC 2.1.285 / 2.1.287 default 1M there for Sonnet 5+ / Opus 4.7+ / Fable; see the `sonnet` alias entry and the provider-window paragraph at the top of this file): there
```

Design notes: the fragment ends at ": there" so the shared tail ("a stripped
suffix can still land on 200K, and the legacy hazard below applies.") is
untouched; line grows 1311→1378 chars; line 68 ("the 200K-budget paths named
above" + legacy-model enumeration) stays accurate against the corrected
enumeration — no edit (spec §3.1).

### Byte-delta summary (measured at plan time)

| Edit | Old chars | New chars | Delta |
|------|-----------|-----------|-------|
| C1 (cwm l16) | 382 | 481 | +99 |
| P1 (mp l90) | 555 | 814 | +259 |
| P2 (mp l22) | 950 | 1146 | +196 |
| P3 (mp l28) | 515 | 630 | +115 |
| P4 (mp l66) | 1311 | 1378 | +67 |

Iteration-1 re-measure: C1 382→481 bytes (LC_ALL=C, +99 B — D2/D3 additions);
P3 515→630 chars (D4, +115); model-policy total +637. A merged single-list
phrasing of the two upstream defaults was rejected as unfaithful to the
changelog (see the C1 design notes).

Projected post-edit sizes: cwm ≈ 7,098 B (always-loaded; AC bound ≤ 7,099 B);
model-policy ≈ 27,688 B (`paths:`-scoped — no always-loaded duty).

## §C Verification plan (keys per verification-plan-contract)

One owner per key; before running, query the shared snapshot; reuse only on
an exact key match. Every key below names its owner phase (run unless marked).

| Key | tree_key | command | toolchain | environment | scope |
|-----|----------|---------|-----------|-------------|-------|
| VK-RED-* (E1-E16) | `c50da9c2f` clean (measured, plan phase — reuse, do not re-run) | the 16 grep commands in acceptance.md §Evidence Ledger | BSD grep (macOS), BRE patterns, no GNU-only flags | darwin 27, zsh, worktree t1415 | one named file per command |
| VK-CMP-C | post-M1 working tree (re-pin SHA at execution) | `cmp` live vs mirror, cwm pair | POSIX cmp | same | 2 files |
| VK-CMP-P | post-M2 working tree (re-pin) | `cmp` live vs mirror, model-policy pair | POSIX cmp | same | 2 files |
| VK-GREEN-* | post-M2 working tree (re-pin) | the AC-CTM-001/002 green greps + anchors (acceptance.md) | BSD grep | same | one named file per command |
| VK-SIZE | post-M2 working tree (re-pin) | `wc -c` on the two live files | POSIX wc | same | 2 files |
| VK-BUILD | post-M2 working tree (re-pin) | `make build` | make + Go toolchain | same | repository (MUTATING — runs alone, never batched; not read-only batch-safe per verification-batch-pattern.md) |
| VK-NEUTRAL | post-M2 working tree (re-pin) | neutrality greps (AC-CTM-007) | BSD grep | same | 4 files |
| VK-SCOPE | post-M2 working tree (re-pin) | `git status --short` (AC-CTM-008) | git | same | working tree |

Tool-provenance note (verification-claim-integrity.md §2.2): grep/cmp/wc/git
are OS tools with no repository-lag axis; `make build` builds FROM this tree
by construction. RED keys measured at `c50da9c2f` are cited from acceptance.md
without re-execution; a post-edit re-run of a RED key is expected to flip
(green cell) — that flip is the point, not a key violation.

Command-output discipline: keep every command bounded (single file greps
return one line); `git status --short` and `git diff --stat` are bounded by
nature here; `make build` output redirected to a file with exit code +
bounded tail if verbose.

## §D Commit convention (lane executes)

- Plan-phase artifacts commit (if not yet landed on this branch):
  `feat(SPEC-CTX-TABLE1M-001): plan-phase artifacts (Tier S scope, 3 artifacts per dispatch)` — card id t1415 in the body; `Authored-By-Agent: manager-spec` trailer; `🗿 MoAI` trailer.
- M1 commit: `docs(SPEC-CTX-TABLE1M-001): correct 200K-sessions row for CC 2.1.285/2.1.287 gateway and provider defaults` — both cwm copies in one commit (card id t1415 in body).
- M2 commit: `docs(SPEC-CTX-TABLE1M-001): correct gateway/provider 200K claims in model-policy` — both model-policy copies (card id t1415 in body).
- M3 commit: `chore(SPEC-CTX-TABLE1M-001): embed refresh after template mirror correction` (only if `make build` produces a tracked-file delta; otherwise fold into M2).
- Status transition `draft → in-progress` rides the first run-phase commit (manager-develop owns it per the ownership matrix).
- Never `--no-verify`; no `git add -A` — stage by explicit pathspec; re-read `git status --short` immediately before staging.

## §E PRESERVE list (constraints for the run phase)

1. Only the five line edits of §B, applied to the four files of spec §5.
   No other byte of the four files changes (`cmp` + `git diff` prove it).
2. Line 19 of cwm (precedence sentence), line 43 and line 68 of
   model-policy: untouched.
3. No Go source, hook, CI workflow, agent definition, or skill body changes.
4. CHANGELOG/README/docs-site: sync-phase territory — not this run phase.
5. Runtime-managed files (`.moai/state/`, `.moai/logs/`, `.moai/cache/`,
   `.moai/harness/`): untouched.
6. Template neutrality: the mirrored rule texts carry no card id, no SPEC ID,
   no internal dates, no report paths (C1-C8 of
   template-internal-isolation-doctrine.md §25.1). CC version citations
   (`CC 2.1.285`, `CC 2.1.287`) are upstream release facts of the same class
   as the existing `CC 2.1.223` / `CC 2.1.198` citations in these files —
   permitted. AC-CTM-007 greps enforce this.
7. Sync of the corrected text to the live copies is done ONLY by the explicit
   per-file edits of §A — never `moai update` (it deletes local-only files
   under managed roots and requires git-strategy reapply afterward; this card
   syncs exactly two named files). Recorded at repair iteration 1 per D6.

## §F Risks and notes

- **R1 — mirror/live divergence mid-edit.** Between editing the mirror and
  the live copy, `cmp` exits non-zero. That state is transient and expected;
  the milestone closes only on `cmp` exit 0. Do not "fix" by copying the
  mirror over the live file with `cp` — apply the identical Edit so the diff
  stays reviewable.
- **R2 — false-zero on green greps.** Every absence-form green check is
  paired with a presence anchor (acceptance.md §Swept-set discipline) so a
  zero cannot masquerade as a pass on an empty or wrong-path sweep.
- **R3 — line-number drift.** Line numbers in §B are valid at tree
  `c50da9c2f`. If an earlier absorb moved them, re-locate by the verbatim
  old text (each old_string is unique in its file); the text, not the number,
  is the edit target.
- **R4 — concurrent absorbs.** This worktree is a card tree; if `develop`
  must be absorbed mid-card, re-run the §C green keys on the post-absorb tree
  and re-pin the tree_key (verification-claim-integrity.md §4 corollary).
- **R5 — agents-emit-check inside `make build`.** Read-only; this card edits
  no agent definitions, so it must pass unchanged. If it fails, stop — that
  is a foreign defect, report it, do not regenerate `.codex/` emissions from
  this card.
