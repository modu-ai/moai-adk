#!/usr/bin/env python3
"""
Enhanced TAG Ledger Sync PostToolUse Hook

Implements comprehensive ledger synchronization with:
- Ledger append operations (append-only)
- Index rebuilding for fast lookups
- Automatic correction suggestions
- Snapshot creation for recovery
- RESCIND operation handling
- Chain state updates
"""

import json
import os
import re
import shutil
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Any, Set

# Configuration paths
POLICY_PATH = Path(".moai/tag-policy.json")
LEDGER_PATH = Path(".moai/tags/ledger.jsonl")
INDEX_PATH = Path(".moai/tags/index.json")
COUNTERS_PATH = Path(".moai/tags/counters.json")
SNAPSHOTS_DIR = Path(".moai/tags/snapshots")
LOCK_PATH = Path(".moai/tags/.lock")

def load_policy() -> Dict[str, Any]:
    """Load tag policy configuration"""
    if not POLICY_PATH.exists():
        return {}
    with open(POLICY_PATH, 'r', encoding='utf-8') as f:
        return json.load(f)

def load_ledger() -> List[Dict[str, Any]]:
    """Load all ledger records"""
    if not LEDGER_PATH.exists():
        return []
    with open(LEDGER_PATH, 'r', encoding='utf-8') as f:
        return [json.loads(line) for line in f if line.strip()]

def load_index() -> Dict[str, Any]:
    """Load current tag index"""
    if not INDEX_PATH.exists():
        return {}
    with open(INDEX_PATH, 'r', encoding='utf-8') as f:
        return json.load(f)

def is_eligible_file(file_path: str, policy: Dict[str, Any]) -> bool:
    """Check if file should be processed for TAG management"""
    import fnmatch

    eligible = any(fnmatch.fnmatch(file_path, pat) for pat in policy.get("eligible", []))
    excluded = any(fnmatch.fnmatch(file_path, pat) for pat in policy.get("excluded", []))

    return eligible and not excluded

def extract_tag_from_file(file_path: str, policy: Dict[str, Any]) -> Optional[str]:
    """Extract topline TAG from file"""
    try:
        content = Path(file_path).read_text(encoding='utf-8', errors='ignore')
    except Exception:
        return None

    lines = content.splitlines()
    max_scan = policy.get("topline_rules", {}).get("max_scan_lines", 20)

    if not lines:
        return None

    # Skip shebang
    start_line = 0
    if lines[0].startswith('#!'):
        start_line = 1

    # Skip header blocks
    skip_patterns = policy.get("topline_rules", {}).get("skip_header_block", [])
    block_comments = policy.get("topline_rules", {}).get("block_comments", [])

    in_block_comment = False
    block_comment_start = None

    for i in range(start_line, min(len(lines), start_line + max_scan)):
        line = lines[i].strip()

        # Handle block comments
        if not in_block_comment:
            for j, start_marker in enumerate(block_comments):
                if start_marker in line:
                    if j % 2 == 0:  # Start of block comment
                        in_block_comment = True
                        block_comment_start = block_comments[j + 1]
                        break
            if in_block_comment:
                continue

        if in_block_comment:
            if block_comment_start in line:
                in_block_comment = False
            continue

        # Skip empty lines and header patterns
        if not line:
            continue

        if any(pattern in line.lower() for pattern in skip_patterns):
            continue

        # Look for TAG pattern
        tag_match = re.match(r'^[#/<!--\s]*@([A-Z_]+):([A-Z_]+-[0-9]{3,})', line)
        if tag_match:
            return f"@{tag_match.group(1)}:{tag_match.group(2)}"

    return None

def extract_relates_from_file(file_path: str, policy: Dict[str, Any]) -> List[str]:
    """Extract Relates: references from file"""
    try:
        content = Path(file_path).read_text(encoding='utf-8', errors='ignore')
    except Exception:
        return []

    relates_pattern = policy.get("relates", {}).get("pattern", r"^(# |// |<!-- )Relates: (.+)$")
    relates_tags = []

    for line in content.splitlines():
        match = re.match(relates_pattern, line.strip())
        if match:
            relates_part = match.group(2)
            # Extract individual TAGs
            tags = re.findall(r'@[A-Z_]+:[A-Z_]+-[0-9]{3,}', relates_part)
            relates_tags.extend(tags)

    return relates_tags

def append_to_ledger(operation: Dict[str, Any]) -> None:
    """Append operation to ledger (append-only)"""
    LEDGER_PATH.parent.mkdir(parents=True, exist_ok=True)

    # Ensure timestamp
    if "ts" not in operation:
        operation["ts"] = datetime.utcnow().isoformat() + "Z"

    with open(LEDGER_PATH, 'a', encoding='utf-8') as f:
        f.write(json.dumps(operation, ensure_ascii=False) + "\n")

def rebuild_index_from_ledger() -> Dict[str, Any]:
    """Rebuild index from ledger (for consistency and recovery)"""
    ledger = load_ledger()
    index = {}

    # Process operations in chronological order
    for entry in sorted(ledger, key=lambda x: x.get("ts", "")):
        tag_id = entry.get("id")
        if not tag_id:
            continue

        op = entry.get("op", "")
        if op == "RESCIND":
            if tag_id in index:
                index[tag_id]["state"] = "rescinded"
            continue

        if op == "DEPRECATED":
            if tag_id in index:
                index[tag_id]["state"] = "deprecated"
            continue

        if op == "MIGRATE":
            from_id = entry.get("from")
            to_id = entry.get("to")
            if from_id in index:
                index[from_id]["state"] = "migrated"
                index[from_id]["migrated_to"] = to_id
            continue

        # Handle CREATE, RESERVE, ISSUE operations
        if op in ["CREATE", "RESERVE", "ISSUE"]:
            if tag_id not in index:
                index[tag_id] = {
                    "id": tag_id,
                    "type": tag_id.split(":")[0][1:],  # Remove @ and get TYPE
                    "primary_path": entry.get("primary_path"),
                    "related_paths": entry.get("related_paths", []),
                    "links": entry.get("links", {"SPEC": [], "TEST": [], "CODE": [], "DOC": []}),
                    "state": entry.get("state", "unknown"),
                    "created": entry.get("ts"),
                    "domain": entry.get("domain", "UNKNOWN")
                }

        # Update paths and links
        if tag_id in index:
            if "paths" in entry:
                paths = entry["paths"]
                if paths:
                    # First path becomes primary if not set
                    if not index[tag_id].get("primary_path"):
                        index[tag_id]["primary_path"] = paths[0]
                    else:
                        # Additional paths go to related_paths
                        for path in paths[1:]:
                            if path not in index[tag_id].get("related_paths", []):
                                index[tag_id].setdefault("related_paths", []).append(path)

            if "links" in entry:
                index[tag_id]["links"].update(entry["links"])

            if "state" in entry and index[tag_id]["state"] not in ["migrated", "deprecated", "rescinded"]:
                index[tag_id]["state"] = entry["state"]

    return index

def create_snapshot() -> str:
    """Create a snapshot for recovery"""
    SNAPSHOTS_DIR.mkdir(parents=True, exist_ok=True)

    timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
    snapshot_path = SNAPSHOTS_DIR / f"snapshot_{timestamp}.jsonl"

    # Copy current ledger to snapshot
    if LEDGER_PATH.exists():
        shutil.copy2(LEDGER_PATH, snapshot_path)

    return str(snapshot_path)

def find_expired_reservations(index: Dict[str, Any], policy: Dict[str, Any]) -> List[Dict[str, Any]]:
    """Find expired reservations that need RESCIND operation"""
    expired = []

    block_hours = policy.get("autoplan", {}).get("expire_hours_block", 72)
    cutoff_time = datetime.utcnow() - timedelta(hours=block_hours)

    for tag_id, entry in index.items():
        if entry.get("state") == "reserved":
            created = entry.get("created")
            if created:
                try:
                    created_time = datetime.fromisoformat(created.replace("Z", "+00:00"))
                    if created_time < cutoff_time:
                        expired.append({
                            "tag_id": tag_id,
                            "domain": entry.get("domain", "UNKNOWN"),
                            "created": created
                        })
                except (ValueError, TypeError):
                    pass

    return expired

def suggest_corrections(changed_files: List[str], index: Dict[str, Any], policy: Dict[str, Any]) -> List[Dict[str, Any]]:
    """Generate automatic correction suggestions"""
    suggestions = []

    for file_path in changed_files:
        if not is_eligible_file(file_path, policy):
            continue

        tag = extract_tag_from_file(file_path, policy)
        relates = extract_relates_from_file(file_path, policy)

        # Check for missing topline when related exists
        if not tag and relates:
            suggestions.append({
                "type": "add-topline-for-related",
                "file": file_path,
                "related_tags": relates,
                "message": f"Relates 태그가 있지만 topline TAG가 없습니다: {file_path}",
                "suggestion": f"파일 상단에 TAG를 추가하거나 관련 TAG와의 연결을 확인하세요."
            })

        # Check for chain completeness
        if tag:
            tag_type = tag.split(":")[0][1:]  # Remove @ and get TYPE
            tag_domain_id = tag.split(":")[1]

            # Find corresponding SPEC
            spec_id = f"@SPEC:{tag_domain_id}"
            if tag_type != "SPEC" and spec_id not in index:
                suggestions.append({
                    "type": "missing-spec-in-chain",
                    "file": file_path,
                    "tag": tag,
                    "missing_spec": spec_id,
                    "message": f"체인이 불완전합니다: {tag}에 해당하는 SPEC가 없습니다",
                    "suggestion": f"SPEC을 생성하거나 TAG 도메인을 확인하세요."
                })

    return suggestions

def on_post_tool_use(changed_files: List[str], operation_context: Dict[str, Any]) -> Dict[str, Any]:
    """
    Main hook entry point for ledger synchronization

    Args:
        changed_files: List of changed file paths
        operation_context: Context about the operation

    Returns:
        Dict with sync results and suggestions
    """
    policy = load_policy()
    if not policy:
        return {"success": False, "error": "Policy file not found"}

    operations = []
    index = load_index()

    # Process each changed file
    for file_path in changed_files:
        if not is_eligible_file(file_path, policy):
            continue

        tag = extract_tag_from_file(file_path, policy)
        relates = extract_relates_from_file(file_path, policy)

        if tag:
            # Create or update operation
            tag_type = tag.split(":")[0][1:]  # Remove @ and get TYPE
            tag_domain_id = tag.split(":")[1]
            domain = tag_domain_id.split("-")[0]

            # Determine state
            state = "active"
            if tag_type == "SPEC":
                # Check if this is just a stub
                try:
                    content = Path(file_path).read_text(encoding='utf-8')
                    if "TBD" in content or "(목적을 입력하세요)" in content:
                        state = "reserved"
                except Exception:
                    pass

            operation = {
                "op": "CREATE" if tag not in index else "UPDATE",
                "id": tag,
                "primary_path": file_path,
                "related_paths": [],
                "links": {"SPEC": [], "TEST": [], "CODE": [], "DOC": []},
                "state": state,
                "domain": domain,
                "actor": operation_context.get("actor", "unknown")
            }

            # Add links based on tag type
            if tag_type == "SPEC":
                operation["links"]["TEST"] = [f"@TEST:{tag_domain_id}"]
                operation["links"]["CODE"] = [f"@CODE:{tag_domain_id}"]
            elif tag_type == "TEST":
                operation["links"]["SPEC"] = [f"@SPEC:{tag_domain_id}"]
                operation["links"]["CODE"] = [f"@CODE:{tag_domain_id}"]
            elif tag_type == "CODE":
                operation["links"]["SPEC"] = [f"@SPEC:{tag_domain_id}"]
                operation["links"]["TEST"] = [f"@TEST:{tag_domain_id}"]
            elif tag_type == "DOC":
                operation["links"]["SPEC"] = [f"@SPEC:{tag_domain_id}"]
                operation["links"]["TEST"] = [f"@TEST:{tag_domain_id}"]
                operation["links"]["CODE"] = [f"@CODE:{tag_domain_id}"]

            operations.append(operation)

        # Process Relates tags
        for relates_tag in relates:
            if relates_tag in index:
                # Update existing tag to include this file in related_paths
                existing_entry = index[relates_tag]
                if file_path not in existing_entry.get("related_paths", []):
                    operations.append({
                        "op": "UPDATE",
                        "id": relates_tag,
                        "related_paths": [file_path],
                        "actor": operation_context.get("actor", "unknown")
                    })

    # Append operations to ledger
    for op in operations:
        append_to_ledger(op)

    # Rebuild index for consistency
    new_index = rebuild_index_from_ledger()

    # Save updated index
    with open(INDEX_PATH, 'w', encoding='utf-8') as f:
        json.dump(new_index, f, indent=2, ensure_ascii=False)

    # Create snapshot
    snapshot_path = create_snapshot()

    # Handle expired reservations
    expired_reservations = find_expired_reservations(new_index, policy)
    rescind_operations = []

    for expired in expired_reservations:
        rescind_op = {
            "op": "RESCIND",
            "id": expired["tag_id"],
            "reason": policy.get("rescind", {}).get("reason_default", "expired"),
            "actor": "system"
        }
        rescind_operations.append(rescind_op)
        append_to_ledger(rescind_op)

    # Generate correction suggestions
    suggestions = suggest_corrections(changed_files, new_index, policy)

    # Rebuild index again after RESCIND operations
    final_index = rebuild_index_from_ledger()
    with open(INDEX_PATH, 'w', encoding='utf-8') as f:
        json.dump(final_index, f, indent=2, ensure_ascii=False)

    return {
        "success": True,
        "operations_processed": len(operations),
        "rescind_operations": len(rescind_operations),
        "suggestions": suggestions,
        "snapshot_path": snapshot_path,
        "total_tags": len(final_index),
        "active_tags": len([t for t in final_index.values() if t.get("state") == "active"]),
        "reserved_tags": len([t for t in final_index.values() if t.get("state") == "reserved"]),
        "expired_tags": len(expired_reservations)
    }

# Hook entry point
if __name__ == "__main__":
    import sys

    # Example usage for testing
    changed_files = sys.argv[1:] if len(sys.argv) > 1 else []
    operation_context = {"actor": "user"}  # Can be overridden

    result = on_post_tool_use(changed_files, operation_context)

    print(f"✅ TAG 원장 동기화 완료:")
    print(f"  - 처리된 작업: {result['operations_processed']}")
    print(f"  - 만료 처리: {result['rescind_operations']}")
    print(f"  - 전체 TAG: {result['total_tags']}")
    print(f"  - 활성 TAG: {result['active_tags']}")
    print(f"  - 예약 TAG: {result['reserved_tags']}")
    print(f"  - 스냅샷: {result['snapshot_path']}")

    if result['suggestions']:
        print(f"\n💡 수정 제안:")
        for suggestion in result['suggestions']:
            print(f"  - {suggestion.get('message', 'Unknown suggestion')}")
            print(f"    → {suggestion.get('suggestion', 'No suggestion available')}")