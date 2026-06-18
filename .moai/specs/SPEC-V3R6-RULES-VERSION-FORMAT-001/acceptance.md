# Acceptance Criteria — SPEC-V3R6-RULES-VERSION-FORMAT-001

All ACs are grep-verifiable in BOTH the deployed file (`.claude/rules/moai/**`) and its template mirror (`internal/template/templates/.claude/rules/moai/**`), plus the Go-test trio + `moai constitution validate` gates. Commands are illustrative; run-phase substitutes exact line targets.

## §D. AC Matrix

| AC ID | REQ | Scope | Verification (deployed + template) | Severity |
|-------|-----|-------|-----------------------------------|----------|
| AC-VFM-001a | REQ-VFM-001 | zone-registry clause↔source (substring, validator-independent) | for EACH edited clause (CONST-V3R2-028 L291, CONST-V3R2-029 L299): `grep -Fq "<clause text>" <source file>` succeeds (the clause is a `strings.Contains` substring of its identically-edited source). NOT "`moai constitution validate` → 0 SentinelDrift" — that gate is baseline-RED today (registry load aborts on CONST-V3R6-001, see §D.6 blocker). | MUST |
| AC-VFM-001b | REQ-VFM-001 | full validator green-gate (BLOCKED — sequencing dep) | `moai constitution validate` exits 0 — DEFERRED: blocked on sibling `SPEC-V3R6-RULES-CONST-RULEID-001` landing first (broadens `ruleIDPattern` to admit V3R6). Re-verify AFTER that SPEC merges. Documented as a precondition-blocker, NOT a within-this-SPEC gate. | SHOULD (post-dependency) |
| AC-VFM-001c | REQ-VFM-001 | constitution Principle 4/5 | delta-grep (§D.5): bare current-substrate "Opus 4.7" drops to 0 at the Principle 4/5 lines AND "Opus 4.7+ / 4.8" framing appears, deployed + template | MUST |
| AC-VFM-001d | REQ-VFM-001 | zone-registry clause↔source (verbatim, EXCLUDING L845) | each edited clause is a verbatim `grep -F` substring of its source line — for CONST-V3R2-028 (L291) + CONST-V3R2-029 (L299) ONLY. **EXCLUDES CONST-V3R5-022 (L845)** — see AC-VFM-001e. | MUST |
| AC-VFM-001e | REQ-VFM-001 | L845 carve-out (pre-existing non-substring) | CONST-V3R5-022 (L845) clause `"...1M context (Opus 4.7) = 50%, 200K context (Sonnet/Opus/Haiku) = 90%"` is ALREADY a non-substring of its source (`grep -cF '1M context (Opus 4.7)' context-window-management.md` == 0 — source uses a TABLE `\| Opus 4.8 (1M) \| ... \| 50% \|`). L845 is SHOULD-scoped and LEFT UNEDITED by this SPEC (editing it cannot produce a clean substring without a table-text rewrite, out of scope). Verified: this AC asserts the count stays 0 and L845 is not touched. | SHOULD |
| AC-VFM-002a | REQ-VFM-002 | csharp.md | `.NET 10`+`C# 14` present; `.NET 8`+`C# 12` absent | MUST |
| AC-VFM-002b | REQ-VFM-002 | kotlin.md | `Kotlin 2.2`+`Ktor 3.5` (or 3.x) present; `Kotlin 2.0`+`Ktor 3.0` identity absent | MUST |
| AC-VFM-002c | REQ-VFM-002 | ruby.md (Ruby + Rails, paired) | delta-grep (§D.5): `Ruby 4.0` present at version-identity pin (L11/L16) AND `Rails 8.0` present (L11/L17/L60) AND old `Rails 7.2` drops to 0. Rails MUST be bumped alongside Ruby — a Ruby-4.0/Rails-7.2 pairing is internally inconsistent (Rails 8.x is current). | MUST |
| AC-VFM-002d | REQ-VFM-002 | go.md | `1.26-alpine` present; `golang:1.23-alpine` absent; `Go 1.23+` floor retained | MUST |
| AC-VFM-002e | REQ-VFM-002 | verified-correct pins | python/swift/php/elixir/rust/ts-React19-Next16 UNCHANGED | MUST |
| AC-VFM-003 | REQ-VFM-003 | footer policy | footer-consistency policy statement present (consistent-by-absence); named SSOT files carry footer | MUST |
| AC-VFM-004a | REQ-VFM-004 | settings-management.md | no Korean instruction prose; alwaysLoad note reads English | MUST |
| AC-VFM-004b | REQ-VFM-004 | session-handoff Diet/V0 | Diet/V0 blocks English; localization examples `전제 검증:`/`실행:`/`머지 후:` PRESERVED | MUST |
| AC-VFM-004c | REQ-VFM-004 | manager-develop §검증 | §verification block English; no Korean prose | MUST |
| AC-VFM-005 | REQ-VFM-005 | skill-writing-craft + ab-testing | `grep -c '✅\|❌'` == 0 in both files (deployed + template) — matcher covers BOTH emoji to stay aligned with the run-phase scan | MUST |
| AC-VFM-006a | REQ-VFM-006 | orchestrator-templates | no "N days/hours" planning estimate ("2 days" gone); phase ordering present | MUST |
| AC-VFM-006b | REQ-VFM-006 | manager-develop targets | no wall-time target ("≤30분"/"91분") survives | MUST |
| AC-VFM-006c | REQ-VFM-006 | team-pattern-cookbook | no "Day N" label; phase ordering present | MUST |
| AC-VFM-007a | REQ-VFM-007 | session-handoff Block 1 | ultrathink→`effort: xhigh` + Adaptive-Thinking-distinct-axis framing | MUST |
| AC-VFM-007b | REQ-VFM-007 | hooks-system Setup | framing names internal `EventSetup` retirement AND states upstream CC `Setup` event is current | MUST |
| AC-VFM-008a | REQ-VFM-008 | enrolled-byte-parity files | `TestRuleTemplateMirrorDrift` PASS — NON-vacuous for the enrolled subset I edit: `core/hooks-system.md` + `workflow/session-handoff.md` (both in `workflowOptMirroredPaths`) | MUST |
| AC-VFM-008b | REQ-VFM-008 | non-enrolled byte-identical files | explicit per-file `diff -q <deployed> <mirror>` == identical for the 12 byte-identical-but-NOT-enrolled files I edit (see §D.4 file-mechanism table) — NOT a vacuous `TestRuleTemplateMirrorDrift PASS` | MUST |
| AC-VFM-008c | REQ-VFM-008 | zone-registry (intentionally divergent) | `TestTemplateNoInternalContentLeak` PASS + scoped `diff` confirming ONLY the 3 Opus-clause lines (L291/L299/L845 region) changed identically deployed↔mirror; NEVER whole-file `diff -q` (false-fails on the deployed-only `CONST-V3R6-001` block at L1006-1018) | MUST |
| AC-VFM-008d | REQ-VFM-008 | neutrality (whole corpus) | `TestTemplateNoInternalContentLeak` + `TestTemplateNeutralityAudit` PASS | MUST |

## §D.1 Given-When-Then scenarios

### Scenario 1 — Zone-registry clause stays a valid substring after model-identity edit
- **Given** `core/zone-registry.md` clause CONST-V3R2-028 carries "Opus 4.7 does not auto-spawn subagents" and `core/moai-constitution.md` L56 carries the same substring.
- **When** the run-phase agent generalizes the model identity in the constitution source line AND edits the registry clause to the identical new text (atomic pair), then runs `moai constitution validate`.
- **Then** the validator emits 0 `SentinelDrift` findings (the clause remains a whitespace-normalized substring of its source).

### Scenario 2 — Stale csharp version is corrected everywhere, including the mirror
- **Given** `languages/csharp.md` (and its template mirror) read `.NET 8 / C# 12` in title, features, and body.
- **When** the agent replaces every occurrence with `.NET 10 (LTS) / C# 14` in deployed AND template.
- **Then** `grep '.NET 10'` matches and `grep '.NET 8'` returns no version-identity match in BOTH files, and `TestRuleTemplateMirrorDrift` passes.

### Scenario 3 — Korean instruction prose is translated but example renderings are preserved
- **Given** `workflow/session-handoff.md` has English-target instruction prose in § Diet Constraints / § V0 Abort Gate, AND canonical Korean paste-ready example renderings (`전제 검증:`, `실행:`, `머지 후:`).
- **When** the agent translates ONLY the instruction-prose blocks to English.
- **Then** the Diet/V0 instruction prose reads English AND `grep '전제 검증:'` still matches (the locale-output examples are intentionally preserved) in deployed + template.

### Scenario 4 — hooks-system Setup framing is disambiguated, not over-corrected
- **Given** `core/hooks-system.md` L62 reads "Setup | REMOVED" implying the upstream CC event is gone.
- **When** the agent reframes to state the moai-adk Go `EventSetup` constant + `moai hook setup` subcommand were retired (internal) while the upstream Claude Code `Setup` event remains current.
- **Then** grep finds both the internal-retirement statement and the "upstream Setup event is current" clarification; the file no longer implies the upstream event is gone (deployed + template).

### Scenario 5 — team-pattern-cookbook time-fix lands without colliding with SSOT-DEDUP
- **Given** `workflow/team-pattern-cookbook.md` has "Day 1/2/3" labels AND is a structural-edit target of SPEC-V3R6-RULES-SSOT-DEDUP-001.
- **When** this SPEC replaces the day labels with phase ordering as an isolated edit (and the orchestrator sequences the two SPECs rather than running them blind-parallel).
- **Then** `grep 'Day [0-9]'` returns no planning label, phase ordering is present, and the SSOT-DEDUP structural edit rebases cleanly (different section) in deployed + template.

### Scenario 6 — A forgotten mirror edit on a non-enrolled file is caught (anti-vacuity)
- **Given** `languages/csharp.md` is byte-identical-by-discipline deployed↔template but is NOT enrolled in `workflowOptMirroredPaths`, and the run-phase agent edits the deployed file but forgets the template mirror.
- **When** verification runs the §D.4 mechanism — an explicit `diff -q .claude/rules/moai/languages/csharp.md internal/template/templates/.claude/rules/moai/languages/csharp.md` — rather than relying on `TestRuleTemplateMirrorDrift` (which does not enroll this file and stays green).
- **Then** the `diff -q` reports the files differ and the AC FAILS, surfacing the forgotten mirror edit that the vacuous test-only assertion would have missed.

### Scenario 7 — zone-registry Opus-clause edit verified without whole-file false-fail
- **Given** `core/zone-registry.md` deployed carries a deployed-only `CONST-V3R6-001` block (L1006-1018) absent from the §25-neutralized mirror, so the file is intentionally divergent, while my Opus-clause edits (L291/L299/L845 region) are in the byte-identical region.
- **When** verification asserts mirror parity via `TestTemplateNoInternalContentLeak` PASS + a SCOPED `diff` of only the Opus-clause lines — NOT a whole-file `diff -q`.
- **Then** the scoped diff confirms the Opus-clause lines changed identically on both sides, the leak-test stays green, and no false-fail is triggered by the unrelated `CONST-V3R6-001` divergence; AND `moai constitution validate` reports 0 `SentinelDrift` because each edited clause remains a verbatim substring of its (identically-edited) source.

## §D.2 Edge cases

- A language file whose version pin is verified-correct (python/swift/php/elixir/rust) is touched ZERO times — AC-VFM-002e guards against accidental edits.
- A zone-registry clause whose source-sync is too entangled is scoped SHOULD with a note (REQ-VFM-001 escape) rather than half-edited — no `SentinelDrift` results.
- The `10s` hook-timeout in orchestrator-templates and the `0.5s` example test-output in manager-develop are literal values, NOT planning estimates — they are LEFT unchanged (AC-VFM-006 greps target planning estimates only).

## §D.4 Mirror-parity mechanism per edited file (anti-vacuity)

[HARD] **Mirror-parity AC vacuity avoidance.** `TestRuleTemplateMirrorDrift` enrolls ONLY the explicit `workflowOptMirroredPaths` allowlist (`internal/template/rule_template_mirror_test.go`). For files I edit that are byte-identical-by-discipline but NOT enrolled, a bare "TestRuleTemplateMirrorDrift PASS" clause is **vacuously green** (a forgotten mirror edit would pass both Go tests silently). The mechanism below is assigned from VERIFIED `diff -q` ground-truth (run at plan-time), NOT from assumption.

| Edited file | Enrolled in `workflowOptMirroredPaths`? | Deployed↔template (plan-time `diff`) | Correct AC mechanism |
|-------------|------------------------------------------|--------------------------------------|----------------------|
| `core/hooks-system.md` | YES | identical | `TestRuleTemplateMirrorDrift` (non-vacuous) + `diff -q` belt-and-suspenders |
| `workflow/session-handoff.md` | YES (comment L47-51: byte-identical post-neutralization) | identical | `TestRuleTemplateMirrorDrift` (non-vacuous) + `diff -q` belt-and-suspenders |
| `core/moai-constitution.md` | NO | identical | explicit `diff -q` byte-identity |
| `core/settings-management.md` | NO | identical | explicit `diff -q` byte-identity |
| `workflow/worktree-integration.md` | NO | identical | explicit `diff -q` byte-identity |
| `workflow/context-window-management.md` | NO | identical | explicit `diff -q` byte-identity |
| `workflow/team-pattern-cookbook.md` | NO | identical | explicit `diff -q` byte-identity |
| `development/agent-authoring.md` | NO | identical | explicit `diff -q` byte-identity |
| `development/skill-authoring.md` | NO | identical | explicit `diff -q` byte-identity |
| `development/orchestrator-templates.md` | NO | identical | explicit `diff -q` byte-identity |
| `development/manager-develop-prompt-template.md` | NO | identical (live tree; the test-source comment lists it as a historical leak target, but it is byte-identical TODAY) | explicit `diff -q` byte-identity + re-confirm leak-test green |
| `development/skill-writing-craft.md` | NO | identical | explicit `diff -q` byte-identity |
| `development/skill-ab-testing.md` | NO | identical | explicit `diff -q` byte-identity |
| `languages/csharp.md` | NO | identical | explicit `diff -q` byte-identity |
| `languages/kotlin.md` | NO | identical | explicit `diff -q` byte-identity |
| `languages/ruby.md` | NO | identical | explicit `diff -q` byte-identity |
| `languages/go.md` | NO | identical | explicit `diff -q` byte-identity |
| `core/zone-registry.md` | NO | **DIVERGENT** (deployed-only `CONST-V3R6-001` block L1006-1018, §25-stripped from neutral mirror) | `TestTemplateNoInternalContentLeak` + **scoped** `diff` of ONLY the 2 edited Opus-clause lines (L291 CONST-V3R2-028, L299 CONST-V3R2-029) — NEVER whole-file `diff -q`. L845 (CONST-V3R5-022) is NOT edited (AC-VFM-001e carve-out). |

Note on §25-sanitized files: the coordinator-relayed advisory predicted `settings-management.md` AND `manager-develop-prompt-template.md` would be §25-divergent (needing leak-test + diff-only-my-edit). Plan-time `diff` ground-truth REFUTES that for the live tree: both are byte-identical TODAY, so a plain `diff -q` is the correct (and stronger) mechanism. Only `zone-registry.md` is actually divergent, and its divergence is far from my edit lines. The run-phase agent MUST re-run `diff -q` immediately before committing (TOCTOU guard) — if any file has become divergent since plan-time, downgrade it to the leak-test + scoped-diff mechanism.

## §D.5 Delta-grep tightening (anti-vacuity for presence checks)

[HARD] Every version/string-correction AC uses a **before/after delta grep**, not a bare presence check (a presence-only grep can be pre-satisfied and thus vacuous). Each such AC asserts BOTH:
- the distinctive OLD string drops to 0 occurrences (`grep -c '<old>' == 0`), AND
- the NEW string appears (`grep -c '<new>' >= 1`),
in deployed AND template. Applies to AC-VFM-001c (Opus identity), AC-VFM-002a/b/c/d (language pins incl. ruby Rails), AC-VFM-005 (emoji `grep -c '✅\|❌' == 0`), AC-VFM-006a/b/c (time labels drop to 0). Pure-presence checks are permitted only where there is no removable old token (e.g. AC-VFM-003 footer-policy statement, AC-VFM-007a/b reframing — these assert new-text presence because no distinctive old token is being removed).

## §D.6 Keystone-gate dependency blocker (D1 — baseline-RED validator)

[HARD] **`moai constitution validate` is RED at baseline today — independent of this SPEC.** Verified deterministically:

```
$ moai constitution validate
Error: load registry: entry 114 validation error: rule ID "CONST-V3R6-001" does not match pattern "^CONST-V3R[25]-\d{3,}$"
```

`internal/constitution/rule.go:10` defines `ruleIDPattern = ^CONST-V3R[25]-\d{3,}$` (admits only V3R2/V3R5). The deployed `zone-registry.md` carries a `CONST-V3R6-001` entry (the `CONST-V3R6-NNN` namespace block, L1006-1018) that the pattern rejects, so `LoadRegistry` aborts BEFORE the clause-substring drift loop ever runs. The full green-gate is therefore **unreachable until a separate fix lands**.

Consequence for this SPEC's ACs:
- **AC-VFM-001a is re-scoped** to assert the substring fact DIRECTLY via `grep -Fq` (the verifiable invariant: each edited clause text IS a `strings.Contains` substring of its identically-edited source), bypassing the dead validator. This is verifiable TODAY.
- **AC-VFM-001b (full `moai constitution validate` exits 0)** is a SHOULD, BLOCKED on the sibling SPEC below, and re-verified only after it merges.

Sequencing dependency: **`SPEC-V3R6-RULES-CONST-RULEID-001` → `SPEC-V3R6-RULES-VERSION-FORMAT-001`.** The sibling SPEC owns the 1-line Go fix broadening `ruleIDPattern` to admit `V3R6` (`^CONST-V3R[256]-\d{3,}$` or equivalent). It is OUT OF SCOPE here (Go code change; this SPEC is doc-only). The run-phase orchestrator MUST sequence CONST-RULEID-001 first if the full validator green-gate is wanted before merge; otherwise this SPEC ships with the `grep -Fq` substring proof (AC-VFM-001a) and AC-VFM-001b deferred.

## §D.3 Quality gates / Definition of Done

- [ ] All MUST ACs pass with delta-grep discipline per §D.5 (old→0 AND new≥1, deployed + template).
- [ ] Mirror parity verified per §D.4 file-mechanism table: enrolled files via `TestRuleTemplateMirrorDrift` (non-vacuous) + `diff -q`; 12 non-enrolled byte-identical files via explicit per-file `diff -q`; `zone-registry.md` via leak-test + scoped Opus-clause-line `diff` (NEVER whole-file `diff -q`).
- [ ] Per-file `diff -q` re-run immediately before commit (TOCTOU guard) — any newly-divergent file downgraded to leak-test + scoped-diff.
- [ ] zone-registry clause-substring invariant verified DIRECTLY via `grep -Fq` (AC-VFM-001a) for CONST-V3R2-028 + CONST-V3R2-029 — NOT via `moai constitution validate` (baseline-RED today per §D.6).
- [ ] L845 (CONST-V3R5-022) left UNEDITED; `grep -cF '1M context (Opus 4.7)' context-window-management.md` stays 0 (AC-VFM-001e carve-out).
- [ ] (Deferred, post-dependency) full `moai constitution validate` exits 0 — re-verify ONLY after `SPEC-V3R6-RULES-CONST-RULEID-001` merges (§D.6 sequencing dependency).
- [ ] Go-test trio (`TestRuleTemplateMirrorDrift`, `TestTemplateNoInternalContentLeak`, `TestTemplateNeutralityAudit`) GREEN.
- [ ] `go test ./...` GREEN (no production-code regression; zero Go source change expected).
- [ ] No archived-agent reference touched (CATALOG-SCRUB lane); no SSOT de-dup performed (SSOT-DEDUP lane).
- [ ] Language-skeleton restructuring NOT performed (deferred to SPEC-V3R6-RULES-LANG-SKELETON-001).
- [ ] Footer policy = consistent-by-absence stated; bulk footer-adding NOT performed.
