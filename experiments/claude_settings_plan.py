"""Preview a top-level Claude settings merge without reading or writing files."""
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
