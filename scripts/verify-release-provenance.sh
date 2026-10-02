#!/usr/bin/env bash
# scripts/verify-release-provenance.sh — release provenance gate (7 checks)
#
# Usage:
#   scripts/verify-release-provenance.sh <tag>      # e.g. v3.1.0
#
# Run from the repository root. Read-only: it inspects the tag and the tagged
# commit through git and writes nothing. It does NOT fetch — the caller (the
# verify-provenance job of .github/workflows/release.yml) fetches the tag and
# origin/main first, because those two fetches need the network and this script
# is meant to be runnable against a scratch repository.
#
# The checks (the first failing one decides the verdict):
#   1  the tag is an annotated tag
#   2  its annotation carries the `Released-via: harness:release` trailer
#   3  the trailer's Release-version equals the tag name
#   4  the trailer's Release-commit equals the tagged commit
#   5  CHANGELOG.md at the tagged commit has this version's section
#   6  .moai/config/sections/system.yaml at the tagged commit has version: <tag>
#   7  the tagged commit is an ancestor of origin/main
#
# Pre-release tags (SemVer -rc.N) skip checks 5 and 6 and keep 1-4 and 7: a
# release candidate has no CHANGELOG section or version bump of its own, but who
# tagged which commit, by which route, is exactly as binding as for a final tag.
# No substitute check takes the place of the two that are skipped, and the
# verdict says so explicitly. Tags without a pre-release suffix run all seven
# checks unchanged. Only the project's own `-rc.N` form counts; the legacy
# undotted `-rcN` and other pre-release identifiers keep all seven.
#
# Exit 0 = every applicable check passed; exit 1 = a check failed (the verdict
# line names it); exit 2 = bad usage.
#
# Honest limitation: the trailer lives in a public tag annotation, so anyone with
# push rights can copy it. The gate stops accidental and habitual releases
# through unsanctioned routes; the substantive protection is checks 5-7 and, for
# a pre-release tag, check 7 and the commit binding of check 4.

set -euo pipefail

if [ "$#" -ne 1 ] || [ -z "${1}" ]; then
  echo "usage: scripts/verify-release-provenance.sh <tag>" >&2
  exit 2
fi
TAG="$1"

fail() {
  echo "::error::RELEASE_PROVENANCE_GATE: $1"
  echo "Releases must be produced by the /harness:release maintainer harness."
  exit 1
}

# Check 1 — annotated tag (lightweight tags from `gh release create` or
# the GitHub web UI resolve to `commit` and are rejected here).
TAG_TYPE="$(git cat-file -t "${TAG}")"
if [ "${TAG_TYPE}" != "tag" ]; then
  fail "check 1 (annotated tag): '${TAG}' is a ${TAG_TYPE} object, not an annotated tag."
fi

ANNOTATION="$(git for-each-ref "refs/tags/${TAG}" --format='%(contents)')"

# Check 2 — provenance trailer written by scripts/release.sh.
# No `grep -q` in any pipeline below: -q exits on first match and
# closes the pipe, the producer dies of SIGPIPE (141), and pipefail
# then reports failure for a check that MATCHED. Redirect instead so
# grep consumes all of its input.
if ! printf '%s\n' "${ANNOTATION}" | grep -xF 'Released-via: harness:release' >/dev/null; then
  fail "check 2 (provenance trailer): tag annotation has no 'Released-via: harness:release' line."
fi

# Check 3 — the trailer's version equals the pushed tag name.
TRAILER_VERSION="$(printf '%s\n' "${ANNOTATION}" | sed -n 's/^Release-version: //p' | tail -n1)"
if [ "${TRAILER_VERSION}" != "${TAG}" ]; then
  fail "check 3 (version binding): trailer Release-version='${TRAILER_VERSION}' != pushed tag '${TAG}'."
fi

# Check 4 — the trailer's commit equals the commit the tag points to.
TAG_COMMIT="$(git rev-list -n1 "${TAG}")"
TRAILER_COMMIT="$(printf '%s\n' "${ANNOTATION}" | sed -n 's/^Release-commit: //p' | tail -n1)"
if [ "${TRAILER_COMMIT}" != "${TAG_COMMIT}" ]; then
  fail "check 4 (commit binding): trailer Release-commit='${TRAILER_COMMIT}' != tagged commit '${TAG_COMMIT}'."
fi

# Pre-release tags (-rc.N) skip checks 5 and 6; see the header. The grammar is
# the project's own `-rc.N` (no leading zero), the form scripts/release.sh
# produces. A tag outside it — the legacy undotted `-rcN`, `-beta.1`, build
# metadata — is treated as a formal tag and keeps all seven checks.
if [[ "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-rc\.(0|[1-9][0-9]*)$ ]]; then
  echo "RELEASE_PROVENANCE_GATE: pre-release tag ${TAG}: skipping check 5 (CHANGELOG) and check 6 (version SSOT) per the rc rule."
  APPLICABLE_CHECKS="5 applicable checks"
  SKIPPED_NOTE="; checks 5 and 6 skipped (pre-release)"
else
  # Check 5 — CHANGELOG.md at the tagged commit has this version's section.
  # Formal sections are bare ('## [3.1.0]'); pre-release (rc) sections
  # carry the v prefix ('## [v3.0.0-rc12]'). 'v?' accepts both forms.
  VERSION_NO_V="${TAG#v}"
  if ! git show "${TAG_COMMIT}:CHANGELOG.md" | grep -E "^## \[v?${VERSION_NO_V}\]" >/dev/null; then
    fail "check 5 (CHANGELOG): CHANGELOG.md at ${TAG_COMMIT} has no '## [${VERSION_NO_V}]' or '## [v${VERSION_NO_V}]' section."
  fi

  # Check 6 — version SSOT (.moai/config/sections/system.yaml stores the
  # v-prefixed form, e.g. `version: v3.0.1`).
  SSOT_PATH='.moai/config/sections/system.yaml'
  if ! git show "${TAG_COMMIT}:${SSOT_PATH}" | grep -E "^[[:space:]]*version:[[:space:]]*${TAG}[[:space:]]*$" >/dev/null; then
    ACTUAL="$(git show "${TAG_COMMIT}:${SSOT_PATH}" | sed -n 's/^[[:space:]]*version:[[:space:]]*//p' | head -n1)"
    fail "check 6 (version SSOT): ${SSOT_PATH} at ${TAG_COMMIT} has version='${ACTUAL}', expected '${TAG}'."
  fi
  APPLICABLE_CHECKS="all 7 checks"
  SKIPPED_NOTE=""
fi

# Check 7 — the tagged commit is an ancestor of origin/main.
if ! git merge-base --is-ancestor "${TAG_COMMIT}" origin/main; then
  fail "check 7 (main ancestry): tagged commit ${TAG_COMMIT} is not an ancestor of origin/main."
fi

echo "RELEASE_PROVENANCE_GATE: ${APPLICABLE_CHECKS} passed for ${TAG} (${TAG_COMMIT})${SKIPPED_NOTE}."
