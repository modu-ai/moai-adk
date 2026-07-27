# SPEC-SEC-GUARDIAN-001 — Design

## §A Architecture — 3-layer hook topology

```
┌─────────────────────────────────────────────────────────────────────────┐
│  MoAI Claude Code session (working tree)                                  │
│                                                                           │
│  Edit/Write ──PostToolUse──► handle-security-scan.sh ──► moai hook        │
│   (L1)         (Write|Edit|MultiEdit, async, 5s)          security-scan   │
│                                                            │              │
│                                    internal/hook/security/ ▼ (Go)         │
│                                    ┌──────────────────────────────┐       │
│                                    │ patterns.go  (single-source) │       │
│                                    │ scan.go      (buffer scan)   │◄──┐   │
│                                    │ diff.go      (git-diff scan) │   │   │
│                                    └──────────────────────────────┘   │   │
│  turn end ────Stop──────────► handle-security-turn.sh ──► moai hook   │   │
│   (L2)         (Stop, 5s)                                 security-turn┘  │
│                                    (runs pattern engine over turn diff)   │
│                                                                           │
│  git commit ──(L3 surface)──► handle-security-commit.sh ─► moai hook      │
│   (L3, opt-in)  (dormant unless MOAI_SECURITY_COMMIT_REVIEW)  security-   │
│                                    (cross-file data-flow review)  commit  │
└─────────────────────────────────────────────────────────────────────────┘
        │ block signal (structured JSON on exit-0 stdout channel)
        ▼
   orchestrator translates → AskUserQuestion(accept / --skip-hook / abort)
```

**Key design principle**: the three layers share ONE Go pattern engine (`internal/hook/security/`). Layer 1 scans a single write buffer; Layer 2 scans the turn's diff; Layer 3 does cross-file data-flow reasoning. The shell wrappers are thin forwarders (3-tier `moai` resolution + fail-open, identical to the existing 31 wrappers). The Go handlers are compiled into the binary and tested — NOT template content (REQ-SG-052). The wrappers + settings.json wiring ARE template content (16-language-neutral).

## §B Layer contract table

| Layer | Event | Matcher | Wrapper | `moai hook` subcmd | Input | Output (advisory default) | Default |
|-------|-------|---------|---------|--------------------|-------|---------------------------|---------|
| L1 | PostToolUse | `Write\|Edit\|MultiEdit` | `handle-security-scan.sh` | `security-scan` | `tool_input.content` / `.new_string` | `additionalContext` (async) | ON |
| L2 | Stop | (none) | `handle-security-turn.sh` | `security-turn` | working-tree diff | `systemMessage` | ON (regex-only) |
| L3 | (see §L3) | (see §L3) | `handle-security-commit.sh` | `security-commit` | commit changed-files + related files | `systemMessage` / structured escalation | OFF (dormant) |

All three: exit 0 always; blocking only via opt-in env flag; fail-open silent no-op on missing `moai`/`jq`/`git`.

## §C Exit-code / structured-JSON contract (REQ-SG-061)

Per the hook event schema (`.claude/rules/moai/core/hooks-system.md` § Hook Event stdin/stdout Reference) + the hook invocation surface (`agent-common-protocol.md` § Hook Invocation Surface):

- **L1 (PostToolUse, async)**: PostToolUse cannot block; async delivers ONLY `additionalContext` (next turn). Advisory findings ride `additionalContext`. Never emits `decision`.
- **L2 (Stop)**: advisory default emits `{"systemMessage": "..."}`. Opt-in blocking emits `{"hookSpecificOutput":{"hookEventName":"Stop","decision":"block","reason":"..."},"systemMessage":"..."}` on exit-0 stdout (the blocking channel is honored only on exit 0; exit 2 discards stdout). Mirrors `sync-phase-quality-gate.sh` exactly.
- **L3**: schema of whichever event it binds to (§L3). Advisory `systemMessage` by default; opt-in block via the event's decision channel.
- **Orchestrator translation (REQ-SG-041)**: on a block signal, the orchestrator parses the JSON (`decision`/`reason`) and runs `AskUserQuestion(accept / override --skip-hook / abort)`. No hook calls AskUserQuestion (REQ-SG-040).
- **Recovery-Signal Carve-Out**: on a recovery turn (`stopReason` ∈ {sync-failure, compact, PTL, max_output_tokens, media_size}), guardian gates SHOULD defer (exit 0) — documentation-only at this layer (hooks do not parse `stopReason`), same posture as the other gates (`runtime-recovery-doctrine.md` §4).

## §P Pattern-config schema (single-source, REQ-SG-053) + vulnerability-class taxonomy

`internal/hook/security/patterns.go` — one exported table, organized by CLASS:

```go
// VulnClass groups patterns by vulnerability class, NOT by language (REQ-SG-011).
type VulnClass struct {
    Name        string          // e.g. "unsafe-deserialization"
    Severity    Severity        // critical | high | medium | low
    Description string          // one-line human explanation
    Patterns    []*regexp.Regexp // language-agnostic regexes; a class applies across all 16 languages
    Langs       []string        // languages the class is meaningful for (empty = all 16)
}
```

Design choice (design fork recorded): **language-agnostic regexes keyed by class** (a `yaml.load(`/`pickle.load(`/`innerHTML =` regex matches regardless of which of the 16 languages the file is) is preferred over a per-language-variant map, because (a) most dangerous idioms are token-shaped and language-portable, (b) it keeps the table small and single-source, (c) it satisfies 16-language neutrality by construction (no PRIMARY language). The `Langs` field narrows a class only when a pattern is genuinely language-specific.

### Vulnerability-class taxonomy (≈25 patterns across ~10 classes — REQ-SG-012)

| Class | Severity | Example patterns (token-shaped, cross-language) |
|-------|----------|-------------------------------------------------|
| unsafe-deserialization | high | `yaml.load(` w/o SafeLoader, `pickle.load(`, `torch.load(` w/ `weights_only=False`, `Marshal.load`, `ObjectInputStream`, `unserialize(` |
| dom-injection-xss | high | `.innerHTML =`, `dangerouslySetInnerHTML`, `document.write(`, `v-html=`, `.html(` (jQuery) |
| hardcoded-secret | critical | `api[_-]?key\s*[:=]\s*["']`, `password\s*[:=]\s*["']`, `-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----`, AWS/token literals |
| code-injection-eval | high | `eval(` on external input, `exec(`, `Function(` ctor, `new Function`, `os.system(`, `child_process.exec(` w/ interpolation |
| sql-injection | high | string-concatenated SQL (`"SELECT ... " + `, f-string/interp into a query), `.raw(` with interpolation |
| command-injection | high | `subprocess.*shell=True`, `child_process.exec(` w/ unsanitized arg, `os/exec`+`sh -c`, backtick shell in Ruby/PHP |
| path-traversal | medium | user-supplied path joined into `open(`/`readFile(` without sanitization heuristic |
| ssrf | medium | user-supplied URL into `requests.get(`/`fetch(`/`http.Get(` without allowlist heuristic |
| weak-crypto | medium | `MD5`/`SHA1` for password/token, `DES`/`ECB` mode, hardcoded IV |
| insecure-random | medium | `Math.random()` for token/secret, `rand()` for crypto, non-CSPRNG |

The taxonomy draws vocabulary from `moai-ref-owasp-checklist`, `moai-ref-llm-security` (the `torch.load`/`yaml.load` ML-poisoning idioms), `moai-ref-secops`, and `moai-ref-supply-chain` — consulted at authoring time, not shipped.

### False-positive posture

Advisory-only (never blocks) means a false positive costs one advisory line, not a broken edit — so the patterns favor recall over precision. Where a pattern is inherently noisy (e.g. any `eval(`), the finding is `medium` and phrased as "review", not "vulnerability". Benign-fixture tests (`patterns_test.go`) pin the false-positive baseline so tuning is measured, not guessed.

## §L3 — Layer-3 surface (SETTLED: L3-A)

Two realizations, both dormant unless `MOAI_SECURITY_COMMIT_REVIEW` is enabled:

| Option | Event | Shape | Pros | Cons |
|--------|-------|-------|------|------|
| **L3-A (RECOMMENDED)** — extend the sync-gate model | Stop | A sibling `security-commit-review` Stop hook that detects a commit landed this turn (commit-subject / HEAD-changed, like `sync-phase-quality-gate.sh` detects sync commits), runs the Go engine over the commit's changed files + reads related files (Go `Grep`/`Glob` equivalents) to surface cross-file candidates, then emits a structured escalation signal the orchestrator turns into an agentic `Agent()` review | deterministic hook stays fast + boundary-clean; reuses the proven commit-aware gate model (REQ-SG-033); agentic reasoning is orchestrator-mediated (hook→orchestrator translation) | the deep cross-file reasoning is not IN the hook — it is escalated (one extra hop) |
| **L3-B** — native `type: agent` commit hook | PreToolUse `Bash(git commit *)` (via `if: "Bash(git commit *)"`) | Claude Code natively spawns a `type: "agent"` hook with Read/Grep/Glob that traces cross-file data flow and returns `{ok, reason}` | truest to the absorbed "reviewer reads related files" shape; can block (ok:false) pre-commit | heavier per-commit (spawns a subagent, 30-60s agent-hook timeout); blocking-capable → must be carefully advisory-gated; runs the agent even on trivial commits |

**Decision: L3-A (settled by orchestrator).** It keeps every guardian HOOK deterministic + fast + boundary-clean, reuses the `sync-phase-quality-gate.sh` commit-aware model directly (REQ-SG-033 "extend the sync-gate model"), and delivers the agentic cross-file reasoning through the sanctioned hook→orchestrator-translation path (the hook emits a structured signal; the orchestrator spawns the `Agent()` review) rather than an in-hook subagent. L3-B remains a documented, NOT-chosen alternative.

Both options are **dormant by default** (REQ-SG-032): the handler self-gates to a silent no-op (exit 0, empty) unless `MOAI_SECURITY_COMMIT_REVIEW` is enabled — so no per-commit cost is paid until the user opts in. This mirrors the `team-ac-verify.sh` "dormant by default" registration precedent.

## §E Opt-in env-flag matrix (advisory-first, REQ-SG-042)

| Flag | Default | Effect |
|------|---------|--------|
| (none) | — | L1 + L2 advisory (regex, non-blocking); L3 dormant |
| `MOAI_SECURITY_TURN_REVIEW=llm` | unset | L2 escalates the turn-diff review to a fast-model / agentic pass (opt-in) |
| `MOAI_SECURITY_COMMIT_REVIEW` (on) | unset (off) | L3 activates (dormant→advisory cross-file review) |
| `MOAI_SECURITY_BLOCKING` (or per-layer) | unset (advisory) | promotes a layer's finding to a blocking decision (opt-in, `MOAI_SYNC_GATE_BLOCKING`-aligned) |
| `--skip-hook` (first arg) | — | one-shot bypass, logged to `.moai/logs/hook-skip.log` (REQ-SG-043) |

Exact flag names are a run-phase detail; the CONTRACT is: advisory is the floor, every escalation is opt-in, every bypass is audit-logged.

## §H Hook-independence check (`hook-independence.md` §7)

Each new guardian hook answered against the authoring checklist:

1. Depends on `moai` binary? YES → carries the 3-tier resolution chain + silent `exit 0` fallback (mode A, acceptable-by-design; the wrapper does not crash when `moai` is absent).
2. New shared condition? NO new shared config/binary/env beyond the existing `${CLAUDE_PROJECT_DIR:-$PWD}` convention (mode E, benign).
3. `--skip-hook` bypass? logged to `.moai/logs/hook-skip.log` (mode B, acceptable).
4. Graceful degradation? YES → no-op on missing `moai`/`jq`/`git` (REQ-SG-060).
5. Surfaced degradation? Layer 1 is silent-on-clean by design; a one-time "moai unresolvable" warning is out of scope (the existing mode-A recommendation, deferred).

## §M Distribution design (run-phase order of operations)

1. Go FIRST: author `internal/hook/security/{patterns,scan,diff}.go` + per-layer handlers + `internal/cli/hook.go` subcommand registration + tests (`*_test.go`, t.TempDir, no OTEL env parallel).
2. Template tree: author `internal/template/templates/.claude/hooks/moai/handle-security-{scan,turn,commit}.sh.tmpl` (3-tier resolution + fail-open) → wire settings.json.tmpl (PostToolUse + Stop entries; L3 self-gating dormant).
3. Local sync: byte-lockstep `.sh` siblings + rendered `settings.json`.
4. `make build` recompiles the binary + re-embeds the template tree.
5. Verify: `go test ./...`, `go build` (host + windows), neutrality + catalog-count guards, boundary grep.

## §T Test/CI mapping

| Gate | Proves |
|------|--------|
| `internal/hook/security/patterns_test.go` | vuln-class coverage + 16-language applicability + false-positive baseline (REQ-SG-011, -012) |
| `internal/hook/security/scan_test.go` | L1 findings on dangerous fixture, silence on clean, never-block (REQ-SG-010, -013, -014) — t.TempDir, no OTEL env |
| `internal/hook/security/diff_test.go` | L2 turn-diff scan; L3 changed-files scan (REQ-SG-020, -030) |
| boundary grep (`internal/hook/security/` + `.claude/hooks/moai/`) | no AskUserQuestion (REQ-SG-040) |
| `diff` wrapper local↔template + settings.json local↔template | byte-lockstep (REQ-SG-050) |
| `internal_content_leak_test.go` + `template-neutrality-check.yaml` | no internal SPEC-ID/date/SHA; no PRIMARY language (REQ-SG-051) |
| `catalog_tier_audit_test.go` (agent/skill counts unchanged) | no new agent/skill (REQ-SG-002 non-goal) |
| `go build ./...` + `GOOS=windows go build` + `go test ./...` | regression freedom + cross-platform (REQ-SG-052) |
