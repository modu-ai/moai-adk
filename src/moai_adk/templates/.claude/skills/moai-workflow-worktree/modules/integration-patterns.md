# Integration Patterns Module

Purpose: Comprehensive integration patterns for moai-workflow-worktree with MoAI-ADK workflow, development tools, and external systems.

Version: 1.0.0
Last Updated: 2025-11-29

---

## Quick Reference (30 seconds)

Integration Points:
- MoAI-ADK Workflow: Seamless integration with `/moai:1-plan`, `/moai:2-run`, `/moai:3-sync`
- Development Tools: IDEs, editors, terminal emulators, and development servers
- Git Workflows: Branch management, CI/CD pipelines, and code review processes
- Team Collaboration: Shared worktrees, code sharing, and coordination patterns

Core Integration Pattern:
```bash
# Plan Phase - Automatic worktree creation
/moai:1-plan "User Authentication" → auto-creates SPEC-001 worktree

# Development Phase - Isolated development
cd $(moai-workflow-worktree go SPEC-001)
/moai:2-run SPEC-001

# Sync Phase - Clean integration
moai-workflow-worktree sync SPEC-001
/moai:3-sync SPEC-001
```

---

## MoAI-ADK Workflow Integration

### Plan Phase Integration (`/moai:1-plan`)

Automatic Worktree Creation:

```python
# Integration in /moai:1-plan command
def create_worktree_after_spec(spec_id: str, spec_title: str) -> None:
 """Create worktree automatically after SPEC creation."""

 try:
 # Create worktree with automatic branching
 branch_name = f"feature/SPEC-{spec_id}-{spec_title.lower().replace(' ', '-')}"

 result = subprocess.run([
 "moai-workflow-worktree", "new", spec_id, spec_title,
 "--branch", branch_name,
 "--template", "spec-development"
 ], capture_output=True, text=True)

 if result.returncode == 0:
 print(f" Created worktree: {spec_id}")
 print(f" Branch: {branch_name}")
 print(f" Path: {extract_worktree_path(result.stdout)}")

 print("\nNext steps:")
 print(f"1. Switch to worktree: moai-workflow-worktree switch {spec_id}")
 print(f"2. Or use shell eval: eval $(moai-workflow-worktree go {spec_id})")
 print(f"3. Start development: /moai:2-run {spec_id}")
 else:
 print(f" Worktree creation failed: {result.stderr}")

 except Exception as e:
 print(f" Error creating worktree: {e}")

# Enhanced /moai:1-plan output with worktree information
def display_worktree_integration_info(spec_id: str, spec_title: str) -> None:
 """Display comprehensive worktree integration guidance."""

 print(f"""
 Worktree Integration for SPEC-{spec_id}

 Worktree Information:
 SPEC ID: {spec_id}
 Title: {spec_title}
 Branch: feature/SPEC-{spec_id}-{spec_title.lower().replace(' ', '-')}
 Path: ~/workflows/{os.path.basename(os.getcwd())}/SPEC-{spec_id}

 Quick Start Options:
 1. Switch to worktree: moai-workflow-worktree switch SPEC-{spec_id}
 2. Shell integration: eval $(moai-workflow-worktree go SPEC-{spec_id})
 3. Start development: cd $(moai-workflow-worktree go SPEC-{spec_id}) && /moai:2-run SPEC-{spec_id}

 Development Workflow:
 Phase 1: SPEC created
 Phase 2: → Worktree ready for development
 Phase 3: → Implement with /moai:2-run
 Phase 4: → Sync with /moai:3-sync
 Phase 5: → Clean up with moai-workflow-worktree clean

 Advanced Options:
 • Create with custom template: moai-workflow-worktree new {spec_id} "{spec_title}" --template <template>
 • Create from develop branch: moai-workflow-worktree new {spec_id} "{spec_title}" --base develop
 • Create with shallow clone: moai-workflow-worktree new {spec_id} "{spec_title}" --shallow
""")
```

Template-Based SPEC Development:

```python
# SPEC development template configuration
SPEC_DEVELOPMENT_TEMPLATE = {
 "setup_commands": [
 "echo 'Setting up SPEC development environment...'",
 "npm run setup:spec 2>/dev/null || echo 'No npm setup required'",
 "python -m pip install -r requirements-dev.txt 2>/dev/null || echo 'No Python requirements'",
 "echo 'Environment ready for SPEC development'"
 ],
 "files": {
 ".spec-config": """
# SPEC Development Configuration
SPEC_ID={spec_id}
SPEC_TITLE={spec_title}
DEV_MODE=spec
LOG_LEVEL=debug
""",
 ".vscode/tasks.json": """{
 "version": "2.0.0",
 "tasks": [
 {{
 "label": "Run SPEC Tests",
 "type": "shell",
 "command": "/moai:2-run {spec_id}",
 "group": "test",
 "presentation": {{
 "echo": true,
 "reveal": "always",
 "focus": false,
 "panel": "shared"
 }}
 }}
 ]
}}"""
 },
 "env_vars": {
 "SPEC_MODE": "true",
 "DEVELOPMENT_SPEC": spec_id,
 "SPEC_BRANCH": f"feature/SPEC-{spec_id}"
 }
}
```

### Development Phase Integration (`/moai:2-run`)

Worktree-Aware TDD Implementation:

```python
# Enhanced manager-tdd for worktree environments
class WorktreeAwareTDDManager:
 def __init__(self, spec_id: str):
 self.spec_id = spec_id
 self.worktree_path = self._detect_worktree_path()
 self.is_worktree_env = self._is_in_worktree()

 def _detect_worktree_path(self) -> Optional[Path]:
 """Detect if running in worktree environment."""
 current_path = Path.cwd()

 # Check if current directory is a worktree
 if current_path.name.startswith("SPEC-"):
 return current_path

 # Check if parent is worktrees directory
 if "worktrees" in current_path.parts:
 for i, part in enumerate(current_path.parts):
 if part == "worktrees" and i + 1 < len(current_path.parts):
 spec_part = current_path.parts[i + 1]
 if spec_part.startswith("SPEC-"):
 return current_path / spec_part

 return None

 def execute_tdd_in_worktree(self) -> TDDResult:
 """Execute TDD cycle in worktree context."""

 if not self.is_worktree_env:
 print(f" Not in worktree environment. Consider running in worktree:")
 print(f" moai-workflow-worktree switch {self.spec_id}")
 print(f" or: eval $(moai-workflow-worktree go {self.spec_id})")

 # TDD execution with worktree awareness
 result = self._run_tdd_cycle()

 # Update worktree metadata
 if self.worktree_path:
 self._update_worktree_metadata(result)

 return result

 def _update_worktree_metadata(self, tdd_result: TDDResult) -> None:
 """Update worktree metadata after TDD execution."""

 try:
 # Update registry with TDD results
 subprocess.run([
 "moai-workflow-worktree", "config", "set",
 f"last_tdd_result.{self.spec_id}",
 json.dumps(tdd_result.to_dict())
 ], check=False)

 # Update last access time
 subprocess.run([
 "moai-workflow-worktree", "status", self.spec_id, "--update-access"
 ], check=False)

 except Exception as e:
 print(f" Failed to update worktree metadata: {e}")
```

### Sync Phase Integration (`/moai:3-sync`)

Automated Worktree Synchronization:

```python
# Enhanced /moai:3-sync with worktree integration
class WorktreeSyncManager:
 def sync_spec_with_worktree(self, spec_id: str) -> SyncResult:
 """Synchronize SPEC development with main repository."""

 # Check if worktree exists
 worktree_info = self._get_worktree_info(spec_id)
 if not worktree_info:
 print(f" No worktree found for {spec_id}")
 print(f" Create with: moai-workflow-worktree new {spec_id}")
 return SyncResult(skipped=True, reason="No worktree")

 # Sync worktree with base branch
 print(f" Syncing worktree {spec_id}...")

 try:
 sync_result = subprocess.run([
 "moai-workflow-worktree", "sync", spec_id,
 "--auto-resolve",
 "--include", "src/", "docs/", "tests/"
 ], capture_output=True, text=True)

 if sync_result.returncode == 0:
 print(f" Worktree {spec_id} synchronized successfully")

 # Continue with regular sync process
 return self._perform_regular_sync(spec_id, worktree_info)
 else:
 print(f" Worktree sync failed: {sync_result.stderr}")
 return SyncResult(success=False, error=sync_result.stderr)

 except Exception as e:
 print(f" Error syncing worktree: {e}")
 return SyncResult(success=False, error=str(e))

 def _perform_regular_sync(self, spec_id: str, worktree_info: dict) -> SyncResult:
 """Perform regular documentation sync after worktree sync."""

 # Collect changes from worktree
 worktree_path = worktree_info['path']

 # Extract documentation updates
 docs_updates = self._extract_documentation_updates(worktree_path)

 # Generate updated documentation
 updated_docs = self._generate_updated_documentation(spec_id, docs_updates)

 # Create pull request if needed
 if worktree_info.get('git_info', {}).get('commits_ahead', 0) > 0:
 self._create_pull_request_for_spec(spec_id, worktree_info)

 return SyncResult(
 success=True,
 worktree_synced=True,
 documentation_updated=updated_docs
 )
```

### Automated Cleanup Integration

Post-PR Cleanup Workflow:

```python
# Automatic cleanup after successful integration
def cleanup_completed_spec(spec_id: str, merge_status: str) -> None:
 """Clean up worktree after successful SPEC integration."""

 if merge_status == "merged":
 print(f" SPEC {spec_id} successfully merged!")

 # Offer cleanup options
 cleanup_choice = input("""
Clean up worktree options:
1. Remove worktree (recommended)
2. Keep worktree for reference
3. Archive worktree
4. Skip cleanup

Choice [1-4]: """)

 if cleanup_choice == "1":
 # Remove worktree
 result = subprocess.run([
 "moai-workflow-worktree", "remove", spec_id, "--force"
 ], capture_output=True, text=True)

 if result.returncode == 0:
 print(f" Worktree {spec_id} removed successfully")
 else:
 print(f" Error removing worktree: {result.stderr}")

 elif cleanup_choice == "3":
 # Archive worktree
 archive_path = f"~/workflows/archives/{spec_id}-{datetime.now().strftime('%Y%m%d')}"
 subprocess.run([
 "moai-workflow-worktree", "remove", spec_id, "--backup"
 ], capture_output=True, text=True)

 print(f" Worktree {spec_id} archived to {archive_path}")

 # Update registry
 subprocess.run([
 "moai-workflow-worktree", "config", "set",
 f"completed_specs.{spec_id}", json.dumps({
 "merged_at": datetime.utcnow().isoformat(),
 "cleanup_action": cleanup_choice
 })
 ], check=False)
```

---

## Development Tools Integration

### IDE Integration

VS Code Multi-Root Workspace:

```json
// .vscode/workspaces.json - Auto-generated
{
 "version": "0.1.0",
 "folders": [
 {
 "name": "Main Repository",
 "path": "."
 }
 ],
 "extensions": {
 "recommendations": [
 "ms-vscode.vscode-typescript-next",
 "ms-python.python",
 "bradlc.vscode-tailwindcss"
 ]
 },
 "launch": {
 "version": "0.2.0",
 "configurations": []
 }
}
```

Dynamic Workspace Generator:

```python
# generate_vscode_workspace.py
def generate_workspace_with_worktrees():
 """Generate VS Code workspace including all active worktrees."""

 worktrees = get_active_worktrees()
 workspace_config = {
 "version": "0.1.0",
 "folders": [
 {
 "name": "Main Repository",
 "path": "."
 }
 ],
 "extensions": {
 "recommendations": get_workspace_extensions()
 },
 "launch": {
 "version": "0.2.0",
 "configurations": []
 },
 "tasks": {
 "version": "2.0.0",
 "tasks": []
 }
 }

 # Add worktree folders
 for worktree in worktrees:
 workspace_config["folders"].append({
 "name": f"SPEC-{worktree['id']}",
 "path": worktree['path']
 })

 # Add worktree-specific tasks
 workspace_config["tasks"]["tasks"].extend(
 generate_worktree_tasks(worktree)
 )

 # Write workspace file
 with open(".vscode/workspaces.json", "w") as f:
 json.dump(workspace_config, f, indent=2)

 print(" VS Code workspace updated with worktrees")

def generate_worktree_tasks(worktree: dict) -> List[dict]:
 """Generate VS Code tasks for specific worktree."""

 spec_id = worktree['id']
 worktree_path = worktree['path']

 return [
 {
 "label": f"Run SPEC-{spec_id} Tests",
 "type": "shell",
 "command": f"cd {worktree_path} && /moai:2-run {spec_id}",
 "group": "test",
 "presentation": {
 "echo": true,
 "reveal": "always",
 "focus": false,
 "panel": "shared"
 }
 },
 {
 "label": f"Sync SPEC-{spec_id}",
 "type": "shell",
 "command": f"moai-workflow-worktree sync {spec_id}",
 "group": "build"
 },
 {
 "label": f"Switch to SPEC-{spec_id}",
 "type": "shell",
 "command": f"moai-workflow-worktree switch {spec_id}",
 "group": "navigation"
 }
 ]
```

### Terminal Integration

Enhanced Shell Profile:

```bash
# ~/.bashrc or ~/.zshrc - Worktree integration

# Worktree completion
_moai_worktree_completion() {
 local worktrees=($(moai-workflow-worktree list --format json 2>/dev/null | jq -r '.worktrees[].id' 2>/dev/null))
 COMPREPLY=($(compgen -W "${worktrees[*]}" "${COMP_WORDS[COMP_CWORD]}"))
}

complete -F _moai_worktree_completion moai-workflow-worktree

# Worktree-aware prompt
update_prompt_with_worktree() {
 if [ -f "../.moai-workflow-worktree-registry.json" ]; then
 local spec_id=$(basename $(pwd) 2>/dev/null | grep '^SPEC-' || echo "")
 if [ -n "$spec_id" ]; then
 export WORKTREE_SPEC=$spec_id
 export PS1="\[\033[36m\]$spec_id\[\033[0m\]:$PS1"
 fi
 else
 unset WORKTREE_SPEC
 fi
}

PROMPT_COMMAND=update_prompt_with_worktree

# Worktree navigation aliases
alias mw='moai-workflow-worktree'
alias mwl='moai-workflow-worktree list'
alias mws='moai-workflow-worktree switch'
alias mwg='eval $(moai-workflow-worktree go'
alias mwsync='moai-workflow-worktree sync'
alias mwclean='moai-workflow-worktree clean'

# Quick worktree functions
mwnew() {
 local spec_id="$1"
 local description="$2"
 moai-workflow-worktree new "$spec_id" "$description"
 moai-workflow-worktree switch "$spec_id"
}

mwdev() {
 local spec_id="$1"
 moai-workflow-worktree switch "$spec_id"
 /moai:2-run "$spec_id"
}

mwpush() {
 local spec_id="${1:-$WORKTREE_SPEC}"
 if [ -n "$spec_id" ]; then
 moai-workflow-worktree sync "$spec_id"
 cd $(moai-workflow-worktree go "$spec_id")
 git push origin "feature/SPEC-$spec_id"
 else
 echo "No SPEC ID provided or detected"
 fi
}
```

### Git Integration

Enhanced Git Hooks:

```bash
# .git/hooks/post-checkout
#!/bin/bash
echo "Post-checkout hook: Checking worktree status"

# Check if we're in a worktree
if [ -f "../.moai-workflow-worktree-registry.json" ]; then
 SPEC_ID=$(basename $(pwd))
 echo "Switched to worktree: $SPEC_ID"

 # Update last access time
 moai-workflow-worktree status "$SPEC_ID" --update-access 2>/dev/null || true

 # Check if sync is needed
 SYNC_STATUS=$(moai-workflow-worktree status "$SPEC_ID" --sync-check 2>/dev/null || echo "unknown")
 if echo "$SYNC_STATUS" | grep -q "needs sync"; then
 echo " Worktree needs synchronization"
 echo " Run: moai-workflow-worktree sync $SPEC_ID"
 fi

 # Load worktree-specific environment
 if [ -f ".worktree-env" ]; then
 echo "Loading worktree environment..."
 source .worktree-env
 fi
fi

# .git/hooks/pre-push
#!/bin/bash
echo "Pre-push hook: Validating worktree state"

# Check if we're pushing from a worktree
if [ -f "../.moai-workflow-worktree-registry.json" ]; then
 SPEC_ID=$(basename $(pwd))
 echo "Pushing from worktree: $SPEC_ID"

 # Check for uncommitted changes
 if ! git diff --quiet || ! git diff --cached --quiet; then
 echo " Uncommitted changes detected"
 echo " Commit changes before pushing or use --force"
 exit 1
 fi

 # Check if worktree is synced with base
 SYNC_STATUS=$(moai-workflow-worktree status "$SPEC_ID" --sync-check 2>/dev/null || echo "unknown")
 if echo "$SYNC_STATUS" | grep -q "behind"; then
 echo " Worktree is behind base branch"
 echo " Run: moai-workflow-worktree sync $SPEC_ID"
 echo " Continue anyway? (y/N)"
 read -r response
 if [[ ! "$response" =~ ^[Yy]$ ]]; then
 exit 1
 fi
 fi

 # Update worktree metadata
 moai-workflow-worktree config set "last_push.$SPEC_ID" "$(date -Iseconds)" 2>/dev/null || true
fi
```

---

## Team Collaboration Integration

### Shared Worktree Patterns

Team Worktree Registry:

```python
# Shared registry configuration
class TeamWorktreeRegistry:
 def __init__(self, team_name: str, shared_registry_path: str):
 self.team_name = team_name
 self.shared_registry_path = Path(shared_registry_path)
 self.local_registry_path = Path.home() / "worktrees" / team_name / ".moai-workflow-worktree-registry.json"

 def sync_with_team_registry(self) -> None:
 """Synchronize local registry with team shared registry."""

 try:
 # Fetch latest team registry
 if self.shared_registry_path.exists():
 team_registry = self.load_registry(self.shared_registry_path)
 local_registry = self.load_registry(self.local_registry_path)

 # Merge registries
 merged_registry = self.merge_registries(team_registry, local_registry)

 # Save merged registry
 self.save_registry(merged_registry, self.local_registry_path)

 print(f" Synced with team registry: {self.team_name}")

 except Exception as e:
 print(f" Failed to sync with team registry: {e}")

 def share_worktree_with_team(self, spec_id: str, share_config: dict) -> None:
 """Share worktree configuration with team."""

 # Update team registry
 team_registry = self.load_registry(self.shared_registry_path)

 worktree_info = team_registry['worktrees'].get(spec_id, {})
 worktree_info.update({
 'shared_with': share_config.get('team_members', []),
 'share_level': share_config.get('level', 'read-only'),
 'shared_at': datetime.utcnow().isoformat(),
 'shared_by': os.getenv('USER', 'unknown')
 })

 team_registry['worktrees'][spec_id] = worktree_info
 self.save_registry(team_registry, self.shared_registry_path)

 print(f" Worktree {spec_id} shared with team")
```

Collaborative Development Workflow:

```bash
# Team collaboration script
collaborative_development() {
 local spec_id="$1"
 local team_members="$2" # Comma-separated list

 echo "Setting up collaborative development for $spec_id..."

 # Create shared worktree
 moai-workflow-worktree new "$spec_id" "Team Collaborative Development" --template collaborative

 # Configure team access
 IFS=',' read -ra MEMBERS <<< "$team_members"
 for member in "${MEMBERS[@]}"; do
 moai-workflow-worktree config set "team_access.$spec_id.$member" "read-write"
 echo " Granted read-write access to $member"
 done

 # Create shared configuration
 cd $(moai-workflow-worktree go "$spec_id")

 cat > .team-config << EOF
team_members: [$(echo "$team_members" | sed 's/,/, /g')]
collaboration_mode: active
shared_branch: feature/SPEC-$spec_id-collaborative
review_required: true
EOF

 echo " Team collaboration setup completed"
 echo "Team members can now join with:"
 echo " moai-workflow-worktree join $spec_id --team-member <name>"
}
```

---

## External System Integration

### CI/CD Pipeline Integration

GitHub Actions Worktree Integration:

```yaml
# .github/workflows/worktree-integration.yml
name: Worktree Integration CI

on:
 push:
 branches: [ "feature/SPEC-*" ]
 pull_request:
 branches: [ "main", "develop" ]

jobs:
 detect-spec:
 runs-on: ubuntu-latest
 outputs:
 spec-id: ${{ steps.spec.outputs.id }}
 spec-title: ${{ steps.spec.outputs.title }}
 is-worktree-branch: ${{ steps.spec.outputs.is-worktree }}
 steps:
 - uses: actions/checkout@v3

 - name: Detect SPEC information
 id: spec
 run: |
 if [[ "${{ github.ref }}" == refs/heads/feature/SPEC-* ]]; then
 SPEC_ID=$(echo "${{ github.ref }}" | sed 's/.*feature\/SPEC-\([0-9]*\).*/SPEC-\1/')
 SPEC_TITLE=$(echo "${{ github.ref }}" | sed 's/.*feature\/SPEC-[0-9]*-\(.*\)/\1/')
 echo "id=$SPEC_ID" >> $GITHUB_OUTPUT
 echo "title=$SPEC_TITLE" >> $GITHUB_OUTPUT
 echo "is-worktree=true" >> $GITHUB_OUTPUT
 else
 echo "id=none" >> $GITHUB_OUTPUT
 echo "title=none" >> $GITHUB_OUTPUT
 echo "is-worktree=false" >> $GITHUB_OUTPUT
 fi

 worktree-tests:
 needs: detect-spec
 if: needs.detect-spec.outputs.is-worktree == 'true'
 runs-on: ubuntu-latest
 strategy:
 matrix:
 test-type: [unit, integration, e2e]
 steps:
 - uses: actions/checkout@v3
 with:
 fetch-depth: 0

 - name: Setup worktree environment
 run: |
 SPEC_ID="${{ needs.detect-spec.outputs.spec-id }}"
 echo "Setting up worktree for $SPEC_ID"

 # Create worktree directory structure
 mkdir -p ./worktrees/$SPEC_ID

 # Checkout worktree branch
 git checkout "feature/SPEC-${SPEC_ID#SPEC-}-*" 2>/dev/null || git checkout main

 # Copy source files to simulate worktree
 mkdir -p ./worktrees/$SPEC_ID/src
 cp -r src/* ./worktrees/$SPEC_ID/src/

 # Setup test environment
 cd ./worktrees/$SPEC_ID

 if [ -f "../requirements.txt" ]; then
 pip install -r ../requirements.txt
 fi

 if [ -f "../package.json" ]; then
 npm install
 fi

 - name: Run ${{ matrix.test-type }} tests
 run: |
 cd ./worktrees/${{ needs.detect-spec.outputs.spec-id }}

 case "${{ matrix.test-type }}" in
 unit)
 python -m pytest tests/unit/ -v || npm test -- --coverage
 ;;
 integration)
 python -m pytest tests/integration/ -v || npm run test:integration
 ;;
 e2e)
 npm run test:e2e || python -m pytest tests/e2e/ -v
 ;;
 esac

 - name: Upload test results
 uses: actions/upload-artifact@v3
 if: always()
 with:
 name: test-results-${{ matrix.test-type }}-${{ needs.detect-spec.outputs.spec-id }}
 path: |
 ./worktrees/${{ needs.detect-spec.outputs.spec-id }}/test-results/
 ./worktrees/${{ needs.detect-spec.outputs.spec-id }}/coverage/
```

### Monitoring and Analytics Integration

Worktree Usage Analytics:

```python
# Worktree analytics collector
class WorktreeAnalytics:
 def __init__(self, registry_path: Path):
 self.registry_path = registry_path

 def collect_usage_metrics(self) -> dict:
 """Collect comprehensive usage metrics."""

 registry = self.load_registry()
 worktrees = registry.get('worktrees', {})

 metrics = {
 'timestamp': datetime.utcnow().isoformat(),
 'total_worktrees': len(worktrees),
 'active_worktrees': len([w for w in worktrees.values() if w.get('status') == 'active']),
 'disk_usage': self.calculate_total_disk_usage(worktrees),
 'sync_frequency': self.calculate_sync_frequency(worktrees),
 'developer_activity': self.analyze_developer_activity(worktrees),
 'performance_metrics': self.collect_performance_metrics(worktrees)
 }

 return metrics

 def export_analytics(self, output_format: str = 'json') -> str:
 """Export analytics in specified format."""

 metrics = self.collect_usage_metrics()

 if output_format == 'json':
 return json.dumps(metrics, indent=2)
 elif output_format == 'csv':
 return self.convert_to_csv(metrics)
 elif output_format == 'prometheus':
 return self.convert_to_prometheus(metrics)
 else:
 raise ValueError(f"Unsupported format: {output_format}")

 def send_to_monitoring_system(self, metrics: dict) -> None:
 """Send metrics to external monitoring system."""

 # Example: Send to Prometheus Pushgateway
 import requests

 pushgateway_url = os.getenv('PROMETHEUS_PUSHGATEWAY_URL')
 if pushgateway_url:
 try:
 response = requests.post(
 f"{pushgateway_url}/metrics/job/worktree-analytics",
 data=self.convert_to_prometheus(metrics),
 headers={'Content-Type': 'text/plain'}
 )

 if response.status_code == 200:
 print(" Metrics sent to monitoring system")
 else:
 print(f" Failed to send metrics: {response.status_code}")

 except Exception as e:
 print(f" Error sending metrics: {e}")
```

---

## Performance Optimization

### Resource Management Integration

Intelligent Resource Allocation:

```python
# Resource manager for parallel worktree operations
class WorktreeResourceManager:
 def __init__(self):
 self.cpu_count = os.cpu_count()
 self.memory_gb = self.get_available_memory_gb()
 self.max_concurrent_worktrees = min(self.cpu_count, 4)

 def optimize_parallel_operations(self, worktrees: List[str]) -> List[str]:
 """Optimize worktree operations based on available resources."""

 # Sort worktrees by priority and size
 prioritized_worktrees = self.prioritize_worktrees(worktrees)

 # Group worktrees for parallel execution
 operation_groups = self.group_worktrees_for_parallel_execution(
 prioritized_worktrees
 )

 return operation_groups

 def execute_with_resource_limits(self, operations: List[Callable]) -> List[Any]:
 """Execute operations with resource limits."""

 import concurrent.futures

 results = []

 with concurrent.futures.ThreadPoolExecutor(
 max_workers=self.max_concurrent_worktrees
 ) as executor:
 # Submit all operations
 future_to_operation = {
 executor.submit(op): op for op in operations
 }

 # Collect results as they complete
 for future in concurrent.futures.as_completed(future_to_operation):
 operation = future_to_operation[future]
 try:
 result = future.result()
 results.append(result)
 except Exception as e:
 print(f" Operation failed: {e}")
 results.append(None)

 return results
```

---

## Error Handling and Recovery

### Integration Error Recovery

Comprehensive Error Handling:

```python
# Integration error recovery system
class IntegrationErrorHandler:
 def __init__(self):
 self.error_log = []
 self.recovery_strategies = {
 'worktree_not_found': self.recover_missing_worktree,
 'sync_conflict': self.recover_sync_conflict,
 'permission_denied': self.recover_permission_error,
 'disk_space': self.recover_disk_space_error,
 'network_error': self.recover_network_error
 }

 def handle_integration_error(self, error: Exception, context: dict) -> bool:
 """Handle integration error with appropriate recovery strategy."""

 error_type = self.classify_error(error)

 if error_type in self.recovery_strategies:
 try:
 recovery_result = self.recovery_strategies[error_type](error, context)

 if recovery_result.success:
 print(f" Recovered from {error_type}: {recovery_result.message}")
 return True
 else:
 print(f" Recovery failed for {error_type}: {recovery_result.message}")
 return False

 except Exception as recovery_error:
 print(f" Recovery failed with exception: {recovery_error}")
 return False
 else:
 print(f" No recovery strategy for error type: {error_type}")
 return False

 def recover_missing_worktree(self, error: Exception, context: dict) -> RecoveryResult:
 """Recover from missing worktree error."""

 spec_id = context.get('spec_id')
 if not spec_id:
 return RecoveryResult(success=False, message="No SPEC ID in context")

 try:
 # Offer to recreate worktree
 print(f"Worktree {spec_id} not found. Recreating...")

 # Attempt recreation with backup if available
 backup_path = self.find_worktree_backup(spec_id)

 if backup_path:
 print(f"Found backup: {backup_path}")
 restore_result = self.restore_worktree_from_backup(spec_id, backup_path)

 if restore_result:
 return RecoveryResult(
 success=True,
 message=f"Worktree {spec_id} restored from backup"
 )

 # Create new worktree
 subprocess.run([
 "moai-workflow-worktree", "new", spec_id,
 "Auto-recreated after error"
 ], check=True)

 return RecoveryResult(
 success=True,
 message=f"Worktree {spec_id} recreated successfully"
 )

 except Exception as e:
 return RecoveryResult(
 success=False,
 message=f"Failed to recreate worktree: {e}"
 )
```

---

Version: 1.0.0
Last Updated: 2025-11-29
Module: Comprehensive integration patterns for moai-workflow-worktree with MoAI-ADK workflow, development tools, and external systems
