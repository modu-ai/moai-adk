---
name: moai-domain-toon
description: TOON 형식 전문가 - LLM 프롬프트 최적화를 위한 토큰 효율적 데이터 인코딩
allowed-tools:
  - Read
  - Bash
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
version: 1.0.0
tier: Domain-Specific
status: Active
created: 2025-11-21
updated: 2025-11-21
---

# TOON 형식 전문가

## 개요

**TOON (Token-Oriented Object Notation)**은 LLM 입력을 위해 특별히 설계된 컴팩트하고 인간이 읽을 수 있는 인코딩 형식입니다.

**핵심 특징**:
- 토큰 효율성: JSON 대비 39.6% 절감
- 높은 정확도: 73.9% (JSON: 69.7%)
- 무손실 JSON 호환
- 명시적 구조 선언
- LLM 파싱 최적화

---

## 사용 시기

**자동 트리거**:
- LLM 프롬프트 토큰 최적화
- 대량 데이터 포맷팅
- API 비용 절감 전략
- 구조화 데이터 임베딩

**수동 호출**:
- "TOON 형식으로 변환"
- "LLM 토큰 최적화"
- "JSON을 TOON으로 마이그레이션"

---

# Quick Reference

## TOON vs JSON

**JSON**:
```json
{
  "users": [
    {"id": 1, "name": "Alice", "age": 30},
    {"id": 2, "name": "Bob", "age": 25}
  ]
}
```

**TOON** (40% 토큰 절감):
```toon
users[2]{id,name,age}:
  1,Alice,30
  2,Bob,25
```

## 기본 문법

### 객체
```toon
user:
  name: Alice
  age: 30
```

### 배열
```toon
# 단순 배열
colors[3]: red,green,blue

# 테이블 배열
items[2]{id,name}:
  1,Widget
  2,Gadget
```

## 성능 (2025)

| 메트릭 | TOON | JSON |
|--------|------|------|
| 정확도 | 73.9% | 69.7% |
| 토큰 | 60.4% | 100% |
| 절감률 | 39.6% | - |

---

# Core Implementation

## TypeScript 설치

```bash
npm install @toon-format/toon
```

## 기본 사용

```typescript
import { parse, stringify } from '@toon-format/toon'

// TOON → JSON
const data = parse(`
users[2]{name,age}:
  Alice,30
  Bob,25
`)

// JSON → TOON
const toon = stringify({
  users: [
    { name: 'Alice', age: 30 },
    { name: 'Bob', age: 25 }
  ]
})
```

## CLI 도구

```bash
# 설치
npm install -g @toon-format/cli

# JSON → TOON
toon convert data.json --output data.toon

# TOON → JSON
toon convert data.toon --output data.json

# 검증
toon validate data.toon

# 벤치마크
toon benchmark data.json
```

---

# Advanced

## 고급 기능

### Key Folding
```toon
user.profile.name: Alice
user.profile.age: 30
```

### Strict Mode
```typescript
const data = decode(toon, { strict: true })
```

### Custom Delimiter
```bash
toon convert data.json --delimiter '|'
```

## 의사결정 가이드

```
균일한 객체 배열?
  YES → TOON (40% 절감)
  NO → 복잡한 중첩?
    YES → JSON
    NO → TOON
```

---

# Best Practices

## DO
- 테이블 형식 사용 (균일 배열)
- 명시적 길이 선언
- 일관된 들여쓰기
- Context7로 최신 패턴 확인

## DON'T
- 깊은 중첩 (5단계 이상)
- 혼합 구분자
- 필드 수 불일치
- 탭/스페이스 혼용

---

# Context7 통합

```typescript
// 최신 TOON 문서 접근
const docs = await context7.get_library_docs({
  context7_library_id: '/toon-format/toon',
  topic: 'optimization patterns 2025',
  tokens: 5000
})
```

---

# 참고 자료

- **공식 사이트**: https://toonformat.dev
- **GitHub**: https://github.com/toon-format/toon
- **NPM**: @toon-format/toon
- **스펙**: https://github.com/toon-format/spec

## 추가 파일

- [examples.md](examples.md) - 실전 예제
- [reference.md](reference.md) - 완전 레퍼런스
- [patterns.md](patterns.md) - 패턴 및 안티패턴

---

## Works Well With

- `moai-lang-typescript` (TypeScript 통합)
- `moai-lang-python` (Python 통합)
- `moai-context7-integration` (최신 문서)
- `moai-essentials-perf` (성능 최적화)
- `moai-domain-backend` (백엔드)
- `moai-domain-frontend` (프론트엔드)

---

**생성일**: 2025-11-21  
**버전**: 1.0.0  
**상태**: Production Ready  
**라이센스**: MIT

---

**End of TOON 형식 전문가 Skill**
