#!/bin/zsh
# t278 M1 — re-fetch the 16 jobs files that collected a secondary-rate-limit
# body instead of job rows during sweep v1 (single-flight, 2s spacing).
set -u
for id in \
  31499288220 31626420856 31628085639 31627777786 31639333372 \
  31641842385 31717046489 31734591740 31728442148 31791185823 \
  31797542345 31744543706 31783552189 31789379728 31797322925 \
  31853270653
do
  gh api "repos/modu-ai/moai-adk/actions/runs/$id/jobs?per_page=100" \
    --jq '.jobs[] | "\(.name)\t\(.conclusion)"' > "/tmp/t278-sweep/$id.jobs" 2>/dev/null
  sleep 2
done
echo "REFETCH_DONE"
