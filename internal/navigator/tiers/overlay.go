package tiers

// Enrich emits the tiers.json overlay into projectRoot's
// .moai/project/navigator/ directory.
//
// @MX:TODO: M4.2 STUB — engines land in M4.3-M4.6 (Chunk 2).
// @MX:REASON: SPEC-NAVIGATOR-SYNC-003 plan.md §F milestones — M4.2 ships the
// non-overlap guardrail (RED state intended); the contract/blueprint/ADR/
// symbol engines that actually populate TiersOverlay land in M4.3-M4.6.
//
// This stub is intentionally a no-op so the M4.2 runtime-fixture non-overlap
// test (TestNonOverlap_RuntimeFixtureWriteSurface) compiles and captures the
// intended RED state: tiers.json is absent until the real engine lands. The
// function signature is stable so Chunk 2 fills the body without touching
// callers.
//
// REQ-NS3-016 (consumer-only): the real implementation MUST NOT modify M0/M1
// producer paths. REQ-NS3-018 (write-surface isolation): the real
// implementation writes ONLY to the 6 named surfaces and NEVER overwrites
// nav-graph.json.
func Enrich(projectRoot string) error {
	// STUB: no-op. Chunk 2 implements the real emission.
	_ = projectRoot
	return nil
}
