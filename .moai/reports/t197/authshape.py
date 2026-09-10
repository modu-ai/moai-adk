"""Print the SHAPE of <CODEX_HOME>/auth.json plus the two non-secret values the
classifier needs. No token or key value is ever printed."""
import json
import os

home = os.environ.get("CODEX_HOME") or os.path.expanduser("~/.codex")
with open(os.path.join(home, "auth.json")) as fh:
    d = json.load(fh)


def shape(o, depth=0):
    if isinstance(o, dict):
        return {k: (shape(v, depth + 1) if depth < 1 else type(v).__name__) for k, v in o.items()}
    return type(o).__name__


print(json.dumps(shape(d), indent=1))
print("auth_mode =", repr(d.get("auth_mode")))
print("OPENAI_API_KEY populated:", bool(d.get("OPENAI_API_KEY")))
print("tokens present:", isinstance(d.get("tokens"), dict))
print("tokens non-empty values:", sum(1 for v in (d.get("tokens") or {}).values() if v))
