# Acceptance Criteria — SPEC-HARNESS-MCP-PROVISION-001

> SSOT for the AC matrix. 13 ACs covering all 11 REQs (100% AC→REQ coverage).
> Verification conventions: (1) every grep is anchored to a content token, not a
> line number; (2) preservation / NO-WRITE ACs assert absence explicitly;
> (3) template↔local parity is a byte `diff` check on each touched file; (4) this is
> a doc/config-only SPEC, so ACs are grep / diff / `make build` / `moai harness doctor`
> based (no Go test / coverage ACs).

## §A. Given-When-Then Scenarios

### GWT-1 — Phase 3.6 provisions MCP between LSP and dev-mode

- **Given** a `/moai project` run whose Phase 3.5 (LSP) has completed and whose
  `harness-spec.yaml` declares `ui_surface: has-ui`,
- **When** the flow reaches the new Phase 3.6,
- **Then** it detects the web-frontend stack, selects recommended servers from
  `mcp-matrix.yaml` (Playwright + Chrome DevTools), surfaces them via the
  orchestrator AskUserQuestion, and — on approval — writes them into `.mcp.json` at
  the repo root; Phase 3.7 (dev-mode) then runs after Phase 3.6.

### GWT-2 — Credentialed server requires explicit per-server approval

- **Given** a backend-db stack whose matrix recommendation includes a credentialed
  server (read-only Postgres),
- **When** Phase 3.6 prepares to write that server,
- **Then** it requires an EXPLICIT per-server AskUserQuestion approval before writing
  it, and — when written — expresses the secret as `${DATABASE_URL}` (env-var form),
  never as an inlined literal token.

### GWT-3 — .mcp.json write is additive, never a clobber

- **Given** a repo that already has a `.mcp.json` with an unrelated user server,
- **When** Phase 3.6 writes the selected servers,
- **Then** the selected servers are MERGED into the existing `.mcp.json` (the
  pre-existing user server survives), and no file is written under `.moai/specs/**`.

### GWT-4 — Harness generation emits MCP fragment only when declared

- **Given** a harness whose PLAN declares MCP needs (from `harness-spec.yaml`
  `external_systems`),
- **When** the harness Builder runs GENERATE,
- **Then** it emits the OPTIONAL artifact 6 (`.mcp.json` fragment via
  `artifact_type=mcp-server`); and for a harness PLAN that declares NO MCP need,
  GENERATE omits artifact 6 and the output stays byte-identical to the 5-artifact set.

## §B. AC ↔ REQ Mapping

| AC | REQ | Title |
|----|-----|-------|
| AC-HMP-001 | REQ-HMP-001 | Phase 3.6 heading present, ordered between Phase 3.5 and Phase 3.7 |
| AC-HMP-002 | REQ-HMP-001 | Phase 3.6 detects stack + references the MCP matrix |
| AC-HMP-003 | REQ-HMP-002 | orchestrator-held AskUserQuestion approval + subagent-no-prompt clause |
| AC-HMP-004 | REQ-HMP-004 | mcp-matrix.yaml exists with web / mobile / backend rows |
| AC-HMP-005 | REQ-HMP-005 | 3-5 server cap + vendor-maintained preference present |
| AC-HMP-006 | REQ-HMP-006 | credentialed server per-server approval + never-auto-write |
| AC-HMP-007 | REQ-HMP-007 | .mcp.json additive/merge + project-scope + `${VAR}` + no-literal-token |
| AC-HMP-008 | REQ-HMP-008 | harness-builder.md artifact-6 section present |
| AC-HMP-009 | REQ-HMP-009 | conditional-emission clause (emit iff MCP need / byte-identical omission) |
| AC-HMP-010 | REQ-HMP-010 | moai harness doctor tolerates optional manifest mcp block (exit 0) |
| AC-HMP-011 | REQ-HMP-011 | template↔local byte-parity on every touched file |
| AC-HMP-012 | REQ-HMP-011 | NO-SPEC guard: no .moai/specs/ write path in project flow |
| AC-HMP-013 | REQ-HMP-011 | template neutrality + internal-content-leak guard green + 16-lang neutral matrix |

## §C. Verification Commands (per AC)

### AC-HMP-001 (REQ-HMP-001) — Phase 3.6 present and ordered

```bash
grep -c -i "Phase 3.6" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
# ordering: 3.5 line < 3.6 line < 3.7 line
grep -n -i "Phase 3.5\|Phase 3.6\|Phase 3.7" .claude/skills/moai/workflows/project/doc-generation.md
```

Expected: a Phase 3.6 heading exists and its line number falls between Phase 3.5 and
Phase 3.7 (inserted between LSP and dev-mode).

### AC-HMP-002 (REQ-HMP-001) — stack detection + matrix reference

```bash
grep -c -i "mcp-matrix\|mcp matrix\|recommendation matrix" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c -i "external_systems\|ui_surface\|stack\|framework detection" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
```

Expected: Phase 3.6 references the MCP matrix AND reuses the stack signal
(`external_systems` / `ui_surface` / language-framework detection).

### AC-HMP-003 (REQ-HMP-002) — orchestrator approval + subagent no-prompt

```bash
grep -c -i "AskUserQuestion\|orchestrator.*approv\|approv.*orchestrator" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c -i "subagent.*not.*prompt\|not prompt.*user\|blocker report\|orchestrator-held" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
```

Expected: approval routes through the orchestrator AskUserQuestion; the subagent-no-prompt
boundary is stated.

### AC-HMP-004 (REQ-HMP-004) — matrix config file exists with rows

```bash
ls .moai/config/sections/mcp-matrix.yaml   # expect: file exists
grep -c -i "web-frontend\|web frontend" .moai/config/sections/mcp-matrix.yaml   # expect >= 1
grep -c -i "mobile" .moai/config/sections/mcp-matrix.yaml                        # expect >= 1
grep -c -i "backend-db\|backend" .moai/config/sections/mcp-matrix.yaml           # expect >= 1
# skill prose carries at most a pointer, not a hardcoded duplicate of the rows:
grep -c -i "mcp-matrix.yaml" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1 (fallback pointer)
```

Expected: `mcp-matrix.yaml` exists with web / mobile / backend rows; the skill points
to it rather than duplicating the matrix.

### AC-HMP-005 (REQ-HMP-005) — 3-5 cap + vendor-maintained

```bash
grep -c -i "3-5\|3 to 5\|three.*five\|maximum.*5\|max.*server" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c -i "vendor-maintained\|vendor maintained\|vendor.maintained" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
```

Expected: the 3-5 server cap AND the vendor-maintained preference are both stated in
Phase 3.6.

### AC-HMP-006 (REQ-HMP-006) — credential per-server approval, never auto-write

```bash
grep -c -i "credential\|token\|requires_credentials\|secret" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c -i "per-server\|explicit.*approv\|never auto-write\|not auto-write\|MUST NOT auto" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
```

Expected: a credentialed server requires an explicit per-server approval and is never
auto-written.

### AC-HMP-007 (REQ-HMP-007) — additive merge + project scope + env-var secrets

```bash
grep -c -i "additive\|merge\|idempotent\|not clobber\|never clobber" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c "\.mcp.json" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1 (repo-root project scope)
grep -c -i "\${.*}\|env-var\|env var\|environment variable\|never.*literal" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1
grep -c "\.moai/specs/.*mcp" .claude/skills/moai/workflows/project/doc-generation.md   # expect 0 (never under specs)
```

Expected: `.mcp.json` write is additive/merge at project scope; secrets in `${VAR}`
form; never inlined; never under `.moai/specs/`.

### AC-HMP-008 (REQ-HMP-008) — harness artifact-6 section present

```bash
grep -c -i "artifact 6\|artifact-6\|sixth artifact\|mcp.*fragment\|artifact_type=mcp-server\|mcp-server" .claude/skills/moai/workflows/harness-builder.md   # expect >= 1
```

Expected: `harness-builder.md` GENERATE documents the optional artifact 6 (`.mcp.json`
fragment via `artifact_type=mcp-server`).

### AC-HMP-009 (REQ-HMP-009) — conditional emission + byte-identical omission

```bash
grep -c -i "optional\|only when\|declared.*mcp\|mcp.*need\|omit\|byte-identical\|when no MCP" .claude/skills/moai/workflows/harness-builder.md   # expect >= 1
```

Expected: artifact 6 is emitted ONLY when the harness PLAN declares MCP needs; when it
does not, artifact 6 is omitted and output stays byte-identical to the 5-artifact set.

### AC-HMP-010 (REQ-HMP-010) — doctor tolerates optional manifest mcp block

```bash
# Documented-tolerance check (default = tolerate-only, no Go change):
grep -c -i "tolerate\|optional.*mcp\|mcp.*block\|lenient\|doctor" .claude/skills/moai/workflows/harness-builder.md   # expect >= 1
# Smoke: doctor exits 0 on a project whose manifest carries the optional mcp block.
# (Run against a sandbox/fixture harness manifest containing an "mcp" block.)
go run ./cmd/moai harness doctor 2>&1; echo "exit=$?"   # expect exit=0 (mcp block tolerated by lenient json.Unmarshal)
# Confirm the decoder stays lenient (no strictness regression):
grep -c "DisallowUnknownFields" internal/harness/applier.go internal/cli/harness/doctor.go   # expect 0
```

Expected: `moai harness doctor` stays exit-0 with an optional `mcp` block; no
`DisallowUnknownFields` was introduced. (If the M3 clarification resolves to
active-validation, this AC is re-scoped to a Go validation check and the SPEC tier is
revisited — record the decision in progress.md §E.2.)

### AC-HMP-011 (REQ-HMP-011) — template↔local byte-parity

```bash
diff -q .moai/config/sections/mcp-matrix.yaml \
  internal/template/templates/.moai/config/sections/mcp-matrix.yaml && echo "PARITY OK: mcp-matrix.yaml" || echo "DRIFT"
for f in project/doc-generation.md harness-builder.md; do
  diff -q ".claude/skills/moai/workflows/$f" "internal/template/templates/.claude/skills/moai/workflows/$f" \
    && echo "PARITY OK: $f" || echo "DRIFT: $f"
done
# expect: PARITY OK for every touched file (byte-identical mirror)
```

Expected: every touched file is byte-identical between the local and template trees.

### AC-HMP-012 (REQ-HMP-011) — NO-SPEC scope guard

```bash
grep -rn "\.moai/specs/" .claude/skills/moai/workflows/project/doc-generation.md | wc -l   # expect 0 (no specs write path in Phase 3.6)
grep -c "\.mcp.json" .claude/skills/moai/workflows/project/doc-generation.md   # expect >= 1 (.mcp.json at repo root, not specs)
```

Expected: 0 `.moai/specs/` references in the Phase 3.6 section; the artifact target is
the repo-root `.mcp.json`.

### AC-HMP-013 (REQ-HMP-011) — neutrality + 16-language

```bash
make build ; echo "exit=$?"                                  # expect 0
go test ./internal/template/... 2>&1 | tail -2               # expect ok (neutrality + internal-content-leak guards)
grep -rn "SPEC-HARNESS-MCP-PROVISION-001" internal/template/templates/ | wc -l   # expect 0 (no internal SPEC ID leaked)
# 16-language neutrality: matrix keyed by project-TYPE, not a privileged language:
grep -c -i "web-frontend\|mobile\|backend" internal/template/templates/.moai/config/sections/mcp-matrix.yaml   # expect >= 1 (type-keyed)
grep -c -i "PRIMARY LANGUAGE\|primary: go\|primary: python" internal/template/templates/.moai/config/sections/mcp-matrix.yaml   # expect 0 (no privileged language)
```

Expected: `make build` clean; template guards green; no internal SPEC ID / date / SHA
in the template tree; matrix is project-type-keyed (16-language neutral).

## §D. Edge Cases

- **E1 ambiguous / unknown stack**: when the stack cannot be classified into
  web-frontend / mobile / backend-db, Phase 3.6 falls back to the
  `universal_starter` row (GitHub + Context7 + Playwright) rather than skipping MCP
  provisioning silently.
- **E2 user declines all servers**: when the user rejects the recommendation entirely
  via AskUserQuestion, Phase 3.6 writes NO `.mcp.json` entry and proceeds to Phase 3.7
  — a declined recommendation is not an error.
- **E3 existing `.mcp.json` with overlapping server**: when the existing file already
  contains a server with the same key, the additive merge keeps the existing entry
  (no duplicate, no clobber); Phase 3.6 does not silently overwrite a user-tuned
  entry.
- **E4 credentialed server with no env var available**: when a credentialed server is
  approved but the required `${VAR}` is not set in the environment, the entry is still
  written with the `${VAR}` placeholder (config-time); actual credential resolution is
  a runtime concern (out of scope) — never inline a literal.
- **E5 harness PLAN declares no MCP need**: artifact 6 is omitted and GENERATE output
  is byte-identical to the current 5-artifact set (REQ-HMP-009) — a no-MCP harness is
  unchanged from today.
- **E6 mcp-matrix config surface undecided at run-time**: if the M1 clarification is
  unresolved, apply the plan.md default (standalone data resource read as
  prose-context, no Go loader) and record the decision in progress.md §E.2; do NOT
  silently add a Go config-struct field.
- **E7 doctor validate-vs-tolerate undecided at run-time**: if the M3 clarification is
  unresolved, apply the plan.md default (tolerate-only, no Go change) and record it in
  progress.md §E.2; do NOT add `DisallowUnknownFields` or a Go `MCP` struct field.

## §E. Quality Gates

- TRUST 5: Tested (doc/config-only SPEC — verification is grep / diff / `make build` /
  `moai harness doctor`, not unit tests; state test ACs as N/A honestly); Readable
  (workflow prose stays consistent with surrounding sections); Unified (byte-identical
  template mirror); Secured (no literal secrets — `${VAR}` only; credentialed servers
  gated on explicit approval; NO-SPEC guard preserved); Trackable (Conventional Commits
  per milestone, `🗿 MoAI` trailer).
- Neutrality: `template-neutrality-check.yaml` + `internal_content_leak_test.go` green
  on the final push; matrix is project-type-keyed (16-language neutral).
- Doctor tolerance: `moai harness doctor` exit-0 on a manifest carrying the optional
  `mcp` block; no `DisallowUnknownFields` regression.
- Non-regression: `go build ./...` + `go test ./...` exit 0 (no Go code changed).

## §F. Definition of Done

1. All 13 ACs PASS (or documented N/A / PASS-WITH-DEBT with rationale) with verbatim
   command output recorded in progress.md §E.2.
2. Phase 3.6 inserted between Phase 3.5 (LSP) and Phase 3.7 (dev-mode) in
   `doc-generation.md`: stack detect → matrix select (3-5 cap, vendor-maintained) →
   orchestrator approval (per-server for credentialed) → additive `.mcp.json` write at
   project scope (`${VAR}` secrets, never literal).
3. `mcp-matrix.yaml` (web / mobile / backend + universal_starter) created under
   `.moai/config/sections/`, referenced (not duplicated) by the skill.
4. `harness-builder.md` GENERATE documents the optional artifact 6 with the
   conditional-emission (emit iff MCP need / byte-identical omission) clause; the
   optional manifest `mcp` block is doctor-tolerant.
5. Every touched file byte-identical between local and template trees; `make build` +
   neutrality guards green; `moai harness doctor` exit-0 with the optional mcp block.
6. The two `[NEEDS CLARIFICATION]` markers resolved (mcp-matrix config surface +
   doctor validate-vs-tolerate) before Implementation Kickoff Approval.
7. Sync-phase close by manager-docs per the Status Transition Ownership Matrix.
