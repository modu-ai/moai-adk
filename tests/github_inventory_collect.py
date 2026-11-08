"""Read-only GitHub settings inventory and duplicate map for RC1 Issue 00-B (#1986).

Two modes, one pipeline:
  remote:  --repo OWNER/NAME --record DIR   fetch with GET only, redact to a field whitelist,
                                             write fixtures to DIR, then build from them
  offline: --fixtures DIR                    build from previously recorded fixtures

The build step is a pure function of the fixtures (plus the RC1 issue manifest and the form
files), so two recordings of the same remote state yield byte-identical outputs apart from
`fetched_at`. Standard library only, no sibling imports: it must run under `python3 -I`.

Every remote call goes through `gh_get`, which always sends an explicit GET and refuses any
argument that would make `gh api` send a body (and so switch to a writing method).
"""
import argparse
import hashlib
import json
import pathlib
import re
import subprocess
import sys

SCHEMA_VERSION = 1
ROOT = pathlib.Path(__file__).resolve().parent.parent
DEFAULT_MANIFEST = ROOT / ".moai-cli/manifests/github-issues.json"
PROJECTS_UNAVAILABLE = "unavailable(read:project scope missing)"

# Field whitelists (REQ-00B-09): nothing outside these keys is ever written.
ISSUE_FIELDS = ("number", "id", "node_id", "title", "state", "state_reason", "labels", "milestone", "type")
PR_FIELDS = ("number", "id", "node_id", "state", "merged")
LABEL_FIELDS = ("id", "name", "color", "description", "default")
MILESTONE_FIELDS = ("number", "id", "title", "state", "open_issues", "closed_issues")
ISSUE_TYPE_FIELDS = ("id", "name", "color", "is_enabled")

# Arguments that make `gh api` send a request body or another method.
FORBIDDEN_GH_ARGS = ("-f", "-F", "--field", "--raw-field", "--input", "-X", "--method")

# Label classification (REQ-00B-02). Explicit tables only; anything unmatched is `unknown`.
NAMESPACE_AXES = {"type:": "type", "priority:": "priority", "status:": "status", "area:": "area", "host:": "host"}
# canonical -> (members, basis). "issue-body" pairs are stated in the #1986 design; "name-match"
# pairs are proposed from the label names alone and are candidates for the HOLD-00B-1 decision.
SYNONYM_GROUPS = {
    "type:fix": (["bug"], "issue-body"),
    "type:feature": (["feature", "feature-request"], "issue-body"),
    "type:chore": (["chore"], "issue-body"),
    "type:ci": (["ci"], "name-match"),
    "type:docs": (["docs", "documentation"], "name-match"),
    "type:performance": (["performance"], "name-match"),
    "type:refactor": (["refactor"], "name-match"),
    "type:security": (["security"], "name-match"),
    "type:test": (["test", "tests"], "name-match"),
    "priority:P0": (["critical", "priority:critical"], "issue-body"),
    "priority:P1": (["high", "priority:high"], "name-match"),
    "priority:P2": (["medium", "priority:medium"], "name-match"),
    "priority:P3": (["low", "priority:low"], "name-match"),
    "status:blocked": (["blocked"], "issue-body"),
    "status:in-progress": (["in-progress"], "issue-body"),
    "status:review": (["review"], "issue-body"),
    "area:hooks": (["hooks"], "name-match"),
    "area:templates": (["templates"], "name-match"),
    "area:config": (["config"], "name-match"),
}
# Bare status labels the #1986 design names that have no namespaced counterpart.
BARE_STATUS = {"ready", "done", "testing", "triage", "investigating"}
# GitHub's documented default repository labels plus Dependabot's `dependencies`.
FUNCTIONAL = {"bug", "documentation", "duplicate", "enhancement", "good first issue", "help wanted",
              "invalid", "question", "wontfix", "dependencies"}

# Form inventory (REQ-00B-06): repository-relative path -> (tracked, adoption).
FORM_PATHS = {
    ".github/ISSUE_TEMPLATE/moai-feature.yml": ("tracked", "adopted"),
    ".github/ISSUE_TEMPLATE/moai-bugfix.yml": ("tracked", "adopted"),
    ".github/ISSUE_TEMPLATE/moai-investigation.yml": ("tracked", "adopted"),
    ".github/ISSUE_TEMPLATE/agentic-factory-01-parent-issue.yml": ("untracked", "not_adopted_preserved"),
    ".github/ISSUE_TEMPLATE/agentic-factory-02-child-issue.yml": ("untracked", "not_adopted_preserved"),
    ".github/ISSUE_TEMPLATE/agentic-factory-03-spec-change.yml": ("untracked", "not_adopted_preserved"),
    ".github/PULL_REQUEST_TEMPLATE/agentic-factory-11-handoff.md": ("untracked", "not_adopted_preserved"),
    ".github/workflows/label-sync.yml": ("untracked", "not_adopted_preserved"),
}
LABEL_FILE_PATHS = {
    ".github/moai/labels.yml": ("tracked", "adopted"),
    ".github/labels.yml": ("untracked", "not_adopted_preserved"),
    ".github/agentic-factory/labels.proposed.yml": ("untracked", "not_adopted_preserved"),
}

RC1_KEY = re.compile(r"\b(EP-0[1-9]|0[0-9]-[A-D])\b")
RC1_SPEC = re.compile(r"SPEC-RC1-[A-Z0-9-]+")
ANY_SPEC = re.compile(r"SPEC-[A-Z0-9-]+")
TITLE_SIMILARITY = 0.5


class WriteRefused(RuntimeError):
    pass


def gh_argv(path, paginate=True, include=False):
    """Build a `gh api` argv that can only send GET (REQ-00B-05)."""
    argv = ["gh", "api", "-X", "GET"]
    if paginate:
        argv += ["--paginate", "--slurp"]
    if include:
        argv.append("-i")
    argv.append(path)
    return argv


def check_get_only(argv):
    if argv[:4] != ["gh", "api", "-X", "GET"]:
        raise WriteRefused("not an explicit GET: %r" % argv[:4])
    if any(arg in FORBIDDEN_GH_ARGS for arg in argv[4:]):
        raise WriteRefused("argument would change the request method or send a body: %r" % argv)


def gh_get(path, log, paginate=True, include=False):
    argv = gh_argv(path, paginate, include)
    check_get_only(argv)
    log.append({"method": "GET", "path": path})
    out = subprocess.run(argv, check=True, capture_output=True, text=True).stdout
    if include:
        return out
    data = json.loads(out)
    if paginate:
        return [item for page in data for item in page] if data and isinstance(data[0], list) else data
    return data


def pick(obj, fields):
    return {k: obj.get(k) for k in fields}


def redact_issue(raw):
    item = pick(raw, ISSUE_FIELDS)
    item["labels"] = sorted(lab["name"] if isinstance(lab, dict) else lab for lab in raw.get("labels") or [])
    item["milestone"] = (raw.get("milestone") or {}).get("title")
    item["type"] = (raw.get("type") or {}).get("name") if isinstance(raw.get("type"), dict) else raw.get("type")
    return item


def redact_pr(raw):
    return {"number": raw["number"], "id": raw["id"], "node_id": raw["node_id"], "state": raw["state"],
            "merged": bool((raw.get("pull_request") or {}).get("merged_at"))}


def split_issues_and_prs(raw_items):
    """REQ-00B-04: items carrying a `pull_request` key are PRs, never Issues."""
    issues = [i for i in raw_items if "pull_request" not in i]
    prs = [i for i in raw_items if "pull_request" in i]
    return issues, prs


def by_number(items):
    return sorted(items, key=lambda x: x["number"])


def parse_scopes(headers_text):
    for line in headers_text.splitlines():
        if line.lower().startswith("x-oauth-scopes:"):
            return sorted(s.strip() for s in line.split(":", 1)[1].split(",") if s.strip())
    return []


def write_json(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def record(repo, fixtures_dir, now):
    """Fetch with GET only and write redacted fixtures."""
    log = []
    owner = repo.split("/")[0]
    repo_meta = gh_get("repos/%s" % repo, log, paginate=False)
    scopes = parse_scopes(gh_get("rate_limit", log, paginate=False, include=True))
    labels = gh_get("repos/%s/labels?per_page=100" % repo, log)
    types = gh_get("orgs/%s/issue-types" % owner, log, paginate=False)
    milestones = gh_get("repos/%s/milestones?state=all&per_page=100" % repo, log)
    everything = gh_get("repos/%s/issues?state=all&per_page=100" % repo, log)
    issues, prs = split_issues_and_prs(everything)
    sub_issues, dependencies = {}, {}
    for issue in by_number(issues):
        n = issue["number"]
        if (issue.get("sub_issues_summary") or {}).get("total"):
            children = gh_get("repos/%s/issues/%d/sub_issues?per_page=100" % (repo, n), log)
            sub_issues[str(n)] = sorted(c["number"] for c in children)
        if (issue.get("issue_dependencies_summary") or {}).get("total_blocked_by"):
            blockers = gh_get("repos/%s/issues/%d/dependencies/blocked_by?per_page=100" % (repo, n), log)
            dependencies[str(n)] = sorted(b["number"] for b in blockers)
    gh_version = subprocess.run(["gh", "--version"], check=True, capture_output=True, text=True).stdout.splitlines()[0]
    meta = {
        "fetched_at": now,
        "repository": {"full_name": repo, "id": repo_meta["id"], "default_branch": repo_meta["default_branch"],
                       "has_projects": repo_meta["has_projects"]},
        "gh_version": gh_version,
        "api_version_header": None,
        # Only the fact REQ-00B-08 needs; the token's other scopes are not recorded.
        "project_scope": "read:project" in scopes or "project" in scopes,
        "requests": log,
    }
    files = {
        "meta.json": meta,
        "labels.json": sorted((pick(lab, LABEL_FIELDS) for lab in labels), key=lambda x: x["name"]),
        "issue-types.json": sorted((pick(t, ISSUE_TYPE_FIELDS) for t in types), key=lambda x: x["id"]),
        "milestones.json": by_number(pick(m, MILESTONE_FIELDS) for m in milestones),
        "issues-open.json": by_number(redact_issue(i) for i in issues if i["state"] == "open"),
        "issues-closed.json": by_number(redact_issue(i) for i in issues if i["state"] == "closed"),
        "prs.json": by_number(redact_pr(p) for p in prs),
        "sub-issues.json": sub_issues,
        "dependencies.json": dependencies,
    }
    for name, data in files.items():
        write_json(fixtures_dir / name, data)


def load_fixtures(fixtures_dir):
    names = ["meta", "labels", "issue-types", "milestones", "issues-open", "issues-closed", "prs",
             "sub-issues", "dependencies"]
    return {n: json.loads((fixtures_dir / (n + ".json")).read_text(encoding="utf-8")) for n in names}


def classify_label(name):
    for canonical, (members, basis) in SYNONYM_GROUPS.items():
        if name in members:
            return "legacy-synonym", canonical, basis
    for prefix, axis in NAMESPACE_AXES.items():
        if name.startswith(prefix):
            return axis, None, None
    if name in BARE_STATUS:
        return "status", None, None
    if name in FUNCTIONAL:
        return "functional", None, None
    return "unknown", None, None


def label_inventory(labels):
    out, counts = [], {}
    present = {lab["name"] for lab in labels}
    for label in labels:
        axis, synonym_of, basis = classify_label(label["name"])
        counts[axis] = counts.get(axis, 0) + 1
        out.append(dict(label, axis=axis, synonym_of=synonym_of, synonym_basis=basis))
    groups = []
    for canonical, (members, basis) in SYNONYM_GROUPS.items():
        found = [m for m in members if m in present]
        if found:
            groups.append({"canonical": canonical, "canonical_present": canonical in present,
                           "members": found, "basis": basis})
    return out, dict(sorted(counts.items())), groups


def read_shasum(path):
    """Parse `shasum -a 256` output (`<hex>  <relative path>` per line)."""
    out = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        if line.strip():
            digest, rel = line.split(None, 1)
            out[rel.strip()] = digest
    return out


def file_inventory(table, root, recorded=None):
    out = []
    for rel, (tracked, adoption) in sorted(table.items()):
        if recorded is not None:
            sha = recorded.get(rel)
        else:
            path = root / rel
            sha = hashlib.sha256(path.read_bytes()).hexdigest() if path.is_file() else None
        out.append({"path": rel, "tracked": tracked, "adoption": adoption, "sha256": sha})
    return out


def manifest_index(manifest):
    return {v["number"]: k for k, v in manifest["issues"].items()}


def reconcile(manifest, sub_issues, dependencies):
    want_links = {(lk["parent"], lk["child"]) for lk in manifest["links"] if lk.get("status") == "linked"}
    want_deps = {(d["issue"], d["blocked_by"]) for d in manifest["dependencies"] if d.get("status") == "linked"}
    have_links = {(int(p), c) for p, cs in sub_issues.items() for c in cs}
    have_deps = {(int(i), b) for i, bs in dependencies.items() for b in bs}

    def pairs(s):
        return [list(p) for p in sorted(s)]

    return {
        "sub_issues": {"manifest_linked": len(want_links), "remote": len(have_links),
                       "missing_remote": pairs(want_links - have_links), "extra_remote": pairs(have_links - want_links)},
        "dependencies": {"manifest_linked": len(want_deps), "remote": len(have_deps),
                         "missing_remote": pairs(want_deps - have_deps), "extra_remote": pairs(have_deps - want_deps)},
    }


def tokens(title):
    title = re.sub(r"^(\[[^\]]*\]\s*)+", "", title.lower())
    return {t for t in re.findall(r"\w+", title) if len(t) >= 2}


def duplicate_map(fx, manifest, now):
    index = manifest_index(manifest)
    open_rows = []
    for issue in fx["issues-open"]:
        key = index.get(issue["number"])
        open_rows.append({"number": issue["number"], "id": issue["id"], "node_id": issue["node_id"],
                          "title": issue["title"], "logical_key": key, "non_rc1": key is None})
    rc1 = [(r["number"], r["logical_key"], tokens(r["title"])) for r in open_rows if r["logical_key"]]
    candidates, best = [], 0.0
    for issue in fx["issues-closed"]:
        reasons, matches = [], set()
        if RC1_SPEC.search(issue["title"]):
            reasons.append("spec_id")
        key_hits = set(RC1_KEY.findall(issue["title"]))
        closed_toks = tokens(issue["title"])
        for number, key, toks in rc1:
            if key in key_hits:
                matches.add(number)
                if "logical_key" not in reasons:
                    reasons.append("logical_key")
            union = toks | closed_toks
            score = len(toks & closed_toks) / len(union) if union else 0.0
            best = max(best, score)
            if score >= TITLE_SIMILARITY:
                matches.add(number)
                if "title_similar" not in reasons:
                    reasons.append("title_similar")
        if reasons:
            candidates.append({"number": issue["number"], "id": issue["id"], "title": issue["title"],
                               "state_reason": issue["state_reason"], "status": "candidate",
                               "reason": ",".join(reasons), "matches_open": sorted(matches)})
    return {
        "schema_version": SCHEMA_VERSION,
        "fetched_at": now,
        "repository": fx["meta"]["repository"]["full_name"],
        "rules": {"spec_id": RC1_SPEC.pattern, "logical_key": RC1_KEY.pattern,
                  "title_similar": "token Jaccard >= %.2f after stripping [..] prefixes" % TITLE_SIMILARITY,
                  "decision": "candidate only; a person confirms or rejects each duplicate"},
        "counts": {"issues_open": len(open_rows), "issues_closed": len(fx["issues-closed"]),
                   "open_mapped_rc1": sum(1 for r in open_rows if r["logical_key"]),
                   "open_non_rc1": sum(1 for r in open_rows if r["non_rc1"]),
                   "duplicate_candidates": len(candidates)},
        # Evidence that an empty candidate list was measured, not skipped.
        "observed": {"max_title_similarity": round(best, 2),
                     "closed_with_any_spec_id": sum(1 for i in fx["issues-closed"] if ANY_SPEC.search(i["title"])),
                     "closed_with_rc1_spec_id": sum(1 for i in fx["issues-closed"] if RC1_SPEC.search(i["title"])),
                     "closed_with_rc1_key": sum(1 for i in fx["issues-closed"] if RC1_KEY.search(i["title"]))},
        "issues_open": open_rows,
        "duplicate_candidates": candidates,
    }


def count_by(items, key):
    out = {}
    for item in items:
        value = item.get(key)
        value = "null" if value is None else str(value)
        out[value] = out.get(value, 0) + 1
    return dict(sorted(out.items()))


def settings_inventory(fx, manifest, forms_root, recorded=None):
    meta = fx["meta"]
    labels, axis_counts, groups = label_inventory(fx["labels"])
    issues = fx["issues-open"] + fx["issues-closed"]
    return {
        "schema_version": SCHEMA_VERSION,
        "fetched_at": meta["fetched_at"],
        "repository": meta["repository"],
        "gh_version": meta["gh_version"],
        "api_version": meta["api_version_header"] or "gh default (no X-GitHub-Api-Version header sent)",
        "project_scope": meta["project_scope"],
        "projects": "readable" if meta["project_scope"] else PROJECTS_UNAVAILABLE,
        "requests": meta["requests"],
        "labels": labels,
        "label_axis_counts": axis_counts,
        "synonym_groups": groups,
        "issue_types": fx["issue-types"],
        "milestones": fx["milestones"],
        "issue_counts": {"open": len(fx["issues-open"]), "closed": len(fx["issues-closed"]),
                         "by_state_reason": count_by(issues, "state_reason"),
                         "by_type": count_by(issues, "type")},
        "pr_counts": {"total": len(fx["prs"]), "by_state": count_by(fx["prs"], "state"),
                      "merged": sum(1 for p in fx["prs"] if p["merged"])},
        "prs_open": [p for p in fx["prs"] if p["state"] == "open"],
        "sub_issues": fx["sub-issues"],
        "dependencies": fx["dependencies"],
        "manifest_reconciliation": reconcile(manifest, fx["sub-issues"], fx["dependencies"]),
        "forms_hash_source": "recorded shasum file" if recorded is not None else "hashed from --forms-root",
        "forms": file_inventory(FORM_PATHS, forms_root, recorded),
        "label_files": file_inventory(LABEL_FILE_PATHS, forms_root, recorded),
    }


def forms_sha256(inventory):
    return "".join("%s  %s\n" % (f["sha256"], f["path"]) for f in inventory["forms"] if f["sha256"])


def build(fixtures_dir, out_dir, manifest_path, forms_root, forms_sha256_path=None):
    fx = load_fixtures(fixtures_dir)
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    recorded = read_shasum(forms_sha256_path) if forms_sha256_path else None
    inventory = settings_inventory(fx, manifest, forms_root, recorded)
    write_json(out_dir / "github-settings-inventory.json", inventory)
    write_json(out_dir / "github-duplicate-map.json", duplicate_map(fx, manifest, fx["meta"]["fetched_at"]))
    (out_dir / "github-forms.sha256").write_text(forms_sha256(inventory), encoding="utf-8")
    return inventory


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    src = parser.add_mutually_exclusive_group(required=True)
    src.add_argument("--repo", help="remote mode: OWNER/NAME (GET only)")
    src.add_argument("--fixtures", type=pathlib.Path, help="offline mode: recorded fixture directory")
    parser.add_argument("--record", type=pathlib.Path, help="remote mode: where to write redacted fixtures")
    parser.add_argument("--out", type=pathlib.Path, required=True)
    parser.add_argument("--manifest", type=pathlib.Path, default=DEFAULT_MANIFEST)
    parser.add_argument("--forms-root", type=pathlib.Path, default=ROOT,
                        help="checkout whose .github files are hashed (untracked forms live only in the primary)")
    parser.add_argument("--forms-sha256", type=pathlib.Path,
                        help="use hashes recorded by `shasum -a 256` in the checkout that holds the forms")
    parser.add_argument("--now", help="fetched_at override (UTC ISO 8601)")
    args = parser.parse_args(argv)
    if args.repo:
        if not args.record:
            parser.error("--repo requires --record")
        import datetime
        now = args.now or datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        record(args.repo, args.record, now)
        fixtures = args.record
    else:
        fixtures = args.fixtures
    build(fixtures, args.out, args.manifest, args.forms_root, args.forms_sha256)
    return 0


if __name__ == "__main__":
    sys.exit(main())
