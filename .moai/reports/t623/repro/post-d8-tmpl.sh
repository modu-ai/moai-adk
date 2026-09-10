# Detect syscall mentions whose own section carries no build-tag constraint or exemption
if [ ! -r new-spec.md ]; then
  echo "GAP: new-spec.md is not readable — D8 was not observed"
else
  awk '
    function flush() {
      if (has_sys && !has_tag)
        printf "BLOCKING: section \"%s\" references syscall but carries no //go:build constraint or EXCL justification\n", head
    }
    BEGIN { head = "(before the first heading)" }
    /^#+ / { flush(); head = $0; has_sys = 0; has_tag = 0 }
    /syscall/ { has_sys = 1 }
    /\/\/go:build|cross-platform exemption|EXCL.*syscall/ { has_tag = 1 }
    END { flush() }
  ' new-spec.md
fi
