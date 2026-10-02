# SPEC-CC-ULTRACODE-TOGGLE-001 — Implementation Plan

Milestones are ordered by decision-reversibility: the canonical wording (the decision most likely to change) comes first, mechanical mirroring and verification last. No time estimates; complete each milestone before starting the next.

## §A Context

Wording-only correction across 18 files (2 rule files + 4 docs pages x 4 locales). No Go source changes. Facts F1-F6 and open questions OQ-1..OQ-3 are fixed in `spec.md` Context.

## §B Known Issues / Traps

- The `--effort ultracode` launch flag and SDK `effortLevel: "ultracode"` DO set `xhigh` (F3). A blanket "never xhigh" rewrite would be a new false claim; the rule-source bullet therefore keeps one `xhigh` mention, inside the launch-flag exception. ACs pin forbidden coupling phrases, not the bare word.
- "Applies to the current session only and resets in a new session" is still true for the `/effort` route (F5). Keep it; add the `ultracode` settings key as the persistent route. Do not extend it to the slider toggle (OQ-1).
- `docs-site/.locale-parity-baseline` already lists `advanced/ultracode-workflows.md` as divergent (ko 24 headings vs 20). Edit prose only; add no heading.
- ko `ultracode-workflows.md` L65 says "세 가지가 함께 바뀝니다" (three things change together) and then lists the `xhigh` raise as the first. Removing that bullet breaks the count sentence; the sentence must be reworded in the same edit.
- `make build` does NOT regenerate rule mirrors: `internal/template/scripts/gen-catalog-hashes.go` has zero matches for `rules` / `dynamic-workflows` (measured at `c50da9c2f`), and the mirror is embedded by `//go:embed all:templates` (`internal/template/embed.go`). The mirror is therefore hand-edited identically; `cmp` proves parity, and `make build` + `go test ./internal/template/...` prove the embed still compiles and passes.

## §C Pre-flight (run-phase, before the first edit)

1. Re-run the RED-now commands of the `acceptance.md` Evidence ledger (AC-001..AC-008 cells) on the run tree and confirm the printed counts and exit codes match; re-pin the tree SHA if the base moved.
2. Confirm no foreign session writes the tree (pre-edit sync check per `agent-common-protocol.md`).

## §D Constraints

- 4-locale same-change-set obligation (`hns-oss-docs-i18n-rules` §2); ko canonical, en/ja/zh derived (§1).
- No emoji in body text, Mermaid TD-only, URL blacklist (`hns-oss-docs-i18n-rules` §3-§6).
- Literal tokens verbatim in every locale: `v2.1.284`, `/effort ultracode off`, `--effort ultracode`, `"ultracode": true`.
- Non-English locale text must read as native prose (`native-idiom-and-register.md`); docs register is the clean written register.

## §E Milestones

### M1 — Canonical wording in the rule source (highest change likelihood)

Edit the L111 bullet of `.claude/rules/moai/workflow/dynamic-workflows.md` per REQ-001..REQ-003. This is the English wording every derived surface follows, so it is settled first. Retain the per-prompt-vs-session distinction, the "every task then uses more tokens" caution, and the `ultrathink.`-does-not-restore sentence (adding the settings-key route to it).

Exact phrases the bullet must contain verbatim (these are the AC-001/AC-002 positive pins and the REQ-001..REQ-003 text; copy them, do not paraphrase):

| Phrase | Pin | REQ |
|--------|-----|-----|
| `independent on/off toggle` | R3 | REQ-001 |
| `leaves the effort level unchanged` | R2 | REQ-001 |
| `v2.1.284` (and, on the same line, `"ultracode": true`) | R4, R11 | REQ-001, REQ-003, DEC-1 |
| `/effort ultracode off` | R5 | REQ-002 |
| `--effort ultracode` | R6 | REQ-002 |
| `starts the session at `xhigh`` (the launch-flag clause; the only `xhigh` in the rule file — AC-001 `[R1s]`/`[R1c]` count it) | R7 | REQ-002 |
| `current session` | R10 | REQ-003 |
| `"ultracode": true` | R8 | REQ-003 |

Suggested wording of the flag clause: "the `--effort ultracode` launch flag (and the Agent SDK `effortLevel: "ultracode"` value) also starts the session at `xhigh`." The bullet must contain exactly one `xhigh`, and it is this one — AC-001's structural guard (`[R1s]` no line with two `xhigh`, `[R1c]` exactly one line with `xhigh` in the file) plus AC-002's `[R7]` enforce it by counting, not by wording, so no other sentence of the bullet (or file) may mention `xhigh` at all, however it is phrased. The retired verb-list pin is explained in acceptance.md AC-001; the rewording probe and its results are in acceptance.md § Mutant probe.

### M2 — ko canonical docs, then en/ja/zh derivation

Order within the milestone: ko of all four pages (workflows row, multi-llm comment, ultracode-workflows, commands) -> en -> ja and zh. Per page, edit only the lines in the spec §3 change map. Keep the table row to one line and the code comment to one line. Re-read each edited line after the edit (`completion claims need post-edit readback`).

Workflows-row literals (AC-004): each locale's row carries `v2.1.284`, `/effort ultracode off`, `--effort ultracode`, and `"ultracode": true` (the last with `v2.1.284` on the same line), retains its current-session scope wording, and holds exactly one `xhigh` — inside the `--effort ultracode` clause, in one clause with that literal and no `. 。 ; ； , ， 、` between the literal and `xhigh` (so the old leading "xhigh + orchestration" opening must go, and the flag clause is worded without internal commas). The ultracode-workflows pages name `/effort ultracode off` as the off route and lose their `xhigh` effects mention without gaining another (REQ-007; `[U1c]` en/ja/zh `0`, ko `4`). The commands pages carry a line with `ultracode` and the locale's toggle word (REQ-008); the ko/en L137 level-list sentence is reworded so `ultracode` is no longer listed among the levels. No page mentions the slider (REQ-011).

### M3 — Template mirror (mechanical)

Apply the identical M1 edit to `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` and prove `cmp` exit 0. Edit the local file and the mirror in the same step; the pair is byte-identical today (measured: `cmp` exit 0 at `c50da9c2f`).

### M4 — Verification

1. `cmp .claude/rules/moai/workflow/dynamic-workflows.md internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` -> exit 0.
2. `make build` -> exit 0; then `go test ./internal/template/...` -> exit 0 (scope the test run to the template package; CI runs the full suite).
3. docs-site exit gate from `hns-oss-docs-verify` §1-§4 and the 4-locale section-count parity: `cd docs-site && hugo --minify --gc` warn-free, URL-blacklist grep empty, Mermaid LR/RL grep empty, heading counts equal to the baselines in `acceptance.md` AC-008 (the hugo/URL/Mermaid gate itself is AC-009).
4. All AC commands of `acceptance.md`, judged on printed count (a `grep -c` of `0` exits 1 and is a PASS on zero-match assertions).

## §F Risks

| Risk | Mitigation |
|------|------------|
| A locale drifts from ko (fact omitted or translated differently) | per-locale ACs with identical pins; REQ-009 literal-token rule |
| Over-correction to "never xhigh" | REQ-002 launch-flag exception + AC pinning `--effort ultracode` |
| Slider persistence asserted by accident | REQ-011; OQ-1 stays open |
| Scope creep into session-handoff/model-policy | REQ-012 + spec §4 classification with reasons |

## §G Anti-Patterns

- Editing only the card-named files and leaving `advanced/ultracode-workflows.md` still saying "reasoning effort set to xhigh" next to a corrected table row.
- Translating `/effort ultracode off` or the settings key literal.
- Hand-editing only the local rule copy and relying on `make build` to sync the mirror (it does not).

## §H Cross-References

- `.claude/rules/moai/workflow/dynamic-workflows.md` § MoAI Integration Notes (target)
- `.claude/skills/hns-oss-docs-i18n-rules/SKILL.md`, `.claude/skills/hns-oss-docs-verify/SKILL.md`
- Neighbouring wording-correction SPEC shape: `.moai/specs/SPEC-CC-GD124-001/`
