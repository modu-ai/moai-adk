# t1911 Verdict — t1632 preserved work: narrow reproduction + disposition

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1911` (branch `WT-mcp-server-go`, HEAD `8249bff8c3`, clean before and after)
- Preservation sources (read-only, untouched):
  `/Users/goos/MoAI/moai-adk-go/.moai/reports/adk-pr-priority-20261019/continuation/priority-handoff-20261020/t1632-preservation/`
- Verdict: **PRESERVE** — the preserved patch is bound to a different lineage and cannot be reproduced on this base.

## 1. Claim

1. The t1632 preservation patch applies cleanly to this tree — **REFUTED** (does not apply).
2. The preserved test file, copied byte-identically, compiles and its tests run on this base — **REFUTED** (build-fail: unmet dependencies).
3. The applied content is byte-identical to the manifest — test file **CONFIRMED** byte-identical; patched `mcp_server.go` result **NOT REPRODUCIBLE** here (patch cannot apply, so the manifest's `5d27453e…` post-patch hash is unreachable on this base).
4. Base tree itself is healthy (`go build ./internal/cli/...` exits 0) — **CONFIRMED**.

## 2. Evidence (commands + verbatim outputs)

### 2.1 Patch application — FAILS

```
$ git apply --check .../t1632-preservation/mcp_server.patch
error: patch failed: internal/cli/mcp_server.go:269
error: internal/cli/mcp_server.go: patch does not apply
APPLY_CHECK_EXIT=1
```

`git apply` was not attempted after the failed check (same deterministic failure).

### 2.2 sha256 comparison (manifest vs this run)

| File | Manifest | This run | Match |
|---|---|---|---|
| `internal/cli/mcp_audit_cache_guard_test.go` (4361 B) | `8f8ce6ed0e6b83f03a26f4300167b1d3b340899f7234c526c4205b4d3659d886` | `8f8ce6ed0e6b83f03a26f4300167b1d3b340899f7234c526c4205b4d3659d886` | YES (byte-identical) |
| `internal/cli/mcp_server.go` (post-patch, 82215 B) | `5d27453ec86e0dae97f622c170a498522383edaa50ee986d0f01a721024305ee` | n/a — patch cannot apply on this base | N/A (unreachable) |
| `mcp_server.patch` (the preservation artifact itself) | (not in manifest) | `fb0ad32eff7f6a2ba2fd30053cdee17e7e44a093bb64f3c556804cc503ffd535` | recorded for traceability |

Patch pre-image blob check: patch header says `index 22021f6d79..f1518d039d`;
`git rev-parse 9419bb9268:internal/cli/mcp_server.go` → `22021f6d796d…` — the patch is
faithful to the t1632 base, but that base file does not exist in this lineage.

### 2.3 Lineage divergence (root cause)

```
$ git merge-base --is-ancestor 9419bb9268 8249bff8c3  → NOT ancestor
$ git merge-base 9419bb9268 8249bff8c3               → 48bfe97afa
9419bb9268 = "Merge main into WT-3-2-0-2 (card t1632, absorb before landing remeasure)"
this tree mcp_server.go: 67,061 bytes, blob a947df01b4 (HEAD 8249bff8c3)
t1632  mcp_server.go:    82,215 bytes, blob 22021f6d79 (HEAD 9419bb9268)
```

Absent from this lineage entirely (grep over `internal/cli/**.go`, non-test):
`mcpStalenessGuard`, `serverStaleness`, `mcpcat.IsWriteCapable`,
`resolveToolProjectRootWithSource`, `resolveRequiredToolProjectRoot`,
`addMCPStalenessMetadata`, `mcpCallIsWriteCapable` — the patch modifies/depends on
machinery that only exists on the `WT-3-2-0-2` line. Only `handleAuditCache` (line 1019)
and `toolJSON` exist here.

### 2.4 Narrow reproduction (executed once) — BUILD FAIL

Test file copied to `internal/cli/mcp_audit_cache_guard_test.go` (byte-identical), then:

```
$ unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED
$ TMPDIR=/tmp go test ./internal/cli -run 'TestAuditCacheStoreRequiresProjectRoot|TestStaleAuditCacheStoreRefusesWithMetadata' -count=1 -p 1 -json
REPRO_EXIT=1
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:15:10: undefined: sdServerFixture\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:34:71: undefined: mcpTextOf\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:36:18: undefined: mcpTextOf\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:42:10: undefined: sdServerFixture\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:49:13: undefined: mcpStalenessGuard\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:63:56: undefined: mcpTextOf\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:70:13: undefined: mcpStalenessGuard\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:82:58: undefined: mcpTextOf\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:103:2: undefined: writeFileOfSize\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:104:2: undefined: withStalenessSeams\n"}
{"Action":"build-output","Output":"internal/cli/mcp_audit_cache_guard_test.go:104:2: too many errors\n"}
{"Action":"build-fail"}
FAIL  github.com/modu-ai/moai-adk/internal/cli [build failed]
```

Note: the preserved test file contains THREE test functions (the dispatch named two;
`TestStaleReadResultMetadataIsMirroredToTextJSON` is the third) + helpers
`forceStaleMCPServer` / `assertStaleMetadataInBothChannels` / `mcpHasText`. None compiled.

### 2.5 Neighbor run (touched-area)

```
$ TMPDIR=/tmp go test ./internal/cli -run 'TestStaleness|TestAuditCache' -count=1 -p 1 -v
NEIGHBOR_EXIT=0
testing: warning: no tests to run
PASS
ok    github.com/modu-ai/moai-adk/internal/cli  0.960s [no tests to run]
```

Selects nothing: no `TestStaleness*` / `TestAuditCache*` tests exist in this lineage —
consistent with the staleness-guard machinery being absent here.

### 2.6 Baseline build (tree restored to clean first)

```
$ rm internal/cli/mcp_audit_cache_guard_test.go
$ TMPDIR=/tmp go build ./internal/cli/...
BUILD_EXIT=0            (no output)
$ git status --short    → (empty); HEAD = 8249bff8c3
```

## 3. Baseline-attribution

Every observation above was produced in THIS run against THIS tree
(`.moai/worktrees/t1911` @ `8249bff8c3`): `git apply --check` (exit 1), `shasum -a 256`
on the preservation copies, the `-json` repro run (`REPRO_EXIT=1`), the neighbor run
(`NEIGHBOR_EXIT=0`, no tests to run), and the baseline build (`BUILD_EXIT=0`).
Raw outputs: `/tmp/t1911-repro.json`, `/tmp/t1911-neighbor.log`, `/tmp/t1911-build.log`.
The t1632 originals were never touched (read-only source paths only; no command in this
session wrote to `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1632`).

## 4. Gaps

- The narrow repro could not execute any test to a PASS/FAIL verdict — it died at compile.
  What was observed is the compiler's dependency verdict, not test behavior. That is the
  honest limit of reproduction on this base.
- **Guard refusal (recorded per §3.1)**: the worktree-isolation guard refused one compound
  command form (cp + env-unset go test in one line: "a construct too complex to verify").
  Remediation: split into plain separate commands and re-run — the identical repro command
  then ran once and its output is §2.4. The refused form was not retried.
- The manifest's post-patch `mcp_server.go` hash (5d27453e…) is UNVERIFIED against this
  tree by construction (patch cannot apply). It remains verified only against the t1632
  tree's own state, which this dispatch may not touch.
- Whether the preserved work compiles green on its OWN lineage (9419bb9268) was NOT
  measured here — out of scope for this dispatch (t1632 tree off-limits). The PRESERVE
  verdict does not claim it is red there; it claims it cannot be reproduced here.

## 5. Residual-risk

- `48bfe97afa` (merge-base) means `WT-3-2-0-2` and `WT-mcp-server-go` share history; if a
  future merge brings the staleness-guard machinery (and helpers) into this lineage, the
  patch may become applicable — re-run this reproduction before assuming it.
- The manifest records the preserved state at one moment (2026-10-20T09:07Z @ 9419bb9268).
  If the t1632 tree's uncommitted work has since evolved, this verdict describes the
  snapshot only.
- Helper-symbol absence was confirmed by both grep and the compiler; however, the helper
  definitions' exact home files on the `WT-3-2-0-2` lineage were not enumerated (no need
  for the verdict; relevant only if the leader chooses a port).

## Disposition recommendation: PRESERVE

Reason: the preserved work is **lineage-bound to `WT-3-2-0-2`** (base `9419bb9268`). On this
base (`8249bff8c3`, diverged at `48bfe97afa`) the patch does not apply, and the test file
byte-identically copied fails to compile against 10 missing symbols spanning the entire
staleness-guard feature (`mcpStalenessGuard`, `mcpcat.IsWriteCapable`, staleness seams and
test fixtures). A "reproduction" here would require porting an entire untrunked dependency
chain — that would be new engineering, not the preserved t1632 work, and would violate the
byte-identity requirement.

For the leader (choice per dispatch): **commit the two preserved files in the t1632 tree /
`WT-3-2-0-2` lineage where their dependencies exist** (recommended — the work belongs to
that line and was measured only against it), or explicitly open a separate scoped card to
port the staleness-guard feature + tests to main as new work. What should NOT happen: committing
these two files onto `WT-mcp-server-go` as-is (they do not compile here).

Kill discipline: no pattern kills (no pkill/killall/kill $(pgrep)) were used in this session —
no process kill was needed at all; a guard deny (worktree-isolation, §4) was treated as final
and remediated by splitting the command form, not by retrying it.
