# TOON 형식 레퍼런스

## 공식 규격 (TOON Specification v2.0)

### 기본 문법

#### 객체 (Object)
```ebnf
object ::= (key ':' value | key ':' newline indent object dedent)*
key ::= identifier | quoted_string
value ::= primitive | array | object
```

#### 배열 (Array)
```ebnf
array ::= simple_array | tabular_array

simple_array ::= '[' length ']' ':' value (',' value)*

tabular_array ::= '[' length ']' '{' fields '}' ':' newline
                  (row_data newline)*

fields ::= identifier (',' identifier)*
row_data ::= value (',' value)*
```

#### 원시값 (Primitives)
```ebnf
primitive ::= number | boolean | string | null

number ::= integer | decimal
boolean ::= 'true' | 'false'
string ::= identifier | quoted_string
null ::= 'null'
```

### 토큰화 규칙

#### 들여쓰기 (Indentation)
- **스페이스**: 2칸 또는 4칸 권장
- **탭**: 사용 가능하지만 혼용 금지
- **일관성**: 파일 전체에서 동일한 방식 유지

#### 구분자 (Delimiter)
- **기본**: `,` (쉼표)
- **대안**: `|`, `;`, `\t` (탭)
- **설정**: `--delimiter` 옵션으로 변경

#### 따옴표 (Quoting)
**따옴표 불필요** (식별자):
- 영문자로 시작
- 영문자, 숫자, 언더스코어만 포함
- 예: `user_id`, `firstName`, `item123`

**따옴표 필수** (문자열):
- 특수 문자 포함: `,`, `:`, `[`, `]`, `{`, `}`
- 공백 포함
- 빈 문자열
- 예: `"Hello, World"`, `"name: value"`, `""`

### 타입 시스템

#### 숫자 (Number)
```toon
# 정수
count: 42
negative: -17

# 부동소수점
price: 19.99
scientific: 1.23e-4

# 특수값
infinity: Infinity
notANumber: NaN
```

#### 불리언 (Boolean)
```toon
isActive: true
hasError: false

# 대소문자 구분
valid: true
invalid: True  # 오류 (대문자 T)
```

#### 문자열 (String)
```toon
# 따옴표 없음 (식별자)
name: Alice
category: electronics

# 따옴표 있음 (특수 문자)
message: "Hello, World!"
path: "/usr/local/bin"

# 이스케이프 시퀀스
escaped: "Line1\nLine2"
quoted: "She said \"Hello\""
```

#### Null
```toon
value: null
optional: null
```

### 배열 선언

#### 길이 명시
```toon
# 필수: [N] 형식
items[3]: item1,item2,item3

# 오류: 길이 누락
items: item1,item2,item3  # 잘못된 문법
```

#### 필드 헤더
```toon
# 테이블 형식
users[2]{name,age}:
  Alice,30
  Bob,25

# 오류: 필드 수 불일치
users[2]{name,age}:
  Alice,30,extra  # 3개 필드 (2개 예상)
```

## TypeScript 타입 정의

### 기본 타입
```typescript
// TOON 값 타입
type ToonValue = 
  | string
  | number
  | boolean
  | null
  | ToonObject
  | ToonArray

// TOON 객체
interface ToonObject {
  [key: string]: ToonValue
}

// TOON 배열
type ToonArray = ToonValue[]
```

### 옵션 인터페이스

#### EncodeOptions
```typescript
interface EncodeOptions {
  /**
   * 엄격 모드 활성화
   * @default false
   */
  strict?: boolean
  
  /**
   * 배열 구분자
   * @default ','
   */
  delimiter?: string
  
  /**
   * 들여쓰기 문자열
   * @default '  ' (2 spaces)
   */
  indentation?: string
  
  /**
   * 테이블 형식 자동 감지
   * @default true
   */
  detectTabular?: boolean
  
  /**
   * 키 폴딩 활성화
   * @default false
   */
  keyFolding?: boolean
  
  /**
   * 경로 확장 활성화
   * @default false
   */
  pathExpansion?: boolean
}
```

#### DecodeOptions
```typescript
interface DecodeOptions {
  /**
   * 엄격 모드 활성화
   * @default false
   */
  strict?: boolean
  
  /**
   * 구조 검증 수행
   * @default true
   */
  validateStructure?: boolean
  
  /**
   * 부분 파싱 허용
   * @default false
   */
  allowPartial?: boolean
  
  /**
   * 타입 강제 변환
   * @default true
   */
  coerceTypes?: boolean
}
```

### 에러 타입

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
