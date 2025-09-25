# moai_adk.core.git_strategy

Git Strategy Classes

개인/팀 모드에 따른 Git 작업 전략을 구현합니다.

@DESIGN:GIT-STRATEGY-001 - Strategy 패턴으로 Git 워크플로우 관리
@PERF:BRANCH-FAST - 브랜치 작업 최적화 (빠른 전환)
@SEC:GIT-MED - Git 작업 보안 강화

## Functions

### __init__

```python
__init__(self, project_dir, config)
```

### get_current_branch

현재 브랜치명 반환

강화된 에러 처리 및 fallback 메커니즘 (@SEC:GIT-MED)

Returns:
    현재 브랜치명

```python
get_current_branch(self)
```

### _get_fallback_branch

Fallback 브랜치명 반환

Returns:
    전략별 기본 브랜치명

```python
_get_fallback_branch(self)
```

### is_git_repository

Git 저장소인지 확인

Returns:
    Git 저장소이면 True

```python
is_git_repository(self)
```

### get_repository_status

저장소 상태 정보 반환

Returns:
    저장소 상태 정보 딕셔너리

```python
get_repository_status(self)
```

### work_context

작업 컨텍스트 - feature 브랜치 생성

Args:
    feature_name: 기능명 (브랜치명에 사용)

Yields:
    None

```python
work_context(self, feature_name)
```

### validate_feature_name

기능명 검증 및 정규화

Args:
    feature_name: 검증할 기능명

Returns:
    정규화된 기능명

Raises:
    ValueError: 유효하지 않은 기능명

```python
validate_feature_name(self, feature_name)
```

### log_git_operation

Git 작업 로깅

Args:
    operation: 작업명
    details: 작업 세부사항

```python
log_git_operation(self, operation, details)
```

### _get_base_branch

베이스 브랜치 확인

Returns:
    베이스 브랜치명 (main, master, develop 등)

```python
_get_base_branch(self)
```

### _create_feature_branch

feature 브랜치 생성

Args:
    feature_name: 기능명

Returns:
    생성된 브랜치명

```python
_create_feature_branch(self, feature_name)
```

### _branch_exists

브랜치 존재 여부 확인

Args:
    branch_name: 확인할 브랜치명

Returns:
    브랜치가 존재하면 True

```python
_branch_exists(self, branch_name)
```

### _pull_latest_changes

최신 변경사항 가져오기 (있는 경우만)

Args:
    branch_name: 업데이트할 브랜치명

```python
_pull_latest_changes(self, branch_name)
```

### get_feature_branch_info

현재 feature 브랜치 정보 반환

Returns:
    feature 브랜치 정보 딕셔너리

```python
get_feature_branch_info(self)
```

## Classes

### GitStrategyBase

Git 전략 기본 클래스

TRUST 원칙 적용:
- T: 추상화를 통한 테스트 가능성 향상
- R: 명확한 인터페이스 정의
- U: 책임 분리 (전략별 구현)
- S: 보안 검증 및 로깅
- T: 작업 추적 가능성

### PersonalGitStrategy

개인 모드: main 브랜치에서 직접 작업

브랜치 생성 없이 현재 브랜치에서 바로 작업을 수행합니다.

특징:
- 브랜치 생성/전환 없음 (@PERF:BRANCH-FAST)
- 단순한 워크플로우 (50% 간소화 달성)
- 개인 프로젝트 최적화

### TeamGitStrategy

팀 모드: feature 브랜치 생성 후 작업

feature 브랜치를 생성하고 해당 브랜치에서 작업을 수행합니다.

특징:
- 체계적인 브랜치 관리
- 협업 친화적 워크플로우
- PR/MR 준비 자동화
