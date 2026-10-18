#!/bin/sh
# RED-now probe for AC-PAS-009 / AC-PAS-005 (REQ-PAS-005, REQ-PAS-009).
# The hook records (a) the commit-accurate list (git diff --cached --name-only --no-renames, environment as git
# gave it) and (b) the list in the form the gate's existing staged-file helper computes it
# (git diff --cached --name-only with gitenv.RepoScopingVars removed from the environment).
# Exits 1 when (b) differs from (a) in any of the three commit shapes.
# Usage: PROBE_DIR=<scratch dir> sh probe009.sh
set -u
W="${PROBE_DIR:-/private/tmp/pas-probe}/p009"
rm -rf "$W"; mkdir -p "$W/hooks" "$W/repo/a" "$W/repo/d" "$W/repo/old"
cat > "$W/hooks/pre-commit" <<EOF
#!/bin/sh
SCRUB="env -u GIT_DIR -u GIT_COMMON_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_OBJECT_DIRECTORY -u GIT_ALTERNATE_OBJECT_DIRECTORIES -u GIT_NAMESPACE -u GIT_PREFIX -u GIT_SUPER_PREFIX -u GIT_CEILING_DIRECTORIES -u GIT_QUARANTINE_PATH"
git diff --cached --name-only --no-renames | sort | tr '\n' ' ' > "$W/exact.txt"
\$SCRUB git diff --cached --name-only | sort | tr '\n' ' ' > "$W/helper.txt"
exit 0
EOF
chmod +x "$W/hooks/pre-commit"
cd "$W/repo" || exit 2
git init -q .
git config user.email probe@example.com
git config user.name probe
git config core.hooksPath "$W/hooks"
echo 1 > a/x.go; echo 1 > d/z.go; echo 1 > old/o.go
git add -A
git commit -q -m base --no-verify
RC=0
report() {
  E="$(cat "$W/exact.txt")"; H="$(cat "$W/helper.txt")"
  if [ "$E" = "$H" ]; then S=same; else S=DIFFERENT; RC=1; fi
  echo "$1"; echo "  commit records : $E"; echo "  helper form    : $H   [$S]"
}
echo 2 > a/x.go; echo 2 > d/z.go; git add a/x.go
git commit -q -a -m ii
report "(ii)  git commit -a            (a/x.go staged, d/z.go modified unstaged)"
echo 3 > a/x.go; echo 3 > d/z.go; git add a/x.go
git commit -q -m iii d/z.go
report "(iii) git commit d/z.go        (path-limited)"
git reset -q; mkdir -p new; git mv old/o.go new/o.go
git commit -q -m iv
report "(iv)  git mv old/o.go new/o.go (rename)"
echo "exit=$RC"
exit $RC
