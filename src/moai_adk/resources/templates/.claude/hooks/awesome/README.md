# Awesome Hooks Collection

이 디렉토리는 개발 생산성을 향상시키는 범용 hook 스크립트들을 포함합니다.

## 📋 Hook 목록

### auto_formatter.py
**Smart Code Formatter**
- 파일 타입에 따라 자동으로 적절한 포매터 실행
- 지원 포매터:
  - JavaScript/TypeScript: Prettier
  - Python: Black
  - Go: gofmt
  - Rust: rustfmt
  - PHP: php-cs-fixer

### auto_git_commit.py
**Intelligent Git Auto-Commit**
- 파일 변경사항에 기반한 의미있는 커밋 메시지 자동 생성
- 이모지를 포함한 구조화된 메시지
- 변경 크기에 따른 분류 (minor/moderate/major)
- 파일 타입별 특화 메시지

### backup_before_edit.py
**Backup Before Edit**
- 파일 편집 전 자동 백업 생성
- 타임스탬프가 포함된 백업 파일명
- 자동 백업 정리 (최근 5개만 유지)
- 기존 파일이 있을 때만 백업 생성

### test_runner.py
**Automatic Test Runner**
- 코드 변경 후 자동으로 관련 테스트 실행
- 프로젝트 타입 자동 감지 (JS/TS, Python, Ruby, Go)
- 다양한 테스트 프레임워크 지원
- 테스트 파일 자체는 실행 제외

### security_scanner.py
**Security Vulnerability Scanner**
- 코드 수정 후 보안 취약점 자동 스캔
- 하드코딩된 시크릿 탐지
- 위험한 함수 사용 패턴 탐지
- 외부 도구 통합 (Semgrep, Bandit, GitLeaks)

## 🔧 설정 방법

### settings.json에 추가 (권장)
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/awesome/auto_formatter.py",
            "description": "Smart code formatter"
          }
        ]
      },
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/awesome/auto_git_commit.py",
            "description": "Intelligent auto-commit"
          }
        ]
      }
    ]
  }
}
```

## 💡 사용 예시

### Auto Formatter
파일 편집 후 자동으로 실행:
- `.py` 파일 → Black 자동 실행
- `.js/.ts` 파일 → Prettier 자동 실행
- `.go` 파일 → gofmt 자동 실행

### Auto Git Commit
파일 수정/생성 시 자동으로 커밋:
- 새 파일: "✨ Add new file: example.py"
- 작은 변경: "🔧 Update config.json (minor: 5 lines)"
- 큰 변경: "🚀 Update main.py (major: 120 lines)"
- 문서 수정: "📝 Update documentation: README.md"
- 테스트 수정: "🧪 Update tests: test_main.py"

## ⚙️ 요구사항

### Auto Formatter
- Node.js & npm (Prettier용)
- Python & Black
- Go
- Rust & rustfmt
- PHP & php-cs-fixer

### Auto Git Commit
- Git
- Python 3.6+

## 🎯 장점

1. **일관된 코드 스타일**: 모든 파일이 자동으로 포매팅됨
2. **의미있는 커밋 히스토리**: 자동 생성된 설명적 커밋 메시지
3. **시간 절약**: 수동 포매팅과 커밋 메시지 작성 불필요
4. **프로젝트별 커스터마이징 가능**: 필요에 따라 쉽게 수정 가능

## 📝 커스터마이징

각 hook 파일은 Python으로 작성되어 있어 쉽게 수정 가능합니다:
- 포매터 추가/제거
- 커밋 메시지 형식 변경
- 파일 타입별 처리 로직 커스터마이징