# Internal: versioning

When to read: before you ask the `versioning` rows of the question bank (`internal/interview`) in the PROJECT route, and before you compute or check a release version for a project whose profile has a `versioning` key.

MoAI offers three schemes (card C-69): SemVer is the default, HeadVer the second, CalVer the third. Anything else is the user's own words, recorded as `custom`. The choice is the profile key `versioning`: `scheme` is `semver`, `headver`, `calver` or `custom`; `format` is the CalVer format, or for `custom` the user's words in one line (at most 128 characters; no slash, `<`, `>`, backtick, leading `#` or invisible character); `head` is the HeadVer head number, at most 9999. `project-plan` refuses any other scheme, a `format` on SemVer or HeadVer, a `head` on anything but HeadVer, and a `versioning` that is not an object. A profile without the key has no recorded choice: ask, never assume SemVer. The `AGENTS.md` managed section carries the choice as one line, `Versioning: …`.

## Keep the existing scheme

`moai project-detect` reports `versioning` beside `detected`, read only: `scheme` (`semver`, `calver`, `headver`, `mixed` or `none`), `latest` (the highest version it recognised, as written), `evidence` (the five highest tags, then the version in the root `Cargo.toml`, `package.json` and `pyproject.toml`, each with the `shape` it was read as) and `tags` (`read`, `no_repository`, or `unavailable` with a `tags_reason`). "Highest" is version order, not date: numbers compared segment by segment, a release above its own pre-release. Tag names are read from the files of the project's own `.git` (`refs/tags` and `packed-refs`; a linked worktree's pointer is followed, nothing else is); no program is started, so a name like `v1.9.9` tagged after `v2.0.0` does not become `latest`.

1. When `scheme` names one scheme, keeping it is the first and recommended option, shown with that scheme and `latest` so the user can correct a wrong reading. Never recommend a change the user did not ask for: published versions, and every tool that sorts them, depend on the scheme.
2. Tags decide. The manifests decide only when no tag has a shape.
3. Shapes overlap, and the evidence rows say where. A row with `shape` `ambiguous` (`26.4.1` is valid SemVer and a possible CalVer; `16.4` is a possible CalVer or a plain two-part version) is not counted: with `none` and such a row, show the row in the question; there is no keep option to offer. Every HeadVer version is also valid SemVer, and its row carries the note "also valid SemVer": HeadVer shapes beside SemVer tags are counted as SemVer (a minor that walks through `1.999.0`, `1.1000.0`, `1.1001.0`), and HeadVer shapes alone are reported as `headver`, which a SemVer project whose minor happens to look like a year and week would also show. The user's pick decides, not the detection.
4. `mixed` means the tags show two schemes: say which, then ask without a keep option.
5. `tags` `unavailable` or `no_repository` means no tag was read: `none` then says nothing about the project's history; say so when you ask, with the `tags_reason`.
6. A change of scheme must still sort above every published version. SemVer to HeadVer: `head` is the old MAJOR plus one. SemVer to CalVer: the year segment must be greater than the old MAJOR. From CalVer with a full year to SemVer or HeadVer no smaller first number sorts above the old versions: say so before the user picks.

How `project-detect` reads one version, after dropping a leading `v` and a prefix that ends in `/` or `@`, and ignoring what follows `-` or `+`:

| Shape | Rule |
| --- | --- |
| `calver` | Two to four numeric segments; the first is a four-digit year from 2000, or a two-digit number from 10 followed by a month from 1 to 12 (with only two segments the month must be zero-padded: `26.04`) |
| `headver` | Three numeric segments; the middle one has four digits, a two-digit year from 10 and a week from 01 to 53 |
| `semver` | Three numeric segments without leading zeros that fit neither rule above |
| `ambiguous` | Fits two readings: a two-digit year and a month that is also valid SemVer (`26.4.1`), two segments without a padded month (`16.4`, `16.10`), or a HeadVer middle segment behind a four-digit year |
| `unrecognised` | Anything else |

## SemVer

- Format: `MAJOR.MINOR.PATCH`, three non-negative integers without leading zeros. A tag usually carries a `v` in front (`v3.2.1`); the `v` is not part of the version.
- Who bumps what: a person decides from what the change does to users of the public interface. MAJOR for an incompatible change, MINOR for a feature added in a backward-compatible manner, PATCH for a backward-compatible fix. `0.y.z` is initial development, where anything may change.
- Pick it when: other code depends on the project's interface (a library, an API, a command line that scripts call) and a reader must know whether an upgrade can break them.
- Next version: from the newest release and the changes since it. Any incompatible change gives MAJOR plus one with MINOR and PATCH 0; else any feature gives MINOR plus one with PATCH 0; else PATCH plus one. The call is a human judgement; automation that reads commit messages (semantic-release style: `fix:` is PATCH, `feat:` is MINOR, a `BREAKING CHANGE` footer is MAJOR) makes it consistent, and only for a project that already writes such messages.
- Pre-release: a hyphen and dot-separated identifiers (`1.0.0-rc.1`), which sort before the release. Build metadata follows `+` (`1.0.0+build.7`) and is ignored in ordering.
- Examples: `v3.2.1` and a fix give `v3.2.2`; `v3.2.2` and an incompatible change give `v4.0.0`.

## HeadVer

- Format: `{head}.{yearweek}.{build}`. `yearweek` always has four digits: the two-digit year, then the two-digit ISO-8601 week number (week 34 of 2021 is `2134`).
- Who bumps what: `head` is the only number a person sets; it is zero-based and raised when the team decides a release should be noticed. `yearweek` comes from the build date. `build` is an incremental number from the build server.
- Pick it when: an app or a service ships on a steady rhythm from a build server and nobody depends on its interface by version, so the version should say when it was built and which build it is.
- Next version: keep `head` unless a person raised it; take `yearweek` from the build day; take `build` from the build server's counter. Without such a counter the next version cannot be computed: say so and never invent the number.
- Pre-release: the scheme has no marking of its own. A hyphen suffix is allowed and advised against (raise `head` instead); a `+meta` suffix is acceptable.
- Examples: `3.2647.59` is head 3, week 47 of 2026, build 59; the next build one week later is `3.2648.60`. A project that moves from `v2.1.0` starts at head 3.

## CalVer

- Format: the project states its own, built from `YYYY` (full year), `YY` (short year: 6, 16, 106), `0Y` (zero-padded year), `M` and `0M` (month), `W` and `0W` (week of the year), `D` and `0D` (day), and `MAJOR`, `MINOR`, `MICRO`, `MODIFIER`. MoAI's default is `YYYY.0M.MICRO`. The profile takes two to four of these tokens joined by dots, none twice, the first one the only year token (`MM`, `WW` and `DD`, the spelling before spec 26.0, are still read); any other form is recorded as `custom`.
- Who bumps what: the calendar sets the date segments on the release day. `MICRO` counts the releases inside one period from 0 and is the only number a person or a script increments.
- Pick it when: time or outside change drives the releases more than interface changes do (a large or constantly changing scope, a time-sensitive project). It does not signal a breaking change: pair it with a changelog and a deprecation policy.
- Next version: take the date segments from the release day. When they equal the newest release's, `MICRO` plus one; otherwise `MICRO` 0. A format that ends in the day (`YYYY.0M.0D`) appends a micro segment for a second release on the same day.
- Pre-release: the `MODIFIER`, written the way the project's tags already write it; with nothing to follow, `-rc.1` (`2026.11.0-rc.1`).
- Examples: with `YYYY.0M.MICRO`, `2026.11.0` and a second release in November give `2026.11.1`, and the first one in December is `2026.12.0`. Elsewhere: Ubuntu `YY.0M` (`26.04`), yt-dlp `YYYY.0M.0D` (`2026.03.17`), pip `YY.MINOR.MICRO`.

Sources, read 2026-11-17: https://semver.org (2.0.0), https://github.com/line/headver, https://calver.org (spec 26.0).
