# @DOC:TAG-MISSING-ROOT-CAUSE-ANALYSIS 태깅 실수 근본 원인 분석 및 예방 시스템

## 📋 개요

**날짜**: 2025-11-13
**이슈**: SPEC 파일 3개(plan.md, acceptance.md)에 `@SPEC:` 태그 누락
**영향**: PostToolUse:Edit 훅 에러 발생
**심각도**: 중간 (경고 수준이지만 검증 실패 초래)

---

## 🔍 근본 원인 분석

### 1. 문제의 원인 체인

```
[SPEC 파일 생성 시점]
    ↓
[Alfred의 spec-builder 에이전트]
    ↓
[템플릿으로부터 파일 생성]
    ↓
❌ ISSUE: 3개 파일 중 plan.md, acceptance.md에 @SPEC: 태그 누락
    ↓
[파일 저장됨 (훅 검증 이전)]
    ↓
[Edit 작업 후 훅 실행]
    ↓
[post_tool__tag_auto_corrector.py 실행]
    ↓
[policy_validator.validate_after_modification() 호출]
    ↓
[Missing TAGs 검출]
    ↓
⚠️ PreToolUse:Edit / PostToolUse:Edit 훅 에러 표시
```

### 2. 직접 원인 (Direct Cause)

**Alfred의 SPEC 생성 워크플로우에서 template 불일치**

```
src/moai_adk/templates/.moai/specs/SPEC-{ID}/
├── spec.md          ✅ @SPEC:{ID} 포함
├── plan.md          ❌ @SPEC:{ID} 미포함  ← 템플릿 버그
└── acceptance.md    ❌ @SPEC:{ID} 미포함  ← 템플릿 버그
```

실제로는:
- `spec.md` - 라인 34에 `# @SPEC:CLI-ANALYSIS-001: CLI 명령어 분석...`
- `plan.md` - `# SPEC-CLI-ANALYSIS-001 구현 계획` (태그 없음)
- `acceptance.md` - `# SPEC-CLI-ANALYSIS-001 수용 기준...` (태그 없음)

### 3. 근본 원인 (Root Cause)

**MoAI-ADK의 SPEC 템플릿 구조 불일치**

**위치**: `src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE/`

**문제점**:
1. **비일관적 템플릿**: plan.md와 acceptance.md 템플릿에 `@SPEC:` 태그 미포함
2. **검증 누락**: spec-builder 에이전트가 생성 후 검증하지 않음
3. **정책 미적용**: 모든 SPEC 파일에 `@SPEC:` 태그 포함 강제가 template 단계에 없음

### 4. 환경 원인 (Environmental Cause)

**Configuration에서의 TAG 정책**:

```json
{
  "tags": {
    "policy": {
      "enforcement_mode": "strict",      ← Strict 모드 활성화
      "require_spec_before_code": true,
      "mandatory_directories": [
        "src/",
        "tests/",
        ".moai/specs/"                   ← SPEC 디렉토리 검증 필수
      ]
    }
  }
}
```

**문제**:
- Config는 "strict" 모드를 강제하지만
- 템플릿은 이를 반영하지 않음
- 결과: 생성된 파일이 정책을 위반함

---

## 🏗️ 원인 분석 요약표

| 계층 | 원인 | 설명 | 심각도 |
|------|------|------|--------|
| **Template Layer** | SPEC 템플릿 불일치 | plan.md, acceptance.md에 @SPEC: 태그 없음 | 높음 |
| **Generation Layer** | spec-builder 검증 누락 | 생성 후 TAG 검증 단계 미실행 | 높음 |
| **Policy Layer** | 정책-템플릿 동기화 미흡 | strict 모드와 template이 일관성 없음 | 중간 |
| **Validation Layer** | 훅이 경고만 함 | 훅이 에러를 강제하지 않고 경고만 표시 | 낮음 |

---

## 🛡️ 예방 시스템 구축

### 1단계: 템플릿 수정 (Template Layer)

**파일**: `src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE/`

#### plan.md 수정

```markdown
# @SPEC:{{SPEC_ID}} 구현 계획

## 프로젝트 개요
...
```

**변경사항**:
- 라인 1: `# SPEC...` → `# @SPEC:{{SPEC_ID}} 구현 계획`
- 템플릿 변수: `{{SPEC_ID}}`로 자동 치환

#### acceptance.md 수정

```markdown
# @SPEC:{{SPEC_ID}} 수용 기준 및 테스트 계획

## 개요
...
```

### 2단계: spec-builder 에이전트 강화 (Generation Layer)

**파일**: `.claude/agents/spec-builder.md` (또는 `spec-builder.py` if Python-based)

**추가 검증 로직**:

```python
def validate_spec_files_have_tags(spec_id: str, spec_dir: Path) -> bool:
    """
    생성된 SPEC 파일들이 모두 @SPEC: 태그를 포함하는지 검증

    Returns:
        True if all files have @SPEC: tags
    """
    required_files = ['spec.md', 'plan.md', 'acceptance.md']
    pattern = rf'@SPEC:{spec_id}'

    for filename in required_files:
        filepath = spec_dir / filename
        if not filepath.exists():
            return False

        content = filepath.read_text()
        if not re.search(pattern, content):
            # Auto-fix: add tag to file
            if filename == 'spec.md':
                # Already has tag by design
                continue
            elif filename in ['plan.md', 'acceptance.md']:
                # Add @SPEC: to first line
                lines = content.split('\n')
                if lines[0].startswith('# '):
                    lines[0] = f'# @SPEC:{spec_id} {lines[0][2:]}'
                    filepath.write_text('\n'.join(lines))

    return True
```

**실행 위치**: spec-builder 에이전트의 "REFACTOR" 단계

### 3단계: Config 정책 강화 (Policy Layer)

**파일**: `.moai/config/config.json`

**추가 설정**:

```json
{
  "tags": {
    "policy": {
      "enforcement_mode": "strict",

      "template_sync": {
        "enabled": true,
        "auto_sync_from_package": true,
        "validate_on_sync": true,
        "verify_tag_coverage": true,
        "description": "Ensure all SPEC templates have proper TAG annotations"
      },

      "spec_file_requirements": {
        "required_files": ["spec.md", "plan.md", "acceptance.md"],
        "required_tag_in_all": "@SPEC:",
        "tag_position": "file_header",
        "enforce_in_template": true,
        "description": "All SPEC files MUST have @SPEC: tag in header"
      }
    }
  }
}
```

### 4단계: 훅 강화 (Validation Layer)

**파일**: `.claude/hooks/alfred/pre_tool__spec_tag_validator.py` (신규)

```python
#!/usr/bin/env python3
"""PreToolUse Hook: SPEC File TAG Validation

Blocks SPEC file creation/modification if TAGs are missing.
Enforces that all SPEC files have proper @SPEC: annotations.
"""

def validate_spec_file_tags(file_path: str, content: str) -> bool:
    """
    Validate that SPEC files have @SPEC: tags

    Args:
        file_path: Path to file being created/modified
        content: File content

    Returns:
        True if valid, False if invalid
    """
    # Only validate files in .moai/specs/
    if '.moai/specs/SPEC-' not in file_path:
        return True

    # Files like spec.md, plan.md, acceptance.md MUST have @SPEC: tag
    if file_path.endswith(('spec.md', 'plan.md', 'acceptance.md')):
        if not re.search(r'@SPEC:\S+', content):
            raise PolicyViolation(
                level=PolicyViolationLevel.HIGH,
                type=PolicyViolationType.MISSING_TAGS,
                message=f"SPEC file missing @SPEC: tag: {file_path}",
                guidance="Add @SPEC:{SPEC_ID} to the file header",
                auto_fix_possible=True
            )

    return True
```

**배포 위치**: `.claude/hooks/alfred/pre_tool__spec_tag_validator.py`

---

## 📋 완전 해결 체크리스트

### Phase 1: 즉시 조치 (Current Session)

- [x] **수정**: plan.md, acceptance.md에 @SPEC: 태그 추가 (이미 완료)
- [x] **검증**: 모든 SPEC 파일에 @SPEC: 태그 확인

### Phase 2: 단기 조치 (This Sprint)

- [ ] **템플릿 수정**: `src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE/plan.md` 수정
  - `# SPEC-{{SPEC_ID}} 구현 계획` → `# @SPEC:{{SPEC_ID}} 구현 계획`

- [ ] **템플릿 수정**: `src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE/acceptance.md` 수정
  - `# SPEC-{{SPEC_ID}} 수용 기준...` → `# @SPEC:{{SPEC_ID}} 수용 기준...`

- [ ] **로컬 템플릿 동기화**:
  ```bash
  # 패키지 템플릿 → 로컬 프로젝트 동기화
  cp -r src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE .moai/specs/SPEC-TEMPLATE/
  ```

- [ ] **config.json 업데이트**: spec_file_requirements 섹션 추가

### Phase 3: 중기 조치 (Next Release v0.23.0)

- [ ] **spec-builder 강화**: 생성 후 TAG 검증 로직 추가
- [ ] **Pre-Tool 훅 구현**: `.claude/hooks/alfred/pre_tool__spec_tag_validator.py` 생성
- [ ] **문서 업데이트**: README.md에 SPEC 파일 구조 설명 추가
- [ ] **테스트 추가**:
  - `tests/unit/test_spec_builder_tags.py` - TAG 검증 테스트
  - `tests/hooks/test_spec_tag_validator.py` - 훅 검증 테스트

### Phase 4: 장기 조치 (v0.24.0+)

- [ ] **자동 복구 기능**: 훅이 누락된 TAG를 자동으로 추가
- [ ] **SPEC 감시**: 모든 SPEC 생성 시 자동 검증
- [ ] **정기 감사**: 주 1회 모든 SPEC 파일 검증
- [ ] **CI/CD 통합**: GitHub Actions에서 SPEC TAG 검증

---

## 🎓 학습 내용

### 왜 이런 문제가 발생했는가?

**1. Template과 Policy의 불일치**
```
Config (strict mode) ≠ Template (no tags in plan.md)
```

MoAI-ADK는 엄격한 TAG 정책을 강제하지만, SPEC 생성 템플릿이 이를 반영하지 않았습니다.

**2. 생성 후 검증 부재**

Alfred의 spec-builder가:
- ✅ spec.md는 올바르게 생성 (태그 포함)
- ❌ plan.md, acceptance.md는 템플릿 버그로 태그 누락
- ❌ 생성 후 검증 단계 없음

**3. 훅의 한계**

PostToolUse 훅은:
- ✅ 에러를 감지함 (정책 위반 감지)
- ✅ 경고를 표시함 (⚠️ Missing TAG detected)
- ❌ 하지만 강제하지 않음 (경고만 표시)
- ❌ Pre-Tool 훅이 없어서 생성 단계에서 차단 불가

### 교훈

**핵심 원칙**:
> **"Template → Policy 검증 체인이 필요하다"**

```
Generation (spec-builder)
    ↓ [검증]
Policy (config.json)
    ↓ [검증]
File Creation (Edit/Write)
    ↓ [훅 검증]
Post-Tool Hook
```

**개선된 접근**:
1. **Template 우선**: 템플릿부터 정책을 반영
2. **생성 후 검증**: 생성 직후 필수 TAG 검증
3. **Pre-Tool 차단**: 잘못된 파일 생성을 애초에 차단
4. **자동 복구**: 감지된 문제를 자동 수정

---

## 🚀 적용 방법

### 옵션 A: 즉시 적용 (추천)

```bash
# 1. 패키지 템플릿 수정
# src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE/ 수정

# 2. 템플릿 동기화
cp -r src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE .moai/specs/SPEC-TEMPLATE/

# 3. 검증
rg '@SPEC:' .moai/specs/SPEC-TEMPLATE/

# 4. 커밋
git add src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE/
git add .moai/specs/SPEC-TEMPLATE/
git commit -m "fix: SPEC 템플릿에 @SPEC: 태그 추가 (SPEC-TAG-COVERAGE-001)"
```

### 옵션 B: 계획적 적용 (SPEC 생성)

```bash
# /alfred:1-plan "SPEC 템플릿 TAG 커버리지 강화"
# SPEC-TAG-TEMPLATE-FIX-001 생성
#
# Phase 1: 템플릿 수정
# Phase 2: spec-builder 강화
# Phase 3: 훅 추가
# Phase 4: 테스트 및 검증
```

---

## 📚 참고 자료

- **설정 파일**: `.moai/config/config.json` → `tags.policy`
- **훅 구현**: `.claude/hooks/alfred/post_tool__tag_auto_corrector.py`
- **정책 검증**: `src/moai_adk/core/tags/policy_validator.py`
- **SPEC 템플릿**: `src/moai_adk/templates/.moai/specs/SPEC-TEMPLATE/`
- **관련 SPEC**: @SPEC:TAG-POLICY-001, @SPEC:SPEC-BUILDER-001

---

## 결론

**이 문제는 시스템적으로 해결 가능합니다.**

1. **즉시**: 현재 파일 수정 ✅ (완료)
2. **단기**: 템플릿 수정 + Config 업데이트
3. **중기**: spec-builder 강화 + Pre-Tool 훅 추가
4. **장기**: 자동 복구 + 정기 감시

이 체계적 접근을 통해 **다시는 태깅 실수가 발생하지 않을 것**입니다. 🛡️

---

**문서 ID**: @DOC:TAG-MISSING-ROOT-CAUSE-ANALYSIS
**생성일**: 2025-11-13
**상태**: 완료
