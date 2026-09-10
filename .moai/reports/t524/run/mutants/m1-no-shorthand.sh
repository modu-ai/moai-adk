# MUTANT m1 (card t524): the Group 4 verb with the shorthand-expansion loop removed.
# Expected: fixture A now reports UNCOVERED: REQ-FIXA-002.
spec="<new-spec.md>"
acc="<acceptance.md>"
if [ ! -r "$spec" ]; then
  echo "GAP: $spec is not readable — traceability was not observed"
else
  set -- "$spec"
  accstate=absent
  if [ -r "$acc" ]; then set -- "$@" "$acc"; accstate=read; fi
  awk -v accstate="$accstate" '
    function scan(seg,   tok, prefix, rest, c) {
      while (match(seg, id)) {
        tok = substr(seg, RSTART, RLENGTH)
        mapped[tok] = 1
        prefix = tok
        sub(/[0-9]+$/, "", prefix)
        seg = substr(seg, RSTART + RLENGTH)
      }
    }
    BEGIN {
      id = "REQ-([A-Z][A-Z0-9]*-)*[0-9]+"
      deflist = "^[ \t]*[-*+][ \t]+(\\*\\*" id "|" id "[ \t]*:)"
      defhead = "^#+[ \t]+(\\*\\*)?" id
      defrow = "^[ \t]*\\|[ \t]*(\\*\\*)?" id "(\\*\\*)?[ \t]*\\|"
    }
    { sub(/\r$/, "") }
    FILENAME == ARGV[1] && ($0 ~ deflist || $0 ~ defhead || $0 ~ defrow) {
      match($0, id)
      def[substr($0, RSTART, RLENGTH)] = 1
    }
    (" " $0) ~ /[^A-Za-z0-9_-]AC-([A-Z0-9]+-)*[0-9]+/ {
      n = split($0, cells, "|")
      for (i = 1; i <= n; i++) scan(cells[i])
    }
    END {
      count = 0
      for (k in def) count++
      printf "COLLECTED: %d REQ definitions (acceptance input: %s)\n", count, accstate
      if (count == 0) { print "GAP: 0 REQ definitions collected — traceability not observed"; exit }
      for (k in def) if (!(k in mapped)) print "UNCOVERED: " k
      for (k in mapped) if (!(k in def)) print "ORPHAN: " k
    }
  ' "$@" | LC_ALL=C sort
fi
