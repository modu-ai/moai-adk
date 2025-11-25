# TOON Format Reference

## Official Specification (TOON Specification v2.0)

### Basic Grammar

#### Object
```ebnf
object ::= (key ':' value | key ':' newline indent object dedent)*
key ::= identifier | quoted_string
value ::= primitive | array | object
```

#### Array
```ebnf
array ::= simple_array | tabular_array

simple_array ::= '[' length ']' ':' value (',' value)*

tabular_array ::= '[' length ']' '{' fields '}' ':' newline
                  (row_data newline)*

fields ::= identifier (',' identifier)*
row_data ::= value (',' value)*
```

#### Primitives
```ebnf
primitive ::= number | boolean | string | null

number ::= integer | decimal
boolean ::= 'true' | 'false'
string ::= identifier | quoted_string
null ::= 'null'
```

### Tokenization Rules

#### Indentation
- **Spaces**: 2 or 4 spaces recommended
- **Tabs**: Available but mixing prohibited
- **Consistency**: Maintain same method throughout file

#### Delimiter
- **Default**: `,` (comma)
- **Alternatives**: `|`, `;`, `\t` (tab)
- **Configuration**: Change with `--delimiter` option

#### Quoting
**No quotes needed** (identifiers):
- Starts with letters
- Contains only letters, numbers, underscores
- Examples: `user_id`, `firstName`, `item123`

**Quotes required** (strings):
- Contains special characters: `,`, `:`, `[`, `]`, `{`, `}`
- Contains spaces
- Empty strings
- Examples: `"Hello, World"`, `"name: value"`, `""`

### Type System

#### Number
```toon
# Integers
count: 42
negative: -17

# Floating point
price: 19.99
scientific: 1.23e-4

# Special values
infinity: Infinity
notANumber: NaN
```

#### Boolean
```toon
isActive: true
hasError: false

# Case sensitive
valid: true
invalid: True  # Error (uppercase T)
```

#### String
```toon
# No quotes (identifiers)
name: Alice
category: electronics

# With quotes (special characters)
message: "Hello, World!"
path: "/usr/local/bin"

# Escape sequences
escaped: "Line1\nLine2"
quoted: "She said \"Hello\""
```

#### Null
```toon
value: null
optional: null
```

### Array Declaration

#### Length Specification
```toon
# Required: [N] format
items[3]: item1,item2,item3

# Error: Missing length
items: item1,item2,item3  # Invalid syntax
```

#### Field Header
```toon
# Table format
users[2]{name,age}:
  Alice,30
  Bob,25

# Error: Field count mismatch
users[2]{name,age}:
  Alice,30,extra  # 3 fields (2 expected)
```

## TypeScript Type Definitions

### Basic Types
```typescript
// TOON value type
type ToonValue =
  | string
  | number
  | boolean
  | null
  | ToonObject
  | ToonArray

// TOON object
interface ToonObject {
  [key: string]: ToonValue
}

// TOON array
type ToonArray = ToonValue[]
```

### Option Interfaces

#### EncodeOptions
```typescript
interface EncodeOptions {
  /**
   * Enable strict mode
   * @default false
   */
  strict?: boolean

  /**
   * Array delimiter
   * @default ','
   */
  delimiter?: string

  /**
   * Indentation string
   * @default '  ' (2 spaces)
   */
  indentation?: string

  /**
   * Auto-detect tabular format
   * @default true
   */
  detectTabular?: boolean
  
  /**
   * Enable key folding
   * @default false
   */
  keyFolding?: boolean

  /**
   * Enable path expansion
   * @default false
   */
  pathExpansion?: boolean
}
```

#### DecodeOptions
```typescript
interface DecodeOptions {
  /**
   * Enable strict mode
   * @default false
   */
  strict?: boolean

  /**
   * Perform structure validation
   * @default true
   */
  validateStructure?: boolean

  /**
   * Allow partial parsing
   * @default false
   */
  allowPartial?: boolean

  /**
   * Type coercion
   * @default true
   */
  coerceTypes?: boolean
}
```

### Error Types

#### ParseError
```typescript
class ParseError extends Error {
  readonly line: number
  readonly column: number
  readonly position: number
  
  constructor(
    message: string,
    line: number,
    column: number,
    position: number
  )
}
```

#### ValidationError
```typescript
class ValidationError extends Error {
  readonly field: string
  readonly expected: string
  readonly actual: string
  
  constructor(
    message: string,
    field: string,
    expected: string,
    actual: string
  )
}
```

## CLI 명령어 레퍼런스

### convert
```bash
toon convert [input] [options]

Options:
  -o, --output <file>     Output file path
  -f, --format <fmt>      Output format (json|toon)
  -d, --delimiter <char>  Delimiter character
  -i, --indent <num>      Indentation spaces
  --strict                Enable strict mode
  --key-folding           Enable key folding
  --detect-tabular        Auto-detect tabular format
  -h, --help              Show help

Examples:
  toon convert data.json -o data.toon
  toon convert data.toon --format json
  cat data.json | toon convert -f toon
```

### validate
```bash
toon validate [file] [options]

Options:
  --strict                Enable strict mode
  --round-trip            Test round-trip conversion
  --schema <file>         Validate against JSON schema
  -v, --verbose           Verbose output
  -h, --help              Show help

Examples:
  toon validate data.toon
  toon validate data.toon --round-trip
  toon validate data.toon --schema schema.json
```

### benchmark
```bash
toon benchmark [file] [options]

Options:
  --compare <formats>     Compare formats (json,toon,csv)
  --accuracy-test         Run accuracy test
  --model <name>          LLM model for accuracy test
  -o, --output <file>     Output report file
  --quiet                 Suppress output
  -h, --help              Show help

Examples:
  toon benchmark data.json
  toon benchmark data.json --compare json,toon
  toon benchmark data.json --accuracy-test
```

### format
```bash
toon format [file] [options]

Options:
  -o, --output <file>     Output file path
  -i, --indent <num>      Indentation spaces
  -d, --delimiter <char>  Delimiter character
  --in-place              Format file in-place
  -h, --help              Show help

Examples:
  toon format data.toon -o formatted.toon
  toon format data.toon --in-place
  toon format data.toon -i 4 --delimiter '|'
```

## 성능 특성

### 토큰 효율성

#### 데이터 타입별 절감률
| 데이터 타입 | JSON | TOON | 절감률 |
|-------------|------|------|--------|
| 균일 배열 | 1000 | 620 | 38% |
| 중첩 객체 | 850 | 560 | 34% |
| 혼합 구조 | 1200 | 730 | 39% |
| 순수 테이블 | 800 | 480 | 40% |

#### 모델별 성능
| 모델 | 정확도 | 토큰 절감 |
|------|--------|-----------|
| Claude Haiku 4-5 | 59.8% | 39.6% |
| Gemini 2.5 Flash | 87.6% | 39.6% |
| GPT-5 Nano | 90.9% | 39.6% |
| Grok-4 Fast | 57.4% | 39.6% |

### 파싱 속도

#### 벤치마크 결과 (2025)
```
데이터 크기: 1MB
반복 횟수: 1000회

Format  | Parse Time | Stringify Time | Total
--------|------------|----------------|-------
JSON    | 12.5ms     | 8.3ms          | 20.8ms
TOON    | 15.2ms     | 10.1ms         | 25.3ms
Overhead: +21.6%
```

#### 대용량 데이터 (10MB)
```
Format  | Parse Time | Memory Usage
--------|------------|-------------
JSON    | 125ms      | 15.2MB
TOON    | 152ms      | 16.8MB
Overhead: +21.6%     | +10.5%
```

## 호환성 매트릭스

### 언어 구현
| 언어 | 버전 | 상태 | 패키지 이름 |
|------|------|------|-------------|
| TypeScript | ≥4.0 | Stable | @toon-format/toon |
| Python | ≥3.8 | Stable | toon-format |
| Go | ≥1.18 | Stable | github.com/toon-format/toon-go |
| Rust | ≥1.60 | Stable | toon-format |
| .NET | ≥6.0 | Stable | ToonFormat |
| Java | ≥11 | Beta | com.toonformat |
| PHP | ≥8.0 | Beta | toon-format/toon |

### JSON 호환성
| 기능 | JSON | TOON | 참고 |
|------|------|------|------|
| 객체 | ✅ | ✅ | 완전 호환 |
| 배열 | ✅ | ✅ | 완전 호환 |
| 중첩 | ✅ | ✅ | 완전 호환 |
| 숫자 | ✅ | ✅ | 완전 호환 |
| 불리언 | ✅ | ✅ | 완전 호환 |
| Null | ✅ | ✅ | 완전 호환 |
| 문자열 | ✅ | ✅ | 완전 호환 |
| 유니코드 | ✅ | ✅ | 완전 호환 |
| 이스케이프 | ✅ | ✅ | 완전 호환 |

### 제약사항
- **깊은 중첩**: 5단계 이상 중첩 시 JSON 권장
- **스파스 배열**: JavaScript 스파스 배열 미지원
- **순환 참조**: 순환 참조 감지 및 예외 발생
- **BigInt**: JavaScript BigInt 미지원 (문자열로 변환 필요)

## 보안 고려사항

### 입력 검증
```typescript
// 안전한 파싱
function safeParse(toon: string): ToonValue | null {
  try {
    // 크기 제한
    if (toon.length > 10_000_000) {
      throw new Error('Input too large')
    }
    
    // 파싱
    const data = decode(toon, {
      strict: true,
      validateStructure: true
    })
    
    return data
  } catch (error) {
    console.error('Parse error:', error)
    return null
  }
}
```

### 주입 공격 방지
```typescript
// 사용자 입력 새니타이즈
function sanitizeUserInput(input: string): string {
  // 특수 문자 이스케이프
  return input
    .replace(/,/g, '\\,')
    .replace(/:/g, '\\:')
    .replace(/\[/g, '\\[')
    .replace(/\]/g, '\\]')
}
```

### 리소스 제한
```typescript
// 파싱 타임아웃
import { promiseTimeout } from './utils'

async function parseWithTimeout(
  toon: string,
  timeoutMs: number = 5000
): Promise<ToonValue> {
  return promiseTimeout(
    decode(toon),
    timeoutMs,
    'Parse timeout'
  )
}
```

## 공식 리소스

### 문서
- **공식 사이트**: https://toonformat.dev
- **API 문서**: https://toonformat.dev/docs/api
- **스펙 문서**: https://github.com/toon-format/spec/blob/main/SPEC.md
- **마이그레이션 가이드**: https://toonformat.dev/migration

### 저장소
- **TypeScript**: https://github.com/toon-format/toon
- **Python**: https://github.com/toon-format/toon-python
- **Go**: https://github.com/toon-format/toon-go
- **Rust**: https://github.com/toon-format/toon-rust

### 커뮤니티
- **Discord**: https://discord.gg/toon-format
- **Stack Overflow**: [toon-format] 태그
- **Reddit**: r/toonformat
- **Twitter**: @toonformat

### 도구
- **온라인 플레이그라운드**: https://toonformat.dev/playground
- **VSCode 확장**: https://marketplace.visualstudio.com/items?itemName=toon-format.toon-vscode
- **CLI 도구**: https://www.npmjs.com/package/@toon-format/cli

---

**이 레퍼런스는 TOON 형식의 완전한 기술 사양을 제공합니다.**
