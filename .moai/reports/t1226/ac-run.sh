#!/bin/bash
# Non-git AC checks for SPEC-ALWAYS-LOADED-HEADROOM-001 (acceptance.md), verbatim where possible.
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1226 || exit 1
SCRATCH=${SCRATCH:?set SCRATCH to the session scratch dir}
V=.moai/reports/t1226/verdict.md
echo "### rerun harness tail"; grep -E -- '--- (PASS|SKIP|FAIL): TestHeadroomInitSurfaceExport|^ok|^FAIL' $SCRATCH/rerun.txt

echo "### AC-ALH-001"
head -20 .moai/reports/t1226/verdict.md | grep -c -E '^레인 백엔드: [^_ ][^(]* \(출처: [^_][^)]*\)$'
head -20 .moai/reports/t1226/verdict.md | grep -c -F '7fe658815eb0d4110b9acadad56e5a85bee3ed3f'
head -20 .moai/reports/t1226/verdict.md | grep -c -E '199,?111'
head -20 .moai/reports/t1226/verdict.md | grep -c -E '49,?111'
grep -E '197,?897|198,?361' .moai/reports/t1226/verdict.md | grep -v -c '폐기'
grep -c -E '^F = ' .moai/reports/t1226/verdict.md

echo "### AC-ALH-002"
locale charmap
grep -c -E '^charmap = UTF-8$' "$V"
grep -c -E '^build_head = [0-9a-f]{40}$' "$V"
grep -c -E '^count_set_init = (18|17|observed)$' "$V"
P18="CLAUDE.md AGENTS.md .moai/config/sections/user.yaml .moai/config/sections/language.yaml
.claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/core/askuser-protocol.md
.claude/rules/moai/core/moai-constitution.md .claude/rules/moai/core/moai-mcp-tools.md
.claude/rules/moai/core/native-idiom-and-register.md .claude/rules/moai/core/verification-claim-integrity.md
.claude/rules/moai/workflow/cache-aware-execution.md .claude/rules/moai/workflow/context-window-management.md
.claude/rules/moai/workflow/cross-session-messaging.md .claude/rules/moai/workflow/goal-directive.md
.claude/rules/moai/workflow/kanban-dispatch.md .claude/rules/moai/workflow/main-checkout-branch-guard.md
.claude/rules/moai/workflow/session-handoff.md .claude/rules/moai/workflow/skill-routing.md"
printf '%s\n' $P18 | sort | diff - <(sort .moai/reports/t1226/count-set-live.txt) && echo PIN-live-OK
printf '%s\n' $P18 | sort | diff - <(sort .moai/reports/t1226/count-set-init.txt) && echo PIN-init-OK
(cd "$SCRATCH/live" && xargs wc -m < "$OLDPWD/.moai/reports/t1226/count-set-live.txt" | tail -1)
(cd "$SCRATCH/init-surface" && xargs wc -m < "$OLDPWD/.moai/reports/t1226/count-set-init.txt" | tail -1)
grep -E '^total_(live|init) = [0-9]+$' "$V"
awk '/^## S_init/,/^## S_live/' "$V" | grep -c '_미측정'
awk '/^## S_live/,/^## P절 재조정/' "$V" | grep -c '_미측정'
grep -E '^missing_init = ' "$V"

echo "### AC-ALH-003"
for s in init live; do
  root="$SCRATCH/init-surface"; [ "$s" = live ] && root="$SCRATCH/live"
  python3 .moai/reports/t1226/candidates.py --keys --paths ".moai/reports/t1226/count-set-$s.txt" "$root" | sort > "$SCRATCH/k-$s"
  tail -n +2 ".moai/reports/t1226/candidates-$s.tsv" | cut -f1,2,3,6 | sort | diff - "$SCRATCH/k-$s" && echo "KEYS-$s-OK"
done
for s in init live; do
  root="$SCRATCH/init-surface"; [ "$s" = live ] && root="$SCRATCH/live"
  awk -F'\t' 'NR>1 && $2 !~ / ¶[0-9]+$/ {g[$1]+=$3} END {for (f in g) print f"\t"g[f]}' ".moai/reports/t1226/candidates-$s.tsv" | sort > "$SCRATCH/g-$s"
  (cd "$root" && xargs wc -m < "$OLDPWD/.moai/reports/t1226/count-set-$s.txt" | grep -v ' total$' | awk '{print $2"\t"$1}' | sort) > "$SCRATCH/w-$s"
  diff "$SCRATCH/g-$s" "$SCRATCH/w-$s" > /dev/null && echo "SPLIT-$s-OK"
done
for s in init live; do
  f=".moai/reports/t1226/candidates-$s.tsv"
  awk -F'\t' 'NR>1 {
    if (NF!=15) bad++;
    if ($13!="ADMIT" && $13!="REJECT" && $13!="UNTRIED") bad++;
    if ($13=="ADMIT" && ($6+0>0 || $7!="N" || $8!="Y" || $14!="-" || $4 !~ /^[0-9]+$/)) bad++;
    if ($13=="ADMIT" && $5=="M1" && ($9!="Y"||$10!="Y"||$11!="Y"||$12=="-")) bad++;
    if ($1 ~ /AGENTS\.md/ && $5!="M2" && $13=="ADMIT") bad++;
    if ($13!="ADMIT" && $4!="-") bad++;
    if ($13=="UNTRIED" && ($5!="M2" || $14 !~ /^untried:.+/)) bad++;
    if ($13=="REJECT") {
      ok=0;
      if ($14=="bind>0" && $6+0>0) ok=1;
      if ($14=="governs-scope" && $7=="a") ok=1;
      if ($14=="narrows-scope" && $7=="b") ok=1;
      if ($14=="citation-breaks" && $9=="N") ok=1;
      if ($14=="dest-unreachable" && $10=="N") ok=1;
      if ($14=="dest-over-40k" && $11=="N") ok=1;
      if ($14=="agents-md-no-m1" && $1 ~ /AGENTS\.md/ && $5!="M2") ok=1;
      if ($14=="config-data" && $1 ~ /\.yaml$/) ok=1;
      if ($14=="net-negative" && $5=="M1" && $6+0==0 && $7=="N" && $8=="Y" && $9=="Y" && $10=="Y" && $11=="Y" && $12!="-") ok=1;
      if ($14=="rewraps-binding-line" && $5=="M2") ok=1;
      if (!ok) bad++;
    }
    if ($3 !~ /^[0-9]+$/ || $15=="") bad++;
    if ($13=="ADMIT" && $4+0 > $3+0) bad++;
    if ($13=="ADMIT" && $5 ~ /^M1/ && $4+0 != $3+0) bad++;
  } END {print "ROWS-'"$s"' BAD="bad+0}' "$f"
  awk -F'\t' 'NR>1 && $13=="ADMIT" {b=$2; isp=sub(/ ¶[0-9]+$/,"",b); k=$1 SUBSEP b; if (isp) p[k]=1; else s[k]=1} END {for (k in s) if (k in p) bad++; print "OVERLAP-'"$s"' BAD="bad+0}' "$f"
  awk -F'\t' 'NR>1 && $13=="ADMIT" && $5=="M1" {print ".claude/rules/moai/"$12}' "$f" | sort -u \
    | grep -x -F -f ".moai/reports/t1226/count-set-$s.txt" | wc -l | awk '{print "DESTIN-'"$s"' BAD="$1}'
  awk -F'\t' 'NR>1 && $13=="ADMIT" && $5=="M1p" {print $15}' "$f" | while read -r p; do grep -q -E '^dup_source = .+' "$p" || echo "m1p-bad $p"; done | wc -l | awk '{print "M1P-'"$s"' BAD="$1}'
  awk -F'\t' 'NR>1 && $14=="net-negative" {print $3"\t"$15}' "$f" | while IFS="$(printf '\t')" read -r g p; do
    P=$(grep -E '^pointer_chars = [0-9]+$' "$p" | awk '{print $3}'); [ -n "$P" ] && [ "$P" -ge "$g" ] || echo "netneg-bad $p"
  done | wc -l | awk '{print "NETNEG-'"$s"' BAD="$1}'
  tail -n +2 "$f" | cut -f15 | sort -u | while read -r p; do [ -f "$p" ] || echo "missing $p"; done | wc -l | awk '{print "EVID-'"$s"' MISSING="$1}'
done
for s in init live; do
  awk -F'\t' 'FNR==NR {live[$1]=$2; tmpl[$1]=$3; next}
    FNR>1 && $13=="ADMIT" && $5=="M1" {add[$12]+=$4}
    END {for (d in add) { if (!(d in live) || live[d]+add[d]>=40000 || tmpl[d]+add[d]>=40000) bad++ } print "DEST-'"$s"' BAD="bad+0}' \
    .moai/reports/t1226/dest-sizes.tsv ".moai/reports/t1226/candidates-$s.tsv"
done

echo "### AC-ALH-004"
for s in init live; do
  root="$SCRATCH/init-surface"; [ "$s" = live ] && root="$SCRATCH/live"
  (cd "$root" && grep '\.md$' "$OLDPWD/.moai/reports/t1226/count-set-$s.txt" | xargs grep -hE '\[HARD\]|MUST|shall ' \
     | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | sort | shasum -a 256 | awk '{print $1}')
done
grep -c -E '^hash_(init|live) = [0-9a-f]{64}$' "$V"
for s in init live; do
  H=$(grep -E "^hash_${s} = " "$V" | awk '{print $3}')
  awk -F'\t' 'NR>1 && $5=="M2" && ($13=="ADMIT" || $14=="rewraps-binding-line") {print $13"\t"$15"\t"$3"\t"$4}' \
    ".moai/reports/t1226/candidates-${s}.tsv" \
  | while IFS="$(printf '\t')" read -r v p g c; do
      got=$(grep -E '^post_hash = ' "$p" | awk '{print $3}')
      if [ "$v" = "ADMIT" ] && [ "$got" != "$H" ]; then echo "bad $p"; fi
      if [ "$v" != "ADMIT" ] && [ "$got" = "$H" ]; then echo "bad $p"; fi
      if [ "$v" = "ADMIT" ]; then
        pre=$(grep -E '^pre_chars = [0-9]+$' "$p" | awk '{print $3}'); post=$(grep -E '^post_chars = [0-9]+$' "$p" | awk '{print $3}')
        if [ -z "$pre" ] || [ -z "$post" ] || [ "$pre" != "$g" ] || [ "$c" != "$((pre - post))" ]; then echo "chars-bad $p"; fi
      fi
    done
done | wc -l | awk '{print "HASH-BAD="$1}'

echo "### AC-ALH-005"
shasum -a 256 .moai/reports/t1226/sec.py | awk '{print $1}'
grep -c -E '^P절 재조정: \((a|b|공표값|기타)\)$' "$V"
grep -c -E '^J_includes_kanban_scope = (yes|no)$' "$V"
grep -c -E '^F\(172ef22eb\) = [0-9]+$' "$V"
ls .moai/reports/t1226/sec-172ef22eb.txt
cut -f15 .moai/reports/t1226/candidates-init.tsv .moai/reports/t1226/candidates-live.tsv | grep -c -E '^(/tmp|/private/tmp)'
grep -c -E '^evidence: *(/tmp|/private/tmp)' "$V"

echo "### AC-ALH-006"
SR=.claude/rules/moai/workflow/skill-routing.md
calc() {
  s=$1; x=$2; sfx=$3
  A=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x && $13=="ADMIT" {s+=$4} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  U=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x && $13=="UNTRIED" {s+=$3} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  R=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x {s+=$3} END {print s+0}' ".moai/reports/t1226/pointers-${s}.tsv")
  C=$(awk -F'\t' -v x="$x" 'NR>1 && $1!=x && $2 !~ / ¶[0-9]+$/ {s+=$3} END {print s+0}' ".moai/reports/t1226/candidates-${s}.tsv")
  awk -v k="$sfx" -v A="$A" -v U="$U" -v R="$R" -v C="$C" '
    $1=="total_"k {tt=$3} $1=="current_"k {c=$3} $1=="A_adm_"k {a=$3} $1=="U_"k {u=$3} $1=="R_"k {r=$3} $1=="T_min_"k {t=$3}
    END { if (tt==C && c==C && a==A && u==U && r==R && t==C-A+R) print "CHECK-"k"=OK"; else print "CHECK-"k"=BAD" }' "$V"
}
calc init "" init
calc live "" live
grep -q -E '^T_min_init_17 = ' "$V" && calc init "$SR" init_17
for s in init live; do
  awk -F'\t' 'NR>1 && $13=="ADMIT" && $5=="M1" {print $1"\t"$12}' ".moai/reports/t1226/candidates-${s}.tsv" | sort -u > "$SCRATCH/pair-$s"
  tail -n +2 ".moai/reports/t1226/pointers-${s}.tsv" | cut -f1,2 | sort > "$SCRATCH/ptr-$s"
  diff "$SCRATCH/pair-$s" "$SCRATCH/ptr-$s" && echo "PAIRS-$s=OK"
done
echo "# review: pointer_chars == wc -m of committed pointer text"
for s in init live; do
  for p in .moai/reports/t1226/evidence/pointers/$s/*.txt; do printf '%s %s\n' "$(wc -m < "$p" | tr -d ' ')" "$p"; done
  tail -n +2 .moai/reports/t1226/pointers-$s.tsv | cut -f3 | sort -n | tr '\n' ' '; echo
done

echo "### AC-ALH-007"
rule() { awk -v k="$1" '
    $1=="T_min_"k {t=$3} $1=="U_"k {u=$3} $1=="verdict_"k {v=$3; n++}
    END { if (t<150000) e="ACHIEVABLE"; else if (t-u>=150000) e="STRUCTURALLY-INFEASIBLE-UNDER-FREEZE"; else e="UNDETERMINED";
          print e"\t"v"\t"n }' "$V"; }
for k in live init_17; do
  grep -q -E "^T_min_${k} = " "$V" || continue
  rule "$k" | awk -F'\t' -v k="$k" '{print ($3==1 && $1==$2) ? "TOKEN-"k"=OK" : "TOKEN-"k"=BAD" }'
done
E18=$(rule init | cut -f1); VI=$(rule init | cut -f2); N=$(rule init | cut -f3)
V17=$(grep -E '^verdict_init_17 = ' "$V" | awk '{print $3}')
if [ -n "$V17" ] && [ "$V17" != "$E18" ]; then
  [ "$N" = 1 ] && [ "$VI" = UNDETERMINED ] && grep -q -E '^verdict_init_reason = count-set$' "$V" && echo TOKEN-init=OK || echo TOKEN-init=BAD
else
  [ "$N" = 1 ] && [ "$VI" = "$E18" ] && echo TOKEN-init=OK || echo TOKEN-init=BAD
fi
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '^### \((a|b|c|d)\)'
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '^RECOMMEND: '
awk '/^## 동결 해제 상신 절차/,/^## Gaps/' "$V" | grep -c -E '결정했다|승인했다|해제한다\.?$'

echo "### AC-ALH-009 (non-git part)"
H=.moai/reports/t1226/harness-run.txt
grep -c -- '--- PASS: TestHeadroomInitSurfaceExport ' "$H"
grep -c -- '--- SKIP: TestHeadroomInitSurfaceExport ' "$H"
head -1 "$H" | awk '{print $3}'
grep -E '^build_head = ' "$V"
grep -c -E '^harness_sha256 = [0-9a-f]{64}$' "$V"
grep -E '^harness_sha256 = ' "$V"
test -s .moai/reports/t1226/commands.log && echo LOG-OK
grep -c -E 'moai([[:space:]]+-[^[:space:]]+)*[[:space:]]+init([[:space:]]|$)' .moai/reports/t1226/commands.log

echo "### AC-ALH-010"
grep -c -E '^runtime_files_init = [0-9]+$' "$V"; grep -c -E '^runtime_total_init = .+' "$V"
grep -c -E '^runtime_isolated = yes$' "$V"; grep -c -E '^runtime_source = .+' "$V"
grep -c -E '^count_set_init = observed$' "$V"
grep -c -E '^runtime_observed = no$' "$V"
grep -c -E '^count_set_init = 18$' "$V"
grep -c -E '^(total|current|A_adm|R|U|T_min)_init_17 = [0-9]+$' "$V"
grep -c -E '^verdict_init_17 = (ACHIEVABLE|STRUCTURALLY-INFEASIBLE-UNDER-FREEZE|UNDETERMINED)$' "$V"
