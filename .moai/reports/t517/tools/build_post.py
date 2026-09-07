#!/usr/bin/env python3
"""Build a browser-equivalent POST /save body from the rendered settings page.

Parses the <form id="settings-form"> subtree of the rendered HTML and emits
url-encoded form data mirroring what a browser submits when the user clicks
Save without changing anything: text inputs carry their value=, selects carry
their selected option, checkboxes carry value=1 only when checked, and every
rendered hidden companion (name+"__present") is included. Unchecked toggles
without a rendered companion input are omitted (not submitted -> preserve).

Evidence tool for SPEC-WEB-WRITE-SAFETY-001 M1(d)/M2 — value-invariant save
reproduction. usage: build_post.py settings.html > body.txt (stderr = field log)
"""
import html.parser
import sys
import urllib.parse


class FormParser(html.parser.HTMLParser):
    def __init__(self):
        super().__init__()
        self.in_form = 0
        self.fields = []  # list of (name, value)
        self.cur = None  # current select/textarea pending entry

    def handle_starttag(self, tag, attrs):
        a = dict(attrs)
        if tag == "form":
            if a.get("id") == "settings-form":
                self.in_form += 1
            elif self.in_form:
                self.in_form += 1
            return
        if not self.in_form:
            return
        name = a.get("name")
        if tag == "input":
            if not name:
                return
            typ = (a.get("type") or "text").lower()
            if typ in ("checkbox", "radio"):
                # templ renders the bare attribute form (`checked` with no
                # value); html.parser hands that through as ('checked', None),
                # so presence-in-dict is the correct test, not is-not-None.
                if "checked" in a:
                    self.fields.append((name, a.get("value", "on")))
            elif typ in ("submit", "button", "image"):
                return
            else:
                self.fields.append((name, a.get("value", "")))
        elif tag == "select":
            self.cur = {"name": name, "picked": None}
        elif tag == "option" and self.cur is not None:
            # templ renders `selected` as a bare attribute; html.parser hands it
            # through as ('selected', None), so presence-in-dict is the test.
            sel = "selected" in a
            if sel or self.cur["picked"] is None:
                self.cur["picked"] = a.get("value", "")
        elif tag == "textarea" and name:
            self.cur = {"name": name, "picked": "", "textarea": True}

    def handle_endtag(self, tag):
        if tag == "form" and self.in_form:
            self.in_form -= 1
        if tag == "select" and self.cur is not None:
            if self.cur["picked"] is not None:
                self.fields.append((self.cur["name"], self.cur["picked"]))
            self.cur = None
        if tag == "textarea" and self.cur is not None:
            self.fields.append((self.cur["name"], self.cur["picked"]))
            self.cur = None

    def handle_data(self, data):
        if self.cur is not None and self.cur.get("textarea"):
            self.cur["picked"] += data


def main():
    with open(sys.argv[1], "r", encoding="utf-8") as f:
        html_text = f.read()
    p = FormParser()
    p.feed(html_text)
    pairs = []
    for name, value in p.fields:
        pairs.append((name, value))
    body = urllib.parse.urlencode(pairs)
    sys.stdout.write(body)
    sys.stderr.write(f"FIELDS={len(pairs)}\n")
    for name, value in pairs:
        sys.stderr.write(f"  {name}={value!r}\n")


if __name__ == "__main__":
    main()
