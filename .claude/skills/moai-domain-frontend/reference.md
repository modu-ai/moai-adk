# moai-domain-frontend - CLI Reference

_Last updated: 2025-10-22_

## Quick Reference

### Installation

```bash
# Installation commands
```

### Common Commands

```bash
# Test
# Lint
# Format
# Build
```

## Tool Versions (2025-10-22)

- **React**: 19.0.0
- **Vue**: 3.5.13
- **Angular**: 19.0.0
- **Vite**: 6.0.5

---

_For detailed usage, see SKILL.md_

## Skill 표준 준수 검증

이 Skill의 Enterprise v4.0 준수 여부를 확인하세요:

### 빠른 검증 (YAML 메타데이터만)

```bash
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
            print('✅ PASS: YAML metadata complete')
            sys.exit(0)
        else:
            print(f'❌ FAIL: Missing fields: {missing}')
            sys.exit(1)
except Exception as e:
    print(f'❌ ERROR: {str(e)[:100]}')
    sys.exit(1)
"
```

### 전체 검증 (moai-skill-validator 사용)

완전한 Skill 검증을 위해 validator를 호출하세요:

```
Skill("moai-skill-validator")
```

이 명령어는 다음을 검증합니다:
- YAML 메타데이터 구조
- 필수 파일 존재 (SKILL.md, reference.md, examples.md)
- Progressive Disclosure 구조
- 보안 검증 (API 키, eval/exec 패턴 감지)
- TAG 시스템 준수
