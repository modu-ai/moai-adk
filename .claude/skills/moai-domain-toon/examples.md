# Examples

```json
{
  "context": {
    "task": "Our favorite hikes together",
    "location": "Boulder",
    "season": "spring_2025"
  },
  "friends": ["ana", "luis", "sam"],
  "hikes": [
    {
      "id": 1,
      "name": "Blue Lake Trail",
      "distanceKm": 7.5,
      "elevationGain": 320,
      "companion": "ana",
      "wasSunny": true
    },
    {
      "id": 2,
      "name": "Ridge Overlook",
      "distanceKm": 9.2,
      "elevationGain": 540,
      "companion": "luis",
      "wasSunny": false
    }
  ]
}
```

```toon
context:
  task: Our favorite hikes together
  location: Boulder
  season: spring_2025
friends[3]: ana,luis,sam
hikes[2]{id,name,distanceKm,elevationGain,companion,wasSunny}:
  1,Blue Lake Trail,7.5,320,ana,true
  2,Ridge Overlook,9.2,540,luis,false
```

```toon
# 들여쓰기 기반 중첩
user:
  name: Alice
  age: 30
  email: alice@example.com
```

```toon
# 단순 배열 (길이 명시)
colors[3]: red,green,blue

# 객체 배열 (필드 헤더 포함)
users[2]{name,age}:
  Alice,30
  Bob,25
```

```toon
# 숫자
count: 42
price: 19.99

# 불리언
isActive: true
hasError: false

# 문자열 (따옴표 선택적)
title: Hello World
description: "Special chars: , : [ ]"
```

```toon
company:
  name: Acme Corp
  employees[3]{id,name,role}:
    1,Alice,Engineer
    2,Bob,Designer
    3,Charlie,Manager
  locations[2]: Seoul,Busan
```

```bash
npm install @toon-format/toon
```

```typescript
import { parse, stringify } from '@toon-format/toon'

// TOON → JSON 파싱
const data = parse(`
  users[2]{name,age}:
    Alice,30
    Bob,25
`)
console.log(data)
// { users: [{ name: 'Alice', age: 30 }, { name: 'Bob', age: 25 }] }

// JSON → TOON 변환
const toon = stringify({
  users: [
    { name: 'Alice', age: 30 },
    { name: 'Bob', age: 25 }
  ]
})
console.log(toon)
// users[2]{name,age}:
//   Alice,30
//   Bob,25
```

```typescript
import { parse, stringify, EncodeOptions } from '@toon-format/toon'

// 인코딩 옵션
const options: EncodeOptions = {
  strict: true,           // 엄격 모드 (기본값: false)
  delimiter: ',',         // 구분자 (기본값: ,)
  indentation: '  ',      // 들여쓰기 (기본값: 2 spaces)
  detectTabular: true     // 테이블 자동 감지
}

const toon = stringify(data, options)
```

```bash
npm install -g @toon-format/cli
```

```bash
# JSON → TOON
toon convert data.json --output data.toon

# TOON → JSON
toon convert data.toon --output data.json

# 표준 입력/출력 사용
cat data.json | toon convert --format toon

# 대안 구분자 사용
toon convert data.json --delimiter '|' --output data.toon
```

```bash
# TOON 파일 검증
toon validate data.toon

# JSON 호환성 테스트
toon validate data.toon --round-trip

# 상세 오류 출력
toon validate data.toon --verbose
```

```bash
# 토큰 통계 분석
toon benchmark data.json

# 상세 비교 리포트
toon benchmark data.json --compare json,toon,csv

# 정확도 테스트
toon benchmark data.json --accuracy-test
```

```bash
pip install toon-format
```

```python
from toon import parse, stringify

# TOON → Python dict
data = parse("""
users[2]{name,age}:
  Alice,30
  Bob,25
""")

# Python dict → TOON
toon_str = stringify({
    'users': [
        {'name': 'Alice', 'age': 30},
        {'name': 'Bob', 'age': 25}
    ]
})
```

```toon
# [길이]{필드1,필드2,...}:
items[3]{id,name,price}:
  1,Widget,9.99
  2,Gadget,19.99
  3,Tool,14.99
```

```toon
# 숫자, 불리언, 문자열 혼합
products[2]{id,name,inStock,price}:
  1,Laptop,true,999.99
  2,Mouse,false,29.99
```

```toon
# 복잡한 구조
orders[2]:
  order:
    id: 1
    items[2]{product,qty}:
      Laptop,1
      Mouse,2
  order:
    id: 2
    items[1]{product,qty}:
      Keyboard,1
```

```toon
# 특수 문자 포함 시 따옴표 필수
messages[3]: 
  "Hello, world!"
  "Line1
Line2"
  "Contains: comma, colon: bracket["
```

```toon
# 영문자, 숫자, 언더스코어만 포함 시 따옴표 불필요
identifiers[4]: user_id,product_name,item123,category_A
```

```toon
# 올바른 예시 (2칸 들여쓰기)
company:
  name: Acme
  departments:
    engineering:
      size: 50
    sales:
      size: 30

# 잘못된 예시 (혼합 들여쓰기)
company:
  name: Acme
    departments:  # 오류: 들여쓰기 불일치
      engineering:
        size: 50
```

```typescript
function encode(data: any, options?: EncodeOptions): string
```

```typescript
const data = {
  users: [
    { name: 'Alice', age: 30 },
    { name: 'Bob', age: 25 }
  ]
}

const toon = encode(data, {
  strict: true,
  detectTabular: true
})
```

```typescript
function decode(toon: string, options?: DecodeOptions): any
```

```typescript
const toon = `
users[2]{name,age}:
  Alice,30
  Bob,25
`

const data = decode(toon, {
  strict: false,
  validateStructure: true
})
```

```typescript
function* encodeLines(data: any, options?: EncodeOptions): Generator<string>
```

```typescript
const largeData = { /* ... */ }

for (const line of encodeLines(largeData)) {
  process.stdout.write(line)
  process.stdout.write('\n')
}
```

```typescript
interface EncodeOptions {
  strict?: boolean           // 엄격 모드 (기본: false)
  delimiter?: string         // 구분자 (기본: ',')
  indentation?: string       // 들여쓰기 (기본: '  ')
  detectTabular?: boolean    // 테이블 자동 감지 (기본: true)
  keyFolding?: boolean       // 키 폴딩 활성화 (기본: false)
  pathExpansion?: boolean    // 경로 확장 (기본: false)
}
```

```typescript
// strict: false (기본값)
// - 느슨한 파싱 (문법 오류 허용)
// - 호환성 우선

// strict: true
// - 엄격한 파싱 (문법 오류 시 예외 발생)
// - 정확성 우선
const data = decode(toon, { strict: true })
```

```typescript
// 기본 구분자: ,
users[2]{name,age}:
  Alice,30
  Bob,25

// 대안 구분자: |
const toon = encode(data, { delimiter: '|' })
// users[2]{name|age}:
//   Alice|30
//   Bob|25
```

```typescript
// detectTabular: true (기본값)
// - 균일한 객체 배열을 테이블로 변환

// detectTabular: false
// - 모든 배열을 명시적 객체 구조로 유지
const toon = encode(data, { detectTabular: false })
```

```typescript
interface DecodeOptions {
  strict?: boolean              // 엄격 모드
  validateStructure?: boolean   // 구조 검증 (기본: true)
  allowPartial?: boolean        // 부분 파싱 허용 (기본: false)
}
```

```bash
# JSON → TOON
toon convert input.json --output output.toon

# TOON → JSON
toon convert input.toon --output output.json

# 자동 형식 감지
toon convert data.json  # 확장자 기반 형식 추론
```

```bash
# 엄격 모드
toon convert data.json --strict --output data.toon

# 대안 구분자
toon convert data.json --delimiter '|' --output data.toon

# 들여쓰기 설정
toon convert data.json --indent 4 --output data.toon

# 키 폴딩 활성화
toon convert data.json --key-folding --output data.toon
```

```bash
# 표준 입력/출력
cat data.json | toon convert --format toon > data.toon

# 압축과 함께 사용
cat large.json | toon convert --format toon | gzip > data.toon.gz

# 원격 파일 변환
curl https://api.example.com/data | toon convert --format toon
```

```bash
# 기본 검증
toon validate data.toon

# 상세 오류 출력
toon validate data.toon --verbose

# JSON 호환성 테스트
toon validate data.toon --round-trip

# 스키마 검증
toon validate data.toon --schema schema.json
```

```bash
# 성공 (exit code: 0)
✓ Valid TOON format
✓ Round-trip test passed
✓ 100% structure integrity

# 실패 (exit code: 1)
✗ Parse error at line 5: Missing delimiter
✗ Structure mismatch: Expected 3 fields, got 2
✗ Round-trip test failed: Data loss detected
```

```bash
# 기본 통계
toon benchmark data.json

출력 예시:
Format   | Tokens | Size (KB) | Reduction
---------|--------|-----------|----------
JSON     | 1,247  | 8.2       | 0%
TOON     | 753    | 4.9       | 39.6%
CSV      | 695    | 4.5       | 44.3%
```

```bash
# 여러 형식 비교
toon benchmark data.json --compare json,toon,csv,yaml

# 정확도 테스트
toon benchmark data.json --accuracy-test --model claude-haiku-4-5

# 파일 출력
toon benchmark data.json --output-report report.md
```

```bash
# 디렉토리 전체 분석
toon benchmark-dir ./data/*.json --summary

# 병렬 처리
toon benchmark-dir ./data/*.json --parallel 4

# 결과 집계
toon benchmark-dir ./data/*.json --aggregate results.csv
```

```toon
user:
  profile:
    name: Alice
    age: 30
  settings:
    theme: dark
    notifications: true
```

```toon
user.profile.name: Alice
user.profile.age: 30
user.settings.theme: dark
user.settings.notifications: true
```

```typescript
const toon = encode(data, {
  keyFolding: true
})
```

```typescript
// 입력 (Key Folding 형식)
const folded = `
user.profile.name: Alice
user.profile.age: 30
`

// 출력 (Path Expansion)
const expanded = decode(folded, {
  pathExpansion: true
})
// {
//   user: {
//     profile: {
//       name: 'Alice',
//       age: 30
//     }
//   }
// }
```

```typescript
// strict: true
// - 문법 오류 시 예외 발생
// - 필드 수 불일치 감지
// - 타입 검증 강화

try {
  const data = decode(toon, { strict: true })
} catch (error) {
  console.error('Parse error:', error.message)
}
```

```typescript
// strict: false (기본값)
// - 문법 오류 허용 (최대한 파싱)
// - 필드 수 불일치 무시
// - 타입 자동 변환

const data = decode(toon, { strict: false })
```

```toon
# 최상위 배열
[3]{name,age}:
  Alice,30
  Bob,25
  Charlie,35

# 최상위 객체 (암묵적)
name: Alice
age: 30
```

```typescript
// 자동 감지
const data1 = decode(`
[2]{name,age}:
  Alice,30
  Bob,25
`)
// [{ name: 'Alice', age: 30 }, { name: 'Bob', age: 25 }]

const data2 = decode(`
name: Alice
age: 30
`)
// { name: 'Alice', age: 30 }
```

```
데이터 포맷 선택
    ↓
균일한 객체 배열?
    ├─ YES → TOON (테이블 형식)
    │         - 토큰 절감: 35-40%
    │         - 정확도: 73.9%
    │
    └─ NO → 복잡한 중첩 구조?
              ├─ YES → JSON (전통적 방식)
              │         - 호환성 우수
              │         - 도구 지원 풍부
              │
              └─ NO → 순수 테이블 데이터?
                        ├─ YES → CSV (최대 압축)
                        │         - 토큰 절감: 45%
                        │         - 단순 구조만
                        │
                        └─ NO → TOON (균형)
                                  - 유연성
                                  - 토큰 효율성
```

```typescript
// API 응답 데이터 포함
const prompt = `
Analyze these user activities:

${encode(activityData)}

Provide insights on engagement patterns.
`
```

```bash
# 5000개 레코드 변환
toon convert large_dataset.json --output dataset.toon

# 토큰 절감 확인
toon benchmark dataset.toon
```

```toon
# config.toon
database:
  host: localhost
  port: 5432
  credentials:
    user: admin
    password: "secret123"

services[3]{name,port,enabled}:
  api,8080,true
  worker,8081,true
  scheduler,8082,false
```

```toon
# 시계열 데이터
metrics[24]{hour,users,requests,errors}:
  0,120,1450,2
  1,95,1120,1
  2,78,890,0
  ...
```

```json
{
  "company": {
    "departments": {
      "engineering": {
        "teams": {
          "backend": {
            "members": {
              "seniors": [...],
              "juniors": [...]
            }
          }
        }
      }
    }
  }
}
```

```csv
name,age,email
Alice,30,alice@example.com
Bob,25,bob@example.com
```

```typescript
// 실시간 API
app.post('/api/data', (req, res) => {
  const data = JSON.parse(req.body)  // 빠름
  // vs
  const data = decode(req.body)      // 약간 느림
})
```

```
Tabular Eligibility = (Uniform Fields ÷ Total Fields) × 100%

적합성:
- ≥80%: TOON 테이블 형식 강력 추천
- 50-79%: TOON 사용 고려
- <50%: JSON 유지 권장
```

```typescript
const data = {
  users: [
    { name: 'Alice', age: 30, email: 'alice@x.com' },      // 3 fields
    { name: 'Bob', age: 25, email: 'bob@x.com' },          // 3 fields
    { name: 'Charlie', age: 35 }                           // 2 fields
  ]
}

// Uniform Fields: 2 (name, age)
// Total Fields: 3 (name, age, email)
// Eligibility: 66.7% → TOON 사용 고려
```

```typescript
function assessTabularEligibility(data: any[]): number {
  const fieldCounts = data.map(obj => Object.keys(obj).length)
  const minFields = Math.min(...fieldCounts)
  const maxFields = Math.max(...fieldCounts)
  
  return (minFields / maxFields) × 100
}

const eligibility = assessTabularEligibility(users)
if (eligibility >= 80) {
  return encode(users, { detectTabular: true })
} else {
  return JSON.stringify(users)
}
```

```bash
# 기존 JSON 파일 분석
toon benchmark existing_data.json --detailed

# 출력:
# Tabular Eligibility: 85%
# Token Reduction: 38%
# Accuracy Improvement: +4.2%
# Recommendation: MIGRATE
```

```bash
# 배치 변환
find ./data -name "*.json" -exec toon convert {} \;

# 검증
find ./data -name "*.toon" -exec toon validate {} \;
```

```typescript
// 애플리케이션 코드 업데이트
import { decode } from '@toon-format/toon'
import fs from 'fs'

// Before
const data = JSON.parse(fs.readFileSync('data.json', 'utf-8'))

// After
const data = decode(fs.readFileSync('data.toon', 'utf-8'))
```

```typescript
// 성능 추적
const metrics = {
  tokensBefore: countTokens(JSON.stringify(data)),
  tokensAfter: countTokens(encode(data)),
  reduction: calculateReduction()
}

console.log(`Token reduction: ${metrics.reduction}%`)
```

```
ParseError: Unexpected delimiter at line 5
```

```toon
# 잘못된 예시
items[3]{name,price}:
  Widget,9.99
  Gadget,19.99,InStock  # 오류: 3개 필드 (2개 예상)

# 올바른 예시
items[3]{name,price,status}:
  Widget,9.99,InStock
  Gadget,19.99,OutOfStock
  Tool,14.99,InStock
```

```
ParseError: Unexpected character ':' in unquoted string
```

```toon
# 잘못된 예시
title: Hello: World  # 오류: : 포함

# 올바른 예시
title: "Hello: World"
```

```
ParseError: Inconsistent indentation at line 8
```

```bash
# 들여쓰기 확인
cat -A data.toon  # ^I는 탭, 공백은 스페이스

# 자동 수정
toon format data.toon --indent 2 --output data_fixed.toon
```

```
ValidationError: Expected 3 items, got 2
```

```toon
# 잘못된 예시
items[3]: item1,item2  # 오류: 2개 (3개 예상)

# 올바른 예시
items[2]: item1,item2
```

```javascript
// 예상: 숫자 30
// 실제: 문자열 "30"
```

```typescript
// Strict 모드 사용
const data = decode(toon, { strict: true })

// 또는 명시적 타입 변환
const age = Number(data.age)
```

```typescript
import { decode } from '@toon-format/toon'

try {
  const data = decode(toon, {
    strict: true,
    verbose: true  // 상세 로그 출력
  })
} catch (error) {
  console.error('Parse error:', error)
  console.error('Line:', error.line)
  console.error('Column:', error.column)
}
```

```bash
# 작은 파일부터 시작
toon convert small.json --output small.toon
toon validate small.toon --round-trip

# 성공하면 대용량 파일로 확장
toon convert large.json --output large.toon
```

```bash
# 원본과 변환 결과 비교
toon convert data.toon --output recovered.json
diff original.json recovered.json
```

```typescript
// JSON Schema 정의
const schema = {
  type: 'object',
  properties: {
    users: {
      type: 'array',
      items: {
        type: 'object',
        required: ['name', 'age']
      }
    }
  }
}

// 검증
const data = decode(toon)
validate(data, schema)  // JSON Schema 라이브러리 사용
```

```bash
npm install @toon-format/toon
```

```bash
npm install -g @toon-format/cli
```

```bash
code --install-extension toon-format.toon-vscode
```

```bash
pip install toon-format
```

```python
from toon import parse, stringify

data = parse(toon_string)
toon = stringify(python_dict)
```

```bash
go get github.com/toon-format/toon-go
```

```go
import "github.com/toon-format/toon-go"

data, err := toon.Parse(toonString)
toonString, err := toon.Stringify(data)
```

```bash
cargo add toon-format
```

```rust
use toon_format::{parse, stringify};

let data = parse(toon_str)?;
let toon = stringify(&data)?;
```

```bash
dotnet add package ToonFormat
```

```csharp
using ToonFormat;

var data = Toon.Parse(toonString);
var toon = Toon.Stringify(data);
```

```typescript
// Context7를 통한 최신 TOON 문서 접근
import { Context7Client } from '@context7/client'

const context7 = new Context7Client()

const toonDocs = await context7.get_library_docs({
  context7_library_id: '/toon-format/toon',
  topic: 'API reference encoding decoding 2025',
  tokens: 5000
})

console.log(toonDocs)
```

```typescript
// Context7를 통한 성능 최적화 패턴 접근
const perfPatterns = await context7.get_library_docs({
  context7_library_id: '/toon-format/toon',
  topic: 'performance optimization benchmarks token efficiency',
  tokens: 3000
})

// 최신 벤치마크 데이터 적용
const optimizedConfig = applyContext7Patterns(perfPatterns)
```

```typescript
// Context7를 통한 보안 패턴 접근
const securityDocs = await context7.get_library_docs({
  context7_library_id: '/toon-format/toon',
  topic: 'security validation input sanitization',
  tokens: 2000
})

// 보안 검증 로직 구현
implementSecurityPatterns(securityDocs)
```

```typescript
// 라이브러리 이름을 Context7 ID로 변환
const libraryId = await context7.resolve_library_id('toon')
// 결과: /toon-format/toon

// 최신 문서 가져오기
const docs = await context7.get_library_docs({
  context7_library_id: libraryId,
  tokens: 4000
})
```

```typescript
// 특정 버전 문서 접근
const v2Docs = await context7.get_library_docs({
  context7_library_id: '/toon-format/toon/v2.0.0',
  topic: 'breaking changes migration guide',
  tokens: 3000
})
```

```json
{
  "users": [
    {
      "id": 1,
      "name": "Alice Kim",
      "email": "alice@example.com",
      "role": "Engineer",
      "active": true
    },
    {
      "id": 2,
      "name": "Bob Lee",
      "email": "bob@example.com",
      "role": "Designer",
      "active": true
    },
    {
      "id": 3,
      "name": "Charlie Park",
      "email": "charlie@example.com",
      "role": "Manager",
      "active": false
    }
  ],
  "metadata": {
    "total": 3,
    "page": 1,
    "perPage": 10
  }
}
```

```toon
users[3]{id,name,email,role,active}:
  1,Alice Kim,alice@example.com,Engineer,true
  2,Bob Lee,bob@example.com,Designer,true
  3,Charlie Park,charlie@example.com,Manager,false
metadata:
  total: 3
  page: 1
  perPage: 10
```

```json
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "myapp",
    "pool": {
      "min": 2,
      "max": 10,
      "idle": 30000
    }
  },
  "redis": {
    "host": "localhost",
    "port": 6379,
    "db": 0
  },
  "services": [
    {
      "name": "api",
      "port": 8080,
      "enabled": true
    },
    {
      "name": "worker",
      "port": 8081,
      "enabled": true
    }
  ]
}
```

```toon
database:
  host: localhost
  port: 5432
  name: myapp
  pool:
    min: 2
    max: 10
    idle: 30000
redis:
  host: localhost
  port: 6379
  db: 0
services[2]{name,port,enabled}:
  api,8080,true
  worker,8081,true
```

```json
{
  "products": [
    {
      "id": "P001",
      "name": "Laptop",
      "category": "Electronics",
      "price": 999.99,
      "stock": 15,
      "active": true
    },
    {
      "id": "P002",
      "name": "Mouse",
      "category": "Electronics",
      "price": 29.99,
      "stock": 150,
      "active": true
    },
    {
      "id": "P003",
      "name": "Keyboard",
      "category": "Electronics",
      "price": 79.99,
      "stock": 80,
      "active": true
    }
  ]
}
```

```toon
products[3]{id,name,category,price,stock,active}:
  P001,Laptop,Electronics,999.99,15,true
  P002,Mouse,Electronics,29.99,150,true
  P003,Keyboard,Electronics,79.99,80,true
```

```toon
metrics:
  source: server-01
  interval: 1h
  start: 2025-11-21T00:00:00Z
data[24]{hour,cpu,memory,requests,errors}:
  0,12.5,45.2,1450,2
  1,11.8,43.1,1120,1
  2,10.2,41.5,890,0
  3,9.8,40.2,780,0
  4,10.5,41.8,820,1
  # ... 19 more hours
```