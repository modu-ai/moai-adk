# t232 — zone-registry clause drift: measured findings

Card: t232 (#1616). Tree: `.claude/worktrees/t232` @ `294b4b6ab`, branch `WT-zone-registry-drift`.
Binary under test: `bin/moai` built from this tree (`make build`, version stamp v3.1.2 / commit 294b4b6ab).

## 1. Reproduction — a freshly-initialized project fails immediately

```console
$ bin/moai init --root <scratch>/t232repro --non-interactive --language go
  (succeeds)

$ cd <scratch>/t232repro && moai constitution validate; echo exit=$?
exit=1
  Constitution validate: found 67 error(s).

$ grep -c DRIFT validate.txt
67
```

Every one of the 67 is the single sentinel `[DRIFT] … clause "…" not found in source "…"`.
Verbatim output: `validate-repro.txt` (this directory).

`moai doctor` in the same fresh tree surfaces it as a hard failure — the t200/#1612 fix that
landed at this HEAD makes doctor call `constitution.Validate` for real:

```
  fail    Constitution Registry  registry loads (101 entries) but validate found 67 error(s)
  Pass 22    Warn 2    Fail 1
```

So the card's premise holds and is, if anything, understated: the very first `moai doctor` a new
user runs reports a failure. The issue's figure of 65 was measured on `v3.1.3-rc.0`; at this HEAD
it is **67** — the count grows as the rules text keeps moving.

## 2. The registry the user gets is byte-identical to the template

```console
$ diff -q .claude/rules/moai/core/zone-registry.md \
          internal/template/templates/.claude/rules/moai/core/zone-registry.md
IDENTICAL
```

Both drift identically (68 clause failures by the independent analyzer — see §3 for the −1
reconciliation). The template was last touched in `b689f492c` (#1526, 2026-08-14); the rules text
it cites has kept changing since. Downstream repair is impossible: `.claude/rules/moai` is a
managed root, so `moai update` re-deploys the stale file over any local correction.

## 3. Direction of the drift — the template is stale AND the clauses were never quotes

The matcher is not over-matching. `internal/constitution/validator.go:261-264` does exactly what it
documents: `normalizeWhitespace(clause)` must be a substring of
`normalizeWhitespace(stripCodeFences(source))`. Re-implementing that rule independently
(`analyze.py`) reproduces the validator's set exactly:

```console
$ python3 analyze.py <fresh-project> analysis-repro.json
entries=101 clause_fail=68 anchor_fail=17
both_fail=8 clause_fail_anchor_ok=60 clause_ok_anchor_fail=9 missing_file=0

$ python3 reconcile.py analysis-repro.json validate-repro.txt
validator_drift=67 analyzer_clause_fail=68
analyzer_only=['CONST-V3R5-009']   # analyzer artifact: single-quoted YAML value with
validator_only=[]                  # embedded double quotes, mis-parsed by my reader
```

The one-entry gap is a parser artifact on my side, not a 68th defect. **67 is the real number, and
the two implementations agree on the exact same set of IDs.**

Two distinct causes are mixed in those 67:

- **~61 paraphrases.** The `clause:` value is a reworded summary of the doctrine, never a quotation
  of it. `CONST-V3R2-008`'s clause reads `"Language-Aware Responses: All user-facing responses MUST
  be in user's conversation_language…"` — a sentence that appears nowhere in any source file. These
  are *structurally* unpassable under an exact-substring matcher; they were authored against a
  contract they never satisfied.
- **6 short summary labels.** `SPEC+EARS format`, `@MX TAG protocol`, `16-language neutrality`,
  `Template-First discipline`, `AskUserQuestion monopoly`, `Claude Code substrate`. Same cause,
  shorter form.

## 4. A second defect the validator cannot see: 17 stale anchors

`Validate` checks `clause` only. It never resolves `anchor:` against the named file's headings.
Resolving them independently: **17 of 101 entries name an anchor that matches no heading slug in
their own `file:`**, and **9 of those 17 have a passing clause** — so anchor rot is independent of
clause rot and would survive a clause-only repair untouched.

This is what the issue means by "a fix must update `file:`/`anchor:`, not only `clause:`":
`CLAUDE.md` §1 was compressed into three bullets and the verbatim doctrine for
`CONST-V3R2-008..011` now lives in `moai-constitution.md` under `## Response Language` /
`## Parallel Execution` / `## Output Format`. Nothing was lost; the pointer is simply stale.

## 5. Why it rotted silently: nothing runs `validate`

Neither guard invokes the checker that would have caught this:

```console
$ grep -n zone-registry Makefile
76:	MOAI_CONSTITUTION_REGISTRY=… ./bin/moai constitution list --format json > /dev/null

$ grep -n constitution .github/workflows/ci.yml
445:  constitution-check:
469:            ./bin/moai constitution list --format json | …
475:            ./bin/moai constitution list --zone frozen | tail -1
```

`make constitution-check` and the CI `constitution-check` job both run `constitution **list**`,
which only parses the registry. `constitution validate` is run by nothing — not the Makefile, not
CI, not any Go test. That absence is the reason the drift accumulated to two-thirds of the catalog
without a single red signal.

## 6. Consequence for the fix

Re-syncing `clause` to verbatim text is the minimum, but on its own it is a snapshot repair: the
matcher is exact, the rules text keeps moving, and nothing re-checks. The three things that
together make the repair hold:

1. `clause:` re-synced to a genuine verbatim substring of the named file, and `file:`/`anchor:`
   corrected where the doctrine moved.
2. A mechanical guard that fails on drift (a Go test over the shipped template + the local mirror,
   and/or `constitution validate` wired into the existing CI job), so the next rules edit that
   breaks an entry is caught by the PR that breaks it.
3. Anchor resolution actually checked — otherwise the 9 clause-OK/anchor-broken entries stay broken
   and invisible.

### The mutant this card must not accept

An implementation that reaches "validate passes" by **weakening the matcher** — fuzzy or
token-overlap matching, skipping short clauses, dropping entries, or flipping `canary_gate` — would
satisfy an acceptance criterion phrased as "validate exits 0" while leaving every stale pointer in
place and destroying the drift detector itself. Acceptance must pin the matcher's behavior
unchanged (its own tests still pass, byte-identical semantics) *and* require each clause to be a
verbatim substring, verified by an implementation independent of the one being repaired.

## Evidence files in this directory

| File | What it is |
|---|---|
| `validate-repro.txt` | verbatim `moai constitution validate` output from the fresh project |
| `analysis-repro.json` | per-entry clause/anchor verdicts, fresh project |
| `analysis-devrepo.json` | same, against this worktree (identical results — mirror is byte-identical) |
| `analyze.py` | independent re-implementation of the DRIFT rule + anchor resolver |
| `reconcile.py` | set-comparison of analyzer verdicts vs validator output |
