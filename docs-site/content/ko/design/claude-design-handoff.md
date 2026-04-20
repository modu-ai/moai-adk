---
title: Claude Design 핸드오프
description: Claude Design 공식 기능과 handoff bundle 활용법
weight: 30
draft: false
---

# Claude Design 핸드오프

## Claude Design 소개

**Claude Design** 은 Anthropic이 2026년 4월 17일 출시한 **AI 기반 디자인 생성 도구**입니다. Claude.ai 내 전용 인터페이스에서 자연어로 UI/UX 디자인을 생성할 수 있습니다.

- **기반 모델:** Claude 3.5 Opus
- **접근 위치:** https://claude.ai/design
- **출력 형식:** 설계 토큰, 컴포넌트 명세, 정적 자산, handoff bundle

## 지원 구독 플랜

| 플랜 | Claude Design 지원 | 참고 |
|---|---|---|
| Free | 미지원 | Claude.ai 계정만으로 접근 불가 |
| Pro | 지원 | $20/월 |
| Max | 지원 | $200/월 |
| Team | 지원 (기본값: 관리자 off) | 팀당 월 청구 |
| Enterprise | 지원 (기본값: 관리자 off) | 계약 기반 |

**주의:** Team 및 Enterprise 플랜은 관리자가 기능을 **기본 비활성화**합니다. 팀 관리자에게 활성화를 요청하세요.

## 지원 입력 형식

Claude Design은 다양한 형식의 입력을 수용합니다:

| 형식 | 설명 |
|---|---|
| **텍스트** | 자연어로 디자인 요구사항 기술 |
| **이미지** | 참고할 디자인 시안 업로드 |
| **DOCX/PPTX** | 기존 문서 또는 프레젠테이션 |
| **XLSX** | 데이터 테이블 및 구조화된 정보 |
| **웹캡처** | URL에서 웹사이트 스크린샷 |
| **Figma** | Figma 파일 및 프레임 임포트 |
| **GitHub** | GitHub 저장소 코드 및 README |

## Handoff Bundle 내보내기

### 1단계: Claude.ai/design 접속

브라우저에서 **https://claude.ai/design** 열기

### 2단계: 디자인 작성

Claude Design 인터페이스에서:
- 자연어로 디자인 설명 입력
- 참고 이미지/문서 업로드
- 실시간으로 UI 프리뷰 생성

예시 프롬프트:
```
기술 회사를 위한 랜딩 페이지를 만들어줘.
- 히로 섹션: 큰 헤더, 가치 제안, CTA 버튼
- 기능 섹션: 카드 3개로 주요 기능 설명
- 색상: 진한 파란색(#1E40AF), 밝은 파란색(#3B82F6)
- 타이포그래피: 모던, 청결함
```

### 3단계: Bundle 내보내기

Claude Design의 **Export** 또는 **Share** 메뉴에서:
- **ZIP 형식:** 모든 설계 파일, 토큰, 자산 포함
- **PDF 형식:** 정적 문서 버전 (선택)
- **Canva/Figma 형식:** 외부 도구로 계속 편집 (선택)
- **HTML/Claude Code:** 코드 스니펫 포함

**권장:** ZIP 형식으로 내보내기

### 4단계: 로컬 저장

내보낸 파일을 로컬 파일시스템에 저장:

```bash
# 예: ~/Downloads/my-design.zip
```

## MoAI-ADK에 Bundle 임포트

### 5단계: /moai design 재실행

Claude Code에서:

```
/moai design
```

경로 A (Claude Design) 선택 후:

```
Bundle 경로를 입력하세요: ~/Downloads/my-design.zip
```

### 6단계: 자동 변환

`moai-workflow-design-import` 스킬이:
- Bundle 파싱
- 설계 토큰을 JSON으로 변환
- 컴포넌트 명세 추출
- 정적 자산 복사

결과 파일:
```
.moai/design/
├── tokens.json          # 설계 토큰 (색상, 타이포, 간격)
├── components.json      # 컴포넌트 명세
└── assets/              # 이미지, 아이콘
```

## Bundle 지원 버전

지원되는 bundle 형식:

| Bundle 버전 | Claude Design 출시 | 상태 | 비고 |
|---|---|---|---|
| v1.0 (초기) | 2026-04-17 | 지원 | 표준 ZIP 형식 |
| v1.1 | 2026-05-xx | 지원 예정 | 확장 메타데이터 |
| v2.0 (미리보기) | 향후 | 미지원 | 호환성 수동 업데이트 필요 |

**화이트리스트:** `.moai/config/sections/design.yaml` 의 `supported_bundle_versions` 참조

## 에러 코드

Bundle 임포트 실패 시:

| 에러 코드 | 원인 | 해결 방법 |
|---|---|---|
| `DESIGN_IMPORT_NOT_FOUND` | Bundle 파일을 찾을 수 없음 | 경로 확인, 파일 존재 확인 |
| `UNSUPPORTED_FORMAT` | ZIP이 아닌 다른 형식 | ZIP 형식으로 다시 내보내기 |
| `UNSUPPORTED_VERSION` | Bundle 버전이 지원되지 않음 | Claude Design에서 최신 버전으로 다시 내보내기 |
| `SECURITY_REJECT` | 보안 검사 실패 (악성 스크립트 감지) | 관리자에게 문의 |
| `MISSING_MANIFEST` | Bundle 구조 손상 | 새로운 Bundle 생성 후 재시도 |

## 폴백 경로

Claude Design이 사용 불가능하거나 Bundle 임포트에 실패한 경우:

### 옵션 1: 경로 B로 전환

```
Bundle 임포트 실패. 경로 B로 전환하겠습니까?
→ [예] [아니오]
```

선택 시 코드 기반 설계 워크플로우로 자동 전환

### 옵션 2: Bundle 재생성

Claude.ai/design으로 돌아가 Bundle 재 내보내기:
1. 기존 디자인 수정 또는 새로 생성
2. Export 메뉴에서 ZIP 재생성
3. 새 파일로 `/moai design` 재실행

## 팀 협업

Claude Design의 Team 구독 기능:

- **실시간 협업:** 여러 팀 멤버가 동시에 디자인 편집
- **공유 Link:** 팀 외부인에게 설계 공유 (읽기 전용)
- **버전 히스토리:** 이전 버전 복원 가능
- **댓글:** 설계에 대한 피드백 기록

**주의:** 기본 비활성화 상태이므로 팀 관리자가 활성화 필요

## 다음 단계

- Bundle 임포트 후 [GAN Loop](./gan-loop.md) 가이드 참고
- 코드 구현 과정에서 [Sprint Contract 프로토콜](./gan-loop.md#sprint-contract-프로토콜) 확인
- 평가 기준 및 반복 프로세스 [4차원 스코어링](./gan-loop.md#4차원-스코어링) 참고
