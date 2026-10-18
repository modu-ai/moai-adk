#!/bin/sh
# RED-now probe for AC-PAS-016 / AC-PAS-001 (REQ-PAS-001, REQ-PAS-002):
# what argv does the pre-commit gate hand to `go test`?
# Usage: PROBE_DIR=<scratch dir> sh probe016.sh <moai binary built from the tree under measurement>
set -u
MOAI="$1"
W="${PROBE_DIR:-/private/tmp/pas-probe}/p016"
rm -rf "$W"; mkdir -p "$W/bin" "$W/hooks" "$W/repo/a" "$W/repo/b" "$W/repo/c" "$W/repo/.moai/config/sections"
cat > "$W/bin/go" <<'EOF'
#!/bin/sh
echo "$@" >> "$FAKE_GO_LOG"
exit 0
EOF
chmod +x "$W/bin/go"
cat > "$W/hooks/pre-commit" <<EOF
#!/bin/sh
MOAI_PRECOMMIT=1 CLAUDE_PROJECT_DIR="$W/repo" FAKE_GO_LOG="$W/go.log" PATH="$W/bin:\$PATH" "$MOAI" gate
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
: > "$W/go.log"
git commit -q -m base --no-verify
printf 'package a\n\nfunc A() int { return 2 }\n' > a/a.go
git add a/a.go
: > "$W/go.log"
git commit -q -m "change a"
echo "commit exit=$?"
echo "--- fake go argv:"
cat "$W/go.log"
WANT="test ./a ./b"
GOT="$(sed -n 1p "$W/go.log")"
if [ "$GOT" = "$WANT" ]; then echo "RESULT: GREEN (argv == $WANT)"; exit 0; fi
echo "RESULT: RED (wanted argv '$WANT', got '$GOT')"; exit 1
