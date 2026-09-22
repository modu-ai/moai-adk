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
import json
import optparse
import re
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

# Reversible effect kinds a manifest entry may exercise (REQ-AFG-012). Save-
# and submit-family controls are outside the allowlist: the probe must never
# write project configuration.
ALLOWED_EFFECTS = {"visibility", "label", "clipboard", "tab", "swap"}

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


def lint_manifest():
    """--lint-manifest: offline structural self-check of the manifest.

    Exit 0 = manifest well-formed; 1 = a rule is violated (named); 2 = the
    mode itself was misused.
    """
    problems = []
    if len(ENTRIES) == 0:
        problems.append("manifest has zero entries — a green with an empty manifest measures nothing")
    for e in ENTRIES:
        for field in ("id", "page", "selector", "effect", "check"):
            if not e.get(field):
                problems.append("entry %r missing field %r" % (e.get("id"), field))
        if e.get("effect") not in ALLOWED_EFFECTS:
            problems.append(
                "entry %r effect %r outside the reversible allowlist %s"
                % (e.get("id"), e.get("effect"), sorted(ALLOWED_EFFECTS))
            )
    for x in EXCLUSIONS:
        if not x.get("selector") or not x.get("reason"):
            problems.append("exclusion for line_group %r needs selector and reason" % (x.get("line_group"),))
    entry_groups = {e["line_group"] for e in ENTRIES if e.get("line_group")}
    exclusion_groups = {x["line_group"] for x in EXCLUSIONS if x.get("line_group")}
    overlap = entry_groups & exclusion_groups
    if overlap:
        problems.append("line_group(s) %s appear as both entry and exclusion" % sorted(overlap))
    covered = entry_groups | exclusion_groups
    if len(covered) != INVENTORY_TOTAL:
        problems.append(
            "inventory coverage: %d unique groups covered, INVENTORY_TOTAL=%d — every click/change "
            "registration group must be an entry or an exclusion" % (len(covered), INVENTORY_TOTAL)
        )
    if not any(e.get("post_swap") for e in ENTRIES):
        problems.append("no entry exercises an indicator AFTER the hx-boost swap (REQ-AFG-007)")
    if problems:
        for p in problems:
            print("LINT: " + p)
        return 1
    print(
        "LINT OK: %d entries + %d exclusions cover %d inventory groups; all effects within %s; post-swap entry present"
        % (len(ENTRIES), len(EXCLUSIONS), len(covered), sorted(ALLOWED_EFFECTS))
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


# ── Judgement (three-value exit contract) ───────────────────────────────────
#
# exit 0 = every manifest indicator fired AND every selector matched AND zero
#          ReferenceErrors in the load/swap windows.
# exit 1 = an indicator collapsed, a selector matched nothing, or a
#          ReferenceError appeared — the report names WHAT failed.
# exit 2 = machine fault (server unreachable, CDP unreachable) — never a
#          product defect.


def judge(rep):
    """Return the list of failure names for the observed report."""
    failures = []
    missing = []
    by_id = {
        "glm_reveal": rep.get("p2_glm_handler_fired"),
        "popover_open": rep.get("p3_popover_open_fired"),
        "popover_close_btn": rep.get("p3_popover_close_btn_fired"),
        "popover_outside_close": rep.get("p3_popover_outside_close_fired"),
        "settings_tabs": rep.get("p4_settings_tabs_fired"),
        "swap_todo_nav": rep.get("p5_url_after_swap") == "/todo" and rep.get("p5_swap_clicked") is True,
        "popover_after_swap": rep.get("p6_popover_after_swap_fired"),
        "copy_button": rep.get("p7_copy_handler_fired"),
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
    }
    for entry in ENTRIES:
        eid = entry["id"]
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
    for window in ("p1_load_referenceerrors", "p5_swap_referenceerrors", "p7_load_referenceerrors"):
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

    rep = {"label": label, "port": port, "base_url": base, "cdp_port": opts.cdp_port}

    # Machine-fault preflight: the server must answer before Chrome is engaged.
    try:
        urllib.request.urlopen(base + "/", timeout=5).read(1024)
    except Exception as exc:
        print(json.dumps({"label": label, "error": "server unreachable: %s" % exc}))
        return 2

    tab = new_tab(cdp_host)
    try:
        async with websockets.connect(tab["webSocketDebuggerUrl"], max_size=20_000_000) as ws:
            cdp = CDP(ws)
            await cdp.send("Runtime.enable")
            await cdp.send("Log.enable")
            await cdp.send("Page.enable")
            rep.update(await run_scenario(cdp, base))
    finally:
        close_tab(cdp_host, tab["id"])

    failures = judge(rep)
    rep["exit"] = 0 if not failures else 1
    print(json.dumps(rep, ensure_ascii=False, indent=2))
    return rep["exit"]


def main():
    p = optparse.OptionParser(usage="%prog [--cdp-port N] [--base-url URL] [--lint-manifest] <server-port> <label>")
    p.add_option("--cdp-port", default="9222", help="Chrome DevTools protocol port")
    p.add_option("--base-url", default=None, help="base URL of the console server (default http://127.0.0.1:<port>)")
    p.add_option("--lint-manifest", action="store_true", default=False, help="validate the manifest offline and exit")
    global opts
    (opts, args) = p.parse_args()
    if opts.lint_manifest:
        return lint_manifest()
    if len(args) != 2:
        p.error("expected <server-port> <label>")
    try:
        return asyncio.run(main_async(args))
    except Exception as exc:  # machine fault: CDP / protocol / navigation failure
        print(json.dumps({"label": args[1] if args else "", "error": "machine fault: %s" % exc}))
        return 2


if __name__ == "__main__":
    sys.exit(main())
