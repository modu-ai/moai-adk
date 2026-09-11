# AC-GDP-015 reading record (M4, run-phase part 2)

Read by: manager-develop (run-phase part 2, card t622). Tree: HEAD `5708e04d2` (regenerated
`.toml` committed), CARD_BASE `f1f034bb43b06dbde7f7a93c1c51edc5b4d8f5cf`.

Source: `ac015-template.diff` = `git diff develop...HEAD -- internal/template/templates/`
(`test -e` 0, `test -s` 0). Files in the diff (`ac015-files.txt`, 9): template `manager-git.md`,
`sync.md.tmpl`, `moai/SKILL.md`, `references/reference.md`, `workflows/sync.md`, `delivery.md`,
`doc-execution.md`, `quality-gates-context.md`, and the generated `manager-git.toml`. No published
skill (`.agents/skills/`) is in the diff — consistent with AC-GDP-030 case (A).

Added lines: 38 (`ac015-added-count.txt`).

| Check | Command output | Result |
|---|---|---|
| SPEC ID on an added line | `ac015-specid.txt`, grep exit 1 | none |
| REQ token on an added line | `ac015-req.txt`, grep exit 1 | none |
| date on an added line | `ac015-date.txt`, grep exit 1 | none |
| `CLAUDE.local` on an added line | `ac015-local-ref.txt`, grep exit 1 | none |
| 7-40 char hex word with a-f letter | `ac015-sha-letter.txt`, `test -e` 0, `test -s` 1 | none |
| digit-only 7-40 char words | `ac015-sha-digits.txt` — 0 lines (`ac015-hex-tokens.txt` 0 lines) | nothing to read; no commit SHA |

Detector control (re-run in part 2 on the pre-flight fixture `ac015-fixture.diff`):
`p2-ac015-fixture-tokens.txt` → letter tokens 5, digit tokens 2 — the same detector does find
planted SHA-shaped words, so the empty result above is not a dead detector.

Programming-language bias (read, every added line): the added lines describe git/`gh` commands,
the `--auto-merge` / `--merge` / `--no-merge` flags, the `merge_method` configuration key, and the
team / personal / manual mode approval conditions. None names or privileges any of the 16 supported
programming languages. The Korean words on the `workflows/sync.md` `**Flags**:` line (`PR 생성`,
`MX 검증 스킵`) are carried over unchanged from the BASE template line (`base-sync-template.md:104`);
the edit added only the `--auto-merge` item and the `--merge` alias wording. This is a locale-text
observation, not a programming-language bias, and is pre-existing.

CI `template-neutrality-check` result: not observed in this run — the card branch is not pushed
(lanes do not push). Recorded as a gap; the lead's develop push triggers it.

Result: all mechanical absence checks empty on a non-empty added-line set; reading finds no
programming-language bias.
