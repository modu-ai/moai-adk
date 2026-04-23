---
title: "긱뉴스 게시 양식"
type: "geeknews-submission"
---

# MoAI-ADK: Claude Code를 위한 에이전트 개발 키트

## 제목
**MoAI-ADK: 937 Stars, 4-Language Docs인 Claude Code용 오픈소스 에이전트 개발 키트**

## URL
https://github.com/modu-ai/moai-adk

## 한 줄 요약
Go로 만든 단일 바이너리. Claude Code 내에서 `/moai plan` → `/moai run` → `/moai sync` 워크플로우로 자동으로 코드를 생성, 테스트, 문서화합니다.

## 주요 내용

1. **필수 워크플로우 (Plan → Run → Sync)**
   - Plan: SPEC 문서 자동 생성 (EARS 형식)
   - Run: 코드 자동 생성 + 테스트 + 린팅 검증
   - Sync: 문서 자동화 + PR 생성
   - 각 단계 사이 컨텍스트 자동 초기화로 토큰 낭비 방지

2. **24개 전문 에이전트 + 52개 스킬**
   - 백엔드, 프론트엔드, 보안, 디버깅, 테스트, 문서화 전문가
   - 각 에이전트는 특정 도메인에 특화

3. **16개 프로그래밍 언어 지원**
   - Go, Python, TypeScript, Rust, Java 등
   - 언어 자동 감지 후 맞는 LSP/테스트/린팅 도구 자동 로드

4. **자동 품질 보증 (TRUST 5 프레임워크)**
   - Tested: 85% 커버리지 강제
   - Readable: 명확한 이름 및 주석
   - Unified: 코드 스타일 자동화
   - Secured: OWASP 보안 검사
   - Trackable: Conventional Commits 강제

5. **4개 언어 문서**
   - 한국어, 영어, 일본어, 중국어
   - 문서 자동 동기화 시스템

## 기술 사양

- **언어**: Go
- **배포**: 단일 바이너리 (의존성 0개)
- **라이선스**: Apache 2.0
- **GitHub Stars**: 937 (2026년 4월)
- **최신 버전**: v2.12.0 (2026년 4월)

## 주목할 점

- **일관된 코드 품질** (TRUST 5 프레임워크: 테스트 85%+, 린팅, OWASP, Conventional Commits)
- **4개국어 공식 문서** (한/영/일/중 — 7개월 차 OSS로는 이례적)
- **Opus 4.7 + Adaptive Thinking 네이티브 지원**
- **Agent Teams 병렬 실행 모드** (tmux 기반 다중 에이전트 협업)

## 사용자 대상

- Claude Code 활용 개발자
- AI 보조 개발에 관심 있는 팀
- 코드 품질과 자동화를 중시하는 프로젝트 리더
- 다국어 프로젝트 운영 팀
