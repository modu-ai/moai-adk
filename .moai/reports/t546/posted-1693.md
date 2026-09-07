Also confirmed and fixed on the same line (merge `a9d2fb641`) — these two reports described two
sides of one locator defect.

The status locator used to scan the document body, where prose or table cells that
merely mention `Status` / `Notes` could be picked up as if they were the SPEC's
status, producing the false drift you saw (including matches inside backticked
prose). The locator is now anchored to the YAML frontmatter `status:` field only:
body cells — both table-row and backticked-prose variants — are never read as the
status, verified with mutation tests that confirm a body `Status:` cell is ignored
while the real frontmatter status still drives detection and updates.

One note for transparency: drift *summaries* can still legitimately report real
mismatches between the frontmatter status and the lifecycle record — that behavior
is unchanged by design. What's gone is the class of false positives you reported.
We'll post here as soon as a release with this fix ships.
