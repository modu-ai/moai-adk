# t51 — rc release path: CHANGELOG version-key dual-form fix

- Branch: `WT-t51` (renamed from `worktree-agent-a3b7729a864e4caef`)
- Base: merged `origin/release/v3.1.1` (FETCH_HEAD) → pre-fix HEAD `09516ff0c`
- Files: `scripts/release.sh` (validation 8, L158-176), `.github/workflows/release.yml` (check 5, L102-108)
- Date: 2026-08-17

## (a) Option chosen: (b) — key computation accepts BOTH v-prefixed and bare forms

Measured convention in `CHANGELOG.md` (this worktree, full heading inventory):

```
## [Unreleased]
## [3.1.0] - 2026-08-15
## [3.0.2] - 2026-07-29
## [3.0.1] - 2026-07-21
## [3.0.0] - 2026-07-20
## [3.0.0] - 2026-07-20 (한국어)
## [v3.0.0-rc12] - 2026-07-14
## [v3.0.0-rc11] - 2026-07-13
```

Formal headings are bare; rc headings carry the `v`. Both gates stripped `v`
unconditionally, so every rc tag died at validation 8 / check 5.

Why (b) beats the alternatives:

- **Not (a)** (unify headings without v + migrate `[v3.0.0-rc12]`): rewrites the
  documented history of two already-shipped releases (rc11/rc12 sections),
  wider blast radius for zero runtime gain — any external link or anchor to
  those headings breaks, and the migration touches CHANGELOG content outside
  the card's named file scope (the card names release.sh L159-162, check 5,
  and the heading convention as the *assumption*, not as rewrite targets).
- **Not (c)** (exempt rc from CHANGELOG): changes the documented contract
  (release.sh header "전제 조건" requires the CHANGELOG section; release.yml's
  own honest-limitations comment names checks 5-7 as "the substantive
  protection"). Exempting rc would make rc releases strictly less gated than
  formal ones — the opposite of what a provenance gate should do.
- **(b)** accepts the measured reality with the smallest possible diff: the
  formal path is byte-identical when the bare heading exists; the v-prefixed
  form is a fallback. No CHANGELOG content is touched; no contract changes;
  the gate still dies when NEITHER form exists (negative control below).

## (b) RED — verbatim (before fix)

```
=== RED-1: release.sh L159-164 key computation (rc12) ===
$ VERSION=v3.0.0-rc12
$ CHANGELOG_VERSION="${VERSION#v}"
CHANGELOG_VERSION=3.0.0-rc12
$ grep -c "^## \[$CHANGELOG_VERSION\]" CHANGELOG.md
0
exit=1
$ grep -c "^## \[v3.0.0-rc12\]" CHANGELOG.md   # actual heading form
1
exit=0

=== RED-1b: validation-8 gate condition dies (grep -q) ===
$ grep -q "^## \[$CHANGELOG_VERSION\]" CHANGELOG.md && echo MATCH || echo "DIE: no match -> validation 8 fails"
DIE: no match -> validation 8 fails

=== RED-1c: tag-annotation extraction (awk L218-222) with bare header yields nothing ===
       0
exit=0 (0 lines -> L224 die: Failed to extract CHANGELOG section)

=== RED-2: release.yml check 5 (L103-105) logic, rc12 ===
$ TAG=v3.0.0-rc12
$ VERSION_NO_V="${TAG#v}"
VERSION_NO_V=3.0.0-rc12
$ grep -E "^## \[${VERSION_NO_V}\]" CHANGELOG.md   # (git show TAG_COMMIT:CHANGELOG.md replaced by worktree file)
exit=1
-> no match => fail "check 5 (CHANGELOG)" fires
```

(`git show ${TAG_COMMIT}:CHANGELOG.md` was replaced by the worktree
`CHANGELOG.md` in the check-5 probes — the grep expression under test is the
verbatim workflow line; only the input stream differs.)

## (c) GREEN — verbatim (after fix)

Patched release.sh validation 8 (verbatim from the edited file):

```
CHANGELOG_VERSION="${VERSION#v}" # v2.14.0 → 2.14.0
CHANGELOG_HEADER="## [$CHANGELOG_VERSION]"

# CHANGELOG heading convention is split by release family: formal sections are
# bare ('## [3.1.0]'), while pre-release (rc) sections carry the v prefix
# ('## [v3.0.0-rc12]'). Accept both forms: when the bare heading is absent,
# fall back to the v-prefixed one. CHANGELOG_HEADER must point at the form
# that actually matched because the tag-annotation extraction below matches
# it literally.
if ! grep -q "^## \[$CHANGELOG_VERSION\]" CHANGELOG.md; then
    if grep -q "^## \[v$CHANGELOG_VERSION\]" CHANGELOG.md; then
        CHANGELOG_HEADER="## [v$CHANGELOG_VERSION]"
    else
        die "CHANGELOG.md missing section '## [$CHANGELOG_VERSION]' (or '## [v$CHANGELOG_VERSION]'). Add release notes first."
    fi
fi
log_ok "CHANGELOG.md contains $CHANGELOG_HEADER section"
```

Probes:

```
=== GREEN-1: patched validation-8 logic (verbatim expressions from release.sh), rc12 ===
OK: CHANGELOG.md contains ## [v3.0.0-rc12] section

=== GREEN-2: tag-annotation extraction with the header the patched logic selects ===
extracted line count: 33
## [v3.0.0-rc12] - 2026-07-14

=== GREEN-3: patched check-5 pattern (release.yml L106), rc12 ===
## [v3.0.0-rc12] - 2026-07-14
exit=0

=== GREEN-4: dual-form acceptance on new-style rc headings (scratch CHANGELOG) ===
v3.1.1-rc.1 -> OK: contains ## [3.1.1-rc.1]
  check-5 pattern: MATCH
v3.1.1-rc.2 -> OK: contains ## [v3.1.1-rc.2]
  check-5 pattern: MATCH

=== NEGATIVE CONTROL: version with no section still dies (gate not weakened) ===
DIE: CHANGELOG.md missing section '## [9.8.7-rc.9]' (or '## [v9.8.7-rc.9]').
check-5: no match -> fail fires
```

GREEN-4 used a scratch file under `/tmp` containing one bare and one
v-prefixed new-style (`-rc.N`, per `.moai/docs/version-management.md`) heading
— both forms accepted by both patched gates.

## (d) Formal-path regression — verbatim

Before the fix (same computation, real heading):

```
=== BASELINE (pre-change): formal version v3.1.0 through SAME computation ===
$ VERSION=v3.1.0
$ grep -c "^## \[$CHANGELOG_VERSION\]" CHANGELOG.md
1
exit=0
$ grep -E "^## \[${VERSION#v}\]" CHANGELOG.md   # check-5 form
## [3.1.0] - 2026-08-15
exit=0

=== BASELINE (pre-change): extraction for formal 3.1.0 works ===
## [3.1.0] - 2026-08-15

### Summary
```

After the fix (same probes through the patched logic):

```
=== REGRESSION: formal v3.1.0 through the SAME patched logic (post-change) ===
OK: CHANGELOG.md contains ## [3.1.0] section
extracted line count: 267
## [3.1.0] - 2026-08-15
exit=0
```

Formal path unchanged: bare header selected, check-5 pattern matches,
extraction still produces the full section (267 lines incl. the trailing
Korean subsection).

## (e) Diff summary

```
 .github/workflows/release.yml |  6 ++++--
 scripts/release.sh            | 12 +++++++++++-
 2 files changed, 15 insertions(+), 3 deletions(-)
```

- `scripts/release.sh`: validation 8 keeps the bare-form grep first; on miss,
  falls back to the v-prefixed grep and re-points `CHANGELOG_HEADER` at the
  matched form (required — the awk extraction at L218+ and the L343+ echoes
  consume that variable). die message now names both accepted forms.
- `.github/workflows/release.yml`: check 5 pattern gains `v?` (one
  expression accepts both forms); fail message and a 2-line comment updated.
  Deliberately minimal — sibling card t54 audits check 5 against the
  release-notes contract (.moai/release-notes SSOT) and needs a small,
  well-commented surface.

Syntax gates:

```
bash -n scripts/release.sh        -> OK
actionlint .github/workflows/release.yml -> PASS (actionlint on PATH)
```

## (f) Residual risks

1. **Heading-form choice for FUTURE rc sections is now unconstrained** — both
   `## [3.1.1-rc.1]` and `## [v3.1.1-rc.1]` pass. That is inherent to option
   (b); if a single convention is wanted, that is option (a)'s territory
   (separate card). t54's release-notes contract may pin this.
2. **Existing tags v3.1.0-rc.0/.1/.2 still have no CHANGELOG sections** —
   those went out before this fix; the card records them as undocumented-path
   releases. This fix only revives the path for FUTURE rc releases; it does
   not retro-create sections (out of scope).
3. **grep dot-wildcard looseness is preexisting and unchanged** — version
   dots are unescaped in both BRE and ERE patterns (e.g. `3.1.0` matches
   `3x1y0`); `v?` adds no new ambiguity. Escaping would change matching
   semantics beyond the card scope.
4. **Not executed end-to-end** — card constraints forbid real tags and a full
   release run; the probes exercise the exact expressions but not the script
   in situ (`--dry-run` still performs `git fetch`/`ls-remote` and CI checks,
   so it was not card-scoped either). First real rc release should be
   watched.
5. **`git show ${TAG_COMMIT}:CHANGELOG.md` was proxied by the worktree
   CHANGELOG.md** in check-5 probes — the workflow itself runs at tag-push
   time against the tagged commit; the grep expression under test is
   verbatim, only the input stream differs.
