package sign

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/contract"
)

// Writing a signed contract keeps the author's text. The edits are small —
// the acceptance binding, a missing budget, the signature block — so they are
// spliced into the original lines at positions taken from the parsed
// yaml.Node tree, and every other byte (comments, blank lines, quoting,
// indentation) is left as written. When the document has a shape the splice
// does not handle (a flow-style acceptance mapping, a multi-line value), the
// edits fall back to the yaml.Node tree re-encoded, which keeps comments but
// not blank lines. Either result is accepted only when it decodes to exactly
// the intended contract (see renderBody).

// bodyEdit is the contract body sign writes: the measured acceptance binding
// and, when the contract has none, the default budget. The signature block is
// always removed from the body.
type bodyEdit struct {
	sha    string
	count  int
	bind   bool             // write sha and count (false: acceptance.md is absent)
	budget *contract.Budget // nil: keep the contract's budget
}

var errNoSplice = errors.New("contract sign: document shape not spliceable")

// renderBody returns raw with the signature block removed and e applied. It
// verifies the result against want: the rendered body must decode strictly
// and digest to the same value as want, and carry no signature.
func renderBody(raw []byte, e bodyEdit, want *contract.Contract) ([]byte, error) {
	wantDigest, err := contract.Digest(want)
	if err != nil {
		return nil, err
	}
	accept := func(out []byte) bool {
		c, err := contract.Decode(out)
		if err != nil || c.Signature != nil {
			return false
		}
		d, err := contract.Digest(c)
		return err == nil && d == wantDigest
	}
	if out, err := spliceBody(raw, e); err == nil && accept(out) {
		return out, nil
	}
	out, err := reencodeBody(raw, e)
	if err != nil {
		return nil, err
	}
	if !accept(out) {
		return nil, fmt.Errorf("%w: rendered contract does not match the intended body", ErrInternal)
	}
	return out, nil
}

// appendSignature appends the signature block to a rendered body.
func appendSignature(body []byte, sig contract.Signature) ([]byte, error) {
	var sigNode yaml.Node
	if err := sigNode.Encode(sig); err != nil {
		return nil, err
	}
	doc := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: contract.SectionSignature}, &sigNode,
	}}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indentWidth(body))
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	eol := lineEnding(body)
	out := ensureTrailingEOL(body, eol)
	block := buf.String()
	if eol != "\n" {
		block = strings.ReplaceAll(block, "\n", eol)
	}
	return append(out, block...), nil
}

// rootMapping parses raw and returns its root mapping node.
func rootMapping(raw []byte) (*yaml.Node, *yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, nil, err
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, nil, errNoSplice
	}
	return &doc, doc.Content[0], nil
}

// pairIndex returns the index of key in mapping m's Content, or -1.
func pairIndex(m *yaml.Node, key string) int {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return i
		}
	}
	return -1
}

// lastLine is the highest 1-based line of any node in n's subtree.
func lastLine(n *yaml.Node) int {
	hi := n.Line
	for _, c := range n.Content {
		hi = max(hi, lastLine(c))
	}
	return hi
}

func isNullNode(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.ShortTag() == "!!null"
}

// singleLineScalar reports whether n is a scalar that cannot span lines.
func singleLineScalar(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Style&(yaml.LiteralStyle|yaml.FoldedStyle) == 0 &&
		!strings.Contains(n.Value, "\n")
}

// lineEdit replaces lines[start:end] (0-based) with repl.
type lineEdit struct {
	start, end int
	repl       []string
}

// spliceBody applies e to raw's text. Lines keep their own terminators; new
// lines take the document's line ending.
func spliceBody(raw []byte, e bodyEdit) ([]byte, error) {
	_, root, err := rootMapping(raw)
	if err != nil {
		return nil, err
	}
	eol := lineEnding(raw)
	unit := strings.Repeat(" ", indentWidth(raw))
	lines := strings.SplitAfter(string(raw), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	var edits []lineEdit

	if i := pairIndex(root, contract.SectionSignature); i >= 0 {
		k, v := root.Content[i], root.Content[i+1]
		if v.Kind == yaml.ScalarNode && !singleLineScalar(v) {
			return nil, errNoSplice
		}
		edits = append(edits, lineEdit{k.Line - 1, lastLine(v), nil})
	}

	if i := pairIndex(root, contract.SectionAcceptance); i >= 0 && e.bind {
		ed, err := spliceAcceptance(root.Content[i+1], e, eol)
		if err != nil {
			return nil, err
		}
		edits = append(edits, ed...)
	}

	var appendBudget []string
	if e.budget != nil {
		block := budgetLines(unit, *e.budget, eol)
		if i := pairIndex(root, contract.SectionBudget); i >= 0 {
			k, v := root.Content[i], root.Content[i+1]
			// Only a null budget reaches here (a present one is kept); a
			// null written on a later line is left to the fallback.
			if !isNullNode(v) || v.Line > k.Line {
				return nil, errNoSplice
			}
			edits = append(edits, lineEdit{k.Line - 1, k.Line, block})
		} else {
			appendBudget = block
		}
	}

	// Apply bottom-up so earlier line numbers stay valid; at an equal start
	// the wider edit (a removal) goes first and an insertion lands before it.
	slices.SortFunc(edits, func(a, b lineEdit) int {
		if a.start != b.start {
			return b.start - a.start
		}
		return b.end - a.end
	})
	for i := 1; i < len(edits); i++ {
		if edits[i].end > edits[i-1].start {
			return nil, errNoSplice // overlapping edits
		}
	}
	for _, ed := range edits {
		if ed.start < 0 || ed.end > len(lines) || ed.start > ed.end {
			return nil, errNoSplice
		}
		lines = slices.Concat(lines[:ed.start], ed.repl, lines[ed.end:])
	}
	out := []byte(strings.Join(lines, ""))
	if appendBudget != nil {
		out = append(ensureTrailingEOL(out, eol), strings.Join(appendBudget, "")...)
	}
	return out, nil
}

// spliceAcceptance returns the edits that set acceptance.sha256 and
// acceptance.ac_count in a block acceptance mapping.
func spliceAcceptance(m *yaml.Node, e bodyEdit, eol string) ([]lineEdit, error) {
	if m.Kind != yaml.MappingNode || m.Style&yaml.FlowStyle != 0 || len(m.Content) == 0 {
		return nil, errNoSplice
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if !singleLineScalar(m.Content[i+1]) || m.Content[i].Line != m.Content[i+1].Line {
			return nil, errNoSplice
		}
	}
	indent := strings.Repeat(" ", m.Content[0].Column-1)
	fields := []struct{ key, value string }{
		{"sha256", strconv.Quote(e.sha)},
		{"ac_count", strconv.Itoa(e.count)},
	}
	var edits []lineEdit
	var inserted []string
	for _, f := range fields {
		i := pairIndex(m, f.key)
		if i < 0 {
			inserted = append(inserted, indent+f.key+": "+f.value+eol)
			continue
		}
		k, v := m.Content[i], m.Content[i+1]
		comment := v.LineComment
		if comment == "" {
			comment = k.LineComment
		}
		line := indent + f.key + ": " + f.value
		if comment != "" {
			line += " " + comment
		}
		edits = append(edits, lineEdit{k.Line - 1, k.Line, []string{line + eol}})
	}
	if inserted != nil {
		at := lastLine(m)
		edits = append(edits, lineEdit{at, at, inserted})
	}
	return edits, nil
}

func budgetLines(unit string, b contract.Budget, eol string) []string {
	return []string{
		contract.SectionBudget + ":" + eol,
		fmt.Sprintf("%sturns: %d%s", unit, b.Turns, eol),
		fmt.Sprintf("%soperations: %d%s", unit, b.Operations, eol),
		fmt.Sprintf("%saudit_retries: %d%s", unit, b.AuditRetries, eol),
	}
}

// reencodeBody applies e to the yaml.Node tree and re-encodes it. Comments
// survive; blank lines do not.
func reencodeBody(raw []byte, e bodyEdit) ([]byte, error) {
	doc, root, err := rootMapping(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}
	if i := pairIndex(root, contract.SectionSignature); i >= 0 {
		root.Content = slices.Delete(root.Content, i, i+2)
	}
	if i := pairIndex(root, contract.SectionAcceptance); i >= 0 && e.bind {
		m := root.Content[i+1]
		if m.Kind == yaml.MappingNode {
			setScalar(m, "sha256", e.sha, "!!str")
			setScalar(m, "ac_count", strconv.Itoa(e.count), "!!int")
		}
	}
	if e.budget != nil {
		var b yaml.Node
		if err := b.Encode(*e.budget); err != nil {
			return nil, err
		}
		if i := pairIndex(root, contract.SectionBudget); i >= 0 {
			root.Content[i+1] = &b
		} else {
			root.Content = append(root.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: contract.SectionBudget}, &b)
		}
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indentWidth(raw))
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// setScalar sets key in mapping m to a scalar, keeping an existing node's
// comments.
func setScalar(m *yaml.Node, key, value, tag string) {
	style := yaml.Style(0)
	if tag == "!!str" {
		style = yaml.DoubleQuotedStyle
	}
	if i := pairIndex(m, key); i >= 0 {
		v := m.Content[i+1]
		v.Kind, v.Tag, v.Value, v.Style, v.Content = yaml.ScalarNode, tag, value, style, nil
		return
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value, Style: style})
}

// indentWidth is the document's indentation unit: the smallest leading-space
// count of an indented content line, clamped to [2, 8]; 2 when none.
func indentWidth(raw []byte) int {
	best := 0
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimLeft(line, " ")
		n := len(line) - len(trimmed)
		if n == 0 || strings.TrimSpace(trimmed) == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if best == 0 || n < best {
			best = n
		}
	}
	return min(max(best, 2), 8)
}

func lineEnding(raw []byte) string {
	if bytes.Contains(raw, []byte("\r\n")) {
		return "\r\n"
	}
	return "\n"
}

func ensureTrailingEOL(b []byte, eol string) []byte {
	out := bytes.Clone(b)
	if len(out) > 0 && !bytes.HasSuffix(out, []byte("\n")) {
		out = append(out, eol...)
	}
	return out
}
