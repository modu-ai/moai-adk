package update

// preview.go establishes the M3c preview entry point (SPEC-CLI-TUX-V3-003).
//
// It is the single convergence surface for the two `merge.ConfirmMerge` call
// sites in update.go (plan Known Issue #7) and routes BOTH the Bubble Tea v2
// TUI table+viewport (AC-TUX3-008/009) and the plain-text fallback
// (AC-TUX3-010) through the SAME classification model (REQ-TUX3-001/002 — no
// parallel heuristic, plan.md §G anti-pattern defense).
//
// The neutral FilePreviewInput type decouples the preview from merge.FileAnalysis
// so M3d's package move does not ripple into the preview layer: the caller maps
// analysis.Files → []FilePreviewInput at the call site.

// FilePreviewInput is the neutral per-file input the preview consumes. It is
// deliberately decoupled from merge.FileAnalysis (which has no Exists/Conflict
// fields) so M3d's decomposition can move the merge types without forcing a
// preview-layer rewrite. The caller (update.go's two ConfirmMerge sites) maps
// its analysis result into this type at the call site, deriving Exists/Conflict
// from the same signals the deploy stage will later enforce.
type FilePreviewInput struct {
	RelPath  string
	Exists   bool
	Conflict bool
	Diff     string
}

// PreviewOptions governs TUI-vs-fallback selection and rendering dimensions.
type PreviewOptions struct {
	// Interactive selects the TUI path when true; false selects the text
	// fallback (the --yes / non-TTY path, AC-TUX3-010).
	Interactive bool
	// NoColor requests ANSI-free output. The fallback is structurally
	// color-free regardless; this field documents the caller's intent for
	// future TUI color policy and is asserted by AC-TUX3-010 fallback tests.
	NoColor bool
	// Width is the terminal width in columns (TUI table sizing).
	Width int
	// Height is the terminal height in rows (viewport sizing).
	Height int
}

// classifyAll maps a slice of FilePreviewInput into a slice of FileClassification
// via the single Classify entry point (REQ-TUX3-001/002 — one source of truth).
// The preview TUI and the text fallback both consume the output of this
// function; neither re-derives classification.
func classifyAll(in []FilePreviewInput, isUserOwned UserOwnedPredicate) []FileClassification {
	out := make([]FileClassification, 0, len(in))
	for _, f := range in {
		out = append(out, FileClassification{
			RelPath: f.RelPath,
			Class:   Classify(f.RelPath, f.Exists, f.Conflict, isUserOwned),
		})
	}
	return out
}

// PreviewClassification is the single preview entry point consumed by BOTH
// update.go ConfirmMerge call sites (plan Known Issue #7 convergence).
//
// When opts.Interactive is true, it runs the Bubble Tea v2 TUI (table +
// viewport) and blocks until the user confirms or cancels. When false
// (--yes or non-TTY), it renders the plain-text classification summary to
// stdout and returns confirmed=true without launching a TUI (AC-TUX3-010).
//
// Both paths classify through update.Classify (REQ-TUX3-001/002 single source
// of truth).
//
// @MX:ANCHOR: [AUTO] single preview entry point converging both ConfirmMerge call sites
// @MX:REASON: REQ-TUX3-008/009/010 + plan Known Issue #7 — fan_in = 2 (update.go:677, :1038); the deploy-stage enforcement coherence (REQ-TUX3-016) shares the same Classify source
func PreviewClassification(in []FilePreviewInput, isUserOwned UserOwnedPredicate, opts PreviewOptions) (bool, error) {
	if !opts.Interactive {
		classes := classifyAll(in, isUserOwned)
		return renderFallbackToStdout(classes, opts), nil
	}
	return runPreviewTUI(in, isUserOwned, opts)
}
