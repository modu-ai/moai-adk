auditor-model: glm-5.3
verdict: FAIL
audited_sha: 927cd8b6518f3f3aed0285a69dcab5e069a68d02

# t2117 sync-audit iter1 — instructions-handling defects C1-C4

Card t2117 (Class B — acceptance = card text) · branch `WT-main-c1-instructions` ·
audited tree = worktree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2117` ·
card diff = `b00c35d103..927cd8b651` (7 files, +383/−29; commits RED `108c9bed5e` + GREEN `927cd8b651`).

Overall Verdict: **FAIL** (must-pass firewall: Functionality — one stated
acceptance criterion demonstrably fails). The four dispatched defects remain
repaired and verified RED→GREEN by independently reproduced mechanical
evidence; the FAIL is driven by ONE blocking in-diff finding (F3+): the card's
C3 acceptance clause states "**absolute/escaping refused**", and the new
`instructionImportMembers` refuses only LEXICAL escaping — an import path that
is a symlink to a regular file outside repoRoot passes `os.Stat` (follows
links) and the string-form `filepath.Rel` check and is then read and counted,
so an escaping target IS followed. The code's own diff comment ("targets
escaping repoRoot are not followed") is false for that reachable case. F7
(coexistence under-count) rides as Medium/optional — the card text literally
prescribed first/fallback — and may fold into the same repair.

**Verdict history, recorded transparently.** Pre-receipt, this audit's token
was PASS-WITH-DEBT with the unserved codex gate as the (environmental) debt.
The receipt (`rcpt-570cb4d72e7e4f0aaab856cb`, fifth call after four
`stale_server` refusals) records codex's independent verdict on the delta:
`fail`, two P2 findings. The Stop-hook receipt guard refused the pass-family
token on exactly that ground ("receipt records a FAIL verdict"), which forced
a re-examination of my classification of those findings against the card's own
acceptance text. The re-examination — not the guard — moves the token: F3+ is
a failure of a stated criterion ("escaping refused"), not a beyond-prescription
residual as first classified; under the must-pass firewall that forces FAIL.
Codex's `fail` and this verdict now agree.

## Dimension Scores (harmonic 85.3 — OVERRIDDEN: must-pass firewall forces FAIL on Functionality's stated-criterion failure)

| Dimension | Score | Verdict | Evidence (verbatim) |
|-----------|-------|---------|---------------------|
| Functionality (40%) | 75 | **FAIL** (stated criterion) | The four dispatched defects remain GREEN (evidence below, unchanged). The FAIL is the C3 acceptance clause "**absolute/escaping refused**": demonstrated bypass — an import path that is a symlink to a regular file outside repoRoot passes `os.Stat` (follows links) and the lexical `filepath.Rel(repoRoot, resolved)` check, then `os.ReadFile` reads and counts the external file (mechanism verified by this auditor against the code at token_budget_guard.go:379-401; codex P2 #2 concurrence, receipt rcpt-570cb4d72e7e4f0aaab856cb). The diff's own comment "targets escaping repoRoot are not followed" is false for this reachable case. HEAD, attributed env: `ok github.com/modu-ai/moai-adk/internal/hook 1.308s`; `ok github.com/modu-ai/moai-adk/internal/cli 2.475s`; `ok github.com/modu-ai/moai-adk/internal/config 0.492s` (targeted `-count=1`, swept sets non-empty via `go test -list`: hook 19 / cli 7 / config 4 tests). RED tree `108c9bed5e` re-run by this auditor: `--- FAIL: TestInstructionFileSetSkipsFIFOAnchor (5.02s)`; `--- FAIL: TestMigrateLocalInstructions_SourceChangeRace (0.10s) ... migration exited 0 although CLAUDE.local.md changed mid-flight`; `--- FAIL: TestAlwaysLoadedSurfaceIncludesLocalAndImports ... omits .../AGENTS.local.md`; `--- FAIL: TestInstructionFileSetNestedRelativeImports ... got [.../more.md ...] want [.../sub/dir/more.md]` — all four RED claims reproduced on the pre-fix tree. |
| Security (25%) | 82 | FAIL (confinement clause) | C1/C2 fixes sound (evidence unchanged): FIFO guard `if info, lerr := os.Lstat(cur); lerr != nil \|\| !info.Mode().IsRegular() { continue }` (instructions_loaded.go:201-203); C2 abort `if !bytes.Equal(cur, body) { ...rollback... "refusing to migrate: %s changed between the read and the remove; ..." }`. Secrets probe on added diff lines: zero printed hits. The dimension FAILs on the stated project-internal confinement being bypassable via symlink in the new config BFS (F3+): the guard's measurement can be pointed at arbitrary readable files, and its reproducibility claim breaks. Hook prompt surface unaffected (Lstat refuses symlinks). Residual windows F1/F2 remain accepted card subsets. |
| Craft (20%) | 95 | PASS | RED-first TDD honored (all four REDs independently reproduced above; the directory-disguise companion honestly documented as coverage-only — it also passed at RED, consistent with the record). Seam minimality verified: `grep -rn migratePreRemoveHook internal/` — production carries only the declaration (`:22`), comments (`:20,:94`), and the nil-guarded call (`:98-99`); set only in the test (`:177`) with reset (`:180`) — **nil-in-production holds**. Parser dedup single implementation: `ParseInstructionImports` (config:317) with hook `parseImports` delegating (`instructions_loaded.go:245`); one call site each. Targeted-subset coverage measured, not claimed as package coverage: hook `coverage: 2.8% of statements` from the Instruction subset. |
| Consistency (15%) | 92 | PASS | `gofmt -l internal/hook/ internal/cli/ internal/config/` → empty list; `go vet` on the three packages → `VET-EXIT=0`. C2 abort paths reuse the file's pre-existing rollback pattern (destination remove + named-error guidance, mirroring the `os.Remove` failure branch). Naming/comments match file conventions; MX tags carry `[AUTO]` + card ref. Minor: Stat-vs-Lstat discipline divergence (F3) and the escape-check logic duplicated across hook/config (forced by the hook→config dependency direction; cross-referenced in comments). |

**Must-pass firewall: Functionality fails a stated acceptance criterion ("escaping
refused") → overall FAIL regardless of the other dimension scores.** F3+ below is
the single **blocking** finding; the repair re-audit (second call of the card's
cap-2 codex allowance, unspent) is scoped to the enumerated defect delta.

## Findings

| id | loc | severity | disposition | file:line | description |
|----|-----|----------|-------------|-----------|-------------|
| F1 | in-diff | Low | optional | internal/hook/instructions_loaded.go:201-204 | TOCTOU remains between `Lstat` and `ReadFile`: an entry swapped for a FIFO inside that window still hangs the read. Card deliberately adopted the Lstat subset (`O_NOFOLLOW`/open-time variant out of scope); brief marks this acceptable-to-note. Required fix if ever escalated: open-time refusal (`os.OpenFile` with `O_NONBLOCK`/`O_NOFOLLOW`) in place of the two-step check. |
| F2 | in-diff | Low | optional | internal/cli/migrate_local_instructions.go:102-118 | Residual window between the re-read and the remove: a writer landing after `bytes.Equal` but before `os.Remove` is still destroyed. bytes-compare narrows the window (the full backup+destination span) to a few instructions but cannot close it; the closing redesign (rename-into-backup) is card t2114. Recorded as accepted subset. |
| F3 | in-diff | Low | optional | internal/config/token_budget_guard.go:358,379 (os.Stat) vs internal/hook/instructions_loaded.go:201 (os.Lstat) | Symlink discipline diverges between the two surfaces enumerating the same logical file set: config follows symlinks (Stat), hook refuses them (Lstat), so a symlinked instruction member is measured by the budget guard but skipped by the hook. No hang risk on either side (Stat resolves a FIFO target to non-regular → skipped; hook never follows). Consistency note only; unify when the surfaces next move. |
| F4 | in-diff | Info | optional | internal/cli/migrate_local_instructions.go:111-118 | The divergence-abort path leaves the timestamped backup of the (still-live) original on disk. Matches the pre-existing remove-failure rollback behavior exactly; a retry writes a fresh `MkdirTemp` directory, so no retry interference. Cosmetic. |
| F5 | in-diff | Info | optional | internal/hook/instructions_loaded.go:201 | User-visible delta within fix scope: symlinked instruction anchors/members, previously followed by `ReadFile`, are now excluded from the set (and thus from hook measurement). Intended consequence of the card's Lstat-subset adoption; coordinate with t2114 (renames). |
| F6 | in-diff | Info | optional | internal/cli/migrate_local_instructions.go:104-109 | In the re-check-failure branch the message asserts the source "is untouched" even in the vanished-file sub-case (concurrent actor removed it). The action taken (abort + rollback) is the safe one; wording only. |
| F7 | in-diff | Medium | optional (this card) → follow-up routed | internal/config/token_budget_guard.go:338 (`localInstructionSlot`) | **Codex P2 #1** (receipt rcpt-570cb4d72e7e4f0aaab856cb): when `AGENTS.local.md` and `CLAUDE.local.md` coexist, the slot returns only the first, so the legacy file's body AND its `@`-import closure fall out of the budget surface — while the real injection combines both bodies (grounded: AGENTS.md §8 local-launch form, "reads … `AGENTS.local.md` first, then the legacy file, and injects the combined bodies"; the primary checkout carries both files today). Under-count can mask an overshoot during the coexistence state. The card text itself prescribed first/fallback semantics, so the implementation matches the dispatch — the gap is prescription-level. Required fix (follow-up): enumerate BOTH slots as seeds when both are present as regular files. |
| F3+ | in-diff | Medium | **BLOCKING** | internal/config/token_budget_guard.go:381 (`instructionImportMembers`, resolution at :396-401) | **Codex P2 #2, re-classified blocking on re-examination**: the card's C3 acceptance clause states "**absolute/escaping refused**", but the refusal is lexical only — an import path that is a symlink to a regular file outside repoRoot passes `os.Stat` (follows links) and the `filepath.Rel(repoRoot, resolved)` string check, and the external file is then read and counted. The stated project-internal confinement is bypassable; the diff's own comment ("targets escaping repoRoot are not followed") is false for this reachable case. **Required fix**: resolve the real path (`filepath.EvalSymlinks`) BEFORE the boundary check and reject targets (and symlinked ancestors) whose resolved path escapes repoRoot; add a RED test with an escaping symlink import. Scope note: the hook's prompt-facing BFS is NOT affected (its Lstat refuses symlinks); the pre-card hook shared the same lexical discipline, so this is a new-code acceptance failure, not a regression. |

No `outside` findings. No `touched-file` findings outside the diff hunks.

## Codex second-opinion gate — SERVED on the fifth call (receipt minted); four environmental refusals preceded

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
- **Third refusal after host reconnect, recorded (final).** The operator
  reconnected the lane's `/mcp`; HEAD was re-verified first — it had moved to
  `810f0897a841fcf80acc0f3e3ee934ebb268d906`, a docs-only commit adding exactly
  `.moai/reports/t2117/progress.md` (+78) and this verdict file (+178), zero
  code-surface change (`git diff --stat 927cd8b651..810f0897a8`). The instructed
  single retry (same parameters: `target="baseBranch"`,
  `delta_base=b00c35d103`, `project_root=<own toplevel>`) was refused again with
  the identical deterministic error:
  `"error":"codex_audit: the serving moai build is stale — stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from the installed build (/Users/goos/go/bin/moai) — reconnect the moai MCP server (restart the host session) so the installed build serves"`.
  Serving-build notation, complete: the responding server process self-identifies
  as `_server: {"commit":"b00c35d103","version":"v3.2.1-rc.1"}` — i.e. the
  post-reconnect respawn IS fresh relative to rc.1 — while the installed binary
  read this same run reports `moai-adk v3.2.1-rc.2` (`moai version`): the rc
  rollout replaced the on-disk binary again after the reconnect, so the
  staleness check (serving rc.1 vs on-disk rc.2) fires against a moving target.
  Per the coordinator's instruction: no further retries; refusal reported
  verbatim; verdict token unchanged. Resolution remains downstream: respawn the
  server from a host whose installed binary has stopped moving, run the codex
  audit against this tree, cite via rebind-evidence.
- **Stop-hook loop, closed.** The Stop hook refused this session's finish twice
  on the same missing receipt (`AUDIT_RECEIPT_VIOLATION ... no receipt cited`),
  before and after the third recorded refusal. The loop is deterministic — the
  only receipt-minting path errors `stale_server` on every call, further retries
  are forbidden by the coordinator's instruction, and altering the verdict token
  to satisfy the guard is forbidden by the standing decision. Disposition per
  that decision (t2094 precedent, leader-accepted): recorded, token unchanged,
  finished. The lane resolves the receipt downstream (rebind-evidence).
- **Fourth refusal (RECOMMEND-path retry, final).** After the operator's second
  `/mcp` reconnect round, HEAD re-verified
  `5158fc0981ccec574e6d096be418cd544793aaac` (docs-only, two commits past the
  audited SHA, zero code change); the single instructed retry (same parameters)
  returned the verbatim-identical refusal:
  `stale_server: the serving moai build (/Users/goos/go/bin/moai) differs from
  the installed build (/Users/goos/go/bin/moai)` with `_server
  {"commit":"b00c35d103","version":"v3.2.1-rc.1"}` still serving against the
  installed `v3.2.1-rc.2`. No receipt minted, nothing bound to any SHA. No
  further calls per instruction; verdict token unchanged. Four calls total,
  zero codex runs.
- **Fifth call — GATE SERVED, receipt minted (final).** After the stale rc.1
  server was killed and replaced by a fresh process (pid 21882, build
  v3.2.1-rc.2, commit 04d8c27a7e — coordinator-verified), HEAD re-verified
  `bf35a3495312b0c050bed702e99ec182f77340a7` (docs-only commits past the
  audited SHA; `git diff --stat` over them shows only `.moai/reports/`
  progress/verdict files — code surface identical to `927cd8b651…`), the call
  (same parameters) succeeded: **receipt `rcpt-570cb4d72e7e4f0aaab856cb`**,
  server provenance `"backend":"codex","resolved_model":"gpt-6.1-sol"`,
  `"build_commit":"04d8c27a7e"`, review scope
  `"review_base":"delta_base (merge base b00c35d103fa175efe3b68340d9b176cba1b067b)"`.
  The result carries no explicit head-SHA field; the reviewed delta is
  merge-base `b00c35d103…` to call-time HEAD `bf35a3495312…`. Codex's
  independent verdict on the delta: **fail**, two P2 findings — folded below
  as F7 (new) and F3 concurrence. The prior debt item (unserved gate) is
  resolved.
- **Receipt-guard corroboration and token revision (final).** The Stop-hook
  receipt guard refused the pass-family token on a NEW ground —
  `AUDIT_RECEIPT_VIOLATION: ... receipt records a FAIL verdict` — i.e. the
  receipt was corroborated as valid but its recorded verdict is fail. That
  refusal forced a re-examination of the findings' classification against the
  card's own acceptance text; the re-examination (not the guard) re-classified
  F3+ as a failure of the stated criterion "escaping refused" and revised the
  verdict token to **FAIL** under the must-pass firewall. Codex's recorded
  verdict and this verdict now agree; the receipt corroborates. The card's
  remaining codex allowance (one repair re-audit of the cap-2) is unspent and
  is the confirming re-audit for the F3+ defect delta.

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
compiled (build-tag shape verified by inspection only). (c) Required codex gate — initially unserved (four `stale_server` refusals,
verbatim above), ultimately SERVED on the fifth call after the stale server
process was replaced: receipt `rcpt-570cb4d72e7e4f0aaab856cb` minted by build
04d8c27a7e against merge-base `b00c35d103…`, call-time HEAD
`bf35a3495312…` (docs-only past the audited SHA). The gate's findings drove the
final verdict: F3+ re-classified blocking (stated criterion "escaping refused"
fails), F7 Medium/optional. The repair re-audit allowance (cap-2, second call)
remains unspent for the confirming delta re-audit. (d) Concurrent-actor environment damage during THIS audit: the freshly
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
