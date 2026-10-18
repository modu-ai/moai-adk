#!/bin/sh
# RED-now probe for AC-PAS-016 / AC-PAS-001 (change to a), AC-PAS-004 (test-only change),
# AC-PAS-006 (docs-only change) and AC-PAS-011 (invalid scope mode).
# It is a reachable green path: once the gate is implemented it prints GREEN for every scenario and exits 0.
# The wrapper `go` on PATH delegates every verb except `test` to the real go (so a `go list` graph query
# works) and records only `test` invocations, returning success for them without running tests.
# Each commit scenario is GREEN only when the commit itself succeeded (exit 0, HEAD moved) AND the recorded
# `go test` argv equals the wanted one ("none" = no `go test` invocation at all).
# Usage: PROBE_DIR=<scratch dir> sh probe016.sh <moai binary built from the tree under measurement>
set -u
MOAI="$1"
REALGO="$(command -v go)"
W="${PROBE_DIR:-/private/tmp/pas-probe}/p016"
rm -rf "$W"; mkdir -p "$W/bin" "$W/hooks" "$W/repo/a" "$W/repo/b" "$W/repo/c" "$W/repo/docs" "$W/repo/.moai/config/sections"
cat > "$W/bin/go" <<EOF
#!/bin/sh
case "\$1" in
  test) echo "\$@" >> "$W/go.log"; exit 0 ;;
  *) exec "$REALGO" "\$@" ;;
esac
EOF
chmod +x "$W/bin/go"
cat > "$W/hooks/pre-commit" <<EOF
#!/bin/sh
MOAI_PRECOMMIT=1 CLAUDE_PROJECT_DIR="$W/repo" PATH="$W/bin:\$PATH" "$MOAI" gate
EOF
chmod +x "$W/hooks/pre-commit"
cd "$W/repo" || exit 2
printf 'module example.com/m\n\ngo 1.23\n' > go.mod
printf 'package a\n\nfunc A() int { return 1 }\n' > a/a.go
printf 'package a\n\nimport "testing"\n\nfunc TestA(t *testing.T) {}\n' > a/a_test.go
printf 'package b\n\nimport "example.com/m/a"\n\nfunc B() int { return a.A() }\n' > b/b.go
printf 'package b\n\nimport "testing"\n\nfunc TestB(t *testing.T) {}\n' > b/b_test.go
printf 'package c\n\nfunc C() int { return 3 }\n' > c/c.go
printf 'package c\n\nimport "testing"\n\nfunc TestC(t *testing.T) {}\n' > c/c_test.go
printf 'notes\n' > docs/note.md
cat > .moai/config/sections/gate.yaml <<'EOF'
gate:
  enabled: true
  disabled_steps:
    "go vet": false
    "golangci-lint": false
  typecheck:
    enabled: false
  pre_commit:
    enabled: true
  ast_grep_gate:
    enabled: false
  graph_freshness:
    enabled: false
EOF
git init -q .
git config user.email probe@example.com
git config user.name probe
git config core.hooksPath "$W/hooks"
git add -A
git commit -q -m base --no-verify
RC=0
scenario() { # name wanted-argv ("none" = no go test call)
  NAME="$1"; WANT="$2"
  H0="$(git rev-parse HEAD)"
  git commit -q -m "$NAME"
  CRC=$?
  H1="$(git rev-parse HEAD)"
  GOT="$(sed -n 1p "$W/go.log")"; [ -n "$GOT" ] || GOT=none
  if [ "$CRC" -eq 0 ] && [ "$H0" != "$H1" ] && [ "$GOT" = "$WANT" ]; then S=GREEN; else S=RED; RC=1; fi
  echo "$NAME: $S  (commit exit=$CRC, HEAD moved=$([ "$H0" != "$H1" ] && echo yes || echo no), wanted go argv '$WANT', got '$GOT')"
}
printf 'package a\n\nfunc A() int { return 2 }\n' > a/a.go; git add a/a.go; : > "$W/go.log"
scenario "AC-016/001 change a" "test ./a ./b"
printf 'package a\n\nimport "testing"\n\nfunc TestA(t *testing.T) { _ = 1 }\n' > a/a_test.go; git add a/a_test.go; : > "$W/go.log"
scenario "AC-004 test-only change" "test ./a"
printf 'more notes\n' >> docs/note.md; git add docs/note.md; : > "$W/go.log"
scenario "AC-006 docs-only change" "none"
# AC-011: an invalid scope mode must be refused with a message naming the key
sed -i.bak 's/^  pre_commit:/  pre_commit:\n    test_scope:\n      mode: bogus/' .moai/config/sections/gate.yaml
OUT="$(MOAI_PRECOMMIT=1 CLAUDE_PROJECT_DIR="$W/repo" PATH="$W/bin:$PATH" "$MOAI" gate 2>&1)"; GRC=$?
case "$OUT" in *gate.pre_commit.test_scope.mode*) NAMED=yes ;; *) NAMED=no ;; esac
if [ "$GRC" -ne 0 ] && [ "$NAMED" = yes ]; then S=GREEN; else S=RED; RC=1; fi
echo "AC-011 mode: bogus: $S  (gate exit=$GRC, message names the key=$NAMED)"
echo "exit=$RC"
exit $RC
