package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// refreshNative discovers only an exact already-authorized UUID beneath its
// private family namespace. It never searches other Claude profiles or imports
// an unknown UUID. Native files remain the source of public conversation data.
func (m *Manager) refreshNative(ctx context.Context, id string) error {
	m.mu.Lock()
	r, ok := m.entries[id]
	m.mu.Unlock()
	if !ok {
		return ErrMissing
	}
	root := filepath.Join(r.ConfigDir, "projects")
	found := ""
	count := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		count++
		if count > 4096 {
			return ErrInvalid
		}
		if d.Type()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != id+".jsonl" {
			return nil
		}
		if found != "" {
			return ErrAmbiguous
		}
		found = path
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if found == "" {
		return nil
	}
	info, err := os.Lstat(found)
	if err != nil {
		return ErrInvalid
	}
	stamp := info.ModTime().UnixNano()
	if stamp <= 0 {
		return ErrInvalid
	}
	return m.Complete(ctx, id, found, uint64(stamp))
}

func (m *Manager) refreshNativeProject(ctx context.Context, project string) error {
	m.mu.Lock()
	ids := []string{}
	for id, r := range m.entries {
		if r.Project == project {
			ids = append(ids, id)
		}
	}
	m.mu.Unlock()
	for _, id := range ids {
		if err := m.refreshNative(ctx, id); err != nil && err != ErrIncomplete {
			return err
		}
	}
	return nil
}

// transcriptModel validates native identity and completion while obtaining only
// the last successful model ID. Legacy fixture records remain readable.
func transcriptModel(r record) (string, error) {
	if r.Transcript == "" || r.Completion == 0 {
		return "", ErrIncomplete
	}
	p := filepath.Join(r.ConfigDir, filepath.FromSlash(r.Transcript))
	if !within(r.ConfigDir, p) {
		return "", ErrInvalid
	}
	for part := p; ; part = filepath.Dir(part) {
		i, e := os.Lstat(part)
		if e != nil || i.Mode()&os.ModeSymlink != 0 {
			return "", ErrInvalid
		}
		if part == r.ConfigDir {
			break
		}
		if filepath.Dir(part) == part {
			return "", ErrInvalid
		}
	}
	info, err := os.Lstat(p)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxIndexBytes {
		return "", ErrInvalid
	}
	f, err := os.Open(p)
	if err != nil {
		return "", ErrInvalid
	}
	defer f.Close()
	decoder := json.NewDecoder(io.LimitReader(f, maxIndexBytes+1))
	complete, native := false, false
	recoverable := false
	model := ""
	rows := 0
	for {
		var row map[string]any
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil || len(row) == 0 {
			return "", ErrInvalid
		}
		rows++
		if rows > 4096 {
			return "", ErrInvalid
		}
		typ, _ := row["type"].(string)
		sid, hasNative := row["sessionId"].(string)
		if !hasNative {
			if typ == "file-history-snapshot" {
				continue
			}
			if native || row["session_id"] != r.UUID || row["project"] != r.Project {
				return "", ErrInvalid
			}
			if typ == "completed" || row["complete"] == true {
				complete = true
			}
			continue
		}
		native = true
		if sid != r.UUID {
			return "", ErrInvalid
		}
		if legacy, ok := row["session_id"]; ok && legacy != sid {
			return "", ErrInvalid
		}
		if cwd, ok := row["cwd"]; ok && cwd != r.CWD {
			return "", ErrInvalid
		}
		switch typ {
		case "user":
			if row["cwd"] != r.CWD {
				return "", ErrInvalid
			}
			if nativeLocalUserRow(row) {
				continue
			}
			recoverable = complete
			complete = false
		case "assistant":
			if row["cwd"] != r.CWD {
				return "", ErrInvalid
			}
			message, ok := row["message"].(map[string]any)
			if !ok {
				return "", ErrInvalid
			}
			// A native API-error display row is not a model response. Permit
			// retrying only when no actual assistant output has started since
			// the last completed answer; receipt validation remains independent.
			if nativeAPIError(row, message) {
				complete = complete || recoverable
				continue
			}
			complete, recoverable = false, false
			current, ok := message["model"].(string)
			if row["isApiErrorMessage"] == true || !ok || current == "" || current == "<synthetic>" {
				continue
			}
			if message["role"] != "assistant" {
				return "", ErrInvalid
			}
			if message["stop_reason"] == "end_turn" {
				complete = true
				model = current
			}
		}
	}
	if !complete {
		return "", ErrIncomplete
	}
	return model, nil
}

// nativeAPIError recognizes Claude Code's terminal text-only failure display.
// Arbitrary synthetic rows and partially produced model/tool output cannot
// authorize recovery of an incomplete turn.
func nativeAPIError(row, message map[string]any) bool {
	if row["isApiErrorMessage"] != true || message["model"] != "<synthetic>" || message["role"] != "assistant" || message["stop_reason"] != "stop_sequence" {
		return false
	}
	category, ok := row["error"].(string)
	if !ok || category == "" {
		return false
	}
	blocks, ok := message["content"].([]any)
	if !ok || len(blocks) == 0 {
		return false
	}
	for _, value := range blocks {
		block, ok := value.(map[string]any)
		if !ok || block["type"] != "text" {
			return false
		}
		if _, ok := block["text"].(string); !ok {
			return false
		}
	}
	return true
}

// Native local commands are persisted as user rows, but do not start a model
// request. Typed/pasted user input retains its origin or promptSource and must
// never gain this exemption by including command markup in its text.
func nativeLocalUserRow(row map[string]any) bool {
	if _, ok := row["origin"]; ok {
		return false
	}
	if _, ok := row["promptSource"]; ok {
		return false
	}
	message, ok := row["message"].(map[string]any)
	if !ok || message["role"] != "user" {
		return false
	}
	content, ok := message["content"].(string)
	if !ok {
		return false
	}
	if row["isMeta"] == true {
		return true
	}
	if strings.HasPrefix(content, "<local-command-stdout>") && strings.HasSuffix(content, "</local-command-stdout>") {
		return true
	}
	return strings.HasPrefix(content, "<command-name>") && strings.Contains(content, "</command-name>") && strings.Contains(content, "<command-message>") && strings.Contains(content, "</command-message>") && strings.Contains(content, "<command-args>") && strings.HasSuffix(content, "</command-args>")
}
