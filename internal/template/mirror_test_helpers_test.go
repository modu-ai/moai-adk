package template

// shared-helpers-for-retained-mirror-tests.go — the helper bodies the
// retained mirror-manifest and embed-skill-guard tests share, recovered from
// the retired skill_mirror_test.go / skill_mirror_fallback_test.go files
// (SPEC-USER-ASSET-INSTALL-001 M7 deleted the tests whose subject retired
// with the project-side placement, but the surviving tests still need these
// fixtures).

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
