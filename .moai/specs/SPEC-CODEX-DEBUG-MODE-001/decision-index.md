# SPEC-CODEX-DEBUG-MODE-001 — Decision Index

Carrier for decisions surfaced during plan-phase assembly that the operator has not settled in an interview (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml:6`). Rows state what is unresolved and why; no row carries a recommendation. `Operator verdict:` lines are empty at authoring.

### Q1: Where does the launcher debug trace go — stderr, a `.moai/logs/` file, or both?

Label: DECIDED
Authority anchor: SPEC-CODEX-LANE-SLOTS-001, `spec.md` §B.5 REQ-012 (pre-seam stderr report discipline; committed implementation `internal/cli/factory_launch_timing.go:85-101` and print sites `internal/cli/codex_launcher.go:1093,1105`); destination precedent `internal/cli/logging.go:78-83` (non-hook subcommands → stderr; `.moai/logs/` writers are for background processes owning no terminal).
Why unresolved: (none — the card delegated the choice with a justify-and-follow-precedent instruction; the launcher precedent and the exec-seam constraint decide it. Recorded here because the card named it as an open axis.)
Operator verdict:

### Q2: Which debug token spellings does the uniform surface accept — `-d` only, `--debug` only, or both?

Label: FOUNDER
Authority anchor: (none — no committed policy or prior SPEC row fixes launcher flag-spelling policy for a token the child CLI does not itself support.)
Why unresolved: cc/glm today forward `-d` (and claude also accepts `--debug`), but codex rejects both (`error: unexpected argument '-d' found`, verified 0.159.3), so "uniform spelling" can only mean the LAUNCHER consumes the token; whether the consumed surface is `-d` alone, `--debug` alone, or both is a user-facing contract choice. The SPEC authors both spellings; the operator may narrow it.
Operator verdict:

### Q3: Is `RUST_LOG=debug` the correct codex child-linkage form, and must it be verified against codex documentation before run-phase?

Label: EVIDENCE-NEEDED
Authority anchor: (none — the evidence itself does not exist locally: codex 0.159.3's help surfaces document no logging flag or env var; the upstream config-docs fetch returned non-codex content; the binary contains the `RUST_LOG` literal without proving wiring.)
Why unresolved: choosing between env injection and a pass-through flag depends on what the codex CLI actually supports, and that data is not obtainable from this machine today. The SPEC ships the tracing-standard fallback (`RUST_LOG=debug`, operator-set value preserved) with the launcher trace as the primary surface; an operator with authoritative codex docs can override before or during run-phase.
Operator verdict:

### Q4: Is debug granularity one level, or per-axis (flag/env/timing separable)?

Label: FOUNDER
Authority anchor: (none — no committed policy fixes debug-surface granularity.)
Why unresolved: the card listed per-axis granularity as a decision the SPEC must make explicitly. Single-level v1 is the minimum that satisfies all three axes; per-axis flags (`-d=timing` shape) are deferred to Out of Scope. An operator who debugs exactly one axis repeatedly may want the split now rather than later.
Operator verdict:

### Q5: What does a debug token on a readout verb (`moai codex status`) do — refuse, or activate-and-ignore?

Label: FOUNDER
Authority anchor: (related precedent, not the identical question: the `--spawn` readout refusal, `internal/cli/codex_launcher.go:70-72` — committed code behavior from a prior launcher SPEC, mirrored by REQ-003.)
Why unresolved: refusing matches the spawn-flag discipline but makes `-d status` an error; activating-and-ignoring is friendlier but silently drops the operator's intent. The SPEC mirrors the spawn refusal; the operator may prefer the lenient form.
Operator verdict:

### Q6: Is Tier M the correct classification?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (the classification table: M = 300–1000 LOC guidance, 5–15 files, 3-file artifact set, 16/16 REQ/AC ceilings).
Why unresolved: (none — the classification criteria are committed policy; the measured shape fits M: launcher parse + trace + env changes across `internal/cli` and `internal/config`, estimated within the M band. The card delegated the judgment with rationale.)
Operator verdict:
