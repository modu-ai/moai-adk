#!/bin/sh
# debug helper: dump what the validator's parser sees for one fixture
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
cd "$HERE/matrix-space-value"
yq -P '.' .github/workflows/ci.yml
