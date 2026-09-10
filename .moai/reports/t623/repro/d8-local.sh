# Detect syscall introduction without build-tag constraint
if grep -q 'syscall' new-spec.md; then
  if ! grep -qE '//go:build|cross-platform exemption|EXCL.*syscall' new-spec.md; then
    echo "BLOCKING: SPEC references syscall but no //go:build constraint or EXCL justification"
  fi
fi
