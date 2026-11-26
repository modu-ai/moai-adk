"""Source-manifest property tests (C-100, SPEC-20261119-property-tests).

The writer is Python (`tools/source_manifest.py`), so the property runs in process: `render`
is idempotent and permutation-invariant. The merge driver is the command
`moai manifest-merge-driver`; its order property is `tests/manifest_merge_driver_cli.rs`. Every
case is seeded; a failure names the seed and the case index, replayable with PROP_SEED=<seed>
PROP_ONLY=<i>.
"""

import json
import os
import random
import sys
import time
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "tools"))

import source_manifest as sm  # noqa: E402

MIN_CASES = 1000
DEFAULT_SEED = 0x0C10_2026_1119
ALPHABET = "abcdefghijklmnopqrstuvwxyz0123456789-_."


def run_property(name, cases, fn):
    """The Python twin of the Rust harness: seeded cases, 30 s budget, seed on failure."""
    seed = int(os.environ.get("PROP_SEED", str(DEFAULT_SEED)), 0)
    only = os.environ.get("PROP_ONLY")
    only = int(only) if only is not None else None
    start = time.monotonic()
    for i in range(cases):
        if only is not None and only != i:
            continue
        rng = random.Random(seed ^ (i * 0x100000001B3))
        try:
            fn(rng)
        except AssertionError as exc:
            raise AssertionError(
                f"property {name} failed at case {i} of {cases} "
                f"(seed {seed}): {exc} — replay with PROP_SEED={seed} PROP_ONLY={i}"
            ) from None
    elapsed = time.monotonic() - start
    assert elapsed < 30, f"property {name} ran {cases} cases in over 30 s (seed {seed})"


def origin(name, reason="prop"):
    return {"kind": "design-spec", "template_id": name, "adoption": "new", "reason": reason}


def sha(rng):
    return "".join(rng.choice("0123456789abcdef") for _ in range(64))


def manifest_dict(rng, entries):
    """A manifest whose `files` entries (`name` → entry) are inserted in a generated order."""
    keys = list(entries)
    rng.shuffle(keys)
    return {"schema_version": 1, "covers": "prop", "files": {k: entries[k] for k in keys}}


def entries_for(rng, names):
    """The entries themselves, drawn once, so permutations hold the same content."""
    return {name: {"sha256": sha(rng), "origin": origin(name)} for name in names}


class ManifestProperties(unittest.TestCase):
    def test_render_is_idempotent_and_permutation_invariant(self):
        def one(rng):
            names = [
                f"{rng.choice(ALPHABET)}{''.join(rng.choice(ALPHABET) for _ in range(rng.randint(0, 8)))}.md"
                for _ in range(rng.randint(1, 8))
            ]
            names = list(dict.fromkeys(names))
            entries = entries_for(rng, names)  # drawn once: both orders hold this content
            base = manifest_dict(rng, entries)
            other = manifest_dict(rng, entries)
            self.assertEqual(sm.render(base), sm.render(other), f"permutation differs: {names}")
            once = sm.render(base)
            twice = sm.render(json.loads(once))
            self.assertEqual(once, twice, f"re-render moved a byte: {names}")
            keys = list(json.loads(once)["files"])
            self.assertEqual(keys, sorted(keys), f"files not sorted: {names}")

        run_property("render idempotent", MIN_CASES, one)


if __name__ == "__main__":
    unittest.main()
