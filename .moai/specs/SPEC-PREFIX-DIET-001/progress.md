# progress.md — SPEC-PREFIX-DIET-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_artifacts: spec.md, plan.md, acceptance.md (Tier M), progress.md
plan_revision: 0.2.0 (plan-audit iter1 FAIL 0.78 + 리더 판정 D1~D3 반영)
plan_closed_decisions: D1 (skillListingBudgetFraction 불변), D2 (§6 핸드오프 동결 확정), D3 (구속 줄 재작성 없음) — `plan.md` §B
plan_open_decisions: D4 (축약 목표·설명 상한, 비차단), D5 (로컬↔템플릿 에이전트 설명 불일치 처리, 비차단) — `plan.md` §B
plan_known_gaps: `spec.md` §H

## §E.2 Run-phase Evidence

Run-phase owner: manager-develop (cycle_type tdd; DDD-style ledger preservation for output-style edits). Worktree `.claude/worktrees/t1450`, branch `WT-prefix-diet-stage2`. Anchor `5d5ff1aae`; run started at HEAD `e5523d672` (verified: `git rev-parse --short HEAD` -> `e5523d672`, `.moai/specs/SPEC-PREFIX-DIET-001/spec.md` present).

Command-form note (Gap, recorded once): the worktree-isolation guard refuses a quoted `|` inside a `go test -run` pattern, so the `\|`-joined patterns of AC-PFD-001/005/006 were run as prefix patterns (`-run '^TestOutputStyle'` selects the four new tests plus the five existing `TestOutputStyles*`) and each name is read from its own `--- PASS: <name>` line.

### M0 — measurement baseline and ledger front (no style/agent edit)

**Claim.** The anchor reproduces within 1% of 154,219 (REQ-PFD-012 stop condition NOT triggered); the ledger front, the extractor, the frozen/localization fixtures and the four guard tests exist and pass on the unedited tree; per-file reduction targets and the droppable totals are fixed below; the catalog.yaml side effect of an agent-description edit is observed and the guard allowlist is widened for it.

**Evidence — anchor first-turn tokens** (command, one plain invocation per run, cwd = worktree root, HEAD `e5523d672`, settings file `{"disableAllHooks": true}`):

```
claude -p ok --output-format json --model claude-opus-5-5 --settings <scratch>/hooks-off.json | grep -o '"usage":{[^}]*}'
```

(the `grep -o` stops at the first `}` so the nested `output_tokens_details` object is cut; the three counted fields are complete)

```
first-turn-input-tokens[anchor#1]=154235 head=e5523d672 usage={"input_tokens":2,"cache_creation_input_tokens":141834,"cache_read_input_tokens":12399,"output_tokens":365,"output_tokens_details":{"thinking_tokens":90
first-turn-input-tokens[anchor#2]=154235 head=e5523d672 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":154233,"output_tokens":377,"output_tokens_details":{"thinking_tokens":96
first-turn-input-tokens[anchor#3]=309692 head=e5523d672 usage={"input_tokens":4,"cache_creation_input_tokens":1222,"cache_read_input_tokens":308466,"output_tokens":679,"output_tokens_details":{"thinking_tokens":203   (OUTLIER: the run made two model calls, so the three fields sum two requests; excluded from min/median, kept for honesty)
first-turn-input-tokens[anchor#4]=154235 head=e5523d672 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":154233,"output_tokens":388,"output_tokens_details":{"thinking_tokens":58   (replacement run; `"num_turns":1` observed in the same JSON)
```

Anchor figure: 154,235 (three of four runs identical, relative difference from the leader's 154,219 = +16 tokens = 0.010%). The leader's definition (sum of the three `usage` fields) reproduces. The two-call outlier shows that a run's sum can double when the CLI makes a second request; later conditions are grepped together with `"num_turns"` so a doubled run is recognised and not mixed in.

**Evidence — baseline package state** (before any edit, HEAD `e5523d672`):

```
go test ./internal/template/ ./internal/config/ -count=1      (exit 0)
ok  	github.com/modu-ai/moai-adk/internal/template	135.858s
ok  	github.com/modu-ai/moai-adk/internal/config	10.811s
```

FAIL-name set at the anchor: empty. AC-PFD-014's pass condition is therefore "no `FAIL`/`ok`-missing line after the change".

**Evidence — sizes, binding tokens, units** (whole-file UTF-16; `python3 .moai/specs/SPEC-PREFIX-DIET-001/tools/diet_ledger.py sizes`, plus the AC-PFD-004 one-liner per file):

| file | whole-file UTF-16 | body UTF-16 | units | `[HARD]` / `MUST NOT` / `MUST` / `shall ` (AC-004 python) |
|---|---|---|---|---|
| moai.md | 62,593 | 62,218 | 251 | 89 / 4 / 27 / 0 |
| moai-easy.md | 29,243 | 28,739 | 185 | 33 / 0 / 0 / 0 |
| moai-learn.md | 28,517 | 28,122 | 143 | 24 / 0 / 7 / 0 |

These equal EL-1 and EL-10. The Go extractor (`internal/template/output_style_diet_helpers_test.go`, `dietExtractUnits`) and the Python extractor (`tools/diet_ledger.py`) segment all three bodies into the same units: the ledger test rebuilds each anchor body from the fixture rows and re-extracts it with the Go code; it passes (`unit count`, `unit text` and `body sha256` all agree), so the two implementations check each other.

**Evidence — ledger front.** `internal/template/testdata/output_style_ledger.json` carries 548 rows (moai.md 251-27 frozen = 224, moai-easy.md 185-4 frozen = 181, moai-learn.md 143; kinds: binding 28 / 6 / 10, rationale 9 / 31 / 9, example 0 / 28 / 0, the rest normative), every row `verbatim`, `after_text == before_text`, zero `rewrite` rows. Row kinds are set by the extractor for `binding` (any binding token in the unit) and by the author for the rest; the kinds `rationale` / `example` are assigned ONLY to the units planned for dropping (classification principle: a unit is `rationale`/`example` only when the same information survives in a named unit or an existing rule; reference material, install guides, pointers, templates, tables and anything procedural stay `normative`). Frozen sections (units excluded from the ledger, text in `output_style_frozen.json`): moai.md `### Session Boundary Handoff [HARD]` (units 66-69) and `### Session Handoff [HARD]` (units 216-238), moai-easy.md `### Banner 7 — Picking Up Next Time (Session Handoff)` (units 111-114). `moai-learn.md` contains no handoff/resume/cut-line text (`grep -n -i "handoff\|resume\|✂"` -> no output), so it has no equivalent section to freeze (recorded; nothing is frozen there). Localization tables (`output_style_localization.json`): every contiguous table whose header row names `Korean` or `ko canonical` (moai.md 11 tables, moai-easy.md 1, moai-learn.md 1).

**D4 — per-file reduction targets (REQ-PFD-002).** The droppable total is the UTF-16 length of the units classified `rationale`/`example` in M0; because rewrite is forbidden (D3) the target equals that total and replaces the draft numbers of plan.md section B:

| file | anchor | droppable total (rationale / example) | target | budget constant after the milestone |
|---|---|---|---|---|
| moai-easy.md | 29,243 | 7,893 (4,016 / 3,877) | -7,893 | 21,350 (draft was 21,000: lowered, not reachable without rewriting) |
| moai-learn.md | 28,517 | 1,507 (1,507 / 0) | -1,507 | 27,010 (draft was 20,000: lowered) |
| moai.md | 62,593 | 1,444 (1,444 / 0) | -1,444 | 61,149 (draft was 45,000: lowered) |

For the leader (REQ-PFD-002 report): the drafts of plan.md section B cannot be met under D3 for any file; the achievable reductions are 27.0% (moai-easy), 5.3% (moai-learn) and 2.3% (moai). The leader's confirmation of these targets is the open part of D4.

**D4/D5 — agent descriptions (input for M5).** Measured with the plan.md section C extractor: template total 11,155, local total 11,146 (matches EL-7). Largest: manager-lead 2,182, manager-docs 1,552, manager-develop 1,094, manager-spec 1,022. D5 input: `manager-git` (template 553 / local 533) and `manager-spec` (template 1,022 / local 1,033) differ between template and local copy; the other ten are identical. The M5 caps are fixed at M5 from these sizes.

**D1 (plan-audit debt) — observed side effect of an agent-description edit** (probe on a scratch edit of `manager-todo.md`, reverted with `git checkout --`; `git status --short` clean afterwards):

```
(probe) description line + " (probe)"  ->  make agents-emit                                       exit 0
go run ./internal/template/scripts/gen-catalog-hashes.go --all (the step `make build` runs)        exit 0
git status --short:
 M internal/template/catalog.yaml
 M internal/template/templates/.claude/agents/moai/manager-todo.md
 M internal/template/templates/.codex/agents/moai/manager-todo.toml
git diff -U0 internal/template/catalog.yaml:   one hunk, one line:  -  hash: c446ed4a…  +  hash: 5f9e9c99…   (the manager-todo entry, `path: templates/.claude/agents/moai/manager-todo.md`)
```

So an agent-description edit rewrites exactly one TOML and exactly one `hash:` line in `catalog.yaml` per edited agent. Output-style files are not catalogued (`grep -n "output-styles\|moai-easy\|moai-learn" internal/template/catalog.yaml` -> no match), so M2-M4 do not touch `catalog.yaml`. Guard change (same commit as this note): `surface_guard.py` gained the surface `catalog-hashes` for `internal/template/catalog.yaml`; it allows ONLY changed `hash:` lines whose entry `path:` is an agent template `.md` that is itself changed in the diff, and reports `VIOLATION catalog-change-beyond-edited-agent-hashes` otherwise. SPEC-text gap for the leader (spec.md/plan.md/acceptance.md bodies belong to manager-spec): REQ-PFD-013 lists a closed allowlist without `catalog.yaml`; plan.md section C (line 49) already provides the widen-and-record procedure, so no behaviour changed, but the REQ text and AC-PFD-012's positive-control list should add this surface.

**Guard at the anchor** (`python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py 5d5ff1aae`, exit 0): `ok docs …` for the SPEC files and tools, `ok tests …` for the three fixtures and two test files, last line `surface-guard=PASS`. `ruff check --select E,F,W --ignore E501` on both Python files: no findings.

**AC-PFD-011 RED-now re-run on the current tree** (document-pinned to `5d5ff1aae`, where `progress.md` does not exist — plan-audit D4):

```
grep -c 'first-turn-input-tokens\[' .moai/specs/SPEC-PREFIX-DIET-001/progress.md      (run at HEAD e5523d672, before this file's edit)
0
exit=1
```

The AC cell is not edited (manager-spec owns it); this line is the current-tree observation.

**Guard tests on the unedited tree** (`go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v`, exit 0): `--- PASS: TestOutputStylesCharBudget` (+3 subtests), `--- PASS: TestOutputStyleBindingLedger` (+9 mutation subtests: missing_anchor_unit_row, binding_unit_relabeled, binding_row_dropped, normative_row_dropped, dropped_without_note, dropped_without_survivor, verbatim_text_changed, dropped_with_tokens, file_total_token_lost, rewrite_row_rejected), `--- PASS: TestOutputStyleHandoffUnitsFrozen` (+ the one-character-mutation subtest), `--- PASS: TestOutputStyleLocalizationTableParity` (+ one_cell_mutation_is_caught, removed_row_is_caught), and the five pre-existing `TestOutputStyles*` all `--- PASS`.

**Carried SPEC-text debts (plan-audit iter3, not edited here):** D2 — AC-PFD-015 maps to no REQ; D3 — spec.md line 139 still says `AC-PFD-001 ~ AC-PFD-015` while acceptance.md carries 016.

**Baseline-attribution.** All measurements: this run, worktree `t1450`, HEAD `e5523d672`, anchor `5d5ff1aae` for the guard and the ledger (`git show 5d5ff1aae:<style path>` is the ledger source). Tool provenance (VCI 2.2): `moai`-built tooling was not used for any M0 measurement; Go tests ran from this tree's source with `go test`, Python tooling ran from this tree.

**Gaps.** The anchor token runs are single invocations under one account/time window (cache state varies: run #1 wrote cache, #2/#4 read it; the sum is stable). The leader's 154,219 itself is not re-attributed here, only reproduced within 0.010%. The ledger test checks that the deployed body equals the ledger-accounted body, but it cannot decide whether a unit the author labelled `rationale`/`example` is really rationale — see Residual-risk.

**Residual-risk.** Author-side misclassification of a `normative` unit as `rationale` is not machine-detectable; the full `dropped` list (source, reason, survivor) is exported in the final section of §E.2 for the sync-audit and the leader.

### M1 — `skillListingBudgetFraction` key guard (no change; operator decision item)

**Claim.** The key stays `0.02` in both settings files and neither file is modified; the `0.01` measurement is recorded as an OPERATOR DECISION ITEM only (leader decision D1).

**Evidence.**

```
grep -n '"skillListingBudgetFraction"' internal/template/templates/.claude/settings.json.tmpl .claude/settings.json
internal/template/templates/.claude/settings.json.tmpl:414:  "skillListingBudgetFraction": 0.02,
.claude/settings.json:410:  "skillListingBudgetFraction": 0.02,

git diff --numstat 5d5ff1aae HEAD -- internal/template/templates/.claude/settings.json.tmpl .claude/settings.json        (exit 0, no output)
```

Operator-item measurement, re-run on this tree (settings file `{"disableAllHooks": true, "skillListingBudgetFraction": 0.01}`, HEAD `089d6fe51`, same command form as M0):

```
first-turn-input-tokens[budget0.01#1]=147310 head=089d6fe51 usage={"input_tokens":2,"cache_creation_input_tokens":134909,"cache_read_input_tokens":12399,"output_tokens":415,"output_tokens_details":{"thinking_tokens":124   ("num_turns":1)
```

147,310 vs this run's anchor 154,235 = -6,925 (-4.5%); the leader's carried figure was 147,025 / -7,194 (-4.7%). One run, variance unknown. Effect on skill-discovery quality: NOT measured. Nothing was changed; the decision belongs to the operator.

### M2 — `moai-easy.md` (default style; the only milestone with a default-user token effect)

**Claim.** `moai-easy.md` is reduced from 29,243 to 21,350 UTF-16 units (-7,893, the whole droppable total) by dropping `rationale`/`example` units only; every binding token is retained; the frozen Banner 7 section and the localization table are byte-identical; template and local copy are identical; the first-turn input tokens fall by 2,828 (-1.83%) under the default `MoAI-Easy` style.

**RED** (budget constant lowered to 21,350 before the edit, `go test ./internal/template/ -run 'TestOutputStylesCharBudget' -count=1 -v`, exit 1; the `$`-anchored pattern is refused by the guard, so the name is unanchored and selects only this test):

```
    output_style_diet_test.go:46: output-style=moai-easy 29243
    output_style_diet_test.go:48: output style moai-easy.md is 29243 UTF-16 units, over its budget 21350
--- FAIL: TestOutputStylesCharBudget (0.00s)
    --- PASS: TestOutputStylesCharBudget/moai
    --- FAIL: TestOutputStylesCharBudget/moai-easy (0.00s)
    --- PASS: TestOutputStylesCharBudget/moai-learn
FAIL	github.com/modu-ai/moai-adk/internal/template	0.420s
```

**GREEN.** Template edit by `python3 .moai/specs/SPEC-PREFIX-DIET-001/tools/diet_ledger.py apply moai-easy.md` (output `apply ok moai-easy.md utf16 29243 -> 21350`); the same call flips the planned rows to `dropped` in the ledger with reason + survivor. Dropped units (ids `moai-easy-NNNN`): 44-45, 53-71, 72-73, 134-142, 158-184 (the `---` units 142, 158, 169, 174, 182 are separators of dropped sections). Then `make build` (exit 0; `catalog.yaml` unchanged, `git status --short` shows only the three files below), local mirror `cp` of the template file, and:

```
diff -rq internal/template/templates/.claude/output-styles .claude/output-styles        (no output, exit 0)
go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v                        (exit 0)
--- PASS: TestOutputStylesCharBudget (output-style=moai 62593 / moai-easy 21350 / moai-learn 28517)
--- PASS: TestOutputStyleBindingLedger        --- PASS: TestOutputStyleHandoffUnitsFrozen
--- PASS: TestOutputStyleLocalizationTableParity
--- PASS: TestOutputStylesExactlyThree / FallbackDocsContract / Encoding / FrontmatterSchema / TemplateLiveParity
```

Binding tokens after the edit: the ledger test's `TOKEN_TOTAL` check equals the anchor (`[33 0 0 0]` for moai-easy.md). Guard (`surface_guard.py 5d5ff1aae`, exit 0): `ok output-styles` for the template file and the local copy, `surface-guard=PASS`. Gap-free side note: the edited file ends with one extra blank line (the unit preceding the dropped tail keeps its own trailing blank line — rewrite is forbidden, so it stays).

**First-turn tokens, default style MoAI-Easy** (HEAD `089d6fe51` + the M2 edit in the worktree; same command as M0; the `num_turns` field is read so a multi-call run is not mixed in — three multi-call runs (2, 2 and 3 calls) were observed and excluded, listed for honesty):

```
first-turn-input-tokens[M2#1]=151407 usage={"input_tokens":2,"cache_creation_input_tokens":139006,"cache_read_input_tokens":12399,"output_tokens":498,"output_tokens_details":{"thinking_tokens":113   ("num_turns":1)
first-turn-input-tokens[M2#2]=303991 usage={"input_tokens":4,"cache_creation_input_tokens":1177,"cache_read_input_tokens":302810,"output_tokens":839,"output_tokens_details":{"thinking_tokens":155   ("num_turns":2; excluded)
first-turn-input-tokens[M2#3]=304044 usage={"input_tokens":4,"cache_creation_input_tokens":1231,"cache_read_input_tokens":302810,"output_tokens":871,"output_tokens_details":{"thinking_tokens":197   ("num_turns":2; excluded)
first-turn-input-tokens[M2#4]=151407 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151405,"output_tokens":592,"output_tokens_details":{"thinking_tokens":108   ("num_turns":1)
first-turn-input-tokens[M2#5]=304201 usage={"input_tokens":4,"cache_creation_input_tokens":1387,"cache_read_input_tokens":302810,"output_tokens":909,"output_tokens_details":{"thinking_tokens":179   ("num_turns":3; excluded)
first-turn-input-tokens[M2#6]=151407 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151405,"output_tokens":386,"output_tokens_details":{"thinking_tokens":109   ("num_turns":1)
```

Single-call runs: min = median = max = 151,407. Anchor single-call runs: 154,235 (x3). Reduction = 2,828 tokens (-1.83%); spread 0, so the effect is measurable. (Run #2/#3 were issued in one parallel turn by mistake; each is a separate plain invocation and the account cache is shared, which does not change the summed fields.) The upper bound for this surface was -11,033 (`outputStyle=default`, whole body removed); this milestone realises 25.6% of it without rewriting any binding line.

**Baseline-attribution.** This run, worktree `t1450`; anchor sizes from `5d5ff1aae`; token baseline = M0 anchor runs in this section, same command and settings file. Tool provenance: `make build` produced `bin/moai` from this tree (`-X …Commit=089d6fe51`); no measurement in this milestone relied on an installed `moai` binary.

**Gaps.** Single account/time window; cache state differs between runs (the first run wrote the cache). The M2 conclusion rests on three single-call runs of equal value. `moai-easy.md` is what a default-style user loads; users who selected `MoAI` or `MoAI-Learn` see no change from this milestone.

**Residual-risk.** The dropped FAQ/philosophy/quick-reference/example units are classified `rationale`/`example` by the author; the survivor named per row is where the information remains. Dropped section numbers leave gaps (`§10`, `§12`-`§15`) because renumbering would rewrite kept lines; no kept unit references a dropped section number (checked with `grep -n "§1[0-5]"`).

### M3 — `moai-learn.md`

**Claim.** `moai-learn.md` is reduced from 28,517 to 27,010 UTF-16 units (-1,507, the whole droppable total); all 24 `[HARD]` and 7 `MUST` tokens retained; the localization catalogue table is cell-identical; the file has no handoff section to freeze; template and local copy are identical. Under `outputStyle=MoAI-Learn` the first-turn input tokens fall by 434 (-0.28%). The default-style user is unaffected.

**RED** (budget constant lowered to 27,010 before the edit, `go test ./internal/template/ -run 'TestOutputStylesCharBudget' -count=1 -v`, exit 1):

```
    output_style_diet_test.go:46: output-style=moai-learn 28517
    output_style_diet_test.go:48: output style moai-learn.md is 28517 UTF-16 units, over its budget 27010
    --- FAIL: TestOutputStylesCharBudget/moai-learn (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/template	0.391s
```

**GREEN.** `diet_ledger.py apply moai-learn.md` -> `apply ok moai-learn.md utf16 28517 -> 27010`. Dropped units (`moai-learn-NNNN`): 5-7, 122, 137-141. `make build` exit 0 (no tracked file other than the edited ones changed), local mirror `cp`, `diff -rq internal/template/templates/.claude/output-styles .claude/output-styles` -> no output, exit 0; `go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v` exit 0 with `--- PASS` for `TestOutputStylesCharBudget` (moai-learn 27010), `TestOutputStyleBindingLedger`, `TestOutputStyleHandoffUnitsFrozen`, `TestOutputStyleLocalizationTableParity` and the five existing `TestOutputStyles*`. Guard `surface_guard.py 5d5ff1aae` exit 0, `surface-guard=PASS`.

**First-turn tokens under `outputStyle=MoAI-Learn`** (settings file `{"disableAllHooks": true, "outputStyle": "MoAI-Learn"}`, a measurement input only; same command; every run below had `"num_turns":1`). Before = HEAD `c30c7132a` (learn file unedited), after = the M3 edit in the worktree:

```
first-turn-input-tokens[M3-before#1]=154999 usage={"input_tokens":2,"cache_creation_input_tokens":142598,"cache_read_input_tokens":12399,"output_tokens":472,"output_tokens_details":{"thinking_tokens":163
first-turn-input-tokens[M3-before#2]=154999 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":154997,"output_tokens":390,"output_tokens_details":{"thinking_tokens":84
first-turn-input-tokens[M3-before#3]=154999 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":154997,"output_tokens":510,"output_tokens_details":{"thinking_tokens":86
first-turn-input-tokens[M3#1]=154565 usage={"input_tokens":2,"cache_creation_input_tokens":142164,"cache_read_input_tokens":12399,"output_tokens":614,"output_tokens_details":{"thinking_tokens":207
first-turn-input-tokens[M3#2]=154565 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":154563,"output_tokens":544,"output_tokens_details":{"thinking_tokens":139
first-turn-input-tokens[M3#3]=154565 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":154563,"output_tokens":489,"output_tokens_details":{"thinking_tokens":125
```

Before min/median/max 154,999 / 154,999 / 154,999; after 154,565 x3; reduction 434 tokens (-0.28%), spread 0 on both sides, so the effect is measurable but small (it is the smallest droppable total of the three styles). This is a style-specific measurement: a default-style session does not load this file.

**Baseline-attribution.** This run, worktree `t1450`, the before/after runs ten minutes apart on one account, same command and settings file. **Gaps.** One account window; learn's droppable total is only 5.3% of the file because the remaining units are binding or normative (Notion install guide, the localization catalogue, response templates). **Residual-risk.** The dropped teaching-philosophy and root-cause units are author-classified rationale; survivors are named per row.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
