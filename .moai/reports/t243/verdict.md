# t243 — Verdict: `handle-session-start-navigator.sh` — RESTORE (re-wired in this card)

> Card t243 (G3 gate batch, from the t216 hook-audit split). Worktree
> `.claude/worktrees/t243`, branch `WT-navigator-hook-restore`, HEAD
> `e45054c56` (local develop, absorbed after it advanced past the
> dispatched `b7462203a` — fast-forward, no conflict, scope-disjoint
> from this card), measured 2026-09-02.
>
> Per the dispatch's [HARD] rule, the call graph was traced before any
> dead/alive call. The trace also re-derived what the card's premise
> means: the thing that was "deleted" is the WIRING ENTRY, not the file.

## Claim

**Restore — and the restore is executed in this card.** The navigator
SessionStart hook was a planned deliverable of SPEC-PROJECT-NAVIGATOR-001
(its own AC-PN-009), its wiring entry was removed as apparent collateral in
the build-recovery commit `7171880a9` (#1409, 2026-08-08) with the removal
recorded nowhere, the hook script and its template twin survived untouched
and functional the whole time, the surrounding feature is alive, and the
distributed reference documentation still describes the wiring as existing.
The hook is fail-open with zero dependencies on the moai binary, so
re-wiring carries no launch risk.

## Evidence

### E1 — what actually got deleted: the wiring, not the file

- File history: `git log --follow -- .claude/hooks/moai/handle-session-start-navigator.sh`
  → **exactly one commit** (`2c87d195f`, SPEC-PROJECT-NAVIGATOR-001 #1354,
  2026-08-05). The file was created and never deleted, locally or in the
  template tree.
- Wiring history: `git log --all --oneline -S "handle-session-start-navigator" -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl`
  → two commits: `2c87d195f` (entry added to the TEMPLATE settings) and
  `7171880a9` (entry removed). Pickaxe over the local `.claude/settings.json`
  returns **nothing — the entry never existed locally** (t216 d2 recorded the
  same: "no (never — pickaxe … is empty)").
- The #1409 settings diff is exactly one hunk (`@@ -10,12 +10,6 @@`)
  deleting the navigator entry (6 lines); the commit body describes only
  internal-Go reconciliation and records **no reason for the settings
  change** — consistent with collateral, and matching t216 d2's reading
  ("removed as apparent collateral in a build-recovery commit").

### E2 — the wiring is a SPEC deliverable, not an optional extra

- `acceptance.md:55-60` — **AC-PN-009 "SessionStart ambient auto-brief
  fires"**: the additionalContext "emitted via hookSpecificOutput
  … from `handle-session-start-navigator.sh`". AC-PN-010 (fails open) and
  AC-PN-012 (staleness advisory) likewise presuppose the hook being wired.
- `plan.md:34` — "EDIT: `internal/template/templates/.claude/settings.json.tmpl`
  — register the new SessionStart hook." (a named plan deliverable).
- `plan.md:67` — the `--brief` design keeps TWO surfaces deliberately:
  "(a) SessionStart hook … the ambient auto-brief … AND (b) `/moai project
  --brief` … **Dropping either leaves a real gap**."
- The shipped feature docs assert the wiring as current:
  `.claude/skills/moai-workflow-project/references/navigator.md:155-183`
  ("### Ambient auto-brief (SessionStart hook) … See the SessionStart hook
  registration in") and its template mirror (:148-176). A user reading the
  shipped docs today is told a registration that does not exist.

### E3 — the feature is alive; nothing superseded the hook

- Hook body: self-contained bash — depends on `git` + coreutils only, never
  on the `moai` binary or `jq` (hook-independence mode A does not apply).
  Fail-open on every error path; exit 0 unconditionally.
- Template twin exists and is byte-identical to the local script (re-measured
  today: both 4382 bytes; t216 d2's table row agrees).
- The Navigator feature ecosystem survives: `navigator-regen.sh` /
  `navigator-enrich.sh` / `navigator-audit.sh` all present in the template
  skill tree; #1409 itself *strengthened* the feature ("Navigator Detect
  wired into postToolHandler.Handle for Write/Edit/NotebookEdit") while
  removing this SessionStart surface — the PostToolUse Detect layer is a
  different axis (impact recording), not a replacement for the session-start
  brief.
- `internal/template/navigator_hook_test.go` executes the script directly —
  a live guard that passes on the current tree.

### E4 — restore executed

- Both `internal/template/templates/.claude/settings.json.tmpl` and the local
  tracked `.claude/settings.json` gain the navigator entry in the SessionStart
  block, in the **current fail-open wrapper form** (#1505 hardening style —
  the entry is restored to today's registration idiom, not the pre-#1409 raw
  form), timeout 5 (the value d2 recorded for the original entry), after the
  existing `handle-session-start.sh` entry where #1409's diff shows it sat.
- Verified: both files parse (local: `python3 json.load` OK; template is a Go
  template — its non-JSON `{{ if .HookOptIn.Enabled }}` token at line 130 is
  pre-existing, not from this change); the two SessionStart blocks are
  byte-identical between local and template (diff of lines 4-21: identical);
  `go test ./internal/template/ -run 'Settings|NavigatorHook|Neutrality|Catalog'`
  → `ok` (includes the hook script execution guard and the neutrality scan).
- `internal/template/catalog.yaml` carries no settings.json.tmpl hash — no
  catalog regeneration needed (grep: 0 hits).

## Baseline-attribution

- All commands: this run, worktree t243 @ `e45054c56`, 2026-09-02.
- Historical claims attributed to their commits by SHA above (`2c87d195f`,
  `7171880a9`); the t216 d2 record (HEAD `a9eb896ce`, 2026-08-24) was read
  from `git show 950cb4399:.moai/reports/t216/d2-unwired-scripts.md` (the
  file is absent from the current develop tree) and its rows independently
  re-confirmed where they are re-measurable today.

## Premise assessment (dispatch's three claims)

| Dispatch premise | Verdict |
|---|---|
| "deleted in the build-recovery commit" | **TRUE — for the wiring entry** (`7171880a9` template settings diff). FALSE for the file: the file was never deleted. |
| "siblings were restored 2 commits later, only this one was left out" | **NOT REPRODUCIBLE from git history.** Neither sibling (compact) was ever deleted as a file (its history: created `adc867545`, edited `7cff166a1` — no deletion), and compact's wiring never existed in any settings file (pickaxe 0 hits). The claim's source record was not found; see Gap 1. |
| "lane-6 measurement overturned the audit's 'intentionally-unwired' verdict" | **PARTIALLY CONFIRMED BY HISTORY, SOURCE NOT LOCATED.** t216 d2 itself already rules the navigator hook "not by-design-unwired" — its 'ambiguous — needs a decision' row says "was wired, then removed as apparent collateral"; the 'intentionally-unwired' reading d2 overturned belongs to the pre-t216 audit. What lane-6's specific measurement added was not found on disk; see Gap 1. |

None of the unresolved premise detail changes the disposition: the verdict
rests on E1-E3, all of which are re-derivable from the current tree.

## Disposition

- **Restore, executed**: template + local settings re-wired (E4).
- Retirement branch is **rejected**, on the record: a SPEC AC (PN-009)
  exists and is being silently un-met; the feature it belongs to is alive;
  the shipped docs describe the wiring as present. Retiring would require
  retiring the AC and the docs with it — a product decision no commit ever
  recorded.
- Revival conditions (for the record, should the feature be retired
  later): removal of the hook + template twin + both wiring entries + the
  AC-PN-009/010/012 rows + the navigator.md hook sections (local and
  template mirror) — all or nothing, per the d2 twin table (removing any
  template-twinned hook is a distributed act).

## Gaps

1. **RESOLVED after publication — the dispatch's source record located and
   measured.** The lead reported the citation target as
   `.moai/reports/t216/d2-unwired-scripts.md`, "existing only in the t216
   worktree, uncommitted". Re-measured here: the file **is committed** —
   `950cb4399` (2026-08-24 21:55:07, the t216 plan-phase investigation
   commit) carries it — but that commit is **not an ancestor of develop**
   (`git merge-base --is-ancestor` → rc=1), is contained only in
   `WT-hook-wiring-drift` (`git branch --contains`), and develop's own
   history has **zero** commits touching `.moai/reports/t216/`. The
   worktree copy is byte-identical to the `950cb4399` blob (diff →
   IDENTICAL). So the dispatch cited a real, committed record that is
   simply unreachable from the integration branch until t216 merges —
   the correct noun is "unmerged", not "uncommitted". With the source in
   hand: d2 classifies the navigator hook "ambiguous — needs a decision"
   with the wired-then-removed reading this verdict's E1 re-derived from
   git history directly, and the "siblings were deleted and restored 2
   commits later" shape appears **nowhere in d2** — it is not in the
   11-script table, not in the compact row. That premise shape remains
   unreproduced on both carriers (git history and the source document).
   The verdict rests on E1-E3, unchanged.
2. **AC-PN-009's real-session verification is not re-run here** — restoring
   the registration and passing the script-execution guard is the lane-scoped
   verification; observing `additionalContext` in a fresh session requires a
   live session start (out of lane scope).
3. **This dogfood project never generated `navigator.md`** (primary checkout
   `.moai/project/navigator/` holds only `symbols/`), so locally the restored
   hook will run and exit-0 silently (fail-open). It becomes observable in
   projects after `/moai project` generates the Navigator.

## Residual-risk

- The t216 d-1 update-path gap stands: projects initialized before
  2026-08-13 cannot receive this template entry via `moai update` (the
  array-element merge hole) — they get the entry only on a fresh `moai init`
  or a hand edit, until that separate fix lands.
- The restored entry spends one 5-second hook slot per session start even
  where `navigator.md` will never exist; the cost is bounded by the fail-open
  early-exit (one `-r` stat) — negligible, but nonzero.
- If a future commit again edits the template settings without the local
  file, the t216 d-1 author-discipline gap recurs silently — the same class
  of defect as this card; the doctor-drift diagnostic (t216 d-1 option C)
  remains the structural guard, deferred to its own card.
