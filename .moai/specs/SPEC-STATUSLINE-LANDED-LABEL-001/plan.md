---
id: SPEC-STATUSLINE-LANDED-LABEL-001
title: "Implementation plan — statusline landed annotation"
---

# Plan — SPEC-STATUSLINE-LANDED-LABEL-001 (card t1281)

## §A Context

Base measured on this tree: `b59a5d69c` (branch `WT-statusline-landed-label`). Annotation introduced by `6ce0b8cd6` (card t456). Decision ③ is fixed by the lead; this plan does not reopen it.

## §B Known Issues (measured)

- `internal/statusline/landed.go`: `RefreshLandedCounts` runs `git log <ref> --format=%B`; `countNamed` matches `\b(?:ids)\b` across whole messages. `pickedCardIDs` returns ids only (no `added_at`).
- `internal/statusline/renderer.go` backlog segment: `seg += fmt.Sprintf(" ✓%d", …)`; the comment above it explains "never subtracts" in terms of the old body-mention criterion.
- `LandedCounts` JSON schema: `{landed, ref, measured, fetched_at}` — no criterion marker, so an old body-criterion cache is indistinguishable from a new one.
- `kanban.ScanLandedSubjects` takes a `kanban.CommandRunner` (`func(name string, args ...string) (string, error)`) — no context and no working directory. The statusline runner `landedGitRunner(ctx, dir, args...)` carries both.
- Lane measurement (this session): cache `landed=9` vs `moai todo auto-done --dry-run` `scanned=29 closed=4` on `origin/develop`; the 4 were plan-only landings.

## §C Pre-flight

- Read `internal/kanban/autodone_scan.go` (done at plan time) and `internal/cli/todo_autodone.go:205-320` (the reference composition: `ScanLandedSubjects` → `LandedAttributions(commits, LandedBranchFromRef(ref))` → `AutoDoneSubjectFresh(hit, it.AddedAt)`).
- Confirm `github.com/mattn/go-runewidth` is already a direct dependency (`go.mod`: v0.0.29) — usable in the width test without a new dependency.

## §D Constraints

- Lint gate: **golangci-lint v2.1.6** (the CI version; a newer local binary is not evidence).
- Run-phase remeasure: `go test ./internal/statusline/... ./internal/kanban/...` (no full-suite local run).
- One git invocation per refresh; zero on the render path.

## §E Self-Verification

Run phase reports each AC (acceptance.md) with the command and verbatim output; `progress.md` §E.2/§E.3 are owned by manager-develop.

## §F Milestones (ordered by decision-reversibility)

### M1 — Cache schema: criterion marker (Priority High, highest change likelihood)

- Add a criterion field to `LandedCounts` (e.g. JSON `criterion`, value a package constant such as `"subject-attribution/v1"`).
- `Known()` additionally requires the current criterion; `maybeRefreshLandedCounts` treats a criterion mismatch as stale regardless of `FetchedAt` (REQ-SLL-007).
- Refresh placeholder: when `prev` carries a mismatched criterion, the timestamp-first write resets it to an unmeasured placeholder under the current criterion (`Measured=false`, `Landed=0`) so a subsequent failure cannot resurrect the old number (REQ-SLL-008). A matching-criterion `prev` rides through unchanged as today.

### M2 — Glyph and renderer comment (Priority High)

- Replace ` ✓%d` with ` ⚑%d` via a named constant (single rune U+2691).
- Rewrite the renderer comment: annotation means "a subject-attributed landing commit exists — verify before done"; it never subtracts because a card stays picked until auto-done / `moai todo done` actually closes it, and a landing (possibly plan-only) is not a close.
- Update the `LandedCounts.Landed` field comment and the `landed.go` file header to the subject criterion.

### M3 — Criterion: reuse kanban's subject scan (Priority High)

- `pickedCardIDs` → return picked cards with `ID` and `AddedAt` (still via `LoadPure`, still filtered by `landedCardToken`).
- Replace the `%B` query with `kanban.ScanLandedSubjects(adapter, ref)`, where `adapter` is a closure `func(name string, args ...string) (string, error)` that calls `landedGitRunner(ctx, boardRoot, args...)` (ignores `name`, which is always `git`). This keeps exactly one invocation and the existing test seam; argv equals `kanban.LandedScanArgs(ref)`.
- Count = number of picked cards `c` with `hit, ok := attributions[c.ID]; ok && kanban.AutoDoneSubjectFresh(hit, c.AddedAt)`, where `attributions = kanban.LandedAttributions(commits, kanban.LandedBranchFromRef(ref))`.
- A `ScanLandedSubjects` error (runner failure or malformed line) takes the existing failed-query path: keep the stale cache, never write zero.
- Remove `countNamed` and its test (`TestCountNamed_WordBoundaryCriterion`); its criterion is retired. No other caller exists in the package (verify with grep at run time).

### M4 — Tests (Priority Medium)

- Update `landed_test.go`: render cases expect `⚑`, assert no `✓`; ref test asserts argv equals `kanban.LandedScanArgs(ref)`; fake git outputs use the `%H\x00%ct\x00%s` shape.
- New cases per acceptance.md AC-SLL-001..011.

### M5 — Documentation (Priority Medium, mechanical)

Measured by grep for `✓` near TODO/statusline across `internal/template/templates` and `docs-site`:

- `internal/template/templates/**`: **no occurrence** — no template change needed.
- `docs-site/content/{ko,en,ja,zh}/advanced/statusline.md` — lines 28, 33, 200 in each locale (12 lines): replace `✓N` / `✓2` with `⚑N` / `⚑2`, and reword the meaning from "named in the integration branch history" to "a subject-attributed landing commit exists — verify before marking done". 4-locale same-change obligation applies (docs-site i18n rules).
- `CHANGELOG.md` carries historical `✓N` mentions — not rewritten (sync phase adds a new entry).

## §G Anti-Patterns

- Writing a new regex/matcher in `internal/statusline` for commit subjects.
- One git call per card (the shape `landed.go` exists to avoid).
- Rendering an old-criterion number under the new glyph "until the next refresh".
- Writing `Landed: 0, Measured: true` on any failure path.
- Using an emoji-presentation or East-Asian-ambiguous glyph.

## §H Risks

- Adapter drops `name` — if kanban ever passes a non-git command the adapter would still run git. Mitigation: the adapter returns an error when `name != "git"`.
- `subjectAttribution` recognises only `t[0-9]+` card tokens; picked ids of other shapes will never count. This matches auto-done behavior and is accepted.
- Cache migration causes one refresh cycle of "no annotation" on upgrade — intended (unknown is shown as nothing).

## §I Cross-References

- `internal/kanban/autodone_scan.go`, `internal/kanban/prlink_landed.go` (`subjectAttribution`, `LandedBranchFromRef`)
- `internal/cli/todo_autodone.go` (reference composition)
- `internal/statusline/github.go` (the cache/refresh pattern `landed.go` mirrors)
