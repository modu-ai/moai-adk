package hook

// protected_zone_unlink_link_test.go — card t1712: the unlink-family (rm,
// unlink) final-component narrowing. The guard's path resolution followed
// symlinks through the FINAL component and judged the TARGET, so `rm
// zone_dir/link_x` — a link outside the zone pointing at a file inside it —
// was denied although the real shell only unlinks the link (measured on
// /bin/bash 3.2.57, darwin: rc=0, the link gone, protected/secret.md intact;
// `rm -r zonelink` on a DIRECTORY symlink likewise removes only the link).
// The narrowed judgment: an rm/unlink operand whose link is its last
// component with nothing after it (no trailing separator, no "." / "..", no
// child path) is judged by the LINK'S OWN path — parent components still
// resolve through symlinks — while every dereference spelling keeps the
// target-following deny, and so do the non-unlink verbs (mv/cp/tee/truncate
// really write through a final-component symlink). Every expectation here
// encodes the sandbox measurement table recorded in the card report, not an
// assumption.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ulZoneManifest declares the protected directory the fixtures link into.
const ulZoneManifest = "  probe_zone:\n    paths: [\"protected/\"]\n"

// newUnlinkLinkRoot builds the card's measured fixture: the protected zone
// with files and a subdirectory; a link INSIDE the zone; an outside zone_dir
// holding a file symlink into the zone (the defect shape) and a broken
// symlink; root-level directory symlinks into the zone and into the zone's
// subdirectory; and a plain root file for the cp/mv controls.
func newUnlinkLinkRoot(t *testing.T) string {
	t.Helper()
	root := newZoneRoot(t, zoneShippedDoc(ulZoneManifest), "")
	protected := filepath.Join(root, "protected")
	if err := os.MkdirAll(filepath.Join(protected, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join("protected", "secret.md"):         "secret\n",
		filepath.Join("protected", "zone-file.md"):      "zonefile\n",
		filepath.Join("protected", "child.md"):          "child\n",
		filepath.Join("protected", "sub", "subfile.md"): "subfile\n",
		"plain.md": "plain\n",
	}
	for rel, body := range files {
		if err := os.WriteFile(filepath.Join(root, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	links := map[string]string{
		// the defect shape: a file symlink OUTSIDE the zone pointing INTO it
		filepath.Join("zone_dir", "link_x"): filepath.Join(protected, "secret.md"),
		// a broken symlink outside the zone: only the link itself can go
		filepath.Join("zone_dir", "broken"): filepath.Join(protected, "missing.md"),
		// a link INSIDE the zone (its own path is zone-covered)
		filepath.Join("protected", "innerlink"): filepath.Join(protected, "secret.md"),
		// directory symlinks: into the zone, and into the zone's subdirectory
		"zonelink": protected,
		"sublink":  filepath.Join(protected, "sub"),
	}
	for link, target := range links {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, link)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	return root
}

// TestProtectedZoneShellUnlinkLinkOwnAllows — the narrowed rows: a bare
// slash-less rm/unlink operand whose link sits OUTSIDE the zone must ALLOW,
// whatever the target behind it is, with flags and `--` before the operand.
// On the pre-change tree every row denies under the target-following
// judgment — the RED state this repair flips.
func TestProtectedZoneShellUnlinkLinkOwnAllows(t *testing.T) {
	cells := []struct{ name, command string }{
		{name: "defect_shape_rm", command: "rm zone_dir/link_x"},
		{name: "rm_f_flag_before_operand", command: "rm -f zone_dir/link_x"},
		{name: "rm_rf_flag_before_operand", command: "rm -rf zone_dir/link_x"},
		{name: "rm_dashdash_separator", command: "rm -- zone_dir/link_x"},
		{name: "unlink_defect_shape", command: "unlink zone_dir/link_x"},
		{name: "rm_root_dir_symlink_bare", command: "rm zonelink"},
		{name: "rm_r_dir_symlink_bare", command: "rm -r zonelink"},
		{name: "rm_broken_symlink", command: "rm zone_dir/broken"},
	}
	if len(cells) == 0 {
		t.Fatal("empty link-own cell list — the sweep would assert nothing")
	}
	t.Logf("unlink link-own sweep: %d cells", len(cells))
	for _, tc := range cells {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := newUnlinkLinkRoot(t)
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.command})
			// the family's allow convention: the control fails only on a deny
			if d == DecisionDeny {
				t.Errorf("%s: decision=%q reason=%q, want allow — the shell unlinks only the link, the zone target survives (measured)", tc.command, d, r)
			}
		})
	}
}

// TestProtectedZoneShellUnlinkDereferenceFormsStayDeny — the bypass-matrix
// family: every dereference spelling (trailing "/", "/.", a child path, a
// ".." past the link), a link whose own path is INSIDE the zone, the plain
// zone-file control, and the non-unlink verbs all STAY deny — the narrowing
// must not open them. The dir-symlink rows are real deletions in the sandbox
// measurement; the trailing-slash/dot rows on a FILE symlink remove nothing
// in reality (ENOTDIR) and keep the fail-closed deny by design.
func TestProtectedZoneShellUnlinkDereferenceFormsStayDeny(t *testing.T) {
	cells := []struct{ name, command string }{
		{name: "rm_parent_dirslink_child", command: "rm zonelink/secret.md"},
		{name: "unlink_parent_dirslink_child", command: "unlink zonelink/secret.md"},
		{name: "rm_trailing_slash_dirslink", command: "rm zonelink/"},
		{name: "rm_r_trailing_slash_dirslink", command: "rm -r zonelink/"},
		{name: "rm_rf_trailing_slash_dirslink", command: "rm -rf zonelink/"},
		{name: "rm_dot_form", command: "rm zonelink/."},
		{name: "rm_dotdot_through_dirslink", command: "rm sublink/../zone-file.md"},
		{name: "rm_dotdot_through_filelink_failclosed", command: "rm zone_dir/link_x/../zone-file.md"},
		{name: "rm_link_itself_inside_zone", command: "rm protected/innerlink"},
		{name: "rm_zone_file_control", command: "rm protected/secret.md"},
		{name: "tee_final_component_link", command: "tee zone_dir/link_x"},
		{name: "truncate_final_component_link", command: "truncate -s 0 zone_dir/link_x"},
		{name: "cp_onto_final_component_link", command: "cp plain.md zone_dir/link_x"},
		{name: "mv_onto_final_component_link", command: "mv plain.md zone_dir/link_x"},
	}
	if len(cells) == 0 {
		t.Fatal("empty dereference cell list — the sweep would assert nothing")
	}
	t.Logf("unlink dereference sweep: %d cells", len(cells))
	for _, tc := range cells {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := newUnlinkLinkRoot(t)
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.command})
			wantZoneDeny(t, tc.name, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestResolveZoneLinkTargetFinalComponentForms — the resolver-side contract
// of the narrowed judgment: a link-own operand yields NO form inside the
// zone (the link's own path is all there is), while a parent-symlink operand
// still yields the in-zone form of the final name rejoined onto the RESOLVED
// parent — including through a ".." that follows the symlink, where the
// physical walk must win over EvalSymlinks' lexical collapse.
func TestResolveZoneLinkTargetFinalComponentForms(t *testing.T) {
	root := newUnlinkLinkRoot(t)
	prev := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prev })

	inZone := func(forms []zoneForm) bool {
		for _, f := range forms {
			if strings.HasPrefix(f.Folded, "protected/") {
				return true
			}
		}
		return false
	}
	swept := 0

	swept++
	if forms := resolveZoneLinkTarget(root, filepath.Join("zone_dir", "link_x")); inZone(forms) {
		t.Errorf("link-own operand: a form resolves inside the zone, want only the link's own path: %+v", forms)
	}
	swept++
	if forms := resolveZoneLinkTarget(root, filepath.Join("zone_dir", "broken")); inZone(forms) {
		t.Errorf("broken symlink: a form resolves inside the zone, want only the link's own path: %+v", forms)
	}
	swept++
	if forms := resolveZoneLinkTarget(root, "zonelink/secret.md"); !inZone(forms) {
		t.Errorf("parent-symlink operand: no form resolves inside the zone, want the final name on the resolved parent: %+v", forms)
	}
	swept++
	if forms := resolveZoneLinkTarget(root, "sublink/../zone-file.md"); !inZone(forms) {
		t.Errorf("dotdot-through-dirlink operand: no form resolves inside the zone — the lexical .. collapse rejoined the name outside, want protected/zone-file.md: %+v", forms)
	}
	swept++
	// the regular-file control: a zone file judged by its own path stays
	// covered — the narrowing only ever un-follows a symlinked final name.
	if forms := resolveZoneLinkTarget(root, filepath.Join("protected", "secret.md")); !inZone(forms) {
		t.Errorf("regular zone file: no form resolves inside the zone, want protected/secret.md covered: %+v", forms)
	}

	if swept != 5 {
		t.Fatalf("swept %d rows, want exactly 5", swept)
	}
	t.Logf("swept=%d", swept)
}

// TestProtectedZoneShellUnlinkAbsoluteRootDotDot — the repair round: an
// ABSOLUTE operand with a LEADING "/.." walks at the filesystem root (POSIX
// root semantics: ".." at the root stays at the root), so /../<root>/alias IS
// <root>/alias and the deletion lands inside the zone — the guard must deny.
// The forced physical walk failed closed at that first ".." (skipped==0),
// losing the parent resolution, and the lexical arm cannot carry the deny
// either: its Clean collapses the leading "/.." into alias/secret.md, which
// matches no zone entry. The old EvalSymlinks fast path normalized the same
// spelling and denied. The trailing-slash row and the plain-absolute control
// pin the shapes that already denied.
func TestProtectedZoneShellUnlinkAbsoluteRootDotDot(t *testing.T) {
	cells := []struct {
		name    string
		command func(root string) string
	}{
		{name: "dotdot_root_dirslink_child", command: func(root string) string {
			return "rm /../" + root + "/zonelink/secret.md"
		}},
		{name: "dotdot_root_dirslink_interior_dotdot", command: func(root string) string {
			return "rm /../" + root + "/sublink/../zone-file.md"
		}},
		{name: "dotdot_root_dirslink_trailing_slash_stays_deny", command: func(root string) string {
			return "rm /../" + root + "/zonelink/"
		}},
		{name: "absolute_plain_stays_deny_control", command: func(root string) string {
			return "rm " + root + "/zonelink/secret.md"
		}},
	}
	if len(cells) == 0 {
		t.Fatal("empty absolute-root-dotdot cell list — the sweep would assert nothing")
	}
	t.Logf("absolute root-dotdot sweep: %d cells", len(cells))
	for _, tc := range cells {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := newUnlinkLinkRoot(t)
			h := zoneTestHandler(t, root)
			d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": tc.command(root)})
			wantZoneDeny(t, tc.name, d, r, harnessLearnerIdentity, "category", "probe_zone")
		})
	}
}

// TestZoneUnlinkLinkOwnWindowsSeparators — the classifier-level repair
// round 2: on WINDOWS a "\" separates, so `zonelink\.` and `zonelink\` name
// the directory THROUGH the link (dereference spellings, target-following
// judgment) and a "\child" suffix keeps the link a parent component; on
// POSIX the backslash is an ORDINARY filename character
// (SPEC-HOOK-ZONE-BACKSLASH-001) and the same texts stay link-own finals.
// The windows flag rides in as an argument — the package's pathSegments
// idiom — so both branches classify identically on every host. Windows shell
// semantics themselves are not executable on this darwin host: this unit is
// the measurable surface, and CI's windows matrix owns the end-to-end
// observation.
func TestZoneUnlinkLinkOwnWindowsSeparators(t *testing.T) {
	cells := []struct {
		cand    string
		windows bool
		want    bool
	}{
		// windows: the backslash separates — dereference spellings route to
		// the target-following judgment, a real final name stays link-own.
		{cand: `zonelink\.`, windows: true, want: false},
		{cand: `zonelink\`, windows: true, want: false},
		{cand: `zonelink\child`, windows: true, want: true},
		{cand: `C:\dir\link`, windows: true, want: true},
		{cand: `zonelink/child`, windows: true, want: true},
		{cand: `zonelink/`, windows: true, want: false},
		// POSIX: the backslash is a filename character — the identical texts
		// stay link-own finals (the backslash-literal SPEC's constraint).
		{cand: `zonelink\.`, windows: false, want: true},
		{cand: `zonelink\`, windows: false, want: true},
		{cand: `zone_dir/link_x`, windows: false, want: true},
		{cand: `zone_dir/link_x/`, windows: false, want: false},
		{cand: `zonelink/.`, windows: false, want: false},
		{cand: `zonelink/..`, windows: false, want: false},
	}
	if len(cells) == 0 {
		t.Fatal("empty separator cell list — the sweep would assert nothing")
	}
	t.Logf("unlink final-component separator sweep: %d cells", len(cells))
	for _, tc := range cells {
		tc := tc
		t.Run(fmt.Sprintf("%s/windows=%v", tc.cand, tc.windows), func(t *testing.T) {
			if got := zoneUnlinkLinkOwn(tc.cand, tc.windows); got != tc.want {
				t.Errorf("zoneUnlinkLinkOwn(%q, windows=%v) = %v, want %v", tc.cand, tc.windows, got, tc.want)
			}
		})
	}
}
