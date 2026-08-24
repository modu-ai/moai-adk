# SPEC Review Report: SPEC-WEB-TODO-QUEUE-001

Iteration: 1/2 (Tier M ceiling)
Verdict: **PASS**
Overall Score: **0.86** — harmonic mean of the four dimensions, as the skeptical-evaluation contract
requires. The arithmetic mean would have been 0.868; the gap is small here only because no dimension
is badly weak, and I report the harmonic figure so a future weak dimension cannot be masked.
Threshold: 0.80 (Tier M)
Blocking findings: **2** (D1 minor, D8 moderate)

Reasoning context ignored per M1 Context Isolation. The three artifacts, the ratified split design
§4, the parent SPEC, and the parent's audit report were read; the parent's *reasoning* was used only
to locate the findings routed here (F3, F6, F11, F13) and the defect shapes to test for — never as
grounds for accepting a claim.

Audit tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207`, HEAD `ee039da30`.
The SPEC attributes its baselines to `dfbf828a6`. That commit exists in this tree and
`git diff --stat dfbf828a6..HEAD -- internal/cli/todo.go internal/web internal/kanban/backlog_store.go`
prints nothing, so every baseline it states is measurable unchanged at HEAD. All re-measurements
below were taken at `ee039da30`.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-WTQ-[0-9]*' spec.md | sort -u` yields exactly
  `REQ-WTQ-001 … REQ-WTQ-008`: eight ids, sequential, no gap, no duplicate, uniform three-digit
  padding. `grep -c '^- \*\*REQ-WTQ-'` prints `8`, so every id is a requirement bullet rather than a
  cross-reference.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-WTQ-*` in
  `spec.md` §B), not the verification layer. All eight carry a GEARS pattern: 001 unwanted
  (`shall not`), 002/003/004/008 ubiquitous (`shall`), 005/006 event-driven (`When …, … shall`), 007
  compound (`While …, when …, shall`). The `Given … When … Then` form of every `AC-WTQ-*` is the
  correct verification-layer format and is graded under Group 4, never here. Two form deductions are
  recorded as D3 and D4 below; neither breaks a pattern.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types:
  `id`, `title`, `version: "0.1.0"` (quoted), `status: draft` (in enum, per
  `spec-frontmatter-schema.md:41` + its Status Enum), `created`/`updated` ISO `2026-08-24`, `author`,
  `priority: P2`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` comma-separated string. No
  rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`) appears. Three extra
  fields (`era`, `tier`, `related_specs`) are permitted additions, and `tier: M` is load-bearing for
  the ceilings below.
- **[N/A] MP-4 Section 22 language neutrality** — single-language SPEC. Every file it touches is Go,
  `.templ`, or the console's own `i18n.js`; §C.4 states, and I confirmed, that none of
  `internal/web`, `internal/kanban`, `internal/cli` has a mirror under
  `internal/template/templates/`. The four-locale obligation in REQ-WTQ-008 is the console's UI
  locales (`en`/`ko`/`ja`/`zh`), a different axis from the 16 programming languages. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — four distinct SPEC ids are referenced. Measured
  statuses: `SPEC-KANBAN-TODO-CLI-001` `in-progress`, `SPEC-SESSION-TELEMETRY-001` `draft`,
  `SPEC-WEB-CONSOLE-015` `draft`, `SPEC-WEB-CONSOLE-REDESIGN-001` `completed`. None is
  `retired`/`superseded`/`archived`; all four directories exist. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` prints `0` on all three
  artifacts. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' plan.md` exits 1 with no
  output. `research.md` is correctly absent at Tier M, so the plan.md half is the whole check here.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75 band | Every requirement has one reading. Deductions: REQ-WTQ-007's second sentence closes on a non-`shall` limitation clause and uses the soft word "correct" (D4); REQ-WTQ-004 states package structure where behaviour would do (D3). Neither would lead two engineers to different implementations. |
| Completeness | 0.85 | 0.75 band | HISTORY (`spec.md:21-25`), WHY (§A), WHAT (§A opening + §B), Requirements §B, Constraints §C, Exclusions §D with **five** `### Out of Scope — <topic>` H3 sub-headings each carrying specific `-` bullets, Acceptance §A-§F with a §E traceability table and a §F Definition of Done. Frontmatter 12/12. Deductions: plan.md §C's file enumeration — presented as exhaustive and feeding the Tier signal — names the wrong test file and omits the home-seam consequence of the relocation (D8); no explicit WHAT/Scope heading. |
| Testability | 0.82 | 0.75 band | Every criterion names a command or an observable runtime property, and every absence-satisfied criterion states a measured baseline. I re-ran eight of them and all eight reproduce exactly (§Evidence). Deductions: D1 (the read-only grep drops the write primitives the SPEC's own hazard analysis is built on), D2 (`<base>` placeholder is not executable as written), D5 (REQ-WTQ-006's "unreadable" has no criterion). |
| Traceability | 0.95 | 1.0 band, one deduction | §E maps all 8 requirements to criteria and all 11 criteria back to requirements; I verified the table against the ids actually present in both files — no orphan, no uncovered requirement. Deduction: REQ-WTQ-007's second (worktree-served) clause is covered by no criterion. Acceptance §D states the reason and the reason is sound — asserting the absence of an event would assert a race — so this is a documented gap rather than a lapse. |

**Aggregate — harmonic mean**: 4 / (1/0.85 + 1/0.85 + 1/0.82 + 1/0.95) = 4 / 4.62509 = **0.8648**.
Reported as 0.86. Above the Tier M threshold of 0.80.

---

## Claim

The SPEC's factual claims about the tree are true; the four routed findings (F3, F6, F11, F13) are
closed or substantially closed; the write hazard is exercised by a criterion that targets the impure
branch specifically; the read divergence is closed by both a requirement and a rendering criterion;
the event-vocabulary argument reproduces in full against the source; the route surface is enumerated
completely and each element is real; the budget is 8/16 requirements and 11/16 criteria; the boundary
with the parent holds in both directions.

## Evidence

### 1. The write hazard and the criterion that exercises it (brief item 1)

```
$ grep -n 'func resolveTodoQueueRoot\|ResolveGitDirs\|func fallbackTodoQueueRoot\|func adoptLocalTodoQueue\|os.MkdirAll\|os.Rename\|os.WriteFile' internal/cli/todo.go
66:func resolveTodoQueueRoot() string {
68:	if dirs, err := gitcore.ResolveGitDirs(base); err == nil && dirs.CommonDir != "" {
89:func fallbackTodoQueueRoot(base string) string {
115:func adoptLocalTodoQueue(base, fallbackRoot string) {
124:	if err := os.MkdirAll(fallbackRoot, 0o755); err != nil {
128:	if err := os.Rename(local, target); err == nil {
139:	_ = os.WriteFile(target, data, 0o600)
```

```
$ awk 'NR>=89 && NR<=102 {print NR": "$0}' internal/cli/todo.go
 94: 	if err != nil {
 97: 		return filepath.Join(base, ".moai", "state", "kanban")
 99: 	root := filepath.Join(home, ".moai", "todo", todoQueueProjectKey(base))
100: 	adoptLocalTodoQueue(base, root)
101: 	return root
$ awk 'NR>=115 && NR<=124 {print NR": "$0}' internal/cli/todo.go
118: 	if _, err := os.Stat(target); err == nil {
119: 		return
121: 	if _, err := os.Stat(local); err != nil {
122: 		return
124: 	if err := os.MkdirAll(fallbackRoot, 0o755); err != nil {
```

Every line number the SPEC cites is exact: `:66`, `:68-70` (git-resolvable, returns
`filepath.Dir(dirs.CommonDir)`, no side effect), `:89-102` (fallback), `:94-98` (home-unresolvable,
returns `<base>/.moai/state/kanban`, no side effect), `:99-101` (home-resolvable —
`adoptLocalTodoQueue` then `return root`), `:115-139`, `:118-120` (target exists → early return),
`:121-123` (no local file → early return), `:124`/`:128`/`:139` (the three write primitives).
§A.2's three-branch table reproduces exactly.

AC-WTQ-006 states its preconditions as *git-unresolvable + project-local file present + fallback root
holding no queue file* and cites `todo.go:118-123` for them. Those are precisely the conditions under
which `adoptLocalTodoQueue` falls through both early returns and reaches `MkdirAll`. The criterion
therefore exercises the **impure** branch, and its own body says so and says why a git-resolvable-only
criterion would prove nothing. This is the one place where a criterion that "already passes on the
untouched tree" would have been fatal, and it does not.

### 2. The read divergence, F3 (brief item 2)

REQ-WTQ-005 (`spec.md:131-134`) states read-through as a requirement: on the fallback branch with no
fallback queue file and a project-local one present, "the entry point the console consumes shall
resolve to the project-local root … and shall still write nothing."

AC-WTQ-007 (`acceptance.md:103-111`) asserts the **render**: "When the todo section is rendered, Then
it lists exactly those N items — not an empty queue", plus the `moai todo` side reporting the same N.
Its own text names the division of labour: "AC-WTQ-006 asserts the disk is untouched; this criterion
asserts what is rendered while it is untouched."

F3's required fix was: state the intended behaviour, and extend the criterion to assert the rendered
count. Both halves are present. **F3 closed.**

The predicate is also correct rather than merely stated. `adoptLocalTodoQueue` early-returns when the
target exists (`:118-120`) and when no local file exists (`:121-123`); read-through fires on exactly
the complement. I walked the four states: no local file → neither fires, both read the fallback;
local only → adoption moves it, read-through reads it, same bytes; both present → adoption
early-returns and read-through does not fire, both read the fallback; neither → both empty. They
agree in all four. The one case where they do not is D6 below.

### 3. The conditional requirement, F6 (brief item 3)

REQ-WTQ-007 (`spec.md:137-142`) carries the condition **in its own body**, in both directions:
"**While** the console is served from the checkout that holds the resolved backlog, **when** the
existing `kanban` live-refresh event fires … **While** the console is served from a checkout other
than the one holding the resolved backlog … no live event is guaranteed for a change to the resolved
backlog." F6's required fix was exactly this wording. **F6 closed.**

It does not overclaim in either direction, and the mechanism verifies:

```
$ grep -n 'func (h \*Hub) Watch' -A 16 internal/web/events.go
117:func (h *Hub) Watch(root string, stop <-chan struct{}) error {
125:	for event, dirs := range watchMap {
127:			abs := filepath.Join(root, d)
128:			if err := w.Add(abs); err != nil {
129:				continue // 아직 없는 디렉터리는 건너뛴다 (프로젝트 초기 상태)
131:			pathEvent[abs] = event
```

`Watch` joins `watchMap` onto the **served** root, so a worktree-served console watches the
worktree's `.moai/state/kanban` while the resolution points at the primary checkout's — the second
clause is true. The fallback-root case (`~/.moai/todo/<key>`, held by no checkout at all) also falls
under the second clause, so the binary condition is exhaustive rather than leaving a third state
unstated. `POLL_MS` is `30000` (`app.js:638`) and `startPolling()` is reached only from the
no-`EventSource` branch (`:717`) and `if (failures >= 3)` (`:743`) — §A.5's claim that the poll does
not engage while SSE is healthy reproduces.

### 4. The event-vocabulary claim (brief item 4) — both halves reproduce

```
$ sed -n '25,32p' internal/web/events.go
var watchMap = map[string][]string{
	"spec":    {".moai/specs"},
	"session": {".moai/state"},
	"goal":    {".moai/state/goal"},
	"verify":  {".moai/state/verify"},
	"kanban":  {".moai/state/kanban"},
	"config":  {".moai/config/sections"},
}
$ sed -n '168,183p' internal/web/events.go
// eventFor 는 변경된 경로를 가장 구체적인 감시 경로에 귀속시킨다.
// `.moai/state` 와 `.moai/state/goal` 이 모두 등록돼 있을 때, goal 하위 변경은
// 반드시 "goal" 로 간다 — map 순회 순서에 좌우되지 않는다.
func eventFor(pathEvent map[string]string, changed string) string {
	dir := filepath.Dir(changed)
	best, bestLen := "", -1
	for p, name := range pathEvent {
		if dir != p && filepath.Dir(dir) != p {
			continue
		}
		if len(p) > bestLen {
			best, bestLen = name, len(p)
		}
	}
	return best
}
$ sed -n '637p;643,649p;729,737p' internal/web/assets/app.js
  var EVENTS = ["spec", "session", "goal", "verify", "kanban", "config"];
  function hasArea(area) {
    return document.querySelector('[data-live="' + area + '"]') !== null;
  }
  function refresh(area) {
    if (!hasArea(area) || refreshing) return;
    EVENTS.forEach(function (name) {
      es.addEventListener(name, function () {
        if (name === "config") { configBanner(); return; }
        refresh(name);
      });
    });
```

- *Half one — reuse needs no producer change*: the queue file is
  `.moai/state/kanban/backlog.json` (`todoBacklogPath`, `internal/cli/todo.go:42-44`), inside
  `watchMap["kanban"]` at `events.go:30`. The client registers one listener per `EVENTS` name and
  dispatches to `refresh(name)`, which gates on `document.querySelector('[data-live="kanban"]')`. A
  `/todo` page carrying that attribute is refreshed by the existing event with zero producer-side
  change. **Reproduces.**
- *Half two — a seventh name would be actively harmful*: `if len(p) > bestLen` is **strictly**
  greater, so on equal-length watch paths the first entry encountered wins and the winner is decided
  by Go's randomized map iteration order. A seventh event watching the same `.moai/state/kanban`
  would be exactly such a tie. The function's header comment (`:168-170`) records the longest-path
  rule as the fix for precisely that ordering dependence. **Reproduces.**

Both halves hold. This is not a blocking finding; it is the opposite — the strongest-evidenced claim
in the SPEC. One wording note only: the header comment records the *fix*, and the SPEC says it
"records as a fixed bug", which is a fair paraphrase rather than a misreading.

### 5. The route surface (brief item 5) — all five elements enumerated, all five real, baselines exact

```
$ sed -n '/templ rail/,/^}/p' internal/web/shell.templ | grep -c '@navRow'
5
$ grep -n '@navRow' internal/web/shell.templ
130:			@navRow(vm, "overview", "Overview", "/")
131:			@navRow(vm, "kanban", "Kanban", "/kanban")
132:			@navRow(vm, "specs", "Specs", "/specs")
133:			@navRow(vm, "monitor", "Monitor", "/monitor")
134:			@navRow(vm, "settings", "Settings", "/settings")
$ awk '/templ iconAt/,/^}$/' internal/web/icons.templ | grep -c 'case "todo"'
0
$ grep -c '"nav\.todo"' internal/web/assets/i18n.js
0
$ grep -c '"nav\.kanban"' internal/web/assets/i18n.js
4
```

`shell.templ:196-204` shows why each of the five is load-bearing rather than a list padded for
completeness:

```
templ navRow(vm ShellVM, id, en, href string) {
	<a class="nav__row" href={ templ.SafeURL(href) } aria-current={ attrCurrent(vm.Area == id) }>
		@iconAt(id, 16)
		<span data-i18n={ "nav." + id }>{ en }</span>
```

`@iconAt(id, 16)` takes the nav id, and `iconAt`'s switch has no `default` arm — a missing
`case "todo"` renders nothing, silently, exactly as §C.2 and plan.md M2 state. `data-i18n` is
`"nav." + id`, which is what makes `nav.todo` mandatory in four maps; `nav.kanban` measuring `4`
(`i18n.js:31`, `:650`, `:1271`, `:1892`) confirms `4` is the right expected count. `aria-current`
keys on `vm.Area == id`, which is the `Area` value (`screens.go:23`). Route registration sits with
its peers (`app.go:157-160` page routes, `/specs` at `:171`). Every §C.2 row is verified, and none
is redundant.

I additionally tested AC-WTQ-002's `awk` range for vacuity, because a range that matched nothing
would return `0` forever regardless of implementation:

```
$ awk '/templ iconAt/,/^}$/' internal/web/icons.templ | wc -l
      36
$ awk '/templ iconAt/,/^}$/' internal/web/icons.templ | grep -c 'case "'
16
$ grep -c 'case "' internal/web/icons.templ
29
```

The range spans `templ iconAt` (`icons.templ:68`) through the templ block's closing `}` and captures
16 of the file's 29 `case "` occurrences — the whole `iconAt` switch and nothing from the sibling
function. The scoping is deliberate and the criterion is non-vacuous.

### 6. Baselines, and the F9 shape (brief item 6)

Every absence-satisfied criterion states a baseline, and each one I re-ran reproduces:

| Criterion | Command | Stated baseline | Measured at HEAD |
|---|---|---|---|
| AC-WTQ-001 | scoped `Mutate(\|acquireLock` scan | `0` | `0` |
| AC-WTQ-001 note | unqualified 3-token scan | `23` | `23` |
| AC-WTQ-002 | `awk … \| grep -c 'case "todo"'` | `0` | `0` |
| AC-WTQ-002 | `sed … \| grep -c '@navRow'` | `5` | `5` |
| AC-WTQ-004 | `grep -c 'ResolveGitDirs' internal/cli/todo.go` | `1` | `1` |
| AC-WTQ-010 | `sed -n '/^var watchMap/,/^}/p' … \| grep -c '":'` | `6` | `6` |
| AC-WTQ-010 | `sed -n '637p' … app.js` | the six-name `EVENTS` line | identical |
| AC-WTQ-011 | `grep -c '"nav\.todo"' … i18n.js` | `0` | `0` |

§A.1's claim that `internal/web` does not reference the backlog today also reproduces:
`grep -rn "Backlog" internal/web` returns exactly one hit,
`internal/web/assets/i18n.js:483: "f.workflow.todo.enabled.title": "Backlog queue (todo)",` — a
translation string, no Go file, no template.

The read-only criterion specifically: F9's shape was a criterion whose prose qualifier ("against any
`.moai/state` path") no grep could express. AC-WTQ-001 does **not** reproduce it — it drops the
qualifier and states the executable form instead (production Go, comment lines excluded), and
`acceptance.md:23-28` explains that substitution explicitly. **F9's shape is avoided.** What it does
instead is drop a *token*, which is D1.

AC-WTQ-011's runtime half is real: `internal/web/i18n_governance_test.go` exists and
`go test -run TestI18n ./internal/web/...` matches nine test functions
(`TestI18nUntranslatedValues:217`, `TestI18nAllowlistNoOrphans:288`, `TestI18nKeyCoverageForward:408`,
`TestI18nKeyCoverageReverse:423`, and five more).

§A.1's store citations are all exact: `NewBacklogStore:240`, `BacklogPathForRoot:249`,
`QueuedBacklogCountForRoot:277`, `Load:299`, `Mutate:341`, `BacklogItem:66-72` (the five fields
`{ID, Text, AddedAt, SpecID *string, State}`), `BacklogState:51-61` (`queued|picked|dropped`).
`Load` (`:299-305`) delegates to `load` (`:309`) and takes no lock — the claim holds.

### 7. Budget and form (brief item 7)

Requirements: **8** of the Tier M ceiling of 16. Criteria: **11** of 16. The two ceilings are applied
independently and neither is approached; plan.md §B:30-32 states the same two counts and draws the
correct conclusion (room for a re-audit to add without raising a ceiling). Verified counts, not
quoted ones: `grep -c '^- \*\*REQ-WTQ-' spec.md` → `8`; the AC ids present in acceptance.md are
`AC-WTQ-001 … AC-WTQ-011`, eleven, sequential, no gap or duplicate.

### 8. Traceability (brief item 8) and the boundary

§E's table maps REQ-WTQ-001…008 to criteria and covers AC-WTQ-001…011, with no id appearing in one
document and not the other. Against the parent: `SPEC-WEB-CONSOLE-015/spec.md:267-271` excludes the
`/todo` route, its nav entry, its icon case, its area value, queue-root resolution, the
resolution/adoption separation, and reading/listing/badging backlog items — the exact set
REQ-WTQ-002/003/004 claim. Nothing is left unclaimed and nothing is claimed twice; the parent's
HISTORY (`:27`) records the carve-out and the deliberate id gap, and its plan.md `:130` names this
SPEC as the owner. Against the ratified design, `spec-split-design.md` §4's B-001…B-008 map
one-to-one onto REQ-WTQ-001…008 and B-010…B-020 onto the eleven criteria.

## Baseline-attribution

Every measurement above was taken in this run, in
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207` at HEAD `ee039da30`, with the commands quoted
verbatim beside their output. The SPEC's own attribution commit `dfbf828a6` was verified to exist
(`git cat-file -t dfbf828a6` → `commit`) and to be identical to HEAD across every cited source file
(`git diff --stat dfbf828a6..HEAD -- internal/cli/todo.go internal/web internal/kanban/backlog_store.go`
→ empty output), so its stated baselines and my re-measurements are attributed to the same tree
state. No figure in this report is carried over from the parent's audit or from any other tree.

## Gaps

Named explicitly; none of these is asserted as a defect.

- **G1 — I executed no Go test.** The brief forbade a full-suite run and I ran no targeted one
  either. Every runtime-shaped criterion (AC-WTQ-002's rendered rail, AC-003, AC-005, AC-006's mtime
  assertion, AC-007, AC-008, AC-009, AC-010's markup half) is judged on whether it is *expressible*
  as a test, not on whether it passes. It cannot pass today — the implementation does not exist.
- **G2 — closed during this audit.** The linked-worktree fixture AC-WTQ-005/006 need already exists:
  `internal/cli/todo_queue_root_test.go:49-54` (`addGitWorktree`, via `git worktree add`) plus the
  `userHomeDirFn` override at `:121-123`. What that reading also produced is D8.
- **G3 — closed during this audit.** No import cycle: `internal/kanban/board.go:19` already imports
  `gitcore "github.com/modu-ai/moai-adk/internal/core/git"`, and
  `grep -rn 'internal/kanban' internal/core/git/` returns nothing.
- **G4 — REQ-WTQ-007's second clause has no criterion**, by design and with a stated reason
  (`acceptance.md:147-150`). I accept the reason; I record the uncovered clause as a gap rather than
  scoring it as a traceability break.
- **G5 — I ran no SPEC-lint tooling** (`moai spec lint` or equivalent) against these artifacts. Form
  judgements above are mine, read from the rubric, not a tool's.

## Residual-risk

- **R1 — `w.Add(abs)` skips a directory that does not exist yet and never retries**
  (`events.go:128-130`). On a project whose `.moai/state/kanban` is absent at server start, no
  `kanban` event ever fires, in any checkout. REQ-WTQ-007's first clause is phrased "when the event
  fires", so it is not falsified, and AC-WTQ-009 covers the absent-file render — but an operator who
  adds the first card while the console is running sees nothing until reload. No criterion would
  catch this, and it is a pre-existing producer property this SPEC neither creates nor owns.
- **R2 — AC-WTQ-001's runtime half says "every console route is exercised in one test run".** As the
  route set grows, that phrase silently changes meaning. It is the non-vacuous half of the criterion
  and I would not weaken it, but its scope is whatever the test enumerates.
- **R3 — `refresh()` re-fetches `window.location.href` and swaps `.body` wholesale**
  (`app.js:652-657`), so "the `/todo` route's section shall be re-fetched" (REQ-WTQ-007) is satisfied
  by a whole-page body swap rather than a section-scoped fetch. Behaviourally equivalent for this
  SPEC; a future section-scoped refresh would not change the requirement's truth but would change
  what a reader expects it to mean.
- **R4 — the three-artifact set carries no design.md**, correct at Tier M, so M1's two-entry-point
  shape lives only in plan.md §D. It is stated precisely there (pure resolver vs adopt-then-resolve,
  with the third branch named), so this is a risk of the tier, not of the authoring.

---

## Defects Found (structured defect-list)

**D1** — `acceptance.md:15-16` (AC-WTQ-001) against `spec.md:156-160` (§C.1) — the executable grep is
narrower than the scan §C.1 describes, and the tokens it drops are the ones the SPEC's own hazard
analysis is built on. §C.1 says "a naive `grep -rnE 'Mutate\(|acquireLock|os\.WriteFile' internal/web`
returns 23 hits … Scoped to production Go and excluding comment lines, **the same scan** returns 0".
The criterion's actual command is `grep -rnE 'Mutate\(|acquireLock'` — `os\.WriteFile` is gone, and
`os\.MkdirAll` / `os\.Rename` were never there. §A.2 identifies exactly those three primitives
(`todo.go:124`, `:128`, `:139`) as the write hazard this SPEC exists to contain. I measured the
scoped **three**-token form and it also returns `0`:

```
$ grep -rnE 'Mutate\(|acquireLock|os\.WriteFile' internal/web --include='*.go' \
    | grep -v '_test\.go' | grep -vE '^[^:]+:[0-9]+:[[:space:]]*//'
(no output)
```

All three production-Go `os.WriteFile` hits — `server.go:9`, `projectconfig.go:158`,
`projectconfig.go:221` — are comments naming the anti-pattern. So the wider pattern costs nothing and
the narrowing is unforced. — Severity: **minor** — Class: **blocking** (spec/acceptance internal
inconsistency: the document says "the same scan" of a scan it then changes) — Required fix: restore
`|os\.WriteFile` (and preferably `|os\.MkdirAll|os\.Rename`) to AC-WTQ-001's command and state the
baseline as `0`, which is measured above; or amend §C.1 to describe the two-token scan it actually
prescribes.

**D2** — `acceptance.md:133` (AC-WTQ-010) — `git diff <base>..HEAD -- …` is not executable as
written; `<base>` is an unbound placeholder. Every other command in the document runs as typed. This
is a residue of the F11 shape ("a criterion naming no executable check") rather than a fresh instance
of it — the criterion does name a check, it just leaves one argument unresolved. — Severity:
**minor** — Class: **optional** — Required fix: bind it, e.g.
`git diff $(git merge-base origin/main HEAD)..HEAD -- …`, or drop the diff half and keep the two
`sed` baselines, which already pin the six names.

**D3** — `spec.md:125-130` (REQ-WTQ-004) — a requirement stating package structure where behaviour
would do: "relocated into a package that both the command layer and the console import, with the
command layer delegating to it rather than retaining its own copy." The parent's F13 asked for
REQ-030/031 to be converted to a behavioural form, and this is a substantial improvement — no file
name, no symbol, no line number survives, and the behavioural half ("shall perform no filesystem
mutation on any branch", "reachable only from the `moai todo` command path") is the load-bearing
part. What remains is a structural prescription. It is defensible on this SPEC's own thesis — that
one resolution existing *is* the requirement — so I record it as partial closure rather than a
reopening. — Severity: **minor** — Class: **optional** — Required fix: if tightened, restate as the
observable property ("the command layer and the console shall resolve the queue root through one
implementation") and leave the package choice to plan.md §D M1, where it already is.

**D4** — `spec.md:140-142` (REQ-WTQ-007, second sentence) — two form defects in one clause. It closes
on "and no live event is guaranteed for a change to the resolved backlog (§A.5)", a statement of
limitation rather than a `shall`-form behaviour, so one requirement carries both a normative and a
descriptive clause. And "the section shall be correct on load" uses "correct", a soft term the
Testability rubric would flag anywhere it was the sole basis for a criterion. Neither is load-bearing
here: the limitation half is deliberately uncovered (G4) and "correct" is operationalised by
AC-WTQ-005. — Severity: **minor** — Class: **optional** — Required fix: move the descriptive half
into §C as a constraint cross-referenced from the requirement, keeping only `shall` clauses in §B;
replace "correct" with the observable ("shall render the items the resolved backlog holds").

**D5** — `spec.md:135-136` (REQ-WTQ-006) against `acceptance.md:123-126` (AC-WTQ-009) — the
requirement names three conditions (absent, empty, **unreadable**); the criterion fixtures three
(absent, empty, **malformed JSON**). Malformed JSON is a decode failure; unreadable is an I/O failure
(EACCES), a different path through `atomicfile.ReadFile` at `backlog_store.go:310`. One of the
requirement's three stated conditions has no criterion. — Severity: **minor** — Class: **optional**
— Required fix: either add a fourth fixture (noting a permission-based fixture is awkward on
Windows, which §F's cross-build requirement makes relevant), or narrow REQ-WTQ-006's wording to the
three conditions actually asserted.

**D6** — `spec.md:81-84` and `plan.md:78-83` — the claim that console and `moai todo` "agree **by
construction** rather than by coincidence" overclaims on one reachable path. `adoptLocalTodoQueue` is
best-effort throughout: `MkdirAll` failing (`:124-126`), or `Rename` failing followed by `ReadFile`
failing (`:128-137`), leaves the local file in place and the fallback root empty — and
`fallbackTodoQueueRoot` returns `root` unconditionally at `:101` regardless. In that state the
console's read-through reads the project-local file and reports N while `moai todo` reads the empty
fallback and reports 0 — the divergence returns, reversed. The read-through predicate mirrors
`adoptLocalTodoQueue`'s two *early returns* but not its three *failure returns*. I verified this from
the source; I did not construct the failing filesystem, so the reachability is read, not executed.
The underlying `moai todo` behaviour is pre-existing and §D explicitly excludes changing it, so this
is a claim-strength defect, not a scope defect. — Severity: **minor** — Class: **optional** —
Required fix: qualify the claim ("they agree whenever adoption succeeds or is correctly skipped; a
best-effort adoption failure leaves `moai todo` reading an empty fallback, which is pre-existing
behaviour §D excludes from this SPEC") rather than changing the design.

**D8** — `plan.md:34-52` (§C File enumeration) and `plan.md:58-90` (M1) — the file enumeration is
presented as exhaustive ("Twelve entries; the `_templ.go` regenerations bring it to fourteen") and
feeds the Tier signal in §B ("Files | 12-14"), but it is wrong in two ways, both verified:

*(a) It names the wrong test file.* Row 4 reads "`internal/cli/todo_test.go` | M1 | existing adoption
tests re-pointed at the command-path entry point". The adoption and queue-root tests are not there:

```
$ grep -rln 'adoptLocalTodoQueue\|fallbackTodoQueueRoot\|userHomeDirFn' internal/cli/ | grep _test
internal/cli/todo_queue_root_test.go
internal/cli/todo_flag_independence_test.go
internal/cli/update_home_seam_test.go
$ grep -n 'userHomeDirFn\|ResolveGitDirs\|worktree\|adoptLocal' internal/cli/todo_test.go
(no output)
```

`internal/cli/todo_queue_root_test.go` is the file that holds `addGitWorktree` (`:49-54`),
`TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary` (`:73`),
`TestResolveTodoQueueRoot_FallbackNoGit` (`:116`),
`TestTodoQueue_FallbackAdoptsExistingLocalQueue` (`:152`) — the very test AC-WTQ-008 pins — and
`TestTodoQueue_WorktreeSeesPrimaryQueue` (`:218`). `todo_test.go` contains none of them. A run
following the enumeration edits a file with nothing to re-point and leaves the real one untouched.

*(b) It omits the home-seam consequence, which the relocation forces.* The fallback branch resolves
the home directory through `userHomeDirFn`, a package-level seam in `internal/cli` that
`todo_queue_root_test.go:121-123` and `:157-159` override to test the fallback and the adoption. That
seam cannot move with the code: nine non-test files in `internal/cli` use it (`glm.go`, `update.go`,
`update_clean_install.go`, `update_preserve_inventory.go`, `graph.go`, `glm_tools.go`,
`update_template_sync.go`, `init.go`, `todo.go`). So M1 must establish an equivalent injection point
inside `internal/kanban` — otherwise AC-WTQ-006 and AC-WTQ-007, both of which require a controlled
fallback root, have no way to control it. Neither §C nor M1 names this, and "the adoption logic
itself moves **verbatim**" (`plan.md:75`) is precisely the instruction that hides it: the logic is
verbatim, its seam is not.

The underlying move is sound — `internal/kanban/board.go:19` already imports
`gitcore "…/internal/core/git"`, and `grep -rn 'internal/kanban' internal/core/git/` returns nothing,
so the relocation creates no import cycle and `internal/paths.Home()` is reachable from the new home.
The defect is the enumeration, not the design. — Severity: **moderate** — Class: **blocking** (an
enumeration the plan states as exhaustive is not, and a run following it silently edits the wrong
file) — Required fix: correct row 4 to `internal/cli/todo_queue_root_test.go`, and add a row (and an
M1 sentence) for the home-injection seam the relocated resolution needs in `internal/kanban`. Restate
the 12-14 file band if it moves.

**D7** — `acceptance.md:57` — AC-WTQ-011 is placed in §B between AC-WTQ-002 and AC-WTQ-003, out of
numeric order, for topical grouping. The id sequence 001-011 is complete with no gap or duplicate, so
this costs nothing mechanically; it only makes the document harder to read against its own §E table.
— Severity: **minor** — Class: **optional** — Required fix: none required; renumber only if the
document is otherwise edited.

---

## Regression Check

Not applicable — iteration 1. The four findings routed here from the parent's audit
(`plan-audit-iter2-independent.md`) are verified as closures above rather than as regressions:

- **F3 closed** — REQ-WTQ-005 states read-through; AC-WTQ-007 asserts the rendered count. Both halves
  of the required fix are present.
- **F6 closed** — REQ-WTQ-007 carries the condition in its own body, in both directions, and the
  two-clause split is exhaustive over the resolution's possible roots.
- **F11 closed** — AC-WTQ-004 gives the mechanical `grep -c 'ResolveGitDirs' internal/cli/todo.go`
  form the finding asked for, with baseline `1` verified at HEAD.
- **F13 partially closed** — the GEARS `Where` misuse does not recur (no `Where` appears anywhere in
  §B); the implementation-detail-in-requirement shape is much reduced but survives in REQ-WTQ-004,
  recorded as D3.

The two other defect shapes the brief named were tested for specifically and neither recurs: the
F9 shape (a criterion whose prose qualifier no grep can express — AC-WTQ-001 states its executable
form instead) and the F2 vacuity shape (a criterion that already passes on the untouched tree — the
three preservation criteria, AC-WTQ-001, AC-WTQ-008, AC-WTQ-010, each declare themselves as
preservation with a measured baseline, and every other criterion's baseline is non-passing today).

---

## Recommendation

**PASS at 0.86 against the Tier M threshold of 0.80.** All seven must-pass criteria pass or are N/A.
This is a well-evidenced SPEC: of the source claims I re-measured, every one reproduced exactly —
each cited line number, each stated baseline, both halves of the event-vocabulary argument. That is
an unusually high rate and it is the main reason the Testability score is not lower despite three
deductions.

The three things the brief asked me to be hardest on all hold. The write-hazard criterion targets the
impure branch and says why a git-resolvable-only criterion would prove nothing. The read divergence
is closed by a requirement **and** by a criterion that asserts what is rendered — the half F3 said
was missing. The event-vocabulary argument reproduces in both halves against `events.go` and
`app.js`, including the strictly-greater comparison that makes the tie-break claim true.

Where it is weakest is the plan's file enumeration rather than the specification. D8 is the one
finding a run would hit on its first hour: it would open `internal/cli/todo_test.go` looking for
adoption tests that are not there, and it would reach the home-seam question with no ruling to
follow. Both halves were verified by grep, not inferred.

Two blocking findings, neither large. D1 is a one-token edit whose replacement baseline I have
already measured as `0`. D8 is a plan.md correction — one wrong filename and one missing row.
Neither requires a full re-audit: a scoped confirmation that AC-WTQ-001's command carries the write
primitives at baseline `0`, and that §C row 4 names `todo_queue_root_test.go` with the home-seam
consequence enumerated, is sufficient.

Numbered fixes, in the order I would apply them:

1. **D8 (blocking)** — correct `plan.md:41` to `internal/cli/todo_queue_root_test.go`, and add the
   home-injection seam `internal/kanban` will need to §C and to M1's shape description. This is the
   only finding whose absence changes what a run does.
2. **D1 (blocking)** — restore `|os\.WriteFile` to `acceptance.md:15`, or amend `spec.md:156-160` so
   §C.1 and the criterion describe the same scan. Baseline stays `0`, measured.
3. **D2** — bind `<base>` in `acceptance.md:133`, or drop that half.
4. **D6** — qualify the "by construction" claim at `spec.md:81-84` / `plan.md:78-83`. The design is
   right; the claim is one degree stronger than the code supports.
5. **D5** — reconcile REQ-WTQ-006's "unreadable" with AC-WTQ-009's three fixtures, in whichever
   direction is cheaper.
6. **D3, D4, D7** — form only. Worth doing if the document is opened for the above; not worth an edit
   round of their own. Per the over-engineering brake, routing all of these into a revision would
   cost more than it returns.
