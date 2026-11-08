"""Preview a top-level Claude settings merge without reading or writing files."""
import hashlib
import copy
import json

_MISSING = object()
_ALLOWED = frozenset({"outputStyle", "statusLine"})


def _same(left, right):
    if left is _MISSING or right is _MISSING:
        return left is right
    return json.dumps(left, sort_keys=True, ensure_ascii=False, allow_nan=False) == json.dumps(
        right, sort_keys=True, ensure_ascii=False, allow_nan=False
    )


def plan_settings(baseline, current, desired):
    """Return a candidate and conflicts; baseline contains previously owned keys only."""
    if not all(isinstance(value, dict) for value in (baseline, current, desired)):
        raise ValueError("settings snapshots must be objects")
    if not set(baseline).issubset(_ALLOWED) or not set(desired).issubset(_ALLOWED):
        raise ValueError("only explicitly managed top-level keys are supported")
    # Validate the whole snapshots before returning a candidate.
    for snapshot in (baseline, current, desired):
        json.dumps(snapshot, allow_nan=False)
    candidate = copy.deepcopy(current)
    conflicts = []
    for key in sorted(set(baseline) | set(desired)):
        old = baseline.get(key, _MISSING)
        now = current.get(key, _MISSING)
        new = desired.get(key, _MISSING)
        if _same(now, new) or _same(new, old):
            continue
        if _same(now, old):
            if new is _MISSING:
                candidate.pop(key, None)
            else:
                candidate[key] = copy.deepcopy(new)
        else:
            conflicts.append(key)
    return (None if conflicts else candidate), conflicts


from pathlib import Path
import os
import tempfile


def read_bytes(path):
    return path.read_bytes() if path.exists() else None


def replace_checked(path, before, after):
    if path.is_symlink() or read_bytes(path) != before:
        raise ValueError("concurrent settings change")
    if before is None:
        with path.open("xb") as stream:
            stream.write(after)
        return
    fd, name = tempfile.mkstemp(prefix=".moai-update-", dir=path.parent)
    try:
        with os.fdopen(fd, "wb") as stream:
            stream.write(after)
            os.fchmod(stream.fileno(), path.stat().st_mode & 0o777)
            stream.flush()
            os.fsync(stream.fileno())
        if path.is_symlink() or read_bytes(path) != before:
            raise ValueError("concurrent settings change")
        os.replace(name, path)
    finally:
        if os.path.exists(name):
            os.unlink(name)


def encode(value):
    return (json.dumps(value, ensure_ascii=False, indent=2, allow_nan=False) + "\n").encode()


def prepare(root, wanted, parse):
    root = root.resolve(strict=True)
    settings = root / ".claude/settings.json"
    receipt = root / ".moai-cli/claude-settings-owner.json"
    for path in (settings, receipt):
        if path.is_symlink() or path.parent.is_symlink():
            raise ValueError("symlink in settings transaction")
    before = read_bytes(settings)
    current = parse(before) if before is not None else {}
    raw = read_bytes(receipt)
    saved = parse(raw) if raw is not None else {"schema_version": 1, "state": "committed", "baseline": {}}
    validate_receipt(saved)
    if saved.get("schema_version") != 1 or saved.get("state") != "committed":
        raise ValueError("pending settings transaction: inspect receipt before recovery")
    baseline = saved["baseline"]
    # Identical pre-existing values do not transfer ownership to MoAI.
    owned = set(baseline) | (set(wanted) - set(current))
    for key in set(wanted) - owned:
        if not _same(current.get(key, _MISSING), wanted[key]):
            raise ValueError("unowned settings conflict: " + key)
    desired = {key: value for key, value in wanted.items() if key in owned}
    candidate, conflicts = plan_settings(baseline, current, desired)
    if conflicts:
        raise ValueError("settings conflict: " + ", ".join(conflicts))
    after = before if _same(candidate, current) and before is not None else encode(candidate)
    return settings, receipt, before, raw, after, baseline, desired


def apply(plan, after_settings=None):
    settings, receipt, before, raw, after, baseline, desired = plan
    committed = encode({"schema_version": 1, "state": "committed", "baseline": desired})
    if before == after and (raw == committed or not desired):
        return
    pending = encode({"schema_version": 1, "state": "pending", "baseline": baseline, "desired": desired, "before_sha256": digest(before), "after_sha256": digest(after)})
    receipt.parent.mkdir(parents=True, exist_ok=True)
    replace_checked(receipt, raw, pending)
    settings.parent.mkdir(parents=True, exist_ok=True)
    if before != after:
        replace_checked(settings, before, after)
    if after_settings is not None:
        after_settings()
    if settings.parent.is_symlink() or settings.is_symlink() or read_bytes(settings) != after:
        raise ValueError("settings changed before receipt commit; preserve pending transaction")
    replace_checked(receipt, pending, committed)


def digest(data):
    return hashlib.sha256(data).hexdigest() if data is not None else None


def validate_receipt(saved):
    if not isinstance(saved, dict) or type(saved.get("schema_version")) is not int or saved["schema_version"] != 1:
        raise ValueError("invalid ownership receipt schema")
    if saved.get("state") not in ("pending", "committed"):
        raise ValueError("invalid receipt state")
    for field in (["baseline", "desired"] if saved["state"] == "pending" else ["baseline"]):
        value = saved.get(field)
        if not isinstance(value, dict) or not set(value).issubset(_ALLOWED):
            raise ValueError("invalid receipt ownership")
    for field in ("before_sha256", "after_sha256"):
        value = saved.get(field)
        if value is not None and (not isinstance(value, str) or len(value) != 64 or any(c not in "0123456789abcdef" for c in value)):
            raise ValueError("invalid receipt digest")


def recovery_plan(root, parse):
    """Read-only diagnostics: never infer rollback authority from matching values."""
    root = root.resolve(strict=True)
    settings = root / ".claude/settings.json"
    receipt = root / ".moai-cli/claude-settings-owner.json"
    for path in (settings, receipt):
        for ancestor in [path, *path.parents]:
            if ancestor == root.parent:
                break
            if ancestor.is_symlink():
                raise ValueError("symlink in recovery path")
    raw = read_bytes(receipt)
    if raw is None:
        return {"state": "unowned", "action": "preserve", "automatic_recovery": False}
    saved = parse(raw)
    validate_receipt(saved)
    if saved["state"] == "committed":
        return {"state": "committed", "action": "none", "automatic_recovery": False}
    current = digest(read_bytes(settings))
    if "before_sha256" not in saved or "after_sha256" not in saved:
        state = "legacy_pending_unknown"
    elif current == saved["after_sha256"]:
        state = "matches_applied_snapshot"
    elif current == saved["before_sha256"]:
        state = "matches_before_snapshot"
    else:
        state = "conflict_preserve_current"
    return {"state": state, "action": "manual_review", "automatic_recovery": False,
            "reason": "No full before snapshot or authenticated ownership; do not overwrite current settings or delete the receipt."}
