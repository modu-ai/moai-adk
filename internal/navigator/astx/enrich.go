package astx

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"log/slog"
)

// Provenance stamps an enrichment run to a git baseline (REQ-NT-009 /
// REQ-NT-012 idempotence): ExtractCommitSHA is `git rev-parse HEAD` and
// CapturedAt is the committer date of that SHA (`git log -1 --format=%cI`).
// No wall-clock timestamp is used, so two runs on the same HEAD produce
// byte-identical output.
type Provenance struct {
	ExtractCommitSHA string
	CapturedAt       string
}

// EnrichedRow is the stable output schema row (design.md §5.2). Fields are
// forward-compatible (additive only).
type EnrichedRow struct {
	SpecID             string   `json:"spec_id"`
	Title              string   `json:"title"`
	ImplementationPath string   `json:"implementation_path"`
	OnDiskVerified     bool     `json:"on_disk_verified"`
	ExtractLanguage    string   `json:"extract_language"`
	PrimaryFiles       []string `json:"primary_files"`
	PrimarySymbols     []Symbol `json:"primary_symbols"`
	SymbolCount        int      `json:"symbol_count"`
	Truncated          bool     `json:"truncated"`
	Supported          bool     `json:"supported"`
}

// EnrichOptions configures an enrichment run.
type EnrichOptions struct {
	// ProjectRoot resolves the capability-map path and implementation paths.
	ProjectRoot string
	// CapabilityMapPath is the path to 001's capability-map.md (absolute or
	// relative to ProjectRoot).
	CapabilityMapPath string
	// MaxFilesPerPath caps the file walk per implementation-path (REQ-NT-014).
	// Default 2000 when zero.
	MaxFilesPerPath int
	// PrimaryFilesN is the top-N files by symbol count (default 5).
	PrimaryFilesN int
	// PrimarySymbolsN is the top-N symbols by frequency (default 10).
	PrimarySymbolsN int
}

// EnrichResult holds the run output: provenance + enriched rows.
type EnrichResult struct {
	Provenance Provenance
	Rows       []EnrichedRow
}

// Default enrichment knobs (plan.md §I open questions, frozen by this SPEC).
const (
	defaultMaxFilesPerPath = 2000
	defaultPrimaryFilesN   = 5
	defaultPrimarySymbolsN = 10
)

// provenanceUnknown is the fail-open placeholder for git failures.
const provenanceUnknown = "<unknown>"

// fileAgg accumulates the extraction result for one walked file.
type fileAgg struct {
	path    string
	lang    string
	symbols map[string][]Symbol
	count   int
}

// CurrentProvenance returns the git provenance for the working-tree HEAD.
// Fail-open: on any git error, returns "<unknown>" values (never aborts).
func CurrentProvenance(projectRoot string) Provenance {
	sha := gitOutput(projectRoot, "rev-parse", "HEAD")
	if sha == "" {
		sha = provenanceUnknown
	}
	captured := gitOutput(projectRoot, "log", "-1", "--format=%cI", sha)
	if captured == "" {
		// Fall back to a HEAD-relative lookup when the explicit SHA log failed
		// (e.g. shallow-clone edge where the SHA object is unavailable).
		captured = gitOutput(projectRoot, "log", "-1", "--format=%cI")
	}
	if captured == "" {
		captured = provenanceUnknown
	}
	return Provenance{ExtractCommitSHA: sha, CapturedAt: captured}
}

// gitOutput runs a git command in dir (empty dir = inherit cwd) and returns
// the trimmed stdout. Any error → "" (fail-open, logged at debug).
func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		slog.Debug("astx: git command failed", "args", args, "dir", dir, "error", err)
		return ""
	}
	return strings.TrimSpace(out.String())
}

// capRow is one parsed capability-map.md data row keyed by header name.
type capRow map[string]string

// parseCapabilityMap reads capability-map.md content and returns the data
// rows as header→value maps (header-driven join per REQ-NT-011 — robust to
// column reordering). Returns an empty slice when the file has no table.
func parseCapabilityMap(content []byte) []capRow {
	var headers []string
	var rows []capRow
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := splitTableRow(line)
		if isSeparatorRow(cells) {
			continue
		}
		if headers == nil {
			headers = normalizeHeaders(cells)
			continue
		}
		row := capRow{}
		for i, c := range cells {
			if i < len(headers) {
				row[headers[i]] = c
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// splitTableRow splits a markdown table row into trimmed cell values,
// stripping the leading/trailing pipes.
func splitTableRow(line string) []string {
	inner := strings.TrimPrefix(line, "|")
	inner = strings.TrimSuffix(inner, "|")
	parts := strings.Split(inner, "|")
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.TrimSpace(p)
	}
	return out
}

// isSeparatorRow reports whether all cells are the markdown table separator
// pattern (dashes/colons, e.g. "---", ":--:", "----").
func isSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		if !strings.Contains(c, "-") {
			return false
		}
		if strings.Trim(c, " :-") != "" {
			return false
		}
	}
	return true
}

// normalizeHeaders maps capability-map header variants to canonical keys.
// Accepts "spec-id"/"owning-spec", "implementation-path"/"path", "title"/"capability".
func normalizeHeaders(cells []string) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = normalizeHeader(c)
	}
	return out
}

func normalizeHeader(h string) string {
	switch strings.ToLower(h) {
	case "spec-id", "owning-spec", "spec":
		return "spec-id"
	case "implementation-path", "path", "module":
		return "implementation-path"
	case "title", "capability", "feature":
		return "title"
	default:
		return strings.ToLower(h)
	}
}

// EnrichRows reads 001's capability-map.md (header-driven join per
// REQ-NT-011), walks each row's implementation-path, extracts symbols, and
// aggregates them into EnrichedRows. It never aborts on a single row or
// file failure (fail-open); rows whose path is absent carry OnDiskVerified=false.
func EnrichRows(opts EnrichOptions) (EnrichResult, error) {
	if opts.MaxFilesPerPath == 0 {
		opts.MaxFilesPerPath = defaultMaxFilesPerPath
	}
	if opts.PrimaryFilesN == 0 {
		opts.PrimaryFilesN = defaultPrimaryFilesN
	}
	if opts.PrimarySymbolsN == 0 {
		opts.PrimarySymbolsN = defaultPrimarySymbolsN
	}

	capPath := opts.CapabilityMapPath
	if !filepath.IsAbs(capPath) && opts.ProjectRoot != "" {
		capPath = filepath.Join(opts.ProjectRoot, capPath)
	}
	content, err := os.ReadFile(capPath)
	if err != nil {
		// Fail-open: absent/unreadable capability-map → empty rows (REQ-NT-002).
		slog.Debug("astx: capability-map read error", "path", capPath, "error", err)
		return EnrichResult{Provenance: CurrentProvenance(opts.ProjectRoot)}, nil
	}

	rows := parseCapabilityMap(content)
	out := make([]EnrichedRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, enrichOneRow(opts, r))
	}
	return EnrichResult{
		Provenance: CurrentProvenance(opts.ProjectRoot),
		Rows:       out,
	}, nil
}

// enrichOneRow resolves a capability row's implementation-path, walks it, and
// aggregates the extracted symbols into an EnrichedRow.
func enrichOneRow(opts EnrichOptions, r capRow) EnrichedRow {
	row := EnrichedRow{
		SpecID:             r["spec-id"],
		Title:              r["title"],
		ImplementationPath: r["implementation-path"],
		PrimaryFiles:       []string{},
		PrimarySymbols:     []Symbol{},
	}
	if row.ImplementationPath == "" {
		return row
	}

	resolved := row.ImplementationPath
	if !filepath.IsAbs(resolved) && opts.ProjectRoot != "" {
		resolved = filepath.Join(opts.ProjectRoot, resolved)
	}
	info, err := os.Stat(resolved)
	if err != nil || info == nil {
		// Path absent → on_disk_verified false, supported stays false.
		slog.Debug("astx: implementation-path absent", "path", resolved, "error", err)
		return row
	}

	// Walk and extract. The ceiling bounds how many files are parsed; once it
	// is hit the walk stops and truncated is set (REQ-NT-014).
	var aggs []fileAgg
	langFileCount := map[string]int{}
	totalFiles := 0
	truncated := false

	_ = filepath.WalkDir(resolved, func(path string, d os.DirEntry, werr error) error {
		if werr != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return nil
		}
		if totalFiles >= opts.MaxFilesPerPath {
			truncated = true
			return filepath.SkipAll
		}
		lang := DetectLanguage(d.Name())
		if lang == "" || IsScaffolded(lang) {
			return nil
		}
		set, _ := Extract(lang, path)
		if !set.Supported {
			return nil
		}
		totalFiles++
		langFileCount[lang]++
		flat := 0
		for _, syms := range set.Symbols {
			flat += len(syms)
		}
		aggs = append(aggs, fileAgg{path: path, lang: lang, symbols: set.Symbols, count: flat})
		return nil
	})

	row.Truncated = truncated
	row.OnDiskVerified = len(aggs) > 0
	row.Supported = row.OnDiskVerified
	row.ExtractLanguage = dominantLanguage(langFileCount)
	row.PrimaryFiles = topFiles(aggs, opts.PrimaryFilesN)
	row.PrimarySymbols, row.SymbolCount = aggregateSymbols(aggs, opts.PrimarySymbolsN)

	// Make paths relative to the project root for stable output.
	if opts.ProjectRoot != "" {
		for i, p := range row.PrimaryFiles {
			if rel, err := filepath.Rel(opts.ProjectRoot, p); err == nil {
				row.PrimaryFiles[i] = rel
			}
		}
	}
	return row
}

// dominantLanguage returns the language with the highest file count (ties
// broken by registration-table order for determinism), or "".
func dominantLanguage(counts map[string]int) string {
	best := ""
	bestN := 0
	for _, m := range supportedLanguages {
		if n := counts[m.Name]; n > bestN {
			bestN = n
			best = m.Name
		}
	}
	return best
}

// topFiles returns the top-N file paths by symbol count (desc, stable ties).
func topFiles(aggs []fileAgg, n int) []string {
	sorted := make([]fileAgg, len(aggs))
	copy(sorted, aggs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].count != sorted[j].count {
			return sorted[i].count > sorted[j].count
		}
		return sorted[i].path < sorted[j].path
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, sorted[i].path)
	}
	return out
}

// aggregateSymbols returns the top-N symbols by frequency across files and
// the total deduplicated symbol count.
func aggregateSymbols(aggs []fileAgg, n int) ([]Symbol, int) {
	type freqSym struct {
		sym Symbol
		n   int
	}
	freq := map[string]*freqSym{}
	for _, a := range aggs {
		for _, syms := range a.symbols {
			for _, s := range syms {
				key := s.Kind + ":" + s.Name
				if existing, ok := freq[key]; ok {
					existing.n++
				} else {
					freq[key] = &freqSym{sym: s, n: 1}
				}
			}
		}
	}
	list := make([]*freqSym, 0, len(freq))
	for _, v := range freq {
		list = append(list, v)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].n != list[j].n {
			return list[i].n > list[j].n
		}
		ki := list[i].sym.Kind + ":" + list[i].sym.Name
		kj := list[j].sym.Kind + ":" + list[j].sym.Name
		return ki < kj
	})
	if n > len(list) {
		n = len(list)
	}
	out := make([]Symbol, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, list[i].sym)
	}
	return out, len(freq)
}

// RenderMarkdown renders the enriched rows to the human-readable
// capability-symbols.md format (design.md §5.1).
func RenderMarkdown(p Provenance, sourceMap string, rows []EnrichedRow) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# Capability Symbols (AST-derived)\n\n")
	fmt.Fprintf(&b, "Extracted at: %s\n", p.CapturedAt)
	fmt.Fprintf(&b, "Extract commit: %s\n", p.ExtractCommitSHA)
	fmt.Fprintf(&b, "Source capability-map: %s\n\n", sourceMap)
	fmt.Fprintf(&b, "| spec-id | title | implementation-path | on-disk | files | symbols | primary-symbols |\n")
	fmt.Fprintf(&b, "|---------|-------|---------------------|---------|-------|---------|-----------------|\n")
	for _, r := range rows {
		mark := "✗"
		if r.OnDiskVerified {
			mark = "✓"
		}
		prim := symbolNames(r.PrimarySymbols)
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %d | %d | %s |\n",
			r.SpecID, r.Title, r.ImplementationPath, mark,
			len(r.PrimaryFiles), r.SymbolCount, strings.Join(prim, ", "))
	}
	return b.Bytes()
}

// MarshalCapabilitySymbolsJSON renders the machine-readable JSON envelope
// (design.md §5.2) deterministically.
func MarshalCapabilitySymbolsJSON(p Provenance, sourceMap string, rows []EnrichedRow) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "{\n")
	fmt.Fprintf(&b, "  \"extracted_at\": %s,\n", jsonStr(p.CapturedAt))
	fmt.Fprintf(&b, "  \"extract_commit\": %s,\n", jsonStr(p.ExtractCommitSHA))
	fmt.Fprintf(&b, "  \"source_capability_map\": %s,\n", jsonStr(sourceMap))
	fmt.Fprintf(&b, "  \"rows\": [")
	if len(rows) == 0 {
		fmt.Fprintf(&b, "]\n}\n")
		return b.Bytes()
	}
	fmt.Fprintf(&b, "\n")
	for i, r := range rows {
		fmt.Fprintf(&b, "    {\n")
		fmt.Fprintf(&b, "      \"spec_id\": %s,\n", jsonStr(r.SpecID))
		fmt.Fprintf(&b, "      \"title\": %s,\n", jsonStr(r.Title))
		fmt.Fprintf(&b, "      \"implementation_path\": %s,\n", jsonStr(r.ImplementationPath))
		fmt.Fprintf(&b, "      \"on_disk_verified\": %v,\n", r.OnDiskVerified)
		fmt.Fprintf(&b, "      \"extract_language\": %s,\n", jsonStr(r.ExtractLanguage))
		fmt.Fprintf(&b, "      \"primary_files\": [")
		for j, f := range r.PrimaryFiles {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(jsonStr(f))
		}
		fmt.Fprintf(&b, "],\n")
		fmt.Fprintf(&b, "      \"primary_symbols\": [")
		for j, s := range r.PrimarySymbols {
			if j > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "{\"name\": %s, \"kind\": %s, \"file\": %s, \"line\": %d}",
				jsonStr(s.Name), jsonStr(s.Kind), jsonStr(s.File), s.Line)
		}
		fmt.Fprintf(&b, "],\n")
		fmt.Fprintf(&b, "      \"symbol_count\": %d,\n", r.SymbolCount)
		fmt.Fprintf(&b, "      \"truncated\": %v,\n", r.Truncated)
		fmt.Fprintf(&b, "      \"supported\": %v\n", r.Supported)
		if i == len(rows)-1 {
			fmt.Fprintf(&b, "    }\n")
		} else {
			fmt.Fprintf(&b, "    },\n")
		}
	}
	fmt.Fprintf(&b, "  ]\n}\n")
	return b.Bytes()
}

func symbolNames(syms []Symbol) []string {
	out := make([]string, 0, len(syms))
	for _, s := range syms {
		out = append(out, s.Name)
	}
	return out
}

// jsonStr renders s as a JSON string literal.
func jsonStr(s string) string {
	var b bytes.Buffer
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
