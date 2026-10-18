package hook

// protected_zone_identity.go — card t1804 (sync-audit rounds 1-2): an identity
// backstop for the protected-zone path judgment. The name-based forms
// (lexical, physical, listing-canonical) all depend on spelling: a symlinked
// parent, a hard link, or a directory that cannot be listed each produced a
// shape where none of them names the zone. This layer asks the OS instead:
// every prefix of an already-found form is stat'ed (symlinks followed) and
// compared with os.SameFile against the manifest's own zone entries, and a
// match ADDS a form made of the zone entry's spelling plus the rest of the
// operand. It needs no listing and no case or normalization fold.
//
// Additive only: forms are appended next to the ones found, never replaced or
// reordered, and any stat failure skips that prefix — the layer can only add
// deny forms, never widen what the name-based forms allow.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// The stat seams: tests count calls and inject failures.
var (
	zoneIdentityStat  = os.Stat
	zoneIdentityLstat = os.Lstat
)

// zoneIdentityEntry is one concrete manifest entry (directory or exact path)
// with the file identity it names on this machine.
type zoneIdentityEntry struct {
	spelling string      // the entry's own spelling, no trailing slash
	followed os.FileInfo // Stat: symlinks followed
	own      os.FileInfo // Lstat: the entry itself
}

// zoneIdentity is the per-guard-call identity state: the zone entries are
// stat'ed once, and so is every operand prefix, however many candidates share
// it. Never carried across guard calls.
type zoneIdentity struct {
	root     string
	zone     config.ProtectedZone
	ready    bool
	realRoot string
	entries  []zoneIdentityEntry
	followed map[string]os.FileInfo // nil value: the stat failed
	own      map[string]os.FileInfo
}

func newZoneIdentity(root string, zone config.ProtectedZone) *zoneIdentity {
	return &zoneIdentity{
		root:     root,
		zone:     zone,
		followed: make(map[string]os.FileInfo),
		own:      make(map[string]os.FileInfo),
	}
}

// init resolves the real root and stats the concrete zone entries on first
// use. An entry that does not exist or cannot be stat'ed is skipped.
func (zi *zoneIdentity) init() {
	if zi.ready {
		return
	}
	zi.ready = true
	nativeRoot := filepath.FromSlash(zoneNativeSlash(zi.root))
	if zi.root == "" || !filepath.IsAbs(nativeRoot) {
		return
	}
	zi.realRoot = nativeRoot
	if real, ok := zoneResolve(nativeRoot); ok {
		zi.realRoot = real
	}
	seen := make(map[string]bool)
	for _, e := range zi.zone.Entries {
		if e.Kind != config.ZoneDir && e.Kind != config.ZoneExact {
			continue
		}
		spelling := strings.TrimSuffix(zoneSlash(e.Raw), "/")
		if spelling == "" || seen[spelling] || strings.ContainsAny(spelling, "*") || zoneIsAbs(spelling) {
			continue
		}
		bad := false
		for _, part := range strings.Split(spelling, "/") {
			if part == "" || part == "." || part == ".." {
				bad = true
			}
		}
		if bad {
			continue
		}
		seen[spelling] = true
		full := filepath.Join(zi.realRoot, filepath.FromSlash(spelling))
		followed, err := zoneIdentityStat(full)
		if err != nil {
			continue
		}
		own, err := zoneIdentityLstat(full)
		if err != nil {
			continue
		}
		zi.entries = append(zi.entries, zoneIdentityEntry{spelling: spelling, followed: followed, own: own})
	}
}

func (zi *zoneIdentity) statFollowed(rel string) os.FileInfo {
	if fi, hit := zi.followed[rel]; hit {
		return fi
	}
	fi, err := zoneIdentityStat(filepath.Join(zi.realRoot, filepath.FromSlash(rel)))
	if err != nil {
		fi = nil
	}
	zi.followed[rel] = fi
	return fi
}

func (zi *zoneIdentity) statOwn(rel string) os.FileInfo {
	if fi, hit := zi.own[rel]; hit {
		return fi
	}
	fi, err := zoneIdentityLstat(filepath.Join(zi.realRoot, filepath.FromSlash(rel)))
	if err != nil {
		fi = nil
	}
	zi.own[rel] = fi
	return fi
}

// add appends the identity forms of every prefix of every form in forms and
// returns the extended slice; the input forms keep their order and content.
// linkOwn is the unlink-family reading (card t1712): the operand's FINAL
// component is the link itself and is never followed, so it is compared by its
// own identity only; every proper prefix is a parent the shell walks through
// and is followed.
func (zi *zoneIdentity) add(forms []zoneForm, linkOwn bool) []zoneForm {
	if len(forms) == 0 || len(zi.zone.Entries) == 0 {
		return forms
	}
	zi.init()
	if len(zi.entries) == 0 {
		return forms
	}
	out := forms
	for _, f := range forms {
		parts := strings.Split(f.Display, "/")
		usable := true
		for _, part := range parts {
			if part == "" || part == "." || part == ".." {
				usable = false
			}
		}
		if !usable {
			continue
		}
		for i := 1; i <= len(parts); i++ {
			prefix := strings.Join(parts[:i], "/")
			final := i == len(parts)
			var followed, own os.FileInfo
			if !final {
				if followed = zi.statFollowed(prefix); followed == nil {
					break // nothing below an unreachable prefix is reachable either
				}
			} else {
				if !linkOwn {
					followed = zi.statFollowed(prefix)
				}
				own = zi.statOwn(prefix)
				if own == nil && followed == nil {
					break
				}
			}
			for _, e := range zi.entries {
				same := followed != nil && os.SameFile(followed, e.followed)
				if final && own != nil && os.SameFile(own, e.own) {
					same = true
				}
				if !same {
					continue
				}
				form := e.spelling
				if i < len(parts) {
					form += "/" + strings.Join(parts[i:], "/")
				}
				out = appendZoneForm(out, form)
			}
		}
	}
	return out
}
