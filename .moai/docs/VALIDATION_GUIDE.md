# Skill 검증 가이드 (Validation Guide)

**Document Version**: 1.0  
**Last Updated**: 2025-11-12  
**Language**: 한국어 (Korean)

---

## 📋 개요

MoAI-ADK의 모든 Skills은 **Enterprise v4.0** 표준을 준수해야 합니다. 이 가이드는 Skills의 품질을 검증하고 유지하는 방법을 설명합니다.

### 검증 범위

- **YAML 메타데이터**: name, version, status, description 필수 필드
- **파일 구조**: SKILL.md, reference.md, examples.md 존재 여부
- **Progressive Disclosure**: 3단계 구조 (Quick Reference, Implementation, Advanced)
- **보안**: API 키, eval/exec 패턴, 민감한 정보 노출 감지

---

## 🚀 빠른 시작

### 1단계: 단일 Skill 검증 (YAML만)

```bash
cd /path/to/skill
python3 -c "
import yaml
import sys
try:
    with open('SKILL.md') as f:
        content = f.read()
        yaml_str = content.split('---')[1]
        metadata = yaml.safe_load(yaml_str)
        required = ['name', 'version', 'status', 'description']
        missing = [f for f in required if f not in metadata]
        if not missing:
            print('PASS: YAML metadata complete')
            sys.exit(0)
        else:
            print(f'FAIL: Missing fields: {missing}')
            sys.exit(1)
except Exception as e:
    print(f'ERROR: {str(e)[:100]}')
    sys.exit(1)
"
```

### 2단계: 전체 Skill 검증

```
Skill("moai-skill-validator")
```

이 명령어는 모든 검증 체크를 수행합니다.

---

## 📊 검증 체크리스트

### Phase 1: YAML 메타데이터 검증

| 필드 | 요구사항 | 예시 |
|------|---------|------|
| `name` | 필수 | `"moai-lang-python"` |
| `version` | 필수, SemVer | `"4.0.0"` |
| `status` | 필수 | `"stable"` |
| `description` | 필수 | `"Enterprise-grade..."` |
| `allowed-tools` | 필수 | `["Read", "Bash"]` |

### Phase 2: 파일 구조 검증

필수 파일:
- SKILL.md - 메인 Skill 문서
- reference.md - 참고자료 링크
- examples.md - 사용 예제

### Phase 3: Progressive Disclosure 검증

SKILL.md에 필수 섹션:
- Quick Reference (또는 Level 1)
- Implementation (또는 Level 2)  
- Advanced (또는 Level 3)

### Phase 4: 보안 검증

감지해야 할 패턴:
- API 키 하드코딩
- eval(), exec() 사용
- 비밀번호 또는 토큰 노출

### Phase 5: TAG 시스템 검증


---

## 📈 검증 결과 해석

### PASS - 성공
모든 검증 통과 - Skill을 프로덕션에 배포 가능

### WARNING - 경고  
일부 검증 실패 - 수정 필요

### FAIL - 실패
주요 검증 실패 - 긴급 수정 필요

---

## 🔧 일반적인 문제 해결

### Q1: YAML parse error

**원인**: SKILL.md의 frontmatter 형식 오류
**해결**: 다음 형식 확인:
```
---
name: "moai-skill-name"
version: "4.0.0"
status: "stable"
description: "Description"
---
```

### Q2: reference.md 또는 examples.md 누락

**원인**: 선택적 파일 미생성
**해결**: 파일 생성하고 검증 섹션 추가

### Q3: Progressive Disclosure 불완전

**원인**: 3단계 구조 섹션 누락
**해결**: SKILL.md에 다음 섹션 추가:
- Level 1: Quick Reference
- Level 2: Implementation
- Level 3: Advanced

### Q4: 보안 문제 감지

**원인**: 하드코딩된 API 키 등
**해결**: 환경 변수로 변경

### Q5: TAG 형식 오류

**원인**: TAG 형식 비준수

---

## 📋 배치 검증

모든 Skills를 한번에 검증:
```
Skill("moai-skill-validator")
```

---

## ✅ 최종 체크리스트

- YAML 메타데이터 완전성
- 필수 파일 존재
- Progressive Disclosure 구조
- 보안 문제 없음
- TAG 형식 올바름

---

**Last Updated**: 2025-11-12
**Version**: 1.0
