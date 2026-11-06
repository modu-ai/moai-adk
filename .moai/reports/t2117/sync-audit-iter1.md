auditor-model: glm-5.3
verdict: PASS-WITH-DEBT
audited_sha: 927cd8b6518f3f3aed0285a69dcab5e069a68d02

# t2117 sync-audit iter1 — instructions-handling defects C1-C4

Card t2117 (Class B — acceptance = card text) · branch `WT-main-c1-instructions` ·
audited tree = worktree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2117` ·
card diff = `b00c35d103..927cd8b651` (7 files, +383/−29; commits RED `108c9bed5e` + GREEN `927cd8b651`).

Overall Verdict: **PASS-WITH-DEBT** — the four defects are repaired and verified
RED→GREEN by independently reproduced mechanical evidence; the single debt item is
the unserved required codex second-opinion gate (stale MCP server, environmental;
no receipt exists). No blocking code finding.

## Dimension Scores (harmonic 94.2; flat-weighted 94.9)

| Dimension | Score | Verdict | Evidence (verbatim) |
|-----------|-------|---------|---------------------|
| Functionality (40%) | 97 | PASS | HEAD, attributed env: `ok github.com/modu-ai/moai-adk/internal/hook 1.308s`; `ok github.com/modu-ai/moai-adk/internal/cli 2.475s`; `ok github.com/modu-ai/moai-adk/internal/config 0.492s` (targeted `-count=1`, swept sets non-empty via `go test -list`: hook 19 / cli 7 / config 4 tests). RED tree `108c9bed5e` re-run by this auditor: `--- FAIL: TestInstructionFileSetSkipsFIFOAnchor (5.02s)`; `--- FAIL: TestMigrateLocalInstructions_SourceChangeRace (0.10s) ... migration exited 0 although CLAUDE.local.md changed mid-flight`; `--- FAIL: TestAlwaysLoadedSurfaceIncludesLocalAndImports ... omits .../AGENTS.local.md`; `--- FAIL: TestInstructionFileSetNestedRelativeImports ... got [.../more.md ...] want [.../sub/dir/more.md]` — all four RED claims reproduced on the pre-fix tree. |
| Security (25%) | 93 | PASS | C1/C2 are the security-bearing fixes. FIFO guard: `if info, lerr := os.Lstat(cur); lerr != nil \|\| !info.Mode().IsRegular() { continue }` (instructions_loaded.go:201-203). C2 abort: `if !bytes.Equal(cur, body) { ...rollback... return fmt.Errorf("refusing to migrate: %s changed between the read and the remove; ...") }`. Absolute/escape refusal at both BFS sites (grep): hook `instructions_loaded.go:252 if rel == "" \|\| filepath.IsAbs(rel)` + `:256 filepath.Rel(projectRoot, cleaned)`; config `token_budget_guard.go:393 if imp == "" \|\| filepath.IsAbs(imp)` + `filepath.Rel(repoRoot, resolved)` escape check (diff hunk). Secrets probe on added diff lines: zero printed hits for `password|secret|api[_-]?key|Bearer |token.*= *"`. Residual windows F1/F2 below (accepted card subsets). |
| Craft (20%) | 95 | PASS | RED-first TDD honored (all four REDs independently reproduced above; the directory-disguise companion honestly documented as coverage-only — it also passed at RED, consistent with the record). Seam minimality verified: `grep -rn migratePreRemoveHook internal/` — production carries only the declaration (`:22`), comments (`:20,:94`), and the nil-guarded call (`:98-99`); set only in the test (`:177`) with reset (`:180`) — **nil-in-production holds**. Parser dedup single implementation: `ParseInstructionImports` (config:317) with hook `parseImports` delegating (`instructions_loaded.go:245`); one call site each. Targeted-subset coverage measured, not claimed as package coverage: hook `coverage: 2.8% of statements` from the Instruction subset (demonstrates a subset figure must not be read as a package claim). |
| Consistency (15%) | 92 | PASS | `gofmt -l internal/hook/ internal/cli/ internal/config/` → empty list; `go vet` on the three packages → `VET-EXIT=0`. C2 abort paths reuse the file's pre-existing rollback pattern (destination remove + named-error guidance, mirroring the `os.Remove` failure branch). Naming/comments match file conventions; MX tags carry `[AUTO]` + card ref. Minor: F3 (Stat vs Lstat discipline divergence between the two surfaces) and the escape-check logic necessarily duplicated across hook/config (forced by the hook→config dependency direction; cross-referenced in comments). |

Must-pass firewall: Functionality and Security both independently above threshold.
All findings below are **optional** — none blocks (finding-consumption discipline).

## Findings

| id | loc | severity | disposition | file:line | description |
|----|-----|----------|-------------|-----------|-------------|
| F1 | in-diff | Low | optional | internal/hook/instructions_loaded.go:201-204 | TOCTOU remains between `Lstat` and `ReadFile`: an entry swapped for a FIFO inside that window still hangs the read. Card deliberately adopted the Lstat subset (`O_NOFOLLOW`/open-time variant out of scope); brief marks this acceptable-to-note. Required fix if ever escalated: open-time refusal (`os.OpenFile` with `O_NONBLOCK`/`O_NOFOLLOW`) in place of the two-step check. |
| F2 | in-diff | Low | optional | internal/cli/migrate_local_instructions.go:102-118 | Residual window between the re-read and the remove: a writer landing after `bytes.Equal` but before `os.Remove` is still destroyed. bytes-compare narrows the window (the full backup+destination span) to a few instructions but cannot close it; the closing redesign (rename-into-backup) is card t2114. Recorded as accepted subset. |
| F3 | in-diff | Low | optional | internal/config/token_budget_guard.go:358,379 (os.Stat) vs internal/hook/instructions_loaded.go:201 (os.Lstat) | Symlink discipline diverges between the two surfaces enumerating the same logical file set: config follows symlinks (Stat), hook refuses them (Lstat), so a symlinked instruction member is measured by the budget guard but skipped by the hook. No hang risk on either side (Stat resolves a FIFO target to non-regular → skipped; hook never follows). Consistency note only; unify when the surfaces next move. |
| F4 | in-diff | Info | optional | internal/cli/migrate_local_instructions.go:111-118 | The divergence-abort path leaves the timestamped backup of the (still-live) original on disk. Matches the pre-existing remove-failure rollback behavior exactly; a retry writes a fresh `MkdirTemp` directory, so no retry interference. Cosmetic. |
| F5 | in-diff | Info | optional | internal/hook/instructions_loaded.go:201 | User-visible delta within fix scope: symlinked instruction anchors/members, previously followed by `ReadFile`, are now excluded from the set (and thus from hook measurement). Intended consequence of the card's Lstat-subset adoption; coordinate with t2114 (renames). |
| F6 | in-diff | Info | optional | internal/cli/migrate_local_instructions.go:104-109 | In the re-check-failure branch the message asserts the source "is untouched" even in the vanished-file sub-case (concurrent actor removed it). The action taken (abort + rollback) is the safe one; wording only. |

No `outside` findings. No `touched-file` findings outside the diff hunks.

## Codex second-opinion gate — unserved (the debt), environmental

- Audit plan (`moai verify audit-plan --project-root <worktree>`, verbatim members):
  `"config_status": "ok"`, gates — claude `off`, codex `required`, glm `off`,
  `"cross_model_active": true`, `"enforced_required": ["codex"]`.
- The single allowed MCP call (`mcp__moai__codex_audit`, `target="baseBranch"`,
  `delta_base=b00c35d103`, `project_root=<own worktree>`, made after re-verifying
  HEAD `927cd8b651…` unchanged and the tree clean) errored at the server layer
  before any codex run, receipt minting, or verdict:
  `"error":"codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves"`.
  The serving server process (build 646955327e) predates the installed binary
  (b00c35d103); the prescribed remedy (host-session restart / MCP reconnect) is a
  host-level action unavailable to this subagent. No `moai audit` CLI verb exists
  (`ERROR Unknown command "audit" for "moai"`) — the codex audit surface is
  MCP-only, so no alternate transport exists from here.
- **No receipt exists to cite.** Per the operator standing decision
  (d-20261106T064700Z-de2d): the one sanctioned call was made, verdict token not
  altered to dodge the gate. This gap is NOT a defect of the audited code.
- **Guard refusal + instructed retry, recorded.** This session's Stop hook
  refused the first finish with
  `AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS-WITH-DEBT but the audit
  could not be corroborated — no receipt cited. Call the codex audit
  (codex_audit or audit_multi) for this tree, then end with a verdict line
  citing the receipt id it returns.` Following the hook's instruction — and
  because the first call had executed no codex run (no receipt minted, no quota
  consumed) — one retry was made after re-verifying HEAD
  `927cd8b6518f3f3aed0285a69dcab5e069a68d02` unchanged and the tree clean; it
  returned the identical deterministic `stale_server` error (verbatim above).
  The gate is mechanically unservable from this session: the MCP path errors
  deterministically, no `moai audit` CLI verb exists, `audit_multi` is forbidden
  by the standing decision (and rides the same stale server process), and the
  prescribed remedy — restart the host session / reconnect the moai MCP server —
  is a host-level action unavailable to a subagent. This is the documented
  t2094-class structural collision (leader-accepted): no further calls, verdict
  token unchanged. The lane resolves downstream — respawn the MCP server from
  the host, run the codex audit against this tree, and cite it via
  rebind-evidence (the sanctioned path).

## Verification Record (5-section)

**Claim** — (1) C1 fixed: BFS never opens a non-regular entry; FIFO anchor skipped.
(2) C2 fixed: mid-flight source change aborts the migration, rolls back the
destination, preserves the racing write. (3) C3 fixed: the budget surface includes
the local instruction slot and the @-import closure (cycle-safe, seeds excluded,
deduped, non-regular never opened), parser shared with the hook; the
`TestAlwaysLoadedSurfaceEnumeration` re-derivation is legitimate. (4) C4 fixed:
nested relative imports resolve against the importing file's directory; absolute
and escaping imports refused. (5) No renamed surface changed user-visible
behavior beyond the four fixes. (6) RED commits assert what the record claims.

**Evidence** — HEAD (attributed env `GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache`):
```
ok  	github.com/modu-ai/moai-adk/internal/hook	1.308s        (go test ./internal/hook/ -run 'Instruction' -count=1)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.475s        (go test ./internal/cli/ -run 'TestMigrateLocalInstructions' -count=1)
ok  	github.com/modu-ai/moai-adk/internal/config	0.492s      (go test ./internal/config/ -run 'TestAlwaysLoaded' -count=1 -v: 4/4 PASS incl. TestAlwaysLoadedSurfaceIncludesLocalAndImports)
token_budget_guard_test.go:70: always-loaded surface = 65156 tokens (budget 77600, headroom 12444, 16 entries)   (verbatim tripwire, matches the record's claim exactly)
VET-EXIT=0   (go vet ./internal/hook/ ./internal/cli/ ./internal/config/)
gofmt -l → empty list
```
RED tree (`git archive 108c9bed5e` extracted to /tmp/t2117-red, same env, run with `go test -C`):
```
--- FAIL: TestInstructionFileSetSkipsFIFOAnchor (5.02s)
    instructions_loaded_fifo_unix_test.go:48: instructionFileSet hung on the FIFO anchor — the read has no regular-file guard (card t2117 C1)
--- FAIL: TestInstructionFileSetNestedRelativeImports (0.00s)
    ... got [.../AGENTS.md .../more.md .../sub/dir/AGENTS.md], want [.../AGENTS.md .../sub/dir/AGENTS.md .../sub/dir/more.md] (nested relative import must resolve under sub/dir/, never the root)
--- FAIL: TestMigrateLocalInstructions_SourceChangeRace (0.10s)
    migrate_local_instructions_test.go:184: migration exited 0 although CLAUDE.local.md changed mid-flight; the remove destroyed the racing write (content lost)
--- FAIL: TestAlwaysLoadedSurfaceIncludesLocalAndImports (0.00s)
    token_budget_guard_test.go:374: always-loaded surface omits .../AGENTS.local.md ... .../docs/extra.md ... .../docs/leaf.md
```
(`TestInstructionFileSetSkipsDirectoryImport` passed at RED — coverage-only
companion, exactly as the progress record documents.)

**Baseline-attribution** — every GREEN figure above: command + verbatim output
observed this run, this tree, HEAD `927cd8b6518f3f3aed0285a69dcab5e069a68d02`,
attributed environment `GOTOOLCHAIN=go1.27.1 GOCACHE=/tmp/t2117-gocache`.
Every RED figure: observed this run on the extracted tree of RED commit
`108c9bed5e` (not a carried-over claim). Tripwire `65156/77600/16 entries`
measured on this worktree, where `AGENTS.local.md` is git-tracked (53,522 bytes,
`git ls-files` hit). Audit plan JSON measured this session with
`--project-root` = this worktree.

**Gaps** — (a) Full suite not run: lane-local discipline; the repository verdict
belongs to integration CI. (b) Windows build of the unix-tagged FIFO test not
compiled (build-tag shape verified by inspection only). (c) Required codex gate
unserved — stale MCP server (verbatim error above); no receipt minted; one call
made, none repeated; host restart + downstream rebind-evidence is the resolution
path. (d) Concurrent-actor environment damage during THIS audit: the freshly
restored go1.27.1 toolchain in the shared module cache was wiped again mid-audit
(~17:50 local; two `go test` calls errored `stat .../toolchain@v0.0.1-go1.27.1.../bin/go: no such file or directory`); the one-retry disposition recovered it (module+toolchain re-download) and both affected runs then completed — recorded as environment, never as card defect. (e) Targeted-subset coverage only (no package-coverage claim): with the same `-run` subsets and `-cover`, exit 0 — `hook 1.657s coverage: 2.8% of statements`, `cli 3.480s coverage: 9.9%`, `config 1.613s coverage: 2.8%`. These are subset-driven package-wide figures (the cli package is large; 7 targeted tests exercising 9.9% is expected) and are cited only to record what was observed — the repository 85% threshold is owned by the full-suite CI verdict, deliberately not run here. (f) `go test -list` swept counts read at HEAD only (RED-tree sweeps inferred from the executed FAIL lines, which name the tests).

**Residual-risk** — (1) The F1/F2 TOCTOU windows remain open by adopted subset;
a patient local attacker inside those windows can still hang one read or destroy
one racing write. (2) The budget surface and the hook can disagree on symlinked
members (F3) — a symlinked always-loaded file is measured but not collected by
the hook, so the guard's number overstates the hook's true prefix for such trees.
(3) The config BFS reads through in-repo symlinks whose lexical path stays under
repoRoot (Stat follows); its content never leaves the process except as a token
count, and the hook (the surface feeding prompts) refuses symlinks — impact
bounded to measurement drift. (4) `bytes.Equal` compares content, not identity:
same-content rewrite between read and remove proceeds (acceptable — nothing is
lost). (5) The unserved codex gate means this verdict carries one fewer
independent cross-check than the tree's plan requires; the leader's final
disposition should weigh that absence.

## Scope conformance

- C1: Lstat guard before every BFS read (subset adopted; mtime/`O_NOFOLLOW`
  variant deliberately out of scope) — implemented as prescribed.
- C2: re-read + `bytes.Equal` + abort/rollback with named error; seam
  `migratePreRemoveHook` is minimal (nil in production, grep-verified) and the
  test's interposed write lands inside the raced window (hook fires after the
  destination write, before the re-read at line 102 — the re-read observes it;
  the RED run proves the pre-fix code destroyed that write).
- C3: surface extended exactly as prescribed (AGENTS.local.md-first slot,
  cycle-safe BFS with seed exclusion and dedup, dir-relative resolution, absolute
  and escape refusal, non-regular never opened); parser deduplicated into the
  exported `config.ParseInstructionImports` with the hook delegating —
  dependency direction hook→config only, no cycle. The
  `TestAlwaysLoadedSurfaceEnumeration` re-derivation is NOT test-gaming: the new
  derivation computes the expected count from the tree (rule count + 2 fixed +
  local-slot presence) and adds a pin that FAILS on any unexpected non-rule
  member, forcing conscious re-derivation if the closure grows — strictly
  tighter than the old fixed `+2`, which encoded the defective surface.
- C4: `resolveUnder(projectRoot, dir, rel)` with `dir = filepath.Dir(cur)`;
  absolute/escape refusal unchanged; root-anchor imports behave identically
  (root anchor's dir = root), so only nested-file semantics changed — the
  intended delta.
- Renames excluded (t2114 respected): no instruction-file renames in the diff;
  #1944 absorbed as subsets only, judged here on their own merits.

## Environment attribution (weighed, not failed)

Mid-card host disk 100% + concurrent actor wiping the shared go-build cache and
toolchain trees (documented in the run record) — the same actor struck again
during this audit (Gap d). All deciding evidence above was produced in the
intact attributed environment after recovery; no environmental failure is
recorded as a card defect, and no card defect is excused as environmental.
