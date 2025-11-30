# Template Optimization & Smart Merge

Intelligent template merging that preserves user customizations while applying updates.

## 6-Phase Optimization Workflow

```
6-Phase Template Optimization:
 Phase 1: Backup Discovery & Analysis
 Scan .moai-backups/ directory
 Analyze backup metadata
 Select most recent backup
 Phase 2: Template Comparison
 Hash-based file comparison
 Detect user customizations
 Identify template defaults
 Phase 3: Smart Merge Algorithm
 Extract user content
 Apply template updates
 Resolve conflicts
 Phase 4: Template Default Detection
 Identify placeholder patterns
 Classify content (template/user/mixed)
 Phase 5: Version Management
 Track template versions
 Update HISTORY section
 Phase 6: Configuration Updates
 Set optimization flags
 Record customizations preserved
```

## Smart Merge Algorithm

```python
def smart_merge(backup, template, current):
 """Three-way merge with intelligence."""

 # Extract user customizations from backup
 user_content = extract_user_customizations(backup)

 # Get latest template defaults
 template_defaults = get_current_templates()

 # Merge with priority
 merged = {
 "template_structure": template_defaults, # Always latest
 "user_config": user_content, # Preserved
 "custom_content": user_content # Extracted
 }

 return merged
```

## Backup Structure

```json
{
 "backup_id": "backup-2025-11-24-v0.28.2",
 "created_at": "2025-11-24T10:30:00Z",
 "template_version": "0.28.2",
 "project_state": {
 "name": "my-project",
 "specs": ["SPEC-001", "SPEC-002"],
 "files_backed_up": 47
 },
 "customizations": {
 "language": "ko",
 "team_settings": {...},
 "domains": ["backend", "frontend"]
 }
}
```

## Restoration Process

```python
def restore_from_backup(backup_id: str):
 """Restore project from specific backup."""

 # Load backup metadata
 backup = load_backup(backup_id)

 # Validate backup integrity
 if not validate_backup_integrity(backup):
 raise BackupIntegrityError("Backup corrupted")

 # Extract user customizations
 customizations = extract_customizations(backup)

 # Apply to current project
 apply_customizations(customizations)

 # Update configuration
 update_config({
 "restored_from": backup_id,
 "restored_at": datetime.now()
 })
```

## Version Tracking

```json
{
 "template_optimization": {
 "last_optimized": "2025-11-24T12:00:00Z",
 "backup_version": "backup-2025-10-15-v0.27.0",
 "template_version": "0.28.2",
 "customizations_preserved": [
 "language",
 "team_settings",
 "domains"
 ],
 "optimization_flags": {
 "merge_applied": true,
 "conflicts_resolved": 0,
 "user_content_extracted": true
 }
 }
}
```

## History Section Updates

```markdown
## Template Update History

### v0.28.2 (2025-11-24)
- Optimization Applied: Yes
- Backup Used: backup-2025-10-15-v0.27.0
- Customizations Preserved: language (ko), team_settings
- Template Updates: 12 files updated
- Conflicts Resolved: 0
```

## Success Metrics

- Customization Preservation: 100% user content retained during updates
- Merge Success Rate: 99% conflicts resolved automatically
