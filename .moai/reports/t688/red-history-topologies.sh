#!/usr/bin/env bash
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)
TEST_FILE=$(mktemp "${TMPDIR:-/tmp}/t688-history-test.XXXXXX.go")
OVERLAY=$(mktemp "${TMPDIR:-/tmp}/t688-history-overlay.XXXXXX.json")
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

func t688CommitFile(t *testing.T, root, path, content, message string) string {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitFix(t, root, "add", path)
	gitFix(t, root, "commit", "-q", "-m", message)
	return gitFix(t, root, "rev-parse", "HEAD")
}

func t688AssertDisposition(t *testing.T, root, stamp string, wantAncestor bool) {
	t.Helper()
	if got := isAncestor(t, root, stamp, "HEAD"); got != wantAncestor {
		t.Fatalf("stamp ancestor of HEAD=%v, want %v", got, wantAncestor)
	}
	writeCodemapsProvenance(t, root, stamp)
	rep, err := checkCodemaps(root, DefaultThresholds())
	if wantAncestor {
		if err != nil {
			t.Fatalf("merge topology must remain freshness-judgeable: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("object-present non-ancestor stamp must be freshness-unmeasured; got verdict=%q value=%d anchor=%q", rep.Verdict, rep.Value, rep.ContentAnchor)
	}
	if rep.Verdict != VerdictAbsent {
		t.Fatalf("verdict=%q, want %q compatibility carrier", rep.Verdict, VerdictAbsent)
	}
	if rep.Value != 0 || rep.ContentAnchor != "" || rep.ContentAnchorSource != "" || rep.Contribution != nil || len(rep.DrivingPaths) != 0 {
		t.Fatalf("unmeasured topology fabricated freshness fields: %+v", rep)
	}
}

func TestT688MergeSquashRebaseLikeTopologies(t *testing.T) {
	t.Run("merge commit preserves ancestry", func(t *testing.T) {
		root := newCheckFixture(t)
		baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
		gitFix(t, root, "switch", "-q", "-c", "stamp-side")
		stamp := t688CommitFile(t, root, "internal/stamp.go", "package internal\n", "stamp source")
		gitFix(t, root, "switch", "-q", baseBranch)
		t688CommitFile(t, root, "internal/base.go", "package internal\n", "base advance")
		gitFix(t, root, "merge", "-q", "--no-ff", "stamp-side", "-m", "merge stamp side")
		t688AssertDisposition(t, root, stamp, true)
	})

	t.Run("squash retains object but drops ancestry", func(t *testing.T) {
		root := newCheckFixture(t)
		baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
		gitFix(t, root, "switch", "-q", "-c", "stamp-side")
		stamp := t688CommitFile(t, root, "internal/stamp.go", "package internal\n", "stamp source")
		gitFix(t, root, "switch", "-q", baseBranch)
		gitFix(t, root, "merge", "-q", "--squash", "stamp-side")
		gitFix(t, root, "commit", "-q", "-m", "squash stamp side")
		t688AssertDisposition(t, root, stamp, false)
	})

	t.Run("rebase-like rewrite retains object but drops ancestry", func(t *testing.T) {
		root := newCheckFixture(t)
		baseBranch := gitFix(t, root, "symbolic-ref", "--short", "HEAD")
		gitFix(t, root, "switch", "-q", "-c", "stamp-side")
		stamp := t688CommitFile(t, root, "internal/stamp.go", "package internal\n", "stamp source")
		gitFix(t, root, "switch", "-q", baseBranch)
		t688CommitFile(t, root, "internal/base.go", "package internal\n", "base advance")
		gitFix(t, root, "cherry-pick", stamp)
		t688AssertDisposition(t, root, stamp, false)
	})
}
GOEOF

python3 - "$ROOT" "$TEST_FILE" "$OVERLAY" <<'PYEOF'
import json
import pathlib
import sys
root, source, overlay = sys.argv[1:]
target = str(pathlib.Path(root) / "internal/graph/t688_red_history_test.go")
pathlib.Path(overlay).write_text(json.dumps({"Replace": {target: source}}))
PYEOF

go test -overlay="$OVERLAY" -count=1 -run '^TestT688MergeSquashRebaseLikeTopologies$' ./internal/graph
