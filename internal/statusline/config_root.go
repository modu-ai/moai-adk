package statusline

import (
	"bytes"
	"encoding/json"
	"io"
)

// ResolveConfigRoot reports the project root a render reads its configuration
// from, and returns a reader replaying the payload it consumed.
//
// A render has TWO levers in .moai/config/sections/statusline.yaml:
// `statusline.segments` / `theme`, read by the CLI entry point, and
// `statusline.forge`, read by forgeOverride on the board root. Until this seam
// existed they resolved their roots independently — segments and theme from a
// cwd walk-up, forge from the state anchor — so a session sitting in a linked
// worktree read one file's two keys out of two different trees, and the
// worktree's own `forge:` was never consulted. Both levers now resolve here.
//
// The rule is that CONFIG FOLLOWS STATE: the statusline already lands its own
// state (the context-usage snapshot, the board caches, the goal read) under
// the anchor, so the configuration governing that render has to name the same
// tree or the render is reading one project while writing another.
//
// fallback is the caller's own root — internal/cli's findProjectRoot cwd
// walk-up — and is used only when the anchor resolves to nothing. That is not
// a divergence: an empty anchor is REQ-SA-003's "no project" verdict, and every
// state consumer SKIPS rather than reading some other tree, so there is no
// second root to disagree with. Falling back keeps statusline.yaml working for
// a .moai project that is not a git repository, which a bare anchor would
// otherwise withdraw it from.
//
// The payload is taken as a reader, and read the same way Build reads it — a
// json.Decoder, which stops at the end of the JSON value and never waits for
// EOF. Reading the whole stream instead would hang a render whose writer keeps
// the pipe open. Whatever the decoder buffered is replayed ahead of the
// unread remainder, so the returned reader is equivalent to the original and
// the caller hands it straight to Build.
//
// @MX:ANCHOR: statusline config-root seam — the segments/theme lever and the forge lever resolve through here
// @MX:REASON: the two levers read the same file; independent roots let one render read two trees (card t957)
func ResolveConfigRoot(r io.Reader, fallback string) (string, io.Reader) {
	if r == nil {
		return fallback, bytes.NewReader(nil)
	}

	var consumed bytes.Buffer
	var input StdinData
	decodeErr := json.NewDecoder(io.TeeReader(r, &consumed)).Decode(&input)
	replay := io.MultiReader(bytes.NewReader(consumed.Bytes()), r)

	if decodeErr != nil {
		// Unparseable or absent payload. Build reaches the same verdict on the
		// replayed bytes and renders without stdin data; the config read keeps
		// the caller's root rather than losing the file entirely.
		return fallback, replay
	}
	if anchor := resolveStateAnchor(&input); anchor != "" {
		return anchor, replay
	}
	return fallback, replay
}
