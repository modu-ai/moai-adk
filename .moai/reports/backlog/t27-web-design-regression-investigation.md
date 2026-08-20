# t27 — `moai web` design regression investigation

**Card**: t27 (v3.1 release gate)
**Scope**: read-only investigation. No repo file was edited.
**Baseline**: `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `3b9b3bf99`
**Mockup**: `/Users/goos/Downloads/MoAI-ADK Web Dashboard-handoff.zip` (3,205,179 bytes, 2026-08-15 09:41), extracted to `/tmp/t27mock/`
**Date**: 2026-08-15

---

## Headline

**No regression occurred.** The redesigned operations shell is present on `main` and renders live. The operator's report is not reproducible against `main` HEAD.

The design work landed as a single squash merge, PR **#1523** (`b73c81ab5`), which absorbed *both* commits of the local `feat/web-console-ds` branch — the rebuild and the transitional-layer retirement. Render evidence from a live server confirms the new shell serves on every route the handoff specifies.

The one real defect found is **documentation**: all four docs-site locales still describe the pre-redesign tab console. That is a genuine v3.1 gate item.

---

## Q1 — Which merge reverted it, if any

### Claim

No merge reverted the redesign. `main` HEAD `3b9b3bf99` carries the full handoff implementation. PR #1523 is the merge that *landed* it, not one that reverted it.

### Evidence

**The handoff work lives on an unmerged local branch whose content is nonetheless on `main`.**

```
$ git rev-parse --verify feat/web-console-ds
e1d9437331c5e87a458c5f2de101681a107c7484

$ git ls-remote --heads origin | grep -i "web-console-ds"
NOT on origin

$ git merge-base --is-ancestor 7e4f0eb66 main
NO - not on main

$ git log --oneline main..feat/web-console-ds
e1d943733 fix(kanban): drop duplicate RoleLead const left behind by the main merge
b1b2eb1c8 Merge branch 'main' into feat/web-console-ds
e2cc143e3 fix(web): stop asserting the GLM anchor on a string only a stored key produces
c7df25296 fix(web): regenerate templ output the way CI does
05d4c8a10 refactor(web): retire the transitional layers — settings into the shell, board into /specs
7e4f0eb66 feat(web): rebuild the console around an operations shell (overview/kanban/specs/monitor)
```

The branch commits are not ancestors of `main` — but that is the expected signature of a **squash merge**, which rewrites SHAs. The content test settles it:

```
$ git diff --stat main feat/web-console-ds -- internal/web/
 internal/web/assets/i18n.js            |  4 ----
 internal/web/g3_profile_matrix_test.go | 10 +++++-----
 internal/web/glm_tier_test.go          |  4 ++--
 3 files changed, 7 insertions(+), 11 deletions(-)
```

**`main` and the handoff branch differ by 3 files and 18 lines across the entire `internal/web/` surface.** None of the three is a design artifact:

- `i18n.js` — 4 deletions, all the single dictionary key `f.llm.glm.models.opt.glm-5.3` across the en/ko/ja/zh dictionaries. This is `main` being *newer* (GLM tier unification, #1526), not the branch being ahead.
- `g3_profile_matrix_test.go`, `glm_tier_test.go` — test files, GLM model-tier assertions, same #1526 lineage.

**PR #1523 absorbed both branch commits.** Its diffstat matches the rebuild commit's shape, including the board deletion and the CSS rewrite that only the *first* branch commit (`7e4f0eb66`) introduced:

```
$ git show --stat --format="%H%n%an%n%ci%n%s" b73c81ab5
b73c81ab5674fcb27be63bf1afb0ef68fce1331f
Goos Kim
2026-08-14 12:41:31 +0000
refactor(web): retire the transitional layers — settings into the shell, board into /specs (#1523)

 internal/web/assets/console.css            | 1422 +++--------
 internal/web/assets/i18n.js                |  220 ++
 internal/web/icons.templ                   |   64 +
 internal/web/icons_templ.go                |  218 ++
 internal/web/events.go                     |  183 ++
 internal/web/board.go                      |  188 --
 internal/web/board.templ                   |  139 --
 internal/web/board_templ.go                |  409 ---
 internal/web/fieldsets_templ.go            | 3727 ++++++++++++----------------
 ...
```

The commit *subject* names only the second branch commit, which is what makes this look like a partial landing on a subject-line reading. The *diff* carries both.

**The other candidate PRs were checked and are unrelated to the redesign** — all predate it and touch narrower surfaces:

| PR | Commit | Date | Relation |
|---|---|---|---|
| #1523 | `b73c81ab5` | 2026-08-14 | **Landed the redesign** (squash of both branch commits) |
| #1479 | `4389caffd` | pre-#1523 | MCP console — superseded by, not reverted by, #1523 |
| #1410 | `b0d3b61f8` | pre-#1523 | Earlier console redesign (the 9-tab generation) |
| #1334 | `9cac87a01` | pre-#1523 | agentfm model badge, single-concern fix |
| #1327 | `98c6a8364` | pre-#1523 | report.format localization, single-concern fix |

The known squash-from-old-base hazard was the leading hypothesis and it **does not apply here**: no later merge reintroduced removed content, and no earlier addition was dropped.

### Baseline-attribution

All commands run at `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `3b9b3bf99`, in this session. The `git diff --stat main feat/web-console-ds` output above is the load-bearing measurement.

### Gaps

- I did **not** retrieve PR #1523's GitHub metadata (base branch, head branch, PR-level diff) via `gh`. The squash conclusion is inferred from local diffstat shape plus the identical commit subject, not from the PR record itself. The content-equality measurement is direct and does not depend on this inference.
- I did not bisect commit-by-commit through the 16 commits in `feat/web-console-ds..main`. The 3-file content diff makes a hidden revert in that range arithmetically impossible for `internal/web/`, but I did not examine surfaces outside `internal/web/`.

### Residual-risk

The operator's observation was real to them, so something produced an old-looking console. Two causes remain live and are **not** excluded by this investigation:

1. **A stale binary elsewhere.** The binary I tested (`~/go/bin/moai`, `62,935,154` bytes, 2026-08-15 11:38, reporting `v3.1.0-rc.2 a84e48961`) is current. `internal/web/assets` is `go:embed`-ed (`internal/web/assets.go:24`), so **any** older `moai` binary on PATH serves the old CSS and markup with no other symptom. If the operator ran a differently-installed binary, or an older release, they would see exactly the reported regression. This repo already carries a lesson on this failure mode (clean reinstall via `rm -f` + `cp`, never `cp`-over).
2. **A worktree on an older base.** Several worktrees exist on this checkout; `moai web` run from one predating `b73c81ab5` would serve the old design.

### Recommended action

Do not open a revert-recovery task. Instead ask the operator for two facts: `which moai` and the output of `moai version` from the shell where the old design appeared. If either points away from `a84e48961`, the card closes as a stale-binary artifact.

Separately, **`feat/web-console-ds` should be deleted** — it is unpushed, its content is on `main`, and leaving it invites a future re-merge that would revert `main`'s newer GLM tier changes (the 3 files above).

---

## Q2 — Does the current code match the mockup

### Claim

Yes, at both source and render level, on the dimensions the handoff specifies. Verified with **live render evidence**, not token inspection alone.

### Evidence

**Handoff inventory.** The zip mirrors the repo layout under `project/handoff/internal/web/` and ships 10 files the README orders for application:

| # | File | Nature per handoff README |
|---|---|---|
| 1 | `assets/console.css` | full replacement — tokens + shell + widgets + screens |
| 2 | `icons.templ` | replacement — inline SVG set + brand slot |
| 3 | `shell.templ` | new — replaces `root.templ`'s shell |
| 4 | `widgets.templ` | widget replacement |
| 5 | `screens.templ` | new — overview / kanban / spec / monitor |
| 6 | `viewmodel_ops.go` | new — viewmodel + state inference |
| 7 | `render_helpers.go` | new — templ helpers |
| 8 | `events.go` | new — SSE + fsnotify + debounce |
| 9 | `assets/app-events.js` | appended to `app.js` |
| 10 | `assets/i18n-additions.js` | merged into the 4 dictionaries |

One extraction note: the archive contains a CP949-encoded filename (`MoAI Web Console 재설계안.html`) that `unzip` cannot write on this filesystem. It is the visual comp, not a spec artifact; excluded via `-x "*Console*.html"`. Everything else extracted cleanly.

**Source-level: the shell is the handoff shell.** `internal/web/shell.templ` on `main` opens with an explicit provenance header naming the handoff and its four local corrections:

```
// 재설계본 셸 — 좌측 레일(5영역 + 설정 탭) + 상단바(문맥 칩 · 실시간 · 저장) + 본문.
//
// handoff 원본 대비 로컬 보정:
//   1) profilePop 의 폼 계약을 실제 서버 라우트에 맞췄다...
//   2) 아이콘 호출을 `iconAt(name, size)` 로 바꿨다...
//   3) 브랜드는 `brandMark()` — 원본의 `<use href="#moai-wordmark">` 는...
```

**Render evidence — live server.** Started `~/go/bin/moai web --port 39411 --no-open --no-reuse`:

```
MoAI Web Console starting on http://127.0.0.1:39411
$ curl -s -o /tmp/t27-root.html -w "HTTP=%{http_code} bytes=%{size_download}\n" http://127.0.0.1:39411/
HTTP=200 bytes=17138
```

All six routes the handoff README step 3 prescribes are live:

```
/            200
/kanban      200
/monitor     200
/settings    200
/specs       200
/events      200   (curl exit 28 = SSE stream held open — expected)
```

Screenshots captured via Playwright at 1440×900 (written to `/tmp/`, then removed from the repo tree — `git status` confirms no t27 artifact remains):

- `/tmp/t27-render-overview.png` — full-page Overview
- `/tmp/t27-render-settings.png` — Settings

**The Overview render shows the handoff design as specified**: left rail with the five areas (Overview / Kanban / Specs / Monitor / Settings), achromatic surface with red as the sole chroma, four stat tiles (SPEC 622 · drift 0 · session 0/5 · verify —), a Kanban chain bar with per-role state, an in-progress SPEC list, a "Needs attention" panel, and a Sessions panel with the PID-confirmation caveat. Live indicator present top-right.

**The Settings render confirms the #1523 migration**: settings is now an *area inside the shell*, not a standalone page. The appbar carries the four key-value chips the handoff README prescribes (`lang ko` · `model opus[1m]` · `effort medium` · `dev tdd`), replacing the old single monospace line. The horizontal tab strip became a vertical subnav, per the same README's `tab_layout_test.go` instruction ("가로 nav → 세로 subnav").

Measured structure:

```
rail nav rows (/)             : 5
settings subnav rows          : 10
unique data-tab values        : 9
```

**Token discipline matches the handoff's deletion list.** The README step 1 orders `--color-success` / `--color-warning` / `--color-info` / derived `--status-*` deleted, and `--color-primary` reduced to an alias then removed:

```
--color-success    served=0
--color-warning    served=0
--color-info       served=0
--status-          served=0
--color-primary    served=1   ← inside a comment only
```

The single `--color-primary` hit is documentation, not a live declaration:

```css
/* 주 액션 — 기존 --color-primary(#3d7d5f)를 대체한다 */
--ink:#1d1d1f;--ink-hover:#000000;--ink-fg:#ffffff;
```

Dark theme correctly absent (`grep -c 'data-theme="dark"'` → `0`), matching the README's "다크부재" expectation.

Token coverage: 27 tokens declared in the handoff CSS, 62 in the served CSS (superset). Three handoff token names are absent — `--color-primary`, `--color-surface`, `--fg-on-primary` — all three being the legacy names the handoff itself instructs to retire in favour of the `--ink*` / `--fill*` family. Their absence is compliance, not drift.

**The contrast with the pre-redesign state is unambiguous.** The handoff bundles `uploads/current-01-console.png`, the design session's capture of the *then-current* console: green `#3d7d5f` primary, mascot illustration, centred 880px page, horizontal 9-tab strip, "MoAI Web Console — Edit profile preferences & project settings" hero. The live render shares none of these. The old design is gone from `main`.

The transitional layers memory-noted as outstanding (`console.css` §12 compat layer, `/specs/board` route) are both retired — `grep` for `§12` / `호환` / `compat` / `transitional` in the served CSS returns nothing, and `board.go` / `board.templ` / `board_templ.go` are deleted in `b73c81ab5`.

### Baseline-attribution

Server run from `~/go/bin/moai` (`v3.1.0-rc.2`, commit `a84e48961`, built 2026-08-15T02:38:17Z) against working tree at HEAD `3b9b3bf99`. Served CSS `32,115` bytes; handoff CSS `26,422` bytes. All measurements taken in this session.

### Gaps

- **No pixel-level comparison against the visual comp.** The comp (`MoAI Web Console 재설계안.html`) could not be extracted due to the filename encoding, and the `.jsx` mockups (`mockup-shell.jsx`, `mockup-ops.jsx`, `mockup-settings.jsx`, `doc.jsx`) were **not read**. My Q2 comparison is against the handoff's *implementation* files and its README prescriptions, not against the visual comp. A fidelity difference in spacing, type scale, or component detail would not be caught by what I did.
- **Only two screens rendered.** `/kanban`, `/monitor`, and `/specs` returned HTTP 200 but were not screenshotted or structurally inspected. Their visual fidelity is unverified.
- **The 10-subnav-rows vs 9-`data-tab`-values discrepancy is unexplained.** I did not determine whether one row legitimately carries no `data-tab` or whether a row is inert. This is a small, real, open question.
- **SSE not functionally tested.** `/events` returns 200 and holds the connection, which shows the route exists; I did not verify that a file change actually pushes an event, nor that the 250ms debounce behaves.
- **`_ds/` design-system bundle not inventoried.** The zip carries `_ds_manifest.json`, `_ds_bundle.js`, `styles.css`, and `_adherence.oxlintrc.json` (an adherence lint config). I did not run the adherence check. That config is the closest thing to a mechanical fidelity gate in the archive and would be the natural next verification step.
- **No governing SPEC exists for this work.** `grep -rl "operations shell|ops shell|운영 콘솔 재설계|web-console-ds" .moai/specs/` returns nothing. `SPEC-WEB-CONSOLE-REDESIGN-001` (status: `completed`) documents the *earlier* 9-tab redesign, not this one. The ops-shell rebuild landed without SPEC coverage.

### Residual-risk

"Matches the mockup" is established at the level of structure, routes, token discipline, and the two screens rendered. It is **not** established at the level of visual fidelity, because the comp was never opened. An operator judging by eye could still find real divergence, and that judgment would not contradict anything measured here.

### Recommended action

Before treating Q2 as closed for the release gate: extract the comp with an encoding-aware tool (`unzip -O cp949`, or `ditto -x -k`), open it beside the live render, and screenshot the three unrendered screens. If a fidelity gate is wanted, run the bundled `_adherence.oxlintrc.json` against the served CSS.

---

## Q3 — docs-site coverage

### Claim

**Structural parity is perfect; content is stale in all four locales.** Every locale describes the pre-redesign tab console. The five-area operations shell — the actual current surface — is documented nowhere.

### Evidence

**Structural parity holds exactly.**

```
files per locale:      en 143 · ko 143 · ja 143 · zh 143
pages mentioning "moai web":  en 7 · ko 7 · ja 7 · zh 7
```

Two dedicated pages exist per locale, 8 files total:

```
docs-site/content/{en,ko,ja,zh}/advanced/moai-web-console.md
docs-site/content/{en,ko,ja,zh}/cli-reference/web.md
```

**The content describes the old console.** English headings:

```
# MoAI Web Console
## What It's For
### Profile preference editing
### Project section editing
## Console Structure — 9 Tabs
### Widget honesty
### GLM honesty badge
## 4-Locale UI
## Getting Started
## Security Model
```

The same structure is mirrored in every locale — this is a synchronized translation of a stale source, not a drift between locales:

| Locale | Structure heading |
|---|---|
| en | `## Console Structure — 9 Tabs` |
| ko | `## 콘솔 구조 — 9개 탭` |
| ja | `## コンソール構造 — 9つのタブ` |
| zh | `## 控制台结构 — 九个选项卡` |

**The new surface has zero documentation coverage.** Grepping both dedicated pages across all four locales:

```
Overview            1     (incidental prose, not the rail area)
Kanban              0
Monitor             0
/specs              0
/monitor            0
/kanban             0
SSE                 0
operations shell    0
실시간               0
```

Nine of ten probes return zero across 8 files. The Kanban board, the Monitor screen, the Specs area, the live SSE update mechanism, and the fact that settings is now one area among five are all undocumented.

**The staleness has a clean chronological explanation.** Both pages were last touched by:

```
2c70346ab 2026-08-13 07:08:57 +0000 docs(SPEC-WEB-CONSOLE-REDESIGN-001): sync-phase close (3-phase plan→run→sync) (#1503)
```

That is **2026-08-13**. The redesign landed **2026-08-14** (`b73c81ab5`). The docs were closed out one day before the surface they describe was replaced, and no docs commit followed. Consistent with Q2's finding that no SPEC governs the ops-shell rebuild — with no SPEC, no sync phase ran, so no docs update was triggered.

### Baseline-attribution

`grep`/`find` over `docs-site/content/` at HEAD `3b9b3bf99`; `git log -1` per file, this session.

### Gaps

- I inspected **headings and keyword presence**, not full body prose. A passage correctly describing a new screen without using any of my ten probe terms would have been missed — unlikely across 8 files and 9 zero-scoring probes, but not excluded.
- The other 5 pages per locale that mention `moai web` were **not** read. Their mentions may be incidental cross-references or may contain their own stale descriptions; I did not check.
- I did not build the site (`hugo`) or check for broken internal links arising from the route changes (`/specs/board` was retired — any doc linking it would now 404). **This is worth checking before release.**
- `README.md` and its three locale variants were not examined for web-console descriptions.

### Residual-risk

Because the four locales are stale *in lockstep*, the usual 4-locale parity check passes and will keep passing. Nothing mechanical will flag this. It surfaces only when a reader opens the docs beside the running console.

### Recommended action

This is the one item in this investigation that genuinely gates v3.1. Recommended sequence:

1. Author a SPEC covering the ops-shell rebuild retroactively, so the docs update has a sync phase to ride and the work stops being SPEC-invisible.
2. Rewrite `advanced/moai-web-console.md` in the canonical locale around the five rail areas, then derive ko/ja/zh in the same PR per the 4-locale obligation.
3. Update `cli-reference/web.md` with the six routes.
4. Grep the docs tree for `/specs/board` and any other retired route before building.

---

## Summary

| Question | Verdict | Confidence |
|---|---|---|
| Q1 — which merge reverted it | **None did.** #1523 (`b73c81ab5`) *landed* the redesign as a squash of both `feat/web-console-ds` commits. `main` and the branch differ by 3 files / 18 lines, none design-related. | High — direct content measurement |
| Q2 — does code match mockup | **Yes**, at structure / route / token level, with live render evidence. Visual comp never opened; three screens unrendered. | Medium-high on what was measured; comp fidelity unverified |
| Q3 — docs-site coverage | **Stale in all four locales.** Perfect structural parity, zero coverage of the new surface. Docs closed 2026-08-13, redesign landed 2026-08-14. | High |

**Most probable explanation for the operator's report**: a stale `moai` binary or an older worktree, not a code regression. `internal/web/assets` is `go:embed`-ed, so an old binary serves the old design silently.

**Action items, ranked**

1. Confirm the operator's `which moai` / `moai version` in the shell that showed the old design — resolves the card.
2. Fix docs-site: 4-locale rewrite around the five-area shell. This is the real gate item.
3. Delete the unpushed `feat/web-console-ds` branch — re-merging it would revert `main`'s newer GLM tier changes.
4. Author a retroactive SPEC for the ops-shell rebuild.
5. Optional fidelity pass: extract the comp with `unzip -O cp949`, render `/kanban` `/monitor` `/specs`, run the bundled adherence lint.

---

## Artifacts

| Path | Contents |
|---|---|
| `/tmp/t27-render-overview.png` | Full-page Overview render, 1440×900 |
| `/tmp/t27-render-settings.png` | Settings render, 1440×900 |
| `/tmp/t27-root.html` | Served `/` HTML, 17,138 bytes |
| `/tmp/t27-settings.html` | Served `/settings` HTML, 97,474 bytes |
| `/tmp/t27-served.css` | Served `console.css`, 32,115 bytes |
| `/tmp/t27-server.log` | Server startup log |
| `/tmp/t27mock/` | Extracted handoff archive |

Artifacts live under `/tmp` deliberately; note that `/tmp` is cleared by the OS, so re-run the capture rather than citing these paths at a later audit.
