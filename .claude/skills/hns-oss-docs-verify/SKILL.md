---
name: hns-oss-docs-verify
description: >
  Mandatory verify recipe for the oss-docs harness — the runnable exit gate
  every specialist executes before returning: warning-free hugo build,
  sitemap existence, URL-blacklist grep, Mermaid LR/RL direction grep,
  4-locale file-existence and section-count parity, README 4-file heading
  parity, and body-emoji scan. All checks are inlined here because
  docs-i18n-check.sh and gen_menu.py do not exist.
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "harness"
  status: "active"
  updated: "2026-07-13"
  tags: "oss-docs,verify,quality-gate,hugo,parity,blacklist"
---

# oss-docs Verify Recipe (exit gate)

Runnable checks for the sprint-contract dimensions. **The scripts
`docs-i18n-check.sh` and `gen_menu.py` DO NOT exist — never shell out to
them; every check is inlined below.** All checks are read-only; this skill
never commits or pushes.

## 1. Build clean (`build-clean`, must_pass, threshold 1.0)

```bash
cd docs-site && hugo --minify --gc
```

- Must exit 0 AND complete **warning-free** (any `WARN`/`ERROR` line = FAIL).

```bash
test -f docs-site/public/sitemap.xml && echo "sitemap OK" || echo "sitemap MISSING"
```

## 2. URL blacklist (`content-fidelity`)

```bash
grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content README*.md
```

- Expected: **no matches**. Only `adk.mo.ai.kr` is valid. Note: the pattern
  `adk\.moai\.kr` does not match `adk.mo.ai.kr` (different dot positions) —
  no false positive on the valid domain.

## 3. Mermaid direction (`style-compliance`)

```bash
grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content
```

- Expected: **no matches** (TD-only rule; `flowchart TD` / `graph TB` pass).

## 4. 4-locale parity (`locale-parity`, must_pass, threshold 1.0)

File-existence parity — every ko page has en/ja/zh counterparts:

```bash
cd docs-site/content && for f in $(cd ko && find . -name '*.md'); do
  for loc in en ja zh; do
    [ -f "$loc/$f" ] || echo "MISSING: $loc/$f"
  done
done
```

Section-count parity per locale tree (compare totals):

```bash
for loc in ko en ja zh; do
  printf '%s: %s\n' "$loc" "$(grep -rc '^## ' docs-site/content/$loc --include='*.md' | awk -F: '{s+=$2} END {print s}')"
done
```

README 4-file heading-count parity:

```bash
grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md
```

- Expected: identical counts across the 4 files (and identical H2 order —
  spot-check with `grep '^## ' <file>`).

## 5. Body-emoji scan (`style-compliance`)

```bash
grep -rnP '[\x{1F300}-\x{1FAFF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]' docs-site/content --include='*.md' | grep -v '{{<' | head -40
```

- Review each hit: body-text emoji = FAIL (use `{{</* icon */>}}`);
  preserved typographic symbols (`→ ← ↓ ✓ ✗`, U+2702 in handoff blocks) and
  branding emoji inside orchestrator-banner example code blocks are allowed —
  judge code-block context before flagging.

## Scoring map (sprint contract)

| Dimension | Checks | Threshold |
|-----------|--------|-----------|
| `locale-parity` | §4 (all parity checks clean = 1.0) | 1.0 (must_pass) |
| `build-clean` | §1 (build warning-free + sitemap = 1.0) | 1.0 (must_pass) |
| `style-compliance` | §3 + §5 (proportion of clean checks) | 0.95 |
| `content-fidelity` | §2 + facts/figures preserved vs canonical | 0.9 |

A must_pass dimension below threshold blocks the harness run result
(`must_pass_ok: false`) — fix and re-verify before handing back to the
orchestrator.
