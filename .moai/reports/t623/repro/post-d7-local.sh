# Surface cross-SPEC reconciliation candidates; the auditor decides BLOCKING
grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' new-spec.md | sort -u | while read SID; do
  if [ -f ".moai/specs/$SID/spec.md" ]; then
    STATUS=$(grep '^status:' ".moai/specs/$SID/spec.md" | head -1 | cut -d: -f2 | tr -d ' ')
    case "$STATUS" in
      retired|superseded|archived)
        echo "REVIEW: $SID has status=$STATUS — confirm explicit reconciliation in the same section or paragraph before emitting BLOCKING"
        # Paragraphs naming $SID next to a reconciliation keyword, for the auditor to read
        awk -v sid="$SID" 'BEGIN { RS = "" }
          index($0, sid) && tolower($0) ~ /revers|supersed|absorb|carve-out/ {
            gsub(/\n/, " "); print "  reconciliation candidate (paragraph " NR "): " $0
          }' new-spec.md
        ;;
    esac
  else
    echo "SHOULD: referenced SPEC $SID not found in .moai/specs/"
  fi
done
