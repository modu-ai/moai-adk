"""Source-manifest property tests (C-100, SPEC-20261119-property-tests).

The writer is Python (`tools/source_manifest.py`) and so is the merge driver
(`tools/merge_driver_manifest.py`), so the properties run in process: `render` is idempotent
and permutation-invariant, and the driver's merged bytes are the same whatever order the
entries of each side were written in. Every case is seeded; a failure names the seed and the
case index, replayable with PROP_SEED=<seed> PROP_ONLY=<i>.
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
import merge_driver_manifest as mdm  # noqa: E402

# `merged_manifest` reads the module through the driver's own global, which `drive` loads on
# its first run; a test loads it the same way once.
mdm.sm = mdm.load_source_manifest()

# The driver's git-settings probe reads two `git config` values that cannot change while one
# test run lasts; the property measures merged bytes, not probe cost, so this file answers the
# probe once and reuses the answer (the tool itself is untouched).
_probe_answered = {}
_probe = mdm.refuse_settings_git_merges_differently


def _probe_once():
    if "ok" not in _probe_answered:
        _probe()
        _probe_answered["ok"] = True


mdm.refuse_settings_git_merges_differently = _probe_once

MIN_CASES = 1000
DEFAULT_SEED = 0x0C10_2026_1119
MANIFEST_PATH = "skills/moai/references/source-manifest.json"
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

    def test_the_driver_merges_the_same_bytes_for_any_entry_order(self):
        def one(rng):
            names = [f"docs/{n}.md" for n in ("a", "b", "c")]
            base_entries = entries_for(rng, names)
            # The two sides add disjoint names, so no entry is ever added twice with
            # different contents (a merge the driver rightly refuses).
            extra = [f"docs/ours-{rng.choice(ALPHABET)}{i}.md" for i in range(rng.randint(0, 3))]
            other_extra = [f"docs/theirs-{rng.choice(ALPHABET)}{i}.md" for i in range(rng.randint(0, 3))]
            side_entries = dict(base_entries, **{n: {"sha256": sha(rng), "origin": origin(n)} for n in extra})
            their_entries = dict(base_entries, **{n: {"sha256": sha(rng), "origin": origin(n)} for n in other_extra})
            base_text = sm.render(manifest_dict(rng, base_entries))
            merged = []
            for _ in range(3):  # three different entry orders of the same two sides
                merged.append(
                    mdm.merged_manifest(
                        base_text,
                        sm.render(manifest_dict(rng, side_entries)),
                        sm.render(manifest_dict(rng, their_entries)),
                        MANIFEST_PATH,
                    )
                )
            self.assertEqual(merged[0], merged[1], "entry order changed the merged bytes")
            self.assertEqual(merged[0], merged[2], "entry order changed the merged bytes")
            canonical = sm.render(json.loads(merged[0]))
            self.assertEqual(merged[0], canonical, "merge is not the writer's canonical bytes")

        run_property("driver permutation", MIN_CASES, one)


if __name__ == "__main__":
    unittest.main()
