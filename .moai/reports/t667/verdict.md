# t667 Verdict — Web Console settings screen rejects gateway family path as profile name

Card: t667 · Tier S · Class B · Branch `WT-console-gateway-cfg` (based on local develop `e708884b0`)

## Claim

The web console `/settings` failure was NOT stored-data pollution (branch a) — it was the
current-profile derivation (branch b) returning the raw `CLAUDE_CONFIG_DIR` path as a profile
name when the config dir lies outside the profile base. Fixed at the producer
(`profile.GetCurrentNameForProject`): out-of-base config dirs degrade to `"default"`.

## Root cause (producer chain, measured)

1. `internal/cli/gateway_session.go:160` — gateway child sessions get
   `CLAUDE_CONFIG_DIR=~/.moai/state/gateway-conversations/families/<uuid>/native`.
2. `internal/profile/profile.go` `GetCurrentNameForProject` — when `CLAUDE_CONFIG_DIR` is set
   but outside `~/.moai/claude-profiles`, `filepath.Rel` yields `..` and the old code
   `return configDir` (raw absolute path as a profile name).
3. `internal/cli/web.go:153` — `moai web` seeds `Config.ProfileName` from that derivation.
4. `internal/web/handlers.go:229` / `screens.go:104` — `/settings` validates the name with
   `profile.IsValidProfileName` → 400 `invalid profile name <family path>`.

Surface split (card question: profile list / active-profile read / both):
- **Active-profile read only.** `GET /settings` → 400. `GET /` → 200. The profile list
  (`profile.List`, directory-name enumeration) is defect-free.
- Branch (a) excluded by measurement: `~/.moai/claude-profiles/launch.yaml` contains no
  `families` path; the polluted value is derived live per launch, never persisted.

## Fix

`internal/profile/profile.go` (`GetCurrentNameForProject`): out-of-base config dir →
`return "default"` (same fallback the ledger path uses) instead of `return configDir`.
All consumers already handle `"default"`; the raw path was incoherent downstream anyway
(`GetProfileDir` returns `""` for invalid names → base prefs).

## Evidence

- RED (pre-fix): `go test ./internal/profile/ -run 'TestGetCurrentName_UnrelatedPath|TestGetCurrentName_GatewayFamilyPath' -count=1`
  → both FAIL; observed `GetCurrentName() = "/Users/goos/.moai/state/gateway-conversations/families/40827bc6-7d20-4373-abb4-58c9138a0ea5/native"`.
- Screen reproduction (pre-fix): binary built from this worktree, launched with the family
  env on port 30415 → `curl /settings` → HTTP 400 with the operator's exact error string;
  `curl /` → 200.
- GREEN (post-fix): `go test ./internal/profile/ ./internal/web/ -count=1` → both `ok`.
- Screen recovery (post-fix): rebuilt, relaunched with the same family env on port 30416 →
  `GET /settings` → 200, no error banner; `GET /` → 200.
- `golangci-lint run internal/profile/...` → `0 issues`.
- `go vet ./internal/profile/ ./internal/cli/` → clean.

## Regression added

- `TestGetCurrentName_GatewayFamilyPath` — exact family-path shape → `"default"`.
- `TestGetCurrentName_UnrelatedPath` — contract updated: out-of-base → `"default"` (was
  pinning the defective raw-path passthrough).

## Gaps

- Full-suite run not executed locally (lane-local verification scope; CI owns the full suite).
- `internal/cli` consumers of `GetCurrentName` (`update`, `init`, `downgrade_locale`,
  `profile current`) not exercised end-to-end in a family-env shell; behavior change for
  them is confined to out-of-base `CLAUDE_CONFIG_DIR` environments, where they now receive
  `"default"` (coherent) instead of an absolute path (which `ReadPreferences` already
  degraded to base prefs).

## Residual-risk

- A user who legitimately runs with an out-of-base `CLAUDE_CONFIG_DIR` and previously relied
  on `moai profile current` printing the raw path will now see `default`. No code consumer
  depended on the raw value (verified by caller inspection); only display output changes.
- The gateway session itself is still not a named profile — the console shows `default`
  preferences in that context, which is the truthful degradation.
