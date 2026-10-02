auditor-model: claude-sonnet-5-5[1m]

verdict: FAIL
audited_sha: e7dbdeef35e6d3d53b1fb886ccb84f77640433e1

# SPEC Review Report: SPEC-CC-ULTRACODE-TOGGLE-001
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.75 (no must-pass failure; aggregate is below the Tier M PASS threshold 0.80, and blocking defects D1-D7 stand)
Plan Artifact Hash: sha256 of `cat acceptance.md plan.md spec.md` = `767e94f2519e0f278c78f2b2683b1b02663b50c8479225246bce693cbe3205cf` (per-file: acceptance 7b149159..., plan f8d7ff45..., spec 61c8bb6d...). This is a convenience hash; the Go `ComputeHash` algorithm was not run.
Auditor Version: plan-auditor (Sonnet 5.5 session)

Reasoning context ignored per M1 Context Isolation. Only the committed artifacts at `e7dbdeef3` were read: spec.md, plan.md, acceptance.md, progress.md.

## Claim

C1. The SPEC's format is sound (12-field frontmatter, GEARS-shaped REQs REQ-001..REQ-012, Out of Scope H3s, Tier M artifact set). Supported — see Must-Pass Results.
C2. The RED-now counts in acceptance.md reproduce on this tree. Supported for every count measured (table below); exit codes are not recorded by the SPEC and are therefore a gap, not a pass.
C3. The upstream facts F1-F4, F6 match the primary sources. Supported. F5 matches. OQ-1 and OQ-2 as written do NOT match the primary source (D4).
C4. The acceptance criteria can distinguish a correct implementation from a wrong one. NOT supported for REQ-001, REQ-005, REQ-008, REQ-011 (D1-D3, D7): wrong implementations pass.

## Evidence

All commands run from `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1416`. Counts are the verbatim printed `grep -c` output.

### E1. Tree attribution

```
git rev-parse HEAD            -> e7dbdeef35e6d3d53b1fb886ccb84f77640433e1
git branch --show-current     -> WT-ultracode-toggle-wording
git status --porcelain        -> (empty) both before and after the hugo run below
```

### E2. RED-now re-measurement, rule source (RS) and mirror (RM)

```
grep -c 'combines `xhigh` reasoning' RS            -> 1     (RM -> 1)    matches AC-001 RED
grep -c -F 'leaves the effort level unchanged' RS  -> 0     (RM -> 0)
grep -c -F 'v2.1.284' RS                            -> 0     (RM -> 0)    matches AC-001 positive-pin RED
grep -c 'step back with `/effort high`' RS          -> 1     (RM -> 1)    matches AC-002 RED
grep -c -F 'effort ultracode off' RS                -> 0     (RM -> 0)
grep -c -F -- '--effort ultracode' RS               -> 0     (RM -> 0)
grep -c -F '"ultracode": true' RS                   -> 0     (RM -> 0)
grep -c -F 'ultrathink.' RS                         -> 1     (RM -> 1)    retention control, green
cmp RS RM ; echo cmp=$?                             -> cmp=0             matches AC-003 pre-state
```
RS=`.claude/rules/moai/workflow/dynamic-workflows.md`, RM=`internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md`. Line 111 of RS is the bullet the SPEC cites; its text contains both "combines `xhigh` reasoning with automatic workflow orchestration" and "step back with `/effort high` for routine work" (read directly).

### E3. RED-now, docs-site (per locale ko / en / ja / zh, same order)

```
WF  grep -c -E '/effort ultracode.*xhigh'        -> 1 1 1 1     (AC-004 RED)
WF  grep -c '/effort ultracode.*/effort high'    -> 1 1 1 1
WF  grep -c 'v2.1.284'                           -> 0 0 0 0
WF  grep -c -F '/effort ultracode off'           -> 0 0 0 0
WF  grep -c -E '^\| `/effort ultracode` \|'      -> 1 1 1 1     (shape control)
ML  grep -c -E '^/effort ultracode # xhigh'      -> 1 1 1 1     (AC-005 RED)
UW  en 'Reasoning effort: set to `xhigh`' -> 1 ; ja 'Reasoning effort: `xhigh` に設定' -> 1 ; zh 'Reasoning effort：设置为 `xhigh`' -> 1
UW  ko '`xhigh`로 올라갑니다' -> 1 ; '/effort high`로 한 단계 내립니다' -> 1 ; '세 가지가 함께 바뀝니다' -> 1 ; '세션 경계를 넘지 않습니다' -> 1 (control)
CM  en 'simultaneously an `/effort` level' -> 1 ; ko '동시에 `/effort` 레벨입니다' -> 1 ; ja '`/effort` のレベルでもあります' -> 1 ; zh '也是一个 `/effort` 等级' -> 1
```
Heading counts (`grep -c '^#'`), matching AC-008 baselines exactly: WF 11/11/11/11; ML 13/13/13/13; UW 24/20/20/20; CM 23/23/18/18 (ko/en/ja/zh).
`docs-site/.locale-parity-baseline` lists `advanced/ultracode-workflows.md` (L18) AND `claude-code/foundations/commands.md` (L35).

### E4. Citation check (spec.md section 3 line numbers)

Observed via `grep -n`: WF ko L113, en/ja/zh L105; ML ko L96, en L101, ja L98, zh L88; UW ko L65/L67/L71, en L112, ja L109, zh L109; CM ko L79 + L137, en L79 + L137, ja L80, zh L80. Every cited line carries the cited sentence. The en/ja/zh UW pages do not contain a `/effort high` return phrase (only ko does), consistent with REQ-007's "(ko only)" qualifiers.

### E5. Sweep for in-scope surfaces the SPEC missed

`grep -rIl -i ultracode` over the repo excluding `.git`, `worktrees`, `specs`, `reports`, `CHANGELOG*` listed 62 files; reading every xhigh / `/effort` / return-phrase line (plus whole-section reads of the four UW pages) found NO additional surface asserting the xhigh coupling or the `/effort high` return beyond the SPEC's change map. Files outside the map that mention ultracode: session-handoff family + moai.md (+ template mirrors), `.moai/docs/session-handoff-appendix.md` (+ mirror), model-policy.md x4, handoff.md x4, sub-agents.md, `AGENTS.md.tmpl`, Go handoff code. Their ultracode statements are keyword/restore-guidance statements and stay true. Out-of-scope reasons checked and true: model-policy lists `/effort low|...|ultracode|auto` as accepted syntax (ko L87, en L92); sub-agents.md L77 "`ultracode` sessions exempt"; `AGENTS.md.tmpl` L36 capability name; `.moai/docs/harness-delivery-strategy.md` header dated 2026-06-03. Unnamed but unaffected: `session-handoff-appendix.md` L67 (+ mirror). `internal/template/contract_mode_guided_test.go` L677 lists the rule file only as a "Kickoff" classification entry ("R"); the edit does not touch the Kickoff word. `internal/template/scripts/gen-catalog-hashes.go`: `grep -c 'rules\|dynamic-workflows'` -> 0, and `internal/template/embed.go` L28 `//go:embed all:templates` — plan section B's claims hold.

### E6. Upstream facts (fetched with `curl`; WebFetch was not available in this session)

Changelog (`raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md`, HTTP 200): line 406, inside `## 2.1.284` (lines 342-444): "Changed Ultracode into its own toggle in `/effort` (Tab, or `/effort ultracode [on|off]`): it no longer forces xhigh effort and stays on at any effort level". F1 confirmed. No later entry (2.1.285-2.1.287) changes ultracode.

model-config page (HTTP 200, tag-stripped text):
- "Ultracode is a Claude Code setting rather than a model effort level ... at whichever effort level the session runs at." F2 confirmed.
- "Turning ultracode on or off with /effort or the ultracode setting leaves the effort level unchanged. The --effort ultracode flag and the Agent SDK effortLevel: "ultracode" value turn it on and also set the level to xhigh. Picking a level in the /effort slider or the /model picker leaves ultracode as it was." F3 confirmed.
- "run /effort ultracode to turn it on for the current session or /effort ultracode off to turn it off. In the /effort slider, press Tab to flip the Ultracode toggle, then Enter to apply it". "The /effort ultracode off form, the slider toggle, and keeping ultracode on at effort levels other than xhigh require Claude Code v2.1.284 or later. Before v2.1.284, turning on ultracode set the session to xhigh effort, picking another level turned it off". F4 confirmed.
- "The persisted effortLevel setting and the CLAUDE_CODE_EFFORT_LEVEL environment variable don't accept ultracode. If ... an effort cap sets the session's level, ultracode stays on at that level." F6 confirmed.

workflows page: "/effort ultracode lasts for the current session; to have every session start with it, set the ultracode setting. Turn it off with /effort ultracode off when you return to routine work." F5 confirmed.

settings-reference page `ultracode` entry (this is the passage the SPEC read past):
"The key doesn't change the session's effort level ... Claude Code reads this key but never writes it: /effort ultracode turns ultracode on for the current session only. ... Default: unset ... Per-session overrides: /effort ultracode ... /effort ultracode off turns it off for one session when this key is true. The --effort ultracode flag also turns it on for one session, at xhigh effort, and requires Claude Code v2.1.203 or later ... This and the /effort ultracode off form require Claude Code v2.1.284 or later. Before v2.1.284, ultracode: true ran the session at xhigh effort, and an effort cap below xhigh kept ultracode off."

### E7. Structural verbs

- Traceability verb (D4-style, AC-4 / AC-5): the verbatim `awk` was REFUSED by the worktree guard (both inline and `awk -f`); a python re-implementation of the same regexes was run instead. Result: `verb-form COLLECTED (colon/bold defs): 0`, `loose-form defs: 12 [REQ-001..REQ-012]`, `UNCOVERED (loose): []`, `ORPHAN: []`. So the verb as specified collects 0 definitions from this spec (GAP form), because the REQs are written `- REQ-001 — ...` (no colon, no bold). Mapping was therefore judged by hand from the AC headings: AC-001 (REQ-001, REQ-004), AC-002 (REQ-002, REQ-003), AC-003 (REQ-004), AC-004 (REQ-005, REQ-009), AC-005 (REQ-006), AC-006 (REQ-007), AC-007 (REQ-008), AC-008 (REQ-009, REQ-010), AC-009 (REQ-010), AC-010 (REQ-011, REQ-012), AC-011 (no REQ named).
- CN-4 ordering verb (python/grep equivalent; the awk form is refused): plan has 4 milestones M1-M4 (`### M1`..`### M4`), 0 `Exit:` bindings (`grep -n -i -E '^(\*\*)?exit'` on plan.md prints nothing). acceptance.md ordering candidates: L27 "after the edit" (retention pin) and L38 "before and after" (shape control); neither names a milestone on a forbidden side. No `CONFLICT:`; the observed absence of exit bindings means milestone order was not mechanically bound to ACs (read by hand: no conflict).
- D8: `grep -c syscall spec.md` -> 0. D7: only self-reference `SPEC-CC-ULTRACODE-TOGGLE-001` (no external SPEC). MP-7: `grep -n 'NEEDS CLARIFICATION' plan.md` -> no output, rc 1; research.md absent.
- `moai spec lint SPEC-CC-ULTRACODE-TOGGLE-001` -> `No findings - all SPEC documents are valid`, rc 0. See Baseline-attribution for why this is NOT cited as corroboration.
- AC-009 control: `grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content` -> no output; Mermaid LR/RL grep -> no output; `hugo --minify --gc --destination <scratch>` in `docs-site/` -> rc 0, 0 lines matching `WARN|ERROR` (output redirected to the session scratchpad, tree stayed clean).

## Baseline-attribution

- Tree measured: worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1416`, HEAD `e7dbdeef35e6d3d53b1fb886ccb84f77640433e1`, clean. The SPEC's RED pin is `c50da9c2f`, HEAD's parent in the log; the SPEC-only commit on top touched no scope file, and the counts above reproduce, so the pin holds.
- Judging build of `moai spec lint` and the `moai` binary: `moai_cp/20260925_122548-1711-gd194083fb`, built 2026-09-30T08:37:30Z. `git merge-base --is-ancestor d194083fb HEAD` -> rc 0, and the reverse -> rc 1: the installed build is a STRICT ANCESTOR of HEAD, i.e. it may lag. By verification-claim-integrity section 2.2 the lint result is attributed to an older build, and by the citation discipline it is also silent on this axis (the traceability verb collects 0 definitions in the same REQ form). The lint pass is therefore recorded as "the tool said nothing on this axis", never as corroboration.
- Upstream pages were read on 2026-10-02 in this run. The SPEC's own fetch tool summarised pages; this audit read the raw page text with `curl` and tag stripping.
- Not carried over from any earlier measurement: every number above was produced in this run.

## Gaps

1. WebFetch was not available in this session; `curl` + tag stripping was used instead. Quoted upstream text is the stripped HTML text.
2. The verbatim traceability `awk` and CN-4 `awk` verbs were refused by the worktree guard (`awk` with a program body and `awk -f`). Replacements: a python re-implementation (traceability) and `grep` over the same records (CN-4). Their output matches the verbs' output form but is not the verb's own output.
3. Two other compound commands were refused (a `sed` with a runtime-computed path variable, and one `git diff --stat` inside a loop); both were re-issued with literal paths and plain commands. Nothing was substituted by reading.
4. AC-011 (`make build`, `go test ./internal/template/...`) was not executed: scoped-verification discipline, and the mirror is not yet edited. The `Grep` tool was also absent; `grep` via Bash was used.
5. OQ-1 (does the slider toggle persist) was not tested live. The primary text narrows it (D4) but does not settle it.
6. No cross-backend audit (`audit_multi`, `codex_audit`, `glm_audit`) was invoked; the task did not request one and no receipt exists.
7. The ko/en/ja/zh target wording does not exist yet (plan phase), so wording quality, native idiom and the Mermaid/emoji rules are unobserved.

## Residual-risk

- Even with all defects fixed, ACs are string-presence checks on prose; they cannot prove a locale's new sentence reads correctly or states all of F2-F5.
- If upstream changes ultracode again (the changelog is at 2.1.287 today), the fix text ages; the SPEC pins v2.1.284 as introduction version only, which limits this.
- The settings-key route is stated as "persistent"; per the primary text it is read-only for Claude Code, and the slider path remains unverified for persistence.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `- REQ-001` through `- REQ-012`, spec.md L48-L59, sequential, no gaps or duplicates, uniform three-digit padding (python loose-form collection: 12 ids, in order).
- [PASS] MP-2 GEARS/EARS format (layer judged: requirement layer, `REQ-XXX` in spec.md; ACs are in acceptance.md and were graded under Group 4): every REQ carries `shall` (L48-L59). REQ-001, 004-008 ubiquitous; REQ-002 `**When** ...`, REQ-003 `**While** ...`, REQ-009 `**When** ...`; REQ-010-012 are `shall not` (legacy negative form, allowed at Score 1.0 but never canonical). Borderline: the `When`/`While` triggers in REQ-002, REQ-003 ("When that bullet names the ways ...", "While describing persistence") are not real events or states (optional finding D9).
- [PASS] MP-3 frontmatter (read field by field against spec-frontmatter-schema.md): id `SPEC-CC-ULTRACODE-TOGGLE-001` (L2); title quoted (L3); version `"0.1.0"` quoted (L4); status `draft` (L5); created / updated `2026-10-02` (L6-L7); author (L8); priority `P2` (L9); phase `"v3.2.0 target"` is a release label, not a lifecycle stage (L10); module `"docs-site/content"` (L11); lifecycle `spec-anchored` (L12); tags comma string (L13); optional `tier: M` (L14). No snake_case alias. plan.md and acceptance.md carry no frontmatter, so no `status:` there. The `id` regex shape with multiple hyphenated segments follows existing SPEC practice; the lint run is unattributed (see Baseline-attribution).
- [N/A] MP-4 language neutrality: wording correction of prose about a Claude Code feature; no language-specific tooling is named or required.
- [PASS] MP-5 D7: the only `SPEC-...` token in spec.md is its own id; no retired/superseded SPEC is referenced.
- [PASS] MP-6 D8: `syscall` appears 0 times in spec.md.
- [PASS] MP-7 clarification gate: no `[NEEDS CLARIFICATION` marker in plan.md; research.md absent (Tier M), so that half is N/A.
- [N/A] MP-8 RED-now re-execution: no acceptance criterion is classified release-blocking (acceptance.md carries no such label), so the four-element obligation is not triggered. Reason stated per the MP-4 precedent. The RED counts were re-executed anyway (E2-E3) and reproduce; the exit-code field is absent from every cell (D5).
- [PASS] MP-9 ordering consistency: CN-4 collected 4 milestones M1-M4, 0 exit bindings, 2 non-binding candidates (acceptance.md L27, L38), 0 `CONFLICT:`. M1 (canonical wording) -> M2 (ko, then en, ja, zh) -> M3 (mirror) -> M4 (verification) matches the AC "Given M1 and M3" / "Given M2" preconditions. Observed absence of `Exit:` lines is stated, not read as a pass of the mechanical check.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | REQs name exact surfaces and lines (E4 all true). Ambiguity: REQ-001 "leaves the session's effort level unchanged" vs AC pin `leaves the effort level unchanged` (D2); REQ-008 "describe it as a toggle offered in /effort" has no stated wording; plan M2 gives no canonical docs-site sentence, so ko wording is left to the run agent. |
| Completeness | 1.00 | 1.0 | HISTORY (L19), Context/WHY (L23), Change Map/WHAT (L66), HOW in plan.md, REQUIREMENTS (L46), ACCEPTANCE in acceptance.md, `### Out of Scope - <topic>` H3s with specific bullets (L78, L86, L90); all 12 fields. The factual error in OQ-2 is scored under D4, not here. |
| Testability | 0.50 | 0.50 | Several ACs pass a wrong implementation: AC-001 zero pin is one literal phrase; AC-004 never pins `--effort ultracode`, `"ultracode": true` or session scope; AC-007 has a vacuous positive pin and no L137 pin; AC-010 slider check is "reviewer reads the bullet". See D1-D3, D7. |
| Traceability | 0.75 | 0.75 | By hand: every REQ is named on an AC heading, no orphan REQ id. AC-011 names no REQ (D7); REQ-008's second sentence (L137), REQ-011's version half and REQ-012's line-level half are named by an AC but not asserted by any command. The mechanical verb collected 0 (D8). |

Aggregate 0.75 against Tier M threshold 0.80 -> below threshold.

## Defects Found (structured defect-list)

D1. AC-004 / REQ-005, REQ-009 — acceptance.md:L36-L38 — The docs-site workflows-row AC pins only `v2.1.284` and `/effort ultracode off`. Nothing asserts `--effort ultracode` (the launch-flag exception), `"ultracode": true` (the persistence route) or the current-session scope in any of the four locales, although REQ-005 requires all of them and REQ-009 requires those literals verbatim. A mutant that deletes the `xhigh` and `/effort high` text and appends only `v2.1.284 ... /effort ultracode off` passes every AC-004 command, and the over-correction mutant the SPEC lists (acceptance.md L77) is caught only on the rule source, not on the 16 docs files. — Severity: major — Class: blocking — Required fix: add per-locale positive pins to AC-004 (`grep -c -F -- '--effort ultracode' WF` >= 1 and `grep -c -F '"ultracode": true' WF` >= 1 for ko/en/ja/zh, plus a per-locale scope pin, e.g. the `ultracode` setting literal), each with a stated RED-now value (0 today, measured).

D2. AC-001 / REQ-001 — acceptance.md:L20, spec.md:L48 — Two problems on the canonical wording. (a) The zero pin `combines \`xhigh\` reasoning` forbids one literal phrase, while REQ-001 forbids "combines or forces xhigh": a rewrite such as "forces `xhigh`" or "runs at `xhigh`" passes the pin, and no AC pins "forces". (b) The positive pin is `leaves the effort level unchanged`, but REQ-001 text reads "leaves the session's effort level unchanged"; an implementer copying the REQ phrase fails the AC, and the plan (M1) never states the pin phrase. — Severity: major — Class: blocking — Required fix: restate the AC zero pin as a regex over the bullet that covers the coupling in any wording (for example `grep -c -E 'ultracode.*(forces|combines|sets|raises).*xhigh'` on the line-111 bullet, with the launch-flag clause worded so it does not match), and make the positive pin phrase identical to the REQ text (or put the exact required phrase in REQ-001 and plan M1). Record the RED-now values.

D3. AC-007 / REQ-008 — acceptance.md:L52-L53, spec.md:L55 — The commands-page AC pins only the removal of the L79/L80 sentence. (a) The positive pin `grep -c 'ultracode' CM` >= 1 is vacuous: it is already green today (ko L48, L79, L137 all contain `ultracode`), so it asserts nothing about the work. (b) REQ-008's second sentence (ko/en L137, "lists it among the levels") has no pin at all. (c) "describe it as a toggle offered in /effort" has no pin. — Severity: major — Class: blocking — Required fix: replace the vacuous pin with a RED-now-0 positive pin on the required replacement wording (per locale), add a zero pin for the ko/en L137 level-list wording, and measure both RED values on this tree.

D4. spec.md OQ-1 / OQ-2 / REQ-011 — spec.md:L40-L41, L58 — OQ-2 asserts "The `ultracode` settings entry in the fetched settings reference names no minimum version." The settings-reference entry says "This and the `/effort ultracode off` form require Claude Code v2.1.284 or later. Before v2.1.284, `ultracode: true` ran the session at xhigh effort" (E6). The SPEC states a fact about a page that the page contradicts, and builds REQ-011's second clause on it. OQ-1 is also under-evidenced: the same entry says "Claude Code reads this key but never writes it: /effort ultracode turns ultracode on for the current session only", which bears directly on whether the slider can persist and is not recorded. This also leaves the settings-key route unqualified: on a Claude Code older than v2.1.284 the key forces xhigh, the very coupling the SPEC removes. — Severity: major — Class: blocking — Required fix: rewrite OQ-2 to what the page says; decide REQ-011's second clause as a deliberate wording choice (omit the version, or state it) rather than an evidence gap, and if the version is stated, add its pin; add the "never writes" sentence to OQ-1's evidence and narrow the open question to the slider alone. Add a pin for whichever way REQ-011's second clause is decided.

D5. RED-now cells — acceptance.md:L3-L5 and every AC — The document adopts "Two-cell discipline" and a "pins its RED-now state" claim, but each cell holds a command and a printed count, no exit code and no verbatim stdout, and acceptance.md L5 says exit codes "were not captured separately". The commands also use abbreviations (`RS`, `WF`, `<loc>`, undefined `ko-UW` at L48) and per-locale loops, so a cell is not a literal single invocation a reader can re-run. No AC is classed release-blocking or regression-guard, so the obligation level of the cells is undeclared. I re-ran the counts (E2-E3) and all reproduce, so this is a completeness defect of the cells, not a wrong RED. — Severity: minor — Class: blocking (the document states a criterion it does not meet) — Required fix: capture exit codes for the RED cells (one run), expand the abbreviations to literal paths in the cells or ledger, define `ko-UW`, and state each AC's class (release-blocking or regression-guard).

D6. plan.md cross-references — plan.md:L19, L47 — "RED-now commands of AC1-AC8" and "baselines in `acceptance.md` AC9" use non-existent IDs; the heading-count baselines are in AC-008, while AC-009 is the hugo gate. A run agent following L47 reads the wrong AC. — Severity: minor — Class: blocking (internal inconsistency) — Required fix: correct to `AC-001`..`AC-008` and `AC-008`.

D7. AC scope vs REQ scope — acceptance.md:L48, L66, L70 — (a) AC-006 demands a positive pin `/effort ultracode off` in all four UW pages, which REQ-007 does not require (REQ-007 only forbids the xhigh/`/effort high` text); the AC is stricter than its REQ, so the two can disagree. (b) AC-010's slider check is "reviewer reads the bullet", not binary, and only targets RS although REQ-011 covers every edited page; REQ-012's "lines outside the change map" has no line-level check. (c) AC-011 names no REQ. — Severity: minor — Class: blocking (binary-testability of a stated criterion) — Required fix: either add the `/effort ultracode off` requirement to REQ-007 or drop the pin; make the slider check a command (for example `grep -c -i 'slider'` per edited file with an expected value, or a forbidden-phrase regex) over all edited files; cite REQ-004 on AC-011.

D8. REQ definition form — spec.md:L48-L59 — REQs are written `- REQ-001 — ...`. The plan-auditor traceability verb reads list-item definitions only when the id is followed by a colon or is bold; it collected 0 definitions here and the SPEC linter reported "No findings" on an unattributed, possibly lagging build (E7). Mechanical traceability is therefore unobserved for this SPEC. — Severity: minor — Class: optional — Required fix: write definitions as `- **REQ-001**:` (or `- REQ-001:`) so collectors read them.

D9. GEARS shape of REQ-002, REQ-003 — spec.md:L49-L50 — the `When` and `While` clauses name no event or state ("When that bullet names the ways ...", "While describing persistence"), so they read as ubiquitous requirements wearing a trigger. — Severity: minor — Class: optional — Required fix: write them as plain ubiquitous `shall` statements or give a real trigger.

D10. Minor accuracy notes — spec.md:L57, L80, L88 — REQ-010 says only `advanced/ultracode-workflows.md` is divergence-listed, but `claude-code/foundations/commands.md` is also listed (baseline L35). `.moai/docs/session-handoff-appendix.md` (+ template mirror) mentions ultracode but is not in the out-of-scope list; it is unaffected. `autonomous-workflow-strategy.md` is described as a "2026-06" record; its header carries no 2026-06 date. — Severity: minor — Class: optional — Required fix: add CM to REQ-010's note, name the appendix pair under "remain accurate", and drop or verify the 2026-06 date claim.

D11. Unobserved assertion in the SPEC — acceptance.md:L62, progress.md:L14 — AC-009 asserts "the build is structurally green-today"; progress.md records that `hugo` was not run at plan time. The claim was unobserved when written. (This audit ran it: rc 0, 0 `WARN|ERROR`, E7, so the claim is true.) — Severity: minor — Class: optional — Required fix: record the observed hugo rc and warning count in the RED/control cell with the tree SHA.

## Regression Check (Iteration 2+ only)

Not applicable: iteration 1.

## Recommendation

Verdict FAIL at 0.75 (Tier M threshold 0.80); no must-pass failure. Fix the blocking list and re-audit; the re-audit is scoped to D1-D7 plus the CN-4 ordering verb in full.

1. D4 first (it settles REQ-011 and OQ wording, which changes what the ACs must pin): correct OQ-2 against settings-reference text; add "never writes" to OQ-1; decide the version clause.
2. D1, D2, D3: add the missing per-locale positive pins and the wording-independent zero pins; align the pin phrases with the REQ text; record each RED-now value by measuring it on the tree.
3. D5, D6, D7: capture exit codes, expand abbreviations, fix `AC1-AC8` / `AC9` references, define `ko-UW`, reconcile AC-006 with REQ-007, make the slider check a command, trace AC-011.
4. Optional (D8-D11): REQ definition form with a colon or bold id, REQ-002/003 phrasing, the small accuracy notes.

What held up under attack: all cited line numbers and quoted sentences exist at the cited places; every RED count reproduces; F1-F6 match the primary sources; no missed in-scope surface in the repo sweep; template mirror claims (no catalog regeneration, `go:embed`) are true; the out-of-scope reasons are true; the hugo gate is green on the unedited tree.

## Operational Notes (unverified)

- Measure the post-fix AC set for vacuity before adoption — run each positive pin on this tree and confirm `0`, run each zero pin and confirm `>= 1`. Status: assumption (not run here; no fix exists yet).
- Re-run `moai spec lint` with a build made from the tree, passing the binary by path, and confirm requirements collected N > 0 before citing a clean lint. Status: inferred (rule: verification-claim-integrity section 2.2 and the citation-discipline clause; installed build d194083fb is a strict ancestor of HEAD).
- Measure OQ-1 live if the slider claim is ever wanted: toggle in a fresh session, `grep -n ultracode` the settings files, start a new session. Status: assumption.
