---
spec_id: SPEC-CODEX-DUAL-AGENTS-001
status: in-progress
tier: M
era: V3R6
plan_complete_at: 2026-08-22
plan_status: audit-ready
---

# progress.md — SPEC-CODEX-DUAL-AGENTS-001

## Phase 1 — SKIP rationale

Research/context phase skipped per delegation: the orchestrator pre-gathered all plan-phase
inputs — the M0 measurement report (t91, `.moai/reports/t91/README.md` + `hook-payloads/`,
primary checkout) and the agent inventory inline in the delegation prompt. No separate
`research.md` is emitted (Tier M artifact set: spec.md + plan.md + acceptance.md +
progress.md). The M0 report was re-read first-hand during authoring (not taken on delegation
faith), and the agent inventory was re-verified against the template tree (6 corrections —
plan.md §A.2/§B.1).

## Phase 2 — Plan-phase summary (2026-08-22, manager-spec)

- Artifacts emitted: spec.md (14 GEARS requirements, Out of Scope naming M1–M4/M6), plan.md
  (verified inventory, §A.3 mapping table as first-class deliverable, Option A/B design
  decision, 4 milestones, 4 [NEEDS CLARIFICATION] markers with probe resolution paths),
  acceptance.md (13 testable ACs + 6 probe ACs + closure gates).
- Design recommendation requiring lead/auditor attention: Option A (`.md` IS the neutral core
  + mapping manifest; `.md` publication is identity) vs Option B (symmetric re-render) —
  plan.md §A.5.
- Unmeasured Codex semantics are probe items (P-01..P-06), never assumed facts; ship-omitted
  fallback rule governs unconfirmed values.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready — plan_complete_at: 2026-08-22

Plan-phase self-verification executed (all observed, not assumed):

- SPEC ID regex check (executed Bash, verbatim output `PASS`).
- ID uniqueness: `ls .moai/specs | grep CODEX-DUAL` → 0 hits (only SPEC-CODEX-PHASE2-001
  exists in the CODEX area).
- spec.md frontmatter: all 12 canonical fields present, schema-conformant, no snake_case
  aliases (validated against `.claude/rules/moai/development/spec-frontmatter-schema.md`).
- Agent inventory verified against the TEMPLATE tree (grep + full file reads) — ground truth
  recorded in plan.md §A.2.
- M0 facts cross-read from the t91 report with per-section citations.
- Out of Scope section satisfies the OutOfScopeRule lint convention (`### Out of Scope —`
  H3 sub-headings with `-` bullets).
- Tier M artifact set complete: spec.md + plan.md + acceptance.md + progress.md (4 files).
- Revision (plan-audit iter-1, 2026-08-22): mechanical fixes applied — D2 inventory cells
  corrected after re-run grep verification (super-advisor 11, sync-auditor 5, union 20/21
  with `goal_arm` absent, Web class +builder-harness, DesignSync = manager-design only);
  D3 AC-P01..P06 reclassified as probe records outside the Tier M AC budget; D4 §F
  documentation-grounding row annotated; D5 R-008 tag relabeled Event-driven + R-003
  rationale relocated to acceptance.md §D.1. D1 (four §A.4 [NEEDS CLARIFICATION] markers)
  intentionally untouched — pending lead decision.
- Lead decisions landed (2026-08-22): the four §A.4 markers converted to recorded decisions
  (probe-first with omit-on-unconfirmed for sandbox_mode and model_reasoning_effort; `model`
  omitted on all 11; subdirectory layout preferred with flat `moai-` prefix fallback);
  Option A lead-approved (2026-08-22, plan.md §A.5). Implementation Kickoff Approval is
  granted conditional on audit iteration 2 PASS (lead pre-approved run-phase entry on PASS,
  batch approval 2026-08-22).

## §E.2 Run-phase Evidence

### MS1 — Emitter core + neutral-layer contract (2026-08-22, manager-develop, TDD)

RED evidence (verbatim, captured BEFORE implementation — stubs returned
`agentemit: not implemented`):

```
$ go test ./internal/template/agentemit/...
--- FAIL: TestParseAgentDocParsesFrontmatterContract (0.00s)
    agentemit_test.go:99: ParseAgentDoc(agents/mdcarrier.md): agentemit: not implemented
--- FAIL: TestParseAgentDocRejectsBrokenSources (0.00s)
    agentemit_test.go:163: missing name: error "agentemit: not implemented" must name the offending value "name"
    agentemit_test.go:163: name stem mismatch: error "agentemit: not implemented" must name the offending value "differentname"
--- FAIL: TestEmitAllRoundTripsBodyVerbatim (0.00s)
    agentemit_test.go:176: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllMCPServerMapping (0.00s)
    agentemit_test.go:206: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllEffortMappingPerManifest (0.00s)
    agentemit_test.go:227: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllOmitsModel (0.00s)
    agentemit_test.go:254: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllSandboxOmittedWhenUnconfirmed (0.00s)
    agentemit_test.go:271: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllDeterministic (0.00s)
    agentemit_test.go:288: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllMarkdownIdentityIsPassThrough (0.00s)
    agentemit_test.go:321: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllFailClosedNegatives (0.00s)
    agentemit_test.go:381: LoadManifest: agentemit: not implemented
--- FAIL: TestEmitAllFailClosedDuplicateName (0.00s)
    agentemit_test.go:407: LoadManifest: agentemit: not implemented
--- FAIL: TestLoadManifestSelfValidates (0.00s)
    agentemit_test.go:430: LoadManifest: agentemit: not implemented
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template/agentemit	0.374s
FAIL
```

GREEN evidence (this run, this tree, HEAD f8b5d9a71 + M1 working tree):

```
$ go test -cover ./internal/template/agentemit/...
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.507s	coverage: 93.5% of statements

$ golangci-lint run --timeout=2m ./internal/template/agentemit/...
0 issues.

$ go build ./... && GOOS=windows GOARCH=amd64 go build ./...
(exit 0, both)

$ go vet ./internal/template/agentemit/...
(exit 0)
```

MS1 decisions recorded:
- TOML validation strategy (plan §A.6 left to MS1): INDEPENDENT test-side
  spec-subset decoder (`tomldecodertest_test.go`) + codex-cli smoke parsing
  (MS2/§D.3) — no new go.mod dependency. go.mod has no TOML library (direct
  or indirect); the emitted grammar is deliberately tiny (3 string forms +
  1 array form); the real consumer parses the artifacts in the probe smoke.
- Loader anchors frontmatter on the FIRST closing `---` (plan-auditor body
  contains a bare `---` hr at a later line — verified).
- Manifest embedded at `internal/template/agentemit/agents-codex.yaml`
  (build input; NOT under templates/; never distributed). `ParseManifest`
  exported for the M4 seam (plan §H).
- Fixture sources verified pre-implementation: 11/11 name==stem, zero `'''`,
  zero CR, all files end `\n`, all UTF-8.

### MS2 — Probes: lock the enums (2026-08-22, manager-develop; t91 §9 pattern)

Harness: isolated `CODEX_HOME=/tmp/t89-probe.H4oYTX/home` (auth copied in,
mode 600); probe project at `/tmp/t89-probe.H4oYTX/proj` with 12 probe agent
TOMLs (11 under `.codex/agents/moai/`, 1 flat under `.codex/agents/`); real
`~/.codex` verified untouched by mtime snapshot diff before/after
(`REAL_CODEX_HOME_UNTOUCHED`, config.toml + auth.json + hooks.json). Scratch
removed after evidence capture (auth copy deleted first). Total model calls:
2 (bounded; no loops, no background load). codex-cli 0.147.0 confirmed:
`codex --version` → `codex-cli 0.147.0` (measured version = manifest pin).

**Run 1 — P-04 layout + per-value file survival** (verbatim):

```
$ CODEX_HOME=<scratch>/home codex exec --dangerously-bypass-approvals-and-sandbox \
    -C <scratch>/proj --json "List every custom agent name available to you for
    delegation (the agent types you can spawn as subagents). Reply with ONLY the
    comma-separated list of names, no prose, no explanation." < /dev/null
{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"t89flatprobe, t89p01danger, t89p01readonly, t89p01wwrite, t89p02bogus, t89p02high, t89p02low, t89p02medium, t89p02xhigh, t89p03bogusmodel, t89p03omit, t89subprobe, default, explorer, worker"}
```

Readings: (a) `t89subprobe` (placed at `.codex/agents/moai/sub-probe.toml`)
IS listed → **P-04: `.codex/agents/` scans subdirectories — subdirectory
layout CONFIRMED** (manifest knob stays subdirectory/moai). (b) `t89p01bogus`
(sandbox_mode = "t89-bogus-sandbox") is ABSENT → a bad sandbox value kills
the whole file (the lead-cited hazard, now measured). (c) All four effort
candidates AND `t89p02bogus` AND `t89p03bogusmodel` register → effort/model
bad values are silently accepted at parse (silent-ignore zone).

**Run 2 — P-01 accepted set (runtime names it) + delegation smoke** (verbatim):

```
$ CODEX_HOME=<scratch>/home codex exec --dangerously-bypass-approvals-and-sandbox \
    -C <scratch>/proj --json "Delegate to the agent t89subprobe with the message
    'identify yourself'. Wait for its reply. Then delegate to the agent t89p01wwrite
    with the same message. Report both exact replies, one per line, prefixed SUB: and WW:." \
    < /dev/null
{"type":"item.completed","item":{"id":"item_0","type":"error","message":"Ignoring malformed agent role definition: failed to deserialize agent role file at /tmp/t89-probe.H4oYTX/proj/.codex/agents/moai/p01-bogus.toml: unknown variant `t89-bogus-sandbox`, expected one of `read-only`, `workspace-write`, `danger-full-access`\n"}}
{"type":"item.completed","item":{"id":"item_5","type":"agent_message","text":"SUB: T89SUBPROBE-OK\nWW: T89P01WWRITE-OK"}}
```

Readings: (a) **P-01: the runtime names the sandbox value set verbatim —
{read-only, workspace-write, danger-full-access}** (exactly 3 values;
malformed file visibly dropped with a named diagnostic). (b) Delegation to
the t91-pattern agent (fields name/description/developer_instructions/
model_reasoning_effort/sandbox_mode all present) returned `T89SUBPROBE-OK`;
delegation to the workspace-write agent returned `T89P01WWRITE-OK` — the
P-01 emitted-candidate is delegation-confirmed.

**P-02 static enum evidence** (0 model calls, t91's binary-string technique):

```
$ strings "<codex-runtime>/0.147.0-aarch64-apple-darwin/bin/codex" | grep -o "minimal[a-z ]*low[a-z ]*medium[a-z ]*high"
minimallowmediumhighxhigh   (x3 occurrences; a broader run adds none/max/ultra)
```

**P-02: {low, medium, high, xhigh} ⊂ the binary's reasoning-effort enum
{minimal, low, medium, high, xhigh}** + all four registered as agents →
identity mapping CONFIRMED and locked in the manifest.

**P-03**: `t89p03omit` (no model key) registered; `sub-probe`/`t89subprobe`
(no model key) delegated successfully → omission inherits the subagent
default and works. `t89p03bogusmodel` (model = "t89-bogus-model-string")
registered silently → arbitrary strings accepted at parse — emitting a
Claude alias would be accepted-but-wrong. **R-011 omit-model CONFIRMED as
the only safe choice.**

**P-05 (skills.config)**: SKIPPED — M1-deferred per plan §A.4 (M5 emits no
skills field regardless; no M5 emission decision depends on it).
**P-06 (per-agent MCP filtering)**: SKIPPED — optional; the coarse
server-level grant is the shipped design either way (documented drop stands).

**Manifest locks applied (RED→GREEN, tests in agentemit)**: sandbox_mode
emit=true value="workspace-write" accepted_values=[read-only,
workspace-write, danger-full-access]; layout subdirectory confirmed;
model_reasoning_effort identity map confirmed; FieldConfig.AcceptedValues
added with ParseManifest membership validation (fail-closed on an
unconfirmed value). MS2 RED evidence (verbatim):

```
$ go test ./internal/template/agentemit/...
--- FAIL: TestParseManifestFailClosed (0.00s)
    agentemit_edge_test.go:116: sandbox emitted outside the measured value set: want error, got nil
--- FAIL: TestEmitAllSandboxPerMeasuredSet (0.00s)
    agentemit_test.go:282: .codex/agents/moai/mdcarrier.toml: sandbox_mode = <nil>, want workspace-write (P-01 measured set member)
    agentemit_test.go:282: .codex/agents/moai/plainagent.toml: sandbox_mode = <nil>, want workspace-write (P-01 measured set member)
    agentemit_test.go:282: .codex/agents/moai/twoskills.toml: sandbox_mode = <nil>, want workspace-write (P-01 measured set member)
FAIL
```

GREEN (this run, this tree, HEAD 7a7a05384 + MS2 working tree):

```
$ go test -cover ./internal/template/agentemit/...
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	7.121s	coverage: 93.7% of statements
```

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Logged by the orchestrator (lane-9) before the first run-phase Agent() spawn.

**Input parameters**: tier M; scope ~10 files (emitter package + manifest + 11 TOML + tests + template tree); domains 2 (Go emitter code + template artifacts); language mix Go + TOML + markdown; concurrency benefit LOW (coding-heavy — single coherent emitter, sequential test build); Agent Teams prereqs N/A.

**Mode evaluation**:
- direct — not selected: >trivial (emitter package + tests + probes)
- serial — **selected**: coding-heavy Tier M implementation per Anthropic's coding-task parallelism caveat; MS1→MS4 are dependent stages (probes lock enums the emitter consumes; mass emission needs the emitter)
- fanout — not selected: 2 domains only; concurrency benefit LOW
- sweep — not selected: ~10 files < ~30 threshold; not a uniform mechanical transform (new code)

**Decision: serial** (manager-develop, cycle_type=tdd, sequential MS delegation)

**Justification**: The emitter core is new-code work with tight internal coupling (manifest schema → loader → TOML writer → validators); probe milestone MS2 feeds decisions MS3 consumes. Sequential single-agent delegation matches the dependency chain; fan-out would only parallelize the mechanical MS3 emission, which is one command once MS1 exists. Kickoff: lead batch approval 2026-08-22 + plan-audit iter-3 PASS (conditional grant fired).

**Plan Audit Gate skip record** (run Phase 1): most recent verdict PASS (iter-3, 0.92 ≥ Tier M 0.80); plan-artifact hash unchanged since the verdict (progress.md §F addition is not a hash subject); skip-eligible per the three-condition contract — recorded here per the skip-recording obligation.
