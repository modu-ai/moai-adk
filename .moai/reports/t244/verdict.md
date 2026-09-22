# t244 — Verdict: `team-ac-verify.sh` — KEEP DORMANT (do not wire, do not retire)

> Card t244 (G3 gate batch, from the t216 hook-audit split). Worktree
> `.claude/worktrees/t244`, branch `WT-team-ac-verify-wiring`, HEAD
> `e85c55fa9` (local develop, absorbed after it advanced past the
> dispatched `b7462203a` — fast-forward, no conflict, scope-disjoint
> from this card), measured 2026-09-02.
>
> Per the dispatch's [HARD] rule, the call graph was traced before any
> dead/alive call. This card produces a verdict only — no code change
> follows, because the correct disposition is to leave the current state
> exactly as it is.

## Claim

**Keep dormant — wiring is declined, retirement is rejected.** The hook's
absence from settings is a **recorded, completed SPEC decision**, not
drift: SPEC-HOOK-DISCIPLINE-WIRING-001 (status: completed) explicitly
excluded it ("의도적으로 제외 — 파일 보존, 미등록") with an exit criterion
that REQUIRES the hook to stay un-registered. The hook body is a MINIMAL
STUB (full AC verification deferred), so wiring it today would put a
gate-shaped no-op into the task-completion path — a false assurance. The
TaskCompleted event itself is already wired through `handle-task-completed.sh`,
so nothing is lost by leaving the gate un-registered. Retirement is
rejected because the SPEC hard-forbids deleting the file and active
doctrine references its reject path.

## Evidence

### E1 — call graph re-traced (nothing dead, nothing hidden)

- **Settings registration history**: `git log --all --oneline -S
  "team-ac-verify" -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl`
  → **zero commits, across all history** (re-measured on this tree;
  t216 d2's "pickaxe empty across all settings paths and all history"
  reproduced). The hook has never been registered, locally or in the
  template — there is nothing to "restore", unlike t243.
- **Template twin**: byte-identical (`diff` → identical), so any future
  disposition is a distributed act either way.
- **Production Go references (2, both comments)**:
  `internal/hook/user_decision_capture.go:16` (exit 2 "is reserved for
  sync-phase-quality-gate.sh / status-transition-ownership.sh /
  team-ac-verify.sh, NOT this capture path" — doctrine treats it as a
  governance-gate family member) and `internal/codexadapter/output.go:26`
  ("team-ac-verify.sh rejects a task…" — the reject-contract reference).
- **Test references**: `internal/template/hook_official_compliance_test.go:57-73`
  (AC-HOC-001) guards the script's CONTENT (the `--reject` printf carrying
  `ledger_note`) — it asserts nothing about registration, so keeping the
  hook unwired breaks no test.
- **The TaskCompleted event is already wired**: `.claude/settings.json:251`
  and `settings.json.tmpl:275` register `handle-task-completed.sh` (the
  moai-forwarding wrapper with the autonomy-tier guard). The event channel
  the gate would attach to is live; what is missing is only this
  verification gate on top of it.

### E2 — the hook body is a stub by its own declaration

`team-ac-verify.sh` (111 lines) read in full:

- Self-gating dormancy: reads `workflow.yaml`, emits
  `{"status":"dormant",…}` and exits 0 unless `team.enabled: true` —
  wiring it would cost solo sessions nothing, but also would verify
  nothing outside team mode.
- The active path is a **MINIMAL STUB**: header lines 18-23 — "The trigger
  is a MINIMAL STUB (--reject test flag) — **full AC-verification logic
  (parsing acceptance.md, running evidence commands, blocking on AC
  failure) is deferred to a follow-up**." The team-mode active path (lines
  103-110) records the AC reference to a log and always emits
  `status: "allow"`.
- The reject path exists only as the `--reject` test stub (static JSON).

Wiring a gate whose every active branch ends in "allow" would add a
*gate-shaped no-op* to task completion — the exact shape the empty-green /
vacuous-pass lessons warn about: it looks like enforcement, enforces
nothing, and would be cited as evidence that AC verification is active.

### E3 — the unwired state is a completed, recorded decision

SPEC-HOOK-DISCIPLINE-WIRING-001, frontmatter `status: completed`:

- plan.md §A: "세 discipline 훅 중 둘 … wiring한다. **`team-ac-verify.sh`는
  의도적으로 제외(파일 보존, 미등록)**한다."
- plan.md [HARD]: "**`team-ac-verify.sh` 파일 삭제 금지 — 미등록만.**"
- plan.md M2 step 3: "`team-ac-verify.sh`는 **어떤 블록에도 추가하지 않음**
  (의도적 제외)."
- M2 exit criteria: "**team-ac-verify grep 0**" — the SPEC's acceptance
  REQUIRES the settings grep to stay at zero. Wiring the hook today would
  violate a completed SPEC's exit criterion, the inverse of the t243
  situation (where the SPEC demanded the wiring and a commit had silently
  removed it).
- The hook's own header (lines 3-4) records the same intent: "dormant by
  default unless team mode is enabled".
- `hook-independence.md` §4 row (g) caveat codifies the operational
  reading: dormant, forward-looking, "activates only under harness
  `thorough` + team mode prerequisites"; "Do not read row (g) as evidence
  the gate is active."

### E4 — the preconditions the card names are real, but insufficient

- Team mode is a sanctioned (experimental, re-allowed) mode — CLAUDE.md §15,
  flag ships enabled. The *trigger context* exists.
- Harness `thorough` exists. The *policy tier* exists.
- But both preconditions being satisfiable does not make a stub useful:
  with the hook wired, a team-mode TaskCompleted would run a script that
  — in its current state — allows everything. The missing piece is the
  deferred verification logic, not the registration.

## Baseline-attribution

- All greps/pickaxe/diff/test reads: this run, worktree t244 @
  `e85c55fa9`, 2026-09-02.
- The SPEC decision is attributed to SPEC-HOOK-DISCIPLINE-WIRING-001 as
  committed (frontmatter completed; created 2026-06-11, updated
  2026-06-15) — read from this tree.

## Disposition

1. **Do not wire.** The un-registered state is the recorded intent of a
   completed SPEC and matches the hook's stub reality.
2. **Do not retire.** The same SPEC [HARD]-forbids deleting the file;
   doctrine (Ledger Closure (b), codexadapter comment, hooks-system team
   table) references its reject contract; the template twin makes removal
   a distributed act.
3. **Revival conditions (both required, on the record)**:
   (a) the deferred full AC-verification logic is implemented (parse
   acceptance.md, run the AC's evidence command, block on failure — the
   hook's own deferral note names the scope); and
   (b) a follow-up decision explicitly reverses
   SPEC-HOOK-DISCIPLINE-WIRING-001's intentional exclusion (a new SPEC or
   amendment recording why the completed exit criterion "grep 0" is being
   superseded).
   At that point wiring lands as one registration in both settings files
   (template + local mirror), with the double-fire question against
   `handle-task-completed.sh` investigated first (see Gap 2).
4. **No code change in this card** — the verdict is the deliverable.

## Gaps

1. **Whether the "deferred to a follow-up" verification work is queued
   anywhere** was not established — the deferral note names no SPEC or
   card. The queue is lead-owned; this lane did not read or mutate it.
2. **The double-fire interaction between `handle-task-completed.sh` (wired)
   and this gate (unwired) under a future wiring was not tested** — both
   would run on the same event; whether the runtime parallel-runs them and
   how their reject outputs merge is unobserved. Recorded as the first
   investigation item for the revival path.
3. The claim in the hook header that team-mode activation also requires
   "harness thorough" is NOT implemented in the script (it gates only on
   `team.enabled`) — a doc-vs-code mismatch inside the hook's own header,
   recorded here rather than fixed; it binds no behavior while dormant.

## Residual-risk

- The stub's existence invites exactly the wiring mistake this verdict
  declines: a future hook sweep that "wires the unregistered gate" would
  ship a vacuous pass. This verdict is the record that refuses it; the
  revival conditions above are the gate it must clear first.
- If team mode grows into daily use while the gate stays a stub, task
  completion in team mode has NO mechanical AC verification — the gap is
  real, but it is closed by implementing the verification, not by
  registering the stub.
