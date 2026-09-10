# Surface cross-artifact ordering candidates; the auditor decides CN-4 and MP-9
plan=".moai/specs/SPEC-STOPCHAIN-TRIM-001/plan.md"
acc=".moai/specs/SPEC-STOPCHAIN-TRIM-001/acceptance.md"
if [ ! -r "$plan" ] || [ ! -r "$acc" ]; then
  [ -r "$plan" ] || echo "GAP: $plan is not readable — milestone order was not observed"
  [ -r "$acc" ] || echo "GAP: $acc is not readable — ordering clauses were not observed"
else
  awk '
    function bindexit(seg, ms,   tok, prefix, rest) {
      while (match(seg, acid)) {
        tok = substr(seg, RSTART, RLENGTH); bind[tok] = ms; nb++
        prefix = tok; sub(/[0-9]+$/, "", prefix)
        seg = substr(seg, RSTART + RLENGTH)
        while (match(seg, /^[ \t]*,[ \t]*[0-9]+/)) {
          rest = substr(seg, RLENGTH + 1)
          tok = substr(seg, 1, RLENGTH); sub(/^[ \t]*,[ \t]*/, "", tok)
          bind[prefix tok] = ms; nb++
          seg = rest
        }
      }
    }
    function flush(   low, kpos, head, tail, subj, rel, ms) {
      if (rec == "") return
      low = tolower(rec)
      if (match(low, kw)) {
        nc++
        printf "CANDIDATE: %s:%d: %s\n", accname, recline, rec
        kpos = RSTART; rel = substr(low, RSTART, RLENGTH)
        head = substr(rec, 1, kpos - 1); tail = substr(rec, kpos + RLENGTH)
        subj = ""
        if (match(head, acid)) subj = substr(head, RSTART, RLENGTH)
        else if (secac != "") subj = secac
        if (match(tail, /M[0-9]+/)) {
          ms = substr(tail, RSTART, RLENGTH)
          if (subj != "" && (subj in bind) && (ms in pos)) {
            if (rel !~ /after/ && pos[bind[subj]] > pos[ms])
              printf "CONFLICT: %s:%d orders %s before %s, but %s binds %s to the exit of %s, which the plan places after %s\n", accname, recline, subj, ms, planname, subj, bind[subj], ms
            if (rel ~ /after/ && pos[bind[subj]] < pos[ms])
              printf "CONFLICT: %s:%d orders %s after %s, but %s binds %s to the exit of %s, which the plan places before %s\n", accname, recline, subj, ms, planname, subj, bind[subj], ms
          }
        }
      }
      rec = ""
    }
    BEGIN {
      acid = "AC-([A-Z][A-Z0-9]*-)*[0-9]+"
      kw = "(before|after|first|prior to|pre-change)"
      planname = ARGV[1]; accname = ARGV[2]
    }
    { sub(/\r$/, "") }
    FILENAME == ARGV[1] {
      if (match($0, /^#+/) && substr($0, RLENGTH + 1) ~ /^[ \t]+(Milestone[ \t]+)?M[0-9]+([^0-9]|$)/) {
        mlevel = RLENGTH; match($0, /M[0-9]+/); cur = substr($0, RSTART, RLENGTH)
        if (!(cur in pos)) { pos[cur] = ++nm; order = order " " cur }
        next
      }
      if (match($0, /^#+[ \t]/) && RLENGTH - 1 <= mlevel) { cur = ""; next }
      if (cur != "" && tolower($0) ~ /^(\*\*)?exit(\*\*)?[ \t]*:/) bindexit($0, cur)
      next
    }
    {
      if ($0 ~ /^[ \t]*$/ || $0 ~ /^#+[ \t]/ || $0 ~ /^[ \t]*([-*+]|[0-9]+\.)[ \t]/ || $0 ~ /^[ \t]*\|/) {
        flush()
        if ($0 ~ /^#+[ \t]/) { secac = ""; if (match($0, acid)) secac = substr($0, RSTART, RLENGTH); next }
        if ($0 ~ /^[ \t]*$/) next
        rec = $0; recline = FNR
        if ($0 ~ /^[ \t]*\|/) flush()
        next
      }
      if (rec == "") { rec = $0; recline = FNR } else rec = rec " " $0
    }
    END {
      flush()
      printf "COLLECTED: %d milestones in plan order (%s), %d exit bindings, %d ordering candidates\n", nm, (nm ? substr(order, 2) : "none"), nb, nc
      if (nm == 0) print "GAP: 0 milestone headings collected from the plan — milestone order not observed"
      if (nc == 0) printf "NONE: 0 records in %s match %s (case-insensitive) — an observed absence\n", accname, kw
    }
  ' "$plan" "$acc"
fi
