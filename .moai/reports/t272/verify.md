# verify.md — SPEC-SKILL-GALLERY-BENCH-001 M3 recipe results

Exit-gate run of Skill("hns-oss-docs-verify") over the touched surfaces,
2026-08-25, worktree `.claude/worktrees/t272`, branch `WT-skillstead-gallery`,
HEAD `71781683c` (origin/main fast-forward applied by the lead before M2).
Every check below states the command run, the verbatim key output observed,
and the verdict — no unobserved pass. Two recipe commands could not be issued
verbatim because the worktree guard rejects `for` loops and `$(...)`
substitution; each was replaced by an equivalent decomposition, named inline.

## §1 build-clean (must_pass) — PASS

Command: `cd docs-site && hugo --minify --gc > /tmp/t272-hugo-verify.txt 2>&1; echo "exit=$?" >> /tmp/t272-hugo-verify.txt`

Observed (verbatim, `/tmp/t272-hugo-verify.txt`):

```
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew
              │ KO  │ EN  │ JA  │ ZH
 Pages        │ 184 │ 182 │ 182 │ 182
 Static files │ 265 │ 265 │ 265 │ 265
Total in 2724 ms
exit=0
```

Zero `WARN`/`ERROR` lines in the full output (the only lines are the version
banner, the page-count table, `Total in 2724 ms`, and the exit line).

Command: `test -f docs-site/public/sitemap.xml && echo "sitemap OK" || echo "sitemap MISSING"`
Observed: `sitemap OK`

## §2 URL blacklist (content-fidelity) — PASS

Command: `grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content README.md README.ko.md README.ja.md README.zh.md`
Observed: no output (0 matches).

## §3 Mermaid direction (style-compliance) — PASS

Command: `grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content`
Observed: no output (0 matches). This branch also adds zero new Mermaid
blocks (measured: the docs diff adds prose + one table per locale).

## §4 locale-parity (must_pass) — PASS

File-existence parity. Recipe's `for`-loop form is rejected by the worktree
guard; equivalent decomposition issued:

```
find ko -name '*.md' | sed 's|^ko/||' | sort > /tmp/t272-pages-ko.txt   (etc. per locale)
comm -23 /tmp/t272-pages-ko.txt /tmp/t272-pages-en.txt   (and ja, zh)
```

Observed: `150 /tmp/t272-pages-ko.txt`, `150 …en`, `150 …ja`, `150 …zh`;
all three `comm` outputs empty → every ko page has its en/ja/zh counterpart.

Section-count ratchet (verbatim recipe awk pipeline):

```
grep -rc '^#\{2,\} ' ko en ja zh --include='*.md' | awk … | sort > /tmp/parity-now.txt   → 54 lines
grep -v '^#' docs-site/.locale-parity-baseline | grep -v '^[[:space:]]*$' | sort > /tmp/parity-base.txt → 54 lines
comm -23 /tmp/parity-now.txt /tmp/parity-base.txt   → empty (0 NEW divergence)
comm -13 /tmp/parity-now.txt /tmp/parity-base.txt   → empty (0 converged; informational)
```

Observed: both sets 54 lines, byte-identical membership — the branch's +1
heading per locale on `advanced/skill-guide.md` left every page in exactly its
prior parity state (the page was already baselined as divergent).

README 4-file heading parity:

Command: `grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md`
Observed: `12 / 12 / 12 / 12`. H2 order spot-checked (ko: v3.1 새 기능 → 왜
moai-adk인가요? → 빠르게 시작 …). Switcher header verified present in all 4
files (links to all three siblings from each).

## §5 body-emoji scan (style-compliance) — PASS

Command: `grep -rnP '[\x{1F300}-\x{1FAFF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]' docs-site/content --include='*.md' | grep -v '{{<' | head -40`
Observed: hits exclusively in untouched, pre-existing pages
(statusline example outputs, `✓`/`✗` typographic symbols in tables,
`✂` handoff markers, spinner glyphs in agent-view examples,
`🤖 Generated with Claude Code` commit-footer examples) — all inside the
recipe's allowed classes (typographic symbols; example-output code context).

Attribution scans for this branch:

- Command: `grep -rnP '[…same class…]' docs-site/content/{ko,en,ja,zh}/advanced/skill-guide.md | grep -v '{{<'`
  Observed: no output (0 hits in the touched pages).
- Command: `git diff origin/main -- docs-site/content README*.md | grep "^+" | grep -cP '[…same class…]'`
  Observed: `0` (zero emoji on any added line of the branch's docs diff).

## §6 version-sync (must_pass) — PASS for this diff / pre-existing absolute FAIL flagged

Command: `grep -E 'version = |releaseDate' docs-site/hugo.toml`
Observed: `version = "v3.1.2"`, `releaseDate = "2026-08-21"` (SSOT).

Command: `grep -rn '🗿 v[0-9]' docs-site/content README*.md` (direct form; the
recipe's `$(…) `-substituted filter is rejected by the worktree guard — the
substitution only removes lines matching the SSOT string, and every observed
line already reads v3.1.2)
Observed: all `🗿 v…` displays read `v3.1.2` = SSOT → zero stale statusline
examples after the filter.

Command: `grep -rn 'Release-v[0-9]' README.md README.ko.md README.ja.md README.zh.md`
Observed: 4 badge lines reading `Release-v3.1.3` — NOT equal to the hugo.toml
SSOT (v3.1.2). Per the recipe's literal rule this is a stale display.

Attribution (measured, not assumed):

- `git show origin/main:README.ko.md | grep -n "Release-v"` → the same
  `Release-v3.1.3` badge is on origin/main; `git show origin/main:docs-site/hugo.toml`
  → the same `version = "v3.1.2"`. The divergence pre-exists this branch.
- `git diff origin/main -- <8 touched files> | grep "^+" | grep -c "Release-v\|🗿 v"`
  → `0`. This branch adds zero version displays (the plan.md §B.3 sidestep).

Routing: the v3.1.2→v3.1.3 badge/SSOT reconciliation belongs to the in-flight
v3.1.3 release workstream, whose release-sync obligation updates hugo.toml
version+releaseDate and the 4 README badges together. Decision routed to the
orchestrator with this report: gate this PR on the diff-attributed PASS, or
hold for the release PR.

## Dimension roll-up

| Dimension | Checks | Threshold | Result |
|-----------|--------|-----------|--------|
| build-clean | §1 | 1.0 must_pass | 1.0 PASS |
| locale-parity | §4 | 1.0 must_pass | 1.0 PASS |
| style-compliance | §3 + §5 | 0.95 | 1.0 PASS (both clean; added lines 0 emoji) |
| content-fidelity | §2 | 0.9 | 1.0 PASS (0 blacklist matches) |
| version-sync | §6 | 1.0 must_pass | diff-attributed PASS; absolute pre-existing FAIL flagged (origin/main badge v3.1.3 vs SSOT v3.1.2 — not introduced by this branch, zero version displays added) |
