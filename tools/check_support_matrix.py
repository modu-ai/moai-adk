#!/usr/bin/env python3
"""AC-08D-1 checker (Issue #2015): every support-matrix row marked supported
must cite a passed 08-C receipt; everything else must read not_run/failed/blocked.

Usage: python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/

Receipts are JSONL (one JSON object per line, schemas/e2e-receipt.schema.json
fields) read from receipts.jsonl plus any *.json files in the receipt directory.
A matrix row is "supported" when its result cell reads passed/certified. Each
supported row must carry a receipt reference whose `ac` matches the row's AC
cell and whose result is passed — a support claim without a passed receipt
exits 1 (RL-10). Rows may also cite nothing at all: those must be not_run,
failed or blocked.

Exit codes: 0 = consistent, 1 = violation or unreadable input, 2 = usage.
"""

import json
import re
import sys
from pathlib import Path

# A row's receipt reference: docs/e2e-receipts/... or receipts.jsonl mention.
RECEIPT_REF = re.compile(r"receipts\.jsonl|docs/e2e-receipts/|\.json\b")
AC_CELL = re.compile(r"AC-\w+-\d+|HA-\d+")

VERDICTS = ("passed", "certified", "not_run", "failed", "blocked", "probed")


def load_receipts(receipt_dir: Path) -> dict:
    """ac -> list of receipt records, from receipts.jsonl and any *.json files."""
    receipts: dict = {}
    if not receipt_dir.is_dir():
        print(f"receipt directory not found: {receipt_dir}", file=sys.stderr)
        raise SystemExit(1)
    sources = [receipt_dir / "receipts.jsonl"] + sorted(receipt_dir.glob("*.json"))
    for path in sources:
        if not path.is_file():
            continue
        for lineno, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            line = line.strip()
            if not line:
                continue
            try:
                record = json.loads(line)
            except json.JSONDecodeError as exc:
                print(f"{path}:{lineno}: not JSON: {exc}", file=sys.stderr)
                raise SystemExit(1)
            ac = record.get("ac")
            if not ac:
                continue
            receipts.setdefault(ac, []).append(record)
    return receipts


def parse_rows(matrix_path: Path) -> list:
    """Rows of the markdown tables whose result cell names a support verdict."""
    rows = []
    for lineno, line in enumerate(
        matrix_path.read_text(encoding="utf-8").splitlines(), 1
    ):
        line = line.strip()
        if not line.startswith("|"):
            continue
        cells = [c.strip() for c in line.strip("|").split("|")]
        if len(cells) < 2 or set(cells[0]) <= {"-", ":", " "}:
            continue  # separator row
        if cells[0] in ("host", "OS/arch", "os/arch") or cells[0].startswith("OS/arch"):
            continue  # header rows
        rows.append((lineno, cells, line))
    return rows


def verdict_of(cells: list):
    """The result cell: the LAST cell whose whole value is a verdict word. Prose
    that merely mentions a verdict word never counts."""
    for cell in reversed(cells):
        if cell.lower() in VERDICTS:
            return cell.lower()
    return None


def main(argv: list) -> int:
    if len(argv) != 3:
        print(__doc__, file=sys.stderr)
        return 2
    matrix_path = Path(argv[1])
    receipt_dir = Path(argv[2])
    if not matrix_path.is_file():
        print(f"matrix not found: {matrix_path}", file=sys.stderr)
        return 1
    receipts = load_receipts(receipt_dir)

    violations = []
    support_rows = 0
    passed_receipts = 0
    for ac in receipts:
        if any(r.get("result") == "passed" for r in receipts[ac]):
            passed_receipts += 1

    for lineno, cells, line in parse_rows(matrix_path):
        verdict = verdict_of(cells)
        if verdict is None:
            continue  # descriptive row (notes, vocabulary) — not a verdict row
        has_ref = bool(RECEIPT_REF.search(line))
        ac_match = AC_CELL.search(line)
        if verdict in ("passed", "certified"):
            support_rows += 1
            if not has_ref:
                violations.append(
                    f"line {lineno}: supported row without a receipt reference"
                )
                continue
            if not ac_match:
                violations.append(
                    f"line {lineno}: supported row without an AC identifier"
                )
                continue
            ac = ac_match.group(0)
            records = receipts.get(ac, [])
            if not any(r.get("result") == "passed" for r in records):
                violations.append(
                    f"line {lineno}: supported row cites {ac} but no passed receipt exists"
                )
        elif verdict == "probed":
            continue  # probed is the fixture-measured status; the verdict cell decides
        else:
            if has_ref and ac_match:
                ac = ac_match.group(0)
                records = receipts.get(ac, [])
                if any(r.get("result") == "passed" for r in records):
                    violations.append(
                        f"line {lineno}: {verdict} row but receipt {ac} is passed"
                    )

    print(f"supported rows: {support_rows}")
    print(f"passed receipts: {passed_receipts} ({sum(len(v) for v in receipts.values())} receipts total)")
    if support_rows != passed_receipts:
        violations.append(
            f"RL-10: supported row count ({support_rows}) != passed receipt count ({passed_receipts})"
        )
    for violation in violations:
        print(f"VIOLATION: {violation}", file=sys.stderr)
    return 1 if violations else 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
