# Release tag ruleset `moai-cli-v*` — draft JSON and apply procedure

Issue: #2197 ([rc.2][E5]) · parent: #2157 (EP-E).
Design basis (all read, not linked-only): decisions-2026-11-08.md rows `15:5x`, `01:1x(11-09)`, `00:4x(11-10) 정정`; `docs/release-checklist.md` section (g) rows on the tag namespace and the HOLD rows 9/48, verification rows 130–131.

## Purpose

REQ-08D-09 (same-version redeploy ban) currently has no ruleset basis: the existing ruleset `Release tag immutability (v*)` (id 19583648) includes only `refs/tags/v*`, which does not cover the Rust release namespace `moai-cli-v<version>` (04-A decision B). This draft adds a sibling tag ruleset for `moai-cli-v*` with the same rule shape (deletion + update blocked, no bypass actors), so a Rust release tag cannot be deleted or moved once pushed.

## Draft JSON — `docs/release-tag-ruleset-moai-cli-v.json`

The file is shaped as a `POST /repos/{owner}/{repo}/rulesets` body: it carries only `name`, `target`, `enforcement`, `conditions`, `rules`, `bypass_actors` — none of the GET-only fields (`id`, `node_id`, `source_type`, `source`, `created_at`, `updated_at`, `current_user_can_bypass`, `_links`).

- `name`: `Release tag immutability (moai-cli-v*)` — mirrors the existing `Release tag immutability (v*)` naming.
- `conditions.ref_name.include`: exactly `["refs/tags/moai-cli-v*"]`, `exclude` empty. This is what satisfies REQ-E5-01 (the ruleset must not block creation of any other tag — Go namespace `v*`/`go-v*` stays untouched): GitHub ruleset ref-name patterns use fnmatch-style patterns (documented in "Creating rulesets for a repository"), and a single include pattern of this shape cannot reach the other namespaces. The name-anchored and case-sensitive matching specifics are not documented; they remain assumptions (E5 audit observation, 2026-11-10), not documented facts.
- `rules`: `deletion` + `update` — identical to the existing `v*` ruleset. Tag creation is NOT restricted; `git tag moai-cli-v<version> <candidate>` keeps working.
- `bypass_actors`: `[]` — same as the existing ruleset (`current_user_can_bypass: never` observed on id 19583648).

## Apply procedure (leader, after one-time operator approval)

0. **Operator approval (one-time gate, from the Issue body).** Repository-settings change. Record the approval in the decisions log.
1. **Apply (leader):**
   ```
   gh api -X POST repos/modu-ai/moai-adk/rulesets --input docs/release-tag-ruleset-moai-cli-v.json
   ```
   Run from the repository root — the `--input` path is repo-relative.
2. **Readback (AC-2).** After the POST, re-read and assert:
   - `gh api repos/modu-ai/moai-adk/rulesets` lists a new entry with target `tag` and the new name;
   - the detail call `gh api repos/modu-ai/moai-adk/rulesets/<new-id>` shows `enforcement: active`, `conditions.ref_name.include == ["refs/tags/moai-cli-v*"]`, `rules` = deletion + update, `bypass_actors: []`.
   Announce the apply only after this readback (no unobserved-write reporting).
3. **REQ-E5-01 verification (Go tag isolation).** Confirm that creating a tag outside the pattern still works. Optional probe: tag an existing pushed commit, push the tag, confirm creation succeeds, then delete it again (the probe matches neither this ruleset nor the existing `v*` ruleset, so both steps must be allowed):
   ```
   git tag go-v0.0.0-ruleset-probe <existing-pushed-commit-sha>
   git push origin go-v0.0.0-ruleset-probe
   git push origin :refs/tags/go-v0.0.0-ruleset-probe   # remote cleanup
   git tag -d go-v0.0.0-ruleset-probe                    # local cleanup
   ```
   `<existing-pushed-commit-sha>` is a commit already on the integration branch at apply time; record the SHA in the decisions log with the probe result. The probe push is a repository write separate from the ruleset POST, so the operator approval must state whether it covers the probe push. The name matches neither ruleset pattern (`moai-cli-v*`, `v*`), and the workflows' `on.push.tags` filter (`v*` on `main`'s `release.yml`) does not match it either (E5 audit measurement, 2026-11-10). If the probe is skipped, record it as not_run with the reason.
4. **REQ-08D-09 basis check.** Structural basis = the apply readback (step 2). Behavioral proof (a delete/recreate attempt against a live `moai-cli-v*` tag being refused) is only measurable after the first real release tag exists; until then record it as pending, not as passed.
5. **Record.** Write the apply date, new ruleset id and readback results into the decisions log, and update the HOLD/not_run rows in `docs/release-checklist.md` (rows 9, 48, and the status table rows 130–131) in the same docs change.

## Rollback procedure (leader, same operator-approval scope as the apply)

Every rollback step is itself a repository-settings change and runs under the same one-time operator approval as the apply, recorded in the decisions log. `<new-id>` is the id captured by the apply readback (step 2). The commands below are described here and are not executed by this PR. Endpoint basis, read 2026-11-10 (`X-GitHub-Api-Version: 2026-03-10` advertised on the page): Update a repository ruleset is `PUT /repos/{owner}/{repo}/rulesets/{ruleset_id}` (200; body `enforcement` accepts `disabled`, `active`, `evaluate` — `evaluate` is GitHub-Enterprise only), Delete a repository ruleset is `DELETE /repos/{owner}/{repo}/rulesets/{ruleset_id}` (success `204`, no body).

a. **Disable — preferred first step, reversible.**

   ```
   gh api -X PUT repos/modu-ai/moai-adk/rulesets/<new-id> -f enforcement=disabled
   ```

   Readback (announce only after observing): `gh api repos/modu-ai/moai-adk/rulesets/<new-id>` shows `"enforcement": "disabled"`. With the ruleset `disabled` or deleted, nothing in the ruleset list matches `refs/tags/moai-cli-v*` any more, so `moai-cli-v*` tags are deletable and movable again — that readback is the tag-unlock verification. A behavioral probe (deleting or re-pushing an actual `moai-cli-v*` tag) is a separate repository write and stays optional under the operator's approval scope; if skipped, record it as not_run with the reason.

b. **Delete — after the disable readback is observed, or directly under operator approval.**

   ```
   gh api -X DELETE repos/modu-ai/moai-adk/rulesets/<new-id>
   ```

   Expects `204`. Readback: `gh api repos/modu-ai/moai-adk/rulesets` no longer lists `<new-id>`, while the pre-existing rulesets `19583648` (tag `v*`) and `26791615` (branch `v2026.11.0`) remain listed unchanged (observed 2026-11-10T10:00Z as the only two rulesets before any apply).

c. **Record.** Write the rollback date, actor, approval reference, and both readback results into the decisions log, and revert the `docs/release-checklist.md` rows the apply step updated (rows 9, 48, and the status table rows 130–131).

## Out of scope (what this ruleset does not solve)

- **Final-release `/releases/latest` exposure** (decisions row `00:4x(11-10) 정정`): tag immutability is orthogonal to `make_latest` behavior. A final `moai-cli-v<version>` release must still be published with `--latest=false` per the release checklist/attestation procedure, or the Go stable updater breaks. This ruleset neither causes nor fixes that.
- **Tag creation restriction**: the draft blocks deletion and update only, matching the existing `v*` ruleset. Creating `moai-cli-v*` tags stays open — the release procedure (`git tag moai-cli-v<version> <candidate>` + `--verify-tag`, `gh release create --prerelease --latest=false --verify-tag`) is unaffected.

## Lane verification performed (see PR Self-Audit section for the measured record)

- Draft JSON parses: `python3 -I -m json.tool docs/release-tag-ruleset-moai-cli-v.json > /dev/null`, exit 0.
- Structural assertions per REQ-E5-01 (include is exactly `refs/tags/moai-cli-v*`; rules are deletion+update; `bypass_actors` empty; no GET-only fields) run as a validator script; a mutation cycle (include widened to `refs/tags/*` → validator FAILED → restored byte-identical vs HEAD → GREEN) measures that the assertion actually bites. The validator script and its mutation record live under the untracked `target/`; the assertions are re-derivable from the four lines above and were recomputed independently by the E5 audit.
- Check set: the full CI check set is measured at the repair head and recorded with exact exits in the PR Self-Audit section (`cargo test --offline --locked`, `cargo clippy --offline --locked --all-targets -- -D warnings`, `cargo fmt --check`, `python3 -m unittest discover -s tests -p '*_test.py'`, `git diff --check`, `python3 -I -m json.tool` on the draft JSON); the PR-body self-audit checker (`python3 -I tools/check_pr_self_audit.py`) runs measured.
