#!/usr/bin/env python3
from pathlib import Path

workflow = Path(".github/workflows/graph-freshness.yml").read_text()
start = workflow.index('if [ -z "${GITHUB_BASE_REF:-}" ]')
end = workflow.index('if [ "${GITHUB_HEAD_REF#release/}"', start)
push_branch = workflow[start:end]

has_head_target = 'TARGET="HEAD"' in push_branch
exits_early = "exit 0" in push_branch
print(f"push_branch_has_target_head={str(has_head_target).lower()}")
print(f"push_branch_exits_before_ancestry={str(exits_early).lower()}")
raise SystemExit(0 if has_head_target and not exits_early else 1)
