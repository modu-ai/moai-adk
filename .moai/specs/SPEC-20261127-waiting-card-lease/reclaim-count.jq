# Reclaim events per UTC day, split by what the card was waiting for, from `moai card export`.
# moai card export --out <export>
# grep -h '^{' <export>/C-*.md > <cards>.jsonl
# jq -nc -f reclaim-count.jq --slurpfile ev <export>/events.jsonl --slurpfile cards <cards>.jsonl
($cards | [.[].notes[] | {key: (.note_id | tostring), value: (.body | split("\n")[0])}] | from_entries) as $first
| ($ev | map(select(.kind == "card.note"))
    | group_by(.card_id)
    | map({key: (.[0].card_id | tostring), value: map({id: .event_id, at: .at, note: (.payload | fromjson | .note_id)})})
    | from_entries) as $notes
| [ $ev[]
    | select(.kind == "card.reclaim")
    | . as $r
    | (.payload | fromjson) as $p
    # the newest note when the old lease ended says what the holder waited for
    | ([ ($notes[$r.card_id | tostring] // [])[] | select(.id < $r.event_id and .at <= $p.from_lease_until) ] | max_by(.id) | .note) as $n
    | {
        day: (.at | strftime("%Y-%m-%d")),
        card: $r.card_id,
        review: ($p.card_state == "review"),
        blocked: ($p.card_state != "review" and $n != null and (($first[$n | tostring] // "") | startswith("LANE-BLOCKED")))
      }
  ]
| group_by(.day)
| map({
    day: .[0].day,
    reclaims: length,
    review: (map(select(.review)) | length),
    blocked: (map(select(.blocked)) | length)
  })
| .[]
