# Design DNA Profiles

A completed Design DNA JSON can be saved under a name and reused, so a later
generation starts from the saved profile instead of re-extracting the same
reference. Without this, every session that wants "the checkout design" re-reads
the reference and re-derives a slightly different profile — the same drift this
skill's intermediate-JSON step exists to remove, moved from one session to the
next.

> **Provenance**: the marker-first profile mechanism is adapted from
> `cathrynlavery/diagram-design` v2.6.1 (MIT) — restated, not copied. No
> upstream code or verbatim text is included.

## Where profiles live

Profiles live in a project-scoped directory at the project root, outside this
skill's own directory:

```
<project-root>/.design-dna/
  profiles/<slug>.json     # saved Design DNA snapshots
  active                   # the marker: one line naming the selected slug
```

The location is the load-bearing decision, and three alternatives were rejected
for it:

- **Inside the skill directory** — dies on skill update; a skill reinstall
  silently deletes every saved profile.
- **Home-global only** (a dot-directory in the user's home) — invisible to the
  repository; a teammate cloning the project gets none of the profiles, and two
  checkouts of the same project see different design data.
- **Token-merge into a central config** — the copy-race shape: two sessions
  writing one merged file clobber each other.

A project-root directory versions with the project, survives skill
reinstallation, and needs no tooling beyond Read and Write. When a profile is
loaded, the snapshot is **read in place** — the skill never copies it into its
own storage, so there is exactly one copy and no copy race.

## The marker — marker-first resolution

`.design-dna/active` is a one-line file whose single line is a slug:

```
checkout-refresh
```

**Where a project carries the marker, resolution is marker-first: the marker's
slug wins over any other lookup**, and the referenced snapshot
(`profiles/<slug>.json`) is read in place. **Where no marker exists,
no profile is guessed** — extraction proceeds profile-less, exactly as it
would with no profile mechanism at all. A stale marker naming a missing snapshot is an error
to report, never a cue to pick the nearest-sounding profile.

Switching the active profile is editing one line; nothing else moves.

## Slug grammar

A slug is lowercase `a-z`, digits `0-9`, and single hyphens; it starts and ends
with a letter or digit. No spaces, no underscores, no slashes — a slug is a
name, not a path — and no task or ticket identifiers: a name a person would
recognize next month (`checkout-refresh`, `brand-2026-refresh` if the team
means the year) outlives `TASK-482-final`.

## Snapshot shape — metadata header

Each snapshot is a Design DNA JSON (the three dimensions of
`references/dna-schema.md`) wrapped in a small `meta` header:

```json
{
  "meta": {
    "slug": "checkout-refresh",
    "origin_reference": "https://example.com/checkout — 3 mobile screenshots, 1 desktop",
    "schema_version": 1,
    "extracted_at": "2026-05-14"
  },
  "design_system": { "...": "per dna-schema.md" },
  "design_style": { "...": "per dna-schema.md" },
  "visual_effects": { "...": "per dna-schema.md" }
}
```

`origin_reference` names what was extracted (URL, screenshot set, video) so a
later reader can re-check the profile against its source. `schema_version`
tracks the three-dimension schema shape (`1` while `dna-schema.md` carries the
current field set). `extracted_at` is the project's own data — a date a user
writes into their artifact — which is why an example date may appear here; the
distinction is between user-generated profile data (theirs) and template-file
chronology (absent by contract).

## Saving a profile (Phase 2 hook)

When Phase 2 completes and the extracted JSON is confirmed:

1. Offer to save it under a slug. Derive a suggested slug from the reference's
   product or team name; the caller confirms or renames it.
2. **Confirm before overwrite.** When the slug already exists under
   `profiles/`, ask before overwriting — an existing snapshot may be the active
   profile of a teammate's session.
3. **Verify by re-reading.** After the write completes, read the saved file
   back and compare it against the JSON that was to be saved. A mismatch is
   reported as a failed save; the save is never assumed from the write alone.
4. Offer to make the new slug active (write the marker line) — active is the
   caller's call, not a side effect of saving.

## Loading a profile (Phase 3 start)

When the marker names a slug and a new generation is requested, start from that
snapshot instead of re-extracting:

1. Read `profiles/<slug>.json` **in place**.
2. **Validate against the three-dimension schema** of `references/dna-schema.md`
   before use.
3. **Backfill missing optional fields with an explicit `"not observed"` value** —
   the same rule Phase 2 applies at extraction time. A dimension missing
   entirely is backfilled as `"not observed"` at the dimension level; a
   dimension is never silently dropped, and a value the snapshot does not carry
   is never invented. A profile that guesses is worse than no profile, because
   it launders the guess through the one artifact meant to be guess-free.
4. Phase 3 then runs unchanged, from the loaded profile as if it were the
   Phase 2 output.

## Boundaries

- One marker per project. Profiles coexist under `profiles/`; `active` names
  the one in play.
- Profiles are project data: commit them or not as the team decides, but the
  directory is ordinary repository content — nothing here writes outside
  `<project-root>/.design-dna/`.
- The analyze → generate flow is unchanged when no profile exists; this
  mechanism adds a save step and a start-from step, never a routing change.
