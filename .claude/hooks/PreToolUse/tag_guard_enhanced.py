#!/usr/bin/env python3
"""
Enhanced TAG Guard PreToolUse Hook

Implements comprehensive TAG validation with:
- Autoplan Reserve for code-first scenarios
- Duplicate primary detection
- Expiration handling (24h warn, 72h block)
- Batch Guard for large changes
- Branch conflict detection
- Topline parsing with various comment styles
"""

import json
import os
import re
import time
import fcntl
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Any

# Configuration paths
POLICY_PATH = Path(".moai/tag-policy.json")
LEDGER_PATH = Path(".moai/tags/ledger.jsonl")
INDEX_PATH = Path(".moai/tags/index.json")
COUNTERS_PATH = Path(".moai/tags/counters.json")
LOCK_PATH = Path(".moai/tags/.lock")

def load_policy() -> Dict[str, Any]:
    """Load tag policy configuration"""
    if not POLICY_PATH.exists():
        return {}
    with open(POLICY_PATH, 'r', encoding='utf-8') as f:
        return json.load(f)

def load_ledger() -> List[Dict[str, Any]]:
    """Load ledger records"""
    if not LEDGER_PATH.exists():
        return []
    with open(LEDGER_PATH, 'r', encoding='utf-8') as f:
        return [json.loads(line) for line in f if line.strip()]

def load_index() -> Dict[str, Any]:
    """Load tag index for O(1) lookups"""
    if not INDEX_PATH.exists():
        return {}
    with open(INDEX_PATH, 'r', encoding='utf-8') as f:
        return json.load(f)

def is_eligible_file(file_path: str, policy: Dict[str, Any]) -> bool:
    """Check if file should be processed for TAG validation"""
    import fnmatch

    # Check if file matches eligible patterns
    eligible = any(fnmatch.fnmatch(file_path, pat) for pat in policy.get("eligible", []))

    # Check if file matches excluded patterns
    excluded = any(fnmatch.fnmatch(file_path, pat) for pat in policy.get("excluded", []))

    return eligible and not excluded

def extract_domain_from_path(file_path: str, policy: Dict[str, Any]) -> Optional[str]:
    """Extract domain from file path using policy domain mapping"""
    # Check domain mappings first
    for mapping in policy.get("domains_map", []):
        import fnmatch
        if fnmatch.fnmatch(file_path, mapping["glob"]):
            return mapping["domain"]

    # Default extraction from path
    parts = Path(file_path).parts
    for part in parts:
        if part.upper() in ["AUTH", "USER", "PAY", "CORE", "DOCS"]:
            return part.upper()

    return "CORE"  # Default domain

def find_tag_topline(file_content: str, policy: Dict[str, Any]) -> Optional[str]:
    """Find TAG in file topline respecting various comment styles"""
    lines = file_content.splitlines()
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

def is_code_file(file_path: str) -> bool:
    """Check if file is a code file"""
    code_extensions = ['.py', '.js', '.ts', '.go', '.java', '.cpp', '.c', '.h']
    return any(file_path.endswith(ext) for ext in code_extensions)

def acquire_lock():
    """Acquire file lock for thread safety"""
    LOCK_PATH.parent.mkdir(parents=True, exist_ok=True)
    lock_file = open(LOCK_PATH, 'w')
    fcntl.flock(lock_file.fileno(), fcntl.LOCK_EX)
    return lock_file

def release_lock(lock_file):
    """Release file lock"""
    lock_file.close()

def reserve_spec_id(domain: str, policy: Dict[str, Any]) -> Tuple[str, str]:
    """Reserve a new SPEC ID and create stub"""
    lock_file = acquire_lock()

    try:
        # Load counters
        counters = {}
        if COUNTERS_PATH.exists():
            with open(COUNTERS_PATH, 'r') as f:
                counters = json.load(f)

        # Get next number
        next_num = counters.get(domain, 0) + 1
        counters[domain] = next_num

        # Save counters
        with open(COUNTERS_PATH, 'w') as f:
            json.dump(counters, f, indent=2)

        # Create SPEC ID
        spec_id = f"@SPEC:{domain}-{next_num:03d}"

        # Create spec directory and stub
        spec_dir = Path(f".moai/specs/{domain}-{next_num:03d}")
        spec_dir.mkdir(parents=True, exist_ok=True)

        # Load template and create stub
        template_path = Path(policy.get("autoplan", {}).get("spec_stub_template", ".moai/templates/spec_stub.md"))
        template_content = ""
        if template_path.exists():
            template_content = template_path.read_text(encoding='utf-8')

        # Fill template
        now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        stub_content = template_content.replace("{{ID}}", spec_id)
        stub_content = stub_content.replace("{{TITLE}}", "TBD")
        stub_content = stub_content.replace("{{PURPOSE}}", "(목적을 입력하세요)")
        stub_content = stub_content.replace("{{CRITERION_1}}", "(수용 기준 1)")
        stub_content = stub_content.replace("{{CRITERION_2}}", "(수용 기준 2)")
        stub_content = stub_content.replace("{{CRITERION_3}}", "(수용 기준 3)")
        stub_content = stub_content.replace("{{NOTES}}", "(참고사항을 입력하세요)")
        stub_content = stub_content.replace("{{DATETIME}}", now)
        stub_content = stub_content.replace("{{ACTOR}}", "alfred")

        (spec_dir / "spec.md").write_text(stub_content, encoding='utf-8')

        # Update ledger
        ledger_entry = {
            "ts": datetime.utcnow().isoformat() + "Z",
            "op": "RESERVE",
            "id": spec_id,
            "actor": "alfred",
            "paths": [str(spec_dir / "spec.md")],
            "state": "reserved",
            "domain": domain,
            "reserved_until": (datetime.utcnow() + timedelta(hours=policy.get("autoplan", {}).get("expire_hours_block", 72))).isoformat() + "Z"
        }

        with open(LEDGER_PATH, 'a', encoding='utf-8') as f:
            f.write(json.dumps(ledger_entry, ensure_ascii=False) + "\n")

        return spec_id, str(spec_dir / "spec.md")

    finally:
        release_lock(lock_file)

def is_duplicate_primary(tag_id: str, file_path: str, index: Dict[str, Any]) -> bool:
    """Check if tag would create duplicate primary"""
    if tag_id not in index:
        return False

    existing = index[tag_id]
    if existing.get("type") != tag_id.split(":")[0][1:]:  # TYPE part
        return False

    existing_primary = existing.get("primary_path")
    if existing_primary and existing_primary != file_path:
        return True

    return False

def is_expired_reservation(tag_id: str, index: Dict[str, Any], policy: Dict[str, Any]) -> Tuple[bool, str]:
    """Check if reservation is expired"""
    if tag_id not in index:
        return False, ""

    entry = index[tag_id]
    if entry.get("state") != "reserved":
        return False, ""

    reserved_until = entry.get("reserved_until")
    if not reserved_until:
        return False, ""

    try:
        expiry_time = datetime.fromisoformat(reserved_until.replace("Z", "+00:00"))
        now = datetime.utcnow()

        if now > expiry_time:
            return True, "expired"

        # Check if within warning period
        warn_hours = policy.get("autoplan", {}).get("expire_hours_warn", 24)
        warn_time = expiry_time - timedelta(hours=warn_hours)

        if now > warn_time:
            return True, "warning"

    except (ValueError, TypeError):
        pass

    return False, ""

def on_pre_tool_use(changed_files: List[str], operation_context: Dict[str, Any]) -> Dict[str, Any]:
    """
    Main hook entry point for TAG validation

    Args:
        changed_files: List of changed file paths
        operation_context: Context about the operation

    Returns:
        Dict with validation results
    """
    policy = load_policy()
    if not policy:
        return {"block": False, "violations": [], "actions": []}

    mode = operation_context.get("mode", policy.get("policy_mode", "balanced"))

    # Check for batch guard
    file_threshold = policy.get("batch_guard", {}).get("file_threshold", 25)
    if len(changed_files) > file_threshold:
        return {
            "block": mode == "strict",
            "violations": [("batch-guard", len(changed_files))],
            "actions": [{
                "type": "batch-guard-warning",
                "message": f"대량 변경 감지 ({len(changed_files)}개 파일). 임계치 {file_threshold}개 초과. 변경 목록을 검토하고 /alfred:3-sync 전에 수동 확인하세요."
            }]
        }

    index = load_index()
    violations = []
    actions = []

    for file_path in changed_files:
        if not is_eligible_file(file_path, policy):
            continue

        try:
            file_content = Path(file_path).read_text(encoding='utf-8', errors='ignore')
        except Exception:
            continue

        # Find topline tag
        tag = find_tag_topline(file_content, policy)

        # Handle code-first scenarios
        if is_code_file(file_path) and not tag:
            if policy.get("autoplan", {}).get("reserve_on_code_first"):
                domain = extract_domain_from_path(file_path, policy)

                if mode == "strict":
                    violations.append(("code-first-require-spec", file_path, domain))
                    actions.append({
                        "type": "require-spec",
                        "file": file_path,
                        "domain": domain,
                        "message": f"코드 파일이 SPEC 없이 생성되었습니다. /alfred:1-plan --reserve를 실행하여 SPEC를 먼저 생성하세요."
                    })
                else:
                    # Autoplan reserve
                    spec_id, spec_path = reserve_spec_id(domain, policy)
                    actions.append({
                        "type": "autoplan-reserve",
                        "file": file_path,
                        "spec_id": spec_id,
                        "spec_path": spec_path,
                        "domain": domain,
                        "message": f"코드 우선 생성 감지 → {spec_id} 예약 및 {spec_path} 생성. 파일 상단에 '# @CODE:{domain}-{spec_id.split('-')[1]}' 추가하세요."
                    })
                    violations.append(("code-first-autoplan", file_path, domain))
            continue

        if not tag:
            continue

        # Check for duplicate primary
        if is_duplicate_primary(tag, file_path, index):
            violations.append(("duplicate-primary", file_path, tag))
            existing_primary = index[tag].get("primary_path")
            actions.append({
                "type": "duplicate-primary-fix",
                "file": file_path,
                "tag": tag,
                "existing_primary": existing_primary,
                "message": f"동일 ID의 primary 파일이 이미 존재합니다: {existing_primary}. 이 파일은 Relates를 사용하세요.",
                "fix": f"파일 상단의 TAG를 제거하고 첫 줄에 '# Relates: {tag}'를 추가하세요."
            })

        # Check for expired reservations
        expired, status = is_expired_reservation(tag, index, policy)
        if expired:
            if status == "expired":
                violations.append(("expired-reservation", file_path, tag))
                actions.append({
                    "type": "expired-reservation",
                    "file": file_path,
                    "tag": tag,
                    "message": f"예약된 TAG가 만료되었습니다: {tag}. /alfred:1-plan으로 새 SPEC를 생성하세요."
                })
            elif status == "warning":
                violations.append(("expiring-reservation", file_path, tag))
                actions.append({
                    "type": "expiring-reservation",
                    "file": file_path,
                    "tag": tag,
                    "message": f"예약된 TAG가 24시간 내에 만료됩니다: {tag}. 빨리 SPEC를 완성하세요."
                })

    # Determine if should block
    should_block = mode == "strict" and any(
        v[0] in ["duplicate-primary", "expired-reservation", "code-first-require-spec", "batch-guard"]
        for v in violations
    )

    return {
        "block": should_block,
        "violations": violations,
        "actions": actions,
        "policy_mode": mode,
        "processed_files": len([f for f in changed_files if is_eligible_file(f, policy)])
    }

# Hook entry point
if __name__ == "__main__":
    import sys

    # Example usage for testing
    changed_files = sys.argv[1:] if len(sys.argv) > 1 else []
    operation_context = {"mode": "balanced"}  # Can be overridden

    result = on_pre_tool_use(changed_files, operation_context)

    if result["block"]:
        print("🚫 TAG 검증 차단:")
        for violation in result["violations"]:
            print(f"  - {violation}")
        print("\n수정 조치:")
        for action in result["actions"]:
            print(f"  - {action.get('message', 'Unknown action')}")
        sys.exit(1)
    else:
        if result["violations"]:
            print("⚠️ TAG 경고:")
            for violation in result["violations"]:
                print(f"  - {violation}")

        if result["actions"]:
            print("\n권장 조치:")
            for action in result["actions"]:
                print(f"  - {action.get('message', 'Unknown action')}")

        print(f"\n✅ {result['processed_files']}개 파일 처리 완료 (모드: {result['policy_mode']})")