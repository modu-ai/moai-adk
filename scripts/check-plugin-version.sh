#!/bin/sh
# check-plugin-version.sh <tag>
#
# SPEC-PLUGIN-MARKETPLACE-001 REQ-024: the plugin version committed in
# plugins/moai/.claude-plugin/plugin.json must equal the release tag with its
# leading "v" stripped. Runs offline; the release workflow calls it inside
# verify-provenance (check 8) with the pushed tag.
#
#   exit 0  the two versions are equal
#   exit 1  they differ; both values are printed
#   exit 2  no tag argument, or the manifest cannot be read
set -eu

if [ "$#" -lt 1 ] || [ -z "$1" ]; then
  echo "usage: check-plugin-version.sh <tag>" >&2
  exit 2
fi

tag="$1"
want="${tag#v}"

root="$(cd "$(dirname "$0")/.." && pwd)"
manifest="$root/plugins/moai/.claude-plugin/plugin.json"
if [ ! -f "$manifest" ]; then
  echo "check-plugin-version: cannot read $manifest" >&2
  exit 2
fi

have="$(sed -n 's/^[[:space:]]*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$manifest" | head -n 1)"
if [ -z "$have" ]; then
  echo "check-plugin-version: no version field in $manifest" >&2
  exit 2
fi

if [ "$have" != "$want" ]; then
  echo "check-plugin-version: plugin version '$have' (plugins/moai/.claude-plugin/plugin.json) != tag '$tag' (expected '$want')" >&2
  exit 1
fi
echo "check-plugin-version: plugin version $have matches tag $tag"
