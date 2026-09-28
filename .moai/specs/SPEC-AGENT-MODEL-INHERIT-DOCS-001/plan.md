# plan.md — SPEC-AGENT-MODEL-INHERIT-DOCS-001 (card t1300)

## A. Context

Card t1300 (Tier M, Class C) is the docs-layer split of t1246 M8: after
SPEC-AGENT-MODEL-INHERIT-001 landed the inheritance model in product and rules, the public docs
still teach the retired per-agent profile matrix. The phase-① survey
(`.moai/reports/t1300/phase1-survey.md`) measured the scope (155 pages × 4 locales = 620) and
classified every touched page into A (deep rewrite), B (flag/CLI docs), C (minor/no change).
Plan-phase re-verification on tree `8a969dfc0` confirmed the survey's A-cluster line maps and
resolved the one open question — `cli-reference/profile.md` documents the surviving
`moai profile` family, not the deleted accessor.

Baseline branch: `WT-model-docs-sweep` in worktree `.moai/worktrees/t1300`, local develop
`afecf81e9..8a969dfc0` absorbed.

## B. Known Issues

- Line numbers in the survey are en-locale and pre-absorb; t1257 touched
  `cli-reference/launchers.md` and `tokens.md` only, so A-cluster line numbers remain valid — but
  every milestone re-verifies against the current tree before editing (grep-first, not line-blind).
- The `claude-code/context-memory/prompt-caching.md` sibling carries 0 per-spawn-injection hits on
  en; other locales are unmeasured — M3 verifies all four before declaring no-change.
- ko/ja/zh A-cluster pages were existence- and symmetry-checked but not line-mapped in phase ①;
  M1 performs the en edit-map → locale transfer at edit time.

## C. Pre-flight

1. Confirm the worktree is on `WT-model-docs-sweep` and HEAD matches the lead-reported SHA.
2. Confirm `internal/cli/model.go` is still absent and init.go still registers the deprecated
   stubs (REQ premise).
3. Run the phase-① RED-now greps (acceptance.md baseline table) and record any drift before
   editing.

## D. Constraints

- [HARD] 4-locale same-change obligation: every content edit lands in ko/en/ja/zh in the same PR,
  authored ko→en→ja/zh (ko is canonical source).
- [HARD] Rewrite wording DERIVES from the canonical doctrine sentence
  (`agent-common-protocol.md` § "Subagent Model and Effort") and the measured CLI surface —
  inventing a parallel narrative is prohibited.
- [HARD] Icon shortcode (`{{< icon >}}`) for icons; no body emoji. Mermaid diagrams stay TD-only.
  Emphasis-marker spacing per `.moai/docs/docs-site-i18n-rules.md` §17.2. URL blacklist §17.1.
- [HARD] Korean copy: clean native written register (문어), no translationese.
- [HARD] H24 (Go strings) is a separate milestone and a separate delegation from the docs-site
  milestones — it does not ride a docs commit.
- No time estimates; priority labels and phase ordering only.

## E. Self-Verification

The run phase closes with the `hns-oss-docs-verify` skill recipe (warning-free hugo build,
sitemap, URL-blacklist grep, Mermaid direction grep, 4-locale file-existence + section parity,
README-heading parity N/A, body-emoji scan) plus `go build ./...` and
`go test ./internal/cli/wizard/... ./internal/cli/...` (affected packages). Evidence per
verification-claim-integrity §3: command + verbatim output, attributed to the measured tree.

## F. Milestones

Ordered by decision-reversibility — the narrative decision (M1) is the highest-change-likelihood
work and leads; mechanical string and verification work trails.

- **M1 — A-cluster inheritance rewrite (Priority High)**: 5 pages × 4 locales = 20 files.
  `multi-llm/model-policy.md` (sections en 89/115/157/191/222/260/289/304/355/369/389 — the
  per-agent assignment table at 115, the `moai model profile` inspection commands at 211-216, the
  resolver narrative at 191-238, the drift/enforcement story at 289-355 all re-told as
  inheritance); `advanced/profile-matrix.md` (30/87/108/119 — page re-centered from matrix to
  resolution narrative, URL preserved); `advanced/agent-guide.md` (drop the "Model / effort" table
  columns at 46/57/66/72/78/91; rewrite frontmatter narrative 96-98); `advanced/no-haiku-3tier.md`
  (93/97/144/160 — tier narrative aligned to inheritance; 39-cell matrix link target replaced);
  `advanced/tokenomics-overview.md` (36/51/106/123/125 — Layer B routing narrative;
  profile-matrix links at 27/109/131 re-pointed). Per-milestone locale transfer: en map →
  ko/ja/zh equivalents located by heading grep, not by line number.
- **M2 — B-cluster deprecated-stub docs (Priority High)**: 6 pages × 4 locales = 24 files.
  `getting-started/cli.md` (en 72/116/452 — flag rows rewritten to deprecated-stub reality),
  `getting-started/init-wizard.md` (132/152/156), `getting-started/introduction.md` (53/147/152),
  `getting-started/faq.md` (93), `core-concepts/what-is-moai-adk.md` (263/522-541), and the
  `cli-reference/profile.md` residue audit (REQ-AMD-004: page retained, no removal, no redirect).
- **M3 — C-cluster sweep and disposition record (Priority Medium)**: drop/simplify the
  per-spawn-injection bullet (`cost-optimization/prompt-caching.md` en 219 + link description
  308, ×4 locales); re-verify the context-memory sibling across all 4 locales; record the
  no-change dispositions for statusline / moai-web-console / decision-memory / agent-teams /
  en `_index` captions in the milestone evidence (a no-change row needs the verification that
  produced it).
- **M4 — H24 wizard strings, Go (Priority High, independent of M1-M3)**: rewrite the 5
  `model_policy` occurrences in `internal/cli/wizard/translations.go` (en/ko/ja/zh question title
  + description) to main-session wording — the policy feeds only the main-session effort fallback
  via `MapModelPolicyToEffort` — with native wording per locale; survey §5 line map 444/456/468/480.
  `go build ./...` + affected `go test ./internal/cli/...`. Separable: M4 may land before, after,
  or between docs milestones.
- **M5 — Verification closure (Priority High)**: run the `hns-oss-docs-verify` recipe + Go
  build/test; assemble the evidence-bearing completion report (five-section format).

## G. Run-phase Delegation

- M1-M3, M5 docs work → `manager-docs`, which spawns the oss-docs harness specialists
  (content-author on ko canonical, locale-translators for en/ja/zh) and loads
  `hns-oss-docs-i18n-rules` first per the harness contract.
- M4 Go strings → `manager-develop` delegating a per-spawn `Agent(general-purpose)` CLI
  specialist (internal/cli domain); the wizard string edit is Go-source work, not docs work.
- The two delegations may proceed in sequence within the run phase (write-capable agents do not
  run concurrently in the same tree).

## H. Cross-References

- Parent: SPEC-AGENT-MODEL-INHERIT-001 (completed) — the landed inheritance model; its design §H24
  defines the wizard-string scope; its REQ-AMI-016 requires the `moai profile setup` surface the
  deprecation warnings point at.
- Sibling: SPEC-ROLE-NAMING-DOCS-001 (t1257) — same docs-site rewrite pattern, already landed.
- Sibling: card t1302 — README layer of the same sweep (absorbed; not re-done here).
- Procedure: `.moai/docs/docs-site-i18n-rules.md` (HARD i18n rules), `hns-oss-docs-verify` skill
  (exit gate), `verification-claim-integrity.md` (evidence format).
