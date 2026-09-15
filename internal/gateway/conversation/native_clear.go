package conversation

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/gateway/receipt"
)

// ValidateNativeClear accepts only a completed native /clear command in the
// launcher's private profile. Request text alone never authorizes a new UUID.
// It does not import receipts or attach the new session to the old model thread.
func ValidateNativeClear(ctx context.Context, config, cwd, id string) error {
	_, idErr := receipt.New(id)
	configInfo, configErr := os.Lstat(config)
	if idErr != nil || !filepath.IsAbs(cwd) || configErr != nil || !configInfo.IsDir() || configInfo.Mode().Perm()&0077 != 0 || configInfo.Mode()&os.ModeSymlink != 0 {
		return ErrInvalid
	}
	root, err := os.OpenRoot(config)
	if err != nil {
		return ErrInvalid
	}
	defer func() { _ = root.Close() }()
	info, err := root.Lstat("projects")
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ErrInvalid
	}
	projects, err := root.Open("projects")
	if err != nil {
		return ErrInvalid
	}
	defer func() { _ = projects.Close() }()
	dirs, err := projects.ReadDir(257)
	if err != nil && err != io.EOF || len(dirs) > 256 {
		return ErrInvalid
	}
	candidate := ""
	var candidateInfo os.FileInfo
	for _, dir := range dirs {
		if !dir.IsDir() || dir.Type()&os.ModeSymlink != 0 {
			continue
		}
		path := filepath.Join("projects", dir.Name(), id+".jsonl")
		info, e := root.Lstat(path)
		if os.IsNotExist(e) {
			continue
		}
		if e != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || candidate != "" {
			return ErrInvalid
		}
		candidate = path
		candidateInfo = info
	}
	if candidate == "" {
		return ErrMissing
	}
	f, err := root.Open(candidate)
	if err != nil {
		return ErrInvalid
	}
	defer func() { _ = f.Close() }()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(candidateInfo, opened) {
		return ErrInvalid
	}
	// The clear marker is at the beginning; do not scan growing conversation bodies.
	decoder := json.NewDecoder(io.LimitReader(f, 64<<10))
	command := ""
	for n := 0; n < 32; n++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var row map[string]any
		if decoder.Decode(&row) != nil {
			return ErrInvalid
		}
		typ, _ := row["type"].(string)
		if sid, ok := row["sessionId"]; ok && sid != id {
			return ErrInvalid
		}
		if typ != "user" && typ != "system" {
			continue
		}
		rowCWD, _ := row["cwd"].(string)
		if row["sessionId"] != id || !nativeClearSameDirectory(rowCWD, cwd) || row["isSidechain"] != false {
			return ErrInvalid
		}
		if typ == "system" && command != "" && row["subtype"] == "local_command" && row["parentUuid"] == command && row["content"] == "<local-command-stdout></local-command-stdout>" {
			return nil
		}
		if typ != "user" {
			continue
		}
		if !nativeLocalUserRow(row) {
			message, _ := row["message"].(map[string]any)
			content, _ := message["content"].(string)
			if strings.HasPrefix(content, "<local-command-caveat>") && strings.HasSuffix(content, "</local-command-caveat>") {
				continue
			}
			return ErrInvalid
		}
		message, _ := row["message"].(map[string]any)
		content, _ := message["content"].(string)
		if strings.Join(strings.Fields(content), " ") == "<command-name>/clear</command-name> <command-message>clear</command-message> <command-args></command-args>" {
			command, _ = row["uuid"].(string)
		}
	}
	return ErrInvalid
}

func nativeClearSameDirectory(a, b string) bool {
	if a == b {
		return true
	}
	ai, ae := os.Stat(a)
	bi, be := os.Stat(b)
	return ae == nil && be == nil && ai.IsDir() && bi.IsDir() && os.SameFile(ai, bi)
}
