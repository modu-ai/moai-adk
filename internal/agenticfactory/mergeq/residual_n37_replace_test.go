package mergeq

// N37 second bullet (audit db0a69698a, residual classification): `git replace c1 c2` makes every git command that honours
// replace refs read c2 where it reads c1. The queue then lands c1 (the commit the card's evidence is about) with the content
// of c2. It needs write access to the repository's refs, which a lander already has by definition; the cure is one
// environment variable on the git commands of the merger (GIT_NO_REPLACE_OBJECTS). The expectation: what lands under the name
// of c1 is c1's content.

import (
	"context"
	"strings"
	"testing"
)

func TestResidualN37_ALandingIgnoresAReplaceRefOfTheCandidate(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	c1 := candidate(t, dir, "c1", "a.txt", "a\n")
	c2 := candidate(t, dir, "c2", "evil.txt", "evil\n")
	git(t, dir, "replace", c1, c2)
	s := queuedCard(t, "t1", c1)
	res, err := realMerger(t, dir, s).Land(ctx, "t1", c1)
	files := strings.Fields(git(t, dir, "ls-tree", "-r", "--name-only", "refs/heads/trunk"))
	t.Logf("land: outcome=%q err=%v; files of the trunk: %v", res.Outcome, err, files)
	for _, f := range files {
		if f == "evil.txt" {
			t.Errorf("the commit that landed under the name of %.8s carries the content of its replacement: trunk files %v", c1, files)
		}
	}
}
