# Change Logging - Practical Examples

## Example 1: Track Changes

```python
# Log changes to database records
def log_change(entity_type: str, entity_id: str, action: str, changes: dict):
    log_entry = {
        'timestamp': datetime.now().isoformat(),
        'entity_type': entity_type,
        'entity_id': entity_id,
        'action': action,  # 'create', 'update', 'delete'
        'changes': changes,
        'user_id': current_user_id()
    }
    save_to_changelog(log_entry)
```

## Example 2: Version Tracking

```python
# Track version changes
class VersionedEntity:
    def __init__(self, data: dict):
        self.data = data
        self.version = 1
        self.change_log = []
    
    def update(self, new_data: dict):
        changes = compute_diff(self.data, new_data)
        self.change_log.append({
            'version': self.version,
            'changes': changes,
            'timestamp': datetime.now()
        })
        self.data = new_data
        self.version += 1
```

**Learn More**: See advanced-patterns.md for detailed logging strategies.
