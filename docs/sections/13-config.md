# MoAI-ADK 설정 파일

## ⚙️ 설정 파일 개요

MoAI-ADK는 두 개의 주요 설정 파일을 사용합니다:
- `.claude/settings.json`: Claude Code 통합 설정
- `.moai/config.json`: MoAI 시스템 설정

## .claude/settings.json

### 기본 구조
```json
{
  "defaultMode": "default",
  "env": {
    "MOAI_PROJECT": "true",
    "MOAI_VERSION": "0.1.14"
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|WebFetch",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/policy_block.py"
          }
        ]
      }
    ]
  },
  "enableAllProjectMcpServers": true,
  "cleanupPeriodDays": 30,
  "includeCoAuthoredBy": true
}
```

### Hook 설정
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/constitution_guard.py"
          },
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/tag_validator.py"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/moai/post_stage_guard.py"
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
            "command": "python3 .claude/hooks/moai/session_start_notice.py"
          }
        ]
      }
    ]
  }
}
```

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
    "version": "0.1.14",
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
export MOAI_VERSION=0.1.14

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
    "MOAI_VERSION": "0.1.14",
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