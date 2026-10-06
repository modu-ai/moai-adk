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
	"sort"
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
// that map one-to-one. The CURRENT vocabulary kinds — the writable three
// and the stored projection kind merged-into, todo merge's write — pass
// through with their own kind: remapping them to relates-to made todo
// merge's cycle guard, which reads the mapped kind, see no merged-into
// edges at all (card t1454 card-review r2 finding 12).
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
		k := BacklogRelationKind(relation)
		if k == CardRelationDuplicates || k == CardRelationSupersedes || k == CardRelationRelatesTo ||
			k == CardRelationMergedInto || k == CardRelationParentOf || k == CardRelationFollowUpOf {
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

// RelationKindClosesCycle reports whether recording `from` —kind→ `to`
// would close a cycle within that kind's mapped edges (legacy names count:
// replaces rides as supersedes, depends as blocks). Exported for the CLI's
// pre-write refusal (REQ-TCI-013).
func (r *BacklogRecord) RelationKindClosesCycle(from, to string, kinds ...string) bool {
	kindSet := map[BacklogRelationKind]bool{}
	for _, k := range kinds {
		kindSet[BacklogRelationKind(k)] = true
	}
	adjacent := map[string][]string{}
	for i := range r.Findings {
		f := r.Findings[i]
		kind, _ := MapLegacyRelation(f.Relation)
		if !kindSet[kind] {
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

// GTDCardRelation is one GTD relation whose endpoints the caller resolved
// onto todo card ids — the GTD store keys relations by GTD item id, and a
// GTD item links to its card through card_id. A relation whose endpoints
// name no card is dropped by the reader.
type GTDCardRelation struct {
	From, To string
	Kind     string // the GTD kind, rendered verbatim — its own vocabulary
	Source   string
}

// RelationEdge is one resolved relation edge naming a card, whatever its
// source: a stored finding mapped onto the unified vocabulary, an issuance
// attribute projection (spawned_by), or a GTD relation the caller resolved
// onto card ids.
type RelationEdge struct {
	// Index is the finding's 1-based record index — the address `todo
	// unrelate` takes; 0 for edges that are not findings.
	Index     int
	From, To  string
	Kind      string
	Qualifier string
	Source    string
}

// ResolveCardEdges is the common relation resolver (card t1454 card-review
// r2 finding 11): every edge the queue knows about the card, from the three
// sources — the stored findings (mapped onto the unified vocabulary), the
// issuance spawned_by projections, and the caller's GTD relations. The
// follow-up projection reads child → origin (the card is the follow-up of
// its origin), the parent projection parent → child. Deterministic:
// findings in record order, then the projections in queue order, then the
// GTD edges in the caller's order.
func ResolveCardEdges(rec *BacklogRecord, gtd []GTDCardRelation, id string) []RelationEdge {
	var out []RelationEdge
	for i := range rec.Findings {
		f := rec.Findings[i]
		if !f.Names(id) {
			continue
		}
		d := ResolveFindingDirection(f)
		out = append(out, RelationEdge{Index: i + 1, From: d.From, To: d.To, Kind: string(d.Kind), Qualifier: d.Qualifier, Source: d.Source})
	}
	for i := range rec.Items {
		it := &rec.Items[i]
		if it.Issuance == nil || it.Issuance.SpawnedBy == "" {
			continue
		}
		// The projection is the child's attribute, but the edge names both
		// ends: resolving either endpoint surfaces it.
		var child, origin string
		if it.ID == id {
			child, origin = id, it.Issuance.SpawnedBy
		} else if it.Issuance.SpawnedBy == id {
			child, origin = it.ID, id
		} else {
			continue
		}
		kind, from, to := CardRelationFollowUpOf, child, origin
		if it.Issuance.Origin == "split" {
			kind, from, to = CardRelationParentOf, origin, child
		}
		out = append(out, RelationEdge{From: string(from), To: string(to), Kind: string(kind), Source: "issuance"})
	}
	for _, g := range gtd {
		if g.From != id && g.To != id {
			continue
		}
		out = append(out, RelationEdge{From: g.From, To: g.To, Kind: g.Kind, Source: g.Source})
	}
	return out
}

// TraceCardRelations walks the resolved relation edges from start id, up to
// depth (0 = unbounded), over the named kinds (empty = all; the GTD kinds
// carry their own vocabulary and appear only in the all-kinds walk). Each
// visited edge renders as one deterministic line: depth, kind, from → to,
// qualifier, source — from → to is the edge's STORED or resolved direction,
// never the walk order (card t1454 card-review r2 finding 13). Cycles
// terminate through the visit set; nothing is written.
func TraceCardRelations(rec *BacklogRecord, gtd []GTDCardRelation, start string, kinds []string, depth int) []string {
	wanted := map[string]bool{}
	for _, k := range kinds {
		wanted[k] = true
	}
	type edge struct {
		depth int
		line  string
		sort  [3]string
	}
	visited := map[string]bool{start: true}
	var out []edge
	type queued struct {
		id    string
		depth int
	}
	queue := []queued{{start, 0}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if depth > 0 && cur.depth >= depth {
			continue
		}
		var edges []RelationEdge
		for _, e := range ResolveCardEdges(rec, gtd, cur.id) {
			if len(wanted) > 0 && !wanted[e.Kind] {
				continue
			}
			edges = append(edges, e)
		}
		sort.Slice(edges, func(i, j int) bool {
			a, b := edges[i], edges[j]
			aOther, bOther := a.To, b.To
			if a.To == cur.id {
				aOther = a.From
			}
			if b.To == cur.id {
				bOther = b.From
			}
			if aOther != bOther {
				return aOther < bOther
			}
			return a.Kind < b.Kind
		})
		for _, e := range edges {
			// The walk follows whichever endpoint is cur; the rendered line
			// keeps the edge's own direction.
			other := e.To
			if e.To == cur.id {
				other = e.From
			}
			if visited[other] {
				continue
			}
			visited[other] = true
			q := ""
			if e.Qualifier != "" {
				q = " (" + e.Qualifier + ")"
			}
			out = append(out, edge{
				depth: cur.depth + 1,
				line: fmt.Sprintf("depth %d  %s  %s → %s%s  source=%s",
					cur.depth+1, e.Kind, e.From, e.To, q, e.Source),
				sort: [3]string{fmt.Sprintf("%03d", cur.depth+1), e.Kind, other},
			})
			queue = append(queue, queued{other, cur.depth + 1})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		for k := range out[i].sort {
			if out[i].sort[k] != out[j].sort[k] {
				return out[i].sort[k] < out[j].sort[k]
			}
		}
		return false
	})
	lines := make([]string, 0, len(out))
	for _, e := range out {
		lines = append(lines, e.line)
	}
	return lines
}

// normalizedFindingPairPair reports whether the stored finding names the
// same normalized pair, relation, and source as f (card t1454 card-review
// r2 finding 15): the stored rows are normalized too, so a pair an older
// writer recorded in the opposite order still maps onto the first record.
func normalizedFindingTuple(f BacklogFinding, existing BacklogFinding) bool {
	if existing.Relation != f.Relation || existing.Source != f.Source {
		return false
	}
	subject, related := f.SubjectID, f.RelatedID
	eSubject, eRelated := existing.SubjectID, existing.RelatedID
	if BacklogRelationIsSymmetricForDedup(f.Relation) {
		subject, related, _ = NormalizeRelationPair(subject, related)
		eSubject, eRelated, _ = NormalizeRelationPair(eSubject, eRelated)
	}
	return subject == eSubject && related == eRelated
}

// HasNormalizedFindingTuple reports whether a finding with f's source,
// relation, and normalized pair is already recorded.
func (r *BacklogRecord) HasNormalizedFindingTuple(f BacklogFinding) bool {
	for _, existing := range r.Findings {
		if normalizedFindingTuple(f, existing) {
			return true
		}
	}
	return false
}

// RecordRelation validates one relate write against the kind's constraints
// (design §5.3): self-edges refused for every writable kind, blocks and
// supersedes cycle-checked through the mapped kind's edges, symmetric pairs
// normalized — the stored rows included (card t1454 card-review r2 finding
// 15) — so an opposite-order re-record maps onto the first record. The
// returned finding is appended by the caller inside its locked write;
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
	if relation == "blocks" {
		if rec.RelationKindClosesCycle(subject, related, "blocks", "depends") {
			return BacklogFinding{}, fmt.Errorf("relation blocks: %s → %s closes a cycle", subject, related)
		}
	}
	if relation == "supersedes" {
		if rec.RelationKindClosesCycle(subject, related, "supersedes", "replaces") {
			return BacklogFinding{}, fmt.Errorf("relation supersedes: %s → %s closes a cycle", subject, related)
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
			fFirst, fSecond, _ := NormalizeRelationPair(f.SubjectID, f.RelatedID)
			if first == fFirst && second == fSecond {
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

// relationClosesCycle is superseded by RelationKindClosesCycle (the
// exported pre-write refusal used by todo relate's blocks/supersedes
// guards); the unexported spelling had one caller and is retained only for
// this comment to record the succession.
