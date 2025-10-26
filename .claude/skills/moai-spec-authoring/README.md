# moai-spec-authoring Skill

**Version**: 1.1.0  
**Created**: 2025-10-23  
**Updated**: 2025-10-27  
**Status**: Active  
**Tier**: Foundation

## Overview

MoAI-ADK SPEC 문서 작성을 위한 종합 가이드입니다. YAML 메타데이터, EARS 문법, 검증 전략을 제공합니다.

## Key Features

- **7 Required + 9 Optional Metadata Fields**: 완전한 레퍼런스와 예제
- **5 EARS Patterns**: Ubiquitous, Event-driven, State-driven, Optional, Constraints
- **Version Lifecycle**: draft에서 production까지 시맨틱 버전 관리
- **TAG Integration**: @SPEC, @TEST, @CODE, @DOC 체인 관리
- **Validation Tools**: 제출 전 체크리스트와 자동화 스크립트
- **Common Pitfalls**: 7가지 주요 이슈에 대한 예방 전략

## File Structure (Progressive Disclosure)

```
.claude/skills/moai-spec-authoring/
├── SKILL.md          # 핵심 개요 + Quick Start (~500 words)
├── reference.md      # 메타데이터 레퍼런스 + EARS 문법 상세
├── examples.md       # 실전 예제 + 패턴 + 트러블슈팅
├── examples/
│   └── validate-spec.sh  # SPEC 검증 스크립트
└── README.md         # 이 파일
```

## Quick Links

- **Quick Start**: [SKILL.md](./SKILL.md#quick-start-5-step-spec-creation)
- **Metadata Reference**: [reference.md](./reference.md#메타데이터-완전-레퍼런스)
- **EARS Syntax**: [reference.md](./reference.md#ears-요구사항-문법)
- **Examples**: [examples.md](./examples.md#실전-ears-예제)
- **Troubleshooting**: [examples.md](./examples.md#트러블슈팅)

## Usage

### Automatic Activation

이 Skill은 다음 경우 자동으로 로드됩니다:
- `/alfred:1-plan` 명령어 실행
- SPEC 문서 생성 요청
- 요구사항 명확화 논의

### Manual Reference

다음 경우 상세 섹션 참조:
- SPEC 작성 모범 사례 학습
- 기존 SPEC 문서 검증
- 메타데이터 이슈 트러블슈팅
- EARS 문법 패턴 이해

## Validation Command

```bash
# SPEC 메타데이터 검증
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-AUTH-001/spec.md

# 중복 ID 확인
rg "@SPEC:AUTH-001" -n .moai/specs/

# 전체 TAG 체인 스캔
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n

# 자동화 스크립트 사용
./examples/validate-spec.sh .moai/specs/SPEC-AUTH-001
```

## Example SPEC Structure

```markdown
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-23
updated: 2025-10-23
author: @YourHandle
priority: high
---

# @SPEC:AUTH-001: JWT Authentication System

## HISTORY
### v0.0.1 (2025-10-23)
- **INITIAL**: JWT 인증 SPEC 초안

## Environment
**Runtime**: Node.js 20.x

## Assumptions
1. 사용자 저장소: PostgreSQL
2. 시크릿 관리: 환경변수

## Requirements

### Ubiquitous Requirements
**UR-001**: 시스템은 JWT 기반 인증을 제공해야 한다.

### Event-driven Requirements
**ER-001**: WHEN 사용자가 유효한 인증 정보를 제출하면, 시스템은 JWT 토큰을 발급해야 한다.

### State-driven Requirements
**SR-001**: WHILE 사용자가 인증된 상태이면, 시스템은 보호된 리소스에 대한 접근을 허용해야 한다.

### Optional Features
**OF-001**: WHERE 다중 인증이 활성화된 경우, 시스템은 OTP 검증을 요구할 수 있다.

### Constraints
**C-001**: IF 토큰이 만료되었다면, 시스템은 접근을 거부해야 한다.
```

## Integration

다음과 원활하게 작동:
- `spec-builder` agent - SPEC 생성
- `moai-foundation-ears` - EARS 문법 패턴
- `moai-foundation-specs` - 메타데이터 검증
- `moai-foundation-tags` - TAG 시스템 통합

## Support

질문이나 이슈가 있을 경우:
1. 포괄적인 문서는 `SKILL.md`, `reference.md`, `examples.md` 참조
2. 가이드된 SPEC 생성을 위해 `/alfred:1-plan` 호출
3. 예제는 `.moai/specs/`의 기존 SPEC 참조

---

**Maintained By**: MoAI-ADK Team  
**Last Updated**: 2025-10-27
