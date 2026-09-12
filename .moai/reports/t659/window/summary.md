# t659 integration window — re-measure on the merged tree

Window held by lane-6 (`moai integration acquire --name lane-6`), granted by the lead together with the internal/cli compile slot. Pre-check before the window and again before the build: `ps` showed 0 running `go test` / `go build` processes. Every command ran serially.

Absorb: local develop `30cf7f422` (equal to `origin/develop`, `git rev-list --count --left-right` = `0 0`) merged into `WT-amend-apply` as `5ace4895d`. One conflict, `CHANGELOG.md` only (`git diff --name-only --diff-filter=U` = 1 file): both sides added entries at the head of `[Unreleased] > Fixed`. Resolved by keeping all four entries, develop's three first and this card's one after, markers removed, no wording changed or deleted — the lead approved this resolution. Post-resolution checks: conflict markers 0; each of the four SPEC ids appears once; `^## \[Unreleased\]` once; `git diff --check` exit 0.

## Re-measure (tree `5ace4895d`)

| # | Command | Exit | Result | Evidence |
|---|---|---|---|---|
| w1 | `go test ./internal/constitution/ -count=1 -cover` | 0 | `ok`, coverage 88.4% | `w1-constitution.txt` |
| w2 | `go test ./internal/cli/ -run '^TestConstitution' -count=1 -v -timeout 600s` | 0 | 13 top-level RUN, 13 PASS, 0 FAIL | `w2-cli-constitution.txt` |
| w3 | `go test ./internal/cli/ -run '^(TestResolveRegistryPath_MatchesExecute\|TestConstitutionAmend_ContainmentCheck_RelativeEnvEscape)$' …` | 0 | both PASS | `w3-cli-selectors.txt` |
| w4 | `go test ./internal/spec/ -run '^TestLinter_AC08_DanglingRuleReference$' -count=1 -v` | 0 | PASS | `w4-spec-preservation.txt` |
| w5 | `go test ./internal/template/... -count=1 -timeout 600s` | 0 | `template` ok 28.089s, `agentemit` ok, `commandemit` ok, `scripts` no test files | `w5-template.txt` |
| w6 | `go test ./internal/template/ -run '^(TestCatalogHashCoversSkillSubfiles\|TestManifestHashFormat)$' -count=1 -v` | 0 | `--- PASS: TestCatalogHashCoversSkillSubfiles`, `--- PASS: TestManifestHashFormat` | `w6-catalog-hash.txt` |
| w7 | `go vet ./internal/constitution/ ./internal/cli/` | 0 | no output | `w7-vet.txt` |
| w8 | `golangci-lint run ./internal/constitution/... ./internal/cli/...` | 0 | `0 issues.` | `w8-lint.txt` |
| w9 | `go build -o <scratch>/moai-tree ./cmd/moai` | 0 | the one build the lead allowed in this window | `w9-build.txt` |
| w10 | `<scratch>/moai-tree spec lint SPEC-CON-AMEND-APPLY-001` | 0 | `✓ No findings — all SPEC documents are valid` | `w10-lint-treebuild.txt` |

w9 + w10 close the verification-claim-integrity §2.2 Gap recorded in progress §E.4 `sync_lint_gap`: the sync-phase lint was judged by the installed build `ed71054d3`, a strict ancestor; this lint was judged by a binary built from the tree under measurement (`5ace4895d`).

## Known-red carried in from develop — not attributable to this card

`TestDestructiveTargetRegistry_CoversAllSites` (`internal/cli/update_destructive_registry_test.go`) FAILs on the merged tree (`w11-known-red.txt`, exit 1). The lead named it before the absorb: a t656 regression, lane-2 repairing it as dr0912.

Attribution measured, not assumed: `git log --first-parent --no-merges --format=%h fe8cc9875..20f844c55 -- internal/cli/update.go internal/cli/update_destructive_registry_test.go` lists 0 commits, while the same range over `internal/cli` lists 2 — so the range is live and this card touched neither file. The `+20` lines `git diff --stat 20f844c55..HEAD -- internal/cli/update.go` reports arrived with the absorb, from develop.
