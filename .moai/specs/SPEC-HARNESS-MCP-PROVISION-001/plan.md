# Plan — SPEC-HARNESS-MCP-PROVISION-001

> Implementation plan for the `/moai project` Phase 3.6 MCP-provisioning insertion
> + the harness GENERATE optional artifact 6. Tier M (3 artifacts). Doc / config-only
> (markdown + yaml) — no Go code. Every touched skill/config file has a byte-identical
> template mirror; Template-First discipline governs every edit. Line-number citations
> are drift-prone — re-anchor by content token at run-phase.

## §A. Context

- **Branch / baseline**: `main` (verify HEAD at run-phase pre-flight; do not assume
  a plan-time SHA).
- **SPEC artifacts**: `.moai/specs/SPEC-HARNESS-MCP-PROVISION-001/{spec,plan,acceptance,progress}.md`.
- **Epic position**: SPEC 2 of the 3-SPEC "Project-Harness Pipeline" Epic.
  `depends_on: [SPEC-PROJECT-HARNESS-BRIDGE-001]` (Depends_on Pre-flight per Phase
  0.5 requires BRIDGE-001 `status: completed` before run-phase entry; BRIDGE-001 is
  currently `draft`, so this gate WILL block until BRIDGE-001 closes — expected).
- **Verified inventory** (research input — NOT to be re-investigated at plan
  authoring; re-verify by content token at run-phase):
  - `/moai project` router: `.claude/skills/moai/workflows/project.md` +
    `project/{mode-detection,codebase-analysis,doc-generation,meta-harness}.md`,
    each byte-mirrored under `internal/template/templates/.claude/skills/moai/workflows/...`.
  - Today there is NO MCP provisioning in the flow. The only tool-provisioning step
    is **Phase 3.5 (LSP detection / install)**. Phases run:
    `3.5 LSP → 3.7 auto dev-mode → 4.1a DB detection → 4 completion`. Phase 3.6 is
    the insertion slot (between 3.5 and 3.7) inside `project/doc-generation.md`.
  - `builder-harness` (`.claude/agents/moai/builder-harness.md`) ALREADY supports
    `artifact_type=mcp-server` → scaffolds `.mcp.json` entries (stdio / http / sse).
    This capability is NOT wired into harness generation.
  - The v4 harness Builder GENERATE phase (`harness-builder.md`) emits exactly 5
    artifact types (thin command / Runner JS / specialist agents / companion skills /
    manifest.json) — NO MCP.
  - **doctor / manifest verified this session (do NOT re-investigate)**: the v4
    manifest is decoded with plain `json.Unmarshal` at `internal/harness/applier.go:630`
    and `internal/cli/harness/doctor.go:130` — NEITHER uses `DisallowUnknownFields`.
    `internal/harness/v4manifest/validate.go` `Validate()` checks only the 8 required
    top-level fields + specialist enums + sprint_contract; it does NOT reject unknown
    fields. **Conclusion: an optional `mcp` block in `manifest.json` is TOLERATED
    (silently ignored) — `moai harness doctor` stays exit-0 with NO Go change.** Active
    validation of the `mcp` block would require adding `MCP *MCPBlock json:"mcp,omitempty"`
    to `internal/harness/v4manifest/types.go` + a `Validate()` branch — described here,
    NOT implemented (see clarification below).
  - Anthropic-verified MCP guidance: `.mcp.json` at repo root (checked in, per-user
    approval prompt), project scope; 3-5 servers max; vendor-maintained preferred
    (2026 CVE surge); `${VAR}` env-var expansion for secrets. Per-project-type matrix:
    web frontend = Playwright + Chrome DevTools (+ Figma Dev Mode for design→code);
    mobile = Maestro (then Appium); backend/DB = read-only Postgres + Context7;
    universal starter = GitHub + Context7 + Playwright.

### Open clarifications (resolve before Implementation Kickoff Approval)

- **[NEEDS CLARIFICATION: mcp-matrix config surface]** — Is
  `.moai/config/sections/mcp-matrix.yaml` a **new config SECTION merged into the Go
  config manager** (which would require a Go struct field + loader wiring, breaching
  the doc/config-only scope), OR a **standalone data resource** read directly by
  `project/doc-generation.md` as prose-context (no typed Go loader)? Default
  assumption if unresolved: **standalone data resource** — read directly by the
  workflow skill as prose-context, mirroring how BRIDGE-001 treats `harness-spec.yaml`
  ("consumers read it as prose-context via the workflow skills, not via a typed
  loader"). This keeps the SPEC doc/config-only. Confirm only if the matrix must be
  machine-validated by a Go loader.
- **[NEEDS CLARIFICATION: doctor manifest-mcp validate-vs-tolerate]** — Should
  `moai harness doctor` merely TOLERATE the optional manifest `mcp` block (current
  lenient `json.Unmarshal` already does this — zero Go change, AC-HMP-010 passes as a
  documented-tolerance check), OR ACTIVELY VALIDATE it (add
  `MCP *MCPBlock json:"mcp,omitempty"` to `internal/harness/v4manifest/types.go` +
  a `Validate()` branch — a small Go change, out of this SPEC's doc-only scope)?
  Default assumption if unresolved: **tolerate only** (no Go change); active
  validation is deferred to a follow-up Go SPEC. If active validation is desired,
  this SPEC's tier / scope changes (it becomes non-doc-only) — surface to the user.

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
- **B-SUBAGENT-BOUNDARY**: any delegated `builder-harness` step (artifact 6 scaffold)
  MUST NOT call AskUserQuestion — the per-server credential approval (REQ-HMP-006) and
  the overall approval (REQ-HMP-002) are ORCHESTRATOR-held. A subagent returns a
  blocker report; the orchestrator runs AskUserQuestion.
- **B-CREDENTIAL (no literal secrets)**: no written `.mcp.json` entry may inline a
  literal credential / token — env-var `${VAR}` form only (REQ-HMP-007). This binds
  BOTH the Phase 3.6 write path AND the artifact-6 fragment example.
- **B-ADDITIVE (no clobber)**: the `.mcp.json` write MUST merge into an existing file,
  never overwrite it. A pre-existing user `.mcp.json` (with unrelated servers) must
  survive the write (REQ-HMP-007).
- **B-MIRROR-PARITY (byte diff)**: after each edit, `diff` the local file against its
  template mirror — byte-identical required (AC-HMP-011). These files carry no
  internal-content tokens, so the mirror should be a clean byte copy.
- **B-BYTE-IDENTICAL-OMISSION**: when the harness PLAN declares no MCP need, the
  GENERATE output MUST be byte-identical to the current 5-artifact set (REQ-HMP-009).
  The artifact-6 extension is a CONDITIONAL branch, not an unconditional 6th write.
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

# 4. Confirm the harness GENERATE 5-artifact contract + the mcp-server capability
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
- The current 5-artifact harness GENERATE output when no MCP need is declared
  (byte-identical omission).
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
- An unconditional 6th artifact write (artifact 6 is CONDITIONAL on declared MCP need).
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
(template-neutrality guard + internal-content-leak test green; `moai harness doctor`
exit-0 tolerance check), E6 commit SHAs + push state, E7 blocker reports (never
AskUserQuestion). E3 coverage and E4 subagent-boundary grep are N/A (no Go code
touched); state them as N/A rather than fabricating output.

## §F. Milestones

### M1 — MCP matrix externalized to config (REQ-HMP-004)

1. Resolve `[NEEDS CLARIFICATION: mcp-matrix config surface]`; default = standalone
   data resource.
2. Create `internal/template/templates/.moai/config/sections/mcp-matrix.yaml` with the
   §D.1 schema (web-frontend / mobile / backend-db rows + universal_starter fallback;
   per-server `name` / `transport` / `install` / `vendor_maintained` /
   `requires_credentials`). Keep it 16-language neutral (keyed by project-type).
3. Mirror to the local `.moai/config/sections/mcp-matrix.yaml`; `make build`.
4. Exit: AC-HMP-004 grep-green (matrix file exists with web / mobile / backend rows);
   AC-HMP-011 byte-parity clean; AC-HMP-013 neutrality green.

### M2 — /moai project Phase 3.6 insertion (REQ-HMP-001/002/003/005/006/007)

1. In `project/doc-generation.md`, insert a **Phase 3.6** section BETWEEN Phase 3.5
   (LSP) and Phase 3.7 (dev-mode): (a) detect stack (reuse existing language /
   framework detection + `harness-spec.yaml` `external_systems` / `ui_surface`);
   (b) select recommended servers from `mcp-matrix.yaml` (fallback pointer, not a
   hardcoded copy); (c) cap at 3-5 servers + prefer vendor-maintained; (d) surface
   the selection via the ORCHESTRATOR AskUserQuestion (subagent never prompts);
   (e) require an EXPLICIT per-server AskUserQuestion for any credentialed server;
   (f) on approval, write `.mcp.json` at project scope — additively / idempotently
   (merge), with secrets in `${VAR}` env-var form (never a literal token).
2. Confirm the write target is the repo-root `.mcp.json` (NOT `.moai/specs/`).
3. Mirror to the template tree; `make build`.
4. Exit: AC-HMP-001 (Phase 3.6 heading ordered between 3.5 and 3.7),
   AC-HMP-002 (stack-detect + matrix reference), AC-HMP-003 (orchestrator approval +
   subagent-no-prompt), AC-HMP-005 (3-5 cap + vendor-maintained), AC-HMP-006
   (credential per-server approval + never-auto-write), AC-HMP-007 (additive-merge +
   `${VAR}` + no-literal-token) grep-green on both trees; AC-HMP-012 NO-SPEC clean.

### M3 — harness GENERATE optional artifact 6 (REQ-HMP-008/009/010)

1. In `harness-builder.md`, extend the GENERATE contract with an OPTIONAL **artifact 6**
   (`.mcp.json` fragment via `builder-harness` `artifact_type=mcp-server`).
2. Add the explicit conditional-emission clause: emit artifact 6 ONLY when the harness
   PLAN declares MCP needs (derived from `harness-spec.yaml` `external_systems` /
   `verification`); when no MCP need is declared, artifact 6 is OMITTED and GENERATE
   stays byte-identical to the current 5-artifact set.
3. Document the optional manifest `mcp` block (§D.3) as doctor-tolerant (lenient
   decoder; no Go change). Resolve `[NEEDS CLARIFICATION: doctor manifest-mcp
   validate-vs-tolerate]`; default = tolerate-only.
4. Mirror to the template tree; `make build`.
5. Exit: AC-HMP-008 (artifact-6 section present), AC-HMP-009 (conditional-emission +
   byte-identical omission clause), AC-HMP-010 (`moai harness doctor` exit-0 with
   optional mcp block / documented tolerance) grep-green on both trees.

### M4 — Template mirror parity + neutrality + doctor smoke (REQ-HMP-011)

1. Full byte-parity sweep: `diff` every touched local file against its template
   mirror — all byte-identical (AC-HMP-011).
2. `make build`; run the template-neutrality guard + internal-content-leak test
   (`go test ./internal/template/...`) — green (AC-HMP-013).
3. `moai harness doctor` smoke on a manifest carrying the optional `mcp` block —
   exit 0 (AC-HMP-010).
4. Whole-repo non-regression: `go build ./...` + `go test ./...` exit 0 (no Go code
   changed).
5. Exit: all 13 ACs PASS or documented PASS-WITH-DEBT.

## §G. Anti-Patterns (this SPEC)

- Inserting Phase 3.6 in the wrong host file or wrong position (must be BETWEEN 3.5
  LSP and 3.7 dev-mode in `doc-generation.md`) — breaks the ordering AC.
- Hardcoding the matrix rows in skill prose instead of `mcp-matrix.yaml` (REQ-HMP-004).
- Auto-writing a credentialed server without the per-server approval (REQ-HMP-006).
- Inlining a literal token into `.mcp.json` instead of `${VAR}` (REQ-HMP-007).
- Clobbering an existing `.mcp.json` instead of merging additively (REQ-HMP-007).
- Emitting artifact 6 unconditionally (must be conditional on declared MCP need;
  no-MCP output must stay byte-identical to the 5-artifact set — REQ-HMP-009).
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
  reused by artifact 6 (READ at run-phase M3).
- `internal/harness/v4manifest/{types,validate}.go` + `internal/cli/harness/doctor.go`
  — the manifest schema + doctor smoke gate; the lenient `json.Unmarshal` that
  tolerates the optional `mcp` block (no Go change).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier M
  Section A-E delegation template applies to every milestone delegation.
- CLAUDE.local.md §2 (Template-First) + §15 (16-language neutrality) + §25 (Template
  Internal-Content Isolation) — the neutrality / mirror discipline this SPEC's edits
  must respect.
