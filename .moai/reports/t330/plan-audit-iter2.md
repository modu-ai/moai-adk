# SPEC Review Report: SPEC-TODO-DESTRUCTIVE-GUARD-001 (iteration 2)

Card: t330 · Tree: `.claude/worktrees/t330` · Branch `WT-todo-destructive-guard` · HEAD `812ee01fc`
SPEC version audited: **0.2.0** (16 REQ / 16 AC / 10 files)
Iteration: **2/2** (Tier M ceiling — this is the last iteration)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.9375** — clears the Tier M threshold of 0.80
Movement: **0.75 → 0.9375, monotonic (+0.1875). No dimension regressed.**

Reasoning context ignored per M1 Context Isolation. Scoped to the iter1 defect delta plus the named regression surface, per the Retry Loop Contract; Decision 1 and the iter1-cleared items were re-confirmed by spot check, not re-derived.

---

## Must-Pass Results (re-run against v0.2.0)

- **[PASS] MP-1** — RAN `grep -o "REQ-TDG-[0-9]*" spec.md | sort -u` → REQ-TDG-001..016, sequential, no gaps, no duplicates, uniform padding. Same for AC-TDG-001..016.
- **[PASS] MP-2** — the two new requirements match GEARS: REQ-TDG-015 (`spec.md:189`) is Ubiquitous + Where (`Where the exported record carries archived rows, the command shall disclose…`); REQ-TDG-016 (`:190`) is Event-driven (`When a card is restored, its archive entry shall be removed…`). The other fourteen are unchanged from iter1 and were cleared there.
- **[PASS] MP-3** — READ `spec.md:1-19`: all 12 canonical fields present, `version: "0.2.0"` quoted, `updated: 2026-08-28` ISO, no rejected snake_case alias.
- **[N/A] MP-4** — single-language project SPEC; unchanged.
- **[PASS] MP-5 (D7 cross-SPEC)** — the three `related_specs` are unchanged from iter1 and were measured this session: `in-progress` / `completed` / `in-progress`. None retired, superseded, or archived.
- **[PASS] MP-6 (D8)** — RAN `grep -c syscall spec.md` → `0`. Auto-PASS.
- **[PASS] MP-7** — RAN `grep -c 'NEEDS CLARIFICATION' plan.md` → `0`. No `research.md` in the Tier M set.

---

## Category Scores

| Dimension | iter1 | iter2 | Movement | Basis |
|---|---|---|---|---|
| Clarity | 0.75 | **1.00** | +0.25 | D1 and D3 both fully resolved; no requirement carries an interpretation a reasonable engineer would resolve differently. |
| Completeness | 1.00 | **1.00** | — | All sections, frontmatter, and four `### Out of Scope` H3s retained; §C.5 added with two new bullets in §D. |
| Testability | 0.50 | **0.75** | +0.25 | All three non-executable ACs are now executable. Two criteria retain a minor precision gap (N1, N2 below), so not 1.0. |
| Traceability | 0.75 | **1.00** | +0.25 | The iter1 dock was for REQ-TDG-006/007 being covered only by non-executable ACs; both now execute, so coverage is real. Table at `spec.md:239-253` maps all 16→16 correctly. |

Aggregate = **0.9375**. Threshold 0.80 → **clears**. Monotonic against iter1 in every dimension.

---

## Regression Check (iter1 defects)

| # | iter1 defect | Status | Evidence |
|---|---|---|---|
| D1 | §A.4 "already true" unconditioned, contradicting plan.md | **RESOLVED** | RAN `grep -n "already true"` over all three files → two hits only, both at `spec.md:91` and `plan.md:69`, and both are explicit *cautions* quoting the incomplete form and naming its missing condition. The two-mode table (`spec.md:81-85`, mirrored `plan.md:49-52`) states Mode 1 false / Mode 2 true with the ref each depends on, and `spec.md:87` relocates the ruling onto "the predicate answering the wrong question". `plan.md:69` additionally records that the condition was dropped once before — the fix is durable, not cosmetic. |
| D2 | export-json vs AC-TDG-007 collision | **RESOLVED** | Decision 3 (`spec.md:194-201`) rules inclusion; REQ-TDG-007 (`:172`) narrowed to `list`/`next`/`why`/`analyze`/state counts with an explicit carve-out; AC-TDG-007 (`:85`) records the exclusion and points at AC-TDG-015 for the opposite direction; M5 (`plan.md:107-116`) budgets the disclosure and states the two directions are opposite on purpose. Assessed on merits below. |
| D3 | "both live backends" framing; three unverifiable ACs | **RESOLVED** | RAN `grep -n "legacy backend\|both backends\|legacy engine\|record file itself"` → three hits, all *corrective* prose (`spec.md:170`, `acceptance.md:29`, `:78`) explaining why the pair is unreachable. No legacy-engine verification arm survives. AC-TDG-003 (`:56`) now reads the archive table and the export field, both reachable; AC-TDG-006 (`:73-78`) became migration-survival; §A.3 (`:27-29`) names `export-json` as the single serialization surface. |
| D4 | downgrade archive loss unaddressed | **RESOLVED** | REQ-TDG-015 second clause + AC-TDG-015 + `spec.md:199` + a dedicated `§D` exclusion at `:232`. The gap is now ruled rather than silent. |
| D5 | verb surface omitted `why` | **RESOLVED** | `spec.md:36-41` lists fifteen and names `AddCommand` at `todo.go:137-141` as the authority; AC-TDG-001 (`acceptance.md:45`) instructs the verifier to read that call or count `--help` output rather than transcribe the doctrine table. I re-confirmed the registration at `internal/cli/todo.go:137-141`. |
| D6 | `--expect` holder set | **RESOLVED — and my iter1 framing was wrong; the correction is right** | RAN `sed -n '145,155p' internal/cli/todo_edit_move.go` → `move` declares `--top`, `--bottom`, `--before`, `--after` and **no** `--expect`. The corrected set at `spec.md:139` (`next`/`edit`/`drop`/`undrop`) is correct. See N3 for a small slip inside the correction. |
| D7 | citation drift | **PARTIALLY RESOLVED — two survive; see S1** | `prlink_landed.go:27`→28 and `backlog_sqlite.go:243-247`→251-253 are fixed everywhere; §H (`plan.md:170-182`) was rebuilt and I spot-checked each line list. Two acceptance.md citations were not updated. |
| D8 | `undone` archive-entry lifecycle unasserted | **RESOLVED, and the reasoning is sound** | REQ-TDG-016 (`spec.md:190`) + AC-TDG-016 (`acceptance.md:141-145`). The stated derivation — "forced by REQ-TDG-002: a retained entry would break byte-identity and permit a double restore" (`spec.md:201`) — is correct given §A.3's oracle: the byte-identity comparison now runs against `export-json`, which under Decision 3 carries the archive, so a retained entry genuinely changes the compared bytes. The two rulings interlock rather than merely coexist. AC-TDG-016's second-`undone` assertion is verifiable and consistent with AC-TDG-011, which already lists "`undone` of an id never archived" as a refusal path. |

---

## Assessment of Decision 3 (judged on its merits, not as a patch)

**The ruling is sound and I would not overturn it.** Its three load-bearing claims each hold against measurement:

- READ `internal/cli/todo_export.go:1-11` — `export-json` is documented as "the deliberate downgrade route back to plain JSON", the swap is one-way "with no config knob selecting an engine", and "that older binary reads only the json filename". The premise that the export *is* what the previous release serves is the file's own stated contract, not an inference.
- READ `internal/cli/todo_export.go:69` — `json.MarshalIndent(rec, "", "  ")`. Inclusion is what the existing code does with any top-level field, so exclusion is the branch that costs work. `spec.md:198`'s framing — "real work to make the product worse" — is accurate, and `plan.md:112`'s insistence that inclusion is a *ruling* rather than an accident is the right defensive move, backed by a risk row (`plan.md:164`) so a later reader who "tidies the leak away" fails AC-TDG-015 instead of passing silently. That is a well-constructed trap.
- The residual loss is genuinely outside reach: Go's `encoding/json` drops unknown keys on unmarshal, and nothing in this SPEC executes inside the older binary.

**Is "make the loss loud at the one point we control" adequate?** As a *category* of response, yes — it is the same standard §A.2 sets when it calls a quiet loss worse than a loud one, applied consistently. But the claim that export time is "the one point in the sequence we control" (`spec.md:199`, `plan.md:114`) is **stronger than the facts support**, and the overreach costs a cheap improvement:

The exported **artifact** is also under our control, and unlike a terminal line it is the thing that actually crosses into the downgrade. A top-level warning string carried beside the archive in the exported JSON would ride the same whole-record marshal at no implementation cost, survive terminal scrollback, and reach whoever opens the file — plausibly a different operator at a much later time than the export. The stderr line fires at export; the loss occurs at the old binary's first write, and the gap between those two moments is exactly where a scrollback-bound warning fails. The old binary would drop that key too, but a human inspecting the file before downgrading would not.

I am raising this as a **strengthening option (N4, optional)**, not a defect: the ruling as written is coherent, honest about its limits, and implementable. But the sentence asserting export is the *only* controllable point should be softened, because it is the kind of overclaim that forecloses the better option.

**Is the stderr disclosure verifiable?** Only partially — see N1. This one is a real gap.

---

## Surviving and newly-introduced defects

**S1 — `acceptance.md:51` and `acceptance.md:111` — two iter1 D7 citations survive, and the progress record reports them fixed — Severity: minor — Class: blocking**

VERIFIED BY RUNNING `sed -n '344,356p' internal/cli/todo.go` with offsets computed against the base:

- `rec.RemoveFindingsNaming(id)` is at **`todo.go:347`**. `spec.md:58` was corrected to 347; `acceptance.md:51` still reads `internal/cli/todo.go:341`.
- The refused-mutation block is the comment at 351-352 plus `return fmt.Errorf("no backlog item %s", id)` at **353** — i.e. **351-353**. `plan.md:76` was corrected to 351-353; `acceptance.md:111` still reads `todo.go:352-354` (which drops the first comment line and picks up the closing `}); err != nil`).

`progress.md:37` states "D7 (four citations refreshed…)". Four *were* refreshed — in spec.md, plan.md, and §H — but the two in acceptance.md were not, and the record does not distinguish. That is a small unobserved-completion claim on the very defect class D7 named, which is why this is classed blocking despite being minor: two one-line edits, plus a corrected disposition line.

**N1 — `acceptance.md:137` (AC-TDG-015) — the criterion cannot verify the stream REQ-TDG-015 specifies — Severity: minor — Class: blocking**

REQ-TDG-015 (`spec.md:189`) requires the disclosure **on stderr**. AC-TDG-015 captures `out=$(moai todo export-json 2>&1); rc=$?`, which merges both streams, so an implementation printing the disclosure to *stdout* passes the criterion unchanged.

The stream is load-bearing, not stylistic: READ `internal/cli/todo.go:20-22` — the package contract is "one structured stdout line out, human-readable errors on stderr", and READ `todo_export.go:81-82` — `export-json`'s single stdout line (`exported %d cards, %d findings to %s`) is that structured line. A disclosure on stdout pollutes a surface agents parse. Fix: capture the streams separately, e.g. `err=$(moai todo export-json 2>&1 >/dev/null)`, and assert the disclosure text in `err`.

**N2 — `acceptance.md:82-83` (AC-TDG-007) — the `why` observable is not mechanically checkable as written — Severity: minor — Class: blocking**

AC-TDG-007 adds `moai todo why t1` to the live-queue readers and asserts "none of them names `t1` as a live card". VERIFIED BY READING `internal/cli/todo_why.go:33-37`: when `rec.FindingsNaming(id)` returns empty, `why` prints `"%s: no findings\n"` — it **always echoes the id it was given**. A mechanical `grep t1` over `why`'s output therefore matches on an archived card, and the criterion's assertion is saved only by the reader's interpretation of "as a live card".

The semantics are right; the observable needs sharpening. Fix: assert `why t1` prints exactly `t1: no findings` (equivalently, zero finding lines) rather than that it does not name `t1`. The other three readers in the list are unaffected.

**N3 — `spec.md:139` — the `move` flag list inside the D6 correction is itself incomplete — Severity: minor — Class: optional**

The corrected C2 says `move` "takes only `--top`/`--bottom`". RAN `sed -n '145,155p' internal/cli/todo_edit_move.go` → it declares four flags: `--top`, `--bottom`, `--before`, `--after`. The load-bearing claim (that `move` does not carry `--expect`) is correct and the argument is unaffected; only the parenthetical enumeration understates. `plan.md:180` avoids the slip by asserting only the absence.

**N4 — `spec.md:199` and `plan.md:114` — "the one point we control" overclaims — Severity: minor — Class: optional**

See the Decision 3 assessment above. The exported artifact is also controllable and is the carrier that actually crosses into the downgrade. Recommend softening the sentence and recording the in-artifact warning as a considered-and-deferred option, so a later reader does not treat the stronger claim as settled.

---

## Regression surface — confirmed unchanged

Confirmed by reading, not by taking the report.

- **Decision 1 and its grounding** — `spec.md:101-129` is textually unchanged from v0.1.0, including M1/M2/M3 and the two consequence bullets. Its grounding was verified in iter1 (`ensureSchema` executes the whole `IF NOT EXISTS` DDL on every open at `backlog_sqlite.go:235`; the CHECK at `:100`) and is not re-derived here. Not reopened.
- **The `schema_version` freeze** — REQ-TDG-005 (`spec.md:167`) and `spec.md:128` unchanged; AC-TDG-005's citation now correctly points at the `default:` abort, which I re-measured at **`backlog_sqlite.go:251-253`**.
- **Findings-restoration teeth** — AC-TDG-002 unchanged, and now *stronger* rather than merely intact: §A.3's oracle is `export-json`, which marshals the whole record including `Findings`, and REQ-TDG-016 keeps the archive from silently absorbing the difference.
- **AC-TDG-004 / AC-TDG-010 self-flagged debts** — both criteria textually unchanged (`acceptance.md:61-65`, `:101-105`). My iter1 judgment stands: the freeze assertion is legitimately base-identical, and AC-TDG-010's pairing requirement is acceptable debt.
- **Non-interactivity** — REQ-TDG-012, AC-TDG-012, `plan.md:75` unchanged. Nothing in Decision 3 introduces a prompt: a stderr write is not an interaction.
- **§A.1 / §A.2 standing preconditions** — unchanged and still `[HARD]` at the head of `acceptance.md`, restated at `plan.md:135-147`. Every non-zero-exit criterion still uses `out=$(cmd 2>&1); rc=$?`.
- **Template-First** — `plan.md:126-133` still names both paths, `make build`, and `cmp`; AC-TDG-014 unchanged. The `cmp`-clean / 13709-byte measurement was verified in iter1 and the tree has not moved (`HEAD` still `812ee01fc`).

---

## Tier M at the ceiling — judgment

**Tier M holds.** Ten files (`plan.md:11-23`) sits inside the 5-15 band; none is constitutional; the storage change is additive by ruling; there is no cross-subsystem redesign. The single added file (`internal/cli/todo_export.go`, for the disclosure) is a small edit to an existing file, not new surface.

**The 16/16 ceiling is a forward-looking split signal, not a present defect**, and the plan already says so in exactly those terms (`plan.md:24`: "any further growth is a signal to split, not to relax the budget"). Declaring the constraint rather than quietly sitting on it is the correct handling.

One caution for the run phase: **all three debt items below are edits to existing criteria, so none consumes budget.** If a fix is ever proposed that would require a 17th REQ or AC, that is the split signal firing — take it rather than relaxing the ceiling.

---

## Recommendation

**PASS-WITH-DEBT at the Tier M iteration ceiling (2/2).** The SPEC is implementable as it stands. All five iter1 blocking defects are genuinely resolved — I verified each fix rather than accepting the report — and Decision 3, made under audit pressure, holds up on its own merits: its premises are the exporter's own documented contract, its reasoning about what lies outside reach is correct, and it protects itself with a positive assertion and a risk row.

**No third audit iteration is warranted.** The residual debt is three one-line edits the lead can confirm by inspection:

1. **S1** — `acceptance.md:51`: `todo.go:341` → **347**. `acceptance.md:111`: `todo.go:352-354` → **351-353**. Correct `progress.md:37`'s D7 disposition to say which files were refreshed.
2. **N1** — `acceptance.md:137`: capture stderr separately so the stream REQ-TDG-015 specifies is actually asserted.
3. **N2** — `acceptance.md:82-83`: assert `why t1` prints `t1: no findings` rather than that it does not name `t1`.

Optional, at the author's discretion: **N3** (complete the `move` flag list) and **N4** (soften "the one point we control" and record the in-artifact warning as considered-and-deferred).

With items 1-3 landed, this is a clean PASS.
