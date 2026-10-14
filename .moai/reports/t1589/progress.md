# Progress — card t1589

- Card: t1589 — internal/cli/factory_card_pr.go GIT_DIR inheritance on the completion delivery path
- Class: B (defect, cause unknown) · Stages: run → sync
- Lane: lane-27 · Run: tmvgvn

## Lease

- `moai factory next` ×3: `no card is available`.
- `moai factory status --run tmvgvn`: `t1589 … owner=lane-27 lease=none`.
- Proceeded without a lease: 임대 없음(고부하, 리더 지시) — leader instruction, high load.
- Worktree: `moai worktree new t1589` from local main `ec193e4777`; branch renamed to `WT-factory-card-pr-git-dir`.

## Commits (branch WT-factory-card-pr-git-dir)

- `6b12d5b830` test(factory): RED — delivery git reads must ignore an inherited GIT_DIR (card t1589)
- `4faf5ff818` fix(factory): scrub the git environment on the delivery edge's reads (card t1589)

## Stage record

- run: RED observed (3 assertions), fix applied, 40-test targeted set GREEN (`exit=0`), gofmt clean.
- sync: verdict.md and card-review.md written under `.moai/reports/t1589/`.
- Remote: none (no push; merged-local + SHA is the endpoint).

## Open for leader

- Card body's strategy-branch defect (git-flow config vs github-flow operation) — not repaired here.
- Two raw git call sites outside the delivery edge (factory_card.go:633, :2272) — not changed.
