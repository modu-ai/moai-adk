// ccversion.go — the Claude Code version reads behind injectable seams
// (SPEC-SESSION-CC-VERSION-001 REQ-SCV-001..004).
//
// A running Claude Code process never picks up an updated binary — the update
// takes effect at the next process start — so a long-lived lane can fall
// behind its own leader invisibly. These reads make that visible: the running
// version comes from the process's own binary mapping (never the installed
// binary, an env var, or the registry record), and the installed version comes
// from the version segment of the claude binary found on PATH. Every
// degradation — a dead pid, an unreadable mapping, a path carrying no version
// segment, an unsupported platform — renders UnknownCCVersion and never an
// error (REQ-SCV-003).
//
// The reads follow the procInfoFunc seam pattern (session_pid.go): the
// platform-dependent probe sits behind package-level vars, so tests inject
// fixture pid→mapping tables and fixture PATH resolutions and never spawn a
// process or read the real process table (REQ-SCV-004).
package session

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// claudeBinaryName is the binary the version reads name: the running read
// anchors on the mapping line whose path names this binary, and the installed
// read resolves it on PATH.
const claudeBinaryName = "claude"

// UnknownCCVersion is the value every degraded version read renders. Degraded
// is a value shape, never an error: no read here can fail a command carrying
// it (REQ-SCV-003).
const UnknownCCVersion = "unknown"

// CCVersions carries the two version reads for one session entry's process.
// Either field may hold UnknownCCVersion.
type CCVersions struct {
	Running   string `json:"running"`
	Installed string `json:"installed"`
}

// @MX:ANCHOR: [AUTO] the cross-package version-view seam (session → cli)
// @MX:REASON: public API boundary consumed by both SPEC-SESSION-CC-VERSION-001 surfaces (session list --cc-version, doctor staleness); REQ-SCV-004 requires it to be the injectable dependency — cli tests substitute it wholesale, so no cli test spawns a process or reads the real process table
// ResolveCCVersions reports the version view for the process a registry entry
// names (the entry's PID). Package var: the injectable seam REQ-SCV-004
// requires. The default resolves the running read per pid (one probe per live
// pid; dead pids are not probed) and the installed read per call.
var ResolveCCVersions = func(pid int) CCVersions {
	return CCVersions{
		Running:   runningCCVersion(pid),
		Installed: installedCCVersion(),
	}
}

// The version-read seams. readProcessMapping is the platform probe (defined
// per build-tag file: lsof on darwin, /proc/<pid>/exe on linux, unsupported
// elsewhere); ccInstalledLookPath resolves the claude binary on PATH.
var (
	readProcessMapping = platformReadProcessMapping

	ccInstalledLookPath = func() (string, bool) {
		path, err := exec.LookPath(claudeBinaryName)
		if err != nil {
			return "", false
		}
		return path, true
	}
)

// runningCCVersion resolves the version the given pid runs, from the
// process's own binary mapping. A dead pid is never probed: the liveness gate
// short-circuits before the platform read.
func runningCCVersion(pid int) string {
	if pid <= 0 || !pidIsAlive(pid) {
		return UnknownCCVersion
	}
	mapping, ok := readProcessMapping(pid)
	if !ok {
		return UnknownCCVersion
	}
	if v := runningCCVersionFromMapping(mapping); v != "" {
		return v
	}
	return UnknownCCVersion
}

// installedCCVersion resolves the installed version as the version segment of
// the resolved claude binary found on PATH (the symlink target). The claude
// process is never spawned for this (REQ-SCV-002).
func installedCCVersion() string {
	path, ok := ccInstalledLookPath()
	if !ok {
		return UnknownCCVersion
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return UnknownCCVersion
	}
	if v := versionSegmentFromPath(resolved); v != "" {
		return v
	}
	return UnknownCCVersion
}

// deletedExeSuffix is the Linux kernel's marker on a /proc/<pid>/exe value
// whose binary was deleted on disk — the binary was replaced while the
// process still runs it, which is exactly the staleness case this read
// exists to surface (card-review round 1, P2).
const deletedExeSuffix = " (deleted)"

// @MX:NOTE: the anchor duty — `-d txt` also lists mapped frameworks and
// dylibs, and macOS framework bundles carry Versions/<n>/ directories whose
// version-shaped segments a naive parse would misread (t1348 §2.2).
// @MX:SPEC: SPEC-SESSION-CC-VERSION-001
// runningCCVersionFromMapping parses the process's binary-mapping evidence
// into the running version. The mapping is the platform read: on darwin the
// raw `lsof -a -d txt -p` output (multi-line), on linux the
// /proc/<pid>/exe link target. Only a mapping line naming the claude binary
// itself may satisfy the read — a version-shaped path on a library mapping
// must not — so each line's path field is anchored on the binary name before
// the version segment is parsed from it. A trailing " (deleted)" is stripped
// before both, so a replaced-while-running binary still yields its version.
func runningCCVersionFromMapping(mapping string) string {
	for _, line := range strings.Split(mapping, "\n") {
		line = strings.TrimSuffix(line, deletedExeSuffix)
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		path := fields[len(fields)-1]
		if !mappingPathNamesClaudeBinary(path) {
			continue
		}
		if v := versionSegmentFromPath(path); v != "" {
			return v
		}
	}
	return ""
}

// mappingPathNamesClaudeBinary anchors the parse on the mapping line naming
// the claude binary itself. Three real-world shapes name it:
//
//   - the native installer's product-directory layout …/claude/versions/<X.Y.Z>
//     (the shipped binary file is named by its version — measured live:
//     /Users/<u>/.local/share/claude/versions/2.1.287);
//   - the npm-style …/claude-code/<X.Y.Z> layout;
//   - a path ending in the binary name itself (…/<ver>/claude).
//
// dyld, framework, and dylib mapping lines satisfy none of the three: macOS
// framework bundles capitalize Versions/ (the match is case-sensitive), and
// no system library lives under a claude product directory — so a
// version-shaped segment on a library mapping cannot satisfy the read.
func mappingPathNamesClaudeBinary(path string) bool {
	return strings.Contains(path, "/claude/versions/") ||
		strings.Contains(path, "/claude-code/") ||
		path == claudeBinaryName ||
		strings.HasSuffix(path, "/"+claudeBinaryName)
}

// versionSegmentRe matches the two house install-path shapes:
// versions/<X.Y.Z> (native installer) and claude-code/<X.Y.Z> (npm-style
// layout). The (?:^|/) guard keeps a prefix that merely contains the word
// ("conversions/") from matching.
var versionSegmentRe = regexp.MustCompile(`(?:^|/)(?:versions|claude-code)/([0-9]+(?:\.[0-9]+)*)`)

// versionSegmentFromPath extracts the version segment from an install path,
// or "" when the path carries none — the caller renders unknown, never an
// inferred value.
//
// @MX:DEBT: version segment is the FIRST versions/<n> match, not anchored to the claude install root — /opt/versions/9/tools/claude/versions/2.1.281 reads as "9" (r5 overlay repro), so a stale process can escape the doctor warn
// @MX:CEILING: correct for install-root-shaped paths (~/.local/share/claude/versions, npm claude-code/); exotic prefixes misreport
// @MX:UPGRADE: t1515 — anchor extraction to the claude install root
func versionSegmentFromPath(path string) string {
	m := versionSegmentRe.FindStringSubmatch(path)
	if m == nil {
		return ""
	}
	return m[1]
}
