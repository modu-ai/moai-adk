# moai-domain-toon

**TOON Format Expert Skill**

This is a specialized skill for token-efficient data encoding optimized for LLM prompts.

## Core Features

- ✅ Complete TOON format guide
- ✅ JSON ↔ TOON conversion
- ✅ 39.6% token savings
- ✅ TypeScript/Python/CLI integration
- ✅ Performance benchmarks
- ✅ Real-world examples

## Quick Start

### Basic Example

**JSON** (100 tokens):
```json
{"users": [{"id": 1, "name": "Alice", "age": 30}]}
```

**TOON** (60 tokens, 40% savings):
```toon
users[1]{id,name,age}:
  1,Alice,30
```

### Installation

```bash
# TypeScript
npm install @toon-format/toon

# Python
pip install toon-format

# CLI
npm install -g @toon-format/cli
```

## File Structure

```
moai-domain-toon/
├── SKILL.md         # Main skill file
├── examples.md      # Real-world examples
├── reference.md     # API reference
├── patterns.md      # Pattern guide
└── README.md        # This file
```

## Main Contents

### SKILL.md
- Overview and core features
- Basic syntax (objects, arrays, primitives)
- Installation and setup
- Performance benchmarks
- Advanced features
- Context7 integration

### examples.md
- Real-world usage examples
- API response optimization
- Database dumps
- Time series analysis
- TypeScript/Python integration
- CLI automation

### reference.md
- Official specification (TOON Spec v2.0)
- TypeScript type definitions
- CLI command reference
- Performance characteristics
- Compatibility matrix
- Security considerations

### patterns.md
- Recommended patterns (7 types)
- Anti-patterns (7 types)
- Optimization patterns
- Debugging patterns
- Migration strategies

## 성능 (2025)

| 메트릭 | TOON | JSON | 개선 |
|--------|------|------|------|
| 정확도 | 73.9% | 69.7% | +4.2% |
| 토큰 | 60.4% | 100% | -39.6% |
| 파싱 속도 | 125ms | 100ms | +25% |

## 사용 시기

**✅ TOON 사용 (최적)**:
- 균일한 객체 배열
- LLM 프롬프트 임베딩
- API 응답 최적화
- 대량 데이터 포맷팅

**❌ TOON 부적합**:
- 깊은 중첩 구조 (5단계+)
- 순수 테이블 데이터 (CSV 사용)
- 지연 시간 중요 (JSON 네이티브)

## 공식 리소스

- **공식 사이트**: https://toonformat.dev
- **GitHub**: https://github.com/toon-format/toon
- **NPM 패키지**: @toon-format/toon
- **스펙 문서**: https://github.com/toon-format/spec
- **Discord**: https://discord.gg/toon-format

## Works Well With

- `moai-lang-typescript` - TypeScript 통합
- `moai-lang-python` - Python 통합
- `moai-context7-integration` - 최신 문서 접근
- `moai-essentials-perf` - 성능 최적화
- `moai-domain-backend` - 백엔드 구현
- `moai-domain-frontend` - 프론트엔드 구현

## 버전 정보

- **버전**: 1.0.0
- **상태**: Production Ready
- **생성일**: 2025-11-21
- **Tier**: Domain-Specific
- **라이센스**: MIT

## 기여 및 피드백

이 Skill은 MoAI-ADK의 일부입니다. 개선 사항이나 버그 리포트는 MoAI-ADK 저장소에 제출해주세요.

---

**MoAI-ADK Skill Factory로 생성됨**
**Last Updated**: 2025-11-21
