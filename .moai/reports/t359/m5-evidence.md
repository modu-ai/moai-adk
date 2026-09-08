# M5 evidence — SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

Compatibility and doctrine: the landing evidence survives the JSON⇄SQLite round trip on BOTH
card-bearing tables and is compared by the migration's own parity check (AC-TLE-017); a pre-change
binary still serves a post-change database (AC-TLE-018); the two doctrine surfaces agree with each
other and with the number of fields the render emits (AC-TLE-021).

Verbatim command + output pairs. Exit codes are captured with `; echo "rc=$?"` on the un-piped
command, never read off a filtered window — M4 recorded one cycle where a `--- FAIL:` sat below a
`head -20` cut and the run was read as green.

---

## 1. Working tree and baseline, read at M5 entry

```
$ git rev-parse HEAD && git rev-parse --show-toplevel && git branch --show-current && git status --short && echo "---rc=$?"
e4cb86920f2b31c817e7cf66cfd1417fe1ae760d
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359
WT-landing-evidence
---rc=0
```

HEAD re-read by this agent as its first command; clean tree. This is the SHA M5 records — the
dispatch's `e4cb86920` was a comparison value, and it agreed.

## 2. Pre-edit baseline (attributable, this tree, this run)

```
$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.542s
kanban rc=0

$ go test ./internal/cli/... -count=1 -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	5.852s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	7.956s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.597s
cli rc=0
```

Both green before any edit, measured in this tree at this HEAD rather than carried over from M4.
`internal/template` was NOT baselined here — see §9 known losses, item 1, and §8 for what that
costs the attribution of its two failures.

---

## 3. AC-TLE-017 — RED, written first

Three assertions, all authored before any production edit.

```
$ go test ./internal/kanban/ -run 'TestBacklogLanding_(RoundTripPreservesEvidence|ParityComparesEvidence|SurvivesDoneAndUndone)' -count=1 -v
=== RUN   TestBacklogLanding_RoundTripPreservesEvidence
    backlog_landing_roundtrip_test.go:96: archived card t2: evidence LOST — want {Ref:origin/develop RefHead:3333333333333333333333333333333333333333 ObservedAt:2026-01-02T00:00:00Z SHA: SHASource: SpecStatus:}, got none
--- FAIL: TestBacklogLanding_RoundTripPreservesEvidence (0.01s)
=== RUN   TestBacklogLanding_SurvivesDoneAndUndone
    backlog_landing_roundtrip_test.go:174: archived t1: evidence LOST — want {...}, got none
    backlog_landing_roundtrip_test.go:189: restored t1: evidence LOST — want {...}, got none
--- FAIL: TestBacklogLanding_SurvivesDoneAndUndone (0.01s)
=== RUN   TestBacklogLanding_ParityComparesEvidence/archived_evidence_dropped
    backlog_landing_roundtrip_test.go:129: parity check passed a archived evidence dropped — the cutover would have flipped authority onto a record missing the operator's evidence
=== RUN   TestBacklogLanding_ParityComparesEvidence/archived_evidence_altered
    backlog_landing_roundtrip_test.go:129: parity check passed a archived evidence altered — ...
=== RUN   TestBacklogLanding_ParityComparesEvidence/live_evidence_dropped
    backlog_landing_roundtrip_test.go:129: parity check passed a live evidence dropped — ...
=== RUN   TestBacklogLanding_ParityComparesEvidence/live_evidence_altered
    backlog_landing_roundtrip_test.go:129: parity check passed a live evidence altered — ...
--- FAIL: TestBacklogLanding_ParityComparesEvidence (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.477s
```

**What the RED says, precisely.** The LIVE card's evidence already survived — M3 plumbed the
`items` read/write and said so. What was lost was the ARCHIVED card's, at
`backlog_landing_roundtrip_test.go:96`. And all four parity subtests failed the other way round:
the check PASSED records it should have rejected, on both tables. That is the vacuity AC-TLE-017
exists for — a parity check that reports success while dropping the column.

**The `landed → done → undone` assertion M3 left for M5.** M3 measured the loss with a `t.Logf`
probe and deleted it, because a probe that passes either way is a vacuous green. The replacement
(`TestBacklogLanding_SurvivesDoneAndUndone`) fails at `:174` and `:189` above and passes after the
fix — the discrimination M3's probe did not have.

## 4. AC-TLE-017 — GREEN

Three sites, all in `internal/kanban/backlog_migrate.go`: the archived SELECT gains `landing` and
decodes it, `writeArchive` binds it through the same `LandingEvidenceValue` seam as the live write,
and `assertBacklogParity` compares it on both tables via `equalLandingEvidence` (null-shape AND
value, the two-part test `equalSpecID` already applies).

```
$ go test ./internal/kanban/ -run 'TestBacklogLanding_(RoundTripPreservesEvidence|ParityComparesEvidence|SurvivesDoneAndUndone)' -count=1 -v ; echo "GREEN rc=$?"
GREEN rc=0
--- PASS: TestBacklogLanding_RoundTripPreservesEvidence (0.01s)
--- PASS: TestBacklogLanding_SurvivesDoneAndUndone (0.01s)
--- PASS: TestBacklogLanding_ParityComparesEvidence (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.478s
```

## 5. AC-TLE-017 — the mandated mutant, twice

### 5.1 Hash before

```
$ shasum internal/kanban/backlog_migrate.go
843bee33cd0416d72c38ebd38a61b126d4f444bb  internal/kanban/backlog_migrate.go
```

### 5.2 Mutant A — the migration drops the landing value on BOTH writes

`landing = nil` inserted after each `LandingEvidenceValue` call (lines 234 and 322).

```
$ shasum internal/kanban/backlog_migrate.go
94b406951c231e51a8c091c75cfd76719205eda1  internal/kanban/backlog_migrate.go

$ go test ./internal/kanban/ -run 'TestBacklogLanding_RoundTripPreservesEvidence' -count=1 -v ; echo "MUTANT rc=$?"
MUTANT rc=1
=== RUN   TestBacklogLanding_RoundTripPreservesEvidence
    backlog_landing_roundtrip_test.go:88: Load (migration + parity) failed: migrate backlog /var/folders/.../backlog.json -> /var/folders/.../backlog.db: parity check failed, legacy file left authoritative: item 0 (t1): landing {origin/main 1111111111111111111111111111111111111111 2026-01-01T00:00:00Z 2222222222222222222222222222222222222222 operator completed} != <nil>
--- FAIL: TestBacklogLanding_RoundTripPreservesEvidence (0.00s)
```

**Which assertion fired, and why it is the one the criterion names.** Not an equality assertion in
the test — the failure text is `parity check failed, legacy file left authoritative`, raised inside
`migrateLegacyBacklog` by `assertBacklogParity`. AC-TLE-017 requires the PARITY VERIFICATION to
fail, not merely some test, because the risk is a cutover that flips authority onto a database
missing the operator's record. It refused the cutover and left the legacy file authoritative.

### 5.3 Mutant B — archived-only drop

Mutant A short-circuits on item 0, so it proves the LIVE half only. Planted alone on the archived
write to reach the archived half of the same check.

```
$ shasum internal/kanban/backlog_migrate.go
a95d7568b30824a1542ffa0ded39e29c8780ef7a  internal/kanban/backlog_migrate.go

$ go test ./internal/kanban/ -run 'TestBacklogLanding_RoundTripPreservesEvidence' -count=1 ; echo "MUTANT_B rc=$?"
MUTANT_B rc=1
    backlog_landing_roundtrip_test.go:88: Load (migration + parity) failed: ... parity check failed, legacy file left authoritative: archived 0 (t2): landing {origin/develop 3333333333333333333333333333333333333333 2026-01-02T00:00:00Z   } != <nil>
```

### 5.4 What these assertions would still miss

- **A malformed-but-present stored value.** `readRecord` / `readArchive` surface an undecodable
  value as a read error before parity runs, so a mutation that stored garbage rather than NULL
  reds at the read, not at the parity check. The criterion's hazard (silent drop) is covered; a
  corrupting mutation is a different shape and is not asserted here.
- **A reordering that preserves every value.** Parity compares index-by-index, so a migration that
  swapped two cards' evidence between two cards whose OTHER fields also swapped would pass. The
  existing per-axis ordering assertions cover the id/text axes; nothing pins the pairing itself.
- **Anything outside the two card-bearing tables.** The evidence column exists only there, so this
  is a boundary rather than a gap — stated so it is not mistaken for coverage.

### 5.5 Revert, and the proof no mutant survived

```
$ shasum internal/kanban/backlog_migrate.go
843bee33cd0416d72c38ebd38a61b126d4f444bb  internal/kanban/backlog_migrate.go

$ grep -c "MUTANT" internal/kanban/backlog_migrate.go ; echo "rc=$?"
0
rc=1
```

Byte-identical to §5.1, and no mutant marker survives.

---

## 6. AC-TLE-018 — the downgrade proof

New file `internal/kanban/backlog_downgrade_test.go`. Three clauses, one post-change database
built through the LIVE open path with landing evidence recorded on its card — so the pre-change
statements run against a row that actually carries a value in the new column.

The frozen DDL replica was **extracted programmatically from the pre-change commit**, never
retyped:

```
$ git show e50964ad3:internal/kanban/backlog_sqlite.go | sed -n '/^const backlogDDL = `$/,/^`$/p' ...
```

and the frozen constants (`frozenSchemaVersion = "1"`, `frozenMetaKeySchemaVer = "schema_version"`)
are LITERALS rather than references to the live consts — a reference would track a bump instead of
refusing it, which is the failure clause (b) exists to detect.

**Clause (c)'s withdrawn conjunct is deliberately NOT reconstructed.** The frozen-switch
accepted-version-set comparison was withdrawn as unrealisable at v0.3.1 (progress.md §E.1, F2);
the DDL byte-identity conjunct, which carries the independent RED, is what this file asserts.

```
$ go test ./internal/kanban/ -run 'TestBacklogDowngrade' -count=1 -v ; echo "rc=$?"
rc=0
--- PASS: TestBacklogDowngrade_PreChangeBinaryStillServes (0.01s)
    --- PASS: .../(a)_pre-change_statements_run_verbatim (0.00s)
    --- PASS: .../(b)_the_reconstructed_pre-change_open_path_accepts_the_database (0.00s)
    --- PASS: .../(c)_the_frozen_replica_has_not_drifted_from_the_live_source (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.400s
```

### 6.1 RED — the drift clause (c) exists to make loud

Hash before: `28def6da929998fc328f2c4b06344cd8f9e0327a  internal/kanban/backlog_sqlite.go`.

Live `backlogDDL` gains an index; the frozen copy is left untouched.

```
$ shasum internal/kanban/backlog_sqlite.go
148897d2e5ba25e169d7d6c171fc4e969249f692  internal/kanban/backlog_sqlite.go

$ go test ./internal/kanban/ -run 'TestBacklogDowngrade' -count=1 -v ; echo "RED_C rc=$?"
RED_C rc=1
    --- PASS: .../(a)_pre-change_statements_run_verbatim (0.00s)
    --- PASS: .../(b)_the_reconstructed_pre-change_open_path_accepts_the_database (0.00s)
    --- FAIL: .../(c)_the_frozen_replica_has_not_drifted_from_the_live_source (0.00s)
```

Exactly the stated shape: (a) and (b) stay green, only (c) fails. That is (b)'s independent failure
mode — without (c) the replica would go stale silently while the test kept reporting that it
exercises the pre-change open path.

### 6.2 RED — the schema_version bump

`backlogSchemaVersion` "1" → "2". Hash `4007d567c9c17616b421fff3b715fb3732fd62f8`.

```
$ go test ./internal/kanban/ -run 'TestBacklogDowngrade' -count=1 -v ; echo "RED_V rc=$?"
RED_V rc=1
    --- FAIL: .../(a)_pre-change_statements_run_verbatim (0.00s)
    --- FAIL: .../(b)_the_reconstructed_pre-change_open_path_accepts_the_database (0.00s)
    --- PASS: .../(c)_the_frozen_replica_has_not_drifted_from_the_live_source (0.00s)
    backlog_downgrade_test.go:143: schema_version = "2", want "1" — a bump breaks every older binary
    backlog_downgrade_test.go:149: reconstructed pre-change open failed: schema /var/folders/.../backlog.db: unsupported schema_version "2" (want "1"): kanban backlog store corrupt
```

Both (a) and (b) red from ONE mutation, exactly as the criterion states — they are not independent
detectors of it — and (c) correctly stays green because no DDL text moved.

### 6.3 FINDING — the third stated RED does not fire where the criterion says it does

AC-TLE-018's RED list includes: *"Give `landing` a `NOT NULL` without a default → the verbatim
INSERT in (a) fails."* Planted (`ALTER TABLE %s ADD COLUMN %s TEXT NOT NULL`, hash
`2b9580bc9e497225d4c4c4cfbc142c50b435db3b`), the mutant reds — but **not at the named assertion**:

```
$ go test ./internal/kanban/ -run 'TestBacklogDowngrade' -count=1 -v ; echo "RED_NN rc=$?"
RED_NN rc=1
--- FAIL: TestBacklogDowngrade_PreChangeBinaryStillServes (0.00s)
    backlog_downgrade_test.go:97: add: write backlog /var/folders/.../backlog.db: item t1: kanban backlog id conflict: constraint failed: NOT NULL constraint failed: items.landing (1299)
```

Line 97 is the FIXTURE (`store.Add`), not clause (a)'s verbatim INSERT. Cause: SQLite accepts the
`ALTER … ADD COLUMN … NOT NULL` on the empty table, and then the first LIVE write — which binds
NULL through `LandingEvidenceValue(nil)` for a card with no record — violates the constraint. Under
this mutation the post-change database cannot be built at all, so clause (a)'s INSERT is
**unreachable rather than failing**.

Recorded as a narrower remaining gap, following M2's precedent. The test was NOT restructured to
route the mutation through: doing so would convert the best available finding into a green line.
What is demonstrated is that the mutation is caught; what is NOT demonstrated is the criterion's
stated mechanism for catching it.

### 6.4 Revert

```
$ shasum internal/kanban/backlog_sqlite.go
28def6da929998fc328f2c4b06344cd8f9e0327a  internal/kanban/backlog_sqlite.go

$ git diff --stat -- internal/kanban/backlog_sqlite.go ; echo "rc=$?"
rc=0
```

Byte-identical to the pre-plant hash, and no diff against HEAD at all — the file was never part of
M5's own change.

---

## 7. AC-TLE-021 — doctrine, mirror parity, and the stated column count

### 7.1 The doctrine delta, applied by hand to BOTH surfaces

Three edits, applied to `.claude/skills/moai/workflows/todo.md` and to
`internal/template/templates/.claude/skills/moai/workflows/todo.md` by the same scripted
substitution — never by `cp` from the local copy, per the standing rule that an existing template
file may be a deliberately neutralized variant.

1. The `moai todo pr` row: six columns → **seven**, naming the evidence column in position 6 with
   the card text still LAST, and stating what the `(operator)` / `(ref-head)` marker distinguishes
   and that `--json` carries the record under a `landing` key.
2. A new `moai todo landed` verb row.
3. A `[HARD]` paragraph placing the verb under the operator-act rule: what it writes is EVIDENCE,
   not a transition — recording a landing does not close, move, or mark a card done.

The two files happened to be byte-identical before the edit and remain so after:

```
$ diff .claude/skills/moai/workflows/todo.md internal/template/templates/.claude/skills/moai/workflows/todo.md ; echo "mirror_parity_rc=$?"
mirror_parity_rc=0
```

Byte-parity is not itself the criterion — AC-TLE-021 asserts the two extracted ROWS are identical,
which is the weaker and correct requirement; the template variant is permitted to diverge
elsewhere.

### 7.2 Template neutrality

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -run TestTemplateNoInternalContentLeak -count=1 ; echo "leak rc=$?"
leak rc=0
ok  	github.com/modu-ai/moai-adk/internal/template	1.113s

$ grep -nE 'SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b|/Users/|CLAUDE\.local' internal/template/templates/.claude/skills/moai/workflows/todo.md ; echo "rc=$?"
rc=1
```

No hits — no SPEC ID, REQ token, internal date, commit SHA, macOS-biased path, or CLAUDE.local
reference in the mirror. The guard catches only those classes, so the manual read matters too: the
added prose names no tier, mode number, or plan-section token, and describes the verb in terms a
user of any of the 16 supported programming languages can act on.

### 7.3 GREEN

```
$ go test ./internal/cli/ -run 'TestTodoDoctrine_MirrorParityAndStatedColumnCount' -count=1 -v ; echo "rc=$?"
rc=0
--- PASS: TestTodoDoctrine_MirrorParityAndStatedColumnCount (0.27s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.256s
```

The count half MEASURES rather than cites: it renders one row against its own fixture queue and
counts tab-separated fields, exactly as a consumer piping through `cut` would. The literal 7 is
recorded in the criterion as a note and appears nowhere as a comparand.

### 7.4 Both stated REDs, observed

Hashes before: both surfaces `c5490d55b7f5064d9122b2657cc50a603429f6f2`.

**RED-1 — mirror drift.** A word changed in the local `landed` row only (local
`afde6554ad499801f389b7c15f9602882c9e81ab`, mirror unchanged):

```
RED21A rc=1
    todo_landed_doc_test.go:67: the "| `moai todo landed" row differs between the live document and its template mirror.
```

**RED-2 — prose disagrees with the render.** "seven" → "six" on BOTH surfaces (both
`e161bb502b26ebb24078c66b1c2ab07ddc038849`), so parity stays green and only the count half fires:

```
RED21B rc=1
    todo_landed_doc_test.go:77: live todo.md states 6 columns; `moai todo pr` renders 7
    todo_landed_doc_test.go:77: template mirror todo.md states 6 columns; `moai todo pr` renders 7
```

**Revert:** both surfaces back to `c5490d55b7f5064d9122b2657cc50a603429f6f2`, `diff` rc=0.

### 7.5 What this assertion would still miss

- **Prose outside the two extracted rows.** Only the `landed` row and the `todo pr` row are
  compared; the rest of either document may drift between the surfaces undetected. That is
  deliberate (the template variant is allowed to diverge) but it means the `[HARD]` operator-act
  paragraph this milestone added is NOT mirror-guarded.
- **A wrong column ORDER.** The count agrees at 7 whether or not the evidence sits in field 6.
  Position is pinned by AC-TLE-015's own assertions, not by this one.
- **A number word outside the table** (`numberWords` covers three..ten). A row rewritten as
  "carries a dozen" fails loudly rather than silently, which is the right direction, but the
  vocabulary is finite.

---

## 8. Build, and an INHERITED failure that is not M5's

```
$ make build ; echo "rc=$?"
rc=2
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)
    golden_test.go:109: .codex/agents/moai/sync-auditor.toml: committed artifact differs from emission (sha256 mismatch) — regenerate or stop hand-editing
agent-emit drift: committed .codex/agents/moai/*.toml differ from the .md source layer — run `make agents-emit`
make: *** [agents-emit-check] Error 1
```

**Attribution.** `make build` runs `agents-emit-check` FIRST, and it fails on `sync-auditor` — an
agent file M5 does not touch:

```
$ git status --short -- .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md internal/template/templates/.codex/agents/moai/sync-auditor.toml
(no output — none modified)

$ git log --oneline -1 -- internal/template/templates/.claude/agents/moai/sync-auditor.md
4244c4a06 docs(t386/t387): sync-auditor export-mandate clause — unblocked after lane-9 t302 settled
```

`4244c4a06` edited the `.md` source layer without the mandatory `make agents-emit` regeneration.
The same stale-emission defect shows on a second axis:

```
--- FAIL: TestManifestHashFormat (0.10s)
    catalog_tier_audit_test.go:451: CATALOG_HASH_UNSTABLE: sync-auditor stored hash=f1b4487f..., computed hash=545d03d9... (source=.claude/agents/moai/sync-auditor.md)
```

**This is a blocker report, not a fix.** Regenerating would fold another card's un-emitted artifact
into M5's commit, outside this SPEC's scope envelope. The corrective is `make agents-emit` plus a
catalog-hash regen, owned by whoever closes `4244c4a06`'s omission.

**The build body was run directly, minus the unrelated pre-check**, so M5's own template change is
verified as embedding:

```
$ go run ./internal/template/scripts/gen-catalog-hashes.go --all ; echo "rc=$?"
rc=0
catalog.yaml updated successfully (12899 bytes)

$ go build -o bin/moai ./cmd/moai ; echo "rc=$?"
rc=0
```

The regen proposed TWO hash changes: the `moai` skill (M5's — the `todo.md` mirror edit) and
`sync-auditor` (the inherited one). Only M5's is kept; the `sync-auditor` line was restored to its
HEAD value by hand and verified byte-identical to `git show HEAD:internal/template/catalog.yaml`,
so the commit carries M5's delta alone and the inherited defect stays visible where it belongs.

---

## 9. Post-implementation verification (the M5 verdict basis)

Issued as one parallel batch; every exit code captured on the un-piped command.

```
$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.668s
kanban rc=0

$ go test ./internal/cli/... -count=1 -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli	432.249s
(17 packages ok)
cli rc=0

$ go test ./internal/template/... -count=1
--- FAIL: TestManifestHashFormat (0.10s)      ← INHERITED, §8
--- FAIL: TestGoldenCommittedArtifactsMatchEmission (0.00s)   ← INHERITED, §8
template rc=1

$ go vet ./internal/kanban/... ./internal/cli/... ./internal/template/...
vet rc=0   (no output)

$ gofmt -l internal/kanban/ internal/cli/ internal/template/
38 files
```

**The gofmt figure is not a regression, and the arithmetic matters.** M4 measured 28 across
`internal/kanban` + `internal/cli`. Re-measured per tree at this HEAD: `internal/kanban` **0**,
`internal/cli` **28** (unchanged from M4's baseline), `internal/template` **10** — a tree M4 never
measured, newly in scope because M5 touches the mirror. All 38 are pre-existing; none is
M5-touched:

```
$ gofmt -l internal/kanban/backlog_migrate.go internal/kanban/backlog_landing_roundtrip_test.go internal/kanban/backlog_downgrade_test.go internal/cli/todo_landed_doc_test.go ; echo "rc=$?"
rc=0   (no output — all four clean)
```

---

## 10. Known losses — deliberately NOT exported, and NOT citable as a verdict basis

Material named here was not measured, or was measured and discarded. None of it may be cited later
as evidence.

1. **`internal/template` has no pre-edit baseline in this run.** §2 baselined `internal/kanban` and
   `internal/cli` only, because `internal/template` entered scope later, when the mirror edit
   landed. §8's attribution of its two failures therefore rests on a DERIVATION (the failing
   assertions read only files with no working-tree modification, and the catalog line was verified
   byte-identical to HEAD) rather than on a before/after measurement. The derivation is sound but
   it is not the same thing as a baseline.
2. **The full local suite was never run**, by policy. Every package outside `internal/kanban`,
   `internal/cli`, and `internal/template` is UNMEASURED at this HEAD; CI owns that verdict.
3. **darwin/arm64 only.** No windows or linux build or test. Nothing M5 adds is
   platform-conditional, but that is an argument, not a measurement.
4. **`golangci-lint` was not run.** Only `go vet` and `gofmt`. The DoD's "project linter" line is
   unperformed for M5, as it was for M3 and M4.
5. **No coverage measurement.** Neither the delta nor the absolute figure for either package.
6. **No end-to-end run of the real binary.** `bin/moai` was built (exit 0) but never invoked; the
   `moai todo landed` → `moai todo pr` sequence remains unexercised as a shipped command, exactly
   as M4 recorded. Every fixture in this milestone is in-process.
7. **`make build` never completed as a whole.** Its body was run in pieces (§8). The `templ-generate`
   step it also depends on was NOT run, so nothing here establishes that step is clean.
8. **The evidence file is unverified against a schema.** No mechanical check confirms these claims
   match the commands above; a reader must compare them by hand.
