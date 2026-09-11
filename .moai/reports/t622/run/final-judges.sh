#!/bin/sh
# Final re-judge of AC-GDP-006/026/027/028/029 on the M1-M3 tree (card t622).
# Read-only. Run from .moai/reports/t622/run after the final-* fragments were extracted.
# Detector strings are copied verbatim from acceptance.md.
for side in local template; do
  p=final-$side
  d=$p-dl.md
  m=$p-mg.md
  echo "== $side"
  awk '/default for worktree contexts|worktree contexts default|no-merge.{1,3}flag NOT set|merges? (automatically|by default)|auto-merge (is )?(the )?default|no-merge.{0,12}(absent|not set|missing|NOT set)/ {print FNR}' $d > $p-006-dl-default.txt
  test -e $p-006-dl-default.txt; e1=$?; test -s $p-006-dl-default.txt; s1=$?
  echo "006 dl-default test-e=$e1 test-s=$s1"
  /usr/bin/grep -n -i -E 'default (to )?auto-merge|worktree contexts default|merges? (automatically|by default)' $p-de.md > $p-006-de-default.txt
  echo "006 de-default grep-exit=$?"
  /usr/bin/grep -c 'manager-[g]it[.]md' $d $p-de.md > $p-006-source.txt
  echo "006 source: $(tr '\n' ' ' < $p-006-source.txt)"
  /usr/bin/grep -c -e '--auto-merge' $p-skill.md $p-ref.md $p-sync.md $d $p-sync-usage.md $p-hint.md $p-dl-next.md > $p-026-i.txt
  /usr/bin/grep -c -e '--auto-merge' $p-qgc-args.md $p-qgc-flags.md > $p-026-i-qgc.txt
  echo "026 (i): $(tr '\n' ' ' < $p-026-i.txt) | qgc: $(tr '\n' ' ' < $p-026-i-qgc.txt)"
  awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ && (!/--merge([^A-Za-z-].*)?[Dd]eprecated alias (of|for) .?--auto-merge([^A-Za-z-]|$)/ || /(not|no longer|never) (a |an |the )?[Dd]eprecated/ || /[Uu]n-?deprecat/) {print FILENAME ":" FNR ": " $0}' $p-skill.md $p-ref.md $p-qgc-args.md $p-qgc-flags.md $p-sync.md $d $p-sync-usage.md $p-hint.md $p-dl-next.md > $p-026-ii.txt
  test -e $p-026-ii.txt; e2=$?; test -s $p-026-ii.txt; s2=$?
  echo "026 (ii) test-e=$e2 test-s=$s2"
  /usr/bin/grep -c -E '(^|[^A-Za-z-])--merge([^A-Za-z-]|$)' $p-skill.md $p-sync.md $p-qgc-flags.md $d > $p-026-merge-presence.txt
  echo "026 (iii): $(tr '\n' ' ' < $p-026-merge-presence.txt)"
  awk '/(^|[^A-Za-z-])--merge([^A-Za-z-]|$)/ {print FILENAME ":" FNR ": " $0}' $p-skill.md $p-ref.md $p-qgc-args.md $p-qgc-flags.md $p-sync.md $d $p-sync-usage.md $p-hint.md $p-dl-next.md > $p-026-merge-lines.txt
  echo "026 merge-lines: $(/usr/bin/grep -c '' $p-026-merge-lines.txt)"
  /usr/bin/grep -c -e '--no-merge' $d > $p-027-dl-count.txt
  awk '/--no-merge/ && !(/no-op/ && /[Dd]eprecat/) {print FILENAME ":" FNR ": " $0}' $p-skill.md $p-ref.md $p-qgc-args.md $p-qgc-flags.md $p-sync.md $d $p-sync-usage.md $p-hint.md $p-dl-next.md > $p-027-i.txt
  test -e $p-027-i.txt; e3=$?; test -s $p-027-i.txt; s3=$?
  awk '/--no-merge/ && (/[Ss]kip/ || /[Nn][Oo][Tt] set/ || /[Pp]revent/ || /unless/ || /[Dd]isabl/ || /[Oo]verrid/ || /[Tt]urns? off/ || /[Ss]uppress/ || /[Bb]ypass/ || /[Cc]ancel/) {print FILENAME ":" FNR ": " $0}' $p-skill.md $p-ref.md $p-qgc-args.md $p-qgc-flags.md $p-sync.md $d $p-sync-usage.md $p-hint.md $p-dl-next.md > $p-027-ii.txt
  test -e $p-027-ii.txt; e4=$?; test -s $p-027-ii.txt; s4=$?
  echo "027 dl-count=$(cat $p-027-dl-count.txt) (i) test-e=$e3 test-s=$s3 (ii) test-e=$e4 test-s=$s4"
  awk '/[Tt]eam mode/ && /--auto-merge/ && /(^|[^A-Za-z])[Aa]ll( [A-Za-z-]+)?( [A-Za-z-]+)? approv/ {c++} END{print c+0}' $m > $p-028-a-mg.txt
  awk '/[Tt]eam mode/ && /--auto-merge/ && /(^|[^A-Za-z])[Aa]ll( [A-Za-z-]+)?( [A-Za-z-]+)? approv/ {c++} END{print c+0}' $d > $p-028-a-dl.txt
  awk '/[Tt]eam mode/ && (/without (any |an )?approv/ || /no approv/ || /approv[a-z]* (are |is )?(not required|not needed|optional|unnecessary)/ || /optional approv/ || /regardless of approv/ || /at least one approv/ || /(a|one) single approv/ || /(after|on|once|upon|with) (any|one|an|a|some|the first) approv/ || /majority/) {print FILENAME ":" FNR ": " $0}' $m $d > $p-028-b.txt
  test -e $p-028-b.txt; e5=$?; test -s $p-028-b.txt; s5=$?
  echo "028 (a) mg=$(cat $p-028-a-mg.txt) dl=$(cat $p-028-a-dl.txt) (b) test-e=$e5 test-s=$s5"
  awk '/[Pp]ersonal/ && /[Mm]anual/ && /--auto-merge/ {c++} END{print c+0}' $m > $p-029-a-mg.txt
  awk '/[Pp]ersonal/ && /[Mm]anual/ && /--auto-merge/ {c++} END{print c+0}' $d > $p-029-a-dl.txt
  awk '(/[Pp]ersonal/ || /[Mm]anual/) && (((/[Aa]pprov/ || /[Rr]eview/) && !(/without (requiring )?(any |an )?(approv|review)/ || /no (approv|review)/ || /not required/ || /not needed/ || /no teammates/)) || /(after|once|until|pending|requires?) (a |an |the |all |one |at least one )?(code )?(approv|review)/) {print FILENAME ":" FNR ": " $0}' $m $d > $p-029-b.txt
  test -e $p-029-b.txt; e6=$?; test -s $p-029-b.txt; s6=$?
  echo "029 (a) mg=$(cat $p-029-a-mg.txt) dl=$(cat $p-029-a-dl.txt) (b) test-e=$e6 test-s=$s6"
  awk '/[Aa]pprov/ || /[Rr]eview/ || /[Tt]eam mode/ || /[Pp]ersonal/ || /[Mm]anual/ || /--auto-merge/ {print FILENAME ":" FNR ": " $0}' $m $d > $p-028-mode-lines.txt
  echo "028 mode-lines: $(/usr/bin/grep -c '' $p-028-mode-lines.txt)"
done
