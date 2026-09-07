Thank you for the report — and in particular for flagging the data-loss risk. This is
confirmed as a defect and it is fixed on the development line (merge `a9d2fb641`).
The release carrying this fix has not been cut yet; we'll follow up here as soon as
it ships.

The root cause was exactly what you described: the `--sync-git` write path never
consulted `--dry-run`, so a status refresh that was supposed to be a preview went
ahead and rewrote `spec.md`. After the fix, `--dry-run` prints the full summary of
what *would* change and writes nothing — verified by re-running the same operations
with the fixture files' content hashes compared byte-for-byte before and after
(identical) and an empty `git status` in the fixture repository. The real (non-dry)
run still updates the frontmatter as before, and a non-recognizable parsed status is
now skipped loudly instead of silently.

We'll keep this issue open until you've been able to confirm the fix in a release;
we'll post here as soon as one ships.
