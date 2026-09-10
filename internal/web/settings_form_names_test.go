package web

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// settingsFormID is the id of the form the save handler parses. Controls
// belong to it either by sitting inside its subtree or, anywhere in the page,
// by carrying form="settings-form".
const settingsFormID = "settings-form"

// TestSettingsRenderFormNamesUnique pins, at the render layer, the property the
// save parser relies on: within the submitted settings form, every form-control
// name appears once.
//
// parseSchemaForm atomically rejects a name submitted more than once (REQ-WWS-006).
// That guard lives downstream of the renderer, so a template that ships a
// duplicate text-field name passes every render test and surfaces only when a
// user saves — with an error that points at the parser rather than the template.
// This test moves the failure to where the defect is introduced.
//
// Scope is the settings form, not the whole page. Browsers submit only the
// controls that belong to the submitted form; the page carries other,
// independent forms that may legitimately reuse a name.
//
// Radio groups are the one documented exception: a browser submits only the
// checked radio of a group, so several radios sharing a name still submit one
// value. A name shared between a radio and any non-radio control is still a
// duplicate.
func TestSettingsRenderFormNamesUnique(t *testing.T) {
	body := renderSettingsGET(newTestApp(t))
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("parse rendered /settings: %v", err)
	}

	type occurrence struct {
		kind string // "input:<type>", "select", or "textarea"
	}
	byName := map[string][]occurrence{}
	foundForm := false

	var walk func(n *html.Node, insideForm bool)
	walk = func(n *html.Node, insideForm bool) {
		if n.Type == html.ElementNode {
			if n.Data == "form" && attrValue(n, "id") == settingsFormID {
				foundForm = true
				insideForm = true
			}
			if kind, ok := submittableControlKind(n); ok {
				name := attrValue(n, "name")
				owner, hasOwner := attrLookup(n, "form")
				belongs := insideForm
				if hasOwner {
					belongs = owner == settingsFormID
				}
				if name != "" && belongs {
					byName[name] = append(byName[name], occurrence{kind: kind})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, insideForm)
		}
	}
	walk(doc, false)

	if !foundForm {
		t.Fatalf("rendered /settings has no <form id=%q>; the uniqueness check would sweep nothing", settingsFormID)
	}

	radioGroups, otherNames := 0, 0
	var dups []string
	for name, occs := range byName {
		allRadio := true
		kinds := map[string]bool{}
		for _, o := range occs {
			kinds[o.kind] = true
			if o.kind != "input:radio" {
				allRadio = false
			}
		}
		if allRadio {
			radioGroups++
			continue
		}
		otherNames++
		if len(occs) > 1 {
			ks := make([]string, 0, len(kinds))
			for k := range kinds {
				ks = append(ks, k)
			}
			sort.Strings(ks)
			dups = append(dups, name+" x"+strconv.Itoa(len(occs))+" ("+strings.Join(ks, ", ")+")")
		}
	}

	// Guard against a vacuous pass: a sweep that found no radio group or no
	// other named control did not measure the rendered form.
	if radioGroups == 0 || otherNames == 0 {
		t.Fatalf("settings form sweep looks empty: radio groups=%d, other named controls=%d", radioGroups, otherNames)
	}

	if len(dups) > 0 {
		sort.Strings(dups)
		t.Errorf("the /settings render ships duplicate form-control names inside #%s:\n  %s\n"+
			"parseSchemaForm rejects the whole save when a name is submitted more than once; "+
			"fix the template that renders the duplicate, not the parser",
			settingsFormID, strings.Join(dups, "\n  "))
	}
}

// submittableControlKind reports whether n is a form control whose name is
// submitted with its form, and its kind. Buttons and button-like inputs are
// excluded: only the activating button's name is ever submitted.
func submittableControlKind(n *html.Node) (string, bool) {
	switch n.Data {
	case "select", "textarea":
		return n.Data, true
	case "input":
		typ := strings.ToLower(attrValue(n, "type"))
		if typ == "" {
			typ = "text"
		}
		switch typ {
		case "submit", "button", "image", "reset":
			return "", false
		}
		return "input:" + typ, true
	}
	return "", false
}

func attrLookup(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

func attrValue(n *html.Node, key string) string {
	v, _ := attrLookup(n, key)
	return v
}
