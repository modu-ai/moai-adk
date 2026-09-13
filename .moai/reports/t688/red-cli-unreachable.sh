#!/usr/bin/env bash
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel)
TEST_FILE=$(mktemp "${TMPDIR:-/tmp}/t688-cli-test.XXXXXX.go")
OVERLAY=$(mktemp "${TMPDIR:-/tmp}/t688-cli-overlay.XXXXXX.json")
cleanup() {
  rm -f "$TEST_FILE" "$OVERLAY"
}
trap cleanup EXIT

cat >"$TEST_FILE" <<'GOEOF'
package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestT688CLIExistingNonAncestorStampExitsTwoWithRecovery(t *testing.T) {
	root := newCheckCLIRepo(t)
	stampAllLayers(t, root)
	baseBranch := checkFixtureGit(t, root, "symbolic-ref", "--short", "HEAD")
	checkFixtureGit(t, root, "switch", "-q", "-c", "stamp-side")
	if err := os.WriteFile(filepath.Join(root, "internal", "side.go"), []byte("package internal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checkFixtureGit(t, root, "add", "internal/side.go")
	checkFixtureGit(t, root, "commit", "-q", "-m", "side stamp")
	stamp := checkFixtureGit(t, root, "rev-parse", "HEAD")
	checkFixtureGit(t, root, "switch", "-q", baseBranch)
	if err := os.WriteFile(filepath.Join(root, ".moai", "project", "codemaps", "provenance.json"), marshalCodemapsProvenance(t, root, stamp), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newGraphCmd()
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs([]string{"check", "--root", root})
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	err := cmd.Execute()
	var codeErr interface{ ExitCode() int }
	if !errors.As(err, &codeErr) || codeErr.ExitCode() != 2 {
		t.Fatalf("existing non-ancestor stamp must exit 2; err=%v stdout=%q stderr=%q", err, out.String(), errOut.String())
	}
	surface := strings.ToLower(errOut.String())
	for _, token := range []string{"unreachable", "freshness unmeasured", "regenerate", "stamp"} {
		if !strings.Contains(surface, token) {
			t.Fatalf("stderr missing recovery token %q: %s", token, errOut.String())
		}
	}
	if strings.Contains(out.String(), "metric=described-source-diff value=") {
		t.Fatalf("unmeasured error rendered a numeric freshness row: %s", out.String())
	}
}
GOEOF

python3 - "$ROOT" "$TEST_FILE" "$OVERLAY" <<'PYEOF'
import json
import pathlib
import sys
root, source, overlay = sys.argv[1:]
target = str(pathlib.Path(root) / "internal/cli/t688_red_ancestry_test.go")
pathlib.Path(overlay).write_text(json.dumps({"Replace": {target: source}}))
PYEOF

go test -overlay="$OVERLAY" -count=1 -run '^TestT688CLIExistingNonAncestorStampExitsTwoWithRecovery$' ./internal/cli
