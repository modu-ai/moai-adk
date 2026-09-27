package closure

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/modu-ai/moai-adk/internal/closure/gitio"
	"github.com/modu-ai/moai-adk/internal/contract"
)

// PushGitSeam is the git access the push evaluation needs, over the tree the
// push runs in. The production seam is gitio-backed (GitSeam); tests stub it
// or use real repositories.
type PushGitSeam interface {
	ResolveRef(tree, ref string) (string, error)
	NonMergeCommits(tree, from, to string) ([]CommitPaths, error)
	Blob(tree, rev, path string) ([]byte, error)
	PathsUnder(tree, rev, prefix string) ([]string, error)
}

// GitSeam is the gitio-backed production seam.
type GitSeam struct{}

// ResolveRef resolves a ref to its commit.
func (GitSeam) ResolveRef(tree, ref string) (string, error) { return gitio.ResolveRef(tree, ref) }

// NonMergeCommits lists the non-merge commits of a range.
func (GitSeam) NonMergeCommits(tree, from, to string) ([]CommitPaths, error) {
	return gitio.NonMergeCommits(tree, from, to)
}

// Blob reads one file at a rev.
func (GitSeam) Blob(tree, rev, path string) ([]byte, error) { return gitio.Blob(tree, rev, path) }

// PathsUnder lists files under a directory at a rev.
func (GitSeam) PathsUnder(tree, rev, prefix string) ([]string, error) {
	return gitio.PathsUnder(tree, rev, prefix)
}

// PushEvalInput is one push to evaluate: the source commit S, the tree it
// comes from, and the effective policy values.
type PushEvalInput struct {
	// Tree is the directory the push runs in (the -C argument or the tool
	// call's working directory).
	Tree string
	// Remote and Integration name the destination ("origin", "develop").
	Remote      string
	Integration string
	// Source is the pushed source commit S.
	Source string
	// SecondReview is the effective workflow.autonomy.contract.second_review.
	SecondReview string
	// Registry values verify needs for the contracts in S's tree.
	RegistryRuleIDs     []string
	RegistryFrozenFiles []string
}

// PushContractResult is one in-push contract's readiness outcome.
type PushContractResult struct {
	SpecID string
	Codes  []string
}

// PushEvalResult is the whole push's evaluation (design.md §C.2/§C.4).
type PushEvalResult struct {
	// Undetermined reports push_check_undetermined (a missing remote ref, a
	// git failure, an unreadable evidence home). It is never allowed through.
	Undetermined bool
	Cause        string
	Results      []PushContractResult
}

// UndeterminedPush is the fail-closed result for an unprovable push.
func UndeterminedPush(cause string) PushEvalResult {
	return PushEvalResult{Undetermined: true, Cause: cause}
}

// EvaluatePush evaluates push readiness for every signed push-develop
// contract that is a candidate on the own-card basis of REQ-CLOSURE-015
// (design.md §C.2). A push with no in-push contract is ready.
func EvaluatePush(seam PushGitSeam, in PushEvalInput) PushEvalResult {
	remoteRef := in.Remote + "/" + in.Integration
	base, err := seam.ResolveRef(in.Tree, remoteRef)
	if err != nil {
		return UndeterminedPush("no " + remoteRef + " ref: " + err.Error())
	}
	commits, err := seam.NonMergeCommits(in.Tree, base, in.Source)
	if err != nil {
		return UndeterminedPush("range listing failed: " + err.Error())
	}
	var rangePaths []string
	for _, c := range commits {
		rangePaths = append(rangePaths, c.Paths...)
	}

	paths, err := seam.PathsUnder(in.Tree, in.Source, ".moai/specs/")
	if err != nil {
		return UndeterminedPush("spec listing failed: " + err.Error())
	}
	specIDs := contractSpecIDs(paths)

	var results []PushContractResult
	for _, specID := range specIDs {
		ownPrefix := ".moai/specs/" + specID
		ownEdit := slices.ContainsFunc(rangePaths, func(p string) bool {
			return p == ownPrefix || strings.HasPrefix(p, ownPrefix+"/")
		})

		raw, err := seam.Blob(in.Tree, in.Source, ownPrefix+"/contract.yaml")
		if err != nil {
			continue // not a contract (the path filter above matched loosely)
		}
		c, derr := contract.Decode(raw)
		if derr != nil || c.Signature == nil || !slices.Contains(c.Actions, "push-develop") {
			continue
		}

		// Candidate on the governed-path branch only when the SPEC is
		// non-terminal in S (REQ-CLOSURE-015).
		if !ownEdit {
			statusRaw, serr := seam.Blob(in.Tree, in.Source, ownPrefix+"/spec.md")
			if serr != nil {
				return UndeterminedPush("spec.md unreadable for " + specID + ": " + serr.Error())
			}
			rep := verifyAt(seam, in, specID, raw, statusRaw)
			if rep.Terminal {
				continue // closed on the remote, SPEC dir unchanged: not a candidate
			}
			governed := slices.ContainsFunc(rangePaths, func(p string) bool {
				return governedPath(c.Ownership.Write, specID, p)
			})
			if !governed {
				continue
			}
		}

		statusRaw, serr := seam.Blob(in.Tree, in.Source, ownPrefix+"/spec.md")
		if serr != nil {
			statusRaw = nil
		}
		rep := verifyAt(seam, in, specID, raw, statusRaw)

		// Card evidence home (§C.3): the live worktree named by the signed
		// card, else the primary checkout. Unreadable → undetermined.
		home, herr := ResolveEvidenceHome(in.Tree, c.Card)
		if herr != nil {
			return UndeterminedPush("card evidence home unresolvable for " + specID + ": " + herr.Error())
		}
		ev := EvidenceFor(home, c.Card)
		evIn := EvidenceInput{}
		if report, ok, _ := readFileOrEmpty(ev.ReportJSON); ok {
			var parsed Report
			if json.Unmarshal(report, &parsed) == nil {
				evIn.ClosureReportFound = true
				evIn.ClosureReportHeadSHA = parsed.HeadSHA
			}
		}
		evIn.SecondReviews, _, _ = LoadSecondReviews(ev.SecondReview)
		evIn.Verdicts, _, _ = LoadVerdictRecords(ev.ClosureVerdict)

		codes := EvaluateReadiness(ContractReadiness{
			SpecID:             specID,
			Card:               c.Card,
			VerifyState:        rep.State,
			ContractDigest:     rep.RecordedContractSHA256,
			WriteGlobs:         writeGlobs(c),
			SecondReviewPolicy: effectiveSecondReview(in.SecondReview),
		}, evIn, in.Source, gitFactsOver(in.Tree))

		if len(codes) > 0 {
			results = append(results, PushContractResult{SpecID: specID, Codes: codes})
		}
	}
	return PushEvalResult{Results: results}
}

// verifyAt runs A1's Verify over the contract bytes read from S's tree.
func verifyAt(seam PushGitSeam, in PushEvalInput, specID string, contractRaw, specMD []byte) contract.Report {
	acceptance, _ := seam.Blob(in.Tree, in.Source, ".moai/specs/"+specID+"/"+contract.AcceptanceFile)
	receipt, _ := seam.Blob(in.Tree, in.Source, ".moai/specs/"+specID+"/"+contract.ReceiptFile)
	ci := contract.Inputs{
		SpecID:            specID,
		Contract:          contractRaw,
		Acceptance:        acceptance,
		AcceptancePresent: acceptance != nil,
		Receipt:           receipt,
		ReceiptPresent:    receipt != nil,
		Policy: contract.Policy{
			SecondReview: in.SecondReview,
			PushDevelop:  true,
			Mode:         "contract",
		},
		RegistryRuleIDs:     in.RegistryRuleIDs,
		RegistryFrozenFiles: in.RegistryFrozenFiles,
		SpecStatus:          parseFrontmatterStatus(specMD),
	}
	return contract.Verify(ci)
}

// contractSpecIDs extracts the SPEC IDs whose contract.yaml exists in the
// given path list, sorted.
func contractSpecIDs(paths []string) []string {
	var out []string
	for _, p := range paths {
		rest := strings.TrimPrefix(p, ".moai/specs/")
		if strings.HasSuffix(rest, "/"+contract.ContractFile) && rest != contract.ContractFile {
			out = append(out, strings.TrimSuffix(rest, "/"+contract.ContractFile))
		}
	}
	slices.Sort(out)
	return out
}

// parseFrontmatterStatus reads the `status:` value of a spec.md frontmatter
// (the same frontmatter-anchored read as spec.ParseStatus, over bytes read
// from a git tree).
func parseFrontmatterStatus(specMD []byte) string {
	if specMD == nil {
		return ""
	}
	lines := strings.Split(string(specMD), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			break
		}
		if v, ok := strings.CutPrefix(line, "status:"); ok {
			s := strings.TrimSpace(v)
			if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
				s = s[1 : len(s)-1]
			}
			return s
		}
	}
	return ""
}

func writeGlobs(c *contract.Contract) []string {
	if c.Ownership != nil {
		return c.Ownership.Write
	}
	return nil
}

// effectiveSecondReview normalizes the policy for the readiness evaluator:
// anything but advisory/off is required (A1 Policy semantics).
func effectiveSecondReview(policy string) string {
	switch policy {
	case PolicyAdvisory, PolicyOff:
		return policy
	default:
		return PolicyRequired
	}
}

func gitFactsOver(tree string) GitFacts {
	return GitFacts{
		IsAncestor:      func(a, b string) (bool, error) { return gitio.IsAncestor(tree, a, b) },
		NonMergeCommits: func(from, to string) ([]CommitPaths, error) { return gitio.NonMergeCommits(tree, from, to) },
	}
}
