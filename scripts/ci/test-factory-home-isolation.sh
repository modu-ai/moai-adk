#!/bin/sh
set -eu
unset GIT_CONFIG GIT_DIR GIT_COMMON_DIR GIT_WORK_TREE GIT_INDEX_FILE \
  GIT_OBJECT_DIRECTORY GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_NAMESPACE \
  GIT_PREFIX GIT_SUPER_PREFIX GIT_CEILING_DIRECTORIES GIT_QUARANTINE_PATH
root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM
mkdir "$tmp/home"
HOME="$tmp/home" USERPROFILE="$tmp/home" GOTELEMETRY=off \
  go test "$root/internal/cli" -run '^(TestFactoryPROpenReleasesLaneForIndependentCard|TestFactorySameOwnerReacquiresExpiredLease|TestFactoryExpiredLeaseReacquireKeepsDirtyTreeAndRejectsDuplicate|TestFactoryExpiredLeaseReacquireRefusalsPreserveState|TestExpiredReacquirePreservesClaimDeadlineAndPRSnapshot|TestExpiredReacquireBusyRecordIsBounded|TestExpiredReacquireUnderDriftLogLock)$' -count=1 -timeout 5m
if [ -e "$tmp/home/.moai" ]; then
  echo 'factory handoff tests left state in the caller HOME' >&2
  exit 1
fi
echo 'caller HOME has no Factory test state: PASS'
