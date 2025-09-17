# MoAI-ADK 설정 파일

## ⚙️ 설정 파일 개요

MoAI-ADK는 두 개의 주요 설정 파일을 사용합니다:
- `.claude/settings.json`: Claude Code 통합 설정
- `.moai/config.json`: MoAI 시스템 설정
- `.moai/version.json`: 설치된 템플릿/패키지 버전 메타데이터

`moai update --check` 또는 `moai status` 명령은 `version.json`을 읽어 템플릿 버전이 최신인지 확인합니다. 업데이트가 적용되면 파일이 자동으로 갱신됩니다.

## .claude/settings.json

### 기본 구조
```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": ["Read(**)", "Grep", "Glob", "Task", "Bash"],
    "deny": []
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write|Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/pre_write_guard.py",
            "timeout": 60
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "python3 -c 'import json,sys;d=json.load(sys.stdin);fp=(d.get(\"tool_input\") or {}).get(\"file_path\",\"\");print(f\"PostToolUse: updated {fp}\")'",
            "timeout": 30
          }
        ]
      }
    ]
  },
  "env": {
    "MOAI_PROJECT": "true",
    "MOAI_VERSION": "vX.Y.Z"
  },
  "enableAllProjectMcpServers": true,
  "cleanupPeriodDays": 30,
  "includeCoAuthoredBy": true
}
```

> **defaultMode 유효 값**: `default`, `acceptEdits`, `plan`, `bypassPermissions`.  
> **명령 패턴**: `Bash`(모든 Bash 허용), `Bash(npm:*)`(특정 명령), `Read(**)` 등.

### Hook 설정 (전체 구성)
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write|Bash",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/pre_write_guard.py",
            "timeout": 60
          }
        ]
      },
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/constitution_guard.py"
          },
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/tag_validator.py"
          }
        ]
      },
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/awesome/backup_before_edit.py",
            "description": "Backup files before editing"
          }
        ]
      },
      {
        "matcher": "Bash|WebFetch",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/policy_block.py"
          }
        ]
      }
    ],
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
      },
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/awesome/test_runner.py",
            "description": "Run tests after code changes"
          }
        ]
      },
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/awesome/security_scanner.py",
            "description": "Security vulnerability scanner"
          }
        ]
      },
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/post_stage_guard.py"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "python3 $CLAUDE_PROJECT_DIR/.claude/hooks/moai/session_start_notice.py"
          }
        ]
      }
    ]
  }
}
```

### Hook 폴더 구조
```
.claude/hooks/
├── moai/              # MoAI 워크플로우 전용 Hook (6개)
│   ├── pre_write_guard.py
│   ├── constitution_guard.py
│   ├── tag_validator.py
│   ├── policy_block.py
│   ├── post_stage_guard.py
│   └── session_start_notice.py
└── awesome/           # 범용 개발 생산성 Hook (5개)
    ├── auto_formatter.py
    ├── auto_git_commit.py
    ├── backup_before_edit.py
    ├── test_runner.py
    └── security_scanner.py
```
```

> **기본 템플릿**은 `pre_write_guard.py`만 활성화된 최소 구성으로 제공됩니다.
> **완전 구성**에서는 MoAI 워크플로우 Hook(6개) + Awesome Hook(5개) = 총 11개 Hook이 활성화됩니다.
>
> **Hook 체인 특징**:
> - **안전성**: 백업 → 검증 → 실행 → 포매팅 → 커밋 → 테스트 → 보안 스캔
> - **자동화**: 개발자 개입 없이 품질 보장
> - **투명성**: 백그라운드에서 실행되어 워크플로우 방해 없음

## .moai/config.json

### 기본 설정
```json
{
  "project": {
    "name": "My Project",
    "version": "1.0.0",
    "type": "web-app",
    "language": "typescript",
    "framework": "nextjs"
  },
  "moai": {
    "version": "vX.Y.Z",
    "constitution_version": "1.0",
    "pipeline_stage": "INIT"
  },
  "agents": {
    "code-generator": {
      "model": "opus",
      "style": "defensive",
      "test_coverage_threshold": 0.8
    },
    "doc-syncer": {
      "auto_commit": false,
      "sync_frequency": "immediate"
    }
  },
  "quality": {
    "constitution_check": true,
    "tag_validation": true,
    "test_coverage_min": 0.8,
    "max_complexity": 10
  },
  "integrations": {
    "github": {
      "auto_pr": false,
      "ci_cd": true
    },
    "slack": {
      "enabled": false,
      "webhook_url": ""
    }
  }
}
```

### 고급 설정
```json
{
  "templates": {
    "spec_template": "ears",
    "task_template": "tdd",
    "custom_variables": {
      "COMPANY_NAME": "My Company",
      "PROJECT_PREFIX": "MP"
    }
  },
  "automation": {
    "auto_sync": true,
    "auto_index": true,
    "auto_commit": false,
    "parallel_tasks": 3
  },
  "notifications": {
    "gate_failure": true,
    "coverage_drop": true,
    "constitution_violation": true
  }
}
```

## 환경 변수

### 시스템 환경 변수
```bash
# MoAI 프로젝트 식별
export MOAI_PROJECT=true
export MOAI_VERSION=vX.Y.Z

# 성능 설정
export MOAI_MAX_PARALLEL_TASKS=5
export MOAI_TIMEOUT=300

# 디버그 설정
export MOAI_DEBUG=false
export MOAI_LOG_LEVEL=INFO
```

### Claude Code 환경 변수
```json
{
  "env": {
    "MOAI_PROJECT": "true",
    "MOAI_VERSION": "vX.Y.Z",
    "CONSTITUTION_MODE": "strict",
    "TAG_VALIDATION": "enabled"
  }
}
```

## 프로젝트별 커스터마이징

### 팀 설정 예시
```json
{
  "team": {
    "name": "Frontend Team",
    "coding_style": "prettier",
    "review_required": true,
    "pair_programming": false
  },
  "workflows": {
    "feature_branch": true,
    "hotfix_branch": true,
    "release_branch": false
  },
  "policies": {
    "force_tests": true,
    "require_docs": true,
    "block_direct_push": true
  }
}
```

### 언어별 설정
```json
{
  "language_config": {
    "typescript": {
      "strict_mode": true,
      "no_any": true,
      "coverage_threshold": 0.9
    },
    "python": {
      "type_hints": true,
      "docstrings": true,
      "coverage_threshold": 0.8
    }
  }
}
```

## 설정 관리 명령어

### 설정 확인
```bash
# 현재 설정 표시
moai config show

# 특정 설정 확인
moai config get agents.code-generator.model

# 설정 유효성 검사
moai config validate
```

### 설정 변경
```bash
# 설정 업데이트
moai config set quality.test_coverage_min 0.9

# 에이전트 모델 변경
moai config set agents.code-generator.model sonnet

# 설정 리셋
moai config reset
```

## 글로벌 vs 프로젝트 설정

### 글로벌 설정 (v0.1.13)
```bash
# 전역 리소스 위치
~/.claude/moai/
~/.moai/resources/

# 글로벌 설정 파일
~/.moai/global_config.json
```

### 프로젝트 설정 우선순위
1. 프로젝트 `.moai/config.json`
2. 글로벌 `~/.moai/global_config.json`
3. 시스템 기본값

설정 파일은 **프로젝트 특성에 맞는 맞춤형 개발 환경**을 제공합니다.

## 템플릿 모드 설정 (vNext)

프로젝트에서 문서 템플릿 설치/참조 방식을 선택할 수 있습니다.

- 옵션: `templates.mode` = `copy` | `package`
- 기본값: `copy`
- 동작:
  - `copy`: 설치 시 `.moai/_templates/`를 프로젝트에 복사합니다(현행 기본).
  - `package`: 설치 시 `.moai/_templates/` 복사를 생략하고, 템플릿 생성 시 패키지 내장 템플릿으로 폴백합니다.

예시
```json
{
  "templates": {
    "mode": "package",
    "spec_template": "ears",
    "task_template": "tdd"
  }
}
```

주의
- `package` 모드에서도 프로젝트별 오버라이드가 필요하면 `.moai/_templates/` 디렉토리를 수동으로 생성하여 원하는 템플릿만 추가하면 됩니다(프로젝트가 우선).
- TemplateEngine 탐색 순서: 프로젝트 `.moai/_templates` → 패키지 `moai_adk.resources/templates/.moai/_templates`.
