# models-revert

`moai models-revert <HOME> --receipt models-… --permission-reference <ref>` puts the launcher
settings file back as the receipt of one `moai update` model save kept it: the text before the
save, byte for byte, or no file when there was none. This happens only while the file is still
the one that save wrote (its hash equals the receipt's `after_sha256`); a file changed since is
left alone and named.

## Order and the resume rule

The revert marker (`<receipt id>.reverted` in the receipts folder) is written **before** the
settings file is touched (SPEC-20261119-models-revert-reporting REQ-1). A failed marker write
refuses before anything changed — the message says `nothing was changed`, the kind is `io`
(retryable), and the same command re-runs whole. A run that dies after the marker but before the
revert leaves the marker with the file still at the saved state: the next attempt sees marker +
matching hash and **resumes** the revert (REQ-2). A marker whose file no longer matches refuses
`already reverted`; a plain changed-since file (no marker) answers `skipped` with
`changed_since_apply` and writes no marker.

## Refusals, kinds and exit codes

Every refusal exits 1 with the error envelope; `retryable` is true only for the `io` marker
failure. The kind comes from the message table (`src/contract.rs`), not a per-op mapping.

| refusal message | kind | exit | retryable |
| --- | --- | --- | --- |
| `permission reference required` | `permission_reference_missing` | 1 | no |
| `invalid models receipt id` | `invalid_input` | 1 | no |
| `models receipt not found` | `not_found` | 1 | no |
| `conflict in models revert: the receipt is not a models receipt` | `conflict` | 1 | no |
| `conflict in models revert: the receipt names another settings file` | `conflict` | 1 | no |
| `invalid models receipt: the fields the revert needs are missing or do not hash` | `invalid_input` | 1 | no |
| `models receipt already reverted` | `conflict` | 1 | no |
| `conflict in models revert: the settings file read back differently after the revert; check the file` | `conflict` | 1 | no |
| `models revert marker write failed: …; nothing was changed` | `io` | 1 | **yes** |

## Success

`{"state":"reverted","receipt":…,"path":…}` with the marker on file; a changed-since file answers
`{"state":"skipped","reason":"changed_since_apply",…}`. The receipt id a save printed under
`models.receipt` is the `--receipt` argument; the permission reference is recorded in the marker
and checked by `permission_reference_ok`.
