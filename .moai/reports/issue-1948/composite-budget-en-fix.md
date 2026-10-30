# TestCompositeBudgetRealAssembly/lane_en_glm — English laneRecheckRule shortened (lane-20)

## Claim 1 — RED (tree a175ed36d4 = integration cc09f4047c + lane-17 3f8ceb2ce7 merged into the AF branch)
- Command: `go test -p 4 -count=1 -v -run '^TestCompositeBudgetRealAssembly$' ./internal/hook/`
- Observed: `RESULT lane_en_glm: composite utf16=11058 (cap 10000, margin=-1058)`, `lane_en_glm: composite 11058 exceeds the 10000 cap`,
  `--- FAIL: TestCompositeBudgetRealAssembly/lane_en_glm`; lane_ko_glm 9759, leader_ko 8541, leader_en 9936 passed.
- Cause: the English `laneRecheckRule` (internal/hook/session_start_factory_i18n.go) was 1885 chars, about twice the Korean one.

## Claim 2 — GREEN
- The English sentence is rewritten to 1342 chars with every obligation kept: one relative next wake via `ScheduleWakeup` in a
  self-paced `/loop`, bound %[1]s; set at takeover and after a reset; nothing remains → schedule nothing, `idle: backlog empty <time>` once, stop;
  no recurring `CronCreate`, no absolute-clock one-shot; `/moai goal` armed (`cmd:`/`model:`, "arming starts no work of its own"); watchdog first
  (API-error stall such as 429, delegate report that never comes); read disk evidence before replying; ladder contract/SPEC → `jev_ask` →
  one advisory model, only if configured and available (`codex_task`/`claude_audit`, never GLM) → own judgment recorded; never wait for a reply.
  The cap is not raised.
- Command: `go test -p 4 -count=1 -v -run '^(TestRoleInjectionAssemblyBudget|TestCompositeBudgetRealAssembly)$' ./internal/hook/`
- Observed: `RESULT lane_en_glm: composite utf16=9979 (cap 10000, margin=21)`; the 4 `TestCompositeBudgetRealAssembly/*` subtests PASS and the 4
  `TestRoleInjectionAssemblyBudget/*` subtests PASS. Also `-run 'LaneRecheck|RoleCore|SD_|Cron|LaneRenew|SessionStartFactory|Ladder|SelfPaced|Wake|RoleInjection|CompositeBudget'` ok.
- A first shortening (margin 35) failed two pinned phrases (`only if configured and available`, `arming starts no work of its own`);
  they were put back verbatim in the same commit, so no pinned test text needed re-pinning.

## Gaps / residual risk
- The margin is 21 UTF-16 units: the next English addition to the lane rule or the role core breaks the cap again; the lane_en_glm budget has no room left.
- `role_composite_budget_meas_test.go` logs `lane_en_estimated: composite utf16=10878 (over=878)`; it is a measurement line of a differently-assumed
  estimate, not an assertion. Whether it was over before this change was not measured.
- ko/ja/zh sentences are unchanged.
