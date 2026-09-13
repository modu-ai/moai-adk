#!/usr/bin/env bash
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)
TEST_FILE=$(mktemp "${TMPDIR:-/tmp}/t688-ancestry-test.XXXXXX.go")
OVERLAY=$(mktemp "${TMPDIR:-/tmp}/t688-overlay.XXXXXX.json")
cleanup() {
  rm -f "$TEST_FILE" "$OVERLAY"
}
trap cleanup EXIT

cat >"$TEST_FILE" <<'GOEOF'
package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestT688ExistingNonAncestorStampIsUnmeasured(t *testing.T) {
	root := newCheckFixture(t)
	baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
	gitFix(t, root, "switch", "-q", "-c", "stamp-side")
	if err := os.WriteFile(filepath.Join(root, "internal", "side.go"), []byte("package internal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitFix(t, root, "add", "internal/side.go")
	gitFix(t, root, "commit", "-q", "-m", "side stamp")
	stamp := gitFix(t, root, "rev-parse", "HEAD")
	gitFix(t, root, "switch", "-q", baseBranch)
	writeCodemapsProvenance(t, root, stamp)

	rep, err := checkCodemaps(root, DefaultThresholds())
	if err == nil {
		t.Fatalf("existing non-ancestor stamp must be freshness-unmeasured with a system error; got verdict=%q value=%d threshold=%d content_anchor=%q", rep.Verdict, rep.Value, rep.Threshold, rep.ContentAnchor)
	}
	if rep.Verdict != VerdictAbsent {
		t.Fatalf("verdict=%q, want %q carrier", rep.Verdict, VerdictAbsent)
	}
	if rep.Value != 0 || rep.ContentAnchor != "" || rep.ContentAnchorSource != "" || rep.Contribution != nil || len(rep.DrivingPaths) != 0 {
		t.Fatalf("unmeasured path fabricated freshness fields: %+v", rep)
	}
}
GOEOF

python3 - "$ROOT" "$TEST_FILE" "$OVERLAY" <<'PYEOF'
import json
import pathlib
import sys
root, source, overlay = sys.argv[1:]
target = str(pathlib.Path(root) / "internal/graph/t688_red_ancestry_test.go")
pathlib.Path(overlay).write_text(json.dumps({"Replace": {target: source}}))
PYEOF

go test -overlay="$OVERLAY" -count=1 -run '^TestT688ExistingNonAncestorStampIsUnmeasured$' ./internal/graph
