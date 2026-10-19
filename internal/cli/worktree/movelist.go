package worktree

// movelist.go — SPEC-WORKTREE-HOME-ROOT-001 M1: tree identity and the
// move-list reader (REQ-WHR-012 reader half, REQ-WHR-014 witness;
// design.md D-11).
//
// The immutable identity of a linked worktree is its admin id (the leaf of
// the `gitdir:` pointer in the tree's .git file) plus its instance token,
// written once to `<common-dir>/worktrees/<admin id>/moai-instance` — git's
// administrative directory, never the working tree. HEAD and branch are
// display-only and are never part of the identity. Git reuses an admin id
// once the tree that held it is removed, so the token is what tells two
// trees apart.
//
// The move list lives at <primary>/.moai/state/worktree-migrate/
// move-list.jsonl, one JSON object per line, written by the M6 engine. The
// reader here is consumed from M1: the lookup answers where a tree went
// after a move, and the reverse lookup witnesses that the migration moved
// the tree standing at a path (the M3 cleanup guard consumes it before the
// writer exists). The lookup fixes the expected identity once from the
// requester, follows the old_path -> new_path chain through strictly later
// location entries, and verifies every candidate against registration and
// identity — the latest line naming a path never decides.

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/worktreeroot"
)

// moveListPath is the move-list location under the primary checkout
// (design.md D-11); the primary's state area is the one place every tree of
// the repository can reach.
func moveListPath(primary string) string {
	return filepath.Join(primary, ".moai", "state", "worktree-migrate", "move-list.jsonl")
}

// tokenFileName is the instance-token file inside the tree's git
// administrative directory.
const tokenFileName = "moai-instance"

// TreeIdentity is the immutable part of a linked worktree's identity.
// Token is empty for a tree that was never tokened (made before this SPEC,
// or by another tool).
type TreeIdentity struct {
	AdminID string
	Token   string
}

// Ident renders the identity the move list records: "<admin id>:<token>".
func (i TreeIdentity) Ident() string {
	return i.AdminID + ":" + i.Token
}

// gitDirPointer reads the `gitdir:` pointer out of a tree's .git file.
func gitDirPointer(treePath string) (string, error) {
	body, err := os.ReadFile(filepath.Join(treePath, ".git"))
	if err != nil {
		return "", fmt.Errorf("reading the .git pointer of %s: %w", treePath, err)
	}
	line := strings.TrimSpace(string(body))
	if rest, ok := strings.CutPrefix(line, "gitdir:"); ok {
		p := strings.TrimSpace(rest)
		if p != "" {
			return p, nil
		}
	}
	return "", fmt.Errorf("%s/.git is not a gitdir pointer to a linked worktree", treePath)
}

// TreeIdent reads a linked worktree's identity: the admin id from the .git
// pointer and the instance token from the tree's git administrative
// directory. Read-only: a missing token stays missing (REQ-WHR-012). Paths
// that are not linked worktrees error.
func TreeIdent(treePath string) (TreeIdentity, error) {
	gitdir, err := gitDirPointer(treePath)
	if err != nil {
		return TreeIdentity{}, err
	}
	token := ""
	if body, err := os.ReadFile(filepath.Join(gitdir, tokenFileName)); err == nil {
		token = strings.TrimSpace(string(body))
	}
	return TreeIdentity{AdminID: filepath.Base(gitdir), Token: token}, nil
}

// EnsureInstanceToken returns the tree's identity, writing the instance
// token once when it is absent and never again. The token lands in the
// tree's git administrative directory only — nothing is written into the
// working tree. An unwritable administrative directory fails without
// creating the token.
func EnsureInstanceToken(treePath string) (TreeIdentity, error) {
	gitdir, err := gitDirPointer(treePath)
	if err != nil {
		return TreeIdentity{}, err
	}
	admin := filepath.Base(gitdir)
	tokenFile := filepath.Join(gitdir, tokenFileName)
	if body, err := os.ReadFile(tokenFile); err == nil {
		return TreeIdentity{AdminID: admin, Token: strings.TrimSpace(string(body))}, nil
	}
	token, err := newToken()
	if err != nil {
		return TreeIdentity{}, err
	}
	if err := os.WriteFile(tokenFile, []byte(token+"\n"), 0o644); err != nil {
		return TreeIdentity{}, fmt.Errorf("writing the instance token to %s: %w", tokenFile, err)
	}
	return TreeIdentity{AdminID: admin, Token: token}, nil
}

// newToken returns 16 random bytes as hex — the one-time instance token.
func newToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// MoveRequester is the requester's identity for the move-list lookup: a
// card id, a tree name, an expected ident, or any of them (design D-11).
type MoveRequester struct {
	Card  string
	Tree  string
	Ident string
}

// mlRecord is one parsed move-list line. Only the fields the reader consults
// are carried; the display-only head and branch are ignored here.
type mlRecord struct {
	Card    string `json:"card"`
	Tree    string `json:"tree"`
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	Result  string `json:"result"`
	Ident   string `json:"ident"`
}

// locationKinds are the results that name where a tree went or will go;
// skipped, failed and unparseable lines never enter a decision.
func locationKind(r string) bool {
	switch r {
	case "moved", "restored", "reversed", "pending":
		return true
	}
	return false
}

// loadMoveList parses the move list of primary; a missing file is no
// entries, and every unparseable line is skipped (steps 8).
func loadMoveList(primary string) []mlRecord {
	body, err := os.ReadFile(moveListPath(primary))
	if err != nil {
		return nil
	}
	var out []mlRecord
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec mlRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		out = append(out, rec)
	}
	return out
}

// registeredWorktreePaths returns the paths git registers for primary.
func registeredWorktreePaths(primary string) ([]string, error) {
	out, err := exec.Command("git", "-C", primary, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("listing the worktrees of %s: %w", primary, err)
	}
	recs, err := ParseWorktreeList(string(out))
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(recs))
	for _, r := range recs {
		paths = append(paths, r.Path)
	}
	return paths, nil
}

func isRegistered(paths []string, p string) bool {
	for _, reg := range paths {
		if worktreeroot.Same(reg, p) {
			return true
		}
	}
	return false
}

// requesterMatches reports whether the line belongs to the requester: the
// line's card or tree equals the requester's, and a requester that names
// neither matches every line (design D-11).
func requesterMatches(rec mlRecord, req MoveRequester) bool {
	if req.Card != "" && rec.Card == req.Card {
		return true
	}
	if req.Tree != "" && rec.Tree == req.Tree {
		return true
	}
	return req.Card == "" && req.Tree == ""
}

// carriesE reports whether the line carries the expected identity; every
// line carries it when E is empty.
func carriesE(rec mlRecord, e string) bool {
	return e == "" || rec.Ident == e
}

// @MX:ANCHOR: [AUTO] move-list lookup — every reader of a moved tree's new path resolves here
// @MX:REASON: [AUTO] consumed by launcher cleanup, card readers and the migration engine alike; a second lookup that trusts the latest line would hand a stale pointer another tree (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-012, design D-11)
func MovedTo(primary, oldPath string, requester MoveRequester, cardPath string) (string, bool) {
	entries := loadMoveList(primary)

	// The expected identity E is fixed once, before any candidate is tried:
	// the supplied ident, else the one ident of the requester's own
	// ident-bearing lines naming the held path; two distinct idents mean no
	// answer, none means E is empty.
	e := requester.Ident
	if e == "" {
		seen := map[string]bool{}
		for _, rec := range entries {
			if rec.Ident == "" || !worktreeroot.Same(rec.OldPath, oldPath) || !requesterMatches(rec, requester) {
				continue
			}
			seen[rec.Ident] = true
		}
		if len(seen) > 1 {
			return "", false
		}
		for id := range seen {
			e = id
		}
	}

	candidates := make([]string, 0, 4)
	if cardPath != "" {
		candidates = append(candidates, cardPath)
	}

	// Step 2: the latest location entry of the requester's lines naming the
	// held path starts the chain; the chain follows new_path through
	// strictly later location entries carrying E.
	start := -1
	for i, rec := range entries {
		if !locationKind(rec.Result) || !carriesE(rec, e) || !worktreeroot.Same(rec.OldPath, oldPath) || !requesterMatches(rec, requester) {
			continue
		}
		start = i
	}
	if start >= 0 {
		cur := start
		for {
			rec := entries[cur]
			if rec.Result == "pending" {
				// An unsettled pending leaves both of its paths as
				// candidates (the tree may stand at either).
				candidates = append(candidates, rec.OldPath, rec.NewPath)
			} else {
				candidates = append(candidates, rec.NewPath)
			}
			next := -1
			for i := cur + 1; i < len(entries); i++ {
				en := entries[i]
				if locationKind(en.Result) && carriesE(en, e) && worktreeroot.Same(en.OldPath, rec.NewPath) {
					next = i // latest strictly later entry wins
				}
			}
			if next < 0 {
				break
			}
			cur = next
		}
	}

	if len(candidates) == 0 {
		return "", false
	}
	registrations, err := registeredWorktreePaths(primary)
	if err != nil {
		return "", false
	}
	for _, c := range candidates {
		if !isRegistered(registrations, c) {
			continue
		}
		if e != "" {
			id, err := TreeIdent(c)
			if err != nil || id.Ident() != e {
				continue
			}
			return c, true
		}
		// An empty expected identity accepts a candidate on registration
		// alone only when no ident-bearing line names its path.
		if identBearingLineNames(entries, c) {
			continue
		}
		return c, true
	}
	return "", false
}

// identBearingLineNames reports whether any line carrying an ident names
// path as either of its paths.
func identBearingLineNames(entries []mlRecord, path string) bool {
	for _, rec := range entries {
		if rec.Ident == "" {
			continue
		}
		if worktreeroot.Same(rec.OldPath, path) || worktreeroot.Same(rec.NewPath, path) {
			return true
		}
	}
	return false
}

// @MX:ANCHOR: [AUTO] moved-by-the-migration witness — the cleanup guard refuses on it
// @MX:REASON: [AUTO] a second witness keyed on paths instead of the tree's own ident would let a foreign line switch the protection off (SPEC-WORKTREE-HOME-ROOT-001 REQ-WHR-014, design D-11)
func MovedBy(primary, path string) bool {
	id, err := TreeIdent(path)
	if err != nil || id.Token == "" {
		// A tree with no token was never moved by the migration.
		return false
	}
	entries := loadMoveList(primary)
	latest := -1
	for i, rec := range entries {
		if rec.Ident != id.Ident() || !isLocationLine(rec.Result) {
			continue
		}
		latest = i
	}
	if latest < 0 {
		return false
	}
	rec := entries[latest]
	switch rec.Result {
	case "restored", "reversed":
		// Only a later restored or reversed line releases the protection.
		return false
	case "moved", "pending", "refused":
		if !worktreeroot.Same(rec.NewPath, path) {
			return false
		}
		registrations, err := registeredWorktreePaths(primary)
		if err != nil {
			return false
		}
		return isRegistered(registrations, path)
	}
	return false
}

// isLocationLine names the results that are the tree's location lines for
// the reverse lookup: moved, pending, refused, restored, reversed.
func isLocationLine(r string) bool {
	switch r {
	case "moved", "pending", "refused", "restored", "reversed":
		return true
	}
	return false
}
