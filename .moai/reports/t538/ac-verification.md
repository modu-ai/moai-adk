# t538 — AC verification raw output

Tree: `.claude/worktrees/t538` · branch `WT-docs-v313-locales` · base `bce6d7e08`
Baseline SHA: `d787897ca` (docs untouched since base) · after SHA: `91fd27bb6`
Tool: `/usr/bin/grep` — the shell `grep` here is a ugrep wrapper that skips silently.
Baselines below are re-derived from the baseline SHA via `git show`, so every figure in this file is re-derivable from the file alone.

```
### AC-001  en 'not yet provided'   expect 1 -> 0
baseline: 1
after:    0

### AC-002  zh 尚未提供   expect 1 -> 0
baseline: 1
after:    0

### AC-003/004  desktop-native en,zh   expect 0 -> >=4 each
baseline en: 0
baseline zh: 0
after    en: 7
after    zh: 7

### AC-005  table rows '^|' and headings '^#'   expect en,zh 42 -> 46 ; headings 18 throughout
baseline rows ko: 46
baseline rows en: 42
baseline rows ja: 46
baseline rows zh: 42
after    rows ko: 46
after    rows en: 46
after    rows ja: 46
after    rows zh: 46
after headings ko: 18
after headings en: 18
after headings ja: 18
after headings zh: 18

### AC-005  host-OS token, native per locale   expect 1 each
ko 호스트 OS 규칙: 1
ja ホスト OS ルール: 1
en host OS rule:   1
zh 主机 OS 规则:    1

### AC-006  doctor example rows   anchored expect ja,zh 0 -> 2 ; whole-file 2 -> 4
baseline anchored ja: 0
baseline anchored zh: 0
after    anchored ja: 2
after    anchored zh: 2
baseline whole    ja: 2
baseline whole    zh: 2
after    whole    ja: 4
after    whole    zh: 4

### AC-007  bold-parenthetical scan of doctor.md   expect ko,ja,zh 1 -> 0 ; en 0 held
baseline ko: 1
baseline en: 0
baseline ja: 1
baseline zh: 1
after    ko: 0
after    en: 0
after    ja: 0
after    zh: 0

### AC-008  skill-guide SVG rule families   expect 0 -> >=2 each (amended: occurrence count)
baseline ko:        0
baseline en:        0
baseline ja:        0
baseline zh:        0
after    ko:        2
after    en:        2
after    ja:        2
after    zh:        2
(superseded grep -c form, retained for the AC-008 both-readings claim)
after -c ko: 2
after -c en: 2
after -c ja: 2
after -c zh: 2

### AC-009  boundary — ko/ja e2e unchanged, no badge added
ko/ja e2e diff exit: 0  (0 = unchanged)
added badge lines: 0

### AC-011  landing set, amended (scope-limited) expression
scope-limited count:       10
CHANGELOG.md
docs-site/content/en/advanced/skill-guide.md
docs-site/content/en/utility-commands/moai-e2e.md
docs-site/content/ja/advanced/skill-guide.md
docs-site/content/ja/cli-reference/doctor.md
docs-site/content/ko/advanced/skill-guide.md
docs-site/content/ko/cli-reference/doctor.md
docs-site/content/zh/advanced/skill-guide.md
docs-site/content/zh/cli-reference/doctor.md
docs-site/content/zh/utility-commands/moai-e2e.md
CHANGELOG t538 entries: 1
whole-branch count (REPORTING figure, not a bar):       18
```
