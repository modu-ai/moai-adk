#!/usr/bin/env python3
"""app.js handler fire probe (SPEC-APPJS-FIRE-GUARD-001, card t1060).

Derived from the field-proven t1041 browser probe
(.claude/worktrees/t1041/.moai/reports/t1041/browser-probe.py) — that probe
caught a real regression in two builds (t1041 verdict E3: all five indicators
died on the broken build, all five lived on the fixed one). This probe keeps
the same CDP-over-websockets machinery and adds what the SPEC requires:

  - an explicit (page, selector, observable effect) MANIFEST covering every
    click/change-family addEventListener group in internal/web/assets/app.js
    (each group is a manifest entry or an exclusion with a stated reason),
  - a three-value exit contract: 0 = every indicator fired AND zero
    ReferenceErrors on load/swap windows AND every manifest selector matched;
    1 = an indicator collapsed or a selector matched nothing (the report names
    WHAT flipped); 2 = machine fault (CDP unreachable, server unreachable) —
    a fault is not a product defect and must not read as red against the app,
  - a --lint-manifest offline self-check (effect-kind allowlist, coverage
    count, post-swap presence).

Orthogonality: this guard is orthogonal to the static sibling
SPEC-APPJS-IIFE-GUARD-001 (static-scope analysis). Neither guard's green
implies the other's. A static IIFE-scope pass does not prove a handler fires
in a real browser; a green run here does not pinpoint the offending
identifier or line — static-scope stays the owner of precise cause reports.

What this guard does NOT cover (SPEC-APPJS-FIRE-GUARD-001 spec.md §F):
  1. interactions outside the manifest (excluded groups are not measured —
     the quality of the exclusion reasons bounds the guard),
  2. paths the fixed scenario never exercises (one scenario, one order),
  3. browsers other than Chrome (single-engine guard),
  4. precise static-cause attribution (that is the static guard's job).

Usage:
  appjs_fire_probe.py [--cdp-port N] [--base-url URL] <server-port> <label>
  appjs_fire_probe.py --lint-manifest

Dependencies: stdlib (asyncio, json, sys, urllib.request) + websockets (the
only third-party dependency; pinned in the test-browser CI job — never a Go
module). Chrome is located by the caller (the Go driver or the CI job) and
reached through its CDP port.
"""

import asyncio
import hashlib
import json
import optparse
import os
import sys
import urllib.request

import websockets

# ── Manifest ────────────────────────────────────────────────────────────────
#
# inventory_total is the count of click/change-family addEventListener
# registration sites in internal/web/assets/app.js, measured 2026-09-22 on
# tree WT-appjs-handler-guard (spec.md §B.3):
#
#   grep -n -E "addEventListener\((['\"])(click|submit|change|input)" \
#     internal/web/assets/app.js   -> 13 sites
#
# Every group is either a manifest entry (line_group records the source line
# at authoring time — a documentation anchor only; the load-bearing checks are
# selector survival and the count cross-check the Go driver performs against
# the live asset) or an exclusion with a stated reason. If app.js gains or
# loses a registration group, the Go driver's inventory cross-check goes red
# and this manifest must be updated — a stale manifest is never silently green.

INVENTORY_TOTAL = 13

# Reversible effect kinds a manifest entry may exercise unconditionally
# (REQ-AFG-012). Save- and submit-family controls are outside this family: the
# probe must never write project configuration.
ALLOWED_EFFECTS = {"visibility", "label", "clipboard", "tab", "swap"}

# Conditional persistence family (REQ-AFG-014, card t1106). A CLOSED
# enumeration whose only member today is `validation-reject`: such an entry may
# exercise a form submit ONLY when it carries BOTH markers below. This is an
# enumeration, not a rule - "any persisting kind that carries the markers"
# would be a widening, and adding a member is a SPEC amendment.
CONDITIONAL_EFFECTS = {"validation-reject"}

# The two manifest markers a conditional entry must carry, inseparably:
#   requires_sandbox_serving    -> REQ-AFG-014 (1): a dedicated second server
#                                  serves this entry from a disposable project
#                                  copy, never from the real repo root. This
#                                  marker IS the routing key (route_for_entry).
#   requires_no_write_assertion -> REQ-AFG-014 (2): the probe snapshots the
#                                  sandbox root immediately before the submit
#                                  and right after the reject render settles,
#                                  and fails on a single changed byte.
# REQ-AFG-014 (3) (lifetime bound to t.TempDir()) is deliberately NOT a
# manifest marker: it is a Go-side lifetime property a committed manifest
# cannot carry, and the driver judges it (AC-AFG-011 (d)).
REQUIRED_CONDITIONAL_MARKERS = ("requires_sandbox_serving", "requires_no_write_assertion")

# Paths the exercised request's write seams can reach inside the sandbox root.
# POST /save persists project configuration through SyncToProjectConfig and
# writeProjectConfig (internal/web/handlers.go), and both land under
# .moai/config/sections/. A no-write comparison exclusion MUST NOT cover this
# subtree even with a stated reason (REQ-AFG-014 (2)): excluding it would make
# the assertion green by construction exactly where a regression would appear.
WRITE_SEAM_PREFIXES = (".moai/config/sections",)

# Paths excluded from the byte-invariance comparison, each with its reason.
# EMPTY today, and that emptiness is measured rather than assumed: the sandbox
# server's profile store lives OUTSIDE the sandbox root (ProfileBaseDir is a
# separate t.TempDir()), so the reject path touches nothing inside it. An
# entry added here must state WHY, and lint_manifest refuses any entry that
# reaches into WRITE_SEAM_PREFIXES.
NO_WRITE_EXCLUSIONS = []  # [{"path": "<relative path>", "reason": "<why>"}]

ENTRIES = [
    {
        "id": "popover_open",
        "line_group": 73,
        "page": "/settings",
        "selector": '[data-pop="profile"]',
        "effect": "visibility",
        "check": "click opens the profile popover panel (hidden true -> false)",
    },
    {
        "id": "popover_close_btn",
        "line_group": 83,
        "page": "/settings",
        "selector": "[data-pop-close]",
        "effect": "visibility",
        "check": "the panel's close button hides the popover panel",
    },
    {
        "id": "popover_outside_close",
        "line_group": 109,
        "page": "/settings",
        "selector": '[data-pop="profile"]',
        "effect": "visibility",
        "check": "an outside document click hides the open popover panel",
    },
    {
        "id": "settings_tabs",
        "line_group": 352,
        "page": "/settings",
        "selector": '.subnav__row[role="tab"]',
        "effect": "tab",
        "check": "clicking an inactive tab selects it (aria-selected true)",
    },
    {
        "id": "glm_reveal",
        "line_group": 603,
        "page": "/settings",
        "selector": "#glmKeyReveal",
        "effect": "visibility",
        "check": "click reveals the GLM key output node (hidden true -> false)",
    },
    {
        "id": "swap_todo_nav",
        "line_group": None,
        "page": "/",
        "selector": 'a[href="/todo"]',
        "effect": "swap",
        "check": "clicking the nav link performs an hx-boost body swap to /todo",
    },
    {
        "id": "popover_after_swap",
        "line_group": 73,
        "page": "/todo",
        "selector": '[data-pop="profile"]',
        "effect": "visibility",
        "post_swap": True,
        "check": "REQ-AFG-007: an indicator still fires AFTER the hx-boost swap",
    },
    {
        "id": "copy_button",
        "line_group": 648,
        "page": "/specs",
        "selector": "[data-copy]",
        "effect": "label",
        "check": "click flashes the copy button label to the check mark",
    },
    {
        # card t1106 - the validation-reject submit surface. line_group is None
        # for the same reason swap_todo_nav's is: this is an htmx-boost +
        # server-render surface, not an app.js addEventListener registration
        # group, so it does not move INVENTORY_TOTAL.
        "id": "validation_reject_banner",
        "line_group": None,
        "page": "/settings",
        "selector": "#settings-form",
        "effect": "validation-reject",
        "requires_sandbox_serving": True,
        "requires_no_write_assertion": True,
        "submit_button": 'button[type="submit"][form="settings-form"]',
        "banner_selector": '.banner[role="status"]',
        "invalid_field": "permission_mode",
        "invalid_value": "bogus",
        "check": "submitting an invalid permission_mode paints the validation-reject banner (the card t1105 fix) on screen",
    },
]

EXCLUSIONS = [
    {
        "line_group": 158,
        "selector": "#serverShutdown",
        "reason": "destructive: POST /__shutdown__ stops the server (REQ-AFG-012 data-loss path)",
    },
    {
        "line_group": 327,
        "selector": "#uiLangSelect",
        "reason": "persisting side effect: writes the locale to browser localStorage; excluded by default per REQ-AFG-012 / plan §C",
    },
    {
        "line_group": 406,
        "selector": 'select[name^="agentfm."]',
        "reason": "unsaved form dirty-state mutation (radio/check state) — effect kind outside the reversible allowlist",
    },
    {
        "line_group": 414,
        "selector": 'input[name="performance_tier"]',
        "reason": "rewrites multiple unsaved form select values from the tier matrix — outside the allowlist",
    },
    {
        "line_group": 473,
        "selector": 'select[name^="agentfm."][name$=".model"]',
        "reason": "haiku effort lock: option disabled-state pairing — form state outside the allowlist",
    },
    {
        "line_group": 513,
        "selector": 'select[name^="llm.glm.models."]',
        "reason": "GLM flash effort lock: option disabled-state pairing — form state outside the allowlist",
    },
    {
        "line_group": 530,
        "selector": 'select[name="statusline_preset"]',
        "reason": "dead surface: no served template renders select[name=statusline_preset] or #custom-segments today; listener is inert (guard inside app.js)",
    },
]


def effect_problems(entry):
    """Judge one entry's effect kind against the two closed families.

    Returns a list of problem strings, each NAMING what is missing (AC-AFG-012
    (b) requires the rejection to say WHICH condition is absent).
    """
    eid = entry.get("id")
    effect = entry.get("effect")
    if effect in ALLOWED_EFFECTS:
        return []
    if effect in CONDITIONAL_EFFECTS:
        missing = [m for m in REQUIRED_CONDITIONAL_MARKERS if entry.get(m) is not True]
        return [
            "entry %r declares effect %r but is missing the required condition marker %r "
            "- the effect kind and its conditions are inseparable (REQ-AFG-014)"
            % (eid, effect, m)
            for m in missing
        ]
    return [
        "entry %r effect %r belongs to neither closed family: unconditional %s / conditional %s"
        % (eid, effect, sorted(ALLOWED_EFFECTS), sorted(CONDITIONAL_EFFECTS))
    ]


def route_for_entry(entry, primary_base, sandbox_base):
    """Return the base URL this entry is driven against (REQ-AFG-014 (1)).

    The sandbox-serving marker IS the routing key, in BOTH directions: a marked
    entry never runs against the primary base, and an unmarked entry never runs
    against the sandbox base. Returns None when a marked entry has no sandbox
    base - the caller turns that into exit 2 (caller wiring fault), never a
    silent skip and never a fallback to the primary base.
    """
    if entry.get("requires_sandbox_serving") is True:
        return sandbox_base or None
    return primary_base


def lint_manifest(extra_entry=None):
    """--lint-manifest: offline structural self-check of the manifest.

    Exit 0 = manifest well-formed; 1 = a rule is violated (named); 2 = the
    mode itself was misused. --extra-entry appends ONE synthetic entry to the
    judged set without touching the committed manifest, so the reverse
    direction of the rule (does it actually reject?) can be OBSERVED rather
    than assumed (AC-AFG-012 (b)).
    """
    problems = []
    entries = list(ENTRIES)
    if extra_entry is not None:
        entries.append(extra_entry)
    if len(entries) == 0:
        problems.append("manifest has zero entries - a green with an empty manifest measures nothing")
    for e in entries:
        for field in ("id", "page", "selector", "effect", "check"):
            if not e.get(field):
                problems.append("entry %r missing field %r" % (e.get("id"), field))
        problems.extend(effect_problems(e))
    for x in EXCLUSIONS:
        if not x.get("selector") or not x.get("reason"):
            problems.append("exclusion for line_group %r needs selector and reason" % (x.get("line_group"),))
    # No-write comparison exclusions: each needs a reason, and none may cover a
    # path the exercised write seam can reach (REQ-AFG-014 (2)).
    for x in NO_WRITE_EXCLUSIONS:
        if not x.get("path") or not x.get("reason"):
            problems.append("no-write exclusion %r needs both path and reason" % (x,))
            continue
        rel = x["path"].lstrip("./")
        for prefix in WRITE_SEAM_PREFIXES:
            if rel == prefix or rel.startswith(prefix + "/"):
                problems.append(
                    "no-write exclusion %r covers %r, which the exercised request's write seam reaches "
                    "- excluding it would make the byte-invariance assertion green by construction"
                    % (x["path"], prefix)
                )
    entry_groups = {e["line_group"] for e in ENTRIES if e.get("line_group")}
    exclusion_groups = {x["line_group"] for x in EXCLUSIONS if x.get("line_group")}
    overlap = entry_groups & exclusion_groups
    if overlap:
        problems.append("line_group(s) %s appear as both entry and exclusion" % sorted(overlap))
    covered = entry_groups | exclusion_groups
    if len(covered) != INVENTORY_TOTAL:
        problems.append(
            "inventory coverage: %d unique groups covered, INVENTORY_TOTAL=%d - every click/change "
            "registration group must be an entry or an exclusion" % (len(covered), INVENTORY_TOTAL)
        )
    if not any(e.get("post_swap") for e in ENTRIES):
        problems.append("no entry exercises an indicator AFTER the hx-boost swap (REQ-AFG-007)")
    if problems:
        for p in problems:
            print("LINT: " + p)
        return 1
    print(
        "LINT OK: %d entries + %d exclusions cover %d inventory groups; effects within "
        "unconditional %s or conditional %s (conditional entries carry %s); post-swap entry present"
        % (
            len(entries),
            len(EXCLUSIONS),
            len(covered),
            sorted(ALLOWED_EFFECTS),
            sorted(CONDITIONAL_EFFECTS),
            list(REQUIRED_CONDITIONAL_MARKERS),
        )
    )
    return 0


# ── CDP machinery (inherited from the t1041 probe) ──────────────────────────


def new_tab(cdp_host):
    req = urllib.request.Request(cdp_host + "/json/new?about:blank", method="PUT")
    return json.load(urllib.request.urlopen(req, timeout=10))


def close_tab(cdp_host, tid):
    try:
        urllib.request.urlopen(cdp_host + "/json/close/" + tid, timeout=5).read()
    except Exception:
        pass


class CDP:
    def __init__(self, ws):
        self.ws = ws
        self.n = 0
        self.events = []

    async def send(self, method, params=None, timeout=20):
        self.n += 1
        mid = self.n
        await self.ws.send(json.dumps({"id": mid, "method": method, "params": params or {}}))
        while True:
            msg = json.loads(await asyncio.wait_for(self.ws.recv(), timeout=timeout))
            if msg.get("id") == mid:
                if msg.get("error"):
                    raise RuntimeError("CDP %s failed: %s" % (method, msg["error"]))
                return msg
            if "method" in msg:
                self.events.append(msg)

    async def drain(self, seconds):
        try:
            while True:
                msg = json.loads(await asyncio.wait_for(self.ws.recv(), timeout=seconds))
                if "method" in msg:
                    self.events.append(msg)
        except asyncio.TimeoutError:
            pass

    def take_errors(self):
        out = []
        for e in self.events:
            method = e.get("method")
            p = e.get("params", {})
            if method == "Runtime.exceptionThrown":
                d = p.get("exceptionDetails", {})
                ex = d.get("exception") or {}
                out.append((d.get("text", "") + " " + (ex.get("description") or "")).strip())
            elif method == "Runtime.consoleAPICalled" and p.get("type") == "error":
                out.append(" ".join(str(a.get("value", a.get("description", ""))) for a in p.get("args", [])))
            elif method == "Log.entryAdded" and p.get("entry", {}).get("level") == "error":
                out.append(p["entry"].get("text", ""))
        self.events = []
        return out


def only_reference_errors(errors):
    return [e for e in errors if "ReferenceError" in e]


async def ev(cdp, expr):
    r = await cdp.send(
        "Runtime.evaluate",
        {"expression": expr, "returnByValue": True, "awaitPromise": True, "userGesture": True},
    )
    return r.get("result", {}).get("result", {}).get("value")


async def poll(cdp, expr, want, timeout=5.0, interval=0.2):
    """Poll a JS expression until it equals `want` (condition wait, not sleep)."""
    waited = 0.0
    while waited < timeout:
        val = await ev(cdp, expr)
        if val == want:
            return val
        await asyncio.sleep(interval)
        waited += interval
    return await ev(cdp, expr)


async def navigate(cdp, url, drain_seconds=4.0):
    await cdp.send("Page.navigate", {"url": url})
    await cdp.drain(drain_seconds)


# ── Scenario ────────────────────────────────────────────────────────────────

# ReferenceError collection windows: page load, the post-swap window, and the
# final fresh load. A ReferenceError in ANY of them fails the run.


async def run_scenario(cdp, base):
    rep = {}

    # Phase 1 — load /settings (carries the GLM reveal control, the tabs, and
    # the shell rail popover).
    await navigate(cdp, base + "/settings")
    rep["p1_load_referenceerrors"] = only_reference_errors(cdp.take_errors())
    rep["p1_has_glm_btn"] = await ev(cdp, "!!document.querySelector('#glmKeyReveal')")

    # Phase 2 — glm_reveal: click reveals the output node. Poll the visibility
    # flip (the reveal fetch may resolve or fail; both paths un-hide the node).
    rep["p2_revealed_hidden_before"] = await ev(
        cdp, "(document.getElementById('glmKeyRevealed')||{}).hidden"
    )
    await ev(cdp, "document.querySelector('#glmKeyReveal').click()")
    rep["p2_revealed_hidden_after"] = await poll(
        cdp, "(document.getElementById('glmKeyRevealed')||{}).hidden", False
    )
    rep["p2_glm_handler_fired"] = (
        rep["p2_revealed_hidden_before"] is True and rep["p2_revealed_hidden_after"] is False
    )
    cdp.take_errors()

    # Phase 3 — popover_open / popover_close_btn / popover_outside_close.
    panel_sel = '[data-pop-panel="profile"]'
    rep["p3_panel_hidden_before"] = await ev(cdp, "(document.querySelector('%s')||{}).hidden" % panel_sel)
    await ev(cdp, "document.querySelector('[data-pop=\"profile\"]').click()")
    rep["p3_panel_hidden_after_open"] = await poll(cdp, "(document.querySelector('%s')||{}).hidden" % panel_sel, False)
    rep["p3_popover_open_fired"] = (
        rep["p3_panel_hidden_before"] is True and rep["p3_panel_hidden_after_open"] is False
    )
    rep["p3_has_close_btn"] = await ev(cdp, "!!document.querySelector('[data-pop-close]')")
    await ev(cdp, "document.querySelector('[data-pop-close]').click()")
    rep["p3_panel_hidden_after_close"] = await poll(
        cdp, "(document.querySelector('%s')||{}).hidden" % panel_sel, True
    )
    rep["p3_popover_close_btn_fired"] = rep["p3_panel_hidden_after_close"] is True
    await ev(cdp, "document.querySelector('[data-pop=\"profile\"]').click()")
    await poll(cdp, "(document.querySelector('%s')||{}).hidden" % panel_sel, False)
    await ev(cdp, "document.body.click()")
    rep["p3_panel_hidden_after_outside"] = await poll(
        cdp, "(document.querySelector('%s')||{}).hidden" % panel_sel, True
    )
    rep["p3_popover_outside_close_fired"] = rep["p3_panel_hidden_after_outside"] is True
    cdp.take_errors()

    # Phase 4 — settings_tabs: click the tab after the active one.
    rep["p4_tab_count"] = await ev(
        cdp, "document.querySelectorAll('.subnav__row[role=\"tab\"]').length"
    )
    rep["p4_tab_selected_before"] = await ev(
        cdp,
        "(function(){var t=document.querySelectorAll('.subnav__row[role=\"tab\"]');"
        "for(var i=0;i<t.length;i++){if(t[i].getAttribute('aria-selected')==='true')return i}return -1})()",
    )
    rep["p4_tab_clicked"] = await ev(
        cdp,
        "(function(){var t=document.querySelectorAll('.subnav__row[role=\"tab\"]');"
        "var i=(%d===-1)?0:((%d+1)%%t.length);t[i].click();return i})()" % (rep["p4_tab_selected_before"] or 0, rep["p4_tab_selected_before"] or 0),
    )
    rep["p4_tab_selected_after"] = await poll(
        cdp,
        "(function(){var t=document.querySelectorAll('.subnav__row[role=\"tab\"]');"
        "for(var i=0;i<t.length;i++){if(t[i].getAttribute('aria-selected')==='true')return i}return -1})()",
        rep["p4_tab_clicked"] if isinstance(rep["p4_tab_clicked"], int) else 1,
    )
    rep["p4_settings_tabs_fired"] = (
        isinstance(rep["p4_tab_selected_before"], int)
        and isinstance(rep["p4_tab_selected_after"], int)
        and rep["p4_tab_selected_before"] != rep["p4_tab_selected_after"]
        and rep["p4_tab_selected_after"] == rep["p4_tab_clicked"]
    )
    cdp.take_errors()

    # Phase 5 — swap_todo_nav: click the nav link; htmx boost swaps the body
    # without a full reload. URL must land on /todo with no ReferenceErrors.
    rep["p5_swap_clicked"] = await ev(
        cdp,
        "(function(){var a=document.querySelector('a[href=\"/todo\"]');"
        "if(a){a.click();return true}return false})()",
    )
    rep["p5_url_after_swap"] = await poll(cdp, "location.pathname", "/todo", timeout=8.0)
    rep["p5_swap_referenceerrors"] = only_reference_errors(cdp.take_errors())
    cdp.take_errors()

    # Phase 6 — popover_after_swap (REQ-AFG-007): the swap killed the old DOM;
    # a wired popover trigger on the NEW body must still fire (initConsole
    # re-runs on htmx:afterSettle).
    rep["p6_panel_hidden_before"] = await ev(cdp, "(document.querySelector('%s')||{}).hidden" % panel_sel)
    await ev(cdp, "document.querySelector('[data-pop=\"profile\"]').click()")
    rep["p6_panel_hidden_after"] = await poll(cdp, "(document.querySelector('%s')||{}).hidden" % panel_sel, False)
    rep["p6_popover_after_swap_fired"] = (
        rep["p6_panel_hidden_before"] is True and rep["p6_panel_hidden_after"] is False
    )
    cdp.take_errors()

    # Phase 7 — fresh load of /specs, then copy_button label flash.
    await navigate(cdp, base + "/specs")
    rep["p7_load_referenceerrors"] = only_reference_errors(cdp.take_errors())
    rep["p7_has_copy_btn"] = await ev(cdp, "!!document.querySelector('[data-copy]')")
    rep["p7_label_before_click"] = await ev(
        cdp,
        "(function(){var b=document.querySelector('[data-copy]');if(!b)return null;"
        "return (b.querySelector('[data-i18n]')||b).textContent})()",
    )
    await ev(
        cdp,
        "(function(){var b=document.querySelector('[data-copy]');if(!b)return false;b.click();return true})()",
    )
    rep["p7_label_after_click"] = await poll(
        cdp,
        "(function(){var b=document.querySelector('[data-copy]');if(!b)return null;"
        "return (b.querySelector('[data-i18n]')||b).textContent})()",
        "✓",
    )
    rep["p7_copy_handler_fired"] = rep["p7_label_after_click"] == "✓"
    cdp.take_errors()

    return rep


# ── Byte invariance over the sandbox root (REQ-AFG-014 (2)) ───────────


def no_write_excluded(rel):
    """True when `rel` is a stated exclusion from the byte comparison.

    NO_WRITE_EXCLUSIONS is empty today, so this returns False for everything
    and the comparison covers the whole sandbox root. lint_manifest refuses an
    exclusion that reaches into WRITE_SEAM_PREFIXES, so this can never be made
    to look away from the paths the exercised request actually writes.
    """
    for x in NO_WRITE_EXCLUSIONS:
        path = x.get("path", "").lstrip("./")
        if path and (rel == path or rel.startswith(path + "/")):
            return True
    return False


def snapshot_tree(root):
    """Content hash of every file under `root`, keyed by relative path."""
    out = {}
    for dirpath, _dirnames, filenames in os.walk(root):
        for fn in filenames:
            full = os.path.join(dirpath, fn)
            rel = os.path.relpath(full, root).replace(os.sep, "/")
            if no_write_excluded(rel):
                continue
            try:
                with open(full, "rb") as fh:
                    out[rel] = hashlib.sha256(fh.read()).hexdigest()
            except OSError as exc:
                out[rel] = "unreadable: %s" % exc
    return out


def diff_snapshots(before, after):
    """Name every path that changed — a count alone would not say WHAT moved."""
    changed = []
    for rel, digest in after.items():
        if rel not in before:
            changed.append({"path": rel, "change": "added"})
        elif before[rel] != digest:
            changed.append({"path": rel, "change": "modified"})
    for rel in before:
        if rel not in after:
            changed.append({"path": rel, "change": "removed"})
    return sorted(changed, key=lambda c: c["path"])


# ── Sandbox scenario: the validation-reject submit (card t1106) ──────────


async def poll_nonempty(cdp, expr, timeout=10.0, interval=0.2):
    """Poll until a JS expression yields a non-empty value (condition wait)."""
    waited = 0.0
    val = None
    while waited < timeout:
        val = await ev(cdp, expr)
        if val:
            return val
        await asyncio.sleep(interval)
        waited += interval
    return val


# The paint predicate (REQ-AFG-015). Presence in the DOM is NOT the question —
# the response-body layer already answers that. This asks whether the banner
# REACHED THE SCREEN: a layout box of non-zero area, no hiding ancestor
# anywhere up the chain, and non-empty text. What it deliberately does not
# judge is named in limit 5 of the header: wording, contrast, scroll position
# and focus are outside this guard.
PAINT_JS = """(function(){
  var b=document.querySelector(%s);
  if(!b){return {found:false};}
  var r=b.getBoundingClientRect();
  var hidden=false, by='';
  for(var n=b;n&&n.nodeType===1;n=n.parentElement){
    var cs=window.getComputedStyle(n);
    if(n.hasAttribute('hidden')||cs.display==='none'||cs.visibility==='hidden'||
       cs.visibility==='collapse'||parseFloat(cs.opacity)===0){
      hidden=true; by=n.tagName+(n.className?('.'+String(n.className).split(' ').join('.')):''); break;
    }
  }
  var txt=(b.textContent||'').trim();
  return {found:true, box:(r.width>0&&r.height>0), width:r.width, height:r.height,
          hidden:hidden, hidden_by:by, text_len:txt.length, text:txt.slice(0,200)};
})()"""


async def run_sandbox_scenario(cdp, base, sandbox_root, entry):
    """Drive the one sandbox-served entry: submit invalid, observe the paint.

    The byte-invariance window is narrow on purpose (plan §A0): the snapshots
    bracket the submit and the settled reject render, not the page load, so a
    server's ordinary read-path side effects never enter the comparison as
    noise.
    """
    rep = {}
    banner_sel = json.dumps(entry["banner_selector"])
    submit_sel = json.dumps(entry["submit_button"])
    form_sel = json.dumps(entry["selector"])

    await navigate(cdp, base + entry["page"])
    rep["s_load_referenceerrors"] = only_reference_errors(cdp.take_errors())
    rep["s_has_form"] = await ev(cdp, "!!document.querySelector(%s)" % form_sel)
    rep["s_has_submit"] = await ev(cdp, "!!document.querySelector(%s)" % submit_sel)
    # A banner already on screen before the submit would make "painted after
    # the reject" vacuous — the run must start from its absence.
    rep["s_banner_before_submit"] = await ev(cdp, "!!document.querySelector(%s)" % banner_sel)

    before = snapshot_tree(sandbox_root)
    rep["s_snapshot_files"] = len(before)

    # Place a value the server-side validator rejects. The control is a select
    # with no such option, so the option is appended first: the point is to
    # exercise the server's reject path, not to simulate a reachable keystroke.
    rep["s_invalid_set"] = await ev(
        cdp,
        """(function(){
  var f=document.querySelector(%s); if(!f){return 'no-form';}
  var el=f.querySelector('[name=%s]'); if(!el){return 'no-field';}
  if(el.tagName==='SELECT'){var o=document.createElement('option');o.value=%s;o.textContent=%s;el.appendChild(o);}
  el.value=%s;
  return el.value;
})()"""
        % (
            form_sel,
            json.dumps(entry["invalid_field"]),
            json.dumps(entry["invalid_value"]),
            json.dumps(entry["invalid_value"]),
            json.dumps(entry["invalid_value"]),
        ),
    )
    rep["s_submit_clicked"] = await ev(
        cdp,
        "(function(){var b=document.querySelector(%s); if(!b){return false;} b.click(); return true;})()" % submit_sel,
    )
    rep["s_banner_text"] = await poll_nonempty(
        cdp, "(function(){var b=document.querySelector(%s); return b?(b.textContent||'').trim():'';})()" % banner_sel
    )

    # Mutation-probe hooks. Both are OFF unless the caller asks for them, and
    # both exist so the red direction of an assertion can be OBSERVED instead
    # of argued: a guard nobody has seen fail is a guard nobody has measured.
    if opts.inject_sandbox_write:
        target = os.path.join(sandbox_root, opts.inject_sandbox_write)
        with open(target, "ab") as fh:
            fh.write(b"\n")
        rep["s_injected_write"] = opts.inject_sandbox_write
    if opts.inject_banner_hidden:
        rep["s_injected_banner_hidden"] = await ev(
            cdp,
            "(function(){var b=document.querySelector(%s); if(!b){return false;} b.style.display='none'; return true;})()"
            % banner_sel,
        )

    rep["s_paint"] = await ev(cdp, PAINT_JS % banner_sel)
    rep["s_window_referenceerrors"] = only_reference_errors(cdp.take_errors())

    after = snapshot_tree(sandbox_root)
    rep["s_changed_paths"] = diff_snapshots(before, after)
    return rep


# ── Judgement (three-value exit contract) ───────────────────────────────────
#
# exit 0 = every manifest indicator fired AND every selector matched AND zero
#          ReferenceErrors in the load/swap windows.
# exit 1 = an indicator collapsed, a selector matched nothing, or a
#          ReferenceError appeared — the report names WHAT failed.
# exit 2 = machine fault (server unreachable, CDP unreachable) — never a
#          product defect.


def judge(rep, driven_ids):
    """Return the list of failure names for the observed report.

    Only the DRIVEN entries are judged: an entry the run declared out of scope
    is reported by name in the run's accounting (REQ-AFG-014 (1)(ii)), never
    silently judged against observations that were never made.
    """
    failures = []
    missing = []
    paint = rep.get("s_paint") or {}
    by_id = {
        "glm_reveal": rep.get("p2_glm_handler_fired"),
        "popover_open": rep.get("p3_popover_open_fired"),
        "popover_close_btn": rep.get("p3_popover_close_btn_fired"),
        "popover_outside_close": rep.get("p3_popover_outside_close_fired"),
        "settings_tabs": rep.get("p4_settings_tabs_fired"),
        "swap_todo_nav": rep.get("p5_url_after_swap") == "/todo" and rep.get("p5_swap_clicked") is True,
        "popover_after_swap": rep.get("p6_popover_after_swap_fired"),
        "copy_button": rep.get("p7_copy_handler_fired"),
        # card t1106: the banner must be PAINTED, the submit must have carried
        # a value the validator rejects, and the sandbox must be untouched.
        "validation_reject_banner": (
            rep.get("s_invalid_set") == "bogus"
            and paint.get("found") is True
            and paint.get("box") is True
            and paint.get("hidden") is False
            and (paint.get("text_len") or 0) > 0
            and not rep.get("s_changed_paths")
        ),
    }
    selector_found = {
        "glm_reveal": rep.get("p1_has_glm_btn"),
        "popover_open": rep.get("p3_panel_hidden_before") is not None,
        "popover_close_btn": rep.get("p3_has_close_btn"),
        "popover_outside_close": rep.get("p3_panel_hidden_before") is not None,
        "settings_tabs": isinstance(rep.get("p4_tab_count"), int) and rep.get("p4_tab_count", 0) > 0,
        "swap_todo_nav": rep.get("p5_swap_clicked") is True,
        "popover_after_swap": rep.get("p6_panel_hidden_before") is not None,
        "copy_button": rep.get("p7_has_copy_btn"),
        "validation_reject_banner": rep.get("s_has_form") is True and rep.get("s_has_submit") is True,
    }
    for entry in ENTRIES:
        eid = entry["id"]
        if eid not in driven_ids:
            continue
        if selector_found.get(eid) is not True:
            missing.append({"entry": eid, "selector": entry["selector"]})
            # A selector that matches nothing is itself a failure (REQ-AFG-004):
            # manifest staleness must be red, never a quiet zero.
            failures.append(
                {"entry": eid, "reason": "selector matched nothing", "selector": entry["selector"]}
            )
            continue
        if by_id.get(eid) is not True:
            failures.append({"entry": eid, "reason": "indicator did not fire", "check": entry["check"]})
    windows = ["p1_load_referenceerrors", "p5_swap_referenceerrors", "p7_load_referenceerrors"]
    if "validation_reject_banner" in driven_ids:
        windows += ["s_load_referenceerrors", "s_window_referenceerrors"]
        # A changed byte is its own failure, named by path: "an indicator did
        # not fire" would say nothing about WHAT the reject path wrote.
        for changed in rep.get("s_changed_paths") or []:
            failures.append(
                {
                    "entry": "validation_reject_banner",
                    "reason": "sandbox root changed across the submit (%s)" % changed.get("change"),
                    "detail": changed.get("path"),
                }
            )
    for window in windows:
        for err in rep.get(window, []) or []:
            failures.append({"entry": None, "reason": "ReferenceError in %s" % window, "detail": err})
    rep["failures"] = failures
    rep["missing_selectors"] = missing
    return failures


async def main_async(args):
    port = args[0]
    label = args[1]
    base = opts.base_url or ("http://127.0.0.1:" + port)
    cdp_host = "http://127.0.0.1:" + str(opts.cdp_port)

    sandbox_base = opts.sandbox_base_url
    rep = {
        "label": label,
        "port": port,
        "base_url": base,
        "sandbox_base_url": sandbox_base,
        "sandbox_root": opts.sandbox_root,
        "cdp_port": opts.cdp_port,
    }

    # Which entries does THIS run drive? A run may narrow to the real-root
    # family, but only by saying so: the reduction is an affirmative
    # declaration, never inferred from a missing sandbox base (REQ-AFG-014
    # (1)). The accounting below is what keeps a narrowed run honest — the
    # driven count and the names of the excluded entries both travel in the
    # report, so a reader can tell WHICH cycle produced a green.
    marked = [e for e in ENTRIES if e.get("requires_sandbox_serving") is True]
    if opts.primary_entries_only:
        driven = [e for e in ENTRIES if e.get("requires_sandbox_serving") is not True]
        excluded = [e["id"] for e in marked]
    else:
        driven = list(ENTRIES)
        excluded = []
    rep["reduction_declared"] = bool(opts.primary_entries_only)
    rep["driven_entries"] = [e["id"] for e in driven]
    rep["driven_count"] = len(driven)
    rep["excluded_entries"] = excluded

    if not driven:
        rep["failures"] = [{"entry": None, "reason": "the declaration drives no entry at all", "detail": ""}]
        rep["exit"] = 1
        print(json.dumps(rep, ensure_ascii=False, indent=2))
        return 1

    driven_marked = [e for e in driven if e.get("requires_sandbox_serving") is True]
    # Caller-wiring faults are exit 2, not exit 1: a missing sandbox base is a
    # defect in how the probe was invoked, not in the product. Skipping the
    # marked entry silently is forbidden by name — that would be this guard
    # committing the very sin it exists to catch (an unmeasured green).
    if driven_marked and not sandbox_base:
        rep["error"] = (
            "sandbox-serving entries were driven without --sandbox-base-url: %s "
            "(declare --primary-entries-only to drive the real-root family only; absence is not a declaration)"
            % [e["id"] for e in driven_marked]
        )
        print(json.dumps(rep, ensure_ascii=False, indent=2))
        return 2
    if driven_marked and not opts.sandbox_root:
        rep["error"] = (
            "sandbox-serving entries were driven without --sandbox-root: %s — "
            "there is no tree to assert byte invariance over" % [e["id"] for e in driven_marked]
        )
        print(json.dumps(rep, ensure_ascii=False, indent=2))
        return 2

    # Machine-fault preflight: the servers must answer before Chrome is engaged.
    for name, url in (("primary", base), ("sandbox", sandbox_base if driven_marked else None)):
        if not url:
            continue
        try:
            urllib.request.urlopen(url + "/", timeout=5).read(1024)
        except Exception as exc:
            print(json.dumps({"label": label, "error": "%s server unreachable: %s" % (name, exc)}))
            return 2

    tab = new_tab(cdp_host)
    try:
        async with websockets.connect(tab["webSocketDebuggerUrl"], max_size=20_000_000) as ws:
            cdp = CDP(ws)
            await cdp.send("Runtime.enable")
            await cdp.send("Log.enable")
            await cdp.send("Page.enable")
            if [e for e in driven if e.get("requires_sandbox_serving") is not True]:
                rep.update(await run_scenario(cdp, base))
            for entry in driven_marked:
                rep.update(await run_sandbox_scenario(cdp, sandbox_base, opts.sandbox_root, entry))
    finally:
        close_tab(cdp_host, tab["id"])

    failures = judge(rep, set(rep["driven_entries"]))
    rep["exit"] = 0 if not failures else 1
    print(json.dumps(rep, ensure_ascii=False, indent=2))
    return rep["exit"]


def main():
    p = optparse.OptionParser(usage="%prog [--cdp-port N] [--base-url URL] [--lint-manifest] <server-port> <label>")
    p.add_option("--cdp-port", default="9222", help="Chrome DevTools protocol port")
    p.add_option("--base-url", default=None, help="base URL of the console server (default http://127.0.0.1:<port>)")
    p.add_option("--lint-manifest", action="store_true", default=False, help="validate the manifest offline and exit")
    p.add_option(
        "--sandbox-base-url",
        default=None,
        help="base URL of the SECOND server, the one serving the disposable project copy; "
        "entries carrying the sandbox-serving marker are driven against this base and no other",
    )
    p.add_option(
        "--sandbox-root",
        default=None,
        help="filesystem path of the disposable project copy — the tree whose byte invariance the "
        "no-write assertion measures across the submit (REQ-AFG-014 (2))",
    )
    p.add_option(
        "--primary-entries-only",
        action="store_true",
        default=False,
        help="affirmative reduction declaration (REQ-AFG-014 (1)): drive ONLY the entries without "
        "the sandbox-serving marker. Absence of --sandbox-base-url never implies this — without the "
        "declaration a marked entry with no sandbox base is exit 2, never a silent skip",
    )
    p.add_option(
        "--inject-sandbox-write",
        default=None,
        help="mutation probe: append one byte to <sandbox-root>/<PATH> between the two snapshots, so the "
        "red direction of the byte-invariance assertion can be observed (AC-AFG-011 (c))",
    )
    p.add_option(
        "--inject-banner-hidden",
        action="store_true",
        default=False,
        help="mutation probe: hide the reject banner before measuring paint, so a predicate that only "
        "checks node existence is caught passing a banner nobody can see (AC-AFG-010)",
    )
    p.add_option(
        "--print-routing",
        action="store_true",
        default=False,
        help="offline: print the per-entry base-URL routing decision as JSON and exit, so the routing "
        "can be judged in both directions without a browser (AC-AFG-013)",
    )
    p.add_option(
        "--extra-entry",
        default=None,
        help="mutation probe: JSON object appended to the judged entry set for --lint-manifest only "
        "(the committed manifest is never modified) - lets the reverse direction of the rule be observed",
    )
    global opts
    (opts, args) = p.parse_args()
    if opts.lint_manifest:
        extra = None
        if opts.extra_entry:
            try:
                extra = json.loads(opts.extra_entry)
            except ValueError as exc:
                print("LINT: --extra-entry is not valid JSON: %s" % exc)
                return 2
            if not isinstance(extra, dict):
                print("LINT: --extra-entry must be a JSON object")
                return 2
        return lint_manifest(extra)
    if opts.print_routing:
        routes = {}
        for e in ENTRIES:
            routes[e["id"]] = route_for_entry(e, opts.base_url, opts.sandbox_base_url)
        print(
            json.dumps(
                {
                    "routes": routes,
                    "sandbox_marked": {e["id"]: e.get("requires_sandbox_serving") is True for e in ENTRIES},
                },
                ensure_ascii=False,
                indent=2,
            )
        )
        return 0
    if len(args) != 2:
        p.error("expected <server-port> <label>")
    try:
        return asyncio.run(main_async(args))
    except Exception as exc:  # machine fault: CDP / protocol / navigation failure
        print(json.dumps({"label": args[1] if args else "", "error": "machine fault: %s" % exc}))
        return 2


if __name__ == "__main__":
    sys.exit(main())
