import json
import sys

def get_value_by_path(data, path):
    keys = path.split('.')
    curr = data
    for key in keys:
        if isinstance(curr, dict) and key in curr:
            curr = curr[key]
        else:
            return None
    return curr

def verify_schema(schema_path, config_path):
    try:
        with open(schema_path, 'r') as f:
            schema = json.load(f)
        with open(config_path, 'r') as f:
            config = json.load(f)
    except Exception as e:
        print(f"Error loading files: {e}")
        return

    errors = []
    fields_checked = 0

    for tab in schema.get('tabs', []):
        for batch in tab.get('batches', []):
            for question in batch.get('questions', []):
                field_path = question.get('field')
                if not field_path:
                    continue
                
                fields_checked += 1
                value = get_value_by_path(config, field_path)
                
                if value is None:
                    # Special case for null values in config which are valid but return None here
                    # We need to check if the key actually exists
                    keys = field_path.split('.')
                    curr = config
                    exists = True
                    for key in keys:
                        if isinstance(curr, dict) and key in curr:
                            curr = curr[key]
                        else:
                            exists = False
                            break
                    if not exists:
                        errors.append(f"Field not found in config: {field_path}")
                
                # Optional: Check type consistency if needed, but existence is primary

    if errors:
        print(f"Found {len(errors)} errors:")
        for err in errors:
            print(f"- {err}")
    else:
        print(f"Success! Verified {fields_checked} fields. All exist in config.json.")

if __name__ == "__main__":
    verify_schema(
        '/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-project-batch-questions/tab_schema.json',
        '/Users/goos/MoAI/MoAI-ADK/.moai/config/config.json'
    )
