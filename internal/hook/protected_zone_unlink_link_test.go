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
	"runtime"
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
// Repair round 9 branches the EXPECTATION on GOOS: since round 8 windows
// defers the narrowing entirely, these operands take the target-following
// judgment there and DENY — the documented windows over-block, whose deny
// reason carries the same zone sentinel (probe_zone) the dereference cells
// assert. POSIX keeps the allow.
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
			if runtime.GOOS == "windows" {
				// round 8: windows defers the narrowing — the target-
				// following judgment follows the final link and denies the
				// zone access the POSIX narrowing allows through
				wantZoneDeny(t, tc.name, d, r, harnessLearnerIdentity, "category", "probe_zone")
				return
			}
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

// TestZoneLinkTargetDefersToFullResolverWithoutExactEntry — repair round 14:
// EXACT-MATCH-ONLY narrowing. The link's own judgment is complete only when
// the operand's final component is byte-equal to an entry of the resolved
// parent; every other spelling (fold/normalization/short-name variants,
// nonexistent finals, listing failures) defers WHOLESALE to the full
// resolver — the two semantics never mix for one operand, so
// resolveZoneLinkTarget returns EXACTLY what resolveZoneTarget returns for
// such an operand.
func TestZoneLinkTargetDefersToFullResolverWithoutExactEntry(t *testing.T) {
	root := newUnlinkLinkRoot(t)
	prev := zoneGetwd
	zoneGetwd = func() (string, error) { return root, nil }
	t.Cleanup(func() { zoneGetwd = prev })

	for _, raw := range []string{
		filepath.Join("zonelink", "missing.md"), // nonexistent final
		filepath.Join("zonelink", "SECRET.md"),  // case variant of a real entry
		filepath.Join("zone_dir", "LINK_X"),     // case variant on the link name
	} {
		linkOwn := resolveZoneLinkTarget(root, raw)
		full := resolveZoneTarget(root, raw)
		if fmt.Sprint(linkOwn) != fmt.Sprint(full) {
			t.Errorf("deferral: resolveZoneLinkTarget(%q) = %+v, resolveZoneTarget = %+v — no exact entry match must defer to the full resolver wholesale", raw, linkOwn, full)
		}
	}
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
			return "rm " + zoneAbsOperand(root, "/zonelink/secret.md")
		}},
		{name: "dotdot_root_dirslink_interior_dotdot", command: func(root string) string {
			return "rm " + zoneAbsOperand(root, "/sublink/../zone-file.md")
		}},
		{name: "dotdot_root_dirslink_trailing_slash_stays_deny", command: func(root string) string {
			return "rm " + zoneAbsOperand(root, "/zonelink/")
		}},
		{name: "absolute_plain_stays_deny_control", command: func(root string) string {
			// the plain control carries no "/.." — the same QUOTING applies
			// (a windows-origin backslash path must not reach the bash
			// parser unquoted), the volume stays untouched
			return `rm "` + root + `/zonelink/secret.md"`
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

// zoneAbsOperand composes the QUOTED rooted ".." operand for a shell-command
// fixture (repair round 7 P2): the path rides zoneRootedDotDotPath (volume
// split, round 6) and the double quotes keep a windows-origin path's
// backslashes out of the bash parser's escape handling.
func zoneAbsOperand(root, final string) string {
	return `"` + zoneRootedDotDotPath(root, final, false) + `"`
}

// zoneBashText rewrites the native separators of a shell-command fixture to
// "/" (card t1893). The command text is parsed as BASH: an unquoted
// backslash is an escape the parser removes, so a Windows-host path such as
// C:\Users\RUNNER~1\Temp\x concatenated into a command reaches the guard as
// C:UsersRUNNER~1Tempx — a path that names nothing, which the guard
// correctly allows (windows-latest job 117833434303: every denied row of
// TestUserRootSymlinkSiblings read "expected denied, observed allowed").
// Git-bash spells such a path with "/" as well. A no-op on POSIX hosts.
func zoneBashText(command string) string {
	return strings.ReplaceAll(command, "\\", "/")
}

// TestZoneAbsOperandQuotedAndVolumeFree — repair round 7 P2: the absolute
// shell-command fixture composes through the round-6 helper (volume split
// BEFORE the ".." prefix) and QUOTES the operand — on a Windows host the raw
// composition embeds the drive letter mid-path ("\..\C:\...") and the
// unquoted backslashes would be eaten by the bash parser. The volume is a
// property of the root string, so the property runs on every OS.
func TestZoneAbsOperandQuotedAndVolumeFree(t *testing.T) {
	const syntheticRoot = `C:\Users\goos\T\root001`
	got := zoneAbsOperand(syntheticRoot, "/zonelink/secret.md")
	// the operand stays a BASH string: POSIX "/.." separators with the
	// volume-split remainder's own backslashes preserved, QUOTED so the
	// bash parser cannot eat them
	if want := `"/../Users\goos\T\root001/zonelink/secret.md"`; got != want {
		t.Errorf("zoneAbsOperand(%q, ...) = %q, want %q — the operand must ride the volume-split composition and be double-quoted", syntheticRoot, got, want)
	}
}

// TestZoneNameIndexExactMembership — repair round 14: the narrowing is
// EXACT-MATCH-ONLY. The index answers ONE question — is the operand's final
// component byte-equal to a real entry of the resolved parent? — and every
// non-exact spelling (fold/normalization/short-name variants, nonexistent
// finals) defers to the full resolver; the rounds-11..13 fold/normalization
// matcher cells are deleted with the machinery. Host-portable: the cells are
// pure over a names slice, and the ReadDir-backed cell harvests names
// exactly the way the resolver does.
func TestZoneNameIndexExactMembership(t *testing.T) {
	cells := []struct {
		names []string
		final string
		want  bool
	}{
		// byte-equal entries match — including when a differently-cased
		// sibling shares the listing (a case-sensitive volume can hold both)
		{names: []string{"Ä.txt", "ä.txt"}, final: "ä.txt", want: true},
		{names: []string{"Ä.txt", "ä.txt"}, final: "Ä.txt", want: true},
		{names: []string{"secret.md", "zz.md"}, final: "secret.md", want: true},
		// every variant spelling is NOT an exact match — those operands defer
		// to the full resolver (round 14); the deleted fold matcher used to
		// answer true for this one
		{names: []string{"Ä.txt"}, final: "ä.txt", want: false},
		{names: []string{"report~1.txt"}, final: "REPORT~1.TXT", want: false},
		// locale fold pairs (ı/I — NTFS locale case tables equate them) are
		// byte-inequal like any other variant: deferred, never link-own
		{names: []string{"I.txt"}, final: "ı.txt", want: false},
		{names: []string{"secret.md", "zz.md"}, final: "missing.txt", want: false},
		{names: nil, final: "ä.txt", want: false},
	}
	if len(cells) == 0 {
		t.Fatal("empty exact-membership cell list — the sweep would assert nothing")
	}
	t.Logf("exact-membership sweep: %d cells", len(cells))
	for _, tc := range cells {
		if got := newZoneNameIndex(tc.names).has(tc.final); got != tc.want {
			t.Errorf("exact membership(%q, %q) = %v, want %v", tc.names, tc.final, got, tc.want)
		}
	}
	// a real listing harvested the way the resolver does — os.ReadDir on the
	// resolved parent — feeding the same index
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Ä.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	idx := newZoneNameIndex(names)
	// repair round 15 P2-b: the fixture name is queried AS THE FILESYSTEM
	// RETURNED it — HFS+-style volumes hand back the DECOMPOSED (NFD)
	// spelling of the NFC source literal, and the byte-exact index (correct
	// by design) must be probed with the returned spelling, not the source
	// literal; the cell then passes on NFC-storage and NFD-storage volumes
	// alike. The negative probe uses a byte-different fold variant, false on
	// both storage forms.
	if len(names) == 0 {
		t.Fatal("ReadDir-backed fixture produced no entries — the probes would assert nothing")
	}
	if !idx.has(names[0]) {
		t.Errorf("ReadDir-backed membership(%q, %q) = false, want true — the exact index must accept the spelling the filesystem returned", names, names[0])
	}
	if idx.has("ä.txt") {
		t.Errorf("ReadDir-backed fold-variant membership(%q, %q) = true, want false — a fold variant is not an exact match and defers to the full resolver", names, "ä.txt")
	}
}

// TestZoneParentListingReadOncePerGuardCall — repair round 12 P2: N
// candidates sharing one resolved parent read that parent's listing ONCE
// per guard call (the cache the candidate loop threads through every
// resolution), never list-per-candidate and never across guard calls.
func TestZoneParentListingReadOncePerGuardCall(t *testing.T) {
	root := newUnlinkLinkRoot(t)
	prev := zoneParentLister
	counts := map[string]int{}
	zoneParentLister = func(dir string) ([]os.DirEntry, error) {
		counts[dir]++
		return os.ReadDir(dir)
	}
	t.Cleanup(func() { zoneParentLister = prev })
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{
		"command": "rm zone_dir/link_x zone_dir/broken zone_dir/innermissing",
	})
	if runtime.GOOS == "windows" {
		// round 8: windows defers the narrowing — every operand takes the
		// target-following judgment and denies (the documented over-block);
		// the link-own candidate loop never runs there, so no link-own
		// listing is taken at all (repair round 15 P2-a)
		wantZoneDeny(t, "listing cache", d, r, harnessLearnerIdentity, "category", "probe_zone")
		listedWindows := 0
		for dir := range counts {
			if strings.HasSuffix(filepath.ToSlash(dir), "/zone_dir") {
				listedWindows++
			}
		}
		if listedWindows != 0 {
			t.Errorf("windows link-own listing: %d /zone_dir keyed entries observed, want 0 — windows never routes link-own", listedWindows)
		}
		return
	} else if d == DecisionDeny {
		t.Fatalf("listing cache: decision=%q reason=%q, want the operands judged (links outside the zone)", d, r)
	}
	listed := 0
	for dir, n := range counts {
		if !strings.HasSuffix(filepath.ToSlash(dir), "/zone_dir") {
			continue
		}
		listed++
		if n != 1 {
			t.Errorf("parent listing re-read: %s listed %d times in ONE guard call, want once — the per-call cache absorbs the sibling candidates", filepath.ToSlash(dir), n)
		}
	}
	if listed != 1 {
		t.Errorf("parent listing re-read: %d /zone_dir keyed entries observed, want exactly one", listed)
	}
}

// TestZoneUnlinkLinkOwnWindowsSeparators — the classifier-level repair
// round 2: on WINDOWS a "\" separates, so `zonelink\.` and `zonelink\` name
// the directory THROUGH the link (dereference spellings, target-following
// judgment) and a "\child" suffix keeps the link a parent component; on
// POSIX the backslash is an ORDINARY filename character
// (SPEC-HOOK-ZONE-BACKSLASH-001) and the same texts stay link-own finals.
// Repair round 5 adds the Win32 NON-CANONICAL finals: normalization strips
// trailing dots and spaces from the last component, so `secret.md.` /
// `secret.md␣` alias `secret.md` — on windows such a final must NOT route
// to the link-own resolver (which preserves the raw name) and stays on the
// target-following judgment instead, while POSIX keeps the literal filename
// link-own. Repair round 8 closes the aliasing CLASS structurally: the
// narrowing is DISABLED on windows outright — every windows=true cell below
// wants false, and the rounds-2..7 guard cells (separator split, trailing
// dot/space, ~digit shape) stay as BELT-AND-SUSPENDERS documentation of the
// narrowing's intended semantics for any future windows re-enable, not live
// routing. POSIX (windows=false) keeps the narrowing and these cells pin it.
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
		// windows (narrowing disabled, round 8): EVERY operand — plain names
		// included — routes to the target-following full resolution, whose
		// EvalSymlinks/Lstat canonicalization restores the real entry name.
		// The per-mechanism guard cells below document why each spelling was
		// never link-own-safe and stay green as defense-in-depth.
		{cand: `zonelink\.`, windows: true, want: false},
		{cand: `zonelink\`, windows: true, want: false},
		{cand: `zonelink\child`, windows: true, want: false},
		{cand: `C:\dir\link`, windows: true, want: false},
		{cand: `zonelink/child`, windows: true, want: false},
		{cand: `zonelink/`, windows: true, want: false},
		// windows: Win32 strips trailing dots/spaces from the final component
		// — these alias a DIFFERENT real name, so they are not the link's own
		// entry and keep the fail-closed target-following judgment.
		{cand: `secret.md.`, windows: true, want: false},
		{cand: `secret.md `, windows: true, want: false},
		{cand: `dir\secret.md.`, windows: true, want: false},
		// windows: an NTFS 8.3 short-name shape (~N) aliases the long name
		// (SECRET~1.MD removes secret-long-name.md) and the stdlib cannot
		// expand it — never judge it as the link's own entry.
		{cand: `SECRET~1.MD`, windows: true, want: false},
		{cand: `dir\REPORT~2.TXT`, windows: true, want: false},
		// POSIX: the backslash is a filename character — the identical texts
		// stay link-own finals (the backslash-literal SPEC's constraint).
		{cand: `zonelink\.`, windows: false, want: true},
		{cand: `zonelink\`, windows: false, want: true},
		// POSIX: a trailing dot or space is an ORDINARY filename character —
		// `secret.md.` is a real, distinct entry and stays link-own. A "~N"
		// substring is likewise an ordinary name on POSIX — no 8.3 aliasing
		// exists there.
		{cand: `secret.md.`, windows: false, want: true},
		{cand: `secret.md `, windows: false, want: true},
		{cand: `dir/secret.md.`, windows: false, want: true},
		{cand: `SECRET~1.MD`, windows: false, want: true},
		{cand: `dir/REPORT~2.TXT`, windows: false, want: true},
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

// TestZoneResolveWalkWindowsRootedDotDot — repair round 3 P1: on WINDOWS a
// leading "\" ROOTS the path on the current drive, so `\..\<root>\zonelink`
// IS `<root>\zonelink` and the walk must continue at the filesystem root —
// the same root rule the "/" spelling got in repair round 1. Treated as
// relative, the walk fails closed at that ".." (skipped==0), the parent
// symlink is never resolved, and the rooted deletion through it reads allow.
// The windows flag rides in as an argument (the pathSegments idiom); Windows
// end-to-end semantics are not executable on this darwin host — this walk
// unit is the measurable surface, CI's windows matrix owns the guard-level
// observation.
func TestZoneResolveWalkWindowsRootedDotDot(t *testing.T) {
	root := newUnlinkLinkRoot(t)
	// A drive-rooted spelling names the CURRENT drive, and the fixture composes
	// it from the root with the volume split off (card t1893). When the temp
	// root and the process cwd sit on different drives — windows-latest keeps
	// TEMP on C: and the workspace on D: — "\Users\..." resolves on the cwd's
	// drive, where no fixture exists, and every walk below stops at the first
	// missing component. Anchoring the cwd on the root's volume makes the
	// current drive the fixture's drive; a no-op where they already coincide.
	t.Chdir(root)
	windowsPath := zoneRootedDotDotPath(root, "/zonelink", true)
	resolved, ok := zoneResolveWalk(windowsPath, 0, true)
	if !ok {
		t.Fatalf("zoneResolveWalk(%q, windows=true) failed closed — a leading \\ roots the path on the current drive, the walk must continue at the filesystem root", windowsPath)
	}
	if got := filepath.ToSlash(resolved); !strings.HasSuffix(got, "/protected") {
		t.Errorf("zoneResolveWalk(%q, windows=true) = %q, want the rooted parent resolved through the symlink (…/protected)", windowsPath, got)
	}
	// the POSIX control pins the "/" spelling's round-1 rule unchanged
	resolved, ok = zoneResolveWalk(zoneRootedDotDotPath(root, "/zonelink", false), 0, false)
	if !ok || !strings.HasSuffix(filepath.ToSlash(resolved), "/protected") {
		t.Errorf("zoneResolveWalk(/..-rooted, windows=false) = %q ok=%v, want the root-rule resolution …/protected", resolved, ok)
	}
	// repair round 6 P1: a resolved dir symlink adds MANY real components but
	// a SINGLE counter unit, so the walked DEFICIT (alias chain length) can
	// exceed the counter the prefix funded — then some ".." is IGNORED-at-
	// root while real bash is still climbing INSIDE the zone. The chain must
	// be deeper than any host's resolved prefix (darwin temp roots resolve
	// to ~8 components), so the cell composes both from one constant: twelve
	// dirs deep, twelve "..". The buggy counter runs out mid-chain and the
	// walk sticks at a dN dir (missing name → allow) while real bash lands
	// on …/protected and removes secret.md inside the zone. Root-reach must
	// come from the resolved prefix string, and the pop from it — never the
	// counter.
	const aliasDepth = 12
	chain := make([]string, 0, aliasDepth+1)
	chain = append(chain, "protected")
	for i := 0; i < aliasDepth; i++ {
		chain = append(chain, fmt.Sprintf("d%d", i))
	}
	deep := filepath.Join(root, filepath.Join(chain...))
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(chain...), filepath.Join(root, "deepalias")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	final := "/deepalias/" + strings.Repeat("../", aliasDepth) + "secret.md"
	resolved, ok = zoneResolveWalk(zoneRootedDotDotPath(root, final, false), 0, false)
	if !ok || !strings.HasSuffix(filepath.ToSlash(resolved), "/protected/secret.md") {
		t.Errorf("zoneResolveWalk(deep-alias .. chain, windows=false) = %q ok=%v, want the resolved-prefix pop landing on the protected entry …/protected/secret.md — the guard's deny", resolved, ok)
	}
}

// zoneRootedDotDotPath composes the drive-rooted ".." spelling a test walks:
// the fixture root with the volume split off, prefixed "/../" (slash-flipped
// on windows). Repair round 6 P2: splitting the volume FIRST keeps a
// Windows-host root (e.g. "C:\...\001") composing "\..\...\final" — rooted
// on the current drive — instead of splicing the drive letter mid-path after
// the "..", which no Win32 API resolves.
func zoneRootedDotDotPath(root, final string, windows bool) string {
	volume := filepath.VolumeName(root)
	if volume == "" && len(root) >= 2 && root[1] == ':' &&
		(('A' <= root[0] && root[0] <= 'Z') || ('a' <= root[0] && root[0] <= 'z')) {
		// a Windows drive volume is a property of the root STRING, not of
		// the running host — filepath.VolumeName only recognizes the form
		// on a windows GOOS, and this composition must behave identically
		// on every OS
		volume = root[:2]
	}
	remainder := strings.TrimPrefix(strings.TrimPrefix(root[len(volume):], "/"), "\\")
	rooted := "/../" + remainder + final
	if windows {
		return strings.ReplaceAll(rooted, "/", "\\")
	}
	return rooted
}

// TestZoneRootedDotDotPathSplitsVolume — repair round 6 P2: the rooted ".."
// composition strips the fixture root's volume BEFORE prefixing, so a
// Windows-style root composes "\..\Users\...\zonelink" with no drive letter
// after the "..". The volume is a property of the root STRING, not of the
// host, so injecting a synthetic "C:\"-style root exercises the composition
// on every OS.
func TestZoneRootedDotDotPathSplitsVolume(t *testing.T) {
	const syntheticRoot = `C:\Users\goos\T\root001`
	got := zoneRootedDotDotPath(syntheticRoot, "/zonelink", true)
	if want := `\..\Users\goos\T\root001\zonelink`; got != want {
		t.Errorf("zoneRootedDotDotPath(%q, windows=true) = %q, want %q — the volume must split off before the \"..\" prefix (a drive letter after \"\\\\..\\\\\" does not resolve on Windows)", syntheticRoot, got, want)
	}
	posix := zoneRootedDotDotPath(syntheticRoot, "/zonelink", false)
	if strings.Contains(posix, "C:") {
		t.Errorf("zoneRootedDotDotPath(%q, windows=false) = %q, want the volume split off here too", syntheticRoot, posix)
	}
}

// TestProtectedZoneShellCandidateCapCountsUniquePaths — repair round 3 P2:
// the candidate cap bounds UNIQUE candidate paths, not list memberships. One
// command naming the same paths under rm (link-own list) and tee
// (target-following list) lands each path in BOTH lists; counting the lists
// additively denied a set whose deduplicated population is within the cap —
// 2 lists × 2,100 memberships = 4,200 > 4,096 while 2,100 unique paths sit
// comfortably under it. All operands sit outside the zone, so over-cap is
// the only possible deny source and the allow outcome is the assertion.
func TestProtectedZoneShellCandidateCapCountsUniquePaths(t *testing.T) {
	const perSemantics = 2100
	var rmArgs, teeArgs strings.Builder
	for i := 0; i < perSemantics; i++ {
		fmt.Fprintf(&rmArgs, " f%04d.md", i)
		fmt.Fprintf(&teeArgs, " f%04d.md", i)
	}
	command := "rm" + rmArgs.String() + "; tee" + teeArgs.String()
	root := newZoneRoot(t, zoneShippedDoc(ulZoneManifest), "")
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
	if d == DecisionDeny {
		t.Errorf("cap counting: decision=%q reason=%q, want allow — %d unique candidates across both semantics are within the %d cap; only list MEMBERSHIPS (%d) exceed it", d, r, perSemantics, zoneCandidateCap, 2*perSemantics)
	}
}

// TestProtectedZoneShellCandidateCapCountsExactStrings — repair round 4 P2:
// the cap union dedups by EXACT candidate string, never by folded form. On a
// case-sensitive filesystem `f0000.md` and `F0000.md` are DIFFERENT paths, so
// a command carrying more than zoneCandidateCap case-variant candidates is a
// set too large to verify soundly and must deny fail-closed; a folded union
// key collapses the variants onto one key and lets the flood reach
// per-candidate resolution. 2,100 base names in both case spellings, named
// under both semantics: 4,200 exact strings (> the 4,096 cap) folding to
// 2,100 keys. The over-cap deny is the assertion — all operands sit outside
// the zone, so over-cap is the only possible deny source.
func TestProtectedZoneShellCandidateCapCountsExactStrings(t *testing.T) {
	const perSemantics = 2100
	var rmArgs, teeArgs strings.Builder
	for i := 0; i < perSemantics; i++ {
		fmt.Fprintf(&rmArgs, " f%04d.md F%04d.md", i, i)
		fmt.Fprintf(&teeArgs, " f%04d.md F%04d.md", i, i)
	}
	command := "rm" + rmArgs.String() + "; tee" + teeArgs.String()
	root := newZoneRoot(t, zoneShippedDoc(ulZoneManifest), "")
	h := zoneTestHandler(t, root)
	d, r := zoneCall(t, h, "Bash", harnessLearnerIdentity, map[string]any{"command": command})
	if d != DecisionDeny {
		t.Errorf("cap counting: decision=%q reason=%q, want over-cap deny — %d EXACT candidate strings exceed the %d cap (case variants are distinct paths on a case-sensitive filesystem) and must not reach per-candidate resolution", d, r, 2*perSemantics, zoneCandidateCap)
	}
}
