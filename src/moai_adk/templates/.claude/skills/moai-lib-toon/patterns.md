# TOON 형식 패턴 및 안티패턴

## 권장 패턴

### 패턴 1: 테이블 우선 설계

**설명**: 균일한 객체 배열은 항상 테이블 형식으로 변환

**좋은 예시**:
```toon
users[3]{id,name,email,role}:
  1,Alice,alice@example.com,admin
  2,Bob,bob@example.com,user
  3,Charlie,charlie@example.com,user
```

**나쁜 예시**:
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

**이유**: 테이블 형식이 40% 더 효율적

### 패턴 2: 명시적 길이 선언

**설명**: 배열 길이를 항상 명시하여 모델 파싱 신뢰성 향상

**좋은 예시**:
```toon
items[5]: item1,item2,item3,item4,item5
```

**나쁜 예시**:
```toon
items: item1,item2,item3,item4,item5  # 길이 누락
```

**이유**: 명시적 길이는 LLM 검증을 가능하게 함

### 패턴 3: 일관된 들여쓰기

**설명**: 파일 전체에서 동일한 들여쓰기 방식 유지

**좋은 예시**:
```toon
company:
  name: Acme
  departments:
    engineering:
      size: 50
    sales:
      size: 30
```

**나쁜 예시**:
```toon
company:
  name: Acme
    departments:  # 잘못된 들여쓰기
      engineering:
        size: 50
  sales:  # 일관성 없는 들여쓰기
    size: 30
```

**이유**: 일관성은 파싱 오류를 방지

### 패턴 4: 스마트 필드 순서

**설명**: 중요한 필드를 먼저 배치하여 가독성 향상

**좋은 예시**:
```toon
products[3]{id,name,price,category,inStock}:
  1,Laptop,1200000,Electronics,true
  2,Mouse,30000,Electronics,true
  3,Keyboard,80000,Electronics,false
```

**나쁜 예시**:
```toon
products[3]{inStock,category,price,name,id}:
  true,Electronics,1200000,Laptop,1
  true,Electronics,30000,Mouse,2
  false,Electronics,80000,Keyboard,3
```

**이유**: 자연스러운 필드 순서는 데이터 이해도를 높임

### 패턴 5: 중첩 최소화

**설명**: 3단계 이상 중첩을 피하고 필요시 키 폴딩 사용

**좋은 예시**:
```toon
# 키 폴딩 사용
user.profile.name: Alice
user.profile.age: 30
user.settings.theme: dark
```

**나쁜 예시**:
```toon
user:
  profile:
    details:
      personal:
        name: Alice  # 5단계 중첩
```

**이유**: 깊은 중첩은 토큰 효율성을 감소시킴

### 패턴 6: 타입 명시적 변환

**설명**: 숫자와 문자열을 명확히 구분

**좋은 예시**:
```toon
data[2]{id,count,price,name}:
  1,100,19.99,Widget
  2,200,29.99,Gadget
```

**나쁜 예시**:
```toon
data[2]{id,count,price,name}:
  "1","100","19.99","Widget"  # 불필요한 따옴표
  2,200,29.99,Gadget
```

**이유**: 명시적 타입은 파싱 정확도를 높임

### 패턴 7: 주석 활용

**설명**: 복잡한 구조에는 주석으로 설명 추가

**좋은 예시**:
```toon
# 사용자 분석 데이터
# 기간: 2025-11-01 ~ 2025-11-30
analytics:
  totalUsers: 10000
  activeUsers: 8500  # 월간 활성 사용자
  # 일별 통계
  daily[30]{date,visitors,conversions}:
    2025-11-01,1250,45
    # ... 28 more days
```

**이유**: 주석은 데이터 맥락을 제공

## 안티패턴

### 안티패턴 1: 불필요한 객체 중첩

**문제**:
```toon
# 비효율적
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

**해결**:
```toon
# 효율적 (40% 토큰 절감)
items[3]{id,name}:
  1,Widget
  2,Gadget
  3,Tool
```

### 안티패턴 2: 혼합 구분자

**문제**:
```toon
users[2]{name,age}:
  Alice,30
  Bob|25  # 오류: 혼합 구분자
```

**해결**:
```toon
users[2]{name,age}:
  Alice,30
  Bob,25
```

### 안티패턴 3: 필드 수 불일치

**문제**:
```toon
products[3]{id,name,price}:
  1,Laptop,1200000
  2,Mouse,30000,InStock  # 4개 필드 (3개 예상)
  3,Keyboard  # 2개 필드 (3개 예상)
```

**해결**:
```toon
# 옵션 1: 필드 추가
products[3]{id,name,price,status}:
  1,Laptop,1200000,null
  2,Mouse,30000,InStock
  3,Keyboard,80000,null

# 옵션 2: 객체 형식 사용
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

### 안티패턴 4: 탭과 스페이스 혼용

**문제**:
```toon
company:
  name: Acme  # 2 spaces
	address: Seoul  # 1 tab (오류)
```

**해결**:
```toon
company:
  name: Acme
  address: Seoul  # 일관된 2 spaces
```

### 안티패턴 5: 특수 문자 미이스케이프

**문제**:
```toon
messages[2]: Hello, World,How are you?  # 쉼표 충돌
```

**해결**:
```toon
messages[2]: "Hello, World","How are you?"
```

### 안티패턴 6: 과도한 키 폴딩

**문제**:
```toon
# 가독성 저하
user.profile.personal.details.name.first: Alice
user.profile.personal.details.name.last: Kim
user.profile.personal.details.age: 30
```

**해결**:
```toon
# 균형있는 중첩
user:
  profile:
    name.first: Alice
    name.last: Kim
    age: 30
```

### 안티패턴 7: 무의미한 배열 길이

**문제**:
```toon
items[0]:  # 빈 배열 (무의미)
```

**해결**:
```toon
items: []  # 또는 필드 제거
```

## 최적화 패턴

### 패턴 A: 배치 변환

**시나리오**: 여러 JSON 파일을 TOON으로 일괄 변환

**구현**:
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

### 패턴 B: 스트리밍 변환

**시나리오**: 대용량 데이터 스트리밍 처리

**구현**:
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

### 패턴 C: 조건부 변환

**시나리오**: 데이터 특성에 따라 형식 선택

**구현**:
```typescript
function smartEncode(data: any): string {
  const eligibility = calculateTabularEligibility(data)

  if (eligibility >= 80) {
    // 테이블 형식 최적
    return encode(data, { detectTabular: true })
  } else if (eligibility >= 50) {
    // 혼합 형식
    return encode(data, {
      detectTabular: true,
      keyFolding: true
    })
  } else {
    // JSON 유지
    return JSON.stringify(data)
  }
}
```

### 패턴 D: 캐싱 전략

**시나리오**: 반복 변환 최적화

**구현**:
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

## 디버깅 패턴

### 패턴 X: 단계별 검증

**설명**: 변환 각 단계에서 검증 수행

**구현**:
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

### 패턴 Y: 오류 복구

**설명**: 파싱 오류 시 자동 복구 시도

**구현**:
```typescript
function resilientDecode(toon: string): any {
  try {
    // Strict 모드 시도
    return decode(toon, { strict: true })
  } catch (error) {
    console.warn('Strict mode failed, retrying with non-strict...')

    try {
      // Non-strict 모드 폴백
      return decode(toon, { strict: false })
    } catch (error2) {
      console.error('Both modes failed')
      throw error2
    }
  }
}
```

### 패턴 Z: 차이점 분석

**설명**: 변환 전후 데이터 비교

**구현**:
```typescript
function analyzeDifference(original: any, converted: any) {
  const diff = {
    added: [],
    removed: [],
    changed: []
  }

  // 재귀적 비교 로직
  function compare(obj1: any, obj2: any, path: string = '') {
    // ... 구현 생략
  }

  compare(original, converted)

  return diff
}
```

## 마이그레이션 패턴

### 패턴 M1: 점진적 전환

**설명**: 기존 JSON 시스템을 단계적으로 TOON으로 전환

**구현**:
```typescript
class HybridFormat {
  encode(data: any, format: 'json' | 'toon' | 'auto'): string {
    if (format === 'auto') {
      // 데이터 특성 분석
      const eligibility = calculateTabularEligibility(data)
      format = eligibility >= 80 ? 'toon' : 'json'
    }

    return format === 'toon'
      ? encode(data)
      : JSON.stringify(data)
  }

  decode(text: string): any {
    // 형식 자동 감지
    if (text.includes('{') && text.includes('}')) {
      return JSON.parse(text)
    } else {
      return decode(text)
    }
  }
}
```

### 패턴 M2: 호환성 레이어

**설명**: JSON/TOON 양방향 호환성 보장

**구현**:
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

## 성능 최적화 패턴

### 패턴 P1: 지연 변환

**설명**: 필요한 시점까지 변환 지연

**구현**:
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

### 패턴 P2: 메모이제이션

**설명**: 동일 입력에 대한 결과 캐싱

**구현**:
```typescript
import memoize from 'lodash/memoize'

const memoizedEncode = memoize(
  (data: any) => encode(data),
  (data: any) => JSON.stringify(data)  // 캐시 키
)
```

### 패턴 P3: 병렬 처리

**설명**: 여러 파일 동시 변환

**구현**:
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

**이 패턴 가이드는 TOON 형식의 효과적 사용법을 제공합니다.**
