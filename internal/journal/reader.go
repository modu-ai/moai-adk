package journal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Reader parses journal entries from JSONL files.
type Reader struct{}

// NewReader creates a new journal reader.
func NewReader() *Reader {
	return &Reader{}
}

// ReadAll reads all journal entries from a SPEC directory.
func (r *Reader) ReadAll(specDir string) ([]Entry, error) {
	path := filepath.Join(specDir, "journal.jsonl")
	return r.readFile(path)
}

// ReadLast reads the last N entries from a journal file.
func (r *Reader) ReadLast(specDir string, n int) ([]Entry, error) {
	entries, err := r.ReadAll(specDir)
	if err != nil {
		return nil, err
	}
	if len(entries) <= n {
		return entries, nil
	}
	return entries[len(entries)-n:], nil
}

// BuildResumeContext synthesizes a ResumeContext from journal entries.
func (r *Reader) BuildResumeContext(specDir string) (*ResumeContext, error) {
	entries, err := r.ReadAll(specDir)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no journal entries found in %s", specDir)
	}

	ctx := &ResumeContext{Resumable: false}

	// Count sessions
	sessionSet := make(map[string]bool)
	for _, e := range entries {
		if e.SessionID != "" {
			sessionSet[e.SessionID] = true
		}
		if e.SpecID != "" {
			ctx.SpecID = e.SpecID
		}
	}
	ctx.SessionCount = len(sessionSet)

	// Find last session's state
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		switch e.Type {
		case "session_end":
			ctx.LastSessionID = e.SessionID
			ctx.LastPhase = e.Phase
			ctx.LastStatus = e.Status
			ctx.TokensUsed = e.TokensUsed
			if reason, ok := e.Context["reason"]; ok {
				ctx.EndReason = reason
			}
			if e.Status == "interrupted" {
				ctx.Resumable = true
			}
		case "checkpoint":
			if ctx.NextAction == "" {
				if next, ok := e.Context["next_step"]; ok {
					ctx.NextAction = next
				}
			}
			if files, ok := e.Context["files_modified"]; ok && len(ctx.FilesModified) == 0 {
				ctx.FilesModified = append(ctx.FilesModified, files)
			}
		}
		if ctx.LastSessionID != "" && ctx.NextAction != "" {
			break
		}
	}

	// If no session_end found, the last session crashed
	if ctx.LastSessionID == "" {
		for i := len(entries) - 1; i >= 0; i-- {
			if entries[i].Type == "session_start" {
				ctx.LastSessionID = entries[i].SessionID
				ctx.LastPhase = entries[i].Phase
				ctx.EndReason = "crash"
				ctx.Resumable = true
				break
			}
		}
	}

	return ctx, nil
}

// FindResumableSPECs scans all SPEC directories for resumable work.
func (r *Reader) FindResumableSPECs(specsDir string) ([]ResumeContext, error) {
	dirs, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("read specs directory: %w", err)
	}

	var results []ResumeContext
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		specDir := filepath.Join(specsDir, d.Name())
		ctx, err := r.BuildResumeContext(specDir)
		if err != nil {
			continue
		}
		if ctx.Resumable {
			results = append(results, *ctx)
		}
	}

	// Sort by most recent first
	sort.Slice(results, func(i, j int) bool {
		return results[i].SessionCount > results[j].SessionCount
	})

	return results, nil
}

func (r *Reader) readFile(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open journal file: %w", err)
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue // skip malformed lines
		}
		entries = append(entries, e)
	}
	return entries, scanner.Err()
}
