#!/usr/bin/env bash
# Count canonical 12 frontmatter fields per SPEC artifact.
cd "$(git rev-parse --show-toplevel)/.moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001" || exit 1
for f in spec.md plan.md acceptance.md design.md research.md progress.md; do
    n=$(awk '/^---$/{c++; next} c==1' "$f" | grep -cE '^(id|title|version|status|created|updated|author|priority|phase|module|lifecycle|tags):')
    echo "$f: $n/12"
done
echo "--- rejected snake_case aliases (expect none) ---"
grep -nE '^(created_at|updated_at|labels|spec_id):' ./*.md || echo "none"
echo "--- Out of Scope H3 sub-headings in spec.md ---"
grep -c '^### Out of Scope — ' spec.md
