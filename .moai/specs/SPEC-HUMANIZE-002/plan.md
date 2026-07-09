# SPEC-HUMANIZE-002 — Implementation Plan

Tier M. Doc-only (markdown + catalog.yaml). Route A (Hybrid Trunk main-direct: commit + push to `main`, no PR unless `--pr`).

## §A Context

- **Repo**: `/Users/goos/MoAI/moai-adk-go`, branch `main`, plan baseline SHA `39c74d77787621b6645aebe81e470277ba3c97cb`.
- **SPEC artifacts**: `.moai/specs/SPEC-HUMANIZE-002/{spec.md, plan.md, acceptance.md, progress.md}` (Tier M set + progress).
- **Target (Template-First order)**:
  1. `internal/template/templates/.claude/skills/moai-domain-humanize/modules/copy-review.md` (NEW)
  2. `internal/template/templates/.claude/skills/moai-domain-humanize/modules/design-copy.md` (NEW)
  3. `internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md` (EDIT — genre routing + version 1.2.0)
  4. `internal/template/catalog.yaml` (EDIT — humanize entry 1.1.0 → 1.2.0; hash via `make build`)
  5. Local mirror `.claude/skills/moai-domain-humanize/{SKILL.md, modules/copy-review.md, modules/design-copy.md}` (SYNC — byte-identical)
- **Sources (read-only, outside repo)**: the four claude.mo.ai.kr files inventoried in spec.md §A.1.
- **PRESERVE (never touch)**: the 8 existing language-module files (`{korean,english,japanese,chinese}.md` × 2 trees), all other skills, all Go source, `.moai/state/*`, `.moai/harness/*`, unrelated SPEC dirs.
- **Plan-time facts** (baseline-attributed, measured at `39c74d777`): template skill dir count = 28; template↔local humanize trees byte-identical (`diff -rq` clean); catalog humanize entry `version: 1.1.0`, `tier: optional-pack:design`.

## §B Known Issues (filtered to relevant categories)

- **B4 Frontmatter schema**: SPEC artifacts use `created:`/`updated:`/`tags:` canonical names (done at plan). SKILL.md frontmatter is skill-schema, not SPEC-schema — only `metadata.version` + `metadata.updated` change.
- **B6 spec-lint heading**: spec.md carries `### Out of Scope — <topic>` H3 subsections (done); do not restructure during run.
- **B8/B10 Working-tree hygiene & scope**: commit with specific paths only (`git add <path>`); parallel sessions are active on this checkout historically — run the pre-spawn sync check (`git fetch` + `rev-list --left-right`) before commits; never `git add -A`.
- **B9 Commit/push**: manager-develop commits + pushes per milestone. Conventional Commits: `feat(SPEC-HUMANIZE-002): M<N> <subject>` + `🗿 MoAI` trailer. First run-phase commit flips frontmatter `draft → in-progress`.
- **Domain-specific pitfalls**:
  - *Neutrality leak from sources*: S1–S4 contain plugin namespaces (`moai-coworker:` etc.), `3-point sync` HTML comments, and internal tokens — strip ALL during porting; AC-H2-019 greps for them.
  - *Windowed-grep undercount*: per-language section checks MUST use flag-based awk windowing (`/^## .*Japanese/{f=1;next} /^## /{f=0} f`), never `sed -n` line windows (memory-documented failure mode).
  - *`grep -P` portability*: script-range greps (`[가-힣]`, kana) need GNU grep/`ggrep` or `rg`; stock BSD grep may mis-handle multibyte ranges — acceptance.md notes the fallback.
  - *Catalog hash*: never hand-edit the hash; run `make build` and verify `git status --porcelain internal/template/catalog.yaml` is clean post-commit.
  - *Template-leak test*: `go test ./internal/template/ -run TestTemplateNoInternalContentLeak` may be red for pre-existing unrelated leaks — the gate is "humanize dir absent from the violation list", not whole-repo green (SPEC-HUMANIZE-001 precedent).

## §C Pre-flight (before any edit)

```bash
git -C /Users/goos/MoAI/moai-adk-go branch --show-current && git rev-parse HEAD
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# Mirror parity baseline (expect: identical)
diff -rq internal/template/templates/.claude/skills/moai-domain-humanize .claude/skills/moai-domain-humanize
# Skill-count baseline (expect: 28)
ls -d internal/template/templates/.claude/skills/*/ | wc -l
# Existing-ID inventory for dedup pass
grep -oE '\b(ENC-[0-9]|JA-1[0-4]|CN-[LMNOPQ]|A-2[0-5]|L-[1-8]|M-[1-3])\b' \
  internal/template/templates/.claude/skills/moai-domain-humanize/modules/*.md | sort -u
```

## §D Constraints (DO NOT VIOLATE)

- Write whitelist: the 5 target files (§A) in template tree + 3-file local mirror + `.moai/specs/SPEC-HUMANIZE-002/*`. Nothing else.
- The 8 existing language-module files are byte-frozen (REQ-H2-019) — verified by `git diff 39c74d777..HEAD -- <paths>` empty.
- Source repo `/Users/goos/MoAI/claude.mo.ai.kr/` read-only.
- No `--no-verify`, no `--amend` on pushed commits, no force-push.
- ID namespaces: new entries use `CR-{KO|EN|JA|ZH}-N`, `CRS-1..7`, `DCG-*` only — never mint entries in the existing A/L/M/ENC/JA/CN series.
- Severity vocabulary: S1/S2/S3 only; the source "Tier 1/2" labels must not survive into shipped content.
- Blockers (e.g., JA/ZH grounding below minimum, dedup ambiguity requiring spec change) → structured blocker report to orchestrator; never AskUserQuestion, never silent scope change.

## §E Self-Verification Deliverables (run-phase completion report)

Reported per verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk):

- **E1** — AC binary PASS/FAIL matrix for all 24 checks (acceptance.md §D) with executed command + verbatim output per mechanical/hybrid AC.
- **E2** — `make build` exit 0 (embeds templates + regenerates catalog hash); `go build ./...` exit 0 (sanity — no Go touched).
- **E3** — Coverage: n/a (doc-only); replaced by scoped template-leak check (humanize dir absent from violation list).
- **E4** — Byte-parity evidence: `diff -rq` template vs local → identical.
- **E5** — Byte-freeze evidence: `git diff 39c74d777..HEAD -- <8 language-module paths>` → empty.
- **E6** — Commit SHAs per milestone + `git push origin main` result.
- **E7** — Blocker report (if any) with 3-4 option structure.

## §F Milestones (priority-ordered, no time estimates)

### M1 — KO-base module authoring (template tree)

- Author `modules/copy-review.md`: 6-stage pipeline + input classes; review-only gate mode; fix-proposal format; report template; `CRS-1`…`CRS-7` playbook with before/after examples; `CR-KO-*` formula dictionary (≥6 entries after dedup) with S1/S2/S3 severities; explicit dedup cross-references into `korean.md` (M-1/M-3/A-20…/L-*/prose D/J) — zero re-definitions.
- Author `modules/design-copy.md`: `DCG-*` landing-page rules (structure constraints + vague-claim → concrete-mechanism repair table + element checklist + tone profile) and short-form card/slide rules; KO adaptation block.
- Neutrality discipline applied at authoring time (no plugin namespaces, no sync comments, no internal tokens).
- Exit: AC-H2-003/004/005/006/008 KO-portion green; dedup evidence recorded.
- Commit: `feat(SPEC-HUMANIZE-002): M1 copy-review + design-copy KO-base modules` (flips `draft → in-progress`).

### M2 — 4-language expansion

- `CR-EN-*` (≥8): port S2/S3 EN formula dictionary + structural anti-patterns, dedup vs ENC-1…9 (cross-reference overlaps such as Unleash/Transform → ENC-1, No-more-X → ENC-2, A-B-C-D listing → ENC-3); web-ground any new additions.
- `CR-JA-*` (≥4) and `CR-ZH-*` (≥4): independently grounded language-native entries; each non-v1.1.0-mapped entry carries a grounding note in the module sources section; cross-reference JA-11/JA-13/CN-N/CN-Q where the tell already exists.
- Per-language adaptation blocks in `design-copy.md` for all 4 languages (native length measures — no verbatim KO character limits in EN/JA/ZH).
- All examples in target script; instruction prose English; per-language headings normative (`## … Korean|English|Japanese|Chinese`) for AC windowing.
- Exit: AC-H2-007/009 (full), AC-H2-011/012 green; AC-H2-013 grounding trail complete.
- Commit: `feat(SPEC-HUMANIZE-002): M2 4-language formula dictionaries + genre adaptations`.

### M3 — Integration, parity, close-out

- `SKILL.md`: genre-module routing extension (copy-review gate routing + design-copy display-surface routing; Language Routing table untouched); `metadata.version` + footer → 1.2.0; `metadata.updated` refresh.
- `internal/template/catalog.yaml`: humanize `version: 1.2.0`; `make build` to regenerate hash; verify clean porcelain after commit.
- Sync local tree (3 files) → `diff -rq` byte-identical.
- Sweeps: neutrality 7-class grep set → 0; license greps (zero MIT, `license: Apache-2.0` unchanged, im-not-ai credit intact); byte-freeze diff on 8 language modules; skill-count check (28).
- Run full AC matrix (24 checks) + 3 manual scenarios; populate progress.md §E.2/§E.3.
- Commit: `feat(SPEC-HUMANIZE-002): M3 SKILL routing + catalog 1.2.0 + mirror sync`; push.

## §G Risks & mitigations

| # | Risk | Mitigation |
|---|------|------------|
| R1 | JA/ZH web-grounding honestly yields < minimum entries | Blocker report proposing threshold amendment; fabrication prohibited (REQ-H2-015) |
| R2 | Dedup ambiguity (formula overlaps an existing category only partially) | Conservative default: cross-reference the existing ID + add a delta note inside the new module's entry prose; never re-define |
| R3 | Script-range grep portability (BSD grep) | Use `rg`/`ggrep`; fallback documented in acceptance.md preamble |
| R4 | Parallel-session race on shared checkout | Pre-spawn sync check; specific-path commits; re-fetch before each push |
| R5 | `make build` churns unrelated catalog hashes | Inspect `git diff internal/template/catalog.yaml` — only the humanize entry (+its hash) may change; otherwise halt and report |
| R6 | Over-porting (absorbing source content that duplicates SKILL.md machinery, e.g., output contracts) | New modules reference SKILL.md shared sections; the dedup pass covers machinery, not just entries |

## §H Cross-references

- SPEC-HUMANIZE-001 artifacts (`.moai/specs/SPEC-HUMANIZE-001/`) — precedent for AC style, neutrality command set, license gates, 8-sample scenario pattern.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter + status transition ownership.
- `CLAUDE.local.md` §2.1 / `.moai/docs/template-internal-isolation-doctrine.md` §25.1 — neutrality content classes.
- `.claude/rules/moai/core/verification-claim-integrity.md` — evidence format for E1–E7.
