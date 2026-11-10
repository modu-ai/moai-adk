# Release tag ruleset `moai-cli-v*` — draft JSON and apply procedure

Issue: #2197 ([rc.2][E5]) · parent: #2157 (EP-E).
Design basis (all read, not linked-only): decisions-2026-11-08.md rows `15:5x`, `01:1x(11-09)`, `00:4x(11-10) 정정`; `docs/release-checklist.md` section (g) rows on the tag namespace and the HOLD rows 9/48, verification rows 130–131.

## Purpose

REQ-08D-09 (same-version redeploy ban) currently has no ruleset basis: the existing ruleset `Release tag immutability (v*)` (id 19583648) includes only `refs/tags/v*`, which does not cover the Rust release namespace `moai-cli-v<version>` (04-A decision B). This draft adds a sibling tag ruleset for `moai-cli-v*` with the same rule shape (deletion + update blocked, no bypass actors), so a Rust release tag cannot be deleted or moved once pushed.

## Draft JSON — `docs/release-tag-ruleset-moai-cli-v.json`

The file is shaped as a `POST /repos/{owner}/{repo}/rulesets` body: it carries only `name`, `target`, `enforcement`, `conditions`, `rules`, `bypass_actors` — none of the GET-only fields (`id`, `node_id`, `source_type`, `source`, `created_at`, `updated_at`, `current_user_can_bypass`, `_links`).

- `name`: `Release tag immutability (moai-cli-v*)` — mirrors the existing `Release tag immutability (v*)` naming.
- `conditions.ref_name.include`: exactly `["refs/tags/moai-cli-v*"]`, `exclude` empty. This is what satisfies REQ-E5-01 (the ruleset must not block creation of any other tag — Go namespace `v*`/`go-v*` stays untouched): GitHub ruleset patterns are anchored fnmatch, so a single include pattern cannot reach other tags.
- `rules`: `deletion` + `update` — identical to the existing `v*` ruleset. Tag creation is NOT restricted; `git tag moai-cli-v<version> <candidate>` keeps working.
- `bypass_actors`: `[]` — same as the existing ruleset (`current_user_can_bypass: never` observed on id 19583648).

## Apply procedure (leader, after one-time operator approval)

0. **Operator approval (one-time gate, from the Issue body).** Repository-settings change. Record the approval in the decisions log.
1. **Apply (leader):**
   ```
   gh api -X POST repos/modu-ai/moai-adk/rulesets --input docs/release-tag-ruleset-moai-cli-v.json
   ```
2. **Readback (AC-2).** After the POST, re-read and assert:
   - `gh api repos/modu-ai/moai-adk/rulesets` lists a new entry with target `tag` and the new name;
   - the detail call `gh api repos/modu-ai/moai-adk/rulesets/<new-id>` shows `enforcement: active`, `conditions.ref_name.include == ["refs/tags/moai-cli-v*"]`, `rules` = deletion + update, `bypass_actors: []`.
   Announce the apply only after this readback (no unobserved-write reporting).
3. **REQ-E5-01 verification (Go tag isolation).** Confirm that creating a tag outside the pattern still works. Optional probe: create tag `go-v0.0.0-ruleset-probe`, confirm creation succeeds, then delete it (the probe matches neither this ruleset nor the existing `v*` ruleset, so both steps must be allowed). If the probe is skipped, record it as not_run with the reason.
4. **REQ-08D-09 basis check.** Structural basis = the apply readback (step 2). Behavioral proof (a delete/recreate attempt against a live `moai-cli-v*` tag being refused) is only measurable after the first real release tag exists; until then record it as pending, not as passed.
5. **Record.** Write the apply date, new ruleset id and readback results into the decisions log, and update the HOLD/not_run rows in `docs/release-checklist.md` (row 48 and the status table rows 130–131) in the same docs change.

## Out of scope (what this ruleset does not solve)

- **Final-release `/releases/latest` exposure** (decisions row `00:4x(11-10) 정정`): tag immutability is orthogonal to `make_latest` behavior. A final `moai-cli-v<version>` release must still be published with `--latest=false` per the release checklist/attestation procedure, or the Go stable updater breaks. This ruleset neither causes nor fixes that.
- **Tag creation restriction**: the draft blocks deletion and update only, matching the existing `v*` ruleset. Creating `moai-cli-v*` tags stays open — the release procedure (`git tag moai-cli-v<version> <candidate>` + `--verify-tag`, `gh release create --prerelease --latest=false --verify-tag`) is unaffected.

## Lane verification performed (see PR Self-Audit section for the measured record)

- Draft JSON parses: `python3 -I -m json.tool docs/release-tag-ruleset-moai-cli-v.json > /dev/null`, exit 0.
- Structural assertions per REQ-E5-01 (include is exactly `refs/tags/moai-cli-v*`; rules are deletion+update; `bypass_actors` empty; no GET-only fields) run as a validator script; a mutation cycle (include widened to `refs/tags/*` → validator FAILED → restored byte-identical vs HEAD → GREEN) measures that the assertion actually bites.
- Check set: docs-only change, no Rust/Python source surface — the cargo/Python suites are not_run for this PR (no reachable behavior change); the PR-body self-audit checker (`python3 -I tools/check_pr_self_audit.py`) runs measured.
