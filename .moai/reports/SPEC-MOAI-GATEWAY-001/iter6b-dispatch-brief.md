# iter6b dispatch brief — SPEC-MOAI-GATEWAY-001 0.7.0 amendment (GLM slot keys)

Orchestrator dispatch record (session 650cbd95, 2026-09-11). Not an audit report.

The 0.7.0 pass (brief `iter6-dispatch-brief.md`) is written, lint-clean, and **not yet audited**. Its writer left one contradiction unresolved; the operator has now decided it. Amend 0.7.0 in place: **keep `version: "0.7.0"`**, and extend the existing HISTORY 0.7.0 entry and the progress.md 0.7.0 line as the final writes. Replace contradicting sentences; do not annotate beside them. SPEC prose stays Korean in clean native written register.

At start, invoke Skill("moai-workflow-spec") for the GEARS format.

## 0. Tree, ownership, prohibitions

- Worktree root / CWD: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. Branch `WT-unified-gateway`, HEAD `81c1d58f9`.
- **You are the sole writer of `.moai/specs/SPEC-MOAI-GATEWAY-001/`.** Do not message any agent. Do not touch code or `.moai/reports/`. No state-changing git.
- Use `/usr/bin/grep`, never the shell `grep` (ugrep wrapper that silently skips files). Open every cited code line before citing it; trust code over this brief and report differences.
- Budget fixed: 24 live REQ / 24 AC. **No new REQ or AC number.**
- If an item needs an operator decision not given here, return a blocker report for that item and finish the rest.

## 1. The contradiction (as the 0.7.0 writer reported it, re-checked by the orchestrator)

- 0.7.0 REQ-MG-021 "모델 슬롯" bullet (`spec.md:578-583`) scrubs inherited `ANTHROPIC_DEFAULT_*_MODEL` and does not require the launcher to set slot keys.
- REQ-MG-018 (`spec.md:507-509`, `design.md:567`) only says tier mapping moves "inside the gateway"; no mechanism is written anywhere. AC-MG-025 (`acceptance.md:437-440`) judges mapping only for requests that already carry a GLM model ID.
- Today `moai glm` puts GLM IDs into the slot keys (`internal/cli/glm.go:364-370`, `setGLMEnv`: OPUS←`Models.High`, SONNET←`Models.Medium`, HAIKU←`Models.Low`, FABLE←`Models.Fable`), so subagent alias requests (`opus`/`sonnet`/`haiku`) carry GLM IDs. With no slot keys, those requests would carry Claude Code's default Claude IDs, and the exact-match registry (REQ-MG-011) would route them to Anthropic — a `moai glm` user's subagents silently on Claude (inference from Claude Code's documented slot variables; not run).

## 2. Operator decision (2026-09-11): the gateway `moai glm` launcher sets the GLM slot keys

Write:
1. **REQ-MG-021 "모델 슬롯" bullet** — rewrite: after the 14-key scrub, the gateway `moai glm` launcher adds `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` with the configured GLM tier IDs (same mapping as today's `setGLMEnv`), so alias-tier requests carry GLM IDs and resolve by exact match to the GLM route. `moai cc` and `moai gpt` add no slot keys. Picker composition, the "Default" row risk, and slot keys for cc/gpt remain `SPEC-MOAI-GATEWAY-PICKER-001` (proposed).
2. **REQ-MG-018** — replace "tier별 모델 매핑은 gateway 안쪽 책임" with the actual mechanism: the launcher-set slot keys plus the registry's exact-match routing of the configured GLM IDs; Anthropic beta header removal stays inside the gateway. Keep the in-process teammate exception and the TEAMMATE pointer for tmux-pane tier mapping. Update `design.md:567` and the `spec.md:580` cross-reference to match.
3. **Catalog precondition — verify, do not assume.** Exact match only works if the session catalog registers the configured GLM tier IDs with the Z.AI route. Locate where the SPEC defines catalog assembly (REQ-MG-019, design registry/catalog sections, M2/M3). If it already registers the `llm.glm.models` values, cite it. If not, add that requirement inside REQ-MG-019 or REQ-MG-018 text (no new number). If the catalog's GLM IDs come from somewhere that conflicts with `llm.glm.models`, report it rather than choosing.
4. **design.md** — §3.2 step 6b: add the glm slot keys to the launcher additions. §6.1 process-env projection: note that for `moai glm` four of the scrubbed keys are re-added by the launcher with configured values. §6.7: narrow it — slot keys for gateway `moai glm` tier mapping are now core (REQ-MG-021/REQ-MG-018); picker composition, the "Default" row risk, and cc/gpt slot use stay with PICKER. Keep the name-sharing-with-the-cleanup-set note, now stated precisely (scrub first, then add).
5. **AC-MG-018 (a) inherited-key judgment** (`acceptance.md:260-285`) — for `moai glm`, the four slot keys are present and their values equal the configured tier fixture values (distinct from the inherited markers). For cc/gpt keep the inherited-marker-absence judgment. Rewrite the explanatory paragraph that says PICKER "may decide" slot keys: glm sets them now; for cc/gpt PICKER may. Add a mutant: a glm builder that omits the slot keys (or copies inherited ones) is red.
6. **AC-MG-025** — add that, for a `moai glm` gateway session built from a `llm.glm.models` fixture, each configured tier ID resolves in the session registry to the Z.AI mock route (and a Claude default model ID does not resolve to the Z.AI route). Keep the existing header/auth/subcommand judgments and the Gap wording.
7. **plan.md** — M7 "Claude child env의 상속 키 정리" bullet (`:234-238`): add the glm slot keys. Decision 11 PICKER list names `design.md` §6.7 — adjust so it no longer implies all slot keys moved. Decision 12's 0.7.0 bullet: one clause on the glm slot keys.
8. **§E** — one line: that Claude Code sends the slot values for subagent alias requests is documented client behaviour, not measured here.
9. Keep every other 0.7.0 edit intact. Re-check consistency with REQ-MG-022 and design §6.5 (a stale tmux slot key in a new pane is still the TEAMMATE residual).

## 3. Final writes

- HISTORY 0.7.0 entry: append the slot-key decision (2026-09-11, operator) and what changed.
- `progress.md` §E.1 0.7.0 line: append the same in one clause. Status stays "0.7.0 작성 완료, 감사 미실시".

## 4. Return (to the orchestrator, English)

- Per file: what changed (section + line ranges).
- Verbatim output of `moai spec lint SPEC-MOAI-GATEWAY-001`, live REQ/AC counts (`/usr/bin/grep -cE '^\*\*REQ-MG-[0-9]{3}\*\* \(' spec.md` and the AC equivalent), and `wc -c` of the six artifacts.
- The catalog-precondition finding (item 3) with citations.
- Every code-vs-brief difference and any unresolved contradiction.
