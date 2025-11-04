# Examples

## Correct Document Placement

### Internal Documentation

```python
# When creating implementation guide during /alfred:2-run
def create_implementation_guide(spec_id):
    location = f".moai/docs/implementation-{spec_id}.md"
    content = generate_implementation_content(spec_id)
    
    # Correct placement in .moai/docs/
    with open(location, 'w') as f:
        f.write(content)
    
    return f"Implementation guide created at {location}"
```

### SPEC Documents

```python
# During /alfred:1-plan when creating SPEC
def create_spec_structure(spec_id):
    spec_dir = f".moai/specs/{spec_id}/"
    
    # Create directory structure
    os.makedirs(spec_dir, exist_ok=True)
    
    # Create required SPEC files
    files = {
        "spec.md": generate_spec_content(),
        "plan.md": generate_plan_content(), 
        "acceptance.md": generate_acceptance_criteria()
    }
    
    for filename, content in files.items():
        with open(f"{spec_dir}{filename}", 'w') as f:
            f.write(content)
    
    return f"SPEC structure created at {spec_dir}"
```

### Sync Reports

```python
# During /alfred:3-sync when generating reports
def create_sync_report():
    timestamp = datetime.now().strftime("%Y%m%d")
    location = f".moai/reports/sync-report-{timestamp}.md"
    
    # Generate sync analysis
    analysis = analyze_sync_status()
    
    # Correct placement in .moai/reports/
    with open(location, 'w') as f:
        f.write(f"# Sync Report - {timestamp}\n\n")
        f.write(analysis)
    
    return f"Sync report created at {location}"
```

## ❌ Common Mistakes to Avoid

### Root Directory Violations

```python
# WRONG - Creating documentation in project root
def create_implementation_report_wrong():
    # ❌ This violates the rules
    with open("IMPLEMENTATION_REPORT.md", 'w') as f:
        f.write("Implementation details...")
    
    # Should be in .moai/docs/ instead
```

```python
# CORRECT - Proper placement
def create_implementation_report_correct():
    # ✅ Follows the rules
    location = ".moai/docs/implementation-SPEC-001.md"
    with open(location, 'w') as f:
        f.write("Implementation details...")
```

### Wrong Document Types in Wrong Locations

```python
# WRONG - Sync report in docs/
def create_sync_report_wrong():
    # ❌ Sync reports should go in .moai/reports/
    with open(".moai/docs/sync-report.md", 'w') as f:
        f.write("Sync analysis...")
```

```python
# CORRECT - Sync report in proper location
def create_sync_report_correct():
    # ✅ Follows the rules
    location = ".moai/reports/sync-report-20251105.md"
    with open(location, 'w') as f:
        f.write("Sync analysis...")
```

## User Interaction Examples

### Asking for Permission

```python
# When user hasn't explicitly requested documentation
def check_documentation_request(doc_type, location):
    if not user_explicitly_requested():
        return AskUserQuestion(
            question=f"Would you like me to create a {doc_type} at {location}?",
            options=[
                "YES - Create the document",
                "NO - Don't create documentation"
            ]
        )
    return True
```

### Auto-Creation Scenarios

```python
# During SPEC creation - auto-create without asking
def plan_command_handler():
    spec_id = generate_spec_id()
    
    # Auto-create SPEC structure (no user approval needed)
    create_spec_structure(spec_id)
    
    # But ask before creating extra documentation
    if should_create_implementation_guide():
        return check_documentation_request(
            "implementation guide", 
            f".moai/docs/implementation-{spec_id}.md"
        )
```

## Directory Creation Examples

### Setting Up MoAI Structure

```bash
# Create complete .moai directory structure
mkdir -p .moai/{docs,specs,reports,analysis,memory}

# Verify structure exists
ls -la .moai/
# Expected: docs/ specs/ reports/ analysis/ memory/ config.json
```

### SPEC Directory Structure

```python
def ensure_spec_structure(spec_id):
    """Ensure SPEC directory has required structure"""
    spec_path = f".moai/specs/{spec_id}/"
    
    # Create directory if doesn't exist
    os.makedirs(spec_path, exist_ok=True)
    
    # Check for required files
    required_files = ["spec.md", "plan.md", "acceptance.md"]
    missing_files = []
    
    for file in required_files:
        if not os.path.exists(f"{spec_path}{file}"):
            missing_files.append(file)
    
    return missing_files
```

## File Naming Examples

### Implementation Guides

```python
def generate_implementation_filename(spec_id):
    """Generate proper implementation guide filename"""
    return f".moai/docs/implementation-{spec_id}.md"

# Examples:
# .moai/docs/implementation-SPEC-001.md
# .moai/docs/implementation-user-auth.md
```

### Exploration Reports

```python
def generate_exploration_filename(topic):
    """Generate proper exploration report filename"""
    sanitized_topic = topic.lower().replace(" ", "-")
    return f".moai/docs/exploration-{sanitized_topic}.md"

# Examples:
# .moai/docs/exploration-performance-optimization.md
# .moai/docs/exploration-database-migration.md
```

### Sync Reports

```python
def generate_sync_report_filename():
    """Generate timestamped sync report filename"""
    timestamp = datetime.now().strftime("%Y%m%d")
    return f".moai/reports/sync-report-{timestamp}.md"

# Examples:
# .moai/reports/sync-report-20251105.md
# .moai/reports/sync-report-20251104.md
```
