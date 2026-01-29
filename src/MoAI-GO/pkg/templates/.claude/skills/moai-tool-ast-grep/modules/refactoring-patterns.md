# AST-Grep Refactoring Patterns

Common code transformation patterns for large-scale refactoring.

## API Migration

### Function Rename

```bash
# Simple rename
sg run --pattern 'oldFunction($$$ARGS)' --rewrite 'newFunction($$$ARGS)' --lang python

# With method chain
sg run --pattern '$OBJ.oldMethod($$$ARGS)' --rewrite '$OBJ.newMethod($$$ARGS)' --lang javascript
```

### Library Migration

```yaml
# axios to fetch
id: axios-to-fetch-get
language: typescript
rule:
  pattern: 'axios.get($URL)'
fix: 'fetch($URL).then(res => res.json())'

---
id: axios-to-fetch-post
language: typescript
rule:
  pattern: 'axios.post($URL, $DATA)'
fix: |
  fetch($URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify($DATA)
  }).then(res => res.json())
```

### Import Statement Update

```yaml
id: update-import-path
language: typescript
rule:
  pattern: "import { $$$IMPORTS } from 'old-package'"
fix: "import { $$$IMPORTS } from '@new-scope/new-package'"
```

## Code Modernization

### var to const/let

```yaml
id: var-to-const
language: javascript
rule:
  pattern: 'var $NAME = $VALUE'
  not:
    has:
      pattern: '$NAME = $OTHER'
      inside:
        not:
          pattern: 'var $NAME = $VALUE'
fix: 'const $NAME = $VALUE'
message: 'Prefer const for variables that are never reassigned'
```

### Callback to Async/Await

```yaml
id: callback-to-async
language: javascript
rule:
  pattern: |
    $FUNC($$$ARGS, function($ERR, $DATA) {
      $$$BODY
    })
fix: |
  const $DATA = await $FUNC($$$ARGS)
  $$$BODY
message: 'Consider using async/await instead of callbacks'
```

### Promise.then to Async/Await

```yaml
id: then-to-await
language: typescript
rule:
  pattern: '$PROMISE.then($CALLBACK)'
  inside:
    pattern: 'async function $NAME($$$ARGS) { $$$BODY }'
fix: 'await $PROMISE'
```

## React Patterns

### Class to Functional Component

```yaml
id: class-to-functional-state
language: typescriptreact
rule:
  pattern: 'this.state.$PROP'
  inside:
    pattern: 'class $NAME extends $$$BASE { $$$BODY }'
message: 'Consider converting to functional component with useState'
```

### componentDidMount to useEffect

```yaml
id: componentDidMount-to-useEffect
language: typescriptreact
rule:
  pattern: |
    componentDidMount() {
      $$$BODY
    }
fix: |
  useEffect(() => {
    $$$BODY
  }, [])
```

### PropTypes to TypeScript

```yaml
id: proptypes-to-typescript
language: typescriptreact
rule:
  pattern: '$COMPONENT.propTypes = { $$$PROPS }'
message: 'Consider using TypeScript interfaces instead of PropTypes'
```

## Python Patterns

### String Format to f-string

```yaml
id: format-to-fstring
language: python
rule:
  pattern: '"$$$STR".format($ARG)'
fix: 'f"$$$STR"'  # Note: requires manual adjustment of placeholders
message: 'Consider using f-strings for cleaner formatting'
```

### Dict Comprehension

```yaml
id: loop-to-dict-comprehension
language: python
rule:
  pattern: |
    $DICT = {}
    for $KEY in $ITER:
        $DICT[$KEY] = $VALUE
fix: '$DICT = {$KEY: $VALUE for $KEY in $ITER}'
```

### Context Manager

```yaml
id: file-to-context-manager
language: python
rule:
  pattern: |
    $FILE = open($PATH)
    $$$BODY
    $FILE.close()
fix: |
  with open($PATH) as $FILE:
      $$$BODY
```

## Go Patterns

### Error Handling

```yaml
id: error-wrap
language: go
rule:
  pattern: 'return $ERR'
  inside:
    pattern: 'if $ERR != nil { $$$BODY }'
fix: 'return fmt.Errorf("$FUNC: %w", $ERR)'
message: 'Wrap errors with context'
```

### Defer Pattern

```yaml
id: close-without-defer
language: go
rule:
  pattern: '$RESOURCE.Close()'
  not:
    inside:
      pattern: 'defer $RESOURCE.Close()'
message: 'Consider using defer for resource cleanup'
```

## TypeScript Patterns

### Type Assertion

```yaml
id: type-assertion-style
language: typescript
rule:
  pattern: '<$TYPE>$EXPR'
fix: '$EXPR as $TYPE'
message: 'Prefer "as" syntax for type assertions'
```

### Optional Chaining

```yaml
id: use-optional-chaining
language: typescript
rule:
  pattern: '$OBJ && $OBJ.$PROP'
fix: '$OBJ?.$PROP'
```

### Nullish Coalescing

```yaml
id: use-nullish-coalescing
language: typescript
rule:
  pattern: '$VAR !== null && $VAR !== undefined ? $VAR : $DEFAULT'
fix: '$VAR ?? $DEFAULT'
```

## Batch Refactoring

### Multi-file Transformation

```bash
# Find all files
sg run --pattern '$OLD($$$ARGS)' --lang python src/

# Preview changes
sg run --pattern '$OLD($$$ARGS)' --rewrite '$NEW($$$ARGS)' --lang python src/ --interactive

# Apply changes
sg run --pattern '$OLD($$$ARGS)' --rewrite '$NEW($$$ARGS)' --lang python src/ --update-all
```

### JSON Output for Review

```bash
sg scan --config rules.yml --json > changes.json
# Review changes before applying
cat changes.json | jq '.[] | {file: .path, line: .range.start.line, fix: .fix}'
```
