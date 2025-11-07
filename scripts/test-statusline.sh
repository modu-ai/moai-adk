#!/bin/bash
# Test script for MoAI-ADK Statusline
# Claude Code Statusline functionality testing
#
# @CODE:STATUSLINE-TEST-001

set -e

echo "=========================================="
echo "  MoAI-ADK Statusline Integration Test"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function for test output
test_result() {
    local test_name=$1
    local exit_code=$2

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        ((TESTS_FAILED++))
    fi
}

# Test 1: Module imports
echo "${BLUE}Test 1: Module Imports${NC}"
python3 << 'EOF'
try:
    from moai_adk.statusline.renderer import StatuslineRenderer, StatuslineData
    from moai_adk.statusline.git_collector import GitCollector, GitInfo
    from moai_adk.statusline.metrics_tracker import MetricsTracker
    from moai_adk.statusline.alfred_detector import AlfredDetector, AlfredTask
    from moai_adk.statusline.version_reader import VersionReader
    from moai_adk.statusline.update_checker import UpdateChecker, UpdateInfo
    from moai_adk.statusline.main import build_statusline_data
    print("All modules imported successfully")
except ImportError as e:
    print(f"Import error: {e}")
    exit(1)
EOF
test_result "Module imports" $?
echo ""

# Test 2: Compact mode rendering
echo "${BLUE}Test 2: Compact Mode Rendering${NC}"
python3 << 'EOF'
from moai_adk.statusline.renderer import StatuslineRenderer, StatuslineData

renderer = StatuslineRenderer()
data = StatuslineData(
    model="H 4.5",
    duration="5m",
    directory="MoAI-ADK",
    version="0.20.1",
    branch="feature/SPEC-001",
    git_status="+2 M1 ?0",
    active_task="[PLAN]",
    update_available=False,
    latest_version=None
)

statusline = renderer.render(data, mode="compact")
print(f"Output: {statusline}")
print(f"Length: {len(statusline)} chars")

assert len(statusline) <= 80, f"Compact mode must be ≤80 chars, got {len(statusline)}"
assert "|" in statusline, "Separator not found"
assert "MoAI-ADK" in statusline, "Directory not found"
print("Compact mode test passed")
EOF
test_result "Compact mode rendering" $?
echo ""

# Test 3: Extended mode rendering
echo "${BLUE}Test 3: Extended Mode Rendering${NC}"
python3 << 'EOF'
from moai_adk.statusline.renderer import StatuslineRenderer, StatuslineData

renderer = StatuslineRenderer()
data = StatuslineData(
    model="Haiku 4.5",
    duration="1h 30m",
    directory="MoAI-ADK",
    version="0.20.1",
    branch="feature/SPEC-AUTH-001",
    git_status="+5 M3 ?2",
    active_task="[RUN-GREEN]",
    update_available=True,
    latest_version="0.21.0"
)

statusline = renderer.render(data, mode="extended")
print(f"Output: {statusline}")
print(f"Length: {len(statusline)} chars")

assert len(statusline) <= 120, f"Extended mode must be ≤120 chars, got {len(statusline)}"
assert "Haiku" in statusline, "Full model name not found"
print("Extended mode test passed")
EOF
test_result "Extended mode rendering" $?
echo ""

# Test 4: Minimal mode rendering
echo "${BLUE}Test 4: Minimal Mode Rendering${NC}"
python3 << 'EOF'
from moai_adk.statusline.renderer import StatuslineRenderer, StatuslineData

renderer = StatuslineRenderer()
data = StatuslineData(
    model="H",
    duration="5m",
    directory="project",
    version="0.20.1",
    branch="feature/SPEC",
    git_status="+2M",
    active_task="",
    update_available=False,
    latest_version=None
)

statusline = renderer.render(data, mode="minimal")
print(f"Output: {statusline}")
print(f"Length: {len(statusline)} chars")

assert len(statusline) <= 40, f"Minimal mode must be ≤40 chars, got {len(statusline)}"
print("Minimal mode test passed")
EOF
test_result "Minimal mode rendering" $?
echo ""

# Test 5: Git collector (no errors)
echo "${BLUE}Test 5: Git Information Collection${NC}"
python3 << 'EOF'
from moai_adk.statusline.git_collector import GitCollector

collector = GitCollector()
git_info = collector.collect_git_info()

print(f"Branch: {git_info.branch}")
print(f"Staged: {git_info.staged}")
print(f"Modified: {git_info.modified}")
print(f"Untracked: {git_info.untracked}")

assert git_info.branch is not None, "Branch not detected"
print("Git collection test passed")
EOF
test_result "Git information collection" $?
echo ""

# Test 6: Metrics tracker (duration calculation)
echo "${BLUE}Test 6: Session Metrics Tracking${NC}"
python3 << 'EOF'
from moai_adk.statusline.metrics_tracker import MetricsTracker

tracker = MetricsTracker()
duration = tracker.get_duration()

print(f"Duration: {duration}")
assert duration is not None and duration != "", "Duration not calculated"
assert any(unit in duration for unit in ['m', 'h', 's']), "Duration format incorrect"
print("Metrics tracking test passed")
EOF
test_result "Session metrics tracking" $?
echo ""

# Test 7: Version reader
echo "${BLUE}Test 7: Version Information Reading${NC}"
python3 << 'EOF'
from moai_adk.statusline.version_reader import VersionReader

reader = VersionReader()
version = reader.get_version()

print(f"Version: {version}")
assert version is not None and version != "", "Version not read"
print("Version reading test passed")
EOF
test_result "Version information reading" $?
echo ""

# Test 8: Alfred detector (no errors)
echo "${BLUE}Test 8: Alfred Task Detection${NC}"
python3 << 'EOF'
from moai_adk.statusline.alfred_detector import AlfredDetector

detector = AlfredDetector()
task = detector.detect_active_task()

print(f"Command: {task.command}")
print(f"Spec ID: {task.spec_id}")
print(f"Stage: {task.stage}")
print("Alfred detection test passed (no active task expected)")
EOF
test_result "Alfred task detection" $?
echo ""

# Test 9: Update checker (no errors on API failure)
echo "${BLUE}Test 9: Update Availability Check${NC}"
python3 << 'EOF'
from moai_adk.statusline.update_checker import UpdateChecker

checker = UpdateChecker()
update_info = checker.check_update("0.20.1")

print(f"Update available: {update_info.available}")
print(f"Latest version: {update_info.latest_version}")
print("Update check test passed (API failure handled gracefully)")
EOF
test_result "Update availability check" $?
echo ""

# Test 10: Main entry point
echo "${BLUE}Test 10: Main Entry Point${NC}"
echo '{}' | python3 -m moai_adk.statusline.main > /dev/null 2>&1
test_result "Main entry point execution" $?
echo ""

# Test 11: Full integration with session context
echo "${BLUE}Test 11: Full Integration Test${NC}"
python3 << 'EOF'
from moai_adk.statusline.main import build_statusline_data

context = {
    "model": {"display_name": "H 4.5"},
    "cwd": "/Users/goos/MoAI/MoAI-ADK",
    "statusline": {"mode": "compact"}
}

statusline = build_statusline_data(context, mode="compact")
print(f"Full statusline: {statusline}")
assert statusline is not None, "Statusline not generated"
assert len(statusline) <= 80, "Length exceeded for compact mode"
print("Full integration test passed")
EOF
test_result "Full integration test" $?
echo ""

# Test 12: Unit tests with pytest
echo "${BLUE}Test 12: Pytest Unit Tests${NC}"
python3 -m pytest tests/statusline/ -v --tb=short 2>&1 | tail -20
PYTEST_EXIT=${PIPESTATUS[0]}
test_result "All unit tests" $PYTEST_EXIT
echo ""

# Summary
echo "=========================================="
echo "  Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Restart Claude Code"
    echo "2. Look for statusline in the terminal status bar"
    echo "3. Check docs/guides/statusline/setup.md for configuration"
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    exit 1
fi
