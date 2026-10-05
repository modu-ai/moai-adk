// backlog_relation.go — SPEC-TODO-CARD-ISSUANCE-001 M3: the relation
// vocabulary, its per-kind constraints, and the card↔GTD resolver
// (REQ-TCI-012..014).
//
// Stored finding rows are NEVER rewritten: the seven-kind vocabulary is a
// READ-TIME mapping over the legacy relation names (design §5.2), and the
// new write kinds (duplicates, supersedes, relates-to) extend the findings
// vocabulary so old rows read alongside new ones. Only the measures live in
// backlog_analysis.go; only the classifier's skip-dropped doctrine lives
// there — this file never calls ClassifyCardText.
package factory

import (
	"fmt"
	"strings"
)

// BacklogRelationKind names the seven relation kinds of the unified
// vocabulary (REQ-TCI-012). The CardRelation* prefix separates them from
// the GTDRelationKind family (gtd_relation.go) — the two vocabularies map
// onto each other at read time and must not share constants.
type BacklogRelationKind string

const (
	CardRelationBlocks     BacklogRelationKind = "blocks"
	CardRelationDuplicates BacklogRelationKind = "duplicates"
	CardRelationParentOf   BacklogRelationKind = "parent-of"
	CardRelationFollowUpOf BacklogRelationKind = "follow-up-of"
	CardRelationMergedInto BacklogRelationKind = "merged-into"
	CardRelationSupersedes BacklogRelationKind = "supersedes"
	CardRelationRelatesTo  BacklogRelationKind = "relates-to"
)

// BacklogWriteableRelations is the set `todo relate` may write: today's six
// semantic names PLUS the three new kinds. The projection kinds
// (parent-of, follow-up-of, merged-into) are deliberately absent — they are
// projections of the add-time attributes or written only by `todo merge`,
// and a relation verb handing them out would be the natural over-extension
// mistake TestTodoRelateRefusesProjectionKinds pins (REQ-TCI-013).
var BacklogWriteableRelations = append(append([]string{}, BacklogSemanticRelations...),
	"duplicates", "supersedes", "relates-to")

// BacklogRelationIsSymmetricForDedup normalizes at write time: a pair
// recorded again in the opposite order maps onto the first record instead
// of creating a second one (REQ-TCI-013). blocks and supersedes are
// directed; everything else writable is symmetric for dedup purposes.
func BacklogRelationIsSymmetricForDedup(relation string) bool {
	switch relation {
	case "blocks", "depends", "supersedes", "replaces":
		return false
	default:
		return true
	}
}

// issuanceIDNum extracts the numeric part of a t<N> card id; unparsable ids
// sort by string so the normalization stays total.
func issuanceIDNum(id string) (int, bool) {
	if len(id) < 2 || id[0] != 't' {
		return 0, false
	}
	n := 0
	for k := 1; k < len(id); k++ {
		if id[k] < '0' || id[k] > '9' {
			return 0, false
		}
		n = n*10 + int(id[k]-'0')
	}
	return n, true
}


// NormalizeRelationPair orders a symmetric pair (smaller numeric id first)
// and reports whether the input was swapped into canonical order.
func NormalizeRelationPair(a, b string) (first, second string, swapped bool) {
	na, oka := issuanceIDNum(a)
	nb, okb := issuanceIDNum(b)
	if oka && okb && na > nb {
		return b, a, true
	}
	return a, b, false
}

// MapLegacyRelation maps a stored finding's relation name onto the unified
// vocabulary at read time (REQ-TCI-012): depends normalizes through
// WaitsOnOf's direction, near-duplicate and duplicate-forced to duplicates,
// replaces to supersedes, contains/absorbs/conflicts to relates-to with the
// legacy name kept as a qualifier. The returned qualifier is "" for names
// that map one-to-one.
func MapLegacyRelation(relation string) (kind BacklogRelationKind, qualifier string) {
	switch relation {
	case "blocks", "depends":
		return CardRelationBlocks, ""
	case "near-duplicate", "duplicate-forced":
		return CardRelationDuplicates, relation
	case "replaces":
		return CardRelationSupersedes, ""
	case "contains", "absorbs", "conflicts":
		return CardRelationRelatesTo, relation
	default:
		// New-kind rows (duplicates/supersedes/relates-to) pass through.
		k := BacklogRelationKind(relation)
		if k == CardRelationDuplicates || k == CardRelationSupersedes || k == CardRelationRelatesTo {
			return k, ""
		}
		return CardRelationRelatesTo, relation
	}
}

// BacklogDirection is the resolved direction of one mapped finding.
type BacklogDirection struct {
	From, To  string
	Kind      BacklogRelationKind
	Qualifier string
	Source    string
}

// ResolveFindingDirection maps one stored finding onto the unified
// vocabulary, normalizing the legacy `depends` direction through the same
// semantics WaitsOnOf uses (the waiter is the SUBJECT: subject depends on
// related, so the unified blocks direction is related → subject).
func ResolveFindingDirection(f BacklogFinding) BacklogDirection {
	kind, qualifier := MapLegacyRelation(f.Relation)
	from, to := f.SubjectID, f.RelatedID
	if f.Relation == "depends" {
		// legacy `depends`: subject depends on related ⇒ related blocks
		// subject. The unified blocks kind reads predecessor → successor.
		from, to = f.RelatedID, f.SubjectID
	}
	return BacklogDirection{From: from, To: to, Kind: kind, Qualifier: qualifier, Source: f.Source}
}

// RecordRelation validates one relate write against the kind's constraints
// (design §5.3): self-edges refused for every writable kind, blocks and
// supersedes cycle-checked through the existing guard, symmetric pairs
// normalized so an opposite-order re-record maps onto the first record.
// The returned finding is appended by the caller inside its locked write;
// nothing here touches the record.
func RecordRelation(rec *BacklogRecord, subject, related, relation string) (BacklogFinding, error) {
	if subject == related {
		return BacklogFinding{}, fmt.Errorf("relation %s: self-edge %s refused", relation, subject)
	}
	writable := false
	for _, r := range BacklogWriteableRelations {
		if r == relation {
			writable = true
			break
		}
	}
	if !writable {
		return BacklogFinding{}, fmt.Errorf("--relation must be one of %s (got %q)",
			strings.Join(BacklogWriteableRelations, ", "), relation)
	}
	// Cycle guards: blocks (with legacy depends rows counting as blocks)
	// and supersedes walk their own kind over the mapped record.
	if relation == "blocks" || relation == "supersedes" {
		if relationClosesCycle(rec, subject, related, relation) {
			return BacklogFinding{}, fmt.Errorf("relation %s: %s → %s closes a cycle",
				relation, subject, related)
		}
	}
	if BacklogRelationIsSymmetricForDedup(relation) {
		first, second, _ := NormalizeRelationPair(subject, related)
		for i := range rec.Findings {
			f := rec.Findings[i]
			kind, _ := MapLegacyRelation(f.Relation)
			if kind != BacklogRelationKind(relation) {
				continue
			}
			if first == f.SubjectID && second == f.RelatedID {
				return BacklogFinding{}, fmt.Errorf("relation %s: %s/%s already recorded",
					relation, first, second)
			}
		}
		subject, related = first, second
	}
	return BacklogFinding{
		SubjectID: subject,
		RelatedID: related,
		Relation:  relation,
		Source:    "agent",
		At:        "",
	}, nil
}

// relationClosesCycle walks the mapped kind's edges from the prospective
// successor and reports whether the prospective predecessor is reachable —
// the same visit-set termination the trace command uses, so legacy cycles
// in OTHER kinds cannot hang the walk.
func relationClosesCycle(rec *BacklogRecord, from, to, relation string) bool {
	adjacent := map[string][]string{}
	for i := range rec.Findings {
		f := rec.Findings[i]
		kind, _ := MapLegacyRelation(f.Relation)
		if kind != BacklogRelationKind(relation) {
			continue
		}
		d := ResolveFindingDirection(f)
		adjacent[d.From] = append(adjacent[d.From], d.To)
	}
	seen := map[string]bool{}
	queue := []string{to}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node == from {
			return true
		}
		if seen[node] {
			continue
		}
		seen[node] = true
		queue = append(queue, adjacent[node]...)
	}
	return false
}
