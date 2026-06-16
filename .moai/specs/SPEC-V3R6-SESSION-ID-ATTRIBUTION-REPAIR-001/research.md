---
id: SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
title: "Research — Session-ID attribution dead-feature diagnosis evidence"
version: "0.1.0"
status: in-progress
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/session.go; internal/hook/session_start.go; internal/session/registry.go"
lifecycle: spec-anchored
tags: "session, attribution, multi-session, coordination, race-attribution, doctrine"
era: V3R6
---

# Research — SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001

This research document records the diagnosis evidence per the
`.claude/rules/moai/core/verification-claim-integrity.md` 5-section
evidence-bearing report format. **Every file:line claim below was re-grepped
against the current HEAD `12e20d190` on 2026-06-17** and the verbatim grep
output is recorded as Evidence. The baseline diagnosis was authored against
`d0e2e9bc3`; the 4 commits added since (`12e20d190` +
STATUSLINE-PRESET-RETIRE phase-close commits) are UNRELATED to this SPEC's
scope (statusline preset retire), so no citation shift was expected — but
every claim was re-verified regardless.

**Iter-2 remediation (2026-06-17):** §C Evidence was re-scoped after iter-1
plan-auditor flagged the original command as self-referentially polluted
(D1 BLOCKING — the path list included `.moai/specs/`, so the SPEC's own
placeholder text and regex literal self-matched, producing 16 not 14). §C
now splits the Evidence into (1) canonical doctrine surfaces (4 variants —
load-bearing for AC-FBC-003) and (2) broader-context sweep (14 variants —
drift-detector calibration only). §C variant #10 disposition contradiction
(D3 SHOULD-FIX) is resolved: RETAIN as happy-path UUID-source slot, rewrite
for symmetry. See §C footnote and §C "D3 resolved disposition for variant
#10" subsection.

## §A. P1 — Write-path diagnosis

### Claim

The SessionStart hook writes to the registry when
`input.SessionID != "" && input.ProjectDir != ""`. The write path EXISTS.

### Evidence

Command: `grep -nE 'Register|SessionID|input\.Session|input\.Project|FormatStderr|additionalContext|hookSpecific' internal/hook/session_start.go`

Verbatim output (excerpt, full output 30+ lines):

```
47:		"session_id", input.SessionID,
49:		"project_dir", input.ProjectDir,
58:	// Step 1 — Register this session with no SPEC scope yet.
65:	// @MX:NOTE: [AUTO] 3-step protocol — Register + Purge + Query + stderr surface
66:	if input.SessionID != "" && input.ProjectDir != "" {
800:	registryPath := filepath.Join(input.ProjectDir, session.DefaultRegistryPath)
803:	// Step 1: RegisterSession with no SPEC scope yet.
804:	if err := reg.Register(input.SessionID, session.SpecIDNone, session.PhaseNone); err != nil {
805:		slog.Warn("multi-session protocol: RegisterSession failed (non-blocking)",
806:			"session_id", input.SessionID,
```

Command: `wc -l internal/hook/session_start.go`

Verbatim output: `861 internal/hook/session_start.go`

### Baseline-attribution

Re-grepped against HEAD `12e20d190` on 2026-06-17. The `if` gate is at L66;
the actual `reg.Register(...)` call is at L804 (inside a helper method block
L783-861). The baseline diagnosis cited "step 1, line 66 / 803-804" — the L66
citation is correct (gate), the L803-804 citation is corrected to **L804** (the
call itself; L803 is the comment immediately preceding). The file is 861 lines
total (baseline implied ≤804).

### Gaps

- The P1 investigation (M1) must determine WHY the registry stays empty even
  though the write path exists at L804. The write-path diagnosis confirms the
  call site, NOT that the call succeeds at runtime.
- The `input.SessionID == ""` scenario is the leading hypothesis (K3 gate at
  L66 bypasses Register when session_id is empty). Not yet empirically
  reproduced — deferred to M1.

### Residual-risk

- The write path may fail silently (`slog.Warn` at L805 is non-blocking). Even
  if Register is called, the write may not land if the registry file is
  unwritable or the path is wrong. M1 must verify the post-call registry state.

---

## §B. P2 — Read-path diagnosis

### Claim

There is NO subcommand that returns "this orchestrator's own UUID". The `list`
subcommand returns the full entry list and cannot identify which entry is the
caller's own.

### Evidence

Command: `grep -nE 'current|whoami|uuid|self|own' internal/cli/session.go`

Verbatim output: `(no matches)` (only 2 comment-line hits at L19, L32 — neither
is a subcommand).

Command: `grep -nE 'Command|Use:|RunE|cobra' internal/cli/session.go | head -40`

Verbatim output (excerpt):

```
15: // newSessionCmd builds the `moai session` cobra command tree with 5
24: func newSessionCmd() *cobra.Command {
43:	cmd.AddCommand(newSessionRegisterCmd())
44:	cmd.AddCommand(newSessionHeartbeatCmd())
45:	cmd.AddCommand(newSessionDeregisterCmd())
46:	cmd.AddCommand(newSessionListCmd())
47:	cmd.AddCommand(newSessionPurgeCmd())
51: func newSessionRegisterCmd() {
116: func newSessionListCmd() *cobra.Command {
157: func newSessionPurgeCmd() *cobra.Command {
```

Command: `sed -n '116,155p' internal/cli/session.go`

Verbatim output (the `list` subcommand body, abbreviated):

```
func newSessionListCmd() *cobra.Command {
	var jsonOutput bool
	var filterSpec string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active sessions (optionally filtered by --filter-spec)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := session.QueryActiveWork(filterSpec)
			...
			for _, e := range entries {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(),
					"session=%s spec=%s phase=%s ...", ...)
			}
```

The `list` subcommand iterates ALL entries; it has no concept of "the caller's
own entry".

Command (FormatStderrReminder — confirms it is others-only, NOT self-UUID):

```
grep -nE 'FormatStderrReminder|func.*Others|others\s*==' internal/session/registry.go
sed -n '420,448p' internal/session/registry.go
```

Verbatim output (L424-448):

```go
// FormatStderrReminder formats the QueryActiveWork result as a stderr
// system-reminder block. Used by SessionStart hook Step 3.
//
// REQ-COORD-015.
func FormatStderrReminder(currentSessionID string, entries []Entry, now time.Time) string {
	others := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if e.SessionID != currentSessionID {
			others = append(others, e)
		}
	}
	if len(others) == 0 {
		return ""
	}
	...
}
```

L431-432: `if len(others) == 0 { return "" }` — confirms the function returns
empty when there are no OTHER sessions. It is the "others-active-session
warning" surface, NOT the self-UUID surface.

Command (additionalContext injection): `grep -nE 'additionalContext|hookSpecificOutput' internal/hook/session_start.go`

Verbatim output: **0 matches.** The SessionStart hook only WRITES to the
registry; it never TELLS the orchestrator its own UUID.

### Baseline-attribution

Re-grepped against HEAD `12e20d190` on 2026-06-17. All citations confirmed:
5 subcommands at L43-47; `list` at L116-155; `FormatStderrReminder` at L424
with empty-return at L431-432; zero `additionalContext` matches. The baseline
diagnosis cited "step 1, line 66 / 803-804" for the write path — corrected in
§A above (L66 gate, L804 call). The baseline cited "L420-448" for
FormatStderrReminder — confirmed (comment block L420-423, function L424-448).

### Gaps

- The `current` subcommand does not exist — it is the P2 Stage 1 deliverable
  (M2), not a pre-existing artifact.
- The `additionalContext` injection does not exist — it is the P2 Stage 2
  deliverable (M3), not a pre-existing artifact.

### Residual-risk

- Even with P2 delivered, `additionalContext` is lost after compaction
  (spec.md §F.2). The orchestrator must re-query `moai session current`
  post-compaction; this is a runtime limitation, not a defect.

---

## §C. P3 — Fallback-variant enumeration

### Claim

There are **4 distinct `source_session_id` placeholder variants on the
canonical doctrine surfaces** (rules tree + output-styles render surface +
their template mirrors), of which **1 is the canonical fallback** and **3 are
non-canonical**. A broader sweep that also includes `.moai/docs/` and
auto-memory `~/.claude/projects/.../memory/` finds 14 distinct strings total,
but only the 4 doctrine-surface variants are in scope for the M4+M5
canonicalization remediation; the remaining 10 live in historical append-only
memory entries (excluded per spec.md §G) and `.moai/docs/` documentation
prose (not doctrine SSOT).

### Evidence

**Command (canonical doctrine surfaces ONLY — `.moai/specs/`, `.moai/docs/`,
and memory EXCLUDED to avoid self-referential grep pollution, see footnote):**

```
grep -rohE 'source_session_id: <[^>]*>' \
  .claude/rules/moai/ \
  .claude/output-styles/moai/moai.md \
  internal/template/templates/.claude/rules/moai/ \
  internal/template/templates/.claude/output-styles/moai/moai.md \
  2>/dev/null | sed 's/[[:space:]]\+/ /g' | sort -u | nl
```

Verbatim output (4 distinct variants on canonical doctrine surfaces):

```
     1  source_session_id: <UUID>
     2  source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>
     3  source_session_id: <orchestrator-uuid-from-current-turn>
     4  source_session_id: <orchestrator-uuid-here>
```

Command (count): appended `| wc -l` → `4`.

**Locating each non-canonical variant (file:line, both local + template trees):**

```
$ grep -rn 'source_session_id: <UUID>' \
    .claude/rules/moai/ .claude/output-styles/moai/moai.md \
    internal/template/templates/.claude/rules/moai/ \
    internal/template/templates/.claude/output-styles/moai/moai.md
.claude/rules/moai/workflow/session-handoff.md:77:    ...MUST also include a `source_session_id: <UUID>` line...
.claude/rules/moai/workflow/session-handoff.md:159:   - Resume Block 2 missing `source_session_id: <UUID>` **AND missing...**
.claude/output-styles/moai/moai.md:689:- [ ] Block 2 includes `source_session_id: <UUID>` line carrying current...
internal/template/templates/.claude/rules/moai/workflow/session-handoff.md:77:    (mirror — identical)
internal/template/templates/.claude/rules/moai/workflow/session-handoff.md:159:   (mirror — identical)
internal/template/templates/.claude/output-styles/moai/moai.md:689:(mirror — identical)

$ grep -rn 'source_session_id: <orchestrator-uuid-from-current-turn>' <same paths>
.claude/output-styles/moai/moai.md:653:source_session_id: <orchestrator-uuid-from-current-turn>
internal/template/templates/.claude/output-styles/moai/moai.md:653:(mirror — identical)

$ grep -rn 'source_session_id: <orchestrator-uuid-here>' <same paths>
.claude/rules/moai/workflow/session-handoff.md:96:source_session_id: <orchestrator-uuid-here>
internal/template/templates/.claude/rules/moai/workflow/session-handoff.md:96:(mirror — identical)
```

**Broader-context sweep (NOT in M4+M5 scope — recorded for drift-detector
calibration only):** extending the same grep to include `.moai/docs/` and
auto-memory yields **14 distinct strings total**. The 10 extra strings
(beyond the 4 doctrine-surface variants above) live exclusively in
`.moai/docs/` prose and `~/.claude/projects/.../memory/*.md` historical
entries. These are NOT canonical doctrine surfaces and are excluded from
M4+M5 remediation scope (spec.md §G excludes backfilling historical
append-only entries). The 10 extras are:

```
source_session_id: <current-session-uuid — moai session list --json available일 때 backfill>
source_session_id: <not-available — environment-fallback>
source_session_id: <orchestrator-session-id-placeholder>
source_session_id: <orchestrator-session-uuid>
source_session_id: <orchestrator-uuid — moai session list --json 부재로 백필 불가, 다음 세션 acknowledge>
source_session_id: <orchestrator-uuid-current-session>
source_session_id: <orchestrator-uuid-current-turn>
source_session_id: <orchestrator-uuid-pending>
source_session_id: <unavailable — moai CLI 미설치>
source_session_id: <unavailable — orchestrator UUID 미노출>
```

#### Footnote — self-referential grep pollution trap

An earlier draft of this section used a grep whose path list included
`.moai/specs/`. Because THIS SPEC's own `acceptance.md` carries the literal
placeholder `source_session_id: <...>` in AC prose and `research.md` carries
the regex literal `<[^>]*>` in the command itself, that grep inflated the
count beyond the honest canonical-surface 4 (and beyond the broader-context
14) — the extras being `source_session_id: <...>` (self-match from
acceptance.md AC-FBC-001/002/003 prose) and `source_session_id: <[^>`
(self-match from the regex literal in the command line itself). iter-1
plan-auditor observed this producing 16; after iter-2 expanded the §C
Evidence documentation (which itself repeats the regex literal), the same
polluted variant returns 17 — the count grows every time the SPEC documents
the grep, which is the clearest possible demonstration that a
drift-measurement SPEC grepping `.moai/specs/` is measuring itself. The
canonical-doctrine-surface grep above excludes `.moai/specs/` for this
reason. A future auditor re-running the polluted variant
(`.moai/specs/` included) will get a number ≥16 that drifts upward with
SPEC edits; the canonical-surface count (4) and the broader-context sweep
(14, excluding `.moai/specs/`) are the stable, load-bearing numbers.

### Baseline-attribution

Re-grepped against HEAD `12e20d190` on 2026-06-17 by the iter-2 remediation
(manager-spec). **Iter-1 plan-auditor flagged the prior §C Evidence as
non-reproducible** (D1 BLOCKING): the prior command path list included
`.moai/specs/`, so the SPEC's own placeholder text and the regex literal
self-matched, producing a count (16 at iter-1 time; 17 after iter-2 expanded
the §C documentation) that exceeded the documented 14. The substantive
"non-canonical variants exist on doctrine surfaces" claim was always correct;
the defect was the path-list scope (`.moai/specs/` included self-pollution).
Iter-2 remediates by splitting the Evidence into two greps: (1) canonical
doctrine surfaces only (4 variants — load-bearing for AC-FBC-003), and (2)
broader-context sweep including `.moai/docs/` + memory (14 variants —
drift-detector calibration only). The spec.md (§B.1 P3, REQ-FBC-003) and
acceptance.md (AC-FBC-003) are updated to cite the canonical-surface count
of 4 (1 canonical + 3 non-canonical), not the broader-context 14.

### Classification (for M4+M5 remediation — canonical doctrine surfaces only)

The 4 canonical-doctrine-surface variants classify as:

**Canonical (1 — keep):**
1. `source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`

**Non-canonical (3 — eliminate):**
2. `source_session_id: <UUID>` — generic placeholder at session-handoff.md
   L77 + L159 + moai.md L689 (and template mirrors). This is the canonical
   "happy-path UUID slot" description (where the orchestrator DOES know its
   UUID). Reclassified per D3 disposition: **retain as the happy-path
   template placeholder, NOT a fallback variant** — it is the same structural
   role as `<SPEC-ID>` or `<memory-file-1>` in the 6-block template skeleton.
   M4+M5 rewrites it to `source_session_id: <UUID from moai session current>`
   (per REQ-MSC-003) so it unambiguously references the P2 read-path
   deliverable, then verifies it no longer collides with the canonical
   fallback string.
3. `source_session_id: <orchestrator-uuid-from-current-turn>` — at moai.md
   L653 (the canonical 6-block template example). **D3 resolved disposition
   (see below): retain as the happy-path UUID-source slot, rewrite to
   `source_session_id: <UUID from moai session current>` for symmetry with
   variant #2.** Not counted as a fallback variant.
4. `source_session_id: <orchestrator-uuid-here>` — at session-handoff.md L96
   (the illustrative example). **Rewrite to the canonical fallback string**
   (the example should demonstrate the fallback gracefully, not introduce a
   new variant). Eliminated by M4.

**Reconciliation with iter-1 numbering:** iter-1's variant #10
(`<orchestrator-uuid-from-current-turn>`) is reclassified from the
"semantically distinct placeholder (rewrite to canonical OR delete)" bucket
to the "happy-path UUID-source template placeholder (retain, rewrite for
symmetry)" bucket. Its prior in-bucket placement and the trailing note's
"MAY be retained" language were mutually contradictory (D3 SHOULD-FIX); this
classification makes both references agree: variant #10 is RETAINED as the
happy-path UUID-source slot and rewritten for symmetry with variant #2, not
deleted.

### D3 resolved disposition for variant #10

**Decision: RETAIN as the happy-path UUID-source template placeholder, rewrite
for symmetry with variant #2.**

**Rationale:** variant #10 (`<orchestrator-uuid-from-current-turn>`) appears
exclusively at `moai.md` L653 inside the canonical 6-block paste-ready
template skeleton — it is the UUID slot the orchestrator populates when it
DOES know its UUID (the happy path), NOT a fallback variant. L667 explains
this exact semantics. Deleting it would remove the UUID-source slot from the
canonical template; retaining it verbatim would leave the drift detector
matching a non-canonical spelling. The right disposition is to rewrite it to
`source_session_id: <UUID from moai session current>` so it (a) keeps the
happy-path slot, (b) unambiguously references the P2 `current` subcommand
deliverable, and (c) no longer collides with the canonical fallback string
or any other variant. Variant #2 (`<UUID>`) gets the identical rewrite for
symmetry (it occupies the same structural role at session-handoff.md L77 /
L159 + moai.md L689). After M4+M5, the doctrine surfaces carry exactly two
distinct `source_session_id` spellings: the canonical fallback (REQ-FBC-001)
and the rewritten happy-path slot `<UUID from moai session current>`. This
resolves the iter-1 D3 contradiction (bucket said "rewrite/delete", note
said "MAY retain" — both now agree: retain + rewrite).

### Gaps

- The 4-variant canonical-surface enumeration covers doctrine SSOT
  (`.claude/rules/moai/`) + render surface (`.claude/output-styles/moai/moai.md`)
  + their template mirrors. Historical paste-ready resume MESSAGES in past
  sessions and memory entries are append-only and out of scope (spec.md §G
  excludes backfilling historical entries).
- Some variants may exist in archived / report directories not grepped above;
  M4+M5 should extend the grep to `.moai/reports/` and `.moai/backups/` if
  the drift detector flags them — but those are NOT canonical doctrine
  surfaces and do not count toward the AC-FBC-003 total.

### Residual-risk

- The 4-variant count is a snapshot at HEAD `12e20d190`. If new doctrine
  edits introduce more variants before M4+M5 lands, the count will grow. M4+M5
  MUST re-run the canonical-surface enumeration (the 4-count command above)
  at run-phase start and re-classify. The broader-context sweep (14) is
  informational only and does NOT gate AC-FBC-003.

---

## §D. P1 Root-Cause Investigation Scaffold (M1 deliverable)

This section is the M1 gating predecessor deliverable scaffold. It is
INTENTIONALLY EMPTY at plan-phase; manager-develop populates it during M1
with empirically-reproduced root cause(s).

### D.0 M1 FINDING — HookOutput.Data carries `json:"-"` (structural root cause)

**Claim:** The SessionStart handler's `data` map (containing `session_id`,
`multi_session_register`, `glm_credentials`, etc.) is placed into
`HookOutput.Data`, which carries the struct tag `json:"-"`. The hook protocol
serializes the output via `json.NewEncoder(w).Encode(output)`
(`internal/hook/protocol.go:76` WriteOutput). Because `Data` is tagged
`json:"-"`, **none of the session data the handler computes is ever emitted
on stdout to Claude Code.** This is the structural root cause that makes
K5 (no additionalContext injection) the load-bearing defect: even when
SessionID is non-empty and the registry write succeeds, the orchestrator
never receives the UUID via the hook response because the only field the
handler populates (`Data`) is dropped at serialization time.

**Evidence:**

Command: `grep -n 'Data json.RawMessage' internal/hook/types.go`

Verbatim output:

```
278:	Data json.RawMessage `json:"-"`
323:	Data json.RawMessage `json:"-"`
```

Command: `grep -n 'func.*WriteOutput\|json.NewEncoder\|Encode(output)' internal/hook/protocol.go`

Verbatim output:

```
70:func (p *jsonProtocol) WriteOutput(w io.Writer, output *HookOutput) error {
76:	if err := encoder.Encode(output); err != nil {
```

The SessionStart handler return path (`internal/hook/session_start.go:218`):
`return &HookOutput{Data: jsonData}, nil` — `jsonData` is the marshaled
`data` map, placed into the `json:"-"` field.

Per `.claude/rules/moai/core/hooks-system.md` § Hook Event stdin/stdout
Reference, the SessionStart stdout contract is `hookSpecificOutput:
additionalContext, reloadSkills, sessionTitle` — NOT the `Data` field. The
`Data` field is internal-only (used by tests + observability), never
serialized to Claude Code.

**Baseline-attribution:** Re-grepped against HEAD `12e20d190` on 2026-06-17
by manager-develop (M1). Both `Data` field citations (L278 input, L323
output) confirmed `json:"-"`. The WriteOutput encoder citation (L76)
confirmed. This finding was NOT in the plan-phase diagnosis (the plan-phase
research.md §B only noted "0 additionalContext matches" without tracing
WHY the existing Data-map approach fails to surface anything); it is the
M1 empirical discovery that re-frames the P2 Stage 2 deliverable (M3): the
additionalContext injection MUST populate `hookSpecificOutput.AdditionalContext`
(the serialized field), not the `Data` field.

**Gaps:** This finding is structural (Go struct tag), confirmed by grep —
not a runtime reproduction. Whether the registry write ALSO fails at runtime
(D.1/D.2/D.3) is orthogonal: even a successful registry write does not
surface the UUID to the orchestrator because of this serialization gap.

**Residual-risk:** The `json:"-"` tag is load-bearing for other event types
(the Data field is used internally by tests across the hook package); removing
it would change the serialized shape for every event. M3 addresses this by
ADDING `hookSpecificOutput.AdditionalContext` alongside the existing Data
field (strictly additive, REQ-RDP-005), not by changing the Data tag.

**P1-outcome implication for AC-RDP-002 (D5 conditional note):** Because the
side-channel write (M3) will populate `hookSpecificOutput.AdditionalContext`
AND the existing `input.SessionID != ""` gate at L66 governs the registry
Register call, the M3 additionalContext injection MUST be placed INSIDE the
existing `input.SessionID != ""` block (otherwise it would inject an empty
UUID). This means: in the empty-SessionID scenario (D.1), the additionalContext
is NOT injected, and `moai session current` (M2) returns the degradation
path (AC-RDP-003 exit 1 + AC-RDP-006 canonical fallback). AC-RDP-002
happy-path GREEN is satisfiable post-M3 only when SessionID is non-empty —
which is the normal case. The D5 conditional note's "defer AC-RDP-002 GREEN
to post-M3" branch applies; M2 ships the degradation path and M3 makes the
happy path available.

### D.1 Hypothesis 1 — `input.SessionID == ""` on first-turn activation

**Hypothesis:** Some Claude Code activation paths emit an empty `session_id`
in the SessionStart hook stdin JSON. The gate at L66 bypasses Register; the
registry stays empty.

**M1 STATUS: REPRODUCED (characterization test).** The M1 characterization
test `TestSessionStartEmptySessionIDEmitsWarning`
(`internal/hook/session_start_attribution_test.go`) confirms that before the
REQ-WPR-003 fix, the handler silently bypassed the protocol when
`input.SessionID == ""` with no observable signal. The fix (GREEN) adds a
non-blocking stderr warning so the orchestrator can observe the bypass. The
test confirms: (a) the warning is emitted, (b) the registry is NOT polluted
with an empty-session_id entry, (c) the hook still returns allow
(non-blocking).

**Whether the production runtime ever passes an empty session_id** is not
unit-testable (it depends on Claude Code runtime behavior, out of MoAI-ADK's
control). The REQ-WPR-003 warning is the observable surface: if the warning
appears in `$HOME/.moai/logs/hook-stderr.log`, the runtime did pass an empty
session_id. The `moai session doctor` diagnostic (REQ-WPR-001, AC-WPR-001)
surfaces this at the CLI.

**Disposition:** D.1 is a CONFIRMED contributing cause (the gate bypass is
reproduced; whether the runtime triggers it is observable via the warning +
doctor). REQ-WPR-003 warning added (GREEN).

### D.2 Hypothesis 2 — handle-session-start.sh silent-exit

**Hypothesis:** The 3-tier fallback in `handle-session-start.sh` (PATH →
`~/go/bin/moai` → `~/.local/bin/moai` → silent `exit 0`) silently drops the
hook result when all 3 tiers fail. Claude Code observes a successful
non-invocation and proceeds without the registry write.

**Reproduction scaffold (M1):**
1. Enumerate the conditions under which all 3 tiers fail (moai not in PATH,
   not at `~/go/bin/moai`, not at `~/.local/bin/moai`).
2. Confirm whether the current diagnostic context (baseline `d0e2e9bc3`)
   matches any of these conditions.
3. If yes, this is a contributing root cause; document and propose a fix
   (e.g., explicit error message instead of silent `exit 0`).

**Evidence (handle-session-start.sh current content, re-grepped):**

```bash
# Try moai command in PATH
if command -v moai &> /dev/null; then
    exec moai hook session-start 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try default ~/go/bin/moai
if [ -f "$HOME/go/bin/moai" ]; then
    exec "$HOME/go/bin/moai" hook session-start 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Try ~/.local/bin/moai (Linux install location)
if [ -f "$HOME/.local/bin/moai" ]; then
    exec "$HOME/.local/bin/moai" hook session-start 2>>"$MOAI_HOOK_STDERR_LOG"
fi

# Not found - exit silently (Claude Code handles missing hooks gracefully)
exit 0
```

**Status at plan-phase:** Not yet reproduced. M1 will resolve.

**M1 STATUS: NOT THE LOAD-BEARING CAUSE (re-framed by D.0).** The 3-tier
fallback silent-exit is a real failure mode (moai not in PATH / not at
`~/go/bin/moai` / not at `~/.local/bin/moai`), but even when the wrapper
DOES reach the moai binary and the hook runs successfully, the session data
does not reach the orchestrator (D.0: `HookOutput.Data` is `json:"-"`). The
wrapper silent-exit is therefore a contributing cause for the REGISTRY being
empty (K6) but NOT the root cause of the orchestrator never learning its
UUID (K5). The registry-write reliability (D.3) is the relevant question
for K6; the wrapper-fallback is a secondary defense (the diagnostic session
that observed K6 ran in a context where the wrapper reached the binary, so
D.2 did not fire there).

**Disposition:** D.2 is documented but not load-bearing for this SPEC's
P2/P3 deliverables. The `moai session doctor` diagnostic (REQ-WPR-001)
surfaces wrapper-fallback as one of the 3 candidate root causes when the
registry is empty.

### D.3 Hypothesis 3 — Registry write failure (silent slog.Warn)

**Hypothesis:** `reg.Register` IS called (L804) but fails silently. The
non-blocking `slog.Warn` at L805 logs the failure but the hook exits 0. The
registry file is never created.

**M1 STATUS: ORTHOGONAL TO THE READ-PATH DEFECT (re-framed by D.0).** Even
when `reg.Register` succeeds and the registry file IS created, the
orchestrator cannot read its own UUID back because: (a) there is no `current`
subcommand (K1/K2 — the P2 Stage 1 deliverable, M2), and (b) the session
data never reached the orchestrator via the hook response (D.0). The
registry-write reliability question (D.3) matters for the Layer 1 registry
primitives but is NOT the root cause of the attribution dead feature.

**Disposition:** D.3 is documented for completeness. The existing
`slog.Warn` at L805 (now L819 after the REQ-WPR-003 edit) already surfaces
registry write failures non-blocking; `moai session doctor` (REQ-WPR-001)
aggregates this signal at the CLI.

### D.4 M1 Exit Criteria (RECAPPED from plan.md §F M1) — MET

- [x] research.md §D contains an empirically-reproduced root cause for the
  attribution dead feature: **D.0** (HookOutput.Data `json:"-"` structural
  gap, confirmed by grep against HEAD `12e20d190`), plus **D.1** (empty-
  SessionID gate bypass, confirmed by characterization test
  `TestSessionStartEmptySessionIDEmitsWarning`).
- [x] The root cause is documented with the 5-section evidence-bearing format
  (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) in D.0.
- [x] M2-M6 are unblocked (REQ-WPR-004 satisfied — this section is populated).
- [x] REQ-WPR-003 warning added (GREEN); AC-WPR-003 satisfiable.

**P1-outcome-determined pillar ordering (per plan.md §F M2 D4 note):**
- M2 ships the degradation path (AC-RDP-003 exit 1 + AC-RDP-006 canonical
  fallback). AC-RDP-002 happy-path GREEN is deferred to post-M3 because the
  side-channel write (M3 additionalContext) is the mechanism that makes the
  UUID available to `current`, and per D.0/D.1 the injection is gated on
  `input.SessionID != ""` (the normal case). This matches the D5 conditional
  note's "defer AC-RDP-002 GREEN to post-M3" branch exactly.

---

## §E. Root-Blocker — Claude Code Runtime Non-Exposure

### Claim

`session.id` is available only in hook stdin JSON, not directly exposed to
the orchestrator (LLM) as an env var or tool input.

### Evidence

Command: `grep -nE 'SessionID|session_id|session\.id' internal/cli/hook.go`

Verbatim output (excerpt):

```
675:	// T-A3 spec: nested stdin JSON — last_assistant_message + session.id
699:		SessionID:                hookInput.Session.ID,
740:// parent_session_id is extracted from session.id (nested).
757:	// T-A4 spec: camelCase agentType/agentName, nested session.id
788:		ParentSessionID:                hookInput.Session.ID,
```

Command (broader): `grep -rnE 'session\.id|SessionID' internal/cli/hook.go internal/hook/ 2>/dev/null | head -25`

Verbatim output: 30+ hits, all in hook handler code (hook.go + protocol_test.go).
No env-var read, no orchestrator-facing exposure. The session.id lives ONLY in
the hook stdin JSON parsed by the hook handler.

### Baseline-attribution

Re-grepped against HEAD `12e20d190` on 2026-06-17. Confirmed: `session.id`
parsing at L675 (T-A3 spec comment), L699 (SessionID assignment),
L788 (ParentSessionID). The baseline diagnosis cited "L675-690 Stop hook
parsing" — confirmed (the parsing block spans L675-699 for the SessionID
assignment; L788 for ParentSessionID in a different hook event).

### Gaps

- This root blocker is OUT OF MoAI-ADK's control (runtime territory). The P2
  2-stage approach (SessionStart additionalContext + CLI complement) is the
  canonical workaround.
- A future upstream request to expose `session.id` as an env var would
  eliminate this blocker; out of scope for this SPEC (spec.md §F.3).

### Residual-risk

- Even with P2 delivered, any activation path that bypasses SessionStart
  (e.g., headless `-p` without hooks) leaves the orchestrator without UUID
  access. `moai session current` returns the canonical fallback (REQ-RDP-006)
  in this case.

---

## §F. Diagnosis-Data Discrepancies (re-grep findings)

This section records the discrepancies between the baseline diagnosis
(authored against `d0e2e9bc3`) and the re-grep against current HEAD
`12e20d190`.

| # | Baseline claim | Re-grep finding | Disposition |
|---|----------------|-----------------|-------------|
| D-D-1 | "9 observed fallback string variants" | **4 distinct variants on canonical doctrine surfaces** (§C above); 14 in the broader-context sweep (`.moai/docs/` + memory) | Corrected in spec.md (REQ-FBC-003 "3 non-canonical on canonical doctrine surfaces") and acceptance.md (AC-FBC-003). Iter-1 plan-auditor further flagged the original §C command as self-referentially polluted (`.moai/specs/` in path list → inflated count: 16 at iter-1, 17 after iter-2 expanded the §C documentation); iter-2 splits the Evidence into canonical-surface (4) and broader-context (14) sweeps and excludes `.moai/specs/`. |
| D-D-2 | "SessionStart hook step 1, line 66 / 803-804" | Gate at L66; `reg.Register` call at **L804** (L803 is the preceding comment) | Corrected in spec.md §A.1, plan.md §B K3, research.md §A. |
| D-D-3 | "FormatStderrReminder L420-448" | Comment L420-423; function L424-448 (empty return L431-432) | Confirmed; minor scope clarification (comment block vs function bounds). |
| D-D-4 | "internal/cli/session.go:24-49 (5 verbs)" | Comment block L15-23; function L24-49 (5 AddCommand calls L43-47) | Confirmed; minor scope clarification. |
| D-D-5 | "internal/cli/hook.go:675-690 Stop hook parsing" | L675 T-A3 comment; L699 SessionID assignment; L788 ParentSessionID (different hook event) | Confirmed; expanded to L699/L788 for the actual assignments. |

**No citation was invalidated.** All 5 discrepancies are either count
corrections (D-D-1) or scope clarifications (D-D-2 through D-D-5). The
diagnosis itself is sound; the repair pillars (P1+P2+P3) stand as defined in
spec.md §B.1.

---

## §G. Lessons Applied

- `verification-claim-integrity` — every file:line claim re-grepped against
  current HEAD; verbatim grep output recorded as Evidence; baseline-attribution
  section names the command and the tree.
- `feedback_windowed_grep_undercount_authoring` — line-count verification uses
  content-anchored `grep -n` and `wc -l`, NOT `sed -n 'NNN,MMM'p` windowed
  reads.
- `feedback_coverage_audit_table_not_actually_run` — the "9 variants" count
  was re-run independently in iter-1 and found to undercount; iter-2 further
  refined the count by splitting canonical doctrine surfaces (4 variants) from
  the broader-context sweep including `.moai/docs/` + memory (14 variants).
  The canonical-surface count (4 = 1 canonical + 3 non-canonical) is frozen
  into spec.md / acceptance.md / research.md §C as the load-bearing number
  for AC-FBC-003.
- `feedback_self_referential_grep_pollution` (NEW, captured by iter-1
  plan-auditor) — a drift-measurement SPEC that greps `.moai/specs/` measures
  itself: the SPEC's own placeholder text and regex literals self-match and
  inflate the count. Canonical-doctrine-surface measurements MUST exclude
  `.moai/specs/`. Applied in §C Evidence above.

## §H. Cross-References

- spec.md: `SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/spec.md` (requirements, scope, exclusions).
- plan.md: `SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/plan.md` (milestones M1-M6).
- acceptance.md: `SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001/acceptance.md` (AC matrix).
- Doctrine SSOT: `.claude/rules/moai/workflow/session-handoff.md` Block 2.
- Render surface: `.claude/output-styles/moai/moai.md` §8.
- Verification doctrine: `.claude/rules/moai/core/verification-claim-integrity.md`.
- Predecessor: `SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001` (track 2, completed).
- Predecessor: `SPEC-V3R6-MULTI-SESSION-COORD-001` (Layer 1 primitives).
