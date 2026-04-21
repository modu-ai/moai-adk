---
title: 시작하기
description: /moai design 명령어로 하이브리드 디자인 워크플로우 시작
weight: 20
draft: false
---

# 시작하기

## 전제 조건

- MoAI-ADK 프로젝트 초기화 완료
- `.moai/project/brand/` 디렉터리가 존재하거나 새로 생성될 예정
- Claude Code 데스크톱 클라이언트 v2.1.50 이상

## 브랜드 컨텍스트 인터뷰

`/moai design` 실행 시 먼저 **브랜드 컨텍스트 인터뷰**가 진행됩니다.

### 3개 브랜드 파일 생성

인터뷰는 다음 3개 파일을 `.moai/project/brand/` 에 생성합니다:

1. **brand-voice.md** — 브랜드 톤, 용어, 메시지 가이드
2. **visual-identity.md** — 색상, 타이포그래피, 시각 언어
3. **target-audience.md** — 타겟 고객 프로필 및 선호도

### 인터뷰 과정

1. Claude Code에서 `/moai design` 명령어 실행
2. "완료되지 않은 브랜드 컨텍스트가 있습니다"라는 메시지 표시
3. 브랜드 인터뷰 선택 시 `manager-spec` 에이전트가 질문 제시
4. 질문에 대한 답변 입력 (자유 형식)
5. 3개 파일이 자동으로 생성됨

예시 질문:
- "당신의 브랜드 톤은 전문적인가, 친근한가?"
- "주요 브랜드 색상 3가지를 선택하세요"
- "타겟 고객의 주요 문제점은 무엇인가?"

## 경로 선택

브랜드 컨텍스트 작성 후, 두 가지 경로 선택 UI가 나타납니다:

### 옵션 1 (권장) — Claude Design 활용

**Claude.ai/design** 에서 직접 디자인을 생성한 후 **handoff bundle** 로 내보내기

**필요 조건:**
- Claude.ai Pro, Max, Team, 또는 Enterprise 구독

**장점:**
- 직관적인 UI/UX
- 실시간 협업 (팀 구독)
- 다양한 입력 형식 지원 (텍스트, 이미지, Figma, GitHub repo)

### 옵션 2 — 코드 기반 설계

**moai-domain-copywriting** 과 **moai-domain-brand-design** 스킬 활용

**필요 조건:**
- 완료된 `brand-voice.md` 와 `visual-identity.md`

**장점:**
- 추가 구독 불필요
- 자동화된 설계 토큰 생성
- 버전 관리 편의성

## 첫 번째 실행

```bash
# Claude Code에서 실행
/moai design
```

실행 순서:
1. `.agency/` 검사 (migration 안내 표시)
2. 브랜드 컨텍스트 확인
3. 부족한 파일 인터뷰 진행
4. 경로 선택 UI 표시
5. 선택한 경로의 워크플로우 진행

## 설정 확인

생성된 브랜드 파일 확인:

```bash
ls -la .moai/project/brand/
# brand-voice.md
# visual-identity.md
# target-audience.md
```

## 다음 단계

- **경로 A 선택 시:** [Claude Design 핸드오프](./claude-design-handoff.md) 가이드 참고
- **경로 B 선택 시:** [코드 기반 경로](./code-based-path.md) 가이드 참고

## 문제 해결

### 브랜드 컨텍스트 업데이트하기

기존 파일 수정:

```bash
# 원하는 텍스트 편집기에서 수정
vim .moai/project/brand/brand-voice.md
```

변경 사항은 다음 `/moai design` 실행 시 자동으로 반영됩니다.

### 다시 인터뷰 실행하기

```bash
# 현재 파일 백업 후 재실행
mv .moai/project/brand .moai/project/brand.backup
/moai design
```
