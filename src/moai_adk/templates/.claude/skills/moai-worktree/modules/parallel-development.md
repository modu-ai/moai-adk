# Parallel Development Module

Purpose: Advanced patterns and workflows for parallel SPEC development using isolated worktrees, enabling true concurrent development without context switching overhead.

Version: 1.0.0
Last Updated: 2025-11-29

---

## Quick Reference (30 seconds)

Parallel Development Benefits:
- Context Isolation: Each SPEC has independent Git state, files, and environment
- Zero Switching Cost: Instant switching between worktrees without loading/unloading
- Concurrent Development: Multiple SPECs developed simultaneously by single developer
- Safe Experimentation: Isolated environments prevent conflicts and contamination
- Clean Integration: Automatic sync and conflict resolution maintain code integrity

Core Workflow:
```bash
# Create parallel environments
moai-worktree new SPEC-001 "User Authentication"
moai-worktree new SPEC-002 "Payment Gateway"
moai-worktree new SPEC-003 "Dashboard Analytics"

# Parallel development
cd $(moai-worktree go SPEC-001) && /moai:2-run SPEC-001 &
cd $(moai-worktree go SPEC-002) && /moai:2-run SPEC-002 &
cd $(moai-worktree go SPEC-003) && /moai:2-run SPEC-003 &

# Integration and sync
moai-worktree sync --all
moai-worktree clean --merged-only
```

---

## Parallel Development Architecture

### Worktree Isolation Model

Complete Isolation Layers:

```

 Developer Machine 

 Main Repository (project_root/) 
 .git/ 
 src/ 
 docs/ 
 .moai-worktree-registry.json 

 Worktree Root (~/workflows/project-name/) 
 
 SPEC-001/ ← Isolated Environment 1 
 .git/ ← Worktree Git metadata 
 src/ ← Complete project copy 
 .env.local ← Worktree-specific env vars 
 .vscode/ ← Worktree-specific IDE config 
 
 SPEC-002/ ← Isolated Environment 2 
 .git/ 
 src/ 
 node_modules/ ← Independent dependencies 
 build/ ← Independent build artifacts 
 
 SPEC-003/ ← Isolated Environment 3 
 .git/ 
 src/ 
 .venv/ ← Independent Python environment 
 test_results/ ← Independent test results 

```

Isolation Benefits:

1. Git State Isolation: Each worktree has independent branch state, commits, and history
2. File System Isolation: Complete project copy with independent modifications
3. Dependency Isolation: Separate node_modules, .venv, and build artifacts
4. Configuration Isolation: Worktree-specific .env, IDE settings, and tool configurations
5. Process Isolation: Independent development servers, test runners, and build processes

### Parallel Development Patterns

#### Pattern 1: Independent SPEC Development

Use Case: Multiple unrelated features developed simultaneously

```bash
# Phase 1: Setup parallel environments
echo "Setting up parallel development environments..."

# Create worktrees for different feature areas
moai-worktree new SPEC-AUTH-001 "User Authentication System"
moai-worktree new SPEC-PAY-001 "Payment Gateway Integration"
moai-worktree new SPEC-DASH-001 "Analytics Dashboard"

# Phase 2: Initialize each environment
echo "Initializing development environments..."

# Setup authentication worktree
cd $(moai-worktree go SPEC-AUTH-001)
npm install
npm run setup:auth
npm run dev &
AUTH_PID=$!

# Setup payment worktree
cd $(moai-worktree go SPEC-PAY-001)
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python manage.py migrate &
PAY_PID=$!

# Setup dashboard worktree
cd $(moai-worktree go SPEC-DASH-001)
yarn install
yarn build &
DASH_PID=$!

# Phase 3: Parallel development
echo "Starting parallel development..."

# Work on authentication in first terminal
moai-worktree switch SPEC-AUTH-001
/moai:2-run SPEC-AUTH-001

# Work on payment in second terminal (simultaneous)
moai-worktree switch SPEC-PAY-001
/moai:2-run SPEC-PAY-001

# Work on dashboard in third terminal (simultaneous)
moai-worktree switch SPEC-DASH-001
/moai:2-run SPEC-DASH-001

# Phase 4: Integration
echo "Integrating parallel development work..."

moai-worktree sync --all
/moai:3-sync --parallel SPEC-AUTH-001,SPEC-PAY-001,SPEC-DASH-001
```

#### Pattern 2: Sequential Feature Development

Use Case: Features with dependencies, developed in sequence with overlap

```bash
# Sequential development with preparation overlap
echo "Sequential development with preparation..."

# Step 1: Start foundation work (SPEC-FOUND-001)
moai-worktree new SPEC-FOUND-001 "Foundation Services"
cd $(moai-worktree go SPEC-FOUND-001)
/moai:2-run SPEC-FOUND-001

# Step 2: While foundation develops, prepare dependent feature
moai-worktree new SPEC-API-001 "API Layer" --base develop
cd $(moai-worktree go SPEC-API-001)
# Setup API structure, wait for foundation

# Step 3: Foundation complete → switch to API
moai-worktree switch SPEC-FOUND-001
# Complete foundation, merge to develop
git checkout develop
git merge feature/SPEC-FOUND-001-foundation-services
git push origin develop

# Step 4: API development with foundation available
moai-worktree switch SPEC-API-001
moai-worktree sync SPEC-API-001 # Get foundation changes
/moai:2-run SPEC-API-001

# Step 5: While API develops, prepare UI layer
moai-worktree new SPEC-UI-001 "User Interface" --base develop
cd $(moai-worktree go SPEC-UI-001)
# Setup UI structure, wait for API
```

#### Pattern 3: Experiment-Production Parallel

Use Case: Experimental features alongside stable production work

```bash
# Parallel experimental and production work
echo "Experimental and production parallel development..."

# Production worktree (stable)
moai-worktree new SPEC-PROD-001 "Production Bug Fixes" --base main
cd $(moai-worktree go SPEC-PROD-001)
# Stable production fixes

# Experimental worktree (innovative)
moai-worktree new SPEC-EXP-001 "Experimental Feature X" --base develop
cd $(moai-worktree go SPEC-EXP-001)
# Experimental development without affecting production

# Context switching for reviews
alias prod="cd \$(moai-worktree go SPEC-PROD-001)"
alias exp="cd \$(moai-worktree go SPEC-EXP-001)"

# Quick comparison
prod
git log --oneline -10
exp
git log --oneline -10
```

---

## Advanced Parallel Workflows

### Multi-Developer Coordination

Team Parallel Development:

```bash
# Developer A - Authentication worktree
moai-worktree new SPEC-AUTH-001 "User Authentication" --developer alice
cd $(moai-worktree go SPEC-AUTH-001)
/moai:2-run SPEC-AUTH-001

# Developer B - Payment worktree
moai-worktree new SPEC-PAY-001 "Payment Processing" --developer bob
cd $(moai-worktree go SPEC-PAY-001)
/moai:2-run SPEC-PAY-001

# Developer C - Integration worktree
moai-worktree new SPEC-INT-001 "Service Integration" --developer charlie
cd $(moai-worktree go SPEC-INT-001)
# Integrates AUTH and PAY services

# Team coordination commands
moai-worktree list --all-developers
moai-worktree status --team-overview
moai-worktree sync --all --parallel
```

Conflict Prevention Strategies:

```bash
# Pre-development coordination
echo "Setting up coordinated development..."

# Define integration points early
moai-worktree new SPEC-AUTH-001 "Authentication API" --base main
moai-worktree new SPEC-PAY-001 "Payment Service" --base main

# Create integration contract
cd $(moai-worktree go SPEC-AUTH-001)
mkdir -p contracts
echo '{"auth_endpoint": "/api/auth", "version": "1.0"}' > contracts/api.json

cd $(moai-worktree go SPEC-PAY-001)
mkdir -p contracts
echo '{"payment_endpoint": "/api/pay", "auth_required": true}' > contracts/api.json

# Shared integration worktree
moai-worktree new SPEC-INT-001 "Service Integration"
cd $(moai-worktree go SPEC-INT-001)
# Pulls contracts from AUTH and PAY worktrees
```

### Continuous Integration Parallel

CI/CD Pipeline Integration:

```yaml
# .github/workflows/parallel-development.yml
name: Parallel Development CI

on:
 push:
 branches: [ "feature/SPEC-*" ]

jobs:
 detect-worktree:
 runs-on: ubuntu-latest
 outputs:
 spec-id: ${{ steps.detect.outputs.spec-id }}
 worktree-type: ${{ steps.detect.outputs.worktree-type }}
 steps:
 - name: Detect SPEC and worktree type
 id: detect
 run: |
 SPEC_ID=$(echo "${{ github.ref }}" | sed 's/.*feature\/SPEC-\([0-9]*\).*/SPEC-\1/')
 WORKTREE_TYPE=$(echo "${{ github.ref }}" | sed 's/.*feature\/SPEC-[0-9]*-\(.*\)/\1/')
 echo "spec-id=$SPEC_ID" >> $GITHUB_OUTPUT
 echo "worktree-type=$WORKTREE_TYPE" >> $GITHUB_OUTPUT

 test-worktree:
 needs: detect-worktree
 runs-on: ubuntu-latest
 strategy:
 matrix:
 worktree-type: [auth, payment, dashboard]
 steps:
 - uses: actions/checkout@v3
 with:
 fetch-depth: 0

 - name: Setup worktree environment
 run: |
 # Simulate worktree isolation
 mkdir -p ./worktrees/${{ needs.detect-worktree.outputs.spec-id }}
 cd ./worktrees/${{ needs.detect-worktree.outputs.spec-id }}

 # Copy project structure
 cp -r ../../src ./
 cp -r ../../tests ./
 cp -r ../../requirements.txt ./

 # Setup environment based on worktree type
 if [ "${{ matrix.worktree-type }}" = "auth" ]; then
 pip install -r requirements-auth.txt
 elif [ "${{ matrix.worktree-type }}" = "payment" ]; then
 pip install -r requirements-payment.txt
 fi

 - name: Run worktree-specific tests
 run: |
 cd ./worktrees/${{ needs.detect-worktree.outputs.spec-id }}
 python -m pytest tests/${{ matrix.worktree-type }}/ -v

 sync-worktrees:
 needs: [detect-worktree, test-worktree]
 runs-on: ubuntu-latest
 if: success()
 steps:
 - name: Sync worktrees
 run: |
 echo "All worktree tests passed for ${{ needs.detect-worktree.outputs.spec-id }}"
 echo "Triggering worktree sync..."
 # This would trigger worktree sync in development environment
```

### Performance Optimization for Parallel Development

Resource Management:

```bash
# Resource-aware worktree management
echo "Optimizing for parallel development..."

# Lightweight worktrees for rapid iteration
moai-worktree new SPEC-PROTO-001 "Prototype" --shallow --depth 1
cd $(moai-worktree go SPEC-PROTO-001)
# Fast setup, minimal history

# Full worktrees for comprehensive development
moai-worktree new SPEC-FULL-001 "Complete Feature"
cd $(moai-worktree go SPEC-FULL-001)
# Full history, complete setup

# Background operations for parallel work
sync_all_background() {
 for spec in $(moai-worktree list --status active --format json | jq -r '.worktrees[].id'); do
 (moai-worktree sync "$spec" --background) &
 done
 wait
}

# Memory-efficient switching
quick_switch() {
 local target_spec=$1
 # Pre-load worktree metadata in background
 (moai-worktree status "$target_spec" --quiet > /dev/null) &

 # Switch immediately
 moai-worktree switch "$target_spec"

 # Background process completes metadata loading
 wait
}
```

Development Server Management:

```bash
# Multi-server development environment
echo "Managing parallel development servers..."

# Start servers for each worktree
start_parallel_servers() {
 local worktrees=($(moai-worktree list --status active --format json | jq -r '.worktrees[].id'))

 for spec in "${worktrees[@]}"; do
 (
 echo "Starting server for $spec..."
 cd $(moai-worktree go "$spec")

 case $spec in
 *AUTH*)
 npm run dev:auth --port 3001 &
 echo $! > .auth_server.pid
 ;;
 *PAY*)
 python manage.py runserver --port 8002 &
 echo $! > .payment_server.pid
 ;;
 *DASH*)
 yarn start --port 3003 &
 echo $! > .dashboard_server.pid
 ;;
 esac
 ) &
 done

 wait
 echo "All development servers started"
}

# Stop all servers
stop_parallel_servers() {
 local worktrees=($(moai-worktree list --status active --format json | jq -r '.worktrees[].id'))

 for spec in "${worktrees[@]}"; do
 (
 cd $(moai-worktree go "$spec")

 # Kill servers by PID files
 for pid_file in ./*_server.pid; do
 if [ -f "$pid_file" ]; then
 kill $(cat "$pid_file") 2>/dev/null || true
 rm "$pid_file"
 fi
 done

 # Kill by port as fallback
 for port in 3001 8002 3003; do
 lsof -ti:$port | xargs kill -9 2>/dev/null || true
 done
 ) &
 done

 wait
 echo "All development servers stopped"
}
```

---

## Integration Patterns

### IDE Integration

VS Code Multi-Root Setup:

```json
// .vscode/workspaces.json
{
 "folders": [
 {
 "name": "Main Repository",
 "path": "."
 },
 {
 "name": "SPEC-AUTH-001",
 "path": "~/workflows/project-name/SPEC-AUTH-001"
 },
 {
 "name": "SPEC-PAY-001",
 "path": "~/workflows/project-name/SPEC-PAY-001"
 },
 {
 "name": "SPEC-DASH-001",
 "path": "~/workflows/project-name/SPEC-DASH-001"
 }
 ],
 "settings": {
 "git.autoFetch": false,
 "git.autorefresh": true,
 "workbench.editor.enablePreview": false
 },
 "extensions": {
 "recommendations": [
 "ms-vscode.vscode-json",
 "ms-python.python",
 "bradlc.vscode-tailwindcss"
 ]
 }
}
```

Worktree-Specific Settings:

```json
// ~/workflows/project-name/SPEC-AUTH-001/.vscode/settings.json
{
 "python.defaultInterpreterPath": "./.venv/bin/python",
 "python.linting.enabled": true,
 "python.linting.pylintEnabled": true,
 "python.formatting.provider": "black",
 "files.exclude": {
 "/__pycache__": true,
 "/.pytest_cache": true
 }
}
```

### Shell Integration

Parallel Development Shell Functions:

```bash
# Add to ~/.bashrc or ~/.zshrc

# Enhanced worktree switching with context
mwswitch() {
 local spec_id="$1"
 local current_spec=$(basename $(pwd) 2>/dev/null | grep '^SPEC-' || echo "")

 if [ "$current_spec" = "$spec_id" ]; then
 echo "Already in worktree: $spec_id"
 return 0
 fi

 # Save current context
 if [ -n "$current_spec" ]; then
 echo "$current_spec" > ~/.last_worktree
 echo "Saved context: $current_spec → ~/.last_worktree"
 fi

 # Switch to new worktree
 moai-worktree switch "$spec_id"

 # Load worktree-specific environment
 if [ -f ".worktree-env" ]; then
 source .worktree-env
 echo "Loaded worktree environment"
 fi
}

# Quick toggle between last two worktrees
mwtoggle() {
 local last_spec=$(cat ~/.last_worktree 2>/dev/null)

 if [ -z "$last_spec" ]; then
 echo "No previous worktree saved"
 return 1
 fi

 local current_spec=$(basename $(pwd) 2>/dev/null | grep '^SPEC-' || echo "")
 mwswitch "$last_spec"
 echo "$current_spec" > ~/.last_worktree
}

# Parallel development status
mwstatus() {
 echo "=== Parallel Development Status ==="
 moai-worktree status --all

 echo ""
 echo "=== Active Development Servers ==="

 local worktrees=($(moai-worktree list --status active --format json | jq -r '.worktrees[].id'))

 for spec in "${worktrees[@]}"; do
 echo ""
 echo "$spec:"
 cd $(moai-worktree go "$spec") 2>/dev/null

 # Check for running servers
 if [ -f ".auth_server.pid" ] && kill -0 $(cat .auth_server.pid) 2>/dev/null; then
 echo " Auth server running (PID: $(cat .auth_server.pid))"
 fi

 if [ -f ".payment_server.pid" ] && kill -0 $(cat .payment_server.pid) 2>/dev/null; then
 echo " Payment server running (PID: $(cat .payment_server.pid))"
 fi

 # Check for uncommitted changes
 if ! git diff --quiet || ! git diff --cached --quiet; then
 echo " Has uncommitted changes"
 fi
 done
}
```

---

## Advanced Use Cases

### Feature Flag Development

Parallel Feature Flag Implementation:

```bash
# Feature flag worktree pattern
echo "Setting up feature flag development..."

# Create worktrees for different feature flag configurations
moai-worktree new SPEC-FF-001 "Feature Flag Backend"
moai-worktree new SPEC-FF-002 "Feature Flag Frontend"
moai-worktree new SPEC-FF-003 "Feature Flag Integration"

# Backend worktree - flag implementation
cd $(moai-worktree go SPEC-FF-001)
cat > src/flags.py << 'EOF'
FEATURE_NEW_AUTH = os.getenv('FEATURE_NEW_AUTH', 'false').lower() == 'true'
FEATURE_ENHANCED_PAYMENT = os.getenv('FEATURE_ENHANCED_PAYMENT', 'false').lower() == 'true'
EOF

# Frontend worktree - flag consumption
cd $(moai-worktree go SPEC-FF-002)
cat > src/featureFlags.ts << 'EOF'
export const flags = {
 NEW_AUTH: process.env.REACT_APP_NEW_AUTH === 'true',
 ENHANCED_PAYMENT: process.env.REACT_APP_ENHANCED_PAYMENT === 'true'
};
EOF

# Integration worktree - testing different flag combinations
cd $(moai-worktree go SPEC-FF-003)

# Test different flag configurations
test_flag_combination() {
 local auth_flag=$1
 local payment_flag=$2

 export FEATURE_NEW_AUTH=$auth_flag
 export REACT_APP_NEW_AUTH=$auth_flag
 export FEATURE_ENHANCED_PAYMENT=$payment_flag
 export REACT_APP_ENHANCED_PAYMENT=$payment_flag

 echo "Testing flags: NEW_AUTH=$auth_flag, ENHANCED_PAYMENT=$payment_flag"

 # Run integration tests
 npm run test:integration

 # Run E2E tests
 npm run test:e2e
}

# Test all combinations
for auth in true false; do
 for payment in true false; do
 test_flag_combination $auth $payment
 done
done
```

### Microservices Parallel Development

Microservice Architecture Pattern:

```bash
# Microservices parallel development
echo "Setting up microservices parallel development..."

# Service-specific worktrees
SERVICES=("auth-service" "payment-service" "notification-service" "gateway-service")

for service in "${SERVICES[@]}"; do
 spec_id="SPEC-MS-${service//service/}"
 description="Microservice: ${service^}"

 moai-worktree new "$spec_id" "$description" --template microservice

 cd $(moai-worktree go "$spec_id")

 # Service-specific setup
 echo "PORT=$((3000 + ${#SERVICES[@]}))" > .env
 echo "SERVICE_NAME=$service" >> .env
 echo "DATABASE_URL=postgresql://localhost:5432/${service}_dev" >> .env

 # Start service in background
 docker-compose up -d &
done

# Gateway worktree for integration
moai-worktree new SPEC-MS-GATEWAY "API Gateway Integration"
cd $(moai-worktree go SPEC-MS-GATEWAY)

# Configure service discovery
cat > config/services.yaml << 'EOF'
services:
 auth-service:
 url: http://localhost:3001
 health_check: /health
 payment-service:
 url: http://localhost:3002
 health_check: /health
 notification-service:
 url: http://localhost:3003
 health_check: /health
EOF

# Parallel testing
test_microservices() {
 echo "Testing microservices in parallel..."

 # Test each service independently
 for service in "${SERVICES[@]}"; do
 spec_id="SPEC-MS-${service//service/}"
 (
 cd $(moai-worktree go "$spec_id")
 npm run test &
 )
 done

 wait

 # Test integration through gateway
 cd $(moai-worktree go SPEC-MS-GATEWAY)
 npm run test:integration
}
```

---

## Performance and Optimization

### Memory and CPU Optimization

Resource-Aware Development:

```bash
# Resource monitoring for parallel development
monitor_parallel_resources() {
 echo "=== Parallel Development Resource Monitor ==="

 local worktrees=($(moai-worktree list --status active --format json | jq -r '.worktrees[].id'))

 for spec in "${worktrees[@]}"; do
 echo ""
 echo "$spec:"

 # Check for running processes
 local processes=$(pgrep -f "$(moai-worktree go "$spec")" | wc -l)
 echo " Processes: $processes"

 # Check disk usage
 local disk_usage=$(du -sh "$(moai-worktree go "$spec")" 2>/dev/null | cut -f1)
 echo " Disk Usage: $disk_usage"

 # Check for memory-intensive processes
 local memory_usage=$(ps aux | grep "$(moai-worktree go "$spec")" | awk '{sum+=$6} END {print sum/1024}' 2>/dev/null)
 echo " Memory Usage: ${memory_usage:-0}MB"
 done

 # System-wide resource usage
 echo ""
 echo "System Resources:"
 echo " CPU Load: $(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }')"
 echo " Memory Usage: $(free -h | awk 'NR==2{printf "%.1f%%", $3*100/$2}')"
 echo " Disk Usage: $(df -h . | awk 'NR==2{print $5}')"
}

# Optimize worktree for parallel development
optimize_worktree() {
 local spec_id="$1"

 echo "Optimizing worktree: $spec_id"

 cd $(moai-worktree go "$spec_id")

 # Clean unnecessary files
 find . -name "*.log" -delete 2>/dev/null || true
 find . -name ".pytest_cache" -type d -exec rm -rf {} + 2>/dev/null || true
 find . -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null || true

 # Optimize Git repository
 git gc --aggressive --prune=now

 # Compress large files if they exist
 find . -name "*.tar.gz" -exec gzip {} \; 2>/dev/null || true

 echo "Worktree optimization completed"
}
```

---

Version: 1.0.0
Last Updated: 2025-11-29
Module: Advanced parallel development patterns with isolation, coordination, and optimization strategies
