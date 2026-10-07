package hook

// protected_zone_backslash_repro_test.go — RED reproduction for
// SPEC-HOOK-ZONE-BACKSLASH-001 (card t1566): on POSIX a "\" inside a path
// component is an ordinary filename character, but zoneSlash rewrites it to "/"
// before the zone target resolution touches the filesystem, so a project entry
// literally named `lnk\dir` symlinked into the protected zone is validated under
// the fictional `lnk/dir` spelling and the guard allows a Write that the OS
// lands inside the zone. The repair is resolution fidelity (the symlink arm
// must follow the literal component), never a character blacklist — the
// positive-control rows pin that.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// hzbBackslashManifest declares the protected zone_dir directory under the
// probe category, the same fixture shape the guard tests use.
const hzbBackslashManifest = "  probe_zone:\n    paths: [\"zone_dir/\"]\n"

// hzbLinkName is the literal POSIX component name carrying a backslash: the OS
// reads it as ONE component, the rewrite reads it as two.
const hzbLinkName = `lnk\dir`

// hzbNewBackslashZoneRoot builds a project root whose protected zone_dir exists
// and whose entry named hzbLinkName is a directory symlink to it. Skips
// gracefully where the platform cannot create directory symlinks unprivileged.
func hzbNewBackslashZoneRoot(t *testing.T) string {
	t.Helper()
	return hzbNewZoneRootWithLink(t, hzbLinkName)
}

// hzbNewZoneRootWithLink is hzbNewBackslashZoneRoot with the link name chosen
// by the caller.
func hzbNewZoneRootWithLink(t *testing.T, linkName string) string {
	t.Helper()
	root := newZoneRoot(t, zoneShippedDoc(hzbBackslashManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "zone_dir"), filepath.Join(root, linkName)); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	return root
}

// TestCheckProtectedZonePosixBackslashLinkBypass — AC-HZB-001/002: a Write
// through a backslash-named link into the protected zone must be denied. On the
// pre-repair tree the fixture demonstrates the in-zone landing (the measured
// round-8 shape: decision allow + the protected file changed) — the non-deny
// branch performs the ACTUAL write through the literal path and reads it back
// from zone_dir. The positive-control row proves a fix cannot pass by denying
// every backslash-bearing path: the same component name over an ORDINARY
// directory outside the zone stays allowed.
func TestCheckProtectedZonePosixBackslashLinkBypass(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: on Windows a backslash is a separator and the rewrite is correct")
	}
	swept := 0

	for _, shape := range []struct {
		name string
		raw  func(root string) string
	}{
		{"absolute raw path", func(root string) string {
			return filepath.Join(root, hzbLinkName, "secret.md")
		}},
		{"relative raw path", func(string) string {
			return filepath.Join(hzbLinkName, "secret.md")
		}},
	} {
		swept++
		root := hzbNewBackslashZoneRoot(t)
		h := zoneTestHandler(t, root)
		d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite(shape.raw(root)))
		if d != DecisionDeny {
			// RED demonstration (AC-HZB-002): the guard did not deny, so the
			// actual write follows the literal backslash-named symlink —
			// perform it and prove where the OS lands it.
			literal := filepath.Join(root, hzbLinkName, "secret.md")
			if werr := os.WriteFile(literal, []byte("bypass"), 0o644); werr != nil {
				t.Errorf("%s: decision=%q reason=%q, want deny; and the demonstration write failed: %v", shape.name, d, r, werr)
				continue
			}
			landed, rerr := os.ReadFile(filepath.Join(root, "zone_dir", "secret.md"))
			if rerr != nil {
				t.Errorf("%s: decision=%q reason=%q, want deny; the write did not land inside the zone: %v", shape.name, d, r, rerr)
				continue
			}
			t.Errorf("%s: BYPASS — decision=%q reason=%q, want deny; the write through the literal backslash link landed INSIDE the protected zone (zone_dir/secret.md=%q)", shape.name, d, r, landed)
			continue
		}
		wantZoneDeny(t, shape.name, d, r, harnessLearnerIdentity, "category", "probe_zone")
		// the deny branch means the write never ran: the protected file must
		// stay absent (AC-HZB-002 green path)
		if _, err := os.Lstat(filepath.Join(root, "zone_dir", "secret.md")); !os.IsNotExist(err) {
			t.Errorf("%s: deny observed but zone_dir/secret.md exists anyway", shape.name)
		}
	}

	// positive control (plan-auditor non-blocking #2): the SAME component name
	// over an ORDINARY directory outside the zone stays allowed — a mutant that
	// blanket-denies backslash-bearing POSIX paths fails here, not on the rows
	// above.
	swept++
	root := newZoneRoot(t, zoneShippedDoc(hzbBackslashManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, hzbLinkName), 0o755); err != nil {
		t.Fatal(err)
	}
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite(filepath.Join(root, hzbLinkName, "secret.md")))
	if d == DecisionDeny || strings.Contains(r, SentinelHarnessFrozenProtectedZone) {
		t.Fatalf("positive control ordinary directory: decision=%q reason=%q, want allowed — a backslash inside a component name is a legal filename character on POSIX", d, r)
	}
	if err := os.WriteFile(filepath.Join(root, hzbLinkName, "secret.md"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "zone_dir", "secret.md")); !os.IsNotExist(err) {
		t.Errorf("positive control ordinary directory: the write reached the zone: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(root, hzbLinkName, "secret.md")); err != nil || string(data) != "ok" {
		t.Errorf("positive control ordinary directory: the write did not land in the ordinary directory: %v", err)
	}

	if swept != 3 {
		t.Fatalf("swept %d rows, want exactly 3 (2 bypass shapes + 1 positive control)", swept)
	}
	t.Logf("swept=%d", swept)
}

// TestCheckProtectedZonePosixBackslashConvertedAbsoluteness — the absoluteness
// vector of the same named instance (gate round 8, leader-forwarded): the cwd
// prepend decision reads the slash-CONVERTED spelling, so a POSIX relative raw
// `\alias/secret.md` converts to a "/"-leading form and a `C:\alias/secret.md`
// to a drive-letter form — both wrongly judged absolute, the cwd prepend is
// skipped, and no arm resolves the literal `\alias` (or `C:\alias`) symlink the
// OS would follow into the zone. Measured pre-amendment: forms=[], decision
// allow, the write landed inside the protected zone. The backslash-free
// drive-letter row pins REQ-HZB-003: `c:/x` (no backslash) keeps its
// pre-existing outside-root reading.
func TestCheckProtectedZonePosixBackslashConvertedAbsoluteness(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: on Windows a backslash is a separator and the rewrite is correct")
	}
	const aliasName = `\alias`
	swept := 0

	for _, shape := range []struct {
		name string
		link string // the literal component the raw names — the symlink's name
		raw  func(link string) string
	}{
		{"leading backslash relative", aliasName, func(link string) string {
			return filepath.Join(link, "secret.md")
		}},
		{"drive-letter prefixed relative", `C:\alias`, func(link string) string {
			return filepath.Join(link, "secret.md")
		}},
	} {
		swept++
		root := hzbNewZoneRootWithLink(t, shape.link)
		h := zoneTestHandler(t, root)
		d, r := zoneCall(t, h, "Write", harnessLearnerIdentity, zoneWrite(shape.raw(shape.link)))
		if d != DecisionDeny {
			// RED demonstration: the guard did not deny — the actual write
			// follows the literal component the OS names and lands inside the
			// protected zone.
			literal := filepath.Join(root, shape.link, "secret.md")
			if werr := os.WriteFile(literal, []byte("bypass"), 0o644); werr != nil {
				t.Errorf("%s: decision=%q reason=%q, want deny; and the demonstration write failed: %v", shape.name, d, r, werr)
				continue
			}
			landed, rerr := os.ReadFile(filepath.Join(root, "zone_dir", "secret.md"))
			if rerr != nil {
				t.Errorf("%s: decision=%q reason=%q, want deny; the write did not land inside the zone: %v", shape.name, d, r, rerr)
				continue
			}
			t.Errorf("%s: BYPASS — decision=%q reason=%q, want deny; the write through the literal component landed INSIDE the protected zone (zone_dir/secret.md=%q)", shape.name, d, r, landed)
			continue
		}
		wantZoneDeny(t, shape.name, d, r, harnessLearnerIdentity, "category", "probe_zone")
		if _, err := os.Lstat(filepath.Join(root, "zone_dir", "secret.md")); !os.IsNotExist(err) {
			t.Errorf("%s: deny observed but zone_dir/secret.md exists anyway", shape.name)
		}
	}

	// resolver level: the leading-backslash relative raw must yield an in-zone
	// form once the walk receives its own platform-correct cwd prepend.
	swept++
	root := hzbNewZoneRootWithLink(t, aliasName)
	prev := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prev })
	forms := resolveZoneTarget(root, filepath.Join(aliasName, "secret.md"))
	inZone := false
	for _, f := range forms {
		if strings.HasPrefix(f.Folded, "zone_dir/") {
			inZone = true
		}
	}
	if !inZone {
		t.Errorf("resolver: no returned form resolves inside the protected zone for the leading-backslash relative raw (want a folded form under zone_dir/): %+v", forms)
	}

	// REQ-HZB-003 regression: a backslash-free drive-letter-form raw keeps its
	// pre-existing reading — treated absolute by the shared rule, outside the
	// root, no forms, left to the outside-project check.
	swept++
	plainRoot := newZoneRoot(t, zoneShippedDoc(hzbBackslashManifest), "")
	if err := os.MkdirAll(filepath.Join(plainRoot, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	plainForms := resolveZoneTarget(plainRoot, "c:/x.zonefile")
	if len(plainForms) != 0 {
		t.Errorf("backslash-free drive-letter raw: resolveZoneTarget returned %+v, want no forms (pre-existing outside-root reading must not change)", plainForms)
	}

	if swept != 4 {
		t.Fatalf("swept %d rows, want exactly 4 (2 bypass shapes + 1 resolver + 1 regression)", swept)
	}
	t.Logf("swept=%d", swept)
}

// TestResolveZoneTargetPosixBackslashLinkDivergence — AC-HZB-003: the zone
// target resolver must see the real target through the backslash component.
// Pre-repair both arms resolve the rewritten `lnk/dir` spelling and no returned
// form lands inside zone_dir/; post-repair the symlink arm follows the literal
// component. The ordinary-directory row is the resolver-side shadow of the
// guard positive control.
func TestResolveZoneTargetPosixBackslashLinkDivergence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-specific: on Windows a backslash is a separator and the rewrite is correct")
	}
	// inZone reports whether any returned form resolves inside the protected
	// zone directory (a folded form under zone_dir/).
	inZone := func(forms []zoneForm) bool {
		for _, f := range forms {
			if strings.HasPrefix(f.Folded, "zone_dir/") {
				return true
			}
		}
		return false
	}
	pinCwd := func(root string) {
		t.Helper()
		prev := zoneGetwd
		zoneGetwd = func() (string, error) { return root, nil }
		t.Cleanup(func() { zoneGetwd = prev })
	}
	swept := 0

	for _, shape := range []struct {
		name string
		raw  func(root string) string
	}{
		{"absolute raw path", func(root string) string {
			return filepath.Join(root, hzbLinkName, "secret.md")
		}},
		{"relative raw path", func(string) string {
			return filepath.Join(hzbLinkName, "secret.md")
		}},
	} {
		swept++
		root := hzbNewBackslashZoneRoot(t)
		pinCwd(root)
		forms := resolveZoneTarget(root, shape.raw(root))
		if !inZone(forms) {
			t.Errorf("%s: no returned form resolves inside the protected zone for a path that goes through the backslash-named symlink (want a folded form under zone_dir/): %+v", shape.name, forms)
		}
	}

	// resolver-side positive control: the literal component over an ORDINARY
	// directory must never yield a zone form.
	swept++
	root := newZoneRoot(t, zoneShippedDoc(hzbBackslashManifest), "")
	if err := os.MkdirAll(filepath.Join(root, "zone_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, hzbLinkName), 0o755); err != nil {
		t.Fatal(err)
	}
	pinCwd(root)
	forms := resolveZoneTarget(root, filepath.Join(root, hzbLinkName, "secret.md"))
	if inZone(forms) {
		t.Errorf("positive control ordinary directory: a form resolves inside the zone, want outside: %+v", forms)
	}

	if swept != 3 {
		t.Fatalf("swept %d rows, want exactly 3 (2 bypass shapes + 1 positive control)", swept)
	}
	t.Logf("swept=%d", swept)
}
