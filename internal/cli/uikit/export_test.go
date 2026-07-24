package uikit

// Test-only exports (SPEC-CLI-TUX-INIT-UPDATE-001 M3).
//
// BannerString exposes the unexported compact-band builder so the external
// uikit_test package can assert the compact ◆ MoAI-ADK band stays logo-free at
// the bannerString layer. The restored large logo stacks ONLY in PrintBanner
// (the composition layer), never in bannerString — the reversal-minimizing
// invariant of §A.1 L3 / §B R6 (SPEC-CLI-TUX-V3-004's "compact band stays
// compact" intent survives at the band layer).
var BannerString = bannerString
