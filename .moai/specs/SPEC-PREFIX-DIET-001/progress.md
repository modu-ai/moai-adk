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

**Guard tests on the unedited tree** (`go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v`, exit 0): `--- PASS: TestOutputStylesCharBudget` (+3 subtests), `--- PASS: TestOutputStyleBindingLedger` (+10 mutation subtests: missing_anchor_unit_row, binding_unit_relabeled, binding_row_dropped, normative_row_dropped, dropped_without_note, dropped_without_survivor, verbatim_text_changed, dropped_with_tokens, file_total_token_lost, rewrite_row_rejected), `--- PASS: TestOutputStyleHandoffUnitsFrozen` (+ the one-character-mutation subtest), `--- PASS: TestOutputStyleLocalizationTableParity` (+ one_cell_mutation_is_caught, removed_row_is_caught), and the five pre-existing `TestOutputStyles*` all `--- PASS`.

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

### M4 — `moai.md` (89 `[HARD]` lines, two frozen handoff sections)

**Claim.** `moai.md` is reduced from 62,593 to 61,149 UTF-16 units (-1,444, the whole droppable total, 2.3%): the draft target of 45,000 is not reachable without rewriting binding lines (REQ-PFD-002: the target was lowered to the droppable total and is reported to the leader). All 89 `[HARD]`, 4 `MUST NOT` and 27 `MUST` tokens retained; the two frozen handoff sections (`### Session Boundary Handoff [HARD]`, `### Session Handoff [HARD]`) and all 11 localization tables are byte/cell-identical; template and local copy are identical. Under `outputStyle=MoAI` the first-turn input tokens fall by 453 (-0.27%).

**RED** (budget constant lowered to 61,149 before the edit, exit 1): `output-style=moai 62593`, `output style moai.md is 62593 UTF-16 units, over its budget 61149`, `--- FAIL: TestOutputStylesCharBudget/moai`.

**GREEN.** `diet_ledger.py apply moai.md` -> `apply ok moai.md utf16 62593 -> 61149`. Dropped units (`moai-NNNN`): 6-7 (Core Traits), 64-65 (the persistence provenance sentence and the auto-memory note), 246-250 (`---` + section 12 Service Philosophy). The frozen sections (units 66-69 and 216-238) lie outside every dropped range; unit 65 sits directly in front of the frozen heading and is a separate unit, so the frozen hash is unchanged (`--- PASS: TestOutputStyleHandoffUnitsFrozen`). `make build` exit 0, local mirror `cp`, `diff -rq` empty (exit 0), `go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v` exit 0, guard `surface_guard.py 5d5ff1aae` exit 0 `surface-guard=PASS`.

**First-turn tokens under `outputStyle=MoAI`** (settings file `{"disableAllHooks": true, "outputStyle": "MoAI"}`, a measurement input only; before = HEAD `576ba50ff`, moai.md unedited):

```
first-turn-input-tokens[M4-before#1]=168841 usage={"input_tokens":2,"cache_creation_input_tokens":52821,"cache_read_input_tokens":116018,"output_tokens":359,"output_tokens_details":{"thinking_tokens":100   ("num_turns":1)
first-turn-input-tokens[M4-before#2]=168841 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":168839,"output_tokens":346,"output_tokens_details":{"thinking_tokens":110   ("num_turns":1)
first-turn-input-tokens[M4-before#3]=168841 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":168839,"output_tokens":327,"output_tokens_details":{"thinking_tokens":95   ("num_turns":1)
first-turn-input-tokens[M4#1]=168388 usage={"input_tokens":2,"cache_creation_input_tokens":155987,"cache_read_input_tokens":12399,"output_tokens":482,"output_tokens_details":{"thinking_tokens":73   ("num_turns":1)
first-turn-input-tokens[M4#2]=168388 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":168386,"output_tokens":600,"output_tokens_details":{"thinking_tokens":154   ("num_turns":1)
first-turn-input-tokens[M4#3]=337964 usage={"input_tokens":4,"cache_creation_input_tokens":1188,"cache_read_input_tokens":336772,"output_tokens":785,"output_tokens_details":{"thinking_tokens":161   ("num_turns":2; excluded)
first-turn-input-tokens[M4#4]=168388 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":168386,"output_tokens":704,"output_tokens_details":{"thinking_tokens":173   ("num_turns":1)
```

Before 168,841 x3, after 168,388 x3 (single-call runs), reduction 453 tokens (-0.27%), spread 0. Style-specific: a default-style session does not load this file. The `MoAI` style is itself 14,606 tokens heavier than the default style (168,841 vs 154,235 anchor) because its file is 62,593 units; the 453-token cut is small against that and exists only because every other unit is a binding line, a normative instruction, a template, or a frozen/localization unit.

**Baseline-attribution / Gaps / Residual-risk.** As M3. The dropped Core Traits / persistence note / Service Philosophy units are author-classified rationale; the note (unit 65) points at `.claude/rules/moai/workflow/moai-memory.md` § Official Claude Code Auto-Memory Feature, where the content already lives (named as survivor together with the retained persistence paragraph, unit 63).

### M5 — agent description cap, then closing measurement

**D5 decision (leader decision pending — default taken, flagged).** `manager-spec`: the only template/local difference is the dated phrase `per the 2026-05-25 Anthropic catalog consolidation` in the local copy; the template is the cleaner text, so the template was shortened and the local block set equal to it. `manager-git`: the template (553) and local (533) texts differ in substance (template: "invoked for PR creation only when the SPEC is Tier L or the operator selects `--pr` … Route A"; local: "owns every push and delivery decision … explicitly configured WT integration route"). Choosing one text is a delivery-route policy decision, not a size matter, and the agent is small, so `manager-git` was NOT edited and the pre-existing mismatch is reported to the leader (carried gap).

**RED** (`go test ./internal/template/ -run 'TestAgentDescriptionBudget' -count=1 -v`, exit 1, on the unedited agent files, caps 10,460 / 1,815 already in the test):

```
agent_description_budget_test.go:143: agent-description-total=11155 largest=2182 agents=12
agent_description_budget_test.go:149: AGENT_OVER_CAP manager-lead.md description is 2182 UTF-16 units, over the per-agent cap 1815
agent_description_budget_test.go:149: AGENT_OVER_TOTAL the description blocks total 11155 UTF-16 units, over the budget 10460
--- FAIL: TestAgentDescriptionBudget (0.00s)
```

The Go block extractor measures 11,155 / 2,182, equal to the python extractor of plan.md section C (anchor values). Caps are fixed at the achieved values (ratchet): sum 10,460, per-agent 1,815 — both below 11,155 / 2,182.

**GREEN.** Only the `description:` block of four agents changed (template and local mirror carry the identical block; script `tools/agent_desc_apply.py`): manager-lead 2,182 -> 1,815, manager-docs 1,552 -> 1,383, manager-develop 1,094 -> 1,066, manager-spec 1,022 (template) / 1,033 (local) -> 891; total 11,155 -> 10,460 (-695, -6.2%). Role statement, invocation trigger and every `NOT for:` clause kept: `grep -c 'NOT for:'` per template agent = builder-harness 1, e2e-tester 1, manager-design 1, manager-develop 1, manager-docs 1, manager-git 1, manager-lead 1, manager-spec 2, manager-todo 1, plan-auditor 1, super-advisor 3, sync-auditor 1 (equal to EL-7). Removed: dated/provenance phrases (catalog-consolidation history), the lane-label detail and the `Evidence is read before advancing; /clear between phases` sentence of manager-lead's Role B (the lane naming `lane-1` appears twice and `/clear` several times in the agent body, `grep -c "lane-1"` -> 2, and the kanban-dispatch rule the description points at carries both), and parenthetical restatements.

```
make agents-emit                                   exit 0   (regenerated 4 TOMLs: manager-develop/docs/lead/spec.toml)
make build                                         exit 0   (rewrote exactly four `hash:` lines in internal/template/catalog.yaml)
make agents-emit-check                             exit 0
go test ./internal/template/ -run 'TestAgentDescriptionBudget' -count=1 -v      exit 0
agent-description-total=10460 largest=1815 agents=12
--- PASS: TestAgentDescriptionBudget (+ oversized_description_names_file_and_size)
go test ./internal/template/ -run 'TestAgentFrontmatterAudit' -count=1 -v       exit 0   --- PASS: TestAgentFrontmatterAudit
go test ./internal/template/agentemit/... -run 'TestGoldenCommittedArtifactsMatchEmission' -count=1 -v    exit 0   --- PASS
```

Description blocks of the four edited agents, template vs local, compared with the plan.md section C extractor: identical (all four).

**Closing first-turn tokens** (clean tree, HEAD `c1443d00c`, default style `MoAI-Easy`, same command and settings file as M0; M5 itself is below the like-for-like noise floor and is not separately measured — see Gaps):

```
first-turn-input-tokens[final#1]=151057 head=c1443d00c usage={"input_tokens":2,"cache_creation_input_tokens":138656,"cache_read_input_tokens":12399,"output_tokens":332,"output_tokens_details":{"thinking_tokens":51   ("num_turns":1)
first-turn-input-tokens[final#2]=151057 head=c1443d00c usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151055,"output_tokens":305,"output_tokens_details":{"thinking_tokens":71   ("num_turns":1)
first-turn-input-tokens[final#3]=303342 head=c1443d00c usage={"input_tokens":4,"cache_creation_input_tokens":1228,"cache_read_input_tokens":302110,"output_tokens":684,"output_tokens_details":{"thinking_tokens":198   ("num_turns":2; excluded)
first-turn-input-tokens[final#4]=151057 head=c1443d00c usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151055,"output_tokens":517,"output_tokens_details":{"thinking_tokens":147   ("num_turns":1)
```

Single-call runs: min = median = max = 151,057; anchor single-call runs 154,235 (x3). Reduction -3,178 tokens (-2.06%).
result: first-turn-nonregression=true

| milestone | state measured | default-style first-turn tokens | note |
|---|---|---|---|
| anchor | clean, `e5523d672` | 154,235 | min = median = max over 3 single-call runs |
| M2 | 5 modified files uncommitted | 151,407 | -2,828 vs anchor |
| M5 uncommitted (one run, not used) | 15 modified files uncommitted | 151,435 | `first-turn-input-tokens[M5-dirty#1]=151435 usage={"input_tokens":2,"cache_creation_input_tokens":139034,"cache_read_input_tokens":12399,"output_tokens":609,"output_tokens_details":{"thinking_tokens":135` (`"num_turns":1`); +28 vs M2 although the descriptions were shorter — the git-status block listed 15 changed files, which is what made me re-measure on a clean tree |
| M5 final | clean, `c1443d00c` | 151,057 | -3,178 vs anchor; -350 vs M2 (agent-description cut -695 UTF-16 plus git-status-block difference) |

Style-specific (not the default style): `MoAI-Learn` 154,999 -> 154,565 (-434), `MoAI` 168,841 -> 168,388 (-453).

**Gaps / noise (honest limits).** (1) The prompt carries Claude Code's git-status snapshot (branch, changed files, last five commit subjects), so a like-for-like comparison needs equal git state: anchor and final were measured on clean trees (different last-five-commit subjects), M2/M3/M4 after-runs on trees with 5 uncommitted modified files (each listed line adds tokens), so those per-milestone reductions UNDERSTATE the true reduction by roughly the tokens of the listed file lines (tens of tokens); the headline anchor-vs-final comparison is clean-vs-clean. (2) M5's own effect (-695 UTF-16, estimated about -170 tokens) was not measured like-for-like and is not claimed. (3) The number of model calls per `claude -p ok` run varies (1, 2 or 3); only single-call runs are compared, multi-call runs are listed and excluded. (4) Single account/time window. (5) Hooks are off in this measurement, so SessionStart-injected context is not included.

**Residual-risk.** Shortened agent descriptions keep role, trigger and `NOT for:`, but a routing hint that lived only in a removed phrase (for example manager-lead's lane-label wording) is now only in the agent body, which loads on spawn rather than at selection time; `TestAgentFrontmatterAudit` and the golden Codex test cannot judge routing quality.

### Closing verification — acceptance criteria and exported ledger

**Claim.** AC-PFD-001..016 hold on the final tree (HEAD at the time of the last test run = the M5 commit `c1443d00c` plus two uncommitted files at that moment: this file and a one-line unused-variable removal in `output_style_diet_helpers_test.go`, both committed afterwards). Package regression: baseline and final both `ok`.

| AC | result | evidence (this run, this tree) |
|---|---|---|
| 001 | PASS | `grep -rl TestOutputStylesCharBudget internal/template` -> `internal/template/output_style_diet_test.go`; `--- PASS: TestOutputStylesCharBudget`; `output-style=moai 61149`, `output-style=moai-easy 21350`, `output-style=moai-learn 27010`; python UTF-16 per file prints 61149 / 21350 / 27010 (anchor 62593 / 29243 / 28517) |
| 002 | PASS | `--- PASS: TestOutputStyleBindingLedger` with the ten mutation subtests, each asserting the targeted check fired (`MISSING_ROW`, `KIND_MISMATCH`, `DROPPED_KIND` x2, `DROPPED_NOTE`, `DROPPED_SURVIVOR`, `VERBATIM_DIFF`, `DROPPED_TOKENS`, `TOKEN_TOTAL`, `TREATMENT`) |
| 003 | PASS | `python3 -c "…sum(1 for r in …['rows'] if r['treatment']=='rewrite')"` -> `0`, exit 0; `rewrite_row_rejected` subtest `--- PASS` |
| 004 | PASS | AC-004 one-liner: moai `HARD=89 MUST_NOT=4 MUST=27 shall=0`, moai-easy `HARD=33 MUST_NOT=0 MUST=0 shall=0`, moai-learn `HARD=24 MUST_NOT=0 MUST=7 shall=0` (equal to EL-10); mutant probe: removing one `[HARD]` from moai-easy.md text gives `HARD=32` |
| 005 | PASS | `--- PASS: TestOutputStyleHandoffUnitsFrozen` (3 sections; the one-character-mutation subtest names section and both hashes) and `--- PASS: TestOutputStyleLocalizationTableParity` (13 tables; one-cell and removed-row mutation subtests) |
| 006 | PASS | `diff -rq internal/template/templates/.claude/output-styles .claude/output-styles` -> no output, exit 0; `--- PASS` for `TestOutputStylesEncoding`, `TestOutputStylesFallbackDocsContract`, `TestOutputStylesExactlyThree`, `TestOutputStylesFrontmatterSchema`, `TestOutputStylesTemplateLiveParity` (run through the `^TestOutputStyle` prefix, see the command-form note); `make build` exit 0 at M2, M3, M4, M5 |
| 007 | PASS | line counts per file (SPEC/card, date, hex hash, language names) via python equivalents of the four `grep -cE` expressions (the guard refuses a quoted `|`): moai 15/0/0/3, moai-easy 0/0/0/0, moai-learn 0/1/0/1 = EL-9 |
| 008 | PASS | `settings.json.tmpl:414: "skillListingBudgetFraction": 0.02,` and `.claude/settings.json:410: …0.02,`; `git diff --numstat 5d5ff1aae HEAD -- <both>` -> no output |
| 009 | PASS | `--- PASS: TestAgentDescriptionBudget`, total 10,460 < 11,155, largest 1,815 < 2,182; the oversize mutation names `manager-develop.md` and the size |
| 010 | PASS | `NOT for:` counts equal EL-7 (listed in M5); `--- PASS: TestAgentFrontmatterAudit`; template-vs-local description blocks of the four edited agents identical |
| 011 | PASS | the first-turn token lines of M0, M1, M2, M3, M4, M5 (`grep -c` on the line prefix -> 29 lines, 29 raw `usage` JSON carriers) each carry the raw `usage` JSON; anchor and final each have three single-call runs; `result: first-turn-nonregression=true` (151,057 <= 154,235) |
| 012 | PASS | `surface_guard.py 5d5ff1aae` -> exit 0, 32 `ok` lines, `surface-guard=PASS`; positive controls below |
| 013 | PASS | `"outputStyle": "MoAI-Easy"` at `settings.json.tmpl:418` and `.claude/settings.json:417`; `git diff --numstat` on both settings files -> no output; `git diff --name-only 5d5ff1aae HEAD -- <skills, rules, CLAUDE.md, AGENTS*, internal/hook>` -> no output; `git merge-base --is-ancestor 5d5ff1aae HEAD` -> exit 0 |
| 014 | PASS | baseline `ok … internal/template 135.858s`, `ok … internal/config 10.811s` (exit 0) -> final `ok … internal/template 146.983s`, `ok … internal/config 6.057s` (exit 0): no new `FAIL` name |
| 015 | PASS | `bin/moai spec lint .moai/specs/SPEC-PREFIX-DIET-001/spec.md` (binary built from this tree) -> `✓ No findings — all SPEC documents are valid`, exit 0; `grep -c '^## §E\.' progress.md` -> 4 |
| 016 | PASS | `AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run 'TestGoldenCommittedArtifactsMatchEmission' -count=1 -v` -> `--- PASS`; `make agents-emit-check` exit 0; the four TOMLs changed in the M5 commit; the RED side of this property is the plan-phase observation EL-14 (mismatch before `make agents-emit`) — re-observed here only as the passing end state |

`golangci-lint run ./internal/template/` (v2.1.6): first run flagged one `unused` variable in the new helper (`dietTokenNames`), removed; second run exit 0.

**AC-012 positive controls** (each applied on top of the final tree, observed, then restored with `git checkout --`/`rm`; `git status --short` showed only this SPEC's pending edits afterwards; outputs read from `surface_guard.py 5d5ff1aae`):

| mutation | observed output | exit |
|---|---|---|
| forbidden surfaces `output-styles agents codex-tomls` (command-line args) | 18 x `VIOLATION forbidden-surface …`, `surface-guard=FAIL` | 1 |
| `echo mutant >>` to template `manager-todo.md` (body) | `VIOLATION agent-body-or-nondescription-frontmatter-changed agents internal/template/templates/.claude/agents/moai/manager-todo.md` | 1 |
| `name: MoAI-Easy` -> `MoAI-Easyx` in the moai-easy.md template frontmatter | `VIOLATION frontmatter-changed output-styles internal/template/templates/.claude/output-styles/moai/moai-easy.md` | 1 |
| `.claude/settings.json` `skillListingBudgetFraction` 0.02 -> 0.01 | `VIOLATION outside-allowlist .claude/settings.json` | 1 |
| new file `.claude/rules/x.md` | `VIOLATION outside-allowlist .claude/rules/x.md` | 1 |
| catalog.yaml: hash line of the UNEDITED `manager-todo` entry changed | `VIOLATION catalog-change-beyond-edited-agent-hashes (hash-of-unedited-entry templates/.claude/agents/moai/manager-todo.md) catalog-hashes internal/template/catalog.yaml` | 1 |
| catalog.yaml: `tier: core` -> `tier: corex` on that entry | `VIOLATION catalog-change-beyond-edited-agent-hashes (non-hash-line - tier: core; non-hash-line + tier: corex) …` | 1 |

**Exported `dropped` ledger rows** (77 rows, 10,844 UTF-16 units in 16 contiguous groups; every row is in `internal/template/testdata/output_style_ledger.json` with `before_text`, `note` and `survivor`; groups listed so the sync-audit and the leader can read them without opening 548 rows):

| file | units | rows | kind | UTF-16 | content (first unit lines) | survivor | reason |
|---|---|---|---|---|---|---|---|
| moai.md | 6-7 | 2 | rationale | 213 | ### Core Traits; **Persistence** (never walk away mid-task) · **Transparenc | moai.md section 1 'Operating Principles' | core-trait list restates principles 3 and 5 and the language rule of section 9 |
| moai.md | 64-65 | 2 | rationale | 722 | This is the 2026 Anthropic-recommended persistence pattern; > Note: the memory directory is a **native Claude Code aut | moai.md section 6 persistence paragraph (memory directory path) and .claude/rules/moai/workflow/moai-memory.md | provenance claim and an explanatory note pointing at the memory rule; the persistence behaviour itself survives |
| moai.md | 246-250 | 5 | rationale | 509 | ## 12. Service Philosophy; I'm a **pair programming orchestrator**, not a task-runner; Every time we work together, I aim for: ... | moai.md section 1 'Operating Principles' | service philosophy restates the operating principles of section 1 |
| moai-easy.md | 44-45 | 2 | rationale | 230 | ### For beginners; You don't have to memorize any of this — honestly, don't e | moai-easy.md section 5 'In plain words' / 'When I delegate vs. do it myself' | reassurance for beginners; the delegation behaviour is stated in the surviving section 5 units |
| moai-easy.md | 53-71 | 19 | example | 1299 | ### Examples; / Term / Plain-language explanation /; /------/----------------------------/ ... | moai-easy.md section 6 'How it works' (the first-mention pattern and its blockquote example) | sixteen sample glossary entries; the plain-language rule itself and a worked first-mention example survive |
| moai-easy.md | 72-73 | 2 | rationale | 249 | ### When to slow down; If you ever go "wait, what does X mean?" — I'll stop right | moai-easy.md section 11 table row 'Lost on a term' and section 6 'How it works' | restates the pause-and-explain behaviour that the section 11 table row already specifies |
| moai-easy.md | 134-142 | 9 | example | 2578 | ## 10. Banner Examples (What Each Looks Like in Real Use); Let me give you a quick tour of each banner in action, so ; ### Banner 1 — Let's Begin ... | moai-easy.md section 7 banner skeletons (Banner 1-6 templates) | worked instances of the six banners; the skeleton of each banner survives in section 7 |
| moai-easy.md | 158-169 | 12 | rationale | 1780 | ## 12. Questions Beginners Often Have (FAQ); **Q: Do I need to know how to code to use MoAI-Easy?**; **Q: Will you explain what the code does?** ... | moai-easy.md section 1 (switching styles, who I am), section 4 (check step), section 5 (delegation), section 11 (I'm lost) | FAQ restating answers already given in sections 1, 4, 5 and 11; switch mechanism survives in section 1 |
| moai-easy.md | 170-174 | 5 | rationale | 782 | ## 13. My Teaching Philosophy; > *"You don't have to know everything. You just have to kn; Here's what I believe: ... | moai-easy.md section 2 'My Promise to You (Operating Principles)' | teaching philosophy restates the operating principles of section 2 |
| moai-easy.md | 175-182 | 8 | rationale | 553 | ## 14. Quick Reference — When to Switch Styles; / If you want... / Switch to /; /----------------/-----------/ ... | moai-easy.md section 1 sibling table and switch instruction | style-switch quick reference duplicates the section 1 sibling table and the /output-style switch mechanism |
| moai-easy.md | 183-184 | 2 | rationale | 422 | ## 15. Friendly Reminders; I'm your companion here. We go at your pace — always. | moai-easy.md section 11 situation table | friendly reminders restate the section 11 situation/response table |
| moai-learn.md | 5-7 | 3 | rationale | 334 | ### The MoAI-Learn Principle; > *"Make everything as simple as possible, but no simpler.; Let me be upfront: I won't hide behind jargon on the first | moai-learn.md section 1 'Core Mission' and section 3 'Phase 2 - Teach' | principle quote and upfront-honesty paragraph restate the jargon-free first-pass rule of Phase 2 |
| moai-learn.md | 122-122 | 1 | rationale | 411 | Root cause of the defect: a prior version's §9 said "trans | moai-learn.md section 8 anti-pattern catalogue intro (units before the table) | root-cause narrative for the catalogue; the HARD violation statement and the catalogue itself survive |
| moai-learn.md | 137-141 | 5 | rationale | 762 | ## 11. Teaching Philosophy; > *"The important thing is never to stop questioning. Curi; What I hold to: ... | moai-learn.md section 1 'Core Mission' and section 3 phase units | teaching philosophy restates the mission bullets and the five-phase protocol |

**Review pointers for the sync-audit / leader (known weak survivors).** (a) moai-easy FAQ row "What if I make a mistake?" (code is reversible through version control) has only a partial survivor: section 2 principle 4 and section 3 "No big surprises" survive, the explicit version-control reassurance does not. (b) moai-easy section 14 dropped the explicit `/output-style MoAI-Learn` switch text; the generic switch mechanism for MoAI survives in section 1 and the runtime lists all styles under `/output-style`. (c) moai.md unit 65 (the memory-store note) cites a rule file as survivor, not a unit of the same file; the retained unit 63 still names the memory path.

**Gaps (whole run).** The leader's confirmation of the D4 targets and of D5 (manager-git) is still open. The installed `moai` binary was not compared with HEAD; `bin/moai` was built from the tree and used for `spec lint`. `go vet` was run, `golangci-lint` on `./internal/template/` only. The independence of the two extractors is bounded by one author. Codex and other harnesses were not measured (SPEC section H gap unchanged). The AC-014 package run started before the one-line unused-variable removal; the four new tests and the budget test were re-run afterwards (exit 0).

**Residual-risk.** Author classification of `rationale`/`example` (listed above); manager-git template/local description mismatch persists; per-milestone token numbers carry the git-status-block confounder described in M5; moai.md's reduction is 2.3% because 89 `[HARD]` lines and the normative templates cannot be dropped or rewritten under D3 — a larger moai.md reduction needs a leader decision that allows rewriting or relocating binding text.

### Sync-audit repair (F1, F2 and the strict survivor sweep) — supersedes the sizes and the dropped set above

**Claim.** The sync-audit findings F1 (`moai-easy-0183`, "just do it, I trust you" -> minimal check-ins) and F2 (`moai-learn-0122`, "locales beyond these four … don't transliterate") are repaired; a strict sweep of all 77 dropped rows restored 25 units verbatim; the 52 rows that stay dropped carry a survivor pointer that the ledger test resolves. The sizes, dropped counts and token figures in M2-M5 and in the closing verification above describe the state BEFORE this repair; the figures below are current. Repair started from HEAD `f1078478d` (verified before any edit; the sync-phase commits `1d711d13b`, `b7ea0d823`, `f1078478d` were already on the branch).

**Restored verbatim** (ledger `dropped` -> `verbatim`, template first, `make build` exit 0, local mirrors copied, `diff -rq` empty):

| id | first words | why no survivor |
|---|---|---|
| moai-0006, 0007 | `### Core Traits` / `**Persistence** (never walk away mid-task) · **Transparency**…` | Transparency and Efficiency appear nowhere else |
| moai-easy-0044, 0045 | `### For beginners` / `You don't have to memorize any of this…` | the "I handle the routing behind the scenes" promise has no other carrier |
| moai-easy-0158, 0159, 0167, 0169 | `---` / `## 12. Questions Beginners Often Have (FAQ)` / `**Q: What if I make a mistake?**` / `---` | the reversibility promise ("I'll show you how to undo anything we do") is only here (F3); heading and separators restored with it |
| moai-easy-0170..0174 | `## 13. My Teaching Philosophy` … `**How I measure success**` … `---` | "explain it to a friend" success test and several principles have no literal survivor |
| moai-easy-0182, 0183, 0184 | `---` / `## 15. Friendly Reminders` / `I'm your companion here…` | F1: "just do it, I trust you -> proceed with minimal check-ins" is the sole carrier |
| moai-learn-0005..0007 | `### The MoAI-Learn Principle` / `> "Make everything as simple as possible…"` / `Let me be upfront: I won't hide behind jargon…` | the middle-schooler test of the first explanation has no literal survivor |
| moai-learn-0122 | `Root cause of the defect: a prior version's §9 said…` | F2: "for locales beyond these four … don't transliterate" is the sole carrier |
| moai-learn-0137..0141 | `## 11. Teaching Philosophy` … `What I hold to:` … `**How I know it worked**` … `---` | "depth over breadth" and the self-test of understanding have no survivor (also F3 "note you keep") |

**Rows that stay dropped (52) and their pointers** (`survivor_file` + `survivor_anchor`, resolved by `TestOutputStyleBindingLedger`; all resolve): moai.md 64 -> `keep working right through auto-compaction` (moai.md); 65 -> `moai memory doctor` in `templates/.claude/rules/moai/core/moai-constitution.md` (always loaded); 246-247 -> `### Operating Principles`; 248 -> `pair programming partner`; 249 -> `Intent-First`; 250 -> `Verify Every Step`. moai-easy.md 72-73 -> `Pause, explain X with an analogy, then continue`; 53-71 -> `A **function** (a reusable recipe`; 134-142 -> `### Banner 1 — Let's Begin (Step 1: Understand)`; FAQ 160 `is plenty`, 161 `one-line note on what just changed`, 162 `never race past the confusion`, 163 `change this`, 164 `/output-style MoAI`, 165 `| **MoAI-Learn** | Learning a concept deeply`, 166 `"I don't understand" is always a perfectly good thing to say`, 168 `Bring back the result`; 175-180 -> `### How I'm different from my siblings`; 181 -> ``just type `/output-style MoAI` right here in the chat``. moai-learn.md: no dropped row remains. The pointer proves the anchor exists, not that the anchor carries the meaning; that judgement stays with the reviewer (named in Residual-risk).

**Guard strengthening (RED first).** `dietRow` gained `survivor_file` / `survivor_anchor`; the ledger test emits `SURVIVOR_UNRESOLVED` unless the file is readable and contains the anchor. RED on the previous fixture (`go test ./internal/template/ -run 'TestOutputStyleBindingLedger' -count=1 -v`, exit 1): 77 x `SURVIVOR_UNRESOLVED`, `--- FAIL: TestOutputStyleBindingLedger/dropped_survivor_anchor_missing`, `--- FAIL: …/dropped_survivor_file_missing`. GREEN after the repair: both mutation subtests pass (anchor `no-such-anchor-in-that-file-7f3a` and file `templates/no/such/file.md` are rejected).

**Current whole-file UTF-16 and budgets** (python one-liner of AC-PFD-001 equals the test log): moai.md 62,593 -> 61,362 (-1,231); moai-easy.md 29,243 -> 23,036 (-6,207); moai-learn.md 28,517 -> 28,517 (0). Budget constants set to these sizes. Dropped rows: 52 (was 77); dropped UTF-16 total = 1,231 + 6,207 = 7,438 (was 10,844). Zero `rewrite` rows; binding-token totals still equal the anchor (ledger `TOKEN_TOTAL` check passes).

**Verification (scoped; exit codes).** `go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v` exit 0 (budget, ledger with 12 mutation subtests, frozen, localization, five existing); `go test ./internal/template/ -run 'TestAgentDescriptionBudget' -count=1` exit 0; `make agents-emit-check` exit 0; `make build` exit 0; `golangci-lint run ./internal/template/` exit 0; `diff -rq internal/template/templates/.claude/output-styles .claude/output-styles` no output. `surface_guard.py 5d5ff1aae`: every file this repair touched is `ok`; the run still prints `VIOLATION outside-allowlist CHANGELOG.md` and `surface-guard=FAIL` because the sync-phase commit `b7ea0d823` (not part of this repair) changed `CHANGELOG.md`, which the run-phase allowlist deliberately does not cover (guard left unchanged as instructed; the guard is a run-phase, pre-sync instrument).

**First-turn tokens, clean tree, HEAD `e5be9868f`** (default style MoAI-Easy, same command/settings file, every run `"num_turns":1`):

```
first-turn-input-tokens[repair#1]=151673 head=e5be9868f usage={"input_tokens":2,"cache_creation_input_tokens":139272,"cache_read_input_tokens":12399,"output_tokens":362,"output_tokens_details":{"thinking_tokens":91
first-turn-input-tokens[repair#2]=151673 head=e5be9868f usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151671,"output_tokens":412,"output_tokens_details":{"thinking_tokens":76
first-turn-input-tokens[repair#3]=151673 head=e5be9868f usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151671,"output_tokens":292,"output_tokens_details":{"thinking_tokens":64
```

151,673 x3 vs anchor 154,235: -2,562 tokens (-1.66%); previously -3,178. The repair gives back 616 tokens (the restored text) to keep the promises the dropped units carried.

**Residual-risk (updated).** A pointer proves the anchor exists in the named file, not that it carries the dropped instruction; the 52 remaining rows rest on that reviewer judgement. The moai.md note (row 65) points at a rule file, not a style unit.

### Repair 2 (sync-audit iter2 F-D3, F-D4, F-D5; leader-ordered conservative restore)

Started from HEAD `c7ba13ab4` (verified). **Restored verbatim** in moai-easy.md (template first, `make build` exit 0, local mirror copied, `diff -rq` empty): `moai-easy-0073` (`If you ever go "wait, what does X mean?" — I'll stop right there…`, the pause-until-comfortable reassurance; its heading unit 0072 stays dropped), `moai-easy-0163` (`**Q: Can I change my mind mid-task?**`), `moai-easy-0181` (`Switch any time via /config → Output style → choose…`, the explicit `/output-style MoAI-Learn` route). **F-D5:** the free-text `survivor` field of the remaining 49 dropped rows is cleared; the ledger test's `DROPPED_SURVIVOR` check now keys off the `survivor_file` + `survivor_anchor` pointer (its mutation subtest clears the pointer) and `SURVIVOR_UNRESOLVED` is unchanged, so the rule is not weakened.

Whole-file UTF-16: moai.md 61,362 (unchanged); moai-easy.md 23,036 -> 23,586 (+550); moai-learn.md 28,517 (unchanged); budget constant for moai-easy set to 23,586. Dropped rows 52 -> 49, total 6,888 UTF-16 units (was 7,438).

Verification exit codes: `go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v` 0 (all nine tests PASS); `go test ./internal/template/ -run 'TestAgentDescriptionBudget' -count=1` 0; `make agents-emit-check` 0; `make build` 0; `golangci-lint run ./internal/template/` 0.

First-turn tokens, clean tree, HEAD `76e480974`, default style, every run `"num_turns":1`:

```
first-turn-input-tokens[repair2#1]=151873 head=76e480974 usage={"input_tokens":2,"cache_creation_input_tokens":139472,"cache_read_input_tokens":12399,"output_tokens":361,"output_tokens_details":{"thinking_tokens":109
first-turn-input-tokens[repair2#2]=151873 head=76e480974 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151871,"output_tokens":319,"output_tokens_details":{"thinking_tokens":88
first-turn-input-tokens[repair2#3]=151873 head=76e480974 usage={"input_tokens":2,"cache_creation_input_tokens":0,"cache_read_input_tokens":151871,"output_tokens":302,"output_tokens_details":{"thinking_tokens":59
```

151,873 x3 vs anchor 154,235: -2,362 (-1.53%); the previous repair measured -2,562. The restored text costs 200 tokens.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_artifacts: output-style ledger fixtures (`internal/template/testdata/output_style_{ledger,frozen,localization}.json`), guard tests (`internal/template/output_style_diet_test.go`, `output_style_diet_helpers_test.go`, `agent_description_budget_test.go`), guard script `surface_guard.py`, tools `tools/diet_ledger.py` and `tools/agent_desc_apply.py`
run_milestones: M0 (baseline ledger front, own commit before any edit) · M1 (key guard, no change) · M2 moai-easy -7,893 · M3 moai-learn -1,507 · M4 moai -1,444 · M5 agent descriptions -695 and caps
run_headline: default-style first-turn input tokens 154,235 -> 151,057 (-3,178, -2.06%); per-style: MoAI-Learn -434, MoAI -453
run_open_for_leader: D4 targets (all three below the plan drafts, REQ-PFD-002), D5 manager-git, SPEC-text gaps (REQ-PFD-013/AC-PFD-012 catalog surface; AC-PFD-015 unmapped; spec.md line 139 count)

## §E.4 Sync-phase Audit-Ready Signal — SUPERSEDED (kept as history; the re-close signal below is current)

superseded_by: the re-close §E.4 block below, after the sync-audit FAIL 81.7 repair and the 0.4.0 in-place amendment; the figures in this first block (-3,178 / -2.06% and the pre-repair sizes) were replaced by §E.2 "Sync-audit repair"
sync_complete_at: 2026-10-03
prior_sync_commit_sha: b7ea0d823e9d
sync_status: complete (sync-audit and the leader's evidence read are still owed; this record claims neither)
b12_self_test_a: PASS — `grep -c "SPEC-PREFIX-DIET-001" CHANGELOG.md` printed `0` before the entry was written (no duplicate)
b12_self_test_b: PASS with a named deviation — ac_source=`.moai/specs/SPEC-PREFIX-DIET-001/acceptance.md`, tier=M; live count 16 (AC-PFD-001..016), no `[RETIRED]`/`[REF]` token and no `moai-ac-prefix` line in the file (`grep -c -F` -> 0 for each), so no identifier is ambiguous; the CHANGELOG entry states 16. The deviation: the worktree guard refused `awk -f`, so the contract's awk counter was not executed; the count was rebuilt from `grep -o -E "AC-[A-Z0-9]+-[0-9]+[a-z]?"` (28 occurrences, 16 distinct ids, 001..016) by hand
b12_self_test_c: PASS — `ls` of all 17 paths named in the CHANGELOG entry (3 output-style templates, 4 agent templates, 3 test files, 3 testdata fixtures, `surface_guard.py`, `tools/diet_ledger.py`, `tools/agent_desc_apply.py`, `catalog.yaml`) returned each path, no error
changelog_entry_position: `CHANGELOG.md` `[Unreleased]` -> first `### Changed` block (opens at line 178), first entry; the closest recent sync-close entries sit under `### Added` and `### Changed` of the same section, and a change of this kind (shortened prompt bodies, new guards) belongs under Changed
frontmatter_status_transitions: `spec.md` `in-progress -> implemented -> completed` on this single sync commit (version stays 0.3.1; `updated:` already reads 2026-10-03). `plan.md`, `acceptance.md` and `progress.md` carry no `status:` field (`spec-frontmatter-schema.md` § Artifact Statelessness), so no status line changed in them; their `updated:` already reads 2026-10-03. No body text of `spec.md` / `plan.md` / `acceptance.md` was edited
canary_compliance_check: not applicable (this SPEC defines no forward-looking policy that its own sync tests)

### Sync-phase scope decisions

- CHANGELOG: edited (the repo lists SPEC sync-close entries for changes of this kind; English only).
- README (4 locales) and docs-site (4 locales): NOT edited. No behaviour, command, flag, setting or install step changed; `grep -rli "Questions Beginners Often Have" . --include="*.md"` (excluding `.git`, `.moai`, `node_modules`) -> no output, and `grep -rli "Teaching Philosophy" docs-site` -> no output, so no user-facing page cites a dropped output-style section. Only those two section names were searched; the other dropped section titles were not.
- Codemaps / project docs (`product.md`, `structure.md`, `tech.md`): NOT regenerated. No directory, package or dependency edge was added (new files are tests and fixtures inside the existing `internal/template/` package), so the structural-change trigger of `sync/doc-execution.md` does not fire.

### Evidence-bearing report

**Claim.** (1) The run-phase guards still pass on the final tree. (2) The SPEC lints clean and the audit engine finds no lifecycle drift after the `completed` transition. (3) The CHANGELOG entry is unique, counts the acceptance criteria correctly and cites only existing paths. (4) One guard result is red after the sync edit and is explained below, not hidden.

**Evidence** (commands run in this tree, cwd = worktree root, HEAD `1d711d13b` plus the uncommitted sync edits named per row; a tool result with no error is recorded as exit 0, the Bash tool prints a code only on failure):

```
go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v        exit 0 (HEAD 1d711d13b, before the sync edits)
  --- PASS: TestOutputStylesCharBudget   (moai 61149 / moai-easy 21350 / moai-learn 27010)
  --- PASS: TestOutputStyleBindingLedger (10 mutation subtests)  --- PASS: TestOutputStyleHandoffUnitsFrozen
  --- PASS: TestOutputStyleLocalizationTableParity               --- PASS: the five TestOutputStyles* parity tests
  ok  github.com/modu-ai/moai-adk/internal/template  0.417s
go test ./internal/template/ -run 'TestAgentDescriptionBudget' -count=1 -v      exit 0 (same tree)
  agent-description-total=10460 largest=1815 agents=12
  --- PASS: TestAgentDescriptionBudget (+ oversized_description_names_file_and_size)
  ok  github.com/modu-ai/moai-adk/internal/template  0.199s
make agents-emit-check                                                           exit 0 (same tree)
  ok  github.com/modu-ai/moai-adk/internal/template/agentemit  0.178s
python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py 5d5ff1aae              exit 0 (same tree, before the sync edits)
  32 `ok` lines ... surface-guard=PASS
moai spec lint .moai/specs/SPEC-PREFIX-DIET-001                                  exit 0, run twice: before the sync edits and again after the status transition + CHANGELOG edit
  ✓ No findings — all SPEC documents are valid
mcp__moai__spec_audit (filter_spec=SPEC-PREFIX-DIET-001, after the status transition)
  total_specs 1, modern_era_clean 1, grandfathered 0, one INFO EraAutoDetected (H-5), no drift finding
python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py 5d5ff1aae              exit 1 (re-run after the CHANGELOG edit)
  VIOLATION outside-allowlist CHANGELOG.md   ... surface-guard=FAIL   (CHANGELOG.md is the only violation line)
```

The last row is the sync-phase edit itself: `CHANGELOG.md` is a mandated sync deliverable but is not in the guard's run-phase allowlist, and `spec.md` §F / REQ-PFD-013 are silent on it. `grep -rln "surface_guard" . --include="*.go" --include="Makefile" --include="*.yml" --include="*.yaml" --include="*.sh"` (excluding `.git`, `node_modules`) -> no output, so nothing in CI or the test suite runs the guard; it is a SPEC-local script run by hand. Nothing in the guard was edited here (it is a SPEC artifact).

**Baseline-attribution.** All rows: this run, worktree `t1450`, branch `WT-prefix-diet-stage2`, HEAD `1d711d13b`. Tool provenance (VCI 2.2): `moai` is the installed `v3.2.0-rc.27` (`moai version`, commit `0732cc699`, built 2026-10-03T03:34:50Z); `git merge-base --is-ancestor 0732cc699 HEAD` exits 0 and `git rev-list --count 0732cc699..HEAD` prints 157, so the installed build IS a strict ancestor of HEAD (157 commits behind); `git diff --stat 0732cc699 HEAD -- internal/spec` prints nothing, so the SPEC lint and audit code is identical between that build and this tree — the stale build does not change the lint/audit verdicts, but the build was not rebuilt from this tree. The Go tests and `make agents-emit-check` ran from this tree's source. The first-turn token figures (-3,178, -2.06%) are not re-measured in this sync; they are the run-phase measurement in §E.2 (anchor `e5523d672` clean vs `c1443d00c` clean, single-call runs).

**Gaps.** (1) The AC counter's awk script was not run (guard refusal); the count of 16 is a manual reduction of `grep -o` output. (2) `moai spec lint` and the audit ran after the status transition but before this §E.4 text was written; a final lint pass after the §E.4 write is reported in the hand-off message, not recorded here (a record cannot cite a run that follows it). (3) No sync-audit (4-dimension or cold sync-auditor) was run by this agent — that verdict belongs to the orchestrator / sync-auditor. (4) Only two dropped section titles were searched in docs-site and the repo's markdown, not every dropped group. (5) Codex and other harnesses were not measured (unchanged from §E.2). (6) The sync-phase quality-gate Stop hook was not observed by this agent.

**Residual-risk.** Carried debts, listed without softening:

- **D5 — `manager-git` description mismatch.** The template (553) and local (533) `description:` blocks differ in substance (PR-only-on-Tier-L/`--pr` versus owns-every-push); it is a delivery-route policy question, so it was left unedited and remains open for the leader.
- **Weak-survivor dropped groups.** (a) moai-easy FAQ row "What if I make a mistake?" lost its explicit version-control reassurance (partial survivor only); (b) moai-easy section 14 dropped the explicit `/output-style MoAI-Learn` switch text (the generic switch mechanism survives in section 1); (c) moai.md unit 65 names a rule file, not a unit of the same file, as survivor. Survivor classification is author-side; no tool can judge whether a dropped unit was really rationale.
- **`plan.md` still shows draft targets.** Its section B carries the draft budget numbers (21,000 / 20,000 / 45,000 UTF-16 units) that §E.2 D4 superseded with the measured droppable totals (21,350 / 27,010 / 61,149); `spec.md` 0.3.1 states the measured values, but a sync-phase agent may not edit `plan.md` body, so the stale numbers stay until `manager-spec` amends them.
- **Per-milestone token numbers were taken on dirty trees.** The M2, M3 and M4 after-runs had five modified files listed in the prompt's git-status block and M5's own effect was not measured like-for-like, so those per-milestone reductions slightly understate the cut. Only the headline (anchor clean vs final clean: 154,235 -> 151,057, -3,178, -2.06%) is a like-for-like comparison. One account/time window, hooks off.
- **Guard vs CHANGELOG.** `surface_guard.py` reports `VIOLATION outside-allowlist CHANGELOG.md` once the sync entry exists; run it against the run-phase HEAD to get PASS. A `manager-spec` amendment (allow `CHANGELOG.md` in the sync phase, or state the guard is run-phase only) would remove the ambiguity.
- **Routing quality of shortened descriptions.** `TestAgentFrontmatterAudit` and the golden Codex test cannot judge whether a routing hint that lived only in a removed phrase (for example `manager-lead`'s lane-label wording) is still found at selection time; it is now only in the agent body, which loads on spawn.
- **Not pushed.** The branch is local; integration and the remote CI verdict are the leader's.

## §E.4 Sync-phase Audit-Ready Signal — re-close (current; supersedes the block above)

sync_complete_at: 2026-10-03
sync_commit_sha: 6b2fe3a1f454
sync_status: complete (the first sync-audit FAILED at 81.7, receipt rcpt-8cf2e184c500ae809e6bd8cd; a new sync-audit of the repaired tree and the leader's evidence read are still owed, and this record claims neither)
b12_self_test_a: PASS — `grep -c "SPEC-PREFIX-DIET-001" CHANGELOG.md` prints `1` (counts lines): this run corrects the single existing t1450 entry in place, so 1 is the expected count and no second entry was added
b12_self_test_b: PASS with the same named deviation as the first block — ac_source=`.moai/specs/SPEC-PREFIX-DIET-001/acceptance.md`, tier=M; the awk counter was not run (worktree guard refuses `awk -f`); `grep -o -E "AC-[A-Z0-9]+-[0-9]+[a-z]?"` printed 28 occurrences of 16 distinct ids (AC-PFD-001..016), `grep -c -F "[REF]"` and `grep -c -F "[RETIRED]"` both print `0`, so none is ambiguous; the CHANGELOG entry still states 16
b12_self_test_c: PASS — `ls` of the 17 paths the CHANGELOG entry names returned every path, no error (the re-run was needed because the repair rewrote three style templates, two test files, the ledger fixture and `catalog.yaml`)
changelog_entry_position: unchanged — `CHANGELOG.md` `[Unreleased]` -> first `### Changed` block, first entry; corrected in place (commit range, token result, sizes, restored units, mutation-case count, survivor-pointer rule, closing sentence)
frontmatter_status_transitions: `spec.md` `in-progress -> implemented -> completed` on this ONE re-close commit, the amendment re-close of `completed -> in-progress (amendment)` (version 0.4.0, `amendment_of: SPEC-PREFIX-DIET-001`, `## Amendments` records `prior_completed_sha: b7ea0d823`). `plan.md`, `acceptance.md` and `progress.md` carry no `status:` (§ Artifact Statelessness); their `updated:` already reads 2026-10-03. No body of `spec.md` / `plan.md` / `acceptance.md` was edited
canary_compliance_check: not applicable

### Evidence-bearing report (re-close)

**Claim.** (1) After the repair the guard tests, the agent-description guard, the Codex golden and the package lint are green on the final tree. (2) The SPEC lints clean. (3) The CHANGELOG entry now states the repaired numbers and no stale pre-repair figure except the deliberate "superseded" sentence. (4) One guard line is red and expected.

**Evidence** (this run, cwd = worktree root, HEAD `812cc20e7` plus only this re-close's uncommitted edits to `CHANGELOG.md` and `progress.md`; a tool result without an error is recorded as exit 0, the Bash tool prints a code only on failure):

```
go test ./internal/template/ -run '^TestOutputStyle' -count=1 -v             exit 0
  --- PASS: TestOutputStylesCharBudget   (moai 61362 / moai-easy 23036 / moai-learn 28517)
  --- PASS: TestOutputStyleBindingLedger (12 mutation subtests, incl. dropped_survivor_anchor_missing and dropped_survivor_file_missing)
  --- PASS: TestOutputStyleHandoffUnitsFrozen  --- PASS: TestOutputStyleLocalizationTableParity
  --- PASS: the five TestOutputStyles* parity tests        ok  github.com/modu-ai/moai-adk/internal/template  0.249s
go test ./internal/template/ -run 'TestAgentDescriptionBudget' -count=1 -v   exit 0
  agent-description-total=10460 largest=1815 agents=12    --- PASS: TestAgentDescriptionBudget (+ oversized subtest)
make agents-emit-check                                                        exit 0   ok  .../template/agentemit  0.153s
moai spec lint .moai/specs/SPEC-PREFIX-DIET-001                               exit 0   "No findings — all SPEC documents are valid"
golangci-lint run ./internal/template/                                        exit 0   "0 issues."
python3 .moai/specs/SPEC-PREFIX-DIET-001/surface_guard.py 5d5ff1aae           exit 1   every file `ok` except `VIOLATION outside-allowlist CHANGELOG.md` -> surface-guard=FAIL
grep -c -E "151,057|3,178|61,149|21,350|27,010|2\.06%|2\.3%" CHANGELOG.md     prints 2: line 180 is the deliberate "An earlier -3,178 / -2.06% ... is superseded" sentence; the other hit is an unrelated entry (`runCodexReviewGate` 92.3%)
```

The guard's red line is the expected one: REQ-PFD-013 (0.4.0) states `CHANGELOG.md` is a sync-phase deliverable outside the run-phase allowlist, and the guard was not edited. The guard also lists `tools/diet_restore.py` as `ok`.

**Baseline-attribution.** All rows: this run, worktree `t1450`, branch `WT-prefix-diet-stage2`, HEAD `812cc20e7`. Tool provenance: Go tests, `make agents-emit-check` and `golangci-lint` ran from this tree's source. `moai spec lint` ran with the installed `moai` v3.2.0-rc.27 (commit `0732cc699`, from `moai version` in the first sync pass); that build is a strict ancestor of HEAD and `git diff --stat 0732cc699 HEAD -- internal/spec` printed nothing in the first pass, so the SPEC lint code was identical then; it was not re-compared against `812cc20e7` in this pass. Token and size figures are NOT re-measured here: they are the repair-phase measurements in §E.2 "Sync-audit repair" (HEAD `e5be9868f`, clean tree, three identical single-call runs 154,235 -> 151,673, -2,562, -1.66%; whole-file UTF-16 61,362 / 23,036 / 28,517; 52 dropped rows, 7,438 units, zero rewrite rows).

**Gaps.** (1) The AC counter's awk script was not run (guard refusal); 16 comes from `grep -o` output reduced by hand. (2) No new sync-audit was run by this agent. (3) The `MoAI` and `MoAI-Learn` first-turn effects were not re-measured after the repair, so the CHANGELOG states none. (4) `spec audit` (MCP) and a final `moai spec lint` after this §E.4 was written are reported in the hand-off message, not recorded here. (5) The installed `moai` build was not re-compared with `812cc20e7`. (6) Codex and other harnesses were not measured.

**Residual-risk.** A survivor pointer proves its anchor exists in the named file, not that the anchor carries the dropped instruction; the 52 rows that stay dropped rest on reviewer judgement (§E.2 "Sync-audit repair"). Still open from the first close: D5 (`manager-git` description differs between template 553 and local 533, unedited), the `plan.md` draft targets (manager-spec's to amend), the routing quality of the four shortened agent descriptions (no test can judge it), and the per-milestone token figures of M2-M4, which predate the repair and were taken on dirty trees (only the repaired clean-tree headline is like-for-like). Not pushed; integration and the remote CI verdict are the leader's.
