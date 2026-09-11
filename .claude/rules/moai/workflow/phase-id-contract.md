# Stable Phase ID Contract

Phase routing uses stable string IDs, not display numbers. A phase record has
`phase_id`, `status`, `input_key`, and `next_ids`. Decimal labels such as
`2.75` are display aliases only and MUST map to one canonical ID in the
resolver table.

- `skip` accepts canonical IDs and records the skipped ID plus reason.
- `resume` compares completed canonical IDs against the DAG and never guesses
  from the last printed number.
- A missing target, duplicate ID, or cycle is a blocker before execution.
- Deprecated display labels remain in an explicit alias map until removed; they
  are never interpreted as a new phase.
