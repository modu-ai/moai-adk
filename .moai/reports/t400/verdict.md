# t400 — shared feature-flag slot documented as a fifth availability constraint

Branch `WT-messaging-slot-axis`, worktree `.claude/worktrees/t400`, base `e7746e95d` (= origin/develop at start), commit `25b341dcb`.

## Claim

The card's item (1) is delivered: `cross-session-messaging.md` now carries the shared-slot axis its availability list omitted, the companion carries the mechanism and the measurement, and both template mirrors follow. The card's item (2) — injecting `CLAUDE_CODE_HARBOR_KITE=1` at `moai glm` launch — is deliberately not implemented; the card marks it review-only, and the operator confirmed that scope before run-phase entry.

## Evidence

| Check | Command | Result |
|---|---|---|
| No code surface in this tree | `git grep -nEi "tengu_harbor_kite\|harbor_kite\|HARBOR_KITE" origin/develop -- .` | 0 hits |
| Template package tests | `go test ./internal/template/...` | `ok github.com/modu-ai/moai-adk/internal/template 24.999s`, `ok .../agentemit 0.267s`, exit 0 |
| Embed rebuild | `make build` | exit 0; `catalog.yaml` byte-unchanged (absent from `git status --short`) |
| Mirror parity, rule file | `diff .claude/... internal/template/templates/.claude/...` | only `142,143d141` — the deliberate `> Origin: SPEC-...` omission |
| Mirror parity, companion | same, on `-detail.md` | exit 0, byte-identical |
| Template neutrality | grep for SPEC/REQ/AC ids, ISO dates, issue numbers, absolute paths over both mirrored files | 0 hits |
| Always-loaded growth | `wc -c` before vs after on `cross-session-messaging.md` | 15,852 → 16,987 (+1,135) |

## Baseline-attribution

Every figure above was measured in this run, in this worktree, against `e7746e95d` before the commit and against the working tree after it. The "0 hits" grep was run against the `origin/develop` ref rather than the primary checkout's working tree, because the primary sits on `main`, 719 commits behind (`git rev-list --count --left-right origin/develop...HEAD` → `719 0`).

The four-run causal experiment (slot `true` → socket ×2, `false` → none ×2) and the 36-second write-timestamp observation are the **external reporter's** measurements, cited as reported. This session did not reproduce them, and the documentation text says "the measurement this rests on" rather than claiming them as ours.

## Gaps

- The reporter's causal experiment was not independently reproduced here. Reproducing it means flipping a machine-global slot that every session on this machine reads, so it was not attempted while other work could be running.
- `CLAUDE_CODE_HARBOR_KITE=1` is documented as an escape hatch on the strength of the reporter's socket observation and the quoted gate ordering. This session did not itself start a session with the variable set and confirm a socket appeared.
- The gate function text quoted in the report was not re-extracted from a local binary.
- CI has not run: the branch is unpushed at the time of writing.

## Residual-risk

- The slot name, the gate's env-variable check, and the evaluation endpoint are all upstream internals with no compatibility promise. A rename upstream silently rots the bullet, the diagnostic command, and the escape hatch together. Nothing in this tree detects that — the documentation asserts a fact about another program, and no check here can go red when it stops being true.
- The always-loaded surface grew 1,135 bytes for a constraint most sessions never hit. The judgment that the silence of the failure justifies the charge is a judgment, not a measurement.
