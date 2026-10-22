#!/usr/bin/env python3
"""Refuse a diagnostic success without a named run, pass and package pass."""

import argparse
import json
from pathlib import Path
import sys


def verify(path, package, test):
    runs, passes, package_passes = [], [], []
    with path.open(encoding="utf-8-sig") as stream:
        for number, line in enumerate(stream, 1):
            row = json.loads(line)
            if not isinstance(row, dict) or row.get("Package") != package:
                raise ValueError(f"event {number} does not belong to the expected package")
            action, name = row.get("Action"), row.get("Test")
            if action in {"fail", "skip"}:
                raise ValueError(f"event {number} records {action}: {name or package}")
            if name == test and action == "run":
                runs.append(number)
            if name == test and action == "pass":
                passes.append(number)
            if not name and action == "pass":
                package_passes.append(number)
    if not (len(runs) == len(passes) == len(package_passes) == 1):
        raise ValueError("expected exactly one named run, named pass and package pass")
    if not runs[0] < passes[0] < package_passes[0]:
        raise ValueError("named run, named pass and package pass are out of order")
    return {"package": package, "test": test, "run": 1, "pass": 1, "package_pass": 1}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--events", type=Path, required=True)
    parser.add_argument("--package", required=True)
    parser.add_argument("--test", required=True)
    args = parser.parse_args()
    try:
        receipt = verify(args.events, args.package, args.test)
    except (OSError, UnicodeError, ValueError) as error:
        print(f"named-test census: {error}", file=sys.stderr)
        return 1
    print(json.dumps(receipt, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
