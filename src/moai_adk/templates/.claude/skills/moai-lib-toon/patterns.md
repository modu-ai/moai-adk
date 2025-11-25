# TOON Format Patterns and Anti-Patterns

## Recommended Patterns

### Pattern 1: Table-First Design

**Description**: Always convert uniform object arrays to table format

**Good Example**:
```toon
users[3]{id,name,email,role}:
  1,Alice,alice@example.com,admin
  2,Bob,bob@example.com,user
  3,Charlie,charlie@example.com,user
```

**Bad Example**:
```toon
users[3]:
  user:
    id: 1
    name: Alice
    email: alice@example.com
    role: admin
  user:
    id: 2
    name: Bob
    email: bob@example.com
    role: user
  user:
    id: 3
    name: Charlie
    email: charlie@example.com
    role: user
```

**Reason**: Table format is 40% more efficient

### Pattern 2: Explicit Length Declaration

**Description**: Always declare array length to improve model parsing reliability

**Good Example**:
```toon
items[5]: item1,item2,item3,item4,item5
```

**Bad Example**:
```toon
items: item1,item2,item3,item4,item5  # Missing length
```

**Reason**: Explicit length enables LLM verification

### Pattern 3: Consistent Indentation

**Description**: Maintain same indentation style throughout the file

**Good Example**:
```toon
company:
  name: Acme
  departments:
    engineering:
      size: 50
    sales:
      size: 30
```

**Bad Example**:
```toon
company:
  name: Acme
    departments:  # Wrong indentation
      engineering:
        size: 50
  sales:  # Inconsistent indentation
    size: 30
```

**Reason**: Consistency prevents parsing errors

### Pattern 4: Smart Field Ordering

**Description**: Place important fields first to improve readability

**Good Example**:
```toon
products[3]{id,name,price,category,inStock}:
  1,Laptop,1200000,Electronics,true
  2,Mouse,30000,Electronics,true
  3,Keyboard,80000,Electronics,false
```

**Bad Example**:
```toon
products[3]{inStock,category,price,name,id}:
  true,Electronics,1200000,Laptop,1
  true,Electronics,30000,Mouse,2
  false,Electronics,80000,Keyboard,3
```

**Reason**: Natural field order increases data comprehension

### Pattern 5: Minimize Nesting

**Description**: Avoid 3+ level nesting, use key folding when necessary

**Good Example**:
```toon
# Use key folding
user.profile.name: Alice
user.profile.age: 30
user.settings.theme: dark
```

**Bad Example**:
```toon
user:
  profile:
    details:
      personal:
        name: Alice  # 5-level nesting
```

**Reason**: Deep nesting reduces token efficiency

### Pattern 6: Explicit Type Conversion

**Description**: Clearly distinguish numbers and strings

**Good Example**:
```toon
data[2]{id,count,price,name}:
  1,100,19.99,Widget
  2,200,29.99,Gadget
```

**Bad Example**:
```toon
data[2]{id,count,price,name}:
  "1","100","19.99","Widget"  # Unnecessary quotes
  2,200,29.99,Gadget
```

**Reason**: Explicit types improve parsing accuracy

### Pattern 7: Comment Usage

**Description**: Add explanations with comments for complex structures

**Good Example**:
```toon
# User analytics data
# Period: 2025-11-01 ~ 2025-11-30
analytics:
  totalUsers: 10000
  activeUsers: 8500  # Monthly active users
  # Daily statistics
  daily[30]{date,visitors,conversions}:
    2025-11-01,1250,45
    # ... 28 more days
```

**Reason**: Comments provide data context

## Anti-Patterns

### Anti-Pattern 1: Unnecessary Object Nesting

**Problem**:
```toon
# Inefficient
items[3]:
  item:
    id: 1
    name: Widget
  item:
    id: 2
    name: Gadget
  item:
    id: 3
    name: Tool
```

**Solution**:
```toon
# Efficient (40% token savings)
items[3]{id,name}:
  1,Widget
  2,Gadget
  3,Tool
```

### Anti-Pattern 2: Mixed Delimiters

**Problem**:
```toon
users[2]{name,age}:
  Alice,30
  Bob|25  # Error: Mixed delimiters
```

**Solution**:
```toon
users[2]{name,age}:
  Alice,30
  Bob,25
```

### Anti-Pattern 3: Field Count Mismatch

**Problem**:
```toon
products[3]{id,name,price}:
  1,Laptop,1200000
  2,Mouse,30000,InStock  # 4 fields (3 expected)
  3,Keyboard  # 2 fields (3 expected)
```

**Solution**:
```toon
# Option 1: Add field
products[3]{id,name,price,status}:
  1,Laptop,1200000,null
  2,Mouse,30000,InStock
  3,Keyboard,80000,null

# Option 2: Use object format
products[3]:
  product:
    id: 1
    name: Laptop
    price: 1200000
  product:
    id: 2
    name: Mouse
    price: 30000
    status: InStock
  product:
    id: 3
    name: Keyboard
    price: 80000
```

### Anti-Pattern 4: Tab and Space Mixing

**Problem**:
```toon
company:
  name: Acme  # 2 spaces
	address: Seoul  # 1 tab (error)
```

**Solution**:
```toon
company:
  name: Acme
  address: Seoul  # Consistent 2 spaces
```

### Anti-Pattern 5: Unescaped Special Characters

**Problem**:
```toon
messages[2]: Hello, World,How are you?  # Comma collision
```

**Solution**:
```toon
messages[2]: "Hello, World","How are you?"
```

### Anti-Pattern 6: Excessive Key Folding

**Problem**:
```toon
# Reduced readability
user.profile.personal.details.name.first: Alice
user.profile.personal.details.name.last: Kim
user.profile.personal.details.age: 30
```

**Solution**:
```toon
# Balanced nesting
user:
  profile:
    name.first: Alice
    name.last: Kim
    age: 30
```

### Anti-Pattern 7: Meaningless Array Length

**Problem**:
```toon
items[0]:  # Empty array (meaningless)
```

**Solution**:
```toon
items: []  # Or remove field
```

## Optimization Patterns

### Pattern A: Batch Conversion

**Scenario**: Convert multiple JSON files to TOON in batch

**Implementation**:
```bash
#!/bin/bash
# batch_convert.sh

INPUT_DIR="./json_files"
OUTPUT_DIR="./toon_files"

mkdir -p "$OUTPUT_DIR"

for json_file in "$INPUT_DIR"/*.json; do
  base_name=$(basename "$json_file" .json)
  toon convert "$json_file" --output "$OUTPUT_DIR/${base_name}.toon"

  if [ $? -eq 0 ]; then
    echo "✓ $base_name converted"
  else
    echo "✗ $base_name failed"
  fi
done
```

### Pattern B: Streaming Conversion

**Scenario**: Process large data streaming

**Implementation**:
```typescript
import { encodeLines } from '@toon-format/toon'
import { createWriteStream } from 'fs'

async function streamConvert(data: any, outputPath: string) {
  const stream = createWriteStream(outputPath)

  for (const line of encodeLines(data)) {
    stream.write(line + '\n')
  }

  stream.end()
}
```

### Pattern C: Conditional Conversion

**Scenario**: Select format based on data characteristics

**Implementation**:
```typescript
function smartEncode(data: any): string {
  const eligibility = calculateTabularEligibility(data)

  if (eligibility >= 80) {
    // Table format optimal
    return encode(data, { detectTabular: true })
  } else if (eligibility >= 50) {
    // Mixed format
    return encode(data, {
      detectTabular: true,
      keyFolding: true
    })
  } else {
    // Keep JSON
    return JSON.stringify(data)
  }
}
```

### Pattern D: Caching Strategy

**Scenario**: Optimize repeated conversions

**Implementation**:
```typescript
class ToonCache {
  private cache = new Map<string, string>()

  encode(data: any): string {
    const key = JSON.stringify(data)

    if (this.cache.has(key)) {
      return this.cache.get(key)!
    }

    const toon = encode(data)
    this.cache.set(key, toon)

    return toon
  }

  clear() {
    this.cache.clear()
  }
}
```

## Debugging Patterns

### Pattern X: Step-by-Step Validation

**Description**: Perform validation at each conversion step

**Implementation**:
```typescript
function debugConvert(data: any): string {
  console.log('1. Input validation...')
  validateInput(data)

  console.log('2. Encoding...')
  const toon = encode(data)

  console.log('3. Round-trip test...')
  const decoded = decode(toon)

  console.log('4. Deep equality check...')
  assert.deepStrictEqual(data, decoded)

  console.log('✓ All checks passed')
  return toon
}
```

### Pattern Y: Error Recovery

**Description**: Attempt automatic recovery on parsing errors

**Implementation**:
```typescript
function resilientDecode(toon: string): any {
  try {
    // Try strict mode
    return decode(toon, { strict: true })
  } catch (error) {
    console.warn('Strict mode failed, retrying with non-strict...')

    try {
      // Non-strict mode fallback
      return decode(toon, { strict: false })
    } catch (error2) {
      console.error('Both modes failed')
      throw error2
    }
  }
}
```

### Pattern Z: Difference Analysis

**Description**: Compare data before and after conversion

**Implementation**:
```typescript
function analyzeDifference(original: any, converted: any) {
  const diff = {
    added: [],
    removed: [],
    changed: []
  }

  // Recursive comparison logic
  function compare(obj1: any, obj2: any, path: string = '') {
    // ... Implementation omitted
  }

  compare(original, converted)

  return diff
}
```

## Migration Patterns

### Pattern M1: Gradual Transition

**Description**: Phase-wise migration of existing JSON system to TOON

**Implementation**:
```typescript
class HybridFormat {
  encode(data: any, format: 'json' | 'toon' | 'auto'): string {
    if (format === 'auto') {
      // Analyze data characteristics
      const eligibility = calculateTabularEligibility(data)
      format = eligibility >= 80 ? 'toon' : 'json'
    }

    return format === 'toon'
      ? encode(data)
      : JSON.stringify(data)
  }

  decode(text: string): any {
    // Auto-detect format
    if (text.includes('{') && text.includes('}')) {
      return JSON.parse(text)
    } else {
      return decode(text)
    }
  }
}
```

### Pattern M2: Compatibility Layer

**Description**: Ensure JSON/TOON bidirectional compatibility

**Implementation**:
```typescript
interface UniversalData {
  format: 'json' | 'toon'
  data: any
  metadata: {
    created: Date
    version: string
  }
}

class DataAdapter {
  static from(input: string): UniversalData {
    const format = detectFormat(input)
    const data = format === 'json'
      ? JSON.parse(input)
      : decode(input)

    return {
      format,
      data,
      metadata: {
        created: new Date(),
        version: '1.0.0'
      }
    }
  }

  static to(universal: UniversalData, targetFormat: 'json' | 'toon'): string {
    return targetFormat === 'json'
      ? JSON.stringify(universal.data)
      : encode(universal.data)
  }
}
```

## Performance Optimization Patterns

### Pattern P1: Lazy Conversion

**Description**: Delay conversion until needed

**Implementation**:
```typescript
class LazyToon {
  private data: any
  private _toon: string | null = null

  constructor(data: any) {
    this.data = data
  }

  get toon(): string {
    if (this._toon === null) {
      this._toon = encode(this.data)
    }
    return this._toon
  }
}
```

### Pattern P2: Memoization

**Description**: Cache results for identical inputs

**Implementation**:
```typescript
import memoize from 'lodash/memoize'

const memoizedEncode = memoize(
  (data: any) => encode(data),
  (data: any) => JSON.stringify(data)  // Cache key
)
```

### Pattern P3: Parallel Processing

**Description**: Convert multiple files simultaneously

**Implementation**:
```typescript
import { Worker } from 'worker_threads'

async function parallelConvert(files: string[]): Promise<string[]> {
  const workers = files.map(file => {
    return new Promise((resolve, reject) => {
      const worker = new Worker('./convert-worker.js', {
        workerData: { file }
      })

      worker.on('message', resolve)
      worker.on('error', reject)
    })
  })

  return Promise.all(workers)
}
```

---

**This pattern guide provides effective usage of TOON format.**