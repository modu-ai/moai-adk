# t54 — release.yml check 5 alignment with the new release-notes contract

- Branch: `WT-t54` (renamed from `worktree-agent-a55dd6a3e0ddbb399`)
- Base: `origin/release/v3.1.1` (FETCH_HEAD, ff to `09516ff0c`) + sibling lane `WT-t51` (ff to `1fbea5ecf`)
- Files analyzed: `.github/workflows/release.yml` (check 5, L102-108, post-t51 state). **No file modified.**
- Date: 2026-08-17

## Step 1 — precondition: PR #1563 merge state

```
$ gh pr view 1563 --json state,mergedAt,title,mergeCommit
{"mergeCommit":{"oid":"2b3eae590974a35b482b87b358079da81fed1d1b"},
 "mergedAt":"2026-08-16T07:44:11Z","state":"MERGED",
 "title":"fix(ci): stop the release provenance gate failing on checks that matched"}
```

MERGED. Unblock condition satisfied. `git log` confirms the ancestry in this
worktree: release.yml was last touched by `2b3eae590` (#1563), then `1fbea5ecf`
(WT-t51), both ancestors of HEAD.

## Step 2 — what check 5 asserts (post-t51 combined state, verbatim)

```yaml
          # Check 5 — CHANGELOG.md at the tagged commit has this version's section.
          # Formal sections are bare ('## [3.1.0]'); pre-release (rc) sections
          # carry the v prefix ('## [v3.0.0-rc12]'). 'v?' accepts both forms.
          VERSION_NO_V="${TAG#v}"
          if ! git show "${TAG_COMMIT}:CHANGELOG.md" | grep -E "^## \[v?${VERSION_NO_V}\]" >/dev/null; then
            fail "check 5 (CHANGELOG): CHANGELOG.md at ${TAG_COMMIT} has no '## [${VERSION_NO_V}]' or '## [v${VERSION_NO_V}]' section."
          fi
```

Purpose (per the workflow header comment, L38-41): check 5 is one of the three
"substantive protections" — an ad-hoc or out-of-process tag fails it because the
sanctioned flow (`scripts/release.sh` validation 8) refuses to tag unless the
version's CHANGELOG section already exists. It is a **provenance proxy**, not a
release-notes completeness check.

## Step 3 — the new contract (from the two split SPECs, read in the `card-relnotes` authoring worktree)

Contract sources (both `status: draft`, v0.2.0 / v0.1.0 — **not yet landed**):
`.moai/specs/SPEC-RELEASE-NOTES-ASSEMBLY-001/spec.md` and
`.moai/specs/SPEC-CHANGELOG-BUDGET-001/spec.md` in worktree `card-relnotes`.

| Surface | Old (de-facto today) | New (per the SPECs) |
|---|---|---|
| Release-note content | CHANGELOG.md per-version section (full narrative, up to 260,955 B for `## [3.1.0]`) | `.moai/release-notes/vX.Y.Z.en.md` + `.ko.md` are the SSOT (ASSEMBLY REQ-1); assembled by `moai release notes <version>` (REQ-2) |
| CHANGELOG.md | full narrative | **summary index**: per-version sections survive, each carrying a link to the version's `.en.md` SSOT + an unreleased marker while no release exists (BUDGET REQ-4); per-entry 400-char / per-section 25,000-char caps (REQ-1/2) |
| Version heading format | formal bare `## [3.1.0]`, rc v-prefixed `## [v3.0.0-rc12]` | unchanged — neither SPEC touches heading form; BUDGET REQ-4 keys on "버전 섹션이 생성될 때", i.e. sections are still created in CHANGELOG.md |
| `scripts/release.sh` tag-annotation path | extracts CHANGELOG section + appends `.ko.md` | **preserved verbatim** (ASSEMBLY §E: "이 SPEC은 그 경로를 현행 동작 그대로 보존하며"); rc-prefix defect in that path was t51's separate card |
| release.yml | — | explicitly **out of scope for both SPECs** ("이 SPEC은 이 파일을 편집하지 않는다"); check-5 alignment was split out to this backlog card (t54) |

Key load-bearing fact for check 5: **the summary-index conversion keeps the
per-version `## [v?X.Y.Z]` heading as the section key in CHANGELOG.md.** What
changes is the section's *content* (narrative → link + summary), which check 5
never inspects.

## Sub-check classification

| # | Sub-check | Classification | Reason |
|---|---|---|---|
| 5a | `VERSION_NO_V="${TAG#v}"` | WORKS-AS-IS | Tag naming (`vX.Y.Z` / `vX.Y.Z-rc.N`) is untouched by the release-notes contract; check 3 binds annotation version to TAG and check 6 binds system.yaml to the v-prefixed TAG, so the prefix invariant is independently enforced. |
| 5b | `git show "${TAG_COMMIT}:CHANGELOG.md"` (file exists at tagged commit) | WORKS-AS-IS | CHANGELOG.md is not removed by the transition — BUDGET REQ-4 continues to create per-version sections in it (as index entries). |
| 5c | `grep -E "^## \[v?${VERSION_NO_V}\]"` (heading match) | WORKS-AS-IS | Heading form is unchanged by both SPECs; formal stays bare, rc stays v-prefixed, and t51's `v?` already accepts both. Pattern is prefix-anchored with no trailing `$`, so heading suffixes (dates today, BUDGET REQ-4's unreleased markers tomorrow) cannot break the match — verified by probe E below. Semantic note: what the check *means* shifts from "the narrative exists at the tagged commit" to "the version's index entry exists at the tagged commit" — the provenance-proxy function survives because release.sh validation 8 still requires the section before tagging (path explicitly preserved by ASSEMBLY §E). |
| 5d | pipe discipline (no `grep -q`, `>/dev/null` redirect) | WORKS-AS-IS | This is PR #1563's SIGPIPE fix, which ASSEMBLY §E pins ("PR #1563 보존"). Re-verified empirically under `set -o pipefail` against the largest historical CHANGELOG (416,407 B at the v3.1.0 commit) — probe A. The new contract shrinks CHANGELOG (25,000-char section cap), moving further away from the pipe-buffer class, not closer. |
| 5e | `fail` message text (names both heading forms) | WORKS-AS-IS | Accurate post-t51; cosmetic surface. |
| — | full workflow job (tag push → verify-provenance → release) | UNTESTABLE-LOCALLY | Requires a real `v*` tag push to trigger; card constraints forbid release runs and remote writes. Check 5 in isolation is fully replicated below; the surrounding job is not. |

## Evidence — verbatim probes (run in this worktree, 2026-08-17)

Heading inventory of the current worktree CHANGELOG.md:

```
$ grep -n '^## \[' CHANGELOG.md
8:## [Unreleased]
19:## [3.1.0] - 2026-08-15
286:## [3.0.2] - 2026-07-29
349:## [3.0.1] - 2026-07-21
393:## [3.0.0] - 2026-07-20
451:## [3.0.0] - 2026-07-20 (한국어)
495:## [v3.0.0-rc12] - 2026-07-14
528:## [v3.0.0-rc11] - 2026-07-13
```

Tagged commits resolved for replication: `v3.1.0` → `ed04e40e6fd8a078dd2de242b55839e268725ca6`,
`v3.0.0-rc12` → `93161bf785a83711e79617e1e4ee67bf2f37ebed`.

Probe A — formal release, exact post-t51 expression under pipefail against the
416,407 B historical CHANGELOG (both the t51 heading fix and the #1563 SIGPIPE
fix exercised):

```
$ bash -c 'set -o pipefail; git show "$(git rev-list -n1 v3.1.0)":CHANGELOG.md | grep -E "^## \[v?3.1.0\]" >/dev/null; echo "formal v3.1.0 post-t51 pattern rc=$?"'
formal v3.1.0 post-t51 pattern rc=0
```

Probe B — rc release with the post-t51 `v?` pattern:

```
$ bash -c 'set -o pipefail; git show "$(git rev-list -n1 v3.0.0-rc12)":CHANGELOG.md | grep -E "^## \[v?3.0.0-rc12\]" >/dev/null; echo "rc12 post-t51 pattern rc=$?"'
rc12 post-t51 pattern rc=0
```

Probe C — same rc commit with the PRE-t51 expression (negative control,
confirming t51's fix is load-bearing and reproducing the SPEC's
`grep -c '^## \[3.0.0-rc12\]'` → 0 finding):

```
$ bash -c 'set -o pipefail; git show "$(git rev-list -n1 v3.0.0-rc12)":CHANGELOG.md | grep -E "^## \[3.0.0-rc12\]" >/dev/null; echo "rc12 PRE-t51 pattern rc=$?"'
rc12 PRE-t51 pattern rc=1
```

Probe D — negative control, nonexistent version (out-of-process-tag detection):

```
$ bash -c 'set -o pipefail; git show "$(git rev-list -n1 v3.1.0)":CHANGELOG.md | grep -E "^## \[v?9.9.9\]" >/dev/null; echo "nonexistent-version negative control rc=$?"'
nonexistent-version negative control rc=1
```

Probe E — anchoring robustness: neither a suffix-bearing future-form heading
(BUDGET REQ-4 unreleased marker) nor prefix-similar versions cross-match; a
heading suffix does not break the genuine match:

```
$ printf '## [3.1.1-rc1] - 2026-01-01\n## [3.1.10] - 2026-01-01\n## [3.1.1] - 2026-01-01\n## [3.1.1] — unreleased marker suffix\n' | grep -E "^## \[v?3.1.1\]"; echo "cross-match control rc=$?"
## [3.1.1] - 2026-01-01
## [3.1.1] — unreleased marker suffix
cross-match control rc=0
```

Probe F — SSOT directory state at the tagged commits (what exists in git at the
commits check 5 reads):

```
$ git ls-tree ed04e40e6fd8a078dd2de242b55839e268725ca6 .moai/release-notes/
100644 blob 926d49b3e1cb46bab983e0812e6df94c4436b654	.moai/release-notes/v3.1.0.ko.md

$ git ls-tree 93161bf785a83711e79617e1e4ee67bf2f37ebed .moai/release-notes/
(no output — directory did not exist at rc12)
```

Sizes for context: CHANGELOG.md at the v3.1.0 commit = 416,407 B
(`git cat-file -s ed04e40e6...:CHANGELOG.md`); current worktree = 421,933 B.

## Verdict — NO-CHANGE-NEEDED

No edit was made to `release.yml` (or anything else). "변경이 필요 없다는
결론도 결론이다" (card, quoting the pre-split plan). Rationale:

1. Check 5 reads only CHANGELOG.md, and the new contract keeps per-version
   `## [v?X.Y.Z]` headings in CHANGELOG.md as the index key (BUDGET REQ-4).
   Nothing check 5 greps for disappears under the new contract.
2. The rc dual-form acceptance is already t51's work, verified here against the
   real rc12 tagged commit (probes B/C).
3. The pipe discipline is already #1563's work, pinned by the ASSEMBLY SPEC and
   re-verified under pipefail against the largest historical CHANGELOG (probe A).
4. The one candidate enhancement — also asserting `.moai/release-notes/vX.Y.Z.en.md`
   exists at the tagged commit — would be wrong to add now: the `.en.md` half of
   the SSOT does not exist yet for ANY version (probe F: only `v3.1.0.ko.md`
   exists, and nothing at all existed at rc12), both contract SPECs are still
   `draft`, and both explicitly declare release.yml out of scope. Adding the
   check today would break the next real release against a contract that has not
   landed. If an SSOT-presence gate is wanted, it belongs to the ASSEMBLY
   SPEC's own landing (it can amend REQ scope then) — noted here as a hand-off,
   not acted on.

## Baseline-attribution

All probes ran in this worktree (`WT-t54`, HEAD `1fbea5ecf`, =
`origin/release/v3.1.1` @ `09516ff0c` + WT-t51) on 2026-08-17, against the
tagged commits resolved above and the worktree's own CHANGELOG.md. PR state
read live via `gh pr view 1563`.

## Gaps (not observed)

- The workflow job itself was never executed — no tag was pushed, no
  Actions run observed. Check 5 was replicated command-by-command, not as a
  GitHub Actions job.
- The future-contract states (`.en.md` files, the `moai release notes`
  assembler, the summary-index CHANGELOG) do not exist in-tree; the
  classification of 5b/5c against them rests on the REQ text of two draft
  SPECs, not on a measured post-landing tree.
- BUDGET REQ-4's unreleased-marker placement (heading vs body) is not yet
  pinned by the draft; probe E covers both placements by construction, but the
  landed form was not inspected because it does not exist.
- v3.1.1's own CHANGELOG section does not exist yet (only `## [Unreleased]`),
  so the next formal release could not be end-to-end probed either.

## Residual-risk

- `VERSION_NO_V` is interpolated into `grep -E` from the pushed tag name
  (external input). A tag like `v.*` passes the `on.push.tags: v*` glob and
  would interpolate regex metacharacters. This is pre-existing, unchanged by
  the new contract, and bounded by checks 2/3/4/7 (trailer + version equality
  + commit binding + main ancestry) — the workflow header already documents
  the honest limit that the gate is not an authentication mechanism. No action
  taken (out of card scope; noting for the record).
- If a future SPEC ever removes CHANGELOG.md or re-keys headings (e.g. table
  rows instead of `## [...]`), check 5 breaks silently-by-design (fails every
  release). Neither current draft does this; the risk is conditional on a
  contract change not currently proposed.
