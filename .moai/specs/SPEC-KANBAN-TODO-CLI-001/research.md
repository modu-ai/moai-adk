# research.md — SPEC-KANBAN-TODO-CLI-001

Verified evidence gathered during plan phase (2026-08-14). Every claim below names the command or file read that produced it. Nothing here is inferred from text patterns alone.

## 1. The CLI subcommand does not exist

- `moai todo` → `Unknown command "todo"` (orchestrator probe, twice, 2026-08-14).
- `ls internal/cli/todo*.go` → no matches (re-confirmed this session). No `newTodoCmd` factory exists; `internal/cli/root.go` registers sibling commands via `rootCmd.AddCommand(newXxxCmd())` factories (`newConstitutionCmd()`, `newStateCmd()`, `newChainCmd()`, `newHarnessRouterCmd()`, …) — a `newTodoCmd()` registered there is the repo convention for the new verb.
- `grep -n 'Backlog' internal/kanban/*.go` → only `ColumnBacklog` (the board column enum). No backlog-store symbols exist yet; the name space is free.

## 2. The current surface races (the motivating incident)

- `.moai/state/kanban/backlog.json` (read in full, 2026-08-14): card t12 records the incident — "3 cards lost + t4/t5/t6 ID collision". Observable in the same file: t4 is absent while t13–t19 exist; id sequence is non-monotonic (t1, t2, t10, t5, t8, t6, t7, t11, t12, t13–t19). t6 and t7 both carry `[DROPPED — …]` prefixes — post-hoc human repair of collided entries.
- `.claude/skills/moai/workflows/todo.md` instructs the model to Read and Write `backlog.json` directly, with "Write the file atomically (write a sibling temp file, then rename)". The rename protects a crash mid-write; it does not close the minutes-wide read→write window between sessions. Five sessions share this checkout (kanban mode), so concurrent model writers are the normal case, not an edge.
- Skill contract lines preserved verbatim into REQ-TODO-013/014: "A missing file is an empty queue, not an error. A malformed file is reported and left untouched — never silently reset, because the operator's intent is the one thing here that cannot be regenerated." Also: ids are "never reused after removal"; `picked` items persist; `done` removes the row.

## 3. Record shape (load compatibility is mandatory)

Production file (`.moai/state/kanban/backlog.json`, v1):

```json
{"version":1,"items":[{"id":"t1","text":"…","added_at":"<RFC3339>","spec_id":null,"state":"picked"}]}
```

`state` observed values: `queued`, `picked`, `dropped`. `spec_id` is `null` or a SPEC-ID string. The store must read this shape unchanged; the only permitted change is the additive high-water mark (plan.md §D decision c).

## 4. Lock substrate — internal/kanban (the recommended reuse)

Read in full this session:

- `internal/kanban/board_lock.go` (137 lines): `acquireBoardLockImpl(lockPath)` is **path-parameterized** — the substrate is not board-specific. `BoardLockOwner{PID, CreatedAt}` recorded in the artifact at acquisition. Header states the substrate "reuses the repository's existing cross-process per-scope lock PATTERN (internal/spec/lock.go and its platform counterparts: flock on Unix, atomic-create on Windows); internal/lockfile's in-process mutex is neither used nor upgraded."
- `internal/kanban/board_lock_unix.go` (56 lines, read in full): `unix.Open(O_CREAT|O_RDWR|O_CLOEXEC, 0o644)` → `flock(LOCK_EX|LOCK_NB)` → on failure close fd + `ErrBoardLockHeld` (loser writes nothing to the artifact); on success `ftruncate(0)` + write owner record. Kernel releases the flock when the fd closes (incl. process death), so a dead Unix holder blocks nothing — the stale-lock requirement is a Windows-substrate concern (board_lock_clear_windows.go exists for the board).
- `internal/kanban/board_store.go` (290 lines, read in full): `acquireBoardLockSerialized` = 25 ms × 40 retries ≈ 1 s bounded window, then surfaces the error — "two concurrent transitions must BOTH reach the admission decision in turn… The window is bounded so a genuinely stuck holder surfaces as an error instead of a hang." `WriteBoardState` shape: `requireLeadRole` → serialized acquire → defer release with error join ("mutation landed but lock release failed" — Windows release removes the artifact) → `LoadBoard` (absent = empty board legal; unreadable = `ErrBoardUnknown`, repairs nothing) → mutate callback → validation sweep → `writeBoardAtomic` (`MarshalIndent` + `\n`, `os.CreateTemp(dir, ".board-*.tmp")`, `atomicfile.Replace`).
- `internal/kanban/board.go`: `LoadBoard` reads via `atomicfile.ReadFile` (absorbs the Windows delete-pending window on the read side); "the read path performs no repair — recovery is an explicit operator-visible act, never a fallback a read takes (AP-21)."

The backlog reuses the **substrate** (path-parameterized acquire + serialized wrapper + atomic write), not `WriteBoardState` itself: the board entry point is welded to `requireLeadRole`, which the backlog must NOT apply (plan.md §D decision b).

## 5. Lock substrate — internal/session/registry.go (pattern reference, not the substrate)

`Registry.withLock` (registry.go ~449–533, read earlier this session): MkdirAll parent → in-process path-keyed mutex (NB-flock kernel-fairness starvation guard, REQ-CFS3-001) → jittered NB-flock retry to a deadline (`ErrLockTimeout`) → read → mutate callback → `sort.Slice` (deterministic output) → `writeAtomic` → release. Reads (`Query`) bypass the lock — eventual consistency (AP-MSC-002). The in-process mutex layer is registry-specific; importing `internal/session` from `internal/kanban` would be a reverse dependency. The backlog takes the **shape** (lock → read → mutate → atomic write → release; reads lock-free) from here and the **substrate** from internal/kanban.

## 6. Path layout — backlog and board are different directories

- Board: `boardDirSegments = []string{".moai", "state", "kanban-board"}` (board.go) — deliberately distinct from the session record's `stateDirSegments` (AP-24).
- Backlog: `.moai/state/kanban/` (the existing production file's directory). The new lock artifact is a sibling: `.moai/state/kanban/backlog.lock`. No board path constant is reused or changed.

## 7. atomicfile surface (verified exports)

`internal/atomicfile/replace.go`: `Replace(oldpath, newpath string) error`, `ReadFile(path string) ([]byte, error)` (with `read_unix.go`/`read_windows.go` platform split absorbing the Windows delete-pending race), `Claim(path, perm) error`. Same-directory temp + rename; rename(2) is atomic only within one filesystem (AP-20 / REQ-KB-018 precedent).

## 8. Template twin status

`diff .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md` → **identical** (and both clean in `git status`). So today there is no intentionally-neutralized divergence between the twins (the hazard memory warns about from other rules); run phase applies the SAME delta to both and re-verifies with `MOAI_TEMPLATE_LEAK_STRICT=1 go test` (template neutrality: no SPEC ids, no internal dates — the shrunk skill must not cite SPEC-KANBAN-TODO-CLI-001). `make build` re-embeds templates (`//go:embed all:templates` in `internal/template/embed.go`; catalog.yaml regeneration must be committed).

## 9. CLI conventions (internal/cli/CLAUDE.md, read earlier)

- No-prompt static guard: "every interactive-shaped subcommand needs the equivalent grep-based test" — `TestNew_NoAskUserQuestion` precedent.
- Cobra registration via `newXxxCmd()` factory + `rootCmd.AddCommand` (duplicate `Use:` panics at runtime).
- Exit codes 0/1/2; stdout structured for `--json`, stderr human-readable; `filepath.Abs` for user paths; env names as constants in `internal/config/envkeys.go`.
- Cross-platform claim discipline (CLAUDE.local.md §6): `GOOS=windows go build ./...` does NOT compile `_test.go`; the claim "compiles on Windows" requires `GOOS=windows go vet ./...` with cited exit code. This becomes a run-phase AC.

## 10. kanban-dispatch.md dependency on these verbs

`.claude/rules/moai/workflow/kanban-dispatch.md` (template copy re-read this session) names `/moai todo` as THE entry mechanism: "`/moai todo \"<description>\"` appends an item… `/moai todo` alone lists the queue. After a `/clear`, the lead presents the queue through `AskUserQuestion` and the operator picks." The shrunk skill keeps pointing at this rule; the verbs must therefore support exactly: append, list, present-for-selection (read-only `next`), and pick (`next <n> [--spec]`). The rule's "The lead never picks for the operator" is why REQ-TODO-005 keeps bare `next` read-only.

## 11. SPEC ID pre-write self-check (executed)

```bash
ID="SPEC-KANBAN-TODO-CLI-001"
[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
```

Verbatim output: `PASS`. Existing `SPEC-KANBAN-*` directories (BOARD / BOOTSTRAP / RENAME / WORKTREE-001) contain no TODO-CLI — id is unique.

## 12. Release-target basis for `phase`

Recent main commit `2c40e8106` "chore(version): align ldflags fallback and project config to v3.1.0-rc.2" — the next unreleased version is v3.1.0, hence `phase: "v3.1.0 target"` (phase is a release target, never a lifecycle stage token).
