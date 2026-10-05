# t345 verdict — policy-rule application evidence (the citation convention + dormancy recipe)

Card: t345 (G2a, lane-9, Tier M) · Branch: WT-policy-rule-evidence · Base: develop `1cfd9e8a0` + merge of prior card branch `WT-audit-verdict-owner` (pre-empts the §6/§7 same-file merge conflict — both cards append at the file tail)

## Claim

The card's question — "정책 층 규칙이 실제로 인용되고 적용되는지를 무엇이 확인하는가 — 규칙에는 실행 기록이 없다" — now has a three-part answer landed in the doctrine that governs verification-artifact authoring: **(1)** a [HARD] citation obligation (a decision resting on a policy-layer rule names `<rule-file>.md §x.y` in the artifact carrying the decision); **(2)** an on-demand dormancy recipe (sweep decision artifacts for citations of a rule; zero-citation windows are a dormancy signal, not a proven defect); **(3)** an honest limit statement (citation proves observability + reviewability, never obedience). Section: `verification-completeness.md` §7, byte-identical in both trees, Version 1.1.0 → 1.2.0.

## Evidence

1. **The premise, measured**: `find .moai/reports -name verdict.md | wc -l` → 106 (t302 tree) and 108 (this tree); citations of the two load-bearing policy rules: 3/106-class for `verification-claim-integrity.md`, 3→5 for this rule across the two trees — ~97% of decision artifacts carry no application record for the rules their own report format is built on. Full numbers + citing-file lists: `.moai/reports/t345/recipe-baseline.txt`.
2. **The recipe, exercised (fires demonstration)**: the §7 sweep command was run on this tree and named its citing artifacts (t241, t302, t333, t344, t371 for this rule; t156, t190, t308 for VCI). The recipe returns a READABLE dormancy reading — not a binary verdict — exactly as the section specifies. Baseline-attribution note: the citation counts MOVE between measurement trees (106/3/3 → 108/5/3), which is why the doctrine's evidence block was reworded to generic phrasing (exact counts live here, not in the shipped template — the file's own §1-6 evidence convention is generic prose).
3. **The convention's first instances (organic, not forced)**: the G2a card verdicts already carry clause-level citations — t302's verdict cites `verification-completeness.md §1.3` for the sentinel finding; t344's cites `verification-claim-integrity.md §1.1 surface 3/4` for the refuted-adoption judgment — this card's rule records what this batch was already doing, and makes future silence visible instead of invisible.
4. Both trees edited (template source first per Template-First, then byte-parity local mirror); `diff -q` → IDENTICAL; `make build` → EXIT=0.

## Baseline-attribution

All measurements this run, this tree (WT-policy-rule-evidence @ merge of develop `1cfd9e8a0` + `WT-audit-verdict-owner`):
- premise counts: `find`/`grep -rl | wc -l` commands as captured in recipe-baseline.txt, this run.
- recipe output: same file, citing-file lists verbatim, this run.
- parity: `diff -q` both copies → IDENTICAL, rc=0, this run.
- build: `make build` → exit 0.

## Gaps (explicitly NOT observed)

- **Obedience is not observed by design** — §7 says so itself; a citation can decorate a violation. The residual is owned by review.
- No gate/hook was wired for the citation obligation (deliberate — the dormancy recipe is on-demand; wiring a blocking gate for prose citations was rejected as over-engineering with a false-positive surface, same reasoning as t300's rejected commit-topology gate).
- The recipe was exercised on two rules only; other policy rules' dormancy readings are unmeasured here (the recipe is the tool, not a one-time sweep).
- citation-based dormancy cannot distinguish "rule not loaded" from "loaded and not applied" without reading an artifact — §7 states this three-way ambiguity as a feature of the signal, not a resolved defect.

## Residual-risk

- The convention binds decision authors through the rules system's loading path (paths-scoped: `**/.moai/specs/**`, `.claude/rules/**`, hooks, scripts). A decision artifact authored outside those paths (e.g., a chat-only dispatch) is outside the rule's reach — citation discipline there remains custom.
- §6/§7 same-file sequencing: this branch already merged `WT-audit-verdict-owner`, so the develop-side merge order is settled for the pair; a THIRD card touching this file's tail would re-open the sequencing question.
