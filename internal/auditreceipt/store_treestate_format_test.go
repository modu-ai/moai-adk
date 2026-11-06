package auditreceipt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// The head-bound digest keeps the format it always had: the stream starts with the
// head line. Recomputed here independently from the same git reads.
func TestTreeState_DigestFormatIsUnchanged(t *testing.T) {
	root := initContentRepo(t)
	head, digest, err := TreeState(root)
	if err != nil {
		t.Fatalf("TreeState: %v", err)
	}
	headOut, _ := runGit(t, root, "rev-parse", "HEAD")
	statusOut, _ := gitInRaw(root, "status", "--porcelain", "-z")
	diffOut, _ := runGit(t, root, "diff", "HEAD")
	sum := sha256.Sum256([]byte(fmt.Sprintf("head:%s\x00status\x00%s\x00diff\x00%s", headOut, statusOut, diffOut)))
	if want := hex.EncodeToString(sum[:]); digest != want || head != headOut {
		t.Errorf("TreeState = (%s, %s), want (%s, %s)", head, digest, headOut, want)
	}
}
