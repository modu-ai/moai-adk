# moai-domain-toon

**TOON 형식 전문가 Skill**

LLM 프롬프트 최적화를 위한 토큰 효율적 데이터 인코딩 전문 Skill입니다.

## 핵심 기능

- ✅ TOON 형식 완전 가이드
- ✅ JSON ↔ TOON 변환
- ✅ 39.6% 토큰 절감
- ✅ TypeScript/Python/CLI 통합
- ✅ 성능 벤치마크
- ✅ 실전 예제

## 빠른 시작

### 기본 예제

**JSON** (100 토큰):
```json
{"users": [{"id": 1, "name": "Alice", "age": 30}]}
```

**TOON** (60 토큰, 40% 절감):
```toon
users[1]{id,name,age}:
  1,Alice,30
```

### 설치

```bash
# TypeScript
npm install @toon-format/toon

# Python
pip install toon-format

# CLI
npm install -g @toon-format/cli
```

## 파일 구조

```
moai-domain-toon/
├── SKILL.md         # 메인 스킬 파일
├── examples.md      # 실전 예제
├── reference.md     # API 레퍼런스
├── patterns.md      # 패턴 가이드
└── README.md        # 이 파일
```

## 주요 내용

### SKILL.md
- 개요 및 핵심 특징
- 기본 문법 (객체, 배열, 원시값)
- 설치 및 설정
- 성능 벤치마크
- 고급 기능
- Context7 통합

### examples.md
- 실전 사용 예제
- API 응답 최적화
- 데이터베이스 덤프
- 시계열 분석
- TypeScript/Python 통합
- CLI 자동화

### reference.md
- 공식 규격 (TOON Spec v2.0)
- TypeScript 타입 정의
- CLI 명령어 레퍼런스
- 성능 특성
- 호환성 매트릭스
- 보안 고려사항

### patterns.md
- 권장 패턴 (7가지)
- 안티패턴 (7가지)
- 최적화 패턴
- 디버깅 패턴
- 마이그레이션 전략

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
