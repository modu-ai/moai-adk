# SPEC-FACTORY-LANE-NO-CLEAR-001 — Acceptance Criteria

This file is the verification layer. Each criterion is written as Given-When-Then, names the command that decides it, and states the observable result that counts as a pass. Verification commands run from the card worktree with literal absolute paths. Each command's verbatim output is written into the verdict file `.moai/reports/t1793/verdict.md` during the run phase; this plan phase only defines the criteria.

Targeted test packages only: `./internal/cli`, `./internal/hook`, and `./internal/template/...`. The full suite and background load are out of scope for every criterion.

## §A Scope of Verification

- Sixteen criteria, AC-LNC-001 through AC-LNC-016, one per verification surface; each traces to at least one requirement in `spec.md` § B.
- AC-LNC-003 is a run-phase observation on the deployed runtime. It cannot be discharged in the plan phase.
- Each criterion's pass result is attributed to the command run and the output observed in the run. A remembered figure is not a pass (see the attribution rule in the verification doctrine).

## §B Acceptance Criteria

### AC-LNC-001 — The transition passage names the in-session steps

**Given** the template copies of the factory rule family are edited,
**When** the template full body is searched for the step phrases,
**Then** each step phrase occurs at least once.

**Verification command**

    grep -c -F 'ExitWorktree(keep)' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
    grep -c -F 'EnterWorktree(' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
    grep -c -F 'factory next' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md
    grep -c -F 'conversation context' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md

**Expected observable result:** each command prints a count of 1 or more.

**Traces to:** REQ-LNC-001, REQ-LNC-002, REQ-LNC-003, REQ-LNC-005.

### AC-LNC-002 — No lane clear remains in the transition passages

**Given** every EDIT row of the location ledger is applied,
**When** the template copies of the rule family and the local-only lane protocol are searched for the clear literal,
**Then** every printed line is a leader-session sentence (ledger rows L01, L02, L07, L14), matched by its sentence text because line numbers shift after the edit.

**Verification command**

    grep -n -F '/clear' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/local/gitflow-lane-protocol.md

**Expected observable result:** the printed lines are the leader-session sentences only: three in the full body, one in the stub, and none in the companions, the auto-semantics rule, or the local-only lane protocol.

**Traces to:** REQ-LNC-001, REQ-LNC-006, REQ-LNC-013.

### AC-LNC-003 — The in-session switch works on the deployed runtime (run phase)

**Given** a lane session holds a leased card whose worktree exists under the project's worktree directory,
**When** the lane runs the exit that keeps the tree, then the entry into the leased worktree by its absolute path, then `git rev-parse --show-toplevel` from inside the entered tree,
**Then** the toplevel equals the leased worktree's absolute path; if any step is refused, the verbatim refusal is recorded and the lane follows the blocker path of REQ-LNC-004.

**Verification command**

    git rev-parse --show-toplevel

(run inside the entered tree, after the exit and entry steps, each step's verbatim output recorded in the verdict file)

**Expected observable result:** the printed toplevel is the leased worktree path. A refusal is recorded verbatim under Gaps in the verdict file, and no clear appears in the lane's output.

**Traces to:** REQ-LNC-001, REQ-LNC-003, REQ-LNC-004. Decision Q7 is open until this criterion is observed.

### AC-LNC-004 — The factory completion output carries no clear instruction

**Given** the clear-policy printer is changed and its tests are updated,
**When** the factory completion tests run for every clear-policy value,
**Then** each completion output carries zero clear instructions and names the in-session transition step.

**Verification command**

    go test ./internal/cli -run '^TestSD_AC020_ClearPolicies$' -count=1 -v

**Expected observable result:** the package result is `ok`; the verbose output shows the completion subtests (clear-each, absent, clear-when-full at threshold, complete output) passing with a zero clear-instruction count.

**Traces to:** REQ-LNC-007, REQ-LNC-009.

### AC-LNC-005 — The todo auto lane cycle output carries no clear instruction

**Given** the lane cycle tests gain an assertion over the completion path output and the exit guidance,
**When** the lane cycle tests run,
**Then** every test passes and no completion or exit line names a clear of the session.

**Verification command**

    go test ./internal/cli -run '^(TestAutoLaneCycleAssignedCardBeforeQueuedRanking|TestAutoLaneCycleAssignedCardWithNoQueuedCandidates|TestAutoLaneCycleEvidenceMissKeepsLease|TestAutoLaneCycleHonorsLauncherRunID|TestAutoLaneCycleJevRankBeatsQueueOrderAndContinues|TestAutoLaneCycleKeepSetRefusalFallsThrough|TestAutoLaneCycleNoteNamesNextCandidate|TestAutoLaneCycleQueuedCandidatesOnly|TestAutoLaneCycleQuotaHoldOwnAssignedCardProceeds|TestAutoLaneCycleQuotaHoldTakesNoNewCard|TestAutoLaneCycleRunsThroughCommandSurface|TestAutoLaneCycleSerialSlotRefusalStops|TestAutoLaneCycleSkipsExcludedAssignedCard)$' -count=1 -v

**Expected observable result:** the package result is `ok`, and the verbose output includes the new no-clear assertion as a passing subtest.

**Traces to:** REQ-LNC-002, REQ-LNC-008.

### AC-LNC-006 — Every clear-policy value names no clear; relaunch keeps its end-session line

**Given** the printer keeps the three accepted values,
**When** the clear-policy and carrier tests run,
**Then** clear-each and absent print no clear instruction, clear-when-full prints nothing at or above the threshold, and relaunch prints its end-session line that names no clear.

**Verification command**

    go test ./internal/cli -run '^(TestSD_AC020_ClearPolicies|TestSD_M6_ClearPolicyCarrierConstantOnly)$' -count=1 -v

**Expected observable result:** the package result is `ok`; relaunch's end-session subtest passes; no subtest asserts a clear instruction for any value.

**Traces to:** REQ-LNC-009.

### AC-LNC-007 — The SessionStart lane-join text agrees across the four languages

**Given** the English next-card text and the four recheck texts are edited,
**When** the hook tests run and the old phrases are searched for,
**Then** the hook tests pass with a takeover-trigger assertion for each language, and neither old phrase remains.

**Verification command**

    go test ./internal/hook -run '^(TestAgentStartHandler_RoutesViaFactory|TestFactoryBindBudgetStaysInsideTheHookTimeout|TestFactoryBootstrapNoticeSilentForOrdinarySession|TestFactoryBootstrapNoticeStartupOnly|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer|TestFactoryGuideAutoModeMarkerBlocksEveryLocale|TestFactoryGuideDocsPointerInBothNotices|TestFactoryGuideFallbackCompleteNotice|TestFactoryGuideNamesWorkerJoinInEveryLocale|TestFactoryGuideNewFieldsCompleteInEveryLocale|TestFactoryGuideNewFieldsLayoutInvariants|TestFactoryGuideStaleRunStaysEnglishOnAgentChannel|TestFactoryGuideTeachesLaneFormsInEveryLocale|TestFactoryHandoffOwnerHelperProcess|TestFactoryHookBatchDeadlineIsNotReportedAsUnbound|TestFactoryHookBatchForRunOpensTheNamedRun|TestFactoryHookBenchmarkBudget|TestFactoryHookBindsIntoResumedRun|TestFactoryHookContextAndContinuationSafety|TestFactoryHookPeerRefusesLegacyLabel|TestFactoryHookZeroTurnAndCapabilityTruth|TestFactoryLaneHandoffInteractiveStateMachine|TestFactoryLaneHandoffRebindVsLaunchBindRace|TestFactoryLaneJoinClosesOnTheThirdHop|TestFactoryLaneRecordsItsNumberAndLeadRecordsZero|TestFactoryLeadNoticeCarriesDispatchDiscipline|TestFactoryLeadNoticeCarriesLaneLinesSocketAndEntryGuide|TestFactoryLeadNoticeDispatchDisciplineKorean|TestFactoryLeadNoticeEmptyWithoutRunID|TestFactoryLeadNoticeIsLaneCountIndependent|TestFactoryLeadNoticePrintsLaneCommandOnce|TestFactoryLeadProviderLaneGuidance|TestFactoryMessagesLocaleFallback|TestFactoryNetSessionRecord|TestFactoryNetSessionStartNotices|TestFactoryRelaunchProseUnchanged|TestFactoryRoleFromEnvLegacyLabelsNotRecognized|TestFactoryRoleFromEnvNewVocabulary|TestFactoryRoleFromEnvReadsOnlyLaneLabels|TestFactorySessionRecord_NoStrayTreeInSubdirCWD|TestFactorySessionStartCannotRotateAuthoritativeUserPromptBinding|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptRegistrationSurfacesHandoffRefusals|TestFactoryUserPromptSubmitExhaustedBudgetStaysDegraded|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRecoversAfterFirstTurnFailure|TestFactoryWorkerNoticeCarriesSpawnAuthority|TestFactoryWorkerNoticeLocaleWordOrders|TestFactoryWorkerNoticeNamesLabel|TestNonFactorySessionWritesNoRecord|TestRebindEligibleSessionStartSilent|TestSessionEndSourceFactoryDefault|TestSessionRecordIgnoresRetiredFactoryMarkers|TestSessionStart_AuditDisabled|TestSessionStart_BlockingComparerDoesNotStallSessionStart|TestSessionStart_CacheHitStartsNoFillAndRendersAdvisory|TestSessionStart_CacheMissStartsExactlyOneFill|TestSessionStart_DeferredScanDoesNotBlockReturn|TestSessionStart_DeferredScanJoinsWithinBound|TestSessionStart_DriftCacheProbe|TestSessionStart_EnabledByDefault|TestSessionStart_FreshFilesNotWrapped|TestSessionStart_GuardLivenessAdvisoryArrivesWithNoOperatorInput|TestSessionStart_GuardLivenessRefreshOnEveryActivation|TestSessionStart_GuardLivenessReportsAContractViolation|TestSessionStart_HandleInputLagBudget|TestSessionStart_HandleStageProbe|TestSessionStart_InvokesMigrationRunner|TestSessionStart_LagAdvisoryReachesAdditionalContext|TestSessionStart_LagAdvisorySerializesUnderAdditionalContext|TestSessionStart_MemoryStaleAggregated|TestSessionStart_MemoryStaleWrap|TestSessionStart_MigrationFailure_DoesNotBlockSession|TestSessionStart_MigrationFailure_SurfacesViaSystemMessage|TestSessionStart_MigrationsDisabled_SkipsRunner|TestSessionStart_MigrationSuccess_EmitsNoMigrationMessage|TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift|TestSessionStart_NoArchiveInvocation|TestSessionStart_NoLagAdvisoryWhenBinaryMatchesHead|TestSessionStart_NoMemoryDir|TestSessionStart_OptOutStartsNoFill|TestSessionStart_StageClock|TestSessionStart_SyncDeferredScansStartNoFill|TestSessionStart_SynchronousSideEffectsPreserved|TestSessionStartAdditionalContextSkippedOnEmptySessionID|TestSessionStartAdditionalContextStrictlyAdditive|TestSessionStartAutoPathSplit|TestSessionStartBannerAtRoot|TestSessionStartBannerBackfill|TestSessionStartBannerEmptyProjectDir|TestSessionStartBannerNoContext|TestSessionStartDeferredSiblingsProceedDuringCapWalk|TestSessionStartEmitsNoKanbanNotice|TestSessionStartEmptySessionIDEmitsWarning|TestSessionStartGoalPruneFailOpen|TestSessionStartHandler_EventType|TestSessionStartHandler_GetConfig_NilProvider|TestSessionStartHandler_GetConfig_ProviderReturnsNil|TestSessionStartHandler_Handle|TestSessionStartHandler_Handle_GLMGuardrailInjected|TestSessionStartHandler_Handle_NilCfgProvider|TestSessionStartHandler_Handle_NonGLMNoGuardrail|TestSessionStartHandler_Handle_NonWindowsGuard|TestSessionStartHandler_Handle_PreserveGLMEnvPath|TestSessionStartHandler_Handle_WindowsOnlyInjection|TestSessionStartHandler_Handle_WithProjectConfig|TestSessionStartHandler_Handle_WithProjectDirAndConfig|TestSessionStartHandlerCountIsUnchanged|TestSessionStartHygieneBestEffort|TestSessionStartInjectsAdditionalContext|TestSessionStartLineageBanner|TestSessionStartMemoryBudget_AsyncScanKeepsItsReader|TestSessionStartMemoryBudget_BelowAbsentUnreadableKillswitch|TestSessionStartMemoryBudget_CancelledAsyncScanIsSilent|TestSessionStartMemoryBudget_FourSources|TestSessionStartMemoryBudget_JoinBound|TestSessionStartMemoryBudget_JoinBoundBelowHookTimeout|TestSessionStartMemoryBudget_RecorderSeesAdvisoryRead|TestSessionStartMemoryBudget_StaysUnderTempHome|TestSessionStartMultiSessionProtocol|TestSessionStartMultiSessionProtocolEmptySessionID|TestSessionStartMultiSessionProtocolWithExistingEntry|TestSessionStartMXColdStartIntegration|TestSessionStartProbe_AC002_All3TiersAbsentEmitsWarning|TestSessionStartProbe_AC003_NonBlockingExit0|TestSessionStartProbe_AC004_OncePerSessionDedup|TestSessionStartProbe_AC006_TemplateFirstMirrorParity|TestSessionStartProbe_AC007_WarningContentActionability|TestSessionStartProbe_EdgeCases_EmptyAndMalformedStdin|TestSessionStartProfileLeaseDirectAndTokenFallback|TestSessionStartProfileLeaseUnavailableFailsClosedWithoutPanic|TestSessionStartPruneFailsOpenOnMalformedConfig|TestSessionStartPruneFailsOpenOnUnreadableConfig|TestSessionStartPruneHonorsConfiguredWindow|TestSessionStartPruneHonorsZeroRetentionOverride|TestSessionStartPrunesExpiredRecordsByDefault|TestSessionStartPrunesGoalOrphans|TestSessionStartPruneSparesRecentRecords|TestSessionStartRoleRulesCodexOnlyDeploymentLayout|TestSessionStartRoleRulesCompositeNilGuard|TestSessionStartRoleRulesCoverageGuard|TestSessionStartRoleRulesDirectivePathsResolveFromDisk|TestSessionStartRoleRulesFailVisible|TestSessionStartRoleRulesFifoFailVisible|TestSessionStartRoleRulesHandleIntegration|TestSessionStartRoleRulesInjectionPerMarkerAndSource|TestSessionStartRoleRulesMarkerSequence|TestSessionStartRoleRulesNoInjection|TestSessionStartRoleRulesOperatorLocales|TestSessionStartRoleRulesOverflowDirectiveSurvivesPrefixCut|TestSessionStartRoleRulesRootFollowsSessionCWD|TestSessionStartRoleRulesRootFromSubdirectoryCWD|TestSessionStartRoleRulesRootStopsAtNestedMoAIProject|TestSessionStartRoleRulesRootStopsAtProjectBoundary|TestSessionStartRoleRulesSizeGate|TestSessionStartWritesSideChannel|TestStaleRunNoticeFactoryLegacyLabel)$' -count=1 -v
    grep -c -F 'follow the clear policy' internal/hook/session_start_factory_i18n.go
    grep -c -F 'after every clear' internal/hook/session_start_factory_i18n.go

**Expected observable result:** the package result is `ok`; both grep counts print `0`.

**Traces to:** REQ-LNC-010.

### AC-LNC-008 — The role-core injection test asserts the new transition rule

**Given** the ALB-0294 row of the role-core test is replaced with the new transition phrases asserted in the injected core,
**When** the role-core test runs,
**Then** every row passes against the template and the local full body.

**Verification command**

    go test ./internal/hook -run '^TestInjectedCoreCarriesObligationElements$' -count=1 -v

**Expected observable result:** the package result is `ok`, and the verbose output includes the new transition row for both the template and the local surface.

**Traces to:** REQ-LNC-015, REQ-LNC-013.

### AC-LNC-009 — The leader-side surfaces do not request a lane clear

**Given** the leader's description phrase and return template line are edited in the local and template agent definitions and the template codex TOML is regenerated,
**When** the three surfaces are searched for the lane clear request and the old description phrase,
**Then** no line matches.

**Verification command**

    grep -rn -F '/clear {session-name}' .claude/agents/moai/manager-lead.md internal/template/templates/.claude/agents/moai/manager-lead.md internal/template/templates/.codex/agents/moai/manager-lead.toml
    grep -rn -F 'between cards' .claude/agents/moai/manager-lead.md internal/template/templates/.claude/agents/moai/manager-lead.md

**Expected observable result:** both commands print nothing (exit status 1).

**Traces to:** REQ-LNC-011.

### AC-LNC-010 — The operator-session serial cycle keeps its guidance, scoped to the operator session

**Given** the operator-session guidance is retained (decision Q1 default) and each surface that states it is scoped,
**When** the guidance test runs and the scoped surfaces are searched,
**Then** the guidance test passes and each surface names the operator session.

**Verification command**

    go test ./internal/cli -run '^TestTodoAutoClearGuidancePerCard$' -count=1 -v
    grep -c -F 'operator session' internal/template/templates/.claude/skills/moai/workflows/gtd.md
    grep -c -F 'operator session' internal/template/templates/.claude/agents/moai/manager-todo.md

**Expected observable result:** the package result is `ok`; each grep count is 1 or more.

**Traces to:** REQ-LNC-012.

### AC-LNC-011 — The factory doctrine family states one transition rule and drops the old isolation clause

**Given** the stub, the full body, and the companions are edited,
**When** the family is searched for the retired isolation wording and for the transition phrase,
**Then** the retired wording is absent from every family copy, and the transition phrase appears in the full body and the stub.

**Verification command**

    grep -rln -F 'moving into another card worktree is prohibited' internal/template/templates/.claude/rules/moai/workflow .moai/policies/workflow
    grep -rln -F 'end the session and launch inside the new tree' internal/template/templates/.claude/rules/moai/workflow .moai/policies/workflow
    grep -c -F 'ExitWorktree(keep)' internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md

**Expected observable result:** the first two commands print nothing; the third prints 1 or more.

**Traces to:** REQ-LNC-013.

### AC-LNC-012 — Template-first byte parity for every changed artifact

**Given** the template sources are edited and `make build` has passed (AC-LNC-014),
**When** each local counterpart is compared with its template source,
**Then** every pair is byte-identical.

**Verification command**

    cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md
    cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md .claude/rules/moai/workflow/factory-dispatch-core.md
    cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md .claude/rules/moai/workflow/factory-dispatch-detail.md
    cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md .claude/rules/moai/workflow/factory-dispatch-cards.md
    cmp internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md .claude/rules/moai/workflow/factory-dispatch-mechanics.md
    cmp internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md .claude/rules/moai/workflow/auto-semantics.md
    cmp internal/template/templates/.claude/agents/moai/manager-lead.md .claude/agents/moai/manager-lead.md
    cmp internal/template/templates/.claude/agents/moai/manager-todo.md .claude/agents/moai/manager-todo.md
    cmp internal/template/templates/.claude/skills/moai/workflows/gtd.md .claude/skills/moai/workflows/gtd.md

**Expected observable result:** every command exits 0 with no output. The root dogfood codex TOML and the policy copies are outside this criterion (decisions Q4 and Q5).

**Traces to:** REQ-LNC-014.

### AC-LNC-013 — Policy copies carry the transition passages; generated codex definitions match their emission

**Given** the surgical policy edits (decision Q5 default) and the regenerated codex TOMLs,
**When** the transition phrases are counted in the policy copy and the rule file, and the emission check runs,
**Then** each phrase appears in both files and the emission check passes.

**Verification command**

    grep -c -F 'ExitWorktree(keep)' .moai/policies/workflow/factory-dispatch.md
    grep -c -F 'EnterWorktree(' .moai/policies/workflow/factory-dispatch.md
    grep -c -F 'conversation context' .moai/policies/workflow/factory-dispatch.md
    make agents-emit-check

**Expected observable result:** each grep count is 1 or more; `make agents-emit-check` exits 0.

**Traces to:** REQ-LNC-014.

### AC-LNC-014 — The build gate passes with the edited templates

**Given** the template sources and the regenerated TOMLs are in place,
**When** the build target runs,
**Then** the build succeeds, including the emission, command-skill, tool-policy, and embed-manifest checks.

**Verification command**

    make build

**Expected observable result:** the command exits 0 and reports no drift message from the emission, command-skill, tool-policy, or embed-manifest checks.

**Traces to:** REQ-LNC-014.

### AC-LNC-015 — The template parity and golden tests pass

**Given** the template tree and the local copies are edited and synchronized,
**When** the template parity, mirror, and golden tests run,
**Then** every matching test passes; each package that matched at least one test shows at least one passing test line.

**Verification command**

    go test ./internal/template/... -run '^(TestBacklogJSONDisclosure_EmbeddedTemplatesMatchSource|TestBacklogJSONDisclosure_TemplateMirrorIsComplete|TestBranchProtectionDetectFilterCoversParityInput|TestBranchProtectionParity|TestCGEmbeddedRetirementPreservesRoutingAndAudit|TestContractModeLocalTemplateParity|TestDeclaredRuleMirrorForks|TestDotgithubDeployedParity|TestEmbed_SkillNamePrefixInvariant|TestEmbed_TemplateTreeHasNoSymlinks|TestEmbeddedGitignoreCoversGeneratedArtifacts|TestEmbeddedGitignoreHasMoaiBackupsSlash|TestEmbeddedMoaiSkillNames|TestEmbeddedSetEqualsManifest|TestEmbeddedTemplates_AgentDefinitions|TestEmbeddedTemplates_AGENTSmd|TestEmbeddedTemplates_Announcements|TestEmbeddedTemplates_Gitignore|TestEmbeddedTemplates_LLMConfig|TestEmbeddedTemplates_NoCacheOrOSFiles|TestEmbeddedTemplates_NoGitHooks|TestEmbeddedTemplates_NoLSPConfig|TestEmbeddedTemplates_NoPythonHooks|TestEmbeddedTemplates_NoSettingsJSON|TestEmbeddedTemplates_OutputStyles|TestEmbeddedTemplates_PathStripping|TestEmbeddedTemplates_ReturnsValidFS|TestEmbeddedTemplates_RuleFiles|TestEmbeddedTemplates_SkillDefinitions|TestEmbeddedTemplates_WalkDirTotalCount|TestEmbeddedTemplatesExcludeCILoopSkill|TestEmbeddedTemplatesNoRuntimeContamination|TestEmbedFSPresenceAndByteEquality|TestFactoryStopHookSynchronousParity|TestGitignore_IgnoresSkillMirrorOnly|TestGoalAutoWorkflowContractAndMirrorParity|TestGoldenCommittedArtifactsMatchEmission|TestGTDCanonicalSurfaceGolden|TestHMPMatcherParityLocal|TestHMPMatcherParityMutants|TestHMPMatcherParityTemplate|TestHtmlReportOutputPathParity|TestHtmlReportOutputPathParity_MutationObservedRed|TestLateBranchTemplateMirror|TestLoadEmbeddedCatalog_LoadCatalogErrorWrapping|TestLoadEmbeddedCatalog_Success|TestOutputStyleLocalizationTableParity|TestOutputStylesTemplateLiveParity|TestParityVerdictAggregate|TestPipelineCarrySignal_EmbeddedMatchesSource|TestPipelineCarrySignal_LocalMirrorMatches|TestPipelineCarrySignal_TemplateMirror|TestRenderHookEntriesFromEmbeddedTemplate|TestRuleTemplateMirrorDrift|TestSanitizedPairParity|TestSettingsTemplatePowerShellDenyParity|TestShippedRetiredModelKeys_ReadsTheEmbeddedTemplate|TestStandingPolicyParity|TestSyncPhaseGateCheckerSweepParityWithReportsKey|TestSyncPhaseGateSurvivingRootsSweepParityWithReportsKey)$' -count=1 -v

**Expected observable result:** the result is `ok` for each package; the verbose output shows at least one `--- PASS` line for each package that carries matching tests, and no `--- FAIL` line.

**Traces to:** REQ-LNC-014, REQ-LNC-015.

### AC-LNC-016 — The remaining `/clear` references match the ledger, file by file

**Given** every EDIT row is applied and every KEEP and PENDING row is left as recorded,
**When** the count of literal clear references is taken for each scoped file,
**Then** each count equals the number the table below states, and every remaining line is recorded in the verdict file with its kind.

**Verification command**

    grep -c -F '/clear' <each file in the table below>

| Scoped file | Expected count | Remaining kinds |
|-------------|----------------|-----------------|
| `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md` (and the local copy) | 3 | leader session's own clear (three lines) |
| `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md` (and the local copy) | 1 | leader session's own clear |
| `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md` (and the local copy) | 0 | none |
| `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-cards.md` (and the local copy) | 0 | none |
| `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` (and the local copy) | 0 | none |
| `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` (and the local copy) | 0 | none |
| `.claude/rules/local/gitflow-lane-protocol.md` | 0 | none |
| `.claude/agents/moai/manager-lead.md` (and the template) | 2 | leader session's own clear; leader's own context fallback |
| `.claude/agents/moai/manager-todo.md` (and the template) | 1 | operator-session serial cycle |
| `.claude/skills/moai/workflows/gtd.md` (and the template) | 2 | leader session's own clear; operator-session serial cycle |
| `.moai/policies/workflow/factory-dispatch.md` | 2 | leader session's own clear (two lines) |
| `.moai/policies/workflow/factory-dispatch-detail.md` and `factory-dispatch-cards.md` | 0 | none |
| `.moai/policies/workflow/factory-dispatch-mechanics.md` | 0 | none |
| `internal/template/templates/.codex/agents/moai/manager-lead.toml` | 2 | leader session's own clear; leader's own context fallback |
| `internal/template/templates/.codex/agents/moai/manager-todo.toml` | 1 | operator-session serial cycle |
| `internal/cli/factory_card.go` | 0 | none |
| `internal/cli/factory_relaunch.go` | 0 | none |
| `internal/hook/session_start_factory_i18n.go` | 0 | none |

Go test files (`internal/cli/factory_m6_test.go`, `internal/hook/role_core_elements_test.go`) are reconciled in the verdict file: each remaining literal line is an absence assertion or a comment that names an absence, and carries its ledger row.

**Expected observable result:** every count matches the table; the verdict file lists each remaining line with its ledger row and kind.

**Traces to:** REQ-LNC-011, REQ-LNC-012, REQ-LNC-016.

## §C Edge Cases

- **Clear-when-full above threshold** (decision Q2 default): the completion output prints no clear instruction even when the recorded context is at or above the threshold. Covered by AC-LNC-006.
- **Absent clear policy**: the default reads as clear-each and prints no clear instruction. Covered by AC-LNC-004 and AC-LNC-006.
- **Relaunch value**: prints its end-session line and names no clear. Covered by AC-LNC-006.
- **In-session entry refused**: the verbatim refusal is recorded and the lane takes the blocker path; no clear appears. Covered by AC-LNC-003.
- **Lease contention**: the lane waits out the claim with the verb's wait option; the wait output names no clear. Covered by AC-LNC-005.
- **Source clear at session start**: the SessionStart lane rule still fires on source clear and on source startup, unchanged. Covered by AC-LNC-007.
- **Operator-session cycle**: keeps its guidance, scoped to the operator session. Covered by AC-LNC-010.
- **Four conversation languages**: each recheck text carries the takeover trigger. Covered by AC-LNC-007.

## §D REQ-to-AC Matrix

| Requirement | Acceptance criteria |
|-------------|---------------------|
| REQ-LNC-001 | AC-LNC-001, AC-LNC-002, AC-LNC-003 |
| REQ-LNC-002 | AC-LNC-001, AC-LNC-005 |
| REQ-LNC-003 | AC-LNC-001, AC-LNC-003 |
| REQ-LNC-004 | AC-LNC-003 |
| REQ-LNC-005 | AC-LNC-001 |
| REQ-LNC-006 | AC-LNC-002 |
| REQ-LNC-007 | AC-LNC-004 |
| REQ-LNC-008 | AC-LNC-005 |
| REQ-LNC-009 | AC-LNC-004, AC-LNC-006 |
| REQ-LNC-010 | AC-LNC-007 |
| REQ-LNC-011 | AC-LNC-009, AC-LNC-016 |
| REQ-LNC-012 | AC-LNC-010, AC-LNC-016 |
| REQ-LNC-013 | AC-LNC-002, AC-LNC-008, AC-LNC-011 |
| REQ-LNC-014 | AC-LNC-012, AC-LNC-013, AC-LNC-014, AC-LNC-015 |
| REQ-LNC-015 | AC-LNC-008, AC-LNC-015 |
| REQ-LNC-016 | AC-LNC-016 |

## §E Quality Gates

- Template first: every changed doctrine or agent artifact is edited in its template source before its local copy is synchronized (AC-LNC-012, AC-LNC-014).
- Targeted tests only; no full-suite run and no background load (every criterion above).
- The SPEC itself passes the spec lint with no new non-advisory finding against the existing baseline (run after commit; the baseline file is not edited).

## §F Definition of Done

- AC-LNC-001 through AC-LNC-016 each pass, with the command and its observed output recorded in `.moai/reports/t1793/verdict.md`. AC-LNC-003 is observed on the deployed runtime, not inferred.
- Decision rows Q1 through Q8 remain without operator verdicts. The authoring and run phases do not write any verdict line.
- Every `/clear` reference in the scoped files carries its ledger row and kind (AC-LNC-016).
- The release-facing text carries no time estimate and no emoji.
