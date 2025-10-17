# @DOC:CONFIG-ADV-001 | Chain: @SPEC:DOCS-003 -> @DOC:CONFIG-001

# Advanced Settings

고급 사용자를 위한 세부 설정 옵션입니다.

## TRUST 임계값 조정

TRUST 원칙 준수 수준을 조정합니다:

```json
{
  "trust_threshold": 90,  // 기본값: 85
  "trust_strict_mode": true
}
```

- **trust_threshold**: 코드 품질 최소 점수 (0-100)
- **trust_strict_mode**: 엄격 모드 활성화 (경고를 오류로 처리)

---

## TAG 전략 커스터마이징

TAG 생성 및 검증 전략:

```json
{
  "tag_strategy": "hybrid",
  "tag_prefix": "CUSTOM",  // 기본값: "" (없음)
  "tag_auto_link": true     // 자동 체인 연결
}
```

### TAG 전략 옵션

- **auto**: 완전 자동 TAG 생성
- **manual**: 수동 TAG 작성
- **hybrid**: 자동 생성 + 수동 검증

---

## 테스트 설정

테스트 실행 및 커버리지 설정:

```json
{
  "test_coverage_min": 85,     // 최소 커버리지 (%)
  "test_parallel": true,        // 병렬 테스트 실행
  "test_timeout": 300           // 테스트 타임아웃 (초)
}
```

---

## 문서 설정

MkDocs 및 문서 생성 옵션:

```json
{
  "docs_auto_build": true,      // 자동 문서 빌드
  "docs_theme": "material",     // 테마 선택
  "docs_language": "ko"         // 문서 언어
}
```

---

## Hook 설정

커스텀 Hook 활성화:

```json
{
  "hooks": {
    "session_start": "custom_hooks.SessionStartHook",
    "pre_tool_use": "custom_hooks.PreToolUseHook",
    "post_tool_use": "custom_hooks.PostToolUseHook"
  }
}
```

자세한 내용: [Custom Hooks](../hooks/custom-hooks.md)

---

## 전체 설정 예제

```json
{
  "project_name": "advanced-project",
  "project_language": "python",
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "mode": "team",
  "git_strategy": "gitflow",

  "trust_threshold": 90,
  "trust_strict_mode": true,

  "tag_strategy": "hybrid",
  "tag_auto_link": true,

  "test_coverage_min": 90,
  "test_parallel": true,

  "docs_auto_build": true,
  "docs_theme": "material",

  "hooks": {
    "session_start": "custom_hooks.SessionStartHook"
  }
}
```

---

**다음**: [Workflow →](../workflow.md)
