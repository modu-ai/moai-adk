# Plan — SPEC-HARNESS-MCP-PROVISION-001

> Implementation plan for the `/moai project` Phase 3.6 MCP-provisioning insertion
> + the harness GENERATE optional artifact 7. Tier M (3 artifacts). Doc / config-only
> (markdown + yaml) — no Go code. Every touched skill/config file has a byte-identical
> template mirror; Template-First discipline governs every edit. Line-number citations
> are drift-prone — re-anchor by content token at run-phase.

## §A. Context

- **Branch / baseline**: `main` (verify HEAD at run-phase pre-flight; do not assume
  a plan-time SHA).
- **SPEC artifacts**: `.moai/specs/SPEC-HARNESS-MCP-PROVISION-001/{spec,plan,acceptance,progress}.md`.
- **Epic position**: SPEC 2 of the 3-SPEC "Project-Harness Pipeline" Epic.
  `depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]` (Depends_on Pre-flight per Phase
  0.5 requires BRIDGE-001 `status: completed` before run-phase entry).
  **Observed state (measured 2026-07-11 at v0.1.3):**
  `grep -n "^status:" .moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/spec.md` → `status: completed`.
  The gate is therefore **SATISFIED — it does NOT block run-phase entry.**
  (This CORRECTS the v0.1.0-v0.1.2 claim that BRIDGE-001 was "currently `draft`, so this
  gate WILL block until BRIDGE-001 closes". That was an **unobserved state claim** — the
  dependency had already closed. Per `verification-claim-integrity.md` §1.1, re-verify by
  execution at run-phase pre-flight rather than trusting this line.)
- **Verified inventory** (research input — NOT to be re-investigated at plan
  authoring; re-verify by content token at run-phase):
  - `/moai project` **ROUTER**: `.claude/skills/moai/workflows/project.md` +
    `project/{mode-detection,codebase-analysis,doc-generation,meta-harness}.md`,
    each byte-mirrored under `internal/template/templates/.claude/skills/moai/workflows/...`.
    The router declares the canonical phase sequence in TWO surfaces — a **Phase Routing
    Table** (one `| Phase N: … | \`project/<sub-skill>.md\` | <description> |` row per
    phase) and an **Invocation Flow diagram** (one `├─ Phase N: …` line per phase).
    **The router is an M2 EDIT TARGET** (AC-HMP-016) — see §F M2.
  - **Measured router drift (2026-07-11)**: `grep -c -F "Phase 3.2" project.md` → **0**.
    Phase 3.2 (`harness-spec.yaml` emission) was added to `doc-generation.md` by
    SPEC-PROJECT-HARNESS-BRIDGE-001 and **never registered in the router**. The same
    file-pair drift this SPEC must avoid has already shipped once — M2 backfills it.
  - Today there is NO MCP provisioning in the flow. The only tool-provisioning step
    is **Phase 3.5 (LSP detection / install)**. Phases run:
    `3.5 LSP → 3.7 auto dev-mode → 4.1a DB detection → 4 completion`. Phase 3.6 is
    the insertion slot (between 3.5 and 3.7) inside `project/doc-generation.md`, AND
    must be registered in BOTH router surfaces above.
  - `builder-harness` (`.claude/agents/moai/builder-harness.md`) ALREADY supports
    `artifact_type=mcp-server` → scaffolds `.mcp.json` entries (stdio / http / sse).
    This capability is NOT wired into harness generation.
  - The v4 harness Builder GENERATE phase (`harness-builder.md`) documents **5 base
    artifact types** (thin command / Runner JS / specialist agents / companion skills /
    manifest.json) — NO MCP. Per the Epic canonical artifact order the contract grows
    to 5 base + verify skill (artifact 6, mandatory — SPEC-HARNESS-VERIFY-PROMOTE-001)
    + optional MCP fragment (artifact 7, this SPEC); this SPEC reconciles the
    "exactly 5" prose so the bare count is not left uncontextualized (AC-HMP-015).
  - **doctor / manifest verified this session (do NOT re-investigate)**: the v4
    manifest is decoded with plain `json.Unmarshal` at `internal/harness/applier.go:630`
    and `internal/cli/harness/doctor.go:130` — NEITHER uses `DisallowUnknownFields`.
    `internal/harness/v4manifest/validate.go` `Validate()` checks only the 8 required
    top-level fields + specialist enums + sprint_contract; it does NOT reject unknown
    fields. **Conclusion: an optional `mcp` block in `manifest.json` is TOLERATED
    (silently ignored) — `moai harness doctor` stays exit-0 with NO Go change.** Active
    validation of the `mcp` block would require adding `MCP *MCPBlock json:"mcp,omitempty"`
    to `internal/harness/v4manifest/types.go` + a `Validate()` branch — described here,
    NOT implemented (TOLERATE-ONLY decision; see Resolved clarifications below).
  - Anthropic-verified MCP guidance: `.mcp.json` at repo root (checked in, per-user
    approval prompt), project scope; 3-5 servers max; vendor-maintained preferred
    (2026 CVE surge); `${VAR}` env-var expansion for secrets. Per-project-type matrix:
    web frontend = Playwright + Chrome DevTools (+ Figma Dev Mode for design→code);
    mobile = Maestro (then Appium); backend/DB = read-only Postgres + Context7;
    universal starter = GitHub + Context7 + Playwright.

### Resolved clarifications (both settled at plan-phase — no open markers remain)

Both plan-phase clarifications are RESOLVED with the values below; the resolution is
recorded in progress.md §E.1. No unresolved markers remain (a re-audit MP-7 marker
grep on plan.md returns 0).

- **mcp-matrix config surface → RESOLVED: standalone DATA RESOURCE.**
  `.moai/config/sections/mcp-matrix.yaml` is a standalone template data resource
  (authored at `internal/template/templates/.moai/config/sections/mcp-matrix.yaml`,
  mirrored to the local copy) read by `project/doc-generation.md` as prose-context —
  NOT a new config SECTION merged into the Go config manager, and it adds NO Go struct
  field or typed loader. This mirrors how BRIDGE-001 treats `harness-spec.yaml`
  (consumed as prose-context via the workflow skills, not via a typed loader) and keeps
  this SPEC doc/config-only.
- **doctor manifest-mcp validate-vs-tolerate → RESOLVED: TOLERATE-ONLY (zero Go
  change).** The manifest decoders use plain `json.Unmarshal` with no
  `DisallowUnknownFields` (`internal/cli/harness/doctor.go`,
  `internal/harness/applier.go`) and `internal/harness/v4manifest/validate.go` checks
  only the required fields, so an optional `mcp` block is silently tolerated with no Go
  change. AC-HMP-010 encodes this as a documented-tolerance grep + the regression guard
  `grep -c "DisallowUnknownFields" internal/harness/v4manifest/*.go  # expect 0`. Active
  validation (adding `MCP *MCPBlock` to `v4manifest/types.go` + a `Validate()` branch)
  is explicitly OUT OF SCOPE, deferred to a follow-up Go SPEC.

## §B. Known Issues (filtered, Tier M — doc/config-only)

- **B-TF (Template-First)**: every touched skill/config file exists in BOTH trees
  (local `.claude/...` + `internal/template/templates/.claude/...`; and
  `.moai/config/sections/mcp-matrix.yaml` local + template mirror). Edit the template
  FIRST, mirror to local, then `make build`. An unmirrored local edit fails
  AC-HMP-011 (byte-parity diff).
- **B-NEUTRAL (template neutrality + 16-language)**: the template tree MUST NOT gain
  internal SPEC IDs, dates, or SHAs. Do NOT paste `SPEC-HARNESS-MCP-PROVISION-001`
  or this SPEC's dates into any `internal/template/templates/**` file. The matrix
  MUST stay 16-language neutral — it is keyed by project-TYPE (web-frontend / mobile /
  backend-db), not by a privileged language; the neutrality CI guard
  (`template-neutrality-check.yaml` + `internal_content_leak_test.go`) must stay green
  (AC-HMP-013).
- **B-NOSPEC (scope guard)**: the project workflow must never write to
  `.moai/specs/**`. `.mcp.json` lives at the repo root; `mcp-matrix.yaml` under
  `.moai/config/sections/`. Do NOT introduce any `.moai/specs/` write path
  (AC-HMP-012).
- **B-SUBAGENT-BOUNDARY**: any delegated `builder-harness` step (artifact 7 scaffold)
  MUST NOT call AskUserQuestion — the per-server credential approval (REQ-HMP-006) and
  the overall approval (REQ-HMP-002) are ORCHESTRATOR-held. A subagent returns a
  blocker report; the orchestrator runs AskUserQuestion.
- **B-CREDENTIAL (no literal secrets)**: no written `.mcp.json` entry may inline a
  literal credential / token — env-var `${VAR}` form only (REQ-HMP-007). This binds
  BOTH the Phase 3.6 write path AND the artifact-7 fragment example.
- **B-ADDITIVE (no clobber)**: the `.mcp.json` write MUST merge into an existing file,
  never overwrite it. A pre-existing user `.mcp.json` (with unrelated servers) must
  survive the write (REQ-HMP-007).
- **B-MIRROR-PARITY (byte diff)**: after each edit, `diff` the local file against its
  template mirror — byte-identical required (AC-HMP-011). These files carry no
  internal-content tokens, so the mirror should be a clean byte copy.
- **B-BYTE-IDENTICAL-OMISSION**: when the harness PLAN declares no MCP need, the
  GENERATE output MUST be byte-identical to the without-artifact-7 baseline
  (REQ-HMP-009). The artifact-7 extension is a CONDITIONAL branch, not an
  unconditional additional write.
- **B-DOCTOR-TOLERANCE**: do NOT add `DisallowUnknownFields` anywhere; the doctor
  tolerance relies on the current lenient decoder. Adding strictness would break
  AC-HMP-010 and every grandfathered manifest.

## §C. Pre-flight Checklist (run before any change)

```bash
# 1. Baseline
git branch --show-current && git rev-parse HEAD

# 2. Depends_on gate — BRIDGE-001 must be completed (Phase 0.5 pre-flight)
grep -n "^status:" .moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/spec.md   # expect: completed (else run-gate blocks)

# 3. Locate the Phase 3.5 / 3.7 anchors + confirm no existing 3.6 / MCP step
grep -n -i "Phase 3.5\|Phase 3.7\|LSP\|dev-mode\|mcp" .claude/skills/moai/workflows/project/doc-generation.md | head

# 4. Confirm the harness GENERATE base-artifact contract (5 base + verify skill 6) + the mcp-server capability
grep -n -i "GENERATE\|artifact\|manifest.json\|Runner\|companion skill" .claude/skills/moai/workflows/harness-builder.md | head
grep -n -i "mcp-server\|artifact_type\|.mcp.json\|transport" .claude/agents/moai/builder-harness.md | head

# 5. Confirm both trees exist for every target file (Template-First)
for f in project/doc-generation.md harness-builder.md; do
  ls ".claude/skills/moai/workflows/$f" "internal/template/templates/.claude/skills/moai/workflows/$f"
done
ls .moai/config/sections/ internal/template/templates/.moai/config/sections/   # confirm the sections dir exists in both trees

# 6. Doctor tolerance baseline (lenient decoder — must remain lenient)
grep -n "DisallowUnknownFields" internal/harness/applier.go internal/cli/harness/doctor.go || echo "lenient decoder confirmed (good)"

# 7. NO-SPEC scope baseline (must remain 0 in touched sections after edit)
grep -rn "\.moai/specs/" .claude/skills/moai/workflows/project/ || echo "no .moai/specs write path (good)"
```

## §D. Constraints (DO NOT VIOLATE)

**PRESERVE list (assert STILL EXISTS at exit)**:
- The `/moai project` NO-SPEC scope guard — no `.moai/specs/**` write path.
- The pre-MCP harness GENERATE output when no MCP need is declared — artifact 7
  omitted, byte-identical to the without-artifact-7 baseline.
- The `builder-harness` `artifact_type=mcp-server` internals (reused, unchanged).
- The `harness-spec.yaml` schema + adaptive interview (owned by BRIDGE-001; consumed,
  not modified).
- The lenient `json.Unmarshal` manifest decoder (doctor tolerance depends on it).

**Forbidden**:
- Writing internal SPEC IDs / dates / SHAs into `internal/template/templates/**`.
- Local-only edits not mirrored to the template tree (or vice versa).
- Any `.moai/specs/` write path in the project workflow.
- Inlining a literal credential / token into `.mcp.json` (env-var `${VAR}` only).
- Clobbering an existing `.mcp.json` (must merge additively).
- A subagent (`builder-harness`) calling AskUserQuestion — approval is
  orchestrator-held.
- An unconditional artifact-7 write (artifact 7 is CONDITIONAL on declared MCP need).
- Adding `DisallowUnknownFields` or a Go `MCP` struct field (out of doc-only scope;
  deferred per the clarification).
- Hardcoding the matrix rows in skill prose beyond a fallback pointer (REQ-HMP-004).

**Required**: Conventional Commits (`feat(SPEC-HARNESS-MCP-PROVISION-001): M{N} …`
for run-phase; the plan-phase artifact commit uses the `feat(` prefix per the
plan-commit-subject lesson), `🗿 MoAI` trailer, specific-path `git add`.

## §E. Self-Verification Deliverables

Per the manager-develop prompt template §E (E1-E7), each milestone completion report
carries: E1 AC PASS/FAIL matrix (verbatim command output), E2 build result
(`make build` exit 0 — doc/config-only SPEC, so the cross-platform Go build is a
non-regression check, not a feature check), E5 lint / neutrality
(template-neutrality guard + internal-content-leak test green; doctor tolerance =
documented-tolerance grep + `DisallowUnknownFields == 0` regression guard, NOT a live
repo-wide doctor smoke), E6 commit SHAs + push state, E7 blocker reports (never
AskUserQuestion). E3 coverage and E4 subagent-boundary grep are N/A (no Go code
touched); state them as N/A rather than fabricating output.

## §F. Milestones

### M1 — MCP matrix externalized to config (REQ-HMP-004)

1. Apply the resolved decision (see §A Resolved clarifications): `mcp-matrix.yaml` is a
   standalone data resource read as prose-context — NO Go config section / loader.
2. Create `internal/template/templates/.moai/config/sections/mcp-matrix.yaml` with the
   §D.1 schema (web-frontend / mobile / backend-db rows + universal_starter fallback;
   per-server `name` / `transport` / `install` / `vendor_maintained` /
   `requires_credentials`). Keep it 16-language neutral (keyed by project-type).
3. Mirror to the local `.moai/config/sections/mcp-matrix.yaml`; `make build`.
4. Exit: AC-HMP-004 grep-green (matrix file exists with web / mobile / backend rows);
   AC-HMP-011 byte-parity clean; AC-HMP-013 neutrality green.

### M2 — /moai project Phase 3.6 insertion + ROUTER registration (REQ-HMP-001/002/003/005/006/007)

**Edit targets (4 files — 2 logical files × 2 trees; ALL must be mirrored):**
- `.claude/skills/moai/workflows/project/doc-generation.md` (the sub-skill that IMPLEMENTS Phase 3.6)
- `internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md`
- `.claude/skills/moai/workflows/project.md` (**the ROUTER that DECLARES Phase 3.6**)
- `internal/template/templates/.claude/skills/moai/workflows/project.md`

**M2a — implement Phase 3.6 in the sub-skill (`project/doc-generation.md`)**

1. Insert a **Phase 3.6** section BETWEEN Phase 3.5 (LSP) and Phase 3.7 (dev-mode):
   (a) detect stack (reuse existing language / framework detection + `harness-spec.yaml`
   `external_systems` / `ui_surface`); (b) select recommended servers from
   `mcp-matrix.yaml` (fallback pointer, not a hardcoded copy); (c) cap at 3-5 servers +
   prefer vendor-maintained; (d) surface the selection via the ORCHESTRATOR
   AskUserQuestion (subagent never prompts — it returns a blocker report);
   (e) require an EXPLICIT per-server AskUserQuestion for any credentialed server, and
   never auto-write one; (f) on approval, write `.mcp.json` at project scope —
   additively / idempotently (merge, never clobber), with secrets in `${VAR}` env-var
   form (never a literal token).
2. Confirm the write target is the repo-root `.mcp.json` (NOT `.moai/specs/`).
   (AC-HMP-012's NO-SPEC guard is prohibition-aware and explicitly permits this
   `(NOT \`.moai/specs/\`)` phrasing — it filters prohibitions, not write paths.)

**M2b — register Phase 3.6 in the ROUTER (`project.md`) — REQ-HMP-001 / AC-HMP-016**

3. The router declares the canonical phase sequence in TWO surfaces. Phase 3.6 MUST be
   added to BOTH, or the router keeps documenting `3.5 → 3.7` with Phase 3.6 nowhere in
   it (the reachability gap this SPEC exists to close, reproduced one file over):
   - **Phase Routing Table** — add a row AFTER the `| Phase 3.5: Dev Environment Check |`
     row and BEFORE the `| Phase 3.7: Dev Methodology Config |` row, matching the
     existing 3-column shape:
     `| Phase 3.6: MCP Server Provisioning | \`project/doc-generation.md\` | <one-line description> |`
   - **Invocation Flow diagram** — add a line AFTER `├─ Phase 3.5: LSP check` and BEFORE
     `├─ Phase 3.7: Dev mode config`, matching the existing tree-branch shape:
     `├─ Phase 3.6: MCP provisioning`
4. **Phase 3.2 BACKFILL (user-approved scope addition).** `grep -c -F "Phase 3.2" project.md`
   measures **0** — Phase 3.2 (`harness-spec.yaml` emission) was added to
   `doc-generation.md` by SPEC-PROJECT-HARNESS-BRIDGE-001 and never registered in the
   router. In the SAME edit, backfill it: a routing-table row after the Phase 3.1 row and
   before the Phase 3.3 row, plus a diagram line after `├─ Phase 3.1: Audit (conditional)`
   and before `├─ Phase 3.3: Codemaps`.
   [HARD] **READ `project/doc-generation.md` § "Phase 3.2: harness-spec.yaml Emission"
   before writing its row — describe what it ACTUALLY does; do NOT invent a description.**
   Verified behavior (2026-07-11): it emits the machine-readable
   `.moai/project/harness-spec.yaml` artifact by READING the answers already recorded in
   `.moai/project/interview.md` and mapping them onto an 8-field schema; it runs
   automatically after Phase 3 with NO user interaction and does NOT re-interview;
   re-run semantics are OVERWRITE.
5. Mirror BOTH files to the template tree; `make build`.
6. Exit: AC-HMP-001 (Phase 3.6 heading ordered between 3.5 and 3.7),
   AC-HMP-002 (stack-detect + matrix reference), AC-HMP-003 (orchestrator approval +
   subagent-no-prompt), AC-HMP-005 (3-5 cap + vendor-maintained), AC-HMP-006
   (credential per-server approval + never-auto-write), AC-HMP-007 (additive-merge +
   `${VAR}` + no-literal-token), AC-HMP-014 (write-on-approval at project scope —
   REQ-HMP-003), **AC-HMP-016 (router: Phase 3.6 in routing table + flow diagram, ordered
   3.5 < 3.6 < 3.7, pointing at `doc-generation.md`; Phase 3.2 backfilled)** grep-green on
   both trees; AC-HMP-012 NO-SPEC clean.

### M3 — harness GENERATE optional artifact 7 (REQ-HMP-008/009/010)

1. In `harness-builder.md`, extend the GENERATE contract with an OPTIONAL **artifact 7**
   (`.mcp.json` fragment via `builder-harness` `artifact_type=mcp-server`) as a new
   `### Artifact 7 — ...` section under `## GENERATE Output Contract`.

   **Reconcile the stale artifact-count prose at ALL SIX sites** (measured 2026-07-11 via
   `grep -n -E "5 artifacts?|the 5 artifact" harness-builder.md`). Reconciling only the
   body sentence would leave the H2 heading announcing "the 5 artifact types" directly
   above a section listing seven — REQ-HMP-008's intent unmet:

   | # | Site (content-token anchor — line numbers drift) | Current text |
   |---|--------------------------------------------------|--------------|
   | 1 | frontmatter `description:` block | "contract (the 5 artifact types M3/M4 consume)" |
   | 2 | `### Phase 3 — GENERATE [...]` heading | "orchestrator fan-out emits 5 artifact types" |
   | 3 | `**Primitive**:` line under Phase 3 | "emit the 5 artifact types in § GENERATE Output Contract" |
   | 4 | "On Proceed, the orchestrator fans out..." line | "emit the **5 artifacts**" (note: no "types") |
   | 5 | `## GENERATE Output Contract (the 5 artifact types)` H2 heading | "(the 5 artifact types)" |
   | 6 | first body sentence of that section | "The GENERATE phase emits **exactly 5 artifact types**." |

   Canonical reconciled order: **5 base + verify skill (artifact 6, mandatory — owned by
   SPEC-HARNESS-VERIFY-PROMOTE-001) + optional MCP fragment (artifact 7, this SPEC)**.
   Sites 5 and 6 are pinned by AC-HMP-015 reverse deltas (`(the 5 artifact types)` → 0
   and `exactly 5 artifact types` → 0). Sites 1-4 are not individually AC-pinned but are
   in scope for REQ-HMP-008 — reconcile them in the same edit.

   [HARD] Do NOT rename the `## GENERATE Output Contract` H2 prefix — only its
   parenthetical suffix. The AC-HMP-015 `goc()` extractor prefix-matches
   `^## GENERATE Output Contract`; changing the prefix would break it.
2. Add the explicit conditional-emission clause: emit artifact 7 ONLY when the harness
   PLAN declares MCP needs (derived from `harness-spec.yaml` `external_systems` /
   `verification`); when no MCP need is declared, artifact 7 is OMITTED and GENERATE
   stays byte-identical to the without-artifact-7 baseline.
3. Document the optional manifest `mcp` block (§D.3) as doctor-tolerant. Apply the
   resolved decision (see §A Resolved clarifications): TOLERATE-ONLY, no Go change
   (lenient decoder); no `DisallowUnknownFields` and no Go `MCP` struct field is added.
4. Mirror to the template tree; `make build`.
5. Exit: AC-HMP-008 (artifact-7 section present), AC-HMP-009 (conditional-emission +
   byte-identical omission clause), AC-HMP-010 (documented-tolerance grep +
   `DisallowUnknownFields == 0` regression guard), AC-HMP-015 ("exactly 5" prose
   reconciled) grep-green on both trees.

### M4 — Template mirror parity + neutrality + doctor smoke (REQ-HMP-011)

1. Full byte-parity sweep: `diff` every touched local file against its template
   mirror — all byte-identical (AC-HMP-011). **Four logical files** at v0.1.3:
   `mcp-matrix.yaml`, `project/doc-generation.md`, `harness-builder.md`, and
   **`project.md` (the router — added at v0.1.3 per AC-HMP-016)**. Existence is checked
   BEFORE `diff` (parity between two absent files is vacuous).
2. `make build`; run the template-neutrality guard + internal-content-leak test
   (`go test ./internal/template/...`) — green (AC-HMP-013).
3. Doctor tolerance (AC-HMP-010): documented-tolerance grep in `harness-builder.md`
   + `grep -c "DisallowUnknownFields" internal/harness/v4manifest/*.go` == 0 (and the
   decode sites `internal/harness/applier.go` / `internal/cli/harness/doctor.go` == 0)
   — deterministic grep+guard, no live repo-wide doctor run.
4. Whole-repo non-regression: `go build ./...` + `go test ./...` exit 0 (no Go code
   changed).
5. Exit: all 15 ACs PASS or documented PASS-WITH-DEBT.

## §G. Anti-Patterns (this SPEC)

- Inserting Phase 3.6 in the wrong host file or wrong position (must be BETWEEN 3.5
  LSP and 3.7 dev-mode in `doc-generation.md`) — breaks the ordering AC.
- Hardcoding the matrix rows in skill prose instead of `mcp-matrix.yaml` (REQ-HMP-004).
- Auto-writing a credentialed server without the per-server approval (REQ-HMP-006).
- Inlining a literal token into `.mcp.json` instead of `${VAR}` (REQ-HMP-007).
- Clobbering an existing `.mcp.json` instead of merging additively (REQ-HMP-007).
- Emitting artifact 7 unconditionally (must be conditional on declared MCP need;
  no-MCP output must stay byte-identical to the without-artifact-7 baseline — REQ-HMP-009).
- A `builder-harness` subagent calling AskUserQuestion for approval — approval is
  orchestrator-held (REQ-HMP-002).
- Adding `DisallowUnknownFields` / a Go `MCP` struct field (out of doc-only scope;
  breaks doctor tolerance / expands scope — REQ-HMP-010).
- Making the matrix language-biased (must be project-type-keyed, 16-lang neutral).
- Pasting `SPEC-HARNESS-MCP-PROVISION-001` / this SPEC's dates into the template tree
  — fails neutrality.

## §H. Cross-References

- `spec.md` §D — the `mcp-matrix.yaml` + `.mcp.json` + manifest `mcp` block schemas.
- `acceptance.md` — AC matrix (SSOT), GWT scenarios, quality gates, DoD.
- `.moai/specs/SPEC-PROJECT-HARNESS-BRIDGE-001/spec.md` §D — the `harness-spec.yaml`
  schema this SPEC consumes (`external_systems` / `ui_surface` / `verification`).
- `.claude/agents/moai/builder-harness.md` — the `artifact_type=mcp-server` capability
  reused by artifact 7 (READ at run-phase M3).
- `internal/harness/v4manifest/{types,validate}.go` + `internal/cli/harness/doctor.go`
  — the manifest schema + doctor smoke gate; the lenient `json.Unmarshal` that
  tolerates the optional `mcp` block (no Go change).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier M
  Section A-E delegation template applies to every milestone delegation.
- CLAUDE.local.md §2 (Template-First) + §15 (16-language neutrality) + §25 (Template
  Internal-Content Isolation) — the neutrality / mirror discipline this SPEC's edits
  must respect.
