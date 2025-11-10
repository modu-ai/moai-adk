---
title: 프로젝트 관리 가이드
description: MoAI-ADK 프로젝트 초기화, 설정, 배포 완전 가이드
status: stable
---

# 프로젝트 관리 가이드

MoAI-ADK 프로젝트의 전체 생명주기를 관리하는 방법을 배웁니다. 초기화부터 설정, 배포까지 모든 단계를 다룹니다.

## 🎯 프로젝트 관리 3단계

### [1. 프로젝트 초기화](init.md)
- `moai-adk init` 명령어로 새 프로젝트 생성
- 프로젝트 템플릿 선택
- 필요한 파일 구조 자동 생성

**주요 생성 항목**:
- `.moai/config.json` - 프로젝트 메타데이터
- `.claude/` - Claude Code 설정 (agents, commands, skills, hooks)
- `pyproject.toml` - Python 프로젝트 설정
- `pytest.ini` - 테스트 설정

### [2. 설정 관리](config.md)
- `.moai/config.json` 상세 설정
- 언어 및 지역화 설정
- 개발 환경 커스터마이징
- Hook과 Agent 설정

**핵심 설정**:
- 프로젝트 메타데이터 (이름, 버전, 설명)
- 언어 및 도메인 설정
- Git 워크플로우 설정
- 보고서 생성 정책

### [3. 배포 전략](deploy.md)
- 로컬 개발 환경 구성
- Docker 컨테이너화
- 클라우드 플랫폼 배포 (Vercel, Railway, AWS)
- CI/CD 파이프라인 구축
- 모니터링 및 로깅

## 📊 프로젝트 구조

```
my-awesome-project/
├── .moai/              # MoAI-ADK 메타데이터
│   ├── config.json     # 프로젝트 설정
│   ├── docs/           # 자동 생성 문서
│   └── reports/        # 분석 및 리포트
├── .claude/            # Claude Code 설정
│   ├── agents/         # Sub-agent 커스터마이징
│   ├── commands/       # 슬래시 커맨드
│   ├── skills/         # 프로젝트별 Skill
│   └── hooks/          # 자동화 Hooks
├── src/                # 소스 코드
├── tests/              # 테스트 코드
├── docs/               # 프로젝트 문서
└── pyproject.toml      # Python 프로젝트 설정
```

## 🔄 Alfred 통합

프로젝트 관리는 Alfred SuperAgent와 완벽하게 통합됩니다:

- `/alfred:0-project` - 프로젝트 설정 최적화
- `/alfred:1-plan` - 요구사항 SPEC 작성
- `/alfred:2-run` - TDD 구현
- `/alfred:3-sync` - 문서 동기화 및 배포

[Alfred 워크플로우 완전 가이드](../alfred/index.md)

## 📋 체크리스트

프로젝트 설정 시 확인해야 할 항목:

- [ ] 프로젝트 초기화 완료 (`moai-adk init`)
- [ ] `.moai/config.json` 검토 및 커스터마이징
- [ ] Git 워크플로우 설정 확인
- [ ] 개발 환경 구성 완료
- [ ] CI/CD 파이프라인 설정 (선택)
- [ ] 배포 전략 결정

## 🚀 다음 단계

- [프로젝트 초기화: init.md](init.md)
- [설정 관리: config.md](config.md)
- [배포 전략: deploy.md](deploy.md)
- [Alfred 0-project: 설정 최적화](../alfred/index.md)

---

**Learn more**: 프로젝트 관리는 MoAI-ADK 워크플로우의 기초입니다. 올바른 설정으로 시작하면 개발 생산성이 크게 향상됩니다.
