# t189 — `TestConstitutionCrossReference` regression on release/v3.1.3

Card class: urgent batch blocker, Tier S. Fixed directly on `release/v3.1.3` (no worktree,
per the lead's dispatch).

---

## Claim

1. `TestConstitutionCrossReference` (`internal/cli/agentlint/agent_lint_test.go:1218`) failed
   on `release/v3.1.3` and passes now.
2. The cause is the one the card names: the diet that reduced `moai-constitution.md`
   replaced a direct pointer to `agent-authoring.md` with a paraphrase inside the
   detail-companion line, and the test asserts on the literal filename.
3. The repair restores a direct, resolvable pointer rather than relaxing the assertion.
4. Local and template mirrors stay byte-identical.

---

## Evidence

### The failure, reproduced before the fix

```
$ go test ./internal/cli/agentlint/ -run TestConstitutionCrossReference -count=1
--- FAIL: TestConstitutionCrossReference (0.00s)
    agent_lint_test.go:1249: moai-constitution.md should cross-reference agent-authoring.md for effort matrix
    agent_lint_test.go:1257: Checked moai-constitution.md at: …/release-v313/.claude/rules/moai/core/moai-constitution.md
FAIL	github.com/modu-ai/moai-adk/internal/cli/agentlint	0.424s
```

`grep -c 'agent-authoring' .claude/rules/moai/core/moai-constitution.md` → `0` on this
branch. On `origin/main` the same grep finds the line at `:50`:

> `- Per-agent effort calibration: see `.claude/rules/moai/development/agent-authoring.md` § Effort-Level Calibration Matrix for the retained-agent default-effort table and the archived-agent legacy reference.`

### The change

The release-side text carried the pointer only as a paraphrase —
"Rationale, **the per-agent effort-calibration matrix pointer**, and the model-id table:
`moai-constitution-detail.md` …" — naming the companion but not the file the reader needs.
Split into two sentences so the destination is named:

```
Per-agent effort calibration: `agent-authoring.md` § Effort-Level Calibration Matrix.
Rationale and the model-id table: `moai-constitution-detail.md` § Opus 5 / 4.8 Prompt
Philosophy.
```

Applied identically to `.claude/rules/moai/core/moai-constitution.md` and its template
mirror. The destination is real: `moai-constitution-detail.md:36` cites the same
`agent-authoring.md` § Effort-Level Calibration Matrix, so the constitution now points at
the section directly instead of only through the companion.

### Verification batch (this tree)

| Command | Observed |
|---|---|
| `go test ./internal/cli/agentlint/ -count=1` | `ok … 5.397s` (whole package, not just the one test) |
| `go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$'` | `67025 tokens (budget 76000, headroom 8975, 18 entries)` — PASS |
| `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -count=1` | `ok … 48.102s` (mirror parity + neutrality) |
| `git status --short` | only the two `moai-constitution.md` files modified |

---

## Baseline-attribution

All four rows were measured by me in `.claude/worktrees/release-v313` on branch
`release/v3.1.3`, on the tree carrying the final two-file diff. The RED transcript is the
same tree with the fix hunk absent.

**On the budget arithmetic the card asks to be recorded**: the always-loaded surface reads
67,025 tokens here against a 76,000 budget. The card's t82 ratchet figure (66,371) is
exceeded by 654 — and that overage predates this card: the immediately-preceding measurement
on this branch, taken during t183 before either constitution file was touched, was 67,016.
This fix accounts for 9 of those tokens; the other 645 arrived before it. The ratchet is not
the guard that runs in CI — `TestAlwaysLoadedTokenBudget` checks against 76,000, and it
passes.

---

## Gaps — what was NOT observed

- **No full-suite run.** Only `internal/cli/agentlint`, `internal/config` (one test), and
  `internal/template` were run. The change is two prose lines in a markdown rule file with no
  Go code touched, so no other package can observe it; CI on the batch PR is the full-suite
  measurement.
- **Whether other assertions lost pointers in the same diet commit was not swept.** This card
  fixes the one regression it names. `243eb07ef` reduced the file by ~3,500 characters and
  nothing here checks the rest of that reduction for other dropped cross-references — a
  broader sweep is a separate card, not silently in scope here.
- **The relax-the-assertion alternative was not built or measured.** The card offers it as a
  fallback if restoration is awkward. Restoration was not awkward (two lines, budget clear),
  so the comparison was not needed and no evidence about that option exists.
- **`make build` was not run.** The template mirror is a source-of-truth edit; the embedded
  copy recompiles from `templates/` at build time, and `internal/template` mirror-parity
  passed. No binary was reinstalled from this tree.

---

## Residual-risk

- **The pointer is now stated in two places** (the constitution and its detail companion), so
  a future rename of `agent-authoring.md` must update both. The test guards only the
  constitution's copy.
- **The test asserts a bare filename substring**, so it would also be satisfied by a mention
  that does not resolve — it detects deletion, not link rot. That weakness is unchanged by
  this card.
- **A future diet pass could drop the line again** by the same mechanism (paraphrasing a
  pointer while compressing). The test catches it, which is how this one surfaced.
