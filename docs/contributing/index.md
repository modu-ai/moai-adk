# Contributing

MoAI-ADK 기여 가이드입니다.

## 목차

- [Code of Conduct](code-of-conduct.md) - 행동 강령
- [Pull Request Guide](pull-request-guide.md) - PR 가이드

## 기여 방법

MoAI-ADK는 오픈소스 프로젝트로, 여러분의 기여를 환영합니다!

### 1. 이슈 생성

버그 리포트나 기능 제안은 [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)에 작성해주세요.

### 2. 개발 환경 설정

```bash
# 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 의존성 설치
bun install

# 빌드 및 테스트
bun run build
bun test
```

### 3. SPEC-First TDD 준수

모든 기여는 MoAI-ADK의 핵심 원칙을 따라야 합니다:

1. **SPEC 작성**: `/alfred:1-spec`로 요구사항 명세
2. **TDD 구현**: `/alfred:2-build`로 테스트 주도 개발
3. **문서 동기화**: `/alfred:3-sync`로 문서 업데이트

### 4. Pull Request 제출

- 명확한 PR 제목과 설명 작성
- 테스트 커버리지 ≥ 85% 유지
- TRUST 5원칙 준수
- @TAG 추적성 보장

## 행동 강령

모든 기여자는 [Code of Conduct](code-of-conduct.md)를 준수해야 합니다.

## 질문이 있으신가요?

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- [Discord 커뮤니티](https://discord.gg/moai-adk)
