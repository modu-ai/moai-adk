# MoAI-ADK 스크립트 라이브러리

> **위치**: `.moai/scripts/`
>
> **상태**: Production Ready
>
> **배포**: 일부 포함 (패키지 배포본에 포함되는 스크립트 안내 참고)

---

## 📂 디렉토리 구조

```
.moai/scripts/
│
├── README.md (이 파일)
│
├── utils/                              # 사용자 유틸리티 (패키지 포함 ✅)
│   ├── README.md
│   ├── feedback-collect-info.py       # GitHub 이슈 생성 정보 수집
│   ├── session_analyzer.py            # 세션 성능 분석
│   └── statusline.py                  # 프로젝트 상태 표시줄
│
├── monitoring/                         # TAG 모니터링 (패키지 포함 ✅)
│   ├── README.md
│   └── tag_health_monitor.py          # 주간 TAG 건강 검사
│
├── validation/                         # 검증 도구 (패키지 포함 ✅)
│   ├── README.md
│   ├── tag_dedup_manager.py           # TAG 중복 탐지 및 수정 (통합)
│   └── validate_all_skills.py         # Skill 메타데이터 검증
│
├── dev/                                # 개발자 전용 (패키지 제외 ❌)
│   ├── README.md
│   ├── fix-missing-spec-tags.py       # @SPEC: 태그 자동 추가
│   ├── lint_korean_docs.py            # 한국어 문서 검증
│   ├── validate_mermaid_diagrams.py   # Mermaid 다이어그램 검증
│   ├── init-dev-config.sh             # 개발 환경 설정
│   └── skill-pattern-validator.sh     # Skill 구조 검증
│
├── conversion/                         # 변환 도구
│   └── fix-internal-links.js          # 내부 링크 변환
│
├── analysis/ (예약)                    # 분석 도구 (향후)
└── maintenance/ (예약)                 # 유지보수 도구 (향후)
```

---

## 🎯 빠른 시작

### 모든 스크립트 확인
```bash
# 전체 스크립트 목록
find .moai/scripts -type f \( -name "*.py" -o -name "*.sh" \) | sort
```

### 각 디렉토리별 상세 가이드
- **[utils/README.md](utils/)** - 사용자 유틸리티 (피드백, 분석, 상태)
- **[monitoring/README.md](monitoring/)** - TAG 시스템 모니터링
- **[validation/README.md](validation/)** - TAG 및 Skill 검증
- **[dev/README.md](dev/)** - 개발자 도구

---

## 📊 스크립트 요약

| 스크립트 | 목적 | 빈도 | 배포 | 위치 |
|---------|------|------|------|------|
| **feedback-collect-info** | GitHub 이슈 정보 수집 | 필요시 | ✅ | utils/ |
| **session_analyzer** | 세션 성능 분석 | 주간 | ✅ | utils/ |
| **statusline** | 상태 표시줄 | 지속 | ✅ | utils/ |
| **tag_health_monitor** | TAG 건강 점검 | 주간 | ✅ | monitoring/ |
| **tag_dedup_manager** | TAG 중복 관리 | 필요시 | ✅ | validation/ |
| **validate_all_skills** | Skill 검증 | 배포 전 | ✅ | validation/ |
| **fix-missing-spec-tags** | @SPEC 태그 자동 추가 | 개발 중 | ❌ | dev/ |
| **lint_korean_docs** | 한국어 문서 검증 | 개발 중 | ❌ | dev/ |
| **validate_mermaid_diagrams** | 다이어그램 검증 | 개발 중 | ❌ | dev/ |
| **init-dev-config** | 개발 환경 설정 | 설치 후 | ❌ | dev/ |
| **skill-pattern-validator** | Skill 구조 검증 | 개발 중 | ❌ | dev/ |

---

## 🚀 일반적인 워크플로우

### 일일 작업
```bash
# 프로젝트 상태 확인
python3 .moai/scripts/utils/statusline.py

# SPEC 작업
/alfred:1-plan "기능 설명"

# TAG 중복 확인
python3 .moai/scripts/validation/tag_dedup_manager.py --scan-only
```

### 주간 작업
```bash
# 월요일: TAG 건강 검사
python3 .moai/scripts/monitoring/tag_health_monitor.py --weekly

# 금요일: 성능 분석
python3 .moai/scripts/utils/session_analyzer.py --report html
```

### 배포 전 검증
```bash
# 1. Skill 검증
python3 .moai/scripts/validation/validate_all_skills.py

# 2. TAG 검증
python3 .moai/scripts/validation/tag_dedup_manager.py --scan-only

# 3. 이슈 수집
/alfred:9-feedback
```

---

## 📦 패키지 배포 정책

### 배포본에 포함되는 스크립트 (✅)

**utils/** (3개)
- feedback-collect-info.py
- session_analyzer.py
- statusline.py

**monitoring/** (1개)
- tag_health_monitor.py

**validation/** (2개)
- tag_dedup_manager.py (통합)
- validate_all_skills.py

**총 6개 스크립트 배포**

### 배포본에 제외되는 스크립트 (❌)

**dev/** (5개) - 패키지 개발자 전용
- fix-missing-spec-tags.py
- lint_korean_docs.py
- validate_mermaid_diagrams.py
- init-dev-config.sh
- skill-pattern-validator.sh

**이유**: 패키지 개발/유지보수 목적, 최종 사용자 불필요

### pyproject.toml 설정

```toml
[tool.poetry]
packages = [{include = "moai_adk"}]
exclude = [
  ".moai/scripts/dev/*",   # 개발자 전용 제외
  ".moai/scripts/**/test_*.py",
]
```

---

## 🔄 스크립트 개발 정책

### 새 스크립트 추가 기준

✅ **포함되어야 함**:
- 최종 사용자가 필요한 기능
- 패키지 기능의 일부
- 정기적 유지보수 필요

❌ **제외되어야 함**:
- 개발/유지보수 전용 도구
- 로컬 환경 설정
- 테스트 또는 디버깅 전용

### 추가 절차

1. **목적 명확화**: 누가, 언제, 왜 사용하는가?
2. **카테고리 선택**: utils/, monitoring/, validation/, dev/ 중 선택
3. **README 작성**: 스크립트 목적 및 사용 방법 문서화
4. **배포 정책**: pyproject.toml에 반영 (필요시 제외)
5. **Git 커밋**: `chore(scripts): Add {script-name}`

---

## 🛠️ 스크립트 유지보수

### 버전 관리

각 스크립트는 다음 형식의 헤더 주석 포함:
```python
#!/usr/bin/env python3
"""
Script Name and Purpose

Version: 1.0.0 (2025-11-13)
Maintained by: MoAI-ADK Team
"""
```

### 문서화

- 각 스크립트마다 상세한 docstring
- 각 디렉토리마다 README.md
- 사용 예시 포함
- 에러 메시지 명확화

### 테스트

스크립트 배포 전:
1. 로컬에서 수동 테스트
2. 다양한 환경에서 테스트
3. 에러 처리 검증
4. 성능 벤치마크

---

## 📚 관련 문서

- **Alfred 워크플로우**: `CLAUDE.md` - 4-Step Agent-Based Workflow Logic
- **TAG 시스템**: `.moai/specs/TAG-REFERENCE.md`
- **Skill 시스템**: `.moai/skills/`
- **개발 가이드**: `CONTRIBUTING.md`

---

## ❓ 자주 묻는 질문

**Q: 새로운 스크립트를 만들어야 합니까?**
- A: 먼저 기존 스크립트가 이미 해당 기능을 제공하는지 확인하세요. 필요한 경우 기존 스크립트 확장을 고려하세요.

**Q: 개발자 스크립트를 배포할 수 있습니까?**
- A: 아니요. `dev/` 디렉토리의 스크립트는 패키지 배포본에 포함되지 않습니다.

**Q: 내 스크립트가 배포본에 포함되려면?**
- A: `utils/`, `monitoring/`, 또는 `validation/` 디렉토리에 배치하고 README에서 최종 사용자 용도임을 명확히 하세요.

---

**마지막 업데이트**: 2025-11-13
**상태**: Production Ready
**관리자**: MoAI-ADK Team
