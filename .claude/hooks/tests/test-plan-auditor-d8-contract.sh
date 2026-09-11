#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)
auditor="$root/.claude/agents/moai/plan-auditor.md"
grep -q 'scoped per section' "$auditor"
grep -q 'cannot cover a `syscall` mention' "$auditor"

fixture=$(mktemp -d)
trap 'rm -rf "$fixture"' EXIT
cat > "$fixture/untagged.md" <<'EOF'
### Runtime syscall

The implementation uses syscall.Flock.
EOF
cat > "$fixture/tagged.md" <<'EOF'
### Runtime syscall

//go:build unix
The implementation uses syscall.Flock.
EOF

scan() {
  awk '
    function flush() {
      if (has_sys && !has_tag)
        printf "BLOCKING: section %s references syscall without a local constraint\\n", head
    }
    BEGIN { head = "(before the first heading)" }
    /^#+ / { flush(); head = $0; has_sys = 0; has_tag = 0 }
    /syscall/ { has_sys = 1 }
    /\/\/go:build|cross-platform exemption|EXCL.*syscall/ { has_tag = 1 }
    END { flush() }
  ' "$1"
}

untagged=$(scan "$fixture/untagged.md")
grep -q 'BLOCKING:' <<< "$untagged"
tagged=$(scan "$fixture/tagged.md")
[[ -z "$tagged" ]]
printf 'PASS: D8 binds syscall coverage to its own section\n'
