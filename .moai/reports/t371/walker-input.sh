#!/bin/sh
# t371: dump the StatusGitConsistency walker's input per SPEC-ID.
# Replicates internal/spec/drift.go getGitImpliedStatus:
#   git log <branch> --oneline --no-merges --grep=<specID> -50
while read -r id; do
  printf '### %s\n' "$id"
  printf 'frontmatter: '
  grep -m1 '^status:' ".moai/specs/$id/spec.md" 2>/dev/null || echo '(none)'
  printf 'main-walk:\n'
  git log main --oneline --no-merges --grep="$id" -6
  printf 'develop-walk:\n'
  git log develop --oneline --no-merges --grep="$id" -6
  printf '\n'
done < .moai/reports/t371/statusgit-18-ids.txt
