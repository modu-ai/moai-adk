# SPEC-LAUNCHER-ENTRY-FLAGS-001 — Decision Index

Decisions surfaced during plan assembly. Q1–Q15, Q17, and Q19 were answered by the operator on 2026-10-02 and are recorded as decided (Q1 is superseded by Q10; Q9's count sentence is superseded by Q17); Q7 is closed as moot. Q16 was raised as a lane-orchestrator ruling on measured evidence and was then CONFIRMED by the operator ("proceed as is"). Q18 was settled in two steps: the operator first chose an option whose premise and three safeguards the update code contradicts, the facts were returned, and the operator then chose Option X. The options weighed and their measured impact stay recorded in `plan.md` §B.2. No row is open. Labels use the fixed four-label vocabulary. Authority register consulted: `.moai/project/product.md`, completed SPECs' HISTORY and `## Amendments` rows, `.moai/config/sections/*.yaml` operator settings, the constitution — no committed artifact decided any row before the operator answered, so none was `DECIDED` or `POLICY-COVERED`; the operator verdicts are not yet a committed artifact and are not cited as authority. No row carries a recommendation.

Verdict source for every decided row: operator answer via AskUserQuestion, lane-3, 2026-10-02 (Q16 additionally records its origin as a lane-orchestrator ruling on evidence).

### Q1: Does "exactly two forms" mean two new root entry tokens beside an unchanged verb surface, or the whole launcher entry surface reduced to two forms? (plan.md OD-6)

Label: FOUNDER
Authority anchor: none — no committed artifact decides the scope of the card's phrase.
Why unresolved: the card names `-f` and `-l` and the existing factory surface but not whether the kanban `-k` shapes are in scope.
Operator verdict: DECIDED, then SUPERSEDED — first verdict: new tokens only; the kanban `-k` entry and everything not named in the other verdicts stay untouched. Superseded by the Q10 verdict (Kanban Mode removed, Tier L). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q2: How does the entry choose Claude, GLM, or Codex? (plan.md OD-3)

Label: FOUNDER
Authority anchor: none — no committed artifact decides how an entry names a backend.
Why unresolved: launchers default to Claude when no backend is named and no default-backend setting was found; the options ranged from no new token to a new configuration key.
Operator verdict: DECIDED — the flags ride on the backend verbs (leader: `moai cc -f`, `moai glm -f`; lane: `moai cc -l`, `moai glm -l`, `moai codex -l`); no verb-less `moai -f` / `moai -l` root form is created; the verb carries backend selection, so there is no `--backend` token and no default-backend setting. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q3: What happens to the short `-l` that today means `--leader <name>`? (plan.md OD-1)

Label: FOUNDER
Authority anchor: none — the short was introduced as "free in this tree" by a completed SPEC (SPEC-FACTORY-LANE-JOIN-SOCKET-001), which does not decide a conflict with a later lane entry.
Why unresolved: retiring the short breaks any script using it; keeping both meanings leaves one letter with two meanings.
Operator verdict: DECIDED — the short `-l` of `--leader <name>` is retired; the long `--leader <name>` stays; `-l` becomes the lane-entry flag on the verbs above. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q4: Are the existing `-f lane` / `-f lane-<n>` / `moai codex -f lane` forms and the `-f <N>` shape kept, deprecated, or removed, and over what window? (plan.md OD-4)

Label: FOUNDER
Authority anchor: none — no repository policy file states a back-compat window for launcher flags (`.moai/config/sections/sunset.yaml` governs quality-gate relaxation, not flags); the repository's own precedent refuses legacy spellings with a named canonical form instead of running a warning window.
Why unresolved: removal depended on the backend question (Q2) because the verb forms were the only route to a GLM lane and a Codex lane.
Operator verdict: DECIDED — remove now, with no deprecation window; each removed form is refused with a message naming the canonical form: `moai cc|glm -f lane`, `moai cc|glm -f lane-<n>`, `moai codex -f lane`, and the `-f <N>` shape. Re-measurement on tree `a6d3e6fd4` found `-f <N>` already refused on cc, glm, and codex (exit 1, one line), so for that shape the requirement pins the refusal and rewrites its message. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02; re-measurement in progress.md PV-16.)

### Q5: Which text changes with this SPEC? (plan.md OD-5)

Label: FOUNDER
Authority anchor: none — no committed artifact scopes documentation for this change.
Why unresolved: the footprint ranged from help sites only to every document naming an entry command, across four locales.
Operator verdict: DECIDED — everything that mentions an entry command: help text, rules, hooks, agent/skill/template mirrors, READMEs (4 locales), docs-site (4 locales), AGENTS.md; the four-locale same-commit rule and the Template-First mirror cycle stay explicit. The scope widened with Q10 and Q12 to everything that carries the word kanban (156 files by the measured pattern, research.md §R1). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q6: Is the lane entry bare `-l` only, or `-l [lane-<n>]` with an optional numbered label? (plan.md OD-2)

Label: FOUNDER
Authority anchor: none — the numbered form existed on the verb surface as `-f lane-<n>`, but no committed artifact decided whether the lane entry carries it.
Why unresolved: the stale-run notice and the leader notice print the numbered form in four locales.
Operator verdict: DECIDED — `-l` takes no numbered label; `-l lane-<n>` and any argument form is refused with one line naming the canonical `-l`; slot assignment stays automatic through the existing shared lane-slot claim (reused, not re-invented). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q7: Do operators or scripts use the short `-l <leader>` today?

Label: EVIDENCE-NEEDED
Authority anchor: none.
Why unresolved: the answer needs usage data that does not exist yet — no launcher usage telemetry was found.
Operator verdict: closed as moot — Q3 was decided to retire the short, so the measurement no longer informs a pending choice; the unmeasured usage stays recorded as a gap.

### Q8: Is `-f` combined with a lane-shaped `--name` (an explicit-name lane entry) refused, or left as it is? (plan.md OD-7)

Label: FOUNDER
Authority anchor: none — the earlier verdicts named `-f lane`, `-f lane-<n>`, and `-f <N>`; none named this spelling.
Why unresolved: by the dispatch truth table `-f --name lane-<n>` selects the lane branch today (code reading, not run), so a numbered lane entry spelled with `-f` survives beside `-l`.
Operator verdict: DECIDED — refuse the explicit-name lane spelling (`moai cc|glm -f --name lane-<n>` and equivalents) with one line naming the canonical `-l`; the requirement is a real refusal, not a conditional. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q9: With automatic slots, does the leader notice print one launch line per declared lane or the command once? (plan.md OD-8)

Label: FOUNDER
Authority anchor: none — the notice printed one numbered line per lane and a test pinned the line count; no committed artifact decides the shape once lines are identical.
Why unresolved: every line would read `moai <entry> -l`; one line per lane keeps the pinned contract, printing once changes the notice shape and that test.
Operator verdict: DECIDED — the leader notice prints the lane launch command once, not once per lane; the pinned line-count test is updated. The accompanying count sentence is superseded by Q17. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q10: Do the kanban-token numbered and count shapes stay, or is Kanban Mode removed — and if removed, how far, and in which SPEC? (plan.md OD-9)

Label: FOUNDER
Authority anchor: none — completed SPECs (the SPEC-KANBAN-* family) introduced the mode; none decides its removal.
Why unresolved: verdict Q1 had left `-k` untouched, which kept a leader-count and numbered-lane entry reachable; Q4 and Q6 removed the same ideas from the `-f` and `-l` spellings.
Operator verdict: DECIDED — "-k is deleted and retired, we no longer use it, REMOVE KANBAN MODE". Extent: the WHOLE kanban mode — the launcher `-k` entry and parser, the companion plan/run/sync session join path, launcher help and notices, the kanban rules/skills/hooks/agent mentions, and docs/mirrors. Placement: included in this SPEC, re-tiered to Tier L. Back-compat position: `-k` in every shape is refused after removal, with a message naming what to use instead (factory mode `-f` / `-l`), reusing the retired-entry refusal mechanism. Supersedes Q1. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q11: Are the six `MOAI_KANBAN*` markers the factory reads kept, renamed with a compatibility window, or renamed outright — and does the Codex lane child keep the `MOAI_KANBAN_LABEL` stamp? (plan.md OD-10)

Label: FOUNDER
Authority anchor: none — `internal/config/envkeys.go` names the markers and states they are load-bearing, but no committed artifact decides their fate under the mode's removal.
Why unresolved: six markers (`_ID`, `_LEAD_ADDR`, `_LEAD_NAME`, `_SETTINGS_INJECTED`, `_BACKEND`, `_CARD`) are read by factory code across 95 files (34 production, 61 test), and a joining lane reads `MOAI_KANBAN_ID` from a live leader process's environment by name, so a rename without a dual read breaks discovery across binary versions. Separately, the Codex lane child is stamped with `MOAI_KANBAN_LABEL` under a code comment that says the factory card verbs read it; the measurement found those verbs read `MOAI_FACTORY_WORKER`.
Operator verdict: DECIDED — KEEP the names and string values of the six factory-read markers (`MOAI_KANBAN_ID`, `_LEAD_ADDR`, `_LEAD_NAME`, `_SETTINGS_INJECTED`, `_BACKEND`, `_CARD`); no rename, no dual read. The Go constants that carry them are renamed (the string values are frozen) and an acceptance criterion pins the frozen strings (REQ-015, AC-016). REMOVE the `MOAI_KANBAN_LABEL` stamp on the Codex lane child (one production file and four tests), with a RED/GREEN check that the Codex lane still launches and identifies itself through `MOAI_FACTORY_WORKER`; readers of the label outside this repository's Go code remain a Gap (REQ-012, AC-013). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q12: Which names that carry "kanban" survive the mode's removal — the Go package, state directories, web route, SSE key, statusline segment, types and helpers? (plan.md OD-11)

Label: FOUNDER
Authority anchor: none.
Why unresolved: `internal/kanban` also holds the todo queue, factory slots, locks, and landing code (179 importing files); the web route, the live-update key, and `data-live` are shared with the todo queue card and a declared frontend contract; the statusline label reads the todo queue.
Operator verdict: DECIDED — rename EVERYTHING that carries the word and survives, including the Go package `internal/kanban`, with the old→new mapping stated (design.md §4.7); every persisted or wire name is stated as renamed or frozen with its evidence (design.md §8); a rename that would break existing user data or cross-version behavior is returned as an open question — the code showed one such case, the renamed distributed rule files in user projects (Q18); no persisted data name was found that breaks. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q13: Are the `kanban-dispatch*.md` rules and the kanban-named docs pages renamed, or kept under their names with the kanban-only sections stripped? (plan.md OD-12)

Label: FOUNDER
Authority anchor: none.
Why unresolved: the three rules are the operating doctrine factory leaders and lanes run under (one always-loaded); 52 files outside specs reference the path, four tests reference it, and one companion rule is a declared fork pair.
Operator verdict: DECIDED — RENAME the three rule files to factory names (local and template), strip the kanban-only sections, re-point the references and the pinned tests; docs-site pages (12: three per locale), menu entries, redirects, and the home banner are each renamed with a redirect or deleted with a redirect, four locales in the same change (REQ-019, REQ-022; design.md §5). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q14: What happens to `workflows/factory.md`, the `workflows/moai.md` factory contract, and the chain-contract record fields? (plan.md OD-13)

Label: FOUNDER
Authority anchor: none.
Why unresolved: the skill documents a "Factory Mode" that is the single-session plan→run→verify→sync chain entered by `--factory`/`-f` — the pre-rename chain lineage — while `-f` today is the multi-lane leader; the file never says "kanban", so the removal search by word does not find it.
Operator verdict: DECIDED — REWRITE `workflows/factory.md`, the `moai.md:210` sentence, and the Record chain-field documentation to match today's `-f`/`-l` model, stating precisely what the text asserts (design.md §6, assertions A1–A7), with acceptance criteria that check the text against observed behavior (REQ-020, AC-021). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q15: Is the foreman skill, the bare-loop driver, and the factory leader's reference to them removed, or kept as an unattended queue watcher under another name? (plan.md OD-14)

Label: FOUNDER
Authority anchor: none.
Why unresolved: the foreman skill and `.claude/loop.md` are kanban-named, but the factory leader notice tells the operator the queue is polled by "the kanban foreman loop (bare `/loop`)" in four locales, and `moai todo --auto` implements the foreman contract from the CLI.
Operator verdict: DECIDED — KEEP the foreman skill, `.claude/loop.md`, `moai todo --auto`, and the leader notice's queue sentence, under FACTORY names: the skill directory and name become `moai-factory-foreman`, with the catalog entry, hash, three catalog tests, and references updated (REQ-019). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q16: Do the four kanban sentences in `CLAUDE.md`, `moai-constitution.md`, and `agent-common-protocol.md` go through the constitution amend gate, or are they edited directly with `moai constitution validate` before and after? (plan.md OD-15)

Label: FOUNDER
Authority anchor: none — measured instead: none of the sentences' distinctive phrases is in the zone registry, no kanban rule carries a `[ZONE:Frozen]` tag, `moai constitution guard` takes rule IDs rather than file diffs, `amend` requires a registered rule ID, and `moai constitution validate` is OK on this tree (progress.md PV-26..PV-28, PV-38).
Why unresolved: the amend command takes a registered rule ID and the four sentences have none, so the amend route would first need them registered; whether the operator wants that is a policy question the registry cannot answer.
Operator verdict: DECIDED — CONFIRMED by the operator ("proceed as is") after being raised as a lane-orchestrator ruling on evidence. Edit the four constitution-slot sentences directly, running `moai constitution validate` before the first edit and after each edit; do not touch registered Frozen clause strings (`CONST-V3R2-036..038` in the section anchored `#user-interaction-boundary`); the stop condition is kept — any validate failure, or any need to touch a registered string, is a blocker returned to the orchestrator (REQ-021, AC-022). Nuance recorded: `agent-common-protocol.md:27` sits in that section, and `:75` is in the body of `#language-handling` (`CONST-V3R2-039`, registered clause is the header sentence). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02; origin: lane orchestrator ruling on evidence.)

### Q17: What does the leader notice state about lanes now that new launches always declare one lane? (plan.md OD-16)

Label: FOUNDER
Authority anchor: none.
Why unresolved: with `-f <N>` and `-k N` removed, a new launch always declares one lane and lanes join on demand through `-l`; a sentence stating the declared count would always say 1, a sentence stating the claimed slots would be a live figure, and a capacity-open wording would state no number.
Operator verdict: DECIDED — the leader notice drops the lane COUNT entirely and states only the command: to start a lane, enter `moai cc -l` / `moai glm -l` / `moai codex -l` in a new terminal; no per-lane lines, no number. The notice code, the four locale strings, and the pinned line-count test change; a text-pinning RED/GREEN pair replaces the line-count test (REQ-009, AC-010). Consequence recorded (author reading, Q19 item b): the free-slot line, built from the declared count, is dropped too. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02.)

### Q18: What happens to the three old rule files already installed in user projects once the `kanban-dispatch*.md` rules are renamed? (plan.md OD-17)

Label: FOUNDER
Authority anchor: none — `internal/cli/update_archive.go:45` archives retired SKILLS; the rule files are handled by the managed-root clean (`internal/cli/update/deploy/deploy.go:75-78,107-185`, spec.md §A.2 rows 21–22).
Why unresolved (before the verdict): the operator chose Option A — a retired-rule-file cleanup step in `moai update` removing the old installed `kanban-dispatch.md`, `kanban-dispatch-detail.md`, and `kanban-dispatch-mechanics.md`, with five safeguards: (1) respect the namespace-protection contract; (2) delete only on a released-content hash match, and back up and report a user-modified file instead of deleting silently; (3) idempotent and fail-open; (4) a fixture RED/GREEN pair for unmodified, user-modified, and absent; (5) a named list in one place, mirrored for Template-First. The code shows: every `moai update` already removes the whole MoAI-managed `.claude/rules/moai` root and first copies every file the template does not carry into `.moai-backups/<timestamp>/pre-clean/`, aborting the removal if that copy fails, so the old files are removed and backed up once the template stops shipping them; (1) holds (not a user-owned namespace); (2) "retain a modified file" cannot coexist with the managed-root removal without changing that contract, and the hash list would be built from 65 committed revisions; (3) conflicts with the clean's tested abort-on-backup-failure; the `defs.DeprecatedPaths` route cannot carry it (its sweep aborts the update on error and a registered path flips V2 detection). The old `kanban-dispatch.md` is always-loaded (26,352 B local), so the outcome matters. The question: X — rely on the existing clean and add only the fixture proof (no production change; modified copies are removed from the live tree and preserved in the backup; deletion is not hash-conditional); Y — add a dedicated retired-rule step with a one-file list of released-content hashes that classifies, backs up, and reports each file by name, fail-open at the step level, while the managed clean still removes and still aborts on a failed backup; Z — Y plus a carve-out so a user-modified copy is retained, changing the managed-root contract, its registry, and its tests, and leaving an always-loaded stale rule in place. Measured impact per option is in `plan.md` §B.2.
Operator verdict: DECIDED in two steps. Step 1: Option A chosen, with the instruction to return a precise question where the namespace-protection contract or the hash-match requirement conflicts with the design in a way the code cannot resolve. Step 2: the conflict was returned with the code facts, and the operator, shown them, chose **Option X** — rely on the existing managed-root clean (pre-clean backup, then removal); add NO production change and NO dedicated retired-rule step; the proof is the fixture test `TestUpdateRemovesRetiredRuleFilesWithBackup` (unmodified, user-modified, absent, bystanders, second-run idempotence), exactly as REQ-019 and AC-020 describe. Accepted consequences, stated plainly: a user-modified copy of an old rule file is backed up and removed (not retained); deletion is not hash-conditional; the progress line names a count, not paths; a failed backup still aborts the removal and the update. AC-018's retired-name allowlist is unchanged (no new Go file). (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02, both steps.)

### Q19: Are the choices the author made inside the verdicts acceptable? (spec.md §D; plan.md §B.3)

Label: FOUNDER
Authority anchor: none.
Why unresolved (before the verdict): each was an author choice within a verdict, not an operator answer: (a) the package name `internal/factory`; (b) the free-slot line is dropped from the leader notice; (c) `GET /kanban` redirects to `/factory`; (d) the `moai chain` and chain-lineage sections of the kanban docs page move to a new page `advanced/origin-trail-chain`; (e) the home banner and the five-sessions image are removed; (f) the `MOAI_FACTORY_WORKERS` value stays a number; (g) the four dependents of the chain contract (`moai.md`, `run.md`, `mode-orchestration.md`, `quality-gates-quality.md`) are reworded and the verify exit gate is retained as specification; (h) the retired-name literals are confined to four named files (the `-k` retirement refusal, the archive list, the legacy state-directory name, the legacy web route); (i) the web live-update key is named `factory`; (j) the three removed docs pages redirect to `advanced/factory-mode`; (k) `FactoryFreeSlots` is left in the slots package, caller-less.
Operator verdict: DECIDED in two steps — every row of spec.md §D is now operator-accepted as written. Step 1: seven rows accepted (rows 2–6, 9, and 10; row 1, the OD-15 ruling, confirmed separately). Step 2: the four rows the first acceptance had not covered — row 7 (`MOAI_FACTORY_WORKERS` stays a number), row 8 (the four chain-contract dependents reworded; the verify exit gate retained as specification), row 11 (the three removed docs pages redirect to `advanced/factory-mode`), and row 12 (`FactoryFreeSlots` stays caller-less) — accepted as written. (Source: operator answer via AskUserQuestion, lane-3, 2026-10-02, both steps.)
