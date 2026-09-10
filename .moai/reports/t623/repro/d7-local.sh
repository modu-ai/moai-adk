# Extract SPEC-ID references and check their cross-SPEC status
grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' new-spec.md | sort -u | while read SID; do
  if [ -f ".moai/specs/$SID/spec.md" ]; then
    STATUS=$(grep '^status:' ".moai/specs/$SID/spec.md" | head -1 | cut -d: -f2 | tr -d ' ')
    case "$STATUS" in
      retired|superseded|archived)
        echo "BLOCKING: $SID has status=$STATUS but is referenced without reconciliation"
        ;;
    esac
  else
    echo "SHOULD: referenced SPEC $SID not found in .moai/specs/"
  fi
done
