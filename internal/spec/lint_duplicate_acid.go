package spec

// lint_duplicate_acid.go — DuplicateAcceptanceIDRule (card t564).
//
// WHY LINT NEVER SAW THIS. The inline parser has always returned
// DuplicateAcceptanceID, but parseSPECDoc discarded the whole error slice
// (`criteria, _ :=`) and no rule emitted the code, so a duplicate id read as 0
// through `moai spec lint` whether or not it fired. `moai spec view` treated the
// same error as fatal. One event, silent on one surface and fatal on the other.
//
// WHAT THE PARSER NOW DOES. It cannot tell a bullet that cites an id from the
// bullet that declares it, so it keeps the first line's text and collects the
// REQ mappings of every line carrying the id. That can count coverage the author
// did not intend — a citing bullet with a `maps` tail. This finding is the
// counterweight: the duplicate is always reported, so the widened coverage never
// goes unseen.
//
// SEVERITY. Warning, not advisory, so --strict escalates it. The live corpus
// carried 0 duplicate ids when this rule landed, so escalation reddens nothing.
// The code is deliberately NOT in eraDemotableCodes: that map demotes errors
// only, so an entry would be inert while reading as intent.

import "fmt"

// DuplicateAcceptanceIDRule reports an inline AC id declared on more than one line.
type DuplicateAcceptanceIDRule struct{}

func (r *DuplicateAcceptanceIDRule) Code() string { return "DuplicateAcceptanceID" }

// Check emits one finding per repeated occurrence, at its body-relative line.
func (r *DuplicateAcceptanceIDRule) Check(doc *SPECDoc, _ []*SPECDoc) []Finding {
	var findings []Finding
	for _, dup := range doc.DuplicateACIDs {
		findings = append(findings, Finding{
			File:     doc.Path,
			Line:     dup.Line,
			Severity: SeverityWarning,
			Code:     r.Code(),
			Message: fmt.Sprintf("AC %s is declared on more than one line; the first line's text is kept "+
				"and the REQ mappings of every line count as coverage", dup.ID),
		})
	}
	return findings
}
